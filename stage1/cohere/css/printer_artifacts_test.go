package css

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestProduct_CSSPrinterSanitizedAndLowered(t *testing.T) {
	t.Parallel()
	requireCSSPrinter(t, cssPrinterProduct(t, -1))
}
func TestProduct_CSSPrinterSemicolonMutant(t *testing.T) {
	t.Parallel()
	requireCSSPrinter(t, cssPrinterProduct(t, 0))
}
func TestProduct_CSSPrinterIndentMutant(t *testing.T) {
	t.Parallel()
	requireCSSPrinter(t, cssPrinterProduct(t, 1))
}
func TestProduct_CSSPrinterWidthMutant(t *testing.T) {
	t.Parallel()
	requireCSSPrinter(t, cssPrinterProduct(t, 2))
}
func TestProduct_CSSPrinterDarwinLeaks(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "darwin" {
		t.Skip("macOS leaks requires an unsanitized executable")
	}
	requireCSSPrinter(t, cssPrinterProduct(t, 3))
}

func requireCSSPrinter(t *testing.T, product cssExecutable) {
	t.Helper()
	manifest := filepath.Join(t.TempDir(), "empty-manifest")
	if err := os.WriteFile(manifest, nil, 0600); err != nil {
		t.Fatal(err)
	}
	buildcache.RequireArtifacts(t, filepath.Dir(product.binary), buildcache.Artifact{Name: "native", Executable: true, Arguments: []string{manifest}}, buildcache.Artifact{Name: "main.c"}, buildcache.Artifact{Name: "main.mjs"}, buildcache.Artifact{Name: "sources/css/print_main.ts"})
}
