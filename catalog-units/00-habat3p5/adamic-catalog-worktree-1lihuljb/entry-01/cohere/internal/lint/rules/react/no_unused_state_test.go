package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is upstream's, extracted by stubbing its RuleTester and capturing the case objects the
// test file hands it, so no case was retyped and no escape was typed by hand.
//
// All 103 cases were replayed against the installed build before anything was written down. 101
// reproduced their upstream verdict exactly, with zero verdict mismatches and no substitution of
// any kind. The two that are absent below both fail to PARSE as TypeScript, and both are tagged
// `flow` in upstream's own corpus: they use Flow's type cast expression, `(expr: any)`, which our
// parser cannot produce at all. That is what upstream's `uncast` helper exists for, and it is why
// this port has no counterpart to it. Excluding a case our parser cannot express is a fact about
// the parser rather than a weakening of the rule, and it is recorded here rather than silently
// dropped.
const noUnusedStateFile = "/repository/source/NoUnusedState.tsx"

func TestNoUnusedStateStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{"upstream valid 0", "\n        function StatelessFnUnaffectedTest(props) {\n          return <SomeComponent foo={props.foo} />;\n        };\n      "},
		{"upstream valid 1", "\n        var NoStateTest = createReactClass({\n          render: function() {\n            return <SomeComponent />;\n          }\n        });\n      "},
		{"upstream valid 2", "\n        var NoStateMethodTest = createReactClass({\n          render() {\n            return <SomeComponent />;\n          }\n        });\n      "},
		{"upstream valid 3", "\n        var GetInitialStateTest = createReactClass({\n          getInitialState: function() {\n            return { foo: 0 };\n          },\n          render: function() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        });\n      "},
		{"upstream valid 4", "\n        var ComputedKeyFromVariableTest = createReactClass({\n          getInitialState: function() {\n            return { [foo]: 0 };\n          },\n          render: function() {\n            return <SomeComponent />;\n          }\n        });\n      "},
		{"upstream valid 5", "\n        var ComputedKeyFromBooleanLiteralTest = createReactClass({\n          getInitialState: function() {\n            return { [true]: 0 };\n          },\n          render: function() {\n            return <SomeComponent foo={this.state[true]} />;\n          }\n        });\n      "},
		{"upstream valid 6", "\n        var ComputedKeyFromNumberLiteralTest = createReactClass({\n          getInitialState: function() {\n            return { [123]: 0 };\n          },\n          render: function() {\n            return <SomeComponent foo={this.state[123]} />;\n          }\n        });\n      "},
		{"upstream valid 7", "\n        var ComputedKeyFromExpressionTest = createReactClass({\n          getInitialState: function() {\n            return { [foo + bar]: 0 };\n          },\n          render: function() {\n            return <SomeComponent />;\n          }\n        });\n      "},
		{"upstream valid 8", "\n        var ComputedKeyFromBinaryExpressionTest = createReactClass({\n          getInitialState: function() {\n            return { ['foo' + 'bar' * 8]: 0 };\n          },\n          render: function() {\n            return <SomeComponent />;\n          }\n        });\n      "},
		{"upstream valid 9", "\n        var ComputedKeyFromStringLiteralTest = createReactClass({\n          getInitialState: function() {\n            return { ['foo']: 0 };\n          },\n          render: function() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        });\n      "},
		{"upstream valid 10", "\n        var ComputedKeyFromTemplateLiteralTest = createReactClass({\n          getInitialState: function() {\n            return { [`foo${bar}`]: 0 };\n          },\n          render: function() {\n            return <SomeComponent />;\n          }\n        });\n      "},
		{"upstream valid 11", "\n        var ComputedKeyFromTemplateLiteralTest = createReactClass({\n          getInitialState: function() {\n            return { [`foo`]: 0 };\n          },\n          render: function() {\n            return <SomeComponent foo={this.state['foo']} />;\n          }\n        });\n      "},
		{"upstream valid 12", "\n        var GetInitialStateMethodTest = createReactClass({\n          getInitialState() {\n            return { foo: 0 };\n          },\n          render() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        });\n      "},
		{"upstream valid 13", "\n        var SetStateTest = createReactClass({\n          onFooChange(newFoo) {\n            this.setState({ foo: newFoo });\n          },\n          render() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        });\n      "},
		{"upstream valid 14", "\n        var MultipleSetState = createReactClass({\n          getInitialState() {\n            return { foo: 0 };\n          },\n          update() {\n            this.setState({foo: 1});\n          },\n          render() {\n            return <SomeComponent onClick={this.update} foo={this.state.foo} />;\n          }\n        });\n      "},
		{"upstream valid 15", "\n        class NoStateTest extends React.Component {\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid 16", "\n        class CtorStateTest extends React.Component {\n          constructor() {\n            this.state = { foo: 0 };\n          }\n          render() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        }\n      "},
		{"upstream valid 17", "\n        class ComputedKeyFromVariableTest extends React.Component {\n          constructor() {\n            this.state = { [foo]: 0 };\n          }\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid 18", "\n        class ComputedKeyFromBooleanLiteralTest extends React.Component {\n          constructor() {\n            this.state = { [false]: 0 };\n          }\n          render() {\n            return <SomeComponent foo={this.state['false']} />;\n          }\n        }\n      "},
		{"upstream valid 19", "\n        class ComputedKeyFromNumberLiteralTest extends React.Component {\n          constructor() {\n            this.state = { [345]: 0 };\n          }\n          render() {\n            return <SomeComponent foo={this.state[345]} />;\n          }\n        }\n      "},
		{"upstream valid 20", "\n        class ComputedKeyFromExpressionTest extends React.Component {\n          constructor() {\n            this.state = { [foo + bar]: 0 };\n          }\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid 21", "\n        class ComputedKeyFromBinaryExpressionTest extends React.Component {\n          constructor() {\n            this.state = { [1 + 2 * 8]: 0 };\n          }\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid 22", "\n        class ComputedKeyFromStringLiteralTest extends React.Component {\n          constructor() {\n            this.state = { ['foo']: 0 };\n          }\n          render() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        }\n      "},
		{"upstream valid 23", "\n        class ComputedKeyFromTemplateLiteralTest extends React.Component {\n          constructor() {\n            this.state = { [`foo${bar}`]: 0 };\n          }\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid 24", "\n        class ComputedKeyFromTemplateLiteralTest extends React.Component {\n          constructor() {\n            this.state = { [`foo`]: 0 };\n          }\n          render() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        }\n      "},
		{"upstream valid 25", "\n        class SetStateTest extends React.Component {\n          onFooChange(newFoo) {\n            this.setState({ foo: newFoo });\n          }\n          render() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        }\n      "},
		{"upstream valid 26", "\n        class ClassPropertyStateTest extends React.Component {\n          state = { foo: 0 };\n          render() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        }\n      "},
		{"upstream valid 27", "\n        class OptionalChaining extends React.Component {\n          constructor() {\n            this.state = { foo: 0 };\n          }\n          render() {\n            return <SomeComponent foo={this.state?.foo} />;\n          }\n        }\n      "},
		{"upstream valid 28", "\n        class VariableDeclarationTest extends React.Component {\n          constructor() {\n            this.state = { foo: 0 };\n          }\n          render() {\n            const foo = this.state.foo;\n            return <SomeComponent foo={foo} />;\n          }\n        }\n      "},
		{"upstream valid 29", "\n      class DestructuringTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        render() {\n          const {foo: myFoo} = this.state;\n          return <SomeComponent foo={myFoo} />;\n        }\n      }\n    "},
		{"upstream valid 30", "\n      class ShorthandDestructuringTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        render() {\n          const {foo} = this.state;\n          return <SomeComponent foo={foo} />;\n        }\n      }\n    "},
		{"upstream valid 31", "\n      class AliasDeclarationTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        render() {\n          const state = this.state;\n          return <SomeComponent foo={state.foo} />;\n        }\n      }\n    "},
		{"upstream valid 32", "\n      class AliasAssignmentTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        render() {\n          let state;\n          state = this.state;\n          return <SomeComponent foo={state.foo} />;\n        }\n      }\n    "},
		{"upstream valid 33", "\n      class DestructuringAliasTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        render() {\n          const {state: myState} = this;\n          return <SomeComponent foo={myState.foo} />;\n        }\n      }\n    "},
		{"upstream valid 34", "\n      class ShorthandDestructuringAliasTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        render() {\n          const {state} = this;\n          return <SomeComponent foo={state.foo} />;\n        }\n      }\n    "},
		{"upstream valid 35", "\n      class RestPropertyTest extends React.Component {\n        constructor() {\n          this.state = {\n            foo: 0,\n            bar: 1,\n          };\n        }\n        render() {\n          const {foo, ...others} = this.state;\n          return <SomeComponent foo={foo} bar={others.bar} />;\n        }\n      }\n    "},
		{"upstream valid 36", "\n        class DeepDestructuringTest extends React.Component {\n          state = { foo: 0, bar: 0 };\n          render() {\n            const {state: {foo, ...others}} = this;\n            return <SomeComponent foo={foo} bar={others.bar} />;\n          }\n        }\n      "},
		{"upstream valid 37", "\n      class MethodArgFalseNegativeTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        consumeFoo(foo) {}\n        render() {\n          this.consumeFoo(this.state.foo);\n          return <SomeComponent />;\n        }\n      }\n    "},
		{"upstream valid 38", "\n      class AssignedToObjectFalseNegativeTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        render() {\n          const obj = { foo: this.state.foo, bar: 0 };\n          return <SomeComponent bar={obj.bar} />;\n        }\n      }\n    "},
		{"upstream valid 39", "\n      class ComputedAccessFalseNegativeTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0, bar: 1 };\n        }\n        render() {\n          const bar = 'bar';\n          return <SomeComponent bar={this.state[bar]} />;\n        }\n      }\n    "},
		{"upstream valid 40", "\n      class JsxSpreadFalseNegativeTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        render() {\n          return <SomeComponent {...this.state} />;\n        }\n      }\n    "},
		{"upstream valid 41", "\n      class AliasedJsxSpreadFalseNegativeTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        render() {\n          const state = this.state;\n          return <SomeComponent {...state} />;\n        }\n      }\n    "},
		{"upstream valid 42", "\n      class ObjectSpreadFalseNegativeTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        render() {\n          const attrs = { ...this.state, foo: 1 };\n          return <SomeComponent foo={attrs.foo} />;\n        }\n      }\n    "},
		{"upstream valid 43", "\n      class ShadowingFalseNegativeTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0 };\n        }\n        render() {\n          const state = this.state;\n          let foo;\n          {\n            const state = { foo: 5 };\n            foo = state.foo;\n          }\n          return <SomeComponent foo={foo} />;\n        }\n      }\n    "},
		{"upstream valid 44", "\n      class NonRenderClassMethodFalseNegativeTest extends React.Component {\n        constructor() {\n          this.state = { foo: 0, bar: 0 };\n        }\n        doSomething() {\n          const { foo } = this.state;\n          return this.state.foo;\n        }\n        doSomethingElse() {\n          const { state: { bar }} = this;\n          return bar;\n        }\n        render() {\n          return <SomeComponent />;\n        }\n      }\n    "},
		{"upstream valid 46", "\n        class ArrowFunctionClassMethodDestructuringFalseNegativeTest extends React.Component {\n          constructor() {\n            this.state = { foo: 0 };\n          }\n\n          doSomething = () => {\n            const { state: { foo } } = this;\n\n            return foo;\n          }\n\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid 47", "\n        class ArrowFunctionClassMethodWithClassPropertyTransformFalseNegativeTest extends React.Component {\n          state = { foo: 0 };\n\n          doSomething = () => {\n            const { state:{ foo } } = this;\n\n            return foo;\n          }\n\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid 48", "\n        class ArrowFunctionClassMethodDeepDestructuringFalseNegativeTest extends React.Component {\n          state = { foo: { bar: 0 } };\n\n          doSomething = () => {\n            const { state: { foo: { bar }}} = this;\n\n            return bar;\n          }\n\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid 49", "\n        class ArrowFunctionClassMethodDestructuringAssignmentFalseNegativeTest extends React.Component {\n          state = { foo: 0 };\n\n          doSomething = () => {\n            const { state: { foo: bar }} = this;\n\n            return bar;\n          }\n\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid 50", "\n        class ThisStateAsAnObject extends React.Component {\n          state = {\n            active: true\n          };\n\n          render() {\n            return <div className={classNames('overflowEdgeIndicator', className, this.state)} />;\n          }\n        }\n      "},
		{"upstream valid 51", "\n        class GetDerivedStateFromPropsTest extends Component {\n          constructor(props) {\n            super(props);\n            this.state = {\n              id: 123,\n            };\n          }\n          static getDerivedStateFromProps(nextProps, otherState) {\n            if (otherState.id === nextProps.id) {\n              return {\n                selected: true,\n              };\n            }\n            return null;\n          }\n          render() {\n            return (\n              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>\n            );\n          }\n        }\n      "},
		{"upstream valid 52", "\n        class ComponentDidUpdateTest extends Component {\n          constructor(props) {\n            super(props);\n            this.state = {\n              id: 123,\n            };\n          }\n\n          componentDidUpdate(someProps, someState) {\n            if (someState.id === someProps.id) {\n              doStuff();\n            }\n          }\n          render() {\n            return (\n              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>\n            );\n          }\n        }\n      "},
		{"upstream valid 53", "\n        class ShouldComponentUpdateTest extends Component {\n          constructor(props) {\n            super(props);\n            this.state = {\n              id: 123,\n            };\n          }\n          shouldComponentUpdate(nextProps, nextState) {\n            return nextState.id === nextProps.id;\n          }\n          render() {\n            return (\n              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>\n            );\n          }\n        }\n      "},
		{"upstream valid 54", "\n        class NestedScopesTest extends Component {\n          constructor(props) {\n            super(props);\n            this.state = {\n              id: 123,\n            };\n          }\n          shouldComponentUpdate(nextProps, nextState) {\n            return (function() {\n              return nextState.id === nextProps.id;\n            })();\n          }\n          render() {\n            return (\n              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>\n            );\n          }\n        }\n      "},
		{"upstream valid 55", "\n        class Foo extends Component {\n          state = {\n            initial: 'foo',\n          }\n          handleChange = () => {\n            this.setState(state => ({\n              current: state.initial\n            }));\n          }\n          render() {\n            const { current } = this.state;\n            return <div>{current}</div>\n          }\n        }\n      "},
		{"upstream valid 56", "\n        class Foo extends Component {\n          constructor(props) {\n            super(props);\n            this.state = {\n              initial: 'foo',\n            }\n          }\n          handleChange = () => {\n            this.setState(state => ({\n              current: state.initial\n            }));\n          }\n          render() {\n            const { current } = this.state;\n            return <div>{current}</div>\n          }\n        }\n      "},
		{"upstream valid 57", "\n        class Foo extends Component {\n          constructor(props) {\n            super(props);\n            this.state = {\n              initial: 'foo',\n            }\n          }\n          handleChange = () => {\n            this.setState((state, props) => ({\n              current: state.initial\n            }));\n          }\n          render() {\n            const { current } = this.state;\n            return <div>{current}</div>\n          }\n        }\n      "},
		{"upstream valid 58", "\n        var Foo = createReactClass({\n          getInitialState: function() {\n            return { initial: 'foo' };\n          },\n          handleChange: function() {\n            this.setState(state => ({\n              current: state.initial\n            }));\n          },\n          render() {\n            const { current } = this.state;\n            return <div>{current}</div>\n          }\n        });\n      "},
		{"upstream valid 59", "\n        var Foo = createReactClass({\n          getInitialState: function() {\n            return { initial: 'foo' };\n          },\n          handleChange: function() {\n            this.setState((state, props) => ({\n              current: state.initial\n            }));\n          },\n          render() {\n            const { current } = this.state;\n            return <div>{current}</div>\n          }\n        });\n      "},
		{"upstream valid 60", "\n        class SetStateDestructuringCallback extends Component {\n          state = {\n              used: 1, unused: 2\n          }\n          handleChange = () => {\n            this.setState(({unused}) => ({\n              used: unused * unused,\n            }));\n          }\n          render() {\n            return <div>{this.state.used}</div>\n          }\n        }\n      "},
		{"upstream valid 61", "\n        class SetStateCallbackStateCondition extends Component {\n          state = {\n              isUsed: true,\n              foo: 'foo'\n          }\n          handleChange = () => {\n            this.setState((prevState) => (prevState.isUsed ? {foo: 'bar', isUsed: false} : {}));\n          }\n          render() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        }\n      "},
		{"upstream valid 62", "\n        class Foo extends Component {\n          handleChange = function() {\n            this.setState(() => ({ foo: value }));\n          }\n          render() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        }\n      "},
		{"upstream valid 63", "\n        class Foo extends Component {\n          handleChange = function() {\n            this.setState(state => ({ foo: value }));\n          }\n          render() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        }\n      "},
		{"upstream valid 64", "\n        class Foo extends Component {\n          static handleChange = () => {\n            this.setState(state => ({ foo: value }));\n          }\n          render() {\n            return <SomeComponent foo={this.state.foo} />;\n          }\n        }\n      "},
		{"upstream valid 65", "\n        class Foo extends Component {\n          state = {\n            thisStateAliasProp,\n            thisStateAliasRestProp,\n            thisDestructStateAliasProp,\n            thisDestructStateAliasRestProp,\n            thisDestructStateDestructRestProp,\n            thisSetStateProp,\n            thisSetStateRestProp,\n          } as unknown\n\n          constructor() {\n            // other methods of defining state props\n            ((this as unknown).state as unknown) = { thisStateProp } as unknown;\n            ((this as unknown).setState as unknown)({ thisStateDestructProp } as unknown);\n            ((this as unknown).setState as unknown)(state => ({ thisDestructStateDestructProp } as unknown));\n          }\n\n          thisStateAlias() {\n            const state = (this as unknown).state as unknown;\n\n            (state as unknown).thisStateAliasProp as unknown;\n            const { ...thisStateAliasRest } = state as unknown;\n            (thisStateAliasRest as unknown).thisStateAliasRestProp as unknown;\n          }\n\n          thisDestructStateAlias() {\n            const { state } = this as unknown;\n\n            (state as unknown).thisDestructStateAliasProp as unknown;\n            const { ...thisDestructStateAliasRest } = state as unknown;\n            (thisDestructStateAliasRest as unknown).thisDestructStateAliasRestProp as unknown;\n          }\n\n          thisSetState() {\n            ((this as unknown).setState as unknown)(state => (state as unknown).thisSetStateProp as unknown);\n            ((this as unknown).setState as unknown)(({ ...thisSetStateRest }) => (thisSetStateRest as unknown).thisSetStateRestProp as unknown);\n          }\n\n          render() {\n            ((this as unknown).state as unknown).thisStateProp as unknown;\n            const { thisStateDestructProp } = (this as unknown).state as unknown;\n            const { state: { thisDestructStateDestructProp, ...thisDestructStateDestructRest } } = this as unknown;\n            (thisDestructStateDestructRest as unknown).thisDestructStateDestructRestProp as unknown;\n\n            return null;\n          }\n        }\n      "},
		{"upstream valid 66", "\n        interface Props {}\n\n        interface State {\n          flag: boolean;\n        }\n\n        export default class RuleTest extends React.Component<Props, State> {\n          readonly state: State = {\n            flag: false,\n          };\n\n          static getDerivedStateFromProps = (props: Props, state: State) => {\n            const newState: Partial<State> = {};\n            if (!state.flag) {\n              newState.flag = true;\n            }\n            return newState;\n          };\n        }\n      "},
		{"upstream valid 67", "\n        class Foo extends React.Component {\n          onCancel = (data) => {\n            console.log('Cancelled', data)\n            this.setState({ status: 'Cancelled. Try again?' })\n          }\n          render() {\n            const { status } = this.state;\n            return <div>{status}</div>\n          }\n        }\n      "},
		{"upstream valid 68", "\n        class KarmaRefundPillComponent extends GenericPillComponent {\n          renderContent = () => {\n            const { action } = this.props\n\n            return (\n              <Box fontSize={[1]} mx={[2]} minWidth=\"10px\" minHeight=\"26px\" alignItems=\"center\">\n                <FormattedText\n                  fields={getKarmaClaimLevel1Fields(action)}\n                  i18nKey=\"pillTemplates.karmarefund.summary\"\n                  fontSize={[1]}\n                />\n              </Box>\n            )\n          }\n        }\n      "},
		{"upstream valid 69", "\n        class AutoControlledComponent<P = {}, S = {}> extends UIComponent<P, S> {\n          static getDerivedStateFromProps: React.GetDerivedStateFromProps<any, any>\n        }\n      "},
		{"upstream valid 70", "\n        export const commonMixinWrapper = ComposeComponent => class extends ComposeComponent {\n          static getDerivedStateFromProps = ComposeComponent.getDerivedStateFromProps;\n          render() { return <div />; }\n        }\n      "},
		{"upstream valid 71", "\n        import React, { PureComponent } from 'react';\n\n        class TestNoUnusedState extends React.Component {\n          constructor(props) {\n            super(props);\n            this.state = {\n              id: null,\n            };\n          }\n\n          static getDerivedStateFromProps = (props, state) => {\n            if (state.id !== props.id) {\n              return {\n                id: props.id,\n              };\n            }\n\n            return null;\n          };\n\n          render() {\n            return <h1>{this.state.id}</h1>;\n          }\n        }\n\n        export default TestNoUnusedState;\n      "},
		{"upstream valid 72", "\n        class Component extends React.Component {\n          static getDerivedStateFromProps = ({value, disableAnimation}: ToggleProps, {isControlled, isOn}: ToggleState) => {\n            return { isControlled, isOn };\n          };\n\n          render() {\n            const { isControlled, isOn } = this.state;\n            return <div>{isControlled ? 'controlled' : ''}{isOn ? 'on' : ''}</div>;\n          }\n        }\n      "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestNoUnusedStateFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		messageIds       []string
	}{
		{"upstream invalid 0", "\n        var UnusedGetInitialStateTest = createReactClass({\n          getInitialState: function() {\n            return { foo: 0 };\n          },\n          render: function() {\n            return <SomeComponent />;\n          }\n        })\n      ", []string{"unusedStateField"}},
		{"upstream invalid 1", "\n        var UnusedComputedStringLiteralKeyStateTest = createReactClass({\n          getInitialState: function() {\n            return { ['foo']: 0 };\n          },\n          render: function() {\n            return <SomeComponent />;\n          }\n        })\n      ", []string{"unusedStateField"}},
		{"upstream invalid 2", "\n        var UnusedComputedTemplateLiteralKeyStateTest = createReactClass({\n          getInitialState: function() {\n            return { [`foo`]: 0 };\n          },\n          render: function() {\n            return <SomeComponent />;\n          }\n        })\n      ", []string{"unusedStateField"}},
		{"upstream invalid 3", "\n        var UnusedComputedNumberLiteralKeyStateTest = createReactClass({\n          getInitialState: function() {\n            return { [123]: 0 };\n          },\n          render: function() {\n            return <SomeComponent />;\n          }\n        })\n      ", []string{"unusedStateField"}},
		{"upstream invalid 4", "\n        var UnusedComputedBooleanLiteralKeyStateTest = createReactClass({\n          getInitialState: function() {\n            return { [true]: 0 };\n          },\n          render: function() {\n            return <SomeComponent />;\n          }\n        })\n      ", []string{"unusedStateField"}},
		{"upstream invalid 5", "\n        var UnusedGetInitialStateMethodTest = createReactClass({\n          getInitialState() {\n            return { foo: 0 };\n          },\n          render() {\n            return <SomeComponent />;\n          }\n        })\n      ", []string{"unusedStateField"}},
		{"upstream invalid 6", "\n        var UnusedSetStateTest = createReactClass({\n          onFooChange(newFoo) {\n            this.setState({ foo: newFoo });\n          },\n          render() {\n            return <SomeComponent />;\n          }\n        });\n      ", []string{"unusedStateField"}},
		{"upstream invalid 7", "\n        class UnusedCtorStateTest extends React.Component {\n          constructor() {\n            this.state = { foo: 0 };\n          }\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 8", "\n        class UnusedSetStateTest extends React.Component {\n          onFooChange(newFoo) {\n            this.setState({ foo: newFoo });\n          }\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 9", "\n        class UnusedClassPropertyStateTest extends React.Component {\n          state = { foo: 0 };\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 10", "\n        class UnusedComputedStringLiteralKeyStateTest extends React.Component {\n          state = { ['foo']: 0 };\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 11", "\n        class UnusedComputedTemplateLiteralKeyStateTest extends React.Component {\n          state = { [`foo`]: 0 };\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 12", "\n        class UnusedComputedTemplateLiteralKeyStateTest extends React.Component {\n          state = { [`foo \\n bar`]: 0 };\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 13", "\n        class UnusedComputedBooleanLiteralKeyStateTest extends React.Component {\n          state = { [true]: 0 };\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 14", "\n        class UnusedComputedNumberLiteralKeyStateTest extends React.Component {\n          state = { [123]: 0 };\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 15", "\n        class UnusedComputedFloatLiteralKeyStateTest extends React.Component {\n          state = { [123.12]: 0 };\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 16", "\n        class UnusedStateWhenPropsAreSpreadTest extends React.Component {\n          constructor() {\n            this.state = { foo: 0 };\n          }\n          render() {\n            return <SomeComponent {...this.props} />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 17", "\n        class AliasOutOfScopeTest extends React.Component {\n          constructor() {\n            this.state = { foo: 0 };\n          }\n          render() {\n            const state = this.state;\n            return <SomeComponent />;\n          }\n          someMethod() {\n            const outOfScope = state.foo;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 18", "\n        class MultipleErrorsTest extends React.Component {\n          constructor() {\n            this.state = {\n              foo: 0,\n              bar: 1,\n              baz: 2,\n              qux: 3,\n            };\n          }\n          render() {\n            let {state} = this;\n            return <SomeComponent baz={state.baz} qux={state.qux} />;\n          }\n        }\n      ", []string{"unusedStateField", "unusedStateField"}},
		{"upstream invalid 19", "\n        class MultipleErrorsForSameKeyTest extends React.Component {\n          constructor() {\n            this.state = { foo: 0 };\n          }\n          onFooChange(newFoo) {\n            this.setState({ foo: newFoo });\n          }\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      ", []string{"unusedStateField", "unusedStateField"}},
		{"upstream invalid 20", "\n        class UnusedRestPropertyFieldTest extends React.Component {\n          constructor() {\n            this.state = {\n              foo: 0,\n              bar: 1,\n            };\n          }\n          render() {\n            const {bar, ...others} = this.state;\n            return <SomeComponent bar={bar} />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 21", "\n        class UnusedStateArrowFunctionMethodTest extends React.Component {\n          constructor() {\n            this.state = { foo: 0 };\n          }\n          doSomething = () => {\n            return null;\n          }\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 23", "\n        class UnusedDeepDestructuringTest extends React.Component {\n          state = { foo: 0, bar: 0 };\n          render() {\n            const {state: {foo}} = this;\n            return <SomeComponent foo={foo} />;\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 24", "\n        class FakePrevStateVariableTest extends Component {\n          constructor(props) {\n            super(props);\n            this.state = {\n              id: 123,\n              foo: 456\n            };\n          }\n\n          componentDidUpdate(someProps, someState) {\n            if (someState.id === someProps.id) {\n              const prevState = { foo: 789 };\n              console.log(prevState.foo);\n            }\n          }\n          render() {\n            return (\n              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>\n            );\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 25", "\n        class UseStateParameterOfNonLifecycleTest extends Component {\n          constructor(props) {\n            super(props);\n            this.state = {\n              foo: 123,\n            };\n          }\n          nonLifecycle(someProps, someState) {\n            doStuff(someState.foo)\n          }\n          render() {\n            return (\n              <SomeComponent />\n            );\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 26", "\n        class MissingStateParameterTest extends Component {\n          constructor(props) {\n            super(props);\n            this.state = {\n              id: 123\n            };\n          }\n\n          componentDidUpdate(someProps) {\n            const prevState = { id: 456 };\n            console.log(prevState.id);\n          }\n          render() {\n            return (\n              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>\n            );\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 27", "\n        class Foo extends Component {\n          state = {\n            initial: 'foo',\n          }\n          handleChange = () => {\n            this.setState(() => ({\n              current: 'hi'\n            }));\n          }\n          render() {\n            const { current } = this.state;\n            return <div>{current}</div>\n          }\n        }\n      ", []string{"unusedStateField"}},
		{"upstream invalid 28", "\n        wrap(class NotWorking extends React.Component {\n            state = {\n                dummy: null\n            };\n        });\n      ", []string{"unusedStateField"}},
		{"upstream invalid 29", "\n        class Foo extends Component {\n          state = {\n            thisStateAliasPropUnused,\n            thisStateAliasRestPropUnused,\n            thisDestructStateAliasPropUnused,\n            thisDestructStateAliasRestPropUnused,\n            thisDestructStateDestructRestPropUnused,\n            thisSetStatePropUnused,\n            thisSetStateRestPropUnused,\n          } as unknown\n\n          constructor() {\n            // other methods of defining state props\n            ((this as unknown).state as unknown) = { thisStatePropUnused } as unknown;\n            ((this as unknown).setState as unknown)({ thisStateDestructPropUnused } as unknown);\n            ((this as unknown).setState as unknown)(state => ({ thisDestructStateDestructPropUnused } as unknown));\n          }\n\n          thisStateAlias() {\n            const state = (this as unknown).state as unknown;\n\n            (state as unknown).thisStateAliasProp as unknown;\n            const { ...thisStateAliasRest } = state as unknown;\n            (thisStateAliasRest as unknown).thisStateAliasRestProp as unknown;\n          }\n\n          thisDestructStateAlias() {\n            const { state } = this as unknown;\n\n            (state as unknown).thisDestructStateAliasProp as unknown;\n            const { ...thisDestructStateAliasRest } = state as unknown;\n            (thisDestructStateAliasRest as unknown).thisDestructStateAliasRestProp as unknown;\n          }\n\n          thisSetState() {\n            ((this as unknown).setState as unknown)(state => (state as unknown).thisSetStateProp as unknown);\n            ((this as unknown).setState as unknown)(({ ...thisSetStateRest }) => (thisSetStateRest as unknown).thisSetStateRestProp as unknown);\n          }\n\n          render() {\n            ((this as unknown).state as unknown).thisStateProp as unknown;\n            const { thisStateDestructProp } = (this as unknown).state as unknown;\n            const { state: { thisDestructStateDestructProp, ...thisDestructStateDestructRest } } = this as unknown;\n            (thisDestructStateDestructRest as unknown).thisDestructStateDestructRestProp as unknown;\n\n            return null;\n          }\n        }\n      ", []string{"unusedStateField", "unusedStateField", "unusedStateField", "unusedStateField", "unusedStateField", "unusedStateField", "unusedStateField", "unusedStateField", "unusedStateField", "unusedStateField"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestNoUnusedStateGiveUpConditions pins the four inputs that abandon the whole class.
//
// This is the rule's most important behaviour and the reason people keep it enabled. When an access
// cannot be resolved to a key, upstream reports NOTHING for that component rather than guessing.
// Reporting a field that is genuinely read through a computed access is a false positive nobody can
// act on, because the read is visible in the source and the rule simply cannot see it.
//
// All four measured against the installed build on 2026-08-27, each beside a control that reports.
func TestNoUnusedStateGiveUpConditions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"the control reports when nothing obscures the reads",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return null; } }",
			1,
		},
		{
			"a computed access with a non literal key gives up",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return this.state[key]; } }",
			0,
		},
		{
			"the state object handed to a call gives up",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return f(this.state); } }",
			0,
		},
		{
			"spreading state as JSX attributes gives up",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return <D {...this.state} />; } }",
			0,
		},
		{
			"spreading state into an object gives up",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return { ...this.state }; } }",
			0,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "unusedStateField")
		})
	}
}

