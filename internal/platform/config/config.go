package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
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
	CORSOrigins []string
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

	if raw := strings.TrimSpace(getenv("CORS_ALLOWED_ORIGINS")); raw != "" {
		for origin := range strings.SplitSeq(raw, ",") {
			origin = strings.TrimSpace(origin)
			if !isOrigin(origin) {
				problems = append(problems, fmt.Sprintf("CORS_ALLOWED_ORIGINS entry %q is invalid: use scheme://host[:port], e.g. https://loja.example.com", origin))
				continue
			}
			cfg.CORSOrigins = append(cfg.CORSOrigins, strings.ToLower(origin))
		}
	}

	if len(problems) > 0 {
		return Config{}, errors.New("invalid configuration (see .env.example):\n  - " + strings.Join(problems, "\n  - "))
	}
	return cfg, nil
}

func isOrigin(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || strings.Contains(u.Host, "*") {
		return false
	}
	return strings.EqualFold(u.Scheme+"://"+u.Host, raw)
}
