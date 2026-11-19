package lower

import (
	"testing"
)

// overloadViewWidenedShape is step 1's road for a branded supertype: tag's first overload promises a
// TaggedNode over an implementation answering Node, so the call's result is checked and read through
// TaggedNode's view from the call on. This body lies: the node it returns has no name.
const overloadViewWidenedShape = `interface Node { readonly kind: number }
interface TaggedNode extends Node { readonly __tagged: undefined; readonly name: string }
function tag(node: Node, sure: true): TaggedNode;
function tag(node: Node, sure: boolean): Node;
function tag(node: Node, sure: boolean): Node { return node; }
const r = tag({ kind: 1 }, true);
`

// A read through the view itself is checked where it's read, so the program lowers, and the lie stops
// at its read rather than reading a field the value doesn't have.
func TestOverloadResultViewDirectReadLowers(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, overloadViewWidenedShape+"console.log(r.name);\n", ".ts"); err != nil {
		t.Fatalf("want the read through the view lowered, got %v", err)
	}
}

// The caller's local holds the result as the call's declared type, never a narrowing, so the check's call
// is where the value enters the view (overloadCheckCalls). A read through a wider type the result is
// assigned or passed to is weighed as the predicate road's is (Language's q06 and Optimizer's matched pair
// on #z5d0jye: the same reads segfaulted natively on step 1's road while the predicate road stopped them),
// and is checked where it's made (#zpzm9ta), so the lie stops at that read on sanitized native and the
// JavaScript backend, after what the program printed before it. Mutant: overloadResultCheck not recording
// its call, and each read is a plain one again.
func TestOverloadResultViewWidenedReadIsChecked(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, tail, stdout, read string }{
		{"through a structural local", "const w: { readonly name: string } = r;\nconsole.log(w.name.toUpperCase());\n", "", "w.name"},
		{"through a parameter", "function show(v: { readonly name: string }): void { console.log(v.name.toUpperCase()); }\nshow(r);\n", "", "v.name"},
		// A second call site of the same overload finds its view cached and registers none, so its call is an
		// entry only because every check call is (Language's v10 on a44f6b8c, and its reverse). Mutant: only a
		// call whose check registered a view recorded, and the second site's read is a plain one.
		{"a second site widened", "const b: { readonly name: string } = tag({ kind: 2 }, true);\nconsole.log(String(r.kind));\nconsole.log(b.name.toUpperCase());\n", "1\n", "b.name"},
		{"the first site widened", "const w: { readonly name: string } = r;\nconst b = tag({ kind: 2 }, true);\nconsole.log(String(b.kind));\nconsole.log(w.name.toUpperCase());\n", "2\n", "w.name"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, overloadViewWidenedShape+probe.tail, ".ts")
			if err != nil {
				t.Fatalf("want the widened read lowered as a checked read, got %v", err)
			}
			if checked, _ := widenedReadsOf(t, program, "name"); checked == 0 {
				t.Fatalf("%s isn't a widened read checked where it's made", probe.read)
			}
			overloadCheckStops(t, overloadViewWidenedShape+probe.tail, probe.stdout, "adamic: panic: field read failed: "+probe.read+" is not initialized; expected string, found missing")
		})
	}
}
