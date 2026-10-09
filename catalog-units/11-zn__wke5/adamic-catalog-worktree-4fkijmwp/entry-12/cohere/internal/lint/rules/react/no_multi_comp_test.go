package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is upstream's, extracted by stubbing its RuleTester and capturing what the test file
// hands it, so no case was retyped and no escape was typed by hand. All 36 cases were replayed
// against the installed build first and every one reproduced, so the expectations below are
// measured rather than transcribed.
//
// The filename is .tsx because that is what makes our parser produce JSX nodes. The absence of a
// suffix gate is pinned separately by TestNoMultiCompHasNoFileSuffixGate.
const noMultiCompFile = "/repository/source/NoMultiComp.tsx"

// TestNoMultiCompStaysSilent is every passing case upstream ships, grouped by the options it runs
// them under.
func TestNoMultiCompStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    NoMultiCompOptions
	}{
		{
			name:       "upstream valid 0 (default)",
			sourceText: "\n        var Hello = require('./components/Hello');\n        var HelloJohn = createReactClass({\n          render: function() {\n            return <Hello name=\"John\" />;\n          }\n        });\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
		},
		{
			name:       "upstream valid 1 (default)",
			sourceText: "\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
		},
		{
			name:       "upstream valid 2 (default)",
			sourceText: "\n        var Heading = createReactClass({\n          render: function() {\n            return (\n              <div>\n                {this.props.buttons.map(function(button, index) {\n                  return <Button {...button} key={index}/>;\n                })}\n              </div>\n            );\n          }\n        });\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
		},
		{
			name:       "upstream valid 3 (ignoreStateless true)",
			sourceText: "\n        function Hello(props) {\n          return <div>Hello {props.name}</div>;\n        }\n        function HelloAgain(props) {\n          return <div>Hello again {props.name}</div>;\n        }\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: true},
		},
		{
			name:       "upstream valid 4 (ignoreStateless true)",
			sourceText: "\n        function Hello(props) {\n          return <div>Hello {props.name}</div>;\n        }\n        class HelloJohn extends React.Component {\n          render() {\n            return <Hello name=\"John\" />;\n          }\n        }\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: true},
		},
		{
			name:       "upstream valid 5 (default)",
			sourceText: "\n        import React, { createElement } from \"react\"\n        const helperFoo = () => {\n          return true;\n        };\n        function helperBar() {\n          return false;\n        };\n        function RealComponent() {\n          return createElement(\"img\");\n        };\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
		},
		{
			name:       "upstream valid 6 (ignoreStateless true)",
			sourceText: "\n        const Hello = React.memo(function(props) {\n          return <div>Hello {props.name}</div>;\n        });\n        class HelloJohn extends React.Component {\n          render() {\n            return <Hello name=\"John\" />;\n          }\n        }\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: true},
		},
		{
			name:       "upstream valid 7 (ignoreStateless false)",
			sourceText: "\n        class StoreListItem extends React.PureComponent {\n          // A bunch of stuff here\n        }\n        export default React.forwardRef((props, ref) => <StoreListItem {...props} forwardRef={ref} />);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
		},
		{
			name:       "upstream valid 8 (ignoreStateless false)",
			sourceText: "\n        class StoreListItem extends React.PureComponent {\n          // A bunch of stuff here\n        }\n        export default React.forwardRef((props, ref) => {\n          return <StoreListItem {...props} forwardRef={ref} />\n        });\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
		},
		{
			name:       "upstream valid 9 (ignoreStateless false)",
			sourceText: "\n        const HelloComponent = (props) => {\n          return <div></div>;\n        }\n        export default React.forwardRef((props, ref) => <HelloComponent {...props} forwardRef={ref} />);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
		},
		{
			name:       "upstream valid 10 (ignoreStateless true)",
			sourceText: "\n        class StoreListItem extends React.PureComponent {\n          // A bunch of stuff here\n        }\n        export default React.forwardRef(\n          function myFunction(props, ref) {\n            return <StoreListItem {...props} forwardedRef={ref} />;\n          }\n        );\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: true},
		},
		{
			name:       "upstream valid 11 (ignoreStateless true)",
			sourceText: "\n        const HelloComponent = (props) => {\n          return <div></div>;\n        }\n        class StoreListItem extends React.PureComponent {\n          // A bunch of stuff here\n        }\n        export default React.forwardRef(\n          function myFunction(props, ref) {\n            return <StoreListItem {...props} forwardedRef={ref} />;\n          }\n        );\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: true},
		},
		{
			name:       "upstream valid 12 (ignoreStateless true)",
			sourceText: "\n        const HelloComponent = (props) => {\n          return <div></div>;\n        }\n        export default React.memo((props, ref) => <HelloComponent {...props} />);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: true},
		},
		{
			name:       "upstream valid 13 (default)",
			sourceText: "\n        import React from 'react';\n        function memo() {\n          var outOfScope = \"hello\"\n          return null;\n        }\n        class ComponentY extends React.Component {\n          memoCities = memo((cities) => cities.map((v) => ({ label: v })));\n          render() {\n            return (\n              <div>\n                <div>Counter</div>\n              </div>\n            );\n          }\n        }\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
		},
		{
			name:       "upstream valid 14 (default)",
			sourceText: "\n        const MenuList = forwardRef(({onClose, ...props}, ref) => {\n          const {t} = useTranslation();\n          const handleLogout = useLogoutHandler();\n\n          const onLogout = useCallback(() => {\n            onClose();\n            handleLogout();\n          }, [onClose, handleLogout]);\n\n          return (\n            <MuiMenuList ref={ref} {...props}>\n              <MuiMenuItem key=\"logout\" onClick={onLogout}>\n                {t('global-logout')}\n              </MuiMenuItem>\n            </MuiMenuList>\n          );\n        });\n\n        MenuList.displayName = 'MenuList';\n\n        MenuList.propTypes = {\n          onClose: PropTypes.func,\n        };\n\n        MenuList.defaultProps = {\n          onClose: () => null,\n        };\n\n        export default MenuList;\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
		},
		{
			name:       "upstream valid 15 (default)",
			sourceText: "\n        const MenuList = forwardRef(({ onClose, ...props }, ref) => {\n          const onLogout = useCallback(() => {\n            onClose()\n          }, [onClose])\n\n          return (\n            <BlnMenuList ref={ref} {...props}>\n              <BlnMenuItem key=\"logout\" onClick={onLogout}>\n                Logout\n              </BlnMenuItem>\n            </BlnMenuList>\n          )\n        })\n\n        MenuList.displayName = 'MenuList'\n\n        MenuList.propTypes = {\n          onClose: PropTypes.func\n        }\n\n        MenuList.defaultProps = {\n          onClose: () => null\n        }\n\n        export default MenuList\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMultiComp, noMultiCompFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoMultiCompFires is every reporting case upstream ships. The finding count per case is
// upstream's own `errors` length, which is 2 for the one file holding three components.
func TestNoMultiCompFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    NoMultiCompOptions
		wantIds    []string
	}{
		{
			name:       "upstream invalid 0 (default)",
			sourceText: "\r        var Hello = createReactClass({\r          render: function() {\r            return <div>Hello {this.props.name}</div>;\r          }\r        });\r        var HelloJohn = createReactClass({\r          render: function() {\r            return <Hello name=\"John\" />;\r          }\r        });\r      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 1 (default)",
			sourceText: "\r        class Hello extends React.Component {\r          render() {\r            return <div>Hello {this.props.name}</div>;\r          }\r        }\r        class HelloJohn extends React.Component {\r          render() {\r            return <Hello name=\"John\" />;\r          }\r        }\r        class HelloJohnny extends React.Component {\r          render() {\r            return <Hello name=\"Johnny\" />;\r          }\r        }\r      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent", "onlyOneComponent"},
		},
		{
			name:       "upstream invalid 2 (default)",
			sourceText: "\n        function Hello(props) {\n          return <div>Hello {props.name}</div>;\n        }\n        function HelloAgain(props) {\n          return <div>Hello again {props.name}</div>;\n        }\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 3 (default)",
			sourceText: "\n        function Hello(props) {\n          return <div>Hello {props.name}</div>;\n        }\n        class HelloJohn extends React.Component {\n          render() {\n            return <Hello name=\"John\" />;\n          }\n        }\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 4 (default)",
			sourceText: "\n        export default {\n          RenderHello(props) {\n            let {name} = props;\n            return <div>{name}</div>;\n          },\n          RenderHello2(props) {\n            let {name} = props;\n            return <div>{name}</div>;\n          }\n        };\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 5 (default)",
			sourceText: "\n        exports.Foo = function Foo() {\n          return <></>\n        }\n\n        exports.createSomeComponent = function createSomeComponent(opts) {\n          return function Foo() {\n            return <>{opts.a}</>\n          }\n        }\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 6 (ignoreStateless false)",
			sourceText: "\n        class StoreListItem extends React.PureComponent {\n          // A bunch of stuff here\n        }\n        export default React.forwardRef((props, ref) => <div><StoreListItem {...props} forwardRef={ref} /></div>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 7 (ignoreStateless false)",
			sourceText: "\n        const HelloComponent = (props) => {\n          return <div></div>;\n        }\n        const HelloComponent2 = React.forwardRef((props, ref) => <div></div>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 8 (ignoreStateless false)",
			sourceText: "\n        const HelloComponent = (0, (props) => {\n          return <div></div>;\n        });\n        const HelloComponent2 = React.forwardRef((props, ref) => <><HelloComponent></HelloComponent></>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 9 (ignoreStateless false)",
			sourceText: "\n        const forwardRef = React.forwardRef;\n        const HelloComponent = (0, (props) => {\n          return <div></div>;\n        });\n        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 10 (ignoreStateless false)",
			sourceText: "\n        const memo = React.memo;\n        const HelloComponent = (props) => {\n          return <div></div>;\n        };\n        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 11 (ignoreStateless false)",
			sourceText: "\n        const {forwardRef} = React;\n        const HelloComponent = (0, (props) => {\n          return <div></div>;\n        });\n        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 12 (ignoreStateless false)",
			sourceText: "\n        const {memo} = React;\n        const HelloComponent = (0, (props) => {\n          return <div></div>;\n        });\n        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 13 (ignoreStateless false)",
			sourceText: "\n        import React, { memo } from 'react';\n        const HelloComponent = (0, (props) => {\n          return <div></div>;\n        });\n        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 14 (ignoreStateless false)",
			sourceText: "\n        import {forwardRef} from 'react';\n        const HelloComponent = (0, (props) => {\n          return <div></div>;\n        });\n        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 15 (ignoreStateless false)",
			sourceText: "\n        const { memo } = require('react');\n        const HelloComponent = (0, (props) => {\n          return <div></div>;\n        });\n        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 16 (ignoreStateless false)",
			sourceText: "\n        const {forwardRef} = require('react');\n        const HelloComponent = (0, (props) => {\n          return <div></div>;\n        });\n        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 17 (ignoreStateless false)",
			sourceText: "\n        const forwardRef = require('react').forwardRef;\n        const HelloComponent = (0, (props) => {\n          return <div></div>;\n        });\n        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
		{
			name:       "upstream invalid 18 (ignoreStateless false)",
			sourceText: "\n        const memo = require('react').memo;\n        const HelloComponent = (0, (props) => {\n          return <div></div>;\n        });\n        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);\n      ",
			options:    NoMultiCompOptions{IgnoreStateless: false},
			wantIds:    []string{"onlyOneComponent"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMultiComp, noMultiCompFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoMultiCompMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite is a differential table.
//
// Every row was run against the installed `eslint-plugin-react` build on 2026-08-27 and the `want`
// column is what that build reported, not what this port produces. The rows exist because upstream's
// corpus is silent on all of them, and eight real defects in this port were found here after every
// imported fixture was already green:
//
//	the factory name surface        `React.createClass` and a bare `createClass` are NOT the factory,
//	                                so `react.IsEs5ComponentCall` is too wide for this rule
//	parenthesized heritage          `class A extends (React.Component) {}` counts, and
//	                                `react.IsEs6ComponentClass` answers false for it
//	nested pragma wrappers          `memo(forwardRef(fn))` is TWO components, not one
//	the return traversal list       a return inside a do-while or a for-of is invisible upstream,
//	                                so widening the list reported on inputs upstream leaves alone
//	named function under a property `{a: function A(){...}}` is judged by the function's name rather
//	                                than by the key, so the property arm must fall through
//	member assignment targets       `exports.foo = fn` is silent and `exports.Foo = fn` reports,
//	                                with `module.exports` exempted from the capitalization test
//	plain assignment targets        `foo = fn` is silent and `Foo = fn` reports
//	the capitalization predicate    `$foo`, `_1`, `테스트`, `_Foo` are all components upstream
//
// Six of those eight are places where a shelf helper or an obvious improvement was the wrong answer,
// which is why the table is kept rather than deleted once green: it is the only thing standing
// between this rule and a later simplification back to any of them.
func TestNoMultiCompMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		sourceText      string
		wantFindings    int
		ignoreStateless bool
	}{
		{"name Foo", "const Ref = (props) => <div/>;\nconst Foo = (props) => <div/>;\n", 1, false},
		{"name foo", "const Ref = (props) => <div/>;\nconst foo = (props) => <div/>;\n", 0, false},
		{"name _Foo", "const Ref = (props) => <div/>;\nconst _Foo = (props) => <div/>;\n", 1, false},
		{"name _foo", "const Ref = (props) => <div/>;\nconst _foo = (props) => <div/>;\n", 0, false},
		{"name __Foo", "const Ref = (props) => <div/>;\nconst __Foo = (props) => <div/>;\n", 1, false},
		{"name $foo", "const Ref = (props) => <div/>;\nconst $foo = (props) => <div/>;\n", 1, false},
		{"name $Foo", "const Ref = (props) => <div/>;\nconst $Foo = (props) => <div/>;\n", 1, false},
		{"name _1", "const Ref = (props) => <div/>;\nconst _1 = (props) => <div/>;\n", 1, false},
		{"name 테스트", "const Ref = (props) => <div/>;\nconst 테스트 = (props) => <div/>;\n", 1, false},
		{"name Émile", "const Ref = (props) => <div/>;\nconst Émile = (props) => <div/>;\n", 1, false},
		{"name Ωmega", "const Ref = (props) => <div/>;\nconst Ωmega = (props) => <div/>;\n", 1, false},
		{"returns null", "const Ref = (props) => <div/>;\nconst A = (props) => null;\n", 1, false},
		{"returns undefined", "const Ref = (props) => <div/>;\nconst A = (props) => undefined;\n", 0, false},
		{"returns number", "const Ref = (props) => <div/>;\nconst A = (props) => 5;\n", 0, false},
		{"returns conditional jsx/null", "const Ref = (props) => <div/>;\nconst A = (props) => props.x ? <div/> : null;\n", 1, false},
		{"returns logical and", "const Ref = (props) => <div/>;\nconst A = (props) => props.x && <div/>;\n", 1, false},
		{"returns nullish", "const Ref = (props) => <div/>;\nconst A = (props) => props.x ?? <div/>;\n", 1, false},
		{"returns parenthesized jsx", "const Ref = (props) => <div/>;\nconst A = (props) => { return (<div/>); };\n", 1, false},
		{"returns double parenthesized", "const Ref = (props) => <div/>;\nconst A = (props) => { return ((<div/>)); };\n", 1, false},
		{"returns createElement", "const Ref = (props) => <div/>;\nconst A = (props) => React.createElement(\"div\");\n", 1, false},
		{"returns bare createElement", "const Ref = (props) => <div/>;\nconst A = (props) => createElement(\"div\");\n", 0, false},
		{"returns fragment", "const Ref = (props) => <div/>;\nconst A = (props) => <></>;\n", 1, false},
		{"return inside if", "const Ref = (props) => <div/>;\nconst A = (props) => { if (props.x) { return <div/>; } return null; };\n", 1, false},
		{"return inside for", "const Ref = (props) => <div/>;\nconst A = (props) => { for(;;){ return <div/>; } };\n", 1, false},
		{"return inside switch", "const Ref = (props) => <div/>;\nconst A = (props) => { switch(1){ case 1: return <div/>; } };\n", 1, false},
		{"return only in NESTED fn", "const Ref = (props) => <div/>;\nconst A = (props) => { function inner(){ return <div/>; } };\n", 0, false},
		{"class extends React.Component", "class A extends React.Component { render(){return <div/>;} }\nclass B extends React.Component { render(){return <div/>;} }\n", 1, false},
		{"class extends Component bare", "import {Component} from 'react';\nclass A extends Component { render(){return <div/>;} }\nclass B extends Component { render(){return <div/>;} }\n", 1, false},
		{"class extends PureComponent", "class A extends React.PureComponent {}\nclass B extends React.PureComponent {}\n", 1, false},
		{"class no heritage", "class A { render(){return <div/>;} }\nclass B { render(){return <div/>;} }\n", 0, false},
		{"class extends parenthesized", "class A extends (React.Component) {}\nclass B extends (React.Component) {}\n", 1, false},
		{"class extends Other.Component", "class A extends Other.Component {}\nclass B extends Other.Component {}\n", 0, false},
		{"class expression", "const A = class extends React.Component {};\nconst B = class extends React.Component {};\n", 1, false},
		{"createReactClass x2", "var A = createReactClass({render(){return <div/>;}});\nvar B = createReactClass({render(){return <div/>;}});\n", 1, false},
		{"React.createClass x2", "var A = React.createClass({render(){return <div/>;}});\nvar B = React.createClass({render(){return <div/>;}});\n", 0, false},
		{"bare createClass x2", "var A = createClass({render(){return <div/>;}});\nvar B = createClass({render(){return <div/>;}});\n", 0, false},
		{"async generator fn", "const Ref = (props) => <div/>;\nasync function* A(){ return <div/>; }\n", 0, false},
		{"async fn", "const Ref = (props) => <div/>;\nasync function A(){ return <div/>; }\n", 1, false},
		{"generator fn", "const Ref = (props) => <div/>;\nfunction* A(){ return <div/>; }\n", 1, false},
		{"memo nested in forwardRef", "import {memo, forwardRef} from 'react';\nconst A = (p)=><div/>;\nconst B = memo(forwardRef((p,r)=><span/>));\n", 2, false},
		{"React.memo wrapping unknown", "class A extends React.PureComponent {}\nconst B = React.memo((p)=><Unknown/>);\n", 1, false},
		{"React.memo wrapping known", "class A extends React.PureComponent {}\nconst B = React.memo((p)=><A/>);\n", 0, false},
		{"React.memo wrapping known in div", "class A extends React.PureComponent {}\nconst B = React.memo((p)=><div><A/></div>);\n", 1, false},
		{"React.memo wrapping fragment", "class A extends React.PureComponent {}\nconst B = React.memo((p)=><></>);\n", 1, false},
		{"fn + class ignoreStateless", "function A(p){return <div/>;}\nclass B extends React.Component { render(){return <div/>;} }\n", 0, true},
		{"class + class ignoreStateless", "class A extends React.Component {}\nclass B extends React.Component {}\n", 1, true},
		{"fn + fn + class ignoreStateless", "function A(p){return <div/>;}\nfunction B(p){return <div/>;}\nclass C extends React.Component {}\n", 0, true},
		{"wrapper + class ignoreStateless", "import {memo} from 'react';\nconst A = memo((p)=><span/>);\nclass B extends React.Component {}\n", 0, true},
		{"three fn comps", "function A(){return <div/>;}\nfunction B(){return <div/>;}\nfunction C(){return <div/>;}\n", 2, false},
		{"obj methods x2", "export default { A(p){return <div/>;}, B(p){return <div/>;} };\n", 1, false},
		{"obj arrow props x2", "export default { A: (p)=><div/>, B: (p)=><div/> };\n", 1, false},
		{"obj lowercase keys", "export default { a: (p)=><div/>, b: (p)=><div/> };\n", 0, false},
		{"obj computed key", "export default { [k]: (p)=><div/>, B: (p)=><div/> };\n", 0, false},
		{"export default arrow returning null", "const Ref = (props) => <div/>;\nexport default () => null;\n", 0, false},
		{"export default arrow returning jsx", "const Ref = (props) => <div/>;\nexport default () => <div/>;\n", 1, false},
		{"memo(SomeComponent) identifier arg", "import {memo} from 'react';\nconst A=(p)=><div/>;\nconst B = memo(A);\n", 0, false},
		{"memo() no args", "import {memo} from 'react';\nconst A=(p)=><div/>;\nconst B = memo();\n", 0, false},
		{"memo(fn) fn returns nothing", "import {memo} from 'react';\nconst A=(p)=><div/>;\nconst B = memo((p)=>{});\n", 1, false},
		{"memo(fn) fn returns null", "import {memo} from 'react';\nconst A=(p)=><div/>;\nconst B = memo((p)=>null);\n", 1, false},
		{"triple nest memo(memo(forwardRef))", "import {memo, forwardRef} from 'react';\nconst A=(p)=><div/>;\nconst B = memo(memo(forwardRef((p,r)=><span/>)));\n", 2, false},
		{"React.memo(React.forwardRef(fn))", "const A=(p)=><div/>;\nconst B = React.memo(React.forwardRef((p,r)=><span/>));\n", 2, false},
		{"import memo aliased", "import {memo as m} from 'react';\nconst A=(p)=><div/>;\nconst B = m((p)=><span/>);\n", 0, false},
		{"component inside component", "function Outer(){ function Inner(){ return <div/>; } return <div/>; }\n", 1, false},
		{"arrow inside arrow returning jsx", "const Outer = () => { const Inner = () => <div/>; return <div/>; };\n", 1, false},
		{"iife component", "const Ref = (props) => <div/>;\nconst A = (function(){ return () => <div/>; })();\n", 1, false},
		{"return in try block", "const Ref = (props) => <div/>;\nconst A = (p) => { try { return <div/>; } catch(e){} };\n", 0, false},
		{"return in labeled stmt", "const Ref = (props) => <div/>;\nconst A = (p) => { lbl: { return <div/>; } };\n", 0, false},
		{"return in nested block", "const Ref = (props) => <div/>;\nconst A = (p) => { { { return <div/>; } } };\n", 1, false},
		{"return in while", "const Ref = (props) => <div/>;\nconst A = (p) => { while(1){ return <div/>; } };\n", 1, false},
		{"return in do-while", "const Ref = (props) => <div/>;\nconst A = (p) => { do { return <div/>; } while(0); };\n", 0, false},
		{"return in for-of", "const Ref = (props) => <div/>;\nconst A = (p) => { for (const x of y){ return <div/>; } };\n", 0, false},
		{"returns self-closing", "const Ref = (props) => <div/>;\nconst A = (p) => <div/>;\n", 1, false},
		{"returns nested ternary", "const Ref = (props) => <div/>;\nconst A = (p) => a ? (b ? <div/> : null) : null;\n", 1, false},
		{"returns sequence ending jsx", "const Ref = (props) => <div/>;\nconst A = (p) => (0, <div/>);\n", 1, false},
		{"returns sequence ending null", "const Ref = (props) => <div/>;\nconst A = (p) => (0, null);\n", 1, false},
		{"returns jsx member tag", "const Ref = (props) => <div/>;\nconst A = (p) => <Foo.Bar/>;\n", 1, false},
		{"class extends bare Component no import", "class A extends Component {}\nclass B extends Component {}\n", 1, false},
		{"class expression const", "const A = class extends React.Component {};\nconst B = class extends React.Component {};\n", 1, false},
		{"createReactClass non-object arg", "var A = createReactClass(x);\nvar B = createReactClass({render(){return <div/>;}});\n", 0, false},
		{"createReactClass no args", "var A = createReactClass();\nvar B = createReactClass({render(){return <div/>;}});\n", 0, false},
		{"single component ignoreStateless", "function A(p){return <div/>;}\n", 0, true},
		{"two wrappers ignoreStateless", "import {memo} from 'react';\nconst A = memo((p)=><span/>);\nconst B = memo((p)=><span/>);\n", 0, true},
		{"class+fn+fn ignoreStateless", "class A extends React.Component {}\nfunction B(p){return <div/>;}\nfunction C(p){return <div/>;}\n", 0, true},
		{"nested obj methods", "export default { outer: { A(p){return <div/>;}, B(p){return <div/>;} } };\n", 1, false},
		{"obj shorthand fn expression", "export default { A: function(){ return <div/>; }, B: function(){ return <div/>; } };\n", 1, false},
		{"obj named fn expressions lowercase", "export default { a: function A(){ return <div/>; }, b: function B(){ return <div/>; } };\n", 1, false},
		{"this.Foo = fn", "const Ref = (props) => <div/>;\nthis.Foo = function(){ return <div/>; };\n", 1, false},
		{"a.b.Foo = fn", "const Ref = (props) => <div/>;\na.b.Foo = function(){ return <div/>; };\n", 1, false},
		{"namespaced wrapper returning a member tag is not exempted", "class A extends React.PureComponent {}\nconst C = React.memo((p)=><A.B/>);\n", 1, false},
		{"namespaced wrapper returning a plain known tag is exempted", "class A extends React.PureComponent {}\nconst C = React.memo((p)=><A/>);\n", 0, false},
		{"member tag against an arrow component of the same head", "const A = (p)=><div/>;\nconst C = React.memo((p)=><A.B/>);\n", 1, false},
		{"arrow as the first operand of a comma is not the value", "const Ref=(p)=><div/>;\nconst A = ((p) => <div/>, 0);\n", 0, false},
		{"arrow as the last operand of a comma is the value", "const Ref=(p)=><div/>;\nconst A = (0, (p) => <div/>);\n", 1, false},
		{"only the last of two arrow operands counts", "const Ref=(p)=><div/>;\nconst A = ((p) => <div/>, (p) => <span/>);\n", 1, false},
		{"exports.Foo capitalized member", "const Ref = (props) => <div/>;\nexports.Foo = function () { return <div/>; };\n", 1, false},
		{"exports.foo lowercase member", "const Ref = (props) => <div/>;\nexports.foo = function () { return <div/>; };\n", 0, false},
		{"module.exports assignment", "const Ref = (props) => <div/>;\nmodule.exports = function () { return <div/>; };\n", 1, false},
		{"module.exports arrow", "const Ref = (props) => <div/>;\nmodule.exports = () => <div/>;\n", 1, false},
		{"plain Foo assignment", "const Ref = (props) => <div/>;\nFoo = function () { return <div/>; };\n", 1, false},
		{"plain foo assignment", "const Ref = (props) => <div/>;\nfoo = function () { return <div/>; };\n", 0, false},
		{"exports.Foo returning only null", "const Ref = (props) => <div/>;\nexports.Foo = function () { return null; };\n", 1, false},
		{"a.b.Foo deep member", "const Ref = (props) => <div/>;\na.b.Foo = function () { return <div/>; };\n", 1, false},
		{"a.b.foo deep member lowercase", "const Ref = (props) => <div/>;\na.b.foo = function () { return <div/>; };\n", 0, false},
		{"named fn expression name beats a lowercase key", "const Ref = (props) => <div/>;\nexports.foo = function Bar() { return <div/>; };\n", 1, false},
		{"exports.Foo = arrow", "const Ref = (props) => <div/>;\nexports.Foo = () => <div/>;\n", 1, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMultiComp, noMultiCompFile, testCase.sourceText,
				NoMultiCompOptions{IgnoreStateless: testCase.ignoreStateless})
			if len(result.Diagnostics) != testCase.wantFindings {
				t.Errorf("installed build reports %d findings, this rule reports %d",
					testCase.wantFindings, len(result.Diagnostics))
			}
		})
	}
}

