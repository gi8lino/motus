package sounds

// Option describes an available sound option.
type Option struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	File        string `json:"file"`
	LeadSeconds int    `json:"leadSeconds,omitempty"`
}

// BuiltinOptions defines the bundled sound choices.
var BuiltinOptions = []Option{
	{Key: "beep", Label: "Soft Beep", File: "/sounds/beep.wav"},
	{Key: "chime", Label: "Gentle Chime", File: "/sounds/chime.wav"},
	{Key: "click", Label: "Click", File: "/sounds/click.wav"},
	{Key: "soft1", Label: "Friendly Soft 1", File: "/sounds/soft1.wav"},
	{Key: "soft2", Label: "Friendly Soft 2", File: "/sounds/soft2.wav"},
	{Key: "soft3", Label: "Friendly Soft 3", File: "/sounds/soft3.wav"},
	{Key: "soft4", Label: "Friendly Soft 4", File: "/sounds/soft4.wav"},
	{Key: "countdown", Label: "Countdown Tut-Tut-Tuuu", File: "/sounds/countdown.wav"},
	{Key: "race", Label: "Race Start", File: "/sounds/race.wav"},
	{Key: "count321", Label: "3-2-1 Action", File: "/sounds/count321.wav", LeadSeconds: 3},
}

// lookup is a map of sound keys to options.
var lookup = func() map[string]Option {
	m := make(map[string]Option, len(BuiltinOptions))
	for _, opt := range BuiltinOptions {
		m[opt.Key] = opt
	}
	return m
}()

// soundURLByKey resolves a sound key to its file path.
func URLByKey(key string) string {
	if opt, ok := lookup[key]; ok {
		return opt.File
	}
	return ""
}

// ValidKey reports whether a sound key is recognized.
func ValidKey(key string) bool {
	if key == "" {
		return true
	}
	_, ok := lookup[key]
	return ok
}
