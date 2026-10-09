package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// unusedLabelsFile is where the fixtures pretend to live.
const unusedLabelsFile = "/repository/source/UnusedLabels.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_unused_labels.rs`:
// 10 pass, 21 fail. The extractor flagged a discrepancy worth stating, because a fixture asserting
// one finding per input would have been wrong here: the snapshot records 32 diagnostics against
// those 21 inputs. Five inputs report more than once, and the counts were recovered by aligning each
// diagnostic on the source line it prints rather than by walking the snapshot in order, because two
// of the nested-label inputs are prefixes of each other and an in-order walk misattributes them.
//
//	A: B: 'foo'              2
//	A: B: C: 'foo'           3
//	A: B: C: D: 'foo'        4
//	A: B: C: D: E: 'foo'     5
//	label: while (true) { (() => { label: while (false) {} })(); }   2
//
// 16 remaining inputs report once. 16 + 2 + 3 + 4 + 5 + 2 = 32, which is the snapshot's total.
func TestNoUnusedLabelsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"a label on a variable statement", "A: var foo = 0;", 1},
		{"a label on a block nothing jumps to", "A: { foo(); bar(); }", 1},
		{"a label on an if nothing jumps to", "A: if (a) { foo(); bar(); }", 1},
		{"a loop whose break is unlabeled", "A: for (var i = 0; i < 10; ++i) { foo(); if (a) break; bar(); }", 1},
		{"a loop whose continue is unlabeled", "A: for (var i = 0; i < 10; ++i) { foo(); if (a) continue; bar(); }", 1},
		{"an inner label while the outer one is used", "A: for (var i = 0; i < 10; ++i) { B: break A; }", 1},
		{"a label shadowed by a variable of the same name", "A: { var A = 0; console.log(A); }", 1},
		{"a comment between the label and the body", "A: /* comment */ foo", 1},
		{"a comment between the label and the colon", "A /* comment */: foo", 1},
		{"a label on a string that would become a directive", `A: "use strict"`, 1},
		{"a label after a directive has already run", `"use strict"; foo: "bar"`, 1},
		{"a label on a parenthesized string", `A: ("use strict")`, 1},
		{"a label on a template literal", "A: `use strict`", 1},
		{"a label nested inside a block", "if (foo) { bar: 'baz' }", 1},
		{"two nested labels, both unused", "A: B: 'foo'", 2},
		{"three nested labels, all unused", "A: B: C: 'foo'", 3},
		{"four nested labels, all unused", "A: B: C: D: 'foo'", 4},
		{"five nested labels, all unused", "A: B: C: D: E: 'foo'", 5},
		{"a label on a bare number", "A: 42", 1},
		{"an inner label inside a function while the outer is used", "A: { function f() { B: { } } break A; }", 1},
		{"a label reused inside an arrow, neither jumped to", "label: while (true) { (() => { label: while (false) {} })(); }", 2},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "unusedLabel"
			}
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile, testCase.sourceText), expected...)
		})
	}
}

// The clean cases are the whole discrimination, and each one fails a different way.
//
// The two that matter most are the break/continue pair. A port that hooks only `break` reports the
// `continue A` case and silently halves the rule, and no other fixture here would catch it: every
// other pass case is reachable by a break-only implementation.
//
// The last four are about scope, and they are the cases a naive implementation gets wrong in the
// opposite direction. A label is used by a jump lexically inside its own body, so a rule that
// matched labels to jumps by name across the whole file would call the outer label used in the
// function and arrow cases. Upstream's `label: while (true) { f = function() { ... } }` case pins
// that a nested function's own `break label` does not reach the outer label of the same name, while
// the outer one is separately used and so stays clean.
func TestNoUnusedLabelsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a label broken to immediately", "A: break A;"},
		{"a label broken to from inside a block", "A: { foo(); break A; bar(); }"},
		{"a label broken to from inside an if", "A: if (a) { foo(); if (b) break A; bar(); }"},
		{"a loop label broken to by name", "A: for (var i = 0; i < 10; ++i) { foo(); if (a) break A; bar(); }"},
		{"a loop label continued to by name", "A: for (var i = 0; i < 10; ++i) { foo(); if (a) continue A; bar(); }"},
		{"three labels each used by a differently named jump", "A: { B: break B; C: for (var i = 0; i < 10; ++i) { foo(); if (a) break A; if (c) continue C; bar(); } }"},
		{"a label used despite a variable shadowing its name", "A: { var A = 0; console.log(A); break A; console.log(A); }"},
		{"a label reused inside a function, both used", "label: while (true) { f = function() { label: while (true) { break label; } }; break label; }"},
		{"a label inside a nested function, both used", "outer: { function f() { inner: { break inner; } } break outer; }"},
		{"a label inside an arrow, both used", "A: { const f = () => { B: { break B; } }; break A; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile, testCase.sourceText))
		})
	}
}

