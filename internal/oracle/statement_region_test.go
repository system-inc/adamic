package oracle

import "testing"

func TestStatementRegionEnd(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "region_end.a")
}

func TestStatementRegionBig(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "regions_big.a")
}

func TestStatementRegionEscapes(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "regions_escapes.a")
}

func TestStatementRegionPaths(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "regions_paths.a")
}

func TestStatementRegionSpreadFresh(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "regions_spread_fresh.a")
}
