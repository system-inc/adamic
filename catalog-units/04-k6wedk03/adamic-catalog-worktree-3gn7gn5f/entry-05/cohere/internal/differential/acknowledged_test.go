package differential

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// acknowledgementSource is a tree of two small files, read the way Run reads the real one.
var acknowledgementSource = map[string][]string{
	"app/Probe.tsx": {
		"import { read } from './read';",
		"const stored = window.localStorage;",
		"export const other = window.localStorage;",
		"export function Probe() { return read(stored); }",
	},
	"app/Other.tsx": {
		"const stored = window.localStorage;",
	},
}

func readAcknowledgementSource(file string, line int) (string, bool) {
	lines := acknowledgementSource[file]
	if line < 1 || line > len(lines) {
		return "", false
	}
	return lines[line-1], true
}

// compareWithAcknowledgements runs Compare over cohere-only findings with one acknowledgement.
func compareWithAcknowledgements(known []AcknowledgedDifference, findings ...Finding) Report {
	return Compare(Inputs{
		CohereFindings:   findings,
		CoherePopulation: Population{Findings: len(findings), FilesWalked: 2},
		GatePopulation:   Population{FilesWalked: 2},
		CohereRules:      map[string]bool{"storage-no-direct-local-storage": true, "no-self-assign": true},
		ConfiguredRules:  map[string]bool{"storage-no-direct-local-storage": true, "no-self-assign": true},
		Acknowledged:     known,
		ReadSourceLine:   readAcknowledgementSource,
	})
}

var storedAcknowledgement = AcknowledgedDifference{
	File: "app/Probe.tsx", Rule: "storage-no-direct-local-storage", Side: SideCohere,
	Anchor: "const stored = window.localStorage", Reason: "the gate cannot see this shape",
}

// An acknowledged difference is excused by the text on its line, wherever the line has moved to.
func TestAnAcknowledgementExcusesTheFindingItsAnchorNames(t *testing.T) {
	t.Parallel()
	report := compareWithAcknowledgements([]AcknowledgedDifference{storedAcknowledgement},
		Finding{File: "app/Probe.tsx", Line: 2, Rule: "storage-no-direct-local-storage"})
	if got := report.Differences[0].Classification; got != ClassificationAcknowledged {
		t.Fatalf("the anchored finding was classified %q rather than acknowledged", got)
	}
	if len(report.StaleAcknowledgements) != 0 || len(report.AmbiguousAcknowledgements) != 0 {
		t.Fatalf("an entry that excused its one finding was reported out of date: %v %v",
			report.StaleAcknowledgements, report.AmbiguousAcknowledgements)
	}
}

// The other direction, and it is the one that matters: an acknowledgement must not excuse anything
// else in the same rule, file, or line.
//
// Without this, an index that matched on rule alone would pass the test above while silently
// excusing every future drift in that rule, which is the failure this whole package exists to
// prevent. The excuse has to be exactly as narrow as the reason written beside it. Each finding here
// is the only one in its run, so the entry is also stale, which is the run failing as it should.
func TestAnAcknowledgementExcusesNothingElse(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name    string
		finding Finding
		side    Side
	}{
		{"a line in the same file without the anchor", Finding{File: "app/Probe.tsx", Line: 4, Rule: "storage-no-direct-local-storage"}, SideCohere},
		{"the same text in a different file", Finding{File: "app/Other.tsx", Line: 1, Rule: "storage-no-direct-local-storage"}, SideCohere},
		{"a different rule on the anchored line", Finding{File: "app/Probe.tsx", Line: 2, Rule: "no-self-assign"}, SideCohere},
		{"the same finding from the other side", Finding{File: "app/Probe.tsx", Line: 2, Rule: "storage-no-direct-local-storage"}, SideGate},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			inputs := Inputs{
				CoherePopulation: Population{FilesWalked: 2},
				GatePopulation:   Population{FilesWalked: 2},
				CohereRules:      map[string]bool{"storage-no-direct-local-storage": true, "no-self-assign": true},
				ConfiguredRules:  map[string]bool{"storage-no-direct-local-storage": true, "no-self-assign": true},
				Acknowledged:     []AcknowledgedDifference{storedAcknowledgement},
				ReadSourceLine:   readAcknowledgementSource,
			}
			if testCase.side == SideCohere {
				inputs.CohereFindings = []Finding{testCase.finding}
			} else {
				inputs.GateFindings = []Finding{testCase.finding}
			}
			report := Compare(inputs)
			if got := report.Differences[0].Classification; got == ClassificationAcknowledged {
				t.Error("an unrelated difference was excused by the acknowledgement")
			}
			if len(report.StaleAcknowledgements) != 1 {
				t.Errorf("the entry excused nothing in this run, so it must be stale, got %v", report.StaleAcknowledgements)
			}
		})
	}
}

