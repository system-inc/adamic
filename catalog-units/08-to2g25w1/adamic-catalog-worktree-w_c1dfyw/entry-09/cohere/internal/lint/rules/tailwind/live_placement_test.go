package tailwind

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

// The corpus placement measurement for the rules moved onto the live design system.
//
// # Why a placement count and not only a findings count
//
// This is the acceptance criterion #3r6cxrb was written around, and it exists because of the way the
// previous attempt failed. A migration is scored by comparing the rule's answers before and after,
// and a rule that stops having answers scores perfectly on every comparison that only counts
// disagreements: silence reads as agreement. `#4q5dsn3` was blocked over exactly that, on a
// regression check that named a number without naming the population it had to be measured over.
//
// So every measurement here reports how many classes the rule actually PLACED — resolved to a root,
// or to a property set, or to an existence answer — beside whatever else it reports. A rule that went
// quiet fails on the denominator rather than passing on the rate.
//
// # The corpus
//
// `internal/lint/rules/tailwind/collapse/testdata/variant_fixtures.json`, the same 2,396 class lists and 11,314 class
// occurrences `class_order_live_test.go` measures against, captured from two real repositories. Real
// literals rather than constructed ones, because the failure these rules have historically had is on
// shapes nobody would think to construct.
//
// Every helper here is prefixed `livePlacement` for the reason design_system_test.go states: several
// agents work in this package at once, and a helper whose name does not say which component owns it
// is a collision waiting to happen.

// livePlacementCorpus is the slice of variant_fixtures.json this file reads.
type livePlacementCorpus struct {
	Cases []struct {
		EntryPath  string `json:"entryPath"`
		ClassOrder *struct {
			Sorted []string `json:"sorted"`
		} `json:"classOrder"`
		ClassLists []struct {
			ClassOrder struct {
				Sorted []string `json:"sorted"`
			} `json:"classOrder"`
		} `json:"classLists"`
	} `json:"cases"`
}

// livePlacementLiterals returns the corpus's class lists, grouped by the design system they came
// from.
//
// Grouped rather than flattened because a class's answer is a function of the repository it was
// written in, which is the whole premise of the migration. Measuring www-connected-app's classes
// against ahra's design system would be measuring the defect rather than the fix.
func livePlacementLiterals(t *testing.T) map[string][][]string {
	t.Helper()

	path := filepath.Join("collapse", "testdata", "variant_fixtures.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var fixture livePlacementCorpus
	if err := json.Unmarshal(contents, &fixture); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	literals := map[string][][]string{}
	for _, aCase := range fixture.Cases {
		if aCase.EntryPath == "" {
			continue
		}
		if aCase.ClassOrder != nil && len(aCase.ClassOrder.Sorted) > 1 {
			literals[aCase.EntryPath] = append(literals[aCase.EntryPath], aCase.ClassOrder.Sorted)
		}
		for _, classList := range aCase.ClassLists {
			if len(classList.ClassOrder.Sorted) > 1 {
				literals[aCase.EntryPath] = append(literals[aCase.EntryPath], classList.ClassOrder.Sorted)
			}
		}
	}
	if len(literals) == 0 {
		t.Fatal("the corpus yielded no multi-class literals, so every measurement below would be empty")
	}
	return literals
}

// livePlacementSystem builds the design system one corpus entry point names.
//
// The fixture spells the entry point inside its corpus, so an unset corpus skips the test naming its
// variable, and a corpus that is set but lacks the stylesheet fails. A stylesheet that is there but
// will not load yields a result carrying its error, which the population floors then catch.
func livePlacementSystem(t *testing.T, spelling string) DesignSystemResult {
	t.Helper()

	entryPoint := corpus.Resolve(t, spelling)
	packageRoot := findTailwindPackageRoot(filepath.Dir(entryPoint), diskFileExists)
	if packageRoot == "" {
		return DesignSystemResult{Err: os.ErrNotExist}
	}
	system, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
		EntryPoint:          entryPoint,
		TailwindPackageRoot: packageRoot,
	})
	if err != nil {
		return DesignSystemResult{Err: err}
	}
	return DesignSystemResult{System: system, Table: tailwindengine.NewTable(system)}
}

