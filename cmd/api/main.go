package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/database"
)

const serviceName = "ecommerce-backend"

func main() {
	if err := run(context.Background(), os.Stdout, os.Getenv("DATABASE_URL")); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", serviceName, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer, databaseURL string) error {
	databaseURL = strings.TrimSpace(databaseURL)
	if databaseURL == "" {
		return errors.New("DATABASE_URL não definida")
	}

	pool, err := database.Open(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	_, err = fmt.Fprintf(out, "%s: api inicializada\n", serviceName)
	return err
}
