package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// A record on a cycle joins its graph region (#1vb0x5s). Before, a record was never a graph member, so a
// record on a cycle was a counted allocation inside it and the cycle leaked: a.r = { k: a.next.slice() }
// freed 5 of 20 by the counted build on main fd8b74b9, clean under ASan (#a9e21hc), and a record written
// into a graph slot on a cycle stopped. Now graphTypes selects a record type on a cycle, or one holding
// graph values, its literals are made in its region (adamic_record_new_graph), and every write into it is
// seeded where it's made, its value taken into the region by the runtime. So each road here that stopped
// compiles, agrees with Node and balances by the counted build, and the clean twins stay as they were. A
// record of a type the graph doesn't select, written into a graph slot on a cycle, still stops
// (noteRecordSlot); a record of loose values never is one (internal/lower's
// TestGraphLooseRecordsAreNeverSelected). The mutants are graphRecordMutants, below.
const graphRecordNode = `interface Node { name: string; next: Node[]; r: Record<string, Node[]>; rs: Record<string, Node[]>[] }
const a: Node = { name: 'a' + '', next: [], r: {}, rs: [] };
const b: Node = { name: 'b' + '', next: [a], r: {}, rs: [] };
a.next.push(b);
`

// graphRecordTable is a cycle through a record field, whose writes the roads below make.
const graphRecordTable = `interface Node { name: string; next: Node[]; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], r: {} };
const b: Node = { name: 'b' + '', next: [a], r: {} };
a.next.push(b);
`

// graphRecordCaught is how a road's program writes a TypeError it caught.
const graphRecordCaught = `catch(error) {
    if(error instanceof Error) console.log(` + "`${error.name}: ${error.message}`" + `);
}
`

// graphRecordRoads are TestGraphRecordOnACycle's rows: a road that stops names its NotYet, and every
// other compiles, agrees with Node and balances by the counted build.
var graphRecordRoads = []struct{ name, source, stopped string }{
	{"a record literal stored into a graph field", graphRecordNode + `a.r = { k: a.next.slice() };
const got = a.r['k'];
console.log(` + "`${a.name} ${got === undefined ? -1 : got.length}`" + `);
`, ""},
	{"a record literal pushed into a graph array", graphRecordNode + `a.rs.push({ k: a.next.slice() });
console.log(` + "`${a.name} ${a.rs.length}`" + `);
`, ""},
	// A captured variable assigned again with a record holding the closure that captures it: the cell
	// holds the record, the record the closure, the closure the cell. It's a function's local, since a
	// top-level variable is a global, whose cell never seeds. The record's type is on the cell's cycle,
	// so it's selected, and the cell's stop doesn't fire.
	{"a reassigned captured variable holding a record", `interface Node { name: string; next: Node[] }
const a: Node = { name: 'a' + '', next: [] };
const b: Node = { name: 'b' + '', next: [a] };
a.next.push(b);
function run(): number {
  let table: Record<string, () => number> = {};
  const size = (): number => (table['size'] === undefined ? 0 : 1);
  table = { size: size };
  return size();
}
console.log(` + "`${a.name} ${run()}`" + `);
`, ""},
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
	// the record, the record the node. 22 of 102 freed on main 8e3a784a (#t5q4prw); the record's type is
	// on the cycle, so it's a graph member now.
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
`, ""},
	// A record a closure on the node captures: the node holds the closure, the closure the record's
	// cell, the record the node. 22 of 92 freed on main 8e3a784a (#t5q4prw); a graph member now.
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
`, ""},
	// The record reaches the seed through a binding typed by a view that omits it: the written
	// expression's own type is { tag: number }, and only the shapes seen as it hold r (Runtime's road
	// on #t5q4prw). The selection follows them as reachedRecords does, so the record is a graph member.
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
`, ""},
	// The closure that captures the record is written through a const alias, not as a literal, so the
	// write is read by the slot's type, which reaches the record (Runtime's road on #t5q4prw): its type is
	// selected, so it's a graph member.
	{"a record captured by a closure written through an alias", `interface Node { name: string; count: () => number }
