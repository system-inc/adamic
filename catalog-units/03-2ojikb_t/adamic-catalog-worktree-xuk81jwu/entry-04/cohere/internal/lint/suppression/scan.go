package suppression

import "github.com/system-inc/cohere/internal/lint/ecmascript/directives"

// commentSpan is one comment's byte range in a file.
type commentSpan struct {
	pos int
	end int
}

// scanComments finds every comment in a file, skipping the ones that are not comments.
//
// The naive version of this — search the text for the directive — is wrong in a way that is
// invisible when it fires. This codebase contains, in two build scripts, the literal string
// `'/* eslint-disable */\n'`: source that *writes* a suppression into a generated file. A text
// search reads that as a file-level suppression and silences every rule for the rest of the script.
// Nothing would report that it happened.
//
// So the text is scanned as TypeScript lexes it: string literals, template literals, and regular
// expressions are skipped rather than searched, and only real comments are returned. The two build
// scripts, and the four prose sentences that mention `eslint-disable`, all fall out correctly —
// the strings because they are skipped here, the prose because a directive must begin its comment.
//
// This is a lexer for exactly the question "where are the comments", not a parser. It does not need
// to know what any token means, only where each one ends, which is what keeps it short enough to
// read in one sitting.
func scanComments(text string) []commentSpan {
	comments := []commentSpan{}
	scanCode(text, 0, false, &comments)
	return comments
}

// scanCode lexes code from index, appending every comment it passes, and returns where it stopped:
// the end of the text, or, inside an interpolation, just past the `}` that closes the `${`.
//
// One loop for top-level code and for an interpolation body, because the body of `${...}` is code: it
// holds strings, templates, regular expressions and comments exactly as the file does. An earlier
// interpolation skipper knew strings and nested templates and nothing else, so in
//
//	return `"${key.replace(/"/g, '""')}"`;
//
// it read the regex's `"` as a string, paired the wrong quotes, and left the scan inside a phantom
// template for the rest of the file. Base's SqliteAdapter.ts lost every directive after that line that
// way: its block enable, so a disable for one rule ran to the end of the file, and two next-line
// disables, so their findings were reported (#59rmvb5).
func scanCode(text string, index int, inInterpolation bool, comments *[]commentSpan) int {
	// previousToken is the last significant character seen, which is the only way to tell a regular
	// expression from a division: `/` after a value divides, `/` after an operator opens a regex. An
	// interpolation starts an expression, so it starts as the start of the file does.
	previousToken := byte(0)

	// depth counts the braces opened inside an interpolation, so `${ {a: 1}.a }` closes on its own
	// brace rather than the object literal's.
	depth := 0

	for index < len(text) {
		character := text[index]

		switch {
		case character == '/' && index+1 < len(text) && text[index+1] == '/':
			start := index
			index += 2
			for index < len(text) && text[index] != '\n' {
				index++
			}
			*comments = append(*comments, commentSpan{pos: start, end: index})

		case character == '/' && index+1 < len(text) && text[index+1] == '*':
			start := index
			index += 2
			for index+1 < len(text) && !(text[index] == '*' && text[index+1] == '/') {
				index++
			}
			if index+1 < len(text) {
				index += 2
			} else {
				index = len(text)
			}
			*comments = append(*comments, commentSpan{pos: start, end: index})

		case character == '"' || character == '\'':
			index = skipQuoted(text, index, character)
			previousToken = character

		case character == '`':
			index = skipTemplate(text, index, comments)
			previousToken = character

		case character == '/' && opensRegularExpression(previousToken):
			index = skipRegularExpression(text, index)
			previousToken = '/'

		case character == '}' && inInterpolation && depth == 0:
			return index + 1

		default:
			switch character {
			case '{':
				depth++
			case '}':
				depth--
			}
			if !isSpace(character) {
				previousToken = character
			}
			index++
		}
	}

	return index
}

// enableSpan is one enable directive, resolved to the line it sits on and the rules it names.
type enableSpan struct {
	line  int
	rules []string
}

// scanEnables finds the comments that close a block suppression.
//
// It re-scans rather than sharing Build's pass because only a file that actually contains a
// file-scope directive needs enables resolved at all, and that is 4 files in the corpus. Paying a
// second scan on those four is cheaper than threading enable state through the common path.
func scanEnables(text string, lineOf func(offset int) int) []enableSpan {
	found := []enableSpan{}

	for _, comment := range scanComments(text) {
		names, isEnable := directives.ParseEnable(text[comment.pos:comment.end])
		if !isEnable {
			continue
		}
		found = append(found, enableSpan{
			line:  lineOf(comment.pos),
			rules: names,
		})
	}

	return found
}

// buildLineIndex returns a function from byte offset to zero-based line.
//
// Built once per file rather than counting newlines per lookup, which would be quadratic on a file
// with many directives.
func buildLineIndex(text string) func(offset int) int {
	starts := []int{0}
	for index := 0; index < len(text); index++ {
		if text[index] == '\n' {
			starts = append(starts, index+1)
		}
	}
	return func(offset int) int {
		low, high := 0, len(starts)-1
		for low < high {
			middle := (low + high + 1) / 2
			if starts[middle] <= offset {
				low = middle
			} else {
				high = middle - 1
			}
		}
		return low
	}
}

// skipQuoted advances past a single- or double-quoted string, honoring backslash escapes.
func skipQuoted(text string, index int, quote byte) int {
	index++
	for index < len(text) {
		switch text[index] {
		case '\\':
			index += 2
			continue
		case quote:
			return index + 1
		case '\n':
			// An unterminated string literal. The file will not parse, but this scan must still
			// terminate rather than swallowing the rest of the file as a string.
			return index + 1
		}
		index++
	}
	return index
}

// skipTemplate advances past a template literal, lexing each `${...}` as code, because a comment,
// a string or a regular expression can live inside an interpolation.
func skipTemplate(text string, index int, comments *[]commentSpan) int {
	index++
	for index < len(text) {
		switch {
		case text[index] == '\\':
			index += 2
			continue
		case text[index] == '`':
			return index + 1
		case text[index] == '$' && index+1 < len(text) && text[index+1] == '{':
			index = scanCode(text, index+2, true, comments)
			continue
		}
		index++
	}
	return index
}

// skipRegularExpression advances past a regex literal, honoring escapes and character classes.
//
// A `/` inside `[...]` does not end the pattern, which is why the class has to be tracked.
func skipRegularExpression(text string, index int) int {
	index++
	inClass := false
	for index < len(text) {
		switch text[index] {
		case '\\':
			index += 2
			continue
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '/':
			if !inClass {
				return index + 1
			}
		case '\n':
			// Not a regex after all — a lone division that reached the end of its line. Back out
			// rather than consuming the rest of the file.
			return index
		}
		index++
	}
	return index
}

// opensRegularExpression decides whether a `/` starts a pattern or divides.
//
// After a value — an identifier, a literal, a closing bracket — a slash divides. After an operator,
// a comma, or an opening bracket, it opens a pattern. This is the standard heuristic and it is
// wrong in the same rare places every JavaScript lexer is wrong; getting it wrong here costs a
// misread comment boundary in a file that contains both a regex and a suppression, which the
// fixtures cover.
func opensRegularExpression(previousToken byte) bool {
	switch previousToken {
	case 0, '(', ',', '=', ':', '[', '!', '&', '|', '?', '{', '}', ';', '+', '-', '*', '~', '^', '<', '>', '%':
		return true
	}
	return false
}

func isSpace(character byte) bool {
	return character == ' ' || character == '\t' || character == '\n' || character == '\r'
}