// TestNoMultiCompAnchorsOnTheComponentNotTheDeclaration pins where each finding points.
//
// `ExpectFindings` asserts ids and counts and can see none of this, and the anchor is not a detail:
// upstream returns a different node per shape and the rule reports whatever it returns. A function
// declaration reports on itself, an arrow assigned to a const reports on the ARROW rather than the
// declaration, a class reports on itself, a factory reports on its object literal argument, and a
// pragma wrapper reports on the whole wrapping call.
//
// Every expected span was measured against the installed build on 2026-08-27 by slicing its
// reported range out of the source.
func TestNoMultiCompAnchorsOnTheComponentNotTheDeclaration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSpans  []string
	}{
		{
			name:       "function declaration reports on itself",
			sourceText: "function A(){ return <div/>; }\nfunction B(){ return <div/>; }\n",
			wantSpans:  []string{"function B(){ return <div/>; }"},
		},
		{
			name:       "const arrow reports on the arrow rather than the declaration",
			sourceText: "const A = (p) => <div/>;\nconst B = (p) => <div/>;\n",
			wantSpans:  []string{"(p) => <div/>"},
		},
		{
			name:       "class reports on the class",
			sourceText: "class A extends React.Component { render(){ return <div/>; } }\nclass B extends React.Component { render(){ return <div/>; } }\n",
			wantSpans:  []string{"class B extends React.Component { render(){ return <div/>; } }"},
		},
		{
			name:       "factory reports on the object argument rather than the call",
			sourceText: "var A = createReactClass({ render: function(){ return <div/>; } });\nvar B = createReactClass({ render: function(){ return <div/>; } });\n",
			wantSpans:  []string{"{ render: function(){ return <div/>; } }"},
		},
		{
			name:       "pragma wrapper reports on the whole wrapping call",
			sourceText: "import React, { memo } from 'react';\nconst A = (p) => <div/>;\nconst B = memo((p) => <A/>);\n",
			wantSpans:  []string{"memo((p) => <A/>)"},
		},
		{
			name:       "every component after the first reports, in source order",
			sourceText: "function A(){ return <div/>; }\nfunction B(){ return <div/>; }\nfunction C(){ return <div/>; }\n",
			wantSpans: []string{
				"function B(){ return <div/>; }",
				"function C(){ return <div/>; }",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMultiComp, noMultiCompFile, testCase.sourceText,
				DefaultNoMultiCompOptions())
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("want %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			// The harness writes the fixture as it received it here, since this rule does not need
			// the trim-sensitive path, but the source is sliced rather than the literal so the
			// assertion cannot drift from what was actually parsed.
			for index, diagnostic := range result.Diagnostics {
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.wantSpans[index] {
					t.Errorf("finding %d points at %q, want %q", index, reported, testCase.wantSpans[index])
				}
			}
		})
	}
}

