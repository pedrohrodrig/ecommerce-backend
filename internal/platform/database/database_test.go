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

var secrets = []string{"s3cr3tPw", "Abc", "Def123", "12345", "ZqXsecret7", "Pw9q", "ZqXss", "Wq9rd"}

func assertNoSecret(t *testing.T, msg string) {
	t.Helper()
	for _, secret := range secrets {
		if strings.Contains(msg, secret) {
			t.Errorf("error %q exposes %q", msg, secret)
		}
	}
}

func TestOpenFailsWithoutExposingPassword(t *testing.T) {
	original := pingTimeout
	pingTimeout = 500 * time.Millisecond
	t.Cleanup(func() { pingTimeout = original })

	t.Setenv("PGHOST", "localhost")
	t.Setenv("PGPORT", "1")

	silentAddr := startSilentServer(t)

	tests := []struct {
		name string
		url  string
		want string
	}{
		{"parse: invalid port", "postgres://u:s3cr3tPw@localhost:porta/db", "invalid database URL"},
		{"parse: password with spaces in keyword/value", "host=localhost password = s3cr3tPw port=abc", "invalid database URL"},
		{"parse: unencoded @ in password", "postgres://u:Abc@Def123@localhost:abc/db", "invalid database URL"},
		{"parse: password only", "s3cr3tPw", "invalid database URL"},
		{"parse: unencoded @ with valid port", "postgres://u:Abc@Def123@localhost:1/db", "invalid database URL"},
		{"parse: unencoded / in password", "postgres://u:12345/ZqXsecret7@localhost:1/db", "invalid database URL"},
		{"parse: unencoded @ and / in password", "postgres://u:pa@ZqXss/Wq9rd@localhost:1/db", "invalid database URL"},
		{"connect: closed port", "postgres://u:s3cr3tPw@localhost:1/db", "database unreachable at localhost:1: connection refused"},
		{"connect: password with spaces in keyword/value", "host=localhost password = s3cr3tPw port=1", "database unreachable at localhost:1: "},
		{"connect: unencoded ? in password", "postgres://u:12?Pw9q@localhost:1/db", "database unreachable at localhost:1: "},
		{"connect: unencoded # in password", "postgres://u:12#Pw9q@localhost:1/db", "database unreachable at localhost:1: "},
		{"connect: variable name in value", "DATABASE_URL=s3cr3tPw", "database unreachable at localhost:1: "},
		{"connect: no user", "postgres://localhost:1/db", "database unreachable at localhost:1: "},
		{"connect: IPv6 host", "postgres://u:s3cr3tPw@[::1]:1/db", "database unreachable at [::1]:1: "},
		{"connect: multiple hosts", "postgres://u:s3cr3tPw@localhost:1,127.0.0.1:1/db", "database unreachable at localhost:1: "},
		{"connect: encoded @ in password", "postgres://u:Abc%40Def123@localhost:1/db", "database unreachable at localhost:1: connection refused"},
		{"connect: no response", "postgres://u:s3cr3tPw@" + silentAddr + "/db", "database unreachable at " + silentAddr + ": no response within 500ms"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := Open(t.Context(), tt.url)
			if err == nil {
				pool.Close()
				t.Fatal("Open() returned no error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q does not contain %q", err, tt.want)
			}
			assertNoSecret(t, err.Error())
		})
	}
}

func TestConnectFailureCauseDoesNotLeakURLText(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"timeout", context.DeadlineExceeded, "no response within"},
		{"wrong password", &pgconn.PgError{Code: "28P01", Message: `password authentication failed for user "s3cr3tPw"`}, "invalid user or password"},
		{"denied by pg_hba", &pgconn.PgError{Code: "28000", Message: `no pg_hba.conf entry for user "s3cr3tPw"`}, "server rejected the connection (SQLSTATE 28000)"},
		{"missing database", &pgconn.PgError{Code: "3D000", Message: `database "s3cr3tPw" does not exist`}, "database does not exist"},
		{"other server error", &pgconn.PgError{Code: "53300", Message: `too many connections for role "s3cr3tPw"`}, "server rejected the connection (SQLSTATE 53300)"},
		{"host not found", &net.DNSError{Err: "no such host", Name: "Def123@localhost", IsNotFound: true}, "host not found"},
		{"connection refused", &net.OpError{Op: "dial", Addr: &net.TCPAddr{}, Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}, "connection refused"},
		{"connection refused on Windows", &net.OpError{Op: "dial", Addr: &net.TCPAddr{}, Err: os.NewSyscallError("connectex", wsaECONNREFUSED)}, "connection refused"},
		{"system error", &net.OpError{Op: "dial", Addr: &net.TCPAddr{}, Err: syscall.Errno(1)}, syscall.Errno(1).Error()},
		{"unknown", errors.New("failed with s3cr3tPw"), "connection failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := connectFailureCause(tt.err)
			if !strings.Contains(got, tt.want) {
				t.Errorf("connectFailureCause() = %q, want it to contain %q", got, tt.want)
			}
			assertNoSecret(t, got)
		})
	}
}

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
