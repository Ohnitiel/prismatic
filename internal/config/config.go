package config

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	"ohnitiel/prismatic/internal/locale"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
)

type Environment struct {
	Host     string `toml:"host"`
	Port     uint16 `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Database string `toml:"database"`
	Disabled bool
}

type Connection struct {
	Engine      string `toml:"engine"`
	Host        string `toml:"host"`
	Port        uint16 `toml:"port"`
	Database    string `toml:"database"`
	Username    string `toml:"username"`
	Password    string `toml:"password"`
	SSLMode     string `toml:"sslmode"`
	Environment map[string]*Environment
}

type LoggerConfigs struct {
	ConsoleLevel  string `toml:"console_level"`
	ConsoleOutput string `toml:"console_output"`
	FileLevel     string `toml:"file_level"`
	FileOutput    string `toml:"file_output"`
}

type PathConfigs struct {
	Connections string `toml:"connections"`
}

type CacheConfig struct {
	UseCache   bool   `toml:"use_cache"`
	TimeToLive uint16 `toml:"time_to_live"`
	MaxAge     time.Duration
}

type Config struct {
	Cache                CacheConfig            `toml:"cache"`
	Locale               string                 `toml:"locale"`
	MaxWorkers           uint8                  `toml:"max_workers"`
	MaxRetries           uint8                  `toml:"max_retries"`
	MaxConnections       uint8                  `toml:"max_connections"`
	Timeout              uint8                  `toml:"timeout"`
	Paths                PathConfigs            `toml:"paths"`
	Connections          map[string]*Connection `toml:"connections"`
	Logging              LoggerConfigs          `toml:"logger"`
	ConnectionColumnName string                 `toml:"connection_column_name"`
}

func NewConfig() *Config {
	return &Config{}
}

func FromFile(path string) (*Config, error) {
	conf := NewConfig()

	_, err := toml.DecodeFile(path, conf)
	if err != nil {
		return nil, fmt.Errorf("Error loading config TOML: %w", err)
	}
	conf.Cache.MaxAge = time.Duration(conf.Cache.TimeToLive) * time.Second

	err = conf.loadConnections()
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func (c *Config) GetConnection(name string) *Connection {
	return c.Connections[name]
}

func (c *Config) GetConnections() map[string]*Connection {
	return c.Connections
}

func (c *Config) validateLoggerConfig() error {
	consoleOutputs := []string{"stderr", "stdout"}

	if !slices.Contains(consoleOutputs, c.Logging.ConsoleOutput) {
		return fmt.Errorf("%s is not in valid console outputs [%v]!", c.Logging.ConsoleOutput, consoleOutputs)
	}

	return nil
}

func (c *Connection) Resolve(env *Environment) {
	if env.Host == "" {
		slog.Warn(locale.L.Logs.NoHostSpecified)
		env.Disabled = true
		return
	}
	if env.Database == "" {
		env.Database = c.Database
	}
	if env.Port == 0 {
		env.Port = c.Port
	}
	if env.Username == "" {
		env.Username = c.Username
	}
	if env.Password == "" {
		env.Password = c.Password
	} else {
		env.Password = getPasswordFromEnv(env)
	}
}

func (c *Config) loadConnections() error {
	var connections map[string]*Connection

	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("Error loading .env file: %w", err)
	}

	_, err = toml.DecodeFile(c.Paths.Connections, &connections)
	if err != nil {
		return fmt.Errorf("Error loading connections TOML: %w", err)
	}

	for _, conn := range connections {
		conn.Password = getPasswordFromEnv(conn)

		for _, env := range conn.Environment {
			conn.Resolve(env)
		}
	}

	c.Connections = connections

	return nil
}

func getPasswordFromEnv(info any) string {
	var password string

	switch v := info.(type) {
	case *Environment:
		password = v.Password
	case *Connection:
		password = v.Password
	default:
		return ""
	}

	if strings.HasPrefix(password, "${") && strings.HasSuffix(password, "}") {
		envVar := strings.TrimPrefix(strings.TrimSuffix(password, "}"), "${")
		return os.Getenv(envVar)
	}
	return password
}
