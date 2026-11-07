package lint

import (
	"path/filepath"
	"testing"
)

// Declare the exact recipes fetched by these independently selected shards.
// C-emission performance is tracked separately; no shard budget exemption.
func TestProduct_EmittedMismatchLowered(test *testing.T) {
	test.Parallel()
	requireLintArtifacts(test, emittedMismatchLowered(test), "lint.c", "lint.mjs")
}
func TestProduct_EmittedMismatchNative(test *testing.T) {
	test.Parallel()
	requireLintExecutable(test, emittedMismatchNative(test))
}
func TestProduct_RulesAgreeLowered(test *testing.T) {
	test.Parallel()
	requireLintArtifacts(test, rulesAgreeLowered(test), "lint.c", "lint.mjs")
}
func TestProduct_RulesAgreeNative(test *testing.T) {
	test.Parallel()
	requireLintExecutable(test, filepath.Join(rulesAgreeNative(test), "scanner"))
}
func TestProduct_LintFindsRulesAgreeCapture(test *testing.T) {
	test.Parallel()
	rulesAgreeCapture(test)
}
