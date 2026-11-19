package config

import (
	"strings"
	"testing"
)

func env(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

func TestLoadAplicaPadroes(t *testing.T) {
	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://localhost/db"}))
	if err != nil {
		t.Fatalf("Load() erro inesperado: %v", err)
	}

	want := Config{Env: Development, HTTPPort: 8080, DatabaseURL: "postgres://localhost/db"}
	if cfg != want {
		t.Errorf("Load() = %+v, quer %+v", cfg, want)
	}
}

func TestLoadLeTodasAsVariaveis(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"APP_ENV":      " production ",
		"HTTP_PORT":    "9090",
		"DATABASE_URL": " postgres://localhost/db ",
	}))
	if err != nil {
		t.Fatalf("Load() erro inesperado: %v", err)
	}

	want := Config{Env: Production, HTTPPort: 9090, DatabaseURL: "postgres://localhost/db"}
	if cfg != want {
		t.Errorf("Load() = %+v, quer %+v", cfg, want)
	}
}

func TestLoadRejeitaConfiguracaoInvalida(t *testing.T) {
	tests := []struct {
		name string
		vars map[string]string
		want []string
	}{
		{"DATABASE_URL ausente", map[string]string{}, []string{"DATABASE_URL não definida"}},
		{"DATABASE_URL só com espaços", map[string]string{"DATABASE_URL": " \t\n"}, []string{"DATABASE_URL não definida"}},
		{"APP_ENV desconhecido", map[string]string{"DATABASE_URL": "x", "APP_ENV": "staging"}, []string{`APP_ENV "staging" inválido`}},
		{"HTTP_PORT não numérica", map[string]string{"DATABASE_URL": "x", "HTTP_PORT": "abc"}, []string{`HTTP_PORT "abc" inválida`}},
		{"HTTP_PORT zero", map[string]string{"DATABASE_URL": "x", "HTTP_PORT": "0"}, []string{`HTTP_PORT "0" inválida`}},
		{"HTTP_PORT acima do limite", map[string]string{"DATABASE_URL": "x", "HTTP_PORT": "65536"}, []string{`HTTP_PORT "65536" inválida`}},
		{
			"todos os problemas de uma vez",
			map[string]string{"APP_ENV": "prod", "HTTP_PORT": "-1"},
			[]string{"DATABASE_URL não definida", `APP_ENV "prod" inválido`, `HTTP_PORT "-1" inválida`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(env(tt.vars))
			if err == nil {
				t.Fatal("Load() não retornou erro")
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("erro %q não contém %q", err, want)
				}
			}
		})
	}
}
