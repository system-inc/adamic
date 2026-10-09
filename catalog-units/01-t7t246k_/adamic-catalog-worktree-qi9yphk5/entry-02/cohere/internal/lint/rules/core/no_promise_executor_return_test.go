package core

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// promiseExecutorFile is where the fixtures pretend to live.
const promiseExecutorFile = "/repository/source/PromiseExecutor.ts"

// decodedPromiseExecutorOptions routes a fixture's options through the rule's own exported
// decoder rather than building the struct directly.
//
// An empty string means the rule is configured as bare "error", which is what the live config
// does and what hands the rule nil.
func decodedPromiseExecutorOptions(t *testing.T, raw string) any {
	t.Helper()
	if raw == "" {
		return nil
	}
	options, err := DecodeNoPromiseExecutorReturnOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return options
}

// The corpus is ESLint's own, at `tests/lib/rules/no-promise-executor-return.js`, extracted
// programmatically by loading that file with a stub rule tester rather than retyped, so every
// string below is upstream's own bytes.
//
// 66 clean cases upstream, of which 59 are here. The other seven need a globals or a commonjs
// global-return surface this harness does not have; they are recorded in
// `TestNoPromiseExecutorReturnCasesThisHarnessCannotExpress` rather than deleted or greened.
func TestNoPromiseExecutorReturnStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
	}{
		{"function foo(resolve, reject) { return 1; }", ""},
		{"function Promise(resolve, reject) { return 1; }", ""},
		{"(function (resolve, reject) { return 1; })", ""},
		{"(function foo(resolve, reject) { return 1; })", ""},
		{"(function Promise(resolve, reject) { return 1; })", ""},
		{"var foo = function (resolve, reject) { return 1; }", ""},
		{"var foo = function Promise(resolve, reject) { return 1; }", ""},
		{"var Promise = function (resolve, reject) { return 1; }", ""},
		{"(resolve, reject) => { return 1; }", ""},
		{"(resolve, reject) => 1", ""},
		{"var foo = (resolve, reject) => { return 1; }", ""},
		{"var Promise = (resolve, reject) => { return 1; }", ""},
		{"var foo = (resolve, reject) => 1", ""},
		{"var Promise = (resolve, reject) => 1", ""},
		{"var foo = { bar(resolve, reject) { return 1; } }", ""},
		{"var foo = { Promise(resolve, reject) { return 1; } }", ""},
		{"new foo(function (resolve, reject) { return 1; });", ""},
		{"new foo(function bar(resolve, reject) { return 1; });", ""},
		{"new foo(function Promise(resolve, reject) { return 1; });", ""},
		{"new foo((resolve, reject) => { return 1; });", ""},
		{"new foo((resolve, reject) => 1);", ""},
		{"new promise(function foo(resolve, reject) { return 1; });", ""},
		{"new Promise.foo(function foo(resolve, reject) { return 1; });", ""},
		{"new foo.Promise(function foo(resolve, reject) { return 1; });", ""},
		{"new Promise.Promise(function foo(resolve, reject) { return 1; });", ""},
		{"new Promise()(function foo(resolve, reject) { return 1; });", ""},
		{"Promise(function (resolve, reject) { return 1; });", ""},
		{"Promise((resolve, reject) => { return 1; });", ""},
		{"Promise((resolve, reject) => 1);", ""},
		{"new Promise(foo, function (resolve, reject) { return 1; });", ""},
		{"new Promise(foo, (resolve, reject) => { return 1; });", ""},
		{"new Promise(foo, (resolve, reject) => 1);", ""},
		{"let Promise; new Promise(function (resolve, reject) { return 1; });", ""},
		{"function f() { new Promise((resolve, reject) => { return 1; }); var Promise; }", ""},
		{"function f(Promise) { new Promise((resolve, reject) => 1); }", ""},
		{"if (x) { const Promise = foo(); new Promise(function (resolve, reject) { return 1; }); }", ""},
		{"x = function Promise() { new Promise((resolve, reject) => { return 1; }); }", ""},
		{"new Promise(function (resolve, reject) { return; });", ""},
		{"new Promise(function (resolve, reject) { reject(new Error()); return; });", ""},
		{"new Promise(function (resolve, reject) { if (foo) { return; } });", ""},
		{"new Promise((resolve, reject) => { return; });", ""},
		{"new Promise((resolve, reject) => { if (foo) { resolve(1); return; } reject(new Error()); });", ""},
		{"new Promise(function (resolve, reject) { throw new Error(); });", ""},
		{"new Promise((resolve, reject) => { throw new Error(); });", ""},
		{"new Promise(function (resolve, reject) { function foo() { return 1; } });", ""},
		{"new Promise((resolve, reject) => { (function foo() { return 1; })(); });", ""},
		{"new Promise(function (resolve, reject) { () => { return 1; } });", ""},
		{"new Promise((resolve, reject) => { () => 1 });", ""},
		{"function foo() { return new Promise(function (resolve, reject) { resolve(bar); }) };", ""},
		{"foo => new Promise((resolve, reject) => { bar(foo, (err, data) => { if (err) { reject(err); return; } resolve(data); })});", ""},
		{"new Promise(function (resolve, reject) {}); function foo() { return 1; }", ""},
		{"new Promise((resolve, reject) => {}); (function () { return 1; });", ""},
		{"new Promise(function (resolve, reject) {}); () => { return 1; };", ""},
		{"new Promise((resolve, reject) => {}); () => 1;", ""},
		{"new Promise((r) => void cbf(r));", "{\"allowVoid\": true}"},
		{"new Promise(r => void 0)", "{\"allowVoid\": true}"},
		{"new Promise(r => { return void 0 })", "{\"allowVoid\": true}"},
		{"new Promise(r => { if (foo) { return void 0 } return void 0 })", "{\"allowVoid\": true}"},
		{"new Promise(r => {0})", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoPromiseExecutorReturn,
				promiseExecutorFile, testCase.sourceText, decodedPromiseExecutorOptions(t, testCase.options)))
		})
	}
}

