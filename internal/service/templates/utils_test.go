package templates

import (
	"testing"

	domaintemplates "github.com/gi8lino/motus/internal/domain/templates"
	"github.com/gi8lino/motus/internal/service"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	t.Run("MapsNotFound", func(t *testing.T) {
		t.Parallel()
		svc := New(&fakeTemplateStore{})
		err := svc.mapError(&domaintemplates.Error{Kind: domaintemplates.KindNotFound, Message: "nope"})
		if !service.IsKind(err, service.ErrorNotFound) {
			t.Fatalf("expected not found kind")
		}
	})
}
