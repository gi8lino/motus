package exercises

import (
	"testing"

	domainexercises "github.com/gi8lino/motus/internal/domain/exercises"
	"github.com/gi8lino/motus/internal/service"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	t.Run("MapsValidation", func(t *testing.T) {
		t.Parallel()
		svc := New(&fakeStore{})
		err := svc.mapError(&domainexercises.Error{Kind: domainexercises.KindValidation, Message: "bad"})
		if !service.IsKind(err, service.ErrorValidation) {
			t.Fatalf("expected validation kind")
		}
	})
}
