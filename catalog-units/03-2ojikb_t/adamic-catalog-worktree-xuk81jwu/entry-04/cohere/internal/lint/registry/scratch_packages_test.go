package registry

import (
	"path/filepath"
	"strings"
	"testing"
)

// scratchPackageSuffixes are the names porters give throwaway packages.
//
// The convention arose on its own: several porters independently reached for an isolated package
// when the shared rule package would not compile, and independently named it for their rule plus a
// marker. It is a good pattern and the brief now sanctions it. What it lacks is an ending.
var scratchPackageSuffixes = []string{"iso", "probe"}

// TestNoScratchPackagesRemain refuses a throwaway package left in the tree.
//
// The isolated-package route is how a porter measures anything when a sibling has the shared package
// red at compile time, and `-run` scoping cannot rescue a compile error. So these get created, and
// the instruction to delete them afterwards lived only in a brief nobody re-reads at the end.
//
// Eight were left behind in one session. Most were harmless clutter. One was not: a scratch copy
// registered the same rule name as the real package, so `TestEveryRuleIsRegistered` failed for every
// other agent in the tree, and the failure named a package none of them had heard of.
//
// That is the cost this catches. A porter who forgets pays for it themselves at their own commit,
// rather than everyone paying for it until a coordinator notices.
//
// The suffix list is deliberately narrow. A real package named for what it does will not end in
// `iso` or `probe`, and if one ever legitimately does, adding it to an exemption here is a decision
// somebody makes on purpose rather than a guard quietly widening.
func TestNoScratchPackagesRemain(t *testing.T) {
	t.Parallel()
	entries, err := filepath.Glob("../rules/*")
	if err != nil {
		t.Fatalf("globbing rule packages: %v", err)
	}
	if len(entries) == 0 {
		// A sweep with nothing to look at passes for the wrong reason, which is the same shape as
		// the defect it guards against.
		t.Fatal("found no rule packages, so this test examined nothing")
	}

	for _, entry := range entries {
		name := filepath.Base(entry)
		for _, suffix := range scratchPackageSuffixes {
			if !strings.HasSuffix(name, suffix) {
				continue
			}
			t.Errorf("internal/rules/%s looks like a throwaway package left behind; delete it file "+
				"by file before committing. A scratch copy that registers a rule name breaks "+
				"TestEveryRuleIsRegistered for every other agent in the tree, and the failure names "+
				"a package they have never heard of", name)
		}
	}
}
