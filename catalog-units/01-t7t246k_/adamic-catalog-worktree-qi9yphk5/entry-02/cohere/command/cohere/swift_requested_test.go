package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// renderRequested renders a stream for a command line that asked for the phases requested names, the way
// runSwiftEngine does, and returns what renderRecords returns.
func renderRequested(t *testing.T, lines []string, requested map[phaseName]bool, engineExit int) (string, int, error) {
	t.Helper()
	var out bytes.Buffer
	run := newSwiftRun(&out, swiftModeCheck, "", time.Now())
	run.requested = requested
	for index, line := range lines {
		if err := run.accept([]byte(line)); err != nil {
			return out.String(), 1, &recordError{index: index, err: err}
		}
	}
	exitCode, err := run.finish(engineExit, "exited")
	return out.String(), exitCode, err
}

// flagsOf states a command line's boolean flags for swiftRequestedPhases.
func flagsOf(names ...string) (map[string]bool, func(string) string) {
	given := map[string]bool{}
	for _, name := range names {
		given[name] = true
	}
	return given, func(name string) string {
		if given[name] {
			return "true"
		}
		return "false"
	}
}

// editRecords rewrites the records of a fixture: drop removes a record, and edit changes one in place.
func editRecords(t *testing.T, lines []string, edit func(record map[string]any) (keep bool)) []string {
	t.Helper()
	edited := []string{}
	for _, line := range lines {
		record := map[string]any{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		if !edit(record) {
			continue
		}
		encoded, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		edited = append(edited, string(encoded))
	}
	return edited
}

// A command line asks for the phases the TypeScript run would run for it.
func TestSwiftRequestedPhasesReadTheFlagsAsTheTypeScriptRunDoes(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		flags []string
		want  string
	}{
		{nil, "fix types lint"},
		{[]string{"no-fix"}, "fix types lint"},
		{[]string{"lint"}, "lint"},
		{[]string{"types"}, "types"},
		{[]string{"fix"}, "fix"},
		{[]string{"fix", "lint"}, "fix lint"},
		{[]string{"unused"}, "fix types lint unused"},
		{[]string{"lint", "unused-deep"}, "lint unused"},
	} {
		requested := swiftRequestedPhases(flagsOf(testCase.flags...))
		names := []string{}
		for _, name := range phaseOrder {
			if requested[name] {
				names = append(names, string(name))
			}
		}
		if got := strings.Join(names, " "); got != testCase.want {
			t.Errorf("cohere %v asks for %q, want %q", testCase.flags, got, testCase.want)
		}
	}
}

// `cohere --lint` on a Swift package: the engine skips fix and types as not requested and calls the run
// complete, which it is (swift/Contract.md, `complete`). LintOnly.jsonl is a real stream of that run.
// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
func TestALintRunThatSkipsWhatItWasNotAskedForIsComplete(t *testing.T) {
	lint := swiftRequestedPhases(flagsOf("lint"))

	output, exitCode, err := renderRequested(t, contractFixture(t, "LintOnly.jsonl"), lint, 1)
	if err != nil || exitCode != 1 {
		t.Fatalf("a --lint run with two findings: exit %d, err %v\n%s", exitCode, err, output)
	}
	requireLines(t, output,
		"[cohere-swift/toolchain-require-upcoming-features/upcomingFeatureMissing]",
		"phases: fix skipped (not requested) · types skipped (not requested) · lint ran",
	)

	// The same run with nothing found passes.
	clean := editRecords(t, contractFixture(t, "LintOnly.jsonl"), func(record map[string]any) bool {
		switch record["kind"] {
		case "finding":
			return false
		case "lint", "summary":
			record["findings"] = 0
			if record["kind"] == "summary" {
				record["exitCode"] = 0
			}
		}
		return true
	})
	output, exitCode, err = renderRequested(t, clean, lint, 0)
	if err != nil || exitCode != 0 {
		t.Fatalf("a clean --lint run: exit %d, err %v\n%s", exitCode, err, output)
	}
}

// A phase the caller asked for that comes back skipped still fails, whatever the engine calls it.
// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
func TestARequestedPhaseThatComesBackSkippedStillFails(t *testing.T) {
	// Skipped as not requested when it was requested: a bare run reported as though it were --lint. The
	// engine is refused rather than believed.
	_, _, err := renderRequested(t, contractFixture(t, "LintOnly.jsonl"), swiftRequestedPhases(flagsOf()), 1)
	if err == nil || !strings.Contains(err.Error(), "the fix phase was asked for, and the engine skipped it as \"not requested\"") {
		t.Fatalf("a bare run whose fix came back not requested was accepted: %v", err)
	}

	// Skipped for any other reason, under --lint, while calling the run complete: the contradiction the
	// summary check exists for.
	otherSkip := func(complete bool) []string {
		return editRecords(t, contractFixture(t, "LintOnly.jsonl"), func(record map[string]any) bool {
			if record["kind"] == "phase" && record["name"] == "lint" {
				record["outcome"] = "skipped"
				record["detail"] = "no rule was configured"
			}
			if record["kind"] == "summary" {
				record["complete"] = complete
			}
			return true
		})
	}
	lint := swiftRequestedPhases(flagsOf("lint"))
	_, _, err = renderRequested(t, otherSkip(true), lint, 1)
	if err == nil || !strings.Contains(err.Error(), "calls the run complete") {
		t.Fatalf("a requested lint skipped for another reason, called complete, was accepted: %v", err)
	}

	// Called incomplete instead, it is believed and fails as an incomplete run.
	output, exitCode, err := renderRequested(t, otherSkip(false), lint, 1)
	if err != nil || exitCode != 1 || !strings.Contains(output, "this run did not check everything") {
		t.Fatalf("a requested lint skipped and called incomplete: exit %d, err %v\n%s", exitCode, err, output)
	}
}
