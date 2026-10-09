package javascript

import (
	"fmt"
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// print/jsx.js, second half: the wrapping parentheses, attributes, expression containers, opening and
// closing elements and fragments, and the printJsx dispatch. printJsxElementInternal and
// printJsxChildren are in print_jsx_children.go.

// isNoWrapParent is upstream's isNoWrapParent, a createTypeCheckFunction.
func isNoWrapParent(node Node) bool {
	return node.Is(
		"ArrayExpression",
		"JSXAttribute",
		"JSXElement",
		"JSXExpressionContainer",
		"JSXFragment",
		"ExpressionStatement",
		"NewExpression",
		"CallExpression",
		"OptionalCallExpression",
		"ConditionalExpression",
		"JsExpressionRoot",
		"MatchExpressionCase",
	)
}

// maybeWrapJsxElementInParens is upstream's maybeWrapJsxElementInParens.
func maybeWrapJsxElementInParens(path *Path, elem Doc, options *Options) Doc {
	parent := parentOf(path)

	if isNoWrapParent(parent) {
		return elem
	}

	shouldBreak := shouldBreakJsxElement(path)
	needsParens := needsParentheses(path, options)

	var openParen, closeParen Doc = emptyDoc, emptyDoc
	if !needsParens {
		openParen, closeParen = ifBreak("(", nil), ifBreak(")", nil)
	}
	return groupWithIn(path, concatIn(path, openParen,
		indentIn(path, concatIn(path, softline, elem)),
		softline,
		closeParen,
	),
		doc.GroupOptions{ShouldBreak: shouldBreak},
	)
}

// shouldBreakJsxElement is upstream's shouldBreakJsxElement.
func shouldBreakJsxElement(path *Path) bool {
	isCallArgument := func(value any, key any, _ int, _ bool) bool {
		node, _ := value.(Node)
		return key == "arguments" && isCallExpression(node)
	}
	return path.Match(
		nil,
		keyIsType("body", "ArrowFunctionExpression"),
		isCallArgument,
	) &&
		// Babel
		(path.Match(
			nil,
			nil,
			nil,
			keyIsType("expression", "JSXExpressionContainer"),
		) ||
			// Estree
			path.Match(
				nil,
				nil,
				nil,
				keyIsType("expression", "ChainExpression"),
				keyIsType("expression", "JSXExpressionContainer"),
			))
}

// printJsxAttribute is upstream's printJsxAttribute. jsxSingleQuote is never set in our
// configurations, so the preferred quote is always the double quote.
func printJsxAttribute(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	parts := []any{print("name", nil)}

	if current.Child("value") != nil {
		var res Doc
		if isStringLiteral(current.Child("value")) {
			raw := getRaw(current.Child("value"))
			// Remove enclosing quotes and unescape
			// all quotes so we get an accurate preferred quote
			final := raw[1 : len(raw)-1]
			final = strings.ReplaceAll(final, "&apos;", "'")
			final = strings.ReplaceAll(final, "&quot;", `"`)
			quote := getPreferredQuote(final, false)
			if quote == `"` {
				final = strings.ReplaceAll(final, `"`, "&quot;")
			} else {
				final = strings.ReplaceAll(final, "'", "&apos;")
			}
			res = call(path, func(path *Path) Doc {
				return printing.PrintComments(path, doc.ReplaceEndOfLine(doc.Text(quote+final+quote), nil), options, nil)
			}, "value")
		} else {
			res = print("value", nil)
		}
		parts = append(parts, "=", res)
	}

	return concatIn(path, parts...)
}

// printJsxExpressionContainer is upstream's printJsxExpressionContainer.
func printJsxExpressionContainer(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)

	var shouldInline func(node Node, parent Node) bool
	shouldInline = func(node Node, parent Node) bool {
		return node.Is("JSXEmptyExpression") ||
			(!hasAnyComment(node) &&
				(isArrayExpression(node) ||
					isObjectExpression(node) ||
					node.Is("ArrowFunctionExpression") ||
					(node.Is("AwaitExpression") &&
						(shouldInline(node.Child("argument"), node) ||
							node.Child("argument").Is("JSXElement"))) ||
					isCallExpression(stripChainElementWrappers(node)) ||
					node.Is("FunctionExpression") ||
					node.Is("TemplateLiteral") ||
					node.Is("TaggedTemplateExpression") ||
					node.Is("DoExpression") ||
					(isJsxElement(parent) &&
						(node.Is("ConditionalExpression") || isBinaryish(node)))))
	}

	if shouldInline(current.Child("expression"), parentOf(path)) {
		return groupIn(path, concatIn(path, "{", print("expression", nil), lineSuffixBoundary, "}"))
	}

	return groupIn(path, concatIn(path, "{",
		indentIn(path, concatIn(path, softline, print("expression", nil))),
		softline,
		lineSuffixBoundary,
		"}",
	))
}

