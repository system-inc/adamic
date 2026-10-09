// Package yaml is Prettier's YAML printer (src/language-yaml) ported to Go, printing the tree
// internal/format/yaml/unist describes: yaml-unist-parser's, over eemeli/yaml.
//
// The printer prints comments itself. Prettier's parser deletes root.comments so the core never
// attaches or prints any, and so the core's comment hooks stay nil here and PrintAstToDoc is given none.
package yaml

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/printing"
	"github.com/system-inc/cohere/internal/format/yaml/unist"
)

// src/language-yaml/printer-yaml.js.
//
// massageAstNode (the AST comparison for Prettier's own tests) and insertPragma (we never insert a
// pragma) are not ported.

type (
	unistNode = unist.Node
	astPath   = printing.AstPath[*unist.Node]
	options   = printing.Options[*unist.Node]
)

// settings are the Prettier options the YAML printer reads, carried on options.Settings, and the
// per-print state upstream keeps elsewhere.
type settings struct {
	bracketSpacing bool
	singleQuote    bool
	proseWrap      string
	tabWidth       int
	trailingComma  string

	// filePath is upstream's options.filepath, which only the .prettierrc embed reads. Empty is
	// upstream's undefined.
	filePath string

	// isNextEmptyLinePrinted is printNextEmptyLine's printedEmptyLineCache entry for this print's root.
	isNextEmptyLinePrinted map[int]bool
}

func settingsOf(options *options) *settings { return options.Settings.(*settings) }

func currentNode(path *astPath) *unist.Node {
	node, _ := path.Node()
	return node
}

func parentNode(path *astPath) *unist.Node {
	node, _ := path.Parent()
	return node
}

func nextNode(path *astPath) *unist.Node {
	node, _ := path.Next()
	return node
}

// mapPrint is upstream's path.map(print, name): print receives each element's index as its args, as
// upstream's map passes (path, index, value) to print, so these prints are not cached.
func mapPrint(path *astPath, print printing.PrintFunc, names ...any) []doc.Doc {
	return printing.Map(path, func(_ *astPath, index int, _ any) doc.Doc { return print(nil, index) }, names...)
}

// fillWords is fill(join(line, words)).
func fillWords(words []string) doc.Doc {
	parts := make([]doc.Doc, 0, 2*len(words))
	for index, word := range words {
		if index > 0 {
			parts = append(parts, doc.LineDoc)
		}
		parts = append(parts, doc.Text(word))
	}
	return doc.NewFill(parts)
}

// yamlPrinter is the printer object printer-yaml.js exports.
var yamlPrinter = &printing.Printer[*unist.Node]{
	Print:            genericPrint,
	LocStart:         locStart,
	LocEnd:           locEnd,
	VisitorKeys:      getVisitorKeys,
	Embed:            embed,
	EmbedVisitorKeys: embedVisitorKeys,
	Preprocess:       preprocess,
}

// Print formats a parsed YAML file: text is the source the tree was parsed from (the byte order mark
// removed and line endings normalized to \n, as Prettier's core does before parsing), and the tree's
// positions are byte offsets into it. The result ends in options.EndOfLine; restoring a byte order mark
// is the caller's, as it is the core's upstream.
//
// No file name is given, so the .prettierrc embed never applies, as upstream's without
// options.filepath. PrintFile takes one.
func Print(root *unist.Node, text string, options formatoptions.Options, textToDoc printing.TextToDoc) (string, error) {
	return PrintFile("", root, text, options, textToDoc)
}

// PrintDoc is Print's doc, before layout: what markdown's front matter embeds. The trailing hardline is
// still there; upstream's textToDoc strips it with stripTrailingHardline (doc.StripTrailingHardline),
// and the caller is expected to do the same.
func PrintDoc(root *unist.Node, text string, options formatoptions.Options, textToDoc printing.TextToDoc) (doc.Doc, error) {
	return PrintDocFile("", root, text, options, textToDoc)
}

