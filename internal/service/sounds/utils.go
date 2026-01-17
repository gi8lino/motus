package sounds

import "github.com/gi8lino/motus/internal/utils"

// lookup is a map of sound keys to options, built once to avoid recomputing per call.
var lookup = func() map[string]Option {
	m := make(map[string]Option, len(BuiltinOptions))
	for _, opt := range BuiltinOptions {
		m[utils.NormalizeToken(opt.Key)] = opt
	}
	return m
}()

// URLByKey resolves a sound key to its file path.
func URLByKey(key string) string {
	if k := utils.NormalizeToken(key); k != "" {
		if opt, ok := lookup[k]; ok {
			return opt.File
		}
	}
	return ""
}

// ValidKey reports whether a sound key is recognized.
func ValidKey(key string) bool {
	if key == "" {
		return true
	}
	_, ok := lookup[utils.NormalizeToken(key)]
	return ok
}