// Every reporting case upstream ships, with its own suggestion list.
//
// `suggestions: null` and `suggestions: []` are different upstream and both are reproduced: null
// means the rule offers nothing, and the empty list means it considered offering and declined,
// which happens for an unnamed function or class body that braces would make invalid syntax.
// Both arrive here as a zero-length want, and the distinction is recorded in the rule.
func TestNoPromiseExecutorReturnFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		options     string
		wantSuggest []string
	}{
		{"new Promise(function (resolve, reject) { return 1; })", "", nil},
		{"new Promise((resolve, reject) => resolve(1))", "{\"allowVoid\": true}", []string{"prependVoid", "wrapBraces"}},
		{"new Promise((resolve, reject) => { return 1 })", "{\"allowVoid\": true}", []string{"prependVoid"}},
		{"new Promise(r => 1)", "{\"allowVoid\": true}", []string{"prependVoid", "wrapBraces"}},
		{"new Promise(r => 1 ? 2 : 3)", "{\"allowVoid\": true}", []string{"prependVoid", "wrapBraces"}},
		{"new Promise(r => (1 ? 2 : 3))", "{\"allowVoid\": true}", []string{"prependVoid", "wrapBraces"}},
		{"new Promise(r => (1))", "{\"allowVoid\": true}", []string{"prependVoid", "wrapBraces"}},
		{"new Promise(r => () => {})", "{\"allowVoid\": true}", []string{"prependVoid", "wrapBraces"}},
		{"new Promise(r => null)", "{\"allowVoid\": true}", []string{"prependVoid", "wrapBraces"}},
		{"new Promise(r => null)", "{\"allowVoid\": false}", []string{"wrapBraces"}},
		{"new Promise(r => /*hi*/ ~0)", "{\"allowVoid\": true}", []string{"prependVoid", "wrapBraces"}},
		{"new Promise(r => /*hi*/ ~0)", "{\"allowVoid\": false}", []string{"wrapBraces"}},
		{"new Promise(r => { return 0 })", "{\"allowVoid\": true}", []string{"prependVoid"}},
		{"new Promise(r => { return 0 })", "{\"allowVoid\": false}", nil},
		{"new Promise(r => { if (foo) { return void 0 } return 0 })", "{\"allowVoid\": true}", []string{"prependVoid"}},
		{"new Promise(resolve => { return (foo = resolve(1)); })", "{\"allowVoid\": true}", []string{"prependVoid"}},
		{"new Promise(resolve => r = resolve)", "{\"allowVoid\": true}", []string{"prependVoid", "wrapBraces"}},
		{"new Promise(r => { return(1) })", "{\"allowVoid\": true}", []string{"prependVoid"}},
		{"new Promise(r =>1)", "{\"allowVoid\": true}", []string{"prependVoid", "wrapBraces"}},
		{"new Promise(r => ((1)))", "{\"allowVoid\": true}", []string{"prependVoid", "wrapBraces"}},
		{"new Promise(function foo(resolve, reject) { return 1; })", "", nil},
		{"new Promise((resolve, reject) => { return 1; })", "", nil},
		{"new Promise(function (resolve, reject) { return undefined; })", "", nil},
		{"new Promise((resolve, reject) => { return null; })", "", nil},
		{"new Promise(function (resolve, reject) { return false; })", "", nil},
		{"new Promise((resolve, reject) => resolve)", "", []string{"wrapBraces"}},
		{"new Promise((resolve, reject) => null)", "", []string{"wrapBraces"}},
		{"new Promise(function (resolve, reject) { return resolve(foo); })", "", nil},
		{"new Promise((resolve, reject) => { return reject(foo); })", "", nil},
		{"new Promise((resolve, reject) => x + y)", "", []string{"wrapBraces"}},
		{"new Promise((resolve, reject) => { return Promise.resolve(42); })", "", nil},
		{"new Promise(function (resolve, reject) { if (foo) { return 1; } })", "", nil},
		{"new Promise((resolve, reject) => { try { return 1; } catch(e) {} })", "", nil},
		{"new Promise(function (resolve, reject) { while (foo){ if (bar) break; else return 1; } })", "", nil},
		{"new Promise(() => { return void 1; })", "", nil},
		{"new Promise(() => (1))", "", []string{"wrapBraces"}},
		{"() => new Promise(() => ({}));", "", []string{"wrapBraces"}},
		{"new Promise(function () { return 1; })", "", nil},
		{"new Promise(() => { return 1; })", "", nil},
		{"new Promise(() => 1)", "", []string{"wrapBraces"}},
		{"function foo() {} new Promise(function () { return 1; });", "", nil},
		{"function foo() { return; } new Promise(() => { return 1; });", "", nil},
		{"function foo() { return 1; } new Promise(() => { return 2; });", "", nil},
		{"function foo () { return new Promise(function () { return 1; }); }", "", nil},
		{"function foo() { return new Promise(() => { bar(() => { return 1; }); return false; }); }", "", nil},
		{"() => new Promise(() => { if (foo) { return 0; } else bar(() => { return 1; }); })", "", nil},
		{"function foo () { return 1; return new Promise(function () { return 2; }); return 3;}", "", nil},
		{"() => 1; new Promise(() => { return 1; })", "", nil},
		{"new Promise(function () { return 1; }); function foo() { return 1; } ", "", nil},
		{"() => new Promise(() => { return 1; });", "", nil},
		{"() => new Promise(() => 1);", "", []string{"wrapBraces"}},
		{"() => new Promise(() => () => 1);", "", []string{"wrapBraces"}},
		{"() => new Promise(() => async () => 1);", "", []string{"wrapBraces"}},
		{"() => new Promise(() => function () {});", "", nil},
		{"() => new Promise(() => class {});", "", nil},
		{"() => new Promise(() => function foo() {});", "", []string{"wrapBraces"}},
		{"() => new Promise(() => class Foo {});", "", []string{"wrapBraces"}},
		{"() => new Promise(() => []);", "", []string{"wrapBraces"}},
		{"new Promise((Promise) => { return 1; })", "", nil},
		{"new Promise(function Promise(resolve, reject) { return 1; })", "", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoPromiseExecutorReturn,
				promiseExecutorFile, testCase.sourceText, decodedPromiseExecutorOptions(t, testCase.options))
			rule_testing.ExpectFindings(t, result, "returnsValue")
			got := []string{}
			for _, suggestion := range result.Diagnostics[0].Suggestions {
				got = append(got, suggestion.Message.Id)
			}
			if strings.Join(got, ",") != strings.Join(testCase.wantSuggest, ",") {
				t.Errorf("offered suggestions %v, want %v", got, testCase.wantSuggest)
			}
		})
	}
}

