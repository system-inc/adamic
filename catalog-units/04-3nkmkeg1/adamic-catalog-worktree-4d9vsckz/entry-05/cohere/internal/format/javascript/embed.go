package javascript

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// embed/index.js, embed/utilities.js, embed/graphql.js, embed/css.js, and the tests of embed/html.js
// and embed/markdown.js: template literals in another language, printed by that language's printer.
//
// GraphQL and CSS (as scss) have native printers to embed. HTML, Angular and markdown keep upstream's
// tests, because a whitespace-only template in any embedded language prints as `` before its printer
// is consulted, and their print fails, which upstream treats as an embed that could not format: the
// template prints as written. Markdown also needs its printer's __inJsTemplate option (fences in `~` rather than
// backticks), which the markdown port does not carry yet.

// embedPrint is a printer upstream's embed returns: given textToDoc, the doc for the template.
type embedPrint = func(textToDoc printing.TextToDoc, print PrintFunc, path *Path, options *Options) (Doc, error)

type embedPrinter struct {
	test  func(path *Path) bool
	print embedPrint
}

// embedPrinters is upstream's printers list, in its order: the first whose test passes prints.
var embedPrinters = []embedPrinter{
	{test: isEmbedCss, print: printEmbedCss},
	{test: isEmbedGraphQL, print: printEmbedGraphQL},
	{test: isEmbedHtml, print: printEmbedUnsupported("html")},
	{test: isAngularComponentTemplate, print: printEmbedUnsupported("angular")},
	{test: isEmbedMarkdown, print: printEmbedUnsupported("markdown")},
}

// embed is upstream's embed, the printer's embed hook.
func embed(path *Path, options *Options) embedPrint {
	current := node(path)
	if !current.Is("TemplateLiteral") ||
		// Bail out if any of the quasis have an invalid escape sequence
		// (which would make the `cooked` value be `null`)
		hasInvalidCookedValue(current) {
		return nil
	}

	var printer *embedPrinter
	for index := range embedPrinters {
		if embedPrinters[index].test(path) {
			printer = &embedPrinters[index]
			break
		}
	}
	if printer == nil {
		return nil
	}

	// Special case: whitespace-only template literals
	if quasis := current.List("quasis"); len(quasis) == 1 && estree.TrimJavaScript(templateElementRaw(quasis[0])) == "" {
		return func(printing.TextToDoc, PrintFunc, *Path, *Options) (Doc, error) { return doc.Text("``"), nil }
	}

	return createTemplateLiteralPrint(printer.print)
}

// holdsTemplateLiteral reports whether a TemplateLiteral is reachable from node through its visitor keys:
// embed answers for nothing else, so a tree without one has no embed to find. It walks the nodes directly,
// allocating nothing, where the embed walk goes through the path.
func holdsTemplateLiteral(node *estree.Node) bool {
	if node == nil {
		return false
	}
	if node.Is("TemplateLiteral") {
		return true
	}
	for _, key := range estree.VisitorKeys(node) {
		switch child := node.Get(key).(type) {
		case *estree.Node:
			if holdsTemplateLiteral(child) {
				return true
			}
		case []*estree.Node:
			for _, element := range child {
				if holdsTemplateLiteral(element) {
					return true
				}
			}
		}
	}
	return false
}

// createTemplateLiteralPrint is upstream's createTemplateLiteralPrint: the printed doc is labelled as
// an embed, which print/call-arguments.js and print/arrow-function.js read.
func createTemplateLiteralPrint(print embedPrint) embedPrint {
	return func(textToDoc printing.TextToDoc, printFunction PrintFunc, path *Path, options *Options) (Doc, error) {
		printed, err := print(textToDoc, printFunction, path, options)
		if err != nil || printed == nil {
			return nil, err
		}
		return label("embed", printed), nil
	}
}

func hasInvalidCookedValue(node Node) bool {
	for _, quasi := range node.List("quasis") {
		if value, _ := quasi.Get("value").(*estree.TemplateValue); value == nil || value.Cooked == nil {
			return true
		}
	}
	return false
}

