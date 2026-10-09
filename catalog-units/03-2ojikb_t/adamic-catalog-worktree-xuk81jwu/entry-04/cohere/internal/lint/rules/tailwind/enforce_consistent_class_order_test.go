package tailwind

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/tailwind/vendored"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// Expectations measured by diffing the rule against the engine's own sort over the whole corpus,
// not by reading upstream: upstream delegates the entire question to `getClassOrder` in three
// lines, so its source says nothing about how the order is built.
//
// # Why every fixture here runs through a program
//
// The rule reads the live design system, so it declares `ProgramReads` and takes `ctx.Program` to
// find the repository's stylesheet. `rule_testing.Run` hands a rule a nil Program, and the rule's own
// decline path returns nil listeners for that case rather than reporting. So every fixture in this
// file that used `rule_testing.Run` after the swap would have exercised the nil-Program branch: the
// reporting half would have failed loudly, and the whole silent half would have passed while
// proving nothing at all.
//
// That is the vacuous-probe shape `rule_testing.RunTyped`'s own comment describes one level up, and it
// is worth naming because it is the failure that hides: a suite going green on a rule that never
// ran. So the fixtures run through `RunTypedFiles`, which writes a real tsconfig and a real
// stylesheet into a temp directory, and the rule finds that stylesheet through the same entry-point
// search it uses on a real repository.
//
// The fixture stylesheet declares no tokens of its own, deliberately. A fixture theme that added
// tokens would test the fixture rather than the framework, and per-repository behaviour is what the
// differential in `internal/tailwind` measures against two real design systems.
//
// # The `@import` has to reach a real package, and a skip is the only honest alternative
//
// `findTailwindPackageRoot` walks upward from the stylesheet looking for `node_modules/tailwindcss`.
// A fixture written into `t.TempDir()` sits under the system temp directory, where that walk finds
// nothing, so `@import "tailwindcss"` cannot resolve and the rule declines.
//
// That decline is the rule behaving correctly, and it is exactly why it had to be fixed rather than
// accommodated: every fixture in the silent half would have passed on a `designSystemUnavailable`
// finding instead of on the order being right. Measured before the fix, this file reported
// `designSystemUnavailable` on 11 of the silent cases and on both message fixtures.
//
// So the import is pointed at the installed package by absolute path. When there is no installed
// tailwindcss to point at, the fixtures skip: a suite that cannot build a design system has not
// measured this rule, and saying "not measured" is the difference between an honest gap and a green
// run over nothing.
//
// # A skip is honest and it is still not free, so it is loud
//
// `cohere` vendors no `node_modules` of its own, so the walk has to start somewhere that has one. It
// starts at the corpus repository, which is where the class-order fixtures in `internal/tailwind`
// were captured from and the only installed 4.3.3 on this machine.
//
// The first version of this helper walked upward from `.`, the package directory. That finds nothing,
// so every fixture in this file skipped, and `go test` reports a file whose every case skipped as
// `ok`. Measured: 7 of 7 reporting cases and 13 of 13 silent ones skipped while the package printed
// PASS. A skip that reads as a pass is the same failure as silence that reads as agreement, one layer
// out, so `TestClassOrderFixturesActuallyRan` below fails rather than skips when the package cannot
// be found, and that test is the one that makes the rest of this file's greenness mean something.
//
// # The package is symlinked into the fixture rather than imported by absolute path
//
// Pointing the fixture's `@import` at an absolute path does not work, and the reason is worth stating
// because it looks like it should. The rule does not resolve the import itself: it calls
// `findTailwindPackageRoot`, which walks UP from the stylesheet's own directory looking for
// `node_modules/tailwindcss`, and hands the result to `LoadDesignSystem` as the resolver's root. A
// fixture in `t.TempDir()` has nothing above it, so that walk fails and the rule declines before it
// ever reads what the stylesheet imports. Measured: every reporting fixture came back
// `designSystemUnavailable` while the stylesheet held a perfectly good absolute import.
//
// So the fixture gets a real `node_modules/tailwindcss`, symlinked to the installed one. That makes
// the rule take exactly the path it takes on a repository — its own upward walk, its own resolver,
// the bare `tailwindcss` specifier — rather than a path arranged for the test.
const classOrderFixtureStylesheetPath = "app/_theme/styles/theme.css"

