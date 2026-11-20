package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunFailsBeforeStartingServer(t *testing.T) {
	tests := []struct {
		name string
		vars map[string]string
		want string
	}{
		{"DATABASE_URL missing", map[string]string{}, "DATABASE_URL is not set"},
		{"HTTP_PORT invalid", map[string]string{"DATABASE_URL": "postgres://localhost/db", "HTTP_PORT": "x"}, `HTTP_PORT "x" is invalid`},
		{"database unreachable", map[string]string{"DATABASE_URL": "postgres://u:s3cr3tPw@localhost:1/db"}, "database unreachable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			getenv := func(key string) string { return tt.vars[key] }

			err := run(t.Context(), getenv, &out)
			if err == nil {
				t.Fatal("run() returned no error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q does not contain %q", err, tt.want)
			}
			if out.Len() != 0 {
				t.Errorf("run() wrote %q before starting the server", out.String())
			}
		})
	}
}
