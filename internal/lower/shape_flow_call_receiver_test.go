package lower

import (
	"sync/atomic"
	"testing"
	"time"
)

// callReceiverFrameBound is the most projected query frames lowering any one of the call-rooted programs
// below may open (#r1ef873). With calls unkeyed each passed 20,000 in a fifth of a second, so the bound
// sits far below that and well above what each opens keyed.
const callReceiverFrameBound = 1000000 // measure

const callReceiverTypes = `enum Kind { Leaf, Pair, Wrap }
interface Node { readonly kind: Kind; }
interface Leaf extends Node { readonly kind: Kind.Leaf; readonly value: number; }
interface Pair extends Node { readonly kind: Kind.Pair; readonly left: Node; readonly right: Node; }
interface Wrap extends Node { readonly kind: Kind.Wrap; readonly first: Node; readonly second: Node; }
const one: Leaf = { kind: Kind.Leaf, value: 1 };
const two: Leaf = { kind: Kind.Leaf, value: 2 };
const wrap: Wrap = { kind: Kind.Wrap, first: one, second: two };
const pair: Pair = { kind: Kind.Pair, left: wrap, right: one };
function value(node: Node): number { return node.kind === Kind.Leaf ? (node as Leaf).value : 0; }
`

// A function that returns a field of its own call seen through a cast reaches that call again, one
// level deeper, at every level of the allocation-flow projection: follow demands the call's result node,
// and that node's source is a property of the same call. Before calls were keyed they had neither the
// memo nor the open-receiver cut, so two such returns opened 2^32 frames before the depth cap; on tsc's
// whole program that ran the first link's 55 minutes after the Read-rooted cut had landed (#r1ef873,
// Correctness's row 1 at 2a97945f). Each program must lower in bounded work, counted by
// projectionFramesOpened, and agree with Node. Mutant: calls dropped from projectionValueKey's roots, and
// each count passes its bound.
//
// Not parallel: projectionFramesOpened counts every projected query in the package while it's set, and no
// parallel test runs beside a sequential one.
func TestShapeFlowCallReceiverFinishes(t *testing.T) {
	programs := []struct{ name, source string }{
		{"call", `function pick(node: Node): Node {
	switch (node.kind) {
		case Kind.Pair:
			return pick((node as Pair).left).kind === Kind.Pair ? (pick((node as Pair).left) as Pair).right : node;
		case Kind.Wrap:
			return pick((node as Wrap).first).kind === Kind.Wrap ? (pick((node as Wrap).first) as Wrap).second : node;
	}
	return node;
}
console.log(String(value(pick(pair))));
console.log(String(value(pick(wrap))));
`},
		{"closure", `const pick = (node: Node): Node => {
	switch (node.kind) {
		case Kind.Pair:
			return pick((node as Pair).left).kind === Kind.Pair ? (pick((node as Pair).left) as Pair).right : node;
		case Kind.Wrap:
			return pick((node as Wrap).first).kind === Kind.Wrap ? (pick((node as Wrap).first) as Wrap).second : node;
	}
	return node;
};
console.log(String(value(pick(pair))));
console.log(String(value(pick(wrap))));
`},
		{"call_pair", `function left(node: Node): Node {
	switch (node.kind) {
		case Kind.Pair:
			return right((node as Pair).left).kind === Kind.Pair ? (right((node as Pair).left) as Pair).right : node;
		case Kind.Wrap:
			return right((node as Wrap).first).kind === Kind.Wrap ? (right((node as Wrap).first) as Wrap).second : node;
	}
	return node;
}
function right(node: Node): Node {
	switch (node.kind) {
		case Kind.Pair:
			return left((node as Pair).right).kind === Kind.Pair ? (left((node as Pair).right) as Pair).left : node;
		case Kind.Wrap:
			return left((node as Wrap).second).kind === Kind.Wrap ? (left((node as Wrap).second) as Wrap).first : node;
	}
	return node;
}
console.log(String(value(left(pair))));
console.log(String(value(right(wrap))));
`},
	}
	for _, program := range programs {
		t.Run(program.name, func(t *testing.T) {
			source := callReceiverTypes + program.source
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
					if opened := frames.Load(); opened > callReceiverFrameBound {
						t.Fatalf("lowering %s opened %d projected query frames and was still going; the bound is %d", program.name, opened, callReceiverFrameBound)
					}
				}
			}
			opened := frames.Load()
			if opened > callReceiverFrameBound {
				t.Fatalf("lowering %s opened %d projected query frames; the bound is %d", program.name, opened, callReceiverFrameBound)
			}
			t.Logf("lowering %s opened %d projected query frames", program.name, opened)
			projectionFramesOpened.Store(nil)
			lowersAndAgreesWithNodeAs(t, source, ".ts")
		})
	}
}
