//go:build integration

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
			fmt.Fprintf(os.Stderr, "terminate PostgreSQL container: %v\n", err)
		}
	}()
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration tests require Docker to be running; could not start %s: %v\n", postgresImage, err)
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
		return fmt.Errorf("get PostgreSQL URL: %w", err)
	}

	conn, err := sql.Open("pgx", url)
	if err != nil {
		return fmt.Errorf("open connection for migrations: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, conn, migrations.FS)
	if err == nil {
		_, err = provider.Up(ctx)
	}
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("apply migrations: %w", err)
	}

	databaseURL, db, migrator = url, conn, provider
	return nil
}
