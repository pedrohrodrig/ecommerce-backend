package httpserver

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHealthRoutes(t *testing.T) {
	dbDown := errors.New("dial tcp 10.0.0.5:5432: connection refused")

	tests := []struct {
		name       string
		path       string
		pingErr    error
		wantStatus int
		wantBody   string
	}{
		{"liveness ignores the database", "/healthz", dbDown, http.StatusOK, `{"status":"ok"}`},
		{"readiness with database up", "/readyz", nil, http.StatusOK, `{"status":"ok"}`},
		{"readiness with database down", "/readyz", dbDown, http.StatusServiceUnavailable, `{"status":"unavailable"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newTestRouter(t, tt.pingErr, io.Discard), http.MethodGet, tt.path, nil)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("body = %s, want %s", got, tt.wantBody)
			}
		})
	}
}

func TestReadinessLogsCauseWithoutExposingIt(t *testing.T) {
	var logs bytes.Buffer
	rec := serve(t, newTestRouter(t, errors.New("connection refused"), &logs), http.MethodGet, "/readyz", nil)

	if strings.Contains(rec.Body.String(), "connection refused") {
		t.Errorf("response exposes the failure cause: %s", rec.Body.String())
	}
	if !strings.Contains(logs.String(), `"msg":"readiness check failed"`) || !strings.Contains(logs.String(), "connection refused") {
		t.Errorf("failure cause was not logged: %s", logs.String())
	}
}
