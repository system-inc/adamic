package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// requireOptimizationFile is where the fixtures pretend to live.
//
// A `.tsx` name because several cases carry JSX. It is NOT load-bearing and there is no file gate:
// `TestRequireOptimizationHasNoFileGate` reports on a `.ts` file too. Three shipped rules in this
// package used to gate on the suffix through an `isJsxFileName` helper, which was oxc residue rather
// than upstream behavior and blinded them across every `.ts` file; the gate and the helper are both
// gone now, and this rule never had one.
const requireOptimizationFile = "/repository/source/RequireOptimization.tsx"

// The corpus is `eslint-plugin-react`'s own, extracted rather than retyped.
//
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/require-optimization.js` holds 15 valid and
// 8 invalid cases. Every string below was pulled out of that file by loading it with a stubbed
// `RuleTester` and serializing the captured object to JSON, then emitted into these tables by a
// generator. No case was typed by hand.
//
// All 23 were run against the installed build, 7.37.5, through the ESLint Linter API with the
// TypeScript parser, and the clone and the installed build agreed on every one. Five cases use
// decorators, which the TypeScript parser reads natively.
//
// # The options column is RAW JSON, on purpose
//
// Every case routes through `DecodeRequireOptimizationOptions`, the same function the config layer
// calls, so the decoder's empty-input path is under test rather than bypassed. An empty string means
// a bare severity, which is what hands a real rule nil options.

// TestRequireOptimizationFires runs upstream's reporting cases.
func TestRequireOptimizationFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantIds    []string
	}{
		{"upstream invalid-0", "\n        import React from \"react\";\n        class YourComponent extends React.Component {}\n      ", "", []string{"noShouldComponentUpdate"}},
		{"upstream invalid-1", "\n        import React from \"react\";\n        class YourComponent extends React.Component {\n          handleClick() {}\n          render() {\n            return <div onClick={this.handleClick}>123</div>\n          }\n        }\n      ", "", []string{"noShouldComponentUpdate"}},
		{"upstream invalid-2", "\n        import React from \"react\";\n        class YourComponent extends React.Component {\n          handleClick = () => {}\n          render() {\n            return <div onClick={this.handleClick}>123</div>\n          }\n        }\n      ", "", []string{"noShouldComponentUpdate"}},
		{"upstream invalid-3", "\n        import React, {Component} from \"react\";\n        class YourComponent extends Component {}\n      ", "", []string{"noShouldComponentUpdate"}},
		{"upstream invalid-4", "\n        import React from \"react\";\n        createReactClass({})\n      ", "", []string{"noShouldComponentUpdate"}},
		{"upstream invalid-5", "\n        import React from \"react\";\n        createReactClass({\n          mixins: [RandomMixin]\n        })\n      ", "", []string{"noShouldComponentUpdate"}},
		{"upstream invalid-6", "\n        @reactMixin.decorate(SomeOtherMixin)\n        class DecoratedComponent extends Component {}\n      ", "", []string{"noShouldComponentUpdate"}},
		{"upstream invalid-7", "\n        @bar\n        @pure\n        @foo\n        class DecoratedComponent extends Component {}\n      ", "{\"allowDecorators\":[\"renderPure\",\"pureRender\"]}", []string{"noShouldComponentUpdate"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runRequireOptimization(t, testCase.sourceText, testCase.rawOptions)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestRequireOptimizationStaysSilent runs upstream's clean cases.
//
// The three function-component cases are the ones worth naming: a function declaration, a function
// expression and an arrow are all exempt by design upstream, with the comment "Stateless Functional
// Components cannot be optimized (yet)". That is why this rule is portable without the component
// registry, and these three are what pin it.
func TestRequireOptimizationStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"upstream valid-0", "\n        class A {}\n      ", ""},
		{"upstream valid-1", "\n        import React from \"react\";\n        class YourComponent extends React.Component {\n          shouldComponentUpdate () {}\n        }\n      ", ""},
		{"upstream valid-2", "\n        import React, {Component} from \"react\";\n        class YourComponent extends Component {\n          shouldComponentUpdate () {}\n        }\n      ", ""},
		{"upstream valid-3", "\n        import React, {Component} from \"react\";\n        @reactMixin.decorate(PureRenderMixin)\n        class YourComponent extends Component {\n          componentDidMount () {}\n          render() {}\n        }\n      ", ""},
		{"upstream valid-4", "\n        import React from \"react\";\n        createReactClass({\n          shouldComponentUpdate: function () {}\n        })\n      ", ""},
		{"upstream valid-5", "\n        import React from \"react\";\n        createReactClass({\n          mixins: [PureRenderMixin]\n        })\n      ", ""},
		{"upstream valid-6", "\n        @reactMixin.decorate(PureRenderMixin)\n        class DecoratedComponent extends Component {}\n      ", ""},
		{"upstream valid-7", "\n        const FunctionalComponent = function (props) {\n          return <div />;\n        }\n      ", ""},
		{"upstream valid-8", "\n        function FunctionalComponent(props) {\n          return <div />;\n        }\n      ", ""},
		{"upstream valid-9", "\n        const FunctionalComponent = (props) => {\n          return <div />;\n        }\n      ", ""},
		{"upstream valid-10", "\n        @bar\n        @pureRender\n        @foo\n        class DecoratedComponent extends Component {}\n      ", "{\"allowDecorators\":[\"renderPure\",\"pureRender\"]}"},
		{"upstream valid-11", "\n        import React from \"react\";\n        class YourComponent extends React.PureComponent {}\n      ", "{\"allowDecorators\":[\"renderPure\",\"pureRender\"]}"},
		{"upstream valid-12", "\n        import React, {PureComponent} from \"react\";\n        class YourComponent extends PureComponent {}\n      ", "{\"allowDecorators\":[\"renderPure\",\"pureRender\"]}"},
		{"upstream valid-13", "\n        const obj = { prop: [,,,,,] }\n      ", ""},
		{"upstream valid-14", "\n        import React from \"react\";\n        class YourComponent extends React.Component {\n          handleClick = () => {}\n          shouldComponentUpdate(){\n            return true;\n          }\n          render() {\n            return <div onClick={this.handleClick}>123</div>\n          }\n        }\n      ", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runRequireOptimization(t, testCase.sourceText, testCase.rawOptions))
		})
	}
}

// runRequireOptimization drives the rule through its own decoder.
//
// `Run` rather than `RunTyped` because the rule is purely syntactic: it asks what a class extends,
// what members it declares, and what decorators it carries, and every one of those is on the AST.
// `TestRequireOptimizationNeedsNoChecker` pins that a reporting case is identical under both
// harnesses, so a later reader adding `NeedsTypeChecker` has to justify it.
func runRequireOptimization(t *testing.T, sourceText string, rawOptions string) rule_testing.Result {
	t.Helper()
	decoded, err := DecodeRequireOptimizationOptions([]byte(rawOptions))
	if err != nil {
		t.Fatalf("decoding %q: %v", rawOptions, err)
	}
	return rule_testing.RunWithOptions(t, RequireOptimization, requireOptimizationFile, sourceText, decoded)
}

// TestRequireOptimizationObjectArmReproducesUpstreamsAccident pins the strangest thing here.
//
// `createReactClass({})` reports and `createReactClass({ foo: function () {} })` does not, and the
// difference has nothing to do with `shouldComponentUpdate`. Upstream's `FunctionExpression` and
// `ArrowFunctionExpression` listeners mark their enclosing component as already optimized, and for a
// `createReactClass` call that component is the OBJECT, so any function value anywhere in it exempts
// the whole thing.
//
// The corpus cannot see this. Both of its reporting object cases contain no function at all, so a
// port that had written the intuitive rule, "an object with no shouldComponentUpdate", would pass
// them and then report on nearly every `createReactClass` in a real codebase. Every row below was
// measured on the installed build before it was written here.
func TestRequireOptimizationObjectArmReproducesUpstreamsAccident(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"an empty object reports", "declare function createReactClass(spec: any): any;\ncreateReactClass({});\n", []string{"noShouldComponentUpdate"}},
		{"a non-function value does not exempt", "declare function createReactClass(spec: any): any;\ncreateReactClass({ foo: 1 });\n", []string{"noShouldComponentUpdate"}},
		{"an unrelated mixin does not exempt", "declare function createReactClass(spec: any): any;\ndeclare const RandomMixin: any;\ncreateReactClass({ mixins: [RandomMixin] });\n", []string{"noShouldComponentUpdate"}},
		{"ANY function expression exempts", "declare function createReactClass(spec: any): any;\ncreateReactClass({ foo: function () { return 1; } });\n", nil},
		{"ANY arrow exempts", "declare function createReactClass(spec: any): any;\ncreateReactClass({ foo: () => 1 });\n", nil},
		{"ANY method shorthand exempts", "declare function createReactClass(spec: any): any;\ncreateReactClass({ foo() { return 1; } });\n", nil},
		{"the mixin arm exempts", "declare function createReactClass(spec: any): any;\ndeclare const PureRenderMixin: any;\ncreateReactClass({ mixins: [PureRenderMixin] });\n", nil},
		{"a shouldComponentUpdate property exempts", "declare function createReactClass(spec: any): any;\ncreateReactClass({ shouldComponentUpdate: 1 });\n", nil},
		// A bare object with no `createReactClass` around it is not a component at all, so the
		// object arm never sees it. Upstream reaches the object through its component registry;
		// this port reaches it through the call, and this row pins that the two agree.
		{"a bare object literal is not a component", "const x = {};\n", nil},
		{"a bare object with a function is not a component either", "const x = { foo: function () { return 1; } };\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runRequireOptimization(t, testCase.sourceText, ""), testCase.wantIds...)
		})
	}
}

// TestRequireOptimizationMatchesUpstreamOnShapesTheCorpusOmits pins measured verdicts.
//
// Each row was run against the installed build before it was written. The class-expression row and
// the two property-spelling rows are the ones a port is most likely to get wrong in the widening
// direction, because all three read as obviously equivalent to a shape upstream does exempt.
func TestRequireOptimizationMatchesUpstreamOnShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"a class EXPRESSION is silent, only a declaration is listened for",
			"declare const React: any;\nconst C = class extends React.Component { render() { return null; } };\n",
			nil,
		},
		{
			"a shouldComponentUpdate class PROPERTY does not exempt",
			"declare const React: any;\nclass C extends React.Component { shouldComponentUpdate = () => true; render() { return null; } }\n",
			[]string{"noShouldComponentUpdate"},
		},
		{
			"a computed shouldComponentUpdate key does not exempt",
			"declare const React: any;\nclass C extends React.Component { ['shouldComponentUpdate']() { return true; } render() { return null; } }\n",
			[]string{"noShouldComponentUpdate"},
		},
		{
			"a STATIC shouldComponentUpdate does exempt",
			"declare const React: any;\nclass C extends React.Component { static shouldComponentUpdate() { return true; } render() { return null; } }\n",
			nil,
		},
		{
			"a class with no heritage is not a component",
			"class C { render() { return null; } }\n",
			nil,
		},
		{
			"a foreign namespace is not a React base",
			"declare const Foo: any;\nclass C extends Foo.Component { render() { return null; } }\n",
			nil,
		},
		{
			"a foreign PureComponent is not exempt and is not a component either",
			"declare const Foo: any;\nclass C extends Foo.PureComponent { render() { return null; } }\n",
			nil,
		},
		{
			"a bare PureComponent base exempts",
			"declare const PureComponent: any;\nclass C extends PureComponent { render() { return null; } }\n",
			nil,
		},
		{
			"extending a call is not a component",
			"declare const React: any;\ndeclare function mixin(base: any): any;\nclass C extends mixin(React.Component) { render() { return null; } }\n",
			nil,
		},
		{
			"an exported class still reports",
			"declare const React: any;\nexport class C extends React.Component { render() { return null; } }\n",
			[]string{"noShouldComponentUpdate"},
		},
		{
			"two components report twice",
			"declare const React: any;\nclass A extends React.Component { render() { return null; } }\nclass B extends React.Component { render() { return null; } }\n",
			[]string{"noShouldComponentUpdate", "noShouldComponentUpdate"},
		},
		{
			"a nested class reports",
			"declare const React: any;\nfunction f() { class C extends React.Component { render() { return null; } } }\n",
			[]string{"noShouldComponentUpdate"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runRequireOptimization(t, testCase.sourceText, ""), testCase.wantIds...)
		})
	}
}

// TestRequireOptimizationDecoratorArms pins both decorator paths.
//
// The built-in arm compares three separate things and upstream tests all three, so each of the two
// near-miss rows below reports. The configured arm compares the decorator expression's own `name`,
// which only an Identifier has, so `@pure` exempts and `@pure()` does not. Every verdict measured on
// the installed build.
func TestRequireOptimizationDecoratorArms(t *testing.T) {
	t.Parallel()

	const preamble = "declare const reactMixin: any;\ndeclare const PureRenderMixin: any;\ndeclare const SomeOtherMixin: any;\ndeclare const other: any;\ndeclare const Component: any;\ndeclare const pure: any;\ndeclare const bar: any;\n"
	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantIds    []string
	}{
		{"the built-in decoration exempts", preamble + "@reactMixin.decorate(PureRenderMixin)\nclass C extends Component {}\n", "", nil},
		{"a different mixin argument reports", preamble + "@reactMixin.decorate(SomeOtherMixin)\nclass C extends Component {}\n", "", []string{"noShouldComponentUpdate"}},
		{"a different object reports", preamble + "@other.decorate(PureRenderMixin)\nclass C extends Component {}\n", "", []string{"noShouldComponentUpdate"}},
		{"a different property reports", preamble + "@reactMixin.other(PureRenderMixin)\nclass C extends Component {}\n", "", []string{"noShouldComponentUpdate"}},
		{"a listed bare decorator exempts", preamble + "@pure\nclass C extends Component {}\n", `{"allowDecorators":["pure"]}`, nil},
		{"a listed decorator CALLED does not exempt", preamble + "@pure()\nclass C extends Component {}\n", `{"allowDecorators":["pure"]}`, []string{"noShouldComponentUpdate"}},
		{"an unlisted bare decorator reports", preamble + "@bar\nclass C extends Component {}\n", `{"allowDecorators":["pure"]}`, []string{"noShouldComponentUpdate"}},
		{"a bare decorator with no option list reports", preamble + "@pure\nclass C extends Component {}\n", "", []string{"noShouldComponentUpdate"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runRequireOptimization(t, testCase.sourceText, testCase.rawOptions), testCase.wantIds...)
		})
	}
}

// TestRequireOptimizationSpans asserts WHERE each finding points.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule reporting on the class NAME or
// on the heritage clause passes every fixture above while pointing somewhere the reader was never
// shown. Upstream reports on the component node, which is the whole class declaration for the class
// arm and the whole OBJECT for the call arm, and those are two different anchors that no id
// assertion can tell apart.
//
// The text is sliced from the source the harness wrote rather than from the literal, and `Run` does
// not trim, so the two agree here; the slice is taken from the harness output anyway so a later
// switch to the typed harness cannot silently shift it.
func TestRequireOptimizationSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantTexts  []string
	}{
		{
			"the whole class declaration",
			"declare const React: any;\nclass C extends React.Component {}",
			[]string{"class C extends React.Component {}"},
		},
		{
			"an exported class points at the class, not the export",
			"declare const React: any;\nexport class C extends React.Component {}",
			[]string{"class C extends React.Component {}"},
		},
		{
			"the OBJECT, not the call",
			"declare function createReactClass(spec: any): any;\ncreateReactClass({});",
			[]string{"{}"},
		},
		{
			"an object with members still points at the object",
			"declare function createReactClass(spec: any): any;\ndeclare const RandomMixin: any;\ncreateReactClass({ mixins: [RandomMixin] });",
			[]string{"{ mixins: [RandomMixin] }"},
		},
		// Decorators are INCLUDED where `export` is excluded, and the two are the same kind of node
		// here: our parser makes both modifiers on the class, while upstream's puts `export` on a
		// wrapping declaration and keeps decorators on the class itself. Measured on the installed
		// build, the decorated class reports from the first decorator. A mutant neutralizing the
		// decorator branch of the span survived the rest of this suite until these rows existed.
		{
			"a decorated class reports from the DECORATOR",
			"declare const reactMixin: any;\ndeclare const SomeOtherMixin: any;\ndeclare const Component: any;\n@reactMixin.decorate(SomeOtherMixin)\nclass C extends Component {}",
			[]string{"@reactMixin.decorate(SomeOtherMixin)\nclass C extends Component {}"},
		},
		{
			"several decorators report from the FIRST one",
			"declare const bar: any;\ndeclare const pure: any;\ndeclare const foo: any;\ndeclare const Component: any;\n@bar\n@pure\n@foo\nclass C extends Component {}",
			[]string{"@bar\n@pure\n@foo\nclass C extends Component {}"},
		},
		// A decorator written BEFORE `export` is attached to the export declaration upstream, not
		// to the class, so upstream drops it from the span. Measured: it reports `class C extends
		// Component {}` alone. Written the other way round, `export @bar class`, the decorator is
		// on the class and IS included. Two arrangements, two answers, and only measurement
		// separates them; the intuitive reading is that a decorator is a decorator.
		{
			"a decorator BEFORE export is not part of the span",
			"declare const bar: any;\ndeclare const Component: any;\n@bar\nexport class C extends Component {}",
			[]string{"class C extends Component {}"},
		},
		{
			"a decorator AFTER export is part of the span",
			"declare const bar: any;\ndeclare const Component: any;\nexport @bar class C extends Component {}",
			[]string{"@bar class C extends Component {}"},
		},
		// Two more arrangements that separate a backward scan from a forward one, and that pin
		// which keywords upstream's export wrapper actually consumes. Only `export` and `default`
		// move the start; `abstract` stays on the class and stays in the span. Both measured on the
		// installed build, and the second corrected this port: it was dropping `abstract` too.
		{
			"a decorator between export and default is included",
			"declare const bar: any;\ndeclare const Component: any;\nexport default @bar class C extends Component {}",
			[]string{"@bar class C extends Component {}"},
		},
		{
			"abstract stays in the span where export does not",
			"declare const bar: any;\ndeclare const Component: any;\n@bar\nexport abstract class C extends Component {}",
			[]string{"abstract class C extends Component {}"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runRequireOptimization(t, testCase.sourceText, "")
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantTexts))
			}
			source := result.SourceFile.Text()
			for index, diagnostic := range result.Diagnostics {
				got := source[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != testCase.wantTexts[index] {
					t.Errorf("finding %d points at %q, want %q", index, got, testCase.wantTexts[index])
				}
			}
		})
	}
}

