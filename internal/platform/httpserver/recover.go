package httpserver

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func recoverPanics(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(rec)
				}

				logger.LogAttrs(r.Context(), slog.LevelError, "panic recovered",
					slog.String("panic", fmt.Sprint(rec)),
					slog.String("stack", string(debug.Stack())),
					slog.String("route", chi.RouteContext(r.Context()).RoutePattern()),
					slog.String("request_id", RequestID(r.Context())),
				)

				// Once the status line is out, a 500 can no longer be sent; aborting
				// drops the connection so the client does not take a partial body as success.
				if ww.Status() != 0 {
					panic(http.ErrAbortHandler)
				}
				WriteProblem(ww, http.StatusInternalServerError, "")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
