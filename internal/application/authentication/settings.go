package authentication

import (
	"strings"

	"github.com/gi8lino/motus/internal/db"
)

// Settings is the persisted, non-secret browser authentication configuration.
type Settings = db.AuthenticationSettings

// HeaderList normalizes a comma-separated persisted trusted-header list.
func HeaderList(value string) []string {
	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
