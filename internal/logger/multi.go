package logger

import (
	"context"
	"log/slog"
)

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for _, handler := range m.handlers {
		if handler.Enabled(ctx, l) {
			return true
		}
	}

	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, destHandler := range m.handlers {
		if destHandler.Enabled(ctx, r.Level) {
			if err := destHandler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))

	for i, handler := range m.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}

	return NewMultiHandler(newHandlers...)
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))

	for i, handler := range m.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}

	return NewMultiHandler(newHandlers...)
}
