package trainings

import (
	"testing"

	"github.com/gi8lino/motus/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewStateFromWorkout(t *testing.T) {
	t.Parallel()

	t.Run("Defaults and corrections", func(t *testing.T) {
		t.Parallel()

		workout := &Workout{
			ID:     "w1",
			UserID: "u1",
			Name:   "Workout",
			Steps: []WorkoutStep{
				{
					ID:   "s1",
					Type: utils.StepTypePause.String(),
					Name: "Break",
				},
				{
					ID:   "s2",
					Type: utils.StepTypeSet.String(),
					Name: "Set",
					Subsets: []WorkoutSubset{
						{
							Name:      "Superset",
							Superset:  true,
							Exercises: []SubsetExercise{{Name: "Push", Type: utils.ExerciseTypeRep}},
						},
						{
							Name: "Normal",
							Exercises: []SubsetExercise{
								{Name: "Pull", Type: utils.ExerciseTypeStopwatch},
							},
						},
					},
				},
			},
		}

		state := NewStateFromWorkout(workout, func(key string) string { return "/sounds/" + key })
		assert.Len(t, state.Steps, 3)
		assert.Equal(t, "Break", state.Steps[0].Name)
		assert.True(t, state.Steps[1].Superset)
		assert.Equal(t, "Pull", state.Steps[2].Exercises[0].Name)
	})

	t.Run("Timed exercises marked as a superset become sequential steps", func(t *testing.T) {
		t.Parallel()

		workout := &Workout{
			ID:     "w1",
			UserID: "u1",
			Name:   "Workout",
			Steps: []WorkoutStep{
				{
					ID:   "s1",
					Type: utils.StepTypeSet.String(),
					Name: "Timed sequence",
					Subsets: []WorkoutSubset{
						{
							Name:     "Work",
							Superset: true,
							Exercises: []SubsetExercise{
								{Name: "One", Type: utils.ExerciseTypeCountdown, Duration: "20s"},
								{Name: "Two", Type: utils.ExerciseTypeCountdown, Duration: "30s"},
							},
						},
					},
				},
			},
		}

		state := NewStateFromWorkout(workout, func(string) string { return "" })
		assert.Len(t, state.Steps, 2)
		assert.False(t, state.Steps[0].Superset)
		assert.False(t, state.Steps[1].Superset)
		assert.True(t, state.Steps[0].AutoAdvance)
		assert.True(t, state.Steps[1].AutoAdvance)
	})

}
