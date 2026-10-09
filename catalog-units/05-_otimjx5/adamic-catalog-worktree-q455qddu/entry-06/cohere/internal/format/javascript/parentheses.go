package javascript

import (
	"slices"
	"strings"

	"github.com/system-inc/cohere/internal/format/printing"
)

// parentheses/needs-parentheses.js. Flow and Babel type names stay in the cases, as upstream has them,
// even though only the TypeScript and ESTree names occur in a typescript-estree tree.

// needsParentheses is upstream's needsParentheses.
func needsParentheses(path *Path, options *Options) bool {
	if path.IsRoot() {
		return false
	}

	current := node(path)
	key := keyOf(path)
	parent := parentOf(path)

	// Dropped: the `options.__isInHtmlInterpolation` check (and its endsWithRightBracket and
	// isFollowedByRightBracket helpers), which only the HTML host sets.

	// Only statements don't need parentheses.
	if isStatement(current) {
		return false
	}

	if current.Is("Identifier") {
		return shouldAddParenthesesToIdentifier(path)
	}

	if current.Is("ObjectExpression", "FunctionExpression", "ClassExpression", "DoExpression") {
		ancestor, _ := path.FindAncestor(func(node Node) bool { return node.Is("ExpressionStatement") })
		expression := ancestor.Child("expression")
		if expression != nil &&
			startsWithNoLookaheadToken(expression, func(leftmostNode Node) bool { return leftmostNode == current }) {
			return true
		}
	}

	if current.Is("ObjectExpression") {
		ancestor, _ := path.FindAncestor(func(node Node) bool { return node.Is("ArrowFunctionExpression") })
		arrowFunctionBody := ancestor.Child("body")
		if arrowFunctionBody != nil &&
			!arrowFunctionBody.Is("SequenceExpression") && // these have parens added anyway
			!arrowFunctionBody.Is("AssignmentExpression") &&
			startsWithNoLookaheadToken(arrowFunctionBody, func(leftmostNode Node) bool { return leftmostNode == current }) {
			return true
		}
	}

	if parentCheckResult, decided := parentNeedsParentheses(path, options, needsParentheses); decided {
		return parentCheckResult
	}

	switch current.Type() {
	case "UpdateExpression":
		if parent.Is("UnaryExpression") {
			return current.Truthy("prefix") &&
				(current.String("operator") == "++" && parent.String("operator") == "+" ||
					current.String("operator") == "--" && parent.String("operator") == "-")
		}
		fallthrough
	case "UnaryExpression":
		switch parent.Type() {
		case "UnaryExpression":
			return current.String("operator") == parent.String("operator") &&
				(current.String("operator") == "+" || current.String("operator") == "-")

		case "BindExpression":
			return true

		case "MemberExpression", "OptionalMemberExpression":
			return key == "object"

		case "TaggedTemplateExpression":
			return true

		case "NewExpression", "CallExpression", "OptionalCallExpression":
			return key == "callee"

		case "BinaryExpression":
			// A user typing `!foo instanceof Bar` probably intended
			// `!(foo instanceof Bar)`, so format to `(!foo) instance Bar` to what is
			// really happening
			if key == "left" && current.Is("UnaryExpression") &&
				(parent.String("operator") == "in" || parent.String("operator") == "instanceof") {
				return true
			}

			return key == "left" && parent.String("operator") == "**"

		case "TSNonNullExpression":
			return true

		default:
			return false
		}

	case "BinaryExpression":
		if parent.Is("UpdateExpression") {
			return true
		}

		// We add parentheses to any `a in b` inside `ForStatement` initializer
		// https://github.com/prettier/prettier/issues/907#issuecomment-284304321
		if current.String("operator") == "in" && isPathInForStatementInitializer(path) {
			return true
		}
		if current.String("operator") == "|>" && current.Parenthesized {
			grandParent := grandparentOf(path)
			if grandParent.Is("BinaryExpression") && grandParent.String("operator") == "|>" {
				return true
			}
		}
		fallthrough
	case "TSTypeAssertion", "TSAsExpression", "TSSatisfiesExpression", "AsExpression", "AsConstExpression",
		"SatisfiesExpression", "LogicalExpression":
		switch parent.Type() {
		case "TSAsExpression", "TSSatisfiesExpression", "AsExpression", "AsConstExpression", "SatisfiesExpression":
			// examples:
			//   foo as unknown as Bar
			//   foo satisfies unknown satisfies Bar
			//   foo satisfies unknown as Bar
			//   foo as unknown satisfies Bar
			return !isBinaryCastExpression(current)

		case "ConditionalExpression":
			return isBinaryCastExpression(current) || isNullishCoalescing(current)

		case "CallExpression", "NewExpression", "OptionalCallExpression":
			return key == "callee"

		case "ClassExpression", "ClassDeclaration":
			return key == "superClass"

		case "TSTypeAssertion", "TaggedTemplateExpression", "JSXSpreadAttribute", "SpreadElement", "BindExpression",
			"AwaitExpression", "TSNonNullExpression", "UpdateExpression":
			return true
		case "UnaryExpression":
			// `UnaryExpression` adds parentheses and indention when argument has comment
			if !hasAnyComment(current) {
				return true
			}

		case "MemberExpression", "OptionalMemberExpression":
			return key == "object"

		case "AssignmentExpression", "AssignmentPattern":
			return key == "left" && (current.Is("TSTypeAssertion") || isBinaryCastExpression(current))

		case "LogicalExpression":
			if current.Is("LogicalExpression") {
				return parent.String("operator") != current.String("operator")
			}
			fallthrough

		case "BinaryExpression":
			operator := current.String("operator")
			if operator == "" && !current.Is("TSTypeAssertion") {
				return true
			}

			// getPrecedence answers false where upstream's is undefined, and undefined compares unequal
			// to every number under <, > and ===, but equal to itself under ===.
			precedence, known := getPrecedence(operator)
			parentOperator := parent.String("operator")
			parentPrecedence, parentKnown := getPrecedence(parentOperator)
			bothKnown := known && parentKnown
			equalPrecedence := known == parentKnown && parentPrecedence == precedence

			if bothKnown && parentPrecedence > precedence {
				return true
			}

			if key == "right" && equalPrecedence {
				return true
			}

			if equalPrecedence && !shouldFlatten(parentOperator, operator) {
				return true
			}

			if bothKnown && parentPrecedence < precedence && operator == "%" &&
				(parentOperator == "+" || parentOperator == "-") {
				return true
			}

			// Add parenthesis when working with bitwise operators
			// It's not strictly needed but helps with code understanding
			if isBitwiseOperator(parentOperator) {
				return true
			}

			return false

		default:
			return false
		}

	case "SequenceExpression":
		// Although parentheses wouldn't hurt around sequence
		// expressions in the head of for loops, traditional style
		// dictates that e.g. i++, j++ should not be wrapped with
		// parentheses.
		if parent.Is("ForStatement") {
			return false
		}

		// Otherwise err on the side of overparenthesization, adding
		// explicit exceptions above if this proves overzealous.
		return true

	case "YieldExpression":
		if parent.Is("AwaitExpression") || parent.Is("TSTypeAssertion") {
			return true
		}
		fallthrough
	case "AwaitExpression":
		switch parent.Type() {
		case "TaggedTemplateExpression", "UnaryExpression", "LogicalExpression", "SpreadElement", "TSAsExpression",
			"TSSatisfiesExpression", "TSNonNullExpression", "AsExpression", "AsConstExpression", "SatisfiesExpression",
			"BindExpression":
			return true

		case "MemberExpression", "OptionalMemberExpression":
			return key == "object"

		case "NewExpression", "CallExpression", "OptionalCallExpression":
			return key == "callee"

		case "ConditionalExpression":
			return key == "test"

		case "BinaryExpression":
			if !current.Truthy("argument") && parent.String("operator") == "|>" {
				return false
			}

			return true

		default:
			return false
		}

	case "TSFunctionType":
		if path.Match(
			func(value any, _ any, _ int, _ bool) bool {
				node, _ := value.(Node)
				return node.Is("TSFunctionType")
			},
			keyIsType("typeAnnotation", "TSTypeAnnotation"),
			keyIsType("returnType", "ArrowFunctionExpression"),
		) {
			return true
		}
		fallthrough
	case "TSConditionalType", "TSConstructorType", "ConditionalTypeAnnotation":
		if key == "extendsType" && isConditionalType(current) && parent.Type() == current.Type() {
			return true
		}

		// `const foo = <Foo extends (Bar extends Baz ? A : B)>() => true;`
		//                            ^^^^^^^^^^^^^^^^^^^^^^^
		if key == "constraint" && current.Is("TSConditionalType") && parent.Is("TSTypeParameter") ||
			key == "typeAnnotation" && current.Is("ConditionalTypeAnnotation") && parent.Is("TypeAnnotation") &&
				grandparentOf(path).Is("TypeParameter") && grandparentOf(path).Child("bound") == parent &&
				grandparentOf(path).Truthy("usesExtendsBound") {
			return true
		}

		if key == "checkType" && isConditionalType(parent) {
			return true
		}

		if key == "extendsType" && parent.Is("TSConditionalType") {
			annotation := current.Child("returnType")
			if annotation == nil {
				annotation = current.Child("typeAnnotation")
			}
			typeAnnotation := annotation.Child("typeAnnotation")

			if typeAnnotation.Is("TSTypePredicate") && typeAnnotation.Truthy("typeAnnotation") {
				typeAnnotation = typeAnnotation.Child("typeAnnotation").Child("typeAnnotation")
			}

			if typeAnnotation.Is("TSInferType") && typeAnnotation.Child("typeParameter").Truthy("constraint") {
				return true
			}
		}

		fallthrough
	case "TSUnionType", "TSIntersectionType":
		if isUnionType(parent) || isIntersectionType(parent) {
			return true
		}
		fallthrough
	case "TSInferType":
		if current.Is("TSInferType") {
			if parent.Is("TSRestType") {
				return false
			}

			if key == "types" && parent.Is("TSUnionType", "TSIntersectionType") &&
				current.Child("typeParameter").Is("TSTypeParameter") &&
				current.Child("typeParameter").Truthy("constraint") {
				return true
			}
		}
		fallthrough
	case "TSTypeOperator":
		return parent.Is("TSArrayType") ||
			parent.Is("TSOptionalType") ||
			parent.Is("TSRestType") ||
			key == "objectType" && parent.Is("TSIndexedAccessType") ||
			parent.Is("TSTypeOperator") ||
			parent.Is("TSTypeAnnotation") && strings.HasPrefix(grandparentOf(path).Type(), "TSJSDoc")
	case "TSTypeQuery":
		return key == "objectType" && parent.Is("TSIndexedAccessType") ||
			key == "elementType" && parent.Is("TSArrayType")
	// Same as `TSTypeOperator`, but for Flow syntax
	case "TypeOperator":
		return parent.Is("ArrayTypeAnnotation") ||
			parent.Is("NullableTypeAnnotation") ||
			key == "objectType" && parent.Is("IndexedAccessType", "OptionalIndexedAccessType") ||
			parent.Is("TypeOperator")
	// Same as `TSTypeQuery`, but for Flow syntax
	case "TypeofTypeAnnotation", "KeyofTypeAnnotation":
		return key == "objectType" && parent.Is("IndexedAccessType", "OptionalIndexedAccessType") ||
			key == "elementType" && parent.Is("ArrayTypeAnnotation")
	case "ArrayTypeAnnotation":
		return parent.Is("NullableTypeAnnotation")

	case "IntersectionTypeAnnotation", "UnionTypeAnnotation":
		return parent.Is("TypeOperator") ||
			parent.Is("KeyofTypeAnnotation") ||
			parent.Is("ArrayTypeAnnotation") ||
			parent.Is("NullableTypeAnnotation") ||
			parent.Is("IntersectionTypeAnnotation") ||
			parent.Is("UnionTypeAnnotation") ||
			key == "objectType" && parent.Is("IndexedAccessType", "OptionalIndexedAccessType")
	case "InferTypeAnnotation", "NullableTypeAnnotation":
		if parent.Is("ArrayTypeAnnotation") ||
			key == "objectType" && parent.Is("IndexedAccessType", "OptionalIndexedAccessType") {
			return true
		}

	case "ComponentTypeAnnotation", "FunctionTypeAnnotation":
		if current.Is("ComponentTypeAnnotation") && current.Get("rendersType") == nil {
			return false
		}

		if path.Match(
			nil,
			keyIsType("typeAnnotation", "TypeAnnotation"),
			keyIsType("returnType", "ArrowFunctionExpression"),
		) {
			return true
		}

		// If the return type is a nullable arrow function, then we need a paren
		// otherwise the inner => can be assumed to be for the outer one.
		if path.Match(
			nil,
			keyIsType("typeAnnotation", "NullableTypeAnnotation"),
			keyIsType("typeAnnotation", "TypeAnnotation"),
			keyIsType("returnType", "ArrowFunctionExpression"),
		) {
			return true
		}

		// Matches the following case in Flow:
		//
		//     const a = (x: any): x is (number => string) => true;
		//
		// This case is not necessary in TS since `number => string` is not a valid
		// arrow type there.
		if path.Match(
			nil,
			keyIsType("typeAnnotation", "TypePredicate"),
			keyIsType("typeAnnotation", "TypeAnnotation"),
			keyIsType("returnType", "ArrowFunctionExpression"),
		) {
			return true
		}

		ancestor := parent
		if parent.Is("NullableTypeAnnotation") {
			ancestor = grandparentOf(path)
		}

		return ancestor.Is("UnionTypeAnnotation") ||
			ancestor.Is("IntersectionTypeAnnotation") ||
			ancestor.Is("ArrayTypeAnnotation") ||
			key == "objectType" && ancestor.Is("IndexedAccessType", "OptionalIndexedAccessType") ||
			key == "checkType" && parent.Is("ConditionalTypeAnnotation") ||
			key == "extendsType" && parent.Is("ConditionalTypeAnnotation") &&
				current.Child("returnType").Is("InferTypeAnnotation") &&
				current.Child("returnType").Child("typeParameter").Truthy("bound") ||
			// We should check ancestor's parent to know whether the parentheses
			// are really needed, but since ??T doesn't make sense this check
			// will almost never be true.
			ancestor.Is("NullableTypeAnnotation") ||
			// See #5283
			// `parent.name === null` is strict: a missing name is undefined, not null.
			parent.Is("FunctionTypeParam") && parent.Has("name") && parent.Get("name") == nil &&
				slices.ContainsFunc(getFunctionParameters(current), func(parameter Node) bool {
					return parameter.Child("typeAnnotation").Is("NullableTypeAnnotation")
				})

	case "OptionalIndexedAccessType":
		return key == "objectType" && parent.Is("IndexedAccessType")

	case "StringLiteral", "NumericLiteral", "Literal":
		_, valueIsString := current.Get("value").(string)
		_, directiveIsString := parent.Get("directive").(string)
		if valueIsString && parent.Is("ExpressionStatement") && !directiveIsString {
			// To avoid becoming a directive
			grandParent := grandparentOf(path)

			return grandParent.Is("Program") || grandParent.Is("BlockStatement")
		}

		return key == "object" && isMemberExpression(parent) && isNumericLiteral(current)

	case "AssignmentExpression":
		if (key == "init" || key == "update") && parent.Is("ForStatement") {
			return false
		}

		if key == "expression" && !current.Child("left").Is("ObjectPattern") && parent.Is("ExpressionStatement") {
			return false
		}

		if key == "key" && parent.Is("TSPropertySignature") {
			return false
		}

		if parent.Is("AssignmentExpression") {
			return false
		}

		if key == "expressions" && parent.Is("SequenceExpression") &&
			path.Match(nil, nil, func(value any, name any, _ int, _ bool) bool {
				node, _ := value.(Node)
				return (name == "init" || name == "update") && node.Is("ForStatement")
			}) {
			return false
		}

		if key == "value" && parent.Is("Property") &&
			path.Match(nil, nil, keyIsType("properties", "ObjectPattern")) {
			return false
		}

		if parent.Is("NGChainedExpression") {
			return false
		}

		if key == "node" && parent.Is("JsExpressionRoot") {
			return false
		}

		return true

	case "ConditionalExpression":
		switch parent.Type() {
		case "TaggedTemplateExpression", "UnaryExpression", "SpreadElement", "BinaryExpression", "LogicalExpression",
			"NGPipeExpression", "AwaitExpression", "JSXSpreadAttribute", "TSTypeAssertion", "TypeCastExpression",
			"TSAsExpression", "TSSatisfiesExpression", "AsExpression", "AsConstExpression", "SatisfiesExpression",
			"TSNonNullExpression":
			return true

		case "NewExpression", "CallExpression", "OptionalCallExpression":
			return key == "callee"

		case "ConditionalExpression":
			// TODO remove this case entirely once we've removed this flag.
			// options.experimentalTernaries is always false here: formatoptions.Options does not carry it.
			return key == "test"

		case "MemberExpression", "OptionalMemberExpression":
			return key == "object"

		default:
			return false
		}

	case "FunctionExpression":
		switch parent.Type() {
		case "NewExpression", "CallExpression", "OptionalCallExpression":
			// Not always necessary, but it's clearer to the reader if IIFEs are wrapped in parentheses.
			// Is necessary if it is `expression` of `ExpressionStatement`.
			return key == "callee"
		case "TaggedTemplateExpression":
			return true // This is basically a kind of IIFE.
		case "ExportDefaultDeclaration":
			return key == "declaration"
		default:
			return false
		}

	case "ArrowFunctionExpression":
		switch parent.Type() {
		case "BinaryExpression":
			return parent.String("operator") != "|>" || current.Parenthesized
		case "NewExpression", "CallExpression", "OptionalCallExpression":
			return key == "callee"

		case "MemberExpression", "OptionalMemberExpression":
			return key == "object"

		case "TSAsExpression", "TSSatisfiesExpression", "AsExpression", "AsConstExpression", "SatisfiesExpression",
			"TSNonNullExpression", "BindExpression", "TaggedTemplateExpression", "UnaryExpression", "LogicalExpression",
			"AwaitExpression", "TSTypeAssertion", "MatchExpressionCase":
			return true

		case "TSInstantiationExpression":
			return key == "expression"

		case "ConditionalExpression":
			return key == "test"

		default:
			return false
		}

	case "ClassExpression":
		switch parent.Type() {
		case "NewExpression":
			return key == "callee"
		case "ExportDefaultDeclaration":
			return key == "declaration"
		default:
			return false
		}
	case "OptionalMemberExpression", "OptionalCallExpression", "ChainExpression", "TSNonNullExpression":
		if shouldAddParenthesesToChainElement(path) {
			return true
		}
		fallthrough
	case "CallExpression", "MemberExpression", "TaggedTemplateExpression", "ImportExpression":
		if key == "callee" && parent.Is("BindExpression", "NewExpression") {
			object := current
			for object != nil {
				switch object.Type() {
				case "CallExpression", "ImportExpression":
					return true
				case "MemberExpression", "OptionalMemberExpression", "BindExpression":
					object = object.Child("object")
				// tagged templates are basically member expressions from a grammar perspective
				// see https://tc39.github.io/ecma262/#prod-MemberExpression
				case "TaggedTemplateExpression":
					object = object.Child("tag")
				case "TSNonNullExpression":
					object = object.Child("expression")
				default:
					return false
				}
			}
		}

		return false

	case "BindExpression":
		return key == "callee" && parent.Is("BindExpression", "NewExpression") ||
			key == "object" && isMemberExpression(parent)
	case "NGPipeExpression":
		if parent.Is("NGRoot") ||
			parent.Is("NGMicrosyntaxExpression") ||
			parent.Is("ObjectProperty") &&
				// Preserve parens for compatibility with AngularJS expressions
				!current.Parenthesized ||
			isArrayExpression(parent) ||
			key == "arguments" && isCallExpression(parent) ||
			key == "right" && parent.Is("NGPipeExpression") ||
			key == "property" && parent.Is("MemberExpression") ||
			parent.Is("AssignmentExpression") {
			return false
		}
		return true
	case "JSXFragment", "JSXElement":
		return key == "callee" ||
			key == "left" && parent.Is("BinaryExpression") && parent.String("operator") == "<" ||
			!isArrayExpression(parent) &&
				!parent.Is("ArrowFunctionExpression") &&
				!parent.Is("AssignmentExpression") &&
				!parent.Is("AssignmentPattern") &&
				!parent.Is("BinaryExpression") &&
				!parent.Is("ConditionalExpression") &&
				!parent.Is("ExpressionStatement") &&
				!parent.Is("JsExpressionRoot") &&
				!parent.Is("JSXAttribute") &&
				!parent.Is("JSXElement") &&
				!parent.Is("JSXExpressionContainer") &&
				!parent.Is("JSXFragment") &&
				!parent.Is("LogicalExpression") &&
				!isCallOrNewExpression(parent) &&
				!isObjectProperty(parent) &&
				!isReturnOrThrowStatement(parent) &&
				!parent.Is("TypeCastExpression") &&
				!parent.Is("VariableDeclarator") &&
				!parent.Is("YieldExpression") &&
				!parent.Is("MatchExpressionCase") &&
				!(key == "declaration" && parent.Is("ExportDefaultDeclaration"))

	case "TSInstantiationExpression":
		return key == "object" && isMemberExpression(parent)

	case "MatchOrPattern":
		return parent.Is("MatchAsPattern")
	}

	return false
}

