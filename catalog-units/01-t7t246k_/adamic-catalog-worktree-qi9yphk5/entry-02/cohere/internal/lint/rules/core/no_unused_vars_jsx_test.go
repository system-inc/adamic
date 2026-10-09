package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The JSX half of upstream's corpus, run under a `.tsx` filename because that is what makes the
// parser produce JSX nodes at all and what upstream's own react suite uses.
//
// These are separated from the main corpus file rather than merged into it because the extension is
// part of the case: the same source under `.ts` parses `<Foo />` as a type assertion and asserts
// something else entirely.
var noUnusedVarsUpstreamJsxClean = []string{
	"export const Foo = ({ onDismiss }) => {\n            const { remove } = useToaster();\n            return (\n                <button onClick={() => (onDismiss?.(), remove())}>x</button>\n            );\n        };",
	"export const Component = () => <button onClick={function onClick(e) { console.log(e) }} />",
	"\n        import React from 'react';\n\n        export const Foo = () => <div />;\n        ",
	"\n        // React 17 API\n        import React from 'react';\n        import ReactDOM from 'react-dom';\n\n        interface Props {}\n        const Component = React.forwardRef<HTMLElement, Props>(\n            function Component(props, ref) {\n                return <div ref={ref} {...props} />\n            }\n        );\n\n        ReactDOM.render(<Component />, document.getElementById('root'));\n        ",
	"const 테스트 = () => <div>Hello</div>; <테스트 />;",
	"const $foo = () => <div>Hello</div>; <$foo />;",
	"const _foo = () => <div>Hello</div>; <_foo />;",
	"\n            var App;\n            React.render(<App/>);\n        ",
	"\n            function foo() {\n                var App;\n                var bar = React.render(<App/>);\n                return bar;\n            };\n            foo()\n        ",
	"\n            var a = 1;\n            React.render(<img src={a} />);\n        ",
	"\n            var App;\n            function f() {\n                return <App />;\n            }\n            f();\n        ",
	"\n            var App;\n            <App.Hello />\n        ",
	"\n            class HelloMessage {};\n            <HelloMessage />\n        ",
	"\n            class HelloMessage {\n                render() {\n                var HelloMessage = <div>Hello</div>;\n                return HelloMessage;\n                }\n            };\n            <HelloMessage />\n        ",
	"\n            function foo() {\n                var App = { Foo: { Bar: {} } };\n                var bar = React.render(<App.Foo.Bar/>);\n                return bar;\n            };\n            foo()\n        ",
	"\n            function foo() {\n                var App = { Foo: { Bar: { Baz: {} } } };\n                var bar = React.render(<App.Foo.Bar.Baz/>);\n                return bar;\n            };\n            foo()\n        ",
	"\n            var object;\n            React.render(<object.Tag />);\n        ",
	"\n            var object;\n            React.render(<object.tag />);\n        ",
	"\n            var App;\n            React.render(<App/>);\n        ",
	"\n            function foo() {\n                var App;\n                var bar = React.render(<App/>);\n                return bar;\n            };\n            foo()\n        ",
	"\n            var a = 1;\n            React.render(<img src={a} />);\n        ",
	"\n            var App;\n            function f() {\n                return <App />;\n            }\n            f();\n        ",
	"\n            var App;\n            <App.Hello />\n        ",
	"\n            class HelloMessage {};\n            <HelloMessage />\n        ",
	"\n            class HelloMessage {\n                render() {\n                var HelloMessage = <div>Hello</div>;\n                return HelloMessage;\n                }\n            };\n            <HelloMessage />\n        ",
	"\n            function foo() {\n                var App = { Foo: { Bar: {} } };\n                var bar = React.render(<App.Foo.Bar/>);\n                return bar;\n            };\n            foo()\n        ",
	"\n            function foo() {\n                var App = { Foo: { Bar: { Baz: {} } } };\n                var bar = React.render(<App.Foo.Bar.Baz/>);\n                return bar;\n            };\n            foo()\n        ",
	"\n            var object;\n            React.render(<object.Tag />);\n        ",
	"\n            var object;\n            React.render(<object.tag />);\n        ",
	"\n        import { TypeA } from './interface';\n        export const a = <GenericComponent<TypeA> />;\n        ",
	"\n        const text = 'text';\n        export function Foo() {\n          return (\n            <div>\n              <input type=\"search\" size={30} placeholder={text} />\n            </div>\n          );\n        }",
	"\n                import React from 'react';\n\n                export const ComponentFoo: React.FC = () => {\n                  return <div>Foo Foo</div>;\n                };\n              ",
	"\n                import { h } from 'some-other-jsx-lib';\n\n                export const ComponentFoo: h.FC = () => {\n                  return <div>Foo Foo</div>;\n                };\n              ",
	"\n                import { Fragment } from 'react';\n\n                export const ComponentFoo: Fragment = () => {\n                  return <>Foo Foo</>;\n                };\n              ",
}

