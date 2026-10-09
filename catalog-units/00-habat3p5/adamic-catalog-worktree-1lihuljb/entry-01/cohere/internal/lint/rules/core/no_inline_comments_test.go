package core

import (
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// noInlineCommentsFile is where the fixtures pretend to live.
//
// A `.ts` name: this rule has no file gate, and nothing in its corpus is JSX.
const noInlineCommentsFile = "/repository/source/NoInlineComments.tsx"

// noInlineCommentsCase is one row of upstream's corpus.
type noInlineCommentsCase struct {
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

// noInlineCommentsFiresCases are the rows upstream reports on.
var noInlineCommentsFiresCases = []noInlineCommentsCase{
	{
		name:    "invalid-0",
		source:  "var a = 1; /*A block comment inline after code*/",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-1",
		source:  "/*A block comment inline before code*/ var a = 2;",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-2",
		source:  "/* something */ var a = 2;",
		options: "{\"ignorePattern\": \"otherthing\"}",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-3",
		source:  "var a = 3; //A comment inline with code",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-4",
		source:  "var a = 3; // someday use eslint-disable-line here",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-5",
		source:  "var a = 3; // other line comment",
		options: "{\"ignorePattern\": \"something\"}",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-6",
		source:  "var a = 4;\n/**A\n * block\n * comment\n * inline\n * between\n * code*/ var foo = a;",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-7",
		source:  "var a = \n{/**/}",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-8",
		source:  "var a = (\n                <div>{/* comment */}</div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-9",
		source:  "var a = (\n                <div>{// comment\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-10",
		source:  "var a = (\n                <div>{/* comment */\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-11",
		source:  "var a = (\n                <div>{/*\n                       * comment\n                       */\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-12",
		source:  "var a = (\n                <div>{/*\n                       * comment\n                       */}\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-13",
		source:  "var a = (\n                <div>{/*\n                       * comment\n                       */}</div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-14",
		source:  "var a = (\n                <div>\n                {/*\n                  * comment\n                  */}</div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-15",
		source:  "var a = (\n                <div>\n                {\n                 /*\n                  * comment\n                  */}</div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-16",
		source:  "var a = (\n                <div>\n                {\n                /* comment */}</div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-17",
		source:  "var a = (\n                <div>\n                {b/* comment */}\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-18",
		source:  "var a = (\n                <div>\n                {/* comment */b}\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-19",
		source:  "var a = (\n                <div>\n                {// comment\n                    b\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-20",
		source:  "var a = (\n                <div>\n                {/* comment */\n                    b\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-21",
		source:  "var a = (\n                <div>\n                {/*\n                  * comment\n                  */\n                    b\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-22",
		source:  "var a = (\n                <div>\n                {\n                    b// comment\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-23",
		source:  "var a = (\n                <div>\n                {\n                    /* comment */b\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-24",
		source:  "var a = (\n                <div>\n                {\n                    b/* comment */\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-25",
		source:  "var a = (\n                <div>\n                {\n                    b\n                /*\n                 * comment\n                 */}\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-26",
		source:  "var a = (\n                <div>\n                {\n                    b\n                /* comment */}\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-27",
		source:  "var a = (\n                <div>\n                {\n                    { /* this is an empty object literal, not braces for js code! */ }\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-28",
		source:  "var a = (\n                <div>\n                {\n                    {// comment\n                    }\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-29",
		source:  "var a = (\n                <div>\n                {\n                    {\n                    /* comment */}\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment"},
	},
	{
		name:    "invalid-30",
		source:  "var a = (\n                <div>\n                { /* two comments on the same line... */ /* ...are not allowed, same as with a non-JSX code */}\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment", "unexpectedInlineComment"},
	},
	{
		name:    "invalid-31",
		source:  "var a = (\n                <div>\n                {\n                    /* overlapping\n                    */ /*\n                       lines */\n                }\n                </div>\n            )",
		options: "",
		ids:     []string{"unexpectedInlineComment", "unexpectedInlineComment"},
	},
}