// printJsxOpeningElement is upstream's printJsxOpeningElement. singleAttributePerLine is never set in
// our configurations, so attributeLine is always line.
func printJsxOpeningElement(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	attributes := current.List("attributes")

	nameHasComments := hasAnyComment(current.Child("name")) || hasAnyComment(current.Child("typeArguments"))

	// Don't break self-closing elements with no attributes and no comments
	if current.Bool("selfClosing") && len(attributes) == 0 && !nameHasComments {
		return concatIn(path, "<", print("name", nil), print("typeArguments", nil), " />")
	}

	// don't break up opening elements with a single long text attribute
	if len(attributes) == 1 &&
		isStringLiteral(attributes[0].Child("value")) &&
		!strings.Contains(attributes[0].Child("value").String("value"), "\n") &&
		// We should break for the following cases:
		// <div
		//   // comment
		//   attr="value"
		// >
		// <div
		//   attr="value"
		//   // comment
		// >
		!nameHasComments &&
		!hasAnyComment(attributes[0]) {
		parts := []any{
			"<",
			print("name", nil),
			print("typeArguments", nil),
			" ",
		}
		for _, printed := range printAll(path, print, "attributes") {
			parts = append(parts, printed)
		}
		if current.Bool("selfClosing") {
			parts = append(parts, " />")
		} else {
			parts = append(parts, ">")
		}
		return groupIn(path, concatIn(path, parts...))
	}

	// We should print the opening element expanded if any prop value is a
	// string literal with newlines
	shouldBreak := false
	for _, attribute := range attributes {
		if isStringLiteral(attribute.Child("value")) && strings.Contains(attribute.Child("value").String("value"), "\n") {
			shouldBreak = true
			break
		}
	}

	attributeLine := line

	parts := []any{
		"<",
		print("name", nil),
		print("typeArguments", nil),
		indentIn(path, mapPath(path, func(path *Path, _ int) Doc {
			var separator Doc
			if path.IsFirst() {
				separator = attributeLine
			} else if isNextLineEmptyAfter(previousOf(path), options) {
				separator = concatIn(path, hardline, hardline)
			} else {
				separator = attributeLine
			}
			return concatIn(path, separator, print(nil, nil))
		}, "attributes"),
		),
	}
	parts = append(parts, printEndOfOpeningTag(current, options, nameHasComments)...)
	return groupWithIn(path, concatIn(path, parts...), doc.GroupOptions{ShouldBreak: shouldBreak})
}

// printEndOfOpeningTag is upstream's printEndOfOpeningTag; it returns the array upstream spreads.
func printEndOfOpeningTag(node Node, options *Options, nameHasComments bool) []any {
	if node.Bool("selfClosing") {
		return []any{line, "/>"}
	}
	bracketSameLine := shouldPrintBracketSameLine(
		node,
		options,
		nameHasComments,
	)
	if bracketSameLine {
		return []any{">"}
	}
	return []any{softline, ">"}
}

// shouldPrintBracketSameLine is upstream's shouldPrintBracketSameLine. The deprecated
// jsxBracketSameLine is not an option our configurations can set, so only bracketSameLine is read.
func shouldPrintBracketSameLine(node Node, options *Options, nameHasComments bool) bool {
	attributes := node.List("attributes")
	lastAttrHasTrailingComments := len(attributes) > 0 &&
		hasComment(attributes[len(attributes)-1], commentTrailing, nil)
	// Simple tags (no attributes and no comment in tag name) should be
	// kept unbroken regardless of `bracketSameLine`.
	// jsxBracketSameLine is deprecated in favour of bracketSameLine,
	// but is still needed for backwards compatibility.
	return (len(attributes) == 0 && !nameHasComments) ||
		(settingsOf(options).BracketSameLine &&
			// We should print the bracket in a new line for the following cases:
			// <div
			//   // comment
			// >
			// <div
			//   attr // comment
			// >
			(!nameHasComments || len(attributes) > 0) &&
			!lastAttrHasTrailingComments)
}

// printJsxClosingElement is upstream's printJsxClosingElement.
func printJsxClosingElement(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	parts := []any{"</"}

	printed := print("name", nil)
	if hasComment(current.Child("name"), commentLeading|commentLine, nil) {
		parts = append(parts, indentIn(path, concatIn(path, hardline, printed)), hardline)
	} else if hasComment(current.Child("name"), commentLeading|commentBlock, nil) {
		parts = append(parts, " ", printed)
	} else {
		parts = append(parts, printed)
	}

	parts = append(parts, ">")

	return concatIn(path, parts...)
}

