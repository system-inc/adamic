package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// The Tier 2 differential, run.
//
// `classorder_differential.go` is the instrument; this file loads the fixture, wires the live Go
// path into it, plants the controls, and asserts. Everything here is prefixed `classOrder` because
// three other components are being written into this package concurrently and two agents in one
// evening independently defined a helper called `describeCandidate`, which blocked the whole
// package's tests. A name collision in a test file is a compile error for everybody.

// classOrderFixture is the whole committed fixture.
type classOrderFixture struct {
	TailwindVersion string `json:"tailwindVersion"`
	GroundTruth     string `json:"groundTruth"`
	Divergence      struct {
		SharedClasses     int `json:"sharedClasses"`
		DivergentReadings int `json:"divergentReadings"`
		UniqueToOneSystem int `json:"uniqueToOneSystem"`
	} `json:"divergence"`
	Systems []classOrderSystem `json:"systems"`
}

type classOrderSystem struct {
	Name               string   `json:"name"`
	EntryPoint         string   `json:"entryPoint"`
	TailwindVersion    string   `json:"tailwindVersion"`
	CorpusRoot         string   `json:"corpusRoot"`
	CorpusNullExamples []string `json:"corpusNullExamples"`
	Counts             struct {
		RegistryClasses      int `json:"registryClasses"`
		CorpusOccurrences    int `json:"corpusOccurrences"`
		CorpusClasses        int `json:"corpusClasses"`
		RegistryNullReadings int `json:"registryNullReadings"`
		CorpusNullReadings   int `json:"corpusNullReadings"`
		DistinctReadings     int `json:"distinctReadings"`
	} `json:"counts"`
	Registration struct {
		Prefix       string              `json:"prefix"`
		UtilityRoots map[string][]string `json:"utilityRoots"`
		VariantRoots map[string]struct {
			Kind          string `json:"kind"`
			Compounds     int    `json:"compounds"`
			CompoundsWith int    `json:"compoundsWith"`
		} `json:"variantRoots"`
	} `json:"registration"`
	Readings []struct {
		Order []int `json:"order"`
		Count int   `json:"count"`
	} `json:"readings"`
	RegistryCases []classOrderCase `json:"registryCases"`
	CorpusCases   []classOrderCase `json:"corpusCases"`
	// DescriptorTable and Context are the two per-repository halves of the lookup table, captured by
	// the extractors this fixture's generator orchestrates.
	DescriptorTable extractedTableFile   `json:"descriptorTable"`
	Context         extractedContextFile `json:"context"`
}

// classOrderCase is one class and an index into the system's interned readings.
//
// `readingIndex` of -1 is the engine declining, and it is deliberately a value in the same column as
// every other answer rather than an absent field. This harness exists to refuse to score silence as
// agreement, and spelling silence as missing data would be a poor start.
type classOrderCase struct {
	ClassName    string `json:"className"`
	ReadingIndex int    `json:"readingIndex"`
}

var (
	classOrderFixtureOnce  sync.Once
	classOrderFixtureValue *classOrderFixture
	classOrderFixtureError error
)

func classOrderLoadFixture(t *testing.T) *classOrderFixture {
	t.Helper()
	classOrderFixtureOnce.Do(func() {
		path := filepath.Join("testdata", "classorder_fixtures.json")
		contents, err := os.ReadFile(path)
		if err != nil {
			classOrderFixtureError = err
			return
		}
		classOrderFixtureValue = &classOrderFixture{}
		if err := json.Unmarshal(contents, classOrderFixtureValue); err != nil {
			classOrderFixtureError = fmt.Errorf("parse %s: %w", path, err)
		}
	})
	if classOrderFixtureError != nil {
		t.Fatalf("loading the class-order fixture: %v", classOrderFixtureError)
	}
	return classOrderFixtureValue
}

// classOrderCases turns one fixture system into the harness's population.
func classOrderCases(system *classOrderSystem) SystemCases {
	cases := make([]ClassCase, 0, len(system.RegistryCases)+len(system.CorpusCases))
	appendCases := func(source []classOrderCase, population string) {
		for _, entry := range source {
			aCase := ClassCase{ClassName: entry.ClassName, Population: population}
			if entry.ReadingIndex >= 0 {
				reading := system.Readings[entry.ReadingIndex]
				aCase.EngineReading = Reading{Order: reading.Order, Count: reading.Count}
				aCase.EngineAnswered = true
			}
			cases = append(cases, aCase)
		}
	}
	appendCases(system.RegistryCases, PopulationRegistry)
	appendCases(system.CorpusCases, PopulationCorpus)
	return SystemCases{SystemName: system.Name, Cases: cases}
}

