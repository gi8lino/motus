package middleware

import (
	"log/slog"
	"net/http"

	httpkitmiddleware "github.com/containeroo/httpkit/middleware"
	"github.com/gi8lino/motus/internal/http/response"
)

// SecurityHeaders adds browser-hardening headers that do not constrain Motus assets.
func SecurityHeaders() Middleware {
	return httpkitmiddleware.SecurityHeaders(http.Header{
		"X-Content-Type-Options": {"nosniff"},
		"X-Frame-Options":        {"DENY"},
		"Referrer-Policy":        {"same-origin"},
	})
}

// RejectCrossSiteWrites blocks browser requests that would otherwise send a Motus session cross-origin.
func RejectCrossSiteWrites(logger *slog.Logger) Middleware {
	return httpkitmiddleware.RejectCrossOriginWrites(logger, func(w http.ResponseWriter, _ *http.Request) {
		response.Problem(w, http.StatusForbidden, "Forbidden.")
	})
}
