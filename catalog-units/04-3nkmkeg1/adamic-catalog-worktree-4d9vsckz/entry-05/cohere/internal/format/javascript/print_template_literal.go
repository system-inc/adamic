package javascript

import (
	"math"
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
)

// print/template-literal.js.
//
// Embedded languages (css`...`, graphql`...`, html`...`, markdown`...`) are not printed here: upstream
// formats them in the printer's separate `embed` hook, language-js/embed/index.js, which this package
// does not have (printer.go). Every template literal therefore prints through printTemplateLiteral.

/*
- `TemplateLiteral`
- `TSTemplateLiteralType` (TypeScript)
*/

// printTemplateLiteral is upstream's printTemplateLiteral.
func printTemplateLiteral(path *Path, options *Options, print PrintFunc) Doc {
	if isJestEachTemplateLiteral(path) {
		printed := printJestEachTemplateLiteral(path, options, print)
		if printed != nil {
			return printed
		}
	}

	expressionDocs := printTemplateExpressions(path, options, print)
	quasiCount := len(node(path).List("quasis"))
	parts := mapPath(path, func(_ *Path, index int) Doc {
		isLast := index == quasiCount-1
		if isLast {
			return concatIn(path, print(nil, nil), "")
		}
		return concatIn(path, print(nil, nil), expressionDocs[index])
	}, "quasis")

	result := []any{lineSuffixBoundary, "`"}
	for _, part := range parts {
		result = append(result, part)
	}
	result = append(result, "`")
	return concatIn(path, result...)
}

// printTaggedTemplateExpression is upstream's printTaggedTemplateExpression.
//
// Upstream labels the result `{ tagged: true, ...quasiDoc.label }` when the quasi carries a label.
// Go labels are strings and nothing reads `tagged`, so the quasi's own label is carried through.
func printTaggedTemplateExpression(path *Path, options *Options, print PrintFunc) Doc {
	quasiDoc := print("quasi", nil)
	current := node(path)

	var space Doc = emptyDoc
	quasiLeadingComments := getComments(current.Child("quasi"), commentLeading, nil)
	if len(quasiLeadingComments) > 0 {
		quasiLeadingComment := quasiLeadingComments[0]
		// node.typeArguments ?? node.tag
		typeArgumentsOrTag := current.Child("typeArguments")
		if typeArgumentsOrTag == nil {
			typeArgumentsOrTag = current.Child("tag")
		}
		if hasNewlineInRange(
			originalText(options),
			locEnd(typeArgumentsOrTag),
			locStart(quasiLeadingComment),
		) {
			space = softline
		} else {
			space = doc.Text(" ")
		}
	}

	return label(labelOf(quasiDoc), concatIn(path, print("tag", nil),
		print("typeArguments", nil),
		space,
		lineSuffixBoundary,
		quasiDoc,
	))
}

// jestEachTableRow is one row of upstream's tableBody: `{ hasLineBreak, cells }`.
type jestEachTableRow struct {
	hasLineBreak bool
	cells        []string
}

// jestEachHeaderSeparator is upstream's /\s*\|\s*/.
var jestEachHeaderSeparator = regexp.MustCompile(
	`[\t\n\v\f\r \x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}]*\|` +
		`[\t\n\v\f\r \x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}]*`,
)

