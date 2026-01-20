package handler

import "net/http"

// Healthz responds with a simple health check.
func (a *API) Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.HealthStore.Ping(r.Context()); err != nil {
			a.businessLogger(r).Error(
				"health check failed",
				"event", "health_check_failed",
				"err", err,
			)
			a.respondJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		a.respondJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	}
}