// The spans upstream states, converted into the slice the finding names. The columns for the
// parenthesized rows were measured against the installed eslint at 10.8.1 rather than read off
// this rule.
//
// Those rows are the one place our parser and upstream genuinely disagree about what node this is.
// Espree folds parentheses away, so `new Promise(r => (1))` gives upstream the literal `1` as the
// arrow body and it reports the `1`. typescript-go produces a real `KindParenthesizedExpression`,
// so reporting the body node directly would span `(1)`. `((1))` unwraps all the way down to `1`,
// which is why the rule unwraps in a loop rather than one step.
//
// The FIX anchors run the OTHER way. `wrapBraces` on `new Promise(r => (1))` writes `{(1)}`, so
// the braces sit outside the parentheses the span excludes. One node, two answers, and only a
// span assertion beside a suggestion-output assertion records which is which.
func TestNoPromiseExecutorReturnSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		options      string
		wantReported string
	}{
		{"new Promise(r => (1))", "{\"allowVoid\":true}", "1"},
		{"new Promise(r => ((1)))", "{\"allowVoid\":true}", "1"},
		{"new Promise(r => (1 ? 2 : 3))", "{\"allowVoid\":true}", "1 ? 2 : 3"},
		{"new Promise(() => ({}))", "", "{}"},
		{"new Promise(r => 1)", "{\"allowVoid\":true}", "1"},
		{"new Promise(r => /*hi*/ ~0)", "{\"allowVoid\":true}", "~0"},
		{"new Promise(function (resolve, reject) { return 1; })", "", "return 1;"},
		{"new Promise(r => { return 0 })", "", "return 0"},
		{"() => new Promise(() => function () {});", "", "function () {}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoPromiseExecutorReturn,
				promiseExecutorFile, testCase.sourceText, decodedPromiseExecutorOptions(t, testCase.options))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want one finding, got %d", len(result.Diagnostics))
			}
			source := result.SourceFile.Text()
			reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantReported {
				t.Errorf("the finding points at %q, want %q", reported, testCase.wantReported)
			}
		})
	}
}

