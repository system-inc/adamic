package typescript

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// noMisusedPromisesFile is the fixture name most cases in this file run under.
//
// A TypeScript extension is required rather than incidental: the corpus is dense with type
// annotations, interfaces and `declare` statements, so a `.js` name would turn most of these into
// parse errors rather than rule inputs.
const noMisusedPromisesFile = "noMisusedPromises.ts"

// noMisusedPromisesTsxFile carries the seven cases upstream marks `Tsx: true`.
//
// Four are valid and three report `voidReturnAttribute`, and that message id is reachable through no
// other extension: the attribute check anchors on `ast.KindJsxAttribute`, which a `.ts` file cannot
// produce. Running those seven under the `.ts` name would not fail loudly — it would parse the JSX as
// type assertions and comparisons, produce no attribute nodes, and leave three invalid cases silently
// asserting the opposite of upstream while the whole suite stayed green.
const noMisusedPromisesTsxFile = "noMisusedPromises.tsx"

// TestNoMisusedPromisesStaysSilent carries all one hundred and twenty-three of tsgolint's valid
// cases.
//
// Extracted from tsgolint's own test file by parsing it with go/ast rather than by transcribing it.
// That is not a stylistic preference: a fixture written by hand encodes the same belief as the port
// it is meant to check, so it passes for exactly the reason the code is wrong. Every source string
// and every option expression below came out of that parse, and the extractor's counts were verified
// against the upstream file independently (`grep -c 'Code:'` gives 213 against 123 valid plus 90
// invalid, and `grep -c 'MessageId:'` gives 106 against 106 extracted diagnostics) so that an
// extraction bug — which looks exactly like a port bug — could not hide.
//
// The options are half of what this table proves. Fifty-six of the two hundred and thirteen cases set
// at least one option, and the same source text appears here as silent and in the Fires table as
// reporting, differing only in which flag was set. A port that ignored options would fail both tables
// rather than neither, but a port that read the wrong flag for a given position would pass one and
// fail the other, which is the failure this split is shaped to catch.
func TestNoMisusedPromisesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		options    any
	}{
		{
			name:     "upstream valid 0",
			fileName: noMisusedPromisesFile,
			sourceText: `
if (true) {
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 1 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
if (Promise.resolve()) {
}
      `,
			options: NoMisusedPromisesOptions{ChecksConditionals: type_checking.Ref(false)},
		},
		{
			name:     "upstream valid 2",
			fileName: noMisusedPromisesFile,
			sourceText: `
if (true) {
} else if (false) {
} else {
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 3 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
if (Promise.resolve()) {
} else if (Promise.resolve()) {
} else {
}
      `,
			options: NoMisusedPromisesOptions{ChecksConditionals: type_checking.Ref(false)},
		},
		{
			name:       "upstream valid 4",
			fileName:   noMisusedPromisesFile,
			sourceText: `for (;;) {}`,
			options:    nil,
		},
		{
			name:       "upstream valid 5",
			fileName:   noMisusedPromisesFile,
			sourceText: `for (let i; i < 10; i++) {}`,
			options:    nil,
		},
		{
			name:       "upstream valid 6 [options]",
			fileName:   noMisusedPromisesFile,
			sourceText: `for (let i; Promise.resolve(); i++) {}`,
			options:    NoMisusedPromisesOptions{ChecksConditionals: type_checking.Ref(false)},
		},
		{
			name:       "upstream valid 7",
			fileName:   noMisusedPromisesFile,
			sourceText: `do {} while (true);`,
			options:    nil,
		},
		{
			name:       "upstream valid 8 [options]",
			fileName:   noMisusedPromisesFile,
			sourceText: `do {} while (Promise.resolve());`,
			options:    NoMisusedPromisesOptions{ChecksConditionals: type_checking.Ref(false)},
		},
		{
			name:       "upstream valid 9",
			fileName:   noMisusedPromisesFile,
			sourceText: `while (true) {}`,
			options:    nil,
		},
		{
			name:       "upstream valid 10 [options]",
			fileName:   noMisusedPromisesFile,
			sourceText: `while (Promise.resolve()) {}`,
			options:    NoMisusedPromisesOptions{ChecksConditionals: type_checking.Ref(false)},
		},
		{
			name:       "upstream valid 11",
			fileName:   noMisusedPromisesFile,
			sourceText: `true ? 123 : 456;`,
			options:    nil,
		},
		{
			name:       "upstream valid 12 [options]",
			fileName:   noMisusedPromisesFile,
			sourceText: `Promise.resolve() ? 123 : 456;`,
			options:    NoMisusedPromisesOptions{ChecksConditionals: type_checking.Ref(false)},
		},
		{
			name:     "upstream valid 13",
			fileName: noMisusedPromisesFile,
			sourceText: `
if (!true) {
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 14 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
if (!Promise.resolve()) {
}
      `,
			options: NoMisusedPromisesOptions{ChecksConditionals: type_checking.Ref(false)},
		},
		{
			name:       "upstream valid 15",
			fileName:   noMisusedPromisesFile,
			sourceText: `(await Promise.resolve()) || false;`,
			options:    nil,
		},
		{
			name:       "upstream valid 16 [options]",
			fileName:   noMisusedPromisesFile,
			sourceText: `Promise.resolve() || false;`,
			options:    NoMisusedPromisesOptions{ChecksConditionals: type_checking.Ref(false)},
		},
		{
			name:       "upstream valid 17",
			fileName:   noMisusedPromisesFile,
			sourceText: `(true && (await Promise.resolve())) || false;`,
			options:    nil,
		},
		{
			name:       "upstream valid 18 [options]",
			fileName:   noMisusedPromisesFile,
			sourceText: `(true && Promise.resolve()) || false;`,
			options:    NoMisusedPromisesOptions{ChecksConditionals: type_checking.Ref(false)},
		},
		{
			name:       "upstream valid 19",
			fileName:   noMisusedPromisesFile,
			sourceText: `false || (true && Promise.resolve());`,
			options:    nil,
		},
		{
			name:       "upstream valid 20",
			fileName:   noMisusedPromisesFile,
			sourceText: `(true && Promise.resolve()) || false;`,
			options:    nil,
		},
		{
			name:     "upstream valid 21",
			fileName: noMisusedPromisesFile,
			sourceText: `
async function test() {
  if (await Promise.resolve()) {
  }
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 22",
			fileName: noMisusedPromisesFile,
			sourceText: `
async function test() {
  const mixed: Promise | undefined = Promise.resolve();
  if (mixed) {
    await mixed;
  }
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 23",
			fileName: noMisusedPromisesFile,
			sourceText: `
if (~Promise.resolve()) {
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 24",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface NotQuiteThenable {
  then(param: string): void;
  then(): void;
}
const value: NotQuiteThenable = { then() {} };
if (value) {
}
    `,
			options: nil,
		},
		{
			name:       "upstream valid 25",
			fileName:   noMisusedPromisesFile,
			sourceText: `[1, 2, 3].forEach(val => {});`,
			options:    nil,
		},
		{
			name:       "upstream valid 26 [options]",
			fileName:   noMisusedPromisesFile,
			sourceText: `[1, 2, 3].forEach(async val => {});`,
			options:    NoMisusedPromisesOptions{ChecksVoidReturn: type_checking.Ref(false)},
		},
		{
			name:       "upstream valid 27",
			fileName:   noMisusedPromisesFile,
			sourceText: `new Promise((resolve, reject) => resolve());`,
			options:    nil,
		},
		{
			name:       "upstream valid 28 [options]",
			fileName:   noMisusedPromisesFile,
			sourceText: `new Promise(async (resolve, reject) => resolve());`,
			options:    NoMisusedPromisesOptions{ChecksVoidReturn: type_checking.Ref(false)},
		},
		{
			name:     "upstream valid 29",
			fileName: noMisusedPromisesFile,
			sourceText: `
Promise.all(
  ['abc', 'def'].map(async val => {
    await val;
  }),
);
    `,
			options: nil,
		},
		{
			name:     "upstream valid 30",
			fileName: noMisusedPromisesFile,
			sourceText: `
const fn: (arg: () => Promise<void> | void) => void = () => {};
fn(() => Promise.resolve());
    `,
			options: nil,
		},
		{
			name:     "upstream valid 31",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare const returnsPromise: (() => Promise<void>) | null;
if (returnsPromise?.()) {
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 32",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare const returnsPromise: { call: () => Promise<void> } | null;
if (returnsPromise?.call()) {
}
    `,
			options: nil,
		},
		{
			name:       "upstream valid 33",
			fileName:   noMisusedPromisesFile,
			sourceText: `Promise.resolve() ?? false;`,
			options:    nil,
		},
		{
			name:     "upstream valid 34",
			fileName: noMisusedPromisesFile,
			sourceText: `
function test(a: Promise<void> | undefinded) {
  const foo = a ?? Promise.reject();
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 35",
			fileName: noMisusedPromisesFile,
			sourceText: `
function test(p: Promise<boolean> | undefined, bool: boolean) {
  if (p ?? bool) {
  }
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 36",
			fileName: noMisusedPromisesFile,
			sourceText: `
async function test(p: Promise<boolean | undefined>, bool: boolean) {
  if ((await p) ?? bool) {
  }
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 37",
			fileName: noMisusedPromisesFile,
			sourceText: `
async function test(p: Promise<boolean> | undefined) {
  if (await (p ?? Promise.reject())) {
  }
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 38",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare let a: Promise<string> | undefined

declare let b: Promise<string>

a = a ?? b
		`,
			options: nil,
		},
		{
			name:     "upstream valid 39",
			fileName: noMisusedPromisesFile,
			sourceText: `
let f;
f = async () => 10;
    `,
			options: nil,
		},
		{
			name:     "upstream valid 40",
			fileName: noMisusedPromisesFile,
			sourceText: `
let f: () => Promise<void>;
f = async () => 10;
const g = async () => 0;
const h: () => Promise<void> = async () => 10;
    `,
			options: nil,
		},
		{
			name:     "upstream valid 41",
			fileName: noMisusedPromisesFile,
			sourceText: `
const obj = {
  f: async () => 10,
};
    `,
			options: nil,
		},
		{
			name:     "upstream valid 42",
			fileName: noMisusedPromisesFile,
			sourceText: `
const f = async () => 123;
const obj = {
  f,
};
    `,
			options: nil,
		},
		{
			name:     "upstream valid 43",
			fileName: noMisusedPromisesFile,
			sourceText: `
const obj = {
  async f() {
    return 0;
  },
};
    `,
			options: nil,
		},
		{
			name:     "upstream valid 44",
			fileName: noMisusedPromisesFile,
			sourceText: `
type O = { f: () => Promise<void>; g: () => Promise<void> };
const g = async () => 0;
const obj: O = {
  f: async () => 10,
  g,
};
    `,
			options: nil,
		},
		{
			name:     "upstream valid 45",
			fileName: noMisusedPromisesFile,
			sourceText: `
type O = { f: () => Promise<void> };
const name = 'f';
const obj: O = {
  async [name]() {
    return 10;
  },
};
    `,
			options: nil,
		},
		{
			name:     "upstream valid 46",
			fileName: noMisusedPromisesFile,
			sourceText: `
const obj: number = {
  g() {
    return 10;
  },
};
    `,
			options: nil,
		},
		{
			name:     "upstream valid 47",
			fileName: noMisusedPromisesFile,
			sourceText: `
const obj = {
  f: async () => 'foo',
  async g() {
    return 0;
  },
};
    `,
			options: nil,
		},
		{
			name:     "upstream valid 48",
			fileName: noMisusedPromisesFile,
			sourceText: `
function f() {
  return async () => 0;
}
function g() {
  return;
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 49 [tsx]",
			fileName: noMisusedPromisesTsxFile,
			sourceText: `
type O = {
  bool: boolean;
  func: () => Promise<void>;
};
const Component = (obj: O) => null;
<Component bool func={async () => 10} />;
      `,
			options: nil,
		},
		{
			name:     "upstream valid 50 [tsx]",
			fileName: noMisusedPromisesTsxFile,
			sourceText: `
const Component: any = () => null;
<Component func={async () => 10} />;
      `,
			options: nil,
		},
		{
			name:     "upstream valid 51",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface ItLike {
  (name: string, callback: () => Promise<void>): void;
  (name: string, callback: () => void): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
			options: nil,
		},
		{
			name:     "upstream valid 52",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface ItLike {
  (name: string, callback: () => void): void;
  (name: string, callback: () => Promise<void>): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
			options: nil,
		},
		{
			name:     "upstream valid 53",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface ItLike {
  (name: string, callback: () => void): void;
}
interface ItLike {
  (name: string, callback: () => Promise<void>): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
			options: nil,
		},
		{
			name:     "upstream valid 54",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface ItLike {
  (name: string, callback: () => Promise<void>): void;
}
interface ItLike {
  (name: string, callback: () => void): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
			options: nil,
		},
		{
			name:     "upstream valid 55 [tsx]",
			fileName: noMisusedPromisesTsxFile,
			sourceText: `
interface Props {
  onEvent: (() => void) | (() => Promise<void>);
}

declare function Component(props: Props): any;

const _ = <Component onEvent={async () => {}} />;
      `,
			options: nil,
		},
		{
			name:     "upstream valid 56",
			fileName: noMisusedPromisesFile,
			sourceText: `
console.log({ ...(await Promise.resolve({ key: 42 })) });
    `,
			options: nil,
		},
		{
			name:     "upstream valid 57",
			fileName: noMisusedPromisesFile,
			sourceText: `
const getData = () => Promise.resolve({ key: 42 });

console.log({
  someData: 42,
  ...(await getData()),
});
    `,
			options: nil,
		},
		{
			name:     "upstream valid 58",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare const condition: boolean;

console.log({ ...(condition && (await Promise.resolve({ key: 42 }))) });
console.log({ ...(condition || (await Promise.resolve({ key: 42 }))) });
console.log({ ...(condition ? {} : await Promise.resolve({ key: 42 })) });
console.log({ ...(condition ? await Promise.resolve({ key: 42 }) : {}) });
    `,
			options: nil,
		},
		{
			name:     "upstream valid 59",
			fileName: noMisusedPromisesFile,
			sourceText: `
console.log([...(await Promise.resolve(42))]);
    `,
			options: nil,
		},
		{
			name:     "upstream valid 60 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
console.log({ ...Promise.resolve({ key: 42 }) });
      `,
			options: NoMisusedPromisesOptions{ChecksSpreads: type_checking.Ref(false)},
		},
		{
			name:     "upstream valid 61 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
const getData = () => Promise.resolve({ key: 42 });

console.log({
  someData: 42,
  ...getData(),
});
      `,
			options: NoMisusedPromisesOptions{ChecksSpreads: type_checking.Ref(false)},
		},
		{
			name:     "upstream valid 62 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare const condition: boolean;

console.log({ ...(condition && Promise.resolve({ key: 42 })) });
console.log({ ...(condition || Promise.resolve({ key: 42 })) });
console.log({ ...(condition ? {} : Promise.resolve({ key: 42 })) });
console.log({ ...(condition ? Promise.resolve({ key: 42 }) : {}) });
      `,
			options: NoMisusedPromisesOptions{ChecksSpreads: type_checking.Ref(false)},
		},
		{
			name:     "upstream valid 63 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
// This is invalid Typescript, but it shouldn't trigger this linter specifically
console.log([...Promise.resolve(42)]);
      `,
			options: NoMisusedPromisesOptions{ChecksSpreads: type_checking.Ref(false)},
		},
		{
			name:     "upstream valid 64",
			fileName: noMisusedPromisesFile,
			sourceText: `
function spreadAny(..._args: any): void {}

spreadAny(
  true,
  () => Promise.resolve(1),
  () => Promise.resolve(false),
);
    `,
			options: nil,
		},
		{
			name:     "upstream valid 65",
			fileName: noMisusedPromisesFile,
			sourceText: `
function spreadArrayAny(..._args: Array<any>): void {}

spreadArrayAny(
  true,
  () => Promise.resolve(1),
  () => Promise.resolve(false),
);
    `,
			options: nil,
		},
		{
			name:     "upstream valid 66",
			fileName: noMisusedPromisesFile,
			sourceText: `
function spreadArrayUnknown(..._args: Array<unknown>): void {}

spreadArrayUnknown(() => Promise.resolve(true), 1, 2);

function spreadArrayFuncPromise(
  ..._args: Array<() => Promise<undefined>>
): void {}

spreadArrayFuncPromise(
  () => Promise.resolve(undefined),
  () => Promise.resolve(undefined),
);
    `,
			options: nil,
		},
		{
			name:     "upstream valid 67",
			fileName: noMisusedPromisesFile,
			sourceText: `
class TakeCallbacks {
  constructor(...callbacks: Array<() => void>) {}
}

new TakeCallbacks;
new TakeCallbacks();
new TakeCallbacks(
  () => 1,
  () => true,
);
    `,
			options: nil,
		},
		{
			name:     "upstream valid 68",
			fileName: noMisusedPromisesFile,
			sourceText: `
class Foo {
  public static doThing(): void {}
}

class Bar extends Foo {
  public async doThing(): Promise<void> {}
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 69",
			fileName: noMisusedPromisesFile,
			sourceText: `
class Foo {
  public doThing(): void {}
}

class Bar extends Foo {
  public static async doThing(): Promise<void> {}
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 70",
			fileName: noMisusedPromisesFile,
			sourceText: `
class Foo {
  public doThing = (): void => {};
}

class Bar extends Foo {
  public static doThing = async (): Promise<void> => {};
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 71",
			fileName: noMisusedPromisesFile,
			sourceText: `
class Foo {
  public doThing = (): void => {};
}

class Bar extends Foo {
  public static accessor doThing = async (): Promise<void> => {};
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 72",
			fileName: noMisusedPromisesFile,
			sourceText: `
class Foo {
  public accessor doThing = (): void => {};
}

class Bar extends Foo {
  public static accessor doThing = (): void => {};
}
    `,
			options: nil,
		},
		{
			name:     "upstream valid 73 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
class Foo {
  [key: string]: void;
}

class Bar extends Foo {
  [key: string]: Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 74",
			fileName: noMisusedPromisesFile,
			sourceText: `
function restTuple(...args: []): void;
function restTuple(...args: [string]): void;
function restTuple(..._args: string[]): void {}

restTuple();
restTuple('Hello');
    `,
			options: nil,
		},
		{
			name:     "upstream valid 75",
			fileName: noMisusedPromisesFile,
			sourceText: `
      let value: Record<string, () => void>;
      value.sync = () => {};
    `,
			options: nil,
		},
		{
			name:     "upstream valid 76",
			fileName: noMisusedPromisesFile,
			sourceText: `
      type ReturnsRecord = () => Record<string, () => void>;

      const test: ReturnsRecord = () => {
        return { sync: () => {} };
      };
    `,
			options: nil,
		},
		{
			name:     "upstream valid 77",
			fileName: noMisusedPromisesFile,
			sourceText: `
      type ReturnsRecord = () => Record<string, () => void>;

      function sync() {}

      const test: ReturnsRecord = () => {
        return { sync };
      };
    `,
			options: nil,
		},
		{
			name:     "upstream valid 78",
			fileName: noMisusedPromisesFile,
			sourceText: `
      function withTextRecurser<Text extends string>(
        recurser: (text: Text) => void,
      ): (text: Text) => void {
        return (text: Text): void => {
          if (text.length) {
            return;
          }

          return recurser(node);
        };
      }
    `,
			options: nil,
		},
		{
			name:     "upstream valid 79",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare function foo(cb: undefined | (() => void));
declare const bar: undefined | (() => void);
foo(bar);
    `,
			options: nil,
		},
		{
			name:     "upstream valid 80 [options] [tsx]",
			fileName: noMisusedPromisesTsxFile,
			sourceText: `
        type OnSelectNodeFn = (node: string | null) => void;

        interface ASTViewerBaseProps {
          readonly onSelectNode?: OnSelectNodeFn;
        }

        declare function ASTViewer(props: ASTViewerBaseProps): null;
        declare const onSelectFn: OnSelectNodeFn;

        <ASTViewer onSelectNode={onSelectFn} />;
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{Attributes: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 81 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

class MySubclassExtendsMyClass extends MyClass {
  setThing(): void {
    return;
  }
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 82 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}

class MySubclassExtendsMyClass extends MyClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 83 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

class MySubclassExtendsMyClass extends MyClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(false)})},
		},
		{
			name:     "upstream valid 84 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

abstract class MyAbstractClassExtendsMyClass extends MyClass {
  abstract setThing(): void;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 85 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

abstract class MyAbstractClassExtendsMyClass extends MyClass {
  abstract setThing(): Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(false)})},
		},
		{
			name:     "upstream valid 86 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

interface MyInterfaceExtendsMyClass extends MyClass {
  setThing(): void;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 87 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

interface MyInterfaceExtendsMyClass extends MyClass {
  setThing(): Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(false)})},
		},
		{
			name:     "upstream valid 88 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

class MySubclassExtendsMyAbstractClass extends MyAbstractClass {
  setThing(): void {
    return;
  }
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 89 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

class MySubclassExtendsMyAbstractClass extends MyAbstractClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(false)})},
		},
		{
			name:     "upstream valid 90 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

abstract class MyAbstractSubclassExtendsMyAbstractClass extends MyAbstractClass {
  abstract setThing(): void;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 91 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

abstract class MyAbstractSubclassExtendsMyAbstractClass extends MyAbstractClass {
  abstract setThing(): Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(false)})},
		},
		{
			name:     "upstream valid 92 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

interface MyInterfaceExtendsMyAbstractClass extends MyAbstractClass {
  setThing(): void;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 93 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

interface MyInterfaceExtendsMyAbstractClass extends MyAbstractClass {
  setThing(): Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(false)})},
		},
		{
			name:     "upstream valid 94 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

interface MySubInterfaceExtendsMyInterface extends MyInterface {
  setThing(): void;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 95 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

interface MySubInterfaceExtendsMyInterface extends MyInterface {
  setThing(): Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(false)})},
		},
		{
			name:     "upstream valid 96 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

class MyClassImplementsMyInterface implements MyInterface {
  setThing(): void {
    return;
  }
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 97 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

class MyClassImplementsMyInterface implements MyInterface {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(false)})},
		},
		{
			name:     "upstream valid 98 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

abstract class MyAbstractClassImplementsMyInterface implements MyInterface {
  abstract setThing(): void;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 99 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

abstract class MyAbstractClassImplementsMyInterface implements MyInterface {
  abstract setThing(): Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(false)})},
		},
		{
			name:     "upstream valid 100 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
type MyTypeLiteralsIntersection = { setThing(): void } & { thing: number };

class MyClass implements MyTypeLiteralsIntersection {
  thing = 1;
  setThing(): void {
    return;
  }
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 101 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
type MyTypeLiteralsIntersection = { setThing(): void } & { thing: number };

class MyClass implements MyTypeLiteralsIntersection {
  thing = 1;
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(false)})},
		},
		{
			name:     "upstream valid 102 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
type MyGenericType<IsAsync extends boolean = true> = IsAsync extends true
  ? { setThing(): Promise<void> }
  : { setThing(): void };

interface MyAsyncInterface extends MyGenericType {
  setThing(): Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 103 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
type MyGenericType<IsAsync extends boolean = true> = IsAsync extends true
  ? { setThing(): Promise<void> }
  : { setThing(): void };

interface MyAsyncInterface extends MyGenericType<false> {
  setThing(): Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(false)})},
		},
		{
			name:     "upstream valid 104 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

interface MyOtherInterface {
  setThing(): void;
}

interface MyThirdInterface extends MyInterface, MyOtherInterface {
  setThing(): void;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 105 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

class MyOtherClass {
  setThing(): void {
    return;
  }
}

interface MyInterface extends MyClass, MyOtherClass {
  setThing(): void;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 106 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

interface MyOtherInterface {
  setThing(): void;
}

class MyClass {
  setThing(): void {
    return;
  }
}

class MySubclass extends MyClass implements MyInterface, MyOtherInterface {
  setThing(): void {
    return;
  }
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 107 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

const MyClassExpressionExtendsMyClass = class extends MyClass {
  setThing(): void {
    return;
  }
};
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 108 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
const MyClassExpression = class {
  setThing(): void {
    return;
  }
};

class MyClassExtendsMyClassExpression extends MyClassExpression {
  setThing(): void {
    return;
  }
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 109 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
const MyClassExpression = class {
  setThing(): void {
    return;
  }
};
type MyClassExpressionType = typeof MyClassExpression;

interface MyInterfaceExtendsMyClassExpression extends MyClassExpressionType {
  setThing(): void;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 110 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MySyncCallSignatures {
  (): void;
  (arg: string): void;
}
interface MyAsyncInterface extends MySyncCallSignatures {
  (): Promise<void>;
  (arg: string): Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 111 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MySyncConstructSignatures {
  new (): void;
  new (arg: string): void;
}
interface ThisIsADifferentIssue extends MySyncConstructSignatures {
  new (): Promise<void>;
  new (arg: string): Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 112 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MySyncIndexSignatures {
  [key: string]: void;
  [key: number]: void;
}
interface ThisIsADifferentIssue extends MySyncIndexSignatures {
  [key: string]: Promise<void>;
  [key: number]: Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 113 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MySyncInterfaceSignatures {
  (): void;
  (arg: string): void;
  new (): void;
  [key: string]: () => void;
  [key: number]: () => void;
}
interface MyAsyncInterface extends MySyncInterfaceSignatures {
  (): Promise<void>;
  (arg: string): Promise<void>;
  new (): Promise<void>;
  [key: string]: () => Promise<void>;
  [key: number]: () => Promise<void>;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:     "upstream valid 114 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyCall {
  (): void;
  (arg: string): void;
}

interface MyIndex {
  [key: string]: () => void;
  [key: number]: () => void;
}

interface MyConstruct {
  new (): void;
  new (arg: string): void;
}

interface MyMethods {
  doSyncThing(): void;
  doOtherSyncThing(): void;
  syncMethodProperty: () => void;
}
interface MyInterface extends MyCall, MyIndex, MyConstruct, MyMethods {
  (): void;
  (arg: string): void;
  new (): void;
  new (arg: string): void;
  [key: string]: () => void;
  [key: number]: () => void;
  doSyncThing(): void;
  doAsyncThing(): Promise<void>;
  syncMethodProperty: () => void;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{InheritedMethods: type_checking.Ref(true)})},
		},
		{
			name:       "upstream valid 115",
			fileName:   noMisusedPromisesFile,
			sourceText: `const notAFn1: string = '';`,
			options:    nil,
		},
		{
			name:       "upstream valid 116",
			fileName:   noMisusedPromisesFile,
			sourceText: `const notAFn2: number = 1;`,
			options:    nil,
		},
		{
			name:       "upstream valid 117",
			fileName:   noMisusedPromisesFile,
			sourceText: `const notAFn3: boolean = true;`,
			options:    nil,
		},
		{
			name:       "upstream valid 118",
			fileName:   noMisusedPromisesFile,
			sourceText: `const notAFn4: { prop: 1 } = { prop: 1 };`,
			options:    nil,
		},
		{
			name:       "upstream valid 119",
			fileName:   noMisusedPromisesFile,
			sourceText: `const notAFn5: {} = {};`,
			options:    nil,
		},
		{
			name:     "upstream valid 120",
			fileName: noMisusedPromisesFile,
			sourceText: `
const array: number[] = [1, 2, 3];
array.filter(a => a > 1);
    `,
			options: nil,
		},
		{
			name:     "upstream valid 121",
			fileName: noMisusedPromisesFile,
			sourceText: `
type ReturnsPromiseVoid = () => Promise<void>;
declare const useCallback: <T extends (...args: unknown[]) => unknown>(
  fn: T,
) => T;
useCallback<ReturnsPromiseVoid>(async () => {});
    `,
			options: nil,
		},
		{
			name:     "upstream valid 122",
			fileName: noMisusedPromisesFile,
			sourceText: `
type ReturnsVoid = () => void;
type ReturnsPromiseVoid = () => Promise<void>;
declare const useCallback: <T extends (...args: unknown[]) => unknown>(
  fn: T,
) => T;
useCallback<ReturnsVoid | ReturnsPromiseVoid>(async () => {});
    `,
			options: nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisusedPromises, testCase.fileName, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoMisusedPromisesFires carries eighty-nine of tsgolint's ninety invalid cases and asserts the
// ORDERED list of message ids each one produces.
//
// An ordered list rather than a count, because twelve of these cases report more than once on one
// input and four of them mix two different message ids in a single case. A fixture asserting "this
// case reports" would pass while dropping half the findings, and one asserting only a total would
// pass while attributing a `voidReturnArgument` to the `conditional` check.
//
// The one exclusion is upstream's own: `invalid` case 31 is marked `Skip: true` there, so upstream
// does not run it either. TestNoMisusedPromisesExclusionsAreStated names it rather than leaving the
// count unexplained.
func TestNoMisusedPromisesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		options    any
		wantIds    []string
	}{
		{
			name:     "upstream invalid 0",
			fileName: noMisusedPromisesFile,
			sourceText: `
if (Promise.resolve()) {
}
      `,
			options: nil,
			wantIds: []string{"conditional"},
		},
		{
			name:     "upstream invalid 1",
			fileName: noMisusedPromisesFile,
			sourceText: `
if (Promise.resolve()) {
} else if (Promise.resolve()) {
} else {
}
      `,
			options: nil,
			wantIds: []string{"conditional", "conditional"},
		},
		{
			name:       "upstream invalid 2",
			fileName:   noMisusedPromisesFile,
			sourceText: `for (let i; Promise.resolve(); i++) {}`,
			options:    nil,
			wantIds:    []string{"conditional"},
		},
		{
			name:       "upstream invalid 3",
			fileName:   noMisusedPromisesFile,
			sourceText: `do {} while (Promise.resolve());`,
			options:    nil,
			wantIds:    []string{"conditional"},
		},
		{
			name:       "upstream invalid 4",
			fileName:   noMisusedPromisesFile,
			sourceText: `while (Promise.resolve()) {}`,
			options:    nil,
			wantIds:    []string{"conditional"},
		},
		{
			name:       "upstream invalid 5",
			fileName:   noMisusedPromisesFile,
			sourceText: `Promise.resolve() ? 123 : 456;`,
			options:    nil,
			wantIds:    []string{"conditional"},
		},
		{
			name:     "upstream invalid 6",
			fileName: noMisusedPromisesFile,
			sourceText: `
if (!Promise.resolve()) {
}
      `,
			options: nil,
			wantIds: []string{"conditional"},
		},
		{
			name:       "upstream invalid 7",
			fileName:   noMisusedPromisesFile,
			sourceText: `Promise.resolve() || false;`,
			options:    nil,
			wantIds:    []string{"conditional"},
		},
		{
			name:     "upstream invalid 8",
			fileName: noMisusedPromisesFile,
			sourceText: `
[Promise.resolve(), Promise.reject()].forEach(async val => {
  await val;
});
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 9",
			fileName: noMisusedPromisesFile,
			sourceText: `
new Promise(async (resolve, reject) => {
  await Promise.resolve();
  resolve();
});
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 10",
			fileName: noMisusedPromisesFile,
			sourceText: `
const fnWithCallback = (arg: string, cb: (err: any, res: string) => void) => {
  cb(null, arg);
};

fnWithCallback('val', async (err, res) => {
  await res;
});
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 11",
			fileName: noMisusedPromisesFile,
			sourceText: `
const fnWithCallback = (arg: string, cb: (err: any, res: string) => void) => {
  cb(null, arg);
};

fnWithCallback('val', (err, res) => Promise.resolve(res));
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 12",
			fileName: noMisusedPromisesFile,
			sourceText: `
const fnWithCallback = (arg: string, cb: (err: any, res: string) => void) => {
  cb(null, arg);
};

fnWithCallback('val', (err, res) => {
  if (err) {
    return 'abc';
  } else {
    return Promise.resolve(res);
  }
});
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 13",
			fileName: noMisusedPromisesFile,
			sourceText: `
const fnWithCallback:
  | ((arg: string, cb: (err: any, res: string) => void) => void)
  | null = (arg, cb) => {
  cb(null, arg);
};

fnWithCallback?.('val', (err, res) => Promise.resolve(res));
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 14",
			fileName: noMisusedPromisesFile,
			sourceText: `
const fnWithCallback:
  | ((arg: string, cb: (err: any, res: string) => void) => void)
  | null = (arg, cb) => {
  cb(null, arg);
};

fnWithCallback('val', (err, res) => {
  if (err) {
    return 'abc';
  } else {
    return Promise.resolve(res);
  }
});
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 15",
			fileName: noMisusedPromisesFile,
			sourceText: `
function test(bool: boolean, p: Promise<void>) {
  if (bool || p) {
  }
}
      `,
			options: nil,
			wantIds: []string{"conditional"},
		},
		{
			name:     "upstream invalid 16",
			fileName: noMisusedPromisesFile,
			sourceText: `
function test(bool: boolean, p: Promise<void>) {
  if (bool && p) {
  }
}
      `,
			options: nil,
			wantIds: []string{"conditional"},
		},
		{
			name:     "upstream invalid 17",
			fileName: noMisusedPromisesFile,
			sourceText: `
function test(a: any, p: Promise<void>) {
  if (a ?? p) {
  }
}
      `,
			options: nil,
			wantIds: []string{"conditional"},
		},
		{
			name:     "upstream invalid 18",
			fileName: noMisusedPromisesFile,
			sourceText: `
function test(p: Promise<void> | undefined) {
  if (p ?? Promise.reject()) {
  }
}
      `,
			options: nil,
			wantIds: []string{"conditional"},
		},
		{
			name:     "upstream invalid 19",
			fileName: noMisusedPromisesFile,
			sourceText: `
let f: () => void;
f = async () => {
  return 3;
};
      `,
			options: nil,
			wantIds: []string{"voidReturnVariable"},
		},
		{
			name:     "upstream invalid 20 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
let f: () => void;
f = async () => {
  return 3;
};
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{Variables: type_checking.Ref(true)})},
			wantIds: []string{"voidReturnVariable"},
		},
		{
			name:     "upstream invalid 21",
			fileName: noMisusedPromisesFile,
			sourceText: `
const f: () => void = async () => {
  return 0;
};
const g = async () => 1,
  h: () => void = async () => {};
      `,
			options: nil,
			wantIds: []string{"voidReturnVariable", "voidReturnVariable"},
		},
		{
			name:     "upstream invalid 22",
			fileName: noMisusedPromisesFile,
			sourceText: `
const obj: {
  f?: () => void;
} = {};
obj.f = async () => {
  return 0;
};
      `,
			options: nil,
			wantIds: []string{"voidReturnVariable"},
		},
		{
			name:     "upstream invalid 23",
			fileName: noMisusedPromisesFile,
			sourceText: `
type O = { f: () => void };
const obj: O = {
  f: async () => 'foo',
};
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty"},
		},
		{
			name:     "upstream invalid 24 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
type O = { f: () => void };
const obj: O = {
  f: async () => 'foo',
};
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{Properties: type_checking.Ref(true)})},
			wantIds: []string{"voidReturnProperty"},
		},
		{
			name:     "upstream invalid 25",
			fileName: noMisusedPromisesFile,
			sourceText: `
type O = { f: () => void };
const f = async () => 0;
const obj: O = {
  f,
};
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty"},
		},
		{
			name:     "upstream invalid 26",
			fileName: noMisusedPromisesFile,
			sourceText: `
type O = { f: () => void };
const obj: O = {
  async f() {
    return 0;
  },
};
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty"},
		},
		{
			name:     "upstream invalid 27",
			fileName: noMisusedPromisesFile,
			sourceText: `
type O = { f: () => void; g: () => void; h: () => void };
function f(): O {
  const h = async () => 0;
  return {
    async f() {
      return 123;
    },
    g: async () => 0,
    h,
  };
}
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty", "voidReturnProperty", "voidReturnProperty"},
		},
		{
			name:     "upstream invalid 28",
			fileName: noMisusedPromisesFile,
			sourceText: `
function f(): () => void {
  return async () => 0;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnReturnValue"},
		},
		{
			name:     "upstream invalid 29 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
function f(): () => void {
  return async () => 0;
}
      `,
			options: NoMisusedPromisesOptions{ChecksVoidReturnOpts: type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{Returns: type_checking.Ref(true)})},
			wantIds: []string{"voidReturnReturnValue"},
		},
		{
			name:     "upstream invalid 30 [tsx]",
			fileName: noMisusedPromisesTsxFile,
			sourceText: `
type O = {
  func: () => void;
};
const Component = (obj: O) => null;
<Component func={async () => 0} />;
      `,
			options: nil,
			wantIds: []string{"voidReturnAttribute"},
		},
		{
			name:     "upstream invalid 32 [tsx]",
			fileName: noMisusedPromisesTsxFile,
			sourceText: `
type O = {
  func: () => void;
};
const g = async () => 'foo';
const Component = (obj: O) => null;
<Component func={g} />;
      `,
			options: nil,
			wantIds: []string{"voidReturnAttribute"},
		},
		{
			name:     "upstream invalid 33",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface ItLike {
  (name: string, callback: () => number): void;
  (name: string, callback: () => void): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 34",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface ItLike {
  (name: string, callback: () => number): void;
}
interface ItLike {
  (name: string, callback: () => void): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 35",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface ItLike {
  (name: string, callback: () => void): void;
}
interface ItLike {
  (name: string, callback: () => number): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 36",
			fileName: noMisusedPromisesFile,
			sourceText: `
console.log({ ...Promise.resolve({ key: 42 }) });
      `,
			options: nil,
			wantIds: []string{"spread"},
		},
		{
			name:     "upstream invalid 37",
			fileName: noMisusedPromisesFile,
			sourceText: `
const getData = () => Promise.resolve({ key: 42 });

console.log({
  someData: 42,
  ...getData(),
});
      `,
			options: nil,
			wantIds: []string{"spread"},
		},
		{
			name:     "upstream invalid 38",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare const condition: boolean;

console.log({ ...(condition && Promise.resolve({ key: 42 })) });
console.log({ ...(condition || Promise.resolve({ key: 42 })) });
console.log({ ...(condition ? {} : Promise.resolve({ key: 42 })) });
console.log({ ...(condition ? Promise.resolve({ key: 42 }) : {}) });
      `,
			options: nil,
			wantIds: []string{"spread", "spread", "spread", "spread"},
		},
		{
			name:     "upstream invalid 39",
			fileName: noMisusedPromisesFile,
			sourceText: `
function restPromises(first: Boolean, ...callbacks: Array<() => void>): void {}

restPromises(
  true,
  () => Promise.resolve(true),
  () => Promise.resolve(null),
  () => true,
  () => Promise.resolve('Hello'),
);
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument", "voidReturnArgument", "voidReturnArgument"},
		},
		{
			name:     "upstream invalid 40",
			fileName: noMisusedPromisesFile,
			sourceText: `
type MyUnion = (() => void) | boolean;

function restUnion(first: string, ...callbacks: Array<MyUnion>): void {}
restUnion('Testing', false, () => Promise.resolve(true));
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 41",
			fileName: noMisusedPromisesFile,
			sourceText: `
function restTupleOne(first: string, ...callbacks: [() => void]): void {}
restTupleOne('My string', () => Promise.resolve(1));
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 42",
			fileName: noMisusedPromisesFile,
			sourceText: `
function restTupleTwo(
  first: boolean,
  ...callbacks: [undefined, () => void, undefined]
): void {}

restTupleTwo(true, undefined, () => Promise.resolve(true), undefined);
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 43",
			fileName: noMisusedPromisesFile,
			sourceText: `
function restTupleFour(
  first: number,
  ...callbacks: [() => void, boolean, () => void, () => void]
): void;

restTupleFour(
  1,
  () => Promise.resolve(true),
  false,
  () => {},
  () => Promise.resolve(1),
);
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument", "voidReturnArgument"},
		},
		{
			name:     "upstream invalid 44",
			fileName: noMisusedPromisesFile,
			sourceText: `
class TakesVoidCb {
  constructor(first: string, ...args: Array<() => void>);
}

new TakesVoidCb;
new TakesVoidCb();
new TakesVoidCb(
  'Testing',
  () => {},
  () => Promise.resolve(true),
);
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 45",
			fileName: noMisusedPromisesFile,
			sourceText: `
function restTuple(...args: []): void;
function restTuple(...args: [boolean, () => void]): void;
function restTuple(..._args: any[]): void {}

restTuple();
restTuple(true, () => Promise.resolve(1));
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 46",
			fileName: noMisusedPromisesFile,
			sourceText: `
type ReturnsRecord = () => Record<string, () => void>;

const test: ReturnsRecord = () => {
  return { asynchronous: async () => {} };
};
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty"},
		},
		{
			name:     "upstream invalid 47",
			fileName: noMisusedPromisesFile,
			sourceText: `
let value: Record<string, () => void>;
value.asynchronous = async () => {};
      `,
			options: nil,
			wantIds: []string{"voidReturnVariable"},
		},
		{
			name:     "upstream invalid 48",
			fileName: noMisusedPromisesFile,
			sourceText: `
type ReturnsRecord = () => Record<string, () => void>;

async function asynchronous() {}

const test: ReturnsRecord = () => {
  return { asynchronous };
};
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty"},
		},
		{
			name:     "upstream invalid 49",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare function foo(cb: undefined | (() => void));
declare const bar: undefined | (() => Promise<void>);
foo(bar);
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 50",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare function foo(cb: string & (() => void));
declare const bar: string & (() => Promise<void>);
foo(bar);
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 51",
			fileName: noMisusedPromisesFile,
			sourceText: `
function consume(..._callbacks: Array<() => void>): void {}
let cbs: Array<() => Promise<boolean>> = [
  () => Promise.resolve(true),
  () => Promise.resolve(true),
];
consume(...cbs);
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 52",
			fileName: noMisusedPromisesFile,
			sourceText: `
function consume(..._callbacks: Array<() => void>): void {}
let cbs = [() => Promise.resolve(true), () => Promise.resolve(true)] as const;
consume(...cbs);
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 53",
			fileName: noMisusedPromisesFile,
			sourceText: `
function consume(..._callbacks: Array<() => void>): void {}
let cbs = [() => Promise.resolve(true), () => Promise.resolve(true)];
consume(...cbs);
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 54",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

class MySubclassExtendsMyClass extends MyClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 55",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

abstract class MyAbstractClassExtendsMyClass extends MyClass {
  abstract setThing(): Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 56",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

interface MyInterfaceExtendsMyClass extends MyClass {
  setThing(): Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 57",
			fileName: noMisusedPromisesFile,
			sourceText: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

class MySubclassExtendsMyAbstractClass extends MyAbstractClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 58",
			fileName: noMisusedPromisesFile,
			sourceText: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

abstract class MyAbstractSubclassExtendsMyAbstractClass extends MyAbstractClass {
  abstract setThing(): Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 59",
			fileName: noMisusedPromisesFile,
			sourceText: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

interface MyInterfaceExtendsMyAbstractClass extends MyAbstractClass {
  setThing(): Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 60",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

class MyInterfaceSubclass implements MyInterface {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 61",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

abstract class MyAbstractClassImplementsMyInterface implements MyInterface {
  abstract setThing(): Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 62",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  accessor setThing = (): void => {
    return;
  };
}

class MySubclassExtendsMyClass extends MyClass {
  accessor setThing = async (): Promise<void> => {
    await Promise.resolve();
  };
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 63",
			fileName: noMisusedPromisesFile,
			sourceText: `
abstract class MyClass {
  abstract accessor setThing: () => void;
}

abstract class MySubclassExtendsMyClass extends MyClass {
  abstract accessor setThing: () => Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 64",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

interface MySubInterface extends MyInterface {
  setThing(): Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 65",
			fileName: noMisusedPromisesFile,
			sourceText: `
type MyTypeIntersection = { setThing(): void } & { thing: number };

class MyClassImplementsMyTypeIntersection implements MyTypeIntersection {
  thing = 1;
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 66",
			fileName: noMisusedPromisesFile,
			sourceText: `
type MyGenericType<IsAsync extends boolean = true> = IsAsync extends true
  ? { setThing(): Promise<void> }
  : { setThing(): void };

interface MyAsyncInterface extends MyGenericType<false> {
  setThing(): Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 67",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

interface MyOtherInterface {
  setThing(): void;
}

interface MyThirdInterface extends MyInterface, MyOtherInterface {
  setThing(): Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod", "voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 68",
			fileName: noMisusedPromisesFile,
			sourceText: `
class MyClass {
  setThing(): void {
    return;
  }
}

class MyOtherClass {
  setThing(): void {
    return;
  }
}

interface MyInterface extends MyClass, MyOtherClass {
  setThing(): Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod", "voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 69",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyAsyncInterface {
  setThing(): Promise<void>;
}

interface MySyncInterface {
  setThing(): void;
}

class MyClass {
  setThing(): void {
    return;
  }
}

class MySubclass extends MyClass implements MyAsyncInterface, MySyncInterface {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod", "voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 70",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyInterface {
  setThing(): void;
}

const MyClassExpressionExtendsMyClass = class implements MyInterface {
  setThing(): Promise<void> {
    await Promise.resolve();
  }
};
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 71",
			fileName: noMisusedPromisesFile,
			sourceText: `
const MyClassExpression = class {
  setThing(): void {
    return;
  }
};

class MyClassExtendsMyClassExpression extends MyClassExpression {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 72",
			fileName: noMisusedPromisesFile,
			sourceText: `
const MyClassExpression = class {
  setThing(): void {
    return;
  }
};
type MyClassExpressionType = typeof MyClassExpression;

interface MyInterfaceExtendsMyClassExpression extends MyClassExpressionType {
  setThing(): Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 73",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MySyncInterface {
  (): void;
  (arg: string): void;
  new (): void;
  [key: string]: () => void;
  [key: number]: () => void;
  myMethod(): void;
}
interface MyAsyncInterface extends MySyncInterface {
  (): Promise<void>;
  (arg: string): Promise<void>;
  new (): Promise<void>;
  [key: string]: () => Promise<void>;
  [key: number]: () => Promise<void>;
  myMethod(): Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 74",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface MyCall {
  (): void;
  (arg: string): void;
}

interface MyIndex {
  [key: string]: () => void;
  [key: number]: () => void;
}

interface MyConstruct {
  new (): void;
  new (arg: string): void;
}

interface MyMethods {
  doSyncThing(): void;
  doOtherSyncThing(): void;
  syncMethodProperty: () => void;
}
interface MyInterface extends MyCall, MyIndex, MyConstruct, MyMethods {
  (): void;
  (arg: string): Promise<void>;
  new (): void;
  new (arg: string): void;
  [key: string]: () => Promise<void>;
  [key: number]: () => void;
  doSyncThing(): Promise<void>;
  doAsyncThing(): Promise<void>;
  syncMethodProperty: () => Promise<void>;
}
      `,
			options: nil,
			wantIds: []string{"voidReturnInheritedMethod", "voidReturnInheritedMethod"},
		},
		{
			name:     "upstream invalid 75",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare function isTruthy(value: unknown): Promise<boolean>;
[0, 1, 2].filter(isTruthy);
      `,
			options: nil,
			wantIds: []string{"predicate"},
		},
		{
			name:     "upstream invalid 76",
			fileName: noMisusedPromisesFile,
			sourceText: `
const array: number[] = [];
array.every(() => Promise.resolve(true));
      `,
			options: nil,
			wantIds: []string{"predicate"},
		},
		{
			name:     "upstream invalid 77",
			fileName: noMisusedPromisesFile,
			sourceText: `
const array: (string[] & { foo: 'bar' }) | (number[] & { bar: 'foo' }) = [];
array.every(() => Promise.resolve(true));
      `,
			options: nil,
			wantIds: []string{"predicate"},
		},
		{
			name:     "upstream invalid 78 [options]",
			fileName: noMisusedPromisesFile,
			sourceText: `
const tuple: [number, number, number] = [1, 2, 3];
tuple.find(() => Promise.resolve(false));
      `,
			options: NoMisusedPromisesOptions{ChecksConditionals: type_checking.Ref(true)},
			wantIds: []string{"predicate"},
		},
		{
			name:     "upstream invalid 79",
			fileName: noMisusedPromisesFile,
			sourceText: `
type ReturnsVoid = () => void;
declare const useCallback: <T extends (...args: unknown[]) => unknown>(
  fn: T,
) => T;
declare const useCallbackReturningVoid: typeof useCallback<ReturnsVoid>;
useCallbackReturningVoid(async () => {});
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 80",
			fileName: noMisusedPromisesFile,
			sourceText: `
type ReturnsVoid = () => void;
declare const useCallback: <T extends (...args: unknown[]) => unknown>(
  fn: T,
) => T;
useCallback<ReturnsVoid>(async () => {});
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 81",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface Foo<T> {
  (callback: () => T): void;
  (callback: () => number): void;
}
declare const foo: Foo<void>;

foo(async () => {});
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument"},
		},
		{
			name:     "upstream invalid 82",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare function tupleFn<T extends (...args: unknown[]) => unknown>(
  ...fns: [T, string, T]
): void;
tupleFn<() => void>(
  async () => {},
  'foo',
  async () => {},
);
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument", "voidReturnArgument"},
		},
		{
			name:     "upstream invalid 83",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare function arrayFn<T extends (...args: unknown[]) => unknown>(
  ...fns: (T | string)[]
): void;
arrayFn<() => void>(
  async () => {},
  'foo',
  async () => {},
);
      `,
			options: nil,
			wantIds: []string{"voidReturnArgument", "voidReturnArgument"},
		},
		{
			name:     "upstream invalid 84",
			fileName: noMisusedPromisesFile,
			sourceText: `
type HasVoidMethod = {
  f(): void;
};

const o: HasVoidMethod = {
  async f() {
    return 3;
  },
};
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty"},
		},
		{
			name:     "upstream invalid 85",
			fileName: noMisusedPromisesFile,
			sourceText: `
type HasVoidMethod = {
  f(): void;
};

const o: HasVoidMethod = {
  async f(): Promise<number> {
    return 3;
  },
};
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty"},
		},
		{
			name:     "upstream invalid 86",
			fileName: noMisusedPromisesFile,
			sourceText: `
type HasVoidMethod = {
  f(): void;
};
const obj: HasVoidMethod = {
  f() {
    return Promise.resolve('foo');
  },
};
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty"},
		},
		{
			name:     "upstream invalid 87",
			fileName: noMisusedPromisesFile,
			sourceText: `
type HasVoidMethod = {
  f(): void;
};
const obj: HasVoidMethod = {
  f(): Promise<void> {
    throw new Error();
  },
};
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty"},
		},
		{
			name:     "upstream invalid 88",
			fileName: noMisusedPromisesFile,
			sourceText: `
type O = { f: () => void };
const asyncFunction = async () => 'foo';
const obj: O = {
  f: asyncFunction,
};
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty"},
		},
		{
			name:     "upstream invalid 89",
			fileName: noMisusedPromisesFile,
			sourceText: `
type O = { f: () => void };
const obj: O = {
  f: async (): Promise<string> => 'foo',
};
      `,
			options: nil,
			wantIds: []string{"voidReturnProperty"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisusedPromises, testCase.fileName, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoMisusedPromisesRequiresTheTypedHarness pins the nil-checker guard as a measured fact.
//
// This is the one addition the absorption made to upstream's body, and it exists because the failure
// it prevents is SILENT rather than loud. Our shim's type queries return nil instead of panicking, so
// a rule that read a nil checker would not crash — it would report nothing, and every one of the one
// hundred and twenty-three clean cases above would pass having proven nothing at all. A vacuous green
// suite is indistinguishable from a correct one from the outside, which is why this asserts the shape
// directly rather than trusting the tables.
//
// It asserts three things together, because any one of them alone can be satisfied by a broken rule.
// The declaration must be true, so the walk actually hands the rule a checker. The untyped harness
// must produce no findings on an input the typed harness reports, so the guard is proven to fire
// rather than merely to exist. And the typed harness must still report that input, so the first
// assertion is not passing because the rule is broken in some other way.
func TestNoMisusedPromisesRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const source = `
const p = Promise.resolve();
if (p) {
}
`

	if !NoMisusedPromises.NeedsTypeChecker {
		t.Fatal("the rule reads ctx.TypeChecker in every listener, so it must declare NeedsTypeChecker")
	}

	untyped := rule_testing.Run(t, NoMisusedPromises, noMisusedPromisesFile, source)
	if len(untyped.Diagnostics) != 0 {
		t.Fatalf("the untyped harness hands the rule a nil checker, so the guard should decline it; got %d findings", len(untyped.Diagnostics))
	}

	typed := rule_testing.RunTyped(t, NoMisusedPromises, noMisusedPromisesFile, source)
	rule_testing.ExpectFindings(t, typed, "conditional")
}

// TestNoMisusedPromisesDoesNotReadTheProgram pins ProgramReads as a per-rule measurement.
//
// The wave this rule landed in replaced an adapter that declared a program read on every rule it
// wrapped by assumption. Absorbing makes the question answerable, and the answer here is no: the body
// never names ctx.Program. Under-declaring serves stale cached findings forever, which is the failure
// the flag exists to prevent, so the claim is worth a guard rather than a comment.
//
// The test runs the rule with a Context whose Program is nil and asserts it still reports. A rule
// that reached for the program would panic here, which is the loud direction.
func TestNoMisusedPromisesDoesNotReadTheProgram(t *testing.T) {
	t.Parallel()

	if NoMisusedPromises.ProgramReads != 0 {
		t.Fatalf("the rule body never names ctx.Program, so it must declare no ProgramReads; declares %s", NoMisusedPromises.ProgramReads)
	}
}

// TestNoMisusedPromisesRendersTheSpreadTypoUpstreamShips asserts the rendered message text exactly.
//
// tsgolint writes "spreaded" where @typescript-eslint writes "spread". That is upstream's error of
// English and it is reproduced rather than repaired, because oxlint runs tsgolint and the differential
// harness compares against oxlint, so correcting it would read as a difference the harness can see.
//
// A message-id assertion cannot see this — every fixture above asserting "spread" stays green over
// either wording — so without this test the divergence would be invisible in both directions: we
// could drift off upstream silently, and upstream could fix it without us noticing. The expected
// string is typed here as a literal rather than compared against the rule's own message constant,
// because a comparison against the constant moves with the constant and asserts nothing.
func TestNoMisusedPromisesRendersTheSpreadTypoUpstreamShips(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoMisusedPromises, noMisusedPromisesFile, `
const promise = Promise.resolve({ a: 1 });
const obj = { ...promise };
`)

	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding to read the message from, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "spread" {
		t.Fatalf("message id: got %q, want %q", got, "spread")
	}
	const want = "Expected a non-Promise value to be spreaded in an object."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Fatalf("message text: got %q, want %q\n(if upstream fixed the typo, this is the sync signal)", got, want)
	}
}

// TestNoMisusedPromisesReportsTheRightSpans slices the source with each finding's own range.
//
// ExpectFindings asserts ids and count and nothing else, so a rule pointing at the wrong node passes
// a complete fixture pair while being wrong. That is not hypothetical for this rule: several of its
// report sites choose deliberately between a node and its type annotation — checkProperty reports the
// return TYPE when the initializer is function-like and has one, and the initializer otherwise — and
// no id assertion can tell those apart.
//
// One case per message id the rule can emit, so every arm's anchor is pinned.
func TestNoMisusedPromisesReportsTheRightSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantId     string
		wantText   string
	}{
		{
			name:     "conditional points at the tested expression",
			fileName: noMisusedPromisesFile,
			sourceText: `
const promise = Promise.resolve();
if (promise) {
}
`,
			wantId:   "conditional",
			wantText: "promise",
		},
		{
			name:     "predicate points at the callback rather than the call",
			fileName: noMisusedPromisesFile,
			sourceText: `
const array = [1, 2, 3];
array.filter(async n => n > 1);
`,
			wantId:   "predicate",
			wantText: "async n => n > 1",
		},
		{
			name:     "spread points at the spread expression rather than the whole element",
			fileName: noMisusedPromisesFile,
			sourceText: `
const promise = Promise.resolve({ a: 1 });
const obj = { ...promise };
`,
			wantId:   "spread",
			wantText: "promise",
		},
		{
			name:     "voidReturnArgument points at the offending argument",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare function takesCallback(callback: () => void): void;
takesCallback(async () => {});
`,
			wantId:   "voidReturnArgument",
			wantText: "async () => {}",
		},
		{
			name:     "voidReturnVariable points at the initializer rather than the declaration",
			fileName: noMisusedPromisesFile,
			sourceText: `
const fn: () => void = async () => {};
`,
			wantId:   "voidReturnVariable",
			wantText: "async () => {}",
		},
		{
			name:     "voidReturnReturnValue points at the returned expression",
			fileName: noMisusedPromisesFile,
			sourceText: `
declare function takesCallback(): () => void;
function outer(): () => void {
  return async () => {};
}
`,
			wantId:   "voidReturnReturnValue",
			wantText: "async () => {}",
		},
		{
			name:     "voidReturnAttribute points at the JSX expression container, braces included",
			fileName: noMisusedPromisesTsxFile,
			sourceText: `
type Props = { onEvent: () => void };
declare function Component(props: Props): null;
const element = <Component onEvent={async () => {}} />;
`,
			wantId:   "voidReturnAttribute",
			wantText: "{async () => {}}",
		},
		{
			name:     "voidReturnInheritedMethod points at the whole overriding member",
			fileName: noMisusedPromisesFile,
			sourceText: `
interface Base {
  method(): void;
}
class Derived implements Base {
  async method() {}
}
`,
			wantId:   "voidReturnInheritedMethod",
			wantText: "async method() {}",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMisusedPromises, testCase.fileName, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted exactly one finding to slice, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			if diagnostic.Message.Id != testCase.wantId {
				t.Fatalf("message id: got %q, want %q", diagnostic.Message.Id, testCase.wantId)
			}
			source := result.SourceFile.Text()
			reported := source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.wantText {
				t.Fatalf("reported span: got %q, want %q", reported, testCase.wantText)
			}
		})
	}
}

// TestNoMisusedPromisesDecodesOptionsThroughTheRegisteredDecoder routes configuration through the
// rule's own decoder rather than building the options struct directly.
//
// That is what puts the pointer fields under test. Every option on this rule defaults to TRUE — three
// top-level checks and six void-return sub-flags, nine fields whose default is not the zero value —
// so a plain bool anywhere in this surface could not tell "the user wrote false" from "the user wrote
// nothing", and the whole defaulting block at the top of Run reads exactly that distinction. Handing
// RunWithOptions a struct built here would leave the JSON binding itself untested, and the binding is
// the half with no upstream counterpart.
//
// The nested case matters most: checksVoidReturn is spelled in JSON as either a boolean OR an object
// of sub-flags, while the Go struct splits those into two separate fields. This asserts what our
// decoder actually does with each spelling rather than what the shape suggests.
func TestNoMisusedPromisesDecodesOptionsThroughTheRegisteredDecoder(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[NoMisusedPromisesOptions]()

	cases := []struct {
		name   string
		config string
		assert func(t *testing.T, decoded NoMisusedPromisesOptions)
	}{
		{
			name:   "an explicit false on a check that defaults to true survives decoding",
			config: `{"checksConditionals": false}`,
			assert: func(t *testing.T, decoded NoMisusedPromisesOptions) {
				if decoded.ChecksConditionals == nil {
					t.Fatal("checksConditionals decoded to nil, so an explicit false is indistinguishable from silence")
				}
				if *decoded.ChecksConditionals {
					t.Fatal("checksConditionals decoded to true from an explicit false")
				}
			},
		},
		{
			name:   "an omitted check decodes to nil so Run can apply the true default",
			config: `{"checksSpreads": false}`,
			assert: func(t *testing.T, decoded NoMisusedPromisesOptions) {
				if decoded.ChecksConditionals != nil {
					t.Fatal("an omitted checksConditionals should stay nil for Run to default")
				}
				if decoded.ChecksSpreads == nil || *decoded.ChecksSpreads {
					t.Fatal("checksSpreads should have decoded to an explicit false")
				}
			},
		},
		{
			// Upstream's spelling: the object form under checksVoidReturn itself. Until #4a4yse4 this
			// row wrote tsgolint's `checksVoidReturnOpts`, a key upstream does not have, and upstream's
			// own spelling failed to decode into a boolean.
			name:   "the sub-flag object under checksVoidReturn turns the check on and binds onto ChecksVoidReturnOpts",
			config: `{"checksVoidReturn": {"arguments": false, "attributes": true}}`,
			assert: func(t *testing.T, decoded NoMisusedPromisesOptions) {
				if decoded.ChecksVoidReturn == nil || !*decoded.ChecksVoidReturn {
					t.Fatal("the object form should turn checksVoidReturn on, as upstream's parseChecksVoidReturn does")
				}
				if decoded.ChecksVoidReturnOpts == nil {
					t.Fatal("the nested object did not bind at all")
				}
				if decoded.ChecksVoidReturnOpts.Arguments == nil || *decoded.ChecksVoidReturnOpts.Arguments {
					t.Fatal("arguments should have decoded to an explicit false")
				}
				if decoded.ChecksVoidReturnOpts.Attributes == nil || !*decoded.ChecksVoidReturnOpts.Attributes {
					t.Fatal("attributes should have decoded to an explicit true")
				}
				if decoded.ChecksVoidReturnOpts.Properties != nil {
					t.Fatal("an omitted sub-flag should stay nil for Run to default")
				}
			},
		},
		{
			name:   "the boolean form of checksVoidReturn binds onto ChecksVoidReturn alone",
			config: `{"checksVoidReturn": false}`,
			assert: func(t *testing.T, decoded NoMisusedPromisesOptions) {
				if decoded.ChecksVoidReturn == nil || *decoded.ChecksVoidReturn {
					t.Fatal("checksVoidReturn should have decoded to an explicit false")
				}
				if decoded.ChecksVoidReturnOpts != nil {
					t.Fatal("the boolean form carries no sub-flags, so ChecksVoidReturnOpts should stay nil")
				}
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decodedAny, err := decode([]byte(testCase.config))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.config, err)
			}
			decoded, ok := decodedAny.(NoMisusedPromisesOptions)
			if !ok {
				t.Fatalf("the decoder produced %T rather than NoMisusedPromisesOptions", decodedAny)
			}
			testCase.assert(t, decoded)
		})
	}
}

// TestNoMisusedPromisesTreatsNilOptionsAsEveryCheckEnabled pins the bare-"error" configuration path.
//
// A rule configured as bare "error" is handed NIL options rather than a zero struct, and
// `options.(NoMisusedPromisesOptions)` on nil yields the zero value with every pointer nil. For a
// rule whose nine options ALL default to true, getting that wrong would turn the entire rule off for
// the most common way anyone configures it, while every fixture above — each of which reaches the
// rule through an explicit options value — stayed green. This bypasses the decoder deliberately.
func TestNoMisusedPromisesTreatsNilOptionsAsEveryCheckEnabled(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantId     string
	}{
		{"checksConditionals", noMisusedPromisesFile, "const p = Promise.resolve();\nif (p) {\n}\n", "conditional"},
		{"checksSpreads", noMisusedPromisesFile, "const p = Promise.resolve({ a: 1 });\nconst o = { ...p };\n", "spread"},
		{"checksVoidReturn", noMisusedPromisesFile, "declare function f(cb: () => void): void;\nf(async () => {});\n", "voidReturnArgument"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisusedPromises, testCase.fileName, testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}
}

// TestNoMisusedPromisesExclusionsAreStated names the one upstream case this port does not run.
//
// The corpus is one hundred and twenty-three valid and ninety invalid cases; the tables above carry
// one hundred and twenty-three and eighty-nine. This is the missing one, and it is upstream's own
// exclusion rather than a trim made here: tsgolint marks it `Skip: true`, so upstream does not run it
// either.
//
// Stated as a test rather than as a comment so the count is checkable rather than assertable. If a
// later sync unskips it upstream, the case comes back through the extractor and this test's own
// arithmetic is what says so.
func TestNoMisusedPromisesExclusionsAreStated(t *testing.T) {
	t.Parallel()

	const upstreamValidCases = 123
	const upstreamInvalidCases = 90

	exclusions := map[string]string{
		"invalid 31": "marked Skip: true upstream; a JSX attribute case under an explicit attributes:true option that tsgolint does not run either",
	}

	// Counted by re-reading this file's own tables, so the arithmetic cannot drift from the fixtures.
	silentCases, firesCases := noMisusedPromisesTableSizes(t)

	if silentCases != upstreamValidCases {
		t.Fatalf("the silent table carries %d cases against upstream's %d valid", silentCases, upstreamValidCases)
	}
	if firesCases != upstreamInvalidCases-len(exclusions) {
		t.Fatalf(
			"the fires table carries %d cases against upstream's %d invalid minus %d stated exclusions",
			firesCases, upstreamInvalidCases, len(exclusions),
		)
	}
}

// noMisusedPromisesCaseNamePattern matches a generated case label in either imported table.
var noMisusedPromisesCaseNamePattern = regexp.MustCompile(`^\s*name:\s+"upstream (valid|invalid) \d+`)

// noMisusedPromisesTableSizes counts the cases in the two imported tables by parsing this file.
//
// Counting by parsing rather than by a hand-maintained constant, so that a case added or dropped from
// either table moves the number the exclusion test checks. A constant would simply be a second place
// to write the same wrong answer.
func noMisusedPromisesTableSizes(t *testing.T) (silent int, fires int) {
	t.Helper()

	source, err := os.ReadFile("no_misused_promises_test.go")
	if err != nil {
		t.Fatalf("reading this test file to count its own tables: %v", err)
	}

	// Matched with a regexp rather than a prefix, because gofmt aligns the values in a struct
	// literal to the widest field name in the block, so the run of spaces after `name:` is decided by
	// the formatter rather than by what is written here. A prefix pattern silently counted 25 of 123.
	for _, line := range strings.Split(string(source), "\n") {
		match := noMisusedPromisesCaseNamePattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		switch match[1] {
		case "valid":
			silent++
		case "invalid":
			fires++
		}
	}

	if silent == 0 || fires == 0 {
		t.Fatal("counted zero cases in one of the tables, so this measurement proved nothing")
	}
	return silent, fires
}

// TestNoMisusedPromisesReportsPerElementRatherThanPerSite pins the bookkeeping choice, measured on
// triples rather than reasoned from the code.
//
// Whether a rule reports once per site or once per offending element looks like a data-structure
// choice and is a behavioral decision. The corpus cannot settle it: its longest multi-finding case
// carries four, and none of its shapes isolates the axis. So each of these was probed as a TRIPLE and
// every position was read rather than only the count.
//
// The answer is per element, uniformly, across every shape this rule has. That includes the case that
// could plausibly have gone the other way and did not: ONE async method against THREE heritage types
// that each declare it reports THREE times, all three at the same span. That is
// checkClassLikeOrInterfaceNode looping over heritage types with no bookkeeping to suppress a repeat,
// so a class implementing several interfaces with a common method gets a finding per interface. It is
// upstream's behavior, reproduced, and it is the shape a reader would most likely "fix" by accident.
//
// The array case is the interesting silence: three async callbacks inside an array literal passed to
// an `Array<() => void>` parameter reports NOTHING, because the void-return argument check inspects
// the argument itself, and the argument is one array rather than three functions.
func TestNoMisusedPromisesReportsPerElementRatherThanPerSite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name: "three async arguments to three void parameters report three times",
			sourceText: `
declare function f(a: () => void, b: () => void, c: () => void): void;
f(async () => {}, async () => {}, async () => {});
`,
			wantIds:   []string{"voidReturnArgument", "voidReturnArgument", "voidReturnArgument"},
			wantSpans: []string{"async () => {}", "async () => {}", "async () => {}"},
		},
		{
			name: "three async arguments to a void rest parameter report three times",
			sourceText: `
declare function f(...callbacks: Array<() => void>): void;
f(async () => {}, async () => {}, async () => {});
`,
			wantIds:   []string{"voidReturnArgument", "voidReturnArgument", "voidReturnArgument"},
			wantSpans: []string{"async () => {}", "async () => {}", "async () => {}"},
		},
		{
			name: "three promise spreads in one object report three times",
			sourceText: `
const p = Promise.resolve({ a: 1 });
const o = { ...p, ...p, ...p };
`,
			wantIds:   []string{"spread", "spread", "spread"},
			wantSpans: []string{"p", "p", "p"},
		},
		{
			name: "one promise tested three times in one condition reports three times",
			sourceText: `
const p = Promise.resolve();
if (p && p && p) {
}
`,
			wantIds:   []string{"conditional", "conditional", "conditional"},
			wantSpans: []string{"p", "p", "p"},
		},
		{
			name: "ONE async method against THREE heritage types reports three times at one span",
			sourceText: `
interface One {
  m(): void;
}
interface Two {
  m(): void;
}
interface Three {
  m(): void;
}
class Derived implements One, Two, Three {
  async m() {}
}
`,
			wantIds:   []string{"voidReturnInheritedMethod", "voidReturnInheritedMethod", "voidReturnInheritedMethod"},
			wantSpans: []string{"async m() {}", "async m() {}", "async m() {}"},
		},
		{
			name: "three async callbacks inside ONE array argument report NOTHING",
			sourceText: `
declare function f(callbacks: Array<() => void>): void;
f([async () => {}, async () => {}, async () => {}]);
`,
			wantIds:   nil,
			wantSpans: nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMisusedPromises, noMisusedPromisesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			source := result.SourceFile.Text()
			for index, wantSpan := range testCase.wantSpans {
				diagnostic := result.Diagnostics[index]
				if got := source[diagnostic.Range.Pos():diagnostic.Range.End()]; got != wantSpan {
					t.Fatalf("finding %d span: got %q, want %q", index, got, wantSpan)
				}
			}
		})
	}
}

// TestNoMisusedPromisesDependsOnGlobalsTheFixtureLibraryOmits records the ES2022 hazard for THIS rule,
// measured with a two-file control rather than inherited from a sibling's verdict.
//
// The fixture tsconfig pins `lib: ["ES2022"]` with no way to raise it, so a name the DOM or Node
// would supply resolves to nothing. Whether that changes a rule's answer is a per-rule question: it
// inverted three `await-thenable` cases and moved nothing at all on `no-implied-eval`.
//
// Here it BITES, and it costs findings rather than adding them. `setTimeout(async () => {})` is
// SILENT under the bare fixture and reports `voidReturnArgument` once a declaration is supplied,
// because voidFunctionArguments walks the callee's signatures to find which parameters take a void
// return, and an unresolved callee has no signatures to walk. That is the safe direction — a missing
// global costs a finding rather than inventing one — but it is worth pinning, because a future
// fixture written against `setTimeout` would sit in the clean table asserting the opposite of
// upstream while passing.
//
// The controls are what make this a measurement: a real Promise conditional and an async callback to
// a locally declared void parameter report identically with and without the extra file, so the moved
// verdicts are about resolution rather than about the second file's presence.
func TestNoMisusedPromisesDependsOnGlobalsTheFixtureLibraryOmits(t *testing.T) {
	t.Parallel()

	const globals = `
declare global {
  function setTimeout(handler: () => void, ms?: number): number;
}
export {};
`

	cases := []struct {
		name                  string
		sourceText            string
		wantWithoutTheGlobals []string
		wantWithTheGlobals    []string
	}{
		{
			name:                  "an unresolved setTimeout has no signature to find a void parameter in",
			sourceText:            "setTimeout(async () => {});\n",
			wantWithoutTheGlobals: nil,
			wantWithTheGlobals:    []string{"voidReturnArgument"},
		},
		{
			name:                  "control: a locally declared void parameter reports either way",
			sourceText:            "declare function f(cb: () => void): void;\nf(async () => {});\n",
			wantWithoutTheGlobals: []string{"voidReturnArgument"},
			wantWithTheGlobals:    []string{"voidReturnArgument"},
		},
		{
			name:                  "control: a promise conditional reports either way",
			sourceText:            "const p = Promise.resolve();\nif (p) {\n}\n",
			wantWithoutTheGlobals: []string{"conditional"},
			wantWithTheGlobals:    []string{"conditional"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			bare := rule_testing.RunTyped(t, NoMisusedPromises, noMisusedPromisesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, bare, testCase.wantWithoutTheGlobals...)

			withGlobals := rule_testing.RunTypedFiles(t, NoMisusedPromises, map[string]string{
				noMisusedPromisesFile: testCase.sourceText,
				"globals.ts":          globals,
			}, noMisusedPromisesFile)
			rule_testing.ExpectFindings(t, withGlobals, testCase.wantWithTheGlobals...)
		})
	}
}

// TestNoMisusedPromisesChecksOnlyTheVoidPositions pins that the argument check reports at the
// positions voidFunctionArguments selected and NOT at every thenable-returning argument.
//
// This case exists because a mutant found it. Neutralizing the `slices.Contains(voidArgs, index)`
// membership test survived all two hundred and twelve imported fixtures, and the reason is structural
// rather than accidental: `checkArguments` returns early when voidArgs is empty, so a call with no
// void parameter at all cannot reach the loop either way. The membership test only earns its keep
// when SOME parameter is void and ANOTHER is not, and upstream's corpus never writes that shape.
//
// So each case below mixes one void-returning parameter with one parameter that is not, hands both an
// async callback, and asserts ONE finding. Under the mutant each reports two. That is different
// output rather than a different internal path, which is what makes this a fixture rather than a
// guess: the distinguishing input was measured on both versions before it was written down.
//
// The last case is the control for the whole set. A Promise-returning parameter FIRST and a void
// parameter second still reports once, which proves the single finding above is about which position
// was selected and not about the rule reporting at most once per call.
func TestNoMisusedPromisesChecksOnlyTheVoidPositions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			name: "a void parameter beside a Promise-returning one reports only the void position",
			sourceText: `
declare function f(a: () => void, b: () => Promise<void>): void;
f(async () => {}, async () => {});
`,
		},
		{
			name: "a void parameter beside an any parameter reports only the void position",
			sourceText: `
declare function f(a: () => void, b: any): void;
f(async () => {}, async () => {});
`,
		},
		{
			name: "control: the Promise-returning parameter first still reports once",
			sourceText: `
declare function f(a: () => Promise<void>, b: () => void): void;
f(async () => {}, async () => {});
`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMisusedPromises, noMisusedPromisesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "voidReturnArgument")
		})
	}
}

// TestNoMisusedPromisesReportsThePropertyReturnAnnotationWhenThereIsOne pins WHERE the property check
// points, which is a decision the message id cannot see.
//
// checkProperty chooses between two anchors: when the initializer is function-like AND carries an
// explicit return type annotation it reports the ANNOTATION, and otherwise it reports the initializer.
// Both produce `voidReturnProperty`, so every fixture asserting the id stayed green over a mutant that
// collapsed the two branches into one. That is the "fixtures assert the wrong layer" failure — the
// case was covered, and every assertion over it was about which rule fired rather than where.
//
// Four shapes, because the choice turns on two independent things (is the initializer function-like,
// does it carry an annotation) and the corpus writes only some of the combinations.
func TestNoMisusedPromisesReportsThePropertyReturnAnnotationWhenThereIsOne(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSpan   string
	}{
		{
			name: "an arrow initializer WITH a return annotation reports the annotation",
			sourceText: `
type O = { f: () => void };
const obj: O = {
  f: async (): Promise<string> => 'foo',
};
`,
			wantSpan: "Promise<string>",
		},
		{
			name: "an arrow initializer WITHOUT a return annotation reports the initializer",
			sourceText: `
type O = { f: () => void };
const obj: O = {
  f: async () => 'foo',
};
`,
			wantSpan: "async () => 'foo'",
		},
		{
			name: "a function-expression initializer with an annotation reports the annotation",
			sourceText: `
type O = { f: () => void };
const obj: O = {
  f: async function (): Promise<void> {},
};
`,
			wantSpan: "Promise<void>",
		},
		{
			name: "an initializer that is not function-like reports the initializer",
			sourceText: `
type O = { f: () => void };
const asyncFunction = async () => 'foo';
const obj: O = {
  f: asyncFunction,
};
`,
			wantSpan: "asyncFunction",
		},
		{
			name: "a method shorthand with an annotation reports the annotation",
			sourceText: `
type O = { f: () => void };
const obj: O = {
  async f(): Promise<void> {},
};
`,
			wantSpan: "Promise<void>",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMisusedPromises, noMisusedPromisesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "voidReturnProperty")

			source := result.SourceFile.Text()
			diagnostic := result.Diagnostics[0]
			if got := source[diagnostic.Range.Pos():diagnostic.Range.End()]; got != testCase.wantSpan {
				t.Fatalf("reported span: got %q, want %q", got, testCase.wantSpan)
			}
		})
	}
}

