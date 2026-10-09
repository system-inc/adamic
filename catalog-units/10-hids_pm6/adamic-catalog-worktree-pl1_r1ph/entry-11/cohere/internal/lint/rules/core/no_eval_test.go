package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// evalFile is where the fixtures pretend to live.
const evalFile = "/repository/source/Eval.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_eval.rs`: 64 pass,
// 52 fail, and the snapshot records 54 diagnostics from 52 fail inputs. The two extra come from
// `eval(eval)` and `eval?.(eval)`, each of which reports twice: once for the callee and once for
// the argument. A fixture asserting one finding per input would have been wrong on exactly those
// two, which is what the extractor's discrepancy line was telling me.
//
// Cases upstream gates behind an `env` configuration we do not have are marked at the case. See the
// divergence note on the rule itself.

// TestNoEvalRequiresTypedHarness exists because a revert to rule_testing.Run would be silent.
//
// The rule declares NeedsTypeChecker, and the plain harness hands it a nil checker. Every clean case
// below would then pass vacuously while the rule reported nothing at all. This asserts the
// difference is observable: a direct call fires under the typed harness, and the whole rule goes
// quiet without it.
func TestNoEvalRequiresTypedHarness(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.RunTyped(t, NoEval, evalFile, "var EVAL = eval; EVAL('foo')"), "noEval")
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoEval, evalFile, "var EVAL = eval; EVAL('foo')"))
}

func TestNoEvalFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a direct call", "eval(foo)", []string{"noEval"}},
		{"a direct call on a literal", "eval('foo')", []string{"noEval"}},
		// Two findings: the callee is a direct call, the argument is an indirect reference.
		{"a direct call taking eval", "eval(eval)", []string{"noEval", "noEval"}},
		{"an optional call taking eval", "eval?.(eval)", []string{"noEval", "noEval"}},
		{"a non-null asserted call", "eval!('foo')", []string{"noEval"}},
		{"an as-expression call", "(eval as any)('foo')", []string{"noEval"}},
		{"a property read off eval", "eval.toString()", []string{"noEval"}},
		{"a parameter shadowing eval, still called", "function foo(eval) { eval('foo') }", []string{"noEval"}},
		{"the comma-sequence indirection", "(0, eval)('foo')", []string{"noEval"}},
		{"an alias assignment", "var EVAL = eval; EVAL('foo')", []string{"noEval"}},
		{"eval passed as an argument", "(function(exe){ exe('foo') })(eval);", []string{"noEval"}},
		{"eval passed to a callback taker", "callbacks.findLastIndex(function (cb) { return cb(eval); }, this);", []string{"noEval"}},

		// The global-object arm. Upstream gates these behind `env: { browser: true }` or
		// `env: { node: true }`, which our configuration surface does not have; see the divergence
		// note on the rule.
		{"a window property call", "window.eval('foo')", []string{"noEval"}},
		{"a doubled window property call", "window.window.eval('foo')", []string{"noEval"}},
		{"a doubled window subscript", "window.window['eval']('foo')", []string{"noEval"}},
		{"a global property call", "global.eval('foo')", []string{"noEval"}},
		{"a doubled global property call", "global.global.eval('foo')", []string{"noEval"}},
		{"a doubled global template subscript", "global.global[`eval`]('foo')", []string{"noEval"}},
		{"a globalThis property call", "globalThis.eval('foo')", []string{"noEval"}},
		{"a doubled globalThis property call", "globalThis.globalThis.eval('foo')", []string{"noEval"}},
		{"a doubled globalThis subscript", "globalThis.globalThis['eval']('foo')", []string{"noEval"}},
		{"a sequenced window property", "(0, window.eval)('foo')", []string{"noEval"}},
		{"a sequenced window subscript", "(0, window['eval'])('foo')", []string{"noEval"}},
		{"a sequenced globalThis property", "(0, globalThis.eval)('foo')", []string{"noEval"}},
		{"a sequenced globalThis subscript", "(0, globalThis['eval'])('foo')", []string{"noEval"}},
		{"a globalThis alias assignment", "var EVAL = globalThis.eval; EVAL('foo')", []string{"noEval"}},
		{"an optional window property call", "window?.eval('foo')", []string{"noEval"}},
		{"a parenthesised optional window property", "(window?.eval)('foo')", []string{"noEval"}},
		{"an optional doubled window property", "(window?.window).eval('foo')", []string{"noEval"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoEval, evalFile, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoEvalStaysSilent holds every upstream pass case this port is able to represent.
//
// The clean cases are the whole discrimination. `Eval(foo)` is a different name. `setTimeout('foo')`
// evaluates a string but is not this rule's business. The shadow cases are the ones that decide
// whether the rule resolves names or merely spells them: `function foo() { var eval = 'foo';
// window[eval]('foo') }` subscripts with a *variable* holding the string, so the property read is
// whatever that variable names and is not statically `eval`.
func TestNoEvalStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a different name", "Eval(foo)"},
		{"setTimeout with a string", "setTimeout('foo')"},
		{"setInterval with a string", "setInterval('foo')"},
		{"window.setTimeout", "window.setTimeout('foo')"},
		{"window.setInterval", "window.setInterval('foo')"},
		{"a near-miss window property", "window.noeval('foo')"},
		{"a near-miss global property", "global.noeval('foo')"},
		{"a near-miss globalThis property", "globalThis.noneval('foo')"},
		{"a near-miss this property", "this.noeval('foo');"},

		// The shadow cases. A computed subscript through a *variable* is not a static `eval`
		// property, whatever that variable happens to hold.
		{"a window subscript through a shadowing variable", "function foo() { var eval = 'foo'; window[eval]('foo') }"},
		{"a global subscript through a shadowing variable", "function foo() { var eval = 'foo'; global[eval]('foo') }"},
		{"a globalThis subscript through a shadowing variable", "function foo() { var eval = 'foo'; globalThis[eval]('foo') }"},

		// `this` inside a method, an object literal function, or a class body is not the global
		// object, so `this.eval` there is a call on the receiver rather than global eval.
		{"a strict function's this", "function foo() { 'use strict'; this.eval('foo'); }"},
		{"an object literal method", "var obj = {foo: function() { this.eval('foo'); }}"},
		{"a function assigned to a member", "var obj = {}; obj.foo = function() { this.eval('foo'); }"},
		{"a function returned from an IIFE", "obj.foo = (function() { return function() { this.eval('foo'); }; })()"},
		{"a function returned from an arrow IIFE", "obj.foo = (() => function() { this.eval('foo'); })()"},
		{"an arrow inside a strict function", "function f() { 'use strict'; () => { this.eval('foo') } }"},
		{"an arrow inside a strict function expression", "(function f() { 'use strict'; () => { this.eval('foo') } })"},
		{"a class heritage function", "class C extends function () { this.eval('foo'); } {}"},
		{"a class method", "class A { foo() { this.eval(); } }"},
		{"a static class method", "class A { static foo() { this.eval(); } }"},
		{"a class field initializer", "class A { field = this.eval(); }"},
		{"an arrow class field initializer", "class A { field = () => this.eval(); }"},
		{"a static class field initializer", "class A { static field = this.eval(); }"},
		{"an arrow static class field initializer", "class A { static field = () => this.eval(); }"},
		{"an accessor field initializer", "class A { accessor field = this.eval(); }"},
		{"a static accessor arrow field", "class A { static accessor field = () => this.eval(); }"},
		{"a class static block", "class A { static { this.eval(); } }"},

		// A function passed a `thisArg` is invoked with that receiver, so `this.eval` inside it is
		// not global eval. Upstream enumerates the standard-library methods that take one.
		{"findLast with a thisArg", "array.findLast(function (x) { return this.eval.includes(x); }, { eval: ['foo', 'bar'] });"},
		{"findLastIndex with this as thisArg", "callbacks.findLastIndex(function (cb) { return cb(this.eval); }, this);"},
		{"flatMap with a thisArg", "['1+1'].flatMap(function (str) { return this.eval(str); }, new Evaluator);"},
		{"Array.from with a thisArg", "Array.from(values, function (value) { return this.eval(value); }, context);"},
		{"Array.fromAsync with a thisArg", "Array.fromAsync(values, async function (value) { return this.eval(await value); }, context);"},
		{"Uint8Array.from with a thisArg", "Uint8Array.from(values, function (value) { return this.eval(value); }, context);"},

		// A top-level `this` in a module is undefined rather than the global object. Our tree is
		// modules, so this is the shape a real file has and it must not fire.
		{"a module top-level this", "export {}; this.eval('foo');"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoEval, evalFile, testCase.sourceText))
		})
	}
}

