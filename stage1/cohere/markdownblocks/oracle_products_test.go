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
