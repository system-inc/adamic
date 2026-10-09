package tailwind

import (
	"strings"
	"testing"
)

// The harness's own controls: proof that the instrument can fail.
//
// `classorder_differential_test.go` plants four controls inside every run, and those prove that the
// *comparison* detects a corrupted Go side. They do not prove that the run's *assertions* detect a
// corrupted comparison. Those are different claims and the second one is the one a green line rests
// on: a Trustworthy that never returns false, or a ControlsProven that cannot say no, would make
// every run above pass regardless of what it measured.
//
// This is the same argument one level up, and it is the argument the whole tier is built on. A
// component reachable only through its caller is not tested. The assertions are reachable only
// through a run that passes, so they need a caller that makes them fail on purpose.

// classOrderStubSource answers from a map, so a test can state exactly what each side said.
type classOrderStubSource struct {
	name     string
	readings map[string]Reading
}

func (source classOrderStubSource) Name() string { return source.name }

func (source classOrderStubSource) Reading(className string) (Reading, bool) {
	reading, found := source.readings[className]
	return reading, found
}

// classOrderSyntheticPopulation builds a population large enough to clear the plausibility floors,
// so a test can isolate the assertion it means to exercise instead of failing on volume.
func classOrderSyntheticPopulation(registryCount, corpusCount, engineNullCount int) ([]SystemCases, map[string]Reading) {
	readings := map[string]Reading{}
	cases := make([]ClassCase, 0, registryCount+corpusCount)

	for index := 0; index < registryCount; index++ {
		className := "reg-" + itoa(index)
		reading := Reading{Order: []int{index % 300}, Count: 1}
		cases = append(cases, ClassCase{ClassName: className, Population: PopulationRegistry, EngineReading: reading, EngineAnswered: true})
		readings[className] = reading
	}
	for index := 0; index < corpusCount; index++ {
		className := "cor-" + itoa(index)
		aCase := ClassCase{ClassName: className, Population: PopulationCorpus}
		if index >= engineNullCount {
			reading := Reading{Order: []int{index % 300}, Count: 2}
			aCase.EngineReading, aCase.EngineAnswered = reading, true
			readings[className] = reading
		}
		cases = append(cases, aCase)
	}

	return []SystemCases{{SystemName: "synthetic", Cases: cases}}, readings
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := [20]byte{}
	position := len(digits)
	for value > 0 {
		position--
		digits[position] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[position:])
}

// classOrderHealthyDivergence is a divergence block that satisfies RepositoriesDiverge, so a test of
// something else does not fail on that assertion.
func classOrderHealthyDivergence() SystemDivergence {
	return SystemDivergence{SharedRegistryClasses: 95134, DivergentReadings: 0, UniqueToOneSystem: 462}
}

// TestClassOrderHarnessScoresMutualSilenceAsFailure is the requirement this whole file exists for.
//
// A side reporting nothing must be a failed run and not a clean one. The check is deliberately
// adversarial: the Go side is silent on exactly the classes the engine is silent on, which produces
// zero disagreements, zero go-silent and zero engine-silent. Every divergence-counting assertion
// sees a perfectly clean run. Only the both-silent ceiling and the empty-denominator rule catch it.
func TestClassOrderHarnessScoresMutualSilenceAsFailure(t *testing.T) {
	t.Parallel()
	systems, readings := classOrderSyntheticPopulation(minimumPlausibleRegistryClasses+1, minimumPlausibleCorpusClasses+1, 0)

	// Silence both sides on the same classes: a tenth of the registry, well over the ceiling.
	silenced := 0
	for index := range systems[0].Cases {
		if systems[0].Cases[index].Population != PopulationRegistry || index%10 != 0 {
			continue
		}
		systems[0].Cases[index].EngineAnswered = false
		delete(readings, systems[0].Cases[index].ClassName)
		silenced++
	}

	report := Compare("4.3.3", systems, classOrderHealthyDivergence(),
		classOrderStubSource{name: "stub", readings: readings}, nil)

	if disagreed := report.Populations[0].ByOutcome[OutcomeDisagreed]; disagreed != 0 {
		t.Fatalf("the setup is wrong: %d disagreements, expected a run that is clean by every divergence count", disagreed)
	}
	if silent := report.Populations[0].ByOutcome[OutcomeBothSilent]; silent != silenced {
		t.Fatalf("expected %d both-silent classes, got %d", silenced, silent)
	}

	trustworthy, reasons := report.Trustworthy()
	if trustworthy {
		t.Fatal("a run where a tenth of the population went mutually silent was reported as trustworthy; " +
			"silence is being scored as agreement, which is the failure this harness exists to refuse")
	}
	if !classOrderAnyReasonContains(reasons, "neither side answer") {
		t.Errorf("the run was untrustworthy but not for the silence; reasons were %v", reasons)
	}
}

