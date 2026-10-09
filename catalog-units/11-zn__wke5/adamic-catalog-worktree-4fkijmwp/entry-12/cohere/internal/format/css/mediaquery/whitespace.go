package mediaquery

import (
	"strings"
	"unicode/utf8"
)

// The library's whitespace is JavaScript's: the regular expression class \s, and String.prototype.trim,
// which strip the same set (ECMAScript WhiteSpace and LineTerminator). It is not Go's unicode.IsSpace,
// which takes U+0085 and leaves out U+FEFF. Every member is in the Basic Multilingual Plane, so one
// UTF-16 unit in JavaScript is one rune here and testing runes answers as the library's unit tests do.

// isWhitespace is `character.search(/\s/) !== -1` for one character.
func isWhitespace(character rune) bool {
	switch character {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0xa0, 0x1680, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000, 0xfeff:
		return true
	}
	return character >= 0x2000 && character <= 0x200a
}

// leadingWhitespace is `/^(\s*)/.exec(text)[1]`.
func leadingWhitespace(text string) string {
	return text[:len(text)-len(strings.TrimLeftFunc(text, isWhitespace))]
}

// trailingWhitespace is `/(\s*)$/.exec(text)[1]`: the first position where \s* reaches the end, which
// is where the trailing run of whitespace starts.
func trailingWhitespace(text string) string {
	return text[len(strings.TrimRightFunc(text, isWhitespace)):]
}

// trim is String.prototype.trim.
func trim(text string) string {
	return strings.TrimFunc(text, isWhitespace)
}

// startsWithWhitespace is `text[0].search(/\s/) !== -1` for a non-empty text.
func startsWithWhitespace(text string) bool {
	character, _ := utf8.DecodeRuneInString(text)
	return isWhitespace(character)
}
