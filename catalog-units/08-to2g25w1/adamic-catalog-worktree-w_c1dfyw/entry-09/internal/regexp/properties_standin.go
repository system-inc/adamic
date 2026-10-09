package regexp

import (
	"fmt"
	"strings"
	"unicode"
)

// UnavailablePropertyError refuses a valid property whose data is not installed.
// It is separate from syntax errors and a completed match returning nil.
type UnavailablePropertyError struct{ Property, Version string }

func (e *UnavailablePropertyError) Error() string {
	return fmt.Sprintf("regexp: Unicode property %q unavailable in Go stand-in (%s)", e.Property, e.Version)
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
			for _, t := range unicode.Scripts {
				known = append(known, goRanges(t)...)
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
	for _, r := range table.R16 {
		if r.Stride == 1 {
			out = append(out, RuneRange{rune(r.Lo), rune(r.Hi)})
		} else {
			for c := uint32(r.Lo); c <= uint32(r.Hi); c += uint32(r.Stride) {
				out = append(out, RuneRange{rune(c), rune(c)})
			}
		}
	}
	for _, r := range table.R32 {
		if r.Stride == 1 {
			out = append(out, RuneRange{rune(r.Lo), rune(r.Hi)})
		} else {
			for c := r.Lo; c <= r.Hi; c += r.Stride {
				out = append(out, RuneRange{rune(c), rune(c)})
			}
		}
	}
	return out
}
