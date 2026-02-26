package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5"

	"ohnitiel/prismatic/internal/config"
	"ohnitiel/prismatic/internal/locale"
)

type state int

const (
	active state = iota
	idle
	transaction
)

type Connection struct {
	db    *sql.DB
	err   error
	state state
}

// Tests the connection.
// Retuns true if the connection is active.
// Will attempt up to maxAttempts to handle transient errors
func (c *Connection) TestConnection(name string, maxAttempts uint8) bool {
	var attempt uint8
	for attempt = 1; attempt <= maxAttempts; attempt++ {
		err := c.db.Ping()
		if err != nil {
			slog.Warn(locale.L.Logs.ConnectionFailed,
				"connection", name,
				"attempt", attempt,
				"max_attempts", maxAttempts,
				"error", err,
			)
			time.Sleep(time.Second * time.Duration(attempt*2))
		} else {
			return true
		}
	}
	c.err = fmt.Errorf("connection to %s timeout", name)

	return false
}

// TODO: Implement caching
// TODO: Implement connection pooling
func (c *Connection) ExecuteQuery(
	ctx context.Context, query string, useCache bool,
	commitTransaction bool, conf *config.Config,
	name string, command string,
) (*ResultSet, error) {
	if ctx.Err() != nil {
		slog.ErrorContext(ctx, locale.L.Logs.ContextAlreadyCancelled, "connection", name)
		return nil, ctx.Err()
	}

	// if useCache {
	// 	if results, ok := cache.Get(name, query); ok {
	// 		slog.InfoContext(ctx, locale.L.Logs.QueryResultCache, "connection", name)
	// 		return results, nil
	// 	}
	// }

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		slog.ErrorContext(ctx, locale.L.Logs.ErrorStartingTransaction, "connection", name, "error", err)
		return nil, err
	}

	if commitTransaction {
		defer slog.InfoContext(ctx, locale.L.Logs.CommittingTransaction, "connection", name)
		defer tx.Commit()
	} else {
		defer slog.InfoContext(ctx, locale.L.Logs.RollingBackTransaction, "connection", name)
		defer tx.Rollback()
	}

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		slog.ErrorContext(ctx, locale.L.Logs.ErrorPreparingStatement, "connection", name, "error", err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, locale.L.Logs.ErrorRunningQuery, "error", err)
		return nil, fmt.Errorf("error running query: %w", err)
	}
	defer rows.Close()

	if command == "export" {
		res, err := getQueryResults(ctx, rows)
		if err != nil {
			slog.ErrorContext(ctx, locale.L.Logs.ErrorRunningQuery, "connection", name, "error", err)
			return nil, err
		}

		return res, nil
	}

	// if useCache && cache != nil {
	// 	cache.Set(name, query, res)
	// }

	return nil, nil
}

func getQueryResults(ctx context.Context, rows *sql.Rows) (*ResultSet, error) {
	start := time.Now()

	cols, err := rows.ColumnTypes()
	if err != nil {
		slog.ErrorContext(ctx, locale.L.Logs.ErrorIdentifyingColumns, "error", err)
		return nil, fmt.Errorf("error identifying columns: %w", err)
	}

	results := &ResultSet{
		Columns: make([]Column, len(cols)),
		Rows:    make([][]any, 0, 100),
	}

	for i, col := range cols {
		nullable, _ := col.Nullable()
		results.Columns[i] = Column{
			Ordinal:  i,
			Name:     col.Name(),
			Type:     col.ScanType().Name(),
			Nullable: nullable,
		}
	}

	colPointers := make([]any, len(cols))
	colValues := make([]any, len(cols))

	for rows.Next() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		for i := range colValues {
			colPointers[i] = &colValues[i]
		}

		if err := rows.Scan(colPointers...); err != nil {
			slog.ErrorContext(ctx, locale.L.Logs.ErrorScanningRows, "error", err)
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		row := make([]any, len(cols))
		for i, v := range colValues {
			switch v := v.(type) {
			case []byte:
				row[i] = string(v)
			default:
				row[i] = v
			}
		}
		results.Rows = append(results.Rows, row)
		results.RowCount++
	}

	if err = rows.Err(); err != nil {
		slog.ErrorContext(ctx, locale.L.Logs.GenericRowError, "error", err)
		return nil, fmt.Errorf("generic row error: %w", err)
	}

	results.Duration = time.Since(start)

	return results, nil
}
