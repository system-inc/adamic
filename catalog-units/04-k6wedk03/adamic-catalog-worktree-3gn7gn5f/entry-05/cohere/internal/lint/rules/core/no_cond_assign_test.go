package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const condAssignFile = "/repository/source/Thing.ts"

// alwaysOptions is the non-default mode, spelled the way the config spells it.
var alwaysOptions = NoCondAssignOptions(NoCondAssignAlways)

// exceptParensOptions is the default, stated explicitly where upstream states it explicitly.
var exceptParensOptions = NoCondAssignOptions(NoCondAssignExceptParens)

// Fires, default mode. Every case here is upstream's `fail` list carrying `None` for options,
// meaning the default `except-parens`.
func TestNoCondAssignFiresUnderExceptParens(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"if test", "var x; if (x = 0) { var b = 1; }"},
		{"while test", "var x; while (x = 0) { var b = 1; }"},
		{"do while test", "var x = 0, y; do { y = x; } while (x = x + 1);"},
		// A compound operator is an assignment too, which is why the operator check asks
		// `IsAssignmentOperator` rather than comparing against `=`.
		{"for test with compound operator", "var x; for(; x+=1 ;){};"},
		// Parenthesized *operands* are not a parenthesized assignment. The assignment itself is
		// still bare in the test position, so the except-parens escape does not apply.
		{"parenthesized operands", "var x; if ((x) = (0));"},
		{"ternary test", "var x; var b = (x = 0) ? 1 : 0;"},
		// Reported under the default even though the assignment IS wrapped in parentheses. See the
		// asymmetry note on the rule: upstream strips parentheses for a ternary test regardless of
		// mode, and only for a ternary test.
		{"parenthesized ternary test", "(((3496.29)).bkufyydt = 2e308) ? foo : bar;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile, testCase.sourceText, exceptParensOptions)
			rule_testing.ExpectFindings(t, result, "condAssign")
		})
	}
}

// Fires, always mode. Upstream's `fail` entries carrying `["always"]`.
//
// Nine of these report TWICE upstream, at a byte-identical span, because its statement branch and
// its assignment branch both fire on a bare assignment in a test position. This asserts one finding
// each. The divergence is deliberate and is argued at the rule.
func TestNoCondAssignFiresUnderAlways(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// Assignment nested inside another expression: only the always mode reaches it.
		{"nested in if", "if (someNode || (someNode = parentNode)) { }"},
		{"nested in while", "while (someNode || (someNode = parentNode)) { }"},
		{"nested in do while", "do { } while (someNode || (someNode = parentNode));"},
		{"nested in for test", "for (; (typeof l === 'undefined' ? (l = 0) : l); i++) { }"},
		// Bare in the test position. Upstream double-reports each of these six.
		{"bare if", "if (x = 0) { }"},
		{"bare while", "while (x = 0) { }"},
		{"bare do while", "do { } while (x = x + 1);"},
		{"bare for", "for(; x = y; ) { }"},
		// Wrapped in parentheses, which the always mode refuses to accept as intent.
		{"parenthesized if", "if ((x = 0)) { }"},
		{"parenthesized while", "while ((x = 0)) { }"},
		{"parenthesized do while", "do { } while ((x = x + 1));"},
		{"parenthesized for", "for(; (x = y); ) { }"},
		{"nested in ternary test", "var x; var b = x && (y = 0) ? 1 : 0;"},
		// The `=` inside the comment is not the operator. This is the case that decides whether the
		// reported span is scanned from the token or taken from the node's position, which includes
		// leading trivia and would cover the comment.
		{"assignment operator after a comment", "while (a /* = */ = b) {}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile, testCase.sourceText, alwaysOptions)
			rule_testing.ExpectFindings(t, result, "condAssign")
		})
	}
}

