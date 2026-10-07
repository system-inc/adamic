package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Read-only stock TypeScript audit of all measured refusals, including NotYet
// diagnostics that the runner's checker-code branch does not ask tsc about.
// This inventory does not change any runner verdict or classification.
func TestRegExpRemainingTypeScriptAudit(t *testing.T) {
	results, corpus, output := os.Getenv("ADAMIC_REGEX_REMAINING_RESULTS"), os.Getenv("ADAMIC_REGEX_REMAINING_CORPUS"), os.Getenv("ADAMIC_REGEX_REMAINING_AUDIT")
	if results == "" || corpus == "" || output == "" {
		t.Skip("explicit survey, corpus and audit output required")
	}
	data, err := os.ReadFile(results)
	if err != nil {
		t.Fatal(err)
	}
	oracle, err := startTypescript("../..", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer oracle.close()
	type audited struct {
		Path, Reason string
		Codes        []string
		Program      string
	}
	var rows []audited
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		var one result
		if err := json.Unmarshal([]byte(line), &one); err != nil {
			t.Fatal(err)
		}
		if one.Kind != outcomeRefused {
			continue
		}
		source, err := os.ReadFile(filepath.Join(corpus, "test", one.Path))
		if err != nil {
			t.Fatal(err)
		}
		test := classify(one.Path, string(source), true)
		if test.Skip != "" {
			t.Fatalf("measured refusal now skipped: %s", one.Path)
		}
		codes, err := oracle.check(test.Program)
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, audited{one.Path, one.Reason, codes, test.Program})
	}
	encoded, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, append(encoded, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("audited %d refusals with stock TypeScript %s", len(rows), oracle.stats.Version)
}
