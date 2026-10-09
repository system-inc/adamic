package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind/vendored"
)

// The fixture is what the shipped Tailwind 4.3.3 engine did with this repository's own `@utility`
// blocks, captured by internal/lint/rules/tailwind/tools/generate_utility and checked in next to this test.
//
// The population is chosen so that a clean result is evidence rather than a coincidence. The
// registry half is the one Phase 0 measured, so the 18 known exceptions are covered exactly as they
// were reported. The sweep half reaches what `getClassList()` does not advertise and what the
// registry therefore cannot test: fractions, which drive the ratio splice; decimals, where `1.25`
// survives and `1.3` drops on the same declaration; arbitrary values with and without a typehint;
// and modifier forms, which are what the four post-conditions gate.
//
// The `@utility` blocks are carried as CSS text rather than as a pre-parsed tree, and the Go side
// parses them with cssparser.go. That composes the two components the way production will compose
// them: a fixture holding a tree the Go parser might never produce would test this evaluator
// against an input it never sees.

type utilityCorpus struct {
	TailwindVersion string `json:"tailwindVersion"`
	// EntryPath is the stylesheet the design system was loaded from, so the Go side resolves the
	// same theme from the same `@import` graph.
	EntryPath string `json:"entryPoint"`
	// UtilityBlocks is every `@utility` block in the graph, as CSS text.
	UtilityBlocks []utilityBlockFixture `json:"utilityBlocks"`
	// PerDeclarationRoots is the set of roots the engine resolves per declaration, measured rather
	// than named. See TestUtilityFindsThePerDeclarationRoots.
	PerDeclarationRoots []string `json:"perDeclarationRoots"`
	// KnownExceptions is the 18 registry classes the descriptor model mispredicts, as reported by
	// generate_descriptors. This is the acceptance criterion for this component.
	KnownExceptions []utilityExceptionFixture `json:"knownExceptions"`
	// ShadowQuirkProbes is the other exception, recorded and not modelled: four sweep probes hitting
	// an upstream shadow quirk on `[16/9]`. No `@utility` block is involved and no registry contains
	// them. See TestUtilityDoesNotClaimTheShadowQuirk.
	ShadowQuirkProbes []utilityQuirkFixture `json:"shadowQuirkProbes"`
	// SyntheticCases are standalone design systems whose `@utility` blocks reach the branches this
	// repository's own blocks never do. See TestUtilitySyntheticCasesMatchEngine.
	SyntheticCases []utilitySyntheticCase `json:"syntheticCases"`
	Cases          []utilityCase          `json:"cases"`
}

// utilitySyntheticCase is one small design system built only to reach a branch.
//
// The theme is carried as key/value pairs rather than as a stylesheet path, because these systems
// exist only in the generator: the Go side rebuilds the theme from ThemeEntries and the definitions
// from UtilityBlocks, so the case is self-contained and does not depend on a file that was deleted
// after it was measured.
type utilitySyntheticCase struct {
	Name string `json:"name"`
	// Stylesheet is the source the engine was given, kept so a failure can be read against it.
	Stylesheet string `json:"stylesheet"`
	// RejectionOnly marks a case whose every probe the engine rejects, which for one case is the
	// point rather than a gap: a block using `--modifier(…)` and no `--value(…)` never compiles,
	// because post-condition 1 reads `usedValueFunction`, which only `--value(…)` sets. The flag is
	// on the case rather than inferred from the readings, so a case that stopped compiling anything
	// for some other reason still fails.
	RejectionOnly bool `json:"rejectionOnly"`
	// ThemeEntries is the resolved theme, which for these systems is small and framework-free
	// enough that the entries the blocks consult are all declared in the stylesheet itself.
	ThemeEntries  map[string]string     `json:"themeEntries"`
	UtilityBlocks []utilityBlockFixture `json:"utilityBlocks"`
	Cases         []utilityCase         `json:"cases"`
}

type utilityBlockFixture struct {
	// Name is the `@utility` parameter, including the `-*` suffix for a functional block.
	Name   string `json:"name"`
	Source string `json:"source"`
}

type utilityExceptionFixture struct {
	ClassName string `json:"className"`
	Order     []int  `json:"order"`
	Count     int    `json:"count"`
	// PredictedByDescriptorModel is what the table said, kept so a failure names both answers and
	// the reader can see which side moved.
	PredictedByDescriptorModel string `json:"predictedByDescriptorModel"`
}

type utilityQuirkFixture struct {
	ClassName string `json:"className"`
	Predicted string `json:"predicted"`
	Actual    string `json:"actual"`
}