// TestNoMultiCompMessage pins the rendered message exactly.
//
// The message has no format verbs, so there is nothing to interpolate and nothing to render wrong,
// but the text is upstream's own string and a paraphrase would diverge silently. Asserted against a
// literal typed here rather than against the rule's own constant, so the two cannot move together.
func TestNoMultiCompMessage(t *testing.T) {
	t.Parallel()

	if messageOnlyOneComponent.Id != "onlyOneComponent" {
		t.Errorf("message id is %q, want %q", messageOnlyOneComponent.Id, "onlyOneComponent")
	}
	if messageOnlyOneComponent.Description != "Declare only one React component per file" {
		t.Errorf("message description is %q, want upstream's own text",
			messageOnlyOneComponent.Description)
	}
}

// TestNoMultiCompHasNoFileSuffixGate pins the absence of the gate three siblings in this package
// carry.
//
// Those siblings were ported from oxc, which gates on the file being read as JSX. This rule's
// authority has no filename condition anywhere, and copying one by pattern-matching a neighbour
// would blind the rule to every `.ts` file in the tree. The source deliberately holds no JSX, so a
// `.ts` file parses cleanly and the parser's opinion cannot be mistaken for the rule's.
func TestNoMultiCompHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	// Only the two TypeScript suffixes are listed, and that is a fact about the harness rather than
	// about the rule. The typed harness builds a real program, and its tsconfig includes only
	// TypeScript extensions, so a `.js` or `.jsx` fixture fails with TS18003 before any rule runs.
	// The `.ts` row is the one that carries the weight regardless: it is exactly the file a `.tsx`
	// gate would have silenced, and it is where the three ported-from-oxc siblings lose findings.
	sourceText := "class A extends React.Component {}\nclass B extends React.Component {}\n"
	for _, fileName := range []string{
		"/repository/source/Suffix.tsx",
		"/repository/source/Suffix.ts",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMultiComp, fileName, sourceText,
				DefaultNoMultiCompOptions())
			rule_testing.ExpectFindings(t, result, "onlyOneComponent")
		})
	}
}

