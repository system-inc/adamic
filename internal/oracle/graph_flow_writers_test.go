package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// graphFlows seeded a demand only at a graph-typed local, a return, and a property, index, push, Map
// set or Set add into a graph slot, so a copy only follow plans (toSorted here) reaching a graph slot
// any other way stayed counted inside the cycle and leaked (#dnb0yhj, Codex 01a1bd47 on #y7wexf5's
// head: s1 freed 3 of 19, s2 4 of 22, s3 4 of 19, s5 5 of 20). Each is now seeded: new Map's entries
// through members at the slot write, fill in place and splice's items as a push is, and a for-of
// binding over an array literal by its elements. Each control, the same program over slice(), which
// graphAllocation plans by type, freed everything before. All held to Node and the counted build.
//
// Mutants, each red on its own row by the counted build: members without its MapNew arm (s1); collect
// without its ArrayFill arm (s2), its ArraySplice arm (s3), or its ForOf arm (s5).
func TestGraphFlowWriters(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ name, source string }{
		{"s1, new Map's entries", `// s1: a copy stored through new Map's entries
interface Node { name: string; next: Node[]; kids: Node[][]; m: Map<string, Node[]>; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], kids: [], m: new Map<string, Node[]>(), r: {} };
const b: Node = { name: 'b' + '', next: [a], kids: [], m: new Map<string, Node[]>(), r: {} };
a.next.push(b); // a cycle: the program is a graph program
const byName = (x: Node, y: Node): number => (x.name < y.name ? -1 : 1);
a.m = new Map<string, Node[]>([['k', a.next.toSorted(byName)]]);
const got = a.m.get('k'); console.log(` + "`" + `${got === undefined ? -1 : got.length}` + "`" + `);
`},
		{"s1 control", `// s1: a copy stored through new Map's entries (control: slice, planned by graphAllocation)
interface Node { name: string; next: Node[]; kids: Node[][]; m: Map<string, Node[]>; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], kids: [], m: new Map<string, Node[]>(), r: {} };
const b: Node = { name: 'b' + '', next: [a], kids: [], m: new Map<string, Node[]>(), r: {} };
a.next.push(b); // a cycle: the program is a graph program
const byName = (x: Node, y: Node): number => (x.name < y.name ? -1 : 1);
a.m = new Map<string, Node[]>([['k', a.next.slice()]]);
const got = a.m.get('k'); console.log(` + "`" + `${got === undefined ? -1 : got.length}` + "`" + `);
`},
		{"s2, fill", `// s2: a copy stored by fill
interface Node { name: string; next: Node[]; kids: Node[][]; m: Map<string, Node[]>; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], kids: [], m: new Map<string, Node[]>(), r: {} };
const b: Node = { name: 'b' + '', next: [a], kids: [], m: new Map<string, Node[]>(), r: {} };
a.next.push(b); // a cycle: the program is a graph program
const byName = (x: Node, y: Node): number => (x.name < y.name ? -1 : 1);
a.kids = [[], []]; a.kids.fill(a.next.toSorted(byName));
console.log(` + "`" + `${a.kids.length} ${a.kids[0] === undefined ? -1 : a.kids[0].length}` + "`" + `);
`},
		{"s2 control", `// s2: a copy stored by fill (control: slice, planned by graphAllocation)
interface Node { name: string; next: Node[]; kids: Node[][]; m: Map<string, Node[]>; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], kids: [], m: new Map<string, Node[]>(), r: {} };
const b: Node = { name: 'b' + '', next: [a], kids: [], m: new Map<string, Node[]>(), r: {} };
a.next.push(b); // a cycle: the program is a graph program
const byName = (x: Node, y: Node): number => (x.name < y.name ? -1 : 1);
a.kids = [[], []]; a.kids.fill(a.next.slice());
console.log(` + "`" + `${a.kids.length} ${a.kids[0] === undefined ? -1 : a.kids[0].length}` + "`" + `);
`},
		{"s3, splice", `// s3: a copy inserted by splice
interface Node { name: string; next: Node[]; kids: Node[][]; m: Map<string, Node[]>; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], kids: [], m: new Map<string, Node[]>(), r: {} };
const b: Node = { name: 'b' + '', next: [a], kids: [], m: new Map<string, Node[]>(), r: {} };
a.next.push(b); // a cycle: the program is a graph program
const byName = (x: Node, y: Node): number => (x.name < y.name ? -1 : 1);
a.kids.splice(0, 0, a.next.toSorted(byName));
console.log(` + "`" + `${a.kids.length} ${a.kids[0] === undefined ? -1 : a.kids[0].length}` + "`" + `);
`},
		{"s3 control", `// s3: a copy inserted by splice (control: slice, planned by graphAllocation)
interface Node { name: string; next: Node[]; kids: Node[][]; m: Map<string, Node[]>; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], kids: [], m: new Map<string, Node[]>(), r: {} };
const b: Node = { name: 'b' + '', next: [a], kids: [], m: new Map<string, Node[]>(), r: {} };
a.next.push(b); // a cycle: the program is a graph program
const byName = (x: Node, y: Node): number => (x.name < y.name ? -1 : 1);
a.kids.splice(0, 0, a.next.slice());
console.log(` + "`" + `${a.kids.length} ${a.kids[0] === undefined ? -1 : a.kids[0].length}` + "`" + `);
`},
		{"s5, a for-of binding", `// s5: a copy bound by for-of, then pushed into a graph slot
interface Node { name: string; next: Node[]; kids: Node[][]; m: Map<string, Node[]>; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], kids: [], m: new Map<string, Node[]>(), r: {} };
const b: Node = { name: 'b' + '', next: [a], kids: [], m: new Map<string, Node[]>(), r: {} };
a.next.push(b); // a cycle: the program is a graph program
const byName = (x: Node, y: Node): number => (x.name < y.name ? -1 : 1);
for (const c of [a.next.toSorted(byName)]) { a.kids.push(c); }
console.log(` + "`" + `${a.kids.length} ${a.kids[0] === undefined ? -1 : a.kids[0].length}` + "`" + `);
`},
		{"s5 control", `// s5: a copy bound by for-of, then pushed into a graph slot (control: slice, planned by graphAllocation)
interface Node { name: string; next: Node[]; kids: Node[][]; m: Map<string, Node[]>; r: Record<string, Node[]> }
const a: Node = { name: 'a' + '', next: [], kids: [], m: new Map<string, Node[]>(), r: {} };
const b: Node = { name: 'b' + '', next: [a], kids: [], m: new Map<string, Node[]>(), r: {} };
a.next.push(b); // a cycle: the program is a graph program
const byName = (x: Node, y: Node): number => (x.name < y.name ? -1 : 1);
for (const c of [a.next.slice()]) { a.kids.push(c); }
console.log(` + "`" + `${a.kids.length} ${a.kids[0] === undefined ? -1 : a.kids[0].length}` + "`" + `);
`},
	} {
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