// TestRequireOptimizationMessageReadsAsWritten pins the rendered text.
//
// The message interpolates nothing, so there is no format string to get wrong, but the Description
// is what a reader sees and a silent edit is otherwise invisible to every other test here. Asserted
// against a literal typed in this file rather than against the rule's own constant, because
// comparing a diagnostic to the constant it was reported with is an equality whose two sides move
// together under mutation.
func TestRequireOptimizationMessageReadsAsWritten(t *testing.T) {
	t.Parallel()

	result := runRequireOptimization(t, "declare const React: any;\nclass C extends React.Component {}\n", "")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Message.Id != "noShouldComponentUpdate" {
		t.Errorf("id is %q", result.Diagnostics[0].Message.Id)
	}
	if !strings.HasPrefix(result.Diagnostics[0].Message.Description,
		"This class component re-renders whenever its parent does") {
		t.Errorf("description reads %q", result.Diagnostics[0].Message.Description)
	}
}

// TestRequireOptimizationDecodesItsOptions exercises the decoder the config layer calls.
//
// The empty-input row is the one with no upstream counterpart and the reason the decoder is
// hand-written: `rule.DecodeOptionsInto` errors there, and a bare `"error"` configuration hands a
// rule exactly that.
func TestRequireOptimizationDecodesItsOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{"empty input answers the default", "", nil},
		{"an empty object answers the default", "{}", nil},
		{"one decorator", `{"allowDecorators":["pure"]}`, []string{"pure"}},
		{"several decorators", `{"allowDecorators":["renderPure","pureRender"]}`, []string{"renderPure", "pureRender"}},
		{"an explicit empty list", `{"allowDecorators":[]}`, []string{}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeRequireOptimizationOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %q: %v", testCase.raw, err)
			}
			settings, ok := decoded.(RequireOptimizationOptions)
			if !ok {
				t.Fatalf("decoded to %T", decoded)
			}
			if len(settings.AllowDecorators) != len(testCase.want) {
				t.Fatalf("decoded %d decorators, want %d", len(settings.AllowDecorators), len(testCase.want))
			}
			for index, name := range testCase.want {
				if settings.AllowDecorators[index] != name {
					t.Errorf("decorator %d is %q, want %q", index, settings.AllowDecorators[index], name)
				}
			}
		})
	}
}