// TestNoUnusedStateComputedLiteralIsARead separates a give-up from an ordinary read.
//
// A computed access is only a give-up when its key is not a literal. `this.state["a"]` reads `a`,
// and the third row proves the key is read correctly rather than merely suppressing the report:
// reading `"b"` still reports `a`. Without that row a rule that treated every computed access as a
// blanket give-up would pass the first two.
func TestNoUnusedStateComputedLiteralIsARead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"a computed string literal reads that key",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return this.state[\"a\"]; } }",
			0,
		},
		{
			"a computed non literal gives up",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return this.state[key]; } }",
			0,
		},
		{
			"a computed literal naming a different key still reports the unread one",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return this.state[\"b\"]; } }",
			1,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "unusedStateField")
		})
	}
}

// TestNoUnusedStateDestructuringIsARead pins the shape a port written from ESTree gets wrong.
//
// Our parser gives a shorthand `{foo}` a binding element whose PropertyName is nil and whose Name is
// the identifier, while a renaming `{foo: bar}` fills both. ESTree carries `key` and `value`
// separately in both cases. So the property being read is PropertyName when present and Name
// otherwise, and reading Name unconditionally would record the LOCAL name of a renamed binding.
//
// The renaming row is the one that separates the two designs: reading Name would record `z` rather
// than `a`, so `a` would report as unused. Measured silent upstream.
func TestNoUnusedStateDestructuringIsARead(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{
			"a shorthand destructuring is a read",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ const { a } = this.state; return a; } }",
		},
		{
			"a renaming destructuring reads the PROPERTY, not the local name",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ const { a: z } = this.state; return z; } }",
		},
		{
			"a rest element aliases the whole state object",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ const { ...rest } = this.state; return rest.a; } }",
		},
		{
			"destructuring state off this makes an alias",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ const { state } = this; return state.a; } }",
		},
		{
			"a nested pattern over this reads the keys directly",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ const { state: { a } } = this; return a; } }",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUnusedStateDoesNotPanicOnBindingPatterns is a crash guard, not a behaviour fixture.
