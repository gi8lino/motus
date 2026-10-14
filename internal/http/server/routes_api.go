package server

import (
	"net/http"

	"github.com/gi8lino/motus/internal/http/middleware"
)

// registerAPIRoutes registers authenticated Motus API routes.
func registerAPIRoutes(mux *http.ServeMux, config Config) {
	routes := http.NewServeMux()
	registerBrowserRoutes(routes, config)
	registerAdminRoutes(routes, config)
	mux.Handle("/api/", http.StripPrefix("/api", middleware.Chain(routes, middleware.Authenticate(config.Browser.Authenticator))))
}