// TestDecodeNoMultiCompOptions routes configuration through the rule's own decoder.
//
// Building the options struct directly would leave the decoder untested, and the decoder is where a
// default inversion would live. The empty-input row is the one that matters: a rule configured as a
// bare `"error"` is handed nil, and this asserts that resolves to the documented default rather
// than to an error or a surprise.
func TestDecodeNoMultiCompOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		raw     string
		want    bool
		wantErr bool
	}{
		{name: "empty input is the bare error configuration", raw: "", want: false},
		{name: "explicit false", raw: `{"ignoreStateless": false}`, want: false},
		{name: "explicit true", raw: `{"ignoreStateless": true}`, want: true},
		{name: "absent key falls back to the default", raw: `{}`, want: false},
		{name: "wrong type is an error", raw: `{"ignoreStateless": "yes"}`, wantErr: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoMultiCompOptions([]byte(testCase.raw))
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("want an error, got %#v", decoded)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			options, isOptions := decoded.(NoMultiCompOptions)
			if !isOptions {
				t.Fatalf("decoder returned %T, want NoMultiCompOptions", decoded)
			}
			if options.IgnoreStateless != testCase.want {
				t.Errorf("IgnoreStateless is %v, want %v", options.IgnoreStateless, testCase.want)
			}
		})
	}
}

