package javascript

// print/binary-cast-expression.js.

// printBinaryCastExpression is upstream's printBinaryCastExpression. Upstream's isFlowAsConstExpression
// branch (Flow's AsConstExpression prints "const") is Flow-only; TSAsExpression and
// TSSatisfiesExpression always print their typeAnnotation.
func printBinaryCastExpression(path *Path, options *Options, print PrintFunc) Doc {
	parent := parentOf(path)
	current := node(path)
	key := keyOf(path)
	typeAnnotationDoc := print("typeAnnotation", nil)

	keyword := "as"
	if isSatisfiesExpression(current) {
		keyword = "satisfies"
	}
	parts := []any{
		print("expression", nil),
		" ",
		keyword,
		" ",
		typeAnnotationDoc,
	}

	if key == "callee" && isCallOrNewExpression(parent) ||
		key == "object" && isMemberExpression(parent) {
		return groupIn(path, concatIn(path, indentIn(path, concatIn(path, append([]any{softline}, parts...)...)), softline))
	}

	return concatIn(path, parts...)
}
