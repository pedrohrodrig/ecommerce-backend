//go:build integration

package integration

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/database"
	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/httpserver"
)

func TestReadinessFollowsDatabase(t *testing.T) {
	pool, err := database.Open(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	t.Cleanup(pool.Close)

	router := httpserver.NewRouter(httpserver.RouterConfig{
		ServiceName: "ecommerce-backend",
		DB:          pool,
		Logger:      slog.New(slog.DiscardHandler),
	})
	readyz := func() int {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", http.NoBody))
		return rec.Code
	}

	if got := readyz(); got != http.StatusOK {
		t.Fatalf("/readyz with database up = %d, want %d", got, http.StatusOK)
	}

	pool.Close()

	if got := readyz(); got != http.StatusServiceUnavailable {
		t.Errorf("/readyz with database down = %d, want %d", got, http.StatusServiceUnavailable)
	}
}