//
// `Node.Text()` PANICS on a binding pattern rather than returning a name, measured directly in a
// probe: an object pattern and an array pattern both panic while a plain identifier is safe. The
// walk recovers per FILE rather than per rule, so one such panic costs every rule in the tree its
// verdict on that file. That is the 167-file class this project has already paid for once.
//
// Every place this rule reads a name off a binding is kind-guarded. These rows drive each of them
// with a pattern in the position a name is expected. No `ExpectFindings` assertion can see a panic,
// so the assertion here is simply that the run completes.
func TestNoUnusedStateDoesNotPanicOnBindingPatterns(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{
			"an array pattern as the setState updater parameter",
			"class C extends React.Component { f(){ this.setState(([a]) => ({ b: a })); } render(){ return null; } }",
		},
		{
			"an array pattern destructuring state",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ const [ x ] = this.state; return x; } }",
		},
		{
			"a destructured second parameter on a lifecycle method",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } componentDidUpdate(prev, { a }){ return a; } render(){ return null; } }",
		},
		{
			"a destructured second parameter on getDerivedStateFromProps",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } static getDerivedStateFromProps(props, { a }){ return null; } render(){ return null; } }",
		},
		{
			"a nested pattern inside a rest binding",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ const { state: [ x ] } = this; return x; } }",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Reaching the end of this subtest is the assertion. A panic in the rule would fail the
			// test rather than being swallowed, because the harness does not recover.
			rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
		})
	}
}

