package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is upstream's, extracted by stubbing its RuleTester and capturing the case objects the
// test file hands it, so no case was retyped.
//
// One mechanical substitution was applied and it is a correction rather than an adaptation.
// Upstream writes six of its nine failing cases with `React.createClass`, which the version this
// repository runs does NOT recognise: `componentUtil` reads the factory name from
// `settings.react.createClass`, whose default is `createReactClass`, and upstream's own test
// harness never sets it. Measured against the installed build, all six are SILENT as written and
// report on byte-identical bodies once the call is spelled `createReactClass`. So `React.createClass(`
// was rewritten to `createReactClass(` and every one of the nineteen cases was then replayed; all
// nineteen reproduce their upstream verdict exactly. Without the rewrite, six of the nine failing
// cases would have been imported as evidence of behaviour the running rule does not have.
//
// cohere reads no React settings, so `createReactClass` is the only reachable spelling here too,
// which is why the substitution is also the right shape for this port rather than only for the
// measurement.
const noAccessStateInSetstateFile = "/repository/source/NoAccessStateInSetstate.tsx"

func TestNoAccessStateInSetstateStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{"upstream valid 0", "\n        var Hello = createReactClass({\n          onClick: function() {\n            this.setState(state => ({value: state.value + 1}))\n          }\n        });\n      "},
		{"upstream valid 1", "\n        var Hello = createReactClass({\n          multiplyValue: function(obj) {\n            return obj.value*2\n          },\n          onClick: function() {\n            var value = this.state.value\n            this.multiplyValue({ value: value })\n          }\n        });\n      "},
		{"upstream valid 2", "\n        var SearchForm = createReactClass({\n          render: function () {\n            return (\n              <div>\n                {(function () {\n                  if (this.state.prompt) {\n                    return <div>{this.state.prompt}</div>\n                  }\n                }).call(this)}\n              </div>\n            );\n          }\n        });\n      "},
		{"upstream valid 3", "\n        var Hello = createReactClass({\n          onClick: function() {\n            this.setState({}, () => console.log(this.state));\n          }\n        });\n      "},
		{"upstream valid 4", "\n        var Hello = createReactClass({\n          onClick: function() {\n            this.setState({}, () => 1 + 1);\n          }\n        });\n      "},
		{"upstream valid 5", "\n        var Hello = createReactClass({\n          onClick: function() {\n            var nextValueNotUsed = this.state.value + 1\n            var nextValue = 2\n            this.setState({value: nextValue})\n          }\n        });\n      "},
		{"upstream valid 6", "\n        function testFunction({a, b}) {\n        };\n      "},
		{"upstream valid 7", "\n        class ComponentA extends React.Component {\n          state = {\n            greeting: 'hello',\n          };\n\n          myFunc = () => {\n            this.setState({ greeting: 'hi' }, () => this.doStuff());\n          };\n\n          doStuff = () => {\n            console.log(this.state.greeting);\n          };\n        }\n      "},
		{"upstream valid 8", "\n        class Foo extends Abstract {\n          update = () => {\n            const result = this.getResult ( this.state.foo );\n            return this.setState ({ result });\n          };\n        }\n      "},
		{"upstream valid 9", "\n        class StateContainer extends Container {\n          anything() {\n            return this.setState({value: this.state.value + 1})\n          }\n        };\n      "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestNoAccessStateInSetstateFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		messageIds       []string
	}{
		{"upstream invalid 0", "\n        var Hello = createReactClass({\n          onClick: function() {\n            this.setState({value: this.state.value + 1})\n          }\n        });\n      ", []string{"useCallback"}},
		{"upstream invalid 1", "\n        var Hello = createReactClass({\n          onClick: function() {\n            this.setState(() => ({value: this.state.value + 1}))\n          }\n        });\n      ", []string{"useCallback"}},
		{"upstream invalid 2", "\n        var Hello = createReactClass({\n          onClick: function() {\n            var nextValue = this.state.value + 1\n            this.setState({value: nextValue})\n          }\n        });\n      ", []string{"useCallback"}},
		{"upstream invalid 3", "\n        var Hello = createReactClass({\n          onClick: function() {\n            var {state, ...rest} = this\n            this.setState({value: state.value + 1})\n          }\n        });\n      ", []string{"useCallback"}},
		{"upstream invalid 4", "\n        function nextState(state) {\n          return {value: state.value + 1}\n        }\n        var Hello = createReactClass({\n          onClick: function() {\n            this.setState(nextState(this.state))\n          }\n        });\n      ", []string{"useCallback"}},
		{"upstream invalid 5", "\n        var Hello = createReactClass({\n          onClick: function() {\n            this.setState(this.state, () => 1 + 1);\n          }\n        });\n      ", []string{"useCallback"}},
		{"upstream invalid 6", "\n        var Hello = createReactClass({\n          onClick: function() {\n            this.setState(this.state, () => console.log(this.state));\n          }\n        });\n      ", []string{"useCallback"}},
		{"upstream invalid 7", "\n        var Hello = createReactClass({\n          nextState: function() {\n            return {value: this.state.value + 1}\n          },\n          onClick: function() {\n            this.setState(nextState())\n          }\n        });\n      ", []string{"useCallback"}},
		{"upstream invalid 8", "\n        class Hello extends React.Component {\n          onClick() {\n            this.setState(this.state, () => console.log(this.state));\n          }\n        }\n      ", []string{"useCallback"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestNoAccessStateInSetstateSpans asserts where each finding points.
//
// The finding anchors on the `this.state` access rather than on the `setState` call, measured on
// the installed build at columns 27 through 37 for `this.state` inside a one-line body. A
// message-id assertion cannot see this, and the three routes to a finding anchor on three
// different recorded nodes, so an anchor mistake on the variable or method route would be invisible
// from the direct one.
func TestNoAccessStateInSetstateSpans(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText, reported string }{
		{
			"the direct route points at the state access",
			"class Hello extends React.Component {\n  onClick() {\n    this.setState({value: this.state.value + 1});\n  }\n}",
			"this.state",
		},
		{
			"the variable route points at the state access, not the use",
			"var Hello = createReactClass({\n  onClick: function() {\n    var v = this.state.value;\n    this.setState({value: v});\n  }\n});",
			"this.state",
		},
		{
			"the object pattern route points at the destructured key",
			"var Hello = createReactClass({\n  onClick: function() {\n    var {state, ...rest} = this;\n    this.setState({value: state.value});\n  }\n});",
			"state",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
				testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.reported {
				t.Fatalf("finding points at %q, wanted %q", reported, testCase.reported)
			}
		})
	}
}

// TestNoAccessStateInSetstateMessageText asserts the rendered text exactly.
//
// Asserted against a literal typed here rather than against the rule's own constant: comparing a
// finding to the constant it was reported with moves both sides together under mutation and passes
// either way.
func TestNoAccessStateInSetstateMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
		"class Hello extends React.Component {\n  onClick() {\n    this.setState({value: this.state.value + 1});\n  }\n}")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	want := "Use callback in setState when referencing the previous state. React batches state " +
		"updates, so `this.state` read while building the next state is whatever was there before " +
		"the pending updates applied, and two updates in one batch will read the same stale value. " +
		"The callback form is handed the state React is about to apply."
	if result.Diagnostics[0].Message.Description != want {
		t.Fatalf("message is %q", result.Diagnostics[0].Message.Description)
	}
	if result.Diagnostics[0].Message.Id != "useCallback" {
		t.Fatalf("id is %q", result.Diagnostics[0].Message.Id)
	}
	// This rule proposes no repair, and upstream ships none either.
	if len(result.Diagnostics[0].Fixes) != 0 || len(result.Diagnostics[0].Suggestions) != 0 {
		t.Fatal("this rule must propose no repair")
	}
}

