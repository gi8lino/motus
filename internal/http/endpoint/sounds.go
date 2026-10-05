package endpoint

import (
	"github.com/gi8lino/motus/internal/http/response"
	"log/slog"
	"net/http"

	"github.com/gi8lino/motus/internal/service/sounds"
)

// ListSounds returns available sound options.
func ListSounds(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.Respond(w, http.StatusOK, sounds.BuiltinOptions)
	}
}
