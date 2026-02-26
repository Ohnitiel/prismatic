package config

import (
	"fmt"
	"log"
	"testing"

	"ohnitiel/prismatic/internal/locale"
)

func TestFileLoad(t *testing.T) {
	_, err := locale.Load("")
	if err != nil {
		log.Fatalf("Failed to load locale: %v", err)
	}
	cfg, err := FromFile("../../config/config.toml")
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
