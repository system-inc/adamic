// The Tier 2 differential: `{order, count}` per class, over a whole registry, against the engine.
//
// This is the tier above the per-component fixtures rather than a replacement for them. Six
// components have landed and each carries an engine-verified corpus of its own, which is the right
// shape for testing a component and cannot see the thing this file exists to see.
//
// # What a component fixture cannot reach
//
// `descriptor_fixtures.json` hands the Go side a value that has already been decoded: a
// `{kind, root, value, modifier}` the JavaScript extractor parsed, with the question "does Lookup
// map this to the right reading". That tests the lookup exactly, and it steps over the seam. A class
// is a *string*. Turning the string into a candidate is `ParseCandidate`'s job, and a suite that
// begins from a candidate has tested both halves while never testing the join.
//
// The precedent is on the record in this slice. `segment` was verified exclusively through
// `InferDataType`, which launders its errors, so mutations to `segment` changed nothing any test
// could observe. A component reachable only through its caller is not tested — and a *seam* is
// reachable only through a fixture that has already crossed it, so the same hole applies one level
// up. This file starts from the class name and ends at the reading, and everything between is the
// Go side's problem, which is the point.
//
// # The outcome lattice, and why silence is a bucket rather than a pass
//
// Every class lands in exactly one Outcome. The lattice is the whole design, because the failure
// this harness was built to refuse is a differential that scores mutual silence as agreement. In the
// ahra corpus 10 of 1,208 distinct classes return null from the engine, and the expanded registry
// adds 30 more per system where the engine advertises a modifier it then refuses. Those are the
// classes most likely to break and the ones a naive differential is blind to, so `BothSilent` is
// counted, reported, and bounded — never folded into agreement.
//
// The rate a reader trusts is therefore computed over *answered* classes, and the silent counts are
// asserted separately against measured figures. A run in which every class went silent scores an
// agreement rate of zero over an empty denominator rather than a hundred percent over nothing, and
// a Population that fell below its floor disqualifies the comparison before any rate is read.
//
// # The vocabulary is `internal/differential`'s
//
// `Population`, `Control` and `ControlsProven` are that package's ideas and are not renamed here:
// the empty diff is not proof of agreement, and a harness that has never returned a positive has not
// been shown able to. `Finding` does not transfer, since a per-class `{order, count}` comparison is
// not keyed like a rule finding, so this file carries `Divergence` instead and says so rather than
// inventing a second word for the same concept.
package tailwind

import (
	"fmt"
	"sort"
	"strings"
)

// ReadingSource is the Go side of the differential, pluggable so a component wires in as it lands.
//
// The interface exists because the port arrives in phases and the harness is built alongside them
// rather than after. Today `LiveReadingSource` runs the real path, `ParseCandidate` into
// `Table.Lookup`. As the `@utility` evaluator and the variant work land they replace or wrap it, and
// the lattice, the controls and the assertions below do not change. A harness written against the
// concrete path would have to be rewritten by whoever lands the next component, which is how a
// harness stops being run.
//
// Reading returns `ok` false for "this side has no answer", which is a first-class outcome and never
// an error. The distinction is load-bearing: declining to answer is what the descriptor model does
// on the 18 per-declaration roots by design, and it must be visible as silence rather than
// laundered into a zero Reading that compares equal to something.
type ReadingSource interface {
	// Name identifies the source in the report.
	Name() string
	// Reading returns the reading for a class name, and whether this side can answer at all.
	Reading(className string) (Reading, bool)
}

// Outcome is which of the five states a single class landed in.
//
// Five rather than two, and the three beyond agree/disagree are the reason this type exists. A
// two-state comparison has to put "neither side answered" somewhere, and every place it can put it
// is wrong: as agreement it is blind, as disagreement it is noise on classes the engine genuinely
// does not compile.
type Outcome string