function make(name: string): Node { return { name, count: () => 0 }; }
let total = 0;
for (let i = 0; i < 10; i++) {
  const a = make(` + "`a${i}`" + `);
  const rec: Record<string, Node> = { k: a };
  const counter = (): number => Object.keys(rec).length;
  a.count = counter;
  total += a.count();
}
console.log(` + "`${total}`" + `);
`, ""},
	// A graph object holding an empty record it never writes: {} holds nothing, so the cycle a.next
	// closes doesn't run through it, and it lowers and frees whole. The seed's slot reaches the record
	// type, whose values reach the node, which stopped every program of this shape at 7e4d4fb9
	// (TestGraphFlowWriters).
	{"an empty record a graph object holds", graphRecordNode + `console.log(` + "`${a.name} ${Object.keys(a.r).length}`" + `);
`, ""},
	// An empty record a graph object holds, filled to close the cycle by each road that writes a record
	// (#t5q4prw's TestGraphEmptyRecordFilledLaterStops before records joined graph regions): each lowers
	// now, its record a graph member, and frees whole (#1vb0x5s).
	{"an empty record filled by set", `interface Node { name: string; next: Node[]; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], r: {} };
const b: Node = { name: 'b' + '', next: [a], r: {} };
a.next.push(b);
a.r['k'] = a.next;
console.log(` + "`${a.name} ${Object.keys(a.r).length}`" + `);
`, ""},
	{"an empty record filled by coalesce", `interface Node { name: string; next: Node[]; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], r: {} };
const b: Node = { name: 'b' + '', next: [a], r: {} };
a.next.push(b);
a.r['k'] ??= a.next;
console.log(` + "`${a.name} ${Object.keys(a.r).length}`" + `);
`, ""},
	{"an empty record filled by computed key", `interface Node { name: string; next: Node[]; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], r: {} };
const b: Node = { name: 'b' + '', next: [a], r: {} };
a.next.push(b);
const key = 'k' + '';
a.r[key] = a.next;
console.log(` + "`${a.name} ${Object.keys(a.r).length}`" + `);
`, ""},
	{"an empty record filled by callee", `interface Node { name: string; next: Node[]; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], r: {} };
const b: Node = { name: 'b' + '', next: [a], r: {} };
a.next.push(b);
function put(n: Node): void { n.r['k'] = n.next; }
put(a);
console.log(` + "`${a.name} ${Object.keys(a.r).length}`" + `);
`, ""},
	{"an empty record filled by defineHidden", `interface Node { name: string; next: Node[]; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], r: {} };
const b: Node = { name: 'b' + '', next: [a], r: {} };
a.next.push(b);
Object.defineProperty(a.r, 'k', { enumerable: false, writable: false, value: a.next });
console.log(` + "`${a.name} ${Object.keys(a.r).length}`" + `);
`, ""},
	{"an empty record filled by push into a stored container", `interface Node { name: string; next: Node[]; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], r: {} };
const b: Node = { name: 'b' + '', next: [a], r: {} };
a.next.push(b);
const list: Node[] = [];
a.r['k'] = list;
list.push(a);
console.log(` + "`${a.name} ${Object.keys(a.r).length}`" + `);
`, ""},
	// The two record self-cycles lower's TestProgramRecordCycleLowersInBothModes lowers, held here (#1vb0x5s).
	{"a record holding itself", `interface R { [key: string]: R }
const r: R = {};
r['self'] = r;
console.log(Object.keys(r).join(','));
`, ""},
	{"a node's record of links holding it", `interface Node { links: Record<string, Node> }
const a: Node = { links: {} };
a.links['self'] = a;
console.log(Object.keys(a.links).join(','));
`, ""},
	// A node pushed into an array its record holds, with no cycle through the record: a.next.push(b) puts
	// a and b in one graph region, so a counted record and array would make a counted path out of it and
	// back in, and the region never let go (Emit on #t5q4prw: 11 allocated, 2 freed on main 89614cc0).
	// The record and its array are graph members now, and it frees whole (#1vb0x5s).
	{"a node pushed into an array its record holds", `interface Node { name: string; r: Record<string, Node[]>; next: Node[] }
const a: Node = { name: "a", r: {}, next: [] };
const b: Node = { name: "b", r: {}, next: [] };
a.next.push(b);
a.r["k"] = [];
a.r["k"].push(b);
console.log(String(Object.keys(a.r).length));
`, ""},
	// A record type closing its own cycle, a declared field beside its index signature: refused before
	// records joined graph regions (lower's TestGraphRecordCycleLowers), held here to Node and the counted
	// build (#1vb0x5s).
	{"a record type closing its own cycle", `interface Ring { self: Ring | undefined; [key: string]: Ring | undefined }
