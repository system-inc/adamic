package main

import (
	"fmt"
	"strings"
	"time"
)

// runFinding is one finding, a rule's or a type error, whose rule is then its TypeScript code (TS2322).
// Both views print the same fields: the human line `path:line:col severity rule message`, and `--json`.
type runFinding struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity string `json:"severity"`
	Rule     string `json:"rule"`
	// MessageID names which of a rule's messages this is. Empty for a type error.
	MessageID string `json:"messageId,omitempty"`
	Message   string `json:"message"`
}

// runSummary is everything a finished run can say about itself, gathered in one place so every way of
// printing it reads the same facts: the footer a default run ends with, the full account under
// `--verbose`, and whatever comes next. Rendering is kept out of it, and each renderer is thin.
//
// It holds what ran and what it cost, what was found and changed, how much was checked, what the cache
// supplied, and, kept apart because no renderer may drop it, every way the run fell short of checking
// everything.
type runSummary struct {
	// Label is the project's path in a repository with several. Empty for a project checked on its own.
	Label string

	// Total is the run's wall clock, the number a developer waited for.
	Total time.Duration
	// Graph is how long reading, parsing and binding the program took, zero when no graph was built.
	Graph time.Duration
	// Phases is each phase's outcome and time, as the pipeline recorded them.
	Phases []phaseRecord
	// Rules is the rules that ran.
	Rules int
	// Formatting is the part of the fix phase's time with a format in flight, zero when no formatter ran.
	// The fix phase's own time holds it; the footer shows 🪄 as the rest.
	Formatting time.Duration

	Cache cacheUse

	TypeErrors int
	Findings   int

	// Changed is each file the run rewrote, by fixing, formatting or both: the files it cohered, the
	// brand's verb for altering. WouldChange is the files a `--no-fix` run found a fix or formatting
	// would rewrite, which fail it.
	Changed     []changedFile
	WouldChange int

	// FilesInScope is every file the run was to account for. FilesChecked is the ones it examined fresh,
	// and FilesCached the ones the cache answered for, so the two add up to FilesInScope; countsAgree
	// holds them to it. The split is the point: a warm run that checked three files and took the rest
	// from the cache must not read like one that checked them all again.
	FilesInScope int
	FilesChecked int
	FilesCached  int
	// Nodes is the syntax nodes walked this run, so only in the checked files.
	Nodes int

	// Adamic is the run's readiness, nil for a run that did not lint or did not ask. See readiness.go.
	Adamic *readinessSummary

	// AdamicRequested is --adamic-readiness, so a run that asked and could not measure says why.
	AdamicRequested bool

	// Skips is each rule that declined every file it was offered. Whether one is a gap depends on its cover
	// having run, which uncoveredSkips decides once the phases are known, into Gaps.
	Skips []ruleSkip

	Gaps runGaps
}

// ruleSkip is a rule that declined every file it was offered: why, and the check that reports the same
// thing in its place, empty when none does.
type ruleSkip struct {
	Rule      string
	Reason    string
	CoveredBy string
}

// uncoveredSkips counts the skips that left something unchecked: every skip with no cover, and every
// covered skip whose cover did not run this run. The one cover today is the types phase, which ran only
// when it ran to its end and lint was not cut off behind it, since a bail is a types phase that stopped
// the run rather than one that stood in for a rule.
func (summary runSummary) uncoveredSkips() int {
	count := 0
	for _, skip := range summary.Skips {
		if skip.CoveredBy == "" || !summary.coverRan(skip.CoveredBy) {
			count++
		}
	}
	return count
}

func (summary runSummary) coverRan(cover string) bool {
	if cover != "types" {
		return false
	}
	typesRan, lintCutOff := false, false
	for _, record := range summary.Phases {
		switch {
		case record.Name == phaseTypes && record.Outcome == outcomeRan:
			typesRan = true
		case record.Name == phaseLint && record.Outcome == outcomeNotReached:
			lintCutOff = true
		}
	}
	return typesRan && !lintCutOff
}

// cohered is the files the run rewrote, which is what the 🪄 and 💅 lines above the footer list.
func (summary runSummary) cohered() int {
	return len(summary.Changed)
}

