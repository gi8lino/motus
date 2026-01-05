package handler

import "net/http"

// ListSounds returns available sound options.
func (a *API) ListSounds() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, builtinSoundOptions)
	}
}
