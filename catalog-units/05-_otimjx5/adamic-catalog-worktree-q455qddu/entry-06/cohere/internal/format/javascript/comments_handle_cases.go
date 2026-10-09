package javascript

import "github.com/system-inc/cohere/internal/format/printing"

// comments/handle-comments.js, from isAssignmentLikeNode to the end of the file. The first half is in
// comments_handle.go.

// isAssignmentLikeNode is upstream's createTypeCheckFunction of that name.
func isAssignmentLikeNode(node Node) bool {
	return node.Is("VariableDeclarator", "AssignmentExpression", "TypeAlias", "TSTypeAliasDeclaration")
}

// isComplexExprNode is upstream's createTypeCheckFunction of that name.
func isComplexExprNode(node Node) bool {
	return node.Is("ObjectExpression", "ArrayExpression", "TemplateLiteral", "TaggedTemplateExpression",
		"ObjectTypeAnnotation", "TSTypeLiteral")
}

// handleAssignmentLikeComments is upstream's handleAssignmentLikeComments.
func handleAssignmentLikeComments(context *commentContext) bool {
	comment, enclosingNode, followingNode, options, placement := context.Comment, context.Enclosing,
		context.Following, context.Options, context.Placement
	if isAssignmentLikeNode(enclosingNode) &&
		followingNode != nil &&
		placement == "endOfLine" &&
		(isComplexExprNode(followingNode) || isBlockComment(comment)) {
		return addLeadingCommentToPossibleUnionType(followingNode, context)
	}

	// Ideally, we should only check cases that the right side is a single-element union or intersection type
	// We already strip the wrapper, there is no way to know it, so we only check "type alias"
	if isTypeAlias(enclosingNode) && followingNode != nil {
		leftSide := enclosingNode.Child("id")
		equalsTokenIndex := indexOfFrom(stripComments(options), "=", locEnd(leftSide))

		if locStart(comment) >= equalsTokenIndex {
			return addLeadingCommentToPossibleUnionType(followingNode, context)
		}
	}

	return false
}

// handleTSFunctionTrailingComments is upstream's handleTSFunctionTrailingComments.
func handleTSFunctionTrailingComments(context *commentContext) bool {
	comment, enclosingNode, precedingNode, followingNode, text := context.Comment, context.Enclosing,
		context.Preceding, context.Following, context.Text
	if followingNode == nil &&
		enclosingNode.Is("TSMethodSignature", "TSDeclareFunction", "TSAbstractMethodDefinition") &&
		(precedingNode == nil || precedingNode != enclosingNode.Child("returnType")) &&
		getNextNonSpaceNonCommentCharacter(text, locEnd(comment)) == ";" {
		printing.AddTrailingComment(enclosingNode, comment)
		return true
	}
	return false
}

// handleIgnoreComments is upstream's handleIgnoreComments.
func handleIgnoreComments(context *commentContext) bool {
	comment, enclosingNode, followingNode := context.Comment, context.Enclosing, context.Following
	if isPrettierIgnoreComment(comment) &&
		enclosingNode.Is("TSMappedType") &&
		followingNode == enclosingNode.Child("key") {
		enclosingNode.Set("prettierIgnore", true)
		comment.Set("unignore", true)
		return true
	}
	return false
}

// isBeforeMappedTypeOpeningBracket is upstream's isBeforeMappedTypeOpeningBracket.
func isBeforeMappedTypeOpeningBracket(node Node, comment Node, options *Options) bool {
	bracketIndex := indexOfFrom(stripComments(options), "[", locStart(node))
	return locEnd(comment) < bracketIndex
}

// handleTSMappedTypeComments is upstream's handleTSMappedTypeComments.
func handleTSMappedTypeComments(context *commentContext) bool {
	comment, enclosingNode, options := context.Comment, context.Enclosing, context.Options
	if !enclosingNode.Is("TSMappedType") {
		return false
	}

	if isBeforeMappedTypeOpeningBracket(enclosingNode, comment, options) {
		printing.AddDanglingComment(enclosingNode, comment, "")
		return true
	}
	return false
}

