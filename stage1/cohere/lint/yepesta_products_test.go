package lint

import "testing"

// The build phase and independent shards use the same content-addressed recipes.
func TestProduct_YepestaOracle(t *testing.T) {
	t.Parallel()
	yepestaOracle(t)
}

func TestProduct_YepestaLowered(t *testing.T) {
	t.Parallel()
	jsxTextnodesLowered(t)
}

func TestProduct_YepestaNative(t *testing.T) {
	t.Parallel()
	jsxTextnodesNative(t)
}

func TestProduct_YepestaBundle(t *testing.T) {
	t.Parallel()
	yepestaFetch(t)
}
