package oracle

import "testing"

func TestCheckedViewRanked21OriginalArrays(t *testing.T) {
	originalRankedArrayOracle(t, "21", 4, 6, 6, 6, 5)
}

// Not parallel: refreshing writes this group's measured rows.
func TestCheckedViewRanked21ArrayCounts(t *testing.T) {
	originalRankedArrayCountTest(t, "21")
}
