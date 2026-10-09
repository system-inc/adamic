package core

import (
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// noContinueFile is where the fixtures pretend to live.
//
// A `.ts` name: this rule has no file gate, and nothing in its corpus is JSX.
const noContinueFile = "/repository/source/NoContinue.ts"

// noContinueCase is one row of upstream's corpus.
type noContinueCase struct {
	// name is the corpus list and index the row came from, so a failure names a case that
	// can be found in upstream's own file rather than a number local to this table.
	name string

	// source is upstream's `code`, byte for byte.
	source string

	// options is the RAW JSON of upstream's options object, routed through the rule's own
	// exported decoder rather than built as a struct, so the decoder's defaults and its
	// empty-input path are under test.
	options string

	// ids are the message ids upstream produced for this input, in order.
	ids []string
}

// noContinueFiresCases are the rows upstream reports on.
var noContinueFiresCases = []noContinueCase{
	{
		name:    "invalid-0",
		source:  "var sum = 0, i; for(i = 0; i < 10; i++){ if(i <= 5) { continue; } sum += i; }",
		options: "",
		ids:     []string{"unexpected"},
	},
	{
		name:    "invalid-1",
		source:  "var sum = 0, i; myLabel: for(i = 0; i < 10; i++){ if(i <= 5) { continue myLabel; } sum += i; }",
		options: "",
		ids:     []string{"unexpected"},
	},
	{
		name:    "invalid-2",
		source:  "var sum = 0, i = 0; while(i < 10) { if(i <= 5) { i++; continue; } sum += i; i++; }",
		options: "",
		ids:     []string{"unexpected"},
	},
	{
		name:    "invalid-3",
		source:  "var sum = 0, i = 0; myLabel: while(i < 10) { if(i <= 5) { i++; continue myLabel; } sum += i; i++; }",
		options: "",
		ids:     []string{"unexpected"},
	},
}

// noContinueSilentCases are the rows upstream is clean on.
var noContinueSilentCases = []noContinueCase{
	{
		name:    "valid-0",
		source:  "var sum = 0, i; for(i = 0; i < 10; i++){ if(i > 5) { sum += i; } }",
		options: "",
	},
	{
		name:    "valid-1",
		source:  "var sum = 0, i = 0; while(i < 10) { if(i > 5) { sum += i; } i++; }",
		options: "",
	},
}

// decodedNoContinue routes a row's raw JSON through the rule's own exported decoder.
// no-continue has no option surface (upstream's schema is []), so there is no decoder to route
// through and the harness is handed nil, which is what the config layer passes for a bare "error".
func decodedNoContinue(t *testing.T, raw string) any {
	t.Helper()
	if raw != "" {
		t.Fatalf("no-continue takes no options, got %q", raw)
	}
	return nil
}

func TestNoContinueFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range noContinueFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoContinue, noContinueFile, testCase.source,
				decodedNoContinue(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

func TestNoContinueStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range noContinueSilentCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoContinue, noContinueFile, testCase.source,
				decodedNoContinue(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}
