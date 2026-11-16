package httpauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gi8lino/motus/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type localRepositoryStub struct {
	user         *db.User
	hash         string
	createdToken string
	createdHash  string
	legacyToken  string
	updated      bool
	revoked      bool
}

func (s *localRepositoryStub) CreateSession(_ context.Context, token, _ string, _ time.Time) error {
	s.createdToken = token
	return nil
}
func (s *localRepositoryStub) CreateSessionHash(_ context.Context, hash, _ string, _ time.Time) error {
	s.createdHash = hash
	return nil
}
func (s *localRepositoryStub) GetSessionUserByHash(_ context.Context, hash string, _ time.Time) (*db.User, error) {
	if hash == s.createdHash && s.user != nil {
		return s.user, nil
	}
	return nil, pgx.ErrNoRows
}
func (*localRepositoryStub) DeleteSessionHash(context.Context, string) error { return nil }
func (s *localRepositoryStub) GetSessionUser(_ context.Context, token string, _ time.Time) (*db.User, error) {
	if token != s.legacyToken {
		return nil, pgx.ErrNoRows
	}
	return s.user, nil
}
func (*localRepositoryStub) DeleteSession(context.Context, string) error { return nil }
func (s *localRepositoryStub) DeleteUserSessions(context.Context, string) error {
	s.revoked = true
	return nil
}
func (s *localRepositoryStub) GetUserWithPassword(context.Context, string) (*db.User, string, error) {
	return s.user, s.hash, nil
}
func (s *localRepositoryStub) UpdateUserPassword(context.Context, string, string) error {
	s.updated = true
	return nil
}
func (*localRepositoryStub) CreateInitialAdministrator(context.Context, string, string) (*db.User, error) {
	return nil, errors.New("not implemented")
}

func TestLocalSignInOwnsSessionLifecycle(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	require.NoError(t, err)
	repository := &localRepositoryStub{user: &db.User{ID: "user@example.com"}, hash: string(hash)}
	local := NewLocal(repository, time.Hour)

	user, token, err := local.SignIn(context.Background(), "user@example.com", "secret")
	require.NoError(t, err)
	require.Equal(t, "user@example.com", user.ID)
	require.NotEmpty(t, token)
	require.NotEqual(t, token, repository.createdHash)
	require.NotEmpty(t, repository.createdHash)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	local.WriteSessionCookie(recorder, request, token)
	require.NotEmpty(t, recorder.Result().Cookies())
}

func TestLocalChangePasswordRevokesSessions(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("old"), bcrypt.DefaultCost)
	require.NoError(t, err)
	repository := &localRepositoryStub{user: &db.User{ID: "user@example.com"}, hash: string(hash)}

	err = NewLocal(repository, time.Hour).ChangePassword(context.Background(), "user@example.com", "old", "new")
	require.NoError(t, err)
	require.True(t, repository.updated)
	require.True(t, repository.revoked)
}

func TestLocalRejectsSessionWithoutSignInOrigin(t *testing.T) {
	repository := &localRepositoryStub{user: &db.User{ID: "legacy@example.com"}, legacyToken: "older-valid-session"}
	local := NewLocal(repository, time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: localSessionCookie, Value: repository.legacyToken})
	_, _, err := local.sessionOrigin(req)
	require.ErrorIs(t, err, ErrUnauthenticated)
}

func TestLocalRejectsMissingSession(t *testing.T) {
	repository := &localRepositoryStub{user: &db.User{ID: "user@example.com"}}
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: localSessionCookie, Value: "not-a-valid-token"})
	_, _, err := NewLocal(repository, time.Hour).sessionOrigin(req)
	require.ErrorIs(t, err, ErrUnauthenticated)
}

// TestBrowserSessionIssuerChange exercises the persisted token through Motus's request authenticator.
func TestBrowserSessionIssuerChange(t *testing.T) {
	t.Parallel()
	account := db.User{ID: "alice"}
	local := NewLocal(&localRepositoryStub{user: &account}, time.Hour)
	token, err := local.sessions.StartScoped(t.Context(), account, "oidc:https://original.example")
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.AddCookie(&http.Cookie{Name: localSessionCookie, Value: token})

	b := browserAuthenticator{mode: ModeOIDC, local: local, oidc: &OIDC{issuer: "https://original.example"}}
	user, err := b.Authenticate(request)
	require.NoError(t, err)
	require.Equal(t, account.ID, user.ID)

	b.publish(configuredAuthenticator{mode: ModeOIDC, oidc: &OIDC{issuer: "https://replacement.example"}})
	_, err = b.Authenticate(request)
	require.ErrorIs(t, err, ErrUnauthenticated)
}
