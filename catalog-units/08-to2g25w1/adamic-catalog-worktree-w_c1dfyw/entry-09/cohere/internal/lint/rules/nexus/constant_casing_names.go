package nexus

import (
	"strings"
	"unicode"
)

// The casing predicates and converters for consistency-require-constant-casing.
//
// These are deliberately not the toPascalCase and toCamelCase already in this package. Those serve
// consistency-no-screaming-snake-case, where every word arrives all-uppercase and lowercasing a part
// whole is exactly right. Here the input is an ordinary identifier whose internal boundaries are
// load bearing: folding numberAbsentPlaceholder the shouting way yields numberabsentplaceholder.
// Two converters that do almost the same thing is worse than one of the wrong shape, so the
// difference is stated rather than shared.

// isPascalCase and isCamelCase each allow a leading underscore.
//
// A leading underscore marks a binding that is declared and knowingly unused, which every linter in
// this stack reads that way, so it is orthogonal to casing rather than a violation of it. Rejecting
// it made "_results" report as not-camelCase while the converter handed back "_results", an error
// with no edit that satisfies it. The underscore is skipped and the rest is judged.
func isPascalCase(name string) bool {
	return matchesCasing(name, unicode.IsUpper)
}

func isCamelCase(name string) bool {
	return matchesCasing(name, unicode.IsLower)
}

// matchesCasing checks the shared shape: leading underscores, then a first letter the caller judges,
// then letters and digits only. An underscore after the first letter disqualifies the name, which is
// what sends snake_case to a converter rather than letting it pass as either casing.
func matchesCasing(name string, firstLetterMatches func(rune) bool) bool {
	body := strings.TrimLeft(name, "_")
	if body == "" {
		return false
	}

	for index, character := range body {
		if index == 0 {
			if !unicode.IsLetter(character) || !firstLetterMatches(character) {
				return false
			}
			continue
		}
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) {
			return false
		}
	}
	return true
}

// constantNameWords folds the separators a name carries into word boundaries, so a suggestion is a
// name the predicates above would actually accept.
//
// Both predicates reject an underscore anywhere in the body, so a converter that did not remove one
// would report a name and then hand back its own spelling: `Constant "user_id" should be camelCase
// ("user_id")`. That is an error no edit can satisfy, and a reader who tries the suggestion gets the
// same report again.
//
// The leading underscores are kept and returned separately, because they carry the unused-binding
// claim rather than a casing one, and that is not this rule's to overrule.
func constantNameWords(name string) (leadingUnderscores string, words []string) {
	body := strings.TrimLeft(name, "_")
	leadingUnderscores = name[:len(name)-len(body)]

	for _, part := range strings.Split(body, "_") {
		if part == "" {
			continue
		}
		// An all-uppercase part has no internal boundaries to preserve, so lowercase it whole:
		// leaving it produces "utmSOURCE" from "utm_SOURCE", which is neither casing.
		if isAllUppercase(part) {
			part = strings.ToLower(part)
		}
		words = append(words, part)
	}
	return leadingUnderscores, words
}

// isAllUppercase reports whether a word carries no lowercase letter, so it has no boundary to keep.
func isAllUppercase(part string) bool {
	for _, character := range part {
		if unicode.IsLower(character) {
			return false
		}
	}
	return true
}

// constantNameToPascalCase raises the first character of each word, so numberAbsentPlaceholder
// suggests NumberAbsentPlaceholder and apple_root_ca suggests AppleRootCa.
func constantNameToPascalCase(name string) string {
	leadingUnderscores, words := constantNameWords(name)
	if len(words) == 0 {
		return name
	}

	var builder strings.Builder
	builder.WriteString(leadingUnderscores)
	for _, word := range words {
		builder.WriteString(raiseFirstCharacter(word))
	}
	return builder.String()
}

// constantNameToCamelCase is the mirror: the first word starts lowercase, every later one is raised.
func constantNameToCamelCase(name string) string {
	leadingUnderscores, words := constantNameWords(name)
	if len(words) == 0 {
		return name
	}

	var builder strings.Builder
	builder.WriteString(leadingUnderscores)
	builder.WriteString(lowerFirstCharacter(words[0]))
	for _, word := range words[1:] {
		builder.WriteString(raiseFirstCharacter(word))
	}
	return builder.String()
}

func raiseFirstCharacter(word string) string {
	runes := []rune(word)
	if len(runes) == 0 {
		return ""
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// lowerFirstCharacter lowers only the first character, since the rest of the word carries boundaries
// the caller wants kept. constantNameWords has already lowered an all-uppercase word whole, so this
// never produces the "tHRESHOLDS" shape.
func lowerFirstCharacter(word string) string {
	runes := []rune(word)
	if len(runes) == 0 {
		return ""
	}
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// reservedWords are the words a program cannot use as a binding name.
//
// A trailing underscore is the usual escape when a keyword is the name the domain wants, and
// ChatUsageReportService used exactly that for a Durable Object handle: "do_". Folding the separator
// away produced "do", which parses as the loop keyword and took the file with it. A suggestion that
// cannot be typed is worse than no suggestion, so a name whose escape is load bearing is left alone.
var reservedWords = map[string]bool{
	"arguments": true, "await": true, "break": true, "case": true, "catch": true, "class": true,
	"const": true, "continue": true, "debugger": true, "default": true, "delete": true, "do": true,
	"else": true, "enum": true, "eval": true, "export": true, "extends": true, "false": true,
	"finally": true, "for": true, "function": true, "if": true, "implements": true, "import": true,
	"in": true, "instanceof": true, "interface": true, "let": true, "new": true, "null": true,
	"package": true, "private": true, "protected": true, "public": true, "return": true,
	"static": true, "super": true, "switch": true, "this": true, "throw": true, "true": true,
	"try": true, "typeof": true, "var": true, "void": true, "while": true, "with": true,
	"yield": true,
}

// escapesAReservedWord reports a name that only escapes a keyword, so recasing it would produce
// something unusable.
func escapesAReservedWord(name string) bool {
	return reservedWords[strings.ToLower(strings.TrimRight(name, "_"))]
}
