package lint

import (
	"github.com/system-inc/adamic/stage1/cohere/lint/oracles"

	"testing"
)

// lintGoOracleProduct is the Go oracle over this package's live registry and rule oracles, laid over cohere
// (goOracleOverlay), built ahead as a buildcache.GoBuildIn product: toolchain's go in cohere under its environment,
// keyed by go's own answer there (cohere's go.mod, go.sum and go.work, the packages it reads, the overlay's map and
// each replacement's content, go's settings and release). Every consumer of the live port's oracle reads this one
// product.
func lintGoOracleProduct(test *testing.T) string { test.Helper(); return oracles.LintGoOracle(test) }
func TestProduct_LintFindsGoOracle(test *testing.T) {
	test.Parallel()
	lintGoOracleProduct(test)
}
