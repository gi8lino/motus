package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainusers "github.com/gi8lino/motus/internal/domain/users"
	"github.com/gi8lino/motus/internal/service"
)

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{
			listUsersFn: func(context.Context) ([]domainusers.User, error) {
				return []domainusers.User{{ID: "u1"}}, nil
			},
		}, "", false)
		users, err := svc.List(context.Background())
		require.NoError(t, err)
		assert.Len(t, users, 1)
	})
}

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{}, "", false)
		_, err := svc.Get(context.Background(), " ")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorValidation))
	})
}
