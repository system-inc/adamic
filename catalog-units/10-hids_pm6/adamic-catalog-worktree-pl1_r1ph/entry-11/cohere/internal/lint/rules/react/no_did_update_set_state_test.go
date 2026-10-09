package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// didUpdateSetStateFile is where the fixtures pretend to live.
//
// A `.tsx` name because upstream's corpus is written in JSX and two of its clean cases carry a JSX
// return. This rule reads no JSX node and declares no file gate, unlike `no-string-refs` in this
// same package, so the suffix is only about the fixtures parsing.
const didUpdateSetStateFile = "/repository/source/DidUpdate.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/react/no_did_update_set_state.rs`:
// 16 pass and 18 fail. The extractor reported 18 diagnostics against those 18 fail inputs, so the
// mapping is one finding per input and no recovery by line alignment was needed. That is worth
// stating because one fail case carries two `setState` calls and reports only once, which is the
// nesting judgment rather than a counting accident.
//
// Four of those 34 were written under oxc's `allowed` spelling of the default, two passing and two
// failing, each byte-identical to a default-mode case kept here. No ESLint version accepts
// `allowed`, so the decoder refuses it and those four rows were dropped (#d21war2).
//
// The strings were decoded from the extractor's `-dump` output and then checked byte against byte
// into the Rust source by a script before any Go was written. All 34 matched, and none of them
// carries a backslash or a quote, so the cooking hazard that has bitten three previous porters
// could not arise here. That was measured rather than assumed.
//
// The option column is the raw upstream spelling: "" for an absent option, and otherwise the single
// string upstream passes positionally.
func TestNoDidUpdateSetStateFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		option     string
	}{
		{"setState directly inside a createReactClass componentDidUpdate", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {\n                this.setState({\n                  name: this.props.name.toUpperCase()\n                });\n              },\n              render: function() {\n                return <div>Hello {this.state.name}</div>;\n              }\n            });\n            ", ""},
		{"setState inside a named function expression lifecycle", "\n            var Hello = createReactClass({\n              componentDidUpdate: function componentDidUpdate() {\n                this.setState({\n                  name: this.props.name.toUpperCase()\n                });\n              }\n            });\n            ", ""},
		{"setState directly inside a class componentDidUpdate", "\n            class Hello extends React.Component {\n              componentDidUpdate() {\n                this.setState({\n                  name: this.props.name.toUpperCase()\n                });\n              }\n            }\n            ", ""},
		{"setState inside an arrow-valued componentDidUpdate class property", "\n            class Hello extends React.Component {\n              componentDidUpdate = () => {\n                this.setState({\n                  name: this.props.name.toUpperCase()\n                });\n              }\n            }\n            ", ""},
		{"a direct setState beside one inside a callback, only the direct one reports", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {\n                this.setState({ data: 1 });\n                someClass.onSomeEvent(function(data) {\n                  this.setState({ data: 2 });\n                })\n              }\n            });\n            ", ""},
		{"setState inside an if block in a class component", "\n            class Hello extends React.Component {\n              componentDidUpdate() {\n                if (true) {\n                  this.setState({ data: 123 });\n                }\n              }\n            }\n            ", ""},
		{"setState inside an if block in a createReactClass component", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {\n                if (true) {\n                  this.setState({ data: 123 });\n                }\n              }\n            });\n            ", ""},
		{"setState inside a conditional expression", "\n            class Hello extends React.Component {\n              componentDidUpdate() {\n                const x = true ? this.setState({ data: 123 }) : null;\n              }\n            }\n            ", ""},
		{"a direct setState under disallow-in-func", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {\n                this.setState({\n                  name: this.props.name.toUpperCase()\n                });\n              }\n            });\n            ", "disallow-in-func"},
		{"a function-expression callback under disallow-in-func", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {\n                someClass.onSomeEvent(function(data) {\n                  this.setState({\n                    data: data\n                  });\n                })\n              }\n            });\n            ", "disallow-in-func"},
		{"a setTimeout arrow under disallow-in-func", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {\n                setTimeout(() => {\n                  this.setState({ data: 123 });\n                }, 100);\n              }\n            });\n            ", "disallow-in-func"},
		{"a promise-then arrow under disallow-in-func", "\n            class Hello extends React.Component {\n              componentDidUpdate() {\n                Promise.resolve().then(() => {\n                  this.setState({ data: 123 });\n                });\n              }\n            }\n            ", "disallow-in-func"},
		{"a direct setState in a class component under disallow-in-func", "\n            class Hello extends React.Component {\n              componentDidUpdate() {\n                this.setState({\n                  data: data\n                });\n              }\n            }\n            ", "disallow-in-func"},
		{"a function-expression callback in a class component under disallow-in-func", "\n            class Hello extends React.Component {\n              componentDidUpdate() {\n                someClass.onSomeEvent(function(data) {\n                  this.setState({\n                    data: data\n                  });\n                })\n              }\n            }\n            ", "disallow-in-func"},
		{"a concise-body arrow callback under disallow-in-func", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {\n                someClass.onSomeEvent((data) => this.setState({data: data}));\n              }\n            });\n            ", "disallow-in-func"},
		{"a concise-body arrow callback in a class component under disallow-in-func", "\n            class Hello extends React.Component {\n              componentDidUpdate() {\n                someClass.onSomeEvent((data) => this.setState({data: data}));\n              }\n            }\n            ", "disallow-in-func"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoDidUpdateSetState, didUpdateSetStateFile,
				testCase.sourceText, noMethodSetStateOptions(t, testCase.option))
			rule_testing.ExpectFindings(t, result, "noSetStateInComponentDidUpdate")
		})
	}
}

