package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An array literal a binding is inferred from, bound to a constraint that declares nothing to read
// through ({}, object or unknown), is judged whole as its type (judgeBoundArray, #x856a4v): the
// constraint has no element type, and needs none, since nothing read through it can hide a field or
// widen a slot through an element. e01 is tsc's builder.ts append(x, [fileId, ...]) over core.ts's
// append<T extends {}> overloads, a [number, string] into a [number, string][] | undefined; e02 a fresh
// object literal element under the same {}; e04 an unknown constraint; e05 an object constraint. Each
// lowers and agrees with Node on release, sanitized native and JavaScript. e03 is a constraint that
// declares something, { readonly 0: Options }, handed [narrow], whose element hides Options'
// overrideImportMode, which the body reads: it keeps the fail-closed stop. Each program's stop is pinned
// in its .stop file as generic_body_rigid's are (pinnedStop), or "lowers".
// internal/lower/bound-array-empty-constraint.mutants.md has the mutant.
func TestBoundArrayEmptyConstraint(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/bound_array_empty_constraint"
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
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					t.Fatalf("%s lowered, want %q", entry.Name(), stop)
				}
				if got := pinnedStop(err, strings.TrimSuffix(absolute, path)); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %s", pinnedStop(err, strings.TrimSuffix(absolute, path)))
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("%s source Node: exit %d stderr %q", entry.Name(), truth.exitCode, truth.stderr)
			}
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Fatalf("%s %s: %s: Node %q, %s %q", entry.Name(), backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 10 {
		t.Fatalf("found %d programs, want 10", programs)
	}
}
