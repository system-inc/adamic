package lower

import (
	"errors"
	"testing"
)

// A generic that instantiates itself deeper at each step is a new type at every level, whether its
// target nests in itself or grows another type's arguments, and
// collectViewedElements' walk ran on it until Go's stack overflowed, with no stop (#k031ey0, Optimizer's
// find on #68x1t2y). Bounded as tsc's isDeeplyNestedType bounds it, the walk gives up seeing to the end,
// every element read is checked, and the program lowers, or stops as not yet, but never crashes. Mutant:
// the before.count >= viewedElementNesting test taken out, and each row overflows the stack. Run that
// mutant with -run on this test alone, since the overflow ends the whole test binary.
func TestViewedElementsStopOnAGenericThatDeepensItself(t *testing.T) {
	t.Parallel()
	tail := `interface Id extends N { readonly kind: 1; readonly text: string }
function isId(n: N): n is Id { return n.kind === 1; }
const id: Id = { kind: 1, text: "a" };
const n: N = id;
if (isId(n)) { console.log(n.text); }
`
	for _, row := range []struct{ name, growth string }{
		// The target nests in itself (Optimizer's find on #68x1t2y).
		{"in itself", `interface Grow<T> { readonly value?: T; readonly next?: Grow<Grow<T>> }
interface N { readonly kind: number; readonly grow?: Grow<number> }
`},
		// The growth is in another generic's arguments, Array's (Optimizer's should on #k031ey0).
		{"in an array", `interface G<T> { readonly value?: T; readonly next?: G<T[]> }
interface N { readonly kind: number; readonly grow?: G<number> }
`},
		// The growth is an anonymous object type, which no reference names.
		{"in an object literal type", `interface G<T> { readonly value?: T; readonly next?: G<{ readonly v: T }> }
interface N { readonly kind: number; readonly grow?: G<number> }
`},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.growth+tail, ".ts")
			var notYet *NotYet
			if err != nil && !errors.As(err, &notYet) {
				t.Fatalf("want the program lowered or not yet, got %v", err)
			}
		})
	}
}