// TestClassOrderHarnessRefusesAnEmptyDenominator is the degenerate case of the same rule.
//
// Every class silent on both sides is zero disagreements over zero answered classes. A rate computed
// over the whole population would report a hundred percent agreement; over answered classes it is
// zero of zero, and the run must be refused rather than scored.
func TestClassOrderHarnessRefusesAnEmptyDenominator(t *testing.T) {
	t.Parallel()
	systems, _ := classOrderSyntheticPopulation(minimumPlausibleRegistryClasses+1, minimumPlausibleCorpusClasses+1, 0)
	for index := range systems[0].Cases {
		systems[0].Cases[index].EngineAnswered = false
	}

	report := Compare("4.3.3", systems, classOrderHealthyDivergence(),
		classOrderStubSource{name: "silent", readings: map[string]Reading{}}, nil)

	if answered := report.Populations[0].Answered(); answered != 0 {
		t.Fatalf("the setup is wrong: %d answered classes, expected none", answered)
	}
	trustworthy, reasons := report.Trustworthy()
	if trustworthy {
		t.Fatal("a run in which nothing was answered was reported as trustworthy")
	}
	if !classOrderAnyReasonContains(reasons, "nothing to agree about") {
		t.Errorf("expected the empty-denominator reason, got %v", reasons)
	}
}

// TestClassOrderHarnessRefusesALostPopulation proves the volume floors fire.
//
// The recorded failure this guards against is a sweep that reported twelve skills clean: a run that
// measured almost nothing looks exactly like a run that found nothing wrong.
func TestClassOrderHarnessRefusesALostPopulation(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name          string
		registryCount int
		corpusCount   int
		expectReason  string
	}{
		{"registry lost", 10, minimumPlausibleCorpusClasses + 1, "below the plausible floor"},
		{"corpus lost", minimumPlausibleRegistryClasses + 1, 3, "the corpus is where the engine's null readings live"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			systems, readings := classOrderSyntheticPopulation(testCase.registryCount, testCase.corpusCount, 1)
			report := Compare("4.3.3", systems, classOrderHealthyDivergence(),
				classOrderStubSource{name: "stub", readings: readings}, nil)

			trustworthy, reasons := report.Trustworthy()
			if trustworthy {
				t.Fatalf("a run over %d registry and %d corpus classes was reported as trustworthy",
					testCase.registryCount, testCase.corpusCount)
			}
			if !classOrderAnyReasonContains(reasons, testCase.expectReason) {
				t.Errorf("expected a reason containing %q, got %v", testCase.expectReason, reasons)
			}
		})
	}
}

// TestClassOrderHarnessRefusesAPopulationWithNoEngineNulls guards the population that proves the
// silence rule is live.
//
// If the fixture stopped carrying classes the engine declines, every assertion about not scoring
// silence would still pass while testing nothing, and it would do so invisibly.
func TestClassOrderHarnessRefusesAPopulationWithNoEngineNulls(t *testing.T) {
	t.Parallel()
	systems, readings := classOrderSyntheticPopulation(minimumPlausibleRegistryClasses+1, minimumPlausibleCorpusClasses+1, 0)
	report := Compare("4.3.3", systems, classOrderHealthyDivergence(),
		classOrderStubSource{name: "stub", readings: readings}, nil)

	if nulls := report.Populations[0].EngineNullReadings; nulls != 0 {
		t.Fatalf("the setup is wrong: %d engine nulls, expected none", nulls)
	}
	trustworthy, reasons := report.Trustworthy()
	if trustworthy {
		t.Fatal("a population in which the engine declined nothing was reported as trustworthy")
	}
	if !classOrderAnyReasonContains(reasons, "the engine declined no class at all") {
		t.Errorf("expected the missing-nulls reason, got %v", reasons)
	}
}

