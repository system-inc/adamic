package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// funcAssignFile is where the fixtures pretend to live.
//
// A real path matters more here than for a syntactic rule: this rule reads the checker, so the
// harness builds an actual program and the file has to sit somewhere a tsconfig can reach.
const funcAssignFile = "/repository/source/FuncAssign.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_func_assign.rs`:
// one Tester block, 7 pass and 9 fail. The snapshot records 9 diagnostics from those 9 inputs, so
// the extractor reported no discrepancy and one finding per input is right here.
func TestNoFuncAssignFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a reassignment after the declaration", "function foo() {}; foo = bar;",
			[]string{"noFuncAssign"}},
		{"a reassignment from inside the function's own body", "function foo() { foo = bar; }",
			[]string{"noFuncAssign"}},
		{"a reassignment hoisted above the declaration", "foo = bar; function foo() { };",
			[]string{"noFuncAssign"}},
		{"an array destructuring target above the declaration", "[foo] = bar; function foo() { };",
			[]string{"noFuncAssign"}},
		{"an object destructuring target with a default above the declaration",
			"({x: foo = 0} = bar); function foo() { };", []string{"noFuncAssign"}},
		{"an array destructuring target inside the function's own body",
			"function foo() { [foo] = bar; }", []string{"noFuncAssign"}},
		{"a destructuring target in an IIFE that also declares the function",
			"(function() { ({x: foo = 0} = bar); function foo() { }; })();",
			[]string{"noFuncAssign"}},
		// The two function-expression cases. A named function expression binds its own name inside
		// its body, so this is a real reassignment of the function and not of the variable.
		{"a named function expression reassigned inside its own body",
			"var a = function foo() { foo = 123; };", []string{"noFuncAssign"}},
		{"a named function expression assigned to a let, reassigned inside its body",
			"let a = function hello() { hello = 123;};", []string{"noFuncAssign"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoFuncAssign, funcAssignFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// The clean cases are the whole discrimination, and every one of them is a shadow or a variable.
//
// `function foo(foo) { foo = bar; }` writes to a parameter, `function foo() { var foo = bar; }` to
// a local `var` that shadows the function name inside its own body, and `var foo = function() {};
// foo = bar;` to the variable holding an anonymous function expression rather than to any function
// name at all. Their text is nearly identical to the failing forms and only name resolution
// separates them, which is why this rule reads the checker.
func TestNoFuncAssignStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a local var shadowing the function name inside its body",
			"function foo() { var foo = bar; }"},
		{"a parameter shadowing the function name", "function foo(foo) { foo = bar; }"},
		{"a local var declared then written inside the body",
			"function foo() { var foo; foo = bar; }"},
		{"a write to a variable holding an arrow function", "var foo = () => {}; foo = bar;"},
		{"a write to a variable holding an anonymous function expression",
			"var foo = function() {}; foo = bar;"},
		{"a write to the variable from inside an anonymous function expression",
			"var foo = function() { foo = bar; };"},
		{"a local var shadowing the function name in a module",
			"import bar from 'bar'; function foo() { var foo = bar; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoFuncAssign, funcAssignFile, testCase.sourceText))
		})
	}
}

// The span, which the message-id fixtures above cannot see.
//
// Upstream labels the write rather than the declaration, and our engine has one range per finding.
// A port pointing at the function declaration passes every fixture above, because `ExpectFindings`
// asserts ids and count and nothing else.
//
// Sliced out of the source with the finding's own range rather than compared against an offset,
// since an offset computed by the test is wrong in the same direction as the code that produced it.
func TestNoFuncAssignPointsAtTheWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       []string
	}{
		{"a reassignment after the declaration", "function foo() {}; foo = bar;",
			[]string{"foo"}},
		{"a hoisted reassignment", "foo = bar; function foo() { };", []string{"foo"}},
		// A parenthesized assignment target, which neither upstream's corpus nor this one had
		// until a sibling rule tripped over it in the shared write detector. The climb
		// reassigns `child` to the parent, so at the binary node `child` was the parenthesis while
		// the arm compared against `ast.SkipParentheses(binary.Left)`, which is the identifier, and
		// `(foo) = 1` read as a read. The span assertion is the useful half: it pins the finding to
		// the name rather than to the parentheses around it.
		{"a parenthesized target", "function foo() {} (foo) = 1;", []string{"foo"}},
		{"a reassignment from inside the body", "function foo() { foo = bar; }",
			[]string{"foo"}},
		{"an array destructuring target", "[foo] = bar; function foo() { };",
			[]string{"foo"}},
		{"a named function expression reassigned in its body",
			"var a = function foo() { foo = 123; };", []string{"foo"}},
		{"two writes point at their own sites", "function foo() {} foo = 0; foo = 1;",
			[]string{"foo", "foo"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoFuncAssign, funcAssignFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.want), len(result.Diagnostics))
			}
			// The harness trims and appends a newline, so read positions against that same text.
			source := strings.TrimSpace(testCase.sourceText) + "\n"
			for i, want := range testCase.want {
				reported := source[result.Diagnostics[i].Range.Pos():result.Diagnostics[i].Range.End()]
				if reported != want {
					t.Errorf("finding %d pointed at %q, wanted %q", i, reported, want)
				}
			}
		})
	}
}

