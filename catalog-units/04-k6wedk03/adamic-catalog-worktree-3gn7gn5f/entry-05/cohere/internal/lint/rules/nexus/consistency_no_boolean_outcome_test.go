package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const outcomeFile = "/repository/source/Thing.ts"

func TestConsistencyNoBooleanOutcomeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"interface with success and error", "export interface LookupResult {\n    success: boolean;\n    error: string;\n}\n"},
		{"type literal with ok and value", "export type LookupResultType = {\n    ok: boolean;\n    value: string;\n};\n"},
		{"failed with a reason", "export interface CallResult {\n    failed: boolean;\n    reason: string;\n}\n"},
		{"succeeded with data", "export interface CallResult {\n    succeeded: boolean;\n    data: string;\n}\n"},
		{"isError alone with a message", "export interface CallResult {\n    isError: boolean;\n    message: string;\n}\n"},
		// A third-party envelope exempts only on its whole documented field set, so a shape of our own
		// that shares three of Cloudflare v4's four fields is still ours.
		{"three of cloudflare's four fields", "export interface SyncResult {\n    success: boolean;\n    errors: string[];\n    result: string;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoBooleanOutcome, outcomeFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "booleanOutcome")
		})
	}
}

func TestConsistencyNoBooleanOutcomeStaysSilent(t *testing.T) {
	t.Parallel()

	// Every one of these is a shape that would fire if the rule keyed on the flag name alone. The
	// exemptions are the difference between a rule people keep and one they switch off.
	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		{"a flag with no companion is state", outcomeFile, "export interface ToggleState {\n    isVisible: boolean;\n    succeeded: boolean;\n}\n"},
		{"a lone flag on a database row", outcomeFile, "export interface JobRow {\n    id: string;\n    succeeded: boolean;\n}\n"},
		{"react properties are display state", outcomeFile, "export interface BannerProperties {\n    isError: boolean;\n    message: string;\n}\n"},
		{"tanstack field set is not ours", outcomeFile, "export interface QueryResult {\n    isError: boolean;\n    isLoading: boolean;\n    error: string;\n}\n"},
		{"cloudflare's v4 envelope is not ours", outcomeFile, "export interface R2CloudflareApiResponseInterface {\n    success: boolean;\n    errors: unknown;\n    messages: unknown;\n    result: unknown;\n}\n"},
		{"isSuccess and isError are one settled state", outcomeFile, "export interface QueryResult {\n    isSuccess: boolean;\n    isError: boolean;\n    error: string;\n}\n"},
		// A discriminated union is the shape this rule steers toward, so it must never fire on one.
		{"a union alias is already the answer", outcomeFile, "export type PostOutcomeType =\n    | { outcome: 'Found'; value: string }\n    | { outcome: 'NotFound'; message: string };\n"},
		// A literal arm inside a discriminated union is fine, since only a bare boolean counts.
		{"a literal true is not a bare boolean", outcomeFile, "export interface FoundArm {\n    success: true;\n    value: string;\n}\n"},
		{"an unrelated interface", outcomeFile, "export interface Point {\n    x: number;\n    y: number;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoBooleanOutcome, testCase.fileName, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Generated code is held to the rule like hand-written code (Kirk's ruling on generated code, #c076xbg):
// a generator emitting a boolean outcome is fixed at the generator. The rule used to skip these files.
func TestConsistencyNoBooleanOutcomeHoldsGeneratedFilesToTheRule(t *testing.T) {
	t.Parallel()
	for _, fileName := range []string{"/repository/source/generated/Api.ts", "/repository/source/Api.generated.ts"} {
		result := rule_testing.Run(t, ConsistencyNoBooleanOutcome, fileName, "export interface LookupResult {\n    success: boolean;\n    error: string;\n}\n")
		rule_testing.ExpectFindings(t, result, "booleanOutcome")
	}
}

func TestConsistencyNoBooleanOutcomeRespectsTheAllowedTypeNamesOption(t *testing.T) {
	t.Parallel()

	// Telemetry envelopes carry timing or cost on both paths, since a failed model call still bills.
	// Splitting those into a union puts the cost in the success arm and loses it where it is needed.
	sourceText := "export interface ClaudeCallResultInterface {\n    success: boolean;\n    error: string;\n}\n"

	withoutOption := rule_testing.Run(t, ConsistencyNoBooleanOutcome, outcomeFile, sourceText)
	rule_testing.ExpectFindings(t, withoutOption, "booleanOutcome")

	withOption := rule_testing.RunWithOptions(t, ConsistencyNoBooleanOutcome, outcomeFile, sourceText,
		ConsistencyNoBooleanOutcomeOptions{AllowedTypeNames: []string{"ClaudeCallResultInterface"}})
	rule_testing.ExpectClean(t, withOption)
}

// The suggestion strips role suffixes so the name it offers is one someone would actually write.
func TestConsistencyNoBooleanOutcomeSuggestsAStrippedName(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ConsistencyNoBooleanOutcome, outcomeFile,
		"export interface ClaudeCallResultInterface {\n    success: boolean;\n    error: string;\n}\n")
	rule_testing.ExpectFindings(t, result, "booleanOutcome")

	description := result.Diagnostics[0].Message.Description
	if !strings.Contains(description, "ClaudeCallOutcomeType") {
		t.Fatalf("expected a stripped suggestion, got: %s", description)
	}
}
