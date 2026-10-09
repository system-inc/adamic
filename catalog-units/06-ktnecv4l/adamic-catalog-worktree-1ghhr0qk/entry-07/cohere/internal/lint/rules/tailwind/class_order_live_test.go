package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

// The acceptance suite for the live class-order sort, measured against the engine's own answers.
//
// Every helper here is prefixed `classOrderLive` for the reason design_system_test.go states: several
// agents work in this package at once, and a helper whose name does not say which component owns it
// is a collision waiting to happen.
//
// # What this measures, and the trap it is built to avoid
//
// The regression check on this task was "15 disagreeing pairs of 4,265 must go to zero." That number
// is satisfiable two ways: by the rule becoming correct, or by the rule declining to have an opinion,
// and a differential scores both as agreement because both sides go silent. A previous agent refused
// to build against this task for exactly that reason, and was right to.
//
// So this suite reports a denominator alongside every rate. `classOrderLiveMeasure` counts the classes
// the rule actually PLACED, and the tests below assert that count against a floor as well as asserting
// the disagreement count against a ceiling. A run that placed nothing scores zero disagreements over
// zero pairs and fails on the floor rather than passing on the empty diff.
//
// # The ground truth
//
// `internal/lint/rules/tailwind/collapse/testdata/variant_fixtures.json` holds 2,400 class lists captured from
// `designSystem.getClassOrder(classes)` on the two corpus repositories, plus 37 synthetic cases built
// to separate rules that agree on ordinary input. `sorted` is what the engine returned. Nothing here
// asserts against what the Go code produces; the fixture is the authority and the Go side is the
// thing under test.

// classOrderLiveCorpus is the slice of variant_fixtures.json this suite reads.
type classOrderLiveCorpus struct {
	TailwindVersion string `json:"tailwindVersion"`
	Cases           []struct {
		Name       string `json:"name"`
		Source     string `json:"source"`
		EntryPath  string `json:"entryPath"`
		ClassOrder *struct {
			Sorted []string `json:"sorted"`
		} `json:"classOrder"`
		ClassLists []struct {
			Name       string `json:"name"`
			ClassOrder struct {
				Sorted []string `json:"sorted"`
			} `json:"classOrder"`
		} `json:"classLists"`
	} `json:"cases"`
}

// classOrderLiveList is one engine-sorted class list, with the design system it was captured against.
type classOrderLiveList struct {
	name      string
	entryPath string
	sorted    []string
}

// classOrderLiveLoadCorpus reads the fixture and flattens it to the lists this suite sorts.
//
// Both shapes are collected: the synthetic `classOrder` on a case, and the `classLists` array a
// repository case carries. They answer different questions and dropping either would leave a gap the
// count could not show, since the synthetic cases are the only ones built to separate the depth-first
// rule from the mask and the repository lists are the only ones that prove it on code people wrote.
func classOrderLiveLoadCorpus(t *testing.T) []classOrderLiveList {
	t.Helper()

	path := filepath.Join("collapse", "testdata", "variant_fixtures.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var fixture classOrderLiveCorpus
	if err := json.Unmarshal(contents, &fixture); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if fixture.TailwindVersion != "4.3.3" {
		t.Fatalf("the fixture was captured from Tailwind %s and this port targets 4.3.3",
			fixture.TailwindVersion)
	}

	var lists []classOrderLiveList
	for _, aCase := range fixture.Cases {
		if aCase.ClassOrder != nil && len(aCase.ClassOrder.Sorted) > 1 {
			lists = append(lists, classOrderLiveList{
				name:      aCase.Name,
				entryPath: aCase.EntryPath,
				sorted:    aCase.ClassOrder.Sorted,
			})
		}
		for _, classList := range aCase.ClassLists {
			if len(classList.ClassOrder.Sorted) < 2 {
				continue
			}
			lists = append(lists, classOrderLiveList{
				name:      classList.Name,
				entryPath: aCase.EntryPath,
				sorted:    classList.ClassOrder.Sorted,
			})
		}
	}

	if len(lists) == 0 {
		t.Fatal("the fixture yielded no multi-class lists, so every assertion below would pass vacuously")
	}
	return lists
}

// classOrderLiveSystems builds one live design system per entry point the corpus names.
//
// Built from the repository on disk rather than from the fixture, which is the whole point of the
// swap: the thing under test is the rule reading the repository in front of it. The fixture spells each
// entry point inside its corpus, so an unset corpus skips the test naming its variable, and a corpus
// that is set but lacks the stylesheet fails.
func classOrderLiveSystems(t *testing.T, lists []classOrderLiveList) map[string]DesignSystemResult {
	t.Helper()

	systems := map[string]DesignSystemResult{}
	for _, list := range lists {
		if list.entryPath == "" {
			continue
		}
		if _, built := systems[list.entryPath]; built {
			continue
		}
		entryPoint := corpus.Resolve(t, list.entryPath)
		packageRoot := findTailwindPackageRoot(filepath.Dir(entryPoint), diskFileExists)
		if packageRoot == "" {
			systems[list.entryPath] = DesignSystemResult{
				Err: fmt.Errorf("no installed tailwindcss beside %s", entryPoint),
			}
			continue
		}
		system, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
			EntryPoint:          entryPoint,
			TailwindPackageRoot: packageRoot,
		})
		if err != nil {
			systems[list.entryPath] = DesignSystemResult{Err: err}
			continue
		}
		systems[list.entryPath] = DesignSystemResult{
			System: system,
			Table:  tailwindengine.NewTable(system),
		}
	}
	return systems
}

