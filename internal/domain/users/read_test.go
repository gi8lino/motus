package users_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/domain/users"
)

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		mgr := users.NewManager(&fakeUsersStore{
			listUsersFn: func(context.Context) ([]db.User, error) {
				return []db.User{{ID: "u1"}}, nil
			},
		}, "", false)
		usersList, err := mgr.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, usersList, 1)
	})
}
