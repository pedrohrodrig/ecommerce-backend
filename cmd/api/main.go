package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/config"
	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/database"
	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/httpserver"
)

const serviceName = "ecommerce-backend"

func main() {
	if err := run(context.Background(), os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", serviceName, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getenv func(string) string, out io.Writer) error {
	cfg, err := config.Load(getenv)
	if err != nil {
		return err
	}

	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	srv := httpserver.New(cfg.HTTPPort, httpserver.NewRouter(serviceName))

	if _, err := fmt.Fprintf(out, "%s: listening on http://localhost%s/v1 (%s)\n", serviceName, srv.Addr, cfg.Env); err != nil {
		return err
	}
	return srv.ListenAndServe()
}
