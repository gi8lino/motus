package service

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {
	t.Parallel()

	t.Run("Message", func(t *testing.T) {
		t.Parallel()
		err := &Error{Kind: ErrorValidation, Err: errors.New("boom")}
		if err.Error() != "boom" {
			t.Fatalf("expected error message")
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("root")
		err := &Error{Kind: ErrorInternal, Err: cause}
		if !errors.Is(err, cause) {
			t.Fatalf("expected unwrap to expose cause")
		}
	})
}

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

func TestNewErrorWithScope(t *testing.T) {
	t.Parallel()

	t.Run("SetsScope", func(t *testing.T) {
		t.Parallel()
		err := NewErrorWithScope(ErrorNotFound, "missing", "workouts")
		var svcErr *Error
		if !errors.As(err, &svcErr) {
			t.Fatalf("expected service error")
		}
		if svcErr.Scope != "workouts" {
			t.Fatalf("expected scope to be set")
		}
	})
}

func TestWrapError(t *testing.T) {
	t.Parallel()

	t.Run("WrapsCause", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("boom")
		err := WrapError(ErrorInternal, cause)
		if !errors.Is(err, cause) {
			t.Fatalf("expected wrapped cause")
		}
	})

	t.Run("WrapsWithScope", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("boom")
		err := WrapErrorWithScope(ErrorInternal, "templates", cause)
		var svcErr *Error
		if !errors.As(err, &svcErr) {
			t.Fatalf("expected service error")
		}
		if svcErr.Scope != "templates" {
			t.Fatalf("expected scope to be set")
		}
	})
}
