package text

import "unicode"

// GraphemeCount is how many user-perceived characters a string holds, as `Intl.Segmenter` counts
// them: typescript-eslint's `getStringLength` and ESLint's `getGraphemeCount` both measure length this
// way for anything that is not plain ASCII.
//
// Go's standard library has no grapheme segmenter and this module takes no dependency for one, so
// this follows Unicode's extended grapheme cluster rules (UAX #29) for the shapes source text holds,
// from the character categories Go ships:
//
//   - CR LF is one cluster, and nothing else joins a control character, a line or paragraph
//     separator, or a format character that is not one of the joiners below.
//   - A combining or enclosing mark, a spacing mark, a zero width joiner or non-joiner, an emoji
//     skin tone modifier, a variation selector or a tag character continues the cluster before it.
//   - A pictograph after a zero width joiner continues a cluster that began with a pictograph, which
//     is what makes a family emoji one character.
//   - Regional indicators pair up, so a flag is one character and three flags are three.
//   - In six Indic scripts a virama joins the consonant after it, and Thai and Lao SARA AM joins
//     the letter before it.
//   - Hangul jamo join as a syllable: a leading consonant before a vowel or syllable, a vowel or
//     vowel syllable before a vowel or trailing consonant, a trailing one after either.
//
// Extended_Pictographic is not one of Go's tables, so a pictograph is read as a symbol (`So`) or a
// code point in the emoji blocks. Prepend characters, which begin a cluster in a few Indic scripts,
// are not handled. TestGraphemeCountAgreesWithIntlSegmenter scores this against Node's
// `Intl.Segmenter` on every shape it was written for.
func GraphemeCount(value string) int {
	count := 0
	var previous rune
	clusterStartsWithPictograph := false
	regionalIndicators := 0
	for index, character := range value {
		if index > 0 && continuesCluster(previous, character, clusterStartsWithPictograph, regionalIndicators) {
			if isRegionalIndicator(character) {
				regionalIndicators++
			}
			previous = character
			continue
		}
		count++
		clusterStartsWithPictograph = isPictograph(character)
		regionalIndicators = 0
		if isRegionalIndicator(character) {
			regionalIndicators = 1
		}
		previous = character
	}
	return count
}

// continuesCluster reports whether character belongs to the cluster previous is in.
func continuesCluster(previous rune, character rune, startsWithPictograph bool, regionalIndicators int) bool {
	switch {
	case previous == '\r' && character == '\n':
		return true
	case isGraphemeControl(previous):
		return false
	case isGraphemeExtend(character):
		return true
	case previous == zeroWidthJoiner && startsWithPictograph && isPictograph(character):
		return true
	case isRegionalIndicator(previous) && isRegionalIndicator(character):
		return regionalIndicators%2 == 1
	case isIndicLinker(previous) && unicode.Is(unicode.Lo, character) && previous>>7 == character>>7:
		return true
	}
	return hangulJoins(previous, character)
}

const zeroWidthJoiner = 0x200D

// isGraphemeExtend is UAX #29's Extend, ZWJ and SpacingMark together: what never begins a cluster.
func isGraphemeExtend(character rune) bool {
	switch {
	case unicode.In(character, unicode.Mn, unicode.Me, unicode.Mc):
		return true
	case character == 0x200C || character == zeroWidthJoiner:
		return true
	case character >= 0x1F3FB && character <= 0x1F3FF:
		// Emoji skin tone modifiers, which Go files under Sk rather than as marks.
		return true
	case character >= 0xE0020 && character <= 0xE007F:
		// Tag characters, the tail of a subdivision flag such as England's.
		return true
	case character == 0x0E33 || character == 0x0EB3:
		// Thai and Lao SARA AM, filed as letters and joined as spacing marks.
		return true
	}
	return false
}

// isGraphemeControl is UAX #29's Control, CR and LF: a cluster of its own, which nothing extends.
func isGraphemeControl(character rune) bool {
	if unicode.In(character, unicode.Cc, unicode.Zl, unicode.Zp) {
		return true
	}
	return unicode.Is(unicode.Cf, character) && !isGraphemeExtend(character)
}

// isIndicLinker is a virama UAX #29 lets join two consonants into one cluster (GB9c): Devanagari,
// Bengali, Gujarati, Oriya, Telugu and Malayalam. The consonant after it must be a letter in the same
// 128-code-point block, which is how these scripts are laid out.
func isIndicLinker(character rune) bool {
	switch character {
	case 0x094D, 0x09CD, 0x0ACD, 0x0B4D, 0x0C4D, 0x0D4D:
		return true
	}
	return false
}

func isRegionalIndicator(character rune) bool {
	return character >= 0x1F1E6 && character <= 0x1F1FF
}

// isPictograph stands in for Extended_Pictographic, which Go does not ship.
func isPictograph(character rune) bool {
	switch {
	case character >= 0x1F000 && character <= 0x1FAFF && !isRegionalIndicator(character) &&
		!(character >= 0x1F3FB && character <= 0x1F3FF):
		return true
	case character >= 0x2600 && character <= 0x27BF:
		return true
	}
	return unicode.Is(unicode.So, character)
}

// The Hangul jamo classes. A precomposed syllable is LV when it has no trailing consonant and LVT
// when it does, which its offset in the syllable block says.
func hangulLeading(character rune) bool { return character >= 0x1100 && character <= 0x115F }
func hangulVowel(character rune) bool   { return character >= 0x1160 && character <= 0x11A7 }
func hangulTrailing(character rune) bool {
	return character >= 0x11A8 && character <= 0x11FF
}
func hangulSyllable(character rune) bool { return character >= 0xAC00 && character <= 0xD7A3 }
func hangulVowelSyllable(character rune) bool {
	return hangulSyllable(character) && (character-0xAC00)%28 == 0
}

func hangulJoins(previous rune, character rune) bool {
	switch {
	case hangulLeading(previous):
		return hangulLeading(character) || hangulVowel(character) || hangulSyllable(character)
	case hangulVowel(previous) || hangulVowelSyllable(previous):
		return hangulVowel(character) || hangulTrailing(character)
	case hangulTrailing(previous) || hangulSyllable(previous):
		return hangulTrailing(character)
	}
	return false
}
