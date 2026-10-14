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
	// Mode is the configured primary browser authentication mode.
	Mode Mode
	// LocalLoginEnabled reports whether recovery local login is available.
	LocalLoginEnabled bool
	// ApplySettings validates and immediately activates persisted authentication settings.
	ApplySettings func(context.Context, db.AuthenticationSettings) error
}
