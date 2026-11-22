package config

import (
	"log/slog"
	"reflect"
	"strings"
	"testing"
)

func env(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

func TestLoadAppliesDefaults(t *testing.T) {
	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://localhost/db"}))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	want := Config{Env: Development, HTTPPort: 8080, DatabaseURL: "postgres://localhost/db", LogLevel: slog.LevelInfo, LogProbes: true}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("Load() = %+v, want %+v", cfg, want)
	}
}

func TestLoadReadsAllVariables(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"APP_ENV":              " production ",
		"HTTP_PORT":            "9090",
		"DATABASE_URL":         " postgres://localhost/db ",
		"LOG_LEVEL":            " DEBUG ",
		"LOG_PROBES":           " false ",
		"CORS_ALLOWED_ORIGINS": " https://Loja.example.com , http://localhost:3000",
	}))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	want := Config{Env: Production, HTTPPort: 9090, DatabaseURL: "postgres://localhost/db", LogLevel: slog.LevelDebug, LogProbes: false,
		CORSOrigins: []string{"https://loja.example.com", "http://localhost:3000"},
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("Load() = %+v, want %+v", cfg, want)
	}
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name string
		vars map[string]string
		want []string
	}{
		{"DATABASE_URL missing", map[string]string{}, []string{"DATABASE_URL is not set"}},
		{"DATABASE_URL only whitespace", map[string]string{"DATABASE_URL": " \t\n"}, []string{"DATABASE_URL is not set"}},
		{"APP_ENV unknown", map[string]string{"DATABASE_URL": "x", "APP_ENV": "staging"}, []string{`APP_ENV "staging" is invalid`}},
		{"HTTP_PORT not numeric", map[string]string{"DATABASE_URL": "x", "HTTP_PORT": "abc"}, []string{`HTTP_PORT "abc" is invalid`}},
		{"HTTP_PORT zero", map[string]string{"DATABASE_URL": "x", "HTTP_PORT": "0"}, []string{`HTTP_PORT "0" is invalid`}},
		{"HTTP_PORT above range", map[string]string{"DATABASE_URL": "x", "HTTP_PORT": "65536"}, []string{`HTTP_PORT "65536" is invalid`}},
		{"LOG_LEVEL unknown", map[string]string{"DATABASE_URL": "x", "LOG_LEVEL": "verbose"}, []string{`LOG_LEVEL "verbose" is invalid`}},
		{"LOG_PROBES not boolean", map[string]string{"DATABASE_URL": "x", "LOG_PROBES": "no"}, []string{`LOG_PROBES "no" is invalid`}},
		{"CORS origin with wildcard", map[string]string{"DATABASE_URL": "x", "CORS_ALLOWED_ORIGINS": "*"}, []string{`CORS_ALLOWED_ORIGINS entry "*" is invalid`}},
		{"CORS origin with wildcard subdomain", map[string]string{"DATABASE_URL": "x", "CORS_ALLOWED_ORIGINS": "https://*.example.com"}, []string{`CORS_ALLOWED_ORIGINS entry "https://*.example.com" is invalid`}},
		{"CORS origin with path", map[string]string{"DATABASE_URL": "x", "CORS_ALLOWED_ORIGINS": "https://loja.example.com/"}, []string{`CORS_ALLOWED_ORIGINS entry "https://loja.example.com/" is invalid`}},
		{"CORS origin without scheme", map[string]string{"DATABASE_URL": "x", "CORS_ALLOWED_ORIGINS": "loja.example.com"}, []string{`CORS_ALLOWED_ORIGINS entry "loja.example.com" is invalid`}},
		{"CORS origin not http", map[string]string{"DATABASE_URL": "x", "CORS_ALLOWED_ORIGINS": "ftp://loja.example.com"}, []string{`CORS_ALLOWED_ORIGINS entry "ftp://loja.example.com" is invalid`}},
		{"CORS empty entry", map[string]string{"DATABASE_URL": "x", "CORS_ALLOWED_ORIGINS": "https://a.example.com,,https://b.example.com"}, []string{`CORS_ALLOWED_ORIGINS entry "" is invalid`}},
		{
			"all problems at once",
			map[string]string{"APP_ENV": "prod", "HTTP_PORT": "-1", "LOG_LEVEL": "trace", "LOG_PROBES": "yes"},
			[]string{
				"DATABASE_URL is not set", `APP_ENV "prod" is invalid`, `HTTP_PORT "-1" is invalid`,
				`LOG_LEVEL "trace" is invalid`, `LOG_PROBES "yes" is invalid`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(env(tt.vars))
			if err == nil {
				t.Fatal("Load() returned no error")
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err, want)
				}
			}
		})
	}
}

func TestLoadAcceptsEveryLogLevel(t *testing.T) {
	tests := map[string]slog.Level{"debug": slog.LevelDebug, "info": slog.LevelInfo, "Warn": slog.LevelWarn, "ERROR": slog.LevelError}

	for raw, want := range tests {
		t.Run(raw, func(t *testing.T) {
			cfg, err := Load(env(map[string]string{"DATABASE_URL": "x", "LOG_LEVEL": raw}))
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if cfg.LogLevel != want {
				t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, want)
			}
		})
	}
}
