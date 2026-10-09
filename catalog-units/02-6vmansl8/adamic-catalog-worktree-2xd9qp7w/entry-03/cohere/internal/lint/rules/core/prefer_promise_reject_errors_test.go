package core_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/core"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is ESLint's own, taken verbatim from
// /tmp/lint-sources/eslint/tests/lib/rules/prefer-promise-reject-errors.js by evaluating its
// arrays and emitting each source through a JSON encoder, so no case was retyped and no escape
// sequence was typed on the way in.
//
// 35 valid and 39 invalid upstream. Two valid cases are not reproduced here and are recorded in
// TestGlobalsOffCasesHaveNoHarnessCounterpart below.

func TestPreferPromiseRejectErrorsStaysSilent(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		// ---- upstream valid, no options: 31 cases ----
		"Promise.resolve(5)",
		"Foo.reject(5)",
		"Promise.reject(foo)",
		"Promise.reject(foo.bar)",
		"Promise.reject(foo.bar())",
		"Promise.reject(new Error())",
		"Promise.reject(new TypeError)",
		"Promise.reject(new Error('foo'))",
		"Promise.reject(foo || 5)",
		"Promise.reject(5 && foo)",
		"new Foo((resolve, reject) => reject(5))",
		"new Promise(function(resolve, reject) { return function(reject) { reject(5) } })",
		"new Promise(function(resolve, reject) { if (foo) { const reject = somethingElse; reject(5) } })",
		"new Promise(function(resolve, {apply}) { apply(5) })",
		"new Promise(function(resolve, reject) { resolve(5, reject) })",
		"async function foo() { Promise.reject(await foo); }",
		"Promise.reject(obj?.foo)",
		"Promise.reject(obj?.foo())",
		"Promise.reject(foo = new Error())",
		"Promise.reject(foo ||= 5)",
		"Promise.reject(foo.bar ??= 5)",
		"Promise.reject(foo[bar] ??= 5)",
		"class C { #reject; foo() { Promise.#reject(5); } }",
		"class C { #error; foo() { Promise.reject(this.#error); } }",
		"let Promise; Promise.reject('x');",
		"function f() { Promise.reject('x'); var Promise; }",
		"function f(Promise) { return Promise.reject('x'); }",
		"{ class Promise { static reject(x) { return x; } } Promise.reject('x'); }",
		"function g(Promise) { return new Promise((resolve, reject) => { reject('x'); }); }",
		"function f(undefined) { Promise.reject(undefined); }",
		"function f(undefined) { new Promise((resolve, reject) => reject(undefined)); }",
	} {
		result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", source)
		rule_testing.ExpectClean(t, result)
	}
}

func TestPreferPromiseRejectErrorsFires(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		// ---- upstream invalid, no options: 36 cases ----
		"Promise.reject(5)",
		"Promise.reject('foo')",
		"Promise.reject(`foo`)",
		"Promise.reject(!foo)",
		"Promise.reject(void foo)",
		"Promise.reject()",
		"Promise.reject(undefined)",
		"Promise.reject({ foo: 1 })",
		"Promise.reject([1, 2, 3])",
		"Promise.reject('foo', somethingElse)",
		"new Promise(function(resolve, reject) { reject(5) })",
		"new Promise((resolve, reject) => { reject(5) })",
		"new Promise((resolve, reject) => reject(5))",
		"new Promise((resolve, reject) => reject())",
		"new Promise(function(yes, no) { no(5) })",
		"\n          new Promise((resolve, reject) => {\n            fs.readFile('foo.txt', (err, file) => {\n              if (err) reject('File not found')\n              else resolve(file)\n            })\n          })\n        ",
		"new Promise(({foo, bar, baz}, reject) => reject(5))",
		"new Promise(function(reject, reject) { reject(5) })",
		"new Promise(function(foo, arguments) { arguments(5) })",
		"new Promise((foo, arguments) => arguments(5))",
		"new Promise(function({}, reject) { reject(5) })",
		"new Promise(({}, reject) => reject(5))",
		"new Promise((resolve, reject, somethingElse = reject(5)) => {})",
		"Promise.reject?.(5)",
		"Promise?.reject(5)",
		"Promise?.reject?.(5)",
		"(Promise?.reject)(5)",
		"(Promise?.reject)?.(5)",
		"Promise.reject(foo += new Error())",
		"Promise.reject(foo -= new Error())",
		"Promise.reject(foo **= new Error())",
		"Promise.reject(foo <<= new Error())",
		"Promise.reject(foo |= new Error())",
		"Promise.reject(foo &= new Error())",
		"Promise.reject(foo && 5)",
		"Promise.reject(foo &&= 5)",
	} {
		result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", source)
		// Every invalid case in upstream's corpus carries exactly one error, because the tester
		// maps a single `rejectAnError` entry onto all of them.
		rule_testing.ExpectFindings(t, result, "rejectAnError")
	}
}