// The six clean cases upstream ships that this harness cannot express, recorded rather than
// deleted or quietly greened.
//
// Two need a globals surface, one through `languageOptions.globals` and one through a
// `/* globals Promise:off */` directive comment, both of which turn the global `Promise` off so the
// executor is not one. Five need commonjs or `globalReturn`, where a `return` at the top level of a
// module is legal; our harness parses TypeScript modules where a bare top-level return is a syntax
// error rather than a construct the rule declines.
//
// The directive-comment one is worth naming separately because it is the case that FAILED rather
// than merely being unexpressible: this port reports it, because `Promise` there does resolve to the
// standard library and nothing reads the comment. Reproducing upstream's silence needs a
// configured-globals layer, which cohere does not carry.
//
// Both are facts about the harness rather than about the rule, so hiding them inside a relaxed
// rule would turn them into facts about the rule. What IS asserted here is the part that does not
// depend on the missing surface: each input reaches the rule and the rule does not crash on it.
func TestNoPromiseExecutorReturnCasesThisHarnessCannotExpress(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		reason     string
	}{
		{"new Promise((resolve, reject) => { return 1; });", "needs a configured-globals surface"},
		{"/* globals Promise:off */ new Promise(function (resolve, reject) { return 1; });", "needs a directive-comment globals surface"},
		{"return 1;", "needs commonjs or globalReturn, where a top-level return is legal"},
		{"return 1;", "needs commonjs or globalReturn, where a top-level return is legal"},
		{"return 1; function foo(){ return 1; } return 1;", "needs commonjs or globalReturn, where a top-level return is legal"},
		{"function foo(){} return 1; var bar = function*(){ return 1; }; return 1; var baz = () => {}; return 1;", "needs commonjs or globalReturn, where a top-level return is legal"},
		{"new Promise(function (resolve, reject) {}); return 1;", "needs commonjs or globalReturn, where a top-level return is legal"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			// Not asserted clean, because the surface that makes them clean upstream is absent here.
			// Asserted only to run, so a later crash on one of these shapes fails loudly.
			rule_testing.RunTypedWithOptions(t, NoPromiseExecutorReturn, promiseExecutorFile,
				testCase.sourceText, nil)
			t.Logf("upstream holds this clean, and reproducing that %s", testCase.reason)
		})
	}
}

