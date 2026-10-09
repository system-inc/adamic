package main

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// Coverage, said as counts a reader can check rather than as a wall.
//
// Every printer that used to live here exists because a silent gap once shipped: four rules in the
// gate this replaces were dead for months under a green run, a rule turned off across a directory read
// exactly like a rule with nothing to report, and forty withheld findings printed the same green as
// none. The doctrine stands: a run that checked nothing must never look like a run that found nothing.
//
// What changed is the shape. A clean run on ahra printed about 170 notes, a rule nobody configured
// printed twice ("offered no files" and "not in the config"), and a rule the config turned off printed
// twice ("offered no files" and "scoped off for 3785 files"). A reader skims 170 lines and learns
// nothing, which is the same failure as printing nothing: the line that matters drowns. So every rule
// lands in exactly one category, the categories are counted on one line whose terms add up to the rule
// total, and each rule is named once, under its category, behind `--coverage`. What needs action
// prints in full by default whatever the mode: findings, crashed and unreadable files, config keys that
// match no rule, reasonless suppressions, the parity line, and a rule that registered nothing when it
// should have. See `ahra tasks show psj5k5q`.

// coverageCategory is the one coverage fact that describes a rule. Every rule lands in exactly one,
// which is what lets the counts add up and what keeps a rule from being described twice.
type coverageCategory string

const (
	coverageFoundSomething     coverageCategory = "FoundSomething"
	coverageRanOnPartOfTheTree coverageCategory = "RanOnPartOfTheTree"
	coverageRanAndFoundNothing coverageCategory = "RanAndFoundNothing"
	coverageNoListener         coverageCategory = "NoListener"
	coverageListenedToNothing  coverageCategory = "ListenedToNothing"
	coverageOffByTheConfig     coverageCategory = "OffByTheConfig"
	coverageNotConfigured      coverageCategory = "NotConfigured"
	coverageOfferedNoFiles     coverageCategory = "OfferedNoFiles"
	coverageNotPorted          coverageCategory = "NotPorted"
)

// coverageCategoryOrder is the order the counts and the details print in: what ran first, then what
// did not and why.
var coverageCategoryOrder = []coverageCategory{
	coverageFoundSomething,
	coverageRanOnPartOfTheTree,
	coverageRanAndFoundNothing,
	coverageNoListener,
	coverageListenedToNothing,
	coverageOffByTheConfig,
	coverageNotConfigured,
	coverageOfferedNoFiles,
	coverageNotPorted,
}

// coverageCategoryTerms is each category as a term in the counted line.
var coverageCategoryTerms = map[coverageCategory]string{
	coverageFoundSomething:     "found something",
	coverageRanOnPartOfTheTree: "ran on part of the tree",
	coverageRanAndFoundNothing: "ran and found nothing",
	coverageNoListener:         "registered no listener",
	coverageListenedToNothing:  "listened to no files",
	coverageOffByTheConfig:     "off by the config",
	coverageNotConfigured:      "nobody has configured",
	coverageOfferedNoFiles:     "offered no files",
	coverageNotPorted:          "not ported",
}

// coverageCategoryMeanings is what each category means, said once at its heading under `--coverage`
// rather than once per rule. The per-rule sentence was the wall: 52 copies of "nobody has said whether
// it should" carry one fact.
var coverageCategoryMeanings = map[coverageCategory]string{
	coverageFoundSomething: "its findings are printed above",
	// A rule that looked at real code and had nothing to say is either a clean tree or a rule that
	// cannot see. Two false positives shipped past a full fixture pair and were caught only by running
	// against the tree, which is why this is counted on every run rather than left to a habit.
	coverageRanAndFoundNothing: "a clean tree and a rule that cannot see look identical here",
	// A rule on at the top level and turned off by an override for most files read as "ran and found
	// nothing" if any file still ran it. phi_api's twelve core rules did: its whole-tree override kept
	// them off for every .ts and .tsx file, they ran on the .cjs files alone, and coverage said they ran
	// (#7n4zxrb). So a quiet rule the config kept from some files is counted apart, with the files it ran
	// on beside the files it did not.
	coverageRanOnPartOfTheTree: "ran and found nothing on the files it ran on, and the config turned it off or never enabled it for the rest, so its silence covers only those files",
	// Three different rules look identical here by their numbers. A rule whose whole job is answered
	// from the file itself does its work in Run and returns no listener
	// (network-require-hook-request-suffix inspects every hook and reports before returning). A rule
	// that declines files it is not about registers nothing on a tree with none of them. And a rule its
	// options leave inert (init-declarations with no mode) registers nothing anywhere. The first two
	// are working and the third is dead, so each rule declares which it is (rule.NoListenerKind), and a
	// rule declaring neither prints by default as having checked nothing.
	coverageNoListener:        "offered files and registered no listener on any; each says whether it answers in Run, declines files it is not about, or declares neither and checked nothing",
	coverageListenedToNothing: "nothing gave it a file, or it declined every one, so its silence says nothing about the tree",
	coverageOffByTheConfig:    "the config turns it off for every file it would have run on",
	// "Someone turned this rule off" and "nobody has said whether this rule should run" are different
	// facts, and a rule newly added to the registry is a decision waiting to be made rather than one
	// already made. Collapsing them would describe a brand-new rule as though it had been excluded.
	coverageNotConfigured: "not in the config, so it ran on no files; nobody has said whether it should",
	// A rule offered nothing was never wired, and its silence says nothing about the tree at all. With
	// the config's own two reasons counted apart above, what lands here is a rule every file it could
	// have run on was ignored for.
	coverageOfferedNoFiles: "nothing wired it, so its silence says nothing about the tree",
	coverageNotPorted:      "the config asks for it and this binary does not implement it, so nothing here checked it",
}

