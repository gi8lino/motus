package authentication

import (
	"errors"
	"fmt"
	"testing"

	appkitauth "github.com/containeroo/appkit/auth"
	"github.com/gi8lino/motus/internal/db"
	"github.com/stretchr/testify/require"
)

// TestClassifyExternalIdentityError preserves typed identity denials and infrastructure failures.
func TestClassifyExternalIdentityError(t *testing.T) {
	t.Parallel()
	for _, err := range []error{
		db.ErrIdentityRegistrationDisabled,
		db.ErrIdentityEmailInUse,
		db.ErrIdentityEmailInvalid,
	} {
		result := classifyExternalIdentityError(fmt.Errorf("lookup: %w", err))
		require.ErrorIs(t, result, appkitauth.ErrForbidden)
		require.ErrorIs(t, result, err)
	}
	databaseErr := errors.New("connection failed")
	require.ErrorIs(t, classifyExternalIdentityError(databaseErr), databaseErr)
	require.NotErrorIs(t, classifyExternalIdentityError(databaseErr), appkitauth.ErrForbidden)
	require.NoError(t, classifyExternalIdentityError(nil))
}
