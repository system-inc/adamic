package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// guardForInFile is where the fixtures pretend to live.
const guardForInFile = "/repository/source/GuardForIn.ts"

// The corpus is ESLint's own, copied rather than rewritten.
//
// Every case below is verbatim from `eslint/tests/lib/rules/guard-for-in.js`, extracted by loading
// that file with the rule tester stubbed out, so the bytes are upstream's rather than retyped.
// 6 valid and 6 invalid, and every invalid case names exactly one
// `wrap`, so one finding per input is stated by the corpus rather than assumed.

// The clean cases are the whole discrimination and each one exempts for a different reason.
//
// A bodyless loop and an empty block have nothing to guard. A bare `if` as the body is the shape the
// rule is asking for. A block holding only an `if` is the same shape wearing braces. And the last two
// are the guard written as an early `continue`, in both the braced and unbraced spellings, which is
// the idiom the rule exists to permit rather than to flag.
func TestGuardForInStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"for (var x in o);",
		"for (var x in o) {}",
		"for (var x in o) if (x) f();",
		"for (var x in o) { if (x) { f(); } }",
		"for (var x in o) { if (x) continue; f(); }",
		"for (var x in o) { if (x) { continue; } f(); }",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, GuardForIn, guardForInFile, sourceText))
		})
	}
}

// The failing cases, and what separates each from its neighbour above.
//
// The first two are the sharpest: an `if` whose consequent is a block containing a `continue` and
// something else is NOT the early-exit idiom, in either order, because the other statement runs. The
// third and fourth guard part of the body and then run `g()` unguarded. The last two have no guard
// at all, braced and unbraced.
func TestGuardForInFires(t *testing.T) {
	t.Parallel()

	cases := []string{
		"for (var x in o) { if (x) { f(); continue; } g(); }",
		"for (var x in o) { if (x) { continue; f(); } g(); }",
		"for (var x in o) { if (x) { f(); } g(); }",
		"for (var x in o) { if (x) f(); g(); }",
		"for (var x in o) { foo() }",
		"for (var x in o) foo();",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, GuardForIn, guardForInFile, sourceText), "wrap")
		})
	}
}

// The span, which no message-id fixture above can see.
//
// Upstream reports on the loop STATEMENT rather than on the body or on the `for` keyword, so the
// finding covers the whole loop including its head. A port anchoring on the body would pass every
// assertion above while pointing somewhere the reader was not shown.
func TestGuardForInReportsOnTheWholeLoop(t *testing.T) {
	t.Parallel()

	const sourceText = "for (var x in o) foo();"

	result := rule_testing.Run(t, GuardForIn, guardForInFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}

	// `rule_testing.Run` does not trim, so the literal above and the file on disk agree and slicing the
	// literal is safe here. `RunTyped` would trim and this slice would be off by one.
	reported := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != sourceText {
		t.Fatalf("finding spans %q, wanted the whole loop %q", reported, sourceText)
	}

	if result.Diagnostics[0].Message.Id != "wrap" {
		t.Fatalf("message id is %q, wanted %q", result.Diagnostics[0].Message.Id, "wrap")
	}
}

// A case written from reading OUR code rather than upstream's, pinning the asymmetry the doc
// comment describes.
//
// The corpus has `{ if (x) { f(); } }` clean and `{ if (x) { f(); } g(); }` reporting, but it never
// isolates why: the single-statement arm accepts ANY consequent while the longer-block arm demands
// a bare `continue`. A port collapsing the two arms into one `continue` test would report the first
// of these, which upstream accepts. A port collapsing them the other way would accept the second.
func TestGuardForInSeparatesTheTwoBlockArms(t *testing.T) {
	t.Parallel()

	// One statement, a non-skipping consequent: clean, because the `if` is the whole body.
	rule_testing.ExpectClean(t, rule_testing.Run(t, GuardForIn, guardForInFile,
		"for (var x in o) { if (x) { f(); } }"))

	// Two statements, the same non-skipping consequent: reports, because `g()` runs unfiltered.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, GuardForIn, guardForInFile,
		"for (var x in o) { if (x) { f(); } g(); }"), "wrap")
}

// `for-of` never walks the prototype chain, so it has nothing to guard.
//
// Written because our parser gives `for-in` and `for-of` the SAME node struct
// (`ForInOrOfStatement`) and separates them only by kind. A listener keyed on the shared shape
// rather than on `KindForInStatement` would report every unguarded `for-of` in the tree, and no
// imported fixture can see it: upstream's corpus writes no `for-of` at all, because its parser
// gives the two different node types and the mistake is not available there.
func TestGuardForInDeclinesForOf(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, GuardForIn, guardForInFile, "for (var x of o) foo();"))

	// The control: the same body under `for-in` does report, so the silence above is about the
	// loop kind rather than about the fixture failing to reach the rule.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, GuardForIn, guardForInFile,
		"for (var x in o) foo();"), "wrap")
}

// Cases the corpus does not write, each measured against the INSTALLED eslint 10.8.1 build rather
// than read off the source.
//
// The corpus covers six clean shapes and six failing ones and leaves every neighbouring shape
// unstated. These are the ones where a defensible reading of the source gives the wrong answer, so
// each row is a place a port could drift silently. Driven with the ESLint Linter API on
// `sourceType: "script"`, `ecmaVersion: 2022`.
func TestGuardForInMatchesTheInstalledBuild(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantFindings int
	}{
		// A labeled continue is a `LabeledStatement`, not a `ContinueStatement`, so it is not the
		// bare skip and the trailing `g()` runs unfiltered. A port unwrapping labels to find the
		// continue inside would go silent here.
		{"a labeled continue is not a bare skip", "for (var x in o) { if (x) label: continue; g(); }", 1},

		// The guard nested one block deeper is not the leading statement of the loop body, so the
		// body starts with a `Block` rather than an `If`. Upstream does not look through it.
		{"a guard wrapped in its own block", "for (var x in o) { { if (x) continue; } g(); }", 1},

		// Only the CONSEQUENT is inspected. An `if` with an else whose consequent is a bare
		// continue still exempts, even though the else arm runs unfiltered work.
		{"an else arm beside a bare continue", "for (var x in o) { if (x) continue; else f(); g(); }", 0},

		// An empty statement inside a block is a statement. The block is not empty, its first
		// statement is not an `if`, so it reports -- unlike the bodyless `for (var x in o);` and
		// the empty `{}`, which the corpus has clean. Three spellings of "does nothing", two
		// verdicts.
		{"a block holding only an empty statement", "for (var x in o) { ; }", 1},

		// A continue nested one block deeper inside the consequent is not the bare skip: upstream
		// requires the consequent block's ONE statement to be a continue, and here it is a block.
		{"a continue nested inside the consequent", "for (var x in o) { if (x) { { continue; } } g(); }", 1},

		// The guard is not required to come FIRST for the single-statement arm, but for a longer
		// block it is: a statement before the guard runs on every key.
		{"a statement before the guard", "for (var x in o) { var y = 1; if (x) continue; }", 1},

		// Automatic semicolon insertion still produces a `ContinueStatement`, so the missing
		// semicolon changes nothing.
		{"a bare continue without its semicolon", "for (var x in o) { if (x) { continue } g(); }", 0},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, GuardForIn, guardForInFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.wantFindings {
				t.Fatalf("got %d findings, wanted %d (measured on eslint 10.8.1)",
					len(result.Diagnostics), testCase.wantFindings)
			}
		})
	}
}