// classOrderDesignSystem answers the parser's four questions from the tables the engine reported.
//
// It carries its own registration tables rather than borrowing `candidate_fixtures.json`, which
// belongs to the candidate parser's suite and whose shape is that author's to change. A harness
// whose population silently depends on another component's fixture breaks when that component is
// refactored, and the break arrives disguised as a differential finding.
type classOrderDesignSystem struct {
	system *classOrderSystem
}

func (designSystem *classOrderDesignSystem) Prefix() string {
	return designSystem.system.Registration.Prefix
}

func (designSystem *classOrderDesignSystem) HasUtility(root string, kind UtilityKind) bool {
	for _, declared := range designSystem.system.Registration.UtilityRoots[root] {
		if declared == string(kind) {
			return true
		}
	}
	return false
}

func (designSystem *classOrderDesignSystem) HasVariant(root string) bool {
	_, exists := designSystem.system.Registration.VariantRoots[root]
	return exists
}

func (designSystem *classOrderDesignSystem) VariantKind(root string) ParsedVariantKind {
	return ParsedVariantKind(designSystem.system.Registration.VariantRoots[root].Kind)
}

// VariantCompoundsWith reproduces upstream's `Variants.compoundsWith`, over this fixture's own
// bitmasks.
//
// Both masks are carried and both are consulted: `compounds` is what a variant produces and
// `compoundsWith` is what it accepts, and they differ on the same variant (`not` is 2 and 3).
// Inferring one from the other would make every compound-variant parse agree for the wrong reason.
func (designSystem *classOrderDesignSystem) VariantCompoundsWith(parent string, child ParsedVariant) bool {
	parentInfo, parentExists := designSystem.system.Registration.VariantRoots[parent]
	if !parentExists {
		return false
	}

	var childCompounds int
	if child.Kind == ParsedVariantKindArbitrary {
		childCompounds = int(compoundsForSelectors([]string{child.Selector}))
	} else {
		childInfo, childExists := designSystem.system.Registration.VariantRoots[child.Root]
		if !childExists {
			return false
		}
		childCompounds = childInfo.Compounds
	}

	return parentInfo.CompoundsWith&childCompounds != 0
}

// classOrderLiveSource is the real Go path: a class name in, a reading out.
//
// This is the seam under test and the reason the whole tier exists. `ParseCandidate` turns the
// string into candidates, the first is taken (which is what `getClassOrder` does at `sort.ts:22`),
// and `Table.Lookup` reads it. Nothing here decodes anything on the fixture's behalf.
type classOrderLiveSource struct {
	name         string
	table        *Table
	designSystem DesignSystem
}

func (source *classOrderLiveSource) Name() string { return source.name }

func (source *classOrderLiveSource) Reading(className string) (Reading, bool) {
	// The variant prefix is stripped before lookup, not before parsing.
	//
	// `md:flex` and `flex` have the same reading, since variants change the selector and the variant
	// bitmask rather than the declarations, and the parser must still see the variants: a class
	// whose variant does not parse is not a class. Parsing the whole string and reading the utility
	// off the parsed candidate is what keeps both true at once.
	candidates := ParseCandidate(className, source.designSystem)
	if len(candidates) == 0 {
		return Reading{}, false
	}
	return source.table.Lookup(&candidates[0])
}

