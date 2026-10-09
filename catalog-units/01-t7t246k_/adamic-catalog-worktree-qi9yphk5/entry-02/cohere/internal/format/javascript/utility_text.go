package javascript

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/printing"
)

// The generic text utilities in src/utilities that the JavaScript printer calls, over the core's skip
// functions (internal/format/printing/text.go), which already port skip, skipNewline, hasNewline,
// skipInlineComment and skipTrailingComment. Upstream's `false` position is the core's notFound, -2.

const notFound = -2

var (
	skipSpaces     = printing.SkipSpaces
	skipWhitespace = printing.SkipWhitespace
	skipToLineEnd  = printing.SkipToLineEnd
)

// hasNewline is upstream's hasNewline(text, index).
func hasNewline(text string, index int) bool { return printing.HasNewline(text, index, false) }

// hasNewlineBackwards is upstream's hasNewline(text, index, { backwards: true }).
func hasNewlineBackwards(text string, index int) bool { return printing.HasNewline(text, index, true) }

// hasNewlineInRange is upstream's hasNewlineInRange: a "\n" in [start, end).
func hasNewlineInRange(text string, start int, end int) bool {
	if start < 0 {
		start = 0
	}
	if end > len(text) {
		end = len(text)
	}
	return start < end && strings.IndexByte(text[start:end], '\n') >= 0
}

// isNextLineEmpty is upstream's isNextLineEmpty(text, startIndex), utilities/is-next-line-empty.js.
func isNextLineEmpty(text string, startIndex int) bool {
	oldIndex := notFound - 1
	index := startIndex
	for index != oldIndex {
		oldIndex = index
		index = skipToLineEnd(text, index, false)
		index = printing.SkipInlineComment(text, index)
		index = skipSpaces(text, index, false)
	}
	index = printing.SkipTrailingComment(text, index)
	index = printing.SkipNewline(text, index, false)
	return index != notFound && hasNewline(text, index)
}

// isPreviousLineEmpty is upstream's isPreviousLineEmpty.
func isPreviousLineEmpty(text string, startIndex int) bool {
	return printing.IsPreviousLineEmpty(text, startIndex)
}

// getNextNonSpaceNonCommentCharacterIndex is upstream's function of that name.
func getNextNonSpaceNonCommentCharacterIndex(text string, startIndex int) int {
	oldIndex := notFound - 1
	nextIndex := startIndex
	for nextIndex != oldIndex {
		oldIndex = nextIndex
		nextIndex = skipSpaces(text, nextIndex, false)
		nextIndex = printing.SkipInlineComment(text, nextIndex)
		nextIndex = printing.SkipTrailingComment(text, nextIndex)
		nextIndex = printing.SkipNewline(text, nextIndex, false)
	}
	return nextIndex
}

// getNextNonSpaceNonCommentCharacter is upstream's function of that name: the character, or "".
func getNextNonSpaceNonCommentCharacter(text string, startIndex int) string {
	index := getNextNonSpaceNonCommentCharacterIndex(text, startIndex)
	if index < 0 || index >= len(text) {
		return ""
	}
	return charAt(text, index)
}

// charAt is JavaScript's text.charAt(index) on a byte offset: the whole character starting there.
func charAt(text string, index int) string {
	if index < 0 || index >= len(text) {
		return ""
	}
	for end := index + 1; end <= len(text); end++ {
		if end == len(text) || text[end]&0xC0 != 0x80 {
			return text[index:end]
		}
	}
	return text[index:]
}

const (
	singleQuote = "'"
	doubleQuote = `"`
)

// getPreferredQuote is upstream's getPreferredQuote: the preferred quote unless it occurs more often
// than the alternate.
func getPreferredQuote(text string, preferSingleQuote bool) string {
	preferred, alternate := doubleQuote, singleQuote
	if preferSingleQuote {
		preferred, alternate = singleQuote, doubleQuote
	}
	preferredCount := strings.Count(text, preferred)
	alternateCount := strings.Count(text, alternate)
	if preferredCount > alternateCount {
		return alternate
	}
	return preferred
}

