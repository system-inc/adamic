package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// constAssignFile is where the fixtures pretend to live.
//
// A real path matters more here than for a syntactic rule: this rule reads the checker, so the
// harness builds an actual program and the file has to sit somewhere a tsconfig can reach.
const constAssignFile = "/repository/source/ConstAssign.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_const_assign.rs`:
// one Tester block, 20 pass and 24 fail. The snapshot records 25 diagnostics from those 24 inputs,
// so one finding per input is wrong here, and the extractor said so before any code was written.
// The extra diagnostic belongs to `const x = 0; x = 1; x = 2;`, which writes twice and reports
// twice. Recovered from the snapshot by source line rather than by an in-order walk, since that
// case prints two entries with byte-identical source text.
func TestNoConstAssignFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a plain reassignment", "const x = 0; x = 1;", []string{"noConstAssign"}},
		{"a reassignment of a destructured binding", "const {a: x} = {a: 0}; x = 1;",
			[]string{"noConstAssign"}},
		{"a shorthand destructuring target", "const x = 0; ({x} = {x: 1});",
			[]string{"noConstAssign"}},
		{"a destructuring target with a default", "const x = 0; ({a: x = 1} = {});",
			[]string{"noConstAssign"}},
		{"a compound assignment", "const x = 0; x += 1;", []string{"noConstAssign"}},
		{"a prefix increment", "const x = 0; ++x;", []string{"noConstAssign"}},
		{"an increment in a for head", "for (const i = 0; i < 10; ++i) { foo(i); }",
			[]string{"noConstAssign"}},
		// The discrepancy case. Two writes, two findings.
		{"two reassignments of the same binding", "const x = 0; x = 1; x = 2;",
			[]string{"noConstAssign", "noConstAssign"}},
		{"a reassignment from inside a function", "const x = 0; function foo() { x = x + 1; }",
			[]string{"noConstAssign"}},
		{"a reassignment from a function with a parameter",
			"const x = 0; function foo(a) { x = a; }", []string{"noConstAssign"}},
		{"a reassignment inside a loop body", "const x = 0; while (true) { x = x + 1; }",
			[]string{"noConstAssign"}},
		{"a reassignment from a doubly nested function",
			"const x = 0; function foo(a) { function bar(b) { x = b; } bar(123); }",
			[]string{"noConstAssign"}},
		// Upstream's comment: error even if the declaration comes after the assignment, which
		// aligns with eslint.
		{"a reassignment hoisted above the declaration", "x = 123; const x = 1;",
			[]string{"noConstAssign"}},
		// Binding patterns
		{"a reassignment of a deeply nested rest binding",
			"const [a, b, ...[c, ...d]] = [1, 2, 3, 4, 5]; d = 123", []string{"noConstAssign"}},
		{"a nested rest element as a destructuring target",
			"const d = 123; [a, b, ...[c, ...d]] = [1, 2, 3, 4, 5]", []string{"noConstAssign"}},
		{"an object rest as a destructuring target",
			"const b = 0; ({a, ...b} = {a: 1, c: 2, d: 3})", []string{"noConstAssign"}},
		// using + await using
		{"a reassignment of a using binding", "using x = foo(); x = 1;",
			[]string{"noConstAssign"}},
		{"a reassignment of an await using binding", "await using x = foo(); x = 1;",
			[]string{"noConstAssign"}},
		{"a logical assignment to a using binding", "using x = foo(); x ??= bar();",
			[]string{"noConstAssign"}},
		{"a logical assignment to an await using binding", "await using x = foo(); x ||= bar();",
			[]string{"noConstAssign"}},
		{"an array destructuring target on a using binding", "using x = foo(); [x, y] = bar();",
			[]string{"noConstAssign"}},
		{"an array destructuring target with a default on an await using binding",
			"await using x = foo(); [x = baz, y] = bar();", []string{"noConstAssign"}},
		{"an object destructuring target on a using binding",
			"using x = foo(); ({a: x} = bar());", []string{"noConstAssign"}},
		{"an object destructuring target with a default on an await using binding",
			"await using x = foo(); ({a: x = baz} = bar());", []string{"noConstAssign"}},

		// Beyond upstream, from reading our code.

		// The kind-versus-identity discriminator, and the reason this rule compares declaration
		// nodes rather than declaration kinds. Both anchors are `KindVariableDeclaration`, so a
		// kind comparison calls the inner declaration a match for the outer one as well and reports
		// twice. Nothing in the imported corpus can see this: upstream's shadow cases resolve to a
		// declaration this rule never anchored on, so they exit through a different path and a kind
		// mutant survives all 25 of them.
		{"a nested redeclaration written in its own block", "const A = 1; { const A = 2; A = 3; }",
			[]string{"noConstAssign"}},
		// The same shape with the write in the outer scope, so the reported one is the outer
		// binding. Guards against a port that anchors on whichever declaration it saw last.
		{"a nested redeclaration with the outer binding written",
			"const A = 1; { const A = 2; foo(A); } A = 3;", []string{"noConstAssign"}},
		// A postfix update. Upstream's corpus has `++x` but never `x++`, and the sibling rule
		// `no-ex-assign` shipped missing postfix entirely because its corpus had the same hole.
		{"a postfix increment", "const x = 0; x++;", []string{"noConstAssign"}},
		{"a postfix decrement", "const x = 0; x--;", []string{"noConstAssign"}},
		{"a prefix decrement", "const x = 0; --x;", []string{"noConstAssign"}},
		// A `for (x of ...)` head assigning to an existing const on every iteration. Upstream's
		// corpus only has the declaring form, which is clean, so this direction is untested there.
		{"a for-of head assigning to a const", "const x = 0; for (x of [1,2]) { foo(x); }",
			[]string{"noConstAssign"}},
		{"a for-in head assigning to a const", "const x = 0; for (x in {a:1}) { foo(x); }",
			[]string{"noConstAssign"}},
		// Two separate declarators in one list. The rule anchors on the list, so a port looping
		// only over the first declarator passes every upstream case and misses this.
		{"a second declarator in the same list reassigned", "const a = 0, b = 1; b = 2;",
			[]string{"noConstAssign"}},
		{"both declarators in one list reassigned", "const a = 0, b = 1; a = 2; b = 3;",
			[]string{"noConstAssign", "noConstAssign"}},
		// A nested object binding pattern, which upstream exercises only for arrays.
		{"a reassignment of a nested object pattern binding",
			"const {a: {b: c}} = {a: {b: 0}}; c = 1;", []string{"noConstAssign"}},
		// A write through a parenthesised target, which the shelf accessor handles and a
		// hand-written climb would have to remember.
		{"a parenthesised assignment target", "const x = 0; (x) = 1;", []string{"noConstAssign"}},
		// A const redeclared inside a nested function and written there. Written first as a clean
		// case on the assumption it was a shadow, which was wrong and the rule caught it: the inner
		// binding is itself a const being reassigned, so this is one violation rather than none.
		// Kept because getting it wrong is the easy mistake, and a rule that anchored only on the
		// outermost declaration would report zero here.
		{"a const redeclared in a nested function and written there",
			"const A = 1; function f() { const A = 2; A = 3; }", []string{"noConstAssign"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoConstAssign, constAssignFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// The clean cases are the whole discrimination, and most of them are a shadow of some shape.
//
// `const x = 0; { let x; x = 1; }` writes to a block-scoped shadow, `function a(x) { x = 1; }` to a
// parameter, and `const a = 1; { let a = 2; { a += 1; } }` to a shadow from a nested block. Those
// are the reason this rule reads the checker at all: no bounded walk distinguishes them from the
// fail cases, because the text is identical.
//
// The rest are the structural half. `x.key = 1` and `foo(x)` both resolve to the const binding and
// neither reassigns it, so a rule built on symbol identity alone reports both.
func TestNoConstAssignStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a write to a block-scoped shadow", "const x = 0; { let x; x = 1; }"},
		{"a write to a parameter shadowing the const", "const x = 0; function a(x) { x = 1; }"},
		{"a read of the const", "const x = 0; foo(x);"},
		{"a for-in head declaring a fresh const", "for (const x in [1,2,3]) { foo(x); }"},
		{"a for-of head declaring a fresh const", "for (const x of [1,2,3]) { foo(x); }"},
		{"a property write on a const object", "const x = {key: 0}; x.key = 1;"},
		{"a write to a var", "var x = 0; x = 1;"},
		{"a write to a let", "let x = 0; x = 1;"},
		{"a write to a function declaration", "function x() {} x = 1;"},
		{"a write to a parameter", "function foo(x) { x = 1; }"},
		{"a write to a class binding", "class X {} X = 1;"},
		{"a write to a catch parameter", "try {} catch (x) { x = 1; }"},
		{"a compound write to a shadow from a nested block",
			"const a = 1; { let a = 2; { a += 1; } }"},
		{"a const used as a computed key on another binding",
			"const foo = 1;let bar;bar[foo ?? foo] = 42;"},
		{"a const used as a default value in a destructuring target",
			"const FOO = 1; ({ files = FOO } = arg1); "},
		// using + await using
		{"a using declaration with no write", "using x = foo();"},
		{"an await using declaration with no write", "await using x = foo();"},
		{"a read of a using binding", "using x = foo(); bar(x);"},
		{"a read of an await using binding", "await using x = foo(); bar(x);"},
		{"a type alias with a rest parameter", "type t = (a, ...b) => void"},

		// Beyond upstream, from reading our code.

		// The mask note in the rule doc. `NodeFlagsAwaitUsing` is `Const|Using` rather than a
		// distinct bit, so a port testing individual flags the obvious way misclassifies `let` and
		// `var`. These two assert the negative direction of that mask.
		{"a compound write to a let", "let x = 0; x += 1;"},
		{"an increment of a var", "var x = 0; x++;"},
		// The read side of every write shape the fires table exercises, so a rule reporting on any
		// occurrence rather than on a write fails here.
		{"a const read on the right of an assignment", "const x = 0; let y; y = x;"},
		{"a const read as a call argument in a nested function",
			"const x = 0; function foo() { bar(x); }"},
		{"a const read through a property access chain", "const x = {a:{b:0}}; foo(x.a.b);"},
		// An element write through a const array binding, which mutates the object rather than
		// rebinding the name.
		{"an element write on a const array", "const x = [0]; x[0] = 1;"},
		// A property whose key is spelled like the const. Only the value side of a property
		// assignment binds, so this is a read of nothing.
		{"a property key spelled like the const", "const x = 0; const o = {x: 1}; foo(o, x);"},
		// A shorthand property in an object literal that is not a destructuring target. Textually
		// this is `{x}` exactly as the failing shorthand case is, and only the surrounding
		// construct differs.
		{"a shorthand property in a plain object literal", "const x = 0; const o = {x}; foo(o);"},
		// The outer binding of a nested redeclaration, read rather than written. Pairs with the
		// kind-versus-identity fires case.
		{"a nested redeclaration where neither binding is written",
			"const A = 1; { const A = 2; foo(A); } foo(A);"},
		// The pair to the nested-function fires case. Here the inner binding is a `let`, so it is a
		// real shadow and the write is legal. Identical text to that case apart from one keyword,
		// which is the discrimination this rule exists to make.
		{"a const shadowed by a let in a nested function, only the let written",
			"const A = 1; function f() { let A = 2; A = 3; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoConstAssign, constAssignFile, testCase.sourceText))
		})
	}
}

