package javascript

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// print/binaryish.js.
//
// Our parser is always "typescript" and experimentalOperatorPosition is always "end", so the branches
// for NGPipeExpression (Angular), the Hack pipeline `|>` (Babel's __isUsingHackPipeline), Vue filter
// sequences (isVueFilterSequenceExpression, only the __vue_expression parsers) and the "start" operator
// position are not ported. With those gone, upstream's lineBeforeOperator is always false and
// rightSuffix is always "".

/*
- `BinaryExpression`
- `LogicalExpression`
- `NGPipeExpression`(Angular)
*/
// printBinaryishExpression is upstream's printBinaryishExpression.
func printBinaryishExpression(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	parent := parentOf(path)
	grandparent := grandparentOf(path)
	key := keyOf(path)
	isInsideParenthesis := key != "body" &&
		parent.Is("IfStatement", "WhileStatement", "SwitchStatement", "DoWhileStatement")

	parts := printBinaryishExpressions(path, options, print, false /* isNested */, isInsideParenthesis)

	//   if (
	//     this.hasPlugin("dynamicImports") && this.lookahead().type === tt.parenLeft
	//   ) {
	//
	// looks super weird, we want to break the children if the parent breaks
	//
	//   if (
	//     this.hasPlugin("dynamicImports") &&
	//     this.lookahead().type === tt.parenLeft
	//   ) {
	if isInsideParenthesis {
		return doc.Concat(parts)
	}

	// Upstream's isHackPipeline branch (`|>` under Babel's hack pipeline) is unreachable for TypeScript.

	// Break between the parens in
	// unaries or in a member or specific call expression, i.e.
	//
	//   (
	//     a &&
	//     b &&
	//     c
	//   ).call()
	if key == "callee" && isCallOrNewExpression(parent) ||
		// `UnaryExpression` adds parentheses and indention when argument has comment
		parent.Is("UnaryExpression") && !hasAnyComment(current) ||
		isMemberExpression(parent) && !parent.Bool("computed") {
		return groupIn(path, concatIn(path, indentIn(path, concatIn(path, softline, doc.Concat(parts))), softline))
	}

	// Avoid indenting sub-expressions in some cases where the first sub-expression is already
	// indented accordingly. We should indent sub-expressions where the first case isn't indented.
	// Upstream's NGRoot and NGMicrosyntaxExpression clauses are Angular-only and are not ported.
	shouldNotIndent := isReturnOrThrowStatement(parent) ||
		parent.Is("JSXExpressionContainer") && grandparent.Is("JSXAttribute") ||
		current.String("operator") != "|" && parent.Is("JsExpressionRoot") ||
		current == parent.Child("body") && parent.Is("ArrowFunctionExpression") ||
		current != parent.Child("body") && parent.Is("ForStatement") ||
		parent.Is("ConditionalExpression") &&
			!isReturnOrThrowStatement(grandparent) &&
			!isCallOrNewExpression(grandparent) ||
		parent.Is("TemplateLiteral") ||
		key == "argument" && parent.Is("UnaryExpression") ||
		key == "arguments" && isBooleanTypeCoercion(parent)

	shouldIndentIfInlining := parent.Is(
		"AssignmentExpression",
		"VariableDeclarator",
		"ClassProperty",
		"PropertyDefinition",
		"TSAbstractPropertyDefinition",
		"ClassPrivateProperty",
	) || isObjectProperty(parent)

	samePrecedenceSubExpression := isBinaryish(current.Child("left")) &&
		shouldFlatten(current.String("operator"), current.Child("left").String("operator"))

	if shouldNotIndent ||
		shouldInlineLogicalExpression(current) && !samePrecedenceSubExpression ||
		!shouldInlineLogicalExpression(current) && shouldIndentIfInlining {
		return groupIn(path, doc.Concat(parts))
	}

	if len(parts) == 0 {
		return emptyDoc
	}

	// If the right part is a JSX node, we include it in a separate group to
	// prevent it breaking the whole chain, so we can print the expression like:
	//
	//   foo && bar && (
	//     <Foo>
	//       <Bar />
	//     </Foo>
	//   )

	hasJsx := isJsxElement(current.Child("right"))

	firstGroupIndex := -1
	for index, part := range parts {
		if _, isGroup := part.(*doc.Group); isGroup {
			firstGroupIndex = index
			break
		}
	}

	// Separate the leftmost expression, possibly with its leading comments.
	headLength := 1
	if firstGroupIndex != -1 {
		headLength = firstGroupIndex + 1
	}
	headLength = min(headLength, len(parts))
	headParts := parts[:headLength]

	// parts.slice(headParts.length, hasJsx ? -1 : undefined): a negative end counts from the end, and
	// an end before the start is an empty slice.
	restEnd := len(parts)
	if hasJsx {
		restEnd = len(parts) - 1
	}
	var rest []Doc
	if restEnd > headLength {
		rest = parts[headLength:restEnd]
	}

	// Upstream names the Symbol "logicalChain-" + ++uid; the name is for debugging only, so the module
	// counter is not ported.
	groupID := newGroupID("logicalChain")

	chainParts := make([]Doc, 0, len(headParts)+1)
	// Don't include the initial expression in the indentation
	// level. The first item is guaranteed to be the first
	// left-most expression.
	chainParts = append(chainParts, headParts...)
	chainParts = append(chainParts, indentIn(path, doc.Concat(rest)))
	chain := groupWithIn(path, doc.Concat(chainParts), doc.GroupOptions{ID: groupID})

	if !hasJsx {
		return chain
	}

	jsxPart := parts[len(parts)-1]
	return groupIn(path, concatIn(path, chain, indentIfBreak(jsxPart, groupID, false)))
}