// printEmbedUnsupported stands in for an embedded language with no native printer: it fails, and the
// template prints as written, as upstream does when an embedded format throws.
func printEmbedUnsupported(language string) embedPrint {
	return func(textToDoc printing.TextToDoc, _ PrintFunc, _ *Path, _ *Options) (Doc, error) {
		return nil, errUnsupportedEmbed(language)
	}
}

type errUnsupportedEmbed string

func (language errUnsupportedEmbed) Error() string {
	return "no native printer embeds " + string(language) + " in a template literal yet"
}

// predicate adapts a (node, key) test to the print core's path.match predicate.
func predicate(test func(node Node, name string) bool) printing.Predicate[Node] {
	return func(value any, name any, _ int, _ bool) bool {
		current, _ := value.(Node)
		key, _ := name.(string)
		return current != nil && test(current, key)
	}
}

// embed/utilities.js

var angularComponentObjectExpressionPredicates = []printing.Predicate[Node]{
	predicate(func(node Node, name string) bool { return name == "properties" && node.Is("ObjectExpression") }),
	predicate(func(node Node, name string) bool {
		return name == "arguments" && node.Is("CallExpression") && node.Child("callee").Is("Identifier") &&
			node.Child("callee").String("name") == "Component"
	}),
	predicate(func(node Node, name string) bool { return name == "expression" && node.Is("Decorator") }),
}

func isTemplateLiteralPredicate() printing.Predicate[Node] {
	return predicate(func(node Node, _ string) bool { return node.Is("TemplateLiteral") })
}

// isObjectPropertyNamed is upstream's isObjectPropertyNamedStyles and the template property test.
func isObjectPropertyNamed(name string) printing.Predicate[Node] {
	return predicate(func(node Node, key string) bool {
		return isObjectProperty(node) && !node.Bool("computed") && node.Child("key").Is("Identifier") &&
			node.Child("key").String("name") == name && key == "value"
	})
}

// isAngularComponentStyles is upstream's isAngularComponentStyles.
func isAngularComponentStyles(path *Path) bool {
	return path.Match(append([]printing.Predicate[Node]{
		isTemplateLiteralPredicate(),
		predicate(func(node Node, name string) bool { return isArrayExpression(node) && name == "elements" }),
		isObjectPropertyNamed("styles"),
	}, angularComponentObjectExpressionPredicates...)...) ||
		path.Match(append([]printing.Predicate[Node]{
			isTemplateLiteralPredicate(),
			isObjectPropertyNamed("styles"),
		}, angularComponentObjectExpressionPredicates...)...)
}

// isAngularComponentTemplate is upstream's isAngularComponentTemplate.
func isAngularComponentTemplate(path *Path) bool {
	return path.Match(append([]printing.Predicate[Node]{
		isTemplateLiteralPredicate(),
		isObjectPropertyNamed("template"),
	}, angularComponentObjectExpressionPredicates...)...)
}

// hasLeadingBlockCommentWithName is upstream's hasLeadingBlockCommentWithName.
//
// This checks for a leading comment that is exactly `/* GraphQL */`
// In order to be in line with other implementations of this comment tag
// we will not trim the comment value and we will expect exactly one space on
// either side of the GraphQL string
func hasLeadingBlockCommentWithName(node Node, languageName string) bool {
	return hasComment(node, commentBlock|commentLeading, func(comment Node) bool {
		return comment.String("value") == " "+languageName+" "
	})
}

// hasLanguageComment is upstream's hasLanguageComment.
func hasLanguageComment(path *Path, languageName string) bool {
	current, parent := node(path), parentOf(path)
	return hasLeadingBlockCommentWithName(current, languageName) ||
		isAsConstExpression(parent) && hasLeadingBlockCommentWithName(parent, languageName) ||
		parent.Is("ExpressionStatement") && hasLeadingBlockCommentWithName(parent, languageName)
}

