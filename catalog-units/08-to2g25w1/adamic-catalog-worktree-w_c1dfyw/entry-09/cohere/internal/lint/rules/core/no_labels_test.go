package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// noLabelsFile is where the fixtures pretend to live.
const noLabelsFile = "/repository/source/NoLabels.ts"

// The corpus is ESLint's own, copied rather than rewritten.
//
// Verbatim from `eslint/tests/lib/rules/no-labels.js`, extracted by loading that file with the rule
// tester stubbed out. 8 valid and 21 invalid. Several invalid cases name TWO
// message ids and one names three, so the per-input finding count is stated by the corpus rather
// than assumed, and a fixture asserting one finding per input would be wrong for most of them.

// The clean cases, and note that five of the eight are clean only under an option.
//
// The first is the trap: `{ label: foo () }` is an object property named `label`, not a labeled
// statement, and a rule matching on the identifier reports it. The next three are unlabeled loop
// jumps, which this rule never touches. The last four are labels the options permit.
func TestNoLabelsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
	}{
		{"var f = { label: foo ()}", nil},
		{"while (true) {}", nil},
		{"while (true) { break; }", nil},
		{"while (true) { continue; }", nil},
		{"A: while (a) { break A; }", NoLabelsOptions{AllowLoop: true, AllowSwitch: false}},
		{"A: do { if (b) { break A; } } while (a);", NoLabelsOptions{AllowLoop: true, AllowSwitch: false}},
		{"A: for (var a in obj) { for (;;) { switch (a) { case 0: continue A; } } }", NoLabelsOptions{AllowLoop: true, AllowSwitch: false}},
		{"A: switch (a) { case 0: break A; }", NoLabelsOptions{AllowLoop: false, AllowSwitch: true}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(
				t, NoLabels, noLabelsFile, testCase.sourceText, testCase.options))
		})
	}
}

// The failing cases, with the exact ids the corpus names, in the corpus's own order.
//
// The ordering is load-bearing rather than incidental. Upstream reports the label itself on
// `LabeledStatement:exit` and reports a labeled `break` or `continue` on entry, so a nested label
// emits its findings inside-out while the jumps interleave. `A: switch (a) { case 0: B: { break A; }
// default: break; };` is the case that pins it: two `unexpectedLabel` then one
// `unexpectedLabelInBreak`, which only comes out in that order if the label reports on exit.
func TestNoLabelsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
		wantIds    []string
	}{
		{"label: while(true) {}", nil, []string{"unexpectedLabel"}},
		{"label: while (true) { break label; }", nil, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"label: while (true) { continue label; }", nil, []string{"unexpectedLabel", "unexpectedLabelInContinue"}},
		{"A: var foo = 0;", nil, []string{"unexpectedLabel"}},
		{"A: break A;", nil, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: { if (foo()) { break A; } bar(); };", nil, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: if (a) { if (foo()) { break A; } bar(); };", nil, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: switch (a) { case 0: break A; default: break; };", nil, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: switch (a) { case 0: B: { break A; } default: break; };", nil, []string{"unexpectedLabel", "unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: var foo = 0;", NoLabelsOptions{AllowLoop: true, AllowSwitch: false}, []string{"unexpectedLabel"}},
		{"A: break A;", NoLabelsOptions{AllowLoop: true, AllowSwitch: false}, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: { if (foo()) { break A; } bar(); };", NoLabelsOptions{AllowLoop: true, AllowSwitch: false}, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: if (a) { if (foo()) { break A; } bar(); };", NoLabelsOptions{AllowLoop: true, AllowSwitch: false}, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: switch (a) { case 0: break A; default: break; };", NoLabelsOptions{AllowLoop: true, AllowSwitch: false}, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: var foo = 0;", NoLabelsOptions{AllowLoop: false, AllowSwitch: true}, []string{"unexpectedLabel"}},
		{"A: break A;", NoLabelsOptions{AllowLoop: false, AllowSwitch: true}, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: { if (foo()) { break A; } bar(); };", NoLabelsOptions{AllowLoop: false, AllowSwitch: true}, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: if (a) { if (foo()) { break A; } bar(); };", NoLabelsOptions{AllowLoop: false, AllowSwitch: true}, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: while (a) { break A; }", NoLabelsOptions{AllowLoop: false, AllowSwitch: true}, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: do { if (b) { break A; } } while (a);", NoLabelsOptions{AllowLoop: false, AllowSwitch: true}, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
		{"A: for (var a in obj) { for (;;) { switch (a) { case 0: break A; } } }", NoLabelsOptions{AllowLoop: false, AllowSwitch: true}, []string{"unexpectedLabel", "unexpectedLabelInBreak"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(
				t, NoLabels, noLabelsFile, testCase.sourceText, testCase.options),
				testCase.wantIds...)
		})
	}
}