// utilityCase is one class and the reading the engine gave it.
//
// The parsed candidate's fields are carried alongside the reading so a disagreement can be read as
// "the candidate parsed differently" rather than only as "the reading differed". The evaluator
// consumes a ParsedCandidate, so a candidate-parser regression would otherwise arrive here as an
// unexplained reading change.
type utilityCase struct {
	ClassName     string `json:"className"`
	CandidateKind string `json:"candidateKind"`
	Root          string `json:"root"`
	ValueKind     string `json:"valueKind"`
	Value         string `json:"value"`
	ValueDataType string `json:"valueDataType"`
	Fraction      string `json:"fraction"`
	ModifierKind  string `json:"modifierKind"`
	Modifier      string `json:"modifier"`
	// Reading is null when the engine rejected the class, which is a different answer than `[]#0`
	// and must never be scored as agreement: an empty order sorts last, not first.
	Reading *utilityReadingFixture `json:"reading"`
}

type utilityReadingFixture struct {
	Order []int `json:"order"`
	Count int   `json:"count"`
}

// utilityTailwindPackageRoot is the tailwindcss install the fixture's theme resolves
// `@import "tailwindcss"` against.
//
// The vendored snapshot of 4.3.3, the version the fixture was generated against, so it is present on any
// machine (#sycrdr6).
var utilityTailwindPackageRoot = vendored.TailwindPackageRoot()

func utilityLoadCorpus(t *testing.T) utilityCorpus {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", "utility_fixtures.json"))
	if err != nil {
		t.Fatalf("read utility fixture: %v", err)
	}
	var fixture utilityCorpus
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode utility fixture: %v", err)
	}
	if len(fixture.Cases) == 0 {
		t.Fatal("utility fixture holds no cases")
	}
	return fixture
}

// utilityBuildEvaluator builds the evaluator the way production will: the theme from the repository's
// own stylesheet graph, and the `@utility` definitions parsed out of the block text with the Go CSS
// parser.
//
// The stylesheet is spelled inside its corpus, so this skips naming the variable when the corpus is
// unset, and skips loudly, because a silently absent corpus turns this suite into one that compares
// nothing and still prints green. A corpus that is set but lacks the stylesheet fails.
func utilityBuildEvaluator(t *testing.T, fixture utilityCorpus) (*UtilityEvaluator, bool) {
	t.Helper()

	entryPath := corpus.Resolve(t, fixture.EntryPath)
	theme, _, err := LoadThemeFromFile(entryPath, NodeStylesheetResolver(utilityTailwindPackageRoot))
	if err != nil {
		t.Fatalf("load theme from %s: %v", entryPath, err)
	}

	definitions := utilityParseDefinitions(t, fixture)
	return NewUtilityEvaluator(theme, definitions), true
}

// utilityParseDefinitions turns the fixture's `@utility` block text into definitions.
//
// Only the functional blocks, the ones whose name ends `-*`, are definitions this evaluator answers
// for. A static `@utility` takes no value and has no `--value()` to resolve, so it is a constant
// reading the descriptor table already carries; handling it here would be a second implementation of
// something already measured.
func utilityParseDefinitions(t *testing.T, fixture utilityCorpus) []*UtilityDefinition {
	t.Helper()

	var definitions []*UtilityDefinition
	for _, block := range fixture.UtilityBlocks {
		if !strings.HasSuffix(block.Name, "-*") {
			continue
		}
		nodes, err := ParseCSS(block.Source)
		if err != nil {
			t.Fatalf("parse @utility %s: %v", block.Name, err)
		}
		if len(nodes) != 1 || nodes[0].Kind != KindAtRule || nodes[0].Name != "@utility" {
			t.Fatalf("@utility %s parsed to %d nodes, expected one at-rule", block.Name, len(nodes))
		}
		definitions = append(definitions, &UtilityDefinition{
			Name:  strings.TrimSuffix(block.Name, "-*"),
			Nodes: nodes[0].Nodes,
		})
	}
	if len(definitions) == 0 {
		t.Fatal("fixture holds no functional @utility blocks")
	}
	return definitions
}

