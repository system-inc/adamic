package javascript

import "strings"

// utilities/node-types.js, utilities/literal.js, utilities/get-raw.js and the small type predicates
// built on them. Flow and Babel type names stay in the lists, as upstream has them, so a list reads
// against its source even though only the TypeScript and ESTree names occur here.

func isBinaryCastExpression(node Node) bool {
	return node.Is("TSAsExpression", "TSSatisfiesExpression", "AsExpression", "AsConstExpression", "SatisfiesExpression")
}
func isSatisfiesExpression(node Node) bool {
	return node.Is("SatisfiesExpression", "TSSatisfiesExpression")
}
func isUnionType(node Node) bool { return node.Is("TSUnionType", "UnionTypeAnnotation") }
func isIntersectionType(node Node) bool {
	return node.Is("TSIntersectionType", "IntersectionTypeAnnotation")
}
func isTupleType(node Node) bool { return node.Is("TupleTypeAnnotation", "TSTupleType") }
func isConditionalType(node Node) bool {
	return node.Is("TSConditionalType", "ConditionalTypeAnnotation")
}
func isTypeAlias(node Node) bool { return node.Is("TSTypeAliasDeclaration", "TypeAlias") }
func isReturnOrThrowStatement(node Node) bool {
	return node.Is("ReturnStatement", "ThrowStatement")
}
func isExportDeclaration(node Node) bool {
	return node.Is("ExportDefaultDeclaration", "DeclareExportDeclaration", "ExportNamedDeclaration",
		"ExportAllDeclaration", "DeclareExportAllDeclaration")
}
func isArrayExpression(node Node) bool  { return node.Is("ArrayExpression") }
func isObjectExpression(node Node) bool { return node.Is("ObjectExpression") }
func isLiteral(node Node) bool {
	return node.Is("Literal", "BooleanLiteral", "BigIntLiteral", "DirectiveLiteral", "NullLiteral", "NumericLiteral",
		"RegExpLiteral", "StringLiteral")
}
func isObjectType(node Node) bool {
	return node.Is("ObjectTypeAnnotation", "TSTypeLiteral", "TSMappedType")
}
func isFunctionOrArrowExpression(node Node) bool {
	return node.Is("FunctionExpression", "ArrowFunctionExpression")
}
func isJsxElement(node Node) bool { return node.Is("JSXElement", "JSXFragment") }
func isBinaryish(node Node) bool {
	return node.Is("BinaryExpression", "LogicalExpression", "NGPipeExpression")
}
func isCallExpression(node Node) bool {
	return node.Is("CallExpression", "OptionalCallExpression")
}
func isMemberExpression(node Node) bool {
	return node.Is("MemberExpression", "OptionalMemberExpression")
}
func isCallOrNewExpression(node Node) bool {
	return node.Is("CallExpression", "OptionalCallExpression", "NewExpression")
}
func isCallLikeExpression(node Node) bool {
	return node.Is("CallExpression", "OptionalCallExpression", "NewExpression", "ImportExpression")
}
func isChainElementWrapper(node Node) bool { return node.Is("ChainExpression", "TSNonNullExpression") }
func isArrayType(node Node) bool           { return node.Is("TSArrayType", "ArrayTypeAnnotation") }
func isTypeParameterInstantiation(node Node) bool {
	return node.Is("TSTypeParameterInstantiation", "TypeParameterInstantiation")
}

// isBigIntLiteral is upstream's isBigIntLiteral.
func isBigIntLiteral(node Node) bool {
	return node.Is("BigIntLiteral") || node.Is("Literal") && node.Truthy("bigint")
}

// isBooleanLiteral is upstream's isBooleanLiteral.
func isBooleanLiteral(node Node) bool {
	if node.Is("BooleanLiteral") {
		return true
	}
	_, isBool := node.Get("value").(bool)
	return node.Is("Literal") && isBool
}

// isNumericLiteral is upstream's isNumericLiteral.
func isNumericLiteral(node Node) bool {
	if node.Is("NumericLiteral") {
		return true
	}
	_, isNumber := node.Get("value").(float64)
	return node.Is("Literal") && isNumber
}

// isRegExpLiteral is upstream's isRegExpLiteral.
func isRegExpLiteral(node Node) bool {
	return node.Is("RegExpLiteral") || node.Is("Literal") && node.Get("regex") != nil
}

// isStringLiteral is upstream's isStringLiteral.
func isStringLiteral(node Node) bool {
	if node.Is("StringLiteral") {
		return true
	}
	_, isString := node.Get("value").(string)
	return node.Is("Literal") && isString
}

