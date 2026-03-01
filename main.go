package main

import (
	"embed"
	"log"
	"log/slog"
	"os"

	"ohnitiel/prismatic/cmd/cli"
	"ohnitiel/prismatic/internal/config"
	"ohnitiel/prismatic/internal/locale"
	"ohnitiel/prismatic/internal/logger"
)

//go:embed config/locales/*
//go:embed config/config.toml
//go:embed config/connections.example.toml
var cfgPath embed.FS

func main() {
	installConfig("config")

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

func installConfig(dir string) {
	os.MkdirAll(dir, os.FileMode(0o744))

	subDir, err := cfgPath.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range subDir {
		fileName := dir + "/" + f.Name()
		if f.IsDir() {
			installConfig(fileName)
		} else {
			if _, err := os.Stat(fileName); err == nil {
				continue
			}
			file, err := cfgPath.ReadFile(fileName)
			if err != nil {
				log.Fatal(err)
			}
			os.WriteFile(fileName, file, os.FileMode(0o744))
		}
	}
}
