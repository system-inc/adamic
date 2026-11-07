package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A container made right there goes only where its elements are read as they're stored (#b8fwyfx). On
// main ab8a85be each s row lowered and segfaulted natively (ASan's SEGV on unknown address) where Node
// printed the elements, most of them printing undefined first: [Kind.A, Kind.B] into "string" |
// (number | string)[] through an object literal's field (s01), a declaration (s02), an assignment (s03),
// an argument (s04), a return (s05), and [[0, 1]] into (number | string)[][] | string (s06); and, in a
// plain slot, a copy (s07), a new Map of written-out entries (s08), of Object.entries (s09, tsc's
// commandLineParser shape, read only, where native printed "a = undefined" for Node's "a = 0") and of
// named entries (s10). s11 is [0, 1] into string | number[] | (number | string)[]: the first array member
// takes it as it's stored, the second doesn't, and narrowing can pick either. Each now stops in
// elementsHeldOtherwise's words, in .a and .ts; a later fix that lowers one must agree with Node to pass.
// The k rows stopped on main and still stop: an element of a union-element array (k01) and an object
// member's array field (k02) by object.go's narrower-literal guard, a copy, a new Map and a new Set into
// a union slot by invariant-mutable (k03, k06, k07), and Array.from (k04). k08 to k12 hold the roads
// where a container meets a slot member that reads elements without being a container, which
// elementsHeldOtherwise doesn't walk (Compiler's question): a literal into Iterable<number | string>, into
// ArrayLike<number | string>, and into "s" | Iterable<number | string>, a copy into an Iterable and a new
// Map<string, number> into an Iterable of pairs. Each stops before the judgment, at shape_fit's plain
// object stop or at ArrayLike's own, and its pin goes red the day one lowers, so that lift meets this
// judgment first; a new Set<number> into a
// Set<number | string> (k05) now stops at the judgment, before the Set's own element stop. The controls
// lower and agree with Node on release, sanitized native and the JavaScript backend: a mixed literal into
// the union slot (c01) and a plain one (c02), a number literal into "string" | number[] (c03), numbers
// read as number | undefined from a literal in a union slot (c04) and a plain one (c05), and a new Map of
// an enum's values into Map<string, number> (c06). Each program's whole stop is pinned in its .stop file,
// or "lowers". Mutant: elementsHeldOtherwise judging nothing turns every s row back into the native crash
// and red here.
func TestUnionSlotLiteralLayout(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/union_slot_literal_layout"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		test.Fatal(err)
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
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				test.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				test.Fatal(err)
			}
			truth := onNode(test, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				test.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			program, err := lowered(test, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					test.Fatalf("lowered, want %q", stop)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					test.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				test.Fatalf("Lower: %v", err)
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
	if programs != 58 {
		test.Fatalf("found %d programs, want 58", programs)
	}
}
