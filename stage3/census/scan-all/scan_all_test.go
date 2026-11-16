package scanall

import (
	"os/exec"
	"testing"
)

// Compiler consumer: github.com/system-inc/adamic/internal/lower.
func TestScanAllMainAnchorsMatchExactly(t *testing.T) {
	t.Parallel()
	out, err := exec.Command("python3", "make_overlay.py", "--check").CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
}
