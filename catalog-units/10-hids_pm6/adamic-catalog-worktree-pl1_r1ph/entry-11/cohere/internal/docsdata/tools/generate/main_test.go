package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/docsdata"
	"github.com/system-inc/cohere/internal/docsdata/capture"
)

// moduleRoot is the module's root: go test runs in this package's directory, four below it.
const moduleRoot = "../../../.."

// moduleCopy copies every file a check reads into a fresh directory: docs/data, and the sources beside it
// that SourceInputs reads. A test plants a change in the copy and never touches the module's own files.
func moduleCopy(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	copyFile := func(relative string) {
		contents, err := os.ReadFile(filepath.Join(moduleRoot, relative))
		if err != nil {
			t.Fatal(err)
		}
		destination := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destination, contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, relative := range []string{"swift/HouseRuleVerdicts.json", "swift/Rules.json", "CHANGELOG.md"} {
		copyFile(relative)
	}
	records, err := filepath.Glob(filepath.Join(moduleRoot, "bench", "results", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		copyFile(filepath.Join("bench", "results", filepath.Base(record)))
	}
	err = filepath.WalkDir(filepath.Join(moduleRoot, docsdata.Directory), func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(moduleRoot, path)
		if err != nil {
			return err
		}
		copyFile(relative)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// recordsReproducing is capture records from which PickExamples picks exactly these examples: the shown
// cases as they were asserted, and every other message id and fix kind as a case run with options, which
// counts and is never shown.
func recordsReproducing(examples docsdata.Examples) []capture.Record {
	var records []capture.Record
	if firing := examples.Firing; firing != nil {
		outcome := capture.OutcomeFindings
		if firing.FixedSource != "" {
			outcome = capture.OutcomeFixed
		}
		records = append(records, capture.Record{Rule: examples.Rule, File: firing.File, Source: firing.Source,
			Outcome: outcome, Findings: firing.Findings, FixedSource: firing.FixedSource})
	}
	if clean := examples.Clean; clean != nil {
		records = append(records, capture.Record{Rule: examples.Rule, File: clean.File, Source: clean.Source, Outcome: capture.OutcomeClean})
	}
	suggestions := 0
	if examples.AssertsSuggestion {
		suggestions = 1
	}
	for _, id := range examples.AssertedMessageIds {
		records = append(records, capture.Record{Rule: examples.Rule, File: "options.ts", Source: id, Options: json.RawMessage(`{}`),
			Outcome: capture.OutcomeFindings, Findings: []capture.Finding{{MessageId: id, Fix: examples.AssertsFix, Suggestions: suggestions}}})
	}
	if len(records) == 0 {
		records = append(records, capture.Record{Rule: examples.Rule, File: "options.ts", Options: json.RawMessage(`{}`), Outcome: capture.OutcomeClean})
	}
	return records
}

// lintTestsAsCommitted is a capture whose lint tests assert exactly the module's committed examples: what
// the real capture records when the committed examples are current.
func lintTestsAsCommitted(t *testing.T) func(directory string) error {
	t.Helper()
	committed, err := docsdata.ReadCommittedExamples(moduleRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(committed) == 0 {
		t.Fatal("the module has no committed examples, so a check against them would prove nothing")
	}
	return func(directory string) error {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return err
		}
		var lines []string
		for _, examples := range committed {
			for _, record := range recordsReproducing(examples) {
				encoded, err := json.Marshal(record)
				if err != nil {
					return err
				}
				lines = append(lines, string(encoded))
			}
		}
		sort.Strings(lines)
		return os.WriteFile(filepath.Join(directory, "capture-1.jsonl"), []byte(strings.Join(lines, "\n")+"\n"), 0o644)
	}
}

// noHelp leaves cli.json out of the check, since building the binary is not what this test is about.
func noHelp(string) ([]byte, map[string][]byte, error) { return nil, nil, nil }

// plantStaleExample rewrites one committed example's firing source in root, as an example looks when its
// rule's tests changed and nobody recaptured, and returns its path.
func plantStaleExample(t *testing.T, root string) string {
	t.Helper()
	committed, err := docsdata.ReadCommittedExamples(root)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for name, examples := range committed {
		if examples.Firing != nil {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	planted := committed[names[0]]
	firing := *planted.Firing
	firing.Source += "// a case the rule's tests no longer assert\n"
	planted.Firing = &firing
	encoded, err := json.MarshalIndent(planted, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	path := docsdata.ExamplePath(planted.Rule)
	if err := os.WriteFile(filepath.Join(root, path), append(encoded, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestCheckRecapturesTheExamples is #d7vx76v in both directions: -check passes a tree whose examples are
// what the lint tests assert, and names an example planted stale. The rendering depth, which reads the
// examples back, passes the planted tree, which is why it must never be the one called a check.
func TestCheckRecapturesTheExamples(t *testing.T) {
	t.Parallel()
	capturing := lintTestsAsCommitted(t)

	t.Run("a fresh tree passes -check", func(t *testing.T) {
		t.Parallel()
		root := moduleCopy(t)
		stale, err := staleFiles(sources{root: root, help: noHelp, capture: capturing}, t.TempDir(), true)
		if err != nil {
			t.Fatal(err)
		}
		if len(stale) > 0 {
			t.Errorf("a copy of the module's own files is stale: %v", stale)
		}
	})

	t.Run("a planted stale example fails -check", func(t *testing.T) {
		t.Parallel()
		root := moduleCopy(t)
		planted := plantStaleExample(t, root)
		stale, err := staleFiles(sources{root: root, help: noHelp, capture: capturing}, t.TempDir(), true)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Contains(stale, planted) {
			t.Errorf("-check missed the planted %s: %v", planted, stale)
		}
	})

	t.Run("the rendering depth cannot see a stale example", func(t *testing.T) {
		t.Parallel()
		root := moduleCopy(t)
		plantStaleExample(t, root)
		noCapture := func(string) error {
			t.Error("the rendering depth ran the lint tests")
			return nil
		}
		stale, err := staleFiles(sources{root: root, help: noHelp, capture: noCapture}, t.TempDir(), false)
		if err != nil {
			t.Fatal(err)
		}
		if len(stale) > 0 {
			t.Errorf("the rendering depth reads the examples back, so it should find nothing stale here: %v", stale)
		}
	})

	t.Run("a capture that recorded nothing is refused", func(t *testing.T) {
		t.Parallel()
		recordsNothing := func(directory string) error { return os.MkdirAll(directory, 0o755) }
		_, err := staleFiles(sources{root: moduleCopy(t), help: noHelp, capture: recordsNothing}, t.TempDir(), true)
		if err == nil || !strings.Contains(err.Error(), "recorded no case") {
			t.Errorf("an empty capture passed as a check: %v", err)
		}
	})
}
