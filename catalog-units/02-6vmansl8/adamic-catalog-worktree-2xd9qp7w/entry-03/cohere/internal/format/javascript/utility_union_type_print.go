package javascript

// utilities/union-type-print.js.

func isVoidType(node Node) bool {
	return node.Is(
		"VoidTypeAnnotation",
		"TSVoidKeyword",
		"NullLiteralTypeAnnotation",
		"TSNullKeyword",
	)
}

func isObjectLikeType(node Node) bool {
	return node.Is(
		"ObjectTypeAnnotation",
		"TSTypeLiteral",
		// This is a bit aggressive but captures Array<{x}>
		"GenericTypeAnnotation",
		"TSTypeReference",
	)
}

// isMultipleTupleTypeElement is upstream's isMultipleTupleTypeElement.
func isMultipleTupleTypeElement(path *Path) bool {
	return keyOf(path) == "elementTypes" &&
		isTupleType(parentOf(path)) &&
		len(parentOf(path).List("elementTypes")) > 1
}

// shouldHugUnionType is upstream's shouldHugUnionType.
func shouldHugUnionType(node Node) bool {
	types := node.List("types")
	for _, member := range types {
		if hasAnyComment(member) {
			return false
		}
	}

	var objectType Node
	for _, member := range types {
		if isObjectLikeType(member) {
			objectType = member
			break
		}
	}
	if objectType == nil {
		return false
	}

	for _, member := range types {
		if !(member == objectType || isVoidType(member)) {
			return false
		}
	}
	return true
}

// shouldUnionTypePrintOwnComments is upstream's shouldUnionTypePrintOwnComments.
func shouldUnionTypePrintOwnComments(path *Path) bool {
	key, node, parent := keyOf(path), node(path), parentOf(path)
	if shouldHugUnionType(node) ||
		// If it's `types` of union type, parent will print comment for it
		key == "types" && isUnionType(parent) ||
		// Inside intersection type let the comment print outside of parentheses
		key == "types" && isIntersectionType(parent) ||
		isMultipleTupleTypeElement(path) {
		return false
	}

	return true
}
