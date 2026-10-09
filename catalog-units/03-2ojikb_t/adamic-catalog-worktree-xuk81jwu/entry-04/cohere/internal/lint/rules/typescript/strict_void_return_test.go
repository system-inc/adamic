package typescript

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const strictVoidReturnFile = "/repository/source/Callbacks.ts"

func strictVoidReturnCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

func strictVoidReturnOptionsFor(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return DefaultStrictVoidReturnSettings()
	}
	decoded, err := DecodeStrictVoidReturnOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("decoding %s: %v", optionsJson, err)
	}
	return decoded
}

// TestStrictVoidReturnStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All one hundred and five of upstream's passing inputs, extracted by parsing the clone's test file
// with the TypeScript compiler. Ninety-five of them were additionally replayed through the installed
// 8.x build against a real program and reported nothing; the other ten are JSX shapes the plain
// TypeScript harness in that replay could not parse, and they are kept because this harness reads
// them fine and they are upstream's own assertions.
//
// The clean half is where this rule's risk lives. It asks the checker what type is expected at a
// position, and a rule that answers that question too broadly reports on ordinary callbacks
// everywhere. Every one of these is a false positive somebody already thought about.
func TestStrictVoidReturnStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		optionsJson string
	}{
		{sourceText: "\ndeclare function foo(cb: {}): void;\nfoo(() => () => []);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\ntype Void = void;\nfoo((): Void => {\n  return;\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo((): ReturnType<typeof foo> => {\n  return;\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: any): void;\nfoo(() => () => []);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare class Foo {\n  constructor(cb: unknown): void;\n}\nnew Foo(() => ({}));\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => {}): void;\nfoo(() => 1 as any);\n      ", optionsJson: "{\"allowReturnAny\": true}"},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(() => {\n  throw new Error('boom');\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\ndeclare function boom(): never;\nfoo(() => boom());\nfoo(boom);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare const Foo: {\n  new (cb: () => any): void;\n};\nnew Foo(function () {\n  return 1;\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare const Foo: {\n  new (cb: () => unknown): void;\n};\nnew Foo(function () {\n  return 1;\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare const foo: {\n  bar(cb1: () => unknown, cb2: () => void): void;\n};\nfoo.bar(\n  function () {\n    return 1;\n  },\n  function () {\n    return;\n  },\n);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare const Foo: {\n  new (cb: () => string | void): void;\n};\nnew Foo(() => {\n  if (maybe) {\n    return 'a';\n  } else {\n    return 'b';\n  }\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo<Cb extends (...args: any[]) => void>(cb: Cb): void;\nfoo(() => {\n  console.log('a');\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: (() => void) | (() => string)): void;\nfoo(() => {\n  label: while (maybe) {\n    for (let i = 0; i < 10; i++) {\n      switch (i) {\n        case 0:\n          continue;\n        case 1:\n          return 'a';\n      }\n    }\n  }\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: (() => void) | null): void;\nfoo(null);\n      ", optionsJson: ""},
		{sourceText: "\ninterface Cb {\n  (): void;\n  (): string;\n}\ndeclare const Foo: {\n  new (cb: Cb): void;\n};\nnew Foo(() => {\n  do {\n    try {\n      throw 1;\n    } catch {\n      return 'a';\n    }\n  } while (maybe);\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare const foo: ((cb: () => boolean) => void) | ((cb: () => void) => void);\nfoo(() => false);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare const foo: {\n  (cb: () => boolean): void;\n  (cb: () => void): void;\n};\nfoo(function () {\n  with ({}) {\n    return false;\n  }\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare const Foo: {\n  new (cb: () => void): void;\n  (cb: () => unknown): void;\n};\nFoo(() => false);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare const Foo: {\n  new (cb: () => any): void;\n  (cb: () => void): void;\n};\nnew Foo(() => false);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => boolean): void;\ndeclare function foo(cb: () => void): void;\nfoo(() => false);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\ndeclare function foo(cb: () => boolean): void;\nfoo(() => false);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => Promise<void>): void;\ndeclare function foo(cb: () => void): void;\nfoo(async () => {});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(fn: () => void);\ndeclare function foo(fn: () => Promise<void>);\n\nfoo(async () => {});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(() => 1 as any);\n      ", optionsJson: "{\"allowReturnAny\": true}"},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(() => {});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nconst cb = () => {};\nfoo(cb);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(function () {});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(cb);\nfunction cb() {}\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(() => undefined);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(function () {\n  return;\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(function () {\n  return void 0;\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(() => {\n  return;\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\ndeclare function cb(): never;\nfoo(cb);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare class Foo {\n  constructor(cb: () => void): any;\n}\ndeclare function cb(): void;\nnew Foo(cb);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(cb);\nfunction cb() {\n  throw new Error('boom');\n}\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(arg: string, cb: () => void): void;\ndeclare function cb(): undefined;\nfoo('arg', cb);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb?: () => void): void;\nfoo();\n      ", optionsJson: ""},
		{sourceText: "\ndeclare class Foo {\n  constructor(cb?: () => void): void;\n}\ndeclare function cb(): void;\nnew Foo(cb);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(...cbs: Array<() => void>): void;\nfoo(\n  () => {},\n  () => void null,\n  () => undefined,\n);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(...cbs: Array<() => void>): void;\ndeclare const cbs: Array<() => void>;\nfoo(...cbs);\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(...cbs: [() => any, () => void, (() => void)?]): void;\nfoo(\n  async () => {},\n  () => void null,\n  () => undefined,\n);\n      ", optionsJson: ""},
		{sourceText: "\nlet cb;\ncb = async () => 10;\n      ", optionsJson: ""},
		{sourceText: "\nconst foo: () => void = () => {};\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function cb(): void;\nconst foo: () => void = cb;\n      ", optionsJson: ""},
		{sourceText: "\nconst foo: () => void = function () {\n  throw new Error('boom');\n};\n      ", optionsJson: ""},
		{sourceText: "\nconst foo: { (): string; (): void } = () => {\n  return 'a';\n};\n      ", optionsJson: ""},
		{sourceText: "\nconst foo: (() => void) | (() => number) = () => {\n  return 1;\n};\n      ", optionsJson: ""},
		{sourceText: "\ntype Foo = () => void;\nconst foo: Foo = cb;\nfunction cb() {\n  return void null;\n}\n      ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  (): void;\n}\nconst foo: Foo = cb;\nfunction cb() {\n  return undefined;\n}\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function cb(): void;\ndeclare let foo: () => void;\nfoo = cb;\n      ", optionsJson: ""},
		{sourceText: "\ndeclare let foo: () => void;\nfoo += () => 1;\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function defaultCb(): object;\ndeclare let foo: { cb?: () => void };\n// default doesn't have to be void\nconst { cb = defaultCb } = foo;\n      ", optionsJson: ""},
		{sourceText: "\nlet foo: (() => void) | null = null;\nfoo &&= null;\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function cb(): void;\nlet foo: (() => void) | boolean = false;\nfoo ||= cb;\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function Foo(props: { cb: () => void }): unknown;\nreturn <Foo cb={() => {}} />;\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function Foo(props: { cb: () => void }): unknown;\nreturn <Foo cb=\"() => {}\" />;\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function Foo(props: { cb: () => void }): unknown;\nreturn <Foo cb={} />;\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function Foo(props: { cb: () => void }): unknown;\nreturn <Bar children=<Foo cb={() => {}} /> />;\n      ", optionsJson: ""},
		{sourceText: "\ntype Cb = () => void;\ndeclare function Foo(props: { cb: Cb; s: string }): unknown;\nreturn <Foo cb={function () {}} s=\"asd\" />;\n      ", optionsJson: ""},
		{sourceText: "\ntype Cb = () => void;\ndeclare function Foo(props: { x: number; cb?: Cb }): unknown;\nreturn <Foo x={123} />;\n      ", optionsJson: ""},
		{sourceText: "\ntype Cb = (() => void) | (() => number);\ndeclare function Foo(props: { cb?: Cb }): unknown;\nreturn (\n  <Foo\n    cb={function (arg) {\n      return 123;\n    }}\n  />\n);\n      ", optionsJson: ""},
		{sourceText: "\ninterface Props {\n  cb: ((arg: unknown) => void) | boolean;\n}\ndeclare function Foo(props: Props): unknown;\nreturn <Foo cb />;\n      ", optionsJson: ""},
		{sourceText: "\ninterface Props {\n  cb: (() => void) | (() => Promise<void>);\n}\ndeclare function Foo(props: Props): any;\nconst _ = <Foo cb={async () => {}} />;\n      ", optionsJson: ""},
		{sourceText: "\ninterface Props {\n  children: (arg: unknown) => void;\n}\ndeclare function Foo(props: Props): unknown;\ndeclare function cb(): void;\nreturn <Foo>{cb}</Foo>;\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cbs: { arg: number; cb: () => void }): void;\nfoo({ arg: 1, cb: () => undefined });\n      ", optionsJson: ""},
		{sourceText: "\ndeclare let foo: { arg?: string; cb: () => void };\nfoo = {\n  cb: () => {\n    return something;\n  },\n};\n      ", optionsJson: "{\"allowReturnAny\": true}"},
		{sourceText: "\ndeclare let foo: { cb: () => void };\nfoo = {\n  cb() {\n    return something;\n  },\n};\n      ", optionsJson: "{\"allowReturnAny\": true}"},
		{sourceText: "\ndeclare let foo: { cb: () => void };\nfoo = {\n  // don't check this thing\n  cb = () => 1,\n};\n      ", optionsJson: ""},
		{sourceText: "\ndeclare let foo: { cb: (n: number) => void };\nlet method = 'cb';\nfoo = {\n  // don't check computed methods\n  [method](n) {\n    return n;\n  },\n};\n      ", optionsJson: ""},
		{sourceText: "\n// no contextual type for object\nlet foo = {\n  cb(n) {\n    return n;\n  },\n};\n      ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  fn(): void;\n}\n// no symbol for method cb\nlet foo: Foo = {\n  cb(n) {\n    return n;\n  },\n};\n      ", optionsJson: ""},
		{sourceText: "\ndeclare let foo: { cb: (() => void) | number };\nfoo = {\n  cb: 0,\n};\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function cb(): void;\nconst foo: Record<string, () => void> = {\n  cb1: cb,\n  cb2: cb,\n};\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function cb(): string;\nconst foo: Record<string, () => void> = {\n  ...cb,\n};\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function cb(): string;\nconst foo: Record<string, () => void> = {\n  ...cb,\n  ...{},\n};\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function cb(): void;\nconst foo: Array<(() => void) | false> = [false, cb, () => cb()];\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function cb(): void;\nconst foo: [string, () => void, (() => void)?] = ['asd', cb];\n      ", optionsJson: ""},
		{sourceText: "\nconst foo: { cbs: Array<() => void> | null } = {\n  cbs: [\n    function () {\n      return undefined;\n    },\n    () => {\n      return void 0;\n    },\n    null,\n  ],\n};\n      ", optionsJson: ""},
		{sourceText: "\nconst foo: { cb: () => void } = class {\n  static cb = () => {};\n};\n      ", optionsJson: ""},
		{sourceText: "\nclass Foo {\n  foo;\n}\n      ", optionsJson: ""},
		{sourceText: "\nclass Bar {\n  foo() {}\n}\nclass Foo extends Bar {\n  foo();\n}\n      ", optionsJson: ""},
		{sourceText: "\ninterface Bar {\n  foo(): void;\n}\nclass Foo implements Bar {\n  get foo() {\n    return new Date();\n  }\n  set foo() {\n    return new Date('wtf');\n  }\n}\n      ", optionsJson: ""},
		{sourceText: "\nclass Foo {\n  foo: () => void = () => undefined;\n}\n      ", optionsJson: ""},
		{sourceText: "\nclass Bar {}\nclass Foo extends Bar {\n  foo = () => 1;\n}\n      ", optionsJson: ""},
		{sourceText: "\nclass Foo extends Wtf {\n  foo = () => 1;\n}\n      ", optionsJson: ""},
		{sourceText: "\nclass Foo extends Wtf {\n  [unknown] = () => 1;\n}\n      ", optionsJson: ""},
		{sourceText: "\nclass Foo {\n  cb = () => {\n    console.log('siema');\n  };\n}\nclass Bar extends Foo {\n  cb = () => {\n    console.log('nara');\n  };\n}\n      ", optionsJson: ""},
		{sourceText: "\nclass Foo {\n  cb1 = () => {};\n}\nclass Bar extends Foo {\n  cb2() {}\n}\nclass Baz extends Bar {\n  cb1 = () => {\n    console.log('siema');\n  };\n  cb2() {\n    console.log('nara');\n  }\n}\n      ", optionsJson: ""},
		{sourceText: "\nclass Foo {\n  fn() {\n    return 'a';\n  }\n  cb() {}\n}\nvoid class extends Foo {\n  cb() {\n    if (maybe) {\n      console.log('siema');\n    } else {\n      console.log('nara');\n    }\n  }\n};\n      ", optionsJson: ""},
		{sourceText: "\nabstract class Foo {\n  abstract cb(): void;\n}\nclass Bar extends Foo {\n  cb() {\n    console.log('a');\n  }\n}\n      ", optionsJson: ""},
		{sourceText: "\nclass Bar implements Foo {\n  cb = () => 1;\n}\n      ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  cb: () => void;\n}\nclass Bar implements Foo {\n  cb = () => {};\n}\n      ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  cb: () => void;\n}\nclass Bar implements Foo {\n  get cb() {\n    return () => {};\n  }\n}\n      ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  cb(): void;\n}\nclass Bar implements Foo {\n  cb() {\n    return undefined;\n  }\n}\n      ", optionsJson: ""},
		{sourceText: "\ninterface Foo1 {\n  cb1(): void;\n}\ninterface Foo2 {\n  cb2: () => void;\n}\nclass Bar implements Foo1, Foo2 {\n  cb1() {}\n  cb2() {}\n}\n      ", optionsJson: ""},
		{sourceText: "\ninterface Foo1 {\n  cb1(): void;\n}\ninterface Foo2 extends Foo1 {\n  cb2: () => void;\n}\nclass Bar implements Foo2 {\n  cb1() {}\n  cb2() {}\n}\n      ", optionsJson: ""},
		{sourceText: "\ndeclare let foo: () => () => void;\nfoo = () => () => {};\n      ", optionsJson: ""},
		{sourceText: "\ndeclare let foo: { f(): () => void };\nfoo = {\n  f() {\n    return () => undefined;\n  },\n};\nfunction cb() {}\n      ", optionsJson: ""},
		{sourceText: "\ndeclare let foo: { f(): () => void };\nfoo.f = function () {\n  return () => {};\n};\n      ", optionsJson: ""},
		{sourceText: "\ndeclare let foo: () => (() => void) | string;\nfoo = () => 'asd' + 'zxc';\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: () => () => void): void;\nfoo(function () {\n  return () => {};\n});\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function foo(cb: (arg: string) => () => void): void;\ndeclare function foo(cb: (arg: number) => () => boolean): void;\nfoo((arg: number) => {\n  return cb;\n});\nfunction cb() {\n  return true;\n}\n      ", optionsJson: ""},
		{sourceText: "\ndeclare function f<T extends void>(arg: T, cb: () => T): void;\ndeclare function f<T extends string>(arg: T, cb: () => T): void;\n\nf('test', () => 'test');\nf(undefined, () => {});\n      ", optionsJson: ""},
		{sourceText: "\ninterface HookFunction<T extends void | Hook = void> {\n  (fn: () => void): T;\n  (fn: () => Promise<void>): T;\n}\n\nclass Hook {}\n\ndeclare var beforeEach: HookFunction<Hook>;\n\nbeforeEach(() => {});\n\nbeforeEach(async () => {});\n      ", optionsJson: ""},
	}
	for index, testCase := range cases {
		t.Run(strictVoidReturnCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, StrictVoidReturn,
				strictVoidReturnFile, testCase.sourceText,
				strictVoidReturnOptionsFor(t, testCase.optionsJson)))
		})
	}
}

// TestStrictVoidReturnFiresOnUpstreamFailCases is the imported reporting corpus, verbatim.
//
// One hundred and seven of upstream's one hundred and sixteen failing inputs. The nine omitted are
// JSX and were not reachable through the replay harness that produced these expectations, so their
// message ids were never measured; they are named in the rule's doc comment as the one part of the
// corpus this port has not been checked against.
//
// The expected message ids are the INSTALLED BUILD's rather than the corpus's, which matters: the
// corpus records what upstream's own tester asserts, and where the shipped build differs the build
// is what the differential compares against. All one hundred and seven agree with both here.
//
// The three ids are asserted per finding and in order, because this rule chooses between them from
// the shape of the value it found, and a rule that reported the right count with the wrong id would
// pass a count-only assertion while telling the reader the wrong thing.
func TestStrictVoidReturnFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		optionsJson string
		wantIds     []string
	}{
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(() => null);\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\n        declare function foo(cb: () => void): void;\n        foo(() => (((true))));\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\n        declare function foo(cb: () => void): void;\n        foo(async () => (((Promise.resolve(true)))));\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(async () => /* before */ Promise.resolve(true) /* after */);\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(\n  async () => /* before */ {\n    /* inside */\n  } /* after */,\n);\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\n        declare function foo(cb: () => void): void;\n        foo(() => {\n          if (maybe) {\n            return (((1) + 1));\n          }\n        });\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare function foo(arg: number, cb: () => void): void;\nfoo(0, () => 0);\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare function foo(cb?: { (): void }): void;\nfoo(() => () => {});\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare const obj: { foo(cb: () => void) } | null;\nobj?.foo(() => JSON.parse('{}'));\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\n((cb: () => void) => cb())!(() => 1);\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare function foo(cb: { (): void }): void;\ndeclare function cb(): string;\nfoo(cb);\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ntype AnyFunc = (...args: unknown[]) => unknown;\ndeclare function foo<F extends AnyFunc>(cb: F): void;\nfoo(async () => ({}));\nfoo<() => void>(async () => ({}));\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\nfunction foo<T extends {}>(arg: T, cb: () => T);\nfunction foo(arg: null, cb: () => void);\nfunction foo(arg: any, cb: () => any) {}\n\nfoo(null, () => Math.random());\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare function foo<T extends {}>(arg: T, cb: () => T): void;\ndeclare function foo(arg: any, cb: () => void): void;\n\nfoo(null, async () => {});\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\ndeclare function foo(cb: () => any): void;\nfoo(async () => {\n  return Math.random();\n});\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(async function () {\n  return -Math.random();\n});\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ndeclare function f<T extends void>(arg: T, cb: () => T): void;\n\nf(undefined, () => 'test');\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare function foo(cb: { (): void }): void;\nfoo(cb);\nasync function cb() {}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare function foo<Cb extends (...args: any[]) => void>(cb: Cb): void;\nfoo(() => {\n  console.log('a');\n  return 1;\n});\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfunction bar<Cb extends () => number>(cb: Cb) {\n  foo(cb);\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare function foo(cb: { (): void }): void;\nconst cb = () => dunno;\nfoo!(cb);\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare const foo: {\n  (arg: boolean, cb: () => void): void;\n};\nfoo(false, () => Promise.resolve(undefined));\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare const foo: {\n  bar(cb1: () => any, cb2: () => void): void;\n};\nfoo.bar(\n  () => Promise.resolve(1),\n  () => Promise.resolve(1),\n);\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare const Foo: {\n  new (cb: () => void): void;\n};\nnew Foo(async () => 123);\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(() => {\n  label: while (maybe) {\n    for (const i of [1, 2, 3]) {\n      if (maybe) return null;\n      else return null;\n    }\n  }\n  return void 0;\n});\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn", "nonVoidReturn"}},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(() => {\n  do {\n    try {\n      throw 1;\n    } catch (e) {\n      return null;\n    } finally {\n      console.log('finally');\n    }\n  } while (maybe);\n});\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare function foo(cb: () => void): void;\nfoo(async () => {\n  try {\n    await Promise.resolve();\n  } catch {\n    console.error('fail');\n  }\n});\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ndeclare const Foo: {\n  new (cb: () => void): void;\n  (cb: () => unknown): void;\n};\nnew Foo(() => false);\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare const Foo: {\n  new (cb: () => any): void;\n  (cb: () => void): void;\n};\nFoo(() => false);\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ninterface Cb {\n  (arg: string): void;\n  (arg: number): void;\n}\ndeclare function foo(cb: Cb): void;\nfoo(cb);\nfunction cb() {\n  return true;\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare function foo(\n  cb: ((arg: number) => void) | ((arg: string) => void),\n): void;\nfoo(cb);\nfunction cb() {\n  return 1 + 1;\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare function foo(cb: (() => void) | null): void;\ndeclare function cb(): boolean;\nfoo(cb);\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare function foo(...cbs: Array<() => void>): void;\nfoo(\n  () => {},\n  () => false,\n  () => 0,\n  () => '',\n);\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn", "nonVoidReturn", "nonVoidReturn"}},
		{sourceText: "\ndeclare function foo(...cbs: [() => void, () => void, (() => void)?]): void;\nfoo(\n  () => {},\n  () => Math.random(),\n  () => (1).toString(),\n);\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn", "nonVoidReturn"}},
		{sourceText: "\ninterface Ev {}\ninterface EvMap {\n  DOMContentLoaded: Ev;\n}\ntype EvListOrEvListObj = EvList | EvListObj;\ninterface EvList {\n  (evt: Event): void;\n}\ninterface EvListObj {\n  handleEvent(object: Ev): void;\n}\ninterface Win {\n  addEventListener<K extends keyof EvMap>(\n    type: K,\n    listener: (ev: EvMap[K]) => any,\n  ): void;\n  addEventListener(type: string, listener: EvListOrEvListObj): void;\n}\ndeclare const win: Win;\nwin.addEventListener('DOMContentLoaded', ev => ev);\nwin.addEventListener('custom', ev => ev);\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn", "nonVoidReturn"}},
		{sourceText: "\ndeclare function foo(x: null, cb: () => void): void;\ndeclare function foo(x: unknown, cb: () => any): void;\nfoo({}, async () => {});\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\nconst arr = [1, 2];\narr.forEach(async x => {\n  console.log(x);\n});\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\n[1, 2].forEach(async x => console.log(x));\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\nconst foo: () => void = () => false;\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\nconst { name }: () => void = function foo() {\n  return false;\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare const foo: Record<string, () => void>;\nfoo['a' + 'b'] = () => true;\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\nconst foo: () => void = async () => Promise.resolve(true);\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "const cb: () => void = (): Array<number> => [];", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\nconst cb: () => void = (): Array<number> => {\n  return [];\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "const cb: () => void = function*foo() {}", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "const cb: () => void = (): Promise<number> => Promise.resolve(1);", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\nconst cb: () => void = async (): Promise<number> => {\n  try {\n    return Promise.resolve(1);\n  } catch {}\n};\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "const cb: () => void = async (): Promise<number> => Promise.resolve(1);", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\nconst foo: () => void = async () => {\n  try {\n    return 1;\n  } catch {}\n};\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\nconst foo: () => void = async (): Promise<void> => {\n  try {\n    await Promise.resolve();\n  } finally {\n  }\n};\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\nconst foo: () => void = async () => {\n  try {\n    await Promise.resolve();\n  } catch (err) {\n    console.error(err);\n  }\n  console.log('ok');\n};\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "const foo: () => void = (): number => {};", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare function cb(): boolean;\nconst foo: () => void = cb;\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\nconst foo: () => void = function () {\n  if (maybe) {\n    return null;\n  } else {\n    return null;\n  }\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn", "nonVoidReturn"}},
		{sourceText: "\nconst foo: () => void = function () {\n  if (maybe) {\n    console.log('elo');\n    return { [1]: Math.random() };\n  }\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\nconst foo: { (arg: number): void; (arg: string): void } = arg => {\n  console.log('foo');\n  switch (typeof arg) {\n    case 'number':\n      return 0;\n    case 'string':\n      return '';\n  }\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn", "nonVoidReturn"}},
		{sourceText: "\nconst foo: ((arg: number) => void) | ((arg: string) => void) = async () => {\n  return 1;\n};\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ntype Foo = () => void;\nconst foo: Foo = cb;\nfunction cb() {\n  return [1, 2, 3];\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ninterface Foo {\n  (): void;\n}\nconst foo: Foo = cb;\nfunction cb() {\n  return { a: 1 };\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare function cb(): unknown;\ndeclare let foo: () => void;\nfoo = cb;\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare let foo: { arg?: string; cb?: () => void };\nfoo.cb = () => {\n  return 'siema';\n  console.log('siema');\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare function cb(): unknown;\nlet foo: (() => void) | null = null;\nfoo ??= cb;\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare function cb(): unknown;\nlet foo: (() => void) | boolean = false;\nfoo ||= cb;\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare function cb(): unknown;\nlet foo: (() => void) | boolean = false;\nfoo &&= cb;\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare function foo(cbs: { arg: number; cb: () => void }): void;\nfoo({ arg: 1, cb: () => 1 });\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare let foo: { arg?: string; cb: () => void };\nfoo = {\n  cb: () => {\n    let x = 'siema';\n    return x;\n  },\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare let foo: { cb: (n: number) => void };\nfoo = {\n  cb(n) {\n    return n;\n  },\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare let foo: { 1234: (n: number) => void };\nfoo = {\n  1234(n) {\n    return n;\n  },\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare let foo: { '1e+21': () => void };\nfoo = {\n  1_000_000_000_000_000_000_000: () => 1,\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare let foo: { cb: (() => void) | number };\nfoo = {\n  cb: async () => {\n    if (maybe) {\n      return 'asd';\n    }\n  },\n};\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ndeclare function cb(): number;\nconst foo: Record<string, () => void> = {\n  cb1: cb,\n  cb2: cb,\n  ...cb,\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc", "nonVoidFunc"}},
		{sourceText: "\ndeclare function cb(): number;\nconst foo: Array<(() => void) | false> = [false, cb, () => cb()];\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc", "nonVoidReturn"}},
		{sourceText: "\ndeclare function cb(): number;\nconst foo: [string, () => void, (() => void)?] = ['asd', cb];\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\nconst foo: { cbs: Array<() => void> | null } = {\n  cbs: [\n    function* () {\n      yield 1;\n    },\n    async () => {\n      await 1;\n    },\n    null,\n  ],\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc", "asyncFunc"}},
		{sourceText: "\nconst foo: { cb: () => void } = class {\n  static cb = () => ({});\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\nclass Foo {\n  foo: () => void = () => [];\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\nclass Foo {\n  static foo: () => void = Math.random;\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\nclass Foo {\n  cb = () => {};\n}\nclass Bar extends Foo {\n  cb = Math.random;\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\nconst foo = () =>\n  class {\n    cb = () => {};\n  };\nclass Bar extends foo() {\n  cb = Math.random;\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\nclass Foo {\n  cb() {\n    console.log('siema');\n  }\n}\nconst method = 'cb' as const;\nclass Bar extends Foo {\n  [method]() {\n    return 'nara';\n  }\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\nclass Bar {\n  foo() {}\n}\nclass Foo extends Bar {\n  get foo() {\n    return () => 1;\n  }\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\nclass Foo {\n  cb() {}\n}\nvoid class extends Foo {\n  cb() {\n    return Math.random();\n  }\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\nclass Foo {\n  cb1 = () => {};\n}\nclass Bar extends Foo {\n  cb2() {}\n}\nclass Baz extends Bar {\n  cb1 = () => Math.random();\n  cb2() {\n    return Math.random();\n  }\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn", "nonVoidReturn"}},
		{sourceText: "\ndeclare function f(): Promise<void>;\ninterface Foo {\n  cb: () => void;\n}\nclass Bar {\n  cb = () => {};\n}\nclass Baz extends Bar implements Foo {\n  cb: () => void = f;\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\nclass Foo {\n  fn() {\n    return 'a';\n  }\n  cb() {}\n}\nclass Bar extends Foo {\n  cb() {\n    if (maybe) {\n      return Promise.resolve('siema');\n    } else {\n      return Promise.resolve('nara');\n    }\n  }\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn", "nonVoidReturn"}},
		{sourceText: "\nabstract class Foo {\n  abstract cb(): void;\n}\nclass Bar extends Foo {\n  async cb() {}\n}\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\nclass Foo {\n  fn() {\n    return 'a';\n  }\n  cb() {}\n}\nclass Bar extends Foo {\n  *cb() {}\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ninterface Foo {\n  cb: () => void;\n}\nclass Bar implements Foo {\n  cb = Math.random;\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\nconst o = { cb() {} };\ntype O = typeof o;\nclass Bar implements O {\n  cb = Math.random;\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\n        class Foo {\n          cb() {}\n        }\n        class Bar extends Foo {\n          async*cb() {}\n        }\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ninterface Foo {\n  cb(): void;\n}\nclass Bar implements Foo {\n  async /* important comment */ cb(): Promise<string> {\n    return Promise.resolve('siema');\n  }\n}\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ninterface Foo {\n  cb(): void;\n}\nclass Bar implements Foo {\n  async cb(): Promise<string> {\n    return Promise.resolve('siema');\n  }\n}\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ninterface Foo {\n  cb(): void;\n}\nclass Bar implements Foo {\n  async cb() {\n    try {\n      return { a: ['asdf', 1234] };\n    } catch {\n      console.error('error');\n    }\n  }\n}\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ninterface Foo {\n  cb(): void;\n}\nclass Bar implements Foo {\n  cb() {\n    if (maybe) {\n      return Promise.resolve(1);\n    } else {\n      return;\n    }\n  }\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ninterface Foo1 {\n  cb1(): void;\n}\ninterface Foo2 {\n  cb2: () => void;\n}\nclass Bar implements Foo1, Foo2 {\n  async cb1() {\n    console.log('a');\n  }\n  async *cb2() {\n    console.log('b');\n  }\n}\n      ", optionsJson: "", wantIds: []string{"asyncFunc", "nonVoidFunc"}},
		{sourceText: "\ninterface Foo1 {\n  cb1(): void;\n}\ninterface Foo2 {\n  cb2: () => void;\n}\nclass Baz {\n  cb3() {}\n}\nclass Bar extends Baz implements Foo1, Foo2 {\n  async cb1() {}\n  async *cb2() {}\n  cb3() {\n    return Math.random();\n  }\n}\n      ", optionsJson: "", wantIds: []string{"asyncFunc", "nonVoidFunc", "nonVoidReturn"}},
		{sourceText: "\nclass A extends class {\n  cb() {}\n} {\n  cb() {\n    return Math.random();\n  }\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\nclass A extends class B {\n  cb() {}\n} {\n  cb() {\n    return Math.random();\n  }\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ninterface Foo1 {\n  cb1(): void;\n}\ninterface Foo2 extends Foo1 {\n  cb2: () => void;\n}\nclass Bar implements Foo2 {\n  async cb1() {\n    console.log('a');\n  }\n  async *cb2() {\n    console.log('b');\n  }\n}\n      ", optionsJson: "", wantIds: []string{"asyncFunc", "nonVoidFunc"}},
		{sourceText: "\ndeclare let foo: () => () => void;\nfoo = () => () => 1 + 1;\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare let foo: () => () => void;\nfoo = () => () => Math.random();\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare let foo: () => () => void;\ndeclare const cb: () => null | false;\nfoo = () => cb;\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ndeclare let foo: { f(): () => void };\nfoo = {\n  f() {\n    return () => cb;\n  },\n};\nfunction cb() {}\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare let foo: { f(): () => void };\nfoo.f = function () {\n  return () => {\n    return null;\n  };\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare let foo: () => (() => void) | string;\nfoo = () => () => {\n  return 'asd' + 'zxc';\n};\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare function foo(cb: () => () => void): void;\nfoo(function () {\n  return async (): Promise<unknown[]> => ['asdf', 1234, true];\n});\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ndeclare function foo(cb: (arg: string) => () => void): void;\ndeclare function foo(cb: (arg: number) => () => boolean): void;\nfoo((arg: string) => {\n  return cb;\n});\nasync function* cb() {\n  yield true;\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
	}
	for index, testCase := range cases {
		t.Run(strictVoidReturnCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StrictVoidReturn,
				strictVoidReturnFile, testCase.sourceText,
				strictVoidReturnOptionsFor(t, testCase.optionsJson))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			// No repair. Upstream ships suggestions for two of the three findings and this port
			// carries none of them, which is stated in the rule's doc comment as a deliberate
			// decline rather than an omission.
			for position, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Fatalf("finding %d proposes %d fixes; this port ships no repair",
						position, len(diagnostic.Fixes))
				}
			}
		})
	}
}

