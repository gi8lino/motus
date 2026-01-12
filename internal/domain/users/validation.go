package users

import "github.com/gi8lino/motus/internal/utils"

// requireEntityID normalizes and validates a user/entity identifier.
func requireEntityID(value, message string) (string, error) {
	if trimmed := utils.NormalizeToken(value); trimmed != "" {
		return trimmed, nil
	}
	return "", validation(message)
}