// The span, which the message-id fixtures above cannot see.
//
// Upstream carries two labels per diagnostic, the declaration and the write, and our engine has one
// range per finding. The write is the one chosen: it is the line the reader has to change, and the
// declaration is usually correct code that some other line is abusing. A port pointing at the
// declaration instead passes every fixture above, because `ExpectFindings` asserts ids and count and
// nothing else.
//
// Sliced out of the source with the finding's own range rather than compared against an offset the
// test computed, since an offset computed by the test is wrong in the same direction as the code
// that produced it.
//
// The destructuring and rest cases are here deliberately. Those are the shapes where a port is most
// likely to report the enclosing pattern or the whole assignment rather than the identifier, and the
// text of a one-character binding makes that invisible unless the span is checked.
func TestNoConstAssignPointsAtTheWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       []string
	}{
		{"a plain reassignment", "const x = 0; x = 1;", []string{"x"}},
		{"a hoisted reassignment", "x = 123; const x = 1;", []string{"x"}},
		{"a compound assignment", "const x = 0; x += 1;", []string{"x"}},
		{"a prefix increment", "const x = 0; ++x;", []string{"x"}},
		{"a postfix increment", "const x = 0; x++;", []string{"x"}},
		{"an increment in a for head", "for (const i = 0; i < 10; ++i) { foo(i); }",
			[]string{"i"}},
		{"a shorthand destructuring target", "const x = 0; ({x} = {x: 1});", []string{"x"}},
		{"a destructuring target with a default", "const x = 0; ({a: x = 1} = {});",
			[]string{"x"}},
		// The rest shapes, which are the ones `ast.IsWriteAccess` declines on its own and which
		// `reference.WritesToBinding` reaches through a spread wrapper. A port reporting the spread element
		// rather than the identifier inside it would still produce one finding with the right id.
		{"a nested rest element as a destructuring target",
			"const d = 123; [a, b, ...[c, ...d]] = [1, 2, 3, 4, 5]", []string{"d"}},
		{"an object rest as a destructuring target",
			"const b = 0; ({a, ...b} = {a: 1, c: 2, d: 3})", []string{"b"}},
		{"a using binding reassigned", "using x = foo(); x = 1;", []string{"x"}},
		{"an await using binding reassigned", "await using x = foo(); x = 1;", []string{"x"}},
		// Two findings in one file, so the ordering of the reported spans is asserted too.
		{"two reassignments of the same binding", "const x = 0; x = 1; x = 2;",
			[]string{"x", "x"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoConstAssign, constAssignFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("expected %d findings, got %d", len(testCase.want),
					len(result.Diagnostics))
			}
			for index, diagnostic := range result.Diagnostics {
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.want[index] {
					t.Errorf("finding %d points at %q, want %q", index, reported,
						testCase.want[index])
				}
			}
		})
	}
}

