package db

import (
	"context"
	"log/slog"
	"sync"

	"ohnitiel/prismatic/internal/config"
	"ohnitiel/prismatic/internal/db/sql"
)

type Executor struct {
	manager *Manager
	// cache   *Cache
}

func NewExecutor(manager *Manager) *Executor {
	return &Executor{
		manager: manager,
		// cache:   cache,
	}
}

// Executes a query on multiple connections in parallel
// TODO: Add a caching mechanism when DQL
// TODO: Add a connection column name for DQL
// TODO: Make more memory efficient
func (ex *Executor) ParallelExecution(
	ctx context.Context, workers uint8, query string, useCache bool,
	commitTransaction bool, conf *config.Config, command string,
) (map[string]*ResultSet, map[string]error) {
	type result struct {
		name string
		data *ResultSet
		err  error
	}
	var wg sync.WaitGroup

	resChann := make(chan result, len(ex.manager.connections))
	sem := make(chan struct{}, workers)

	queryType, err := sql.SimpleQueryIdentifier(query)
	if err != nil {
		slog.WarnContext(ctx, "Unable to identify query type")
	}
	slog.InfoContext(ctx, "Identified query type", "query_type", queryType)

	if command == "run" && queryType == sql.DQL {
		slog.WarnContext(ctx, "Running SELECT query without saving results")
	}

	for name, conn := range ex.manager.connections {
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

			res, err := conn.ExecuteQuery(ctx, query, useCache, commitTransaction, conf, name, command)

			if err != nil {
				slog.ErrorContext(ctx, "Error running query on connection", "connection", name, "error", err)
			} else {
				slog.InfoContext(ctx, "Query successful on connection", "connection", name)
			}

			resChann <- result{name: name, data: res, err: err}
		}()
	}

	go func() {
		wg.Wait()
		close(resChann)
	}()

	results := make(map[string]*ResultSet)
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
