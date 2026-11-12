package lower

import (
	"errors"
	"strings"
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
// is where the value enters the view (viewSource.checkCall). A read through a wider type the result is
// assigned or passed to is weighed as the predicate road's is, and stays NotYet until it checks against
// the view (Language's q06 and Optimizer's matched pair on #z5d0jye: the same reads segfaulted natively
// on step 1's road while the predicate road stopped them). Mutant: overloadResultCheck leaving checkCall
// unset, and both lower.
func TestOverloadResultViewWidenedReadIsWeighed(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, tail string }{
		{"through a structural local", "const w: { readonly name: string } = r;\nconsole.log(w.name.toUpperCase());\n"},
		{"through a parameter", "function show(v: { readonly name: string }): void { console.log(v.name.toUpperCase()); }\nshow(r);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, overloadViewWidenedShape+probe.tail, ".ts")
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), "a read of field name through type") || !strings.Contains(err.Error(), "which may hold a checked view's value") {
				t.Fatalf("want NotYet at the widened read, got %v", err)
			}
		})
	}
}
