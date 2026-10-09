package javascript

import (
	"github.com/system-inc/cohere/internal/format/doc"
)

// print/function-parameters.js.

// `ArrowFunctionExpression` has other dangling comments
//
// Upstream reads `comment.mark`, a property nothing sets (addDanglingComment writes `comment.marker`),
// so the filter is always true. The Go reads the same missing property to keep that behavior: the
// "commentBeforeArrow" comments still count for hasComment in printDanglingCommentsInList, and
// printDanglingComments drops them by its own marker check.
func functionParameterDanglingCommentFilter(comment Node) bool {
	return comment.String("mark") != "commentBeforeArrow"
}

/*
- `ArrowFunctionExpression`
- `FunctionDeclaration`
- `FunctionExpression`
- `ObjectMethod`
- `Property`
- `ObjectProperty`
- `ClassMethod`
- `ClassPrivateMethod`
- `MethodDefinition
- `TSFunctionType` (TypeScript)
- `TSCallSignatureDeclaration` (TypeScript)
- `TSConstructorType` (TypeScript)
- `TSConstructSignatureDeclaration` (TypeScript)
- `TSDeclareFunction`(TypeScript)
- `TSAbstractMethodDefinition` (TypeScript)
- `TSDeclareMethod` (TypeScript)
- `TSEmptyBodyFunctionExpression` (TypeScript)
- `TSMethodSignature` (TypeScript)
- `FunctionTypeAnnotation` (Flow)
- `HookDeclaration` (Flow)
- `HookTypeAnnotation` (Flow)
- `ComponentDeclaration` (Flow)
- `DeclareComponent` (Flow)
- `ComponentTypeAnnotation` (Flow)
*/
// printFunctionParameters is upstream's printFunctionParameters.
func printFunctionParameters(
	path *Path,
	options *Options,
	print PrintFunc,
	shouldExpandParameters bool,
	shouldPrintTypeParameters bool,
) Doc {
	functionNode := node(path)
	parameters := getFunctionParameters(functionNode)
	var typeParametersDoc Doc = emptyDoc
	if shouldPrintTypeParameters && functionNode.Truthy("typeParameters") {
		typeParametersDoc = print("typeParameters", nil)
	}

	if len(parameters) == 0 {
		return concatIn(path, typeParametersDoc,
			"(",
			printDanglingCommentsInList(
				path,
				options,
				functionParameterDanglingCommentFilter,
			),
			")",
		)
	}

	parent := parentOf(path)
	isParametersInTestCall := isTestCall(parent, nil)
	shouldHugParameters := shouldHugTheOnlyFunctionParameter(functionNode)
	var printed []Doc
	iterateFunctionParametersPath(path, func(parameterPath *Path, index int) {
		isLastParameter := index == len(parameters)-1
		if isLastParameter && functionNode.Truthy("rest") {
			printed = append(printed, toDoc("..."))
		}
		printed = append(printed, print(nil, nil))
		if isLastParameter {
			return
		}
		printed = append(printed, toDoc(","))
		if isParametersInTestCall || shouldHugParameters {
			printed = append(printed, toDoc(" "))
		} else if isNextLineEmptyAfter(parameters[index], options) {
			printed = append(printed, hardline, hardline)
		} else {
			printed = append(printed, line)
		}
	})

	// If the parent is a call with the first/last argument expansion and this is the
	// params of the first/last argument, we don't want the arguments to break and instead
	// want the whole expression to be on a new line.
	//
	// Good:                 Bad:
	//   verylongcall(         verylongcall((
	//     (a, b) => {           a,
	//     }                     b,
	//   )                     ) => {
	//                         })
	if shouldExpandParameters && !isDecoratedFunction(path) {
		if doc.WillBreak(typeParametersDoc) || doc.WillBreak(doc.Concat(printed)) {
			// Removing lines in this case leads to broken or ugly output
			panic(argExpansionBailout{})
		}
		return groupIn(path, concatIn(path, doc.RemoveLines(typeParametersDoc),
			"(",
			doc.RemoveLines(doc.Concat(printed)),
			")",
		))
	}

	// Single object destructuring should hug
	//
	// function({
	//   a,
	//   b,
	//   c
	// }) {}
	hasNotParameterDecorator := true
	for _, parameter := range parameters {
		if len(parameter.List("decorators")) > 0 {
			hasNotParameterDecorator = false
			break
		}
	}
	if shouldHugParameters && hasNotParameterDecorator {
		return wrapParameters(typeParametersDoc, printed)
	}

	// don't break in specs, eg; `it("should maintain parens around done even when long", (done) => {})`
	if isParametersInTestCall {
		return wrapParameters(typeParametersDoc, printed)
	}

	// Upstream's isFlowShorthandWithOneArg branch is dropped: every parent it tests
	// (isFlowObjectTypePropertyAFunction, isTypeAnnotationAFunction, TypeAlias, UnionTypeAnnotation,
	// IntersectionTypeAnnotation, FunctionTypeAnnotation) is a Flow node, and its parameter test
	// (`parameters[0].name === null`) only a Flow FunctionTypeParam passes.

	// Upstream also drops the trailing comma when `path.root.type === "NGRoot"` (Angular does not
	// allow trailing comma); the parser here is always typescript, so that test is always false.
	var trailingComma Doc = emptyDoc
	if !hasRestParameter(functionNode) {
		trailingComma = printTrailingComma(options, "all")
	}
	return concatIn(path, typeParametersDoc,
		"(",
		indentIn(path, append([]Doc{softline}, printed...)),
		trailingComma,
		softline,
		")",
	)
}

