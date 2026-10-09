package typescript

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noConfusingVoidExpressionFile is where the fixtures pretend to live.
const noConfusingVoidExpressionFile = "/repository/source/Void.ts"

// noConfusingVoidExpressionJsxFile is where the one JSX case pretends to live. Parsing `.tsx` as
// plain TypeScript does not fail loudly: the JSX is read as type assertions and the tree simply
// holds no JSX node, so the case would pass for the wrong reason.
const noConfusingVoidExpressionJsxFile = "/repository/source/Void.tsx"

// noConfusingVoidExpressionCase is one upstream case.
//
// `optionsJson` is raw config text rather than a built struct, so every case is routed through the
// rule's own decoder. Half of upstream's firing cases carry options, and all three of them are
// what make five of the seven reachable message ids reachable at all.
type noConfusingVoidExpressionCase struct {
	source      string
	optionsJson string
	// jsx says the case configures ecmaFeatures.jsx. Our harness decides JSX parsing from the file
	// extension rather than from an option, so this routes the case to a .tsx filename.
	jsx         bool
	fixedSource *string
	messageIds  []string
}

func noConfusingVoidExpressionStringPointer(value string) *string { return &value }

// decodeNoConfusingVoidExpressionOptionsForTest routes a case's options through the decoder.
func decodeNoConfusingVoidExpressionOptionsForTest(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return nil
	}
	decoded, err := DecodeNoConfusingVoidExpressionOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("decoding options %q: %v", optionsJson, err)
	}
	return decoded
}

