package printing

import (
	"strings"
	"unicode/utf8"
)

/*
 * Upstream's text-scanning utilities, src/utilities/skip.js and its neighbours, on byte offsets.
 *
 * Upstream steps through text by UTF-16 code unit with charAt. Here positions are byte offsets into
 * UTF-8, so these step by rune instead: forward by decoding the rune at the cursor, backward by decoding
 * the rune that ends at it. For ASCII the two agree exactly. They differ for the characters these
 * functions actually test that are not ASCII: U+2028 and U+2029, which skipNewline treats as newlines,
 * and the Unicode whitespace JavaScript's \s matches (U+00A0, U+3000, U+FEFF and the rest). A byte-wise
 * port would land inside those characters and report a position that is not a character boundary.
 *
 * Upstream returns false for "no position". Here that is -2, notFound, so every function stays one int
 * in and one int out like upstream's; -1 and len(text) remain the real "ran off the start" and "ran off
 * the end" results.
 */

// notFound is upstream's `false` position.
const notFound = -2

// runeAt returns the rune whose first byte is at cursor, and its width.
func runeAt(text string, cursor int) (rune, int) {
	return utf8.DecodeRuneInString(text[cursor:])
}

// runeBefore returns the rune that ends at cursor+1 when stepping backwards from a rune starting at
// cursor, which is the rune at cursor itself; and the start of the rune before it.
func runeBackwards(text string, cursor int) (rune, int) {
	character, _ := runeAt(text, cursor)
	_, width := utf8.DecodeLastRuneInString(text[:cursor])
	if cursor == 0 {
		return character, -1
	}
	return character, cursor - width
}

// skip is upstream's skip(characters): from startIndex, step while the character matches.
func skip(matches func(rune) bool) func(text string, startIndex int, backwards bool) int {
	return func(text string, startIndex int, backwards bool) int {
		if startIndex == notFound {
			return notFound
		}
		cursor := startIndex
		for cursor >= 0 && cursor < len(text) {
			if backwards {
				character, previous := runeBackwards(text, cursor)
				if !matches(character) {
					return cursor
				}
				cursor = previous
			} else {
				character, width := runeAt(text, cursor)
				if !matches(character) {
					return cursor
				}
				cursor += width
			}
		}
		if cursor == -1 || cursor == len(text) {
			return cursor
		}
		return notFound
	}
}

// isJavaScriptWhitespace is JavaScript's \s: WhiteSpace and LineTerminator in ECMA-262.
func isJavaScriptWhitespace(character rune) bool {
	switch character {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0x00A0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return character >= 0x2000 && character <= 0x200A
}

var (
	// SkipWhitespace is upstream's skipWhitespace, skip(/\s/).
	SkipWhitespace = skip(isJavaScriptWhitespace)
	// SkipSpaces is upstream's skipSpaces, skip(" \t").
	SkipSpaces = skip(func(character rune) bool { return character == ' ' || character == '\t' })
	// SkipToLineEnd is upstream's skipToLineEnd, skip(",; \t").
	SkipToLineEnd = skip(func(character rune) bool { return strings.ContainsRune(",; \t", character) })
	// SkipEverythingButNewLine is upstream's skipEverythingButNewLine, skip(/[^\n\r]/).
	SkipEverythingButNewLine = skip(func(character rune) bool { return character != '\n' && character != '\r' })
)

func isNewlineCharacter(character rune) bool {
	return character == '\n' || character == '\r' || character == 0x2028 || character == 0x2029
}

// SkipNewline is upstream's skipNewline: step over one newline, CRLF counting as one.
func SkipNewline(text string, startIndex int, backwards bool) int {
	if startIndex == notFound || startIndex < 0 || startIndex >= len(text) {
		// upstream's charAt out of range is "", which is no newline, so the index comes back unchanged.
		return startIndex
	}
	character, width := runeAt(text, startIndex)
	if backwards {
		if startIndex > 0 && text[startIndex-1] == '\r' && character == '\n' {
			// Upstream's startIndex - 2 is the character before the \r, which is two UTF-16 units back
			// only because \r and \n are one unit each. In bytes it is the start of whatever rune precedes
			// the \r, and the first version's startIndex - 2 landed inside a multi-byte one: the text
			// differential found it on "\u2028\r\n", where it split the U+2028.
			return previousRuneStart(text, startIndex-1)
		}
		if isNewlineCharacter(character) {
			_, previous := runeBackwards(text, startIndex)
			return previous
		}
		return startIndex
	}
	if character == '\r' && startIndex+1 < len(text) && text[startIndex+1] == '\n' {
		return startIndex + 2
	}
	if isNewlineCharacter(character) {
		return startIndex + width
	}
	return startIndex
}

// HasNewline is upstream's hasNewline: is there a newline right after startIndex, past spaces and tabs.
// Backwards looks before it instead.
func HasNewline(text string, startIndex int, backwards bool) bool {
	start := startIndex
	if backwards {
		start = previousRuneStart(text, startIndex)
	}
	index := SkipSpaces(text, start, backwards)
	return index != SkipNewline(text, index, backwards)
}

// previousRuneStart is upstream's `startIndex - 1` on a byte offset: the start of the rune before it.
func previousRuneStart(text string, startIndex int) int {
	if startIndex <= 0 {
		return startIndex - 1
	}
	if startIndex > len(text) {
		return startIndex - 1
	}
	_, width := utf8.DecodeLastRuneInString(text[:startIndex])
	return startIndex - width
}

// IsPreviousLineEmpty is upstream's isPreviousLineEmpty.
func IsPreviousLineEmpty(text string, startIndex int) bool {
	index := previousRuneStart(text, startIndex)
	index = SkipSpaces(text, index, true)
	index = SkipNewline(text, index, true)
	index = SkipSpaces(text, index, true)
	return index != SkipNewline(text, index, true)
}

// SkipInlineComment is upstream's skipInlineComment: past a /* */ comment starting at startIndex.
func SkipInlineComment(text string, startIndex int) int {
	if startIndex == notFound {
		return notFound
	}
	// upstream's charAt outside the text is "", which starts no comment, so the index comes back as is.
	if startIndex < 0 || startIndex >= len(text) {
		return startIndex
	}
	if strings.HasPrefix(text[startIndex:], "/*") {
		if end := strings.Index(text[startIndex+2:], "*/"); end >= 0 {
			return startIndex + 2 + end + 2
		}
	}
	return startIndex
}

// SkipTrailingComment is upstream's skipTrailingComment: to the end of a // comment at startIndex.
func SkipTrailingComment(text string, startIndex int) int {
	if startIndex == notFound {
		return notFound
	}
	// Not clamped to the text: an earlier version clamped -1 to 0 and so reported a // comment at the
	// start of a file for an index before it, where upstream's charAt(-1) is "" and returns -1.
	if startIndex < 0 || startIndex >= len(text) {
		return startIndex
	}
	if strings.HasPrefix(text[startIndex:], "//") {
		return SkipEverythingButNewLine(text, startIndex, false)
	}
	return startIndex
}
