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
		if err := encode(w, r, http.StatusOK, configResponse{
			AuthHeaderEnabled: a.AuthHeader != "",
			AllowRegistration: a.AllowRegistration,
			Version:           a.Version,
			Commit:            a.Commit,
		}); err != nil {
			a.Logger.Error("config encode", "err", err)
		}
	}
}

// CurrentUser resolves the authenticated user.
func (a *API) CurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("current user encode", "err", err)
			}
			return
		}

		user, err := a.Users.Get(r.Context(), userID)
		if err != nil || user == nil {
			if err := encode(w, r, http.StatusUnauthorized, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("current user encode", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, user); err != nil {
			a.Logger.Error("current user encode", "err", err)
		}
	}
}
