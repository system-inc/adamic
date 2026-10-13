package printer

import (
	"path/filepath"
	"sync"
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

// Each set of products is looked up once per process: a lookup lists the repository and keys every file it names, and
// the five whitespace units all read the copy's.
type printerProductsOnce struct {
	once     sync.Once
	products printerProducts
}

func (p *printerProductsOnce) get(t *testing.T, prepare func() printerProducts) printerProducts {
	t.Helper()
	p.once.Do(func() { p.products = prepare() })
	if p.products.sanitized == "" {
		t.Fatal("shared GraphQL product preparation failed earlier in this process (see that test's log)")
	}
	return p.products
}

var inTreeProducts, copyProducts printerProductsOnce

func printerInTreeProducts(t *testing.T) printerProducts {
	t.Helper()
	return inTreeProducts.get(t, func() printerProducts { return preparePrinterProducts(t, printerInTreeSource(t)) })
}

// printerCopyProducts are the products of an unmutated copy of the port. Every copy holds the same bytes, so the first
// caller's products are every copy's; each caller still runs Node on a copy of its own, since the first caller's copy
// leaves with its test.
func printerCopyProducts(t *testing.T) printerProducts {
	t.Helper()
	source := printerDirectory(t, "", "", "")
	products := copyProducts.get(t, func() printerProducts { return preparePrinterProducts(t, source) })
	products.source = source
	return products
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
	printerInTreeProducts(t)
}

// The whitespace gap units lower an unmutated copy of the port (printerDirectory with no change) rather than the
// package's own main.ts; this builds what they read, keyed as they key it.
func TestProduct_GraphQLPrinterCopyNative(t *testing.T) {
	t.Parallel()
	printerCopyProducts(t)
}
