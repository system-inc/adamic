package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// # How upstream's corpus was translated, and why it needed translating at all
//
// Upstream's corpus is mostly configuration. Every one of its 76 invalid cases declares the global
// it calls, through an explicit `globals` block, through `sourceType: "commonjs"` for `global`, or
// through `ecmaVersion: 2020` for `globalThis`, and ZERO of them fire with nothing declared.
// Measured by extracting the corpus and grouping on what supplies the name: 57 name a `globals`
// block, 15 rely on `commonjs`, 4 on the ecma version. Replaying `setTimeout("x = 1;")` through the
// installed eslint 10.8.1 reports once with `setTimeout` declared and not at all without it.
//
// cohere has no globals configuration. It asks the checker whether anything in SOURCE declares the
// name, and treats "nothing does" as the global. So a case's configuration translates into which
// names its PROGRAM declares, and the fixtures below run in one of two programs accordingly.
//
// # Nine sources carry opposite verdicts, and keying a fixture on source alone loses them
//
// The first version of this file was generated keyed on the source text, and nine sources appear in
// both of upstream's lists with the verdict decided only by `languageOptions`:
//
//	window.setTimeout('foo')       valid with nothing declared, invalid with `globals.browser`
//	global.setTimeout('foo')       valid with `globals.browser`, invalid with `sourceType: commonjs`
//	globalThis.setTimeout('foo')   valid at ecmaVersion 6, invalid at ecmaVersion 2020
//
// Collapsing those onto one row produced eleven failures that all read as rule defects and were
// none of them. Five further valid cases carry an EMPTY `globals: {}` block, which is upstream
// saying the name is deliberately not a global, and they belong in the bare program for the same
// reason. The split below is on what each case needs declared, not on what it says.

// impliedEvalAmbientGlobals is the program shape for a case whose root name upstream declared.
//
// It stands in for `lib.dom.d.ts`, which the fixture harness does not load: `rule_testing` pins
// `lib: ["ES2022"]` with no DOM, while the ahra tree's own tsconfig carries
// `lib: ["dom", "dom.iterable", "esnext"]` and therefore has all of these. Supplying them through an
// ambient declaration file is the same mechanism a real lib file uses rather than an arrangement
// made for the test.
//
// Probed before the rule was written, in internal/no_implied_eval_core_probe: with this file in the
// program, `setTimeout` resolves to one declaration in a declaration file, and a parameter named
// `setTimeout` resolves to a non-declaration file, so the two are separable. Without it,
// `setTimeout` resolves to nothing at all.
//
// All seven root names upstream's firing cases use are declared here, because upstream declares
// every one of them: `globals.browser` supplies `window`, `self`, `setTimeout`, `setInterval` and
// `execScript`, `sourceType: "commonjs"` supplies `global`, and `ecmaVersion: 2020` supplies
// `globalThis`. Counted over the 74 firing cases, the roots are window 22, self 21, global 17,
// setTimeout 8, globalThis 4, setInterval 1, execScript 1.
//
// `execScript` and `global` matter most here, because no TypeScript library declares either. A
// program that does not declare them cannot report them, and neither can upstream: measured against
// eslint 10.8.1, `execScript("x")` with nothing declared reports zero and reports once as soon as
// `execScript` is added to the globals block. That is a real limit of this rule on an ordinary
// TypeScript tree and it is upstream's limit too, not a gap in the port.
const impliedEvalAmbientGlobals = `declare function setTimeout(handler: unknown, timeout?: number, ...rest: unknown[]): number;
declare function setInterval(handler: unknown, timeout?: number, ...rest: unknown[]): number;
declare function execScript(source: unknown, ...rest: unknown[]): void;
interface ImpliedEvalGlobalObject {
	setTimeout(handler: unknown, timeout?: number, ...rest: unknown[]): number;
	setInterval(handler: unknown, timeout?: number, ...rest: unknown[]): number;
	execScript(source: unknown, ...rest: unknown[]): void;
	window: ImpliedEvalGlobalObject;
	self: ImpliedEvalGlobalObject;
	global: ImpliedEvalGlobalObject;
	globalThis: ImpliedEvalGlobalObject;
}
declare var window: ImpliedEvalGlobalObject;
declare var self: ImpliedEvalGlobalObject;
declare var global: ImpliedEvalGlobalObject;
`

// runImpliedEvalWithGlobals runs a case in a program that declares the browser globals.
func runImpliedEvalWithGlobals(t *testing.T, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, NoImpliedEval, map[string]string{
		"globals.d.ts": impliedEvalAmbientGlobals,
		"subject.ts":   source,
	}, "subject.ts")
}

// runImpliedEvalWithoutGlobals runs a case in a program that declares none of them, which is what
// upstream expresses as an absent or empty `globals` block.
func runImpliedEvalWithoutGlobals(t *testing.T, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, NoImpliedEval, map[string]string{
		"subject.ts": source,
	}, "subject.ts")
}

