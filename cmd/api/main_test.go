package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunFalhaAntesDeSubirOServidor(t *testing.T) {
	tests := []struct {
		name string
		vars map[string]string
		want string
	}{
		{"DATABASE_URL ausente", map[string]string{}, "DATABASE_URL não definida"},
		{"HTTP_PORT inválida", map[string]string{"DATABASE_URL": "postgres://localhost/db", "HTTP_PORT": "x"}, `HTTP_PORT "x" inválida`},
		{"banco inacessível", map[string]string{"DATABASE_URL": "postgres://u:segredo@localhost:1/db"}, "banco de dados inacessível"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			getenv := func(key string) string { return tt.vars[key] }

			err := run(t.Context(), getenv, &out)
			if err == nil {
				t.Fatal("run() não retornou erro")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("erro %q não contém %q", err, tt.want)
			}
			if out.Len() != 0 {
				t.Errorf("run() escreveu %q antes de subir o servidor", out.String())
			}
		})
	}
}
