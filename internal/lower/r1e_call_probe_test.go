package lower

import (
	"sync/atomic"
	"testing"
	"time"
)

// Probe for #r1ef873, never for main: which receiver shapes the open-receiver cut misses. Each test
// lowers one program with projectionFramesOpened set and prints how many frames it opened, stopping at a
// cap so a runaway reports instead of running.

const r1eProbeCap = 20000

const r1eProbeTypes = `enum Kind { Leaf, Pair, Wrap }
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

func r1eProbe(t *testing.T, name string, source string) {
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
	for {
		select {
		case err := <-done:
			t.Logf("PROBE %s finished frames=%d err=%v", name, frames.Load(), err)
			return
		case <-poll.C:
			if opened := frames.Load(); opened > r1eProbeCap {
				t.Logf("PROBE %s runaway frames>%d", name, r1eProbeCap)
				return
			}
		}
	}
}

// c1: a recursive function whose returns read a field of its own call, seen through a cast.
func TestR1eProbeC1Call(t *testing.T) {
	r1eProbe(t, "c1_call", r1eProbeTypes+`function pick(node: Node): Node {
	switch (node.kind) {
		case Kind.Pair:
			return (pick((node as Pair).left) as Pair).right;
		case Kind.Wrap:
			return (pick((node as Wrap).first) as Wrap).second;
	}
	return node;
}
console.log(String(value(pick(pair))));
console.log(String(value(pick(wrap))));
`)
}

// c2: the same through a closure held in a const.
func TestR1eProbeC2Closure(t *testing.T) {
	r1eProbe(t, "c2_closure", r1eProbeTypes+`const pick = (node: Node): Node => {
	switch (node.kind) {
		case Kind.Pair:
			return (pick((node as Pair).left) as Pair).right;
		case Kind.Wrap:
			return (pick((node as Wrap).first) as Wrap).second;
	}
	return node;
};
console.log(String(value(pick(pair))));
console.log(String(value(pick(wrap))));
`)
}

// c3: two functions each reading a field of the other's call.
func TestR1eProbeC3CallPair(t *testing.T) {
	r1eProbe(t, "c3_call_pair", r1eProbeTypes+`function left(node: Node): Node {
	switch (node.kind) {
		case Kind.Pair:
			return (right((node as Pair).left) as Pair).right;
		case Kind.Wrap:
			return (right((node as Wrap).first) as Wrap).second;
	}
	return node;
}
function right(node: Node): Node {
	switch (node.kind) {
		case Kind.Pair:
			return (left((node as Pair).right) as Pair).left;
		case Kind.Wrap:
			return (left((node as Wrap).second) as Wrap).first;
	}
	return node;
}
console.log(String(value(left(pair))));
console.log(String(value(right(wrap))));
`)
}
