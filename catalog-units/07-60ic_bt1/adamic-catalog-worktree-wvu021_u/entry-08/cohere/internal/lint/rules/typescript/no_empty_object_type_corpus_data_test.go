package typescript

// Code generated from typescript-eslint v8.71.0's tests/rules/no-empty-object-type.test.ts and edge rows. DO NOT EDIT BY HAND.
//
// The 42 upstream rows are every row in the test file, 14 valid and 28 invalid, extracted by loading it with
// RuleTester stubbed. 18 edge rows were added for what they leave open. Before recording, every upstream row
// was linted untyped, as upstream's RuleTester runs it, and agreed with upstream's own assertions: ids, lines,
// columns, and each suggestion's id and output. Every row was then replayed against the INSTALLED rule
// (@typescript-eslint/eslint-plugin 8.71.0, one program per row) under noEmptyObjectTypeCorpusTsconfig, on the
// bytes this harness writes: the source trimmed plus one newline, with `export {};` appended to an edge row
// parsed as a module that has no import or export. A suggestion's output is the file with that one suggestion
// applied. TestNoEmptyObjectTypeUpstreamCorpus replays this table.

// noEmptyObjectTypeCorpusTsconfig is the no-unused-vars options corpus's tsconfig.
const noEmptyObjectTypeCorpusTsconfig = "{\n \"compilerOptions\": {\n  \"strict\": true,\n  \"target\": \"ES2022\",\n  \"lib\": [\n   \"ES2022\",\n   \"ESNext.Disposable\"\n  ],\n  \"types\": []\n },\n \"include\": [\n  \"file.ts\"\n ]\n}"

// noEmptyObjectTypeCorpusSuggestion is one suggestion: upstream's message id and the file with it applied.
type noEmptyObjectTypeCorpusSuggestion struct {
	id     string
	output string
}

// noEmptyObjectTypeCorpusFinding is one finding by upstream's message id, with its suggestions.
type noEmptyObjectTypeCorpusFinding struct {
	id          string
	text        string
	suggestions []noEmptyObjectTypeCorpusSuggestion
}

// noEmptyObjectTypeCorpusCase is one row with the installed rule's verdict. edge is empty for upstream's own rows.
type noEmptyObjectTypeCorpusCase struct {
	index    int
	edge     string
	valid    bool
	source   string
	options  string
	findings []noEmptyObjectTypeCorpusFinding
}

// noEmptyObjectTypeCorpusUpstreamValid and noEmptyObjectTypeCorpusUpstreamInvalid pin how many of upstream's rows
// are here, so a filter cannot empty either direction.
const noEmptyObjectTypeCorpusUpstreamValid, noEmptyObjectTypeCorpusUpstreamInvalid = 14, 28

