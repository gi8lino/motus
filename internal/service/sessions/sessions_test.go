package sessions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

func TestSessionStateFromWorkout(t *testing.T) {
	t.Parallel()

	t.Run("Auto-advance pause detected", func(t *testing.T) {
		t.Parallel()

		workout := &db.Workout{
			ID:     "w1",
			UserID: "u1",
			Name:   "Test",
			Steps: []db.WorkoutStep{
				{
					ID:               "s1",
					Type:             string(utils.StepTypePause),
					Name:             "Pause",
					EstimatedSeconds: 10,
					SoundKey:         "beep",
					PauseOptions:     db.PauseOptions{AutoAdvance: true},
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

func TestBuildSessionHistoryItems(t *testing.T) {
	t.Parallel()

	t.Run("Maps logs into response", func(t *testing.T) {
		t.Parallel()

		started := time.Now().Add(-2 * time.Hour)
		completed := time.Now().Add(-time.Hour)
		logs := []db.SessionLog{{
			ID:          "s1",
			WorkoutID:   "w1",
			WorkoutName: "Workout",
			UserID:      "u1",
			StartedAt:   started,
			CompletedAt: completed,
		}}
		stepMap := map[string][]db.SessionStepLog{
			"s1": {{ID: "s1-0", SessionID: "s1", StepOrder: 0}},
		}

		items := BuildSessionHistoryItems(logs, stepMap)
		require.Len(t, items, 1)
		assert.Equal(t, "s1", items[0].SessionID)
		assert.Equal(t, "w1", items[0].WorkoutID)
		require.Len(t, items[0].Steps, 1)
	})
}
