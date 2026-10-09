package main

import (
	"bytes"
	"testing"
	"time"
)

// TestASwiftWarningIsAWarning holds the contract's spelling: a compiler warning is `warning`, and the
// view spells it warn, as the settings do; an error stays an error. Both count as type errors, since
// both gate.
// Not parallel: it swaps the package-level activeOutput and activeSummary, and acceptFinding counts into activeSummary
func TestASwiftWarningIsAWarning(t *testing.T) {
	saved, savedSummary := activeOutput, activeSummary
	defer func() { activeOutput, activeSummary = saved, savedSummary }()
	activeOutput, activeSummary = outputSettings{Mode: outputHuman}, runSummary{}

	var out bytes.Buffer
	run := newSwiftRun(&out, swiftModeCheck, "", time.Now())
	for _, severity := range []string{"warning", "error"} {
		record := &swiftFindingRecord{Source: "compiler", File: "Sources/A.swift", Line: 3, Column: 7,
			Severity: severity, Rule: "VariableNeverMutated", Message: "never mutated"}
		if err := run.acceptFinding(record); err != nil {
			t.Fatal(err)
		}
	}
	want := "Sources/A.swift:3:7 warn #VariableNeverMutated never mutated\n" +
		"Sources/A.swift:3:7 error #VariableNeverMutated never mutated\n"
	if out.String() != want {
		t.Errorf("findings:\n got  %q\n want %q", out.String(), want)
	}
	if activeSummary.TypeErrors != 2 {
		t.Errorf("counted %d type errors, want both", activeSummary.TypeErrors)
	}
}

// TestASwiftGapSaysItself is the footer's wording for files nothing checked: the count and the kinds,
// since the notes that name the files are --verbose's.
func TestASwiftGapSaysItself(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		withoutRecord, unreadable int
		want                      string
	}{
		{1, 1, "2 files nothing checked: 1 with no compiler record, 1 unreadable"},
		{1, 0, "1 file nothing checked: 1 with no compiler record"},
		{0, 3, "3 files nothing checked: 3 unreadable"},
		{0, 0, ""},
	} {
		if got := swiftUncheckedFiles(testCase.withoutRecord, testCase.unreadable); got != testCase.want {
			t.Errorf("swiftUncheckedFiles(%d, %d) = %q, want %q", testCase.withoutRecord, testCase.unreadable, got, testCase.want)
		}
	}
}
