package endpoint

import (
	"github.com/gi8lino/motus/internal/http/response"
	"net/http"

	"github.com/gi8lino/motus/internal/db"
)

// Health serves process liveness without depending on external services.
func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		response.Respond(w, http.StatusOK, statusResponse{Status: "ok"})
	}
}

// Ready serves application readiness and fails while a required dependency is unavailable.
func Ready(store db.HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.Ping(r.Context()); err != nil {
			response.Problem(w, http.StatusServiceUnavailable, "Database unavailable.")
			return
		}
		response.Respond(w, http.StatusOK, statusResponse{Status: "ok"})
	}
}