// The option cases are routed through the rule's own registered decoder rather than by building
// the options struct, so the wire shape is under test and not only the struct.
//
// The wire shape is the bare object. Upstream writes `options: [{ allowEmptyReject: true }]`, and
// cohere's config layer unwraps the severity tuple before dispatch, so the decoder receives what is
// inside the array.
func decodePreferPromiseRejectErrorsOptions(t *testing.T, wire string) any {
	t.Helper()
	decode := rule.DecodeOptionsInto[core.PreferPromiseRejectErrorsOptions]()
	decoded, err := decode(json.RawMessage(wire))
	if err != nil {
		t.Fatalf("decoding %s: %v", wire, err)
	}
	return decoded
}

func TestPreferPromiseRejectErrorsAllowEmptyReject(t *testing.T) {
	t.Parallel()

	allowed := decodePreferPromiseRejectErrorsOptions(t, `{"allowEmptyReject": true}`)
	denied := decodePreferPromiseRejectErrorsOptions(t, `{"allowEmptyReject": false}`)

	// allowEmptyReject exempts the no-argument call, in both the static and the executor shape.
	for _, source := range []string{
		"Promise.reject()",
		"new Promise(function(resolve, reject) { reject() })",
	} {
		result := rule_testing.RunTypedWithOptions(t, core.PreferPromiseRejectErrors, "input.ts", source, allowed)
		rule_testing.ExpectClean(t, result)
	}

	// Written out explicitly rather than relying on the default, because upstream's corpus carries
	// both spellings and they must not diverge.
	for _, source := range []string{
		"Promise.reject()",
		"new Promise(function(resolve, reject) { reject() })",
	} {
		result := rule_testing.RunTypedWithOptions(t, core.PreferPromiseRejectErrors, "input.ts", source, denied)
		rule_testing.ExpectFindings(t, result, "rejectAnError")
	}

	// The option exempts only the EMPTY call. An explicit `undefined` still reports under it, which
	// is upstream's own case and the reason the argument-count test comes before the option test.
	result := rule_testing.RunTypedWithOptions(t, core.PreferPromiseRejectErrors, "input.ts",
		"Promise.reject(undefined)", allowed)
	rule_testing.ExpectFindings(t, result, "rejectAnError")
}

// A rule configured as a bare "error" is handed nil options, and `options.(T)` on nil yields the
// zero value. That is correct here only because upstream's default for allowEmptyReject is false,
// so this pins the coincidence rather than trusting it.
func TestPreferPromiseRejectErrorsDefaultsWithoutADecoder(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", "Promise.reject()")
	rule_testing.ExpectFindings(t, result, "rejectAnError")

	decoded := decodePreferPromiseRejectErrorsOptions(t, `{}`)
	settings, ok := decoded.(core.PreferPromiseRejectErrorsOptions)
	if !ok {
		t.Fatalf("the decoder returned %T rather than the options type", decoded)
	}
	if settings.AllowEmptyReject {
		t.Errorf("an empty options object decoded allowEmptyReject to true, inverting the default")
	}
}

