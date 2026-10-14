package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	appkitsecrets "github.com/containeroo/appkit/secrets"
	appauth "github.com/gi8lino/motus/internal/application/authentication"
	"github.com/gi8lino/motus/internal/flags"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
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
	if _, err := appkitsecrets.New(cfg.Authentication.EncryptionKey); err != nil {
		return nil, fmt.Errorf("configure secrets: %w", err)
	}
	if err := httpauth.ValidateConfig(
		httpauth.BrowserConfig{
			ModeOverride:                  httpauth.Mode(*cfg.Authentication.ModeOverride),
			AllowUserRegistrationOverride: cfg.Authentication.AllowRegistrationOverride,
			OIDC: httpauth.OIDCConfig{
				Issuer: cfg.OIDC.Issuer, ClientID: cfg.OIDC.ClientID, ClientSecret: cfg.OIDC.ClientSecret, SessionSecret: cfg.OIDC.SessionSecret,
				PublicURL: cfg.Server.PublicURL,
			},
			LocalLoginEnabled: cfg.Authentication.LocalLogin,
			SessionLifetime:   cfg.Authentication.SessionLifetime,
		}); err != nil {
		return nil, fmt.Errorf("validate browser authentication: %w", err)
	}

	if err := SeedCoreExercises(ctx, store, logger, cfg.CoreExercisesFile); err != nil {
		return nil, fmt.Errorf("load core exercises: %w", err)
	}

	browser, err := httpauth.ConfigureBrowserAuth(ctx, httpauth.BrowserConfig{
		ModeOverride:                  httpauth.Mode(*cfg.Authentication.ModeOverride),
		LocalLoginEnabled:             cfg.Authentication.LocalLogin,
		AllowUserRegistrationOverride: cfg.Authentication.AllowRegistrationOverride,
		SessionLifetime:               cfg.Authentication.SessionLifetime,
		OIDC: httpauth.OIDCConfig{
			Issuer:        cfg.OIDC.Issuer,
			ClientID:      cfg.OIDC.ClientID,
			ClientSecret:  cfg.OIDC.ClientSecret,
			SessionSecret: cfg.OIDC.SessionSecret,
			PublicURL:     cfg.Server.PublicURL,
		},
	}, store)
	if err != nil {
		return nil, fmt.Errorf("configure browser authentication: %w", err)
	}

	return &Application{
		browser: browser,
		users: users.New(store, func() bool {
			state := browser.Current()
			return state.Mode == httpauth.ModeLocal && state.AllowRegistration
		}),
		exercises:      exercises.New(store),
		workouts:       workouts.New(store),
		trainings:      trainings.New(store, sounds.URLByKey),
		authentication: appauth.NewConfiguration(store),
	}, nil
}
