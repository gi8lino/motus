package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	httpkitmiddleware "github.com/containeroo/httpkit/middleware"
	"github.com/gi8lino/motus/internal/http/response"
)

// RecoverPanics contains a handler panic and records it without leaking details to clients.
func RecoverPanics(logger *slog.Logger) func(http.Handler) http.Handler {
	return httpkitmiddleware.RecoverPanics(
		func(w http.ResponseWriter, _ *http.Request, recovered any) {
			response.InternalServerError(logger.With("operation", "request_panic"), w, fmt.Errorf("request panic: %v", recovered))
		},
		func(w http.ResponseWriter) bool {
			request, ok := w.(*response.RequestWriter)
			return ok && request.Status != 0
		},
	)
}
