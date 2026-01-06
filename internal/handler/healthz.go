package handler

import "net/http"

// Healthz responds with a simple health check.
func (a *API) Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.Store.Ping(r.Context()); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	}
}