var noUnusedVarsUpstreamJsxReports = []string{
	"\n        const React = {};\n\n        export const Foo = () => <div />\n        ",
	"const foo = () => <div>Hello</div>; <foo />;",
	"\n            var App;\n              var unused;\n              React.render(<App unused=\"\"/>);\n        ",
	"\n            var App;\n              var Hello;\n              React.render(<App:Hello/>);\n        ",
	"\n            var Button;\n            var Input;\n            React.render(<Button.Input unused=\"\"/>);\n        ",
	"\n            class HelloMessage {\n                render() {\n                    var HelloMessage = <div>Hello</div>;\n                    return HelloMessage;\n                }\n            }\n        ",
	"\n            import {Hello} from 'Hello';\n            function Greetings() {\n                const Hello = require('Hello').default;\n                return <Hello />;\n            }\n            Greetings();\n        ",
	"\n            var lowercase;\n              React.render(<lowercase />);\n        ",
	"\n            function Greetings(div) {\n                return <div />;\n            }\n            Greetings();\n        ",
	"\n            var App;\n              var unused;\n              React.render(<App unused=\"\"/>);\n        ",
	"\n            var App;\n              var Hello;\n              React.render(<App:Hello/>);\n        ",
	"\n            var Button;\n            var Input;\n            React.render(<Button.Input unused=\"\"/>);\n        ",
	"\n            class HelloMessage {\n                render() {\n                    var HelloMessage = <div>Hello</div>;\n                    return HelloMessage;\n                }\n            }\n        ",
	"\n            import {Hello} from 'Hello';\n            function Greetings() {\n                const Hello = require('Hello').default;\n                return <Hello />;\n            }\n            Greetings();\n        ",
	"\n            var lowercase;\n              React.render(<lowercase />);\n        ",
	"\n            function Greetings(div) {\n                return <div />;\n            }\n            Greetings();\n        ",
	"\n        import React from 'react';\n        import { Fragment } from 'react';\n\n        export const ComponentFoo = () => {\n          return <div>Foo Foo</div>;\n        };\n              ",
}

// TestNoUnusedVarsStaysSilentOnUpstreamJsxCleanCases and its reporting counterpart were written
// expecting to find gaps and found none: all 51 JSX cases passed on the first run, before any
// JSX-aware code existed in the rule.
//
// That is worth stating rather than quietly banking, because it is evidence about the design. The
// rule resolves through the type checker rather than pattern-matching syntax, and typescript-go
// already treats a JSX tag name as an ordinary identifier reference, so `<Foo />` reads `Foo`
// without the rule containing the word JSX anywhere. The one place JSX genuinely needed handling
// was the opposite direction: the factory import exemption, which has its own fixture and which
// these cases would not have caught.
func TestNoUnusedVarsStaysSilentOnUpstreamJsxCleanCases(t *testing.T) {
	t.Parallel()

	for _, source := range noUnusedVarsUpstreamJsxClean {
		result := rule_testing.RunTyped(t, NoUnusedVars, "a.tsx", source)
		if len(result.Diagnostics) != 0 {
			t.Errorf("want clean, got %d findings for %q", len(result.Diagnostics), source)
		}
	}
}

// TestNoUnusedVarsFiresOnUpstreamJsxReportingCases asserts every reporting JSX case fires.
//
// An exact zero-silence assertion rather than a count with a gap list, because there are no gaps
// here: 17 of 17 report. If that ever stops being true this fails and names the case.
func TestNoUnusedVarsFiresOnUpstreamJsxReportingCases(t *testing.T) {
	t.Parallel()

	for _, source := range noUnusedVarsUpstreamJsxReports {
		result := rule_testing.RunTyped(t, NoUnusedVars, "a.tsx", source)
		if len(result.Diagnostics) == 0 {
			t.Errorf("want at least one finding, got none for %q", source)
		}
	}
}

// TestNoUnusedVarsJsxDoesNotReopenTheFactoryExemption guards the interaction the JSX work is most
// likely to break.
//
// The first real-tree run of this rule produced 512 findings and roughly 470 were a React import in
// a file containing markup. Nothing in the 51 JSX corpus cases covers that shape, so a change that
// re-opened it would pass every case above.
func TestNoUnusedVarsJsxDoesNotReopenTheFactoryExemption(t *testing.T) {
	t.Parallel()

	const source = "import React from 'react';\nexport const A = () => <div />;\n"
	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.tsx", source); len(result.Diagnostics) != 0 {
		t.Errorf("the JSX factory import must stay exempt; got %d findings", len(result.Diagnostics))
	}
	// The control: a genuinely unused import in the same file still reports, so the case above
	// cannot be passing because the rule went silent on `.tsx` altogether.
	const control = "import React from 'react';\nimport { unused } from './m';\nexport const A = () => <div />;\n"
	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.tsx", control); len(result.Diagnostics) != 1 {
		t.Errorf("control: an unused import beside the factory must report; got %d",
			len(result.Diagnostics))
	}
}
