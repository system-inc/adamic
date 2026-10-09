package javascript

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// print/index.js, print/estree.js and print/typescript.js: the dispatch from a node to its printer.
// The JSX, Flow and Angular printers come before these upstream; Flow and Angular never match a
// TypeScript tree, so only JSX is consulted.

// shouldPrintDirectly is upstream's shouldPrintDirectly: class members print their own decorators and
// never need parentheses or a leading semicolon.
func shouldPrintDirectly(node Node) bool {
	return node.Is("ClassMethod", "ClassPrivateMethod", "ClassProperty", "ClassAccessorProperty", "AccessorProperty",
		"TSAbstractAccessorProperty", "PropertyDefinition", "TSAbstractPropertyDefinition", "ClassPrivateProperty",
		"MethodDefinition", "TSAbstractMethodDefinition", "TSDeclareMethod")
}

// printWithoutParentheses is upstream's printWithoutParentheses.
func printWithoutParentheses(path *Path, options *Options, print PrintFunc, args any) Doc {
	if printed := printJsx(path, options, print); printed != nil {
		return printed
	}
	if printed := printTypescript(path, options, print, args); printed != nil {
		return printed
	}
	return printEstreeNode(path, options, print, args)
}

// printEstree is upstream's print, print/index.js: the printer's entry for every node.
func printEstree(path *Path, options *Options, print PrintFunc, args any) Doc {
	current := node(path)

	var printed Doc
	if isIgnored(path) {
		printed = printIgnored(path, options)
	} else {
		printed = printWithoutParentheses(path, options, print, args)
	}
	if printed == nil || isEmptyString(printed) {
		return emptyDoc
	}

	if shouldPrintDirectly(current) {
		return printed
	}

	printed = printCommentsForFunction(path, options, printed)

	var decoratorsDoc Doc
	if !current.Is("ClassExpression") && len(current.List("decorators")) > 0 {
		decoratorsDoc = printDecorators(path, options, print)
	}
	hasDecorators := decoratorsDoc != nil && !isEmptyString(decoratorsDoc)

	needsParens := needsParentheses(path, options)
	if !hasDecorators && !needsParens {
		return printed
	}

	return doc.InheritLabel(printed, func(printed Doc) Doc {
		open, close := "", ""
		if needsParens {
			open, close = "(", ")"
		}
		if hasDecorators {
			return concatIn(path, open, groupIn(path, concatIn(path, decoratorsDoc, printed)), close)
		}
		return concatIn(path, open, printed, close)
	})
}

// printCommentsForFunction is upstream's printCommentsForFunction.
func printCommentsForFunction(path *Path, options *Options, printed Doc) Doc {
	current := node(path)
	if (hasComment(current, commentLeading, nil) || hasComment(current, commentTrailing, nil)) &&
		isIifeCalleeOrTaggedTemplateExpressionTag(path) {
		return concatIn(path, indentIn(path, concatIn(path, softline, printing.PrintComments(path, printed, options, nil))), softline)
	}
	return printed
}

var lowercaseLetterEnd = regexp.MustCompile(`[a-z]$`)

