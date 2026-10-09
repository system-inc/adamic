package tailwind

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/tailwind/vendored"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The fixture corpus is the one the migration earned, and every entry in it was a real finding at
// some point. Several are the specific finding a cheaper approach silently lost.
//
// # Why every fixture here runs through a program
//
// Same reason `enforce_consistent_class_order_test.go` and `no_unknown_classes_test.go` give. The
// rule now reads the live design system to split a class into root and value, so it declares
// `ProgramReads` and returns nil listeners when the Program is nil. `rule_testing.Run` hands it exactly
// that, so a fixture left on it would exercise the nil-Program branch: the reporting half fails
// loudly and the silent half passes while proving nothing.
//
// `TestCanonicalFixturesActuallyRan` is what stops the converted file from going green on nothing in
// the other direction, where the walk for an installed tailwindcss finds none and every fixture
// skips while `go test` prints ok.

// runCanonicalFixture runs the rule against a one-file program that has a real design system.
//
// The fixture stylesheet declares no tokens of its own. Unlike `no-unknown-classes`, whose whole
// subject is what a repository adds, this rule's subject is the collapse families, and those are
// framework facts: 44 on ahra and 44 on www-connected-app, with none on either side alone.
func runCanonicalFixture(t *testing.T, fileName string, source string) rule_testing.Result {
	t.Helper()
	return runCanonicalFixtureWithOptions(t, fileName, source, nil)
}

// runCanonicalFixtureWithOptions is runCanonicalFixture for the cases that configure the ignore list.
func runCanonicalFixtureWithOptions(
	t *testing.T,
	fileName string,
	source string,
	options any,
) rule_testing.Result {
	t.Helper()

	packageRoot := unknownFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine, so the rule's design system cannot be " +
			"built and these fixtures would measure a decline rather than a collapse")
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
		return rule_testing.RunTypedFilesWithSetup(t, EnforceCanonicalClasses, files, fileName, plantPackage)
	}
	return rule_testing.RunTypedFilesWithSetupAndOptions(
		t, EnforceCanonicalClasses, files, fileName, options, plantPackage)
}

