package lint

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"sync"
	"testing"
)

var lintOracleOnce sync.Once
var lintOraclePath string

// lintGoOracleProduct is the Go oracle over this package's live registry and rule oracles, laid over cohere
// (goOracleOverlay), built ahead as a buildcache.GoBuildIn product: toolchain's go in cohere under its environment,
// keyed by go's own answer there (cohere's go.mod, go.sum and go.work, the packages it reads, the overlay's map and each
// replacement's content, go's settings and release). Every consumer of the live port's oracle reads this one product.
func lintGoOracleProduct(t *testing.T) string {
	t.Helper()
	lintOracleOnce.Do(func() {
		prepareRegistry(t, packageDirectory)
		path, err := goOracleOverlay(packageDirectory, t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		// Each rule's oracle.go is built only through the overlay (//go:build lintoracle), never in its own directory.
		lintOraclePath = buildcache.GoBuildIn(t, "cohere", "oracle", "./"+goOraclePackage, []string{"-tags=lintoracle", "-overlay=" + path})
	})
	if lintOraclePath == "" {
		t.Fatal("lint Go oracle build failed earlier in this process")
	}
	return lintOraclePath
}
func TestProduct_LintGoOracle(t *testing.T) {
	t.Parallel()
	lintGoOracleProduct(t)
}
