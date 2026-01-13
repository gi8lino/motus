package sessions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/utils"
)

func TestSessionStateFromWorkout(t *testing.T) {
	t.Parallel()

	t.Run("Auto-advance pause detected", func(t *testing.T) {
		t.Parallel()

		workout := &Workout{
			ID:     "w1",
			UserID: "u1",
			Name:   "Test",
			Steps: []WorkoutStep{
				{
					ID:               "s1",
					Type:             string(utils.StepTypePause),
					Name:             "Pause",
					EstimatedSeconds: 10,
					SoundKey:         "beep",
					PauseOptions:     PauseOptions{AutoAdvance: true},
				},
			},
		}

		state := SessionStateFromWorkout(workout, func(key string) string { return "/" + key })
		require.Len(t, state.Steps, 1)
		step := state.Steps[0]
		assert.True(t, step.PauseOptions.AutoAdvance)
		assert.Equal(t, "/beep", step.SoundURL)
	})
}

func TestCreateState(t *testing.T) {
	t.Parallel()

	t.Run("Validation", func(t *testing.T) {
		t.Parallel()

		_, err := CreateState(context.Background(), &fakeStore{}, " ", func(string) string { return "" })
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{
			workoutFn: func(context.Context, string) (*Workout, error) {
				return &Workout{ID: "w1", UserID: "u1", Name: "Workout"}, nil
			},
		}
		state, err := CreateState(context.Background(), store, "w1", func(string) string { return "" })
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.WorkoutID != "w1" {
			t.Fatalf("unexpected workout id")
		}
	})
}

func TestCreateStateMethod(t *testing.T) {
	t.Parallel()

	t.Run("UsesServiceStore", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{
			workoutFn: func(context.Context, string) (*Workout, error) {
				return &Workout{ID: "w2", UserID: "u1", Name: "Workout"}, nil
			},
		}
		svc := New(store, func(string) string { return "" })
		state, err := svc.CreateState(context.Background(), "w2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.WorkoutID != "w2" {
			t.Fatalf("unexpected workout id")
		}
	})
}

func TestFetchStepTimings(t *testing.T) {
	t.Parallel()

	t.Run("Validation", func(t *testing.T) {
		t.Parallel()

		_, err := FetchStepTimings(context.Background(), &fakeStore{}, " ")
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestFetchStepTimingsMethod(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{
			stepTimingsFn: func(context.Context, string) ([]SessionStepLog, error) {
				return []SessionStepLog{{ID: "step"}}, nil
			},
		}
		svc := New(store, func(string) string { return "" })
		steps, err := svc.FetchStepTimings(context.Background(), "sess")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(steps) != 1 {
			t.Fatalf("expected steps")
		}
	})
}

func TestBuildSessionHistory(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{
			stepTimingsFn: func(context.Context, string) ([]SessionStepLog, error) {
				return []SessionStepLog{{ID: "s1-0", SessionID: "s1", StepOrder: 0}}, nil
			},
		}
		history := []SessionLog{{ID: "s1", WorkoutID: "w1"}}
		items, err := BuildSessionHistory(context.Background(), store, history)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected history items")
		}
	})
}

func TestBuildSessionHistoryMethod(t *testing.T) {
	t.Parallel()

	t.Run("Validation", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{}, func(string) string { return "" })
		_, err := svc.BuildSessionHistory(context.Background(), " ", 10)
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestBuildSessionHistoryItems(t *testing.T) {
	t.Parallel()

	t.Run("Maps logs into response", func(t *testing.T) {
		t.Parallel()

		started := time.Now().Add(-2 * time.Hour)
		completed := time.Now().Add(-time.Hour)
		logs := []SessionLog{{
			ID:          "s1",
			WorkoutID:   "w1",
			WorkoutName: "Workout",
			UserID:      "u1",
			StartedAt:   started,
			CompletedAt: completed,
		}}
		stepMap := map[string][]SessionStepLog{
			"s1": {{ID: "s1-0", SessionID: "s1", StepOrder: 0}},
		}

		items := BuildSessionHistoryItems(logs, stepMap)
		require.Len(t, items, 1)
		assert.Equal(t, "s1", items[0].SessionID)
		assert.Equal(t, "w1", items[0].WorkoutID)
		require.Len(t, items[0].Steps, 1)
	})
}
