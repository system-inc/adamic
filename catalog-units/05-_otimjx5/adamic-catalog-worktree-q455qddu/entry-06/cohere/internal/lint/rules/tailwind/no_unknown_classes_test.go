package tailwind

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind/vendored"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// Expectations measured against the real plugin, and against the real corpus for the silent half.
// The silent cases matter more than usual here: this rule's failure mode is reporting a class that
// works, and a rule that flags working classes gets turned off the first time it does it.
//
// # Why every fixture here runs through a program
//
// Same reason `enforce_consistent_class_order_test.go` gives, and the same trap. The rule now reads
// the live design system, so it declares `ProgramReads` and returns nil listeners when the Program
// is nil. `rule_testing.Run` hands it exactly that. So every fixture in this file that kept using
// `rule_testing.Run` after the swap would exercise the nil-Program branch: the reporting half would fail
// loudly and the whole silent half would pass while proving nothing.
//
// Measured before these were converted: with the rule swapped and the fixtures still on
// `rule_testing.Run`, the eight reporting cases failed and all thirteen silent cases passed. Thirteen
// green assertions over a rule that never ran.
//
// `TestUnknownClassFixturesActuallyRan` is what stops the converted file from going green on nothing
// in the other direction, where the walk for an installed tailwindcss finds none and every fixture
// skips while `go test` prints ok.

// unknownFixtureStylesheetPath is where the fixture writes its root stylesheet.
//
// The first entry in `tailwindEntryPointCandidates`, so the rule's own entry-point search finds it
// by the same path it uses on a real repository.
const unknownFixtureStylesheetPath = "app/_theme/styles/theme.css"

// unknownFixtureStylesheet is the fixture's root stylesheet, and it declares utilities of its own.
//
// This is the one place this file's fixtures differ from the class-order ones, and the difference is
// the whole point of the migration. Those declare no tokens deliberately, because per-repository
// behaviour is not what they measure. Here it is exactly what is measured: a class this repository
// declares and no framework contains must read as known, and the shipped table could only answer
// that for the repository it was generated from.
//
// The names come from `internal/lint/rules/tailwind/tools/generate_descriptor_base/testdata/independent_theme.css`, which
// exists so a framework fact can be told from a repository fact. `synthetic-static` and
// `synthetic-fn-*` appear in no framework table and in neither corpus repository, so a rule that
// calls them known has consulted this stylesheet rather than a table.
const unknownFixtureStylesheet = `@import "tailwindcss";

@theme {
    --frobnicate-small: 2px;
    --frobnicate-large: 9px;
}

@utility synthetic-static {
    display: grid;
    opacity: 0.5;
}

@utility synthetic-fn-* {
    margin: --value(--frobnicate-*, [length]);
}`

// unknownFixturePackageRoot is the installed tailwindcss the fixture stylesheet imports: the vendored
// snapshot, so the fixtures run on any machine (#sycrdr6). It used to be found by walking up from ahra's
// checkout, which only one laptop has.
func unknownFixturePackageRoot() string {
	return vendored.TailwindPackageRoot()
}

// runUnknownFixture runs the rule against a one-file program that has a real design system.
//
// The symlink is planted before the program is built, into the same temp directory
// `rule_testing.RunTypedFiles` writes the fixture files to, so the rule's own upward walk finds it.
// Pointing the `@import` at an absolute path instead does not work: the rule never resolves the
// import itself, it walks up for `node_modules/tailwindcss` and declines when the walk fails.
func runUnknownFixture(t *testing.T, fileName string, source string) rule_testing.Result {
	t.Helper()
	return runUnknownFixtureWithOptions(t, fileName, source, nil)
}

// runUnknownFixtureWithOptions is runUnknownFixture for the cases that configure the ignore list.
func runUnknownFixtureWithOptions(
	t *testing.T,
	fileName string,
	source string,
	options any,
) rule_testing.Result {
	t.Helper()

	packageRoot := unknownFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine, so the rule's design system cannot be " +
			"built and these fixtures would measure a decline rather than an existence check")
	}

	files := map[string]string{
		fileName:                     source,
		unknownFixtureStylesheetPath: unknownFixtureStylesheet,
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
		return rule_testing.RunTypedFilesWithSetup(t, NoUnknownClasses, files, fileName, plantPackage)
	}
	return rule_testing.RunTypedFilesWithSetupAndOptions(
		t, NoUnknownClasses, files, fileName, options, plantPackage)
}