var noConfusingVoidExpressionCleanCases = []noConfusingVoidExpressionCase{
	{source: "() => Math.random();", optionsJson: "", jsx: false},
	{source: "console.log('foo');", optionsJson: "", jsx: false},
	{source: "foo && console.log(foo);", optionsJson: "", jsx: false},
	{source: "foo || console.log(foo);", optionsJson: "", jsx: false},
	{source: "foo ? console.log(true) : console.log(false);", optionsJson: "", jsx: false},
	{source: "console?.log('foo');", optionsJson: "", jsx: false},
	{source: "\n() => console.log('foo');\n      ", optionsJson: "{\"ignoreArrowShorthand\":true}", jsx: false},
	{source: "\nfoo => foo && console.log(foo);\n      ", optionsJson: "{\"ignoreArrowShorthand\":true}", jsx: false},
	{source: "\nfoo => foo || console.log(foo);\n      ", optionsJson: "{\"ignoreArrowShorthand\":true}", jsx: false},
	{source: "\nfoo => (foo ? console.log(true) : console.log(false));\n      ", optionsJson: "{\"ignoreArrowShorthand\":true}", jsx: false},
	{source: "\n!void console.log('foo');\n      ", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false},
	{source: "\n+void (foo && console.log(foo));\n      ", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false},
	{source: "\n-void (foo || console.log(foo));\n      ", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false},
	{source: "\n() => void ((foo && void console.log(true)) || console.log(false));\n      ", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false},
	{source: "\nconst x = void (foo ? console.log(true) : console.log(false));\n      ", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false},
	{source: "\n!(foo && void console.log(foo));\n      ", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false},
	{source: "\n!!(foo || void console.log(foo));\n      ", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false},
	{source: "\nconst x = (foo && void console.log(true)) || void console.log(false);\n      ", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false},
	{source: "\n() => (foo ? void console.log(true) : void console.log(false));\n      ", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false},
	{source: "\nreturn void console.log('foo');\n      ", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false},
	{source: "\nfunction cool(input: string) {\n  return (console.log(input), input);\n}\n    ", optionsJson: "", jsx: false},
	{source: "\nfunction cool(input: string) {\n  return (input, console.log(input), input);\n}\n      ", optionsJson: "", jsx: false},
	{source: "\nfunction test(): void {\n  return console.log('bar');\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nconst test = (): void => {\n  return console.log('bar');\n};\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nconst test = (): void => console.log('bar');\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nfunction test(): void {\n  {\n    return console.log('foo');\n  }\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nconst obj = {\n  test(): void {\n    return console.log('foo');\n  },\n};\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nclass Foo {\n  test(): void {\n    return console.log('foo');\n  }\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nfunction test() {\n  function nestedTest(): void {\n    return console.log('foo');\n  }\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ntype Foo = () => void;\nconst test = (() => console.log()) as Foo;\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ntype Foo = {\n  foo: () => void;\n};\nconst test: Foo = {\n  foo: () => console.log(),\n};\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nconst test = {\n  foo: () => console.log(),\n} as {\n  foo: () => void;\n};\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nconst test: {\n  foo: () => void;\n} = {\n  foo: () => console.log(),\n};\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ntype Foo = {\n  foo: { bar: () => void };\n};\n\nconst test = {\n  foo: { bar: () => console.log() },\n} as Foo;\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ntype Foo = {\n  foo: { bar: () => void };\n};\n\nconst test: Foo = {\n  foo: { bar: () => console.log() },\n};\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ntype MethodType = () => void;\n\nclass App {\n  private method: MethodType = () => console.log();\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ninterface Foo {\n  foo: () => void;\n}\n\nfunction bar(): Foo {\n  return {\n    foo: () => console.log(),\n  };\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ntype Foo = () => () => () => void;\nconst x: Foo = () => () => () => console.log();\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ntype Foo = {\n  foo: () => void;\n};\n\nconst test = {\n  foo: () => console.log(),\n} as Foo;\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ntype Foo = () => void;\nconst test: Foo = () => console.log('foo');\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "const foo = <button onClick={() => console.log()} />;", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: true},
	{source: "\ndeclare function foo(arg: () => void): void;\nfoo(() => console.log());\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ndeclare function foo(arg: (() => void) | (() => string)): void;\nfoo(() => console.log());\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ndeclare function foo(arg: (() => void) | (() => string) | string): void;\nfoo(() => console.log());\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ndeclare function foo(arg: () => void | string): void;\nfoo(() => console.log());\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ndeclare function foo(options: { cb: () => void }): void;\nfoo({ cb: () => console.log() });\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nconst obj = {\n  foo: { bar: () => console.log() },\n} as {\n  foo: { bar: () => void };\n};\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nfunction test(): void & void {\n  return console.log('foo');\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ntype Foo = void;\n\ndeclare function foo(): Foo;\n\nfunction test(): Foo {\n  return foo();\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\ntype Foo = void;\nconst test = (): Foo => console.log('err');\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nconst test: () => any = (): void => console.log();\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nfunction test(): void | string {\n  return console.log('bar');\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
	{source: "\nexport function makeDate(): Date;\nexport function makeDate(m: number): void;\nexport function makeDate(m?: number): Date | void {\n  if (m !== undefined) {\n    return console.log('123');\n  }\n  return new Date();\n}\n\ndeclare const test: (cb: () => void) => void;\n\ntest((() => {\n  return console.log('123');\n}) as typeof makeDate | (() => string));\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false},
}

var noConfusingVoidExpressionFiringCases = []noConfusingVoidExpressionCase{
	{source: "\nconst x = console.log('foo');\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExpr"}},
	{source: "\nconst x = console?.log('foo');\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExpr"}},
	{source: "\nconsole.error(console.log('foo'));\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExpr"}},
	{source: "\n[console.log('foo')];\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExpr"}},
	{source: "\n({ x: console.log('foo') });\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExpr"}},
	{source: "\nvoid console.log('foo');\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExpr"}},
	{source: "\nconsole.log('foo') ? true : false;\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExpr"}},
	{source: "\n(console.log('foo') && true) || false;\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExpr"}},
	{source: "\n(cond && console.log('ok')) || console.log('error');\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExpr"}},
	{source: "\n!console.log('foo');\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExpr"}},
	{source: "\nfunction notcool(input: string) {\n  return (input, console.log(input));\n}\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExpr"}},
	{source: "() => console.log('foo');", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("() => { console.log('foo'); };"), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "foo => foo && console.log(foo);", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExprArrow"}},
	{source: "(foo: undefined) => foo && console.log(foo);", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("(foo: undefined) => { foo && console.log(foo); };"), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "foo => foo || console.log(foo);", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExprArrow"}},
	{source: "(foo: undefined) => foo || console.log(foo);", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("(foo: undefined) => { foo || console.log(foo); };"), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "(foo: void) => foo || console.log(foo);", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("(foo: void) => { foo || console.log(foo); };"), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "foo => (foo ? console.log(true) : console.log(false));", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("foo => { foo ? console.log(true) : console.log(false); };"), messageIds: []string{"invalidVoidExprArrow", "invalidVoidExprArrow"}},
	{source: "\nfunction f() {\n  return console.log('foo');\n  console.log('bar');\n}\n      ", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nfunction f() {\n  console.log('foo'); return;\n  console.log('bar');\n}\n      "), messageIds: []string{"invalidVoidExprReturn"}},
	{source: "\n        function f() {\n          console.log('foo')\n          return ['bar', 'baz'].forEach(console.log)\n          console.log('quux')\n        }\n      ", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\n        function f() {\n          console.log('foo')\n          ;['bar', 'baz'].forEach(console.log); return;\n          console.log('quux')\n        }\n      "), messageIds: []string{"invalidVoidExprReturn"}},
	{source: "\nfunction f() {\n  console.log('foo');\n  return console.log('bar');\n}\n      ", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nfunction f() {\n  console.log('foo');\n  console.log('bar');\n}\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\n        function f() {\n          console.log('foo')\n          return ['bar', 'baz'].forEach(console.log)\n        }\n      ", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\n        function f() {\n          console.log('foo')\n          ;['bar', 'baz'].forEach(console.log);\n        }\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\nconst f = () => {\n  if (cond) {\n    return console.error('foo');\n  }\n  console.log('bar');\n};\n      ", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nconst f = () => {\n  if (cond) {\n    console.error('foo'); return;\n  }\n  console.log('bar');\n};\n      "), messageIds: []string{"invalidVoidExprReturn"}},
	{source: "\nconst f = function () {\n  if (cond) return console.error('foo');\n  console.log('bar');\n};\n      ", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nconst f = function () {\n  if (cond) { console.error('foo'); return; }\n  console.log('bar');\n};\n      "), messageIds: []string{"invalidVoidExprReturn"}},
	{source: "\nconst f = function () {\n  let num = 1;\n  return num ? console.log('foo') : num;\n};\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\nconst f = function () {\n  let undef = undefined;\n  return undef ? console.log('foo') : undef;\n};\n      ", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nconst f = function () {\n  let undef = undefined;\n  undef ? console.log('foo') : undef;\n};\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\nconst f = function () {\n  let num = 1;\n  return num || console.log('foo');\n};\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\nconst f = function () {\n  let bar = void 0;\n  return bar || console.log('foo');\n};\n      ", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nconst f = function () {\n  let bar = void 0;\n  bar || console.log('foo');\n};\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\nlet num = 1;\nconst foo = () => (num ? console.log('foo') : num);\n      ", optionsJson: "", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\nlet bar = void 0;\nconst foo = () => (bar ? console.log('foo') : bar);\n      ", optionsJson: "", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nlet bar = void 0;\nconst foo = () => { bar ? console.log('foo') : bar; };\n      "), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "return console.log('foo');", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("return void console.log('foo');"), messageIds: []string{"invalidVoidExprReturnWrapVoid"}},
	{source: "console.error(console.log('foo'));", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExprWrapVoid"}},
	{source: "console.log('foo') ? true : false;", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExprWrapVoid"}},
	{source: "const x = foo ?? console.log('foo');", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExprWrapVoid"}},
	{source: "foo => foo || console.log(foo);", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("foo => foo || void console.log(foo);"), messageIds: []string{"invalidVoidExprArrowWrapVoid"}},
	{source: "!!console.log('foo');", optionsJson: "{\"ignoreVoidOperator\":true}", jsx: false, fixedSource: nil, messageIds: []string{"invalidVoidExprWrapVoid"}},
	{source: "\nfunction test() {\n  return console.log('foo');\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nfunction test() {\n  console.log('foo');\n}\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "const test = () => console.log('foo');", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("const test = () => { console.log('foo'); };"), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\nconst test = () => {\n  return console.log('foo');\n};\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nconst test = () => {\n  console.log('foo');\n};\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\nfunction foo(): void {\n  const bar = () => {\n    return console.log();\n  };\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nfunction foo(): void {\n  const bar = () => {\n    console.log();\n  };\n}\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\n(): any => console.log('foo');\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\n(): any => { console.log('foo'); };\n      "), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\n(): unknown => console.log('foo');\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\n(): unknown => { console.log('foo'); };\n      "), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\nfunction test(): void {\n  () => () => console.log();\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nfunction test(): void {\n  () => () => { console.log(); };\n}\n      "), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\ntype Foo = any;\n(): Foo => console.log();\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\ntype Foo = any;\n(): Foo => { console.log(); };\n      "), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\ntype Foo = unknown;\n(): Foo => console.log();\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\ntype Foo = unknown;\n(): Foo => { console.log(); };\n      "), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\nfunction test(): any {\n  () => () => console.log();\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nfunction test(): any {\n  () => () => { console.log(); };\n}\n      "), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\nfunction test(): unknown {\n  return console.log();\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nfunction test(): unknown {\n  console.log();\n}\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\nfunction test(): any {\n  return console.log();\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nfunction test(): any {\n  console.log();\n}\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\ntype Foo = () => any;\n(): Foo => () => console.log();\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\ntype Foo = () => any;\n(): Foo => () => { console.log(); };\n      "), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\ntype Foo = () => unknown;\n(): Foo => () => console.log();\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\ntype Foo = () => unknown;\n(): Foo => () => { console.log(); };\n      "), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\ntype Foo = () => any;\nconst test: Foo = () => console.log();\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\ntype Foo = () => any;\nconst test: Foo = () => { console.log(); };\n      "), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\ntype Foo = () => unknown;\nconst test: Foo = () => console.log();\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\ntype Foo = () => unknown;\nconst test: Foo = () => { console.log(); };\n      "), messageIds: []string{"invalidVoidExprArrow"}},
	{source: "\ntype Foo = () => void;\n\nconst foo: Foo = function () {\n  function bar() {\n    return console.log();\n  }\n};\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\ntype Foo = () => void;\n\nconst foo: Foo = function () {\n  function bar() {\n    console.log();\n  }\n};\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\nconst foo = function () {\n  function bar() {\n    return console.log();\n  }\n};\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nconst foo = function () {\n  function bar() {\n    console.log();\n  }\n};\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\nreturn console.log('foo');\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\n{ console.log('foo'); return; }\n      "), messageIds: []string{"invalidVoidExprReturn"}},
	{source: "\nfunction test(): void;\nfunction test(arg: string): any;\nfunction test(arg?: string): any | void {\n  if (arg) {\n    return arg;\n  }\n  return console.log();\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nfunction test(): void;\nfunction test(arg: string): any;\nfunction test(arg?: string): any | void {\n  if (arg) {\n    return arg;\n  }\n  console.log();\n}\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
	{source: "\nfunction test(arg: string): any;\nfunction test(): void;\nfunction test(arg?: string): any | void {\n  if (arg) {\n    return arg;\n  }\n  return console.log();\n}\n      ", optionsJson: "{\"ignoreVoidReturningFunctions\":true}", jsx: false, fixedSource: noConfusingVoidExpressionStringPointer("\nfunction test(arg: string): any;\nfunction test(): void;\nfunction test(arg?: string): any | void {\n  if (arg) {\n    return arg;\n  }\n  console.log();\n}\n      "), messageIds: []string{"invalidVoidExprReturnLast"}},
}

// noConfusingVoidExpressionAmbientGlobals declares what upstream's corpus assumes its environment
// provides.
//
// Every case in that corpus is written with `console.log`, and the fixture harness pins
// `lib: ["ES2022"]` with `types: []`, so `console` resolves to nothing and its calls have no type.
// Measured: without this, 55 of upstream's 57 firing cases go silent while a case written with a
// locally-declared void function reports correctly. The rule was right and the environment was
// missing.
//
// Declared ambiently rather than by widening the harness's lib, which would change every other
// rule's fixtures in this package. `Math.random` is the corpus's stand-in for a non-void call, so
// it is here too and is deliberately typed as returning a number rather than void.
const noConfusingVoidExpressionAmbientGlobals = "declare const console: { log(...data: unknown[]): void; error(...data: unknown[]): void };\n" +
	"declare const Math: { random(): number };\n"

// noConfusingVoidExpressionJsxIntrinsics declares the one JSX element the corpus uses.
//
// The single JSX case is clean upstream because the arrow's contextual type comes from the
// element's `onClick` prop. The fixture harness has no React types, so a JSX intrinsic has no
// contextual type at all: measured, the arrow answers nil there while an identically-shaped object
// property answers `() => void`. Declaring the intrinsic locally supplies exactly what the real
// environment would and nothing more, which is the difference between reproducing upstream's
// verdict and weakening the rule until the case passes.
const noConfusingVoidExpressionJsxIntrinsics = "declare global {\n" +
	"  namespace JSX {\n" +
	"    interface IntrinsicElements { button: { onClick?: () => void } }\n" +
	"    interface Element {}\n" +
	"  }\n" +
	"}\n"

// runNoConfusingVoidExpressionCase drives one case with the ambient preamble prepended.
func runNoConfusingVoidExpressionCase(t *testing.T, testCase noConfusingVoidExpressionCase) rule_testing.Result {
	t.Helper()
	fileName := noConfusingVoidExpressionFile
	preamble := noConfusingVoidExpressionAmbientGlobals
	if testCase.jsx {
		fileName = noConfusingVoidExpressionJsxFile
		preamble += noConfusingVoidExpressionJsxIntrinsics
	}
	return rule_testing.RunTypedWithOptions(t, NoConfusingVoidExpression, fileName,
		preamble+testCase.source,
		decodeNoConfusingVoidExpressionOptionsForTest(t, testCase.optionsJson))
}

// TestNoConfusingVoidExpressionFires runs upstream's 57 invalid cases.
func TestNoConfusingVoidExpressionFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range noConfusingVoidExpressionFiringCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := runNoConfusingVoidExpressionCase(t, testCase)
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)

			// The repair, which no message-id assertion can see. Thirty seven of upstream's cases
			// carry an `output`, and a fixer anchored correctly can still write the wrong bytes.
			//
			// The comparison has to account for two harness facts. `RunTyped` writes the fixture
			// as TrimSpace(contents)+"\n", and this runner prepends a preamble, so the expected
			// output is transformed the same way rather than compared raw.
			if testCase.fixedSource == nil {
				// Upstream's `output: null`: reported and deliberately not repaired. Reproducing
				// the decline is the assertion.
				for i, diagnostic := range result.Diagnostics {
					if len(diagnostic.Fixes) != 0 {
						t.Errorf("finding %d proposes %d fixes; upstream declines this case",
							i, len(diagnostic.Fixes))
					}
				}
				return
			}
			preamble := noConfusingVoidExpressionAmbientGlobals
			if testCase.jsx {
				preamble += noConfusingVoidExpressionJsxIntrinsics
			}
			rule_testing.ExpectFixedSource(t, result,
				strings.TrimSpace(preamble+*testCase.fixedSource)+"\n")
		})
	}
}

