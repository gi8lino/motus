package templates

import (
	"context"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// Apply clones a template into a new workout.
func (s *Service) Apply(ctx context.Context, templateID, userID, name string) (*Workout, error) {
	tid, err := requireID(templateID, "template id is required")
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}
	uid, err := requireID(userID, "userId is required")
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}

	workout, err := s.store.CreateWorkoutFromTemplate(ctx, tid, uid, name)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return workout, nil
}