// The clean cases carry the whole rule, and they decline for four different reasons.
//
// Five of them are the nesting judgment in default mode: a callback, a function declared inside the
// lifecycle and passed out, an arrow handler, a `setTimeout` and a promise continuation. Those are
// the cases the option exists to change, and every one of them reports once `disallow-in-func` is
// passed.
//
// Three are the lifecycle name: `componentDidMount` and `componentWillUpdate` hold a `setState`
// that a neighbouring rule reports and this one must not. Those two matter more here than anywhere
// else in this package, because `no-did-mount-set-state` is being ported beside this rule and the
// two differ in exactly one string.
//
// One is `this.setState` written without calling it, which is an assignment rather than a call and
// never reaches the call listener. One is a plain function with no component anywhere, which is the
// component gate. One is an empty lifecycle body, which is the corpus checking the rule does not
// fire on the method itself.
func TestNoDidUpdateSetStateStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		option     string
	}{
		{"a render method with no lifecycle at all", "\n            var Hello = createReactClass({\n              render: function() {\n                return <div>Hello {this.props.name}</div>;\n              }\n            });\n            ", ""},
		{"an empty componentDidUpdate", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {}\n            });\n            ", ""},
		{"setState referenced but never called", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {\n                someNonMemberFunction(arg);\n                this.someHandler = this.setState;\n              }\n            });\n            ", ""},
		{"setState inside a function-expression callback", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {\n                someClass.onSomeEvent(function(data) {\n                  this.setState({\n                    data: data\n                  });\n                })\n              }\n            });\n            ", ""},
		{"setState inside a function declared in the lifecycle and passed out", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {\n                function handleEvent(data) {\n                  this.setState({\n                    data: data\n                  });\n                }\n                someClass.onSomeEvent(handleEvent)\n              }\n            });\n            ", ""},
		{"setState inside an arrow callback in a class component", "\n            class Hello extends React.Component {\n              componentDidUpdate() {\n                this.handleEvent(() => {\n                  this.setState({ data: 123 });\n                });\n              }\n            }\n            ", ""},
		{"setState directly inside componentDidMount", "\n            class Hello extends React.Component {\n              componentDidMount() {\n                this.setState({ data: 123 });\n              }\n            }\n            ", ""},
		{"setState directly inside componentWillUpdate", "\n            class Hello extends React.Component {\n              componentWillUpdate() {\n                this.setState({ data: 123 });\n              }\n            }\n            ", ""},
		{"setState directly inside a createReactClass componentDidMount", "\n            var Hello = createReactClass({\n              componentDidMount: function() {\n                this.setState({ data: 123 });\n              }\n            });\n            ", ""},
		{"setState in a plain function that is no component", "\n            function Hello() {\n              this.setState({ data: 123 });\n            }\n            ", ""},
		{"setState inside a setTimeout arrow", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {\n                setTimeout(() => {\n                  this.setState({ data: 123 });\n                }, 100);\n              }\n            });\n            ", ""},
		{"setState inside a promise-then arrow", "\n            class Hello extends React.Component {\n              componentDidUpdate() {\n                Promise.resolve().then(() => {\n                  this.setState({ data: 123 });\n                });\n              }\n            }\n            ", ""},
		{"an empty componentDidUpdate under disallow-in-func", "\n            var Hello = createReactClass({\n              componentDidUpdate: function() {}\n            });\n            ", "disallow-in-func"},
		{"a render method with no lifecycle under disallow-in-func", "\n            var Hello = createReactClass({\n              render: function() {\n                return <div>Hello {this.props.name}</div>;\n              }\n            });\n            ", "disallow-in-func"},

		// Everything below is ours rather than upstream's, and every one of them was measured on
		// the release oxlint binary rather than reasoned about. They cover the discriminations our
		// AST forces us to spell differently than oxc does, and the two places our own shelf would
		// have given a different answer than upstream gives.

		// The component gate, which ESLint does not have and oxc does. ESLint's factory matches a
		// property named `componentDidUpdate` and stops, so it reports on all three of these. oxc
		// requires the walk to reach an ES5 or ES6 component after finding the lifecycle name, and
		// is silent on all three. oxc wins because oxc is what the differential gate compares
		// against. Measured: all three are clean on the release binary.
		{"a lifecycle method on a class extending nothing", "\nclass Hello {\n  componentDidUpdate() {\n    this.setState({ data: 123 });\n  }\n}\n", ""},
		{"a lifecycle property on a plain object literal", "\nvar Hello = {\n  componentDidUpdate: function() {\n    this.setState({ data: 123 });\n  }\n};\n", ""},
		{"a lifecycle property in a call that is not createReactClass", "\nvar Hello = somethingElse({\n  componentDidUpdate: function() {\n    this.setState({ data: 123 });\n  }\n});\n", ""},

		// `createClass` and `React.createClass` are NOT components to this rule, and this is the
		// one place the shelf would have over-reported. `utilsreact.IsEs5ComponentCall` accepts
		// both of those names as well as `createReactClass`, matching a different oxc helper, while
		// `no_did_update_set_state.rs` reaches `is_es5_component`, which accepts only the name
		// `createReactClass`. Measured: both spellings below are clean on the release binary, so
		// reaching for the shelf here would have reported twice where upstream says nothing.
		{"a bare createClass call is not the ES5 component this rule means", "\nvar Hello = createClass({\n  componentDidUpdate: function() {\n    this.setState({ data: 123 });\n  }\n});\n", ""},
		{"React.createClass is not the ES5 component this rule means either", "\nvar Hello = React.createClass({\n  componentDidUpdate: function() {\n    this.setState({ data: 123 });\n  }\n});\n", ""},

		// The heritage check is on React specifically. `Foo.Component` is somebody else's class.
		{"a class extending a namespaced Component that is not React", "\nclass Hello extends Foo.Component {\n  componentDidUpdate() {\n    this.setState({ data: 123 });\n  }\n}\n", ""},

		// Upstream does not implement the `UNSAFE_` prefix for this rule. ESLint's factory takes an
		// optional callback enabling it and `no-did-update-set-state` passes none, and oxc has no
		// notion of it at all. Measured clean on the release binary; reproduced rather than
		// improved on.
		{"the UNSAFE_ prefixed spelling, which upstream does not implement", "\nclass Hello extends React.Component {\n  UNSAFE_componentDidUpdate() {\n    this.setState({ data: 123 });\n  }\n}\n", ""},

		// A method call on something other than `this`. The receiver check is not incidental: a
		// store or a child ref with a `setState` of its own is ordinary code.
		{"setState called on something that is not this", "\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    other.setState({ data: 123 });\n    this.other.setState({ data: 456 });\n  }\n}\n", ""},

		// A different method on `this`. The property name is the other half of the callee check.
		{"a different method called on this", "\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    this.forceUpdate();\n  }\n}\n", ""},

		// An interpolated computed key cannot name the lifecycle statically, and reading its text
		// is what panics the shim. This case exists so that a port reaching for `Text()` on a
		// computed key crashes here rather than in Kirk's tree.
		{"an interpolated computed key beside a real one", "\nvar Hello = createReactClass({\n  [`componentDidUpdate${suffix}`]: function() {\n    this.setState({ data: 123 });\n  }\n});\n", ""},

		// A component nested inside the lifecycle. The walk stops at the FIRST component it reaches
		// after the lifecycle name, and here it never reaches the lifecycle name at all: the inner
		// `render` is not `componentDidUpdate`, so the counting loop runs to the top and answers
		// nothing. Measured clean on the release binary.
		{"a component defined inside the lifecycle of another component", "\nclass Outer extends React.Component {\n  componentDidUpdate() {\n    var Inner = createReactClass({\n      render: function() {\n        this.setState({ data: 1 });\n        return null;\n      }\n    });\n  }\n}\n", ""},

		// A callback inside a getter named for the lifecycle, which is the accessor's own function
		// scope being counted. Found by a mutation sweep: dropping the accessors from the
		// method-like count survived every fixture, because the getter case above reports at either
		// depth and nothing distinguished one from two. This input does distinguish them. Measured
		// silent by default and reporting under the option on the release binary, so without
		// counting the accessor the default mode reports and upstream does not.
		{"a callback inside a getter named for the lifecycle", "\nclass Hello extends React.Component {\n  get componentDidUpdate() {\n    someClass.on(function() {\n      this.setState({ data: 123 });\n    });\n    return 1;\n  }\n}\n", ""},

		// Parentheses, which are the two cases this rule gets right by NOT being clever and which
		// a mutation sweep found unguarded. Both were measured silent on the release binary before
		// the rule was written, and then no fixture asserted either, so a mutant adding the
		// paren-skip that reads like an improvement survived twice. oxc destructures
		// `Expression::ThisExpression` for the receiver and reaches `as_member_expression()` for the
		// callee, and neither looks through a parenthesized expression.
		//
		// Worth contrasting with `(createReactClass)({...})`, which DOES report, because the ES5
		// component check goes through `get_identifier_reference()`, which does skip parentheses.
		// The asymmetry is upstream's and the fires table below pins that half.
		{"a parenthesized this receiver", "\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    (this).setState({ data: 123 });\n  }\n}\n", ""},
		{"a parenthesized callee", "\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    (this.setState)({ data: 123 });\n  }\n}\n", ""},

		// Depth two by way of an object literal that is not the component's own. Silent by default
		// and reported under the option, which is the same judgment as upstream's callback cases
		// reached by a shape upstream does not write. Both halves measured on the release binary.
		{"a nested callback under a lifecycle-named property of a foreign object", "\nclass Hello extends React.Component {\n  render() {\n    var obj = {\n      componentDidUpdate: function() {\n        someClass.on(function() {\n          this.setState({ data: 1 });\n        });\n      }\n    };\n    return null;\n  }\n}\n", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoDidUpdateSetState, didUpdateSetStateFile,
				testCase.sourceText, noMethodSetStateOptions(t, testCase.option))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Cases that report which upstream's corpus does not write, all measured on the release binary.
