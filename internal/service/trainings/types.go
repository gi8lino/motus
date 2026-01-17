// Package trainings provides domain types and training state builders.
package trainings

import (
	"time"

	"github.com/gi8lino/motus/internal/db"
)

// Workout is the domain-level DTO for training workouts.
type Workout = db.Workout

// WorkoutStep is the domain-level DTO for workout steps.
type WorkoutStep = db.WorkoutStep

// WorkoutSubset is the domain-level DTO for workout subsets.
type WorkoutSubset = db.WorkoutSubset

// SubsetExercise is the domain-level DTO for subset exercises.
type SubsetExercise = db.SubsetExercise

// PauseOptions controls auto-advance for pauses.
type PauseOptions = db.PauseOptions

// TrainingLog is the domain-level DTO for completed training logs.
type TrainingLog = db.TrainingLog

// TrainingStepLog is the domain-level DTO for training step timing logs.
type TrainingStepLog = db.TrainingStepLog

// TrainingState captures the runtime status that the SPA consumes for an active training.
type TrainingState struct {
	TrainingID   string              `json:"trainingId"`
	WorkoutID    string              `json:"workoutId"`
	UserID       string              `json:"userId"`
	WorkoutName  string              `json:"workoutName"`
	CurrentIndex int                 `json:"currentIndex"`
	Running      bool                `json:"running"`
	Done         bool                `json:"done"`
	StartedAt    time.Time           `json:"startedAt"`
	CompletedAt  time.Time           `json:"completedAt"`
	Steps        []TrainingStepState `json:"steps"`
}

// TrainingStepState describes a single card/step inside a training view.
type TrainingStepState struct {
	ID                     string       `json:"id"`
	Name                   string       `json:"name"`
	Type                   string       `json:"type"`
	EstimatedSeconds       int          `json:"estimatedSeconds"`
	SoundURL               string       `json:"soundUrl"`
	SoundKey               string       `json:"soundKey,omitempty"`
	SubsetEstimatedSeconds int          `json:"subsetEstimatedSeconds,omitempty"`
	Running                bool         `json:"running"`
	Completed              bool         `json:"completed"`
	Current                bool         `json:"current"`
	ElapsedMillis          int64        `json:"elapsedMillis"`
	Exercises              []Exercise   `json:"exercises"`
	PauseOptions           PauseOptions `json:"pauseOptions"`
	AutoAdvance            bool         `json:"autoAdvance"`
	LoopIndex              int          `json:"loopIndex,omitempty"`
	LoopTotal              int          `json:"loopTotal,omitempty"`
	SubsetID               string       `json:"subsetId,omitempty"`
	Superset               bool         `json:"superset,omitempty"`
	SubsetLabel            string       `json:"subsetLabel,omitempty"`
	HasMultipleSubsets     bool         `json:"hasMultipleSubsets,omitempty"`
	SetName                string       `json:"setName,omitempty"`
}

// Exercise represents a configured exercise inside a training step.
type Exercise struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Reps     string `json:"reps"`
	Weight   string `json:"weight"`
	Duration string `json:"duration"`
	SoundKey string `json:"soundKey,omitempty"`
}

// TrainingHistoryItem is the API payload for a completed training.
type TrainingHistoryItem struct {
	ID          string            `json:"id"`                    // ID is the history item identifier.
	TrainingID  string            `json:"trainingId"`            // TrainingID links to the completed training.
	WorkoutID   string            `json:"workoutId"`             // WorkoutID references the workout definition.
	WorkoutName string            `json:"workoutName"`           // WorkoutName is the display name at completion time.
	UserID      string            `json:"userId"`                // UserID owns the training.
	StartedAt   *time.Time        `json:"startedAt,omitempty"`   // StartedAt is when the training began.
	CompletedAt *time.Time        `json:"completedAt,omitempty"` // CompletedAt is when the training finished.
	Steps       []TrainingStepLog `json:"steps,omitempty"`       // Steps contains logged timings when available.
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