// For binary expressions to be consistent, we need to group
// subsequent operators with the same precedence level under a single
// group. Otherwise they will be nested such that some of them break
// onto new lines but not all. Operators with the same precedence
// level should either all break or not. Because we group them by
// precedence level and the AST is structured based on precedence
// level, things are naturally broken up correctly, i.e. `&&` is
// broken before `+`.
func printBinaryishExpressions(path *Path, options *Options, print PrintFunc, isNested bool, isInsideParenthesis bool) []Doc {
	current := node(path)

	// Simply print the node normally.
	if !isBinaryish(current) {
		return []Doc{groupIn(path, print(nil, nil))}
	}

	var parts []Doc

	// We treat BinaryExpression and LogicalExpression nodes the same.

	// Put all operators with the same precedence level in the same
	// group. The reason we only need to do this with the `left`
	// expression is because given an expression like `1 + 2 - 3`, it
	// is always parsed like `((1 + 2) - 3)`, meaning the `left` side
	// is where the rest of the expression will exist. Binary
	// expressions on the right side mean they have a difference
	// precedence level and should be treated as a separate group, so
	// print them normally. (This doesn't hold for the `**` operator,
	// which is unique in that it is right-associative.)
	if shouldFlatten(current.String("operator"), current.Child("left").String("operator")) {
		// Flatten them out by recursively calling this function.
		parts = call(path, func(path *Path) []Doc {
			return printBinaryishExpressions(path, options, print, true /* isNested */, isInsideParenthesis)
		}, "left")
	} else {
		parts = append(parts, groupIn(path, print("left", nil)))
	}

	shouldInline := shouldInlineLogicalExpression(current)
	rightNodeToCheckComments := current.Child("right")
	if rightNodeToCheckComments.Is("ChainExpression") {
		rightNodeToCheckComments = rightNodeToCheckComments.Child("expression")
	}
	// Upstream's lineBeforeOperator is true only for NGPipeExpression, `|>` and Vue filter sequences,
	// none of which the typescript parser produces.
	lineBeforeOperator := false
	// Upstream's hasTypeCastComment and commentBeforeOperator feed only the "start" operator position.

	operator := current.String("operator")
	// Upstream's rightSuffix is non-empty only for NGPipeExpression arguments.

	var right Doc
	if shouldInline {
		var rightContent Doc
		if hasLeadingOwnLineComment(originalText(options), rightNodeToCheckComments) {
			rightContent = indentIn(path, concatIn(path, line, print("right", nil), emptyDoc))
		} else {
			rightContent = concatIn(path, " ", print("right", nil), emptyDoc)
		}
		right = concatIn(path, operator, rightContent)
	} else {
		// Upstream's isHackPipeline recursion into "right" is unreachable for TypeScript.
		rightContent := print("right", nil)
		var beforeOperator Doc = emptyDoc
		var afterOperator Doc = line
		if lineBeforeOperator {
			beforeOperator = line
			afterOperator = doc.Text(" ")
		}
		right = concatIn(path, beforeOperator, operator, afterOperator, rightContent, emptyDoc)
	}

	// If there's only a single binary expression, we want to create a group
	// in order to avoid having a small right part like -1 be on its own line.
	parent := parentOf(path)
	shouldBreak := hasComment(current.Child("left"), commentTrailing|commentLine, nil)
	shouldGroup := shouldBreak ||
		!(isInsideParenthesis && current.Is("LogicalExpression")) &&
			parent.Type() != current.Type() &&
			current.Child("left").Type() != current.Type() &&
			current.Child("right").Type() != current.Type()
	if shouldGroup {
		right = groupWithIn(path, right, doc.GroupOptions{ShouldBreak: shouldBreak})
	}

	if lineBeforeOperator {
		parts = append(parts, emptyDoc, right)
	} else {
		parts = append(parts, doc.Text(" "), right)
	}

	// The root comments are already printed, but we need to manually print
	// the other ones since we don't call the normal print on BinaryExpression,
	// only for the left and right parts
	if isNested && hasAnyComment(current) {
		printed := doc.CleanDoc(printing.PrintComments(path, doc.Concat(parts), options, nil))
		/* c8 ignore next 3 */
		if fillDoc, isFill := printed.(*doc.Fill); isFill {
			return fillDoc.Parts
		}

		if concatenated, isArray := doc.Parts(printed); isArray {
			return concatenated
		}
		return []Doc{printed}
	}

	return parts
}

// shouldInlineLogicalExpression is upstream's shouldInlineLogicalExpression.
func shouldInlineLogicalExpression(node Node) bool {
	if !node.Is("LogicalExpression") {
		return false
	}

	if isObjectExpression(node.Child("right")) && len(node.Child("right").List("properties")) > 0 {
		return true
	}

	if isArrayExpression(node.Child("right")) && len(node.Child("right").List("elements")) > 0 {
		return true
	}

	if isJsxElement(node.Child("right")) {
		return true
	}

	return false
}