// TestClassOrderHarnessRequiresBothSilentDirections proves ControlsProven cannot be satisfied by one
// direction of detection standing in for the other.
//
// A harness that can see a wrong reading and cannot see an invented one reports a clean run for
// every class the Go side answers and the engine does not, which is the direction nothing else in
// the suite watches.
func TestClassOrderHarnessRequiresBothSilentDirections(t *testing.T) {
	t.Parallel()
	systems, readings := classOrderSyntheticPopulation(minimumPlausibleRegistryClasses+1, minimumPlausibleCorpusClasses+1, 5)
	honest := classOrderStubSource{name: "stub", readings: readings}

	goSilentOnly := Control{
		Name: "go-silent only", Detail: "declines everything",
		Corrupt: func(ReadingSource) ReadingSource {
			return classOrderStubSource{name: "silent", readings: map[string]Reading{}}
		},
		ExpectedOutcome: OutcomeGoSilent, MinimumDetections: 1,
	}
	engineSilentOnly := Control{
		Name: "engine-silent only", Detail: "answers everything",
		Corrupt: func(ReadingSource) ReadingSource {
			return classOrderAlwaysAnswers{}
		},
		ExpectedOutcome: OutcomeEngineSilent, MinimumDetections: 1,
	}

	if report := Compare("4.3.3", systems, classOrderHealthyDivergence(), honest, []Control{goSilentOnly}); report.ControlsProven() {
		t.Error("a single go-silent control was accepted as proof of detection in both directions")
	}
	if report := Compare("4.3.3", systems, classOrderHealthyDivergence(), honest, []Control{engineSilentOnly}); report.ControlsProven() {
		t.Error("a single engine-silent control was accepted as proof of detection in both directions")
	}
	report := Compare("4.3.3", systems, classOrderHealthyDivergence(), honest, []Control{goSilentOnly, engineSilentOnly})
	if !report.ControlsProven() {
		t.Error("both directions were demonstrated and ControlsProven still said no")
	}

	// And a run with no controls at all has demonstrated nothing, which must not read as success.
	if Compare("4.3.3", systems, classOrderHealthyDivergence(), honest, nil).ControlsProven() {
		t.Error("a run carrying no controls reported that detection had been proven")
	}
}

// classOrderAlwaysAnswers answers every class, including the ones the engine declined.
type classOrderAlwaysAnswers struct{}

func (classOrderAlwaysAnswers) Name() string { return "always answers" }

func (classOrderAlwaysAnswers) Reading(string) (Reading, bool) {
	return Reading{Order: []int{1}, Count: 1}, true
}

// TestClassOrderHarnessRefusesAControlThatDidNotFire proves a silent control invalidates a run.
//
// A control that did not fire is not a minor gap. It means the harness has not been shown able to
// detect anything, so the agreement figure it produced rests on nothing.
func TestClassOrderHarnessRefusesAControlThatDidNotFire(t *testing.T) {
	t.Parallel()
	systems, readings := classOrderSyntheticPopulation(minimumPlausibleRegistryClasses+1, minimumPlausibleCorpusClasses+1, 5)

	// A control whose corruption changes nothing: it hands back the honest source unchanged.
	inert := Control{
		Name: "inert", Detail: "corrupts nothing",
		Corrupt:         func(honest ReadingSource) ReadingSource { return honest },
		ExpectedOutcome: OutcomeDisagreed, MinimumDetections: 1,
	}

	report := Compare("4.3.3", systems, classOrderHealthyDivergence(),
		classOrderStubSource{name: "stub", readings: readings}, []Control{inert})

	if report.Controls[0].Detected {
		t.Fatal("a control that corrupted nothing reported that it fired")
	}
	if report.ControlsProven() {
		t.Error("a run with an unfired control reported that detection had been proven")
	}
	trustworthy, reasons := report.Trustworthy()
	if trustworthy {
		t.Fatal("a run with an unfired control was reported as trustworthy")
	}
	if !classOrderAnyReasonContains(reasons, "did not fire") {
		t.Errorf("expected the unfired-control reason, got %v", reasons)
	}
}