//
// These are the shapes our AST spells differently than oxc does, so each is a place a port can be
// silently narrower than upstream while every imported fixture stays green. oxc reads one
// `MemberExpression` where we have two node kinds, and one `static_name()` where we have three key
// spellings, so the count of things that have to be right here is larger on our side than on
// upstream's and none of it is visible from the imported corpus.
func TestNoDidUpdateSetStateFiresOnShapesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		option     string
	}{
		// Our AST splits oxc's member expression in two. `this["setState"]` is an
		// `ElementAccessExpression` here and a computed `MemberExpression` there, and
		// `static_property_name()` resolves it, so upstream reports. Measured: reports.
		{"setState reached through a computed member", "\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    this[\"setState\"]({ data: 123 });\n  }\n}\n", ""},

		// Both optional spellings report upstream. `this?.setState(...)` is a
		// PropertyAccessExpression carrying a question-dot token here, and `this.setState?.(...)`
		// is an ordinary access under a call marked optional. Measured: both report, two findings
		// from the same file, so a port matching only the plain spelling loses both.
		{"setState called through an optional access", "\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    this?.setState({ data: 123 });\n  }\n}\n", ""},
		{"setState called through an optional call", "\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    this.setState?.({ data: 123 });\n  }\n}\n", ""},

		// The lifecycle key resolves statically through three spellings, not one. oxc asks
		// `static_name()`, which answers for an identifier, a string literal and a computed key
		// holding a literal. All three measured reporting on the release binary, and the computed
		// one is the shape that panics a port reading key text before checking kind.
		{"a string-literal lifecycle key", "\nvar Hello = createReactClass({\n  \"componentDidUpdate\": function() {\n    this.setState({ data: 123 });\n  }\n});\n", ""},
		{"a computed lifecycle key holding a string literal", "\nvar Hello = createReactClass({\n  [\"componentDidUpdate\"]: function() {\n    this.setState({ data: 123 });\n  }\n});\n", ""},

		// A getter named for the lifecycle. oxc sees a `MethodDefinition` whose kind is Get, which
		// its `is_lifecycle_component_method` accepts without asking what kind of method it is, and
		// the getter's own function is what makes the depth 1. Measured: reports. This is the case
		// `scope.EnclosingFunctionLike` documents four rules disagreeing about, and the answer here
		// comes from upstream rather than from that discussion.
		{"a getter named for the lifecycle", "\nclass Hello extends React.Component {\n  get componentDidUpdate() {\n    this.setState({ data: 123 });\n    return 1;\n  }\n}\n", ""},

		// A static method. Nothing in either implementation asks whether the lifecycle method is on
		// the instance, and a static `componentDidUpdate` is not a lifecycle method at all, so this
		// is upstream over-reporting. Reproduced rather than corrected: measured reporting on the
		// release binary, and a port that quietly declined it would be a divergence nothing states.
		{"a static method named for the lifecycle", "\nclass Hello extends React.Component {\n  static componentDidUpdate() {\n    this.setState({ data: 123 });\n  }\n}\n", ""},

		// The component the walk finds does not have to be the one owning the lifecycle method.
		// A plain object literal with a lifecycle-named property, nested anywhere inside any
		// component, reports. This is the single most surprising thing the probes found and it is
		// upstream behavior on both implementations: the loop sets `in_lifecycle` on a name match
		// and then returns at the first component above it, without ever checking the two belong
		// together. Measured: reports on the release binary.
		{"a foreign object literal's lifecycle property inside a component", "\nclass Hello extends React.Component {\n  render() {\n    var obj = {\n      componentDidUpdate: function() {\n        this.setState({ data: 123 });\n      }\n    };\n    return null;\n  }\n}\n", ""},

		// A lifecycle name nested inside a lifecycle name. Counting stops at the innermost match,
		// so the depth is 1 and this reports in default mode. An implementation that kept counting
		// past the first match would make it 2 and go silent. Measured: reports in both modes.
		{"a lifecycle-named property inside the real lifecycle method", "\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    var o = { componentDidUpdate: function() { this.setState({ data: 1 }); } };\n  }\n}\n", ""},

		// The bare heritage spellings a file importing the base classes directly writes.
		{"a class extending a bare PureComponent", "\nclass Hello extends PureComponent {\n  componentDidUpdate() {\n    this.setState({ data: 123 });\n  }\n}\n", ""},
		{"a class extending a bare Component", "\nclass Hello extends Component {\n  componentDidUpdate() {\n    this.setState({ data: 123 });\n  }\n}\n", ""},

		// The parenthesized ES5 factory, which is the half of the parenthesis story that reports.
		// Measured on the release binary in the same run as the two silent cases above.
		{"a parenthesized createReactClass call", "\nvar Hello = (createReactClass)({\n  componentDidUpdate: function() {\n    this.setState({ data: 1 });\n  }\n});\n", ""},

		// A class expression rather than a declaration. Same class, written where a declaration
		// will not fit, and our heritage reader has to handle both node kinds.
		{"a class expression component", "\nvar Hello = class extends React.Component {\n  componentDidUpdate() {\n    this.setState({ data: 123 });\n  }\n};\n", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoDidUpdateSetState, didUpdateSetStateFile,
				testCase.sourceText, noMethodSetStateOptions(t, testCase.option))
			rule_testing.ExpectFindings(t, result, "noSetStateInComponentDidUpdate")
		})
	}
}

