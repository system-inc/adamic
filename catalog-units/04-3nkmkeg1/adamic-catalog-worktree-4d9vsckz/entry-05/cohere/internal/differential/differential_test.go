// Tests for the instrument that decides whether cohere agrees with the gate.
//
// The thing being defended here is one level up from the usual. This package exists to refuse
// untrustworthy results, so a bug in it does not produce a wrong answer that someone notices — it
// produces a clean, confident answer that nobody questions, about a comparison that never happened.
// Every test below is therefore written as "prove this can fail", not "prove this passes on good
// input". A detector that has never returned a positive has not been shown to work.
package differential

import (
	"strings"
	"testing"
)

// The decoration test, first because it guards a vacuous-probe generator sitting inside the
// vacuous-probe detector.
//
// The two gates decorate rule names in opposite directions. If normalization breaks, every finding
// is keyed under a different rule name on each side, so every finding becomes one-sided and the
// report is a total mismatch — which is indistinguishable, in shape, from a real 130-finding
// disagreement. A well-formed report of a comparison that did not happen.
func TestRuleNamesFromBothGatesCollapseToTheSameName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw    string
		expect string
	}{
		// The gate puts the plugin outside.
		{"nexus(consistency-no-shouting)", "consistency-no-shouting"},
		{"structure(react-component-no-multiple-primary)", "react-component-no-multiple-primary"},
		{"typescript(no-explicit-any)", "no-explicit-any"},
		// cohere puts the message id after.
		{"consistency-no-shouting/noShouting", "consistency-no-shouting"},
		{"consistency-no-abbreviated-identifier/noParams", "consistency-no-abbreviated-identifier"},
		// Already bare, from either side.
		{"consistency-no-enum", "consistency-no-enum"},
		{"  consistency-no-enum  ", "consistency-no-enum"},
	}

	for _, testCase := range cases {
		if got := NormalizeRuleName(testCase.raw); got != testCase.expect {
			t.Errorf("NormalizeRuleName(%q) = %q, want %q", testCase.raw, got, testCase.expect)
		}
	}

	// The property that actually matters, stated as a property rather than as a pair of cases: the
	// same rule reaching the harness through the two different formats must key identically. This
	// is the assertion that fails if someone "simplifies" normalization in a way that happens to
	// satisfy every case above individually.
	fromGate := NormalizeRuleName("nexus(consistency-no-enum)")
	fromCohere := NormalizeRuleName("consistency-no-enum/noEnum")
	if fromGate != fromCohere {
		t.Fatalf("the same rule normalized to %q from the gate and %q from cohere, so no finding could ever match", fromGate, fromCohere)
	}
}

// Prove the comparison can report a difference, in each direction separately.
//
// Both directions are tested because they fail independently: a harness that only ever sees the
// gate's findings reports a clean diff for every defect cohere finds alone, and it looks exactly
// like agreement.
func TestCompareDetectsADifferenceInEachDirection(t *testing.T) {
	t.Parallel()
	rules := Inputs{
		CohereRules:     map[string]bool{"consistency-no-enum": true},
		ConfiguredRules: map[string]bool{"consistency-no-enum": true},
	}

	onlyCohere := rules
	onlyCohere.CohereFindings = []Finding{{File: "a.ts", Line: 3, Rule: "consistency-no-enum", Side: SideCohere}}
	onlyCohere.CoherePopulation = Population{Findings: 1, FilesWalked: 3407, Rules: 22}
	onlyCohere.GatePopulation = Population{Findings: 0, FilesWalked: 3407}

	report := Compare(onlyCohere)
	if len(report.Differences) != 1 {
		t.Fatalf("a finding only cohere reported must surface as one difference, got %d", len(report.Differences))
	}
	if report.Differences[0].OnlyOn != SideCohere {
		t.Fatalf("difference attributed to %q, want %q", report.Differences[0].OnlyOn, SideCohere)
	}

	onlyGate := rules
	onlyGate.GateFindings = []Finding{{File: "a.ts", Line: 3, Rule: "consistency-no-enum", Side: SideGate}}
	onlyGate.CoherePopulation = Population{Findings: 0, FilesWalked: 3407, Rules: 22}
	onlyGate.GatePopulation = Population{Findings: 1, FilesWalked: 3407}

	report = Compare(onlyGate)
	if len(report.Differences) != 1 {
		t.Fatalf("a finding only the gate reported must surface as one difference, got %d", len(report.Differences))
	}
	if report.Differences[0].OnlyOn != SideGate {
		t.Fatalf("difference attributed to %q, want %q", report.Differences[0].OnlyOn, SideGate)
	}
}

