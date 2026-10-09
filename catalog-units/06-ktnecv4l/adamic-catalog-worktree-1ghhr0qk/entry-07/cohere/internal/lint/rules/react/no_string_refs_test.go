package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// stringRefsFile is where the fixtures pretend to live.
//
// A `.tsx` name because the corpus is JSX and needs a parser that reads it. It is NOT load-bearing
// the way the previous port's was: this rule has no file gate, and `TestNoStringRefsHasNoFileGate`
// pins that by running the same source under four extensions.
const stringRefsFile = "/repository/source/StringRefs.tsx"

// The corpus is `eslint-plugin-react`'s own, extracted rather than retyped.
//
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/no-string-refs.js` holds 4 valid and 7
// invalid cases carrying 10 diagnostics between them. Every string below was pulled out of that
// file by loading it with a stubbed `RuleTester` and serializing the captured object, then compared
// byte against byte back into the source with a script rather than by eye.
//
// Two columns encode what upstream expresses as test configuration:
//
//	noTemplateLiterals   the case's `options[0].noTemplateLiterals`
//	checkThisRefs        upstream's `settings.react.version` resolved through the version gate,
//	                     `< 18.3.0`. Three of the eleven cases exist only to vary this: upstream's
//	                     fourth valid case and its first invalid case are byte-identical and differ
//	                     only in carrying `18.3.0` against `18.2.0`, and its last invalid case is
//	                     its sixth with the version raised so one of the two findings disappears.
//	                     Absent means `999.999.999`, so false.
//
// That pair is the whole reason this rule grew a `checkThisRefs` option. Without it the corpus
// cannot be expressed: two of its cases are the same bytes with opposite verdicts.
func TestNoStringRefsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name               string
		sourceText         string
		noTemplateLiterals bool
		checkThisRefs      bool
		findings           []string
	}{
		{"upstream invalid 0", "\n        var Hello = createReactClass({\n          componentDidMount: function() {\n            var component = this.refs.hello;\n          },\n          render: function() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        });\n      ", false, true, []string{"thisRefsDeprecated"}},
		{"upstream invalid 1", "\n        var Hello = createReactClass({\n          render: function() {\n            return <div ref=\"hello\">Hello {this.props.name}</div>;\n          }\n        });\n      ", false, true, []string{"stringInRefDeprecated"}},
		{"upstream invalid 2", "\n        var Hello = createReactClass({\n          render: function() {\n            return <div ref={'hello'}>Hello {this.props.name}</div>;\n          }\n        });\n      ", false, true, []string{"stringInRefDeprecated"}},
		{"upstream invalid 3", "\n        var Hello = createReactClass({\n          componentDidMount: function() {\n            var component = this.refs.hello;\n          },\n          render: function() {\n            return <div ref=\"hello\">Hello {this.props.name}</div>;\n          }\n        });\n      ", false, true, []string{"thisRefsDeprecated", "stringInRefDeprecated"}},
		{"upstream invalid 4", "\n        var Hello = createReactClass({\n          componentDidMount: function() {\n          var component = this.refs.hello;\n          },\n          render: function() {\n            return <div ref={`hello`}>Hello {this.props.name}</div>;\n          }\n        });\n      ", true, true, []string{"thisRefsDeprecated", "stringInRefDeprecated"}},
		{"upstream invalid 5", "\n        var Hello = createReactClass({\n          componentDidMount: function() {\n          var component = this.refs.hello;\n          },\n          render: function() {\n            return <div ref={`hello${index}`}>Hello {this.props.name}</div>;\n          }\n        });\n      ", true, true, []string{"thisRefsDeprecated", "stringInRefDeprecated"}},
		{"upstream invalid 6", "\n        var Hello = createReactClass({\n          componentDidMount: function() {\n          var component = this.refs.hello;\n          },\n          render: function() {\n            return <div ref={`hello${index}`}>Hello {this.props.name}</div>;\n          }\n        });\n      ", true, false, []string{"stringInRefDeprecated"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoStringRefs, stringRefsFile, testCase.sourceText,
				NoStringRefsOptions{
					NoTemplateLiterals: testCase.noTemplateLiterals,
					CheckThisRefs:      testCase.checkThisRefs,
				})
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// The clean cases, and each declines for a different reason.
//
// The first passes a ref callback, which is the repair this rule asks for. The second and third
// write template refs with the option off, which is that option's entire subject. The fourth is the
// version gate: byte-identical to the first firing case and clean only because upstream's
// `18.3.0` setting turns the `this.refs` half off.
func TestNoStringRefsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name               string
		sourceText         string
		noTemplateLiterals bool
		checkThisRefs      bool
	}{
		{"upstream valid 0", "\n        var Hello = createReactClass({\n          componentDidMount: function() {\n            var component = this.hello;\n          },\n          render: function() {\n            return <div ref={c => this.hello = c}>Hello {this.props.name}</div>;\n          }\n        });\n      ", false, false},
		{"upstream valid 1", "\n        var Hello = createReactClass({\n          render: function() {\n            return <div ref={`hello`}>Hello {this.props.name}</div>;\n          }\n        });\n      ", false, false},
		{"upstream valid 2", "\n        var Hello = createReactClass({\n          render: function() {\n            return <div ref={`hello${index}`}>Hello {this.props.name}</div>;\n          }\n        });\n      ", false, false},
		{"upstream valid 3", "\n        var Hello = createReactClass({\n          componentDidMount: function() {\n            var component = this.refs.hello;\n          },\n          render: function() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        });\n      ", false, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoStringRefs, stringRefsFile,
				testCase.sourceText, NoStringRefsOptions{
					NoTemplateLiterals: testCase.noTemplateLiterals,
					CheckThisRefs:      testCase.checkThisRefs,
				}))
		})
	}
}

// Cases upstream does not ship, each covering a discrimination its corpus leaves untested.
//
// Every shape here was run against the installed `eslint-plugin-react` build, version 7.37.5,
// through the ESLint Linter API before being written down, so these assert measured upstream
// behavior rather than a reading of the JavaScript. The reading was wrong about five of them.
func TestNoStringRefsDiscriminationsUpstreamDoesNotCover(t *testing.T) {
	t.Parallel()

	fires := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		// A ref attribute needs no enclosing component, unlike the other half of this rule. This is
		// the asymmetry the rule doc describes, and upstream writes every ref fixture inside a
		// component so nothing in the imported corpus pins it. Measured reporting.
		{
			"a string ref outside any component",
			"const a = <div ref=\"hello\" />;\n",
			[]string{"stringInRefDeprecated"},
		},
		// A double-quoted string inside the container, which upstream only ever writes
		// single-quoted. Measured reporting.
		{
			"a double-quoted string ref inside a container",
			"const a = <div ref={\"hello\"} />;\n",
			[]string{"stringInRefDeprecated"},
		},
		// A member-expression element name is still an element with a ref attribute. Measured.
		{
			"a string ref on a namespaced component",
			"const a = <Foo.Bar ref=\"hello\" />;\n",
			[]string{"stringInRefDeprecated"},
		},
	}

	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoStringRefs, stringRefsFile, testCase.sourceText),
				testCase.findings...)
		})
	}

	silent := []struct {
		name       string
		sourceText string
	}{
		// JSX attribute names are case-sensitive and `REF` is a different attribute. Measured
		// silent. A rule comparing case-insensitively survives the entire imported corpus, because
		// upstream never writes the attribute any other way.
		{
			"a ref attribute in the wrong case",
			"const a = <div REF=\"hello\" />;\n",
		},
		// A bare attribute has no value at all, and `containsStringLiteral` opens with `!!node.value`.
		{
			"a ref attribute with no value",
			"const a = <div ref />;\n",
		},
		// A numeric Literal is a Literal whose value is not a string, so it falls out of
		// `typeof node.value.value === 'string'` rather than out of a kind check. Measured silent.
		{
			"a ref holding a number",
			"const a = <div ref={123} />;\n",
		},
		{
			"a ref holding an identifier",
			"const a = <div ref={hello} />;\n",
		},
		// A knowable string that is not a Literal. Upstream does not evaluate, so this is silent
		// even though its value is obvious. Measured.
		{
			"a ref holding a concatenation of two strings",
			"const a = <div ref={\"a\" + \"b\"} />;\n",
		},
		// A namespaced attribute name is a JSXNamespacedName whose `.name` is an object rather than
		// a string, so the comparison against `'ref'` fails on type. Measured silent.
		{
			"a namespaced ref attribute",
			"const a = <svg xlink:ref=\"hello\" />;\n",
		},
		// An attribute whose name merely starts with `ref`.
		{
			"an attribute whose name only starts with ref",
			"const a = <div refs=\"hello\" />;\n",
		},
		// A spread supplying the ref is not a JSXAttribute and is never visited. Measured silent.
		{
			"a ref supplied through a spread",
			"const a = <div {...{ ref: \"hello\" }} />;\n",
		},
	}

	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoStringRefs, stringRefsFile, testCase.sourceText))
		})
	}
}

// How `this.refs` is spelled, which upstream's corpus writes exactly one way.
//
// The corpus writes `this.refs.hello` inside `createReactClass` and nothing else, so every
// discrimination in `isRefsUsage` beyond that one shape is untested upstream. Each case here was
// measured on the installed build with `settings.react.version` set to `18.2.0`, which is what
// `checkThisRefs: true` reproduces.
//
// The computed group is the reason this test exists. `node.property.name` is undefined for a
// computed key unless that key is an identifier, and `computed` is never checked, so the four
// spellings split in a way no reading of the rule name would predict. The previous port of this
// rule had them exactly inverted.
func TestNoStringRefsSpellingsOfTheRefsRead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{
			"the dotted spelling",
			"class H extends React.Component { m() { return this.refs.x; } }\n",
			[]string{"thisRefsDeprecated"},
		},
		{
			"a bare refs read with nothing taken off it",
			"class H extends React.Component { m() { return this.refs; } }\n",
			[]string{"thisRefsDeprecated"},
		},
		// A string-literal key has no `.name`, so upstream is silent. Measured.
		{
			"a computed read written with a string",
			"class H extends React.Component { m() { return this[\"refs\"].x; } }\n",
			nil,
		},
		// A template key likewise. Measured.
		{
			"a computed read written with a bare template",
			"class H extends React.Component { m() { return this[`refs`].x; } }\n",
			nil,
		},
		// An IDENTIFIER key does have `.name`, and it is `refs`, and `computed` is never checked.
		// So an unrelated variable named `refs` is read as the registry and reports. Measured on
		// the installed build. This is an upstream defect, reproduced rather than corrected.
		{
			"a computed read whose key is a variable named refs",
			"class H extends React.Component { m() { return this[refs].x; } }\n",
			[]string{"thisRefsDeprecated"},
		},
		// Parentheses are transparent in ESTree, which has no node for them. Measured reporting;
		// the previous port was silent here.
		{
			"a parenthesized this",
			"class H extends React.Component { m() { return (this).refs.x; } }\n",
			[]string{"thisRefsDeprecated"},
		},
		// The object must be `this` specifically. Measured silent.
		{
			"a refs read off something that is not this",
			"class H extends React.Component { m() { return that.refs.x; } }\n",
			nil,
		},
		// The same shape one level in: a nested object whose property happens to be named `refs`.
		{
			"a refs property on a plain object",
			"class H extends React.Component { m() { return this.props.refs.x; } }\n",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoStringRefs, stringRefsFile, testCase.sourceText,
				NoStringRefsOptions{CheckThisRefs: true})
			if len(testCase.findings) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// Which enclosing component counts, which is a SCOPE walk and not a parent-chain walk.
//
// This is the group that the previous port got wrong in two directions, and neither was visible
// from the imported corpus, which writes exactly one nesting shape. `getParentES6Component` stops
// at the first class scope rather than continuing upward, and `getParentES5Component` walks
// function scopes asking whether each one's block sits inside a factory call.
//
// Every case measured on the installed build at `settings.react.version` of `18.2.0`.
func TestNoStringRefsEnclosingComponentIsAScopeWalk(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		reports    bool
	}{
		// The ES6 walk stops at `Inner`, which is not a component, and returns null rather than
		// continuing up to `H`. Measured silent. A parent-chain walk reports here, which is the
		// false positive the previous port shipped.
		{
			"a plain class nested inside a component hides it",
			"class H extends React.Component { m() { class Inner { k() { return this.refs.x; } } } }\n",
			false,
		},
		// The mirror image: the first class scope IS the component, so the plain class above it
		// does not matter. Measured reporting.
		{
			"a component nested inside a plain class still counts",
			"class Outer { m() { class H extends React.Component { k() { return this.refs.x; } } } }\n",
			true,
		},
		// An arrow opens no class scope, so the walk passes through it to the component.
		{
			"an arrow function inside a component method",
			"class H extends React.Component { m() { const f = () => this.refs.x; } }\n",
			true,
		},
		{
			"a static method",
			"class H extends React.Component { static m() { return this.refs.x; } }\n",
			true,
		},
		{
			"a class field initializer",
			"class H extends React.Component { f = this.refs.x; }\n",
			true,
		},
		{
			"a class expression assigned to a variable",
			"var H = class extends React.Component { m() { return this.refs.x; } };\n",
			true,
		},
		// The ES5 walk reads `scope.block.parent.parent` at each scope. With no function between
		// the read and the module there is no such scope, so this is silent even though the read
		// sits lexically inside the factory call. Measured.
		{
			"an ES5 property value with no function around it",
			"var H = createReactClass({ m: this.refs.x });\n",
			false,
		},
		// A function nested arbitrarily deep inside the factory argument still answers, because the
		// ES5 walk continues upward rather than stopping at the first scope.
		{
			"an ES5 factory with a nested plain object holding the function",
			"var H = createReactClass({ m: function() { var o = { k: function() { return this.refs.x; } }; } });\n",
			true,
		},
		// The four cases below were found by mutation: deleting `KindArrowFunction` and deleting
		// `KindMethodDeclaration` from `opensFunctionScope` both survived every fixture above,
		// because upstream's corpus writes the factory property exactly one way, as a function
		// expression. eslint-scope opens a scope for each of these, so each reaches the factory
		// test and each reports. All four measured on the installed build before being written.
		{
			"an ES5 factory property written as an arrow",
			"var H = createReactClass({ m: () => this.refs.x });\n",
			true,
		},
		{
			"an ES5 factory property written as a shorthand method",
			"var H = createReactClass({ m() { return this.refs.x; } });\n",
			true,
		},
		{
			"an ES5 factory property written as a getter",
			"var H = createReactClass({ get m() { return this.refs.x; } });\n",
			true,
		},
		// An arrow nested inside the function property. The arrow's own hops land on nothing, and
		// the walk continues to the function above it, which is what makes this a case about the
		// loop continuing rather than about the arrow.
		{
			"an arrow nested inside an ES5 factory property",
			"var H = createReactClass({ m: function() { return (() => this.refs.x)(); } });\n",
			true,
		},
		// The factory name is the default `createClass` pragma and nothing wider. The shelf's
		// `IsEs5ComponentCall` accepts the two below and would report on both. Measured silent.
		{
			"the React.createClass factory spelling",
			"var H = React.createClass({ m: function() { return this.refs.x; } });\n",
			false,
		},
		{
			"the bare createClass factory spelling",
			"var H = createClass({ m: function() { return this.refs.x; } });\n",
			false,
		},
		// The namespaced form of the pragma does count. Measured reporting.
		{
			"the React.createReactClass factory spelling",
			"var H = React.createReactClass({ m: function() { return this.refs.x; } });\n",
			true,
		},
		// Parentheses on the callee are transparent in ESTree. Measured reporting.
		{
			"a parenthesized factory callee",
			"var H = (createReactClass)({ m: function() { return this.refs.x; } });\n",
			true,
		},
		// The ES6 base names are `Component` and `PureComponent`, bare or namespaced.
		{
			"a bare Component base",
			"class H extends Component { m() { return this.refs.x; } }\n",
			true,
		},
		{
			"a PureComponent base",
			"class H extends React.PureComponent { m() { return this.refs.x; } }\n",
			true,
		},
		{
			"an unrelated base class",
			"class H extends React.Foo { m() { return this.refs.x; } }\n",
			false,
		},
		{
			"a class with no heritage at all",
			"class H { m() { return this.refs.x; } }\n",
			false,
		},
		// Upstream's first valid case in spirit: no component of any kind encloses the read.
		{
			"a plain function, which is no component",
			"var Hello = function() { return this.refs; };\n",
			false,
		},
		// A component sitting beside the read rather than around it.
		{
			"a read outside the component beside it",
			"var Hello = function() { return this.refs; };\n" +
				"class Other extends React.Component { render() { let x; } }\n",
			false,
		},
		// Upstream's `isExplicitComponent` reads an `@extends` JSDoc tag through doctrine, behind
		// the `componentDetection` option most configurations leave off. Measured silent on the
		// installed build rather than assumed, and not implemented here. Recorded so the absence
		// reads as measured rather than as forgotten.
		{
			"a JSDoc extends tag naming React.Component",
			"/** @extends React.Component */ class H extends Foo { m() { return this.refs.x; } }\n",
			false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoStringRefs, stringRefsFile, testCase.sourceText,
				NoStringRefsOptions{CheckThisRefs: true})
			if testCase.reports {
				rule_testing.ExpectFindings(t, result, "thisRefsDeprecated")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Upstream has no file gate, and the previous port of this rule had one.
//
// oxc declares `should_run` on `source_type().is_jsx()`; `eslint-plugin-react` declares nothing of
// the kind, and the same class reports identically under `probe.jsx`, `probe.js`, `probe.cjs` and
// `probe.mjs` on the installed build. Reproducing oxc's gate costs every finding in every `.ts`
// file, which is most of this tree.
//
// The `this.refs` half rather than the attribute half, deliberately: a `.ts` file cannot hold a JSX
// attribute at all, so gating on a ref attribute would pass whether or not a gate exists.
func TestNoStringRefsHasNoFileGate(t *testing.T) {
	t.Parallel()

	sourceText := "class Hello extends React.Component {\n" +
		"  componentDidMount() {\n" +
		"    var component = this.refs.hello;\n" +
		"  }\n" +
		"}\n"

	for _, fileName := range []string{
		"/repository/source/Hello.tsx",
		"/repository/source/Hello.ts",
		"/repository/source/Hello.jsx",
		"/repository/source/Hello.js",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunWithOptions(t, NoStringRefs, fileName, sourceText,
					NoStringRefsOptions{CheckThisRefs: true}),
				"thisRefsDeprecated")
		})
	}
}

// The option decoder, routed through the rule's own exported entry point.
//
// Building the options struct directly leaves the decoder untested, and the decoder is where the
// one non-upstream key lives. `checkThisRefs` has to be distinguishable as absent, so a struct
// literal cannot exercise the branch that decides what absence means.
func TestDecodeNoStringRefsOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want NoStringRefsOptions
	}{
		{"empty input is upstream's unconfigured answer", "", NoStringRefsOptions{}},
		{"an empty object is the same", "{}", NoStringRefsOptions{}},
		{"the template option alone", `{"noTemplateLiterals":true}`,
			NoStringRefsOptions{NoTemplateLiterals: true}},
		{"the refs half turned on", `{"checkThisRefs":true}`,
			NoStringRefsOptions{CheckThisRefs: true}},
		{"an explicit false is still off", `{"checkThisRefs":false}`, NoStringRefsOptions{}},
		{"both keys together", `{"noTemplateLiterals":true,"checkThisRefs":true}`,
			NoStringRefsOptions{NoTemplateLiterals: true, CheckThisRefs: true}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoStringRefsOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %q: %v", testCase.raw, err)
			}
			got, ok := decoded.(NoStringRefsOptions)
			if !ok {
				t.Fatalf("decoded %q to %T", testCase.raw, decoded)
			}
			if got != testCase.want {
				t.Errorf("decoded %q to %+v, want %+v", testCase.raw, got, testCase.want)
			}
		})
	}

	if _, err := DecodeNoStringRefsOptions([]byte("not json")); err == nil {
		t.Error("decoding malformed input returned no error")
	}
}

// A rule handed nil options must reach the same answer as one handed an empty object.
//
// The brief's measured failure here is a rule reporting registrations while finding nothing,
// because a bare `"error"` in the config hands the rule nil and every fixture reached it through
// the decoder. This bypasses the decoder entirely, which is the only way to see that.
func TestNoStringRefsWithNilOptions(t *testing.T) {
	t.Parallel()

	// The template half is off, so a template ref is clean.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoStringRefs, stringRefsFile,
		"const a = <div ref={`hello`} />;\n"))
	// The refs half is off, so a component's `this.refs` is clean.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoStringRefs, stringRefsFile,
		"class H extends React.Component { m() { return this.refs.x; } }\n"))
	// The attribute half is on regardless, so a string ref still reports. Without this the two
	// assertions above would pass on a rule that registered no listeners at all.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoStringRefs, stringRefsFile,
		"const a = <div ref=\"hello\" />;\n"), "stringInRefDeprecated")
}

// Where each finding points, which no message-id assertion can see.
//
// Both spans are the node upstream hands to `report`, and both were read off the installed build
// rather than inferred. The `this.refs` finding underlines the member expression and stops there,
// so `this.refs.hello` reports against `this.refs` and the outer access is left alone. The
// attribute finding underlines the whole `ref="hello"` rather than the name span that
// `no-children-prop` in this same package reports, so the two rules here disagree about where a JSX
// attribute finding belongs and that disagreement is upstream's rather than ours.
func TestNoStringRefsPointsAtTheRightNode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		reported   []string
	}{
		{
			"the refs read, not the property taken off it",
			"class Hello extends React.Component {\n  m() { return this.refs.hello; }\n}\n",
			[]string{"this.refs"},
		},
		{
			"a bare refs read with nothing taken off it",
			"class Hello extends React.Component {\n  m() { return this.refs; }\n}\n",
			[]string{"this.refs"},
		},
		// The parenthesis is inside the reported range upstream, measured, because ESTree's member
		// expression starts at the open paren.
		{
			"a parenthesized this, parenthesis included",
			"class Hello extends React.Component {\n  m() { return (this).refs.hello; }\n}\n",
			[]string{"(this).refs"},
		},
		{
			"the whole attribute, not its name",
			"const a = <div ref=\"hello\" />;\n",
			[]string{"ref=\"hello\""},
		},
		{
			"the whole attribute when the value is in a container",
			"const a = <div ref={'hello'} />;\n",
			[]string{"ref={'hello'}"},
		},
		{
			"both halves in one component, in source order",
			"class Hello extends React.Component {\n" +
				"  m() { return this.refs.hello; }\n" +
				"  render() { return <div ref=\"hello\" />; }\n" +
				"}\n",
			[]string{"this.refs", "ref=\"hello\""},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoStringRefs, stringRefsFile, testCase.sourceText,
				NoStringRefsOptions{CheckThisRefs: true})
			if len(result.Diagnostics) != len(testCase.reported) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.reported))
			}
			for i, want := range testCase.reported {
				got := testCase.sourceText[result.Diagnostics[i].Range.Pos():result.Diagnostics[i].Range.End()]
				if got != want {
					t.Errorf("finding %d underlines %q, want %q", i, got, want)
				}
			}
		})
	}
}

// The two messages, asserted by equality rather than by containment.
//
// A `strings.Contains` check on a message that interpolates nothing still passes when the message
// is wrong in a way that only adds text, so the ids are asserted against literals typed here rather
// than against the rule's own constants, which would move together under mutation. The two
// descriptions also have to stay distinguishable, since this rule's whole structure rests on the
// two judgments being separately named.
func TestNoStringRefsMessagesReadCorrectly(t *testing.T) {
	t.Parallel()

	if messageThisRefsDeprecated.Id != "thisRefsDeprecated" {
		t.Errorf("this.refs message id is %q", messageThisRefsDeprecated.Id)
	}
	if messageStringInRefDeprecated.Id != "stringInRefDeprecated" {
		t.Errorf("string ref message id is %q", messageStringInRefDeprecated.Id)
	}
	if messageThisRefsDeprecated.Description == messageStringInRefDeprecated.Description {
		t.Error("the two judgments render the same description, so a reader cannot tell them apart")
	}
	// The description says why the code is wrong rather than restating the rule name, so it must
	// not simply be the rule name back again.
	for _, message := range []string{
		messageThisRefsDeprecated.Description,
		messageStringInRefDeprecated.Description,
	} {
		if len(message) < 60 || !strings.Contains(message, "React 19") {
			t.Errorf("description does not explain the removal: %q", message)
		}
	}
}

// The wire type accepts exactly upstream's key set plus the one this port adds.
//
// `meta.schema` declares `additionalProperties: false`, which ESLint enforces by refusing the whole
// configuration by name, and so does this now (#4a4yse4); an unknown key used to be ignored.
func TestNoStringRefsRefusesUnknownOptionKeys(t *testing.T) {
	t.Parallel()

	decoded, err := DecodeNoStringRefsOptions([]byte(`{"noTemplateLiterals":true}`))
	if err != nil {
		t.Fatalf("baseline: %v", err)
	}
	if got := decoded.(NoStringRefsOptions); got != (NoStringRefsOptions{NoTemplateLiterals: true}) {
		t.Errorf("decoded to %+v", got)
	}
	_, err = DecodeNoStringRefsOptions([]byte(`{"bogus":true,"noTemplateLiterals":true}`))
	if err == nil || !strings.Contains(err.Error(), "bogus") {
		t.Fatalf("want a refusal naming the unknown key, got %v", err)
	}
}
