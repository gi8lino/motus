package server

import (
	"net/http"

	httpkitmiddleware "github.com/containeroo/httpkit/middleware"
	"github.com/gi8lino/motus/internal/http/middleware"
)

// registerAPIRoutes registers authenticated Motus API routes.
func registerAPIRoutes(mux *http.ServeMux, config Config) {
	routes := http.NewServeMux()
	registerBrowserRoutes(routes, config)
	registerAdminRoutes(routes, config)
	mux.Handle("/api/", http.StripPrefix("/api", httpkitmiddleware.Chain(routes, middleware.Authenticate(config.Browser.Authenticator))))
}
