package differential

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A control the comparison did not report must come back missed, not detected.
//
// This is the assertion the whole runner exists to make, so it is also the one most worth proving
// can fail. A checkControls that returned Detected unconditionally would satisfy every happy-path
// test ever written against it and would silently bless a completely blind harness.
func TestAControlTheComparisonMissedIsReportedMissed(t *testing.T) {
	t.Parallel()
	planted := []Control{{
		Name:         "cohere-only",
		RelativePath: "control/Planted.ts",
		Rule:         "consistency-no-enum",
		ExpectedSide: SideCohere,
	}}

	// A report with no differences at all: the pipeline dropped it somewhere.
	missed := checkControls(planted, Report{}, nil, nil)
	if len(missed) != 1 {
		t.Fatalf("expected one control result, got %d", len(missed))
	}
	if missed[0].Detected {
		t.Fatal("a control absent from the differences must never be reported detected")
	}
	if !strings.Contains(missed[0].Detail, "consistency-no-enum") {
		t.Fatalf("the detail must name the rule that failed to fire, got %q", missed[0].Detail)
	}

	// The same control, this time actually reported. Without this the test above would pass against
	// a checkControls that can never detect anything.
	found := checkControls(planted, Report{Differences: []Difference{{
		Finding: Finding{File: "control/Planted.ts", Line: 3, Rule: "consistency-no-enum"},
		OnlyOn:  SideCohere,
	}}}, nil, nil)
	if !found[0].Detected {
		t.Fatalf("a control present in the differences on the expected side must be detected, got %q", found[0].Detail)
	}
}

// A control found on the wrong side is not a pass.
//
// This is the failure a naive check misses. The finding is there, the rule matches, the file
// matches — and the harness has proven the opposite of what it set out to prove, because the
// direction is the whole claim.
func TestAControlOnTheWrongSideIsNotDetected(t *testing.T) {
	t.Parallel()
	planted := []Control{{
		Name:         "cohere-only",
		RelativePath: "control/Planted.ts",
		Rule:         "consistency-no-enum",
		ExpectedSide: SideCohere,
	}}

	results := checkControls(planted, Report{Differences: []Difference{{
		Finding: Finding{File: "control/Planted.ts", Line: 3, Rule: "consistency-no-enum"},
		OnlyOn:  SideGate,
	}}}, nil, nil)

	if results[0].Detected {
		t.Fatal("a control that fired on the opposite side must not count as detected")
	}
	if !strings.Contains(results[0].Detail, "attributed to") {
		t.Fatalf("the detail must say which side actually saw it, got %q", results[0].Detail)
	}
}

// A missing coverage line must read as zero, so the vacuity guard refuses the run.
//
// The tempting bug here is a friendly default. Any plausible non-zero number would make an
// incomplete run look complete, which is the vacuous pass this package exists to prevent,
// reintroduced by the file that reports the evidence.
func TestAbsentCoverageReadsAsZeroRatherThanAGuess(t *testing.T) {
	t.Parallel()
	filesWalked, rulesRun := cohereCoverageFrom([]string{
		"graph built in 1.642s — 9973 files in the program, 3407 of them ours",
		"  note: rule consistency-no-utils-folder listened to no files",
	})
	if filesWalked != 0 || rulesRun != 0 {
		t.Fatalf("a summary with no lint coverage line must read as 0/0, got %d files and %d rules", filesWalked, rulesRun)
	}

	// The real line, so this test cannot pass by never parsing anything.
	filesWalked, rulesRun = cohereCoverageFrom([]string{
		"lint: 1 findings — 23 rules over 3407 files, 2098302 nodes visited, in 813ms",
	})
	if filesWalked != 3407 || rulesRun != 23 {
		t.Fatalf("expected 3407 files and 23 rules from the real coverage line, got %d and %d", filesWalked, rulesRun)
	}
}

// Planting must never clobber an existing file.
//
// On a tree several authors are working in, overwriting somebody's file to run a self-test would be
// a worse defect than the one being looked for.
func TestPlantingRefusesToOverwriteAnExistingFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	existing := filepath.Join(root, "Occupied.ts")
	if err := os.WriteFile(existing, []byte("export const mine = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, cleanup, err := plantControls(RunOptions{
		Root:     root,
		Controls: []Control{{Name: "clobber", RelativePath: "Occupied.ts", Contents: "export enum Bad {}\n"}},
	})
	cleanup()

	if err == nil {
		t.Fatal("planting over an existing file must be refused")
	}
	contents, readErr := os.ReadFile(existing)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(contents) != "export const mine = 1;\n" {
		t.Fatalf("the existing file was modified: %q", contents)
	}
}