// keyIsType is the path.match predicate upstream writes inline as
// `(node, key) => key === name && node.type === type`.
func keyIsType(name string, nodeType string) printing.Predicate[Node] {
	return func(value any, key any, _ int, _ bool) bool {
		node, _ := value.(Node)
		return key == name && node.Is(nodeType)
	}
}

// isStatement is upstream's isStatement.
func isStatement(node Node) bool {
	return node.Is("BlockStatement", "BreakStatement", "ComponentDeclaration", "ClassBody", "ClassDeclaration",
		"ClassMethod", "ClassProperty", "PropertyDefinition", "ClassPrivateProperty", "ContinueStatement",
		"DebuggerStatement", "DeclareComponent", "DeclareClass", "DeclareExportAllDeclaration",
		"DeclareExportDeclaration", "DeclareFunction", "DeclareHook", "DeclareInterface", "DeclareModule",
		"DeclareModuleExports", "DeclareNamespace", "DeclareVariable", "DeclareEnum", "DoWhileStatement",
		"EnumDeclaration", "ExportAllDeclaration", "ExportDefaultDeclaration", "ExportNamedDeclaration",
		"ExpressionStatement", "ForInStatement", "ForOfStatement", "ForStatement", "FunctionDeclaration",
		"HookDeclaration", "IfStatement", "ImportDeclaration", "InterfaceDeclaration", "LabeledStatement",
		"MethodDefinition", "ReturnStatement", "SwitchStatement", "ThrowStatement", "TryStatement",
		"TSDeclareFunction", "TSEnumDeclaration", "TSImportEqualsDeclaration", "TSInterfaceDeclaration",
		"TSModuleDeclaration", "TSNamespaceExportDeclaration", "TypeAlias", "VariableDeclaration", "WhileStatement",
		"WithStatement")
}

// isPathInForStatementInitializer is upstream's isPathInForStatementInitializer.
func isPathInForStatementInitializer(path *Path) bool {
	index := 0
	current := node(path)
	for current != nil {
		parent := nodeAt(path, index+1)
		index++
		if parent.Is("ForStatement") && parent.Child("init") == current {
			return true
		}
		current = parent
	}

	return false
}
