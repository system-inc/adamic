package tailwind

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The fixture is what the shipped Tailwind 4.3.3 engine answered about variant ordering, captured
// by internal/lint/rules/tailwind/tools/generate_variant and checked in next to this test.
//
// Measured rather than transcribed, because the two are different claims and `getVariantOrder` is
// where they part. Reading `variants.ts` shows every variant holding an `order` number and invites
// a static table of variant to position; cohere has shipped exactly that table, 145 entries. Asking
// the engine says the number a consumer sees is a dense rank assigned per run over only the
// variants that run parsed, with ties collapsed. A transcribed table agrees with the engine on any
// corpus whose variant set matches the one it was generated from, which is why a table generated
// from this repository passed on this repository for as long as it did.
//
// The corpus has three halves and needs all three. The repository half is the exit criterion: 2,400
// real class lists from two repositories, sorted by the engine, each compared class by class. The
// synthetic half reaches the pairs a real corpus never writes, and one of those pairs is what
// answers this component's open question. The `@custom-variant` half separates a custom variant
// that appends a new position from one that overwrites a framework name in place, which look the
// same in a stylesheet and behave differently in the sort.

type variantOrderCorpus struct {
	TailwindVersion     string             `json:"tailwindVersion"`
	SyntheticCaseCount  int                `json:"syntheticCaseCount"`
	RepositoryCaseCount int                `json:"repositoryCaseCount"`
	Cases               []variantOrderCase `json:"cases"`
}

type variantOrderCase struct {
	Name       string                         `json:"name"`
	Source     string                         `json:"source"`
	Input      string                         `json:"input"`
	EntryPath  string                         `json:"entryPath"`
	Registry   *variantOrderRegistryFixture   `json:"registry"`
	VariantOrd *variantOrderIndicesFixture    `json:"variantOrder"`
	ClassOrder *variantOrderClassOrderFixture `json:"classOrder"`
	ClassLists []variantOrderClassListFixture `json:"classLists"`
}

type variantOrderRegistryFixture struct {
	Entries []struct {
		Name  string            `json:"name"`
		Order int               `json:"order"`
		Kind  ParsedVariantKind `json:"kind"`
	} `json:"entries"`
	CompareFnOrders []int `json:"compareFnOrders"`
}

type variantOrderIndicesFixture struct {
	Size      int `json:"size"`
	Positions []struct {
		Raw     string                    `json:"raw"`
		Variant *variantOrderShapeFixture `json:"variant"`
		Index   *int                      `json:"index"`
	} `json:"positions"`
}

type variantOrderClassOrderFixture struct {
	Ranks []struct {
		ClassName string  `json:"className"`
		Rank      *string `json:"rank"`
	} `json:"ranks"`
	Masks []struct {
		ClassName string  `json:"className"`
		Mask      *string `json:"mask"`
		Variants  []struct {
			Variant *variantOrderShapeFixture `json:"variant"`
			Index   *int                      `json:"index"`
		} `json:"variants"`
	} `json:"masks"`
	Sorted []string `json:"sorted"`
}

type variantOrderClassListFixture struct {
	Name       string                        `json:"name"`
	Classes    []string                      `json:"classes"`
	ClassOrder variantOrderClassOrderFixture `json:"classOrder"`
}

// variantOrderShapeFixture is the engine's own description of a parsed variant.
//
// Read into the port's ParsedVariant by variantOrderVariantFromFixture rather than compared field by field,
// because what is under test here is the ordering of variants and not the parsing of them. The
// candidate parser is a sibling component with its own fixture; taking its answers from this
// fixture instead of re-deriving them keeps a parser defect from reading as an ordering defect.
type variantOrderShapeFixture struct {
	Kind     ParsedVariantKind         `json:"kind"`
	Root     string                    `json:"root"`
	Selector string                    `json:"selector"`
	Relative bool                      `json:"relative"`
	Value    *variantOrderValueFixture `json:"value"`
	Modifier *variantOrderValueFixture `json:"modifier"`
	Variant  *variantOrderShapeFixture `json:"variant"`
}

type variantOrderValueFixture struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

func variantOrderVariantFromFixture(fixture *variantOrderShapeFixture) *ParsedVariant {
	if fixture == nil {
		return nil
	}
	variant := &ParsedVariant{
		Kind:     fixture.Kind,
		Root:     fixture.Root,
		Selector: fixture.Selector,
		Relative: fixture.Relative,
		Variant:  variantOrderVariantFromFixture(fixture.Variant),
	}
	if fixture.Value != nil {
		variant.Value = &ParsedVariantValue{Kind: ParsedValueKind(fixture.Value.Kind), Value: fixture.Value.Value}
	}
	if fixture.Modifier != nil {
		variant.Modifier = &ParsedModifier{Kind: ParsedModifierKind(fixture.Modifier.Kind), Value: fixture.Modifier.Value}
	}
	return variant
}

