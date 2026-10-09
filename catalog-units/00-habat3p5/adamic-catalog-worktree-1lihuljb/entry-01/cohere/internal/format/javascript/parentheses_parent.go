package javascript

// parentheses/parent-needs-parentheses.js.

// TODO[@fisker]: Remove `needsParentheses`

// parentNeedsParentheses answers (result, decided): upstream returns a boolean when the parent decides
// and undefined when it does not, which is decided == false here.
func parentNeedsParentheses(
	path *Path,
	options *Options,
	needsParentheses func(path *Path, options *Options) bool,
) (bool, bool) {
	current := node(path)
	key := keyOf(path)
	parent := parentOf(path)

	switch parent.Type() {
	case "ReturnStatement", "ThrowStatement":
		if willReturnOrThrowStatementBreak(path, options) {
			return false, true
		}

	case "ParenthesizedExpression":
		return false, true
	case "ClassDeclaration", "ClassExpression":
		// Add parens around the extends clause of a class. It is needed for almost
		// all expressions.
		if key == "superClass" {
			superClass := stripChainElementWrappers(current)
			if superClass.Is("ArrowFunctionExpression") ||
				superClass.Is("AssignmentExpression") ||
				superClass.Is("AwaitExpression") ||
				superClass.Is("BinaryExpression") ||
				superClass.Is("ConditionalExpression") ||
				superClass.Is("LogicalExpression") ||
				superClass.Is("NewExpression") ||
				superClass.Is("ObjectExpression") ||
				superClass.Is("SequenceExpression") ||
				superClass.Is("TaggedTemplateExpression") ||
				superClass.Is("UnaryExpression") ||
				superClass.Is("UpdateExpression") ||
				superClass.Is("YieldExpression") ||
				superClass.Is("ClassExpression") &&
					len(superClass.List("decorators")) > 0 {
				return true, true
			}
		}

	case "ExportDefaultDeclaration":
		// `export default function` or `export default class` can't be followed by
		// anything after. So an expression like `export default (function(){}).toString()`
		// needs to be followed by a parentheses
		if shouldWrapFunctionForExportDefault(path, options, needsParentheses) {
			return true, true
		}

	case "Decorator":
		if key == "expression" &&
			!canDecoratorExpressionUnparenthesized(current) {
			return true, true
		}

	case "TypeAnnotation":
		if path.Match(
			nil,
			nil,
			keyIsType("returnType", "ArrowFunctionExpression"),
		) &&
			!(current.Is("NullableTypeAnnotation") &&
				call(path, func(path *Path) bool { return needsParentheses(path, options) }, "typeAnnotation")) &&
			includesFunctionTypeInObjectType(current) {
			return true, true
		}

	case "VariableDeclarator":
		// Legacy syntax
		// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Errors/Invalid_for-in_initializer
		// `for (var a = 1 in b);`
		if key == "init" &&
			path.Match(
				nil,
				nil,
				keyIsType("declarations", "VariableDeclaration"),
				keyIsType("left", "ForInStatement"),
			) {
			return true, true
		}

	case "TSInstantiationExpression":
		if key == "expression" &&
			(current.Is("AwaitExpression") || current.Is("YieldExpression")) {
			return true, true
		}
	}

	return false, false
}

func willReturnOrThrowStatementBreak(path *Path, options *Options) bool {
	key := keyOf(path)
	parent := parentOf(path)
	if !(key == "argument" && isReturnOrThrowStatement(parent)) {
		return false
	}

	/*
	   When `ReturnStatement` or `ThrowStatement` breaks, parentheses will be added around it's argument.
	   So don't need add parentheses again.
	   But we can't know how the argument printed, so only matches cases that will break for sure
	*/

	current := node(path)

	if (current.Is("SequenceExpression") ||
		current.Is("AssignmentExpression")) &&
		returnArgumentHasLeadingComment(current, options) {
		return true
	}

	return false
}

func shouldWrapFunctionForExportDefault(
	path *Path,
	options *Options,
	needsParentheses func(path *Path, options *Options) bool,
) bool {
	current := node(path)
	parent := parentOf(path)

	if current.Is("FunctionExpression") || current.Is("ClassExpression") {
		return parent.Is("ExportDefaultDeclaration") ||
			// in some cases the function is already wrapped
			// (e.g. `export default (function() {})();`)
			// in this case we don't need to add extra parens
			!needsParentheses(path, options)
	}

	if !hasNakedLeftSide(current) ||
		!parent.Is("ExportDefaultDeclaration") &&
			needsParentheses(path, options) {
		return false
	}

	return call(
		path,
		func(path *Path) bool { return shouldWrapFunctionForExportDefault(path, options, needsParentheses) },
		getLeftSidePathName(current)...,
	)
}

// Based on babel implementation
// https://github.com/nicolo-ribaudo/babel/blob/c4b88a4e5005364255f7e964fe324cf7bfdfb019/packages/babel-generator/src/node/index.ts#L111
func canDecoratorExpressionUnparenthesized(node Node) bool {
	if node.Is("ChainExpression") {
		node = node.Child("expression")
	}

	return isDecoratorMemberExpression(node) ||
		isCallExpression(node) &&
			!node.Truthy("optional") &&
			isDecoratorMemberExpression(node.Child("callee"))
}

func isDecoratorMemberExpression(node Node) bool {
	if node.Is("Identifier") {
		return true
	}

	if isMemberExpression(node) {
		return !node.Truthy("computed") &&
			!node.Truthy("optional") &&
			node.Child("property").Is("Identifier") &&
			isDecoratorMemberExpression(node.Child("object"))
	}

	return false
}

func includesFunctionTypeInObjectType(node Node) bool {
	return hasNode(
		node,
		func(node Node) bool {
			return node.Is("ObjectTypeAnnotation") &&
				hasNode(node, func(node Node) bool { return node.Is("FunctionTypeAnnotation") })
		},
	)
}
