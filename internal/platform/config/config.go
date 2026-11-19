package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

const defaultHTTPPort = 8080

type Config struct {
	Env         Environment
	HTTPPort    int
	DatabaseURL string
}

func Load(getenv func(string) string) (Config, error) {
	var problems []string

	cfg := Config{
		Env:         Development,
		HTTPPort:    defaultHTTPPort,
		DatabaseURL: strings.TrimSpace(getenv("DATABASE_URL")),
	}

	if cfg.DatabaseURL == "" {
		problems = append(problems, "DATABASE_URL não definida")
	}

	if raw := strings.TrimSpace(getenv("APP_ENV")); raw != "" {
		cfg.Env = Environment(raw)
		if cfg.Env != Development && cfg.Env != Production {
			problems = append(problems, fmt.Sprintf("APP_ENV %q inválido: use %s ou %s", raw, Development, Production))
		}
	}

	if raw := strings.TrimSpace(getenv("HTTP_PORT")); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil || port < 1 || port > 65535 {
			problems = append(problems, fmt.Sprintf("HTTP_PORT %q inválida: use um número entre 1 e 65535", raw))
		}
		cfg.HTTPPort = port
	}

	if len(problems) > 0 {
		return Config{}, errors.New("configuração inválida (veja .env.example):\n  - " + strings.Join(problems, "\n  - "))
	}
	return cfg, nil
}
