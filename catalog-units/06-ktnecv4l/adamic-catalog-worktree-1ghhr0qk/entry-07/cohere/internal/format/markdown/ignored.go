package markdown

import (
	"regexp"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// src/language-markdown/print/ignored.js.

var trailingBlockquoteMarkerPattern = regexp.MustCompile(`\n>[` + javaScriptSpace + `]*$`)

func printPrettierIgnored(path *astPath, options *options, _ printing.PrintFunc, _ any) doc.Doc {
	node := currentNode(path)
	originalText := options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset]

	if node.NodeType == "list" &&
		path.HasAncestor(func(node *Node) bool { return node.NodeType == "blockquote" }) &&
		settingsOf(options).proseWrap != "always" {
		if location := trailingBlockquoteMarkerPattern.FindStringIndex(originalText); location != nil {
			return doc.Text(originalText[:location[0]] + originalText[location[1]:])
		}
	}
	return doc.Text(originalText)
}

// hasPrettierIgnore is the printer's: the previous sibling is a `<!-- prettier-ignore -->`.
func hasPrettierIgnore(path *astPath) bool {
	index, present := path.Index()
	return present && index > 0 && isPrettierIgnore(previousNode(path)) == "next"
}
