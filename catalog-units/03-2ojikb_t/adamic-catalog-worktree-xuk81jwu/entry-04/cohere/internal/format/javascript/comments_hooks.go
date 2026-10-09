package javascript

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/estree"
)

// The comment hooks the print core calls: print/comment.js, comments/can-attach-comment.js and
// comments/will-print-own-comments.js. comments/is-gap.js only answers for Flow parsers, so the core's
// default gap test applies and no hook is set.

// printComment is upstream's printComment, print/comment.js.
func printComment(path *Path, options *Options) Doc {
	comment := node(path)
	if isLineComment(comment) {
		return concatIn(path, estree.TrimEndJavaScript(options.OriginalText[locStart(comment):locEnd(comment)]))
	}
	if estree.IsIndentableBlockComment(comment) {
		return printIndentableBlockComment(comment)
	}
	if isBlockComment(comment) {
		return concatIn(path, "/*", replaceEndOfLine(comment.String("value")), "*/")
	}
	panic("not a comment: " + comment.Type())
}

// replaceEndOfLine is upstream's replaceEndOfLine(text) for a string: lines joined by literalline.
func replaceEndOfLine(text string) Doc {
	lines := strings.Split(text, "\n")
	parts := make([]Doc, len(lines))
	for index, line := range lines {
		parts[index] = concat(line)
	}
	return join(literalline, parts)
}

// printIndentableBlockComment is upstream's printIndentableBlockComment.
func printIndentableBlockComment(comment Node) Doc {
	lines := estree.IndentableBlockCommentLines(comment)
	value := comment.String("value")
	isJsdoc := len(value) > 0 && value[0] == '*' && (len(value) < 2 || value[1] != '*')
	parts := make([]any, len(lines))
	for index, line := range lines {
		switch {
		case index == 0:
			parts[index] = concat(estree.TrimEndJavaScript(line), hardline)
		case index == len(lines)-1:
			parts[index] = concat(" ", line)
		default:
			trimmed := estree.TrimEndJavaScript(line)
			content := concat(" ", trimmed)
			if isJsdoc && trimmed != "*" && strings.HasSuffix(line, "  ") {
				parts[index] = concat(content, "  ", markAsRoot(literalline))
			} else {
				parts[index] = concat(content, hardline)
			}
		}
	}
	return concat("/", concat(parts...), "/")
}

// isNodeCantAttachComment is upstream's isNodeCantAttachComment.
func isNodeCantAttachComment(node Node) bool {
	return node.Is("File", "TemplateElement", "TSEmptyBodyFunctionExpression", "ChainExpression")
}

// isChildWontPrint is upstream's isChildWontPrint.
func isChildWontPrint(node Node, parent Node) bool {
	return parent.Is("ObjectProperty") && parent.Bool("shorthand") && parent.Child("key") == node &&
		parent.Child("value") != parent.Child("key") ||
		parent.Is("Property") && parent.Bool("shorthand") && parent.Child("key") == node && !isMethod(parent) &&
			parent.Child("value") != parent.Child("key") ||
		parent.Is("ImportSpecifier") && isShorthandSpecifier(parent) && parent.Child("local") == node &&
			parent.Child("local") != parent.Child("imported") ||
		parent.Is("ExportSpecifier") && isShorthandSpecifier(parent) && parent.Child("exported") == node &&
			parent.Child("local") != parent.Child("exported")
}

// isClassMethodCantAttachComment is upstream's isClassMethodCantAttachComment.
//
// Upstream also tests `!isNonEmptyArray(node.typeParameters)`, but typeParameters is a
// TSTypeParameterDeclaration object, never an array, so that clause is always true and is omitted.
func isClassMethodCantAttachComment(node Node, parent Node) bool {
	return node.Is("FunctionExpression") && parent.Is("MethodDefinition") && parent.Child("value") == node &&
		len(node.List("params")) == 0 && node.Child("returnType") == nil && node.Child("body") != nil
}

// isTsAsConstTypeReference is upstream's isTsAsConstTypeReference.
func isTsAsConstTypeReference(node Node, parent Node) bool {
	return parent != nil && parent.Child("typeAnnotation") == node && isTsAsConstExpression(parent)
}

// isTsAsConst is upstream's isTsAsConst.
func isTsAsConst(node Node, ancestors []Node) bool {
	var parent, grandparent Node
	if len(ancestors) > 0 {
		parent = ancestors[0]
	}
	if len(ancestors) > 1 {
		grandparent = ancestors[1]
	}
	return isTsAsConstTypeReference(node, parent) ||
		parent != nil && parent.Child("typeName") == node && isTsAsConstTypeReference(parent, grandparent)
}

// canAttachComment is upstream's canAttachComment, comments/can-attach-comment.js.
func canAttachComment(node Node, ancestors []Node) bool {
	var parent Node
	if len(ancestors) > 0 {
		parent = ancestors[0]
	}
	if isNodeCantAttachComment(node) || isChildWontPrint(node, parent) || isClassMethodCantAttachComment(node, parent) {
		return false
	}
	if node.Is("EmptyStatement") {
		return isMeaningfulEmptyStatementNode(node, parent)
	}
	if isTsAsConst(node, ancestors) {
		return false
	}
	if node.Is("TSTypeAnnotation") && parent.Is("TSPropertySignature") {
		return false
	}
	return true
}

// isClassOrInterface is upstream's isClassOrInterface.
func isClassOrInterface(node Node) bool {
	return node.Is("ClassDeclaration", "ClassExpression", "DeclareClass", "DeclareInterface", "InterfaceDeclaration",
		"TSInterfaceDeclaration")
}

// willPrintOwnComments is upstream's willPrintOwnComments, comments/will-print-own-comments.js.
func willPrintOwnComments(path *Path, options *Options) bool {
	key := keyOf(path)
	parent := parentOf(path)
	if key == "types" && isUnionType(parent) ||
		key == "argument" && parent.Is("JSXSpreadAttribute") ||
		key == "expression" && parent.Is("JSXSpreadChild") ||
		key == "superClass" && parent.Is("ClassDeclaration", "ClassExpression") ||
		(key == "id" || key == "typeParameters") && isClassOrInterface(parent) ||
		key == "patterns" && parent.Is("MatchOrPattern") ||
		isIifeCalleeOrTaggedTemplateExpressionTag(path) {
		return true
	}
	current := node(path)
	if hasNodeIgnoreComment(current) {
		return false
	}
	if current.Is("ExpressionStatement") {
		return shouldExpressionStatementPrintOwnComments(path, options)
	}
	if isUnionType(current) {
		return shouldUnionTypePrintOwnComments(path)
	}
	return isJsxElement(current)
}
