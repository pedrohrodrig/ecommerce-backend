package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

const readinessTimeout = 2 * time.Second

type Pinger interface {
	Ping(ctx context.Context) error
}

var (
	statusOK          = []byte(`{"status":"ok"}`)
	statusUnavailable = []byte(`{"status":"unavailable"}`)
)

func liveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, statusOK)
}

func readiness(db Pinger, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			logger.WarnContext(r.Context(), "readiness check failed", "check", "database", "error", err)
			writeJSON(w, http.StatusServiceUnavailable, statusUnavailable)
			return
		}
		writeJSON(w, http.StatusOK, statusOK)
	}
}

func writeJSON(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