// TestUnknownClassesPlacesEveryCorpusClass is `no-unknown-classes`'s population measurement.
//
// The rule's output is a finding per class it cannot place, so its population is every class it was
// asked about and its finding count is the complement. Both are asserted, because either alone is
// satisfiable by a broken rule: a rule that answers "known" for everything reports zero findings, and
// a rule that answers "unknown" for everything reports the whole corpus.
//
// Before the migration this measured 11,431 classes with zero reported, on tables generated from
// these same two repositories. That zero was not evidence the rule was right — it was evidence the
// tables described the corpus, which is the thing the migration removes. The number to watch is that
// it is still zero now that the answer comes from each repository's own stylesheet.
func TestUnknownClassesPlacesEveryCorpusClass(t *testing.T) {
	t.Parallel()
	literals := livePlacementLiterals(t)

	asked, unknown := 0, 0
	unknownExamples := map[string]int{}

	for entryPoint, lists := range literals {
		designSystem := livePlacementSystem(t, entryPoint)
		if designSystem.Err != nil {
			continue
		}
		for _, classes := range lists {
			for _, className := range classes {
				asked++
				if classExistsIn(className, designSystem.System) {
					continue
				}
				unknown++
				unknownExamples[className]++
			}
		}
	}

	t.Logf("no-unknown-classes: asked about %d classes, reported %d as unknown", asked, unknown)
	for _, example := range livePlacementTopKeys(unknownExamples, 20) {
		t.Logf("  %4d  %s", unknownExamples[example], example)
	}

	// The population guard, first, because every assertion below it is satisfiable by a rule that was
	// never asked anything.
	if asked < 11000 {
		t.Fatalf(
			"the rule was asked about %d classes and the corpus holds 11,314; a population that "+
				"dropped this far means the corpus stopped loading, and the finding count below "+
				"would be measuring an empty run",
			asked,
		)
	}

	// Both corpus repositories write only classes their own design systems define, which is what
	// makes zero the right answer here and what makes it worth asserting: a rule that reported on
	// real code would be reporting classes that work.
	if unknown != 0 {
		t.Errorf(
			"the rule reported %d of %d corpus classes as unknown. Every class here is one a real "+
				"repository writes against the design system it was measured with, so a finding is a "+
				"false positive rather than a defect found",
			unknown, asked,
		)
	}
}