// TestNoImpliedEvalFires carries 74 of upstream's 76 invalid cases. The other two are in
// TestNoImpliedEvalDivergesFromUpstreamOnStaticValues, with the measurement that put them there.
//
// The sources are upstream's, taken by evaluating its test module and rendering each string through
// a Go quoter, so no case here was retyped and none could have been silently cooked on the way in.
func TestNoImpliedEvalFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		ids    []string
	}{
		{source: "setTimeout(\"x = 1;\");", ids: []string{"impliedEval"}},
		{source: "setTimeout(\"x = 1;\", 100);", ids: []string{"impliedEval"}},
		{source: "setInterval(\"x = 1;\");", ids: []string{"impliedEval"}},
		{source: "execScript(\"x = 1;\");", ids: []string{"execScript"}},
		{source: "window.setTimeout('foo')", ids: []string{"impliedEval"}},
		{source: "window.setInterval('foo')", ids: []string{"impliedEval"}},
		{source: "window.execScript('foo')", ids: []string{"execScript"}},
		{source: "window['setTimeout']('foo')", ids: []string{"impliedEval"}},
		{source: "window['setInterval']('foo')", ids: []string{"impliedEval"}},
		{source: "window[`setInterval`]('foo')", ids: []string{"impliedEval"}},
		{source: "window['execScript']('foo')", ids: []string{"execScript"}},
		{source: "window[`execScript`]('foo')", ids: []string{"execScript"}},
		{source: "window.window['setInterval']('foo')", ids: []string{"impliedEval"}},
		{source: "window.window['execScript']('foo')", ids: []string{"execScript"}},
		{source: "global.setTimeout('foo')", ids: []string{"impliedEval"}},
		{source: "global.setInterval('foo')", ids: []string{"impliedEval"}},
		{source: "global.execScript('foo')", ids: []string{"execScript"}},
		{source: "global['setTimeout']('foo')", ids: []string{"impliedEval"}},
		{source: "global['setInterval']('foo')", ids: []string{"impliedEval"}},
		{source: "global[`setInterval`]('foo')", ids: []string{"impliedEval"}},
		{source: "global['execScript']('foo')", ids: []string{"execScript"}},
		{source: "global[`execScript`]('foo')", ids: []string{"execScript"}},
		{source: "global.global['setInterval']('foo')", ids: []string{"impliedEval"}},
		{source: "global.global['execScript']('foo')", ids: []string{"execScript"}},
		{source: "globalThis.setTimeout('foo')", ids: []string{"impliedEval"}},
		{source: "globalThis.setInterval('foo')", ids: []string{"impliedEval"}},
		{source: "globalThis.execScript('foo')", ids: []string{"execScript"}},
		{source: "setTimeout(`foo${bar}`)", ids: []string{"impliedEval"}},
		{source: "window.setTimeout(`foo${bar}`)", ids: []string{"impliedEval"}},
		{source: "window.window.setTimeout(`foo${bar}`)", ids: []string{"impliedEval"}},
		{source: "global.global.setTimeout(`foo${bar}`)", ids: []string{"impliedEval"}},
		{source: "setTimeout('foo' + bar)", ids: []string{"impliedEval"}},
		{source: "setTimeout(foo + 'bar')", ids: []string{"impliedEval"}},
		{source: "setTimeout(`foo` + bar)", ids: []string{"impliedEval"}},
		{source: "setTimeout(1 + ';' + 1)", ids: []string{"impliedEval"}},
		{source: "window.setTimeout('foo' + bar)", ids: []string{"impliedEval"}},
		{source: "window.setTimeout(foo + 'bar')", ids: []string{"impliedEval"}},
		{source: "window.setTimeout(`foo` + bar)", ids: []string{"impliedEval"}},
		{source: "window.setTimeout(1 + ';' + 1)", ids: []string{"impliedEval"}},
		{source: "window.window.setTimeout(1 + ';' + 1)", ids: []string{"impliedEval"}},
		{source: "global.setTimeout('foo' + bar)", ids: []string{"impliedEval"}},
		{source: "global.setTimeout(foo + 'bar')", ids: []string{"impliedEval"}},
		{source: "global.setTimeout(`foo` + bar)", ids: []string{"impliedEval"}},
		{source: "global.setTimeout(1 + ';' + 1)", ids: []string{"impliedEval"}},
		{source: "global.global.setTimeout(1 + ';' + 1)", ids: []string{"impliedEval"}},
		{source: "globalThis.setTimeout('foo' + bar)", ids: []string{"impliedEval"}},
		{source: "setTimeout('foo' + (function() {\n   setTimeout(helper);\n   execScript('str');\n   return 'bar';\n})())", ids: []string{"impliedEval", "execScript"}},
		{source: "window.setTimeout('foo' + (function() {\n   setTimeout(helper);\n   window.execScript('str');\n   return 'bar';\n})())", ids: []string{"impliedEval", "execScript"}},
		{source: "global.setTimeout('foo' + (function() {\n   setTimeout(helper);\n   global.execScript('str');\n   return 'bar';\n})())", ids: []string{"impliedEval", "execScript"}},
		{source: "window?.setTimeout('code', 0)", ids: []string{"impliedEval"}},
		{source: "(window?.setTimeout)('code', 0)", ids: []string{"impliedEval"}},
		{source: "window?.execScript('code')", ids: []string{"execScript"}},
		{source: "(window?.execScript)('code')", ids: []string{"execScript"}},
		{source: "self.setTimeout('foo')", ids: []string{"impliedEval"}},
		{source: "self.setInterval('foo')", ids: []string{"impliedEval"}},
		{source: "self.execScript('foo')", ids: []string{"execScript"}},
		{source: "self['setTimeout']('foo')", ids: []string{"impliedEval"}},
		{source: "self['setInterval']('foo')", ids: []string{"impliedEval"}},
		{source: "self[`setInterval`]('foo')", ids: []string{"impliedEval"}},
		{source: "self['execScript']('foo')", ids: []string{"execScript"}},
		{source: "self[`execScript`]('foo')", ids: []string{"execScript"}},
		{source: "self.self['setInterval']('foo')", ids: []string{"impliedEval"}},
		{source: "self.self['execScript']('foo')", ids: []string{"execScript"}},
		{source: "self.setTimeout(`foo${bar}`)", ids: []string{"impliedEval"}},
		{source: "self.self.setTimeout(`foo${bar}`)", ids: []string{"impliedEval"}},
		{source: "self.setTimeout('foo' + bar)", ids: []string{"impliedEval"}},
		{source: "self.setTimeout(foo + 'bar')", ids: []string{"impliedEval"}},
		{source: "self.setTimeout(`foo` + bar)", ids: []string{"impliedEval"}},
		{source: "self.setTimeout(1 + ';' + 1)", ids: []string{"impliedEval"}},
		{source: "self.self.setTimeout(1 + ';' + 1)", ids: []string{"impliedEval"}},
		{source: "self?.setTimeout('code', 0)", ids: []string{"impliedEval"}},
		{source: "(self?.setTimeout)('code', 0)", ids: []string{"impliedEval"}},
		{source: "self?.execScript('code')", ids: []string{"execScript"}},
		{source: "(self?.execScript)('code')", ids: []string{"execScript"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := runImpliedEvalWithGlobals(t, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

// TestNoImpliedEvalStaysSilentWithGlobalsDeclared carries the 58 valid cases whose root name
// upstream declared as a global. These are silent for a reason other than the global test: the
// argument is a function, the call is not a call at all, or the property name does not match.
func TestNoImpliedEvalStaysSilentWithGlobalsDeclared(t *testing.T) {
	t.Parallel()

	cases := []struct{ source string }{
		{source: "setTimeout();"},
		{source: "window.setTimeout;"},
		{source: "window.setTimeout = foo;"},
		{source: "window['setTimeout'];"},
		{source: "window['setTimeout'] = foo;"},
		{source: "global.setTimeout;"},
		{source: "global.setTimeout = foo;"},
		{source: "global['setTimeout'];"},
		{source: "global['setTimeout'] = foo;"},
		{source: "globalThis['setTimeout'] = foo;"},
		{source: "window[`SetTimeOut`]('foo', 100);"},
		{source: "global[`SetTimeOut`]('foo', 100);"},
		{source: "global[`setTimeout${foo}`]('foo', 100);"},
		{source: "globalThis[`setTimeout${foo}`]('foo', 100);"},
		{source: "setTimeout(function() { x = 1; }, 100);"},
		{source: "setInterval(function() { x = 1; }, 100);"},
		{source: "execScript(function() { x = 1; }, 100);"},
		{source: "window.setTimeout(function() { x = 1; }, 100);"},
		{source: "window.setInterval(function() { x = 1; }, 100);"},
		{source: "window.execScript(function() { x = 1; }, 100);"},
		{source: "window.setTimeout(foo, 100);"},
		{source: "window.setInterval(foo, 100);"},
		{source: "window.execScript(foo, 100);"},
		{source: "global.setTimeout(function() { x = 1; }, 100);"},
		{source: "global.setInterval(function() { x = 1; }, 100);"},
		{source: "global.execScript(function() { x = 1; }, 100);"},
		{source: "global.setTimeout(foo, 100);"},
		{source: "global.setInterval(foo, 100);"},
		{source: "global.execScript(foo, 100);"},
		{source: "globalThis.setTimeout(foo, 100);"},
		{source: "setTimeout(foo, 10)"},
		{source: "setInterval(1, 10)"},
		{source: "execScript(2)"},
		{source: "setTimeout(function() {}, 10)"},
		{source: "setInterval(foo, 10)"},
		{source: "setInterval(function() {}, 10)"},
		{source: "execScript(foo)"},
		{source: "execScript(function() {})"},
		{source: "setTimeout(foo + bar, 10)"},
		{source: "setTimeout(foobar, 'buzz')"},
		{source: "setTimeout(foobar, foo + 'bar')"},
		{source: "setTimeout(function() { return 'foobar'; }, 10)"},
		{source: "var window; window.setTimeout('foo', 100);"},
		{source: "var global; global.setTimeout('foo', 100);"},
		{source: "\n\t\t\tfunction execScript(string) {\n\t\t\t\tconsole.log(\"This is not your grandparent's execScript().\");\n\t\t\t}\n\n\t\t\texecScript('wibble');\n\t\t\t"},
		{source: "\n\t\t\tfunction setTimeout(string) {\n\t\t\t\tconsole.log(\"This is not your grandparent's setTimeout().\");\n\t\t\t}\n\n\t\t\tsetTimeout('wibble');\n\t\t\t"},
		{source: "\n\t\t\tfunction setInterval(string) {\n\t\t\t\tconsole.log(\"This is not your grandparent's setInterval().\");\n\t\t\t}\n\n\t\t\tsetInterval('wibble');\n\t\t\t"},
		{source: "self.setTimeout;"},
		{source: "self.setTimeout = foo;"},
		{source: "self['setTimeout'];"},
		{source: "self['setTimeout'] = foo;"},
		{source: "self[`SetTimeOut`]('foo', 100);"},
		{source: "self[`setTimeout${foo}`]('foo', 100);"},
		{source: "self.setTimeout(function() { x = 1; }, 100);"},
		{source: "self.setInterval(function() { x = 1; }, 100);"},
		{source: "self.execScript(function() { x = 1; }, 100);"},
		{source: "self.setTimeout(foo, 100);"},
		{source: "var self; self.setTimeout('foo', 100);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := runImpliedEvalWithGlobals(t, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoImpliedEvalStaysSilentWithoutGlobalsDeclared carries the 41 valid cases that need the name
// NOT to be a global: a local shadow, an unrelated receiver, or upstream's empty `globals: {}`.
//
// This is the half of the corpus the rule reads the checker for. Twelve of these are shadows
// specifically, in every shape that can produce one: a function declaration, a block-scoped const, a
// parameter, and a nested function whose shadow sits in an enclosing scope.
func TestNoImpliedEvalStaysSilentWithoutGlobalsDeclared(t *testing.T) {
	t.Parallel()

	cases := []struct{ source string }{
		{source: "setTimeout;"},
		{source: "setTimeout = foo;"},
		{source: "window.setTimeout('foo')"},
		{source: "window.setInterval('foo')"},
		{source: "window['setTimeout']('foo')"},
		{source: "window['setInterval']('foo')"},
		{source: "window.setTimeout('foo')"},
		{source: "window.setInterval('foo')"},
		{source: "window['setTimeout']('foo')"},
		{source: "window['setInterval']('foo')"},
		{source: "global.setTimeout('foo')"},
		{source: "global.setInterval('foo')"},
		{source: "global['setTimeout']('foo')"},
		{source: "global['setInterval']('foo')"},
		{source: "global[`setTimeout${foo}`]('foo', 100);"},
		{source: "foo.setTimeout('hi')"},
		{source: "foo.setInterval('hi')"},
		{source: "foo.execScript('hi')"},
		{source: "setTimeoutFooBar('Foo Bar')"},
		{source: "foo.window.setTimeout('foo', 100);"},
		{source: "foo.global.setTimeout('foo', 100);"},
		{source: "function foo(window) { window.setTimeout('foo', 100); }"},
		{source: "function foo(global) { global.setTimeout('foo', 100); }"},
		{source: "foo('', window.setTimeout);"},
		{source: "foo('', global.setTimeout);"},
		{source: "\n\t\t\tfunction outer() {\n\t\t\t\tfunction setTimeout(string) {\n\t\t\t\t\tconsole.log(\"Shadowed setTimeout\");\n\t\t\t\t}\n\t\t\t\tsetTimeout('code');\n\t\t\t}\n\t\t\t"},
		{source: "\n\t\t\tfunction outer() {\n\t\t\t\tfunction setInterval(string) {\n\t\t\t\t\tconsole.log(\"Shadowed setInterval\");\n\t\t\t\t}\n\t\t\t\tsetInterval('code');\n\t\t\t}\n\t\t\t"},
		{source: "\n\t\t\tfunction outer() {\n\t\t\t\tfunction execScript(string) {\n\t\t\t\t\tconsole.log(\"Shadowed execScript\");\n\t\t\t\t}\n\t\t\t\texecScript('code');\n\t\t\t}\n\t\t\t"},
		{source: "\n\t\t\t{\n\t\t\t\tconst setTimeout = function(string) {\n\t\t\t\t\tconsole.log(\"Block-scoped setTimeout\");\n\t\t\t\t};\n\t\t\t\tsetTimeout('code');\n\t\t\t}\n\t\t\t"},
		{source: "\n\t\t\t{\n\t\t\t\tconst setInterval = function(string) {\n\t\t\t\t\tconsole.log(\"Block-scoped setInterval\");\n\t\t\t\t};\n\t\t\t\tsetInterval('code');\n\t\t\t}\n\t\t\t"},
		{source: "setTimeout('code');"},
		{source: "setInterval('code');"},
		{source: "execScript('code');"},
		{source: "window.setTimeout('code');"},
		{source: "foo.self.setTimeout('foo', 100);"},
		{source: "function foo(self) { self.setTimeout('foo', 100); }"},
		{source: "foo('', self.setTimeout);"},
		{source: "\n\t\t\tfunction outer() {\n\t\t\t\tfunction self() {\n\t\t\t\t\tconsole.log(\"Shadowed self\");\n\t\t\t\t}\n\t\t\t\tself.setTimeout('code');\n\t\t\t}"},
		{source: "self.setTimeout('code');"},
	}

	for _, testCase := range cases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := runImpliedEvalWithoutGlobals(t, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoImpliedEvalDivergesFromUpstreamOnStaticValues pins the two corpus cases this port does not
// reproduce, and records why as a measurement rather than as an opinion.
//
// Upstream reports a string first argument by EITHER of two tests. The syntactic one is ported. The
// other is `getStaticValue` from eslint-utils, a constant expression evaluator of roughly 786 lines
// that follows bindings, folds conditionals, and evaluates about a hundred whitelisted builtin calls
// whose whitelist is keyed on JavaScript function IDENTITY rather than on names.
//
// The cost of omitting it was measured rather than estimated. Upstream's own rule was copied with
// the `getStaticValue` call replaced by `null`, and both versions were replayed over all 175 corpus
// cases through the installed eslint 10.8.1. Exactly two verdicts moved, both invalid cases, and no
// valid case moved at all. A control confirmed the neutered rule still caught a plain string
// literal, so the two are a real difference rather than a broken instrument.
//
// Both divergences are toward SILENCE, the safe direction for a rule proposing no repair: the cost
// is a missed finding rather than a rewritten file. Counted with the TypeScript parser over 3,469
// files in the ahra tree, neither shape occurs: 206 calls to the three eval-like functions and not
// one passes a string, a template, a concatenation, or a const-bound string.
//
// If it ever needs closing, the smallest honest version follows a `const` binding whose initializer
// is already a string by the syntactic test. That covers the first case and not the second, and it
// is a much smaller thing than the evaluator.
func TestNoImpliedEvalDivergesFromUpstreamOnStaticValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		// upstreamIds is what eslint reports for this case. It is recorded rather than asserted,
		// because the assertion below is that this port is SILENT. A later change that closes the
		// gap should turn these into the expectation and fail here first.
		upstreamIds []string
	}{
		{source: "const s = 'x=1'; setTimeout(s, 100);", upstreamIds: []string{"impliedEval"}},
		{source: "setTimeout(String('x=1'), 100);", upstreamIds: []string{"impliedEval"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := runImpliedEvalWithGlobals(t, testCase.source)
			rule_testing.ExpectClean(t, result)
			if len(testCase.upstreamIds) == 0 {
				t.Errorf("case records no upstream verdict, so the divergence is undocumented")
			}
		})
	}
}

// TestNoImpliedEvalReportsGlobalThisUpstreamGatesByEcmaVersion records the port's one behavioural
// divergence outside the static-value gap, and it runs the other way: this rule reports two cases
// upstream lists as valid.
//
// Both are `globalThis` at an ecma version predating it. `globalThis` entered the language in
// ES2020, and upstream's globals table is version-indexed, so at ecmaVersion 6 and 2017 the name is
// not a global and the call is not the global's. Measured against eslint 10.8.1, the same source is
// silent at 6 and at 2017 and reports at 2020, with nothing else changed:
//
//	globalThis.setTimeout('foo')        ecma 6      0 findings
//	globalThis.setTimeout('foo')        ecma 2020   1 finding
//	globalThis['setInterval']('foo')    ecma 2017   0 findings
//
// cohere has no equivalent gate to reproduce. The version a file targets is a compiler option
// rather than a property of the name, and TypeScript's own libraries declare `globalThis`
// unconditionally, so a program built from this tree's tsconfig has it at every target. There is no
// state in which the checker would answer that `globalThis` is not declared.
//
// Reproducing upstream here would mean reading the program's target and suppressing the name below
// ES2020, which would be inventing a gate rather than porting one, and it would misfire on any file
// whose target is low while the runtime is modern. The divergence is left in place and pinned here.
//
// It costs nothing on this tree: `globalThis` appears in no eval-like call anywhere in ahra, and
// both shapes are ones upstream itself reports at any current ecma version, which is every version
// this codebase compiles under.
func TestNoImpliedEvalReportsGlobalThisUpstreamGatesByEcmaVersion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		// upstreamEcmaVersion is the setting that makes upstream call this valid. At 2020 and
		// above upstream reports it too, which is what makes this a version gate rather than a
		// disagreement about the rule.
		upstreamEcmaVersion string
		ids                 []string
	}{
		{
			source:              "globalThis.setTimeout('foo')",
			upstreamEcmaVersion: "6",
			ids:                 []string{"impliedEval"},
		},
		{
			source:              "globalThis['setInterval']('foo')",
			upstreamEcmaVersion: "2017",
			ids:                 []string{"impliedEval"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := runImpliedEvalWithGlobals(t, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.ids...)
			if testCase.upstreamEcmaVersion == "" {
				t.Errorf("case records no upstream version gate, so the divergence is undocumented")
			}
		})
	}
}

// TestNoImpliedEvalDeclinesAGlobalReceiverOutsideTheCandidateSet covers the global-object name
// check, which upstream's corpus cannot exercise.
//
// Upstream walks only four receiver names, `global`, `window`, `globalThis` and `self`, because
// those are the ones that ARE the global object. Every other receiver names an ordinary object
// whose `setTimeout` is its own method, and calling it with a string runs nothing.
//
// The corpus tests this only with receivers that are not globals at all, such as `foo.setTimeout`,
// so a rule that dropped the name check entirely and relied on the global test alone passes all 175
// imported cases. Found by mutation: replacing the candidate-name test with `if false` survived the
// whole corpus.
//
// The distinguishing input is a receiver that is a REAL declared global and is not one of the four.
// `document` is the natural one and it is in this codebase's own globals list. Measured against
// eslint 10.8.1 with each name declared: `document.setTimeout('x')`, `navigator.setTimeout('x')` and
// `top.setTimeout('x')` all report zero, while `window.setTimeout('x')` reports one, with nothing
// else changed between the runs.
func TestNoImpliedEvalDeclinesAGlobalReceiverOutsideTheCandidateSet(t *testing.T) {
	t.Parallel()

	// A program declaring three globals that are NOT global-object candidates, each carrying a
	// method with an eval-like name. Without these declarations the receiver would be declined for
	// being undeclared and the case would pass for the wrong reason.
	const nonCandidateGlobals = `interface ImpliedEvalOrdinaryObject {
	setTimeout(handler: unknown, timeout?: number): number;
	setInterval(handler: unknown, timeout?: number): number;
	execScript(source: unknown): void;
}
declare var document: ImpliedEvalOrdinaryObject;
declare var navigator: ImpliedEvalOrdinaryObject;
declare var top: ImpliedEvalOrdinaryObject;
`

	silent := []string{
		"document.setTimeout('x = 1;')",
		"document['setTimeout']('x = 1;')",
		"navigator.setInterval('x = 1;')",
		"top.execScript('x = 1;')",
	}
	for _, source := range silent {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFiles(t, NoImpliedEval, map[string]string{
				"globals.d.ts": nonCandidateGlobals,
				"subject.ts":   source,
			}, "subject.ts")
			rule_testing.ExpectClean(t, result)
		})
	}

	// The control for the four cases above: the same shape with a candidate receiver reports, so
	// their silence is the name check rather than the program being wrong or the rule being inert.
	t.Run("control: a candidate receiver in the same program reports", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTypedFiles(t, NoImpliedEval, map[string]string{
			"globals.d.ts": nonCandidateGlobals + impliedEvalAmbientGlobals,
			"subject.ts":   "window.setTimeout('x = 1;')",
		}, "subject.ts")
		rule_testing.ExpectFindings(t, result, "impliedEval")
	})
}

// TestNoImpliedEvalDeclinesAVariableBracketKey covers the guard that refuses to read a bare
// identifier as a bracketed property name, which upstream's corpus never writes.
//
// `window[setTimeout]('x')` reads whichever property the VARIABLE `setTimeout` names, which is not
// knowable before it runs, so it is not a call to the global's `setTimeout` at all. Upstream
// declines it through `getStaticPropertyName`, which returns null for a computed key that is a bare
// identifier. Measured against eslint 10.8.1 with both `window` and `setTimeout` declared:
// `window[setTimeout]('x')` and `window[key]('x')` report zero while `window['setTimeout']('x')`
// reports one.
//
// The shelf's `property.Name` declines a bare identifier only under `Computed`, which is the shape
// of an object-literal or class key. An element access hands its key over directly rather than
// wrapped, so it reaches the `KindIdentifier` arm and is ACCEPTED. That is correct for the shelf's
// other callers and wrong here, so the rule guards it at the call site.
//
// Found by mutation: removing the guard survived all 175 imported cases, because the corpus writes
// no bracketed variable key anywhere.
func TestNoImpliedEvalDeclinesAVariableBracketKey(t *testing.T) {
	t.Parallel()

	silent := []string{
		"window[setTimeout]('x = 1;')",
		"window[setInterval]('x = 1;')",
		"self[setTimeout]('x = 1;')",
	}
	for _, source := range silent {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			result := runImpliedEvalWithGlobals(t, source)
			rule_testing.ExpectClean(t, result)
		})
	}

	// The control: the same receiver and the same property, spelled as a string, reports. Without
	// this the four cases above would pass equally well against a rule that had gone inert.
	t.Run("control: the same property as a string key reports", func(t *testing.T) {
		t.Parallel()
		result := runImpliedEvalWithGlobals(t, "window['setTimeout']('x = 1;')")
		rule_testing.ExpectFindings(t, result, "impliedEval")
	})
}