// The two writes must point at different places, which the fixture above cannot quite prove.
//
// `[]string{"foo", "foo"}` is satisfied by a rule reporting the same identifier twice, and a rule
// looping over the wrong collection does exactly that. This pins that they are distinct offsets.
func TestNoFuncAssignReportsEachWriteAtItsOwnOffset(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoFuncAssign, funcAssignFile, "function foo() {} foo = 0; foo = 1;")
	if len(result.Diagnostics) != 2 {
		t.Fatalf("wanted 2 findings, got %d", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Range.Pos() == result.Diagnostics[1].Range.Pos() {
		t.Fatalf("both findings pointed at offset %d, so one write went unreported",
			result.Diagnostics[0].Range.Pos())
	}
}

// The typed harness is load-bearing, and a revert to `rule_testing.Run` must fail loudly.
//
// This rule declares NeedsTypeChecker, so the plain harness hands it a nil checker and it goes
// completely silent. Every StaysSilent case above would then pass for the wrong reason, and the
// Fires cases would fail in a way that reads as a rule bug rather than as a harness mistake. This
// asserts the silence directly, so the vacuous configuration is a named, tested state rather than
// something a later edit can drift into unnoticed.
func TestNoFuncAssignNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoFuncAssign.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker, so the engine will not lock the file")
	}
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoFuncAssign, funcAssignFile, "function foo() {}; foo = bar;"))
}

