package fixtures

import "testing"

func TestFixturesAssignmentProofs(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "assignment-proofs")
}