// The span, which the message-id fixtures above cannot see.
//
// Every finding must point at the label identifier alone, not at the labeled statement. Upstream's
// snapshot is explicit about this: the caret under `A: var foo = 0;` covers one column, and under
// `label: while (true) ...` it covers five. A rule reporting the whole statement passes every
// assertion in TestNoUnusedLabelsFires while pointing at the wrong thing, and a finding spanning a
// multi-line labeled loop is one an author cannot suppress with a -next-line directive.
//
// The nested case is the sharper one. The two findings for `A: B: 'foo'` must land on `A` and `B`
// separately, so a rule reporting the outer LabeledStatement's own range twice is caught here.
func TestNoUnusedLabelsPointsAtTheLabelOnly(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		reported   []string
	}{
		{"a single-character label", "A: var foo = 0;", []string{"A"}},
		{"a multi-character label", "label: while (false) {}", []string{"label"}},
		{"a label indented inside a block", "if (foo) { bar: 'baz' }", []string{"bar"}},
		{"a label carrying a comment before its colon", "A /* comment */: foo", []string{"A"}},
		{"each of two nested labels", "A: B: 'foo'", []string{"A", "B"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.reported) {
				t.Fatalf("expected %d findings, got %d", len(testCase.reported), len(result.Diagnostics))
			}
			source := result.SourceFile.Text()
			seen := make([]string, 0, len(result.Diagnostics))
			for _, diagnostic := range result.Diagnostics {
				seen = append(seen, source[diagnostic.Range.Pos():diagnostic.Range.End()])
			}
			for _, want := range testCase.reported {
				found := false
				for _, got := range seen {
					if got == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("no finding pointed at %q; findings pointed at %q", want, seen)
				}
			}
		})
	}
}

// The fix, asserted by applying it rather than by reading its text.
//
// A fix writing the right string over the wrong span passes a text comparison, so these compare the
// rewritten source. Every pair is verbatim from upstream's `expect_fix` table.
func TestNoUnusedLabelsFixRemovesTheLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"a variable statement", "A: var foo = 0;", "var foo = 0;"},
		{"a block", "A: { foo(); bar(); }", "{ foo(); bar(); }"},
		{"an if statement", "A: if (a) { foo(); bar(); }", "if (a) { foo(); bar(); }"},
		{
			"a loop with an unlabeled break",
			"A: for (var i = 0; i < 10; ++i) { foo(); if (a) break; bar(); }",
			"for (var i = 0; i < 10; ++i) { foo(); if (a) break; bar(); }",
		},
		{
			"a loop with an unlabeled continue",
			"A: for (var i = 0; i < 10; ++i) { foo(); if (a) continue; bar(); }",
			"for (var i = 0; i < 10; ++i) { foo(); if (a) continue; bar(); }",
		},
		{
			"only the inner label when the outer is used",
			"A: for (var i = 0; i < 10; ++i) { B: break A; }",
			"A: for (var i = 0; i < 10; ++i) { break A; }",
		},
		{"a block shadowing the label name", "A: { var A = 0; console.log(A); }", "{ var A = 0; console.log(A); }"},
		{"a label nested in a block", "if (foo) { bar: 'baz' }", "if (foo) { 'baz' }"},
		{"a bare number", "A: 42", "42"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile, testCase.sourceText), testCase.wantSource)
		})
	}
}