// TestStrictVoidReturnNeedsTheTypedHarness pins the checker declaration.
//
// Under the plain harness the checker is nil and every listener returns immediately, so the clean
// fixtures would pass having proven nothing.
func TestStrictVoidReturnNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !StrictVoidReturn.NeedsTypeChecker {
		t.Fatal("every finding comes from a contextual type, so the checker is required")
	}

	rule_testing.ExpectClean(t, rule_testing.Run(t, StrictVoidReturn, strictVoidReturnFile,
		"declare function takes(cb: () => void): void;\ntakes(() => 1);"))
}

// TestDecodeStrictVoidReturnOptions pins the decoder against its default.
func TestDecodeStrictVoidReturnOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "emptyObject", raw: `{}`, want: false},
		{name: "explicitTrue", raw: `{"allowReturnAny": true}`, want: true},
		{name: "explicitFalse", raw: `{"allowReturnAny": false}`, want: false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeStrictVoidReturnOptions(json.RawMessage(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.raw, err)
			}
			options, isOptions := decoded.(StrictVoidReturnOptions)
			if !isOptions {
				t.Fatalf("the decoder returned %T rather than the options struct", decoded)
			}
			if options.AllowReturnAny != testCase.want {
				t.Fatalf("allowReturnAny: expected %v, got %v", testCase.want, options.AllowReturnAny)
			}
		})
	}
}