const (
	// OutcomeAgreed is both sides answering, with equal readings. The only clean state.
	OutcomeAgreed Outcome = "agreed"
	// OutcomeDisagreed is both sides answering different readings. A defect in the port.
	OutcomeDisagreed Outcome = "disagreed"
	// OutcomeGoSilent is the engine answering and the Go side declining.
	//
	// A defect unless it is a known boundary of the model. The descriptor model declines the 18
	// per-declaration roots deliberately, and that is the difference between a boundary that was
	// measured and a component that quietly stopped answering, which look identical in a count.
	OutcomeGoSilent Outcome = "go-silent"
	// OutcomeEngineSilent is the Go side answering where the engine returned null.
	//
	// Always a defect, and the direction that is easiest to miss. The Go side inventing a reading
	// for a class the engine does not compile is a rule that will sort a class nobody can write, and
	// no count of agreements will surface it.
	OutcomeEngineSilent Outcome = "engine-silent"
	// OutcomeBothSilent is neither side answering.
	//
	// Counted and bounded, never scored. This is requirement three of the harness: a side reporting
	// nothing is a failed run and not a clean one, so this bucket has a ceiling of its own and a run
	// that drifts into silence fails on the ceiling rather than passing on an empty diff.
	OutcomeBothSilent Outcome = "both-silent"
)

// Divergence is one class the two sides did not agree on.
//
// Named apart from `internal/differential`'s `Finding` because it is not one: a Finding is a rule
// firing at a file and a line, keyed so two gates' spellings can be collapsed onto each other. This
// is a class and two readings. Reusing the name would imply a shared key that does not exist.
type Divergence struct {
	// ClassName is the class both sides were asked about.
	ClassName string
	// Population is which population it came from, `registry` or `corpus`.
	Population string
	// Outcome is which of the four non-agreeing states this landed in.
	Outcome Outcome
	// EngineReading is what the engine said, valid only when EngineAnswered.
	EngineReading Reading
	// EngineAnswered is whether the engine produced a reading at all.
	EngineAnswered bool
	// GoReading is what the Go side said, valid only when GoAnswered.
	GoReading Reading
	// GoAnswered is whether the Go side produced a reading at all.
	GoAnswered bool
	// Planted marks a divergence a control caused rather than one observed in the design system.
	//
	// Carried for the reason `internal/differential` carries it: a planted divergence and an
	// observed one are the same shape in a table, and a run with three controls once reported
	// "differences: 3" on a tree whose real gap was zero. Three problems, and zero problems plus
	// three proofs the instrument works, are not the same sentence. The controls still appear,
	// because hiding them makes the evidence of detection invisible where a reader is looking.
	Planted bool
}

// Population is what a run covered, as distinct from what it found.
//
// The name and the reason are `internal/differential`'s: without this a report cannot tell agreement
// from mutual silence, and a side claiming zero classes is enough on its own to disqualify the
// comparison. Recorded per design system, since two systems is the floor and one of them going empty
// while the other carries the run is exactly the shape a total would hide.
type Population struct {
	// SystemName is the design system this population was measured on.
	SystemName string
	// RegistryClasses and CorpusClasses are the two populations, kept apart because they answer
	// different questions. The registry is what the engine advertises and the corpus is what someone
	// happened to write, and an earlier table in this slice built from a corpus missed nine
	// families.
	RegistryClasses int
	CorpusClasses   int
	// EngineNullReadings is how many classes the engine itself declined, across both populations.
	// A run where this is zero on the corpus has almost certainly lost its corpus.
	EngineNullReadings int
	// UnknownPopulationClasses is how many classes named a population this package does not know.
	//
	// Never zero by accident: it can only be non-zero if a caller invented a population name, and
	// that miscounts the two floors below. Reported rather than absorbed, because absorbing it into
	// the registry count is exactly the shape of silent miscount the floors exist to catch.
	UnknownPopulationClasses int
	// ByOutcome counts every class by where it landed. The five sum to the two class counts.
	ByOutcome map[Outcome]int
}