// The decoder is the line with no upstream counterpart. `allowVoid` defaults to FALSE, so the
// zero value happens to be right, and the fallback is written out anyway so a later default change
// cannot invert the rule silently.
func TestDecodeNoPromiseExecutorReturnOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{"empty input falls back to the default", "", false},
		{"an empty object falls back to the default", "{}", false},
		{"an explicit false", `{"allowVoid": false}`, false},
		{"an explicit true", `{"allowVoid": true}`, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoPromiseExecutorReturnOptions(json.RawMessage(testCase.raw))
			if err != nil {
				t.Fatalf("the decoder refused %q: %v", testCase.raw, err)
			}
			options, isOptions := decoded.(NoPromiseExecutorReturnOptions)
			if !isOptions {
				t.Fatalf("the decoder returned %T rather than NoPromiseExecutorReturnOptions", decoded)
			}
			got := options.AllowVoid != nil && *options.AllowVoid
			if got != testCase.want {
				t.Errorf("allowVoid decodes to %v, want %v", got, testCase.want)
			}
		})
	}
}

// A rule configured as bare "error" reaches Run with nil rather than with an options struct. No
// fixture routed through the decoder can see that path.
func TestNoPromiseExecutorReturnWithNilOptionsDefaultsToDisallowingVoid(t *testing.T) {
	t.Parallel()

	// Under the default, `return void 0` still reports: allowVoid is what exempts it.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoPromiseExecutorReturn,
		promiseExecutorFile, "new Promise(() => { return void 1; })", nil), "returnsValue")
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoPromiseExecutorReturn,
		promiseExecutorFile, "new Promise(function (resolve, reject) { return; })", nil))
}

// The rule declares NeedsTypeChecker, so the plain harness hands it a nil checker and the guard
// makes it go completely silent. That is the more dangerous of the two failure modes, because
// every clean case would then pass vacuously.
func TestNoPromiseExecutorReturnRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	// Both listeners, because they guard separately and a test covering only one leaves the other's
	// guard unmeasured. A mutation removing the arrow listener's guard survived a version of this
	// test that only used a `return` statement.
	cases := []string{
		"new Promise(() => { return 1; })",
		"new Promise(() => 1)",
	}
	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoPromiseExecutorReturn,
				promiseExecutorFile, sourceText, nil), "returnsValue")
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoPromiseExecutorReturn,
				promiseExecutorFile, sourceText, nil))
		})
	}
}

// The message text is asserted against a literal typed here rather than against the rule's own
// constant, because comparing a diagnostic to the constant it was built from moves both sides
// together under mutation and asserts nothing.
func TestNoPromiseExecutorReturnMessages(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedWithOptions(t, NoPromiseExecutorReturn, promiseExecutorFile,
		"new Promise(r => 1)", decodedPromiseExecutorOptions(t, `{"allowVoid": true}`))
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "returnsValue" {
		t.Errorf("message id is %q, want %q", got, "returnsValue")
	}
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, "This returns a value from a promise executor") {
		t.Errorf("message description starts %q, which is not the sentence this rule reports", got)
	}
	wantSuggestionIds := []string{"prependVoid", "wrapBraces"}
	for index, suggestion := range result.Diagnostics[0].Suggestions {
		if suggestion.Message.Id != wantSuggestionIds[index] {
			t.Errorf("suggestion %d is %q, want %q", index, suggestion.Message.Id, wantSuggestionIds[index])
		}
	}
}

