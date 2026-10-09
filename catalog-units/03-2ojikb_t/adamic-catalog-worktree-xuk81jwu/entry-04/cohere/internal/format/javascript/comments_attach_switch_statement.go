package javascript

import "github.com/system-inc/cohere/internal/format/printing"

// comments/attach/handle-switch-statement-comments.js

// handleSwitchStatementComments is upstream's handleSwitchStatementComments.
func handleSwitchStatementComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode, text := context.Comment, context.Preceding, context.Enclosing,
		context.Following, context.Text
	if !(enclosingNode.Is("SwitchStatement") &&
		len(enclosingNode.List("cases")) == 0 &&
		followingNode == nil &&
		precedingNode == enclosingNode.Child("discriminant")) {
		return false
	}

	nextCharacter := getNextNonSpaceNonCommentCharacter(text, locEnd(comment))

	if nextCharacter == "}" {
		printing.AddDanglingComment(enclosingNode, comment, "")
		return true
	}

	return false
}