// Total is every class this population covered, unknown ones included.
//
// Included rather than dropped so that the both-silent rate is computed over what was actually
// walked. A denominator that quietly excluded some classes would understate every rate derived from
// it, which is the direction that makes a run look better than it was.
func (population Population) Total() int {
	return population.RegistryClasses + population.CorpusClasses + population.UnknownPopulationClasses
}

// Answered is the classes on which both sides produced a reading.
//
// This is the denominator of the agreement rate, and it is deliberately not `Total`. Scoring
// agreement over every class would let a run that went entirely silent report a high rate, since the
// silent classes are not disagreements. Over answered classes a silent run reports zero of zero,
// which reads as the absence of evidence it is.
func (population Population) Answered() int {
	return population.ByOutcome[OutcomeAgreed] + population.ByOutcome[OutcomeDisagreed]
}

// Control is a deliberate defect the harness must detect.
//
// The concept is `internal/differential`'s and so is the reason: a check that has never returned a
// positive has not been shown able to. It is planted differently here, because there is no tree to
// write a file into. A control wraps the Go `ReadingSource` and corrupts its answers, and the run is
// then required to land a stated number of classes in a stated Outcome.
//
// The required outcome is named rather than inferred from "something failed", and that is the part
// worth being strict about. A control that corrupts a reading and a control that removes a root
// both make the run fail, and they fail in different directions: one produces disagreements and the
// other produces Go-side silence. A harness that accepted either as proof of the other would report
// detection it had not demonstrated, which is the failure the whole package refuses.
type Control struct {
	// Name identifies the control in the report.
	Name string
	// Detail says what the control corrupts, in one line, for the report.
	Detail string
	// Corrupt wraps the honest Go side in one that is deliberately wrong.
	Corrupt func(honest ReadingSource) ReadingSource
	// ExpectedOutcome is the bucket the corruption must land classes in.
	ExpectedOutcome Outcome
	// MinimumDetections is how many classes must land there. A floor rather than an exact count,
	// because the registry grows and a control that has to be re-tuned on every token added is a
	// control that gets deleted. It only has to be high enough that a corruption which changed
	// nothing cannot pass.
	MinimumDetections int
}

// ControlResult is what a control actually did.
type ControlResult struct {
	// Control is the control that ran.
	Control Control
	// Detections is how many classes landed in the expected outcome.
	Detections int
	// Detected is whether it met its floor.
	Detected bool
	// Detail explains a miss, empty on a hit.
	Detail string
}

// Report is one differential run: what it covered, what it found, and whether it can be trusted.
type Report struct {
	// TailwindVersion is the engine version the fixture was captured against.
	TailwindVersion string
	// GoSideName identifies the Go implementation under test.
	GoSideName string
	// Populations is one entry per design system, in fixture order.
	Populations []Population
	// Divergences is every non-agreeing class, controls included and marked.
	Divergences []Divergence
	// Controls is what each control did.
	Controls []ControlResult
	// SystemDivergence records the two design systems against each other rather than against Go.
	// See RepositoriesDiverge.
	SystemDivergence SystemDivergence
}

// SystemDivergence is how far apart the design systems themselves are.
//
// It is a property of the fixture rather than of a run, and it is passed into Compare rather than
// defaulted, so that a caller which forgets it fails the divergence assertion loudly instead of
// reporting three zeros. Three zeros here read as "the systems are identical", which is the exact
// claim RepositoriesDiverge exists to refuse to make by accident.
type SystemDivergence struct {
	// SharedRegistryClasses is how many registry classes both systems hold.
	SharedRegistryClasses int
	// DivergentReadings is how many of those shared classes read differently. Measured zero: order
	// indexes the declared property, and two systems declaring the same property with different
	// values read alike.
	DivergentReadings int
	// UniqueToOneSystem is how many registry classes exist in exactly one system. Measured 462, and
	// this is where the per-repository divergence actually lives.
	UniqueToOneSystem int
}

