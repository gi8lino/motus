package sessions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/utils"
)

func TestBuildSessionLog(t *testing.T) {
	t.Parallel()

	t.Run("Defaults and corrections", func(t *testing.T) {
		t.Parallel()

		started := time.Now().Add(10 * time.Second)
		completed := started.Add(-2 * time.Second)
		log, steps, err := BuildSessionLog(CompleteRequest{
			SessionID:   "sess",
			WorkoutID:   "work",
			WorkoutName: "Workout",
			UserID:      "user",
			StartedAt:   started,
			CompletedAt: completed,
			Steps: []SessionStepState{{
				Name:          "Step 1",
				Type:          string(utils.StepTypeSet),
				ElapsedMillis: 1200,
			}},
		})
		require.NoError(t, err)
		assert.False(t, log.CompletedAt.Before(log.StartedAt))
		assert.Equal(t, time.Second, log.CompletedAt.Sub(log.StartedAt))
		require.Len(t, steps, 1)
		assert.Equal(t, "sess", steps[0].SessionID)
		assert.Equal(t, 0, steps[0].StepOrder)
	})
}

func TestRecordSession(t *testing.T) {
	t.Parallel()

	t.Run("PersistsLog", func(t *testing.T) {
		t.Parallel()

		called := false
		store := &fakeStore{
			recordFn: func(context.Context, SessionLog, []SessionStepLog) error {
				called = true
				return nil
			},
		}
		_, err := RecordSession(context.Background(), store, CompleteRequest{
			SessionID:   "s1",
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
			t.Fatalf("expected RecordSession to be called")
		}
	})
}

func TestRecordSessionMethod(t *testing.T) {
	t.Parallel()

	t.Run("DelegatesToStore", func(t *testing.T) {
		t.Parallel()

		called := false
		store := &fakeStore{
			recordFn: func(context.Context, SessionLog, []SessionStepLog) error {
				called = true
				return nil
			},
		}
		svc := New(store, func(string) string { return "" })
		_, err := svc.RecordSession(context.Background(), CompleteRequest{
			SessionID:   "s2",
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
			t.Fatalf("expected RecordSession to be called")
		}
	})
}
