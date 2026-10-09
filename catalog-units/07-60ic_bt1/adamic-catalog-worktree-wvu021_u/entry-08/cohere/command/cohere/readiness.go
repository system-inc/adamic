package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/types/program"
)

/*
 * Adamic readiness (#drbrp8c): how much of the project is TypeScript whose types are true, the language
 * Adamic compiles native. A file is Adamic-ready when the type phase reports nothing in it and no
 * cohere:adamic rule finds anything in it, before suppression, whatever the project's chain enables. The
 * share of ready files is a segment of the footer, and the `adamic` object of the --json summary, on a run
 * that passes --adamic-readiness:
 *
 *	✓ 💎 2.4s (488 rules • 3,926 checked) • 87% Adamic-ready (3,814 of 4,387)
 *	✗ ☠️ 0.8s • 2 findings (…) • 87% Adamic-ready (3,814 of 4,387; tsconfig lacks exactOptionalPropertyTypes)
 *	✗ ☠️ 1.1s • 4 type errors (…) • Adamic readiness not measured: types bailed
 *
 * Readiness never reads as a number it did not measure. When the lint walk did not run, or the type
 * phase did not, it says not measured and why; it never falls back to 0% or to a remembered number. A file
 * a set rule skipped, or the findings cache replayed without a count, is unmeasured, never ready, and the
 * segment names how many when there are any.
 *
 * Measured only when asked (Kirk, 2026-10-05 09:25, #9tgm3dq, reversing "readiness on every run" from 00:42):
 * the measure-only rules cost a cold run about 1.6 CPU-s and 4M objects before #m6tyg79's cut, which a run
 * that does not want the number should not pay. Unasked, none of them runs, and the footer and --json say
 * nothing about readiness.
 */

// readinessSetName is the set a readiness run measures against.
const readinessSetName = configuration.SetPrefix + "adamic"

// adamicOptions are the compiler options Adamic sets itself that change what the checker proves. A
// project without one is measured under weaker options than Adamic's, and the segment says which.
var adamicOptions = []struct {
	name string
	on   func(options *core.CompilerOptions) bool
}{
	{"strict", func(options *core.CompilerOptions) bool {
		for _, value := range []core.Tristate{options.NoImplicitAny, options.StrictNullChecks, options.StrictFunctionTypes,
			options.StrictBindCallApply, options.StrictPropertyInitialization, options.UseUnknownInCatchVariables} {
			if !options.GetStrictOptionValue(value) {
				return false
			}
		}
		return true
	}},
	{"noUncheckedIndexedAccess", func(options *core.CompilerOptions) bool { return options.NoUncheckedIndexedAccess.IsTrue() }},
	{"exactOptionalPropertyTypes", func(options *core.CompilerOptions) bool { return options.ExactOptionalPropertyTypes.IsTrue() }},
}

// loadReadiness is the set's rules with their raw options, read the way the loader reads any chain. A
// cohere that carries no such set measures nothing, and says so, rather than measuring against nothing.
func loadReadiness() (*program.Readiness, error) {
	set, err := configuration.Load(readinessSetName)
	if err != nil {
		return nil, err
	}
	resolution := set.Resolve("index.ts")
	options := map[string][]json.RawMessage{}
	for _, name := range set.RuleKeys() {
		if status, _ := resolution.StatusOf(name); status == configuration.StatusEnabled {
			options[name] = resolution.RawOptionsFor(name)
		}
	}
	if len(options) == 0 {
		return nil, fmt.Errorf("%s turns on no rule, so it would measure nothing", readinessSetName)
	}
	return &program.Readiness{Options: options}, nil
}

// readinessSummary is a run's readiness, as the footer and --json say it.
type readinessSummary struct {
	// Measured is false when the run could not say, and Reason then says why, in a clause.
	Measured bool   `json:"measured"`
	Reason   string `json:"reason,omitempty"`

	// Files is every file the walk measured or tried to, Ready the ones that pass, and Unmeasured the ones a
	// set rule skipped or the cache replayed without a count.
	Files      int `json:"files"`
	Ready      int `json:"ready"`
	Unmeasured int `json:"unmeasured"`

	// OptionsMissing is the Adamic compiler options the project's tsconfig leaves off.
	OptionsMissing []string `json:"optionsMissing"`

	// FailingFilesByRule counts, by set rule, the files it found something in.
	FailingFilesByRule map[string]int `json:"failingFilesByRule"`
}

// notMeasured is a readiness the run could not take.
func notMeasured(reason string) *readinessSummary {
	return &readinessSummary{Reason: reason, OptionsMissing: []string{}, FailingFilesByRule: map[string]int{}}
}

