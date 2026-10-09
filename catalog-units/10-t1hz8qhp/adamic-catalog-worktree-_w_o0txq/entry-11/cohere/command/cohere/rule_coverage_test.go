package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// renderLintReport writes a lint report to a string, in the default mode unless the report asks for
// details.
func renderLintReport(report lintReport) string {
	var out strings.Builder
	writeLintReport(&out, report)
	return out.String()
}

// countedLineTerms parses the counted line back into its total and its terms, so a test can check the
// arithmetic the reader is asked to check rather than the struct behind it.
func countedLineTerms(t *testing.T, output string) (int, map[string]int) {
	t.Helper()
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "coverage: ") {
			continue
		}
		head, tail, _ := strings.Cut(strings.TrimPrefix(line, "coverage: "), " = ")
		total, err := strconv.Atoi(strings.TrimSuffix(head, " rules"))
		if err != nil {
			t.Fatalf("the counted line does not start with a rule total: %q", line)
		}
		terms := map[string]int{}
		if tail == "" {
			return total, terms
		}
		for _, term := range strings.Split(tail, " + ") {
			count, label, _ := strings.Cut(term, " ")
			value, err := strconv.Atoi(count)
			if err != nil {
				t.Fatalf("a term with no count: %q in %q", term, line)
			}
			terms[label] = value
		}
		return total, terms
	}
	t.Fatalf("no counted line in:\n%s", output)
	return 0, nil
}

// The counts add up to the rule total, for every shape a rule's coverage can take.
//
// Every combination of the five per-rule numbers is a rule here, plus rules the config asks for that
// the binary lacks, so a rule whose shape no case of the classifier names would fall out of the sum and
// fail this rather than vanish from the line silently. The arithmetic is read back off the printed
// line, because that line is what a reader checks.
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestCoverageCountsAddUpToTheRuleTotal(t *testing.T) {
	rules := []rule.Rule{}
	coverage := program.Coverage{
		RulesOffered:      map[string]int{},
		RulesListening:    map[string]int{},
		RulesReporting:    map[string]int{},
		RulesScopedOff:    map[string]int{},
		RulesUnconfigured: map[string]int{},
	}
	for shape := range 32 {
		name := fmt.Sprintf("rule-%02d", shape)
		rules = append(rules, rule.Rule{Name: name})
		for bit, counts := range []map[string]int{
			coverage.RulesReporting, coverage.RulesListening, coverage.RulesOffered,
			coverage.RulesScopedOff, coverage.RulesUnconfigured,
		} {
			if shape&(1<<bit) != 0 {
				counts[name] = 7 + bit
			}
		}
	}
	config := &configuration.Config{Rules: map[string]configuration.RuleSetting{
		"rule-00":         {Severity: configuration.SeverityError},
		"unported-one":    {Severity: configuration.SeverityError},
		"plugin/unported": {Severity: configuration.SeverityWarn},
		"unported-off":    {Severity: configuration.SeverityOff},
	}}

	summary := classifyTypeScriptCoverage(rules, coverage, config)
	if summary.total() != len(rules)+2 {
		t.Fatalf("the summary accounts for %d rules, want %d registered plus 2 not ported", summary.total(), len(rules))
	}
	sum := 0
	for _, entry := range summary.Entries {
		if _, known := coverageCategoryTerms[entry.Category]; !known {
			t.Fatalf("rule %s landed in category %q, which the counted line has no term for", entry.Name, entry.Category)
		}
	}
	for _, category := range coverageCategoryOrder {
		sum += summary.count(category)
	}
	if sum != summary.total() {
		t.Fatalf("the categories sum to %d and the total is %d", sum, summary.total())
	}

	output := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, LintConfig: config, WalkCost: "in 1s"})
	total, terms := countedLineTerms(t, output)
	printedSum := 0
	for _, value := range terms {
		printedSum += value
	}
	if total != len(rules)+2 || printedSum != total {
		t.Fatalf("the counted line says %d rules and its terms sum to %d, want %d both:\n%s", total, printedSum, len(rules)+2, output)
	}

	// The Swift record's parts add up to its whole the same way.
	swift, err := classifySwiftCoverage(&swiftLintRecord{
		RulesRun: 9, RulesWatchedAndQuiet: 3, RulesSilent: []string{"a", "b", "c"},
		RulesNotConfigured: []string{"a"}, RulesScopedOff: map[string]int{"b": 2, "quiet-but-partly-off": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if swift.total() != 9 || swift.count(coverageFoundSomething) != 3 {
		t.Fatalf("the Swift summary accounts for %d rules with %d found something, want 9 and 3", swift.total(), swift.count(coverageFoundSomething))
	}
}

// A Swift record whose parts add up to more than its whole is refused rather than printed, because the
// counted line would state arithmetic nobody can vouch for.
func TestSwiftCoverageRefusesPartsLargerThanTheWhole(t *testing.T) {
	t.Parallel()
	_, err := classifySwiftCoverage(&swiftLintRecord{RulesRun: 2, RulesWatchedAndQuiet: 2, RulesSilent: []string{"a"}})
	if err == nil || !strings.Contains(err.Error(), "more than that") {
		t.Fatalf("a record counting 3 rules among 2 was accepted: %v", err)
	}
}

// The coverage categories distinguish a rule nobody wired from a rule that was offered files and
// registered nothing.
//
// Both states produce zero findings and both used to print "listened to no files", so a passing rule
// read exactly like a dead one. They are opposite defects: one was handed files and declined them, the
// other was never handed anything. The pair was once separable only because a second line about missing
// config happened to print for one of them, which is why this asserts the categories themselves.
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestCoverageSeparatesUnwiredFromOfferedAndSilent(t *testing.T) {
	rules := []rule.Rule{{Name: "satisfied-rule", NoListener: rule.NoListenerDeclinesIrrelevantFiles}, {Name: "unwired-rule"}}
	coverage := program.Coverage{
		RulesOffered:   map[string]int{"satisfied-rule": 3407},
		RulesListening: map[string]int{},
	}

	output := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, WalkCost: "in 1s", Details: true})
	requireLines(t, output,
		"coverage: 2 rules = 1 registered no listener + 1 offered no files\n",
		"  registered no listener (1): offered files and registered no listener on any",
		"    satisfied-rule (offered 3407 files, declines files it is not about)\n",
		"  offered no files (1): nothing wired it",
		"    unwired-rule\n",
	)
}

