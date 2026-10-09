package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const emptyCharacterClassFile = "/repository/source/Thing.ts"

// The cases below are upstream's own, taken from oxlint's test corpus for this rule rather than
// invented here. That matters more than usual: this rule's whole difficulty is discriminating
// between bracket sequences that look alike, and a corpus written by the porter tends to contain
// the cases the porter already thought of. These contain the ones they did not.
func TestNoEmptyCharacterClassFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"at the end of a pattern", "export const Pattern = /^abc[]/;\n"},
		{"in the middle of a pattern", "export const Pattern = /foo[]bar/;\n"},
		// `[]]` is the class `[]` followed by a literal `]`, not a class containing `]`. A scanner
		// that treats the first `]` as closable-over reads this as non-empty and stays silent.
		{"followed by a literal bracket", "export const Pattern = /[]]/;\n"},
		// The leading `\[` is an escaped bracket and opens nothing, so the class is the later pair.
		{"after an escaped open bracket", "export const Pattern = /\\[[]/;\n"},
		{"after several escaped brackets", "export const Pattern = /\\[\\[\\]a-z[]/;\n"},
		{"with an unrelated flag set", "export const Pattern = /[]]/d;\n"},
		{"inside a call argument", "export function run(value: string) {\n    return value.match(/^abc[]/);\n}\n"},
		{"as a call receiver", "export function run(value: string) {\n    return /^abc[]/.test(value);\n}\n"},
		{"under the unicode flag", "export const Pattern = /[(]\\u{0}*[]/u;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoEmptyCharacterClass, emptyCharacterClassFile, testCase.sourceText),
				"unexpectedEmptyCharacterClass")
		})
	}
}

// Under `v`, classes nest. These are separated from the cases above because they exercise the
// recursion rather than the emptiness test, and because a scanner that ignores the `v` flag passes
// every case above while failing all of these.
func TestNoEmptyCharacterClassFiresInsideNestedClasses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare empty class", "export const Pattern = /[]/v;\n"},
		{"an empty class nested in one other", "export const Pattern = /[[]]/v;\n"},
		{"an empty class beside a non-empty sibling", "export const Pattern = /[[a][]]/v;\n"},
		{"an empty class nested three deep", "export const Pattern = /[a[[b[]c]]d]/v;\n"},
		{"the right operand of a difference", "export const Pattern = /[a--[]]/v;\n"},
		{"the left operand of a difference", "export const Pattern = /[[]--b]/v;\n"},
		{"the right operand of an intersection", "export const Pattern = /[a&&[]]/v;\n"},
		{"the left operand of an intersection", "export const Pattern = /[[]&&b]/v;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoEmptyCharacterClass, emptyCharacterClassFile, testCase.sourceText),
				"unexpectedEmptyCharacterClass")
		})
	}
}

// The clean half is where this rule is actually decided. Every case here contains brackets, and
// several contain the exact byte pair `[]`, so a rule that searches for that pair rather than
// parsing the pattern fires on all of them.
func TestNoEmptyCharacterClassStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a populated class", "export const Pattern = /^abc[a-zA-Z]/;\n"},
		{"no class at all", "export const Pattern = /^abc/;\n"},
		// The boundary case. `[^]` is one character away from `[]` and means the opposite: it
		// matches everything, including newlines, and is the shortest way to write that.
		{"a negated empty class", "export const Pattern = /[^]/;\n"},
		{"an escaped open bracket inside a class", "export const Pattern = /[\\[]/;\n"},
		{"an escaped close bracket inside a class", "export const Pattern = /[\\]]/;\n"},
		{"an escaped close bracket with flags", "export const Pattern = /[\\]]/uy;\n"},
		{"an escaped close bracket under dotall", "export const Pattern = /[\\]]/s;\n"},
		{"a class holding a range and a bracket", "export const Pattern = /[a-zA-Z\\[]/;\n"},
		{"a raw open bracket inside a class", "export const Pattern = /[[]/;\n"},
		{"a class opened and closed around an escape", "export const Pattern = /[\\[a-z[]]/;\n"},
		{"a class of escaped punctuation", "export const Pattern = /[\\-\\[\\]\\/\\{\\}\\(\\)\\*\\+\\?\\.\\\\^\\$\\|]/g;\n"},
		{"a pattern with no brackets at all", "export const Pattern = /\\s*:\\s*/gim;\n"},
		// `\[]` escapes the opening bracket, so this pattern contains the literal characters `[`
		// and `]` and opens no class whatsoever. This is the case a substring search cannot get
		// right, since the bytes `[]` are present and adjacent.
		{"an escaped bracket followed by a close", "export const Pattern = /\\[]/;\n"},
		// The constructor form is upstream's deliberate omission, not an oversight of ours.
		{"the RegExp constructor", "export const Pattern = new RegExp('^abc[]');\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoEmptyCharacterClass, emptyCharacterClassFile, testCase.sourceText))
		})
	}
}

