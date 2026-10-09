// Package differential compares what cohere found against what the gate it replaces found.
//
// The claim this package exists to test is "cohere agrees with the gate." That claim is not one
// claim. With 181 rules configured it is 181 separate claims, and an aggregate agreement can hold
// while any number of individual rules disagree in ways that cancel: cohere missing three findings
// on one rule and inventing three on another sums to zero and looks like agreement.
//
// So a difference is keyed by file, line, and rule, and the report is per rule rather than a count.
//
// The harder problem is the empty diff. Two gates that both report nothing produce identical output
// whether one of them checked the tree or checked no files at all, and that is precisely the failure
// the gate cohere replaces shipped for days. A diff harness inherits that failure mode and makes it
// worse, because an empty diff over two vacuous runs reads as proof of agreement.
//
// The defense is Population, carried on every Report. A comparison states how many findings each
// side produced and how many files each side says it walked, so a reader can tell "they agree" from
// "neither of them looked." Compare refuses to call a run comparable when a side reports no coverage
// at all, and SelfTest proves the detector can detect by planting a violation each side sees alone.
package differential

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Side names which gate produced a finding.
type Side string

const (
	// SideCohere is ours: the Go binary.
	SideCohere Side = "cohere"
	// SideGate is theirs: oxlint behind RunCachedOxlint.
	SideGate Side = "gate"
)

// Finding is one report from one gate, reduced to the parts both gates can express.
//
// Column is deliberately absent from the key. The two gates locate a finding at different offsets
// within the same line often enough that keying on column would report a difference for every
// finding they agree about, drowning the real disagreements in noise. Line is the granularity a
// person acts on.
type Finding struct {
	File string
	Line int
	// Column is carried for the reader and never compared, for the reason stated above. It is the
	// difference between "go look at this" and "go find this", and a harness whose output cannot be
	// navigated to is a harness people stop running.
	Column  int
	Rule    string
	Message string
	Side    Side
}

// Key is the identity two findings must share to be called the same finding.
func (finding Finding) Key() string {
	return fmt.Sprintf("%s:%d:%s", finding.File, finding.Line, finding.Rule)
}

// NormalizeRuleName strips the decoration each gate puts around a rule name.
//
// The two formats decorate in opposite directions. The gate prints `structure(rule-name)` or
// `nexus(rule-name)`, putting the plugin outside. cohere prints `rule-name/messageId`, putting the
// message identifier after. Neither decoration is part of the rule's identity, and comparing
// decorated names would report every single rule as a disagreement — a total mismatch that still
// looks like well-formed output, which is the failure this whole package is built to refuse.
func NormalizeRuleName(raw string) string {
	name := strings.TrimSpace(raw)

	// `plugin(rule-name)` — take what is inside the parentheses.
	if open := strings.IndexByte(name, '('); open >= 0 && strings.HasSuffix(name, ")") {
		name = name[open+1 : len(name)-1]
	}

	// A slash means two opposite things depending on who wrote it, and an earlier version of this
	// function assumed there was only one meaning:
	//
	//	rule-name/messageId          cohere's findings   the name is BEFORE the slash
	//	plugin/rule-name             the config's keys   the name is AFTER the slash
	//
	// Taking the first segment unconditionally is right for the first and silently wrong for the
	// second, and the wrongness does not look like a parse failure. Every config key collapses to
	// its plugin, so the configured-rule set becomes `{nexus, structure, typescript}` and no real
	// rule name is in it. Every rule then reads as unconfigured and every genuine disagreement is
	// filed as not-configured, which is the harness excusing precisely the findings it exists to
	// surface. It reached a live run and showed up as one wrong word in a table.
	//
	// The two are told apart by shape rather than by a plugin list, so a new plugin needs no edit
	// here: rule names in both catalogs are hyphenated and message ids are camelCase.
	//
	// **A hyphen in the first segment is not enough**, and the one plugin that proves it is
	// `better-tailwindcss`. Its name is hyphenated, so `better-tailwindcss/no-duplicate-classes`
	// collapsed to `better-tailwindcss` and all eight of its rules became one entry. That is the
	// exact failure this comment describes one paragraph above, surviving the fix that was written
	// for it: every tailwind rule read as unconfigured, so a genuine disagreement in that family
	// would have been excused as not-configured rather than surfaced.
	//
	// The distinguishing shape is the second segment rather than the first. A message id is
	// camelCase and therefore never hyphenated, so a hyphen *after* the slash means the config's
	// `plugin/rule-name` form and the name is after; anything else is cohere's `rule-name/messageId`
	// form and the name is before. That reading is correct for both plugins whose names contain a
	// hyphen and plugins whose names do not.
	if slash := strings.IndexByte(name, '/'); slash >= 0 {
		before, after := name[:slash], name[slash+1:]
		if strings.Contains(after, "-") {
			name = after
		} else {
			name = before
		}
	}

	return strings.TrimSpace(name)
}