// A rule that ran is counted on the default line and not named there.
//
// Naming every rule that ran and found nothing is the wall this replaced: 370 of them on ahra, which
// buried the lines a reader must act on. The count is the honest summary, and the names are one flag
// away.
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestARuleThatRanIsCountedAndNotNamedByDefault(t *testing.T) {
	rules := []rule.Rule{{Name: "quiet-rule"}, {Name: "working-rule"}, {Name: "eager-rule", NoListener: rule.NoListenerAnswersInRun}, {Name: "unconfigured-rule"}}
	coverage := program.Coverage{
		RulesOffered:      map[string]int{"quiet-rule": 3407, "working-rule": 3407, "eager-rule": 3407},
		RulesListening:    map[string]int{"quiet-rule": 3407, "working-rule": 412},
		RulesReporting:    map[string]int{"working-rule": 7},
		RulesUnconfigured: map[string]int{"unconfigured-rule": 3407},
	}

	output := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, WalkCost: "in 1s"})
	requireLines(t, output,
		"coverage: 4 rules = 1 found something + 1 ran and found nothing + 1 registered no listener + 1 nobody has configured\n",
		"details: cohere --coverage",
	)
	forbidLines(t, output, "quiet-rule", "working-rule", "eager-rule", "unconfigured-rule", "coverage details")
}

