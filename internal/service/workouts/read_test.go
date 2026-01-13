package workouts

import (
	"context"
	"testing"
)

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		svc := New(&fakeStore{
			getFn: func(context.Context, string) (*Workout, error) {
				return &Workout{ID: "w1"}, nil
			},
		})
		workout, err := svc.Get(context.Background(), "w1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if workout == nil {
			t.Fatalf("expected workout")
		}
	})
}

func TestExport(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		svc := New(&fakeStore{
			getFn: func(context.Context, string) (*Workout, error) {
				return &Workout{ID: "w1"}, nil
			},
		})
		workout, err := svc.Export(context.Background(), "w1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if workout == nil {
			t.Fatalf("expected workout")
		}
	})
}

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		svc := New(&fakeStore{
			listFn: func(context.Context, string) ([]Workout, error) {
				return []Workout{{ID: "w1"}}, nil
			},
		})
		workouts, err := svc.List(context.Background(), "u1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(workouts) != 1 {
			t.Fatalf("expected workouts")
		}
	})
}