// ruleCoverageEntry is one rule's coverage fact: its category, and what else is true of it.
type ruleCoverageEntry struct {
	Name     string
	Category coverageCategory
	// Details are the rule's numbers and the partial facts that do not change its category: a rule
	// that ran on most files and is off by the config for six is a rule that ran, with "off by the
	// config for 6 files" beside its name rather than a second line in a second category.
	Details []string
	// NeedsAction marks a rule that was offered files, registered nothing, and declares no reason it may
	// (rule.NoListenerKind). That is the dead-rule signal these notes were invented for, so it prints in
	// full by default.
	NeedsAction bool
	// OffWithoutReason marks a rule the config turns off at top level with no file saying why. Inside our
	// tiers the loader refuses one, so this is a project outside them going its own way (#bfxz13m): it is
	// counted by default and named under `--coverage`, so the choice stays visible.
	OffWithoutReason bool
}

// coverageCrash is one file a rule could not finish, in either engine's terms.
type coverageCrash struct {
	File  string
	Cause string
}

// coverageRuleCrash is one rule that panicked on one file while the file's other rules finished.
type coverageRuleCrash struct {
	Rule  string
	File  string
	Cause string
}

// suppressionCoverage is what disable comments withheld. Nil when the engine reports none.
type suppressionCoverage struct {
	Suppressed              int
	SuppressedWithoutReason int
	// Dead is the directives that silenced nothing while their rule ran; ForUnrunRules is the ones
	// naming only rules this run did not run. They ask for opposite actions: dead scaffolding is worth
	// deleting, and a directive for a rule cohere has not ported yet is load-bearing today. Measured at
	// 63 of 171 rules ported, 82 of 100 unused directives were the second kind, so one number told a
	// reader to delete comments the gate still needed.
	Dead          int
	ForUnrunRules int

	// DeadSites is where each dead directive sits, listed under `--coverage`, sorted by file and line.
	// Nil for an engine that counts them without saying where.
	DeadSites []program.DeadSuppression
}

// coverageSummary is everything the coverage block says, built by each engine from its own records
// and printed by one writer, so a Swift run and a TypeScript run say the same thing about the same
// fact.
type coverageSummary struct {
	Entries []ruleCoverageEntry
	// Unnamed counts rules a category holds that the engine's record does not name: the Swift record
	// counts its quiet rules and names none. They count toward the total and print as a count.
	Unnamed map[coverageCategory]int
	// PartialFacts are facts about rules the record counts and does not name, said once by name under
	// `--coverage`: a Swift rule counted as quiet that the record also names as off for some files.
	PartialFacts []string
	Crashes      []coverageCrash
	RuleCrashes  []coverageRuleCrash
	FilesIgnored int
	Suppression  *suppressionCoverage
	// Notes is what rules counted rather than reported (rule.Context.Note), summed across files: rule
	// name to note key to count. Listed under `--coverage`, so an exemption reads as a number.
	Notes map[string]map[string]int
}

// count is how many rules a category holds, named or not.
func (s coverageSummary) count(category coverageCategory) int {
	total := s.Unnamed[category]
	for _, entry := range s.Entries {
		if entry.Category == category {
			total++
		}
	}
	return total
}

// total is every rule the summary accounts for. It is the sum of the categories by construction, and
// a test pins that the counted line says so.
func (s coverageSummary) total() int {
	total := len(s.Entries)
	for _, count := range s.Unnamed {
		total += count
	}
	return total
}

// noListenerReasons is how a declared reason reads beside a rule's name under `--coverage`.
var noListenerReasons = map[rule.NoListenerKind]string{
	rule.NoListenerAnswersInRun:            "answers in Run",
	rule.NoListenerDeclinesIrrelevantFiles: "declines files it is not about",
}

