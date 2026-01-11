package sessions

import "time"

// SessionState captures the runtime status that the SPA consumes for an active session.
type SessionState struct {
	SessionID    string             `json:"sessionId"`
	WorkoutID    string             `json:"workoutId"`
	UserID       string             `json:"userId"`
	WorkoutName  string             `json:"workoutName"`
	CurrentIndex int                `json:"currentIndex"`
	Running      bool               `json:"running"`
	Done         bool               `json:"done"`
	StartedAt    time.Time          `json:"startedAt"`
	CompletedAt  time.Time          `json:"completedAt"`
	Steps        []SessionStepState `json:"steps"`
}

// SessionStepState describes a single card/step inside a session view.
type SessionStepState struct {
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

// Exercise represents a configured exercise inside a session step.
type Exercise struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Reps     string `json:"reps"`
	Weight   string `json:"weight"`
	Duration string `json:"duration"`
	SoundKey string `json:"soundKey,omitempty"`
}

// PauseOptions controls auto-advance for pauses.
type PauseOptions struct {
	AutoAdvance bool `json:"autoAdvance"`
}
