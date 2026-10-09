package markdown

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// src/language-markdown/print/mdast.js, for the markdown parser.

type (
	astPath = printing.AstPath[*Node]
	options = printing.Options[*Node]
)

// settings are the Prettier options the markdown printer reads, carried on options.Settings.
type settings struct {
	proseWrap   string
	singleQuote bool
	tabWidth    int
	printWidth  int
	useTabs     bool

	// nodes is the arena this format's split text comes from, released once the file is printed.
	nodes *arena.Arena[Node]
}

func settingsOf(options *options) *settings { return options.Settings.(*settings) }

func currentNode(path *astPath) *Node {
	node, _ := path.Node()
	return node
}

func previousNode(path *astPath) *Node {
	node, _ := path.Previous()
	return node
}

func nextNode(path *astPath) *Node {
	node, _ := path.Next()
	return node
}

func parentNode(path *astPath) *Node {
	node, _ := path.Parent()
	return node
}

func pathIndex(path *astPath) int {
	index, _ := path.Index()
	return index
}

// siblingAt is path.siblings[index], nil when out of range or not in a list.
func siblingAt(path *astPath, index int) *Node {
	siblings, isList := path.Siblings().([]*Node)
	if !isList || index < 0 || index >= len(siblings) {
		return nil
	}
	return siblings[index]
}

var spaceAtEndPattern = func(text string) bool {
	if text == "" {
		return false
	}
	runes := []rune(text)
	return inClass(spaceSeparatorRanges, runes[len(runes)-1])
}

var spaceAtStartPattern = func(text string) bool {
	if text == "" {
		return false
	}
	return inClass(spaceSeparatorRanges, []rune(text)[0])
}

func prevOrNextWord(path *astPath) bool {
	previous, next := previousNode(path), nextNode(path)
	return (previous != nil && previous.NodeType == "sentence" &&
		len(previous.Children) > 0 && previous.Children[len(previous.Children)-1].NodeType == "word" &&
		!previous.Children[len(previous.Children)-1].HasTrailingPunctuation &&
		// https://spec.commonmark.org/0.31.2/#unicode-whitespace-character
		!spaceAtEndPattern(previous.Children[len(previous.Children)-1].Value)) ||
		(next != nil && next.NodeType == "sentence" &&
			len(next.Children) > 0 && next.Children[0].NodeType == "word" &&
			!next.Children[0].HasLeadingPunctuation &&
			!spaceAtStartPattern(next.Children[0].Value))
}

func hasFakeWhitespaceAfterNextToken(path *astPath) bool {
	if path.Siblings() == nil {
		return false
	}
	if _, present := path.Index(); !present {
		return false
	}
	afterNext := siblingAt(path, pathIndex(path)+2)
	return afterNext != nil && afterNext.NodeType == "whitespace" && afterNext.Value == ""
}

func isNextTokenFakeSetextH2Line(path *astPath) bool {
	next := nextNode(path)
	if !isNewLine(currentNode(path)) || next == nil || next.Value != "-" {
		return false
	}

	afterNext := siblingAt(path, pathIndex(path)+2)
	return afterNext == nil || isNewLine(afterNext)
}

var syntaxLeadingPattern = regexp.MustCompile(`^>|^(?:[*+-]|#{1,6}|\d+[).])$`)

