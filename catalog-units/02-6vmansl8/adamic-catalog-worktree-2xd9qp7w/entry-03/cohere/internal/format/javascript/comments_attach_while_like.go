package javascript

import "github.com/system-inc/cohere/internal/format/printing"

// comments/attach/handle-while-like-comments.js

// handleWhileLikeComments is upstream's handleWhileLikeComments.
func handleWhileLikeComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode, text := context.Comment, context.Preceding, context.Enclosing,
		context.Following, context.Text
	if !(enclosingNode.Is("WhileStatement", "WithStatement") && followingNode != nil) {
		return false
	}

	// We unfortunately have no way using the AST or location of nodes to know
	// if the comment is positioned before the condition parenthesis:
	//   while (a /* comment */) {}
	// The only workaround I found is to look at the next character to see if
	// it is a ).
	nextCharacter := getNextNonSpaceNonCommentCharacter(text, locEnd(comment))
	if nextCharacter == ")" {
		printing.AddTrailingComment(precedingNode, comment)
		return true
	}

	if enclosingNode.Child("body") == followingNode {
		printing.AddLeadingComment(followingNode, comment)
		return true
	}

	return false
}