var noEmptyObjectTypeCorpus = []noEmptyObjectTypeCorpusCase{
	{0, "", true, "interface Base {\n  name: string;\n}\n", "", []noEmptyObjectTypeCorpusFinding{}},
	{1, "", true, "interface Base {\n  name: string;\n}\n\ninterface Derived {\n  age: number;\n}\n\n// valid because extending multiple interfaces can be used instead of a union type\ninterface Both extends Base, Derived {}\n", "", []noEmptyObjectTypeCorpusFinding{}},
	{2, "", true, "interface Base {}\n", "{\"allowInterfaces\":\"always\"}", []noEmptyObjectTypeCorpusFinding{}},
	{3, "", true, "interface Base {\n  name: string;\n}\n\ninterface Derived extends Base {}\n", "{\"allowInterfaces\":\"with-single-extends\"}", []noEmptyObjectTypeCorpusFinding{}},
	{4, "", true, "interface Base {\n  props: string;\n}\n\ninterface Derived extends Base {}\n\nclass Derived {}\n", "{\"allowInterfaces\":\"with-single-extends\"}", []noEmptyObjectTypeCorpusFinding{}},
	{5, "", true, "let value: object;\n", "", []noEmptyObjectTypeCorpusFinding{}},
	{6, "", true, "let value: Object;\n", "", []noEmptyObjectTypeCorpusFinding{}},
	{7, "", true, "let value: { inner: true };\n", "", []noEmptyObjectTypeCorpusFinding{}},
	{8, "", true, "type MyNonNullable<T> = T & {};\n", "", []noEmptyObjectTypeCorpusFinding{}},
	{9, "", true, "type Base = {};\n", "{\"allowObjectTypes\":\"always\"}", []noEmptyObjectTypeCorpusFinding{}},
	{10, "", true, "type Base = {};\n", "{\"allowWithName\":\"Base\"}", []noEmptyObjectTypeCorpusFinding{}},
	{11, "", true, "type BaseProps = {};\n", "{\"allowWithName\":\"Props$\"}", []noEmptyObjectTypeCorpusFinding{}},
	{12, "", true, "interface Base {}\n", "{\"allowWithName\":\"Base\"}", []noEmptyObjectTypeCorpusFinding{}},
	{13, "", true, "interface BaseProps {}\n", "{\"allowWithName\":\"Props$\"}", []noEmptyObjectTypeCorpusFinding{}},
	{14, "", false, "interface Base {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "type Base = object\n"}, {"replaceEmptyInterface", "type Base = unknown\n"}}}}},
	{15, "", false, "interface Base {}\n", "{\"allowInterfaces\":\"never\"}", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "type Base = object\n"}, {"replaceEmptyInterface", "type Base = unknown\n"}}}}},
	{16, "", false, "interface Base {\n  props: string;\n}\n\ninterface Derived extends Base {}\n\nclass Other {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Derived", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterfaceWithSuper", "interface Base {\n  props: string;\n}\n\ntype Derived = Base\n\nclass Other {}\n"}}}}},
	{17, "", false, "interface Base {\n  props: string;\n}\n\ninterface Derived extends Base {}\n\nclass Derived {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Derived", []noEmptyObjectTypeCorpusSuggestion{}}}},
	{18, "", false, "interface Base {\n  props: string;\n}\n\ninterface Derived extends Base {}\n\nconst derived = class Derived {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Derived", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterfaceWithSuper", "interface Base {\n  props: string;\n}\n\ntype Derived = Base\n\nconst derived = class Derived {};\n"}}}}},
	{19, "", false, "interface Base {}\n\ninterface Base {\n  name: string;\n}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{}}}},
	{20, "", false, "interface Base {}\n\ninterface Base {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{}}, {"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{}}}},
	{21, "", false, "interface Base {}\n\nfunction foo() {\n  interface Base {\n    name: string;\n  }\n}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "type Base = object\n\nfunction foo() {\n  interface Base {\n    name: string;\n  }\n}\n"}, {"replaceEmptyInterface", "type Base = unknown\n\nfunction foo() {\n  interface Base {\n    name: string;\n  }\n}\n"}}}}},
	{22, "", false, "interface Base {\n  props: string;\n}\n\ninterface Derived extends Base {}\n\ninterface Derived {\n  name: string;\n}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Derived", []noEmptyObjectTypeCorpusSuggestion{}}}},
	{23, "", false, "export default interface Base {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{}}}},
	{24, "", false, "export default interface Derived extends Base {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Derived", []noEmptyObjectTypeCorpusSuggestion{}}}},
	{25, "", false, "export interface Base {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "export type Base = object\n"}, {"replaceEmptyInterface", "export type Base = unknown\n"}}}}},
	{26, "", false, "interface Base {\n  name: string;\n}\n\ninterface Derived extends Base {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Derived", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterfaceWithSuper", "interface Base {\n  name: string;\n}\n\ntype Derived = Base\n"}}}}},
	{27, "", false, "interface Base extends Array<number> {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterfaceWithSuper", "type Base = Array<number>\n"}}}}},
	{28, "", false, "interface Base extends Array<number | {}> {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterfaceWithSuper", "type Base = Array<number | {}>\n"}}}, {"noEmptyObject", "{}", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyObjectType", "interface Base extends Array<number | object> {}\n"}, {"replaceEmptyObjectType", "interface Base extends Array<number | unknown> {}\n"}}}}},
	{29, "", false, "interface Derived {\n  property: string;\n}\ninterface Base extends Array<Derived> {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterfaceWithSuper", "interface Derived {\n  property: string;\n}\ntype Base = Array<Derived>\n"}}}}},
	{30, "", false, "type R = Record<string, unknown>;\ninterface Base extends R {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterfaceWithSuper", "type R = Record<string, unknown>;\ntype Base = R\n"}}}}},
	{31, "", false, "interface Base<T> extends Derived<T> {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterfaceWithSuper", "type Base<T> = Derived<T>\n"}}}}},
	{32, "", false, "declare namespace BaseAndDerived {\n  type Base = typeof base;\n  export interface Derived extends Base {}\n}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Derived", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterfaceWithSuper", "declare namespace BaseAndDerived {\n  type Base = typeof base;\n  export type Derived = Base\n}\n"}}}}},
	{33, "", false, "type Base = {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyObject", "{}", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyObjectType", "type Base = object;\n"}, {"replaceEmptyObjectType", "type Base = unknown;\n"}}}}},
	{34, "", false, "type Base = {};\n", "{\"allowObjectTypes\":\"never\"}", []noEmptyObjectTypeCorpusFinding{{"noEmptyObject", "{}", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyObjectType", "type Base = object;\n"}, {"replaceEmptyObjectType", "type Base = unknown;\n"}}}}},
	{35, "", false, "let value: {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyObject", "{}", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyObjectType", "let value: object;\n"}, {"replaceEmptyObjectType", "let value: unknown;\n"}}}}},
	{36, "", false, "let value: {};\n", "{\"allowObjectTypes\":\"never\"}", []noEmptyObjectTypeCorpusFinding{{"noEmptyObject", "{}", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyObjectType", "let value: object;\n"}, {"replaceEmptyObjectType", "let value: unknown;\n"}}}}},
	{37, "", false, "let value: {/* ... */};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyObject", "{/* ... */}", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyObjectType", "let value: object;\n"}, {"replaceEmptyObjectType", "let value: unknown;\n"}}}}},
	{38, "", false, "type MyUnion<T> = T | {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyObject", "{}", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyObjectType", "type MyUnion<T> = T | object;\n"}, {"replaceEmptyObjectType", "type MyUnion<T> = T | unknown;\n"}}}}},
	{39, "", false, "type Base = {} | null;\n", "{\"allowWithName\":\"Base\"}", []noEmptyObjectTypeCorpusFinding{{"noEmptyObject", "{}", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyObjectType", "type Base = object | null;\n"}, {"replaceEmptyObjectType", "type Base = unknown | null;\n"}}}}},
	{40, "", false, "type Base = {};\n", "{\"allowWithName\":\"Mismatch\"}", []noEmptyObjectTypeCorpusFinding{{"noEmptyObject", "{}", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyObjectType", "type Base = object;\n"}, {"replaceEmptyObjectType", "type Base = unknown;\n"}}}}},
	{41, "", false, "interface Base {}\n", "{\"allowWithName\":\".*Props$\"}", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "type Base = object\n"}, {"replaceEmptyInterface", "type Base = unknown\n"}}}}},
	{42, "a parenthesized empty literal in an intersection", false, "type A = B & ({});\ntype B = { b: 1 };\nexport {};\n", "", []noEmptyObjectTypeCorpusFinding{}},
	{43, "a parenthesized empty literal named by allowWithName", false, "type BaseProps = ({});\nexport {};\n", "{\"allowWithName\":\"Props$\"}", []noEmptyObjectTypeCorpusFinding{}},
	{44, "declare interface", false, "declare interface Base {}\nexport {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "type Base = object\nexport {};\n"}, {"replaceEmptyInterface", "type Base = unknown\nexport {};\n"}}}}},
	{45, "export declare interface", false, "export declare interface Base {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "export type Base = object\n"}, {"replaceEmptyInterface", "export type Base = unknown\n"}}}}},
	{46, "an empty interface with type parameters merged with a class", false, "interface Base<T> {}\nclass Base<T> {}\nexport {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "type Base<T> = object\nclass Base<T> {}\nexport {};\n"}, {"replaceEmptyInterface", "type Base<T> = unknown\nclass Base<T> {}\nexport {};\n"}}}}},
	{47, "an empty interface merged with a namespace", false, "interface Base {}\nnamespace Base {}\nexport {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "type Base = object\nnamespace Base {}\nexport {};\n"}, {"replaceEmptyInterface", "type Base = unknown\nnamespace Base {}\nexport {};\n"}}}}},
	{48, "an empty interface merged across blocks", false, "interface Base {}\n{\n  class Base {}\n}\nexport {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "type Base = object\n{\n  class Base {}\n}\nexport {};\n"}, {"replaceEmptyInterface", "type Base = unknown\n{\n  class Base {}\n}\nexport {};\n"}}}}},
	{49, "an empty interface in a namespace merged in the same namespace", false, "namespace N {\n  interface Base {}\n  class Base {}\n}\nexport {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{}}}},
	{50, "an empty interface augmenting the lib in a script", false, "interface Window {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Window", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "type Window = object\n"}, {"replaceEmptyInterface", "type Window = unknown\n"}}}}},
	{51, "an empty interface extending two, with-single-extends", false, "interface A { a: 1 }\ninterface B { b: 1 }\ninterface C extends A, B {}\nexport {};\n", "{\"allowInterfaces\":\"with-single-extends\"}", []noEmptyObjectTypeCorpusFinding{}},
	{52, "allowWithName is a regular expression, not a name", false, "interface XBaseY {}\ntype XBaseZ = {};\nexport {};\n", "{\"allowWithName\":\"Base\"}", []noEmptyObjectTypeCorpusFinding{}},
	{53, "allowWithName with a unicode escape", false, "type Café = {};\nexport {};\n", "{\"allowWithName\":\"\\\\u{E9}$\"}", []noEmptyObjectTypeCorpusFinding{}},
	{54, "an empty interface with a type parameter and a comment", false, "interface Base<T extends string = 'a'> /* note */ {}\nexport {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterface", "type Base<T extends string = 'a'> = object\nexport {};\n"}, {"replaceEmptyInterface", "type Base<T extends string = 'a'> = unknown\nexport {};\n"}}}}},
	{55, "an empty literal as a type argument", false, "let value: Array<{}>;\nexport {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyObject", "{}", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyObjectType", "let value: Array<object>;\nexport {};\n"}, {"replaceEmptyObjectType", "let value: Array<unknown>;\nexport {};\n"}}}}},
	{56, "an empty literal in a mapped type's value", false, "type M = { [K in 'a']: {} };\nexport {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyObject", "{}", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyObjectType", "type M = { [K in 'a']: object };\nexport {};\n"}, {"replaceEmptyObjectType", "type M = { [K in 'a']: unknown };\nexport {};\n"}}}}},
	{57, "declare interface extending one", false, "interface Base {\n  name: string;\n}\ndeclare interface Derived extends Base {}\nexport {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Derived", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterfaceWithSuper", "interface Base {\n  name: string;\n}\ntype Derived = Base\nexport {};\n"}}}}},
	{58, "export declare interface with type parameters extending one", false, "interface Base<T> {\n  value: T;\n}\nexport declare interface Derived<T> extends Base<T> {}\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterfaceWithSuper", "Derived", []noEmptyObjectTypeCorpusSuggestion{{"replaceEmptyInterfaceWithSuper", "interface Base<T> {\n  value: T;\n}\nexport type Derived<T> = Base<T>\n"}}}}},
	{59, "an empty interface in a switch case merged in another case", false, "switch (1 as number) {\n  case 1:\n    interface Base {}\n    break;\n  default:\n    class Base {}\n}\nexport {};\n", "", []noEmptyObjectTypeCorpusFinding{{"noEmptyInterface", "Base", []noEmptyObjectTypeCorpusSuggestion{}}}},
}
