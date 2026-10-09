package typescript

// Code generated from typescript-eslint v8.71.0's tests/rules/no-redeclare.test.ts and edge rows. DO NOT EDIT BY HAND.
//
// The 21 valid and 31 invalid rows were extracted by loading upstream's test file with its RuleTester stubbed,
// and 34 edge rows were added for what that corpus leaves open. Before recording, every upstream row
// was linted untyped, as upstream's RuleTester runs it, with its own parser options, globals and directives, and
// agreed with upstream's own assertions: message ids, lines and columns. Every row was then replayed against the
// INSTALLED rule (@typescript-eslint/eslint-plugin 8.71.0, one program per row) under noRedeclareCorpusTsconfig, its
// lib replaced where the row names one, on the bytes this harness writes: the source trimmed plus one newline.
// Each finding records the message id and the exact text the installed rule's location covers.
//
// cohere's module is TypeScript's, a file with an import or export, so a row upstream parses as
// `sourceType: "module"` with neither gets `export {};` appended (exportAppended), and both engines see a module.
// 5 upstream rows carry what cohere cannot express, named in pinned, and are recorded with that part
// taken out: globals configuration dropped, /*global*/ directives off through noInlineConfig, globalReturn dropped.
// TestNoRedeclareUpstreamCorpus replays this table.

// noRedeclareCorpusTsconfig is the dot-notation corpus's tsconfig, types emptied so nothing installed near
// the test directory leaks in. A row with lib set replaces its lib.
const noRedeclareCorpusTsconfig = "{\n \"compilerOptions\": {\n  \"jsx\": \"preserve\",\n  \"target\": \"es2015\",\n  \"module\": \"commonjs\",\n  \"strict\": true,\n  \"types\": [],\n  \"lib\": [\n   \"es2015\",\n   \"es2017\",\n   \"esnext\"\n  ],\n  \"experimentalDecorators\": true,\n  \"stableTypeOrdering\": true\n },\n \"include\": [\n  \"file.ts\"\n ]\n}"

// noRedeclareCorpusFinding is one finding the installed rule produced.
type noRedeclareCorpusFinding struct {
	id   string
	text string
}

// noRedeclareCorpusCase is one row with the installed rule's verdict. edge is empty for upstream's own rows.
type noRedeclareCorpusCase struct {
	index          int
	edge           string
	source         string
	options        string
	lib            []string
	exportAppended bool
	pinned         []string
	findings       []noRedeclareCorpusFinding
}

// noRedeclareCorpusUpstreamValid, noRedeclareCorpusUpstreamInvalid and noRedeclareCorpusUpstreamPinned pin how
// many of upstream's rows are here and how many carry a pin, so a filter cannot empty either direction.
const noRedeclareCorpusUpstreamValid, noRedeclareCorpusUpstreamInvalid, noRedeclareCorpusUpstreamPinned = 21, 31, 5