// classOrderFixturePackageRoot is the installed tailwindcss the fixture stylesheets import: the vendored
// snapshot of 4.3.3, the version the class-order corpus in collapse/testdata was captured against, so the
// fixtures and the corpus agree on which engine they mean, on any machine (#sycrdr6).
func classOrderFixturePackageRoot() string {
	return vendored.TailwindPackageRoot()
}

// classOrderFixtureStylesheet is the fixture's root stylesheet.
//
// The bare specifier a real repository writes, resolved through the symlink `runClassOrderFixture`
// plants. Nothing here is arranged for the test beyond the symlink itself.
const classOrderFixtureStylesheet = `@import "tailwindcss";`

// runClassOrderFixture runs the rule against a one-file program that has a real design system.
//
// Every caller passes source for one file, so the stylesheet is added here rather than at each call
// site: a fixture that forgot it would not fail, it would decline, and a declining rule reports
// nothing on the silent half.
//
// The symlink is planted before the program is built, into the same temp directory
// `rule_testing.RunTypedFiles` writes the fixture files to, so the rule's own upward walk finds it.
func runClassOrderFixture(t *testing.T, fileName string, source string) rule_testing.Result {
	t.Helper()

	packageRoot := classOrderFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine, so the rule's design system cannot be " +
			"built and these fixtures would measure a decline rather than an order")
	}

	return rule_testing.RunTypedFilesWithSetup(t, EnforceConsistentClassOrder, map[string]string{
		fileName:                        source,
		classOrderFixtureStylesheetPath: classOrderFixtureStylesheet,
	}, fileName, func(root string) {
		modules := filepath.Join(root, "node_modules")
		if err := os.MkdirAll(modules, 0o755); err != nil {
			t.Fatalf("creating the fixture node_modules: %v", err)
		}
		if err := os.Symlink(packageRoot, filepath.Join(modules, "tailwindcss")); err != nil {
			t.Fatalf("linking the installed tailwindcss into the fixture: %v", err)
		}
	})
}