// TestNoConfusingVoidExpressionStaysSilent runs upstream's 53 valid cases.
func TestNoConfusingVoidExpressionStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range noConfusingVoidExpressionCleanCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoConfusingVoidExpressionCase(t, testCase))
		})
	}
}

// TestNoConfusingVoidExpressionJudgesVoidNotUndefined is where cohere is deliberately quieter than
// ESLint.
//
// Upstream tests `TypeFlags.VoidLike`, which is void OR undefined, and its corpus writes no
// undefined-typed expression anywhere, so all 110 imported cases agree either way. Measured against
// the installed rule before this change: an undefined-returning call assigned or in an arrow
// shorthand reports exactly as a void-returning one does. That is a false positive. `void` says
// "there is no result, do not look", and `undefined` is a value somebody returned on purpose, which
// a test asserting `toBeUndefined()` reads by design. The rule's name and messages are about void.
//
// The silent rows include the two real sites in ahra that ESLint reports:
// `nexus/source/security/random/Random.test.ts:6`, `const result = arrayGetRandom([])` where `T`
// infers to `never` and the call types as `undefined`, and
// `nexus/source/coordination/TrackedPromise.test.ts:132`, an `await` of a `Promise<undefined>`.
//
// The reporting rows are the controls that keep this to undefined: the same positions with a real
// `void`, including `nexus/source/coordination/RandomSleep.test.ts:250`, an `await` of a
// `Promise<void>`, which is a true positive and stays one.
func TestNoConfusingVoidExpressionJudgesVoidNotUndefined(t *testing.T) {
	t.Parallel()

	const declarations = "declare function returnsUndefined(): undefined;\n" +
		"declare function returnsVoid(): void;\n" +
		"declare function arrayGetRandom<T>(array: readonly T[]): T | undefined;\n" +
		"declare const undefinedPromise: Promise<undefined>;\n" +
		"declare function randomSleep(minimum: number, maximum: number): Promise<void>;\n"

	for _, testCase := range []struct {
		name       string
		source     string
		messageIds []string
	}{
		{name: "an undefined call in statement position", source: "returnsUndefined();"},
		{name: "an undefined call assigned", source: "const value = returnsUndefined();"},
		{name: "an undefined call in an arrow shorthand", source: "const wrapped = () => returnsUndefined();"},
		{name: "an undefined call returned", source: "function f() { return returnsUndefined(); }"},
		{name: "Random.test.ts:6, a generic that infers to undefined", source: "const result = arrayGetRandom([]);"},
		{name: "TrackedPromise.test.ts:132, awaiting Promise<undefined>", source: "async function f() { const result = await undefinedPromise; }"},
		{name: "a void call assigned", source: "const value = returnsVoid();", messageIds: []string{"invalidVoidExpr"}},
		{name: "a void call in an arrow shorthand", source: "const wrapped = () => returnsVoid();", messageIds: []string{"invalidVoidExprArrow"}},
		{name: "a void call returned", source: "function f() { return returnsVoid(); }", messageIds: []string{"invalidVoidExprReturnLast"}},
		{name: "RandomSleep.test.ts:250, awaiting Promise<void>", source: "async function f() { const result = await randomSleep(100, 200); }", messageIds: []string{"invalidVoidExpr"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoConfusingVoidExpression,
				noConfusingVoidExpressionFile,
				noConfusingVoidExpressionAmbientGlobals+declarations+testCase.source, nil)
			if len(testCase.messageIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestNoConfusingVoidExpressionDeclinesToRepairOverAComment pins two declines this port adds.
//
// Upstream replaces the gap between the arrow token and the body, and the gap between the `return`
// keyword and its argument, unconditionally. A comment written in either stretch is inside the
// replaced range and is deleted by that repair. This port reports and withholds instead, which is
// this tree's standing rule for a fixer that would otherwise remove what it was not asked to
// remove.
//
// A deliberate narrowing rather than a different repair: upstream offers a fix here and this does
// not. Stated so the missing fix reads as a decision rather than an unfinished port.
func TestNoConfusingVoidExpressionDeclinesToRepairOverAComment(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name      string
		source    string
		messageId string
	}{
		{
			name:      "a comment between the arrow and its body",
			source:    "const wrapped = () => /* keep me */ console.log('foo');",
			messageId: "invalidVoidExprArrow",
		},
		{
			name:      "a comment between return and its argument",
			source:    "function f() { return /* keep me */ console.log('foo'); }",
			messageId: "invalidVoidExprReturnLast",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoConfusingVoidExpression,
				noConfusingVoidExpressionFile,
				noConfusingVoidExpressionAmbientGlobals+testCase.source, nil)
			rule_testing.ExpectFindings(t, result, testCase.messageId)
			for i, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("finding %d proposes %d fixes; the repair would delete the comment "+
						"sitting inside the range it replaces", i, len(diagnostic.Fixes))
				}
			}
		})
	}
}

