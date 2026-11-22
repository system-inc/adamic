package cloud

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// agreement-daily's diff_test.py, run here so a gate runs it, as meter-daily's outcome_test.py is (#wfnmhe9). It holds
// diff.py's read of a day out of its shards' job logs, round trip and refusals, and its compare of two days: each
// section, a void that never moves, and the exit that opens a task. It needs nothing but python3: temporary logs and
// reports, no network.
func TestAgreementDailyDiff(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "agreement-daily-diff.txt")
	output, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	command := toolchain.Command("python3", "-B", "-m", "unittest", "-v", "diff_test")
	command.Dir = "agreement-daily"
	command.Stdout = output
	command.Stderr = output
	runError := command.Run()
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if runError != nil {
		t.Fatalf("agreement-daily diff tests: %v\n%s", runError, data)
	}
	t.Logf("agreement-daily diff tests passed:\n%s", data)
}
