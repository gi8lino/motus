package exercises

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Parallel()

	t.Run("Nil", func(t *testing.T) {
		t.Parallel()
		var err *Error
		assert.Equal(t, "", err.Error())
	})

	t.Run("Message", func(t *testing.T) {
		t.Parallel()
		err := &Error{Message: "boom"}
		assert.Equal(t, "boom", err.Error())
	})
}

func TestUnwrap(t *testing.T) {
	t.Parallel()

	t.Run("Nil", func(t *testing.T) {
		t.Parallel()
		var err *Error
		assert.Nil(t, err.Unwrap())
	})

	t.Run("Wrapped", func(t *testing.T) {
		t.Parallel()
		wrapped := assert.AnError
		err := &Error{Message: "boom", Err: wrapped}
		assert.Equal(t, wrapped, err.Unwrap())
	})
}

func TestDomainKind(t *testing.T) {
	t.Parallel()

	t.Run("Nil", func(t *testing.T) {
		t.Parallel()
		var err *Error
		assert.Equal(t, int(KindInternal), err.DomainKind())
	})

	t.Run("Value", func(t *testing.T) {
		t.Parallel()
		err := &Error{Kind: KindForbidden}
		assert.Equal(t, int(KindForbidden), err.DomainKind())
	})
}

func TestValidation(t *testing.T) {
	t.Parallel()

	t.Run("CreatesError", func(t *testing.T) {
		t.Parallel()
		err := validation("bad")
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindValidation, domainErr.Kind)
	})
}

func TestNotFound(t *testing.T) {
	t.Parallel()

	t.Run("CreatesError", func(t *testing.T) {
		t.Parallel()
		err := notFound("missing")
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindNotFound, domainErr.Kind)
	})
}

func TestForbidden(t *testing.T) {
	t.Parallel()

	t.Run("CreatesError", func(t *testing.T) {
		t.Parallel()
		err := forbidden("nope")
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindForbidden, domainErr.Kind)
	})
}

func TestInternal(t *testing.T) {
	t.Parallel()

	t.Run("Nil", func(t *testing.T) {
		t.Parallel()
		err := internal(nil)
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindInternal, domainErr.Kind)
	})

	t.Run("Wrapped", func(t *testing.T) {
		t.Parallel()
		err := internal(assert.AnError)
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindInternal, domainErr.Kind)
	})
}
