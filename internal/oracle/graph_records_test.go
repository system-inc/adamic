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
// 20 by the counted build on main fd8b74b9, clean under ASan (#a9e21hc). A record written into a graph
// slot on a cycle now stops, until #1vb0x5s gives records their region. A record only near a cycle's
// types, never written into a slot on it, lowers: method_bind.a's dictionary is the control in the
// corpus, beside the three here.
//
// Mutants: make noteRecordSlot (internal/lower/graph_records.go) return at once, and the field and
// element cases lower and leak by the counted build; drop the reassigned cell's stop in graphTypes, and
// the captured-variable case lowers. Make reachesRecord return false at once, and the record one field
// deep and the record a closure captures lower and leak by the counted build (#t5q4prw).
const graphRecordNode = `interface Node { name: string; next: Node[]; r: Record<string, Node[]>; rs: Record<string, Node[]>[] }
const a: Node = { name: 'a' + '', next: [], r: {}, rs: [] };
const b: Node = { name: 'b' + '', next: [a], r: {}, rs: [] };
a.next.push(b);
`

func TestGraphRecordOnACycle(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, source, stopped string }{
		{"a record literal stored into a graph field", graphRecordNode + `a.r = { k: a.next.slice() };
const got = a.r['k'];
console.log(` + "`${a.name} ${got === undefined ? -1 : got.length}`" + `);
`, "a record written into a graph slot"},
		{"a record literal pushed into a graph array", graphRecordNode + `a.rs.push({ k: a.next.slice() });
console.log(` + "`${a.name} ${a.rs.length}`" + `);
`, "a record written into a graph slot"},
		// A captured variable assigned again with a record holding the closure that captures it: the cell
		// holds the record, the record the closure, the closure the cell. It's a function's local, since a
		// top-level variable is a global, whose cell never seeds.
		{"a reassigned captured variable holding a record", graphRecordNode + `function run(): number {
  let table: Record<string, () => number> = {};
  const size = (): number => (table['size'] === undefined ? 0 : 1);
  table = { size: size };
  return size();
}
console.log(` + "`${a.name} ${run()}`" + `);
`, "a record held by 'table'"},
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
		// A record one field deep on a cycle, which the slot seed doesn't see: the node holds a box, the box
		// the record, the record the node. 22 of 102 freed on main 8e3a784a (#t5q4prw).
		{"a record inside an object field on a cycle", `interface Box { r: Record<string, Node> }
interface Node { name: string; box: Box }
function make(name: string): Node { return { name, box: { r: {} } }; }
let total = 0;
for (let i = 0; i < 10; i++) {
  const a = make(` + "`a${i}`" + `);
  a.box = { r: { k: a } };
  total += Object.keys(a.box.r).length;
}
console.log(` + "`${total}`" + `);
`, "a record written into a graph slot"},
		// A record a closure on the node captures: the node holds the closure, the closure the record's
		// cell, the record the node. 22 of 92 freed on main 8e3a784a (#t5q4prw).
		{"a record a closure on the node captures", `interface Node { name: string; count: () => number }
function make(name: string): Node { return { name, count: () => 0 }; }
let total = 0;
for (let i = 0; i < 10; i++) {
  const a = make(` + "`a${i}`" + `);
  const rec: Record<string, Node> = { k: a };
  a.count = () => Object.keys(rec).length;
  total += a.count();
}
console.log(` + "`${total}`" + `);
`, "a record written into a graph slot"},
		// The record reaches the seed through a binding typed by a view that omits it: the written
		// expression's own type is { tag: number }, and only the shapes seen as it hold r (Runtime's road
		// on #t5q4prw). reachesRecord follows them, so the write stops.
		{"a record behind a view at the seed", `interface Box { tag: number }
interface Node { name: string; box: Box }
function make(name: string): Node { return { name, box: { tag: 0 } }; }
let total = 0;
for (let i = 0; i < 10; i++) {
  const a = make(` + "`a${i}`" + `);
  const rec: Record<string, Node> = { k: a };
  const full = { tag: 1, r: rec };
  const hidden: Box = full;
  a.box = hidden;
  total += a.box.tag;
}
console.log(` + "`${total}`" + `);
`, "a record written into a graph slot"},
		// The two shapes above with no record: each cycle is the graph plan's, and frees whole.
		{"the field shape with no record", `interface Box { n: Node | undefined }
interface Node { name: string; box: Box }
function make(name: string): Node { return { name, box: { n: undefined } }; }
let total = 0;
for (let i = 0; i < 10; i++) {
  const a = make(` + "`a${i}`" + `);
  a.box = { n: a };
  total += a.box.n === a ? 1 : 0;
}
console.log(` + "`${total}`" + `);
`, ""},
		{"the capture shape with no record", `interface Node { name: string; count: () => number }
function make(name: string): Node { return { name, count: () => 0 }; }
let total = 0;
for (let i = 0; i < 10; i++) {
  const a = make(` + "`a${i}`" + `);
  const held: Node[] = [a];
  a.count = () => held.length;
  total += a.count();
}
console.log(` + "`${total}`" + `);
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
