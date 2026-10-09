package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const nonoctalDecimalEscapeFile = "/repository/source/Thing.ts"

// The cases are upstream's own corpus rather than cases invented here. That matters for this rule
// more than for most: every case is a sequence of backslashes and a digit, and a porter writing
// their own fixtures tends to write the parities they already understand.

func TestNoNonoctalDecimalEscapeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare escape", "export const A = '\\8';\n"},
		{"the nine form", "export const A = '\\9';\n"},
		{"in a double quoted string", "export const A = \"\\8\";\n"},
		{"after one character", "export const A = 'f\\9';\n"},
		{"after several characters", "export const A = 'foo\\9';\n"},
		{"with text on both sides", "export const A = 'foo\\8bar';\n"},
		{"after an escaped backslash", "export const A = '\\\\\\8';\n"},
		{"after two escaped backslashes", "export const A = '\\\\\\\\\\9';\n"},
		{"after an octal escape", "export const A = '\\1\\9';\n"},
		{"after a newline escape", "export const A = '\\n\\8';\n"},
		{"before an escaped backslash", "export const A = '\\8\\\\9';\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoNonoctalDecimalEscape, nonoctalDecimalEscapeFile, testCase.sourceText),
				"nonoctalDecimalEscape")
		})
	}
}

// The clean half is the whole rule. Every case here contains a backslash and an 8 or a 9, several
// contain the exact bytes of a violation, and none of them is one. A rule that searches for the
// pair rather than walking escape by escape fires on most of these.

func TestNoNonoctalDecimalEscapeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a plain digit", "export const A = '8';\n"},
		{"the nine digit", "export const A = '9';\n"},
		{"a digit after text", "export const A = 'foo8';\n"},
		{"an empty string", "export const A = '';\n"},
		{"no escapes at all", "export const A = 'foo';\n"},
		{"an escaped backslash then a digit", "export const A = '\\\\8';\n"},
		{"an escaped backslash then nine", "export const A = '\\\\9';\n"},
		{"two escaped backslashes then a digit", "export const A = '\\\\\\\\9';\n"},
		{"an octal escape", "export const A = '\\1';\n"},
		{"a seven escape", "export const A = '\\7';\n"},
		{"a null escape", "export const A = '\\0';\n"},
		{"an octal escape followed by eight", "export const A = '\\08';\n"},
		{"an octal escape followed by nine", "export const A = '\\19';\n"},
		{"a newline escape then a digit", "export const A = '\\t9';\n"},
		{"a hex escape", "export const A = '\\x99';\n"},
		{"an escaped backslash before a hex escape", "export const A = '\\\\\\x38';\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoNonoctalDecimalEscape, nonoctalDecimalEscapeFile, testCase.sourceText))
		})
	}
}

// Each escape is its own finding. A rule reporting once per literal agrees with every
// single-violation case above and undercounts the moment a string holds two, which is exactly the
// shape a scan that stops at the first match produces.
func TestNoNonoctalDecimalEscapeReportsEachEscape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       int
	}{
		{"the same digit twice", "export const A = '\\8\\8';\n", 2},
		{"both digits", "export const A = '\\9\\8';\n", 2},
		{"separated by an escaped backslash", "export const A = '\\8\\\\9';\n", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoNonoctalDecimalEscape, nonoctalDecimalEscapeFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.want {
				t.Fatalf("want %d findings, got %d: %v", testCase.want, len(result.Diagnostics), result.MessageIds())
			}
		})
	}
}

// The suggestions are the half a message id says nothing about, and they are where this rule's one
// real judgment lives.
//
// An ordinary violation offers two repairs because two things could have been meant: the digit, or
// a literal backslash before it. They are not interchangeable, which is why these are suggestions
// rather than fixes. An engine picking one unattended would silently change the string every time
// the other was intended.
func TestNoNonoctalDecimalEscapeOffersBothReadings(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoNonoctalDecimalEscape, nonoctalDecimalEscapeFile,
		"export const A = 'foo\\8bar';\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	suggestions := result.Diagnostics[0].Suggestions
	if len(suggestions) != 2 {
		t.Fatalf("want two suggestions, got %d", len(suggestions))
	}
	if text := suggestions[0].Fixes[0].Text; text != "8" {
		t.Fatalf("want the digit reading to write %q, got %q", "8", text)
	}
	// Two source characters before the digit: an escaped backslash. Upstream fixes
	// `'foo\\8bar'` to `'foo\\\\8bar'`, which is this same replacement written out.
	if text := suggestions[1].Fixes[0].Text; text != `\\8` {
		t.Fatalf("want the backslash reading to write %q, got %q", `\\8`, text)
	}
}

