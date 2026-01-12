package workouts

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("CreatesService", func(t *testing.T) {
		t.Parallel()
		svc := New(&fakeStore{})
		if svc == nil {
			t.Fatalf("expected service")
		}
	})
}
