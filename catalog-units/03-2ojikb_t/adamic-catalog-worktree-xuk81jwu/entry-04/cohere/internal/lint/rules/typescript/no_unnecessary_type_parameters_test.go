package typescript

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus below is typescript-eslint's own, extracted by EXECUTING its tester file with a
// stubbed rule tester rather than by reading it, so every case carries the bytes upstream wrote.
// Every extracted string was then checked byte for byte against the tester source; the one case
// that does not appear verbatim is upstream's `noFormat` template literal, whose backticks and
// `${` are escaped in the source and correctly cooked by the extractor.
const noUnnecessaryTypeParametersFile = "/repository/source/TypeParameters.ts"

type noUnnecessaryTypeParametersCase struct {
	name   string
	source string
	ids    []string
}

var noUnnecessaryTypeParametersValidCases = []noUnnecessaryTypeParametersCase{
	{name: "valid0", source: `
class ClassyArray<T> {
  arr: T[];
}
    `},
	{name: "valid1", source: `
class ClassyArray<T> {
  value1: T;
  value2: T;
}
    `},
	{name: "valid2", source: `
class ClassyArray<T> {
  arr: T[];
  constructor(arr: T[]) {
    this.arr = arr;
  }
}
    `},
	{name: "valid3", source: `
class ClassyArray<T> {
  arr: T[];
  workWith(value: T) {
    this.arr.indexOf(value);
  }
}
    `},
	{name: "valid4", source: `
abstract class ClassyArray<T> {
  arr: T[];
  abstract workWith(value: T): void;
}
    `},
	{name: "valid5", source: `
class Box<T> {
  val: T | null = null;
  get() {
    return this.val;
  }
}
    `},
	{name: "valid6", source: `
class Joiner<T extends string | number> {
  join(els: T[]) {
    return els.map(el => '' + el).join(',');
  }
}
    `},
	{name: "valid7", source: `
declare class Foo {
  getProp<T>(this: Record<'prop', T>): T;
}
    `},
	{name: "valid8", source: `type Fn = <T>(input: T) => T;`},
	{name: "valid9", source: `type Fn = <T extends string>(input: T) => T;`},
	{name: "valid10", source: "type Fn = <T extends string>(input: T) => `a${T}b`;"},
	{name: "valid11", source: `type Fn = new <T>(input: T) => T;`},
	{name: "valid12", source: `type Fn = <T>(input: T) => typeof input;`},
	{name: "valid13", source: `type Fn = <T>(input: T) => keyof typeof input;`},
	{name: "valid14", source: `type Fn = <T>(input: Partial<T>) => typeof input;`},
	{name: "valid15", source: `type Fn = <T>(input: Partial<T>) => input is T;`},
	{name: "valid16", source: `type Fn = <T>(input: T) => { [K in keyof T]: K };`},
	{name: "valid17", source: `type Fn = <T>(input: T) => { [K in keyof T as K]: string };`},
	{name: "valid18", source: "type Fn = <T>(input: T) => { [K in keyof T as `${K & string}`]: string };"},
	{name: "valid19", source: `type Fn = <T>(input: T) => Partial<T>;`},
	{name: "valid20", source: `type Fn = <T>(input: { [i: number]: T }) => T;`},
	{name: "valid21", source: `type Fn = <T>(input: { [i: number]: T }) => Partial<T>;`},
	{name: "valid22", source: `type Fn = <T>(input: { [i: string]: T }) => Partial<T>;`},
	{name: "valid23", source: `type Fn = <T>(input: T) => { [i: number]: T };`},
	{name: "valid24", source: `type Fn = <T>(input: T) => { [i: string]: T };`},
	{name: "valid25", source: `type Fn = <T extends unknown[]>(input: T) => Omit<T, 'length'>;`},
	{name: "valid26", source: `
interface I {
  <T>(value: T): T;
}
    `},
	{name: "valid27", source: `
interface I {
  new <T>(value: T): T;
}
    `},
	{name: "valid28", source: `
function identity<T>(arg: T): T {
  return arg;
}
    `},
	{name: "valid29", source: `
function printProperty<T>(obj: T, key: keyof T) {
  console.log(obj[key]);
}
    `},
	{name: "valid30", source: `
function getProperty<T, K extends keyof T>(obj: T, key: K) {
  return obj[key];
}
    `},
	{name: "valid31", source: `
function box<T>(val: T) {
  return { val };
}
    `},
	{name: "valid32", source: `
function doStuff<K, V>(map: Map<K, V>, key: K) {
  let v = map.get(key);
  v = 1;
  map.set(key, v);
  return v;
}
    `},
	{name: "valid33", source: `
function makeMap<K, V>() {
  return new Map<K, V>();
}
    `},
	{name: "valid34", source: `
function makeMap<K, V>(ks: K[], vs: V[]) {
  const r = new Map<K, V>();
  ks.forEach((k, i) => {
    r.set(k, vs[i]);
  });
  return r;
}
    `},
	{name: "valid35", source: `
function arrayOfPairs<T>() {
  return [] as [T, T][];
}
    `},
	{name: "valid36", source: `
function isNonNull<T>(v: T): v is Exclude<T, null> {
  return v !== null;
}
    `},
	{name: "valid37", source: `
function both<Args extends unknown[]>(
  fn1: (...args: Args) => void,
  fn2: (...args: Args) => void,
): (...args: Args) => void {
  return function (...args: Args) {
    fn1(...args);
    fn2(...args);
  };
}
    `},
	{name: "valid38", source: `
function lengthyIdentity<T extends { length: number }>(x: T) {
  return x;
}
    `},
	{name: "valid39", source: `
interface Lengthy {
  length: number;
}
function lengthyIdentity<T extends Lengthy>(x: T) {
  return x;
}
    `},
	{name: "valid40", source: `
function ItemComponent<T>(props: { item: T; onSelect: (item: T) => void }) {}
    `},
	{name: "valid41", source: `
interface ItemProps<T> {
  item: readonly T;
  onSelect: (item: T) => void;
}
function ItemComponent<T>(props: ItemProps<T>) {}
    `},
	{name: "valid42", source: `
function useFocus<T extends HTMLOrSVGElement>(): [
  React.RefObject<T>,
  () => void,
];
    `},
	{name: "valid43", source: `
function findFirstResult<U>(
  inputs: unknown[],
  getResult: (t: unknown) => U | undefined,
): U | undefined;
    `},
	{name: "valid44", source: `
function findFirstResult<T, U>(
  inputs: T[],
  getResult: (t: T) => () => [U | undefined],
): () => [U | undefined];
    `},
	{name: "valid45", source: `
function getData<T>(url: string): Promise<T | null> {
  return Promise.resolve(null);
}
    `},
	{name: "valid46", source: `
function getData<T>(url: string): Promise<T extends null ? T : null> {
  return Promise.resolve(null);
}
    `},
	{name: "valid47", source: "\nfunction getData<T extends string>(url: string): Promise<`a${T}b`> {\n  return Promise.resolve(null);\n}\n    "},
	{name: "valid48", source: `
async function getData<T>(url: string): Promise<T | null> {
  return null;
}
    `},
	{name: "valid49", source: `declare function get(): void;`},
	{name: "valid50", source: `declare function get<T>(param: T[]): T;`},
	{name: "valid51", source: `declare function box<T>(val: T): { val: T };`},
	{name: "valid52", source: `declare function identity<T>(param: T): T;`},
	{name: "valid53", source: `declare function compare<T>(param1: T, param2: T): boolean;`},
	{name: "valid54", source: `declare function example<T>(a: Set<T>): T;`},
	{name: "valid55", source: `declare function example<T>(a: Set<T>, b: T[]): void;`},
	{name: "valid56", source: `declare function example<T>(a: Map<T, T>): void;`},
	{name: "valid57", source: `declare function example<T, U extends T>(t: T, u: U): U;`},
	{name: "valid58", source: `declare function makeSet<K>(): Set<K>;`},
	{name: "valid59", source: `declare function makeSet<K>(): [Set<K>];`},
	{name: "valid60", source: `declare function makeSets<K>(): Set<K>[];`},
	{name: "valid61", source: `declare function makeSets<K>(): [Set<K>][];`},
	{name: "valid62", source: `declare function makeMap<K, V>(): Map<K, V>;`},
	{name: "valid63", source: `declare function makeMap<K, V>(): [Map<K, V>];`},
	{name: "valid64", source: `declare function makeArray<T>(): T[];`},
	{name: "valid65", source: `declare function makeArrayNullish<T>(): (T | null)[];`},
	{name: "valid66", source: `declare function makeTupleMulti<T>(): [T | null, T | null];`},
	{name: "valid67", source: `declare function takeTupleMulti<T>(input: [T, T]): void;`},
	{name: "valid68", source: `declare function takeTupleMultiNullish<T>(input: [T | null, T | null]): void;`},
	{name: "valid69", source: `declare function arrayOfPairs<T>(): [T, T][];`},
	{name: "valid70", source: `declare function fetchJson<T>(url: string): Promise<T>;`},
	{name: "valid71", source: `declare function fetchJsonTuple<T>(url: string): Promise<[T]>;`},
	{name: "valid72", source: `declare function fn<T>(input: T): 0 extends 0 ? T : never;`},
	{name: "valid73", source: `declare function useFocus<T extends HTMLOrSVGElement>(): [React.RefObject<T>];`},
	{name: "valid74", source: `
declare function useFocus<T extends HTMLOrSVGElement>(): {
  ref: React.RefObject<T>;
};
    `},
	{name: "valid75", source: `
interface TwoMethods<T> {
  a(x: T): void;
  b(x: T): void;
}

declare function two<T>(props: TwoMethods<T>): void;
    `},
	{name: "valid76", source: `
type Obj = { a: string };

declare function hasOwnProperty<K extends keyof Obj>(
  obj: Obj,
  key: K,
): obj is Obj & { [key in K]-?: Obj[key] };
    `},
	{name: "valid77", source: `
type AsMutable<T extends readonly unknown[]> = {
  -readonly [Key in keyof T]: T[Key];
};

declare function makeMutable<T>(input: T): MakeMutable<T>;
    `},
	{name: "valid78", source: `
type AsMutable<T extends readonly unknown[]> = {
  -readonly [Key in keyof T]: T[Key];
};

declare function makeMutable<T>(input: T): MakeMutable<typeof input>;
    `},
	{name: "valid79", source: `
type ValueNulls<U extends string> = {} & {
  [P in U]: null;
};

declare function invert<T extends string>(obj: T): ValueNulls<T>;
    `},
	{name: "valid80", source: `
interface Middle {
  inner: boolean;
}

type Conditional<T extends Middle> = {} & (T['inner'] extends true ? {} : {});

function withMiddle<T extends Middle = Middle>(options: T): Conditional<T> {
  return options;
}
    `},
	{name: "valid81", source: `

declare function forEachReturnStatement<T>(
  body: ts.Block,
  visitor: (stmt: ts.ReturnStatement) => T,
): T | undefined;
    `},
	{name: "valid82", source: `

declare const isNodeOfType: <NodeType extends AST_NODE_TYPES>(
  nodeType: NodeType,
) => node is Extract<TSESTree.Node, { type: NodeType }>;
    `},
	{name: "valid84", source: `

export const isNotTokenOfTypeWithConditions =
  <
    TokenType extends AST_TOKEN_TYPES,
    ExtractedToken extends Extract<TSESTree.Token, { type: TokenType }>,
    Conditions extends Partial<ExtractedToken>,
  >(
    tokenType: TokenType,
    conditions: Conditions,
  ): ((
    token: TSESTree.Token | null | undefined,
  ) => token is Exclude<TSESTree.Token, Conditions & ExtractedToken>) =>
  (token): token is Exclude<TSESTree.Token, Conditions & ExtractedToken> =>
    tokenType in conditions;
    `},
	{name: "valid85", source: `
type Foo<T, S> = S extends 'somebody'
  ? T extends 'once'
    ? 'told'
    : 'me'
  : never;

declare function foo<T>(data: T): <S>(other: S) => Foo<T, S>;
    `},
	{name: "valid86", source: `
type Foo<T, S> = S extends 'somebody'
  ? T extends 'once'
    ? 'told'
    : 'me'
  : never;

declare function foo<T>(data: T): <S>(other: S) => Foo<S, T>;
    `},
	{name: "valid87", source: `
declare function mapObj<K extends string, V>(
  obj: { [key in K]?: V },
  fn: (key: K, val: V) => number,
): number[];
    `},
	{name: "valid91", source: `
type Identity<T> = T;

type Mapped<T, Value> = Identity<{ [P in keyof T]: Value }>;

declare function sillyFoo<Data, Value>(
  c: Value,
): (data: Data) => Mapped<Data, Value>;
    `},
	{name: "valid92", source: `
type Silly<T> = { [P in keyof T]: T[P] };

type SillyFoo<T, Value> = Silly<{ [P in keyof T]: Value }>;

type Foo<T, Value> = { [P in keyof T]: Value };

declare function foo<T, Constant>(data: T, c: Constant): Foo<T, Constant>;
declare function foo<T, Constant>(c: Constant): (data: T) => Foo<T, Constant>;

declare function sillyFoo<T, Constant>(
  data: T,
  c: Constant,
): SillyFoo<T, Constant>;
declare function sillyFoo<T, Constant>(
  c: Constant,
): (data: T) => SillyFoo<T, Constant>;
    `},
	{name: "valid93", source: `
const f = <T,>(setValue: (v: T) => void, getValue: () => NoInfer<T>) => {};
    `},
	{name: "valid94", source: `
const f = <T,>(
  setValue: (v: T) => NoInfer<T>,
  getValue: (v: NoInfer<T>) => NoInfer<T>,
) => {};
    `},
}

