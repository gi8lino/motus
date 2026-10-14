package users

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("CreatesService", func(t *testing.T) {
		t.Parallel()
		svc := New(&fakeStore{}, func() bool { return false })
		if svc == nil {
			t.Fatalf("expected service")
		}
	})
}
