package javascript

import "github.com/system-inc/cohere/internal/format/printing"

// comments/attach/utilities.js

// isSingleLineBlockComment is upstream's isSingleLineBlockComment.
func isSingleLineBlockComment(comment Node, text string) bool {
	return isBlockComment(comment) && !hasNewlineInRange(text, locStart(comment), locEnd(comment))
}

// isSingleLineComment is upstream's isSingleLineComment.
func isSingleLineComment(comment Node, text string) bool {
	return isLineComment(comment) || isSingleLineBlockComment(comment, text)
}

// addBlockOrNotComment is upstream's addBlockOrNotComment.
func addBlockOrNotComment(node Node, comment Node) {
	if node.Is("BlockStatement") {
		addBlockStatementFirstComment(node, comment)
	} else {
		printing.AddLeadingComment(node, comment)
	}
}

// addBlockStatementFirstComment is upstream's addBlockStatementFirstComment.
func addBlockStatementFirstComment(node Node, comment Node) {
	// `node.body || node.properties`: an array is truthy even when empty, so body wins whenever present.
	body := node.List("body")
	if !node.Truthy("body") {
		body = node.List("properties")
	}
	var firstNonEmptyNode Node
	for _, child := range body {
		if child.Type() != "EmptyStatement" {
			firstNonEmptyNode = child
			break
		}
	}
	if firstNonEmptyNode != nil {
		printing.AddLeadingComment(firstNonEmptyNode, comment)
	} else {
		printing.AddDanglingComment(node, comment, "")
	}
}