// ControlsProven is whether the harness was shown able to detect a difference at all.
//
// Reported separately from Trustworthy because they answer different questions, which is the
// distinction `internal/differential` draws and the one worth keeping. Trustworthy asks whether this
// run's population was real. This asks whether the instrument works.
//
// Every control must have fired, and they must cover both directions of silence. A harness that can
// only see Go-side silence reports a clean run for every class where Go invents a reading the engine
// never produced, and that direction is the one nothing else in the suite is looking at.
func (report Report) ControlsProven() bool {
	// Stated explicitly, though the loop below would reach the same answer: with no controls both
	// direction flags stay false and the final conjunction returns false anyway. Mutation-tested and
	// confirmed equivalent, so this line is documentation rather than logic, and it is kept because
	// "a run with no controls has proven nothing" is the single most important thing this function
	// says and it should not have to be derived from an empty loop.
	if len(report.Controls) == 0 {
		return false
	}
	sawGoSilentDirection := false
	sawEngineSilentDirection := false
	for _, result := range report.Controls {
		if !result.Detected {
			return false
		}
		switch result.Control.ExpectedOutcome {
		case OutcomeGoSilent:
			sawGoSilentDirection = true
		case OutcomeEngineSilent:
			sawEngineSilentDirection = true
		}
	}
	return sawGoSilentDirection && sawEngineSilentDirection
}

// minimumPlausibleRegistryClasses is the floor below which a run is assumed to have been pointed at
// the wrong thing.
//
// A threshold rather than an exact count, for the reason the same constant carries in
// `internal/differential`: the design systems grow, and a harness that must be edited every time a
// token lands is a harness whose guard gets commented out. The registry measured 95,136 and 95,594
// expanded across the two systems, so this is roughly a quarter of the smaller and cannot be reached
// by a misconfigured run.
const minimumPlausibleRegistryClasses = 20000

// minimumPlausibleCorpusClasses is the same floor for the corpus, which is two orders smaller.
//
// The corpus is the population that carries the engine's null readings, so a run that lost it still
// compares 190,000 registry classes and looks entirely healthy while having stopped testing the one
// case this harness was built for.
const minimumPlausibleCorpusClasses = 400

// maximumPlausibleBothSilentRate bounds the silence.
//
// This is requirement three as a number. Both-silent classes are not scored as agreement, which
// stops them inflating the rate, and that alone does not stop a run drifting into total silence:
// zero disagreements over zero answered classes is not a failure in any comparison that only looks
// at divergences. So the silence has a ceiling. Measured, the expanded registry runs 30 engine nulls
// per system out of 95,136 and the corpus 11 of 1,244 on ahra and 84 of 1,055 on connected, which is
// under one percent overall; five percent is far enough above that to never fire on a real run and
// far below any run that has actually gone quiet.
const maximumPlausibleBothSilentRate = 0.05

