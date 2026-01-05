package db

import (
	"context"
	"fmt"
	"strings"
)

// ListTemplates returns all workout templates.
func (s *Store) ListTemplates(ctx context.Context) ([]Workout, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, name, is_template, created_at
		FROM workouts
		WHERE is_template=TRUE
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []Workout
	for rows.Next() {
		var w Workout
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.IsTemplate, &w.CreatedAt); err != nil {
			return nil, err
		}
		steps, err := s.WorkoutSteps(ctx, w.ID)
		if err != nil {
			return nil, err
		}
		w.Steps = steps
		templates = append(templates, w)
	}
	return templates, rows.Err()
}

// CreateTemplateFromWorkout clones an existing workout as a template.
func (s *Store) CreateTemplateFromWorkout(ctx context.Context, workoutID string, nameOverride string) (*Workout, error) {
	src, err := s.WorkoutWithSteps(ctx, workoutID)
	if err != nil {
		return nil, err
	}
	if src.IsTemplate {
		return nil, fmt.Errorf("workout is already a template")
	}
	template := &Workout{
		UserID: src.UserID,
		Name:   src.Name,
		Steps:  cloneSteps(src.Steps),
	}
	if trimmed := strings.TrimSpace(nameOverride); trimmed != "" {
		template.Name = trimmed
	}
	return s.insertWorkout(ctx, template, true)
}

// CreateWorkoutFromTemplate copies an existing template to a user.
func (s *Store) CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*Workout, error) {
	template, err := s.WorkoutWithSteps(ctx, templateID)
	if err != nil {
		return nil, err
	}
	if !template.IsTemplate {
		return nil, fmt.Errorf("workout is not a template")
	}
	workout := &Workout{
		UserID: userID,
		Name:   strings.TrimSpace(name),
		Steps:  cloneSteps(template.Steps),
	}
	if workout.Name == "" {
		workout.Name = template.Name
	}
	return s.insertWorkout(ctx, workout, false)
}
