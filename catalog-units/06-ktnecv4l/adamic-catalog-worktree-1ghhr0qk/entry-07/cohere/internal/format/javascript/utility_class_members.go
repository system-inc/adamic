package javascript

// utilities/class-members.js.

// iterateClassMembersPath is upstream's iterateClassMembersPath. Upstream's iteratee reads only the
// path (node, next, isLast), never the index, so it takes the path alone. The ObjectTypeAnnotation
// branch, which reorders Flow's interleaved properties, indexers, call properties and internal
// slots and hands the iteratee a path-shaped object, is dropped: only the Flow parsers produce it.
func iterateClassMembersPath(path *Path, iteratee func(path *Path)) {
	node := node(path)
	if node.Is("ClassBody", "TSInterfaceBody") {
		each(path, func(path *Path, _ int) { iteratee(path) }, "body")
		return
	}

	if node.Is("TSTypeLiteral") {
		each(path, func(path *Path, _ int) { iteratee(path) }, "members")
		return
	}

	if node.Is("RecordDeclarationBody") {
		each(path, func(path *Path, _ int) { iteratee(path) }, "elements")
		return
	}
}

// isNonEmptyClassBody is upstream's isNonEmptyClassBody. The ObjectTypeAnnotation branch is dropped
// with the Flow parsers that produce it.
func isNonEmptyClassBody(node Node) bool {
	members := node.List("body")
	if node.Is("RecordDeclarationBody") {
		members = node.List("elements")
	}

	return len(members) > 0
}
