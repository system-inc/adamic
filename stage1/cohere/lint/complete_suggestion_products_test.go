package lint

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestProduct_LintFindsCompleteSuggestionGoOracle(test *testing.T) {
	test.Parallel()
	completeSuggestionProductUnit(test, func(ctx context.Context, directory string) {
		completeSuggestionOracleProduct(ctx, test, directory)
	}, false)
}

func TestProduct_CompleteSuggestionLowered(test *testing.T) {
	test.Parallel()
	completeSuggestionProductUnit(test, func(ctx context.Context, directory string) {
		requireLintArtifacts(test, completeSuggestionLoweredProduct(ctx, test, directory, false), "lint.c", "lint.mjs")
	}, false)
}

func TestProduct_CompleteSuggestionMutantLowered(test *testing.T) {
	test.Parallel()
	completeSuggestionProductUnit(test, func(ctx context.Context, directory string) {
		requireLintArtifacts(test, completeSuggestionLoweredProduct(ctx, test, directory, true), "lint.c", "lint.mjs")
	}, true)
}

func TestProduct_CompleteSuggestionNative(test *testing.T) {
	test.Parallel()
	completeSuggestionProductUnit(test, func(ctx context.Context, directory string) {
		requireLintExecutable(test, filepath.Join(completeSuggestionNativeProduct(ctx, test, directory), "scanner"))
	}, false)
}

func completeSuggestionProductUnit(test *testing.T, build func(context.Context, string), mutant bool) {
	test.Helper()
	// A product unit carries no test-side budget or deadline: the build phase measures it, and a
	// product over 60 s is a build-phase finding (#c5k975w), never a test failure.
	started := time.Now()
	build(context.Background(), completeSuggestionFixture(test, mutant))
	test.Logf("%s: %.3fs", test.Name(), time.Since(started).Seconds())
}