// noInlineCommentsSilentCases are the rows upstream is clean on.
var noInlineCommentsSilentCases = []noInlineCommentsCase{
	{
		name:    "valid-0",
		source:  "// A valid comment before code\nvar a = 1;",
		options: "",
	},
	{
		name:    "valid-1",
		source:  "var a = 2;\n// A valid comment after code",
		options: "",
	},
	{
		name:    "valid-2",
		source:  "// A solitary comment",
		options: "",
	},
	{
		name:    "valid-3",
		source:  "var a = 1; // eslint-disable-line no-debugger",
		options: "",
	},
	{
		name:    "valid-4",
		source:  "var a = 1; /* eslint-disable-line no-debugger */",
		options: "",
	},
	{
		name:    "valid-5",
		source:  "foo(); /* global foo */",
		options: "",
	},
	{
		name:    "valid-6",
		source:  "foo(); /* globals foo */",
		options: "",
	},
	{
		name:    "valid-7",
		source:  "var foo; /* exported foo */",
		options: "",
	},
	{
		name:    "valid-8",
		source:  "var a = (\n            <div>\n            {/*comment*/}\n            </div>\n        )",
		options: "",
	},
	{
		name:    "valid-9",
		source:  "var a = (\n            <div>\n            { /* comment */ }\n            <h1>Some heading</h1>\n            </div>\n        )",
		options: "",
	},
	{
		name:    "valid-10",
		source:  "var a = (\n            <div>\n            {// comment\n            }\n            </div>\n        )",
		options: "",
	},
	{
		name:    "valid-11",
		source:  "var a = (\n            <div>\n            { // comment\n            }\n            </div>\n        )",
		options: "",
	},
	{
		name:    "valid-12",
		source:  "var a = (\n            <div>\n            {/* comment 1 */\n            /* comment 2 */}\n            </div>\n        )",
		options: "",
	},
	{
		name:    "valid-13",
		source:  "var a = (\n            <div>\n            {/*\n              * comment 1\n              */\n             /*\n              * comment 2\n              */}\n            </div>\n        )",
		options: "",
	},
	{
		name:    "valid-14",
		source:  "var a = (\n            <div>\n            {/*\n               multi\n               line\n               comment\n            */}\n            </div>\n        )",
		options: "",
	},
	{
		name:    "valid-15",
		source:  "import(/* webpackChunkName: \"my-chunk-name\" */ './locale/en');",
		options: "{\"ignorePattern\": \"(?:webpackChunkName):\\\\s.+\"}",
	},
	{
		name:    "valid-16",
		source:  "var foo = 2; // Note: This comment is legal.",
		options: "{\"ignorePattern\": \"Note: \"}",
	},
}

// decodedNoInlineComments routes a row's raw JSON through the rule's own exported decoder.
func decodedNoInlineComments(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeNoInlineCommentsOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding options %q: %v", raw, err)
	}
	return decoded
}

func TestNoInlineCommentsFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range noInlineCommentsFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoInlineComments, noInlineCommentsFile, testCase.source,
				decodedNoInlineComments(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

func TestNoInlineCommentsStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range noInlineCommentsSilentCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoInlineComments, noInlineCommentsFile, testCase.source,
				decodedNoInlineComments(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoInlineCommentsDirectiveTestDiffersByCommentKind covers an asymmetry the corpus cannot see.
//
// Upstream's `isDirectiveComment` uses two different tests:
//
//	Line comment    comment.startsWith("eslint-")
//	Block comment   /^(?:eslint[- ]|(?:globals?|exported) )/
//
// So a block `/* global foo */` is a directive and a line `// global foo` is not. Collapsing the two
// into the block pattern SURVIVES the whole 49-case corpus, because it contains three block-form
// `global`/`exported` cases and ZERO line-form ones. The corpus tests the shape it happens to use.
//
// Measured on the installed rule with two controls that fire: `/* global foo */` is clean,
// `// global foo` reports, `// exported a` reports, `// eslint-disable-line no-debugger` is clean,
// and `// eslint disable` -- a SPACE where the line test requires a hyphen -- reports.
func TestNoInlineCommentsDirectiveTestDiffersByCommentKind(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		reports bool
	}{
		// The controls: the shapes the corpus does cover, so a rule that stopped exempting
		// anything, or started exempting everything, fails here too.
		{"control-block-global-is-a-directive", "foo(); /* global foo */\n", false},
		{"control-line-eslint-hyphen-is-a-directive", "var a = 1; // eslint-disable-line no-debugger\n", false},

		// The shapes it does not cover, which is what the mutant exploited.
		{"line-global-is-not-a-directive", "foo(); // global foo\n", true},
		{"line-exported-is-not-a-directive", "var a = 1; // exported a\n", true},
		{"line-eslint-space-is-not-a-directive", "var a = 1; // eslint disable\n", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoInlineComments, noInlineCommentsFile,
				testCase.source, decodedNoInlineComments(t, ""))
			if testCase.reports {
				rule_testing.ExpectFindings(t, result, "unexpectedInlineComment")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}