// Under --coverage every rule is named exactly once, and no rule appears in two categories.
//
// The duplicates this replaced were per-rule: a rule nobody configured printed as "offered no files"
// and again as "not in the config", and a rule the config turned off printed as "offered no files" and
// again as "scoped off". The shapes here are the ones that used to print twice, plus the partial ones
// whose second fact now rides beside the name.
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestCoverageDetailsNameEachRuleOnce(t *testing.T) {
	rules := []rule.Rule{
		{Name: "off-everywhere"}, {Name: "nobody-configured"}, {Name: "partly-off"},
		{Name: "override-only"}, {Name: "off-and-unconfigured"}, {Name: "quiet"}, {Name: "eager", NoListener: rule.NoListenerAnswersInRun},
	}
	coverage := program.Coverage{
		RulesOffered:      map[string]int{"partly-off": 3779, "override-only": 40, "quiet": 3785, "eager": 3785},
		RulesListening:    map[string]int{"partly-off": 3779, "override-only": 40, "quiet": 3785},
		RulesReporting:    map[string]int{"override-only": 2},
		RulesScopedOff:    map[string]int{"off-everywhere": 3785, "partly-off": 6, "off-and-unconfigured": 3000},
		RulesUnconfigured: map[string]int{"nobody-configured": 3785, "override-only": 3745, "off-and-unconfigured": 785},
	}
	config := &configuration.Config{Rules: map[string]configuration.RuleSetting{"not-ported": {Severity: configuration.SeverityError}}}

	output := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, LintConfig: config, WalkCost: "in 1s", Details: true})
	lines := strings.Split(output, "\n")
	for _, name := range []string{"off-everywhere", "nobody-configured", "partly-off", "override-only", "off-and-unconfigured", "quiet", "eager", "not-ported"} {
		named := 0
		for _, line := range lines {
			if line == "    "+name || strings.HasPrefix(line, "    "+name+" (") {
				named++
			}
		}
		if named != 1 {
			t.Errorf("rule %s is named on %d lines under --coverage, want exactly 1:\n%s", name, named, output)
		}
	}
	requireLines(t, output,
		"    partly-off (ran on 3779 of 3785 files, off by the config on 6)\n",
		"    override-only (2 findings, not in the config for 3745 files)\n",
		"    off-and-unconfigured (3000 files, not in the config for 785 more)\n",
	)
	// The pointer is for the default mode; with details it would point at itself.
	forbidLines(t, output, "details: cohere --coverage")
}

// A rule on at the top level and off by an override for most files is counted apart from a rule that
// ran everywhere, and named with where it ran and where it did not.
//
// This is phi_api's shape (#7n4zxrb): core rules on in the base `rules`, an override turning them off
// for `**/*.ts`, and a `.cjs` file still linted. They ran on 12 of 3,242 files and coverage counted
// them with the rules that ran on all of them. The control is a rule that ran on every file, which
// stays where it was; a rule that found something keeps that category and says where it did not run.
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestARuleOffByAnOverrideForMostFilesIsCountedApart(t *testing.T) {
	rules := []rule.Rule{{Name: "eqeqeq"}, {Name: "no-var"}, {Name: "everywhere"}, {Name: "override-only"}}
	coverage := program.Coverage{
		RulesOffered:      map[string]int{"eqeqeq": 12, "no-var": 12, "everywhere": 3242, "override-only": 40},
		RulesListening:    map[string]int{"eqeqeq": 12, "no-var": 9, "everywhere": 3242, "override-only": 40},
		RulesReporting:    map[string]int{},
		RulesScopedOff:    map[string]int{"eqeqeq": 3230, "no-var": 3230},
		RulesUnconfigured: map[string]int{"override-only": 3202},
	}

	report := lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, WalkCost: "in 1s"}
	requireLines(t, renderLintReport(report),
		"coverage: 4 rules = 3 ran on part of the tree + 1 ran and found nothing\n",
	)

	report.Details = true
	output := renderLintReport(report)
	requireLines(t, output,
		"  ran on part of the tree (3): ",
		"    eqeqeq (ran on 12 of 3242 files, off by the config on 3230)\n",
		// A rule offered files it declined says so, so the terms still add up to the files it could
		// have run on.
		"    no-var (ran on 9 of 3242 files, declined 3, off by the config on 3230)\n",
		// A rule enabled only inside an override is the same fact with the other reason.
		"    override-only (ran on 40 of 3242 files, not in the config for 3202)\n",
		"  ran and found nothing (1): ",
		"    everywhere\n",
	)
}