// The `v` flag's clean cases, kept apart for the same reason their firing counterparts are: these
// all nest, and several nest a negated empty class, which is the pairing most likely to be reported
// by a recursion that checks emptiness before it checks negation.
func TestNoEmptyCharacterClassStaysSilentInsideNestedClasses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a nested negated empty class", "export const Pattern = /[[^]]/v;\n"},
		{"a nested escaped close bracket", "export const Pattern = /[[\\]]]/v;\n"},
		{"a nested escaped open bracket", "export const Pattern = /[[\\[]]/v;\n"},
		{"a difference of two populated operands", "export const Pattern = /[a--b]/v;\n"},
		{"an intersection of two populated operands", "export const Pattern = /[a&&b]/v;\n"},
		{"two populated classes side by side", "export const Pattern = /[[a][b]]/v;\n"},
		// `\q{}` is an empty string literal under `v`, which is a real element rather than the
		// absence of one, so the class holding it is not empty.
		{"a class holding an empty string literal", "export const Pattern = /[\\q{}]/v;\n"},
		{"a negated empty class as a difference operand", "export const Pattern = /[[^]--\\p{ASCII}]/v;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoEmptyCharacterClass, emptyCharacterClassFile, testCase.sourceText))
		})
	}
}

// The finding must cover the class itself and nothing else.
//
// This is a separate assertion because every test above passes with the range pointed at the whole
// literal, at the enclosing statement, or at the file. A message id proves a rule noticed something;
// it says nothing about where. Here that gap is load-bearing twice: a finding whose range starts
// before the literal's leading trivia lands on the preceding line and cannot be suppressed by any
// directive the author is able to write, and a finding spanning the whole literal is a different
// report from the one the gate we are matching produces.
func TestNoEmptyCharacterClassReportsTheClassAndNotTheLiteral(t *testing.T) {
	t.Parallel()

	sourceText := "export const Pattern = /foo[]bar/;\n"
	result := rule_testing.Run(t, NoEmptyCharacterClass, emptyCharacterClassFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	reported := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "[]" {
		t.Fatalf("want the finding to cover %q, got %q", "[]", reported)
	}
}

// A comment directly above the literal is what separates a usable finding from an unsuppressable
// one, and it is invisible in every count: the finding is still produced, still reads as real, and
// only fails when somebody tries to silence it. Loc.Pos() sits before leading trivia, so a range
// derived from it lands on the comment's line instead of the code's.
func TestNoEmptyCharacterClassReportsPastLeadingTrivia(t *testing.T) {
	t.Parallel()

	sourceText := "// a comment that must not be swallowed\nexport const Pattern = /foo[]bar/;\n"
	result := rule_testing.Run(t, NoEmptyCharacterClass, emptyCharacterClassFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	reported := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "[]" {
		t.Fatalf("want the finding to cover %q, got %q", "[]", reported)
	}
}

// Each empty class is its own finding. A rule that reports once per literal agrees with a
// single-class corpus on every case above and undercounts the moment a pattern holds two.
func TestNoEmptyCharacterClassReportsEachClassSeparately(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoEmptyCharacterClass, emptyCharacterClassFile,
		"export const Pattern = /[[][]]/v;\n")
	if len(result.Diagnostics) != 2 {
		t.Fatalf("want two findings, got %d: %v", len(result.Diagnostics), result.MessageIds())
	}
}

// Two literals in one file are two independent findings, each scanned against its own text.
//
// This is the case a rule that reads the file rather than the node gets wrong, and it is also the
// smallest statement that the pattern is taken per-literal at all.
func TestNoEmptyCharacterClassScansEachLiteralSeparately(t *testing.T) {
	t.Parallel()

	sourceText := "export const First = /[a]/g;\nexport const Second = /x[]/;\n"
	result := rule_testing.Run(t, NoEmptyCharacterClass, emptyCharacterClassFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d: %v", len(result.Diagnostics), result.MessageIds())
	}
	// Asserted against the source rather than against a literal offset, because an offset written
	// by hand is a number that agrees with today's fixture and stops meaning anything the moment
	// somebody adds a character to the line above it.
	secondLine := strings.Index(sourceText, "export const Second")
	if position := result.Diagnostics[0].Range.Pos(); position < secondLine {
		t.Fatalf("want the finding inside the second literal (at or past %d), got %d", secondLine, position)
	}
}

// A pattern the scanner cannot finish is skipped rather than reported. An unterminated class is a
// syntax error the parser has already refused, so a second complaint here would be noise on a file
// that does not compile. This asserts the rule survives the input at all, which is the half that
// matters: the scanner returns false mid-walk, and a callback that had already fired must not leave
// a finding behind on a pattern nobody can read.
func TestNoEmptyCharacterClassSkipsAnUnterminatedClass(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoEmptyCharacterClass, emptyCharacterClassFile,
		"export const Pattern = new RegExp('[abc');\n"))
}