// The same finding from both sides is agreement, not two differences.
//
// This is the control for the test above: without it, a Compare that reported everything as a
// difference would pass both direction tests and be completely broken.
func TestMatchingFindingsAreNotDifferences(t *testing.T) {
	t.Parallel()
	report := Compare(Inputs{
		CohereFindings:   []Finding{{File: "a.ts", Line: 3, Column: 16, Rule: "consistency-no-enum", Side: SideCohere}},
		GateFindings:     []Finding{{File: "a.ts", Line: 3, Column: 9, Rule: "consistency-no-enum", Side: SideGate}},
		CoherePopulation: Population{Findings: 1, FilesWalked: 3407, Rules: 22},
		GatePopulation:   Population{Findings: 1, FilesWalked: 3407},
		CohereRules:      map[string]bool{"consistency-no-enum": true},
		ConfiguredRules:  map[string]bool{"consistency-no-enum": true},
	})

	// Note the differing columns. They must still match: the two gates locate the same finding at
	// different offsets within a line routinely, and keying on column would report a difference for
	// every finding the two sides agree about.
	if len(report.Differences) != 0 {
		t.Fatalf("the same finding from both sides must not be a difference, got %d: %+v", len(report.Differences), report.Differences)
	}
	if len(report.Agreements) != 1 || report.Agreements[0].Shared != 1 {
		t.Fatalf("expected one rule agreeing with one shared finding, got %+v", report.Agreements)
	}
}

// The three classifications are three different facts and must not collapse into each other.
//
// Getting this wrong is not cosmetic. Most of the current gap is rules cohere has never claimed,
// and reporting those as disagreements reports noise as signal — while reporting a genuine
// disagreement as a coverage gap hides the only class that blocks the claim.
func TestClassificationSeparatesCoverageFromCorrectness(t *testing.T) {
	t.Parallel()
	inputs := Inputs{
		GateFindings: []Finding{
			{File: "a.tsx", Line: 1, Rule: "react-component-no-multiple-primary", Side: SideGate},
			{File: "b.ts", Line: 2, Rule: "import-require-path-alias", Side: SideGate},
			{File: "c.ts", Line: 3, Rule: "consistency-no-enum", Side: SideGate},
		},
		CoherePopulation: Population{Findings: 0, FilesWalked: 3407, Rules: 22},
		GatePopulation:   Population{Findings: 3, FilesWalked: 3407},
		// cohere implements two of the three rules, and the config enables only one of those two.
		CohereRules:     map[string]bool{"import-require-path-alias": true, "consistency-no-enum": true},
		ConfiguredRules: map[string]bool{"consistency-no-enum": true, "react-component-no-multiple-primary": true},
	}

	byRule := map[string]Classification{}
	for _, difference := range Compare(inputs).Differences {
		byRule[difference.Finding.Rule] = difference.Classification
	}

	expected := map[string]Classification{
		// cohere has no such rule, so it cannot disagree — a coverage fact.
		"react-component-no-multiple-primary": ClassificationNotPorted,
		// cohere has the rule, nobody enabled it, so it ran over no files — neither a port gap nor
		// a defect.
		"import-require-path-alias": ClassificationNotConfigured,
		// Both sides had it and both had it on. Somebody is wrong, and this is the only class that
		// blocks the agreement claim.
		"consistency-no-enum": ClassificationBothActive,
	}
	for ruleName, want := range expected {
		if got := byRule[ruleName]; got != want {
			t.Errorf("rule %q classified %q, want %q", ruleName, got, want)
		}
	}
}

