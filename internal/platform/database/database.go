// Package database abre o pool de conexões com o PostgreSQL.
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

// pingTimeout é variável só para os testes encurtarem a espera.
var pingTimeout = 5 * time.Second

// errInvalidURL é fixo porque o texto de erro do pgx ecoa trechos da URL,
// e o mascaramento de senha dele falha em URLs malformadas.
var errInvalidURL = errors.New("URL do banco inválida: use postgres://usuario:senha@host:porta/banco " +
	"e codifique caracteres especiais do usuário e da senha (ex.: @ como %40)")

// Open cria o pool de conexões a partir da URL e confirma que o banco responde.
// As mensagens de erro nunca incluem texto da URL além de host e porta.
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil || hasStrayAt(url) {
		return nil, errInvalidURL
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("criar pool de conexões: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		address := net.JoinHostPort(cfg.ConnConfig.Host, strconv.Itoa(int(cfg.ConnConfig.Port)))
		return nil, fmt.Errorf("banco de dados inacessível em %s: %s", address, connectFailureCause(err))
	}

	return pool, nil
}

// hasStrayAt informa se sobra um @ depois do usuário e da senha da URL. Como o
// libpq, o pgx encerra a senha no primeiro @ antes de qualquer /, então um / ou
// @ sem codificar na senha joga o resto dela para host, porta ou banco.
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

// connectFailureCause descreve a falha de conexão sem repassar o texto do pgx
// nem a mensagem do servidor, que incluem usuário, banco e host vindos da URL.
func connectFailureCause(err error) string {
	var pgErr *pgconn.PgError
	var dnsErr *net.DNSError
	var errno syscall.Errno

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Sprintf("sem resposta em %s", pingTimeout)
	case errors.As(err, &pgErr):
		return describeServerError(pgErr.Code)
	case errors.As(err, &dnsErr):
		return "host não encontrado"
	case errors.As(err, &errno):
		return errno.Error()
	default:
		// shortcut: causas sem classe (ex.: TLS) viram genéricas; classifique quando aparecerem.
		return "falha ao conectar"
	}
}

func describeServerError(sqlState string) string {
	switch sqlState {
	case "28P01", "28000":
		return "usuário ou senha inválidos"
	case "3D000":
		return "o banco de dados não existe"
	default:
		return fmt.Sprintf("o servidor recusou a conexão (SQLSTATE %s)", sqlState)
	}
}