// A finding must carry no automatic fixes at all.
//
// This is asserted separately from the suggestion contents because it fails in the opposite
// direction: a rule that populated Fixes alongside Suggestions would still pass every assertion
// above, and the engine applies Fixes unattended. That would turn a repair only the author can
// choose into a silent rewrite of their string.
func TestNoNonoctalDecimalEscapeProposesNoAutomaticFix(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoNonoctalDecimalEscape, nonoctalDecimalEscapeFile,
		"export const A = '\\8';\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	if fixes := result.Diagnostics[0].Fixes; len(fixes) != 0 {
		t.Fatalf("want no automatic fixes, got %d", len(fixes))
	}
}

// `\0\8` is the case where the obvious repair is wrong: dropping the backslash yields `\08`, which
// is a legacy octal escape, so fixing one legacy escape would produce another. The suggestion set
// changes shape here rather than the rule declining to help.
func TestNoNonoctalDecimalEscapeAvoidsCreatingAnOctalEscape(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoNonoctalDecimalEscape, nonoctalDecimalEscapeFile,
		"export const A = '\\0\\8';\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	suggestions := result.Diagnostics[0].Suggestions
	if len(suggestions) != 3 {
		t.Fatalf("want three suggestions for the null-escape case, got %d", len(suggestions))
	}
	// Not one of them may write the bare digit, because that is the repair that creates `\08`.
	for index, suggestion := range suggestions {
		if suggestion.Fixes[0].Text == "8" {
			t.Fatalf("suggestion %d writes the bare digit, which would produce a legacy octal escape", index)
		}
	}
}

// The null-escape shape applies only to a real `\0`, not to `\01`, which is already a legacy octal
// escape of a different kind. Getting this wrong is invisible: the finding still fires and only the
// suggested repair is wrong.
func TestNoNonoctalDecimalEscapeTreatsOctalZeroAsOrdinary(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoNonoctalDecimalEscape, nonoctalDecimalEscapeFile,
		"export const A = '\\01\\8';\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	if count := len(result.Diagnostics[0].Suggestions); count != 2 {
		t.Fatalf("want the ordinary two suggestions after an octal escape, got %d", count)
	}
}

// Only a real `\0` gets the null-escape treatment, and every other preceding escape gets the
// ordinary two suggestions.
//
// A mutation sweep is what produced this test, and it is the case an id assertion cannot reach.
// Forcing the null-escape flag always-on leaves every fixture above green, because they assert
// which rule fired and not what it offered. The only visible symptom is `'\1\9'` gaining a third
// suggestion that rewrites a `\1` nobody asked about. A finding with a wrong repair still reads as
// a correct finding, so this needs asserting directly.
func TestNoNonoctalDecimalEscapeOffersTheNullShapeOnlyAfterNull(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       int
	}{
		{"after a null escape", "export const A = '\\0\\8';\n", 3},
		{"after an octal one escape", "export const A = '\\1\\9';\n", 2},
		{"after an octal zero-one escape", "export const A = '\\01\\8';\n", 2},
		{"after a newline escape", "export const A = '\\n\\8';\n", 2},
		{"with nothing before it", "export const A = '\\8';\n", 2},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoNonoctalDecimalEscape, nonoctalDecimalEscapeFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want one finding, got %d", len(result.Diagnostics))
			}
			if count := len(result.Diagnostics[0].Suggestions); count != testCase.want {
				t.Fatalf("want %d suggestions, got %d", testCase.want, count)
			}
		})
	}
}
