package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunFalhaSemBanco(t *testing.T) {
	tests := []struct {
		name        string
		databaseURL string
		want        string
	}{
		{"DATABASE_URL vazia", "", "DATABASE_URL não definida"},
		{"DATABASE_URL só com espaços", " \t\n", "DATABASE_URL não definida"},
		{"banco inacessível", "postgres://u:segredo@localhost:1/db", "banco de dados inacessível"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer

			err := run(t.Context(), &out, tt.databaseURL)
			if err == nil {
				t.Fatal("run() não retornou erro")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("erro %q não contém %q", err, tt.want)
			}
			if out.Len() != 0 {
				t.Errorf("run() escreveu %q antes de conectar ao banco", out.String())
			}
		})
	}
}
