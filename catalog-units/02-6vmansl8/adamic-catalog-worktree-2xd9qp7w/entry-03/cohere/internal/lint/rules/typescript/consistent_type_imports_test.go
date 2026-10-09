package typescript

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// consistentTypeImportsCase is one imported corpus entry: the source verbatim, the option
// payload verbatim as upstream wrote it, and the message ids the release binary produced
// for that exact pair.
//
// The options travel as raw JSON and go through DecodeConsistentTypeImportsOptions rather
// than being built as a struct, so every case exercises the decoder's non-zero defaults.
// A case written as a struct literal would leave those defaults untested, and all three of
// them are non-zero.
type consistentTypeImportsCase struct {
	source  string
	options string
	ids     []string
}

// decodeConsistentTypeImportsFixtureOptions routes a case's option payload through the rule's
// own exported decoder. An empty payload means the corpus case carried no options, which is
// the bare `"error"` shape and must reach the rule as the documented defaults.
func decodeConsistentTypeImportsFixtureOptions(t *testing.T, payload string) any {
	t.Helper()
	if payload == "" {
		return DefaultConsistentTypeImportsOptions()
	}
	decoded, err := DecodeConsistentTypeImportsOptions(json.RawMessage(payload))
	if err != nil {
		t.Fatalf("decoding fixture options %s: %v", payload, err)
	}
	return decoded
}

// consistentTypeImportsCleanCases are upstream's forty applicable passing inputs, verbatim.
//
// Four further passing cases exist upstream and are deliberately absent: block two of its corpus
// is four `.vue`, `.astro` and `.svelte` files asserting that the rule declines a template
// language it cannot parse. Our tree compiles TypeScript, those extensions never reach a rule
// here, and a fixture naming one would assert the harness rather than the rule.
var consistentTypeImportsCleanCases = []consistentTypeImportsCase{
	{
		source:  "\n              import Foo from 'foo';\n              const foo: Foo = new Foo();\n            ",
		options: "",
	},
	{
		source:  "\n              import foo from 'foo';\n              const foo: foo.Foo = foo.fn();\n            ",
		options: "",
	},
	{
		source:  "\n              import { A, B } from 'foo';\n              const foo: A = B();\n              const bar = new A();\n            ",
		options: "",
	},
	{
		source:  "\n              import Foo from 'foo';\n                  ",
		options: "",
	},
	{
		source:  "\n              import { A, B } from 'foo';\n              const b = B;\n            ",
		options: "",
	},
	{
		source:  "\n              import { A, B, C as c } from 'foo';\n              const d = c;\n            ",
		options: "",
	},
	{
		source:  "\n              import {} from 'foo'; // empty\n            ",
		options: "",
	},
	{
		source:  "\n              let foo: import('foo');\n              let bar: import('foo').Bar;\n            ",
		options: "{\"disallowTypeAnnotations\": false}",
	},
	{
		source:  "\n              import Foo from 'foo';\n              let foo: Foo;\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
	},
	{
		source:  "\n              import * as Type from 'foo' assert { type: 'json' };\n              const a: typeof Type = Type;\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
	},
	{
		source:  "\n              import { type A } from 'foo';\n              type T = A;\n            ",
		options: "",
	},
	{
		source:  "\n              import { type A, B } from 'foo';\n              type T = A;\n              const b = B;\n            ",
		options: "",
	},
	{
		source:  "\n              import { type A, type B } from 'foo';\n              type T = A;\n              type Z = B;\n            ",
		options: "",
	},
	{
		source:  "\n              import { B } from 'foo';\n              import { type A } from 'foo';\n              type T = A;\n              const b = B;\n            ",
		options: "",
	},
	{
		source:  "\n              import { B, type A } from 'foo';\n              type T = A;\n              const b = B;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\"}",
	},
	{
		source:  "\n              import { B } from 'foo';\n              import type A from 'baz';\n              type T = A;\n              const b = B;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\"}",
	},
	{
		source:  "\n              import { type B } from 'foo';\n              import type { A } from 'foo';\n              type T = A;\n              const b = B;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\"}",
	},
	{
		source:  "\n              import { B, type C } from 'foo';\n              import type A from 'baz';\n              type T = A;\n              type Z = C;\n              const b = B;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
	},
	{
		source:  "\n              import { B } from 'foo';\n              import type { A } from 'foo';\n              type T = A;\n              const b = B;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
	},
	{
		source:  "\n              import { B } from 'foo';\n              import { A } from 'foo';\n              type T = A;\n              const b = B;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"no-type-imports\"}",
	},
	{
		source:  "\n              import Type from 'foo';\n\n              export { Type }; // is a value export\n              export default Type; // is a value export\n            ",
		options: "",
	},
	{
		source:  "\n              import type Type from 'foo';\n\n              export { Type }; // is a type-only export\n              export default Type; // is a type-only export\n              export type { Type }; // is a type-only export\n            ",
		options: "",
	},
	{
		source:  "\n              import { Type } from 'foo';\n\n              export { Type }; // is a value export\n              export default Type; // is a value export\n            ",
		options: "",
	},
	{
		source:  "\n              import type { Type } from 'foo';\n\n              export { Type }; // is a type-only export\n              export default Type; // is a type-only export\n              export type { Type }; // is a type-only export\n            ",
		options: "",
	},
	{
		source:  "\n              import * as Type from 'foo';\n\n              export { Type }; // is a value export\n              export default Type; // is a value export\n            ",
		options: "",
	},
	{
		source:  "\n              import type * as Type from 'foo';\n\n              export { Type }; // is a type-only export\n              export default Type; // is a type-only export\n              export type { Type }; // is a type-only export\n            ",
		options: "",
	},
	{
		source:  "\n              import Type from 'foo';\n\n              export { Type }; // is a type-only export\n              export default Type; // is a type-only export\n              export type { Type }; // is a type-only export\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
	},
	{
		source:  "\n              import { Type } from 'foo';\n\n              export { Type }; // is a type-only export\n              export default Type; // is a type-only export\n              export type { Type }; // is a type-only export\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
	},
	{
		source:  "\n              import * as Type from 'foo';\n\n              export { Type }; // is a type-only export\n              export default Type; // is a type-only export\n              export type { Type }; // is a type-only export\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
	},
	{
		source:  "\n              import React from 'react';\n\n              export const ComponentFoo: React.FC = () => {\n                return <div>Foo Foo</div>;\n              };\n            ",
		options: "",
	},
	{
		source:  "\n              import Default, * as Rest from 'module';\n              const a: typeof Default = Default;\n              const b: typeof Rest = Rest;\n            ",
		options: "",
	},
	{
		source:  "\n              import type * as constants from './constants';\n\n              export type Y = {\n                [constants.X]: ReadonlyArray<string>;\n              };\n            ",
		options: "",
	},
	{
		source:  "\n              import A from 'foo';\n              export = A;\n            ",
		options: "",
	},
	{
		source:  "\n              import type A from 'foo';\n              export = A;\n            ",
		options: "",
	},
	{
		source:  "\n              import type A from 'foo';\n              export = {} as A;\n            ",
		options: "",
	},
	{
		source:  "\n              import { type A } from 'foo';\n              export = {} as A;\n            ",
		options: "",
	},
	{
		source:  "\n              import type T from 'mod';\n              const x = T;\n            ",
		options: "",
	},
	{
		source:  "\n              import type { T } from 'mod';\n              const x = T;\n            ",
		options: "",
	},
	{
		source:  "\n              import { type T } from 'mod';\n              const x = T;\n            ",
		options: "",
	},
	{
		source:  "import { Bar } from './bar';\nexport type { Baz } from './baz';\n\nexport class Foo extends Bar {}\n",
		options: "",
	},
}

