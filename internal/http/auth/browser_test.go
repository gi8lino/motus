package httpauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appkitproxy "github.com/containeroo/appkit/auth/proxy"
	"github.com/gi8lino/motus/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// proxyBrowserRepository implements only the persistence used by these proxy-mode tests.
type proxyBrowserRepository struct {
	browserRepository
	settings db.AuthenticationSettings
}

// AuthenticationSettings provides unmodified persisted authentication settings.
func (s *proxyBrowserRepository) AuthenticationSettings(context.Context) (db.AuthenticationSettings, error) {
	return s.settings, nil
}

// GetIdentityUser simulates an identity not yet registered locally.
func (s *proxyBrowserRepository) GetIdentityUser(context.Context, string, string) (*db.User, error) {
	return nil, pgx.ErrNoRows
}

// ResolveExternalIdentity simulates application-managed registration of a verified proxy identity.
func (s *proxyBrowserRepository) ResolveExternalIdentity(_ context.Context, _, subject, email, _ string, allow bool) (*db.User, error) {
	if !allow {
		return nil, pgx.ErrNoRows
	}
	return &db.User{ID: subject, Name: email}, nil
}

// TestBrowserProxyOverridesApplyBeforeValidation exercises proxy login on an empty database configuration.
func TestBrowserProxyOverridesApplyBeforeValidation(t *testing.T) {
	t.Parallel()

	store := &proxyBrowserRepository{settings: db.AuthenticationSettings{Mode: ModeLocal}}
	username := []string{"X-Auth-Request-Email"}
	email := []string{"X-Auth-Request-Email"}
	cidrs := []string{"127.0.0.1/32"}
	allow := true
	config := BrowserConfig{
		ModeOverride:                  ModeProxy,
		AllowUserRegistrationOverride: &allow,
		SessionLifetime:               time.Hour,
		ProxyOverrides: appkitproxy.Layer{
			TrustedCIDRs:    &cidrs,
			UsernameHeaders: &username,
			EmailHeaders:    &email,
		},
	}

	browser, err := ConfigureBrowserAuth(context.Background(), config, store)
	require.NoError(t, err)
	require.Equal(t, ModeProxy, browser.Current().Mode)
	require.True(t, browser.Current().AllowRegistration)
	require.Empty(t, store.settings.TrustedUsernameHeaders, "deployment headers must not be persisted")
	require.Empty(t, store.settings.TrustedEmailHeaders, "deployment headers must not be persisted")

	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.RemoteAddr = "127.0.0.1:41123"
	request.Header.Set("X-Auth-Request-Email", "alice@example.com")
	user, err := browser.Authenticator.Authenticate(request)
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", user.ID)

	// Saving empty persisted headers remains valid while deployment overrides are active.
	updated := db.AuthenticationSettings{Mode: ModeProxy}
	require.NoError(t, browser.Validate(context.Background(), updated))
	browser.ApplySettings(updated)
	user, err = browser.Authenticator.Authenticate(request)
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", user.ID)
}

// TestBrowserProxyEmailOnlyOverride verifies a verified email claim can be the sole identity source.
func TestBrowserProxyEmailOnlyOverride(t *testing.T) {
	t.Parallel()

	store := &proxyBrowserRepository{settings: db.AuthenticationSettings{Mode: ModeProxy}}
	email := []string{"X-Auth-Request-Email"}
	empty := []string{}
	cidrs := []string{"*"}
	allow := true
	browser, err := ConfigureBrowserAuth(context.Background(), BrowserConfig{
		AllowUserRegistrationOverride: &allow,
		ProxyOverrides: appkitproxy.Layer{
			TrustedCIDRs: &cidrs, UsernameHeaders: &empty, EmailHeaders: &email,
		},
	}, store)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.RemoteAddr = "192.0.2.10:443"
	request.Header.Set("X-Auth-Request-Email", "alice@example.com")
	request.Header.Set("X-Forwarded-User", "spoofed")
	user, err := browser.Authenticator.Authenticate(request)
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", user.ID)
}

// TestBrowserProxyRequiresEffectiveIdentityHeader rejects explicitly empty identity sources.
func TestBrowserProxyRequiresEffectiveIdentityHeader(t *testing.T) {
	t.Parallel()

	store := &proxyBrowserRepository{settings: db.AuthenticationSettings{
		Mode: ModeProxy, TrustedUsernameHeaders: []string{"X-Remote-User"},
	}}
	empty := []string{}
	cidrs := []string{"127.0.0.1/32"}
	_, err := ConfigureBrowserAuth(context.Background(), BrowserConfig{
		ProxyOverrides: appkitproxy.Layer{
			TrustedCIDRs: &cidrs, UsernameHeaders: &empty, EmailHeaders: &empty,
		},
	}, store)
	require.ErrorContains(t, err, "username or email header")
	require.Equal(t, []string{"X-Remote-User"}, store.settings.TrustedUsernameHeaders)
}

