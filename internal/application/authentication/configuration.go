package authentication

import (
	"context"
	"fmt"
	"strings"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/pkg/domain"
)

// SettingsRepository persists browser authentication configuration.
type SettingsRepository interface {
	AuthenticationSettings(context.Context) (db.AuthenticationSettings, error)
	SaveAuthenticationSettings(context.Context, db.AuthenticationSettings) error
}

// Configuration manages persisted browser authentication configuration.
type Configuration struct{ repository SettingsRepository }

// NewConfiguration constructs authentication configuration use cases.
func NewConfiguration(repository SettingsRepository) *Configuration {
	return &Configuration{repository: repository}
}

// Current returns the effective persisted authentication configuration.
func (s *Configuration) Current(ctx context.Context) (Settings, error) {
	return s.repository.AuthenticationSettings(ctx)
}

// Save validates and persists browser authentication configuration.
func (s *Configuration) Save(ctx context.Context, settings Settings) error {
	switch settings.Mode {
	case domain.AuthModeLocal, domain.AuthModeNone:
	case domain.AuthModeTrustedProxy:
		if len(settings.TrustedUsernameHeaders) == 0 {
			return fmt.Errorf("trusted-proxy authentication requires a username header")
		}
	case domain.AuthModeOIDC:
		if strings.TrimSpace(settings.OIDCIssuer) == "" || strings.TrimSpace(settings.OIDCClientID) == "" {
			return fmt.Errorf("oidc issuer and client ID are required")
		}
	default:
		return fmt.Errorf("unknown authentication mode %q", settings.Mode)
	}
	return s.repository.SaveAuthenticationSettings(ctx, settings)
}
