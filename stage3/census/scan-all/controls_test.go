package scanall

import (
	"os"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
)

func scanAllControl(t *testing.T, name string) {
	t.Helper()
	binary, production := scanAllProduct(t, true), scanAllProduct(t, false)
	command, release := boundedrun.Command(55*time.Second, "python3", "controls.py", name)
	defer release()
	command.Env = append(os.Environ(), "SCAN_ALL_BINARY="+binary, "SCAN_ALL_PRODUCTION="+production)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %v\n%s", name, err, output)
	}
	t.Logf("%s", output)
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