// Where this rule deliberately declines to repair, and why each refusal is load-bearing.
//
// This is the one place the port diverges from oxc, and it is a divergence toward ESLint rather than
// an invention. oxc declares `fix` unconditionally and rewrites the labeled statement to its body in
// every case, including these. ESLint gates the same repair behind `isFixable`, and each of its
// three refusals corresponds to a rewrite that parses and means something different. Since the
// engine's guard checks only that the rewritten file parses, none of these would be refused
// downstream; a wrong fix here is applied unattended.
//
// Each refusal was checked against a real parser rather than reasoned about:
//
//	comment    `A: /* comment */ foo` -> `foo` silently discards the comment, and where the comment
//	           sits before the colon there is no span that removes the label and keeps it.
//	directive  `A: "use strict"` -> `"use strict"` promotes the string to a directive. Confirmed
//	           with node --check: the unlabeled form rejects a duplicate parameter name that the
//	           labeled form accepts, so the rewrite changes which programs are legal.
//	ASI        `foo()\nLABEL: [1].forEach(x => x)` -> `foo()\n[1].forEach(x => x)` parses as
//	           `foo()[1]`. Confirmed with node --check: both forms parse, and they are different
//	           programs. This is exactly the failure the parse guard structurally cannot catch.
//
// The finding still fires in all of these. Only the repair is withheld.
func TestNoUnusedLabelsDeclinesUnsafeFixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a comment between the label and the body", "A: /* comment */ foo"},
		{"a comment between the label and the colon", "A /* comment */: foo"},
		{"a string that would become a directive", `A: "use strict"`},
		{"a template that would become a directive", "A: `use strict`"},
		{"an ASI hazard opening with a bracket", "foo()\nLABEL: [1].forEach(x => x)"},
		{"an ASI hazard opening with a parenthesis", "foo()\nLABEL: (function () {})()"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile, testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("expected the label to still be reported, got no findings")
			}
			for _, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("expected no fix, got %d proposing %q",
						len(diagnostic.Fixes), diagnostic.Fixes[0].Text)
				}
			}
		})
	}
}

// Directive position is not only the top of a file, and it is not only the innermost label.
//
// Both cases here were written for surviving mutants, and both guard a rewrite that parses and means
// something different. The imported corpus reaches neither: every directive case upstream carries a
// single label at the top level, so a rule checking only the innermost label's immediate parent
// against the source file passes all of them while withholding nothing it should.
//
//	function body    A directive prologue exists at the top of a function body as well as at the top
//	                 of a file. Confirmed with node --check: `function g() { "use strict"; return
//	                 function f(a, a) { return a }; }` is rejected for the duplicate parameter, and
//	                 the same source with the string labeled is accepted. So removing the label
//	                 changes which programs are legal, exactly as it does at the top level.
//	nested labels    Only the innermost label is refused. `A: B: "use strict"` may lose `A`, because
//	                 what remains is `B: "use strict"`, still labeled and so still not a directive.
//	                 The chain never collapses into one, because the last label to go is refused on
//	                 its own pass.
//
// That second line is a correction rather than a design note. This port briefly walked through the
// label chain to the string underneath and refused the outer labels too, which withheld repairs
// upstream offers. ESLint's Linter was asked directly and fixes `A: B: 'foo'` to `B: 'foo'` and
// `A: B: C: "use strict"` to `C: "use strict"`, so the immediate body is the right question.
//
// The finding fires in every case below. Only the repair is withheld.
func TestNoUnusedLabelsDeclinesFixesInEveryDirectivePosition(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a string at the top of a function body", `function g() { A: "use strict"; }`},
		{"a string at the top of a function expression body", `const g = function () { A: "use strict"; };`},
		{"a string at the top of an arrow body", `const g = () => { A: "use strict"; };`},
		{"a string at the top of a method body", `class C { m() { A: "use strict"; } }`},
		{"a template at the top of a function body", "function g() { A: `use strict`; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile, testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("expected the label to still be reported, got no findings")
			}
			for _, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("expected no fix, got %d proposing %q",
						len(diagnostic.Fixes), diagnostic.Fixes[0].Text)
				}
			}
		})
	}
}

// A label chain over a string: the outer labels are repaired and only the innermost is held back.
//
// This is the pair that pins the correction above, and it is asserted by applying the repairs and
// comparing the source rather than by counting them. Both expectations were taken from ESLint's own
// Linter rather than derived: it rewrites `A: B: C: "use strict"` to `C: "use strict"`, removing two
// labels and refusing the third.
func TestNoUnusedLabelsFixesOuterLabelsOverADirective(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"one outer label over a top-level directive", `A: B: "use strict"`, `B: "use strict"`},
		{"two outer labels over a top-level directive", `A: B: C: "use strict"`, `C: "use strict"`},
		{
			"an outer label over a directive in a function body",
			`function g() { A: B: "use strict"; }`,
			`function g() { B: "use strict"; }`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile, testCase.sourceText), testCase.wantSource)
		})
	}
}

