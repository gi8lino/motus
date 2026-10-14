package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	appauth "github.com/gi8lino/motus/internal/application/authentication"
	"github.com/gi8lino/motus/internal/flags"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/secrets"
	"github.com/gi8lino/motus/internal/service/exercises"
	"github.com/gi8lino/motus/internal/service/sounds"
	"github.com/gi8lino/motus/internal/service/trainings"
	"github.com/gi8lino/motus/internal/service/users"
	"github.com/gi8lino/motus/internal/service/workouts"
)

// Application contains the composed Motus application capabilities.
type Application struct {
	// browser owns the selected HTTP authentication method.
	browser httpauth.BrowserAuth
	// users provides user-management use cases.
	users *users.Service
	// exercises provides exercise-catalog use cases.
	exercises *exercises.Service
	// workouts provides workout use cases.
	workouts *workouts.Service
	// trainings provides training use cases.
	trainings *trainings.Service
	// authentication manages persisted browser-authentication configuration.
	authentication *appauth.Configuration
}

// NewApplication composes application services around initialized process infrastructure.
func NewApplication(
	ctx context.Context,
	cfg flags.Config,
	infrastructure *Infrastructure,
	logger *slog.Logger,
) (*Application, error) {
	store := infrastructure.store
	if _, err := secrets.New(cfg.EncryptionKey); err != nil {
		return nil, fmt.Errorf("configure secrets: %w", err)
	}
	if err := httpauth.ValidateConfig(
		httpauth.BrowserConfig{
			ModeOverride: cfg.AuthMode,
			OIDC: httpauth.OIDCConfig{
				Issuer:        cfg.OIDCIssuer,
				ClientID:      cfg.OIDCClientID,
				ClientSecret:  cfg.OIDCClientSecret,
				SessionSecret: cfg.OIDCSessionSecret,
				PublicURL:     cfg.PublicURL,
			},
			LocalLoginEnabled: cfg.LocalLogin,
			SessionLifetime:   cfg.SessionLifetime,
		}); err != nil {
		return nil, fmt.Errorf("validate browser authentication: %w", err)
	}

	if err := SeedCoreExercises(ctx, store, logger, cfg.CoreExercisesFile); err != nil {
		return nil, fmt.Errorf("load core exercises: %w", err)
	}

	browser, err := httpauth.ConfigureBrowserAuth(ctx, httpauth.BrowserConfig{
		ModeOverride:      cfg.AuthMode,
		LocalLoginEnabled: cfg.LocalLogin,
		SessionLifetime:   cfg.SessionLifetime,
		OIDC: httpauth.OIDCConfig{
			Issuer:        cfg.OIDCIssuer,
			ClientID:      cfg.OIDCClientID,
			ClientSecret:  cfg.OIDCClientSecret,
			SessionSecret: cfg.OIDCSessionSecret,
			PublicURL:     cfg.PublicURL,
		},
	}, store)
	if err != nil {
		return nil, fmt.Errorf("configure browser authentication: %w", err)
	}

	return &Application{
		browser:        browser,
		users:          users.New(store, cfg.AllowRegistration),
		exercises:      exercises.New(store),
		workouts:       workouts.New(store),
		trainings:      trainings.New(store, sounds.URLByKey),
		authentication: appauth.NewConfiguration(store),
	}, nil
}
