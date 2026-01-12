package workouts

import (
	"regexp"

	"github.com/gi8lino/motus/internal/db"
)

var repRangePattern = regexp.MustCompile(`^\d+(-\d+)?$`)

// WorkoutRequest describes the payload for building a workout definition.
type WorkoutRequest struct {
	UserID string      `json:"userId"`
	Name   string      `json:"name"`
	Steps  []StepInput `json:"steps"`
}

// StepInput describes a workout step definition in the domain model.
type StepInput struct {
	Type                  string          `json:"type"`
	Name                  string          `json:"name"`
	Duration              string          `json:"duration"`
	EstimatedSeconds      int             `json:"estimatedSeconds"`
	SoundKey              string          `json:"soundKey"`
	Subsets               []SubsetInput   `json:"subsets"`
	PauseOptions          db.PauseOptions `json:"pauseOptions"`
	RepeatCount           int             `json:"repeatCount"`
	RepeatRestSeconds     int             `json:"repeatRestSeconds"`
	RepeatRestAfterLast   bool            `json:"repeatRestAfterLast"`
	RepeatRestSoundKey    string          `json:"repeatRestSoundKey"`
	RepeatRestAutoAdvance bool            `json:"repeatRestAutoAdvance"`
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