// Only both-active differences block the claim.
func TestUnportedRulesDoNotBlockAgreement(t *testing.T) {
	t.Parallel()
	provenance := provenProvenance()

	notPorted := Compare(Inputs{
		GateFindings:     []Finding{{File: "a.tsx", Line: 1, Rule: "react-component-no-multiple-primary", Side: SideGate}},
		CoherePopulation: Population{FilesWalked: 3407, Rules: 22},
		GatePopulation:   Population{Findings: 1, FilesWalked: 3407},
		CohereRules:      map[string]bool{},
		ConfiguredRules:  map[string]bool{},
	})
	notPorted.Provenance = provenance
	if !notPorted.Agreed() {
		t.Fatal("a rule cohere has never claimed must not count against agreement")
	}

	bothActive := Compare(Inputs{
		CohereFindings:   []Finding{{File: "a.ts", Line: 1, Rule: "consistency-no-enum", Side: SideCohere}},
		CoherePopulation: Population{Findings: 1, FilesWalked: 3407, Rules: 22},
		GatePopulation:   Population{FilesWalked: 3407},
		CohereRules:      map[string]bool{"consistency-no-enum": true},
		ConfiguredRules:  map[string]bool{"consistency-no-enum": true},
	})
	bothActive.Provenance = provenance
	if bothActive.Agreed() {
		t.Fatal("a rule both sides had enabled, differing, must block agreement")
	}
}

// The vacuity guard: a side that walked no files is not a clean side.
//
// This is the exact failure the gate cohere replaces shipped for days, and a diff harness inherits
// it in a worse form, because an empty diff over two empty runs reads as proof of agreement.
func TestARunOverNoFilesIsNotComparable(t *testing.T) {
	t.Parallel()
	report := Compare(Inputs{
		CoherePopulation: Population{FilesWalked: 0, Rules: 22},
		GatePopulation:   Population{FilesWalked: 3407},
	})
	if report.Comparable {
		t.Fatal("cohere walking zero files must make the comparison not comparable")
	}
	if report.Agreed() {
		t.Fatal("a not-comparable report must never claim agreement")
	}

	rendered := &strings.Builder{}
	Write(rendered, report)
	if strings.Contains(rendered.String(), "✓") {
		t.Fatalf("a vacuous run must not render a pass mark: %q", rendered.String())
	}
}

// An empty diff from an unproven harness must not read as agreement.
//
// This is the property the whole package is named for, and it is the one that four vacuous probes
// in one night got past: each had a plausible population and an empty result.
func TestAnEmptyDiffWithoutControlsIsNotAgreement(t *testing.T) {
	t.Parallel()
	unproven := Compare(Inputs{
		CoherePopulation: Population{FilesWalked: 3407, Rules: 22},
		GatePopulation:   Population{FilesWalked: 3407},
	})
	unproven.Provenance = Provenance{
		CohereFilesLinted: 3407,
		GateFilesLinted:   3407,
		CohereRulesRun:    22,
		GateRulesRun:      181,
		// No controls run at all.
	}
	if len(unproven.Differences) != 0 {
		t.Fatalf("this fixture is meant to have no differences, got %d", len(unproven.Differences))
	}
	if unproven.Agreed() {
		t.Fatal("an empty diff from a harness never shown able to detect a difference must not claim agreement")
	}

	rendered := &strings.Builder{}
	Write(rendered, unproven)
	if strings.Contains(rendered.String(), "✓") {
		t.Fatalf("an unproven run must not render a pass mark: %q", rendered.String())
	}

	// And the same comparison, once the controls have fired, must reach the opposite verdict —
	// otherwise this test would pass against a harness that can never say ✓ at all.
	proven := unproven
	proven.Provenance = provenProvenance()
	if !proven.Agreed() {
		t.Fatal("an empty diff from a proven harness over a real population is agreement, and must be reachable")
	}
}