const ring: Ring = { self: undefined };
ring.self = ring;
console.log(Object.keys(ring).join(','));
`, ""},
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
	// A record spread into a new record on the cycle: the spread's values were planned where they were
	// stored, and the new record's own entry is seeded where it's made.
	{"a record spread into a new record on the cycle", graphRecordTable + `a.r = { k: [b] };
b.r = { ...a.r, j: a.next.slice() };
console.log(Object.keys(b.r).join(','));
`, ""},
	// r[k] ??= node, the fallback taken once and kept, then present.
	{"a record's ??= of a node on the cycle", `interface Node { name: string; peers: Record<string, Node> }
const a: Node = { name: 'a' + '', peers: {} };
const b: Node = { name: 'b' + '', peers: {} };
a.peers['b'] ??= b;
b.peers['a'] ??= a;
a.peers['b'] ??= a;
const got = a.peers['b'];
console.log(` + "`${got === undefined ? '-' : got.name} ${Object.keys(b.peers).join(',')}`" + `);
`, ""},
	{"a record's entry deleted on the cycle", graphRecordTable + `a.r['k'] = [b];
a.r['j'] = a.next.slice();
delete a.r['k'];
console.log(Object.keys(a.r).join(','));
`, ""},
	{"a record's value replaced on the cycle", graphRecordTable + `a.r['k'] = [b];
a.r['k'] = [a, b];
a.r['k'] = a.next.slice();
const got = a.r['k'];
console.log(` + "`${got === undefined ? -1 : got.length}`" + `);
`, ""},
	// Object.defineProperty's hidden key, defined again with another value: refused, and the value it was
	// handed let go, as the runtime's take_in leaves a refused write's.
	{"a define refused on a fixed key on the cycle", graphRecordTable + `Object.defineProperty(a.r, 'k', { enumerable: false, writable: false, value: [b] });
try {
    Object.defineProperty(a.r, 'k', { enumerable: false, writable: false, value: a.next.slice() });
    console.log('redefined');
}
` + graphRecordCaught + `console.log(` + "`${String('k' in a.r)} ${Object.keys(a.r).length}`" + `);
`, ""},
	{"a write refused on a read-only key on the cycle", graphRecordTable + `Object.defineProperty(a.r, 'k', { enumerable: false, writable: false, value: [b] });
try {
    a.r['k'] = [a];
    console.log('written');
}
` + graphRecordCaught + `const got = a.r['k'];
console.log(` + "`${got === undefined ? -1 : got.length}`" + `);
`, ""},
	// The node holds a box, the box an inner object, the inner object the record, the record the node.
	{"a record held two hops from the cycle", `interface Inner { r: Record<string, Node> }
interface Box { inner: Inner }
interface Node { name: string; box: Box }
function make(name: string): Node { return { name, box: { inner: { r: {} } } }; }
let total = 0;
for (let i = 0; i < 10; i++) {
  const a = make(` + "`a${i}`" + `);
  a.box = { inner: { r: { k: a } } };
  total += Object.keys(a.box.inner.r).length;
}
console.log(` + "`${total}`" + `);
`, ""},
	// A key known only as it runs. It lowers to a record set, as every record write does, never to an
	// array's SetIndex. Object.values makes the array written, planned only by the write's seed, so M8
	// leaves it counted between members.
	{"a computed-key write on the cycle", graphRecordTable + `const pool: Record<string, Node> = { a: a, b: b };
for (let i = 0; i < 3; i++) {
  a.r['k' + String(i)] = Object.values(pool);
}
console.log(Object.keys(a.r).join(','));
`, ""},
	// Object.assign takes no record target (records.go's recordObjectCallArguments), so no library call
	// writes into a record.
	{"Object.assign into a record target", graphRecordTable + `Object.assign(a.r, { k: [b] });
console.log(Object.keys(a.r).join(','));
`, "Object.assign on a record"},
	// A record of a type the graph doesn't select, written into a graph slot on the cycle: a record of
	// numbers holds no graph value, so graphTypes never selects it, and judgeRecordSlots stops the write
	// (scan-all's record-stop control, Correctness's fixture on #1vb0x5s). This is the stop's positive
	// witness: every other road here lowers.
	{"an unselected record written into a graph slot", `interface Node { name: string; next: Node[]; r: Node[] | Record<string, number> }
