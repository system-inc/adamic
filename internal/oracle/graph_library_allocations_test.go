package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// Lowering builds these allocations outside expression, and in a graph program each was measured
// leaking a cycle through its result by the counted build, with a literal control freeing everything
// (#djwd1t2, main e3aa0bb7): toSorted 3 of 10 freed, toReversed 2 of 9, Map.groupBy 5 of 15, a
// Record's values 6 of 13. graphFlows plans each where lowering makes it (#y7wexf5), so each, written
// inline into a graph member as it was measured, lowers, prints what Node prints and frees everything
// by the counted build. The controls beside them, the same operations over numbers and a record of
// string arrays as tsc reads one, never held anything a cycle could run through.
//
// Mutant: make ArraySort's, ArrayReverse's or Coalesce's arm silent, or mark RecordCall can't hold
// (internal/lower/graph-flow-exhaustive.mutants.md 1, 2, 3 and 5), and its case leaks by the counted
// build.
const graphLibraryNode = `interface Node { name: string; next: Node[] }
const a: Node = { name: 'a' + '', next: [] };
const b: Node = { name: 'b' + '', next: [a] };
a.next.push(b);
`

func TestGraphLibraryAllocations(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, source string }{
		{"toSorted", graphLibraryNode + `a.next = a.next.toSorted((x, y) => (x.name < y.name ? -1 : 1));
console.log(` + "`${a.name} ${a.next.length}`" + `);
`},
		{"toReversed", graphLibraryNode + `a.next = a.next.toReversed();
console.log(` + "`${a.name} ${a.next.length}`" + `);
`},
		{"Map.groupBy", `interface Node { name: string; next: Node[]; groups: Map<string, Node[]>[] }
const a: Node = { name: 'a' + '', next: [], groups: [] };
const b: Node = { name: 'b' + '', next: [a], groups: [] };
a.next.push(b);
a.groups.push(Map.groupBy(a.next, (node) => node.name));
console.log(` + "`${a.name} ${a.groups.length}`" + `);
`},
		{"a record's values", graphLibraryNode + `const record: Record<string, Node> = { b: b };
a.next = Object.values(record);
console.log(` + "`${a.name} ${a.next.length}`" + `);
`},
		// tsc's own shape (executeCommandLine.ts:411): a record of string arrays read by Object.entries,
		// in a graph program. A string[] holds only strings, so nothing a cycle could run through.
		{"a record of string arrays", graphLibraryNode + `const inverted: { [value: string]: string[] } = {};
inverted['es6'] = ['es2015', 'es6'];
const joined = Object.entries(inverted).map(([, synonyms]) => synonyms.join('/'));
console.log(` + "`${a.name} ${joined.join()}`" + `);
`},
		// The same array and map operations over numbers, in the same graph program: nothing a cycle could
		// run through.
		{"over numbers", graphLibraryNode + `const counts: number[] = [3, 1, 2];
const sorted = counts.toSorted((x, y) => x - y);
const reversed = counts.toReversed();
const grouped = Map.groupBy(counts, (n) => (n > 1 ? 'big' : 'small'));
console.log(` + "`${a.name} ${sorted.join()} ${reversed.join()} ${grouped.size}`" + `);
`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
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
			if ran.exitCode != truth.exitCode || string(ran.stdout) != string(truth.stdout) {
				t.Fatalf("counted disagrees with Node: %d %s %s", ran.exitCode, ran.stdout, ran.stderr)
			}
			if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: ran.stdout, Stderr: ran.stderr, ExitCode: ran.exitCode}); report != "" {
				t.Fatal(report)
			}
		})
	}
}