func TestEnforceConsistentClassOrderReportsMisordering(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
		wantIds  []string
	}{
		{
			name:     "two classes reversed",
			fileName: "Component.tsx",
			source:   `const element = <div className="items-center flex" />;`,
			wantIds:  []string{"inconsistentClassOrder"},
		},
		{
			// A variant class before its unprefixed counterpart. The engine groups by variant first,
			// so every bare class precedes every prefixed one.
			name:     "variant before base",
			fileName: "Component.tsx",
			source:   `const element = <div className="hover:px-8 px-4" />;`,
			wantIds:  []string{"inconsistentClassOrder"},
		},
		{
			// `group` is unranked and sorts first, so writing it last is the violation. The reverse
			// of this pair was asserted until the ecosystem convention was checked: Tailwind's
			// Prettier plugin and `better-tailwindcss` both hoist the markers.
			name:     "unranked class written last",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex group" />;`,
			wantIds:  []string{"inconsistentClassOrder"},
		},
		{
			// Two unranked classes keep the order they were written in. Alphabetical would rewrite
			// this one, and `better-tailwindcss` accepts it as written.
			name:     "two unranked classes in non-alphabetical source order",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex peer group" />;`,
			wantIds:  []string{"inconsistentClassOrder"},
		},
		{
			// Every class the engine ranks null leads, not only the markers. This literal was declined
			// until #vf1hd6j measured the plugin hoisting classes like `text-dark`.
			name:     "unknown class written last",
			fileName: "Component.tsx",
			source:   `const element = <div className="items-center flex ahralia-splash" />;`,
			wantIds:  []string{"inconsistentClassOrder"},
		},
		{
			name:     "several classes out of order",
			fileName: "Component.tsx",
			source:   `const element = <div className="gap-2 items-center flex" />;`,
			wantIds:  []string{"inconsistentClassOrder"},
		},
		{
			name:     "on a callee surface",
			fileName: "Component.tsx",
			source:   `const merged = cn('items-center flex');`,
			wantIds:  []string{"inconsistentClassOrder"},
		},
		{
			name:     "on a variable surface",
			fileName: "Styles.ts",
			source:   `const className = 'items-center flex';`,
			wantIds:  []string{"inconsistentClassOrder"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runClassOrderFixture(t, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The silent half. Several of these are shapes an earlier version of the comparator got wrong, and
// each one is a case where reporting would be reporting correct code.
func TestEnforceConsistentClassOrderStaysSilent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			name:     "already ordered",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex items-center gap-2" />;`,
		},
		{
			name:     "base before variant",
			fileName: "Component.tsx",
			source:   `const element = <div className="px-4 hover:px-8" />;`,
		},
		{
			// Variants group together, and within a group the classes order among themselves. An
			// earlier comparator interleaved them with the bare classes and agreed with the engine on
			// only 51% of the corpus.
			name:     "variant group kept together",
			fileName: "Component.tsx",
			source:   `const element = <div className="pointer-events-none hidden md:absolute md:inset-x-0 md:flex md:w-full" />;`,
		},
		{
			name:     "unranked class first",
			fileName: "Component.tsx",
			source:   `const element = <div className="group flex" />;`,
		},
		{
			// A root whose own prefix is also a root. `ring-offset` exists in the ordering table
			// and not in the conflict table, so resolving it through the latter answered with
			// `ring`'s properties and made `ring-offset-1` indistinguishable from `ring-1`. The
			// engine accepts this literal as written; the rule reported it until the lookup was
			// pointed at the table it reads.
			name:     "ring-offset resolves to its own root, not to ring",
			fileName: "Component.tsx",
			source: `const element = <div className="focus-visible:ring-1 focus-visible:ring-(--color-content-informative) ` +
				`focus-visible:ring-offset-1 focus-visible:outline-none" />;`,
		},
		{
			// Both spellings are accepted, which is what pins the tiebreak to source order rather
			// than to any ordering of our own. An alphabetical tiebreak passes the first and fails
			// the second, so the pair is the discriminator.
			name:     "two unranked classes, alphabetical source order",
			fileName: "Component.tsx",
			source:   `const element = <div className="group peer flex" />;`,
		},
		{
			name:     "two unranked classes, reverse-alphabetical source order",
			fileName: "Component.tsx",
			source:   `const element = <div className="peer group flex" />;`,
		},
		{
			// `font-mono` and `text-[10px]` share no root and their relative order is not guessable
			// from their names; it comes from the table.
			name:     "ordering the names do not suggest",
			fileName: "Component.tsx",
			source:   `const element = <div className="truncate font-mono text-[10px]" />;`,
		},
		{
			name:     "single class",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex" />;`,
		},
		{
			// A duplicate is `no-duplicate-classes`'s finding. Reporting an ordering defect here
			// would send the author to fix the wrong thing.
			name:     "duplicate class present",
			fileName: "Component.tsx",
			source:   `const element = <div className="items-center flex flex" />;`,
		},
		{
			// A class the engine ranks null leads, in source order, so this is already in order. It
			// is `no-unknown-classes` that has something to say about `ahralia-splash`.
			name:     "unknown class written first",
			fileName: "Component.tsx",
			source:   `const element = <div className="ahralia-splash flex items-center" />;`,
		},
		{
			name:     "unrelated attribute",
			fileName: "Component.tsx",
			source:   `const element = <div title="items-center flex" />;`,
		},
		{
			name:     "unrelated callee",
			fileName: "Component.tsx",
			source:   `const value = someOtherFunction('items-center flex');`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runClassOrderFixture(t, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestClassOrderMessageNamesTheOrder checks the part an author acts on.
//
// The order is not guessable from the class names, which is the whole reason the rule needs a table.
// A finding that says only "these are misordered" leaves the reader to run the formatter and diff
// the result.
func TestClassOrderMessageNamesTheOrder(t *testing.T) {
	t.Parallel()
	result := runClassOrderFixture(t, "Component.tsx",
		`const element = <div className="gap-2 items-center flex" />;`)

	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	description := result.Diagnostics[0].Message.Description
	if !strings.Contains(description, "\"flex items-center gap-2\"") {
		t.Errorf("message does not name the correct order: %s", description)
	}
}

// TestClassOrderFixMatchesThePlugin pins the rewrite, which is what Prettier's Tailwind plugin wrote
// for every one of these lists until it left.
//
// The rule shipped without a fix, on the reasoning that a rewrite would reflow wrapped lists. The
// plugin never reflowed them: it permutes the classes between the same whitespace runs, and so does
// this fix, which writes only the slots whose class changes. The wrapped and padded cases below are
// that property, and the padding staying put is `no-unnecessary-whitespace`'s to remove.
func TestClassOrderFixMatchesThePlugin(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name, source, want string
	}{
		{
			name:   "two classes reversed",
			source: `const element = <div className="items-center flex" />;`,
			want:   `const element = <div className="flex items-center" />;`,
		},
		{
			name:   "several classes, only the moved slots rewritten",
			source: `const element = <div className="gap-2 items-center flex" />;`,
			want:   `const element = <div className="flex items-center gap-2" />;`,
		},
		{
			name:   "a null leads in source order",
			source: `const element = <div className="items-center peer flex ahralia-splash" />;`,
			want:   `const element = <div className="peer ahralia-splash flex items-center" />;`,
		},
		{
			name:   "a wrapped list stays wrapped",
			source: "const element = <div\n  className=\"items-center\n    flex\n    gap-2\"\n/>;",
			want:   "const element = <div\n  className=\"flex\n    items-center\n    gap-2\"\n/>;",
		},
		{
			name:   "padding is left for the whitespace rule",
			source: `const element = <div className="  items-center   flex " />;`,
			want:   `const element = <div className="  flex   items-center " />;`,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runClassOrderFixture(t, "Component.tsx", testCase.source)
			rule_testing.ExpectFixedSource(t, result, testCase.want+"\n")
		})
	}
}

// TestClassOrderReportsWithoutAFixItCannotPlace covers the cases that keep the finding and drop the
// fix: rewriting at source offsets needs the source to be the value, and a class the deprecation rule
// renames is renamed before it is moved, since both edits would claim its bytes.
func TestClassOrderReportsWithoutAFixItCannotPlace(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, source string
	}{
		{
			name:   "an escape in the literal",
			source: `const merged = cn('items-center\u0020flex');`,
		},
		{
			name:   "a deprecated class that moves",
			source: `const element = <div className="items-center flex-grow" />;`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runClassOrderFixture(t, "Component.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, "inconsistentClassOrder")
			if len(result.Diagnostics[0].Fixes) != 0 {
				t.Fatalf("expected no fix, got %+v", result.Diagnostics[0].Fixes)
			}
		})
	}
}

// The dimension tests below replace five that read the deleted generated tables directly.
//
// Each one asserted a real property and asserted it through `classSortsBefore`, `compareVariants`,
// `declaredPropertiesForOrdering` or `orderingRootOf`, all of which were the pairwise-comparator path
// that #4q5dsn3 removed. The properties survive; what changed is that they are now asked of the live
// design system and their expected answers are quoted from the engine rather than from the port.
//
// Two of them asserted something that is simply false, and those are corrected rather than
// translated. See TestClassOrderDepthIsNotADimension.

// TestClassOrderDimensionsAreAllUsed is the known-dirty control, on the live path.
//
// Replaces TestComparatorDimensionsAreAllUsed. Three simpler orderings each looked right and each was
// measured wrong against the engine: ordering by root alone agreed on 51% of the corpus, adding
// variants took it to 58% while representatives still carried their own prefixes, and unprefixing
// them took it to 91%. The last 9% was the markers being given a position instead of being separated
// out entirely.
//
// So each dimension is exercised by a pair only it can order correctly. The pairs are unchanged from
// the test this replaces; the mechanism asked is the live sort rather than the pairwise comparator.
func TestClassOrderDimensionsAreAllUsed(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {

		testCases := []struct {
			name      string
			first     string
			second    string
			dimension string
		}{
			{
				name:      "class position",
				first:     "flex",
				second:    "items-center",
				dimension: "the design system's own readings; their roots share no ordering the names suggest",
			},
			{
				name:      "variant grouping",
				first:     "px-4",
				second:    "hover:px-8",
				dimension: "the variant mask; without it the two interleave by property alone",
			},
			{
				name:      "markers first",
				first:     "group",
				second:    "flex",
				dimension: "the marker partition; `group` has no reading and must not be sorted by one",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()
				forward := classOrderLiveSort(t, designSystem, []string{testCase.first, testCase.second})
				reverse := classOrderLiveSort(t, designSystem, []string{testCase.second, testCase.first})
				want := []string{testCase.first, testCase.second}

				if strings.Join(forward, " ") != strings.Join(want, " ") {
					t.Errorf("got %v, expected %v, decided by %s", forward, want, testCase.dimension)
				}
				// Sorted from the other input too, because a sort that merely preserves its input would
				// pass the first assertion and order nothing.
				if strings.Join(reverse, " ") != strings.Join(want, " ") {
					t.Errorf("from reversed input got %v, expected %v; the sort is not deciding, it is "+
						"preserving", reverse, want)
				}
			})
		}
	})
}