// The spans, which no message-id fixture above can see.
//
// Three different anchors, and each is a place the port could point wrong while every id assertion
// stayed green. Measured against the installed eslint 10.8.1 build, whose columns are 1-based and
// whose end column is exclusive, so `[1,1,1,23]` on a 22-character input is the whole statement.
//
// A label reports on the WHOLE labeled statement, not on the label identifier: `A: if (a) { break
// A; }` spans all 22 characters. `no-unused-labels` in this same package reports on the identifier
// instead, so a port taking its shape from that neighbour would anchor wrong and pass every id
// fixture. That is exactly the "read siblings for shape, never for semantics" trap.
func TestNoLabelsSpansTheWholeStatement(t *testing.T) {
	t.Parallel()

	const sourceText = "A: if (a) { break A; }"

	result := rule_testing.Run(t, NoLabels, noLabelsFile, sourceText)
	if len(result.Diagnostics) != 2 {
		t.Fatalf("wanted two diagnostics, got %d", len(result.Diagnostics))
	}

	// The label finding covers the entire labeled statement.
	label := result.Diagnostics[0]
	if got := sourceText[label.Range.Pos():label.Range.End()]; got != sourceText {
		t.Fatalf("label finding spans %q, wanted the whole statement %q", got, sourceText)
	}
	if label.Message.Id != "unexpectedLabel" {
		t.Fatalf("first finding is %q, wanted unexpectedLabel", label.Message.Id)
	}

	// The break finding covers the break statement including its semicolon, matching eslint's
	// [1,13,1,21] on this input.
	jump := result.Diagnostics[1]
	if got := sourceText[jump.Range.Pos():jump.Range.End()]; got != "break A;" {
		t.Fatalf("break finding spans %q, wanted %q", got, "break A;")
	}
	if jump.Message.Id != "unexpectedLabelInBreak" {
		t.Fatalf("second finding is %q, wanted unexpectedLabelInBreak", jump.Message.Id)
	}
}

// A continue's span, asserted separately because it reports through a different arm and a different
// message than break.
func TestNoLabelsSpansAContinue(t *testing.T) {
	t.Parallel()

	const sourceText = "A: while (a) { B: while (b) { continue A; } }"

	result := rule_testing.Run(t, NoLabels, noLabelsFile, sourceText)
	if len(result.Diagnostics) != 3 {
		t.Fatalf("wanted three diagnostics, got %d", len(result.Diagnostics))
	}

	jump := result.Diagnostics[2]
	if got := sourceText[jump.Range.Pos():jump.Range.End()]; got != "continue A;" {
		t.Fatalf("continue finding spans %q, wanted %q", got, "continue A;")
	}
	if jump.Message.Id != "unexpectedLabelInContinue" {
		t.Fatalf("third finding is %q, wanted unexpectedLabelInContinue", jump.Message.Id)
	}
}

// The nil-options path, which every fixture above reaches THROUGH the decoder and therefore cannot
// see.
//
// A rule configured as a bare `"error"` is handed nil rather than a decoded struct: the decoder
// errors on empty input and the config turns that into nil. The rule's fallback has to produce the
// strict setting, because that is upstream's default. Passing nil directly bypasses the decoder the
// way the live config does.
func TestNoLabelsDefaultsWithoutTheDecoder(t *testing.T) {
	t.Parallel()

	// Under the default, a label wrapping a loop still reports, and so does the break naming it.
	rule_testing.ExpectFindings(t,
		rule_testing.RunWithOptions(t, NoLabels, noLabelsFile, "A: while (a) { break A; }", nil),
		"unexpectedLabel", "unexpectedLabelInBreak")

	// The control: the same input under allowLoop is clean, so the finding above is about the
	// default rather than about the rule reporting unconditionally.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile,
		"A: while (a) { break A; }", NoLabelsOptions{AllowLoop: true}))
}

