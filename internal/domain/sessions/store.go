package sessions

import "context"

// Store defines the persistence methods needed by session orchestration.
type Store interface {
	SessionStepTimings(ctx context.Context, sessionID string) ([]SessionStepLog, error)
	WorkoutWithSteps(ctx context.Context, id string) (*Workout, error)
	RecordSession(ctx context.Context, log SessionLog, steps []SessionStepLog) error
	SessionHistory(ctx context.Context, userID string, limit int) ([]SessionLog, error)
}
