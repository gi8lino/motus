package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	appkitauth "github.com/containeroo/appkit/auth"
	appkitproxy "github.com/containeroo/appkit/auth/proxy"

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
	if appkitproxy.IsWildcardCIDRs(cfg.Proxy.TrustedCIDRs) {
		logger.Warn("trusted proxy wildcard accepts identities from any valid peer; restrict ingress and strip untrusted identity headers",
			"event", "trusted_proxy_wildcard_enabled")
	}
	store := infrastructure.store
	browserConfig := newBrowserConfig(cfg)
	if err := httpauth.ValidateConfig(browserConfig); err != nil {
		return nil, fmt.Errorf("validate browser authentication: %w", err)
	}

	if err := SeedCoreExercises(ctx, store, logger, cfg.CoreExercisesFile); err != nil {
		return nil, fmt.Errorf("load core exercises: %w", err)
	}

	browser, err := httpauth.ConfigureBrowserAuth(ctx, browserConfig, store)
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

// newBrowserConfig passes explicit deployment overrides to the browser-auth runtime.
// Appkit resolves proxy header defaults against persisted settings at runtime.
func newBrowserConfig(cfg flags.Config) httpauth.BrowserConfig {
	return httpauth.BrowserConfig{
		ModeOverride:                  authenticationMode(cfg.Authentication.ModeOverride),
		LocalLoginEnabled:             cfg.Authentication.LocalLogin,
		AllowUserRegistrationOverride: cfg.Authentication.AllowRegistrationOverride,
		SessionLifetime:               cfg.Authentication.SessionLifetime,
		ProxyOverrides:                cfg.Proxy.DeploymentOverrides(cfg.Overrides),
		PublicURL:                     cfg.Server.PublicURL,
		RoutePrefix:                   cfg.Server.RoutePrefix,
		OIDC: httpauth.OIDCConfig{
			Issuer:        cfg.OIDC.Issuer,
			ClientID:      cfg.OIDC.ClientID,
			ClientSecret:  cfg.OIDC.ClientSecret,
			SessionSecret: cfg.OIDC.SessionSecret,
			PublicURL:     cfg.Server.PublicURL,
			RoutePrefix:   cfg.Server.RoutePrefix,
			GroupClaim:    oidcGroupClaimOverride(cfg),
		},
	}
}

// authenticationMode keeps an absent deployment override distinct from explicit none.
func authenticationMode(override *appkitauth.Mode) httpauth.Mode {
	if override == nil {
		return ""
	}
	return httpauth.Mode(*override)
}

// oidcGroupClaimOverride leaves the database group-claim setting authoritative unless overridden.
func oidcGroupClaimOverride(cfg flags.Config) string {
	if _, ok := cfg.Overrides.Values()["oidc-group-claim"]; ok {
		return cfg.OIDC.GroupClaim
	}
	return ""
}
