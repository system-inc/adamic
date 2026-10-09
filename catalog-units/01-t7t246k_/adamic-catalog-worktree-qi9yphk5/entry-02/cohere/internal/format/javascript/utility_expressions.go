package javascript

import "github.com/system-inc/cohere/internal/format/estree"

// utilities/left-side.js, utilities/starts-with-no-lookahead-token.js,
// utilities/strip-chain-element-wrappers.js, utilities/is-nullish-coalescing.js,
// utilities/is-signed-numeric-literal.js, utilities/is-simple-expression-by-node-count.js,
// utilities/has-node.js (with hasDescendant from src/utilities/ast.js),
// utilities/is-meaningful-empty-statement.js, utilities/is-shorthand-specifier.js.

// hasNakedLeftSide is upstream's hasNakedLeftSide.
func hasNakedLeftSide(node Node) bool {
	return node.Is("AssignmentExpression") ||
		node.Is("BinaryExpression") ||
		node.Is("LogicalExpression") ||
		node.Is("NGPipeExpression") ||
		node.Is("ConditionalExpression") ||
		isCallExpression(node) ||
		isMemberExpression(node) ||
		node.Is("SequenceExpression") ||
		node.Is("TaggedTemplateExpression") ||
		node.Is("BindExpression") ||
		node.Is("UpdateExpression") && !node.Truthy("prefix") ||
		isBinaryCastExpression(node) ||
		isChainElementWrapper(node)
}

// getLeftSide is upstream's getLeftSide. `node.expressions` is truthy for any array, even an empty
// one, whose first element is then undefined: nil here.
func getLeftSide(node Node) Node {
	if node.Truthy("expressions") {
		expressions := node.List("expressions")
		if len(expressions) == 0 {
			return nil
		}
		return expressions[0]
	}
	for _, key := range []string{"left", "test", "callee", "object", "tag", "argument", "expression"} {
		if side := node.Child(key); side != nil {
			return side
		}
	}
	return nil
}

// getLeftSidePathName is upstream's getLeftSidePathName. Note its order differs from getLeftSide's:
// object before callee, as upstream has it.
func getLeftSidePathName(node Node) []any {
	if node.Truthy("expressions") {
		return []any{"expressions", 0}
	}
	if node.Truthy("left") {
		return []any{"left"}
	}
	if node.Truthy("test") {
		return []any{"test"}
	}
	if node.Truthy("object") {
		return []any{"object"}
	}
	if node.Truthy("callee") {
		return []any{"callee"}
	}
	if node.Truthy("tag") {
		return []any{"tag"}
	}
	if node.Truthy("argument") {
		return []any{"argument"}
	}
	if node.Truthy("expression") {
		return []any{"expression"}
	}
	panic("javascript: Unexpected node has no left side.")
}

// startsWithNoLookaheadToken is upstream's startsWithNoLookaheadToken.
//
// Tests if the leftmost node of the expression matches the predicate. E.g.,
// used to check whether an expression statement needs to be wrapped in extra
// parentheses because it starts with:
//
//   - `{`
//   - `function`, `class`, or `do {}`
//   - `let[`
//
// Will be overzealous if there already are necessary grouping parentheses.
func startsWithNoLookaheadToken(node Node, predicate func(leftmostNode Node) bool) bool {
	switch node.Type() {
	case "BinaryExpression",
		"LogicalExpression",
		"AssignmentExpression",
		"NGPipeExpression":
		return startsWithNoLookaheadToken(node.Child("left"), predicate)
	case "MemberExpression",
		"OptionalMemberExpression":
		return startsWithNoLookaheadToken(node.Child("object"), predicate)
	case "TaggedTemplateExpression":
		if node.Child("tag").Is("FunctionExpression") {
			// IIFEs are always already parenthesized
			return false
		}
		return startsWithNoLookaheadToken(node.Child("tag"), predicate)
	case "CallExpression",
		"OptionalCallExpression":
		if node.Child("callee").Is("FunctionExpression") {
			// IIFEs are always already parenthesized
			return false
		}
		return startsWithNoLookaheadToken(node.Child("callee"), predicate)
	case "ConditionalExpression":
		return startsWithNoLookaheadToken(node.Child("test"), predicate)
	case "UpdateExpression":
		return !node.Truthy("prefix") && startsWithNoLookaheadToken(node.Child("argument"), predicate)
	case "BindExpression":
		return node.Truthy("object") && startsWithNoLookaheadToken(node.Child("object"), predicate)
	case "SequenceExpression":
		return startsWithNoLookaheadToken(node.List("expressions")[0], predicate)
	case "ChainExpression",
		"TSNonNullExpression",
		"TSSatisfiesExpression",
		"TSAsExpression",
		"AsExpression",
		"AsConstExpression",
		"SatisfiesExpression":
		return startsWithNoLookaheadToken(node.Child("expression"), predicate)
	default:
		return predicate(node)
	}
}

