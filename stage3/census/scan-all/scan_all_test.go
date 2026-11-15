package scanall

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

// Compiler consumer: github.com/system-inc/adamic/internal/lower.
func scanAllControl(t *testing.T, name string) {
	t.Helper()
	if os.Getenv("SCAN_ALL_BINARY") == "" || os.Getenv("SCAN_ALL_PRODUCTION") == "" {
		t.Skip("run scan-all verify.sh to prepare the two binaries outside test units")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "python3", "controls.py", name).CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %v\n%s", name, err, out)
	}
	t.Log(string(out))
}
func TestScanAllMainAnchorsMatchExactly(t *testing.T) {
	t.Parallel()
	out, err := exec.Command("python3", "make_overlay.py", "--check").CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
}
func TestScanAllCountStopAfterEarlierScanStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "count-stop")
}
func TestScanAllCountCleanAfterEarlierScanStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "count-clean")
}
func TestScanAllRestStopAfterEarlierScanStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "rest-stop")
}
func TestScanAllRestCleanAfterEarlierScanStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "rest-clean")
}
func TestScanAllRecordStopAfterEarlierScanStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "record-stop")
}
func TestScanAllRecordCleanAfterEarlierScanStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "record-clean")
}
func TestScanAllCycleStopAfterEarlierScanStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "cycle-stop")
}
func TestScanAllCycleCleanAfterEarlierScanStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "cycle-clean")
}
