package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixture is what the shipped Tailwind 4.3.3 candidate parser read for every class in the
// corpus, captured by internal/lint/rules/tailwind/tools/generate_candidate and checked in next to this test.
//
// Measured rather than transcribed, and the thing measurement buys here is order. `parseCandidate`
// is a generator; 2,760 of the 5,056 classes in this fixture yield more than one reading, and the
// caller keeps the first that compiles. A hand-written expectation set would almost certainly have
// recorded the readings as a set, agreed with a port that had `border-b` and `border`+`b` the wrong
// way round, and let every `border-b` in the repository resolve to the wrong utility silently.
//
// The fixture also carries the design system's own tables, so this test answers the parser's four
// questions with the same 1,236 utility roots and 88 variant roots the engine had. Without that
// this would be a test of a stub.

type candidateCorpus struct {
	TailwindVersion string `json:"tailwindVersion"`
	// CandidateResolvedByIdentity records that the generator confirmed the export it captured is
	// the ambiguity-producing parser, by requiring the known two-candidate reading of `border-b`.
	// A false here would mean the corpus was captured from something unverified.
	CandidateResolvedByIdentity bool `json:"parseCandidateResolvedByIdentity"`
	// ThemeEntry names the repository stylesheet whose @utility and @theme blocks are in the
	// tables, so a fixture captured against framework defaults alone is visible in the diff.
	ThemeEntry *string `json:"themeEntry"`

	UtilityRootCount         int `json:"utilityRootCount"`
	VariantRootCount         int `json:"variantRootCount"`
	ClassNameOccurrences     int `json:"classNameOccurrences"`
	RepositoryClassCount     int `json:"repositoryClassCount"`
	AmbiguousCount           int `json:"ambiguousCount"`
	AmbiguousRepositoryCount int `json:"ambiguousRepositoryCount"`

	UtilityRoots map[string][]string       `json:"utilityRoots"`
	VariantRoots map[string]fixtureVariant `json:"variantRoots"`

	// RepositoryUtilityMarkers and RepositoryVariantMarkers are roots only this repository's own
	// stylesheets register. They exist because the root counts cannot tell a loaded theme from an
	// unloaded one: Tailwind registers its 1,201 built-in utility roots from JavaScript, so a
	// capture against a stylesheet importing nothing still reports 1,201 roots and 83 variants.
	// That was measured against a control, not assumed. These markers are the part that only
	// appears when the repository's @import graph actually loaded.
	RepositoryUtilityMarkers []string `json:"repositoryUtilityMarkers"`
	RepositoryVariantMarkers []string `json:"repositoryVariantMarkers"`

	CompoundsForSelectorsProbes []fixtureCompoundsProbe `json:"compoundsForSelectorsProbes"`

	Cases []candidateCase `json:"cases"`
}

// fixtureVariant is one variant root's registration, as the engine reported it.
type fixtureVariant struct {
	Kind string `json:"kind"`
	// Compounds and CompoundsWith are the bitmask upstream calls `Compounds`. They are carried as
	// the raw integers the engine held rather than decoded here, so the compatibility test in this
	// package is the same bitwise-and the engine performs.
	Compounds     int `json:"compounds"`
	CompoundsWith int `json:"compoundsWith"`
}

// fixtureCompoundsProbe pins compoundsForSelectors, the one design-system answer the port has to
// compute rather than look up: an arbitrary child variant's compounding is derived from its
// selector.
type fixtureCompoundsProbe struct {
	Parent        string `json:"parent"`
	Selector      string `json:"selector"`
	CompoundsWith bool   `json:"compoundsWith"`
}

// candidateCase is one class and the ordered readings the engine produced for it.
type candidateCase struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Input  string `json:"input"`
	// Candidates is null only when the engine threw, which no case in the corpus does. An empty
	// array is a real reading: the class is not a utility.
	Candidates []fixtureCandidate `json:"candidates"`
	Threw      *string            `json:"threw"`
	Ambiguous  bool               `json:"ambiguous"`
}