func printMdast(path *astPath, options *options, print printing.PrintFunc, _ any) doc.Doc {
	node := currentNode(path)
	settings := settingsOf(options)

	if node.NodeType == "frontMatter" {
		// The core's front matter support prints it raw (src/main/front-matter/print.js).
		return doc.Text(node.FrontMatter.Raw)
	}

	if shouldRemainTheSameContent(path) {
		// We assume parts always meet following conditions:
		// - parts.length is odd
		// - odd (0-indexed) elements are line-like doc
		parts := []doc.Doc{doc.Text("")}
		textsNodes := splitText(options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset], settings.nodes)
		for _, textNode := range textsNodes {
			if textNode.NodeType == "word" {
				parts[len(parts)-1] = doc.Concat{parts[len(parts)-1], doc.Text(textNode.Value)}
				continue
			}
			printed := printWhitespace(path, textNode.Value, settings.proseWrap, true, options)
			if _, isText := printed.(doc.Text); isText {
				parts[len(parts)-1] = doc.Concat{parts[len(parts)-1], printed}
				continue
			}
			// In this path, doc is line. To meet the condition, we need additional element "".
			parts = append(parts, printed, doc.Text(""))
		}
		return doc.NewFill(parts)
	}

	switch node.NodeType {
	case "root":
		if len(node.Children) == 0 {
			return doc.Text("")
		}
		return doc.Concat{printRoot(path, options, print), doc.Hardline}
	case "paragraph":
		return printParagraph(path, print)
	case "sentence":
		return printSentence(path, print)
	case "word":
		return printWord(path, options)
	case "whitespace":
		next := nextNode(path)

		proseWrap := settings.proseWrap
		if next != nil &&
			// leading char that may cause different syntax
			syntaxLeadingPattern.MatchString(next.Value) &&
			// Avoid https://github.com/prettier/prettier/issues/18861
			!hasFakeWhitespaceAfterNextToken(path) &&
			// Next fake setext h2 `-` is going to be escaped, so no need to join adjacent words
			!(settings.proseWrap == "preserve" && isNextTokenFakeSetextH2Line(path)) {
			proseWrap = "never"
		}

		return printWhitespace(path, node.Value, proseWrap, false, options)
	case "emphasis":
		var style string
		if len(node.Children) > 0 && isAutolink(node.Children[0]) {
			style = options.OriginalText[node.Position.Start.Offset : node.Position.Start.Offset+1]
		} else {
			hasPrevOrNextWord := prevOrNextWord(path) // `1*2*3` is considered emphasis but `1_2_3` is not
			// `1***2***3` is considered strong emphasis but `1**_2_**3` is not
			inStrongAndHasPrevOrNextWord := printing.CallParent(path, func(path *astPath) bool {
				return currentNode(path).NodeType == "strong" && prevOrNextWord(path)
			}, 0)
			if hasPrevOrNextWord || inStrongAndHasPrevOrNextWord ||
				path.HasAncestor(func(node *Node) bool { return node.NodeType == "emphasis" }) {
				style = "*"
			} else {
				style = "_"
			}
		}
		return doc.Concat{doc.Text(style), printChildren(path, options, print, nil), doc.Text(style)}
	case "strong":
		return doc.Concat{doc.Text("**"), printChildren(path, options, print, nil), doc.Text("**")}
	case "delete":
		return doc.Concat{doc.Text("~~"), printChildren(path, options, print, nil), doc.Text("~~")}
	case "inlineCode":
		code := node.Value
		if settings.proseWrap != "preserve" {
			code = strings.ReplaceAll(node.Value, "\n", " ")
		}
		if path.HasAncestor(func(node *Node) bool { return node.NodeType == "tableCell" }) {
			code = strings.ReplaceAll(code, "|", `\|`)
		}
		backtickCount := getMinNotPresentContinuousCount(code, "`")
		backtickString := strings.Repeat("`", backtickCount)
		padding := ""
		if strings.HasPrefix(code, "`") || strings.HasSuffix(code, "`") ||
			(len(code) > 0 && strings.ContainsAny(code[:1], "\n ") && strings.ContainsAny(code[len(code)-1:], "\n ") &&
				strings.ContainsFunc(code, func(character rune) bool { return character != '\n' && character != ' ' })) {
			padding = " "
		}
		return doc.Concat{doc.Text(backtickString), doc.Text(padding), doc.Text(code), doc.Text(padding), doc.Text(backtickString)}
	case "wikiLink":
		if node.ValueNull {
			// Upstream prints a null value into the doc, which the doc printer rejects.
			panic("markdown: a wiki link with no target")
		}
		contents := node.Value
		if settings.proseWrap != "preserve" {
			contents = tabsAndNewlinesPattern.ReplaceAllString(node.Value, " ")
		}

		return doc.Concat{doc.Text("[["), doc.Text(contents), doc.Text("]]")}
	case "link":
		switch options.OriginalText[node.Position.Start.Offset] {
		case '<':
			mailto := "mailto:"
			link := node.URL
			// <hello@example.com> is parsed as { url: "mailto:hello@example.com" }
			start := node.Position.Start.Offset + 1
			end := min(start+len(mailto), len(options.OriginalText))
			if strings.HasPrefix(node.URL, mailto) && options.OriginalText[start:end] != mailto {
				link = node.URL[len(mailto):]
			}
			return doc.Concat{doc.Text("<"), doc.Text(link), doc.Text(">")}
		case '[':
			destination := doc.Text("<>")
			if node.URL != "" {
				destination = doc.Text(printURL(node.URL, ")"))
			}
			return doc.Concat{
				doc.Text("["),
				printChildren(path, options, print, nil),
				doc.Text("]("),
				destination,
				doc.Text(printTitle(node.Title, settings, true)),
				doc.Text(")"),
			}
		default:
			return doc.Text(options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset])
		}
	case "image":
		destination := "<>"
		if node.URL != "" {
			destination = printURL(node.URL, ")")
		}
		return doc.Concat{
			doc.Text("!["),
			doc.Text(printImageAlt(node)),
			doc.Text("]("),
			doc.Text(destination),
			doc.Text(printTitle(node.Title, settings, true)),
			doc.Text(")"),
		}
	case "blockquote":
		return doc.Concat{doc.Text("> "), doc.AlignWithString("> ", printChildren(path, options, print, nil))}
	case "heading":
		return printHeading(path, options, print)
	case "code":
		if node.IsIndented {
			// indented code block
			alignment := strings.Repeat(" ", 4)
			return doc.AlignWithString(alignment, doc.Concat{doc.Text(alignment), replaceEndOfLine(node.Value, doc.Hardline)})
		}

		// fenced code block
		style := fenceStyle(node.Value)
		lang := ""
		if node.Lang != nil {
			lang = *node.Lang
		}
		meta := ""
		if node.Meta != nil && *node.Meta != "" {
			meta = " " + *node.Meta
		}
		return doc.Concat{
			doc.Text(style),
			doc.Text(lang),
			doc.Text(meta),
			doc.Hardline,
			replaceEndOfLine(node.Value, doc.Hardline),
			doc.Hardline,
			doc.Text(style),
		}
	case "html":
		parent := parentNode(path)
		value := node.Value
		if parent.NodeType == "root" && path.IsLast() {
			value = strings.TrimRight(value, javaScriptSpaceText)
		}
		if htmlCommentPattern.MatchString(value) {
			return replaceEndOfLine(value, doc.Hardline)
		}
		return replaceEndOfLine(value, doc.MarkAsRoot(doc.Literalline))
	case "list":
		return printList(path, options, print)
	case "thematicBreak":
		ancestors := path.Ancestors()
		counter := -1
		for index, ancestor := range ancestors {
			if ancestor.NodeType == "list" {
				counter = index
				break
			}
		}
		if counter == -1 {
			return doc.Text("---")
		}
		nthSiblingIndex := getNthListSiblingIndex(ancestors[counter], ancestors[counter+1])
		if nthSiblingIndex%2 == 0 {
			return doc.Text("***")
		}
		return doc.Text("---")
	case "linkReference":
		suffix := ""
		switch node.ReferenceType {
		case "full":
			suffix = printLinkReference(node)
		case "collapsed":
			suffix = "[]"
		}
		return doc.Concat{doc.Text("["), printChildren(path, options, print, nil), doc.Text("]"), doc.Text(suffix)}
	case "imageReference":
		alt := printImageAlt(node)

		if node.ReferenceType == "full" {
			return doc.Concat{doc.Text("!["), doc.Text(alt), doc.Text("]"), doc.Text(printLinkReference(node))}
		}
		suffix := ""
		if node.ReferenceType == "collapsed" {
			suffix = "[]"
		}
		return doc.Concat{doc.Text("!"), doc.Text(printLinkReference(node)), doc.Text(suffix)}
	case "definition":
		var lineOrSpace doc.Doc = doc.Text(" ")
		if settings.proseWrap == "always" {
			lineOrSpace = doc.LineDoc
		}
		destination := doc.Text("<>")
		if node.URL != "" {
			destination = doc.Text(printURL(node.URL, ""))
		}
		var title doc.Doc = doc.Text("")
		if node.Title != nil {
			title = doc.Concat{lineOrSpace, doc.Text(printTitle(node.Title, settings, false))}
		}
		return doc.NewGroup(doc.Concat{
			doc.Text(printLinkReference(node)),
			doc.Text(":"),
			doc.NewIndent(doc.Concat{lineOrSpace, destination, title}),
		}, doc.GroupOptions{})
	case "footnoteReference":
		return doc.Text(printFootnoteReference(node))
	case "footnoteDefinition":
		shouldInlineFootnote := len(node.Children) == 1 &&
			node.Children[0].NodeType == "paragraph" &&
			(settings.proseWrap == "never" ||
				(settings.proseWrap == "preserve" &&
					node.Children[0].Position.Start.Line == node.Children[0].Position.End.Line))
		if shouldInlineFootnote {
			return doc.Concat{doc.Text(printFootnoteReference(node)), doc.Text(": "), printChildren(path, options, print, nil)}
		}
		return doc.Concat{
			doc.Text(printFootnoteReference(node)),
			doc.Text(": "),
			doc.NewGroup(doc.Concat{
				// align(" ".repeat(4), ...): a string, not a width.
				doc.AlignWithString("    ", printChildren(path, options, print, func(path *astPath) (doc.Doc, bool) {
					if path.IsFirst() {
						return doc.NewGroup(doc.Concat{doc.Softline, print(nil, nil)}, doc.GroupOptions{}), true
					}
					return print(nil, nil), true
				})),
			}, doc.GroupOptions{}),
		}
	case "table":
		return printTable(path, options, print)
	case "tableCell":
		return printChildren(path, options, print, nil)
	case "break":
		if isJavaScriptSpace(firstRuneAt(options.OriginalText, node.Position.Start.Offset)) {
			return doc.Concat{doc.Text("  "), doc.MarkAsRoot(doc.Literalline)}
		}
		return doc.Concat{doc.Text(`\`), doc.Hardline}
	case "liquidNode":
		return replaceEndOfLine(node.Value, doc.Hardline)
	case "math":
		meta := ""
		if node.Meta != nil && *node.Meta != "" {
			meta = " " + *node.Meta
		}
		var value doc.Doc = doc.Text("")
		if node.Value != "" {
			value = doc.Concat{replaceEndOfLine(node.Value, doc.Hardline), doc.Hardline}
		}
		return doc.Concat{doc.Text("$$"), doc.Text(meta), doc.Hardline, value, doc.Text("$$")}
	case "inlineMath":
		// remark-math trims content but we don't want to remove whitespaces since it's very possible
		// that it's recognized as math accidentally
		return doc.Text(options.OriginalText[locStart(node):locEnd(node)])
	case "text":
		return replaceEndOfLine(node.Value, doc.Hardline)
	default:
		// tableRow is handled in "table" and listItem in "list".
		panic("markdown: unexpected node type " + node.NodeType)
	}
}

var (
	tabsAndNewlinesPattern = regexp.MustCompile(`[\t\n]+`)
	htmlCommentPattern     = regexp.MustCompile(`(?s)^<!--.*-->$`)
	javaScriptSpaceRuns    = regexp.MustCompile(`[` + javaScriptSpace + `]+`)
)

func isJavaScriptSpace(character rune) bool {
	return character >= 0 && strings.ContainsRune(javaScriptSpaceText, character)
}

func firstRuneAt(text string, offset int) rune {
	if offset >= len(text) {
		return -1
	}
	return []rune(text[offset:min(offset+4, len(text))])[0]
}

// fenceStyle is the fence for a code block: at least three backticks, one more than the longest run in
// the value.
func fenceStyle(value string) string {
	return strings.Repeat("`", max(3, getMaxContinuousCount(value, "`")+1))
}

// replaceEndOfLine is the doc utility for a string: the lines joined with the replacement.
func replaceEndOfLine(text string, replacement doc.Doc) doc.Doc {
	lines := strings.Split(text, "\n")
	parts := make([]doc.Doc, len(lines))
	for index, line := range lines {
		parts[index] = doc.Text(line)
	}
	return doc.Join(replacement, parts)
}

func printRoot(path *astPath, options *options, print printing.PrintFunc) doc.Doc {
	type ignorePosition struct{ index, offset int }
	type ignoreRange struct{ start, end ignorePosition }
	var ignoreRanges []ignoreRange

	var ignoreStart *ignorePosition

	children := currentNode(path).Children
	for index, childNode := range children {
		switch isPrettierIgnore(childNode) {
		case "start":
			if ignoreStart == nil {
				ignoreStart = &ignorePosition{index: index, offset: childNode.Position.End.Offset}
			}
		case "end":
			if ignoreStart != nil {
				ignoreRanges = append(ignoreRanges, ignoreRange{
					start: *ignoreStart,
					end:   ignorePosition{index: index, offset: childNode.Position.Start.Offset},
				})
				ignoreStart = nil
			}
		}
	}

	return printChildren(path, options, print, func(path *astPath) (doc.Doc, bool) {
		index := pathIndex(path)
		if len(ignoreRanges) > 0 {
			ignoreRange := ignoreRanges[0]

			if index == ignoreRange.start.index {
				return doc.Concat{
					doc.Text(children[ignoreRange.start.index].Value),
					doc.Text(options.OriginalText[ignoreRange.start.offset:ignoreRange.end.offset]),
					doc.Text(children[ignoreRange.end.index].Value),
				}, true
			}

			if ignoreRange.start.index < index && index < ignoreRange.end.index {
				return nil, false
			}

			if index == ignoreRange.end.index {
				ignoreRanges = ignoreRanges[1:]
				return nil, false
			}
		}

		return print(nil, nil), true
	})
}

// shouldRemainTheSameContent: text inside a reference's label prints as written, because the label is
// matched against a definition by its normalized text.
func shouldRemainTheSameContent(path *astPath) bool {
	node, found := path.FindAncestor(func(node *Node) bool {
		return node.NodeType == "linkReference" || node.NodeType == "imageReference"
	})
	return found && (node.NodeType != "linkReference" || node.ReferenceType != "full")
}

// encodeURL is encodeUrl: each character replaced by its encodeURIComponent.
func encodeURL(link string, characters string) string {
	for _, character := range characters {
		link = strings.ReplaceAll(link, string(character), url.PathEscape(string(character)))
	}
	return link
}

// printURL is printUrl: wrap the url in `<>` when it holds a space or the dangerous character.
func printURL(link string, dangerousCharacter string) string {
	dangerous := strings.Contains(link, " ") || (dangerousCharacter != "" && strings.Contains(link, dangerousCharacter))
	if dangerous {
		return "<" + encodeURL(link, "<>") + ">"
	}
	return link
}

func printTitle(title *string, settings *settings, printSpace bool) string {
	if title == nil || *title == "" {
		return ""
	}
	if printSpace {
		return " " + printTitle(title, settings, false)
	}

	value := *title
	if strings.Contains(value, `"`) && strings.Contains(value, "'") && !strings.Contains(value, ")") {
		return "(" + value + ")" // avoid escaped quotes
	}
	quote := getPreferredQuote(value, settings.singleQuote)
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, quote, `\`+quote)
	return quote + value + quote
}

func printLinkReference(node *Node) string {
	// `remark-parse` lowercase the `label` as `identifier`, we don't want do that
	label := ""
	if node.Label != nil {
		label = javaScriptSpaceRuns.ReplaceAllString(*node.Label, " ")
	}
	var builder strings.Builder
	for _, character := range label {
		if character == '\\' || character == '[' || character == ']' {
			builder.WriteByte('\\')
		}
		builder.WriteRune(character)
	}
	return "[" + builder.String() + "]"
}

func printFootnoteReference(node *Node) string {
	label := ""
	if node.Label != nil {
		label = *node.Label
	}
	return "[^" + label + "]"
}

func printImageAlt(node *Node) string {
	if node.OriginalAltText != nil && *node.OriginalAltText != "" {
		return *node.OriginalAltText
	}
	if node.Alt != nil {
		return *node.Alt
	}
	return ""
}

// getPreferredQuote is src/utilities/get-preferred-quote.js: the preferred quote unless the text holds
// more of it than of the alternate.
func getPreferredQuote(text string, preferSingleQuote bool) string {
	preferred, alternate := `"`, "'"
	if preferSingleQuote {
		preferred, alternate = "'", `"`
	}
	if strings.Count(text, preferred) > strings.Count(text, alternate) {
		return alternate
	}
	return preferred
}

// getMaxContinuousCount is src/utilities/get-max-continuous-count.js for a one-character search.
func getMaxContinuousCount(text string, search string) int {
	maximum, current := 0, 0
	for index := 0; index < len(text); index++ {
		if text[index] == search[0] {
			current++
			maximum = max(maximum, current)
		} else {
			current = 0
		}
	}
	return maximum
}

// getMinNotPresentContinuousCount is src/utilities/get-min-not-present-continuous-count.js for a
// one-character search: the smallest run length (from 1) that does not occur.
func getMinNotPresentContinuousCount(text string, search string) int {
	present := map[int]bool{}
	maximum, current := 0, 0
	flush := func() {
		if current > 0 {
			present[current] = true
			maximum = max(maximum, current)
		}
		current = 0
	}
	for index := 0; index < len(text); index++ {
		if text[index] == search[0] {
			current++
		} else {
			flush()
		}
	}
	flush()

	if maximum == 0 {
		return 1
	}
	for count := 1; count < maximum; count++ {
		if !present[count] {
			return count
		}
	}
	return maximum + 1
}
