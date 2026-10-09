package micromark

import "strings"

// DecodeNamedCharacterReference is decode-named-character-reference: the decoded value, or false.
func DecodeNamedCharacterReference(value string) (string, bool) {
	decoded, present := characterEntities[value]
	return decoded, present
}

// DecodeNumericCharacterReference is micromark-util-decode-numeric-character-reference.
func DecodeNumericCharacterReference(value string, base int) string {
	return decodeNumericCharacterReference(value, base)
}

// DecodeString is micromark-util-decode-string: decode character escapes and references in a string,
// the replace of
//
//	/\\([!-/:-@[-`{-~])|&(#(?:\d{1,7}|x[\da-f]{1,6})|[\da-z]{1,31});/gi
//
// written as a scan. The regex matches a reference only when the whole run of digits or letters fits
// its bound and is followed by `;` (backtracking to a shorter run can never find a `;`), and an
// unknown name decodes to the match itself.
func DecodeString(value string) string {
	var builder strings.Builder
	index := 0
	for index < len(value) {
		character := value[index]

		if character == '\\' && index+1 < len(value) && asciiPunctuation(Code(value[index+1])) {
			builder.WriteByte(value[index+1])
			index += 2
			continue
		}

		if character == '&' {
			if end, decoded, matched := matchReference(value, index); matched {
				builder.WriteString(decoded)
				index = end
				continue
			}
		}

		builder.WriteByte(character)
		index++
	}
	return builder.String()
}

// matchReference tries the reference alternative at `&`: returns the index after `;` and the decoding.
func matchReference(value string, ampersand int) (int, string, bool) {
	start := ampersand + 1
	if start < len(value) && value[start] == '#' {
		cursor := start + 1
		hexadecimal := cursor < len(value) && (value[cursor] == 'x' || value[cursor] == 'X')
		if hexadecimal {
			cursor++
		}
		digitsStart := cursor
		for cursor < len(value) && isReferenceDigit(value[cursor], hexadecimal) {
			cursor++
		}
		digits := cursor - digitsStart
		limit := characterReferenceDecimalSizeMax
		base := 10
		if hexadecimal {
			limit = characterReferenceHexadecimalSizeMax
			base = 16
		}
		if digits >= 1 && digits <= limit && cursor < len(value) && value[cursor] == ';' {
			return cursor + 1, decodeNumericCharacterReference(value[digitsStart:cursor], base), true
		}
		// `#` is not alphanumeric, so the named alternative cannot match here either.
		return 0, "", false
	}

	cursor := start
	for cursor < len(value) && asciiAlphanumeric(Code(value[cursor])) {
		cursor++
	}
	letters := cursor - start
	if letters >= 1 && letters <= characterReferenceNamedSizeMax && cursor < len(value) && value[cursor] == ';' {
		if decoded, present := DecodeNamedCharacterReference(value[start:cursor]); present {
			return cursor + 1, decoded, true
		}
		// An unknown name: upstream returns $0, the match itself, unchanged.
		return cursor + 1, value[ampersand : cursor+1], true
	}
	return 0, "", false
}

func isReferenceDigit(character byte, hexadecimal bool) bool {
	if hexadecimal {
		return asciiHexDigit(Code(character))
	}
	return asciiDigit(Code(character))
}
