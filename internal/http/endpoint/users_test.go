package endpoint

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/service/users"
)

type fakeUserStore struct {
	getUserFn             func(context.Context, string) (*db.User, error)
	listUsersFn           func(context.Context) ([]db.User, error)
	getUserWithPasswordFn func(context.Context, string) (*db.User, string, error)
	updateUserPasswordFn  func(context.Context, string, string) error
	updateUserAdminFn     func(context.Context, string, bool) error
	updateUserNameFn      func(context.Context, string, string) error
	createUserFn          func(context.Context, string, string, string) (*db.User, error)
}

func (f *fakeUserStore) CreateSession(context.Context, string, string, time.Time) error { return nil }
func (f *fakeUserStore) GetSessionUser(context.Context, string, time.Time) (*db.User, error) {
	return nil, nil
}
func (f *fakeUserStore) DeleteSession(context.Context, string) error      { return nil }
func (f *fakeUserStore) DeleteUserSessions(context.Context, string) error { return nil }
func (f *fakeUserStore) SetupRequired(context.Context) (bool, error)      { return false, nil }
func (f *fakeUserStore) CreateInitialAdministrator(context.Context, string, string) (*db.User, error) {
	return nil, nil
}

func (f *fakeUserStore) GetUser(ctx context.Context, id string) (*db.User, error) {
	if f.getUserFn == nil {
		return nil, nil
	}
	return f.getUserFn(ctx, id)
}

func (f *fakeUserStore) ListUsers(ctx context.Context) ([]db.User, error) {
	if f.listUsersFn == nil {
		return nil, nil
	}
	return f.listUsersFn(ctx)
}

func (f *fakeUserStore) GetUserWithPassword(ctx context.Context, id string) (*db.User, string, error) {
	if f.getUserWithPasswordFn == nil {
		return nil, "", nil
	}
	return f.getUserWithPasswordFn(ctx, id)
}

func (f *fakeUserStore) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	if f.updateUserPasswordFn == nil {
		return nil
	}
	return f.updateUserPasswordFn(ctx, id, passwordHash)
}

func (f *fakeUserStore) UpdateUserAdmin(ctx context.Context, id string, isAdmin bool) error {
	if f.updateUserAdminFn == nil {
		return nil
	}
	return f.updateUserAdminFn(ctx, id, isAdmin)
}

func (f *fakeUserStore) UpdateUserName(ctx context.Context, id, name string) error {
	if f.updateUserNameFn == nil {
		return nil
	}
	return f.updateUserNameFn(ctx, id, name)
}

func (f *fakeUserStore) CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error) {
	if f.createUserFn == nil {
		return nil, nil
	}
	return f.createUserFn(ctx, email, avatarURL, passwordHash)
}

func (*fakeUserStore) LinkAuthIdentity(context.Context, string, db.AuthIdentity) error { return nil }

func (*fakeUserStore) DeleteAuthIdentity(context.Context, string, string, string) error { return nil }

func TestUsersHandlers(t *testing.T) {
	t.Run("List users", func(t *testing.T) {
		store := &fakeUserStore{listUsersFn: func(context.Context) ([]db.User, error) {
			return []db.User{{ID: "user@example.com"}}, nil
		}}
		api := &testEndpoints{Logger: slog.Default(), Users: users.New(store, func() bool { return false })}
		h := GetUsers(api.Users, api.Logger)
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload []db.User
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Len(t, payload, 1)
		assert.Equal(t, "user@example.com", payload[0].ID)
	})

	t.Run("Create user", func(t *testing.T) {
		store := &fakeUserStore{createUserFn: func(context.Context, string, string, string) (*db.User, error) {
			return &db.User{ID: "user@example.com"}, nil
		}}
		api := &testEndpoints{Logger: slog.Default(), Users: users.New(store, func() bool { return true }), AllowRegistration: true}
		h := CreateUser(api.Users, httpauth.NewLocal(store, time.Hour), api.Logger)
		body := strings.NewReader(`{"email":"user@example.com","avatarUrl":"","password":"secret"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/users", body)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var payload db.User
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "user@example.com", payload.ID)
	})

	t.Run("Update user role", func(t *testing.T) {
		store := &fakeUserStore{updateUserAdminFn: func(context.Context, string, bool) error { return nil }}
		api := &testEndpoints{Logger: slog.Default(), Users: users.New(store, func() bool { return false })}
		h := UpdateUserRole(api.Users, api.Logger)
		body := strings.NewReader(`{"isAdmin":true}`)
		req := httptest.NewRequest(http.MethodPut, "/api/users/user@example.com/admin", body)
		req.SetPathValue("id", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Login", func(t *testing.T) {
		hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
		require.NoError(t, err)

		store := &fakeUserStore{getUserWithPasswordFn: func(context.Context, string) (*db.User, string, error) {
			return &db.User{ID: "user@example.com"}, string(hash), nil
		}}
		api := &testEndpoints{Logger: slog.Default(), Users: users.New(store, func() bool { return false })}
		h := Login(httpauth.NewLocal(store, time.Hour), func() bool { return true }, api.Logger)
		body := strings.NewReader(`{"email":"user@example.com","password":"secret"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/login", body)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload db.User
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "user@example.com", payload.ID)
	})

	t.Run("Change password", func(t *testing.T) {
		hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
		require.NoError(t, err)

		store := &fakeUserStore{
			getUserWithPasswordFn: func(context.Context, string) (*db.User, string, error) {
				return &db.User{ID: "user@example.com"}, string(hash), nil
			},
			updateUserPasswordFn: func(context.Context, string, string) error { return nil },
		}
		api := &testEndpoints{Logger: slog.Default()}
		h := ChangePassword(httpauth.NewLocal(store, time.Hour), api.Logger)
		body := strings.NewReader(`{"currentPassword":"secret","newPassword":"new"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/users/password", body)
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}