// TestNoUnusedStateUpdaterAndLifecycleParameters pins the two parameter routes.
//
// Both carry a decision that reads as an oversight and is reproduced rather than corrected.
//
// The updater alias only applies to an ARROW function: upstream tests that node type specifically,
// so `this.setState(function(prev) {...})` aliases nothing and its reads are lost. Measured, the
// arrow form is silent and the function-expression form reports.
//
// The lifecycle list is EXACT. A method with the same two-parameter shape and a different name does
// not alias its second parameter, so a read through it does not count. Measured.
func TestNoUnusedStateUpdaterAndLifecycleParameters(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"an arrow updater parameter is an alias",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } f(){ this.setState(prev => ({ b: prev.a })); } render(){ return this.state.b; } }",
			0,
		},
		{
			"a function expression updater parameter is NOT an alias",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } f(){ this.setState(function(prev){ return { b: prev.a }; }); } render(){ return this.state.b; } }",
			1,
		},
		{
			"a destructured updater parameter reads its keys",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } f(){ this.setState(({a}) => ({ b: a })); } render(){ return this.state.b; } }",
			0,
		},
		{
			"the second parameter of componentDidUpdate is state",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } componentDidUpdate(prevProps, prevState){ return prevState.a; } render(){ return null; } }",
			0,
		},
		{
			"the second parameter of shouldComponentUpdate is state",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } shouldComponentUpdate(nextProps, nextState){ return nextState.a; } render(){ return null; } }",
			0,
		},
		{
			"the second parameter of static getDerivedStateFromProps is state",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } static getDerivedStateFromProps(props, state){ return { b: state.a }; } render(){ return this.state.b; } }",
			0,
		},
		{
			"a method that is not a lifecycle hook does NOT alias its second parameter",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } notLifecycle(x, y){ return y.a; } render(){ return null; } }",
			1,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "unusedStateField")
		})
	}
}