// TestClassOrderHarnessRefusesIdenticalDesignSystems is `f7d1d8d`'s discipline as a test.
//
// The claim the port rests on is that a table generated from one repository is wrong for another. If
// the two design systems ever stopped diverging, every run here would still be green and would have
// stopped saying anything, so the claim becoming trivially true has to fail.
func TestClassOrderHarnessRefusesIdenticalDesignSystems(t *testing.T) {
	t.Parallel()
	systems, readings := classOrderSyntheticPopulation(minimumPlausibleRegistryClasses+1, minimumPlausibleCorpusClasses+1, 5)
	honest := classOrderStubSource{name: "stub", readings: readings}

	identical := Compare("4.3.3", systems, SystemDivergence{SharedRegistryClasses: 95134}, honest, nil)
	if diverge, detail := identical.RepositoriesDiverge(); diverge {
		t.Errorf("two identical design systems were reported as diverging: %s", detail)
	}

	// Either axis alone is enough, and both must be recognised: the measured divergence today is
	// membership-only, and a future Tailwind could move it to readings.
	membershipOnly := Compare("4.3.3", systems, SystemDivergence{SharedRegistryClasses: 95134, UniqueToOneSystem: 462}, honest, nil)
	if diverge, _ := membershipOnly.RepositoriesDiverge(); !diverge {
		t.Error("divergence in registry membership alone was not recognised as divergence")
	}
	readingsOnly := Compare("4.3.3", systems, SystemDivergence{SharedRegistryClasses: 95134, DivergentReadings: 3}, honest, nil)
	if diverge, _ := readingsOnly.RepositoriesDiverge(); !diverge {
		t.Error("divergence in readings alone was not recognised as divergence")
	}
}

// TestClassOrderHarnessDistinguishesPlantedFromObserved guards a mistake with a record in this repo.
//
// A run with three controls once reported "differences: 3" on a tree whose real gap was zero, and
// the number was read as remaining exposure. Planted and observed are the same shape in a table, so
// they are counted apart and the actionable list holds only observations.
func TestClassOrderHarnessDistinguishesPlantedFromObserved(t *testing.T) {
	t.Parallel()
	systems, readings := classOrderSyntheticPopulation(minimumPlausibleRegistryClasses+1, minimumPlausibleCorpusClasses+1, 5)

	report := Compare("4.3.3", systems, classOrderHealthyDivergence(),
		classOrderStubSource{name: "stub", readings: readings},
		[]Control{{
			Name: "shifts", Detail: "shifts every count",
			Corrupt: func(honest ReadingSource) ReadingSource {
				return classOrderCorruptedSource{honest: honest, transform: func(reading Reading, answered bool) (Reading, bool) {
					if !answered {
						return reading, answered
					}
					return Reading{Order: reading.Order, Count: reading.Count + 1}, true
				}}
			},
			ExpectedOutcome: OutcomeDisagreed, MinimumDetections: 1,
		}})

	planted := 0
	for _, divergence := range report.Divergences {
		if divergence.Planted {
			planted++
		}
	}
	if planted == 0 {
		t.Fatal("the control fired and planted no divergence, so the evidence of detection is invisible in the report")
	}
	for outcome, divergences := range report.DivergencesByOutcome() {
		for _, divergence := range divergences {
			if divergence.Planted {
				t.Errorf("a planted divergence appeared in the actionable %s list for %s", outcome, divergence.ClassName)
			}
		}
	}
	if !strings.Contains(report.Describe(), "planted by controls") {
		t.Error("the report does not separate planted divergences from observed ones")
	}
}

// TestClassOrderAcknowledgementsRequireAReason proves an unreasoned excuse is refused.
//
// An acknowledgement with no reason is indistinguishable from a suppression added to turn a red
// build green, which is why the index panics rather than skipping it.
func TestClassOrderAcknowledgementsRequireAReason(t *testing.T) {
	t.Parallel()
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Error("an acknowledgement with no reason was accepted")
		}
	}()
	AcknowledgedIndex([]AcknowledgedDivergence{{ClassName: "flex", Outcome: OutcomeGoSilent, Owner: "#nobody"}})
}