// TestCanonicalClassesPlacementIsAccountedFor is `enforce-canonical-classes`'s measurement, and it
// is the one that had to explain a drop rather than only report one.
//
// Measured over the corpus, before and after the split moved onto the live design system:
//
//	placed    9,428 -> 7,287
//	findings      2 ->     2   (`size-[20px]` and `size-[55%]`, the same two)
//
// A placement drop of 2,141 with an unchanged finding count is exactly the shape that should be
// distrusted, so it is decomposed here rather than asserted away. Every dropped class is one whose
// FIRST engine reading is static, and a static has no root-plus-value structure for a collapse
// family to relate: `flex`, `block`, `border-dashed`, `cursor-pointer`. 135 distinct classes, 2,141
// occurrences, zero unparseable and zero unexplained.
//
// The shipped prefix walk placed them because it forced a functional reading that the engine does not
// take — it read `border-dashed` as root `border` with value `dashed`, when `parseCandidate` returns
// `border-dashed` as a static first. Those placements never became findings, because a static's
// bucket never matches a collapse family, so the rule's output is unchanged and 2,141 wasted lookups
// are gone.
//
// `overflow-x-hidden overflow-y-hidden => overflow-hidden` is a real engine collapse between two
// statics, and this rule reported it neither before nor after. It is a static collapse, which the
// rule's own doc comment places out of scope alongside the single-class rewrites.
func TestCanonicalClassesPlacementIsAccountedFor(t *testing.T) {
	t.Parallel()
	literals := livePlacementLiterals(t)

	placed, asked, findings := 0, 0, 0
	droppedStatics, droppedOther := 0, 0
	droppedExamples := map[string]string{}
	outputs := map[string]int{}

	for entryPoint, lists := range literals {
		designSystem := livePlacementSystem(t, entryPoint)
		if designSystem.Err != nil {
			continue
		}
		for _, classes := range lists {
			seen := make(map[string]bool, len(classes))
			remaining := make([]string, 0, len(classes))
			for _, className := range classes {
				asked++
				if seen[className] {
					continue
				}
				seen[className] = true
				remaining = append(remaining, className)

				if _, canSplit := splitCandidateIn(className, designSystem.System); canSplit {
					placed++
					continue
				}

				// Every class the live split declines must be one the engine reads as something other
				// than a functional utility, or the decline is a lost answer rather than a correct
				// one.
				parsed := tailwindengine.ParseCandidate(className, designSystem.System)
				switch {
				case len(parsed) == 0:
					droppedOther++
					droppedExamples[className] = "no reading at all"
				case parsed[0].Kind == tailwindengine.ParsedCandidateKindFunctional:
					// A functional class the split declined is a lost answer, since the split's only
					// other exit is a design system it could not read.
					droppedOther++
					droppedExamples[className] = "reads functionally but was not split"
				default:
					// Static or arbitrary. Neither has the root-plus-value structure a collapse
					// family relates, and `[font:inherit]` in particular is a whole declaration the
					// author wrote out, with no root at all.
					droppedStatics++
				}
			}

			// The rule's own loop, so this counts findings rather than merge opportunities.
			for pass := 0; pass < 6; pass++ {
				inputs, output, didMerge := mergeOnce(remaining, designSystem, canonicalCollapseOptions{logical: true})
				if !didMerge {
					break
				}
				findings++
				outputs[output]++
				next := make([]string, 0, len(remaining))
				merged := map[string]bool{inputs[0]: true, inputs[1]: true}
				for _, className := range remaining {
					if !merged[className] {
						next = append(next, className)
					}
				}
				remaining = append(next, output)
			}
		}
	}

	t.Logf("enforce-canonical-classes: asked about %d classes, placed %d, reported %d findings",
		asked, placed, findings)
	t.Logf("  declined: %d that read as static or arbitrary, %d other", droppedStatics, droppedOther)
	for _, output := range livePlacementTopKeys(outputs, 10) {
		t.Logf("  %4d  => %s", outputs[output], output)
	}

	// The population guard. A rule that placed nothing reports nothing and would otherwise satisfy
	// the finding assertion below.
	if placed < 7000 {
		t.Fatalf(
			"the rule placed %d classes and the corpus places 7,287; a population that dropped this "+
				"far means a whole category of class stopped resolving to a root, and the finding "+
				"count would be measuring that silence",
			placed,
		)
	}

	// The decomposition, and it is the assertion that makes the placement drop reviewable rather than
	// merely reported. A class the live split declines for any reason other than "the engine does not
	// read this as a functional utility" is an answer the migration lost.
	if droppedOther != 0 {
		names := make([]string, 0, len(droppedExamples))
		for className := range droppedExamples {
			names = append(names, className+" ("+droppedExamples[className]+")")
		}
		sort.Strings(names)
		if len(names) > 12 {
			names = names[:12]
		}
		t.Errorf(
			"%d classes were declined despite reading as functional utilities, so the placement drop "+
				"is not fully explained by classes that have no root-plus-value structure: %s",
			droppedOther, strings.Join(names, ", "),
		)
	}

	// The finding count is what the rule actually says, and it is the number that must not move. Two
	// on this corpus, both `size-*`.
	if findings != 2 {
		t.Errorf(
			"the rule reported %d findings and the shipped rule reported 2 on this same corpus "+
				"(size-[20px] and size-[55%%]). A migration that changes what a rule says is a "+
				"behaviour change rather than a portability fix",
			findings,
		)
	}
}

