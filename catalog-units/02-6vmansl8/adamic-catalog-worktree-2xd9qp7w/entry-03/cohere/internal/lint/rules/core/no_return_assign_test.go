package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const returnAssignFile = "/repository/source/Thing.ts"

// alwaysReturnOptions is the non-default mode, spelled the way the config spells it.
var alwaysReturnOptions = NoReturnAssignOptions(NoReturnAssignAlways)

// exceptParensReturnOptions is the default, stated explicitly where upstream states it explicitly.
var exceptParensReturnOptions = NoReturnAssignOptions(NoReturnAssignExceptParens)

// Fires. Upstream's whole `invalid` list, each with the mode it was written under.
//
// The codes are generated from the corpus rather than retyped: the extractor loads upstream's test
// file with a stubbed RuleTester, captures the cases object, and emits these rows through a JSON
// encoder, so no escape sequence is ever typed on the way in. Four of these cases carry embedded
// newlines and would have been the ones a heredoc cooked.
//
// A nil option is upstream's own `options: undefined`, which reaches the rule as no options at all
// rather than as the string "except-parens". That distinction is the reason the rule has an
// explicit fallback, and keeping the nil here rather than substituting the default keeps the
// fallback under test through fourteen of these rows.
func TestNoReturnAssignFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
		wantIds    []string
	}{
		{"bare assignment returned", "function x() { return result = a * b; };", nil, []string{"returnAssignment"}},
		{"parenthesized operands are not a parenthesized assignment", "function x() { return (result) = (a * b); };", nil, []string{"returnAssignment"}},
		{"bare assignment returned, explicit except-parens", "function x() { return result = a * b; };", exceptParensReturnOptions, []string{"returnAssignment"}},
		{"parenthesized operands, explicit except-parens", "function x() { return (result) = (a * b); };", exceptParensReturnOptions, []string{"returnAssignment"}},
		{"bare assignment returned from an arrow block body", "() => { return result = a * b; }", nil, []string{"returnAssignment"}},
		{"bare assignment as an arrow body", "() => result = a * b", nil, []string{"arrowAssignment"}},
		{"bare assignment returned, always", "function x() { return result = a * b; };", alwaysReturnOptions, []string{"returnAssignment"}},
		{"parenthesized assignment returned, always", "function x() { return (result = a * b); };", alwaysReturnOptions, []string{"returnAssignment"}},
		{"assignment nested in a logical expression, always", "function x() { return result || (result = a * b); };", alwaysReturnOptions, []string{"returnAssignment"}},
		{"bare assignment returned across lines", "function foo(){\n                return a = b\n            }", nil, []string{"returnAssignment"}},
		{"assignment of a logical expression returned", "function doSomething() {\n                return foo = bar && foo > 0;\n            }", nil, []string{"returnAssignment"}},
		{"assignment of a function expression returned", "function doSomething() {\n                return foo = function(){\n                    return (bar = bar1)\n                }\n            }", nil, []string{"returnAssignment"}},
		{"assignment of an arrow returned", "function doSomething() {\n                return foo = () => a\n            }", nil, []string{"returnAssignment"}},
		{"arrow body assigning an arrow", "function doSomething() {\n                return () => a = () => b\n            }", nil, []string{"arrowAssignment"}},
		{"assignment returned from a nested function expression", "function foo(a){\n                return function bar(b){\n                    return a = b\n                }\n            }", nil, []string{"returnAssignment"}},
		{"curried arrow body assigning", "const foo = (a) => (b) => a = b", nil, []string{"arrowAssignment"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// Stays silent. Upstream's whole `valid` list, each with the mode it was written under.
//
// Every one of these is a false positive upstream already thought about. The three parenthesized
// rows under the default mode are the idiom the option exists to permit, and the two function-body
// rows are the ancestor walk stopping where it should.
func TestNoReturnAssignStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
	}{
		{"module exports object literal", "module.exports = {'a': 1};", nil},
		{"plain assignment statement", "var result = a * b;", nil},
		{"return of a variable", "function x() { var result = a * b; return result; }", nil},
		{"parenthesized assignment returned", "function x() { return (result = a * b); }", nil},
		{"return of a variable, explicit except-parens", "function x() { var result = a * b; return result; }", exceptParensReturnOptions},
		{"parenthesized assignment returned, explicit except-parens", "function x() { return (result = a * b); }", exceptParensReturnOptions},
		{"return of a variable, always", "function x() { var result = a * b; return result; }", alwaysReturnOptions},
		{"assignment inside a returned function expression, always", "function x() { return function y() { result = a * b }; }", alwaysReturnOptions},
		{"parenthesized assignment in an arrow block body", "() => { return (result = a * b); }", exceptParensReturnOptions},
		{"parenthesized assignment as an arrow body", "() => (result = a * b)", exceptParensReturnOptions},
		{"parenthesized assignment inside a sequence arrow body", "const foo = (a,b,c) => ((a = b), c)", nil},
		{"parenthesized assignment returned across lines", "function foo(){\n            return (a = b)\n        }", nil},
		{"parenthesized assignment in a nested function expression", "function bar(){\n            return function foo(){\n                return (a = b) && c\n            }\n        }", nil},
		{"parenthesized assignment as a curried arrow body", "const foo = (a) => (b) => (a = b)", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Where the finding lands, which the message-id fixtures above cannot see.
//
// This rule has two messages pointing at two DIFFERENT nodes: `returnAssignment` anchors on the
// whole return statement, `arrowAssignment` on the whole arrow function. Neither anchors on the
// assignment, which is the span a port would reach for first because it is the node the listener
// receives. Every id fixture above stays green over that mistake, and the two are not close: on
// `function x() { return result = a * b; }` the assignment is `result = a * b` and the finding is
// `return result = a * b;`, semicolon included.
//
// Measured against the installed build, whose columns give the same two spans.
func TestNoReturnAssignReportsTheEnclosingNode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
		want       string
	}{
		{
			"the whole return statement, semicolon included",
			"function x() { return result = a * b; }",
			nil,
			"return result = a * b;",
		},
		{
			// No semicolon to include, so this separates "the statement" from "the statement's text
			// plus whatever follows".
			"a return statement with no semicolon",
			"function f(){ return a = b }",
			nil,
			"return a = b",
		},
		{
			// The parenthesized form reports only under `always`, and it reaches the report through
			// the same branch. The span still covers the statement rather than the parentheses.
			"the whole return statement under always",
			"function x() { return (result = a * b); }",
			alwaysReturnOptions,
			"return (result = a * b);",
		},
		{
			"the whole arrow function",
			"() => result = a * b",
			nil,
			"() => result = a * b",
		},
		{
			// The inner arrow, not the outer one. A rule reporting the first arrow it walks past
			// would point at `(a) => (b) => a = b` and pass every id fixture.
			"the inner arrow of a curried pair",
			"const foo = (a) => (b) => a = b",
			nil,
			"(b) => a = b",
		},
		{
			// A comment before the reported node. `ReportNode` routes through `TokenRange`, which
			// scans past leading trivia, so the span starts at `(` rather than at the comment. A
			// span taken from the node's own Pos() would begin inside the comment.
			"an arrow preceded by a comment",
			"const f = /*c*/ () => a = b;",
			nil,
			"() => a = b",
		},
		{
			// Same trivia question on the other message.
			"a return statement whose value is preceded by a comment",
			"function f(){ return /*c*/ a = b; }",
			nil,
			"return /*c*/ a = b;",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile, testCase.sourceText, testCase.options)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Fatalf("the finding covers %q, wanted %q", reported, testCase.want)
			}
		})
	}
}

