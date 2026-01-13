package handler

import "net/http"

// Healthz responds with a simple health check.
func (a *API) Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.HealthStore.Ping(r.Context()); err != nil {
			if err := encode(w, r, http.StatusInternalServerError, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("healthz encode", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, statusResponse{Status: "ok"}); err != nil {
			a.Logger.Error("healthz encode", "err", err)
		}
	}
}
