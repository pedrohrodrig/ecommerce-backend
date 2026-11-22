package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

const problemContentType = "application/problem+json"

type problem struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// WriteProblem writes an RFC 9457 Problem Details response. The detail is sent
// to the client, so it must never carry internal causes.
func WriteProblem(w http.ResponseWriter, status int, detail string) {
	body, _ := json.Marshal(problem{Title: http.StatusText(status), Status: status, Detail: detail})

	w.Header().Set("Content-Type", problemContentType)
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	WriteProblem(w, http.StatusNotFound, "")
}

var routableMethods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete,
}

// A custom 405 handler replaces chi's default one, which is the only place that
// sets the Allow header, so the allowed methods are looked up again here.
func methodNotAllowed(routes chi.Routes) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", strings.Join(allowedMethods(routes, r.URL.Path), ", "))
		WriteProblem(w, http.StatusMethodNotAllowed, "")
	}
}

// chi.Routes.Match accepts any method on the root of a mounted sub-router
// ("/v1" and "/v1/"), so those paths are resolved as "/" inside the sub-router.
func allowedMethods(routes chi.Routes, path string) []string {
	for _, route := range routes.Routes() {
		prefix := strings.TrimSuffix(route.Pattern, "/*")
		if route.SubRoutes != nil && (path == prefix || path == prefix+"/") {
			return allowedMethods(route.SubRoutes, "/")
		}
	}

	var allowed []string
	for _, method := range routableMethods {
		if routes.Match(chi.NewRouteContext(), method, path) {
			allowed = append(allowed, method)
		}
	}
	return allowed
}