// TestNoMultiCompNilOptionsFallsBackToTheDefault bypasses the decoder entirely.
//
// A rule configured as a bare `"error"` can reach `Run` with nil options rather than with a decoded
// struct, and `options.(T)` on nil yields the zero value. The zero value happens to be correct here,
// which is exactly why this is pinned: nothing else in the suite would notice if the fallback were
// removed and a later option defaulted the other way.
func TestNoMultiCompNilOptionsFallsBackToTheDefault(t *testing.T) {
	t.Parallel()

	sourceText := "function A(){ return <div/>; }\nfunction B(){ return <div/>; }\n"
	result := rule_testing.RunTypedWithOptions(t, NoMultiComp, noMultiCompFile, sourceText, nil)
	rule_testing.ExpectFindings(t, result, "onlyOneComponent")
}

// TestNoMultiCompRequiresTheTypedHarness asserts the rule declines rather than reports when it has
// no checker.
//
// The rule declares NeedsTypeChecker because it resolves a bare `createElement` through this
// package's `isPragmaCreateElementCall`. A typed rule run on the plain harness receives a nil
// checker and, without the guard, would silently answer a narrower question while every quiet
// fixture passed vacuously. This fails loudly if the guard is ever removed.
func TestNoMultiCompRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	sourceText := "function A(){ return <div/>; }\nfunction B(){ return <div/>; }\n"

	// The control. On the typed harness this input reports, so a zero below means the guard fired
	// rather than that the input was uninteresting.
	typed := rule_testing.RunTypedWithOptions(t, NoMultiComp, noMultiCompFile, sourceText,
		DefaultNoMultiCompOptions())
	rule_testing.ExpectFindings(t, typed, "onlyOneComponent")

	untyped := rule_testing.RunWithOptions(t, NoMultiComp, noMultiCompFile, sourceText,
		DefaultNoMultiCompOptions())
	rule_testing.ExpectClean(t, untyped)
}

