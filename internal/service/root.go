package service

import errorspkg "github.com/gi8lino/motus/internal/service/errors"

type ErrorKind = errorspkg.ErrorKind
type Error = errorspkg.Error

const (
	ErrorValidation   = errorspkg.ErrorValidation
	ErrorForbidden    = errorspkg.ErrorForbidden
	ErrorNotFound     = errorspkg.ErrorNotFound
	ErrorUnauthorized = errorspkg.ErrorUnauthorized
	ErrorInternal     = errorspkg.ErrorInternal
)

func IsKind(err error, kind ErrorKind) bool {
	return errorspkg.IsKind(err, kind)
}

func NewError(kind ErrorKind, message string) error {
	return errorspkg.NewError(kind, message)
}

func MapDomainError(err error, mapKind func(int) (ErrorKind, bool)) error {
	return errorspkg.MapDomainError(err, mapKind)
}