// measureReadiness reads the walk's records against the files the type phase reported in.
func measureReadiness(records map[string]*program.AdamicRecord, typeErrorFiles map[string]bool,
	options *core.CompilerOptions) *readinessSummary {
	summary := &readinessSummary{Measured: true, OptionsMissing: []string{}, FailingFilesByRule: map[string]int{}}
	if options != nil {
		for _, option := range adamicOptions {
			if !option.on(options) {
				summary.OptionsMissing = append(summary.OptionsMissing, option.name)
			}
		}
	}
	for fileName, record := range records {
		summary.Files++
		if record == nil {
			// Replayed from an entry whose run never measured: it has no counts to read.
			summary.Unmeasured++
			continue
		}
		for _, count := range record.Counts {
			if count.Findings > 0 {
				summary.FailingFilesByRule[count.Rule]++
			}
		}
		switch {
		case !record.Measured():
			summary.Unmeasured++
		case record.Findings() == 0 && !typeErrorFiles[fileName]:
			summary.Ready++
		}
	}
	return summary
}

// segment is the footer's readiness segment, empty for a run that took none.
func (summary *readinessSummary) segment() string {
	if summary == nil {
		return ""
	}
	if !summary.Measured {
		return "Adamic readiness not measured: " + summary.Reason
	}
	if summary.Files == 0 {
		return "Adamic readiness not measured: no file was linted"
	}
	text := fmt.Sprintf("%d%% Adamic-ready (%s of %s", summary.Ready*100/summary.Files, grouped(summary.Ready), grouped(summary.Files))
	if summary.Unmeasured > 0 {
		text += "; " + grouped(summary.Unmeasured) + " unmeasured"
	}
	if len(summary.OptionsMissing) > 0 {
		text += "; tsconfig lacks " + joinedOptions(summary.OptionsMissing)
	}
	return text + ")"
}

// topFailingRules is the set rules by files failed, most first, for --verbose.
func (summary *readinessSummary) topFailingRules(limit int) []string {
	if summary == nil {
		return nil
	}
	names := make([]string, 0, len(summary.FailingFilesByRule))
	for name := range summary.FailingFilesByRule {
		names = append(names, name)
	}
	sort.Slice(names, func(left, right int) bool {
		if summary.FailingFilesByRule[names[left]] != summary.FailingFilesByRule[names[right]] {
			return summary.FailingFilesByRule[names[left]] > summary.FailingFilesByRule[names[right]]
		}
		return names[left] < names[right]
	})
	if len(names) > limit {
		names = names[:limit]
	}
	lines := make([]string, 0, len(names))
	for _, name := range names {
		lines = append(lines, fmt.Sprintf("%s %s", name, counted(summary.FailingFilesByRule[name], "file", "files")))
	}
	return lines
}

func joinedOptions(names []string) string {
	switch len(names) {
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	}
	text := ""
	for index, name := range names {
		switch {
		case index == len(names)-1:
			text += ", and " + name
		case index > 0:
			text += ", " + name
		default:
			text = name
		}
	}
	return text
}

// readinessLoaded holds loadReadiness's answer, read once per process: the set is in the binary.
var readinessLoaded = struct {
	once      sync.Once
	readiness *program.Readiness
	err       error
}{}

// readinessSet is the set to measure against, or why there is none.
func readinessSet() (*program.Readiness, error) {
	readinessLoaded.once.Do(func() {
		readinessLoaded.readiness, readinessLoaded.err = loadReadiness()
	})
	return readinessLoaded.readiness, readinessLoaded.err
}

// adamicReadinessRequested is --adamic-readiness: whether this run measures readiness at all.
var adamicReadinessRequested bool

// requestedReadiness is the set this run measures against, nil unless --adamic-readiness asked: a run that did
// not ask runs none of the set's measure-only rules, and its findings cache records no readiness, which a later
// run that asks misses on rather than replays (FindingsReuse.MeasureReadiness).
func requestedReadiness() *program.Readiness {
	if !adamicReadinessRequested {
		return nil
	}
	readiness, _ := readinessSet()
	return readiness
}

// readinessOf is the run's readiness from its walk.
func readinessOf(graph *program.Graph, result program.Result, typesRan bool, typeErrorFiles map[string]bool) *readinessSummary {
	if _, err := readinessSet(); err != nil {
		return notMeasured(err.Error())
	}
	if !typesRan {
		return notMeasured("types did not run, so no file's type errors are known")
	}
	if result.Adamic == nil {
		return notMeasured("the walk measured nothing")
	}
	return measureReadiness(result.Adamic, typeErrorFiles, graph.Program.Options())
}
