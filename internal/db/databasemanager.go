package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"ohnitiel/prismatic/internal/config"

	"golang.org/x/sync/errgroup"
)

type Connection struct {
	db  *sql.DB
	err error
}

func (c *Connection) TestConnection(name string, maxAttempts uint8) error {
	var attempt uint8
	for attempt = 1; attempt <= maxAttempts; attempt++ {
		err := c.db.Ping()
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

func (c *Connection) ExecuteQuery(
	ctx context.Context, query string, useCache bool,
	commitTransaction bool, conf *config.Config,
	name string,
) ([]map[string]any, error) {
	if ctx.Err() != nil {
		slog.ErrorContext(ctx, "Context already cancelled", "connection", name)
		return nil, ctx.Err()
	}

	stmt, err := c.db.PrepareContext(ctx, query)
	if err != nil {
		slog.ErrorContext(ctx, "Error preparing statement", "connection", name, "error", err)
		return nil, err
	}
	defer stmt.Close()

	res, err := getQueryResults(ctx, stmt)
	if err != nil {
		slog.ErrorContext(ctx, "Error running query", "connection", name, "error", err)
		return nil, err
	}

	return res, nil
}

type DatabaseManager struct {
	connections map[string]Connection
}

func NewDatabaseManager(connections map[string]Connection) *DatabaseManager {
	return &DatabaseManager{
		connections: connections,
	}
}

func (dm *DatabaseManager) GetConnection(name string) Connection {
	return dm.connections[name]
}

func (dm *DatabaseManager) GetConnections() map[string]Connection {
	return dm.connections
}

func (dm *DatabaseManager) Close() {
	for _, conn := range dm.connections {
		if conn.db != nil {
			conn.db.Close()
		}
	}
}

func (dm *DatabaseManager) LoadConnections(conf *config.Config) {
	dm.connections = make(map[string]Connection)

	for name, conn := range conf.Connections {
		switch conn.Engine {
		case "postgres", "postgresql":
			dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s pool_max_conns=%d connect_timeout=%d sslmode=%s",
				conn.Host, conn.Port, conn.Database, conn.Username, conn.Password, conf.MaxConnections, conf.Timeout,
				conn.SSLMode,
			)
			db, err := sql.Open("pgx", dsn)
			if err != nil {
				dm.connections[name] = Connection{
					err: fmt.Errorf("unable to connect to %s: %w", conn.Host, err),
				}
			} else {
				dm.connections[name] = Connection{db: db}
			}
		}
	}
}

// Executes a query on multiple connections in parallel
func (dm *DatabaseManager) ExecuteMultiQuery(
	ctx context.Context, workers uint8, query string, useCache bool,
	commitTransaction bool, conf *config.Config,
) (map[string][]map[string]any, map[string]error) {
	type result struct {
		name string
		data []map[string]any
		err  error
	}
	var wg sync.WaitGroup

	resChann := make(chan result, len(dm.connections))
	sem := make(chan struct{}, workers)

	for name, conn := range dm.connections {
		if conn.err != nil {
			slog.ErrorContext(ctx, "Skipping connection due to error", "connection", name, "error", conn.err)
			continue
		}

		if conn.db == nil {
			slog.WarnContext(ctx, "Running query on connection", "connection", name)
			continue
		}

		wg.Add(1)
		name, conn := name, conn

		go func() {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			slog.InfoContext(ctx, "Running query on connection", "connection", name)

			res, err := conn.ExecuteQuery(ctx, query, useCache, commitTransaction, conf, name)

			resChann <- result{name: name, data: res, err: err}
		}()
	}

	go func() {
		wg.Wait()
		close(resChann)
	}()

	results := make(map[string][]map[string]any)
	errors := make(map[string]error)
	for r := range resChann {
		if r.err != nil {
			errors[r.name] = r.err
		} else {
			results[r.name] = r.data
		}
	}
	return results, errors
}

func getQueryResults(ctx context.Context, query *sql.Stmt) ([]map[string]any, error) {
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

	results := make([]map[string]any, 0, 100)

	colPointers := make([]any, len(cols))
	colValues := make([]any, len(cols))

	for rows.Next() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

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
