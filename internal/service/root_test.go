package service

import (
	"errors"
	"testing"
)

type fakeDomainErr struct {
	kind int
}

func (e fakeDomainErr) Error() string   { return "domain boom" }
func (e fakeDomainErr) DomainKind() int { return e.kind }

func TestIsKind(t *testing.T) {
	t.Parallel()

	t.Run("Matches", func(t *testing.T) {
		t.Parallel()
		err := NewError(ErrorValidation, "bad")
		if !IsKind(err, ErrorValidation) {
			t.Fatalf("expected validation kind")
		}
	})
}

func TestNewError(t *testing.T) {
	t.Parallel()

	t.Run("CreatesError", func(t *testing.T) {
		t.Parallel()
		err := NewError(ErrorForbidden, "nope")
		if !IsKind(err, ErrorForbidden) {
			t.Fatalf("expected forbidden kind")
		}
	})
}

func TestMapDomainError(t *testing.T) {
	t.Parallel()

	t.Run("MapsKnownKind", func(t *testing.T) {
		t.Parallel()
		err := MapDomainError(fakeDomainErr{kind: 1}, func(kind int) (ErrorKind, bool) {
			if kind == 1 {
				return ErrorValidation, true
			}
			return ErrorInternal, false
		})
		if !IsKind(err, ErrorValidation) {
			t.Fatalf("expected validation kind")
		}
	})

	t.Run("FallbackInternal", func(t *testing.T) {
		t.Parallel()
		err := MapDomainError(errors.New("boom"), func(int) (ErrorKind, bool) {
			return ErrorValidation, true
		})
		if !IsKind(err, ErrorInternal) {
			t.Fatalf("expected internal kind")
		}
	})
}
