package javascript

import "github.com/system-inc/cohere/internal/format/printing"

// comments/attach/handle-if-statement-comments.js

// There are often comments before the else clause of if statements like
//
//	if (1) { ... }
//	// comment
//	else { ... }
//
// They are being attached as leading comments of the BlockExpression which
// is not well printed. What we want is to instead move the comment inside
// of the block and make it leadingComment of the first element of the block
// or dangling comment of the block if there is nothing inside
//
//	if (1) { ... }
//	else {
//	  // comment
//	  ...
//	}
func handleIfStatementComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode, text := context.Comment, context.Preceding, context.Enclosing,
		context.Following, context.Text
	if !enclosingNode.Is("IfStatement") || followingNode == nil {
		return false
	}

	// We unfortunately have no way using the AST or location of nodes to know
	// if the comment is positioned before the condition parenthesis:
	//   if (a /* comment */) {}
	// The only workaround I found is to look at the next character to see if
	// it is a ).
	nextCharacter := getNextNonSpaceNonCommentCharacter(text, locEnd(comment))
	if nextCharacter == ")" {
		printing.AddTrailingComment(precedingNode, comment)
		return true
	}

	// if comment is positioned between the condition and its body
	if followingNode == enclosingNode.Child("consequent") {
		printing.AddLeadingComment(followingNode, comment)
		return true
	}

	if precedingNode == enclosingNode.Child("consequent") && followingNode == enclosingNode.Child("alternate") {
		return handleCommentsBetween(context)
	}

	return false
}

// handleCommentsBetween is upstream's handleCommentsBetween. Upstream passes a fresh object of the
// same six fields, so the context passes through unchanged.
func handleCommentsBetween(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode, text, options := context.Comment, context.Preceding,
		context.Enclosing, context.Following, context.Text, context.Options
	elseTokenIndex := indexOfFrom(stripComments(options), "else", locEnd(enclosingNode.Child("consequent")))

	// if comment is positioned after the `else` token
	if locStart(comment) >= elseTokenIndex {
		printing.AddLeadingComment(followingNode, comment)
		return true
	}

	isConsequentBlockStatement := precedingNode.Is("BlockStatement")

	// Comments before `else`:
	// - treat as trailing comments of the consequent, if it's a BlockStatement
	// - treat as a dangling comment otherwise
	if !isConsequentBlockStatement &&
		isSingleLineComment(comment, text) &&
		// Comment and `precedingNode` are on same line
		!hasNewlineInRange(text, locEnd(precedingNode), locStart(comment)) {
		// example:
		//   if (cond1) expr1; // comment A
		//   else if (cond2) expr2; // comment A
		//   else expr3;

		printing.AddTrailingComment(precedingNode, comment)
		return true
	}

	printing.AddDanglingComment(enclosingNode, comment, "")
	return true
}