// Every category that needs action prints in full on a default run, whatever else moved behind
// --coverage. One subtest per category in the design, so dropping any of them from the default output
// fails by name.
// Not parallel: its subtests render through writeLintReport and the Swift record stream, which set the package-level
// activeSummary
func TestActionableCoverageAlwaysPrintsByDefault(t *testing.T) {
	cleanCoverage := func() program.Coverage {
		return program.Coverage{
			RulesOffered:   map[string]int{"quiet-rule": 10},
			RulesListening: map[string]int{"quiet-rule": 10},
		}
	}
	rules := []rule.Rule{{Name: "quiet-rule"}}

	// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
	t.Run("findings", func(t *testing.T) {
		coverage := cleanCoverage()
		coverage.RulesReporting = map[string]int{"quiet-rule": 1}
		output := renderLintReport(lintReport{
			Result: program.Result{
				Diagnostics: []rule.Diagnostic{{RuleName: "quiet-rule", Message: rule.Message{Id: "m", Description: "this is wrong"}}},
				Coverage:    coverage,
			},
			Rules: rules, WalkCost: "in 1s",
		})
		requireLines(t, output, "error quiet-rule: this is wrong\n", "lint: 1 findings", "1 found something")
	})

	// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
	t.Run("crashed files", func(t *testing.T) {
		coverage := cleanCoverage()
		coverage.FilesCrashed = []program.FileCrash{{FileName: "/project/Broken.ts", Cause: fmt.Errorf("Node.Text on a kind it does not handle")}}
		output := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, WalkCost: "in 1s"})
		requireLines(t, output,
			"  1 files crashed",
			"  crashed: /project/Broken.ts: cohere crashed on this file (Node.Text on a kind it does not handle). This is a bug in cohere, not in your code: nothing in this file was checked.",
		)
	})

	// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
	t.Run("a rule that could not finish a file", func(t *testing.T) {
		coverage := cleanCoverage()
		coverage.RulesCrashed = []program.RuleCrash{{RuleName: "prefer-arrow-callback", FileName: "/project/Broken.ts", Cause: fmt.Errorf("interface conversion")}}
		output := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, WalkCost: "in 1s"})
		requireLines(t, output,
			"  0 files crashed · 1 rule crashes, each costing one rule one file",
			"  crashed: /project/Broken.ts: cohere's rule prefer-arrow-callback crashed on this file (interface conversion). This is a bug in cohere, not in your code: the rule's verdict on this file is missing, and the file's other rules ran.",
		)
	})

	// Not parallel: it drives the Swift record stream through renderRecords, whose acceptors set the package-level
	// activeSummary
	t.Run("unreadable files", func(t *testing.T) {
		output, _, _ := renderRecords(t, swiftModeCheck, contractFixture(t, "Unreadable.jsonl"), 1)
		requireLines(t, output, "  not checked: /project/Sources/Example/Latin1.swift (could not be read:")
	})

	// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
	t.Run("config keys that match no rule", func(t *testing.T) {
		config := &configuration.Config{Rules: map[string]configuration.RuleSetting{
			"quiet-rule":   {Severity: configuration.SeverityError},
			"no-dupe-keys": {Severity: configuration.SeverityOff},
		}}
		output := renderLintReport(lintReport{Result: program.Result{Coverage: cleanCoverage()}, Rules: rules, LintConfig: config, WalkCost: "in 1s"})
		requireLines(t, output, `  config: key "no-dupe-keys" matches no registered rule, so its off never applies`)
	})

	// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
	t.Run("a rule offered files that registered nothing when it should have", func(t *testing.T) {
		// Three rules with the same numbers: offered 40 files, listened to none, reported nothing. Only
		// the one declaring no reason is dead, and only it prints by default (#j69gvka). Before the
		// declaration every such rule was presumed eager, so the inert one printed nothing.
		coverage := cleanCoverage()
		for _, name := range []string{"inert-rule", "eager-rule", "declining-rule"} {
			coverage.RulesOffered[name] = 40
		}
		silentRules := append([]rule.Rule{
			{Name: "inert-rule"},
			{Name: "eager-rule", NoListener: rule.NoListenerAnswersInRun},
			{Name: "declining-rule", NoListener: rule.NoListenerDeclinesIrrelevantFiles},
		}, rules...)
		output := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: silentRules, WalkCost: "in 1s"})
		requireLines(t, output, "  no listener: rule inert-rule was offered 40 files and registered no listener on any, and it declares no reason it may (answering in Run, or declining files it is not about), so it checked nothing\n")
		forbidLines(t, output, "eager-rule", "declining-rule")

		// And with details each is named once, under its category, with what it declared.
		detailed := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: silentRules, WalkCost: "in 1s", Details: true})
		if strings.Count(detailed, "inert-rule") != 1 {
			t.Errorf("a rule needing action is named %d times under --coverage, want 1:\n%s", strings.Count(detailed, "inert-rule"), detailed)
		}
		requireLines(t, detailed,
			"    declining-rule (offered 40 files, declines files it is not about)\n",
			"    eager-rule (offered 40 files, answers in Run)\n",
			"    inert-rule (offered 40 files, declares no reason to register none, so it checked nothing)\n",
		)
	})

	// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
	t.Run("suppressions without a reason", func(t *testing.T) {
		coverage := cleanCoverage()
		coverage.Suppressed, coverage.SuppressedWithoutReason = 5, 2
		output := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, WalkCost: "in 1s"})
		requireLines(t, output, "5 findings suppressed, 2 without a reason")
	})

	// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
	t.Run("disable comments that silenced nothing while their rule ran", func(t *testing.T) {
		coverage := cleanCoverage()
		coverage.UnusedSuppressions, coverage.UnusedSuppressionsForUnrunRules = 16, 13
		output := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, WalkCost: "in 1s"})
		requireLines(t, output, "3 disable comments silenced nothing while their rule ran")
		// The load-bearing ones are a detail, not an action.
		forbidLines(t, output, "13 unused disable comments")
	})

	// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
	t.Run("dead disable comments are named under --coverage", func(t *testing.T) {
		coverage := cleanCoverage()
		coverage.UnusedSuppressions = 3
		// Out of order, so the listing's sort is under test, and one blanket directive naming no rule.
		coverage.DeadSuppressions = []program.DeadSuppression{
			{File: "/repository/source/Zebra.ts", Line: 4, Rules: []string{"no-console"}},
			{File: "/repository/source/Alpha.ts", Line: 40, Rules: []string{"eqeqeq", "no-var"}},
			{File: "/repository/source/Alpha.ts", Line: 2},
		}
		detailed := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, WalkCost: "in 1s", Details: true})
		requireLines(t, detailed,
			"  dead disable comments (3): each silenced nothing while a rule it names ran\n"+
				"    /repository/source/Alpha.ts:3 every rule\n"+
				"    /repository/source/Alpha.ts:41 eqeqeq, no-var\n"+
				"    /repository/source/Zebra.ts:5 no-console\n",
		)

		// The default report keeps the count and names no site: the sites are a detail.
		output := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, WalkCost: "in 1s"})
		requireLines(t, output, "3 disable comments silenced nothing while their rule ran")
		forbidLines(t, output, "Alpha.ts", "dead disable comments (")
	})

	// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
	t.Run("parity", func(t *testing.T) {
		config := &configuration.Config{Rules: map[string]configuration.RuleSetting{
			"quiet-rule":    {Severity: configuration.SeverityError},
			"not-yet-here":  {Severity: configuration.SeverityError},
			"turned-off-be": {Severity: configuration.SeverityOff},
		}}
		output := renderLintReport(lintReport{Result: program.Result{Coverage: cleanCoverage()}, Rules: rules, LintConfig: config, WalkCost: "in 1s"})
		requireLines(t, output,
			"  parity: 1 of 2 rules the config asks for, so 1 were not checked by anything here\n",
			"1 not ported",
		)
	})

	t.Run("the modified-tree warning", func(t *testing.T) {
		t.Parallel()
		var out strings.Builder
		(&pipelineReport{}).writeProvenanceWarning(&out, true)
		requireLines(t, out.String(), "this binary was built from a modified tree")
	})

	t.Run("this run did not check everything", func(t *testing.T) {
		t.Parallel()
		for name, report := range map[string]*pipelineReport{
			"a phase cut off": func() *pipelineReport {
				report := &pipelineReport{}
				report.record(phaseFix, outcomeSkipped, 0, "--no-fix")
				report.record(phaseTypes, outcomeRan, 0, "")
				report.markRemainingNotReached(phaseTypes, "a type diagnostic")
				return report
			}(),
			"a file a rule crashed on": func() *pipelineReport {
				report := &pipelineReport{incompleteBeyondPhases: namedGapsSentence}
				report.record(phaseFix, outcomeSkipped, 0, "--no-fix")
				report.record(phaseTypes, outcomeRan, 0, "")
				report.record(phaseLint, outcomeRan, 0, "")
				report.record(phaseUnused, outcomeSkipped, 0, "not requested")
				return report
			}(),
		} {
			var out strings.Builder
			report.Write(&out)
			if !strings.Contains(out.String(), "this run did not check everything") {
				t.Errorf("%s: no warning:\n%s", name, out.String())
			}
		}
	})
}

