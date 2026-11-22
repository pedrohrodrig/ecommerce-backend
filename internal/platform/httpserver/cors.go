package httpserver

import (
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/cors"
)

const corsMaxAgeSeconds = 600

// Origins are matched exactly through AllowOriginFunc: go-chi/cors treats an
// empty AllowedOrigins as "allow all" and expands "*" patterns.
func allowOrigins(origins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowOriginFunc: func(_ *http.Request, origin string) bool {
			return slices.Contains(origins, strings.ToLower(origin))
		},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", requestIDHeader},
		ExposedHeaders: []string{requestIDHeader},
		MaxAge:         corsMaxAgeSeconds,
	})
}
