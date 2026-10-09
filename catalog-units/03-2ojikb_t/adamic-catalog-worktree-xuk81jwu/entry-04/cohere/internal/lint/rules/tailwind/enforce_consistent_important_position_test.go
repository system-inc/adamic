package tailwind

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every expectation here was measured against eslint-plugin-better-tailwindcss 4.7.0 on the ahra
// tree rather than read off its source.
//
// These fixtures use `rule_testing.Run` rather than the class-order suite's program harness, and that
// is a claim about the rule rather than a shortcut: the marker's position is a property of the
// string, so this rule asks nothing of the design system and declares no ProgramReads. The proof it
// needs none is `unknown class carries the marker` below, which reports on a class no theme
// defines.
func runImportantPositionFixture(t *testing.T, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.Run(t, EnforceConsistentImportantPosition, "Component.tsx", source)
}

// TestEnforceConsistentImportantPositionReports covers what upstream reports.
func TestEnforceConsistentImportantPositionReports(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		source string
	}{
		{
			name:   "leading marker",
			source: `const element = <div className="!text-red-500" />;`,
		},
		// The marker sits on the utility rather than on the class, so a check that read the whole
		// string would miss every variant-carrying class.
		{
			name:   "leading marker behind a variant",
			source: `const element = <div className="hover:!flex" />;`,
		},
		{
			name:   "leading marker behind several variants",
			source: `const element = <div className="md:hover:!text-sm" />;`,
		},
		/*
		 * Both spellings of a negative collapse onto one answer, and this is the pair a reader
		 * predicts wrongly. Upstream strips the sign before it looks for the marker, so a marker
		 * written on either side of the sign is found, and the rebuild always puts the sign outside.
		 */
		{
			name:   "marker outside the negative sign",
			source: `const element = <div className="!-mt-4" />;`,
		},
		{
			name:   "marker inside the negative sign",
			source: `const element = <div className="-!mt-4" />;`,
		},
		// Membership is not part of the question. This class is in no theme and still reports,
		// which is the evidence the rule needs no design system.
		{
			name:   "unknown class carries the marker",
			source: `const element = <div className="!unknown-class" />;`,
		},
		// An arbitrary value can hold anything, including the marker character itself.
		{
			name:   "arbitrary value containing the marker",
			source: `const element = <div className="!bg-[url('a!b.png')]" />;`,
		},
		{
			name:   "arbitrary property",
			source: `const element = <div className="![color:red]" />;`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runImportantPositionFixture(t, testCase.source), "importantPosition")
		})
	}
}

// TestEnforceConsistentImportantPositionStaysSilent covers what upstream leaves alone.
func TestEnforceConsistentImportantPositionStaysSilent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		source string
	}{
		{
			name:   "trailing marker",
			source: `const element = <div className="text-red-500!" />;`,
		},
		{
			name:   "trailing marker behind a variant",
			source: `const element = <div className="hover:flex!" />;`,
		},
		{
			name:   "trailing marker on a negative",
			source: `const element = <div className="-mt-4!" />;`,
		},
		{
			name:   "no marker",
			source: `const element = <div className="flex items-center" />;`,
		},
		/*
		 * A marker on both ends is accepted, which is upstream's behaviour rather than an omission:
		 * its condition returns early when the marker is present at the position being enforced,
		 * whatever else is true.
		 */
		{
			name:   "marker on both ends",
			source: `const element = <div className="!flex!" />;`,
		},
		/*
		 * A class that is nothing but the marker. Upstream reports it with the replacement `!`, a
		 * finding whose fix changes nothing and which would then report forever. Declined here, and
		 * the divergence is deliberate; see the rule's own comment.
		 */
		{
			name:   "the marker alone",
			source: `const element = <div className="!" />;`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runImportantPositionFixture(t, testCase.source))
		})
	}
}

// TestEnforceConsistentImportantPositionLegacy pins the configured direction.
//
// The option exists because a repository mid-migration wants to name its target rather than have
// one read off its node_modules. Under `legacy` every expectation above inverts.
func TestEnforceConsistentImportantPositionLegacy(t *testing.T) {
	t.Parallel()
	legacy := EnforceConsistentImportantPositionOptions{Position: importantPositionLegacy}

	reported := rule_testing.RunWithOptions(t, EnforceConsistentImportantPosition, "Component.tsx",
		`const element = <div className="text-red-500!" />;`, legacy)
	rule_testing.ExpectFindings(t, reported, "importantPosition")

	silent := rule_testing.RunWithOptions(t, EnforceConsistentImportantPosition, "Component.tsx",
		`const element = <div className="!text-red-500" />;`, legacy)
	rule_testing.ExpectClean(t, silent)
}

// TestEnforceConsistentImportantPositionNamesBothSpellings pins the message content.
func TestEnforceConsistentImportantPositionNamesBothSpellings(t *testing.T) {
	t.Parallel()
	result := runImportantPositionFixture(t, `const element = <div className="hover:!flex" />;`)
	rule_testing.ExpectFindings(t, result, "importantPosition")

	if len(result.Diagnostics) == 0 {
		t.Fatal("expected a finding")
	}
	message := result.Diagnostics[0].Message.Description
	for _, want := range []string{"hover:!flex", "hover:flex!"} {
		if !strings.Contains(message, want) {
			t.Errorf("the message does not name %q: %s", want, message)
		}
	}
}
