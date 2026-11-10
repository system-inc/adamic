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

// A Map's [key, value] pair is made in the runtime, outside graphAllocation: by an entries()
// iterator's next (map_set.c's collection_next), which every road here takes, and by map.c's
// adamic_map_entries. Before #dps0kp6, next never adopted a pair, and map.c only when the map was a
// graph member, so a pair held by a graph slot on a cycle stayed counted and the cycle leaked (e2
// freed 9 of 20 by the counted build on main 32e7abb8, e4 10 of 20, each matching Node and clean
// under ASan). Each pair is now planned by its own tuple type (pairGraphTypes, and the iterator's
// planned state), and a pair a graph program could hold as some other graph type, which no plan
// names, stops (unplannedPair).
//
// Mutants, each red here by the counted build or by the stop's absence:
//   - collection_next ignores planned: "e2", "e4", "for-of" and "a labeled tuple" leak (Codex 01a1c4da).
//   - pairGraphTypes returns nil for a named pair: the same four leak.
//
// unplannedPair is a fail-closed backstop with no witness: every road tried (a labeled tuple, a
// readonly tuple, a wider array, refused first as invariant-mutable) has the pair's own type
// selected, so the pair is planned. Each of those cases here holds either outcome sound.
const mapEntriesPairsNode = `interface Node { name: string; next: Node[]; pairs: [string, Node][] }
const a: Node = { name: 'a' + '', next: [], pairs: [] };
const b: Node = { name: 'b' + '', next: [a], pairs: [] };
a.next.push(b); // a cycle: the program is a graph program
const map = new Map<string, Node>();
map.set('b', b);
`

const mapEntriesPairsPrint = `const pair = a.pairs[0];
if (pair !== undefined) console.log(` + "`${a.pairs.length} ${pair[0]} ${pair[1].name} ${pair[1].next[0] === a}`" + `);
`

func TestMapEntriesPairs(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, source, stopped string }{
		// e2: the entries array, held by a; each pair holds b, b holds a.
		{"e2", mapEntriesPairsNode + "a.pairs = [...map.entries()];\n" + mapEntriesPairsPrint, ""},
		// e4: one pair stored alone, the array dropped.
		{"e4", mapEntriesPairsNode + "const first = [...map.entries()][0];\nif (first !== undefined) a.pairs.push(first);\n" + mapEntriesPairsPrint, ""},
		// The for-of road (iterated): each pair a for-of gives, stored.
		{"for-of", mapEntriesPairsNode + "for (const entry of map) a.pairs.push(entry);\n" + mapEntriesPairsPrint, ""},
		// The control: the same pair as a literal.
		{"literal", mapEntriesPairsNode + "a.pairs = [['b', b]];\n" + mapEntriesPairsPrint, ""},
		// A for-of over the same map whose pairs no slot holds: they're planned all the same, and free.
		{"read only", mapEntriesPairsNode + "let names = '';\nfor (const [name, node] of map) names += name + node.name;\nconsole.log(names);\n", ""},
		// A pair held as a readonly tuple, a different type: the cycle finder relates the pair's own
		// allocation identity into the graph, so the pair is planned and balances (Codex 01a1c4da). A wider
		// array is no road: a tuple is an object, so an array view of it is refused first.
		{"a readonly tuple", `interface Node { name: string; next: Node[]; pairs: (readonly [string, Node])[] }
const a: Node = { name: 'a' + '', next: [], pairs: [] };
const b: Node = { name: 'b' + '', next: [a], pairs: [] };
a.next.push(b);
const map = new Map<string, Node>();
map.set('b', b);
a.pairs = [...map.entries()];
console.log(String(a.pairs.length));
`, "either"},
		// A pair held as a labeled tuple. The checker gives it [string, Node]'s identity today, so the pair
		// is planned and must balance (Codex 01a1c4da); if that changes, it stops as the readonly tuple does.
		// Either is sound.
		{"a labeled tuple", `interface Node { name: string; next: Node[]; pairs: [key: string, value: Node][] }
const a: Node = { name: 'a' + '', next: [], pairs: [] };
const b: Node = { name: 'b' + '', next: [a], pairs: [] };
a.next.push(b);
const map = new Map<string, Node>();
map.set('b', b);
a.pairs = [...map.entries()];
console.log(String(a.pairs.length));
`, "either"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if c.stopped == "either" {
				var notYet *lower.NotYet
				if errors.As(err, &notYet) {
					if !strings.Contains(notYet.What, "pair held as another graph type") {
						t.Fatalf("got %v, want the pair's stop or a balanced program", err)
					}
					t.Logf("stopped: %s", notYet.What)
					return
				}
				t.Log("lowered: its pairs are planned")
			} else if c.stopped != "" {
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
