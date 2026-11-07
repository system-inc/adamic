package printer

import (
	"path/filepath"
	"sync"
	"testing"
)

// printerInTreeSource is the package's own main.ts, which TestPrinterFileDriver and the whitespace gap units lower;
// the GraphQLPrinter products above are the as-Go copy's.
func printerInTreeSource(test *testing.T) string {
	test.Helper()
	path, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return path
}

// Each set of products is looked up once per process: a lookup lists the repository and keys every file it names, and
// the five whitespace units all read the copy's.
type printerProductsOnce struct {
	once     sync.Once
	products printerProducts
}

func (once *printerProductsOnce) get(test *testing.T, prepare func() printerProducts) printerProducts {
	test.Helper()
	once.once.Do(func() { once.products = prepare() })
	if once.products.sanitized == "" {
		test.Fatal("shared GraphQL product preparation failed earlier in this process (see that test's log)")
	}
	return once.products
}

var inTreeProducts, copyProducts printerProductsOnce

func printerInTreeProducts(test *testing.T) printerProducts {
	test.Helper()
	return inTreeProducts.get(
		test,
		func() printerProducts { return preparePrinterProducts(test, printerInTreeSource(test)) },
	)
}

// printerCopyProducts are the products of an unmutated copy of the port. Every copy holds the same bytes, so the first
// caller's products are every copy's; each caller still runs Node on a copy of its own, since the first caller's copy
// leaves with its test.
func printerCopyProducts(test *testing.T) printerProducts {
	test.Helper()
	source := printerDirectory(test, "", "", "")
	products := copyProducts.get(test, func() printerProducts { return preparePrinterProducts(test, source) })
	products.source = source
	return products
}

// No TestProduct_ built the in-tree printer or the whitespace gap's copy, so Workshop's combined proof (Oct 10) found
// the lowered and sanitized printers missing on the reading checkout. These build exactly what preparePrinterProducts
// reads for each.
func TestProduct_GraphQLPrinterInTreeLowered(test *testing.T) {
	test.Parallel()
	printerLoweredProduct(test, printerInTreeSource(test))
}

func TestProduct_GraphQLPrinterInTreeNative(test *testing.T) {
	test.Parallel()
	printerInTreeProducts(test)
}

// The whitespace gap units lower an unmutated copy of the port (printerDirectory with no change) rather than the
// package's own main.ts; this builds what they read, keyed as they key it.
func TestProduct_GraphQLPrinterCopyNative(test *testing.T) {
	test.Parallel()
	printerCopyProducts(test)
}
