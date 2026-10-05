package httpauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gi8lino/motus/internal/db"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type localRepositoryStub struct {
	user         *db.User
	hash         string
	createdToken string
	updated      bool
	revoked      bool
}

func (s *localRepositoryStub) CreateSession(_ context.Context, token, _ string, _ time.Time) error {
	s.createdToken = token
	return nil
}
func (s *localRepositoryStub) GetSessionUser(context.Context, string, time.Time) (*db.User, error) {
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
func (*localRepositoryStub) SetupRequired(context.Context) (bool, error) { return false, nil }
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
	require.Equal(t, token, repository.createdToken)

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
