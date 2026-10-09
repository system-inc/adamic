package tailwind

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every expectation in this file was measured against eslint-plugin-better-tailwindcss 4.7.0 rather
// than read off its source, and the two disagreed twice during the port. Both disagreements are
// preserved as fixtures below because each one is a shape a reader would predict wrongly.
//
// The fixtures share the class-order suite's harness for the reason that file states at length: the
// rule declares `ProgramReads` and reaches for a real stylesheet, so a fixture running through
// `rule_testing.Run` would exercise the nil-Program decline path and pass while proving nothing.
func runVariantOrderFixture(t *testing.T, fileName string, source string) rule_testing.Result {
	t.Helper()

	packageRoot := classOrderFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine, so the rule's design system cannot be " +
			"built and these fixtures would measure a decline rather than an order")
	}

	return rule_testing.RunTypedFilesWithSetup(t, EnforceConsistentVariantOrder, map[string]string{
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

// TestEnforceConsistentVariantOrderReports covers the cases upstream reports.
//
// Every source here was run through upstream on the ahra tree and reported, with the replacement
// this rule prints matching upstream's character for character.
func TestEnforceConsistentVariantOrderReports(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		source string
	}{
		// A media variant written after an element-scoped one. The plainest case and the one the
		// whole rule exists for.
		{
			name:   "media variant after a pseudo variant",
			source: `const element = <div className="hover:sm:flex" />;`,
		},
		// Two globals order against each other, and descending: `sm:lg:` reports as `lg:sm:`. That
		// direction is upstream's `orderB > orderA` returning +1 and reads backwards at first.
		{
			name:   "two breakpoints in ascending order",
			source: `const element = <div className="sm:lg:flex" />;`,
		},
		// `dark` is not global in a default build: its registration is a selector, not a media
		// query. So it sorts after every media variant despite a higher registration number.
		{
			name:   "dark before a breakpoint",
			source: `const element = <div className="dark:md:flex" />;`,
		},
		// A compound variant carries no flag and lands after the media variant.
		{
			name:   "compound variant before a breakpoint",
			source: `const element = <div className="group-hover:md:flex" />;`,
		},
		// An arbitrary variant is ordered by its dense rank and carries no flag either.
		{
			name:   "arbitrary variant before a breakpoint",
			source: `const element = <div className="[@media(hover)]:md:flex" />;`,
		},
		/*
		 * The class that caught the one defect a narrower suite would have missed. `has-[:checked]`
		 * holds a colon inside its brackets, so splitting the prefix with `strings.Split` yields
		 * three pieces for two variants, the length check against the parser fails, and the rule
		 * declines a class it should report. Silent, and found only by differential run.
		 */
		{
			name:   "arbitrary value containing a colon",
			source: `const element = <div className="has-[:checked]:md:flex" />;`,
		},
		// Three variants, two of them sub-global. The pair keeps its written order while the
		// breakpoint moves to the front, which is the stable-sort half of the comparison.
		{
			name:   "two pseudo variants keep their order while a breakpoint moves",
			source: `const element = <div className="focus:hover:lg:text-red-500" />;`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runVariantOrderFixture(t, "Component.tsx", testCase.source), "variantOrder")
		})
	}
}

// TestEnforceConsistentVariantOrderStaysSilent covers what upstream leaves alone.
//
// The first two are the restraint the rule is built around, and a version of this file that used
// the registry's full `Compare` reported both. Upstream's `compareVariantOrder` answers 0 when both
// variants sort below the global flag, so two element-scoped variants are accepted in either order.
func TestEnforceConsistentVariantOrderStaysSilent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		source string
	}{
		{
			name:   "two pseudo variants in either order",
			source: `const element = <div className="hover:focus:underline focus:hover:underline" />;`,
		},
		{
			name:   "dark and hover in either order",
			source: `const element = <div className="dark:hover:flex hover:dark:flex" />;`,
		},
		// Already correct: the media variant is outermost.
		{
			name:   "breakpoint already outermost",
			source: `const element = <div className="sm:hover:flex" />;`,
		},
		// A single variant has nothing to order.
		{
			name:   "one variant",
			source: `const element = <div className="hover:flex" />;`,
		},
		// No variants at all.
		{
			name:   "no variants",
			source: `const element = <div className="flex items-center" />;`,
		},
		/*
		 * `supports` is not global, which is the near miss most likely to be got wrong: a support
		 * query looks like a media query and is not one, because the rule it wraps still carries
		 * `&`. Upstream leaves this alone and so does this rule.
		 */
		{
			name:   "supports after a pseudo variant",
			source: `const element = <div className="hover:supports-[display:grid]:flex" />;`,
		},
		// A functional breakpoint variant keeps its element scope and is not global either.
		{
			name:   "max-md after a pseudo variant",
			source: `const element = <div className="hover:max-md:flex" />;`,
		},
		// A variant the parser does not know leaves the class unresolved, and an opinion formed
		// from a partial parse is worse than silence.
		{
			name:   "unknown variant",
			source: `const element = <div className="hover:unknown-variant:text-red-500" />;`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runVariantOrderFixture(t, "Component.tsx", testCase.source))
		})
	}
}

// TestEnforceConsistentVariantOrderNamesBothSpellings pins the message content.
//
// The replacement is the whole content of the finding: a reader told only that the order is wrong
// has to work out the permutation themselves, and the permutation is what the rule computed.
func TestEnforceConsistentVariantOrderNamesBothSpellings(t *testing.T) {
	t.Parallel()
	result := runVariantOrderFixture(t, "Component.tsx", `const element = <div className="dark:md:flex" />;`)
	rule_testing.ExpectFindings(t, result, "variantOrder")

	if len(result.Diagnostics) == 0 {
		t.Fatal("expected a finding")
	}
	message := result.Diagnostics[0].Message.Description
	for _, want := range []string{"dark:md:flex", "md:dark:flex"} {
		if !strings.Contains(message, want) {
			t.Errorf("the message does not name %q: %s", want, message)
		}
	}
}
