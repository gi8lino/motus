package users_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/domain/users"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	t.Run("ProxyDisabled", func(t *testing.T) {
		t.Parallel()
		mgr := users.NewManager(&fakeUsersStore{}, "X", false)
		_, err := mgr.Login(context.Background(), "user@example.com", "secret")
		var domainErr *users.Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, users.KindForbidden, domainErr.Kind)
	})
}

func TestChangePassword(t *testing.T) {
	t.Parallel()

	t.Run("Allowed", func(t *testing.T) {
		t.Parallel()
		mgr := users.NewManager(&fakeUsersStore{
			getUserWithPassFn: func(context.Context, string) (*db.User, string, error) {
				return &db.User{ID: "user"}, "hash", nil
			},
			updateUserPassFn: func(context.Context, string, string) error {
				return nil
			},
		}, "", false)
		err := mgr.ChangePassword(context.Background(), "user", "current", "new")
		assert.Error(t, err)
		var domainErr *users.Error
		assert.ErrorAs(t, err, &domainErr)
	})
}