// TestClassOrderUnresolvableAlphaPredicateIsNarrow guards the acknowledgement's own blast radius.
//
// The predicate excuses a real disagreement, so it has to match the mechanism and nothing adjacent.
// The discriminator is whether the modifier resolves at build time, which is why `/[--x]` and
// `/[calc(1/2)]` must not match while `/[var(--a)]` does, and why no non-shadow root may match.
func TestClassOrderUnresolvableAlphaPredicateIsNarrow(t *testing.T) {
	t.Parallel()
	for _, className := range []string{
		"shadow-lg/[var(--a)]", "inset-shadow-sm/[var(--a)]", "text-shadow-xs/[var(--a)]",
		"drop-shadow-md/[var(--a)]", "hover:shadow-lg/[var(--a)]",
	} {
		if !classOrderUnresolvableAlpha(className) {
			t.Errorf("%s is the mechanism and was not matched", className)
		}
	}
	for _, className := range []string{
		"shadow-lg/[0.5]", "shadow-lg/[--x]", "shadow-lg/[calc(1/2)]", "shadow-lg/50", "shadow-lg",
		"bg-red-500/[var(--a)]", "text-red-500/[var(--a)]", "border-red-500/[var(--a)]",
	} {
		if classOrderUnresolvableAlpha(className) {
			t.Errorf("%s is not the mechanism and was matched, so the acknowledgement is too broad", className)
		}
	}
}

// TestClassOrderHarnessRefusesAnUnfiredControlEvenWithBothDirections isolates ControlsProven's
// per-control check from its direction check.
//
// Written because a mutation survived. Removing the `if !result.Detected { return false }` branch
// left every existing test green: they exercise it with controls whose expected outcome is
// Disagreed, so the direction check was already returning false and the mutated branch never decided
// anything. The two checks answer different questions and each needs a case where it is the only one
// that can say no.
//
// Here both silent directions are present, so the direction check is satisfied, and one of the two
// controls did not fire. That must still be refused: a harness whose engine-silent control is dead
// cannot see an invented reading, whatever its go-silent control managed.
func TestClassOrderHarnessRefusesAnUnfiredControlEvenWithBothDirections(t *testing.T) {
	t.Parallel()
	systems, readings := classOrderSyntheticPopulation(minimumPlausibleRegistryClasses+1, minimumPlausibleCorpusClasses+1, 5)
	honest := classOrderStubSource{name: "stub", readings: readings}

	firingGoSilent := Control{
		Name: "go-silent, fires", Detail: "declines everything",
		Corrupt: func(ReadingSource) ReadingSource {
			return classOrderStubSource{name: "silent", readings: map[string]Reading{}}
		},
		ExpectedOutcome: OutcomeGoSilent, MinimumDetections: 1,
	}
	// Same direction as a working engine-silent control and deliberately inert, so it cannot fire.
	deadEngineSilent := Control{
		Name: "engine-silent, dead", Detail: "corrupts nothing",
		Corrupt:         func(honest ReadingSource) ReadingSource { return honest },
		ExpectedOutcome: OutcomeEngineSilent, MinimumDetections: 1,
	}

	report := Compare("4.3.3", systems, classOrderHealthyDivergence(), honest,
		[]Control{firingGoSilent, deadEngineSilent})

	if !report.Controls[0].Detected {
		t.Fatal("the setup is wrong: the go-silent control did not fire")
	}
	if report.Controls[1].Detected {
		t.Fatal("the setup is wrong: the inert control reported firing")
	}
	if report.ControlsProven() {
		t.Error("both silent directions were declared and one control was dead, and detection was " +
			"still reported as proven")
	}
}

// TestClassOrderHarnessRefusesNoControlsWithNothingElseWrong pins the empty-controls behaviour.
//
// The guard it exercises is provably equivalent to the loop that follows it: with no controls both
// direction flags stay false and the conjunction returns false regardless. Mutation-testing found
// that and it is recorded on the guard itself rather than papered over, because a test that claims
// to cover a branch it cannot distinguish is the false-confidence this tier exists to refuse.
//
// The test is still worth having, and for the behaviour rather than the branch: a run that is sound
// in every other respect and simply carries no controls must be refused, and that is what a future
// refactor of ControlsProven could break without touching any line this file names.
func TestClassOrderHarnessRefusesNoControlsWithNothingElseWrong(t *testing.T) {
	t.Parallel()
	systems, readings := classOrderSyntheticPopulation(minimumPlausibleRegistryClasses+1, minimumPlausibleCorpusClasses+1, 5)

	report := Compare("4.3.3", systems, classOrderHealthyDivergence(),
		classOrderStubSource{name: "stub", readings: readings}, nil)

	// Everything else about this run is sound: the populations clear their floors, the engine
	// declined classes, and both sides agreed on every class they both answered.
	if population := report.Populations[0]; population.ByOutcome[OutcomeDisagreed] != 0 || population.EngineNullReadings == 0 {
		t.Fatalf("the setup is wrong: %d disagreements and %d engine nulls",
			population.ByOutcome[OutcomeDisagreed], population.EngineNullReadings)
	}
	if len(report.Controls) != 0 {
		t.Fatalf("the setup is wrong: %d controls", len(report.Controls))
	}
	if report.ControlsProven() {
		t.Error("a run carrying no controls reported that detection had been proven")
	}
	trustworthy, reasons := report.Trustworthy()
	if trustworthy {
		t.Error("a run that never demonstrated detection was reported as trustworthy")
	}
	if !classOrderAnyReasonContains(reasons, "no control fired in both silent directions") {
		t.Errorf("expected the unproven-controls reason, got %v", reasons)
	}
}