// TestNoAccessStateInSetstateShapesUpstreamDoesNotWrite covers inputs measured against the
// installed build on 2026-08-27 that upstream's corpus has no case for.
//
// The two-findings rows matter most. Upstream's longest failing case reports once, so nothing in
// the imported corpus can tell a rule that reports per access from one that reports per call, and
// the method route reporting TWICE AT THE SAME SPAN is bookkeeping a reader would reasonably
// "correct" into deduplication.
func TestNoAccessStateInSetstateShapesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		messageIds       []string
	}{
		{
			"a receiver other than this is not setState",
			"var Hello = createReactClass({\n  onClick: function() {\n    that.setState({value: this.state.value + 1});\n  }\n});",
			nil,
		},
		{
			"a plain object literal is not a component",
			"var Hello = {\n  onClick: function() {\n    this.setState({value: this.state.value + 1});\n  }\n};",
			nil,
		},
		{
			"a class extending something other than React is not a component",
			"class Hello extends Container {\n  onClick() {\n    this.setState({value: this.state.value + 1});\n  }\n}",
			nil,
		},
		{
			"PureComponent is a component",
			"class Hello extends React.PureComponent {\n  onClick() {\n    this.setState({value: this.state.value + 1});\n  }\n}",
			[]string{"useCallback"},
		},
		{
			"a bare Component base is a component",
			"class Hello extends Component {\n  onClick() {\n    this.setState({value: this.state.value + 1});\n  }\n}",
			[]string{"useCallback"},
		},
		{
			"an element access spelling of state is not matched",
			"class Hello extends React.Component {\n  onClick() {\n    this.setState({value: this[\"state\"].value + 1});\n  }\n}",
			nil,
		},
		{
			"setState with no arguments cannot have a first argument",
			"class Hello extends React.Component {\n  onClick() {\n    this.setState();\n    var v = this.state.value;\n  }\n}",
			nil,
		},
		{
			"two state accesses in one call report twice",
			"class Hello extends React.Component {\n  onClick() {\n    this.setState({a: this.state.a, b: this.state.b});\n  }\n}",
			[]string{"useCallback", "useCallback"},
		},
		{
			"a renamed object pattern key records a name nothing later uses",
			"var Hello = createReactClass({\n  onClick: function() {\n    var {state: s} = this;\n    this.setState({value: s.value});\n  }\n});",
			nil,
		},
		{
			"a marked method called from two setState calls reports twice at one span",
			"var Hello = createReactClass({\n  nextState: function() {\n    return {value: this.state.value + 1};\n  },\n  onClick: function() {\n    this.setState(nextState());\n    this.setState(nextState());\n  }\n});",
			[]string{"useCallback", "useCallback"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
				testCase.sourceText)
			if len(testCase.messageIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestNoAccessStateInSetstateScopeIdentity pins the variable route's scope comparison.
//
// Upstream compares scope objects by reference, and the consequence is stricter than it reads: a
// `var` moved into a block goes silent even though the name still resolves, because the block is a
// different scope object. All three rows measured against the installed build.
//
// This is the discrimination most likely to be "fixed" by a later reader into name resolution,
// which would report code upstream calls clean.
func TestNoAccessStateInSetstateScopeIdentity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		messageIds       []string
	}{
		{
			"same scope reports",
			"var Hello = createReactClass({\n  onClick: function() {\n    var v = this.state.value;\n    this.setState({value: v});\n  }\n});",
			[]string{"useCallback"},
		},
		{
			"a use inside a nested function is a different scope and is silent",
			"var Hello = createReactClass({\n  onClick: function() {\n    var v = this.state.value;\n    this.setState(function(){ return {value: v}; });\n  }\n});",
			nil,
		},
		{
			"a var declared inside a block is a different scope even though it hoists",
			"var Hello = createReactClass({\n  onClick: function() {\n    { var v = this.state.value; }\n    this.setState({value: v});\n  }\n});",
			nil,
		},
		{
			"a let declared and used inside one block is the same scope",
			"var Hello = createReactClass({\n  onClick: function() {\n    { let v = this.state.value; this.setState({value: v}); }\n  }\n});",
			[]string{"useCallback"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
				testCase.sourceText)
			if len(testCase.messageIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestNoAccessStateInSetstateFirstArgumentOnly pins the argument-position split.
//
// The completion callback is the correct place to read state, so the same access is a finding in
// position one and clean in position two. This pair is the rule's central discrimination and both
// halves are in upstream's corpus; they are restated here as an adjacent pair because reading them
// side by side is what makes the rule's judgment legible.
func TestNoAccessStateInSetstateFirstArgumentOnly(t *testing.T) {
	t.Parallel()

	clean := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
		"var Hello = createReactClass({\n  onClick: function() {\n    this.setState({}, () => console.log(this.state));\n  }\n});")
	rule_testing.ExpectClean(t, clean)

	reports := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
		"var Hello = createReactClass({\n  onClick: function() {\n    this.setState(this.state, () => console.log(this.state));\n  }\n});")
	rule_testing.ExpectFindings(t, reports, "useCallback")
}

// TestNoAccessStateInSetstateOnlyStateIsMatched pins the property name.
//
// A mutation making the property test always-accept survived every other fixture in this file,
// because nothing here writes a `this.<something else>` access inside a `setState` argument.
// Measured against the installed build: `this.props` is silent, so the name comparison is
// load-bearing rather than decorative and this rule is about state specifically.
func TestNoAccessStateInSetstateOnlyStateIsMatched(t *testing.T) {
	t.Parallel()

	silent := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
		"class Hello extends React.Component {\n  onClick() {\n    this.setState({value: this.props.value + 1});\n  }\n}")
	rule_testing.ExpectClean(t, silent)

	reports := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
		"class Hello extends React.Component {\n  onClick() {\n    this.setState({value: this.state.value + 1});\n  }\n}")
	rule_testing.ExpectFindings(t, reports, "useCallback")
}

// TestNoAccessStateInSetstateIdentifierPosition pins where a recorded variable has to appear.
//
// Upstream requires the identifier to be a property's VALUE or a member access's OBJECT. A mutation
// removing that gate survived every other fixture here, and the distinguishing inputs are both
// counterintuitive enough that a reader would call them bugs:
//
//	this.setState(v)      SILENT, even though v holds this.state
//	this.setState({v: 1}) SILENT, v is a key rather than a value
//
// Both measured against the installed build. The first is the one worth pausing on: handing the
// state variable straight to `setState` is exactly the mistake this rule exists to catch, and
// upstream misses it because its identifier arm never looks at a bare argument. Reproduced rather
// than improved on, because widening it changes which files report.
func TestNoAccessStateInSetstateIdentifierPosition(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		messageIds       []string
	}{
		{
			"a bare identifier as the whole first argument is missed",
			"var Hello = createReactClass({\n  onClick: function() {\n    var v = this.state.value;\n    this.setState(v);\n  }\n});",
			nil,
		},
		{
			"a property key sharing the name is not a use",
			"var Hello = createReactClass({\n  onClick: function() {\n    var v = this.state.value;\n    this.setState({v: 1});\n  }\n});",
			nil,
		},
		{
			"a property value is a use",
			"var Hello = createReactClass({\n  onClick: function() {\n    var v = this.state.value;\n    this.setState({value: v});\n  }\n});",
			[]string{"useCallback"},
		},
		{
			"a member access object is a use",
			"var Hello = createReactClass({\n  onClick: function() {\n    var v = this.state;\n    this.setState({value: v.value});\n  }\n});",
			[]string{"useCallback"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
				testCase.sourceText)
			if len(testCase.messageIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestNoAccessStateInSetstateBinaryExpressionClimb pins the walk past enclosing binary expressions.
//
// Without it, `{value: v + 1}` puts the identifier's parent at a binary expression rather than at
// the property, so the position gate declines and the finding is lost. Upstream climbs in a LOOP
// rather than one step, which the nested row is what distinguishes: `v + 1 + 2` nests two deep and
// a single step would still land on a binary expression.
//
// Both measured reporting against the installed build. Upstream's own corpus writes `v + 1` only
// through the `this.state.value + 1` direct route, where the climb is not consulted.
func TestNoAccessStateInSetstateBinaryExpressionClimb(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{
			"one enclosing binary expression",
			"var Hello = createReactClass({\n  onClick: function() {\n    var v = this.state.value;\n    this.setState({value: v + 1});\n  }\n});",
		},
		{
			"two nested binary expressions need the loop rather than one step",
			"var Hello = createReactClass({\n  onClick: function() {\n    var v = this.state.value;\n    this.setState({value: v + 1 + 2});\n  }\n});",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
				testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "useCallback")
		})
	}
}

// TestNoAccessStateInSetstateMethodRoute pins the method route across ten shapes, all measured
// against the installed build on 2026-08-27.
//
// This test exists because the port was WRONG on two of them and no imported case could see it.
// Upstream's corpus writes the method route exactly once, in the object-literal one-hop spelling,
// which happens to be the shape where the divergence does not show.
//
// The asymmetry is upstream's and it is easy to miss reading the source. Its propagation walk tests
// `current.type === 'MethodDefinition'` alone, while its `this.state` walk tests MethodDefinition
// OR a FunctionExpression under a property. So a two-hop chain propagates in a class and not in an
// object literal.
//
// The second half is a PARSER difference rather than an upstream one. ESTree reserves
// MethodDefinition for a class member and spells an object method shorthand as a Property holding a
// FunctionExpression; our parser gives both the same kind. So the propagation arm needs an explicit
// class-parent check, which nothing in the ESTree-shaped source suggests. Without it this port
// reported the object shorthand two-hop chain, which upstream calls clean.
//
// The class-property rows are a third distinction: upstream's `else if` requires a
// FunctionExpression, so a property holding an ARROW is silent and one holding a `function` reports.
func TestNoAccessStateInSetstateMethodRoute(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"a class two hop chain propagates",
			"class Hello extends React.Component {\n  inner() { return this.state.value; }\n  outer() { return inner(); }\n  onClick() { this.setState(outer()); }\n}",
			1,
		},
		{
			"a class one hop needs no propagation",
			"class Hello extends React.Component {\n  inner() { return this.state.value; }\n  onClick() { this.setState(inner()); }\n}",
			1,
		},
		{
			"a class expression propagates like a declaration",
			"var Hello = class extends React.Component {\n  inner() { return this.state.value; }\n  outer() { return inner(); }\n  onClick() { this.setState(outer()); }\n};",
			1,
		},
		{
			"an object literal two hop chain does NOT propagate",
			"var Hello = createReactClass({\n  inner: function() {\n    return this.state.value;\n  },\n  outer: function() {\n    return inner();\n  },\n  onClick: function() {\n    this.setState(outer());\n  }\n});",
			0,
		},
		{
			"an object method shorthand two hop chain does NOT propagate either",
			"var Hello = createReactClass({\n  inner() { return this.state.value; },\n  outer() { return inner(); },\n  onClick() { this.setState(outer()); }\n});",
			0,
		},
		{
			"an object method shorthand one hop still reports",
			"var Hello = createReactClass({\n  inner() { return this.state.value; },\n  onClick() { this.setState(inner()); }\n});",
			1,
		},
		{
			"a method defined after the setState call is not yet recorded",
			"class Hello extends React.Component {\n  inner() { return this.state.value; }\n  onClick() { this.setState(outer()); }\n  outer() { return inner(); }\n}",
			0,
		},
		{
			"a class property holding an arrow is not a recorded method",
			"class Hello extends React.Component {\n  inner = () => this.state.value;\n  onClick() { this.setState(inner()); }\n}",
			0,
		},
		{
			"a class property holding a function expression is a recorded method",
			"class Hello extends React.Component {\n  inner = function() { return this.state.value; };\n  onClick() { this.setState(inner()); }\n}",
			1,
		},
		{
			"a computed method name records no usable name",
			"var Hello = createReactClass({\n  [x]() { return this.state.value; },\n  onClick() { this.setState(inner()); }\n});",
			0,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
				testCase.sourceText)
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			ids := make([]string, testCase.findings)
			for index := range ids {
				ids[index] = "useCallback"
			}
			rule_testing.ExpectFindings(t, result, ids...)
		})
	}
}

// TestNoAccessStateInSetstateObjectPatternGates pins the two conditions on the destructuring route.
//
// Both mutants over these lines survived until these rows existed, and both distinguishing inputs
// were measured silent against the installed build.
func TestNoAccessStateInSetstateObjectPatternGates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		messageIds       []string
	}{
		{
			"destructuring something other than this is not the component state",
			"var Hello = createReactClass({\n  onClick: function() {\n    var {state, ...rest} = other;\n    this.setState({value: state.value});\n  }\n});",
			nil,
		},
		{
			"destructuring a key other than state is not the component state",
			"var Hello = createReactClass({\n  onClick: function() {\n    var {props, ...rest} = this;\n    this.setState({value: props.value});\n  }\n});",
			nil,
		},
		{
			"destructuring state off this is the component state",
			"var Hello = createReactClass({\n  onClick: function() {\n    var {state, ...rest} = this;\n    this.setState({value: state.value});\n  }\n});",
			[]string{"useCallback"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAccessStateInSetstate, noAccessStateInSetstateFile,
				testCase.sourceText)
			if len(testCase.messageIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}
