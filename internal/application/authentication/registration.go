package authentication

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// registrationSettingsRepository loads persisted external-registration settings.
type registrationSettingsRepository interface {
	AuthenticationSettings(context.Context) (db.AuthenticationSettings, error)
}

// registrationPolicy resolves deployment and persisted external-registration settings.
type registrationPolicy struct {
	repository registrationSettingsRepository
	override   *bool
}

// newRegistrationPolicy constructs shared external-registration policy resolution.
func newRegistrationPolicy(repository registrationSettingsRepository, override *bool) registrationPolicy {
	return registrationPolicy{repository: repository, override: override}
}

// enabled resolves the deployment override before persisted settings.
func (p registrationPolicy) enabled(ctx context.Context) (bool, error) {
	if p.override != nil {
		return *p.override, nil
	}
	settings, err := p.repository.AuthenticationSettings(ctx)
	if err != nil {
		return false, err
	}
	return settings.AllowRegistration, nil
}
