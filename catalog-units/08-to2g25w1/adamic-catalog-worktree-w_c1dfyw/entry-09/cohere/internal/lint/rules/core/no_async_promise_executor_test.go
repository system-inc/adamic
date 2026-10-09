package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// asyncPromiseExecutorFile is where the fixtures pretend to live.
const asyncPromiseExecutorFile = "/repository/source/AsyncPromiseExecutor.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_async_promise_executor.rs`
// lines 116-126: 3 pass, 3 fail. The extractor reports one Tester block and no discrepancy, and the
// snapshot records 3 diagnostics from those 3 fail inputs, so one finding per input is measured
// rather than assumed. Copied because a fixture a porter invents encodes the same belief as the
// port, so it passes for exactly the reason the code is wrong.
//
// This is the smallest corpus in the lane, which is why the added cases below outnumber it.
func TestNoAsyncPromiseExecutorFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an async function expression", "new Promise(async function foo(resolve, reject) {})"},
		{"an async arrow", "new Promise(async (resolve, reject) => {})"},
		{"an async arrow behind five parens", "new Promise(((((async () => {})))))"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoAsyncPromiseExecutor, asyncPromiseExecutorFile, testCase.sourceText),
				"noAsyncPromiseExecutor")
		})
	}
}

// The clean cases are the discrimination, and each fails a different way.
//
// The first is the ordinary executor and must never report. The second is the load-bearing one: an
// async function in argument *position 1* is somebody else's callback, not the executor, so a rule
// scanning all arguments reports it. The third pins that the callee name is part of the predicate.
func TestNoAsyncPromiseExecutorStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a plain executor", "new Promise((resolve, reject) => {})"},
		{"an async function in the second argument", "new Promise((resolve, reject) => {}, async function unrelated() {})"},
		{"a different constructor", "new Foo(async (resolve, reject) => {})"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoAsyncPromiseExecutor, asyncPromiseExecutorFile, testCase.sourceText))
		})
	}
}

// Cases written from reading our code rather than upstream's.
//
// The imported corpus is three fail cases and three pass cases, the smallest in the lane, so it is a
// floor and a low one. Each case below covers a decision this rule makes that no upstream fixture
// touches, and each is a spelling a plausible port gets wrong.
func TestNoAsyncPromiseExecutorFiresOnCasesUpstreamDoesNotCover(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The callee side of the paren question. Upstream pins parens on the *argument* and leaves
		// the callee untested, but `is_specific_id` unwraps the callee too, so upstream reports
		// this. Measured here: without `SkipParentheses` on the callee this parses as a
		// ParenthesizedExpression and the identifier test returns false, so the rule goes silent on
		// a case upstream catches.
		{"parens around the callee", "new (Promise)(async () => {})"},
		{"several parens around the callee", "new (((Promise)))(async () => {})"},
		// The async generator. `ast.IsAsyncFunction` answers false for this while oxc reads
		// `r#async` independently of `generator` and reports it. This is the case that decides
		// between the two ways of asking, and no upstream fixture covers it.
		{"an async generator expression", "new Promise(async function* g(resolve, reject) {})"},
		// An async function expression with no name, which takes a different parse path from the
		// named one upstream ships and shares none of its fixtures.
		{"an anonymous async function expression", "new Promise(async function (resolve, reject) {})"},
		// Whitespace before the keyword. This is the case that catches a span computed as
		// `function.Pos() + 5`, which points at `  asy` here while every corpus case has the
		// function flush against the paren and hides the defect.
		{"whitespace before the keyword", "new Promise(  async () => {})"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoAsyncPromiseExecutor, asyncPromiseExecutorFile, testCase.sourceText),
				"noAsyncPromiseExecutor")
		})
	}
}

