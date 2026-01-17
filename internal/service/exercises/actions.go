package exercises

import (
	"context"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// Backfill rebuilds core exercises from workout data.
func (s *Service) Backfill(ctx context.Context) error {
	if err := s.store.BackfillCoreExercises(ctx); err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return nil
}
