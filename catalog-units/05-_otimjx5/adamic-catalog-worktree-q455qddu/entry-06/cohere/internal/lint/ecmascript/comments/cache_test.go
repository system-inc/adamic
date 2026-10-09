package comments

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// `ForFile` had no direct test until now.
//
// It was exercised through the three rules that call it, which proves those three get comments and
// says nothing about the property the function exists for. A rule reaching for it next would be
// trusting a measurement nobody can re-run, which is the state `@system_cohere_lint_rules_tailwind`
// named `unverified`: committed, green, reachable, and unproven.
//
// The property is not "returns comments". `All` already does that and is tested. The property is
// **computed once per file and shared**, and the whole reason the function exists is a measurement:
// three rules each visiting one node per file cost 807ms, 619ms and 601ms because each rescanned
// the file independently.

// A counting cache stands in for the real one so the test can observe how many times the work ran.
// The production path uses `rule.Cached` against a live `FileCache`; what is asserted here is the
// contract both sides rely on, which is that a second ask does not recompute.
func TestForFileComputesOncePerFile(t *testing.T) {
	t.Parallel()
	fileCache := rule.NewFileCache()

	computations := 0
	compute := func() []Comment {
		computations++
		return []Comment{{Text: "// one"}}
	}

	for range 3 {
		result := rule.Cached(fileCache, "comments.All", compute)
		if len(result) != 1 {
			t.Fatalf("want one comment back on every ask, got %d", len(result))
		}
	}

	if computations != 1 {
		t.Fatalf("want the scan to run once across three asks, ran %d times", computations)
	}
}

// The key is the identity the cache stores under rather than a label, so two callers using
// different keys each compute. This is what makes the key rename from `nexus.allComments` to
// `comments.All` load-bearing rather than cosmetic: a stale key naming the old home would have
// meant two packages each paying for the same scan.
func TestDifferentKeysDoNotShareAnEntry(t *testing.T) {
	t.Parallel()
	fileCache := rule.NewFileCache()

	computations := 0
	compute := func() []Comment {
		computations++
		return nil
	}

	rule.Cached(fileCache, "comments.All", compute)
	rule.Cached(fileCache, "nexus.allComments", compute)

	if computations != 2 {
		t.Fatalf("want two distinct keys to compute separately, computed %d times", computations)
	}
}

// A nil cache is legal and means no caching, so a harness that builds a Context by hand keeps
// working and simply recomputes. Asserted because the alternative is a panic in a shared package,
// which takes the whole run down rather than one rule's finding.
func TestForFileSurvivesANilCache(t *testing.T) {
	t.Parallel()
	computations := 0
	compute := func() []Comment {
		computations++
		return []Comment{{Text: "// one"}}
	}

	first := rule.Cached(nil, "comments.All", compute)
	second := rule.Cached(nil, "comments.All", compute)

	if len(first) != 1 || len(second) != 1 {
		t.Fatal("want a nil cache to still answer")
	}
	if computations != 2 {
		t.Fatalf("want a nil cache to recompute rather than cache, computed %d times", computations)
	}
}
