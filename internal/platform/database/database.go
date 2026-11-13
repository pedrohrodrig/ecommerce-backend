// Package database abre o pool de conexões com o PostgreSQL.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const pingTimeout = 5 * time.Second

// Open cria o pool de conexões a partir da URL e confirma que o banco responde.
// As mensagens de erro nunca incluem a URL, que pode conter a senha.
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL inválida: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("banco de dados inacessível: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("banco de dados inacessível: %w", err)
	}

	return pool, nil
}