// A directive can only be the thing a directive prologue holds, and that is a position question.
//
// ESLint refuses the fix only when the labeled statement's outermost non-label ancestor is a Program
// or a function body, because those are the only two places a string expression statement is read as
// a directive. Anywhere else the same string is an ordinary expression and the fix is safe. Upstream
// covers the refusing side with `A: "use strict"` and never the permitting side, so these exist
// because a rule refusing every string literal passes the whole imported corpus while withholding
// repairs it should offer.
func TestNoUnusedLabelsFixesAStringOutsideDirectivePosition(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"a string inside a block", "{ A: 'use strict'; }", "{ 'use strict'; }"},
		{"a string inside an if", "if (a) { A: 'use strict'; }", "if (a) { 'use strict'; }"},
		{"a string inside a loop body", "while (a) { A: 'use strict'; }", "while (a) { 'use strict'; }"},
		// A parenthesized string is never a directive, so the fix applies even at the top level
		// where a bare string would be refused. Verified against a real parser rather than reasoned
		// about: `("use strict")` followed by a function with duplicate parameter names is accepted
		// by node --check, while the unparenthesized form rejects it. ESLint reaches the same answer
		// by a different route, requiring the expression to be a Literal, which parentheses fail.
		//
		// This case is upstream's `A: ("use strict")`, which oxc lists as a fail input and does fix.
		{"a parenthesized string at the top level", `A: ("use strict")`, `("use strict")`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile, testCase.sourceText), testCase.wantSource)
		})
	}
}

// A comment outside the removed span does not block the repair.
//
// Written for a surviving mutant: the comment test asks whether a comment overlaps the span the fix
// would delete, and dropping the half that checks the comment ends after the span begins changed no
// fixture. Upstream's two comment cases both sit inside the span, so nothing covered a comment
// merely earlier in the file.
//
// The mutant was run against these inputs before this was written rather than after: it withheld the
// repair on all four, where the correct rule offers it. A comment anywhere above a label would have
// suppressed every fix in the file, and the finding would still have fired, so the symptom is a rule
// that quietly stops repairing rather than one that reports wrongly.
func TestNoUnusedLabelsFixesDespiteACommentOutsideTheSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"a block comment before the label", "/* leading */ A: foo", "/* leading */ foo"},
		{"a line comment before the label", "// leading\nA: foo", "// leading\nfoo"},
		{"two block comments before the label", "/* one */ /* two */ A: foo;", "/* one */ /* two */ foo;"},
		{"a comment after an earlier statement", "foo(); /* trailing */ A: bar", "foo(); /* trailing */ bar"},
		// The other side of the same overlap test, and its own surviving mutant. Dropping the half
		// that checks the comment starts before the span ends withheld the repair on these two while
		// leaving the four above alone, so each half of the comparison needs a case on its own side.
		{"a block comment after the body", "A: foo /* after */", "foo /* after */"},
		{"a block comment after a terminated body", "A: foo; /* after */", "foo; /* after */"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile, testCase.sourceText), testCase.wantSource)
		})
	}
}

// A case written because somebody read our code rather than upstream's.
//
// Upstream's corpus is JavaScript throughout and never labels a statement in a TypeScript-only
// construct. A label inside a namespace body is the shape most likely to break a port that decides
// directive position by checking for a Program or function ancestor: a namespace body is neither,
// so the fix applies, and the walk has to survive the extra node kind rather than assume it.
func TestNoUnusedLabelsHandlesTypeScriptContainers(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile,
		"namespace Outer { A: { foo(); } }")
	rule_testing.ExpectFindings(t, result, "unusedLabel")
	rule_testing.ExpectFixedSource(t, result, "namespace Outer { { foo(); } }")
}

