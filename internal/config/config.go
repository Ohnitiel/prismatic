package config

import (
	"fmt"
	"slices"

	"github.com/BurntSushi/toml"
)

type Connection struct {
	Engine      string `toml:"engine"`
	Environment string `toml:"environment"`
	Host        string `toml:"host"`
	Port        uint16 `toml:"port"`
	Database    string `toml:"database"`
	Username    string `toml:"username"`
	Password    string `toml:"password"`
	SSLMode     string `toml:"sslmode"`
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
	UseCache   bool `toml:"use_cache"`
	TimeToLive uint `toml:"time_to_live"`
}

type Config struct {
	Cache                CacheConfig           `toml:"cache"`
	Locale               string                `toml:"locale"`
	MaxWorkers           uint8                 `toml:"max_workers"`
	MaxRetries           uint8                 `toml:"max_retries"`
	MaxConnections       uint8                 `toml:"max_connections"`
	Timeout              uint8                 `toml:"timeout"`
	Paths                PathConfigs           `toml:"paths"`
	Connections          map[string]Connection `toml:"connections"`
	Logging              LoggerConfigs         `toml:"logger"`
	ConnectionColumnName string                `toml:"connection_column_name"`
}

func (c *Config) validateLoggerConfig() error {
	consoleOutputs := []string{"stderr", "stdout"}

	if !slices.Contains(consoleOutputs, c.Logging.ConsoleOutput) {
		return fmt.Errorf("%s is not in valid console outputs [%v]!", c.Logging.ConsoleOutput, consoleOutputs)
	}

	return nil
}

func loadConnections(path string) (map[string]Connection, error) {
	var connections map[string]Connection

	_, err := toml.DecodeFile(path, &connections)
	if err != nil {
		return nil, fmt.Errorf("Error loading connections TOML: %w", err)
	}

	return connections, nil
}

func Load(path string) (*Config, error) {
	var conf Config

	_, err := toml.DecodeFile(path, &conf)
	if err != nil {
		return nil, fmt.Errorf("Error loading config TOML: %w", err)
	}

	connections, err := loadConnections(conf.Paths.Connections)
	if err != nil {
		return nil, err
	}

	conf.Connections = connections

	return &conf, nil
}
