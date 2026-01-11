// Package templates provides domain errors for the template domain.
package templates

import "fmt"

// ErrorKind categorizes template domain failures.
type ErrorKind int

const (
	KindValidation ErrorKind = iota
	KindNotFound
	KindInternal
)

// Error represents a template domain rule violation.
type Error struct {
	Kind    ErrorKind
	Message string
	Err     error
}

// Error implements error.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// DomainKind exposes the underlying domain error kind.
func (e *Error) DomainKind() int {
	if e == nil {
		return int(KindInternal)
	}
	return int(e.Kind)
}

// validation creates a validation error.
func validation(message string) error {
	return &Error{Kind: KindValidation, Message: message}
}

// notFound creates a not-found error.
func notFound(message string) error {
	return &Error{Kind: KindNotFound, Message: message}
}

// internal wraps an internal error.
func internal(err error) error {
	if err == nil {
		return &Error{Kind: KindInternal, Message: "internal error"}
	}
	return &Error{Kind: KindInternal, Message: err.Error(), Err: err}
}