// handleSwitchDefaultCaseComments is upstream's handleSwitchDefaultCaseComments.
func handleSwitchDefaultCaseComments(context *commentContext) bool {
	comment, enclosingNode, followingNode := context.Comment, context.Enclosing, context.Following
	// `enclosingNode.consequent[0]`, undefined for an empty list.
	var firstConsequent Node
	if consequent := enclosingNode.List("consequent"); len(consequent) > 0 {
		firstConsequent = consequent[0]
	}
	if enclosingNode == nil ||
		!enclosingNode.Is("SwitchCase") ||
		enclosingNode.Truthy("test") ||
		followingNode == nil ||
		followingNode != firstConsequent {
		return false
	}

	if followingNode.Is("BlockStatement") && isLineComment(comment) {
		addBlockStatementFirstComment(followingNode, comment)
	} else {
		printing.AddDanglingComment(enclosingNode, comment, "")
	}

	return true
}

// handleLastUnionElementInExpression is upstream's handleLastUnionElementInExpression.
//
// Handle `Comment2`, `Comment4`, `Comment6`.
//
//	type Foo = (
//	  | "thing1" // Comment1
//	  | "thing2" // Comment2
//	)[];
//
//	type Foo = (
//	  | "thing1" // Comment3
//	  | "thing2" // Comment4
//	) & Bar;
//
//	type Foo = (
//	  | "thing1" // Comment5
//	  | "thing2" // Comment6
//	) | Bar;
func handleLastUnionElementInExpression(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode := context.Comment, context.Preceding, context.Enclosing,
		context.Following
	if isUnionType(precedingNode) &&
		(isArrayType(enclosingNode) && followingNode == nil ||
			isIntersectionType(enclosingNode) ||
			isUnionType(enclosingNode)) {
		types := precedingNode.List("types")
		printing.AddTrailingComment(types[len(types)-1], comment)
		return true
	}
	return false
}

// handleCommentsInDestructuringPattern is upstream's handleCommentsInDestructuringPattern.
//
//	const [
//	  foo,
//	  // bar
//	  // baz
//	]: Foo = foo();
//
//	const {
//	  foo,
//	  // bar
//	  // baz
//	}: Foo = foo();
func handleCommentsInDestructuringPattern(context *commentContext) bool {
	comment, enclosingNode, precedingNode, followingNode := context.Comment, context.Enclosing, context.Preceding,
		context.Following
	if enclosingNode.Is("ObjectPattern", "ArrayPattern") &&
		followingNode.Is("TSTypeAnnotation") {
		if precedingNode != nil {
			printing.AddTrailingComment(precedingNode, comment)
		} else {
			// const {
			//   // bar
			//   // baz
			// }: Foo = expr;
			printing.AddDanglingComment(enclosingNode, comment, "")
		}
		return true
	}
	return false
}

// handleLastBinaryOperatorOperand is upstream's handleLastBinaryOperatorOperand.
func handleLastBinaryOperatorOperand(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode, text := context.Comment, context.Preceding,
		context.Enclosing, context.Following, context.Text
	// "baz" should be a trailing comment of `cond3`:
	//
	//   !(
	//     cond1 || // foo
	//     cond2 || // bar
	//     cond3 // baz
	//   );
	if followingNode == nil &&
		enclosingNode.Is("UnaryExpression") &&
		precedingNode.Is("LogicalExpression", "BinaryExpression") {
		//   !(
		//     (cond1 || cond2) // foo
		//   );
		// Multiline expression
		if hasNewlineInRange(text, locStart(enclosingNode.Child("argument")), locStart(precedingNode.Child("right"))) &&
			isSingleLineComment(comment, text) &&
			// Comment and `precedingNode.right` are on same line
			!hasNewlineInRange(text, locStart(precedingNode.Child("right")), locStart(comment)) {
			printing.AddTrailingComment(precedingNode.Child("right"), comment)
			return true
		}
	}
	return false
}

// handlePropertySignatureComments is upstream's handlePropertySignatureComments.
func handlePropertySignatureComments(context *commentContext) bool {
	enclosingNode, followingNode, comment := context.Enclosing, context.Following, context.Comment
	if enclosingNode != nil &&
		enclosingNode.Is("TSPropertySignature", "ObjectTypeProperty") &&
		(isUnionType(followingNode) || isIntersectionType(followingNode)) {
		printing.AddLeadingComment(followingNode, comment)
		return true
	}
	return false
}

