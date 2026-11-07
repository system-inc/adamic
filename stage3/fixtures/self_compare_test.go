package fixtures

import "testing"

// The directory owner is the self-compare member from repair a233eec3.
func TestFixturesSelfCompareClosure(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "self-compare", "closure.a")
}

func TestFixturesSelfCompareArrayFill(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "self-compare", "array-fill.a")
}

func TestFixturesSelfCompareSweep(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "self-compare", "sweep.a")
}