// TestConflictingClassesPlacementIsAccountedFor is `no-conflicting-classes`'s measurement.
//
// Seven tables moved, the largest change of the three, so the population is the thing to watch. The
// rule resolves a class to a property set, a selector shape, a variant string and a composition flag,
// and a class it cannot resolve is skipped rather than reported — which is precisely how a migration
// of this rule could go silent without a single test failing.
//
// Measured over the corpus, before and after:
//
//	resolved  11,272 -> 11,267   (913 of them now answered by compiling an `@utility` block)
//	findings       0 ->      0
//	changed resolutions: 0       (no class resolves to a different property set)
//
// The five-occurrence drop is four classes and both are corrections rather than losses:
//
//   - `bg-linear-to-r`, `-to-b`, `-to-t`. The shipped prefix walk read these as root `bg` and gave
//     them `background-color`. They are root `bg-linear`, which sets `background-image`, and
//     `RootDeclaredProperties` has no entry for it — so the old answer was a wrong property set that
//     would collide with any real `bg-*` color class on the same element.
//   - `ring-inset`. Reads static first and `StaticDeclaredProperties` carries no entry, so declining
//     is correct. The walk forced root `ring` and answered `box-shadow`; `ring` is a composing root,
//     so it produced no finding either way.
//
// Zero findings is the right answer on this corpus and it is asserted beside the population for the
// reason the file comment gives: on its own it is satisfiable by a rule that resolved nothing.
func TestConflictingClassesPlacementIsAccountedFor(t *testing.T) {
	t.Parallel()
	literals := livePlacementLiterals(t)

	asked, resolved, findings := 0, 0, 0
	unresolvedExamples := map[string]int{}

	for entryPoint, lists := range literals {
		designSystem := livePlacementSystem(t, entryPoint)
		if designSystem.Err != nil {
			continue
		}
		for _, classes := range lists {
			seen := make(map[string]bool, len(classes))
			distinct := make([]string, 0, len(classes))
			for _, className := range classes {
				asked++
				if seen[className] {
					continue
				}
				seen[className] = true
				distinct = append(distinct, className)
				if _, canResolve := resolveClassFactsIn(className, designSystem); canResolve {
					resolved++
					continue
				}
				unresolvedExamples[className]++
			}
			findings += len(conflictFindingsIn(distinct, designSystem))
		}
	}

	t.Logf("no-conflicting-classes: asked about %d classes, resolved %d, reported %d findings",
		asked, resolved, findings)
	for _, example := range livePlacementTopKeys(unresolvedExamples, 20) {
		t.Logf("  %4d  %s", unresolvedExamples[example], example)
	}

	// The population guard, and this rule is the one it matters most for: seven tables moved at once,
	// and a rule that resolved nothing would report nothing and look exactly like a clean corpus.
	if resolved < 11000 {
		t.Fatalf(
			"the rule resolved %d classes to a property set and the corpus resolves 11,389; a "+
				"population that dropped this far means a whole category of class stopped resolving, "+
				"and the finding count below would be measuring its silence",
			resolved,
		)
	}

	if findings != 0 {
		t.Errorf(
			"the rule reported %d findings and the shipped rule reported 0 on this same corpus. "+
				"Every literal here is one a real repository wrote, so a finding is a conflict "+
				"between classes that work rather than a defect found",
			findings,
		)
	}
}

// livePlacementTopKeys returns the most frequent keys of a count map, for readable logs.
func livePlacementTopKeys(counts map[string]int, limit int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(left int, right int) bool {
		if counts[keys[left]] != counts[keys[right]] {
			return counts[keys[left]] > counts[keys[right]]
		}
		return keys[left] < keys[right]
	})
	if len(keys) > limit {
		keys = keys[:limit]
	}
	return keys
}

// TestConflictingClassesPerRepositoryPlacement reports the placement counts the way #mz0m6k8 asked.
//
// The aggregate test above answers "did the rule go quiet" over both repositories at once, which is
// the question that matters most and is not the whole question. A rule can hold its total while one
// repository's half collapses and the other's grows, and this corpus is two repositories whose
// themes differ, which is exactly the situation where that could happen unnoticed.
//
// Measured before and after the table deletion, per repository:
//
//	ahra                asked 5947, resolved 5924  ->  resolved 5936
//	www-connected-app   asked 5367, resolved 5343  ->  resolved 5352
//	findings                     0             0   ->            0
//
// Both halves moved up, which is the direction the change predicts: the deleted tables were generated
// from a snapshot of one repository and were structurally blind to roots either repository added
// afterwards, while the computed path reads each repository's own design system.
func TestConflictingClassesPerRepositoryPlacement(t *testing.T) {
	t.Parallel()
	literals := livePlacementLiterals(t)

	type placement struct{ asked, resolved, findings int }
	byRepository := map[string]*placement{}

	for entryPoint, lists := range literals {
		designSystem := livePlacementSystem(t, entryPoint)
		if designSystem.Err != nil {
			continue
		}
		counts := byRepository[entryPoint]
		if counts == nil {
			counts = &placement{}
			byRepository[entryPoint] = counts
		}
		for _, classes := range lists {
			seen := make(map[string]bool, len(classes))
			distinct := make([]string, 0, len(classes))
			for _, className := range classes {
				counts.asked++
				if seen[className] {
					continue
				}
				seen[className] = true
				distinct = append(distinct, className)
				if _, canResolve := resolveClassFactsIn(className, designSystem); canResolve {
					counts.resolved++
				}
			}
			counts.findings += len(conflictFindingsIn(distinct, designSystem))
		}
	}

	// An unset corpus has already skipped this test, so both are set and loading neither is a failure.
	if len(byRepository) == 0 {
		t.Fatal("both corpora are set and neither design system loaded, so there is nothing to measure")
	}

	entryPoints := make([]string, 0, len(byRepository))
	for entryPoint := range byRepository {
		entryPoints = append(entryPoints, entryPoint)
	}
	sort.Strings(entryPoints)

	for _, entryPoint := range entryPoints {
		counts := byRepository[entryPoint]
		t.Logf("%s: asked %d, resolved %d, %d findings", entryPoint, counts.asked, counts.resolved, counts.findings)

		// Per repository rather than in total, which is the point of this test. A repository whose
		// half stopped resolving would be invisible in the aggregate if the other half grew.
		if counts.resolved*10 < counts.asked*9 {
			t.Errorf("%s resolved %d of %d classes; a repository resolving under nine tenths of what it "+
				"was asked has had a whole category of class stop resolving, and its finding count "+
				"below would be measuring that silence", entryPoint, counts.resolved, counts.asked)
		}
		if counts.findings != 0 {
			t.Errorf("%s reported %d findings and every literal in this corpus is one that repository "+
				"actually wrote, so a finding is a conflict between classes that work", entryPoint, counts.findings)
		}
	}
}

