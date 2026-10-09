package tailwind

import (
	"os"
	"path/filepath"
	"testing"

	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind/vendored"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// # Why every fixture here runs through a program
//
// Same reason the other three migrated rules' fixtures give. The rule now reads the live design
// system to resolve a class to its declared properties, so it declares `ProgramReads` and returns
// nil listeners when the Program is nil. `rule_testing.Run` hands it exactly that, so a fixture left on
// it would exercise the nil-Program branch: the reporting half fails loudly and the silent half —
// which this rule's own header calls the load-bearing one — passes while proving nothing.
//
// `TestConflictFixturesActuallyRan` is what stops the converted file from going green on nothing in
// the other direction, where the walk for an installed tailwindcss finds none and every fixture
// skips while `go test` prints ok.

// runConflictFixture runs the rule against a one-file program that has a real design system.
//
// The fixture stylesheet declares no tokens of its own. This rule's subject is what a class declares,
// and the classes these fixtures use are the framework's, so a fixture theme adding tokens would test
// the fixture. The repository half is measured against the real corpus in live_placement_test.go,
// where 913 class occurrences are answered by compiling an `@utility` block.
func runConflictFixture(t *testing.T, fileName string, source string) rule_testing.Result {
	t.Helper()
	return runConflictFixtureWithOptions(t, fileName, source, nil)
}

// runConflictFixtureWithOptions is runConflictFixture for the cases that configure the surfaces.
func runConflictFixtureWithOptions(
	t *testing.T,
	fileName string,
	source string,
	options any,
) rule_testing.Result {
	t.Helper()

	packageRoot := unknownFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine, so the rule's design system cannot be " +
			"built and these fixtures would measure a decline rather than a conflict")
	}

	files := map[string]string{
		fileName:                     source,
		unknownFixtureStylesheetPath: `@import "tailwindcss";`,
	}
	plantPackage := func(root string) {
		modules := filepath.Join(root, "node_modules")
		if err := os.MkdirAll(modules, 0o755); err != nil {
			t.Fatalf("creating the fixture node_modules: %v", err)
		}
		if err := os.Symlink(packageRoot, filepath.Join(modules, "tailwindcss")); err != nil {
			t.Fatalf("linking the installed tailwindcss into the fixture: %v", err)
		}
	}

	if options == nil {
		return rule_testing.RunTypedFilesWithSetup(t, NoConflictingClasses, files, fileName, plantPackage)
	}
	return rule_testing.RunTypedFilesWithSetupAndOptions(
		t, NoConflictingClasses, files, fileName, options, plantPackage)
}

// Expectations measured by running `better-tailwindcss/no-conflicting-classes` over probe fixtures
// against the real plugin. Each line is a verdict, and the silent ones are the load-bearing half:
// three of them are cases a plausible port reports and upstream does not.
//
//	flex block                2 findings, one anchored at each class
//	px-4 px-8                 2 findings, on "padding-inline"
//	text-left text-right      2 findings, on "text-align"
//	hover:flex hover:block    2 findings, same variant
//	p-4 px-8                  silent, padding and padding-inline are different names
//	flex hover:block          silent, different states
//	w-8 h-8                   silent, width and height are different properties
//	mt-2 mb-2                 silent
//	flex items-center         silent