// classOrderLiveMeasurement is what one pass over the corpus found.
//
// Placed and skipped are carried beside the disagreement counts for the reason in the file comment: a
// disagreement count means nothing without the population it was measured over, and the failure this
// task is guarded against is a zero produced by an empty population.
type classOrderLiveMeasurement struct {
	// lists is how many class lists the rule produced an order for.
	lists int
	// listsSkipped is how many it declined, because a class in them could not be placed.
	listsSkipped int
	// placedClasses is every class the rule assigned a key to. The denominator that matters.
	placedClasses int
	// comparedPairs is every ordered pair within a sorted list, which is what a disagreement is
	// counted against.
	comparedPairs int
	// disagreeingPairs is the pairs the rule ordered the other way from the engine.
	disagreeingPairs int
	// disagreeingLists is how many lists held at least one such pair.
	disagreeingLists int
	// examples holds up to eight readable disagreements.
	examples []string
}

// classOrderLiveMeasure sorts every corpus list with the live rule and scores it against the engine.
//
// The list is shuffled into reverse before sorting rather than being handed to the rule already
// correct. A sort fed its own output is a sort that only has to be stable to pass, which would let a
// comparator that expresses no opinion at all score a hundred percent.
func classOrderLiveMeasure(
	t *testing.T,
	lists []classOrderLiveList,
	systems map[string]DesignSystemResult,
) classOrderLiveMeasurement {
	t.Helper()

	measurement := classOrderLiveMeasurement{}
	for _, list := range lists {
		designSystem, hasSystem := systems[list.entryPath]
		if !hasSystem || designSystem.Err != nil {
			continue
		}

		unranked, placeable := partitionUnranked(classOrderLiveReversed(list.sorted), designSystem)
		keys, _, resolved := classOrderKeys(placeable, designSystem.System, designSystem.Table)
		if !resolved {
			measurement.listsSkipped++
			continue
		}

		measurement.lists++
		measurement.placedClasses += len(placeable)

		ordered := append(unranked, sortClassesByKey(placeable, keys)...)
		position := make(map[string]int, len(ordered))
		for index, className := range ordered {
			position[className] = index
		}
		// Two nulls keep source order, and the source here is the reversal, so the engine's order
		// between them is no answer at all. Every other pair, a null against a ranked class
		// included, is scored.
		isUnranked := make(map[string]bool, len(unranked))
		for _, className := range unranked {
			isUnranked[className] = true
		}

		// Scored pairwise against the engine's own sequence rather than by string equality, so a
		// single misplaced class is one disagreement rather than a whole list, and the number stays
		// comparable to the 15-of-4,265 baseline, which was also counted in pairs.
		listDisagrees := false
		for left := 0; left < len(list.sorted); left++ {
			for right := left + 1; right < len(list.sorted); right++ {
				leftClass, rightClass := list.sorted[left], list.sorted[right]
				leftPosition, leftPlaced := position[leftClass]
				rightPosition, rightPlaced := position[rightClass]
				if !leftPlaced || !rightPlaced || (isUnranked[leftClass] && isUnranked[rightClass]) {
					continue
				}
				measurement.comparedPairs++
				// The engine put leftClass first, by construction of `sorted`.
				if leftPosition > rightPosition {
					measurement.disagreeingPairs++
					listDisagrees = true
					if len(measurement.examples) < 8 {
						measurement.examples = append(measurement.examples, fmt.Sprintf(
							"%s: engine puts %q before %q; the rule reverses them",
							list.name, leftClass, rightClass,
						))
					}
				}
			}
		}
		if listDisagrees {
			measurement.disagreeingLists++
		}
	}
	return measurement
}

