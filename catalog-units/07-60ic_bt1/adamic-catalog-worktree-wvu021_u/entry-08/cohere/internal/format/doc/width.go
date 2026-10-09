package doc

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf16"
)

// surrogateBase is where UTF-16 surrogate code units live as runes, matching generate_string_width.
const surrogateBase = 0x100000

var (
	emojiPattern       = regexp.MustCompile(emojiPatternSource)
	narrowEmojiPattern = regexp.MustCompile(narrowEmojiPatternSource)
)

// StringWidth is upstream's getStringWidth, src/utilities/get-string-width.js: how many columns text
// takes, which is what decides whether a group fits.
//
// Emoji are matched first and removed, each counting 2 unless narrow-emojis says 1. What remains is
// walked by code point, skipping controls, combining marks and variation selectors, and counting wide
// and fullwidth characters as 2. Ambiguous-width characters count as 1, as upstream chooses.
//
// The emoji match runs over UTF-16 code units, one rune each, because the regex upstream uses has no
// `u` flag and spells astral emoji as surrogate pairs. See generate_string_width for the mapping.
func StringWidth(text string) int {
	if text == "" {
		return 0
	}
	if isPrintableASCII(text) {
		return len(text)
	}

	width := 0
	units := utf16.Encode([]rune(text))
	remaining := emojiPattern.ReplaceAllStringFunc(unitString(units), func(match string) string {
		if narrowEmojiPattern.MatchString(match) {
			width++
		} else {
			width += 2
		}
		return ""
	})

	for _, codePoint := range codePoints(remaining) {
		switch {
		case codePoint <= 0x1F, codePoint >= 0x7F && codePoint <= 0x9F:
			continue
		case codePoint >= 0x300 && codePoint <= 0x36F:
			continue
		case codePoint >= 0xFE00 && codePoint <= 0xFE0F:
			continue
		}
		if isWide(codePoint) {
			width += 2
		} else {
			width++
		}
	}
	return width
}

// isPrintableASCII is upstream's shortcut, !/[^\x20-\x7F]/.test(text). A tab or newline is outside the
// range and takes the slow path, as it does upstream.
func isPrintableASCII(text string) bool {
	for index := 0; index < len(text); index++ {
		if text[index] < 0x20 || text[index] > 0x7F {
			return false
		}
	}
	return true
}

// unitString renders UTF-16 code units as one rune each, surrogates moved to plane 16.
func unitString(units []uint16) string {
	var builder strings.Builder
	builder.Grow(len(units) * 3)
	for _, unit := range units {
		if unit >= 0xD800 && unit <= 0xDFFF {
			builder.WriteRune(rune(surrogateBase + int(unit) - 0xD800))
		} else {
			builder.WriteRune(rune(unit))
		}
	}
	return builder.String()
}

// codePoints reverses unitString and decodes the units the way JavaScript's for...of does: a valid
// surrogate pair is one code point, and a lone surrogate is a code point equal to its own unit.
func codePoints(mapped string) []rune {
	units := make([]uint16, 0, len(mapped))
	for _, character := range mapped {
		if character >= surrogateBase && character <= surrogateBase+0x7FF {
			units = append(units, uint16(int(character)-surrogateBase+0xD800))
		} else {
			units = append(units, uint16(character))
		}
	}
	result := make([]rune, 0, len(units))
	for index := 0; index < len(units); index++ {
		unit := units[index]
		if utf16.IsSurrogate(rune(unit)) && unit < 0xDC00 && index+1 < len(units) && units[index+1] >= 0xDC00 && units[index+1] <= 0xDFFF {
			result = append(result, utf16.DecodeRune(rune(unit), rune(units[index+1])))
			index++
			continue
		}
		result = append(result, rune(unit))
	}
	return result
}

// isWide is get-east-asian-width's isFullWidth || isWide, answered from the captured ranges.
func isWide(codePoint rune) bool {
	index := sort.Search(len(wideRanges), func(index int) bool { return wideRanges[index][1] >= codePoint })
	return index < len(wideRanges) && wideRanges[index][0] <= codePoint
}
