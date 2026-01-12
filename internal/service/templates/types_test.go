package templates

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("CreatesService", func(t *testing.T) {
		t.Parallel()
		svc := New(&fakeTemplateStore{})
		if svc == nil {
			t.Fatalf("expected service")
		}
	})
}