// TestNoMultiCompCustomPragmaIsNotReproduced records the one upstream case this port does not
// express, rather than deleting it.
//
// Upstream's last invalid case sets `settings.react.pragma` to `Foo` and then writes
// `import Foo, { memo, forwardRef } from 'foo'`, so the wrappers resolve through a non-React
// pragma. Our config layer has no per-rule React settings surface and the pragma is fixed to
// `React` across this whole package, so the case cannot be expressed here at all.
//
// The equivalent input under the default pragma is asserted instead, which pins that the wrapper
// machinery works while being explicit that the pragma dimension is untested. If a settings surface
// is ever added, this is the case to bring back.
func TestNoMultiCompCustomPragmaIsNotReproduced(t *testing.T) {
	t.Parallel()

	sourceText := "import React, { memo, forwardRef } from 'react';\n" +
		"const Text = forwardRef(({ text }, ref) => {\n" +
		"  return <div ref={ref}>{text}</div>;\n" +
		"})\n" +
		"const Label = memo(() => <Text />);\n"
	result := rule_testing.RunTypedWithOptions(t, NoMultiComp, noMultiCompFile, sourceText,
		DefaultNoMultiCompOptions())
	rule_testing.ExpectFindings(t, result, "onlyOneComponent")
}

// TestNoMultiCompCurriedFunctionsAreNotComponents pins a defect this rule shipped with.
//
// A function returned by another function is not a component unless it returns real JSX, and
// `statelessComponentFor` was missing that arm. The inner half of a curried pair was therefore
// detected through the outer binding's name, and `demo = () => () => null;` beside one real
// component reported a finding upstream does not.
//
// Measured against the installed build on 2026-08-27: both rows below report zero there. This
// rule's own corpus writes no curried function at all, so nothing in the imported fixture set
// could see it; it surfaced while porting `display-name`, whose corpus has five passing cases of
// exactly this shape.
func TestNoMultiCompCurriedFunctionsAreNotComponents(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			name:       "one curried function beside a real component",
			sourceText: "demo = () => () => null;\nfunction Other(){ return <div/>; }\n",
		},
		{
			name:       "two curried functions beside a real component",
			sourceText: "demo = () => () => null;\nother = () => () => null;\nfunction Other(){ return <div/>; }\n",
		},
		{
			name:       "a curried function expression",
			sourceText: "demo = function() {return function() {return null;};};\nfunction Other(){ return <div/>; }\n",
		},
		{
			name:       "a curried arrow under an object property",
			sourceText: "demo = {\n  property: () => () => null\n};\nfunction Other(){ return <div/>; }\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMultiComp, noMultiCompFile,
				testCase.sourceText, DefaultNoMultiCompOptions())
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoMultiCompCurriedInnerReturningJsxIsStillAComponent is the control for the test above.
//
// The arm declines on real JSX rather than on JSX-or-null, so an inner function that returns an
// element is still a component. Without this row, narrowing the arm to decline everything returned
// by a function would pass the test above while losing a component upstream counts.
func TestNoMultiCompCurriedInnerReturningJsxIsStillAComponent(t *testing.T) {
	t.Parallel()

	sourceText := "const make = () => (props) => <div/>;\nfunction Other(){ return <div/>; }\n"
	result := rule_testing.RunTypedWithOptions(t, NoMultiComp, noMultiCompFile, sourceText,
		DefaultNoMultiCompOptions())
	rule_testing.ExpectFindings(t, result, "onlyOneComponent")
}
