package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"ohnitiel/prismatic/internal/config"
	"ohnitiel/prismatic/internal/locale"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Manager is a thread-safe manager for database connections
type Manager struct {
	connections map[string]*Connection
}

func NewDatabaseManager() *Manager {
	return &Manager{}
}

func (dm *Manager) GetConnection(name string) *Connection {
	return dm.connections[name]
}

func (dm *Manager) GetConnections() map[string]*Connection {
	return dm.connections
}

func (dm *Manager) Close() {
	for _, conn := range dm.connections {
		if conn.db != nil {
			conn.db.Close()
		}
	}
}

// Loads the database connections from the configuration
func (dm *Manager) LoadConnections(ctx context.Context, conf *config.Config, environment string, connections []string) {
	var wg sync.WaitGroup

	dm.connections = make(map[string]*Connection)
	sem := make(chan struct{}, conf.MaxWorkers)

	for name, conn := range conf.Connections {
		env := conn.Environment[environment]
		if env == nil {
			continue
		}
		if env.Disabled {
			slog.WarnContext(ctx, locale.L.Logs.EnvDisabled, "environment", environment)
			continue
		}

		switch conn.Engine {
		case "postgres", "postgresql":
			dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s connect_timeout=%d sslmode=%s",
				env.Host, env.Port, env.Database, env.Username, env.Password, conf.Timeout,
				conn.SSLMode,
			)
			db, err := sql.Open("pgx", dsn)
			if err != nil {
				dm.connections[name] = &Connection{
					err: fmt.Errorf("unable to connect to %s: %w", conn.Host, err),
				}
			} else {
				dm.connections[name] = &Connection{db: db}
			}
		}

		if len(connections) > 0 && !slices.Contains(connections, name) {
			continue
		}

		wg.Add(1)

		go func(name string) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()

			dm.connections[name].TestConnection(ctx, name, conf.MaxRetries)
		}(name)

	}
	wg.Wait()
	close(sem)
}
