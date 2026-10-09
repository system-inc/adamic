package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noDeprecatedFile is where the fixtures pretend to live.
//
// A `.tsx` name because several cases write JSX and need a parser that reads it. It is NOT
// load-bearing: this rule has no file gate, and `TestNoDeprecatedHasNoFileGate` pins that by
// running one reporting source under four extensions.
const noDeprecatedFile = "/repository/source/NoDeprecated.tsx"

// The corpus is `eslint-plugin-react`'s own, extracted rather than retyped.
//
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/no-deprecated.js` holds 22 valid and 33
// invalid cases. Every string below was pulled out of that file by loading it with a stubbed
// `RuleTester` and serializing the captured object, then verified byte against byte by extracting
// the same file a second time through an independent path that writes raw bytes rather than JSON.
// All 55 codes were byte-identical across the two extractions.
//
// All 55 were then run against the installed build, version 7.37.5, through the ESLint Linter API,
// and its verdict agreed with the corpus on every one: 22 clean and 33 reporting with the stated
// counts. So the two authorities are one here, which is not true of every rule in this batch.
//
// # Six cases carry `settings`, and every one of them moves
//
// Upstream configures a React version on five cases and a pragma on two. cohere reads
// no React settings, so neither can reach a rule here by any route, and the faithful
// reading is to reproduce the answer upstream gives when nothing is configured. Each of the six was
// re-run against the installed build with its settings REMOVED, and the verdict below is that run:
//
//	valid-11   version 0.11.0     -> reports; unconfigured means 999.999.999
//	valid-12   version 15.4.0     -> reports
//	valid-13   version 15.4.0     -> stays clean; a bare `PropTypes` matches no key either way
//	valid-14   version 16.8.0     -> reports 3
//	valid-16   version 17.999.999 -> reports 6
//	invalid-1  pragma Foo         -> goes clean; `Foo.renderComponent` is nothing without the pragma
//	invalid-12 pragma Foo         -> goes clean, same reason
//
// Those seven lines are the whole of this rule's divergence from its corpus, and every one of them
// is the version gate or the pragma being absent rather than the judgment differing.
//
// There are no `options` and no `output` columns because `meta.schema` is `[]` and `meta.fixable`
// is unset.

// TestNoDeprecatedFires asserts the RENDERED message, not the id.
//
// This rule has exactly one `messages` entry and interpolates four slots into it, so the id
// separates nothing: every finding in this file carries `deprecated`. A mutation that moved a slot,
// dropped the `, use X instead` clause, or printed the wrong version would leave the id and the
// count fixed and be invisible to `ExpectFindings`. The whole string is compared for that reason.
func TestNoDeprecatedFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		descriptions []string
	}{
		{"upstream valid-11", "React.renderComponent()", []string{"React.renderComponent is deprecated since React 0.12.0, use React.render instead"}},              // upstream sets settings; cohere reads no React settings, so this is the unconfigured answer
		{"upstream valid-12", "React.createClass()", []string{"React.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead"}}, // upstream sets settings; cohere reads no React settings, so this is the unconfigured answer
		{"upstream valid-14", "\n        class Foo extends React.Component {\n          componentWillMount() {}\n          componentWillReceiveProps() {}\n          componentWillUpdate() {}\n        }\n      ", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillReceiveProps is deprecated since React 16.9.0, use UNSAFE_componentWillReceiveProps instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillreceiveprops. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillUpdate is deprecated since React 16.9.0, use UNSAFE_componentWillUpdate instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillupdate. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},                                                                                                               // upstream sets settings; cohere reads no React settings, so this is the unconfigured answer
		{"upstream valid-16", "\n        import { render, hydrate } from 'react-dom';\n        import { renderToNodeStream } from 'react-dom/server';\n        ReactDOM.render(element, container);\n        ReactDOM.unmountComponentAtNode(container);\n        ReactDOMServer.renderToNodeStream(element);\n      ", []string{"ReactDOM.render is deprecated since React 18.0.0, use createRoot instead, see https://reactjs.org/link/switch-to-createroot", "ReactDOM.hydrate is deprecated since React 18.0.0, use hydrateRoot instead, see https://reactjs.org/link/switch-to-createroot", "ReactDOMServer.renderToNodeStream is deprecated since React 18.0.0, use renderToPipeableStream instead, see https://reactjs.org/docs/react-dom-server.html#rendertonodestream", "ReactDOM.render is deprecated since React 18.0.0, use createRoot instead, see https://reactjs.org/link/switch-to-createroot", "ReactDOM.unmountComponentAtNode is deprecated since React 18.0.0, use root.unmount instead, see https://reactjs.org/link/switch-to-createroot", "ReactDOMServer.renderToNodeStream is deprecated since React 18.0.0, use renderToPipeableStream instead, see https://reactjs.org/docs/react-dom-server.html#rendertonodestream"}}, // upstream sets settings; cohere reads no React settings, so this is the unconfigured answer
		{"upstream invalid-0", "React.renderComponent()", []string{"React.renderComponent is deprecated since React 0.12.0, use React.render instead"}},
		{"upstream invalid-2", "/** @jsx Foo */ Foo.renderComponent()", []string{"Foo.renderComponent is deprecated since React 0.12.0, use Foo.render instead"}},
		{"upstream invalid-3", "this.transferPropsTo()", []string{"this.transferPropsTo is deprecated since React 0.12.0, use spread operator ({...}) instead"}},
		{"upstream invalid-4", "React.addons.TestUtils", []string{"React.addons.TestUtils is deprecated since React 15.5.0, use ReactDOM.TestUtils instead"}},
		{"upstream invalid-5", "React.addons.classSet()", []string{"React.addons.classSet is deprecated since React 0.13.0, use the npm module classnames instead"}},
		{"upstream invalid-6", "React.render(element, container);", []string{"React.render is deprecated since React 0.14.0, use ReactDOM.render instead"}},
		{"upstream invalid-7", "React.unmountComponentAtNode(container);", []string{"React.unmountComponentAtNode is deprecated since React 0.14.0, use ReactDOM.unmountComponentAtNode instead"}},
		{"upstream invalid-8", "React.findDOMNode(instance);", []string{"React.findDOMNode is deprecated since React 0.14.0, use ReactDOM.findDOMNode instead"}},
		{"upstream invalid-9", "React.renderToString(element);", []string{"React.renderToString is deprecated since React 0.14.0, use ReactDOMServer.renderToString instead"}},
		{"upstream invalid-10", "React.renderToStaticMarkup(element);", []string{"React.renderToStaticMarkup is deprecated since React 0.14.0, use ReactDOMServer.renderToStaticMarkup instead"}},
		{"upstream invalid-11", "React.createClass({});", []string{"React.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead"}},
		{"upstream invalid-13", "React.PropTypes", []string{"React.PropTypes is deprecated since React 15.5.0, use the npm module prop-types instead"}},
		{"upstream invalid-14", "var {createClass} = require('react');", []string{"React.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead"}},
		{"upstream invalid-15", "var {createClass, PropTypes} = require('react');", []string{"React.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead", "React.PropTypes is deprecated since React 15.5.0, use the npm module prop-types instead"}},
		{"upstream invalid-16", "import {createClass} from 'react';", []string{"React.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead"}},
		{"upstream invalid-17", "import {createClass, PropTypes} from 'react';", []string{"React.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead", "React.PropTypes is deprecated since React 15.5.0, use the npm module prop-types instead"}},
		{"upstream invalid-18", "\n      import React from 'react';\n      const {createClass, PropTypes} = React;\n    ", []string{"React.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead", "React.PropTypes is deprecated since React 15.5.0, use the npm module prop-types instead"}},
		{"upstream invalid-19", "import {printDOM} from 'react-addons-perf';", []string{"ReactPerf.printDOM is deprecated since React 15.0.0, use ReactPerf.printOperations instead"}},
		{"upstream invalid-20", "\n        import ReactPerf from 'react-addons-perf';\n        const {printDOM} = ReactPerf;\n      ", []string{"ReactPerf.printDOM is deprecated since React 15.0.0, use ReactPerf.printOperations instead"}},
		{"upstream invalid-21", "React.DOM.div", []string{"React.DOM is deprecated since React 15.6.0, use the npm module react-dom-factories instead"}},
		{"upstream invalid-22", "\n        class Bar extends React.PureComponent {\n          componentWillMount() {}\n          componentWillReceiveProps() {}\n          componentWillUpdate() {}\n        };\n      ", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillReceiveProps is deprecated since React 16.9.0, use UNSAFE_componentWillReceiveProps instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillreceiveprops. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillUpdate is deprecated since React 16.9.0, use UNSAFE_componentWillUpdate instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillupdate. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"upstream invalid-23", "\n        function Foo() {\n          return class Bar extends React.PureComponent {\n            componentWillMount() {}\n            componentWillReceiveProps() {}\n            componentWillUpdate() {}\n          };\n        }\n      ", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillReceiveProps is deprecated since React 16.9.0, use UNSAFE_componentWillReceiveProps instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillreceiveprops. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillUpdate is deprecated since React 16.9.0, use UNSAFE_componentWillUpdate instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillupdate. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"upstream invalid-24", "\n        class Bar extends PureComponent {\n          componentWillMount() {}\n          componentWillReceiveProps() {}\n          componentWillUpdate() {}\n        };\n      ", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillReceiveProps is deprecated since React 16.9.0, use UNSAFE_componentWillReceiveProps instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillreceiveprops. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillUpdate is deprecated since React 16.9.0, use UNSAFE_componentWillUpdate instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillupdate. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"upstream invalid-25", "\n        class Foo extends React.Component {\n          componentWillMount() {}\n          componentWillReceiveProps() {}\n          componentWillUpdate() {}\n        }\n      ", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillReceiveProps is deprecated since React 16.9.0, use UNSAFE_componentWillReceiveProps instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillreceiveprops. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillUpdate is deprecated since React 16.9.0, use UNSAFE_componentWillUpdate instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillupdate. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"upstream invalid-26", "\n        class Foo extends Component {\n          componentWillMount() {}\n          componentWillReceiveProps() {}\n          componentWillUpdate() {}\n        }\n      ", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillReceiveProps is deprecated since React 16.9.0, use UNSAFE_componentWillReceiveProps instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillreceiveprops. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillUpdate is deprecated since React 16.9.0, use UNSAFE_componentWillUpdate instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillupdate. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"upstream invalid-27", "\n        var Foo = createReactClass({\n          componentWillMount: function() {},\n          componentWillReceiveProps: function() {},\n          componentWillUpdate: function() {}\n        })\n      ", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillReceiveProps is deprecated since React 16.9.0, use UNSAFE_componentWillReceiveProps instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillreceiveprops. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillUpdate is deprecated since React 16.9.0, use UNSAFE_componentWillUpdate instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillupdate. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"upstream invalid-28", "\n        class Foo extends React.Component {\n          constructor() {}\n          componentWillMount() {}\n          componentWillReceiveProps() {}\n          componentWillUpdate() {}\n        }\n      ", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillReceiveProps is deprecated since React 16.9.0, use UNSAFE_componentWillReceiveProps instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillreceiveprops. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.", "componentWillUpdate is deprecated since React 16.9.0, use UNSAFE_componentWillUpdate instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillupdate. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"upstream invalid-29", "\n        import { render } from 'react-dom';\n        ReactDOM.render(<div></div>, container);\n      ", []string{"ReactDOM.render is deprecated since React 18.0.0, use createRoot instead, see https://reactjs.org/link/switch-to-createroot", "ReactDOM.render is deprecated since React 18.0.0, use createRoot instead, see https://reactjs.org/link/switch-to-createroot"}},
		{"upstream invalid-30", "\n        import { hydrate } from 'react-dom';\n        ReactDOM.hydrate(<div></div>, container);\n      ", []string{"ReactDOM.hydrate is deprecated since React 18.0.0, use hydrateRoot instead, see https://reactjs.org/link/switch-to-createroot", "ReactDOM.hydrate is deprecated since React 18.0.0, use hydrateRoot instead, see https://reactjs.org/link/switch-to-createroot"}},
		{"upstream invalid-31", "\n        import { unmountComponentAtNode } from 'react-dom';\n        ReactDOM.unmountComponentAtNode(container);\n      ", []string{"ReactDOM.unmountComponentAtNode is deprecated since React 18.0.0, use root.unmount instead, see https://reactjs.org/link/switch-to-createroot", "ReactDOM.unmountComponentAtNode is deprecated since React 18.0.0, use root.unmount instead, see https://reactjs.org/link/switch-to-createroot"}},
		{"upstream invalid-32", "\n        import { renderToNodeStream } from 'react-dom/server';\n        ReactDOMServer.renderToNodeStream(element);\n      ", []string{"ReactDOMServer.renderToNodeStream is deprecated since React 18.0.0, use renderToPipeableStream instead, see https://reactjs.org/docs/react-dom-server.html#rendertonodestream", "ReactDOMServer.renderToNodeStream is deprecated since React 18.0.0, use renderToPipeableStream instead, see https://reactjs.org/docs/react-dom-server.html#rendertonodestream"}},

		// --- cases upstream does not cover, each measured against the installed build 7.37.5 ---

		// Two findings from one expression: the listener fires on the outer member and again on the
		// inner `React.PropTypes`, and both slices are keys.
		{"m5", "React.PropTypes.component", []string{"React.PropTypes.component is deprecated since React 0.12.0, use React.PropTypes.element instead", "React.PropTypes is deprecated since React 15.5.0, use the npm module prop-types instead"}},

		// The two entries with NO replacement. Their message must end at the version with no trailing
		// comma, which is the clause upstream builds in the report call rather than in the template.
		{"m6", "React.isValidClass()", []string{"React.isValidClass is deprecated since React 0.12.0"}},
		{"m7", "React.addons.LinkedStateMixin", []string{"React.addons.LinkedStateMixin is deprecated since React 15.0.0"}},

		// The only route by which `react-addons-perf`'s second module name is reachable: the bare
		// identifier branch of `getReactModuleName`. The import arm always takes the first name.
		{"m8", "var {printDOM} = Perf;", []string{"Perf.printDOM is deprecated since React 15.0.0, use Perf.printOperations instead"}},
		{"m9", "var {printDOM} = ReactPerf;", []string{"ReactPerf.printDOM is deprecated since React 15.0.0, use ReactPerf.printOperations instead"}},

		// An alias reports under the IMPORTED name, not the local one, and the span covers the whole
		// specifier including the alias.
		{"m10", "import {createClass as cc} from 'react';", []string{"React.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead"}},

		// The destructuring twin of the above: the KEY decides, and the span is the whole property.
		{"m11", "var {createClass: cc} = require('react');", []string{"React.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead"}},

		// A rest element is filtered out and the named sibling still reports.
		{"m12", "var {createClass, ...rest} = require('react');", []string{"React.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead"}},

		// A getter, a static, and a field holding a number all report. This arm tests the NAME only.
		{"m19", "class F extends React.Component { get componentWillMount() {} }", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"m20", "class F extends React.Component { static componentWillMount() {} }", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"m21", "class F extends React.Component { componentWillMount = 1 }", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},

		// All four paren spellings on the heritage clause report. espree produces no parenthesized
		// node so upstream never sees them; our parser does, which is why both positions are skipped.
		{"m22", "class F extends (React.Component) { componentWillMount() {} }", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"m23", "class F extends (React).Component { componentWillMount() {} }", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"m24", "class F extends ((React.Component)) { componentWillMount() {} }", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
		{"m25", "class F extends (Component) { componentWillMount() {} }", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},

		// The ES5 factory callee in parentheses, same mechanism.
		{"m26", "var F = (createReactClass)({ componentWillMount: function() {} })", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},

		// The namespaced factory, which IS `React.createReactClass` and not `React.createClass`.
		{"m28", "var F = React.createReactClass({ componentWillMount: function() {} })", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},

		// The `@jsx` pragma renames React for the whole file.
		{"m29", "/** @jsx Foo */ Foo.renderComponent()", []string{"Foo.renderComponent is deprecated since React 0.12.0, use Foo.render instead"}},

		// A LINE comment carries the pragma too; the scan is not restricted to block comments.
		{"m31", "// @jsx Foo\nFoo.renderComponent()", []string{"Foo.renderComponent is deprecated since React 0.12.0, use Foo.render instead"}},

		// Only the part before the first dot is taken.
		{"m32", "/** @jsx Foo.bar */ Foo.renderComponent()", []string{"Foo.renderComponent is deprecated since React 0.12.0, use Foo.render instead"}},

		// Position in the file does not matter; every comment is scanned.
		{"m33", "Foo.renderComponent()\n/** @jsx Foo */", []string{"Foo.renderComponent is deprecated since React 0.12.0, use Foo.render instead"}},

		// A pragma that is not a valid identifier falls back to React, so this reports as React.
		{"m34", "/** @jsx 9bad */ React.renderComponent()", []string{"React.renderComponent is deprecated since React 0.12.0, use React.render instead"}},

		// The pragma name written flush against the block comment's closing delimiter.
		//
		// This is the input that separates stripping the trailing `*/` from not stripping it. ESLint
		// hands a rule the comment's VALUE, delimiters already removed, so upstream's `[^\s]+`
		// captures `Foo`. Keeping the delimiter would capture `Foo*/`, which fails the identifier
		// check and falls back to React, and this case would go silent while its React twin below
		// started reporting. Measured on the installed build: this reports as `Foo.renderComponent`.
		//
		// Written for a surviving mutant that removed the `*/` strip, and it kills it. The matching
		// `//` strip has no such input and is argued at its own line rather than tested.
		{"block pragma flush against the closing delimiter", "/**@jsx Foo*/ Foo.renderComponent()", []string{"Foo.renderComponent is deprecated since React 0.12.0, use Foo.render instead"}},

		// No whitespace after `@jsx` does not match at all, so the fallback applies.
		{"m35", "/**@jsxFoo*/ React.renderComponent()", []string{"React.renderComponent is deprecated since React 0.12.0, use React.render instead"}},

		// A spread in the object literal does not stop the named sibling from reporting.
		{"m40", "var F = createReactClass({ ...x, componentWillMount: function() {} })", []string{"componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDeprecated, noDeprecatedFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.descriptions) {
				t.Fatalf("want %d findings, got %d: %v", len(testCase.descriptions),
					len(result.Diagnostics), result.MessageIds())
			}
			for index, diagnostic := range result.Diagnostics {
				if diagnostic.Message.Id != "deprecated" {
					t.Errorf("finding %d: want id deprecated, got %q", index, diagnostic.Message.Id)
				}
				if diagnostic.Message.Description != testCase.descriptions[index] {
					t.Errorf("finding %d description:\n got %q\nwant %q", index,
						diagnostic.Message.Description, testCase.descriptions[index])
				}
			}
		})
	}
}

func TestNoDeprecatedStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"upstream valid-0", "var element = React.createElement('p', {}, null);"},
		{"upstream valid-1", "var clone = React.cloneElement(element);"},
		{"upstream valid-2", "ReactDOM.cloneElement(child, container);"},
		{"upstream valid-3", "ReactDOM.findDOMNode(instance);"},
		{"upstream valid-4", "ReactDOM.createPortal(child, container);"},
		{"upstream valid-5", "ReactDOMServer.renderToString(element);"},
		{"upstream valid-6", "ReactDOMServer.renderToStaticMarkup(element);"},
		{"upstream valid-7", "\n        var Foo = createReactClass({\n          render: function() {}\n        })\n      "},
		{"upstream valid-8", "\n        var Foo = createReactClassNonReact({\n          componentWillMount: function() {},\n          componentWillReceiveProps: function() {},\n          componentWillUpdate: function() {}\n        });\n      "},
		{"upstream valid-9", "\n        var Foo = {\n          componentWillMount: function() {},\n          componentWillReceiveProps: function() {},\n          componentWillUpdate: function() {}\n        };\n      "},
		{"upstream valid-10", "\n        class Foo {\n          constructor() {}\n          componentWillMount() {}\n          componentWillReceiveProps() {}\n          componentWillUpdate() {}\n        }\n      "},
		{"upstream valid-13", "PropTypes"}, // upstream sets settings; cohere reads no React settings, so this is the unconfigured answer
		{"upstream valid-15", "\n        import React from \"react\";\n\n        let { default: defaultReactExport, ...allReactExports } = React;\n      "},
		{"upstream valid-17", "\n        import ReactDOM, { createRoot } from 'react-dom/client';\n        ReactDOM.createRoot(container);\n        const root = createRoot(container);\n        root.unmount();\n      "},
		{"upstream valid-18", "\n        import ReactDOM, { hydrateRoot } from 'react-dom/client';\n        ReactDOM.hydrateRoot(container, <App/>);\n        hydrateRoot(container, <App/>);\n      "},
		{"upstream valid-19", "\n        import ReactDOMServer, { renderToPipeableStream } from 'react-dom/server';\n        ReactDOMServer.renderToPipeableStream(<App />, {});\n        renderToPipeableStream(<App />, {});\n      "},
		{"upstream valid-20", "\n        import { renderToString } from 'react-dom/server';\n      "},
		{"upstream valid-21", "\n        const { renderToString } = require('react-dom/server');\n      "},
		{"upstream invalid-1", "Foo.renderComponent()"}, // upstream sets settings; cohere reads no React settings, so this is the unconfigured answer
		{"upstream invalid-12", "Foo.createClass({});"}, // upstream sets settings; cohere reads no React settings, so this is the unconfigured answer

		// --- cases upstream does not cover, each measured against the installed build 7.37.5 ---
		{"m1", "React . renderComponent()"},
		{"m2", "React['renderComponent']()"},
		{"m3", "React?.renderComponent()"},
		{"m4", "React/*x*/.renderComponent()"},

		// A computed key in the destructuring pattern. `property.key.name` is undefined for it.
		{"m13", "var {['createClass']: c} = require('react');"},

		// An array pattern, not an object pattern, so the whole arm declines.
		{"m14", "var [createClass] = require('react');"},

		// A default import lacks `imported` and is filtered out upstream.
		{"m15", "import React from 'react';"},

		// A namespace import, same filter.
		{"m16", "import * as R from 'react';"},

		// A computed member key. `getPropertyName` reads `nameNode.name`, absent here.
		{"m17", "class F extends React.Component { ['componentWillMount']() {} }"},

		// A string-literal member key, same reason: a Literal has `value` and not `name`.
		{"m18", "class F extends React.Component { 'componentWillMount'() {} }"},

		// The bare `createClass` factory. Upstream's createClass pragma defaults to `createReactClass`
		// ALONE, and the shelf helper `react.IsEs5ComponentCall` accepts both. This is the case that
		// separates them, and it is why this rule does not use the shelf helper.
		{"m27", "var F = createClass({ componentWillMount: function() {} })"},

		// Under a `Foo` pragma, `React.renderComponent` is nothing. The pragma REPLACES rather than adds.
		{"m30", "/** @jsx Foo */ React.renderComponent()"},

		// The twin of the flush-delimiter case above. Under a `Foo` pragma however it is spelled,
		// `React.renderComponent` matches no key. If the `*/` strip were dropped the pragma would
		// fall back to React and this case would start reporting, so the pair pins the strip from
		// both sides.
		{"block pragma flush, React silent under it", "/**@jsx Foo*/ React.renderComponent()"},

		// A class with no heritage is not a component.
		{"m36", "class F { componentWillMount() {} }"},

		// A plain object literal is not a component; it must be the argument to the factory.
		{"m37", "var F = { componentWillMount: function() {} }"},

		// The UNSAFE_ prefixed spelling is the replacement, not a deprecation.
		{"m38", "class F extends React.Component { UNSAFE_componentWillMount() {} }"},

		// An identifier that names no known module.
		{"m39", "var {createClass} = Whatever;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDeprecated, noDeprecatedFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoDeprecatedSpans pins WHERE each of the four arms points.
//
// `ExpectFindings` asserts ids and counts and nothing else, and the fires test above additionally
// asserts the rendered message, but neither can see a finding anchored on the wrong node. Each arm
// here has a plausible wrong answer that would leave both green: the member arm could point at the
// whole call, the import arm at the declaration, the destructuring arm at the key rather than the
// property, and the lifecycle arm at the method rather than its name.
//
// Every want below is the slice upstream underlines, taken from the installed build's reported
// columns rather than from reading the source.
func TestNoDeprecatedSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       []string
	}{
		{"member expression", "React.renderComponent()", []string{"React.renderComponent"}},

		// Both findings from one expression, and their spans differ, which is the clearest proof
		// that the inner node is visited rather than the outer one being reported twice.
		{"nested member expressions", "React.PropTypes.component",
			[]string{"React.PropTypes.component", "React.PropTypes"}},

		{"import specifier", "import {createClass} from 'react';", []string{"createClass"}},

		// The alias is INSIDE the span. Reporting the imported name node alone would underline
		// eleven characters where upstream underlines seventeen.
		{"aliased import specifier", "import {createClass as cc} from 'react';",
			[]string{"createClass as cc"}},

		{"destructured property", "var {createClass} = require('react');", []string{"createClass"}},

		// Same shape on the destructuring side: the whole property, alias included.
		{"aliased destructured property", "var {createClass: cc} = require('react');",
			[]string{"createClass: cc"}},

		// The KEY, not the method. Reporting the member would underline the body too.
		{"class member key", "class F extends React.Component { componentWillMount() {} }",
			[]string{"componentWillMount"}},
		{"object member key", "var F = createReactClass({ componentWillMount: function() {} });",
			[]string{"componentWillMount"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDeprecated, noDeprecatedFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("want %d findings, got %d", len(testCase.want), len(result.Diagnostics))
			}
			// The harness does not trim for `rule_testing.Run`, so the source on disk is the literal
			// above and slicing it directly is sound here. `RunTyped` would need the trim applied
			// to the expectation first; this rule declares no checker and uses the untrimmed path.
			source := testCase.sourceText
			for index, diagnostic := range result.Diagnostics {
				reported := source[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.want[index] {
					t.Errorf("finding %d span: got %q, want %q", index, reported, testCase.want[index])
				}
			}
		})
	}
}

// TestNoDeprecatedHasNoFileGate pins that the extension decides nothing.
//
// Upstream registers no filename predicate, and three sibling react rules in this tree carry a
// `.tsx`-only gate inherited from an oxc port that upstream does not have anywhere (task
// `#2abaqvt`). Writing this test is how a later reader learns the `.tsx` in `noDeprecatedFile` is
// a parser choice rather than a rule one.
func TestNoDeprecatedHasNoFileGate(t *testing.T) {
	t.Parallel()

	const source = "React.renderComponent()"
	for _, fileName := range []string{
		"/repository/source/Probe.ts",
		"/repository/source/Probe.tsx",
		"/repository/source/Probe.js",
		"/repository/source/Probe.jsx",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDeprecated, fileName, source)
			rule_testing.ExpectFindings(t, result, "deprecated")
		})
	}
}
