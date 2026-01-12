package workouts

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/gi8lino/motus/internal/domain/sounds"
)

// Manager orchestrates workout domain rules.
type Manager struct {
	store Store
}

// NewManager builds a workout Manager.
func NewManager(store Store) *Manager {
	return &Manager{store: store}
}

// Create stores a new workout for the user.
func (m *Manager) Create(ctx context.Context, req WorkoutRequest) (*Workout, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.Name = strings.TrimSpace(req.Name)
	if req.UserID == "" || req.Name == "" || len(req.Steps) == 0 {
		return nil, validation("name and at least one step are required")
	}

	steps, err := NormalizeSteps(req.Steps, sounds.ValidKey)
	if err != nil {
		return nil, validation(err.Error())
	}

	workout := &Workout{UserID: req.UserID, Name: req.Name, Steps: steps}
	created, err := m.store.CreateWorkout(ctx, workout)
	if err != nil {
		return nil, internal(err)
	}

	return created, nil
}

// Update replaces a workout and its steps.
func (m *Manager) Update(ctx context.Context, id string, req WorkoutRequest) (*Workout, error) {
	id = strings.TrimSpace(id)
	req.UserID = strings.TrimSpace(req.UserID)
	req.Name = strings.TrimSpace(req.Name)
	if id == "" {
		return nil, validation("workout id is required")
	}
	if req.Name == "" || len(req.Steps) == 0 {
		return nil, validation("name and steps are required")
	}

	steps, err := NormalizeSteps(req.Steps, sounds.ValidKey)
	if err != nil {
		return nil, validation(err.Error())
	}

	workout := &Workout{ID: id, UserID: req.UserID, Name: req.Name, Steps: steps}
	updated, err := m.store.UpdateWorkout(ctx, workout)
	if err != nil {
		return nil, internal(err)
	}

	return updated, nil
}

// Delete removes a workout by id.
func (m *Manager) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return validation("workout id is required")
	}

	if err := m.store.DeleteWorkout(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) || isNotFoundError(err) {
			return notFound("workout not found")
		}
		return internal(err)
	}

	return nil
}

// List returns workouts for the given user.
func (m *Manager) List(ctx context.Context, userID string) ([]Workout, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, validation("userId is required")
	}
	workouts, err := m.store.WorkoutsByUser(ctx, userID)
	if err != nil {
		return nil, internal(err)
	}
	return workouts, nil
}

// Get fetches a workout by id.
func (m *Manager) Get(ctx context.Context, id string) (*Workout, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, validation("workout id is required")
	}
	workout, err := m.store.WorkoutWithSteps(ctx, id)
	if err != nil {
		if isNotFoundError(err) {
			return nil, notFound("workout not found")
		}
		return nil, internal(err)
	}
	if workout == nil {
		return nil, notFound("workout not found")
	}
	return workout, nil
}

// Export returns a workout for sharing.
func (m *Manager) Export(ctx context.Context, id string) (*Workout, error) {
	return m.Get(ctx, id)
}

// Import creates a new workout from exported JSON.
func (m *Manager) Import(ctx context.Context, userID string, workout Workout) (*Workout, error) {
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
	created, err := m.store.CreateWorkout(ctx, &Workout{
		UserID: userID,
		Name:   workout.Name,
		Steps:  workout.Steps,
	})
	if err != nil {
		return nil, internal(err)
	}

	return created, nil
}
