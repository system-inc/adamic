package scanall

import "testing"

// The ordinary build has no measurement APIs in internal/lower. The documented
// overlay build replaces this bridge and runs the same eight named tests.
func scanAllControl(t *testing.T, name string, bind, record, completed bool) {
	t.Helper()
	t.Skip("scan-all controls require the generated overlay; see README.md")
}
