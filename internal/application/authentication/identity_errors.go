package authentication

import (
	"errors"
	"fmt"

	"github.com/containeroo/appkit/auth"
	"github.com/gi8lino/motus/internal/db"
)

// classifyExternalIdentityError identifies expected denials without masking infrastructure failures.
func classifyExternalIdentityError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, db.ErrIdentityRegistrationDisabled),
		errors.Is(err, db.ErrIdentityEmailInUse),
		errors.Is(err, db.ErrIdentityEmailInvalid):
		return fmt.Errorf("%w: %w", auth.ErrForbidden, err)
	default:
		return err
	}
}
