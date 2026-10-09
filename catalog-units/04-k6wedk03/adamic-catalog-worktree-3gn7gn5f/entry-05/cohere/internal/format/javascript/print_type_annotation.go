package javascript

// print/type-annotation.js.

// shouldHugType is upstream's shouldHugType.
func shouldHugType(node Node) bool {
	if isSimpleType(node) || isObjectType(node) {
		return true
	}

	if isUnionType(node) {
		return shouldHugUnionType(node)
	}

	return false
}

/*
Normally the `(TS)TypeAnnotation` node starts with `:` token.
If we print `:` in parent node, `cursorNodeDiff` in `/src/main/core.js` will consider `:` is removed, cause cursor moves, see #12491.
Token *before* `(TS)TypeAnnotation.typeAnnotation` should be printed in `getTypeAnnotationFirstToken` function.
*/
// printTypeAnnotationProperty is upstream's printTypeAnnotationProperty(path, print, propertyName =
// "typeAnnotation"). Go has no default parameters, so the property name is an optional trailing
// argument. Upstream's typeAnnotationNodesCheckedLeadingComments WeakSet only feeds a development
// assertion in printTypeAnnotation, so it is not ported.
func printTypeAnnotationProperty(path *Path, print PrintFunc, property ...string) Doc {
	propertyName := "typeAnnotation"
	if len(property) > 0 {
		propertyName = property[0]
	}
	typeAnnotation := node(path).Child(propertyName)

	if typeAnnotation == nil {
		return emptyDoc
	}

	shouldPrintLeadingSpace := false

	if typeAnnotation.Is("TSTypeAnnotation", "TypeAnnotation") {
		firstToken := call(path, getTypeAnnotationFirstToken, propertyName)

		if firstToken == "=>" ||
			(firstToken == ":" &&
				hasComment(typeAnnotation, commentLeading, nil)) {
			shouldPrintLeadingSpace = true
		}
	}

	if shouldPrintLeadingSpace {
		return concatIn(path, " ", print(propertyName, nil))
	}
	return print(propertyName, nil)
}

// getTypeAnnotationFirstToken is upstream's getTypeAnnotationFirstToken. The three Flow matches
// (DeclareFunction, DeclareHook, a TypeParameter's usesExtendsBound) never match a TypeScript tree and
// are dropped.
func getTypeAnnotationFirstToken(path *Path) string {
	if
	// TypeScript
	path.Match(
		func(value any, _ any, _ int, _ bool) bool {
			current, _ := value.(Node)
			return current.Is("TSTypeAnnotation")
		},
		func(value any, name any, _ int, _ bool) bool {
			current, _ := value.(Node)
			key, _ := name.(string)
			return (key == "returnType" || key == "typeAnnotation") &&
				(current.Is("TSFunctionType") || current.Is("TSConstructorType"))
		},
	) {
		return "=>"
	}

	if
	// TypeScript
	path.Match(
		func(value any, _ any, _ int, _ bool) bool {
			current, _ := value.(Node)
			return current.Is("TSTypeAnnotation")
		},
		func(value any, name any, _ int, _ bool) bool {
			current, _ := value.(Node)
			key, _ := name.(string)
			return key == "typeAnnotation" &&
				(current.Is("TSJSDocNullableType") ||
					current.Is("TSJSDocNonNullableType") ||
					current.Is("TSTypePredicate"))
		},
	) {
		return ""
	}

	return ":"
}

/*
- `TSTypeAnnotation` (TypeScript)
- `TypeAnnotation` (Flow)
*/
// printTypeAnnotation is upstream's printTypeAnnotation. We need print space before leading comments,
// `printTypeAnnotationProperty` is responsible for it. Upstream's development-only check that the node
// went through printTypeAnnotationProperty is not ported.
func printTypeAnnotation(path *Path, options *Options, print PrintFunc) Doc {
	token := getTypeAnnotationFirstToken(path)
	if token != "" {
		return concatIn(path, token, " ", print("typeAnnotation", nil))
	}
	return print("typeAnnotation", nil)
}