// handleBinaryCastExpressionComment is upstream's handleBinaryCastExpressionComment.
func handleBinaryCastExpressionComment(context *commentContext) bool {
	enclosingNode, precedingNode, followingNode, comment, text := context.Enclosing, context.Preceding,
		context.Following, context.Comment, context.Text
	// Avoid break before `as` and `satisfies`
	if isBinaryCastExpression(enclosingNode) &&
		precedingNode == enclosingNode.Child("expression") &&
		!isSingleLineComment(comment, text) {
		if followingNode != nil {
			printing.AddLeadingComment(followingNode, comment)
		} else {
			printing.AddTrailingComment(enclosingNode, comment)
		}
		return true
	}
	return false
}

// isCommentBeforeArrowFunctionExpressionArrow is upstream's isCommentBeforeArrowFunctionExpressionArrow.
func isCommentBeforeArrowFunctionExpressionArrow(comment Node, arrowFunctionExpression Node, options *Options) bool {
	arrowTokenIndex := lastIndexOfFrom(stripComments(options), "=>", locStart(arrowFunctionExpression.Child("body")))

	return locEnd(comment) < arrowTokenIndex
}

// handleArrowExpressionComments is upstream's handleArrowExpressionComments.
//
// Avoid attaching multiline comment to node before arrow
//
//	const test = (): any => /* first line
//	second line
//	*\/
//	null;
func handleArrowExpressionComments(context *commentContext) bool {
	comment, enclosingNode, followingNode, precedingNode, options := context.Comment, context.Enclosing,
		context.Following, context.Preceding, context.Options
	if !enclosingNode.Is("ArrowFunctionExpression") ||
		followingNode == nil ||
		precedingNode == nil {
		return false
	}

	isBeforeArrow := isCommentBeforeArrowFunctionExpressionArrow(comment, enclosingNode, options)

	if !isBeforeArrow {
		addBlockOrNotComment(followingNode, comment)
		return true
	}

	return false
}

// handleParenthesizedExpressionTrailingComment is upstream's handleParenthesizedExpressionTrailingComment.
func handleParenthesizedExpressionTrailingComment(context *commentContext) bool {
	comment, enclosingNode, precedingNode, followingNode := context.Comment, context.Enclosing, context.Preceding,
		context.Following
	if followingNode == nil && enclosingNode != nil && precedingNode != nil {
		if enclosingNode.Is("ExpressionStatement") &&
			enclosingNode.Child("expression") == precedingNode {
			printing.AddTrailingComment(enclosingNode, comment)
			return true
		}

		isSequence := precedingNode.Is("SequenceExpression")
		isAssignment := precedingNode.Is("AssignmentExpression")

		if (isSequence || isAssignment) &&
			(enclosingNode.Is("ArrowFunctionExpression") &&
				enclosingNode.Child("body") == precedingNode ||
				enclosingNode.Is("VariableDeclarator") &&
					enclosingNode.Child("init") == precedingNode ||
				enclosingNode.Is("ReturnStatement") &&
					enclosingNode.Child("argument") == precedingNode ||
				enclosingNode.Is("AssignmentExpression") &&
					enclosingNode.Child("right") == precedingNode) {
			target := precedingNode.Child("right")
			if isSequence {
				expressions := precedingNode.List("expressions")
				target = expressions[len(expressions)-1]
			}
			printing.AddTrailingComment(target, comment)
			return true
		}
	}

	return false
}

// handleUnionTypeLeadingComments is upstream's handleUnionTypeLeadingComments.
func handleUnionTypeLeadingComments(context *commentContext) bool {
	followingNode, comment := context.Following, context.Comment

	if shouldAttachToUnionTypeFirstElement(followingNode, context) {
		printing.AddLeadingComment(followingNode.List("types")[0], comment)
		return true
	}

	return false
}

// isRealFunctionLikeNode is upstream's createTypeCheckFunction of that name.
func isRealFunctionLikeNode(node Node) bool {
	return node.Is("ArrowFunctionExpression", "FunctionExpression", "FunctionDeclaration", "ObjectMethod", "ClassMethod",
		"TSDeclareFunction", "TSCallSignatureDeclaration", "TSConstructSignatureDeclaration", "TSMethodSignature",
		"TSConstructorType", "TSFunctionType", "TSDeclareMethod", "HookDeclaration")
}
