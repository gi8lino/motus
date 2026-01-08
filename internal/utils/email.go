package utils

import (
	"fmt"
	"net/mail"
)

// NormalizeEmail trims, lowercases, and validates an email address.
func NormalizeEmail(value string) (string, error) {
	trimmed := NormalizeToken(value)
	if trimmed == "" {
		return "", fmt.Errorf("email is required")
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed {
		return "", fmt.Errorf("valid email is required")
	}

	return trimmed, nil
}
