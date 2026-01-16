// Package trainings provides service access to workout trainings.
package trainings

import (
	"time"

	domainTrains "github.com/gi8lino/motus/internal/domain/trainings"
)

// TrainingHistoryItem is the API payload for a completed training.
type TrainingHistoryItem struct {
	ID          string                         `json:"id"`                    // ID is the history item identifier.
	TrainingID  string                         `json:"trainingId"`            // TrainingID links to the completed training.
	WorkoutID   string                         `json:"workoutId"`             // WorkoutID references the workout definition.
	WorkoutName string                         `json:"workoutName"`           // WorkoutName is the display name at completion time.
	UserID      string                         `json:"userId"`                // UserID owns the training.
	StartedAt   *time.Time                     `json:"startedAt,omitempty"`   // StartedAt is when the training began.
	CompletedAt *time.Time                     `json:"completedAt,omitempty"` // CompletedAt is when the training finished.
	Steps       []domainTrains.TrainingStepLog `json:"steps,omitempty"`       // Steps contains logged timings when available.
}

// Store exposes the persistence requirements for training services.
type Store = domainTrains.Store

type (
	TrainingState     = domainTrains.TrainingState
	TrainingStepState = domainTrains.TrainingStepState
	Exercise          = domainTrains.Exercise
	PauseOptions      = domainTrains.PauseOptions
	TrainingLog       = domainTrains.TrainingLog
	TrainingStepLog   = domainTrains.TrainingStepLog
	Workout           = domainTrains.Workout
	WorkoutStep       = domainTrains.WorkoutStep
)

// Service coordinates training operations.
type Service struct {
	store         Store
	soundURLByKey func(string) string
}

// New creates a new trainings service.
func New(store Store, soundURLByKey func(string) string) *Service {
	return &Service{store: store, soundURLByKey: soundURLByKey}
}

// CompleteRequest captures the payload for logging a finished training.
type CompleteRequest struct {
	TrainingID  string              `json:"trainingId"`  // TrainingID identifies the training.
	WorkoutID   string              `json:"workoutId"`   // WorkoutID identifies the workout.
	WorkoutName string              `json:"workoutName"` // WorkoutName is the display name at completion time.
	UserID      string              `json:"userId"`      // UserID owns the training.
	StartedAt   time.Time           `json:"startedAt"`   // StartedAt records when the training began.
	CompletedAt time.Time           `json:"completedAt"` // CompletedAt records when the training finished.
	Steps       []TrainingStepState `json:"steps"`       // Steps includes timing details.
}