// PrintFile is Print for a named file. The name is upstream's options.filepath: a .prettierrc,
// .stylelintrc or .lintstagedrc is printed as JSON when textToDoc can format it as JSON.
func PrintFile(fileName string, root *unist.Node, text string, options formatoptions.Options, textToDoc printing.TextToDoc) (string, error) {
	return printFile(fileName, root, text, options, "preserve", textToDoc, 0)
}

// PrintDocFile is PrintDoc for a named file.
func PrintDocFile(fileName string, root *unist.Node, text string, options formatoptions.Options, textToDoc printing.TextToDoc) (doc.Doc, error) {
	return printDocFile(fileName, root, text, options, "preserve", textToDoc, 0)
}

// printFile is PrintFile with proseWrap given. None of our repositories sets it, and formatoptions.Options
// does not carry it, so only the tests reach always and never. nodeCount is how many nodes the tree holds,
// zero when unknown, for printing.Options.NodeCount.
func printFile(fileName string, root *unist.Node, text string, prettierOptions formatoptions.Options, proseWrap string, textToDoc printing.TextToDoc, nodeCount int) (string, error) {
	// src/main/core.js, coreFormat: a file that is empty or only whitespace formats to "" without being
	// printed. The doc entry has no such check, as upstream's textToDoc has none.
	if trim(text) == "" {
		return "", nil
	}

	document, err := printDocFile(fileName, root, text, prettierOptions, proseWrap, textToDoc, nodeCount)
	if err != nil {
		return "", err
	}
	formatted := doc.Print(document, doc.Options{
		PrintWidth: prettierOptions.PrintWidth,
		TabWidth:   prettierOptions.TabWidth,
		UseTabs:    prettierOptions.UseTabs,
	})
	// printDocToString writes options.endOfLine for every newline, in text and line breaks alike.
	switch prettierOptions.EndOfLine {
	case "crlf":
		formatted = strings.ReplaceAll(formatted, "\n", "\r\n")
	case "cr":
		formatted = strings.ReplaceAll(formatted, "\n", "\r")
	}
	return formatted, nil
}

func printDocFile(fileName string, root *unist.Node, text string, prettierOptions formatoptions.Options, proseWrap string, textToDoc printing.TextToDoc, nodeCount int) (document doc.Doc, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			document, err = nil, fmt.Errorf("yaml: %v", recovered)
		}
	}()

	printOptions := &options{
		Printer:      yamlPrinter,
		OriginalText: text,
		Settings: &settings{
			bracketSpacing:         prettierOptions.BracketSpacing,
			singleQuote:            prettierOptions.SingleQuote,
			proseWrap:              proseWrap,
			tabWidth:               prettierOptions.TabWidth,
			trailingComma:          prettierOptions.TrailingComma,
			filePath:               fileName,
			isNextEmptyLinePrinted: map[int]bool{},
		},
		EmbeddedLanguageFormatting: "auto",
		TextToDoc:                  textToDoc,
		NodeCount:                  nodeCount,
	}
	// No comments: parser-yaml.js deletes root.comments, because the printer prints them itself.
	return printing.PrintAstToDoc(root, nil, printOptions)
}

