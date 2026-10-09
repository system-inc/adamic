package estree

// LocStart is Prettier's locStart for JavaScript, src/language-js/location/start.js: a node with
// decorators starts at its first decorator, even when the parser's range begins after it.
func LocStart(node *Node) int {
	start := node.Range[0]
	var decorators []*Node
	if declaration := node.Child("declaration"); declaration != nil && declaration.Has("decorators") {
		decorators = declaration.List("decorators")
	} else {
		decorators = node.List("decorators")
	}
	if len(decorators) > 0 && decorators[0] != nil {
		return min(LocStart(decorators[0]), start)
	}
	return start
}

// LocEndWithFullText is Prettier's locEndWithFullText: the range end, before any override.
func LocEndWithFullText(node *Node) int {
	return node.Range[1]
}

// nodeTypesWithContentEnd are upstream's nodeTypesWithContentEnd, location/overrides.js.
var nodeTypesWithContentEnd = map[string]bool{
	"ExpressionStatement":      true,
	"Directive":                true,
	"ImportDeclaration":        true,
	"ExportDefaultDeclaration": true,
	"ExportNamedDeclaration":   true,
	"ExportAllDeclaration":     true,
	"ReturnStatement":          true,
	"ThrowStatement":           true,
	"DoWhileStatement":         true,
}

// ShouldAddContentEnd is upstream's shouldAddContentEnd.
func ShouldAddContentEnd(node *Node) bool {
	return nodeTypesWithContentEnd[node.Type()]
}

// LocEnd is Prettier's locEnd for JavaScript, location/overrides.js.
func LocEnd(node *Node) int {
	switch node.Type() {
	case "IfStatement":
		if alternate := node.Child("alternate"); alternate != nil {
			return LocEnd(alternate)
		}
		return LocEnd(node.Child("consequent"))
	case "ForInStatement", "ForOfStatement", "ForStatement", "LabeledStatement", "WithStatement", "WhileStatement":
		return LocEnd(node.Child("body"))
	case "BreakStatement":
		if label := node.Child("label"); label != nil {
			return LocEnd(label)
		}
		return LocStart(node) + len("break")
	case "ContinueStatement":
		if label := node.Child("label"); label != nil {
			return LocEnd(label)
		}
		return LocStart(node) + len("continue")
	case "DebuggerStatement":
		return LocStart(node) + len("debugger")
	case "VariableDeclaration":
		declarations := node.List("declarations")
		return LocEnd(declarations[len(declarations)-1])
	}
	if nodeTypesWithContentEnd[node.Type()] && node.HasContentEnd {
		return node.ContentEnd
	}
	return LocEndWithFullText(node)
}

// ShouldIgnoredNodePrintSemicolon is upstream's shouldIgnoredNodePrintSemicolon.
func ShouldIgnoredNodePrintSemicolon(node *Node) bool {
	if ShouldAddContentEnd(node) && node.HasContentEnd && node.ContentEnd != 0 {
		return true
	}
	switch node.Type() {
	case "BreakStatement", "ContinueStatement", "DebuggerStatement", "VariableDeclaration":
		return true
	case "IfStatement":
		if alternate := node.Child("alternate"); alternate != nil {
			return ShouldIgnoredNodePrintSemicolon(alternate)
		}
		return ShouldIgnoredNodePrintSemicolon(node.Child("consequent"))
	case "ForInStatement", "ForOfStatement", "ForStatement", "LabeledStatement", "WithStatement", "WhileStatement":
		return ShouldIgnoredNodePrintSemicolon(node.Child("body"))
	}
	return false
}

// HasSameLocStart is upstream's hasSameLocStart.
func HasSameLocStart(left *Node, right *Node) bool {
	return LocStart(left) == LocStart(right)
}

// HasSameLoc is upstream's hasSameLoc.
func HasSameLoc(left *Node, right *Node) bool {
	return HasSameLocStart(left, right) && LocEnd(left) == LocEnd(right)
}