var noUnnecessaryTypeParametersInvalidCases = []noUnnecessaryTypeParametersCase{
	{name: "invalid0", source: `const func = <T,>(param: T) => null;`, ids: []string{"sole"}},
	{name: "invalid1", source: `const func = <T,>(param: [T]) => null;`, ids: []string{"sole"}},
	{name: "invalid2", source: `const func = <T,>(param: T[]) => null;`, ids: []string{"sole"}},
	{name: "invalid4", source: `
interface I {
  <T>(value: T): void;
}
      `, ids: []string{"sole"}},
	{name: "invalid5", source: `
interface I {
  m<T>(x: T): void;
}
      `, ids: []string{"sole"}},
	{name: "invalid6", source: `
class Joiner<T extends string | number> {
  join(el: T, other: string) {
    return [el, other].join(',');
  }
}
      `, ids: []string{"sole"}},
	{name: "invalid7", source: `
declare class C<V> {}
      `, ids: []string{"sole"}},
	{name: "invalid8", source: `
declare class C<T, U> {
  method(param: T): U;
}
      `, ids: []string{"sole", "sole"}},
	{name: "invalid9", source: `
declare class C {
  method<T, U>(param: T): U;
}
      `, ids: []string{"sole", "sole"}},
	{name: "invalid11", source: `
declare class Foo {
  foo<T>(this: T): void;
}
      `, ids: []string{"sole"}},
	{name: "invalid12", source: `
function third<A, B, C>(a: A, b: B, c: C): C {
  return c;
}
      `, ids: []string{"sole", "sole"}},
	{name: "invalid13", source: `
function foo<T>(_: T) {
  const x: T = null!;
  const y: T = null!;
}
      `, ids: []string{"sole"}},
	{name: "invalid14", source: `
function foo<T>(_: T): void {
  const x: T = null!;
  const y: T = null!;
}
      `, ids: []string{"sole"}},
	{name: "invalid15", source: `
function foo<T>(_: T): <T>(input: T) => T {
  const x: T = null!;
  const y: T = null!;
  return null!;
}
      `, ids: []string{"sole"}},
	{name: "invalid16", source: `
function foo<T>(_: T) {
  function withX(): T {
    return null!;
  }
  function withY(): T {
    return null!;
  }
}
      `, ids: []string{"sole"}},
	{name: "invalid17", source: `
function parseYAML<T>(input: string): T {
  return input as any as T;
}
      `, ids: []string{"sole"}},
	{name: "invalid18", source: `
function printProperty<T, K extends keyof T>(obj: T, key: K) {
  console.log(obj[key]);
}
      `, ids: []string{"sole"}},
	{name: "invalid19", source: `
function fn<T>(param: string) {
  let v: T = null!;
  return v;
}
      `, ids: []string{"sole"}},
	{name: "invalid20", source: `
function both<
  Args extends unknown[],
  CB1 extends (...args: Args) => void,
  CB2 extends (...args: Args) => void,
>(fn1: CB1, fn2: CB2): (...args: Args) => void {
  return function (...args: Args) {
    fn1(...args);
    fn2(...args);
  };
}
      `, ids: []string{"sole", "sole"}},
	{name: "invalid21", source: `
function getLength<T extends { length: number }>(x: T) {
  return x.length;
}
      `, ids: []string{"sole"}},
	{name: "invalid22", source: `
interface Lengthy {
  length: number;
}
function getLength<T extends Lengthy>(x: T) {
  return x.length;
}
      `, ids: []string{"sole"}},
	{name: "invalid23", source: `declare function get<T>(): unknown;`, ids: []string{"sole"}},
	{name: "invalid26", source: `declare function take<T>(param: T): void;`, ids: []string{"sole"}},
	{name: "invalid27", source: `declare function take<T extends object>(param: T): void;`, ids: []string{"sole"}},
	{name: "invalid28", source: `declare function take<T, U = T>(param1: T, param2: U): void;`, ids: []string{"sole"}},
	{name: "invalid29", source: `declare function take<T, U extends T>(param: T): U;`, ids: []string{"sole"}},
	{name: "invalid30", source: `declare function take<T, U extends T>(param: U): U;`, ids: []string{"sole"}},
	{name: "invalid31", source: `declare function get<T, U = T>(param: U): U;`, ids: []string{"sole"}},
	{name: "invalid32", source: `declare function get<T, U extends T = T>(param: T): U;`, ids: []string{"sole"}},
	{name: "invalid33", source: `declare function compare<T, U extends T>(param1: T, param2: U): boolean;`, ids: []string{"sole"}},
	{name: "invalid34", source: `declare function get<T>(param: <U, V>(param: U) => V): T;`, ids: []string{"sole", "sole", "sole"}},
	{name: "invalid35", source: `declare function get<T>(param: <T, U>(param: T) => U): T;`, ids: []string{"sole", "sole", "sole"}},
	{name: "invalid39", source: `declare function takeArray<T>(input: T[]): void;`, ids: []string{"sole"}},
	{name: "invalid40", source: `declare function takeArrayNullish<T>(input: (T | null)[]): void;`, ids: []string{"sole"}},
	{name: "invalid41", source: `declare function takeTuple<T>(input: [T]): void;`, ids: []string{"sole"}},
	{name: "invalid42", source: `declare function takeTupleMultiUnrelated<T>(input: [T, number]): void;`, ids: []string{"sole"}},
	{name: "invalid43", source: `
declare function takeTupleMultiUnrelatedNullish<T>(
  input: [T | null, null],
): void;
      `, ids: []string{"sole"}},
	{name: "invalid45", source: `type Fn = <T>() => [];`, ids: []string{"sole"}},
	{name: "invalid46", source: `
type Other = 0;
type Fn = <T>() => Other;
      `, ids: []string{"sole"}},
	{name: "invalid47", source: `
type Other = 0 | 1;
type Fn = <T>() => Other;
      `, ids: []string{"sole"}},
	{name: "invalid48", source: `type Fn = <U>(param: U) => void;`, ids: []string{"sole"}},
	{name: "invalid52", source: `type Fn = <T>(value: unknown) => value is T;`, ids: []string{"sole"}},
	{name: "invalid54", source: `
declare function mapObj<K extends string, V>(
  obj: { [key in K]?: V },
  fn: (key: K) => number,
): number[];
      `, ids: []string{"sole"}},
	{name: "invalid55", source: `
declare function setItem<T>(T): T;
      `, ids: []string{"sole"}},
	{name: "invalid56", source: `
interface StorageService {
  setItem<T>({ key: string, value: T }): Promise<void>;
}
      `, ids: []string{"sole"}},
	// Upstream's invalid57, the `Equal<X, Y>` exact-equality idiom, reports twice upstream and is
	// deliberately silent here. It lives in TestNoUnnecessaryTypeParametersRecognizesTheExactEqualityIdiom,
	// verbatim, with the reasoning.
	{name: "invalid58", source: `
function f<T extends any>(x: T): void {
  // @ts-expect-error
  x.notAMethod();
}
      `, ids: []string{"sole"}},
	{name: "invalid59", source: `
class Joiner {
  join<T extends number>(els: T[]) {
    return els.map(el => '' + el).join(',');
  }
}
      `, ids: []string{"sole"}},
	{name: "invalid60", source: `
function join<T extends string | number>(els: T[]) {
  return els.map(el => '' + el).join(',');
}
      `, ids: []string{"sole"}},
	{name: "invalid61", source: `
function join<T extends string & number>(els: T[]) {
  return els.map(el => '' + el).join(',');
}
      `, ids: []string{"sole"}},
	{name: "invalid62", source: `
function join<T extends (string & number) | boolean>(els: T[]) {
  return els.map(el => '' + el).join(',');
}
      `, ids: []string{"sole"}},
	{name: "invalid63", source: `
function join<T extends (string | number)>(els: T[]) {
  return els.map(el => '' + el).join(',');
}
      `, ids: []string{"sole"}},
	{name: "invalid64", source: `
function join<T extends { hoge: string } | { hoge: number }>(els: T['hoge'][]) {
  return els.map(el => '' + el).join(',');
}
      `, ids: []string{"sole"}},
}

