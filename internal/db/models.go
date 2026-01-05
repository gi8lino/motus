package db

import "time"

// User represents an account owner.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IsAdmin   bool      `json:"isAdmin"`
	AvatarURL string    `json:"avatarUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

// Workout groups stopwatch steps.
type Workout struct {
	ID         string        `json:"id"`
	UserID     string        `json:"userId"`
	Name       string        `json:"name"`
	IsTemplate bool          `json:"isTemplate"`
	CreatedAt  time.Time     `json:"createdAt"`
	Steps      []WorkoutStep `json:"steps"`
}

// PauseOptions captures optional behaviour for pause steps.
type PauseOptions struct {
	AutoAdvance bool `json:"autoAdvance,omitempty"`
}

// WorkoutStep defines a single part of the workout.
type WorkoutStep struct {
	ID               string         `json:"id"`
	WorkoutID        string         `json:"workoutId"`
	Order            int            `json:"order"`
	Type             string         `json:"type"`
	Name             string         `json:"name"`
	EstimatedSeconds int            `json:"estimatedSeconds"`
	SoundKey         string         `json:"soundKey"`
	Exercise         string         `json:"exercise"`
	Amount           string         `json:"amount"`
	Weight           string         `json:"weight"`
	Exercises        []StepExercise `json:"exercises"`
	PauseOptions     PauseOptions   `json:"pauseOptions,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
}

// StepExercise represents a detailed exercise entry for a step.
type StepExercise struct {
	ID         string `json:"id"`
	StepID     string `json:"stepId"`
	Order      int    `json:"order"`
	ExerciseID string `json:"exerciseId"`
	Name       string `json:"name"`
	Amount     string `json:"amount"`
	Weight     string `json:"weight"`
}

// Exercise represents a reusable exercise catalog entry.
type Exercise struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OwnerUserID string    `json:"ownerUserId,omitempty"`
	IsCore      bool      `json:"isCore"`
	CreatedAt   time.Time `json:"createdAt"`
}

// SessionLog represents a completed workout session.
type SessionLog struct {
	ID          string    `json:"id"`
	WorkoutID   string    `json:"workoutId"`
	WorkoutName string    `json:"workoutName"`
	UserID      string    `json:"userId"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
}

// SessionStepLog captures actual timing for a completed step.
type SessionStepLog struct {
	ID               string `json:"id"`
	SessionID        string `json:"sessionId"`
	StepOrder        int    `json:"stepOrder"`
	Type             string `json:"type"`
	Name             string `json:"name"`
	EstimatedSeconds int    `json:"estimatedSeconds"`
	ElapsedMillis    int64  `json:"elapsedMillis"`
}