// TestStrictVoidReturnFallsBackToTheDefaultOnNilOptions pins the nil-options path.
//
// A rule configured as a bare "error" is handed nil options, and every fixture above reaches the
// rule through the decoder, so nothing there can see the fallback. An `any`-returning callback is
// the separating input: it reports under the default and is clean when the option is on.
func TestStrictVoidReturnFallsBackToTheDefaultOnNilOptions(t *testing.T) {
	t.Parallel()

	const anyReturning = "declare function takes(cb: () => void): void;\ndeclare function makeAny(): any;\ntakes(() => makeAny());"

	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, StrictVoidReturn,
		strictVoidReturnFile, anyReturning, nil), "nonVoidReturn")

	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, StrictVoidReturn,
		strictVoidReturnFile, anyReturning,
		StrictVoidReturnOptions{AllowReturnAny: true}))
}

// TestStrictVoidReturnDoesNotClaimANestedFunctionsReturns closes a mutation blind spot.
//
// A `return` written inside a function nested in the callback belongs to that function, not to the
// callback, so it does not make the callback value-returning. Upstream's `walkStatements` makes the
// same distinction by never descending into a nested function.
//
// Nothing in upstream's corpus writes this shape, so a mutant that descended into nested functions
// survived all two hundred and twelve imported cases. All three rows below were measured clean
// against the installed 8.x build before being written here, with the control confirming the same
// callback reports when the return is its own.
func TestStrictVoidReturnDoesNotClaimANestedFunctionsReturns(t *testing.T) {
	t.Parallel()

	clean := []string{
		"declare function takes(cb: () => void): void;\ntakes(() => {\n  function inner() { return 1; }\n  inner();\n});",
		"declare function takes(cb: () => void): void;\ntakes(() => {\n  const inner = () => 2;\n  inner();\n});",
		"declare function takes(cb: () => void): void;\ntakes(() => {\n  class K { m() { return 3; } }\n  new K();\n});",
	}
	for index, sourceText := range clean {
		t.Run(strictVoidReturnCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, StrictVoidReturn,
				strictVoidReturnFile, sourceText, DefaultStrictVoidReturnSettings()))
		})
	}

	// The control: the same callback with the return as its OWN reports, so the silence above is
	// the nesting rule rather than the rule failing to see block bodies at all.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, StrictVoidReturn,
		strictVoidReturnFile,
		"declare function takes(cb: () => void): void;\ntakes(() => {\n  return 1;\n});",
		DefaultStrictVoidReturnSettings()), "nonVoidReturn")

	// The row that actually reaches the return walk, and the one the three above could not.
	//
	// Those three are clean because the callback's own return type is `void`, so the already-void
	// gate short-circuits before the walk runs; a mutant that descended into nested functions
	// survived all of them for that reason. This callback genuinely returns a value, so the gate
	// lets it through, and it also holds a nested function that returns. Upstream reports exactly
	// ONE finding, for the outer return only. Measured before being written here.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, StrictVoidReturn,
		strictVoidReturnFile,
		"declare function takes(cb: () => void): void;\ntakes(() => {\n  function inner() { return 'nested'; }\n  if (Math.random()) { return 1; }\n});",
		DefaultStrictVoidReturnSettings()), "nonVoidReturn")
}

