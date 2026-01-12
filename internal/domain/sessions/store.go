package sessions

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Store defines the persistence methods needed by session orchestration.
type Store interface {
	// SessionStepTimings loads logged steps for a session.
	SessionStepTimings(ctx context.Context, sessionID string) ([]db.SessionStepLog, error)
	// WorkoutWithSteps loads a workout and its steps.
	WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error)
	// RecordSession stores a completed session and its steps.
	RecordSession(ctx context.Context, log db.SessionLog, steps []db.SessionStepLog) error
	// SessionHistory returns completed session logs for a user.
	SessionHistory(ctx context.Context, userID string, limit int) ([]db.SessionLog, error)
}