// TestNoImpliedEvalStopsTheReceiverWalkAtANonCandidateLink covers the middle of the receiver chain.
//
// Upstream advances through a receiver chain only while each link names the global object again, so
// `window.window.setTimeout('s')` reports and `window.document.setTimeout('s')` does not. The second
// reads a `setTimeout` belonging to `document`, which is an ordinary method.
//
// The corpus tests the chain only with a repeated candidate, `window.window` and `global.global` and
// `self.self`, so both halves of the link test are true together in every imported case and a rule
// that weakened the test from OR to AND passes all of them. Found by mutation.
//
// The distinguishing input needs a middle link that IS a settled property name but is NOT one of
// the four candidates. Measured against eslint 10.8.1 with `window`, `document`, `foo` and
// `setTimeout` declared: `window.document.setTimeout('x')` and `window.foo.setTimeout('x')` report
// zero, `window.window.setTimeout('x')` reports one.
func TestNoImpliedEvalStopsTheReceiverWalkAtANonCandidateLink(t *testing.T) {
	t.Parallel()

	// The candidate object carries a `document` property so the middle link resolves to a real
	// settled name rather than to nothing, which is what makes this test the link check and not the
	// resolution check.
	const chainGlobals = `interface ImpliedEvalChainObject {
	setTimeout(handler: unknown, timeout?: number): number;
	setInterval(handler: unknown, timeout?: number): number;
	window: ImpliedEvalChainObject;
	self: ImpliedEvalChainObject;
	document: ImpliedEvalChainObject;
	frames: ImpliedEvalChainObject;
}
declare var window: ImpliedEvalChainObject;
declare var self: ImpliedEvalChainObject;
`

	run := func(t *testing.T, source string) rule_testing.Result {
		t.Helper()
		return rule_testing.RunTypedFiles(t, NoImpliedEval, map[string]string{
			"globals.d.ts": chainGlobals,
			"subject.ts":   source,
		}, "subject.ts")
	}

	silent := []string{
		"window.document.setTimeout('x = 1;')",
		"window.frames.setTimeout('x = 1;')",
		"self.document.setInterval('x = 1;')",
		"window.document['setTimeout']('x = 1;')",
	}
	for _, source := range silent {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, run(t, source))
		})
	}

	// The control: the same two-link shape with a candidate in the middle reports, so the silence
	// above is the link check rather than the walk having stopped working altogether.
	for _, source := range []string{
		"window.window.setTimeout('x = 1;')",
		"self.self.setInterval('x = 1;')",
	} {
		t.Run("control: "+source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, run(t, source), "impliedEval")
		})
	}
}

