package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// preferObjectSpreadFile is where the fixtures pretend to live.
const preferObjectSpreadFile = "/repository/source/Spread.ts"

// preferObjectSpreadCase is one upstream case.
type preferObjectSpreadCase struct {
	source string
	// typescript marks a case upstream runs through a TypeScript parser fixture. Our harness parses
	// TypeScript natively, so the flag is recorded rather than acted on; it exists so the two cases
	// read as deliberate rather than as accidental TypeScript in a JavaScript corpus.
	typescript bool
	// ecmaVersionGated marks a case whose verdict comes from the tester's default ecmaVersion of
	// 2018 rather than from the rule. `globalThis` is ES2020, so at 2018 it is not a known global
	// and the case is clean for that reason alone. Our program is ES2022 and the checker knows the
	// name unconditionally, so those cases cannot be expressed here.
	ecmaVersionGated bool
	// repairNotExpressible marks a case whose FINDINGS are asserted but whose applied text cannot
	// be, and there are exactly two reasons, both structural rather than rule defects.
	//
	// A nested `Object.assign` produces two findings whose repairs overlap. Upstream applies
	// non-overlapping fixes in one pass and re-runs, so its `output` records only the outer
	// repair; this harness refuses to guess which of two overlapping edits wins, and that refusal
	// is correct. Asserting upstream's text would mean asserting a state the real pipeline never
	// produces in one pass.
	//
	// The HTML-comment case uses `<!--`, a sloppy-script legacy production TypeScript does not
	// parse. The call arrives from a recovered parse whose span ends on an identifier rather than
	// a parenthesis, and the fixer's bracket check declines rather than writing a brace over the
	// wrong character. Measured: the span ends on the `d` of `weird`.
	repairNotExpressible bool
	fixedSource          *string
	messageIds           []string
}

func preferObjectSpreadStringPointer(value string) *string { return &value }

