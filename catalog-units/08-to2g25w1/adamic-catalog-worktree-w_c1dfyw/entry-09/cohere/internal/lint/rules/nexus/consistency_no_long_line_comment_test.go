package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const longCommentFile = "/repository/source/Thing.ts"

// lineCommentRun builds a run of count line comments followed by a statement.
func lineCommentRun(count int) string {
	var builder strings.Builder
	for index := 0; index < count; index++ {
		builder.WriteString("// line ")
		builder.WriteString(string(rune('a' + index)))
		builder.WriteString("\n")
	}
	builder.WriteString("export const value = 1;\n")
	return builder.String()
}

func TestConsistencyNoLongLineCommentFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"five lines", lineCommentRun(5)},
		{"nine lines", lineCommentRun(9)},
		{
			"indented run inside a body",
			"export function thing() {\n" +
				"    // one\n    // two\n    // three\n    // four\n    // five\n" +
				"    return 1;\n}\n",
		},
		{
			"a blank comment line does not break the run",
			"// one\n// two\n//\n// four\n// five\nexport const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoLongLineComment, longCommentFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "longLineComment")
		})
	}
}

func TestConsistencyNoLongLineCommentStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		{"four lines is the threshold", longCommentFile, lineCommentRun(4)},
		{"one line", longCommentFile, lineCommentRun(1)},
		{"no comments", longCommentFile, "export const value = 1;\n"},

		// A block comment is already the shape this rule is asking for.
		{
			"a long block comment",
			longCommentFile,
			"/*\n * one\n * two\n * three\n * four\n * five\n * six\n */\nexport const value = 1;\n",
		},

		// A gap breaks the run, so neither half reaches five.
		{
			"blank source line splits the run",
			longCommentFile,
			"// one\n// two\n// three\n\n// four\n// five\n// six\nexport const value = 1;\n",
		},

		// Columns must match. A trailing comment must never join the passage above it.
		{
			"trailing comments do not stack",
			longCommentFile,
			"const a = 1; // one\nconst b = 2; // two\nconst c = 3; // three\nconst d = 4; // four\nconst e = 5; // five\n",
		},

		// The column check in isolation. Every one of these is alone on its line, so the
		// own-line test passes them all and only the differing indentation keeps them apart. The
		// case above cannot prove this: those comments are excluded for sharing a line with code,
		// so the column check never runs and could be deleted without the fixture noticing.
		{
			"a change in indentation breaks the run",
			longCommentFile,
			"// one\n// two\n  // three\n// four\n// five\nexport const value = 1;\n",
		},

		// Folding a directive into a block silently disables it.
		{
			"directives are never folded",
			longCommentFile,
			"// eslint-disable-next-line no-console\n// eslint-disable-next-line no-alert\n" +
				"// @ts-expect-error one\n// @ts-expect-error two\n// prettier-ignore\nexport const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoLongLineComment, testCase.fileName, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestConsistencyNoLongLineCommentFixes pins the rewrite, because a fix that lands wrong destroys
// the comment it was repairing and the finding disappears with it.
// Generated code is held to the rule like hand-written code (Kirk's ruling on generated code, #c076xbg):
// the generator writes a block comment, so a finding here is a defect in the generator. The rule used
// to skip these files.
func TestConsistencyNoLongLineCommentHoldsGeneratedFilesToTheRule(t *testing.T) {
	t.Parallel()
	for _, fileName := range []string{"/repository/source/generated/Thing.ts", "/repository/source/Thing.generated.ts"} {
		result := rule_testing.Run(t, ConsistencyNoLongLineComment, fileName, lineCommentRun(9))
		rule_testing.ExpectFindings(t, result, "longLineComment")
	}
}

func TestConsistencyNoLongLineCommentFixes(t *testing.T) {
	t.Parallel()

	sourceText := "// one\n// two\n// three\n// four\n// five\nexport const value = 1;\n"
	result := rule_testing.Run(t, ConsistencyNoLongLineComment, longCommentFile, sourceText)
	rule_testing.ExpectFindings(t, result, "longLineComment")

	fixes := result.Diagnostics[0].Fixes
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}

	wantText := "/*\n * one\n * two\n * three\n * four\n * five\n */"
	if fixes[0].Text != wantText {
		t.Fatalf("expected:\n%s\ngot:\n%s", wantText, fixes[0].Text)
	}

	// The fix must replace exactly the run and nothing else, or it eats the code after it.
	replaced := sourceText[fixes[0].Range.Pos():fixes[0].Range.End()]
	wantReplaced := "// one\n// two\n// three\n// four\n// five"
	if replaced != wantReplaced {
		t.Fatalf("fix replaces %q, expected %q", replaced, wantReplaced)
	}
}

// TestConsistencyNoLongLineCommentFixPreservesIndentation checks the indented case separately,
// since a fix that resets indentation reformats code it was not asked to touch.
func TestConsistencyNoLongLineCommentFixPreservesIndentation(t *testing.T) {
	t.Parallel()

	sourceText := "export function thing() {\n" +
		"    // one\n    // two\n    // three\n    // four\n    // five\n" +
		"    return 1;\n}\n"
	result := rule_testing.Run(t, ConsistencyNoLongLineComment, longCommentFile, sourceText)
	rule_testing.ExpectFindings(t, result, "longLineComment")

	wantText := "/*\n     * one\n     * two\n     * three\n     * four\n     * five\n     */"
	if result.Diagnostics[0].Fixes[0].Text != wantText {
		t.Fatalf("expected:\n%s\ngot:\n%s", wantText, result.Diagnostics[0].Fixes[0].Text)
	}
}

// TestConsistencyNoLongLineCommentWithholdsUnsafeFix covers the one case the fix must refuse: a
// star-slash in the text would close the block early and change what the rest of the run means.
func TestConsistencyNoLongLineCommentWithholdsUnsafeFix(t *testing.T) {
	t.Parallel()

	sourceText := "// one\n// two\n// a */ sequence\n// four\n// five\nexport const value = 1;\n"
	result := rule_testing.Run(t, ConsistencyNoLongLineComment, longCommentFile, sourceText)
	rule_testing.ExpectFindings(t, result, "longLineComment")

	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("expected no fix, got %d", len(result.Diagnostics[0].Fixes))
	}
}