// TestNoImpliedEvalReadsThroughParenthesesOnTheArgument covers the argument unwrap, which upstream's
// corpus structurally cannot test.
//
// Our parser keeps `KindParenthesizedExpression` as a real node; the parser upstream recurses over
// deletes it before any rule runs. So `setTimeout(('x'))` arrives here as a parenthesized expression
// wrapping a string, and upstream never sees the wrapper at all. That means no imported fixture can
// flag a missing unwrap: upstream had no reason to write one of these cases, because for its parser
// they are the same tree as the unparenthesized form.
//
// Found by mutation: disabling the unwrap survived all 175 imported cases.
//
// Measured against eslint 10.8.1 with `setTimeout` and `window` declared, every one of these reports
// exactly once, which is what the rule has to reproduce:
//
//	setTimeout(('x'))          setTimeout(('a' + bar))    setTimeout(bar + ('a'))
//	setTimeout((('x')))        setTimeout(('a') + bar)    window.setTimeout(('x'))
//
// The unwrap is a loop rather than a single step because the wrapping nests, and the nil check is
// written out rather than delegated to `ast.SkipParentheses`, which dereferences its argument.
func TestNoImpliedEvalReadsThroughParenthesesOnTheArgument(t *testing.T) {
	t.Parallel()

	firing := []string{
		"setTimeout(('x = 1;'))",
		"setTimeout((('x = 1;')))",
		"setTimeout(('foo' + bar))",
		"setTimeout(('foo') + bar)",
		"setTimeout(bar + ('foo'))",
		"window.setTimeout(('x = 1;'))",
	}
	for _, source := range firing {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			result := runImpliedEvalWithGlobals(t, source)
			rule_testing.ExpectFindings(t, result, "impliedEval")
		})
	}

	// The control: a parenthesized argument that is NOT a string stays silent, so the cases above
	// are the unwrap finding a string rather than the rule reporting any parenthesized argument.
	for _, source := range []string{
		"setTimeout((foo))",
		"setTimeout((function() {}))",
	} {
		t.Run("control: "+source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runImpliedEvalWithGlobals(t, source))
		})
	}
}