// A planted control must exist during the run and be gone afterward.
func TestPlantedControlsAreWrittenThenRemoved(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	control := Control{
		Name:         "cohere-only",
		RelativePath: filepath.Join("nested", "Planted.ts"),
		Contents:     "export enum Planted { A = 'A' }\n",
		Rule:         "consistency-no-enum",
		ExpectedSide: SideCohere,
	}

	planted, cleanup, err := plantControls(RunOptions{Root: root, Controls: []Control{control}})
	if err != nil {
		t.Fatal(err)
	}
	if len(planted) != 1 {
		t.Fatalf("expected one planted control, got %d", len(planted))
	}

	absolutePath := filepath.Join(root, control.RelativePath)
	contents, err := os.ReadFile(absolutePath)
	if err != nil {
		t.Fatalf("the control was not written where the gates would lint it: %v", err)
	}
	if string(contents) != control.Contents {
		t.Fatalf("the control's contents did not survive planting: %q", contents)
	}

	cleanup()
	if _, err := os.Stat(absolutePath); !os.IsNotExist(err) {
		t.Fatal("the planted control was left behind, and the next person to run either gate would lint it")
	}
}

// A gate that could not be launched is an error, never an empty result.
//
// A process that never started produces empty stdout, which is byte-identical to a clean run.
func TestAGateThatCannotRunIsAnErrorNotSilence(t *testing.T) {
	t.Parallel()
	output, err := runGate(context.Background(), GateCommand{
		Name:      "missing",
		Program:   filepath.Join(t.TempDir(), "no-such-binary"),
		Directory: t.TempDir(),
	})
	if err == nil {
		t.Fatalf("a gate that does not exist must be an error, got output %q", output)
	}
	if !strings.Contains(err.Error(), "did not run") {
		t.Fatalf("the error must say the gate never ran, got %v", err)
	}
}

