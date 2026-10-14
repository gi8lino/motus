// Package domain contains transport-independent Motus domain types.
package domain

// AuthMode selects the primary browser authentication strategy.
type AuthMode string

const (
	// AuthModeLocal authenticates users with Motus local credentials.
	AuthModeLocal AuthMode = "local"
	// AuthModeTrustedProxy accepts identities asserted by a trusted reverse proxy.
	AuthModeTrustedProxy AuthMode = "trusted-proxy"
	// AuthModeProxy is the preferred name for reverse-proxy authentication.
	AuthModeProxy AuthMode = "proxy"
	// AuthModeOIDC authenticates users through OpenID Connect.
	AuthModeOIDC AuthMode = "oidc"
	// AuthModeNone permits access behind an already trusted boundary.
	AuthModeNone AuthMode = "none"
)
