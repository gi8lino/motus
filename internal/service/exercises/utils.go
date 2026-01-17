package exercises

import (
	"strings"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
	"github.com/gi8lino/motus/internal/utils"
)

// requireUserID ensures a non-empty user identifier.
func requireUserID(value string) (string, error) {
	if trimmed := utils.NormalizeToken(value); trimmed != "" {
		return trimmed, nil
	}
	return "", errpkg.NewErrorWithScope(errpkg.ErrorValidation, "user id is required", errorScope)
}

// requireName ensures a non-empty exercise name.
func requireName(value string) (string, error) {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed, nil
	}
	return "", errpkg.NewErrorWithScope(errpkg.ErrorValidation, "name is required", errorScope)
}

// requireEntityID validates a generic entity identifier.
func requireEntityID(value, msg string) (string, error) {
	if trimmed := utils.NormalizeToken(value); trimmed != "" {
		return trimmed, nil
	}
	return "", errpkg.NewErrorWithScope(errpkg.ErrorValidation, msg, errorScope)
}
