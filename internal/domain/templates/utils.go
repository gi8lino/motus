package templates

import "github.com/gi8lino/motus/internal/utils"

// requireID normalizes and validates a template identifier.
func requireID(value, message string) (string, error) {
	if trimmed := utils.NormalizeToken(value); trimmed != "" {
		return trimmed, nil
	}
	return "", validation(message)
}
