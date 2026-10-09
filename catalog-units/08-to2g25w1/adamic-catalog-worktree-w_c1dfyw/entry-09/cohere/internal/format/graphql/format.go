package graphql

// The printer object of src/language-graphql/printer-graphql.js, loc.js, and the two entries the rest of
// cohere calls: Format for a .graphql or .gql file, and PrintToDoc for GraphQL embedded in a TypeScript
// template, which is upstream's textToDoc (src/main/multiparser.js).

import (
	"fmt"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/printing"
)

// astPath is upstream's AstPath over GraphQL nodes.
type astPath = printing.AstPath[*estree.Node]

// printerOptions is upstream's options object for one format.
type printerOptions = printing.Options[*estree.Node]

// locStart and locEnd are loc.js: a node's or token's loc.start and loc.end, which are the Range here.
func locStart(nodeOrToken *estree.Node) int { return nodeOrToken.Range[0] }
func locEnd(nodeOrToken *estree.Node) int   { return nodeOrToken.Range[1] }

// graphqlPrinter is printer-graphql.js's printer object. massageAstNode is test-only upstream and
// insertPragma runs only under insertPragma, which cohere never sets, so neither is here.
var graphqlPrinter = &printing.Printer[*estree.Node]{
	Print:             genericPrint,
	LocStart:          locStart,
	LocEnd:            locEnd,
	VisitorKeys:       getVisitorKeys,
	HasPrettierIgnore: hasPrettierIgnore,
	PrintComment:      printComment,
	CanAttachComment:  canAttachComment,
}

// settingsOf is the resolved Prettier options a format carries in options.Settings. The printer reads
// one of them, bracketSpacing (options.js).
func settingsOf(options *printerOptions) formatoptions.Options {
	return options.Settings.(formatoptions.Options)
}

// Format is Prettier's format for a GraphQL file: parse, attach comments, print to a doc, lay it out.
// Byte order marks and line endings are normalized by the caller, native.Formatter, for every printer.
func Format(text string, prettierOptions formatoptions.Options) (formatted string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			formatted, err = "", fmt.Errorf("graphql: %v", recovered)
		}
	}()

	document, err := printToDoc(text, prettierOptions)
	if err != nil {
		return "", err
	}
	return doc.Print(document, doc.Options{
		PrintWidth:     prettierOptions.PrintWidth,
		TabWidth:       prettierOptions.TabWidth,
		UseTabs:        prettierOptions.UseTabs,
		ExpectedLength: len(text),
	}), nil
}

// PrintToDoc is upstream's textToDoc for GraphQL embedded in another language: the printed doc with its
// trailing hardline stripped.
func PrintToDoc(text string, prettierOptions formatoptions.Options) (document doc.Doc, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			document, err = nil, fmt.Errorf("graphql: %v", recovered)
		}
	}()

	document, err = printToDoc(text, prettierOptions)
	if err != nil {
		return nil, err
	}
	return doc.StripTrailingHardline(document), nil
}

// printToDoc is parser-graphql.js's parse followed by printAstToDoc (src/main/ast-to-doc.js).
func printToDoc(text string, prettierOptions formatoptions.Options) (doc.Doc, error) {
	document, comments, err := Parse(text)
	if err != nil {
		return nil, err
	}
	printOptions := &printerOptions{
		Printer:                    graphqlPrinter,
		OriginalText:               text,
		Settings:                   prettierOptions,
		EmbeddedLanguageFormatting: "auto",
	}
	return printing.PrintAstToDoc(document, comments, printOptions)
}