// genericPrint is upstream's genericPrint: a node's comments, tag and anchor around printNode.
func genericPrint(path *astPath, options *options, print printing.PrintFunc, _ any) doc.Doc {
	node := currentNode(path)
	settings := settingsOf(options)

	// Room for the usual parts (the node's group and the empty line after it, plus a comment or a
	// property), made once rather than grown from nothing (#vbjv3d6).
	parts := make(doc.Concat, 0, 4)

	if node.NodeType != "mappingValue" && hasLeadingComments(node) {
		parts = append(parts, doc.Concat{doc.Join(doc.Hardline, mapPrint(path, print, "leadingComments")), doc.Hardline})
	}

	tag, anchor := node.Tag, node.Anchor
	if tag != nil {
		parts = append(parts, print("tag", nil))
	}
	if tag != nil && anchor != nil {
		parts = append(parts, doc.Text(" "))
	}
	if anchor != nil {
		parts = append(parts, print("anchor", nil))
	}

	var nextEmptyLine doc.Doc = doc.Text("")

	if isNode(node, "mapping", "sequence", "comment", "directive", "mappingItem", "sequenceItem") &&
		!isLastDescendantNode(path) {
		nextEmptyLine = printNextEmptyLine(path, options.OriginalText, settings)
	}

	if tag != nil || anchor != nil {
		if isNode(node, "sequence", "mapping") && !hasMiddleComments(node) {
			parts = append(parts, doc.Hardline)
		} else {
			parts = append(parts, doc.Text(" "))
		}
	}

	if hasMiddleComments(node) {
		var first doc.Doc = doc.Hardline
		if len(node.MiddleComments) == 1 {
			first = doc.Text("")
		}
		parts = append(parts, doc.Concat{
			first,
			doc.Join(doc.Hardline, mapPrint(path, print, "middleComments")),
			doc.Hardline,
		})
	}

	if hasPrettierIgnore(path) {
		parts = append(parts, doc.ReplaceEndOfLine(
			doc.Text(trimEnd(options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset])),
			nil,
		))
	} else {
		parts = append(parts, doc.NewGroup(printNode(path, options, print), doc.GroupOptions{}))
	}

	if hasTrailingComment(node) && !isNode(node, "document", "documentHead") {
		var space doc.Doc = doc.Text(" ")
		if node.NodeType == "mappingValue" && contentOf(node) == nil {
			space = doc.Text("")
		}
		var breakParent = doc.BreakParent
		if parent := parentNode(path); parent.NodeType == "mappingKey" {
			// path.getParentNode(2): the mapping item's parent.
			if greatGrandparent, _ := path.GetParentNode(2); greatGrandparent.NodeType == "mapping" && isInlineNode(node) {
				breakParent = doc.Text("")
			}
		}
		parts = append(parts, doc.NewLineSuffix(doc.Concat{space, breakParent, print("trailingComment", nil)}))
	}

	if shouldPrintEndComments(node) {
		width := 0
		if node.NodeType == "sequenceItem" {
			width = 2
		}
		endComments := printing.Map(path, func(path *astPath, _ int, _ any) doc.Doc {
			var emptyLine doc.Doc = doc.Text("")
			if printing.IsPreviousLineEmpty(options.OriginalText, locStart(currentNode(path))) {
				emptyLine = doc.Hardline
			}
			return doc.Concat{emptyLine, print(nil, nil)}
		}, "endComments")
		parts = append(parts, alignWithSpaces(width, doc.Concat{doc.Hardline, doc.Join(doc.Hardline, endComments)}))
	}
	parts = append(parts, nextEmptyLine)
	return parts
}

