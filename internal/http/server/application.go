package server

import (
	"net/http"

	"github.com/containeroo/httpkit/prefix"
	"github.com/gi8lino/motus/internal/http/middleware"
	"github.com/gi8lino/motus/internal/http/response"
)

// newApplication constructs public application routes and their middleware.
func newApplication(config Config) http.Handler {
	applicationMux := http.NewServeMux()
	registerPublicRoutes(applicationMux, config)
	registerAPIRoutes(applicationMux, config)
	registerFallbackRoutes(applicationMux, config)

	middlewares := []middleware.Middleware{
		middleware.SecurityHeaders(),
	}
	if config.AccessLog {
		middlewares = append(middlewares, middleware.AccessLog(config.Logger))
	}
	middlewares = append(
		middlewares,
		middleware.RecoverPanics(config.Logger),
		middleware.RejectCrossSiteWrites(config.Logger),
		middleware.CORS(""),
	)

	application := middleware.Chain(applicationMux, middlewares...)
	application = prefix.MountUnderPrefixWithOptions(
		application,
		config.RoutePrefix,
		prefix.WithRedirectRewriting(),
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		application.ServeHTTP(response.NewRequestWriter(w, r), r)
	})
}