// TestNoUnnecessaryTypeParametersStaysSilent runs upstream's valid list.
//
// These are the false positives upstream already thought about, and they are the half of the corpus
// that catches a port over-reporting. The mapped-type rows among them are the ones this port is
// most exposed on, because the walk cannot descend into a mapped type here; each is marked in the
// divergence test below rather than deleted.
func TestNoUnnecessaryTypeParametersStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range noUnnecessaryTypeParametersValidCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUnnecessaryTypeParametersFires runs upstream's invalid list, asserting one message id per
// expected finding so a case reporting twice is distinguished from a case reporting once.
func TestNoUnnecessaryTypeParametersFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range noUnnecessaryTypeParametersInvalidCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

// TestNoUnnecessaryTypeParametersMappedTypeDivergence records the one place this port knowingly
// disagrees with upstream, so the difference is a measured fact in the suite rather than four cases
// quietly deleted from the corpus.
//
// Upstream descends into a mapped type by reading four fields off it: its own type parameter, its
// constraint type, its name type and its template type. On checker.MappedType those four fields are
// unexported, carry no accessor method, and have no entry in shim/checker/extra-shim.json, so they
// cannot be reached without a shim change. Measured with a compiling probe rather than a grep:
// reaching MappedType.TypeParameter() fails to build while the same shape on
// ConditionalType.CheckType() and IndexedAccessType.ObjectType() compiles, which is what separates
// a real absence from a name nobody guessed.
//
// The walk therefore stops at a mapped type instead of counting the uses inside it. That can only
// LOWER a count, so the exposure is a false positive on a clean shape rather than a missed finding,
// which is the safe direction for a rule whose repair is a suggestion a human must choose.
//
// The blast radius is four cases. Twelve of upstream's sixteen mapped-type cases pass here already,
// including all four of its REPORTING ones, because the uses that decide them are visible before
// the walk reaches the mapped type. The four below need what is inside it.
func TestNoUnnecessaryTypeParametersMappedTypeDivergence(t *testing.T) {
	t.Parallel()

	cases := []noUnnecessaryTypeParametersCase{
		{name: "valid88", source: `
declare function mappedReturnType<T extends string>(
  x: T,
): { [K in T]: Capitalize<K> };

function inferredMappedReturnType<T extends string>(x: T) {
  return mappedReturnType(x);
}
    `},
		{name: "valid89", source: `
declare function mappedReturnType<T extends string>(
  x: T,
): { [K in T]: Capitalize<K> };

function inferredMappedReturnType<T extends string>(x: T) {
  return () => mappedReturnType(x);
}
    `},
		{name: "valid90", source: `
declare function mappedReturnType<T extends string>(
  x: T,
): { [K in T]: Capitalize<K> };

function inferredMappedReturnType<T extends string>(x: T) {
  return [{ value: () => mappedReturnType(x) }];
}
    `},
		{name: "valid95", source: `<T extends string>(t: T) => t as { [K in 'a' as T]: 0 };`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			// Upstream is SILENT on each of these. This port reports once. Asserting the divergence
			// rather than the upstream verdict means closing the gap makes this test fail loudly,
			// which is the point: a divergence nobody is told about becomes permanent.
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected the documented divergence of exactly 1 finding, got %d", len(result.Diagnostics))
			}
		})
	}
}

