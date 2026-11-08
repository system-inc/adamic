package markdownblocks

import (
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

func TestProduct_MarkdownblocksFindsLeafLists(test *testing.T) {
	test.Parallel()
	oracles.LeafCompositionGoProduct(test, markdownLayoutProductRoot(test), "adamic_markdown_lists", "list_go.go")
}
func TestProduct_MarkdownblocksFindsLeafDocLayout(test *testing.T) {
	test.Parallel()
	oracles.LeafCompositionGoProduct(
		test,
		markdownLayoutProductRoot(test),
		"adamic_markdown_doclayout",
		"document_go.go",
	)
}
func TestProduct_MarkdownblocksFindsQuote_lists(test *testing.T) {
	test.Parallel()
	oracles.QuoteLayoutGoProducts(test, markdownLayoutProductRoot(test), "lists")
}
func TestProduct_MarkdownblocksFindsQuote_layout(test *testing.T) {
	test.Parallel()
	oracles.QuoteLayoutGoProducts(test, markdownLayoutProductRoot(test), "layout")
}
func TestProduct_MarkdownblocksFindsStructure_adamic_markdown_lists(test *testing.T) {
	test.Parallel()
	oracles.StructureLayoutGoProducts(test, markdownLayoutProductRoot(test), "adamic_markdown_lists")
}
func TestProduct_MarkdownblocksFindsStructure_adamic_markdown_doclayout(test *testing.T) {
	test.Parallel()
	oracles.StructureLayoutGoProducts(test, markdownLayoutProductRoot(test), "adamic_markdown_doclayout")
}
func TestProduct_MarkdownblocksFindsTable_list(test *testing.T) {
	test.Parallel()
	oracles.TableLayoutGoProducts(test, markdownLayoutProductRoot(test), "list")
}
func TestProduct_MarkdownblocksFindsTable_document(test *testing.T) {
	test.Parallel()
	oracles.TableLayoutGoProducts(test, markdownLayoutProductRoot(test), "document")
}
func TestProduct_MarkdownblocksFindsWhitespace_lists(test *testing.T) {
	test.Parallel()
	oracles.WhitespaceLayoutGoProducts(test, test.Context(), markdownLayoutProductRoot(test), "lists")
}
func TestProduct_MarkdownblocksFindsWhitespace_layout(test *testing.T) {
	test.Parallel()
	oracles.WhitespaceLayoutGoProducts(test, test.Context(), markdownLayoutProductRoot(test), "layout")
}

func TestProduct_MarkdownblocksFindsTextFormatter(test *testing.T) {
	test.Parallel()
	oracles.TextFormatterProduct(test, markdownLayoutProductRoot(test))
}

func TestProduct_MarkdownblocksFindsTextGenerator(test *testing.T) {
	test.Parallel()
	oracles.TextGeneratorProduct(test, markdownLayoutProductRoot(test))
}

func TestProduct_MarkdownblocksFindsMarkdownChunksGo(test *testing.T) {
	test.Parallel()
	oracles.ChunksGoOracle(test)
}

func TestProduct_MarkdownblocksFindsMarkdownAstGo(test *testing.T) {
	test.Parallel()
	oracles.ASTGoOracle(test)
}

func TestProduct_MarkdownblocksFindsMarkdownFrontMatterGo(test *testing.T) {
	test.Parallel()
	oracles.FrontMatterGoOracle(test)
}

func TestProduct_MarkdownblocksFindsMarkdownPrefixGo(test *testing.T) {
	test.Parallel()
	oracles.PrefixGoOracle(test)
}

func TestProduct_MarkdownblocksFindsMarkdownIdentifierGo(test *testing.T) {
	test.Parallel()
	oracles.PilotGoProduct(test, "identifier")
}

func TestProduct_MarkdownblocksFindsTokenizerEventsGo(test *testing.T) {
	test.Parallel()
	oracles.PilotGoProduct(test, "events")
}

func TestProduct_MarkdownblocksFindsTokenizerEventsAnswers(test *testing.T) {
	test.Parallel()
	oracles.TokenizerEventsAnswers(test)
}

func TestProduct_MarkdownblocksFindsMarkdownTextGo(test *testing.T) {
	test.Parallel()
	oracles.TextGoBuild(test, "go")
}

func TestProduct_MarkdownblocksFindsMarkdownMalformedEventsGo(test *testing.T) {
	test.Parallel()
	oracles.MalformedEventsGoOracle(test)
}

func TestProduct_MarkdownblocksFindsGrainMarkdown_path_go(test *testing.T) {
	test.Parallel()
	oracles.GrainGoOracle(test, "path")
}

func TestProduct_MarkdownblocksFindsGrainMarkdown_mdast_go(test *testing.T) {
	test.Parallel()
	oracles.GrainGoOracle(test, "mdast")
}

func TestProduct_MarkdownblocksFindsGrainMarkdown_decode_go(test *testing.T) {
	test.Parallel()
	oracles.GrainGoOracle(test, "decode")
}

func TestProduct_MarkdownblocksFindsGrainMarkdown_width_go(test *testing.T) {
	test.Parallel()
	oracles.GrainGoOracle(test, "width")
}

func TestProduct_MarkdownblocksFindsGrainMarkdownEntitiesGenerator(test *testing.T) {
	test.Parallel()
	oracles.GrainTool(test, "entities")
}

func TestProduct_MarkdownblocksFindsGrainMarkdownWidthGenerator(test *testing.T) {
	test.Parallel()
	oracles.GrainTool(test, "width")
}

func TestProduct_MarkdownblocksFindsGrainMarkdownFormatter(test *testing.T) {
	test.Parallel()
	oracles.GrainTool(test, "formatter")
}

func TestProduct_MarkdownblocksFindsGrainMarkdownWidthFullAnswers(test *testing.T) {
	test.Parallel()
	oracles.GrainMarkdownWidthFullAnswers(test)
}

func TestProduct_MarkdownblocksFindsWholeDocumentPreflightGoProducts(test *testing.T) {
	test.Parallel()
	oracles.PreflightGoProducts(test)
}
