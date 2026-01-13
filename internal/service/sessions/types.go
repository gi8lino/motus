// Package sessions provides service access to workout sessions.
package sessions

import (
	"time"

	domainSessions "github.com/gi8lino/motus/internal/domain/sessions"
)

// SessionHistoryItem is the API payload for a completed session.
type SessionHistoryItem struct {
	ID          string                          `json:"id"`                    // ID is the history item identifier.
	SessionID   string                          `json:"sessionId"`             // SessionID links to the completed session.
	WorkoutID   string                          `json:"workoutId"`             // WorkoutID references the workout definition.
	WorkoutName string                          `json:"workoutName"`           // WorkoutName is the display name at completion time.
	UserID      string                          `json:"userId"`                // UserID owns the session.
	StartedAt   *time.Time                      `json:"startedAt,omitempty"`   // StartedAt is when the session began.
	CompletedAt *time.Time                      `json:"completedAt,omitempty"` // CompletedAt is when the session finished.
	Steps       []domainSessions.SessionStepLog `json:"steps,omitempty"`       // Steps contains logged timings when available.
}

// Store exposes the persistence requirements for session services.
type Store = domainSessions.Store

type (
	SessionState     = domainSessions.SessionState
	SessionStepState = domainSessions.SessionStepState
	Exercise         = domainSessions.Exercise
	PauseOptions     = domainSessions.PauseOptions
	SessionLog       = domainSessions.SessionLog
	SessionStepLog   = domainSessions.SessionStepLog
	Workout          = domainSessions.Workout
	WorkoutStep      = domainSessions.WorkoutStep
)

// Service coordinates session operations.
type Service struct {
	store         Store
	soundURLByKey func(string) string
}

// New creates a new sessions service.
func New(store Store, soundURLByKey func(string) string) *Service {
	return &Service{store: store, soundURLByKey: soundURLByKey}
}

// CompleteRequest captures the payload for logging a finished session.
type CompleteRequest struct {
	SessionID   string             `json:"sessionId"`   // SessionID identifies the session.
	WorkoutID   string             `json:"workoutId"`   // WorkoutID identifies the workout.
	WorkoutName string             `json:"workoutName"` // WorkoutName is the display name at completion time.
	UserID      string             `json:"userId"`      // UserID owns the session.
	StartedAt   time.Time          `json:"startedAt"`   // StartedAt records when the session began.
	CompletedAt time.Time          `json:"completedAt"` // CompletedAt records when the session finished.
	Steps       []SessionStepState `json:"steps"`       // Steps includes timing details.
}
