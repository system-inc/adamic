package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The right of ??=, ||= and &&= is judged as the right of = is, by every rule that reads a view site
// (#6jq01hc). viewSite took only =, so on main 71869748 a Dog[] went into an Animal[] by each operator and
// a Cat pushed through it was read as a Dog (v02 to v04: Node throws, both backends print on), a readonly
// field went into a writable slot by ??= and a string write faulted natively under ASan (v06), and a
// value hiding a string b went into an optional numeric b by ??= and panicked natively where Node prints
// str1 (v07). Each now stops at the operator's right in its rule's words, as v01 and v09 always did with
// =. The controls carry an honest value through each operator and agree with Node on release native,
// sanitized native and the JavaScript backend. A stop's .stop file holds its position on the first line
// and words its message must carry on the second. Mutant: the logical operators dropped from viewSite
// (internal/lower/record-storage-roads.mutants.md), and v02 to v07 lower again.
func TestLogicalAssignmentViewsArePinned(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/logical_assignment_views"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, absolute)
			truth := onNode(t, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				position, words, _ := strings.Cut(stop, "\n")
				if err == nil {
					// What the write does once it lowers, so a mutant's red names the fault.
					sanitized, _ := natively(t, program)
					t.Fatalf("lowered, want a stop at %s: Node exit %d %q, sanitized native exit %d %q, stderr %q", position, truth.exitCode, truth.stdout, sanitized.exitCode, sanitized.stdout, sanitized.stderr)
				}
				got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), "")
				if !strings.HasPrefix(got, position+": ") || !strings.Contains(got, words) {
					t.Fatalf("stop: got %q, want one at %s carrying %q", got, position, words)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Fatalf("%s: %s: Node %q, %s %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 12 {
		t.Fatalf("found %d programs, want 12", programs)
	}
}
