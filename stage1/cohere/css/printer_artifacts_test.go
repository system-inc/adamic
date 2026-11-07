package css

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestProduct_CSSPrinterSanitizedAndLowered(test *testing.T) {
	test.Parallel()
	requireCSSPrinter(test, cssPrinterProduct(test, -1))
}
func TestProduct_CSSPrinterSemicolonMutant(test *testing.T) {
	test.Parallel()
	requireCSSPrinter(test, cssPrinterProduct(test, 0))
}
func TestProduct_CSSPrinterIndentMutant(test *testing.T) {
	test.Parallel()
	requireCSSPrinter(test, cssPrinterProduct(test, 1))
}
func TestProduct_CSSPrinterWidthMutant(test *testing.T) {
	test.Parallel()
	requireCSSPrinter(test, cssPrinterProduct(test, 2))
}
func TestProduct_CSSPrinterDarwinLeaks(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "darwin" {
		test.Skip("macOS leaks requires an unsanitized executable")
	}
	requireCSSPrinter(test, cssPrinterProduct(test, 3))
}

func requireCSSPrinter(test *testing.T, product cssExecutable) {
	test.Helper()
	manifest := filepath.Join(test.TempDir(), "empty-manifest")
	if writeError := os.WriteFile(manifest, nil, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	buildcache.RequireArtifacts(
		test,
		filepath.Dir(product.binary),
		buildcache.Artifact{Name: "native", Executable: true, Arguments: []string{manifest}},
		buildcache.Artifact{Name: "main.c"},
		buildcache.Artifact{Name: "main.mjs"},
		buildcache.Artifact{Name: "sources/css/print_main.ts"},
	)
}