// Stays silent. Upstream's whole `pass` list, each with the mode it was written under.
func TestNoCondAssignStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    NoCondAssignOptions
	}{
		{"comparison", "var x = 0; if (x == 0) { var b = 1; }", exceptParensOptions},
		{"comparison under always", "var x = 0; if (x == 0) { var b = 1; }", alwaysOptions},
		{"assignment in a loop body", "var x = 5; while (x < 5) { x = x + 1; }", exceptParensOptions},
		// The idiom the default mode exists to permit: parentheses say the assignment was meant.
		{"parenthesized assignment compared", "if ((someNode = someNode.parentNode) !== null) { }", exceptParensOptions},
		{"parenthesized assignment compared, explicit mode", "if ((someNode = someNode.parentNode) !== null) { }", exceptParensOptions},
		{"parenthesized if test", "if ((a = b));", exceptParensOptions},
		{"parenthesized while test", "while ((a = b));", exceptParensOptions},
		{"parenthesized do while test", "do {} while ((a = b));", exceptParensOptions},
		{"parenthesized for test", "for (;(a = b););", exceptParensOptions},
		{"for with no test", "for (;;) {}", exceptParensOptions},
		{"nested parenthesized in if", "if (someNode || (someNode = parentNode)) { }", exceptParensOptions},
		{"nested parenthesized in while", "while (someNode || (someNode = parentNode)) { }", exceptParensOptions},
		{"nested parenthesized in do while", "do { } while (someNode || (someNode = parentNode));", exceptParensOptions},
		{"nested parenthesized in for", "for (;someNode || (someNode = parentNode););", exceptParensOptions},
		// A function boundary stops the ancestor walk. The assignment is inside a body that merely
		// happens to sit in a test position, and it runs when the function is called rather than
		// when the condition is evaluated.
		{"assignment inside an invoked function expression", "if ((function(node) { return node = parentNode; })(someNode)) { }", exceptParensOptions},
		{"assignment inside an invoked function expression, always", "if ((function(node) { return node = parentNode; })(someNode)) { }", alwaysOptions},
		{"assignment inside an invoked arrow", "if ((node => node = parentNode)(someNode)) { }", exceptParensOptions},
		{"assignment inside an invoked arrow, always", "if ((node => node = parentNode)(someNode)) { }", alwaysOptions},
		{"assignment inside a function expression", "if (function(node) { return node = parentNode; }) { }", exceptParensOptions},
		{"assignment inside a function expression, always", "if (function(node) { return node = parentNode; }) { }", alwaysOptions},
		{"plain assignment statement", "x = 0;", alwaysOptions},
		{"comparison in a ternary test", "var x; var b = (x === 0) ? 1 : 0;", exceptParensOptions},
		// A case label is not a conditional test position, even though a switch is a conditional.
		{"switch case label", "switch (foo) { case a = b: bar(); }", exceptParensOptions},
		{"switch case label, always", "switch (foo) { case a = b: bar(); }", alwaysOptions},
		{"switch case label expression, always", "switch (foo) { case baz + (a = b): bar(); }", alwaysOptions},
		// Assignments in the BODY of a conditional, which is where assignments belong. These are
		// the cases the span containment check exists for.
		{"assignment in an if body", "if (obj.key) { (obj.key=false) }", alwaysOptions},
		{"assignment in a for body", "for (;;) { (obj.key=false) }", alwaysOptions},
		{"assignment in a while body", "while (obj.key) { (obj.key=false) }", alwaysOptions},
		{"assignment in a do body", "do { (obj.key=false) } while (obj.key)", alwaysOptions},
		// oxc issue 6656: an unbraced statement body. Without a BlockStatement to stop the walk,
		// a containment check that was merely "is there a conditional ancestor" reports these.
		{"unbraced if and else bodies", "if (['a', 'b', 'c', 'd'].includes(value)) newValue = value;\nelse newValue = 'default';", alwaysOptions},
		{"unbraced while body", "while(true) newValue = value;", alwaysOptions},
		{"unbraced for body", "for(;;) newValue = value;", alwaysOptions},
		{"ternary in a for test under the default", "for (; (typeof l === 'undefined' ? (l = 0) : l); i++) { }", exceptParensOptions},
		// The for initializer and updater are not the test.
		{"for initializer and body", "for (x = 0;x<10;x++) { x = 0 }", exceptParensOptions},
		{"for updater", "for (x = 0;x<10;(x = x + 1)) { x = 0 }", exceptParensOptions},
		{"for updater compound", "const nums = [1,2,3]; for(let i = 0; i < nums.length; i += 1) { dosomething(); }", exceptParensOptions},
		{"for updater compound, always", "for (let i = 0; i < nums.length; i += 1) { dosomething();}", alwaysOptions},
		// The assignment is in a ternary BRANCH, not its test.
		{"assignment in a ternary branch", "let a = 1; a = a ? (a += 5) : 1;", alwaysOptions},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Cases upstream does not cover, each here for a reason stated at the case.
func TestNoCondAssignBeyondTheUpstreamCorpus(t *testing.T) {
	t.Parallel()

	t.Run("no options at all falls back to except-parens", func(t *testing.T) {
		t.Parallel()
		// The registry hands a rule its decoded options, and a rule configured with a bare severity
		// gets nothing at all. Upstream's corpus always supplies a mode, so nothing there pins what
		// happens when the config says only `"error"`. The default has to be the permissive one, or
		// enabling the rule the ordinary way silently turns on the strict mode.
		result := rule_testing.Run(t, NoCondAssign, condAssignFile, "if ((a = b));")
		rule_testing.ExpectClean(t, result)

		result = rule_testing.Run(t, NoCondAssign, condAssignFile, "if (a = b);")
		rule_testing.ExpectFindings(t, result, "condAssign")
	})

	t.Run("an unrecognized mode string is treated as the default", func(t *testing.T) {
		t.Parallel()
		// A typo in the config must not silently escalate to the strict mode. It relaxes to the
		// documented default instead, which is the same direction the decoder fails in.
		result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile, "if ((a = b));", NoCondAssignOptions("alwyas"))
		rule_testing.ExpectClean(t, result)
	})

	t.Run("logical assignment operators in a test position", func(t *testing.T) {
		t.Parallel()
		// `&&=` and `??=` are assignments the ES2021 grammar added, and a rule matching only `=`
		// and the arithmetic compounds misses them. Upstream's corpus predates them.
		for _, source := range []string{"if (a ||= b) { }", "if (a &&= b) { }", "if (a ??= b) { }"} {
			result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile, source, exceptParensOptions)
			rule_testing.ExpectFindings(t, result, "condAssign")
		}
	})

	t.Run("a nested conditional reports each assignment once", func(t *testing.T) {
		t.Parallel()
		// Two assignments in one test position. A rule reporting per statement rather than per
		// assignment finds one of these; a rule that walks ancestors without stopping finds one of
		// them twice. This is the case that separates those two defects.
		result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile,
			"if ((a = b) || (c = d)) { }", alwaysOptions)
		rule_testing.ExpectFindings(t, result, "condAssign", "condAssign")
	})

	t.Run("a for initializer is not the test, even when the test also assigns", func(t *testing.T) {
		t.Parallel()
		// Found by a surviving mutant. Dropping the start bound of the containment check turns the
		// `for` initializer into a second finding, because the initializer ENDS before the test
		// ends and so satisfies the end bound on its own.
		//
		// Upstream's corpus cannot see this. Its `for (x = 0;x<10;x++)` case has an assignment in
		// the initializer and a comparison in the test, so the initializer stays silent whether or
		// not the start bound is checked. It takes an assignment in BOTH positions to separate them,
		// and the finding must land on the test's assignment rather than the initializer's.
		const source = "for (x = 0; y = 1; z++) { }"
		result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile, source, alwaysOptions)
		rule_testing.ExpectFindings(t, result, "condAssign")
		reported := result.Diagnostics[0].Range.Pos()
		if want := 14; reported != want {
			t.Fatalf("the finding starts at offset %d, wanted %d (the test's operator, not the initializer's)", reported, want)
		}
	})

	t.Run("a for initializer assigning into a parenthesized value", func(t *testing.T) {
		t.Parallel()
		// The same bound, reached a different way: here the initializer NESTS an assignment, so a
		// rule missing the start bound reports two findings on a loop whose test assigns nothing.
		result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile,
			"for (a = (b = 1); c; d) { }", alwaysOptions)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("an else-if test is checked", func(t *testing.T) {
		t.Parallel()
		// The else branch of an if is another IfStatement, so this passes only if the listener sees
		// nested statements rather than top-level ones.
		result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile,
			"if (a) { } else if (b = c) { }", exceptParensOptions)
		rule_testing.ExpectFindings(t, result, "condAssign")
	})

	t.Run("a getter body in a test position stops the ancestor walk", func(t *testing.T) {
		t.Parallel()
		// Found by a surviving mutant, and the only input that distinguishes the Block stop from
		// nothing at all. An object literal sitting in an `if` test is truthy and its getter never
		// runs during the test, so the assignment inside it is a body assignment.
		//
		// Upstream's corpus covers the function-expression and arrow forms of this and not the
		// accessor form. Its own stop set names Function and ArrowFunctionExpression, so a port
		// naming only those two would report here; the Block is what actually stops it.
		result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile,
			"if ({ get p(){ a = 1; return 1; } }) { }", alwaysOptions)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("an arrow with an expression body stops the ancestor walk", func(t *testing.T) {
		t.Parallel()
		// The one function form with no Block to stop at, which is why ArrowFunction is in the stop
		// set and the four other function kinds are not.
		result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile,
			"if ((() => a = 1)()) { }", alwaysOptions)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("an object method body in a test position", func(t *testing.T) {
		t.Parallel()
		// Same shape as the getter, reached through an ordinary method. Stopped by the Block rather
		// than by any method-specific case.
		result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile,
			"if ({ m(){ a = 1; } }) { }", alwaysOptions)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("a class body stops the ancestor walk", func(t *testing.T) {
		t.Parallel()
		// Upstream stops at Function, ArrowFunction, Program, and BlockStatement. A method body is
		// a BlockStatement so it stops, but this pins that a whole class expression sitting in a
		// test position does not leak its assignments into the condition.
		result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile,
			"if (class { m() { let a; a = 1; return a; } }) { }", alwaysOptions)
		rule_testing.ExpectClean(t, result)
	})
}

// Where the finding lands, which the message-id fixtures above cannot see.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule reporting the whole assignment
// rather than just its operator passes every fixture above. Upstream points at the operator token
// specifically, and the two differ by the whole left operand.
func TestNoCondAssignReportsTheOperator(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    NoCondAssignOptions
		want       string
	}{
		{"simple assignment", "if (x = 0) { }", exceptParensOptions, "="},
		// A compound operator is more than one character, so a span hardcoded to width one truncates
		// it and a span taken from the node covers the operand.
		{"compound assignment", "var x; for(; x+=1 ;){};", exceptParensOptions, "+="},
		{"logical assignment", "if (a ??= b) { }", exceptParensOptions, "??="},
		// The decisive one. A node's position runs from before its leading trivia, so a span built
		// from `OperatorToken.Pos()` covers ` /* = */ =` and points the finding at the comment.
		{"operator preceded by a comment", "while (a /* = */ = b) {}", alwaysOptions, "="},
		// Reported through the ancestor walk rather than the test position, which is a different
		// code path and could compute a different span.
		{"nested assignment", "if (someNode || (someNode = parentNode)) { }", alwaysOptions, "="},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile, testCase.sourceText, testCase.options)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Fatalf("the finding covers %q, wanted the operator %q", reported, testCase.want)
			}
		})
	}
}

// The `while (a /* = */ = b) {}` case above pins the comment, and this pins the column upstream
// snapshots for it, so a span that is right in text but shifted in position still fails.
func TestNoCondAssignOperatorSpanMatchesUpstreamOffset(t *testing.T) {
	t.Parallel()

	const source = "while (a /* = */ = b) {}"
	result := rule_testing.RunWithOptions(t, NoCondAssign, condAssignFile, source, alwaysOptions)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	// Upstream's snapshot reports column 18, one-based, which is offset 17 here.
	if got := result.Diagnostics[0].Range.Pos(); got != 17 {
		t.Fatalf("the finding starts at offset %d, wanted 17 to match the upstream column", got)
	}
}
