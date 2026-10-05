package httpauth

import (
	"net/http"
	"time"
)

// BrowserConfig contains deployment-managed browser authentication settings.
type BrowserConfig struct {
	Mode                   Mode
	TrustedProxyUserHeader string
	AllowRegistration      bool
	SessionLifetime        time.Duration
	OIDC                   OIDCConfig
}

// BrowserAuth groups browser identity resolution with its public handlers.
type BrowserAuth struct {
	Authenticator Authenticator
	Login         http.Handler
	Callback      http.Handler
	Local         *Local
	Mode          Mode
}
