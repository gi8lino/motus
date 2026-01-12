package workouts

import (
	"testing"

	domainworkouts "github.com/gi8lino/motus/internal/domain/workouts"
	"github.com/gi8lino/motus/internal/service"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	t.Run("MapsValidation", func(t *testing.T) {
		t.Parallel()
		svc := New(&fakeStore{})
		err := svc.mapError(&domainworkouts.Error{Kind: domainworkouts.KindValidation, Message: "bad"})
		if !service.IsKind(err, service.ErrorValidation) {
			t.Fatalf("expected validation kind")
		}
	})
}