// classOrderLiveReversed returns a class list back to front.
//
// Reversing rather than shuffling, because a shuffle needs a seed to be reproducible and a reversal
// is the one permutation guaranteed to disagree with the target on every separable pair. A sort that
// recovers the engine's order from its exact opposite has been asked the question properly.
func classOrderLiveReversed(classes []string) []string {
	reversed := make([]string, len(classes))
	for index, className := range classes {
		reversed[len(classes)-1-index] = className
	}
	return reversed
}

// TestClassOrderLiveMatchesTheEngineOverTheCorpus is the acceptance test for the swap.
//
// The shipped depth-first rule disagreed with the engine on 15 of 4,265 separable pairs across 7
// class lists, measured by `TestDepthFirstDisagreesWithTheEngineOnTheCorpus` in `internal/tailwind`.
// After the swap that must be zero, and it must be zero over a population that did not shrink, which
// is what the placed-class floor below is for.
//
// The floor is a floor rather than an equality so that a fixture gaining or losing a handful of classes
// does not fail it. Both corpora, ahra and connected, must be set for this to run, and it skips naming
// the variable when either is not.
func TestClassOrderLiveMatchesTheEngineOverTheCorpus(t *testing.T) {
	t.Parallel()
	lists := classOrderLiveLoadCorpus(t)
	systems := classOrderLiveSystems(t, lists)
	measurement := classOrderLiveMeasure(t, lists, systems)

	t.Logf(
		"live rule vs the engine: %d disagreeing pairs of %d compared pairs across %d lists "+
			"(%d classes placed, %d lists declined)",
		measurement.disagreeingPairs, measurement.comparedPairs, measurement.lists,
		measurement.placedClasses, measurement.listsSkipped,
	)
	for _, example := range measurement.examples {
		t.Log("  " + example)
	}

	// The population guard, and it comes first on purpose. Every assertion below is satisfiable by a
	// rule that placed nothing, so the denominator is checked before the rate.
	//
	// # The floor is tight because a loose one let a real mutation through
	//
	// This was written as `< 5000` first, on the reasoning that the corpus holds ~11,300 classes and
	// half of that was clearly "still measuring something". Then the mutation that removes the
	// evaluator escalation in `readingFor` was run against it: every repository `@utility` root goes
	// unanswerable, 724 lists drop out, the population falls from 11,307 to 7,103, and the
	// disagreement count stays at zero because the classes that would have disagreed are simply not
	// being asked about any more. The suite passed.
	//
	// That is precisely the failure this task exists to prevent, arriving inside the test written to
	// prevent it: a number that improves because a rule went quiet. A floor only works if it is set
	// near the real population rather than at a number that merely sounds large.
	//
	// 11,000 of the 11,307 currently placed. Tight enough that losing one class category fails, loose
	// enough that a fixture gaining or losing a handful of classes does not. `listsDeclined` is
	// asserted separately for the same reason: a decline is how classes leave the population, so its
	// count is the leading indicator and it is currently 1.
	if measurement.comparedPairs == 0 {
		t.Fatal("no pairs were compared, so this measured nothing")
	}
	if measurement.placedClasses < 11000 {
		t.Fatalf(
			"the rule placed %d classes and the corpus holds 11,307; a population that dropped this "+
				"far means the rule stopped answering for a whole category of class, and the zero "+
				"disagreements below would be measuring its silence rather than its correctness",
			measurement.placedClasses,
		)
	}
	if measurement.listsSkipped > 10 {
		t.Errorf(
			"the rule declined %d of %d lists; it declined 1 when this was measured, and a rise means "+
				"classes are leaving the population rather than being ordered correctly",
			measurement.listsSkipped, measurement.lists+measurement.listsSkipped,
		)
	}

	if measurement.disagreeingPairs != 0 {
		t.Errorf(
			"the live rule disagrees with the engine on %d of %d pairs across %d lists; the swap was "+
				"supposed to take this to zero",
			measurement.disagreeingPairs, measurement.comparedPairs, measurement.disagreeingLists,
		)
	}
}

