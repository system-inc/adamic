package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const screamingFile = "/repository/source/Thing.ts"

func TestConsistencyNoScreamingSnakeCaseFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		{"local constant", "const MAX_RETRY_COUNT = 3;\n", "noScreamingSnakeCaseLocal"},
		{"exported constant", "export const HTTP_TIMEOUT = 5000;\n", "noScreamingSnakeCaseExported"},
		{"digits in the name", "const API_V2_URL = 'x';\n", "noScreamingSnakeCaseLocal"},
		{"two words", "const ORDER_COLUMNS = [];\n", "noScreamingSnakeCaseLocal"},
		// Rooted at process but not an environment read, so the exemption must not apply. This is
		// the case that separates "rooted at process" from "reads the environment": it is a plain
		// property access, so it reaches the env check rather than exiting early the way
		// process.argv[2] does as an element access.
		{"process.platform is not an env read", "const SOME_PLATFORM = process.platform;\n", "noScreamingSnakeCaseLocal"},
		{"process.argv is not an env read", "const SOME_ARGUMENT = process.argv[2];\n", "noScreamingSnakeCaseLocal"},
		// An env-shaped name whose value comes from somewhere else is still ours to name.
		{"env-looking name from a literal", "const DATABASE_URL = 'postgres://local';\n", "noScreamingSnakeCaseLocal"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoScreamingSnakeCase, screamingFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}
}

func TestConsistencyNoScreamingSnakeCaseStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"camelCase local", "const orderColumns = [];\n"},
		{"PascalCase export", "export const OrderColumns = [];\n"},
		// A single all-caps word is a different naming choice, and this rule does not have that
		// argument. The underscore is what makes the shape unambiguous.
		{"single all-caps word", "const RED = '#ff0000';\n"},
		{"single all-caps export", "export const RED = '#ff0000';\n"},
		// A shouting let is exotic, and the immutability premise the reasoning rests on does not
		// apply to it.
		{"let is not const", "let MAX_RETRY_COUNT = 3;\n"},
		{"var is not const", "var MAX_RETRY_COUNT = 3;\n"},
		// The exemptions, which are the whole design of the rule.
		{"direct process.env read", "const DATABASE_URL = process.env.DATABASE_URL;\n"},
		{"process.env with a nullish fallback", "const DATABASE_URL = process.env.DATABASE_URL ?? 'local';\n"},
		{"process.env with an or fallback", "const DATABASE_URL = process.env.DATABASE_URL || 'local';\n"},
		{"mixed case with underscore", "const Max_Retry = 3;\n"},
		{"lowercase with underscore", "const max_retry = 3;\n"},
		{"no declarations", "export function run() {\n    return 1;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoScreamingSnakeCase, screamingFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestConsistencyNoScreamingSnakeCaseRespectsTheAllowOption(t *testing.T) {
	t.Parallel()

	sourceText := "const STRIPE_WEBHOOK_SECRET = readSecret();\n"

	withoutOption := rule_testing.Run(t, ConsistencyNoScreamingSnakeCase, screamingFile, sourceText)
	rule_testing.ExpectFindings(t, withoutOption, "noScreamingSnakeCaseLocal")

	withOption := rule_testing.RunWithOptions(t, ConsistencyNoScreamingSnakeCase, screamingFile, sourceText,
		ConsistencyNoScreamingSnakeCaseOptions{Allow: []string{"STRIPE_WEBHOOK_SECRET"}})
	rule_testing.ExpectClean(t, withOption)

	// Exact match only. An allowlist that matched by prefix would admit names nobody approved.
	nearMiss := rule_testing.RunWithOptions(t, ConsistencyNoScreamingSnakeCase, screamingFile,
		"const STRIPE_WEBHOOK_SECRET_OLD = readSecret();\n",
		ConsistencyNoScreamingSnakeCaseOptions{Allow: []string{"STRIPE_WEBHOOK_SECRET"}})
	rule_testing.ExpectFindings(t, nearMiss, "noScreamingSnakeCaseLocal")
}

// The message has to carry a name the author can actually use, and the two forms differ by export.
func TestConsistencyNoScreamingSnakeCaseSuggestsTheRightCasing(t *testing.T) {
	t.Parallel()

	local := rule_testing.Run(t, ConsistencyNoScreamingSnakeCase, screamingFile, "const MAX_RETRY_COUNT = 3;\n")
	rule_testing.ExpectFindings(t, local, "noScreamingSnakeCaseLocal")
	if !strings.Contains(local.Diagnostics[0].Message.Description, `"maxRetryCount"`) {
		t.Fatalf("expected a camelCase suggestion, got: %s", local.Diagnostics[0].Message.Description)
	}

	exported := rule_testing.Run(t, ConsistencyNoScreamingSnakeCase, screamingFile, "export const MAX_RETRY_COUNT = 3;\n")
	rule_testing.ExpectFindings(t, exported, "noScreamingSnakeCaseExported")
	if !strings.Contains(exported.Diagnostics[0].Message.Description, `"MaxRetryCount"`) {
		t.Fatalf("expected a PascalCase suggestion, got: %s", exported.Diagnostics[0].Message.Description)
	}
}