// classifyTypeScriptCoverage puts every registered rule, and every rule the config asks for that this
// binary lacks, into exactly one category.
//
// The order of the cases is the precedence, and it is what keeps one rule out of two categories. A
// rule that reported anything found something, whatever else is true of it. A rule that listened ran.
// A rule that was offered files and listened to none registered no listener. Only a rule offered
// nothing is described by why it was offered nothing, and the config's two reasons are told apart
// there: a rule turned off and a rule nobody configured are different decisions.
func classifyTypeScriptCoverage(
	rules []rule.Rule,
	coverage program.Coverage,
	lintConfig *configuration.Config,
) coverageSummary {
	summary := coverageSummary{
		FilesIgnored: coverage.FilesIgnored,
		Suppression: &suppressionCoverage{
			Suppressed:              coverage.Suppressed,
			SuppressedWithoutReason: coverage.SuppressedWithoutReason,
			Dead:                    coverage.UnusedSuppressions - coverage.UnusedSuppressionsForUnrunRules,
			ForUnrunRules:           coverage.UnusedSuppressionsForUnrunRules,
			DeadSites:               coverage.DeadSuppressions,
		},
	}

	for _, subject := range rules {
		name := subject.Name
		offered := coverage.RulesOffered[name]
		scopedOff := coverage.RulesScopedOff[name]
		unconfigured := coverage.RulesUnconfigured[name]
		entry := ruleCoverageEntry{Name: name}

		switch {
		case coverage.RulesReporting[name] > 0:
			entry.Category = coverageFoundSomething
			entry.Details = append(entry.Details, fmt.Sprintf("%d findings", coverage.RulesReporting[name]))
		case coverage.RulesListening[name] > 0 && scopedOff+unconfigured > 0:
			entry.Category = coverageRanOnPartOfTheTree
			entry.Details = append(entry.Details, partOfTheTreeDetail(coverage.RulesListening[name], offered, scopedOff, unconfigured))
		case coverage.RulesListening[name] > 0:
			entry.Category = coverageRanAndFoundNothing
		case offered > 0:
			entry.Category = coverageNoListener
			entry.Details = append(entry.Details, fmt.Sprintf("offered %d files", offered))
			if reason, declared := noListenerReasons[subject.NoListener]; declared {
				entry.Details = append(entry.Details, reason)
			} else {
				entry.NeedsAction = true
			}
		case scopedOff > 0:
			entry.Category = coverageOffByTheConfig
			entry.Details = append(entry.Details, fmt.Sprintf("%d files", scopedOff))
		case unconfigured > 0:
			entry.Category = coverageNotConfigured
			entry.Details = append(entry.Details, fmt.Sprintf("%d files", unconfigured))
		default:
			entry.Category = coverageOfferedNoFiles
		}

		// The partial facts. A rule enabled only inside an override is unconfigured for every file the
		// override does not reach, and it once printed "ran on no files" while reporting thousands of
		// findings in the files the override does reach. Found on the boundaries dry run by
		// @system_cohere_base_rules. A rule that ran keeps its category and says where it did not run.
		if offered > 0 && entry.Category != coverageRanOnPartOfTheTree {
			if scopedOff > 0 {
				entry.Details = append(entry.Details, fmt.Sprintf("off by the config for %d files", scopedOff))
			}
			if unconfigured > 0 {
				entry.Details = append(entry.Details, fmt.Sprintf("not in the config for %d files", unconfigured))
			}
		} else if entry.Category == coverageOffByTheConfig && unconfigured > 0 {
			entry.Details = append(entry.Details, fmt.Sprintf("not in the config for %d more", unconfigured))
		}

		// Why it is off, from the file that said so, so a reader judging an off reads the decision
		// beside it rather than in a file the run never names.
		if entry.Category == coverageOffByTheConfig {
			if offReason, reasoned := lintConfig.OffReasonFor(name); reasoned {
				entry.Details = append(entry.Details, "off because "+offReason.Reason)
			} else if lintConfig.TurnsOffAtTopLevel(name) {
				entry.OffWithoutReason = true
			}
		}

		summary.Entries = append(summary.Entries, entry)
	}

	for _, name := range unportedConfiguredRules(rules, lintConfig) {
		summary.Entries = append(summary.Entries, ruleCoverageEntry{Name: name, Category: coverageNotPorted})
	}

	// A crashed file produced no findings, and a run that lost a file to a panic must not read as a run
	// that found nothing in it. Named rather than counted, because the panic message is the defect.
	for _, crash := range coverage.FilesCrashed {
		summary.Crashes = append(summary.Crashes, coverageCrash{File: crash.FileName, Cause: fmt.Sprint(crash.Cause)})
	}
	for _, crash := range coverage.RulesCrashed {
		summary.RuleCrashes = append(summary.RuleCrashes, coverageRuleCrash{Rule: crash.RuleName, File: crash.FileName, Cause: fmt.Sprint(crash.Cause)})
	}
	sort.Slice(summary.RuleCrashes, func(first, second int) bool {
		if summary.RuleCrashes[first].Rule != summary.RuleCrashes[second].Rule {
			return summary.RuleCrashes[first].Rule < summary.RuleCrashes[second].Rule
		}
		return summary.RuleCrashes[first].File < summary.RuleCrashes[second].File
	})

	summary.sortEntries()
	return summary
}

// partOfTheTreeDetail says where a rule ran and where it did not, as terms that add up to every file it
// could have run on: `ran on 12 of 3242 files, off by the config on 3230`. The files it was offered and
// declined are a term of their own when there are any, so the sum still holds.
func partOfTheTreeDetail(listening, offered, scopedOff, unconfigured int) string {
	terms := []string{fmt.Sprintf("ran on %d of %d files", listening, offered+scopedOff+unconfigured)}
	if declined := offered - listening; declined > 0 {
		terms = append(terms, fmt.Sprintf("declined %d", declined))
	}
	if scopedOff > 0 {
		terms = append(terms, fmt.Sprintf("off by the config on %d", scopedOff))
	}
	if unconfigured > 0 {
		terms = append(terms, fmt.Sprintf("not in the config for %d", unconfigured))
	}
	return strings.Join(terms, ", ")
}

