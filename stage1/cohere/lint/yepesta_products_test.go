package lint

import "testing"

// The build phase and independent shards use the same content-addressed recipes.
func TestProduct_YepestaOracle(t *testing.T) {
	t.Parallel()
	yepestaOracle(t)
}
