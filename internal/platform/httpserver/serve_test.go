package httpserver

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

func listen(t *testing.T) net.Listener {
	t.Helper()
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	return ln
}

func get(ctx context.Context, url string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func TestServeFinishesInFlightRequestsOnShutdown(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusNoContent)
	})

	ln := listen(t)
	url := "http://" + ln.Addr().String()
	ctx, shutdown := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() { served <- Serve(ctx, &http.Server{Handler: handler}, ln, slog.New(slog.DiscardHandler)) }()

	type result struct {
		status int
		err    error
	}
	inFlight := make(chan result, 1)
	go func() {
		status, err := get(t.Context(), url)
		inFlight <- result{status, err}
	}()

	<-started
	shutdown()

	select {
	case err := <-served:
		t.Fatalf("Serve returned (%v) before the in-flight request finished", err)
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	if res := <-inFlight; res.err != nil || res.status != http.StatusNoContent {
		t.Errorf("in-flight request: status = %d, err = %v, want %d", res.status, res.err, http.StatusNoContent)
	}
	if err := <-served; err != nil {
		t.Errorf("Serve() = %v, want nil", err)
	}
	if _, err := get(t.Context(), url); err == nil {
		t.Error("server still accepts requests after shutdown")
	}
}

func TestServeReturnsServerErrors(t *testing.T) {
	ln := listen(t)
	_ = ln.Close()

	err := Serve(t.Context(), &http.Server{Handler: http.NotFoundHandler()}, ln, slog.New(slog.DiscardHandler))
	if err == nil {
		t.Fatal("Serve() returned no error for a closed listener")
	}
}
