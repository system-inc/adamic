package main

import (
	"testing"

	"github.com/system-inc/adamic/internal/fuzz"
)

// The exit status is the worst that happened: a crash outranks any number of findings, and a run
// with neither exits 0 however many programs weren't fit to judge.
func TestExitStatusSaysTheWorst(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		verdicts map[fuzz.Verdict]int
		want     int
	}{
		{"nothing found", map[fuzz.Verdict]int{fuzz.Agreed: 90, fuzz.Checked: 4, fuzz.NotYet: 3, fuzz.Invalid: 2, fuzz.Unfit: 1}, 0},
		{"a finding", map[fuzz.Verdict]int{fuzz.Agreed: 99, fuzz.Finding: 1}, 1},
		{"a crash", map[fuzz.Verdict]int{fuzz.Agreed: 99, fuzz.Crash: 1}, 3},
		{"a crash among findings", map[fuzz.Verdict]int{fuzz.Finding: 40, fuzz.Crash: 1}, 3},
	} {
		if got := exitStatus(test.verdicts); got != test.want {
			t.Errorf("%s: exit %d, want %d", test.name, got, test.want)
		}
	}
	if severity(fuzz.Crash) >= severity(fuzz.Finding) {
		t.Errorf("a crash ranks %d and a finding %d: the crash has to come first", severity(fuzz.Crash), severity(fuzz.Finding))
	}
}
