package oracle

import "testing"

// ??= over a union holding null takes the write, in a local and in a field, natively as on Node: the
// native test once asked only for the null pointer, which is undefined, and printed null where Node
// prints the 7 written. Mutant: dropping the &adamic_null test from logicalAssignment's ??= condition.
func TestNullishAssignmentOverAUnionNull(t *testing.T) {
	t.Parallel()
	loomViewsDAgrees(t, "internal/oracle/testdata/nullish_assignment/union_null.a")
}
