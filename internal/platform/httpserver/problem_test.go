package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func assertProblem(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Errorf("status = %d, want %d", rec.Code, wantStatus)
	}
	if got := rec.Header().Get("Content-Type"); got != problemContentType {
		t.Errorf("Content-Type = %q, want %q", got, problemContentType)
	}

	var p problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("body is not JSON: %q", rec.Body.String())
	}
	if p.Status != wantStatus || p.Title != http.StatusText(wantStatus) {
		t.Errorf("problem = %+v, want status %d and title %q", p, wantStatus, http.StatusText(wantStatus))
	}
}

func TestWriteProblem(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		detail   string
		wantBody string
	}{
		{"without detail", http.StatusNotFound, "", `{"title":"Not Found","status":404}`},
		{"with detail", http.StatusConflict, "SKU already exists", `{"title":"Conflict","status":409,"detail":"SKU already exists"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			WriteProblem(rec, tt.status, tt.detail)

			assertProblem(t, rec, tt.status)
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("body = %s, want %s", got, tt.wantBody)
			}
		})
	}
}
