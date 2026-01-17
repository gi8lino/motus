package workouts

import (
	"context"
	"strings"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// Import creates a new workout from exported JSON.
func (s *Service) Import(ctx context.Context, userID string, workout Workout) (*Workout, error) {
	userID = strings.TrimSpace(userID)
	workout.Name = strings.TrimSpace(workout.Name)
	if userID == "" {
		return nil, errpkg.NewError(errpkg.ErrorValidation, "userId is required")
	}
	if workout.Name == "" || len(workout.Steps) == 0 {
		return nil, errpkg.NewError(errpkg.ErrorValidation, "name and steps are required")
	}

	for idx := range workout.Steps {
		step := &workout.Steps[idx]
		step.ID = ""
		step.WorkoutID = ""
		step.Order = idx
		for subIdx := range step.Subsets {
			sub := &step.Subsets[subIdx]
			sub.ID = ""
			sub.StepID = ""
			sub.Order = subIdx
			for exIdx := range sub.Exercises {
				ex := &sub.Exercises[exIdx]
				ex.ID = ""
				ex.SubsetID = ""
				ex.Order = exIdx
				ex.ExerciseID = ""
			}
		}
	}
	created, err := s.store.CreateWorkout(ctx, &Workout{
		UserID: userID,
		Name:   workout.Name,
		Steps:  workout.Steps,
	})
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}

	return created, nil
}