// classOrderTable builds one design system's lookup table from its half of the fixture.
//
// It reuses `extractedTableFile`, `extractedContextFile` and the four reading helpers that
// `descriptor_table_test.go` declares at package scope, so there is one definition of how a measured
// bucket becomes a `Reading` and this file cannot drift from it. What it does not reuse is
// `loadTestTable`, which reads two fixed paths off disk and therefore only ever builds ahra's table.
// This tier needs one table per design system out of one fixture, so the assembly is here and the
// decoding is shared.
func classOrderTable(t *testing.T, system *classOrderSystem) *Table {
	t.Helper()

	tableFile, contextFile := &system.DescriptorTable, &system.Context

	if len(tableFile.Roots) == 0 || len(contextFile.Namespaces) == 0 {
		// An empty half builds a table that declines everything, which reads downstream as a port
		// that stopped answering rather than as a fixture that lost a field.
		t.Fatalf("%s: the fixture carries %d descriptor roots and %d namespaces; a table cannot be built from that",
			system.Name, len(tableFile.Roots), len(contextFile.Namespaces))
	}

	perDeclaration := make(map[string]bool, len(contextFile.PerDeclarationRoots))
	for _, root := range contextFile.PerDeclarationRoots {
		perDeclaration[root] = true
	}

	table := &Table{
		TailwindVersion: tableFile.TailwindVersion,
		Descriptors:     make(map[string]*Descriptor, len(tableFile.Roots)),
		Statics:         make(map[string]Reading, len(tableFile.Statics)),
		Namespaces:      contextFile.Namespaces,
		KeysByNamespace: make(map[string]map[string]bool, len(contextFile.KeysByNamespace)),
		PropertyOrder:   make(map[string]int, len(contextFile.PropertyOrder)),
	}

	for namespace, keys := range contextFile.KeysByNamespace {
		set := make(map[string]bool, len(keys))
		for _, key := range keys {
			set[key] = true
		}
		table.KeysByNamespace[namespace] = set
	}
	for index, property := range contextFile.PropertyOrder {
		table.PropertyOrder[property] = index
	}
	// The unlisted statics first, so the registry's own measurement wins any collision. `order-none`
	// is the one on this design system: a class an author can write that `getClassList()` never
	// advertised, so a table without it declines a class the engine reads.
	for name, value := range contextFile.UnlistedStatics {
		if value == nil {
			continue
		}
		table.Statics[name] = Reading{Order: value.Order, Count: value.Count}
	}
	for name, value := range tableFile.Statics {
		if value == nil {
			continue
		}
		table.Statics[name] = Reading{Order: value.Order, Count: value.Count}
	}

	for index := range tableFile.Roots {
		entry := &tableFile.Roots[index]
		descriptor := &Descriptor{
			Root:           entry.Root,
			PerDeclaration: perDeclaration[entry.Root],
			ByLiteral:      readingMap(entry.ReadingByLiteral),
			Absent: AxisReadings{
				ByType:      typeReadingMap(entry.ReadingByType, entry.Fallback),
				ByNamespace: readingMap(entry.ReadingByNamespace),
				Fallback:    readingOrZero(entry.Fallback),
				Empty:       readingPointer(entry.Empty),
			},
			Alpha: AxisReadings{
				ByType:      typeReadingMap(entry.ReadingByTypeWithModifier, entry.FallbackWithModifier),
				ByNamespace: readingMap(entry.ReadingByNamespaceWithModifier),
				Fallback:    readingOrZero(entry.FallbackWithModifier),
				Empty:       readingPointer(entry.EmptyWithModifier),
			},
			Themed: AxisReadings{
				ByType:      typeReadingMap(entry.ReadingByTypeWithThemedModifier, entry.FallbackWithThemedModifier),
				ByNamespace: readingMap(entry.ReadingByNamespaceWithThemedModifier),
				Fallback:    readingOrZero(entry.FallbackWithThemedModifier),
				Empty:       readingPointer(entry.EmptyWithThemedModifier),
			},
		}
		// The order is carried verbatim, never sorted. See TestTypeListOrderIsCarriedNotSorted.
		for _, typeName := range entry.TypeList {
			descriptor.TypeList = append(descriptor.TypeList, DataType(typeName))
		}
		table.Descriptors[entry.Root] = descriptor
	}

	return table
}

// The two mechanisms this tier found on its first run, and the tasks that own them.
//
// They are written as *predicates over the fixture* rather than as a list of class names. Both were
// reduced to a reproducible probe against the engine before being written down, and a predicate says
// which mechanism a class falls under, so a class that stops matching stops being excused and a
// class that starts matching is covered without anyone editing a list. A hand-written list would
// drift silently the moment a repository added a shadow token.
//
// Each is a finding handed onward, not a dispensation. Both are bounded below by an expected count,
// so an acknowledgement that quietly stopped applying fails the run rather than shrinking it.

// classOrderUnresolvableAlphaRoots are the roots whose alpha modifier gains a declaration when the
// modifier cannot be resolved at build time.
//
// # The mechanism
//
// Phase 0 measured the modifier as three states and recorded that its *value* is otherwise not
// consulted, so `/25`, `/[0.5]` and `/[var(--a)]` are one bucket. That holds on 28,963 of 28,986
// probed registry classes and fails on 23, all in the shadow families, identically on both design
// systems — which is what says the mechanism belongs to Tailwind rather than to either repository.
//
// It is visible in the emitted CSS. A statically resolvable alpha lets the engine fold the colour at
// build time into one declaration:
//
//	shadow-lg/[0.5]      --tw-shadow: ... oklab(from rgb(0 0 0 / 0.1) l a b / 50%) ...
//
// An alpha holding a runtime `var()` cannot be folded, so the engine emits a literal fallback *and*
// an `@supports (color: lab(from red l a b))` block carrying the relative-colour form:
//
//	shadow-lg/[var(--a)] --tw-shadow: ... rgb(0 0 0 / 0.1) ...
//	                     @supports ... { --tw-shadow: ... oklab(from ... / var(--a)) ... }
//
// That second declaration is the difference between `#3` and `#4`. The discriminator is resolvability
// and not the `var` token: `/[--x]` and `/[calc(1/2)]` both read `#3`. So this is a fourth modifier
// state, and modelling it is the descriptor table's work rather than this harness's.
var classOrderUnresolvableAlphaRoots = []string{"shadow", "inset-shadow", "text-shadow", "drop-shadow"}

