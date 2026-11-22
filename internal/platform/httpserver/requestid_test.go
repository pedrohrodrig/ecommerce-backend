package httpserver

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

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

func TestRequestIDIsAvailableToHandlers(t *testing.T) {
	var seen string
	handler := assignRequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = RequestID(r.Context())
	}))

	rec := serve(t, handler, http.MethodGet, "/", nil)

	if seen == "" || seen != rec.Header().Get(requestIDHeader) {
		t.Errorf("RequestID() = %q, response header = %q", seen, rec.Header().Get(requestIDHeader))
	}
}

func TestRequestIDIsEmptyOutsideRequests(t *testing.T) {
	if got := RequestID(t.Context()); got != "" {
		t.Errorf("RequestID() = %q, want empty", got)
	}
}
