package lint

import (
	"context"
	"testing"
	"time"
)

func TestProduct_CompleteSuggestionGoOracle(t *testing.T) {
	t.Parallel()
	completeSuggestionProductUnit(t, func(ctx context.Context, directory string) {
		completeSuggestionOracleProduct(ctx, t, directory)
	}, false)
}

func TestProduct_CompleteSuggestionLowered(t *testing.T) {
	t.Parallel()
	completeSuggestionProductUnit(t, func(ctx context.Context, directory string) {
		completeSuggestionLoweredProduct(ctx, t, directory, false)
	}, false)
}

func TestProduct_CompleteSuggestionMutantLowered(t *testing.T) {
	t.Parallel()
	completeSuggestionProductUnit(t, func(ctx context.Context, directory string) {
		completeSuggestionLoweredProduct(ctx, t, directory, true)
	}, true)
}

func TestProduct_CompleteSuggestionNative(t *testing.T) {
	t.Parallel()
	completeSuggestionProductUnit(t, func(ctx context.Context, directory string) {
		completeSuggestionNativeProduct(ctx, t, directory)
	}, false)
}

func completeSuggestionProductUnit(t *testing.T, build func(context.Context, string), mutant bool) {
	t.Helper()
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	build(ctx, completeSuggestionFixture(t, mutant))
	elapsed := time.Since(started)
	t.Logf("%s: %.3fs", t.Name(), elapsed.Seconds())
	if elapsed >= 60*time.Second {
		t.Fatal("build product unit exceeds 60s budget")
	}
}
