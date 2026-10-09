package javascript

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// comments/handle-comments.js

// commentContext is upstream's CommentContext: the comment, its preceding, enclosing and following
// nodes, the text, the options, the ast, whether it is the last comment, and its placement
// ("ownLine" | "endOfLine" | "remaining"). An absent node is nil.
type commentContext = printing.CommentContext[*estree.Node]

// commentHandler is one of upstream's handlers. Upstream's handlers that fall off the end return
// undefined, which `.some` reads as false; here they return false.
type commentHandler func(context *commentContext) bool

// someHandler is upstream's `[...].some((fn) => fn(context))`.
func someHandler(context *commentContext, handlers ...commentHandler) bool {
	for _, handler := range handlers {
		if handler(context) {
			return true
		}
	}
	return false
}

// handleOwnLineComment is upstream's handleOwnLineComment.
func handleOwnLineComment(context *commentContext) bool {
	return someHandler(context,
		handleCommentInEmptyParens,
		handleIgnoreComments,
		handleConditionalExpressionComments,
		handleLastFunctionParameterComments,
		handleMemberExpressionComments,
		handleIfStatementComments,
		handleWhileLikeComments,
		handleSwitchStatementComments,
		handleTryStatementComments,
		handleClassComments,
		handleForXStatementComments,
		handleUnionTypeComments,
		handleMatchOrPatternComments,
		handleOnlyComments,
		handleModuleSpecifiersComments,
		handleAssignmentPatternComments,
		handleMethodNameComments,
		handleLabeledStatementComments,
		handleNestedConditionalExpressionComments,
		handleCommentsInDestructuringPattern,
		handleTSMappedTypeComments,
		handleBinaryCastExpressionComment,
		handleUnionTypeLeadingComments,
	)
}

// handleEndOfLineComment is upstream's handleEndOfLineComment.
func handleEndOfLineComment(context *commentContext) bool {
	return someHandler(context,
		handleCommentInEmptyParens,
		handleClosureTypeCastComments,
		handleLastFunctionParameterComments,
		handleConditionalExpressionComments,
		handleModuleSpecifiersComments,
		handleIfStatementComments,
		handleWhileLikeComments,
		handleSwitchStatementComments,
		handleTryStatementComments,
		handleClassComments,
		handleForXStatementComments,
		handleLabeledStatementComments,
		handleCallExpressionComments,
		handlePropertyComments,
		handleOnlyComments,
		handleAssignmentLikeComments,
		handleSwitchDefaultCaseComments,
		handleLastUnionElementInExpression,
		handleLastBinaryOperatorOperand,
		handleTSMappedTypeComments,
		handleArrowExpressionComments,
		handleParenthesizedExpressionTrailingComment,
		handlePropertySignatureComments,
		handleBinaryCastExpressionComment,
	)
}

// handleRemainingComment is upstream's handleRemainingComment.
func handleRemainingComment(context *commentContext) bool {
	return someHandler(context,
		handleCommentInEmptyParens,
		handleIgnoreComments,
		handleIfStatementComments,
		handleWhileLikeComments,
		handleSwitchStatementComments,
		handleForXStatementComments,
		handleMethodNameComments,
		handleOnlyComments,
		handleAssignmentLikeComments,
		handleTSMappedTypeComments,
		handleCommentAfterArrowParams,
		handleFunctionNameComments,
		handleTSFunctionTrailingComments,
		handleParenthesizedExpressionTrailingComment,
		handleBinaryCastExpressionComment,
		handleUnionTypeLeadingComments,
	)
}

// getCommentChildNodes is the printer's getCommentChildNodes hook. The estree printer in printers.js
// defines none, so it never answers and the core's visitor-key children apply.
func getCommentChildNodes(_ Node, _ *Options) ([]Node, bool) {
	return nil, false
}

// handleClosureTypeCastComments is upstream's handleClosureTypeCastComments.
func handleClosureTypeCastComments(context *commentContext) bool {
	comment, followingNode := context.Comment, context.Following
	if followingNode != nil && estree.IsTypeCastComment(comment) {
		printing.AddLeadingComment(followingNode, comment)
		return true
	}
	return false
}

// Same as IfStatement but for TryStatement
func handleTryStatementComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode := context.Comment, context.Preceding, context.Enclosing,
		context.Following
	if !enclosingNode.Is("TryStatement") && !enclosingNode.Is("CatchClause") || followingNode == nil {
		return false
	}

	if enclosingNode.Is("CatchClause") && precedingNode != nil {
		printing.AddTrailingComment(precedingNode, comment)
		return true
	}

	if followingNode.Is("BlockStatement") {
		addBlockStatementFirstComment(followingNode, comment)
		return true
	}

	if followingNode.Is("TryStatement") {
		addBlockOrNotComment(followingNode.Child("finalizer"), comment)
		return true
	}

	if followingNode.Is("CatchClause") {
		addBlockOrNotComment(followingNode.Child("body"), comment)
		return true
	}

	return false
}