// isAsConstExpression is upstream's isAsConstExpression in embed/utilities.js.
func isAsConstExpression(node Node) bool {
	annotation := node.Child("typeAnnotation")
	return node.Is("AsConstExpression") ||
		node.Is("TSAsExpression") && annotation.Is("TSTypeReference") &&
			annotation.Child("typeName").Is("Identifier") && annotation.Child("typeName").String("name") == "const"
}

// embed/css.js

// printEmbedCss is upstream's printEmbedCss: the template with each expression replaced by a
// placeholder, formatted as scss, and the expressions put back.
func printEmbedCss(textToDoc printing.TextToDoc, print PrintFunc, path *Path, options *Options) (Doc, error) {
	current := node(path)

	// Get full template literal with expressions replaced by placeholders
	var text strings.Builder
	for index, quasi := range current.List("quasis") {
		if index > 0 {
			text.WriteString("@prettier-placeholder-" + strconv.Itoa(index-1) + "-id")
		}
		text.WriteString(templateElementRaw(quasi))
	}
	quasisDoc, err := textToDoc(text.String(), "scss")
	if err != nil {
		return nil, err
	}
	expressionDocs := printTemplateExpressions(path, options, print)
	newDoc := replacePlaceholders(quasisDoc, expressionDocs)
	if newDoc == nil {
		return nil, errCouldNotInsertExpressions
	}
	return concatIn(path, "`", indentIn(path, concatIn(path, hardline, newDoc)), softline, "`"), nil
}

var errCouldNotInsertExpressions = errors.New("Couldn't insert all the expressions")

var placeholderPattern = regexp.MustCompile(`@prettier-placeholder-(\d+)-id`)

// replacePlaceholders is upstream's replacePlaceholders: search all the placeholders in the quasisDoc
// tree and replace them with the expression docs one by one. It returns nil when it could not replace
// every expression.
func replacePlaceholders(quasisDoc Doc, expressionDocs []Doc) Doc {
	if len(expressionDocs) == 0 {
		return quasisDoc
	}
	replaceCounter := 0
	newDoc := doc.MapDoc(doc.CleanDoc(quasisDoc), func(current Doc) Doc {
		text, isText := current.(doc.Text)
		if !isText || !strings.Contains(string(text), "@prettier-placeholder") {
			return current
		}
		// When we have multiple placeholders in one line, like:
		// ${Child}${Child2}:not(:first-child)
		//
		// Upstream splits on the pattern with its capture group kept, so the pieces alternate between
		// text (even indexes) and placeholder numbers (odd), which this rebuilds.
		parts := []Doc{}
		position := 0
		for _, match := range placeholderPattern.FindAllStringSubmatchIndex(string(text), -1) {
			parts = append(parts, replaceEndOfLine(string(text)[position:match[0]]))
			number, _ := strconv.Atoi(string(text)[match[2]:match[3]])
			replaceCounter++
			if number < len(expressionDocs) {
				parts = append(parts, expressionDocs[number])
			} else {
				// expressionDocs[component] is undefined upstream, and the doc printer throws on it.
				parts = append(parts, nil)
			}
			position = match[1]
		}
		parts = append(parts, replaceEndOfLine(string(text)[position:]))
		return doc.Concat(parts)
	})
	if len(expressionDocs) != replaceCounter {
		return nil
	}
	return newDoc
}

// embed/css.js's tests.

func isStyledJsx(path *Path) bool {
	return path.Match(
		nil,
		predicate(func(node Node, key string) bool {
			return key == "quasi" && node.Is("TaggedTemplateExpression") &&
				isNodeMatches(node.Child("tag"), []string{"css", "css.global", "css.resolve"})
		}),
	) ||
		path.Match(
			nil,
			predicate(func(node Node, key string) bool { return key == "expression" && node.Is("JSXExpressionContainer") }),
			predicate(func(node Node, key string) bool {
				name := node.Child("openingElement").Child("name")
				if key != "children" || !node.Is("JSXElement") || !name.Is("JSXIdentifier") || name.String("name") != "style" {
					return false
				}
				for _, attribute := range node.Child("openingElement").List("attributes") {
					if attribute.Is("JSXAttribute") && attribute.Child("name").Is("JSXIdentifier") &&
						attribute.Child("name").String("name") == "jsx" {
						return true
					}
				}
				return false
			}),
		)
}

