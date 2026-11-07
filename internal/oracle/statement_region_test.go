package oracle

import "testing"

func TestStatementRegionEnd(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "region_end.a")
}

func TestStatementRegionBig(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "regions_big.a")
}

func TestStatementRegionEscapes(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "regions_escapes.a")
}

func TestStatementRegionPaths(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "regions_paths.a")
}

func TestStatementRegionSpreadFresh(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "regions_spread_fresh.a")
}