// A gate exiting non-zero is normal: that is what a gate with findings does.
func TestANonZeroExitIsNotAFailure(t *testing.T) {
	t.Parallel()
	output, err := runGate(context.Background(), GateCommand{
		Name:      "finds-something",
		Program:   "/bin/sh",
		Arguments: []string{"-c", "echo 'a.ts:1:1: error nexus(consistency-no-enum): found'; exit 1"},
		Directory: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("a non-zero exit from a gate with findings must not be an error: %v", err)
	}
	if !strings.Contains(output, "consistency-no-enum") {
		t.Fatalf("the gate's stdout must be returned even on a non-zero exit, got %q", output)
	}
}

// A shared control passes only when both gates reported it, and each way of failing says which.
//
// The three failures are genuinely different diagnoses and collapsing them would waste the control.
// Neither side means the plant never reached either gate, so the file was written somewhere they do
// not lint. One side means the plant landed but one gate's parse or normalization dropped it, which
// is the silent total-mismatch bug this control exists to catch.
func TestASharedControlNeedsBothGatesToReportIt(t *testing.T) {
	t.Parallel()
	control := Control{
		Name:           "shared-enum",
		RelativePath:   "control/Planted.ts",
		Rule:           "consistency-no-enum",
		ExpectedShared: true,
	}
	planted := []Control{control}
	seen := []Finding{{File: "control/Planted.ts", Line: 4, Rule: "consistency-no-enum"}}

	both := checkControls(planted, Report{}, seen, seen)
	if !both[0].Detected {
		t.Fatalf("a shared control both gates reported must be detected, got %q", both[0].Detail)
	}

	neither := checkControls(planted, Report{}, nil, nil)
	if neither[0].Detected {
		t.Fatal("a shared control neither gate reported must not be detected")
	}
	if !strings.Contains(neither[0].Detail, "neither gate") {
		t.Fatalf("the detail must say neither gate saw it, got %q", neither[0].Detail)
	}

	cohereOnly := checkControls(planted, Report{}, seen, nil)
	if cohereOnly[0].Detected {
		t.Fatal("a shared control only cohere reported must not be detected: one side dropped it")
	}
	if !strings.Contains(cohereOnly[0].Detail, "only cohere") {
		t.Fatalf("the detail must name the side that saw it, got %q", cohereOnly[0].Detail)
	}

	gateOnly := checkControls(planted, Report{}, nil, seen)
	if !strings.Contains(gateOnly[0].Detail, "only the gate") {
		t.Fatalf("the detail must name the side that saw it, got %q", gateOnly[0].Detail)
	}
}

// A shared control must never satisfy ControlsProven.
//
// It proves the pipeline carries a finding; it says nothing about whether a one-sided finding would
// survive. Letting it count would turn the weaker proof into the stronger claim silently, which is
// the same class of defect as every other guard in this package.
func TestASharedControlDoesNotProveDirection(t *testing.T) {
	t.Parallel()
	provenance := Provenance{
		CohereFilesLinted: 3407,
		GateFilesLinted:   3407,
		CohereRulesRun:    23,
		GateRulesRun:      181,
		ControlsRun: []ControlResult{
			{Name: "shared-enum", Rule: "consistency-no-enum", Detected: true},
		},
	}

	if provenance.ControlsProven() {
		t.Fatal("a detected shared control must not count as proving either direction")
	}
	if trustworthy, _ := provenance.Trustworthy(); trustworthy {
		t.Fatal("a run carrying only a shared control has not shown it can detect a difference, so it must not be trustworthy")
	}
}

// Cleanup must unwind the directories the plant created, and only those.
//
// Removing the control file and leaving its directory behind is still a mutation of a tree several
// people work in, and it shows up in their `git status` as an unexplained empty directory. An
// earlier run of this harness did exactly that and somebody else found the leftover.
func TestCleanupRemovesTheDirectoriesItCreated(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	control := Control{
		Name:         "nested",
		RelativePath: filepath.Join("code-quality", "differential-control", "Planted.ts"),
		Contents:     "export enum Planted { A = 'A' }\n",
		Rule:         "consistency-no-enum",
	}

	_, cleanup, err := plantControls(RunOptions{Root: root, Controls: []Control{control}})
	if err != nil {
		t.Fatal(err)
	}

	nested := filepath.Join(root, "code-quality", "differential-control")
	if _, err := os.Stat(nested); err != nil {
		t.Fatalf("the plant should have created %s: %v", nested, err)
	}

	cleanup()

	// Both levels go, because the plant created both.
	if _, err := os.Stat(nested); !os.IsNotExist(err) {
		t.Fatal("the directory the plant created was left behind")
	}
	if _, err := os.Stat(filepath.Join(root, "code-quality")); !os.IsNotExist(err) {
		t.Fatal("the outer directory the plant created was left behind")
	}
	// The tree root is never removed, however far the unwind walks.
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("the tree root must survive cleanup: %v", err)
	}
}

