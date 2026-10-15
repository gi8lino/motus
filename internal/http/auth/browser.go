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
	"github.com/gi8lino/motus/pkg/domain"
)

// Mode is the HTTP authentication alias for the transport-independent authentication mode.
type Mode = domain.AuthMode

const (
	ModeLocal        = domain.AuthModeLocal
	ModeTrustedProxy = domain.AuthModeTrustedProxy
	ModeProxy        = domain.AuthModeProxy
	ModeOIDC         = domain.AuthModeOIDC
	ModeNone         = domain.AuthModeNone
)

// ValidateConfig verifies deployment-managed browser authentication settings.
func ValidateConfig(config BrowserConfig) error {
	switch config.ModeOverride {
	case "", ModeLocal, ModeNone:
		return nil
	case ModeTrustedProxy, ModeProxy:
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
	// prepared holds the runtime validated for the next matching persisted settings update.
	prepared *preparedAuthenticator
}

// configuredAuthenticator is one complete, not-yet-published browser authentication runtime.
type configuredAuthenticator struct {
	// primary authenticates requests after this configuration is published.
	primary Authenticator
	// oidc serves the interactive OIDC flow when OIDC is active.
	oidc *OIDC
	// mode identifies the effective primary authentication method.
	mode Mode
	// allowRegistration controls external identity provisioning for this configuration.
	allowRegistration bool
}

// preparedAuthenticator binds a validated runtime to the exact settings it can safely publish.
type preparedAuthenticator struct {
	// settings are the persisted values accepted during validation.
	settings db.AuthenticationSettings
	// configured is the runtime built from settings.
	configured configuredAuthenticator
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
	local := NewLocal(repository, config.SessionLifetime, CookieOptions{PublicURL: config.PublicURL, RoutePrefix: config.RoutePrefix})
	browser := &browserAuthenticator{local: local, localLoginEnabled: config.LocalLoginEnabled}
	configured, err := browser.configure(ctx, config, repository, settings)
	if err != nil {
		return BrowserAuth{}, err
	}
	browser.publish(configured)
	return BrowserAuth{
		Authenticator: browser,
		Local:         local,
		Current:       browser.current,
		Validate: func(ctx context.Context, settings db.AuthenticationSettings) error {
			return browser.validate(ctx, config, repository, settings)
		},
		ApplySettings: browser.applySettings,
		Login:         http.HandlerFunc(browser.login),
		Callback:      http.HandlerFunc(browser.callback),
	}, nil
}

// validate prepares the exact runtime that can be published after settings persistence succeeds.
func (b *browserAuthenticator) validate(ctx context.Context, config BrowserConfig, repository browserRepository, settings db.AuthenticationSettings) error {
	configured, err := b.configure(ctx, config, repository, settings)
	if err != nil {
		return err
	}
	b.mu.Lock()
	b.prepared = &preparedAuthenticator{settings: settings, configured: configured}
	b.mu.Unlock()
	return nil
}

// applySettings publishes a runtime previously accepted by Validate.
func (b *browserAuthenticator) applySettings(settings db.AuthenticationSettings) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.prepared == nil || !authenticationSettingsEqual(b.prepared.settings, settings) {
		return
	}
	b.publishLocked(b.prepared.configured)
	b.prepared = nil
}

// publish replaces the effective runtime with a fully configured authentication boundary.
func (b *browserAuthenticator) publish(configured configuredAuthenticator) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.publishLocked(configured)
}

// publishLocked publishes configured while b.mu is held.
func (b *browserAuthenticator) publishLocked(configured configuredAuthenticator) {
	b.mode = configured.mode
	b.primary = configured.primary
	b.oidc = configured.oidc
	b.allowRegistration = configured.allowRegistration
}