// classOrderUnresolvableAlpha reports whether a class is the fourth-modifier-state mechanism.
//
// The modifier must be arbitrary and must contain a `var()` reference, which is the only construct
// among the probed shapes that the engine cannot fold at build time.
func classOrderUnresolvableAlpha(className string) bool {
	slash := strings.LastIndex(className, "/[")
	if slash < 0 || !strings.HasSuffix(className, "]") {
		return false
	}
	modifier := className[slash+2 : len(className)-1]
	if !strings.Contains(modifier, "var(") {
		return false
	}
	root := className[:slash]
	if colon := strings.LastIndexByte(root, ':'); colon >= 0 {
		root = root[colon+1:]
	}
	for _, prefix := range classOrderUnresolvableAlphaRoots {
		if root == prefix || strings.HasPrefix(root, prefix+"-") {
			return true
		}
	}
	return false
}

// classOrderExpectedUnresolvableAlpha is how many classes the mechanism must cover per design system.
//
// A floor rather than an exact count, so a repository adding a shadow token does not fail the suite,
// and high enough that an acknowledgement which stopped applying cannot pass as covered.
const classOrderExpectedUnresolvableAlpha = 15

// classOrderExpectedInventedReadings is how many classes the second mechanism must cover.
//
// # The mechanism
//
// `Table.Lookup` maps a bare value through the namespace precedence and falls through to a fallback
// reading when nothing claims it. The engine instead emits no declarations at all, because a utility
// whose value resolved to nothing has nothing to declare. So Go answers and the engine declines.
//
// Phase 0 could not have caught this. It measured nulls on the corpus, reported them separately and
// never scored them, which is correct for a component fixture and means the claim "the model does not
// invent readings for classes the engine declines" was asserted nowhere. This tier asserts it, which
// is the reason the tier exists above the per-component one.
//
// Six registry classes on both systems, every one the colour-keyword-plus-alpha shape
// (`shadow-inherit/50` puts an alpha on a keyword that cannot carry one), plus corpus classes that
// differ per repository because they are words the repository writes that no namespace holds:
// `text-ss` and `text-dark-4/70` on ahra. Ten on ahra and nine on connected.
//
// Bounded above rather than below, because this is the direction where a growing number means the
// port has started inventing readings, which is the failure nothing else in the suite watches for.
const classOrderExpectedInventedReadings = 12

// classOrderControls are the deliberate defects the harness must detect.
//
// Requirement four: prove the check can fail before trusting that it passed. Each control corrupts
// the Go side in a different direction and names the bucket it must land classes in, because a
// control that only had to "make something fail" would let one direction of detection stand in for
// the other. A harness that can see wrong readings and cannot see invented ones reports a clean run
// for every class the Go side answers and the engine does not.
//
// The floors are floors and not exact counts, so the registry can grow without anyone re-tuning a
// control. Each is far enough above zero that a corruption which changed nothing cannot pass, and
// far enough below the measured detection count that a token landing in a theme cannot fail it.
func classOrderControls() []Control {
	return []Control{
		{
			Name:   "shifted-order",
			Detail: "every reading's first order index shifted by one, which is a wrong answer rather than no answer",
			Corrupt: func(honest ReadingSource) ReadingSource {
				return classOrderCorruptedSource{honest: honest, transform: func(reading Reading, answered bool) (Reading, bool) {
					if !answered || len(reading.Order) == 0 {
						return reading, answered
					}
					shifted := make([]int, len(reading.Order))
					copy(shifted, reading.Order)
					shifted[0]++
					return Reading{Order: shifted, Count: reading.Count}, true
				}}
			},
			ExpectedOutcome:   OutcomeDisagreed,
			MinimumDetections: 10000,
		},
		{
			Name:   "shifted-count",
			Detail: "every reading's declaration count incremented, which is the axis an order-only comparison cannot see",
			Corrupt: func(honest ReadingSource) ReadingSource {
				return classOrderCorruptedSource{honest: honest, transform: func(reading Reading, answered bool) (Reading, bool) {
					if !answered {
						return reading, answered
					}
					return Reading{Order: reading.Order, Count: reading.Count + 1}, true
				}}
			},
			ExpectedOutcome:   OutcomeDisagreed,
			MinimumDetections: 10000,
		},
		{
			Name:   "silent-side",
			Detail: "the Go side declines every class, which is the failure that reads as a clean run when silence is scored as agreement",
			Corrupt: func(honest ReadingSource) ReadingSource {
				return classOrderCorruptedSource{honest: honest, transform: func(Reading, bool) (Reading, bool) {
					return Reading{}, false
				}}
			},
			ExpectedOutcome:   OutcomeGoSilent,
			MinimumDetections: 10000,
		},
		{
			Name:   "invented-reading",
			Detail: "the Go side answers where the engine declined, which no count of agreements would surface",
			Corrupt: func(honest ReadingSource) ReadingSource {
				return classOrderCorruptedSource{honest: honest, transform: func(reading Reading, answered bool) (Reading, bool) {
					if answered {
						return reading, true
					}
					// A plausible-looking reading, which is the point: an invented answer does not
					// announce itself, and the engine's silence is the only thing that contradicts it.
					return Reading{Order: []int{1}, Count: 1}, true
				}}
			},
			ExpectedOutcome:   OutcomeEngineSilent,
			MinimumDetections: 40,
		},
	}
}