// consistentTypeImportsReportingCases are upstream's sixty-four failing inputs, verbatim, each
// with the message ids the release binary produced for it under that case's own options.
//
// The counts are a measurement rather than a reconstruction. The snapshot holds seventy-nine
// diagnostics against sixty-four inputs, so a fixture asserting one finding per input would be
// wrong on ten of them, and aligning snapshot text to inputs by line is ambiguous here because
// one diagnostic spans lines. Every case below was instead run through the release binary with
// its own options and its findings read off, which reproduced the snapshot total exactly and
// agreed with upstream's pass/fail labelling on all one hundred and four applicable cases.
var consistentTypeImportsReportingCases = []consistentTypeImportsCase{
	{
		source:  "\n              import Foo from 'foo';\n              let foo: Foo;\n              type Bar = Foo;\n              interface Baz {\n                foo: Foo;\n              }\n              function fn(a: Foo): Foo {}\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import Foo from 'foo';\n              let foo: Foo;\n            ",
		options: "{\"prefer\": \"type-imports\"}",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import Foo from 'foo';\n              let foo: Foo;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import { A, B } from 'foo';\n              let foo: A;\n              let bar: B;\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import { A as a, B as b } from 'foo';\n              let foo: a;\n              let bar: b;\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import Foo from 'foo';\n              type Bar = typeof Foo; // TSTypeQuery\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import foo from 'foo';\n              type Bar = foo.Bar; // TSQualifiedName\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import foo from 'foo';\n              type Baz = (typeof foo.bar)['Baz']; // TSQualifiedName & TSTypeQuery\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import * as A from 'foo';\n              let foo: A.Foo;\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import A, { B } from 'foo';\n              let foo: A;\n              let bar: B;\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import A, {} from 'foo';\n              let foo: A;\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import { A, B } from 'foo';\n              const foo: A = B();\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { A, B, C } from 'foo';\n              const foo: A = B();\n              let bar: C;\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { A, B, C, D } from 'foo';\n              const foo: A = B();\n              type T = { bar: C; baz: D };\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import A, { B, C, D } from 'foo';\n              B();\n              type T = { foo: A; bar: C; baz: D };\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import A, { B } from 'foo';\n              B();\n              type T = A;\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import type Already1Def from 'foo';\n              import type { Already1 } from 'foo';\n              import A, { B } from 'foo';\n              import { C, D, E } from 'bar';\n              import type { Already2 } from 'bar';\n              type T = { b: B; c: C; d: D };\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes", "someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import A, { /* comment */ B } from 'foo';\n              type T = B;\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { A, B, C } from 'foo';\n              import { D, E, F, } from 'bar';\n              type T = A | D;\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes", "someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { A, B, C } from 'foo';\n              import { D, E, F, } from 'bar';\n              type T = B | E;\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes", "someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { A, B, C } from 'foo';\n              import { D, E, F, } from 'bar';\n              type T = C | F;\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes", "someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { Type1, Type2 } from 'named_types';\n              import Type from 'default_type';\n              import * as Types from 'namespace_type';\n              import Default, { Named } from 'default_and_named_type';\n              type T = Type1 | Type2 | Type | Types.A | Default | Named;\n            ",
		options: "",
		ids:     []string{"typeOverValue", "typeOverValue", "typeOverValue", "typeOverValue"},
	},
	{
		source:  "\n              import { Value1, Type1 } from 'named_import';\n              import Type2, { Value2 } from 'default_import';\n              import Value3, { Type3 } from 'default_import2';\n              import Type4, { Type5, Value4 } from 'default_and_named_import';\n              type T = Type1 | Type2 | Type3 | Type4 | Type5;\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes", "someImportsAreOnlyTypes", "someImportsAreOnlyTypes", "someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              let foo: import('foo');\n              let bar: import('foo').Bar;\n            ",
		options: "",
		ids:     []string{"noImportTypeAnnotations", "noImportTypeAnnotations"},
	},
	{
		source:  "\n              let foo: import('foo');\n            ",
		options: "{\"prefer\": \"type-imports\"}",
		ids:     []string{"noImportTypeAnnotations"},
	},
	{
		source:  "\n              import type Foo from 'foo';\n              let foo: Foo;\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
		ids:     []string{"avoidImportType"},
	},
	{
		source:  "\n              import type { Foo } from 'foo';\n              let foo: Foo;\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
		ids:     []string{"avoidImportType"},
	},
	{
		source:  "\n              import Type from 'foo';\n\n              export type { Type }; // is a type-only export\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import { Type } from 'foo';\n\n              export type { Type }; // is a type-only export\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import * as Type from 'foo';\n\n              export type { Type }; // is a type-only export\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import type Type from 'foo';\n\n              export { Type }; // is a type-only export\n              export default Type; // is a type-only export\n              export type { Type }; // is a type-only export\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
		ids:     []string{"avoidImportType"},
	},
	{
		source:  "\n              import type { Type } from 'foo';\n\n              export { Type }; // is a type-only export\n              export default Type; // is a type-only export\n              export type { Type }; // is a type-only export\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
		ids:     []string{"avoidImportType"},
	},
	{
		source:  "\n              import type * as Type from 'foo';\n\n              export { Type }; // is a type-only export\n              export default Type; // is a type-only export\n              export type { Type }; // is a type-only export\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
		ids:     []string{"avoidImportType"},
	},
	{
		source:  "\n              import type /*comment*/ * as AllType from 'foo';\n              import type // comment\n              DefType from 'foo';\n              import type /*comment*/ { Type } from 'foo';\n\n              type T = { a: AllType; b: DefType; c: Type };\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
		ids:     []string{"avoidImportType", "avoidImportType", "avoidImportType"},
	},
	{
		source:  "\n              import Default, * as Rest from 'module';\n              const a: Rest.A = '';\n            ",
		options: "{\"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import Default, * as Rest from 'module';\n              const a: Default = '';\n            ",
		options: "{\"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import Default, * as Rest from 'module';\n              const a: Default = '';\n              const b: Rest.A = '';\n            ",
		options: "{\"prefer\": \"type-imports\"}",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import Default, /*comment*/ * as Rest from 'module';\n              const a: Default = '';\n            ",
		options: "{\"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import Default /*comment1*/, /*comment2*/ { Data } from 'module';\n              const a: Default = '';\n            ",
		options: "{\"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { type A, B } from 'foo';\n              type T = A;\n              const b = B;\n            ",
		options: "{\"prefer\": \"no-type-imports\"}",
		ids:     []string{"avoidImportType"},
	},
	{
		source:  "\n              import { A, B, type C } from 'foo';\n              type T = A | C;\n              const b = B;\n            ",
		options: "{\"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { A, B } from 'foo';\n              let foo: A;\n              let bar: B;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import { A, B } from 'foo';\n\n              let foo: A;\n              B();\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { A, B } from 'foo';\n              type T = A;\n              B();\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { A } from 'foo';\n              import { B } from 'foo';\n              type T = A;\n              type U = B;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"typeOverValue", "typeOverValue"},
	},
	{
		source:  "\n              import { A } from 'foo';\n              import B from 'foo';\n              type T = A;\n              type U = B;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"typeOverValue", "typeOverValue"},
	},
	{
		source:  "\n              import A, { B, C } from 'foo';\n              type T = B;\n              type U = C;\n              A();\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import A, { B, C } from 'foo';\n              type T = B;\n              type U = C;\n              type V = A;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import A, { B, C as D } from 'foo';\n              type T = B;\n              type U = D;\n              type V = A;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import { /* comment */ A, B } from 'foo';\n              type T = A;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { B, /* comment */ A } from 'foo';\n              type T = A;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { A, B, C } from 'foo';\n              import type { D } from 'deez';\n\n              const foo: A = B();\n              let bar: C;\n              let baz: D;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import { A, B, type C } from 'foo';\n              import type { D } from 'deez';\n              const foo: A = B();\n              let bar: C;\n              let baz: D;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import A from 'foo';\n              export = {} as A;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import { A } from 'foo';\n              export = {} as A;\n            ",
		options: "{\"fixStyle\": \"inline-type-imports\", \"prefer\": \"type-imports\"}",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n                import Foo from 'foo';\n                class A {\n                  @deco\n                  foo(): Foo {}\n                }\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n                import Foo from 'foo';\n                class A {\n                  foo(@deco foo: Foo) {}\n                }\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n                import Foo from 'foo';\n                class A {\n                  @deco\n                  set foo(value: Foo) {}\n                }\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n                import Foo from 'foo';\n                class A {\n                  @deco\n                  get foo() {}\n\n                  set foo(value: Foo) {}\n                }\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n                import Foo from 'foo';\n                class A {\n                  @deco\n                  get foo() {}\n\n                  set ['foo'](value: Foo) {}\n                }\n            ",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		source:  "\n              import 'foo';\n              import { Foo, Bar } from 'foo';\n              function test(foo: Foo) {}\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n              import {} from 'foo';\n              import { Foo, Bar } from 'foo';\n              function test(foo: Foo) {}\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n            import type {\n              StorageProvider,\n            } from '../../../fundamentals';\n            import {\n              Config,\n              OnEvent,\n              StorageProviderFactory,\n              URLHelper,\n            } from '../../../fundamentals';\n\n            type A = StorageProvider\n            type B = Config\n            type C = OnEvent\n            type D = URLHelper\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		source:  "\n            import type {\n              TransformResult,\n            } from './plugin'\n            import { defineParallelPlugin,DefineParallelPluginResult } from './plugin'\n\n            type A = TransformResult;\n            type B = DefineParallelPluginResult;\n            const c = defineParallelPlugin()\n            ",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
}

// TestConsistentTypeImportsStaysSilent runs upstream's clean corpus.
//
// These are the cases that catch a port, because each one was added upstream when somebody hit the
// bug it describes. Four of them differ from a reporting case only in what a name resolves to.
func TestConsistentTypeImportsStaysSilent(t *testing.T) {
	t.Parallel()

	for index, testCase := range consistentTypeImportsCleanCases {
		t.Run(fmt.Sprintf("case%02d", index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports,
				"consistent_type_imports.tsx", testCase.source,
				decodeConsistentTypeImportsFixtureOptions(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestConsistentTypeImportsFires runs upstream's reporting corpus, asserting the message ids in
// source order.
func TestConsistentTypeImportsFires(t *testing.T) {
	t.Parallel()

	for index, testCase := range consistentTypeImportsReportingCases {
		t.Run(fmt.Sprintf("case%02d", index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports,
				"consistent_type_imports.tsx", testCase.source,
				decodeConsistentTypeImportsFixtureOptions(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

// consistentTypeImportsMeasuredCases are cases upstream does not write, each measured against
// the release binary before it was recorded here rather than predicted from the source.
//
// The whole set agreed with the prediction, which is worth stating because it is the reason to
// keep them: several pin discriminations that no imported fixture can see, so a port getting
// them backwards passes all one hundred and four upstream cases.
var consistentTypeImportsMeasuredCases = []struct {
	name    string
	source  string
	options string
	ids     []string
}{
	{
		// A class `extends` clause parses to KindExpressionWithTypeArguments here while `implements`
		// parses to KindTypeReference, so the two are separated by the parse shape rather than by a
		// heritage-clause token. Measured on the release binary: this input is SILENT and the
		// `implements` case below REPORTS. Upstream's corpus writes neither, so nothing imported can
		// see a port that treats a superclass as type-only and proposes erasing an import the emitted
		// code needs.
		name:    "classExtendsIsAValueUse",
		source:  "import A from 'foo';\nclass C extends A {}\n",
		options: "",
	},
	{
		// The other half of the pair above, measured the same way.
		name:    "classImplementsIsATypeUse",
		source:  "import A from 'foo';\nclass C implements A {}\n",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		// An interface's `extends` is a type reference where a class's is not, and both are spelled
		// `extends`. Measured; upstream's corpus writes neither form.
		name:    "interfaceExtendsIsATypeUse",
		source:  "import A from 'foo';\ninterface I extends A {}\n",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		// Present upstream as a pass case too, kept here beside its three siblings so the export
		// family reads as one table rather than as four unrelated rows.
		name:    "exportEqualsIsAValueUse",
		source:  "import A from 'foo';\nexport = A;\n",
		options: "",
	},
	{
		// `export default X` and `export = X` share KindExportAssignment in this AST, so one arm
		// answers both. Measured: both silent.
		name:    "exportDefaultIsAValueUse",
		source:  "import Type from 'foo';\nexport default Type;\n",
		options: "",
	},
	{
		// Resolves through GetExportSpecifierLocalTargetSymbol rather than GetSymbolAtLocation, and is
		// then a VALUE use because the export carries no `type` keyword.
		name:    "exportNamedIsAValueUse",
		source:  "import Type from 'foo';\nexport { Type };\n",
		options: "",
	},
	{
		// The keyword sits on the export declaration, two levels above the identifier.
		name:    "exportTypeNamedIsATypeUse",
		source:  "import Type from 'foo';\nexport type { Type };\n",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		// The keyword sits on the specifier instead. Two code paths, one verdict, and upstream's corpus
		// writes only the declaration-level form.
		name:    "exportInlineTypeSpecifierIsATypeUse",
		source:  "import Type from 'foo';\nexport { type Type };\n",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		// IsPartOfTypeNode answers FALSE here, with a parent of KindTypeQuery. Probed directly against
		// the parser. A port resting on that predicate alone goes silent on this and on the qualified
		// name below.
		name:    "typeofIsATypeUse",
		source:  "import Foo from 'foo';\ntype Bar = typeof Foo;\n",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		// Parent is KindQualifiedName and IsPartOfTypeNode again answers false.
		name:    "qualifiedNameLeftIsATypeUse",
		source:  "import foo from 'foo';\ntype Bar = foo.Bar;\n",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		// No JSX, so nothing makes `React` a value, and it reports like any other type-only import.
		// This was a skip by name, oxc's, until phi web found ESLint reporting it. All five React
		// cases here were measured on the installed ESLint and typescript-eslint, 8.67.0.
		name:    "reactDefaultImportWithoutJsxReports",
		source:  "import React from 'react';\ntype T = React.FC;\n",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		name:    "reactNamespaceImportWithoutJsxReports",
		source:  "import * as React from 'react';\ntype T = React.FC;\n",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		// JSX is a value reference to the factory, `React` by default, so the import is a value use.
		name:    "reactWithJsxIsAValueUse",
		source:  "import React from 'react';\ntype T = React.FC;\nconst x = <div />;\n",
		options: "",
	},
	{
		// Only the factory's own name counts: a renamed import is not the pragma, JSX or no JSX.
		name:    "renamedReactWithJsxReports",
		source:  "import Renamed from 'react';\ntype T = Renamed.FC;\nconst x = <div />;\n",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		// The factory is a value use and its type-only sibling still reports, by name.
		name:    "reactValueBesideATypeOnlyName",
		source:  "import React, { FC } from 'react';\nconst C: FC = () => <div />;\n",
		options: "",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		// The distinguishing input for the exemption: identical module, identical use, different local
		// name, and it REPORTS on the release binary. Without this case the exemption could be written
		// as a module-specifier test and every imported fixture would still pass.
		name:    "renamedReactLocalIsNotExempt",
		source:  "import Renamed from 'react';\ntype T = Renamed.FC;\n",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		// The exemption covers the default and namespace forms only. Measured.
		name:    "namedReactSpecifierIsNotExempt",
		source:  "import { React } from 'react';\ntype T = React;\n",
		options: "",
		ids:     []string{"typeOverValue"},
	},
	{
		// A name with no references at all answers no, which is upstream's first line rather than a
		// consequence of the loop. Upstream ships this as a pass case; kept here because the mutant
		// removing that line is otherwise invisible.
		name:    "unusedImportIsSilent",
		source:  "import Foo from 'foo';\n",
		options: "",
	},
	{
		// `import type` cannot carry an import-attributes clause, so upstream withholds the whole
		// declaration finding rather than proposing something unwritable. Measured on the release
		// binary with the older `assert` spelling, which its corpus uses; this writes the modern `with`
		// spelling because our parser accepts both and the decision is on the clause, not the keyword.
		name:    "attributesClauseWithholdsTheDeclarationFinding",
		source:  "import A, { B } from 'foo' with { type: 'json' };\ntype T = A | B;\n",
		options: "",
	},
	{
		// Two anchors under the inverted setting: a type-only declaration reports against the statement
		// and an inline specifier against the specifier. Upstream's corpus exercises them in separate
		// cases, so nothing imported pins that both fire in one file, in source order.
		name:    "noTypeImportsReportsDeclarationAndSpecifierSeparately",
		source:  "import type { A } from 'foo';\nimport { type B } from 'foo';\ntype T = A | B;\n",
		options: "{\"prefer\":\"no-type-imports\"}",
		ids:     []string{"avoidImportType", "avoidImportType"},
	},
	{
		// Pins the default of the option most likely to be got wrong. Every other boolean option in this
		// tree defaults to false; this one defaults to TRUE, so a Go zero value silently disables an arm
		// worth three of the corpus's seventy-nine diagnostics. The empty options payload here is the
		// bare `\"error\"` shape.
		name:    "disallowTypeAnnotationsDefaultsOn",
		source:  "type T = import('foo');\n",
		options: "",
		ids:     []string{"noImportTypeAnnotations"},
	},
	{
		// The inverse of the case above, so the field is pinned in both directions rather than only in
		// the direction the default already produces.
		name:    "disallowTypeAnnotationsCanBeTurnedOff",
		source:  "type T = import('foo');\n",
		options: "{\"disallowTypeAnnotations\":false}",
	},
	{
		// Pins the `prefer` default: under `type-imports` an existing type import is exactly right, and
		// under `no-type-imports` the same input reports. The case below is that inverse.
		name:    "preferDefaultsToTypeImports",
		source:  "import type { A } from 'foo';\ntype T = A;\n",
		options: "",
	},
	{
		// Same source, opposite verdict, which is what makes the default above a measurement rather than
		// a restatement.
		name:    "preferNoTypeImportsInvertsTheSameInput",
		source:  "import type { A } from 'foo';\ntype T = A;\n",
		options: "{\"prefer\":\"no-type-imports\"}",
		ids:     []string{"avoidImportType"},
	},
	{
		// `fixStyle` binds and is decoded, and it changes nothing about what reports because this port
		// proposes no repair. Pinned in both spellings so that a later fix landing here has to decide
		// deliberately whether these verdicts move, rather than discovering it.
		name:    "fixStyleDefaultDoesNotChangeWhatReports",
		source:  "import { A, B } from 'foo';\ntype T = A;\nB();\n",
		options: "{\"fixStyle\":\"inline-type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		// The other spelling of the case above.
		name:    "fixStyleSeparateReportsIdentically",
		source:  "import { A, B } from 'foo';\ntype T = A;\nB();\n",
		options: "{\"fixStyle\":\"separate-type-imports\"}",
		ids:     []string{"someImportsAreOnlyTypes"},
	},
	{
		// An unrecognized enum value keeps the default rather than matching no arm and disabling the
		// rule. Upstream refuses such a configuration outright and no error channel reaches a user
		// here, so the choice is between the documented behavior and a silence that looks like the rule
		// working.
		name:    "unknownEnumValueFallsBackToTheDefault",
		source:  "import type { A } from 'foo';\ntype T = A;\n",
		options: "{\"prefer\":\"nonsense\"}",
	},
}

// TestConsistentTypeImportsMeasuredCases runs the cases written here rather than imported.
func TestConsistentTypeImportsMeasuredCases(t *testing.T) {
	t.Parallel()

	for _, testCase := range consistentTypeImportsMeasuredCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports,
				"consistent_type_imports.tsx", testCase.source,
				decodeConsistentTypeImportsFixtureOptions(t, testCase.options))
			if len(testCase.ids) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

// TestConsistentTypeImportsPointsAtTheRightNode asserts what each finding covers, which no message
// id assertion can see.
//
// Three anchors exist and they are different nodes, so a port pointing all three at the enclosing
// statement passes every id fixture in this file. The text is sliced out of the source with the
// finding's own range and compared against a literal typed here rather than against anything the
// rule computes.
func TestConsistentTypeImportsPointsAtTheRightNode(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name    string
		source  string
		options string
		spans   []string
	}{
		{
			// The declaration, not the clause and not the specifier list.
			name:   "wholeDeclarationForTypeOverValue",
			source: "import Foo from 'foo';\nlet foo: Foo;\n",
			spans:  []string{"import Foo from 'foo';"},
		},
		{
			// Still the whole declaration when only some names are type-only, even though the
			// message names one specifier. Pointing at the specifier would read as more helpful
			// and would be a divergence.
			name:   "wholeDeclarationForSomeImportsAreOnlyTypes",
			source: "import { A, B } from 'foo';\nconst foo: A = B();\n",
			spans:  []string{"import { A, B } from 'foo';"},
		},
		{
			// The `import()` node alone, which is a fragment of the annotation rather than the
			// statement holding it.
			name:   "importTypeNodeOnly",
			source: "type T = import('foo');\n",
			spans:  []string{"import('foo')"},
		},
		{
			// Under the inverted setting the two anchors differ: a declaration for the statement
			// form and a SPECIFIER for the inline one. This is the case that separates them.
			name:    "declarationThenSpecifierUnderNoTypeImports",
			source:  "import type { A } from 'foo';\nimport { type B } from 'foo';\ntype T = A | B;\n",
			options: "{\"prefer\":\"no-type-imports\"}",
			spans:   []string{"import type { A } from 'foo';", "type B"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports,
				"consistent_type_imports.tsx", testCase.source,
				decodeConsistentTypeImportsFixtureOptions(t, testCase.options))
			if len(result.Diagnostics) != len(testCase.spans) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.spans))
			}
			for position, diagnostic := range result.Diagnostics {
				covered := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
				if covered != testCase.spans[position] {
					t.Errorf("finding %d covers %q, want %q", position, covered, testCase.spans[position])
				}
			}
		})
	}
}

// TestConsistentTypeImportsRendersTheNameList asserts the interpolated message text exactly.
//
// This is the one message built with Sprintf, so the id assertion in every other test in this file
// is blind to it: a format string that dropped the names, doubled a separator, or rendered the
// wrong variable would leave the whole suite green. The expected strings below are literals typed
// here and not references to the rule's own message constant, because a constant moves with the
// code under mutation and would agree with a defect.
//
// The three shapes are upstream's own: one name bare, two joined with `and` and no comma, three or
// more with an Oxford comma. The three-name form is the one a reimplementation gets wrong.
func TestConsistentTypeImportsRendersTheNameList(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "oneName",
			source: "import { A, B } from 'foo';\nconst foo: A = B();\n",
			want:   "Imports A are only used as type.",
		},
		{
			name:   "twoNames",
			source: "import { A, B, C } from 'foo';\nconst foo: A = B();\nlet bar: C;\n",
			want:   "Imports A and C are only used as type.",
		},
		{
			name:   "threeNames",
			source: "import { A, B, C, D } from 'foo';\nconst foo: A = B();\ntype T = { bar: C; baz: D };\n",
			want:   "Imports A, C, and D are only used as type.",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports,
				"consistent_type_imports.tsx", testCase.source,
				DefaultConsistentTypeImportsOptions())
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
			}
			if result.Diagnostics[0].Message.Description != testCase.want {
				t.Errorf("rendered %q, want %q",
					result.Diagnostics[0].Message.Description, testCase.want)
			}
		})
	}
}

// TestConsistentTypeImportsRequiresTheTypedHarness asserts the rule declares the checker and goes
// silent without one, so a later revert of either fails loudly rather than quietly.
//
// The source matters here and the first version of this test got it wrong. Only one of the rule's
// three arms consults the checker; the other two, `import()` annotations and the `no-type-imports`
// direction, answer from the syntax alone and fire perfectly well without one. A test written over
// a type-only import alone therefore passes with the nil guard deleted, because that arm goes quiet
// through `GetSymbolAtLocation` returning nil rather than through the guard, and a mutation removing
// the guard survived on exactly that. All three arms are exercised below.
//
// That silence is the more dangerous of the two failure modes and is why the guard stays regardless:
// a nil checker does not panic here, so a rule missing its guard buys a vacuous green rather than an
// obvious crash.
func TestConsistentTypeImportsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !ConsistentTypeImports.NeedsTypeChecker {
		t.Fatal("the rule resolves every name through the checker and must declare it")
	}

	for _, testCase := range []struct {
		name    string
		source  string
		options ConsistentTypeImportsOptions
		typed   []string
	}{
		{
			name:    "typeOverValueArm",
			source:  "import Foo from 'foo';\nlet foo: Foo;\n",
			options: DefaultConsistentTypeImportsOptions(),
			typed:   []string{"typeOverValue"},
		},
		{
			// This arm reads no symbols, so only the guard can silence it.
			name:    "importTypeAnnotationsArm",
			source:  "type T = import('foo');\n",
			options: DefaultConsistentTypeImportsOptions(),
			typed:   []string{"noImportTypeAnnotations"},
		},
		{
			// Nor does this one.
			name:   "noTypeImportsArm",
			source: "import type { A } from 'foo';\ntype T = A;\n",
			options: ConsistentTypeImportsOptions{
				DisallowTypeAnnotations: true,
				FixStyle:                "separate-type-imports",
				Prefer:                  "no-type-imports",
			},
			typed: []string{"avoidImportType"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			typed := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports,
				"consistent_type_imports.tsx", testCase.source, testCase.options)
			rule_testing.ExpectFindings(t, typed, testCase.typed...)

			untyped := rule_testing.RunWithOptions(t, ConsistentTypeImports,
				"consistent_type_imports.tsx", testCase.source, testCase.options)
			rule_testing.ExpectClean(t, untyped)
		})
	}
}

