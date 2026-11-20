package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

const defaultHTTPPort = 8080

var logLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

type Config struct {
	Env         Environment
	HTTPPort    int
	DatabaseURL string
	LogLevel    slog.Level
	LogProbes   bool
}

func Load(getenv func(string) string) (Config, error) {
	var problems []string

	cfg := Config{
		Env:         Development,
		HTTPPort:    defaultHTTPPort,
		LogLevel:    slog.LevelInfo,
		LogProbes:   true,
		DatabaseURL: strings.TrimSpace(getenv("DATABASE_URL")),
	}

	if cfg.DatabaseURL == "" {
		problems = append(problems, "DATABASE_URL is not set")
	}

	if raw := strings.TrimSpace(getenv("APP_ENV")); raw != "" {
		cfg.Env = Environment(raw)
		if cfg.Env != Development && cfg.Env != Production {
			problems = append(problems, fmt.Sprintf("APP_ENV %q is invalid: use %s or %s", raw, Development, Production))
		}
	}

	if raw := strings.TrimSpace(getenv("HTTP_PORT")); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil || port < 1 || port > 65535 {
			problems = append(problems, fmt.Sprintf("HTTP_PORT %q is invalid: use a number between 1 and 65535", raw))
		}
		cfg.HTTPPort = port
	}

	if raw := strings.TrimSpace(getenv("LOG_LEVEL")); raw != "" {
		level, ok := logLevels[strings.ToLower(raw)]
		if !ok {
			problems = append(problems, fmt.Sprintf("LOG_LEVEL %q is invalid: use debug, info, warn or error", raw))
		}
		cfg.LogLevel = level
	}

	if raw := strings.TrimSpace(getenv("LOG_PROBES")); raw != "" {
		logProbes, err := strconv.ParseBool(raw)
		if err != nil {
			problems = append(problems, fmt.Sprintf("LOG_PROBES %q is invalid: use true or false", raw))
		}
		cfg.LogProbes = logProbes
	}

	if len(problems) > 0 {
		return Config{}, errors.New("invalid configuration (see .env.example):\n  - " + strings.Join(problems, "\n  - "))
	}
	return cfg, nil
}