// classOrderCorruptedSource wraps an honest source and bends its answers.
type classOrderCorruptedSource struct {
	honest    ReadingSource
	transform func(reading Reading, answered bool) (Reading, bool)
}

func (source classOrderCorruptedSource) Name() string { return source.honest.Name() + " (corrupted)" }

func (source classOrderCorruptedSource) Reading(className string) (Reading, bool) {
	return source.transform(source.honest.Reading(className))
}

// TestClassOrderDifferential is the Tier 2 run: every class, both design systems, end to end.
func TestClassOrderDifferential(t *testing.T) {
	t.Parallel()
	fixture := classOrderLoadFixture(t)

	systems := make([]SystemCases, 0, len(fixture.Systems))
	sources := make([]*classOrderLiveSource, 0, len(fixture.Systems))
	for index := range fixture.Systems {
		system := &fixture.Systems[index]
		systems = append(systems, classOrderCases(system))
		sources = append(sources, &classOrderLiveSource{
			name:         "ParseCandidate + Table.Lookup",
			table:        classOrderTable(t, system),
			designSystem: &classOrderDesignSystem{system: system},
		})
	}

	// One design system at a time, because the table is per repository. Comparing ahra's classes
	// against connected's table is not a differential, it is the defect this port exists to fix,
	// run deliberately.
	for index := range systems {
		report := Compare(
			fixture.TailwindVersion,
			systems[index:index+1],
			SystemDivergence{
				SharedRegistryClasses: fixture.Divergence.SharedClasses,
				DivergentReadings:     fixture.Divergence.DivergentReadings,
				UniqueToOneSystem:     fixture.Divergence.UniqueToOneSystem,
			},
			sources[index],
			classOrderControls(),
		)
		classOrderAssertReport(t, fixture, sources[index], report)
	}
}