// TestRequireOptimizationRejectsMalformedOptions asserts the decoder surfaces bad input.
func TestRequireOptimizationRejectsMalformedOptions(t *testing.T) {
	t.Parallel()

	if _, err := DecodeRequireOptimizationOptions([]byte(`{"allowDecorators":`)); err == nil {
		t.Fatal("truncated JSON decoded without error")
	}
	if _, err := DecodeRequireOptimizationOptions([]byte(`"error"`)); err == nil {
		t.Fatal("a string decoded into the options struct without error")
	}
}

// TestRequireOptimizationHandlesNilOptions covers the path the fixtures cannot reach.
//
// The config layer turns a decoder error into nil for a non-required rule, and `options.(T)` on nil
// yields the zero value. Here that IS the default, an empty allow list, so nil must behave exactly
// like an absent option object. Asserted rather than assumed, because a rule whose default was not
// the zero value would be silently inverted on this path.
func TestRequireOptimizationHandlesNilOptions(t *testing.T) {
	t.Parallel()

	reporting := rule_testing.RunWithOptions(t, RequireOptimization, requireOptimizationFile,
		"declare const React: any;\nclass C extends React.Component {}\n", nil)
	rule_testing.ExpectFindings(t, reporting, "noShouldComponentUpdate")

	clean := rule_testing.RunWithOptions(t, RequireOptimization, requireOptimizationFile,
		"declare const React: any;\nclass C extends React.PureComponent {}\n", nil)
	rule_testing.ExpectClean(t, clean)

	// The rows above cannot see the allow list at all, because neither carries a decorator, so a
	// mutant giving the nil path a non-empty list survived them. This is the case that separates
	// "nil means no decorators are allowed" from any other default: a bare decorator must still
	// report when nothing configured it.
	decorated := rule_testing.RunWithOptions(t, RequireOptimization, requireOptimizationFile,
		"declare const pure: any;\ndeclare const Component: any;\n@pure\nclass C extends Component {}\n", nil)
	rule_testing.ExpectFindings(t, decorated, "noShouldComponentUpdate")
}