// printNode is upstream's printNode: the node itself, by type.
func printNode(path *astPath, options *options, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	settings := settingsOf(options)
	switch node.NodeType {
	case "root":
		lastDescendantNode := getLastDescendantNode(node)
		shouldPrintHardline := !(isNode(lastDescendantNode, "blockLiteral", "blockFolded") &&
			lastDescendantNode.Chomping == "keep")
		parts := doc.Concat{}
		path.Each(func(path *astPath, _ int, _ any) {
			document := currentNode(path)
			if !path.IsFirst() {
				parts = append(parts, doc.Hardline)
			}
			parts = append(parts, print(nil, nil))
			if shouldPrintDocumentEndMarker(path) {
				if shouldPrintHardline {
					parts = append(parts, doc.Hardline)
				}

				parts = append(parts, doc.Text("..."))
				if hasTrailingComment(document) {
					parts = append(parts, doc.Text(" "), print("trailingComment", nil))
				}
			}
		}, "children")

		if shouldPrintHardline {
			parts = append(parts, doc.Hardline)
		}

		return parts
	case "document":
		parts := []doc.Doc{}
		if shouldPrintDocumentHeadEndMarker(path) {
			head := childAt(node, 0)
			if len(head.Children) > 0 || len(head.EndComments) > 0 {
				parts = append(parts, print("head", nil))
			}

			if hasTrailingComment(head) {
				parts = append(parts, doc.Concat{doc.Text("---"), doc.Text(" "), print([]string{"head", "trailingComment"}, nil)})
			} else {
				parts = append(parts, doc.Text("---"))
			}
		}

		if shouldPrintDocumentBody(node) {
			parts = append(parts, print("body", nil))
		}

		return doc.Join(doc.Hardline, parts)
	case "documentHead":
		children := mapPrint(path, print, "children")
		endComments := mapPrint(path, print, "endComments")
		return doc.Join(doc.Hardline, append(children, endComments...))
	case "documentBody":
		children, endComments := node.Children, node.EndComments
		var separator doc.Doc = doc.Text("")
		if len(children) > 0 && len(endComments) > 0 {
			lastDescendantNode := getLastDescendantNode(node)
			// there's already a newline printed at the end of blockValue (chomping=keep, lastDescendant=true)
			if isNode(lastDescendantNode, "blockFolded", "blockLiteral") {
				// an extra newline for better readability
				if lastDescendantNode.Chomping != "keep" {
					separator = doc.Concat{doc.Hardline, doc.Hardline}
				}
			} else {
				// Check if we need to preserve empty line before end comments
				lastChild := children[len(children)-1]
				shouldPreserveEmptyLine := isNode(lastChild, "mapping") &&
					printing.IsPreviousLineEmpty(options.OriginalText, locStart(endComments[0]))

				if shouldPreserveEmptyLine {
					separator = doc.Concat{doc.Hardline, doc.Hardline}
				} else {
					separator = doc.Hardline
				}
			}
		}

		printedChildren := doc.Join(doc.Hardline, mapPrint(path, print, "children"))
		return doc.Concat{
			printedChildren,
			separator,
			doc.Join(doc.Hardline, mapPrint(path, print, "endComments")),
		}
	case "directive":
		words := make([]doc.Doc, 0, 1+len(node.Parameters))
		words = append(words, doc.Text(node.Name))
		for _, parameter := range node.Parameters {
			words = append(words, doc.Text(parameter))
		}
		return doc.Concat{doc.Text("%"), doc.Join(doc.Text(" "), words)}
	case "comment":
		return doc.Concat{doc.Text("#"), doc.Text(node.Value)}
	case "alias":
		return doc.Concat{doc.Text("*"), doc.Text(node.Value)}
	case "tag":
		return doc.Text(options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset])
	case "anchor":
		return doc.Concat{doc.Text("&"), doc.Text(node.Value)}
	case "plain":
		return printFlowScalarContent(
			node.NodeType,
			options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset],
			settings,
		)
	case "quoteDouble", "quoteSingle":
		const singleQuote = "'"
		const doubleQuote = "\""

		raw := options.OriginalText[node.Position.Start.Offset+1 : node.Position.End.Offset-1]

		if (node.NodeType == "quoteSingle" && strings.Contains(raw, "\\")) ||
			(node.NodeType == "quoteDouble" && hasEscapeOtherThanDoubleQuote(raw)) {
			// only quoteDouble can use escape chars
			// and quoteSingle do not need to escape backslashes
			originalQuote := singleQuote
			if node.NodeType == "quoteDouble" {
				originalQuote = doubleQuote
			}
			return doc.Concat{
				doc.Text(originalQuote),
				printFlowScalarContent(node.NodeType, raw, settings),
				doc.Text(originalQuote),
			}
		}

		if strings.Contains(raw, doubleQuote) {
			content := raw
			if node.NodeType == "quoteDouble" {
				// double quote needs to be escaped by backslash in quoteDouble
				content = strings.ReplaceAll(strings.ReplaceAll(raw, `\"`, doubleQuote), "'", strings.Repeat(singleQuote, 2))
			}
			return doc.Concat{
				doc.Text(singleQuote),
				printFlowScalarContent(node.NodeType, content, settings),
				doc.Text(singleQuote),
			}
		}

		if strings.Contains(raw, singleQuote) {
			content := raw
			if node.NodeType == "quoteSingle" {
				// single quote needs to be escaped by 2 single quotes in quoteSingle
				content = strings.ReplaceAll(raw, "''", singleQuote)
			}
			return doc.Concat{
				doc.Text(doubleQuote),
				printFlowScalarContent(node.NodeType, content, settings),
				doc.Text(doubleQuote),
			}
		}

		quote := doubleQuote
		if settings.singleQuote {
			quote = singleQuote
		}
		return doc.Concat{doc.Text(quote), printFlowScalarContent(node.NodeType, raw, settings), doc.Text(quote)}
	case "blockFolded", "blockLiteral":
		return printBlock(path, options, print)

	case "mapping", "sequence":
		return doc.Join(doc.Hardline, mapPrint(path, print, "children"))
	case "sequenceItem":
		var content doc.Doc = doc.Text("")
		if contentOf(node) != nil {
			content = print("content", nil)
		}
		return doc.Concat{doc.Text("- "), alignWithSpaces(2, content)}
	case "mappingKey", "mappingValue":
		if contentOf(node) == nil {
			return doc.Text("")
		}
		return print("content", nil)
	case "mappingItem", "flowMappingItem":
		return printMappingItem(path, options, print)

	case "flowMapping", "flowSequence":
		return printFlowMapping(path, options, print)
	case "flowSequenceItem":
		return print("content", nil)
	default:
		panic(fmt.Sprintf("Unexpected node type %q in YAML", node.NodeType))
	}
}

