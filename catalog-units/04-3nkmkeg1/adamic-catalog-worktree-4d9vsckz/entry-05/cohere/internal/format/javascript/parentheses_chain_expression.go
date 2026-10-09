package javascript

// parentheses/chain-expression.js.

func shouldAddParenthesesToChainExpression(path *Path) bool {
	key := keyOf(path)
	parent := parentOf(path)

	return key == "expression" && parent.Is("TSNonNullExpression") ||
		key == "object" &&
			parent.Is("MemberExpression") &&
			!parent.Truthy("optional") ||
		key == "callee" &&
			parent.Is("CallExpression") &&
			!parent.Truthy("optional") ||
		key == "callee" && parent.Is("NewExpression") ||
		key == "tag" && parent.Is("TaggedTemplateExpression")
}

// isParenthesized is upstream's file-local `(node) => node.extra?.parenthesized`, nil-safe as the
// optional chain is.
func isParenthesized(node Node) bool {
	return node != nil && node.Parenthesized
}

// isBabelOptionalChainElement is upstream's createTypeCheckFunction over the Babel optional chain
// element types.
func isBabelOptionalChainElement(node Node) bool {
	return node.Is("OptionalCallExpression", "OptionalMemberExpression")
}

func isBabelOptionalChainRoot(path *Path) bool {
	current := node(path)

	child := current
	for child.Is("TSNonNullExpression") {
		child = child.Child("expression")

		if isParenthesized(child) {
			return false
		}
	}

	if !isBabelOptionalChainElement(child) {
		return false
	}

	if isParenthesized(current) {
		return true
	}

	return !(keyOf(path) == "expression" && parentOf(path).Is("TSNonNullExpression"))
	// We should exclude if it's in `OptionalCallExpression` or `OptionalMemberExpression`
	// But in these cases it should not matter, since we don't need parentheses anyway
}

func shouldAddParenthesesToChainElement(path *Path) bool {
	return (node(path).Is("ChainExpression") || isBabelOptionalChainRoot(path)) &&
		shouldAddParenthesesToChainExpression(path)
}
