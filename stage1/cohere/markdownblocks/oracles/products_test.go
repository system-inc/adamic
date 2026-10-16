package oracles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	EnableProductProduction()
	os.Exit(m.Run())
}
func productRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}
func TestProduct_MarkdownblocksLeafLists(t *testing.T) {
	t.Parallel()
	LeafCompositionGoProduct(t, productRoot(t), "adamic_markdown_lists", "list_go.go")
}
func TestProduct_MarkdownblocksLeafDocLayout(t *testing.T) {
	t.Parallel()
	LeafCompositionGoProduct(t, productRoot(t), "adamic_markdown_doclayout", "document_go.go")
}
func TestProduct_MarkdownblocksQuote_lists(t *testing.T) {
	t.Parallel()
	QuoteLayoutGoProducts(t, productRoot(t), "lists")
}
func TestProduct_MarkdownblocksQuote_layout(t *testing.T) {
	t.Parallel()
	QuoteLayoutGoProducts(t, productRoot(t), "layout")
}
func TestProduct_MarkdownblocksStructure_adamic_markdown_lists(t *testing.T) {
	t.Parallel()
	StructureLayoutGoProducts(t, productRoot(t), "adamic_markdown_lists")
}
func TestProduct_MarkdownblocksStructure_adamic_markdown_doclayout(t *testing.T) {
	t.Parallel()
	StructureLayoutGoProducts(t, productRoot(t), "adamic_markdown_doclayout")
}
func TestProduct_MarkdownblocksTable_list(t *testing.T) {
	t.Parallel()
	TableLayoutGoProducts(t, productRoot(t), "list")
}
func TestProduct_MarkdownblocksTable_document(t *testing.T) {
	t.Parallel()
	TableLayoutGoProducts(t, productRoot(t), "document")
}
func TestProduct_MarkdownblocksWhitespace_lists(t *testing.T) {
	t.Parallel()
	WhitespaceLayoutGoProducts(t, t.Context(), productRoot(t), "lists")
}
func TestProduct_MarkdownblocksWhitespace_layout(t *testing.T) {
	t.Parallel()
	WhitespaceLayoutGoProducts(t, t.Context(), productRoot(t), "layout")
}

func TestProduct_MarkdownblocksTextFormatter(t *testing.T) {
	t.Parallel()
	TextFormatterProduct(t, productRoot(t))
}

func TestProduct_MarkdownblocksTextGenerator(t *testing.T) {
	t.Parallel()
	TextGeneratorProduct(t, productRoot(t))
}
