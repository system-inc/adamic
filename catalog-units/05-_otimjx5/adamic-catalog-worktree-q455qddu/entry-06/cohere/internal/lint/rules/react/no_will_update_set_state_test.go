package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// willUpdateSetStateFile is where the fixtures pretend to live.
//
// A `.tsx` name because upstream's corpus is written in JSX and one clean case returns JSX. Unlike
// oxc, this rule declares no file gate; see the note on the rule about `should_run`.
const willUpdateSetStateFile = "/repository/source/WillUpdate.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every upstream case below is verbatim from
// `oxc/crates/oxc_linter/src/rules/react/no_will_update_set_state.rs`: 7 pass and 12 fail, from a
// single tester block. The extractor reported 12 diagnostics against those 12 fail inputs and the
// snapshot carries 12 location headers, so the mapping is one finding per input and no recovery by
// line alignment was needed. There is only one snapshot file for this rule, so the multi-snapshot
// discrepancy hazard the brief warns about could not arise; it was checked rather than assumed.
//
// The strings were decoded from the extractor's `-dump` output as JSON and then compared byte
// against byte into the Rust source by a script before any Go was written. All 19 matched, and a
// second script confirmed none of them contains a backslash or a quote, so the escape-cooking
// hazard that has bitten four previous porters could not arise here.
//
// The option column is the raw upstream spelling: "" for an absent option, and otherwise the single
// positional string upstream passes.
//
// # One upstream pass case is recorded here as reporting, deliberately
//
// Upstream's sixth pass case is an `UNSAFE_componentWillUpdate` under
// `settings.react.version = "16.2.0"`, and it is clean upstream only because that version predates
// the prefix. cohere reads no React settings, so no react version can reach a rule here and oxc's own default
// branch, `is_none_or(supports_unsafe_lifecycle_prefix)`, is the only reachable one. Measured on the
// release binary rather than reasoned about: with no version configured that exact source reports,
// and with `"version": "16.2.0"` configured it is silent, in the same run against the same file.
// So the case is pinned as reporting, which is upstream's answer under the configuration we can
// actually produce, and the reasoning is here so the next reader does not delete the case or break
// the rule to green it.
func TestNoWillUpdateSetStateFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		option     string
	}{
		// Upstream, verbatim.
		{"a direct setState in a createReactClass componentWillUpdate", "\n                    var Hello = createReactClass({\n                      componentWillUpdate: function() {\n                        this.setState({\n                          data: data\n                        });\n                      }\n                    });\n                  ", ""},
		{"a direct setState in a class componentWillUpdate", "\n                    class Hello extends React.Component {\n                      componentWillUpdate() {\n                        this.setState({\n                          data: data\n                        });\n                      }\n                    }\n                  ", ""},
		{"a direct setState in createReactClass under disallow-in-func", "\n                    var Hello = createReactClass({\n                      componentWillUpdate: function() {\n                        this.setState({\n                          data: data\n                        });\n                      }\n                    });\n                  ", "disallow-in-func"},
		{"a direct setState in a class component under disallow-in-func", "\n                    class Hello extends React.Component {\n                      componentWillUpdate() {\n                        this.setState({\n                          data: data\n                        });\n                      }\n                    }\n                  ", "disallow-in-func"},
		{"a function-expression callback in createReactClass under disallow-in-func", "\n                    var Hello = createReactClass({\n                      componentWillUpdate: function() {\n                        someClass.onSomeEvent(function(data) {\n                          this.setState({\n                            data: data\n                          });\n                        })\n                      }\n                    });\n                  ", "disallow-in-func"},
		{"a function-expression callback in a class component under disallow-in-func", "\n                    class Hello extends React.Component {\n                      componentWillUpdate() {\n                        someClass.onSomeEvent(function(data) {\n                          this.setState({\n                            data: data\n                          });\n                        })\n                      }\n                    }\n                  ", "disallow-in-func"},
		{"setState inside an if block in createReactClass", "\n                    var Hello = createReactClass({\n                      componentWillUpdate: function() {\n                        if (true) {\n                          this.setState({\n                            data: data\n                          });\n                        }\n                      }\n                    });\n                  ", ""},
		{"setState inside an if block in a class component", "\n                    class Hello extends React.Component {\n                      componentWillUpdate() {\n                        if (true) {\n                          this.setState({\n                            data: data\n                          });\n                        }\n                      }\n                    }\n                  ", ""},
		{"a concise-body arrow callback in createReactClass under disallow-in-func", "\n                    var Hello = createReactClass({\n                      componentWillUpdate: function() {\n                        someClass.onSomeEvent((data) => this.setState({data: data}));\n                      }\n                    });\n                  ", "disallow-in-func"},
		{"a concise-body arrow callback in a class component under disallow-in-func", "\n                    class Hello extends React.Component {\n                      componentWillUpdate() {\n                        someClass.onSomeEvent((data) => this.setState({data: data}));\n                      }\n                    }\n                  ", "disallow-in-func"},
		{"an UNSAFE prefixed class lifecycle method", "\n                    class Hello extends React.Component {\n                      UNSAFE_componentWillUpdate() {\n                        this.setState({\n                          data: data\n                        });\n                      }\n                    }\n                  ", ""},
		{"an UNSAFE prefixed createReactClass lifecycle property", "\n                    var Hello = createReactClass({\n                      UNSAFE_componentWillUpdate: function() {\n                        this.setState({\n                          data: data\n                        });\n                      }\n                    });\n                  ", ""},

		// Everything below is ours rather than upstream's, and every one was measured on the
		// release oxlint binary rather than reasoned about. They cover the discriminations our
		// abstract syntax tree spells differently than oxc's does, the shapes upstream's corpus is
		// silent on, and the one place our own shelf would have given a different answer.

		// The two conditions are independent, which is the real surprise in this rule. Nothing
		// checks that the lifecycle-named member belongs to the component the walk later reaches.
		// A plain object literal carrying the property, nested inside any component, reports as
		// though it were that component's own method. Measured, because the loop reads as though it
		// were checking a component's method and is not.
		{"a lifecycle property on a plain object literal nested inside a component", "class Hello extends React.Component {\n  render() {\n    var notMine = {\n      componentWillUpdate: function() { this.setState({}); }\n    };\n    return null;\n  }\n}\n", ""},

		// The component predicates, both eras. `React.createReactClass` counts because oxc accepts
		// the namespaced spelling of that one name. The bare `Component` and `PureComponent` count
		// because a file importing them directly writes them that way.
		{"the namespaced createReactClass factory", "var Hello = React.createReactClass({\n  componentWillUpdate: function() { this.setState({}); }\n});\n", ""},
		{"a class extending the bare Component identifier", "class Hello extends Component {\n  componentWillUpdate() { this.setState({}); }\n}\n", ""},
		{"a class extending React.PureComponent", "class Hello extends React.PureComponent {\n  componentWillUpdate() { this.setState({}); }\n}\n", ""},
		{"a class extending the bare PureComponent identifier", "class Hello extends PureComponent {\n  componentWillUpdate() { this.setState({}); }\n}\n", ""},

		// The class property holding an arrow. The arrow is a separate node passed by the walk
		// before the property that names it, so it must not be double counted; it reports at depth
		// one like the two method spellings do.
		{"an arrow-valued componentWillUpdate class property", "class Hello extends React.Component {\n  componentWillUpdate = () => { this.setState({}); }\n}\n", ""},

		// Counting stops at the first name match, which is what lets a lifecycle-named object
		// literal nested inside the real lifecycle method report at depth one rather than two.
		// Written the other way it would be silent, so this case pins the direction.
		{"an object literal named for the lifecycle nested inside the real lifecycle method", "class Hello extends React.Component {\n  componentWillUpdate() {\n    const o = { componentWillUpdate() { this.setState({}); } };\n  }\n}\n", ""},

		// Nothing asks what kind of member carries the name. A static method is not a lifecycle
		// method at all and reports, and both accessors report, because oxc reaches all three
		// through one `MethodDefinition` arm and never asks. Reproduced rather than corrected.
		{"a static method named for the lifecycle", "class Hello extends React.Component {\n  static componentWillUpdate() { this.setState({}); }\n}\n", ""},
		{"a getter named for the lifecycle", "class Hello extends React.Component {\n  get componentWillUpdate() { this.setState({}); return 1; }\n}\n", ""},
		{"a setter named for the lifecycle", "class Hello extends React.Component {\n  set componentWillUpdate(v) { this.setState({}); }\n}\n", ""},

		// The key is resolved statically, not matched as an identifier. oxc calls `static_name()`,
		// which answers for a computed key holding a string literal and for a string-literal key.
		// A port matching only the identifier spelling loses both of these findings, and reading
		// `Text()` on the computed one crashes the linter outright.
		{"a computed lifecycle key holding a string literal", "class Hello extends React.Component {\n  [\"componentWillUpdate\"]() { this.setState({}); }\n}\n", ""},
		{"a string-literal lifecycle key", "class Hello extends React.Component {\n  \"componentWillUpdate\"() { this.setState({}); }\n}\n", ""},

		// The callee spellings. The optional forms need no special handling and were measured
		// rather than assumed. The bracketed ones are our `ElementAccessExpression`, which oxc
		// reaches through the same `static_property_name()` as the dotted form.
		{"an optional receiver, this?.setState", "class Hello extends React.Component {\n  componentWillUpdate() { this?.setState({}); }\n}\n", ""},
		{"an optional call, this.setState?.()", "class Hello extends React.Component {\n  componentWillUpdate() { this.setState?.({}); }\n}\n", ""},
		{"a bracketed string setState", "class Hello extends React.Component {\n  componentWillUpdate() { this[\"setState\"]({}); }\n}\n", ""},
		{"a bracketed template-literal setState", "class Hello extends React.Component {\n  componentWillUpdate() { this[`setState`]({}); }\n}\n", ""},

		// Parentheses are skipped on the factory callee and nowhere else, and that asymmetry is
		// upstream's. The two silent paren shapes are in the clean list below; all three were
		// measured in one invocation of the release binary.
		{"a parenthesized createReactClass callee", "var Hello = (createReactClass)({\n  componentWillUpdate: function() { this.setState({}); }\n});\n", ""},

		// The same six shapes from the clean list, under `disallow-in-func`. These pin the option
		// as the thing that changes the answer rather than the shapes being unreachable, which is
		// what a clean-only fixture would leave ambiguous.
		{"a nested object method inside a createReactClass lifecycle under disallow-in-func", "var Hello = createReactClass({\n  componentWillUpdate: function() {\n    var o = { inner() { this.setState({}); } };\n  }\n});\n", "disallow-in-func"},
		{"a nested object method inside a class lifecycle under disallow-in-func", "class Hello extends React.Component {\n  componentWillUpdate() {\n    var o = { inner() { this.setState({}); } };\n  }\n}\n", "disallow-in-func"},
		{"a method on a class declared inside the lifecycle under disallow-in-func", "class Hello extends React.Component {\n  componentWillUpdate() {\n    class Inner { m() { this.setState({}); } }\n  }\n}\n", "disallow-in-func"},
		{"a nested object getter inside the lifecycle under disallow-in-func", "class Hello extends React.Component {\n  componentWillUpdate() {\n    var o = { get g() { this.setState({}); return 1; } };\n  }\n}\n", "disallow-in-func"},
		{"an arrow callback in a class component under disallow-in-func", "class Hello extends React.Component {\n  componentWillUpdate() { on(() => { this.setState({}); }); }\n}\n", "disallow-in-func"},
		{"an arrow callback in a createReactClass component under disallow-in-func", "var Hello = createReactClass({\n  componentWillUpdate: function() { on(() => { this.setState({}); }); }\n});\n", "disallow-in-func"},

		// Upstream's sixth pass case, recorded as reporting. See the block comment above.
		{"an UNSAFE prefixed lifecycle upstream silences with a React version setting", "\n                    class Hello extends React.Component {\n                      UNSAFE_componentWillUpdate() {\n                        this.setState({\n                          data: data\n                        });\n                      }\n                    }\n                  ", ""}}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoWillUpdateSetState, willUpdateSetStateFile,
				testCase.sourceText, noMethodSetStateOptions(t, testCase.option))
			rule_testing.ExpectFindings(t, result, "noSetStateInComponentWillUpdate")
		})
	}
}