// Upstream turns Promise off through ESLint's `globals` configuration, in two of its valid cases:
//
//	"/* global Promise:off */ Promise.reject('x')"
//	{ code: "Promise.reject('x')", languageOptions: { globals: { Promise: "off" } } }
//
// cohere has no globals surface, so neither input can be expressed and both would REPORT here. That
// is a fact about the harness rather than about the rule: what those cases assert is that a
// non-global Promise is exempt, and the five source-level shadowing cases in the silent list assert
// exactly the same judgment through a mechanism this tree does have.
//
// Recorded as a test rather than dropped, so the next reader sees which two cases are missing and
// why, instead of counting 33 against upstream's 35 and wondering.
func TestGlobalsOffCasesHaveNoHarnessCounterpart(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts",
		"/* global Promise:off */ Promise.reject('x');")
	if len(result.Diagnostics) != 1 {
		t.Errorf("expected the globals comment to be inert here, giving 1 finding, got %d",
			len(result.Diagnostics))
	}

	// The control: the same judgment expressed the way this tree can express it stays silent, which
	// is what makes the divergence a harness gap rather than a rule gap.
	shadowed := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts",
		"let Promise; Promise.reject('x');")
	rule_testing.ExpectClean(t, shadowed)
}

// Where does the finding point, and what does it say?
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule reporting the right thing at
// the wrong node passes a complete fixture pair while being wrong. Upstream anchors on the whole
// CallExpression rather than on the argument, which is worth pinning because the argument is the
// thing at fault and is the natural place to point.
func TestPreferPromiseRejectErrorsSpanAndMessage(t *testing.T) {
	t.Parallel()

	rows := []struct {
		source string
		want   string
	}{
		{"Promise.reject(5)", "Promise.reject(5)"},
		{"Promise.reject()", "Promise.reject()"},
		{"(Promise?.reject)?.(5)", "(Promise?.reject)?.(5)"},
		{"new Promise((resolve, reject) => reject(5))", "reject(5)"},
		{"new Promise((resolve, reject, somethingElse = reject(5)) => {})", "reject(5)"},
	}
	for _, row := range rows {
		result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", row.source)
		if len(result.Diagnostics) != 1 {
			t.Errorf("%s: expected 1 finding, got %d", row.source, len(result.Diagnostics))
			continue
		}
		diagnostic := result.Diagnostics[0]

		// Sliced from the source the HARNESS wrote, not from the literal above: RunTyped trims the
		// fixture before writing it, so slicing the literal would be off by one on any case with
		// leading whitespace.
		written := strings.TrimSpace(row.source) + "\n"
		reported := written[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != row.want {
			t.Errorf("%s: the finding points at %q, expected %q", row.source, reported, row.want)
		}
	}

	// The message text, asserted by equality against a literal typed here rather than against the
	// rule's own constant, which would move with any mutation of it.
	result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", "Promise.reject(5)")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Diagnostics))
	}
	wantMessage := "Expected the Promise rejection reason to be an Error. A rejection carrying a " +
		"string or a plain object arrives at the catch with no stack, so the line that failed is " +
		"lost and the handler cannot tell one failure from another."
	if result.Diagnostics[0].Message.Description != wantMessage {
		t.Errorf("message was %q", result.Diagnostics[0].Message.Description)
	}
	if result.Diagnostics[0].Message.Id != "rejectAnError" {
		t.Errorf("message id was %q, expected rejectAnError", result.Diagnostics[0].Message.Id)
	}
}

// The typed harness is required. A rule declaring NeedsTypeChecker and handed a nil checker goes
// silent rather than crashing, so every Fires case would fail and every StaysSilent case would pass
// vacuously. Pinned so a later revert of the guard fails loudly.
func TestPreferPromiseRejectErrorsDeclinesWithoutAChecker(t *testing.T) {
	t.Parallel()

	if !core.PreferPromiseRejectErrors.NeedsTypeChecker {
		t.Errorf("the rule stopped declaring NeedsTypeChecker, so the registration path will hand it files with no checker")
	}
	untyped := rule_testing.Run(t, core.PreferPromiseRejectErrors, "input.ts", "Promise.reject(5)")
	rule_testing.ExpectClean(t, untyped)

	// The control: the same input through the typed harness reports, which is what makes the
	// silence above a decline rather than a rule that cannot fire at all.
	typed := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", "Promise.reject(5)")
	rule_testing.ExpectFindings(t, typed, "rejectAnError")
}

