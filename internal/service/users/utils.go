package users

import (
	errpkg "github.com/gi8lino/motus/internal/service/errors"
	"github.com/gi8lino/motus/internal/utils"
)

// requireEntityID normalizes and validates a user/entity identifier.
func requireEntityID(value, message string) (string, error) {
	if trimmed := utils.NormalizeToken(value); trimmed != "" {
		return trimmed, nil
	}
	return "", errpkg.NewErrorWithScope(errpkg.ErrorValidation, message, errorScope)
}
