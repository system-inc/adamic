package regexp

import (
	"fmt"
	"strings"
	"unicode"
)

// UnavailablePropertyError refuses a valid property whose data is not installed.
// It is separate from syntax errors and a completed match returning nil.
type UnavailablePropertyError struct{ Property, Version string }

func (propertyError *UnavailablePropertyError) Error() string {
	return fmt.Sprintf(
		"regexp: Unicode property %q unavailable in Go stand-in (%s)",
		propertyError.Property,
		propertyError.Version,
	)
}

// GoProperties is a STAND-IN using Go's Unicode version, General_Category and
// Script only. It deliberately refuses Script_Extensions and binary/string
// properties until the generated UCD provider lands. Alias spelling is exact.
type GoProperties struct{}

func (GoProperties) Lookup(property string) (PropertySet, error) {
	name, value, hasValue := strings.Cut(property, "=")
	var table *unicode.RangeTable
	if !hasValue || name == "gc" || name == "General_Category" {
		category := property
		if hasValue {
			category = value
		}
		if alias := unicode.CategoryAliases[category]; alias != "" {
			category = alias
		}
		table = unicode.Categories[category]
	} else if name == "sc" || name == "Script" {
		if alias := scriptAliases[value]; alias != "" {
			value = alias
		}
		table = unicode.Scripts[value]
		// Go omits the Unknown script table. Its complement is exact Script=Unknown,
		// not Script_Extensions, which this stand-in never guesses.
		if value == "Unknown" {
			var known []RuneRange
			for _, value := range unicode.Scripts {
				known = append(known, goRanges(value)...)
			}
			return PropertySet{Ranges: subtractRanges([]RuneRange{{0, 0x10ffff}}, normalizeRanges(known))}, nil
		}
	}
	if table == nil {
		return PropertySet{}, &UnavailablePropertyError{property, unicode.Version}
	}
	return PropertySet{Ranges: normalizeRanges(goRanges(table))}, nil
}
func goRanges(table *unicode.RangeTable) []RuneRange {
	var out []RuneRange
	for _, range16 := range table.R16 {
		if range16.Stride == 1 {
			out = append(out, RuneRange{rune(range16.Lo), rune(range16.Hi)})
		} else {
			for c := uint32(range16.Lo); c <= uint32(range16.Hi); c += uint32(range16.Stride) {
				out = append(out, RuneRange{rune(c), rune(c)})
			}
		}
	}
	for _, range32 := range table.R32 {
		if range32.Stride == 1 {
			out = append(out, RuneRange{rune(range32.Lo), rune(range32.Hi)})
		} else {
			for c := range32.Lo; c <= range32.Hi; c += range32.Stride {
				out = append(out, RuneRange{rune(c), rune(c)})
			}
		}
	}
	return out
}
