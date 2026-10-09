package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noAwaitInLoopFile is where the fixtures pretend to live.
const noAwaitInLoopFile = "/repository/source/NoAwaitInLoop.ts"

// The corpus is ESLint's own, at `tests/lib/rules/no-await-in-loop.js`, copied rather than
// rewritten: all 20 clean cases and all 17 reporting ones, which is the whole upstream file. The
// rule declares `schema: []`, so no case carries options and there is no decoder to route through.
//
// Every case string was built from a list rather than typed, so nothing could cook an escape on the
// way in, and every verdict was reproduced by driving the installed eslint at 10.8.1 first.
//
// Four of the clean cases are the ones a port is most likely to get wrong, and each is here for a
// different reason: an await on a for-of ITERABLE runs once, a for-await-of that is not itself
// nested is deliberate asynchronous iteration, a plain `using` does not wait, and an `await using`
// in a for-loop INITIALIZER runs once.
func TestNoAwaitInLoopFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an await in a while body", "async function foo() { while (baz) { await bar; } }"},
		{"an await in a while test", "async function foo() { while (await foo()) {  } }"},
		{"a for-await-of nested in a while", "async function foo() { while (baz) { for await (x of xs); } }"},
		{"an await in a for-of body", "async function foo() { for (var bar of baz) { await bar; } }"},
		{"an await as an unbraced for-of body", "async function foo() { for (var bar of baz) await bar; }"},
		{"an await in a for-in body", "async function foo() { for (var bar in baz) { await bar; } }"},
		{"an await in a for body", "async function foo() { for (var i; i < n; i++) { await bar; } }"},
		{"an await in a for test", "async function foo() { for (var i; await foo(i); i++) {  } }"},
		{"an await in a for update", "async function foo() { for (var i; i < n; i = await bar) {  } }"},
		{"an await in a do-while body", "async function foo() { do { await bar; } while (baz); }"},
		{"an await in a do-while test", "async function foo() { do { } while (await bar); }"},
		{"an await deep in a loop body", "async function foo() { while (true) { if (bar) { foo(await bar); } } }"},
		{"an await deep in a loop condition", "async function foo() { while (xyz || 5 > await x) {  } }"},
		{"an await in a while nested inside a for-await-of", "async function foo() { for await (var x of xs) { while (1) await f(x) } }"},
		{"an await using in a while body", "while (true) { await using resource = getResource(); }"},
		{"an await using in a for body", "for (;;) { await using resource = getResource(); }"},
		{"an await using in a for-of left position", "for (await using resource of resources) {}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAwaitInLoop, noAwaitInLoopFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unexpectedAwait")
		})
	}
}

func TestNoAwaitInLoopStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an await outside any loop", "async function foo() { await bar; }"},
		{"an await on the for-in object", "async function foo() { for (var bar in await baz) { } }"},
		{"an await on the for-of iterable", "async function foo() { for (var bar of await baz) { } }"},
		{"an await on a for-await-of iterable", "async function foo() { for await (var bar of await baz) { } }"},
		{"an await in a legacy for-in initializer", "async function foo() { for (var bar = await baz in qux) {} }"},
		{"blocked by a function declaration", "async function foo() { while (true) { async function foo() { await bar; } } }"},
		{"an await in a for-loop initializer", "async function foo() { for (var i = await bar; i < n; i++) {  } }"},
		{"a do-while with no await at all", "async function foo() { do { } while (bar); }"},
		{"blocked by a function expression", "async function foo() { while (true) { var y = async function() { await bar; } } }"},
		{"blocked by a concise arrow", "async function foo() { while (true) { var y = async () => await foo; } }"},
		{"blocked by a block-bodied arrow", "async function foo() { while (true) { var y = async () => { await foo; } } }"},
		{"blocked by a class method", "async function foo() { while (true) { class Foo { async foo() { await bar; } } } }"},
		{"a for-await-of that is not itself nested", "async function foo() { for await (var x of xs) { await f(x) } }"},
		{"a const declaration in a loop", "while (true) { const value = 0; }"},
		{"a let declaration in a loop", "while (true) { let value = 0; }"},
		{"a var declaration in a loop", "while (true) { var value = 0; }"},
		{"an await using outside any loop", "await using resource = getResource();"},
		{"a plain using in a loop", "while (true) { using resource = getResource(); }"},
		{"an await using blocked by a function declaration", "async function foo() { while (true) { async function foo() { await using resource = getResource(); } } }"},
		{"an await using in a for-loop initializer", "for (await using resource = getResource(); ;) {}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAwaitInLoop, noAwaitInLoopFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoAwaitInLoopSpans pins where each of the three waiting shapes points.
//
// The `await using` rows are the reason this test exists. Upstream reports ESTree's
// `VariableDeclaration`, whose span INCLUDES a trailing semicolon in a statement and has none to
// include in a for-of header. Those are two different nodes in this parser, so the rule chooses
// between them, and a port always taking one is off by one character on half the cases while every
// message-id fixture stays green.
//
// The offsets were checked by converting each to a line and column and comparing against the
// installed build, and the semicolon question specifically was settled by running the same source
// with the semicolon removed and watching upstream's end column move.
func TestNoAwaitInLoopSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantStart  int
		wantEnd    int
		wantText   string
	}{
		{"an await using statement keeps its semicolon", "while (true) { await using resource = getResource(); }", 15, 52, "await using resource = getResource();"},
		{"an await using in a for body keeps its semicolon", "for (;;) { await using resource = getResource(); }", 11, 48, "await using resource = getResource();"},
		{"an await using in a for-of header has no semicolon to keep", "for (await using resource of resources) {}", 5, 25, "await using resource"},
		{"a plain await reports the await expression", "async function foo() { while (baz) { await bar; } }", 37, 46, "await bar"},
		{"a nested for-await-of reports the loop", "async function foo() { while (baz) { for await (x of xs); } }", 37, 57, "for await (x of xs);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAwaitInLoop, noAwaitInLoopFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}
			reported := result.Diagnostics[0].Range
			got := testCase.sourceText[reported.Pos():reported.End()]
			if reported.Pos() != testCase.wantStart || reported.End() != testCase.wantEnd {
				t.Fatalf("expected [%d,%d) which is %q, got [%d,%d) which is %q",
					testCase.wantStart, testCase.wantEnd, testCase.wantText,
					reported.Pos(), reported.End(), got)
			}
			if got != testCase.wantText {
				t.Fatalf("expected the finding to cover %q, got %q", testCase.wantText, got)
			}
		})
	}
}

// TestNoAwaitInLoopMessage asserts the message.
//
// Nothing is interpolated, so this is equality on a constant rather than a guard against a format
// string. It is here because the rule has one message and three entry points, and a reworded
// description would otherwise be invisible.
func TestNoAwaitInLoopMessage(t *testing.T) {
	t.Parallel()

	if messageUnexpectedAwaitInLoop.Id != "unexpectedAwait" {
		t.Fatalf("expected id %q, got %q", "unexpectedAwait", messageUnexpectedAwaitInLoop.Id)
	}
	result := rule_testing.Run(t, NoAwaitInLoop, noAwaitInLoopFile,
		"async function foo() { while (baz) { await bar; } }")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Message.Description != messageUnexpectedAwaitInLoop.Description {
		t.Fatalf("the reported description is not the rule's own")
	}
}

