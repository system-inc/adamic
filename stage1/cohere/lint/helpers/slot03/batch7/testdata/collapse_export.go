package tailwind

import (
	"fmt"
	"unicode/utf16"
)

func adamicUnits(text string) string {
	out := ""
	for _, u := range utf16.Encode([]rune(text)) {
		out += fmt.Sprintf("%d,", u)
	}
	return out
}
func AdamicRegistryState(registry *VariantRegistry, last int, present bool, group int) {
	registry.lastOrder = last
	if present {
		registry.groupOrder = &group
	} else {
		registry.groupOrder = nil
	}
}
func AdamicRegistrySnapshot(registry *VariantRegistry, name string, order int) []string {
	group := 0
	if registry.groupOrder != nil {
		group = *registry.groupOrder
	}
	lines := []string{fmt.Sprintf("%d:%t:%d:%d:%d", registry.lastOrder, registry.groupOrder != nil, group, len(registry.registrations), len(registry.comparisons))}
	if value, ok := registry.registrations[name]; ok {
		lines = append(lines, fmt.Sprintf("%s:%d:%s", adamicUnits(value.Name), value.Order, value.Kind))
	} else {
		lines = append(lines, "missing")
	}
	if fn, ok := registry.comparisons[order]; ok {
		lines = append(lines, fmt.Sprint(fn(ParsedVariant{Root: "abc"}, ParsedVariant{Root: "abcdefghijk"})))
	} else {
		lines = append(lines, "missing")
	}
	return lines
}
