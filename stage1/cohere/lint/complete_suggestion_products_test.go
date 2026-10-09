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
	// A product unit carries no test-side budget or deadline: the build phase measures it, and a
	// product over 60 s is a build-phase finding (#c5k975w), never a test failure.
	started := time.Now()
	build(context.Background(), completeSuggestionFixture(t, mutant))
	t.Logf("%s: %.3fs", t.Name(), time.Since(started).Seconds())
}
