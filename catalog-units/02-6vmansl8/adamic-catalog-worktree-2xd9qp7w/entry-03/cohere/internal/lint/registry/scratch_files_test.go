package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scratchFilePrefixes are the names an agent gives a throwaway test file.
//
// `zz_` is the dominant one and it is not arbitrary: Go runs tests in file order within a package,
// so a leading `zz` sorts a probe last, and an agent measuring something wants its probe to run
// after the real fixtures rather than among them. Several arrived at that independently in one
// night.
var scratchFilePrefixes = []string{"zz_", "zzz"}

// scratchFileInfixes are the names a throwaway file carries when it is not prefixed.
//
// Narrow on purpose, and narrower than the first version of this list. That version included
// `scratch`, and the first run flagged this file and its sibling guard, both of which are named for
// the thing they refuse. A guard that cannot survive its own name is measuring the wrong property,
// so the word is gone rather than exempted: an exemption would have hidden a real over-match behind
// a special case, and the next file named for what it guards would have tripped it again.
//
// What remains matches a measurement rather than a subject. A real test named for what it does will
// not contain `probe`, `counterfactual`, or `_measure_`, and if one ever legitimately does, an
// exemption below is a decision somebody makes on purpose.
var scratchFileInfixes = []string{"probe", "_measure_", "counterfactual"}

// scratchFileExemptions are shipped packages whose names would otherwise trip this.
//
// `internal/astprobe` is a real package: it probes the abstract syntax tree on purpose and ships.
// It is listed rather than pattern-excluded so that adding a second one is visible in a diff.
var scratchFileExemptions = []string{"internal/astprobe"}

// TestNoScratchFilesRemain refuses a throwaway test file left inside a shipped package.
//
// This is the file-shaped sibling of TestNoScratchPackagesRemain, and it exists because that guard
// cannot see this violation. It globs directories under `internal/rules`, so it is blind twice over:
// to a stray file inside an otherwise real package, which is the more common shape by far, and to
// every tree outside `internal/rules` -- `internal/utilities/hir` held five scratch files in one night
// and no guard could have named them.
//
// A directory-shaped check cannot see a file-shaped violation. That sentence is the whole reason
// this file exists, and it came from a sibling domain noticing the same class in its own tree.
//
// The cost is smaller than the package case but real. A `zz_` probe carries an agent's private
// measurement helpers, compiles into the package, and runs on every `go test ./...` any other agent
// runs. It is not a failure that names someone else's work, which is what made the package version
// urgent; it is clutter that reads as coverage. A reader counting tests in a package cannot tell a
// probe from a fixture without opening it.
//
// Measured when this was written: one match across `internal`, `tools` and `cmd`, and that match is
// the exempted shipped package. The control is 343 real test files found by the same walk, so a
// clean result here is a measurement rather than a search that examined nothing.
func TestNoScratchFilesRemain(t *testing.T) {
	t.Parallel()
	// Skipped unless asked for, and that is a deliberate weakening rather than an oversight.
	//
	// The first version ran always, and within a minute of being written it went red on two live
	// probes belonging to an agent that was mid-measurement in `internal/utilities/hir`. Those files
	// genuinely should not ship, so the finding was true -- and a guard that is true and fires
	// during ordinary work is a guard people learn to scroll past. The package-shaped sibling avoids
	// this only by accident: scratch packages are rare, scratch files are the common shape, so the
	// same design is quiet there and deafening here.
	//
	// So it runs where it is actionable: before a commit, when a probe left behind is about to
	// become permanent. `COHERE_CHECK_SCRATCH=1 go test ./internal/registry/` is the invocation, and
	// the commit path sets it. Everywhere else it says why it did not run, so a reader never mistakes
	// a skip for a pass.
	if os.Getenv("COHERE_CHECK_SCRATCH") == "" {
		t.Skip("set COHERE_CHECK_SCRATCH=1 to run; skipped during ordinary work because a live " +
			"agent's probe is a legitimate mid-measurement file, and a guard that fires on normal " +
			"work stops being read")
	}

	roots := []string{"../../../internal", "../../../command"}

	examined := 0
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".go") {
				return nil
			}
			examined++

			normalized := filepath.ToSlash(path)
			for _, exemption := range scratchFileExemptions {
				if strings.Contains(normalized, exemption) {
					return nil
				}
			}

			flagged := false
			for _, prefix := range scratchFilePrefixes {
				if strings.HasPrefix(name, prefix) {
					flagged = true
				}
			}
			for _, infix := range scratchFileInfixes {
				if strings.Contains(name, infix) {
					flagged = true
				}
			}
			if !flagged {
				return nil
			}

			t.Errorf("%s looks like a throwaway probe left inside a shipped package; delete it by "+
				"name before committing. The package guard cannot see this: it globs directories "+
				"under internal/rules, so a stray file in a real package is invisible to it and so "+
				"is every tree outside internal/rules", normalized)
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", root, err)
		}
	}

	// A walk that found nothing passes for the wrong reason, which is the same shape as the defect
	// it guards against. The floor is deliberately far below the real count so ordinary growth or
	// deletion never trips it.
	if examined < 200 {
		t.Fatalf("examined only %d Go files across internal, tools and cmd, so this test looked at "+
			"almost nothing and its silence means nothing", examined)
	}
}
