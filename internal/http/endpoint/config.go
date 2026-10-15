package endpoint

import (
	"net/http"

	"github.com/gi8lino/motus/internal/http/response"

	httpauth "github.com/gi8lino/motus/internal/http/auth"
)

// ClientConfiguration describes the runtime settings exposed to the SPA.
// It keeps the browser-facing contract together at the HTTP boundary.
type ClientConfiguration struct {
	// Current returns the process-resident browser authentication state.
	Current func() httpauth.BrowserState
	// Local resolves whether initial setup is required.
	Local *httpauth.Local
	// Version identifies the current application build.
	Version string
	// Commit identifies the source commit of the current build.
	Commit string
}

// configResponse describes the serialized client configuration.
type configResponse struct {
	AuthHeaderEnabled bool   `json:"authHeaderEnabled"` // AuthHeaderEnabled indicates proxy auth usage.
	AllowRegistration bool   `json:"allowRegistration"` // AllowRegistration enables local sign-up.
	OIDCEnabled       bool   `json:"oidcEnabled"`       // OIDCEnabled enables provider sign-in.
	LocalLoginEnabled bool   `json:"localLoginEnabled"` // LocalLoginEnabled enables local recovery sign-in.
	SetupRequired     bool   `json:"setupRequired"`     // SetupRequired enables first-run bootstrap.
	Version           string `json:"version"`           // Version is the build version string.
	Commit            string `json:"commit"`            // Commit is the build commit SHA.
}

// Handler returns the client-configuration endpoint.
func (config ClientConfiguration) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := httpauth.BrowserState{}
		if config.Current != nil {
			state = config.Current()
		}
		setupRequired := false
		if config.Local != nil {
			setupRequired, _ = config.Local.SetupRequired(r.Context())
		}
		response.Respond(w, http.StatusOK, configResponse{
			AuthHeaderEnabled: state.Mode == httpauth.ModeProxy,
			AllowRegistration: state.Mode == httpauth.ModeLocal && state.AllowRegistration,
			OIDCEnabled:       state.Mode == httpauth.ModeOIDC,
			LocalLoginEnabled: state.LocalLoginEnabled,
			SetupRequired:     setupRequired,
			Version:           config.Version,
			Commit:            config.Commit,
		})
	}
}

// CurrentUser returns the user resolved by authentication middleware.
func CurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := httpauth.User(r)
		if !ok {
			response.Problem(w, http.StatusUnauthorized, "authentication required")
			return
		}

		response.Respond(w, http.StatusOK, user)
	}
}
