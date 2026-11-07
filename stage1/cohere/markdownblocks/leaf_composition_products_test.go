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
	p := leafCompositionProducts{root: leafCompositionProductRoot(test), main: main}
	prepareLeafCompositionLowered(test, &p)
	return p
}

func TestProduct_MarkdownLeafLowered(test *testing.T) {
	test.Parallel()
	leafCompositionLoweredProduct(test)
}

func TestProduct_MarkdownLeafNativeSanitized(test *testing.T) {
	test.Parallel()
	p := leafCompositionLoweredProduct(test)
	leafCompositionNative(test, p.source, true)
}

func TestProduct_MarkdownLeafNativeRelease(test *testing.T) {
	test.Parallel()
	p := leafCompositionLoweredProduct(test)
	leafCompositionNative(test, p.source, false)
}
