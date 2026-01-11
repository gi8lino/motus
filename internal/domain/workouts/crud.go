package workouts

import (
	"context"
	"errors"
	"strings"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/domain/sounds"
	"github.com/jackc/pgx/v5"
)

// Create stores a new workout for the user.
func (m *Manager) Create(ctx context.Context, req WorkoutRequest) (*db.Workout, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.Name = strings.TrimSpace(req.Name)
	if req.UserID == "" || req.Name == "" || len(req.Steps) == 0 {
		return nil, validation("name and at least one step are required")
	}

	steps, err := NormalizeSteps(req.Steps, sounds.ValidKey)
	if err != nil {
		return nil, validation(err.Error())
	}

	workout := &db.Workout{UserID: req.UserID, Name: req.Name, Steps: steps}
	created, err := m.store.CreateWorkout(ctx, workout)
	if err != nil {
		return nil, internal(err)
	}

	return created, nil
}

// Update replaces a workout and its steps.
func (m *Manager) Update(ctx context.Context, id string, req WorkoutRequest) (*db.Workout, error) {
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

	workout := &db.Workout{ID: id, UserID: req.UserID, Name: req.Name, Steps: steps}
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