// TestNoUnusedStateLiteralKeyKinds pins which computed keys have a static name.
//
// Upstream renders a key with `String(node.value)` over an estree Literal, and an estree Literal
// covers more than a string and a number. Our parser gives `true`, `false` and `null` their own
// keyword kinds, so a port that stopped at string and numeric goes silent on two of upstream's OWN
// failing cases, which is how this was found rather than reasoned about.
//
// All five measured against the installed build on 2026-08-27.
func TestNoUnusedStateLiteralKeyKinds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"a computed boolean literal key is named true",
			"class C extends React.Component { state = { [true]: 0 }; render(){ return null; } }",
			1,
		},
		{
			"a computed null literal key is named null",
			"class C extends React.Component { state = { [null]: 0 }; render(){ return null; } }",
			1,
		},
		{
			"a computed template with no substitutions is named by its text",
			"class C extends React.Component { state = { [`a`]: 0 }; render(){ return null; } }",
			1,
		},
		{
			"a computed template is also readable",
			"class C extends React.Component { state = { [`a`]: 0 }; render(){ return this.state[`a`]; } }",
			0,
		},
		{
			"a computed identifier key has no static name and is not recorded",
			"class C extends React.Component { state = { [k]: 0 }; render(){ return null; } }",
			0,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "unusedStateField")
		})
	}
}

