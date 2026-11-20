package httpserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

func lastRequestLog(t *testing.T, logs *bytes.Buffer) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(logs.String()), "\n")

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("log line is not JSON: %q", logs.String())
	}
	if entry["msg"] != "request" {
		t.Fatalf("last log line is not a request log: %v", entry)
	}
	return entry
}

func TestRequestLogHasRequiredFields(t *testing.T) {
	tests := []struct {
		method     string
		path       string
		wantRoute  string
		wantStatus float64
	}{
		{http.MethodGet, "/v1/", "/v1", http.StatusOK},
		{http.MethodGet, "/readyz", "/readyz", http.StatusOK},
		{http.MethodPost, "/v1", "/v1", http.StatusMethodNotAllowed},
		{http.MethodGet, "/unknown", "", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			var logs bytes.Buffer
			rec := serve(t, newTestRouter(t, nil, &logs), tt.method, tt.path, nil)
			entry := lastRequestLog(t, &logs)

			if entry["method"] != tt.method || entry["route"] != tt.wantRoute || entry["path"] != tt.path || entry["status"] != tt.wantStatus {
				t.Errorf("unexpected request log: %v", entry)
			}
			if d, ok := entry["duration_ms"].(float64); !ok || d < 0 {
				t.Errorf("duration_ms = %v, want a non-negative number", entry["duration_ms"])
			}
			if id := rec.Header().Get(requestIDHeader); id == "" || entry["request_id"] != id {
				t.Errorf("request_id = %v, response header = %q", entry["request_id"], id)
			}
		})
	}
}

func TestRequestIDReusesValidClientHeader(t *testing.T) {
	tests := []struct {
		name     string
		incoming string
		wantSame bool
	}{
		{"valid id", "abc-123_X.y", true},
		{"missing", "", false},
		{"invalid characters", "abc 123\n", false},
		{"too long", strings.Repeat("a", 65), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newTestRouter(t, nil, io.Discard), http.MethodGet, "/healthz", http.Header{requestIDHeader: {tt.incoming}})
			got := rec.Header().Get(requestIDHeader)

			if (got == tt.incoming) != tt.wantSame {
				t.Errorf("response request id = %q, incoming = %q, want reused = %v", got, tt.incoming, tt.wantSame)
			}
			if !validRequestID.MatchString(got) {
				t.Errorf("response request id %q is not valid", got)
			}
		})
	}
}

func TestGeneratedRequestIDsAreUnique(t *testing.T) {
	router := newTestRouter(t, nil, io.Discard)
	first := serve(t, router, http.MethodGet, "/healthz", nil).Header().Get(requestIDHeader)
	second := serve(t, router, http.MethodGet, "/healthz", nil).Header().Get(requestIDHeader)

	if first == second {
		t.Errorf("two requests got the same id %q", first)
	}
}

func TestRequestLogCanOmitProbes(t *testing.T) {
	tests := []struct {
		path    string
		wantLog bool
	}{
		{"/healthz", false},
		{"/readyz", false},
		{"/v1", true},
		{"/unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			var logs bytes.Buffer
			router := NewRouter(RouterConfig{
				ServiceName: "ecommerce-backend",
				DB:          fakePinger{},
				Logger:      slog.New(slog.NewJSONHandler(&logs, nil)),
			})
			rec := serve(t, router, http.MethodGet, tt.path, nil)

			if got := strings.Contains(logs.String(), `"msg":"request"`); got != tt.wantLog {
				t.Errorf("request logged = %v, want %v (logs: %s)", got, tt.wantLog, logs.String())
			}
			if rec.Header().Get(requestIDHeader) == "" {
				t.Error("response has no request id")
			}
		})
	}
}

func TestReadinessFailureIsLoggedEvenWithoutProbeLogs(t *testing.T) {
	var logs bytes.Buffer
	router := NewRouter(RouterConfig{
		ServiceName: "ecommerce-backend",
		DB:          fakePinger{errors.New("connection refused")},
		Logger:      slog.New(slog.NewJSONHandler(&logs, nil)),
	})
	serve(t, router, http.MethodGet, "/readyz", nil)

	if !strings.Contains(logs.String(), `"msg":"readiness check failed"`) {
		t.Errorf("readiness failure was not logged: %s", logs.String())
	}
}
