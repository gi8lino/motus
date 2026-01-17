// Package workouts provides domain logic for workout definitions.
package workouts

import "github.com/gi8lino/motus/internal/db"

// Workout is the domain-level DTO for workouts.
type Workout = db.Workout

// PauseOptions is the domain-level DTO for pause configuration.
type PauseOptions = db.PauseOptions

// WorkoutStep is the domain-level DTO for workout steps.
type WorkoutStep = db.WorkoutStep

// WorkoutSubset is the domain-level DTO for workout subsets.
type WorkoutSubset = db.WorkoutSubset

// SubsetExercise is the domain-level DTO for subset exercises.
type SubsetExercise = db.SubsetExercise

// errorScope is the service error scope for workouts.
const errorScope = "workouts"

// WorkoutRequest describes the payload for building a workout definition.
type WorkoutRequest struct {
	UserID string      `json:"userId"`
	Name   string      `json:"name"`
	Steps  []StepInput `json:"steps"`
}

// StepInput describes a workout step definition in the domain model.
type StepInput struct {
	Type                  string        `json:"type"`
	Name                  string        `json:"name"`
	Duration              string        `json:"duration"`
	EstimatedSeconds      int           `json:"estimatedSeconds"`
	SoundKey              string        `json:"soundKey"`
	Subsets               []SubsetInput `json:"subsets"`
	PauseOptions          PauseOptions  `json:"pauseOptions"`
	RepeatCount           int           `json:"repeatCount"`
	RepeatRestSeconds     int           `json:"repeatRestSeconds"`
	RepeatRestAfterLast   bool          `json:"repeatRestAfterLast"`
	RepeatRestSoundKey    string        `json:"repeatRestSoundKey"`
	RepeatRestAutoAdvance bool          `json:"repeatRestAutoAdvance"`
}

// SubsetInput describes a logical subset inside a set step.
type SubsetInput struct {
	Name      string          `json:"name"`
	Duration  string          `json:"duration"`
	SoundKey  string          `json:"soundKey"`
	Superset  bool            `json:"superset"`
	Exercises []ExerciseInput `json:"exercises"`
}

// ExerciseInput describes an exercise entry inside a subset definition.
type ExerciseInput struct {
	ExerciseID string `json:"exerciseId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Reps       string `json:"reps"`
	Weight     string `json:"weight"`
	Duration   string `json:"duration"`
	SoundKey   string `json:"soundKey"`
}
