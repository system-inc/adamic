package typescript

// Code generated from typescript-eslint v8.71.0's tests/rules/no-misused-promises.test.ts and edge rows. DO NOT EDIT BY HAND.
//
// The 34 upstream rows are every row in the test file whose options set checksConditionals, 22 valid and 12 invalid, extracted by loading it with
// RuleTester stubbed. 17 edge rows were added for what they leave open. Before recording, every upstream row
// was linted typed over its own bytes, as upstream's typed RuleTester runs it, and agreed with upstream's own
// assertions: ids, lines and columns. Every row was then replayed against the INSTALLED rule
// (@typescript-eslint/eslint-plugin 8.71.0, one program per row) under noMisusedPromisesConditionalsCorpusTsconfig, on the
// bytes this harness writes: the source trimmed plus one newline, with `export {};` appended to an edge row
// parsed as a module that has no import or export.
// TestNoMisusedPromisesConditionalsUpstreamCorpus replays this table.

// noMisusedPromisesConditionalsCorpusTsconfig is the no-unused-vars options corpus's tsconfig.
const noMisusedPromisesConditionalsCorpusTsconfig = "{\n \"compilerOptions\": {\n  \"strict\": true,\n  \"target\": \"ES2022\",\n  \"lib\": [\n   \"ES2022\",\n   \"ESNext.Disposable\"\n  ],\n  \"types\": []\n },\n \"include\": [\n  \"file.ts\"\n ]\n}"

// noMisusedPromisesConditionalsCorpusFinding is one finding by upstream's message id.
type noMisusedPromisesConditionalsCorpusFinding struct {
	id   string
	text string
}

// noMisusedPromisesConditionalsCorpusCase is one row with the installed rule's verdict. edge is empty for upstream's own rows.
type noMisusedPromisesConditionalsCorpusCase struct {
	index    int
	edge     string
	valid    bool
	source   string
	options  string
	findings []noMisusedPromisesConditionalsCorpusFinding
}

// noMisusedPromisesConditionalsCorpusUpstreamValid and noMisusedPromisesConditionalsCorpusUpstreamInvalid pin how many of upstream's rows
// are here, so a filter cannot empty either direction.
const noMisusedPromisesConditionalsCorpusUpstreamValid, noMisusedPromisesConditionalsCorpusUpstreamInvalid = 22, 12

