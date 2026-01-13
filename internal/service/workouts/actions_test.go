package workouts

import (
	"context"
	"testing"
)

func TestImport(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		called := false
		svc := New(&fakeStore{
			createFn: func(context.Context, *Workout) (*Workout, error) {
				called = true
				return &Workout{ID: "w1"}, nil
			},
		})
		workout, err := svc.Import(context.Background(), "u1", Workout{
			Name: "Workout",
			Steps: []WorkoutStep{
				{
					Type: "set",
					Name: "A",
					Subsets: []WorkoutSubset{{
						Exercises: []SubsetExercise{{Name: "X"}},
					}},
				},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called || workout == nil {
			t.Fatalf("expected import to run")
		}
	})
}
