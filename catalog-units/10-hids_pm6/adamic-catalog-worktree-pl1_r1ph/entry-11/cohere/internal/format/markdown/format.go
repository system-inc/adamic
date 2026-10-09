// Package markdown is Prettier's markdown printer (src/language-markdown) for the markdown parser,
// ported to Go over the tree internal/format/markdown/mdast builds with the same parser the fork runs.
package markdown

import (
	"fmt"
	"sync"

	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/markdown/mdast"
	"github.com/system-inc/cohere/internal/format/printing"
)

// visitorKeys is src/language-markdown/traverse/visitor-keys.evaluate.js for the types the markdown
// parser produces. Every other type, front matter included, has none.
var visitorKeys = map[string][]string{
	"root": {"children"}, "paragraph": {"children"}, "sentence": {"children"}, "emphasis": {"children"},
	"strong": {"children"}, "delete": {"children"}, "link": {"children"}, "blockquote": {"children"},
	"heading": {"children"}, "list": {"children"}, "linkReference": {"children"}, "footnote": {"children"},
	"footnoteDefinition": {"children"}, "table": {"children"}, "tableCell": {"children"},
	"tableRow": {"children"}, "listItem": {"children"},
}

// mdastPrinter is src/language-markdown/printers.js.
var mdastPrinter = &printing.Printer[*Node]{
	Print:       printMdast,
	LocStart:    locStart,
	LocEnd:      locEnd,
	VisitorKeys: func(node *Node) []string { return visitorKeys[node.NodeType] },
	Embed:       embed,
	Preprocess: func(ast *Node, options *options) *Node {
		ast = preprocess(ast, options.OriginalText, settingsOf(options).tabWidth, settingsOf(options).nodes)
		// The print cache is made once preprocess returns, and most of the nodes it prints, the sentences,
		// words and whitespace, are only made here. The arena holds every node of the tree from both,
		// so its count, read now, presizes the cache without a walk (#4bq8vyn).
		options.NodeCount = settingsOf(options).nodes.Len()
		return ast
	},
	HasPrettierIgnore:    hasPrettierIgnore,
	PrintPrettierIgnored: printPrettierIgnored,
}

// nodeArenas hold the node memory of formats that have finished, for the next to reuse (#93dpede). A pool,
// because files are formatted on several goroutines at once and each Get is that caller's alone.
var nodeArenas = sync.Pool{New: func() any {
	return &arena.Arena[Node]{Poison: Node{NodeType: "released", Value: "released", IsLiteral: true}}
}}

// Format is Prettier's format for a markdown file: parse, print to a doc, lay the doc out.
//
// textToDoc formats the code blocks and front matter Prettier formats as another language; nil leaves
// them as written, which is also what upstream does when that formatting fails. The core's own steps
// around the printer are here too: a byte order mark is removed and restored, and line endings are
// normalized to \n before parsing, as src/main/core.js does.
func Format(text string, prettierOptions formatoptions.Options, textToDoc printing.TextToDoc) (string, error) {
	// None of our repositories sets proseWrap, and formatoptions.Options does not carry it.
	return formatWithProseWrap(text, prettierOptions, "preserve", textToDoc)
}

// formatWithProseWrap is Format with proseWrap given, so the tests can reach the always and never
// branches the corpora never take.
func formatWithProseWrap(text string, prettierOptions formatoptions.Options, proseWrap string, textToDoc printing.TextToDoc) (formatted string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			formatted, err = "", fmt.Errorf("markdown: %v", recovered)
		}
	}()

	// Byte order marks and line endings, main/core.js's part, are normalized by the caller,
	// native.Formatter, once for every printer.
	// The tree, and the words and whitespace preprocess splits its text into, come from an arena released
	// once the file is printed: doc.Print below has read the last of them before the deferred release runs.
	nodes := nodeArenas.Get().(*arena.Arena[Node])
	defer func() {
		nodes.Reset()
		nodeArenas.Put(nodes)
	}()
	ast, err := mdast.ParseMarkdown(text, nodes)
	if err != nil {
		return "", err
	}

	printOptions := &options{
		Printer:      mdastPrinter,
		OriginalText: text,
		Settings: &settings{
			proseWrap:   proseWrap,
			singleQuote: prettierOptions.SingleQuote,
			tabWidth:    prettierOptions.TabWidth,
			printWidth:  prettierOptions.PrintWidth,
			useTabs:     prettierOptions.UseTabs,
			nodes:       nodes,
		},
		EmbeddedLanguageFormatting: "auto",
		TextToDoc:                  textToDoc,
	}
	document, err := printing.PrintAstToDoc(ast, nil, printOptions)
	if err != nil {
		return "", err
	}

	return doc.Print(document, doc.Options{PrintWidth: prettierOptions.PrintWidth, TabWidth: prettierOptions.TabWidth, UseTabs: prettierOptions.UseTabs}), nil
}
