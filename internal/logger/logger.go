package logger

import (
	"log/slog"
	"os"
	"strings"

	"ohnitiel/prismatic/internal/config"
)

func Setup(cfg config.LoggerConfigs) error {
	var handlers []slog.Handler
	var stdErrLevel slog.Level

	switch strings.ToLower(cfg.ConsoleLevel) {
	case "debug":
		stdErrLevel = slog.LevelDebug
	case "info":
		stdErrLevel = slog.LevelInfo
	case "warn":
		stdErrLevel = slog.LevelWarn
	case "error":
		stdErrLevel = slog.LevelError
	default:
		stdErrLevel = slog.LevelInfo
	}

	stdErrOpts := &slog.HandlerOptions{Level: stdErrLevel}
	handlers = append(handlers, slog.NewTextHandler(os.Stderr, stdErrOpts))

	if cfg.FileOutput != "" {
		logFile, err := os.OpenFile(cfg.FileOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}

		var fileLevel slog.Level

		switch strings.ToLower(cfg.FileLevel) {
		case "debug":
			fileLevel = slog.LevelDebug
		case "info":
			fileLevel = slog.LevelInfo
		case "warn":
			fileLevel = slog.LevelWarn
		case "error":
			fileLevel = slog.LevelError
		default:
			fileLevel = slog.LevelInfo
		}
		fileOpts := &slog.HandlerOptions{
			Level: fileLevel, AddSource: true,
		}

		handlers = append(handlers, slog.NewTextHandler(logFile, fileOpts))
	}

	multi := NewMultiHandler(handlers...)

	slog.SetDefault(slog.New(multi))

	return nil
}
