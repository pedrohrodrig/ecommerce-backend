package logging

import (
	"io"
	"log/slog"

	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/config"
)

func New(w io.Writer, env config.Environment, level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}

	if env == config.Production {
		return slog.New(slog.NewJSONHandler(w, opts))
	}
	return slog.New(slog.NewTextHandler(w, opts))
}