// Population is what a run covered, as distinct from what it found.
//
// Without this a Report cannot distinguish agreement from mutual silence. FilesWalked is what the
// side says about itself and is therefore a claim rather than proof, but a side claiming zero files
// is enough on its own to disqualify the comparison, which is the case that matters.
type Population struct {
	Findings    int
	FilesWalked int
	Rules       int
}

// Difference is one finding exactly one side reported.
type Difference struct {
	Finding Finding
	// OnlyOn is the side that reported it. The other side saw this file and this rule and said
	// nothing, which is the fact a reader has to classify.
	OnlyOn Side
	// Classification is why the two sides disagree, when it can be determined mechanically.
	Classification Classification
	// Planted marks a difference the harness caused by planting a control, as distinct from one it
	// observed in the codebase.
	//
	// A planted finding and an observed finding are the same shape in a difference table, and the
	// summary count is what a reader trusts. So a run with three controls reported "differences: 3
	// total" on a tree whose real gap was zero, and that number was read as the remaining exposure
	// and nearly published. Three problems and zero problems plus three proofs the instrument works
	// are not the same sentence.
	//
	// The controls still appear, because hiding them would make the evidence of detection invisible
	// in the place a reader is looking. They are separated instead of removed.
	Planted bool
}

// printAcknowledgedCaveat states that a green verdict rested on an acknowledgement.
//
// Factored out because there is more than one way to pass, and a caveat attached to only some of
// them is worse than none: it reads as absent on exactly the paths nobody checked. That is the same
// failure the scoped verdict above exists to prevent, arriving through the fix for it.
func printAcknowledgedCaveat(out io.Writer, byClassification map[Classification]int) {
	if acknowledgedCount := byClassification[ClassificationAcknowledged]; acknowledgedCount > 0 {
		fmt.Fprintf(out,
			"  %d differences were acknowledged rather than absent, each with a recorded reason\n",
			acknowledgedCount,
		)
	}
}

// Classification is the mechanical part of "why do these disagree."
//
// It is deliberately coarse. Most differences are configuration rather than correctness, and the
// classes below are the ones a program can decide from the two runs alone. Anything requiring a
// judgment about what a rule ought to do is left Unclassified rather than guessed, because a wrong
// classification is worse than none: it tells the reader the question has been answered.
type Classification string

const (
	// ClassificationNotPorted means cohere has no rule by this name, so the gate finding it alone
	// is expected and says nothing about agreement. This is the largest class during the migration
	// and the one that must not be mistaken for a defect.
	ClassificationNotPorted Classification = "not-ported"
	// ClassificationNotConfigured means the rule exists in cohere but the config never enables it,
	// so cohere ran it over no files.
	ClassificationNotConfigured Classification = "not-configured"
	// ClassificationBothActive means both sides have this rule and both had it enabled, and they
	// still disagree on this file and line. This is the class that is always a defect in one of the
	// two gates, and the only class that blocks the claim.
	ClassificationBothActive Classification = "both-active"
	// ClassificationAcknowledged means somebody looked at this exact finding and decided the
	// difference is correct, with a reason recorded in KnownGateDefects.
	//
	// It is a class rather than a deletion so the difference still appears in the report. A
	// difference that stops being printed stops being reviewed, and an acknowledgement nobody
	// re-reads is how a known bug becomes a permanent one.
	ClassificationAcknowledged Classification = "acknowledged"
	// ClassificationUnclassified is an honest absence rather than a default. It means the harness
	// could not decide from the evidence it has.
	ClassificationUnclassified Classification = "unclassified"
)

// RuleAgreement is the verdict for one rule across the whole tree.
type RuleAgreement struct {
	Rule           string
	Shared         int
	OnlyCohere     int
	OnlyGate       int
	Classification Classification
}

