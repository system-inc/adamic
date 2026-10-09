package oracle

import "testing"

func TestProgramRegionTscNodesAgreeWithNode(t *testing.T) {
	t.Parallel()
	counts := programRegionFixture(t, "tsc_node_membership.a")
	if counts[5] != 7 {
		t.Fatalf("tsc Node graph regions %d, want symbol, three nodes and three arrays", counts[5])
	}
}
