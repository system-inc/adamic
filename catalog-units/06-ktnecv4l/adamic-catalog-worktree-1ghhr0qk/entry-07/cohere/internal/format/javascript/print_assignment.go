package javascript

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
)

// print/assignment.js. experimentalTernaries is always false, so the branches that read it keep only
// the false side.

// printAssignment is upstream's printAssignment. operator is `any` because upstream passes both a
// string (" =", ":") and an array ([" ", node.operator]). An absent rightPropertyName is "".
func printAssignment(
	path *Path,
	options *Options,
	print PrintFunc,
	leftDoc Doc,
	operator any,
	rightPropertyName string,
) Doc {
	layout := chooseLayout(path, options, print, leftDoc, rightPropertyName)

	var rightDoc Doc = emptyDoc
	if rightPropertyName != "" {
		rightDoc = print(rightPropertyName, &printArguments{assignmentLayout: layout})
	}

	switch layout {
	// First break after operator, then the sides are broken independently on their own lines
	case "break-after-operator":
		return groupIn(path, concatIn(path, groupIn(path, leftDoc), operator, groupIn(path, indentIn(path, concatIn(path, line, rightDoc)))))

	// First break right-hand side, then left-hand side
	case "never-break-after-operator":
		return groupIn(path, concatIn(path, groupIn(path, leftDoc), operator, " ", rightDoc))

	// First break right-hand side, then after operator
	case "fluid":
		groupID := newGroupID("assignment")
		return groupIn(path, concatIn(path, groupIn(path, leftDoc),
			operator,
			groupWithIn(path, indentIn(path, line), doc.GroupOptions{ID: groupID}),
			lineSuffixBoundary,
			indentIfBreak(rightDoc, groupID, false),
		))

	case "break-lhs":
		return groupIn(path, concatIn(path, leftDoc, operator, " ", groupIn(path, rightDoc)))

	// Parts of assignment chains aren't wrapped in groups.
	// Once one of them breaks, the chain breaks too.
	case "chain":
		return concatIn(path, groupIn(path, leftDoc), operator, line, rightDoc)

	case "chain-tail":
		return concatIn(path, groupIn(path, leftDoc), operator, indentIn(path, concatIn(path, line, rightDoc)))

	case "chain-tail-arrow-chain":
		return concatIn(path, groupIn(path, leftDoc), operator, rightDoc)

	case "only-left":
		return leftDoc
	}
	// Upstream falls off the switch and returns undefined; chooseLayout never returns another layout.
	panic("javascript: unknown assignment layout " + layout)
}

// printAssignmentExpression is upstream's printAssignmentExpression.
func printAssignmentExpression(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	return printAssignment(
		path,
		options,
		print,
		print("left", nil),
		concatIn(path, " ", current.String("operator")),
		"right",
	)
}

// printVariableDeclarator is upstream's printVariableDeclarator.
func printVariableDeclarator(path *Path, options *Options, print PrintFunc) Doc {
	return printAssignment(path, options, print, print("id", nil), " =", "init")
}

// chooseLayout is upstream's chooseLayout.
func chooseLayout(path *Path, options *Options, print PrintFunc, leftDoc Doc, rightPropertyName string) string {
	current := node(path)
	var rightNode Node
	if rightPropertyName != "" {
		rightNode = current.Child(rightPropertyName)
	}

	if rightNode == nil {
		return "only-left"
	}

	// Short assignment chains (only 2 segments) are NOT formatted as chains.
	//   1) a = b = c; (expression statements)
	//   2) var/let/const a = b = c;

	isTail := !isAssignment(rightNode)
	shouldUseChainFormatting := path.Match(
		func(value any, _ any, _ int, _ bool) bool {
			matched, _ := value.(Node)
			return isAssignment(matched)
		},
		func(value any, _ any, _ int, _ bool) bool {
			matched, _ := value.(Node)
			return isAssignmentOrVariableDeclarator(matched)
		},
		func(value any, _ any, _ int, _ bool) bool {
			matched, _ := value.(Node)
			return !isTail ||
				(!matched.Is("ExpressionStatement") &&
					!matched.Is("VariableDeclaration"))
		},
	)
	if shouldUseChainFormatting {
		if !isTail {
			return "chain"
		}
		if rightNode.Is("ArrowFunctionExpression") &&
			rightNode.Child("body").Is("ArrowFunctionExpression") {
			return "chain-tail-arrow-chain"
		}
		return "chain-tail"
	}
	isHeadOfLongChain := !isTail && isAssignment(rightNode.Child("right"))

	if isHeadOfLongChain ||
		(isUnionType(rightNode) && !shouldHugUnionType(rightNode)) ||
		hasLeadingOwnLineComment(originalText(options), rightNode) ||
		hasComment(rightNode, commentLeading, estree.IsIndentableBlockComment) {
		return "break-after-operator"
	}

	root, _ := path.Root()
	if current.Is("ImportAttribute") ||
		(rightNode.Is("CallExpression") &&
			rightNode.Child("callee").String("name") == "require") ||
		// do not put values on a separate line from the key in JSON
		root.Is("JsonRoot") {
		return "never-break-after-operator"
	}

	canBreakLeftDoc := doc.CanBreak(leftDoc)

	if isComplexDestructuring(current) ||
		hasComplexTypeAnnotation(current) ||
		(isArrowFunctionVariableDeclarator(current) && canBreakLeftDoc) {
		return "break-lhs"
	}

	// wrapping object properties with very short keys usually doesn't add much value
	hasShortKey := isObjectPropertyWithShortKey(current, leftDoc, options)

	if call(path, func(path *Path) bool {
		return shouldBreakAfterOperator(path, options, print, hasShortKey)
	}, rightPropertyName) {
		return "break-after-operator"
	}

	if isComplexTypeAliasParams(current) {
		return "break-lhs"
	}

	if !canBreakLeftDoc &&
		(hasShortKey ||
			rightNode.Is("TemplateLiteral") ||
			rightNode.Is("TaggedTemplateExpression") ||
			isBooleanLiteral(rightNode) ||
			isNumericLiteral(rightNode) ||
			rightNode.Is("ClassExpression")) {
		return "never-break-after-operator"
	}

	return "fluid"
}