// Agrees reports whether the two gates said the same thing about this rule everywhere.
func (agreement RuleAgreement) Agrees() bool {
	return agreement.OnlyCohere == 0 && agreement.OnlyGate == 0
}

// Report is a whole comparison: what differed, per rule, over what population.
type Report struct {
	CoherePopulation Population
	GatePopulation   Population
	Differences      []Difference
	Agreements       []RuleAgreement
	// Comparable is false when a side reported no coverage, which makes the diff meaningless
	// regardless of how it looks. An empty Differences with Comparable false is the vacuous case.
	Comparable bool
	// NotComparableReason states which side had nothing and is empty when Comparable is true.
	NotComparableReason string
	// ComparedRules and ConfiguredRules are how many rules the comparison could speak about and
	// how many the config enables. When they differ, the verdict covers a subset of the gate and
	// must say so.
	ComparedRules   int
	ConfiguredRules int
	// StaleAcknowledgements match no current difference, and AmbiguousAcknowledgements match more
	// than one. Either fails the run: a stale entry is an excuse waiting for whatever lands on its
	// site next, and an ambiguous one would excuse more than its reason was written about, so the
	// table is out of date until someone removes or re-anchors them.
	StaleAcknowledgements     []AcknowledgedDifference
	AmbiguousAcknowledgements []AcknowledgedDifference
	// Provenance is what actually ran, and whether the harness was ever shown able to detect a
	// difference at all.
	//
	// This does not overlap Comparable and neither substitutes for the other, which is the reason
	// both are here. Comparable answers "did both sides report coverage" — a property of this run's
	// population. Provenance answers "has this instrument been demonstrated to work" — a property
	// of the instrument, which a healthy population says nothing about. A run over 3,407 files by a
	// harness that parses one side's format wrong is perfectly Comparable and completely blind.
	Provenance Provenance
}

// Agreed reports whether every rule active on both sides agreed.
//
// Rules the migration has not reached yet do not count against agreement, because a rule cohere
// has never claimed to implement cannot disagree with anything. Only ClassificationBothActive does.
// Both guards are asked before the findings are, and in this order, because each one describes a
// way the finding list can be empty for a reason that has nothing to do with the code being clean.
func (report Report) Agreed() bool {
	if !report.Comparable {
		return false
	}
	// A run whose population is real but whose controls never fired has not been shown able to
	// report a difference, so its silence is not evidence. This is the guard that four vacuous
	// probes in one night got past: every one of them had a plausible population.
	if trustworthy, _ := report.Provenance.Trustworthy(); !trustworthy {
		return false
	}
	if len(report.StaleAcknowledgements) > 0 || len(report.AmbiguousAcknowledgements) > 0 {
		return false
	}
	// Observed only. A planted control is a difference this run caused on purpose, and counting it
	// against agreement would mean the harness could never agree with itself: proving detection
	// requires creating a difference, so a control that classified both-active would fail the very
	// run that demonstrated the instrument works.
	for _, difference := range report.ObservedDifferences() {
		if difference.Classification == ClassificationBothActive {
			return false
		}
	}
	return true
}

// Inputs is everything Compare needs that it cannot derive from the findings themselves.
type Inputs struct {
	CohereFindings   []Finding
	GateFindings     []Finding
	CoherePopulation Population
	GatePopulation   Population
	// CohereRules is every rule name compiled into cohere. A gate finding whose rule is absent here
	// is not-ported rather than a disagreement.
	CohereRules map[string]bool
	// ConfiguredRules is every rule name the lint config enables. A cohere rule absent here ran over
	// no files by design.
	ConfiguredRules map[string]bool
	// Acknowledged are differences someone decided are correct, each with a reason. Nil means none,
	// which is the honest default: an empty list excuses nothing.
	Acknowledged []AcknowledgedDifference
	// ReadSourceLine reads a line of the tree, which is how an acknowledgement's anchor is checked
	// against the finding it names. Nil leaves every acknowledgement stale rather than trusted.
	ReadSourceLine SourceLineReader
}

