package oracle

import "testing"

func TestProgramRegionTscNodesAgreeWithNode(test *testing.T) {
	test.Parallel()
	counts := programRegionFixture(test, "tsc_node_membership.a")
	if counts[5] != 7 {
		test.Fatalf("tsc Node graph regions %d, want symbol, three nodes and three arrays", counts[5])
	}
}