// TestClassOrderLiveStackedVariantsSortByMaskNotByDepth is the known-dirty control.
//
// This is the bug the swap exists to fix, stated as the smallest input that shows it. The shipped
// rule asserted "every single variant precedes every stacked one, whatever they start with", and the
// engine implements nothing of the kind: it ORs one bit per variant into a mask and compares masks
// numerically, so a class's position is decided by its highest-ranked variant.
//
// `group-hover:disabled:flex` carries two variants and mask 5. `dark:flex` carries one and mask 8.
// The engine sorts the stacked class FIRST, and the depth-first rule reversed it. Both fixtures are
// the engine's own answers, quoted from variant_fixtures.json, not this port's output.
//
// The pairs are asserted rather than merely counted so that a corpus that lost its stacked-variant
// cases could not turn this suite green by having nothing left to disagree about.
func TestClassOrderLiveStackedVariantsSortByMaskNotByDepth(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		input    []string
		expected []string
		why      string
	}{
		{
			name:     "a stacked variant sorts before a higher-ranked single one",
			input:    []string{"dark:flex", "group-hover:disabled:flex"},
			expected: []string{"group-hover:disabled:flex", "dark:flex"},
			why: "masks 5 and 8: the stacked class carries only lower bits, so it sorts first " +
				"despite having more variants",
		},
		{
			name:     "and it lands between the singles rather than after them",
			input:    []string{"dark:flex", "group-hover:disabled:flex", "hover:flex", "group-hover:flex"},
			expected: []string{"group-hover:flex", "hover:flex", "group-hover:disabled:flex", "dark:flex"},
			why:      "depth-first would put the two-variant class last; the engine puts it third",
		},
		{
			name:     "an arbitrary variant sorts after a stacked registered one",
			input:    []string{"[&:hover]:flex", "dark:hover:flex", "hover:flex"},
			expected: []string{"hover:flex", "dark:hover:flex", "[&:hover]:flex"},
			why:      "arbitrary variants rank after every registered one, so mask 4 beats mask 3",
		},
		{
			name:     "stacking order does not change the mask",
			input:    []string{"hover:dark:flex", "dark:hover:flex"},
			expected: []string{"dark:hover:flex", "hover:dark:flex"},
			why:      "both set the same two bits, so they tie on the mask and separate on the name",
		},
	}

	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {
		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()
				unranked, placeable := partitionUnranked(testCase.input, designSystem)
				keys, unplaceable, resolved := classOrderKeys(placeable, designSystem.System, designSystem.Table)
				if !resolved {
					t.Fatalf("the rule could not place %q, so this case proves nothing", unplaceable)
				}
				ordered := append(unranked, sortClassesByKey(placeable, keys)...)
				if strings.Join(ordered, " ") != strings.Join(testCase.expected, " ") {
					t.Errorf("got %v, engine says %v: %s", ordered, testCase.expected, testCase.why)
				}
			})
		}
	})
}