// TestNoUnnecessaryTypeParametersUnresolvedImportIsAHarnessLimit pins the one upstream clean case
// this harness cannot express, so it is not mistaken for a rule defect.
//
// Upstream's case imports AST_NODE_TYPES and TSESTree from '@typescript-eslint/types', which the
// rule_testing tsconfig cannot resolve, so both names become error types and the walk cannot count
// through them. The control below is the SAME shape with those two names declared locally: it is
// clean, which is upstream's verdict, and that is what establishes the rule is right and the
// instrument is what differs.
func TestNoUnnecessaryTypeParametersUnresolvedImportIsAHarnessLimit(t *testing.T) {
	t.Parallel()

	withUnresolvableImport := `

const isNodeOfType =
  <NodeType extends AST_NODE_TYPES>(nodeType: NodeType) =>
  (
    node: TSESTree.Node | null,
  ): node is Extract<TSESTree.Node, { type: NodeType }> =>
    node?.type === nodeType;
    `

	// Same shape, nothing unresolved.
	withLocalDeclarations := `
type AST_NODE_TYPES = 'a' | 'b';
declare namespace TSESTree {
  type Node = { type: AST_NODE_TYPES };
}

const isNodeOfType =
  <NodeType extends AST_NODE_TYPES>(nodeType: NodeType) =>
  (
    node: TSESTree.Node | null,
  ): node is Extract<TSESTree.Node, { type: NodeType }> =>
    node?.type === nodeType;
`

	unresolved := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, withUnresolvableImport)
	if len(unresolved.Diagnostics) != 1 {
		t.Fatalf("expected the documented harness divergence of exactly 1 finding, got %d", len(unresolved.Diagnostics))
	}

	// The control. If this ever reports, the cause is the rule rather than the unresolved import,
	// and the comment above is wrong.
	resolved := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, withLocalDeclarations)
	rule_testing.ExpectClean(t, resolved)
}

