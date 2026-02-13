package workouts

import (
	"context"
	"testing"

	"github.com/gi8lino/motus/internal/db"
	errpkg "github.com/gi8lino/motus/internal/service/errors"
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

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		svc := New(&fakeStore{
			updateFn: func(context.Context, *Workout) (*Workout, error) {
				return nil, db.ErrWorkoutNotFound
			},
		})
		workout, err := svc.Update(context.Background(), "w1", WorkoutRequest{Name: "Workout", Steps: []StepInput{{Type: "set", Name: "A", Subsets: []SubsetInput{{Exercises: []ExerciseInput{{Name: "X"}}}}}}})
		if workout != nil {
			t.Fatalf("expected nil workout")
		}
		if err == nil {
			t.Fatalf("expected error")
		}
		if !errpkg.IsKind(err, errpkg.ErrorNotFound) {
			t.Fatalf("expected not_found error, got: %v", err)
		}
		if err.Error() != db.ErrWorkoutNotFound.Error() {
			t.Fatalf("expected workout not found message, got: %v", err)
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