// handleMemberExpressionComments is upstream's handleMemberExpressionComments.
func handleMemberExpressionComments(context *commentContext) bool {
	comment, enclosingNode, followingNode := context.Comment, context.Enclosing, context.Following
	if isMemberExpression(enclosingNode) && followingNode.Is("Identifier") {
		printing.AddLeadingComment(enclosingNode, comment)
		return true
	}

	return false
}

// handleNestedConditionalExpressionComments is upstream's handleNestedConditionalExpressionComments.
// It acts only under experimentalTernaries, which our configs never set, so it always declines; the
// rest of its body is dropped.
func handleNestedConditionalExpressionComments(_ *commentContext) bool {
	return false
}

// handleConditionalExpressionComments is upstream's handleConditionalExpressionComments. The
// experimentalTernaries branch, which adds a dangling comment, is dropped.
func handleConditionalExpressionComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode, text := context.Comment, context.Preceding,
		context.Enclosing, context.Following, context.Text
	isSameLineAsPrecedingNode := precedingNode != nil &&
		!hasNewlineInRange(text, locEnd(precedingNode), locStart(comment))

	if (precedingNode == nil || !isSameLineAsPrecedingNode) &&
		(enclosingNode.Is("ConditionalExpression") || isConditionalType(enclosingNode)) &&
		followingNode != nil {
		printing.AddLeadingComment(followingNode, comment)
		return true
	}
	return false
}

// isClassLikeNode is upstream's createTypeCheckFunction of that name.
func isClassLikeNode(node Node) bool {
	return node.Is("ClassDeclaration", "ClassExpression", "DeclareClass", "DeclareInterface", "InterfaceDeclaration",
		"TSInterfaceDeclaration")
}

// handleClassComments is upstream's handleClassComments.
func handleClassComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode := context.Comment, context.Preceding, context.Enclosing,
		context.Following
	if isClassLikeNode(enclosingNode) {
		decorators := enclosingNode.List("decorators")
		if len(decorators) > 0 && !followingNode.Is("Decorator") {
			printing.AddTrailingComment(decorators[len(decorators)-1], comment)
			return true
		}

		if enclosingNode.Child("body") != nil && followingNode == enclosingNode.Child("body") {
			addBlockStatementFirstComment(enclosingNode.Child("body"), comment)
			return true
		}

		// Don't add leading comments to `implements`, `extends`, `mixins` to
		// avoid printing the comment after the keyword.
		if followingNode != nil {
			superClass := enclosingNode.Child("superClass")
			if superClass != nil &&
				followingNode == superClass &&
				precedingNode != nil &&
				(precedingNode == enclosingNode.Child("id") ||
					precedingNode == enclosingNode.Child("typeParameters")) {
				printing.AddTrailingComment(precedingNode, comment)
				return true
			}

			for _, prop := range []string{"implements", "extends", "mixins"} {
				// `enclosingNode[prop] && followingNode === enclosingNode[prop][0]`: an empty list's [0] is
				// undefined, which a present followingNode never equals.
				list := enclosingNode.List(prop)
				if len(list) > 0 && followingNode == list[0] {
					if precedingNode != nil &&
						(precedingNode == enclosingNode.Child("id") ||
							precedingNode == enclosingNode.Child("typeParameters") ||
							precedingNode == superClass) {
						printing.AddTrailingComment(precedingNode, comment)
					} else {
						printing.AddDanglingComment(enclosingNode, comment, prop)
					}
					return true
				}
			}
		}
	}
	return false
}

// isPropertyLikeNode is upstream's createTypeCheckFunction of that name.
func isPropertyLikeNode(node Node) bool {
	return node.Is("ClassMethod", "ClassProperty", "PropertyDefinition", "TSAbstractPropertyDefinition",
		"TSAbstractMethodDefinition", "TSDeclareMethod", "MethodDefinition", "ClassAccessorProperty", "AccessorProperty",
		"TSAbstractAccessorProperty", "TSParameterProperty")
}

