package typescript

// Code generated from typescript-eslint v8.71.0's tests/rules/dot-notation.test.ts and edge rows. DO NOT EDIT BY HAND.
//
// The 38 valid and 25 invalid rows were extracted by loading upstream's test file with its RuleTester stubbed,
// and 40 edge rows were added for what that corpus leaves open. Every row was replayed against the
// INSTALLED rule (@typescript-eslint/eslint-plugin 8.71.0, one program per row) under dotNotationCorpusTsconfig, on
// the bytes this harness writes: the source trimmed plus one newline. Before recording, every upstream
// row was replayed on its own bytes and agreed with upstream's own assertions. Each finding records the
// message id and the exact text the installed rule's location covers, and output is what one pass of
// its fixes writes, empty when nothing changes. TestDotNotationUpstreamCorpus replays this table.

// dotNotationCorpusTsconfig is upstream's fixture tsconfig with types emptied, so nothing installed
// near the test directory leaks in. A row with flag set also turns on noPropertyAccessFromIndexSignature.
const dotNotationCorpusTsconfig = "{\n \"compilerOptions\": {\n  \"jsx\": \"preserve\",\n  \"target\": \"es2015\",\n  \"module\": \"commonjs\",\n  \"strict\": true,\n  \"types\": [],\n  \"lib\": [\n   \"es2015\",\n   \"es2017\",\n   \"esnext\"\n  ],\n  \"experimentalDecorators\": true,\n  \"stableTypeOrdering\": true\n },\n \"include\": [\n  \"file.ts\"\n ]\n}"

// dotNotationCorpusFinding is one finding the installed rule produced.
type dotNotationCorpusFinding struct {
	id   string
	text string
}

// dotNotationCorpusCase is one row with the installed rule's verdict. edge is empty for upstream's own rows.
type dotNotationCorpusCase struct {
	index    int
	edge     string
	source   string
	options  string
	flag     bool
	findings []dotNotationCorpusFinding
	output   string
}

// dotNotationCorpusUpstreamValid and dotNotationCorpusUpstreamInvalid pin how many of upstream's rows are
// here, so a filter cannot empty either direction.
const dotNotationCorpusUpstreamValid, dotNotationCorpusUpstreamInvalid = 38, 25