// TestClassOrderVariantDimensions pins the variant dimensions, with the engine's answers.
//
// Replaces TestVariantComparisonDimensions. Three of its five cases asserted properties that hold and
// are kept; the expected orders here are quoted from `variant_fixtures.json`, which recorded what
// `getClassOrder` returned, rather than restated from the port.
//
// The two dropped cases are the subject of TestClassOrderDepthIsNotADimension below.
func TestClassOrderVariantDimensions(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {

		testCases := []struct {
			name     string
			input    []string
			expected []string
			fixture  string
			why      string
		}{
			{
				name:     "compound resolves on its inner variant",
				input:    []string{"dark:focus:flex", "dark:placeholder:flex"},
				expected: []string{"dark:placeholder:flex", "dark:focus:flex"},
				fixture:  "class-order/compound-inner-segment",
				why:      "both are `dark:` compounds, so only the inner variant separates them",
			},
			{
				name:     "a named group sorts where its unnamed form does",
				input:    []string{"group-hover/pdf:flex", "group-hover/csv:flex", "group-hover:flex"},
				expected: []string{"group-hover:flex", "group-hover/csv:flex", "group-hover/pdf:flex"},
				fixture:  "class-order/named-groups",
				why:      "the name is a modifier on the variant, and an unnamed compound precedes a named one",
			},
			{
				name:     "unknown functional variants order by their own value",
				input:    []string{"data-[show=true]:flex", "data-[show=false]:flex", "hover:flex"},
				expected: []string{"hover:flex", "data-[show=false]:flex", "data-[show=true]:flex"},
				fixture:  "class-order/functional-data",
				why:      "two `data-` variants tie on registration order and separate on their value text",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()
				ordered := classOrderLiveSort(t, designSystem, testCase.input)
				if strings.Join(ordered, " ") != strings.Join(testCase.expected, " ") {
					t.Errorf("got %v, %s says %v: %s",
						ordered, testCase.fixture, testCase.expected, testCase.why)
				}
			})
		}
	})
}

