package regexpattern

import (
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
)

// singleEscapeValues are the characters a named escape stands for.
//
// The backspace escape only means backspace inside a character class; outside one the same two
// characters are a word boundary, which matches a position and never reaches here. That asymmetry
// lives in readEscape rather than in this table, so the table stays a plain statement of what each
// name denotes.
var singleEscapeValues = map[byte]uint32{
	'n': 0x0A,
	'r': 0x0D,
	't': 0x09,
	'v': 0x0B,
	'f': 0x0C,
	'b': 0x08,
}

// escapeValue returns the code point an escape denotes, or false when the escape has no single one.
//
// The kind is passed in rather than re-derived so the two cannot disagree. A value computed against
// one reading of the spelling while the caller holds another is the kind of divergence that
// produces a correct-looking answer about the wrong character.
func escapeValue(text string, kind CharacterKind) (uint32, bool) {
	if len(text) < 2 || text[0] != '\\' {
		return 0, false
	}

	switch kind {
	case KindSingleEscape:
		value, ok := singleEscapeValues[text[1]]
		return value, ok

	case KindNull:
		return 0, true

	case KindHexadecimalEscape:
		// Exactly two hex digits. Fewer is not a hex escape at all, and the scanner would not have
		// spelled it this way, so a short one is refused rather than padded.
		if len(text) != 4 || !regexsyntax.AllHexDigits(text[2:]) {
			return 0, false
		}
		return regexsyntax.ParseHexUint(text[2:]), true

	case KindUnicodeEscape:
		return unicodeEscapeValue(text)

	case KindControlLetter:
		// The control escape maps a letter onto the control code in its low five bits, so the
		// letter's case does not matter and the arithmetic is the definition rather than a trick.
		if len(text) != 3 {
			return 0, false
		}
		letter := text[2]
		if !isAsciiLetter(letter) {
			return 0, false
		}
		return uint32(letter % 32), true

	case KindOctal:
		return octalEscapeValue(text[1:])

	case KindIdentityEscape:
		// A backslash before a character that needs no escaping denotes that character. This is
		// what makes an escaped space a space, which matters to a caller counting spaces: it is
		// still one, and it is still written down.
		value, width := decodeRune(text, 1)
		if width == 0 {
			return 0, false
		}
		return value, true
	}

	return 0, false
}

// unicodeEscapeValue reads the fixed-width and braced forms of a unicode escape.
//
// The braced form is only valid under the u or v flag, and that check belongs to the scanner that
// decided the spelling rather than here: by the time a kind says unicode escape, the flags have
// already been consulted. Re-checking them at this layer would put the same decision in two places.
func unicodeEscapeValue(text string) (uint32, bool) {
	if len(text) < 3 {
		return 0, false
	}

	if text[2] == '{' {
		if text[len(text)-1] != '}' {
			return 0, false
		}
		digits := text[3 : len(text)-1]
		if digits == "" || !regexsyntax.AllHexDigits(digits) {
			return 0, false
		}
		return regexsyntax.ParseHexUint(digits), true
	}

	if len(text) != 6 || !regexsyntax.AllHexDigits(text[2:]) {
		return 0, false
	}
	return regexsyntax.ParseHexUint(text[2:]), true
}

// octalEscapeValue reads a legacy octal escape's digits.
//
// A leading digit of 8 or 9 is not octal at all, which is what separates this from the legacy
// decimal escapes another rule reports. Values above 255 are refused rather than truncated, since
// the escape does not denote a character in that case.
func octalEscapeValue(digits string) (uint32, bool) {
	if digits == "" {
		return 0, false
	}
	var value uint32
	for index := 0; index < len(digits); index++ {
		digit := digits[index]
		if digit < '0' || digit > '7' {
			return 0, false
		}
		value = value*8 + uint32(digit-'0')
	}
	if value > 0xFF {
		return 0, false
	}
	return value, true
}

// extendOctalEscape widens a two-byte digit escape to cover the octal digits that follow it.
//
// The language allows up to three octal digits, and stops early at a digit outside the octal range
// or at a value above 255: `\400` is `\40` followed by a literal zero, because 0400 does not name a
// character. Reproducing that rather than taking three digits blindly is what keeps the reported
// span equal to the text the escape actually covers.
//
// A backreference is left at one digit deliberately. `\1` in a pattern with capturing groups is a
// reference rather than a character, and the caller that cares decides which by counting groups; a
// span greedily widened here would take digits belonging to the text after the reference.
func extendOctalEscape(pattern string, start int, end int) int {
	if end-start != 2 || start+1 >= len(pattern) {
		return end
	}
	first := pattern[start+1]
	if first < '0' || first > '7' {
		return end
	}

	value := uint32(first - '0')
	scan := start + 2
	// Two bounds, and only one of them is reachable on its own: the 255 ceiling fires first for
	// every input the three-digit cap would catch, because a fourth octal digit always exceeds it.
	// A sweep confirmed the cap can be widened to nine with no fixture noticing. It stays because it
	// is the language's rule stated directly, and leaving the walk to depend on an arithmetic
	// coincidence would be a worse thing to read than one redundant condition.
	for digits := 1; digits < 3 && scan < len(pattern); digits++ {
		next := pattern[scan]
		if next < '0' || next > '7' {
			break
		}
		if value*8+uint32(next-'0') > 0xFF {
			break
		}
		value = value*8 + uint32(next-'0')
		scan++
	}
	return scan
}

// braceQuantifierEnd returns the index past a counted quantifier, or false when the brace does not
// open one.
//
// A brace that is not a quantifier is a literal brace, which is legal in a pattern, so refusing is
// the common path rather than an error path. The accepted shapes are a count, a count with a
// trailing comma, and a range.
func braceQuantifierEnd(pattern string, start int) (int, bool) {
	if start >= len(pattern) || pattern[start] != '{' {
		return 0, false
	}

	index := start + 1
	digitsBefore := consumeDigits(pattern, index)
	if digitsBefore == index {
		return 0, false
	}
	index = digitsBefore

	if index < len(pattern) && pattern[index] == ',' {
		index++
		index = consumeDigits(pattern, index)
	}

	if index >= len(pattern) || pattern[index] != '}' {
		return 0, false
	}
	return index + 1, true
}

// consumeDigits returns the index past a run of decimal digits, which may be empty.
func consumeDigits(pattern string, index int) int {
	for index < len(pattern) && pattern[index] >= '0' && pattern[index] <= '9' {
		index++
	}
	return index
}

// isAsciiLetter reports whether a byte is an unaccented letter, which is what a control escape
// accepts.
func isAsciiLetter(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z')
}