// A clean run prints the counts and nothing per rule: the shape of the default output, held so a
// later printer that adds a per-rule line to every run has to change this test to do it.
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestACleanRunPrintsTwoCoverageLines(t *testing.T) {
	rules := []rule.Rule{{Name: "quiet-rule"}, {Name: "off-rule"}, {Name: "unconfigured-rule"}, {Name: "eager-rule", NoListener: rule.NoListenerAnswersInRun}}
	coverage := program.Coverage{
		RulesOffered:      map[string]int{"quiet-rule": 10, "eager-rule": 10},
		RulesListening:    map[string]int{"quiet-rule": 10},
		RulesScopedOff:    map[string]int{"off-rule": 10},
		RulesUnconfigured: map[string]int{"unconfigured-rule": 10},
		Suppressed:        4,
	}
	output := renderLintReport(lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, WalkCost: "in 1s"})
	want := "lint: 0 findings — 0 rules over 0 files, 0 nodes visited, in 1s\n" +
		"coverage: 4 rules = 1 ran and found nothing + 1 registered no listener + 1 off by the config + 1 nobody has configured\n" +
		"  0 files crashed · 4 findings suppressed, 0 without a reason · details: cohere --coverage\n"
	if output != want {
		t.Errorf("a clean run printed:\n%s\nwant:\n%s", output, want)
	}
}

