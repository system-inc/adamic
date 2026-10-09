package javascript

import "github.com/system-inc/cohere/internal/format/doc"

// print/method-signature.js, print/enum.js, print/type-assertion.js, print/index-signature.js,
// print/indexed-access-type.js, print/infer-type.js, print/rest-type.js, print/type-predicate.js and
// print/type-query.js: the small TypeScript type printers, one upstream file each.

// print/method-signature.js

/*
- `TSMethodSignature` (TypeScript)
*/
// printMethodSignature is upstream's printMethodSignature.
func printMethodSignature(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	var parts []any
	kind := ""
	if current.Truthy("kind") && current.String("kind") != "method" {
		kind = current.String("kind") + " "
	}
	parts = append(parts,
		printTypeScriptAccessibilityToken(current),
		kind,
		printKey(path, options, print),
		printOptionalToken(path),
	)

	parametersDoc := printFunctionParameters(
		path,
		options,
		print,
		/* shouldExpandParameters */ false,
		/* shouldPrintTypeParameters */ true,
	)

	returnTypeDoc := printTypeAnnotationProperty(path, print, "returnType")
	shouldGroupParameters := shouldGroupFunctionParameters(
		current,
		returnTypeDoc,
	)

	if shouldGroupParameters {
		parts = append(parts, groupIn(path, parametersDoc))
	} else {
		parts = append(parts, parametersDoc)
	}

	if current.Truthy("returnType") {
		parts = append(parts, groupIn(path, returnTypeDoc))
	}

	return concatIn(path, groupIn(path, concatIn(path, parts...)), printClassMemberSemicolon(path, options))
}

// print/enum.js. printFlowEnumBody and printLegacyFlowEnumBody print Flow enum bodies and are not
// ported.

/*
- `EnumBooleanMember`(flow)
- `EnumNumberMember`(flow)
- `EnumBigIntMember`(flow)
- `EnumStringMember`(flow)
- `EnumDefaultedMember`(flow)
- `TSEnumMember`(TypeScript)
*/
// printEnumMember is upstream's printEnumMember. Only `TSEnumMember` reaches it in a TypeScript tree,
// so upstream's Flow `id` / `init` branch is dropped.
func printEnumMember(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)

	idDoc := printKey(path, options, print)
	initializerProperty := "initializer"
	if !current.Truthy(initializerProperty) {
		return idDoc
	}

	return concatIn(path, idDoc, " = ", print(initializerProperty, nil))
}

/*
- `DeclareEnum`(flow)
- `EnumDeclaration`(flow)
- `TSEnumDeclaration`(TypeScript)
*/
// printEnumDeclaration is upstream's printEnumDeclaration.
func printEnumDeclaration(path *Path, print PrintFunc) Doc {
	current := node(path)
	constToken := ""
	if current.Truthy("const") {
		constToken = "const "
	}
	return concatIn(path, printDeclareToken(path),
		constToken,
		"enum ",
		print("id", nil),
		" ",
		print("body", nil),
	)
}

// print/type-assertion.js

/*
- `TSTypeAssertion` (TypeScript)
*/
// printTypeAssertion is upstream's printTypeAssertion.
func printTypeAssertion(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	shouldBreakAfterCast := !(isArrayExpression(current.Child("expression")) ||
		isObjectExpression(current.Child("expression")))

	castGroup := groupIn(path, concatIn(path, "<",
		indentIn(path, concatIn(path, softline, print("typeAnnotation", nil))),
		softline,
		">",
	))

	exprContents := concatIn(path, ifBreak("(", ""),
		indentIn(path, concatIn(path, softline, print("expression", nil))),
		softline,
		ifBreak(")", ""),
	)

	if shouldBreakAfterCast {
		return conditionalGroup([]Doc{
			concatIn(path, castGroup, print("expression", nil)),
			concatIn(path, castGroup, groupWithIn(path, exprContents, doc.GroupOptions{ShouldBreak: true})),
			concatIn(path, castGroup, print("expression", nil)),
		}, doc.GroupOptions{})
	}
	return groupIn(path, concatIn(path, castGroup, print("expression", nil)))
}