// TestRequireOptimizationNeedsNoChecker pins that this rule is syntactic.
//
// Every question it asks is on the AST: what the class extends, what members it declares, what
// decorators it carries. The typed harness must therefore produce the identical verdict, and a later
// reader adding `NeedsTypeChecker` has to explain why this stopped being true.
func TestRequireOptimizationNeedsNoChecker(t *testing.T) {
	t.Parallel()

	const source = "declare const React: any;\nclass C extends React.Component {}\n"

	untyped := rule_testing.RunWithOptions(t, RequireOptimization, requireOptimizationFile, source,
		RequireOptimizationOptions{})
	rule_testing.ExpectFindings(t, untyped, "noShouldComponentUpdate")

	typed := rule_testing.RunTypedWithOptions(t, RequireOptimization, requireOptimizationFile, source,
		RequireOptimizationOptions{})
	rule_testing.ExpectFindings(t, typed, "noShouldComponentUpdate")
}

// TestRequireOptimizationHasNoFileGate reports in a `.ts` file, not only in `.tsx`.
//
// Upstream gates on nothing, and a React class in a `.ts` file is ordinary and legal. Three shipped
// rules in this package carried a `.tsx`-only gate that was oxc residue, each with a test asserting
// the gate, so the suite locked the bug in rather than catching it. This pins the opposite.
//
// The source deliberately carries no JSX, because JSX in a `.ts` file is a syntax error and a case
// that failed to parse would produce silence for a reason that has nothing to do with a gate.
func TestRequireOptimizationHasNoFileGate(t *testing.T) {
	t.Parallel()

	const source = "declare const React: any;\nclass C extends React.Component {}\n"
	for _, fileName := range []string{
		"/repository/source/RequireOptimization.tsx",
		"/repository/source/RequireOptimization.ts",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, RequireOptimization, fileName, source,
				RequireOptimizationOptions{})
			rule_testing.ExpectFindings(t, result, "noShouldComponentUpdate")
		})
	}
}

