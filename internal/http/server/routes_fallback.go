package server

import (
	"net/http"

	"github.com/gi8lino/motus/internal/http/endpoint"
)

// registerFallbackRoutes registers the SPA after every more-specific route.
func registerFallbackRoutes(mux *http.ServeMux, config Config) {
	mux.Handle("/", endpoint.SPA(config.Assets, config.RoutePrefix))
}
