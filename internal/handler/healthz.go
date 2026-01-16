package handler

import "net/http"

// Healthz responds with a simple health check.
func (a *API) Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.HealthStore.Ping(r.Context()); err != nil {
			a.Logger.Error("health check failed", "err", err)
			a.respondJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}

		a.Logger.Debug("health check")
		a.respondJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	}
}
