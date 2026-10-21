package flagged_test

import . "testing"

// The rule knows *testing.T by its type, so a dot import of testing is the same test.
func TestDotImported(t *T) { // want `a test runs in parallel unless it says why`
	t.Log("serial")
}
