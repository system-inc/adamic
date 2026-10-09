package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// extraLabelFile is where the fixtures pretend to live.
const extraLabelFile = "/repository/source/ExtraLabel.ts"

// The corpus is ESLint's own, imported verbatim from
// eslint/tests/lib/rules/no-extra-label.js: 15 pass, 19 fail. Every failing case
// reports exactly once, and 15 of them carry an output, so the fix vectors are asserted
// separately below rather than folded into the count.
//
// The cases were extracted by loading upstream's tester module with a stubbed RuleTester and
// serialising what it received, then rendering these literals from that JSON, so no case here
// was retyped and no escape sequence was ever hand-written. The generator refuses any case
// containing a byte outside printable ASCII, which is what would let a cooked escape through.
func TestNoExtraLabelFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a labeled loop whose break names it", "A: while (a) break A;"},
		{"a labeled loop reached past an inner labeled block", "A: while (a) { B: { continue A; } }"},
		{"the innermost of three, where only the redundant one reports", "X: while (x) { A: while (a) { B: { break A; break B; continue X; } } }"},
		{"a labeled do while", "A: do { break A; } while (a);"},
		{"a labeled for", "A: for (;;) { break A; }"},
		{"a labeled for in", "A: for (a in obj) { break A; }"},
		{"a labeled for of", "A: for (a of ary) { break A; }"},
		{"a labeled switch", "A: switch (a) { case 0: break A; }"},
		{"a labeled switch inside another labeled loop", "X: while (x) { A: switch (a) { case 0: break A; } }"},
		{"a labeled loop inside a labeled switch", "X: switch (a) { case 0: A: while (b) break A; }"},
		{"only the break in the outer loop is redundant", "                A: while (true) {\n                    break A;\n                    while (true) {\n                        break A;\n                    }\n                }\n            "},
		{"a comment before the keyword still reports", "A: while(true) { /*comment*/break A; }"},
		{"a comment between break and its label", "A: while(true) { break/**/ A; }"},
		{"a comment between continue and its label", "A: while(true) { continue /**/ A; }"},
		{"a comment before the label with no space", "A: while(true) { break /**/A; }"},
		{"a comment butted against continue and its label", "A: while(true) { continue/**/A; }"},
		{"a comment after the label", "A: while(true) { continue A/*comment*/; }"},
		{"a line comment after the label", "A: while(true) { break A//comment\n }"},
		{"a block comment after the label before a newline", "A: while(true) { break A/*comment*/\nfoo() }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoExtraLabel, extraLabelFile, testCase.sourceText), "unexpected")
		})
	}
}

// The clean cases are the whole discrimination, and the scope stack is what they test.
//
// A rule reporting whenever a jump's label names SOME enclosing label reports every one of these,
// because in all fifteen the label does enclose the jump. What makes them clean is that something
// breakable sits between the jump and the label, so the label is the only thing that reaches it.
// The two continue cases pin the switch frame: a switch is breakable but cannot be continued, so
// upstream stops the search there and leaves the label alone rather than reporting a jump whose
// label is load bearing.
func TestNoExtraLabelStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a label on a bare break with no enclosing loop", "A: break A;"},
		{"a labeled block reached from an if", "A: { if (a) break A; }"},
		{"a labeled block reached past a loop", "A: { while (b) { break A; } }"},
		{"a labeled block reached past a switch", "A: { switch (b) { case 0: break A; } }"},
		{"two nested loops with unlabeled breaks", "A: while (a) { while (b) { break; } break; }"},
		{"an inner loop intercepting the plain break", "A: while (a) { while (b) { break A; } }"},
		{"an inner loop intercepting the plain continue", "A: while (a) { while (b) { continue A; } }"},
		{"a switch intercepting the break", "A: while (a) { switch (b) { case 0: break A; } }"},
		{"a switch that cannot be continued through", "A: while (a) { switch (b) { case 0: continue A; } }"},
		{"a loop inside a labeled switch", "A: switch (a) { case 0: while (b) { break A; } }"},
		{"a switch inside a labeled switch", "A: switch (a) { case 0: switch (b) { case 0: break A; } }"},
		{"a loop inside a labeled for", "A: for (;;) { while (b) { break A; } }"},
		{"a switch inside a labeled do while", "A: do { switch (b) { case 0: break A; break; } } while (a);"},
		{"a loop inside a labeled for in", "A: for (a in obj) { while (b) { break A; } }"},
		{"a switch inside a labeled for of", "A: for (a of ary) { switch (b) { case 0: break A; } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoExtraLabel, extraLabelFile, testCase.sourceText))
		})
	}
}

