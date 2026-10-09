package css

// src/language-css/printer-postcss.js: genericPrint, for the postcss nodes and the three sub-parsers'
// nodes the glue grafts into them (postcss-media-query-parser, postcss-selector-parser,
// postcss-values-parser). The printer object is postcssPrinter in format.go.
//
// The parser is "css" or "scss" (parserOf, format.go): the Less branches (mixins, functions, variables,
// extend) are left out, each with a comment where upstream has it. options.__isHTMLStyleAttribute is an
// embed-only option and false here.

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

var (
	declarationBeforePattern       = regexp.MustCompile(`[` + javaScriptWhitespaceCharacters + `;]`)
	importantPattern               = regexp.MustCompile(`(?i)` + javaScriptWhitespaceClass + `*!` + javaScriptWhitespaceClass + `*important`)
	scssDefaultPattern             = regexp.MustCompile(`(?i)` + javaScriptWhitespaceClass + `*!default`)
	scssGlobalPattern              = regexp.MustCompile(`(?i)` + javaScriptWhitespaceClass + `*!global`)
	afterNameTwoNewlinesPattern    = regexp.MustCompile(`^` + javaScriptWhitespaceClass + `*\n` + javaScriptWhitespaceClass + `*\n`)
	afterNameNewlinePattern        = regexp.MustCompile(`^` + javaScriptWhitespaceClass + `*\n`)
	mediaFeatureSpacesPattern      = regexp.MustCompile(` +`)
	mediaURLOpeningPattern         = regexp.MustCompile(`(?i)^url\(` + javaScriptWhitespaceClass + `+`)
	mediaURLClosingPattern         = regexp.MustCompile(javaScriptWhitespaceClass + `+\)$`)
	valueAllSpacePattern           = regexp.MustCompile(`^ *$`)
	javaScriptWhitespaceCharacters = javaScriptWhitespaceClass[1 : len(javaScriptWhitespaceClass)-1]
)

