package printer

import (
	"path/filepath"
	"testing"
)

// printerInTreeSource is the package's own main.ts, which TestPrinterFileDriver and the whitespace gap units lower;
// the GraphQLPrinter products above are the as-Go copy's.
func printerInTreeSource(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// No TestProduct_ built the in-tree printer or the whitespace gap's copy, so Workshop's combined proof (Oct 10) found
// the lowered and sanitized printers missing on the reading checkout. These build exactly what preparePrinterProducts
// reads for each.
func TestProduct_GraphQLPrinterInTreeLowered(t *testing.T) {
	t.Parallel()
	printerLoweredProduct(t, printerInTreeSource(t))
}

func TestProduct_GraphQLPrinterInTreeNative(t *testing.T) {
	t.Parallel()
	preparePrinterProducts(t, printerInTreeSource(t))
}

// The whitespace gap units lower an unmutated copy of the port (printerDirectory with no change) rather than the
// package's own main.ts; this builds what they read, keyed as they key it.
func TestProduct_GraphQLPrinterCopyNative(t *testing.T) {
	t.Parallel()
	preparePrinterProducts(t, printerDirectory(t, "", "", ""))
}