// TestConsistentTypeImportsDecoderDefaults asserts every default through the exported decoder.
//
// All three are non-zero, so a decoder built from rule.DecodeOptionsInto would hand the rule a
// zero-value struct matching no arm. That failure reports nothing while looking correctly wired,
// and every fixture reaching the rule through the decoder would still pass, which is why this
// asserts the decoded values directly.
func TestConsistentTypeImportsDecoderDefaults(t *testing.T) {
	t.Parallel()

	defaults := DefaultConsistentTypeImportsOptions()
	if defaults.Prefer != "type-imports" {
		t.Errorf("prefer defaults to %q, want %q", defaults.Prefer, "type-imports")
	}
	if defaults.FixStyle != "separate-type-imports" {
		t.Errorf("fixStyle defaults to %q, want %q", defaults.FixStyle, "separate-type-imports")
	}
	if !defaults.DisallowTypeAnnotations {
		t.Error("disallowTypeAnnotations defaults to false, want true")
	}

	for _, testCase := range []struct {
		payload string
		want    ConsistentTypeImportsOptions
	}{
		{payload: "{}", want: defaults},
		{
			payload: "{\"prefer\":\"type-imports\",\"fixStyle\":\"separate-type-imports\"}",
			want: ConsistentTypeImportsOptions{
				DisallowTypeAnnotations: true,
				FixStyle:                "separate-type-imports",
				Prefer:                  "type-imports",
			},
		},
		{
			payload: "{\"fixStyle\":\"inline-type-imports\"}",
			want: ConsistentTypeImportsOptions{
				DisallowTypeAnnotations: true,
				FixStyle:                "inline-type-imports",
				Prefer:                  "type-imports",
			},
		},
		{
			payload: "{\"prefer\":\"no-type-imports\"}",
			want: ConsistentTypeImportsOptions{
				DisallowTypeAnnotations: true,
				FixStyle:                "separate-type-imports",
				Prefer:                  "no-type-imports",
			},
		},
		{
			payload: "{\"disallowTypeAnnotations\":false}",
			want: ConsistentTypeImportsOptions{
				DisallowTypeAnnotations: false,
				FixStyle:                "separate-type-imports",
				Prefer:                  "type-imports",
			},
		},
		{
			// An unrecognized value keeps the default rather than disabling the arm.
			payload: "{\"prefer\":\"nonsense\",\"fixStyle\":\"nonsense\"}",
			want:    defaults,
		},
	} {
		decoded, err := DecodeConsistentTypeImportsOptions([]byte(testCase.payload))
		if err != nil {
			t.Fatalf("decoding %s: %v", testCase.payload, err)
		}
		if decoded != any(testCase.want) {
			t.Errorf("decoding %s gave %+v, want %+v", testCase.payload, decoded, testCase.want)
		}
	}
}

