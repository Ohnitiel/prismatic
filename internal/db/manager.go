package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"

	"ohnitiel/prismatic/internal/config"
)

// Manager is a thread-safe manager for database connections
type Manager struct {
	connections map[string]Connection
}

func NewDatabaseManager() *Manager {
	return &Manager{}
}

func (dm *Manager) GetConnection(name string) Connection {
	return dm.connections[name]
}

func (dm *Manager) GetConnections() map[string]Connection {
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
func (dm *Manager) LoadConnections(conf *config.Config, environment string) {
	dm.connections = make(map[string]Connection)

	for name, conn := range conf.Connections {
		env := conn.Environment[environment]
		if env.Disabled {
			slog.Warn("Environment is disabled", "environment", environment)
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
				dm.connections[name] = Connection{
					err: fmt.Errorf("unable to connect to %s: %w", conn.Host, err),
				}
			} else {
				dm.connections[name] = Connection{db: db}
			}
		}
	}
}
