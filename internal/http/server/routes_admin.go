package server

import (
	"net/http"

	"github.com/gi8lino/motus/internal/http/endpoint"
	"github.com/gi8lino/motus/internal/http/middleware"
)

// registerAdminRoutes registers administrator-only API routes.
func registerAdminRoutes(mux *http.ServeMux, config Config) {
	admin := func(handler http.Handler) http.Handler {
		return middleware.Chain(handler, middleware.RequireAdmin(config.Users))
	}
	mux.Handle("GET /users", admin(endpoint.GetUsers(config.Users, config.Logger)))
	mux.Handle("PUT /users/{id}/admin", admin(endpoint.UpdateUserRole(config.Users, config.Logger)))
}