// sortEntries orders the entries by category, then by name, so the details read in the counted
// line's order and two runs can be diffed.
func (s *coverageSummary) sortEntries() {
	position := make(map[coverageCategory]int, len(coverageCategoryOrder))
	for index, category := range coverageCategoryOrder {
		position[category] = index
	}
	sort.SliceStable(s.Entries, func(first, second int) bool {
		if s.Entries[first].Category != s.Entries[second].Category {
			return position[s.Entries[first].Category] < position[s.Entries[second].Category]
		}
		return s.Entries[first].Name < s.Entries[second].Name
	})
}

// coverageCountedLine is the line whose terms add up to the rule total, said as arithmetic so a reader
// can check it: `coverage: 459 rules = 370 ran and found nothing + 14 registered no listener + ...`.
// A category holding no rule is left out, and the sum still holds.
func (s coverageSummary) coverageCountedLine() string {
	terms := []string{}
	for _, category := range coverageCategoryOrder {
		if count := s.count(category); count > 0 {
			terms = append(terms, fmt.Sprintf("%d %s", count, coverageCategoryTerms[category]))
		}
	}
	line := fmt.Sprintf("coverage: %d rules", s.total())
	if len(terms) > 0 {
		line += " = " + strings.Join(terms, " + ")
	}
	return line
}

// coverageFilesLine is the second line: what happened to files and to findings rather than to rules,
// so none of it is part of the rule arithmetic.
//
// The crash count prints when it is zero, because "0 files crashed" is the claim a clean run makes and
// a reader should be able to see it made. The reasonless count prints every run rather than being
// enforced: measured across the codebase cohere gates, 281 of 306 suppressions of our own rules stated
// no reason, so requiring one would turn working code red for no defect, and printing it is what lets
// the convention be tightened later from evidence rather than from a guess.
func (s coverageSummary) coverageFilesLine(details bool) string {
	terms := []string{fmt.Sprintf("%d files crashed", len(s.Crashes))}
	if len(s.RuleCrashes) > 0 {
		terms = append(terms, fmt.Sprintf("%d rule crashes, each costing one rule one file", len(s.RuleCrashes)))
	}
	if s.FilesIgnored > 0 {
		terms = append(terms, fmt.Sprintf("%d files excluded by ignorePatterns", s.FilesIgnored))
	}
	if s.Suppression != nil {
		terms = append(terms, fmt.Sprintf("%d findings suppressed, %d without a reason",
			s.Suppression.Suppressed, s.Suppression.SuppressedWithoutReason))
		// Dead scaffolding is worth deleting, so it is counted by default when there is any. It is a
		// count rather than a finding because a filtered run makes every directive for an unselected
		// rule look unused, and a number that is wrong under a common flag should not fail a build.
		if s.Suppression.Dead > 0 {
			terms = append(terms, fmt.Sprintf("%d disable comments silenced nothing while their rule ran", s.Suppression.Dead))
		}
	}
	if !details {
		terms = append(terms, "details: cohere --coverage")
	}
	return "  " + strings.Join(terms, " · ")
}

// writeCoverage prints the coverage block: the counted line, the files line, then everything that needs
// action in full, and with details the per-rule listing.
//
// Each per-rule fact prints exactly once in either mode. A rule that needs action prints as its own
// line by default and under its category with details, never both.
func writeCoverage(out io.Writer, summary coverageSummary, details bool) {
	fmt.Fprintln(out, summary.coverageCountedLine())
	fmt.Fprintln(out, summary.coverageFilesLine(details))
	writeCoverageNotes(out, summary, details)
}

// writeCoverageDetails names every rule once, under the one category that describes it, with the
// category's meaning said once at its heading.
func writeCoverageDetails(out io.Writer, summary coverageSummary) {
	fmt.Fprintln(out, "  coverage details, each rule once:")
	for _, category := range coverageCategoryOrder {
		count := summary.count(category)
		if count == 0 {
			continue
		}
		fmt.Fprintf(out, "  %s (%d): %s\n", coverageCategoryTerms[category], count, coverageCategoryMeanings[category])
		if unnamed := summary.Unnamed[category]; unnamed > 0 {
			fmt.Fprintf(out, "    %d the engine's record counts and does not name\n", unnamed)
		}
		for _, entry := range summary.Entries {
			if entry.Category != category {
				continue
			}
			line := "    " + entry.Name
			details := entry.Details
			if entry.NeedsAction {
				details = append(append([]string(nil), details...), "declares no reason to register none, so it checked nothing")
			}
			if entry.OffWithoutReason {
				details = append(append([]string(nil), details...), "off with no reason given")
			}
			if len(details) > 0 {
				line += " (" + strings.Join(details, ", ") + ")"
			}
			fmt.Fprintln(out, line)
		}
	}
	for _, fact := range summary.PartialFacts {
		fmt.Fprintf(out, "  partly: %s\n", fact)
	}
	if summary.Suppression != nil && summary.Suppression.ForUnrunRules > 0 {
		fmt.Fprintf(out, "  suppressions: %d unused disable comments name only rules cohere has not ported yet, so nothing looked and they are not dead\n",
			summary.Suppression.ForUnrunRules)
	}
	if summary.Suppression != nil && len(summary.Suppression.DeadSites) > 0 {
		writeDeadSuppressions(out, summary.Suppression.DeadSites)
	}
	skips, notes := splitRuleSkips(summary.Notes)
	if len(skips) > 0 {
		writeRuleSkips(out, skips)
	}
	if len(notes) > 0 {
		writeRuleNotes(out, notes)
	}
}