// printEstreeNode is upstream's printEstree, print/estree.js.
func printEstreeNode(path *Path, options *Options, print PrintFunc, args any) Doc {
	current := node(path)

	if isLiteral(current) {
		return printLiteral(path, options)
	}

	switch current.Type() {
	case "JsonRoot":
		return concatIn(path, printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{}), print("node", nil), hardline)
	case "ExpressionStatement":
		return printExpressionStatement(path, options, print)
	case "ChainExpression":
		return print("expression", nil)
	// Babel non-standard node. Used for Closure-style type casts. See postprocess.js.
	case "ParenthesizedExpression":
		expression := current.Child("expression")
		shouldHug := !hasAnyComment(expression) && (isObjectExpression(expression) || isArrayExpression(expression))
		if shouldHug {
			return concatIn(path, "(", print("expression", nil), ")")
		}
		return groupIn(path, concatIn(path, "(", indentIn(path, concatIn(path, softline, print("expression", nil))), softline, ")"))
	case "AssignmentExpression":
		return printAssignmentExpression(path, options, print)
	case "VariableDeclarator":
		return printVariableDeclarator(path, options, print)
	case "BinaryExpression", "LogicalExpression":
		return printBinaryishExpression(path, options, print)
	case "AssignmentPattern":
		return concatIn(path, print("left", nil), " = ", print("right", nil))
	case "OptionalMemberExpression", "MemberExpression":
		return printMemberExpression(path, options, print)
	case "MetaProperty":
		return concatIn(path, print("meta", nil), ".", print("property", nil))
	case "Identifier":
		return concatIn(path, current.String("name"),
			printOptionalToken(path),
			printDefiniteToken(path),
			printTypeAnnotationProperty(path, print),
		)
	case "SpreadElement":
		return printSpreadElement(path, print)
	case "RestElement":
		return printRestElement(path, print)
	case "FunctionDeclaration", "FunctionExpression":
		return printFunction(path, options, print, argsOf(args))
	case "ArrowFunctionExpression":
		return printArrowFunction(path, options, print, argsOf(args))
	case "YieldExpression":
		keyword := "yield"
		if current.Bool("delegate") {
			keyword = "yield*"
		}
		var argument Doc = emptyDoc
		if current.Child("argument") != nil {
			argument = concatIn(path, " ", print("argument", nil))
		}
		return concatIn(path, keyword, argument)
	case "AwaitExpression":
		return printAwaitExpression(path, options, print)
	case "ExportDefaultDeclaration", "ExportNamedDeclaration", "ExportAllDeclaration":
		return printExportDeclaration(path, options, print)
	case "ImportDeclaration":
		return printImportDeclaration(path, options, print)
	case "ImportSpecifier", "ExportSpecifier", "ImportNamespaceSpecifier", "ExportNamespaceSpecifier",
		"ImportDefaultSpecifier", "ExportDefaultSpecifier":
		return printModuleSpecifier(path, options, print)
	case "ImportAttribute":
		return printProperty(path, options, print)
	case "Program", "BlockStatement", "StaticBlock":
		return printBlock(path, options, print)
	case "ClassBody":
		return printClassBody(path, options, print)
	case "ThrowStatement":
		return printThrowStatement(path, options, print)
	case "ReturnStatement":
		return printReturnStatement(path, options, print)
	case "NewExpression", "ImportExpression", "OptionalCallExpression", "CallExpression":
		return printCallExpression(path, options, print)
	case "ObjectExpression", "ObjectPattern":
		return printObject(path, options, print)
	case "Property":
		if isMethod(current) {
			return printMethod(path, options, print)
		}
		return printProperty(path, options, print)
	// Babel, which the JSON parsers produce.
	case "ObjectProperty":
		return printProperty(path, options, print)
	case "Decorator":
		return concatIn(path, "@", print("expression", nil))
	case "ArrayExpression", "ArrayPattern":
		return printArray(path, options, print)
	case "SequenceExpression":
		return printSequenceExpression(path, options, print)
	case "ThisExpression":
		return doc.Text("this")
	case "Super":
		return doc.Text("super")
	case "Directive":
		return concatIn(path, print("value", nil), printSemicolon(options))
	case "UnaryExpression":
		operator := current.String("operator")
		parts := []any{operator}
		if lowercaseLetterEnd.MatchString(operator) {
			parts = append(parts, " ")
		}
		argumentDoc := print("argument", nil)
		if hasAnyComment(current.Child("argument")) {
			parts = append(parts, groupIn(path, concatIn(path, "(", indentIn(path, concatIn(path, softline, argumentDoc)), softline, ")")))
		} else {
			parts = append(parts, argumentDoc)
		}
		return concatIn(path, parts...)
	case "UpdateExpression":
		if current.Bool("prefix") {
			return concatIn(path, current.String("operator"), print("argument", nil), "")
		}
		return concatIn(path, "", print("argument", nil), current.String("operator"))
	case "ConditionalExpression":
		return printTernary(path, options, print, argsOf(args))
	case "VariableDeclaration":
		return printVariableDeclaration(path, options, print)
	case "IfStatement":
		return printIfStatement(path, options, print)
	case "ForStatement":
		return printForStatement(path, options, print)
	case "WithStatement", "WhileStatement":
		return printWhileStatement(path, options, print)
	case "DoWhileStatement":
		return printDoWhileStatement(path, options, print)
	case "ForInStatement", "ForOfStatement":
		return printForXStatement(path, options, print)
	case "BreakStatement", "ContinueStatement":
		keyword := "continue"
		if current.Is("BreakStatement") {
			keyword = "break"
		}
		var labelDoc Doc = emptyDoc
		if current.Child("label") != nil {
			labelDoc = concatIn(path, " ", print("label", nil))
		}
		return concatIn(path, keyword, labelDoc, printSemicolon(options))
	case "LabeledStatement":
		separator := ": "
		if current.Child("body").Is("EmptyStatement") && !hasComment(current.Child("body"), commentLeading, nil) {
			separator = ":"
		}
		return concatIn(path, print("label", nil), separator, print("body", nil))
	case "TryStatement":
		return printTryStatement(path, options, print)
	case "CatchClause":
		return printCatchClause(path, options, print)
	case "SwitchStatement":
		return printSwitchStatement(path, options, print)
	case "SwitchCase":
		return printSwitchCase(path, options, print)
	case "DebuggerStatement":
		return concatIn(path, "debugger", printSemicolon(options))
	case "ClassDeclaration", "ClassExpression":
		return printClass(path, options, print)
	case "ClassMethod", "ClassPrivateMethod", "MethodDefinition":
		return printClassMethod(path, options, print)
	case "ClassProperty", "PropertyDefinition", "ClassPrivateProperty", "ClassAccessorProperty", "AccessorProperty":
		return printClassProperty(path, options, print)
	case "TemplateElement":
		return doc.ReplaceEndOfLine(doc.Text(current.Get("value").(*estree.TemplateValue).Raw), nil)
	case "TemplateLiteral":
		return printTemplateLiteral(path, options, print)
	case "TaggedTemplateExpression":
		return printTaggedTemplateExpression(path, options, print)
	case "PrivateIdentifier":
		return concatIn(path, "#", current.String("name"))
	case "EmptyStatement":
		if isMeaningfulEmptyStatement(path) {
			return doc.Text(";")
		}
	}
	panic(fmt.Sprintf("unknown ESTree node type %q", current.Type()))
}