func isStyledIdentifier(node Node) bool {
	return node.Is("Identifier") && node.String("name") == "styled"
}

var leadingUppercase = regexp.MustCompile(`^[A-Z]`)

func isStyledExtend(node Node) bool {
	return leadingUppercase.MatchString(node.Child("object").String("name")) && node.Child("property").String("name") == "extend"
}

// isStyledComponents is upstream's isStyledComponents: styled-components template literals.
func isStyledComponents(path *Path) bool {
	parent := parentOf(path)
	if !parent.Is("TaggedTemplateExpression") {
		return false
	}
	tag := parent.Child("tag")
	if tag.Is("ParenthesizedExpression") {
		tag = tag.Child("expression")
	}
	switch tag.Type() {
	case "MemberExpression":
		// styled.foo`` or Component.extend``
		return isStyledIdentifier(tag.Child("object")) || isStyledExtend(tag)
	case "CallExpression":
		callee := tag.Child("callee")
		return isStyledIdentifier(callee) || // styled(Component)``
			callee.Is("MemberExpression") &&
				(callee.Child("object").Is("MemberExpression") &&
					// styled.foo.attrs({})`` or Component.extend.attrs({})``
					(isStyledIdentifier(callee.Child("object").Child("object")) || isStyledExtend(callee.Child("object"))) ||
					// styled(Component).attrs({})``
					callee.Child("object").Is("CallExpression") && isStyledIdentifier(callee.Child("object").Child("callee")))
	case "Identifier":
		// css``
		return tag.String("name") == "css"
	}
	return false
}

// isCssProp is upstream's isCssProp: the css={`...`} JSX attribute.
func isCssProp(path *Path) bool {
	parent, grandparent := parentOf(path), grandparentOf(path)
	return grandparent.Is("JSXAttribute") && parent.Is("JSXExpressionContainer") &&
		grandparent.Child("name").Is("JSXIdentifier") && grandparent.Child("name").String("name") == "css"
}

func isEmbedCss(path *Path) bool {
	return isStyledJsx(path) || isStyledComponents(path) || isCssProp(path) || isAngularComponentStyles(path)
}

// embed/html.js's test.
func isEmbedHtml(path *Path) bool {
	return hasLanguageComment(path, "HTML") ||
		path.Match(
			isTemplateLiteralPredicate(),
			predicate(func(node Node, name string) bool {
				return node.Is("TaggedTemplateExpression") && node.Child("tag").Is("Identifier") &&
					node.Child("tag").String("name") == "html" && name == "quasi"
			}),
		)
}

// embed/markdown.js's test: md`...` and markdown`...`.
func isEmbedMarkdown(path *Path) bool {
	parent := parentOf(path)
	return parent.Is("TaggedTemplateExpression") && len(node(path).List("quasis")) == 1 &&
		parent.Child("tag").Is("Identifier") &&
		(parent.Child("tag").String("name") == "md" || parent.Child("tag").String("name") == "markdown")
}

// embed/graphql.js

