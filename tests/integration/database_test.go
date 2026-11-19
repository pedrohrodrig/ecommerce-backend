//go:build integration

package integration

import (
	"testing"

	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/database"
)

func TestOpenConectaAoPostgres(t *testing.T) {
	pool, err := database.Open(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	t.Cleanup(pool.Close)

	var one int
	if err := pool.QueryRow(t.Context(), "SELECT 1").Scan(&one); err != nil {
		t.Fatalf("SELECT 1: %v", err)
	}
	if one != 1 {
		t.Errorf("SELECT 1 = %d", one)
	}
}