// shouldBreakAfterOperator is upstream's shouldBreakAfterOperator.
func shouldBreakAfterOperator(path *Path, options *Options, print PrintFunc, hasShortKey bool) bool {
	rightNode := node(path)

	if isBinaryish(rightNode) && !shouldInlineLogicalExpression(rightNode) {
		return true
	}

	switch rightNode.Type() {
	case "StringLiteralTypeAnnotation", "SequenceExpression":
		return true
	case "TSConditionalType", "ConditionalTypeAnnotation":
		// experimentalTernaries is always false.
		if !shouldBreakBeforeConditionalType(rightNode) {
			break
		}
		return true
	case "ConditionalExpression":
		// experimentalTernaries is always false; upstream's other branch is dropped.
		test := rightNode.Child("test")
		return isBinaryish(test) && !shouldInlineLogicalExpression(test)
	case "ClassExpression":
		return len(rightNode.List("decorators")) > 0
	}

	if hasShortKey {
		return false
	}

	current := rightNode
	var propertiesForPath []any
	for {
		if current.Is("UnaryExpression") ||
			current.Is("AwaitExpression") ||
			(current.Is("YieldExpression") && current.Child("argument") != nil) {
			current = current.Child("argument")
			propertiesForPath = append(propertiesForPath, "argument")
		} else if current.Is("TSNonNullExpression") {
			current = current.Child("expression")
			propertiesForPath = append(propertiesForPath, "expression")
		} else {
			break
		}
	}
	if isStringLiteral(current) ||
		call(path, func(path *Path) bool {
			return isPoorlyBreakableMemberOrCallChain(path, options, print, false)
		}, propertiesForPath...) {
		return true
	}

	return false
}

// isComplexDestructuring is upstream's isComplexDestructuring.
//
// prefer to break destructuring assignment
// if it includes default values or non-shorthand properties
func isComplexDestructuring(current Node) bool {
	if isAssignmentOrVariableDeclarator(current) {
		leftNode := current.Child("left")
		if leftNode == nil {
			leftNode = current.Child("id")
		}
		if !leftNode.Is("ObjectPattern") {
			return false
		}
		properties := leftNode.List("properties")
		if len(properties) <= 2 {
			return false
		}
		for _, property := range properties {
			if isObjectProperty(property) &&
				(!property.Bool("shorthand") || property.Child("value").Is("AssignmentPattern")) {
				return true
			}
		}
		return false
	}
	return false
}

// isAssignment is upstream's isAssignment.
func isAssignment(current Node) bool {
	return current.Is("AssignmentExpression")
}

// isAssignmentOrVariableDeclarator is upstream's isAssignmentOrVariableDeclarator.
func isAssignmentOrVariableDeclarator(current Node) bool {
	return isAssignment(current) || current.Is("VariableDeclarator")
}

// isComplexTypeAliasParams is upstream's isComplexTypeAliasParams.
func isComplexTypeAliasParams(current Node) bool {
	typeParams := getTypeParametersFromTypeAlias(current)
	if len(typeParams) > 0 {
		constraintPropertyName := "bound"
		if current.Is("TSTypeAliasDeclaration") {
			constraintPropertyName = "constraint"
		}
		if len(typeParams) > 1 {
			for _, param := range typeParams {
				if param.Child(constraintPropertyName) != nil || param.Child("default") != nil {
					return true
				}
			}
		}
	}
	return false
}

// getTypeParametersFromTypeAlias is upstream's getTypeParametersFromTypeAlias.
func getTypeParametersFromTypeAlias(current Node) []Node {
	if isTypeAlias(current) {
		return current.Child("typeParameters").List("params")
	}
	return nil
}

