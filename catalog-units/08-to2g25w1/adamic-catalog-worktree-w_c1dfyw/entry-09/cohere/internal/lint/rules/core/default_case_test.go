package core

import (
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// defaultCaseFile is where the fixtures pretend to live.
//
// A `.ts` name: this rule has no file gate, and nothing in its corpus is JSX.
const defaultCaseFile = "/repository/source/DefaultCase.ts"

// defaultCaseCase is one row of upstream's corpus.
type defaultCaseCase struct {
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

// defaultCaseFiresCases are the rows upstream reports on.
var defaultCaseFiresCases = []defaultCaseCase{
	{
		name:    "invalid-0",
		source:  "switch (a) { case 1: break; }",
		options: "",
		ids:     []string{"missingDefaultCase"},
	},
	{
		name:    "invalid-1",
		source:  "switch (a) { \n // no default \n case 1: break;  }",
		options: "",
		ids:     []string{"missingDefaultCase"},
	},
	{
		name:    "invalid-2",
		source:  "switch (a) { case 1: break; \n // no default \n // nope \n  }",
		options: "",
		ids:     []string{"missingDefaultCase"},
	},
	{
		name:    "invalid-3",
		source:  "switch (a) { case 1: break; \n // no default \n }",
		options: "{\"commentPattern\": \"skipped default case\"}",
		ids:     []string{"missingDefaultCase"},
	},
	{
		name:    "invalid-4",
		source:  "switch (a) {\ncase 1: break; \n// default omitted intentionally \n// TODO: add default case \n}",
		options: "{\"commentPattern\": \"default omitted\"}",
		ids:     []string{"missingDefaultCase"},
	},
	{
		name:    "invalid-5",
		source:  "switch (a) {\ncase 1: break;\n}",
		options: "{\"commentPattern\": \".?\"}",
		ids:     []string{"missingDefaultCase"},
	},
}

// defaultCaseSilentCases are the rows upstream is clean on.
var defaultCaseSilentCases = []defaultCaseCase{
	{
		name:    "valid-0",
		source:  "switch (a) { case 1: break; default: break; }",
		options: "",
	},
	{
		name:    "valid-1",
		source:  "switch (a) { case 1: break; case 2: default: break; }",
		options: "",
	},
	{
		name:    "valid-2",
		source:  "switch (a) { case 1: break; default: break; \n //no default \n }",
		options: "",
	},
	{
		name:    "valid-3",
		source:  "switch (a) { \n    case 1: break; \n\n//oh-oh \n // no default\n }",
		options: "",
	},
	{
		name:    "valid-4",
		source:  "switch (a) { \n    case 1: \n\n// no default\n }",
		options: "",
	},
	{
		name:    "valid-5",
		source:  "switch (a) { \n    case 1: \n\n// No default\n }",
		options: "",
	},
	{
		name:    "valid-6",
		source:  "switch (a) { \n    case 1: \n\n// no deFAUlt\n }",
		options: "",
	},
	{
		name:    "valid-7",
		source:  "switch (a) { \n    case 1: \n\n// NO DEFAULT\n }",
		options: "",
	},
	{
		name:    "valid-8",
		source:  "switch (a) { \n    case 1: a = 4; \n\n// no default\n }",
		options: "",
	},
	{
		name:    "valid-9",
		source:  "switch (a) { \n    case 1: a = 4; \n\n/* no default */\n }",
		options: "",
	},
	{
		name:    "valid-10",
		source:  "switch (a) { \n    case 1: a = 4; break; break; \n\n// no default\n }",
		options: "",
	},
	{
		name:    "valid-11",
		source:  "switch (a) { // no default\n }",
		options: "",
	},
	{
		name:    "valid-12",
		source:  "switch (a) { }",
		options: "",
	},
	{
		name:    "valid-13",
		source:  "switch (a) { case 1: break; default: break; }",
		options: "{\"commentPattern\": \"default case omitted\"}",
	},
	{
		name:    "valid-14",
		source:  "switch (a) { case 1: break; \n // skip default case \n }",
		options: "{\"commentPattern\": \"^skip default\"}",
	},
	{
		name:    "valid-15",
		source:  "switch (a) { case 1: break; \n /*\nTODO:\n throw error in default case\n*/ \n }",
		options: "{\"commentPattern\": \"default\"}",
	},
	{
		name:    "valid-16",
		source:  "switch (a) { case 1: break; \n// \n }",
		options: "{\"commentPattern\": \".?\"}",
	},
}

// decodedDefaultCase routes a row's raw JSON through the rule's own exported decoder.
func decodedDefaultCase(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeDefaultCaseOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding options %q: %v", raw, err)
	}
	return decoded
}

func TestDefaultCaseFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range defaultCaseFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, DefaultCase, defaultCaseFile, testCase.source,
				decodedDefaultCase(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

func TestDefaultCaseStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range defaultCaseSilentCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, DefaultCase, defaultCaseFile, testCase.source,
				decodedDefaultCase(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestDefaultCaseCommentPatternIsAnchored covers a discrimination upstream's corpus cannot see.
//
// The default pattern is `/^no default$/iu`, anchored at both ends. Removing the anchors SURVIVES
// the whole 23-case corpus: every case that exercises the default pattern writes the comment
// exactly, and the two cases carrying a near-miss comment supply their own `commentPattern` option,
// under which both spellings match either way.
//
// Measured on the installed rule with a control that fires: `// no default` excuses the switch,
// while `// no default here` and `// say no default` both report. So an unanchored pattern would
// silently excuse two shapes upstream flags, in the direction that hides findings.
func TestDefaultCaseCommentPatternIsAnchored(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		reports bool
	}{
		// The control. Without it, a rule that had stopped excusing anything would pass the rest.
		{"control-exact-comment-excuses", "switch (a) { case 1: break;\n// no default\n}\n", false},
		{"trailing-text-does-not-excuse", "switch (a) { case 1: break;\n// no default here\n}\n", true},
		{"leading-text-does-not-excuse", "switch (a) { case 1: break;\n// say no default\n}\n", true},
		// Case-insensitivity is the other half of the flag, and the corpus does cover it; asserted
		// here beside the anchors so the whole pattern is pinned in one place.
		{"case-insensitive-still-excuses", "switch (a) { case 1: break;\n// NO DEFAULT\n}\n", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, DefaultCase, defaultCaseFile, testCase.source,
				decodedDefaultCase(t, ""))
			if testCase.reports {
				rule_testing.ExpectFindings(t, result, "missingDefaultCase")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}