// The clean cases carry the whole rule, and they decline for eight different reasons.
//
// Three are upstream's nesting judgment in default mode: a function-expression callback, a function
// declared inside the lifecycle and passed out, and the same written again under an empty option
// array. Every one of them reports once `disallow-in-func` is passed, which is what the option is
// for. One is `this.setState` referenced without being called, which is an assignment rather than a
// call and never reaches the listener. One is an empty lifecycle body, the corpus checking the rule
// does not fire on the method itself.
//
// The rest are ours: the component gate, the two factory names our shelf accepts and this rule's
// upstream does not, the two neighbouring lifecycle names, four callee shapes, the absence of any
// call-graph analysis, both parenthesized forms, and a tagged template.
func TestNoWillUpdateSetStateStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		option     string
	}{
		// Upstream, verbatim.
		{"a render method with no lifecycle at all", "\n                    var Hello = createReactClass({\n                      render: function() {\n                        return <div>Hello {this.props.name}</div>;\n                      }\n                    });\n                  ", ""},
		{"an empty componentWillUpdate", "\n                    var Hello = createReactClass({\n                      componentWillUpdate: function() {}\n                    });\n                  ", ""},
		{"setState referenced but never called", "\n                    var Hello = createReactClass({\n                      componentWillUpdate: function() {\n                        someNonMemberFunction(arg);\n                        this.someHandler = this.setState;\n                      }\n                    });\n                  ", ""},
		{"setState inside a function-expression callback", "\n                    var Hello = createReactClass({\n                      componentWillUpdate: function() {\n                        someClass.onSomeEvent(function(data) {\n                          this.setState({\n                            data: data\n                          });\n                        })\n                      }\n                    });\n                  ", ""},
		{"setState inside a function declared in the lifecycle and passed out", "\n                    var Hello = createReactClass({\n                      componentWillUpdate: function() {\n                        function handleEvent(data) {\n                          this.setState({\n                            data: data\n                          });\n                        }\n                        someClass.onSomeEvent(handleEvent)\n                      }\n                    });\n                  ", ""},
		{"a function declared in the lifecycle under an empty option array", "\n                    var Hello = createReactClass({\n                      componentWillUpdate: function() {\n                        function handleEvent(data) {\n                          this.setState({\n                            data: data\n                          });\n                        }\n                        someClass.onSomeEvent(handleEvent)\n                      }\n                    });\n                  ", ""},

		// The component gate, which ESLint does not have and oxc does. ESLint's factory matches a
		// property named for the lifecycle method and stops, so it reports on all three of these.
		// oxc requires the walk to reach an ES5 or ES6 component after the name match and is silent
		// on all three. oxc wins because oxc is what the differential gate compares against.
		{"a lifecycle method on a class extending nothing", "class Hello {\n  componentWillUpdate() { this.setState({}); }\n}\n", ""},
		{"a lifecycle property on a bare object literal at the top level", "var Hello = {\n  componentWillUpdate: function() { this.setState({}); }\n};\n", ""},
		{"a lifecycle property in a call that is not createReactClass", "var Hello = somethingElse({\n  componentWillUpdate: function() { this.setState({}); }\n});\n", ""},

		// `createClass` and `React.createClass` are NOT components to this rule, which is why the
		// ES5 predicate is spelled in the rule file rather than taken from the shelf. Both are
		// clean on the release binary and `utilsreact.IsEs5ComponentCall` accepts both, so reaching
		// for it would report where upstream says nothing and no imported fixture could see it.
		{"the createClass factory name, which this rule's upstream does not accept", "var Hello = createClass({\n  componentWillUpdate: function() { this.setState({}); }\n});\n", ""},
		{"the React.createClass factory name, which this rule's upstream does not accept", "var Hello = React.createClass({\n  componentWillUpdate: function() { this.setState({}); }\n});\n", ""},

		// The lifecycle name is one string plus its UNSAFE prefix, and the two neighbouring rules
		// in this package differ from this one in exactly that string. These two matter more here
		// than anywhere else because all three rules live in the same file tree.
		{"setState directly inside componentDidUpdate", "class Hello extends React.Component {\n  componentDidUpdate() { this.setState({}); }\n}\n", ""},
		{"setState directly inside componentDidMount", "class Hello extends React.Component {\n  componentDidMount() { this.setState({}); }\n}\n", ""},

		// The callee gate. A different receiver, a different method name, and two computed keys
		// that name nothing this rule matches.
		{"a receiver that is not this", "class Hello extends React.Component {\n  componentWillUpdate() { that.setState({}); }\n}\n", ""},
		{"a method that is not setState", "class Hello extends React.Component {\n  componentWillUpdate() { this.setStates({}); }\n}\n", ""},
		{"a computed callee whose key names nothing statically", "class Hello extends React.Component {\n  componentWillUpdate() { this[foo]({}); }\n}\n", ""},
		{"a numeric computed callee", "class Hello extends React.Component {\n  componentWillUpdate() { this[0]({}); }\n}\n", ""},

		// There is no call-graph analysis anywhere in either upstream, and this is the thing most
		// likely to be assumed. A helper method that calls setState, invoked from the lifecycle
		// method, is silent at both tools and silent here. Only ancestry counts.
		{"a helper method called from the lifecycle method", "class Hello extends React.Component {\n  componentWillUpdate() { helper(); }\n  helper() { this.setState({}); }\n}\n", ""},

		// Parentheses, measured on the release binary rather than guessed. oxc destructures
		// `Expression::ThisExpression` directly and reaches `callee.as_member_expression()`
		// directly, and neither looks through a parenthesized expression. Skipping here reads as a
		// free correctness improvement and would report where upstream is silent.
		{"a parenthesized this receiver", "class Hello extends React.Component {\n  componentWillUpdate() { (this).setState({}); }\n}\n", ""},
		{"a parenthesized setState callee", "class Hello extends React.Component {\n  componentWillUpdate() { (this.setState)({}); }\n}\n", ""},

		// The nesting judgment applies to the UNSAFE spelling identically; the prefix changes which
		// names latch and nothing else.
		{"a callback inside an UNSAFE prefixed lifecycle method", "class Hello extends React.Component {\n  UNSAFE_componentWillUpdate() { on(function(){ this.setState({}); }); }\n}\n", ""},

		// A nested member that carries its own function scope still counts as a function, and this
		// block is here because a mutation sweep found the rule reporting on all four of these
		// while upstream is silent on all four. The cause is the one place our tree fuses two of
		// oxc's nodes into one: a `MethodDeclaration` is both the name that latches AND the
		// function scope, so counting it only at the latch left every method NOT on the latch
		// uncounted, and a `setState` two scopes deep read as one scope deep. oxc has a separate
		// `Function` node under its `MethodDefinition` and its walk counts that node either way.
		// All four report under `disallow-in-func` and all four are clean by default, on the
		// release binary and here, checked cell by cell.
		{"a nested object method inside a createReactClass lifecycle", "var Hello = createReactClass({\n  componentWillUpdate: function() {\n    var o = { inner() { this.setState({}); } };\n  }\n});\n", ""},
		{"a nested object method inside a class lifecycle", "class Hello extends React.Component {\n  componentWillUpdate() {\n    var o = { inner() { this.setState({}); } };\n  }\n}\n", ""},
		{"a method on a class declared inside the lifecycle method", "class Hello extends React.Component {\n  componentWillUpdate() {\n    class Inner { m() { this.setState({}); } }\n  }\n}\n", ""},
		{"a nested object getter inside the lifecycle method", "class Hello extends React.Component {\n  componentWillUpdate() {\n    var o = { get g() { this.setState({}); return 1; } };\n  }\n}\n", ""},

		// An arrow callback in default mode. Upstream's corpus writes arrow callbacks ONLY under
		// `disallow-in-func`, so nothing imported covers the arrow being counted at all, and a
		// mutation dropping `KindArrowFunction` from the counted set survived the whole suite
		// until these two were added. Both measured silent by default and reporting under the
		// option, on the release binary.
		{"an arrow callback in a class component in default mode", "class Hello extends React.Component {\n  componentWillUpdate() { on(() => { this.setState({}); }); }\n}\n", ""},
		{"an arrow callback in a createReactClass component in default mode", "var Hello = createReactClass({\n  componentWillUpdate: function() { on(() => { this.setState({}); }); }\n});\n", ""},

		// A tagged template is no call expression, so the listener never sees it.
		{"a tagged template on setState, which is no call expression", "class Hello extends React.Component {\n  componentWillUpdate() { this.setState`x`; }\n}\n", ""}}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoWillUpdateSetState, willUpdateSetStateFile,
				testCase.sourceText, noMethodSetStateOptions(t, testCase.option))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The finding points at the callee and carries the rendered message, and both are asserted exactly.
