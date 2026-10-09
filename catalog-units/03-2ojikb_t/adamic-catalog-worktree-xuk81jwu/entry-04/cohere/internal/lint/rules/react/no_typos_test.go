package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noTyposFile is where the fixtures pretend to live.
//
// A `.tsx` name because several cases carry JSX. There is no file gate, and
// `TestNoTyposHasNoFileGate` reports in a `.ts` file too.
const noTyposFile = "/repository/source/NoTypos.tsx"

// The corpus is `eslint-plugin-react`'s own, extracted rather than retyped.
//
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/no-typos.js` holds 43 valid and 53 invalid
// cases. Every string below was pulled out of that file by loading it with a stubbed `RuleTester`
// and serializing the captured object to JSON, then emitted into these tables by a generator. No
// case was typed by hand.
//
// All 96 were run against the installed build, 7.37.5, through the ESLint Linter API with the
// TypeScript parser, and the two authorities agreed on 95.
//
// # The one case they disagree about is held back, and the cause is an ESLint version
//
// `/** @extends React.Component */ class MyComponent extends BaseComponent {}` followed by
// `MyComponent.PROPTYPES = {}` is asserted as reporting in the clone and is SILENT on the installed
// build. The cause is `componentUtil.isExplicitComponent`, which calls `sourceCode.getJSDocComment`;
// that method was removed in ESLint 9, so the branch cannot fire any more. Measured under both the
// TypeScript parser and the default one, with a control confirming the same class reports when it
// extends `React.Component` for real. Recorded in `TestNoTyposJsDocComponentAnnotationIsNotHonored`
// rather than dropped, so the whole corpus stays accounted for.
//
// # The invalid table asserts the rendered text as well as the id
//
// Four of the eight ids interpolate a name, and two of those interpolate two, so an id assertion
// alone cannot see a message naming the wrong thing. The fourth column is the exact string the
// installed build rendered, captured in the same run that verified the ids, and it is asserted as a
// substring because this port rewrites the wording while keeping the interpolated values.

// TestNoTyposFires runs upstream's reporting cases.
func TestNoTyposFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
		wantTexts  []string
	}{
		{"upstream invalid-0", "\n        class Component extends React.Component {\n          static PropTypes = {};\n        }\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-1", "\n        class Component extends React.Component {}\n        Component.PropTypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-2", "\n        function MyComponent() { return (<div>{this.props.myProp}</div>) }\n        MyComponent.PropTypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-3", "\n        class Component extends React.Component {\n          static proptypes = {};\n        }\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-4", "\n        class Component extends React.Component {}\n        Component.proptypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-5", "\n        function MyComponent() { return (<div>{this.props.myProp}</div>) }\n        MyComponent.proptypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-6", "\n        class Component extends React.Component {\n          static ContextTypes = {};\n        }\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-7", "\n        class Component extends React.Component {}\n        Component.ContextTypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-8", "\n        function MyComponent() { return (<div>{this.props.myProp}</div>) }\n        MyComponent.ContextTypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-9", "\n        class Component extends React.Component {\n          static contexttypes = {};\n        }\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-10", "\n        class Component extends React.Component {}\n        Component.contexttypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-11", "\n        function MyComponent() { return (<div>{this.props.myProp}</div>) }\n        MyComponent.contexttypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-12", "\n        class Component extends React.Component {\n          static ChildContextTypes = {};\n        }\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-13", "\n        class Component extends React.Component {}\n        Component.ChildContextTypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-14", "\n        function MyComponent() { return (<div>{this.props.myProp}</div>) }\n        MyComponent.ChildContextTypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-15", "\n        class Component extends React.Component {\n          static childcontexttypes = {};\n        }\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-16", "\n        class Component extends React.Component {}\n        Component.childcontexttypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-17", "\n        function MyComponent() { return (<div>{this.props.myProp}</div>) }\n        MyComponent.childcontexttypes = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-18", "\n        class Component extends React.Component {\n          static DefaultProps = {};\n        }\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-19", "\n        class Component extends React.Component {}\n        Component.DefaultProps = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-20", "\n        function MyComponent() { return (<div>{this.props.myProp}</div>) }\n        MyComponent.DefaultProps = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-21", "\n        class Component extends React.Component {\n          static defaultprops = {};\n        }\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-22", "\n        class Component extends React.Component {}\n        Component.defaultprops = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-23", "\n        function MyComponent() { return (<div>{this.props.myProp}</div>) }\n        MyComponent.defaultprops = {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-24", "\n        Component.defaultprops = {}\n        class Component extends React.Component {}\n      ", []string{"typoStaticClassProp"}, []string{"Typo in static class property declaration"}},
		{"upstream invalid-26", "\n        class Hello extends React.Component {\n          static GetDerivedStateFromProps()  { }\n          ComponentWillMount() { }\n          UNSAFE_ComponentWillMount() { }\n          ComponentDidMount() { }\n          ComponentWillReceiveProps() { }\n          UNSAFE_ComponentWillReceiveProps() { }\n          ShouldComponentUpdate() { }\n          ComponentWillUpdate() { }\n          UNSAFE_ComponentWillUpdate() { }\n          GetSnapshotBeforeUpdate() { }\n          ComponentDidUpdate() { }\n          ComponentDidCatch() { }\n          ComponentWillUnmount() { }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ", []string{"typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod"}, []string{"Typo in component lifecycle method declaration: GetDerivedStateFromProps should be getDerivedStateFromProps", "Typo in component lifecycle method declaration: ComponentWillMount should be componentWillMount", "Typo in component lifecycle method declaration: UNSAFE_ComponentWillMount should be UNSAFE_componentWillMount", "Typo in component lifecycle method declaration: ComponentDidMount should be componentDidMount", "Typo in component lifecycle method declaration: ComponentWillReceiveProps should be componentWillReceiveProps", "Typo in component lifecycle method declaration: UNSAFE_ComponentWillReceiveProps should be UNSAFE_componentWillReceiveProps", "Typo in component lifecycle method declaration: ShouldComponentUpdate should be shouldComponentUpdate", "Typo in component lifecycle method declaration: ComponentWillUpdate should be componentWillUpdate", "Typo in component lifecycle method declaration: UNSAFE_ComponentWillUpdate should be UNSAFE_componentWillUpdate", "Typo in component lifecycle method declaration: GetSnapshotBeforeUpdate should be getSnapshotBeforeUpdate", "Typo in component lifecycle method declaration: ComponentDidUpdate should be componentDidUpdate", "Typo in component lifecycle method declaration: ComponentDidCatch should be componentDidCatch", "Typo in component lifecycle method declaration: ComponentWillUnmount should be componentWillUnmount"}},
		{"upstream invalid-27", "\n        class Hello extends React.Component {\n          static Getderivedstatefromprops() { }\n          Componentwillmount() { }\n          UNSAFE_Componentwillmount() { }\n          Componentdidmount() { }\n          Componentwillreceiveprops() { }\n          UNSAFE_Componentwillreceiveprops() { }\n          Shouldcomponentupdate() { }\n          Componentwillupdate() { }\n          UNSAFE_Componentwillupdate() { }\n          Getsnapshotbeforeupdate() { }\n          Componentdidupdate() { }\n          Componentdidcatch() { }\n          Componentwillunmount() { }\n          Render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ", []string{"typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod"}, []string{"Typo in component lifecycle method declaration: Getderivedstatefromprops should be getDerivedStateFromProps", "Typo in component lifecycle method declaration: Componentwillmount should be componentWillMount", "Typo in component lifecycle method declaration: UNSAFE_Componentwillmount should be UNSAFE_componentWillMount", "Typo in component lifecycle method declaration: Componentdidmount should be componentDidMount", "Typo in component lifecycle method declaration: Componentwillreceiveprops should be componentWillReceiveProps", "Typo in component lifecycle method declaration: UNSAFE_Componentwillreceiveprops should be UNSAFE_componentWillReceiveProps", "Typo in component lifecycle method declaration: Shouldcomponentupdate should be shouldComponentUpdate", "Typo in component lifecycle method declaration: Componentwillupdate should be componentWillUpdate", "Typo in component lifecycle method declaration: UNSAFE_Componentwillupdate should be UNSAFE_componentWillUpdate", "Typo in component lifecycle method declaration: Getsnapshotbeforeupdate should be getSnapshotBeforeUpdate", "Typo in component lifecycle method declaration: Componentdidupdate should be componentDidUpdate", "Typo in component lifecycle method declaration: Componentdidcatch should be componentDidCatch", "Typo in component lifecycle method declaration: Componentwillunmount should be componentWillUnmount", "Typo in component lifecycle method declaration: Render should be render"}},
		{"upstream invalid-28", "\n        class Hello extends React.Component {\n          static getderivedstatefromprops() { }\n          componentwillmount() { }\n          unsafe_componentwillmount() { }\n          componentdidmount() { }\n          componentwillreceiveprops() { }\n          unsafe_componentwillreceiveprops() { }\n          shouldcomponentupdate() { }\n          componentwillupdate() { }\n          unsafe_componentwillupdate() { }\n          getsnapshotbeforeupdate() { }\n          componentdidupdate() { }\n          componentdidcatch() { }\n          componentwillunmount() { }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ", []string{"typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod"}, []string{"Typo in component lifecycle method declaration: getderivedstatefromprops should be getDerivedStateFromProps", "Typo in component lifecycle method declaration: componentwillmount should be componentWillMount", "Typo in component lifecycle method declaration: unsafe_componentwillmount should be UNSAFE_componentWillMount", "Typo in component lifecycle method declaration: componentdidmount should be componentDidMount", "Typo in component lifecycle method declaration: componentwillreceiveprops should be componentWillReceiveProps", "Typo in component lifecycle method declaration: unsafe_componentwillreceiveprops should be UNSAFE_componentWillReceiveProps", "Typo in component lifecycle method declaration: shouldcomponentupdate should be shouldComponentUpdate", "Typo in component lifecycle method declaration: componentwillupdate should be componentWillUpdate", "Typo in component lifecycle method declaration: unsafe_componentwillupdate should be UNSAFE_componentWillUpdate", "Typo in component lifecycle method declaration: getsnapshotbeforeupdate should be getSnapshotBeforeUpdate", "Typo in component lifecycle method declaration: componentdidupdate should be componentDidUpdate", "Typo in component lifecycle method declaration: componentdidcatch should be componentDidCatch", "Typo in component lifecycle method declaration: componentwillunmount should be componentWillUnmount"}},
		{"upstream invalid-29", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.propTypes = {\n            a: PropTypes.Number.isRequired\n        }\n      ", []string{"typoPropType"}, []string{"Typo in declared prop type: Number"}},
		{"upstream invalid-30", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.propTypes = {\n            a: PropTypes.number.isrequired\n        }\n      ", []string{"typoPropTypeChain"}, []string{"Typo in prop type chain qualifier: isrequired"}},
		{"upstream invalid-31", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.number.isrequired\n          }\n        };\n      ", []string{"typoPropTypeChain"}, []string{"Typo in prop type chain qualifier: isrequired"}},
		{"upstream invalid-32", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.Number\n          }\n        };\n      ", []string{"typoPropType"}, []string{"Typo in declared prop type: Number"}},
		{"upstream invalid-33", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.propTypes = {\n            a: PropTypes.Number\n        }\n      ", []string{"typoPropType"}, []string{"Typo in declared prop type: Number"}},
		{"upstream invalid-34", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.propTypes = {\n          a: PropTypes.shape({\n            b: PropTypes.String,\n            c: PropTypes.number.isRequired,\n          })\n        }\n      ", []string{"typoPropType"}, []string{"Typo in declared prop type: String"}},
		{"upstream invalid-35", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.propTypes = {\n          a: PropTypes.oneOfType([\n            PropTypes.bools,\n            PropTypes.number,\n          ])\n        }\n      ", []string{"typoPropType"}, []string{"Typo in declared prop type: bools"}},
		{"upstream invalid-36", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.propTypes = {\n          a: PropTypes.bools,\n          b: PropTypes.Array,\n          c: PropTypes.function,\n          d: PropTypes.objectof,\n        }\n      ", []string{"typoPropType", "typoPropType", "typoPropType", "typoPropType"}, []string{"Typo in declared prop type: bools", "Typo in declared prop type: Array", "Typo in declared prop type: function", "Typo in declared prop type: objectof"}},
		{"upstream invalid-37", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.childContextTypes = {\n          a: PropTypes.bools,\n          b: PropTypes.Array,\n          c: PropTypes.function,\n          d: PropTypes.objectof,\n        }\n      ", []string{"typoPropType", "typoPropType", "typoPropType", "typoPropType"}, []string{"Typo in declared prop type: bools", "Typo in declared prop type: Array", "Typo in declared prop type: function", "Typo in declared prop type: objectof"}},
		{"upstream invalid-38", "\n        import PropTypes from 'prop-types';\n        class Component extends React.Component {};\n        Component.childContextTypes = {\n          a: PropTypes.bools,\n          b: PropTypes.Array,\n          c: PropTypes.function,\n          d: PropTypes.objectof,\n        }\n      ", []string{"typoPropType", "typoPropType", "typoPropType", "typoPropType"}, []string{"Typo in declared prop type: bools", "Typo in declared prop type: Array", "Typo in declared prop type: function", "Typo in declared prop type: objectof"}},
		{"upstream invalid-39", "\n        import PropTypes from 'prop-types';\n        class Component extends React.Component {};\n        Component.propTypes = {\n          a: PropTypes.string.isrequired,\n          b: PropTypes.shape({\n            c: PropTypes.number\n          }).isrequired\n        }\n      ", []string{"typoPropTypeChain", "typoPropTypeChain"}, []string{"Typo in prop type chain qualifier: isrequired", "Typo in prop type chain qualifier: isrequired"}},
		{"upstream invalid-40", "\n        import RealPropTypes from 'prop-types';\n        class Component extends React.Component {};\n        Component.childContextTypes = {\n          a: RealPropTypes.bools,\n          b: RealPropTypes.Array,\n          c: RealPropTypes.function,\n          d: RealPropTypes.objectof,\n        }\n      ", []string{"typoPropType", "typoPropType", "typoPropType", "typoPropType"}, []string{"Typo in declared prop type: bools", "Typo in declared prop type: Array", "Typo in declared prop type: function", "Typo in declared prop type: objectof"}},
		{"upstream invalid-41", "\n      import React from 'react';\n      class Component extends React.Component {};\n      Component.propTypes = {\n        a: React.PropTypes.string.isrequired,\n        b: React.PropTypes.shape({\n          c: React.PropTypes.number\n        }).isrequired\n      }\n    ", []string{"typoPropTypeChain", "typoPropTypeChain"}, []string{"Typo in prop type chain qualifier: isrequired", "Typo in prop type chain qualifier: isrequired"}},
		{"upstream invalid-42", "\n        import React from 'react';\n        class Component extends React.Component {};\n        Component.childContextTypes = {\n          a: React.PropTypes.bools,\n          b: React.PropTypes.Array,\n          c: React.PropTypes.function,\n          d: React.PropTypes.objectof,\n        }\n      ", []string{"typoPropType", "typoPropType", "typoPropType", "typoPropType"}, []string{"Typo in declared prop type: bools", "Typo in declared prop type: Array", "Typo in declared prop type: function", "Typo in declared prop type: objectof"}},
		{"upstream invalid-43", "\n      import { PropTypes } from 'react';\n      class Component extends React.Component {};\n      Component.propTypes = {\n        a: PropTypes.string.isrequired,\n        b: PropTypes.shape({\n          c: PropTypes.number\n        }).isrequired\n      }\n    ", []string{"typoPropTypeChain", "typoPropTypeChain"}, []string{"Typo in prop type chain qualifier: isrequired", "Typo in prop type chain qualifier: isrequired"}},
		{"upstream invalid-44", "\n      import 'react';\n      class Component extends React.Component {};\n    ", []string{"noReactBinding"}, []string{"`'react'` imported without a local `React` binding."}},
		{"upstream invalid-45", "\n        import { PropTypes } from 'react';\n        class Component extends React.Component {};\n        Component.childContextTypes = {\n          a: PropTypes.bools,\n          b: PropTypes.Array,\n          c: PropTypes.function,\n          d: PropTypes.objectof,\n        }\n      ", []string{"typoPropType", "typoPropType", "typoPropType", "typoPropType"}, []string{"Typo in declared prop type: bools", "Typo in declared prop type: Array", "Typo in declared prop type: function", "Typo in declared prop type: objectof"}},
		{"upstream invalid-46", "\n      import PropTypes from 'prop-types';\n      class Component extends React.Component {};\n      Component.propTypes = {\n        a: PropTypes.string.isrequired,\n        b: PropTypes.shape({\n          c: PropTypes.number\n        }).isrequired\n      }\n      ", []string{"typoPropTypeChain", "typoPropTypeChain"}, []string{"Typo in prop type chain qualifier: isrequired", "Typo in prop type chain qualifier: isrequired"}},
		{"upstream invalid-47", "\n      import PropTypes from 'prop-types';\n      class Component extends React.Component {};\n      Component.propTypes = {\n        a: PropTypes.string.isrequired,\n        b: PropTypes.shape({\n          c: PropTypes.number\n        }).isrequired\n      }\n    ", []string{"typoPropTypeChain", "typoPropTypeChain"}, []string{"Typo in prop type chain qualifier: isrequired", "Typo in prop type chain qualifier: isrequired"}},
		{"upstream invalid-48", "\n        import React from 'react';\n        import PropTypes from 'prop-types';\n        const Component = React.createReactClass({\n          propTypes: {\n            a: PropTypes.string.isrequired,\n            b: PropTypes.shape({\n              c: PropTypes.number\n            }).isrequired\n          }\n        });\n      ", []string{"typoPropTypeChain", "typoPropTypeChain"}, []string{"Typo in prop type chain qualifier: isrequired", "Typo in prop type chain qualifier: isrequired"}},
		{"upstream invalid-49", "\n        import React from 'react';\n        import PropTypes from 'prop-types';\n        const Component = React.createReactClass({\n          childContextTypes: {\n            a: PropTypes.bools,\n            b: PropTypes.Array,\n            c: PropTypes.function,\n            d: PropTypes.objectof,\n          }\n        });\n      ", []string{"typoPropType", "typoPropType", "typoPropType", "typoPropType"}, []string{"Typo in declared prop type: bools", "Typo in declared prop type: Array", "Typo in declared prop type: function", "Typo in declared prop type: objectof"}},
		{"upstream invalid-50", "\n        import React from 'react';\n        const Component = React.createReactClass({\n          proptypes: {},\n          childcontexttypes: {},\n          contexttypes: {},\n          getdefaultProps() { },\n          getinitialState() { },\n          getChildcontext() { },\n          ComponentWillMount() { },\n          ComponentDidMount() { },\n          ComponentWillReceiveProps() { },\n          ShouldComponentUpdate() { },\n          ComponentWillUpdate() { },\n          ComponentDidUpdate() { },\n          ComponentWillUnmount() { },\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        });\n      ", []string{"typoPropDeclaration", "typoPropDeclaration", "typoPropDeclaration", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod", "typoLifecycleMethod"}, []string{"Typo in property declaration", "Typo in property declaration", "Typo in property declaration", "Typo in component lifecycle method declaration: getdefaultProps should be getDefaultProps", "Typo in component lifecycle method declaration: getinitialState should be getInitialState", "Typo in component lifecycle method declaration: getChildcontext should be getChildContext", "Typo in component lifecycle method declaration: ComponentWillMount should be componentWillMount", "Typo in component lifecycle method declaration: ComponentDidMount should be componentDidMount", "Typo in component lifecycle method declaration: ComponentWillReceiveProps should be componentWillReceiveProps", "Typo in component lifecycle method declaration: ShouldComponentUpdate should be shouldComponentUpdate", "Typo in component lifecycle method declaration: ComponentWillUpdate should be componentWillUpdate", "Typo in component lifecycle method declaration: ComponentDidUpdate should be componentDidUpdate", "Typo in component lifecycle method declaration: ComponentWillUnmount should be componentWillUnmount"}},
		{"upstream invalid-51", "\n        class Hello extends React.Component {\n          getDerivedStateFromProps() { }\n        }\n      ", []string{"staticLifecycleMethod"}, []string{"Lifecycle method should be static: getDerivedStateFromProps"}},
		{"upstream invalid-52", "\n        import 'prop-types'\n      ", []string{"noPropTypesBinding"}, []string{"`'prop-types'` imported without a local `PropTypes` binding."}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoTypos(t, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoTyposStaysSilent runs upstream's clean cases.
func TestNoTyposStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"upstream valid-0", "\n          import createReactClass from 'create-react-class'\n          function hello (extra = {}) {\n            return createReactClass({\n              noteType: 'hello',\n              renderItem () {\n                return null\n              },\n              ...extra\n            })\n          }\n      "},
		{"upstream valid-1", "\n        class First {\n          static PropTypes = {key: \"myValue\"};\n          static ContextTypes = {key: \"myValue\"};\n          static ChildContextTypes = {key: \"myValue\"};\n          static DefaultProps = {key: \"myValue\"};\n        }\n      "},
		{"upstream valid-2", "\n        class First {}\n        First.PropTypes = {key: \"myValue\"};\n        First.ContextTypes = {key: \"myValue\"};\n        First.ChildContextTypes = {key: \"myValue\"};\n        First.DefaultProps = {key: \"myValue\"};\n      "},
		{"upstream valid-3", "\n        class First extends React.Component {\n          static propTypes = {key: \"myValue\"};\n          static contextTypes = {key: \"myValue\"};\n          static childContextTypes = {key: \"myValue\"};\n          static defaultProps = {key: \"myValue\"};\n        }\n      "},
		{"upstream valid-4", "\n        class First extends React.Component {}\n        First.propTypes = {key: \"myValue\"};\n        First.contextTypes = {key: \"myValue\"};\n        First.childContextTypes = {key: \"myValue\"};\n        First.defaultProps = {key: \"myValue\"};\n      "},
		{"upstream valid-5", "\n        class MyClass {\n          propTypes = {key: \"myValue\"};\n          contextTypes = {key: \"myValue\"};\n          childContextTypes = {key: \"myValue\"};\n          defaultProps = {key: \"myValue\"};\n        }\n      "},
		{"upstream valid-6", "\n        class MyClass {\n          PropTypes = {key: \"myValue\"};\n          ContextTypes = {key: \"myValue\"};\n          ChildContextTypes = {key: \"myValue\"};\n          DefaultProps = {key: \"myValue\"};\n        }\n      "},
		{"upstream valid-7", "\n        class MyClass {\n          proptypes = {key: \"myValue\"};\n          contexttypes = {key: \"myValue\"};\n          childcontextypes = {key: \"myValue\"};\n          defaultprops = {key: \"myValue\"};\n        }\n      "},
		{"upstream valid-8", "\n        class MyClass {\n          static PropTypes() {};\n          static ContextTypes() {};\n          static ChildContextTypes() {};\n          static DefaultProps() {};\n        }\n      "},
		{"upstream valid-9", "\n        class MyClass {\n          static proptypes() {};\n          static contexttypes() {};\n          static childcontexttypes() {};\n          static defaultprops() {};\n        }\n      "},
		{"upstream valid-10", "\n        class MyClass {}\n        MyClass.prototype.PropTypes = function() {};\n        MyClass.prototype.ContextTypes = function() {};\n        MyClass.prototype.ChildContextTypes = function() {};\n        MyClass.prototype.DefaultProps = function() {};\n      "},
		{"upstream valid-11", "\n        class MyClass {}\n        MyClass.PropTypes = function() {};\n        MyClass.ContextTypes = function() {};\n        MyClass.ChildContextTypes = function() {};\n        MyClass.DefaultProps = function() {};\n      "},
		{"upstream valid-12", "\n        function MyRandomFunction() {}\n        MyRandomFunction.PropTypes = {};\n        MyRandomFunction.ContextTypes = {};\n        MyRandomFunction.ChildContextTypes = {};\n        MyRandomFunction.DefaultProps = {};\n      "},
		{"upstream valid-13", "\n        class First extends React.Component {}\n        First[\"prop\" + \"Types\"] = {};\n        First[\"context\" + \"Types\"] = {};\n        First[\"childContext\" + \"Types\"] = {};\n        First[\"default\" + \"Props\"] = {};\n      "},
		{"upstream valid-14", "\n        class First extends React.Component {}\n        First[\"PROP\" + \"TYPES\"] = {};\n        First[\"CONTEXT\" + \"TYPES\"] = {};\n        First[\"CHILDCONTEXT\" + \"TYPES\"] = {};\n        First[\"DEFAULT\" + \"PROPS\"] = {};\n      "},
		{"upstream valid-15", "\n        const propTypes = \"PROPTYPES\"\n        const contextTypes = \"CONTEXTTYPES\"\n        const childContextTypes = \"CHILDCONTEXTTYPES\"\n        const defaultProps = \"DEFAULTPROPS\"\n\n        class First extends React.Component {}\n        First[propTypes] = {};\n        First[contextTypes] = {};\n        First[childContextTypes] = {};\n        First[defaultProps] = {};\n      "},
		{"upstream valid-16", "\n        class Hello extends React.Component {\n          static getDerivedStateFromProps() { }\n          componentWillMount() { }\n          componentDidMount() { }\n          componentWillReceiveProps() { }\n          shouldComponentUpdate() { }\n          componentWillUpdate() { }\n          componentDidUpdate() { }\n          componentWillUnmount() { }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      "},
		{"upstream valid-17", "\n        class Hello extends React.Component {\n          \"componentDidMount\"() { }\n          \"my-method\"() { }\n        }\n      "},
		{"upstream valid-18", "\n        class MyClass {\n          componentWillMount() { }\n          componentDidMount() { }\n          componentWillReceiveProps() { }\n          shouldComponentUpdate() { }\n          componentWillUpdate() { }\n          componentDidUpdate() { }\n          componentWillUnmount() { }\n          render() { }\n        }\n      "},
		{"upstream valid-19", "\n        class MyClass {\n          componentwillmount() { }\n          componentdidmount() { }\n          componentwillreceiveprops() { }\n          shouldcomponentupdate() { }\n          componentwillupdate() { }\n          componentdidupdate() { }\n          componentwillUnmount() { }\n          render() { }\n        }\n      "},
		{"upstream valid-20", "\n        class MyClass {\n          Componentwillmount() { }\n          Componentdidmount() { }\n          Componentwillreceiveprops() { }\n          Shouldcomponentupdate() { }\n          Componentwillupdate() { }\n          Componentdidupdate() { }\n          ComponentwillUnmount() { }\n          Render() { }\n        }\n      "},
		{"upstream valid-21", "\n        function test(b) {\n          return a.bind(b);\n        }\n        function a() {}\n      "},
		{"upstream valid-22", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.propTypes = {\n          a: PropTypes.number.isRequired\n        }\n      "},
		{"upstream valid-23", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.propTypes = {\n          e: PropTypes.shape({\n            ea: PropTypes.string,\n          })\n        }\n      "},
		{"upstream valid-24", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.propTypes = {\n          a: PropTypes.string,\n          b: PropTypes.string.isRequired,\n          c: PropTypes.shape({\n            d: PropTypes.string,\n            e: PropTypes.number.isRequired,\n          }).isRequired\n        }\n      "},
		{"upstream valid-25", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.propTypes = {\n          a: PropTypes.oneOfType([\n            PropTypes.string,\n            PropTypes.number\n          ])\n        }\n      "},
		{"upstream valid-26", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.propTypes = {\n          a: PropTypes.oneOf([\n            'hello',\n            'hi'\n          ])\n        }\n      "},
		{"upstream valid-27", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.childContextTypes = {\n          a: PropTypes.string,\n          b: PropTypes.string.isRequired,\n          c: PropTypes.shape({\n            d: PropTypes.string,\n            e: PropTypes.number.isRequired,\n          }).isRequired\n        }\n      "},
		{"upstream valid-28", "\n        import PropTypes from \"prop-types\";\n        class Component extends React.Component {};\n        Component.contextTypes = {\n          a: PropTypes.string,\n          b: PropTypes.string.isRequired,\n          c: PropTypes.shape({\n            d: PropTypes.string,\n            e: PropTypes.number.isRequired,\n          }).isRequired\n        }\n      "},
		{"upstream valid-29", "\n        import PropTypes from 'prop-types'\n        import * as MyPropTypes from 'lib/my-prop-types'\n        class Component extends React.Component {};\n        Component.propTypes = {\n          a: PropTypes.string,\n          b: MyPropTypes.MYSTRING,\n          c: MyPropTypes.MYSTRING.isRequired,\n        }\n      "},
		{"upstream valid-30", "\n        import PropTypes from \"prop-types\"\n        import * as MyPropTypes from 'lib/my-prop-types'\n        class Component extends React.Component {};\n        Component.propTypes = {\n          b: PropTypes.string,\n          a: MyPropTypes.MYSTRING,\n        }\n      "},
		{"upstream valid-31", "\n        import CustomReact from \"react\"\n        class Component extends React.Component {};\n        Component.propTypes = {\n          b: CustomReact.PropTypes.string,\n        }\n      "},
		{"upstream valid-32", "\n        class Component extends React.Component {};\n        Component.propTypes = {\n          a: PropTypes.shape(),\n        };\n        Component.contextTypes = {\n          a: PropTypes.shape(),\n        };\n      "},
		{"upstream valid-33", "\n        const fn = (err, res) => {\n          const { body: data = {} } = { ...res };\n          data.time = data.time || {};\n        };\n      "},
		{"upstream valid-34", "\n        class Component extends React.Component {};\n        Component.propTypes = {\n          b: string.isRequired,\n          c: PropTypes.shape({\n            d: number.isRequired,\n          }).isRequired\n        }\n      "},
		{"upstream valid-35", "\n        import React from 'react';\n        import PropTypes from 'prop-types';\n        const Component = React.createReactClass({\n          propTypes: {\n            a: PropTypes.string.isRequired,\n            b: PropTypes.shape({\n              c: PropTypes.number\n            }).isRequired\n          }\n        });\n      "},
		{"upstream valid-36", "\n        import React from 'react';\n        import PropTypes from 'prop-types';\n        const Component = React.createReactClass({\n          childContextTypes: {\n            a: PropTypes.bool,\n            b: PropTypes.array,\n            c: PropTypes.func,\n            d: PropTypes.object,\n          }\n        });\n      "},
		{"upstream valid-37", "\n        import React from 'react';\n        const Component = React.createReactClass({\n          propTypes: {},\n          childContextTypes: {},\n          contextTypes: {},\n          componentWillMount() { },\n          componentDidMount() { },\n          componentWillReceiveProps() { },\n          shouldComponentUpdate() { },\n          componentWillUpdate() { },\n          componentDidUpdate() { },\n          componentWillUnmount() { },\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        });\n      "},
		{"upstream valid-38", "\n        import { string, element } from \"prop-types\";\n\n        class Sample extends React.Component {\n          render() { return null; }\n        }\n\n        Sample.propTypes = {\n          title: string.isRequired,\n          body: element.isRequired\n        };\n      "},
		{"upstream valid-39", "\n        import React from 'react';\n\n        const A = { B: 'C' };\n\n        export default class MyComponent extends React.Component {\n          [A.B] () {\n            return null\n          }\n        }\n      "},
		{"upstream valid-40", "\n        const MyComponent = React.forwardRef((props, ref) => <div />);\n        MyComponent.defaultProps = { value: \"\" };\n      "},
		{"upstream valid-41", "\n        import styled from \"styled-components\";\n\n        const MyComponent = styled.div;\n        MyComponent.defaultProps = { value: \"\" };\n      "},
		{"upstream valid-42", "\n        class Editor extends React.Component {\n            #somethingPrivate() {\n              // ...\n            }\n\n            render() {\n            const { value = '' } = this.props;\n\n            return (\n              <textarea>\n                {value}\n              </textarea>\n            );\n          }\n        }\n      "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoTypos(t, testCase.sourceText))
		})
	}
}

// runNoTypos drives the rule on an untyped program.
//
// `Run` rather than `RunTyped`: every question is syntactic. The one place upstream resolves a name
// through its registry is the member-assignment arm, and that is reproduced by searching the file
// for the declaration rather than by asking the checker, because the registry is doing the same
// thing. `TestNoTyposNeedsNoChecker` pins that both harnesses agree.
func runNoTypos(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.Run(t, NoTypos, noTyposFile, sourceText)
}

// TestNoTyposJsDocComponentAnnotationIsNotHonored records the one corpus case held back.
//
// The clone asserts that a class annotated `/** @extends React.Component */` counts as a component,
// so a mis-cased static property on it reports. The installed build is SILENT, because
// `componentUtil.isExplicitComponent` reads the annotation through `sourceCode.getJSDocComment` and
// that method was removed in ESLint 9.
//
// Measured under the TypeScript parser and the default parser, both silent, with the control below
// confirming the same shape reports when the class extends `React.Component` for real. So the
// silence is the annotation path being dead rather than anything else declining, and silence is what
// this port produces.
func TestNoTyposJsDocComponentAnnotationIsNotHonored(t *testing.T) {
	t.Parallel()

	const annotated = "/** @extends React.Component */\nclass MyComponent extends BaseComponent {}\nMyComponent.PROPTYPES = {};\n"
	rule_testing.ExpectClean(t, runNoTypos(t, annotated))

	// The control. Without it, the silence above would be indistinguishable from a rule that cannot
	// see a member assignment at all.
	const real = "declare const React: any;\nclass MyComponent extends React.Component {}\nMyComponent.PROPTYPES = {};\n"
	rule_testing.ExpectFindings(t, runNoTypos(t, real), "typoStaticClassProp")
}

// TestNoTyposImportBindingsGateThePropTypeArms pins the guard that makes two ids inert.
//
// `checkValidProp` returns immediately when neither package name is known, so a file that imports
// neither `prop-types` nor `react` produces no prop-type findings however misspelled it is. That is
// upstream's guard and it is easy to drop, because the corpus imports one of the two in nearly every
// case. Every verdict measured on the installed build.
//
// The source-order rows are the reason this rule walks the file rather than registering per-kind
// listeners: upstream sets the binding as ESLint reaches the import, so a propTypes object written
// ABOVE the import sees no binding.
func TestNoTyposImportBindingsGateThePropTypeArms(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"with a prop-types import the typo reports",
			"import PropTypes from 'prop-types';\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PropTypes.strng };\n",
			[]string{"typoPropType"},
		},
		{
			"with NO import at all the same typo is silent",
			"declare const PropTypes: any;\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PropTypes.strng };\n",
			nil,
		},
		{
			"an import written BELOW the usage does not bind it",
			"declare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PropTypes.strng };\nimport PropTypes from 'prop-types';\n",
			nil,
		},
		{
			"the react namespace binds the pre-15.5 spelling",
			"import React from 'react';\nclass Foo extends React.Component {}\nFoo.propTypes = { a: React.PropTypes.strng };\n",
			[]string{"typoPropType"},
		},
		{
			"a named PropTypes specifier from react binds the bare name",
			"import { PropTypes } from 'react';\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PropTypes.strng };\n",
			[]string{"typoPropType"},
		},
		{
			"a renamed PropTypes specifier binds the local name",
			"import { PropTypes as PT } from 'react';\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PT.strng };\n",
			[]string{"typoPropType"},
		},
		// The rows above cannot see the import guard on their own, and a mutant deleting it survived
		// them: with NO import both the guard AND `isPropTypesPackage` decline, so both versions
		// reach silence by different routes. The three below separate them. A chain qualifier
		// reached through a CALL never consults `isPropTypesPackage` at all, so the import guard is
		// the only thing deciding, and it flips on whether either package has been named.
		{
			"a call-chain typo is silent with no import",
			"declare function something(): any;\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: something().isRequird };\n",
			nil,
		},
		{
			"the same typo reports once react is imported",
			"import React from 'react';\ndeclare function something(): any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: something().isRequird };\n",
			[]string{"typoPropTypeChain"},
		},
		{
			"and reports once prop-types is imported",
			"import PropTypes from 'prop-types';\ndeclare function something(): any;\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: something().isRequird };\n",
			[]string{"typoPropTypeChain"},
		},
		{
			"a side-effect prop-types import reports its own id",
			"import 'prop-types';\n",
			[]string{"noPropTypesBinding"},
		},
		{
			"a side-effect react import reports its own id",
			"import 'react';\n",
			[]string{"noReactBinding"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoTypos(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoTyposReturningJsxDecidesTheMemberArm pins upstream's isReturningJSX.
//
// The member-assignment arm accepts a class component OR a non-class whose own body returns JSX.
// That predicate is narrower than it sounds and the corpus barely exercises it. Every row measured
// on the installed build before it was written here.
//
// The last row is the one a port is most likely to get wrong: JSX inside a nested callback belongs
// to that callback, not to the enclosing function, so a component that only maps over items and
// returns null does NOT count.
func TestNoTyposReturningJsxDecidesTheMemberArm(t *testing.T) {
	t.Parallel()

	const preamble = "import PropTypes from 'prop-types';\ndeclare const items: any[];\n"
	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a function returning JSX counts", preamble + "function Foo() { return <div/>; }\nFoo.PropTypes = {};\n", []string{"typoStaticClassProp"}},
		{"an arrow with an implicit JSX body counts", preamble + "const Foo = () => <div/>;\nFoo.PropTypes = {};\n", []string{"typoStaticClassProp"}},
		{"a function returning null does NOT count", preamble + "function Foo() { return null; }\nFoo.PropTypes = {};\n", nil},
		{"a function returning a string does NOT count", preamble + "function Foo() { return 'x'; }\nFoo.PropTypes = {};\n", nil},
		{"a function returning nothing does NOT count", preamble + "function Foo() { }\nFoo.PropTypes = {};\n", nil},
		{"there is no name gate, lowercase counts too", preamble + "function foo() { return <div/>; }\nfoo.PropTypes = {};\n", []string{"typoStaticClassProp"}},
		{"a nested return counts", preamble + "declare const x: boolean;\nfunction Foo() { if(x) { return <div/>; } return null; }\nFoo.PropTypes = {};\n", []string{"typoStaticClassProp"}},
		// Parenthesized returns. Upstream's parser folds parentheses away so these reach its JSX
		// test as bare elements; ours keeps them, and without an unwrapping loop eight of the
		// corpus's own reporting cases went silent, because wrapping a returned element in
		// parentheses is how a multi-line component is ordinarily written. The double form is why
		// the unwrap is a loop rather than one step.
		{"a parenthesized JSX return counts", preamble + "function Foo() { return (<div/>); }\nFoo.PropTypes = {};\n", []string{"typoStaticClassProp"}},
		{"a doubly parenthesized JSX return counts", preamble + "function Foo() { return ((<div/>)); }\nFoo.PropTypes = {};\n", []string{"typoStaticClassProp"}},
		{"a parenthesized non-JSX return still does not", preamble + "function Foo() { return (null); }\nFoo.PropTypes = {};\n", nil},
		{"JSX only inside a callback does NOT count", preamble + "function Foo() { items.map(() => <div/>); return null; }\nFoo.PropTypes = {};\n", nil},
		// The row above cannot see the nested-function skip, and a mutant removing it survived: an
		// arrow's JSX is a concise BODY rather than a return statement, so the walk never meets a
		// return there and both versions reach silence. These two carry an explicit `return` inside
		// a nested function, which is the only shape where the skip alone decides. Both measured
		// silent on the installed build.
		{"a nested function DECLARATION returning JSX does not count", preamble + "function Foo() { function Inner() { return <div/>; } return null; }\nFoo.PropTypes = {};\n", nil},
		{"a nested function EXPRESSION returning JSX does not count", preamble + "function Foo() { const f = function() { return <div/>; }; return null; }\nFoo.PropTypes = {};\n", nil},
		{"a correct spelling is clean either way", preamble + "function Foo() { return <div/>; }\nFoo.propTypes = {};\n", nil},
		{"a plain object is not a component", preamble + "const foo = { a: 1 };\nfoo.PropTypes = {};\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoTypos(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoTyposAcceptsEveryPropTypesModuleKey pins the accept set, quirks included.
//
// Upstream takes `Object.keys(require('prop-types'))` wholesale, so three module helpers that are
// not prop types at all are accepted: `checkPropTypes`, `resetWarningCache`, and the package's own
// `PropTypes` self-reference. That is an accident rather than a decision, and reproducing it matters
// because narrowing the set would report where upstream is silent. Measured on the installed build.
func TestNoTyposAcceptsEveryPropTypesModuleKey(t *testing.T) {
	t.Parallel()

	const preamble = "import PropTypes from 'prop-types';\ndeclare const React: any;\nclass Foo extends React.Component {}\n"
	accepted := []string{
		"array", "bigint", "bool", "func", "number", "object", "string", "symbol",
		"any", "arrayOf", "element", "elementType", "instanceOf", "node", "objectOf",
		"oneOf", "oneOfType", "shape", "exact",
		// The three that are module helpers rather than prop types.
		"checkPropTypes", "resetWarningCache", "PropTypes",
	}
	for _, name := range accepted {
		t.Run("accepts "+name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoTypos(t,
				preamble+"Foo.propTypes = { a: PropTypes."+name+" };\n"))
		})
	}

	for _, name := range []string{"strng", "boolean", "Array", "func2"} {
		t.Run("rejects "+name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoTypos(t,
				preamble+"Foo.propTypes = { a: PropTypes."+name+" };\n"), "typoPropType")
		})
	}
}

// TestNoTyposRecursesThroughShapeAndOneOfType pins the two recursive arms.
//
// `shape(...)` walks its object literal and `oneOfType([...])` walks each element, and every other
// call is left alone. The corpus covers these; these rows pin the boundary, which it does not.
func TestNoTyposRecursesThroughShapeAndOneOfType(t *testing.T) {
	t.Parallel()

	const preamble = "import PropTypes from 'prop-types';\ndeclare const React: any;\nclass Foo extends React.Component {}\n"
	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a typo inside shape reports", preamble + "Foo.propTypes = { a: PropTypes.shape({ b: PropTypes.strng }) };\n", []string{"typoPropType"}},
		{"a typo inside oneOfType reports", preamble + "Foo.propTypes = { a: PropTypes.oneOfType([PropTypes.strng]) };\n", []string{"typoPropType"}},
		{"a typo nested two deep reports", preamble + "Foo.propTypes = { a: PropTypes.shape({ b: PropTypes.shape({ c: PropTypes.strng }) }) };\n", []string{"typoPropType"}},
		// Only three property names have their object walked, and `defaultProps` is deliberately
		// not among them: it holds VALUES rather than validators, so a name that looks like a
		// misspelled prop type there is just data. A mutant widening the gate to every property
		// survived the rest of this suite until this row existed. Measured silent on the installed
		// build, with the three walked names beside it as controls.
		{"a defaultProps object is NOT walked", preamble + "Foo.defaultProps = { a: PropTypes.strng };\n", nil},
		{"a contextTypes object IS walked", preamble + "Foo.contextTypes = { a: PropTypes.strng };\n", []string{"typoPropType"}},
		{"a childContextTypes object IS walked", preamble + "Foo.childContextTypes = { a: PropTypes.strng };\n", []string{"typoPropType"}},
		{"another call is not walked", preamble + "Foo.propTypes = { a: PropTypes.instanceOf(Thing) };\n", nil},
		{"oneOf is not walked, its arguments are values", preamble + "Foo.propTypes = { a: PropTypes.oneOf(['strng']) };\n", nil},
		{"a chain qualifier after shape reports", preamble + "Foo.propTypes = { a: PropTypes.shape({}).isRequird };\n", []string{"typoPropTypeChain"}},
		{"a correct chain after shape is clean", preamble + "Foo.propTypes = { a: PropTypes.shape({}).isRequired };\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoTypos(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoTyposLifecycleArmsAreIndependent pins that one name can trip two judgments.
//
// A method whose lowercased name is `getderivedstatefromprops` and which is not declared `static`
// reports `staticLifecycleMethod`, and if its casing also differs it reports `typoLifecycleMethod`
// as well. Upstream emits both, in that order, which no single-id fixture can see.
func TestNoTyposLifecycleArmsAreIndependent(t *testing.T) {
	t.Parallel()

	const head = "declare const React: any;\nclass Foo extends React.Component { "
	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"correct casing without static reports only the static id",
			head + "getDerivedStateFromProps() {} }\n",
			[]string{"staticLifecycleMethod"},
		},
		{
			"wrong casing without static reports BOTH, static first",
			head + "getderivedstatefromprops() {} }\n",
			[]string{"staticLifecycleMethod", "typoLifecycleMethod"},
		},
		{
			"wrong casing WITH static reports only the typo id",
			head + "static getderivedstatefromprops() {} }\n",
			[]string{"typoLifecycleMethod"},
		},
		{
			"correct casing with static is clean",
			head + "static getDerivedStateFromProps() {} }\n",
			nil,
		},
		{
			"an instance lifecycle typo reports one id",
			head + "ComponentDidMount() {} }\n",
			[]string{"typoLifecycleMethod"},
		},
		{
			"an instance lifecycle spelled right is clean",
			head + "componentDidMount() {} }\n",
			nil,
		},
		{
			"a method that is not a lifecycle name is clean",
			head + "SomethingElse() {} }\n",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoTypos(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoTyposSpans asserts WHERE each finding points.
//
// The two casing ids point at the KEY, the two prop-type ids point at the offending identifier
// inside the chain, and the two lifecycle ids point at the whole member. Those are three different
// anchors that no id assertion can tell apart, and getting one wrong means underlining a whole
// method where upstream underlines one word.
func TestNoTyposSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantTexts  []string
	}{
		{
			"a static class property points at the key",
			"declare const React: any;\nclass Foo extends React.Component { static PropTypes = {}; }",
			[]string{"PropTypes"},
		},
		{
			"a member assignment points at the property name",
			"declare const React: any;\nclass Foo extends React.Component {}\nFoo.PropTypes = {};",
			[]string{"PropTypes"},
		},
		{
			"a prop type typo points at the misspelled name",
			"import PropTypes from 'prop-types';\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PropTypes.strng };",
			[]string{"strng"},
		},
		{
			"a chain typo points at the qualifier",
			"import PropTypes from 'prop-types';\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PropTypes.string.isRequird };",
			[]string{"isRequird"},
		},
		{
			"a lifecycle typo points at the whole method",
			"declare const React: any;\nclass Foo extends React.Component { ComponentDidMount() {} }",
			[]string{"ComponentDidMount() {}"},
		},
		{
			"a createReactClass property typo points at the key",
			"declare function createReactClass(spec: any): any;\ncreateReactClass({ PropTypes: {} });",
			[]string{"PropTypes"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoTypos(t, testCase.sourceText)
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

// TestNoTyposMessagesReadAsWritten pins the interpolated text.
//
// Four ids interpolate and two of those interpolate twice, so a format slot filled with the wrong
// value passes every id assertion in this file. Asserted against literals typed here rather than
// against the rule's own constants, because comparing a diagnostic to the constant it was reported
// with is an equality whose two sides move together under mutation.
func TestNoTyposMessagesReadAsWritten(t *testing.T) {
	t.Parallel()

	propType := runNoTypos(t,
		"import PropTypes from 'prop-types';\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PropTypes.strng };\n")
	rule_testing.ExpectFindings(t, propType, "typoPropType")
	if !strings.Contains(propType.Diagnostics[0].Message.Description, "`strng` is not a prop type") {
		t.Errorf("typoPropType reads %q", propType.Diagnostics[0].Message.Description)
	}

	chain := runNoTypos(t,
		"import PropTypes from 'prop-types';\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PropTypes.string.isRequird };\n")
	rule_testing.ExpectFindings(t, chain, "typoPropTypeChain")
	if !strings.Contains(chain.Diagnostics[0].Message.Description, "`isRequird` is not it") {
		t.Errorf("typoPropTypeChain reads %q", chain.Diagnostics[0].Message.Description)
	}

	// The two-slot message. A format string reusing one value in both slots would read
	// "ComponentDidMount calls ComponentDidMount", which is exactly the defect the port brief
	// records from a rule whose 37 fixtures all agreed with a wrong message.
	lifecycle := runNoTypos(t,
		"declare const React: any;\nclass Foo extends React.Component { ComponentDidMount() {} }\n")
	rule_testing.ExpectFindings(t, lifecycle, "typoLifecycleMethod")
	description := lifecycle.Diagnostics[0].Message.Description
	if !strings.Contains(description, "`ComponentDidMount` is spelled with the wrong casing") {
		t.Errorf("typoLifecycleMethod is missing the actual name: %q", description)
	}
	if !strings.Contains(description, "React calls `componentDidMount`") {
		t.Errorf("typoLifecycleMethod is missing the expected name: %q", description)
	}

	static := runNoTypos(t,
		"declare const React: any;\nclass Foo extends React.Component { getDerivedStateFromProps() {} }\n")
	rule_testing.ExpectFindings(t, static, "staticLifecycleMethod")
	if !strings.Contains(static.Diagnostics[0].Message.Description, "`getDerivedStateFromProps` is a static lifecycle method") {
		t.Errorf("staticLifecycleMethod reads %q", static.Diagnostics[0].Message.Description)
	}
}

// TestNoTyposComponentGate pins what counts as a component on each arm.
//
// A plain class is not one, and only `createReactClass` counts on the object side: the shelf's
// `react.IsEs5ComponentCall` also accepts `createClass` and the namespaced spelling, and upstream's
// `isES5Component` accepts neither because its createClass pragma defaults to `createReactClass`.
func TestNoTyposComponentGate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a class with no heritage is not a component", "class Foo { static PropTypes = {}; }\n", nil},
		{"a foreign base is not a component", "declare const Other: any;\nclass Foo extends Other.Component { static PropTypes = {}; }\n", nil},
		{"a React.Component class is", "declare const React: any;\nclass Foo extends React.Component { static PropTypes = {}; }\n", []string{"typoStaticClassProp"}},
		{"a bare Component class is", "declare const Component: any;\nclass Foo extends Component { static PropTypes = {}; }\n", []string{"typoStaticClassProp"}},
		{"createReactClass counts", "declare function createReactClass(spec: any): any;\ncreateReactClass({ PropTypes: {} });\n", []string{"typoPropDeclaration"}},
		{"a bare createClass does not", "declare function createClass(spec: any): any;\ncreateClass({ PropTypes: {} });\n", nil},
		{"React.createClass does not", "declare const React: any;\nReact.createClass({ PropTypes: {} });\n", nil},
		// The namespaced form of the PRAGMA does count, and this port was silent on it at first.
		// Eleven corpus cases are written this way and every one of them caught it. The row below it
		// is the control that keeps the namespace load-bearing.
		{"React.createReactClass counts", "declare const React: any;\nReact.createReactClass({ PropTypes: {} });\n", []string{"typoPropDeclaration"}},
		{"a foreign namespace on the pragma does not", "declare const Other: any;\nOther.createReactClass({ PropTypes: {} });\n", nil},
		{"a bare object literal is not a component", "const x = { PropTypes: {} };\n", nil},
		{"a non-static class property is not checked", "declare const React: any;\nclass Foo extends React.Component { PropTypes = {}; }\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoTypos(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoTyposTwoIdsForTheSameCasingError pins which id each surface produces.
//
// A class reports `typoStaticClassProp` and a createReactClass object reports `typoPropDeclaration`,
// and the two carry different text. A port collapsing them would pass every count assertion.
func TestNoTyposTwoIdsForTheSameCasingError(t *testing.T) {
	t.Parallel()

	class := runNoTypos(t,
		"declare const React: any;\nclass Foo extends React.Component { static PropTypes = {}; }\n")
	rule_testing.ExpectFindings(t, class, "typoStaticClassProp")
	if !strings.HasPrefix(class.Diagnostics[0].Message.Description, "This static property is spelled") {
		t.Errorf("class reads %q", class.Diagnostics[0].Message.Description)
	}

	object := runNoTypos(t,
		"declare function createReactClass(spec: any): any;\ncreateReactClass({ PropTypes: {} });\n")
	rule_testing.ExpectFindings(t, object, "typoPropDeclaration")
	if !strings.HasPrefix(object.Diagnostics[0].Message.Description, "This property is spelled") {
		t.Errorf("object reads %q", object.Diagnostics[0].Message.Description)
	}
}

// TestNoTyposNeedsNoChecker pins that this rule is syntactic.
func TestNoTyposNeedsNoChecker(t *testing.T) {
	t.Parallel()

	const source = "declare const React: any;\nclass Foo extends React.Component { static PropTypes = {}; }\n"
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoTypos, noTyposFile, source), "typoStaticClassProp")
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoTypos, noTyposFile, source), "typoStaticClassProp")
}

// TestNoTyposHasNoFileGate reports in a `.ts` file, not only in `.tsx`.
//
// Three shipped rules in this package carried a `.tsx`-only gate that was oxc residue, each with a
// test asserting the gate, so the suite locked the bug in. This pins the opposite. The source
// carries no JSX, because JSX in a `.ts` file is a syntax error and a case that failed to parse
// would be silent for a reason unrelated to any gate.
func TestNoTyposHasNoFileGate(t *testing.T) {
	t.Parallel()

	const source = "declare const React: any;\nclass Foo extends React.Component { static PropTypes = {}; }\n"
	for _, fileName := range []string{
		"/repository/source/NoTypos.tsx",
		"/repository/source/NoTypos.ts",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoTypos, fileName, source), "typoStaticClassProp")
		})
	}
}

// TestNoTyposSurvivesShapesThatWouldPanic drives shapes with pieces missing.
//
// The walk recovers per FILE rather than per rule, so one nil dereference here takes the file away
// from every rule in the tree, and no `ExpectFindings` fixture can see a panic.
func TestNoTyposSurvivesShapesThatWouldPanic(t *testing.T) {
	t.Parallel()

	sources := []string{
		"import 'prop-types';\nimport 'react';\n",
		"import * as PropTypes from 'prop-types';\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PropTypes.strng };\n",
		"declare function createReactClass(spec: any): any;\ncreateReactClass();\n",
		"declare function createReactClass(spec: any): any;\ndeclare const spread: any;\ncreateReactClass({ ...spread });\n",
		"import PropTypes from 'prop-types';\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PropTypes.shape() };\n",
		"import PropTypes from 'prop-types';\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { a: PropTypes.oneOfType(notAnArray) };\n",
		"import PropTypes from 'prop-types';\ndeclare const React: any;\nclass Foo extends React.Component {}\nFoo.propTypes = { [computed]: PropTypes.strng };\n",
		"declare const React: any;\nclass Foo extends React.Component { #secret() {} }\n",
		"declare const React: any;\ndeclare const key: string;\nclass Foo extends React.Component { [key]() {} }\n",
		"(()=>{})();\n",
	}

	for index, sourceText := range sources {
		t.Run(string(rune('a'+index)), func(t *testing.T) {
			t.Parallel()
			runNoTypos(t, sourceText)
		})
	}
}

// TestNoTyposMemberArmAssignmentShapes pins what upstream actually requires around the member.
//
// Upstream tests `node.parent.type === 'AssignmentExpression' && node.parent.right` and nothing
// else: no left-side check, no operator check. This port had both at first, and a surviving mutant
// for each is what sent me to measure rather than reason. Two of my three checks were wrong and the
// mutants were right.
//
// The right-hand row is the surprising one: merely READING `target = Foo.PropTypes` reports, as if
// the property had been written. That reads as an upstream oversight, and it is reproduced because
// narrowing it would silence an input upstream reports and no imported fixture would notice.
//
// Every verdict below was measured on the installed build.
func TestNoTyposMemberArmAssignmentShapes(t *testing.T) {
	t.Parallel()

	const preamble = "declare const React: any;\nclass Foo extends React.Component {}\ndeclare let target: any;\ndeclare function doThing(value: any): void;\n"
	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a plain assignment reports", preamble + "Foo.PropTypes = {};\n", []string{"typoStaticClassProp"}},
		{"a member on the RIGHT of an assignment also reports", preamble + "target = Foo.PropTypes;\n", []string{"typoStaticClassProp"}},
		{"a compound assignment reports", preamble + "Foo.PropTypes ||= {};\n", []string{"typoStaticClassProp"}},
		{"a nullish assignment reports", preamble + "Foo.PropTypes ??= {};\n", []string{"typoStaticClassProp"}},
		{"a bare read is silent", preamble + "Foo.PropTypes;\n", nil},
		{"a comparison is silent", preamble + "if (Foo.PropTypes === 1) {}\n", nil},
		{"an argument is silent", preamble + "doThing(Foo.PropTypes);\n", nil},
		{"a correctly spelled assignment is silent", preamble + "Foo.propTypes = {};\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoTypos(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}