// TestClassOrderLiveReadsTheRepositoryRatherThanATable is the control that the swap actually happened.
//
// Every other test here would pass against the generated tables, because the generated tables were
// generated from this repository and agree with it almost everywhere. What cannot pass against them
// is a class whose reading exists only because this repository declared it: `markdown-content` is one
// of ahra's own `@utility` blocks, present in the live design system and absent from Tailwind.
//
// So this asserts the rule places a class no framework table contains. Without it, a swap that
// silently kept reading the old tables would show green on the whole suite above.
func TestClassOrderLiveReadsTheRepositoryRatherThanATable(t *testing.T) {
	t.Parallel()
	designSystem := classOrderLiveRepositorySystem(t)

	// `markdown-content` is declared by this repository's own `app/_theme/styles/utilities.css` and
	// exists nowhere in Tailwind. Asserted through HasUtility first, so a repository that renames the
	// block fails here with a readable reason rather than further down with a confusing one.
	const repositoryClass = "markdown-content"
	if !designSystem.System.HasUtility(repositoryClass, tailwindengine.UtilityKindStatic) {
		t.Skipf("this repository no longer declares `@utility %s`, so there is nothing only it knows",
			repositoryClass)
	}

	keys, unplaceable, resolved := classOrderKeys(
		[]string{repositoryClass, "flex"}, designSystem.System, designSystem.Table)
	if !resolved {
		t.Fatalf("the rule could not place %q, which this repository's own stylesheet declares; "+
			"that means it is not reading the repository", unplaceable)
	}
	if _, placed := keys[repositoryClass]; !placed {
		t.Fatalf("%q got no key, so the rule is answering from a framework table rather than from "+
			"this repository", repositoryClass)
	}
}

// classOrderLiveRepositorySystem builds the design system of ahra, the corpus these tests were written
// against.
//
// Skips naming the variable when the corpus is unset, so a checkout elsewhere reports "not measured"
// rather than "broken", and fails when it is set but lacks the stylesheet or its installed tailwindcss.
func classOrderLiveRepositorySystem(t *testing.T) DesignSystemResult {
	t.Helper()

	entryPoint := corpus.Resolve(t, "ahra:app/_theme/styles/theme.css")
	packageRoot := findTailwindPackageRoot(filepath.Dir(entryPoint), diskFileExists)
	if packageRoot == "" {
		t.Fatalf("%s is set, but there is no installed tailwindcss beside %s", corpus.Ahra.Variable, entryPoint)
	}
	system, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
		EntryPoint:          entryPoint,
		TailwindPackageRoot: packageRoot,
	})
	if err != nil {
		t.Fatalf("loading the design system: %v", err)
	}
	return DesignSystemResult{System: system, Table: tailwindengine.NewTable(system), EntryPoint: entryPoint}
}

// TestClassOrderLiveDeclinesRatherThanPartiallySorting pins the fail-safe.
//
// A class the design system cannot place must stop the whole literal, not be sorted around. Sorting
// the classes that resolved would reorder correct code on the strength of a lookup that failed, and
// the author would be asked to make a change the engine does not agree with.
func TestClassOrderLiveDeclinesRatherThanPartiallySorting(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {

		_, unplaceable, resolved := classOrderKeys(
			[]string{"items-center", "flex", "ahralia-splash"}, designSystem.System, designSystem.Table)
		if resolved {
			t.Fatal("a class outside the design system was placed, so the rule would sort a literal it " +
				"does not understand")
		}
		if unplaceable != "ahralia-splash" {
			t.Errorf("the decline should name the class that caused it, got %q", unplaceable)
		}
	})
}

// TestClassOrderLiveVariantIndicesAreRanksWithinTheList is the mechanism test.
//
// It is the reason this port sorts a list rather than comparing pairs, and it is stated as an
// assertion because a comment saying "the index is population-scoped" decays and a measurement does
// not. The same variant takes a different index depending on what else the list contains, so a key
// computed for one list is meaningless in another.
func TestClassOrderLiveVariantIndicesAreRanksWithinTheList(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {

		maskOf := func(classes []string, target string) string {
			keys, unplaceable, resolved := classOrderKeys(classes, designSystem.System, designSystem.Table)
			if !resolved {
				t.Fatalf("could not place %q", unplaceable)
			}
			key, placed := keys[target]
			if !placed {
				t.Fatalf("%q got no key", target)
			}
			if key.mask == nil {
				return "0"
			}
			return key.mask.String()
		}

		narrow := maskOf([]string{"dark:flex", "hover:flex"}, "dark:flex")
		wide := maskOf([]string{"dark:flex", "hover:flex", "group-hover:disabled:flex"}, "dark:flex")

		if narrow == wide {
			t.Errorf(
				"`dark:flex` holds mask %s in both a two-class and a four-variant list; the index is "+
					"supposed to be a rank within the population, so a pairwise comparator would be a "+
					"valid shape after all and this port's central claim is wrong",
				narrow,
			)
		}
		t.Logf("`dark:flex` masks %s in the narrow list and %s in the wide one", narrow, wide)
	})
}

