package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
)

func TestUsersHandlers(t *testing.T) {
	t.Run("List users", func(t *testing.T) {
		store := &fakeStore{listUsersFn: func(context.Context) ([]db.User, error) {
			return []db.User{{ID: "user@example.com"}}, nil
		}}
		api := &API{Store: store}
		h := api.GetUsers()
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
		store := &fakeStore{createUserFn: func(context.Context, string, string, string) (*db.User, error) {
			return &db.User{ID: "user@example.com"}, nil
		}}
		api := &API{Store: store, AllowRegistration: true}
		h := api.CreateUser()
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
		store := &fakeStore{updateUserAdminFn: func(context.Context, string, bool) error { return nil }}
		api := &API{Store: store}
		h := api.UpdateUserRole()
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

		store := &fakeStore{getUserWithPasswordFn: func(context.Context, string) (*db.User, string, error) {
			return &db.User{ID: "user@example.com"}, string(hash), nil
		}}
		api := &API{Store: store}
		h := api.Login()
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

		store := &fakeStore{
			getUserWithPasswordFn: func(context.Context, string) (*db.User, string, error) {
				return &db.User{ID: "user@example.com"}, string(hash), nil
			},
			updateUserPasswordFn: func(context.Context, string, string) error { return nil },
		}
		api := &API{Store: store}
		h := api.ChangePassword()
		body := strings.NewReader(`{"currentPassword":"secret","newPassword":"new"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/users/password", body)
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}