// Compare diffs two runs and classifies every difference.
func Compare(inputs Inputs) Report {
	// Per-rule tallies of how many differences were acknowledged, so a rule whose differences are
	// all accounted for is not reported as an open disagreement.
	type ruleDifferenceCounts struct{ total, acknowledged int }
	differencesForRule := map[string]*ruleDifferenceCounts{}
	countDifference := func(ruleName string, classification Classification) {
		counts := differencesForRule[ruleName]
		if counts == nil {
			counts = &ruleDifferenceCounts{}
			differencesForRule[ruleName] = counts
		}
		counts.total++
		if classification == ClassificationAcknowledged {
			counts.acknowledged++
		}
	}

	report := Report{
		CoherePopulation: inputs.CoherePopulation,
		GatePopulation:   inputs.GatePopulation,
		Comparable:       true,
		ComparedRules:    len(inputs.CohereRules),
		ConfiguredRules:  len(inputs.ConfiguredRules),
	}

	// The vacuity guard, before any diffing. A side that walked no files cannot be compared against
	// one that did, and the diff would look clean rather than broken, so this is a refusal and not a
	// warning. This is the exact shape of failure the gate cohere replaces shipped: green over zero.
	switch {
	case inputs.CoherePopulation.FilesWalked == 0:
		report.Comparable = false
		report.NotComparableReason = "cohere walked no files, so its silence is not a result"
	case inputs.GatePopulation.FilesWalked == 0:
		report.Comparable = false
		report.NotComparableReason = "the gate walked no files, so its silence is not a result"
	}

	cohereByKey := indexByKey(inputs.CohereFindings)
	gateByKey := indexByKey(inputs.GateFindings)

	perRule := map[string]*RuleAgreement{}
	agreementFor := func(ruleName string) *RuleAgreement {
		if existing, found := perRule[ruleName]; found {
			return existing
		}
		created := &RuleAgreement{Rule: ruleName}
		perRule[ruleName] = created
		return created
	}

	for key, finding := range cohereByKey {
		agreement := agreementFor(finding.Rule)
		if _, sharedWithGate := gateByKey[key]; sharedWithGate {
			agreement.Shared++
			continue
		}
		agreement.OnlyCohere++
		report.Differences = append(report.Differences, Difference{
			Finding:        finding,
			OnlyOn:         SideCohere,
			Classification: classify(finding.Rule, SideCohere, inputs),
		})
	}

	for key, finding := range gateByKey {
		if _, sharedWithVerify := cohereByKey[key]; sharedWithVerify {
			continue
		}
		agreement := agreementFor(finding.Rule)
		agreement.OnlyGate++
		report.Differences = append(report.Differences, Difference{
			Finding:        finding,
			OnlyOn:         SideGate,
			Classification: classify(finding.Rule, SideGate, inputs),
		})
	}

	// Acknowledgements are matched once every difference is known, because an entry is judged by how
	// many it names: exactly one is excused, none leaves the entry stale, and several leave it
	// ambiguous with none of them excused.
	for _, match := range matchAcknowledgements(inputs.Acknowledged, report.Differences, inputs.ReadSourceLine) {
		switch len(match.differenceIndexes) {
		case 0:
			report.StaleAcknowledgements = append(report.StaleAcknowledgements, match.acknowledgement)
		case 1:
			report.Differences[match.differenceIndexes[0]].Classification = ClassificationAcknowledged
		default:
			report.AmbiguousAcknowledgements = append(report.AmbiguousAcknowledgements, match.acknowledgement)
		}
	}
	for _, difference := range report.Differences {
		countDifference(difference.Finding.Rule, difference.Classification)
	}

	// A rule's own classification is asked from the side that actually differed, because
	// not-ported is only meaningful for a gate finding: cohere cannot report a rule it does not
	// compile. A rule that agreed everywhere is classified from the cohere side, where "both sides
	// had it on" is the true and useful answer.
	for ruleName, agreement := range perRule {
		side := SideCohere
		if agreement.OnlyGate > 0 && agreement.OnlyCohere == 0 {
			side = SideGate
		}
		agreement.Classification = classify(ruleName, side, inputs)

		// A rule whose every difference was acknowledged is acknowledged, not both-active. Leaving
		// it both-active would print the rule as an unresolved defect beside a per-finding line
		// saying it is resolved, and a reader trusts the summary over the detail.
		if differences := differencesForRule[ruleName]; differences != nil && differences.total == differences.acknowledged {
			agreement.Classification = ClassificationAcknowledged
		}
		report.Agreements = append(report.Agreements, *agreement)
	}

	sort.Slice(report.Agreements, func(first, second int) bool {
		return report.Agreements[first].Rule < report.Agreements[second].Rule
	})
	sort.Slice(report.Differences, func(first, second int) bool {
		left, right := report.Differences[first].Finding, report.Differences[second].Finding
		if left.File != right.File {
			return left.File < right.File
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		return left.Rule < right.Rule
	})

	return report
}

// classify decides why a rule's findings differ, from the two runs alone.
func classify(ruleName string, onlyOn Side, inputs Inputs) Classification {
	knownToCohere := inputs.CohereRules[ruleName]
	configured := inputs.ConfiguredRules[ruleName]

	// A nil map means the caller never read the lint config, which is not the same as a config that
	// enables nothing — and Go cannot tell them apart, because reading any key of a nil map returns
	// false. Without this the not-configured branch would swallow every rule and quietly excuse
	// every genuine disagreement, producing a clean tree with maximum confidence. So absent
	// knowledge degrades to the two-way split, which is honest about what the caller actually has.
	configurationKnown := inputs.ConfiguredRules != nil

	switch {
	// The gate found it and cohere has no such rule. Expected during the migration, and the reason
	// a raw count of differences is not a measure of disagreement.
	case onlyOn == SideGate && !knownToCohere:
		return ClassificationNotPorted
	// Cohere has the rule but nothing turned it on, so it walked no files.
	case knownToCohere && configurationKnown && !configured:
		return ClassificationNotConfigured
	// Both sides have it and we cannot say whether it was enabled. Treating that as a disagreement
	// is the conservative direction: it surfaces for a human instead of being filed as expected.
	case knownToCohere && !configurationKnown:
		return ClassificationBothActive
	// Both sides had this rule and both had it on. Somebody is wrong.
	case knownToCohere && configured:
		return ClassificationBothActive
	default:
		return ClassificationUnclassified
	}
}

func indexByKey(findings []Finding) map[string]Finding {
	byKey := make(map[string]Finding, len(findings))
	for _, finding := range findings {
		byKey[finding.Key()] = finding
	}
	return byKey
}

// Write renders a whole comparison for a person.
//
// The population line comes first and unconditionally, before any verdict, because the question a
// reader must be able to answer before believing a diff is "did both of these actually run."
func Write(out *strings.Builder, report Report) {
	fmt.Fprintf(out, "population: cohere %d findings over %d files (%d rules) · gate %d findings over %d files\n",
		report.CoherePopulation.Findings, report.CoherePopulation.FilesWalked, report.CoherePopulation.Rules,
		report.GatePopulation.Findings, report.GatePopulation.FilesWalked,
	)

	// Provenance prints second and unconditionally, next to the population and above the verdict,
	// because a reader deciding whether to believe a diff needs both halves of the question in one
	// place: did both sides look, and has this harness ever been shown able to see.
	fmt.Fprint(out, report.Provenance.Describe())

	if !report.Comparable {
		fmt.Fprintf(out, "✗ not comparable: %s\n", report.NotComparableReason)
		return
	}

	// An untrustworthy run stops here too. It is a different refusal from not-comparable and says
	// so, rather than printing a difference list that a reader would take as a measurement.
	if trustworthy, reasons := report.Provenance.Trustworthy(); !trustworthy {
		fmt.Fprintf(out, "✗ not trustworthy, so the difference list below is not a measurement:\n")
		for _, reason := range reasons {
			fmt.Fprintf(out, "    %s\n", reason)
		}
	}

	// Counted over observed differences only. The planted controls are differences in the same
	// sense and appear in the table below, but a reader trusting this line is asking how far apart
	// the two gates are, and the controls are the harness's own footprint rather than an answer to
	// that. A run reporting "differences: 3 total" on a tree with a real gap of zero was read as
	// three problems and nearly published as the remaining exposure.
	observed := report.ObservedDifferences()
	byClassification := map[Classification]int{}
	for _, difference := range observed {
		byClassification[difference.Classification]++
	}

	plantedCount := len(report.Differences) - len(observed)
	fmt.Fprintf(out, "\ndifferences: %d observed, %d planted by this harness — %d both-active, %d not-ported, %d not-configured, %d unclassified\n",
		len(observed), plantedCount,
		byClassification[ClassificationBothActive],
		byClassification[ClassificationNotPorted],
		byClassification[ClassificationNotConfigured],
		byClassification[ClassificationUnclassified],
	)

	// Both-active differences print in full, individually, because each one is a defect in one of
	// the two gates and a count alone is not actionable. The other classes print per rule, because
	// a hundred not-ported findings from one unported rule is one fact, not a hundred.
	for _, difference := range report.Differences {
		if difference.Classification != ClassificationBothActive {
			continue
		}
		planted := ""
		if difference.Planted {
			planted = "  [planted control, not an observed difference]"
		}
		fmt.Fprintf(out, "\n  %s:%d  %s%s\n    only on: %s\n    %s\n",
			difference.Finding.File, difference.Finding.Line, difference.Finding.Rule, planted,
			difference.OnlyOn, difference.Finding.Message,
		)
	}

	// An out-of-date acknowledgement table prints before the per-rule table, because it is a reason
	// the verdict below fails that the rule table cannot show: a stale entry differs from nothing.
	if len(report.StaleAcknowledgements) > 0 || len(report.AmbiguousAcknowledgements) > 0 {
		fmt.Fprintf(out, "\nacknowledgements that excuse no single current difference:\n")
		for _, stale := range report.StaleAcknowledgements {
			fmt.Fprintf(out, "  stale      %s  (matches none: remove it, or re-anchor it if its site moved)\n", stale.Key())
		}
		for _, ambiguous := range report.AmbiguousAcknowledgements {
			fmt.Fprintf(out, "  ambiguous  %s  (matches several: narrow its anchor to one)\n", ambiguous.Key())
		}
	}

	fmt.Fprintf(out, "\nper-rule agreement:\n")
	for _, agreement := range report.Agreements {
		verdict := "agree"
		if !agreement.Agrees() {
			verdict = "DIFFER"
		}
		fmt.Fprintf(out, "  %-8s %-52s shared %-5d only-cohere %-5d only-gate %-5d  %s\n",
			verdict, agreement.Rule, agreement.Shared, agreement.OnlyCohere, agreement.OnlyGate,
			agreement.Classification,
		)
	}

	if report.Agreed() {
		// The verdict states its own scope, because the sentence a reader carries away is this one
		// and it is easy to read as broader than the test. A comparison can only speak about rules
		// cohere implements: the rest of the config is enabled in the gate, unported, and was never
		// compared. Silence about them reads as coverage.
		//
		// This matters more than it looks. Every finding the gate produces on this tree today comes
		// from a rule cohere already has, and the 138 unported rules find nothing — not because
		// they are worthless, but because the gate has been enforcing them for months and the tree
		// is clean of what they prevent. A clean diff therefore cannot be the signal to remove the
		// old tool: it would keep reading clean right up until somebody wrote a violation of an
		// unported rule, and then keep reading clean while the tree got worse.
		if report.ConfiguredRules > report.ComparedRules {
			fmt.Fprintf(out,
				"\n✓ agrees on the %d rules cohere implements, of %d the config enables\n"+
					"  the other %d are unported and were not compared, so this is not a verdict about them\n",
				report.ComparedRules, report.ConfiguredRules, report.ConfiguredRules-report.ComparedRules,
			)
			printAcknowledgedCaveat(out, byClassification)
			return
		}
		fmt.Fprintf(out, "\n✓ agrees: every rule active on both sides reported the same findings\n")
		printAcknowledgedCaveat(out, byClassification)
		return
	}

	// The two ways to fail say different things, and printing the wrong one is its own defect: a
	// clean tree measured by a blind harness would otherwise read as "0 findings differ", which is
	// a true number and a false claim.
	if trustworthy, _ := report.Provenance.Trustworthy(); !trustworthy {
		fmt.Fprintf(out, "\n✗ no verdict: the harness was not shown able to detect a difference on this run\n")
		return
	}

	fmt.Fprintf(out, "\n✗ disagrees: %d findings differ on rules both sides had enabled\n",
		byClassification[ClassificationBothActive],
	)
	if outOfDate := len(report.StaleAcknowledgements) + len(report.AmbiguousAcknowledgements); outOfDate > 0 {
		fmt.Fprintf(out, "  %d acknowledgements excuse no single current difference, listed above\n", outOfDate)
	}
	if acknowledgedCount := byClassification[ClassificationAcknowledged]; acknowledgedCount > 0 {
		fmt.Fprintf(out, "  %d further differences were acknowledged and are not counted above\n", acknowledgedCount)
	}
}
