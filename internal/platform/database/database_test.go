package database

import (
	"strings"
	"testing"
	"time"
)

func TestOpenFalhaSemExporSenha(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"URL inválida", "postgres://u:segredo@localhost:porta/db", "DATABASE_URL inválida"},
		{"banco inacessível", "postgres://u:segredo@localhost:1/db", "banco de dados inacessível"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			pool, err := Open(t.Context(), tt.url)
			elapsed := time.Since(start)

			if err == nil {
				pool.Close()
				t.Fatal("Open() não retornou erro")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("erro %q não contém %q", err, tt.want)
			}
			if strings.Contains(err.Error(), "segredo") {
				t.Errorf("erro %q expõe a senha da URL", err)
			}
			if elapsed > 6*time.Second {
				t.Errorf("Open() levou %s; esperado no máximo 6s", elapsed)
			}
		})
	}
}