var preferObjectSpreadCleanCases = []preferObjectSpreadCase{
	{source: "Object.assign()", typescript: false, ecmaVersionGated: false},
	{source: "let a = Object.assign(a, b)", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign(a, b)", typescript: false, ecmaVersionGated: false},
	{source: "let a = Object.assign(b, { c: 1 })", typescript: false, ecmaVersionGated: false},
	{source: "const bar = { ...foo }", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign(...foo)", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign(foo, { bar: baz })", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign({}, ...objects)", typescript: false, ecmaVersionGated: false},
	{source: "foo({ foo: 'bar' })", typescript: false, ecmaVersionGated: false},
	{source: "\n        const Object = {};\n        Object.assign({}, foo);\n        ", typescript: false, ecmaVersionGated: false},
	{source: "\n        Object = {};\n        Object.assign({}, foo);\n        ", typescript: false, ecmaVersionGated: false},
	{source: "\n        const Object = {};\n        Object.assign({ foo: 'bar' });\n        ", typescript: false, ecmaVersionGated: false},
	{source: "\n        Object = {};\n        Object.assign({ foo: 'bar' });\n        ", typescript: false, ecmaVersionGated: false},
	{source: "\n        const Object = require('foo');\n        Object.assign({ foo: 'bar' });\n        ", typescript: false, ecmaVersionGated: false},
	{source: "\n        import Object from 'foo';\n        Object.assign({ foo: 'bar' });\n        ", typescript: false, ecmaVersionGated: false},
	{source: "\n        import { Something as Object } from 'foo';\n        Object.assign({ foo: 'bar' });\n        ", typescript: false, ecmaVersionGated: false},
	{source: "\n        import { Object, Array } from 'globals';\n        Object.assign({ foo: 'bar' });\n        ", typescript: false, ecmaVersionGated: false},
	{source: "globalThis.Object.assign({}, foo)", typescript: false, ecmaVersionGated: true},
	{source: "globalThis.Object.assign({}, { foo: 'bar' })", typescript: false, ecmaVersionGated: true},
	{source: "globalThis.Object.assign({}, baz, { foo: 'bar' })", typescript: false, ecmaVersionGated: true},
	{source: "\n                var globalThis = foo;\n                globalThis.Object.assign({}, foo)\n                ", typescript: false, ecmaVersionGated: false},
	{source: "class C { #assign; foo() { Object.#assign({}, foo); } }", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign({ get a() {} }, {})", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign({ set a(val) {} }, {})", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign({ get a() {} }, foo)", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign({ set a(val) {} }, foo)", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign({ foo: 'bar', get a() {}, baz: 'quux' }, quuux)", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign({ foo: 'bar', set a(val) {} }, { baz: 'quux' })", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign({}, { get a() {} })", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign({}, { set a(val) {} })", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign({}, { foo: 'bar', get a() {} }, {})", typescript: false, ecmaVersionGated: false},
	{source: "Object.assign({ foo }, bar, {}, { baz: 'quux', set a(val) {}, quuux }, {})", typescript: false, ecmaVersionGated: false},
}

var preferObjectSpreadFiringCases = []preferObjectSpreadCase{
	{source: "Object.assign({}, foo)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({ ...foo})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign  ({}, foo)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({ ...foo})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({}, { foo: 'bar' })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({ foo: 'bar'})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({}, baz, { foo: 'bar' })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({ ...baz, foo: 'bar'})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({}, { foo: 'bar', baz: 'foo' })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({ foo: 'bar', baz: 'foo'})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({ foo: 'bar' }, baz)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({foo: 'bar', ...baz})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({ foo: 'bar' }, cats, dogs, trees, birds)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({foo: 'bar', ...cats, ...dogs, ...trees, ...birds})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({ foo: 'bar' }, Object.assign({ bar: 'foo' }, baz))", typescript: false, ecmaVersionGated: false, fixedSource: nil, repairNotExpressible: true, messageIds: []string{"useSpreadMessage", "useSpreadMessage"}},
	{source: "Object.assign({ foo: 'bar' }, Object.assign({ bar: 'foo' }, Object.assign({}, { superNested: 'butwhy' })))", typescript: false, ecmaVersionGated: false, fixedSource: nil, repairNotExpressible: true, messageIds: []string{"useSpreadMessage", "useSpreadMessage", "useSpreadMessage"}},
	{source: "Object.assign({foo: 'bar', ...bar}, baz)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({foo: 'bar', ...bar, ...baz})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({}, { foo, bar, baz })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({ foo, bar, baz})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({}, { [bar]: 'foo' })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({ [bar]: 'foo'})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({ ...bar }, { ...baz })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({...bar, ...baz})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({ ...bar }, {\n                // this is a bar\n                foo: 'bar',\n                baz: \"cats\"\n            })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({...bar, // this is a bar\n                foo: 'bar',\n                baz: \"cats\"})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({\n                boo: \"lol\",\n                // I'm a comment\n                dog: \"cat\"\n             }, {\n                // this is a bar\n                foo: 'bar',\n                baz: \"cats\"\n            })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({boo: \"lol\",\n                // I'm a comment\n                dog: \"cat\", // this is a bar\n                foo: 'bar',\n                baz: \"cats\"})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "const test = Object.assign({ ...bar }, {\n                <!-- html comment\n                foo: 'bar',\n                baz: \"cats\"\n                --> weird\n            })", typescript: false, ecmaVersionGated: false, fixedSource: nil, repairNotExpressible: true, messageIds: []string{"useSpreadMessage"}},
	{source: "const test = Object.assign({ ...bar }, {\n                foo: 'bar', // inline comment\n                baz: \"cats\"\n            })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("const test = {...bar, foo: 'bar', // inline comment\n                baz: \"cats\"}"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "const test = Object.assign({ ...bar }, {\n                /**\n                 * foo\n                 */\n                foo: 'bar',\n                baz: \"cats\"\n            })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("const test = {...bar, /**\n                 * foo\n                 */\n                foo: 'bar',\n                baz: \"cats\"}"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "const result = doSomething()\nObject.assign({}, myData)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("const result = doSomething()\n;({ ...myData})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "let a = foo + Object.assign({}, bar)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("let a = foo + ({ ...bar})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "let foo = function() {};\nfoo\nObject.assign({}, bar)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("let foo = function() {};\nfoo\n;({ ...bar})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "foo\nObject.assign({ foo: bar })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("foo\n;({foo: bar})"), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "foo();\nObject.assign({}, bar)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("foo();\n({ ...bar})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "const x = [1]\nObject.assign({}, bar)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("const x = [1]\n;({ ...bar})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "foo\nObject.assign({}, bar).doSomething()", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("foo\n;({ ...bar}).doSomething()"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "foo\nObject.assign({}, bar), 2", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("foo\n;({ ...bar}), 2"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({})", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({})"), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "Object.assign({ foo: bar })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({foo: bar})"), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "\n                const foo = 'bar';\n                Object.assign({ foo: bar })\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                const foo = 'bar';\n                ({foo: bar})\n            "), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "\n                foo = 'bar';\n                Object.assign({ foo: bar })\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                foo = 'bar';\n                ({foo: bar})\n            "), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "let a = Object.assign({})", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("let a = {}"), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "let a = Object.assign({}, a)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("let a = { ...a}"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "let a = Object.assign   ({}, a)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("let a = { ...a}"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "let a = Object.assign({ a: 1 }, b)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("let a = {a: 1, ...b}"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign(  {},  a,      b,   )", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({    ...a,      ...b,   })"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({}, a ? b : {}, b => c, a = 2)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({ ...(a ? b : {}), ...(b => c), ...(a = 2)})"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "\n                const someVar = 'foo';\n                Object.assign({}, a ? b : {}, b => c, a = 2)\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                const someVar = 'foo';\n                ({ ...(a ? b : {}), ...(b => c), ...(a = 2)})\n            "), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "\n                someVar = 'foo';\n                Object.assign({}, a ? b : {}, b => c, a = 2)\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                someVar = 'foo';\n                ({ ...(a ? b : {}), ...(b => c), ...(a = 2)})\n            "), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "[1, 2, Object.assign({}, a)]", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("[1, 2, { ...a}]"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "const foo = Object.assign({}, a)", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("const foo = { ...a}"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "function foo() { return Object.assign({}, a) }", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("function foo() { return { ...a} }"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "foo(Object.assign({}, a));", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("foo({ ...a});"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "const x = { foo: 'bar', baz: Object.assign({}, a) }", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("const x = { foo: 'bar', baz: { ...a} }"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "\n                import Foo from 'foo';\n                Object.assign({ foo: Foo });\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                import Foo from 'foo';\n                ({foo: Foo});\n            "), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "\n                import Foo from 'foo';\n                Object.assign({}, Foo);\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                import Foo from 'foo';\n                ({ ...Foo});\n            "), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "\n                const Foo = require('foo');\n                Object.assign({ foo: Foo });\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                const Foo = require('foo');\n                ({foo: Foo});\n            "), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "\n                import { Something as somethingelse } from 'foo';\n                Object.assign({}, somethingelse);\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                import { Something as somethingelse } from 'foo';\n                ({ ...somethingelse});\n            "), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "\n                import { foo } from 'foo';\n                Object.assign({ foo: Foo });\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                import { foo } from 'foo';\n                ({foo: Foo});\n            "), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "\n                const Foo = require('foo');\n                Object.assign({}, Foo);\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                const Foo = require('foo');\n                ({ ...Foo});\n            "), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "\n                const actions = Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput,\n                    },\n                    this.props.actions\n                );\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                const actions = {\n                    onChangeInput: this.handleChangeInput,\n                    ...this.props.actions\n                };\n            "), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "\n                const actions = Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput, //\n                    },\n                    this.props.actions\n                );\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                const actions = {\n                    onChangeInput: this.handleChangeInput, //\n                    \n                    ...this.props.actions\n                };\n            "), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "\n                const actions = Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput //\n                    },\n                    this.props.actions\n                );\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                const actions = {\n                    onChangeInput: this.handleChangeInput //\n                    ,\n                    ...this.props.actions\n                };\n            "), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "\n                const actions = Object.assign(\n                    (\n                        {\n                            onChangeInput: this.handleChangeInput\n                        }\n                    ),\n                    (\n                        this.props.actions\n                    )\n                );\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                const actions = {\n                    \n                            onChangeInput: this.handleChangeInput\n                        ,\n                    ...(\n                        this.props.actions\n                    )\n                };\n            "), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "\n                eventData = Object.assign({}, eventData, { outsideLocality: `${originLocality} - ${destinationLocality}` })\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                eventData = { ...eventData, outsideLocality: `${originLocality} - ${destinationLocality}`}\n            "), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign({ });", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({});"), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "Object.assign({\n});", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({});"), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "globalThis.Object.assign({ });", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({});"), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "globalThis.Object.assign({\n});", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({});"), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "\n                function foo () { var globalThis = bar; }\n                globalThis.Object.assign({ });\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                function foo () { var globalThis = bar; }\n                ({});\n            "), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "\n                const Foo = require('foo');\n                globalThis.Object.assign({ foo: Foo });\n            ", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("\n                const Foo = require('foo');\n                ({foo: Foo});\n            "), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "Object.assign({ get a() {}, set b(val) {} })", typescript: false, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({get a() {}, set b(val) {}})"), repairNotExpressible: false, messageIds: []string{"useLiteralMessage"}},
	{source: "const obj = Object.assign<{}, Record<string, string[]>>({}, getObject());", typescript: true, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("const obj = { ...getObject()};"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
	{source: "Object.assign<{}, A>({}, foo);", typescript: true, ecmaVersionGated: false, fixedSource: preferObjectSpreadStringPointer("({ ...foo});"), repairNotExpressible: false, messageIds: []string{"useSpreadMessage"}},
}

// preferObjectSpreadGlobalDeclaration gives the fixtures a real global `Object`.
//
// The rule requires `Object` to resolve to a declaration file, which is how it tells the global
// from a shadow. The fixture harness pins `lib: ["ES2022"]`, and that lib does declare `Object`,
// so nothing extra is needed. This constant is empty rather than absent so the reasoning has a
// home: measured, `Object.assign` resolves to a declaration file under the harness's own lib, and
// a parameter, a local binding and an import named `Object` all resolve to source.
const preferObjectSpreadGlobalDeclaration = ""

// TestPreferObjectSpreadFires runs upstream's 63 invalid cases.
//
// Every one carries an `output`, so this corpus is 63 fix vectors as well as 63 judgments, and the
// repair is asserted through `rule_testing.ExpectFixedSource` rather than by eye. That helper is
// also what `TestEveryRuleShipsAFixturePair` looks for by name, so a hand-rolled comparison would
// read to the guard as no proof at all.
func TestPreferObjectSpreadFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range preferObjectSpreadFiringCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			if testCase.ecmaVersionGated {
				t.Skip("upstream runs this at the tester's default ecmaVersion 2018, where " +
					"globalThis is not a known global; our program is ES2022 and the checker " +
					"knows it unconditionally. Measured: the same input reports at 2020 and 2022 " +
					"and is clean at 2018, so this is a language-version gate rather than a rule " +
					"judgment")
			}
			source := preferObjectSpreadGlobalDeclaration + testCase.source
			result := rule_testing.RunTyped(t, PreferObjectSpread, preferObjectSpreadFile, source)
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)

			if testCase.repairNotExpressible {
				// The findings above are the assertion. See the field's own comment for the two
				// reasons the applied text cannot be compared here.
				return
			}
			if testCase.fixedSource == nil {
				for i, diagnostic := range result.Diagnostics {
					if len(diagnostic.Fixes) != 0 {
						t.Errorf("finding %d proposes a fix; upstream declines this case", i)
					}
				}
				return
			}
			// `RunTyped` writes the fixture as TrimSpace(contents)+"\n", so the expected output is
			// transformed the same way rather than compared raw. Padding the rule to make the
			// comparison line up would be the wrong repair for that mismatch.
			want := strings.TrimSpace(preferObjectSpreadGlobalDeclaration+*testCase.fixedSource) + "\n"
			rule_testing.ExpectFixedSource(t, result, want)
		})
	}
}

// TestPreferObjectSpreadStaysSilent runs upstream's 32 valid cases.
func TestPreferObjectSpreadStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range preferObjectSpreadCleanCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			if testCase.ecmaVersionGated {
				t.Skip("upstream runs this at the tester's default ecmaVersion 2018, where " +
					"globalThis is not a known global; our program is ES2022 and the checker " +
					"knows it unconditionally. Measured: the same input reports at 2020 and 2022 " +
					"and is clean at 2018, so this is a language-version gate rather than a rule " +
					"judgment")
			}
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreferObjectSpread,
				preferObjectSpreadFile, preferObjectSpreadGlobalDeclaration+testCase.source))
		})
	}
}

