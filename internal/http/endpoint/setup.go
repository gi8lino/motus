package endpoint

import (
	"github.com/gi8lino/motus/internal/http/response"
	"log/slog"
	"net/http"

	httpauth "github.com/gi8lino/motus/internal/http/auth"
)

// Setup bootstraps Motus's first local administrator exactly once.
func Setup(local *httpauth.Local, logger *slog.Logger) http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		required, err := local.SetupRequired(r.Context())
		if err != nil || !required {
			response.Problem(w, http.StatusForbidden, "setup is unavailable")
			return
		}
		body, err := decode[request](r)
		if err != nil {
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}
		user, token, err := local.Setup(r.Context(), body.Email, body.Password)
		if err != nil {
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}
		local.WriteSessionCookie(w, r, token)
		response.Respond(w, http.StatusCreated, user)
	}
}
