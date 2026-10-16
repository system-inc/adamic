package lint

import (
	"github.com/system-inc/adamic/stage1/cohere/lint/oracles"

	"testing"
)

// lintGoOracleProduct is the Go oracle over this package's live registry and rule oracles, laid over cohere
// (goOracleOverlay), built ahead as a buildcache.GoBuildIn product: toolchain's go in cohere under its environment,
// keyed by go's own answer there (cohere's go.mod, go.sum and go.work, the packages it reads, the overlay's map and each
// replacement's content, go's settings and release). Every consumer of the live port's oracle reads this one product.
func lintGoOracleProduct(t *testing.T) string { t.Helper(); return oracles.LintGoOracle(t) }
func TestProduct_LintFindsGoOracle(t *testing.T) {
	t.Parallel()
	lintGoOracleProduct(t)
}
