package db

import (
	"context"
	"fmt"
	"log"
	"testing"

	"ohnitiel/prismatic/internal/config"
)

func TestLoadConnection(t *testing.T) {
	cfg, err := config.Load("../../config/config.toml")
	if err != nil {
		log.Fatal("Failed to load configuration")
	}

	connections := LoadConnections(cfg)
	for name, conn := range connections {
		pingErr := conn.db.Ping()
		if pingErr != nil {
			fmt.Printf("%s - %v\n", name, pingErr)
		}
	}
}

func TestQuery(t *testing.T) {
	cfg, err := config.Load("../../config/config.toml")
	if err != nil {
		log.Fatal("Failed to load configuration")
	}

	ctx := context.Background()

	connections := LoadConnections(cfg)
	for name, conn := range connections {
		if conn.err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		if err := TestConnection(name, conn.db, 3); err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		query, err := conn.db.Prepare("SELECT * FROM agh.aac_retornos")
		if err != nil {
			log.Fatalf("Failed to prepare statement: %v", err)
		}
		res, err := GetQueryResults(ctx, query)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v", res)
	}
}
