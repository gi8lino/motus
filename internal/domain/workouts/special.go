package workouts

import (
	"context"
	"strings"

	"github.com/gi8lino/motus/internal/db"
)

// Import creates a new workout from exported JSON.
func (m *Manager) Import(ctx context.Context, userID string, workout db.Workout) (*db.Workout, error) {
	userID = strings.TrimSpace(userID)
	workout.Name = strings.TrimSpace(workout.Name)
	if userID == "" {
		return nil, validation("userId is required")
	}
	if workout.Name == "" || len(workout.Steps) == 0 {
		return nil, validation("workout name and steps are required")
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
	created, err := m.store.CreateWorkout(ctx, &db.Workout{
		UserID: userID,
		Name:   workout.Name,
		Steps:  workout.Steps,
	})
	if err != nil {
		return nil, internal(err)
	}

	return created, nil
}
