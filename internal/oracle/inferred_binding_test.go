package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An inferred binding whose type parameter is nested in its parameter's type is judged against the
// constraint as itself, and the argument against its parameter at the resolved binding, by the argument
// site (#c4vhbep). combine is tsc's core.ts:966 shape standalone (append's overloads, tsc's isArray),
// with T a number, a string and an object. add is Compiler's, add<T>(to: T[], value: T) handed a Dog[]
// and an Animal, T inferred and written. p01 to p03 are Language's: a T nested beside { kind: string }
// in an array, a Map's value and a field, which the body may honestly write { kind: "other" } into, and
// p04 is her control, keepFirst<T>(to: T[]) pushing its own to[0] back. add_structural (add over
// interfaces, so no nominal rule) and p05 (p01 with T unconstrained, so no binding is judged) are the
// rows only the argument site can refuse. Each program's whole stop is
// pinned in its .stop file, or "lowers", and a lowering one is held to Node on release, sanitized native
// and JavaScript. MUTANTS.
func TestInferredBindingNestedTypeParameter(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/inferred_binding"
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
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			program, err := lowered(t, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					t.Fatalf("lowered, want %q", stop)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
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
	if programs != 22 {
		t.Fatalf("found %d programs, want 22", programs)
	}
}