func loadVariantOrderCorpus(t *testing.T) variantOrderCorpus {
	t.Helper()

	path := filepath.Join("testdata", "variant_fixtures.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var corpus variantOrderCorpus
	if err := json.Unmarshal(contents, &corpus); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if corpus.TailwindVersion != "4.3.3" {
		t.Fatalf("the fixture was captured from Tailwind %s and this port targets 4.3.3", corpus.TailwindVersion)
	}
	if len(corpus.Cases) == 0 {
		t.Fatal("the fixture holds no cases, so every assertion below would pass vacuously")
	}
	return corpus
}

// variantOrderRegistryFrom rebuilds the engine's registry from a fixture.
//
// The order numbers are taken from the fixture rather than reproduced by replaying registrations,
// because replaying them would mean porting the 1,000 lines of `variants.ts` that build CSS in
// order to reach the `variants.static(...)` calls buried in them. What this component owns is what
// the registry does with those numbers, so the numbers are input and the comparison is the test.
//
// The comparison functions are attached here, keyed by order, because the fixture records which
// orders carry one but cannot record the function itself. All four are breakpoint comparisons in a
// default build, and their directions come from `variants.ts`: `max` and `@max` descend, the
// breakpoint group and `@`/`@min` ascend.
func variantOrderRegistryFrom(t *testing.T, fixture *variantOrderRegistryFixture, resolveBreakpoint func(ParsedVariant) (string, bool)) *VariantRegistry {
	t.Helper()

	registry := NewVariantRegistry()
	descending := map[int]bool{}
	for _, entry := range fixture.Entries {
		registry.registrations[entry.Name] = VariantRegistration{Name: entry.Name, Order: entry.Order, Kind: entry.Kind}
		if entry.Order > registry.lastOrder {
			registry.lastOrder = entry.Order
		}
		if entry.Name == "max" || entry.Name == "@max" {
			descending[entry.Order] = true
		}
	}

	for _, order := range fixture.CompareFnOrders {
		ascending := !descending[order]
		registry.comparisons[order] = func(left ParsedVariant, right ParsedVariant) int {
			leftValue, leftResolves := resolveBreakpoint(left)
			if !leftResolves {
				if ascending {
					return -1
				}
				return 1
			}
			rightValue, rightResolves := resolveBreakpoint(right)
			if !rightResolves {
				if ascending {
					return 1
				}
				return -1
			}
			return CompareBreakpoints(leftValue, rightValue, ascending)
		}
	}

	return registry
}

// TestVariantOrderMatchesTheEngine is the differential: every variant order and every class order
// the engine reported, reproduced by the port.
//
// The counts are printed on success. A suite that compared twelve answers and one that compared
// twenty thousand look the same as a green line, and this component's whole claim is about a
// population rather than about a handful of cases.
func TestVariantOrderMatchesTheEngine(t *testing.T) {
	t.Parallel()
	corpus := loadVariantOrderCorpus(t)

	comparedVariantOrders := 0
	comparedIndices := 0
	comparedClassLists := 0
	comparedMasks := 0
	comparedRanks := 0

	for _, aCase := range corpus.Cases {
		registryFixture := variantOrderRegistryForCase(t, corpus, aCase)
		if registryFixture == nil {
			continue
		}
		resolveBreakpoint := variantOrderBreakpointResolverFor(aCase, corpus)
		registry := variantOrderRegistryFrom(t, registryFixture, resolveBreakpoint)

		if aCase.VariantOrd != nil {
			comparedVariantOrders++
			comparedIndices += checkVariantOrderIndices(t, aCase.Name, registry, aCase.VariantOrd)
		}

		if aCase.ClassOrder != nil {
			comparedClassLists++
			masks, ranks := checkVariantOrderClassOrder(t, aCase.Name, registry, *aCase.ClassOrder)
			comparedMasks += masks
			comparedRanks += ranks
		}

		for _, classList := range aCase.ClassLists {
			comparedClassLists++
			masks, ranks := checkVariantOrderClassOrder(t, classList.Name, registry, classList.ClassOrder)
			comparedMasks += masks
			comparedRanks += ranks
		}
	}

	if comparedIndices == 0 || comparedMasks == 0 {
		t.Fatal("the differential compared nothing, so it cannot have found agreement")
	}

	t.Logf(
		"tailwind %s: %d variant-order cases (%d indices), %d class lists (%d masks, %d ranks)",
		corpus.TailwindVersion, comparedVariantOrders, comparedIndices,
		comparedClassLists, comparedMasks, comparedRanks,
	)
}

// variantOrderRegistryForCase finds the registry a case should be compared under.
//
// A repository's class lists live in a case of their own, separate from the case carrying that
// repository's registry, so the lists are matched back to their repository by name prefix. Falling
// back to the framework registry rather than skipping is deliberate: a synthetic case is built on a
// default design system and has no registry of its own.
func variantOrderRegistryForCase(t *testing.T, corpus variantOrderCorpus, aCase variantOrderCase) *variantOrderRegistryFixture {
	t.Helper()

	if aCase.Registry != nil {
		return aCase.Registry
	}
	if aCase.VariantOrd == nil && aCase.ClassOrder == nil && len(aCase.ClassLists) == 0 {
		return nil
	}

	repository := strings.TrimSuffix(aCase.Name, "/class-lists")
	for _, candidate := range corpus.Cases {
		if candidate.Registry != nil && candidate.Name == repository {
			return candidate.Registry
		}
	}
	for _, candidate := range corpus.Cases {
		if candidate.Registry != nil && candidate.Name == "framework-default" {
			return candidate.Registry
		}
	}
	t.Fatalf("%s: no registry to compare under", aCase.Name)
	return nil
}

// variantOrderBreakpointResolverFor answers what a breakpoint variant resolves to, from the fixture's own data.
//
// The engine resolves these through the theme's `--breakpoint` and `--container` namespaces, which
// is a sibling component. Rather than depend on it, the resolver reads the value straight off the
// variant when it is arbitrary and falls back to the default scale by name, which is what both
// repositories in the corpus use. A name this cannot resolve reports false, which is upstream's own
// `null` branch and is ordered rather than skipped.
func variantOrderBreakpointResolverFor(aCase variantOrderCase, corpus variantOrderCorpus) func(ParsedVariant) (string, bool) {
	return func(variant ParsedVariant) (string, bool) {
		switch variant.Kind {
		case ParsedVariantKindStatic:
			value, isNamed := variantOrderDefaultBreakpointScale[variant.Root]
			return value, isNamed
		case ParsedVariantKindFunctional:
			if variant.Value == nil || variant.Modifier != nil && !strings.HasPrefix(variant.Root, "@") {
				return "", false
			}
			if variant.Value.Kind == ParsedValueKindArbitrary {
				if strings.Contains(variant.Value.Value, "var(") {
					return "", false
				}
				return variant.Value.Value, true
			}
			scale := variantOrderDefaultBreakpointScale
			if strings.HasPrefix(variant.Root, "@") {
				scale = variantOrderDefaultContainerScale
			}
			value, isNamed := scale[variant.Value.Value]
			return value, isNamed
		default:
			return "", false
		}
	}
}

// variantOrderDefaultBreakpointScale is Tailwind 4.3.3's own `--breakpoint-*`, which both corpus repositories
// inherit unchanged.
//
// Hardcoded here and only here, in the test's resolver, never in variant.go. The port reads
// breakpoints from the theme; this table exists so the differential can answer the same question
// without taking a dependency on a sibling component's port.
var variantOrderDefaultBreakpointScale = map[string]string{
	"sm": "40rem", "md": "48rem", "lg": "64rem", "xl": "80rem", "2xl": "96rem",
}

// variantOrderDefaultContainerScale is Tailwind 4.3.3's `--container-*`, for the `@` container-query variants.
var variantOrderDefaultContainerScale = map[string]string{
	"3xs": "16rem", "2xs": "18rem", "xs": "20rem", "sm": "24rem", "md": "28rem",
	"lg": "32rem", "xl": "36rem", "2xl": "42rem", "3xl": "48rem", "4xl": "56rem",
	"5xl": "64rem", "6xl": "72rem", "7xl": "80rem",
}

// checkVariantOrderIndices compares the port's dense indices against the engine's, for one case.
func checkVariantOrderIndices(t *testing.T, name string, registry *VariantRegistry, fixture *variantOrderIndicesFixture) int {
	t.Helper()

	variants := make([]ParsedVariant, 0, len(fixture.Positions))
	for _, position := range fixture.Positions {
		if position.Variant == nil {
			continue
		}
		variants = append(variants, *variantOrderVariantFromFixture(position.Variant))
	}

	order := registry.BuildVariantOrder(variants)
	if order.Len() != fixture.Size {
		t.Errorf("%s: the port ordered %d variants and the engine ordered %d", name, order.Len(), fixture.Size)
	}

	compared := 0
	for _, position := range fixture.Positions {
		if position.Variant == nil || position.Index == nil {
			continue
		}
		got, hasIndex := order.IndexOf(*variantOrderVariantFromFixture(position.Variant))
		if !hasIndex {
			t.Errorf("%s: the port assigned no index to %q, which the engine placed at %d", name, position.Raw, *position.Index)
			continue
		}
		if got != *position.Index {
			t.Errorf("%s: %q is at %d in the port and %d in the engine", name, position.Raw, got, *position.Index)
		}
		compared++
	}
	return compared
}

// checkVariantOrderClassOrder compares the port's bitmasks and resulting order against the engine's.
//
// Both are compared because they answer different questions. The mask is the mechanism, and two
// ports can agree on every rank in a corpus while building the mask differently. The rank is the
// observable behaviour, and a mask that is right in a way that does not reach the sort is not
// worth much either.
func checkVariantOrderClassOrder(t *testing.T, name string, registry *VariantRegistry, fixture variantOrderClassOrderFixture) (int, int) {
	t.Helper()

	// Every variant the engine saw for this class list, which is the population its dense indices
	// were assigned over. Taking it from the fixture rather than re-parsing keeps this a test of
	// ordering rather than of the sibling parser.
	var population []ParsedVariant
	for _, mask := range fixture.Masks {
		for _, entry := range mask.Variants {
			if entry.Variant == nil {
				continue
			}
			population = append(population, *variantOrderVariantFromFixture(entry.Variant))
		}
	}
	order := registry.BuildVariantOrder(population)

	comparedMasks := 0
	masksByClass := map[string]*big.Int{}
	for _, maskFixture := range fixture.Masks {
		if maskFixture.Mask == nil {
			continue
		}
		variants := make([]ParsedVariant, 0, len(maskFixture.Variants))
		for _, entry := range maskFixture.Variants {
			if entry.Variant == nil {
				continue
			}
			variants = append(variants, *variantOrderVariantFromFixture(entry.Variant))
		}

		got, isComplete := order.VariantBitmask(variants)
		if !isComplete {
			t.Errorf("%s: the port could not build a mask for %q", name, maskFixture.ClassName)
			continue
		}
		want, isNumeric := new(big.Int).SetString(*maskFixture.Mask, 10)
		if !isNumeric {
			t.Fatalf("%s: the fixture holds an unreadable mask %q for %q", name, *maskFixture.Mask, maskFixture.ClassName)
		}
		if got.Cmp(want) != 0 {
			t.Errorf("%s: %q has mask %s in the port and %s in the engine", name, maskFixture.ClassName, got, want)
		}
		masksByClass[maskFixture.ClassName] = got
		comparedMasks++
	}

	// The engine's own sorted list, reproduced by sorting on the masks alone. Classes whose masks
	// tie keep their relative order from the engine's list, which is what isolates this comparison
	// to the variant dimension: the property dimensions that break those ties belong to sibling
	// components and are not under test here.
	sorted := make([]string, 0, len(fixture.Sorted))
	for _, className := range fixture.Sorted {
		if _, hasMask := masksByClass[className]; hasMask {
			sorted = append(sorted, className)
		}
	}
	sort.SliceStable(sorted, func(left int, right int) bool {
		return CompareVariantBitmasks(masksByClass[sorted[left]], masksByClass[sorted[right]]) < 0
	})

	// The engine's own list, restricted to the classes that have masks, is what the port's sorted
	// list is compared against position by position.
	expected := make([]string, 0, len(fixture.Sorted))
	for _, className := range fixture.Sorted {
		if _, hasMask := masksByClass[className]; hasMask {
			expected = append(expected, className)
		}
	}

	for index := range sorted {
		if index < len(expected) && expected[index] != sorted[index] {
			t.Errorf("%s: position %d is %q in the port and %q in the engine", name, index, sorted[index], expected[index])
		}
	}

	return comparedMasks, len(sorted)
}

// TestVariantBitmaskIsNotDepthFirst is the answer to this component's open question, and it is a
// divergence.
//
// enforce_consistent_class_order.go compares variants segment by segment and applies an explicit
// rule before any segment is read: "Depth first: every single variant precedes every stacked one,
// whatever they start with." That rule was arrived at by fixture failure, and its comment names the
// evidence honestly as four real class lists.
//
// The engine implements no such rule. `compile.ts` OR-s one bit per variant into a mask and
// compares the masks numerically, so a class's position is decided by its highest-ranked variant
// and only ties on that are broken lower down. A stacked variant whose members all rank low
// therefore sorts BEFORE a single variant that ranks high.
//
// Measured on the engine: `group-hover:disabled:` carries bits 0 and 2, mask 5. `dark:` carries bit
// 3, mask 8. The stacked class sorts first, and the depth-first rule puts it second.
func TestVariantBitmaskIsNotDepthFirst(t *testing.T) {
	t.Parallel()
	corpus := loadVariantOrderCorpus(t)

	var probe *variantOrderCase
	for index, aCase := range corpus.Cases {
		if aCase.Name == "class-order/stacked-against-single" {
			probe = &corpus.Cases[index]
			break
		}
	}
	if probe == nil || probe.ClassOrder == nil {
		t.Fatal("the fixture no longer holds the stacked-against-single case, which is what this assertion measures")
	}

	if len(probe.ClassOrder.Sorted) != 2 {
		t.Fatalf("expected two classes, got %v", probe.ClassOrder.Sorted)
	}

	stacked := "group-hover:disabled:flex"
	single := "dark:flex"
	if probe.ClassOrder.Sorted[0] != stacked || probe.ClassOrder.Sorted[1] != single {
		t.Fatalf(
			"the engine no longer sorts the stacked variant first: got %v. If Tailwind changed this, "+
				"the divergence this test records has closed and the finding needs rewriting rather "+
				"than the assertion flipping",
			probe.ClassOrder.Sorted,
		)
	}

	// The mechanism, not just the outcome: the stacked class's mask must be numerically smaller
	// despite naming more variants. Asserting only the order would pass for a port that happened to
	// be depth-last rather than mask-ordered.
	masks := map[string]*big.Int{}
	for _, mask := range probe.ClassOrder.Masks {
		if mask.Mask == nil {
			continue
		}
		value, _ := new(big.Int).SetString(*mask.Mask, 10)
		masks[mask.ClassName] = value
	}
	if masks[stacked] == nil || masks[single] == nil {
		t.Fatal("both classes should have masks")
	}
	if masks[stacked].BitLen() >= masks[single].BitLen() {
		t.Errorf(
			"the stacked class should carry only lower bits: %s is %s and %s is %s",
			stacked, masks[stacked], single, masks[single],
		)
	}
	if CompareVariantBitmasks(masks[stacked], masks[single]) >= 0 {
		t.Errorf("mask comparison should place %s first", stacked)
	}
}

// TestDepthFirstDisagreesWithTheEngineOnTheCorpus quantifies the divergence.
//
// The open question this component was dispatched to answer is whether the segment comparison that
// enforce_consistent_class_order.go used to carry and upstream's bitmask are the same rule. They are
// not, and a disagreement is only worth reporting with a population attached, so this walks every
// ordered pair in every class list in the corpus and counts the pairs on which a depth-first rule and
// the mask comparison disagree.
//
// The count is logged rather than asserted to be zero, because a nonzero count is the finding. What
// is asserted is that the corpus was actually walked and that the disagreement is real on at least
// the synthetic probes, so a future change that silently emptied the corpus could not read as
// agreement.
//
// # This measures a rule, not the shipped code, and that stayed true on purpose
//
// #4q5dsn3 swapped the rule onto the live design system, so the depth-first comparison this counts
// against no longer ships anywhere: `compareVariants`, `variantSegments` and the rest of that path
// were deleted with it. What this test compares is a depth-first rule REIMPLEMENTED here, in the loop
// below, against the engine's masks from the fixture.
//
// It is kept in that form rather than repointed at the new code, because the two answer different
// questions and both are worth keeping. This one is the standing statement that depth-first is wrong
// and by how much: 15 pairs of 4,265, the number #nr3wtj1 filed. Repointing it at the live sort would
// turn it into a second copy of
// `internal/rules/tailwind`'s TestClassOrderLiveMatchesTheEngineOverTheCorpus, which measures the
// shipped path over the same corpus and reports 0 of 28,133, and this file would lose the record of
// what the divergence was.
//
// So a nonzero count here is correct and expected forever. A zero would mean the corpus lost its
// stacked-variant cases, which is what the assertion at the bottom exists to catch.
func TestDepthFirstDisagreesWithTheEngineOnTheCorpus(t *testing.T) {
	t.Parallel()
	corpus := loadVariantOrderCorpus(t)

	comparedPairs := 0
	disagreeingPairs := 0
	disagreeingLists := 0
	var examples []string

	for _, aCase := range corpus.Cases {
		classOrders := []struct {
			name    string
			fixture variantOrderClassOrderFixture
		}{}
		if aCase.ClassOrder != nil {
			classOrders = append(classOrders, struct {
				name    string
				fixture variantOrderClassOrderFixture
			}{aCase.Name, *aCase.ClassOrder})
		}
		for _, classList := range aCase.ClassLists {
			classOrders = append(classOrders, struct {
				name    string
				fixture variantOrderClassOrderFixture
			}{classList.Name, classList.ClassOrder})
		}

		for _, classOrder := range classOrders {
			depthByClass := map[string]int{}
			maskByClass := map[string]*big.Int{}
			for _, mask := range classOrder.fixture.Masks {
				if mask.Mask == nil {
					continue
				}
				value, isNumeric := new(big.Int).SetString(*mask.Mask, 10)
				if !isNumeric {
					continue
				}
				maskByClass[mask.ClassName] = value
				depthByClass[mask.ClassName] = len(mask.Variants)
			}

			listDisagrees := false
			ordered := classOrder.fixture.Sorted
			for left := 0; left < len(ordered); left++ {
				for right := left + 1; right < len(ordered); right++ {
					leftClass, rightClass := ordered[left], ordered[right]
					leftMask, leftHasMask := maskByClass[leftClass]
					rightMask, rightHasMask := maskByClass[rightClass]
					if !leftHasMask || !rightHasMask {
						continue
					}
					// Only pairs the mask actually separates are counted. A pair the engine ties on
					// is decided by properties, which is a different component's dimension, and
					// counting it here would dilute the measurement with pairs neither rule claims.
					if CompareVariantBitmasks(leftMask, rightMask) == 0 {
						continue
					}
					comparedPairs++

					// The depth-first rule as enforce_consistent_class_order.go states it: fewer
					// variant segments sorts first, whatever they are.
					leftDepth, rightDepth := depthByClass[leftClass], depthByClass[rightClass]
					if leftDepth == rightDepth {
						continue
					}
					depthFirstPutsLeftFirst := leftDepth < rightDepth
					// The engine put leftClass first, by construction of `ordered`.
					if !depthFirstPutsLeftFirst {
						disagreeingPairs++
						listDisagrees = true
						if len(examples) < 8 {
							examples = append(examples, fmt.Sprintf(
								"%s: engine puts %q (mask %s, %d variants) before %q (mask %s, %d variants); depth-first reverses them",
								classOrder.name, leftClass, leftMask, leftDepth, rightClass, rightMask, rightDepth,
							))
						}
					}
				}
			}
			if listDisagrees {
				disagreeingLists++
			}
		}
	}

	if comparedPairs == 0 {
		t.Fatal("no separable pairs were compared, so this measured nothing")
	}

	t.Logf(
		"depth-first vs the engine's bitmask: %d disagreeing pairs of %d separable pairs (%.4f%%), across %d class lists",
		disagreeingPairs, comparedPairs, 100*float64(disagreeingPairs)/float64(comparedPairs), disagreeingLists,
	)
	for _, example := range examples {
		t.Log("  " + example)
	}

	if disagreeingPairs == 0 {
		t.Error(
			"depth-first agreed with the engine on every pair in the corpus, which contradicts the " +
				"synthetic probe in TestVariantBitmaskIsNotDepthFirst. Either the corpus lost its " +
				"stacked-variant cases or the mask comparison stopped being a mask comparison",
		)
	}
}

// TestRegistrationOrderIsSharedWithinAGroup pins the tie that `Variants.group` creates.
//
// In a default build `sm`, `md`, `lg`, `xl`, `2xl` and `min` all hold one order number, and four
// order numbers carry comparison functions. A port that gave each breakpoint its own position would
// pass every test built from a default theme and diverge on any repository that redefines a
// breakpoint, which is the per-repository failure this whole port exists to remove.
func TestRegistrationOrderIsSharedWithinAGroup(t *testing.T) {
	t.Parallel()
	corpus := loadVariantOrderCorpus(t)

	var framework *variantOrderRegistryFixture
	for _, aCase := range corpus.Cases {
		if aCase.Name == "framework-default" && aCase.Registry != nil {
			framework = aCase.Registry
			break
		}
	}
	if framework == nil {
		t.Fatal("the fixture no longer holds the framework-default registry")
	}

	orderByName := map[string]int{}
	namesByOrder := map[int][]string{}
	for _, entry := range framework.Entries {
		orderByName[entry.Name] = entry.Order
		namesByOrder[entry.Order] = append(namesByOrder[entry.Order], entry.Name)
	}

	breakpoints := []string{"sm", "md", "lg", "xl", "2xl", "min"}
	shared, isRegistered := orderByName["sm"]
	if !isRegistered {
		t.Fatal("`sm` is not registered, so the breakpoint group is gone")
	}
	for _, name := range breakpoints {
		order, isRegistered := orderByName[name]
		if !isRegistered {
			t.Errorf("%q is not registered", name)
			continue
		}
		if order != shared {
			t.Errorf("%q holds order %d and `sm` holds %d, so the group no longer shares one number", name, order, shared)
		}
	}

	if len(framework.CompareFnOrders) == 0 {
		t.Fatal("no order carries a comparison function, so every shared-order group would tie")
	}
	carriesComparison := false
	for _, order := range framework.CompareFnOrders {
		if order == shared {
			carriesComparison = true
		}
	}
	if !carriesComparison {
		t.Errorf(
			"the breakpoint group holds order %d and the comparison functions are registered against %v, "+
				"so the breakpoints would tie and their classes interleave",
			shared, framework.CompareFnOrders,
		)
	}
}

// TestCustomVariantOverridingAFrameworkNameKeepsItsOrder pins the per-repository behaviour, and it
// is the opposite of what the task description assumed.
//
// A `@custom-variant` under a NEW name appends a position past the framework's last, and genuinely
// changes what a repository's classes sort against. A `@custom-variant` that reuses a framework
// name overwrites the registration in place: `Variants.set` assigns kind, applyFn and compounds
// onto the existing record and never touches `order`.
//
// Both repositories in this corpus do the second kind. `libraries/structure/source/theme/styles/
// global.css` declares `@custom-variant dark` with a `:has()` selector, which completely replaces
// the CSS `dark:` emits, and leaves its sort position exactly where the framework put it. Their
// variant registries are therefore identical to the framework's, which is a real finding rather
// than a missing measurement: for ordering purposes an overriding custom variant is invisible.
func TestCustomVariantOverridingAFrameworkNameKeepsItsOrder(t *testing.T) {
	t.Parallel()
	corpus := loadVariantOrderCorpus(t)

	registries := map[string]*variantOrderRegistryFixture{}
	for index, aCase := range corpus.Cases {
		if aCase.Registry != nil {
			registries[aCase.Name] = corpus.Cases[index].Registry
		}
	}

	framework := registries["framework-default"]
	if framework == nil {
		t.Fatal("the fixture no longer holds the framework-default registry")
	}
	frameworkOrders := map[string]int{}
	for _, entry := range framework.Entries {
		frameworkOrders[entry.Name] = entry.Order
	}

	overriding := registries["custom-variant/custom-variant-overrides-framework"]
	if overriding == nil {
		t.Fatal("the fixture no longer holds the overriding-custom-variant case")
	}
	for _, entry := range overriding.Entries {
		frameworkOrder, isFramework := frameworkOrders[entry.Name]
		if !isFramework {
			t.Errorf("overriding `dark` should register no new name, but %q appeared", entry.Name)
			continue
		}
		if entry.Order != frameworkOrder {
			t.Errorf("%q moved from order %d to %d when `dark` was overridden", entry.Name, frameworkOrder, entry.Order)
		}
	}
	if len(overriding.Entries) != len(framework.Entries) {
		t.Errorf("overriding a framework variant changed the registry size from %d to %d",
			len(framework.Entries), len(overriding.Entries))
	}

	newName := registries["custom-variant/custom-variant-new-name"]
	if newName == nil {
		t.Fatal("the fixture no longer holds the new-name-custom-variant case")
	}
	if len(newName.Entries) != len(framework.Entries)+1 {
		t.Fatalf("a new custom variant should append exactly one registration, got %d against %d",
			len(newName.Entries), len(framework.Entries))
	}
	appended := ""
	appendedOrder := 0
	for _, entry := range newName.Entries {
		if _, isFramework := frameworkOrders[entry.Name]; !isFramework {
			appended = entry.Name
			appendedOrder = entry.Order
		}
	}
	if appended != "sidebar" {
		t.Fatalf("expected `sidebar` to be appended, got %q", appended)
	}
	highestFramework := 0
	for _, order := range frameworkOrders {
		if order > highestFramework {
			highestFramework = order
		}
	}
	if appendedOrder <= highestFramework {
		t.Errorf(
			"a new custom variant should append past the framework's last order %d, but `sidebar` took %d",
			highestFramework, appendedOrder,
		)
	}

	// Both repositories, asserted rather than assumed: this is the claim that a per-repository port
	// is needed at all, and it deserves to fail loudly if a repository ever adds a new variant name.
	for _, repository := range []string{"ahra", "www-connected-app"} {
		registry := registries[repository]
		if registry == nil {
			t.Errorf("%s has no registry in the fixture", repository)
			continue
		}
		if len(registry.Entries) != len(framework.Entries) {
			t.Logf(
				"%s registers %d variants against the framework's %d, so it adds names of its own",
				repository, len(registry.Entries), len(framework.Entries),
			)
		}
	}
}

// TestVariantOrderIsRunScoped is the known-dirty control for the shipped table.
//
// A static variant-to-position table is the shape cohere ships today and the shape this component
// argues against. The argument is only sound if the engine's index genuinely moves, so this asserts
// it does: the same variant must receive different indices in two runs whose populations differ.
//
// Without this, every other assertion here would still pass against a port that memoized one
// global order and got lucky on a corpus.
func TestVariantOrderIsRunScoped(t *testing.T) {
	t.Parallel()
	corpus := loadVariantOrderCorpus(t)

	indicesByCase := map[string]int{}
	for _, aCase := range corpus.Cases {
		if aCase.VariantOrd == nil {
			continue
		}
		for _, position := range aCase.VariantOrd.Positions {
			if position.Raw == "hover" && position.Index != nil {
				indicesByCase[aCase.Name] = *position.Index
			}
		}
	}

	if len(indicesByCase) < 2 {
		t.Fatalf("`hover` appears in %d cases, which cannot show a difference", len(indicesByCase))
	}

	distinct := map[int]bool{}
	for _, index := range indicesByCase {
		distinct[index] = true
	}
	if len(distinct) < 2 {
		t.Errorf(
			"`hover` received index %v in every case, so the order would be expressible as a static "+
				"table and this component's premise is wrong",
			indicesByCase,
		)
	}

	// And the port reproduces the movement rather than only the engine exhibiting it.
	var framework *variantOrderRegistryFixture
	for _, aCase := range corpus.Cases {
		if aCase.Name == "framework-default" && aCase.Registry != nil {
			framework = aCase.Registry
		}
	}
	registry := variantOrderRegistryFrom(t, framework, func(ParsedVariant) (string, bool) { return "", false })

	small := registry.BuildVariantOrder([]ParsedVariant{
		{Kind: ParsedVariantKindStatic, Root: "hover"},
		{Kind: ParsedVariantKindStatic, Root: "focus"},
	})
	large := registry.BuildVariantOrder([]ParsedVariant{
		{Kind: ParsedVariantKindStatic, Root: "hover"},
		{Kind: ParsedVariantKindStatic, Root: "focus"},
		{Kind: ParsedVariantKindCompound, Root: "group", Variant: &ParsedVariant{Kind: ParsedVariantKindStatic, Root: "hover"}},
	})

	smallIndex, _ := small.IndexOf(ParsedVariant{Kind: ParsedVariantKindStatic, Root: "hover"})
	largeIndex, _ := large.IndexOf(ParsedVariant{Kind: ParsedVariantKindStatic, Root: "hover"})
	if smallIndex == largeIndex {
		t.Errorf(
			"the port gave `hover` index %d in both a two-variant and a three-variant run, so it is "+
				"not reproducing the dense reindex",
			smallIndex,
		)
	}
}

// TestTiedVariantsShareAnIndex pins the tie collapse in `getVariantOrder`.
//
// Upstream advances the index only when `compare` says the next variant differs from the previous
// one, so two variants that compare equal share a bit. A port that assigned an index per variant
// would produce masks one bit too wide for every tie, which shifts every higher variant's bit and
// reorders classes that have nothing to do with the tie.
//
// Reaching an actual tie takes more than a shared registration order, and this test asserted the
// wrong thing before it was measured against the engine. `Variants.group` gives its members one
// order number, but `compare` continues past the order to the root name, so two differently-named
// roots in one group still separate. Measured on the engine, `sm` through `2xl` share order 64 and
// receive five distinct indices. A tie needs two variants that compare equal all the way down,
// which is what a group whose comparison function returns zero produces.
func TestTiedVariantsShareAnIndex(t *testing.T) {
	t.Parallel()
	registry := NewVariantRegistry()
	// A comparison function that ties every member, which is what a real tie looks like. Upstream
	// reaches this state through `compareBreakpointVariants` when two breakpoints resolve to the
	// same value.
	registry.Group(func() {
		registry.Register("alpha", ParsedVariantKindStatic)
		registry.Register("beta", ParsedVariantKindStatic)
	}, func(ParsedVariant, ParsedVariant) int { return 0 })
	registry.Register("gamma", ParsedVariantKindStatic)

	alpha, _ := registry.Get("alpha")
	beta, _ := registry.Get("beta")
	gamma, _ := registry.Get("gamma")
	if alpha.Order != beta.Order {
		t.Fatalf("a group should share one order, got %d and %d", alpha.Order, beta.Order)
	}
	if gamma.Order <= alpha.Order {
		t.Fatalf("a registration after a group should append past it, got %d against %d", gamma.Order, alpha.Order)
	}

	order := registry.BuildVariantOrder([]ParsedVariant{
		{Kind: ParsedVariantKindStatic, Root: "alpha"},
		{Kind: ParsedVariantKindStatic, Root: "beta"},
		{Kind: ParsedVariantKindStatic, Root: "gamma"},
	})

	alphaIndex, _ := order.IndexOf(ParsedVariant{Kind: ParsedVariantKindStatic, Root: "alpha"})
	betaIndex, _ := order.IndexOf(ParsedVariant{Kind: ParsedVariantKindStatic, Root: "beta"})
	gammaIndex, _ := order.IndexOf(ParsedVariant{Kind: ParsedVariantKindStatic, Root: "gamma"})

	if alphaIndex != betaIndex {
		t.Errorf("variants whose comparison returns zero should share an index, got %d and %d", alphaIndex, betaIndex)
	}
	if gammaIndex != alphaIndex+1 {
		t.Errorf(
			"the index should advance by one past a tie, so `gamma` should be %d and it is %d",
			alphaIndex+1, gammaIndex,
		)
	}
}

// TestReRegisteringAVariantKeepsItsOrder pins `Variants.set`'s update branch directly.
//
// The fixture cannot reach this. It supplies order numbers the engine already computed, so every
// differential assertion above passes whether Register preserves an order or reassigns one. Only a
// unit test exercises the branch, and the branch is what makes `@custom-variant dark (...)` change
// the selector `dark:` emits without moving where `dark:` sorts.
//
// Upstream is explicit: the update path is `Object.assign(existing, { kind, applyFn, compounds })`
// and `order` is not among the assigned keys. A port that re-derived the order on every
// registration would push every overridden framework variant to the end of the sort, which is a
// reordering of the whole class list for a stylesheet that only meant to restyle one variant.
func TestReRegisteringAVariantKeepsItsOrder(t *testing.T) {
	t.Parallel()
	registry := NewVariantRegistry()
	registry.Register("first", ParsedVariantKindStatic)
	registry.Register("second", ParsedVariantKindStatic)
	registry.Register("third", ParsedVariantKindStatic)

	first, _ := registry.Get("first")
	originalOrder := first.Order

	// The `@custom-variant first (...)` case: same name, re-registered.
	registry.Register("first", ParsedVariantKindStatic)

	updated, isRegistered := registry.Get("first")
	if !isRegistered {
		t.Fatal("re-registering should not remove the variant")
	}
	if updated.Order != originalOrder {
		t.Errorf(
			"re-registering moved `first` from order %d to %d, so overriding a framework variant "+
				"would reorder every class that uses it",
			originalOrder, updated.Order,
		)
	}

	third, _ := registry.Get("third")
	if updated.Order >= third.Order {
		t.Errorf(
			"`first` should still precede `third`, but holds order %d against %d",
			updated.Order, third.Order,
		)
	}
	if len(registry.Registrations()) != 3 {
		t.Errorf("re-registering should not add a registration, got %d", len(registry.Registrations()))
	}

	// The kind is updated even though the order is not, which is the other half of the assignment.
	registry.Register("second", ParsedVariantKindFunctional)
	second, _ := registry.Get("second")
	if second.Kind != ParsedVariantKindFunctional {
		t.Errorf("re-registering should update the kind, got %q", second.Kind)
	}
}

// TestSharedRegistrationOrderStillSeparatesByRoot is the known-dirty control for the tie above.
//
// Without it, a port that collapsed every shared-order group into one index would pass
// TestTiedVariantsShareAnIndex and be wrong about all five breakpoints. The engine gives `sm`
// through `2xl` five distinct indices despite one shared order, because `compare` continues to the
// root name when no comparison function is registered.
func TestSharedRegistrationOrderStillSeparatesByRoot(t *testing.T) {
	t.Parallel()
	registry := NewVariantRegistry()
	registry.Group(func() {
		registry.Register("alpha", ParsedVariantKindStatic)
		registry.Register("beta", ParsedVariantKindStatic)
	}, nil)

	order := registry.BuildVariantOrder([]ParsedVariant{
		{Kind: ParsedVariantKindStatic, Root: "alpha"},
		{Kind: ParsedVariantKindStatic, Root: "beta"},
	})

	alphaIndex, _ := order.IndexOf(ParsedVariant{Kind: ParsedVariantKindStatic, Root: "alpha"})
	betaIndex, _ := order.IndexOf(ParsedVariant{Kind: ParsedVariantKindStatic, Root: "beta"})
	if alphaIndex == betaIndex {
		t.Errorf(
			"a shared order with no comparison function should still separate by root, but both took "+
				"index %d, which would collapse the five breakpoints onto one bit",
			alphaIndex,
		)
	}
}

// TestCompareBreakpointsBucketsByUnit pins the two behaviours of compareBreakpoints a reader would
// not predict.
//
// Bucketing strips digits and dots, so values in different units never compare numerically at all:
// `40rem` and `1024px` bucket as `rem` and `px` and sort alphabetically. And a value that parses to
// no number falls back to a string comparison, which is upstream's `Number.isNaN` branch.
func TestCompareBreakpointsBucketsByUnit(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name      string
		left      string
		right     string
		ascending bool
		wantFirst string
		why       string
	}{
		{
			name: "same unit ascending", left: "40rem", right: "64rem", ascending: true,
			wantFirst: "40rem", why: "same bucket, so the numbers decide",
		},
		{
			name: "same unit descending", left: "40rem", right: "64rem", ascending: false,
			wantFirst: "64rem", why: "`max-*` descends",
		},
		{
			name: "different units sort by unit name", left: "1024px", right: "40rem", ascending: true,
			wantFirst: "1024px", why: "`px` < `rem` alphabetically, and the numbers are never compared",
		},
		{
			name: "unit name beats magnitude", left: "9999px", right: "1rem", ascending: true,
			wantFirst: "9999px", why: "bucketing happens before any numeric comparison",
		},
		{
			name: "css functions fall back to text", left: "calc(100%-1rem)", right: "calc(100%-2rem)", ascending: true,
			wantFirst: "calc(100%-1rem)", why: "neither parses to a number, so upstream compares the strings",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			order := CompareBreakpoints(testCase.left, testCase.right, testCase.ascending)
			if order == 0 {
				t.Fatalf("%q and %q tied, which the engine does not do for distinct values", testCase.left, testCase.right)
			}
			first := testCase.left
			if order > 0 {
				first = testCase.right
			}
			if first != testCase.wantFirst {
				t.Errorf("expected %q first because %s, got %q", testCase.wantFirst, testCase.why, first)
			}
			// Antisymmetry, because a comparator that is not antisymmetric makes sort.Slice's
			// behaviour undefined rather than merely wrong.
			if reversed := CompareBreakpoints(testCase.right, testCase.left, testCase.ascending); reversed != -order {
				t.Errorf("the comparison is not antisymmetric: %d one way and %d the other", order, reversed)
			}
		})
	}
}