// TestClassOrderDepthIsNotADimension is the correction, and it is the point of this task.
//
// Two cases in the test this replaces asserted the opposite of what the engine does, and they passed
// because the code under test asserted the same wrong thing. Quoted from the deleted
// TestVariantComparisonDimensions:
//
//	{ before: "group-hover:",  after: "group-hover:disabled:",
//	  why: "a stacked variant narrows an already-narrowed selector and lands in a later layer" }
//	{ before: "disabled:",     after: "group-hover:disabled:",
//	  why: "every single variant precedes every stacked one, whatever it starts with" }
//
// The second is false as stated, and the first is true only by coincidence. The engine ORs one bit
// per variant into a mask and compares masks numerically, so a class's position is decided by its
// HIGHEST-ranked variant. A stacked variant sorts first whenever the class it is compared against
// carries a higher bit.
//
// The evidence is the engine's own, from `class-order/stacked-against-single` and
// `class-order/stacked-between-singles` in variant_fixtures.json: `group-hover:disabled:flex` (mask 5)
// precedes `dark:flex` (mask 8), and in a four-class list it lands third rather than last. That is
// the same divergence `#nr3wtj1` filed and `TestDepthFirstDisagreesWithTheEngineOnTheCorpus` measured
// at 15 pairs of 4,265.
//
// The first case is kept, with its reason corrected: `group-hover:` does precede
// `group-hover:disabled:`, not because it is shorter but because the two share the `group-hover` bit
// and the stacked one carries `disabled` on top of it, so its mask is strictly larger.
func TestClassOrderDepthIsNotADimension(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {

		t.Run("a superset mask sorts after the subset it extends", func(t *testing.T) {
			t.Parallel()
			ordered := classOrderLiveSort(t, designSystem,
				[]string{"group-hover:disabled:flex", "group-hover:flex"})
			expected := []string{"group-hover:flex", "group-hover:disabled:flex"}
			if strings.Join(ordered, " ") != strings.Join(expected, " ") {
				t.Errorf("got %v, expected %v: both carry the `group-hover` bit and the stacked class "+
					"carries `disabled` on top of it, so its mask is strictly larger", ordered, expected)
			}
		})

		t.Run("but a stacked variant precedes a higher-ranked single one", func(t *testing.T) {
			t.Parallel()
			ordered := classOrderLiveSort(t, designSystem,
				[]string{"dark:flex", "group-hover:disabled:flex"})
			expected := []string{"group-hover:disabled:flex", "dark:flex"}
			if strings.Join(ordered, " ") != strings.Join(expected, " ") {
				t.Errorf("got %v, class-order/stacked-against-single says %v; the deleted rule asserted "+
					"'every single variant precedes every stacked one' and the engine does no such thing",
					ordered, expected)
			}
		})
	})
}

