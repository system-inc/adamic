package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A closed range type ({ pos; end }) is held as a Union in a program with the NodeArray layout (#hae423v), and a
// NodeArray going where one is held is served: pos and end are read, written, written when optional and
// destructured through it by a kind dispatch, and agree with Node natively, sanitized and on JavaScript. The
// programs in testdata/range_unions send a NodeArray and an object through the same function (a Node alone in
// node_through_range.a), write through tsc's upcast in setTextRangeEnd (upcast_write.a), and read a class with
// get pos() through the dispatch's object arm, which runs the getter (getter_through_range.a, Lower's Must 3).
//
// Mutants (scratch-ha1/): the object arm built as a dynamic read finishAccessors never sees, which turns
// getter_through_range.a red; the array arm's narrow made unchecked, which checkRangeUnionReads stops at
// lowering, turning every program here red.
func TestRangeUnionsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/range_unions"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, directory+"/"+entry.Name())
		})
	}
	if programs != 6 {
		t.Fatalf("%d programs in %s, want 6: a program added or removed needs this count changed with it", programs, directory)
	}
}

// What the dispatch doesn't serve still stops: a NodeArray going where a type with a member beyond pos and end is
// held, which is held as an object, stops at shapeFit as any array there does; and an array interface without
// the NodeArray layout, the nearest a plain array comes to a { pos; end } slot (a plain array has no pos and
// doesn't type-check there), stops where its type is held, before any slot takes it.
func TestRangeUnionsStopWhereNotServed(t *testing.T) {
	t.Parallel()
	const header = "interface TextRange { pos: number; end: number; }\ninterface Node { kind: number; }\ninterface NodeArray<T extends Node> extends Array<T>, TextRange { hasTrailingComma: boolean; transformFlags: number; }\n"
	const nodes = "const nodes = [{ kind: 1 }] as NodeArray<Node>;\nnodes.pos = 0;\nnodes.end = 2;\nnodes.hasTrailingComma = true;\nnodes.transformFlags = 0;\n"
	for _, probe := range []struct {
		name, source, stops string
		runs                bool
	}{
		{"wide_target", header + "interface Wide { pos: number; end: number; hasTrailingComma: boolean; }\nfunction flag(range: Wide): boolean { return range.hasTrailingComma; }\n" + nodes + "console.log(String(flag(nodes)));\n", "an array going where a plain object type is held", true},
		{"array_without_layout", header + "interface Spans extends Array<Node>, TextRange {}\nfunction width(range: TextRange): number { return range.end - range.pos; }\nfunction measure(spans: Spans): number { return width(spans); }\nconsole.log('ok');\n", "Spans", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), probe.name+".a")
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			if probe.runs {
				if truth := onNode(t, path); truth.exitCode != 0 {
					t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
				}
			}
			_, err := lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.stops) {
				t.Fatalf("want it NotYet naming %q, got %v", probe.stops, err)
			}
		})
	}
}
