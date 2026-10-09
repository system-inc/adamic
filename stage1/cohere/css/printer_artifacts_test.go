package css

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestProduct_CSSPrinterParserOracle(t *testing.T) {
	t.Parallel()
	requireCSSOracle(t, cssOracleProduct(t, false))
}
func TestProduct_CSSPrinterOracle(t *testing.T) {
	t.Parallel()
	requireCSSOracle(t, cssOracleProduct(t, true))
}
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

func requireCSSOracle(t *testing.T, binary string) {
	t.Helper()
	buildcache.RequireArtifacts(t, filepath.Dir(binary), buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"-test.list=^$"}})
}
func requireCSSPrinter(t *testing.T, product cssExecutable) {
	t.Helper()
	manifest := filepath.Join(t.TempDir(), "empty-manifest")
	if err := os.WriteFile(manifest, nil, 0600); err != nil {
		t.Fatal(err)
	}
	buildcache.RequireArtifacts(t, filepath.Dir(product.binary), buildcache.Artifact{Name: "native", Executable: true, Arguments: []string{manifest}}, buildcache.Artifact{Name: "main.c"}, buildcache.Artifact{Name: "main.mjs"}, buildcache.Artifact{Name: "sources/css/print_main.ts"})
}
