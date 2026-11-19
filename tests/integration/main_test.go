//go:build integration

// Package integration testa o backend contra um PostgreSQL real em container.
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/pedrohrodrig/ecommerce-backend/migrations"
)

const postgresImage = "postgres:18-alpine"

var (
	databaseURL string
	db          *sql.DB
	migrator    *goose.Provider
)

func TestMain(m *testing.M) {
	os.Exit(runWithPostgres(m))
}

func runWithPostgres(m *testing.M) int {
	ctx := context.Background()

	container, err := startPostgres(ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			fmt.Fprintf(os.Stderr, "encerrar o container do PostgreSQL: %v\n", err)
		}
	}()
	if err != nil {
		fmt.Fprintf(os.Stderr, "os testes de integração exigem o Docker rodando; não foi possível subir %s: %v\n", postgresImage, err)
		return 1
	}

	if err := migrate(ctx, container); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer db.Close()

	return m.Run()
}

func startPostgres(ctx context.Context) (container *postgres.PostgresContainer, err error) {
	// O testcontainers entra em pânico, em vez de devolver erro, quando não acha o Docker.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return postgres.Run(ctx, postgresImage, postgres.WithDatabase("ecommerce"), postgres.BasicWaitStrategies())
}

func migrate(ctx context.Context, container *postgres.PostgresContainer) error {
	url, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("obter a URL do PostgreSQL: %w", err)
	}

	conn, err := sql.Open("pgx", url)
	if err != nil {
		return fmt.Errorf("abrir conexão para as migrations: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, conn, migrations.FS)
	if err == nil {
		_, err = provider.Up(ctx)
	}
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("aplicar as migrations: %w", err)
	}

	databaseURL, db, migrator = url, conn, provider
	return nil
}
