package graphql

import "math"

// graphql-js 17.0.2, language/blockString.js: dedentBlockStringLines, which the lexer calls for every
// block string. isPrintableAsBlockString and printBlockString are graphql-js's printer's, and neither
// the lexer nor Prettier calls them, so they are not here.

// dedentBlockStringLines produces the value of a block string from its parsed raw value, similar to
// CoffeeScript's block string, Python's docstring trim or Ruby's strip_heredoc.
//
// This implements the GraphQL spec's BlockStringValue() static algorithm.
//
// The lines are byte strings where graphql-js's are UTF-16, and the arithmetic still agrees: an indent
// counts tabs and spaces, one byte and one code unit each, so every slice below lands after ASCII.
func dedentBlockStringLines(lines []string) []string {
	commonIndent := math.MaxInt // Number.MAX_SAFE_INTEGER
	firstNonEmptyLine := -1     // null
	lastNonEmptyLine := -1

	for i, line := range lines {
		indent := leadingWhitespace(line)

		if indent == len(line) {
			continue // skip empty lines
		}

		if firstNonEmptyLine == -1 {
			firstNonEmptyLine = i
		}
		lastNonEmptyLine = i

		if i != 0 && indent < commonIndent {
			commonIndent = indent
		}
	}

	// Remove common indentation from all lines but first. JavaScript's slice clamps an index past the
	// end, so a blank line shorter than the indent (or an indent that is still MAX_SAFE_INTEGER) gives "".
	dedented := make([]string, len(lines))
	for i, line := range lines {
		if i == 0 {
			dedented[i] = line
		} else {
			dedented[i] = line[min(commonIndent, len(line)):]
		}
	}
	// Remove leading and trailing blank lines.
	if firstNonEmptyLine == -1 {
		firstNonEmptyLine = 0
	}
	return dedented[firstNonEmptyLine : lastNonEmptyLine+1]
}

func leadingWhitespace(str string) int {
	i := 0
	for i < len(str) && isWhiteSpace(int(str[i])) {
		i++
	}
	return i
}