// Cases written from reading our code rather than upstream's, each covering a write shape the
// imported corpus never exercises.
//
// Upstream gets these for free: `is_write()` is asked of a reference index that already classified
// every occurrence, so an update expression and a for-of head cost it nothing to support. Ours has
// to name each shape, which means each is a place the port can be silently short, and the corpus
// cannot tell us because it contains no such case.
//
// The update expressions are here specifically. `no_ex_assign.go` shipped without them and neither
// its corpus nor this one exercises `f++`, so the gap would have survived the port on both rules.
//
// The reads are here for the same reason from the other side. `foo.x = 0` writes to a property of
// the function object and leaves the binding alone, and it resolves to the function's symbol
// exactly as a real reassignment does, so symbol identity alone reports it. That one is the reason
// this rule needs structural write detection as well as the checker.
func TestNoFuncAssignCoversWriteShapesUpstreamNeverExercises(t *testing.T) {
	t.Parallel()

	fires := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a compound assignment", "function foo() {} foo += 1;", []string{"noFuncAssign"}},
		{"a logical assignment", "function foo() {} foo ||= 1;", []string{"noFuncAssign"}},
		// The three update shapes. Neither corpus has one.
		{"a postfix increment", "function foo() {} foo++;", []string{"noFuncAssign"}},
		{"a prefix decrement", "function foo() {} --foo;", []string{"noFuncAssign"}},
		{"a postfix increment from inside the function's own body",
			"function foo() { foo++; }", []string{"noFuncAssign"}},
		{"a rest element in an array pattern", "function foo() {} [...foo] = [0];",
			[]string{"noFuncAssign"}},
		{"a rest property in an object pattern", "function foo() {} ({...foo} = {});",
			[]string{"noFuncAssign"}},
		{"a shorthand destructuring target", "function foo() {} ({foo} = {});",
			[]string{"noFuncAssign"}},
		{"a shorthand target carrying a default", "function foo() {} ({foo = 1} = {});",
			[]string{"noFuncAssign"}},
		{"a for-of loop head", "function foo() {} for (foo of []) {}", []string{"noFuncAssign"}},
		{"a for-in loop head", "function foo() {} for (foo in {}) {}", []string{"noFuncAssign"}},
		{"a nested destructuring target", "function foo() {} ({b: {c: foo}} = {b: {}});",
			[]string{"noFuncAssign"}},
		{"a write inside a nested arrow", "function foo() { const f = () => { foo = 1; }; }",
			[]string{"noFuncAssign"}},
		// Two functions in one file, one of them written to. The finding must be one, not two: a
		// rule looping the file per function and matching on text alone reports the write once for
		// each anchor, and `wantIds` of length one is what catches that.
		{"one write with two functions in the file", "function foo() {} function bar() {} bar = 0;",
			[]string{"noFuncAssign"}},
		// A function shadowing another function of the same name, which is the case that separates
		// node identity from declaration kind. Both anchors are KindFunctionDeclaration, so a rule
		// comparing kinds instead of nodes calls the write a match for the outer function as well
		// and reports it twice. Upstream's corpus contains no same-kind shadow pair at all.
		{"a block-scoped function shadowing an outer function of the same name",
			"function foo() {} { function foo() {} foo = 1; }", []string{"noFuncAssign"}},
		{"a nested function shadowing an outer function of the same name",
			"function foo() {} function f() { function foo() {} foo = 1; }",
			[]string{"noFuncAssign"}},
		// A named function expression whose name shadows an outer declaration of the same name.
		// The write binds to the expression's own name, so the finding is one, and the anchor that
		// owns it is the expression rather than the declaration.
		{"a named function expression shadowing an outer function declaration",
			"function foo() {} var a = function foo() { foo = 1; };", []string{"noFuncAssign"}},
	}

	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoFuncAssign, funcAssignFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}

	silent := []struct {
		name       string
		sourceText string
	}{
		{"a property write through the function", "function foo() {} foo.x = 0;"},
		{"a read on the right of an assignment", "function foo() {} let b; b = foo;"},
		{"a call of the function", "function foo() {} foo();"},
		{"a same-named property of an unrelated object",
			"const o = { foo: 1 }; function foo() {} o.foo = 2;"},
		{"a shorthand property in an object being built rather than destructured",
			"function foo() {} const o = {foo};"},
		// The function on the left of a binary expression that is not an assignment. A mutant
		// dropping the assignment-operator gate treats every binary with the name on its left as a
		// write, and survives the whole corpus: upstream has no case reading the function in left
		// position, because its `is_write()` classified the reference and never had to ask about
		// operators.
		{"a comparison with the function on the left", "function foo() {} if (foo === 0) {}"},
		{"an arithmetic use with the function on the left",
			"function foo() {} let b = foo + 1;"},
		// A prefix unary that reads rather than writes. Same gap from the other side: a mutant
		// dropping the update-operator gate calls `!foo` and `-foo` writes, and nothing upstream or
		// in the invented cases above exercises a non-update prefix operator on a function name.
		{"a negation of the function", "function foo() {} let b = !foo;"},
		{"a unary minus on the function", "function foo() {} let b = -foo;"},
		// A `for (const foo of ...)` head declares a fresh binding rather than writing the outer
		// one, so the loop variable is a shadow and not a reassignment.
		{"a for-of head declaring a shadowing binding",
			"function foo() {} for (const foo of []) {}"},
		// An arrow function has no name to reassign, so the variable holding it is the only
		// binding and writing it is clean. This is upstream's `var foo = () => {}` case from the
		// other direction: here the write sits inside the arrow's own body.
		{"a write to the variable from inside an arrow", "var foo = () => { foo = 1; };"},
		// A parameter shadow in a named function expression, the expression-side twin of
		// upstream's `function foo(foo) { foo = bar; }`.
		{"a parameter shadowing a named function expression's own name",
			"var a = function foo(foo) { foo = 1; };"},
	}

	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoFuncAssign, funcAssignFile, testCase.sourceText))
		})
	}
}
