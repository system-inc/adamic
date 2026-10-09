package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is upstream's, extracted by stubbing its RuleTester and capturing what the test file
// hands it, so no case was retyped and no escape was typed by hand. Every case was replayed
// against the installed build first and all of them reproduced, so the expectations here are
// measured rather than transcribed.
const noSetStateFile = "/repository/source/NoSetState.tsx"

func TestNoSetStateStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{"upstream valid 0", "\n        var Hello = function() {\n          this.setState({})\n        };\n      "},
		{"upstream valid 1", "\n        var Hello = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        });\n      "},
		{"upstream valid 2", "\n        var Hello = createReactClass({\n          componentDidUpdate: function() {\n            someNonMemberFunction(arg);\n            this.someHandler = this.setState;\n          },\n          render: function() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        });\n      "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoSetState, noSetStateFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestNoSetStateFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		wantIds          []string
	}{
		{"upstream invalid 0", "\n        var Hello = createReactClass({\n          componentDidUpdate: function() {\n            this.setState({\n              name: this.props.name.toUpperCase()\n            });\n          },\n          render: function() {\n            return <div>Hello {this.state.name}</div>;\n          }\n        });\n      ", []string{"noSetState"}},
		{"upstream invalid 1", "\n        var Hello = createReactClass({\n          someMethod: function() {\n            this.setState({\n              name: this.props.name.toUpperCase()\n            });\n          },\n          render: function() {\n            return <div onClick={this.someMethod.bind(this)}>Hello {this.state.name}</div>;\n          }\n        });\n      ", []string{"noSetState"}},
		{"upstream invalid 2", "\n        class Hello extends React.Component {\n          someMethod() {\n            this.setState({\n              name: this.props.name.toUpperCase()\n            });\n          }\n          render() {\n            return <div onClick={this.someMethod.bind(this)}>Hello {this.state.name}</div>;\n          }\n        };\n      ", []string{"noSetState"}},
		{"upstream invalid 3", "\n        class Hello extends React.Component {\n          someMethod = () => {\n            this.setState({\n              name: this.props.name.toUpperCase()\n            });\n          }\n          render() {\n            return <div onClick={this.someMethod.bind(this)}>Hello {this.state.name}</div>;\n          }\n        };\n      ", []string{"noSetState"}},
		{"upstream invalid 4", "\n        class Hello extends React.Component {\n          render() {\n            return <div onMouseEnter={() => this.setState({dropdownIndex: index})} />;\n          }\n        };\n      ", []string{"noSetState"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoSetState, noSetStateFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoSetStateMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite is a differential table.
//
// Upstream's corpus is eight cases and writes almost none of the discriminations this rule makes,
// so every row below was run against the installed `eslint-plugin-react` build on 2026-08-27 and
// the `want` column is what that build reported. Four of them decide real behaviour that no
// imported fixture touches:
//
//	the factory spellings     `createReactClass` counts and a bare `createClass` and a namespaced
//	                          `React.createClass` do not, so the shelf's `IsEs5ComponentCall` is too
//	                          wide and this rule reuses `isCreateReactClassCall` instead
//	the class gate is real    a class with no React heritage is silent, unlike the neighbouring
//	                          `no-is-mounted`, which asks syntactic containment and reports on it
//	the property is by name   a computed `this["setState"]({})` is silent, because upstream reads
//	                          `callee.property.name` and a computed member has none
//	parentheses               `(this).setState({})` and `(this.setState)({})` both report, because
//	                          upstream's parser folds the parenthesis away before the rule sees it
//
// The nested-class row is the one worth reading twice: a plain class written inside a component
// still reports, because the ancestor walk finds the enclosing component and never asks whether a
// nearer class interrupted it. That is upstream's behaviour, reproduced.
func TestNoSetStateMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantFindings int
	}{
		{"class no heritage", "class H { m(){ this.setState({}); } render(){ return <div/>; } }\n", 0},
		{"class extends React.Component", "class H extends React.Component { m(){ this.setState({}); } render(){ return <div/>; } }\n", 1},
		{"class extends React.PureComponent", "class H extends React.PureComponent { m(){ this.setState({}); } render(){ return <div/>; } }\n", 1},
		{"class extends bare Component", "import {Component} from 'react';\nclass H extends Component { m(){ this.setState({}); } }\n", 1},
		{"class extends parenthesized", "class H extends (React.Component) { m(){ this.setState({}); } }\n", 1},
		{"class extends Other.Component", "class H extends Other.Component { m(){ this.setState({}); } }\n", 0},
		{"class expression", "const H = class extends React.Component { m(){ this.setState({}); } };\n", 1},
		{"createReactClass", "var H = createReactClass({ m: function(){ this.setState({}); } });\n", 1},
		{"bare createClass", "var H = createClass({ m: function(){ this.setState({}); } });\n", 0},
		{"React.createClass", "var H = React.createClass({ m: function(){ this.setState({}); } });\n", 0},
		{"private name reports, because upstream reads the bare property name", "class H extends React.Component { #setState(){} m(){ this.#setState({}); } }\n", 1},
		{"a differently named private member is silent", "class H extends React.Component { #other(){} m(){ this.#other({}); } }\n", 0},
		{"computed member", "class H extends React.Component { m(){ this[\"setState\"]({}); } }\n", 0},
		{"parenthesized this", "class H extends React.Component { m(){ (this).setState({}); } }\n", 1},
		{"parenthesized callee", "class H extends React.Component { m(){ (this.setState)({}); } }\n", 1},
		{"not this receiver", "class H extends React.Component { m(){ foo.setState({}); } }\n", 0},
		{"reference not call", "class H extends React.Component { m(){ this.x = this.setState; } }\n", 0},
		{"two calls", "class H extends React.Component { m(){ this.setState({}); this.setState({}); } }\n", 2},
		{"nested arrow in render", "class H extends React.Component { render(){ return <div onClick={()=>this.setState({})}/>; } }\n", 1},
		{"class field arrow", "class H extends React.Component { m = () => { this.setState({}); } }\n", 1},
		{"setState outside any component", "function f(){ this.setState({}); }\n", 0},
		{"setState at top level", "this.setState({});\n", 0},
		{"nested class inside component", "class Outer extends React.Component { m(){ class Inner { n(){ this.setState({}); } } } }\n", 1},
		{"optional call this?.setState()", "class H extends React.Component { m(){ this?.setState({}); } }\n", 1},
		{"no args", "class H extends React.Component { m(){ this.setState(); } }\n", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoSetState, noSetStateFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.wantFindings {
				t.Errorf("installed build reports %d findings, this rule reports %d",
					testCase.wantFindings, len(result.Diagnostics))
			}
		})
	}
}