// TestNoImpliedEvalReadsThroughParenthesesOnTheReceiver is the receiver-side counterpart of the
// argument unwrap above, and it exists for the same parser difference.
//
// `(window).setTimeout('s')` is an ordinary member access whose object our parser wraps in a
// `KindParenthesizedExpression` and upstream's parser does not. Without the unwrap the receiver
// reads as neither an identifier nor a member access, the walk gives up, and every one of these
// goes silent.
//
// Found by mutation, and confirmed reachable before a fixture was written rather than after.
// Measured against eslint 10.8.1 with `window` and `setTimeout` declared, all four report once:
//
//	(window).setTimeout('x')        (window).window.setTimeout('x')
//	((window)).setTimeout('x')      (window)['setTimeout']('x')
//
// Only the chain ROOT can carry parentheses. A middle link cannot: `window.(self).setTimeout` is not
// syntax, so the unwrap inside the walk loop matters on the first turn and is inert afterwards.
func TestNoImpliedEvalReadsThroughParenthesesOnTheReceiver(t *testing.T) {
	t.Parallel()

	firing := []string{
		"(window).setTimeout('x = 1;')",
		"((window)).setTimeout('x = 1;')",
		"(window).window.setTimeout('x = 1;')",
		"(window)['setTimeout']('x = 1;')",
		"(self).setInterval('x = 1;')",
	}
	for _, source := range firing {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			result := runImpliedEvalWithGlobals(t, source)
			rule_testing.ExpectFindings(t, result, "impliedEval")
		})
	}

	// The control: a parenthesized receiver that is not a global-object candidate stays silent, so
	// the cases above are the unwrap reaching a candidate rather than parentheses being waved
	// through.
	t.Run("control: a parenthesized non-candidate receiver stays silent", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, runImpliedEvalWithGlobals(t, "(foo).setTimeout('x = 1;')"))
	})
}

