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
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	return exercises, nil
}
