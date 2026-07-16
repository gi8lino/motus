package middleware

import (
	"net/http"

	"github.com/gi8lino/motus/internal/auth"
)

// Authenticate resolves a principal once for protected API routes.
func Authenticate(store auth.Store, authHeader string, autoCreate bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/config" || r.URL.Path == "/login" || (r.URL.Path == "/users" && r.Method == http.MethodPost) {
				next.ServeHTTP(w, r)
				return
			}
			userID, err := auth.ResolveUserID(r, store, authHeader, autoCreate, "")
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"authentication required"}`))
				return
			}
			next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), userID)))
		})
	}
}
