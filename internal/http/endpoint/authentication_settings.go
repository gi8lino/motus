package endpoint

import (
	"net/http"

	appauth "github.com/gi8lino/motus/internal/application/authentication"
	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/http/response"
)

// AuthenticationSettings exposes administrator browser-authentication configuration.
func AuthenticationSettings(service *appauth.Configuration, browser httpauth.BrowserAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settings, err := service.Current(r.Context())
			if err != nil {
				response.InternalServerError(nil, w, err)
				return
			}
			response.Respond(w, http.StatusOK, settings)
		case http.MethodPut:
			settings, err := decode[db.AuthenticationSettings](r)
			if err != nil {
				response.Problem(w, http.StatusBadRequest, err.Error())
				return
			}
			if browser.Validate != nil {
				if err := browser.Validate(r.Context(), settings); err != nil {
					response.Problem(w, http.StatusUnprocessableEntity, err.Error())
					return
				}
			}
			if err := service.Save(r.Context(), settings); err != nil {
				response.Problem(w, http.StatusUnprocessableEntity, err.Error())
				return
			}
			if browser.ApplySettings != nil {
				if err := browser.ApplySettings(r.Context(), settings); err != nil {
					response.InternalServerError(nil, w, err)
					return
				}
			}
			response.Respond(w, http.StatusNoContent, nil)
		default:
			response.Problem(w, http.StatusMethodNotAllowed, "Method not allowed.")
		}
	}
}
