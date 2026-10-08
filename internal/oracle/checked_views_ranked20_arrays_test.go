package oracle

import "testing"

func TestCheckedViewRanked20OriginalArrays(t *testing.T) {
	originalRankedArrayOracle(t, "20", 3, 9, 8, 7)
}

// Not parallel: the explicit refresh writes this group's rows in the shared counts file.
func TestCheckedViewRanked20ArrayCounts(t *testing.T) {
	originalRankedArrayCountTest(t, "20")
}