var noMisusedPromisesConditionalsCorpus = []noMisusedPromisesConditionalsCorpusCase{
	{0, "", true, "if (Promise.resolve()) {\n}\n", "{\"checksConditionals\":false}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{1, "", true, "if (Promise.resolve()) {\n} else if (Promise.resolve()) {\n} else {\n}\n", "{\"checksConditionals\":false}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{2, "", true, "for (let i; Promise.resolve(); i++) {}\n", "{\"checksConditionals\":false}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{3, "", true, "do {} while (Promise.resolve());\n", "{\"checksConditionals\":false}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{4, "", true, "while (Promise.resolve()) {}\n", "{\"checksConditionals\":false}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{5, "", true, "Promise.resolve() ? 123 : 456;\n", "{\"checksConditionals\":false}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{6, "", true, "if (!Promise.resolve()) {\n}\n", "{\"checksConditionals\":false}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{7, "", true, "Promise.resolve() || false;\n", "{\"checksConditionals\":false}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{8, "", true, "(true && Promise.resolve()) || false;\n", "{\"checksConditionals\":false}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{9, "", true, "declare const f: () => boolean | Promise<boolean>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"none\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{10, "", true, "declare const f: () => boolean | Promise<void>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"none\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{11, "", true, "declare const f: () => boolean | Promise<void>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{12, "", true, "type MyUnion = number | string | undefined;\ntype PromiseUnion = string | number;\ndeclare const f: () => MyUnion | Promise<PromiseUnion>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{13, "", true, "declare const f: () => number | string | Promise<number> | Promise<boolean>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{14, "", true, "declare const f: () => string[] | Promise<number[]>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{15, "", true, "declare const f: () => [number, string] | Promise<[string, number]>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{16, "", true, "declare const f: () => (() => void) | Promise<() => number>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{17, "", true, "interface MyInterface {\n  name: string;\n}\n\ninterface PromiseInterface {\n  age: string;\n}\n\ndeclare const f: () => MyInterface | Promise<PromiseInterface>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{18, "", true, "type Recursive = Promise<Recursive>;\ndeclare const f: () => boolean | Recursive;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{19, "", true, "declare const f: boolean | string;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"all\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{20, "", true, "const fn: () => Promise<boolean> | boolean = () => Promise.resolve(true);\nif (await fn()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"all\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{21, "", true, "type MyUnion = number | string;\ntype PromiseUnion = string | number | undefined;\ndeclare const f: () => MyUnion | Promise<PromiseUnion>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{22, "", false, "const tuple: [number, number, number] = [1, 2, 3];\ntuple.find(() => Promise.resolve(false));\n", "{\"checksConditionals\":true}", []noMisusedPromisesConditionalsCorpusFinding{{"predicate", "() => Promise.resolve(false)"}}},
	{23, "", false, "declare const f: () => number | Promise<number>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "f()"}}},
	{24, "", false, "declare const f: () => number | string | Promise<number> | Promise<string>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "f()"}}},
	{25, "", false, "type MyUnion = number | string | undefined;\ndeclare const f: () => MyUnion | Promise<MyUnion>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "f()"}}},
	{26, "", false, "declare const f: () => string[] | Promise<string[]>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "f()"}}},
	{27, "", false, "declare const f: () => [number, string] | Promise<[number, string]>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "f()"}}},
	{28, "", false, "declare const f: () => (() => void) | Promise<() => void>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "f()"}}},
	{29, "", false, "interface MyInterface {\n  name: string;\n}\ndeclare const f: () => MyInterface | Promise<MyInterface>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "f()"}}},
	{30, "", false, "type MyType<T> = { value: T };\ndeclare const f: () => MyType<number> | Promise<MyType<number>>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "f()"}}},
	{31, "", false, "type T = number | string | undefined;\ndeclare const f: () => T | Promise<T>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "f()"}}},
	{32, "", false, "type MyType<T> = { value: T };\ntype PromiseType<T> = { value: T };\ndeclare const f: () => MyType<number> | Promise<PromiseType<number>>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "f()"}}},
	{33, "", false, "declare const f: () => boolean | Promise<void>;\nif (f()) {\n}\n", "{\"checksConditionals\":{\"flagUnions\":\"all\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "f()"}}},
	{34, "an empty object is none", false, "declare const f: () => Promise<number> | number;\nif (f()) {}\nexport {};\n", "{\"checksConditionals\":{}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{35, "true is none", false, "declare const f: () => Promise<number> | number;\nif (f()) {}\nexport {};\n", "{\"checksConditionals\":true}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{36, "none still reports an always-thenable value", false, "declare const p: Promise<number>;\nif (p) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"none\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "p"}}},
	{37, "all reports a union with undefined", false, "declare const p: Promise<number> | undefined;\nif (p) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"all\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "p"}}},
	{38, "strict skips a union with undefined", false, "declare const p: Promise<number> | undefined;\nif (p) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{39, "strict reports a matching union through a logical test", false, "declare const p: Promise<string> | string;\ndeclare const q: boolean;\nif (q && p) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "p"}}},
	{40, "strict on a union awaited type", false, "declare const p: Promise<string | number> | string | number;\nwhile (p) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "p"}}},
	{41, "strict on a union awaited type, one member short", false, "declare const p: Promise<string | number> | string;\nwhile (p) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{42, "all under negation and a ternary", false, "declare const p: Promise<number> | null;\nconst a = !p;\nconst b = p ? 1 : 2;\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"all\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "p"}, {"conditional", "p"}}},
	{43, "all on two promise types", false, "declare const p: Promise<number> | PromiseLike<string> | 0;\ndo {} while (p);\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"all\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "p"}}},
	{44, "strict on two promise types matching", false, "declare const p: Promise<number> | PromiseLike<string> | number | string;\nfor (; p; ) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "p"}}},
	{45, "all keeps predicates", false, "[1, 2].filter(() => Promise.resolve(true));\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"all\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"predicate", "() => Promise.resolve(true)"}}},
	{46, "all on any", false, "declare const p: any;\nif (p) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"all\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{47, "strict on any in a union", false, "declare const p: Promise<any> | any;\nif (p) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{48, "all on an always-thenable value reports once", false, "declare const p: Promise<number>;\nif (p) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"all\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "p"}}},
	{49, "strict needs each member assignable both ways", false, "declare const p: Promise<{ a: number }> | { a: number; b?: string } | { a: number; c: string };\nif (p) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{}},
	{50, "strict with every member assignable both ways", false, "declare const p: Promise<{ a: number }> | { a: number; b?: string };\nif (p) {}\nexport {};\n", "{\"checksConditionals\":{\"flagUnions\":\"strict\"}}", []noMisusedPromisesConditionalsCorpusFinding{{"conditional", "p"}}},
}