// A directory that already existed, or that something else wrote into, must survive.
//
// The first case would be the harness deleting part of the tree it was asked to measure. The second
// is the one that matters on a shared worktree: if a sibling's file landed in that directory while
// the gates ran, removing it would destroy their work.
func TestCleanupLeavesDirectoriesItDidNotCreate(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	preexisting := filepath.Join(root, "code-quality")
	if err := os.MkdirAll(preexisting, 0o755); err != nil {
		t.Fatal(err)
	}

	control := Control{
		Name:         "nested",
		RelativePath: filepath.Join("code-quality", "differential-control", "Planted.ts"),
		Contents:     "export enum Planted { A = 'A' }\n",
		Rule:         "consistency-no-enum",
	}
	_, cleanup, err := plantControls(RunOptions{Root: root, Controls: []Control{control}})
	if err != nil {
		t.Fatal(err)
	}

	// A sibling writes into the directory the plant created, while the gates would be running.
	sibling := filepath.Join(root, "code-quality", "differential-control", "SomebodyElse.ts")
	if err := os.WriteFile(sibling, []byte("export const theirs = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cleanup()

	if _, err := os.Stat(sibling); err != nil {
		t.Fatalf("a sibling's file inside the planted directory must survive cleanup: %v", err)
	}
	if _, err := os.Stat(preexisting); err != nil {
		t.Fatalf("a directory that existed before the plant must survive cleanup: %v", err)
	}
}

// A gate that broke must not read as a gate that found nothing.
//
// Both exit non-zero, which is the whole difficulty: a linter with findings and a linter that
// crashed are the same exit code, distinguished only by whether stdout carries anything. Treating
// the crash as findings would hand an empty finding list to the comparison, and an empty list on
// one side makes every finding on the other look one-sided while the report stays well-formed.
func TestAGateThatFailedWithNoOutputIsNotAnEmptyResult(t *testing.T) {
	t.Parallel()
	_, err := runGate(context.Background(), GateCommand{
		Name:      "crashes",
		Program:   "/bin/sh",
		Arguments: []string{"-c", "echo 'cannot find module oxlint' >&2; exit 1"},
		Directory: t.TempDir(),
	})
	if err == nil {
		t.Fatal("a gate that exited non-zero having printed no findings must be an error, not an empty result")
	}
	if !strings.Contains(err.Error(), "failed rather than found nothing") {
		t.Fatalf("the error must say the gate broke rather than came back clean, got %v", err)
	}

	// The gate's own explanation has to reach the reader. Diagnosing this from an exit code alone
	// sends them to the harness instead of to the gate, which is where the fault actually is.
	//
	// This assertion does not prove the explicit `execution.Stderr` capture is what carried it:
	// Go's `Output` also fills `ExitError.Stderr` when `cmd.Stderr` is nil, so both spellings
	// surface the text and a mutation removing the capture stays green here. Verified by probing
	// both forms rather than assumed. The assertion is still worth keeping, because what has to
	// hold is that the text reaches the reader, not which mechanism delivered it.
	if !strings.Contains(err.Error(), "cannot find module oxlint") {
		t.Fatalf("the gate's stderr must be surfaced, got %v", err)
	}

	// The control: the same non-zero exit WITH findings on stdout is normal and must still pass.
	output, err := runGate(context.Background(), GateCommand{
		Name:      "finds-something",
		Program:   "/bin/sh",
		Arguments: []string{"-c", "echo 'a.ts:1:1: error nexus(consistency-no-enum): found'; exit 1"},
		Directory: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("a gate with findings exits non-zero and must not be an error: %v", err)
	}
	if !strings.Contains(output, "consistency-no-enum") {
		t.Fatalf("the findings must survive, got %q", output)
	}
}

// A gate that could not be launched surfaces its stderr too.
func TestAGateThatCannotLaunchSurfacesWhatItSaid(t *testing.T) {
	t.Parallel()
	_, err := runGate(context.Background(), GateCommand{
		Name:      "missing",
		Program:   filepath.Join(t.TempDir(), "no-such-binary"),
		Directory: t.TempDir(),
	})
	if err == nil {
		t.Fatal("a gate that does not exist must be an error")
	}
	if !strings.Contains(err.Error(), "did not run") {
		t.Fatalf("the error must say the gate never ran, got %v", err)
	}
}

// stderrTail keeps the end of a long stderr and stays silent when there was none.
func TestStderrTailReportsOnlyWhatThereIs(t *testing.T) {
	t.Parallel()
	if tail := stderrTail("   \n  \n"); tail != "" {
		t.Fatalf("a quiet gate must add nothing to the error, got %q", tail)
	}

	var many []string
	for line := 1; line <= maximumStderrLines+4; line++ {
		many = append(many, fmt.Sprintf("line %d", line))
	}
	tail := stderrTail(strings.Join(many, "\n"))

	// The end is what says what happened; the beginning is usually a banner.
	if !strings.Contains(tail, fmt.Sprintf("line %d", maximumStderrLines+4)) {
		t.Fatalf("the last line must survive, got %q", tail)
	}
	if strings.Contains(tail, "line 1 ") {
		t.Fatalf("a long stderr must be trimmed from the front, got %q", tail)
	}
}

// A support file contributes no evidence and must not be counted as a control that fired.
//
// The tempting implementation is to record it as detected, since nothing went wrong with it. That
// would inflate the number of controls reported as firing and make a run look better proven than it
// is, which is this package's own failure mode: a control that always passes is not a control.
func TestASupportFileIsNotCountedAsAControl(t *testing.T) {
	t.Parallel()
	planted := []Control{
		{Name: "real", RelativePath: "a.ts", Rule: "consistency-no-enum", ExpectedSide: SideCohere},
		{Name: "support", RelativePath: "target/b.ts", ExpectsNothing: true},
	}

	results := checkControls(planted, Report{Differences: []Difference{{
		Finding: Finding{File: "a.ts", Line: 1, Rule: "consistency-no-enum"},
		OnlyOn:  SideCohere,
	}}}, nil, nil)

	if len(results) != 1 {
		t.Fatalf("a support file must produce no control result at all, got %d results: %+v", len(results), results)
	}
	if results[0].Name != "real" {
		t.Fatalf("the surviving result must be the real control, got %q", results[0].Name)
	}

	// And it must not be reachable as evidence: a run carrying only support files has proven
	// nothing and must say so.
	onlySupport := Provenance{
		CohereFilesLinted: 3407,
		GateFilesLinted:   3407,
		CohereRulesRun:    23,
		GateRulesRun:      159,
		ControlsRun:       checkControls([]Control{planted[1]}, Report{}, nil, nil),
	}
	if onlySupport.ControlsProven() {
		t.Fatal("a run whose only planted files were support files must not count as proven")
	}
	if trustworthy, _ := onlySupport.Trustworthy(); trustworthy {
		t.Fatal("a run whose only planted files were support files must not be trustworthy")
	}
}

// The captured version must be the part that distinguishes two binaries.
//
// A local build's first line is the constant "cohere dev" for every binary ever compiled from this
// tree, so keeping it produces a field that is always populated, always plausible, and never
// distinguishes anything. That is worse than an empty field, because it looks like provenance. The
// first implementation did exactly that, and two binaries eleven commits apart both rendered
// identically while producing different results.
func TestTheCapturedVersionDistinguishesTwoBuilds(t *testing.T) {
	t.Parallel()
	older := versionLineFrom("cohere dev\n  platform:       darwin/arm64\n  typescript-go:  unknown (built from cohere bd344fc70555)\n")
	newer := versionLineFrom("cohere dev\n  platform:       darwin/arm64\n  compiler:       unknown (built from cohere 107d82cf8568)\n")

	if older == newer {
		t.Fatalf("two builds from different commits rendered identically as %q, so the field proves nothing", older)
	}
	if !strings.Contains(older, "bd344fc70555") || !strings.Contains(newer, "107d82cf8568") {
		t.Fatalf("the commit must survive: got %q and %q", older, newer)
	}

	// A release build names no commit and states its version first, so that is the fallback.
	release := versionLineFrom("cohere 1.4.0\n  platform: darwin/arm64\n")
	if release != "cohere 1.4.0" {
		t.Fatalf("a release build must fall back to its version line, got %q", release)
	}

	if quiet := versionLineFrom("   \n\n"); !strings.Contains(quiet, "unknown") {
		t.Fatalf("a binary that printed nothing must say so, got %q", quiet)
	}
}

// A finding in a planted file is the harness looking at itself, whatever rule produced it.
//
// The first version of this matched on the control's declared rule, which left any other rule
// firing on the same planted file counted as observed. One did: a control file written for an
// import-alias rule also tripped the gate's import-ordering rule, and the summary read
// "1 observed, 2 planted" when all three were this harness's footprint.
//
// The number in that summary is what a reader trusts, and overstating it by the harness's own
// perturbation is how a tree with a real gap of zero gets reported as having one.
func TestAnyFindingInAPlantedFileIsPlanted(t *testing.T) {
	t.Parallel()
	planted := []Control{
		{Name: "declared", RelativePath: "control/Planted.ts", Rule: "consistency-no-enum", ExpectedSide: SideCohere},
		{Name: "support", RelativePath: "control/Target.ts", ExpectsNothing: true},
	}

	report := Report{Differences: []Difference{
		// The rule the control was written for.
		{Finding: Finding{File: "control/Planted.ts", Line: 3, Rule: "consistency-no-enum"}, OnlyOn: SideCohere},
		// A different rule, same planted file. Still the harness's own footprint.
		{Finding: Finding{File: "control/Planted.ts", Line: 1, Rule: "consistency-organize-imports"}, OnlyOn: SideGate},
		// A support file asserts nothing but a rule may still fire on it.
		{Finding: Finding{File: "control/Target.ts", Line: 1, Rule: "consistency-require-constant-casing"}, OnlyOn: SideGate},
		// A real file the harness never touched.
		{Finding: Finding{File: "app/Real.tsx", Line: 9, Rule: "react-component-no-multiple-primary"}, OnlyOn: SideGate},
	}}

	markPlantedDifferences(&report, planted)

	observed := report.ObservedDifferences()
	if len(observed) != 1 {
		t.Fatalf("only the untouched file's finding is observed, got %d: %+v", len(observed), observed)
	}
	if observed[0].Finding.File != "app/Real.tsx" {
		t.Fatalf("the observed difference must be the one in a file the harness did not write, got %q", observed[0].Finding.File)
	}

	// The control: a report with no planted files must leave everything observed, so this cannot
	// pass against a marker that flags indiscriminately.
	untouched := Report{Differences: []Difference{
		{Finding: Finding{File: "app/Real.tsx", Line: 9, Rule: "react-component-no-multiple-primary"}},
	}}
	markPlantedDifferences(&untouched, nil)
	if len(untouched.ObservedDifferences()) != 1 {
		t.Fatal("with nothing planted, every difference must remain observed")
	}
}

// A planted control must not block agreement.
//
// Proving detection requires creating a difference, so counting a control against the verdict would
// mean the harness could never agree with itself: the very run that demonstrates the instrument
// works would fail on the evidence that it does.
func TestAPlantedBothActiveDifferenceDoesNotBlockAgreement(t *testing.T) {
	t.Parallel()
	report := Compare(Inputs{
		CohereFindings:   []Finding{{File: "control/Planted.ts", Line: 3, Rule: "consistency-no-enum", Side: SideCohere}},
		CoherePopulation: Population{Findings: 1, FilesWalked: 3407, Rules: 23},
		GatePopulation:   Population{FilesWalked: 3407},
		CohereRules:      map[string]bool{"consistency-no-enum": true},
		ConfiguredRules:  map[string]bool{"consistency-no-enum": true},
	})
	report.Provenance = provenProvenance()

	if report.Agreed() {
		t.Fatal("before marking, a both-active difference must block agreement")
	}

	markPlantedDifferences(&report, []Control{
		{Name: "planted", RelativePath: "control/Planted.ts", Rule: "consistency-no-enum", ExpectedSide: SideCohere},
	})

	if !report.Agreed() {
		t.Fatal("once marked as planted, the harness's own control must not count against agreement")
	}
}

// Both halves of controlMatches are required, and the rule half alone is not enough.
//
// This exists because a mutation dropping the path check was read as equivalent: it produced the
// same totals on this tree, so it looked like a no-op. It was not. A rule-only check answers
// correctly here only because every rule a control declares happens to appear in no other file,
// which is a property of today's controls rather than of the matcher. The moment a planted rule
// also fires somewhere real, a finding gets attributed to a control that did not plant it.
//
// Pinned as a property rather than left to the totals, because the totals agreed while the reason
// did not, and a test that only checks the total would agree with them.
func TestControlMatchesNeedsBothTheRuleAndThePath(t *testing.T) {
	t.Parallel()
	control := Control{
		Name:         "gate-only",
		RelativePath: "control/PlantedOrder.ts",
		Rule:         "consistency-organize-imports",
		ExpectedSide: SideGate,
	}

	if !controlMatches(control, Finding{File: "control/PlantedOrder.ts", Line: 3, Rule: "consistency-organize-imports"}) {
		t.Fatal("the finding this control planted must match")
	}

	// Same rule, different file. This is the live case: the gate's ordering rule also fires on the
	// other control's file, and attributing it here would credit a control that did not plant it.
	if controlMatches(control, Finding{File: "control/PlantedImport.ts", Line: 3, Rule: "consistency-organize-imports"}) {
		t.Fatal("a finding of the same rule in a different file must not match: the path half is what rejects it")
	}

	// Same file, different rule. The control did not plant this finding either.
	if controlMatches(control, Finding{File: "control/PlantedOrder.ts", Line: 1, Rule: "consistency-no-enum"}) {
		t.Fatal("a different rule in the control's own file must not match: the rule half is what rejects it")
	}
}

// markPlantedDifferences asks a deliberately different question from controlMatches.
//
// One verifies a named control fired, and a control that fired somewhere else has not fired. The
// other asks whether the harness caused a finding at all, where the rule is irrelevant because the
// file would not exist otherwise. Conflating them in either direction is a defect, and both
// directions have now been written by mistake, so both are pinned.
func TestTheTwoMatchersAnswerDifferentQuestions(t *testing.T) {
	t.Parallel()
	control := Control{
		Name:         "gate-only",
		RelativePath: "control/PlantedOrder.ts",
		Rule:         "consistency-organize-imports",
		ExpectedSide: SideGate,
	}
	// A rule the control never declared, in the file the control planted.
	stray := Finding{File: "control/PlantedOrder.ts", Line: 1, Rule: "consistency-require-constant-casing"}

	if controlMatches(control, stray) {
		t.Fatal("controlMatches must reject a rule the control did not declare, even in its own file")
	}

	report := Report{Differences: []Difference{{Finding: stray}}}
	markPlantedDifferences(&report, []Control{control})
	if len(report.ObservedDifferences()) != 0 {
		t.Fatal("markPlantedDifferences must claim any finding in a planted file, whatever rule produced it")
	}
}

// A local build's own disclaimer must travel with its commit.
//
// `cohere -version` states, in a `note:` line, that a local build's rules are whatever was on disk
// when it compiled. Carrying only the commit is worse than carrying nothing: the sha is real, it
// resolves, and it reads as provenance while saying nothing about uncommitted rules linked in
// beside it.
//
// Three people quoted a rule count from a dev build in one night, each correct about the binary and
// wrong about what they said it measured, with this note one command away from every invocation
// they made. A disclosure that has to be sought is a disclosure that gets skipped.
func TestALocalBuildCarriesItsOwnDisclaimer(t *testing.T) {
	t.Parallel()
	local := versionLineFrom(
		"cohere dev\n" +
			"  platform:       darwin/arm64\n" +
			"  compiler:       unknown (built from cohere ad96b871c377)\n" +
			"  note:           a local build, so the rules are whatever was on disk when it was compiled\n",
	)

	if !strings.Contains(local, "ad96b871c377") {
		t.Fatalf("the commit must survive, got %q", local)
	}
	if !strings.Contains(local, "whatever was on disk") {
		t.Fatalf("a local build's disclaimer must travel with its commit, got %q", local)
	}

	// A build with a commit and no note carries the commit alone rather than inventing a
	// qualification, so the disclaimer means something when it does appear.
	quiet := versionLineFrom("cohere 1.4.0\n  compiler: unknown (built from cohere 107d82cf8568)\n")
	if !strings.Contains(quiet, "107d82cf8568") {
		t.Fatalf("the commit must survive without a note, got %q", quiet)
	}
	if strings.Contains(quiet, "whatever was on disk") {
		t.Fatalf("a build that made no such claim must not be given one, got %q", quiet)
	}

	// A release build naming no commit still falls back to its version line.
	release := versionLineFrom("cohere 1.4.0\n  platform: darwin/arm64\n")
	if release != "cohere 1.4.0" {
		t.Fatalf("a release build must fall back to its version line, got %q", release)
	}
}

// The `-rules` contract is bare names on stdout, and the guard must reject anything else.
//
// This is pinned because the contract was tested by accident rather than by design. `cohere -rules`
// gained a provenance note, and the parser kept working only because the note goes to stderr while
// the names go to stdout. Had it landed on stdout the guard would have rejected the whole list and
// the harness would have fallen back to the slower path, or failed, on correct output.
//
// The guard stays strict rather than learning to skip prose, because a line the parser does not
// understand is a contract change, and silently tolerating it is how a format drift becomes a short
// rule list. A short list is the dangerous failure: every missing name becomes a rule the harness
// believes cohere lacks, which turns real disagreements into excused coverage gaps.
func TestTheRulesContractIsBareNamesOnly(t *testing.T) {
	t.Parallel()
	bare := "consistency-no-enum\nreact-component-no-multiple-primary\n"
	names, err := namesFromRuleLines(bare)
	if err != nil {
		t.Fatalf("bare names must parse: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}

	// The exact note `cohere -rules` emits, which lives on stderr today. If it ever reaches stdout
	// this must fail loudly rather than be absorbed.
	withNote := bare + "note: 39 rules from a local build, so this is whatever was on disk when it was compiled\n"
	if _, err := namesFromRuleLines(withNote); err == nil {
		t.Fatal("a line carrying prose must be rejected: the contract is bare names, and absorbing prose hides a format change")
	}

	if _, err := namesFromRuleLines("   \n\n"); err == nil {
		t.Fatal("an empty list must be rejected: a harness believing cohere implements nothing excuses every disagreement")
	}
}
