package javascript

// print/member.js.

// isCallExpressionWithArguments is upstream's isCallExpressionWithArguments.
func isCallExpressionWithArguments(node Node) bool {
	return isCallExpression(node) && len(getCallArguments(node)) > 0
}

// shouldInlineNewExpressionCallee is upstream's shouldInlineNewExpressionCallee.
func shouldInlineNewExpressionCallee(path *Path) bool {
	child := node(path)
	for _, ancestor := range path.Ancestors() {
		if !(isMemberExpression(ancestor) && ancestor.Child("object") == child ||
			ancestor.Is("TSNonNullExpression") && ancestor.Child("expression") == child) {
			return ancestor.Is("NewExpression") && ancestor.Child("callee") == child
		}

		child = ancestor
	}

	return false
}

// printMemberExpression is upstream's printMemberExpression.
func printMemberExpression(path *Path, options *Options, print PrintFunc) Doc {
	objectDoc := print("object", nil)
	lookupDoc := printMemberLookup(path, options, print)
	current := node(path)
	firstNonMemberParent, _ := path.FindAncestor(func(node Node) bool {
		return !(isMemberExpression(node) || node.Is("TSNonNullExpression"))
	})
	firstNonChainElementWrapperParent, _ := path.FindAncestor(func(node Node) bool {
		return !isChainElementWrapper(node)
	})

	// upstream's `objectDoc.label?.memberChain`: Go labels are strings, and printMemberChain labels
	// its result "member-chain".
	shouldInline := firstNonMemberParent.Is("BindExpression") ||
		(firstNonMemberParent.Is("AssignmentExpression") &&
			!firstNonMemberParent.Child("left").Is("Identifier")) ||
		shouldInlineNewExpressionCallee(path) ||
		current.Bool("computed") ||
		(current.Child("object").Is("Identifier") &&
			current.Child("property").Is("Identifier") &&
			!isMemberExpression(firstNonChainElementWrapperParent)) ||
		(firstNonChainElementWrapperParent.Is("AssignmentExpression", "VariableDeclarator") &&
			(isCallExpressionWithArguments(stripChainElementWrappers(current.Child("object"))) ||
				labelOf(objectDoc) == "member-chain"))

	var lookup Doc
	if shouldInline {
		lookup = lookupDoc
	} else {
		lookup = groupIn(path, indentIn(path, concatIn(path, softline, lookupDoc)))
	}
	return label(labelOf(objectDoc), concatIn(path, objectDoc,
		lineSuffixBoundary,
		lookup,
	))
}

// printMemberLookup is upstream's printMemberLookup.
func printMemberLookup(path *Path, options *Options, print PrintFunc) Doc {
	property := print("property", nil)
	current := node(path)
	optional := printOptionalToken(path)

	if !current.Bool("computed") {
		return concatIn(path, optional, ".", property)
	}

	if current.Child("property") == nil || isNumericLiteral(current.Child("property")) {
		return concatIn(path, optional, "[", property, "]")
	}

	return groupIn(path, concatIn(path, optional, "[", indentIn(path, concatIn(path, softline, property)), softline, "]"))
}