// TestNoUnusedStateTypeAssertionsAreUnwrapped pins the six positions where a TypeScript assertion
// changes the verdict.
//
// Upstream strips its assertion node in eight places; ours has to strip a PARENTHESIS as well,
// because our parser keeps `KindParenthesizedExpression` as a real node and the tree upstream walks
// folds it away. `(this as unknown).state` therefore arrives as an access whose object is a
// parenthesized assertion, two wrappers deep.
//
// Two of these were found only by a corpus case failing rather than by reading: the assertion on an
// updater's arrow BODY, and the assertion on the setState RECEIVER inside an already-unwrapped
// callee. Each row was measured reporting against the installed build.
func TestNoUnusedStateTypeAssertionsAreUnwrapped(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{
			"an assertion on the assigned object",
			"class C extends Component { constructor(){ this.state = { a: 1 } as unknown; } render(){ return null; } }",
		},
		{
			"an assertion on the assignment target and its receiver",
			"class C extends Component { constructor(){ ((this as unknown).state as unknown) = { a: 1 }; } render(){ return null; } }",
		},
		{
			"an assertion on a setState argument",
			"class C extends Component { constructor(){ this.setState({ a: 1 } as unknown); } render(){ return null; } }",
		},
		{
			"an assertion on the setState callee and its receiver",
			"class C extends Component { constructor(){ ((this as unknown).setState as unknown)({ a: 1 }); } render(){ return null; } }",
		},
		{
			"an assertion on an updater arrow body",
			"class C extends Component { constructor(){ this.setState(s => ({ a: 1 } as unknown)); } render(){ return null; } }",
		},
		{
			"an assertion on a class property initializer",
			"class C extends Component { state = { a: 1 } as unknown; render(){ return null; } }",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unusedStateField")
		})
	}
}

