package text

import "strings"

// IsWhitespace reports whether a character is whitespace to JavaScript: what `\s` matches in a
// regular expression and what `String.prototype.trim` removes, which is ECMAScript's WhiteSpace plus
// its LineTerminator, 25 code points in all.
//
// Go's `unicode.IsSpace` is a different set, and a port of a JavaScript tool that trims with it
// answers differently from the tool it ports. The two disagree on two characters: Go counts the
// next-line character (U+0085) as space and JavaScript does not, and JavaScript counts the byte order
// mark (U+FEFF) as space and Go does not. @system_adamic's stream P2 found it porting cohere's
// suppression index: 6 of 8 directive comments holding one of the two got a different answer from
// cohere than from ESLint 10.12.0 (#z4nssqs).
//
// The set is spelled out rather than read from Go's Unicode tables, so it holds whatever Unicode
// version Go ships. TestIsWhitespaceMatchesJavaScript pins it against every code point, from the list
// Node 24 (Unicode 17) printed for both `\s` and `trim`.
func IsWhitespace(character rune) bool {
	switch character {
	case
		// WhiteSpace: tab, vertical tab, form feed, space, no-break space, zero width no-break space.
		'\t', '\v', '\f', ' ', '\u00a0', '\ufeff',
		// The rest of WhiteSpace, the space separators (Zs).
		'\u1680', '\u2000', '\u2001', '\u2002', '\u2003', '\u2004', '\u2005', '\u2006', '\u2007',
		'\u2008', '\u2009', '\u200a', '\u202f', '\u205f', '\u3000',
		// LineTerminator: line feed, carriage return, line separator, paragraph separator.
		'\n', '\r', '\u2028', '\u2029':
		return true
	}
	return false
}

// TrimWhitespace is `String.prototype.trim`: the string without the JavaScript whitespace at either
// end.
func TrimWhitespace(value string) string {
	return strings.TrimFunc(value, IsWhitespace)
}

// WhitespaceFields splits a string around each run of JavaScript whitespace, dropping empty fields,
// as `value.split(/\s+/).filter(Boolean)` does. It is `strings.Fields` over JavaScript's set.
func WhitespaceFields(value string) []string {
	return strings.FieldsFunc(value, IsWhitespace)
}
