package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Binding a type argument to a constrained type parameter is a relation site (#xywvn06, Adamic's
// ruling): read<T extends Options>(options: T) can't be handed a { readonly other: string } that hides
// overrideImportMode: "lie", since the body reads it as Options' number. Language's h1 (the generic
// read), h3 (the same field in arithmetic) and h4 (the hidden field a class's) are refused in .a and .ts
// in the hiding rule's own words, at the argument, where h2, their direct twin through a parameter typed
// Options, already was. So are Language's attacks on the head: a default type argument (a1, at the new
// that binds it), a constraint that is itself a type parameter, inferred (a2) and written (a2b, at the
// type argument), a class's type argument written (a3), inferred (a3b) and in an extends clause (a3c),
// an explicit read<{ readonly other: string }>(narrow) (a4), and a callback returning the hidden value,
// by its expression body (a5) or a block (a5b). On the parent each of these, but h2, lowered and exited
// 70 natively with "compiler bug: a numeric field holds a reference" where Node prints the lie. The
// controls lower and agree with Node on release, sanitized native and the JavaScript backend: a fresh
// literal (c1), a class constructed as its type (c2), Options itself, inferred and written (c3), and a
// type-guard filter, whose library binding is left to the rules at its argument (c4). Each program's
// whole stop is pinned in its .stop file, or "lowers". Mutant: constraintRelationRefusal judging
// nothing turns h1 back into the native compiler-bug panic, and every refused row here red.
func TestConstraintRelationSites(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/constraint_relation_sites"
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
	if programs != 34 {
		t.Fatalf("found %d programs, want 34", programs)
	}
}