// classOrderAssertReport is every assertion one run has to pass.
func classOrderAssertReport(t *testing.T, fixture *classOrderFixture, source *classOrderLiveSource, report Report) {
	t.Helper()

	population := report.Populations[0]

	// Trustworthiness first, and separately from what was found. Every reason it can return
	// describes a run whose agreement numbers mean nothing, and reading a rate off an untrustworthy
	// run is the failure mode the whole file is built against.
	//
	// The single-system floor is checked here rather than through Trustworthy's own two-system rule,
	// because this call deliberately passes one system at a time; the two-system requirement is
	// asserted over the fixture in TestClassOrderDifferentialRequiresTwoDesignSystems.
	for _, reason := range classOrderUntrustworthyReasons(report) {
		t.Errorf("%s: untrustworthy run: %s", population.SystemName, reason)
	}

	system := classOrderSystemNamed(fixture, population.SystemName)
	grouped := report.DivergencesByOutcome()

	// A wrong reading is a defect unless it is the fourth-modifier-state mechanism, which is a
	// finding this tier made and handed to the descriptor table rather than a defect in this file.
	//
	// The split is by predicate and both halves are asserted. An unexplained disagreement fails, and
	// so does an explained one that stopped appearing: an acknowledgement covering nothing is an
	// acknowledgement that has quietly become a suppression of some other class.
	var unexplainedDisagreements, unresolvableAlpha []Divergence
	for _, divergence := range grouped[OutcomeDisagreed] {
		if classOrderUnresolvableAlpha(divergence.ClassName) {
			unresolvableAlpha = append(unresolvableAlpha, divergence)
			continue
		}
		unexplainedDisagreements = append(unexplainedDisagreements, divergence)
	}

	if len(unexplainedDisagreements) > 0 {
		t.Errorf("%s: %d classes read differently in Go and the engine, under no known mechanism",
			population.SystemName, len(unexplainedDisagreements))
		for _, divergence := range classOrderSample(unexplainedDisagreements, 20) {
			t.Errorf("  %s (%s): engine %s, go %s",
				divergence.ClassName, divergence.Population,
				divergence.EngineReading.Format(), divergence.GoReading.Format())
		}
	}
	if len(unresolvableAlpha) < classOrderExpectedUnresolvableAlpha {
		t.Errorf("%s: the unresolvable-alpha mechanism covered %d classes, below the expected %d; "+
			"an acknowledgement that covers nothing is excusing something else",
			population.SystemName, len(unresolvableAlpha), classOrderExpectedUnresolvableAlpha)
	}

	// An invented reading is the direction nothing else in the suite looks at. The engine declining
	// a class is the only evidence that contradicts a Go answer, so an unchecked engine-silent
	// bucket means the port can invent readings freely.
	//
	// Bounded above rather than forbidden, because the mechanism is understood and belongs to the
	// `@utility` evaluator: the model has no membership test for "does this value resolve at all".
	// A ceiling fails the moment the port starts inventing more than the mechanism accounts for.
	if invented := grouped[OutcomeEngineSilent]; len(invented) > classOrderExpectedInventedReadings {
		t.Errorf("%s: %d classes where Go answered and the engine declined, above the %d the known mechanism accounts for",
			population.SystemName, len(invented), classOrderExpectedInventedReadings)
		for _, divergence := range classOrderSample(invented, 20) {
			t.Errorf("  %s (%s): engine declined, go %s",
				divergence.ClassName, divergence.Population, divergence.GoReading.Format())
		}
	}

	// Go-side silence is split by mechanism rather than bounded by a rate.
	//
	// A rate ceiling was the first design here and it was wrong in a way worth recording: 407 of
	// ahra's 408 declines are the model's own declared boundary, so any ceiling loose enough to
	// admit them is loose enough to hide the 408th. Connected has exactly one such class,
	// `max-w-screen-sm`, and a percentage would have absorbed it without a word.
	classOrderAssertGoSilence(t, system, source, population, grouped[OutcomeGoSilent])

	// The coverage assertion, stated as its own requirement: a side reporting nothing is a failed
	// run and not a clean one. Trustworthy already refuses a run over the both-silent ceiling; this
	// asserts the positive, that the run actually answered the bulk of what it was given.
	if answered := population.Answered(); answered < population.Total()/2 {
		t.Errorf("%s: only %d of %d classes were answered by both sides; a differential that mostly says nothing is not a clean run",
			population.SystemName, answered, population.Total())
	}

	// The controls. Proving the check can fail is what makes the passes above mean anything, and a
	// control that did not fire invalidates the run rather than merely noting a gap.
	if !report.ControlsProven() {
		t.Errorf("%s: controls were not proven in both silent directions; detection has not been demonstrated", population.SystemName)
	}
	for _, result := range report.Controls {
		if result.Detected {
			continue
		}
		t.Errorf("%s: control %q did not fire: %s", population.SystemName, result.Control.Name, result.Detail)
	}

	// The fixture's own counts, asserted against what the run actually walked. A fixture that
	// silently lost a population is the failure this catches, and it is invisible in a rate.
	if population.RegistryClasses != system.Counts.RegistryClasses {
		t.Errorf("%s: walked %d registry classes, fixture recorded %d",
			population.SystemName, population.RegistryClasses, system.Counts.RegistryClasses)
	}
	if population.CorpusClasses != system.Counts.CorpusClasses {
		t.Errorf("%s: walked %d corpus classes, fixture recorded %d",
			population.SystemName, population.CorpusClasses, system.Counts.CorpusClasses)
	}

	// Printed on success, because a suite that compared twelve answers and one that compared a
	// hundred and ninety thousand look the same as a green line.
	t.Log("\n" + report.Describe())
}

// classOrderUntrustworthyReasons runs Trustworthy's checks minus the two-system rule.
//
// Factored out rather than parameterised into Trustworthy, because the two-system requirement is
// real and belongs in Trustworthy: a caller handing it one system should be told. This test hands it
// one system deliberately, since the tables are per repository, and asserts the two-system
// requirement over the fixture instead. Suppressing that one reason here and nowhere else keeps
// Trustworthy honest for every other caller.
func classOrderUntrustworthyReasons(report Report) []string {
	_, reasons := report.Trustworthy()
	kept := reasons[:0]
	for _, reason := range reasons {
		if strings.Contains(reason, "design systems measured") {
			continue
		}
		kept = append(kept, reason)
	}
	return kept
}

