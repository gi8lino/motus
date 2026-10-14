package httpauth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	appauth "github.com/gi8lino/motus/internal/application/authentication"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/http/response"
)

// Mode selects the primary browser authentication strategy.
type Mode string

const (
	ModeLocal        Mode = "local"
	ModeTrustedProxy Mode = "trusted-proxy"
	ModeOIDC         Mode = "oidc"
	ModeNone         Mode = "none"
)

// ValidateConfig verifies deployment-managed browser authentication settings.
func ValidateConfig(config BrowserConfig) error {
	switch config.ModeOverride {
	case "", ModeLocal, ModeNone:
		return nil
	case ModeTrustedProxy:
		return nil
	case ModeOIDC:
		if strings.TrimSpace(config.OIDC.Issuer) == "" || strings.TrimSpace(config.OIDC.ClientID) == "" || strings.TrimSpace(config.OIDC.ClientSecret) == "" {
			return errors.New("oidc issuer, client ID, and client secret are required")
		}
		if len(config.OIDC.SessionSecret) < 32 {
			return errors.New("oidc session secret must be at least 32 bytes")
		}
	default:
		return errors.New("unknown authentication mode")
	}
	return nil
}

// browserAuthenticator owns effective mode selection and interactive flow routing.
type browserAuthenticator struct {
	// primary authenticates the configured browser authentication mode.
	primary Authenticator
	// local authenticates recovery local browser sessions.
	local *Local
	// oidc starts and completes the configured OIDC browser flow.
	oidc *OIDC
	// mode identifies the configured primary browser authentication mode.
	mode Mode
	// localLoginEnabled exposes local recovery login outside local mode.
	localLoginEnabled bool
}

// ConfigureBrowserAuth composes the complete browser authentication boundary.
func ConfigureBrowserAuth(ctx context.Context, config BrowserConfig, repository browserRepository) (BrowserAuth, error) {
	if err := ValidateConfig(config); err != nil {
		return BrowserAuth{}, err
	}
	settings, err := repository.AuthenticationSettings(ctx)
	if err != nil {
		return BrowserAuth{}, err
	}
	mode := Mode(settings.Mode)
	if config.ModeOverride != "" {
		mode = config.ModeOverride
	}
	local := NewLocal(repository, config.SessionLifetime)
	externalIdentity := appauth.NewExternalIdentity(repository, settings.AllowRegistration)
	browser := &browserAuthenticator{local: local, mode: mode, localLoginEnabled: config.LocalLoginEnabled}
	switch mode {
	case ModeLocal:
		browser.primary = local
	case ModeTrustedProxy:
		if len(settings.TrustedUsernameHeaders) == 0 {
			return BrowserAuth{}, errors.New("trusted-proxy authentication requires a username header")
		}
		browser.primary = NewTrustedProxy(externalIdentity, settings.TrustedUsernameHeaders[0])
	case ModeNone:
		browser.primary = NewNone(repository)
	case ModeOIDC:
		oidcConfig := config.OIDC
		oidcConfig.Issuer, oidcConfig.ClientID = settings.OIDCIssuer, settings.OIDCClientID
		oidc, err := NewOIDC(ctx, oidcConfig, externalIdentity, local)
		if err != nil {
			return BrowserAuth{}, err
		}
		browser.oidc, browser.primary = oidc, local
	}
	result := BrowserAuth{
		Authenticator:     browser,
		Local:             local,
		Mode:              mode,
		LocalLoginEnabled: mode == ModeLocal || config.LocalLoginEnabled,
		Login:             http.HandlerFunc(browser.login),
	}
	if browser.oidc != nil {
		result.Callback = http.HandlerFunc(browser.callback)
	}
	return result, nil
}

func (b *browserAuthenticator) Authenticate(r *http.Request) (db.User, error) {
	if b.mode == ModeLocal || b.localLoginEnabled {
		user, err := b.local.Authenticate(r)
		if err == nil {
			return user, nil
		}
		if !errors.Is(err, ErrUnauthenticated) {
			return db.User{}, err
		}
	}
	return b.primary.Authenticate(r)
}

func (b *browserAuthenticator) login(w http.ResponseWriter, r *http.Request) {
	if b.mode != ModeOIDC || b.oidc == nil {
		response.Problem(w, http.StatusNotFound, "Interactive login is unavailable.")
		return
	}
	b.oidc.login(w, r)
}

func (b *browserAuthenticator) callback(w http.ResponseWriter, r *http.Request) {
	if b.mode != ModeOIDC || b.oidc == nil {
		response.Problem(w, http.StatusNotFound, "OIDC authentication is unavailable.")
		return
	}
	b.oidc.callback(w, r)
}

var _ Authenticator = (*browserAuthenticator)(nil)
