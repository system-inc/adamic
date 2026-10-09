package javascript

import (
	"fmt"
	"slices"

	"github.com/system-inc/cohere/internal/format/doc"
)

// print/function.js.

// isMethodValue is upstream's isMethodValue.
func isMethodValue(path *Path) bool {
	current := node(path)
	parent := parentOf(path)
	return keyOf(path) == "value" &&
		current.Is("FunctionExpression") &&
		(parent.Is("ObjectMethod") ||
			parent.Is("ClassMethod") ||
			parent.Is("ClassPrivateMethod") ||
			parent.Is("MethodDefinition") ||
			parent.Is("TSAbstractMethodDefinition") ||
			parent.Is("TSDeclareMethod") ||
			(parent.Is("Property") && isMethod(parent)))
}

/*
- `FunctionDeclaration`
- `FunctionExpression`
- `HookDeclaration` (Flow)
- `TSDeclareFunction`(TypeScript)
*/
// printFunction is upstream's printFunction.
func printFunction(path *Path, options *Options, print PrintFunc, args *printArguments) Doc {
	if isMethodValue(path) {
		return printMethodValue(path, options, print)
	}

	current := node(path)

	shouldExpandParameters := false
	if current.Is("FunctionExpression") && args != nil && args.expandLastArg {
		parent := parentOf(path)
		if isCallExpression(parent) &&
			(len(getCallArguments(parent)) > 1 ||
				!slices.ContainsFunc(getFunctionParameters(current), func(param Node) bool {
					// upstream's .every(param => param.type === "Identifier" && !param.typeAnnotation)
					return !(param.Is("Identifier") && !param.Truthy("typeAnnotation"))
				})) {
			shouldExpandParameters = true
		}
	}

	parametersDoc := printFunctionParameters(
		path,
		options,
		print,
		shouldExpandParameters,
		false,
	)
	returnTypeDoc := printReturnType(path, print)
	shouldGroupParameters := shouldGroupFunctionParameters(
		current,
		returnTypeDoc,
	)

	isFlowHookDeclaration := current.Is("HookDeclaration")
	keyword := "function"
	if isFlowHookDeclaration {
		keyword = "hook"
	}

	async := ""
	if current.Truthy("async") {
		async = "async "
	}
	generator := ""
	if current.Truthy("generator") {
		generator = "*"
	}
	// @system-inc: upstream always emits a space here, so an anonymous
	// function expression prints as `function () {}` where the space stands in
	// for the missing name. We only emit it when there is an actual name, so
	// anonymous forms print tight: `function() {}`, `function*() {}`,
	// `async function*() {}`, `function<T>() {}`. Named declarations are
	// unchanged: `function foo() {}`.
	idSpace := ""
	var idDoc Doc = emptyDoc
	if current.Truthy("id") {
		idSpace = " "
		idDoc = print("id", nil)
	}
	var parameters Doc = parametersDoc
	if shouldGroupParameters {
		parameters = groupIn(path, parametersDoc)
	}
	bodySpace := ""
	if current.Truthy("body") {
		bodySpace = " "
	}
	var semicolon Doc = emptyDoc
	if current.Truthy("declare") || !current.Truthy("body") {
		semicolon = printSemicolon(options)
	}

	return concatIn(path, printDeclareToken(path),
		async,
		keyword,
		generator,
		idSpace,
		idDoc,
		print("typeParameters", nil),
		groupIn(path, concatIn(path, parameters,
			returnTypeDoc,
		)),
		bodySpace,
		print("body", nil),
		semicolon,
	)
}