// classOrderAssertGoSilence checks that every class the Go side declined was declined on purpose.
//
// The model has exactly one declared reason to decline a functional class: `PerDeclaration`, the 18
// repository `@utility` roots whose arity is the count of surviving declarations rather than a
// function of any data type. `Lookup` returns false for those deliberately, because no row of the
// table can answer them and a plausible reading would be wrong in the one place the model knows its
// own boundary. Those are expected and counted.
//
// Anything else is a table gap, and this is the assertion that finds them. Measured on the first run:
// 407 of ahra's 408 declines are per-declaration roots, and connected has one more,
// `max-w-screen-sm`. `max-w-screen` is registered as both a static and a functional utility, the
// parser correctly reads the class as functional rooted there, and the descriptor table carries no
// functional row for it — it recorded the root as a static and stopped. Thirteen further roots are
// registered functional with no descriptor row and produce no visible defect yet only because their
// registry classes happen to parse to other roots.
//
// That class of defect is reachable only from a class *name*. A fixture that hands the Go side a
// pre-parsed candidate has already decided the root, so it can never disagree with the parser about
// which root a class has. This is the seam, and this is what testing it finds.
func classOrderAssertGoSilence(t *testing.T, system *classOrderSystem, source *classOrderLiveSource, population Population, silent []Divergence) {
	t.Helper()

	perDeclaration := make(map[string]bool, len(system.Context.PerDeclarationRoots))
	for _, root := range system.Context.PerDeclarationRoots {
		perDeclaration[root] = true
	}
	if len(perDeclaration) == 0 {
		t.Errorf("%s: the fixture names no per-declaration roots, so every decline below is unexplained for the wrong reason",
			population.SystemName)
	}

	byRoot := map[string]int{}
	example := map[string]string{}
	declared, acknowledged := 0, 0
	for _, divergence := range silent {
		root := classOrderRootOf(source, divergence.ClassName)
		if perDeclaration[root] {
			declared++
			continue
		}
		if classOrderMissingDescriptorRoots[root] {
			acknowledged++
			continue
		}
		byRoot[root]++
		if _, seen := example[root]; !seen {
			example[root] = divergence.ClassName
		}
	}

	// Reported so the excused count is visible next to the unexcused one. Whether the
	// acknowledgement is still earning its place is asserted once across both systems, in
	// TestClassOrderAcknowledgementsAreStillEarned, because `max-w-screen` is advertised by only one
	// of the two registries and a per-system assertion would fail on the other for being right.
	if acknowledged > 0 {
		t.Logf("%s: %d declines covered by a known missing descriptor row", population.SystemName, acknowledged)
	}

	if declared < classOrderExpectedPerDeclarationDeclines {
		// The declared boundary vanishing is as much a finding as an undeclared decline appearing:
		// it means either the fixture lost the `@utility` roots or the model started answering
		// classes it has no basis to answer.
		t.Errorf("%s: only %d classes were declined under the %d per-declaration roots, below the expected %d",
			population.SystemName, declared, len(perDeclaration), classOrderExpectedPerDeclarationDeclines)
	}

	if len(byRoot) == 0 {
		return
	}

	counts := make([]classOrderRootCount, 0, len(byRoot))
	for root, count := range byRoot {
		counts = append(counts, classOrderRootCount{root: root, count: count})
	}
	sort.Slice(counts, func(left, right int) bool {
		if counts[left].count != counts[right].count {
			return counts[left].count > counts[right].count
		}
		return counts[left].root < counts[right].root
	})

	// The roots are named rather than the classes counted, because "declined 408 classes" is not
	// actionable and "declined every class rooted at `max-w-screen`, which has no descriptor row" is.
	t.Errorf("%s: %d classes were declined under %d roots that are not per-declaration, so the table is missing rows",
		population.SystemName, len(silent)-declared, len(counts))
	for index, entry := range counts {
		if index >= 15 {
			break
		}
		t.Errorf("  %d declined under root %q, for example %s", entry.count, entry.root, example[entry.root])
	}
}

// classOrderRootCount is one root and how many of its classes the Go side declined.
type classOrderRootCount struct {
	root  string
	count int
}

// classOrderMissingDescriptorRoots are roots the parser reads and the descriptor table cannot answer.
//
// A finding this tier made, handed to the descriptor table's owner rather than fixed here: this file
// is the instrument, and an instrument that repairs what it measures stops being one.
//
// `max-w-screen` is registered as both a static and a functional utility. The parser correctly reads
// `max-w-screen-sm` as functional rooted there; the extractor recorded the root as a static and
// wrote no functional row, so `Lookup` declines a class the engine compiles to `[46]#1`. One class
// on www-connected-app and none on ahra, because ahra's registry does not advertise it.
//
// Thirteen further roots are registered functional with no descriptor row and produce no visible
// decline yet, only because their registry classes happen to parse to other roots:
// `-backdrop-hue-rotate`, `-col`, `-hue-rotate`, `-row`, `backdrop-filter`, `bg-position`,
// `bg-size`, `filter`, `flex-grow`, `flex-shrink`, `font-features`, `mask-position` and
// `max-w-screen`. They are latent rather than acknowledged: this list holds only what actually
// declined, so a class that starts exercising one of the other twelve fails the run.
//
// Named per root rather than waved through by a rate, because that is the whole point. A percentage
// ceiling loose enough to admit the 407 declines the model makes on purpose is loose enough to hide
// this one, which is how the first version of this assertion was written and why it is not this one.
var classOrderMissingDescriptorRoots = map[string]bool{
	"max-w-screen": true,
}