// hasEscapeOtherThanDoubleQuote is /\\[^"]/.test(raw): a backslash followed by any character but a
// double quote, a newline included.
func hasEscapeOtherThanDoubleQuote(raw string) bool {
	for index := 0; index+1 < len(raw); index++ {
		if raw[index] == '\\' && raw[index+1] != '"' {
			return true
		}
	}
	return false
}

// shouldPrintDocumentBody is upstream's shouldPrintDocumentBody.
func shouldPrintDocumentBody(document *unist.Node) bool {
	body := childAt(document, 1)
	return len(body.Children) > 0 || hasEndComments(body)
}

// shouldPrintDocumentEndMarker is upstream's shouldPrintDocumentEndMarker.
func shouldPrintDocumentEndMarker(path *astPath) bool {
	document := currentNode(path)

	if document.DocumentEndMarker {
		return true
	}

	/**
	 *... # trailingComment
	 */
	if hasTrailingComment(document) {
		return true
	}

	if path.IsLast() {
		return false
	}

	nextDocument := nextNode(path)
	nextHead := childAt(nextDocument, 0)

	return len(nextHead.Children) > 0 || // ... %DIRECTIVE ---
		hasEndComments(nextHead) // ... # endComment ---
}

// shouldPrintDocumentHeadEndMarker is upstream's shouldPrintDocumentHeadEndMarker.
func shouldPrintDocumentHeadEndMarker(path *astPath) bool {
	document := currentNode(path)
	head := childAt(document, 0)
	return document.DirectivesEndMarker || // --- preserve the first document head end marker
		len(head.Children) > 0 || // %DIRECTIVE ---
		hasEndComments(head) || // # end comment ---
		hasTrailingComment(head) // --- # trailing comment
}

// printFlowScalarContent is upstream's printFlowScalarContent.
func printFlowScalarContent(nodeType string, content string, settings *settings) doc.Doc {
	if settings.proseWrap == "preserve" {
		return printPreservedFlowScalarContent(content)
	}
	lineContents := getFlowScalarLineContents(nodeType, content, settings)
	lines := make([]doc.Doc, len(lineContents))
	for index, lineContentWords := range lineContents {
		lines[index] = fillWords(lineContentWords)
	}
	return doc.Join(doc.Hardline, lines)
}

