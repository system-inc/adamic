package tailwind

import (
	"fmt"
	"unicode/utf16"
)

func AdamicWave7Dissect(value string) ([]any, []string) {
	variants, base, important := dissectClass(value)
	units := func(value string) string {
		codes := utf16.Encode([]rune(value))
		out := fmt.Sprintf("%d:", len(codes))
		for _, code := range codes {
			out += fmt.Sprintf("%d,", code)
		}
		return out
	}
	return []any{value}, []string{units(variants), units(base), fmt.Sprint(important)}
}
