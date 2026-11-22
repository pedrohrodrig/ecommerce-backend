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

	"github.com/go-chi/chi/v5"
)

func newPanickingRouter(logOutput io.Writer, handler http.HandlerFunc) http.Handler {
	logger := slog.New(slog.NewJSONHandler(logOutput, nil))
	r := chi.NewRouter()
	r.Use(assignRequestID, logRequests(logger, true), recoverPanics(logger))
	r.Get("/boom", handler)
	return r
}

func panicLog(t *testing.T, logs *bytes.Buffer) map[string]any {
	t.Helper()
	for line := range strings.SplitSeq(strings.TrimSpace(logs.String()), "\n") {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err == nil && entry["msg"] == "panic recovered" {
			return entry
		}
	}
	t.Fatalf("no panic log in: %s", logs.String())
	return nil
}

func assertAborted(t *testing.T, recovered any) {
	t.Helper()
	if err, _ := recovered.(error); !errors.Is(err, http.ErrAbortHandler) {
		t.Errorf("recovered %v, want http.ErrAbortHandler", recovered)
	}
}

func TestPanicBecomesProblemAndIsLoggedWithStack(t *testing.T) {
	var logs bytes.Buffer
	router := newPanickingRouter(&logs, func(http.ResponseWriter, *http.Request) {
		panic("secret internal state")
	})

	rec := serve(t, router, http.MethodGet, "/boom", nil)

	assertProblem(t, rec, http.StatusInternalServerError)
	if strings.Contains(rec.Body.String(), "secret") {
		t.Errorf("response exposes the panic value: %s", rec.Body.String())
	}

	entry := panicLog(t, &logs)
	if entry["level"] != "ERROR" || entry["panic"] != "secret internal state" || entry["route"] != "/boom" {
		t.Errorf("unexpected panic log: %v", entry)
	}
	if stack, _ := entry["stack"].(string); !strings.Contains(stack, "recover_test.go") {
		t.Errorf("stack does not point to the panicking handler: %q", stack)
	}
	if entry["request_id"] != rec.Header().Get(requestIDHeader) {
		t.Errorf("request_id = %v, response header = %q", entry["request_id"], rec.Header().Get(requestIDHeader))
	}
	if req := lastRequestLog(t, &logs); req["status"] != float64(http.StatusInternalServerError) {
		t.Errorf("request log status = %v, want 500", req["status"])
	}
}

func TestPanicAfterResponseStartedAbortsConnection(t *testing.T) {
	var logs bytes.Buffer
	router := newPanickingRouter(&logs, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		panic(errors.New("failed mid-response"))
	})

	defer func() {
		assertAborted(t, recover())
		panicLog(t, &logs)
	}()
	serve(t, router, http.MethodGet, "/boom", nil)
	t.Error("panic was swallowed after the response started")
}

func TestAbortHandlerPanicPassesThrough(t *testing.T) {
	var logs bytes.Buffer
	router := newPanickingRouter(&logs, func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	})

	defer func() {
		assertAborted(t, recover())
		if strings.Contains(logs.String(), "panic recovered") {
			t.Errorf("intentional abort was logged as a panic: %s", logs.String())
		}
	}()
	serve(t, router, http.MethodGet, "/boom", nil)
}