// wrapParameters is upstream's `[typeParametersDoc, "(", ...printed, ")"]`, spread flat as upstream's
// array is.
func wrapParameters(typeParametersDoc Doc, printed []Doc) Doc {
	parts := make(doc.Concat, 0, len(printed)+3)
	parts = append(parts, typeParametersDoc, toDoc("("))
	parts = append(parts, printed...)
	parts = append(parts, toDoc(")"))
	return parts
}

// shouldHugTheOnlyFunctionParameter is upstream's shouldHugTheOnlyFunctionParameter.
func shouldHugTheOnlyFunctionParameter(node Node) bool {
	if node == nil {
		return false
	}
	parameters := getFunctionParameters(node)
	if len(parameters) != 1 {
		return false
	}
	parameter := parameters[0]
	return !hasAnyComment(parameter) &&
		(parameter.Is("ObjectPattern") ||
			parameter.Is("ArrayPattern") ||
			(parameter.Is("Identifier") &&
				parameter.Truthy("typeAnnotation") &&
				(parameter.Child("typeAnnotation").Is("TypeAnnotation") ||
					parameter.Child("typeAnnotation").Is("TSTypeAnnotation")) &&
				isObjectType(parameter.Child("typeAnnotation").Child("typeAnnotation"))) ||
			(parameter.Is("FunctionTypeParam") &&
				isObjectType(parameter.Child("typeAnnotation")) &&
				parameter != node.Child("rest")) ||
			(parameter.Is("AssignmentPattern") &&
				(parameter.Child("left").Is("ObjectPattern") ||
					parameter.Child("left").Is("ArrayPattern")) &&
				(parameter.Child("right").Is("Identifier") ||
					(isObjectExpression(parameter.Child("right")) &&
						len(parameter.Child("right").List("properties")) == 0) ||
					(isArrayExpression(parameter.Child("right")) &&
						len(parameter.Child("right").List("elements")) == 0))))
}

// getReturnTypeNode is upstream's getReturnTypeNode. It returns nil where upstream returns undefined.
func getReturnTypeNode(functionNode Node) Node {
	var returnTypeNode Node
	if functionNode.Truthy("returnType") {
		returnTypeNode = functionNode.Child("returnType")
		if returnTypeNode.Truthy("typeAnnotation") {
			returnTypeNode = returnTypeNode.Child("typeAnnotation")
		}
	} else if functionNode.Truthy("typeAnnotation") {
		returnTypeNode = functionNode.Child("typeAnnotation")
	}
	return returnTypeNode
}

