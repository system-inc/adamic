package unicodeproperties

import "sort"

// CanonicalizeUnicode is Canonicalize for the i flag together with u or v
// (ECMA-262 22.2.2.7.3): simple case folding, the C and S mappings in
// CaseFolding.txt. Status F would grow the string (ß folds to ss) and status T
// is Turkic; neither is used. A code point with no such mapping returns itself,
// and so does an argument that is not a code point.
func CanonicalizeUnicode(codePoint rune) rune {
	if codePoint < 0 || codePoint > 0x10FFFF {
		return codePoint
	}
	cp := uint32(codePoint)
	index := sort.Search(len(unicodeFold), func(i int) bool { return unicodeFold[i][0] >= cp })
	if index < len(unicodeFold) && unicodeFold[index][0] == cp {
		return rune(unicodeFold[index][1])
	}
	return codePoint
}

// UnicodeEquivalents is every code point that CanonicalizeUnicode maps to the
// same value as codePoint, sorted, including codePoint. Matching a character
// class under i with u or v compares those canonical values, so a class that
// contains codePoint also matches this set. The slice is newly allocated. An
// argument that is not a code point returns nil.
func UnicodeEquivalents(codePoint rune) []rune {
	if codePoint < 0 || codePoint > 0x10FFFF {
		return nil
	}
	canon := uint32(CanonicalizeUnicode(codePoint))
	index := sort.Search(len(unicodeClassCanon), func(i int) bool { return unicodeClassCanon[i] >= canon })
	if index < len(unicodeClassCanon) && unicodeClassCanon[index] == canon {
		members := unicodeClass[index]
		out := make([]rune, len(members))
		for i, member := range members {
			out[i] = rune(member)
		}
		return out
	}
	return []rune{codePoint}
}

// CanonicalizeLegacy is Canonicalize for the i flag without u or v. codeUnit is
// one UTF-16 code unit. The full locale-independent uppercase of that unit is
// kept only when it is a single code unit and does not turn a non-ASCII unit
// into an ASCII one. U+017F (long s) uppercases to "S" and U+0131 (dotless i)
// uppercases to "I"; both stay themselves because of that second filter, so
// neither matches the ASCII letter. A supplementary character is two code units
// and is not an input here.
func CanonicalizeLegacy(codeUnit uint16) uint16 {
	index := sort.Search(len(legacyFold), func(i int) bool { return legacyFold[i][0] >= codeUnit })
	if index < len(legacyFold) && legacyFold[index][0] == codeUnit {
		return legacyFold[index][1]
	}
	return codeUnit
}

// LegacyEquivalents is every code unit that CanonicalizeLegacy maps to the same
// value as codeUnit, sorted, including codeUnit. A character class under i
// without u or v closes over this set. The slice is newly allocated.
func LegacyEquivalents(codeUnit uint16) []uint16 {
	canon := CanonicalizeLegacy(codeUnit)
	index := sort.Search(len(legacyClassCanon), func(i int) bool { return legacyClassCanon[i] >= canon })
	if index < len(legacyClassCanon) && legacyClassCanon[index] == canon {
		members := legacyClass[index]
		out := make([]uint16, len(members))
		copy(out, members)
		return out
	}
	return []uint16{codeUnit}
}