// TestNoEvalAllowIndirect covers the one option, whose schema ESLint states as
// `{ allowIndirect: boolean }` and which defaults to false.
//
// With it on, only a *direct*, non-optional call to `eval` is reported. Every indirection is
// permitted, because an indirect eval cannot reach the calling scope and so is a far smaller
// hazard than the direct form.
func TestNoEvalAllowIndirect(t *testing.T) {
	t.Parallel()

	allowed := []string{
		"(0, eval)('foo')",
		"(0, window.eval)('foo')",
		"(0, window['eval'])('foo')",
		"var EVAL = eval; EVAL('foo')",
		"var EVAL = this.eval; EVAL('foo')",
		"(function(exe){ exe('foo') })(eval);",
		"window.window.eval('foo')",
		"window.window['eval']('foo')",
		"global.eval('foo')",
		"global.global.eval('foo')",
		"this.eval('foo')",
		"function foo() { this.eval('foo') }",
		"(0, globalThis.eval)('foo')",
		"(0, globalThis['eval'])('foo')",
		"var EVAL = globalThis.eval; EVAL('foo')",
		"function foo() { globalThis.eval('foo') }",
		"globalThis.globalThis.eval('foo');",
		"'use strict'; this.eval('foo');",
		"this.eval('foo');",
		"function foo() { this.eval('foo'); }",
		"() => { this.eval('foo') }",
		"window?.eval('foo')",
		"(window?.eval)('foo')",
		// An optional call is not a direct eval: the spec routes it through optional-chaining
		// evaluation rather than step 6.a.vi of the call runtime semantics.
		"eval?.('foo')",
	}
	for _, sourceText := range allowed {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(
				t, NoEval, evalFile, sourceText, NoEvalOptions{AllowIndirect: true}))
		})
	}

	// The direct call is still reported with the option on. That is the whole point of the option
	// being narrower than off.
	t.Run("a direct call is still reported", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(
			t, NoEval, evalFile, "eval('foo')", NoEvalOptions{AllowIndirect: true}), "noEval")
	})
}

// TestNoEvalSpans asserts where each finding points, which ExpectFindings cannot see.
//
// This rule carries no repair, and the brief is explicit that the span still needs asserting: a rule
// whose defect is where it points passes a complete fixture pair while being wrong. The spans below
// are read off oxc's own snapshot, which prints the reported column for every diagnostic.
func TestNoEvalSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       []string
	}{
		// The callee identifier, not the whole call.
		{"a direct call points at the callee", "eval('foo')", []string{"eval"}},
		// The non-null assertion and the as-expression are part of the callee, so upstream's span
		// covers them: snapshot columns 1:1 spanning `eval!` and `(eval as any)`.
		{"a non-null assertion keeps the operator", "eval!('foo')", []string{"eval!"}},
		{"an as-expression keeps the parentheses", "(eval as any)('foo')", []string{"(eval as any)"}},
		// Two findings, argument first then callee, matching the snapshot's order.
		{"eval(eval) points at both", "eval(eval)", []string{"eval", "eval"}},
		// A property access points at the property, not the object and not the whole member.
		{"a member access points at the property", "window.eval('foo')", []string{"eval"}},
		{"a doubled member points at the last property", "window.window.eval('foo')", []string{"eval"}},
		// A subscript points at the literal including its quotes, per the snapshot's six-column span
		// over `'eval'`.
		{"a subscript points at the literal", "window.window['eval']('foo')", []string{"'eval'"}},
		{"a template subscript points at the template", "global.global[`eval`]('foo')", []string{"`eval`"}},
		// An indirect reference points at the identifier itself.
		{"an alias points at the reference", "var EVAL = eval; EVAL('foo')", []string{"eval"}},
		{"a property read points at the callee", "eval.toString()", []string{"eval"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoEval, evalFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.want))
			}
			for index, want := range testCase.want {
				reported := testCase.sourceText[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != want {
					t.Errorf("finding %d points at %q, want %q", index, reported, want)
				}
			}
		})
	}
}