// A departure from a house ruling prints on every run, with the file and the reason, so a reasoned one
// can never become a quiet allowance (#rkm5a31). The control is a config with none, which prints no
// departure line at all.
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestADepartureFromAHouseRulingPrintsByDefault(t *testing.T) {
	root := "/repository"
	withDeparture := &configuration.Config{
		Rules: map[string]configuration.RuleSetting{"guard-for-in": {Severity: configuration.SeverityError}},
		Root:  root,
		Departures: map[string]configuration.Departure{
			"guard-for-in": {File: root + "/CohereSettings.json", Reason: "no for-in replacement here yet"},
		},
	}
	output := renderLintReport(lintReport{LintConfig: withDeparture})
	want := "  departure: guard-for-in is set differently from the house ruling in CohereSettings.json: no for-in replacement here yet"
	if !strings.Contains(output, want+"\n") {
		t.Errorf("the departure did not print as\n%s\nin:\n%s", want, output)
	}

	without := &configuration.Config{
		Rules: map[string]configuration.RuleSetting{"guard-for-in": {Severity: configuration.SeverityError}},
		Root:  root,
	}
	if output := renderLintReport(lintReport{LintConfig: without}); strings.Contains(output, "departure:") {
		t.Errorf("a config with no departures printed one:\n%s", output)
	}
}

