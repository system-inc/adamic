package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

const anyRelationPrelude = `interface Loose { readonly kind: number; readonly pos: number; readonly text: any; }
interface HasText { readonly text: string; }
`

// A relation that carries a member typed any into a typed member is NotYet at its site, on every road
// relationSite names (#tmdhz1d, step 1): Language's v2 through an initializer, an argument, a return, an
// object literal's field (caught where the literal itself is related, one level out), an any element
// seen as a string, a function whose result is any seen as returning a
// string, and v1, a cast from Loose that the result is then widened from. Mutant: refuseAnyRelation
// answering nil, under which v2 lowers, and native dies under AddressSanitizer reading text as a string.
func TestAnyMemberRelationIsNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"initializer", `function show(n: Loose): void { const h: HasText = n; console.log(h.text); }
const w = { kind: 9, pos: 1, text: 5 }; show(w);`, "a member typed any (text) related to string at the initializer site"},
		{"argument", `function read(h: HasText): void { console.log(h.text); }
function show(n: Loose): void { read(n); }
const w = { kind: 9, pos: 1, text: 5 }; show(w);`, "a member typed any (text) related to string at the argument site"},
		{"return", `function asText(n: Loose): HasText { return n; }
const w = { kind: 9, pos: 1, text: 5 }; console.log(asText(w).text);`, "a member typed any (text) related to string at the return site"},
		{"object field", `function show(n: Loose): void { const box: { readonly inner: HasText } = { inner: n }; console.log(box.inner.text); }
const w = { kind: 9, pos: 1, text: 5 }; show(w);`, "a member typed any (inner.text) related to string at the initializer site"},
		{"array element", `interface Bag { readonly items: any[]; }
interface Names { readonly items: readonly string[]; }
function show(b: Bag): void { const n: Names = b; console.log(String(n.items[0])); }
const w = { items: [5] }; show(w);`, "a member typed any (items[]) related to string at the initializer site"},
		{"result", `interface Make { readonly make: () => any; }
interface MakeText { readonly make: () => string; }
function show(m: Make): void { const t: MakeText = m; console.log(t.make()); }
const w = { make: () => 5 }; show(w);`, "a member typed any (make()) related to string at the initializer site"},
		// The source expression owns the any (Adamic's ruling, Lower's line): a program local counts whoever
		// wrote its type, so the program's own any[] seen as string[] counts (q4), and so does a program
		// parameter Array.isArray narrowed to any[], seen as string[] or passed as one (p2, p8), and a library
		// result first stored in a program local. Mutant: nearest-property ownership, under which q4, p2 and
		// p8 lower.
		{"program's any element", `const loose: any[] = [5];
const t: string[] = loose;
console.log(t[0] ?? "none");`, "a member typed any ([]) related to string at the initializer site"},
		{"library's any element", `function show(x: unknown): void { if (Array.isArray(x)) { const ys: string[] = x; console.log(ys[0] ?? "none"); } }
show([5]);`, "a member typed any ([]) related to string at the initializer site"},
		{"library's any element passed", `function count(ys: readonly string[]): number { return ys.length; }
function show(x: unknown): void { if (Array.isArray(x)) { console.log(String(count(x))); } }
show([5]);`, "a member typed any ([]) related to string at the argument site"},
		// A pruning rule's subtree (a type predicate's) beside it doesn't hide the recorded NotYet (Lower).
		{"after a pruned sibling", `function isWord(x: unknown): x is string { return typeof x === "string"; }
function show(n: Loose): void { const h: HasText = n; console.log(h.text); }
const w = { kind: 9, pos: 1, text: 5 }; console.log(String(isWord(w))); show(w);`, "a member typed any (text) related to string at the initializer site"},
		{"library result stored first", `interface Visible { readonly visible: number; }
function show(view: Visible): void { const e = Object.entries(view); const t: [string, number][] = e; console.log(String(t.length)); }
show({ visible: 1 });`, "a member typed any ([][1]) related to number at the initializer site"},
		{"cast then widened", `interface LitL { readonly kind: 9; readonly pos: number; readonly text: any; }
function show(n: Loose): void { if (n.kind === 9) { const l = n as LitL; const h: HasText = l; console.log(h.text); } }
const w = { kind: 9, pos: 1, text: 5 }; show(w);`, "a member typed any (text) related to string at the initializer site"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, anyRelationPrelude+probe.source)
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet with %q, got %v", probe.want, err)
			}
		})
	}
}

// An honest copy, typed string throughout, and a member typed any seen as unknown, which makes no
// layout claim, both still lower and agree with Node.
func TestAnyMemberRelationLeavesHonestRelations(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"typed copy", `interface Tight { readonly kind: number; readonly pos: number; readonly text: string; }
function show(n: Tight): void { const h: HasText = n; console.log(h.text); }
const w = { kind: 9, pos: 1, text: "x" }; show(w);`},
		{"seen as unknown", `interface Opaque { readonly text: unknown; }
function show(n: Loose): void { const o: Opaque = n; console.log(typeof o.text); }
const w = { kind: 9, pos: 1, text: 5 }; show(w);`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, anyRelationPrelude+probe.source, ".ts")
		})
	}
}

// The any relation is the lowest-priority stop in the refusal walk (Compiler's and Lower's word on
// #tmdhz1d): it records its NotYet and lets the walk go on, so a walk-time Refused later in the module is
// what the program hears, with the NotYet still recorded beside it. If precedence ever flips, the program
// hears the any relation instead and this goes red. Mutant: refuse() not returning the deferred stop,
// under which every NotYet pin above lowers.
func TestAnyMemberRelationLosesToARefusal(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.ts")
	source := anyRelationPrelude + `function show(n: Loose): void { const h: HasText = n; console.log(h.text); }
const w = { kind: 9, pos: 1, text: 5 }; show(w);
const box: { tag?: number } = { tag: 1 }; delete box.tag;`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checker, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{checker: checker, program: program}
	err = l.refuse(file)
	var refused *Refused
	if !errors.As(err, &refused) {
		t.Fatalf("want the later delete's Refused, got %v", err)
	}
	var pending *NotYet
	if !errors.As(l.refusal.deferred, &pending) || !strings.Contains(pending.What, "a member typed any (text) related to string at the initializer site") {
		t.Fatalf("want the any relation's NotYet recorded beside it, got %v", l.refusal.deferred)
	}
}
