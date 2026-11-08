package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Lowering builds these allocations outside expression, so graphAllocation never plans them, and
// in a graph program each was measured leaking a cycle through its result by the counted build, with
// a literal control freeing everything (#djwd1t2, main e3aa0bb7): toSorted 3 of 10 freed, toReversed 2
// of 9, Map.groupBy 5 of 15, a Record's values 6 of 13. Until #y7wexf5 plans them where lowering
// makes them, each stops in a graph program over elements a cycle could run through, and only then:
// the same operations over numbers lower and free everything.
//
// Mutant: drop the graphUnplanned check from graphAdoptionBoundary, and every stopped case lowers,
// and leaks by the counted build.
const graphUnplannedNode = `interface Node { name: string; next: Node[] }
const a: Node = { name: 'a' + '', next: [] };
const b: Node = { name: 'b' + '', next: [a] };
a.next.push(b);
`

func TestGraphUnplannedAllocations(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, source, stopped string }{
		{"toSorted", graphUnplannedNode + `a.next = a.next.toSorted((x, y) => (x.name < y.name ? -1 : 1));
console.log(` + "`${a.name} ${a.next.length}`" + `);
`, "toSorted in a graph program"},
		{"toReversed", graphUnplannedNode + `a.next = a.next.toReversed();
console.log(` + "`${a.name} ${a.next.length}`" + `);
`, "toReversed in a graph program"},
		{"Map.groupBy", `interface Node { name: string; next: Node[]; groups: Map<string, Node[]>[] }
const a: Node = { name: 'a' + '', next: [], groups: [] };
const b: Node = { name: 'b' + '', next: [a], groups: [] };
a.next.push(b);
a.groups.push(Map.groupBy(a.next, (node) => node.name));
console.log(` + "`${a.name} ${a.groups.length}`" + `);
`, "Map.groupBy in a graph program"},
		{"a record's values", graphUnplannedNode + `const record: Record<string, Node> = { b: b };
a.next = Object.values(record);
console.log(` + "`${a.name} ${a.next.length}`" + `);
`, "a record's values in a graph program"},
		// The same array and map operations over numbers, in the same graph program: nothing a cycle could
		// run through. A record over scalars isn't here, since record literals meet their own graph rule.
		{"over numbers", graphUnplannedNode + `const counts: number[] = [3, 1, 2];
const sorted = counts.toSorted((x, y) => x - y);
const reversed = counts.toReversed();
const grouped = Map.groupBy(counts, (n) => (n > 1 ? 'big' : 'small'));
console.log(` + "`${a.name} ${sorted.join()} ${reversed.join()} ${grouped.size}`" + `);
`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if c.stopped != "" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(notYet.What, c.stopped) {
					t.Fatalf("got %v, want NotYet naming %q", err, c.stopped)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			got, _ := natively(t, program)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("native: %s; Node %q; native %q; stderr %q", difference, truth.stdout, got.stdout, got.stderr)
			}
			counted := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), counted, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			ran := execute(t, counted)
			if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: ran.stdout, Stderr: ran.stderr, ExitCode: ran.exitCode}); report != "" {
				t.Fatal(report)
			}
		})
	}
}
