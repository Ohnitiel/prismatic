package main

import (
	"log"
	"log/slog"

	"ohnitiel/prismatic/cmd/cli"
	"ohnitiel/prismatic/internal/config"
	"ohnitiel/prismatic/internal/locale"
	"ohnitiel/prismatic/internal/logger"
)

func main() {
	cfg, err := config.FromFile("./config/config.toml")
	if err != nil {
		log.Fatal(err)
	}

	if err := logger.Setup(cfg.Logging); err != nil {
		log.Fatal(err)
	}

	locale.L, err = locale.Load(cfg.Locale)
	if err != nil {
		slog.Error("Failed to load locale", "error", err)
	}

	cli.Prismatic(cfg)
}