// The 41 suggestion outputs upstream asserts, byte for byte.
//
// `rule_testing` can apply a FIX but not a SUGGESTION, so the applier below is hand-rolled. That
// is not optional coverage: a suggestion writing the right text over the wrong span passes any
// assertion that only compares the text, and this rule's two suggestions are built from two
// separate insertions each, so the span is most of what can go wrong.
//
// Every `output` here is upstream's own, extracted from its corpus rather than retyped.
func TestNoPromiseExecutorReturnSuggestionOutputs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
		index      int
		wantId     string
		wantOutput string
	}{
		{"new Promise((resolve, reject) => resolve(1))", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise((resolve, reject) => void resolve(1))"},
		{"new Promise((resolve, reject) => resolve(1))", "{\"allowVoid\": true}", 1, "wrapBraces", "new Promise((resolve, reject) => {resolve(1)})"},
		{"new Promise((resolve, reject) => { return 1 })", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise((resolve, reject) => { return void 1 })"},
		{"new Promise(r => 1)", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r => void 1)"},
		{"new Promise(r => 1)", "{\"allowVoid\": true}", 1, "wrapBraces", "new Promise(r => {1})"},
		{"new Promise(r => 1 ? 2 : 3)", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r => void (1 ? 2 : 3))"},
		{"new Promise(r => 1 ? 2 : 3)", "{\"allowVoid\": true}", 1, "wrapBraces", "new Promise(r => {1 ? 2 : 3})"},
		{"new Promise(r => (1 ? 2 : 3))", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r => void (1 ? 2 : 3))"},
		{"new Promise(r => (1 ? 2 : 3))", "{\"allowVoid\": true}", 1, "wrapBraces", "new Promise(r => {(1 ? 2 : 3)})"},
		{"new Promise(r => (1))", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r => void (1))"},
		{"new Promise(r => (1))", "{\"allowVoid\": true}", 1, "wrapBraces", "new Promise(r => {(1)})"},
		{"new Promise(r => () => {})", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r => void (() => {}))"},
		{"new Promise(r => () => {})", "{\"allowVoid\": true}", 1, "wrapBraces", "new Promise(r => {() => {}})"},
		{"new Promise(r => null)", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r => void null)"},
		{"new Promise(r => null)", "{\"allowVoid\": true}", 1, "wrapBraces", "new Promise(r => {null})"},
		{"new Promise(r => null)", "{\"allowVoid\": false}", 0, "wrapBraces", "new Promise(r => {null})"},
		{"new Promise(r => /*hi*/ ~0)", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r => /*hi*/ void ~0)"},
		{"new Promise(r => /*hi*/ ~0)", "{\"allowVoid\": true}", 1, "wrapBraces", "new Promise(r => /*hi*/ {~0})"},
		{"new Promise(r => /*hi*/ ~0)", "{\"allowVoid\": false}", 0, "wrapBraces", "new Promise(r => /*hi*/ {~0})"},
		{"new Promise(r => { return 0 })", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r => { return void 0 })"},
		{"new Promise(r => { if (foo) { return void 0 } return 0 })", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r => { if (foo) { return void 0 } return void 0 })"},
		{"new Promise(resolve => { return (foo = resolve(1)); })", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(resolve => { return void (foo = resolve(1)); })"},
		{"new Promise(resolve => r = resolve)", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(resolve => void (r = resolve))"},
		{"new Promise(resolve => r = resolve)", "{\"allowVoid\": true}", 1, "wrapBraces", "new Promise(resolve => {r = resolve})"},
		{"new Promise(r => { return(1) })", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r => { return void (1) })"},
		{"new Promise(r =>1)", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r =>void 1)"},
		{"new Promise(r =>1)", "{\"allowVoid\": true}", 1, "wrapBraces", "new Promise(r =>{1})"},
		{"new Promise(r => ((1)))", "{\"allowVoid\": true}", 0, "prependVoid", "new Promise(r => void ((1)))"},
		{"new Promise(r => ((1)))", "{\"allowVoid\": true}", 1, "wrapBraces", "new Promise(r => {((1))})"},
		{"new Promise((resolve, reject) => resolve)", "", 0, "wrapBraces", "new Promise((resolve, reject) => {resolve})"},
		{"new Promise((resolve, reject) => null)", "", 0, "wrapBraces", "new Promise((resolve, reject) => {null})"},
		{"new Promise((resolve, reject) => x + y)", "", 0, "wrapBraces", "new Promise((resolve, reject) => {x + y})"},
		{"new Promise(() => (1))", "", 0, "wrapBraces", "new Promise(() => {(1)})"},
		{"() => new Promise(() => ({}));", "", 0, "wrapBraces", "() => new Promise(() => {({})});"},
		{"new Promise(() => 1)", "", 0, "wrapBraces", "new Promise(() => {1})"},
		{"() => new Promise(() => 1);", "", 0, "wrapBraces", "() => new Promise(() => {1});"},
		{"() => new Promise(() => () => 1);", "", 0, "wrapBraces", "() => new Promise(() => {() => 1});"},
		{"() => new Promise(() => async () => 1);", "", 0, "wrapBraces", "() => new Promise(() => {async () => 1});"},
		{"() => new Promise(() => function foo() {});", "", 0, "wrapBraces", "() => new Promise(() => {function foo() {}});"},
		{"() => new Promise(() => class Foo {});", "", 0, "wrapBraces", "() => new Promise(() => {class Foo {}});"},
		{"() => new Promise(() => []);", "", 0, "wrapBraces", "() => new Promise(() => {[]});"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText+"/"+testCase.wantId, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoPromiseExecutorReturn,
				promiseExecutorFile, testCase.sourceText, decodedPromiseExecutorOptions(t, testCase.options))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want one finding, got %d", len(result.Diagnostics))
			}
			suggestions := result.Diagnostics[0].Suggestions
			if testCase.index >= len(suggestions) {
				t.Fatalf("want a suggestion at index %d, got %d suggestions", testCase.index, len(suggestions))
			}
			suggestion := suggestions[testCase.index]
			if suggestion.Message.Id != testCase.wantId {
				t.Fatalf("suggestion %d is %q, want %q", testCase.index, suggestion.Message.Id, testCase.wantId)
			}
			// The harness applies fixes and not suggestions, so this replays the suggestion's own
			// fixes the way the engine would: back to front, so an earlier edit cannot move the
			// offsets a later one was computed against.
			//
			// The expectation is transformed the same way `RunTyped` transformed the input, rather
			// than the rule being padded to make the comparison line up. `program.go` writes each
			// fixture as `strings.TrimSpace(contents)+"\n"`, so the file the rule saw ends in a
			// newline that upstream's `output` does not carry, and five byte-correct repairs failed
			// on it before this line existed.
			wantOutput := strings.TrimSpace(testCase.wantOutput) + "\n"
			if got := applySuggestion(t, result.SourceFile.Text(), suggestion); got != wantOutput {
				t.Errorf("applying %q produced:\n  %q\nwant:\n  %q", testCase.wantId, got, wantOutput)
			}
		})
	}
}

