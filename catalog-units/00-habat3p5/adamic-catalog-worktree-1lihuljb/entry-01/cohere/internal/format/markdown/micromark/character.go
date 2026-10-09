package micromark

import (
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
)

// The predicates of micromark-util-character. Upstream builds most of them from a regex tested against
// String.fromCharCode(code), which only ever sees one code unit, and guards with `code !== null &&
// code > -1`, so the virtual codes and the end of input are never members.

func asciiAlpha(code Code) bool {
	return (code >= 'A' && code <= 'Z') || (code >= 'a' && code <= 'z')
}

func asciiAlphanumeric(code Code) bool {
	return asciiDigit(code) || asciiAlpha(code)
}

// asciiAtext is /[#-'*+\--9=?A-Z^-~]/.
func asciiAtext(code Code) bool {
	return (code >= '#' && code <= '\'') || code == '*' || code == '+' || (code >= '-' && code <= '9') ||
		code == '=' || code == '?' || (code >= 'A' && code <= 'Z') || (code >= '^' && code <= '~')
}

func asciiControl(code Code) bool {
	// Special whitespace codes (which have negative values), C0 and Control character DEL.
	return code != CodeEof && (code < CodeSpace || code == CodeDel)
}

func asciiDigit(code Code) bool {
	return code >= '0' && code <= '9'
}

func asciiHexDigit(code Code) bool {
	return asciiDigit(code) || (code >= 'A' && code <= 'F') || (code >= 'a' && code <= 'f')
}

// asciiPunctuation is /[!-/:-@[-`{-~]/.
func asciiPunctuation(code Code) bool {
	return (code >= '!' && code <= '/') || (code >= ':' && code <= '@') || (code >= '[' && code <= '`') ||
		(code >= '{' && code <= '~')
}

func markdownLineEnding(code Code) bool {
	return code != CodeEof && code < CodeHorizontalTab
}

func markdownLineEndingOrSpace(code Code) bool {
	return code != CodeEof && (code < CodeNul || code == CodeSpace)
}

func markdownSpace(code Code) bool {
	return code == CodeHorizontalTab || code == CodeVirtualSpace || code == CodeSpace
}

func unicodePunctuation(code Code) bool {
	return code != CodeEof && code > -1 && inRanges(unicodePunctuationRanges, code)
}

func unicodeWhitespace(code Code) bool {
	return code != CodeEof && code > -1 && inRanges(unicodeWhitespaceRanges, code)
}

func inRanges(ranges [][2]uint16, code Code) bool {
	if code > 0xFFFF {
		return false
	}
	unit := uint16(code)
	index := sort.Search(len(ranges), func(index int) bool { return ranges[index][1] >= unit })
	return index < len(ranges) && ranges[index][0] <= unit
}

// classifyCharacter is micromark-util-classify-character: whitespace, punctuation, or neither (0).
func classifyCharacter(code Code) int {
	if code == CodeEof || markdownLineEndingOrSpace(code) || unicodeWhitespace(code) {
		return characterGroupWhitespace
	}

	if unicodePunctuation(code) {
		return characterGroupPunctuation
	}

	return 0
}

// NormalizeIdentifier is micromark-util-normalize-identifier.
//
// Upstream lowercases then uppercases with JavaScript's full case mapping, so that every form of a
// character folds to one. Go's ToUpper is the simple mapping and differs on a few characters (ß becomes
// SS in JavaScript). The identifier only keys definitions to references, and the printer prints labels,
// not identifiers, so the difference cannot reach output; it is noted rather than ported.
func NormalizeIdentifier(value string) string {
	var builder strings.Builder
	inWhitespace := false
	for _, character := range value {
		if character == '\t' || character == '\n' || character == '\r' || character == ' ' {
			if !inWhitespace {
				builder.WriteByte(' ')
			}
			inWhitespace = true
			continue
		}
		inWhitespace = false
		builder.WriteRune(character)
	}
	collapsed := strings.TrimSuffix(strings.TrimPrefix(builder.String(), " "), " ")
	return strings.ToUpper(strings.ToLower(collapsed))
}

// decodeNumericCharacterReference is micromark-util-decode-numeric-character-reference.
func decodeNumericCharacterReference(value string, base int) string {
	parsed, err := strconv.ParseInt(value, base, 64)
	code := int(parsed)
	if err != nil {
		// Number.parseInt of a too-long run is still a large number; every caller bounds the digits, so
		// an error here is only overflow, which upstream's last clause replaces.
		code = 0x110000
	}
	if code < int(CodeHt) ||
		code == int(CodeVt) ||
		(code > int(CodeCr) && code < int(CodeSpace)) ||
		(code > int(CodeTilde) && code < 160) ||
		(code > 55_295 && code < 57_344) ||
		(code > 64_975 && code < 65_008) ||
		(code&65_535) == 65_535 ||
		(code&65_535) == 65_534 ||
		code > 1_114_111 {
		return "�"
	}

	return string(rune(code))
}

// sliceUnits converts code units to a Go string.
func unitsToString(units []uint16) string {
	return string(utf16.Decode(units))
}

// utf16Length is JavaScript's String.length.
func utf16Length(text string) int {
	return len(utf16.Encode([]rune(text)))
}

// disabled reports whether a construct name is in disable.null.
func disabled(self *Self, name string) bool {
	for _, entry := range self.Parser.Constructs.Disable {
		if entry == name {
			return true
		}
	}
	return false
}
