package fixtures

import "testing"

// The directory owner is the self-compare member from repair a233eec3.
func TestFixturesSelfCompare(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "self-compare")
}