// The decoder itself, routed through the rule's own registration rather than by building the struct.
//
// A serde alias or an inverted default has no upstream counterpart and is exactly what a
// hand-built options struct leaves untested. Both field names are checked, because a typo in either
// tag reads as "the option does nothing" and every fixture above would still pass by constructing
// the struct directly.
func TestNoLabelsDecodesItsOptionNames(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[NoLabelsOptions]()

	cases := []struct {
		raw  string
		want NoLabelsOptions
	}{
		{`{"allowLoop":true}`, NoLabelsOptions{AllowLoop: true}},
		{`{"allowSwitch":true}`, NoLabelsOptions{AllowSwitch: true}},
		{`{"allowLoop":true,"allowSwitch":true}`, NoLabelsOptions{AllowLoop: true, AllowSwitch: true}},
		{`{}`, NoLabelsOptions{}},
	}

	for _, testCase := range cases {
		decoded, err := decode([]byte(testCase.raw))
		if err != nil {
			t.Fatalf("decoding %s: %v", testCase.raw, err)
		}
		if decoded != any(testCase.want) {
			t.Fatalf("decoding %s gave %+v, wanted %+v", testCase.raw, decoded, testCase.want)
		}
	}
}

// The permission is decided by what the label WRAPS, not by what the jump sits inside.
//
// Both rows measured against the installed build. They are the same shape with the same nesting and
// opposite verdicts, and the only difference is which option is set, which is what makes them the
// pair that separates a correct port from one asking "is this jump inside a loop".
func TestNoLabelsPermissionFollowsTheLabelNotTheJump(t *testing.T) {
	t.Parallel()

	const nested = "A: for (var a in obj) { for (;;) { switch (a) { case 0: continue A; } } }"

	// `A` labels a `for`, so allowLoop exempts it even though the jump sits inside a switch.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile, nested,
		NoLabelsOptions{AllowLoop: true}))

	// allowSwitch does not, for the same input, because `A` does not label a switch.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile, nested,
		NoLabelsOptions{AllowSwitch: true}), "unexpectedLabel", "unexpectedLabelInContinue")
}

// A label wrapping another label is not a loop, even when the inner one is.
//
// `ast.IsIterationStatement` takes a `lookInLabeledStatements` flag and passing `true` would see
// through the inner label and exempt `A`. Upstream tests the body node's own type and does not.
// Measured: `A: B: while (a) { break A; }` under allowLoop reports `A` and the break, and leaves
// `B` alone. A port passing `true` goes silent on both findings and no imported fixture sees it,
// because the corpus writes no doubled label under an option.
func TestNoLabelsDoesNotSeeThroughANestedLabel(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile,
		"A: B: while (a) { break A; }", NoLabelsOptions{AllowLoop: true}),
		"unexpectedLabel", "unexpectedLabelInBreak")

	// Naming the INNER label instead leaves only the outer label's own finding, which shows the
	// jump resolved through the stack rather than through its ancestors.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile,
		"A: B: while (a) { break B; }", NoLabelsOptions{AllowLoop: true}), "unexpectedLabel")
}