// TestConsistentTypeImportsProposesOneEditPerFinding pins the shape the fix engine requires.
//
// The engine flattens a report's fixes into independent proposals and refuses two insertions at one
// point, so upstream's repair proposed as its several primitive edits applied nothing at all. The
// repair here is one replacement per finding, and this is what keeps a later change from splitting
// it back up. The `import()` annotation finding carries no repair, as upstream's does not.
func TestConsistentTypeImportsProposesOneEditPerFinding(t *testing.T) {
	t.Parallel()

	for _, source := range []string{"import Foo from 'foo';\nlet foo: Foo;\n", "import Foo, { Bar } from 'foo';\nlet foo: Foo;\nlet bar: Bar;\n", "import { A, B } from 'foo';\nconst foo: A = B();\n"} {
		result := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports,
			"consistent_type_imports.tsx", source, DefaultConsistentTypeImportsOptions())
		if len(result.Diagnostics) == 0 {
			t.Fatalf("expected a finding on %q", source)
		}
		for _, diagnostic := range result.Diagnostics {
			if len(diagnostic.Fixes) != 1 || len(diagnostic.Suggestions) != 0 {
				t.Errorf("%q proposed %d fixes and %d suggestions, want exactly one fix",
					source, len(diagnostic.Fixes), len(diagnostic.Suggestions))
			}
		}
	}

	// And what that one edit writes, for the commonest shape. The full corpus of repairs, upstream's
	// seventy, is in consistent_type_imports_fix_test.go.
	mixed := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports, "consistent_type_imports.tsx",
		"import Foo, { Bar } from 'foo';\nlet foo: Foo;\nconst bar = Bar;\n", DefaultConsistentTypeImportsOptions())
	rule_testing.ExpectFixedSource(t, mixed, "import type Foo from 'foo';\nimport { Bar } from 'foo';\nlet foo: Foo;\nconst bar = Bar;\n")

	annotation := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports,
		"consistent_type_imports.tsx", "type T = import('foo');\n", DefaultConsistentTypeImportsOptions())
	if len(annotation.Diagnostics) != 1 || len(annotation.Diagnostics[0].Fixes) != 0 {
		t.Errorf("the import() annotation finding should report once with no repair, got %+v", annotation.Diagnostics)
	}
}