// classOrderExpectedPerDeclarationDeclines is the floor for declines the model makes on purpose.
//
// The 18 per-declaration roots carry 22 or 23 registry classes each, so the measured figure is 407
// on ahra and 408 on connected. A floor well under that survives a repository dropping an animation
// utility and still fails if the boundary stops applying entirely.
const classOrderExpectedPerDeclarationDeclines = 300

// classOrderRootOf recovers the utility root a class was declined under, by parsing it.
//
// The root cannot be recovered from the class name alone and the first attempt here proved it:
// stripping the final segment reads `fade-in-translate-full` as rooted at `fade-in-translate`, when
// the engine and the parser both root it at `fade-in`, a per-declaration root. That mis-grouping
// turned 18 correctly-declined classes into 18 phantom "missing table rows", which is precisely the
// kind of confident wrong answer this whole tier exists to catch — arriving, fittingly, in the
// harness itself.
//
// So the parser is asked. It is the component that owns the question, it is already wired into this
// test, and a class it cannot parse is reported as unparsed rather than guessed at.
func classOrderRootOf(source *classOrderLiveSource, className string) string {
	candidates := ParseCandidate(className, source.designSystem)
	if len(candidates) == 0 {
		return "(unparsed) " + className
	}
	if candidates[0].Root == "" {
		return "(" + string(candidates[0].Kind) + ") " + className
	}
	return candidates[0].Root
}

func classOrderSample(divergences []Divergence, limit int) []Divergence {
	if len(divergences) <= limit {
		return divergences
	}
	return divergences[:limit]
}

func classOrderSystemNamed(fixture *classOrderFixture, name string) *classOrderSystem {
	for index := range fixture.Systems {
		if fixture.Systems[index].Name == name {
			return &fixture.Systems[index]
		}
	}
	return &classOrderSystem{}
}

// TestClassOrderAcknowledgementsAreStillEarned checks that every excuse still covers something.
//
// An acknowledgement that has stopped applying is worse than one that never existed: it stays in the
// file as a blanket excuse for whatever lands under that root next, and it reads as diligence. So
// each is required to be exercised by at least one design system.
//
// Asserted across both systems rather than within one, because coverage is legitimately per
// repository: `max-w-screen` is advertised by www-connected-app's registry and not by ahra's, so
// demanding it fire on ahra would fail the suite for being correct.
func TestClassOrderAcknowledgementsAreStillEarned(t *testing.T) {
	t.Parallel()
	fixture := classOrderLoadFixture(t)

	exercised := map[string]int{}
	unresolvableAlpha := 0
	invented := 0

	for index := range fixture.Systems {
		system := &fixture.Systems[index]
		source := &classOrderLiveSource{
			name:         "ParseCandidate + Table.Lookup",
			table:        classOrderTable(t, system),
			designSystem: &classOrderDesignSystem{system: system},
		}
		perDeclaration := make(map[string]bool, len(system.Context.PerDeclarationRoots))
		for _, root := range system.Context.PerDeclarationRoots {
			perDeclaration[root] = true
		}

		for _, aCase := range classOrderCases(system).Cases {
			goReading, goAnswered := source.Reading(aCase.ClassName)
			switch classify(aCase, goReading, goAnswered) {
			case OutcomeDisagreed:
				if classOrderUnresolvableAlpha(aCase.ClassName) {
					unresolvableAlpha++
				}
			case OutcomeEngineSilent:
				invented++
			case OutcomeGoSilent:
				root := classOrderRootOf(source, aCase.ClassName)
				if perDeclaration[root] {
					continue
				}
				if classOrderMissingDescriptorRoots[root] {
					exercised[root]++
				}
			}
		}
	}

	for root := range classOrderMissingDescriptorRoots {
		if exercised[root] == 0 {
			t.Errorf("no design system exercised the acknowledged missing descriptor row for %q; "+
				"an acknowledgement covering nothing is an excuse waiting for the next defect under that root", root)
		}
	}
	if unresolvableAlpha == 0 {
		t.Error("no design system exercised the unresolvable-alpha mechanism; the acknowledgement no longer covers anything")
	}
	if invented == 0 {
		t.Error("no design system exercised the invented-reading mechanism; the acknowledgement no longer covers anything")
	}

	t.Logf("acknowledgements exercised across %d design systems: %d unresolvable-alpha, %d invented readings, %v missing descriptor rows",
		len(fixture.Systems), unresolvableAlpha, invented, exercised)
}
