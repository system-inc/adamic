package regexp

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode/utf16"
)

// RuneRange is an inclusive interval in a property's code point set.
type RuneRange struct{ From, To rune }

// PropertySet includes both character ranges and the v flag's strings.
type PropertySet struct {
	Ranges  []RuneRange
	Strings [][]rune
}

// PropertyProvider resolves exact ECMA-262 aliases. An unavailable property must
// return an error, never silently an empty set. Compilation snapshots both
// ranges and strings; providers may reuse their storage.
type PropertyProvider interface {
	Lookup(property string) (PropertySet, error)
}

func wordCharacter(c rune, f Flags) bool {
	if unicodeMode(f) && f.IgnoreCase {
		c = canonicalize(c, f)
	}
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

type characterSet struct {
	ranges  []RuneRange
	strings [][]rune
}

func normalizeRanges(ranges []RuneRange) []RuneRange {
	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].From != ranges[j].From {
			return ranges[i].From < ranges[j].From
		}
		return ranges[i].To < ranges[j].To
	})
	out := make([]RuneRange, 0, len(ranges))
	for _, r := range ranges {
		if r.From > r.To {
			continue
		}
		if len(out) > 0 && r.From <= out[len(out)-1].To+1 {
			if r.To > out[len(out)-1].To {
				out[len(out)-1].To = r.To
			}
		} else {
			out = append(out, r)
		}
	}
	return out
}
func subtractRanges(a, b []RuneRange) []RuneRange {
	var out []RuneRange
	j := 0
	for _, r := range a {
		start := r.From
		for j < len(b) && b[j].To < start {
			j++
		}
		for k := j; k < len(b) && b[k].From <= r.To; k++ {
			if b[k].From > start {
				out = append(out, RuneRange{start, b[k].From - 1})
			}
			if b[k].To >= start {
				start = b[k].To + 1
			}
			if start > r.To {
				break
			}
		}
		if start <= r.To {
			out = append(out, RuneRange{start, r.To})
		}
	}
	return out
}
func intersectRanges(a, b []RuneRange) []RuneRange {
	var out []RuneRange
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		lo, hi := max(a[i].From, b[j].From), min(a[i].To, b[j].To)
		if lo <= hi {
			out = append(out, RuneRange{lo, hi})
		}
		if a[i].To < b[j].To {
			i++
		} else {
			j++
		}
	}
	return out
}
func foldSet(set characterSet, f Flags) characterSet {
	var stringsSet [][]rune
	for _, str := range set.strings {
		if len(str) == 1 {
			set.ranges = append(set.ranges, RuneRange{str[0], str[0]})
		} else {
			stringsSet = append(stringsSet, str)
		}
	}
	set.strings = stringsSet
	set.ranges = normalizeRanges(set.ranges)
	if !f.IgnoreCase {
		return set
	}
	// Canonicalize is identity except at the sparse generated mapping points.
	// Replace only those points, rather than expanding a complemented Unicode
	// set into a million temporary singleton ranges at compile time.
	table := legacyUppercase[:]
	if unicodeMode(f) {
		table = simpleCaseFold[:]
	}
	var sources, targets []RuneRange
	for _, pair := range table {
		if set.contains(pair[0]) {
			sources = append(sources, RuneRange{pair[0], pair[0]})
			targets = append(targets, RuneRange{pair[1], pair[1]})
		}
	}
	set.ranges = normalizeRanges(append(subtractRanges(set.ranges, sources), targets...))
	for i, s := range set.strings {
		copyS := slices.Clone(s)
		for j, c := range s {
			copyS[j] = canonicalize(c, f)
		}
		set.strings[i] = copyS
	}
	return set
}
func negateSet(set characterSet, f Flags) characterSet {
	universe := characterSet{ranges: []RuneRange{{0, maxCharacter(f)}}}
	if f.UnicodeSets {
		universe = foldSet(universe, f)
	}
	return characterSet{ranges: subtractRanges(universe.ranges, set.ranges)}
}
func characterValue(c *Character) rune {
	if c.Kind == Escape && len(c.Raw) == 2 {
		switch c.Raw[1] {
		case 'f':
			return 12
		case 'n':
			return 10
		case 'r':
			return 13
		case 't':
			return 9
		case 'v':
			return 11
		}
	}
	return c.Value
}
func compileCharacter(c *Character, f Flags, p PropertyProvider) (characterSet, error) {
	var set characterSet
	switch c.Kind {
	case PropertyEscape:
		property := c.Raw[3 : len(c.Raw)-1]
		data, err := p.Lookup(property)
		if err != nil {
			return set, err
		}
		set = characterSet{ranges: normalizeRanges(slices.Clone(data.Ranges)), strings: slices.Clone(data.Strings)}
		for index, text := range set.strings {
			set.strings[index] = slices.Clone(text)
		}
		if c.Raw[1] == 'P' {
			if f.UnicodeSets {
				set = foldSet(set, f)
				set = negateSet(set, f)
				return set, nil
			}
			set = negateSet(set, f)
		}
	case ClassEscape:
		switch strings.ToLower(c.Raw[1:]) {
		case "d":
			set.ranges = []RuneRange{{'0', '9'}}
		case "w":
			set.ranges = []RuneRange{{'0', '9'}, {'A', 'Z'}, {'_', '_'}, {'a', 'z'}}
		case "s":
			set.ranges = []RuneRange{{9, 13}, {32, 32}, {0xa0, 0xa0}, {0x1680, 0x1680}, {0x2000, 0x200a}, {0x2028, 0x2029}, {0x202f, 0x202f}, {0x205f, 0x205f}, {0x3000, 0x3000}, {0xfeff, 0xfeff}}
		}
		set = foldSet(set, f)
		if c.Raw[1] >= 'A' && c.Raw[1] <= 'Z' { // Complements are of the canonical character domain in both modes.
			universe := foldSet(characterSet{ranges: []RuneRange{{0, maxCharacter(f)}}}, f)
			set.ranges = subtractRanges(universe.ranges, set.ranges)
		}
		return set, nil
	default:
		value := characterValue(c)
		if value > 0xffff && !unicodeMode(f) {
			h, l := utf16.EncodeRune(value)
			set.ranges = []RuneRange{{h, h}, {l, l}}
		} else {
			set.ranges = []RuneRange{{value, value}}
		}
	}
	return foldSet(set, f), nil
}
func compileClass(e ClassExpression, f Flags, p PropertyProvider) (characterSet, error) {
	switch e := e.(type) {
	case *ClassCharacter:
		return compileCharacter(e.Character, f, p)
	case *ClassRange:
		return foldSet(characterSet{ranges: []RuneRange{{characterValue(e.From), characterValue(e.To)}}}, f), nil
	case *ClassNegation:
		s, err := compileClass(e.Operand, f, p)
		return negateSet(s, f), err
	case *ClassString:
		set := characterSet{}
		for _, s := range e.Alternatives {
			var str []rune
			for _, c := range s {
				str = append(str, characterValue(c))
			}
			if len(str) == 1 {
				set.ranges = append(set.ranges, RuneRange{str[0], str[0]})
			} else {
				set.strings = append(set.strings, str)
			}
		}
		set.ranges = normalizeRanges(set.ranges)
		return foldSet(set, f), nil
	case *ClassUnion:
		result := characterSet{}
		for _, e := range e.Operands {
			s, err := compileClass(e, f, p)
			if err != nil {
				return result, err
			}
			result.ranges = append(result.ranges, s.ranges...)
			result.strings = append(result.strings, s.strings...)
		}
		result.ranges = normalizeRanges(result.ranges)
		return result, nil
	case *ClassIntersection:
		a, err := compileClass(e.Left, f, p)
		if err != nil {
			return a, err
		}
		b, err := compileClass(e.Right, f, p)
		if err != nil {
			return b, err
		}
		return combineSets(a, b, true), nil
	case *ClassSubtraction:
		a, err := compileClass(e.Left, f, p)
		if err != nil {
			return a, err
		}
		b, err := compileClass(e.Right, f, p)
		if err != nil {
			return b, err
		}
		return combineSets(a, b, false), nil
	}
	return characterSet{}, fmt.Errorf("regexp: unsupported class %T", e)
}
func combineSets(a, b characterSet, intersection bool) characterSet {
	result := characterSet{}
	if intersection {
		result.ranges = intersectRanges(a.ranges, b.ranges)
	} else {
		result.ranges = subtractRanges(a.ranges, b.ranges)
	}
	for _, s := range a.strings {
		contains := false
		for _, t := range b.strings {
			if slices.Equal(s, t) {
				contains = true
				break
			}
		}
		if contains == intersection {
			result.strings = append(result.strings, s)
		}
	}
	return result
}
func (s characterSet) contains(c rune) bool {
	i := sort.Search(len(s.ranges), func(i int) bool { return s.ranges[i].To >= c })
	return i < len(s.ranges) && s.ranges[i].From <= c
}
func (s characterSet) match(input []uint16, pos, dir int, f Flags) []int {
	var positions []int
	for _, str := range s.strings {
		at := pos
		ok := true
		for j := 0; j < len(str); j++ {
			index := j
			if dir < 0 {
				index = len(str) - 1 - j
			}
			c, next, exists := readCharacter(input, at, dir, unicodeMode(f))
			if !exists || canonicalize(c, f) != str[index] {
				ok = false
				break
			}
			at = next
		}
		if ok {
			positions = append(positions, at)
		}
	}
	c, next, exists := readCharacter(input, pos, dir, unicodeMode(f))
	if exists && s.contains(canonicalize(c, f)) {
		positions = append(positions, next)
	}
	sort.SliceStable(positions, func(i, j int) bool { return abs(positions[i]-pos) > abs(positions[j]-pos) })
	return slices.Compact(positions)
}
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
