package tailwind

import "strings"

func AdamicWave5Named(value string) bool { return isValidNamedValue(value) }

func AdamicWave5Arbitrary(value string) bool { return isValidArbitrary(value) }

func AdamicWave5Theme(params string) (ThemeOptions, string) { return parseThemeOptions(params) }
func AdamicWave5ThemeSegments(params string) []string       { return segment(strings.TrimSpace(params), ' ') }