// getRaw is upstream's getRaw: node.extra?.raw ?? node.raw. typescript-estree never sets extra.raw.
func getRaw(node Node) string { return node.String("raw") }

// isMethod is upstream's isMethod.
func isMethod(node Node) bool {
	kind := node.String("kind")
	return node.Truthy("method") && kind == "init" || kind == "get" || kind == "set"
}

// isObjectProperty is upstream's isObjectProperty.
func isObjectProperty(node Node) bool {
	return node.Is("ObjectProperty") || node.Is("Property") && !isMethod(node)
}

// isMemberish is upstream's isMemberish.
func isMemberish(node Node) bool {
	return isMemberExpression(node) || node.Is("BindExpression") && node.Truthy("object")
}

// isNodeMatches is upstream's isNodeMatches: whether node is any of the dotted names.
func isNodeMatches(node Node, nameOrPaths []string) bool {
	for _, nameOrPath := range nameOrPaths {
		if isNodeMatchesNameOrPath(node, nameOrPath) {
			return true
		}
	}
	return false
}

// isNodeMatchesNameOrPath is upstream's isNodeMatchesNameOrPath. Upstream splits the path on "." and walks
// the names from the last; this reads the same names from the end of the string, so no list is made: one
// per call was 0.36M objects a cold run on ahra (#wcgw0n4). A name is first (upstream's index 0) when no dot
// is left before it, and second (index 1) when exactly one name is.
func isNodeMatchesNameOrPath(node Node, nameOrPath string) bool {
	rest := nameOrPath
	for {
		dot := strings.LastIndexByte(rest, '.')
		name := rest[dot+1:]
		if dot == -1 {
			return node.Is("Identifier") && node.String("name") == name
		}
		rest = rest[:dot]
		isSecond := strings.IndexByte(rest, '.') == -1
		if isSecond && node.Is("MetaProperty") && node.Child("property").Is("Identifier") &&
			node.Child("property").String("name") == name {
			node = node.Child("meta")
			continue
		}
		if node.Is("MemberExpression") && !node.Truthy("optional") && !node.Truthy("computed") &&
			node.Child("property").Is("Identifier") && node.Child("property").String("name") == name {
			node = node.Child("object")
			continue
		}
		return false
	}
}

// precedence is upstream's PRECEDENCE, utilities/get-precedence.js.
var precedence = func() map[string]int {
	table := map[string]int{}
	for index, operators := range [][]string{
		{"|>"}, {"??"}, {"||"}, {"&&"}, {"|"}, {"^"}, {"&"},
		{"==", "===", "!=", "!=="},
		{"<", ">", "<=", ">=", "in", "instanceof"},
		{">>", "<<", ">>>"},
		{"+", "-"},
		{"*", "/", "%"},
		{"**"},
	} {
		for _, operator := range operators {
			table[operator] = index
		}
	}
	return table
}()

// getPrecedence is upstream's getPrecedence. The second result is false where upstream returns
// undefined, which compares unequal to every number.
func getPrecedence(operator string) (int, bool) {
	value, known := precedence[operator]
	return value, known
}

var (
	equalityOperators       = map[string]bool{"==": true, "!=": true, "===": true, "!==": true}
	multiplicativeOperators = map[string]bool{"*": true, "/": true, "%": true}
	bitshiftOperators       = map[string]bool{">>": true, ">>>": true, "<<": true}
)

// shouldFlatten is upstream's shouldFlatten, utilities/should-flatten.js.
func shouldFlatten(parentOperator string, nodeOperator string) bool {
	nodePrecedence, nodeKnown := getPrecedence(nodeOperator)
	parentPrecedence, parentKnown := getPrecedence(parentOperator)
	if nodeKnown != parentKnown || nodePrecedence != parentPrecedence {
		return false
	}
	if parentOperator == "**" {
		return false
	}
	if equalityOperators[parentOperator] && equalityOperators[nodeOperator] {
		return false
	}
	if nodeOperator == "%" && multiplicativeOperators[parentOperator] ||
		parentOperator == "%" && multiplicativeOperators[nodeOperator] {
		return false
	}
	if nodeOperator != parentOperator && multiplicativeOperators[nodeOperator] && multiplicativeOperators[parentOperator] {
		return false
	}
	if bitshiftOperators[parentOperator] && bitshiftOperators[nodeOperator] {
		return false
	}
	return true
}

// isBitwiseOperator is upstream's isBitwiseOperator.
func isBitwiseOperator(operator string) bool {
	return bitshiftOperators[operator] || operator == "|" || operator == "^" || operator == "&"
}