// TestRequireOptimizationSurvivesShapesThatWouldPanic drives the shapes with nothing to read.
//
// The walk recovers per FILE rather than per rule, so one nil dereference here takes the file away
// from every rule in the tree, and no `ExpectFindings` fixture can see a panic. These are ordinary
// code that reaches the two listeners with a missing piece: a call with no arguments, an argument
// that is not an object, a heritage clause that is `implements` rather than `extends`, and callees
// that are bare keywords.
//
// Every verdict here was measured on the installed build rather than assumed, and THREE of the
// original expectations in this test were wrong in the same direction: a spread-only object, a
// `mixins` that is not an array, and an empty `mixins` array all REPORT upstream, because the object
// arm exempts only on a function value, a `shouldComponentUpdate` key, or `PureRenderMixin` in the
// list. They are in the reporting test above now. Writing them here as clean would have been the
// port brief's exact failure: a fixture encoding my belief rather than upstream's behavior, sitting
// in the silent list asserting the opposite of the truth.
func TestRequireOptimizationSurvivesShapesThatWouldPanic(t *testing.T) {
	t.Parallel()

	sources := []string{
		"declare function createReactClass(spec: any): any;\ncreateReactClass();\n",
		"declare function createReactClass(spec: any): any;\ncreateReactClass(undefined);\n",
		"interface Component { x: number }\nclass C implements Component { x = 1; }\n",
		"(()=>{})();\n",
		"class Base {}\nclass Derived extends Base { constructor() { super(); } }\n",
		"async function load() { await import('./other'); }\n",
	}

	for index, sourceText := range sources {
		t.Run(string(rune('a'+index)), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runRequireOptimization(t, sourceText, ""))
		})
	}
}

