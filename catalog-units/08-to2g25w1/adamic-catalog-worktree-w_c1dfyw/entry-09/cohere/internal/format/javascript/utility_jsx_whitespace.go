package javascript

import (
	"regexp"
	"strings"
)

// utilities/jsx-whitespace.js with the class it instantiates, src/utilities/whitespace-utilities.js,
// and utilities/is-meaningful-jsx-text.js.

// whitespaceUtilities is upstream's WhitespaceUtilities. Every whitespace set it is built with (JSX's
// " \n\r\t") is ASCII, so walking bytes visits the same characters upstream's charAt walk does, and
// the counts it returns are byte counts that slice the same text.
type whitespaceUtilities struct {
	whitespaceCharacters map[byte]bool
	pattern              *regexp.Regexp
	capturingPattern     *regexp.Regexp
}

func newWhitespaceUtilities(whitespaceCharacters string) *whitespaceUtilities {
	set := map[byte]bool{}
	class := ""
	for index := 0; index < len(whitespaceCharacters); index++ {
		set[whitespaceCharacters[index]] = true
		class += regexp.QuoteMeta(whitespaceCharacters[index : index+1])
	}
	return &whitespaceUtilities{
		whitespaceCharacters: set,
		pattern:              regexp.MustCompile("[" + class + "]+"),
		capturingPattern:     regexp.MustCompile("([" + class + "]+)"),
	}
}

func (utilities *whitespaceUtilities) getLeadingWhitespaceCount(text string) int {
	count := 0

	for index := 0; index < len(text) && utilities.whitespaceCharacters[text[index]]; index++ {
		count++
	}

	return count
}

func (utilities *whitespaceUtilities) getTrailingWhitespaceCount(text string) int {
	count := 0

	for index := len(text) - 1; index >= 0 && utilities.whitespaceCharacters[text[index]]; index-- {
		count++
	}

	return count
}

func (utilities *whitespaceUtilities) getLeadingWhitespace(text string) string {
	count := utilities.getLeadingWhitespaceCount(text)
	return text[:count]
}

func (utilities *whitespaceUtilities) getTrailingWhitespace(text string) string {
	count := utilities.getTrailingWhitespaceCount(text)
	return text[len(text)-count:]
}

func (utilities *whitespaceUtilities) hasLeadingWhitespace(text string) bool {
	return len(text) > 0 && utilities.whitespaceCharacters[text[0]]
}

func (utilities *whitespaceUtilities) hasTrailingWhitespace(text string) bool {
	return len(text) > 0 && utilities.whitespaceCharacters[text[len(text)-1]]
}

func (utilities *whitespaceUtilities) trimStart(text string) string {
	count := utilities.getLeadingWhitespaceCount(text)
	return text[count:]
}

func (utilities *whitespaceUtilities) trimEnd(text string) string {
	count := utilities.getTrailingWhitespaceCount(text)
	return text[:len(text)-count]
}

func (utilities *whitespaceUtilities) trim(text string) string {
	return utilities.trimEnd(utilities.trimStart(text))
}

// split is upstream's split: String.prototype.split on a run of whitespace, keeping the runs between
// the words when captureWhitespace is set, as a capturing group does in JavaScript.
func (utilities *whitespaceUtilities) split(text string, captureWhitespace bool) []string {
	pattern := utilities.pattern
	if captureWhitespace {
		pattern = utilities.capturingPattern
	}
	var parts []string
	last := 0
	for _, match := range pattern.FindAllStringIndex(text, -1) {
		parts = append(parts, text[last:match[0]])
		if captureWhitespace {
			parts = append(parts, text[match[0]:match[1]])
		}
		last = match[1]
	}
	return append(parts, text[last:])
}

func (utilities *whitespaceUtilities) hasWhitespaceCharacter(text string) bool {
	for index := 0; index < len(text); index++ {
		if utilities.whitespaceCharacters[text[index]] {
			return true
		}
	}
	return false
}

func (utilities *whitespaceUtilities) hasNonWhitespaceCharacter(text string) bool {
	for index := 0; index < len(text); index++ {
		if !utilities.whitespaceCharacters[text[index]] {
			return true
		}
	}
	return false
}

func (utilities *whitespaceUtilities) isWhitespaceOnly(text string) bool {
	for index := 0; index < len(text); index++ {
		if !utilities.whitespaceCharacters[text[index]] {
			return false
		}
	}
	return true
}

func (utilities *whitespaceUtilities) getMinIndentation(text string) int {
	minIndentation := -1

	for _, line := range strings.Split(text, "\n") {
		if len(line) == 0 {
			continue
		}

		indentation := utilities.getLeadingWhitespaceCount(line)
		if indentation == 0 {
			return 0
		}

		if len(line) == indentation {
			continue
		}

		if minIndentation == -1 || indentation < minIndentation {
			minIndentation = indentation
		}
	}

	if minIndentation == -1 {
		return 0
	}
	return minIndentation
}

func (utilities *whitespaceUtilities) dedentString(text string) string {
	minIndent := utilities.getMinIndentation(text)
	if minIndent == 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	for index, lineText := range lines {
		if len(lineText) < minIndent {
			lines[index] = ""
		} else {
			lines[index] = lineText[minIndent:]
		}
	}
	return strings.Join(lines, "\n")
}

// jsxWhitespace is upstream's jsxWhitespace.
//
// Only the following are treated as whitespace inside JSX.
//
//   - U+0020 SPACE
//   - U+000A LF
//   - U+000D CR
//   - U+0009 TAB
var jsxWhitespace = newWhitespaceUtilities(" \n\r\t")

// isMeaningfulJsxText is upstream's isMeaningfulJsxText.
//
// Meaningful if it contains non-whitespace characters,
// or it contains whitespace without a new line.
func isMeaningfulJsxText(node Node) bool {
	return node.Is("JSXText") &&
		(jsxWhitespace.hasNonWhitespaceCharacter(getRaw(node)) ||
			!strings.Contains(getRaw(node), "\n"))
}
