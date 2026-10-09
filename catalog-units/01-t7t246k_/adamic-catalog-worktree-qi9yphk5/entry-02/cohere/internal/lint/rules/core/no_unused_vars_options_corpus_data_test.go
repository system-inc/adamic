package core

// Code generated from typescript-eslint v8.71.0's tests/rules/no-unused-vars/*.test.ts and edge rows. DO NOT EDIT BY HAND.
//
// The 26 upstream rows are every row in the three test files whose options set ignoreClassWithStaticInitBlock, ignoreUsingDeclarations, reportUsedIgnorePattern:
// 10 valid and 16 invalid, extracted by loading the files with RuleTester stubbed. 17 edge rows were added for
// what they leave open. Before recording, every upstream row was linted untyped, as upstream's RuleTester runs it,
// and agreed with upstream's own assertions. Every row was then replayed against the INSTALLED rule
// (@typescript-eslint/eslint-plugin 8.71.0, one program per row) under noUnusedVarsOptionsCorpusTsconfig, on the bytes
// this harness writes: the source trimmed plus one newline. A row upstream parses as a module with no import or
// export gets `export {};` appended (exportAppended), since cohere's module is TypeScript's. Each finding records
// upstream's message id and the exact text its location covers. TestNoUnusedVarsOptionsUpstreamCorpus replays this.

// noUnusedVarsOptionsCorpusTsconfig is the options tests' tsconfig, with the disposable lib `using` needs.
const noUnusedVarsOptionsCorpusTsconfig = "{\n \"compilerOptions\": {\n  \"strict\": true,\n  \"target\": \"ES2022\",\n  \"lib\": [\n   \"ES2022\",\n   \"ESNext.Disposable\"\n  ],\n  \"types\": []\n },\n \"include\": [\n  \"file.ts\"\n ]\n}"

// noUnusedVarsOptionsCorpusFinding is one finding the installed rule produced, by upstream's message id.
type noUnusedVarsOptionsCorpusFinding struct {
	id   string
	text string
}

// noUnusedVarsOptionsCorpusCase is one row with the installed rule's verdict. edge is empty for upstream's own rows.
type noUnusedVarsOptionsCorpusCase struct {
	index          int
	edge           string
	source         string
	options        string
	exportAppended bool
	pinned         []string
	findings       []noUnusedVarsOptionsCorpusFinding
}

// noUnusedVarsOptionsCorpusUpstreamRows and noUnusedVarsOptionsCorpusUpstreamPinned pin how many of upstream's rows are
// here and how many carry a pin, so a filter cannot empty the table.
const noUnusedVarsOptionsCorpusUpstreamRows, noUnusedVarsOptionsCorpusUpstreamPinned = 26, 0

