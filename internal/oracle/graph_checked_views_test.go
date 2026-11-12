package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// A graph program holds checked views since slice 3 of #ztpcwrr (#ef727kr): a view owns nothing, so a
// read through it retains as any read does, and a write through it takes writeProperty's graph hold.
// Each row puts the viewed objects on a cycle 1,000 times in a function, so every turn's cycle is a
// graph region that must free whole, and holds the answer to Node and the counted build. The views are
// .ts, where a downcast stays a checked view (in .a it's refused, #m8dvvmf). Writing a field of a
// checked type (Identifier here) is NotYet without a source-slot certificate, and a legacy view's field
// can't hold a reference yet, so each cycle closes through a plain Node or the holder of the value. The
// write is a view stored into a graph slot, which the graph hold takes as any store.
//
// Mutant: graphAdoptionBoundary's refusal of ViewOrigins and CheckedFields put back stops every row
// before it lowers.
func TestGraphCheckedViews(t *testing.T) {
	t.Parallel()
	const nodes = `interface Node { readonly kind: 'identifier' | 'number'; readonly pos: number; next: Node | undefined }
interface Identifier extends Node { readonly kind: 'identifier'; readonly name: string }
function identifier(node: Node): Identifier { return node as Identifier; }
`
	for _, c := range []struct{ name, source string }{
		{"a checked read", nodes + `function turn(index: number): string {
	const b: Node = { kind: 'number', pos: index, next: undefined };
	const a: Identifier = { kind: 'identifier', pos: index, next: b, name: 'a' + String(index) };
	b.next = a;
	const seen: Node = a;
	return identifier(seen).name;
}
let last = '';
for (let index = 0; index < 1000; index++) last = turn(index);
console.log(last);
`},
		{"a view stored into a graph slot", nodes + `function turn(index: number): string {
	const b: Node = { kind: 'number', pos: index, next: undefined };
	const a: Identifier = { kind: 'identifier', pos: index, next: b, name: 'a' + String(index) };
	const seen: Node = a;
	const view = identifier(seen);
	b.next = view;
	return view.name + ' ' + String(b.next === a);
}
let last = '';
for (let index = 0; index < 1000; index++) last = turn(index);
console.log(last);
`},
		{"a tagged cast", `type First = { readonly kind: 'a'; readonly name: string };
type Second = { readonly kind: 'b'; readonly name: string };
type Third = { readonly kind: 'c'; readonly name: string };
type Item = First | Second | Third;
interface Holder { item: Item; self: Holder | undefined }
function narrow(value: Item): First | Second { return value as First | Second; }
function turn(index: number): string {
	const holder: Holder = { item: { kind: 'a', name: 'a' + String(index) }, self: undefined };
	holder.self = holder;
	return narrow(holder.item).name;
}
let last = '';
for (let index = 0; index < 1000; index++) last = turn(index);
console.log(last);
`},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if len(program.GraphTypes) == 0 {
				t.Fatal("not a graph program, so it pins nothing about one")
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
			t.Logf("counted: %s", ran.stderr)
		})
	}
}