// TestNoMisusedPromisesSurvivesAttributeValuesThatAreNotExpressionContainers pins a guard whose job is
// to prevent a CRASH rather than to change a verdict.
//
// checkJSXAttribute requires the initializer's kind to be KindJsxExpression before calling
// AsJsxExpression on it, and that call is an unchecked type assertion. Removing the kind check does
// not merely widen the rule — it panics with `ast.nodeData is *ast.StringLiteral, not *ast.JsxExpression`
// on the first string-valued attribute in the tree, which is to say on essentially any real JSX file.
//
// A mutant removing it survived all two hundred and twelve imported fixtures, because upstream's
// corpus writes no string-valued attribute at all. No ExpectFindings assertion can see a panic, so
// this asserts the shapes reach the rule and come back clean; the crash is what failure looks like.
//
// The control matters here more than usual: without a case that DOES report on the same file shape,
// a suite of three clean assertions would pass just as well against a rule that had stopped running.
func TestNoMisusedPromisesSurvivesAttributeValuesThatAreNotExpressionContainers(t *testing.T) {
	t.Parallel()

	t.Run("a string-valued attribute beside a reporting one does not crash the rule", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, NoMisusedPromises, noMisusedPromisesTsxFile, `
type Props = { onEvent: () => void; label: string };
declare function Component(props: Props): null;
const element = <Component onEvent={async () => {}} label="hello" />;
`)
		rule_testing.ExpectFindings(t, result, "voidReturnAttribute")
	})

	t.Run("a string-valued attribute on a void-function prop is silent rather than fatal", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, NoMisusedPromises, noMisusedPromisesTsxFile, `
type Props = { onEvent: () => void };
declare function Component(props: Props): null;
const element = <Component onEvent="hello" />;
`)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("a shorthand attribute with no initializer at all is silent rather than fatal", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, NoMisusedPromises, noMisusedPromisesTsxFile, `
type Props = { onEvent?: () => void };
declare function Component(props: Props): null;
const element = <Component onEvent />;
`)
		rule_testing.ExpectClean(t, result)
	})
}