// printJsxOpeningClosingFragment is upstream's printJsxOpeningClosingFragment.
func printJsxOpeningClosingFragment(path *Path, options *Options) Doc {
	current := node(path)
	nodeHasComment := hasAnyComment(current)
	hasOwnLineComment := hasComment(current, commentLine, nil)
	isOpeningFragment := current.Is("JSXOpeningFragment")

	open := "</"
	if isOpeningFragment {
		open = "<"
	}
	var leading Doc = emptyDoc
	if hasOwnLineComment {
		leading = hardline
	} else if nodeHasComment && !isOpeningFragment {
		leading = doc.Text(" ")
	}
	var trailing Doc = emptyDoc
	if hasOwnLineComment {
		trailing = hardline
	}
	return concatIn(path, open,
		indentIn(path, concatIn(path, leading,
			printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{}),
		)),
		trailing,
		">",
	)
}

// printJsxElement is upstream's printJsxElement.
func printJsxElement(path *Path, options *Options, print PrintFunc) Doc {
	elem := printing.PrintComments(
		path,
		printJsxElementInternal(path, options, print),
		options,
		nil,
	)
	return maybeWrapJsxElementInParens(path, elem, options)
}

// printJsxEmptyExpression is upstream's printJsxEmptyExpression.
func printJsxEmptyExpression(path *Path, options *Options) Doc {
	current := node(path)
	requiresHardline := hasComment(current, commentLine, nil)

	var trailing Doc = emptyDoc
	if requiresHardline {
		trailing = hardline
	}
	return concatIn(path, printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{Indent: requiresHardline}),
		trailing,
	)
}

// printJsxSpreadAttributeOrChild is upstream's printJsxSpreadAttributeOrChild:
// `JSXSpreadAttribute` and `JSXSpreadChild`.
func printJsxSpreadAttributeOrChild(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	name := "expression"
	if current.Is("JSXSpreadAttribute") {
		name = "argument"
	}
	return concatIn(path, "{",
		call(path, func(path *Path) Doc {
			printed := concatIn(path, "...", print(nil, nil))
			if !hasAnyComment(node(path)) {
				return printed
			}
			return concatIn(path, indentIn(path, concatIn(path, softline, printing.PrintComments(path, printed, options, nil))),
				softline,
			)
		}, name),
		"}",
	)
}

// printJsx is upstream's printJsx. It returns nil for a node that is not JSX, where upstream returns
// undefined, so the dispatch moves on to the TypeScript and ESTree printers.
func printJsx(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)

	// JSX nodes always starts with `JSX`
	if !strings.HasPrefix(current.Type(), "JSX") {
		return nil
	}

	switch current.Type() {
	case "JSXAttribute":
		return printJsxAttribute(path, options, print)
	case "JSXIdentifier":
		return doc.Text(current.String("name"))
	case "JSXNamespacedName":
		return join(":", []Doc{print("namespace", nil), print("name", nil)})
	case "JSXMemberExpression":
		return join(".", []Doc{print("object", nil), print("property", nil)})
	case "JSXSpreadAttribute", "JSXSpreadChild":
		return printJsxSpreadAttributeOrChild(path, options, print)
	case "JSXExpressionContainer":
		return printJsxExpressionContainer(path, options, print)
	case "JSXFragment", "JSXElement":
		return printJsxElement(path, options, print)
	case "JSXOpeningElement":
		return printJsxOpeningElement(path, options, print)
	case "JSXClosingElement":
		return printJsxClosingElement(path, options, print)
	case "JSXOpeningFragment", "JSXClosingFragment":
		return printJsxOpeningClosingFragment(path, options /* , print*/)
	case "JSXEmptyExpression":
		return printJsxEmptyExpression(path, options /* , print*/)
	case "JSXText":
		/* c8 ignore next */
		panic("JSXText should be handled by JSXElement")
	default:
		/* c8 ignore next */
		panic(fmt.Sprintf("unexpected node type %q in JSX", current.Type()))
	}
}

// isEmptyJsxElement is upstream's isEmptyJsxElement.
func isEmptyJsxElement(node Node) bool {
	children := node.List("children")
	if len(children) == 0 {
		return true
	}
	if len(children) > 1 {
		return false
	}

	// if there is one text child and does not contain any meaningful text
	// we can treat the element as empty.
	child := children[0]
	return child.Is("JSXText") && !isMeaningfulJsxText(child)
}

// isJsxWhitespaceExpression is upstream's isJsxWhitespaceExpression:
// detect an expression node representing `{" "}`.
func isJsxWhitespaceExpression(node Node) bool {
	return node.Is("JSXExpressionContainer") &&
		isStringLiteral(node.Child("expression")) &&
		node.Child("expression").String("value") == " " &&
		!hasAnyComment(node.Child("expression"))
}