// TestPreferObjectSpreadAddedCases covers shapes upstream's corpus does not write.
//
// Each exists because a mutation survived all 95 imported cases, and each verdict was measured
// against the installed rule rather than derived from reading the fixer.
func TestPreferObjectSpreadAddedCases(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		source      string
		fixedSource string
		reason      string
	}{
		{
			name: "a call inside a statement it does not start needs no semicolon",
			// Upstream asks `isStartOfExpressionStatement`, and the corpus only ever writes calls
			// that DO start one, so a mutation removing the start test survived every imported
			// case. Here the call sits after an operator, so the leading paren cannot be misread
			// and no semicolon is written. Measured: upstream produces `x + ({ ...bar})`.
			source:      "foo\nx + Object.assign({}, bar)",
			fixedSource: "foo\nx + ({ ...bar})",
			reason:      "the start-of-statement test",
		},
		{
			name: "a call that starts a deeper expression still needs the semicolon",
			// The other half of the same test: the call is the leftmost part of a member access
			// which is itself the statement, so the semicolon IS written. Together these two pin
			// the walk rather than just its existence.
			source:      "foo\nObject.assign({}, bar).x",
			fixedSource: "foo\n;({ ...bar}).x",
			reason:      "the start-of-statement walk",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferObjectSpread, preferObjectSpreadFile,
				testCase.source)
			rule_testing.ExpectFindings(t, result, "useSpreadMessage")
			rule_testing.ExpectFixedSource(t, result,
				strings.TrimSpace(testCase.fixedSource)+"\n")
		})
	}
}

