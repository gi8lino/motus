package users

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service"
)

type fakeStore struct {
	createUserFn      func(context.Context, string, string, string) (*db.User, error)
	updateUserAdminFn func(context.Context, string, bool) error
	getUserWithPassFn func(context.Context, string) (*db.User, string, error)
	updateUserPassFn  func(context.Context, string, string) error
	updateUserNameFn  func(context.Context, string, string) error
	getUserFn         func(context.Context, string) (*db.User, error)
	listUsersFn       func(context.Context) ([]db.User, error)
}

func (f *fakeStore) ListUsers(ctx context.Context) ([]db.User, error) {
	if f.listUsersFn == nil {
		return nil, nil
	}
	return f.listUsersFn(ctx)
}

func (f *fakeStore) CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error) {
	if f.createUserFn == nil {
		return nil, nil
	}
	return f.createUserFn(ctx, email, avatarURL, passwordHash)
}

func (f *fakeStore) UpdateUserAdmin(ctx context.Context, id string, isAdmin bool) error {
	if f.updateUserAdminFn == nil {
		return nil
	}
	return f.updateUserAdminFn(ctx, id, isAdmin)
}

func (f *fakeStore) GetUserWithPassword(ctx context.Context, id string) (*db.User, string, error) {
	if f.getUserWithPassFn == nil {
		return nil, "", nil
	}
	return f.getUserWithPassFn(ctx, id)
}

func (f *fakeStore) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	if f.updateUserPassFn == nil {
		return nil
	}
	return f.updateUserPassFn(ctx, id, passwordHash)
}

func (f *fakeStore) UpdateUserName(ctx context.Context, id, name string) error {
	if f.updateUserNameFn == nil {
		return nil
	}
	return f.updateUserNameFn(ctx, id, name)
}

// GetUser is required by the users.Store interface.
// Adjust the parameter name/type here only if your Store.GetUser signature differs.
func (f *fakeStore) GetUser(ctx context.Context, id string) (*db.User, error) {
	if f.getUserFn == nil {
		return nil, nil
	}
	return f.getUserFn(ctx, id)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("Registration disabled", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{}, "", false)
		_, err := svc.Create(context.Background(), "user@example.com", "", "secret")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorForbidden))
	})

	t.Run("Password required", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{}, "", true)
		_, err := svc.Create(context.Background(), "user@example.com", "", " ")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorValidation))
	})

	t.Run("Proxy auth allows empty password", func(t *testing.T) {
		t.Parallel()

		called := false
		svc := New(&fakeStore{
			createUserFn: func(context.Context, string, string, string) (*db.User, error) {
				called = true
				return &db.User{ID: "user"}, nil
			},
		}, "X-User", false)
		user, err := svc.Create(context.Background(), "user@example.com", "", "")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "user", user.ID)
		assert.True(t, called, "expected CreateUser to be called")
	})
}

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
			getUserWithPassFn: func(context.Context, string) (*db.User, string, error) {
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
			getUserWithPassFn: func(context.Context, string) (*db.User, string, error) {
				return &db.User{ID: "user"}, string(hash), nil
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
			getUserWithPassFn: func(context.Context, string) (*db.User, string, error) {
				return &db.User{ID: "user"}, string(hash), nil
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
			getUserWithPassFn: func(context.Context, string) (*db.User, string, error) {
				return &db.User{ID: "user"}, string(hash), nil
			},
			updateUserPassFn: func(context.Context, string, string) error {
				called = true
				return nil
			},
		}, "", false)
		err = svc.ChangePassword(context.Background(), "user", "secret", "newpass")
		require.NoError(t, err)
		assert.True(t, called, "expected UpdateUserPassword to be called")
	})
}

func TestUpdateRole(t *testing.T) {
	t.Parallel()

	t.Run("Validation error", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{}, "", false)
		err := svc.UpdateRole(context.Background(), " ", true)
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorValidation))
	})

	t.Run("Internal error", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{
			updateUserAdminFn: func(context.Context, string, bool) error {
				return errors.New("boom")
			},
		}, "", false)
		err := svc.UpdateRole(context.Background(), "user", true)
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorInternal))
	})
}
