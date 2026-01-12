package users

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainusers "github.com/gi8lino/motus/internal/domain/users"
	"github.com/gi8lino/motus/internal/service"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	t.Run("Disabled with auth header", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{}, "X-User", false)
		_, err := svc.Login(context.Background(), "user@example.com", "secret")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorForbidden))
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{
			getUserWithPassFn: func(context.Context, string) (*domainusers.User, string, error) {
				return nil, "", nil
			},
		}, "", false)
		_, err := svc.Login(context.Background(), "user@example.com", "secret")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorUnauthorized))
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
		require.NoError(t, err)

		svc := New(&fakeStore{
			getUserWithPassFn: func(context.Context, string) (*domainusers.User, string, error) {
				return &domainusers.User{ID: "user"}, string(hash), nil
			},
		}, "", false)
		user, err := svc.Login(context.Background(), "user@example.com", "secret")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "user", user.ID)
	})
}

func TestChangePassword(t *testing.T) {
	t.Parallel()

	t.Run("Proxy managed", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{}, "X-User", false)
		err := svc.ChangePassword(context.Background(), "user", "old", "new")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorForbidden))
	})

	t.Run("Invalid current password", func(t *testing.T) {
		t.Parallel()

		hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
		require.NoError(t, err)

		svc := New(&fakeStore{
			getUserWithPassFn: func(context.Context, string) (*domainusers.User, string, error) {
				return &domainusers.User{ID: "user"}, string(hash), nil
			},
		}, "", false)
		err = svc.ChangePassword(context.Background(), "user", "wrong", "new")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorUnauthorized))
	})

	t.Run("Updates hash", func(t *testing.T) {
		t.Parallel()

		hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
		require.NoError(t, err)

		called := false
		svc := New(&fakeStore{
			getUserWithPassFn: func(context.Context, string) (*domainusers.User, string, error) {
				return &domainusers.User{ID: "user"}, string(hash), nil
			},
			updateUserPassFn: func(context.Context, string, string) error {
				called = true
				return nil
			},
		}, "", false)
		err = svc.ChangePassword(context.Background(), "user", "secret", "new")
		require.NoError(t, err)
		assert.True(t, called)
	})
}
