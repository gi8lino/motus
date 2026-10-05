package utils

import "strings"

// NormalizeToken normalizes a token to lowercase and trims it.
func NormalizeToken(token string) string {
	return strings.TrimSpace(strings.ToLower(token))
}
