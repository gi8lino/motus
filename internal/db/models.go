package db

import "time"

// User represents an account owner.
type User struct {
	ID        string    `json:"id"`        // ID is the unique user identifier.
	Name      string    `json:"name"`      // Name is the display name.
	IsAdmin   bool      `json:"isAdmin"`   // IsAdmin marks admin privileges.
	AvatarURL string    `json:"avatarUrl"` // AvatarURL is the optional avatar image.
	CreatedAt time.Time `json:"createdAt"` // CreatedAt records when the user was created.
}

// Workout groups stopwatch steps.
type Workout struct {
	ID         string        `json:"id"`         // ID is the unique workout identifier.
	UserID     string        `json:"userId"`     // UserID owns the workout.
	Name       string        `json:"name"`       // Name is the workout title.
	IsTemplate bool          `json:"isTemplate"` // IsTemplate marks shared templates.
	CreatedAt  time.Time     `json:"createdAt"`  // CreatedAt records when the workout was created.
	Steps      []WorkoutStep `json:"steps"`      // Steps defines the workout flow.
}

// PauseOptions captures optional behaviour for pause steps.
type PauseOptions struct {
	AutoAdvance bool `json:"autoAdvance,omitempty"` // AutoAdvance skips to the next step on completion.
}

// WorkoutStep defines a single part of the workout.
type WorkoutStep struct {
	ID                    string          `json:"id"`                              // ID is the unique step identifier.
	WorkoutID             string          `json:"workoutId"`                       // WorkoutID links to the parent workout.
	Order                 int             `json:"order"`                           // Order defines the step sequence.
	Type                  string          `json:"type"`                            // Type is set or pause.
	Name                  string          `json:"name"`                            // Name is the step label.
	EstimatedSeconds      int             `json:"estimatedSeconds"`                // EstimatedSeconds is the target duration.
	SoundKey              string          `json:"soundKey"`                        // SoundKey plays on step completion.
	Subsets               []WorkoutSubset `json:"subsets"`                         // Subsets hold exercises for set steps.
	PauseOptions          PauseOptions    `json:"pauseOptions,omitempty"`          // PauseOptions configure pause behavior.
	RepeatCount           int             `json:"repeatCount,omitempty"`           // RepeatCount repeats the step group.
	RepeatRestSeconds     int             `json:"repeatRestSeconds,omitempty"`     // RepeatRestSeconds is rest between repeats.
	RepeatRestAfterLast   bool            `json:"repeatRestAfterLast,omitempty"`   // RepeatRestAfterLast includes rest after last repeat.
	RepeatRestSoundKey    string          `json:"repeatRestSoundKey,omitempty"`    // RepeatRestSoundKey plays during repeat rest.
	RepeatRestAutoAdvance bool            `json:"repeatRestAutoAdvance,omitempty"` // RepeatRestAutoAdvance skips repeat rest automatically.
	RepeatRestName        string          `json:"repeatRestName,omitempty"`        // RepeatRestName overrides the repeat pause label.

	CreatedAt time.Time `json:"createdAt"` // CreatedAt records when the step was created.
}

// NormalizeRepeatSettings clamps and clears repeat rest settings for a step.
func (w *WorkoutStep) NormalizeRepeatSettings() {
	w.RepeatCount = max(w.RepeatCount, 1)
	w.RepeatRestSeconds = max(w.RepeatRestSeconds, 0)
	if w.RepeatCount <= 1 || w.RepeatRestSeconds == 0 {
		w.RepeatRestSeconds = 0
		w.RepeatRestAfterLast = false
		w.RepeatRestSoundKey = ""
		w.RepeatRestAutoAdvance = false
		w.RepeatRestName = ""
	}
}

type WorkoutSubset struct {
	ID               string           `json:"id"`               // ID is the unique subset identifier.
	StepID           string           `json:"stepId"`           // StepID links to the parent step.
	Order            int              `json:"order"`            // Order is the subset sequence within the step.
	Name             string           `json:"name"`             // Name is the subset label.
	EstimatedSeconds int              `json:"estimatedSeconds"` // EstimatedSeconds is the subset target.
	SoundKey         string           `json:"soundKey"`         // SoundKey plays at the subset target.
	Superset         bool             `json:"superset"`         // Superset skips to the next subset on Next.
	Exercises        []SubsetExercise `json:"exercises"`        // Exercises belong to the subset.
	CreatedAt        time.Time        `json:"createdAt"`        // CreatedAt records when the subset was created.
}

type SubsetExercise struct {
	ID         string `json:"id"`         // ID is the unique exercise row identifier.
	SubsetID   string `json:"subsetId"`   // SubsetID links to the parent subset.
	Order      int    `json:"order"`      // Order is the exercise sequence within the subset.
	ExerciseID string `json:"exerciseId"` // ExerciseID links to the catalog entry.
	Name       string `json:"name"`       // Name is the exercise label.
	Type       string `json:"type"`       // Type is rep, stopwatch, or countdown.
	Reps       string `json:"reps"`       // Reps is the repetition text.
	Weight     string `json:"weight"`     // Weight is optional load text.
	Duration   string `json:"duration"`   // Duration is a stopwatch/countdown value.
	SoundKey   string `json:"soundKey"`   // SoundKey overrides the subset sound.
}

// Exercise represents a reusable exercise catalog entry.
type Exercise struct {
	ID          string    `json:"id"`                    // ID is the unique catalog identifier.
	Name        string    `json:"name"`                  // Name is the exercise label.
	OwnerUserID string    `json:"ownerUserId,omitempty"` // OwnerUserID is set for user-owned entries.
	IsCore      bool      `json:"isCore"`                // IsCore marks built-in exercises.
	CreatedAt   time.Time `json:"createdAt"`             // CreatedAt records when the entry was created.
}

// TrainingLog represents a completed workout training.
type TrainingLog struct {
	ID          string    `json:"id"`          // ID is the unique training identifier.
	WorkoutID   string    `json:"workoutId"`   // WorkoutID links to the workout.
	WorkoutName string    `json:"workoutName"` // WorkoutName is the display name at completion time.
	UserID      string    `json:"userId"`      // UserID owns the training.
	StartedAt   time.Time `json:"startedAt"`   // StartedAt is when the training began.
	CompletedAt time.Time `json:"completedAt"` // CompletedAt is when the training finished.
}

// TrainingStepLog captures actual timing for a completed step.
type TrainingStepLog struct {
	ID               string `json:"id"`               // ID is the unique log row identifier.
	TrainingID       string `json:"trainingId"`       // TrainingID links to the training.
	StepOrder        int    `json:"stepOrder"`        // StepOrder preserves training ordering.
	Type             string `json:"type"`             // Type is the step kind.
	Name             string `json:"name"`             // Name is the step label.
	EstimatedSeconds int    `json:"estimatedSeconds"` // EstimatedSeconds is the target duration.
	ElapsedMillis    int64  `json:"elapsedMillis"`    // ElapsedMillis is the observed duration.
}