var dotNotationCorpus = []dotNotationCorpusCase{
	{0, "", "a.b;\n", "", false, []dotNotationCorpusFinding{}, ""},
	{1, "", "a.b.c;\n", "", false, []dotNotationCorpusFinding{}, ""},
	{2, "", "a['12'];\n", "", false, []dotNotationCorpusFinding{}, ""},
	{3, "", "a[b];\n", "", false, []dotNotationCorpusFinding{}, ""},
	{4, "", "a[0];\n", "", false, []dotNotationCorpusFinding{}, ""},
	{5, "", "a.b.c;\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{6, "", "a.arguments;\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{7, "", "a.let;\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{8, "", "a.yield;\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{9, "", "a.eval;\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{10, "", "a[0];\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{11, "", "a['while'];\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{12, "", "a['true'];\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{13, "", "a['null'];\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{14, "", "a[true];\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{15, "", "a[null];\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{16, "", "a.true;\n", "{\"allowKeywords\":true}", false, []dotNotationCorpusFinding{}, ""},
	{17, "", "a.null;\n", "{\"allowKeywords\":true}", false, []dotNotationCorpusFinding{}, ""},
	{18, "", "a['snake_case'];\n", "{\"allowPattern\":\"^[a-z]+(_[a-z]+)+$\"}", false, []dotNotationCorpusFinding{}, ""},
	{19, "", "a['lots_of_snake_case'];\n", "{\"allowPattern\":\"^[a-z]+(_[a-z]+)+$\"}", false, []dotNotationCorpusFinding{}, ""},
	{20, "", "a[`time${range}`];\n", "", false, []dotNotationCorpusFinding{}, ""},
	{21, "", "a[`while`];\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{}, ""},
	{22, "", "a[`time range`];\n", "", false, []dotNotationCorpusFinding{}, ""},
	{23, "", "a.true;\n", "", false, []dotNotationCorpusFinding{}, ""},
	{24, "", "a.null;\n", "", false, []dotNotationCorpusFinding{}, ""},
	{25, "", "a[undefined];\n", "", false, []dotNotationCorpusFinding{}, ""},
	{26, "", "a[void 0];\n", "", false, []dotNotationCorpusFinding{}, ""},
	{27, "", "a[b()];\n", "", false, []dotNotationCorpusFinding{}, ""},
	{28, "", "a[/(?<zero>0)/];\n", "", false, []dotNotationCorpusFinding{}, ""},
	{29, "", "class X {\n  private priv_prop = 123;\n}\n\nconst x = new X();\nx['priv_prop'] = 123;\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{30, "", "class X {\n  protected protected_prop = 123;\n}\n\nconst x = new X();\nx['protected_prop'] = 123;\n", "{\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{31, "", "class X {\n  prop: string;\n  [key: string]: number;\n}\n\nconst x = new X();\nx['hello'] = 3;\n", "{\"allowIndexSignaturePropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{32, "", "interface Nested {\n  property: string;\n  [key: string]: number | string;\n}\n\nclass Dingus {\n  nested: Nested;\n}\n\nlet dingus: Dingus | undefined;\n\ndingus?.nested.property;\ndingus?.nested['hello'];\n", "{\"allowIndexSignaturePropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{33, "", "class X {\n  private priv_prop = 123;\n}\n\nlet x: X | undefined;\nconsole.log(x?.['priv_prop']);\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{34, "", "class X {\n  protected priv_prop = 123;\n}\n\nlet x: X | undefined;\nconsole.log(x?.['priv_prop']);\n", "{\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{35, "", "type Foo = {\n  bar: boolean;\n  [key: `key_${string}`]: number;\n};\ndeclare const foo: Foo;\nfoo['key_baz'];\n", "", true, []dotNotationCorpusFinding{}, ""},
	{36, "", "type Key = Lowercase<string>;\ntype Foo = {\n  BAR: boolean;\n  [key: Lowercase<string>]: number;\n};\ndeclare const foo: Foo;\nfoo['bar'];\n", "", true, []dotNotationCorpusFinding{}, ""},
	{37, "", "type ExtraKey = `extra${string}`;\n\ntype Foo = {\n  foo: string;\n  [extraKey: ExtraKey]: number;\n};\n\nfunction f<T extends Foo>(x: T) {\n  x['extraKey'];\n}\n", "", true, []dotNotationCorpusFinding{}, ""},
	{38, "", "class X {\n  private priv_prop = 123;\n}\n\nconst x = new X();\nx['priv_prop'] = 123;\n", "{\"allowPrivateClassPropertyAccess\":false}", false, []dotNotationCorpusFinding{{"useDot", "'priv_prop'"}}, "class X {\n  private priv_prop = 123;\n}\n\nconst x = new X();\nx.priv_prop = 123;\n"},
	{39, "", "class X {\n  public pub_prop = 123;\n}\n\nconst x = new X();\nx['pub_prop'] = 123;\n", "", false, []dotNotationCorpusFinding{{"useDot", "'pub_prop'"}}, "class X {\n  public pub_prop = 123;\n}\n\nconst x = new X();\nx.pub_prop = 123;\n"},
	{40, "", "a['true'];\n", "", false, []dotNotationCorpusFinding{{"useDot", "'true'"}}, "a.true;\n"},
	{41, "", "a['time'];\n", "", false, []dotNotationCorpusFinding{{"useDot", "'time'"}}, "a.time;\n"},
	{42, "", "a[null];\n", "", false, []dotNotationCorpusFinding{{"useDot", "null"}}, "a.null;\n"},
	{43, "", "a[true];\n", "", false, []dotNotationCorpusFinding{{"useDot", "true"}}, "a.true;\n"},
	{44, "", "a[false];\n", "", false, []dotNotationCorpusFinding{{"useDot", "false"}}, "a.false;\n"},
	{45, "", "a['b'];\n", "", false, []dotNotationCorpusFinding{{"useDot", "'b'"}}, "a.b;\n"},
	{46, "", "a.b['c'];\n", "", false, []dotNotationCorpusFinding{{"useDot", "'c'"}}, "a.b.c;\n"},
	{47, "", "a['_dangle'];\n", "{\"allowPattern\":\"^[a-z]+(_[a-z]+)+$\"}", false, []dotNotationCorpusFinding{{"useDot", "'_dangle'"}}, "a._dangle;\n"},
	{48, "", "a['SHOUT_CASE'];\n", "{\"allowPattern\":\"^[a-z]+(_[a-z]+)+$\"}", false, []dotNotationCorpusFinding{{"useDot", "'SHOUT_CASE'"}}, "a.SHOUT_CASE;\n"},
	{49, "", "a\n  ['SHOUT_CASE'];\n", "", false, []dotNotationCorpusFinding{{"useDot", "'SHOUT_CASE'"}}, "a\n  .SHOUT_CASE;\n"},
	{50, "", "getResource()\n  .then(function () {})\n  ['catch'](function () {})\n  .then(function () {})\n  ['catch'](function () {});\n", "", false, []dotNotationCorpusFinding{{"useDot", "'catch'"}, {"useDot", "'catch'"}}, "getResource()\n  .then(function () {})\n  .catch(function () {})\n  .then(function () {})\n  .catch(function () {});\n"},
	{51, "", "foo\n  .while;\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{{"useBrackets", "while"}}, "foo\n  [\"while\"];\n"},
	{52, "", "foo[/* comment */ 'bar'];\n", "", false, []dotNotationCorpusFinding{{"useDot", "'bar'"}}, ""},
	{53, "", "foo['bar' /* comment */];\n", "", false, []dotNotationCorpusFinding{{"useDot", "'bar'"}}, ""},
	{54, "", "foo['bar'];\n", "", false, []dotNotationCorpusFinding{{"useDot", "'bar'"}}, "foo.bar;\n"},
	{55, "", "foo./* comment */ while;\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{{"useBrackets", "while"}}, ""},
	{56, "", "foo[null];\n", "", false, []dotNotationCorpusFinding{{"useDot", "null"}}, "foo.null;\n"},
	{57, "", "foo['bar'] instanceof baz;\n", "", false, []dotNotationCorpusFinding{{"useDot", "'bar'"}}, "foo.bar instanceof baz;\n"},
	{58, "", "let.if();\n", "{\"allowKeywords\":false}", false, []dotNotationCorpusFinding{{"useBrackets", "if"}}, ""},
	{59, "", "class X {\n  protected protected_prop = 123;\n}\n\nconst x = new X();\nx['protected_prop'] = 123;\n", "{\"allowProtectedClassPropertyAccess\":false}", false, []dotNotationCorpusFinding{{"useDot", "'protected_prop'"}}, "class X {\n  protected protected_prop = 123;\n}\n\nconst x = new X();\nx.protected_prop = 123;\n"},
	{60, "", "class X {\n  prop: string;\n  [key: string]: number;\n}\n\nconst x = new X();\nx['prop'] = 'hello';\n", "{\"allowIndexSignaturePropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'prop'"}}, "class X {\n  prop: string;\n  [key: string]: number;\n}\n\nconst x = new X();\nx.prop = 'hello';\n"},
	{61, "", "type Foo = {\n  bar: boolean;\n  [key: `key_${string}`]: number;\n};\nfoo['key_baz'];\n", "", false, []dotNotationCorpusFinding{{"useDot", "'key_baz'"}}, "type Foo = {\n  bar: boolean;\n  [key: `key_${string}`]: number;\n};\nfoo.key_baz;\n"},
	{62, "", "type ExtraKey = `extra${string}`;\n\ntype Foo = {\n  foo: string;\n  [extraKey: ExtraKey]: number;\n};\n\nfunction f<T extends Foo>(x: T) {\n  x['extraKey'];\n}\n", "", false, []dotNotationCorpusFinding{{"useDot", "'extraKey'"}}, "type ExtraKey = `extra${string}`;\n\ntype Foo = {\n  foo: string;\n  [extraKey: ExtraKey]: number;\n};\n\nfunction f<T extends Foo>(x: T) {\n  x.extraKey;\n}\n"},
	{63, "private under the protected option alone", "class X {\n  private secret = 1;\n}\nconst x = new X();\nx['secret'];\n", "{\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'secret'"}}, "class X {\n  private secret = 1;\n}\nconst x = new X();\nx.secret;\n"},
	{64, "protected under the private option alone", "class X {\n  protected guarded = 1;\n}\nconst x = new X();\nx['guarded'];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'guarded'"}}, "class X {\n  protected guarded = 1;\n}\nconst x = new X();\nx.guarded;\n"},
	{65, "a private parameter property", "class X {\n  constructor(private secret: number) {}\n}\nconst x = new X(1);\nx['secret'];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{66, "a protected readonly parameter property", "class X {\n  constructor(protected readonly guarded: number) {}\n}\nconst x = new X(1);\nx['guarded'];\n", "{\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{67, "a private static", "class X {\n  private static secret = 1;\n}\nX['secret'];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{68, "a static private is a type error but parses", "class X {\n  static secret = 1;\n}\nX['secret'];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'secret'"}}, "class X {\n  static secret = 1;\n}\nX.secret;\n"},
	{69, "readonly alone", "class X {\n  readonly fixed = 1;\n}\nconst x = new X();\nx['fixed'];\n", "{\"allowPrivateClassPropertyAccess\":true,\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'fixed'"}}, "class X {\n  readonly fixed = 1;\n}\nconst x = new X();\nx.fixed;\n"},
	{70, "a protected method called through brackets", "class X {\n  protected clear() {}\n}\nconst x = new X();\nx['clear']();\n", "{\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{71, "a private getter", "class X {\n  private get value() { return 1; }\n}\nconst x = new X();\nx['value'];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{72, "a protected member reached through this in a subclass", "class X {\n  protected guarded = 1;\n}\nclass Y extends X {\n  read() { return this['guarded']; }\n}\n", "{\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{73, "a parenthesized key", "class X {\n  private secret = 1;\n}\nconst x = new X();\nx[('secret')];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{74, "a template key", "class X {\n  private secret = 1;\n}\nconst x = new X();\nx[`secret`];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{75, "a protected abstract", "abstract class X {\n  protected abstract guarded: number;\n}\ndeclare const x: X;\nx['guarded'];\n", "{\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{76, "an override protected", "class X {\n  protected guarded = 1;\n}\nclass Y extends X {\n  protected override guarded = 2;\n}\ndeclare const y: Y;\ny['guarded'];\n", "{\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{77, "an interface property", "interface Shape {\n  side: number;\n}\ndeclare const shape: Shape;\nshape['side'];\n", "{\"allowPrivateClassPropertyAccess\":true,\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'side'"}}, "interface Shape {\n  side: number;\n}\ndeclare const shape: Shape;\nshape.side;\n"},
	{78, "a union whose first member's property is private", "class A {\n  private value = 1;\n}\nclass B {\n  value = 2;\n}\ndeclare const either: A | B;\neither['value'];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'value'"}}, "class A {\n  private value = 1;\n}\nclass B {\n  value = 2;\n}\ndeclare const either: A | B;\neither.value;\n"},
	{79, "a union whose first member's property is public", "class A {\n  value = 2;\n}\nclass B {\n  private value = 1;\n}\ndeclare const either: A | B;\neither['value'];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'value'"}}, "class A {\n  value = 2;\n}\nclass B {\n  private value = 1;\n}\ndeclare const either: A | B;\neither.value;\n"},
	{80, "a property behind undefined", "class X {\n  protected guarded = 1;\n}\ndeclare const x: X | undefined;\nx!['guarded'];\n", "{\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{81, "a number index signature only", "declare const list: { [index: number]: string };\nlist['first'];\n", "{\"allowIndexSignaturePropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'first'"}}, "declare const list: { [index: number]: string };\nlist.first;\n"},
	{82, "a Record of string", "declare const table: Record<string, number>;\ntable['row'];\n", "{\"allowIndexSignaturePropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{83, "a Record of string without the option", "declare const table: Record<string, number>;\ntable['row'];\n", "", false, []dotNotationCorpusFinding{{"useDot", "'row'"}}, "declare const table: Record<string, number>;\ntable.row;\n"},
	{84, "a mapped type over string", "declare const table: { [K in string]: number };\ntable['row'];\n", "{\"allowIndexSignaturePropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{85, "an index signature behind undefined", "declare const table: Record<string, number> | undefined;\ntable!['row'];\n", "{\"allowIndexSignaturePropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{86, "an index signature on an intersection", "declare const table: { known: number } & { [key: string]: number };\ntable['row'];\n", "{\"allowIndexSignaturePropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{87, "a symbol-keyed index signature", "declare const table: { [key: symbol]: number };\ntable['row'];\n", "{\"allowIndexSignaturePropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'row'"}}, "declare const table: { [key: symbol]: number };\ntable.row;\n"},
	{88, "a template-literal index signature without the flag", "declare const table: { [key: `row_${string}`]: number };\ntable['row_one'];\n", "{\"allowIndexSignaturePropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{89, "any", "declare const loose: any;\nloose['row'];\n", "{\"allowIndexSignaturePropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'row'"}}, "declare const loose: any;\nloose.row;\n"},
	{90, "the flag with a Record of string", "declare const table: Record<string, number>;\ntable['row'];\n", "", true, []dotNotationCorpusFinding{}, ""},
	{91, "the flag with a known property beside an index signature", "declare const table: { known: number; [key: string]: number };\ntable['known'];\n", "", true, []dotNotationCorpusFinding{{"useDot", "'known'"}}, "declare const table: { known: number; [key: string]: number };\ntable.known;\n"},
	{92, "the flag with the option explicitly false", "declare const table: Record<string, number>;\ntable['row'];\n", "{\"allowIndexSignaturePropertyAccess\":false}", true, []dotNotationCorpusFinding{}, ""},
	{93, "allowKeywords false with the typed options", "class X {\n  private class = 1;\n}\nconst x = new X();\nx.class;\nx['class'];\n", "{\"allowKeywords\":false,\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useBrackets", "class"}}, "class X {\n  private class = 1;\n}\nconst x = new X();\nx[\"class\"];\nx['class'];\n"},
	{94, "allowPattern with the typed options", "class X {\n  snake_case = 1;\n}\nconst x = new X();\nx['snake_case'];\n", "{\"allowPattern\":\"^[a-z]+(_[a-z]+)+$\",\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{95, "a decorated public member", "declare function tracked(target: object, key: string): void;\nclass X {\n  @tracked value = 1;\n}\nconst x = new X();\nx['value'];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'value'"}}, "declare function tracked(target: object, key: string): void;\nclass X {\n  @tracked value = 1;\n}\nconst x = new X();\nx.value;\n"},
	{96, "a decorated private member", "declare function tracked(target: object, key: string): void;\nclass X {\n  @tracked private value = 1;\n}\nconst x = new X();\nx['value'];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{97, "a null key", "class X {\n  private null = 1;\n}\nconst x = new X();\nx[null];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "null"}}, "class X {\n  private null = 1;\n}\nconst x = new X();\nx.null;\n"},
	{98, "a true key", "class X {\n  private true = 1;\n}\nconst x = new X();\nx[true];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "true"}}, "class X {\n  private true = 1;\n}\nconst x = new X();\nx.true;\n"},
	{99, "the api shape: a test reaching protected methods", "class Entity {\n  protected clearChangedFields() {}\n  protected async afterSave() {}\n}\nconst entity = new Entity();\nentity['clearChangedFields']();\nvoid entity['afterSave']();\n", "{\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{}, ""},
	{100, "the api shape without the option", "class Entity {\n  protected clearChangedFields() {}\n}\nconst entity = new Entity();\nentity['clearChangedFields']();\n", "", false, []dotNotationCorpusFinding{{"useDot", "'clearChangedFields'"}}, "class Entity {\n  protected clearChangedFields() {}\n}\nconst entity = new Entity();\nentity.clearChangedFields();\n"},
	{101, "private written after static, which upstream reads as static", "class X {\n  static private secret = 1;\n}\nX['secret'];\n", "{\"allowPrivateClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'secret'"}}, "class X {\n  static private secret = 1;\n}\nX.secret;\n"},
	{102, "protected written after readonly, which upstream reads as readonly", "class X {\n  readonly protected guarded = 1;\n}\nconst x = new X();\nx['guarded'];\n", "{\"allowProtectedClassPropertyAccess\":true}", false, []dotNotationCorpusFinding{{"useDot", "'guarded'"}}, "class X {\n  readonly protected guarded = 1;\n}\nconst x = new X();\nx.guarded;\n"},
}
