package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterRespondeEmV1(t *testing.T) {
	router := NewRouter("ecommerce-backend")

	for _, path := range []string{"/v1", "/v1/"} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, http.NoBody))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, quer %d", rec.Code, http.StatusOK)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, quer application/json", got)
			}
			if got, want := rec.Body.String(), `{"service":"ecommerce-backend"}`; got != want {
				t.Errorf("corpo = %s, quer %s", got, want)
			}
		})
	}
}

func TestRouterSoAtendeRotasVersionadas(t *testing.T) {
	router := NewRouter("ecommerce-backend")

	tests := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/", http.StatusNotFound},
		{http.MethodGet, "/v2", http.StatusNotFound},
		{http.MethodPost, "/v1", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), tt.method, tt.path, http.NoBody))

			if rec.Code != tt.want {
				t.Errorf("status = %d, quer %d", rec.Code, tt.want)
			}
		})
	}
}

func TestNewDefineEnderecoETimeouts(t *testing.T) {
	srv := New(9090, http.NotFoundHandler())

	if srv.Addr != ":9090" {
		t.Errorf("Addr = %q, quer :9090", srv.Addr)
	}
	if srv.ReadHeaderTimeout <= 0 || srv.ReadTimeout <= 0 || srv.WriteTimeout <= 0 || srv.IdleTimeout <= 0 {
		t.Errorf("servidor sem algum timeout: %+v", srv)
	}
}
