package css

// src/language-css/print/comma-separated-value-group.js. The branches upstream gates on options.parser
// being "scss" read the parser the format carries; the ones gated on "less" alone are left out, and a
// comment marks each place.

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

func printCommaSeparatedValueGroup(path *astPath, options *printerOptions, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	parentNode, _ := path.Parent()
	parentParentNode, _ := path.Grandparent()
	declAncestorProp := getPropOfDeclNode(path)
	isGridValue := declAncestorProp != "" &&
		parentNode.Type() == "value-value" &&
		(declAncestorProp == "grid" ||
			strings.HasPrefix(declAncestorProp, "grid-template"))
	atRuleAncestorNode, _ := path.FindAncestor(func(node *estree.Node) bool { return node.Type() == "css-atrule" })
	isControlDirective := atRuleAncestorNode != nil && isSCSSControlDirectiveNode(atRuleAncestorNode, options)
	groups := node.List("groups")
	hasInlineComment := someNode(groups, isInlineValueCommentNode)

	printed := printAll(path, print, "groups")
	/*
	 * We assume parts always meet following conditions:
	 * - parts.length is odd
	 * - odd (0-indexed) elements are line-like doc
	 * We can achieve this by following way:
	 * - if we push line-like doc, we push empty string after it
	 * - if we push non-line-like doc, push [parts.pop(), doc] instead
	 */
	parts := []doc.Doc{doc.Text("")}
	appendToLast := func(document doc.Doc) {
		parts[len(parts)-1] = doc.Concat{parts[len(parts)-1], document}
	}
	insideURLFunction := insideValueFunctionNode(path, "url")

	insideSCSSInterpolationInString := false
	didBreak := false

	groupAt := func(index int) *estree.Node {
		if index < 0 || index >= len(groups) {
			return nil
		}
		return groups[index]
	}

	for i := range groups {
		iPrevNode := groupAt(i - 1)
		iNode := groups[i]
		iNextNode := groupAt(i + 1)

		// If the node is comment and last node print it in a line suffix
		if isInlineValueCommentNode(iNode) && iNextNode == nil {
			// TODO: Improve this part
			// This `lineSuffix` should be done in `value-comment` print
			// But since we add `line` to groups in `value-paren_group`,
			// The format result looks bad for now
			appendToLast(doc.NewLineSuffix(doc.Concat{doc.Text(" "), printed[i]}))
			continue
		}

		appendToLast(printed[i])

		if insideURLFunction {
			if (iNextNode != nil && isAdditionNode(iNextNode)) || isAdditionNode(iNode) {
				appendToLast(doc.Text(" "))
			}
			continue
		}

		// Ignore SCSS @forward wildcard suffix
		if insideAtRuleNode(path, "forward") &&
			iNode.Type() == "value-word" &&
			iNode.String("value") != "" &&
			iPrevNode != nil &&
			iPrevNode.Type() == "value-word" &&
			iPrevNode.String("value") == "as" &&
			mustNode(iNextNode).Type() == "value-operator" &&
			iNextNode.String("value") == "*" {
			continue
		}

		// Ignore Tailwind `@utility` directive ( https://tailwindcss.com/docs/adding-custom-styles#functional-utilities )
		if insideAtRuleNode(path, "utility") &&
			iNode.Type() == "value-word" &&
			iNextNode != nil &&
			iNextNode.Type() == "value-operator" &&
			iNextNode.String("value") == "*" {
			continue
		}

		// Ignore after latest node (i.e. before semicolon)
		if iNextNode == nil {
			continue
		}

		// Don't print a space before the `;` branch delimiter in the SCSS `if()`
		// function (i.e. `if(condition: value; else: value)`)
		if parserOf(options) == "scss" &&
			iNextNode.Type() == "value-word" &&
			iNextNode.String("value") == ";" &&
			path.Match(
				nil,
				func(node any, key any, _ int, _ bool) bool {
					return key == "groups" && asNode(node).Type() == "value-paren_group"
				},
				func(node any, key any, _ int, _ bool) bool {
					return key == "group" && asNode(node).Type() == "value-func" && asNode(node).String("value") == "if"
				},
			) {
			continue
		}

		// We should keep spaces between words in a embedded JS expression
		// examples:
		//   styled.div` font-size: var(--font-size-h${({ level }) => level}); `;
		//   styled.div` grid-area: area-${({ area }) => area}; `;
		//   styled.div` border: 1px ${solid} red; `;
		if iNode.Type() == "value-word" &&
			isAtWordPlaceholderNode(iNextNode) &&
			locEnd(iNode) == locStart(iNextNode) {
			continue
		}

		// Ignore spaces before/after string interpolation (i.e. `"#{my-fn("_")}"`)
		if iNode.Type() == "value-string" && iNode.Truthy("quoted") {
			positionOfOpeningInterpolation := strings.LastIndex(iNode.String("value"), "#{")
			positionOfClosingInterpolation := strings.LastIndex(iNode.String("value"), "}")
			if positionOfOpeningInterpolation != -1 && positionOfClosingInterpolation != -1 {
				insideSCSSInterpolationInString = positionOfOpeningInterpolation > positionOfClosingInterpolation
			} else if positionOfOpeningInterpolation != -1 {
				insideSCSSInterpolationInString = true
			} else if positionOfClosingInterpolation != -1 {
				insideSCSSInterpolationInString = false
			}
		}

		if insideSCSSInterpolationInString {
			continue
		}

		// Ignore colon (i.e. `:`)
		if isColonNode(iNode) || isColonNode(iNextNode) {
			continue
		}

		// Ignore `@` in Less (i.e. `@@var;`)
		if iNode.Type() == "value-atword" &&
			(iNode.String("value") == "" ||
				/*
				   @var[ @notVarNested ][notVar]
				   ^^^^^
				*/
				strings.HasSuffix(iNode.String("value"), "[")) {
			continue
		}

		/*
		   @var[ @notVarNested ][notVar]
		                       ^^^^^^^^^
		*/
		if iNextNode.Type() == "value-word" && strings.HasPrefix(iNextNode.String("value"), "]") {
			continue
		}

		// Ignore `~` in Less (i.e. `content: ~"^//* some horrible but needed css hack";`)
		if iNode.String("value") == "~" {
			continue
		}

		// The Less property and variable lookups are gated on options.parser === "less".

		// Ignore escape `\`
		if iNode.Type() != "value-string" &&
			iNode.String("value") != "" &&
			strings.Contains(iNode.String("value"), `\`) &&
			iNextNode.Type() != "value-comment" {
			continue
		}

		// Ignore escaped `/`
		if iPrevNode.String("value") != "" &&
			strings.Index(iPrevNode.String("value"), `\`) == len(iPrevNode.String("value"))-1 &&
			iNode.Type() == "value-operator" &&
			iNode.String("value") == "/" {
			continue
		}

		// Ignore `\` (i.e. `$variable: \@small;`)
		if iNode.String("value") == `\` {
			continue
		}

		// Ignore `$$` (i.e. `background-color: $$(style)Color;`)
		if isPostcssSimpleVarNode(iNode, iNextNode) {
			continue
		}

		// Ignore spaces after `#` and after `{` and before `}` in SCSS interpolation (i.e. `#{variable}`)
		if isHashNode(iNode) ||
			isLeftCurlyBraceNode(iNode) ||
			isRightCurlyBraceNode(iNextNode) ||
			(isLeftCurlyBraceNode(iNextNode) && hasEmptyRawBefore(iNextNode)) ||
			(isRightCurlyBraceNode(iNode) && hasEmptyRawBefore(iNextNode)) {
			continue
		}

		// Ignore CSS variables and interpolation in SCSS (i.e. `--#{$var}`)
		if iNode.String("value") == "--" && isHashNode(iNextNode) {
			continue
		}

		// Formatting math operations
		isMathOperator := isMathOperatorNode(iNode)
		isNextMathOperator := isMathOperatorNode(iNextNode)

		// Print spaces before and after math operators beside SCSS interpolation as is
		// (i.e. `#{$var}+5`, `#{$var} +5`, `#{$var}+ 5`, `#{$var} + 5`)
		// (i.e. `5+#{$var}`, `5 +#{$var}`, `5+ #{$var}`, `5 + #{$var}`)
		if ((isMathOperator && isHashNode(iNextNode)) ||
			(isNextMathOperator && isRightCurlyBraceNode(iNode))) &&
			hasEmptyRawBefore(iNextNode) {
			continue
		}

		// Prevent the addition of a space inside type() declarations that include an + operator
		// due to the fact that it is not valid syntax
		// (i.e. `type(<number> +)`, `type(<color> +)`)
		if isAdditionNode(iNextNode) &&
			insideValueFunctionNode(path, "type") &&
			hasEmptyRawBefore(iNextNode) {
			continue
		}

		// absolute paths are only parsed as one token if they are part of url(/abs/path) call
		// but if you have custom -fb-url(/abs/path/) then it is parsed as "division /" and rest
		// of the path. We don't want to put a space after that first division in this case.
		if iPrevNode == nil && isDivisionNode(iNode) {
			continue
		}

		// Print spaces before and after addition and subtraction math operators as is in `calc` function
		// due to the fact that it is not valid syntax
		// (i.e. `calc(1px+1px)`, `calc(1px+ 1px)`, `calc(1px +1px)`, `calc(1px + 1px)`)
		if insideValueFunctionNode(path, "calc") &&
			(isAdditionNode(iNode) ||
				isAdditionNode(iNextNode) ||
				isSubtractionNode(iNode) ||
				isSubtractionNode(iNextNode)) &&
			hasEmptyRawBefore(iNextNode) {
			continue
		}

		// Space before unary minus followed by a function call.
		if parserOf(options) == "scss" &&
			isMathOperator &&
			iNode.String("value") == "-" &&
			iNextNode.Type() == "value-func" &&
			locEnd(iNode) != locStart(iNextNode) {
			appendToLast(doc.Text(" "))
			continue
		}

		iNextNextNode := groupAt(i + 2)

		// Print spaces after `+` and `-` in color adjuster functions as is (e.g. `color(red l(+ 20%))`)
		// Adjusters with signed numbers (e.g. `color(red l(+20%))`) output as-is.
		isColorAdjusterNode := (isAdditionNode(iNode) || isSubtractionNode(iNode)) &&
			i == 0 &&
			(iNextNode.Type() == "value-number" || iNextNode.Truthy("isHex")) &&
			parentParentNode != nil &&
			isColorAdjusterFuncNode(parentParentNode) &&
			!hasEmptyRawBefore(iNextNode)
		requireSpaceBeforeOperator := iNextNextNode.Type() == "value-func" ||
			(iNextNextNode != nil && isWordNode(iNextNextNode)) ||
			iNode.Type() == "value-func" ||
			isWordNode(iNode)
		requireSpaceAfterOperator := iNextNode.Type() == "value-func" ||
			isWordNode(iNextNode) ||
			iPrevNode.Type() == "value-func" ||
			(iPrevNode != nil && isWordNode(iPrevNode))

		// Formatting `/`, `+`, `-` sign
		if !(isMultiplicationNode(iNextNode) || isMultiplicationNode(iNode)) &&
			!insideValueFunctionNode(path, "calc") &&
			!isColorAdjusterNode &&
			((isDivisionNode(iNextNode) && !requireSpaceBeforeOperator) ||
				(isDivisionNode(iNode) && !requireSpaceAfterOperator) ||
				(isAdditionNode(iNextNode) && !requireSpaceBeforeOperator) ||
				(isAdditionNode(iNode) && !requireSpaceAfterOperator) ||
				isSubtractionNode(iNextNode) ||
				isSubtractionNode(iNode)) &&
			(hasEmptyRawBefore(iNextNode) ||
				(isMathOperator &&
					(iPrevNode == nil || (iPrevNode != nil && isMathOperatorNode(iPrevNode))))) {
			continue
		}

		// No space before unary minus followed by an opening parenthesis `-(`
		// (options.parser === "less" is not ported; only "scss" can be true here.)
		if parserOf(options) == "scss" &&
			isMathOperator &&
			iNode.String("value") == "-" &&
			isParenGroupNode(iNextNode) &&
			locEnd(iNode) == locStart(iNextNode.Child("open")) &&
			iNextNode.Child("open").String("value") == "(" {
			continue
		}

		// Add `hardline` after inline comment (i.e. `// comment\n foo: bar;`)
		if isInlineValueCommentNode(iNode) {
			if parentNode.Type() == "value-paren_group" {
				parts = append(parts, doc.Dedent(doc.Hardline), doc.Text(""))
				continue
			}
			parts = append(parts, doc.Hardline, doc.Text(""))
			continue
		}

		// Handle keywords in SCSS control directive
		if isControlDirective &&
			(isEqualityOperatorNode(iNextNode) ||
				isRelationalOperatorNode(iNextNode) ||
				isIfElseKeywordNode(iNextNode) ||
				isEachKeywordNode(iNode) ||
				isForKeywordNode(iNode)) {
			appendToLast(doc.Text(" "))

			continue
		}

		// At-rule `namespace` should be in one line
		if atRuleAncestorNode != nil &&
			toLowerCase(atRuleAncestorNode.String("name")) == "namespace" {
			appendToLast(doc.Text(" "))

			continue
		}

		// Formatting `grid` property
		if isGridValue {
			iNodeLine, iNodeHasSource := sourceStartLine(iNode)
			iNextNodeLine, iNextNodeHasSource := sourceStartLine(iNextNode)
			if iNodeHasSource && iNextNodeHasSource && iNodeLine != iNextNodeLine {
				parts = append(parts, doc.Hardline, doc.Text(""))

				didBreak = true
			} else {
				appendToLast(doc.Text(" "))
			}

			continue
		}

		// Formatting `font` property
		if declAncestorProp != "" &&
			(declAncestorProp == "font" || strings.HasPrefix(declAncestorProp, "--")) {
			if isDivisionNode(iNextNode) &&
				hasEmptyRawBefore(iNextNode) &&
				isPossibleFontSize(iNode) {
				continue
			}

			if isDivisionNode(iNode) &&
				hasEmptyRawBefore(iNode) &&
				isPossibleFontSize(iPrevNode) {
				continue
			}
		}

		// Add `space` before next math operation
		// Note: `grid` property have `/` delimiter and it is not math operation, so
		// `grid` property handles above
		if isNextMathOperator {
			appendToLast(doc.Text(" "))

			continue
		}
		// allow function(returns-list($list)...)
		if iNextNode.String("value") == "..." {
			continue
		}

		if isAtWordPlaceholderNode(iNode) &&
			isAtWordPlaceholderNode(iNextNode) &&
			locEnd(iNode) == locStart(iNextNode) {
			continue
		}

		if isAtWordPlaceholderNode(iNode) &&
			isParenGroupNode(iNextNode) &&
			locEnd(iNode) == locStart(iNextNode.Child("open")) {
			parts = append(parts, doc.Softline, doc.Text(""))
			continue
		}

		if iNode.String("value") == "with" && isParenGroupNode(iNextNode) {
			parts = []doc.Doc{doc.Concat{doc.NewFill(parts), doc.Text(" ")}}
			continue
		}

		if strings.HasSuffix(iNode.String("value"), "#") &&
			iNextNode.String("value") == "{" &&
			isParenGroupNode(mustNode(iNextNode.Child("group"))) {
			continue
		}

		// don't print line when the next node is a comment and last node
		// it will be printed with the comment in a line suffix
		if isInlineValueCommentNode(iNextNode) && iNextNextNode == nil {
			continue
		}

		// align value after block comment
		if atRuleAncestorNode == nil &&
			iNode.Type() == "value-comment" &&
			!iNode.Truthy("inline") &&
			everyNode(groups[:i], func(group *estree.Node) bool { return group.Type() == "value-comment" }) {
			parts = append(parts, doc.Dedent(doc.LineDoc), doc.Text(""))
			continue
		}

		// Be default all values go through `line`
		parts = append(parts, doc.LineDoc, doc.Text(""))
	}

	if hasInlineComment {
		appendToLast(doc.BreakParent)
	}

	if didBreak {
		parts = append([]doc.Doc{doc.Text(""), doc.Hardline}, parts...)
	}

	if isControlDirective {
		return doc.NewGroup(doc.NewIndent(doc.Concat(parts)), doc.GroupOptions{})
	}

	// Indent is not needed for import URL when URL is very long
	// and node has two groups
	// when type is value-comma_group
	// example @import url("verylongurl") projection,tv
	if insideURLFunctionInImportAtRuleNode(path) {
		return doc.NewGroup(doc.NewFill(parts), doc.GroupOptions{})
	}

	return doc.NewGroup(doc.NewIndent(doc.NewFill(parts)), doc.GroupOptions{})
}

func isPossibleFontSize(node *estree.Node) bool {
	if node.Type() == "value-number" {
		return true
	}

	if node.Type() != "value-func" {
		return false
	}

	value := toLowerCase(node.String("value"))
	return value == "var" ||
		value == "calc" ||
		value == "min" ||
		value == "max" ||
		value == "clamp" ||
		strings.HasPrefix(value, "--")
}