// countsAgree checks that checked and cached add up to every file in scope, so neither count can drift
// from the other unnoticed. The other invariant, that the cohered count is the number of 🪄 and 💅 lines,
// holds by construction, since both come from Changed, and a test pins it.
func (summary runSummary) countsAgree() error {
	if summary.FilesChecked+summary.FilesCached != summary.FilesInScope {
		return fmt.Errorf("%d checked and %d cached do not add up to the %d files in scope",
			summary.FilesChecked, summary.FilesCached, summary.FilesInScope)
	}
	return nil
}

// changedFile is one file the run rewrote.
type changedFile struct {
	// Path is relative to the project root, as a reader scanning the list would want it.
	Path      string
	Fixed     bool
	Formatted bool
	// FixedBy counts the fixes applied, by the rule that made them.
	FixedBy map[string]int
}

// filesFixed and filesFormatted count the changed files of each kind; a file can be both.
func (summary runSummary) filesFixed() int {
	count := 0
	for _, file := range summary.Changed {
		if file.Fixed {
			count++
		}
	}
	return count
}

func (summary runSummary) filesFormatted() int {
	count := 0
	for _, file := range summary.Changed {
		if file.Formatted {
			count++
		}
	}
	return count
}

// cacheUse is what the run took from the cache.
type cacheUse struct {
	// Replayed is a run answered whole from the run cache, which ran no phase.
	Replayed bool `json:"replayed"`
	// FilesReplayed is the files whose findings came from the cache: in a run that computed the rest, those it
	// did not, and in a whole replay, every file in scope.
	FilesReplayed int `json:"filesReplayed"`
	// Off is `--no-cache`: nothing read, nothing written.
	Off bool `json:"off"`

	// MissReason is why a run the cache could have answered whole was not, and FindingsMissReason why its
	// files' findings were not replayed: nothing on record, a key whose named part moved, or files with no
	// entry or with changed bytes. Empty when there is nothing to explain (#547dhjz).
	MissReason         string `json:"missReason,omitempty"`
	FindingsMissReason string `json:"findingsMissReason,omitempty"`
}

// missLine is the --verbose line that says why the cache did not answer, empty when there is nothing to say.
func (cache cacheUse) missLine() string {
	var parts []string
	if cache.MissReason != "" {
		parts = append(parts, "not replayed whole: "+cache.MissReason)
	}
	if cache.FindingsMissReason != "" {
		parts = append(parts, "findings: "+cache.FindingsMissReason)
	}
	if len(parts) == 0 {
		return ""
	}
	return "cache: " + strings.Join(parts, "; ")
}

// runGaps is every way a run fell short of checking everything. Each is said in the footer even when
// the run is green.
type runGaps struct {
	// ProgramFiles is the program's size when the run was narrowed to fewer of its files, zero otherwise.
	ProgramFiles int `json:"programFiles"`
	CrashedFiles int `json:"crashedFiles"`
	// RuleCrashes is each rule that crashed on a file, leaving its verdict there missing. Like a crashed
	// file it fails the run: see runCrash.
	RuleCrashes int `json:"ruleCrashes"`
	// RulesSkippingEverything is the rules that declined every file they were offered.
	RulesSkippingEverything int  `json:"rulesSkippingEverything"`
	FormattingNotChecked    bool `json:"formattingNotChecked"`
	NothingToCheck          bool `json:"nothingToCheck"`
	// ModifiedBuild is a binary built from a modified tree, which no commit reproduces.
	ModifiedBuild bool `json:"modifiedBuild"`
	// Unread is files the run could not check for any other reason, said as the run said it.
	Unread string `json:"unread,omitempty"`
	// AdamicIgnored is the Adamic `.a` files the program holds that git's ignore rules leave out of the
	// format walk: checked, and never formatted (#6mhafvb).
	AdamicIgnored int `json:"adamicIgnored,omitempty"`
}

// failed is whether the run found anything or lost a verdict to a crash, which is what the exit code says
// too. A crash is cohere's bug, and a run that could not check a file never exits 0. See runCrash.
func (summary runSummary) failed() bool {
	return summary.TypeErrors > 0 || summary.Findings > 0 || summary.WouldChange > 0 ||
		summary.Gaps.CrashedFiles > 0 || summary.Gaps.RuleCrashes > 0
}