// A control that fired in only one direction is not proof.
func TestOneDirectionOfControlIsNotProven(t *testing.T) {
	t.Parallel()
	oneDirection := Provenance{
		CohereFilesLinted: 3407,
		GateFilesLinted:   3407,
		CohereRulesRun:    22,
		GateRulesRun:      181,
		ControlsRun: []ControlResult{
			{Name: "gate-only", ExpectedSide: SideGate, Rule: "react-component-no-multiple-primary", Detected: true},
		},
	}
	if oneDirection.ControlsProven() {
		t.Fatal("a control in one direction only must not count as proven: a defect on the unexercised side would still report clean")
	}

	missed := provenProvenance()
	missed.ControlsRun[0].Detected = false
	missed.ControlsRun[0].Detail = "planted enum produced no difference"
	if missed.ControlsProven() {
		t.Fatal("a control that did not fire must never count as proven")
	}
	if trustworthy, reasons := missed.Trustworthy(); trustworthy || len(reasons) == 0 {
		t.Fatal("a missed control must make the run untrustworthy and say why")
	}
}

// A population too small to be the tree is its own tell.
func TestAnImplausiblySmallRunIsUntrustworthy(t *testing.T) {
	t.Parallel()
	tiny := provenProvenance()
	tiny.CohereFilesLinted = 2
	trustworthy, reasons := tiny.Trustworthy()
	if trustworthy {
		t.Fatal("a run over 2 files must not be trusted as a run over the tree")
	}
	if !strings.Contains(strings.Join(reasons, " "), "too few") {
		t.Fatalf("the reason must name the implausible file count, got %v", reasons)
	}
}

// provenProvenance is a run that walked the real tree and fired a control in both directions.
func provenProvenance() Provenance {
	return Provenance{
		CohereFilesLinted: 3407,
		GateFilesLinted:   3407,
		CohereRulesRun:    22,
		GateRulesRun:      181,
		CohereCommand:     "cohere -lint",
		GateCommand:       "node RunCachedOxlint.ts --no-cache",
		ControlsRun: []ControlResult{
			{Name: "cohere-only", ExpectedSide: SideCohere, Rule: "consistency-no-enum", Detected: true},
			{Name: "gate-only", ExpectedSide: SideGate, Rule: "react-component-no-multiple-primary", Detected: true},
		},
	}
}

// Absent knowledge must not look like knowledge of absence.
//
// A nil ConfiguredRules means the caller never read the lint config, not that the config enables
// nothing. Go returns false for any read of a nil map, so without an explicit check every rule
// would classify as not-configured — silently converting every genuine disagreement into an excused
// non-finding, and reporting a clean tree with the highest possible confidence.
//
// The degradation is deliberate: fall back to the two-way split, which is honest about what the
// caller actually knows, rather than to the three-way split, which claims knowledge it does not have.
func TestUnknownConfigurationDoesNotExcuseEveryRule(t *testing.T) {
	t.Parallel()
	report := Compare(Inputs{
		CohereFindings:   []Finding{{File: "a.ts", Line: 1, Rule: "consistency-no-enum", Side: SideCohere}},
		CoherePopulation: Population{Findings: 1, FilesWalked: 3407, Rules: 22},
		GatePopulation:   Population{FilesWalked: 3407},
		CohereRules:      map[string]bool{"consistency-no-enum": true},
		// ConfiguredRules deliberately nil: the caller did not read the configuration.
	})

	if len(report.Differences) != 1 {
		t.Fatalf("expected one difference, got %d", len(report.Differences))
	}
	if got := report.Differences[0].Classification; got != ClassificationBothActive {
		t.Fatalf("with configuration unknown, a rule cohere implements and reported must stay %q, got %q — nil must not read as disabled",
			ClassificationBothActive, got)
	}

	report.Provenance = provenProvenance()
	if report.Agreed() {
		t.Fatal("a difference on an implemented rule must still block agreement when the config was never read")
	}
}

