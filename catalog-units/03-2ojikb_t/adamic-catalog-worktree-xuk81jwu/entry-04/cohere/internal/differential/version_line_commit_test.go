package differential

import (
	"strings"
	"testing"
)

// A committed-tree build stamps its compiler, so its `compiler:` line names the compiler commit and
// carries no "built from" fallback. The binary's own commit is on the `commit:` line. Reading only the
// fallback rendered every such binary as "cohere dev", which identifies nothing.
//
// The text below is what the global cohere printed on 2026-10-02 at 7b434bb4105b, verbatim apart from
// the patch line.
func TestACommittedBuildIsIdentifiedByItsCommitLine(t *testing.T) {
	t.Parallel()
	committed := versionLineFrom(
		"cohere dev\n" +
			"  platform:       darwin/arm64\n" +
			"  go:             go1.27.0\n" +
			"  compiler:       1f70213d4922b434345f639b441681e470c7cfc1\n" +
			"  commit:         7b434bb4105b\n" +
			"  note:           a local build of the commit above, with nothing uncommitted in it, so that commit reproduces it\n",
	)
	if !strings.Contains(committed, "7b434bb4105b") {
		t.Fatalf("a committed build's own commit was lost, got %q", committed)
	}
	if !strings.Contains(committed, "so that commit reproduces it") {
		t.Fatalf("the build's note did not travel with its commit, got %q", committed)
	}
	if strings.Contains(committed, "1f70213d4922") {
		t.Fatalf("the compiler commit was taken for the binary's own, got %q", committed)
	}

	// A binary that prints both lines is identified by the one naming itself. Before the `commit:`
	// line was preferred, this took the compiler fallback; the numbers agree there, but the fallback
	// is the line that disappears when the compiler is stamped.
	both := versionLineFrom(
		"cohere dev\n" +
			"  compiler:       unknown (built from cohere 86723d2c576d)\n" +
			"  commit:         86723d2c576d\n" +
			"  source:         built from a tree with uncommitted changes, so no commit reproduces this binary\n" +
			"  note:           a local build, so the rules are whatever was on disk when it was compiled\n",
	)
	if !strings.HasPrefix(both, "commit:") || !strings.Contains(both, "whatever was on disk") {
		t.Fatalf("a binary naming its commit was not identified by that line, or lost its disclaimer, got %q", both)
	}
}