// handleMethodNameComments is upstream's handleMethodNameComments.
func handleMethodNameComments(context *commentContext) bool {
	placement, comment, precedingNode, enclosingNode, followingNode, text := context.Placement, context.Comment,
		context.Preceding, context.Enclosing, context.Following, context.Text
	// This is only needed for estree parsers (Flow, TypeScript) to attach
	// after a method name:
	// obj = { fn /*comment*/() {} };
	if enclosingNode != nil &&
		precedingNode != nil &&
		getNextNonSpaceNonCommentCharacter(text, locEnd(comment)) == "(" &&
		// "MethodDefinition" is handled in `canAttachComment`
		enclosingNode.Is("Property", "TSDeclareMethod", "TSAbstractMethodDefinition") &&
		precedingNode.Is("Identifier") &&
		enclosingNode.Child("key") == precedingNode &&
		// special Property case: { key: /*comment*/(value) };
		// comment should be attached to value instead of key
		getNextNonSpaceNonCommentCharacter(text, locEnd(precedingNode)) != ":" {
		printing.AddTrailingComment(precedingNode, comment)
		return true
	}

	if isPropertyLikeNode(enclosingNode) &&
		followingNode == nil &&
		placement == "remaining" {
		target := enclosingNode
		if getNextNonSpaceNonCommentCharacter(text, locEnd(comment)) == "(" {
			target = precedingNode
		}
		printing.AddTrailingComment(target, comment)
		return true
	}

	// Print comments between decorators and class methods as a trailing comment
	// on the decorator node instead of the method node
	if precedingNode.Is("Decorator") &&
		isPropertyLikeNode(enclosingNode) &&
		(isLineComment(comment) || placement == "ownLine") {
		printing.AddTrailingComment(precedingNode, comment)
		return true
	}

	return false
}

// isFunctionLikeNode is upstream's createTypeCheckFunction of that name.
func isFunctionLikeNode(node Node) bool {
	return node.Is("FunctionDeclaration", "FunctionExpression", "ClassMethod", "MethodDefinition", "ObjectMethod")
}

// handleFunctionNameComments is upstream's handleFunctionNameComments.
func handleFunctionNameComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, text := context.Comment, context.Preceding, context.Enclosing, context.Text
	if getNextNonSpaceNonCommentCharacter(text, locEnd(comment)) != "(" {
		return false
	}
	if precedingNode != nil && isFunctionLikeNode(enclosingNode) {
		printing.AddTrailingComment(precedingNode, comment)
		return true
	}
	return false
}

// handleCommentAfterArrowParams is upstream's handleCommentAfterArrowParams.
func handleCommentAfterArrowParams(context *commentContext) bool {
	comment, enclosingNode, text := context.Comment, context.Enclosing, context.Text
	if !enclosingNode.Is("ArrowFunctionExpression") {
		return false
	}

	index := getNextNonSpaceNonCommentCharacterIndex(text, locEnd(comment))
	// `index !== false && text.slice(index, index + 2) === "=>"`; upstream's false is notFound.
	if index != notFound && index >= 0 && index <= len(text) && strings.HasPrefix(text[index:], "=>") {
		printing.AddDanglingComment(enclosingNode, comment, "commentBeforeArrow")
		return true
	}

	return false
}

// isInArgumentOrParameterParentheses is upstream's isInArgumentOrParameterParentheses.
func isInArgumentOrParameterParentheses(node Node, comment Node, options *Options) bool {
	commentStart := locStart(comment)
	nodeEnd := locEnd(node)
	if commentStart >= nodeEnd {
		return false
	}

	commentEnd := locEnd(comment)
	nodeStart := locStart(node)
	if commentEnd <= nodeStart {
		return false
	}

	text := stripComments(options)
	return strings.HasSuffix(estree.TrimEndJavaScript(text[:locStart(comment)]), "(") &&
		strings.HasPrefix(estree.TrimStartJavaScript(text[locEnd(comment):]), ")")
}

// isFlowComponent is upstream's createTypeCheckFunction of that name. Flow only; kept so
// handleCommentInEmptyParens reads against upstream.
func isFlowComponent(node Node) bool {
	return node.Is("ComponentDeclaration", "DeclareComponent", "ComponentTypeAnnotation")
}

