package config

import (
	"fmt"
	"log"
	"testing"
)

func TestFileLoad(t *testing.T) {
	cfg, err := Load("../../config/config.toml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Loaded config: %+v\n", cfg)
	fmt.Printf("Loaded %d connections.\n", len(cfg.Connections))

	for name, conn := range cfg.Connections {
		fmt.Printf("%s - %s@%s:%d/%s\n",
			name, conn.Username, conn.Host, conn.Port,
			conn.Database)
	}
}
