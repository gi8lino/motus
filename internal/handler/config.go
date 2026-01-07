package handler

import "net/http"

// configResponse describes runtime settings exposed to the SPA.
type configResponse struct {
	AuthHeaderEnabled bool   `json:"authHeaderEnabled"` // AuthHeaderEnabled indicates proxy auth usage.
	AllowRegistration bool   `json:"allowRegistration"` // AllowRegistration enables local sign-up.
	Version           string `json:"version"`           // Version is the build version string.
	Commit            string `json:"commit"`            // Commit is the build commit SHA.
}

// Config returns runtime configuration for the SPA.
func (a *API) Config() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, configResponse{
			AuthHeaderEnabled: a.AuthHeader != "",
			AllowRegistration: a.AllowRegistration,
			Version:           a.Version,
			Commit:            a.Commit,
		})
	}
}

// CurrentUser resolves the authenticated user.
func (a *API) CurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		user, err := a.Store.GetUser(r.Context(), userID)
		if err != nil || user == nil {
			writeJSON(w, http.StatusNotFound, apiError{Error: "user not found"})
			return
		}

		writeJSON(w, http.StatusOK, user)
	}
}