// handleCommentInEmptyParens is upstream's handleCommentInEmptyParens.
func handleCommentInEmptyParens(context *commentContext) bool {
	comment, enclosingNode, options := context.Comment, context.Enclosing, context.Options
	if enclosingNode == nil {
		return false
	}

	if isCallLikeExpression(enclosingNode) &&
		len(getCallArguments(enclosingNode)) == 0 &&
		isInArgumentOrParameterParentheses(enclosingNode, comment, options) {
		printing.AddDanglingComment(enclosingNode, comment, "")
		return true
	}

	// Only add dangling comments to fix the case when no params are present,
	// i.e. a function without any argument.

	var functionNode Node
	if isRealFunctionLikeNode(enclosingNode) ||
		isFlowComponent(enclosingNode) ||
		enclosingNode.Is("HookTypeAnnotation") {
		functionNode = enclosingNode
	} else if enclosingNode.Is("MethodDefinition", "TSAbstractMethodDefinition") ||
		enclosingNode.Is("Property") && isMethod(enclosingNode) {
		functionNode = enclosingNode.Child("value")
	}

	if functionNode != nil &&
		len(getFunctionParameters(functionNode)) == 0 &&
		isInArgumentOrParameterParentheses(functionNode, comment, options) {
		printing.AddDanglingComment(functionNode, comment, "")
		return true
	}

	return false
}

// handleLastFunctionParameterComments is upstream's handleLastFunctionParameterComments. The Flow
// branches stay as type tests, which never match a TypeScript tree.
func handleLastFunctionParameterComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode, text := context.Comment, context.Preceding,
		context.Enclosing, context.Following, context.Text
	// Flow function type definitions
	if precedingNode.Is("FunctionTypeParam") &&
		enclosingNode.Is("FunctionTypeAnnotation") &&
		!followingNode.Is("FunctionTypeParam") {
		printing.AddTrailingComment(precedingNode, comment)
		return true
	}

	// "DeclareComponent" and "ComponentTypeAnnotation" definitions
	if precedingNode.Is("ComponentTypeParameter") &&
		enclosingNode.Is("DeclareComponent", "ComponentTypeAnnotation") &&
		!followingNode.Is("ComponentTypeParameter") {
		printing.AddTrailingComment(precedingNode, comment)
		return true
	}

	// Real functions and TypeScript function type definitions
	if precedingNode.Is("Identifier", "AssignmentPattern", "ObjectPattern", "ArrayPattern", "RestElement",
		"TSParameterProperty") &&
		(isRealFunctionLikeNode(enclosingNode) ||
			// `TSEmptyBodyFunctionExpression` opts out of comment attachment, so
			// the comment walker bubbles up to its wrapper. Three wrappers occur:
			// `TSAbstractMethodDefinition` (always), and `MethodDefinition` in
			// `declare class` or overload position. A plain `MethodDefinition` with
			// a body is a normal method and must not match here, so the
			// `MethodDefinition` branch checks `value.type` to skip it.
			enclosingNode.Is("TSAbstractMethodDefinition", "MethodDefinition") &&
				enclosingNode.Child("value").Is("TSEmptyBodyFunctionExpression")) &&
		getNextNonSpaceNonCommentCharacter(text, locEnd(comment)) == ")" {
		printing.AddTrailingComment(precedingNode, comment)
		return true
	}

	// "ComponentParameter" definitions
	if precedingNode.Is("ComponentParameter", "RestElement") &&
		enclosingNode.Is("ComponentDeclaration", "DeclareComponent") &&
		getNextNonSpaceNonCommentCharacter(text, locEnd(comment)) == ")" {
		printing.AddTrailingComment(precedingNode, comment)
		return true
	}

	// Comment between function parameters parentheses and function body
	if !isBlockComment(comment) &&
		followingNode.Is("BlockStatement") &&
		isFunctionLikeNode(enclosingNode) {
		functionBody := enclosingNode.Child("body")
		if enclosingNode.Is("MethodDefinition") {
			functionBody = enclosingNode.Child("value").Child("body")
		}

		if functionBody == followingNode {
			characterAfterCommentIndex := getNextNonSpaceNonCommentCharacterIndex(text, locEnd(comment))
			if characterAfterCommentIndex == locStart(followingNode) {
				addBlockStatementFirstComment(followingNode, comment)
				return true
			}
		}
	}

	return false
}

// handleLabeledStatementComments is upstream's handleLabeledStatementComments.
func handleLabeledStatementComments(context *commentContext) bool {
	comment, enclosingNode := context.Comment, context.Enclosing
	if enclosingNode.Is("LabeledStatement") {
		printing.AddLeadingComment(enclosingNode, comment)
		return true
	}
	return false
}

// handleCallExpressionComments is upstream's handleCallExpressionComments.
func handleCallExpressionComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, options := context.Comment, context.Preceding, context.Enclosing,
		context.Options
	if isCallOrNewExpression(enclosingNode) &&
		enclosingNode.Child("callee") == precedingNode &&
		len(enclosingNode.List("arguments")) > 0 &&
		isInsideCallOrNewExpressionParentheses(enclosingNode, comment, options) {
		printing.AddLeadingComment(enclosingNode.List("arguments")[0], comment)
		return true
	}
	return false
}

