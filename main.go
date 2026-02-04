package main

import (
	"log"

	"ohnitiel/prismatic/cmd/cli"
	"ohnitiel/prismatic/internal/config"
	"ohnitiel/prismatic/internal/logger"
)

func main() {
	cfg, err := config.Load("./config/config.toml")
	if err != nil {
		log.Fatal(err)
	}

	if err := logger.Setup(cfg.Logging); err != nil {
		log.Fatal(err)
	}

	cli.Prismatic(cfg)
}
