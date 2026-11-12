package flow

import (
	"path/filepath"
	"testing"
)

// A traced generator parks at each yield and resumes after it, as an awaited async function does
// (#7wn0t67), so the consumer's points run in the consumer's call and the generator's own pick up
// where its yield left off. Before, the generator's frame stayed on the trace's stack across the
// yield, and the consumer's points were filed under the generator: walk said a point of one
// function ran inside a call of another, and frames indexed the generator's instructions with the
// consumer's ids. Each program runs every per-program check the corpus runs, and every one of them
// is also held to Node natively by TestGeneratorsAgreeWithNode or TestIterableViewsAgreeWithNode
// (internal/oracle). The mutants are in generator_trace.mutants.md.
func TestFlowGeneratorsTraced(t *testing.T) {
	t.Parallel()
	for _, path := range []string{
		// The reproducer: a generator written in place, one yield, closed at it by its consumer's
		// return from a for-of, which Node delivers as return().
		"../oracle/testdata/generators/trace_in_place.a",
		// The same through a named generator.
		"../oracle/testdata/generators/trace_named.a",
		// A break out of a for-of over a nested generator's call, and a return from inside a
		// sibling's: each leaves the generator parked at a yield, dropped, never resumed.
		"../oracle/testdata/generators/nested_early_break.a",
		// A throw out of the generator's body on a later step, caught by the consumer around its
		// hand-driven next(), and by its caller's try: the body leaves by its own finally, not a park.
		"../oracle/testdata/iterable_views/hand_next_generator_throw.a",
		// Hand-driven next() through an IterableIterator<T> parameter: run to done, left after one
		// step by a break and after two by a return, two steps kept across a later next(), and
		// next() asked again after done.
		"../oracle/testdata/iterable_views/hand_next_generator.a",
		// A throw from the consumer's own loop, which closes the generator at a yield inside its
		// loop: the one call that ends at a yield whose next block runs something, which walk lets
		// end there only because it came back from its park and ran nothing.
		"../oracle/testdata/generators/view_body_throw.a",
	} {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			beginFlowUnit(t)
			checkAllFlowProgram(t, path)
		})
	}
}