// TestPreferObjectSpreadDeclinesOnARecoveredParse pins the bracket sanity check.
//
// The fixer replaces the call's last character with a closing brace, which is only correct when
// that character IS the closing parenthesis. Error recovery breaks that: an HTML-like comment is a
// sloppy-script legacy production TypeScript does not parse, so the call's span ends on an
// identifier instead. Measured: for the corpus's `<!--` case the span ends on the `d` of `weird`.
//
// Without the guard the repair writes a brace over that character and corrupts the file. A
// mutation removing it survived every imported case, because the one case that reaches it is the
// same case whose applied text cannot be compared, so the finding assertion alone could not see it.
// This asserts what the finding OFFERS instead.
func TestPreferObjectSpreadDeclinesOnARecoveredParse(t *testing.T) {
	t.Parallel()

	const source = "const test = Object.assign({ ...bar }, {\n<!-- html comment\nfoo: 'bar'\n--> weird\n})"

	result := rule_testing.RunTyped(t, PreferObjectSpread, preferObjectSpreadFile, source)
	rule_testing.ExpectFindings(t, result, "useSpreadMessage")
	for i, diagnostic := range result.Diagnostics {
		if len(diagnostic.Fixes) != 0 {
			t.Errorf("finding %d proposes %d fixes on a recovered parse whose span does not end "+
				"at a parenthesis; the repair would write a brace over the wrong character",
				i, len(diagnostic.Fixes))
		}
	}
}