/*
- `FunctionDeclaration`
- `FunctionExpression`
- `TSDeclareFunction`(TypeScript)
- `ObjectMethod`
- `Property`
- `ObjectProperty`
- `ClassMethod`
- `ClassPrivateMethod`
- `MethodDefinition
- `TSAbstractMethodDefinition` (TypeScript)
- `TSDeclareMethod` (TypeScript)
*/
// printMethod is upstream's printMethod.
func printMethod(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	kind := current.String("kind")
	// upstream's node.value || node
	value := current.Child("value")
	if value == nil {
		value = current
	}
	var parts []any

	if kind == "" || kind == "init" || kind == "method" || kind == "constructor" {
		if value.Truthy("async") {
			parts = append(parts, "async ")
		}
	} else {
		// upstream's assert.ok(kind === "get" || kind === "set")
		if kind != "get" && kind != "set" {
			panic(fmt.Sprintf("javascript: unexpected method kind %q", kind))
		}

		parts = append(parts, kind, " ")
	}

	// A `getter`/`setter` can't be a generator, but it's recoverable
	if value.Truthy("generator") {
		parts = append(parts, "*")
	}

	optional := ""
	if current.Truthy("optional") {
		optional = "?"
	}
	var valueDoc Doc
	if current == value {
		valueDoc = printMethodValue(path, options, print)
	} else {
		valueDoc = print("value", nil)
	}
	parts = append(parts,
		printKey(path, options, print),
		optional,
		valueDoc,
	)

	return concatIn(path, parts...)
}

/*
- `ObjectMethod`
- `Property`
- `ObjectProperty`
- `ClassMethod`
- `ClassPrivateMethod`
- `MethodDefinition
- `TSAbstractMethodDefinition` (TypeScript)
- `TSDeclareMethod` (TypeScript)
- `TSEmptyBodyFunctionExpression` (TypeScript)
*/
// printMethodValue is upstream's printMethodValue.
func printMethodValue(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	parametersDoc := printFunctionParameters(path, options, print, false, false)
	returnTypeDoc := printReturnType(path, print)
	shouldBreakParameters := shouldBreakFunctionParameters(current)
	shouldGroupParameters := shouldGroupFunctionParameters(
		current,
		returnTypeDoc,
	)
	var parameters Doc
	if shouldBreakParameters {
		parameters = groupWithIn(path, parametersDoc, doc.GroupOptions{ShouldBreak: true})
	} else if shouldGroupParameters {
		parameters = groupIn(path, parametersDoc)
	} else {
		parameters = parametersDoc
	}
	parts := []any{
		print("typeParameters", nil),
		groupIn(path, concatIn(path, parameters,
			returnTypeDoc,
		)),
	}

	if current.Truthy("body") {
		parts = append(parts, " ", print("body", nil))
	} else {
		parts = append(parts, printSemicolon(options))
	}

	return concatIn(path, parts...)
}

// canPrintParamsWithoutParens is upstream's canPrintParamsWithoutParens.
func canPrintParamsWithoutParens(node Node) bool {
	parameters := getFunctionParameters(node)
	return len(parameters) == 1 &&
		!node.Truthy("typeParameters") &&
		!hasComment(node, commentDangling, nil) &&
		parameters[0].Is("Identifier") &&
		!parameters[0].Truthy("typeAnnotation") &&
		!hasAnyComment(parameters[0]) &&
		!parameters[0].Truthy("optional") &&
		!node.Truthy("predicate") &&
		!node.Truthy("returnType")
}

// shouldPrintParamsWithoutParens is upstream's shouldPrintParamsWithoutParens.
func shouldPrintParamsWithoutParens(path *Path, options *Options) bool {
	arrowParens := settingsOf(options).ArrowParens
	if arrowParens == "always" {
		return false
	}

	if arrowParens == "avoid" {
		current := node(path)
		return canPrintParamsWithoutParens(current)
	}

	// Fallback default; should be unreachable
	return false
}

// printReturnType is upstream's printReturnType.
func printReturnType(path *Path, print PrintFunc) Doc {
	current := node(path)
	returnType := printTypeAnnotationProperty(path, print, "returnType")

	parts := []any{returnType}

	if current.Truthy("predicate") {
		parts = append(parts, print("predicate", nil))
	}

	return concatIn(path, parts...)
}
