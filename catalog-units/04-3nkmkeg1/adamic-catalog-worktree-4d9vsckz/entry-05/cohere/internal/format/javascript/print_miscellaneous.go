package javascript

import "github.com/system-inc/cohere/internal/format/printing"

// print/miscellaneous.js.

// printOptionalToken is upstream's printOptionalToken.
func printOptionalToken(path *Path) Doc {
	current := node(path)
	if !current.Bool("optional") || current.Is("Identifier") && current == parentOf(path).Child("key") {
		return emptyDoc
	}
	if isCallExpression(current) || isMemberExpression(current) && current.Bool("computed") ||
		current.Is("OptionalIndexedAccessType") {
		return concatIn(path, "?.")
	}
	return concatIn(path, "?")
}

// printDefiniteToken is upstream's printDefiniteToken.
func printDefiniteToken(path *Path) Doc {
	if node(path).Bool("definite") || path.Match(nil, func(value any, name any, _ int, _ bool) bool {
		parent, _ := value.(Node)
		key, _ := name.(string)
		return key == "id" && parent.Is("VariableDeclarator") && parent.Bool("definite")
	}) {
		return concatIn(path, "!")
	}
	return emptyDoc
}

// isFlowDeclareNode is upstream's isFlowDeclareNode.
func isFlowDeclareNode(node Node) bool {
	return node.Is("DeclareClass", "DeclareComponent", "DeclareFunction", "DeclareHook", "DeclareVariable",
		"DeclareExportDeclaration", "DeclareExportAllDeclaration", "DeclareOpaqueType", "DeclareTypeAlias",
		"DeclareEnum", "DeclareInterface")
}

// shouldPrintDeclareToken is upstream's shouldPrintDeclareToken.
func shouldPrintDeclareToken(path *Path) bool {
	current := node(path)
	if isFlowDeclareNode(current) {
		return !parentOf(path).Is("DeclareExportDeclaration") && !current.Truthy("implicitDeclare")
	}
	return current.Truthy("declare")
}

// printDeclareToken is upstream's printDeclareToken.
func printDeclareToken(path *Path) Doc {
	if shouldPrintDeclareToken(path) {
		return concatIn(path, "declare ")
	}
	return emptyDoc
}

// isTsAbstractNode is upstream's isTsAbstractNode.
func isTsAbstractNode(node Node) bool {
	return node.Is("TSAbstractMethodDefinition", "TSAbstractPropertyDefinition", "TSAbstractAccessorProperty")
}

// printAbstractToken is upstream's printAbstractToken.
func printAbstractToken(path *Path) Doc {
	current := node(path)
	if current.Truthy("abstract") || isTsAbstractNode(current) {
		return concatIn(path, "abstract ")
	}
	return emptyDoc
}

// printTypeScriptAccessibilityToken is upstream's printTypeScriptAccessibilityToken.
func printTypeScriptAccessibilityToken(node Node) Doc {
	if accessibility := node.String("accessibility"); accessibility != "" {
		return concat(accessibility + " ")
	}
	return emptyDoc
}

// isLogicalNot is upstream's isLogicalNot.
func isLogicalNot(node Node) bool {
	return node.Is("UnaryExpression") && node.String("operator") == "!"
}

// shouldInlineCondition is upstream's shouldInlineCondition: `!(a || b)` and `!!(a || b)`.
func shouldInlineCondition(node Node) bool {
	if hasAnyComment(node) {
		return false
	}
	if !isLogicalNot(node) {
		return false
	}
	node = node.Child("argument")
	if isLogicalNot(node) {
		node = node.Child("argument")
	}
	return node.Is("LogicalExpression")
}

// printIfOrWhileConditionOrWithStatementObject is upstream's function of that name, exported upstream
// as printIfStatementCondition, printWhileStatementCondition and printDoWhileStatementCondition.
func printIfOrWhileConditionOrWithStatementObject(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	property := "test"
	if current.Is("WithStatement") {
		property = "object"
	}
	conditionDoc := print(property, nil)
	if shouldInlineCondition(current.Child(property)) {
		return conditionDoc
	}
	return groupIn(path, concatIn(path, indentIn(path, concatIn(path, softline, conditionDoc)), softline))
}

func printIfStatementCondition(path *Path, options *Options, print PrintFunc) Doc {
	return printIfOrWhileConditionOrWithStatementObject(path, options, print)
}

func printWhileStatementCondition(path *Path, options *Options, print PrintFunc) Doc {
	return printIfOrWhileConditionOrWithStatementObject(path, options, print)
}

func printDoWhileStatementCondition(path *Path, options *Options, print PrintFunc) Doc {
	return printIfOrWhileConditionOrWithStatementObject(path, options, print)
}

// printDanglingCommentsInList is upstream's printDanglingCommentsInList. A nil filter includes all.
func printDanglingCommentsInList(path *Path, options *Options, filter func(Node) bool) Doc {
	current := node(path)
	if !hasComment(current, commentDangling, filter) {
		return emptyDoc
	}
	closing := softline
	if hasComment(current, commentDangling|commentLine, filter) {
		closing = hardline
	}
	return concatIn(path, indentIn(path, concatIn(path, softline, printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{Filter: filter}))),
		closing,
	)
}

// printTrailingComma is upstream's printTrailingComma(options, level). level "" is upstream's default
// "es5".
func printTrailingComma(options *Options, level string) Doc {
	if level == "" {
		level = "es5"
	}
	if shouldPrintTrailingComma(options, level) {
		return ifBreak(",", "")
	}
	return emptyDoc
}

// printSemicolon is upstream's printSemicolon.
func printSemicolon(options *Options) Doc {
	if settingsOf(options).Semi {
		return concat(";")
	}
	return emptyDoc
}
