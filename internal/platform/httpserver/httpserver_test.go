package httpserver

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testOrigin = "https://loja.example.com"

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func newTestRouter(t *testing.T, pingErr error, logOutput io.Writer) http.Handler {
	t.Helper()
	return NewRouter(RouterConfig{
		ServiceName: "ecommerce-backend",
		DB:          fakePinger{pingErr},
		Logger:      slog.New(slog.NewJSONHandler(logOutput, nil)),
		LogProbes:   true,
		CORSOrigins: []string{testOrigin},
	})
}

func serve(t *testing.T, router http.Handler, method, path string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), method, path, http.NoBody)
	for key, values := range header {
		req.Header[key] = values
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestRouterServesV1(t *testing.T) {
	router := newTestRouter(t, nil, io.Discard)

	for _, path := range []string{"/v1", "/v1/"} {
		t.Run(path, func(t *testing.T) {
			rec := serve(t, router, http.MethodGet, path, nil)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			if got, want := rec.Body.String(), `{"service":"ecommerce-backend"}`; got != want {
				t.Errorf("body = %s, want %s", got, want)
			}
		})
	}
}

func TestRouterOnlyServesVersionedRoutes(t *testing.T) {
	router := newTestRouter(t, nil, io.Discard)

	tests := []struct {
		method    string
		path      string
		want      int
		wantAllow string
	}{
		{http.MethodGet, "/", http.StatusNotFound, ""},
		{http.MethodGet, "/v2", http.StatusNotFound, ""},
		{http.MethodGet, "/v1/unknown", http.StatusNotFound, ""},
		{http.MethodPost, "/v1", http.StatusMethodNotAllowed, "GET"},
		{http.MethodPut, "/v1/", http.StatusMethodNotAllowed, "GET"},
		{http.MethodDelete, "/healthz", http.StatusMethodNotAllowed, "GET"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := serve(t, router, tt.method, tt.path, nil)

			assertProblem(t, rec, tt.want)
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("Allow = %q, want %q", got, tt.wantAllow)
			}
		})
	}
}

func TestNewSetsAddressAndTimeouts(t *testing.T) {
	srv := New(9090, http.NotFoundHandler())

	if srv.Addr != ":9090" {
		t.Errorf("Addr = %q, want :9090", srv.Addr)
	}
	if srv.ReadHeaderTimeout <= 0 || srv.ReadTimeout <= 0 || srv.WriteTimeout <= 0 || srv.IdleTimeout <= 0 {
		t.Errorf("server is missing a timeout: %+v", srv)
	}
}
