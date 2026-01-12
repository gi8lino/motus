// Package workouts provides service access to workout operations.
package workouts

import domain "github.com/gi8lino/motus/internal/domain/workouts"

// Store exposes the persistence requirements for workout services.
type Store = domain.Store

// WorkoutRequest captures the payload for creating or updating workouts.
type WorkoutRequest = domain.WorkoutRequest

// Workout is the domain-level DTO for workouts.
type Workout = domain.Workout

// WorkoutStep is the domain-level DTO for workout steps.
type WorkoutStep = domain.WorkoutStep

// WorkoutSubset is the domain-level DTO for workout subsets.
type WorkoutSubset = domain.WorkoutSubset

// SubsetExercise is the domain-level DTO for subset exercises.
type SubsetExercise = domain.SubsetExercise

// StepInput mirrors the workout step input payload.
type StepInput = domain.StepInput

// SubsetInput mirrors the workout subset input payload.
type SubsetInput = domain.SubsetInput

// ExerciseInput mirrors the workout exercise input payload.
type ExerciseInput = domain.ExerciseInput

// Service coordinates workout operations.
type Service struct {
	manager *domain.Manager
}

// New creates a new workouts service.
func New(store Store) *Service {
	return &Service{manager: domain.NewManager(store)}
}
