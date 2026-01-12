package utils

import "strings"

// DefaultIfZero returns fallback when value is the zero value.
func DefaultIfZero[T comparable](value, fallback T) T {
	var zero T
	if value == zero {
		return fallback
	}
	return value
}

// NormalizeToken normalizes a token to lowercase and trims it.
func NormalizeToken(token string) string {
	return strings.TrimSpace(strings.ToLower(token))
}
