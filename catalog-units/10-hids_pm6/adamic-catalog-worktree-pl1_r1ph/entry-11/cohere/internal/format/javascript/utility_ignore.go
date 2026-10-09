package javascript

// utilities/has-node-ignore-comment.js, utilities/has-jsx-ignore-comment.js, utilities/is-ignored.js.

// hasNodeIgnoreComment is upstream's hasNodeIgnoreComment. `node.prettierIgnore` is the flag
// handle-comments sets on a union member or mapped type a prettier-ignore comment applies to.
func hasNodeIgnoreComment(node Node) bool {
	return node.Truthy("prettierIgnore") || hasComment(node, commentPrettierIgnore, nil)
}

// hasJsxIgnoreComment is upstream's hasJsxIgnoreComment.
func hasJsxIgnoreComment(path *Path) bool {
	node, parent := node(path), parentOf(path)
	if !isJsxElement(node) || !isJsxElement(parent) {
		return false
	}

	// Lookup the previous sibling, ignoring any empty JSXText elements
	index, siblings := indexOf(path), siblingsOf(path)
	var prevSibling Node
	for candidateIndex := index; candidateIndex > 0; candidateIndex-- {
		candidate := siblings[candidateIndex-1]
		if candidate.Is("JSXText") && !isMeaningfulJsxText(candidate) {
			continue
		}
		prevSibling = candidate
		break
	}

	return prevSibling.Is("JSXExpressionContainer") &&
		prevSibling.Child("expression").Is("JSXEmptyExpression") &&
		hasNodeIgnoreComment(prevSibling.Child("expression"))
}

// isIgnored is upstream's isIgnored.
func isIgnored(path *Path) bool {
	return hasNodeIgnoreComment(node(path)) || hasJsxIgnoreComment(path)
}
