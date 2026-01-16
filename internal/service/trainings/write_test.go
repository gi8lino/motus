package trainings

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/utils"
)

func TestBuildTrainingLog(t *testing.T) {
	t.Parallel()

	t.Run("Defaults and corrections", func(t *testing.T) {
		t.Parallel()

		started := time.Now().Add(10 * time.Second)
		completed := started.Add(-2 * time.Second)
		log, steps, err := BuildTrainingLog(CompleteRequest{
			TrainingID:  "sess",
			WorkoutID:   "work",
			WorkoutName: "Workout",
			UserID:      "user",
			StartedAt:   started,
			CompletedAt: completed,
			Steps: []TrainingStepState{{
				Name:          "Step 1",
				Type:          string(utils.StepTypeSet),
				ElapsedMillis: 1200,
			}},
		})
		require.NoError(t, err)
		assert.False(t, log.CompletedAt.Before(log.StartedAt))
		assert.Equal(t, time.Second, log.CompletedAt.Sub(log.StartedAt))
		require.Len(t, steps, 1)
		assert.Equal(t, "sess", steps[0].TrainingID)
		assert.Equal(t, 0, steps[0].StepOrder)
	})
}

func TestRecordTraining(t *testing.T) {
	t.Parallel()

	t.Run("PersistsLog", func(t *testing.T) {
		t.Parallel()

		called := false
		store := &fakeStore{
			recordFn: func(context.Context, TrainingLog, []TrainingStepLog) error {
				called = true
				return nil
			},
		}
		_, err := RecordTraining(context.Background(), store, CompleteRequest{
			TrainingID:  "s1",
			WorkoutID:   "w1",
			WorkoutName: "Workout",
			UserID:      "u1",
			StartedAt:   time.Now(),
			CompletedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatalf("expected RecordTraining to be called")
		}
	})
}

func TestRecordTrainingMethod(t *testing.T) {
	t.Parallel()

	t.Run("DelegatesToStore", func(t *testing.T) {
		t.Parallel()

		called := false
		store := &fakeStore{
			recordFn: func(context.Context, TrainingLog, []TrainingStepLog) error {
				called = true
				return nil
			},
		}
		svc := New(store, func(string) string { return "" })
		_, err := svc.RecordTraining(context.Background(), CompleteRequest{
			TrainingID:  "s2",
			WorkoutID:   "w1",
			WorkoutName: "Workout",
			UserID:      "u1",
			StartedAt:   time.Now(),
			CompletedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatalf("expected RecordTraining to be called")
		}
	})
}
