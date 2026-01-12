package users_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/domain/users"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("Disabled", func(t *testing.T) {
		t.Parallel()
		mgr := users.NewManager(&fakeUsersStore{}, "", false)
		_, err := mgr.Create(context.Background(), "user@example.com", "", "secret")
		var domainErr *users.Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, users.KindForbidden, domainErr.Kind)
	})
}
