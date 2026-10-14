package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	appkitauth "github.com/containeroo/appkit/auth"

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
	if err := httpauth.ValidateConfig(
		httpauth.BrowserConfig{
			ModeOverride:                  authenticationMode(cfg.Authentication.ModeOverride),
			AllowUserRegistrationOverride: cfg.Authentication.AllowRegistrationOverride,
			OIDC: httpauth.OIDCConfig{
				Issuer: cfg.OIDC.Issuer, ClientID: cfg.OIDC.ClientID, ClientSecret: cfg.OIDC.ClientSecret, SessionSecret: cfg.OIDC.SessionSecret,
				PublicURL:   cfg.Server.PublicURL,
				RoutePrefix: cfg.Server.RoutePrefix,
				GroupClaim:  oidcGroupClaimOverride(cfg),
			},
			LocalLoginEnabled: cfg.Authentication.LocalLogin,
			SessionLifetime:   cfg.Authentication.SessionLifetime,
			TrustedProxyCIDRs: cfg.TrustedProxyCIDRs,
			ProxyOverrides:    proxyOverrides(cfg),
			PublicURL:         cfg.Server.PublicURL,
			RoutePrefix:       cfg.Server.RoutePrefix,
		}); err != nil {
		return nil, fmt.Errorf("validate browser authentication: %w", err)
	}

	if err := SeedCoreExercises(ctx, store, logger, cfg.CoreExercisesFile); err != nil {
		return nil, fmt.Errorf("load core exercises: %w", err)
	}

	browser, err := httpauth.ConfigureBrowserAuth(ctx, httpauth.BrowserConfig{
		ModeOverride:                  authenticationMode(cfg.Authentication.ModeOverride),
		LocalLoginEnabled:             cfg.Authentication.LocalLogin,
		AllowUserRegistrationOverride: cfg.Authentication.AllowRegistrationOverride,
		SessionLifetime:               cfg.Authentication.SessionLifetime,
		TrustedProxyCIDRs:             cfg.TrustedProxyCIDRs,
		ProxyOverrides:                proxyOverrides(cfg),
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

// authenticationMode keeps an absent deployment override distinct from explicit none.
func authenticationMode(override *appkitauth.Mode) httpauth.Mode {
	if override == nil {
		return ""
	}
	return httpauth.Mode(*override)
}

// proxyOverrides distinguishes explicitly selected deployment headers from defaults.
func proxyOverrides(cfg flags.Config) httpauth.ProxyOverrides {
	values := cfg.Overrides.Values()
	var result httpauth.ProxyOverrides
	if _, ok := values["trusted-username-headers"]; ok {
		headers := append([]string(nil), cfg.Proxy.UsernameHeaders...)
		result.Username = &headers
	}
	if _, ok := values["trusted-email-headers"]; ok {
		headers := append([]string(nil), cfg.Proxy.EmailHeaders...)
		result.Email = &headers
	}
	if _, ok := values["trusted-display-name-headers"]; ok {
		headers := append([]string(nil), cfg.Proxy.DisplayNameHeaders...)
		result.DisplayName = &headers
	}
	if _, ok := values["trusted-group-headers"]; ok {
		headers := append([]string(nil), cfg.Proxy.GroupHeaders...)
		result.Groups = &headers
	}
	return result
}

// oidcGroupClaimOverride leaves the database group-claim setting authoritative unless overridden.
func oidcGroupClaimOverride(cfg flags.Config) string {
	if _, ok := cfg.Overrides.Values()["oidc-group-claim"]; ok {
		return cfg.OIDC.GroupClaim
	}
	return ""
}
