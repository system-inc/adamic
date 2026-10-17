package skipcensus

import (
	"os"
	"testing"
)

// Every pending row of census-extra.json is held to its reason here, as the census holds it on a real log
// (checkPending): it names the branch it awaits, and its message says so. Before this, a bad row passed every unit test
// and was caught only when a gate read a log with its skip in it (Planner's mutant on c5db9b09ec, #yzwm4rj). Whether the
// branch is still on origin is the gate's question, asked there of the candidate's origin. Mutant: a row whose message
// drops its awaits wording.
func TestEveryExtraPendingRowNamesWhatItAwaits(t *testing.T) {
	t.Parallel()
	f, err := os.Open("../../cloud/fast-gate/census-extra.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := Load(f)
	if err != nil {
		t.Fatal(err)
	}
	pending := 0
	for _, row := range rows {
		if row.Class != "pending" {
			continue
		}
		pending++
		if err := checkPending(row); err != nil {
			t.Errorf("%s: %v", row.ID, err)
		}
	}
	if pending == 0 {
		t.Fatal("census-extra.json has no pending row, so nothing was checked")
	}
	t.Logf("%d pending rows held to their reason", pending)
}
