package nexus

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const jsDocFile = "/repository/source/Thing.ts"

func TestConsistencyNoSingleLineJsDocFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"one line", "/** Does the thing. */\nexport const value = 1;\n"},
		{
			"multi line wrapper around one line of prose",
			"/**\n * Does the thing.\n */\nexport const value = 1;\n",
		},
		{
			"blank jsdoc lines are furniture",
			"/**\n *\n * Does the thing.\n *\n */\nexport const value = 1;\n",
		},
		{"inside a body", "export function thing() {\n    /** Does the thing. */\n    return 1;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoSingleLineJsDoc, jsDocFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "useSimpleComment")
		})
	}
}

func TestConsistencyNoSingleLineJsDocStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// A tag is the entire reason JSDoc exists, so a comment carrying one is JSDoc at any length.
		{"tag on one line", "/** @returns nothing */\nexport const value = 1;\n"},
		{"description and tag", "/**\n * Does the thing.\n * @returns nothing\n */\nexport const value = 1;\n"},
		{"tag only, multi line", "/**\n * @param name the name\n * @returns nothing\n */\nexport const value = 1;\n"},

		// Two lines of prose is a passage, and JSDoc is the right shape for it.
		{"two lines of prose", "/**\n * Does the thing.\n * Then does another thing.\n */\nexport const value = 1;\n"},

		// Not JSDoc at all.
		{"line comment", "// Does the thing.\nexport const value = 1;\n"},
		{"plain block comment", "/* Does the thing. */\nexport const value = 1;\n"},
		{"plain multi line block", "/*\n * Does the thing.\n */\nexport const value = 1;\n"},

		// An empty JSDoc has no description to promote, so there is nothing to say about it.
		{"empty jsdoc", "/** */\nexport const value = 1;\n"},

		{"no comments", "export const value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoSingleLineJsDoc, jsDocFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestConsistencyNoSingleLineJsDocFixes pins the fix text, because a fix that lands wrong is worse
// than no fix: the finding disappears and the damage is committed.
// TestConsistencyNoSingleLineJsDocFixesWithOnlyWhitespaceAfter keeps the repair for a comment that
// trailing spaces follow, which a line comment swallows harmlessly.
func TestConsistencyNoSingleLineJsDocFixesWithOnlyWhitespaceAfter(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ConsistencyNoSingleLineJsDoc, jsDocFile, "/** Does the thing. */  \t\nexport const value = 1;\n")
	rule_testing.ExpectFindings(t, result, "useSimpleComment")
	rule_testing.ExpectFixedSource(t, result, "// Does the thing.  \t\nexport const value = 1;\n")
}

func TestConsistencyNoSingleLineJsDocFixes(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ConsistencyNoSingleLineJsDoc, jsDocFile, "/** Does the thing. */\nexport const value = 1;\n")
	rule_testing.ExpectFindings(t, result, "useSimpleComment")

	fixes := result.Diagnostics[0].Fixes
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}
	if fixes[0].Text != "// Does the thing." {
		t.Fatalf("expected %q, got %q", "// Does the thing.", fixes[0].Text)
	}
}

// TestConsistencyNoSingleLineJsDocWithholdsUnsafeFixes covers the cases where the rule reports but
// must not repair. Each would produce a file that no longer says what the author wrote, and the last
// two, trpc's (#kq9vtva), a file that does not parse.
func TestConsistencyNoSingleLineJsDocWithholdsUnsafeFixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// A line comment holding a double-slash reads as commented-out code.
		{"description contains a double slash", "/** See // for details. */\nexport const value = 1;\n"},

		// Collapsing three source lines into one moves every position after it.
		{"multi line source", "/**\n * Does the thing.\n */\nexport const value = 1;\n"},

		// A line comment would swallow the code after it on the same line.
		{"code follows inside a call", "declare function create(options?: object): void;\ncreate(/** options */);\n"},
		{"code follows inside a pattern", "declare const values: number[];\nconst [a, b /**, c */] = values;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoSingleLineJsDoc, jsDocFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "useSimpleComment")
			if len(result.Diagnostics[0].Fixes) != 0 {
				t.Fatalf("expected no fix, got %d", len(result.Diagnostics[0].Fixes))
			}
		})
	}
}
