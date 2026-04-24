package mocklogger

import (
	"context"
	"log/slog"
)

func NewMockLogger() *slog.Logger {
	return slog.New(NewMockHandler())
}

type MockHandler struct{}

func NewMockHandler() *MockHandler {
	return &MockHandler{}
}

func (h *MockHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (h *MockHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (h *MockHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *MockHandler) WithGroup(_ string) slog.Handler {
	return h
}
