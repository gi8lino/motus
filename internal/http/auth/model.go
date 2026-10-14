package httpauth

import (
	"context"
	"net/http"
	"time"

	"github.com/gi8lino/motus/internal/db"
)

// BrowserConfig contains deployment-managed browser authentication settings.
type BrowserConfig struct {
	// ModeOverride forces a deployment-managed browser authentication method when non-empty.
	ModeOverride Mode
	// LocalLoginEnabled permits recovery sign-in with a Motus password.
	LocalLoginEnabled bool
	// AllowUserRegistrationOverride overrides persisted external-user registration when non-nil.
	AllowUserRegistrationOverride *bool
	// SessionLifetime controls the lifespan of Motus browser sessions.
	SessionLifetime time.Duration
	// TrustedProxyCIDRs explicitly identifies peers permitted to assert identity headers.
	TrustedProxyCIDRs []string
	// ProxyOverrides contains explicit CLI/environment header overrides, not defaults.
	ProxyOverrides ProxyOverrides
	// PublicURL configures secure session cookies.
	PublicURL string
	// RoutePrefix scopes the browser cookie to this application.
	RoutePrefix string
	// OIDC configures the OpenID Connect browser flow.
	OIDC OIDCConfig
}

// BrowserAuth groups browser identity resolution with its public handlers.
type BrowserAuth struct {
	// Authenticator resolves the current browser user.
	Authenticator Authenticator
	// Login starts the primary interactive login flow.
	Login http.Handler
	// Callback completes the configured OIDC flow when present.
	Callback http.Handler
	// Local owns local-password login and browser sessions.
	Local *Local
	// Current reports the effective browser authentication state without another database read.
	Current func() BrowserState
	// Validate checks whether proposed persisted settings can become the effective runtime configuration.
	Validate func(context.Context, db.AuthenticationSettings) error
	// ApplySettings publishes settings that Validate has accepted after persistence succeeds.
	ApplySettings func(db.AuthenticationSettings)
}

// BrowserState is the process-resident authentication state exposed to browser routes.
type BrowserState struct {
	// Mode is the configured primary browser authentication mode.
	Mode Mode
	// AllowRegistration reports whether the effective settings permit account provisioning.
	AllowRegistration bool
	// LocalLoginEnabled reports whether local sign-in is currently available.
	LocalLoginEnabled bool
}

// ProxyOverrides supplies only explicitly configured identity-header overrides.
type ProxyOverrides struct {
	Username    *[]string
	Email       *[]string
	DisplayName *[]string
	Groups      *[]string
}
