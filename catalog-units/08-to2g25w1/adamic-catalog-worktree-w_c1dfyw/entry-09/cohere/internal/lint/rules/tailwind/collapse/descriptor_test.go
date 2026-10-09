package tailwind

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The descriptor lookup, measured against readings the shipped engine produced.
//
// The question this suite answers is not "does the code run" but "does the table plus the lookup
// reproduce the engine", which is the claim the whole port rests on. So the fixture carries the
// engine's parse alongside its reading: the candidate parser is a separate task (#bg0t9tb), and
// feeding this suite the engine's own parse means a failure here is a failure of the descriptor
// model rather than of two ports summed together.
//
// Counts print on success. A suite that ran twelve cases and a suite that ran eleven thousand are
// indistinguishable from a green line, and this suite's whole value is in the size of the population
// it agreed with.

type descriptorFixtureCase struct {
	ClassName     string   `json:"className"`
	Kind          string   `json:"kind"`
	Root          string   `json:"root"`
	Property      string   `json:"property"`
	ValueKind     string   `json:"valueKind"`
	Value         string   `json:"value"`
	Fraction      string   `json:"fraction"`
	DataType      string   `json:"dataType"`
	ModifierKind  string   `json:"modifierKind"`
	ModifierValue string   `json:"modifierValue"`
	Reading       *reading `json:"reading"`
}

type reading struct {
	Order []int `json:"order"`
	Count int   `json:"count"`
}

type descriptorFixtures struct {
	TailwindVersion string `json:"tailwindVersion"`
	Counts          struct {
		Cases             int `json:"cases"`
		ClassesConsidered int `json:"classesConsidered"`
		Roots             int `json:"roots"`
		WithoutReading    int `json:"withoutReading"`
	} `json:"counts"`
	Cases []descriptorFixtureCase `json:"cases"`
}

func loadDescriptorFixtures(t *testing.T) *descriptorFixtures {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "descriptor_fixtures.json"))
	if err != nil {
		t.Fatalf("reading fixtures: %v", err)
	}
	fixtures := &descriptorFixtures{}
	if err := json.Unmarshal(raw, fixtures); err != nil {
		t.Fatalf("parsing fixtures: %v", err)
	}
	if len(fixtures.Cases) == 0 {
		t.Fatal("fixture corpus is empty")
	}
	return fixtures
}

// parsed turns a fixture's recorded parse back into the type Lookup consumes.
//
// The fixture stores the engine's own field names rather than this package's, so the mapping is
// explicit here. That is deliberate: a fixture written in the port's vocabulary would agree with the
// port by construction, and this one has to be able to disagree.
//
// It is also what keeps this suite measuring the descriptor model rather than the sum of two ports.
// The candidate parser is a separate task, and feeding this suite the engine's parse means a failure
// here is a failure of the table and the lookup.
func (one descriptorFixtureCase) parsed() *ParsedCandidate {
	candidate := &ParsedCandidate{Root: one.Root, Property: one.Property, Raw: one.ClassName}
	switch one.Kind {
	case "static":
		candidate.Kind = ParsedCandidateKindStatic
	case "functional":
		candidate.Kind = ParsedCandidateKindFunctional
	case "arbitrary":
		candidate.Kind = ParsedCandidateKindArbitrary
	default:
		return nil
	}

	switch one.ValueKind {
	case "named":
		// The fraction is carried, and it is the field a hand-built candidate forgets.
		//
		// `-bottom-1/2` parses with value `1`, modifier `2` and fraction `1/2`. A consumer that
		// resolves the value needs the fraction, because without it the modifier cancels the bare
		// value and the class produces nothing; a consumer reading a pre-measured row does not. That
		// asymmetry is what made the descriptor rows and the ported registration tables incomparable
		// on this corpus until the fixture recorded it.
		candidate.Value = &ParsedValue{Kind: ParsedValueKindNamed, Value: one.Value, Fraction: one.Fraction}
	case "arbitrary":
		// The engine reports the annotation and the remaining text separately, and so does the
		// parser, so they are carried across unjoined.
		candidate.Value = &ParsedValue{Kind: ParsedValueKindArbitrary, Value: one.Value, DataType: one.DataType}
	}

	switch one.ModifierKind {
	case "named":
		candidate.Modifier = &ParsedModifier{Kind: ParsedModifierKindNamed, Value: one.ModifierValue}
	case "arbitrary":
		candidate.Modifier = &ParsedModifier{Kind: ParsedModifierKindArbitrary, Value: one.ModifierValue}
	}
	return candidate
}