// makeStringPattern is upstream's /\\(["'\\])|(["'])/g.
var makeStringPattern = regexp.MustCompile(`\\(["'\\])|(["'])`)

// makeString is upstream's makeString: re-quote raw text, escaping and unescaping quotes as needed.
func makeString(rawText string, enclosingQuote string) string {
	otherQuote := doubleQuote
	if enclosingQuote == doubleQuote {
		otherQuote = singleQuote
	}
	raw := makeStringPattern.ReplaceAllStringFunc(rawText, func(match string) string {
		if len(match) == 2 {
			escaped := match[1:]
			if escaped == otherQuote {
				return otherQuote
			}
			return match
		}
		if match == enclosingQuote {
			return `\` + match
		}
		return match
	})
	return enclosingQuote + raw + enclosingQuote
}

// printString is upstream's printString, utilities/print-string.js. The JSON parsers always print
// double quotes; json5 and the HTML-attribute case are parsers and embeds we never run.
func printString(raw string, options *Options) string {
	rawContent := raw[1 : len(raw)-1]
	var enclosingQuote string
	switch settingsOf(options).Parser {
	case "json", "jsonc", "json-stringify":
		enclosingQuote = `"`
	default:
		enclosingQuote = getPreferredQuote(rawContent, settingsOf(options).SingleQuote)
	}
	if raw[:1] == enclosingQuote {
		return raw
	}
	return makeString(rawContent, enclosingQuote)
}

var (
	numberScientificPlusAndZeros = regexp.MustCompile(`^([+-]?[\d.]+e)(?:\+|(-))?0*(\d)`)
	numberUselessScientific      = regexp.MustCompile(`^([+-]?[\d.]+)e[+-]?0+$`)
	numberLeadingDot             = regexp.MustCompile(`^([+-])?\.`)
	numberTrailingZeros          = regexp.MustCompile(`(\.\d+?)0+(e|$)`)
	numberTrailingDot            = regexp.MustCompile(`\.(e|$)`)
)

// printNumber is upstream's printNumber, utilities/print-number.js. Go's regexp has no lookahead, so
// each `(?=x)` is captured and put back.
func printNumber(rawNumber string) string {
	if len(rawNumber) == 1 {
		return rawNumber
	}
	value := strings.ToLower(rawNumber)
	value = replaceFirst(numberScientificPlusAndZeros, value, "${1}${2}${3}")
	value = replaceFirst(numberUselessScientific, value, "${1}")
	value = replaceFirst(numberLeadingDot, value, "${1}0.")
	value = replaceFirst(numberTrailingZeros, value, "${1}${2}")
	value = replaceFirst(numberTrailingDot, value, "${1}")
	return value
}

// replaceFirst is JavaScript's String.prototype.replace with a non-global pattern.
func replaceFirst(pattern *regexp.Regexp, text string, template string) string {
	match := pattern.FindStringSubmatchIndex(text)
	if match == nil {
		return text
	}
	expanded := pattern.ExpandString(nil, template, text, match)
	return text[:match[0]] + string(expanded) + text[match[1]:]
}

// getAlignmentSize is upstream's getAlignmentSize: columns, with tabs to the next tab stop.
func getAlignmentSize(text string, tabWidth int, startIndex int) int {
	size := 0
	for index := startIndex; index < len(text); {
		if text[index] == '\t' {
			size = size + tabWidth - size%tabWidth
			index++
			continue
		}
		// One per UTF-16 code unit, as upstream counts `text[i]`.
		character, width := utf8.DecodeRuneInString(text[index:])
		if character > 0xFFFF {
			size += 2
		} else {
			size++
		}
		index += width
	}
	return size
}

// getIndentSize is upstream's getIndentSize: the alignment of the last line's leading whitespace.
func getIndentSize(value string, tabWidth int) int {
	lastNewline := strings.LastIndexByte(value, '\n')
	if lastNewline == -1 {
		return 0
	}
	rest := value[lastNewline+1:]
	end := 0
	for end < len(rest) && (rest[end] == '\t' || rest[end] == ' ') {
		end++
	}
	return getAlignmentSize(rest[:end], tabWidth, 0)
}