func TestNoUnknownClassesReportsTypos(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
		wantIds  []string
	}{
		{
			name:     "misspelled utility",
			fileName: "Component.tsx",
			source:   `const element = <div className="flx items-center" />;`,
			wantIds:  []string{"unknownClass"},
		},
		{
			name:     "invented utility",
			fileName: "Component.tsx",
			source:   `const element = <div className="not-a-real-class-xyz" />;`,
			wantIds:  []string{"unknownClass"},
		},
		{
			// A project's own CSS class. Real, but not Tailwind's, so it needs the ignore list rather
			// than a fix. Reported so the choice is deliberate.
			name:     "stylesheet class Tailwind does not know",
			fileName: "Component.tsx",
			source:   `const element = <div className="ahralia-splash" />;`,
			wantIds:  []string{"unknownClass"},
		},
		{
			// Misspellings that start with a real root and continue without a separator. These are
			// the cases the value-boundary check exists for: `px4` begins with root `px`, `wfull`
			// with `w`, `textcenter` with `text`. Dropping the boundary check makes every one of
			// them look valid, which is a rule that catches almost no typos while appearing to work.
			//
			// The boundary is now `parseCandidate`'s rather than this rule's own prefix walk, which
			// is the correction: the walk approximated where a root ends, and the same approximation
			// in a different guise is what lost `border-x` during the migration.
			name:     "typo starting with a real root",
			fileName: "Component.tsx",
			source:   `const element = <div className="px4" />;`,
			wantIds:  []string{"unknownClass"},
		},
		{
			name:     "another root-prefixed typo",
			fileName: "Component.tsx",
			source:   `const element = <div className="wfull textcenter" />;`,
			wantIds:  []string{"unknownClass", "unknownClass"},
		},
		{
			name:     "two unknowns in one literal",
			fileName: "Component.tsx",
			source:   `const element = <div className="flx grd" />;`,
			wantIds:  []string{"unknownClass", "unknownClass"},
		},
		{
			// A variant this design system never declared makes the whole class unreadable, which is
			// correct: `notavariant:flex` compiles to nothing. Only reachable now that the variants
			// are parsed by the same design system that answers for the utility; the shipped rule
			// stripped everything before the last colon and never looked at it.
			name:     "unknown variant on a real utility",
			fileName: "Component.tsx",
			source:   `const element = <div className="notavariant:flex" />;`,
			wantIds:  []string{"unknownClass"},
		},
		{
			name:     "on a callee surface",
			fileName: "Component.tsx",
			source:   `const merged = cn('flx');`,
			wantIds:  []string{"unknownClass"},
		},
		{
			name:     "on a variable surface",
			fileName: "Styles.ts",
			source:   `const className = 'not-a-real-class-xyz';`,
			wantIds:  []string{"unknownClass"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runUnknownFixture(t, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The silent half, and every case here is one an earlier version of this rule reported.
//
// Answering existence from the property tables produced 12 findings on valid classes across the real
// corpus. These are those shapes, pinned so the rule cannot regress into asking the wrong question.
func TestNoUnknownClassesStaysSilent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			name:     "ordinary utilities",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex items-center gap-2" />;`,
		},
		{
			// Declares only `--tw-gradient-from`, so it is absent from the property tables and real.
			name:     "gradient stops",
			fileName: "Component.tsx",
			source:   `const element = <div className="from-black/70 to-transparent" />;`,
		},
		{
			// Emits several rules, so `declaredProperties` skips it deliberately.
			name:     "multi-rule component class",
			fileName: "Component.tsx",
			source:   `const element = <div className="container" />;`,
		},
		{
			// Markers that generate no CSS of their own; they exist to be referenced by
			// `group-hover:` and `peer-checked:` on other elements. `ParseCandidate` returns no
			// readings for them for exactly that reason, so they must be recognised before it is
			// asked or the two most common classes in any codebase get reported.
			name:     "group and peer",
			fileName: "Component.tsx",
			source:   `const element = <div className="group peer" />;`,
		},
		{
			name:     "group with a name",
			fileName: "Component.tsx",
			source:   `const element = <div className="group/item" />;`,
		},
		{
			// The author wrote the declaration directly, so there is no root to look up and nothing
			// to check it against.
			name:     "arbitrary property",
			fileName: "Component.tsx",
			source:   `const element = <div className="[font:inherit]" />;`,
		},
		{
			name:     "arbitrary value",
			fileName: "Component.tsx",
			source:   `const element = <div className="w-[13px] h-[calc(100%-3px)]" />;`,
		},
		{
			name:     "variants",
			fileName: "Component.tsx",
			source:   `const element = <div className="hover:px-4 sm:hover:flex" />;`,
		},
		{
			name:     "importance",
			fileName: "Component.tsx",
			source:   `const element = <div className="px-4!" />;`,
		},
		{
			// This fixture's own `@utility` blocks, which no generated table contains. The static
			// one and the functional one exercise the two halves of a repository's contribution.
			name:     "this repository's own static utility",
			fileName: "Component.tsx",
			source:   `const element = <div className="synthetic-static" />;`,
		},
		{
			name:     "this repository's own functional utility",
			fileName: "Component.tsx",
			source:   `const element = <div className="synthetic-fn-small synthetic-fn-large" />;`,
		},
		{
			name:     "this repository's own utility under a variant",
			fileName: "Component.tsx",
			source:   `const element = <div className="hover:synthetic-static" />;`,
		},
		{
			name:     "unrelated attribute",
			fileName: "Component.tsx",
			source:   `const element = <div title="flx" />;`,
		},
		{
			name:     "unrelated callee",
			fileName: "Component.tsx",
			source:   `const value = someOtherFunction('flx');`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runUnknownFixture(t, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestUnknownIgnoreExemptsProjectClasses covers the option this rule cannot ship without.
//
// Any project with hand-written CSS has class names Tailwind does not define. On the ahra tree those
// are the five `ahralia-splash*` classes, and they are the entire real-world finding count.
func TestUnknownIgnoreExemptsProjectClasses(t *testing.T) {
	t.Parallel()
	options := NoUnknownClassesOptions{Ignore: []string{`^ahralia-`}}

	exempted := runUnknownFixtureWithOptions(t, "Component.tsx",
		`const element = <div className="ahralia-splash flex" />;`, options)
	if len(exempted.Diagnostics) != 0 {
		t.Errorf("an ignored class should not be reported, got %d findings", len(exempted.Diagnostics))
	}

	// Without the exemption it reports, so the test above is not passing because the rule is inert.
	unexempted := runUnknownFixture(t, "Component.tsx",
		`const element = <div className="ahralia-splash flex" />;`)
	if len(unexempted.Diagnostics) == 0 {
		t.Fatal("the rule stayed silent without an ignore option, so the exemption proves nothing")
	}

	// And the exemption must not swallow real typos alongside it.
	stillReported := runUnknownFixtureWithOptions(t, "Component.tsx",
		`const element = <div className="ahralia-splash flx" />;`, options)
	if len(stillReported.Diagnostics) != 1 {
		t.Errorf("the ignore pattern should exempt only what it matches, got %d findings",
			len(stillReported.Diagnostics))
	}
}

// TestExistenceIsNotReadFromPropertyTables is the known-dirty control.
//
// The tempting implementation asks the property tables whether a class is known, because they are
// already there and every other rule uses them. It passes every violation fixture above and reports
// 12 valid classes on the real corpus, because those tables deliberately exclude utilities that
// declare nothing a rule can compare.
//
// So the shortcut is implemented here and required to be wrong about classes the real rule gets
// right.
func TestExistenceIsNotReadFromPropertyTables(t *testing.T) {
	t.Parallel()
	system := unknownFixtureLiveSystem(t)

	// Classes that are real and absent from the property tables.
	realButUndeclared := []string{
		"container",
		"from-black/70",
		"to-transparent",
		"fade-in",
	}

	lostByShortcut := 0
	for _, className := range realButUndeclared {
		if !classExistsIn(className, system) {
			t.Errorf("the real rule reports %q as unknown, and it is a valid class", className)
			continue
		}

		// The shortcut: resolve through the property tables, as the other rules do.
		if _, canResolve := resolveClassFactsIn(className, DesignSystemResult{System: system}); !canResolve {
			lostByShortcut++
		}
	}

	if lostByShortcut == 0 {
		t.Fatal("the property-table shortcut resolved every one of these, so this control no longer " +
			"demonstrates the defect it exists to demonstrate")
	}
	t.Logf("the property-table shortcut would report %d of %d valid classes as unknown",
		lostByShortcut, len(realButUndeclared))
}

// TestKnownRootWithUnknownValueIsReported is what the pinned gap became.
//
// The gap this replaces said `text-huge` had root `text`, did not compile, and was not reported,
// because catching it "needs the theme's per-root value scales plus the arbitrary-value grammar,
// which is closer to reimplementing the utility resolver than to reading a table". Its own comment
// asked for the assertion to be deleted deliberately when that changed.
//
// It changed. `ResolveFunctionalUtilityValue` is that resolver, ported, and #31bbwty gave it a
// description for all 57 closure-registered roots, so `classExistsIn` asks whether the value resolves
// rather than stopping at whether the class parses.
//
// The complement is asserted beside it, since a rule that reports everything would pass the first
// half alone.
func TestKnownRootWithUnknownValueIsReported(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {
		system := designSystem.System

		if classExistsIn("text-huge", system) {
			t.Error("`text-huge` has root `text` and compiles to nothing, so the rule must report it")
		}
		if !classExistsIn("text-lg", system) {
			t.Error("`text-lg` compiles, so reporting it would be a false positive on working code")
		}

		// An unknown root is still caught, which is the half that worked before this change.
		if classExistsIn("txt-huge", system) {
			t.Error("an unknown root must still be reported, or the rule catches nothing")
		}
	})
}

// TestExistenceComesFromTheRepositoryRatherThanATable is the measurement the swap exists for.
//
// It asks two design systems about the same classes and requires them to disagree. A rule reading a
// generated table cannot produce a disagreement at all, which is what makes the disagreement count
// the assertion that fails if the swap is ever reverted.
//
// The second system is built from `independent_theme.css`, which exists precisely because the two
// corpus repositories cannot answer this question: they both vendor `libraries/structure`, so a fact
// they agree on may be a fact about the submodule rather than about Tailwind. Its own file comment
// records that generating against it produced 258 of 301 shared roots differing.
//
// `flex` and `px-4` are asserted known on both. Without them the disagreement count would be
// satisfiable by a rule that had simply stopped answering.
func TestExistenceComesFromTheRepositoryRatherThanATable(t *testing.T) {
	t.Parallel()
	ahra := unknownFixtureLiveSystem(t)
	independent := independentLiveSystem(t)

	testCases := []struct {
		className     string
		onAhra        bool
		onIndependent bool
		why           string
	}{
		{
			className:     "synthetic-static",
			onAhra:        false,
			onIndependent: true,
			why:           "an `@utility` block only the independent stylesheet declares",
		},
		{
			className:     "synthetic-fn-small",
			onAhra:        false,
			onIndependent: true,
			why:           "a functional `@utility` block only the independent stylesheet declares",
		},
		{
			className:     "flex",
			onAhra:        true,
			onIndependent: true,
			why:           "a framework static, known everywhere",
		},
		{
			className:     "px-4",
			onAhra:        true,
			onIndependent: true,
			why:           "a framework functional root, known everywhere",
		},
	}

	disagreements := 0
	for _, testCase := range testCases {
		gotAhra := classExistsIn(testCase.className, ahra)
		gotIndependent := classExistsIn(testCase.className, independent)

		if gotAhra != testCase.onAhra {
			t.Errorf("%s (%s): the ahra design system says exists=%v, want %v",
				testCase.className, testCase.why, gotAhra, testCase.onAhra)
		}
		if gotIndependent != testCase.onIndependent {
			t.Errorf("%s (%s): the independent design system says exists=%v, want %v",
				testCase.className, testCase.why, gotIndependent, testCase.onIndependent)
		}
		if testCase.onAhra != testCase.onIndependent {
			disagreements++
		}
	}

	// The population guard. Every assertion above is satisfiable by a rule that answers false for
	// everything, and a table-reading rule answers identically on both systems by construction. So
	// the disagreement count is asserted directly: it is the one number no table can produce.
	if disagreements < 2 {
		t.Fatalf(
			"only %d of the classes answered differently on the two design systems; a rule reading a "+
				"generated table produces zero disagreements by construction, so this count is the "+
				"evidence that existence is being asked of the repository",
			disagreements,
		)
	}
	t.Logf("%d of %d classes answered differently on the two design systems",
		disagreements, len(testCases))
}

// TestRepositoryNamesInTheFrameworkTablesAreStillVouchedFor pins a gap this task did not close.
//
// The migration fixed one direction of the repository/framework confusion and not the other, and the
// half that remains is in a component this task does not own. Pinned as a failing-shaped assertion
// on the CURRENT behaviour, so closing it upstream breaks this test deliberately rather than leaving
// a comment nobody reads.
//
// # What is fixed
//
// A repository's own `@utility` classes are no longer reported as unknown. `synthetic-static` and
// `synthetic-fn-small` read as known on the design system that declares them, which the shipped
// tables could not do for any repository except the one they were generated from. That is the
// failure the task names and TestExistenceComesFromTheRepositoryRatherThanATable measures.
//
// # What was not fixed here, and now is
//
// `KnownStatics` and `KnownRoots` were enumerated from `designSystem.getClassList()` on a LOADED
// design system, which returns the repository's `@utility` classes alongside the framework's. The
// generator ran against ahra's `theme.css`, so ahra's own tokens sat in tables headed
// `Source: Tailwind 4.3.3`, and `HasUtility` consulted them as its framework fallback.
//
// M9 replaced that fallback with the ported registrations: `FrameworkStaticDeclarations` for statics
// and the union of the wave tables and the descriptor rows for functional roots. The test that used
// to measure the leak said to delete it and assert the corrected behaviour instead, so this is that
// assertion.
func TestRepositoryNamesAreNoLongerVouchedForByTheFramework(t *testing.T) {
	t.Parallel()
	ahra := unknownFixtureLiveSystem(t)
	independent := independentLiveSystem(t)

	// Names ahra declares and the independent system does not, measured rather than listed.
	//
	// `zoom-in` and `zoom-out` are deliberately absent from this sample, and the reason is worth
	// stating because writing it wrong is what surfaced it. Ahra declares both as `@utility` blocks,
	// so a list built from "ahra declares it" includes them. But they also parse as the framework
	// root `zoom` with the value `in`, which every design system registers, so they are correctly
	// known everywhere and flagging them measured the parser rather than the tables.
	//
	// The rest are static names with no functional root behind them, so their only source is a
	// registration, which is exactly what this asserts.
	var declaredByAhra []string
	for _, name := range []string{
		"markdown-content", "typing-dots", "prose", "scrollbar-hide",
		"fade-in", "fade-out",
		"slide-in-from-top", "slide-out-to-left",
	} {
		if ahra.Utilities() == nil || !ahra.Utilities().Has(name) {
			continue
		}
		declaredByAhra = append(declaredByAhra, name)
	}

	if len(declaredByAhra) == 0 {
		t.Fatal("ahra declares none of the sampled utilities, so this measurement compared nothing; " +
			"either the fixture stopped loading the repository's blocks or the names changed")
	}

	for _, name := range declaredByAhra {
		if classExistsIn(name, independent) {
			t.Errorf("%q is declared by ahra and still vouched for on a design system that does not "+
				"declare it, so a repository token is being carried as a framework fact", name)
		}
		if !classExistsIn(name, ahra) {
			t.Errorf("%q is declared by ahra and no longer known there, so the fix removed more than "+
				"the leak", name)
		}
	}
	t.Logf("checked %d repository-declared utilities: known on ahra, unknown on an independent system", len(declaredByAhra))
}

// TestUnknownClassFixturesActuallyRan is what stops this file from going green on nothing.
//
// Every fixture above skips when no installed tailwindcss can be found, and a skip is the honest
// answer for a machine that genuinely has none. The problem is that `go test` prints `ok` for a
// package whose every case skipped, so a mistake in where the walk STARTS looks identical to a
// machine that lacks the package. That was not hypothetical for the class-order fixtures and the
// same helper shape is used here.
//
// So this test fails where the others skip, and it asserts the end-to-end claim rather than only the
// precondition: the fixture stylesheet's own `@utility` block must read as known and a typo must
// still report, in the same program. A design system that built but resolved nothing would fail the
// first; a rule that had stopped answering would fail the second.
func TestUnknownClassFixturesActuallyRan(t *testing.T) {
	t.Parallel()
	packageRoot := unknownFixturePackageRoot()
	if packageRoot == "" {
		t.Fatalf(
			"the vendored tailwindcss at %s is missing, so every fixture in this file would skip and the "+
				"package still report ok; restore it (internal/lint/rules/tailwind/vendored)",
			vendored.TailwindPackageRoot(),
		)
	}

	result := runUnknownFixture(t, "Component.tsx",
		`const element = <div className="synthetic-static flx" />;`)
	rule_testing.ExpectFindings(t, result, "unknownClass")
}

// unknownFixtureLiveSystem builds the corpus repository's design system, for the unit-level tests.
//
// The repository on disk rather than the fixture's temp directory, because these tests ask about
// classes the fixture stylesheet does not declare and the corpus repository does. ahra is private, so
// it skips naming COHERE_CORPUS_AHRA where the corpus is not set, and fails where it is set and the
// stylesheet is missing (#sycrdr6). The stylesheet's tailwindcss is the vendored snapshot, the same
// version ahra installs.
func unknownFixtureLiveSystem(t *testing.T) *tailwindengine.LoadedDesignSystem {
	t.Helper()

	entryPoint := corpus.Ahra.Path(t, "app", "_theme", "styles", "theme.css")
	if _, err := os.Stat(entryPoint); err != nil {
		t.Fatalf("the ahra corpus has no stylesheet at %s, so there is no live design system to ask", entryPoint)
	}
	packageRoot := unknownFixturePackageRoot()

	system, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
		EntryPoint:          entryPoint,
		TailwindPackageRoot: packageRoot,
	})
	if err != nil {
		t.Fatalf("loading the corpus design system: %v", err)
	}
	return system
}

// independentLiveSystem builds a design system that shares nothing with either corpus repository.
//
// The stylesheet is copied beside the corpus repository's `node_modules` rather than read in place,
// because `@import "tailwindcss"` resolves through whichever `node_modules` is reachable from where
// the file sits, and this checkout vendors none. Its own file comment says the same.
func independentLiveSystem(t *testing.T) *tailwindengine.LoadedDesignSystem {
	t.Helper()

	// The theme is this package's own testdata, under its generator, so it is read from the package
	// directory and is never absent. It used to be read from five directories up, outside the repository,
	// where it never was, so every test that asked for it skipped on every machine (#sycrdr6).
	source := filepath.Join("tools", "generate_descriptor_base", "testdata", "independent_theme.css")
	contents, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("reading the independent theme: %v", err)
	}

	packageRoot := unknownFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine")
	}

	// Beside the installed package's own tree, so the bare specifier resolves by the same upward
	// walk a repository uses. Written into a directory this test owns and removed after, rather than
	// into the corpus repository's working tree.
	staging := t.TempDir()
	if err := os.Symlink(packageRoot, filepath.Join(staging, "node_modules_link")); err != nil {
		t.Fatalf("linking the installed tailwindcss beside the independent stylesheet: %v", err)
	}
	modules := filepath.Join(staging, "node_modules")
	if err := os.MkdirAll(modules, 0o755); err != nil {
		t.Fatalf("creating the staging node_modules: %v", err)
	}
	if err := os.Symlink(packageRoot, filepath.Join(modules, "tailwindcss")); err != nil {
		t.Fatalf("linking the installed tailwindcss into the staging directory: %v", err)
	}

	entryPoint := filepath.Join(staging, "independent_theme.css")
	if err := os.WriteFile(entryPoint, contents, 0o644); err != nil {
		t.Fatalf("staging the independent stylesheet: %v", err)
	}

	system, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
		EntryPoint:          entryPoint,
		TailwindPackageRoot: findTailwindPackageRoot(staging, diskFileExists),
	})
	if err != nil {
		t.Fatalf("loading the independent design system: %v", err)
	}
	return system
}

