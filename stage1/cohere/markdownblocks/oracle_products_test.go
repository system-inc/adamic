package markdownblocks

import (
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

func TestProduct_MarkdownblocksFindsLeafLists(t *testing.T) {
	t.Parallel()
	oracles.LeafCompositionGoProduct(t, markdownLayoutProductRoot(t), "adamic_markdown_lists", "list_go.go")
}
func TestProduct_MarkdownblocksFindsLeafDocLayout(t *testing.T) {
	t.Parallel()
	oracles.LeafCompositionGoProduct(t, markdownLayoutProductRoot(t), "adamic_markdown_doclayout", "document_go.go")
}
func TestProduct_MarkdownblocksFindsQuote_lists(t *testing.T) {
	t.Parallel()
	oracles.QuoteLayoutGoProducts(t, markdownLayoutProductRoot(t), "lists")
}
func TestProduct_MarkdownblocksFindsQuote_layout(t *testing.T) {
	t.Parallel()
	oracles.QuoteLayoutGoProducts(t, markdownLayoutProductRoot(t), "layout")
}
func TestProduct_MarkdownblocksFindsStructure_adamic_markdown_lists(t *testing.T) {
	t.Parallel()
	oracles.StructureLayoutGoProducts(t, markdownLayoutProductRoot(t), "adamic_markdown_lists")
}
func TestProduct_MarkdownblocksFindsStructure_adamic_markdown_doclayout(t *testing.T) {
	t.Parallel()
	oracles.StructureLayoutGoProducts(t, markdownLayoutProductRoot(t), "adamic_markdown_doclayout")
}
func TestProduct_MarkdownblocksFindsTable_list(t *testing.T) {
	t.Parallel()
	oracles.TableLayoutGoProducts(t, markdownLayoutProductRoot(t), "list")
}
func TestProduct_MarkdownblocksFindsTable_document(t *testing.T) {
	t.Parallel()
	oracles.TableLayoutGoProducts(t, markdownLayoutProductRoot(t), "document")
}
func TestProduct_MarkdownblocksFindsWhitespace_lists(t *testing.T) {
	t.Parallel()
	oracles.WhitespaceLayoutGoProducts(t, t.Context(), markdownLayoutProductRoot(t), "lists")
}
func TestProduct_MarkdownblocksFindsWhitespace_layout(t *testing.T) {
	t.Parallel()
	oracles.WhitespaceLayoutGoProducts(t, t.Context(), markdownLayoutProductRoot(t), "layout")
}

func TestProduct_MarkdownblocksFindsTextFormatter(t *testing.T) {
	t.Parallel()
	oracles.TextFormatterProduct(t, markdownLayoutProductRoot(t))
}

func TestProduct_MarkdownblocksFindsTextGenerator(t *testing.T) {
	t.Parallel()
	oracles.TextGeneratorProduct(t, markdownLayoutProductRoot(t))
}

func TestProduct_MarkdownblocksFindsMarkdownChunksGo(t *testing.T) {
	t.Parallel()
	oracles.ChunksGoOracle(t)
}

func TestProduct_MarkdownblocksFindsMarkdownASTGo(t *testing.T) { t.Parallel(); oracles.ASTGoOracle(t) }

func TestProduct_MarkdownblocksFindsMarkdownFrontMatterGo(t *testing.T) {
	t.Parallel()
	oracles.FrontMatterGoOracle(t)
}

func TestProduct_MarkdownblocksFindsMarkdownPrefixGo(t *testing.T) {
	t.Parallel()
	oracles.PrefixGoOracle(t)
}

func TestProduct_MarkdownblocksFindsMarkdownIdentifierGo(t *testing.T) {
	t.Parallel()
	oracles.PilotGoProduct(t, "identifier")
}

func TestProduct_MarkdownblocksFindsTokenizerEventsGo(t *testing.T) {
	t.Parallel()
	oracles.PilotGoProduct(t, "events")
}

func TestProduct_MarkdownblocksFindsTokenizerEventsAnswers(t *testing.T) {
	t.Parallel()
	oracles.TokenizerEventsAnswers(t)
}

func TestProduct_MarkdownblocksFindsMarkdownTextGo(t *testing.T) {
	t.Parallel()
	oracles.TextGoBuild(t, "go")
}

func TestProduct_MarkdownblocksFindsMarkdownMalformedEventsGo(t *testing.T) {
	t.Parallel()
	oracles.MalformedEventsGoOracle(t)
}

func TestProduct_MarkdownblocksFindsGrainMarkdown_path_go(t *testing.T) {
	t.Parallel()
	oracles.GrainGoOracle(t, "path")
}

func TestProduct_MarkdownblocksFindsGrainMarkdown_mdast_go(t *testing.T) {
	t.Parallel()
	oracles.GrainGoOracle(t, "mdast")
}

func TestProduct_MarkdownblocksFindsGrainMarkdown_decode_go(t *testing.T) {
	t.Parallel()
	oracles.GrainGoOracle(t, "decode")
}

func TestProduct_MarkdownblocksFindsGrainMarkdown_width_go(t *testing.T) {
	t.Parallel()
	oracles.GrainGoOracle(t, "width")
}

func TestProduct_MarkdownblocksFindsGrainMarkdownEntitiesGenerator(t *testing.T) {
	t.Parallel()
	oracles.GrainTool(t, "entities")
}

func TestProduct_MarkdownblocksFindsGrainMarkdownWidthGenerator(t *testing.T) {
	t.Parallel()
	oracles.GrainTool(t, "width")
}

func TestProduct_MarkdownblocksFindsGrainMarkdownFormatter(t *testing.T) {
	t.Parallel()
	oracles.GrainTool(t, "formatter")
}

func TestProduct_MarkdownblocksFindsGrainMarkdownWidthFullAnswers(t *testing.T) {
	t.Parallel()
	oracles.GrainMarkdownWidthFullAnswers(t)
}

func TestProduct_MarkdownblocksFindsWholeDocumentPreflightGoProducts(t *testing.T) {
	t.Parallel()
	oracles.PreflightGoProducts(t)
}
