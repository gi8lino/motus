package httpauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	allow := true
	config := BrowserConfig{
		ModeOverride:                  ModeProxy,
		AllowUserRegistrationOverride: &allow,
		SessionLifetime:               time.Hour,
		TrustedProxyCIDRs:             []string{"127.0.0.1/32"},
		ProxyOverrides: ProxyOverrides{
			Username: &username,
			Email:    &email,
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

// TestBrowserProxyEmailOnlyOverride exercises the supported email-as-username fallback.
func TestBrowserProxyEmailOnlyOverride(t *testing.T) {
	t.Parallel()

	store := &proxyBrowserRepository{settings: db.AuthenticationSettings{Mode: ModeProxy}}
	email := []string{"X-Auth-Request-Email"}
	allow := true
	browser, err := ConfigureBrowserAuth(context.Background(), BrowserConfig{
		AllowUserRegistrationOverride: &allow,
		TrustedProxyCIDRs:             []string{"*"},
		ProxyOverrides:                ProxyOverrides{Email: &email},
	}, store)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.RemoteAddr = "192.0.2.10:443"
	request.Header.Set("X-Auth-Request-Email", "alice@example.com")
	user, err := browser.Authenticator.Authenticate(request)
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", user.ID)
}

// TestBrowserProxyRequiresEffectiveIdentityHeader rejects empty effective proxy headers.
func TestBrowserProxyRequiresEffectiveIdentityHeader(t *testing.T) {
	t.Parallel()

	store := &proxyBrowserRepository{settings: db.AuthenticationSettings{Mode: ModeProxy}}
	_, err := ConfigureBrowserAuth(context.Background(), BrowserConfig{TrustedProxyCIDRs: []string{"127.0.0.1/32"}}, store)
	require.ErrorContains(t, err, "username or email header")

	// An explicit empty override takes precedence over a nonempty stored header.
	stored := &proxyBrowserRepository{settings: db.AuthenticationSettings{
		Mode:                   ModeProxy,
		TrustedUsernameHeaders: []string{"X-Remote-User"},
	}}
	empty := []string{}
	_, err = ConfigureBrowserAuth(context.Background(), BrowserConfig{
		TrustedProxyCIDRs: []string{"127.0.0.1/32"},
		ProxyOverrides:    ProxyOverrides{Username: &empty},
	}, stored)
	require.ErrorContains(t, err, "username or email header")
	require.Equal(t, []string{"X-Remote-User"}, stored.settings.TrustedUsernameHeaders)
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
	browser, err := ConfigureBrowserAuth(context.Background(), BrowserConfig{
		TrustedProxyCIDRs: []string{"127.0.0.1/32"},
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
