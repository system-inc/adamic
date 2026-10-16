package markdownblocks

import (
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

// These finders put the ordinary tests' and composites' oracle reads in this package's product list, so its units fetch
// them. The oracles package owns production; cloud/fast-gate/product-order.json starts each finder after its owner.

func TestProduct_MarkdownFindsChunksGo(t *testing.T) {
	t.Parallel()
	oracles.ChunksGoOracle(t)
}

func TestProduct_MarkdownFindsASTGo(t *testing.T) {
	t.Parallel()
	oracles.ASTGoOracle(t)
}

func TestProduct_MarkdownFindsFrontMatterGo(t *testing.T) {
	t.Parallel()
	oracles.FrontMatterGoOracle(t)
}

func TestProduct_MarkdownFindsPrefixGo(t *testing.T) {
	t.Parallel()
	oracles.PrefixGoOracle(t)
}

func TestProduct_MarkdownFindsIdentifierGo(t *testing.T) {
	t.Parallel()
	oracles.PilotGoProduct(t, "identifier")
}

func TestProduct_MarkdownFindsTokenizerEventsGo(t *testing.T) {
	t.Parallel()
	oracles.PilotGoProduct(t, "events")
}

func TestProduct_MarkdownFindsTokenizerEventsAnswers(t *testing.T) {
	t.Parallel()
	oracles.TokenizerEventsAnswers(t)
}

func TestProduct_MarkdownFindsTextGo(t *testing.T) {
	t.Parallel()
	oracles.TextGoBuild(t, "go")
}

func TestProduct_MarkdownFindsTextFormatter(t *testing.T) {
	t.Parallel()
	oracles.TextGoBuild(t, "formatter")
}

func TestProduct_MarkdownFindsTextGenerator(t *testing.T) {
	t.Parallel()
	oracles.TextGoBuild(t, "generator")
}

func TestProduct_MarkdownFindsMalformedEventsGo(t *testing.T) {
	t.Parallel()
	oracles.MalformedEventsGoOracle(t)
}

func TestProduct_MarkdownFindsGrain_path_go(t *testing.T) {
	t.Parallel()
	oracles.GrainGoOracle(t, "path")
}

func TestProduct_MarkdownFindsGrain_mdast_go(t *testing.T) {
	t.Parallel()
	oracles.GrainGoOracle(t, "mdast")
}

func TestProduct_MarkdownFindsGrain_decode_go(t *testing.T) {
	t.Parallel()
	oracles.GrainGoOracle(t, "decode")
}

func TestProduct_MarkdownFindsGrain_width_go(t *testing.T) {
	t.Parallel()
	oracles.GrainGoOracle(t, "width")
}

func TestProduct_MarkdownFindsGrainEntitiesGenerator(t *testing.T) {
	t.Parallel()
	oracles.GrainTool(t, "entities")
}

func TestProduct_MarkdownFindsGrainWidthGenerator(t *testing.T) {
	t.Parallel()
	oracles.GrainTool(t, "width")
}

func TestProduct_MarkdownFindsGrainFormatter(t *testing.T) {
	t.Parallel()
	oracles.GrainTool(t, "formatter")
}

func TestProduct_MarkdownFindsGrainWidthFullAnswers(t *testing.T) {
	t.Parallel()
	oracles.GrainMarkdownWidthFullAnswers(t)
}

func TestProduct_MarkdownFindsLeafGoLists(t *testing.T) {
	t.Parallel()
	oracles.LeafGoOracle(t, "adamic_markdown_lists", "list_go.go")
}

func TestProduct_MarkdownFindsLeafGoDocLayout(t *testing.T) {
	t.Parallel()
	oracles.LeafGoOracle(t, "adamic_markdown_doclayout", "document_go.go")
}

func TestProduct_MarkdownFindsQuote_lists(t *testing.T) {
	t.Parallel()
	oracles.QuoteGoOracle(t, "lists")
}

func TestProduct_MarkdownFindsQuote_layout(t *testing.T) {
	t.Parallel()
	oracles.QuoteGoOracle(t, "layout")
}

func TestProduct_MarkdownFindsStructure_adamic_markdown_lists(t *testing.T) {
	t.Parallel()
	oracles.StructureGoOracle(t, "adamic_markdown_lists")
}

func TestProduct_MarkdownFindsStructure_adamic_markdown_doclayout(t *testing.T) {
	t.Parallel()
	oracles.StructureGoOracle(t, "adamic_markdown_doclayout")
}

func TestProduct_MarkdownFindsTable_list(t *testing.T) {
	t.Parallel()
	oracles.TableGoOracle(t, "list")
}

func TestProduct_MarkdownFindsTable_document(t *testing.T) {
	t.Parallel()
	oracles.TableGoOracle(t, "document")
}

func TestProduct_MarkdownFindsWhitespace_lists(t *testing.T) {
	t.Parallel()
	oracles.WhitespaceGoOracle(t, t.Context(), "lists")
}

func TestProduct_MarkdownFindsWhitespace_layout(t *testing.T) {
	t.Parallel()
	oracles.WhitespaceGoOracle(t, t.Context(), "layout")
}

func TestProduct_MarkdownFindsWholeDocumentPreflightGoProducts(t *testing.T) {
	t.Parallel()
	oracles.PreflightGoProducts(t)
}