// Trustworthy reports whether this run's population was real, with the reasons it was not.
//
// Separate from whether it found anything, and checked first. Every reason below describes a run
// whose agreement numbers mean nothing, and the point of returning them as strings is that a caller
// prints them: a boolean false here reads as "the port is broken" when it means "this run did not
// measure anything".
func (report Report) Trustworthy() (bool, []string) {
	var reasons []string

	if len(report.Populations) < 2 {
		// One design system is not a population. Per-repository divergence is the defect the whole
		// port exists to fix, and it is unobservable from inside a single repository.
		reasons = append(reasons, fmt.Sprintf(
			"%d design systems measured, and per-repository divergence cannot be observed from one",
			len(report.Populations),
		))
	}

	for _, population := range report.Populations {
		if population.RegistryClasses < minimumPlausibleRegistryClasses {
			reasons = append(reasons, fmt.Sprintf(
				"%s: %d registry classes is below the plausible floor of %d, so the run was pointed at the wrong thing",
				population.SystemName, population.RegistryClasses, minimumPlausibleRegistryClasses,
			))
		}
		if population.CorpusClasses < minimumPlausibleCorpusClasses {
			reasons = append(reasons, fmt.Sprintf(
				"%s: %d corpus classes is below the plausible floor of %d, and the corpus is where the engine's null readings live",
				population.SystemName, population.CorpusClasses, minimumPlausibleCorpusClasses,
			))
		}
		if population.Answered() == 0 {
			reasons = append(reasons, fmt.Sprintf(
				"%s: no class was answered by both sides, so there is nothing to agree about",
				population.SystemName,
			))
		}
		if total := population.Total(); total > 0 {
			bothSilentRate := float64(population.ByOutcome[OutcomeBothSilent]) / float64(total)
			if bothSilentRate > maximumPlausibleBothSilentRate {
				reasons = append(reasons, fmt.Sprintf(
					"%s: %d of %d classes (%.1f%%) had neither side answer, above the %.0f%% ceiling — a side reporting nothing is a failed run",
					population.SystemName, population.ByOutcome[OutcomeBothSilent], total,
					bothSilentRate*100, maximumPlausibleBothSilentRate*100,
				))
			}
		}
		if population.UnknownPopulationClasses > 0 {
			reasons = append(reasons, fmt.Sprintf(
				"%s: %d classes named a population this package does not know, so the registry and corpus floors were measured against the wrong counts",
				population.SystemName, population.UnknownPopulationClasses,
			))
		}
		if population.EngineNullReadings == 0 {
			// The engine declines classes in both design systems and always has. Zero means the
			// fixture stopped carrying the classes it declines, which removes exactly the population
			// this harness was built to keep honest, and it removes it invisibly.
			reasons = append(reasons, fmt.Sprintf(
				"%s: the engine declined no class at all, so the population that proves silence is not scored has gone missing",
				population.SystemName,
			))
		}
	}

	if !report.ControlsProven() {
		reasons = append(reasons, "no control fired in both silent directions, so the harness has not been shown able to detect a difference")
	}
	for _, result := range report.Controls {
		if result.Detected {
			continue
		}
		reasons = append(reasons, fmt.Sprintf(
			"control %q did not fire: %s — the harness has not been shown able to detect a difference",
			result.Control.Name, result.Detail,
		))
	}

	return len(reasons) == 0, reasons
}

// RepositoriesDiverge reports whether the two design systems differ at all, and how.
//
// This is `f7d1d8d`'s discipline: take the claim the port rests on and make it an assertion that
// fails if it ever becomes trivially true. The claim is that a table generated from one repository
// is wrong for another.
//
// Measured, the divergence is not where reading the theme result suggests. Theme resolution diverges
// on *values*, 744 entries against 748 with 22 shared keys carrying different values, and a reading
// is `{order, count}` where order indexes the declared *property*. A shared class declaring the same
// property with a different value therefore reads identically in both, and all 95,134 shared
// registry classes do.
//
// The divergence is in *membership*: 462 classes exist in exactly one system, because a
// `--color-brand-*` token in one repository's `@theme` is what makes `accent-brand` a class at all.
// A table generated against ahra has no row answering `accent-brand`, so it declines a class the
// engine compiles. That is a silent-side defect rather than a wrong reading, and it is the sharper
// statement of why a per-repository table is required.
//
// Both axes are returned so that a future Tailwind which starts diverging readings shows up as a
// number moving off zero rather than as a fixture that quietly still passes.
func (report Report) RepositoriesDiverge() (bool, string) {
	if report.SystemDivergence.UniqueToOneSystem == 0 && report.SystemDivergence.DivergentReadings == 0 {
		return false, fmt.Sprintf(
			"the two design systems agree on all %d shared registry classes and neither holds a class the other lacks, "+
				"so this fixture no longer demonstrates that a per-repository table is required",
			report.SystemDivergence.SharedRegistryClasses,
		)
	}
	return true, fmt.Sprintf(
		"%d shared registry classes read identically, %d read differently, and %d classes exist in only one system",
		report.SystemDivergence.SharedRegistryClasses, report.SystemDivergence.DivergentReadings, report.SystemDivergence.UniqueToOneSystem,
	)
}

