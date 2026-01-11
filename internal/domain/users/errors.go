// Package users defines domain errors for the user domain.
package users

import "fmt"

// ErrorKind enumerates user domain error categories.
type ErrorKind int

const (
	KindValidation ErrorKind = iota
	KindForbidden
	KindUnauthorized
	KindInternal
)

// Error reports rule violations happening in the domain layer.
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

// forbidden creates a forbidden error.
func forbidden(message string) error {
	return &Error{Kind: KindForbidden, Message: message}
}

// unauthorized creates an unauthorized error.
func unauthorized(message string) error {
	return &Error{Kind: KindUnauthorized, Message: message}
}

// internal wraps an internal error.
func internal(err error) error {
	if err == nil {
		return &Error{Kind: KindInternal, Message: "internal error"}
	}
	return &Error{Kind: KindInternal, Message: err.Error(), Err: err}
}