// The option is what changes the answer, and the corpus only ever exercises it one way at a time.
//
// Every case here is one source read under both settings, so a rule that ignored the option would
// fail rather than pass half of them vacuously. Upstream ships the halves separately, which cannot
// see a port that hardcodes either mode.
func TestNoDidUpdateSetStateOptionChangesTheAnswer(t *testing.T) {
	t.Parallel()

	nestedCallback := "\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    someClass.on(function() {\n      this.setState({ data: 123 });\n    });\n  }\n}\n"

	t.Run("silent by default", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, NoDidUpdateSetState, didUpdateSetStateFile,
			nestedCallback, noMethodSetStateOptions(t, ""))
		rule_testing.ExpectClean(t, result)
	})

	// The same shape one level deeper, through an accessor rather than a method. This is the pair
	// that pins the accessor's function scope being counted; see the clean case above.
	t.Run("a callback inside a getter reports under disallow-in-func", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, NoDidUpdateSetState, didUpdateSetStateFile,
			"\nclass Hello extends React.Component {\n  get componentDidUpdate() {\n    someClass.on(function() {\n      this.setState({ data: 123 });\n    });\n    return 1;\n  }\n}\n",
			noMethodSetStateOptions(t, "disallow-in-func"))
		rule_testing.ExpectFindings(t, result, "noSetStateInComponentDidUpdate")
	})

	t.Run("reports under disallow-in-func", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, NoDidUpdateSetState, didUpdateSetStateFile,
			nestedCallback, noMethodSetStateOptions(t, "disallow-in-func"))
		rule_testing.ExpectFindings(t, result, "noSetStateInComponentDidUpdate")
	})

	// A rule offered no options at all gets the zero value through the registry, which is the
	// permissive default. Asserted rather than assumed because the decode path returns an error
	// alongside the zero value when nothing is configured, and a rule reading the error instead of
	// the value would invert this.
	t.Run("no options at all reads as the default", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.Run(t, NoDidUpdateSetState, didUpdateSetStateFile, nestedCallback)
		rule_testing.ExpectClean(t, result)
	})
}