// utilityCandidateFromCase rebuilds the ParsedCandidate the engine parsed, from the fixture's record
// of it.
//
// Rebuilt from the fixture rather than re-parsed with ParseCandidate, deliberately. This test is
// about the evaluator, and feeding it a candidate the Go parser produced would make every failure
// ambiguous between the two components. The candidate parser has its own differential
// (candidate_test.go) against the same engine; TestUtilityAgreesThroughTheGoCandidateParser below
// closes the loop by running the whole chain.
func utilityCandidateFromCase(aCase utilityCase) *ParsedCandidate {
	candidate := &ParsedCandidate{
		Kind: ParsedCandidateKind(aCase.CandidateKind),
		Root: aCase.Root,
		Raw:  aCase.ClassName,
	}
	if aCase.ValueKind != "" {
		candidate.Value = &ParsedValue{
			Kind:     ParsedValueKind(aCase.ValueKind),
			Value:    aCase.Value,
			DataType: aCase.ValueDataType,
			Fraction: aCase.Fraction,
		}
	}
	if aCase.ModifierKind != "" {
		candidate.Modifier = &ParsedModifier{
			Kind:  ParsedModifierKind(aCase.ModifierKind),
			Value: aCase.Modifier,
		}
	}
	return candidate
}

// TestUtilityMatchesEngine is the differential: every reading the evaluator gives, against the
// reading the engine gave for the same class.
//
// Only functional candidates on roots this evaluator defines are scored. A static candidate is the
// descriptor table's job and a candidate on an unknown root is nobody's, so counting either here
// would inflate the agreement figure with answers this component did not produce.
//
// A null reading in the fixture is a class the engine rejected, and it is compared as a rejection
// rather than skipped. That direction is where a port fails silently: an evaluator that accepted
// everything would agree on every class that compiles and be wrong on the 6,000-odd that do not,
// while looking identical in a summary that only counted matches.
func TestUtilityMatchesEngine(t *testing.T) {
	t.Parallel()
	fixture := utilityLoadCorpus(t)
	evaluator, ok := utilityBuildEvaluator(t, fixture)
	if !ok {
		return
	}

	var compared, agreedCompiling, agreedRejecting int
	var disagreements []string

	for _, aCase := range fixture.Cases {
		if aCase.CandidateKind != string(ParsedCandidateKindFunctional) || !evaluator.Has(aCase.Root) {
			continue
		}
		compared++

		reading, compiled := evaluator.Reading(utilityCandidateFromCase(aCase))

		if aCase.Reading == nil {
			if compiled {
				if len(disagreements) < 40 {
					disagreements = append(disagreements, fmt.Sprintf(
						"%s: evaluator read %s, engine rejected the class",
						aCase.ClassName, utilityDescribeReading(reading),
					))
				}
				continue
			}
			agreedRejecting++
			continue
		}

		if !compiled {
			if len(disagreements) < 40 {
				disagreements = append(disagreements, fmt.Sprintf(
					"%s: evaluator rejected the class, engine read %s",
					aCase.ClassName, utilityDescribeReading(Reading{Order: aCase.Reading.Order, Count: aCase.Reading.Count}),
				))
			}
			continue
		}

		want := Reading{Order: aCase.Reading.Order, Count: aCase.Reading.Count}
		if !reading.Equal(want) {
			if len(disagreements) < 40 {
				disagreements = append(disagreements, fmt.Sprintf(
					"%s: evaluator read %s, engine read %s",
					aCase.ClassName, utilityDescribeReading(reading), utilityDescribeReading(want),
				))
			}
			continue
		}
		agreedCompiling++
	}

	if len(disagreements) > 0 {
		t.Errorf("%d of %d classes disagreed with the engine:\n  %s", compared-agreedCompiling-agreedRejecting, compared, strings.Join(disagreements, "\n  "))
	}

	// Coverage, printed on success. A suite that compared twelve answers and one that compared
	// thousands are indistinguishable from a green line, and both halves have to be large: the
	// compiling half is what proves arity is counted right, and the rejecting half is what proves
	// the post-conditions reject anything at all.
	if agreedCompiling < 500 {
		t.Errorf("only %d compiling classes were compared; the fixture is too small to have shown anything about arity", agreedCompiling)
	}
	if agreedRejecting < 500 {
		t.Errorf("only %d rejected classes were compared; without them the post-conditions have never been shown to reject", agreedRejecting)
	}

	t.Logf(
		"tailwind %s: %d functional classes on %d @utility roots; %d compiled and agreed, %d rejected and agreed",
		fixture.TailwindVersion, compared, len(evaluator.Definitions), agreedCompiling, agreedRejecting,
	)
}

