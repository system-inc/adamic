package javascript

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/estree"
)

// utilities/is-next-line-empty.js, utilities/is-previous-line-empty.js, utilities/strip-comments.js,
// utilities/get-shebang.js, utilities/should-print-trailing-comma.js.

// isNextLineEmptyAfter is language-js's isNextLineEmpty(node, options). utility_text.go already holds
// src/utilities' isNextLineEmpty(text, index), which upstream imports here as isNextLineEmptyAfterIndex,
// so this node-taking wrapper takes a distinct name.
func isNextLineEmptyAfter(node Node, options *Options) bool {
	end := locEnd(node)

	if isNextLineEmpty(options.OriginalText, end) {
		return true
	}

	endWithSemicolon := estree.LocEndWithFullText(node)
	if endWithSemicolon == end {
		return false
	}

	return isNextLineEmpty(options.OriginalText, endWithSemicolon)
}

// isPreviousLineEmptyBefore is language-js's isPreviousLineEmpty(node, options), named apart from
// src/utilities' isPreviousLineEmpty(text, index) in utility_text.go for the same reason.
func isPreviousLineEmptyBefore(node Node, options *Options) bool {
	return isPreviousLineEmpty(options.OriginalText, locStart(node))
}

// stripComments is upstream's stripComments: the original text with every comment blanked to spaces,
// computed once per format. Upstream caches it in a WeakMap keyed on the comments array; the Go keeps
// it on settings, as the parse's postprocess made it when it ran one, since that stripped the same text of
// the same comments.
func stripComments(options *Options) string {
	if settingsOf(options).strippedText == nil {
		settingsOf(options).strippedText = estree.NewStrippedText(options.OriginalText, options.Comments)
	}
	return settingsOf(options).strippedText.Text()
}

// getShebang is upstream's getShebang.
func getShebang(text string) string {
	if !strings.HasPrefix(text, "#!") {
		return ""
	}
	index := strings.IndexByte(text, '\n')
	if index == -1 {
		return text
	}
	return text[:index]
}

// shouldPrintTrailingComma is upstream's shouldPrintTrailingComma. Upstream's default level is
// "es5"; Go has no default parameters, so callers pass it.
func shouldPrintTrailingComma(options *Options, level string) bool {
	return settingsOf(options).TrailingComma == "es5" && level == "es5" ||
		settingsOf(options).TrailingComma == "all" && (level == "all" || level == "es5")
}

// utf16Length is JavaScript's String.length, UTF-16 code units, for the places upstream compares a
// `.length` rather than a getStringWidth.
func utf16Length(text string) int {
	length := 0
	for _, character := range text {
		if character > 0xFFFF {
			length += 2
		} else {
			length++
		}
	}
	return length
}