// An override that gives a reason prints it on every run, with its files, its file and the rules it
// sets, so an override standing in for unfinished work stays in front of every reader until it goes.
// The control is an override with no reason beside it, which prints nothing: a scoped override needs
// no reason, and printing every one would bury the few that carry one.
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestAnOverrideWithAReasonPrintsByDefault(t *testing.T) {
	root := "/repository"
	config := &configuration.Config{
		Root: root,
		Overrides: []configuration.Override{
			{
				Files: []string{"**/*.test.ts"},
				Rules: map[string]configuration.RuleSetting{"max-lines": {Severity: configuration.SeverityOff}},
				File:  root + "/CohereSettings.json",
			},
			{
				Files: []string{"**/api/graphql/*.generated.ts"},
				Rules: map[string]configuration.RuleSetting{
					"nexus/consistency-no-abbreviated-identifier": {Severity: configuration.SeverityOff},
					"@typescript-eslint/no-explicit-any":          {Severity: configuration.SeverityOff},
				},
				Reason: "until the generator emits real scalar types",
				File:   root + "/structure/StructureCohereSettings.json",
			},
		},
	}
	output := renderLintReport(lintReport{LintConfig: config})
	want := "  override: **/api/graphql/*.generated.ts in structure/StructureCohereSettings.json sets " +
		"@typescript-eslint/no-explicit-any off, nexus/consistency-no-abbreviated-identifier off: " +
		"until the generator emits real scalar types"
	if !strings.Contains(output, want+"\n") {
		t.Errorf("the override did not print as\n%s\nin:\n%s", want, output)
	}
	if strings.Count(output, "override:") != 1 {
		t.Errorf("only the override with a reason should print, got:\n%s", output)
	}
}

// An off says why beside the rule wherever coverage names it, so whoever reads `--coverage` judges the
// decision with its reason in front of them (#2qq4yr7). An off no file explains loads only outside our
// tiers, where it is the project's choice (#bfxz13m): a default run counts it, and `--coverage` names it.
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestAnOffPrintsItsReasonAndAnUnreasonedOffIsCountedByDefault(t *testing.T) {
	root := "/repository"
	rules := []rule.Rule{{Name: "prefer-arrow-callback"}, {Name: "no-continue"}}
	coverage := program.Coverage{
		RulesOffered:      map[string]int{},
		RulesListening:    map[string]int{},
		RulesReporting:    map[string]int{},
		RulesScopedOff:    map[string]int{"prefer-arrow-callback": 40, "no-continue": 40},
		RulesUnconfigured: map[string]int{},
	}
	config := &configuration.Config{
		Root: root,
		Rules: map[string]configuration.RuleSetting{
			"prefer-arrow-callback": {Severity: configuration.SeverityOff},
			"no-continue":           {Severity: configuration.SeverityOff},
		},
		OffReasons: map[string]configuration.OffReason{
			"prefer-arrow-callback": {File: root + "/CohereSettings.json", Reason: "we write function callbacks on purpose"},
		},
	}
	report := lintReport{Result: program.Result{Coverage: coverage}, Rules: rules, LintConfig: config, WalkCost: "in 1s"}

	output := renderLintReport(report)
	requireLines(t, output,
		"  off with no reason: 1 rule your settings turn off without saying why; cohere --coverage names them\n",
	)
	forbidLines(t, output, "prefer-arrow-callback", "no-continue")

	report.Details = true
	requireLines(t, renderLintReport(report),
		"    prefer-arrow-callback (40 files, off because we write function callbacks on purpose)\n",
		"    no-continue (40 files, off with no reason given)\n",
	)
}

// What rules counted rather than reported is listed under `--coverage`, summed across files and sorted
// by rule and key, so a `@processState` tag reads as its type and how many writes it exempted (#dz42gce).
// A default run counts and names nothing of it: it is a fact to read, not an action.
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestCoverageListsWhatRulesNoted(t *testing.T) {
	coverage := program.Coverage{
		RulesOffered:   map[string]int{"nexus/correctness-no-caller-data-mutation": 2},
		RulesListening: map[string]int{"nexus/correctness-no-caller-data-mutation": 2},
	}
	result := program.Result{
		Coverage: coverage,
		Notes: map[string]program.RuleNotes{
			"/project/Hub.ts":    {"nexus/correctness-no-caller-data-mutation": {"HubInterface in /project/Hub.ts": 3}},
			"/project/Helper.ts": {"nexus/correctness-no-caller-data-mutation": {"HubInterface in /project/Hub.ts": 2, "AnotherHubInterface in /project/Another.ts": 1}},
		},
	}
	rules := []rule.Rule{{Name: "nexus/correctness-no-caller-data-mutation"}}

	detailed := renderLintReport(lintReport{Result: result, Rules: rules, WalkCost: "in 1s", Details: true})
	requireLines(t, detailed,
		"  notes (6): what rules counted rather than reported\n"+
			"    nexus/correctness-no-caller-data-mutation\n"+
			"      AnotherHubInterface in /project/Another.ts: 1\n"+
			"      HubInterface in /project/Hub.ts: 5\n",
	)

	plain := renderLintReport(lintReport{Result: result, Rules: rules, WalkCost: "in 1s"})
	if strings.Contains(plain, "notes (") {
		t.Errorf("a default run listed notes:\n%s", plain)
	}
}

