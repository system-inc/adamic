package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// octalEscapeFile is where the fixtures pretend to live.
const octalEscapeFile = "/repository/source/OctalEscape.ts"

// The corpus is ESLint's own, copied rather than rewritten.
//
// Every case below is verbatim from `eslint/tests/lib/rules/no-octal-escape.js`: 34 pass and 60
// fail, and every failing case names exactly one `octalEscapeSequence`, so one finding per input is
// stated by the corpus rather than assumed. The cases were lifted by evaluating that test file with
// its `RuleTester` stubbed and serialising the spec object it was handed, so the strings here are
// the ones upstream runs rather than a retyping of them, and the sequence beside each failing case
// is upstream's own `data.sequence`.
//
// Upstream runs this corpus at `ecmaVersion: 5` with `sourceType: "script"`, because a legacy octal
// escape is a syntax error in a module and ESLint would report a parse failure instead of the rule.
// Our parser has no such gate: measured over all 94 cases, 93 produce exactly one string literal in
// a `.ts` file and the parser cooks the escape rather than refusing it. The one that produces none
// is `var foo = /([abc]) \1/g;`, a regular expression, which upstream also declines through its
// `typeof value !== "string"` test. So the whole corpus is expressible here, and the correspondence
// was measured rather than hoped for.
//
// Copied because a fixture a porter invents encodes the same belief as the port, and the case that
// catches a bug is the one nobody would think to write.
func TestNoOctalEscapeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantSequence string
	}{
		{"var foo = \"foo \\01 bar\";", "01"},
		{"var foo = \"foo \\000 bar\";", "000"},
		{"var foo = \"foo \\377 bar\";", "377"},
		{"var foo = \"foo \\378 bar\";", "37"},
		{"var foo = \"foo \\37a bar\";", "37"},
		{"var foo = \"foo \\381 bar\";", "3"},
		{"var foo = \"foo \\3a1 bar\";", "3"},
		{"var foo = \"foo \\251 bar\";", "251"},
		{"var foo = \"foo \\258 bar\";", "25"},
		{"var foo = \"foo \\25a bar\";", "25"},
		{"var foo = \"\\3s51\";", "3"},
		{"var foo = \"\\77\";", "77"},
		{"var foo = \"\\78\";", "7"},
		{"var foo = \"\\5a\";", "5"},
		{"var foo = \"\\751\";", "75"},
		{"var foo = \"foo \\400 bar\";", "40"},
		{"var foo = \"\\t\\1\";", "1"},
		{"var foo = \"\\\\\\751\";", "75"},
		{"'\\0\\1'", "1"},
		{"'\\0 \\1'", "1"},
		{"'\\0\\01'", "01"},
		{"'\\0 \\01'", "01"},
		{"'\\0a\\1'", "1"},
		{"'\\0a\\01'", "01"},
		{"'\\0\\08'", "0"},
		{"'\\1'", "1"},
		{"'\\2'", "2"},
		{"'\\7'", "7"},
		{"'\\00'", "00"},
		{"'\\01'", "01"},
		{"'\\02'", "02"},
		{"'\\07'", "07"},
		{"'\\08'", "0"},
		{"'\\09'", "0"},
		{"'\\10'", "10"},
		{"'\\12'", "12"},
		{"' \\1'", "1"},
		{"'\\1 '", "1"},
		{"'a\\1'", "1"},
		{"'\\1a'", "1"},
		{"'a\\1a'", "1"},
		{"' \\01'", "01"},
		{"'\\01 '", "01"},
		{"'a\\01'", "01"},
		{"'\\01a'", "01"},
		{"'a\\01a'", "01"},
		{"'a\\08a'", "0"},
		{"'\\n\\1'", "1"},
		{"'\\n\\01'", "01"},
		{"'\\n\\08'", "0"},
		{"'\\\\\\1'", "1"},
		{"'\\\\\\01'", "01"},
		{"'\\\\\\08'", "0"},
		{"'\\\n\\1'", "1"},
		{"'\\01\\02'", "01"},
		{"'\\02\\01'", "02"},
		{"'\\01\\2'", "01"},
		{"'\\2\\01'", "2"},
		{"'\\08\\1'", "0"},
		{"'foo \\1 bar \\2'", "1"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoOctalEscape, octalEscapeFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "octalEscapeSequence")

			// The sequence is interpolated into the message, so the id assertion above cannot see
			// it. Upstream states a `data.sequence` per case and it is not always the digits a
			// reader would guess: `\378` names `37` and `\08` names `0`.
			//
			// Asserted as an exact prefix rather than with `strings.Contains`, because a
			// containment test on this value is weaker than the property it guards: the message
			// for `\01` contains the needle a message for `\0` would carry, so a rule capturing
			// one digit too few would stay green on half the corpus. The literal is typed here
			// rather than read from the rule's own constant, so both sides cannot move together.
			wantPrefix := "Do not use the octal escape `\\" + testCase.wantSequence + "`."
			if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, wantPrefix) {
				t.Errorf("message %q does not begin with %q", got, wantPrefix)
			}
		})
	}
}

