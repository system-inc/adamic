package cloud

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// meter-daily's outcome_test.py, run here so a gate runs it: a suite only a hand runs can't fail (#mpesmv7, Correctness
// on #ny0c3gb). It holds which outcome run.sh records for each meter exit, OverBudget and BudgetUnknown apart from
// Failed, and which outcomes submit.sh commits. It needs nothing but python3: temporary report.json files, no network.
func TestMeterDailyOutcomes(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "meter-daily-outcomes.txt")
	output, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	command := toolchain.Command("python3", "-B", "-m", "unittest", "-v", "outcome_test")
	command.Dir = "meter-daily"
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
		t.Fatalf("meter-daily outcome tests: %v\n%s", runError, data)
	}
	t.Logf("meter-daily outcome tests passed:\n%s", data)
}