// hasComplexTypeAnnotation is upstream's hasComplexTypeAnnotation.
func hasComplexTypeAnnotation(current Node) bool {
	if !current.Is("VariableDeclarator") {
		return false
	}
	typeAnnotation := current.Child("id").Child("typeAnnotation")
	if typeAnnotation == nil || typeAnnotation.Child("typeAnnotation") == nil {
		return false
	}
	typeParams := getTypeParametersFromTypeReference(typeAnnotation.Child("typeAnnotation"))
	if len(typeParams) <= 1 {
		return false
	}
	for _, param := range typeParams {
		if len(getTypeParametersFromTypeReference(param)) > 0 || param.Is("TSConditionalType") {
			return true
		}
	}
	return false
}

// isArrowFunctionVariableDeclarator is upstream's isArrowFunctionVariableDeclarator.
func isArrowFunctionVariableDeclarator(current Node) bool {
	return current.Is("VariableDeclarator") &&
		current.Child("init").Is("ArrowFunctionExpression")
}

// getTypeParametersFromTypeReference is upstream's getTypeParametersFromTypeReference.
func getTypeParametersFromTypeReference(current Node) []Node {
	var typeArguments Node
	switch current.Type() {
	case "GenericTypeAnnotation":
		typeArguments = current.Child("typeParameters")
	case "TSTypeReference":
		typeArguments = current.Child("typeArguments")
	}
	return typeArguments.List("params")
}

// isPoorlyBreakableMemberOrCallChain is upstream's isPoorlyBreakableMemberOrCallChain.
//
// A chain with no calls at all or whose calls are all without arguments or with lone short arguments,
// excluding chains printed by `printMemberChain`
func isPoorlyBreakableMemberOrCallChain(path *Path, options *Options, print PrintFunc, deep bool) bool {
	current := node(path)
	goDeeper := func(path *Path) bool {
		return isPoorlyBreakableMemberOrCallChain(path, options, print, true)
	}

	if isChainElementWrapper(current) {
		return call(path, goDeeper, "expression")
	}

	if isCallExpression(current) {
		// Upstream tests `doc.label?.memberChain`; printMemberChain labels its doc "member-chain" here.
		printed := printCallExpression(path, options, print)
		if labelOf(printed) == "member-chain" {
			return false
		}

		args := getCallArguments(current)
		isPoorlyBreakableCall := len(args) == 0 ||
			(len(args) == 1 && isLoneShortArgument(args[0], options))
		if !isPoorlyBreakableCall {
			return false
		}

		if isCallExpressionWithComplexTypeArguments(current, print) {
			return false
		}

		return call(path, goDeeper, "callee")
	}

	if isMemberExpression(current) {
		return call(path, goDeeper, "object")
	}

	return deep && (current.Is("Identifier") || current.Is("ThisExpression"))
}

// isObjectPropertyWithShortKey is upstream's isObjectPropertyWithShortKey.
func isObjectPropertyWithShortKey(current Node, keyDoc Doc, options *Options) bool {
	if !isObjectProperty(current) {
		return false
	}
	// TODO: for performance, it might make sense to use a more lightweight
	// version of cleanDoc, such that it would stop once it detects that
	// the doc can't be reduced to a string.
	keyDoc = doc.CleanDoc(keyDoc)
	const minOverlapForBreak = 3
	//   ↓↓ - insufficient overlap for a line break
	// key1: longValue1,
	//   ↓↓↓↓↓↓ - overlap is long enough to break
	// key2abcd:
	//   longValue2
	//
	// Upstream's `typeof keyDoc === "string"`; getStringWidth is doc.StringWidth.
	text, isText := keyDoc.(doc.Text)
	return isText &&
		doc.StringWidth(string(text)) < settingsOf(options).TabWidth+minOverlapForBreak
}

// isCallExpressionWithComplexTypeArguments is upstream's isCallExpressionWithComplexTypeArguments.
func isCallExpressionWithComplexTypeArguments(current Node, print PrintFunc) bool {
	typeArgs := current.Child("typeArguments").List("params")
	if len(typeArgs) > 0 {
		if len(typeArgs) > 1 {
			return true
		}
		if len(typeArgs) == 1 {
			firstArg := typeArgs[0]
			if isUnionType(firstArg) ||
				isIntersectionType(firstArg) ||
				firstArg.Is("TSTypeLiteral") ||
				firstArg.Is("ObjectTypeAnnotation") {
				return true
			}
		}
		if doc.WillBreak(print("typeArguments", nil)) {
			return true
		}
	}
	return false
}

// isGeneric is upstream's isGeneric.
func isGeneric(current Node) bool {
	switch current.Type() {
	case "FunctionTypeAnnotation", "GenericTypeAnnotation", "TSFunctionType":
		return current.Child("typeParameters") != nil
	case "TSTypeReference":
		return current.Child("typeArguments") != nil
	default:
		return false
	}
}

// shouldBreakBeforeConditionalType is upstream's shouldBreakBeforeConditionalType.
func shouldBreakBeforeConditionalType(current Node) bool {
	return isGeneric(current.Child("checkType")) || isGeneric(current.Child("extendsType"))
}