// printPreservedFlowScalarContent is printFlowScalarContent under proseWrap preserve, which every one of our
// repositories uses, built in one pass (#v6ksqg3). The doc is the one the general path builds:
// getFlowScalarLineContents splits the content on "\n", trims each line as it does (the first only at its
// end, the last only at its start, a lone line not at all), and makes each line one word, or none when it
// is empty; fillWords makes each a fill; doc.Join puts hardlines between them. Here the lines are cut and
// trimmed in place and the joined concat is built directly, without the split, the trimmed copy and the
// per-line word lists between.
func printPreservedFlowScalarContent(content string) doc.Doc {
	count := strings.Count(content, "\n") + 1
	last := count - 1
	joined := make(doc.Concat, 0, 2*count)
	rest := content
	for index := 0; index < count; index++ {
		line, after, _ := strings.Cut(rest, "\n")
		rest = after
		switch {
		case index == 0 && index == last:
		case index != 0 && index != last:
			line = trim(line)
		case index == 0:
			line = trimEnd(line)
		default:
			line = trimStart(line)
		}
		if index > 0 {
			joined = append(joined, doc.Hardline)
		}
		// fillWords of one word, or of none: the same parts, with the same room.
		if line == "" {
			joined = append(joined, doc.NewFill(make([]doc.Doc, 0)))
		} else {
			joined = append(joined, doc.NewFill(append(make([]doc.Doc, 0, 2), doc.Text(line))))
		}
	}
	return joined
}

// The remaining files of src/language-yaml, small enough to share this one.

// locStart and locEnd are loc.js.
func locStart(node *unist.Node) int { return node.Position.Start.Offset }
func locEnd(node *unist.Node) int   { return node.Position.End.Offset }

// preprocess is print-preprocess.js: mapNode(ast, defineShortcuts).
//
// Upstream copies every node that has children ({...node, children}) and defines the shortcut getters
// on the copies. unist.Node.Field already answers the shortcuts, and no copy is needed: nothing the
// printer reads distinguishes a copy from its original. The copies drop _parent, which the printer never
// reads, and the printedEmptyLineCache upstream keys on the copied root is this print's own here.
func preprocess(ast *unist.Node, _ *options) *unist.Node { return ast }

// prettierRcPattern is /(?:[/\\]|^)\.(?:prettier|stylelint|lintstaged)rc$/.
var prettierRcPattern = regexp.MustCompile(`(?:[/\\]|^)\.(?:prettier|stylelint|lintstaged)rc$`)

// embedPrint is what an embed returns: the function that prints the node as another language.
type embedPrint = func(printing.TextToDoc, printing.PrintFunc, *astPath, *options) (doc.Doc, error)

// embed is embed.js: try to format `.prettierrc`, `.stylelintrc` and `.lintstagedrc` as JSON first.
// When that fails, the core drops the embed and the file prints as YAML.
func embed(path *astPath, options *options) embedPrint {
	node := currentNode(path)
	settings := settingsOf(options)

	if node.NodeType == "root" && settings.filePath != "" && prettierRcPattern.MatchString(settings.filePath) {
		return func(textToDoc printing.TextToDoc, _ printing.PrintFunc, _ *astPath, _ *printing.Options[*unist.Node]) (doc.Doc, error) {
			embedded, err := textToDoc(options.OriginalText, "json")
			if err != nil {
				return nil, err
			}
			// doc ? [doc, hardline] : undefined
			if text, isText := embedded.(doc.Text); embedded == nil || (isText && text == "") {
				return nil, nil
			}
			return doc.Concat{embedded, doc.Hardline}, nil
		}
	}
	return nil
}

// embedVisitorKeys is embed.getVisitorKeys: only the root may print as JSON, and [] keeps
// printEmbeddedLanguages from walking deeper.
func embedVisitorKeys(*unist.Node) []string { return []string{} }
