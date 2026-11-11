package oracle

import "testing"

// A shorthand property of a nested function is the function's canonical value inside its declaring
// function, from a sibling and from an arrow, natively, sanitized with leaks checked, and on the
// JavaScript backend (#qk8rztp, the shorthand road).
func TestNestedShorthandIdentityAgreesWithNode(t *testing.T) {
	t.Parallel()
	generatorAgrees(t, "internal/oracle/testdata/nested_shorthand/identity.a")
}