// printJestEachTemplateLiteral is upstream's printJestEachTemplateLiteral. It returns nil where upstream
// returns undefined, a header row with nothing in it.
func printJestEachTemplateLiteral(path *Path, options *Options, print PrintFunc) Doc {
	/**
	 * a    | b    | expected
	 * ${1} | ${1} | ${2}
	 * ${1} | ${2} | ${3}
	 * ${2} | ${1} | ${3}
	 */
	current := node(path)
	quasis := current.List("quasis")
	headerNames := jestEachHeaderSeparator.Split(estree.TrimJavaScript(templateElementRaw(quasis[0])), -1)
	someHeaderNameIsNonEmpty := false
	for _, headerName := range headerNames {
		if len(headerName) > 0 {
			someHeaderNameIsNonEmpty = true
			break
		}
	}
	if len(headerNames) > 1 || someHeaderNameIsNonEmpty {
		settingsOf(options).inJestEach = true
		expressions := printTemplateExpressions(path, options, print)
		settingsOf(options).inJestEach = false
		// printDocToString(doc, { ...options, printWidth: Number.POSITIVE_INFINITY, endOfLine: "lf" });
		// doc.Print always ends lines with "\n".
		stringifiedExpressions := make([]string, len(expressions))
		for index, expression := range expressions {
			stringifiedExpressions[index] = doc.Print(expression, doc.Options{
				PrintWidth: math.MaxInt,
				TabWidth:   settingsOf(options).TabWidth,
				UseTabs:    settingsOf(options).UseTabs,
			})
		}

		tableBody := []*jestEachTableRow{{hasLineBreak: false, cells: nil}}
		for i := 1; i < len(quasis); i++ {
			row := tableBody[len(tableBody)-1]
			correspondingExpression := stringifiedExpressions[i-1]

			row.cells = append(row.cells, correspondingExpression)
			if strings.Contains(correspondingExpression, "\n") {
				row.hasLineBreak = true
			}

			if strings.Contains(templateElementRaw(quasis[i]), "\n") {
				tableBody = append(tableBody, &jestEachTableRow{hasLineBreak: false, cells: nil})
			}
		}

		maxColumnCount := len(headerNames)
		for _, row := range tableBody {
			maxColumnCount = max(maxColumnCount, len(row.cells))
		}

		maxColumnWidths := make([]int, maxColumnCount)
		table := []*jestEachTableRow{{cells: headerNames}}
		for _, row := range tableBody {
			if len(row.cells) > 0 {
				table = append(table, row)
			}
		}
		for _, row := range table {
			if row.hasLineBreak {
				continue
			}

			for index, cell := range row.cells {
				// getStringWidth(cell)
				maxColumnWidths[index] = max(
					maxColumnWidths[index],
					doc.StringWidth(cell),
				)
			}
		}

		rowDocs := make([]Doc, len(table))
		for rowIndex, row := range table {
			cellDocs := make([]Doc, len(row.cells))
			for index, cell := range row.cells {
				if row.hasLineBreak {
					cellDocs[index] = doc.Text(cell)
				} else {
					cellDocs[index] = doc.Text(cell + strings.Repeat(" ", maxColumnWidths[index]-doc.StringWidth(cell)))
				}
			}
			rowDocs[rowIndex] = join(" | ", cellDocs)
		}

		return concatIn(path, lineSuffixBoundary,
			"`",
			indentIn(path, concatIn(path, hardline,
				join(hardline, rowDocs),
			)),
			hardline,
			"`",
		)
	}
	return nil
}

// templateLiteralExpressionIndent is one entry of upstream's per-template-literal sizes:
// `{ indentSize, previousQuasiText }`.
type templateLiteralExpressionIndent struct {
	indentSize        int
	previousQuasiText string
}

// getTemplateLiteralExpressionIndent is upstream's getTemplateLiteralExpressionIndent. Upstream caches
// the sizes per template literal in a WeakMap (templateLiteralIndentCache); they depend only on the
// node and tabWidth, so they are recomputed here instead.
func getTemplateLiteralExpressionIndent(path *Path, options *Options) templateLiteralExpressionIndent {
	templateLiteral, index := parentOf(path), indexOf(path)
	tabWidth := settingsOf(options).TabWidth
	previousQuasiIndentSize := 0
	quasis := templateLiteral.List("quasis")
	sizes := make([]templateLiteralExpressionIndent, len(quasis))
	for quasiIndex, quasi := range quasis {
		text := templateElementRaw(quasi)
		indentSize := previousQuasiIndentSize
		if strings.Contains(text, "\n") {
			indentSize = getIndentSize(text, tabWidth)
		}
		previousQuasiIndentSize = indentSize
		sizes[quasiIndex] = templateLiteralExpressionIndent{indentSize: indentSize, previousQuasiText: text}
	}

	return sizes[index]
}

/*
- `TemplateLiteral`
- `TSTemplateLiteralType` (TypeScript)
*/