// TestUtilityReproducesTheKnownExceptions is the acceptance criterion for this component.
//
// These 18 registry classes are what the descriptor table declines by design, and they are enumerated
// in the fixture by the same tool that measured the model. Reproducing them is the whole reason this
// evaluator exists, so they are asserted by name and by reading rather than absorbed into the
// aggregate above: an aggregate that dropped these 18 and gained 18 elsewhere would still be green.
//
// The count is asserted too. If a repository change made the exception set smaller, this component's
// justification changed and that should be read rather than silently passed.
func TestUtilityReproducesTheKnownExceptions(t *testing.T) {
	t.Parallel()
	fixture := utilityLoadCorpus(t)
	evaluator, ok := utilityBuildEvaluator(t, fixture)
	if !ok {
		return
	}

	if len(fixture.KnownExceptions) != 18 {
		t.Errorf(
			"fixture holds %d known exceptions, Phase 0 measured 18; the population this component exists for has changed and the change should be read rather than absorbed",
			len(fixture.KnownExceptions),
		)
	}

	byClassName := make(map[string]utilityCase, len(fixture.Cases))
	for _, aCase := range fixture.Cases {
		byClassName[aCase.ClassName] = aCase
	}

	reproduced := 0
	for _, exception := range fixture.KnownExceptions {
		aCase, present := byClassName[exception.ClassName]
		if !present {
			t.Errorf("%s: the fixture names it an exception and holds no case for it", exception.ClassName)
			continue
		}
		if !evaluator.Has(aCase.Root) {
			t.Errorf("%s: root %q is not defined by any @utility block the evaluator holds", exception.ClassName, aCase.Root)
			continue
		}

		reading, compiled := evaluator.Reading(utilityCandidateFromCase(aCase))
		if !compiled {
			t.Errorf("%s: evaluator rejected the class, engine read %s", exception.ClassName, utilityDescribeReading(Reading{Order: exception.Order, Count: exception.Count}))
			continue
		}
		want := Reading{Order: exception.Order, Count: exception.Count}
		if !reading.Equal(want) {
			t.Errorf(
				"%s: evaluator read %s, engine read %s (the descriptor table predicted %s)",
				exception.ClassName, utilityDescribeReading(reading), utilityDescribeReading(want), exception.PredictedByDescriptorModel,
			)
			continue
		}
		// The exception is only an exception because the table's answer differs. If the two ever
		// agree, this class stopped needing the evaluator and the fixture is stale.
		if utilityDescribeReading(want) == exception.PredictedByDescriptorModel {
			t.Errorf(
				"%s: the descriptor table predicted %s, which is what the engine says; this class is no longer an exception and the fixture is stale",
				exception.ClassName, exception.PredictedByDescriptorModel,
			)
			continue
		}
		reproduced++
	}

	if reproduced != len(fixture.KnownExceptions) {
		t.Errorf("reproduced %d of %d known exceptions", reproduced, len(fixture.KnownExceptions))
	}
	t.Logf("reproduced all %d registry classes the descriptor model declines, each with a reading the table predicts differently", reproduced)
}

// TestUtilityFindsThePerDeclarationRoots asserts the mechanism is detectable from the evaluator's own
// answers, on the roots the engine resolves per declaration.
//
// The property is the one that makes these roots impossible to table: a value satisfying two
// resolution paths produces a higher count than one satisfying a single path, on the same root. It is
// asserted here directly rather than inferred from the aggregate, and it is asserted to be non-empty:
// a repository with no such `@utility` block would make every other test in this file pass over an
// empty population, which is the failure mode this whole slice has been bitten by.
func TestUtilityFindsThePerDeclarationRoots(t *testing.T) {
	t.Parallel()
	fixture := utilityLoadCorpus(t)
	evaluator, ok := utilityBuildEvaluator(t, fixture)
	if !ok {
		return
	}

	if len(fixture.PerDeclarationRoots) == 0 {
		t.Fatal("the fixture reports no per-declaration roots; this component's whole population is empty and every other test here is passing over nothing")
	}

	byClassName := make(map[string]utilityCase, len(fixture.Cases))
	for _, aCase := range fixture.Cases {
		byClassName[aCase.ClassName] = aCase
	}

	checked := 0
	for _, root := range fixture.PerDeclarationRoots {
		if !evaluator.Has(root) {
			t.Errorf("%s: the fixture reports it as per-declaration and no @utility block defines it", root)
			continue
		}

		// `4` is only an integer. `50` is an integer and a `--percentage` key. Same root, same
		// declaration list, and the count has to differ.
		singlePathCase, hasSinglePath := byClassName[root+"-4"]
		multiPathCase, hasMultiPath := byClassName[root+"-50"]
		if !hasSinglePath || !hasMultiPath {
			t.Errorf("%s: the fixture is missing the -4 or -50 probe that separates the two paths", root)
			continue
		}

		singlePath, singleOK := evaluator.Reading(utilityCandidateFromCase(singlePathCase))
		multiPath, multiOK := evaluator.Reading(utilityCandidateFromCase(multiPathCase))
		if !singleOK || !multiOK {
			t.Errorf("%s: one of the two probes did not compile (-4 ok=%v, -50 ok=%v)", root, singleOK, multiOK)
			continue
		}
		if multiPath.Count <= singlePath.Count {
			t.Errorf(
				"%s: -50 read %s and -4 read %s; a value satisfying two resolution paths must survive more declarations than one satisfying a single path, or this root is not per-declaration and the fixture is wrong about it",
				root, utilityDescribeReading(multiPath), utilityDescribeReading(singlePath),
			)
			continue
		}
		checked++
	}

	t.Logf("%d roots resolve per declaration, each proven by a value satisfying two paths outcounting one satisfying a single path", checked)
}