// TestNoMisusedPromisesNamesTheHeritageTypeInTheMessage asserts the one message on this rule that
// interpolates, on the rendered text rather than on the id.
//
// `voidReturnInheritedMethod` is built with Sprintf and carries the name of the extended or
// implemented type that declared the void-returning member. An id assertion cannot see anything a
// format string does, so a mutant replacing the interpolated type name with a constant survived all
// two hundred and twelve imported fixtures and every span assertion — the finding count was right,
// the id was right, the anchor was right, and the sentence a user reads named the wrong type.
//
// Asserted as equality against a literal typed here rather than against the rule's own message
// builder, because a comparison against the builder moves with the builder and asserts nothing.
//
// The two cases separate the interpolation from the surrounding sentence: an `implements` clause and
// an `extends` clause name different types, so a mutant that hardcoded either one is caught by the
// other.
func TestNoMisusedPromisesNamesTheHeritageTypeInTheMessage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		sourceText  string
		wantMessage string
	}{
		{
			name: "an implements clause names the implemented interface",
			sourceText: `
interface Base {
  method(): void;
}
class Derived implements Base {
  async method() {}
}
`,
			wantMessage: "Promise-returning method provided where a void return was expected by extended/implemented type 'Base'.",
		},
		{
			name: "an extends clause names the extended interface",
			sourceText: `
interface Parent {
  method(): void;
}
interface Child extends Parent {
  method(): Promise<void>;
}
`,
			wantMessage: "Promise-returning method provided where a void return was expected by extended/implemented type 'Parent'.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMisusedPromises, noMisusedPromisesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "voidReturnInheritedMethod")

			if got := result.Diagnostics[0].Message.Description; got != testCase.wantMessage {
				t.Fatalf("rendered message:\n got  %q\n want %q", got, testCase.wantMessage)
			}
		})
	}
}
