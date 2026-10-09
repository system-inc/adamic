package react

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// preferStatelessFunctionFile is where the fixtures pretend to live.
//
// A .tsx extension because the corpus is full of JSX, and NO file-suffix gate exists in this rule.
// Three shipped react rules here used to gate on .tsx or .jsx through an isJsxFileName helper,
// which was oxc residue rather than upstream behavior; all three gates were removed and the helper
// deleted. TestPreferStatelessFunctionHasNoFileSuffixGate below pins the absence.
const preferStatelessFunctionFile = "/repository/source/Pure.tsx"

// runPreferStatelessFunction routes every fixture through the rule's OWN exported decoder rather
// than building the options struct directly.
//
// That is what puts the decoder under test, and it is the line with no upstream counterpart. The
// raw json is the BARE object, not upstream's [{...}] array: cohere's config layer unwraps the
// [severity, options] tuple before dispatch, so copying ESLint's spelling fails on every row.
func runPreferStatelessFunction(t *testing.T, sourceText string, rawOptions string) rule_testing.Result {
	t.Helper()
	var decoded any
	var err error
	if rawOptions == "" {
		decoded, err = DecodePreferStatelessFunctionOptions(nil)
	} else {
		decoded, err = DecodePreferStatelessFunctionOptions(json.RawMessage(rawOptions))
	}
	if err != nil {
		t.Fatalf("decoding %q: %v", rawOptions, err)
	}
	return rule_testing.RunWithOptions(t, PreferStatelessFunction, preferStatelessFunctionFile, sourceText, decoded)
}