// TestNoImpliedEvalRequiresTheTypedHarness fails loudly if somebody reverts the checker
// declaration or removes the nil guard, either of which would buy a vacuous green.
//
// Measured on this substrate rather than assumed: `GetSymbolAtLocation` on a nil checker returns
// nil rather than panicking, so an untyped run goes completely SILENT rather than crashing. Silence
// is the more dangerous failure, because every clean fixture then passes for the wrong reason and
// nothing announces itself.
//
// The guard sits in `Run` rather than in the listener, so it declines the file once instead of once
// per node. `NeedsTypeChecker` governs the registration path only: a `rule.Context` built by hand
// still arrives with a nil checker, and `TestNoRegisteredRuleCrashesOnAbsentOptionalNodes` builds
// exactly that.
func TestNoImpliedEvalRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoImpliedEval.NeedsTypeChecker {
		t.Fatalf("this rule resolves an identifier to its declaration and must declare NeedsTypeChecker")
	}

	// The untyped harness hands the rule a nil checker. It must decline rather than guess.
	if result := rule_testing.Run(t, NoImpliedEval, "input.ts", `setTimeout("x = 1;");`); len(result.Diagnostics) != 0 {
		t.Errorf("the untyped harness must produce no findings, got %d", len(result.Diagnostics))
	}

	// The control for that zero: the same source through the typed harness, in a program that
	// declares the global, reaches one finding. Without this the assertion above would pass equally
	// well against a rule that had stopped working entirely.
	if result := runImpliedEvalWithGlobals(t, `setTimeout("x = 1;");`); len(result.Diagnostics) != 1 {
		t.Errorf("the typed harness must reach one finding, got %d", len(result.Diagnostics))
	}
}

