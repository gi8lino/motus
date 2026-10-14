package server

import (
	"net/http"

	appkitweb "github.com/containeroo/appkit/web"
)

// registerFallbackRoutes registers the SPA after every more-specific route.
func registerFallbackRoutes(mux *http.ServeMux, config Config) {
	mux.Handle("/", (appkitweb.SPA{Assets: config.Assets, BasePath: config.RoutePrefix}).Handler())
}
