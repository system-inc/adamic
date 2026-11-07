package lint

import "testing"

// The build phase and independent shards use the same content-addressed recipes. These are every product the
// TestMutantsReactJsxNoCommentTextnodes_ shards read, so Workshop's loom build-tree builds each once and a shard only
// reads them: the setup measured at 323 s inside a shard on Oct 10 (lowering 80 s, the sanitized native 231 s).
func TestProduct_JsxTextnodesMutantLowered(test *testing.T) {
	test.Parallel()
	directory, _ := jsxTextnodesLowered(test)
	requireLintArtifacts(test, directory, "main.c", "lint.mjs")
}

func TestProduct_JsxTextnodesMutantNative(test *testing.T) {
	test.Parallel()
	requireLintExecutable(test, jsxTextnodesNative(test))
}

func TestProduct_LintFindsYepestaSetup(test *testing.T) {
	test.Parallel()
	yepestaPrepare(test)
}
