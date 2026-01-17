package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

type fakeStore struct {
	createUserFn      func(context.Context, string, string, string) (*User, error)
	updateUserAdminFn func(context.Context, string, bool) error
	getUserWithPassFn func(context.Context, string) (*User, string, error)
	updateUserPassFn  func(context.Context, string, string) error
	updateUserNameFn  func(context.Context, string, string) error
	getUserFn         func(context.Context, string) (*User, error)
	listUsersFn       func(context.Context) ([]User, error)
}

func (f *fakeStore) ListUsers(ctx context.Context) ([]User, error) {
	if f.listUsersFn == nil {
		return nil, nil
	}
	return f.listUsersFn(ctx)
}

func (f *fakeStore) CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*User, error) {
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

func (f *fakeStore) GetUserWithPassword(ctx context.Context, id string) (*User, string, error) {
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
func (f *fakeStore) GetUser(ctx context.Context, id string) (*User, error) {
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
		assert.True(t, errpkg.IsKind(err, errpkg.ErrorForbidden))
	})

	t.Run("Password required", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{}, "", true)
		_, err := svc.Create(context.Background(), "user@example.com", "", " ")
		require.Error(t, err)
		assert.True(t, errpkg.IsKind(err, errpkg.ErrorValidation))
	})

	t.Run("Proxy auth allows empty password", func(t *testing.T) {
		t.Parallel()

		called := false
		svc := New(&fakeStore{
			createUserFn: func(context.Context, string, string, string) (*User, error) {
				called = true
				return &User{ID: "user"}, nil
			},
		}, "X-User", false)
		user, err := svc.Create(context.Background(), "user@example.com", "", "")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "user", user.ID)
		assert.True(t, called, "expected CreateUser to be called")
	})
}

func TestUpdateRole(t *testing.T) {
	t.Parallel()

	t.Run("UpdatesAdminFlag", func(t *testing.T) {
		t.Parallel()

		called := false
		svc := New(&fakeStore{
			updateUserAdminFn: func(context.Context, string, bool) error {
				called = true
				return nil
			},
		}, "", false)
		if err := svc.UpdateRole(context.Background(), "user", true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatalf("expected UpdateUserAdmin to be called")
		}
	})
}

func TestUpdateName(t *testing.T) {
	t.Parallel()

	t.Run("UpdatesName", func(t *testing.T) {
		t.Parallel()

		called := false
		svc := New(&fakeStore{
			updateUserNameFn: func(context.Context, string, string) error {
				called = true
				return nil
			},
		}, "", false)
		if err := svc.UpdateName(context.Background(), "user", "New"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatalf("expected UpdateUserName to be called")
		}
	})
}