func TestEnforceCanonicalClassesReportsCollapses(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
		wantIds  []string
	}{
		{
			name:     "padding axes",
			fileName: "Component.tsx",
			source:   `const element = <div className="px-4 py-4" />;`,
			wantIds:  []string{"canonicalCollapse"},
		},
		{
			// Different CSS properties merging into a shorthand. Filtering pairs by declared property
			// family measured fifteen times faster and silently stopped reporting this.
			name:     "width and height into size",
			fileName: "Component.tsx",
			source:   `const element = <div className="w-8 h-8" />;`,
			wantIds:  []string{"canonicalCollapse"},
		},
		{
			// Roots with no value at all. A dash-splitter reads `border-l` as root `border` value
			// `l` and stops reporting this.
			name:     "border edges",
			fileName: "Component.tsx",
			source:   `const element = <div className="border-l border-r" />;`,
			wantIds:  []string{"canonicalCollapse"},
		},
		{
			// A named scale rather than a numeric one, which the generator's first probe could not
			// reach at all.
			name:     "rounded corners",
			fileName: "Component.tsx",
			source:   `const element = <div className="rounded-tl-md rounded-tr-md" />;`,
			wantIds:  []string{"canonicalCollapse"},
		},
		{
			name:     "gap axes",
			fileName: "Component.tsx",
			source:   `const element = <div className="gap-x-2 gap-y-2" />;`,
			wantIds:  []string{"canonicalCollapse"},
		},
		{
			// Variants match on both sides, so the merge is allowed and carries the variant through.
			name:     "collapse under a shared variant",
			fileName: "Component.tsx",
			source:   `const element = <div className="sm:px-4 sm:py-4" />;`,
			wantIds:  []string{"canonicalCollapse"},
		},
		{
			// Negative roots are their own roots rather than a sign flag on a positive one.
			name:     "negative margins",
			fileName: "Component.tsx",
			source:   `const element = <div className="-mt-1 -mb-1" />;`,
			wantIds:  []string{"canonicalCollapse"},
		},
		{
			// Arbitrary values merge like named ones, as long as both sides carry the same value.
			name:     "arbitrary values",
			fileName: "Component.tsx",
			source:   `const element = <div className="px-[3px] py-[3px]" />;`,
			wantIds:  []string{"canonicalCollapse"},
		},
		{
			// Recursion: three merges to reach `m-1`, reported as three steps. The engine reports one
			// finding for the whole set; reporting each step is more actionable and does not change
			// what the author has to do.
			name:     "four margins collapse through three steps",
			fileName: "Component.tsx",
			source:   `const element = <div className="mt-1 mb-1 ml-1 mr-1" />;`,
			wantIds:  []string{"canonicalCollapse", "canonicalCollapse", "canonicalCollapse"},
		},
		{
			name:     "on a callee surface",
			fileName: "Component.tsx",
			source:   `const merged = cn('px-4 py-4');`,
			wantIds:  []string{"canonicalCollapse"},
		},
		{
			name:     "on a variable surface",
			fileName: "Styles.ts",
			source:   `const className = 'w-8 h-8';`,
			wantIds:  []string{"canonicalCollapse"},
		},
		{
			// Unrelated neighbours must not prevent the merge or join it.
			name:     "collapse among other classes",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex px-4 items-center py-4 gap-2" />;`,
			wantIds:  []string{"canonicalCollapse"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runCanonicalFixture(t, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The clean half. The near-misses are the load-bearing ones: each differs from a real collapse in
// exactly one of the three preconditions, and relaxing any of them reports correct code.
func TestEnforceCanonicalClassesStaysSilent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// Values differ, so nothing merges.
			name:     "different values",
			fileName: "Component.tsx",
			source:   `const element = <div className="px-4 py-2" />;`,
		},
		{
			name:     "different values on width and height",
			fileName: "Component.tsx",
			source:   `const element = <div className="w-8 h-4" />;`,
		},
		{
			// Variants differ, so the two apply in different states and cannot merge.
			name:     "one side carries a variant",
			fileName: "Component.tsx",
			source:   `const element = <div className="px-4 sm:py-4" />;`,
		},
		{
			// Importance differs.
			name:     "one side is important",
			fileName: "Component.tsx",
			source:   `const element = <div className="px-4 py-4!" />;`,
		},
		{
			// Roots that are not a family, however similar they look.
			name:     "unrelated roots",
			fileName: "Component.tsx",
			source:   `const element = <div className="ml-2 pt-2" />;`,
		},
		{
			name:     "single class",
			fileName: "Component.tsx",
			source:   `const element = <div className="ml-2" />;`,
		},
		{
			name:     "statics",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex items-center truncate" />;`,
		},
		{
			// A collapse is not a duplicate: `space-x-2` alone has nothing to merge with.
			name:     "no pair present",
			fileName: "Component.tsx",
			source:   `const element = <div className="space-x-2 text-left" />;`,
		},
		{
			name:     "unrelated attribute",
			fileName: "Component.tsx",
			source:   `const element = <div title="px-4 py-4" />;`,
		},
		{
			name:     "unrelated callee",
			fileName: "Component.tsx",
			source:   `const value = someOtherFunction('px-4 py-4');`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runCanonicalFixture(t, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestCanonicalMessageNamesTheShorterSpelling checks the part an author acts on.
//
// A finding that says only "these collapse" leaves the reader to work out into what, which for
// `w-8 h-8` is `size-8` and is not guessable from the class names.
func TestCanonicalMessageNamesTheShorterSpelling(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		source string
		want   string
	}{
		{source: `const element = <div className="px-4 py-4" />;`, want: "p-4"},
		{source: `const element = <div className="w-8 h-8" />;`, want: "size-8"},
		{source: `const element = <div className="border-l border-r" />;`, want: "border-x"},
		{source: `const element = <div className="sm:px-4 sm:py-4" />;`, want: "sm:p-4"},
		{source: `const element = <div className="px-[3px] py-[3px]" />;`, want: "p-[3px]"},
		{source: `const element = <div className="rounded-tl-md rounded-tr-md" />;`, want: "rounded-t-md"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.want, func(t *testing.T) {
			t.Parallel()
			result := runCanonicalFixture(t, "Component.tsx", testCase.source)
			if len(result.Diagnostics) == 0 {
				t.Fatal("expected a finding, got none")
			}
			description := result.Diagnostics[0].Message.Description
			if !strings.Contains(description, "\""+testCase.want+"\"") {
				t.Errorf("message does not name %q: %s", testCase.want, description)
			}
		})
	}
}

// TestCanonicalProposesNoFix guards a deliberate absence.
//
// The collapse is mechanical; splicing the replacement into a class list is not, because the
// resulting order is `enforce-consistent-class-order`'s concern and a rewrite that also reordered
// would fight it.
func TestCanonicalProposesNoFix(t *testing.T) {
	t.Parallel()
	result := runCanonicalFixture(t, "Component.tsx",
		`const element = <div className="px-4 py-4" />;`)

	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	if len(result.Diagnostics[0].Fixes) != 0 || len(result.Diagnostics[0].Suggestions) != 0 {
		t.Fatal("this rule reports the shorter spelling and leaves the rewrite to the author, because " +
			"splicing it in decides an ordering another rule owns")
	}
}

// TestPreconditionsAreCheckedNotAssumed is the known-dirty control.
//
// The tempting implementation looks up the two roots in the table and stops there. It passes every
// violation fixture above and reports `px-4 py-2`, `px-4 sm:py-4` and `px-4 py-4!`, all of which are
// correct code. A rule that fires on correct code is one somebody turns off, so the preconditions
// are the rule rather than a detail of it.
func TestPreconditionsAreCheckedNotAssumed(t *testing.T) {
	t.Parallel()
	correctCode := []struct {
		name   string
		source string
	}{
		{name: "values differ", source: `const element = <div className="px-4 py-2" />;`},
		{name: "variants differ", source: `const element = <div className="px-4 sm:py-4" />;`},
		{name: "importance differs", source: `const element = <div className="px-4 py-4!" />;`},
	}

	for _, testCase := range correctCode {
		result := runCanonicalFixture(t, "Component.tsx", testCase.source)
		if len(result.Diagnostics) != 0 {
			t.Errorf("%s: reported %d findings on correct code, so a precondition is not being checked",
				testCase.name, len(result.Diagnostics))
		}

		// And the control has to be able to fail: the same roots with everything matching must
		// report, or the silence above proves only that the table lookup is broken.
		matching := runCanonicalFixture(t, "Component.tsx",
			`const element = <div className="px-4 py-4" />;`)
		if len(matching.Diagnostics) == 0 {
			t.Fatal("the rule stayed silent on a real collapse, so the silent cases above prove nothing")
		}
	}
}

// TestLongestRootWinsInCollapse pins the root resolution that has been wrong three times.
//
// `border-l` is a root in its own right, not `border` with value `l`. Taking the shorter match
// makes `border-l` and `border-r` look like the same root with different values, which fails the
// merge-key test and silently stops reporting `border-x`.
//
// The boundary is now `parseCandidate`'s rather than a prefix walk over a generated table, which is
// the correction the migration carries: the walk re-derived where a root ends, and re-deriving it is
// what lost `border-x` in the first place. `bg-linear-to-b` is the case that shows the difference on
// real code — the walk read it as root `bg` with value `linear-to-b`, and `bg-linear` is a root.
func TestLongestRootWinsInCollapse(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {
		system := designSystem.System

		parts, canParse := splitCandidateIn("border-l", system)
		if !canParse {
			t.Fatal("border-l did not resolve to a root at all")
		}
		if parts.Root != "border-l" {
			t.Errorf("border-l resolved to root %q with value %q, want root border-l and no value. "+
				"Taking the shorter match is what silently stopped border-x being reported.",
				parts.Root, parts.Value)
		}
		if parts.Value != "" {
			t.Errorf("border-l carries value %q, want none", parts.Value)
		}

		// And the value boundary: `p` must not claim `px-4`.
		pxParts, canParsePx := splitCandidateIn("px-4", system)
		if !canParsePx {
			t.Fatal("px-4 did not resolve to a root")
		}
		if pxParts.Root != "px" {
			t.Errorf("px-4 resolved to root %q, want px. A root matching across a value boundary would "+
				"conflate padding with padding-inline.", pxParts.Root)
		}

		// A root the shipped prefix walk could not find, because it is longer than the one it settled on.
		linearParts, canParseLinear := splitCandidateIn("bg-linear-to-b", system)
		if !canParseLinear {
			t.Fatal("bg-linear-to-b did not resolve to a root")
		}
		if linearParts.Root != "bg-linear" {
			t.Errorf("bg-linear-to-b resolved to root %q, want bg-linear. The prefix walk read this as "+
				"root bg with value linear-to-b, which buckets it with every other bg utility.",
				linearParts.Root)
		}
	})
}

// TestSplitValueCarriesEverythingAfterTheRoot pins the two pieces a parsed value does not include.
//
// Both were found by differencing the live split against the shipped prefix walk over the corpus,
// and both are the permissive direction: a value missing a piece makes two different classes compare
// equal on the merge precondition, which reports correct code.
//
// Measured before the fix: 88 of the corpus's classes split differently, 84 of them because the
// modifier was dropped.
func TestSplitValueCarriesEverythingAfterTheRoot(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {
		system := designSystem.System

		testCases := []struct {
			className string
			wantRoot  string
			wantValue string
			why       string
		}{
			{
				className: "bg-black/20",
				wantRoot:  "bg",
				wantValue: "black/20",
				why: "the modifier is a sibling field of the value, so dropping it makes bg-black/20 " +
					"and bg-black/60 compare equal",
			},
			{
				className: "-translate-x-1/2",
				wantRoot:  "-translate-x",
				wantValue: "1/2",
				why: "a fraction is carried beside the value rather than inside it, because the slash " +
					"is ambiguous and the parser refuses to guess",
			},
			{
				className: "bg-emerald-500/[0.07]",
				wantRoot:  "bg",
				wantValue: "emerald-500/[0.07]",
				why: "an ARBITRARY modifier is the only shape the modifier branch is reachable for, " +
					"because a named one is mirrored into Fraction and taken by the branch above it",
			},
			{
				className: "border-l",
				wantRoot:  "border-l",
				wantValue: "",
				why:       "a root with nothing after it, which rebuildClass must not append a dash for",
			},
		}

		for _, testCase := range testCases {
			parts, canParse := splitCandidateIn(testCase.className, system)
			if !canParse {
				t.Errorf("%s did not resolve to a root", testCase.className)
				continue
			}
			if parts.Root != testCase.wantRoot || parts.Value != testCase.wantValue {
				t.Errorf("%s split to root %q value %q, want root %q value %q: %s",
					testCase.className, parts.Root, parts.Value,
					testCase.wantRoot, testCase.wantValue, testCase.why)
			}
		}
	})
}

// TestRebuiltClassIsWritable pins that a suggested class can actually be typed into a class list.
//
// `parseCandidate` decodes arbitrary values: `grid-cols-[1fr_auto]` becomes `1fr auto` and
// `ring-(--x)` becomes `[var(--x)]`. Those decodings are correct and are what the merge precondition
// must compare, and printing either back would name a class nobody can write — the first has a space
// in it, and a class attribute is split on whitespace.
//
// So the parts carry both forms. Measured on the corpus: 16 classes decode to something different
// from what was written, every one an arbitrary value carrying `_` or the `(--x)` shorthand.
func TestRebuiltClassIsWritable(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, designSystem DesignSystemResult) {
		system := designSystem.System

		for _, className := range []string{
			"grid-cols-[1fr_auto]",
			"ring-(--color-content--2)",
			"hover:bg-(--background--1)",
			"px-4",
			"bg-black/20",
			"-translate-x-1/2",
			"px-[3px]",
			"px-4!",
		} {
			parts, canParse := splitCandidateIn(className, system)
			if !canParse {
				t.Errorf("%s did not resolve to a root", className)
				continue
			}
			// Rebuilt around its own root, so the only thing that can differ is the spelling.
			if rebuilt := rebuildClass(parts, parts.Root); rebuilt != className {
				t.Errorf("%s rebuilds as %q, which is not the class the author wrote. A suggestion "+
					"spelled this way cannot be pasted into a class attribute.", className, rebuilt)
			}
		}
	})
}