// Cases upstream's corpus cannot write, because its parser folds them away or because they are
// TypeScript rather than JavaScript.
//
// The parenthesized receiver is the important one. Our parser keeps a KindParenthesizedExpression
// that upstream's deletes before its rule ever sees it, so an unwrap that upstream has no reason to
// write is load bearing here, in a place no imported fixture can reach.
func TestPreferPromiseRejectErrorsShapesUpstreamCannotWrite(t *testing.T) {
	t.Parallel()

	reports := []string{
		"(Promise).reject(5);",
		"((Promise)).reject(5);",
		"(Promise.reject)(5);",
		"Promise.reject((5));",
		"Promise[`reject`](5);",
		"new Promise((resolve, reject) => (reject)(5));",
		"new Promise((resolve, reject) => { if (true) { reject('x'); } });",
		"Promise.reject(5 as any as string);",
		"Promise.reject(<string>'x');",
	}
	for _, source := range reports {
		result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.tsx", source)
		if len(result.Diagnostics) != 1 {
			t.Errorf("%s: expected 1 finding, got %d", source, len(result.Diagnostics))
		}
	}

	clean := []string{
		"(Promise).reject(new Error());",
		"Promise[0](5);",
		"Promise[someName](5);",
		"class C { #reject: any; foo() { Promise.#reject(5); } }",
		// A destructured second parameter is not tracked, and reading its name would panic. The
		// array pattern is the shape upstream's corpus does not write at all.
		"new Promise(function(resolve, [first]) { first(5) });",
		"new Promise(function(resolve, ...rest) { rest[0](5) });",
		// An interface named Promise merges with the global rather than shadowing the value, so
		// this still reports; the class below replaces the value and does not.
		"class Promise2 { static reject(x: string) { return x; } } Promise2.reject('x');",
	}
	for _, source := range clean {
		result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", source)
		rule_testing.ExpectClean(t, result)
	}
}

// The receiver's NAME is load bearing, and upstream's corpus structurally cannot show it.
//
// Both arms test the receiver twice: it must be spelled `Promise`, and it must resolve to a
// declaration file. Removing the name test survived the whole imported corpus, because upstream's
// only wrong-receiver case is `Foo.reject(5)` and `Foo` is undeclared, so it fails the ambient test
// anyway and the name test is never consulted.
//
// Measured: `Reflect`, `JSON`, `Math` and `Object` all resolve entirely to declaration files, so
// without the name test each of these reports. Upstream is silent on all six, confirmed by driving
// the installed rule.
func TestPreferPromiseRejectErrorsOtherAmbientGlobalsAreNotPromise(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"Reflect.reject(5);",
		"JSON.reject(5);",
		"Math.reject(5);",
		"Object.reject(5);",
		"new Reflect((resolve: any, reject: any) => reject(5));",
		"new Object((resolve: any, reject: any) => reject(5));",
	} {
		result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", source)
		rule_testing.ExpectClean(t, result)
	}

	// The control: the same shape with the right receiver reports, so the silence above is the name
	// test declining rather than the rule being unable to fire on this shape at all.
	control := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", "Promise.reject(5);")
	rule_testing.ExpectFindings(t, control, "rejectAnError")
}

// The executor kind test is CRASH PROTECTION, not only fidelity.
//
// `Parameters()` nil-dereferences on a node with no parameter list, so a `new Promise(executor)`
// with a non-function argument takes the walk down. The walk recovers per FILE rather than per
// rule, so one such expression would cost every rule in the package its verdict on that file, and
// `new Promise(executor)` is ordinary code.
//
// Measured: an object literal, an array literal, a numeric literal and a bare identifier all panic
// on `Parameters()`, while an arrow function answers two. Upstream reports nothing on any of them,
// so the guard is right for its stated reason as well.
//
// No ExpectFindings fixture can see a panic, which is why removing the guard survived the whole
// imported corpus. These cases exist so a later revert crashes the suite instead of the tree.
func TestPreferPromiseRejectErrorsSurvivesANonFunctionExecutor(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"new Promise(executor);",
		"new Promise({ a: 1, b: 2 });",
		"new Promise([1, 2]);",
		"new Promise(5);",
		"declare const made: any; new Promise(made.executor);",
	} {
		result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", source)
		rule_testing.ExpectClean(t, result)
	}

	// The control: a real executor through the same harness still reports, so the silence above is
	// the guard declining rather than the listener never running.
	control := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts",
		"new Promise((resolve, reject) => reject(5));")
	rule_testing.ExpectFindings(t, control, "rejectAnError")
}

