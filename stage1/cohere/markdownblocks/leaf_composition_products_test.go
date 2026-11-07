package markdownblocks

import (
	"path/filepath"
	"testing"
)

func leafCompositionProductRoot(test *testing.T) string {
	test.Helper()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return root
}

func leafCompositionLoweredProduct(test *testing.T) leafCompositionProducts {
	test.Helper()
	main, absolutePathError := filepath.Abs("testdata/list_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	products := leafCompositionProducts{root: leafCompositionProductRoot(test), main: main}
	prepareLeafCompositionLowered(test, &products)
	return products
}

func TestProduct_MarkdownLeafLowered(test *testing.T) {
	test.Parallel()
	leafCompositionLoweredProduct(test)
}

func TestProduct_MarkdownLeafNativeSanitized(test *testing.T) {
	test.Parallel()
	products := leafCompositionLoweredProduct(test)
	leafCompositionNative(test, products.source, true)
}

func TestProduct_MarkdownLeafNativeRelease(test *testing.T) {
	test.Parallel()
	products := leafCompositionLoweredProduct(test)
	leafCompositionNative(test, products.source, false)
}
