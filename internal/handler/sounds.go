package handler

import (
	"net/http"

	"github.com/gi8lino/motus/internal/service/sounds"
)

// ListSounds returns available sound options.
func (a *API) ListSounds() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, sounds.BuiltinOptions)
	}
}
