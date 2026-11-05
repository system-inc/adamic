package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A named container seen through a wider readonly slot is read as the slot holds its elements, while it's
// stored as its own type says (#3vsp1ds). invariant-mutable lets it in, since the slot can't write, and on
// main aea71e38 each n row lowered: a number[] through readonly (number | string)[] (n01, and n06 as an
// argument to such a parameter) and through a readonly field of that type (n02, and n04 by a spread that
// copies the field) segfaulted natively after printing undefined where Node printed 0 and 1, and a
// Map<string, number> through ReadonlyMap<string, number | string> (n03, and n05 in a union with string)
// printed "a = undefined" for Node's "a = 0". Each now stops in elementsHeldOtherwise's words, in .a and
// .ts. The controls lower and agree with Node on release, sanitized native and the JavaScript backend:
// numbers read as number | undefined (c01), whose NaN-packed undefined holds every number as itself, and
// strings (c02) and objects (c03) read as a union, whose box for a reference is the reference, and a
// never[] (c04, tsc's emptyArray) seen through two wider readonly slots, since it holds nothing to read.
// A destructuring declaration with no type of its own reads each element at the initializer's type, so
// the pattern's implied [any, any] is no slot: a tuple a call returns (c05, collections.a's shape) and a
// Map entry and an as const literal (c06, host_map_union.a's and gaps.a's) lower and agree. An annotated
// pattern is a slot its bindings read through, and n07, a number[] destructured as readonly
// (number | string)[], stops.
// k01, a number[] passed as readonly unknown[], keeps its erased-storage stop. Each program's whole stop
// is pinned in its .stop file, or "lowers". Mutant: elementsHeldOtherwise judging only fresh values again,
// and every n row lowers and goes red here.
func TestNamedContainerLayout(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/named_container_layout"
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
	if programs != 28 {
		t.Fatalf("found %d programs, want 28", programs)
	}
}