// A repository's own `@utility` root is never reported, whatever its value looks like.
//
// `shadow--3` is the control this rule needs and the one a value check gets wrong most easily. It
// reads as root `shadow-`, declared by `@utility shadow--*` with a `--shadow---3` through
// `--shadow--9` scale behind it, and no framework description covers that shape. Asking a framework
// description about it would report a class this repository writes and the engine compiles.
//
// `content--0`, `border--0` and `background--0` are the same family, the trailing-dash roots
// `4f23e9f` and `7a72125` found sitting in generated tables headed with a Tailwind version.
//
// Read from the design system rather than listed, so a repository adding an `@utility` block joins
// this test rather than needing to be added to it.
func TestRepositoryUtilityRootsAreNeverReported(t *testing.T) {
	t.Parallel()
	system := unknownFixtureLiveSystem(t)
	if system == nil {
		t.Skip("no design system loaded")
	}

	var checked int
	for _, className := range []string{"shadow--3", "content--0", "border--0", "background--0"} {
		candidates := tailwindengine.ParseCandidate(className, system)
		if len(candidates) == 0 {
			continue
		}
		if !system.DeclaresFunctionalUtility(candidates[0].Root) {
			continue
		}
		checked++
		if !classExistsIn(className, system) {
			t.Errorf("%s reads as the repository root %q and is reported, which is a false positive on a declared utility",
				className, candidates[0].Root)
		}
	}

	t.Logf("repository utility classes checked: %d", checked)
	if checked == 0 {
		t.Skip("this design system declares none of these roots")
	}
}