func genericPrint(path *astPath, options *printerOptions, print printing.PrintFunc, _ any) doc.Doc {
	node := currentNode(path)

	// The core's front matter support prints front matter that is not embedded as written
	// (src/main/front-matter/print.js), ahead of the printer.
	if isFrontMatter(node) {
		return doc.Text(node.String("raw"))
	}

	switch node.Type() {
	case "css-root":
		nodes := printSequence(path, options, print)
		after := trim(rawString(node, "after"))
		if strings.HasPrefix(after, ";") {
			after = trim(after[1:])
		}

		var frontMatter doc.Doc = doc.Text("")
		if node.Child("frontMatter") != nil {
			frontMatter = doc.Concat{
				print("frontMatter", nil),
				doc.Hardline,
				choose(len(node.List("nodes")) > 0, func() doc.Doc { return doc.Hardline }),
			}
		}
		return doc.Concat{
			frontMatter,
			nodes,
			choose(after != "", func() doc.Doc { return doc.Text(" " + after) }),
			choose(len(node.List("nodes")) > 0, func() doc.Doc { return doc.Hardline }),
		}

	case "css-comment":
		isInlineComment := node.Truthy("inline") || estree.IsTruthy(rawsOf(node)["inline"])

		text := sliceText(options.OriginalText, locStart(node), locEnd(node))

		if isInlineComment {
			return doc.Text(trimEnd(text))
		}
		return doc.Text(text)

	case "css-rule":
		selector := node.Child("selector")
		var block doc.Doc = doc.Text(";")
		if isPresent(node.Get("nodes")) {
			var opening doc.Doc = doc.Text("")
			if selector.Type() == "selector-unknown" && lastLineHasInlineComment(selector.String("value")) {
				opening = doc.LineDoc
			} else if selector != nil {
				opening = doc.Text(" ")
			}
			block = doc.Concat{
				opening,
				doc.Text("{"),
				choose(len(node.List("nodes")) > 0, func() doc.Doc {
					return doc.NewIndent(doc.Concat{doc.Hardline, printSequence(path, options, print)})
				}),
				doc.Hardline,
				doc.Text("}"),
				choose(isDetachedRulesetDeclarationNode(node), func() doc.Doc { return doc.Text(";") }),
			}
		}
		return doc.Concat{
			print("selector", nil),
			choose(node.Truthy("important"), func() doc.Doc { return doc.Text(" !important") }),
			block,
		}

	case "css-decl":
		return printDeclaration(path, options, print)

	case "css-atrule":
		return printAtRule(path, options, print)

	// postcss-media-query-parser
	case "media-query-list":
		parts := []doc.Doc{}
		path.Each(func(path *astPath, _ int, _ any) {
			node := currentNode(path)
			if node.Type() == "media-query" && node.String("value") == "" {
				return
			}
			parts = append(parts, print(nil, nil))
		}, "nodes")

		return doc.NewGroup(doc.NewIndent(doc.Join(doc.LineDoc, parts)), doc.GroupOptions{})

	case "media-query":
		return doc.Concat{
			doc.Join(doc.Text(" "), printAll(path, print, "nodes")),
			choose(!path.IsLast(), func() doc.Doc { return doc.Text(",") }),
		}

	case "media-type":
		return doc.Text(adjustNumbers(adjustStrings(node.String("value"), options)))

	case "media-feature-expression":
		if !isPresent(node.Get("nodes")) {
			return doc.Text(node.String("value"))
		}
		return doc.Concat(append(append([]doc.Doc{doc.Text("(")}, printAll(path, print, "nodes")...), doc.Text(")")))

	case "media-feature":
		return doc.Text(maybeToLowerCase(
			adjustStrings(mediaFeatureSpacesPattern.ReplaceAllString(node.String("value"), " "), options),
		))

	case "media-colon":
		return doc.Concat{doc.Text(node.String("value")), doc.Text(" ")}

	case "media-value":
		return doc.Text(adjustNumbers(adjustStrings(node.String("value"), options)))

	case "media-keyword":
		return doc.Text(adjustStrings(node.String("value"), options))

	case "media-url":
		value := mediaURLOpeningPattern.ReplaceAllString(node.String("value"), "url(")
		value = mediaURLClosingPattern.ReplaceAllString(value, ")")
		return doc.Text(adjustStrings(value, options))

	case "media-unknown":
		return doc.Text(node.String("value"))

	// postcss-selector-parser
	case "selector-root":
		var customSelector doc.Doc = doc.Text("")
		if insideAtRuleNode(path, "custom-selector") {
			atRule, _ := path.FindAncestor(func(node *estree.Node) bool { return node.Type() == "css-atrule" })
			customSelector = doc.Concat{doc.Text(atRule.String("customSelector")), doc.LineDoc}
		}
		var separator doc.Doc = doc.Hardline
		if insideAtRuleNode(path, "extend", "custom-selector", "nest") {
			separator = doc.LineDoc
		}
		return doc.NewGroup(doc.Concat{
			customSelector,
			doc.Join(doc.Concat{doc.Text(","), separator}, printAll(path, print, "nodes")),
		}, doc.GroupOptions{})

	case "selector-selector":
		shouldIndent := len(node.List("nodes")) > 2
		var contents doc.Doc = doc.Concat(printAll(path, print, "nodes"))
		if shouldIndent {
			contents = doc.NewIndent(contents)
		}
		return doc.NewGroup(contents, doc.GroupOptions{})

	case "selector-comment":
		return doc.Text(node.String("value"))

	case "selector-string":
		return doc.Text(adjustStrings(node.String("value"), options))

	case "selector-tag":
		value := node.String("value")
		previous, _ := path.Previous()
		if previous.Type() != "selector-nesting" {
			if isKeyframeAtRuleKeywords(path, value) {
				value = toLowerCase(value)
			}
			value = adjustNumbers(value)
		}
		return doc.Concat{printNamespace(node), doc.Text(value)}

	case "selector-id":
		return doc.Concat{doc.Text("#"), doc.Text(node.String("value"))}

	case "selector-class":
		return doc.Concat{doc.Text("."), doc.Text(adjustNumbers(adjustStrings(node.String("value"), options)))}

	case "selector-attribute":
		var value doc.Doc = doc.Text("")
		if node.String("value") != "" {
			value = doc.ReplaceEndOfLine(
				doc.Text(quoteAttributeValue(adjustStrings(trim(node.String("value")), options), options)),
				doc.LiterallineWithoutBreakParent,
			)
		}
		return doc.Concat{
			doc.Text("["),
			printNamespace(node),
			doc.Text(trim(node.String("attribute"))),
			doc.Text(node.String("operator")),
			value,
			choose(node.Truthy("insensitive"), func() doc.Doc { return doc.Text(" i") }),
			doc.Text("]"),
		}

	case "selector-combinator":
		value := node.String("value")
		if value == "+" || value == ">" || value == "~" || value == ">>>" {
			parentNode, _ := path.Parent()
			var leading doc.Doc = doc.LineDoc
			if parentNode.Type() == "selector-selector" && firstNode(parentNode.List("nodes")) == node {
				leading = doc.Text("")
			}

			return doc.Concat{leading, doc.Text(value), choose(!path.IsLast(), func() doc.Doc { return doc.Text(" ") })}
		}

		var leading doc.Doc = doc.Text("")
		if strings.HasPrefix(trimStart(value), "(") {
			leading = doc.LineDoc
		}
		var printed doc.Doc = doc.LineDoc
		if adjusted := adjustNumbers(adjustStrings(trim(value), options)); adjusted != "" {
			printed = doc.Text(adjusted)
		}

		return doc.Concat{leading, printed}

	case "selector-universal":
		return doc.Concat{printNamespace(node), doc.Text(node.String("value"))}

	case "selector-pseudo":
		return doc.Concat{
			doc.Text(maybeToLowerCase(node.String("value"))),
			choose(len(node.List("nodes")) > 0, func() doc.Doc {
				return doc.NewGroup(doc.Concat{
					doc.Text("("),
					doc.NewIndent(doc.Concat{doc.Softline, doc.Join(doc.Concat{doc.Text(","), doc.LineDoc}, printAll(path, print, "nodes"))}),
					doc.Softline,
					doc.Text(")"),
				}, doc.GroupOptions{})
			}),
		}

	case "selector-nesting":
		return doc.Text(node.String("value"))

	case "selector-unknown":
		return printSelectorUnknown(path, options)

	// postcss-values-parser
	case "value-value", "value-root":
		return print("group", nil)

	case "value-comment":
		text := sliceText(options.OriginalText, locStart(node), locEnd(node))
		if node.Truthy("inline") {
			return doc.NewLineSuffix(doc.Text(trimEnd(text)))
		}
		return doc.Text(text)

	case "value-comma_group":
		return printCommaSeparatedValueGroup(path, options, print)

	case "value-paren_group":
		return printParenthesizedValueGroup(path, options, print)

	case "value-func":
		return doc.Concat{
			doc.Text(node.String("value")),
			choose(insideAtRuleNode(path, "supports") && isMediaAndSupportsKeywords(node), func() doc.Doc { return doc.Text(" ") }),
			print("group", nil),
		}

	case "value-paren":
		return doc.Text(node.String("value"))

	case "value-number":
		return doc.Concat{doc.Text(printCssNumber(node.String("value"))), doc.Text(printUnit(node.String("unit")))}

	case "value-operator":
		return doc.Text(node.String("value"))

	case "value-word":
		if (node.Truthy("isColor") && node.Truthy("isHex")) || isWideKeywords(node.String("value")) {
			return doc.Text(toLowerCase(node.String("value")))
		}

		return doc.Text(node.String("value"))

	case "value-colon":
		previous, _ := path.Previous()
		previousValue, previousValueIsString := previous.Get("value").(string)
		var after doc.Doc = doc.LineDoc
		// Don't add spaces on escaped colon `:`, e.g: grid-template-rows: [row-1-00\:00] auto;
		if (previousValueIsString && strings.HasSuffix(previousValue, `\`)) ||
			// Don't add spaces on `:` in `url` function (i.e. `url(fbglyph: cross-outline, fig-white)`)
			insideValueFunctionNode(path, "url") {
			after = doc.Text("")
		}
		return doc.NewGroup(doc.Concat{doc.Text(node.String("value")), after}, doc.GroupOptions{})

	case "value-string":
		quote := rawString(node, "quote")
		return doc.Text(printString(quote+node.String("value")+quote, options))

	case "value-atword":
		return doc.Concat{doc.Text("@"), doc.Text(node.String("value"))}

	case "value-unicode-range":
		return doc.Text(node.String("value"))

	case "value-unknown":
		return doc.Text(node.String("value"))
	}

	// "front-matter" is handled in core, "value-comma" in `value-comma_group`.
	panic(fmt.Sprintf("Unexpected PostCSS node type: %q.", node.Type()))
}

func printDeclaration(path *astPath, options *printerOptions, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	parentNode, _ := path.Parent()

	rawBetween := rawString(node, "between")
	trimmedBetween := trim(rawBetween)
	isColon := trimmedBetween == ":"
	hasSpaceAfterColon := strings.HasSuffix(rawBetween, " ") && isColon
	stringValue, valueIsString := node.Get("value").(string)
	isValueAllSpace := valueIsString && valueAllSpacePattern.MatchString(stringValue)
	var value doc.Doc
	if valueIsString {
		value = doc.Text(stringValue)
	} else {
		value = print("value", nil)
	}

	if hasComposesNode(node) {
		value = doc.RemoveLines(value)
	}

	if !isColon &&
		lastLineHasInlineComment(trimmedBetween) &&
		!printing.Call(path, func(path *astPath) bool { return shouldBreakList(path) }, "value", "group", "group") {
		value = doc.NewIndent(doc.Concat{doc.Hardline, doc.Dedent(value)})
	}

	prop := node.String("prop")
	if !((parentNode.Type() == "css-atrule" && parentNode.Truthy("variable")) || insideIcssRuleNode(path)) {
		prop = maybeToLowerCase(prop)
	}

	var space doc.Doc = doc.Text(" ")
	if node.Truthy("extend") ||
		isValueAllSpace ||
		(!hasSpaceAfterColon &&
			node.Truthy("isNested") &&
			// node.value.group.group reads without optional chaining, a TypeError upstream where a
			// link is missing (mustNode); only .groups?.[0] is guarded.
			(isAtWordPlaceholderNode(mustNode(node.Child("value").Child("group")).Child("group")) ||
				isAtWordPlaceholderNode(firstNode(mustNode(node.Child("value").Child("group").Child("group")).List("groups"))))) {
		space = doc.Text("")
	}

	var important doc.Doc = doc.Text("")
	if rawImportant := rawString(node, "important"); rawImportant != "" {
		important = doc.Text(replaceFirst(importantPattern, rawImportant, " !important"))
	} else if node.Truthy("important") {
		important = doc.Text(" !important")
	}

	var scssDefault doc.Doc = doc.Text("")
	if rawScssDefault := rawString(node, "scssDefault"); rawScssDefault != "" {
		scssDefault = doc.Text(replaceFirst(scssDefaultPattern, rawScssDefault, " !default"))
	} else if node.Truthy("scssDefault") {
		scssDefault = doc.Text(" !default")
	}

	var scssGlobal doc.Doc = doc.Text("")
	if rawScssGlobal := rawString(node, "scssGlobal"); rawScssGlobal != "" {
		scssGlobal = doc.Text(replaceFirst(scssGlobalPattern, rawScssGlobal, " !global"))
	} else if node.Truthy("scssGlobal") {
		scssGlobal = doc.Text(" !global")
	}

	var ending doc.Doc
	switch {
	case isPresent(node.Get("nodes")):
		ending = doc.Concat{
			doc.Text(" {"),
			choose(len(node.List("nodes")) > 0, func() doc.Doc {
				return doc.NewIndent(doc.Concat{doc.Softline, printSequence(path, options, print)})
			}),
			doc.Softline,
			doc.Text("}"),
		}
	case isTemplatePropNode(node) &&
		!estree.IsTruthy(rawsOf(parentNode)["semicolon"]) &&
		characterBefore(options.OriginalText, locEnd(node)) != ";":
		ending = doc.Text("")
	default:
		ending = doc.Text(";")
	}

	return doc.Concat{
		doc.Text(declarationBeforePattern.ReplaceAllString(rawString(node, "before"), "")),
		doc.Text(prop),
		choose(strings.HasPrefix(trimmedBetween, "//"), func() doc.Doc { return doc.Text(" ") }),
		doc.Text(trimmedBetween),
		space,
		// The Less `extend(...)` selector is gated on options.parser === "less".
		value,
		important,
		scssDefault,
		scssGlobal,
		ending,
	}
}

func printAtRule(path *astPath, options *printerOptions, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	parentNode, _ := path.Parent()
	isTemplatePlaceholderNodeWithoutSemiColon := isTemplatePlaceholderNode(node) &&
		!estree.IsTruthy(rawsOf(parentNode)["semicolon"]) &&
		characterBefore(options.OriginalText, locEnd(node)) != ";"

	// Less mixins, functions and variables are gated on options.parser === "less".

	name := node.String("name")
	params := node.Get("params")
	stringParams, paramsIsString := params.(string)
	isImportUnknownValueEndsWithSemiColon := name == "import" &&
		node.Child("params").Type() == "value-unknown" &&
		strings.HasSuffix(node.Child("params").String("value"), ";")

	// If a Less file ends up being parsed with the SCSS parser, Less
	// variable declarations will be parsed as at-rules with names ending
	// with a colon, so keep the original case then.
	printedName := name
	if !(isDetachedRulesetCallNode(node) || strings.HasSuffix(name, ":") || isTemplatePlaceholderNode(node)) {
		printedName = maybeToLowerCase(name)
	}

	var printedParams doc.Doc = doc.Text("")
	if estree.IsTruthy(params) {
		var leading doc.Doc = doc.Text(" ")
		if isDetachedRulesetCallNode(node) {
			leading = doc.Text("")
		} else if isTemplatePlaceholderNode(node) {
			afterName := rawString(node, "afterName")
			switch {
			case afterName == "":
				leading = doc.Text("")
			case strings.HasSuffix(name, ":"):
				leading = doc.Text(" ")
			case afterNameTwoNewlinesPattern.MatchString(afterName):
				leading = doc.Concat{doc.Hardline, doc.Hardline}
			case afterNameNewlinePattern.MatchString(afterName):
				leading = doc.Hardline
			}
		}
		var paramsDoc doc.Doc
		if paramsIsString {
			paramsDoc = doc.Text(stringParams)
		} else {
			paramsDoc = print("params", nil)
		}
		printedParams = doc.Concat{leading, paramsDoc}
	}

	selector := node.Child("selector")
	var printedSelector doc.Doc = doc.Text("")
	if selector != nil {
		printedSelector = doc.NewIndent(doc.Concat{doc.Text(" "), print("selector", nil)})
	}

	var printedValue doc.Doc = doc.Text("")
	if estree.IsTruthy(node.Get("value")) {
		var after doc.Doc = doc.Text("")
		if isSCSSControlDirectiveNode(node, options) {
			if hasParensAroundNode(node) {
				after = doc.Text(" ")
			} else {
				after = doc.LineDoc
			}
		}
		printedValue = doc.NewGroup(doc.Concat{doc.Text(" "), print("value", nil), after}, doc.GroupOptions{})
	} else if name == "else" {
		printedValue = doc.Text(" ")
	}

	var ending doc.Doc
	if isPresent(node.Get("nodes")) {
		var opening doc.Doc = doc.Text(" ")
		_, selectorValueIsString := selector.Get("value").(string)
		if isSCSSControlDirectiveNode(node, options) {
			opening = doc.Text("")
		} else if (selector != nil &&
			!isPresent(selector.Get("nodes")) &&
			selectorValueIsString &&
			lastLineHasInlineComment(selector.String("value"))) ||
			(selector == nil &&
				paramsIsString &&
				lastLineHasInlineComment(stringParams)) {
			opening = doc.LineDoc
		}
		ending = doc.Concat{
			opening,
			doc.Text("{"),
			choose(len(node.List("nodes")) > 0, func() doc.Doc {
				return doc.NewIndent(doc.Concat{doc.Softline, printSequence(path, options, print)})
			}),
			doc.Softline,
			doc.Text("}"),
		}
	} else if isTemplatePlaceholderNodeWithoutSemiColon || isImportUnknownValueEndsWithSemiColon {
		ending = doc.Text("")
	} else {
		ending = doc.Text(";")
	}

	return doc.Concat{
		doc.Text("@"),
		doc.Text(printedName),
		printedParams,
		printedSelector,
		printedValue,
		ending,
	}
}

func printSelectorUnknown(path *astPath, options *printerOptions) doc.Doc {
	node := currentNode(path)
	ruleAncestorNode, _ := path.FindAncestor(func(node *estree.Node) bool { return node.Type() == "css-rule" })

	// Nested SCSS property
	if ruleAncestorNode.Truthy("isScssNestedProperty") {
		return doc.Text(adjustNumbers(adjustStrings(maybeToLowerCase(node.String("value")), options)))
	}

	// originalText has to be used for Less, see replaceQuotesInInlineComments in loc.js
	parentNode, _ := path.Parent()
	if selector := rawString(parentNode, "selector"); selector != "" {
		start := locStart(parentNode)
		end := start + len(selector)
		return doc.Text(trim(sliceText(options.OriginalText, start, end)))
	}

	// Same reason above
	grandParent, _ := path.Grandparent()
	if parentNode.Type() == "value-paren_group" &&
		grandParent.Type() == "value-func" &&
		grandParent.String("value") == "selector" {
		start := locEnd(parentNode.Child("open")) + 1
		end := locStart(parentNode.Child("close"))
		selector := trim(sliceText(options.OriginalText, start, end))

		if lastLineHasInlineComment(selector) {
			return doc.Concat{doc.BreakParent, doc.Text(selector)}
		}
		return doc.Text(selector)
	}

	return doc.Text(node.String("value"))
}

// printNamespace is the namespace prefix selector-tag, selector-attribute and selector-universal share:
// node.namespace ? [node.namespace === true ? "" : node.namespace.trim(), "|"] : "".
func printNamespace(node *estree.Node) doc.Doc {
	namespace := node.Get("namespace")
	if !estree.IsTruthy(namespace) {
		return doc.Text("")
	}
	if namespace == true {
		return doc.Text("|")
	}
	text, _ := namespace.(string)
	return doc.Concat{doc.Text(trim(text)), doc.Text("|")}
}