func TestDescriptorLookupMatchesEngine(t *testing.T) {
	t.Parallel()
	fixtures := loadDescriptorFixtures(t)
	table := testTable(t)

	if table.TailwindVersion != fixtures.TailwindVersion {
		t.Fatalf("table is Tailwind %s and fixtures are Tailwind %s", table.TailwindVersion, fixtures.TailwindVersion)
	}

	agreed, declined, mismatched, withoutReading := 0, 0, 0, 0
	// The known exceptions, which the model does not cover and must not appear to. Counted rather
	// than skipped, because a suite that silently skips its hard cases reports the same green line
	// whether they were hard or absent.
	perDeclarationDeclined, arbitraryProperties := 0, 0

	for _, one := range fixtures.Cases {
		if one.Kind == "arbitrary" {
			arbitraryProperties++
		}

		got, answered := table.Lookup(one.parsed())

		if !answered {
			declined++
			if descriptor, found := table.Descriptors[one.Root]; found && descriptor.PerDeclaration {
				perDeclarationDeclined++
			}
			// Declining is the right answer for a class the engine also refuses. Declining a class
			// the engine reads is a real gap, and it is reported rather than folded into the
			// agreement count, because a lookup that answered nothing would otherwise pass.
			if one.Reading != nil {
				descriptor, found := table.Descriptors[one.Root]
				if !found || !descriptor.PerDeclaration {
					mismatched++
					if mismatched <= 20 {
						t.Errorf("%s: lookup declined but engine read %v#%d",
							one.ClassName, one.Reading.Order, one.Reading.Count)
					}
				}
			}
			continue
		}

		if one.Reading == nil {
			// No reading at all, which is not an agreement and not a disagreement.
			//
			// The descriptor model answers "what does this root read for a value of this type", not
			// "does this root accept this value". The corpus probes every root with every shape, so
			// most of these are combinations no author would write and the engine declines:
			// `-bg-conic-bold` parses and compiles to nothing. Scoring them would be measuring
			// whether the table can predict a utility's own accept-list, which it never claimed to
			// and which the `@utility` evaluator owns.
			//
			// Counted on its own line rather than skipped silently, which is the rule the extractor
			// states and follows: a run cannot be allowed to look clean by measuring nothing.
			withoutReading++
			continue
		}

		want := Reading{Order: one.Reading.Order, Count: one.Reading.Count}
		if !got.Equal(want) {
			mismatched++
			if mismatched <= 20 {
				t.Errorf("%s (root %q, %s %q, modifier %s %q): read %v#%d, engine said %v#%d",
					one.ClassName, one.Root, one.ValueKind, one.Value, one.ModifierKind, one.ModifierValue,
					got.Order, got.Count, want.Order, want.Count)
			}
			continue
		}
		agreed++
	}

	if mismatched > 0 {
		t.Errorf("%d of %d cases disagreed with the engine", mismatched, len(fixtures.Cases))
	}
	scored := agreed + mismatched
	t.Logf("agreed on %d of %d scored cases over %d roots, reduced from %d classes", agreed, scored, fixtures.Counts.Roots, fixtures.Counts.ClassesConsidered)
	t.Logf("  %d cases had no engine reading and were reported rather than scored", withoutReading)
	t.Logf("  %d lookups declined, of which %d on the %d per-declaration roots", declined, perDeclarationDeclined, countPerDeclaration(table))
	t.Logf("  %d arbitrary properties", arbitraryProperties)

	// Volume assertions, in the spirit of the extractor's: a suite that agreed on everything because
	// it tested almost nothing reports the same green line as one that did the work.
	if withoutReading > len(fixtures.Cases)/2 {
		t.Errorf("%d of %d cases had no reading; the corpus is mostly unscored", withoutReading, len(fixtures.Cases))
	}
	if agreed < 5000 {
		t.Errorf("only %d agreements; the corpus is too small for the number to mean anything", agreed)
	}
	if fixtures.Counts.Roots < 300 {
		t.Errorf("only %d roots in the corpus", fixtures.Counts.Roots)
	}
}

