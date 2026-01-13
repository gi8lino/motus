package sounds

import domain "github.com/gi8lino/motus/internal/domain/sounds"

// URLByKey resolves a sound key to its file path through the domain helper.
func URLByKey(key string) string {
	return domain.URLByKey(key)
}

// ValidKey reports whether a given key is a recognized sound.
func ValidKey(key string) bool {
	return domain.ValidKey(key)
}
