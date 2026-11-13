package lower

import (
	"errors"
	"testing"
)

// A generic that instantiates itself deeper at each step is a new type at every level, and
// collectViewedElements' walk ran on it until Go's stack overflowed, with no stop (#k031ey0, Optimizer's
// find on #68x1t2y). Bounded as tsc's isDeeplyNestedType bounds it, the walk gives up seeing to the end,
// every element read is checked, and the program lowers, or stops as not yet, but never crashes. Mutant:
// the open[generic] >= viewedElementNesting test taken out, and this overflows the stack. Run that
// mutant with -run on this test alone, since the overflow ends the whole test binary.
func TestViewedElementsStopOnAGenericThatDeepensItself(t *testing.T) {
	t.Parallel()
	source := `interface Grow<T> { readonly value?: T; readonly next?: Grow<Grow<T>> }
interface N { readonly kind: number; readonly grow?: Grow<number> }
interface Id extends N { readonly kind: 1; readonly text: string }
function isId(n: N): n is Id { return n.kind === 1; }
const id: Id = { kind: 1, text: "a" };
const n: N = id;
if (isId(n)) { console.log(n.text); }
`
	_, err := lowerSourceAs(t, source, ".ts")
	var notYet *NotYet
	if err != nil && !errors.As(err, &notYet) {
		t.Fatalf("want the program lowered or not yet, got %v", err)
	}
}
