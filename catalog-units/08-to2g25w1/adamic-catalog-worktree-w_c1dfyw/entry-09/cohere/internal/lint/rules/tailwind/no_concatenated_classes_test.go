package tailwind

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every expectation in this file was measured by running `better-tailwindcss/no-concatenated-classes`
// over probe fixtures against the real plugin, not read off its source. The rule's name suggests two
// different behaviors, both wrong, and only running it settles which one it has.
//
// The measured verdicts, verbatim:
//
//	`px-${size}`             1 finding
//	`flex ${other}`          silent
//	`${prefix}-4`            1 finding
//	`${prefix} flex`         silent
//	`flex ${a} gap-${b}`     1 finding   (only the `gap-` seam)
//	`bg-[${color}]`          2 findings  (both seams)
//	'px-' + size             silent      (string concatenation is out of scope)
//	'bg-[' + color + ']'     silent
//	'a ' + 'b'               silent

func TestNoConcatenatedClassesReportsGluedFragments(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
		wantIds  []string
	}{
		{
			// The canonical case: a prefix waiting for a value.
			name:     "fragment before the hole",
			fileName: "Component.tsx",
			source:   "const element = <div className={`px-${size}`} />;",
			wantIds:  []string{"concatenatedClass"},
		},
		{
			// The mirror: a suffix after the hole. Easy to miss, because the eye reads the
			// interpolation as the start of the class.
			name:     "fragment after the hole",
			fileName: "Component.tsx",
			source:   "const element = <div className={`${prefix}-4`} />;",
			wantIds:  []string{"concatenatedClass"},
		},
		{
			// Both seams of one hole are separate defects, so an arbitrary value assembled around an
			// interpolation reports twice. Measured: upstream reports 2 here.
			name:     "both sides of one hole",
			fileName: "Component.tsx",
			source:   "const element = <div className={`bg-[${color}]`} />;",
			wantIds:  []string{"concatenatedClass", "concatenatedClass"},
		},
		{
			// A clean seam and a glued seam in one template: exactly one finding, at the glued one.
			// This is the case that separates a boundary-aware rule from one that flags any
			// interpolated literal.
			name:     "one clean seam and one glued seam",
			fileName: "Component.tsx",
			source:   "const element = <div className={`flex ${a} gap-${b}`} />;",
			wantIds:  []string{"concatenatedClass"},
		},
		{
			// The callee surface, which an attribute-only rule cannot see.
			name:     "inside a class-merging call",
			fileName: "Component.tsx",
			source:   "const merged = cn(`px-${size}`);",
			wantIds:  []string{"concatenatedClass"},
		},
		{
			// The variable surface.
			name:     "assigned to a class-named variable",
			fileName: "Styles.ts",
			source:   "const className = `rounded-${radius}`;",
			wantIds:  []string{"concatenatedClass"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoConcatenatedClasses, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The clean half, and for this rule it carries more weight than usual.
//
// The ahra tree holds 141 interpolated class literals and zero of them are findings, so a rule that
// reported every interpolated literal would produce 141 false positives on a tree that is actually
// correct. Every case here is a shape that appears in real code and must stay silent.
func TestNoConcatenatedClassesStaysSilent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// The shape the rule is teaching people to write.
			name:     "whitespace before the hole",
			fileName: "Component.tsx",
			source:   "const element = <div className={`flex ${extra}`} />;",
		},
		{
			name:     "whitespace after the hole",
			fileName: "Component.tsx",
			source:   "const element = <div className={`${extra} flex`} />;",
		},
		{
			name:     "whitespace on both sides",
			fileName: "Component.tsx",
			source:   "const element = <div className={`flex ${extra} gap-2`} />;",
		},
		{
			// A tab separates class names as well as a space does. Testing for `' '` alone would
			// report every tab-indented multi-line class list, which is a large fraction of real
			// markup.
			name:     "tab before the hole",
			fileName: "Component.tsx",
			source:   "const element = <div className={`flex\t${extra}`} />;",
		},
		{
			// A newline is whitespace too, and multi-line class templates are common.
			name:     "newline before the hole",
			fileName: "Component.tsx",
			source:   "const element = <div className={`flex\n${extra}`} />;",
		},
		{
			// Nothing static touches the holes at all.
			name:     "adjacent holes",
			fileName: "Component.tsx",
			source:   "const element = <div className={`${a}${b}`} />;",
		},
		{
			// A template with no substitutions is a plain string; the string reader owns it.
			name:     "template with no holes",
			fileName: "Component.tsx",
			source:   "const element = <div className={`flex items-center`} />;",
		},
		{
			// Out of scope, measured: upstream stays silent on `+` concatenation.
			name:     "string concatenation with a variable",
			fileName: "Component.tsx",
			source:   `const element = <div className={'px-' + size} />;`,
		},
		{
			// The real-tree shape that a first version of the measurement harness wrongly flagged,
			// in TableHeaderCell.tsx.
			name:     "arbitrary value by concatenation",
			fileName: "Component.tsx",
			source:   `const element = <div className={'bg-[' + color + ']'} />;`,
		},
		{
			// The real-tree shape from TableTheme.ts: two complete literals joined.
			name:     "two complete literals joined",
			fileName: "Styles.ts",
			source:   `const className = 'border-b border--2 ' + 'data-[state=selected]:x';`,
		},
		{
			// Choosing between whole class names is the repair this rule wants, so it must not fire
			// on it.
			name:     "ternary between whole classes",
			fileName: "Component.tsx",
			source:   `const element = <div className={isWide ? 'px-8' : 'px-4'} />;`,
		},
		{
			// A template on an attribute nobody said carries classes.
			name:     "unrelated attribute",
			fileName: "Component.tsx",
			source:   "const element = <div title={`px-${size}`} />;",
		},
		{
			// A template in a call nobody said combines classes.
			name:     "unrelated callee",
			fileName: "Component.tsx",
			source:   "const value = someOtherFunction(`px-${size}`);",
		},
		{
			name:     "unrelated variable",
			fileName: "Styles.ts",
			source:   "const description = `px-${size}`;",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoConcatenatedClasses, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoConcatenatedClassesNamesTheFragment checks the message, which is the whole value here.
//
// The upstream message says only "avoid dynamic class construction". In a template holding thirty
// classes and three holes, that tells the reader nothing about which seam to fix. Naming the
// fragment is the difference between an actionable finding and one that gets suppressed.
func TestNoConcatenatedClassesNamesTheFragment(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		source       string
		wantFragment string
	}{
		{
			name:         "prefix before the hole",
			source:       "const element = <div className={`px-${size}`} />;",
			wantFragment: "px-",
		},
		{
			name:         "suffix after the hole",
			source:       "const element = <div className={`${prefix}-4`} />;",
			wantFragment: "-4",
		},
		{
			name:         "the glued seam among clean ones",
			source:       "const element = <div className={`flex ${a} gap-${b}`} />;",
			wantFragment: "gap-",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoConcatenatedClasses, "Component.tsx", testCase.source)
			if len(result.Diagnostics) == 0 {
				t.Fatal("expected a finding, got none")
			}
			description := result.Diagnostics[0].Message.Description
			if !strings.Contains(description, "\""+testCase.wantFragment+"\"") {
				t.Errorf("message does not name the fragment %q: %s", testCase.wantFragment, description)
			}
		})
	}
}