// TestNoUnnecessaryTypeParametersReportsAtTheTypeParameter asserts WHERE each finding points and
// WHAT its message says, neither of which ExpectFindings can see.
//
// The span matters here because the natural wrong answer is defensible: a reader could reasonably
// expect the finding on the declaration that owns the parameter, or on the parameter's name rather
// than the whole parameter. Upstream reports the esTypeParameter node, which is the parameter
// INCLUDING its constraint, so `K extends keyof T` reports across the whole of `K extends keyof T`
// and not across `K`. An id-only fixture is green over either choice.
//
// The message text is asserted by EQUALITY on a literal typed here, not against the rule's own
// message constant: comparing a finding to the constant it was built from moves both sides together
// under mutation and proves nothing.
func TestNoUnnecessaryTypeParametersReportsAtTheTypeParameter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		source      string
		wantSpans   []string
		wantMessage string
	}{
		{
			name:        "the whole parameter including its constraint",
			source:      "function printProperty<T, K extends keyof T>(obj: T, key: K) {\n  console.log(obj[key]);\n}\n",
			wantSpans:   []string{"K extends keyof T"},
			wantMessage: "Type parameter K is used only once in the function signature.",
		},
		{
			name:        "a parameter that is never used at all",
			source:      "declare class C<V> {}\n",
			wantSpans:   []string{"V"},
			wantMessage: "Type parameter V is never used in the class signature.",
		},
		{
			name:      "two findings report at their own parameters",
			source:    "declare class C<T, U> {\n  method(param: T): U;\n}\n",
			wantSpans: []string{"T", "U"},
			// The message is asserted on the first finding only; the second is covered by the
			// single-finding rows above.
			wantMessage: "Type parameter T is used only once in the class signature.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("expected %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}

			// Slice the source the HARNESS wrote rather than the literal above: RunTyped trims the
			// fixture and appends a newline, so a span sliced from the Go literal is off by one on
			// any case carrying a leading newline.
			written := strings.TrimSpace(testCase.source) + "\n"
			for index, wantSpan := range testCase.wantSpans {
				diagnostic := result.Diagnostics[index]
				gotSpan := written[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d: expected span %q, got %q", index, wantSpan, gotSpan)
				}
			}

			if got := result.Diagnostics[0].Message.Description; got != testCase.wantMessage {
				t.Errorf("expected message %q, got %q", testCase.wantMessage, got)
			}
		})
	}
}

// TestNoUnnecessaryTypeParametersOffersTheConstraintSuggestion asserts the repair SURFACE.
//
// Upstream ships `replaceUsagesWithConstraint` as a SUGGESTION rather than a fix, and that
// distinction is part of what is being ported: rewriting every use of a parameter to its constraint
// and deleting the parameter changes what the signature means to callers, so a human chooses it and
// the engine never applies it unattended. A rule that shipped this as a fix would pass every
// message-id fixture while rewriting code upstream would only have offered to rewrite.
func TestNoUnnecessaryTypeParametersOffersTheConstraintSuggestion(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile,
		"declare class C<V> {}\n")

	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Diagnostics))
	}
	diagnostic := result.Diagnostics[0]

	if len(diagnostic.Fixes) != 0 {
		t.Errorf("expected no unattended fixes, got %d", len(diagnostic.Fixes))
	}
	if len(diagnostic.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(diagnostic.Suggestions))
	}
	if got := diagnostic.Suggestions[0].Message.Id; got != "replaceUsagesWithConstraint" {
		t.Errorf("expected suggestion id replaceUsagesWithConstraint, got %q", got)
	}
	if got := diagnostic.Suggestions[0].Message.Description; got != "Replace all usages of type parameter with its constraint." {
		t.Errorf("unexpected suggestion text %q", got)
	}
}

// TestNoUnnecessaryTypeParametersDeclinesWithoutATypeChecker pins the nil guard in Run.
//
// NeedsTypeChecker governs the REGISTRATION path only; the harness builds a rule.Context by hand,
// which is how a typed rule panics with a nil receiver while its declaration reads as correct. The
// assertion is here so a later revert fails loudly rather than going vacuously green.
func TestNoUnnecessaryTypeParametersDeclinesWithoutATypeChecker(t *testing.T) {
	t.Parallel()

	if !NoUnnecessaryTypeParameters.NeedsTypeChecker {
		t.Fatal("the rule reads the checker, so it must declare NeedsTypeChecker")
	}

	// The untyped harness hands the rule a nil checker. Run must decline rather than panic.
	result := rule_testing.Run(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile,
		"declare class C<V> {}\n")
	rule_testing.ExpectClean(t, result)
}

// TestNoUnnecessaryTypeParametersArrayCarveOut covers the one discrimination upstream's corpus
// leaves untested, found by a mutation that survived all 158 imported cases.
//
// The syntactic fast path treats a type parameter passed as a type ARGUMENT as repeated on sight,
// because it cannot see inside the alias to count the uses. Array and ReadonlyArray are carved out
// of that and deferred to the type phase, where `T[]` is correctly counted as a single use.
//
// Upstream's corpus writes the carve-out only in its shorthand form, `T[]`, which the fast path
// never classifies as a type argument in the first place. So the carve-out is dead against every
// imported case, and deleting it leaves all 158 green. The spelled-out forms are what reach it, and
// they are exactly the shapes upstream had no reason to write.
//
// Every verdict below was measured against the installed 8.67.0 build rather than reasoned about,
// with a harness that refuses to report unless a firing control fires and a clean control stays
// clean.
func TestNoUnnecessaryTypeParametersArrayCarveOut(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		source       string
		wantFindings int
	}{
		// The carve-out's own subjects. Without it these go silent while upstream reports.
		{name: "Array spelled out", source: "declare function e<T>(param: Array<T>): void;\n", wantFindings: 1},
		{name: "ReadonlyArray spelled out", source: "declare function f<T>(param: ReadonlyArray<T>): void;\n", wantFindings: 1},

		// The shorthand forms, which upstream's corpus does cover, kept here so the pair reads
		// together and a future edit cannot move one without the other being visible.
		{name: "array shorthand", source: "declare function d<T>(param: T[]): void;\n", wantFindings: 1},
		{name: "readonly array shorthand", source: "declare function g<T>(param: readonly T[]): void;\n", wantFindings: 1},

		// The other side of the same branch: a generic that is NOT an array stays exempt, so these
		// are silent. Without them a mutation deleting the whole type-argument test would be
		// indistinguishable from one deleting only the carve-out.
		{name: "Partial is exempt", source: "declare function a<T>(param: Partial<T>): void;\n", wantFindings: 0},
		{name: "Set is exempt", source: "declare function b<T>(param: Set<T>): void;\n", wantFindings: 0},
		{name: "Map is exempt", source: "declare function c<T>(param: Map<T, string>): void;\n", wantFindings: 0},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			if len(result.Diagnostics) != testCase.wantFindings {
				t.Fatalf("expected %d findings, got %d", testCase.wantFindings, len(result.Diagnostics))
			}
		})
	}
}