func TestNoConflictingClassesReportsSymmetrically(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
		wantIds  []string
	}{
		{
			// The headline case, and the one a table that lost `flex`'s static reading would miss.
			name:     "two display utilities",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex block" />;`,
			wantIds:  []string{"conflictingClasses", "conflictingClasses"},
		},
		{
			// Same property, different values.
			name:     "same functional root twice",
			fileName: "Component.tsx",
			source:   `const element = <div className="px-4 px-8" />;`,
			wantIds:  []string{"conflictingClasses", "conflictingClasses"},
		},
		{
			name:     "two alignment utilities",
			fileName: "Component.tsx",
			source:   `const element = <div className="text-left text-right" />;`,
			wantIds:  []string{"conflictingClasses", "conflictingClasses"},
		},
		{
			// Same variant on both, so they genuinely collide in that state.
			name:     "conflict inside one variant",
			fileName: "Component.tsx",
			source:   `const element = <div className="hover:flex hover:block" />;`,
			wantIds:  []string{"conflictingClasses", "conflictingClasses"},
		},
		{
			name:     "on a callee surface",
			fileName: "Component.tsx",
			source:   `const merged = cn('flex block');`,
			wantIds:  []string{"conflictingClasses", "conflictingClasses"},
		},
		{
			name:     "on a variable surface",
			fileName: "Styles.ts",
			source:   `const className = 'px-4 px-8';`,
			wantIds:  []string{"conflictingClasses", "conflictingClasses"},
		},
		{
			// Three classes on one property is three findings, each naming the other two.
			name:     "three classes on one property",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex block grid" />;`,
			wantIds:  []string{"conflictingClasses", "conflictingClasses", "conflictingClasses"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runConflictFixture(t, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The silent half, which for this rule is where the real difficulty is.
//
// Each of these is a case a reasonable port reports and upstream does not, so getting them wrong
// means firing on correct code, which is how a rule gets turned off.
func TestNoConflictingClassesStaysSilent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// Different properties under the same value. These collapse into `size-8`, which is
			// `enforce-canonical-classes`'s finding rather than this rule's.
			name:     "width and height",
			fileName: "Component.tsx",
			source:   `const element = <div className="w-8 h-8" />;`,
		},
		{
			// Shorthands are not normalised: `padding` and `padding-inline` are different names.
			name:     "padding and padding-inline",
			fileName: "Component.tsx",
			source:   `const element = <div className="p-4 px-8" />;`,
		},
		{
			// Different states, so both take effect where they belong. A port comparing bare class
			// names reports this and is wrong.
			name:     "same property under different variants",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex hover:block" />;`,
		},
		{
			name:     "different axes of margin",
			fileName: "Component.tsx",
			source:   `const element = <div className="mt-2 mb-2" />;`,
		},
		{
			name:     "unrelated properties",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex items-center gap-2" />;`,
		},
		{
			name:     "one class",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex" />;`,
		},
		{
			// A repeat is `no-duplicate-classes`'s finding. Pairing a class with itself would report
			// every duplicate in the tree as a conflict.
			name:     "the same class twice",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex flex" />;`,
		},
		{
			name:     "unrelated attribute",
			fileName: "Component.tsx",
			source:   `const element = <div title="flex block" />;`,
		},
		{
			name:     "unrelated callee",
			fileName: "Component.tsx",
			source:   `const value = someOtherFunction('flex block');`,
		},
		{
			// Different elements. `divide-*` emits under `:where(.CLASS > :not(:last-child))` and
			// styles children; `border-*` styles the element. Both declare `border-color` and they
			// do not collide. This is the pair that made the rule look engine-bound.
			name:     "divide against border",
			fileName: "Component.tsx",
			source:   `const element = <div className="divide-neutral-200 border-neutral-200" />;`,
		},
		{
			// Both declare `box-shadow` and both layer through their own custom properties, so they
			// compose rather than overwrite.
			name:     "shadow against ring",
			fileName: "Component.tsx",
			source:   `const element = <div className="shadow-lg ring-1" />;`,
		},
		{
			// `border` sets style plus width, `border-dashed` sets style alone. Overlapping is not
			// enough: the whole property set has to match.
			name:     "border against border-dashed",
			fileName: "Component.tsx",
			source:   `const element = <div className="border border-dashed" />;`,
		},
		{
			// `ring-inset` declares only a custom property, so it has nothing to collide with.
			name:     "ring against ring-inset",
			fileName: "Component.tsx",
			source:   `const element = <div className="ring-1 ring-inset" />;`,
		},
		{
			// `container` emits several rules and is deliberately absent from the tables, so it is
			// silent rather than wrong.
			name:     "container against w-full",
			fileName: "Component.tsx",
			source:   `const element = <div className="w-full container" />;`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runConflictFixture(t, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestConflictingClassesProposeSuggestionsNotFixes guards the distinction that decides whether the
// rule silently rewrites markup.
//
// Upstream declares `autofix: true`. On our side removing either of two conflicting classes changes
// what renders, and only the author knows which one was meant, so this must be a Suggestion. The
// edit engine never applies suggestions, which makes the difference concrete: ported as a Fix it
// would silently pick one; ported as a suggestion a human chooses.
//
// This was caught by `@system_cohere_lint_fix`'s research pass rather than by reading the source,
// where the `autofix` flag reads as an instruction.
func TestConflictingClassesProposeSuggestionsNotFixes(t *testing.T) {
	t.Parallel()
	result := runConflictFixture(t, "Component.tsx",
		`const element = <div className="flex block" />;`)

	if len(result.Diagnostics) == 0 {
		t.Fatal("expected findings, got none")
	}
	for _, diagnostic := range result.Diagnostics {
		if len(diagnostic.Fixes) != 0 {
			t.Fatal("this rule must not propose an automatic fix: removing either conflicting class " +
				"changes what renders, and the edit engine would apply the choice unattended")
		}
	}
}

// TestPropertyLookupUsesTheLongestRoot is the known-dirty control.
//
// Resolving a class to its properties by taking the first dash-delimited segment is the obvious
// implementation and it is wrong in the direction that reports false conflicts: `border-l-4` would
// resolve as `border` rather than `border-l`, so it would be given `border-width` instead of
// `border-left-width` and would read as conflicting with `border-r-4`.
//
// That is the same shortcut, in a different guise, that silently stopped `border-x` being reported
// during the migration.
func TestPropertyLookupUsesTheLongestRoot(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {

		// The real lookup.
		left, canResolveLeft := resolveClassFactsIn("border-l-4", designSystem)
		right, canResolveRight := resolveClassFactsIn("border-r-4", designSystem)
		if !canResolveLeft || !canResolveRight {
			t.Fatal("border-l-4 or border-r-4 did not resolve, so this control proves nothing")
		}
		leftProperties, rightProperties := left.Properties, right.Properties

		if len(leftProperties) == 0 || len(rightProperties) == 0 {
			t.Fatal("border-l-4 or border-r-4 resolved to no properties, so this control proves nothing")
		}

		if len(sharedProperties(leftProperties, rightProperties)) != 0 {
			t.Errorf("border-l-4 declares %v and border-r-4 declares %v, which share a property. They "+
				"are different edges and must not be reported as conflicting.", leftProperties, rightProperties)
		}

		// And the shortcut it guards against: taking the first segment gives both the same root. The
		// root now comes from `parseCandidate` rather than a longest-prefix walk over a generated table,
		// so this asserts the parser's answer rather than the walk's.
		if root, _ := functionalRootIn("border-l-4", designSystem.System); root == "border" {
			t.Error("the root lookup returned the shortest match rather than the longest, which gives " +
				"every border edge the same properties")
		}

		// The lookup must also not let a shorter root match across a value boundary.
		if root, _ := functionalRootIn("px-4", designSystem.System); root == "p" {
			t.Error("`p` matched `px-4`, so padding and padding-inline would be conflated and `p-4 px-8` " +
				"would report a conflict upstream does not report")
		}

		// A root the shipped prefix walk got wrong on real code. `bg-linear-to-r` is root `bg-linear`,
		// which sets `background-image`; the walk read root `bg` and answered `background-color`, a
		// property the class does not declare and which collides with every real `bg-*` color class.
		if root, isFunctional := functionalRootIn("bg-linear-to-r", designSystem.System); !isFunctional ||
			root != "bg-linear" {
			t.Errorf("bg-linear-to-r resolved to root %q, want bg-linear", root)
		}
	})
}

// TestRepositoryUtilitiesAreCompiledRatherThanLookedUp pins the half of this rule that went live.
//
// A repository's own `@utility` block is answered by compiling it, which is the only source that can
// answer for a repository this port has never seen. The framework's utilities stay on the generated
// tables because nothing in shipped Go compiles them, and conflict_facts.go's file comment records
// why a sort reading is not a substitute.
//
// Asserted against a stylesheet that declares its own utility, so the compiled answer cannot be
// coming from a table: `synthetic-static` sets `display` and `opacity`, and no generated table in
// this repository contains the name.
func TestRepositoryUtilitiesAreCompiledRatherThanLookedUp(t *testing.T) {
	t.Parallel()
	packageRoot := unknownFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine")
	}

	// Two classes under one design system: one the repository declares, one the framework does.
	result := runConflictFixture(t, "Component.tsx",
		`const element = <div className="flex block" />;`)
	rule_testing.ExpectFindings(t, result, "conflictingClasses", "conflictingClasses")

	// And the repository half, against a class the corpus repository declares and no framework
	// contains. `background--0` is a functional `@utility` block in ahra's own stylesheet, compiled
	// here rather than looked up.
	ahra := DesignSystemResult{System: unknownFixtureLiveSystem(t)}
	facts, canResolve := resolveClassFactsIn("background--0", ahra)
	if !canResolve {
		t.Fatal("background--0 did not resolve, so the repository half is not being compiled")
	}
	if len(facts.Properties) != 1 || facts.Properties[0] != "background-color" {
		t.Errorf("background--0 resolved to %v, want [background-color] from compiling its @utility "+
			"block", facts.Properties)
	}

	// The population, so this is not one lucky class. Every corpus class the repository half answers
	// is one whose properties come from compiling a block rather than from any table.
	compiled := 0
	for _, className := range []string{
		"background--0", "background--1", "background--2", "background--3",
		"content--0", "content--1", "content--placeholder", "content--positive",
		"border--0", "border--focus", "hover:background--1", "dark:content--2",
	} {
		if _, wasCompiled := repositoryClassFacts(className, "", ahra); wasCompiled {
			compiled++
		}
	}
	if compiled < 10 {
		t.Errorf("only %d of 12 repository utilities were answered by compiling their @utility "+
			"block; the repository half is not carrying the population it should", compiled)
	}

	// The complement, and it asserts the SOURCE rather than the answer. `background--0` still
	// resolves against a design system that does not declare it, because `RootDeclaredProperties`
	// carries the root `background-` — the same generated-table leak
	// TestRepositoryNamesInTheFrameworkTablesAreStillVouchedFor pins for `KnownStatics`. What must
	// differ is which half answered: the repository's compiled block on ahra, the framework table
	// elsewhere.
	independent := DesignSystemResult{System: independentLiveSystem(t)}
	if _, compiledHere := repositoryClassFacts("background--0", "", ahra); !compiledHere {
		t.Error("background--0 was not answered by compiling ahra's own @utility block, so the " +
			"repository half is not being consulted")
	}
	if _, compiledElsewhere := repositoryClassFacts("background--0", "", independent); compiledElsewhere {
		t.Error("background--0 was answered by compiling a block on a design system that declares " +
			"none, so the repository half is not reading the repository in front of the rule")
	}
}

// TestStaticRepositoryUtilitiesStillReadFromTheFrameworkTables pins a gap this task did not close.
//
// The repository half of the resolution compiles functional `@utility` blocks, which is where 913 of
// the corpus's class occurrences are answered. A STATIC `@utility` block cannot be compiled from
// here: `LoadedDesignSystem.staticUtilityNodes` is unexported and has no accessor, and the only path
// to it is `Table.Statics`, which holds a `{order, count}` reading rather than property names.
// conflict_facts.go's file comment records why a reading is not a substitute for a declaration list.
//
// So `synthetic-static` and its kind fall through to `StaticDeclaredProperties`, which carries 16
// names ahra declares in its own stylesheet and nothing from any other repository. The consequence
// is that a repository's static `@utility` blocks are resolvable exactly when the generated table
// happens to contain them.
//
// Not closed here because the accessor belongs to `internal/tailwind` and the seam it would cross is
// the same one #gnqbn4b tracks. The error direction is the safe one: an unresolvable class is skipped
// rather than reported, so the rule under-reports on a repository whose statics it cannot see.
//
// Pinned on the CURRENT behaviour so closing it upstream breaks this test deliberately.
func TestStaticRepositoryUtilitiesStillReadFromTheFrameworkTables(t *testing.T) {
	t.Parallel()
	independent := DesignSystemResult{System: independentLiveSystem(t)}

	if _, canResolve := resolveClassFactsIn("synthetic-static", independent); canResolve {
		t.Error("a static @utility block now resolves against the repository that declares it, which " +
			"means the design system grew a way to reach static utility bodies. That is the fix this " +
			"test exists to notice: delete it and assert the corrected behaviour instead")
	}

	// The complement: a FUNCTIONAL block on the same stylesheet does resolve, so the gap is about
	// static blocks specifically rather than about the repository half being inert.
	if _, canResolve := resolveClassFactsIn("synthetic-fn-small", independent); !canResolve {
		t.Error("a functional @utility block did not resolve either, so the repository half is not " +
			"working at all and this test is measuring the wrong thing")
	}
}

// TestConflictFixturesActuallyRan is what stops this file from going green on nothing.
//
// Every fixture above skips when no installed tailwindcss can be found, and `go test` prints ok for a
// package whose every case skipped, so a mistake in where the walk starts looks identical to a
// machine that lacks the package.
func TestConflictFixturesActuallyRan(t *testing.T) {
	t.Parallel()
	packageRoot := unknownFixturePackageRoot()
	if packageRoot == "" {
		t.Fatalf(
			"the vendored tailwindcss at %s is missing, so every fixture in this file would skip and the "+
				"package still report ok; restore it (internal/lint/rules/tailwind/vendored)",
			vendored.TailwindPackageRoot(),
		)
	}

	result := runConflictFixture(t, "Component.tsx",
		`const element = <div className="flex block" />;`)
	rule_testing.ExpectFindings(t, result, "conflictingClasses", "conflictingClasses")
}

// TestVariantsAreComparedNotStripped pins the distinction that decides false positives.
//
// Two classes conflict only in the same state. Comparing bare names reports `flex hover:block`,
// which is correct code, and a rule that fires on correct code is one somebody turns off.
func TestVariantsAreComparedNotStripped(t *testing.T) {
	t.Parallel()
	sameVariant := runConflictFixture(t, "Component.tsx",
		`const element = <div className="hover:flex hover:block" />;`)
	if len(sameVariant.Diagnostics) == 0 {
		t.Fatal("two display utilities under the same variant do collide and must be reported")
	}

	differentVariant := runConflictFixture(t, "Component.tsx",
		`const element = <div className="flex hover:block" />;`)
	if len(differentVariant.Diagnostics) != 0 {
		t.Fatalf("classes in different states both take effect where they belong, so this is correct "+
			"code; got %d findings", len(differentVariant.Diagnostics))
	}
}

// TestSelectorShapeCannotProduceAWrongFinding closes the `RootSelectorShapes` question.
//
// The table is repository-invariant across the systems that exist, measured by generating against
// `independent_theme.css` and diffing: 327 roots and 53,301 pairs on ahra against 302 and 45,451 on
// the independent system, both producing the same 8 entries with none on either side alone.
//
// That is not sufficient on its own, and this test is the reason. A repository `@utility` block can
// declare a nested selector, and one that does contributes a ninth entry to the generator's output:
// `@utility gutter-*` wrapping `& > :not(:last-child)` enumerates as
// `gutter -> .CLASS > :not(:last-child)`. So a repository could hold a root whose shape the table
// does not carry, and the invariance measurement could not have found that.
//
// What makes the table safe to keep is that such a class cannot reach the comparison the shape
// decides. A nested block compiles to a single `rule` node carrying no property, `repositoryClassFacts`
// counts only top-level declarations, and the class resolves to false and is skipped before pairing.
//
// Asserted on a design system built for this, rather than on the corpus, because neither corpus
// repository writes a nested `@utility` block and a test that could not construct the case would be
// asserting its absence.
func TestSelectorShapeCannotProduceAWrongFinding(t *testing.T) {
	t.Parallel()
	packageRoot := unknownFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine")
	}

	staging := t.TempDir()
	if err := os.MkdirAll(filepath.Join(staging, "node_modules"), 0o755); err != nil {
		t.Fatalf("creating the staging node_modules: %v", err)
	}
	if err := os.Symlink(packageRoot, filepath.Join(staging, "node_modules", "tailwindcss")); err != nil {
		t.Fatalf("linking the installed tailwindcss: %v", err)
	}

	// A repository utility that emits under a nested selector, which is the one mechanism that can
	// put a repository token into RootSelectorShapes.
	stylesheet := `@import 'tailwindcss';
@theme { --frobnicate-small: 2px; }
@utility gutter-* {
    & > :not(:last-child) { margin-inline-end: --value(--frobnicate-*, [length]); }
}
`
	entryPoint := filepath.Join(staging, "theme.css")
	if err := os.WriteFile(entryPoint, []byte(stylesheet), 0o644); err != nil {
		t.Fatalf("writing the staging stylesheet: %v", err)
	}

	system, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
		EntryPoint:          entryPoint,
		TailwindPackageRoot: findTailwindPackageRoot(staging, diskFileExists),
	})
	if err != nil {
		t.Fatalf("loading the nested-utility design system: %v", err)
	}
	designSystem := DesignSystemResult{System: system, Table: tailwindengine.NewTable(system)}

	// The class whose shape the table does not carry declines rather than resolving to a wrong one.
	if _, canResolve := resolveClassFactsIn("gutter-small", designSystem); canResolve {
		t.Error("a nested-selector repository utility now resolves, so its selector shape decides a " +
			"comparison and RootSelectorShapes cannot answer for it. That is the change this test " +
			"exists to notice: the table must move to the live system, or the shape must be derived")
	}

	// And the classes it would have been compared against still resolve, so the silence above is the
	// nested class declining rather than the whole design system failing to load.
	for _, className := range []string{"me-4", "mr-4"} {
		if _, canResolve := resolveClassFactsIn(className, designSystem); !canResolve {
			t.Fatalf("%s did not resolve, so this test is measuring a broken design system rather "+
				"than the nested-utility case", className)
		}
	}

	// The finding count, which is the thing that actually matters. `gutter-small` lands on child
	// elements and `me-4` lands on the element itself; both declare margin-inline-end, so a rule that
	// resolved the first without its shape would report a conflict on correct code.
	for _, pair := range [][]string{{"gutter-small", "me-4"}, {"gutter-small", "mr-4"}} {
		if findings := conflictFindingsIn(pair, designSystem); len(findings) != 0 {
			t.Errorf("%v produced %d findings; a nested-selector utility and a class on the element "+
				"itself do not collide, and reporting them is the false positive a missing selector "+
				"shape would cause", pair, len(findings))
		}
	}
}