// TestNoUnusedStateSpanAndMessage asserts where the finding points and what it says.
//
// The finding anchors on the PROPERTY that declared the field rather than on the class or the
// assignment, so a reader is shown the declaration to delete. The key name is interpolated, so the
// text is asserted in full against a literal typed here rather than against the rule's own builder,
// which would move both sides together under mutation.
func TestNoUnusedStateSpanAndMessage(t *testing.T) {
	t.Parallel()

	source := "class C extends React.Component { constructor(p){ super(p); this.state = { alpha: 1 }; } render(){ return null; } }"
	result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "alpha: 1" {
		t.Fatalf("finding points at %q, wanted %q", reported, "alpha: 1")
	}
	want := "Unused state field: 'alpha'. Nothing in this component reads it, so every write to " +
		"it re-renders for a value no one looks at, and a reader has to search the whole class " +
		"before they can be sure it is dead."
	if result.Diagnostics[0].Message.Description != want {
		t.Fatalf("message is %q", result.Diagnostics[0].Message.Description)
	}
	if result.Diagnostics[0].Message.Id != "unusedStateField" {
		t.Fatalf("id is %q", result.Diagnostics[0].Message.Id)
	}
	// Upstream marks this rule not fixable and neither does this port: deleting a state field is
	// not a spelling change, since the write that set it may have a side effect.
	if len(result.Diagnostics[0].Fixes) != 0 || len(result.Diagnostics[0].Suggestions) != 0 {
		t.Fatal("this rule must propose no repair")
	}
}

// TestNoUnusedStateFindingOrderIsSourceOrder pins that findings come out in declaration order.
//
// The accumulator is a slice rather than a map for exactly this reason. A map would make the order
// depend on Go's randomisation, which produces a test that passes most of the time and is therefore
// worse than no test.
func TestNoUnusedStateFindingOrderIsSourceOrder(t *testing.T) {
	t.Parallel()

	source := "class C extends React.Component { state = { zeta: 1, alpha: 2, mid: 3 }; render(){ return null; } }"
	result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, source)
	rule_testing.ExpectFindings(t, result, "unusedStateField", "unusedStateField", "unusedStateField")
	want := []string{"zeta", "alpha", "mid"}
	for index, name := range want {
		reported := source[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
		if !strings.HasPrefix(reported, name) {
			t.Fatalf("finding %d points at %q, wanted one starting %q", index, reported, name)
		}
	}
}

// TestNoUnusedStateWhichComponentsAreJudged pins the component gate.
//
// A class extending a React base, in either the declaration or the expression spelling, and an
// object handed to the `createReactClass` factory. `React.createClass` is silent because the factory
// name comes from a setting whose default is `createReactClass` and upstream's own harness never
// sets it, which is the same measurement this package's `no-access-state-in-setstate` records.
func TestNoUnusedStateWhichComponentsAreJudged(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"a class extending React.Component",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return null; } }",
			1,
		},
		{
			"a class extending React.PureComponent",
			"class C extends React.PureComponent { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return null; } }",
			1,
		},
		{
			"a class extending a bare Component",
			"class C extends Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return null; } }",
			1,
		},
		{
			"a class expression",
			"var C = class extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return null; } };",
			1,
		},
		{
			"a class extending something else is not a component",
			"class C extends Other { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return null; } }",
			0,
		},
		{
			"the createReactClass factory",
			"var C = createReactClass({ getInitialState: function(){ return { a: 1 }; }, render: function(){ return null; } });",
			1,
		},
		{
			"the factory with getInitialState written as a method shorthand",
			"var C = createReactClass({ getInitialState() { return { a: 1 }; }, render() { return null; } });",
			1,
		},
		{
			"React.createClass is not the configured factory name",
			"var C = React.createClass({ getInitialState: function(){ return { a: 1 }; }, render: function(){ return null; } });",
			0,
		},
		{
			"a getInitialState on a plain object contributes nothing",
			"var o = { getInitialState: function(){ return { a: 1 }; } };",
			0,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "unusedStateField")
		})
	}
}

// TestNoUnusedStateStateAssignmentOnlyCountsInTheConstructor pins where `this.state = {}` is read.
//
// Upstream climbs to the nearest enclosing function and records the fields only when that function
// is the constructor, so the same assignment written in another method contributes nothing.
// Measured both ways against the installed build.
func TestNoUnusedStateStateAssignmentOnlyCountsInTheConstructor(t *testing.T) {
	t.Parallel()

	inConstructor := rule_testing.Run(t, NoUnusedState, noUnusedStateFile,
		"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } render(){ return null; } }")
	rule_testing.ExpectFindings(t, inConstructor, "unusedStateField")

	inAnotherMethod := rule_testing.Run(t, NoUnusedState, noUnusedStateFile,
		"class C extends React.Component { f(){ this.state = { a: 1 }; } render(){ return null; } }")
	rule_testing.ExpectClean(t, inAnotherMethod)
}

