package endpoint

import (
	"github.com/gi8lino/motus/internal/http/response"
	"net/http"

	httpauth "github.com/gi8lino/motus/internal/http/auth"
)

// Setup bootstraps Motus's first local administrator exactly once.
func Setup(setupRequired func() bool, local *httpauth.Local) http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if setupRequired == nil || !setupRequired() {
			response.Problem(w, http.StatusNotFound, "Not found.")
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
