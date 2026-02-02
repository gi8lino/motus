package utils

import (
	"errors"
	"net/mail"
)

// NormalizeEmail trims, lowercases, and validates an email address.
func NormalizeEmail(value string) (string, error) {
	trimmed := NormalizeToken(value)
	if trimmed == "" {
		return "", errors.New("email is required")
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed {
		return "", errors.New("valid email is required")
	}

	return trimmed, nil
}
