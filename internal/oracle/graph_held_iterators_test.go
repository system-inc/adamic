package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// Row 13 of #bvsjdte's inventory: a collection's library iterator held in a graph slot on a cycle,
// the iterator holding its collection and the collection holding the graph back. The iterator's
// state, next closure and object are adopted when the collection is a graph member (map_set.c's
// adamic_collection_iterator), and the re-probes freed everything by the counted build, matching
// Node, with an object-literal control beside them (#djwd1t2, Codex 01a1c4c5-3c8e on main 32e7abb8:
// 18 of 18 for each). These pin it, held to Node and balanced by the counted build.
//
// Mutant: map_set.c adopts none of the iterator's three allocations when its collection is a graph
// member, and "a Map's values()" and "a Set's values()" leak by the counted build.
func TestGraphHeldIterators(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, source string }{
		{"a Map's values()", `interface Node { name: string; next: Node[]; iterators: MapIterator<Node>[] }
const a: Node = { name: 'a' + '', next: [], iterators: [] };
const b: Node = { name: 'b' + '', next: [a], iterators: [] };
a.next.push(b); // a cycle: the program is a graph program
const map = new Map<string, Node>([[a.name, a]]);
a.iterators.push(map.values()); // a holds the iterator, the iterator holds map, map holds a
const got = map.get('a');
console.log(` + "`${a.name} ${a.iterators.length} ${map.size} ${got === undefined ? -1 : got.iterators.length}`" + `);
`},
		{"a Set's values()", `interface Node { name: string; next: Node[]; iterators: SetIterator<Node>[] }
const a: Node = { name: 'a' + '', next: [], iterators: [] };
const b: Node = { name: 'b' + '', next: [a], iterators: [] };
a.next.push(b); // a cycle: the program is a graph program
const set = new Set<Node>([a]);
a.iterators.push(set.values());
console.log(` + "`${a.name} ${a.iterators.length} ${set.size}`" + `);
`},
		{"the control", `interface Node { name: string; next: Node[]; iterators: { map: Map<string, Node> }[] }
const a: Node = { name: 'a' + '', next: [], iterators: [] };
const b: Node = { name: 'b' + '', next: [a], iterators: [] };
a.next.push(b); // a cycle: the program is a graph program
const map = new Map<string, Node>([[a.name, a]]);
a.iterators.push({ map: map });
const got = map.get('a');
console.log(` + "`${a.name} ${a.iterators.length} ${map.size} ${got === undefined ? -1 : got.iterators.length}`" + `);
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
			if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: ran.stdout, Stderr: ran.stderr, ExitCode: ran.exitCode}); report != "" {
				t.Fatal(report)
			}
		})
	}
}