// splitRuleSkips separates the notes rule.Context.Skip recorded, keyed by reason without the prefix, from
// every other note, so a skip is listed as a skip rather than as one more fact a rule counted.
func splitRuleSkips(notes map[string]map[string]int) (skips, others map[string]map[string]int) {
	skips = map[string]map[string]int{}
	others = map[string]map[string]int{}
	for ruleName, counts := range notes {
		for key, count := range counts {
			target, label := others, key
			if reason, isSkip := strings.CutPrefix(key, rule.SkippedNotePrefix); isSkip {
				target, label = skips, reason
			} else if rest, isCovered := strings.CutPrefix(key, rule.CoveredSkipNotePrefix); isCovered {
				cover, reason, _ := strings.Cut(rest, ": ")
				target, label = skips, reason+" (covered by "+cover+")"
			}
			if target[ruleName] == nil {
				target[ruleName] = map[string]int{}
			}
			target[ruleName][label] += count
		}
	}
	return skips, others
}

// rulesSkippingEveryFile is each rule that declined every file it was offered, with its reason and, for a
// covered skip, the check that covers it. A rule that declined only some files still checked the rest, so
// it is not here. Offered is the same measure excuseRulesThatSkippedEveryFile uses.
func rulesSkippingEveryFile(notes map[string]map[string]int, offered map[string]int) []ruleSkip {
	var skipping []ruleSkip
	for ruleName, counts := range notes {
		for key, count := range counts {
			if offered[ruleName] == 0 || count < offered[ruleName] {
				continue
			}
			if reason, isSkip := strings.CutPrefix(key, rule.SkippedNotePrefix); isSkip {
				skipping = append(skipping, ruleSkip{Rule: ruleName, Reason: reason})
			} else if rest, isCovered := strings.CutPrefix(key, rule.CoveredSkipNotePrefix); isCovered {
				cover, reason, _ := strings.Cut(rest, ": ")
				skipping = append(skipping, ruleSkip{Rule: ruleName, Reason: reason, CoveredBy: cover})
			}
		}
	}
	sort.Slice(skipping, func(left, right int) bool { return skipping[left].Rule < skipping[right].Rule })
	return skipping
}

// skipLines renders each rule's skips as `rule: skipped on N files, reason`, sorted by rule and reason. A
// skip is noted once per file the rule declined, so the count is a count of files.
func skipLines(skips map[string]map[string]int) []string {
	var lines []string
	for ruleName, reasons := range skips {
		for reason, count := range reasons {
			files := "files"
			if count == 1 {
				files = "file"
			}
			lines = append(lines, fmt.Sprintf("%s: skipped on %d %s, %s", ruleName, count, files, reason))
		}
	}
	sort.Strings(lines)
	return lines
}

// writeRuleSkips lists the rules that declined files because a compiler option or a missing precondition
// left them unable to judge, each with how many files and why. A rule that skips reads exactly like one
// that found nothing unless it is named here (#pa7k7zv).
func writeRuleSkips(out io.Writer, skips map[string]map[string]int) {
	lines := skipLines(skips)
	fmt.Fprintf(out, "  skipped (%d): rules that declined files on a compiler option or a missing precondition, so they checked nothing there\n", len(lines))
	for _, line := range lines {
		fmt.Fprintf(out, "    %s\n", line)
	}
}

// sumRuleNotes adds each file's notes together, by rule and key.
func sumRuleNotes(notes map[string]program.RuleNotes) map[string]map[string]int {
	summed := map[string]map[string]int{}
	for _, fileNotes := range notes {
		for ruleName, counts := range fileNotes {
			if summed[ruleName] == nil {
				summed[ruleName] = map[string]int{}
			}
			for key, count := range counts {
				summed[ruleName][key] += count
			}
		}
	}
	return summed
}

// excuseRulesThatSkippedEveryFile clears the dead-rule signal from a rule whose skips cover every file it
// was offered. A skip is a reason given at run time, as NoListener is one given at registration, and the
// skipped line already names it, so the same rule must not also read as having checked nothing for no
// reason: nexus/correctness-no-implicit-return stands down in every project that sets noImplicitReturns,
// and printed both (#yj96emr). A rule that skipped some files and registered nothing on the rest still
// needs action, since nothing explains the rest.
func (summary *coverageSummary) excuseRulesThatSkippedEveryFile(coverage program.Coverage) {
	skips, _ := splitRuleSkips(summary.Notes)
	for index := range summary.Entries {
		if !summary.Entries[index].NeedsAction {
			continue
		}
		skipped := 0
		for _, count := range skips[summary.Entries[index].Name] {
			skipped += count
		}
		if skipped >= coverage.RulesOffered[summary.Entries[index].Name] {
			summary.Entries[index].NeedsAction = false
			summary.Entries[index].Details = append(summary.Entries[index].Details, "skipped every file it was offered")
		}
	}
}

