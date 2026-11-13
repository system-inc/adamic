package lower

import (
	"errors"
	"strings"
	"testing"
)

// A generic that instantiates itself deeper at each step is a new type at every level, whether its
// target nests in itself or grows another type's arguments or an object literal type's. Every structural
// walk over a view's target whose only stop was a set of types it had seen ran on it without end:
// collectViewedElements until Go's stack overflowed (Optimizer's find on #68x1t2y), then
// prepareUntaggedStructuralRead and viewContract, which hung (J-rae10tx, J-g91t70n). The view pre-pass sees
// every view target first, so each such program now stops there, not yet at the view (#k031ey0). Mutant:
// collectViewedElements' stop dropped (the !ok branch returning without setting stop), and each row hangs
// in a later walk. Run that mutant with -run on this test alone and a short -timeout, since a hang ends
// the whole test binary.
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
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "instantiates itself deeper at each step") {
				t.Fatalf("want the growing type's not yet at its view, got %v", err)
			}
		})
	}
}
