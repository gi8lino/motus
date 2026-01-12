package sessions

import (
	"time"

	"github.com/gi8lino/motus/internal/db"
	domainSessions "github.com/gi8lino/motus/internal/domain/sessions"
)

// SessionHistoryItem is the API payload for a completed session.
type SessionHistoryItem struct {
	ID          string              `json:"id"`                    // ID is the history item identifier.
	SessionID   string              `json:"sessionId"`             // SessionID links to the completed session.
	WorkoutID   string              `json:"workoutId"`             // WorkoutID references the workout definition.
	WorkoutName string              `json:"workoutName"`           // WorkoutName is the display name at completion time.
	UserID      string              `json:"userId"`                // UserID owns the session.
	StartedAt   *time.Time          `json:"startedAt,omitempty"`   // StartedAt is when the session began.
	CompletedAt *time.Time          `json:"completedAt,omitempty"` // CompletedAt is when the session finished.
	Steps       []db.SessionStepLog `json:"steps,omitempty"`       // Steps contains logged timings when available.
}

// Store exposes the persistence requirements for session services.
type Store = domainSessions.Store

type (
	SessionState     = domainSessions.SessionState
	SessionStepState = domainSessions.SessionStepState
	Exercise         = domainSessions.Exercise
	PauseOptions     = domainSessions.PauseOptions
)