// TestClassOrderLiveUnrankedLeadInSourceOrder pins the one dimension the engine has no opinion on.
//
// `getClassOrder` returns null for `group` and `peer` and for every class that compiles to nothing,
// so the corpus agreement above says nothing about where they belong. The convention is Tailwind's
// own Prettier plugin's: every null leads, and two nulls keep the order they were written in.
//
// The non-marker nulls are the case #vf1hd6j found missing. `text-dark` and `dark:bg-dark-2` are
// classes the plugin ranked null on the committed trees, and `ahralia-splash` does not parse at all.
// `ahra` declares no `dark` colour, which this asserts rather than assumes: the test is about nulls
// and must not quietly turn into a test about two ranked classes.
func TestClassOrderLiveUnrankedLeadInSourceOrder(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {

		for _, nullClass := range []string{"peer", "group", "text-dark", "dark:bg-dark-2", "ahralia-splash"} {
			if !isMarkerClass(nullClass) && classCompilesIn(nullClass, designSystem) {
				t.Fatalf("%q compiles in this repository, so it is not a null and this test proves nothing", nullClass)
			}
		}

		for _, testCase := range []struct {
			input, unranked []string
		}{
			{[]string{"flex", "peer", "group", "items-center"}, []string{"peer", "group"}},
			{[]string{"flex", "group", "peer", "items-center"}, []string{"group", "peer"}},
			{[]string{"flex", "text-dark", "peer", "items-center"}, []string{"text-dark", "peer"}},
			{[]string{"items-center", "ahralia-splash", "dark:bg-dark-2", "flex"}, []string{"ahralia-splash", "dark:bg-dark-2"}},
		} {
			unranked, placeable := partitionUnranked(testCase.input, designSystem)
			// Source order between them, which is what makes `peer group` and `group peer` both legal.
			if strings.Join(unranked, " ") != strings.Join(testCase.unranked, " ") {
				t.Errorf("from %v the nulls are %v; they should be %v, in source order", testCase.input, unranked, testCase.unranked)
			}
			if len(placeable) != 2 {
				t.Errorf("expected two placeable classes from %v, got %v", testCase.input, placeable)
			}
		}
	})
}

// TestClassOrderLiveDeclinesOnlyTheKnownBoundary names the one list the rule cannot order.
//
// `TestClassOrderLiveMatchesTheEngineOverTheCorpus` reports "1 lists declined" and a decline is how
// classes leave the population, so an unexplained one is exactly the thing that would let a future
// silence hide. This says which list, which class, and why, so the count is a fact with a reason
// attached rather than a number nobody has looked at.
//
// The list is `ahra/list-179` and the class is `data-[show=false]:fade-out`. Its root is declared
// twice in this repository's own stylesheets:
//
//	@utility fade-out    { --exit-opacity: 0; }
//	@utility fade-out-*  { --exit-opacity: calc(--value(number) / 100); … }
//
// `LoadedDesignSystem.utilityRoots` maps a root to ONE kind, and designsystem.go's own comment says
// so and says why: it exists to answer the parser's question, and the parser asks about one kind at a
// time. Sixteen roots in this repository have that shape. So the functional registration wins the
// map, `ParseCandidate` reads the bare `fade-out` as a functional candidate with no value, the
// descriptor declines it as per-declaration, and the evaluator declines it because a functional block
// with no value does not compile.
//
// That is a boundary in a component this task does not own, and the right behaviour on reaching it is
// the one the rule takes: decline the whole literal rather than sort around the class it cannot
// place. It is pinned here so that a future change which fixes it, or which widens it, shows up as a
// failure in a test that names the cause instead of as a silent shift in a count.
func TestClassOrderLiveDeclinesOnlyTheKnownBoundary(t *testing.T) {
	t.Parallel()
	lists := classOrderLiveLoadCorpus(t)
	systems := classOrderLiveSystems(t, lists)

	declined := map[string]string{}
	for _, list := range lists {
		designSystem, hasSystem := systems[list.entryPath]
		if !hasSystem || designSystem.Err != nil {
			continue
		}
		_, placeable := partitionUnranked(classOrderLiveReversed(list.sorted), designSystem)
		if _, unplaceable, resolved := classOrderKeys(
			placeable, designSystem.System, designSystem.Table); !resolved {
			declined[list.name] = unplaceable
		}
	}

	for name, unplaceable := range declined {
		t.Logf("declined %s on %q", name, unplaceable)
	}

	// Every decline must be a root this repository declares both statically and functionally, which
	// is the boundary above. A decline on anything else is a class the rule stopped answering for,
	// and that is the failure this whole suite is built to refuse.
	for name, unplaceable := range declined {
		_, base := splitVariants(unplaceable)
		root := strings.SplitN(base, "-", 2)[0]
		if root != "fade" && root != "zoom" && root != "slide" && root != "spin" {
			t.Errorf(
				"%s declined on %q, which is not the known static/functional collision; the rule has "+
					"stopped answering for a class it used to place",
				name, unplaceable)
		}
	}
	if len(declined) > 3 {
		t.Errorf("%d lists declined; 1 did when this was measured, and a rise means classes are "+
			"leaving the population", len(declined))
	}
}

