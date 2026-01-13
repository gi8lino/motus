package handler

import (
	"net/http"

	"github.com/gi8lino/motus/internal/service/sounds"
)

// ListSounds returns available sound options.
func (a *API) ListSounds() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := encode(w, r, http.StatusOK, sounds.BuiltinOptions); err != nil {
			a.Logger.Error("sounds encode", "err", err)
		}
	}
}