// TestCanonicalFixturesActuallyRan is what stops this file from going green on nothing.
//
// Every fixture above skips when no installed tailwindcss can be found, and `go test` prints ok for
// a package whose every case skipped, so a mistake in where the walk starts looks identical to a
// machine that lacks the package. That was not hypothetical for the class-order fixtures and this
// file uses the same helper shape.
func TestCanonicalFixturesActuallyRan(t *testing.T) {
	t.Parallel()
	packageRoot := unknownFixturePackageRoot()
	if packageRoot == "" {
		t.Fatalf(
			"the vendored tailwindcss at %s is missing, so every fixture in this file would skip and the "+
				"package still report ok; restore it (internal/lint/rules/tailwind/vendored)",
			vendored.TailwindPackageRoot(),
		)
	}

	result := runCanonicalFixture(t, "Component.tsx",
		`const element = <div className="px-4 py-4" />;`)
	rule_testing.ExpectFindings(t, result, "canonicalCollapse")
}

// TestIgnoredClassesAreExempt covers the option the oxlint configuration actually supplies.
//
// A rule declaring no options at all made oxlint refuse the whole plugin with "does not accept
// options", which failed as a silent zero-finding run rather than a crash.
func TestIgnoredClassesAreExempt(t *testing.T) {
	t.Parallel()
	options := EnforceCanonicalClassesOptions{Ignore: []string{`^px-4$`}}

	result := runCanonicalFixtureWithOptions(t, "Component.tsx",
		`const element = <div className="px-4 py-4" />;`, options)
	if len(result.Diagnostics) != 0 {
		t.Errorf("an ignored class should not participate in a collapse, got %d findings",
			len(result.Diagnostics))
	}

	// Without the exemption the same source reports, so the test above is not passing because the
	// rule is broken.
	unignored := runCanonicalFixture(t, "Component.tsx",
		`const element = <div className="px-4 py-4" />;`)
	if len(unignored.Diagnostics) == 0 {
		t.Fatal("the rule stayed silent without any ignore option, so the exemption proves nothing")
	}
}

