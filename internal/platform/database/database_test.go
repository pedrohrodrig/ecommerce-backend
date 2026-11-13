package database

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// secrets são trechos das senhas usadas nos casos; nenhum pode aparecer num erro.
var secrets = []string{"segredo", "Abc", "Def123", "12345", "ZqXsecret7", "Senha9", "ZqXss", "word"}

func assertNoSecret(t *testing.T, msg string) {
	t.Helper()
	for _, secret := range secrets {
		if strings.Contains(msg, secret) {
			t.Errorf("erro %q expõe %q", msg, secret)
		}
	}
}

func TestOpenFalhaSemExporSenha(t *testing.T) {
	original := pingTimeout
	pingTimeout = 500 * time.Millisecond
	t.Cleanup(func() { pingTimeout = original })

	// Strings keyword/value sem host e porta usam PGHOST/PGPORT; a porta 1 recusa conexões.
	t.Setenv("PGHOST", "localhost")
	t.Setenv("PGPORT", "1")

	silentAddr := startSilentServer(t)

	tests := []struct {
		name string
		url  string
		want string
	}{
		{"parse: porta inválida", "postgres://u:segredo@localhost:porta/db", "URL do banco inválida"},
		{"parse: senha com espaços em keyword/value", "host=localhost password = segredo port=abc", "URL do banco inválida"},
		{"parse: @ não codificado na senha", "postgres://u:Abc@Def123@localhost:abc/db", "URL do banco inválida"},
		{"parse: só a senha", "segredo", "URL do banco inválida"},
		{"parse: @ não codificado com porta válida", "postgres://u:Abc@Def123@localhost:1/db", "URL do banco inválida"},
		{"parse: / não codificado na senha", "postgres://u:12345/ZqXsecret7@localhost:1/db", "URL do banco inválida"},
		{"parse: @ e / não codificados na senha", "postgres://u:pa@ZqXss/word@localhost:1/db", "URL do banco inválida"},
		{"conexão: porta fechada", "postgres://u:segredo@localhost:1/db", "banco de dados inacessível em localhost:1: conexão recusada"},
		{"conexão: senha com espaços em keyword/value", "host=localhost password = segredo port=1", "banco de dados inacessível em localhost:1: "},
		// Como no libpq, ? e # antes do @ ficam na senha; a URL é aceita.
		{"conexão: ? não codificado na senha", "postgres://u:12?Senha9@localhost:1/db", "banco de dados inacessível em localhost:1: "},
		{"conexão: # não codificado na senha", "postgres://u:12#Senha9@localhost:1/db", "banco de dados inacessível em localhost:1: "},
		{"conexão: nome da variável no valor", "DATABASE_URL=segredo", "banco de dados inacessível em localhost:1: "},
		{"conexão: sem resposta", "postgres://u:segredo@" + silentAddr + "/db", "banco de dados inacessível em " + silentAddr + ": sem resposta em 500ms"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := Open(t.Context(), tt.url)
			if err == nil {
				pool.Close()
				t.Fatal("Open() não retornou erro")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("erro %q não contém %q", err, tt.want)
			}
			assertNoSecret(t, err.Error())
		})
	}
}

func TestConnectFailureCauseNaoRepassaTextoDaURL(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"timeout", context.DeadlineExceeded, "sem resposta em"},
		{"senha errada", &pgconn.PgError{Code: "28P01", Message: `password authentication failed for user "segredo"`}, "usuário ou senha inválidos"},
		{"acesso negado pelo pg_hba", &pgconn.PgError{Code: "28000", Message: `no pg_hba.conf entry for user "segredo"`}, "o servidor recusou a conexão (SQLSTATE 28000)"},
		{"banco inexistente", &pgconn.PgError{Code: "3D000", Message: `database "segredo" does not exist`}, "o banco de dados não existe"},
		{"outro erro do servidor", &pgconn.PgError{Code: "53300", Message: `too many connections for role "segredo"`}, "o servidor recusou a conexão (SQLSTATE 53300)"},
		{"host não encontrado", &net.DNSError{Err: "no such host", Name: "Def123@localhost", IsNotFound: true}, "host não encontrado"},
		{"conexão recusada", &net.OpError{Op: "dial", Addr: &net.TCPAddr{}, Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}, "conexão recusada"},
		{"conexão recusada no Windows", &net.OpError{Op: "dial", Addr: &net.TCPAddr{}, Err: os.NewSyscallError("connectex", wsaECONNREFUSED)}, "conexão recusada"},
		{"erro do sistema", &net.OpError{Op: "dial", Addr: &net.TCPAddr{}, Err: syscall.Errno(1)}, syscall.Errno(1).Error()},
		{"desconhecido", errors.New("falhou com segredo"), "falha ao conectar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := connectFailureCause(tt.err)
			if !strings.Contains(got, tt.want) {
				t.Errorf("connectFailureCause() = %q; esperado conter %q", got, tt.want)
			}
			assertNoSecret(t, got)
		})
	}
}

// startSilentServer aceita conexões TCP e nunca responde, como um banco travado.
func startSilentServer(t *testing.T) string {
	t.Helper()

	var lc net.ListenConfig
	listener, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		var conns []net.Conn
		for {
			conn, err := listener.Accept()
			if err != nil {
				for _, c := range conns {
					_ = c.Close()
				}
				return
			}
			conns = append(conns, conn)
		}
	}()

	return listener.Addr().String()
}
