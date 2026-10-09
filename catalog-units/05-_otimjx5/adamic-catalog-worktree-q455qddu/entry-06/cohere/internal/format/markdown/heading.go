package markdown

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// src/language-markdown/print/heading.js.

func printAtxHeading(path *astPath, options *options, print printing.PrintFunc) doc.Doc {
	return doc.Concat{doc.Text(strings.Repeat("#", currentNode(path).Depth) + " "), printChildren(path, options, print, nil)}
}

func printSetextHeading(path *astPath, options *options, print printing.PrintFunc) doc.Doc {
	originalText := options.OriginalText
	end := currentNode(path).Position.End.Offset
	lineStart := strings.LastIndex(originalText[:max(end, 0)], "\n") + 1
	lastLine := originalText[lineStart:end]
	start := max(strings.Index(lastLine, "="), strings.Index(lastLine, "-"))
	underline := lastLine[max(start, 0):]
	if start < 0 {
		// lastLine.slice(-1) when neither is found.
		underline = lastLine[max(len(lastLine)-1, 0):]
	}
	return doc.Concat{printChildren(path, options, print, nil), doc.Hardline, doc.Text(underline)}
}

func printHeading(path *astPath, options *options, print printing.PrintFunc) doc.Doc {
	if isSetextHeading(currentNode(path)) {
		return printSetextHeading(path, options, print)
	}

	return printAtxHeading(path, options, print)
}