// printTemplateExpression is upstream's printTemplateExpression.
func printTemplateExpression(path *Path, options *Options, print PrintFunc) Doc {
	current, index := node(path), indexOf(path)
	expressionDoc := print(nil, nil)

	templateLiteral := parentOf(path)
	quasis := templateLiteral.List("quasis")
	start := locEnd(quasis[index])
	end := locStart(quasis[index+1])

	interpolationHasNewline := hasNewlineInRange(
		originalText(options),
		start,
		end,
	)

	if !interpolationHasNewline {
		// Never add a newline to an interpolation which didn't already have one...
		renderedExpression := doc.Print(expressionDoc, doc.Options{
			PrintWidth: math.MaxInt,
			TabWidth:   settingsOf(options).TabWidth,
			UseTabs:    settingsOf(options).UseTabs,
		})

		// ... unless one will be introduced anyway, e.g. by a nested function.
		// This case is rare, so we can pay the cost of re-rendering.
		if strings.Contains(renderedExpression, "\n") {
			interpolationHasNewline = true
		} else {
			expressionDoc = doc.Text(renderedExpression)
		}
	}

	// Breaks at the template element boundaries (${ and }) are preferred to breaking
	// in the middle of a MemberExpression
	if interpolationHasNewline &&
		(hasAnyComment(current) ||
			current.Is("Identifier") ||
			isMemberExpression(stripChainElementWrappers(current)) ||
			current.Is("ConditionalExpression") ||
			current.Is("SequenceExpression") ||
			isBinaryCastExpression(current) ||
			isBinaryish(current)) {
		expressionDoc = concatIn(path, indentIn(path, concatIn(path, softline, expressionDoc)), softline)
	}

	// For a template literal of the following form:
	//   `someQuery {
	//     ${call({
	//       a,
	//       b,
	//     })}
	//   }`
	// the expression is on its own line (there is a \n in the previous
	// quasi literal), therefore we want to indent the JavaScript
	// expression inside at the beginning of ${ instead of the beginning
	// of the `.
	expressionIndent := getTemplateLiteralExpressionIndent(path, options)
	indentSize := expressionIndent.indentSize
	// In `jest.each`, we know expression will at least indent 2 level
	if settingsOf(options).inJestEach {
		indentSize = max(indentSize, settingsOf(options).TabWidth)
	}
	if indentSize == 0 && strings.HasSuffix(expressionIndent.previousQuasiText, "\n") {
		// align(Number.NEGATIVE_INFINITY, expressionDoc)
		expressionDoc = dedentToRoot(expressionDoc)
	} else {
		expressionDoc = addAlignmentToDoc(expressionDoc, indentSize, settingsOf(options).TabWidth)
	}

	return groupIn(path, concatIn(path, "${", expressionDoc, lineSuffixBoundary, "}"))
}

// printTemplateExpressions is upstream's printTemplateExpressions.
func printTemplateExpressions(path *Path, options *Options, print PrintFunc) []Doc {
	property := "expressions"
	if node(path).Is("TSTemplateLiteralType") {
		property = "types"
	}
	return mapPath(path, func(path *Path, _ int) Doc {
		return printTemplateExpression(path, options, print)
	}, property)
}

// templateBackticks is upstream's /(\\*)`/g.
var templateBackticks = regexp.MustCompile("(\\\\*)`")

// escapeTemplateCharacters is upstream's escapeTemplateCharacters. Only the embed printers call it.
func escapeTemplateCharacters(document Doc, raw bool) Doc {
	return doc.MapDoc(document, func(currentDoc Doc) Doc {
		if text, isText := currentDoc.(doc.Text); isText {
			if raw {
				return doc.Text(templateBackticks.ReplaceAllString(string(text), "${1}${1}\\`"))
			}
			return doc.Text(uncookTemplateElementValue(string(text)))
		}

		return currentDoc
	})
}

// templateCookedSpecials is upstream's /([\\`]|\$\{)/g.
var templateCookedSpecials = regexp.MustCompile("([\\\\`]|\\$\\{)")

// uncookTemplateElementValue is upstream's uncookTemplateElementValue. Only the embed printers call it.
func uncookTemplateElementValue(cookedValue string) string {
	return templateCookedSpecials.ReplaceAllString(cookedValue, "\\${1}")
}

/**
 * describe.each`table`(name, fn)
 * describe.only.each`table`(name, fn)
 * describe.skip.each`table`(name, fn)
 * test.each`table`(name, fn)
 * test.only.each`table`(name, fn)
 * test.skip.each`table`(name, fn)
 *
 * Ref: https://github.com/facebook/jest/pull/6102
 */
var jestEachTriggerRegex = regexp.MustCompile(`^[fx]?(?:describe|it|test)$`)

// isJestEachTemplateLiteral is upstream's isJestEachTemplateLiteral.
func isJestEachTemplateLiteral(path *Path) bool {
	current, parent := node(path), parentOf(path)
	tag := parent.Child("tag")
	return current.Is("TemplateLiteral") &&
		parent.Is("TaggedTemplateExpression") &&
		parent.Child("quasi") == current &&
		tag.Is("MemberExpression") &&
		tag.Child("property").Is("Identifier") &&
		tag.Child("property").String("name") == "each" &&
		((tag.Child("object").Is("Identifier") &&
			jestEachTriggerRegex.MatchString(tag.Child("object").String("name"))) ||
			(tag.Child("object").Is("MemberExpression") &&
				tag.Child("object").Child("property").Is("Identifier") &&
				(tag.Child("object").Child("property").String("name") == "only" ||
					tag.Child("object").Child("property").String("name") == "skip") &&
				tag.Child("object").Child("object").Is("Identifier") &&
				jestEachTriggerRegex.MatchString(tag.Child("object").Child("object").String("name"))))
}