// Operator arms upstream's corpus does not exercise, with every verdict taken from driving the
// installed rule rather than from reading its switch.
//
// The corpus writes no sequence expression, no nullish coalescing and no conditional at all, so
// three arms of couldBeError sit unexercised by the imported cases. Each of these pairs an input
// where the deciding side could be an Error against one where it could not, so a mutation reading
// the wrong side of any of the three fails here.
func TestPreferPromiseRejectErrorsOperatorArmsUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	reports := []string{
		// A sequence takes its LAST expression's value, so a trailing literal reports even when the
		// leading operand is opaque.
		"Promise.reject((foo, 5));",
		"Promise.reject((5, 6));",
		// Nullish coalescing can yield either side, so it reports only when NEITHER could be an
		// Error.
		"Promise.reject(5 ?? 6);",
		// A conditional can yield either branch, same reasoning.
		"Promise.reject(5 ? 6 : 7);",
	}
	for _, source := range reports {
		result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", source)
		rule_testing.ExpectFindings(t, result, "rejectAnError")
	}

	clean := []string{
		"Promise.reject((5, foo));",
		"Promise.reject((foo, bar));",
		"Promise.reject(foo ?? 5);",
		"Promise.reject(5 ?? foo);",
		"Promise.reject(foo ? 5 : bar);",
	}
	for _, source := range clean {
		result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", source)
		rule_testing.ExpectClean(t, result)
	}
}

// TypeScript's type-only wrappers are transparent, so the value INSIDE decides.
//
// This is fidelity rather than an improvement. ESTree has no node for `as`, `satisfies`, a
// bracket assertion or `!`, so upstream's switch never mentions them and never sees one. The
// wrapper it does have settles the question: `Promise.reject((5))` reports and
// `Promise.reject((error))` does not, both measured against the installed rule, so the verdict
// follows the inner expression rather than the wrapper.
//
// Found by the dry run rather than by the corpus, which structurally cannot contain any of these.
// Before the unwrap covered them, `reject(error as Error)` in modules/kingdom/pentair/PentairApi.ts
// was reported: source asserting the value IS an Error, flagged for not being one.
func TestPreferPromiseRejectErrorsSeesThroughTypeOnlyWrappers(t *testing.T) {
	t.Parallel()

	// An opaque inner value stays clean through every wrapper.
	for _, source := range []string{
		"declare const error: unknown; Promise.reject(error as Error);",
		"declare const error: unknown; Promise.reject(<Error>error);",
		"declare const error: Error; Promise.reject(error satisfies Error);",
		"declare const error: Error | undefined; Promise.reject(error!);",
		"declare const error: unknown; Promise.reject((error as Error));",
		"declare const error: unknown; new Promise((resolve, reject) => reject(error as Error));",
		"declare const error: unknown; Promise.reject(error as unknown as Error);",
	} {
		result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", source)
		rule_testing.ExpectClean(t, result)
	}

	// A literal inner value still reports through every wrapper, which is what makes the unwrap a
	// descent rather than an exemption.
	for _, source := range []string{
		"Promise.reject(5 as unknown as Error);",
		"Promise.reject(<Error>(<unknown>5));",
		"Promise.reject('x' satisfies string);",
		"Promise.reject((5 as any));",
		"new Promise((resolve, reject) => reject(5 as any));",
	} {
		result := rule_testing.RunTyped(t, core.PreferPromiseRejectErrors, "input.ts", source)
		rule_testing.ExpectFindings(t, result, "rejectAnError")
	}
}
