package typescript

import (
	"sort"
	"strconv"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const noConfusingNonNullAssertionFile = "/repository/source/Confusing.ts"

func noConfusingNonNullAssertionCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoConfusingNonNullAssertionStaysSilent is upstream's eight passing cases verbatim, plus
// shapes upstream does not write whose silence was measured against the installed 8.67.0 build.
//
// Four of upstream's eight are parenthesized, which is not decoration: they are the whole of its
// `tokenAfterLeft !== ')'` test, and they are what keeps this port from reporting a left operand
// that merely CONTAINS an assertion.
func TestNoConfusingNonNullAssertionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		// Upstream's valid list, byte for byte.
		"a == b!;",
		"a = b!;",
		"a !== b;",
		"a != b;",
		"(a + b!) == c;",
		"(a + b!) = c;",
		"(a + b!) in c;",
		"(a || b!) instanceof c;",

		// Measured silent on the installed build. Each would report under a plausible wrong port,
		// and none is expressible from upstream's corpus.

		// The operators NOT in the confusing set, which is the half of the set definition the
		// corpus never exercises. `a! != b` is the shape a naive reading flags first and it is
		// silent, because the assertion and the operator are already spelled apart.
		"a! != b;\n",
		"a! !== b;\n",
		"a! < b;\n",
		"a! += b;\n",

		// A bang that is the last CHARACTER of the left operand without being its last TOKEN.
		// This is the only way the text test could diverge from upstream's token test, and the
		// parser does not produce it: the closing quote or bracket is the final character in each.
		"'a!' == b;\n",
		"`a!` == b;\n",
		"x[`!`] == b;\n",

		// `a!!== b` parses as `a!` with the operator `!==`, so the left text DOES end in a bang
		// and the operator test is what declines it. This is the one input where the two halves of
		// the judgment could disagree, and it is pinned so a change to either is visible.
		"a!!== b;\n",

		// A bang inside a comment beside the operand contributes nothing, because the range is the
		// trimmed token range rather than Pos() to End().
		"a /* ! */ == b;\n",
	}
	for index, sourceText := range cases {
		t.Run(noConfusingNonNullAssertionCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoConfusingNonNullAssertion,
				noConfusingNonNullAssertionFile, sourceText))
		})
	}
}