// Describe renders the run as prose a reader can check, controls separated from observations.
func (report Report) Describe() string {
	builder := &strings.Builder{}

	fmt.Fprintf(builder, "tailwind %s, go side %q\n", report.TailwindVersion, report.GoSideName)

	for _, population := range report.Populations {
		fmt.Fprintf(builder,
			"  %s: %d registry + %d corpus classes, %d answered by both\n",
			population.SystemName, population.RegistryClasses, population.CorpusClasses, population.Answered(),
		)
		fmt.Fprintf(builder,
			"    agreed %d, disagreed %d, go-silent %d, engine-silent %d, both-silent %d (engine declined %d)\n",
			population.ByOutcome[OutcomeAgreed], population.ByOutcome[OutcomeDisagreed],
			population.ByOutcome[OutcomeGoSilent], population.ByOutcome[OutcomeEngineSilent],
			population.ByOutcome[OutcomeBothSilent], population.EngineNullReadings,
		)
		if answered := population.Answered(); answered > 0 {
			fmt.Fprintf(builder, "    agreement over answered classes: %.4f%%\n",
				float64(population.ByOutcome[OutcomeAgreed])/float64(answered)*100)
		}
	}

	// Observed and planted are counted apart, never summed. A run with three controls reporting
	// "3 divergences" on a design system whose real gap was zero is a sentence a reader takes as
	// remaining exposure, and that has happened in this repository.
	observed, planted := 0, 0
	for _, divergence := range report.Divergences {
		if divergence.Planted {
			planted++
			continue
		}
		observed++
	}
	fmt.Fprintf(builder, "  divergences: %d observed, %d planted by controls\n", observed, planted)

	for _, result := range report.Controls {
		status := "did not fire"
		if result.Detected {
			status = "fired"
		}
		fmt.Fprintf(builder, "    control %-22s %s: %d detections (floor %d) — %s\n",
			result.Control.Name, status, result.Detections, result.Control.MinimumDetections, result.Control.Detail)
	}

	_, divergenceDetail := report.RepositoriesDiverge()
	fmt.Fprintf(builder, "  design systems: %s\n", divergenceDetail)

	return builder.String()
}

// DivergencesByOutcome groups the observed divergences, sorted, for a report a reader can act on.
//
// Planted divergences are excluded. They are proof the instrument works and they are not findings,
// and a list that mixed them would hand a reader control noise to triage.
func (report Report) DivergencesByOutcome() map[Outcome][]Divergence {
	grouped := map[Outcome][]Divergence{}
	for _, divergence := range report.Divergences {
		if divergence.Planted {
			continue
		}
		grouped[divergence.Outcome] = append(grouped[divergence.Outcome], divergence)
	}
	for outcome := range grouped {
		sort.Slice(grouped[outcome], func(left, right int) bool {
			return grouped[outcome][left].ClassName < grouped[outcome][right].ClassName
		})
	}
	return grouped
}

// Format renders a reading the way the extractor's `readingKey` does, so a Go failure message and a
// JavaScript one can be compared by eye without translating between two spellings.
func (reading Reading) Format() string {
	parts := make([]string, len(reading.Order))
	for index, order := range reading.Order {
		parts[index] = fmt.Sprint(order)
	}
	return "[" + strings.Join(parts, ",") + "]#" + fmt.Sprint(reading.Count)
}

