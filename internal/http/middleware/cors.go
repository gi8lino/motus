package middleware

import (
	"net/http"

	"github.com/gi8lino/motus/internal/http/response"
)

// CORS adds cross-origin response headers when an explicit allowed origin is configured.
func CORS(origin string) Middleware {
	return func(next http.Handler) http.Handler {
		if origin == "" {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions {
				response.Respond(w, http.StatusNoContent, nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