// Two shapes ESLint refuses at PARSE time and our parser recovers from, so the rule has to answer
// them and no imported fixture can say what the answer should be.
//
// Both were found as surviving mutants, and reading upstream did not settle either one: driving the
// installed eslint 10.8.1 build returned zero findings for both with a `fatal` parse error, so
// upstream has no behaviour here to port.
//
//	A: while (a) { A: switch (b) { ... } }   Parsing error: Label 'A' is already declared
//	break A;                                 Parsing error: Unsyntactic break
//
// typescript-go recovers from illegal source and hands the rule a real tree for both, which is a
// parser difference rather than a rule decision. The choices pinned here follow the language's own
// resolution rules, which is the only authority left once upstream declines to have an opinion:
//
//	a duplicate label resolves INNERMOST-first, matching how every scoped name in the language
//	  resolves, so `break A` inside the inner label names the inner one
//	a jump naming no live label is NOT permitted, because no option can have exempted a label that
//	  does not exist, and defaulting to permitted would silently swallow the finding
//
// A mutant reversing the stack search direction and a mutant making an unresolved name permitted
// both survived the whole imported corpus, which is how these were found.
func TestNoLabelsAnswersShapesEslintRefusesToParse(t *testing.T) {
	t.Parallel()

	// Duplicate label: the inner `A` names the switch, so `allowSwitch` exempts the break while
	// `allowLoop` does not. Reversing the stack search flips both rows.
	const duplicate = "A: while (a) { A: switch (b) { case 0: break A; } }"

	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile, duplicate,
		NoLabelsOptions{AllowSwitch: true}), "unexpectedLabel")

	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile, duplicate,
		NoLabelsOptions{AllowLoop: true}), "unexpectedLabel", "unexpectedLabelInBreak")

	// The mirror image, which is what makes the pair a measurement rather than one row: the same
	// two options give the opposite verdicts when the nesting is reversed.
	const reversed = "A: switch (b) { case 0: A: while(a) { break A; } }"

	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile, reversed,
		NoLabelsOptions{AllowLoop: true}), "unexpectedLabel")

	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile, reversed,
		NoLabelsOptions{AllowSwitch: true}), "unexpectedLabel", "unexpectedLabelInBreak")

	// A jump naming a label that does not exist reports, under every option setting, because no
	// option can exempt a label nothing declared.
	for _, settings := range []any{
		nil,
		NoLabelsOptions{AllowLoop: true},
		NoLabelsOptions{AllowSwitch: true},
		NoLabelsOptions{AllowLoop: true, AllowSwitch: true},
	} {
		rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile,
			"while(a) { break A; }", settings), "unexpectedLabelInBreak")
		rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile,
			"while(a) { continue A; }", settings), "unexpectedLabelInContinue")
	}
}

// The label stack has to POP when a label's scope ends, and this is the input that shows it.
//
// Found as a surviving mutant: deleting the pop entirely left all 29 corpus cases green, because
// upstream's corpus never writes two labels in sequence -- only nested ones, where a stale entry is
// indistinguishable from a live one.
//
//	A: while(a) {} A: while(b) { break A; }
//
// Two SEQUENTIAL labels sharing a name, which is legal (unlike the nested duplicate above, which is
// a parse error in both engines' grammars) and which eslint 10.8.1 accepts with no fatal and
// reports clean under allowLoop. Without the pop, the first `A` is still on the stack when the
// second is pushed, so a search finding the stale entry first would answer about the wrong label.
// Both labels here wrap loops so the verdict is the same either way; what the case actually pins is
// that the stack does not GROW without bound across siblings, which is what the second assertion
// measures.
func TestNoLabelsPopsTheStackWhenScopeEnds(t *testing.T) {
	t.Parallel()

	// Sequential same-name labels, both wrapping loops: clean under allowLoop. Measured on eslint
	// 10.8.1, which parses this without a fatal and reports nothing.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile,
		"A: while(a) {} A: while(b) { break A; }", NoLabelsOptions{AllowLoop: true}))

	// The measurement that needs the pop: a label wrapping a SWITCH, followed by a sibling label of
	// the same name wrapping a loop. If the first entry survives into the second label's subtree, a
	// stack search reaches a switch-bodied `A` and `allowLoop` stops exempting the break.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile,
		"A: switch(a) {} A: while(b) { break A; }", NoLabelsOptions{AllowLoop: true}),
		"unexpectedLabel")

	// The control for the row above, showing the finding is the SWITCH label rather than the break:
	// under allowSwitch the switch label is exempt and the loop label reports instead, and the
	// break stays clean under both because innermost-first resolution reaches the loop.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoLabels, noLabelsFile,
		"A: switch(a) {} A: while(b) { break A; }", NoLabelsOptions{AllowSwitch: true}),
		"unexpectedLabel", "unexpectedLabelInBreak")
}