// TestNoUnnecessaryTypeParametersConstraintRecursion covers the constraint recursion inside the
// type walk, found by a mutation that survived all 158 imported cases plus every fixture above.
//
// When the walk arrives at a type parameter it also visits that parameter's CONSTRAINT, so a
// parameter mentioned only in a sibling's constraint still accumulates a count. Deleting the
// recursion leaves every verdict unchanged and changes the MESSAGE, from "used only once" to
// "never used", because the count falls from 2 to 1 while staying under the report threshold.
//
// That is why this test asserts the rendered text rather than the count. A message-id fixture and a
// findings-count fixture are both satisfied by a finding that says the wrong thing about the code,
// and this is the only shape in the corpus where the difference shows.
//
// Every expectation below was measured against the installed 8.67.0 build, through a harness that
// refuses to report unless a firing control fires and a clean control stays clean.
func TestNoUnnecessaryTypeParametersConstraintRecursion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		source       string
		wantMessages []string
	}{
		{
			name:   "a parameter used only in a sibling constraint is used once, not never",
			source: "declare function a<T, U extends T>(param: U): void;\n",
			wantMessages: []string{
				"Type parameter T is used only once in the function signature.",
				"Type parameter U is used only once in the function signature.",
			},
		},
		{
			name:         "the same with the sibling also returned",
			source:       "declare function b<T, U extends T>(param: U): U;\n",
			wantMessages: []string{"Type parameter T is used only once in the function signature."},
		},
		{
			name:   "the constraint reaches through an array",
			source: "declare function c<T, U extends T[]>(param: U): void;\n",
			wantMessages: []string{
				"Type parameter T is used only once in the function signature.",
				"Type parameter U is used only once in the function signature.",
			},
		},
		{
			name:         "a constraint naming no type parameter is unaffected",
			source:       "declare function d<T extends string>(param: T): void;\n",
			wantMessages: []string{"Type parameter T is used only once in the function signature."},
		},
		{
			name:         "the constraint reaches through a generic",
			source:       "declare function e<T, U extends Set<T>>(param: U): void;\n",
			wantMessages: []string{"Type parameter U is used only once in the function signature."},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			if len(result.Diagnostics) != len(testCase.wantMessages) {
				t.Fatalf("expected %d findings, got %d", len(testCase.wantMessages), len(result.Diagnostics))
			}
			for index, wantMessage := range testCase.wantMessages {
				if got := result.Diagnostics[index].Message.Description; got != wantMessage {
					t.Errorf("finding %d: expected %q, got %q", index, wantMessage, got)
				}
			}
		})
	}
}