// The resolution comparison that justified the deletion, and why it is not here.
//
// `TestConflictingClassesResolutionsDidNotChange` compared the computed answer against the four
// deleted tables, class for class, over this corpus: 11,260 class occurrences answered by both,
// 11,182 identical, 11 different. All 11 were `text-[<size>]` with no modifier, where the tables
// said `{font-size, line-height}` and the computed answer says `font-size` alone. Upstream at v4.3.3
// returns `[decl('font-size', value)]` outside `if (candidate.modifier)`, so the computed side is
// right and the correction makes the rule stricter rather than quieter.
//
// It is deleted along with the tables it read, because a comparison with one side missing is not a
// comparison. What survives is the per-repository placement above, which catches the rule going
// quiet, and `TestComputedPropertiesMakeTheDistinctionsTheOverrideTablesHeld` in the engine package,
// which pins each distinction the deleted tables encoded, including this `text` one in both
// directions.

// A repository's own colour tokens are recognised as colours, which the deleted table could not do.
//
// `valueIsColorIn` read `ColorNames`, 561 names generated from ahra, and www-connected-app declares
// three the table did not carry. A class using one of them read as not-a-colour, so `bg-brand` was
// answered on the wrong arm: `background-image` rather than `background-color`, which is a wrong
// property set rather than a missing answer, and it fails to conflict with `bg-red-500` when it
// should.
//
// Asserted against the live theme rather than a list, so a repository adding a colour joins this
// test rather than needing to be added to it. The count is asserted non-zero because a theme that
// resolved nothing would pass every comparison below without measuring anything.
func TestRepositoryColorTokensReadAsColors(t *testing.T) {
	t.Parallel()
	var checkedRepositories, checkedTokens int

	for entryPoint := range livePlacementLiterals(t) {
		designSystem := livePlacementSystem(t, entryPoint)
		if designSystem.Err != nil {
			continue
		}
		checkedRepositories++

		for _, key := range designSystem.System.Theme().KeysInNamespaces([]string{"--color"}) {
			if !valueIsColorIn("bg-"+key, "bg", designSystem.System) {
				t.Errorf("%s declares --color-%s and `bg-%s` does not read as a colour", entryPoint, key, key)
			}
			checkedTokens++
		}
	}

	t.Logf("colour tokens read from live themes: %d across %d repositories", checkedTokens, checkedRepositories)

	// An unset corpus has already skipped this test, so an empty count is a failure.
	if checkedRepositories == 0 || checkedTokens == 0 {
		t.Fatal("no design system loaded, so this test measured nothing")
	}
}

// TestColorTokensReadAsColorsOnThePublicTheme is TestRepositoryColorTokensReadAsColors on the
// independent theme, so it runs on every machine (#f598zk0): every `--color-*` key reads as a colour,
// the framework palette's and the theme's own `--color-weird-*` alike.
func TestColorTokensReadAsColorsOnThePublicTheme(t *testing.T) {
	t.Parallel()
	system := publicDesignSystem(t).System

	checkedTokens, ownTokens := 0, 0
	for _, key := range system.Theme().KeysInNamespaces([]string{"--color"}) {
		if !valueIsColorIn("bg-"+key, "bg", system) {
			t.Errorf("the public theme declares --color-%s and `bg-%s` does not read as a colour", key, key)
		}
		checkedTokens++
		if strings.HasPrefix(key, "weird-") {
			ownTokens++
		}
	}
	// The theme's own keys are the ones no framework table could vouch for, so they must be among
	// those checked, or the test proves only the palette.
	if checkedTokens < 200 || ownTokens != 2 {
		t.Fatalf("checked %d colour tokens, %d of them the theme's own; want the framework palette and both --color-weird keys",
			checkedTokens, ownTokens)
	}
}
