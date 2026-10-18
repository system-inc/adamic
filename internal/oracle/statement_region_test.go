package oracle

import "testing"

func TestStatementRegionEnd(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "region_end.a")
}