//
// `ExpectFindings` asserts message ids and count and nothing else, so a rule pointing at the wrong
// node or rendering the wrong text passes a complete fixture pair while being wrong. Upstream
// underlines thirteen characters for `this.setState({...})`, stopping before the parenthesis, on
// all twelve of its snapshot diagnostics; the span is `call_expr.callee.span()` rather than the
// call's. The bracketed spelling is longer and is checked too, because a port reporting the whole
// call would pass a span assertion written only against the dotted form.
//
// The message text is compared with equality rather than `strings.Contains`, because a predicate
// weaker than the property it guards is not a guard: a doubled or truncated rendering contains the
// needle and stays green.
func TestNoWillUpdateSetStateReportsAtTheCallee(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{
			"the dotted spelling underlines this.setState and not the call",
			"class Hello extends React.Component {\n  componentWillUpdate() { this.setState({ data: 1 }); }\n}\n",
			"this.setState",
		},
		{
			"the bracketed spelling underlines the whole member expression",
			"class Hello extends React.Component {\n  componentWillUpdate() { this[\"setState\"]({ data: 1 }); }\n}\n",
			"this[\"setState\"]",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoWillUpdateSetState, willUpdateSetStateFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want exactly one finding, got %d", len(result.Diagnostics))
			}
			finding := result.Diagnostics[0]
			reported := testCase.sourceText[finding.Range.Pos():finding.Range.End()]
			if reported != testCase.wantText {
				t.Errorf("reported span is %q, want %q", reported, testCase.wantText)
			}
			if got := finding.Message.Description; got != messageNoSetStateInComponentWillUpdate.Description {
				t.Errorf("rendered description is %q, want %q", got,
					messageNoSetStateInComponentWillUpdate.Description)
			}
			if !strings.HasPrefix(finding.Message.Description, "Updating state from ") {
				t.Errorf("description does not open by naming the mistake: %q",
					finding.Message.Description)
			}
		})
	}
}