type fixtureCandidate struct {
	Kind      string               `json:"kind"`
	Root      string               `json:"root"`
	Property  string               `json:"property"`
	Value     json.RawMessage      `json:"value"`
	Modifier  *fixtureModifier     `json:"modifier"`
	Variants  []fixtureVariantNode `json:"variants"`
	Important bool                 `json:"important"`
	Raw       string               `json:"raw"`
}

type fixtureModifier struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type fixtureUtilityValue struct {
	Kind     string  `json:"kind"`
	DataType *string `json:"dataType"`
	Value    string  `json:"value"`
	Fraction *string `json:"fraction"`
}

type fixtureVariantNode struct {
	Kind     string               `json:"kind"`
	Root     string               `json:"root"`
	Selector string               `json:"selector"`
	Relative bool                 `json:"relative"`
	Value    *fixtureVariantValue `json:"value"`
	Modifier *fixtureModifier     `json:"modifier"`
	Variant  *fixtureVariantNode  `json:"variant"`
}

type fixtureVariantValue struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

func loadCandidateCorpus(t *testing.T) candidateCorpus {
	t.Helper()

	path := filepath.Join("testdata", "candidate_fixtures.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var corpus candidateCorpus
	if err := json.Unmarshal(contents, &corpus); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if len(corpus.Cases) == 0 {
		t.Fatalf("%s holds no cases", path)
	}
	if !corpus.CandidateResolvedByIdentity {
		t.Fatalf("%s was captured without confirming parseCandidate by identity", path)
	}
	return corpus
}

// fixtureDesignSystem answers the parser's four questions from the tables the engine reported, so a
// disagreement in this suite is a parser defect and can be nothing else.
type fixtureDesignSystem struct {
	prefix       string
	utilityRoots map[string][]string
	variantRoots map[string]fixtureVariant
}

func newFixtureDesignSystem(corpus candidateCorpus) *fixtureDesignSystem {
	return &fixtureDesignSystem{
		utilityRoots: corpus.UtilityRoots,
		variantRoots: corpus.VariantRoots,
	}
}

func (system *fixtureDesignSystem) Prefix() string { return system.prefix }

func (system *fixtureDesignSystem) HasUtility(root string, kind UtilityKind) bool {
	for _, declared := range system.utilityRoots[root] {
		if declared == string(kind) {
			return true
		}
	}
	return false
}

func (system *fixtureDesignSystem) HasVariant(root string) bool {
	_, exists := system.variantRoots[root]
	return exists
}

func (system *fixtureDesignSystem) VariantKind(root string) ParsedVariantKind {
	return ParsedVariantKind(system.variantRoots[root].Kind)
}

// VariantCompoundsWith reproduces upstream's `Variants.compoundsWith`, including the branch where
// the child is an arbitrary variant and its compounding is computed from its selector rather than
// looked up. That computation is compoundsForSelectors, ported below and pinned by its own test
// against the probes in the fixture.
func (system *fixtureDesignSystem) VariantCompoundsWith(parent string, child ParsedVariant) bool {
	parentInfo, parentExists := system.variantRoots[parent]
	if !parentExists {
		return false
	}

	var childCompounds int
	if child.Kind == ParsedVariantKindArbitrary {
		childCompounds = int(compoundsForSelectors([]string{child.Selector}))
	} else {
		childInfo, childExists := system.variantRoots[child.Root]
		if !childExists {
			return false
		}
		childCompounds = childInfo.Compounds
	}

	if parentInfo.Kind != string(ParsedVariantKindCompound) {
		return false
	}
	if childCompounds == int(VariantCompoundsNever) {
		return false
	}
	if parentInfo.CompoundsWith == int(VariantCompoundsNever) {
		return false
	}
	return parentInfo.CompoundsWith&childCompounds != 0
}

// describeParsedCandidates renders readings as one line each, in order, so a mismatch reports which
// reading differed and at what position rather than dumping two structs.
//
// Order is rendered explicitly with an index because it is the property under test: two lists
// holding the same readings in a different order must read as obviously different here.
func describeParsedCandidates(candidates []ParsedCandidate) string {
	if len(candidates) == 0 {
		return "  <no readings>\n"
	}
	var builder strings.Builder
	for index, candidate := range candidates {
		fmt.Fprintf(&builder, "  [%d] %s\n", index, describeParsedCandidate(candidate))
	}
	return builder.String()
}

func describeParsedCandidate(candidate ParsedCandidate) string {
	var parts []string
	parts = append(parts, "kind="+string(candidate.Kind))

	switch candidate.Kind {
	case ParsedCandidateKindArbitrary:
		parts = append(parts, "property="+quoteForDiff(candidate.Property))
		parts = append(parts, "value="+quoteForDiff(candidate.PropertyValue))
	default:
		parts = append(parts, "root="+quoteForDiff(candidate.Root))
	}

	if candidate.Value == nil {
		parts = append(parts, "value=<none>")
	} else if candidate.Kind != ParsedCandidateKindArbitrary {
		value := "value={" + string(candidate.Value.Kind) + " " + quoteForDiff(candidate.Value.Value)
		if candidate.Value.DataType != "" {
			value += " dataType=" + quoteForDiff(candidate.Value.DataType)
		}
		if candidate.Value.Fraction != "" {
			value += " fraction=" + quoteForDiff(candidate.Value.Fraction)
		}
		parts = append(parts, value+"}")
	}

	parts = append(parts, "modifier="+describeParsedModifier(candidate.Modifier))
	parts = append(parts, fmt.Sprintf("important=%t", candidate.Important))
	parts = append(parts, "raw="+quoteForDiff(candidate.Raw))

	variants := make([]string, 0, len(candidate.Variants))
	for _, variant := range candidate.Variants {
		variants = append(variants, describeParsedVariant(variant))
	}
	parts = append(parts, "variants=["+strings.Join(variants, " ")+"]")

	return strings.Join(parts, " ")
}

func describeParsedModifier(modifier *ParsedModifier) string {
	if modifier == nil {
		return "<none>"
	}
	return "{" + string(modifier.Kind) + " " + quoteForDiff(modifier.Value) + "}"
}

func describeParsedVariant(variant ParsedVariant) string {
	switch variant.Kind {
	case ParsedVariantKindArbitrary:
		return fmt.Sprintf("{arbitrary selector=%s relative=%t}", quoteForDiff(variant.Selector), variant.Relative)
	case ParsedVariantKindStatic:
		return "{static root=" + quoteForDiff(variant.Root) + "}"
	case ParsedVariantKindFunctional:
		value := "<none>"
		if variant.Value != nil {
			value = "{" + string(variant.Value.Kind) + " " + quoteForDiff(variant.Value.Value) + "}"
		}
		return "{functional root=" + quoteForDiff(variant.Root) + " value=" + value +
			" modifier=" + describeParsedModifier(variant.Modifier) + "}"
	case ParsedVariantKindCompound:
		nested := "<none>"
		if variant.Variant != nil {
			nested = describeParsedVariant(*variant.Variant)
		}
		return "{compound root=" + quoteForDiff(variant.Root) + " modifier=" + describeParsedModifier(variant.Modifier) +
			" variant=" + nested + "}"
	default:
		return "{" + string(variant.Kind) + "}"
	}
}

// candidatesFromFixture converts the engine's serialized readings into the port's type, so the
// comparison is between two values of the same Go type and the conversion is an explicit step this
// test controls rather than something struct tags do invisibly.
func candidatesFromFixture(t *testing.T, fixtures []fixtureCandidate) []ParsedCandidate {
	t.Helper()

	candidates := make([]ParsedCandidate, 0, len(fixtures))
	for _, fixture := range fixtures {
		candidate := ParsedCandidate{
			Kind:      ParsedCandidateKind(fixture.Kind),
			Root:      fixture.Root,
			Important: fixture.Important,
			Raw:       fixture.Raw,
			Variants:  variantsFromFixture(t, fixture.Variants),
		}

		if fixture.Modifier != nil {
			candidate.Modifier = &ParsedModifier{
				Kind:  ParsedModifierKind(fixture.Modifier.Kind),
				Value: fixture.Modifier.Value,
			}
		}

		switch candidate.Kind {
		case ParsedCandidateKindArbitrary:
			candidate.Property = fixture.Property
			// An arbitrary candidate's `value` is a bare JSON string rather than an object, which
			// is why Value is RawMessage on the fixture struct: the field is genuinely two
			// different shapes depending on the kind.
			var propertyValue string
			if err := json.Unmarshal(fixture.Value, &propertyValue); err != nil {
				t.Fatalf("arbitrary candidate value was not a string: %v", err)
			}
			candidate.PropertyValue = propertyValue

		case ParsedCandidateKindFunctional:
			if len(fixture.Value) > 0 && string(fixture.Value) != "null" {
				var utilityValue fixtureUtilityValue
				if err := json.Unmarshal(fixture.Value, &utilityValue); err != nil {
					t.Fatalf("functional candidate value was not an object: %v", err)
				}
				value := &ParsedValue{
					Kind:  ParsedValueKind(utilityValue.Kind),
					Value: utilityValue.Value,
				}
				if utilityValue.DataType != nil {
					value.DataType = *utilityValue.DataType
				}
				if utilityValue.Fraction != nil {
					value.Fraction = *utilityValue.Fraction
				}
				candidate.Value = value
			}
		}

		candidates = append(candidates, candidate)
	}
	return candidates
}

func variantsFromFixture(t *testing.T, fixtures []fixtureVariantNode) []ParsedVariant {
	t.Helper()

	variants := make([]ParsedVariant, 0, len(fixtures))
	for _, fixture := range fixtures {
		variants = append(variants, variantFromFixture(t, fixture))
	}
	return variants
}

func variantFromFixture(t *testing.T, fixture fixtureVariantNode) ParsedVariant {
	t.Helper()

	variant := ParsedVariant{
		Kind:     ParsedVariantKind(fixture.Kind),
		Root:     fixture.Root,
		Selector: fixture.Selector,
		Relative: fixture.Relative,
	}
	if fixture.Value != nil {
		variant.Value = &ParsedVariantValue{
			Kind:  ParsedValueKind(fixture.Value.Kind),
			Value: fixture.Value.Value,
		}
	}
	if fixture.Modifier != nil {
		variant.Modifier = &ParsedModifier{
			Kind:  ParsedModifierKind(fixture.Modifier.Kind),
			Value: fixture.Modifier.Value,
		}
	}
	if fixture.Variant != nil {
		nested := variantFromFixture(t, *fixture.Variant)
		variant.Variant = &nested
	}
	return variant
}

// TestParseCandidateMatchesEngine is the whole contract: for every class in the corpus, the readings
// this port produces are the readings the engine produced, in the same order.
//
// The counts are printed on success rather than only on failure, because a suite that ran six cases
// and a suite that ran five thousand are indistinguishable from a green line.
func TestParseCandidateMatchesEngine(t *testing.T) {
	t.Parallel()
	corpus := loadCandidateCorpus(t)
	designSystem := newFixtureDesignSystem(corpus)

	comparedClasses := 0
	comparedCandidates := 0
	comparedAmbiguous := 0

	for _, testCase := range corpus.Cases {
		if testCase.Threw != nil {
			// No case in the corpus does this, and if a future capture produces one it is a finding
			// about the corpus rather than about the port.
			t.Errorf("%s: the engine threw %q; this port has no error path to compare against", testCase.Name, *testCase.Threw)
			continue
		}

		expected := candidatesFromFixture(t, testCase.Candidates)
		actual := ParseCandidate(testCase.Input, designSystem)

		comparedClasses++
		comparedCandidates += len(expected)
		if testCase.Ambiguous {
			comparedAmbiguous++
		}

		expectedText := describeParsedCandidates(expected)
		actualText := describeParsedCandidates(actual)
		if expectedText != actualText {
			t.Errorf(
				"%s: reading %s (%s)\nengine:\n%sport:\n%s",
				testCase.Name, quoteForDiff(testCase.Input), testCase.Source, expectedText, actualText,
			)
		}
	}

	// The volume assertions. A suite that compared almost nothing looks exactly like a suite that
	// compared everything and agreed, so what makes the green line mean something is asserted here
	// rather than assumed.
	if comparedClasses < 1000 {
		t.Fatalf("only %d classes compared; the corpus is too small to mean anything", comparedClasses)
	}
	if comparedAmbiguous < 500 {
		t.Fatalf("only %d ambiguous classes compared; the ordering claim rests on too few", comparedAmbiguous)
	}
	if corpus.UtilityRootCount < 500 || corpus.VariantRootCount < 50 {
		t.Fatalf(
			"the captured design system holds only %d utility roots and %d variant roots; even the framework defaults are missing",
			corpus.UtilityRootCount, corpus.VariantRootCount,
		)
	}
	// The counts above pass against framework defaults alone. These do not: they are roots only
	// this repository's own `@utility` blocks and `@theme` breakpoints register, so they are what
	// distinguishes a fixture that read this repository from one that read Tailwind's own defaults
	// while claiming otherwise.
	if len(corpus.RepositoryUtilityMarkers) == 0 || len(corpus.RepositoryVariantMarkers) == 0 {
		t.Fatalf("the fixture names no repository markers, so it cannot show that the theme graph loaded")
	}
	for _, marker := range corpus.RepositoryUtilityMarkers {
		if _, exists := corpus.UtilityRoots[marker]; !exists {
			t.Errorf("the repository utility root %q is absent; these tables are framework defaults", marker)
		}
	}
	for _, marker := range corpus.RepositoryVariantMarkers {
		if _, exists := corpus.VariantRoots[marker]; !exists {
			t.Errorf("the repository variant root %q is absent; the @theme block did not load", marker)
		}
	}

	t.Logf(
		"compared %d classes against tailwind %s, %d readings in total; "+
			"%d classes read more than one way and their order was compared positionally; "+
			"the design system held %d utility roots and %d variant roots; "+
			"%d of the classes are ones this repository actually writes, %d of those ambiguous",
		comparedClasses, corpus.TailwindVersion, comparedCandidates,
		comparedAmbiguous, corpus.UtilityRootCount, corpus.VariantRootCount,
		corpus.RepositoryClassCount, corpus.AmbiguousRepositoryCount,
	)
	t.Logf(
		"the design system carried this repository's own roots, confirmed by %d utility and %d variant markers",
		len(corpus.RepositoryUtilityMarkers), len(corpus.RepositoryVariantMarkers),
	)
}

// TestParseCandidateOrderIsCompared is the guard on the guard.
//
// The suite above compares rendered text, and rendered text would compare equal for two orderings if
// the renderer ever stopped showing position. This asserts that a reversed reading list really does
// produce a different rendering, so the ordering assertion cannot quietly become a set comparison.
//
// A check that has never returned a positive has not been shown to be able to.
func TestParseCandidateOrderIsCompared(t *testing.T) {
	t.Parallel()
	corpus := loadCandidateCorpus(t)
	designSystem := newFixtureDesignSystem(corpus)

	checked := 0
	for _, testCase := range corpus.Cases {
		if !testCase.Ambiguous {
			continue
		}

		actual := ParseCandidate(testCase.Input, designSystem)
		if len(actual) < 2 {
			t.Errorf("%s: the fixture calls this ambiguous but the port produced %d readings", testCase.Name, len(actual))
			continue
		}

		reversed := make([]ParsedCandidate, len(actual))
		for index, candidate := range actual {
			reversed[len(actual)-1-index] = candidate
		}

		if describeParsedCandidates(actual) == describeParsedCandidates(reversed) {
			t.Errorf("%s: reversing the readings did not change the rendering, so order is not being compared", testCase.Name)
		}

		checked++
		if checked >= 200 {
			break
		}
	}

	if checked < 100 {
		t.Fatalf("only %d ambiguous classes were available to check; the fixture is not exercising order", checked)
	}
	t.Logf("confirmed on %d ambiguous classes that a reversed reading list renders differently", checked)
}

// TestBorderBReadsBothWaysInOrder pins the single case the whole task is built around, by name.
//
// The suite above would catch this too, as one failure among five thousand. This one names it, so a
// regression reports `border-b` rather than reporting a diff someone has to read to understand.
func TestBorderBReadsBothWaysInOrder(t *testing.T) {
	t.Parallel()
	corpus := loadCandidateCorpus(t)
	designSystem := newFixtureDesignSystem(corpus)

	readings := ParseCandidate("border-b", designSystem)
	if len(readings) != 2 {
		t.Fatalf("border-b produced %d readings, want 2:\n%s", len(readings), describeParsedCandidates(readings))
	}

	if readings[0].Root != "border-b" || readings[0].Value != nil {
		t.Errorf("the first reading of border-b must be the root `border-b` with no value, got %s", describeParsedCandidate(readings[0]))
	}
	if readings[1].Root != "border" || readings[1].Value == nil || readings[1].Value.Value != "b" {
		t.Errorf("the second reading of border-b must be the root `border` with the named value `b`, got %s", describeParsedCandidate(readings[1]))
	}
}

// TestCompoundsForSelectorsMatchesEngine pins the one design-system answer the port computes rather
// than looks up.
//
// Every other question the parser asks is a table lookup the fixture can answer directly. An
// arbitrary child variant's compounding is not: it is derived from the selector text, so the
// derivation is ported and has to be measured like everything else.
func TestCompoundsForSelectorsMatchesEngine(t *testing.T) {
	t.Parallel()
	corpus := loadCandidateCorpus(t)
	designSystem := newFixtureDesignSystem(corpus)

	if len(corpus.CompoundsForSelectorsProbes) == 0 {
		t.Fatalf("the fixture holds no compoundsForSelectors probes")
	}

	for _, probe := range corpus.CompoundsForSelectorsProbes {
		child := ParsedVariant{
			Kind:     ParsedVariantKindArbitrary,
			Selector: probe.Selector,
			Relative: len(probe.Selector) > 0 && (probe.Selector[0] == '>' || probe.Selector[0] == '+' || probe.Selector[0] == '~'),
		}
		actual := designSystem.VariantCompoundsWith(probe.Parent, child)
		if actual != probe.CompoundsWith {
			t.Errorf(
				"compoundsWith(%s, %s) = %t, engine says %t",
				probe.Parent, quoteForDiff(probe.Selector), actual, probe.CompoundsWith,
			)
		}
	}

	t.Logf("compared %d compoundsForSelectors probes", len(corpus.CompoundsForSelectorsProbes))
}

// TestGateCandidatesPreservesOrder pins the property gate.go depends on across the conversion.
//
// gate.go's computeGroupKey takes candidates[0]. If GateCandidates ever reordered, sorted, or
// deduplicated, that choice would silently start reading a different candidate, which is the same
// class of silent wrong answer the whole component is built to prevent.
func TestGateCandidatesPreservesOrder(t *testing.T) {
	t.Parallel()
	corpus := loadCandidateCorpus(t)
	designSystem := newFixtureDesignSystem(corpus)

	checked := 0
	for _, testCase := range corpus.Cases {
		if !testCase.Ambiguous {
			continue
		}

		parsed := ParseCandidate(testCase.Input, designSystem)
		gated := GateCandidates(parsed)

		if len(gated) != len(parsed) {
			t.Fatalf("%s: %d readings became %d gate candidates", testCase.Name, len(parsed), len(gated))
		}
		for index := range parsed {
			if gated[index].Root != parsed[index].Root {
				t.Fatalf(
					"%s: gate candidate %d has root %q but reading %d has root %q; the conversion reordered",
					testCase.Name, index, gated[index].Root, index, parsed[index].Root,
				)
			}
		}

		checked++
		if checked >= 500 {
			break
		}
	}

	if checked < 100 {
		t.Fatalf("only %d ambiguous classes were checked; the order claim is not being exercised", checked)
	}
	t.Logf("confirmed order is preserved through the gate conversion on %d ambiguous classes", checked)
}

// TestGateCandidatesSeparatesWhatMustNotMerge asserts the two facts the conversion could plausibly
// lose: a modifier and a compound variant's inner variant.
//
// Both are dropped by a naive flattening, and both are the difference between two classes that
// produce different CSS. A bucketing view is allowed to lose information; it is not allowed to lose
// this.
func TestGateCandidatesSeparatesWhatMustNotMerge(t *testing.T) {
	t.Parallel()
	corpus := loadCandidateCorpus(t)
	designSystem := newFixtureDesignSystem(corpus)

	keyOf := func(className string) string {
		gated := GateCandidates(ParseCandidate(className, designSystem))
		if len(gated) == 0 {
			t.Fatalf("%s produced no readings, so it cannot be compared", className)
		}
		return variantsKey(gated[0].Variants) + "|" + importanceKey(gated[0].Important) + "|" + valueKey(gated[0].Value)
	}

	mustDiffer := [][2]string{
		// A modifier changes the colour, so these are not interchangeable.
		{"bg-red-500/50", "bg-red-500"},
		{"bg-red-500/50", "bg-red-500/75"},
		// The arbitrary-modifier pairs are the ones that actually exercise the fold. A named
		// modifier is already carried by Fraction, so dropping the fold still separates those; an
		// arbitrary modifier suppresses Fraction by design, and then the fold is the only thing
		// keeping these apart. Found by mutation: removing the fold left the named pairs passing.
		{"bg-red-500/[0.5]", "bg-red-500"},
		{"bg-red-500/[0.5]", "bg-red-500/[0.75]"},
		{"bg-red-500/[var(--a)]", "bg-red-500/[var(--b)]"},
		// A compound variant's inner variant is a different selector.
		{"group-hover:flex", "group-focus:flex"},
		// A compound's own modifier names which group is watched.
		{"group-hover/one:flex", "group-hover/two:flex"},
		// Variant order changes the nesting and therefore the CSS.
		{"sm:hover:flex", "hover:sm:flex"},
		// Importance changes specificity.
		{"bg-red-500!", "bg-red-500"},
	}

	for _, pair := range mustDiffer {
		if keyOf(pair[0]) == keyOf(pair[1]) {
			t.Errorf("%s and %s produced the same gate key; the conversion lost what separates them", pair[0], pair[1])
		}
	}

	t.Logf("confirmed %d pairs that must not share a bucket do not", len(mustDiffer))
}

// TestArbitraryTypehintIsSplitFromTheValue pins the coupling the descriptor table depends on.
//
// `bg-[color:var(--x)]` carries an author-written type annotation. The engine trusts that
// annotation instead of inferring, which matters because `InferDataType` declines `var(...)`
// outright: without the annotation reaching the table, every annotated class falls to the axis
// fallback and reads as something else. The descriptor lookup reads `Value.DataType` and expects
// `Value.Value` to hold the value alone.
//
// So the split is a contract between two components rather than an internal detail, and it has the
// failure shape this slice keeps meeting: a parse that looks correct and a reading that is wrong.
// A port that left `color:` on the front of the value, or dropped the annotation instead of
// carrying it, would still produce a plausible candidate and would silently change the reading.
//
// The corpus test covers these cases too, as two rows among five thousand. This one names the
// contract, so a regression reports what broke rather than a diff someone has to interpret.
func TestArbitraryTypehintIsSplitFromTheValue(t *testing.T) {
	t.Parallel()
	corpus := loadCandidateCorpus(t)
	designSystem := newFixtureDesignSystem(corpus)

	cases := []struct {
		className string
		dataType  string
		value     string
	}{
		// Both spellings of the same annotation must land identically, because the `(…)` shorthand
		// is rewritten into the bracketed form during parsing.
		{"bg-[color:var(--my-color)]", "color", "var(--my-color)"},
		{"bg-(color:--my-var)", "color", "var(--my-var)"},
		// An unannotated arbitrary value must carry no type, so the table infers rather than
		// trusting an annotation that was never written.
		{"bg-[#0088cc]", "", "#0088cc"},
	}

	for _, testCase := range cases {
		readings := ParseCandidate(testCase.className, designSystem)
		if len(readings) == 0 {
			t.Errorf("%s produced no readings", testCase.className)
			continue
		}

		value := readings[0].Value
		if value == nil {
			t.Errorf("%s: the first reading carries no value", testCase.className)
			continue
		}
		if value.Kind != ParsedValueKindArbitrary {
			t.Errorf("%s: the first reading's value is %s, want arbitrary", testCase.className, value.Kind)
			continue
		}
		if value.DataType != testCase.dataType {
			t.Errorf("%s: DataType is %q, want %q", testCase.className, value.DataType, testCase.dataType)
		}
		// The annotation must be gone from the value, not merely also present in DataType.
		if value.Value != testCase.value {
			t.Errorf("%s: Value is %q, want %q", testCase.className, value.Value, testCase.value)
		}
	}
}
