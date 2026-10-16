package lint

import (
	"path/filepath"
	"testing"
)

// Declare the exact recipes fetched by these independently selected shards.
// C-emission performance is tracked separately; no shard budget exemption.
func TestProduct_EmittedMismatchLowered(t *testing.T) {
	t.Parallel()
	requireLintArtifacts(t, emittedMismatchLowered(t), "lint.c", "lint.mjs")
}
func TestProduct_EmittedMismatchNative(t *testing.T) {
	t.Parallel()
	requireLintExecutable(t, emittedMismatchNative(t))
}
func TestProduct_RulesAgreeLowered(t *testing.T) {
	t.Parallel()
	requireLintArtifacts(t, rulesAgreeLowered(t), "lint.c", "lint.mjs")
}
func TestProduct_RulesAgreeNative(t *testing.T) {
	t.Parallel()
	requireLintExecutable(t, filepath.Join(rulesAgreeNative(t), "scanner"))
}
func TestProduct_LintFindsRulesAgreeCapture(t *testing.T) { t.Parallel(); rulesAgreeCapture(t) }