// A jump may not name a label outside the function holding it, and the rule has to agree.
//
// Written for a surviving mutant: deleting the function and class boundary from the upward walk
// changed no fixture, because every case that had one also had a nearer label of the same name, and
// the name match stopped the walk before the boundary could matter. The boundary only decides
// anything when a jump names a label that exists outside the function and not inside it.
//
// Such a program is not legal JavaScript. Confirmed with node --check, which rejects
// `outer: while (true) { (function () { break outer; })(); }` with "Undefined label 'outer'", so on
// valid input this branch can never change the verdict. It is still not dead: the TypeScript parser
// is error-tolerant and hands a rule the tree for source a runtime would refuse, and a linter that
// misreads such a file reports the wrong thing rather than nothing. Here the outer label is genuinely
// unused, because the jump naming it cannot reach it, and a rule without the boundary calls it used
// and stays silent.
func TestNoUnusedLabelsDoesNotLetAJumpEscapeItsFunction(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a function expression", "outer: while (true) { (function () { break outer; })(); }"},
		{"an arrow function", "outer: while (true) { (() => { break outer; })(); }"},
		{"a continue rather than a break", "outer: while (true) { (function () { continue outer; })(); }"},
		{"a method in a class body", "outer: while (true) { class C { m() { break outer; } } }"},
		// The class arm of the boundary, which the four above do not reach. Written for a
		// surviving mutant: dropping `ast.IsClassLike` from the walk while keeping
		// `ast.IsFunctionLikeDeclaration` changed no fixture, because in every case above the
		// jump sits inside a method and the function arm stops the walk before the class arm
		// can matter. A class static initialization block is the one place a jump's ancestor
		// chain reaches a class before any function, confirmed by walking the parsed tree.
		//
		// Such a program is not legal JavaScript. ESLint's own parser rejects it outright with
		// "Unsyntactic break", so on valid input this branch cannot change the verdict. It is
		// still not dead, for the same reason the function boundary is not: the TypeScript
		// parser is error-tolerant and hands a rule the tree for source a runtime would refuse.
		// Here the outer label is genuinely unused, because the jump naming it cannot reach it,
		// and a rule without the class arm calls it used and stays silent.
		{"a class static initialization block", "outer: while (true) { class C { static { break outer; } } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unusedLabel")

			source := result.SourceFile.Text()
			reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != "outer" {
				t.Errorf("expected the finding to point at the unreachable label, pointed at %q", reported)
			}
		})
	}
}

// Two labels of the same name where the inner one is used and the outer is not.
//
// Upstream has the reverse (`A: { function f() { B: { } } break A; }`, outer used and inner not) but
// never this ordering. It pins that `markAsUsed` stops at the nearest enclosing label with a
// matching name rather than marking every label of that name: a rule walking the whole stack and
// setting each match would call the outer label used and report nothing.
func TestNoUnusedLabelsMarksOnlyTheNearestMatchingLabel(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile,
		"A: { A: { break A; } }")
	rule_testing.ExpectFindings(t, result, "unusedLabel")

	source := result.SourceFile.Text()
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "A" {
		t.Fatalf("expected the finding to point at a label, pointed at %q", reported)
	}
	// The outer label is the one at offset zero, and it is the one that must report.
	if result.Diagnostics[0].Range.Pos() != 0 {
		t.Errorf("expected the outer label to be the unused one, reported at offset %d",
			result.Diagnostics[0].Range.Pos())
	}
}

// A safe token before the label defeats the ASI hazard, even when the body opens unsafely.
//
// Written for a surviving mutant: dropping the half of the ASI test that asks whether the preceding
// token already terminates a statement changed no fixture. The imported corpus never pairs a safe
// preceding token with an unsafe body opener, so every fix case it carries has a body starting with
// an ordinary token and the preceding-token half is never consulted.
//
// The mutant withholds the repair on all four of these, where the correct rule offers it. The
// symptom is a rule that quietly stops repairing rather than one that reports wrongly, which is why
// no message-id fixture could have seen it.
//
// Both expectations were taken from ESLint's own Linter rather than derived: cohereAndFix rewrites
// `{ A: (foo) }` to `{ (foo) }` and `foo; A: (bar)` to `foo; (bar)`. Nothing can rejoin across a
// `{`, a `;`, or a `:`, and at the start of the file there is no previous token to rejoin with.
func TestNoUnusedLabelsFixesWhenThePrecedingTokenTerminates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"an open brace before the label", "{ A: (foo) }", "{ (foo) }"},
		{"a semicolon before the label", "foo; A: (bar)", "foo; (bar)"},
		{"nothing at all before the label", "A: (foo)", "(foo)"},
		{"an open brace before a bracket body", "function f() { A: [1]; }", "function f() { [1]; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoUnusedLabels, unusedLabelsFile, testCase.sourceText), testCase.wantSource)
		})
	}
}
