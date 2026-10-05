package server

import (
	"net/http"
	"strings"

	"github.com/gi8lino/motus/internal/http/endpoint"
)

// newManagement constructs operational routes outside application middleware.
func newManagement(config Config) http.Handler {
	mux := http.NewServeMux()
	registerOperationalRoutes(mux, config)
	return mux
}

// registerOperationalRoutes registers operational routes outside application middleware.
func registerOperationalRoutes(mux *http.ServeMux, config Config) {
	mux.Handle("GET /healthz", endpoint.Health())
	mux.Handle("/healthz", methodNotAllowed(http.MethodGet, http.MethodHead))
	mux.Handle("GET /readyz", endpoint.Ready(config.HealthStore))
	mux.Handle("/readyz", methodNotAllowed(http.MethodGet, http.MethodHead))
}

// methodNotAllowed rejects unsupported methods on operational endpoints.
func methodNotAllowed(methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", strings.Join(methods, ", "))
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	})
}
