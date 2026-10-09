package javascript

// print/function-type.js.

/*
- `TSFunctionType` (TypeScript)
- `TSCallSignatureDeclaration` (TypeScript)
- `TSConstructorType` (TypeScript)
- `TSConstructSignatureDeclaration` (TypeScript)
- `FunctionTypeAnnotation` (Flow)
*/
// printFunctionType is upstream's printFunctionType. The Flow `FunctionTypeAnnotation` branch and its
// helper isFlowArrowFunctionTypeAnnotation never match a TypeScript tree and are dropped.
func printFunctionType(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	parts := []any{
		// `TSConstructorType` only
		printAbstractToken(path),
	}

	if current.Is("TSConstructorType") ||
		current.Is("TSConstructSignatureDeclaration") {
		parts = append(parts, "new ")
	}

	parametersDoc := printFunctionParameters(
		path,
		options,
		print,
		/* shouldExpandParameters */ false,
		/* shouldPrintTypeParameters */ true,
	)

	returnTypeDoc := concatIn(path, printTypeAnnotationProperty(path, print, "returnType"))

	if shouldGroupFunctionParameters(current, returnTypeDoc) {
		parametersDoc = groupIn(path, parametersDoc)
	}

	parts = append(parts, parametersDoc, returnTypeDoc)

	var semicolon Doc = emptyDoc
	if current.Is("TSConstructSignatureDeclaration") ||
		current.Is("TSCallSignatureDeclaration") {
		semicolon = printClassMemberSemicolon(path, options)
	}
	return concatIn(path, groupIn(path, concatIn(path, parts...)), semicolon)
}