// writeRuleNotes lists what each rule counted rather than reported, by rule and then by key, each with
// its count over the run. A `@processState` tag lands here as the type and how many writes it exempted,
// so a tag on a widely used type shows up as a number someone reads rather than as a silence.
func writeRuleNotes(out io.Writer, notes map[string]map[string]int) {
	ruleNames := make([]string, 0, len(notes))
	total := 0
	for ruleName, counts := range notes {
		ruleNames = append(ruleNames, ruleName)
		for _, count := range counts {
			total += count
		}
	}
	sort.Strings(ruleNames)
	fmt.Fprintf(out, "  notes (%d): what rules counted rather than reported\n", total)
	for _, ruleName := range ruleNames {
		fmt.Fprintf(out, "    %s\n", ruleName)
		keys := make([]string, 0, len(notes[ruleName]))
		for key := range notes[ruleName] {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(out, "      %s: %d\n", key, notes[ruleName][key])
		}
	}
}

// writeDeadSuppressions names each disable comment that silenced nothing while a rule it names ran,
// one per line as file:line and the rules it names, so the count on the files line can be acted on
// from a whole-program run rather than hunted for one file at a time.
func writeDeadSuppressions(out io.Writer, sites []program.DeadSuppression) {
	sorted := append([]program.DeadSuppression(nil), sites...)
	sort.Slice(sorted, func(left, right int) bool {
		if sorted[left].File != sorted[right].File {
			return sorted[left].File < sorted[right].File
		}
		return sorted[left].Line < sorted[right].Line
	})
	fmt.Fprintf(out, "  dead disable comments (%d): each silenced nothing while a rule it names ran\n", len(sorted))
	for _, site := range sorted {
		named := "every rule"
		if len(site.Rules) > 0 {
			named = strings.Join(site.Rules, ", ")
		}
		// The directive's line is zero-based; a reader's editor counts from one.
		fmt.Fprintf(out, "    %s:%d %s\n", site.File, site.Line+1, named)
	}
}

// writeCrashedNote names one file a rule could not finish. Printed in full in both modes and before the
// details, because a file nothing checked needs action.
func writeCrashedNote(out io.Writer, fileName string, cause any) {
	fmt.Fprintf(out, "  crashed: %s: %s\n", fileName, runCrash{Path: fileName, Cause: fmt.Sprint(cause)}.message())
}

// namedGapsSentence is what "this run did not check everything" says when every phase ran and a file
// still went unchecked, named in the notes above it. One sentence for both engines.
const namedGapsSentence = "the notes above name the files nothing checked"

// namedRuleGapsSentence is the same fact when what went unchecked is one rule's verdict on a file rather
// than the whole file: the file's other rules ran, and saying nothing checked it would be its own lie.
const namedRuleGapsSentence = "the notes above name the rules that could not finish a file"

