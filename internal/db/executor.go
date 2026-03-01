package db

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"ohnitiel/prismatic/internal/config"
	"ohnitiel/prismatic/internal/db/sql"
	"ohnitiel/prismatic/internal/locale"
)

type Executor struct {
	manager *Manager
	// cache   *Cache
}

type Summary struct {
	Sucessful int
	Failed    int
	Errors    map[string]error
}

func NewExecutor(manager *Manager) *Executor {
	return &Executor{
		manager: manager,
		// cache:   cache,
	}
}

// Executes a query on multiple connections in parallel
// TODO: Add a caching mechanism when DQL
// TODO: Make more memory efficient
func (ex *Executor) ParallelExecution(
	ctx context.Context, workers uint8, query string, useCache bool,
	commitTransaction bool, conf *config.Config, command string,
	connections []string,
) (map[string]*ResultSet, map[string]error) {
	type result struct {
		name string
		data *ResultSet
		err  error
	}
	var wg sync.WaitGroup
	summary := Summary{Errors: make(map[string]error)}

	resChann := make(chan result, len(ex.manager.connections))
	sem := make(chan struct{}, workers)

	queryType, err := sql.SimpleQueryIdentifier(query)
	if err != nil {
		slog.WarnContext(ctx, locale.L.Logs.UnableIdentifyQueryType)
	}
	slog.InfoContext(ctx, locale.L.Logs.IdentifiedQueryType, "query_type", queryType)

	if command == "run" && queryType == sql.DQL {
		slog.WarnContext(ctx, locale.L.Logs.RunningSelectWithoutSaving)
	}

	for name, conn := range ex.manager.connections {
		if len(connections) > 0 && !slices.Contains(connections, name) {
			continue
		}

		wg.Add(1)
		name, conn := name, conn

		go func() {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			if conn.err != nil {
				slog.ErrorContext(ctx, locale.L.Logs.SkippingConnectionError, "connection", name, "error", conn.err)
				summary.Errors[name] = conn.err
				summary.Failed++
				fmt.Println("Inside conn.err")
				resChann <- result{name: name, data: nil, err: conn.err}
				return
			}

			if conn.db == nil {
				slog.WarnContext(ctx, locale.L.Logs.RunningQueryOnConn, "connection", name)
				err = fmt.Errorf("connection to %s is null", name)
				summary.Errors[name] = err
				summary.Failed++
				fmt.Println("Null db")
				resChann <- result{name: name, data: nil, err: err}
				return
			}

			slog.InfoContext(ctx, locale.L.Logs.RunningQueryOnConn, "connection", name)

			res, err := conn.ExecuteQuery(ctx, query, useCache, commitTransaction, conf, name, command)

			if err != nil {
				slog.ErrorContext(ctx, locale.L.Logs.ErrorRunningQueryOnConn, "connection", name, "error", err)
				summary.Failed++
				summary.Errors[name] = err
			} else {
				slog.InfoContext(ctx, locale.L.Logs.QuerySuccessfulOnConn, "connection", name)
				summary.Sucessful++
			}

			resChann <- result{name: name, data: res, err: err}
		}()
	}

	go func() {
		wg.Wait()
		close(resChann)
		close(sem)
	}()

	results := make(map[string]*ResultSet)
	errors := make(map[string]error)
	for r := range resChann {
		if r.err != nil {
			fmt.Println(r.name)
			errors[r.name] = r.err
		} else {
			results[r.name] = r.data
		}
	}

	textSummary := fmt.Sprintf(locale.L.Logs.QuerySummary, summary.Sucessful, summary.Failed)
	fmt.Println(textSummary)

	return results, errors
}