// The corpus is eslint-plugin-react's own, imported verbatim from
// tests/lib/rules/prefer-stateless-function.js by evaluating the upstream tester with a stub
// RuleTester and serializing what it was handed, so no case was retyped and no escape could be
// cooked on the way in. 30 valid and 20 invalid, every invalid case naming exactly one
// entry in its errors array.
//
// Seven valid cases carry ignorePureComponents, and cases 2 and 3 are the same shape as invalid
// cases 2 and 3 with only the option flipped, which is what pins the option rather than the rule.
func TestPreferStatelessFunctionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
		findings   []string
	}{
		{"\n        class Foo extends React.Component {\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          render() {\n            return <div>{this['props'].foo}</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.PureComponent {\n          render() {\n            return <div>foo</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.PureComponent {\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          static get displayName() {\n            return 'Foo';\n          }\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          static displayName = 'Foo';\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          static get propTypes() {\n            return {\n              name: PropTypes.string\n            };\n          }\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          static propTypes = {\n            name: PropTypes.string\n          };\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          props: {\n            name: string;\n          };\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          constructor() {\n            super();\n          }\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          render() {\n            let {props:{foo}, context:{bar}} = this;\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          render() {\n            if (!this.props.foo) {\n              return null;\n            }\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        var Foo = createReactClass({\n          render: function() {\n            if (!this.props.foo) {\n              return null;\n            }\n            return <div>{this.props.foo}</div>;\n          }\n        });\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          render() {\n            return true ? <div /> : null;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          static defaultProps = {\n            foo: true\n          }\n          render() {\n            const { foo } = this.props;\n            return foo ? <div /> : null;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          static get defaultProps() {\n            return {\n              foo: true\n            };\n          }\n          render() {\n            const { foo } = this.props;\n            return foo ? <div /> : null;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          render() {\n            const { foo } = this.props;\n            return foo ? <div /> : null;\n          }\n        }\n        Foo.defaultProps = {\n          foo: true\n        };\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          static contextTypes = {\n            foo: PropTypes.boolean\n          }\n          render() {\n            const { foo } = this.context;\n            return foo ? <div /> : null;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          static get contextTypes() {\n            return {\n              foo: PropTypes.boolean\n            };\n          }\n          render() {\n            const { foo } = this.context;\n            return foo ? <div /> : null;\n          }\n        }\n      ", "", []string{"componentShouldBePure"}},
		{"\n        class Foo extends React.Component {\n          render() {\n            const { foo } = this.context;\n            return foo ? <div /> : null;\n          }\n        }\n        Foo.contextTypes = {\n          foo: PropTypes.boolean\n        };\n      ", "", []string{"componentShouldBePure"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runPreferStatelessFunction(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// **Three upstream VALID cases are deliberately absent from this table, and they are not lost.**
// Each appears byte-identically in the fires table above, and the corpus lists it in both places
// with opposite verdicts. The only discriminator is `settings.react.version`, which upstream reads
// through `testReactVersion`: a stateless component could not return null before React 15, so a
// render returning null was a real reason to keep the class.
//
// Measured on the installed build with the version set explicitly, the same input both ways:
//
//	settings react 0.14.0   ->  CLEAN
//	settings react 19.0     ->  REPORTS
//
// cohere has no React-version setting and this tree is on React 19, so the modern verdict is the
// only one reachable here and the 0.14 rows would assert the opposite of what the rule must do.
// Recorded rather than deleted, because a case whose verdict is decided ABOVE the rule is a fact
// about the configuration rather than about the rule, and the next reader meeting three
// duplicate-looking entries deserves to know which axis separates them.
//
// The extractor originally dropped `settings`, which is what surfaced this: three silent cases
// failed, and the failure read like a rule defect until the corpus was re-read with that field
// rendered.
func TestPreferStatelessFunctionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
	}{
		{"\n        const Foo = function(props) {\n          return <div>{props.foo}</div>;\n        };\n      ", ""},
		{"const Foo = ({foo}) => <div>{foo}</div>;", ""},
		{"\n        class Foo extends React.PureComponent {\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", "{\"ignorePureComponents\":true}"},
		{"\n        class Foo extends React.PureComponent {\n          render() {\n            return <div>{this.context.foo}</div>;\n          }\n        }\n      ", "{\"ignorePureComponents\":true}"},
		{"\n        const Foo = class extends React.PureComponent {\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        };\n      ", "{\"ignorePureComponents\":true}"},
		{"\n        class Foo extends React.Component {\n          shouldComponentUpdate() {\n            return false;\n          }\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Foo extends React.Component {\n          changeState() {\n            this.setState({foo: \"clicked\"});\n          }\n          render() {\n            return <div onClick={this.changeState.bind(this)}>{this.state.foo || \"bar\"}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Foo extends React.Component {\n          doStuff() {\n            this.refs.foo.style.backgroundColor = \"red\";\n          }\n          render() {\n            return <div ref=\"foo\" onClick={this.doStuff}>{this.props.foo}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Foo extends React.Component {\n          doStuff() {}\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Foo extends React.Component {\n          constructor() {}\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Foo extends React.Component {\n          constructor() {\n            doSpecialStuffs();\n          }\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Foo extends React.Component {\n          constructor() {\n            foo;\n          }\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Foo extends React.Component {\n          constructor(props)\n\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Foo extends React.Component {\n          render() {\n            return <div>{this.bar}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Foo extends React.Component {\n          render() {\n            let {props:{foo}, bar} = this;\n            return <div>{foo}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Foo extends React.Component {\n          render() {\n            return <div>{this[bar]}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Foo extends React.Component {\n          render() {\n            return <div>{this['bar']}</div>;\n          }\n        }\n      ", ""},
		{"\n        export default (Component) => (\n          class Test extends React.Component {\n            componentDidMount() {}\n            render() {\n              return <Component />;\n            }\n          }\n        );\n      ", ""},
		{"\n        class Foo extends React.Component {\n          render() {\n            return <div>{this.props.children}</div>;\n          }\n        }\n        Foo.childContextTypes = {\n          color: PropTypes.string\n        };\n      ", ""},
		{"\n        @foo\n        class Foo extends React.Component {\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", ""},
		{"\n        @foo(\"bar\")\n        class Foo extends React.Component {\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", ""},
		{"\n        @foo\n        @bar()\n        class Foo extends React.Component {\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ", ""},
		{"\n        class Child extends PureComponent {\n          render() {\n            return <h1>I don't</h1>;\n          }\n        }\n      ", "{\"ignorePureComponents\":true}"},
		{"\n        import React, {PureComponent, PropTypes} from 'react'\n\n        export default function errorDecorator (options) {\n          return WrappedComponent => {\n            class Wrapper extends PureComponent {\n              static propTypes = {\n                error: PropTypes.string\n              }\n              render () {\n                const {error, ...props} = this.props\n                if (error) {\n                  return <div>Error! {error}</div>\n                } else {\n                  return <WrappedComponent {...props} />\n                }\n              }\n            }\n            return Wrapper\n          }\n        }\n      ", "{\"ignorePureComponents\":true}"},
		{"\n        import React, {PureComponent, PropTypes} from 'react'\n\n        export default function errorDecorator (options) {\n          return WrappedComponent =>\n            class Wrapper extends PureComponent {\n              static propTypes = {\n                error: PropTypes.string\n              }\n              render () {\n                const {error, ...props} = this.props\n                if (error) {\n                  return <div>Error! {error}</div>\n                } else {\n                  return <WrappedComponent {...props} />\n                }\n              }\n            }\n        }\n      ", "{\"ignorePureComponents\":true}"},
		{"\n        export default function errorDecorator (options) {\n          return WrappedComponent =>\n            class Wrapper extends React.PureComponent {\n              static propTypes = {\n                error: PropTypes.string\n              }\n              render () {\n                const {error, ...props} = this.props\n                if (error) {\n                  return <div>Error! {error}</div>\n                } else {\n                  return <WrappedComponent {...props} />\n                }\n              }\n            }\n        }\n      ", "{\"ignorePureComponents\":true}"},
		{"\n        /**\n         * @param a.\n         */\n        function Comp() {\n          return <a></a>\n        }\n      ", ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runPreferStatelessFunction(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Beyond the imported corpus. Each verdict below was measured against the installed
// eslint-plugin-react through the Linter API before being written here.

// TestPreferStatelessFunctionOnlyClassesAndFactoriesReport is the test that justifies the whole
// shape of this port.
//
// Upstream wraps this rule in Components.detect, a 959-line registry whose expensive half decides
// whether a plain function is a component. None of that half can produce a finding here, because the
// final filter is isES5Component or isES6Component. If that is wrong, this port is wrong in a way no
// other fixture would show, so it is pinned directly.
func TestPreferStatelessFunctionOnlyClassesAndFactoriesReport(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"an es6 class component", "class Foo extends React.Component { render() { return <div/>; } }", []string{"componentShouldBePure"}},
		{"the bare heritage spelling", "class Foo extends Component { render() { return <div/>; } }", []string{"componentShouldBePure"}},
		{"a class expression", "const Foo = class extends React.Component { render() { return <div/>; } };", []string{"componentShouldBePure"}},
		{"an es5 factory call", "var Foo = createReactClass({ render: function() { return <div/>; } });", []string{"componentShouldBePure"}},
		{"a function declaration is never a candidate", "function Foo() { return <div/>; }", nil},
		{"nor an arrow", "const Foo = () => <div/>;", nil},
		{"nor a function expression", "const Foo = function() { return <div/>; };", nil},
		{"a class with no heritage is not a component", "class Foo { render() { return <div/>; } }", nil},
		{"nor one extending something else", "class Foo extends Bar { render() { return <div/>; } }", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runPreferStatelessFunction(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// TestPreferStatelessFunctionMemberBoundary pins the allowed-member list.
//
// Upstream asks whether ANY member is something other than the five React reads as configuration, so
// this list is the allowed set and anything outside it exempts. Read by NAME rather than by member
// kind, which is why a getter and a static field with the same name give the same answer.
func TestPreferStatelessFunctionMemberBoundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"displayName as a static field", "class Foo extends React.Component { static displayName = 'F'; render() { return <div/>; } }", []string{"componentShouldBePure"}},
		{"displayName as a getter, the same name through a different kind", "class Foo extends React.Component { static get displayName() { return 'F'; } render() { return <div/>; } }", []string{"componentShouldBePure"}},
		{"propTypes", "class Foo extends React.Component { static propTypes = {}; render() { return <div/>; } }", []string{"componentShouldBePure"}},
		{"defaultProps", "class Foo extends React.Component { static defaultProps = {}; render() { return <div/>; } }", []string{"componentShouldBePure"}},
		{"contextTypes", "class Foo extends React.Component { static contextTypes = {}; render() { return <div/>; } }", []string{"componentShouldBePure"}},
		{"any other method exempts", "class Foo extends React.Component { other() {} render() { return <div/>; } }", nil},
		{"any other field exempts", "class Foo extends React.Component { other = 1; render() { return <div/>; } }", nil},
		{"a childContextTypes member exempts, since it is not in the allowed set", "class Foo extends React.Component { childContextTypes = {}; render() { return <div/>; } }", nil},
		{"a bare props member exempts, having no type annotation", "class Foo extends React.Component { props; render() { return <div/>; } }", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runPreferStatelessFunction(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// TestPreferStatelessFunctionThisBoundary pins which uses of `this` disqualify.
//
// The subscript rows are the ones nobody would guess and none of them appears in the corpus in both
// directions. Upstream reads `property.name || property.value`, so an identifier and a string
// subscript both answer with a name while a computed key answers undefined, which is not props.
func TestPreferStatelessFunctionThisBoundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"this.props is exempt", "class Foo extends React.Component { render() { return <div>{this.props.foo}</div>; } }", []string{"componentShouldBePure"}},
		{"a string subscript reads as the same name", "class Foo extends React.Component { render() { return <div>{this['props'].foo}</div>; } }", []string{"componentShouldBePure"}},
		{"this.context is exempt too", "class Foo extends React.Component { render() { return <div>{this.context.foo}</div>; } }", []string{"componentShouldBePure"}},
		{"any other member is a real use of this", "class Foo extends React.Component { render() { return <div>{this.bar}</div>; } }", nil},
		{"including through a string subscript", "class Foo extends React.Component { render() { return <div>{this['bar']}</div>; } }", nil},
		{"and a computed key, whose name cannot be read", "declare const bar: string; class Foo extends React.Component { render() { return <div>{this[bar]}</div>; } }", nil},
		{"destructuring only props and context is exempt", "class Foo extends React.Component { render() { let {props:{foo}, context:{bar}} = this; return <div/>; } }", []string{"componentShouldBePure"}},
		{"one stray name disqualifies the whole pattern", "class Foo extends React.Component { render() { let {props:{foo}, bar} = this; return <div/>; } }", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runPreferStatelessFunction(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// TestPreferStatelessFunctionChildContextTypes pins the branch a first draft wrongly declined.
//
// I wrote this off as needing a scope query and documented it as a divergence. An imported clean
// case caught that, which is exactly why step 5 exists. The branch turns out to be a name
// comparison, and these four rows are the measurements that bound it.
func TestPreferStatelessFunctionChildContextTypes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"an out-of-band declaration naming this component exempts", "class Foo extends React.Component { render() { return <div/>; } }\nFoo.childContextTypes = {};", nil},
		{"naming a different identifier does not", "declare const Bar: any;\nclass Foo extends React.Component { render() { return <div/>; } }\nBar.childContextTypes = {};", []string{"componentShouldBePure"}},
		{"a different property does not", "class Foo extends React.Component { render() { return <div/>; } }\nFoo.somethingElse = {};", []string{"componentShouldBePure"}},
		{"a read rather than a declaration on another object does not", "declare const y: any;\nclass Foo extends React.Component { render() { return <div/>; } }\nconst x = y.childContextTypes;", []string{"componentShouldBePure"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runPreferStatelessFunction(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// TestPreferStatelessFunctionReturnBoundary pins what a render may return.
//
// React 15 and later allow a stateless component to return null, and this tree is on React 19, so
// the null branch is live. Three of upstream's corpus cases exist twice with opposite verdicts on
// exactly this axis; see the note above the silent table.
func TestPreferStatelessFunctionReturnBoundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"jsx", "class Foo extends React.Component { render() { return <div/>; } }", []string{"componentShouldBePure"}},
		{"a fragment", "class Foo extends React.Component { render() { return <></>; } }", []string{"componentShouldBePure"}},
		{"null, allowed since React 15", "class Foo extends React.Component { render() { return null; } }", []string{"componentShouldBePure"}},
		{"false", "class Foo extends React.Component { render() { return false; } }", []string{"componentShouldBePure"}},
		{"a conditional with one qualifying branch", "class Foo extends React.Component { render() { return true ? <div/> : null; } }", []string{"componentShouldBePure"}},
		{"a number is neither jsx nor null", "class Foo extends React.Component { render() { return 42; } }", nil},
		{"a string likewise", "class Foo extends React.Component { render() { return 'x'; } }", nil},
		{"a bare return has no argument at all", "class Foo extends React.Component { render() { return; } }", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runPreferStatelessFunction(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// TestPreferStatelessFunctionDecodesOptions exercises the decoder directly, including the
// unconfigured path.
//
// A rule configured as a bare severity is handed nil rather than a struct, and the port brief
// records a rule that registered on 3,407 files and was completely inert because nothing in its
// suite reached that path. Every fixture above goes through the decoder, so this covers the nil
// input and the wire name.
func TestPreferStatelessFunctionDecodesOptions(t *testing.T) {
	t.Parallel()

	t.Run("absent options give the documented default", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodePreferStatelessFunctionOptions(nil)
		if err != nil {
			t.Fatalf("decoding nil: %v", err)
		}
		if decoded.(PreferStatelessFunctionOptions).IgnorePureComponents {
			t.Fatal("ignorePureComponents defaulted to true, upstream defaults it to false")
		}
	})

	t.Run("the wire name is ignorePureComponents", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodePreferStatelessFunctionOptions(json.RawMessage(`{"ignorePureComponents":true}`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		if !decoded.(PreferStatelessFunctionOptions).IgnorePureComponents {
			t.Fatal("the option did not decode")
		}
	})

	t.Run("the option is what separates these, not the rule", func(t *testing.T) {
		t.Parallel()
		const sourceText = "class Foo extends React.PureComponent { render() { return <div>{this.props.foo}</div>; } }"
		rule_testing.ExpectFindings(t, runPreferStatelessFunction(t, sourceText, ""), "componentShouldBePure")
		rule_testing.ExpectClean(t, runPreferStatelessFunction(t, sourceText, `{"ignorePureComponents":true}`))
	})

	t.Run("ignorePureComponents does not exempt a plain Component", func(t *testing.T) {
		t.Parallel()
		// The distinguishing case for the option's predicate. react.IsEs6ComponentClass accepts
		// Component and PureComponent alike and is the RIGHT predicate for the outer gate and the
		// WRONG one here; routing the option through it would exempt every class component.
		const sourceText = "class Foo extends React.Component { render() { return <div>{this.props.foo}</div>; } }"
		rule_testing.ExpectFindings(t, runPreferStatelessFunction(t, sourceText, `{"ignorePureComponents":true}`), "componentShouldBePure")
	})

	t.Run("the rule falls back to the default when handed nil rather than the struct", func(t *testing.T) {
		t.Parallel()
		// Bypasses the decoder entirely, which is the path every other fixture misses. A
		// PureComponent is the case that measures anything: it reports under the real default and
		// would be clean under a wrong fallback of ignorePureComponents true.
		result := rule_testing.RunWithOptions(t, PreferStatelessFunction, preferStatelessFunctionFile,
			"class Foo extends React.PureComponent { render() { return <div/>; } }", nil)
		rule_testing.ExpectFindings(t, result, "componentShouldBePure")
	})
}

// TestPreferStatelessFunctionSpan asserts WHERE the finding points.
//
// Upstream reports on the component node, which for a factory is the CALL rather than the variable
// and for a class expression is the expression rather than the declaration. ExpectFindings cannot
// see any of this.
func TestPreferStatelessFunctionSpan(t *testing.T) {
	t.Parallel()

	t.Run("a class declaration reports on the whole class", func(t *testing.T) {
		t.Parallel()
		const sourceText = "class Foo extends React.Component { render() { return <div/>; } }"
		result := runPreferStatelessFunction(t, sourceText, "")
		rule_testing.ExpectFindings(t, result, "componentShouldBePure")
		got := result.SourceFile.Text()[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if got != sourceText {
			t.Fatalf("reported on %q", got)
		}
	})

	t.Run("a factory reports on the call, not the declaration", func(t *testing.T) {
		t.Parallel()
		const sourceText = "var Foo = createReactClass({ render: function() { return <div/>; } });"
		const want = "createReactClass({ render: function() { return <div/>; } })"
		result := runPreferStatelessFunction(t, sourceText, "")
		rule_testing.ExpectFindings(t, result, "componentShouldBePure")
		got := result.SourceFile.Text()[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if got != want {
			t.Fatalf("reported on %q, wanted %q", got, want)
		}
	})

	t.Run("the message identifies the rule and explains the defect", func(t *testing.T) {
		t.Parallel()
		result := runPreferStatelessFunction(t, "class Foo extends React.Component { render() { return <div/>; } }", "")
		rule_testing.ExpectFindings(t, result, "componentShouldBePure")
		if result.Diagnostics[0].Message.Id != "componentShouldBePure" {
			t.Fatalf("id was %q", result.Diagnostics[0].Message.Id)
		}
		// Compared against a literal typed here rather than against the rule's own constant,
		// because a comparison to the constant moves with it under mutation.
		const wantPrefix = "This class component uses nothing a class provides."
		if !strings.HasPrefix(result.Diagnostics[0].Message.Description, wantPrefix) {
			t.Fatalf("message was %q", result.Diagnostics[0].Message.Description)
		}
	})
}

// TestPreferStatelessFunctionHasNoFileSuffixGate pins the ABSENCE of a suffix gate.
//
// Three shipped react rules in this package used to decline every file not ending .tsx or .jsx,
// which was oxc residue rather than upstream behavior, and each carried a TEST asserting the gate so
// the suite locked the bug in. All three gates are gone and the helper is deleted.
//
// A .ts file cannot hold JSX, so the JSX-carrying shapes are unreachable there. This uses a render
// returning null, which is legal TypeScript and which upstream reports, so all four suffixes are
// genuinely exercised rather than passing vacuously.
func TestPreferStatelessFunctionHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	const sourceText = "class Foo extends React.Component { render() { return null; } }"
	for _, fileName := range []string{
		"/repository/source/Pure.ts",
		"/repository/source/Pure.tsx",
		"/repository/source/Pure.js",
		"/repository/source/Pure.jsx",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodePreferStatelessFunctionOptions(nil)
			if err != nil {
				t.Fatalf("decoding: %v", err)
			}
			result := rule_testing.RunWithOptions(t, PreferStatelessFunction, fileName, sourceText, decoded)
			rule_testing.ExpectFindings(t, result, "componentShouldBePure")
		})
	}
}
