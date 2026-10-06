// Package unicodeproperties is the data behind ECMA-262's \p and \P escapes:
// the General_Category, Script and Script_Extensions values, the binary
// properties in the spec's table of binary Unicode properties, and the
// properties of strings the v flag admits.
//
// The tables are generated from the Unicode Character Database of the version
// Node reports (process.versions.unicode: 17.0, whose data files are 17.0.0).
// Names are matched exactly, the way the spec's UnicodeMatchProperty does:
// case, spaces, hyphens and underscores all count, and an alias the spec does
// not list does not match. Node is the oracle for membership, not the source
// of the tables; internal/unicodeproperties/node_test.go is that check.
//
// \P is the complement of what Lookup returns, over every code point from 0 to
// U+10FFFF, surrogates included. A property of strings has no code-point
// complement; the caller rejects \P of one, which the spec makes a syntax error.
//
// Case-insensitive matching is Canonicalize (ECMA-262 22.2.2.7.3), not these
// sets. CanonicalizeUnicode and CanonicalizeLegacy are that operation for the
// u or v flag and for the flagless i, and UnicodeEquivalents and
// LegacyEquivalents are the sets a character class has to close over.
package unicodeproperties

// tables.go is generated from the Unicode Character Database. TestVersionMatchesNode
// fails when the Node the oracle runs has moved to a different Unicode version.
//go:generate go run generate.go

import "sort"

// Version is the Unicode Character Database version these tables were built from.
// NodeUnicodeVersion is what process.versions.unicode prints for that version
// (Node says "17.0" for 17.0.0).
const (
	Version            = "17.0.0"
	NodeUnicodeVersion = "17.0"
)

// Range is an inclusive interval of code points. A Set's ranges are sorted,
// non-overlapping and not adjacent, so [1,2][3,4] is stored as [1,4].
type Range struct {
	Start uint32
	End   uint32
}

// Set is a set of code points, including surrogates. Nil is empty.
type Set struct {
	Ranges []Range
}

// Contains reports whether codePoint is in the set. Values outside 0..U+10FFFF
// are not, and a negative rune is not a code point.
func (s *Set) Contains(codePoint rune) bool {
	if s == nil || codePoint < 0 || codePoint > 0x10FFFF {
		return false
	}
	cp := uint32(codePoint)
	ranges := s.Ranges
	lo, hi := 0, len(ranges)
	for lo < hi {
		mid := (lo + hi) / 2
		if ranges[mid].End < cp {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo < len(ranges) && ranges[lo].Start <= cp
}

// Complement is the set of code points from 0 to U+10FFFF that are not in s.
// That is what \P matches when \p matches s.
func (s *Set) Complement() *Set {
	ranges := []Range(nil)
	if s != nil {
		ranges = s.Ranges
	}
	var out []Range
	next := uint32(0)
	for _, r := range ranges {
		if r.Start > next {
			out = append(out, Range{Start: next, End: r.Start - 1})
		}
		if r.End == 0x10FFFF {
			return &Set{Ranges: out}
		}
		next = r.End + 1
	}
	if next <= 0x10FFFF {
		out = append(out, Range{Start: next, End: 0x10FFFF})
	}
	return &Set{Ranges: out}
}

// Kind distinguishes a property that matches code points from one that matches
// strings (the v flag's RGI_Emoji and its kin).
type Kind int

const (
	KindCodePoints Kind = iota
	KindStrings
)

// Property is one resolved \p{...} escape.
type Property struct {
	Kind Kind
	// Name is the canonical property name: "ASCII", "General_Category",
	// "Script", "Script_Extensions", "RGI_Emoji".
	Name string
	// Value is the canonical property value. Binary properties and properties
	// of strings are "True", the only value the spec's compilation accepts for
	// them (Node rejects \p{ASCII=Yes}). General_Category, Script and
	// Script_Extensions use the long name from PropertyValueAliases.txt
	// ("Lowercase_Letter", "Latin").
	Value string
	// Set is non-nil for KindCodePoints. Callers must not modify it.
	Set *Set
	// Sequences is non-nil for KindStrings: each string is one sequence the
	// property matches, sorted in code-point order. Callers must not modify it.
	Sequences []string
}

// Contains reports whether a code point is in a code-point property.
func (p Property) Contains(codePoint rune) bool {
	if p.Kind != KindCodePoints {
		return false
	}
	return p.Set.Contains(codePoint)
}

// ContainsSequence reports whether text is one of a string property's sequences.
func (p Property) ContainsSequence(text string) bool {
	if p.Kind != KindStrings {
		return false
	}
	i := sort.Search(len(p.Sequences), func(i int) bool {
		return p.Sequences[i] >= text
	})
	return i < len(p.Sequences) && p.Sequences[i] == text
}

// Lookup resolves the text inside \p{...}. unicodeSets is true for a pattern
// with the v flag; properties of strings exist only then.
//
// The expression is taken exactly. "gc=Ll", "General_Category=Lowercase_Letter"
// and "Ll" are the same set; "ll", "Latin" and "ascii" are not properties.
// "Script=Latin" is, and "scx=Latn" is the extensions of that script, a
// different set. A binary property is only the bare name ("White_Space",
// "space", "WSpace"), never "White_Space=Yes".
func Lookup(expression string, unicodeSets bool) (Property, bool) {
	if !expressionOK(expression) {
		return Property{}, false
	}
	if name, value, isPair := splitPair(expression); isPair {
		var table map[string]entry
		switch name {
		case "General_Category", "gc":
			table = generalCategoryNames
		case "Script", "sc":
			table = scriptNames
		case "Script_Extensions", "scx":
			table = scriptExtensionNames
		default:
			return Property{}, false
		}
		return codePoint(table, value)
	}
	// A lone name is a General_Category value first, then a binary property,
	// then (v flag only) a property of strings. The spec checks in that order.
	// No value of one is a name of another today; the order is what would decide.
	if p, ok := codePoint(generalCategoryNames, expression); ok {
		return p, true
	}
	if p, ok := codePoint(binaryNames, expression); ok {
		return p, true
	}
	if unicodeSets {
		if s, ok := stringProperties[expression]; ok {
			return Property{Kind: KindStrings, Name: s.name, Value: "True", Sequences: s.sequences}, true
		}
	}
	return Property{}, false
}

func codePoint(table map[string]entry, name string) (Property, bool) {
	e, ok := table[name]
	if !ok {
		return Property{}, false
	}
	return Property{Kind: KindCodePoints, Name: e.name, Value: e.value, Set: sets[e.set]}, true
}

// expressionOK is the lexical shape of UnicodePropertyValueExpression: letters,
// digits and underscores, with at most one "=". The grammar admits nothing else,
// so a space or a hyphen is a different expression and not a property.
func expressionOK(expression string) bool {
	if expression == "" || expression == "=" {
		return false
	}
	equals := 0
	for i := 0; i < len(expression); i++ {
		c := expression[i]
		switch {
		case c == '=':
			equals++
			if equals > 1 || i == 0 || i == len(expression)-1 {
				return false
			}
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '_':
		default:
			return false
		}
	}
	return true
}

func splitPair(expression string) (string, string, bool) {
	for i := 0; i < len(expression); i++ {
		if expression[i] == '=' {
			return expression[:i], expression[i+1:], true
		}
	}
	return "", "", false
}

// entry is one alias's canonical name, canonical value and set index.
type entry struct {
	name  string
	value string
	set   int
}

type sequences struct {
	name      string
	sequences []string
}