// An entry that names no current difference is stale, and a stale entry fails a run that would
// otherwise agree: it is an excuse waiting for whatever lands on its site next.
func TestAStaleAcknowledgementFailsTheRun(t *testing.T) {
	t.Parallel()
	report := compareWithAcknowledgements([]AcknowledgedDifference{storedAcknowledgement})
	report.Provenance = provenProvenance()
	if len(report.StaleAcknowledgements) != 1 || report.StaleAcknowledgements[0].Key() != storedAcknowledgement.Key() {
		t.Fatalf("an entry matching nothing was not reported stale: %v", report.StaleAcknowledgements)
	}
	if report.Agreed() {
		t.Fatal("a run with a stale acknowledgement agreed")
	}

	report.StaleAcknowledgements = nil
	if !report.Agreed() {
		t.Fatal("the control: the same run without the stale entry must agree, or the test above proves nothing")
	}

	var out strings.Builder
	stale := compareWithAcknowledgements([]AcknowledgedDifference{storedAcknowledgement})
	stale.Provenance = provenProvenance()
	Write(&out, stale)
	if !strings.Contains(out.String(), "stale      "+storedAcknowledgement.Key()) ||
		!strings.Contains(out.String(), "1 acknowledgements excuse no single current difference") {
		t.Fatalf("the report must name the stale entry and say why the run failed, got:\n%s", out.String())
	}
}

// An anchor that names several differences excuses none of them, since it would excuse more than its
// reason was written about.
func TestAnAmbiguousAcknowledgementExcusesNone(t *testing.T) {
	t.Parallel()
	broad := storedAcknowledgement
	broad.Anchor = "window.localStorage"
	report := compareWithAcknowledgements([]AcknowledgedDifference{broad},
		Finding{File: "app/Probe.tsx", Line: 2, Rule: "storage-no-direct-local-storage"},
		Finding{File: "app/Probe.tsx", Line: 3, Rule: "storage-no-direct-local-storage"})
	for _, difference := range report.Differences {
		if difference.Classification == ClassificationAcknowledged {
			t.Errorf("line %d was excused by an anchor that names two findings", difference.Finding.Line)
		}
	}
	if len(report.AmbiguousAcknowledgements) != 1 {
		t.Fatalf("the entry naming two findings was not reported ambiguous: %v", report.AmbiguousAcknowledgements)
	}
	report.Provenance = provenProvenance()
	if report.Agreed() {
		t.Fatal("a run with an ambiguous acknowledgement agreed")
	}
}

// A comparison that cannot read the tree cannot tell an anchor that holds from one that moved, so it
// trusts none of them.
func TestWithoutTheSourceEveryAcknowledgementIsStale(t *testing.T) {
	t.Parallel()
	report := Compare(Inputs{
		CohereFindings:   []Finding{{File: "app/Probe.tsx", Line: 2, Rule: "storage-no-direct-local-storage"}},
		CoherePopulation: Population{Findings: 1, FilesWalked: 2},
		GatePopulation:   Population{FilesWalked: 2},
		Acknowledged:     []AcknowledgedDifference{storedAcknowledgement},
	})
	if report.Differences[0].Classification == ClassificationAcknowledged {
		t.Fatal("a finding was excused without its line ever being read")
	}
	if len(report.StaleAcknowledgements) != 1 {
		t.Fatalf("the unread entry must be stale, got %v", report.StaleAcknowledgements)
	}
}

// Every acknowledgement states a reason and an anchor, because one without a reason is
// indistinguishable from a suppression added to turn a red run green, and one without an anchor would
// match every line of its file.
func TestEveryKnownGateDefectStatesAReasonAndAnAnchor(t *testing.T) {
	t.Parallel()
	if len(KnownGateDefects) == 0 {
		t.Skip("no acknowledgements are recorded, so this proves nothing")
	}
	for _, defect := range KnownGateDefects {
		if defect.Reason == "" {
			t.Errorf("%s has no reason recorded", defect.Key())
		}
		if defect.File == "" || defect.Rule == "" || strings.TrimSpace(defect.Anchor) == "" {
			t.Errorf("%s is not keyed to an exact finding, so it would excuse more than one", defect.Key())
		}
	}
}

// Run reads the tree's lines 1-based, as both gates number them, and refuses a line past the end
// rather than handing back an empty one an anchor could never match anyway.
func TestSourceLineReaderReadsTheTreesLinesOneBased(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "app", "Probe.tsx"), []byte("first\nsecond\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	read := sourceLineReaderAt(root)
	for _, testCase := range []struct {
		file     string
		line     int
		text     string
		readable bool
	}{
		{"app/Probe.tsx", 1, "first", true},
		{"app/Probe.tsx", 2, "second", true},
		{"app/Probe.tsx", 0, "", false},
		{"app/Probe.tsx", 4, "", false},
		{"app/Missing.tsx", 1, "", false},
	} {
		text, readable := read(testCase.file, testCase.line)
		if text != testCase.text || readable != testCase.readable {
			t.Errorf("%s:%d read %q, %v; expected %q, %v", testCase.file, testCase.line, text, readable, testCase.text, testCase.readable)
		}
	}
}