// TestClassOrderReadingsKeepCustomProperties pins that `--tw-*` properties stay in a reading.
//
// Replaces TestOrderingPropertiesKeepCustomProperties, and the assertion is corrected rather than
// translated, because the original was right about the property and wrong about where it shows.
//
// The original asserted that `shadow-lg` and `ring-1` "lead with the same property" is the defect.
// The engine says otherwise. From classorder_fixtures.json, the readings it captured on this
// repository:
//
//	shadow-lg   {order: [315, 316], count: 2}
//	ring-1      {order: [315, 318], count: 2}
//
// They SHARE their leading index and separate on the second. Index 315 is the `--tw-*` custom
// property they both emit; 316 and 318 are their differing second declarations. So the property that
// matters is that the reading holds more than one index at all: a lookup that dropped custom
// properties, the way the conflict tables do, would leave both as a single `[box-shadow]` index and
// they would tie completely. Ten real class lists came out wrong when a previous version read the
// conflict tables here.
//
// Asserting on `order[0]` differing, as the original did, is an assertion the engine fails. It passed
// only because the deleted `declaredPropertiesForOrdering` returned property NAMES from a different
// table with a different shape, so the two tests were never asking the same question.
func TestClassOrderReadingsKeepCustomProperties(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {

		keys, unplaceable, resolved := classOrderKeys(
			[]string{"shadow-lg", "ring-1"}, designSystem.System, designSystem.Table)
		if !resolved {
			t.Fatalf("could not place %q", unplaceable)
		}

		shadow, ring := keys["shadow-lg"], keys["ring-1"]
		// Both indices, because one index each is what dropping the custom property would leave.
		if len(shadow.order) < 2 || len(ring.order) < 2 {
			t.Fatalf("expected both to declare a custom property and box-shadow, got %v and %v; a single "+
				"index each means the `--tw-*` property was dropped and the two cannot be separated",
				shadow.order, ring.order)
		}
		// And they must still differ somewhere, which for these two is the second index.
		if shadow.order[0] == ring.order[0] && shadow.order[1] == ring.order[1] {
			t.Errorf("shadow-lg %v and ring-1 %v are identical readings, so the sort cannot separate "+
				"them and the engine does", shadow.order, ring.order)
		}
		// The engine's own numbers, quoted so a drift in either reading fails here rather than silently
		// changing every class list that holds a shadow and a ring.
		if shadow.order[0] != 315 || shadow.order[1] != 316 {
			t.Errorf("shadow-lg reads %v; classorder_fixtures.json recorded [315 316]", shadow.order)
		}
		if ring.order[0] != 315 || ring.order[1] != 318 {
			t.Errorf("ring-1 reads %v; classorder_fixtures.json recorded [315 318]", ring.order)
		}
	})
}

