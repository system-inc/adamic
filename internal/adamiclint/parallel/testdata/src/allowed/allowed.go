// Nothing here may be flagged: each test calls t.Parallel() first or says why, or is no test.
package allowed

import "testing"

// A file that isn't a _test.go file holds no tests, whatever its functions are called.
func TestShapedOutsideATestFile(t *testing.T) {}