// applySuggestion replays one suggestion's fixes into the source, back to front.
//
// Same ordering the fix engine and `rule_testing.ExpectFixedSource` use, and for the same reason.
// Written here rather than in the harness because it is the only rule in this batch that needs it
// and the harness has no suggestion support at all; if a second rule wants it, it belongs there.
func applySuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
	t.Helper()
	if len(suggestion.Fixes) == 0 {
		t.Fatalf("suggestion %q carries no fixes, so there is nothing to apply", suggestion.Message.Id)
	}
	ordered := make([]rule.Fix, len(suggestion.Fixes))
	copy(ordered, suggestion.Fixes)
	sort.Slice(ordered, func(first, second int) bool {
		return ordered[first].Range.Pos() > ordered[second].Range.Pos()
	})
	for _, fix := range ordered {
		if fix.Range.Pos() < 0 || fix.Range.End() > len(source) || fix.Range.Pos() > fix.Range.End() {
			t.Fatalf("fix range [%d,%d) is outside the source, which is a defect in the rule",
				fix.Range.Pos(), fix.Range.End())
		}
		source = source[:fix.Range.Pos()] + fix.Text + source[fix.Range.End():]
	}
	return source
}

// Other global constructors called with `new` and a function first argument.
//
// A mutation dropping `callee.Text() != "Promise"` survived the whole imported corpus, because the
// only non-Promise callees upstream writes are `new promise` (lowercase, which resolves to nothing
// here) and `new foo` (undeclared), so the name test and the global-resolution test agree on every
// one of them. These five resolve to the standard library exactly as `Promise` does and differ only
// in the name, which is what separates the two tests.
//
// All five measured clean against the installed eslint at 10.8.1.
func TestNoPromiseExecutorReturnStaysSilentOnOtherGlobalConstructors(t *testing.T) {
	t.Parallel()

	cases := []string{
		"new Array(function (resolve, reject) { return 1; });",
		"new Map(function (resolve, reject) { return 1; });",
		"new Function(function (resolve, reject) { return 1; });",
		"new Set(() => 1);",
		"new Date(() => 1);",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoPromiseExecutorReturn,
				promiseExecutorFile, sourceText, nil))
		})
	}
}
