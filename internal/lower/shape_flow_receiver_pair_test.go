package lower

import (
	"sync/atomic"
	"testing"
	"time"
)

// receiverPairFrameBound is the most projected query frames lowering two mutually recursive projections
// may open (#r1ef873): it opens 58 at the head and 2,400 with the open-receiver cut dropped, so the bound
// sits well above the first and far below the second.
const receiverPairFrameBound = 300

// Two receivers whose fields' sources are each other: a projection of left's receiver reaches right's, and
// right's reaches left's, each level a deeper frame the depth-keyed memo never matched, so with any fan-out
// the query opened fan-out^32 frames before the depth cap. On tsc's whole program that ran the first link's
// 55 minutes in reachingAllocations and projectedSources (#r1ef873, Correctness's J-k7dy1eb). A receiver
// already open on the projection stack now answers Unknown, as a busy Read does. Lowering must finish in
// bounded work, counted by projectionFramesOpened as TestShapeFlowCastRecursionFinishes counts it, and the
// program agree with Node. Mutant: the open-receiver cut dropped, and the count passes its bound.
//
// Not parallel: projectionFramesOpened counts every projected query in the package while it's set, and no
// parallel test runs beside a sequential one.
func TestShapeFlowReceiverPairFinishes(t *testing.T) {
	const source = `enum Kind { Leaf, Pair, Wrap }
interface Node { readonly kind: Kind; }
interface Leaf extends Node { readonly kind: Kind.Leaf; readonly value: number; }
interface Pair extends Node { readonly kind: Kind.Pair; readonly left: Node; readonly right: Node; }
interface Wrap extends Node { readonly kind: Kind.Wrap; readonly first: Node; readonly second: Node; }
function sumPair(node: Node): number {
	switch (node.kind) {
		case Kind.Pair:
			return sumWrap((node as Pair).left) + sumWrap((node as Pair).right);
		case Kind.Wrap:
			return sumWrap((node as Wrap).first) + sumWrap((node as Wrap).second);
	}
	return node.kind === Kind.Leaf ? (node as Leaf).value : 0;
}
function sumWrap(node: Node): number {
	switch (node.kind) {
		case Kind.Wrap:
			return sumPair((node as Wrap).first) + sumPair((node as Wrap).second);
		case Kind.Pair:
			return sumPair((node as Pair).left) + sumPair((node as Pair).right);
	}
	return node.kind === Kind.Leaf ? (node as Leaf).value : 0;
}
const one: Leaf = { kind: Kind.Leaf, value: 1 };
const two: Leaf = { kind: Kind.Leaf, value: 2 };
const wrap: Wrap = { kind: Kind.Wrap, first: one, second: two };
const pair: Pair = { kind: Kind.Pair, left: wrap, right: one };
console.log(String(sumPair(pair)));
console.log(String(sumWrap(wrap)));
`
	var frames atomic.Int64
	projectionFramesOpened.Store(&frames)
	defer projectionFramesOpened.Store(nil)
	done := make(chan error, 1)
	go func() {
		_, err := lowerSourceAs(t, source, ".ts")
		done <- err
	}()
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
			if opened := frames.Load(); opened > receiverPairFrameBound {
				t.Fatalf("lowering the receiver pair opened %d projected query frames and was still going; the bound is %d", opened, receiverPairFrameBound)
			}
		}
	}
	opened := frames.Load()
	if opened > receiverPairFrameBound {
		t.Fatalf("lowering the receiver pair opened %d projected query frames; the bound is %d", opened, receiverPairFrameBound)
	}
	t.Logf("lowering the receiver pair opened %d projected query frames", opened)
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}
