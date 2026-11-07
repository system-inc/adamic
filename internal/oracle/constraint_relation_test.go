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
// Options, already was. r1, Language's readonly road (reset<T extends { n: number }> writing a Holder's
// readonly n), is refused at the argument in invariant-mutable's words, as its direct twin r2 is. So are
// Language's attacks on the head: a default type argument (a1, at the new that binds it), a constraint
// that is itself a type parameter, inferred (a2) and written (a2b, at the type argument), a class's type
// argument written (a3), inferred (a3b) and in an extends clause (a3c), an explicit
// read<{ readonly other: string }>(narrow) (a4), an in-place callback returning the hidden value, by its
// expression body (a5) or a block (a5b), and the hidden value inside a literal argument, refused at its
// property (a6). On the parent each of these, but h2 and r2, lowered: Node printed the lie, and r1 wrote
// the readonly field on every leg. The controls lower and agree with Node on release, sanitized native
// and the JavaScript backend: a fresh literal (c1), one with a property the constraint doesn't declare
// (c5), a class constructed as its type (c2), Options itself, inferred and written (c3), and a type-guard
// filter (c4). l1 backs the library
// exemption: lib's filter<S extends T> binds S to { readonly kind: "circle" }, which hides Circle's
// radius, and it agrees with Node natively. Each program's whole stop is pinned in its .stop file, or
// "lowers". Mutant: constraintRelationRefusal judging nothing turns h1 back into the native compiler-bug
// panic, and every refused row here but h2 and r2 red.
func TestConstraintRelationSites(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/constraint_relation_sites"
	entries, readDirectoryError := os.ReadDir(filepath.Join(repository, directory))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			expected, readError := os.ReadFile(absolute + ".stop")
			if readError != nil {
				test.Fatal(readError)
			}
			truth := onNode(test, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				test.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			program, loweredError := lowered(test, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if loweredError == nil {
					test.Fatalf("lowered, want %q", stop)
				}
				if got := strings.ReplaceAll(
					loweredError.Error(),
					strings.TrimSuffix(absolute, path),
					"",
				); got != stop {
					test.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if loweredError != nil {
				test.Fatalf("Lower: %v", loweredError)
			}
			sanitized, _ := natively(test, program)
			for _, backend := range []struct {
				name string
				got  run
			}{
				{"release", released(test, program)},
				{"sanitized", sanitized},
				{"JavaScript", onJavaScriptBackend(test, program)},
			} {
				if difference := disagreement(truth, backend.got); difference != "" {
					test.Fatalf(
						"%s: %s: Node %q, %s %q",
						backend.name,
						difference,
						truth.stdout,
						backend.name,
						backend.got.stdout,
					)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 44 {
		test.Fatalf("found %d programs, want 44", programs)
	}
}
