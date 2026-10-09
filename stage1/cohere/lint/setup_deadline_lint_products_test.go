package lint

import (
	"context"
	"testing"
)

// Build-phase units call the same recipes and keys as independently selected shards.
func TestProduct_EmittedMismatchOracle(t *testing.T) {
	t.Parallel()
	emittedMismatchOracle(t)
}
func TestProduct_EmittedMismatchLowered(t *testing.T) {
	t.Parallel()
	emittedMismatchLowered(t)
}
func TestProduct_EmittedMismatchNative(t *testing.T) {
	t.Parallel()
	emittedMismatchNative(t)
}
func TestProduct_FactoryHooksLowered(t *testing.T) {
	t.Parallel()
	factoryHooksLowered(t, context.Background(), 0)
}
func TestProduct_FactoryHooksMutantLowered(t *testing.T) {
	t.Parallel()
	factoryHooksLowered(t, context.Background(), 1)
}
func TestProduct_FactoryHooksNative(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	factoryHooksNative(t, ctx, factoryHooksLowered(t, ctx, 0))
}
func TestProduct_RulesAgreeLowered(t *testing.T) {
	t.Parallel()
	rulesAgreeLowered(t)
}
func TestProduct_RulesAgreeNative(t *testing.T) {
	t.Parallel()
	rulesAgreeNative(t)
}

// Independently selectable shared preparation check; it imposes no setup deadline.
func TestRulesAgree_Setup(t *testing.T) {
	t.Parallel()
	rulesAgreeSetup(t)
}