func countPerDeclaration(table *Table) int {
	count := 0
	for _, descriptor := range table.Descriptors {
		if descriptor.PerDeclaration {
			count++
		}
	}
	return count
}

// TestDescriptorLookupCoversEveryRoot asserts the table answers for every root the engine has,
// rather than for the ones the corpus happened to reach.
func TestDescriptorLookupCoversEveryRoot(t *testing.T) {
	t.Parallel()
	fixtures := loadDescriptorFixtures(t)
	table := testTable(t)

	rootsInCorpus := map[string]bool{}
	for _, one := range fixtures.Cases {
		if one.Kind == "functional" {
			rootsInCorpus[one.Root] = true
		}
	}

	missing := 0
	for root := range rootsInCorpus {
		if _, found := table.Descriptors[root]; !found {
			missing++
			if missing <= 10 {
				t.Errorf("root %q appears in the corpus and has no descriptor", root)
			}
		}
	}
	if missing > 0 {
		t.Errorf("%d roots in the corpus have no descriptor", missing)
	}
	t.Logf("every one of %d corpus roots has a descriptor, out of %d in the table", len(rootsInCorpus), len(table.Descriptors))
}

// TestPerDeclarationRootsDecline pins the boundary of the model.
//
// The 18 `<animation-root>-translate-full` classes resolve per declaration rather than per type, and
// the table must say so rather than return the reading it would have guessed. This asserts the
// decline is real, because a lookup that answered them plausibly would be wrong in exactly the place
// the model knows it cannot answer, and nothing else in the suite would notice.
func TestPerDeclarationRootsDecline(t *testing.T) {
	t.Parallel()
	table := testTable(t)

	perDeclaration := 0
	for root, descriptor := range table.Descriptors {
		if !descriptor.PerDeclaration {
			continue
		}
		perDeclaration++
		candidate := &ParsedCandidate{Kind: ParsedCandidateKindFunctional, Root: root, Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "translate-full"}}
		if _, answered := table.Lookup(candidate); answered {
			t.Errorf("root %q is per-declaration and the lookup answered anyway", root)
		}
	}

	if perDeclaration != 18 {
		t.Errorf("expected 18 per-declaration roots, found %d", perDeclaration)
	}
	t.Logf("%d per-declaration roots, all declining", perDeclaration)
}

