package lower

import (
	"errors"
	"strings"
	"testing"
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
