package docsdata

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/system-inc/cohere/internal/docsdata/capture"
)

// Examples is examples/<rule>.json: the cases a rule's own tests assert, the shortest of each kind, and
// what every asserted case says about the rule.
type Examples struct {
	Rule string `json:"rule"`
	// Firing is the shortest case the tests assert findings on, with its fixed source when the rule's
	// tests assert one. Clean is the shortest case they assert silence on.
	Firing *Example `json:"firing,omitempty"`
	Clean  *Example `json:"clean,omitempty"`
	// AssertedMessageIds is every message id any asserted case reported, sorted.
	AssertedMessageIds []string `json:"assertedMessageIds,omitempty"`
	// AssertsFix and AssertsSuggestion are whether any asserted case carried a fix, or a suggestion.
	AssertsFix        bool `json:"assertsFix,omitempty"`
	AssertsSuggestion bool `json:"assertsSuggestion,omitempty"`
}

// Example is one asserted case, run with no options.
//
// It has no options field on purpose. A test hands a rule its decoded options value, never the JSON a
// settings file writes, and for most rules the two differ: a list decoder's struct has no settings-file
// shape at all, and many decoders normalize what they read ("type-annotation" decodes to TypeAnnotation,
// an untagged field encodes under its Go name). Measured 2026-10-03: of 96 shown cases that ran with
// options, 157 option keys were Go field names. So a case run with options is never shown, and the field
// that could carry them does not exist.
type Example struct {
	File        string            `json:"file"`
	Source      string            `json:"source"`
	Findings    []capture.Finding `json:"findings,omitempty"`
	FixedSource string            `json:"fixedSource,omitempty"`
}

// MessageIds is what rules.json reports as the rule's message ids.
func (examples Examples) MessageIds() []string {
	return examples.AssertedMessageIds
}

// FixKind is what rules.json reports as the rule's fix kind, empty when no asserted case shows either.
func (examples Examples) FixKind() string {
	switch {
	case examples.AssertsFix && examples.AssertsSuggestion:
		return "both"
	case examples.AssertsFix:
		return "fix"
	case examples.AssertsSuggestion:
		return "suggestion"
	}
	return ""
}

// ReadCaptures reads every record the harness wrote into directory.
func ReadCaptures(directory string) ([]capture.Record, error) {
	paths, err := filepath.Glob(filepath.Join(directory, "capture-*.jsonl"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	var records []capture.Record
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 0, 1<<20), 1<<28)
		for scanner.Scan() {
			var record capture.Record
			if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
				file.Close()
				return nil, fmt.Errorf("docsdata: %s: %w", path, err)
			}
			records = append(records, record)
		}
		err = scanner.Err()
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("docsdata: %s: %w", path, err)
		}
	}
	return records, nil
}

// PickExamples chooses each rule's examples from its captured records, and returns how many distinct
// cases ran with options, which count toward a rule's message ids and fix kind and are never shown.
//
// The shortest case wins, among those a reader can run as shown: run with no options (see Example), and
// a case standing alone before one whose program held other files. A rule whose tests assert a fix shows
// a case with its fixed source. Ties break on the text, so the choice never depends on the order the
// tests ran in.
//
// One case asserted twice, or run by two tests, is one case: records are deduplicated by rule, file,
// source, options and outcome before anything is counted or picked.
func PickExamples(records []capture.Record) (map[string]Examples, int) {
	type caseKey struct{ rule, file, source, options, outcome string }
	seen := map[caseKey]bool{}
	byRule := map[string][]capture.Record{}
	withOptions := 0
	for _, record := range records {
		if record.Rule == "" {
			continue
		}
		key := caseKey{record.Rule, record.File, record.Source, string(record.Options), record.Outcome}
		if seen[key] {
			continue
		}
		seen[key] = true
		if len(record.Options) > 0 {
			withOptions++
		}
		byRule[record.Rule] = append(byRule[record.Rule], record)
	}

	picked := make(map[string]Examples, len(byRule))
	for name, cases := range byRule {
		examples := Examples{Rule: name}
		messageIds := map[string]bool{}
		var firing, clean, fixed []capture.Record
		for _, record := range cases {
			for _, finding := range record.Findings {
				messageIds[finding.MessageId] = true
				examples.AssertsFix = examples.AssertsFix || finding.Fix
				examples.AssertsSuggestion = examples.AssertsSuggestion || finding.Suggestions > 0
			}
			if len(record.Options) > 0 {
				continue
			}
			switch record.Outcome {
			case capture.OutcomeClean:
				clean = append(clean, record)
			case capture.OutcomeFindings:
				firing = append(firing, record)
			case capture.OutcomeFixed:
				fixed = append(fixed, record)
			}
		}
		for id := range messageIds {
			examples.AssertedMessageIds = append(examples.AssertedMessageIds, id)
		}
		sort.Strings(examples.AssertedMessageIds)
		if len(fixed) > 0 {
			firing = fixed
		}
		examples.Firing = shortest(firing)
		examples.Clean = shortest(clean)
		picked[name] = examples
	}
	return picked, withOptions
}

// shortest is the case PickExamples prefers among candidates, or nil when there are none.
func shortest(candidates []capture.Record) *Example {
	if len(candidates) == 0 {
		return nil
	}
	sort.Slice(candidates, func(left, right int) bool {
		a, b := candidates[left], candidates[right]
		if (a.OtherFiles > 0) != (b.OtherFiles > 0) {
			return a.OtherFiles == 0
		}
		if len(a.Source) != len(b.Source) {
			return len(a.Source) < len(b.Source)
		}
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		return a.File < b.File
	})
	best := candidates[0]
	return &Example{
		File:        best.File,
		Source:      best.Source,
		Findings:    best.Findings,
		FixedSource: best.FixedSource,
	}
}

// ReadCommittedExamples reads the example files under root back, keyed by rule name, for a run that
// does not recapture.
func ReadCommittedExamples(root string) (map[string]Examples, error) {
	examples := map[string]Examples{}
	directory := filepath.Join(root, ExamplesDirectory)
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
		if os.IsNotExist(err) && path == directory {
			return filepath.SkipDir
		}
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".json") {
			return err
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var example Examples
		if err := json.Unmarshal(contents, &example); err != nil {
			return fmt.Errorf("docsdata: %s: %w", path, err)
		}
		examples[example.Rule] = example
		return nil
	})
	return examples, err
}
