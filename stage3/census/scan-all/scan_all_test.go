package scanall

import (
	"os/exec"
	"testing"
)

// Compiler consumer: github.com/system-inc/adamic/internal/lower.
// The gate runs only the anchor check. The eight controls need two binaries built
// outside any test unit, so verify.sh runs them through controls.py, never as a
// test that skips without them (the skip census refuses such a skip).
func TestScanAllMainAnchorsMatchExactly(t *testing.T) {
	t.Parallel()
	out, err := exec.Command("python3", "make_overlay.py", "--check").CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
}
