package lint

import "testing"

// The build phase and independent shards use the same content-addressed recipes. These are every product the
// TestMutantsReactJsxNoCommentTextnodes_ shards read, so Workshop's loom build-tree builds each once and a shard only
// reads them: the setup measured at 323 s inside a shard on Oct 10 (lowering 80 s, the sanitized native 231 s).
func TestProduct_YepestaOracle(t *testing.T) {
	t.Parallel()
	yepestaOracle(t)
}

func TestProduct_JsxTextnodesMutantLowered(t *testing.T) {
	t.Parallel()
	jsxTextnodesLowered(t)
}

func TestProduct_JsxTextnodesMutantNative(t *testing.T) {
	t.Parallel()
	jsxTextnodesNative(t)
}

func TestProduct_YepestaSetup(t *testing.T) {
	t.Parallel()
	yepestaPrepare(t)
}
