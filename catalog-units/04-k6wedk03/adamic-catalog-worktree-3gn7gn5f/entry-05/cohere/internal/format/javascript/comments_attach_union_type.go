package javascript

import "github.com/system-inc/cohere/internal/format/printing"

// comments/attach/union-type.js

// shouldAttachToUnionTypeFirstElement is upstream's shouldAttachToUnionTypeFirstElement.
func shouldAttachToUnionTypeFirstElement(node Node, context *commentContext) bool {
	comment, text, options := context.Comment, context.Text, context.Options
	if isUnionType(node) &&
		isSingleLineBlockComment(comment, text) &&
		!isPrettierIgnoreComment(comment) {
		text := stripComments(options)
		// JavaScript's slice answers "" for a reversed range, where Go's would panic.
		textBetween := ""
		if locEnd(comment) < locStart(node) {
			textBetween = text[locEnd(comment):locStart(node)]
		}
		// /^[ \t]*$/
		for index := 0; index < len(textBetween); index++ {
			if textBetween[index] != ' ' && textBetween[index] != '\t' {
				return false
			}
		}
		return true
	}

	return false
}

// addLeadingCommentToPossibleUnionType is upstream's addLeadingCommentToPossibleUnionType.
func addLeadingCommentToPossibleUnionType(node Node, context *commentContext) bool {
	target := node
	if shouldAttachToUnionTypeFirstElement(node, context) {
		target = node.List("types")[0]
	}
	printing.AddLeadingComment(target, context.Comment)
	return true
}
