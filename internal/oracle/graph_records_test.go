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

// A record is never a graph member: a literal makes it as ir.RecordLiteral, which nothing adopts. So
// a record on a cycle is a counted allocation inside it, and the cycle leaks. The write rule in
// cycles.go refuses a cycle closed by a write into a record (a.r['k'] = ...), but a record built whole
// and stored into a graph field closes one with no such write. a.r = { k: a.next.slice() } freed 5 of
// 20 by the counted build on main fd8b74b9, clean under ASan (#a9e21hc). It now stops, as a Promise
// on a cycle does, until #1vb0x5s gives records their region; a record off every cycle lowers.
//
// Mutant: drop the throughRecord stop from graphTypes (graph_types.go), and the first case lowers
// and leaks by the counted build.
const graphRecordNode = `interface Node { name: string; next: Node[]; r: Record<string, Node[]>; tags: Record<string, string> }
const a: Node = { name: 'a' + '', next: [], r: {}, tags: {} };
const b: Node = { name: 'b' + '', next: [a], r: {}, tags: {} };
a.next.push(b);
`

func TestGraphRecordOnACycle(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, source, stopped string }{
		{"a record literal stored into a graph field", graphRecordNode + `a.r = { k: a.next.slice() };
const got = a.r['k'];
console.log(` + "`${a.name} ${got === undefined ? -1 : got.length}`" + `);
`, "a record whose values can reach back to what holds it, in a graph program"},
		// The same cycle through an array field: its allocations are graph members, so it frees.
		{"the same cycle through an array", `interface Node { name: string; next: Node[]; kids: Node[][] }
const a: Node = { name: 'a' + '', next: [], kids: [] };
const b: Node = { name: 'b' + '', next: [a], kids: [] };
a.next.push(b);
a.kids = [a.next.slice()];
const got = a.kids[0];
console.log(` + "`${a.name} ${got === undefined ? -1 : got.length}`" + `);
`, ""},
		// A record of graph values that nothing in the graph holds is off every cycle: it holds b from
		// outside, as any counted holder does.
		{"a record of graph values off every cycle", `interface Node { name: string; next: Node[] }
const a: Node = { name: 'a' + '', next: [] };
const b: Node = { name: 'b' + '', next: [a] };
a.next.push(b);
const record: Record<string, Node> = { b: b };
const got = record['b'];
console.log(` + "`${a.name} ${got === undefined ? '-' : got.name}`" + `);
`, ""},
		// A graph object's record of strings: no value it holds leads back, so it's off every cycle.
		{"a graph object's record of strings", `interface Node { name: string; next: Node[]; tags: Record<string, string> }
const a: Node = { name: 'a' + '', next: [], tags: {} };
const b: Node = { name: 'b' + '', next: [a], tags: {} };
a.next.push(b);
a.tags = { kind: 'leaf' };
console.log(` + "`${a.name} ${a.tags['kind']}`" + `);
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