// TestClassOrderReadingEqualComparesBothAxes guards the comparison the whole tier rests on.
//
// Two utilities can declare the same properties and differ in arity, which is exactly what separates
// `slide-in-from-top-50` from `slide-in-from-top-4`. An Equal that compared only Order would report
// agreement on every such pair, and the shifted-count control exists because that axis is the one an
// order-only comparison cannot see.
func TestClassOrderReadingEqualComparesBothAxes(t *testing.T) {
	t.Parallel()
	base := Reading{Order: []int{315, 316}, Count: 2}

	if !base.Equal(Reading{Order: []int{315, 316}, Count: 2}) {
		t.Error("two identical readings compared unequal")
	}
	if base.Equal(Reading{Order: []int{315, 316}, Count: 3}) {
		t.Error("readings differing only in Count compared equal, so arity is invisible to the differential")
	}
	if base.Equal(Reading{Order: []int{315, 317}, Count: 2}) {
		t.Error("readings differing only in Order compared equal")
	}
	if base.Equal(Reading{Order: []int{315}, Count: 2}) {
		t.Error("readings of different order length compared equal")
	}
	// An empty order is meaningful rather than absent: a class that declares nothing sorts last, and
	// a Phase 0 control that got this backwards reported 365,174 spurious mismatches.
	empty := Reading{Order: nil, Count: 3}
	if !empty.Equal(Reading{Order: []int{}, Count: 3}) {
		t.Error("a nil order and an empty order compared unequal")
	}
	if empty.Equal(Reading{Order: nil, Count: 2}) {
		t.Error("empty-order readings differing in Count compared equal")
	}
}

// TestClassOrderHarnessRefusesAnUnknownPopulation proves a miscounted population is loud.
//
// The counting switch used to have a `default` that folded anything unrecognised into the registry.
// That is the silent-miscount shape the floors exist to catch, arriving through the code that feeds
// them: a run whose corpus was mislabelled would report a full registry, an empty corpus, and fail
// on the corpus floor while naming the wrong cause.
func TestClassOrderHarnessRefusesAnUnknownPopulation(t *testing.T) {
	t.Parallel()
	systems, readings := classOrderSyntheticPopulation(minimumPlausibleRegistryClasses+1, minimumPlausibleCorpusClasses+1, 5)
	systems[0].Cases = append(systems[0].Cases, ClassCase{
		ClassName:      "mislabelled",
		Population:     "registryy",
		EngineReading:  Reading{Order: []int{1}, Count: 1},
		EngineAnswered: true,
	})

	report := Compare("4.3.3", systems, classOrderHealthyDivergence(),
		classOrderStubSource{name: "stub", readings: readings}, nil)

	if unknown := report.Populations[0].UnknownPopulationClasses; unknown != 1 {
		t.Fatalf("expected 1 class in an unknown population, got %d", unknown)
	}
	if report.Populations[0].RegistryClasses != minimumPlausibleRegistryClasses+1 {
		t.Error("the mislabelled class was folded into the registry count")
	}
	_, reasons := report.Trustworthy()
	if !classOrderAnyReasonContains(reasons, "a population this package does not know") {
		t.Errorf("expected the unknown-population reason, got %v", reasons)
	}
}

func classOrderAnyReasonContains(reasons []string, needle string) bool {
	for _, reason := range reasons {
		if strings.Contains(reason, needle) {
			return true
		}
	}
	return false
}