// unportedConfiguredRules names the rules the config turns on that this binary does not implement.
//
// Names are compared on the `/` boundary the config resolver uses: plain suffix matching would let a
// config entry for `no-enum` claim `consistency-no-enum`, and the count would read better than the
// truth. A rule turned off ran over no files regardless, so it is not something this run failed to
// check.
//
// **The rules block is not the whole config.** Forty rules are enforced by the `plugins` declarations
// and named in no rules block; they are held in `configuration.PluginDefaultRules` and resolved into
// the config's rules on load, so a denominator taken from the block alone would understate by exactly
// the rules nobody wrote down.
func unportedConfiguredRules(rules []rule.Rule, lintConfig *configuration.Config) []string {
	if lintConfig == nil {
		return nil
	}
	implemented := make(map[string]bool, len(rules))
	for _, registered := range rules {
		implemented[registered.Name] = true
	}
	missing := []string{}
	for name, setting := range lintConfig.Rules {
		if setting.Severity == configuration.SeverityOff {
			continue
		}
		if !implementsConfiguredRule(name, implemented) {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}

// wantedConfiguredRules counts the rules the config turns on, the parity line's denominator.
func wantedConfiguredRules(lintConfig *configuration.Config) int {
	wanted := 0
	for _, setting := range lintConfig.Rules {
		if setting.Severity != configuration.SeverityOff {
			wanted++
		}
	}
	return wanted
}

// writeParityCoverage says how many of the rules the config asks for this binary can actually run.
//
// The lint line reports how many rules ran, which is what this binary contains. It says nothing about
// how many were wanted, and while a port is in progress those are different numbers. A reader seeing a
// rule count has no way to learn whether the config asked for more, and the whole argument of this
// tool is that a run which checked less than it appears to must say so. Printed in full by default.
//
// The gap this closes was once wide enough to quote, and quoting it is what made the comment wrong
// within weeks: the figures drift with every rule that lands, while the reason they matter does not.
// The live numbers belong in the line this prints, which is derived, and not in a comment. This is
// the same omission the differential harness carried until `7b590f6`.
func writeParityCoverage(out io.Writer, rules []rule.Rule, lintConfig *configuration.Config) {
	if lintConfig == nil {
		return
	}
	missing := len(unportedConfiguredRules(rules, lintConfig))
	if missing == 0 {
		return
	}
	wanted := wantedConfiguredRules(lintConfig)
	fmt.Fprintf(out, "  parity: %d of %d rules the config asks for, so %d were not checked by anything here\n",
		wanted-missing, wanted, missing)
}

// lintReport is what the lint phase prints, gathered so the whole report can be written to any writer
// and every category a reader must act on can be tested to print by default.
type lintReport struct {
	Result     program.Result
	Rules      []rule.Rule
	LintConfig *configuration.Config
	// WalkCost says how the walk was paid for: `in 5.64s`, or which phase walked it.
	WalkCost string
	// Details is `--coverage`: name every rule once rather than only counting them.
	Details bool
}

// writeLintReport prints the findings, the lint line, and the coverage block.
//
// Coverage prints unconditionally, alongside the verdict rather than behind a flag. A run that checked
// nothing must not be able to look like a run that found nothing, and the only way to guarantee that is
// to make the population as visible as the findings. `--coverage` widens what is named; it never
// decides whether the counts print.
func writeLintReport(out io.Writer, report lintReport) {
	for _, diagnostic := range report.Result.Diagnostics {
		printRuleDiagnostic(out, diagnostic, report.LintConfig)
	}
	coverage := report.Result.Coverage
	for _, crash := range crashesFrom(coverage) {
		printCrash(out, crash)
	}
	// The account below is `--verbose`'s; the footer carries what a reader needs of it in every view.
	account := accountOutput(out)
	fmt.Fprintf(accountOutput(invocationOutput(out)),
		"lint: %d findings — %d rules over %d files, %d nodes visited%s, %s%s\n",
		len(report.Result.Diagnostics), coverage.RulesRun, coverage.FilesWalked, coverage.NodesVisited, nodesReplayedClause(coverage),
		report.WalkCost,
		replayedFromCache(report.Result),
	)

	summary := classifyTypeScriptCoverage(report.Rules, coverage, report.LintConfig)
	summary.Notes = sumRuleNotes(report.Result.Notes)
	summary.excuseRulesThatSkippedEveryFile(coverage)
	fmt.Fprintln(account, summary.coverageCountedLine())
	fmt.Fprintln(account, summary.coverageFilesLine(report.Details))
	writeParityCoverage(account, report.Rules, report.LintConfig)
	writeOrphanedConfigKeys(account, report.Rules, report.LintConfig)
	writeDepartures(account, report.LintConfig)
	writeOverrideReasons(account, report.LintConfig)
	writeCoverageNotes(account, summary, report.Details)
	writeUnrunRuleReferences(account, coverage.UnrunRuleReferences)

	// What the footer says of lint: the rules that ran, the files it checked fresh against those the cache
	// answered for, the nodes it walked, and the two gaps a green run must still name.
	activeSummary.Rules = coverage.RulesRun
	activeSummary.FilesCached = report.Result.FilesReplayed
	// The --json summary's cache.filesReplayed said 0 on every run, since nothing set it, and read as a full
	// check on runs that replayed 3,891 files (#547dhjz).
	activeSummary.Cache.FilesReplayed = report.Result.FilesReplayed
	activeSummary.FilesChecked = coverage.FilesWalked - report.Result.FilesReplayed
	activeSummary.FilesInScope = coverage.FilesWalked
	activeSummary.Nodes = coverage.NodesVisited
	activeSummary.Findings += len(report.Result.Diagnostics)
	activeSummary.Gaps.CrashedFiles = len(summary.Crashes)
	activeSummary.Gaps.RuleCrashes = len(summary.RuleCrashes)
	activeSummary.Skips = rulesSkippingEveryFile(summary.Notes, coverage.RulesOffered)
}

// writeUnrunRuleReferences notes the rules disable comments named that cohere doesn't run, under
// --verbose. They are real rules, a core rule cohere hasn't ported or another plugin's, so a directive
// naming one is a note rather than a finding (#v1ah2qq): it silences nothing here, and a repository
// written for ESLint carries many. Most named first, each with how many comments named it.
func writeUnrunRuleReferences(out io.Writer, references map[string]int) {
	if len(references) == 0 {
		return
	}
	names := make([]string, 0, len(references))
	total := 0
	for name, count := range references {
		names = append(names, name)
		total += count
	}
	sort.Slice(names, func(first, second int) bool {
		if references[names[first]] != references[names[second]] {
			return references[names[first]] > references[names[second]]
		}
		return names[first] < names[second]
	})
	named := make([]string, 0, len(names))
	for _, name := range names {
		named = append(named, fmt.Sprintf("%s %d", name, references[name]))
	}
	fmt.Fprintf(out, "  directives: %s cohere doesn't run, which silences nothing here: %s\n",
		counted(total, "disable comment names a rule", "disable comments name rules"), strings.Join(named, ", "))
}

// writeDepartures names every rule a config sets differently from a file it extends, with the reason it
// gave, on every run and by default.
//
// A departure is a project choosing to differ from a house ruling (#rkm5a31). Loading already refuses
// one with no reason; printing it is what stops a reasoned one becoming a quiet allowance, because a
// reader of any run can see each place the project and the house disagree and judge whether the reason
// still holds.
func writeDepartures(out io.Writer, lintConfig *configuration.Config) {
	if lintConfig == nil || len(lintConfig.Departures) == 0 {
		return
	}
	names := make([]string, 0, len(lintConfig.Departures))
	for name := range lintConfig.Departures {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		departure := lintConfig.Departures[name]
		file := departure.File
		if relative, err := filepath.Rel(lintConfig.Root, departure.File); err == nil {
			file = relative
		}
		fmt.Fprintf(out, "  departure: %s is set differently from the house ruling in %s: %s\n", name, file, departure.Reason)
	}
}

// writeOverrideReasons names every override that gave a reason, with the rules it sets and the reason,
// on every run and by default. A scoped override needs no reason, so one that gives a reason is usually
// standing in for unfinished work, and printing it keeps that work in front of whoever reads a run until
// the override can go.
func writeOverrideReasons(out io.Writer, lintConfig *configuration.Config) {
	if lintConfig == nil {
		return
	}
	for _, override := range lintConfig.Overrides {
		if override.Reason == "" {
			continue
		}
		file := override.File
		if relative, err := filepath.Rel(lintConfig.Root, override.File); err == nil {
			file = relative
		}
		names := make([]string, 0, len(override.Rules))
		for name, setting := range override.Rules {
			names = append(names, name+" "+setting.Severity.String())
		}
		sort.Strings(names)
		fmt.Fprintf(out, "  override: %s in %s sets %s: %s\n",
			strings.Join(override.Files, ", "), file, strings.Join(names, ", "), override.Reason)
	}
}

// offWithoutReasonCount is how many rules the settings turn off with no reason given.
func (s coverageSummary) offWithoutReasonCount() int {
	count := 0
	for _, entry := range s.Entries {
		if entry.OffWithoutReason {
			count++
		}
	}
	return count
}

// writeCoverageNotes is the part of the coverage block after the two counted lines: crashed files in
// full, then the rules that need action by default or every rule once with details.
func writeCoverageNotes(out io.Writer, summary coverageSummary, details bool) {
	for _, crash := range summary.Crashes {
		writeCrashedNote(out, crash.File, crash.Cause)
	}
	for _, crash := range summary.RuleCrashes {
		fmt.Fprintf(out, "  crashed: %s: %s\n", crash.File, runCrash{Path: crash.File, Rule: crash.Rule, Cause: crash.Cause}.message())
	}
	if details {
		writeCoverageDetails(out, summary)
		return
	}
	// A skip is named by default, not only under --coverage: a rule that declined every file is the case
	// no other line shows, since it still counts as having run (#pa7k7zv).
	skips, _ := splitRuleSkips(summary.Notes)
	for _, line := range skipLines(skips) {
		fmt.Fprintf(out, "  skipped: %s\n", line)
	}
	if offWithoutReason := summary.offWithoutReasonCount(); offWithoutReason > 0 {
		fmt.Fprintf(out, "  off with no reason: %d rule%s your settings turn off without saying why; cohere --coverage names them\n", offWithoutReason, plural(offWithoutReason))
	}
	for _, entry := range summary.Entries {
		if !entry.NeedsAction {
			continue
		}
		fmt.Fprintf(out, "  no listener: rule %s was %s and registered no listener on any, and it declares no reason it may (answering in Run, or declining files it is not about), so it checked nothing\n",
			entry.Name, strings.Join(entry.Details, ", "))
	}
}

// nodesReplayedClause is the lint line's account of the nodes the cache answered for, which this run did not
// walk: ", 2,360,633 more covered by replay". Empty when the walk covered every node itself.
func nodesReplayedClause(coverage program.Coverage) string {
	if coverage.NodesReplayed == 0 {
		return ""
	}
	return fmt.Sprintf(", %d more covered by replay", coverage.NodesReplayed)
}

// replayedFromCache says how much of a verdict was remembered rather than walked, so a reader can always
// tell the two apart, and how many replayed files ran type-aware rules again because something they import
// changed: content-keyed rules on every importer of a changed file, shape-keyed ones only where a shape
// changed. The design-system rules are counted apart, since what re-runs them is a stylesheet rather than an
// import, and so are the derived ones, re-run by the program-wide data their fingerprint covers. Empty when
// nothing was replayed.
func replayedFromCache(result program.Result) string {
	if result.FilesReplayed == 0 {
		return ""
	}
	replayed := fmt.Sprintf("; %d of %d files replayed from cache", result.FilesReplayed, result.Coverage.FilesWalked)
	if result.TypeAwareRerun > 0 || result.ShapeKeyedRerun > 0 {
		replayed += fmt.Sprintf(" (type-aware rules ran again on %d of them, shape-keyed on %d)", result.TypeAwareRerun, result.ShapeKeyedRerun)
	}
	if result.DesignSystemRerun > 0 {
		replayed += fmt.Sprintf(" (design-system rules ran again on %d of them)", result.DesignSystemRerun)
	}
	if result.DerivedRerun > 0 {
		replayed += fmt.Sprintf(" (derived rules ran again on %d of them)", result.DerivedRerun)
	}
	return replayed
}