// PopulationRegistry and PopulationCorpus are the two populations, named rather than spelled at each
// use site.
//
// They answer different questions and the difference is load-bearing. The registry is what the
// engine advertises, which is the population; the corpus is what someone happened to write, which is
// a sample, and an earlier table in this slice built from a corpus missed nine families. They also
// carry different evidence: the corpus is where the engine's null readings live, so a run that lost
// it still compares 190,000 registry classes and looks healthy while having stopped testing the case
// this harness exists for.
const (
	PopulationRegistry = "registry"
	PopulationCorpus   = "corpus"
)

// ClassCase is one class and the engine's answer for it.
type ClassCase struct {
	// ClassName is the class as an author would write it.
	ClassName string
	// Population is PopulationRegistry or PopulationCorpus.
	Population string
	// EngineReading is what the engine computed, valid only when EngineAnswered.
	EngineReading Reading
	// EngineAnswered is whether the engine produced a reading. False is a real measurement and not
	// missing data: the engine declines `border/50`, which it advertised, and 10 corpus classes on
	// ahra.
	EngineAnswered bool
}

// SystemCases is one design system's whole population.
type SystemCases struct {
	// SystemName identifies the design system.
	SystemName string
	// Cases is every class, both populations, in fixture order.
	Cases []ClassCase
}

// Compare runs one Go side against the engine's answers and returns the report.
//
// The controls run inside the same call rather than in a second pass, for the reason
// `internal/differential` plants before running: two passes over a population that can change
// between them is a window, and the corrupted side must be measured against exactly the answers the
// honest side was measured against or the detection count means something else.
func Compare(tailwindVersion string, systems []SystemCases, divergence SystemDivergence, goSide ReadingSource, controls []Control) Report {
	report := Report{
		TailwindVersion:  tailwindVersion,
		GoSideName:       goSide.Name(),
		SystemDivergence: divergence,
	}

	for _, system := range systems {
		population := Population{SystemName: system.SystemName, ByOutcome: map[Outcome]int{}}

		for _, aCase := range system.Cases {
			// Counted by an exhaustive switch rather than by a default, so a population this
			// function does not know is a loud unknown instead of a silent registry class. A
			// miscounted population moves the floors that decide whether the run is trustworthy at
			// all, which is the one number that must not be quietly wrong.
			switch aCase.Population {
			case PopulationCorpus:
				population.CorpusClasses++
			case PopulationRegistry:
				population.RegistryClasses++
			default:
				population.UnknownPopulationClasses++
			}
			if !aCase.EngineAnswered {
				population.EngineNullReadings++
			}

			goReading, goAnswered := goSide.Reading(aCase.ClassName)
			outcome := classify(aCase, goReading, goAnswered)
			population.ByOutcome[outcome]++
			if outcome == OutcomeAgreed {
				continue
			}
			report.Divergences = append(report.Divergences, Divergence{
				ClassName:      aCase.ClassName,
				Population:     aCase.Population,
				Outcome:        outcome,
				EngineReading:  aCase.EngineReading,
				EngineAnswered: aCase.EngineAnswered,
				GoReading:      goReading,
				GoAnswered:     goAnswered,
			})
		}

		report.Populations = append(report.Populations, population)
	}

	for _, control := range controls {
		report.Controls = append(report.Controls, runControl(systems, goSide, control, &report))
	}

	return report
}

// classify places one class in the lattice.
//
// Written as an explicit two-by-two on "did each side answer" rather than as a chain of early
// returns, because the case that matters is the one an early-return chain drops: both sides silent
// is not a `continue`, it is a bucket with a name and a ceiling.
func classify(aCase ClassCase, goReading Reading, goAnswered bool) Outcome {
	switch {
	case aCase.EngineAnswered && goAnswered:
		if aCase.EngineReading.Equal(goReading) {
			return OutcomeAgreed
		}
		return OutcomeDisagreed
	case aCase.EngineAnswered && !goAnswered:
		return OutcomeGoSilent
	case !aCase.EngineAnswered && goAnswered:
		return OutcomeEngineSilent
	default:
		return OutcomeBothSilent
	}
}