// The two messages carry different text, and `ExpectFindings` compares only ids.
//
// A rule reporting `returnAssignment` where upstream reports `arrowAssignment` would be caught by
// the id fixtures, but a rule whose two `rule.Message` values had drifted to share a Description
// would not be, since nothing above reads one. `rule.Message` is `{Id, Description}` with no
// interpolation, so this asserts the two values directly rather than any rendered text.
func TestNoReturnAssignMessagesAreDistinct(t *testing.T) {
	t.Parallel()

	if messageReturnAssignment.Id != "returnAssignment" {
		t.Fatalf("the return message id is %q", messageReturnAssignment.Id)
	}
	if messageArrowAssignment.Id != "arrowAssignment" {
		t.Fatalf("the arrow message id is %q", messageArrowAssignment.Id)
	}
	if messageReturnAssignment.Description == messageArrowAssignment.Description {
		t.Fatal("the two messages share one description, so a reader cannot tell which judgment fired")
	}
	for _, message := range []struct {
		name  string
		value string
	}{
		{"returnAssignment", messageReturnAssignment.Description},
		{"arrowAssignment", messageArrowAssignment.Description},
	} {
		if message.value == "" {
			t.Fatalf("%s has no description", message.name)
		}
	}
}

// Cases upstream's corpus does not cover, each measured against the installed build and each here
// for a reason stated at the case.
//
// Every verdict below was taken by running eslint 10.8.1's own rule, whose file is byte-identical
// to the clone, through the Linter API from ~/Projects/ahra. None was reasoned out from the source.
func TestNoReturnAssignBeyondTheUpstreamCorpus(t *testing.T) {
	t.Parallel()

	t.Run("no options at all falls back to except-parens", func(t *testing.T) {
		t.Parallel()
		// The registry hands a rule its decoded options, and a rule configured with a bare severity
		// gets nothing at all. Upstream's corpus supplies `undefined` on most cases, which its own
		// `defaultOptions` turns into "except-parens" before the rule sees it; ours has no such
		// layer, so the fallback lives in the rule and this is what pins it. Without it, enabling
		// the rule the ordinary way would silently run the strict mode.
		result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, "function f(){ return (a = b); }")
		rule_testing.ExpectClean(t, result)

		result = rule_testing.Run(t, NoReturnAssign, returnAssignFile, "function f(){ return a = b; }")
		rule_testing.ExpectFindings(t, result, "returnAssignment")
	})

	t.Run("an unrecognized mode string is treated as the default", func(t *testing.T) {
		t.Parallel()
		// A typo in the config must not silently escalate to the strict mode. It relaxes to the
		// documented default instead.
		result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile,
			"function f(){ return (a = b); }", NoReturnAssignOptions("alwyas"))
		rule_testing.ExpectClean(t, result)
	})

	// The positional parenthesis test. These are the rows where a structural port over our
	// ParenthesizedExpression node disagrees with upstream, and upstream's corpus writes none of
	// them: it has no call-argument assignment at all.
	t.Run("a call's own parentheses read as the author's wrapping", func(t *testing.T) {
		t.Parallel()
		// SILENT upstream under the default. The assignment has no wrapper of its own; the token
		// before it is the call's `(` and the token after is the call's `)`, which is all
		// `isParenthesised` asks. A structural port reports this.
		result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, "function f(){ return foo(a = b); }")
		rule_testing.ExpectClean(t, result)

		// And the same input under `always`, where the predicate is not consulted at all, reports.
		// That pairing is what proves the silence above comes from the parenthesis test rather than
		// from the walk failing to reach the return.
		result = rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile,
			"function f(){ return foo(a = b); }", alwaysReturnOptions)
		rule_testing.ExpectFindings(t, result, "returnAssignment")
	})

	t.Run("a new expression's parentheses read the same way", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, "function f(){ return new Foo(a = b); }")
		rule_testing.ExpectClean(t, result)
	})

	t.Run("brackets and braces are not parentheses", func(t *testing.T) {
		t.Parallel()
		// The other half of the same predicate. An array literal, a computed member access, and an
		// object property all surround the assignment, and none of them reports as wrapped, because
		// the adjacent tokens are `[`, `[`, and `:`.
		for _, source := range []string{
			"function f(){ return [a = b]; }",
			"function f(){ return a[b = c]; }",
			"function f(){ return {k: a = b}; }",
		} {
			result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, source)
			rule_testing.ExpectFindings(t, result, "returnAssignment")
		}
	})

	t.Run("a sequence inside parentheses is not a parenthesized assignment", func(t *testing.T) {
		t.Parallel()
		// `(a, b = c)` has a real wrapper, but not around the assignment: the token immediately
		// before `b = c` is the comma. Reports upstream. This is the row that separates the
		// positional test from any notion of an enclosing wrapper.
		result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, "function f(){ return (a, b = c); }")
		rule_testing.ExpectFindings(t, result, "returnAssignment")
	})

	t.Run("an assignment leading a parenthesized sequence is not wrapped", func(t *testing.T) {
		t.Parallel()
		// `(a = b, c)` opens with `(` immediately before the assignment, so the first half of the
		// predicate answers yes, and the token AFTER the assignment is a comma rather than the
		// close paren. Reports upstream. This is the only shape in either corpus that exercises the
		// close-paren half on its own: every other reporting row already fails the open-paren half,
		// so a rule checking only the open paren stays green over all of them.
		//
		// Found by a surviving mutant. Weakening `after < len(text) && text[after] == ')'` to an
		// `||` silences this input and nothing else in the suite.
		result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, "function f(){ return (a = b, c); }")
		rule_testing.ExpectFindings(t, result, "returnAssignment")
	})

	t.Run("a parenthesized left operand is not a parenthesized assignment", func(t *testing.T) {
		t.Parallel()
		// `(a) = b` opens with `(` but the token after the assignment is `;`. Reports upstream, and
		// upstream's own corpus carries the sibling shape `(result) = (a * b)`.
		result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, "function f(){ return (a) = b; }")
		rule_testing.ExpectFindings(t, result, "returnAssignment")
	})

	t.Run("a wrapper around only part of the returned expression still counts", func(t *testing.T) {
		t.Parallel()
		// The predicate does not ask whether the wrapper encloses what is returned, only whether it
		// hugs the assignment. Both of these are SILENT upstream under the default even though the
		// returned expression as a whole is unwrapped.
		for _, source := range []string{
			"function f(){ return (a = b) + 1; }",
			"function f(){ return (a = b).c; }",
		} {
			result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, source)
			rule_testing.ExpectClean(t, result)
		}
	})

	t.Run("a parenthesis in a string is not the adjacent token", func(t *testing.T) {
		t.Parallel()
		// The byte comparison in `isPositionallyParenthesized` is only faithful because a `(` inside
		// a literal is never at a token boundary next to the assignment. Here the adjacent token is
		// the comma, and upstream reports.
		result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, `function f(){ return "(" , a = b; }`)
		rule_testing.ExpectFindings(t, result, "returnAssignment")
	})

	t.Run("comments between the parentheses and the assignment are skipped", func(t *testing.T) {
		t.Parallel()
		// Upstream's token reader skips comments, so both of these are SILENT. This is the pair that
		// decides whether the previous token is found at the node's raw Pos() or at its trimmed
		// token start: the trimmed start lands inside the comment's trailing gap and gets both rows
		// wrong.
		for _, source := range []string{
			"function f(){ return foo /*(*/ (a = b); }",
			"function f(){ return ( /*c*/ a = b ); }",
		} {
			result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, source)
			rule_testing.ExpectClean(t, result)
		}
	})

	// The sentinel set. Upstream's regex is a test on ESTree type NAMES, so which nodes stop the
	// walk is not derivable from what a statement is, and these are the rows that pin it.
	t.Run("a class expression stops the walk", func(t *testing.T) {
		t.Parallel()
		// `ClassExpression` is one of the three expression forms named by hand in the regex. The
		// assignment is in a heritage clause, which is inside a return, and upstream is SILENT.
		result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile,
			"function f(){ return class extends (a = b) {}; }", alwaysReturnOptions)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("a class property initializer stops the walk", func(t *testing.T) {
		t.Parallel()
		// Same sentinel, reached through a field rather than a heritage clause.
		result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile,
			"function f(){ return class { p = (a = b); }; }", alwaysReturnOptions)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("a function expression stops the walk from a parameter default", func(t *testing.T) {
		t.Parallel()
		// `FunctionExpression` is the second of the three expression forms named by hand in the
		// regex, and the assignment here is in a parameter default rather than in a body. That
		// matters: a body assignment is an ExpressionStatement and stops one level lower, so a
		// parameter default is the ONLY position where the FunctionExpression case decides anything.
		//
		// Found by a surviving mutant. Upstream's corpus covers the body form twice and the
		// parameter form never, so dropping FunctionExpression from the set is invisible to all
		// thirty imported cases. Measured silent against the installed build.
		for _, source := range []string{
			"function f(){ return function(p = (a = b)){}; }",
			"function f(){ return [function(p = (a = b)){}]; }",
		} {
			result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile, source, alwaysReturnOptions)
			rule_testing.ExpectClean(t, result)
		}
	})

	t.Run("a switch inside an arrow body stops the walk before the arrow", func(t *testing.T) {
		t.Parallel()
		// The plain `switch (a = b) {}` form is silent whether or not SwitchStatement is a sentinel,
		// because the walk above it runs out of ancestors either way. It takes an arrow around the
		// switch to make the difference visible: without the SwitchStatement case the walk climbs to
		// the arrow and reports `arrowAssignment` on code upstream leaves alone.
		//
		// Found by a surviving mutant. Measured silent against the installed build.
		for _, source := range []string{
			"function f(){ return () => { switch (a = b) {} }; }",
			"function f(){ return (() => { switch (a = b) {} })(); }",
		} {
			result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile, source, alwaysReturnOptions)
			rule_testing.ExpectClean(t, result)
		}
	})

	t.Run("a switch statement stops the walk", func(t *testing.T) {
		t.Parallel()
		// `SwitchStatement` matches the statement half of the regex. There is no block below it in
		// the discriminant position, so the SwitchStatement itself is what stops this.
		result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile,
			"function f(){ switch (a = b) { } }", alwaysReturnOptions)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("an arrow parameter default is not an arrow body", func(t *testing.T) {
		t.Parallel()
		// The walk reaches the arrow, but through the parameter list rather than the body, so the
		// body-identity check declines it. Without that check both of these report, and upstream's
		// corpus contains no parameter default at all.
		for _, source := range []string{
			"const f = (a = b) => c",
			"const f = (a = (b = c)) => d",
		} {
			result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile, source, alwaysReturnOptions)
			rule_testing.ExpectClean(t, result)
		}
	})

	t.Run("a parenthesized assignment is still the arrow's body under always", func(t *testing.T) {
		t.Parallel()
		// The walk arrives at the arrow through a ParenthesizedExpression rather than through the
		// assignment itself, so the body-identity check must compare the arrow's body against the
		// child the walk came up THROUGH, not against the assignment it started from. Upstream has
		// no such distinction to make, because its parser discards parentheses and the assignment
		// IS the body node; ours keeps them, so this is a shape upstream's corpus cannot describe.
		//
		// Found by a surviving mutant. Upstream ships both of these codes, but only under the
		// default mode, where the parenthesis guard declines them before the walk runs at all. Under
		// `always` the guard is skipped and the walk decides, and both report `arrowAssignment`
		// against the installed build.
		for _, source := range []string{
			"() => (result = a * b)",
			"const foo = (a,b,c) => ((a = b), c)",
		} {
			result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile, source, alwaysReturnOptions)
			rule_testing.ExpectFindings(t, result, "arrowAssignment")
		}
	})

	t.Run("an arrow block body is not an arrow body assignment", func(t *testing.T) {
		t.Parallel()
		// The block is a sentinel, so the walk stops there and never reaches the arrow. Reporting
		// `arrowAssignment` here would be wrong twice over: the id and the span.
		result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile,
			"const f = () => { a = b };", alwaysReturnOptions)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("an assignment in a statement head inside an arrow block body", func(t *testing.T) {
		t.Parallel()
		// The whole statement half of the sentinel set, exercised where it actually decides
		// something. An assignment in a statement's HEAD (a condition, a loop clause, a throw
		// operand, a with subject) is not an expression statement, so ExpressionStatement cannot
		// stop the walk and each statement kind has to stop it itself. Put the statement inside an
		// arrow's block body and a missing member climbs all the way to the arrow and reports
		// `arrowAssignment`.
		//
		// Found by two surviving mutants, on IfStatement and ForStatement, and generalized here
		// because every member of the statement half has the same exposure. Upstream's corpus writes
		// none of these: it has no loop, no if, no throw, no with, and no try anywhere. All eleven
		// measured silent against the installed build.
		for _, source := range []string{
			"const f = () => { if (a = b) c; };",
			"const f = () => { for (a = b;;) {} };",
			"const f = () => { for (;a = b;) {} };",
			"const f = () => { for (;;a = b) {} };",
			"const f = () => { while (a = b) {} };",
			"const f = () => { do {} while (a = b); };",
			"const f = () => { throw a = b; };",
			"const f = () => { for (const k in (a = b)) {} };",
			"const f = () => { for (const k of [a = b]) {} };",
			// `with` reaches the same way through its subject expression. It is not writable in
			// strict mode, but the parser accepts it and the sentinel decides before any of that.
			"const f = () => { with (a = b) {} };",
			// The only position under a `try` that is not already a statement: a catch binding's
			// destructuring default. Found by a probe after the sweep called TryStatement subsumed
			// on the strength of bodies alone, which was wrong.
			"const f = () => { try { } catch({p = (a = b)}) {} };",
		} {
			result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile, source, alwaysReturnOptions)
			rule_testing.ExpectClean(t, result)
		}
	})

	t.Run("a declarator initializer in an arrow block body", func(t *testing.T) {
		t.Parallel()
		// `var z = a = b` inside an arrow's block body. The assignment sits in a declarator
		// initializer, which is NOT an expression statement, so the ExpressionStatement case cannot
		// stop this walk and the VariableStatement case is what does.
		//
		// This is the shape where our parser and upstream's diverge structurally and land in the
		// same place. Upstream stops at the `BlockStatement`, which is in its sentinel set; we do
		// not carry Block, because it is subsumed everywhere ELSE by ExpressionStatement, and this
		// is the one position where nothing below it would stop the walk. So `VariableStatement`
		// covers exactly the gap Block would have covered, through a different node.
		//
		// Found by a surviving mutant. Dropping VariableStatement makes both of these report
		// `arrowAssignment` on code upstream leaves alone, and no imported fixture can see it: the
		// corpus writes no declarator initializer inside any function body.
		for _, source := range []string{
			"const f = () => { var z = a = b; };",
			"function f(){ return () => { var z = a = b; }; }",
		} {
			result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile, source, alwaysReturnOptions)
			rule_testing.ExpectClean(t, result)
		}
	})

	t.Run("assignments outside any return or arrow body", func(t *testing.T) {
		t.Parallel()
		// The walk runs out of sentinels that report. A labeled statement and a for initializer both
		// stop it at a statement that is neither a return nor an arrow.
		for _, source := range []string{
			"a = b;",
			"function f(){ label: a = b; }",
			"function f(){ for (a = b;;) {} }",
		} {
			result := rule_testing.RunWithOptions(t, NoReturnAssign, returnAssignFile, source, alwaysReturnOptions)
			rule_testing.ExpectClean(t, result)
		}
	})

	t.Run("compound and logical assignment operators", func(t *testing.T) {
		t.Parallel()
		// `IsAssignmentOperator` rather than a comparison against `=`. The logical forms `&&=`,
		// `||=` and `??=` postdate upstream's corpus entirely, and a rule matching only `=` and the
		// arithmetic compounds is silent on all three.
		for _, source := range []string{
			"function f(){ return a += b; }",
			"function f(){ return a ||= b; }",
			"function f(){ return a ??= b; }",
			"function f(){ return a &&= b; }",
		} {
			result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, source)
			rule_testing.ExpectFindings(t, result, "returnAssignment")
		}
	})

	t.Run("destructuring assignment", func(t *testing.T) {
		t.Parallel()
		// An array pattern target is still an assignment and reports; the object form is silent only
		// because it must be wrapped in parentheses to parse at all, which the predicate then reads
		// as deliberate.
		result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, "function f(){ return [a] = b; }")
		rule_testing.ExpectFindings(t, result, "returnAssignment")

		result = rule_testing.Run(t, NoReturnAssign, returnAssignFile, "function f(){ return ({a} = b); }")
		rule_testing.ExpectClean(t, result)
	})

	t.Run("an assignment inside a template substitution", func(t *testing.T) {
		t.Parallel()
		// A template span is not a sentinel and its braces are not parentheses, so the walk reaches
		// the return and the predicate declines. Reports upstream.
		result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, "function f(){ return `${a = b}`; }")
		rule_testing.ExpectFindings(t, result, "returnAssignment")
	})

	t.Run("returns from every function form", func(t *testing.T) {
		t.Parallel()
		// A method, a generator, a conditional return, and a try block. All four reach a
		// ReturnStatement and all four report upstream. The method and generator rows matter because
		// their bodies are Blocks, which ARE sentinels, so the walk must stop at the return before
		// it reaches the block.
		for _, source := range []string{
			"class C { m(){ return a = b; } }",
			"function* g(){ return a = b; }",
			"function f(){ if (x) return a = b; }",
			"function f(){ try { return a = b; } catch (e) {} }",
		} {
			result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, source)
			rule_testing.ExpectFindings(t, result, "returnAssignment")
		}
	})

	t.Run("two assignments in one return statement report twice", func(t *testing.T) {
		t.Parallel()
		// The listener is per assignment, not per return, so one statement holding two assignments
		// produces two findings at the SAME span. Measured upstream, which reports both rows twice
		// at byte-identical columns. A rule reporting per return statement finds one, and every
		// single-assignment fixture above stays green over that.
		for _, source := range []string{
			"function f(){ return a = b, c = d; }",
			"function f(){ return a = b = c; }",
		} {
			result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, source)
			rule_testing.ExpectFindings(t, result, "returnAssignment", "returnAssignment")
		}
	})

	t.Run("an inner return reports at the inner statement", func(t *testing.T) {
		t.Parallel()
		// A returned arrow with a block body holding its own return. The finding belongs to the
		// inner return, and the outer one is not a finding at all.
		const source = "function f(){ return () => { return a = b; }; }"
		result := rule_testing.Run(t, NoReturnAssign, returnAssignFile, source)
		rule_testing.ExpectFindings(t, result, "returnAssignment")
		reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if want := "return a = b;"; reported != want {
			t.Fatalf("the finding covers %q, wanted the inner statement %q", reported, want)
		}
	})
}
