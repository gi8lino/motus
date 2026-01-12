package exercises

import "fmt"

// ErrorKind classifies domain rule violations.
type ErrorKind int

const (
	KindValidation ErrorKind = iota
	KindForbidden
	KindNotFound
	KindInternal
)

// Error represents a domain-level failure in the exercises domain.
type Error struct {
	Kind    ErrorKind
	Message string
	Err     error
}

// Error implements the error interface.
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

// forbidden creates a forbidden error.
func forbidden(message string) error {
	return &Error{Kind: KindForbidden, Message: message}
}

// internal wraps an internal error.
func internal(err error) error {
	if err == nil {
		return &Error{Kind: KindInternal, Message: "internal error"}
	}
	return &Error{Kind: KindInternal, Message: err.Error(), Err: err}
}
