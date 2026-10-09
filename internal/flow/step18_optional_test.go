package flow

import "testing"

// Expressions are atomic graph instructions; optional argument effects must
// remain conservative while the emitted runtime follows every graph edge.
func TestStep18OptionalMethodPaths(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"runtime-chain-index.a", "runtime-chain-intrinsics.a", "runtime-chain-callable.a", "runtime-chain-method-continuation.a", "runtime-methods.a", "runtime-bound-methods.a", "runtime-bound-method-arguments.a", "runtime-bound-method-selection.a"} {
		t.Run(name, func(t *testing.T) {
			observed := traced(t, "../../docs/step-18/fixtures/"+name)
			for _, problem := range walk(observed.graphs, observed.marked, observed.events) {
				t.Error(problem)
			}
		})
	}
}