var noRedeclareCorpus = []noRedeclareCorpusCase{
	{0, "", "var a = 3;\nvar b = function () {\n  var a = 10;\n};\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{1, "", "var a = 3;\na = 10;\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{2, "", "if (true) {\n  let b = 2;\n} else {\n  let b = 3;\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{3, "", "var Object = 0;\n", "{\"builtinGlobals\":false}", nil, false, nil, []noRedeclareCorpusFinding{}},
	{4, "", "var Object = 0;\nexport {};\n", "{\"builtinGlobals\":true}", nil, true, nil, []noRedeclareCorpusFinding{}},
	{5, "", "var Object = 0;\n", "{\"builtinGlobals\":true}", nil, false, []string{"ecmaFeatures.globalReturn"}, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Object"}}},
	{6, "", "var top = 0;\n", "{\"builtinGlobals\":false}", nil, false, nil, []noRedeclareCorpusFinding{}},
	{7, "", "var top = 0;\n", "{\"builtinGlobals\":true}", nil, false, nil, []noRedeclareCorpusFinding{}},
	{8, "", "var top = 0;\n", "{\"builtinGlobals\":true}", nil, false, []string{"ecmaFeatures.globalReturn"}, []noRedeclareCorpusFinding{}},
	{9, "", "var top = 0;\nexport {};\n", "{\"builtinGlobals\":true}", nil, true, nil, []noRedeclareCorpusFinding{}},
	{10, "", "var self = 1;\n", "{\"builtinGlobals\":true}", nil, false, nil, []noRedeclareCorpusFinding{}},
	{11, "", "function foo({ bar }: { bar: string }) {\n  console.log(bar);\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{12, "", "type AST<T extends ParserOptions> = TSESTree.Program &\n  (T['range'] extends true ? { range: [number, number] } : {}) &\n  (T['tokens'] extends true ? { tokens: TSESTree.Token[] } : {}) &\n  (T['comment'] extends true ? { comments: TSESTree.Comment[] } : {});\ninterface ParseAndGenerateServicesResult<T extends ParserOptions> {\n  ast: AST<T>;\n  services: ParserServices;\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{13, "", "function A<T>() {}\ninterface B<T> {}\ntype C<T> = Array<T>;\nclass D<T> {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{14, "", "function a(): string;\nfunction a(): number;\nfunction a() {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{15, "", "interface A {}\ninterface A {}\n", "{\"ignoreDeclarationMerge\":true}", nil, false, nil, []noRedeclareCorpusFinding{}},
	{16, "", "interface A {}\nclass A {}\n", "{\"ignoreDeclarationMerge\":true}", nil, false, nil, []noRedeclareCorpusFinding{}},
	{17, "", "class A {}\nnamespace A {}\n", "{\"ignoreDeclarationMerge\":true}", nil, false, nil, []noRedeclareCorpusFinding{}},
	{18, "", "interface A {}\nclass A {}\nnamespace A {}\n", "{\"ignoreDeclarationMerge\":true}", nil, false, nil, []noRedeclareCorpusFinding{}},
	{19, "", "enum A {}\nnamespace A {}\n", "{\"ignoreDeclarationMerge\":true}", nil, false, nil, []noRedeclareCorpusFinding{}},
	{20, "", "function A() {}\nnamespace A {}\n", "{\"ignoreDeclarationMerge\":true}", nil, false, nil, []noRedeclareCorpusFinding{}},
	{21, "", "var a = 3;\nvar a = 10;\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{22, "", "switch (foo) {\n  case a:\n    var b = 3;\n  case b:\n    var b = 4;\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "b"}}},
	{23, "", "var a = 3;\nvar a = 10;\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{24, "", "var a = {};\nvar a = [];\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{25, "", "var a;\nfunction a() {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{26, "", "function a() {}\nfunction a() {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{27, "", "var a = function () {};\nvar a = function () {};\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{28, "", "var a = function () {};\nvar a = new Date();\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{29, "", "var a = 3;\nvar a = 10;\nvar a = 15;\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}, {"redeclared", "a"}}},
	{30, "", "var a;\nvar a;\nexport {};\n", "", nil, true, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{31, "", "export var a;\nvar a;\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{32, "", "var Object = 0;\n", "{\"builtinGlobals\":true}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Object"}}},
	{33, "", "var top = 0;\n", "{\"builtinGlobals\":true}", nil, false, []string{"globals configuration"}, []noRedeclareCorpusFinding{}},
	{34, "", "var a;\nvar { a = 0, b: Object = 0 } = {};\n", "{\"builtinGlobals\":true}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}, {"redeclaredAsBuiltin", "Object"}}},
	{35, "", "var a;\nvar { a = 0, b: Object = 0 } = {};\nexport {};\n", "{\"builtinGlobals\":true}", nil, true, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{36, "", "var a;\nvar { a = 0, b: Object = 0 } = {};\n", "{\"builtinGlobals\":true}", nil, false, []string{"ecmaFeatures.globalReturn"}, []noRedeclareCorpusFinding{{"redeclared", "a"}, {"redeclaredAsBuiltin", "Object"}}},
	{37, "", "var a;\nvar { a = 0, b: Object = 0 } = {};\n", "{\"builtinGlobals\":false}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{38, "", "/*global b:false*/ var b = 1;\n", "{\"builtinGlobals\":true}", nil, false, []string{"global comment directive"}, []noRedeclareCorpusFinding{}},
	{39, "", "type T = 1;\ntype T = 2;\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "T"}}},
	{40, "", "type NodeListOf = 1;\n", "{\"builtinGlobals\":true}", []string{"dom"}, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "NodeListOf"}}},
	{41, "", "interface A {}\ninterface A {}\n", "{\"ignoreDeclarationMerge\":false}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "A"}}},
	{42, "", "interface A {}\nclass A {}\n", "{\"ignoreDeclarationMerge\":false}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "A"}}},
	{43, "", "class A {}\nnamespace A {}\n", "{\"ignoreDeclarationMerge\":false}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "A"}}},
	{44, "", "interface A {}\nclass A {}\nnamespace A {}\n", "{\"ignoreDeclarationMerge\":false}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "A"}, {"redeclared", "A"}}},
	{45, "", "class A {}\nclass A {}\nnamespace A {}\n", "{\"ignoreDeclarationMerge\":true}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "A"}}},
	{46, "", "function A() {}\nnamespace A {}\n", "{\"ignoreDeclarationMerge\":false}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "A"}}},
	{47, "", "function A() {}\nfunction A() {}\nnamespace A {}\n", "{\"ignoreDeclarationMerge\":true}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "A"}}},
	{48, "", "function A() {}\nclass A {}\n", "{\"ignoreDeclarationMerge\":false}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "A"}}},
	{49, "", "enum A {}\nnamespace A {}\nenum A {}\n", "{\"ignoreDeclarationMerge\":true}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "A"}}},
	{50, "", "function A() {}\nclass A {}\nnamespace A {}\n", "{\"ignoreDeclarationMerge\":false}", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "A"}, {"redeclared", "A"}}},
	{51, "", "type something = string;\nconst something = 2;\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "something"}}},
	{52, "let at a script's top level is global", "let Object = 0;\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Object"}}},
	{53, "a lone augmenting interface in a script", "interface Window {}\n", "", []string{"es2015", "dom"}, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Window"}}},
	{54, "two augmenting interfaces merge", "interface Window {}\ninterface Window {}\n", "", []string{"es2015", "dom"}, false, nil, []noRedeclareCorpusFinding{}},
	{55, "a lone interface Object in a script", "interface Object {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Object"}}},
	{56, "an interface and a namespace merging with the lib", "interface Object {}\nnamespace Object {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{57, "two classes after a builtin both report", "class Object {}\nclass Object {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Object"}, {"redeclaredAsBuiltin", "Object"}}},
	{58, "two vars after a builtin both report", "var Object = 0;\nvar Object = 1;\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Object"}, {"redeclaredAsBuiltin", "Object"}}},
	{59, "ESLint's own NaN is a builtin", "var NaN = 1;\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "NaN"}}},
	{60, "a function redeclaring a builtin", "function Object() {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Object"}}},
	{61, "a declare function is an overload signature", "declare function Object(): void;\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{62, "a namespace redeclaring a lib namespace", "namespace Intl {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Intl"}}},
	{63, "an enum redeclaring a lib interface", "enum Symbol {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Symbol"}}},
	{64, "a type alias redeclaring a lib type", "type Partial<T> = T;\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Partial"}}},
	{65, "inside a function is not the global scope", "function f() {\n  var Object = 0;\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{66, "a var in a top-level block is global", "if (true) {\n  var Object = 0;\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Object"}}},
	{67, "a for-of var at a script's top level is global", "for (var Object of []) {\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Object"}}},
	{68, "an array pattern redeclaring a builtin", "var [Object] = [];\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "Object"}}},
	{69, "builtinGlobals off leaves a lone builtin alone", "interface Object {}\n", "{\"builtinGlobals\":false}", nil, false, nil, []noRedeclareCorpusFinding{}},
	{70, "a module's declare global is not the global scope", "export {};\ndeclare global {\n  interface Window {}\n}\n", "", []string{"es2015", "dom"}, false, nil, []noRedeclareCorpusFinding{}},
	{71, "a var hoists out of a block to the file", "var a;\nif (true) {\n  var a;\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{72, "a var hoists out of a block to its function", "function f() {\n  var a;\n  {\n    var a;\n  }\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{73, "a var meets a let in the same body", "function f() {\n  let a;\n  var a;\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{74, "upstream never looks inside a namespace body", "namespace N {\n  var a;\n  {\n    var a;\n  }\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{75, "upstream never looks inside a static block", "class C {\n  static {\n    var a;\n    {\n      var a;\n    }\n  }\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{76, "a catch variable is not hoisted", "try {\n} catch (e) {\n  var e;\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{77, "a let in a block stays in the block", "let a;\nif (true) {\n  let a;\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{78, "a parameter's pattern is not a declaration here", "var a;\nfunction f({ a }: { a: number }) {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{79, "a nested pattern name", "var a;\nvar { b: { a } } = { b: { a: 1 } };\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{80, "two lets in a namespace are clean to upstream", "namespace N {\n  let a;\n  let a;\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{81, "two lets in a static block are clean to upstream", "class C {\n  static {\n    let a;\n    let a;\n  }\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
	{82, "two lets in a block inside a static block report", "class C {\n  static {\n    {\n      let a;\n      let a;\n    }\n  }\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclared", "a"}}},
	{83, "an ESLint global with no lib type", "function toString() {}\n", "", nil, false, nil, []noRedeclareCorpusFinding{{"redeclaredAsBuiltin", "toString"}}},
	{84, "a lib value with no ESLint global is not a builtin", "var top = 0;\nvar window = 1;\n", "", []string{"es2015", "dom"}, false, nil, []noRedeclareCorpusFinding{}},
	{85, "a duplicate in a catch clause's pattern is clean to upstream", "try {\n} catch ({ a, b: a }) {\n}\n", "", nil, false, nil, []noRedeclareCorpusFinding{}},
}