// A skip is listed apart from what rules noted, by rule and reason with its count of files, under
// `--coverage`, and named by default too, since a rule that declined every file still counts as having
// run and no other line would show it (#pa7k7zv). A rule's other notes stay under notes.
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestCoverageNamesWhatRulesSkipped(t *testing.T) {
	ruleName := "@typescript-eslint/no-useless-default-assignment"
	skip := rule.SkippedNotePrefix + "strictNullChecks is off"
	result := program.Result{
		Coverage: program.Coverage{
			RulesOffered:   map[string]int{ruleName: 3, "nexus/correctness-no-caller-data-mutation": 1},
			RulesListening: map[string]int{"nexus/correctness-no-caller-data-mutation": 1},
		},
		Notes: map[string]program.RuleNotes{
			"/project/a.ts": {ruleName: {skip: 1}},
			"/project/b.ts": {ruleName: {skip: 1}},
			"/project/c.ts": {ruleName: {skip: 1}, "nexus/correctness-no-caller-data-mutation": {"HubInterface in /project/Hub.ts": 2}},
		},
	}
	rules := []rule.Rule{{Name: ruleName}, {Name: "nexus/correctness-no-caller-data-mutation"}}

	detailed := renderLintReport(lintReport{Result: result, Rules: rules, WalkCost: "in 1s", Details: true})
	requireLines(t, detailed,
		"  skipped (1): rules that declined files on a compiler option or a missing precondition, so they checked nothing there\n"+
			"    "+ruleName+": skipped on 3 files, strictNullChecks is off\n"+
			"  notes (2): what rules counted rather than reported\n"+
			"    nexus/correctness-no-caller-data-mutation\n"+
			"      HubInterface in /project/Hub.ts: 2\n",
	)
	if strings.Contains(detailed, "      "+skip) {
		t.Errorf("a skip was listed among the notes as well:\n%s", detailed)
	}

	plain := renderLintReport(lintReport{Result: result, Rules: rules, WalkCost: "in 1s"})
	requireLines(t, plain, "  skipped: "+ruleName+": skipped on 3 files, strictNullChecks is off\n")

	// The skip is the reason it registered nothing, so it is not also named as a rule that checked
	// nothing for no reason (#yj96emr)
	forbidLines(t, plain, "no listener:")
	requireLines(t, detailed, "    "+ruleName+" (offered 3 files, skipped every file it was offered)\n")
	if strings.Contains(detailed, "declares no reason to register none") {
		t.Errorf("a rule that skipped every file it was offered reads as dead:\n%s", detailed)
	}
}

// A rule that skipped some of the files it was offered and registered nothing on the rest gave no reason
// for the rest, so it still prints as having checked nothing there
// Not parallel: it renders through writeLintReport, which sets the package-level activeSummary
func TestARuleThatSkippedSomeFilesStillNeedsActionForTheRest(t *testing.T) {
	ruleName := "@typescript-eslint/no-useless-default-assignment"
	result := program.Result{
		Coverage: program.Coverage{RulesOffered: map[string]int{ruleName: 3}},
		Notes: map[string]program.RuleNotes{
			"/project/a.ts": {ruleName: {rule.SkippedNotePrefix + "strictNullChecks is off": 1}},
		},
	}
	rules := []rule.Rule{{Name: ruleName}}

	plain := renderLintReport(lintReport{Result: result, Rules: rules, WalkCost: "in 1s"})
	requireLines(t, plain,
		"  skipped: "+ruleName+": skipped on 1 file, strictNullChecks is off\n",
		"  no listener: rule "+ruleName+" was offered 3 files and registered no listener on any",
	)
}
