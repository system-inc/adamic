package tailwind

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every expectation was measured against eslint-plugin-better-tailwindcss 4.7.0, and the table this
// rule matches on was extracted mechanically from its source rather than retyped: it is 48 rules,
// and a transcription slip would be a silent miss on one utility family rather than a failure
// anywhere a suite would notice.
func runShorthandFixture(t *testing.T, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.Run(t, EnforceShorthandClasses, "Component.tsx", source)
}

// TestEnforceShorthandClassesReports covers what upstream reports.
func TestEnforceShorthandClassesReports(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name       string
		source     string
		correction string
	}{
		{
			name:       "logical inline pair",
			source:     `const element = <div className="ps-4 pe-4" />;`,
			correction: "px-4",
		},
		{
			name:       "physical block pair",
			source:     `const element = <div className="pt-2 pb-2" />;`,
			correction: "py-2",
		},
		// The longest rule in a family wins, so four sides collapse straight to one class rather
		// than to two axis classes that would then collapse again.
		{
			name:       "all four margin sides",
			source:     `const element = <div className="ml-1 mr-1 mt-1 mb-1" />;`,
			correction: "m-1",
		},
		{
			name:       "width and height",
			source:     `const element = <div className="w-4 h-4" />;`,
			correction: "size-4",
		},
		{
			name:       "all four inset sides",
			source:     `const element = <div className="top-0 right-0 bottom-0 left-0" />;`,
			correction: "inset-0",
		},
		// The variants have to agree, and the rebuilt class carries them.
		{
			name:       "pair behind the same variant",
			source:     `const element = <div className="hover:ps-4 hover:pe-4" />;`,
			correction: "hover:px-4",
		},
		// So does the negative sign.
		{
			name:       "negative pair",
			source:     `const element = <div className="-mt-2 -mb-2" />;`,
			correction: "-my-2",
		},
		// And the important marker, in either position.
		{
			name:       "pair with a leading marker",
			source:     `const element = <div className="!pt-2 !pb-2" />;`,
			correction: "!py-2",
		},
		{
			name:       "pair with a trailing marker",
			source:     `const element = <div className="pt-2! pb-2!" />;`,
			correction: "py-2!",
		},
		// An arbitrary value is just another captured value.
		{
			name:       "pair with arbitrary values",
			source:     `const element = <div className="ps-[2px] pe-[2px]" />;`,
			correction: "px-[2px]",
		},
		// An arbitrary variant groups the same way a named one does.
		{
			name:       "pair behind an arbitrary variant",
			source:     `const element = <div className="[&_button]:border-s-0 [&_button]:border-e-0" />;`,
			correction: "[&_button]:border-x-0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runShorthandFixture(t, testCase.source)
			rule_testing.ExpectFindings(t, result, "shorthandClasses")
			if len(result.Diagnostics) == 0 {
				t.Fatal("expected a finding")
			}
			if !strings.Contains(result.Diagnostics[0].Message.Description, testCase.correction) {
				t.Errorf("the message does not name the correction %q: %s",
					testCase.correction, result.Diagnostics[0].Message.Description)
			}
		})
	}
}

// TestEnforceShorthandClassesStaysSilent covers what upstream leaves alone.
//
// Each case here is one of the four things that must agree across a group. Dropping any one of the
// checks widens the rule silently, which is why each gets its own fixture rather than being taken
// on trust from the value check.
func TestEnforceShorthandClassesStaysSilent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		source string
	}{
		{
			name:   "values disagree",
			source: `const element = <div className="ps-4 pe-8" />;`,
		},
		{
			name:   "variants disagree",
			source: `const element = <div className="ps-4 hover:pe-4" />;`,
		},
		{
			name:   "negative sign disagrees",
			source: `const element = <div className="-mt-2 mb-2" />;`,
		},
		{
			name:   "important marker disagrees",
			source: `const element = <div className="!pt-2 pb-2" />;`,
		},
		// One half of a pair is not a pair.
		{
			name:   "only one side present",
			source: `const element = <div className="ps-4" />;`,
		},
		// The shorthand is already written alongside, so there is nothing to say.
		{
			name:   "shorthand already present",
			source: `const element = <div className="px-4 py-4 p-4" />;`,
		},
		{
			name:   "no shorthand candidates at all",
			source: `const element = <div className="flex items-center" />;`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runShorthandFixture(t, testCase.source))
		})
	}
}

// TestEnforceShorthandClassesReportsOncePerFamily pins the outer grouping.
//
// Several rules in one family can match the same classes. `px-4 py-4` matches both the axis-pair
// rule and, through them, nothing else; but `ml-1 mr-1 mt-1 mb-1` matches the four-side rule and
// both two-side rules. Upstream takes one per family, so this reports once rather than three times.
func TestEnforceShorthandClassesReportsOncePerFamily(t *testing.T) {
	t.Parallel()
	result := runShorthandFixture(t, `const element = <div className="ml-1 mr-1 mt-1 mb-1" />;`)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected exactly one finding, got %d", len(result.Diagnostics))
	}
	if !strings.Contains(result.Diagnostics[0].Message.Description, "m-1") {
		t.Errorf("expected the four-side collapse: %s", result.Diagnostics[0].Message.Description)
	}
}

// TestEnforceShorthandClassesReportsEachFamilySeparately pins the other half of that grouping.
//
// Two families in one literal are two findings, because the one-per-family rule is per family
// rather than per literal.
func TestEnforceShorthandClassesReportsEachFamilySeparately(t *testing.T) {
	t.Parallel()
	result := runShorthandFixture(t, `const element = <div className="ps-4 pe-4 mt-2 mb-2" />;`)
	if len(result.Diagnostics) != 2 {
		t.Fatalf("expected two findings, got %d", len(result.Diagnostics))
	}
}

// The suggested shorthand has to exist, which only a design system can say.
//
// `w-screen h-screen` matches the `w-(.*) h-(.*)` pattern and would suggest `size-screen`, a class
// Tailwind never had: the two set `100vw` and `100vh`. The pair stays, and `w-auto h-auto` beside it
// is the control that the rule still fires through the same harness. Before the check, both
// reported, and nested literals put two `w-screen h-screen` findings on ahra.
func TestEnforceShorthandClassesSuggestsOnlyClassesThatExist(t *testing.T) {
	t.Parallel()
	packageRoot := classOrderFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine, so there is no design system to ask")
	}
	run := func(source string) rule_testing.Result {
		return rule_testing.RunTypedFilesWithSetup(t, EnforceShorthandClasses, map[string]string{
			"Component.tsx":                 source,
			classOrderFixtureStylesheetPath: classOrderFixtureStylesheet,
		}, "Component.tsx", func(root string) {
			modules := filepath.Join(root, "node_modules")
			if err := os.MkdirAll(modules, 0o755); err != nil {
				t.Fatalf("creating the fixture node_modules: %v", err)
			}
			if err := os.Symlink(packageRoot, filepath.Join(modules, "tailwindcss")); err != nil {
				t.Fatalf("linking the installed tailwindcss into the fixture: %v", err)
			}
		})
	}

	rule_testing.ExpectClean(t, run(`const element = <div className="h-screen w-screen" />;`))
	rule_testing.ExpectFindings(t, run(`const element = <div className="h-auto w-auto" />;`), "shorthandClasses")
}