// TestNoSetStateAnchorsOnTheCalleeNotTheCall pins where each finding points.
//
// `ExpectFindings` asserts ids and counts and can see none of this. Upstream reports on
// `callee`, the `this.setState` member access, rather than on the whole call expression, so the
// argument list is outside the span. Every expected span was measured against the installed build
// on 2026-08-27 by slicing its reported range out of the source.
//
// The parenthesized row is the interesting one: upstream's span INCLUDES the parentheses, because
// its parser folded them away and the member access it reports covers that text. Our parser keeps
// the node, and reporting on the member access reproduces the same span for a different reason.
func TestNoSetStateAnchorsOnTheCalleeNotTheCall(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSpans  []string
	}{
		{
			name:       "the callee, not the call, so the argument list is outside the span",
			sourceText: "class H extends React.Component { m(){ this.setState({a:1}); } }\n",
			wantSpans:  []string{"this.setState"},
		},
		{
			name:       "a parenthesized receiver is inside the span",
			sourceText: "class H extends React.Component { m(){ (this).setState({}); } }\n",
			wantSpans:  []string{"(this).setState"},
		},
		{
			name:       "every usage reports, in source order",
			sourceText: "class H extends React.Component { m(){ this.setState({}); this.setState({}); } }\n",
			wantSpans:  []string{"this.setState", "this.setState"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoSetState, noSetStateFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("want %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			for index, diagnostic := range result.Diagnostics {
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.wantSpans[index] {
					t.Errorf("finding %d points at %q, want %q", index, reported, testCase.wantSpans[index])
				}
			}
		})
	}
}

// TestNoSetStateMessage pins the rendered message exactly.
//
// The message has no format verbs, so there is nothing to interpolate, but the text is upstream's
// own string. Asserted against a literal typed here rather than against the rule's own constant, so
// the two cannot move together under mutation.
func TestNoSetStateMessage(t *testing.T) {
	t.Parallel()

	if messageNoSetState.Id != "noSetState" {
		t.Errorf("message id is %q, want %q", messageNoSetState.Id, "noSetState")
	}
	if messageNoSetState.Description != "Do not use setState" {
		t.Errorf("message description is %q, want upstream's own text", messageNoSetState.Description)
	}
}

// TestNoSetStateHasNoFileSuffixGate pins the absence of the gate three siblings in this package
// carry.
//
// Those siblings were ported from oxc, which gates on the file being read as JSX. This rule's
// authority has no filename condition anywhere. The source deliberately holds no JSX, so a `.ts`
// file parses cleanly and the parser's opinion cannot be mistaken for the rule's. The `.ts` row is
// the one that carries the weight: it is exactly the file a `.tsx` gate would have silenced.
func TestNoSetStateHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	sourceText := "class H extends React.Component { m(){ this.setState({}); } }\n"
	for _, fileName := range []string{
		"/repository/source/Suffix.tsx",
		"/repository/source/Suffix.jsx",
		"/repository/source/Suffix.ts",
		"/repository/source/Suffix.js",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoSetState, fileName, sourceText)
			rule_testing.ExpectFindings(t, result, "noSetState")
		})
	}
}