// configure creates one complete browser authentication runtime without publishing it.
func (b *browserAuthenticator) configure(ctx context.Context, config BrowserConfig, repository browserRepository, settings db.AuthenticationSettings) (configuredAuthenticator, error) {
	mode := settings.Mode
	if config.ModeOverride != "" {
		mode = config.ModeOverride
	}
	// Validate the effective configuration, not the unmodified database values.
	// The original settings remain available for persisted-settings comparisons.
	settings = effectiveProxySettings(settings, config.ProxyDefaults, config.ProxyOverrides)
	if err := validateSettings(mode, config, settings); err != nil {
		return configuredAuthenticator{}, err
	}
	allowRegistration := settings.AllowRegistration
	if config.AllowUserRegistrationOverride != nil {
		allowRegistration = *config.AllowUserRegistrationOverride
	}
	oidcLogin := appauth.NewOIDC(repository, config.AllowUserRegistrationOverride)
	trustedProxyLogin := appauth.NewTrustedProxy(repository, config.AllowUserRegistrationOverride)
	var primary Authenticator
	var oidcAuth *OIDC
	switch mode {
	case ModeLocal:
		primary = b.local
	case ModeNone:
		primary = NewNone(repository)
	case ModeTrustedProxy, ModeProxy:
		var err error
		primary, err = NewTrustedProxy(trustedProxyLogin, TrustedProxyHeaders{Username: settings.TrustedUsernameHeaders, Email: settings.TrustedEmailHeaders, DisplayName: settings.TrustedDisplayNameHeaders, Groups: settings.TrustedGroupHeaders, AdminGroup: settings.TrustedAdminGroup}, config.TrustedProxyCIDRs)
		if err != nil {
			return configuredAuthenticator{}, err
		}
	case ModeOIDC:
		oidcConfig := config.OIDC
		oidcConfig.RoutePrefix = config.RoutePrefix
		if oidcConfig.GroupClaim == "" {
			oidcConfig.GroupClaim = settings.OIDCGroupClaim
		}
		if oidcConfig.Issuer == "" {
			oidcConfig.Issuer = settings.OIDCIssuer
		}
		if oidcConfig.ClientID == "" {
			oidcConfig.ClientID = settings.OIDCClientID
		}
		var err error
		oidcAuth, err = NewOIDC(ctx, oidcConfig, oidcLogin, b.local)
		if err != nil {
			return configuredAuthenticator{}, err
		}
		primary = b.local
	default:
		return configuredAuthenticator{}, errors.New("unknown authentication mode")
	}
	return configuredAuthenticator{primary: primary, oidc: oidcAuth, mode: mode, allowRegistration: allowRegistration}, nil
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

// effectiveProxySettings resolves defaults, persisted settings, then explicit overrides.
// The returned settings are runtime-only; none of these changes are persisted.
func effectiveProxySettings(settings db.AuthenticationSettings, defaults TrustedProxyHeaders, overrides ProxyOverrides) db.AuthenticationSettings {
	settings.TrustedUsernameHeaders = effectiveProxyHeaders(settings.TrustedUsernameHeaders, defaults.Username, overrides.Username)
	settings.TrustedEmailHeaders = effectiveProxyHeaders(settings.TrustedEmailHeaders, defaults.Email, overrides.Email)
	settings.TrustedDisplayNameHeaders = effectiveProxyHeaders(settings.TrustedDisplayNameHeaders, defaults.DisplayName, overrides.DisplayName)
	settings.TrustedGroupHeaders = effectiveProxyHeaders(settings.TrustedGroupHeaders, defaults.Groups, overrides.Groups)
	return settings
}

// effectiveProxyHeaders uses application defaults only when no persisted headers exist.
// An explicit deployment override takes precedence, including an empty list.
func effectiveProxyHeaders(persisted, defaults []string, override *[]string) []string {
	if override != nil {
		return append([]string(nil), (*override)...)
	}
	if len(persisted) != 0 {
		return append([]string(nil), persisted...)
	}
	return append([]string(nil), defaults...)
}

// validateSettings checks the effective non-secret and deployment-managed OIDC configuration.
func validateSettings(mode Mode, config BrowserConfig, settings db.AuthenticationSettings) error {
	switch mode {
	case ModeLocal, ModeNone:
		return nil
	case ModeTrustedProxy, ModeProxy:
		if len(settings.TrustedUsernameHeaders) == 0 && len(settings.TrustedEmailHeaders) == 0 {
			return errors.New("trusted-proxy authentication requires a username or email header")
		}
		return nil
	case ModeOIDC:
		if (strings.TrimSpace(settings.OIDCIssuer) == "" && strings.TrimSpace(config.OIDC.Issuer) == "") ||
			(strings.TrimSpace(settings.OIDCClientID) == "" && strings.TrimSpace(config.OIDC.ClientID) == "") {
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

// authenticationSettingsEqual compares every persisted value used to configure browser authentication.
func authenticationSettingsEqual(left, right db.AuthenticationSettings) bool {
	if left.Mode != right.Mode || left.AllowRegistration != right.AllowRegistration ||
		left.TrustedAdminGroup != right.TrustedAdminGroup || left.OIDCIssuer != right.OIDCIssuer ||
		left.OIDCClientID != right.OIDCClientID || left.OIDCGroupClaim != right.OIDCGroupClaim ||
		left.OIDCGroupSync != right.OIDCGroupSync || left.OIDCGroupsAuthoritative != right.OIDCGroupsAuthoritative ||
		left.OIDCAdminGroup != right.OIDCAdminGroup {
		return false
	}
	return stringSlicesEqual(left.TrustedUsernameHeaders, right.TrustedUsernameHeaders) &&
		stringSlicesEqual(left.TrustedEmailHeaders, right.TrustedEmailHeaders) &&
		stringSlicesEqual(left.TrustedDisplayNameHeaders, right.TrustedDisplayNameHeaders) &&
		stringSlicesEqual(left.TrustedGroupHeaders, right.TrustedGroupHeaders)
}

// stringSlicesEqual reports whether two ordered string slices contain the same values.
func stringSlicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

var _ Authenticator = (*browserAuthenticator)(nil)