// TestNoImpliedEvalDoesNotCrashOnACallWithNoArguments pins the guard that reads the argument list.
//
// `setTimeout()` is upstream's very first valid case and it is also the shape that would index an
// empty list. No `ExpectFindings` fixture can see a panic, so the sweep reports a survivor
// identically whether the guard matters or not; this asks what the line PREVENTS rather than what it
// decides.
//
// The other shapes here are calls whose callee is a bare keyword rather than a name, which reach the
// listener as ordinary call expressions and have crashed rules in this tree before.
func TestNoImpliedEvalDoesNotCrashOnACallWithNoArguments(t *testing.T) {
	t.Parallel()

	sources := []string{
		"setTimeout();",
		"window.setTimeout();",
		"window['setTimeout']();",
		"class A { constructor() { super(); } }",
		"async function f() { await import('./other'); }",
		"const x = (function () {})();",
	}
	for _, source := range sources {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			// Not asserting the verdict, only that the rule survives the shape. A panic here takes
			// every rule's verdict on the file, not just this one's.
			runImpliedEvalWithGlobals(t, source)
		})
	}
}

// TestNoImpliedEvalReportsTheWholeCallAndSaysWhy asserts where the finding points and what it says.
//
// `ExpectFindings` compares message ids and counts and nothing else, so a rule pointing at the wrong
// node or rendering the wrong sentence passes a complete fixture pair while being wrong. Upstream
// reports the CallExpression rather than the callee or the argument, and this asserts that by
// slicing the source with the finding's own range.
func TestNoImpliedEvalReportsTheWholeCallAndSaysWhy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source      string
		wantSpan    string
		wantId      string
		wantPhrases []string
	}{
		{
			source:      `setTimeout("x = 1;", 100);`,
			wantSpan:    `setTimeout("x = 1;", 100)`,
			wantId:      "impliedEval",
			wantPhrases: []string{"string", "function"},
		},
		{
			source:      `window['setInterval']('foo');`,
			wantSpan:    `window['setInterval']('foo')`,
			wantId:      "impliedEval",
			wantPhrases: []string{"string", "function"},
		},
		{
			source:      `execScript("x = 1;");`,
			wantSpan:    `execScript("x = 1;")`,
			wantId:      "execScript",
			wantPhrases: []string{"execScript"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := runImpliedEvalWithGlobals(t, testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected exactly one finding, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]

			// The harness writes the fixture through RunTypedFiles, which does not trim, so the
			// span offsets index this same string.
			reported := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.wantSpan {
				t.Errorf("finding points at %q, want %q", reported, testCase.wantSpan)
			}
			if diagnostic.Message.Id != testCase.wantId {
				t.Errorf("message id is %q, want %q", diagnostic.Message.Id, testCase.wantId)
			}
			for _, phrase := range testCase.wantPhrases {
				if !strings.Contains(diagnostic.Message.Description, phrase) {
					t.Errorf("description does not mention %q: %s", phrase, diagnostic.Message.Description)
				}
			}
			// The description must say why rather than restate the rule name.
			if len(diagnostic.Message.Description) < 40 {
				t.Errorf("description is too short to explain the problem: %q", diagnostic.Message.Description)
			}
		})
	}
}
