package lower

import (
	"errors"
	"strings"
	"testing"
)

// A read the checker narrowed to never of a let a call writes stops as not yet at the read
// (narrowedNeverReadGap, #v59xyph), wherever its value would flow: into unknown (nv2), a number slot
// (nv6), an object slot that's read (nv8) and an object slot whose string field is called (nv9, which on
// main read the number in kind as a string pointer and crashed natively where Node throws). Each is
// Language's program from #p6vawef (language/probe-never-between, probe/never-between). nv2 agreed with
// Node before and stops now, the cost the stop takes until #p6vawef's earliest-narrowing build. Mutant 1
// in narrowed_never.mutants.md (the rule unregistered) lowers every row again.
func TestNeverReadOfACallWrittenLetIsNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"nv2", `enum K { A = 1, B = 2 }
interface NodeA { readonly kind: K.A; readonly a: number }
interface NodeB { readonly kind: K.B; readonly b: number }
type N = NodeA | NodeB;
function stray(value: unknown): string { return "stray " + typeof value; }
function describe(start: N): string {
	let node: N = start;
	const flip = (): void => { node = { kind: K.A, a: 1 }; };
	if (node.kind === K.A) return 'a';
	flip();
	if (node.kind === K.B) return 'b';
	{ const seen: unknown = node; return stray(seen); }
}
console.log(describe({ kind: K.B, b: 2 }));
`},
		{"nv6", `enum K { A = 1, B = 2 }
interface NodeA { readonly kind: K.A; readonly a: number }
interface NodeB { readonly kind: K.B; readonly b: number }
type N = NodeA | NodeB;
function stray(value: unknown): string { return "stray " + typeof value; }
function describe(start: N): string {
	let node: N = start;
	const flip = (): void => { node = { kind: K.A, a: 1 }; };
	if (node.kind === K.A) return 'a';
	flip();
	if (node.kind === K.B) return 'b';
	{ const count: number = node; return "count " + String(count + 1); }
}
console.log(describe({ kind: K.B, b: 2 }));
`},
		{"nv8", `enum K { A = 1, B = 2 }
interface NodeA { readonly kind: K.A; readonly a: number }
interface NodeB { readonly kind: K.B; readonly b: number }
type N = NodeA | NodeB;
function stray(value: unknown): string { return "stray " + typeof value; }
function describe(start: N): string {
	let node: N = start;
	const flip = (): void => { node = { kind: K.A, a: 1 }; };
	if (node.kind === K.A) return 'a';
	flip();
	if (node.kind === K.B) return 'b';
	{ const other: { readonly label: string } = node; return "label " + String(other.label === undefined); }
}
console.log(describe({ kind: K.B, b: 2 }));
`},
		{"nv9", `enum K { A = 1, B = 2 }
interface NodeA { readonly kind: K.A; readonly a: number }
interface NodeB { readonly kind: K.B; readonly b: number }
type N = NodeA | NodeB;
function stray(value: unknown): string { return "stray " + typeof value; }
function describe(start: N): string {
	let node: N = start;
	const flip = (): void => { node = { kind: K.A, a: 1 }; };
	if (node.kind === K.A) return 'a';
	flip();
	if (node.kind === K.B) return 'b';
	{ const other: { readonly kind: string } = node; return other.kind.toUpperCase(); }
}
console.log(describe({ kind: K.B, b: 2 }));
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("got %v, want a not yet saying it reads never off a binding a call may write", err)
			}
			if !strings.Contains(err.Error(), "a read the checker narrowed to never of a binding a call may write") {
				t.Errorf("got %v, want a not yet saying it reads never off a binding a call may write", err)
			}
		})
	}
}

// The exhaustive control, nv4: a parameter only its own function writes is narrowed to never by checks
// with no call between them, so its never is the checker's proof and it lowers as before. Mutant 2 in
// narrowed_never.mutants.md (the writtenAcrossFunctions gate dropped) stops it.
func TestNeverReadOfAnUnwrittenParameterLowers(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, `enum K { A = 1, B = 2 }
interface NodeA { readonly kind: K.A; readonly a: number }
interface NodeB { readonly kind: K.B; readonly b: number }
type N = NodeA | NodeB;
function stray(value: unknown): string { return "stray " + typeof value; }
function describe(node: N): string {
	if (node.kind === K.A) return 'a';
	if (node.kind === K.B) return 'b';
	{ const seen: unknown = node; return stray(seen); }
}
console.log(describe({ kind: K.B, b: 2 }));
console.log(describe({ kind: K.A, a: 1 }));
`, ".ts"); err != nil {
		t.Fatalf("want nv4 to lower, got %v", err)
	}
}