// The fix vectors, which are upstream's own output fields.
//
// These assert what the repair WRITES rather than that a finding appeared, which is the only
// assertion that can see a fixer repairing the right span with the wrong text. The comment cases
// are the interesting half: a comment before the keyword or after the label sits outside the
// removal span and the repair is still offered.
func TestNoExtraLabelFixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"a labeled loop whose break names it", "A: while (a) break A;", "A: while (a) break;"},
		{"a labeled loop reached past an inner labeled block", "A: while (a) { B: { continue A; } }", "A: while (a) { B: { continue; } }"},
		{"the innermost of three, where only the redundant one reports", "X: while (x) { A: while (a) { B: { break A; break B; continue X; } } }", "X: while (x) { A: while (a) { B: { break; break B; continue X; } } }"},
		{"a labeled do while", "A: do { break A; } while (a);", "A: do { break; } while (a);"},
		{"a labeled for", "A: for (;;) { break A; }", "A: for (;;) { break; }"},
		{"a labeled for in", "A: for (a in obj) { break A; }", "A: for (a in obj) { break; }"},
		{"a labeled for of", "A: for (a of ary) { break A; }", "A: for (a of ary) { break; }"},
		{"a labeled switch", "A: switch (a) { case 0: break A; }", "A: switch (a) { case 0: break; }"},
		{"a labeled switch inside another labeled loop", "X: while (x) { A: switch (a) { case 0: break A; } }", "X: while (x) { A: switch (a) { case 0: break; } }"},
		{"a labeled loop inside a labeled switch", "X: switch (a) { case 0: A: while (b) break A; }", "X: switch (a) { case 0: A: while (b) break; }"},
		{"only the break in the outer loop is redundant", "                A: while (true) {\n                    break A;\n                    while (true) {\n                        break A;\n                    }\n                }\n            ", "                A: while (true) {\n                    break;\n                    while (true) {\n                        break A;\n                    }\n                }\n            "},
		{"a comment before the keyword still reports", "A: while(true) { /*comment*/break A; }", "A: while(true) { /*comment*/break; }"},
		{"a comment after the label", "A: while(true) { continue A/*comment*/; }", "A: while(true) { continue/*comment*/; }"},
		{"a line comment after the label", "A: while(true) { break A//comment\n }", "A: while(true) { break//comment\n }"},
		{"a block comment after the label before a newline", "A: while(true) { break A/*comment*/\nfoo() }", "A: while(true) { break/*comment*/\nfoo() }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoExtraLabel, extraLabelFile, testCase.sourceText), testCase.wantSource)
		})
	}
}

// The four cases upstream reports and deliberately declines to fix.
//
// Each carries output: null upstream, which is a decision rather than an omission: a comment sits
// inside the span the removal would delete, so applying the repair unattended would destroy it.
// Asserting the source is UNCHANGED is the only thing that can see a port which repairs a case
// upstream refuses to touch, since the finding itself is identical either way.
func TestNoExtraLabelDeclinesToFixOverAComment(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a comment between break and its label", "A: while(true) { break/**/ A; }"},
		{"a comment between continue and its label", "A: while(true) { continue /**/ A; }"},
		{"a comment before the label with no space", "A: while(true) { break /**/A; }"},
		{"a comment butted against continue and its label", "A: while(true) { continue/**/A; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoExtraLabel, extraLabelFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unexpected")
			// The decline is asserted as the ABSENCE of a proposal rather than through
			// ExpectFixedSource, which refuses a result carrying no fixes rather than treating it
			// as an unchanged rewrite. Asserting the count is what separates a withheld repair from
			// one that lands and happens to write the same bytes.
			for _, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("expected no repair to be offered over a comment, got %d",
						len(diagnostic.Fixes))
				}
			}
		})
	}
}

// The finding points at the label identifier, not at the jump statement.
//
// Upstream reports at node.label, and its corpus states the column of every case: the first is
// column 20 of `A: while (a) break A;`, which is the `A` after the keyword rather than the one at
// the start of the line. Every ExpectFindings assertion above is satisfied by a rule reporting the
// whole `break A;` statement, so the span is the only thing separating the two readings, and a
// reader shown the statement would not be shown the token the repair deletes.
func TestNoExtraLabelPointsAtTheLabel(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoExtraLabel, extraLabelFile, "A: while (a) break A;")
	rule_testing.ExpectFindings(t, result, "unexpected")

	source := result.SourceFile.Text()
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "A" {
		t.Fatalf("expected the finding to point at the label, pointed at %q", reported)
	}
	// Upstream's column 20 is one-based, so the label sits at offset 19.
	if result.Diagnostics[0].Range.Pos() != 19 {
		t.Errorf("expected the label after the keyword at offset 19, reported at offset %d",
			result.Diagnostics[0].Range.Pos())
	}
	// The message is asserted against a literal rather than against the rule's own constant, since
	// comparing to the constant moves both sides together under mutation.
	if result.Diagnostics[0].Message.Id != "unexpected" {
		t.Errorf("expected the message id \"unexpected\", got %q", result.Diagnostics[0].Message.Id)
	}
}

// A labeled block pushes a frame, and only a function boundary makes that observable.
//
// Written for two surviving mutants: making a labeled statement push unconditionally, and making it
// never push, both survived all 38 imported cases. The arm pushes a frame that is labeled and not
// breakable, and such a frame can never itself produce a report, since the report condition wants a
// frame that is both. It can only stop the outward search early, and stopping early changes the
// answer only when an outer frame of the SAME name would otherwise have reported.
//
// Two labels of one name in one nest are a syntax error, so that shape is reachable only across a
// function boundary, where the inner label starts a fresh scope. Upstream's corpus writes no
// function at all, which is why nothing in it could see either mutant.
//
// Every expectation here was measured against ESLint's own Linter rather than derived, and the two
// reporting cases were checked by column as well as by count: upstream reports at column 52 and 55
// respectively, which are the labels inside and after the function rather than the outer one.
func TestNoExtraLabelStopsAtALabeledBlockOfTheSameName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"a labeled block inside a function shadows the outer loop label", "A: while (a) { function f() { A: { break A; } } }", 0},
		{"the same through an arrow", "A: while (a) { (() => { A: { break A; } })(); }", 0},
		{"a labeled loop inside a function reports on its own", "A: while (a) { function f() { A: while (b) { break A; } } }", 1},
		{"the outer loop still reports for a jump outside the function", "A: while (a) { function f() { A: { break A; } } break A; }", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "unexpected"
			}
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoExtraLabel, extraLabelFile, testCase.sourceText), expected...)
		})
	}
}
