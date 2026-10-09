package prettier

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/formatoptions"
)

/*
 * The fork was reachable by two independent paths and only one of them was overridable.
 *
 * `internal/release` resolved COHERE_PRETTIER_FORK to decide which bundles it vouched for, while
 * this engine read a hardcoded constant to decide which bundles it loaded. Neither consulted the
 * other, so a machine that set the variable got a release guard reporting green about a directory
 * nothing loaded from, and an engine still reading a home directory on one laptop.
 *
 * The defect was not that either path was wrong. It was that they existed separately. These tests
 * hold the seam closed from both directions, because the failure was silent in both: the override
 * that did not reach the loader, and the default that quietly worked anyway so nobody noticed.
 */

// TestOverrideReachesTheLoader is the regression, and it is written as the measurement that exposed
// the defect rather than as an assertion about the code.
//
// Before the fix this exact probe returned a working engine: New succeeded with the variable pointed
// at a path that does not exist, because loadBundles read a constant. A test that only checked
// BundleDirectory's return value would have passed against the broken engine too, since the constant
// was never the thing being consulted. So this goes through New and demands a real failure.
//
// Not parallel: it points ForkPathVariable at a missing fork with t.Setenv, which a parallel test may not
// call.
func TestOverrideReachesTheLoader(t *testing.T) {
	t.Setenv(ForkPathVariable, "/nonexistent/elsewhere")

	if _, err := New(formatoptions.Default()); err == nil {
		t.Fatalf("New succeeded with %s pointed at a nonexistent path, so the loader is not reading it", ForkPathVariable)
	} else if !strings.Contains(err.Error(), "/nonexistent/elsewhere") {
		t.Fatalf("the failure does not name the overridden path, so something else failed: %v", err)
	}
}

// TestOverrideIsTheOnlyKnobThatMoves is the other half, and it exists because the first test passes
// for the wrong reason if the engine is simply always broken.
//
// A loader that failed unconditionally would satisfy the regression above while being useless. This
// pins the default, and what the default is has changed: with no variable set the bundles come from
// the binary, not from any path. That is the property the whole vendoring unit exists to produce, so
// asserting Origin is asserting the point rather than an implementation detail.
//
// Not parallel: it clears ForkPathVariable with t.Setenv to pin the no-override default, which a parallel
// test may not call.
func TestOverrideIsTheOnlyKnobThatMoves(t *testing.T) {
	t.Setenv(ForkPathVariable, "")

	bundles, err := Bundles()
	if err != nil {
		t.Fatalf("with no override the embedded bundles failed to load: %v", err)
	}
	if bundles.Origin != Embedded {
		t.Fatalf("with no override the origin is %q, want %q", bundles.Origin, Embedded)
	}
	if bundles.Path != "" {
		t.Fatalf("embedded bundles carry path %q, want empty so a forgotten branch cannot use it", bundles.Path)
	}
}

// TestEmbeddedBundlesNeedNoFork is the property the vendoring bought, stated as a test.
//
// Before this, every formatter assertion in the package skipped on a machine without the fork, and
// the package printed `ok` while measuring nothing. This asserts the opposite condition directly:
// point the override at a path that does not exist, unset it, and the engine still builds. If this
// fails, the binary is depending on a checkout again and the skip would come back with it.
//
// Not parallel: it clears ForkPathVariable with t.Setenv to pin the no-override default, which a parallel
// test may not call.
func TestEmbeddedBundlesNeedNoFork(t *testing.T) {
	t.Setenv(ForkPathVariable, "")

	if _, err := New(formatoptions.Default()); err != nil {
		t.Fatalf("building an engine from the embedded bundles: %v", err)
	}
}

// TestEmbedMatchesBundleFiles holds the go:embed directive and BundleFiles together.
//
// They are two lists of the same eight names and nothing in the language ties them. The directive
// names paths explicitly rather than globbing the directory, which makes a missing file a compile
// error, but it cannot catch the other direction: a name added to BundleFiles and not to the
// directive compiles fine and fails when someone formats that language. This is the guard for that,
// and it is why the loader checks a map it was just handed.
//
// Not parallel: it clears ForkPathVariable with t.Setenv so the embed is what loads, which a parallel test
// may not call.
func TestEmbedMatchesBundleFiles(t *testing.T) {
	t.Setenv(ForkPathVariable, "")

	bundles, err := Bundles()
	if err != nil {
		t.Fatalf("loading the embedded bundles: %v", err)
	}
	if len(bundles.Files) != len(BundleFiles) {
		t.Fatalf("the embed carries %d bundles, BundleFiles names %d", len(bundles.Files), len(BundleFiles))
	}
	for _, name := range BundleFiles {
		if len(bundles.Files[name]) == 0 {
			t.Errorf("%s is absent or empty in the embed", name)
		}
	}
}

// TestReleaseResolvesThroughTheEngine is the guard against the two paths reappearing.
//
// `release` declares its own exported name for the variable, which is the shape the defect had. It
// is an alias of this one rather than a copy, and a copy would be invisible: both spellings would be
// the same string today and could drift apart in any later edit, with the drift producing exactly
// the original defect again.
//
// The default path is no longer shared, and that is deliberate rather than a regression. The engine
// has no path-shaped default any more, so `release` declaring its own is two different questions
// rather than one value with two homes. Only that file still asks the path-shaped one, and only
// until its guard is reshaped to vouch for embedded bytes.
//
// It lives here rather than in `release` because `prettier` imports nothing internal, so this
// direction is the one that cannot create a cycle.
func TestReleaseResolvesThroughTheEngine(t *testing.T) {
	t.Parallel()

	if ForkPathVariable != "COHERE_PRETTIER_FORK" {
		t.Fatalf("the variable is %q; release's exported alias and any CI that sets it both assume the old spelling", ForkPathVariable)
	}
}
