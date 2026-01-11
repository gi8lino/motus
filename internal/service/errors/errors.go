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
)

// Error captures a service error with a domain-level kind.
type Error struct {
	Kind ErrorKind
	Err  error
}

// Error returns the wrapped error message.
func (e *Error) Error() string {
	return e.Err.Error()
}

// IsKind reports whether the error matches the requested kind.
func IsKind(err error, kind ErrorKind) bool {
	var svcErr *Error
	return errors.As(err, &svcErr) && svcErr.Kind == kind
}

// NewError wraps a message with a service error kind.
func NewError(kind ErrorKind, message string) error {
	return &Error{Kind: kind, Err: errors.New(message)}
}

type domainError interface {
	error
	DomainKind() int
}

// MapDomainError transforms a domain error into a service error using the provided mapper.
func MapDomainError(err error, mapKind func(int) (ErrorKind, bool)) error {
	if err == nil {
		return nil
	}
	var de domainError
	if errors.As(err, &de) {
		if kind, ok := mapKind(de.DomainKind()); ok {
			return NewError(kind, err.Error())
		}
	}
	return NewError(ErrorInternal, err.Error())
}