// printEmbedGraphQL is upstream's printEmbedGraphQL.
func printEmbedGraphQL(textToDoc printing.TextToDoc, print PrintFunc, path *Path, options *Options) (Doc, error) {
	current := node(path)
	quasis := current.List("quasis")
	numQuasis := len(quasis)

	expressionDocs := printTemplateExpressions(path, options, print)
	parts := []Doc{}

	for index := 0; index < numQuasis; index++ {
		templateElement := quasis[index]
		isFirst := index == 0
		isLast := index == numQuasis-1
		text := *templateElement.Get("value").(*estree.TemplateValue).Cooked

		lines := strings.Split(text, "\n")
		numLines := len(lines)

		// Bail out if an interpolation occurs within a comment.
		if !isLast && commentAtLineEnd.MatchString(lines[numLines-1]) {
			return nil, nil
		}

		startsWithBlankLine := numLines > 2 && estree.TrimJavaScript(lines[0]) == "" && estree.TrimJavaScript(lines[1]) == ""
		endsWithBlankLine := numLines > 2 && estree.TrimJavaScript(lines[numLines-1]) == "" &&
			estree.TrimJavaScript(lines[numLines-2]) == ""

		commentsAndWhitespaceOnly := true
		for _, line := range lines {
			if !isCommentOrWhitespaceLine(line) {
				commentsAndWhitespaceOnly = false
				break
			}
		}

		var printed Doc
		if commentsAndWhitespaceOnly {
			printed = printGraphqlComments(lines)
		} else {
			var err error
			if printed, err = textToDoc(text, "graphql"); err != nil {
				return nil, err
			}
		}

		if printed != nil {
			printed = escapeTemplateCharacters(printed, false)
			if !isFirst && startsWithBlankLine {
				parts = append(parts, emptyDoc)
			}
			parts = append(parts, printed)
			if !isLast && endsWithBlankLine {
				parts = append(parts, emptyDoc)
			}
		} else if !isFirst && !isLast && startsWithBlankLine {
			parts = append(parts, emptyDoc)
		}

		if !isLast {
			parts = append(parts, expressionDocs[index])
		}
	}

	return concatIn(path, "`", indentIn(path, concatIn(path, hardline, join(hardline, parts))), hardline, "`"), nil
}

// commentAtLineEnd is upstream's /#[^\n\r]*$/ over one line.
var commentAtLineEnd = regexp.MustCompile(`#[^\n\r]*$`)

// isCommentOrWhitespaceLine is upstream's /^\s*(?:#[^\n\r]*)?$/ over one line. JavaScript's \s reaches
// past ASCII, so the leading whitespace is trimmed with its definition rather than RE2's.
func isCommentOrWhitespaceLine(line string) bool {
	rest := estree.TrimStartJavaScript(line)
	return rest == "" || rest[0] == '#' && !strings.ContainsAny(rest, "\n\r")
}

// printGraphqlComments is upstream's printGraphqlComments. It returns nil when the lines are
// whitespace only, where upstream returns null.
func printGraphqlComments(lines []string) Doc {
	parts := []Doc{}
	seenComment := false

	trimmed := make([]string, len(lines))
	for index, line := range lines {
		trimmed[index] = estree.TrimJavaScript(line)
	}
	for index, textLine := range trimmed {
		// Lines are either whitespace only, or a comment (with potential whitespace
		// around it). Drop whitespace-only lines.
		if textLine == "" {
			continue
		}

		if index > 0 && trimmed[index-1] == "" && seenComment {
			// If a non-first comment is preceded by a blank (whitespace only) line,
			// add in a blank line.
			parts = append(parts, concat(hardline, textLine))
		} else {
			parts = append(parts, doc.Text(textLine))
		}

		seenComment = true
	}

	// If `lines` was whitespace only, return `null`.
	if len(parts) == 0 {
		return nil
	}
	return join(hardline, parts)
}

// isEmbedGraphQL is upstream's isEmbedGraphQL:
//
//	react-relay and graphql-tag
//	graphql`...`
//	graphql.experimental`...`
//	gql`...`
//	GraphQL comment block
//
// This intentionally excludes Relay Classic tags, as Prettier does not
// support Relay Classic formatting.
func isEmbedGraphQL(path *Path) bool {
	parent := parentOf(path)
	if hasLanguageComment(path, "GraphQL") {
		return true
	}
	if parent == nil {
		return false
	}
	tag := parent.Child("tag")
	return parent.Is("TaggedTemplateExpression") &&
		(tag.Is("MemberExpression") && tag.Child("object").String("name") == "graphql" &&
			tag.Child("property").String("name") == "experimental" ||
			tag.Is("Identifier") && (tag.String("name") == "gql" || tag.String("name") == "graphql")) ||
		parent.Is("CallExpression") && parent.Child("callee").Is("Identifier") &&
			parent.Child("callee").String("name") == "graphql"
}