// handleUnionTypeComments is upstream's handleUnionTypeComments. `node.prettierIgnore` and
// `comment.unignore` are properties set on the nodes, read back by hasNodeIgnoreComment and
// isPrettierIgnoreComment. Upstream throws when followingNode is undefined; Set on a nil node panics
// the same way.
func handleUnionTypeComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode := context.Comment, context.Preceding, context.Enclosing,
		context.Following
	if isUnionType(enclosingNode) {
		if isPrettierIgnoreComment(comment) {
			followingNode.Set("prettierIgnore", true)
			comment.Set("unignore", true)
		}
		if precedingNode != nil {
			printing.AddTrailingComment(precedingNode, comment)
			return true
		}
		return false
	}

	if isUnionType(followingNode) && isPrettierIgnoreComment(comment) {
		followingNode.List("types")[0].Set("prettierIgnore", true)
		comment.Set("unignore", true)
	}

	return false
}

// handleMatchOrPatternComments is upstream's handleMatchOrPatternComments. MatchOrPattern is Flow
// only, so this never acts on a TypeScript tree; it is kept so the handler lists read against upstream.
func handleMatchOrPatternComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, followingNode := context.Comment, context.Preceding, context.Enclosing,
		context.Following
	if enclosingNode != nil && enclosingNode.Is("MatchOrPattern") {
		if isPrettierIgnoreComment(comment) {
			followingNode.Set("prettierIgnore", true)
			comment.Set("unignore", true)
		}
		if precedingNode != nil {
			printing.AddTrailingComment(precedingNode, comment)
			return true
		}
		return false
	}

	if followingNode != nil &&
		followingNode.Is("MatchOrPattern") &&
		isPrettierIgnoreComment(comment) {
		followingNode.List("types")[0].Set("prettierIgnore", true)
		comment.Set("unignore", true)
	}

	return false
}

// handlePropertyComments is upstream's handlePropertyComments.
func handlePropertyComments(context *commentContext) bool {
	comment, enclosingNode := context.Comment, context.Enclosing
	if isObjectProperty(enclosingNode) {
		printing.AddLeadingComment(enclosingNode, comment)
		return true
	}
	return false
}

// handleOnlyComments is upstream's handleOnlyComments.
func handleOnlyComments(context *commentContext) bool {
	comment, enclosingNode, ast, isLastComment := context.Comment, context.Enclosing, context.Ast,
		context.IsLastComment
	// With Flow the enclosingNode is undefined so use the AST instead.
	// `ast?.body?.length === 0`: body must be present and an array.
	if body, isList := ast.Get("body").([]Node); isList && len(body) == 0 {
		if isLastComment {
			printing.AddDanglingComment(ast, comment, "")
		} else {
			printing.AddLeadingComment(ast, comment)
		}
		return true
	}

	if enclosingNode.Is("Program") &&
		len(enclosingNode.List("body")) == 0 &&
		len(enclosingNode.List("directives")) == 0 {
		if isLastComment {
			printing.AddDanglingComment(enclosingNode, comment, "")
		} else {
			printing.AddLeadingComment(enclosingNode, comment)
		}
		return true
	}

	return false
}

// handleModuleSpecifiersComments is upstream's handleModuleSpecifiersComments.
func handleModuleSpecifiersComments(context *commentContext) bool {
	comment, precedingNode, enclosingNode, text := context.Comment, context.Preceding, context.Enclosing, context.Text
	if enclosingNode.Is("ImportSpecifier", "ExportSpecifier") {
		printing.AddLeadingComment(enclosingNode, comment)
		return true
	}

	isImportDeclaration := precedingNode.Is("ImportSpecifier") && enclosingNode.Is("ImportDeclaration")
	isExportDeclaration := precedingNode.Is("ExportSpecifier") && enclosingNode.Is("ExportNamedDeclaration")
	if (isImportDeclaration || isExportDeclaration) &&
		hasNewline(text, locEnd(comment)) {
		printing.AddTrailingComment(precedingNode, comment)
		return true
	}
	return false
}

// handleAssignmentPatternComments is upstream's handleAssignmentPatternComments.
func handleAssignmentPatternComments(context *commentContext) bool {
	comment, enclosingNode := context.Comment, context.Enclosing
	if enclosingNode.Is("AssignmentPattern") {
		printing.AddLeadingComment(enclosingNode, comment)
		return true
	}
	return false
}
