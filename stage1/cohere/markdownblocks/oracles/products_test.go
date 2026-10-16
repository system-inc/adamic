package oracles

import "testing"

func TestProduct_MarkdownChunksGo(t *testing.T) { t.Parallel(); ChunksGoOracle(t) }

func TestProduct_MarkdownASTGo(t *testing.T) { t.Parallel(); ASTGoOracle(t) }

func TestProduct_MarkdownFrontMatterGo(t *testing.T) { t.Parallel(); FrontMatterGoOracle(t) }

func TestProduct_MarkdownPrefixGo(t *testing.T) { t.Parallel(); PrefixGoOracle(t) }

func TestProduct_MarkdownIdentifierGo(t *testing.T) { t.Parallel(); PilotGoProduct(t, "identifier") }

func TestProduct_TokenizerEventsGo(t *testing.T) { t.Parallel(); PilotGoProduct(t, "events") }

func TestProduct_TokenizerEventsAnswers(t *testing.T) { t.Parallel(); TokenizerEventsAnswers(t) }

func TestProduct_MarkdownTextGo(t *testing.T) { t.Parallel(); TextGoBuild(t, "go") }

func TestProduct_MarkdownTextFormatter(t *testing.T) { t.Parallel(); TextGoBuild(t, "formatter") }

func TestProduct_MarkdownTextGenerator(t *testing.T) { t.Parallel(); TextGoBuild(t, "generator") }

func TestProduct_MarkdownMalformedEventsGo(t *testing.T) { t.Parallel(); MalformedEventsGoOracle(t) }

func TestProduct_GrainMarkdown_path_go(t *testing.T) { t.Parallel(); GrainGoOracle(t, "path") }

func TestProduct_GrainMarkdown_mdast_go(t *testing.T) { t.Parallel(); GrainGoOracle(t, "mdast") }

func TestProduct_GrainMarkdown_decode_go(t *testing.T) { t.Parallel(); GrainGoOracle(t, "decode") }

func TestProduct_GrainMarkdown_width_go(t *testing.T) { t.Parallel(); GrainGoOracle(t, "width") }

func TestProduct_GrainMarkdownEntitiesGenerator(t *testing.T) { t.Parallel(); GrainTool(t, "entities") }

func TestProduct_GrainMarkdownWidthGenerator(t *testing.T) { t.Parallel(); GrainTool(t, "width") }

func TestProduct_GrainMarkdownFormatter(t *testing.T) { t.Parallel(); GrainTool(t, "formatter") }

func TestProduct_GrainMarkdownWidthFullAnswers(t *testing.T) {
	t.Parallel()
	GrainMarkdownWidthFullAnswers(t)
}

func TestProduct_MarkdownLeafGoLists(t *testing.T) {
	t.Parallel()
	LeafGoOracle(t, "adamic_markdown_lists", "list_go.go")
}

func TestProduct_MarkdownLeafGoDocLayout(t *testing.T) {
	t.Parallel()
	LeafGoOracle(t, "adamic_markdown_doclayout", "document_go.go")
}

func TestProduct_MarkdownQuote_lists(t *testing.T) { t.Parallel(); QuoteGoOracle(t, "lists") }

func TestProduct_MarkdownQuote_layout(t *testing.T) { t.Parallel(); QuoteGoOracle(t, "layout") }

func TestProduct_MarkdownStructure_adamic_markdown_lists(t *testing.T) {
	t.Parallel()
	StructureGoOracle(t, "adamic_markdown_lists")
}

func TestProduct_MarkdownStructure_adamic_markdown_doclayout(t *testing.T) {
	t.Parallel()
	StructureGoOracle(t, "adamic_markdown_doclayout")
}

func TestProduct_MarkdownTable_list(t *testing.T) { t.Parallel(); TableGoOracle(t, "list") }

func TestProduct_MarkdownTable_document(t *testing.T) { t.Parallel(); TableGoOracle(t, "document") }

func TestProduct_MarkdownWhitespace_lists(t *testing.T) {
	t.Parallel()
	WhitespaceGoOracle(t, t.Context(), "lists")
}

func TestProduct_MarkdownWhitespace_layout(t *testing.T) {
	t.Parallel()
	WhitespaceGoOracle(t, t.Context(), "layout")
}

// TestProduct_WholeDocumentPreflightGoProducts builds the Go printer and its three mutants that every preflight unit
// reads (markdownblocks-whole-document-preflight-go-products, 75 s measured inside a unit on Oct 10), so Workshop
// builds them once and a unit only reads them.
func TestProduct_WholeDocumentPreflightGoProducts(t *testing.T) {
	t.Parallel()
	PreflightGoProducts(t)
}