// runControl replays the population against a corrupted Go side and checks the corruption was seen.
//
// The corrupted divergences are appended to the report marked Planted, so the evidence that
// detection works is visible in the same table a reader is already reading, and the summary counts
// them apart from observations.
func runControl(systems []SystemCases, honest ReadingSource, control Control, report *Report) ControlResult {
	corrupted := control.Corrupt(honest)
	result := ControlResult{Control: control, Detail: control.Detail}

	for _, system := range systems {
		for _, aCase := range system.Cases {
			goReading, goAnswered := corrupted.Reading(aCase.ClassName)
			outcome := classify(aCase, goReading, goAnswered)
			if outcome != control.ExpectedOutcome {
				continue
			}
			result.Detections++
			// One example per control is enough to make the mechanism legible in the report;
			// carrying tens of thousands of planted rows would bury the observed ones.
			if result.Detections == 1 {
				report.Divergences = append(report.Divergences, Divergence{
					ClassName:      aCase.ClassName,
					Population:     aCase.Population,
					Outcome:        outcome,
					EngineReading:  aCase.EngineReading,
					EngineAnswered: aCase.EngineAnswered,
					GoReading:      goReading,
					GoAnswered:     goAnswered,
					Planted:        true,
				})
			}
		}
	}

	result.Detected = result.Detections >= control.MinimumDetections
	if !result.Detected {
		result.Detail = fmt.Sprintf(
			"%s — landed %d classes in %s, below the floor of %d",
			control.Detail, result.Detections, control.ExpectedOutcome, control.MinimumDetections,
		)
	}
	return result
}

// AcknowledgedDivergence is one class the two sides are expected to disagree on, and why.
//
// The concept and the mandatory Reason are `internal/differential`'s, and so is the discipline that
// makes them worth having: an acknowledgement with no reason is indistinguishable from a suppression
// someone added to turn a red build green.
//
// Keyed to the exact class rather than to a root or a pattern. A root-wide excuse would swallow the
// next genuine drift under that root, and every acknowledgement below covers a mechanism whose
// extent was measured, so naming the classes costs nothing that the measurement did not already pay.
type AcknowledgedDivergence struct {
	// ClassName is the exact class this excuses.
	ClassName string
	// Outcome is the bucket it is expected to land in. An acknowledgement excuses one outcome and
	// not the class generally: a class acknowledged as go-silent that starts reading *wrongly* is a
	// new defect and must not be covered by the old excuse.
	Outcome Outcome
	// Owner is the task that will resolve it, so an acknowledgement is a pointer to work rather than
	// a permanent dispensation.
	Owner string
	// Reason is why this divergence is correct today, in a sentence a reader can check.
	Reason string
}

// acknowledgedKey is the identity an acknowledgement shares with the divergence it excuses.
func acknowledgedKey(className string, outcome Outcome) string {
	return className + "\x00" + string(outcome)
}

// AcknowledgedIndex builds the lookup a run uses, and refuses an acknowledgement with no reason.
//
// The refusal is a panic rather than a silent skip because this list is source code reviewed like
// any other, and an unreasoned entry in it is a defect in the review rather than in a run.
func AcknowledgedIndex(acknowledged []AcknowledgedDivergence) map[string]AcknowledgedDivergence {
	index := make(map[string]AcknowledgedDivergence, len(acknowledged))
	for _, one := range acknowledged {
		// Go whitespace: a reason written in cohere's own Go source, which no JavaScript tool reads.
		if strings.TrimSpace(one.Reason) == "" {
			panic("tailwind: acknowledged divergence for " + one.ClassName + " has no reason; an unreasoned acknowledgement is a suppression")
		}
		index[acknowledgedKey(one.ClassName, one.Outcome)] = one
	}
	return index
}