// TestBrowserProxyUsesPersistedHeadersWithoutOverrides keeps existing installations working.
func TestBrowserProxyUsesPersistedHeadersWithoutOverrides(t *testing.T) {
	t.Parallel()

	store := &proxyBrowserRepository{settings: db.AuthenticationSettings{
		Mode:                   ModeProxy,
		AllowRegistration:      true,
		TrustedUsernameHeaders: []string{"X-Remote-User"},
		TrustedEmailHeaders:    []string{"X-Remote-Email"},
	}}
	cidrs := []string{"127.0.0.1/32"}
	browser, err := ConfigureBrowserAuth(context.Background(), BrowserConfig{
		ProxyOverrides: appkitproxy.Layer{TrustedCIDRs: &cidrs},
	}, store)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.RemoteAddr = "127.0.0.1:41123"
	request.Header.Set("X-Remote-User", "alice")
	request.Header.Set("X-Remote-Email", "alice@example.com")
	user, err := browser.Authenticator.Authenticate(request)
	require.NoError(t, err)
	require.Equal(t, "alice", user.ID)
}

// TestBrowserProxyDefaultsSupportEmptyDatabase ensures first-start proxy mode inherits appkit defaults.
func TestBrowserProxyDefaultsSupportEmptyDatabase(t *testing.T) {
	t.Parallel()

	store := &proxyBrowserRepository{settings: db.AuthenticationSettings{Mode: ModeLocal}}
	allow := true
	cidrs := []string{"*"}
	config := BrowserConfig{
		ModeOverride:                  ModeProxy,
		AllowUserRegistrationOverride: &allow,
		ProxyOverrides:                appkitproxy.Layer{TrustedCIDRs: &cidrs},
	}

	browser, err := ConfigureBrowserAuth(context.Background(), config, store)
	require.NoError(t, err)
	require.Equal(t, ModeProxy, browser.Current().Mode)

	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.RemoteAddr = "192.0.2.10:443"
	request.Header.Set("X-Auth-Request-Email", "alice@example.com")
	user, err := browser.Authenticator.Authenticate(request)
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", user.ID)
	require.Empty(t, store.settings.TrustedUsernameHeaders)
	require.Empty(t, store.settings.TrustedEmailHeaders)

	// Runtime changes likewise inherit defaults without persisting them.
	updated := db.AuthenticationSettings{Mode: ModeProxy}
	require.NoError(t, browser.Validate(context.Background(), updated))
	browser.ApplySettings(updated)
	user, err = browser.Authenticator.Authenticate(request)
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", user.ID)
}

// TestProxyResolutionPrecedence keeps stored settings above defaults and deployment overrides above both.
func TestProxyResolutionPrecedence(t *testing.T) {
	t.Parallel()
	stored := db.AuthenticationSettings{
		TrustedUsernameHeaders: []string{"X-Saved-User"},
		TrustedGroupHeaders:    []string{"X-Saved-Groups"},
	}
	cidrs := []string{"*"}
	resolved, err := appkitproxy.Resolve(appkitproxy.Options{
		Persisted: persistedProxyLayer(stored),
		Overrides: appkitproxy.Layer{TrustedCIDRs: &cidrs},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"X-Saved-User"}, resolved.UsernameHeaders)
	require.Contains(t, resolved.EmailHeaders, "X-Auth-Request-Email")
	require.Equal(t, []string{"X-Saved-Groups"}, resolved.GroupHeaders)

	custom := []string{"X-CLI-User"}
	empty := []string{}
	resolved, err = appkitproxy.Resolve(appkitproxy.Options{
		Persisted: persistedProxyLayer(stored),
		Overrides: appkitproxy.Layer{
			TrustedCIDRs: &cidrs, UsernameHeaders: &custom, EmailHeaders: &empty,
		},
	})
	require.NoError(t, err)
	require.Equal(t, custom, resolved.UsernameHeaders)
	require.Empty(t, resolved.EmailHeaders)
	require.Equal(t, []string{"X-Saved-User"}, stored.TrustedUsernameHeaders)
	require.Empty(t, stored.TrustedEmailHeaders)
}

// TestBrowserProxyExplicitEmptyOverridesDisableDefaults rejects intentionally disabling both identity sources.
func TestBrowserProxyExplicitEmptyOverridesDisableDefaults(t *testing.T) {
	t.Parallel()
	empty := []string{}
	cidrs := []string{"*"}
	store := &proxyBrowserRepository{settings: db.AuthenticationSettings{Mode: ModeProxy}}
	_, err := ConfigureBrowserAuth(context.Background(), BrowserConfig{
		ProxyOverrides: appkitproxy.Layer{
			TrustedCIDRs: &cidrs, UsernameHeaders: &empty, EmailHeaders: &empty,
		},
	}, store)
	require.ErrorContains(t, err, "username or email header")
}

// TestBrowserProxyRequiresExplicitTrust ensures appkit's default deny-all peer policy is kept.
func TestBrowserProxyRequiresExplicitTrust(t *testing.T) {
	t.Parallel()
	store := &proxyBrowserRepository{settings: db.AuthenticationSettings{Mode: ModeProxy}}
	_, err := ConfigureBrowserAuth(context.Background(), BrowserConfig{}, store)
	require.ErrorContains(t, err, "trusted-proxy-cidrs")
}
