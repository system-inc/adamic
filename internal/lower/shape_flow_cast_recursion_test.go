package lower

import (
	"testing"
	"time"
)

// binder.ts's isNarrowableOperand recurses on a field of its own parameter seen through a cast,
// `isNarrowableOperand((expr as BinaryExpression).left)`, so the parameter's sources are property
// reads whose receiver is a CheckedCast of the parameter, never a bare Read, and only a bare Read's
// projected query was memoized. Every view origin's query reopened a frame for the cast receiver at
// every level, once per recursive call: three calls make 3^32 frames before the depth cap, and
// lowering never finished (#tzy22kd, found by Performance's unit-C instrument on binder.ts; main
// stopped at binder.ts:1341). Lowering must finish well inside the budget, and the program agree
// with Node, natively and on JavaScript. Mutant: the frame memo and the follow dedupe both dropped,
// so the budget runs out.
func TestShapeFlowCastRecursionFinishes(t *testing.T) {
	t.Parallel()
	const source = `interface Node { readonly kind: number; }
interface Wrap extends Node { readonly inner: Node; }
interface Pair extends Node { readonly left: Node; readonly right: Node; }
function narrowable(n: Node): boolean {
	switch (n.kind) {
		case 1:
			return narrowable((n as Wrap).inner);
		case 2:
			return narrowable((n as Pair).left);
		case 3:
			return narrowable((n as Pair).right);
	}
	return n.kind === 0;
}
const leaf: Node = { kind: 0 };
const wrap: Wrap = { kind: 1, inner: leaf };
const pair: Pair = { kind: 2, left: wrap, right: leaf };
const outer: Pair = { kind: 3, left: leaf, right: pair };
console.log(String(narrowable(outer)));
console.log(String(narrowable({ kind: 5 })));
`
	done := make(chan error, 1)
	go func() {
		_, err := lowerSourceAs(t, source, ".ts")
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("want it lowered, got %v", err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("lowering the cast recursion took longer than 20 seconds")
	}
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "true\nfalse\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "true\nfalse\n")
	}
}
