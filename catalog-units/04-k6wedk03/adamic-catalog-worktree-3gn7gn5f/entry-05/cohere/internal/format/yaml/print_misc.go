package yaml

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
)

// src/language-yaml/print/misc.js.

// printNextEmptyLine is upstream's printNextEmptyLine: a softline after the node when an empty line
// follows it, printed once per end offset.
//
// Upstream keeps the printed offsets in a WeakMap keyed by path.root, one set per parsed file. The set
// lives on this print's settings instead, which is the same lifetime: one root per print.
func printNextEmptyLine(path *astPath, originalText string, settings *settings) doc.Doc {
	node := currentNode(path)

	if !settings.isNextEmptyLinePrinted[node.Position.End.Offset] {
		settings.isNextEmptyLinePrinted[node.Position.End.Offset] = true
		if isNextLineEmpty(node, originalText) && !shouldPrintEndComments(parentNode(path)) {
			return doc.Softline
		}
	}

	return doc.Text("")
}

// shouldPrintEndComments is upstream's shouldPrintEndComments.
func shouldPrintEndComments(node *unistNode) bool {
	return hasEndComments(node) &&
		!isNode(node, "documentHead", "documentBody", "flowMapping", "flowSequence")
}

// alignWithSpaces is upstream's alignWithSpaces: align by a string of width spaces, so the alignment
// stays spaces under useTabs. A width of 0 is the empty string, which upstream's makeAlign treats as no
// alignment at all.
func alignWithSpaces(width int, contents doc.Doc) doc.Doc {
	return doc.AlignWithString(strings.Repeat(" ", width), contents)
}