// print/index-signature.js

// printIndexSignature is upstream's printIndexSignature.
func printIndexSignature(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	// The TypeScript parser accepts multiple parameters here. If you're
	// using them, it makes sense to have a trailing comma. But if you
	// aren't, this is more like a computed property name than an array.
	// So we leave off the trailing comma when there's just one parameter.
	var trailingComma Doc = emptyDoc
	if len(current.List("parameters")) > 1 {
		trailingComma = printTrailingComma(options, "")
	}

	parametersGroup := groupIn(path, concatIn(path, indentIn(path, concatIn(path, softline, join(concatIn(path, ", ", softline), printAll(path, print, "parameters")))),
		trailingComma,
		softline,
	))

	isClassMember := keyOf(path) == "body" && parentOf(path).Is("ClassBody")

	staticToken := ""
	// `static` only allowed in class member
	if isClassMember && current.Truthy("static") {
		staticToken = "static "
	}
	readonly := ""
	if current.Truthy("readonly") {
		readonly = "readonly "
	}
	var parameters Doc = emptyDoc
	// upstream's `node.parameters ?`: an array is truthy even when empty.
	if current.Get("parameters") != nil {
		parameters = parametersGroup
	}

	return concatIn(path, staticToken,
		readonly,
		"[",
		parameters,
		"]",
		printTypeAnnotationProperty(path, print),
		printClassMemberSemicolon(path, options),
	)
}

// print/indexed-access-type.js

/*
- `TSIndexedAccessType`(TypeScript)
- `IndexedAccessType`(flow)
- `OptionalIndexedAccessType`(flow)
*/
// printIndexedAccessType is upstream's printIndexedAccessType.
func printIndexedAccessType(path *Path, options *Options, print PrintFunc) Doc {
	return concatIn(path, print("objectType", nil),
		printOptionalToken(path),
		"[",
		print("indexType", nil),
		"]",
	)
}

// print/infer-type.js

/*
- `TSInferType`(TypeScript)
- `InferTypeAnnotation`(flow)
*/
// printInferType is upstream's printInferType.
func printInferType(path *Path, options *Options, print PrintFunc) Doc {
	return concatIn(path, "infer ", print("typeParameter", nil))
}

// print/rest-type.js

/*
- `TSRestType`(TypeScript)
- `TupleTypeSpreadElement`(flow)
*/
// printRestType is upstream's printRestType. The Flow `TupleTypeSpreadElement` label branch is
// dropped.
func printRestType(path *Path, options *Options, print PrintFunc) Doc {
	return concatIn(path, "...",
		print("typeAnnotation", nil),
	)
}

// print/type-predicate.js

/*
- `TSTypePredicate` (TypeScript)
- `TypePredicate` (flow)
*/
// printTypePredicate is upstream's printTypePredicate. The Flow `TypePredicate` kind prefix is
// dropped.
func printTypePredicate(path *Path, print PrintFunc) Doc {
	current := node(path)
	prefix := ""
	if current.Is("TSTypePredicate") && current.Truthy("asserts") {
		prefix = "asserts "
	}
	var typeAnnotation Doc = emptyDoc
	if current.Truthy("typeAnnotation") {
		typeAnnotation = concatIn(path, " is ", printTypeAnnotationProperty(path, print))
	}
	return concatIn(path, prefix,
		print("parameterName", nil),
		typeAnnotation,
	)
}

// print/type-query.js

/*
- `TSTypeQuery` (TypeScript)
- `TypeofTypeAnnotation` (flow)
*/
// printTypeQuery is upstream's printTypeQuery.
func printTypeQuery(path *Path, print PrintFunc) Doc {
	argumentPropertyName := "argument"
	if node(path).Is("TSTypeQuery") {
		argumentPropertyName = "exprName"
	}
	return concatIn(path, "typeof ", print(argumentPropertyName, nil), print("typeArguments", nil))
}