// TestNoConcatenatedClassesProposesNoFix guards a deliberate absence.
//
// The repair is to write out the class names the code chooses between, which the rule cannot know.
// A fix here would have to invent them, and an autofix that invents class names would silently
// change what the markup renders. Asserting the absence keeps a later well-meaning change from
// adding one without arguing for it.
func TestNoConcatenatedClassesProposesNoFix(t *testing.T) {
	t.Parallel()
	result := rule_testing.Run(t, NoConcatenatedClasses, "Component.tsx",
		"const element = <div className={`px-${size}`} />;")

	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatal("this rule must not propose a fix: the repair requires knowing which class names the " +
			"author meant, and inventing them would change what renders")
	}
	if len(result.Diagnostics[0].Suggestions) != 0 {
		t.Fatal("nor a suggestion, for the same reason")
	}
}

// TestBoundaryReadingIsWhitespaceAware is the known-dirty control.
//
// The tempting shape is "this literal is interpolated, therefore report it", which is what the rule's
// name suggests and what a reader who skipped the boundary logic would write. It measures correct on
// every violation fixture above and produces 141 false positives on the real tree.
//
// So the narrowed reading runs against the silent fixtures and is required to disagree. A rule whose
// discrimination has never been exercised has been shown to detect, never to discriminate.
func TestBoundaryReadingIsWhitespaceAware(t *testing.T) {
	t.Parallel()
	// Sources with an interpolation that must stay silent. A report-any-interpolation rule flags
	// every one of these.
	silentButInterpolated := []string{
		"const element = <div className={`flex ${extra}`} />;",
		"const element = <div className={`${extra} flex`} />;",
		"const element = <div className={`${a}${b}`} />;",
		"const element = <div className={`flex\t${extra}`} />;",
	}

	for _, source := range silentButInterpolated {
		result := rule_testing.Run(t, NoConcatenatedClasses, "Component.tsx", source)
		if len(result.Diagnostics) != 0 {
			t.Errorf("a whitespace-aware rule must stay silent on %q, it reported %d", source, len(result.Diagnostics))
		}
	}

	// And the discrimination has to be real: the same reader must fire when the whitespace is gone.
	glued := rule_testing.Run(t, NoConcatenatedClasses, "Component.tsx",
		"const element = <div className={`flex ${a} gap-${b}`} />;")
	if len(glued.Diagnostics) == 0 {
		t.Fatal("the rule stayed silent on a genuinely glued seam, so the silence above proves nothing")
	}
}