const a: Node = { name: 'a' + '', next: [], r: {} };
const b: Node = { name: 'b' + '', next: [a], r: {} };
a.next.push(b);
a.r = a.next.slice();
a.r = { k: 1 };
console.log(a.name);
`, "a record written into a graph slot"},
	// One record literal, held by both nodes, holding one of them: M7's road, whose emitted C makes
	// exactly one graph record.
	{"one record shared by two nodes", `interface Node { name: string; next: Node[]; r: Record<string, Node> }
const shared: Record<string, Node> = {};
const a: Node = { name: 'a' + '', next: [], r: shared };
const b: Node = { name: 'b' + '', next: [a], r: shared };
a.next.push(b);
shared['k'] = b;
console.log(` + "`${a.name} ${Object.keys(b.r).join(',')}`" + `);
`, ""},
}

func TestGraphRecordOnACycle(t *testing.T) {
	t.Parallel()
	for _, c := range graphRecordRoads {
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

// graphRecordMutants are #1vb0x5s's lowering mutants, each an exact rewrite of a text found once, with the
// road it turns red. M7 rewrites the emitted C, so it's run here, and its road leaks by the counted build.
// M6 and M8 rewrite the lowering's Go, which no test rebuilds in process, so here each is held to its
// text, found exactly once, and it's applied by hand: with it, its road goes red, M6's staying NotYet
// where it should compile, M8's leaving the array Object.values makes counted between members.
var graphRecordMutants = []struct{ name, file, before, after, road string }{
	{
		name:   "M6 noteRecordSlot's stop left firing on a selected record type",
		file:   "internal/lower/graph_records.go",
		before: "\treturn f.l.result.GraphTypes[int(record.Id())] || f.l.result.GraphTypes[int(f.l.concrete(record).Id())]\n",
		after:  "\treturn false\n",
		road:   "a record literal stored into a graph field",
	},
	{
		name:   "M7 dictionaryLiteral emitting adamic_record_new",
		before: "adamic_record_new_graph(adamic_field_reference)",
		after:  "adamic_record_new(adamic_field_reference)",
		road:   "one record shared by two nodes",
	},
	{
		name:   "M8 a record write's seed removed from collect",
		file:   "internal/lower/graph_flow.go",
		before: "\t\t\tcase ir.RecordCall:\n\t\t\t\tif (value.Method == \"set\" || value.Method == \"defineHidden\") && len(value.Arguments) == 3 {\n\t\t\t\t\tmembers(value.Arguments[2], recordSlot(value.Site), graphDemand{site: value.Site})\n\t\t\t\t}\n",
		after:  "",
		road:   "a computed-key write on the cycle",
	},
}

// TestGraphRecordMutants holds each mutant to its text, found exactly once, and runs M7: its road's
// emitted C, its one graph record made counted, must go unbalanced by the counted build.
func TestGraphRecordMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range graphRecordMutants {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			index := slices.IndexFunc(graphRecordRoads, func(road struct{ name, source, stopped string }) bool { return road.name == mutant.road })
			if index < 0 {
				t.Fatalf("no road named %q", mutant.road)
			}
			road := graphRecordRoads[index]
			if road.stopped != "" {
				t.Fatalf("%q stops, so no mutant reds it by its run", mutant.road)
			}
			if mutant.file != "" {
				source, err := os.ReadFile(filepath.Join(repository, mutant.file))
				if err != nil {
					t.Fatal(err)
				}
				if found := strings.Count(string(source), mutant.before); found != 1 {
					t.Fatalf("the mutant's text is in %s %d times, want once", mutant.file, found)
				}
				t.Logf("applied by hand to %s, %s reds %q", mutant.file, mutant.name, mutant.road)
				return
			}
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte(road.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			source := native.C(program)
			if found := strings.Count(source, mutant.before); found != 1 {
				t.Fatalf("the mutant's text is in the emitted C %d times, want once", found)
			}
			counted := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(strings.Replace(source, mutant.before, mutant.after, 1), counted, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			ran := execute(t, counted)
			report := leakcheck.Unbalanced(leakcheck.Run{Stdout: ran.stdout, Stderr: ran.stderr, ExitCode: ran.exitCode})
			if report == "" {
				t.Fatalf("%s left %q balanced by the counted build", mutant.name, mutant.road)
			}
			t.Logf("%s reds %q: %s", mutant.name, mutant.road, report)
		})
	}
}