// A slash means two different things depending on which side of it the rule name sits.
//
// cohere prints `rule-name/messageId`, so the name is before the slash. The lint config keys rules
// as `plugin/rule-name`, so the name is after it. Taking the first segment unconditionally is
// correct for the first and silently wrong for the second, and the wrongness does not look like a
// parse failure: every config key collapses to its plugin, so `nexus/consistency-no-enum` and
// `nexus/consistency-no-shouting` both become `nexus`.
//
// The consequence is not cosmetic. A configured-rule set built that way contains two junk entries
// and no real rule names, so every rule reads as unconfigured, every genuine disagreement is filed
// as not-configured, and the harness excuses exactly the findings it exists to surface. It reached
// a real run and rendered as one wrong word in a table.
func TestAPluginPrefixedConfigKeyKeepsTheRuleName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw    string
		expect string
	}{
		// Config keys: plugin first, rule after.
		{"nexus/consistency-no-enum", "consistency-no-enum"},
		{"structure/react-component-no-multiple-primary", "react-component-no-multiple-primary"},
		{"typescript/no-explicit-any", "no-explicit-any"},
		// cohere findings: rule first, message id after.
		{"consistency-no-enum/noEnum", "consistency-no-enum"},
		{"consistency-no-abbreviated-identifier/noParams", "consistency-no-abbreviated-identifier"},
		// Gate findings: plugin in parentheses.
		{"nexus(consistency-no-enum)", "consistency-no-enum"},
	}
	for _, testCase := range cases {
		if got := NormalizeRuleName(testCase.raw); got != testCase.expect {
			t.Errorf("NormalizeRuleName(%q) = %q, want %q", testCase.raw, got, testCase.expect)
		}
	}

	// The property, stated as a property: all three spellings of one rule must reach the same name,
	// because a config key and a finding have to meet in the same map for classification to work.
	fromConfig := NormalizeRuleName("nexus/consistency-no-enum")
	fromCohere := NormalizeRuleName("consistency-no-enum/noEnum")
	fromGate := NormalizeRuleName("nexus(consistency-no-enum)")
	if fromConfig != fromCohere || fromCohere != fromGate {
		t.Fatalf("one rule normalized three ways: config %q, cohere %q, gate %q", fromConfig, fromCohere, fromGate)
	}
}

// The verdict must state its own scope when it covers less than the configuration.
//
// A comparison can only speak about rules cohere implements. The rest of the config is enabled in
// the gate, unported, and never compared, so a bare "agrees" reads as broader than the test. On
// this tree that gap is most of the catalog, and the reason those rules find nothing is that the
// gate has been enforcing them for months rather than that they are worthless. A clean diff
// therefore cannot be the signal to remove the old tool.
func TestTheVerdictNamesTheRulesItDidNotCompare(t *testing.T) {
	t.Parallel()
	partial := Compare(Inputs{
		CoherePopulation: Population{FilesWalked: 3407, Rules: 33},
		GatePopulation:   Population{FilesWalked: 3407},
		CohereRules:      map[string]bool{"consistency-no-enum": true},
		ConfiguredRules:  map[string]bool{"consistency-no-enum": true, "unported-one": true, "unported-two": true},
	})
	partial.Provenance = provenProvenance()

	rendered := &strings.Builder{}
	Write(rendered, partial)
	text := rendered.String()

	if !strings.Contains(text, "✓") {
		t.Fatalf("a clean partial comparison still agrees about what it compared: %q", text)
	}
	if !strings.Contains(text, "not a verdict about them") {
		t.Fatalf("the verdict must disclaim the rules it could not compare: %q", text)
	}
	if !strings.Contains(text, "the other 2 are unported") {
		t.Fatalf("the verdict must count the uncompared rules, got: %q", text)
	}

	// The control: full coverage must not carry the disclaimer, or it becomes noise a reader learns
	// to skip and stops meaning anything on the run where it matters.
	full := Compare(Inputs{
		CoherePopulation: Population{FilesWalked: 3407, Rules: 1},
		GatePopulation:   Population{FilesWalked: 3407},
		CohereRules:      map[string]bool{"consistency-no-enum": true},
		ConfiguredRules:  map[string]bool{"consistency-no-enum": true},
	})
	full.Provenance = provenProvenance()

	rendered = &strings.Builder{}
	Write(rendered, full)
	if strings.Contains(rendered.String(), "not a verdict about them") {
		t.Fatalf("a comparison covering the whole config must not disclaim anything: %q", rendered.String())
	}
}
