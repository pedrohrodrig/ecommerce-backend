package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func New(port int, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func NewRouter(serviceName string, db Pinger, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(logRequests(logger))

	r.Get("/healthz", liveness)
	r.Get("/readyz", readiness(db, logger))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/", serviceInfo(serviceName))
	})
	return r
}

func serviceInfo(serviceName string) http.HandlerFunc {
	body, _ := json.Marshal(map[string]string{"service": serviceName})

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}
}
