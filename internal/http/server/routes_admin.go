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
	mux.Handle("PUT /users/{id}/local-password", admin(endpoint.SetUserLocalPassword(config.Users, config.Logger)))
	mux.Handle("POST /users/{id}/identities", admin(endpoint.LinkUserIdentity(config.Users, config.Logger)))
	mux.Handle("DELETE /users/{id}/identities", admin(endpoint.UnlinkUserIdentity(config.Users, config.Logger)))
	mux.Handle("POST /users/{id}/sessions/revoke", admin(endpoint.RevokeUserSessions(config.Users, config.Logger)))
}
