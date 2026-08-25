package handler_test

import (
	"io"
	"log/slog"
)

// discardLogger silences workflow logging during tests.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
