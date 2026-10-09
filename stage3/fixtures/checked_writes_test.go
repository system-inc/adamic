package fixtures

import "testing"

func TestFixturesCheckedWrites(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "checked-writes")
}