// A class ranks by the first of its readings that compiles, and a class none of whose readings
// compile is a null. Both answers quoted from Prettier's Tailwind plugin over this repository's
// theme on 2026-10-02, where literals holding either shape were declined until #vf1hd6j's
// acceptance found them: 12 in ahra and 18 in www-phi-health.
//
// `shadow--0` reads first as the repository's `@utility shadow--*` with value `0`, needing a
// `--shadow-0` nobody declared, and second as the framework's `shadow` with value `-0`, which finds
// `--shadow--0`. `hover:content--0-4` reads as `@utility content--*` with value `0-4`, needing a
// `--color-content-0-4` the theme does not have, and as nothing else.
func TestClassOrderRanksByTheReadingThatCompiles(t *testing.T) {
	t.Parallel()
	designSystem := classOrderLiveRepositorySystem(t)

	for _, testCase := range []struct {
		input, want []string
	}{
		{[]string{"shadow--0", "background--0", "border--0", "border"}, []string{"border", "border--0", "background--0", "shadow--0"}},
		{[]string{"shadow--6", "hover:shadow--3", "p-2", "rounded-md"}, []string{"rounded-md", "p-2", "shadow--6", "hover:shadow--3"}},
		{[]string{"flex", "dark:bg-transparent", "hover:content--0-4"}, []string{"hover:content--0-4", "flex", "dark:bg-transparent"}},
	} {
		ordered, decided := orderClasses(testCase.input, designSystem, defaultClassOrderOptions())
		if !decided {
			t.Errorf("%v was declined; every class in it has an engine answer", testCase.input)
			continue
		}
		if strings.Join(ordered, " ") != strings.Join(testCase.want, " ") {
			t.Errorf("%v ordered as %v, the plugin writes %v", testCase.input, ordered, testCase.want)
		}
	}
}

// Existence asks the evaluator for a repository root rather than trusting it, and takes any reading
// that resolves rather than only the first.
//
// The dead ones are classes the plugin ranks null over this theme: `content--0-4` has no
// `--color-content-0-4`, and `background--2/50` puts a modifier on a utility that takes none. Both
// sat on Structure sites with no-unknown-classes reporting nothing. The live ones are the controls,
// and `shadow--0` is the one only a second reading reaches.
func TestClassExistenceAsksTheEvaluatorForRepositoryRoots(t *testing.T) {
	t.Parallel()
	designSystem := classOrderLiveRepositorySystem(t)

	for _, dead := range []string{"content--0-4", "hover:content--0-4", "background--2/50"} {
		if classExistsIn(dead, designSystem.System) {
			t.Errorf("%q generates no CSS and was read as existing", dead)
		}
	}
	for _, live := range []string{"content--1", "content--4", "background--2", "shadow--0", "hover:shadow--3", "markdown-content"} {
		if !classExistsIn(live, designSystem.System) {
			t.Errorf("%q generates CSS and was read as unknown", live)
		}
	}
}
