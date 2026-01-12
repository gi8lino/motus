package exercises

import (
	"strings"

	"github.com/gi8lino/motus/internal/utils"
)

// requireUserID ensures a non-empty user identifier.
func requireUserID(value string) (string, error) {
	if trimmed := utils.NormalizeToken(value); trimmed != "" {
		return trimmed, nil
	}
	return "", validation("userId is required")
}

// requireName ensures a non-empty exercise name.
func requireName(value string) (string, error) {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed, nil
	}
	return "", validation("exercise name is required")
}

// requireEntityID validates a generic entity identifier.
func requireEntityID(value, msg string) (string, error) {
	if trimmed := utils.NormalizeToken(value); trimmed != "" {
		return trimmed, nil
	}
	return "", validation(msg)
}
