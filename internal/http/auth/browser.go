package httpauth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

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
		if strings.TrimSpace(config.OIDC.ClientSecret) == "" {
			return errors.New("oidc client secret is required")
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
	mu sync.RWMutex
	// primary authenticates the configured browser authentication mode.
	primary Authenticator
	// local authenticates recovery local browser sessions.
	local *Local
	// oidc starts and completes the configured OIDC browser flow.
	oidc *OIDC
	// mode identifies the configured primary browser authentication mode.
	mode Mode
	// allowRegistration reports the effective external identity provisioning policy.
	allowRegistration bool
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
	local := NewLocal(repository, config.SessionLifetime)
	browser := &browserAuthenticator{local: local, localLoginEnabled: config.LocalLoginEnabled}
	if err := browser.applySettings(ctx, config, repository, settings); err != nil {
		return BrowserAuth{}, err
	}
	return BrowserAuth{
		Authenticator: browser,
		Local:         local,
		Current:       browser.current,
		ApplySettings: func(ctx context.Context, settings db.AuthenticationSettings) error {
			return browser.applySettings(ctx, config, repository, settings)
		},
		Login:    http.HandlerFunc(browser.login),
		Callback: http.HandlerFunc(browser.callback),
	}, nil
}

// applySettings rebuilds the effective authenticator from persisted configuration.
func (b *browserAuthenticator) applySettings(ctx context.Context, config BrowserConfig, repository browserRepository, settings db.AuthenticationSettings) error {
	mode := Mode(settings.Mode)
	if config.ModeOverride != "" {
		mode = config.ModeOverride
	}
	if err := validateSettings(mode, config, settings); err != nil {
		return err
	}
	allowRegistration := settings.AllowRegistration
	if config.AllowUserRegistrationOverride != nil {
		allowRegistration = *config.AllowUserRegistrationOverride
	}
	externalIdentity := appauth.NewExternalIdentity(repository, allowRegistration)
	var primary Authenticator
	var oidcAuth *OIDC
	switch mode {
	case ModeLocal:
		primary = b.local
	case ModeNone:
		primary = NewNone(repository)
	case ModeTrustedProxy:
		if len(settings.TrustedUsernameHeaders) == 0 {
			return errors.New("trusted-proxy authentication requires a username header")
		}
		primary = NewTrustedProxy(externalIdentity, settings.TrustedUsernameHeaders[0])
	case ModeOIDC:
		oidcConfig := config.OIDC
		oidcConfig.Issuer, oidcConfig.ClientID = settings.OIDCIssuer, settings.OIDCClientID
		var err error
		oidcAuth, err = NewOIDC(ctx, oidcConfig, externalIdentity, b.local)
		if err != nil {
			return err
		}
		primary = b.local
	default:
		return errors.New("unknown authentication mode")
	}
	b.mu.Lock()
	b.mode, b.primary, b.oidc, b.allowRegistration = mode, primary, oidcAuth, allowRegistration
	b.mu.Unlock()
	return nil
}

// current returns a consistent snapshot of browser-facing authentication state.
func (b *browserAuthenticator) current() BrowserState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return BrowserState{
		Mode:              b.mode,
		AllowRegistration: b.allowRegistration,
		LocalLoginEnabled: b.mode == ModeLocal || b.localLoginEnabled,
	}
}

func (b *browserAuthenticator) Authenticate(r *http.Request) (db.User, error) {
	b.mu.RLock()
	mode, primary := b.mode, b.primary
	b.mu.RUnlock()
	if mode == ModeLocal || b.localLoginEnabled {
		user, err := b.local.Authenticate(r)
		if err == nil {
			return user, nil
		}
		if !errors.Is(err, ErrUnauthenticated) {
			return db.User{}, err
		}
	}
	return primary.Authenticate(r)
}

func (b *browserAuthenticator) login(w http.ResponseWriter, r *http.Request) {
	b.mu.RLock()
	mode, oidcAuth := b.mode, b.oidc
	b.mu.RUnlock()
	if mode != ModeOIDC || oidcAuth == nil {
		response.Problem(w, http.StatusNotFound, "Interactive login is unavailable.")
		return
	}
	oidcAuth.login(w, r)
}

func (b *browserAuthenticator) callback(w http.ResponseWriter, r *http.Request) {
	b.mu.RLock()
	mode, oidcAuth := b.mode, b.oidc
	b.mu.RUnlock()
	if mode != ModeOIDC || oidcAuth == nil {
		response.Problem(w, http.StatusNotFound, "OIDC authentication is unavailable.")
		return
	}
	oidcAuth.callback(w, r)
}

// validateSettings checks the effective non-secret and deployment-managed OIDC configuration.
func validateSettings(mode Mode, config BrowserConfig, settings db.AuthenticationSettings) error {
	switch mode {
	case ModeLocal, ModeNone:
		return nil
	case ModeTrustedProxy:
		if len(settings.TrustedUsernameHeaders) == 0 {
			return errors.New("trusted-proxy authentication requires a username header")
		}
		return nil
	case ModeOIDC:
		if strings.TrimSpace(settings.OIDCIssuer) == "" || strings.TrimSpace(settings.OIDCClientID) == "" {
			return errors.New("oidc issuer and client ID are required")
		}
		if strings.TrimSpace(config.OIDC.ClientSecret) == "" || len(config.OIDC.SessionSecret) < 32 {
			return errors.New("oidc client secret and a session secret of at least 32 bytes are required")
		}
		return nil
	default:
		return errors.New("unknown authentication mode")
	}
}

var _ Authenticator = (*browserAuthenticator)(nil)
