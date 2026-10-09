package markdownblocks

import (
	"path/filepath"
	"testing"
)

func leafCompositionProductRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func leafCompositionLoweredProduct(t *testing.T) leafCompositionProducts {
	t.Helper()
	main, err := filepath.Abs("testdata/list_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	p := leafCompositionProducts{root: leafCompositionProductRoot(t), main: main}
	prepareLeafCompositionLowered(t, &p)
	return p
}

func TestProduct_MarkdownLeafGoLists(t *testing.T) {
	t.Parallel()
	leafCompositionGoProduct(t, leafCompositionProductRoot(t), "adamic_markdown_lists", "list_go.go")
}

func TestProduct_MarkdownLeafGoDocLayout(t *testing.T) {
	t.Parallel()
	leafCompositionGoProduct(t, leafCompositionProductRoot(t), "adamic_markdown_doclayout", "document_go.go")
}

func TestProduct_MarkdownLeafLowered(t *testing.T) {
	t.Parallel()
	leafCompositionLoweredProduct(t)
}

func TestProduct_MarkdownLeafNativeSanitized(t *testing.T) {
	t.Parallel()
	p := leafCompositionLoweredProduct(t)
	leafCompositionNative(t, p.source, true)
}

func TestProduct_MarkdownLeafNativeRelease(t *testing.T) {
	t.Parallel()
	p := leafCompositionLoweredProduct(t)
	leafCompositionNative(t, p.source, false)
}