// stripChainElementWrappers is upstream's stripChainElementWrappers.
func stripChainElementWrappers(node Node) Node {
	for isChainElementWrapper(node) {
		node = node.Child("expression")
	}

	return node
}

// isNullishCoalescing is upstream's isNullishCoalescing.
func isNullishCoalescing(node Node) bool {
	return node.Is("LogicalExpression") && node.String("operator") == "??"
}

// isSignedNumericLiteral is upstream's isSignedNumericLiteral.
func isSignedNumericLiteral(node Node) bool {
	return node.Is("UnaryExpression") &&
		(node.String("operator") == "+" || node.String("operator") == "-") &&
		isNumericLiteral(node.Child("argument"))
}

// getExpressionInnerNodeCount is upstream's getExpressionInnerNodeCount. Upstream's `for...in`
// counts every property holding an object with a string `type`, which is every child node and never
// an array, so the Go counts the node-valued properties in insertion order.
func getExpressionInnerNodeCount(node Node, maxCount int) int {
	count := 0
	for _, key := range node.Keys() {
		if property, isNode := node.Get(key).(Node); isNode && property != nil {
			count++
			count += getExpressionInnerNodeCount(property, maxCount-count)
		}

		// Bail early to protect against bad performance.
		if count > maxCount {
			return count
		}
	}
	return count
}

// isSimpleExpressionByNodeCount is upstream's isSimpleExpressionByNodeCount. Upstream's default
// maxInnerNodeCount is 5; Go has no default parameters, so callers pass it.
//
// Attempts to gauge the rough complexity of a node, for example
// to detect deeply-nested booleans, call expressions with lots of arguments, etc.
func isSimpleExpressionByNodeCount(node Node, maxInnerNodeCount int) bool {
	count := getExpressionInnerNodeCount(node, maxInnerNodeCount)
	return count <= maxInnerNodeCount
}

// hasNode is upstream's hasNode.
func hasNode(node Node, predicate func(node Node) bool) bool {
	return predicate(node) || hasDescendant(node, predicate)
}

// hasDescendant is upstream's hasDescendant, src/utilities/ast.js, over the JavaScript visitor keys:
// breadth first, testing each child as it is reached.
func hasDescendant(node Node, predicate func(node Node) bool) bool {
	queue := []Node{node}
	for index := 0; index < len(queue); index++ {
		for _, child := range estree.ChildNodes(queue[index]) {
			if predicate(child) {
				return true
			}
			queue = append(queue, child)
		}
	}

	return false
}

// isMeaningfulEmptyStatement is upstream's isMeaningfulEmptyStatement called with a path.
func isMeaningfulEmptyStatement(path *Path) bool {
	return isMeaningfulEmptyStatementNode(node(path), parentOf(path))
}

// isMeaningfulEmptyStatementNode is upstream's isMeaningfulEmptyStatement called with the duck type
// `{ node, parent }` that comment attachment and AST massage pass in place of a path.
func isMeaningfulEmptyStatementNode(node Node, parent Node) bool {
	if !node.Is("EmptyStatement") {
		return false
	}

	if parent.Is("IfStatement") {
		return parent.Child("consequent") == node || parent.Child("alternate") == node
	}

	if parent.Is(
		"DoWhileStatement",
		"ForInStatement",
		"ForOfStatement",
		"ForStatement",
		"LabeledStatement",
		"WithStatement",
		"WhileStatement",
	) {
		return parent.Child("body") == node
	}

	return false
}

// isShorthandSpecifier is upstream's isShorthandSpecifier.
func isShorthandSpecifier(specifier Node) bool {
	if !specifier.Is("ImportSpecifier") &&
		!specifier.Is("ExportSpecifier") {
		return false
	}

	local := specifier.Child("local")
	importedOrExportedKey := "exported"
	if specifier.Is("ImportSpecifier") {
		importedOrExportedKey = "imported"
	}
	importedOrExported := specifier.Child(importedOrExportedKey)

	if local.Type() != importedOrExported.Type() ||
		!estree.HasSameLoc(local, importedOrExported) {
		return false
	}

	if isStringLiteral(local) {
		return local.Get("value") == importedOrExported.Get("value") &&
			getRaw(local) == getRaw(importedOrExported)
	}

	switch local.Type() {
	case "Identifier":
		return local.String("name") == importedOrExported.String("name")
	default:
		return false
	}
}
