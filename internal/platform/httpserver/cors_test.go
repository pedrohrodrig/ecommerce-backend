package httpserver

import (
	"io"
	"log/slog"
	"net/http"
	"testing"
)

func TestCORSAllowsOnlyConfiguredOrigins(t *testing.T) {
	tests := []struct {
		name       string
		origin     string
		wantOrigin string
	}{
		{"configured origin", testOrigin, testOrigin},
		{"configured origin in another case", "https://LOJA.example.com", "https://LOJA.example.com"},
		{"unknown origin", "https://evil.example.com", ""},
		{"subdomain of configured origin", "https://x.loja.example.com", ""},
		{"same host over http", "http://loja.example.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newTestRouter(t, nil, io.Discard), http.MethodGet, "/v1", http.Header{"Origin": {tt.origin}})

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantOrigin)
			}
			wantExposed := ""
			if tt.wantOrigin != "" {
				wantExposed = requestIDHeader
			}
			if got := rec.Header().Get("Access-Control-Expose-Headers"); got != wantExposed {
				t.Errorf("Access-Control-Expose-Headers = %q, want %q", got, wantExposed)
			}
		})
	}
}

func TestCORSPreflight(t *testing.T) {
	tests := []struct {
		name       string
		origin     string
		method     string
		wantOrigin string
	}{
		{"allowed origin and method", testOrigin, http.MethodPatch, testOrigin},
		{"unknown origin", "https://evil.example.com", http.MethodPatch, ""},
		{"method not allowed", testOrigin, "TRACE", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newTestRouter(t, nil, io.Discard), http.MethodOptions, "/v1", http.Header{
				"Origin":                         {tt.origin},
				"Access-Control-Request-Method":  {tt.method},
				"Access-Control-Request-Headers": {"Content-Type, X-Request-Id"},
			})

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantOrigin)
			}
			if tt.wantOrigin != "" && rec.Header().Get("Access-Control-Max-Age") != "600" {
				t.Errorf("Access-Control-Max-Age = %q, want 600", rec.Header().Get("Access-Control-Max-Age"))
			}
		})
	}
}

func TestCORSDisabledWithoutOrigins(t *testing.T) {
	router := NewRouter(RouterConfig{ServiceName: "ecommerce-backend", DB: fakePinger{}, Logger: slog.New(slog.DiscardHandler)})

	rec := serve(t, router, http.MethodGet, "/v1", http.Header{"Origin": {testOrigin}})

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want none", got)
	}
}