// TestRequireOptimizationObjectArmReportsWithoutAFunction pins three shapes I got backwards.
//
// Each of these reads as a degenerate object that could not possibly be a component worth reporting
// on, and each of them reports upstream. Measured on the installed build after all three failed as
// clean fixtures. The rule was right and the expectation was wrong, which is the direction worth
// recording: the object arm's exemptions are a closed list, and "this object is obviously not a
// real component" is not on it.
func TestRequireOptimizationObjectArmReportsWithoutAFunction(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a spread-only object", "declare function createReactClass(spec: any): any;\ndeclare const spread: any;\ncreateReactClass({ ...spread });\n"},
		{"a mixins value that is not an array", "declare function createReactClass(spec: any): any;\ncreateReactClass({ mixins: 1 });\n"},
		{"an empty mixins array", "declare function createReactClass(spec: any): any;\ncreateReactClass({ mixins: [] });\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runRequireOptimization(t, testCase.sourceText, ""),
				"noShouldComponentUpdate")
		})
	}
}

// TestRequireOptimizationCreateReactClassSpellings pins which factory names count.
//
// The shelf's `react.IsEs5ComponentCall` accepts `createClass` and `React.createClass` as well, and
// upstream's `require-optimization` accepts NEITHER: its `createClass` pragma defaults to
// `createReactClass` and cohere reads no React settings to change it. A mutant dropping the gate
// entirely survived the rest of this suite, because nothing in the corpus writes a call to anything
// other than `createReactClass`. Every verdict below was measured on the installed build.
func TestRequireOptimizationCreateReactClassSpellings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"createReactClass reports", "declare function createReactClass(spec: any): any;\ncreateReactClass({});\n", []string{"noShouldComponentUpdate"}},
		{"a bare createClass declines", "declare function createClass(spec: any): any;\ncreateClass({});\n", nil},
		{"React.createClass declines", "declare const React: any;\nReact.createClass({});\n", nil},
		{"a foreign namespace declines", "declare const Foo: any;\nFoo.createClass({});\n", nil},
		{"React.createReactClass declines", "declare const React: any;\nReact.createReactClass({});\n", nil},
		{"an unrelated call declines", "declare function notCreateReactClass(spec: any): any;\nnotCreateReactClass({});\n", nil},
		{"a call whose name merely contains it declines", "declare function myCreateReactClass(spec: any): any;\nmyCreateReactClass({});\n", nil},
		{"an ordinary call with an object argument declines", "declare function doThing(spec: any): any;\ndoThing({ a: 1 });\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runRequireOptimization(t, testCase.sourceText, ""), testCase.wantIds...)
		})
	}
}