// The two reported spans in `const x = 0; x = 1; x = 2;` must be distinct positions.
//
// The test above compares the sliced text, and both slices read `x`, so it cannot tell one write
// reported twice from two writes reported once each. A rule looping over the wrong collection
// produces exactly that: the right count, the right ids, and the same span twice.
func TestNoConstAssignReportsDistinctWrites(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoConstAssign, constAssignFile, "const x = 0; x = 1; x = 2;")
	if len(result.Diagnostics) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Range.Pos() == result.Diagnostics[1].Range.Pos() {
		t.Errorf("both findings point at offset %d; the two writes are at different positions",
			result.Diagnostics[0].Range.Pos())
	}
}

// The typed harness is required, and a revert to `rule_testing.Run` must fail loudly here.
//
// This rule declares `NeedsTypeChecker`, and the plain harness hands it a nil checker. The rule
// returns early in that case, so every StaysSilent fixture would pass vacuously and the whole clean
// table would stop measuring anything. Asserting the silence directly is what makes that visible.
func TestNoConstAssignNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoConstAssign, constAssignFile, "const x = 0; x = 1;"))
	rule_testing.ExpectFindings(t,
		rule_testing.RunTyped(t, NoConstAssign, constAssignFile, "const x = 0; x = 1;"),
		"noConstAssign")
}