// TestBitmaskExceedsSixtyFourBits is the known-dirty control for the mask's width.
//
// A `uint64` mask is the obvious choice and it is wrong. The dense index counts distinct variants
// in one class list's population, and a large repository reaches well past 64. A `uint64` port
// would shift past the word width, which in Go silently yields zero rather than trapping, so the
// class would carry no bit at all and sort as though it had no variants.
func TestBitmaskExceedsSixtyFourBits(t *testing.T) {
	t.Parallel()
	registry := NewVariantRegistry()
	var variants []ParsedVariant
	for index := 0; index < 100; index++ {
		name := fmt.Sprintf("variant%03d", index)
		registry.Register(name, ParsedVariantKindStatic)
		variants = append(variants, ParsedVariant{Kind: ParsedVariantKindStatic, Root: name})
	}

	order := registry.BuildVariantOrder(variants)
	if order.Len() != 100 {
		t.Fatalf("expected 100 distinct variants, got %d", order.Len())
	}

	highest := variants[len(variants)-1]
	mask, isComplete := order.VariantBitmask([]ParsedVariant{highest})
	if !isComplete {
		t.Fatal("the highest variant should have an index")
	}
	if mask.BitLen() <= 64 {
		t.Errorf("the mask should exceed 64 bits, got %d bits for %s", mask.BitLen(), mask)
	}
	if mask.Sign() == 0 {
		t.Error("the mask is zero, which is what a uint64 shift past the word width would produce")
	}
}

// TestCorpusHoldsRealClassLists guards the population every count above is a count of.
//
// A fixture that regenerated to zero repository class lists would leave every differential passing
// vacuously and every logged percentage a percentage of nothing.
func TestCorpusHoldsRealClassLists(t *testing.T) {
	t.Parallel()
	corpus := loadVariantOrderCorpus(t)

	classListsByRepository := map[string]int{}
	classesByRepository := map[string]int{}
	for _, aCase := range corpus.Cases {
		if aCase.Source != "repository" || len(aCase.ClassLists) == 0 {
			continue
		}
		repository := strings.TrimSuffix(aCase.Name, "/class-lists")
		for _, classList := range aCase.ClassLists {
			classListsByRepository[repository]++
			classesByRepository[repository] += len(classList.Classes)
		}
	}

	if len(classListsByRepository) < 2 {
		t.Fatalf(
			"the corpus holds class lists from %d repositories and the argument for this port needs two",
			len(classListsByRepository),
		)
	}
	for repository, count := range classListsByRepository {
		if count < 100 {
			t.Errorf("%s contributed only %d class lists, which is too few to measure agreement on", repository, count)
		}
		t.Logf("%s: %d class lists, %d classes", repository, count, classesByRepository[repository])
	}
}