// TestNoAwaitInLoopBoundariesBeyondUpstreamsThree covers the function forms upstream reaches through
// ESTree's `FunctionExpression` and this parser gives their own kinds.
//
// Upstream's boundary list names three types and one of them covers several shapes here, so a port
// transcribing those three names literally would report every case below. Its own corpus proves the
// point for a class method; these were measured against the installed build at 10.8.1.
//
// The first draft of this test was wrong in a way worth recording, because it passed. It asserted
// clean for a constructor, a getter, a setter and a static block each containing a bare `await`,
// and all four are clean upstream for a reason that has nothing to do with this rule: `await` is not
// legal in a non-async constructor, accessor or static initialization block, so eslint fails at
// PARSE and no rule runs. That is the brief's "a case whose verdict is decided above the rule",
// arriving as a syntax error rather than as a suppression, and the tell was that eslint returned a
// fatal rather than an empty list. The cases below are legal source, so the boundary is what makes
// them clean, and a firing control was run alongside them to prove the measurement could see a
// report at all.
func TestNoAwaitInLoopBoundariesBeyondUpstreamsThree(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
	}{
		{"an async object method", "async function f(){ while (true) { var o = { async m() { await bar; } }; } }"},
		{"an async class method", "async function f(){ while (true) { class C { async m() { await bar; } } } }"},
		{"an async generator method", "async function f(){ while (true) { class C { async *m() { await bar; } } } }"},
		{"an async arrow returned from a method", "async function f(){ while (true) { class C { m() { return async () => await bar; } } } }"},

		// These four are ILLEGAL source and are here for that reason rather than in spite of it.
		// `await` is not allowed in a non-async constructor, accessor or static initialization
		// block, so eslint fails at parse and upstream can never reach its boundary check with
		// them. typescript-go recovers instead and hands the rule a real await expression inside a
		// real accessor, so the arms naming those kinds are load-bearing HERE and in no upstream
		// test. Measured both ways: with the four kinds removed from the boundary list, every one
		// of these reports; with them present, all four are clean and a firing control still
		// reports. That is the grammar and the parser disagreeing, which is exactly the case a
		// surviving mutant on an arm like this is telling you about.
		{"an await in a constructor, which is illegal and recovered", "async function f(){ while (true) { class C { constructor() { foo(await bar); } } } }"},
		{"an await in a getter, which is illegal and recovered", "async function f(){ while (true) { class C { get x() { return foo(await bar); } } } }"},
		{"an await in a setter, which is illegal and recovered", "async function f(){ while (true) { class C { set x(v) { foo(await bar); } } } }"},
		{"an await in a static block, which is illegal and recovered", "async function f(){ while (true) { class C { static { foo(await bar); } } } }"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAwaitInLoop, noAwaitInLoopFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoAwaitInLoopLoopPositions states the position table in one place, because it is the
// discriminating half of the rule and the corpus covers it unevenly.
//
// A loop header has parts that run once and parts the loop re-enters, and being anywhere inside a
// loop node is not the question. Every row was measured against the installed build at 10.8.1.
func TestNoAwaitInLoopLoopPositions(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
		wantReport bool
	}{
		{"for initializer runs once", "async function f(){ for (var i = await bar; i < n; i++) {} }", false},
		{"for test is re-entered", "async function f(){ for (var i; await foo(i); i++) {} }", true},
		{"for update is re-entered", "async function f(){ for (var i; i < n; i = await bar) {} }", true},
		{"for body is re-entered", "async function f(){ for (var i; i < n; i++) { await bar; } }", true},
		{"for-of iterable runs once", "async function f(){ for (var b of await baz) {} }", false},
		{"for-of body is re-entered", "async function f(){ for (var b of baz) { await b; } }", true},
		{"for-in object runs once", "async function f(){ for (var b in await baz) {} }", false},
		{"for-in body is re-entered", "async function f(){ for (var b in baz) { await b; } }", true},
		{"while test is re-entered", "async function f(){ while (await foo()) {} }", true},
		{"while body is re-entered", "async function f(){ while (baz) { await bar; } }", true},
		{"do-while test is re-entered", "async function f(){ do {} while (await bar); }", true},
		{"do-while body is re-entered", "async function f(){ do { await bar; } while (baz); }", true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAwaitInLoop, noAwaitInLoopFile, testCase.sourceText)
			if testCase.wantReport {
				rule_testing.ExpectFindings(t, result, "unexpectedAwait")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}