var noUnusedVarsOptionsCorpus = []noUnusedVarsOptionsCorpusCase{
	{0, "", "class Foo {\n  static {}\n}\n", "{\"ignoreClassWithStaticInitBlock\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{1, "", "class Foo {\n  static {}\n}\n", "{\"ignoreClassWithStaticInitBlock\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{2, "", "class Foo {\n  static {}\n}\n", "{\"ignoreClassWithStaticInitBlock\":false,\"varsIgnorePattern\":\"^Foo\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{3, "", "const a = 5;\nconst _c = a + 5;\n", "{\"args\":\"all\",\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{4, "", "(function foo(a, _b) {\n  return a + 5;\n})(5);\n", "{\"args\":\"all\",\"argsIgnorePattern\":\"^_\",\"reportUsedIgnorePattern\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{5, "", "const [a, _b, c] = items;\nconsole.log(a + c);\n", "{\"destructuredArrayIgnorePattern\":\"^_\",\"reportUsedIgnorePattern\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{6, "", "class Foo {\n  static {}\n}\n", "{\"ignoreClassWithStaticInitBlock\":false}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"unusedVar", "Foo"}}},
	{7, "", "class Foo {\n  static {\n    var bar;\n  }\n}\n", "{\"ignoreClassWithStaticInitBlock\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"unusedVar", "bar"}}},
	{8, "", "class Foo {}\n", "{\"ignoreClassWithStaticInitBlock\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"unusedVar", "Foo"}}},
	{9, "", "class Foo {\n  static bar;\n}\n", "{\"ignoreClassWithStaticInitBlock\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"unusedVar", "Foo"}}},
	{10, "", "class Foo {\n  static bar() {}\n}\n", "{\"ignoreClassWithStaticInitBlock\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"unusedVar", "Foo"}}},
	{11, "", "const _a = 5;\nconst _b = _a + 5;\n", "{\"args\":\"all\",\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_a"}}},
	{12, "", "const _a = 42;\nfoo(() => _a);\n", "{\"args\":\"all\",\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_a"}}},
	{13, "", "(function foo(_a) {\n  return _a + 5;\n})(5);\n", "{\"args\":\"all\",\"argsIgnorePattern\":\"^_\",\"reportUsedIgnorePattern\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_a"}}},
	{14, "", "const [a, _b] = items;\nconsole.log(a + _b);\n", "{\"destructuredArrayIgnorePattern\":\"^_\",\"reportUsedIgnorePattern\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_b"}}},
	{15, "", "let _x;\n[_x] = arr;\nfoo(_x);\n", "{\"destructuredArrayIgnorePattern\":\"^_\",\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"[iI]gnored\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_x"}}},
	{16, "", "const [ignored] = arr;\nfoo(ignored);\n", "{\"destructuredArrayIgnorePattern\":\"^_\",\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"[iI]gnored\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "ignored"}}},
	{17, "", "try {\n} catch (_err) {\n  console.error(_err);\n}\n", "{\"caughtErrors\":\"all\",\"caughtErrorsIgnorePattern\":\"^_\",\"reportUsedIgnorePattern\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_err"}}},
	{18, "", "type _Foo = 1;\nexport const x: _Foo = 1;\n", "{\"reportUsedIgnorePattern\":false,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{19, "", "export enum Foo {\n  _A,\n}\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{20, "", "using resource = getResource();\nexport {};\n", "{\"ignoreUsingDeclarations\":true}", true, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{21, "", "await using resource = getResource();\nexport {};\n", "{\"ignoreUsingDeclarations\":true}", true, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{22, "", "type _Foo = 1;\nexport const x: _Foo = 1;\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_Foo"}}},
	{23, "", "interface _Foo {}\nexport const x: _Foo = 1;\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_Foo"}}},
	{24, "", "enum _Foo {\n  A = 1,\n}\nexport const x = _Foo.A;\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_Foo"}}},
	{25, "", "namespace _Foo {}\nexport const x = _Foo;\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_Foo"}}},
	{26, "await using is exempt too", "export async function f() {\n  await using resource = getResource();\n}\ndeclare function getResource(): AsyncDisposable;\n", "{\"ignoreUsingDeclarations\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{27, "using without the option reports", "export function f() {\n  using resource = getResource();\n}\ndeclare function getResource(): Disposable;\n", "", false, nil, []noUnusedVarsOptionsCorpusFinding{{"unusedVar", "resource"}}},
	{28, "a const is not a using declaration", "export function f() {\n  const resource = 1;\n}\n", "{\"ignoreUsingDeclarations\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"unusedVar", "resource"}}},
	{29, "a used ignored using declaration still reports", "export function f() {\n  using _resource = getResource();\n  return _resource;\n}\ndeclare function getResource(): Disposable;\n", "{\"ignoreUsingDeclarations\":true,\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_resource"}}},
	{30, "an exported ignored name is used", "export const _x = 1;\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_x"}}},
	{31, "an ignored import that is read", "import { _a } from 'module';\nconsole.log(_a);\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_a"}}},
	{32, "an ignored import nothing reads", "import { _a } from 'module';\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{33, "a used ignored name reports at its last write", "let _a = 1;\n_a = 2;\nconsole.log(_a);\nexport {};\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_a"}}},
	{34, "an ignored name read only as a type", "const _d = {};\nexport type D = typeof _d;\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{35, "a static-block class is set aside before its pattern", "class _Foo {\n  static {}\n}\nnew _Foo();\nexport {};\n", "{\"ignoreClassWithStaticInitBlock\":true,\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{36, "an ignored parameter before a used one", "export function f(_a: number, b: number) {\n  return b;\n}\n", "{\"argsIgnorePattern\":\"^_\",\"reportUsedIgnorePattern\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{37, "args none is asked before the pattern", "export function f(_a: number) {\n  return _a;\n}\n", "{\"args\":\"none\",\"argsIgnorePattern\":\"^_\",\"reportUsedIgnorePattern\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{38, "caughtErrors none is asked before the pattern", "try {\n} catch (_error) {\n  console.log(_error);\n}\nexport {};\n", "{\"caughtErrors\":\"none\",\"caughtErrorsIgnorePattern\":\"^_\",\"reportUsedIgnorePattern\":true}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{39, "vars local is asked before the pattern", "var _global = 1;\nconsole.log(_global);\n", "{\"vars\":\"local\",\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{}},
	{40, "a for-in leading return under a pattern", "export function f(o: object) {\n  for (const _key in o) {\n    return true;\n  }\n  return false;\n}\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_key"}}},
	{41, "a static block in a class nothing reads, option off", "class Foo {\n  static {}\n}\nexport {};\n", "{\"ignoreClassWithStaticInitBlock\":false}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"unusedVar", "Foo"}}},
	{42, "an ignored type alias that is used", "type _T = number;\nexport const value: _T = 1;\n", "{\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", false, nil, []noUnusedVarsOptionsCorpusFinding{{"usedIgnoredVar", "_T"}}},
}
