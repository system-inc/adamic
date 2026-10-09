package tailwind

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every expectation was measured against eslint-plugin-better-tailwindcss 4.7.0, and the
// corrections were checked against its own autofix output rather than only against which classes it
// reports: a rule that reports the right set and rewrites it wrongly passes a set comparison.
//
// `rule_testing.Run` rather than the class-order program harness, because the question is entirely about
// how a class is spelled and this rule declares no ProgramReads.
func runVariableSyntaxFixture(t *testing.T, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.Run(t, EnforceConsistentVariableSyntax, "Component.tsx", source)
}

// TestEnforceConsistentVariableSyntaxReports covers what upstream reports.
func TestEnforceConsistentVariableSyntaxReports(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name       string
		source     string
		correction string
	}{
		// The bare property in brackets.
		{
			name:       "bracketed property",
			source:     `const element = <div className="text-[--my-var]" />;`,
			correction: "text-(--my-var)",
		},
		// The call form. Both bracketed spellings collapse onto the parenthesised one, which is the
		// part a reader predicts wrongly.
		{
			name:       "bracketed var call",
			source:     `const element = <div className="text-[var(--my-var)]" />;`,
			correction: "text-(--my-var)",
		},
		{
			name:       "bracketed var call on a background",
			source:     `const element = <div className="bg-[var(--other)]" />;`,
			correction: "bg-(--other)",
		},
		// The variant is preserved, which is what the prefix split is for.
		{
			name:       "behind a variant",
			source:     `const element = <div className="hover:text-[--v]" />;`,
			correction: "hover:text-(--v)",
		},
		// The modifier rides outside the brackets and has to be put back untouched.
		{
			name:       "with an opacity modifier",
			source:     `const element = <div className="text-[--v]/50" />;`,
			correction: "text-(--v)/50",
		},
		// A property name may hold dashes and underscores.
		{
			name:       "property name with dashes and underscores",
			source:     `const element = <div className="text-[--a-b_c]" />;`,
			correction: "text-(--a-b_c)",
		},
		// A fallback inside the call survives the unwrap.
		{
			name:       "var call with a fallback",
			source:     `const element = <div className="w-[var(--x,10px)]" />;`,
			correction: "w-(--x,10px)",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runVariableSyntaxFixture(t, testCase.source)
			rule_testing.ExpectFindings(t, result, "variableSyntax")
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

// TestEnforceConsistentVariableSyntaxStaysSilent covers what upstream leaves alone.
func TestEnforceConsistentVariableSyntaxStaysSilent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		source string
	}{
		{
			name:   "already parenthesised",
			source: `const element = <div className="text-(--my-var)" />;`,
		},
		/*
		 * A colon in the base means the brackets name a property rather than read one, and
		 * rewriting them would change meaning rather than spelling. Upstream calls this skipping
		 * variable definitions.
		 */
		{
			name:   "arbitrary property",
			source: `const element = <div className="text-[color:red]" />;`,
		},
		/*
		 * The brackets hold arithmetic, not a property reference. Upstream's shorthand test carries
		 * a lookahead for exactly this, and dropping it would rewrite the calc into nonsense.
		 */
		{
			name:   "calc containing a var call",
			source: `const element = <div className="w-[calc(var(--x)*2)]" />;`,
		},
		// Two calls in one bracket are not a single property reference either.
		{
			name:   "several var calls in one value",
			source: `const element = <div className="shadow-[var(--a),var(--b)]" />;`,
		},
		{
			name:   "no custom property at all",
			source: `const element = <div className="flex items-center" />;`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runVariableSyntaxFixture(t, testCase.source))
		})
	}
}

// TestEnforceConsistentVariableSyntaxVariableForm pins the configured direction.
func TestEnforceConsistentVariableSyntaxVariableForm(t *testing.T) {
	t.Parallel()
	variableForm := EnforceConsistentVariableSyntaxOptions{Syntax: variableSyntaxVariable}

	reported := rule_testing.RunWithOptions(t, EnforceConsistentVariableSyntax, "Component.tsx",
		`const element = <div className="text-(--my-var)" />;`, variableForm)
	rule_testing.ExpectFindings(t, reported, "variableSyntax")
	if len(reported.Diagnostics) > 0 &&
		!strings.Contains(reported.Diagnostics[0].Message.Description, "text-[var(--my-var)]") {
		t.Errorf("expected the bracketed call form: %s", reported.Diagnostics[0].Message.Description)
	}

	silent := rule_testing.RunWithOptions(t, EnforceConsistentVariableSyntax, "Component.tsx",
		`const element = <div className="text-[var(--my-var)]" />;`, variableForm)
	rule_testing.ExpectClean(t, silent)
}
