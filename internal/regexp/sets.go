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

func wordCharacter(character rune, flags Flags) bool {
	if unicodeMode(flags) && flags.IgnoreCase {
		character = canonicalize(character, flags)
	}
	return character == '_' ||
		character >= '0' && character <= '9' ||
		character >= 'a' && character <= 'z' ||
		character >= 'A' && character <= 'Z'
}

type characterSet struct {
	ranges  []RuneRange
	strings [][]rune
}

func normalizeRanges(ranges []RuneRange) []RuneRange {
	sort.Slice(ranges, func(left, right int) bool {
		if ranges[left].From != ranges[right].From {
			return ranges[left].From < ranges[right].From
		}
		return ranges[left].To < ranges[right].To
	})
	out := make([]RuneRange, 0, len(ranges))
	for _, runeRange := range ranges {
		if runeRange.From > runeRange.To {
			continue
		}
		if len(out) > 0 && runeRange.From <= out[len(out)-1].To+1 {
			if runeRange.To > out[len(out)-1].To {
				out[len(out)-1].To = runeRange.To
			}
		} else {
			out = append(out, runeRange)
		}
	}
	return out
}
func subtractRanges(left, right []RuneRange) []RuneRange {
	var out []RuneRange
	j := 0
	for _, runeRange := range left {
		start := runeRange.From
		for j < len(right) && right[j].To < start {
			j++
		}
		for k := j; k < len(right) && right[k].From <= runeRange.To; k++ {
			if right[k].From > start {
				out = append(out, RuneRange{start, right[k].From - 1})
			}
			if right[k].To >= start {
				start = right[k].To + 1
			}
			if start > runeRange.To {
				break
			}
		}
		if start <= runeRange.To {
			out = append(out, RuneRange{start, runeRange.To})
		}
	}
	return out
}
func intersectRanges(left, right []RuneRange) []RuneRange {
	var out []RuneRange
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		lo, hi := max(left[i].From, right[j].From), min(left[i].To, right[j].To)
		if lo <= hi {
			out = append(out, RuneRange{lo, hi})
		}
		if left[i].To < right[j].To {
			i++
		} else {
			j++
		}
	}
	return out
}
func foldSet(set characterSet, flags Flags) characterSet {
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
	if !flags.IgnoreCase {
		return set
	}
	// Canonicalize is identity except at the sparse generated mapping points.
	// Replace only those points, rather than expanding a complemented Unicode
	// set into a million temporary singleton ranges at compile time.
	table := legacyUppercase[:]
	if unicodeMode(flags) {
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
		for j, character := range s {
			copyS[j] = canonicalize(character, flags)
		}
		set.strings[i] = copyS
	}
	return set
}
func negateSet(set characterSet, flags Flags) characterSet {
	universe := characterSet{ranges: []RuneRange{{0, maxCharacter(flags)}}}
	if flags.UnicodeSets {
		universe = foldSet(universe, flags)
	}
	return characterSet{ranges: subtractRanges(universe.ranges, set.ranges)}
}
func characterValue(character *Character) rune {
	if character.Kind == Escape && len(character.Raw) == 2 {
		switch character.Raw[1] {
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
	return character.Value
}
func compileCharacter(character *Character, flags Flags, provider PropertyProvider) (characterSet, error) {
	var set characterSet
	switch character.Kind {
	case PropertyEscape:
		property := character.Raw[3 : len(character.Raw)-1]
		data, lookupError := provider.Lookup(property)
		if lookupError != nil {
			return set, lookupError
		}
		set = characterSet{ranges: normalizeRanges(slices.Clone(data.Ranges)), strings: slices.Clone(data.Strings)}
		for index, text := range set.strings {
			set.strings[index] = slices.Clone(text)
		}
		if character.Raw[1] == 'P' {
			if flags.UnicodeSets {
				set = foldSet(set, flags)
				set = negateSet(set, flags)
				return set, nil
			}
			set = negateSet(set, flags)
		}
	case ClassEscape:
		switch strings.ToLower(character.Raw[1:]) {
		case "d":
			set.ranges = []RuneRange{{'0', '9'}}
		case "w":
			set.ranges = []RuneRange{{'0', '9'}, {'A', 'Z'}, {'_', '_'}, {'a', 'z'}}
		case "s":
			set.ranges = []RuneRange{
				{9, 13},
				{32, 32},
				{0xa0, 0xa0},
				{0x1680, 0x1680},
				{0x2000, 0x200a},
				{0x2028, 0x2029},
				{0x202f, 0x202f},
				{0x205f, 0x205f},
				{0x3000, 0x3000},
				{0xfeff, 0xfeff},
			}
		}
		set = foldSet(set, flags)
		if character.Raw[1] >= 'A' &&
			character.Raw[1] <= 'Z' { // Complements are of the canonical character domain in both modes.
			universe := foldSet(characterSet{ranges: []RuneRange{{0, maxCharacter(flags)}}}, flags)
			set.ranges = subtractRanges(universe.ranges, set.ranges)
		}
		return set, nil
	default:
		value := characterValue(character)
		if value > 0xffff && !unicodeMode(flags) {
			h, l := utf16.EncodeRune(value)
			set.ranges = []RuneRange{{h, h}, {l, l}}
		} else {
			set.ranges = []RuneRange{{value, value}}
		}
	}
	return foldSet(set, flags), nil
}
func compileClass(expression ClassExpression, flags Flags, provider PropertyProvider) (characterSet, error) {
	switch typed := expression.(type) {
	case *ClassCharacter:
		return compileCharacter(typed.Character, flags, provider)
	case *ClassRange:
		return foldSet(
			characterSet{ranges: []RuneRange{{characterValue(typed.From), characterValue(typed.To)}}},
			flags,
		), nil
	case *ClassNegation:
		set, compileClassError := compileClass(typed.Operand, flags, provider)
		return negateSet(set, flags), compileClassError
	case *ClassString:
		set := characterSet{}
		for _, alternative := range typed.Alternatives {
			var str []rune
			for _, character := range alternative {
				str = append(str, characterValue(character))
			}
			if len(str) == 1 {
				set.ranges = append(set.ranges, RuneRange{str[0], str[0]})
			} else {
				set.strings = append(set.strings, str)
			}
		}
		set.ranges = normalizeRanges(set.ranges)
		return foldSet(set, flags), nil
	case *ClassUnion:
		result := characterSet{}
		for _, operand := range typed.Operands {
			set, compileClassError := compileClass(operand, flags, provider)
			if compileClassError != nil {
				return result, compileClassError
			}
			result.ranges = append(result.ranges, set.ranges...)
			result.strings = append(result.strings, set.strings...)
		}
		result.ranges = normalizeRanges(result.ranges)
		return result, nil
	case *ClassIntersection:
		a, compileClassError := compileClass(typed.Left, flags, provider)
		if compileClassError != nil {
			return a, compileClassError
		}
		b, compileClassError := compileClass(typed.Right, flags, provider)
		if compileClassError != nil {
			return b, compileClassError
		}
		return combineSets(a, b, true), nil
	case *ClassSubtraction:
		a, compileClassError := compileClass(typed.Left, flags, provider)
		if compileClassError != nil {
			return a, compileClassError
		}
		b, compileClassError := compileClass(typed.Right, flags, provider)
		if compileClassError != nil {
			return b, compileClassError
		}
		return combineSets(a, b, false), nil
	}
	return characterSet{}, fmt.Errorf("regexp: unsupported class %T", expression)
}
func combineSets(left, right characterSet, intersection bool) characterSet {
	result := characterSet{}
	if intersection {
		result.ranges = intersectRanges(left.ranges, right.ranges)
	} else {
		result.ranges = subtractRanges(left.ranges, right.ranges)
	}
	for _, s := range left.strings {
		contains := false
		for _, t := range right.strings {
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
func (set characterSet) contains(character rune) bool {
	index := sort.Search(len(set.ranges), func(index int) bool { return set.ranges[index].To >= character })
	return index < len(set.ranges) && set.ranges[index].From <= character
}
func (set characterSet) match(input []uint16, pos, dir int, flags Flags) []int {
	var positions []int
	for _, str := range set.strings {
		at := pos
		ok := true
		for j := 0; j < len(str); j++ {
			index := j
			if dir < 0 {
				index = len(str) - 1 - j
			}
			character, next, exists := readCharacter(input, at, dir, unicodeMode(flags))
			if !exists || canonicalize(character, flags) != str[index] {
				ok = false
				break
			}
			at = next
		}
		if ok {
			positions = append(positions, at)
		}
	}
	character, next, exists := readCharacter(input, pos, dir, unicodeMode(flags))
	if exists && set.contains(canonicalize(character, flags)) {
		positions = append(positions, next)
	}
	sort.SliceStable(
		positions,
		func(left, right int) bool { return abs(positions[left]-pos) > abs(positions[right]-pos) },
	)
	return slices.Compact(positions)
}
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