// TestCanonicalCollapseOption is upstream's `collapse`, each way, on upstream's own case: on by default
// `w-10 h-10` is `size-10`, and off nothing this rule reports remains.
func TestCanonicalCollapseOption(t *testing.T) {
	t.Parallel()
	const source = `const element = <img className="w-10 h-10 flex" />;`
	off, on := false, true

	if result := runCanonicalFixture(t, "Component.tsx", source); len(result.Diagnostics) != 1 {
		t.Fatalf("by default the pair should collapse, got %d findings", len(result.Diagnostics))
	}
	if result := runCanonicalFixtureWithOptions(t, "Component.tsx", source,
		EnforceCanonicalClassesOptions{Collapse: &on}); len(result.Diagnostics) != 1 {
		t.Fatalf("collapse: true is the default and should collapse too, got %d findings", len(result.Diagnostics))
	}
	if result := runCanonicalFixtureWithOptions(t, "Component.tsx", source,
		EnforceCanonicalClassesOptions{Collapse: &off}); len(result.Diagnostics) != 0 {
		t.Fatalf("collapse: false should report no collapse, got %d findings", len(result.Diagnostics))
	}
}

// TestCanonicalLogicalOption is upstream's `logical`, each way. `mr-2 ml-2` is upstream's own case for
// it: `mx` declares `margin-inline`, which matches `margin-right` and `margin-left` only while the
// canonicalizer reads logical properties as physical ones. `w-8 h-8` merges into `size-8` through a
// utility relationship rather than a logical property, which the generator measured
// (RequiresLogicalToPhysical false), and is the control that `logical: false` silences only the
// families that need it.
func TestCanonicalLogicalOption(t *testing.T) {
	t.Parallel()
	const logicalPair = `const element = <img className="mr-2 ml-2" />;`
	const physicalPair = `const element = <img className="w-8 h-8" />;`
	off := false

	if result := runCanonicalFixture(t, "Component.tsx", logicalPair); len(result.Diagnostics) != 1 {
		t.Fatalf("by default mr-2 ml-2 should collapse to mx-2, got %d findings", len(result.Diagnostics))
	}
	if result := runCanonicalFixtureWithOptions(t, "Component.tsx", logicalPair,
		EnforceCanonicalClassesOptions{Logical: &off}); len(result.Diagnostics) != 0 {
		t.Fatalf("logical: false should leave mr-2 ml-2 alone, got %d findings", len(result.Diagnostics))
	}
	if result := runCanonicalFixtureWithOptions(t, "Component.tsx", physicalPair,
		EnforceCanonicalClassesOptions{Logical: &off}); len(result.Diagnostics) != 1 {
		t.Fatalf("logical: false should still collapse w-8 h-8, got %d findings", len(result.Diagnostics))
	}
}
