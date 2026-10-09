package core

// Code generated from typescript-eslint v8.71.0's tests/rules/no-unused-vars/*.test.ts and edge rows. DO NOT EDIT BY HAND.
//
// The 41 upstream rows are every row that sets enableAutofixRemoval and every row whose errors assert
// suggestions, extracted by loading the three test files with RuleTester stubbed. 15 edge rows were added for
// what they leave open. Before recording, every upstream row was linted untyped, as upstream's RuleTester runs it,
// and agreed with upstream's own assertions: ids, lines, columns, each suggestion's id and output, and the fix
// output. Every row was then replayed against the INSTALLED rule (@typescript-eslint/eslint-plugin
// 8.71.0, one program per row) under noUnusedVarsOptionsCorpusTsconfig, a JSX row as file.tsx with its
// pragma and fragment name as jsxFactory and jsxFragmentFactory, on the bytes this harness writes: the source
// trimmed plus one newline. A suggestion's output is the file with that one suggestion applied, and output is
// what one pass of the fixes writes, empty when nothing changes. TestNoUnusedVarsFixUpstreamCorpus replays this.

// noUnusedVarsFixCorpusSuggestion is one suggestion: upstream's message id and the file with it applied.
type noUnusedVarsFixCorpusSuggestion struct {
	id     string
	output string
}

// noUnusedVarsFixCorpusFinding is one finding by upstream's message id, with its suggestions.
type noUnusedVarsFixCorpusFinding struct {
	id          string
	text        string
	suggestions []noUnusedVarsFixCorpusSuggestion
}

// noUnusedVarsFixCorpusCase is one row with the installed rule's verdict. edge is empty for upstream's own rows.
type noUnusedVarsFixCorpusCase struct {
	index              int
	edge               string
	source             string
	options            string
	jsxFactory         string
	jsxFragmentFactory string
	jsx                bool
	findings           []noUnusedVarsFixCorpusFinding
	output             string
}

// noUnusedVarsFixCorpusUpstreamRows pins how many of upstream's rows are here, so a filter cannot empty the table.
const noUnusedVarsFixCorpusUpstreamRows = 41