// TestUtilityDropIsNotDefault is the first mutation, and it targets the rule the whole component is.
//
// A declaration whose `--value()` fails to resolve is dropped. The plausible wrong port leaves it in
// place, which agrees with the engine on every class where every declaration resolves, and that is
// most of them. This mutates exactly that rule, by keeping every dropped declaration, and asserts the
// differential catches it.
//
// It is written as a re-evaluation with the drop suppressed rather than as an edit to the evaluator,
// because a mutation the test performs on itself is a mutation that stays true after the next
// refactor. The suppression reaches into the same evaluation state the real path uses, so it cannot
// drift away from what it is mutating.
func TestUtilityDropIsNotDefault(t *testing.T) {
	t.Parallel()
	fixture := utilityLoadCorpus(t)
	evaluator, ok := utilityBuildEvaluator(t, fixture)
	if !ok {
		return
	}

	byClassName := make(map[string]utilityCase, len(fixture.Cases))
	for _, aCase := range fixture.Cases {
		byClassName[aCase.ClassName] = aCase
	}

	caught, checked := 0, 0
	for _, exception := range fixture.KnownExceptions {
		aCase, present := byClassName[exception.ClassName]
		if !present {
			continue
		}
		checked++

		mutated, compiled := evaluator.compileKeepingDroppedDeclarations(utilityCandidateFromCase(aCase))
		if !compiled {
			t.Errorf("%s: the mutant did not compile at all, so this class proves nothing about the drop", exception.ClassName)
			continue
		}
		mutatedSort := PropertySort(mutated)
		want := Reading{Order: exception.Order, Count: exception.Count}
		if !(Reading{Order: mutatedSort.Order, Count: mutatedSort.Count}).Equal(want) {
			caught++
			continue
		}
		t.Errorf(
			"%s: keeping dropped declarations still reads %s, which is what the engine says; the drop rule is not what decides this class's arity and the test is not testing it",
			exception.ClassName, utilityDescribeReading(want),
		)
	}

	if checked == 0 {
		t.Fatal("no exception class was available to mutate")
	}
	if caught != checked {
		t.Errorf("the drop mutation was caught on %d of %d exception classes", caught, checked)
	}
	t.Logf("suppressing the resolve-or-drop rule changes the reading on all %d exception classes, so the differential is testing that rule", caught)
}

// TestUtilityRatioSpliceIsLoadBearing is the second mutation, on the other removal.
//
// When `--value(ratio)` resolves, the declarations that resolved a non-ratio `--value()` are removed.
// `slide-in-from-top-1/2` reads `[]#1` because of it: `1` alone satisfies the `integer` declaration,
// so without the splice the class reads `[]#2`. This suppresses the splice and asserts the reading
// moves on every fraction class the corpus holds.
//
// A fraction class the splice does not move is not a failure of the splice, it is a class where no
// non-ratio declaration resolved, so those are counted separately rather than scored as escapes. The
// assertion is that the population where it does matter is non-empty, since a splice that never fires
// is a branch this suite has not exercised.
func TestUtilityRatioSpliceIsLoadBearing(t *testing.T) {
	t.Parallel()
	fixture := utilityLoadCorpus(t)
	evaluator, ok := utilityBuildEvaluator(t, fixture)
	if !ok {
		return
	}

	moved, unaffected := 0, 0
	for _, aCase := range fixture.Cases {
		if aCase.Reading == nil || aCase.CandidateKind != string(ParsedCandidateKindFunctional) {
			continue
		}
		if aCase.Fraction == "" || !evaluator.Has(aCase.Root) {
			continue
		}
		candidate := utilityCandidateFromCase(aCase)

		real, realOK := evaluator.Reading(candidate)
		if !realOK {
			continue
		}
		mutated, mutatedOK := evaluator.compileWithoutRatioSplice(candidate)
		if !mutatedOK {
			t.Errorf("%s: the mutant did not compile while the real path did", aCase.ClassName)
			continue
		}
		mutatedSort := PropertySort(mutated)
		if (Reading{Order: mutatedSort.Order, Count: mutatedSort.Count}).Equal(real) {
			unaffected++
			continue
		}
		moved++
	}

	if moved == 0 {
		t.Error("suppressing the ratio splice changed no reading; either no fraction class in the corpus resolves a non-ratio declaration, or the splice is dead code")
	}
	t.Logf("suppressing the ratio splice changes %d fraction classes and leaves %d unaffected, which are the ones where no non-ratio declaration resolved", moved, unaffected)
}

