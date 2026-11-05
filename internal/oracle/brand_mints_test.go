package oracle

import "testing"

// Brand mints (#yth65gz, ruled on #b5w3ycg Nov 5): tsc's sorted-array constructions agree with Node on
// every backend and leak nothing, so the erased cast hands on the array it was made from.
func TestBrandMintsAgreeWithNode(t *testing.T) {
	t.Parallel()
	generatorAgrees(t, "internal/oracle/testdata/brand_mints/sorted_arrays.ts")
}
