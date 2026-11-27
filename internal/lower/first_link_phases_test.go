package lower

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// firstLinkHarnessReport is the first-link harness's report, from this package's directory. Its BODY_PHASES and
// WHOLE_PROGRAM_PHASES name every phase a trap-mode record may carry, and a record whose phase isn't in either set
// fails the report, so the two sets and this package's calls have to agree (#nm7z8fv).
const firstLinkHarnessReport = "../../stage3/firstlink/report.py"

// firstLinkPhaseCalls are the four call shapes that name a first-link phase, the same patterns the harness's own
// test read before this moved here: a whole-program phase is the first argument of firstLinkPass or
// firstLinkProgramStop, and a body phase is firstLinkCatch's first argument, firstLinkRow's third, or the Phase of
// the ir.Trap a firstLinkRecord takes.
var firstLinkPhaseCalls = []struct {
	whole   bool
	pattern *regexp.Regexp
}{
	{true, regexp.MustCompile(`\.firstLink(?:Pass|ProgramStop)\s*\(\s*"([^"]+)"`)},
	{false, regexp.MustCompile(`\.firstLinkCatch\s*\(\s*"([^"]+)"`)},
	{false, regexp.MustCompile(`\.firstLinkRow\s*\([^,]*,[^,]*,\s*"([^"]+)"`)},
	{false, regexp.MustCompile(`\.firstLinkRecord\s*\(\s*ir\.Trap\{[^}]*\bPhase:\s*"([^"]+)"`)},
}

// firstLinkReportItem is one string in a set literal of report.py, which is read as text and never run.
var firstLinkReportItem = regexp.MustCompile(`"([^"]+)"`)

// firstLinkReportPhases is the named set in report's text, by literal parsing. A set that isn't there, or holds no
// string, fails the test, so the comparison can't pass empty.
func firstLinkReportPhases(t *testing.T, report, name string) []string {
	t.Helper()
	literal := regexp.MustCompile(`(?ms)^` + regexp.QuoteMeta(name) + `\s*=\s*\{(.*?)\}`).FindStringSubmatch(report)
	if literal == nil {
		t.Fatalf("%s holds no %s = {...} set", firstLinkHarnessReport, name)
	}
	var phases []string
	for _, item := range firstLinkReportItem.FindAllStringSubmatch(literal[1], -1) {
		phases = append(phases, item[1])
	}
	if len(phases) == 0 {
		t.Fatalf("%s's %s is empty", firstLinkHarnessReport, name)
	}
	slices.Sort(phases)
	return slices.Compact(phases)
}

// firstLinkLoweringPhases is every phase this package's non-test files name, body and whole-program, each sorted.
// Every *.go file here but the tests is read, so a phase moving to a new file is still seen. override, by file
// name, is read in place of a file's text, so a mutant needs no file of its own. A package that names no phase on
// either side fails the test.
func firstLinkLoweringPhases(t *testing.T, override map[string]string) (body, whole []string) {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	read := 0
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		text, overridden := override[file]
		if !overridden {
			contents, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			text = string(contents)
		}
		read++
		for _, call := range firstLinkPhaseCalls {
			for _, match := range call.pattern.FindAllStringSubmatch(text, -1) {
				if call.whole {
					whole = append(whole, match[1])
				} else {
					body = append(body, match[1])
				}
			}
		}
	}
	if len(body) == 0 || len(whole) == 0 {
		t.Fatalf("this package's %d non-test files name no first-link phase (body %d, whole-program %d)", read, len(body), len(whole))
	}
	slices.Sort(body)
	slices.Sort(whole)
	return slices.Compact(body), slices.Compact(whole)
}

// firstLinkPhaseGap says how the report's two sets differ from the phases lowering names, or is empty when they
// agree.
func firstLinkPhaseGap(reportBody, reportWhole, body, whole []string) string {
	var gaps []string
	for _, side := range []struct {
		name           string
		report, lowers []string
	}{{"BODY_PHASES", reportBody, body}, {"WHOLE_PROGRAM_PHASES", reportWhole, whole}} {
		for _, phase := range side.lowers {
			if !slices.Contains(side.report, phase) {
				gaps = append(gaps, side.name+" lacks "+phase+", which lowering names")
			}
		}
		for _, phase := range side.report {
			if !slices.Contains(side.lowers, phase) {
				gaps = append(gaps, side.name+" holds "+phase+", which lowering never names")
			}
		}
	}
	return strings.Join(gaps, "; ")
}

// The first-link harness's report names exactly the phases this package records: BODY_PHASES the four body call
// shapes', WHOLE_PROGRAM_PHASES the whole-program ones'. Mutants, each named: report.py's BODY_PHASES without
// Statics; a firstLinkPass("probeOnly", ...) call in a lowering file.
func TestFirstLinkPhasesMatchHarnessReport(t *testing.T) {
	t.Parallel()
	// Reads: stage3/firstlink/report.py
	contents, err := os.ReadFile(firstLinkHarnessReport)
	if err != nil {
		t.Fatal(err)
	}
	report := string(contents)
	reportBody, reportWhole := firstLinkReportPhases(t, report, "BODY_PHASES"), firstLinkReportPhases(t, report, "WHOLE_PROGRAM_PHASES")
	body, whole := firstLinkLoweringPhases(t, nil)
	if gap := firstLinkPhaseGap(reportBody, reportWhole, body, whole); gap != "" {
		t.Fatalf("the harness's report and this package differ: %s", gap)
	}

	dropped := `"Statics", `
	if strings.Count(report, dropped) != 1 {
		t.Fatalf("%s no longer holds %s once; the mutant can't be planted", firstLinkHarnessReport, dropped)
	}
	withoutStatics := firstLinkReportPhases(t, strings.Replace(report, dropped, "", 1), "BODY_PHASES")
	if gap := firstLinkPhaseGap(withoutStatics, reportWhole, body, whole); !strings.Contains(gap, "Statics") {
		t.Errorf("with Statics dropped from BODY_PHASES: got %q, want the gap to name Statics", gap)
	}

	passes, err := os.ReadFile("first_link_passes.go")
	if err != nil {
		t.Fatal(err)
	}
	planted := string(passes) + "\nfunc (l *lowering) firstLinkProbeOnly() error { return l.firstLinkPass(\"probeOnly\", false, nil, nil) }\n"
	plantedBody, plantedWhole := firstLinkLoweringPhases(t, map[string]string{"first_link_passes.go": planted})
	if gap := firstLinkPhaseGap(reportBody, reportWhole, plantedBody, plantedWhole); !strings.Contains(gap, "probeOnly") {
		t.Errorf("with a firstLinkPass(\"probeOnly\", ...) call planted: got %q, want the gap to name probeOnly", gap)
	}
}