// TestUtilityAgreesThroughTheGoCandidateParser closes the loop the other tests deliberately leave
// open.
//
// Every test above feeds the evaluator a candidate rebuilt from the fixture, so a failure is
// unambiguously this component's. That isolation is worth having and it is not the production path:
// in production the class string goes through ParseCandidate and its output goes here. This runs the
// whole chain and asserts the same answers, so a mismatch between what the fixture recorded and what
// the Go parser produces is a failure here rather than a surprise in the rule.
func TestUtilityAgreesThroughTheGoCandidateParser(t *testing.T) {
	t.Parallel()
	fixture := utilityLoadCorpus(t)
	evaluator, ok := utilityBuildEvaluator(t, fixture)
	if !ok {
		return
	}

	designSystem := &utilityTestDesignSystem{evaluator: evaluator}

	var compared, agreed int
	var disagreements []string
	for _, aCase := range fixture.Cases {
		if aCase.CandidateKind != string(ParsedCandidateKindFunctional) || !evaluator.Has(aCase.Root) {
			continue
		}

		parsed := ParseCandidate(aCase.ClassName, designSystem)
		if len(parsed) == 0 {
			// The Go parser produced no candidate. That is the candidate parser's business and it
			// has its own differential; counting it here would score this component on another's
			// answer. It is reported in the log so a large number is visible.
			continue
		}
		// The first reading, matching `getClassOrder`, which takes the position of the first parse.
		candidate := parsed[0]
		if candidate.Kind != ParsedCandidateKindFunctional || candidate.Root != aCase.Root {
			continue
		}
		compared++

		reading, compiled := evaluator.Reading(&candidate)
		if aCase.Reading == nil {
			if compiled && len(disagreements) < 20 {
				disagreements = append(disagreements, fmt.Sprintf("%s: read %s through the Go parser, engine rejected it", aCase.ClassName, utilityDescribeReading(reading)))
			}
			if !compiled {
				agreed++
			}
			continue
		}
		want := Reading{Order: aCase.Reading.Order, Count: aCase.Reading.Count}
		if !compiled || !reading.Equal(want) {
			if len(disagreements) < 20 {
				disagreements = append(disagreements, fmt.Sprintf(
					"%s: read %s through the Go parser (compiled=%v), engine read %s",
					aCase.ClassName, utilityDescribeReading(reading), compiled, utilityDescribeReading(want),
				))
			}
			continue
		}
		agreed++
	}

	if len(disagreements) > 0 {
		t.Errorf("%d of %d classes disagreed end to end:\n  %s", compared-agreed, compared, strings.Join(disagreements, "\n  "))
	}
	if agreed < 500 {
		t.Errorf("only %d classes agreed end to end; the chain has not been shown to work at scale", agreed)
	}
	t.Logf("%d of %d classes agree end to end, from class string through ParseCandidate to a reading", agreed, compared)
}

// TestUtilityDoesNotClaimTheShadowQuirk states the boundary of this component in a test rather than
// only in prose.
//
// Four sweep probes hit an upstream shadow quirk on `[16/9]` values, where the shadow handler splits
// an arbitrary value on `/` looking for a colour slot and emits garbage CSS. No `@utility` block is
// involved, no registry contains them, and the task recorded them as not to be modelled. So the
// assertion is that this evaluator declines them: an evaluator that answered would be claiming a
// mechanism it does not implement, which is a right answer for a wrong reason waiting to become a
// wrong one.
func TestUtilityDoesNotClaimTheShadowQuirk(t *testing.T) {
	t.Parallel()
	fixture := utilityLoadCorpus(t)
	evaluator, ok := utilityBuildEvaluator(t, fixture)
	if !ok {
		return
	}

	if len(fixture.ShadowQuirkProbes) != 4 {
		t.Errorf("fixture holds %d shadow-quirk probes, Phase 0 measured 4", len(fixture.ShadowQuirkProbes))
	}

	for _, probe := range fixture.ShadowQuirkProbes {
		root := probe.ClassName
		if index := strings.Index(root, "-["); index != -1 {
			root = root[:index]
		}
		if evaluator.Has(root) {
			t.Errorf(
				"%s: this evaluator defines root %q, so it is answering a class the shadow quirk explains; the quirk is recorded and not modelled",
				probe.ClassName, root,
			)
		}
	}
	t.Logf("%d shadow-quirk probes are outside this component: none of their roots is an @utility block", len(fixture.ShadowQuirkProbes))
}