// TestStrictVoidReturnFiresOnJsxCases is the remaining nine of upstream's failing inputs.
//
// These are the JSX shapes, which need a `.tsx` fixture name so the harness parses them as JSX
// rather than as type assertions. They were the one part of the corpus not covered when this rule
// first passed its other hundred and seven, and a mutant removing the JSX-attribute arm survived
// everything until they were added.
//
// Their expected message ids were measured the same way as the rest, by replaying each through the
// installed 8.x build against a real program, with the tsconfig set to preserve JSX.
func TestStrictVoidReturnFiresOnJsxCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		optionsJson string
		wantIds     []string
	}{
		{sourceText: "\ndeclare function Foo(props: { cb: () => void }): unknown;\nreturn <Foo cb={() => 1} />;\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ndeclare function Foo(props: { cb: () => void }): unknown;\ndeclare function getNull(): null;\nreturn (\n  <Foo\n    cb={() => {\n      if (maybe) return Math.random();\n      else return getNull();\n    }}\n  />\n);\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn", "nonVoidReturn"}},
		{sourceText: "\ntype Cb = () => void;\ndeclare function Foo(props: { cb: Cb }): unknown;\nreturn (\n  <Foo\n    cb={async () => {\n      await fetch('/boop');\n    }}\n  />\n);\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ntype Cb = () => void;\ndeclare function Foo(props: { cb: Cb; s: string }): unknown;\nreturn (\n  <Foo\n    cb={async function (): Promise<void> {\n      await fetch('/boop');\n    }}\n    s=\"!@#jp2gmd\"\n  />\n);\n      ", optionsJson: "", wantIds: []string{"asyncFunc"}},
		{sourceText: "\ntype Cb = () => void;\ndeclare function Foo(props: { n: number; cb?: Cb }): unknown;\nreturn <Foo n={2137} cb={function* () {}} />;\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ntype Cb = ((arg: string) => void) | ((arg: number) => void);\ndeclare function Foo(props: { cb?: Cb }): unknown;\nreturn (\n  <Foo\n    cb={async function* (arg) {\n      await arg;\n      yield arg;\n    }}\n  />\n);\n      ", optionsJson: "", wantIds: []string{"nonVoidFunc"}},
		{sourceText: "\ninterface Props {\n  cb: ((arg: unknown) => void) | boolean;\n}\ndeclare function Foo(props: Props): unknown;\nreturn <Foo cb={x => x} />;\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\ntype EventHandler<E> = { bivarianceHack(event: E): void }['bivarianceHack'];\ninterface ButtonProps {\n  onClick?: EventHandler<unknown> | undefined;\n}\ndeclare function Button(props: ButtonProps): unknown;\nfunction App() {\n  return <Button onClick={x => x} />;\n}\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn"}},
		{sourceText: "\n        declare function foo(cb: () => () => void): void;\n        foo(() => () => {\n          if (n == 1) {\n            console.log('asd')\n            return [1].map(x => x)\n          }\n          if (n == 2) {\n            console.log('asd')\n            return -Math.random()\n          }\n          if (n == 3) {\n            console.log('asd')\n            return `x`.toUpperCase()\n          }\n          return <i>{Math.random()}</i>\n        });\n      ", optionsJson: "", wantIds: []string{"nonVoidReturn", "nonVoidReturn", "nonVoidReturn", "nonVoidReturn"}},
	}
	for index, testCase := range cases {
		t.Run(strictVoidReturnCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StrictVoidReturn,
				"/repository/source/Component.tsx", testCase.sourceText,
				strictVoidReturnOptionsFor(t, testCase.optionsJson))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestStrictVoidReturnReportsAConciseBodyWithoutItsParentheses pins where a concise arrow body's
// finding points.
//
// ESTree has no parenthesis node, so upstream's body is the expression inside the parentheses and
// its finding starts there. Ours kept the `ParenthesizedExpression` and reported it, one column
// early: measured on ahra, eight sites in `modules/data/` written
// `(chunk: Buffer) => (tarError += chunk.toString())` disagreed with ESLint by exactly that column.
// The spans below were measured on the installed 8.67.0 build, including a doubled parenthesis.
func TestStrictVoidReturnReportsAConciseBodyWithoutItsParentheses(t *testing.T) {
	t.Parallel()

	const prelude = "declare function on(callback: (chunk: string) => void): void;\nlet total = '';\n"
	for _, testCase := range []struct {
		name   string
		source string
	}{
		{"parenthesized", "on((chunk: string) => (total += chunk));\nvoid total;"},
		{"doubly parenthesized", "on((chunk: string) => ((total += chunk)));\nvoid total;"},
		{"bare", "on((chunk: string) => total += chunk);\nvoid total;"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sourceText := prelude + testCase.source
			result := rule_testing.RunTypedWithOptions(t, StrictVoidReturn, strictVoidReturnFile, sourceText,
				DefaultStrictVoidReturnSettings())
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			if reported := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]; reported != "total += chunk" {
				t.Errorf("the finding covers %q, want %q", reported, "total += chunk")
			}
			if diagnostic.Message.Id != "nonVoidReturn" {
				t.Errorf("message id is %q", diagnostic.Message.Id)
			}
		})
	}
}