// TestNoUnnecessaryTypeParametersIndexSignaturesCountTwice covers the two index-signature visits in
// the object arm, each found by a mutation that survived every fixture above.
//
// A type reached through an index signature is counted as TWO uses rather than one, because the
// walk cannot tell how many values the index stands for. Counting it once instead makes each of
// these shapes report while upstream is silent, which is a false positive rather than a missed
// finding, and no imported case covers it.
//
// Upstream's corpus does carry index-signature cases, and they pass either way: they place the
// parameter somewhere else in the signature as well, so the count clears the threshold without help
// from this arm. The shapes below put the parameter ONLY behind the index, which is what separates
// the two versions.
//
// The number-index and string-index visits are separate lines and were mutated separately, because
// a mutation of one cannot see a defect in the other.
func TestNoUnnecessaryTypeParametersIndexSignaturesCountTwice(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{name: "number index in parameter position", source: "declare function f<T>(param: { [key: number]: T }): void;\n"},
		{name: "string index in parameter position", source: "declare function g<T>(param: { [key: string]: T }): void;\n"},
		{name: "number index in return position", source: "declare function h<T>(param: string): { [key: number]: T };\n"},
		{name: "string index in return position", source: "declare function i<T>(param: string): { [key: string]: T };\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUnnecessaryTypeParametersRecognizesTheExactEqualityIdiom is where cohere is deliberately
// quieter than ESLint.
//
// `(<T>() => T extends L ? 1 : 2) extends <T>() => T extends R ? 1 : 2` asks whether L and R are
// IDENTICAL, not merely mutually assignable: the checker defers a conditional type whose check type
// is an unresolved type parameter, and relates two deferred conditionals only when their extends
// types are identical. `T` used once is the whole point, and the rule's suggestion, replacing it
// with its constraint, makes both conditionals resolve eagerly and destroys the comparison. Neither
// function type is ever the type of a value, so no cast can hide in it.
//
// The first silent row is the real site, verbatim: `nexus/source/types/UnionFromClasses.test.ts:23`
// in ahra, two findings in ESLint and in cohere before this. The second is upstream's own invalid57,
// which typescript-eslint pins as reporting twice; it is the same idiom and moved here from the
// reporting table.
//
// The idiom is silenced by the type-witness principle (isTypeWitnessParameter), not by a recognizer
// of its own, so the five rows after the first four, once controls that reported, are silent for the
// same reason. The one reporting row is what still separates the idiom from a cast: the conditional
// reached through a value parameter rather than the return.
func TestNoUnnecessaryTypeParametersRecognizesTheExactEqualityIdiom(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name   string
		source string
	}{
		{name: "UnionFromClasses.test.ts:23", source: "type IsExactlyType<TLeft, TRight> =\n    (<T>() => T extends TLeft ? 1 : 2) extends <T>() => T extends TRight ? 1 : 2 ? true : false;\n"},
		{name: "upstream invalid57, the same idiom", source: `
type Compute<A> = A extends Function ? A : { [K in keyof A]: Compute<A[K]> };
type Equal<X, Y> =
  (<T1>() => T1 extends Compute<X> ? 1 : 2) extends
    (<T2>() => T2 extends Compute<Y> ? 1 : 2)
  ? true
  : false;
      `},
		{name: "parenthesized check type inside the witness", source: "type Same<L, R> = (<T>() => (T) extends L ? 1 : 2) extends (<T>() => T extends R ? 1 : 2) ? true : false;\n"},
		{name: "the constructor-type spelling of the idiom", source: "type Same<L, R> = (new <T>() => T extends L ? 1 : 2) extends (new <T>() => T extends R ? 1 : 2) ? true : false;\n"},
		// These five reported while the idiom had its own recognizer, as controls for how narrowly it
		// matched the comparison. Each has no value parameter and uses `T` only in its return, so the
		// witness principle that replaced the recognizer is silent on all five.
		{name: "the witness shape outside a comparison", source: "type Witness = <T>() => T extends string ? 1 : 2;\n"},
		{name: "an operand whose return is not a conditional", source: "type Same<L, R> = (<T>() => T) extends (<T>() => T) ? true : false;\n"},
		{name: "an operand whose conditional checks another type", source: "type Same<L, R> = (<T>() => L extends T ? 1 : 2) extends (<T>() => R extends T ? 1 : 2) ? true : false;\n"},
		{name: "a declaration returning the conditional", source: "declare function witness<T>(): T extends string ? 1 : 2;\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}

	reporting := []struct {
		name   string
		source string
		ids    []string
	}{
		{name: "the conditional in a parameter of an operand", source: "type Same<L> = (<T>(input: T extends L ? 1 : 2) => void) extends (() => void) ? true : false;\n", ids: []string{"sole"}},
		// A method signature has a receiver to cast from, so the witness principle no longer silences it
		// (#gtgw3av), and it reports as it did when the idiom had its own recognizer and as upstream does.
		{name: "a method signature inside a compared type literal", source: "type Same<L> = { m<T>(): T extends L ? 1 : 2 } extends {} ? true : false;\n", ids: []string{"sole"}},
	}
	for _, testCase := range reporting {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

// TestNoUnnecessaryTypeParametersTypeWitnesses pins the type-witness principle (isTypeWitnessParameter):
// a function with no value parameters whose type parameter appears only in its return type cannot be
// a hidden cast, because there is nothing to cast from. Kirk's ruling of 2026-10-01.
//
// typescript-eslint reports every silent row. The first is ahra's own site; the rest are upstream's
// invalid cases of exactly this shape, moved here verbatim from the invalid table (fourteen of its 66),
// so the corpus still holds them and closing the divergence makes this test fail rather than
// silently changing a count. `declare function get<T>(): T` (invalid24) is upstream's canonical
// return-only generic, and it is silent here by the same principle as `typeOnly`: syntax cannot tell
// the two apart, which is why the ruling is a principle rather than a name.
func TestNoUnnecessaryTypeParametersTypeWitnesses(t *testing.T) {
	t.Parallel()

	silent := []noUnnecessaryTypeParametersCase{
		{name: "ObjectTypes.ts:93, typeOnly", source: "export function typeOnly<Shape>(): Shape {\n    return null as unknown as Shape;\n}\n"},
		{name: "invalid3", source: `const f1 = <T,>(): T => {};`},
		{name: "invalid10", source: `
declare class C {
  prop: <P>() => P;
}
      `},
		// A witness body hands back nothing, however it spells it (Kirk's ruling of 2026-10-02).
		{name: "typeOnly returning undefined", source: "export function typeOnly<Shape>(): Shape {\n    return undefined as unknown as Shape;\n}\n"},
		{name: "typeOnly returning void 0 through a non-null assertion", source: "export function typeOnly<Shape>(): Shape {\n    return (void 0)!;\n}\n"},
		{name: "an arrow whose expression body is null", source: "export const typeOnly = <Shape,>(): Shape => null as unknown as Shape;\n"},
		{name: "a body that only throws", source: "export function unreachable<Shape>(): Shape {\n    throw new Error('never called');\n}\n"},
		{name: "a nested function's value return is not the witness's", source: "export function typeOnly<Shape>(): Shape {\n    const unused = () => 1;\n    unused();\n    return null as unknown as Shape;\n}\n"},
		{name: "invalid24", source: `declare function get<T>(): T;`},
		{name: "invalid25", source: `declare function get<T extends object>(): T;`},
		{name: "invalid36", source: `declare function makeReadonlyArray<T>(): readonly T[];`},
		{name: "invalid37", source: `declare function makeReadonlyTuple<T>(): readonly [T];`},
		{name: "invalid38", source: `declare function makeReadonlyTupleNullish<T>(): readonly [T | null];`},
		{name: "invalid44", source: `type Fn = <T>() => T;`},
		{name: "invalid49", source: `type Ctr = new <T>() => T;`},
		{name: "invalid50", source: `type Fn = <T>() => { [K in keyof T]: K };`},
		{name: "invalid51", source: `type Fn = <T>() => { [K in 'a']: T };`},
		{name: "invalid53", source: "type Fn = <T extends string>() => `a${T}b`;"},
		{name: "invalid65", source: `
type A = string;
type B = string;
type C = string;
declare function f<T extends A | B>(): T & C;
      `},
		{name: "invalid66", source: `
type A = string;
type B = string;
type C = string;
type D = string;
declare function f<T extends (A extends B ? C : D)>(): T | null;
      `},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}

	// Each control differs from a witness in one property the principle reads, and reports.
	reporting := []noUnnecessaryTypeParametersCase{
		{name: "a value parameter is something to cast from", source: "declare function parse<T>(input: string): T;\n", ids: []string{"sole"}},
		{name: "typeOnly with a parameter", source: "export function typeOnly<Shape>(seed: unknown): Shape {\n    return seed as Shape;\n}\n", ids: []string{"sole"}},
		{name: "a this parameter is a receiver to cast from", source: "declare function fromThis<T>(this: Window): T;\n", ids: []string{"sole"}},
		{name: "a parameter with a default is still a parameter", source: "declare function make<T>(seed?: number): T;\n", ids: []string{"sole"}},
		{name: "never used is not a witness", source: "declare function nothing<T>(): void;\n", ids: []string{"sole"}},
		{name: "used in another parameter's constraint, outside the return", source: "declare function pair<T, U extends T>(): U;\n", ids: []string{"sole"}},
		{name: "a method with a parameter", source: "interface Store { read<T>(key: string): T }\n", ids: []string{"sole"}},
		// A method's receiver is something to cast from, as a this parameter is (#gtgw3av). The three are
		// api-phi-health's WorkerQueueProcessor.ts:67, OrmTrackingEntity.ts:70 and IntegrationTestEnvironment.ts:68.
		{name: "a method returning a field as its type parameter", source: "export class Processor {\n    private message: unknown;\n    getMessage<MessageType = unknown>(): MessageType {\n        return this.message as MessageType;\n    }\n}\n", ids: []string{"sole"}},
		{name: "a method cloning its receiver as its type parameter", source: "export class Entity {\n    clone<T extends this>(): T {\n        return Object.create(this) as T;\n    }\n}\n", ids: []string{"sole"}},
		{name: "a method widening a field with its type parameter", source: "interface Base { a: string }\nexport class Environment {\n    private variables: Base = { a: '' };\n    getEnvironmentVariables<T>(): Base & T {\n        return this.variables as Base & T;\n    }\n}\n", ids: []string{"sole"}},
		{name: "a parameterless method signature", source: "interface Store { current<T>(): T }\n", ids: []string{"sole"}},
		// The default spelling of the row above. A use inside a parameter's OWN constraint has no control:
		// the type walk counts every self-constrained shape (`<T extends Array<T>>(): T`) three times,
		// past the threshold, so it is silent before the witness test is ever asked.
		{name: "used in another parameter's default, outside the return", source: "declare function pick<T, U = T>(): U;\n", ids: []string{"sole"}},
		{name: "no return annotation to witness through", source: "export function inferred<T>() { return null as unknown as T; }\n", ids: []string{"sole"}},
		// A body that returns a real value cast to the caller's type is a cast from the world, not a
		// witness (Kirk's ruling of 2026-10-02). The first row is api-phi-health's
		// OrmDrizzleConfiguration.ts:38, ormDrizzleCredentialsFromEnvironment, which ESLint reports.
		{name: "a free function returning what it reads from the environment", source: "declare const process: { env: Record<string, string | undefined> };\nexport function credentialsFromEnvironment<CredentialsType>(): CredentialsType {\n    return JSON.parse(process.env.CREDENTIALS ?? '{}') as CredentialsType;\n}\n", ids: []string{"sole"}},
		{name: "an arrow whose expression body is a real value", source: "export const fromStorage = <Shape,>(): Shape => JSON.parse('{}') as Shape;\n", ids: []string{"sole"}},
		{name: "one null return beside one real return", source: "export function maybe<Shape>(): Shape {\n    if(Math.random() > 0.5) {\n        return null as unknown as Shape;\n    }\n    return {} as Shape;\n}\n", ids: []string{"sole"}},
	}
	for _, testCase := range reporting {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

// TestNoUnnecessaryTypeParametersPhantomBrands pins the phantom-brand exemption
// (isPhantomBrandParameter): a class type parameter whose use is the type of a `declare` property is a
// brand, since the property emits nothing. @system_cohere's ruling of 2026-10-02 (#ye9s2jx).
//
// typescript-eslint reports every row, silent and reporting alike (measured on the installed 8.67.0
// build), so the silent rows are the divergence. The first two are Base's own classes. The reporting
// rows are what keep the exemption from being "any class": an ordinary property is a real slot, an
// ambient class's members carry no `declare` modifier of their own, and a brand that names something
// else does not cover a parameter used once in a method.
func TestNoUnnecessaryTypeParametersPhantomBrands(t *testing.T) {
	t.Parallel()

	const typedBinding = "abstract class TypedBinding {\n  constructor(public readonly name: string) {}\n}\n"
	silent := []noUnnecessaryTypeParametersCase{
		{name: "WorkerQueueBinding.ts", source: typedBinding + "export class WorkerQueueBinding<MessageType = unknown> extends TypedBinding {\n  declare private readonly __workerQueue: MessageType;\n}\n"},
		{name: "RpcClientBinding.ts", source: typedBinding + "export class RpcClientBinding<RpcInterface extends object = object> extends TypedBinding {\n  declare private readonly __rpcClient: RpcInterface;\n}\n"},
		{name: "a public brand", source: "export class PublicBrand<M> {\n  declare readonly brand: M;\n}\n"},
		{name: "a brand under a union", source: "export class UnionBrand<M> {\n  declare readonly brand: M | undefined;\n}\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}

	reporting := []noUnnecessaryTypeParametersCase{
		{name: "an ordinary property", source: "export class Holder<M> {\n  readonly value: M | undefined = undefined;\n}\n", ids: []string{"sole"}},
		{name: "a member of an ambient class", source: "export declare class Ambient<M> {\n  brand: M;\n}\n", ids: []string{"sole"}},
		{name: "a brand naming something else", source: "export class BrandElsewhere<M> {\n  declare readonly brand: string;\n  take(message: M): void {\n    void message;\n  }\n}\n", ids: []string{"sole"}},
		// The brand names a TYPE this time, so only resolving the identifier, not spotting one, tells it
		// apart from a brand carrying M.
		{name: "a brand naming another type", source: "export class BrandOther<M> {\n  declare readonly brand: Date;\n  take(message: M): void {\n    void message;\n  }\n}\n", ids: []string{"sole"}},
	}
	for _, testCase := range reporting {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

// TestNoUnnecessaryTypeParametersCountsAnObjectTypesPropertiesOnce is upstream's list guard (#vn5vpfs).
//
// `Extract<ComponentType, { type: Kind }>` distributes over the union, and each constituent's
// conditional carries the same `{ type: Kind }` object. Upstream walks that object's properties once,
// so `Kind` counts once and the function reports: Structure's `surveyFieldRendererFor` at ea00a681,
// which ESLint reported and cohere did not. The site is fixed in Structure 3146950 by narrowing the
// map to `Pick<RenderersType, Kind>`, so this fixture is the only place the shape still lives.
// Both replayed through the installed typescript-eslint 8.67.0: the first reports `sole` at 7:29,
// the second is silent.
func TestNoUnnecessaryTypeParametersCountsAnObjectTypesPropertiesOnce(t *testing.T) {
	t.Parallel()

	fires := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, `interface ShortAnswerInterface { type: 'ShortAnswer'; placeholder: string }
interface RatingInterface { type: 'Rating'; maximum: number }
type ComponentType = ShortAnswerInterface | RatingInterface;
type RendererType<Kind extends ComponentType['type']> = (component: Extract<ComponentType, { type: Kind }>) => string;
type RenderersType = { [Kind in ComponentType['type']]?: RendererType<Kind> };

export function rendererFor<Kind extends ComponentType['type']>(
    renderers: RenderersType,
    component: Extract<ComponentType, { type: Kind }>,
): (() => string) | undefined {
    const renderer = renderers[component.type];
    if(!renderer) {
        return undefined;
    }
    return function() {
        return renderer(component);
    };
}
`)
	rule_testing.ExpectFindings(t, fires, "sole")
	if !strings.HasPrefix(fires.Diagnostics[0].Message.Description, "Type parameter Kind is used only once in the function signature.") {
		t.Errorf("expected upstream's sentence first, got %q", fires.Diagnostics[0].Message.Description)
	}

	silent := rule_testing.RunTyped(t, NoUnnecessaryTypeParameters, noUnnecessaryTypeParametersFile, `interface ShortAnswerInterface { type: 'ShortAnswer'; placeholder: string }
interface RatingInterface { type: 'Rating'; maximum: number }
type ComponentType = ShortAnswerInterface | RatingInterface;
type RendererType<Kind extends ComponentType['type']> = (component: Extract<ComponentType, { type: Kind }>) => string;
type RenderersType = { [Kind in ComponentType['type']]?: RendererType<Kind> };

export function rendererFor<Kind extends ComponentType['type']>(
    renderers: Pick<RenderersType, Kind>,
    component: Extract<ComponentType, { type: Kind }>,
): (() => string) | undefined {
    const renderer = renderers[component.type];
    if(!renderer) {
        return undefined;
    }
    return function() {
        return renderer(component);
    };
}
`)
	rule_testing.ExpectClean(t, silent)
}
