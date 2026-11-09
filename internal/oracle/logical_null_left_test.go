package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// left && x keeps a falsy left, and where the result is a union and the left is a bare reference
// whose type holds null and not undefined, that left is the null pointer a union reads as undefined
// (#ayvxxpc). So `none && none.ready` over a null Item | null answered === null false natively where Node
// answers true (a2, a3), and a read narrowed to boolean past it panicked (a1, a5, a9). Its missing arm is
// now the boxed null. a6 is the present left. b1 is a number result and b3 an array left, both live at
// the parent too. Each lowers and agrees with Node on release,
// sanitized native and the JavaScript backend. Mutant (internal/lower/logical-null-left.mutants.md):
// logicalNullLeft answers nil, and a1 to a5 and a9 disagree again.
func TestLogicalNullLefts(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/logical_null_left"
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
	if programs != 16 {
		t.Fatalf("found %d programs, want 16", programs)
	}
}