// TestClassOrderRootsResolveAgainstTheDesignSystem is the known-dirty control for the table-choice bug.
//
// Replaces TestOrderingRootsResolveInTheOrderingTable. That bug was a lookup resolving an ordering
// question through the conflict table, where `ring-offset` does not exist, so the longest-wins walk
// found `ring` and answered with ring's properties; `ring-offset-1` then read identically to `ring-1`
// and two classes the engine separates collapsed onto one key.
//
// The live path cannot make that mistake in the same way, because `ParseCandidate` resolves the root
// against the design system's own registrations rather than by walking a table of names. The property
// is still worth pinning, and it is pinned the same way: the two classes must not share a reading.
func TestClassOrderRootsResolveAgainstTheDesignSystem(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {

		keys, unplaceable, resolved := classOrderKeys(
			[]string{"ring-offset-1", "ring-1"}, designSystem.System, designSystem.Table)
		if !resolved {
			t.Fatalf("could not place %q", unplaceable)
		}

		offset, ring := keys["ring-offset-1"], keys["ring-1"]
		if len(offset.order) == 0 || len(ring.order) == 0 {
			t.Fatalf("both should declare properties, got %v and %v", offset.order, ring.order)
		}
		if offset.order[0] == ring.order[0] {
			t.Errorf("ring-offset-1 and ring-1 lead with the same property index %d, so the sort cannot "+
				"separate them and the engine does", offset.order[0])
		}
	})
}

// TestClassOrderMarkersLeadRegardlessOfVariant covers the interaction that produced the final three
// corpus divergences.
//
// Replaces TestUnrankedClassesSortFirstRegardlessOfVariant. `group` has no reading and a
// variant-prefixed class has a mask, so an ordering that consulted the mask before the marker
// partition would put `group` in the middle of the list. It belongs at the front, ahead of both.
func TestClassOrderMarkersLeadRegardlessOfVariant(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {

		ordered := classOrderLiveSort(t, designSystem, []string{"hover:px-8", "group", "flex"})
		if ordered[0] != "group" {
			t.Errorf("got %v: a marker must lead even when the other classes carry variants, which "+
				"means the partition has to happen before the mask is consulted", ordered)
		}
	})
}

// TestClassOrderMarkersKeepSourceOrder pins the tiebreak.
//
// Replaces TestUnrankedClassesKeepSourceOrder. The sort must express no preference between two
// markers in either direction, so both `peer group` and `group peer` survive: `better-tailwindcss`
// accepts both, and an alphabetical tiebreak would rewrite the first.
func TestClassOrderMarkersKeepSourceOrder(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {

		for _, input := range [][]string{{"group", "peer", "flex"}, {"peer", "group", "flex"}} {
			ordered := classOrderLiveSort(t, designSystem, input)
			if strings.Join(ordered, " ") != strings.Join(input, " ") {
				t.Errorf("got %v from %v: two markers must keep the order they were written in, and "+
					"alphabetical would rewrite `peer group`", ordered, input)
			}
		}
	})
}

// classOrderLiveSort runs the rule's own ordering over a class list.
//
// The same two calls `reportClassOrder` makes, in the same order, so a test cannot pass against a
// path the rule does not take. Fails rather than skips on an unplaceable class: every class these
// tests use is one this repository's design system knows, so a decline is a defect in the port and
// not a property of the input.
func classOrderLiveSort(t *testing.T, designSystem DesignSystemResult, classes []string) []string {
	t.Helper()

	unranked, placeable := partitionUnranked(classes, designSystem)
	keys, unplaceable, resolved := classOrderKeys(placeable, designSystem.System, designSystem.Table)
	if !resolved {
		t.Fatalf("the rule could not place %q out of %v, so this case proves nothing",
			unplaceable, classes)
	}
	return append(unranked, sortClassesByKey(placeable, keys)...)
}