// TestNoUnusedStateLifecycleParameterGates pins three conditions that survived the sweep.
//
// Each distinguishing input was measured against the installed build before the row was written,
// and each is a shape upstream's corpus does not contain.
//
// The `getDerivedStateFromProps` hook only receives state when it is STATIC, so a non-static method
// of the same name and shape does not alias its second parameter. And both it and the ordinary
// lifecycle methods need at least two parameters before the second one exists to alias; a
// one-parameter hook reads nothing.
//
// All three rows report upstream, which means each guard is what keeps the port from being wider
// than the rule it reproduces.
func TestNoUnusedStateLifecycleParameterGates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"static getDerivedStateFromProps aliases its second parameter",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } static getDerivedStateFromProps(props, state){ return { b: state.a }; } render(){ return this.state.b; } }",
			0,
		},
		{
			"a NON static method of the same name does not",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } getDerivedStateFromProps(props, state){ return { b: state.a }; } render(){ return this.state.b; } }",
			1,
		},
		{
			"a one parameter getDerivedStateFromProps has no state parameter",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } static getDerivedStateFromProps(props){ return null; } render(){ return null; } }",
			1,
		},
		{
			"a one parameter lifecycle method has no state parameter",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } componentDidUpdate(prevProps){ return prevProps.a; } render(){ return null; } }",
			1,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "unusedStateField")
		})
	}
}

// TestNoUnusedStateDerivedStateAsAClassProperty covers the OTHER spelling of the same hook.
//
// This test exists because two mutants survived after fixtures written for them, which the port
// brief says means the hypothesis was wrong rather than the input. It was: upstream handles
// `getDerivedStateFromProps` through TWO separate paths, and so does this port. The METHOD spelling
// resolves through the lifecycle parameter walk, while the CLASS PROPERTY spelling has its own arm
// that reads the parameter's uses directly. The rows above only ever drove the first one, so the
// second one's three guards were never reached by any fixture.
//
// All four measured against the installed build on 2026-08-27. The static requirement, the two
// parameter minimum, and the name comparison are each load-bearing on this path.
func TestNoUnusedStateDerivedStateAsAClassProperty(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"a static property holding the hook aliases its second parameter",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } static getDerivedStateFromProps = (props, state) => ({ b: state.a }); render(){ return this.state.b; } }",
			0,
		},
		{
			"a NON static property of the same name does not",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } getDerivedStateFromProps = (props, state) => ({ b: state.a }); render(){ return this.state.b; } }",
			1,
		},
		{
			"a one parameter property hook has no state parameter",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } static getDerivedStateFromProps = (props) => null; render(){ return null; } }",
			1,
		},
		{
			"a static property with a different name is not the hook",
			"class C extends React.Component { constructor(p){ super(p); this.state = { a: 1 }; } static other = (props, state) => ({ b: state.a }); render(){ return this.state.b; } }",
			1,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "unusedStateField")
		})
	}
}

// TestNoUnusedStateGetInitialStateReadsOnlyTheLastStatement pins a real under-report in upstream.
//
// Upstream reads the LAST statement of `getInitialState` and looks for a return of an object
// literal there. A function returning early therefore contributes only the fields of its final
// return, and the early object's keys are never recorded as written at all.
//
// Measured: a body with an early `{early: 1}` and a final `{late: 1}` reports only `late`. So the
// early field is not reported as unused, which is an under-report rather than a false positive, and
// it is reproduced rather than corrected because widening it would report on code upstream calls
// clean.
//
// A mutation reading the FIRST statement instead survived the whole imported corpus, because every
// one of upstream's factory cases has a single-statement body.
func TestNoUnusedStateGetInitialStateReadsOnlyTheLastStatement(t *testing.T) {
	t.Parallel()

	earlyReturn := rule_testing.Run(t, NoUnusedState, noUnusedStateFile,
		"var C = createReactClass({ getInitialState: function(){ if (x) { return { early: 1 }; } return { late: 1 }; }, render: function(){ return null; } });")
	rule_testing.ExpectFindings(t, earlyReturn, "unusedStateField")
	reported := earlyReturn.Diagnostics[0].Message.Description
	if !strings.Contains(reported, "'late'") {
		t.Fatalf("the finding names %q, wanted the field from the LAST return", reported)
	}
}

// TestNoUnusedStateClassPropertyGates pins which class property declares state.
//
// The name has to be `state` and the property must not be static. A mutation removing the name
// comparison survived every other fixture here, because nothing else in this file writes a class
// property with a different name holding an object. Both distinguishing rows measured silent
// against the installed build.
//
// The static row matters separately: `static state = {}` is a class-level constant rather than
// component state, and reporting its keys as unused state would be wrong about what the code means.
func TestNoUnusedStateClassPropertyGates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"a property named state declares state",
			"class C extends React.Component { state = { a: 1 }; render(){ return null; } }",
			1,
		},
		{
			"a property with another name does not",
			"class C extends React.Component { notState = { a: 1 }; render(){ return null; } }",
			0,
		},
		{
			"a STATIC property named state does not",
			"class C extends React.Component { static state = { a: 1 }; render(){ return null; } }",
			0,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "unusedStateField")
		})
	}
}

// TestNoUnusedStateReceiverMustBeThis pins that only `this.state` is this component's state.
//
// A mutation dropping the receiver check survived every other fixture here, because nothing else in
// this file writes `something.state` on a receiver that is not `this`. Both rows measured against
// the installed build, and the direction matters: without the check a read of ANOTHER object's
// `state` would silence this rule, which is a false negative rather than a false positive and
// therefore invisible in a findings count.
func TestNoUnusedStateReceiverMustBeThis(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"reading this.state is a read",
			"class C extends React.Component { state = { a: 1 }; render(){ return this.state.a; } }",
			0,
		},
		{
			"reading another object's state is NOT a read",
			"class C extends React.Component { state = { a: 1 }; render(){ return other.state.a; } }",
			1,
		},
		{
			"calling setState on another object does not declare state",
			"class C extends React.Component { state = { a: 1 }; f(){ other.setState({ b: 1 }); } render(){ return null; } }",
			1,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedState, noUnusedStateFile, testCase.sourceText)
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "unusedStateField")
		})
	}
}
