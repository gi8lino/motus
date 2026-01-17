package exercises

import (
	"context"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// List returns the exercise catalog for a user.
func (s *Service) List(ctx context.Context, userID string) ([]Exercise, error) {
	uid, err := requireUserID(userID)
	if err != nil {
		return nil, err
	}
	exercises, err := s.store.ListExercises(ctx, uid)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return exercises, nil
}
