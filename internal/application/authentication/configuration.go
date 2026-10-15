package authentication

import (
	"context"
	"fmt"

	appkitauth "github.com/containeroo/appkit/auth"
	"github.com/gi8lino/motus/internal/db"
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

// Save validates the persisted authentication mode and saves the settings.
// Effective provider and header validation is performed by browser.Validate before Save.
func (s *Configuration) Save(ctx context.Context, settings Settings) error {
	switch settings.Mode {
	case appkitauth.ModeLocal, appkitauth.ModeNone, appkitauth.ModeProxy, appkitauth.ModeOIDC:
	default:
		return fmt.Errorf("unknown authentication mode %q", settings.Mode)
	}
	return s.repository.SaveAuthenticationSettings(ctx, settings)
}
