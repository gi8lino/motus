package httpauth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	appkitoidc "github.com/containeroo/appkit/auth/oidc"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/http/response"
)

// OIDCConfig configures one externally trusted identity provider.
type OIDCConfig struct {
	Issuer        string
	ClientID      string
	ClientSecret  string
	SessionSecret string
	PublicURL     string
	RoutePrefix   string
	GroupClaim    string
}

// OIDC maps verified external identities into local Motus browser sessions.
type OIDC struct {
	flow         *appkitoidc.Flow
	loginService oidcLoginService
	local        *Local
}

// NewOIDC discovers and configures appkit's verified OIDC authorization flow.
func NewOIDC(ctx context.Context, cfg OIDCConfig, loginService oidcLoginService, local *Local) (*OIDC, error) {
	base := strings.TrimRight(cfg.PublicURL, "/")
	if cfg.RoutePrefix != "" && !strings.HasSuffix(base, strings.TrimRight(cfg.RoutePrefix, "/")) {
		base += strings.TrimRight(cfg.RoutePrefix, "/")
	}
	callback := base + "/auth/oidc/callback"
	cookiePath := "/"
	if cfg.RoutePrefix != "" {
		cookiePath = strings.TrimRight(cfg.RoutePrefix, "/") + "/"
	}
	flow, err := appkitoidc.New(ctx, appkitoidc.Config{
		Issuer:       cfg.Issuer,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		StateSecret:  cfg.SessionSecret,
		CallbackURL:  callback,
		CookieName:   "motus_oidc_state",
		CookiePath:   cookiePath,
		GroupClaim:   cfg.GroupClaim,
	})
	if err != nil {
		return nil, err
	}
	return &OIDC{flow: flow, loginService: loginService, local: local}, nil
}

// login delegates secure browser state and provider redirect handling to appkit.
func (o *OIDC) login(w http.ResponseWriter, r *http.Request) {
	if err := o.flow.Begin(w, r, r.URL.Query().Get("next")); err != nil {
		response.InternalServerError(nil, w, err)
	}
}

// callback resolves a verified provider identity, issues a session and redirects.
func (o *OIDC) callback(w http.ResponseWriter, r *http.Request) {
	identity, next, err := o.flow.Complete(w, r)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Problem(w, http.StatusUnauthorized, "Invalid identity.")
		} else {
			response.InternalServerError(nil, w, err)
		}
		return
	}
	username := identity.Username
	if username == "" {
		username = identity.Subject
	}
	user, err := o.loginService.Login(r.Context(), identity.Issuer, identity.Subject, username, identity.Email, identity.DisplayName)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrIdentityEmailInUse):
			response.Problem(w, http.StatusForbidden, "The external identity must be linked to your existing account by an administrator.")
		case errors.Is(err, ErrInvalidCredentials):
			response.Problem(w, http.StatusUnauthorized, "Invalid identity.")
		case errors.Is(err, ErrRegistrationDisabled):
			response.Problem(w, http.StatusForbidden, "Login is not permitted.")
		default:
			response.InternalServerError(nil, w, err)
		}
		return
	}
	if user == nil {
		response.Problem(w, http.StatusForbidden, "Login is not permitted.")
		return
	}
	token, err := o.local.StartSession(r.Context(), *user)
	if err != nil {
		response.InternalServerError(nil, w, err)
		return
	}
	o.local.WriteSessionCookie(w, r, token)
	http.Redirect(w, r, next, http.StatusFound)
}