// Where the finding points, and what it says. `ExpectFindings` asserts neither.
//
// The span is the callee rather than the call: upstream underlines thirteen characters for
// `this.setState({...})`, stopping before the parenthesis, and the snapshot shows the same for
// every one of its eighteen diagnostics. A rule reporting the whole call passes every fixture above
// while pointing at the wrong range, and a rule reporting only `setState` passes them too.
func TestNoDidUpdateSetStatePointsAtTheCallee(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{
			"a dotted call underlines this.setState and not the arguments",
			"\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    this.setState({ data: 123 });\n  }\n}\n",
			"this.setState",
		},
		{
			"a computed call underlines the whole computed access",
			"\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    this[\"setState\"]({ data: 123 });\n  }\n}\n",
			"this[\"setState\"]",
		},
		{
			"a conditional puts the finding on the branch rather than the statement",
			"\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    const x = true ? this.setState({ data: 123 }) : null;\n  }\n}\n",
			"this.setState",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoDidUpdateSetState, didUpdateSetStateFile,
				testCase.sourceText, noMethodSetStateOptions(t, ""))
			rule_testing.ExpectFindings(t, result, "noSetStateInComponentDidUpdate")
			diagnostic := result.Diagnostics[0]
			reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.want {
				t.Fatalf("finding underlines %q, want %q", reported, testCase.want)
			}
		})
	}
}

// The rendered message, asserted on equality rather than containment.
//
// A `strings.Contains` predicate here would stay green through a doubled substitution, which is the
// exact defect that shipped from another rule in this tree. There is nothing interpolated into this
// message, which is precisely why the assertion is cheap enough to have no excuse.
func TestNoDidUpdateSetStateMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, NoDidUpdateSetState, didUpdateSetStateFile,
		"\nclass Hello extends React.Component {\n  componentDidUpdate() {\n    this.setState({ data: 123 });\n  }\n}\n",
		noMethodSetStateOptions(t, ""))
	rule_testing.ExpectFindings(t, result, "noSetStateInComponentDidUpdate")

	description := result.Diagnostics[0].Message.Description
	if !strings.HasPrefix(description, "Updating state from `componentDidUpdate`") {
		t.Fatalf("message description starts %q", description)
	}
	if strings.Count(description, "componentDidUpdate") != 1 {
		t.Fatalf("message names the lifecycle method %d times, want 1: %q",
			strings.Count(description, "componentDidUpdate"), description)
	}
}