// TestUtilitySpacingMultiplierRejects pins the numeric guard that separates a surviving
// `--value(number)` from a dropped one, in isolation from the corpus.
//
// The corpus covers it, but only through classes where several other things are also true. Pinned
// here it names the rule: a non-negative multiple of 0.25 written the way JavaScript prints it. The
// cases are the ones a reasonable guess gets wrong, since `1.50` and `+1` are multiples of 0.25 and
// are not how `String(Number(x))` spells them.
func TestUtilitySpacingMultiplierRejects(t *testing.T) {
	t.Parallel()
	accepted := []string{"0", "1", "0.25", "0.5", "1.25", "2.75", "100"}
	rejected := []string{"1.3", "0.1", "-0.25", "1.50", "+1", "1e2", "", "abc", ".25"}

	for _, value := range accepted {
		if !isValidSpacingMultiplier(value) {
			t.Errorf("isValidSpacingMultiplier(%q) = false, want true", value)
		}
	}
	for _, value := range rejected {
		if isValidSpacingMultiplier(value) {
			t.Errorf("isValidSpacingMultiplier(%q) = true, want false", value)
		}
	}
	t.Logf("%d accepted and %d rejected spacing multipliers", len(accepted), len(rejected))
}

// TestUtilityNormalizesValueArguments pins the registration-time rewrites, which are invisible in
// every reading they are right about.
//
// An argument spelling that resolves nothing produces a dropped declaration and a count one lower,
// so a port that skipped normalization would fail quietly on exactly the classes this component
// exists to get right. The cases are upstream's own documented examples.
func TestUtilityNormalizesValueArguments(t *testing.T) {
	t.Parallel()
	cases := []struct{ argument, want string }{
		{"--spacing", "--spacing-*"},
		{"--spacing- *", "--spacing-*"},
		{"--text- * --line-height", "--text-*--line-height"},
		{"--text --line-height", "--text-*--line-height"},
		{`--text-\* --line-height`, "--text-*--line-height"},
		{"[ *]", "[*]"},
		{"--percentage-*", "--percentage-*"},
		{"--default(4)", "--default(4)"},
		{"integer", "integer"},
		{"[length]", "[length]"},
	}
	for _, aCase := range cases {
		if got := normalizeValueFunctionArgument(aCase.argument); got != aCase.want {
			t.Errorf("normalizeValueFunctionArgument(%q) = %q, want %q", aCase.argument, got, aCase.want)
		}
	}
	t.Logf("%d argument spellings normalize as upstream documents", len(cases))
}

