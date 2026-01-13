package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/db"
)

type fakeUsersStore struct {
	getUserFn         func(context.Context, string) (*db.User, error)
	getUserWithPassFn func(context.Context, string) (*db.User, string, error)
	updateUserAdminFn func(context.Context, string, bool) error
	updateUserPassFn  func(context.Context, string, string) error
	updateUserNameFn  func(context.Context, string, string) error
	createUserFn      func(context.Context, string, string, string) (*db.User, error)
	listUsersFn       func(context.Context) ([]db.User, error)
}

func (f *fakeUsersStore) ListUsers(ctx context.Context) ([]db.User, error) {
	if f.listUsersFn == nil {
		return nil, nil
	}
	return f.listUsersFn(ctx)
}

func (f *fakeUsersStore) GetUser(ctx context.Context, id string) (*db.User, error) {
	if f.getUserFn == nil {
		return nil, nil
	}
	return f.getUserFn(ctx, id)
}

func (f *fakeUsersStore) CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error) {
	if f.createUserFn == nil {
		return nil, nil
	}
	return f.createUserFn(ctx, email, avatarURL, passwordHash)
}

func (f *fakeUsersStore) UpdateUserAdmin(ctx context.Context, id string, isAdmin bool) error {
	if f.updateUserAdminFn == nil {
		return nil
	}
	return f.updateUserAdminFn(ctx, id, isAdmin)
}

func (f *fakeUsersStore) GetUserWithPassword(ctx context.Context, id string) (*db.User, string, error) {
	if f.getUserWithPassFn == nil {
		return nil, "", nil
	}
	return f.getUserWithPassFn(ctx, id)
}

func (f *fakeUsersStore) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	if f.updateUserPassFn == nil {
		return nil
	}
	return f.updateUserPassFn(ctx, id, passwordHash)
}

func (f *fakeUsersStore) UpdateUserName(ctx context.Context, id, name string) error {
	if f.updateUserNameFn == nil {
		return nil
	}
	return f.updateUserNameFn(ctx, id, name)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("Disabled", func(t *testing.T) {
		t.Parallel()
		mgr := NewManager(&fakeUsersStore{}, "", false)
		_, err := mgr.Create(context.Background(), "user@example.com", "", "secret")
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindForbidden, domainErr.Kind)
	})
}

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		mgr := NewManager(&fakeUsersStore{
			listUsersFn: func(context.Context) ([]db.User, error) {
				return []db.User{{ID: "u1"}}, nil
			},
		}, "", false)
		usersList, err := mgr.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, usersList, 1)
	})
}

func TestLogin(t *testing.T) {
	t.Parallel()

	t.Run("ProxyDisabled", func(t *testing.T) {
		t.Parallel()
		mgr := NewManager(&fakeUsersStore{}, "X", false)
		_, err := mgr.Login(context.Background(), "user@example.com", "secret")
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindForbidden, domainErr.Kind)
	})
}

func TestChangePassword(t *testing.T) {
	t.Parallel()

	t.Run("Allowed", func(t *testing.T) {
		t.Parallel()
		mgr := NewManager(&fakeUsersStore{
			getUserWithPassFn: func(context.Context, string) (*db.User, string, error) {
				return &db.User{ID: "user"}, "hash", nil
			},
			updateUserPassFn: func(context.Context, string, string) error {
				return nil
			},
		}, "", false)
		err := mgr.ChangePassword(context.Background(), "user", "current", "new")
		assert.Error(t, err)
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
	})
}
