// Package unicodeproperties is the data behind ECMA-262's \p and \P escapes:
// the General_Category, Script and Script_Extensions values, the binary
// properties in the spec's table of binary Unicode properties, and the
// properties of strings the v flag admits.
//
// The code-point properties are cohere's unicodeproperties, the one home for
// them, read through its Lookup; this package keeps no copy. The properties of
// strings and the Canonicalize tables are generated here, from the Unicode
// Character Database of the same version, which is the version Node reports
// (process.versions.unicode: 17.0, whose data files are 17.0.0). Names are
// matched exactly, the way the spec's UnicodeMatchProperty does: case, spaces,
// hyphens and underscores all count, and an alias the spec does not list does
// not match. Node is the oracle for membership, not the source of the tables;
// internal/unicodeproperties/node_test.go is that check, of the sets this
// package answers.
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

// tables.go (the properties of strings) and canonicalize_tables.go are generated
// from the Unicode Character Database. TestVersionMatchesNode fails when the Node
// the oracle runs has moved to a different Unicode version.
//go:generate go run generate.go

import (
	"sort"
	"sync"

	cohere "github.com/system-inc/cohere/unicode_properties"
)

// Version is the Unicode Character Database version every table here is from.
// cohere's is the source: its code-point tables are the ones Lookup reads, and
// generate.go writes the strings and Canonicalize tables from that version's
// files. NodeUnicodeVersion is what process.versions.unicode prints for it
// (Node says "17.0" for 17.0.0).
const (
	Version            = cohere.Version
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

// Property is one resolved \p{...} escape. Every spelling of one property
// resolves to the same Set, or for a property of strings the same Sequences,
// so that identity is the property's: "gc=Ll" and "Lowercase_Letter" share a
// Set, and "sc=Latin" and "scx=Latin" do not.
type Property struct {
	Kind Kind
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
//
// A code-point property is cohere's: its Lookup takes the same exact spellings
// (#cz54235 held its 1,715 Names, and each one's ranges, to what this package's
// own tables accepted before they went). The spec reads a lone name as a
// General_Category value, then a binary property, then (v flag only) a property
// of strings. No name is two of those today, so cohere's single table and then
// the strings is that order.
func Lookup(expression string, unicodeSets bool) (Property, bool) {
	if property, ok := cohere.Lookup(expression); ok {
		// No cohere property is empty; one would be an empty Set.
		set := &Set{}
		if len(property.Ranges) > 0 {
			set = codePointSets()[&property.Ranges[0]]
		}
		// codePointSets is built from cohere's Names, so a spelling its Lookup
		// takes and its Names leaves out has no Set. That's cohere's two lists
		// disagreeing, and it stops here by name rather than as a nil Set read
		// far from the cause (TestEveryNameHasASet).
		if set == nil {
			panic("unicodeproperties: cohere's Lookup takes " + expression + ", which its Names doesn't list")
		}
		return Property{Kind: KindCodePoints, Set: set}, true
	}
	if unicodeSets {
		if sequences, ok := stringProperties[expression]; ok {
			return Property{Kind: KindStrings, Sequences: sequences}, true
		}
	}
	return Property{}, false
}

// codePointSets is each of cohere's properties as a Set, keyed by the address
// of its first range. cohere's spellings of one property share its ranges, so
// they share one Set here. Its ranges are cohere's, sorted and disjoint, and
// cohere merges adjacent ones, which is what a Set needs (TestSetBoundaries
// fails the day one isn't). Made once, on the first Lookup.
var codePointSets = sync.OnceValue(func() map[*cohere.Range]*Set {
	sets := map[*cohere.Range]*Set{}
	for _, name := range cohere.Names() {
		property, _ := cohere.Lookup(name)
		if len(property.Ranges) == 0 || sets[&property.Ranges[0]] != nil {
			continue
		}
		set := &Set{Ranges: make([]Range, len(property.Ranges))}
		for index, interval := range property.Ranges {
			set.Ranges[index] = Range{Start: uint32(interval.Lo), End: uint32(interval.Hi)}
		}
		sets[&property.Ranges[0]] = set
	}
	return sets
})
