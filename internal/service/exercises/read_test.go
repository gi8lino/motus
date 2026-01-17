package exercises

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("Validation error", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{})
		_, err := svc.List(context.Background(), " ")
		require.Error(t, err)
		assert.True(t, errpkg.IsKind(err, errpkg.ErrorValidation))
	})
}