// TestConsistentTypeImportsWithNilOptions bypasses the decoder entirely, which every other fixture
// in this file goes through.
//
// A rule configured as a bare "error" is handed nil rather than a struct, and all three of this
// rule's defaults are non-zero, so the zero value is a valid-looking configuration that means
// something else. A rule whose fallback was written as a zero-value struct reports nothing on
// `import()` annotations while still reporting on imports, which is a third of its arms silently off
// on the real tree with every fixture green.
//
// Found by a mutation: rewriting the fallback to a zero-value struct survived all one hundred and
// twenty-eight fixtures, because each one reached the rule through a decoder that never produces
// nil. The `import()` case below is the input that separates the two versions, and the import case
// beside it is the control proving the harness is passing nil rather than failing to run.
func TestConsistentTypeImportsWithNilOptions(t *testing.T) {
	t.Parallel()

	annotations := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports,
		"consistent_type_imports.tsx", "type T = import('foo');\n", nil)
	rule_testing.ExpectFindings(t, annotations, "noImportTypeAnnotations")

	imported := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports,
		"consistent_type_imports.tsx", "import Foo from 'foo';\nlet foo: Foo;\n", nil)
	rule_testing.ExpectFindings(t, imported, "typeOverValue")

	clean := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports,
		"consistent_type_imports.tsx", "import type { Foo } from 'foo';\nlet foo: Foo;\n", nil)
	rule_testing.ExpectClean(t, clean)
}

