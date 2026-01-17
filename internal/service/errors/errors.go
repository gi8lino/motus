// Package errors provides the shared service error model.
package service

import "errors"

// ErrorKind describes the category of a service error.
type ErrorKind string

const (
	ErrorValidation   ErrorKind = "validation"
	ErrorForbidden    ErrorKind = "forbidden"
	ErrorNotFound     ErrorKind = "not_found"
	ErrorUnauthorized ErrorKind = "unauthorized"
	ErrorInternal     ErrorKind = "internal"
	ErrorNoRows       ErrorKind = "no_rows"
)

// Error captures a service error with a domain-level kind.
type Error struct {
	Kind  ErrorKind
	Scope string
	Err   error
}

// Error returns the wrapped error message.
func (e *Error) Error() string {
	return e.Err.Error()
}

// Unwrap returns the wrapped cause.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// IsKind reports whether the error matches the requested kind.
func IsKind(err error, kind ErrorKind) bool {
	var svcErr *Error
	return errors.As(err, &svcErr) && svcErr.Kind == kind
}

// NewError wraps a message with a service error kind.
func NewError(kind ErrorKind, message string) error {
	return NewErrorWithScope(kind, message, "")
}

// NewErrorWithScope wraps a message with a service error kind and scope.
func NewErrorWithScope(kind ErrorKind, message, scope string) error {
	return &Error{Kind: kind, Scope: scope, Err: errors.New(message)}
}

// WrapError wraps an error with a service error kind.
func WrapError(kind ErrorKind, err error) error {
	return WrapErrorWithScope(kind, "", err)
}

// WrapErrorWithScope wraps an error with a service error kind and scope.
func WrapErrorWithScope(kind ErrorKind, scope string, err error) error {
	if err == nil {
		return NewErrorWithScope(kind, "internal error", scope)
	}
	return &Error{Kind: kind, Scope: scope, Err: err}
}