// TestClassOrderFixturesActuallyRan is what stops this file from going green on nothing.
//
// Every fixture above skips when no installed tailwindcss can be found, and a skip is the honest
// answer for a machine that genuinely has none. The problem is that `go test` prints `ok` for a
// package whose every case skipped, so a mistake in where the walk STARTS looks identical to a
// machine that lacks the package.
//
// That was not hypothetical. The first version of the helper walked upward from `.`, which is this
// package's own directory inside a repository that vendors no `node_modules`. Every fixture skipped,
// the package reported PASS, and the swap under test was never exercised by a single one of them.
//
// So this test fails where the others skip. It asserts the package is findable and that the rule can
// actually build a design system from the fixture stylesheet, which together are the precondition
// every other fixture in this file silently depends on.
func TestClassOrderFixturesActuallyRan(t *testing.T) {
	t.Parallel()
	packageRoot := classOrderFixturePackageRoot()
	if packageRoot == "" {
		t.Fatalf(
			"the vendored tailwindcss at %s is missing, so every fixture in this file would skip and the "+
				"package still report ok; restore it (internal/lint/rules/tailwind/vendored)",
			vendored.TailwindPackageRoot(),
		)
	}

	// And the stylesheet the fixtures write must actually produce findings, which is the end-to-end
	// version of the same claim: a design system that builds but resolves nothing would let the
	// silent half pass while the reporting half failed.
	result := runClassOrderFixture(t, "Component.tsx",
		`const element = <div className="items-center flex" />;`)
	rule_testing.ExpectFindings(t, result, "inconsistentClassOrder")
}

// TestClassOrderSortsTemplateRunsLikeThePlugin pins the template half, each case quoted from what
// Prettier's Tailwind plugin wrote for the same input on 2026-10-02 (ahra's pinned 0.8.1).
//
// Each static run sorts on its own, and a class glued to a hole stays put: `px-` is half of a class
// the hole completes. Whitespace is left exactly where it was, since collapsing it is
// `no-unnecessary-whitespace`'s, so the cases with padding show the order alone.
func TestClassOrderSortsTemplateRunsLikeThePlugin(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name, source, want string
	}{
		{
			name:   "a run after a glued hole keeps its glued class",
			source: "const element = <div className={`px-${size} items-center flex`} />;",
			want:   "const element = <div className={`px-${size} flex items-center`} />;",
		},
		{
			name:   "a run before a glued hole keeps its glued class",
			source: "const element = <div className={`items-center flex px-${size}`} />;",
			want:   "const element = <div className={`flex items-center px-${size}`} />;",
		},
		{
			name:   "a class glued after a hole stays first",
			source: "const element = <div className={`${x}items-center flex block`} />;",
			want:   "const element = <div className={`${x}items-center block flex`} />;",
		},
		{
			name:   "both runs, separators untouched",
			source: "const merged = cn(`items-center   flex ${x} gap-2   block`);",
			want:   "const merged = cn(`flex   items-center ${x} block   gap-2`);",
		},
		{
			name:   "a template in a conditional's branch",
			source: "const element = <div className={open ? `items-center flex ${x}` : 'flex'} />;",
			want:   "const element = <div className={open ? `flex items-center ${x}` : 'flex'} />;",
		},
		{
			name:   "a template inside another template's hole",
			source: "const element = <div className={`flex ${a ? `items-center block ${b}` : ''}`} />;",
			want:   "const element = <div className={`flex ${a ? `block items-center ${b}` : ''}`} />;",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runClassOrderFixture(t, "Component.tsx", testCase.source)
			rule_testing.ExpectFixedSource(t, result, testCase.want+"\n")
		})
	}
}

// The template runs the plugin leaves alone, and so must this rule.
func TestClassOrderLeavesTemplateRunsThePluginLeaves(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, source string
	}{
		{
			// `flex-` is glued to `${grow}`, so `items-center` is the only sortable class before it,
			// and `block flex` is already in order. The plugin returned this unchanged.
			name:   "glued classes leave one sortable class",
			source: "const element = <div className={`items-center flex-${grow} block flex ${x}`} />;",
		},
		{
			name:   "already ordered on both sides of a hole",
			source: "const element = <div className={`flex items-center ${x} block gap-2`} />;",
		},
		{
			// The repeat is no-duplicate-classes' finding and fix; the order lands the pass after, as
			// TestTailwindFixersComposeInOnePass shows for a template whose runs hold repeats.
			name:   "a run holding a repeat",
			source: "const element = <div className={`items-center flex flex ${x}`} />;",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runClassOrderFixture(t, "Component.tsx", testCase.source))
		})
	}
}