// TestConsistentTypeImportsReadsTheJsxFactoryFromTheProgram covers the pragma when `jsxFactory` names
// one, which is how the parser resolves it when a project is configured: the first identifier of the
// factory, so `h` for preact, and `React` stops being special.
func TestConsistentTypeImportsReadsTheJsxFactoryFromTheProgram(t *testing.T) {
	t.Parallel()

	withFactory := func(directory string) {
		config := `{"compilerOptions": {"strict": true, "target": "ES2022", "lib": ["ES2022"], ` +
			`"moduleDetection": "force", "types": [], "jsx": "react", "jsxFactory": "h.createElement", ` +
			`"jsxFragmentFactory": "Fragment"}, ` +
			`"include": ["**/*.ts", "**/*.tsx"]}`
		if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(config), 0o644); err != nil {
			t.Fatalf("writing the tsconfig: %v", err)
		}
	}
	cases := []struct {
		name   string
		source string
		ids    []string
	}{
		{"the factory's binding is a value use", "import { h } from 'preact';\ntype T = h.JSX.Element;\nconst x = <div />;\n", nil},
		{"React is an ordinary name under another factory", "import React from 'react';\ntype T = React.FC;\nconst x = <div />;\n", []string{"typeOverValue"}},
		// The fragment factory is a value use only where a fragment is written. Both measured on the
		// installed ESLint with `jsxPragma: 'h'` and `jsxFragmentName: 'Fragment'`.
		{"a fragment makes the fragment factory a value use", "import { h, Fragment } from 'preact';\ntype T = Fragment;\nconst x = <></>;\n", nil},
		{"without a fragment the fragment factory is a type use", "import { h, Fragment } from 'preact';\ntype T = Fragment;\nconst x = <div />;\n", []string{"someImportsAreOnlyTypes"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFilesWithSetupAndOptions(t, ConsistentTypeImports,
				map[string]string{"/repository/source/Factory.tsx": testCase.source}, "/repository/source/Factory.tsx",
				DefaultConsistentTypeImportsOptions(), withFactory)
			if len(testCase.ids) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}