// The clean cases are the whole discrimination, and they cluster into the distinctions the rule
// makes.
//
// The parity group (`'\\1'`, `'\\01'`, `'\\12'`, `'\\\0'`, `'\\\8'`, `'\0\\'`) is why the walk steps in
// escape-sized units rather than searching for a backslash: in each of these the backslash before
// the digits is itself escaped, so the digits are plain text. A rule scanning for a backslash
// followed by a digit reports all of them.
//
// The bare-null group (`'\0'` and its neighbours with text on either side) pins that `\0` alone is
// the legal null character rather than an octal escape. It is the reason the single-digit
// production is `[1-7]` and excludes zero, and the reason `0(?=[89])` has to be written separately.
//
// The decimal group (`'\8'`, `'\9'`, `'a\8a'`, `'\80'`, `'\81'`) is a different rule's surface:
// `\8` and `\9` are legacy decimal escapes, which no-nonoctal-decimal-escape covers. Reading them
// as octal would report on a sibling rule's cases. Note `'\80'` and `'\81'` are clean while `'\08'`
// reports, which looks backwards until you read the productions: the leading digit is what decides.
//
// The no-escape group (`'0'`, `'01'`, `'08'`, `'12'`) is digits with no backslash at all, and
// `'\x51'`, `'\a'`, `'\n'` are escapes that are not numeric.
func TestNoOctalEscapeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"var foo = \"\\x51\";",
		"var foo = \"foo \\\\251 bar\";",
		"var foo = /([abc]) \\1/g;",
		"var foo = '\\0';",
		"'\\0'",
		"'\\8'",
		"'\\9'",
		"'\\0 '",
		"' \\0'",
		"'a\\0'",
		"'\\0a'",
		"'a\\8a'",
		"'\\0\\8'",
		"'\\8\\0'",
		"'\\80'",
		"'\\81'",
		"'\\\\'",
		"'\\\\0'",
		"'\\\\08'",
		"'\\\\1'",
		"'\\\\01'",
		"'\\\\12'",
		"'\\\\\\0'",
		"'\\\\\\8'",
		"'\\0\\\\'",
		"'0'",
		"'1'",
		"'8'",
		"'01'",
		"'08'",
		"'80'",
		"'12'",
		"'\\a'",
		"'\\n'",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoOctalEscape, octalEscapeFile, sourceText))
		})
	}
}

// The cases below are ours. Each exists because reading our own code, running a mutation sweep, or
// driving the installed ESLint build raised a question the imported corpus does not answer.

// TestNoOctalEscapeProductionWidths pins how many digits each production takes.
//
// The corpus never writes a fourth octal digit, so a rule taking four instead of three passes all
// 94 imported cases. A mutation widening the greedy production's bound from three to four survived
// the whole suite, which is how these were found rather than by reading the pattern again.
//
// Each expectation below was measured against the installed ESLint build (version 10.8.1, driven
// through the Linter API) rather than derived from the regex:
//
//	'\0000'  names 000   the greedy production stops at three
//	'\1234'  names 123   the same, with a non-zero lead
//	'\3777'  names 377   the same, at the top of the greedy range
//	'\7777'  names 77    the [4-7] production takes exactly two, so a third digit is data
//	'\400'   names 40    the corpus's own case for that bound, restated here beside its neighbours
//
// The last two are why the two productions cannot be collapsed into one greedy scan: a value
// starting 4 through 7 overflows a byte with three digits, so upstream stops at two.
func TestNoOctalEscapeProductionWidths(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantSequence string
	}{
		{`'\0000'`, "000"},
		{`'\1234'`, "123"},
		{`'\3777'`, "377"},
		{`'\7777'`, "77"},
		{`'\400'`, "40"},
		{`'\77'`, "77"},
		{`'\7a'`, "7"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoOctalEscape, octalEscapeFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "octalEscapeSequence")
			wantPrefix := "Do not use the octal escape `\\" + testCase.wantSequence + "`."
			if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, wantPrefix) {
				t.Errorf("message %q does not begin with %q", got, wantPrefix)
			}
		})
	}
}

