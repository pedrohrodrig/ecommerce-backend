package database

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pingTimeout = 5 * time.Second

var errInvalidURL = errors.New("invalid database URL: use postgres://user:password@host:port/database " +
	"and percent-encode special characters in the user and password (e.g. @ as %40)")

const wsaECONNREFUSED syscall.Errno = 10061

func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil || hasStrayAt(url) {
		return nil, errInvalidURL
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		address := net.JoinHostPort(cfg.ConnConfig.Host, strconv.Itoa(int(cfg.ConnConfig.Port)))
		return nil, fmt.Errorf("database unreachable at %s: %s", address, connectFailureCause(err))
	}

	return pool, nil
}

func hasStrayAt(connString string) bool {
	rest, ok := strings.CutPrefix(connString, "postgresql://")
	if !ok {
		rest, ok = strings.CutPrefix(connString, "postgres://")
	}
	if !ok {
		return false
	}
	if i := strings.IndexAny(rest, "@/"); i >= 0 && rest[i] == '@' {
		rest = rest[i+1:]
	}
	return strings.Contains(rest, "@")
}

func connectFailureCause(err error) string {
	var pgErr *pgconn.PgError
	var dnsErr *net.DNSError
	var errno syscall.Errno

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Sprintf("no response within %s", pingTimeout)
	case errors.As(err, &pgErr):
		return describeServerError(pgErr.Code)
	case errors.As(err, &dnsErr):
		return "host not found"
	case errors.Is(err, syscall.ECONNREFUSED), errors.Is(err, wsaECONNREFUSED):
		return "connection refused"
	case errors.As(err, &errno):
		return errno.Error()
	default:
		return "connection failed"
	}
}

func describeServerError(sqlState string) string {
	switch sqlState {
	case "28P01":
		return "invalid user or password"
	case "3D000":
		return "database does not exist"
	default:
		return fmt.Sprintf("server rejected the connection (SQLSTATE %s)", sqlState)
	}
}
