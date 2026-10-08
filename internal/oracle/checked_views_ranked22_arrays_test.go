package oracle

import "testing"

func TestCheckedViewRanked22OriginalArrays(t *testing.T) {
	originalRankedArrayOracle(t, "22", 1, 5)
}

// Not parallel: refreshing writes this group's measured rows.
func TestCheckedViewRanked22ArrayCounts(t *testing.T) {
	originalRankedArrayCountTest(t, "22")
}
