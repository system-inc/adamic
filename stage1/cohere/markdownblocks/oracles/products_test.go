package oracles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	EnableProductProduction()
	os.Exit(m.Run())
}
func productRoot(test *testing.T) string {
	test.Helper()
	root, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return root
}
func TestProduct_MarkdownblocksLeafLists(test *testing.T) {
	test.Parallel()
	LeafCompositionGoProduct(test, productRoot(test), "adamic_markdown_lists", "list_go.go")
}
func TestProduct_MarkdownblocksLeafDocLayout(test *testing.T) {
	test.Parallel()
	LeafCompositionGoProduct(test, productRoot(test), "adamic_markdown_doclayout", "document_go.go")
}
func TestProduct_MarkdownblocksQuote_lists(test *testing.T) {
	test.Parallel()
	QuoteLayoutGoProducts(test, productRoot(test), "lists")
}
func TestProduct_MarkdownblocksQuote_layout(test *testing.T) {
	test.Parallel()
	QuoteLayoutGoProducts(test, productRoot(test), "layout")
}
func TestProduct_MarkdownblocksStructure_adamic_markdown_lists(test *testing.T) {
	test.Parallel()
	StructureLayoutGoProducts(test, productRoot(test), "adamic_markdown_lists")
}
func TestProduct_MarkdownblocksStructure_adamic_markdown_doclayout(test *testing.T) {
	test.Parallel()
	StructureLayoutGoProducts(test, productRoot(test), "adamic_markdown_doclayout")
}
func TestProduct_MarkdownblocksTable_list(test *testing.T) {
	test.Parallel()
	TableLayoutGoProducts(test, productRoot(test), "list")
}
func TestProduct_MarkdownblocksTable_document(test *testing.T) {
	test.Parallel()
	TableLayoutGoProducts(test, productRoot(test), "document")
}
func TestProduct_MarkdownblocksWhitespace_lists(test *testing.T) {
	test.Parallel()
	WhitespaceLayoutGoProducts(test, test.Context(), productRoot(test), "lists")
}
func TestProduct_MarkdownblocksWhitespace_layout(test *testing.T) {
	test.Parallel()
	WhitespaceLayoutGoProducts(test, test.Context(), productRoot(test), "layout")
}

func TestProduct_MarkdownblocksTextFormatter(test *testing.T) {
	test.Parallel()
	TextFormatterProduct(test, productRoot(test))
}

func TestProduct_MarkdownblocksTextGenerator(test *testing.T) {
	test.Parallel()
	TextGeneratorProduct(test, productRoot(test))
}

func TestProduct_MarkdownChunksGo(test *testing.T) { test.Parallel(); ChunksGoOracle(test) }

func TestProduct_MarkdownASTGo(test *testing.T) { test.Parallel(); ASTGoOracle(test) }

func TestProduct_MarkdownFrontMatterGo(test *testing.T) { test.Parallel(); FrontMatterGoOracle(test) }

func TestProduct_MarkdownPrefixGo(test *testing.T) { test.Parallel(); PrefixGoOracle(test) }

func TestProduct_MarkdownIdentifierGo(test *testing.T) {
	test.Parallel()
	PilotGoProduct(test, "identifier")
}

func TestProduct_TokenizerEventsGo(test *testing.T) { test.Parallel(); PilotGoProduct(test, "events") }

func TestProduct_TokenizerEventsAnswers(test *testing.T) {
	test.Parallel()
	TokenizerEventsAnswers(test)
}

func TestProduct_MarkdownTextGo(test *testing.T) { test.Parallel(); TextGoBuild(test, "go") }

func TestProduct_MarkdownMalformedEventsGo(test *testing.T) {
	test.Parallel()
	MalformedEventsGoOracle(test)
}

func TestProduct_GrainMarkdown_path_go(test *testing.T) { test.Parallel(); GrainGoOracle(test, "path") }

func TestProduct_GrainMarkdown_mdast_go(test *testing.T) {
	test.Parallel()
	GrainGoOracle(test, "mdast")
}

func TestProduct_GrainMarkdown_decode_go(test *testing.T) {
	test.Parallel()
	GrainGoOracle(test, "decode")
}

func TestProduct_GrainMarkdown_width_go(test *testing.T) {
	test.Parallel()
	GrainGoOracle(test, "width")
}

func TestProduct_GrainMarkdownEntitiesGenerator(test *testing.T) {
	test.Parallel()
	GrainTool(test, "entities")
}

func TestProduct_GrainMarkdownWidthGenerator(test *testing.T) {
	test.Parallel()
	GrainTool(test, "width")
}

func TestProduct_GrainMarkdownFormatter(test *testing.T) {
	test.Parallel()
	GrainTool(test, "formatter")
}

func TestProduct_GrainMarkdownWidthFullAnswers(test *testing.T) {
	test.Parallel()
	GrainMarkdownWidthFullAnswers(test)
}

func TestProduct_WholeDocumentPreflightGoProducts(test *testing.T) {
	test.Parallel()
	PreflightGoProducts(test)
}