// An unconfigured rule gets the permissive default, and the option is what changes that answer.
//
// `rule_testing.Run` hands the rule a nil options value rather than a zero-valued struct, so a rule
// reading its options with an unchecked type assertion panics on every real file while every
// fixture using `RunWithOptions` stays green. This asserts the plain harness works, which is the
// shape the linter actually runs in for a rule configured with a bare severity.
func TestNoWillUpdateSetStateDefaultsToAllowingNestedFunctions(t *testing.T) {
	t.Parallel()

	const callbackSource = "class Hello extends React.Component {\n" +
		"  componentWillUpdate() { on(function() { this.setState({}); }); }\n}\n"

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoWillUpdateSetState, willUpdateSetStateFile,
		callbackSource))

	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoWillUpdateSetState,
		willUpdateSetStateFile, callbackSource,
		noMethodSetStateOptions(t, "disallow-in-func")),
		"noSetStateInComponentWillUpdate")
}

// Two calls in one lifecycle method body report once each.
//
// The nesting judgment is per call site rather than per method, so a method holding two direct
// calls produces two findings. This lives in its own test rather than the table above because
// `ExpectFindings` takes one id per finding and the table asserts exactly one. Measured on the
// release binary, which prints two diagnostics for this source at two different columns.
func TestNoWillUpdateSetStateReportsEachCallSite(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoWillUpdateSetState, willUpdateSetStateFile,
		"class Hello extends React.Component {\n"+
			"  componentWillUpdate() { this.setState({}); this.setState({}); }\n}\n")
	rule_testing.ExpectFindings(t, result, "noSetStateInComponentWillUpdate",
		"noSetStateInComponentWillUpdate")
}
