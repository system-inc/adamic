package lower

import (
	"os"
	"sync/atomic"
	"testing"
	"time"
)

// receiverCycleFrameBound is the most projected query frames lowering two receivers that project to each other
// may open.
const receiverCycleFrameBound = 400

// Two functions that recurse into each other through a field of a cast of their own parameter, as tsc's tree walks
// do, make two receivers project to each other: walkA's parameter takes `(n as B).a` and `(n as A).b` from walkB,
// whose parameter takes the same from walkA, so a projected query of `(n as A)` follows to `(n as B)` and back. The
// frame memo keys a query by its depth, so each reopening was a new frame until the depth cap of 32, and tsc's first
// link spent its whole 55 minutes there (#r1ef873, J-k7dy1eb: goroutine 1 in reachingAllocations and
// projectedSources, two receivers alternating from depth 25 to 31). A receiver already open on the projection stack
// now answers Unknown, which keeps the checks. The pin is the work itself, the count of projected query frames
// lowering opens, which no box's load moves; the 20 s deadline runs only on the budget box
// (ADAMIC_UNIT_BUDGET=1). Mutant: the open-receiver guard dropped, so every reopening is a frame again.
//
// Not parallel: projectionFramesOpened counts every projected query in the package while it's set, and no
// parallel test runs beside a sequential one.
func TestShapeFlowReceiverCycleFinishes(t *testing.T) {
	const source = `enum Kind { A, B, Leaf }
interface Node { readonly kind: Kind; }
interface A extends Node { readonly kind: Kind.A; readonly b: Node; readonly c: Node; }
interface B extends Node { readonly kind: Kind.B; readonly a: Node; readonly d: Node; }
interface Leaf extends Node { readonly kind: Kind.Leaf; }
function walkA(n: Node): number {
	switch (n.kind) {
		case Kind.A:
			return 1 + walkB((n as A).b) + walkB((n as A).c);
		case Kind.B:
			return walkA((n as B).a) + walkA((n as B).d);
	}
	return 0;
}
function walkB(n: Node): number {
	switch (n.kind) {
		case Kind.B:
			return 1 + walkA((n as B).a) + walkA((n as B).d);
		case Kind.A:
			return walkB((n as A).b) + walkB((n as A).c);
	}
	return 0;
}
const leaf: Leaf = { kind: Kind.Leaf };
const inner: B = { kind: Kind.B, a: leaf, d: leaf };
const outer: A = { kind: Kind.A, b: inner, c: leaf };
const top: B = { kind: Kind.B, a: outer, d: inner };
console.log(String(walkA(outer)));
console.log(String(walkB(top)));
console.log(String(walkA(top) + walkB(outer)));
`
	var frames atomic.Int64
	projectionFramesOpened.Store(&frames)
	defer projectionFramesOpened.Store(nil)
	done := make(chan error, 1)
	go func() {
		_, err := lowerSourceAs(t, source, ".ts")
		done <- err
	}()
	var budget <-chan time.Time
	if os.Getenv("ADAMIC_UNIT_BUDGET") == "1" {
		budget = time.After(20 * time.Second)
	}
	poll := time.NewTicker(10 * time.Millisecond)
	defer poll.Stop()
	for lowered := false; !lowered; {
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
			lowered = true
		case <-poll.C:
			if opened := frames.Load(); opened > receiverCycleFrameBound {
				t.Fatalf("lowering two receivers that project to each other opened %d projected query frames and was still going; the bound is %d", opened, receiverCycleFrameBound)
			}
		case <-budget:
			t.Fatal("lowering two receivers that project to each other took longer than 20 seconds")
		}
	}
	opened := frames.Load()
	if opened > receiverCycleFrameBound {
		t.Fatalf("lowering two receivers that project to each other opened %d projected query frames; the bound is %d", opened, receiverCycleFrameBound)
	}
	t.Logf("lowering two receivers that project to each other opened %d projected query frames", opened)
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "2\n3\n3\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "2\n3\n3\n")
	}
}
