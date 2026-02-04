package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"ohnitiel/prismatic/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	// "modernc.org/sqlite"
)

type Connection struct {
	db  *sql.DB
	err error
}

func LoadConnections(conf *config.Config) map[string]Connection {
	connections := make(map[string]Connection, 0)

	for name, conn := range conf.Connections {
		switch conn.Engine {
		case "postgres", "postgresql":
			dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s pool_max_conns=%d connect_timeout=%d sslmode=%s",
				conn.Host, conn.Port, conn.Database, conn.Username, conn.Password, conf.MaxConnections, conf.Timeout,
				conn.SSLMode,
			)
			db, err := sql.Open("pgx", dsn)
			if err != nil {
				connections[name] = Connection{
					err: fmt.Errorf("unable to connect to %s: %w", conn.Host, err),
				}
			} else {
				connections[name] = Connection{db: db}
			}
		}
	}

	return connections
}

func TestConnection(name string, db *sql.DB, maxAttempts uint8) error {
	var attempt uint8
	for attempt = 1; attempt <= maxAttempts; attempt++ {
		err := db.Ping()
		if err != nil {
			slog.Warn("Connection failed",
				"connection", name,
				"attempt", attempt,
				"max_attempts", maxAttempts,
				"error", err,
			)
			time.Sleep(time.Second * time.Duration(attempt*2))
		} else {
			return nil
		}
	}
	return fmt.Errorf("connection to %s timeout", name)
}

func GetQueryResults(ctx context.Context, query *sql.Stmt) ([]map[string]any, error) {
	rows, err := query.QueryContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Error running query", "error", err)
		return nil, fmt.Errorf("error running query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		slog.ErrorContext(ctx, "Error identifying columns", "error", err)
		return nil, fmt.Errorf("error identifying columns: %w", err)
	}

	results := make([]map[string]any, 0)

	for rows.Next() {
		colPointers := make([]any, len(cols))
		colValues := make([]any, len(cols))

		for i := range colValues {
			colPointers[i] = &colValues[i]
		}

		err := rows.Scan(colPointers...)
		if err != nil {
			slog.ErrorContext(ctx, "Error scanning rows", "error", err)
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		row := make(map[string]any)
		for i, name := range cols {
			value := colValues[i]

			switch v := value.(type) {
			case []byte:
				row[name] = string(v)
			case time.Time:
				row[name] = v.Format(time.RFC3339)
			default:
				row[name] = v
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		slog.ErrorContext(ctx, "Generic row error", "error", err)
		return nil, fmt.Errorf("generic row error: %w", err)
	}

	return results, nil
}

func ExecuteMultiQuery(workers uint8, query string, useCache bool,
	commitTransaction bool, connectionColumnName string,
	conf *config.Config,
) map[string][]map[string]any {
	var mu sync.Mutex
	var waitGroup sync.WaitGroup

	sem := make(chan struct{}, workers)
	results := make(map[string][]map[string]any)

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(conf.Timeout)*time.Second)
	defer cancel()

	connections := LoadConnections(conf)
	defer func() {
		for _, conn := range connections {
			if conn.db != nil {
				conn.db.Close()
			}
		}
	}()

	for name, conn := range connections {
		waitGroup.Add(1)
		go func(n string, c Connection) {
			defer waitGroup.Done()

			if c.err != nil {
				slog.ErrorContext(ctx, "Skipping connection due to error", "connection", n, "error", c.err)
				return
			}
			if c.db == nil {
				slog.WarnContext(ctx, "Skipping connection: no database instance", "connection", n)
				return
			}

			sem <- struct{}{}
			defer func() { <-sem }()

			if ctx.Err() != nil {
				slog.WarnContext(ctx, "Context deadline exceeded for connection", "connection", n)
				return
			}

			stmt, err := c.db.PrepareContext(ctx, query)
			if err != nil {
				slog.ErrorContext(ctx, "Context preparation failed", "connection", n, "error", err)
				return
			}
			defer stmt.Close()

			res, err := GetQueryResults(ctx, stmt)
			if err != nil {
				slog.ErrorContext(ctx, "Query execution failed", "connection", n, "error", err)
				return
			}

			mu.Lock()
			results[n] = res
			mu.Unlock()
		}(name, conn)
	}

	waitGroup.Wait()
	return results
}
