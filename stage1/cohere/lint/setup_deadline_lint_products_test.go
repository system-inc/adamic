package lint

import (
	"context"
	"path/filepath"
	"testing"
)

// Build-phase units call the same recipes and keys as independently selected shards.
// Full-port lowering and its dependent native products remain shard preparation
// until their cold builds meet the build phase grain (#c5k975w, #3he8f8g).
func TestProduct_FactoryHooksLowered(test *testing.T) {
	test.Parallel()
	requireLintArtifacts(test, filepath.Dir(factoryHooksLowered(test, context.Background(), 0)), "lint.c", "lint.mjs")
}
func TestProduct_FactoryHooksMutantLowered(test *testing.T) {
	test.Parallel()
	requireLintArtifacts(test, filepath.Dir(factoryHooksLowered(test, context.Background(), 1)), "lint.c", "lint.mjs")
}
func TestProduct_FactoryHooksNative(test *testing.T) {
	test.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	requireLintExecutable(test, factoryHooksNative(test, ctx, factoryHooksLowered(test, ctx, 0)))
}

// Independently selectable shared preparation check; it imposes no setup deadline.
func TestRulesAgree_Setup(test *testing.T) {
	test.Parallel()
	rulesAgreeSetup(test)
}
