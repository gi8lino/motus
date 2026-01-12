package sessions

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

func TestNewStateFromWorkout(t *testing.T) {
	t.Parallel()

	t.Run("Defaults and corrections", func(t *testing.T) {
		t.Parallel()

		workout := &db.Workout{
			ID:     "w1",
			UserID: "u1",
			Name:   "Workout",
			Steps: []db.WorkoutStep{
				{
					ID:   "s1",
					Type: utils.StepTypePause.String(),
					Name: "Break",
				},
				{
					ID:   "s2",
					Type: utils.StepTypeSet.String(),
					Name: "Set",
					Subsets: []db.WorkoutSubset{
						{
							Name:      "Superset",
							Superset:  true,
							Exercises: []db.SubsetExercise{{Name: "Push", Type: utils.ExerciseTypeRep}},
						},
						{
							Name: "Normal",
							Exercises: []db.SubsetExercise{
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
}