// TestNoOctalEscapeReportsTheWholeLiteral asserts where every finding points, which ExpectFindings
// cannot see.
//
// The rule carries no repair, so nothing downstream would notice a finding anchored on the wrong
// node, and upstream's corpus records no columns at all for this rule, only message data. Upstream
// reports the `Literal` node; measured against the installed build, `'\01'` reports columns 1
// through 5, which is the whole literal including both quotes.
//
// The offset cases matter more than the shape: a rule reporting the escape itself rather than the
// literal would pass every message assertion in this file.
func TestNoOctalEscapeReportsTheWholeLiteral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantStart  int
		wantText   string
	}{
		{`'\01'`, 0, `'\01'`},
		{`var foo = "foo \01 bar";`, 10, `"foo \01 bar"`},
		{`foo('\251');`, 4, `'\251'`},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoOctalEscape, octalEscapeFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if result.Diagnostics[0].Range.Pos() != testCase.wantStart || reported != testCase.wantText {
				t.Errorf("finding at [%d,%d) = %q, want start %d and %q",
					result.Diagnostics[0].Range.Pos(), result.Diagnostics[0].Range.End(),
					reported, testCase.wantStart, testCase.wantText)
			}
		})
	}
}

// TestNoOctalEscapeReportsOncePerLiteral pins the count and which escape wins.
//
// The corpus states this for four inputs and they are the sharpest cases in it: `'\01\02'` names
// `01` and `'\02\01'` names `02`, so the finding is the FIRST escape rather than the widest or the
// last. Those two are already above; these add the cases upstream has no equivalent of, where three
// escapes appear and where two literals in one file each report.
//
// A port reporting every escape would name the right id on every input and double the count on
// exactly these.
func TestNoOctalEscapeReportsOncePerLiteral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantFindings int
	}{
		{"three escapes in one literal", `'\1\2\3'`, 1},
		{"two literals, two findings", `var a = '\1'; var b = '\2';`, 2},
		{"one reporting literal beside a clean one", `var a = '\1'; var b = '\\1';`, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoOctalEscape, octalEscapeFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.wantFindings {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), testCase.wantFindings)
			}
		})
	}
}

// TestNoOctalEscapeDeclinesOtherLiteralKinds pins the surface.
//
// Upstream anchors on `Literal` and declines anything whose value is not a string, which is how
// `var foo = /([abc]) \1/g;` stays clean in its corpus: a regular expression's backreference is not
// an octal escape. Our tree gives regular expressions and templates their own kinds, so the
// discrimination lands in the listener map, and this pins that the translation did not widen it.
//
// The template cases have no upstream counterpart at all, because upstream never sees a template
// through this rule. They are here because our listener map is where the decision now lives, and a
// port adding a template kind to it would report on a template holding `\1`, which is a syntax error in a
// template and therefore code that cannot exist rather than code that is wrong.
func TestNoOctalEscapeDeclinesOtherLiteralKinds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a regular expression backreference", `var foo = /([abc]) \1/g;`},
		{"a regular expression octal-looking escape", `var foo = /\251/;`},
		{"a number", "var a = 251;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoOctalEscape, octalEscapeFile, testCase.sourceText))
		})
	}
}

// TestNoOctalEscapeMessage asserts the reported id and that the description is rendered rather than
// left as a template.
//
// The id literal is typed here rather than read from the rule's own constant, so a rename cannot
// move both sides at once and stay green.
func TestNoOctalEscapeMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoOctalEscape, octalEscapeFile, `'\01'`)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "octalEscapeSequence" {
		t.Errorf("message id = %q, want %q", got, "octalEscapeSequence")
	}
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, "Do not use the octal escape `\\01`.") {
		t.Errorf("message %q does not name the sequence", got)
	}
}
