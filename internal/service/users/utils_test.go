package users

import (
	"testing"

	domainusers "github.com/gi8lino/motus/internal/domain/users"
	"github.com/gi8lino/motus/internal/service"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	t.Run("MapsUnauthorized", func(t *testing.T) {
		t.Parallel()
		svc := New(&fakeStore{}, "", false)
		err := svc.mapError(&domainusers.Error{Kind: domainusers.KindUnauthorized, Message: "nope"})
		if !service.IsKind(err, service.ErrorUnauthorized) {
			t.Fatalf("expected unauthorized kind")
		}
	})
}