// When parameters are grouped, the return type annotation breaks first.
//
// shouldGroupFunctionParameters is upstream's shouldGroupFunctionParameters.
func shouldGroupFunctionParameters(functionNode Node, returnTypeDoc Doc) bool {
	returnTypeNode := getReturnTypeNode(functionNode)
	if returnTypeNode == nil {
		return false
	}

	// Upstream's `functionNode.typeParameters?.params` is an array whenever typeParameters exists; a
	// missing and an empty list both fall through both tests below.
	typeParameters := functionNode.Child("typeParameters").List("params")
	if len(typeParameters) > 1 {
		return false
	}
	if len(typeParameters) == 1 {
		typeParameter := typeParameters[0]
		if typeParameter.Truthy("constraint") || typeParameter.Truthy("default") {
			return false
		}
	}

	return len(getFunctionParameters(functionNode)) == 1 &&
		(isObjectType(returnTypeNode) || doc.WillBreak(returnTypeDoc))
}

/**
 * The "decorated function" pattern.
 * The arrow function should be kept hugged even if its signature breaks.
 *
 * ```
 * const decoratedFn = decorator(param1, param2)((
 *   ...
 * ) => {
 *   ...
 * });
 * ```
 */
// isDecoratedFunction is upstream's isDecoratedFunction.
func isDecoratedFunction(path *Path) bool {
	return path.Match(
		func(value any, _ any, _ int, _ bool) bool {
			node, _ := value.(Node)
			return node.Is("ArrowFunctionExpression") &&
				node.Child("body").Is("BlockStatement")
		},
		func(value any, name any, _ int, _ bool) bool {
			node, _ := value.(Node)
			if node.Is("CallExpression") &&
				name == "arguments" &&
				len(node.List("arguments")) == 1 &&
				node.Child("callee").Is("CallExpression") {
				decorator := node.Child("callee").Child("callee")
				return decorator.Is("Identifier") ||
					(decorator.Is("MemberExpression") &&
						!decorator.Truthy("computed") &&
						decorator.Child("object").Is("Identifier") &&
						decorator.Child("property").Is("Identifier"))
			}
			return false
		},
		func(value any, name any, _ int, _ bool) bool {
			node, _ := value.(Node)
			return (node.Is("VariableDeclarator") && name == "init") ||
				(node.Is("ExportDefaultDeclaration") && name == "declaration") ||
				(node.Is("TSExportAssignment") && name == "expression") ||
				(node.Is("AssignmentExpression") &&
					name == "right" &&
					node.Child("left").Is("MemberExpression") &&
					node.Child("left").Child("object").Is("Identifier") &&
					node.Child("left").Child("object").String("name") == "module" &&
					node.Child("left").Child("property").Is("Identifier") &&
					node.Child("left").Child("property").String("name") == "exports")
		},
		func(value any, _ any, _ int, _ bool) bool {
			node, _ := value.(Node)
			return !node.Is("VariableDeclaration") ||
				(node.String("kind") == "const" && len(node.List("declarations")) == 1)
		},
	)
}

// shouldBreakFunctionParameters is upstream's shouldBreakFunctionParameters.
func shouldBreakFunctionParameters(functionNode Node) bool {
	parameters := getFunctionParameters(functionNode)
	if len(parameters) <= 1 {
		return false
	}
	for _, parameter := range parameters {
		if parameter.Is("TSParameterProperty") {
			return true
		}
	}
	return false
}

// shouldHugTheOnlyParameter is upstream's shouldHugTheOnlyParameter.
func shouldHugTheOnlyParameter(node Node, name string) bool {
	return (name == "params" || name == "this" || name == "rest") &&
		shouldHugTheOnlyFunctionParameter(node)
}
