package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/config"
	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/database"
	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/httpserver"
	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/logging"
)

const serviceName = "ecommerce-backend"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, os.Getenv, os.Stdout)
	stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", serviceName, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getenv func(string) string, logOutput io.Writer) error {
	cfg, err := config.Load(getenv)
	if err != nil {
		return err
	}

	logger := logging.New(logOutput, cfg.Env, cfg.LogLevel)

	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	srv := httpserver.New(cfg.HTTPPort, httpserver.NewRouter(serviceName, pool, logger))

	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", srv.Addr, err)
	}

	logger.Info("server started", "url", "http://localhost"+srv.Addr+"/v1", "env", cfg.Env)

	if err := httpserver.Serve(ctx, srv, ln, logger); err != nil {
		return err
	}
	logger.Info("server stopped")
	return nil
}
