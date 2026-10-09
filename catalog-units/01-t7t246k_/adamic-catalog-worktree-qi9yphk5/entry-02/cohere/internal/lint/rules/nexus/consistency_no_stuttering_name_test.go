package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const stutterFile = "/repository/source/Thing.ts"

func TestConsistencyNoStutteringNameFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"result", "export const a = result.result;\n"},
		{"outcome", "export const a = outcome.outcome;\n"},
		{"data", "export const a = data.data;\n"},
		{"value", "export const a = value.value;\n"},
		{"response", "export const a = response.response;\n"},
		// Optional chaining is the same stutter written a different way.
		{"optional chain", "export const a = outcome?.outcome;\n"},
		// Nested in a larger expression, which is where these actually appear.
		{"inside a comparison", "export const a = outcome.outcome === 'Unreadable';\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoStutteringName, stutterFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "stutteringName")
		})
	}
}

func TestConsistencyNoStutteringNameStaysSilent(t *testing.T) {
	t.Parallel()

	// The whole design of this rule is that it stays silent on generic names that are doing their
	// job. The declaration-site version of this idea measured 1,355 hits and roughly forty were
	// real, so these cases are the rule's reason for existing rather than incidental coverage.
	cases := []struct {
		name       string
		sourceText string
	}{
		{"generic name with a real field", "export const a = response.json();\n"},
		{"generic name read plainly", "export const a = result.length;\n"},
		{"a real name that repeats", "export const a = user.user;\n"},
		{"different generic names", "export const a = result.value;\n"},
		{"a stutter on a non-generic word", "export const a = session.session;\n"},
		// A computed key is usually a variable, so only the written-out form is provably a stutter.
		{"computed access", "export const a = result['result'];\n"},
		{"deep access that does not stutter", "export const a = outcome.status.code;\n"},
		{"no member access at all", "export const a = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoStutteringName, stutterFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestConsistencyNoStutteringNameRespectsTheGenericNamesOption(t *testing.T) {
	t.Parallel()

	// The option replaces the default set rather than adding to it, so a name on the default list
	// stops being checked once a caller names its own vocabulary.
	options := ConsistencyNoStutteringNameOptions{GenericNames: []string{"payload"}}

	custom := rule_testing.RunWithOptions(t, ConsistencyNoStutteringName, stutterFile,
		"export const a = payload.payload;\n", options)
	rule_testing.ExpectFindings(t, custom, "stutteringName")

	replaced := rule_testing.RunWithOptions(t, ConsistencyNoStutteringName, stutterFile,
		"export const a = result.result;\n", options)
	rule_testing.ExpectClean(t, replaced)
}

func TestConsistencyNoStutteringNameNamesTheWord(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ConsistencyNoStutteringName, stutterFile, "export const a = outcome.outcome;\n")
	rule_testing.ExpectFindings(t, result, "stutteringName")

	description := result.Diagnostics[0].Message.Description
	if !strings.Contains(description, `"outcome.outcome"`) {
		t.Fatalf("expected the message to name the stutter, got: %s", description)
	}
}
