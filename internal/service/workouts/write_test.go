package workouts

import (
	"context"
	"testing"
)

func TestCreate(t *testing.T) {
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
		workout, err := svc.Create(context.Background(), WorkoutRequest{UserID: "u1", Name: "Workout", Steps: []StepInput{{Type: "set", Name: "A", Subsets: []SubsetInput{{Exercises: []ExerciseInput{{Name: "X"}}}}}}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called || workout == nil {
			t.Fatalf("expected create to run")
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		called := false
		svc := New(&fakeStore{
			updateFn: func(context.Context, *Workout) (*Workout, error) {
				called = true
				return &Workout{ID: "w1"}, nil
			},
		})
		workout, err := svc.Update(context.Background(), "w1", WorkoutRequest{Name: "Workout", Steps: []StepInput{{Type: "set", Name: "A", Subsets: []SubsetInput{{Exercises: []ExerciseInput{{Name: "X"}}}}}}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called || workout == nil {
			t.Fatalf("expected update to run")
		}
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		called := false
		svc := New(&fakeStore{
			deleteFn: func(context.Context, string) error {
				called = true
				return nil
			},
		})
		if err := svc.Delete(context.Background(), "w1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatalf("expected delete to run")
		}
	})
}