// Clean cases written from reading our code, for decisions upstream's three pass cases do not reach.
func TestNoAsyncPromiseExecutorStaysSilentOnCasesUpstreamDoesNotCover(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The call form. Upstream matches `NewExpression` only, so a rule listening on call
		// expressions as well would report this and diverge. Stated in the doc comment as a
		// deliberate gap rather than left to be rediscovered.
		{"a call without new", "Promise(async () => {})"},
		// A spread argument is not an expression oxc can unwrap into a function, so upstream bails.
		// Measured: this parses as a SpreadElement, which the kind gate declines. A rule reaching
		// for the spread's operand would report whatever `args` happens to be.
		{"a spread argument", "new Promise(...args)"},
		// No argument list at all. `Arguments` is nil here rather than an empty list, so a rule
		// trusting a length check to cover it panics rather than missing.
		{"no argument list", "const p = new Promise"},
		{"an empty argument list", "new Promise()"},
		// A member callee. `new Promise.resolve(...)` and `new foo.Promise(...)` both end in the
		// text `Promise` yet neither callee is the identifier, so a rule matching on the name's
		// spelling rather than on the node reports both.
		{"a member expression callee ending in the name", "new foo.Promise(async () => {})"},
		{"a property named Promise on the callee", "new Promise.resolve(async () => {})"},
		// Case matters. The identifier comparison is exact, and a case-insensitive match would
		// report this.
		{"a differently cased callee", "new promise(async () => {})"},
		// A non-async function in position zero with an async one after it, tightening upstream's
		// second pass case: here the executor is a plain *function expression* rather than an
		// arrow, so a rule that gated on kind before position would take a different path.
		{"a plain function executor with an async second argument", "new Promise(function (resolve, reject) {}, async () => {})"},
		// An async function nested inside the executor rather than being the executor. This is
		// legal and common, and a rule walking the executor's subtree for an async modifier
		// reports it.
		{"an async function nested inside a plain executor", "new Promise((resolve, reject) => { void (async () => { await x; })(); })"},
		// An async method on an object literal passed as the executor. Not a function expression or
		// an arrow, so the kind gate declines it, which matches oxc's two-arm match.
		{"an object literal carrying an async method", "new Promise({ async run() {} })"},
		// A class expression carrying an async method, which is the shape that would report if the
		// modifier scan ran without the kind gate.
		{"a class expression with an async method", "new Promise(class { async run() {} })"},
		// A decorated expression, which is the input that separates "find the async modifier" from
		// "take the first modifier". It reaches argument zero carrying a modifier list of exactly
		// one entry, and that entry is the decorator rather than `async`. A scan returning the
		// first modifier reports a finding pointing at `@dec`, which is a false positive on code
		// that has no async executor at all. Found by a surviving mutant rather than by foresight.
		{"a decorated expression whose only modifier is the decorator", "new Promise(@dec async () => {})"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoAsyncPromiseExecutor, asyncPromiseExecutorFile, testCase.sourceText))
		})
	}
}

// The span, which every assertion above is structurally unable to see.
//
// `ExpectFindings` checks message ids and count and nothing else, so a rule pointing at the whole
// function, at the argument list, or at the leading whitespace passes the entire corpus while being
// wrong about the only thing a reader sees. This rule carries no fix, and that is not a reason to
// skip the check: a finding at the wrong line cannot be suppressed by an `eslint-disable-next-line`
// the author is able to write, and nothing in a green suite would say so.
//
// Sliced out of the source with the finding's own range and compared as text, so the assertion
// cannot agree with the rule by sharing its arithmetic.
func TestNoAsyncPromiseExecutorPointsAtTheKeyword(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an async function expression", "new Promise(async function foo(resolve, reject) {})"},
		{"an async arrow", "new Promise(async (resolve, reject) => {})"},
		// Behind parens the keyword is not where the argument starts, which is upstream's own
		// snapshot distinction: it reports column 17 here rather than column 13.
		{"behind five parens", "new Promise(((((async () => {})))))"},
		// Leading whitespace, the case that separates the keyword's trimmed span from its raw
		// position. Measured reporting `  asy` before the span came from `ReportNode`.
		{"with whitespace before the keyword", "new Promise(  async () => {})"},
		// A comment between the paren and the keyword, where the raw node position runs back
		// through the comment and the newline.
		{"with a comment before the keyword", "new Promise(\n  /* c */ async () => {})"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAsyncPromiseExecutor, asyncPromiseExecutorFile,
				testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != "async" {
				t.Fatalf("finding covers %q, wanted %q", reported, "async")
			}
		})
	}
}