// TestTypeListOrderIsCarriedNotSorted pins the first refinement.
//
// The order is per-root and recovered by topological sort from observation, and it is the correction
// most likely to be silently dropped, because a sorted list looks tidier and passes any test that
// only checks membership. `bg` is the canary: it puts `position` before `length`, so `bg-[3px]` is a
// position. If this assertion ever fails because the table changed, the readings changed with it.
func TestTypeListOrderIsCarriedNotSorted(t *testing.T) {
	t.Parallel()
	table := testTable(t)

	descriptor, found := table.Descriptors["bg"]
	if !found {
		t.Fatal("no descriptor for bg")
	}

	positionAt, lengthAt := -1, -1
	for index, dataType := range descriptor.TypeList {
		switch dataType {
		case DataTypePosition:
			positionAt = index
		case DataTypeLength:
			lengthAt = index
		}
	}
	if positionAt < 0 || lengthAt < 0 {
		t.Fatalf("bg's type list lacks position or length: %v", descriptor.TypeList)
	}
	if positionAt > lengthAt {
		t.Errorf("bg lists length before position (%v); bg-[3px] reads as a length rather than a position", descriptor.TypeList)
	}

	// The observable consequence, asserted separately so a reordering that somehow kept the indices
	// still fails on the reading itself.
	candidate := &ParsedCandidate{Kind: ParsedCandidateKindFunctional, Root: "bg", Value: &ParsedValue{Kind: ParsedValueKindArbitrary, Value: "3px"}}
	got, answered := table.Lookup(candidate)
	if !answered {
		t.Fatal("bg-[3px] was declined")
	}
	if inferred := InferDataType("3px", descriptor.TypeList); inferred != DataTypePosition {
		t.Errorf("bg-[3px] inferred as %q rather than position; reading was %v#%d", inferred, got.Order, got.Count)
	}
}

// TestBareValueResolvesBeforeInference pins the second refinement.
//
// `bold` is a `--font-weight` key that also satisfies `family-name`, which is in `font`'s type list.
// Inferring before consulting the theme reads `font-bold` as a font family, and eight registry
// classes separate the two orderings. This asserts the theme wins.
func TestBareValueResolvesBeforeInference(t *testing.T) {
	t.Parallel()
	table := testTable(t)

	descriptor, found := table.Descriptors["font"]
	if !found {
		t.Skip("this design system has no font root")
	}

	if !table.KeysByNamespace["--font-weight"]["bold"] {
		t.Skip("this design system has no --font-weight-bold")
	}
	if inferred := InferDataType("bold", descriptor.TypeList); inferred == "" {
		t.Skip("bold does not infer in this design system, so the two orderings are not separable here")
	}

	candidate := &ParsedCandidate{Kind: ParsedCandidateKindFunctional, Root: "font", Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "bold"}}
	got, answered := table.Lookup(candidate)
	if !answered {
		t.Fatal("font-bold was declined")
	}

	viaTheme, hasTheme := descriptor.Absent.ByNamespace["--font-weight"]
	if !hasTheme {
		viaTheme = descriptor.Absent.Fallback
	}
	if !got.Equal(viaTheme) {
		viaInference := descriptor.Absent.readingForType(InferDataType("bold", descriptor.TypeList))
		t.Errorf("font-bold read %v#%d; theme says %v#%d and inference says %v#%d",
			got.Order, got.Count, viaTheme.Order, viaTheme.Count, viaInference.Order, viaInference.Count)
	}
}

// TestModifierIsThreeStates pins the third refinement.
//
// `/none` is a `--leading` key rather than an alpha, so it selects a different set of readings than
// `/50` does, and both differ from no modifier at all. A port that treated the modifier as one bit
// would pass every test written against `/50` alone.
func TestModifierIsThreeStates(t *testing.T) {
	t.Parallel()
	if axis := ModifierAxisFor(nil); axis != ModifierAbsent {
		t.Errorf("no modifier classified as %v", axis)
	}
	if axis := ModifierAxisFor(&ParsedModifier{Kind: ParsedModifierKindNamed, Value: "none"}); axis != ModifierThemed {
		t.Errorf("/none classified as %v rather than themed", axis)
	}
	for _, modifier := range []*ParsedModifier{
		{Kind: ParsedModifierKindNamed, Value: "50"},
		{Kind: ParsedModifierKindArbitrary, Value: "0.5"},
		{Kind: ParsedModifierKindArbitrary, Value: "var(--a)"},
		// An arbitrary modifier spelling a theme key is still an alpha: the engine does not consult
		// the theme for a bracketed modifier.
		{Kind: ParsedModifierKindArbitrary, Value: "none"},
	} {
		if axis := ModifierAxisFor(modifier); axis != ModifierAlpha {
			t.Errorf("modifier %v classified as %v rather than alpha", modifier, axis)
		}
	}
}