// TestNoConfusingVoidExpressionVoidReturningIsVoidNotVoidLike pins the narrower predicate under
// ignoreVoidReturningFunctions.
//
// The option asks whether a function's return type INCLUDES VOID, and upstream tests
// `isIntrinsicVoidType` per union constituent rather than the VoidLike flag mask it uses for the
// expression itself. The two differ on `undefined`, which is VoidLike and is not void.
//
// A mutation widening this to VoidLike survived all 110 imported cases, because upstream's corpus
// writes no `undefined` annotation under this option. Measured against the installed rule: a
// function annotated `undefined` is NOT exempt, while `void` and `void | undefined` are.
func TestNoConfusingVoidExpressionVoidReturningIsVoidNotVoidLike(t *testing.T) {
	t.Parallel()

	options := `{"ignoreVoidReturningFunctions":true}`

	for _, testCase := range []struct {
		name       string
		source     string
		messageIds []string
	}{
		{
			name:       "an undefined annotation does not exempt",
			source:     "function h(): undefined { return console.log('foo'); }",
			messageIds: []string{"invalidVoidExprReturnLast"},
		},
		{
			name:       "a void annotation exempts",
			source:     "function h(): void { return console.log('foo'); }",
			messageIds: nil,
		},
		{
			name:       "a union including void exempts",
			source:     "function h(): void | undefined { return console.log('foo'); }",
			messageIds: nil,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoConfusingVoidExpression,
				noConfusingVoidExpressionFile,
				noConfusingVoidExpressionAmbientGlobals+testCase.source,
				decodeNoConfusingVoidExpressionOptionsForTest(t, options))
			if len(testCase.messageIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}
