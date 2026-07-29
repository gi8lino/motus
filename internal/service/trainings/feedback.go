package trainings

import (
	"context"
	"strings"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// UpdateFeedback validates and stores post-training notes and effort.
func (s *Service) UpdateFeedback(ctx context.Context, trainingID, userID, notes string, effort *int) error {
	trainingID = strings.TrimSpace(trainingID)
	userID = strings.TrimSpace(userID)
	notes = strings.TrimSpace(notes)
	if trainingID == "" || userID == "" {
		return errpkg.NewErrorWithScope(errpkg.ErrorValidation, "trainingId and userId are required", errorScope)
	}
	if effort != nil && (*effort < 1 || *effort > 10) {
		return errpkg.NewErrorWithScope(errpkg.ErrorValidation, "perceivedEffort must be between 1 and 10", errorScope)
	}
	if err := s.store.UpdateTrainingFeedback(ctx, trainingID, userID, notes, effort); err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return nil
}