// printTypescript is upstream's printTypescript, print/typescript.js. It returns nil for a node that
// is not TypeScript, where upstream returns undefined.
func printTypescript(path *Path, options *Options, print PrintFunc, args any) Doc {
	current := node(path)
	nodeType := current.Type()
	if !strings.HasPrefix(nodeType, "TS") {
		return nil
	}
	if isTsKeywordType(current) {
		return doc.Text(strings.ToLower(nodeType[2 : len(nodeType)-7]))
	}

	switch nodeType {
	case "TSThisType":
		return doc.Text("this")
	case "TSTypeAssertion":
		return printTypeAssertion(path, options, print)
	case "TSDeclareFunction":
		return printFunction(path, options, print, nil)
	case "TSExportAssignment":
		return concatIn(path, "export = ", print("expression", nil), printSemicolon(options))
	case "TSModuleBlock":
		return printBlock(path, options, print)
	case "TSInterfaceBody", "TSTypeLiteral":
		return printClassBody(path, options, print)
	case "TSTypeAliasDeclaration":
		return printTypeAlias(path, options, print)
	case "TSQualifiedName":
		return concatIn(path, print("left", nil), ".", print("right", nil))
	case "TSAbstractMethodDefinition", "TSDeclareMethod":
		return printClassMethod(path, options, print)
	case "TSAbstractAccessorProperty", "TSAbstractPropertyDefinition":
		return printClassProperty(path, options, print)
	case "TSInterfaceHeritage", "TSClassImplements", "TSInstantiationExpression":
		return concatIn(path, print("expression", nil), print("typeArguments", nil))
	case "TSTemplateLiteralType":
		return printTemplateLiteral(path, options, print)
	case "TSNamedTupleMember":
		return printNamedTupleMember(path, options, print)
	case "TSRestType":
		return printRestType(path, options, print)
	case "TSOptionalType":
		return concatIn(path, print("typeAnnotation", nil), "?")
	case "TSInterfaceDeclaration":
		return printClass(path, options, print)
	case "TSTypeParameterDeclaration", "TSTypeParameterInstantiation":
		return printTypeParameters(path, options, print, "params")
	case "TSTypeParameter":
		return printTypeParameter(path, options, print)
	case "TSAsExpression", "TSSatisfiesExpression":
		return printBinaryCastExpression(path, options, print)
	case "TSArrayType":
		return printArrayType(print)
	case "TSPropertySignature":
		readonly := ""
		if current.Bool("readonly") {
			readonly = "readonly "
		}
		return concatIn(path, readonly,
			printKey(path, options, print),
			printOptionalToken(path),
			printTypeAnnotationProperty(path, print),
			printClassMemberSemicolon(path, options),
		)
	case "TSParameterProperty":
		static, override, readonly := "", "", ""
		if current.Bool("static") {
			static = "static "
		}
		if current.Bool("override") {
			override = "override "
		}
		if current.Bool("readonly") {
			readonly = "readonly "
		}
		return concatIn(path, printTypeScriptAccessibilityToken(current), static, override, readonly, print("parameter", nil))
	case "TSTypeQuery":
		return printTypeQuery(path, print)
	case "TSIndexSignature":
		return printIndexSignature(path, options, print)
	case "TSTypePredicate":
		return printTypePredicate(path, print)
	case "TSNonNullExpression":
		return concatIn(path, print("expression", nil), "!")
	case "TSImportType":
		var qualifier Doc = emptyDoc
		if current.Child("qualifier") != nil {
			qualifier = concatIn(path, ".", print("qualifier", nil))
		}
		return concatIn(path, printCallExpression(path, options, print),
			qualifier,
			printTypeParameters(path, options, print, "typeArguments"),
		)
	case "TSLiteralType":
		return print("literal", nil)
	case "TSIndexedAccessType":
		return printIndexedAccessType(path, options, print)
	case "TSTypeOperator":
		return concatIn(path, current.String("operator"), " ", print("typeAnnotation", nil))
	case "TSMappedType":
		return printTypeScriptMappedType(path, options, print)
	case "TSMethodSignature":
		return printMethodSignature(path, options, print)
	case "TSNamespaceExportDeclaration":
		return concatIn(path, "export as namespace ", print("id", nil), printSemicolon(options))
	case "TSEnumDeclaration":
		return printEnumDeclaration(path, print)
	case "TSEnumBody":
		return printObject(path, options, print)
	case "TSEnumMember":
		return printEnumMember(path, options, print)
	case "TSImportEqualsDeclaration":
		return concatIn(path, "import ",
			printImportKind(current, false),
			print("id", nil),
			" = ",
			print("moduleReference", nil),
			printSemicolon(options),
		)
	case "TSExternalModuleReference":
		return printCallExpression(path, options, print)
	case "TSModuleDeclaration":
		return printModuleDeclaration(path, options, print)
	case "TSConditionalType":
		return printTernary(path, options, print, nil)
	case "TSInferType":
		return printInferType(path, options, print)
	case "TSIntersectionType":
		return printIntersectionType(path, options, print)
	case "TSUnionType":
		return printUnionType(path, options, print, argsOf(args))
	case "TSFunctionType", "TSCallSignatureDeclaration", "TSConstructorType", "TSConstructSignatureDeclaration":
		return printFunctionType(path, options, print)
	case "TSTupleType":
		return printArray(path, options, print)
	case "TSTypeReference":
		return concatIn(path, print("typeName", nil), printTypeParameters(path, options, print, "typeArguments"))
	case "TSTypeAnnotation":
		return printTypeAnnotation(path, options, print)
	case "TSEmptyBodyFunctionExpression":
		return printMethodValue(path, options, print)
	}
	panic(fmt.Sprintf("unknown TypeScript node type %q", nodeType))
}