// TestUtilitySyntheticCasesMatchEngine covers the branches the repository's own `@utility` blocks
// cannot reach.
//
// The repository half is the exit criterion and it is not sufficient. Its blocks use `integer`,
// `number`, `ratio`, theme namespaces and bracketed types, and nothing else, so four ported branches
// have no repository class that exercises them: the `percentage` bare type with its integer guard,
// `--modifier(…)` and the post-conditions that gate it, `--default(…)`, and quoted literals.
//
// Each was found by mutation rather than by reading. Deleting the integer-percentage guard changed
// no reading anywhere in the repository corpus and every test still passed, which is what a branch
// with no test looks like from the inside: coverage in the diff, nothing behind it.
//
// The theme is rebuilt from the fixture's entries rather than from a stylesheet, so these cases stay
// self-contained: the generator writes each synthetic stylesheet next to the repository's own entry
// point so `@import "tailwindcss"` resolves, measures it, and deletes it.
func TestUtilitySyntheticCasesMatchEngine(t *testing.T) {
	t.Parallel()
	fixture := utilityLoadCorpus(t)

	if len(fixture.SyntheticCases) == 0 {
		t.Fatal("the fixture holds no synthetic cases; four ported branches then have no test at all")
	}

	var totalCompared, totalCompiling, totalRejecting int
	for _, syntheticCase := range fixture.SyntheticCases {
		// Not parallel: its subtests add to totals the test checks after them, and a parallel subtest would run
		// only after the test had returned, so the check would pass on nothing.
		t.Run(syntheticCase.Name, func(t *testing.T) {
			theme := NewTheme()
			for key, value := range syntheticCase.ThemeEntries {
				if err := theme.Add(key, value, ThemeOptionNone); err != nil {
					t.Fatalf("add theme entry %q: %v", key, err)
				}
			}

			var definitions []*UtilityDefinition
			for _, block := range syntheticCase.UtilityBlocks {
				if !strings.HasSuffix(block.Name, "-*") {
					continue
				}
				nodes, err := ParseCSS(block.Source)
				if err != nil {
					t.Fatalf("parse @utility %s: %v", block.Name, err)
				}
				if len(nodes) != 1 || nodes[0].Kind != KindAtRule {
					t.Fatalf("@utility %s parsed to %d nodes, expected one at-rule", block.Name, len(nodes))
				}
				definitions = append(definitions, &UtilityDefinition{
					Name:  strings.TrimSuffix(block.Name, "-*"),
					Nodes: nodes[0].Nodes,
				})
			}
			if len(definitions) == 0 {
				t.Fatalf("%s declares no functional @utility block", syntheticCase.Name)
			}
			evaluator := NewUtilityEvaluator(theme, definitions)

			compiling, rejecting := 0, 0
			for _, aCase := range syntheticCase.Cases {
				if aCase.CandidateKind != string(ParsedCandidateKindFunctional) || !evaluator.Has(aCase.Root) {
					// A probe the engine did not parse as a functional candidate on one of these
					// roots is not this evaluator's answer to give. `pct` bare parses as a static
					// candidate, and counting it here would score this component on the candidate
					// parser's behaviour.
					continue
				}
				totalCompared++

				reading, compiled := evaluator.Reading(utilityCandidateFromCase(aCase))
				if aCase.Reading == nil {
					if compiled {
						t.Errorf("%s: evaluator read %s, engine rejected the class", aCase.ClassName, utilityDescribeReading(reading))
						continue
					}
					rejecting++
					continue
				}
				want := Reading{Order: aCase.Reading.Order, Count: aCase.Reading.Count}
				if !compiled {
					t.Errorf("%s: evaluator rejected the class, engine read %s", aCase.ClassName, utilityDescribeReading(want))
					continue
				}
				if !reading.Equal(want) {
					t.Errorf("%s: evaluator read %s, engine read %s", aCase.ClassName, utilityDescribeReading(reading), utilityDescribeReading(want))
					continue
				}
				compiling++
			}

			// Each synthetic case has to exercise both directions, or it is not testing the branch it
			// was written for: a case where everything compiles has not shown its guard rejects, and
			// one where nothing does has not shown its resolver resolves.
			if compiling == 0 && !syntheticCase.RejectionOnly {
				t.Errorf("%s: no probe compiled, so this case has not shown its resolution branch works", syntheticCase.Name)
			}
			if compiling > 0 && syntheticCase.RejectionOnly {
				t.Errorf("%s: the fixture marks this case rejection-only and %d probes compiled", syntheticCase.Name, compiling)
			}
			if rejecting == 0 {
				t.Errorf("%s: no probe was rejected, so this case has not shown its guard rejects anything", syntheticCase.Name)
			}
			totalCompiling += compiling
			totalRejecting += rejecting
		})
	}

	t.Logf(
		"%d synthetic design systems; %d probes compared, %d compiled and agreed, %d rejected and agreed",
		len(fixture.SyntheticCases), totalCompared, totalCompiling, totalRejecting,
	)
}

// utilityDescribeReading renders a reading in the `[order]#count` form the extractor and this
// package's other tests print, so a failure here is greppable against a Phase 0 report.
func utilityDescribeReading(reading Reading) string {
	parts := make([]string, len(reading.Order))
	for index, value := range reading.Order {
		parts[index] = fmt.Sprint(value)
	}
	return "[" + strings.Join(parts, ",") + "]#" + fmt.Sprint(reading.Count)
}

// utilityTestDesignSystem is the DesignSystem the candidate parser needs, answering only about the
// roots this evaluator defines.
//
// Minimal on purpose. The parser asks which roots exist and of what kind; every other question it
// could ask concerns variants, which this component does not reach.
type utilityTestDesignSystem struct {
	evaluator *UtilityEvaluator
}

func (system *utilityTestDesignSystem) Prefix() string { return "" }

func (system *utilityTestDesignSystem) HasUtility(root string, kind UtilityKind) bool {
	return kind == UtilityKindFunctional && system.evaluator.Has(root)
}

func (system *utilityTestDesignSystem) HasVariant(string) bool { return false }

func (system *utilityTestDesignSystem) VariantKind(string) ParsedVariantKind {
	return ParsedVariantKindStatic
}

func (system *utilityTestDesignSystem) VariantCompoundsWith(string, ParsedVariant) bool {
	return false
}