// TestNoConfusingNonNullAssertionFires is upstream's eleven reporting cases verbatim, with their
// message ids, their rendered message text, their spans, and every suggestion each one offers.
//
// The suggestions are asserted by APPLYING them and comparing the rewritten source against
// upstream's own `output` field, because rule_testing has no ExpectFixedSource equivalent for a
// suggestion. A suggestion asserted only by its message id is a repair nothing checks, and this
// rule's wrap suggestion writes two separate edits, so a single-edit applier would pass it while
// producing the wrong text.
func TestNoConfusingNonNullAssertionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText      string
		wantId          string
		wantMessage     string
		wantSpan        string
		wantSuggestions []struct {
			id          string
			description string
			rewrite     string
		}
	}{
		{
			sourceText:  "a! == b;",
			wantId:      "confusingEqual",
			wantMessage: "Confusing combination of non-null assertion and equality test like `a! == b`, which looks very similar to `a !== b`.",
			wantSpan:    "a! == b",
			wantSuggestions: []struct {
				id          string
				description string
				rewrite     string
			}{
				{id: "notNeedInEqualTest", description: "Remove unnecessary non-null assertion (!) in equality test.", rewrite: "a == b;"},
			},
		},
		{
			sourceText:  "a! === b;",
			wantId:      "confusingEqual",
			wantMessage: "Confusing combination of non-null assertion and equality test like `a! == b`, which looks very similar to `a !== b`.",
			wantSpan:    "a! === b",
			wantSuggestions: []struct {
				id          string
				description string
				rewrite     string
			}{
				{id: "notNeedInEqualTest", description: "Remove unnecessary non-null assertion (!) in equality test.", rewrite: "a === b;"},
			},
		},
		{
			sourceText:  "a + b! == c;",
			wantId:      "confusingEqual",
			wantMessage: "Confusing combination of non-null assertion and equality test like `a! == b`, which looks very similar to `a !== b`.",
			wantSpan:    "a + b! == c",
			wantSuggestions: []struct {
				id          string
				description string
				rewrite     string
			}{
				{id: "wrapUpLeft", description: "Wrap the left-hand side in parentheses to avoid confusion with \"==\" operator.", rewrite: "(a + b!) == c;"},
			},
		},
		{
			sourceText:  "(obj = new new OuterObj().InnerObj).Name! == c;",
			wantId:      "confusingEqual",
			wantMessage: "Confusing combination of non-null assertion and equality test like `a! == b`, which looks very similar to `a !== b`.",
			wantSpan:    "(obj = new new OuterObj().InnerObj).Name! == c",
			wantSuggestions: []struct {
				id          string
				description string
				rewrite     string
			}{
				{id: "notNeedInEqualTest", description: "Remove unnecessary non-null assertion (!) in equality test.", rewrite: "(obj = new new OuterObj().InnerObj).Name == c;"},
			},
		},
		{
			sourceText:  "(a==b)! ==c;",
			wantId:      "confusingEqual",
			wantMessage: "Confusing combination of non-null assertion and equality test like `a! == b`, which looks very similar to `a !== b`.",
			wantSpan:    "(a==b)! ==c",
			wantSuggestions: []struct {
				id          string
				description string
				rewrite     string
			}{
				{id: "notNeedInEqualTest", description: "Remove unnecessary non-null assertion (!) in equality test.", rewrite: "(a==b) ==c;"},
			},
		},
		{
			sourceText:  "a! = b;",
			wantId:      "confusingAssign",
			wantMessage: "Confusing combination of non-null assertion and assignment like `a! = b`, which looks very similar to `a != b`.",
			wantSpan:    "a! = b",
			wantSuggestions: []struct {
				id          string
				description string
				rewrite     string
			}{
				{id: "notNeedInAssign", description: "Remove unnecessary non-null assertion (!) in assignment left-hand side.", rewrite: "a = b;"},
			},
		},
		{
			sourceText:  "(obj = new new OuterObj().InnerObj).Name! = c;",
			wantId:      "confusingAssign",
			wantMessage: "Confusing combination of non-null assertion and assignment like `a! = b`, which looks very similar to `a != b`.",
			wantSpan:    "(obj = new new OuterObj().InnerObj).Name! = c",
			wantSuggestions: []struct {
				id          string
				description string
				rewrite     string
			}{
				{id: "notNeedInAssign", description: "Remove unnecessary non-null assertion (!) in assignment left-hand side.", rewrite: "(obj = new new OuterObj().InnerObj).Name = c;"},
			},
		},
		{
			sourceText:  "(a=b)! =c;",
			wantId:      "confusingAssign",
			wantMessage: "Confusing combination of non-null assertion and assignment like `a! = b`, which looks very similar to `a != b`.",
			wantSpan:    "(a=b)! =c",
			wantSuggestions: []struct {
				id          string
				description string
				rewrite     string
			}{
				{id: "notNeedInAssign", description: "Remove unnecessary non-null assertion (!) in assignment left-hand side.", rewrite: "(a=b) =c;"},
			},
		},
		{
			sourceText:  "a! in b;",
			wantId:      "confusingOperator",
			wantMessage: "Confusing combination of non-null assertion and `in` operator like `a! in b`, which might be misinterpreted as `!(a in b)`.",
			wantSpan:    "a! in b",
			wantSuggestions: []struct {
				id          string
				description string
				rewrite     string
			}{
				{id: "notNeedInOperator", description: "Remove possibly unnecessary non-null assertion (!) in the left operand of the `in` operator.", rewrite: "a in b;"},
				{id: "wrapUpLeft", description: "Wrap the left-hand side in parentheses to avoid confusion with \"in\" operator.", rewrite: "(a!) in b;"},
			},
		},
		{
			sourceText:  "\na !in b;\n      ",
			wantId:      "confusingOperator",
			wantMessage: "Confusing combination of non-null assertion and `in` operator like `a! in b`, which might be misinterpreted as `!(a in b)`.",
			wantSpan:    "a !in b",
			wantSuggestions: []struct {
				id          string
				description string
				rewrite     string
			}{
				{id: "notNeedInOperator", description: "Remove possibly unnecessary non-null assertion (!) in the left operand of the `in` operator.", rewrite: "\na in b;\n      "},
				{id: "wrapUpLeft", description: "Wrap the left-hand side in parentheses to avoid confusion with \"in\" operator.", rewrite: "\n(a !)in b;\n      "},
			},
		},
		{
			sourceText:  "a! instanceof b;",
			wantId:      "confusingOperator",
			wantMessage: "Confusing combination of non-null assertion and `instanceof` operator like `a! instanceof b`, which might be misinterpreted as `!(a instanceof b)`.",
			wantSpan:    "a! instanceof b",
			wantSuggestions: []struct {
				id          string
				description string
				rewrite     string
			}{
				{id: "notNeedInOperator", description: "Remove possibly unnecessary non-null assertion (!) in the left operand of the `instanceof` operator.", rewrite: "a instanceof b;"},
				{id: "wrapUpLeft", description: "Wrap the left-hand side in parentheses to avoid confusion with \"instanceof\" operator.", rewrite: "(a!) instanceof b;"},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(noConfusingNonNullAssertionCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoConfusingNonNullAssertion,
				noConfusingNonNullAssertionFile, testCase.sourceText)

			rule_testing.ExpectFindings(t, result, testCase.wantId)

			diagnostic := result.Diagnostics[0]
			if diagnostic.Message.Id != testCase.wantId {
				t.Fatalf("message id: expected %q, got %q", testCase.wantId, diagnostic.Message.Id)
			}

			// Equality rather than a substring test. `confusingOperator` splices the operator into
			// its text twice, and a substring predicate cannot see a wrong operator in either slot.
			if diagnostic.Message.Description != testCase.wantMessage {
				t.Fatalf("message: expected %q, got %q", testCase.wantMessage,
					diagnostic.Message.Description)
			}

			// rule_testing.Run does not trim, so the source on disk is the literal above.
			gotSpan := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}

			if len(diagnostic.Suggestions) != len(testCase.wantSuggestions) {
				t.Fatalf("suggestions: expected %d, got %d", len(testCase.wantSuggestions),
					len(diagnostic.Suggestions))
			}
			for suggestionIndex, want := range testCase.wantSuggestions {
				got := diagnostic.Suggestions[suggestionIndex]
				if got.Message.Id != want.id {
					t.Fatalf("suggestion %d id: expected %q, got %q", suggestionIndex, want.id,
						got.Message.Id)
				}
				// The rendered suggestion text, asserted by equality. `notNeedInOperator` and
				// `wrapUpLeft` both splice the operator into their descriptions, and a fixture
				// asserting only the id and the rewrite cannot see a wrong operator there: a
				// mutant hardcoding `in` into that slot survived the whole suite until this line
				// existed.
				if got.Message.Description != want.description {
					t.Fatalf("suggestion %d description: expected %q, got %q", suggestionIndex,
						want.description, got.Message.Description)
				}
				rewritten := applySuggestionFixes(testCase.sourceText, got)
				if rewritten != want.rewrite {
					t.Fatalf("suggestion %d rewrites to %q, wanted %q", suggestionIndex, rewritten,
						want.rewrite)
				}
			}
		})
	}
}

// applySuggestionFixes replays every edit a suggestion carries, highest offset first.
//
// Hand-rolled because rule_testing.ExpectFixedSource applies FIXES and this rule ships only
// suggestions, which the harness deliberately does not apply: a suggestion changes what the code
// means and a human has to choose it. Descending order so an earlier edit does not move a later
// one, which matters here because the wrap suggestion writes an opening and a closing paren as
// two separate edits.
func applySuggestionFixes(sourceText string, suggestion rule.Suggestion) string {
	ordered := make([]rule.Fix, len(suggestion.Fixes))
	copy(ordered, suggestion.Fixes)
	sort.SliceStable(ordered, func(a int, b int) bool {
		return ordered[a].Range.Pos() > ordered[b].Range.Pos()
	})
	for _, fix := range ordered {
		sourceText = sourceText[:fix.Range.Pos()] + fix.Text + sourceText[fix.Range.End():]
	}
	return sourceText
}
