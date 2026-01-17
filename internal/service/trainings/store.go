package trainings

import "context"

// Store defines the persistence methods needed by training orchestration.
type Store interface {
	TrainingStepTimings(ctx context.Context, trainingID string) ([]TrainingStepLog, error)
	WorkoutWithSteps(ctx context.Context, id string) (*Workout, error)
	RecordTraining(ctx context.Context, log TrainingLog, steps []TrainingStepLog) error
	TrainingHistory(ctx context.Context, userID string, limit int) ([]TrainingLog, error)
}
