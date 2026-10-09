package text

import "strings"

// UnescapeStringLiteralText is typescript-estree's unescapeStringLiteralText, which ESLint's
// typescript-eslint parser applies to JSX text and to a JSX attribute's string value: numeric
// character references and the XHTML named entities are decoded, and anything else is left as
// written, including a named entity it does not know and a code point past U+10FFFF.
//
// Shared by the formatter, which prints the decoded value as Prettier does, and by the lint rules
// that read an attribute's string, which must see what ESLint's rules see.
func UnescapeStringLiteralText(text string) string {
	if !strings.Contains(text, "&") {
		return text
	}
	var builder strings.Builder
	for index := 0; index < len(text); {
		if text[index] != '&' {
			builder.WriteByte(text[index])
			index++
			continue
		}
		semicolon := strings.IndexByte(text[index:], ';')
		if semicolon < 2 {
			builder.WriteByte('&')
			index++
			continue
		}
		entity := text[index : index+semicolon+1]
		if replacement, ok := decodeEntity(entity[1 : len(entity)-1]); ok {
			builder.WriteString(replacement)
			index += len(entity)
			continue
		}
		builder.WriteByte('&')
		index++
	}
	return builder.String()
}

// decodeEntity decodes one entity body, `#123`, `#x1F` or a name, matching upstream's pattern
// /&(?:#\d+|#x[\da-fA-F]+|[0-9a-zA-Z]+);/.
func decodeEntity(item string) (string, bool) {
	if item[0] == '#' {
		if len(item) > 2 && item[1] == 'x' {
			value := 0
			for _, character := range item[2:] {
				digit := hexValue(character)
				if digit < 0 {
					return "", false
				}
				value = value*16 + digit
				if value > 0x10FFFF {
					value = 0x110000
				}
			}
			if value > 0x10FFFF {
				return "&" + item + ";", true
			}
			return string(rune(value)), true
		}
		if len(item) < 2 {
			return "", false
		}
		value := 0
		for _, character := range item[1:] {
			if character < '0' || character > '9' {
				return "", false
			}
			value = value*10 + int(character-'0')
			if value > 0x10FFFF {
				value = 0x110000
			}
		}
		if value > 0x10FFFF {
			return "&" + item + ";", true
		}
		return string(rune(value)), true
	}
	for _, character := range item {
		if !(character >= '0' && character <= '9' || character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z') {
			return "", false
		}
	}
	if replacement, known := xhtmlEntities[item]; known {
		return replacement, true
	}
	return "&" + item + ";", true
}

func hexValue(character rune) int {
	switch {
	case character >= '0' && character <= '9':
		return int(character - '0')
	case character >= 'a' && character <= 'f':
		return int(character-'a') + 10
	case character >= 'A' && character <= 'F':
		return int(character-'A') + 10
	}
	return -1
}