var noUnusedVarsFixCorpus = []noUnusedVarsFixCorpusCase{
	{0, "", "import x from 'y';\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "x", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", ""}}}}, ""},
	{1, "", "import { ClassDecoratorFactory } from 'decorators';\nexport class Foo {}\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "ClassDecoratorFactory", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "export class Foo {}\n"}}}}, ""},
	{2, "", "import { Foo, Bar } from 'foo';\nfunction baz<Foo>(): Foo {}\nbaz<Bar>();\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Foo", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedVar", "import {  Bar } from 'foo';\nfunction baz<Foo>(): Foo {}\nbaz<Bar>();\n"}}}}, ""},
	{3, "", "import { Nullable } from 'nullable';\nconst a: string = 'hello';\nconsole.log(a);\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Nullable", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "const a: string = 'hello';\nconsole.log(a);\n"}}}}, ""},
	{4, "", "import { Nullable } from 'nullable';\nimport { SomeOther } from 'other';\nconst a: Nullable<string> = 'hello';\nconsole.log(a);\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "SomeOther", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import { Nullable } from 'nullable';\nconst a: Nullable<string> = 'hello';\nconsole.log(a);\n"}}}}, ""},
	{5, "", "import { Nullable } from 'nullable';\nimport { Another } from 'some';\nclass A {\n  do = (a: Nullable) => {\n    console.log(a);\n  };\n}\nnew A();\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Another", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import { Nullable } from 'nullable';\nclass A {\n  do = (a: Nullable) => {\n    console.log(a);\n  };\n}\nnew A();\n"}}}}, ""},
	{6, "", "import { Nullable } from 'nullable';\nimport { Another } from 'some';\nclass A {\n  do(a: Nullable) {\n    console.log(a);\n  }\n}\nnew A();\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Another", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import { Nullable } from 'nullable';\nclass A {\n  do(a: Nullable) {\n    console.log(a);\n  }\n}\nnew A();\n"}}}}, ""},
	{7, "", "import { Nullable } from 'nullable';\nimport { Another } from 'some';\nclass A {\n  do(): Nullable {\n    return null;\n  }\n}\nnew A();\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Another", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import { Nullable } from 'nullable';\nclass A {\n  do(): Nullable {\n    return null;\n  }\n}\nnew A();\n"}}}}, ""},
	{8, "", "import { Nullable } from 'nullable';\nimport { Another } from 'some';\nexport interface A {\n  do(a: Nullable);\n}\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Another", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import { Nullable } from 'nullable';\nexport interface A {\n  do(a: Nullable);\n}\n"}}}}, ""},
	{9, "", "import { Nullable } from 'nullable';\nimport { Another } from 'some';\nexport interface A {\n  other: Nullable;\n}\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Another", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import { Nullable } from 'nullable';\nexport interface A {\n  other: Nullable;\n}\n"}}}}, ""},
	{10, "", "import { Nullable } from 'nullable';\nfunction foo(a: string) {\n  console.log(a);\n}\nfoo();\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Nullable", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "function foo(a: string) {\n  console.log(a);\n}\nfoo();\n"}}}}, ""},
	{11, "", "import { Nullable } from 'nullable';\nfunction foo(): string | null {\n  return null;\n}\nfoo();\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Nullable", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "function foo(): string | null {\n  return null;\n}\nfoo();\n"}}}}, ""},
	{12, "", "import { Nullable } from 'nullable';\nimport { SomeOther } from 'some';\nimport { Another } from 'some';\nclass A extends Nullable {\n  other: Nullable<Another>;\n}\nnew A();\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "SomeOther", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import { Nullable } from 'nullable';\nimport { Another } from 'some';\nclass A extends Nullable {\n  other: Nullable<Another>;\n}\nnew A();\n"}}}}, ""},
	{13, "", "import { Nullable } from 'nullable';\nimport { SomeOther } from 'some';\nimport { Another } from 'some';\nabstract class A extends Nullable {\n  other: Nullable<Another>;\n}\nnew A();\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "SomeOther", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import { Nullable } from 'nullable';\nimport { Another } from 'some';\nabstract class A extends Nullable {\n  other: Nullable<Another>;\n}\nnew A();\n"}}}}, ""},
	{14, "", "import test from 'test';\nimport baz from 'baz';\nexport interface Bar extends baz.test {}\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "test", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import baz from 'baz';\nexport interface Bar extends baz.test {}\n"}}}}, ""},
	{15, "", "import test from 'test';\nimport baz from 'baz';\nexport class Bar implements baz.test {}\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "test", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import baz from 'baz';\nexport class Bar implements baz.test {}\n"}}}}, ""},
	{16, "", "import React from 'react';\nimport { Fragment } from 'react';\n\nexport const ComponentFoo = () => {\n  return <div>Foo Foo</div>;\n};\n", "", "", "", true, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Fragment", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import React from 'react';\n\nexport const ComponentFoo = () => {\n  return <div>Foo Foo</div>;\n};\n"}}}}, ""},
	{17, "", "import React from 'react';\nimport { h } from 'some-other-jsx-lib';\n\nexport const ComponentFoo = () => {\n  return <div>Foo Foo</div>;\n};\n", "", "h", "", true, []noUnusedVarsFixCorpusFinding{{"unusedVar", "React", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "import { h } from 'some-other-jsx-lib';\n\nexport const ComponentFoo = () => {\n  return <div>Foo Foo</div>;\n};\n"}}}}, ""},
	{18, "", "import React from 'react';\n\nexport const ComponentFoo = () => {\n  return <div>Foo Foo</div>;\n};\n", "", "", "", true, []noUnusedVarsFixCorpusFinding{}, ""},
	{19, "", "namespace Foo {\n  export const foo = 1;\n}\nexport namespace Bar {\n  import TheFoo = Foo;\n}\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "TheFoo", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "namespace Foo {\n  export const foo = 1;\n}\nexport namespace Bar {\n}\n"}}}}, ""},
	{20, "", "import * as Unused from 'module';\nexport {};\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "export {};\n"},
	{21, "", "import Unused from 'module';\nexport {};\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "export {};\n"},
	{22, "", "import { Unused } from 'module';\nexport {};\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "export {};\n"},
	{23, "", "import { Unused, Unused2 } from 'module';\nexport {};\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}, {"unusedVar", "Unused2", []noUnusedVarsFixCorpusSuggestion{}}}, "export {};\n"},
	{24, "", "import { Unused, Used } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "import {  Used } from 'module';\nexport { Used };\n"},
	{25, "", "import { Used, Unused } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "import { Used } from 'module';\nexport { Used };\n"},
	{26, "", "import { Used, Unused, } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "import { Used, } from 'module';\nexport { Used };\n"},
	{27, "", "import { Used, Unused, Used2 } from 'module';\nexport { Used, Used2 };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "import { Used, Used2 } from 'module';\nexport { Used, Used2 };\n"},
	{28, "", "import Unused, { Unused2 } from 'module';\nexport {};\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}, {"unusedVar", "Unused2", []noUnusedVarsFixCorpusSuggestion{}}}, "export {};\n"},
	{29, "", "import Unused, { Used } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "import  { Used } from 'module';\nexport { Used };\n"},
	{30, "", "import Used, { Unused } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "import Used from 'module';\nexport { Used };\n"},
	{31, "", "import Used, { Unused, } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "import Used from 'module';\nexport { Used };\n"},
	{32, "", "import Used, { Used2, Unused } from 'module';\nexport { Used, Used2 };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "import Used, { Used2 } from 'module';\nexport { Used, Used2 };\n"},
	{33, "", "import Used, { Unused, Used2 } from 'module';\nexport { Used, Used2 };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "import Used, {  Used2 } from 'module';\nexport { Used, Used2 };\n"},
	{34, "", "import Unused, { Unused2, Used } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}, {"unusedVar", "Unused2", []noUnusedVarsFixCorpusSuggestion{}}}, "import  {  Used } from 'module';\nexport { Used };\n"},
	{35, "", "import Unused, { Used, Unused2 } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}, {"unusedVar", "Unused2", []noUnusedVarsFixCorpusSuggestion{}}}, "import  { Used } from 'module';\nexport { Used };\n"},
	{36, "", "import { Unused as Unused1, Used } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused1", []noUnusedVarsFixCorpusSuggestion{}}}, "import {  Used } from 'module';\nexport { Used };\n"},
	{37, "", "import { Used, Unused as Unused1 } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused1", []noUnusedVarsFixCorpusSuggestion{}}}, "import { Used } from 'module';\nexport { Used };\n"},
	{38, "", "import { Used, Unused as Unused1, } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused1", []noUnusedVarsFixCorpusSuggestion{}}}, "import { Used, } from 'module';\nexport { Used };\n"},
	{39, "", "/* this is an important comment */ import assert from 'assert';\nexport {};\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "assert", []noUnusedVarsFixCorpusSuggestion{}}}, "/* this is an important comment */ \nexport {};\n"},
	{40, "", "import assert from 'assert'; /* this is an important comment */\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "assert", []noUnusedVarsFixCorpusSuggestion{}}}, " /* this is an important comment */\n"},
	{41, "default mode: a wholly unused declaration is a suggestion", "import { Unused } from 'module';\nexport {};\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "export {};\n"}}}}, ""},
	{42, "default mode: one unused named specifier beside a used one", "import { Unused, Used } from 'module';\nexport { Used };\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedVar", "import {  Used } from 'module';\nexport { Used };\n"}}}}, ""},
	{43, "default mode: an unused import equals", "import unused = require('module');\nexport {};\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "unused", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "export {};\n"}}}}, ""},
	{44, "default mode: an unused namespace import", "import * as Unused from 'module';\nexport {};\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "export {};\n"}}}}, ""},
	{45, "default mode: an unused default beside a used named", "import Unused, { Used } from 'module';\nexport { Used };\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedVar", "import  { Used } from 'module';\nexport { Used };\n"}}}}, ""},
	{46, "default mode: unused named beside a used default", "import Used, { Unused } from 'module';\nexport { Used };\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedVar", "import Used from 'module';\nexport { Used };\n"}}}}, ""},
	{47, "autofix: an unused default beside a used named", "import Unused, { Used } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "import  { Used } from 'module';\nexport { Used };\n"},
	{48, "autofix: a comment inside the list", "import { /* keep */ Unused, Used } from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "import { /* keep */  Used } from 'module';\nexport { Used };\n"},
	{49, "autofix: a declaration sharing its line keeps the line", "import { Unused } from 'module'; export const x = 1;\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, " export const x = 1;\n"},
	{50, "autofix: a declaration over several lines", "import {\n  Unused,\n} from 'module';\nexport {};\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "export {};\n"},
	{51, "autofix: three unused in one declaration", "import { A, B, C } from 'module';\nexport {};\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "A", []noUnusedVarsFixCorpusSuggestion{}}, {"unusedVar", "B", []noUnusedVarsFixCorpusSuggestion{}}, {"unusedVar", "C", []noUnusedVarsFixCorpusSuggestion{}}}, "export {};\n"},
	{52, "autofix off explicitly is the default", "import { Unused } from 'module';\nexport {};\n", "{\"enableAutofixRemoval\":{\"imports\":false}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{{"removeUnusedImportDeclaration", "export {};\n"}}}}, ""},
	{53, "autofix on a used ignored import", "import { _a } from 'module';\nconsole.log(_a);\n", "{\"enableAutofixRemoval\":{\"imports\":true},\"reportUsedIgnorePattern\":true,\"varsIgnorePattern\":\"^_\"}", "", "", false, []noUnusedVarsFixCorpusFinding{{"usedIgnoredVar", "_a", []noUnusedVarsFixCorpusSuggestion{}}}, "console.log(_a);\n"},
	{54, "default mode: an unused variable gets no suggestion", "const unused = 1;\nexport {};\n", "", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "unused", []noUnusedVarsFixCorpusSuggestion{}}}, ""},
	{55, "autofix: an unused namespace beside a used default", "import Used, * as Unused from 'module';\nexport { Used };\n", "{\"enableAutofixRemoval\":{\"imports\":true}}", "", "", false, []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", []noUnusedVarsFixCorpusSuggestion{}}}, "export { Used };\n"},
}
