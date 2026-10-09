package typescript

// consistentTypeImportsUpstreamFixes is every invalid case in upstream's corpus that carries a repair,
// extracted mechanically from typescript-eslint's tests/rules/consistent-type-imports.test.ts (8.67.0
// checkout, effa652) by transpiling the file and capturing the arrays a stubbed RuleTester was handed,
// so no escape in these strings was typed by hand. The first parser-option run, with decorator
// metadata off.
//
// Each row is upstream's index in its invalid array, the options object it ran with (empty for the
// defaults), the source, and the source after upstream's repair.
var consistentTypeImportsUpstreamFixes = []struct {
	index    int
	options  string
	code     string
	upstream string
}{
	{0, ``, `
import Foo from 'foo';
let foo: Foo;
type Bar = Foo;
interface Baz {
  foo: Foo;
}
function fn(a: Foo): Foo {}
          `, `
import type Foo from 'foo';
let foo: Foo;
type Bar = Foo;
interface Baz {
  foo: Foo;
}
function fn(a: Foo): Foo {}
          `},
	{1, `{"prefer": "type-imports"}`, `
import Foo from 'foo';
let foo: Foo;
          `, `
import type Foo from 'foo';
let foo: Foo;
          `},
	{2, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import Foo from 'foo';
let foo: Foo;
          `, `
import type Foo from 'foo';
let foo: Foo;
          `},
	{3, ``, `
import { A, B } from 'foo';
let foo: A;
let bar: B;
          `, `
import type { A, B } from 'foo';
let foo: A;
let bar: B;
          `},
	{4, ``, `
import { A as a, B as b } from 'foo';
let foo: a;
let bar: b;
          `, `
import type { A as a, B as b } from 'foo';
let foo: a;
let bar: b;
          `},
	{5, ``, `
import Foo from 'foo';
type Bar = typeof Foo; // TSTypeQuery
          `, `
import type Foo from 'foo';
type Bar = typeof Foo; // TSTypeQuery
          `},
	{6, ``, `
import foo from 'foo';
type Bar = foo.Bar; // TSQualifiedName
          `, `
import type foo from 'foo';
type Bar = foo.Bar; // TSQualifiedName
          `},
	{7, ``, `
import foo from 'foo';
type Baz = (typeof foo.bar)['Baz']; // TSQualifiedName & TSTypeQuery
          `, `
import type foo from 'foo';
type Baz = (typeof foo.bar)['Baz']; // TSQualifiedName & TSTypeQuery
          `},
	{8, ``, `
import * as A from 'foo';
let foo: A.Foo;
          `, `
import type * as A from 'foo';
let foo: A.Foo;
          `},
	{9, ``, `
import A, { B } from 'foo';
let foo: A;
let bar: B;
          `, `
import type { B } from 'foo';
import type A from 'foo';
let foo: A;
let bar: B;
          `},
	{10, ``, `
import A, {} from 'foo';
let foo: A;
          `, `
import type A from 'foo';
let foo: A;
          `},
	{11, ``, `
import { A, B } from 'foo';
const foo: A = B();
          `, `
import type { A} from 'foo';
import { B } from 'foo';
const foo: A = B();
          `},
	{12, ``, `
import { A, B, C } from 'foo';
const foo: A = B();
let bar: C;
          `, `
import type { A, C } from 'foo';
import { B } from 'foo';
const foo: A = B();
let bar: C;
          `},
	{13, ``, `
import { A, B, C, D } from 'foo';
const foo: A = B();
type T = { bar: C; baz: D };
          `, `
import type { A, C, D } from 'foo';
import { B } from 'foo';
const foo: A = B();
type T = { bar: C; baz: D };
          `},
	{14, ``, `
import A, { B, C, D } from 'foo';
B();
type T = { foo: A; bar: C; baz: D };
          `, `
import type { C, D } from 'foo';
import type A from 'foo';
import { B } from 'foo';
B();
type T = { foo: A; bar: C; baz: D };
          `},
	{15, ``, `
import A, { B } from 'foo';
B();
type T = A;
          `, `
import type A from 'foo';
import { B } from 'foo';
B();
type T = A;
          `},
	{16, ``, `
import type Already1Def from 'foo';
import type { Already1 } from 'foo';
import A, { B } from 'foo';
import { C, D, E } from 'bar';
import type { Already2 } from 'bar';
type T = { b: B; c: C; d: D };
          `, `
import type Already1Def from 'foo';
import type { Already1 , B } from 'foo';
import A from 'foo';
import { E } from 'bar';
import type { Already2 , C, D} from 'bar';
type T = { b: B; c: C; d: D };
          `},
	{17, ``, `
import A, { /* comment */ B } from 'foo';
type T = B;
          `, `
import type { /* comment */ B } from 'foo';
import A from 'foo';
type T = B;
          `},
	{18, ``, `
import { A, B, C } from 'foo';
import { D, E, F, } from 'bar';
type T = A | D;
          `, `
import type { A} from 'foo';
import { B, C } from 'foo';
import type { D} from 'bar';
import { E, F, } from 'bar';
type T = A | D;
          `},
	{19, ``, `
import { A, B, C } from 'foo';
import { D, E, F, } from 'bar';
type T = B | E;
          `, `
import type { B} from 'foo';
import { A, C } from 'foo';
import type { E} from 'bar';
import { D, F, } from 'bar';
type T = B | E;
          `},
	{20, ``, `
import { A, B, C } from 'foo';
import { D, E, F, } from 'bar';
type T = C | F;
          `, `
import type { C } from 'foo';
import { A, B } from 'foo';
import type { F} from 'bar';
import { D, E } from 'bar';
type T = C | F;
          `},
	{21, ``, `
import { Type1, Type2 } from 'named_types';
import Type from 'default_type';
import * as Types from 'namespace_type';
import Default, { Named } from 'default_and_named_type';
type T = Type1 | Type2 | Type | Types.A | Default | Named;
          `, `
import type { Type1, Type2 } from 'named_types';
import type Type from 'default_type';
import type * as Types from 'namespace_type';
import type { Named } from 'default_and_named_type';
import type Default from 'default_and_named_type';
type T = Type1 | Type2 | Type | Types.A | Default | Named;
          `},
	{22, ``, `
import { Value1, Type1 } from 'named_import';
import Type2, { Value2 } from 'default_import';
import Value3, { Type3 } from 'default_import2';
import Type4, { Type5, Value4 } from 'default_and_named_import';
type T = Type1 | Type2 | Type3 | Type4 | Type5;
          `, `
import type { Type1 } from 'named_import';
import { Value1 } from 'named_import';
import type Type2 from 'default_import';
import { Value2 } from 'default_import';
import type { Type3 } from 'default_import2';
import Value3 from 'default_import2';
import type { Type5} from 'default_and_named_import';
import type Type4 from 'default_and_named_import';
import { Value4 } from 'default_and_named_import';
type T = Type1 | Type2 | Type3 | Type4 | Type5;
          `},
	{25, `{"prefer": "no-type-imports"}`, `
import type Foo from 'foo';
let foo: Foo;
          `, `
import Foo from 'foo';
let foo: Foo;
          `},
	{26, `{"prefer": "no-type-imports"}`, `
import type { Foo } from 'foo';
let foo: Foo;
          `, `
import { Foo } from 'foo';
let foo: Foo;
          `},
	{27, ``, `
import Type from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `, `
import type Type from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `},
	{28, ``, `
import { Type } from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `, `
import type { Type } from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `},
	{29, ``, `
import * as Type from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `, `
import type * as Type from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `},
	{30, `{"prefer": "no-type-imports"}`, `
import type Type from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `, `
import Type from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `},
	{31, `{"prefer": "no-type-imports"}`, `
import type { Type } from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `, `
import { Type } from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `},
	{32, `{"prefer": "no-type-imports"}`, `
import type * as Type from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `, `
import * as Type from 'foo';

type T = typeof Type;
type T = typeof Type.foo;
          `},
	{33, ``, `
import Type from 'foo';

export type { Type }; // is a type-only export
          `, `
import type Type from 'foo';

export type { Type }; // is a type-only export
          `},
	{34, ``, `
import { Type } from 'foo';

export type { Type }; // is a type-only export
          `, `
import type { Type } from 'foo';

export type { Type }; // is a type-only export
          `},
	{35, ``, `
import * as Type from 'foo';

export type { Type }; // is a type-only export
          `, `
import type * as Type from 'foo';

export type { Type }; // is a type-only export
          `},
	{36, `{"prefer": "no-type-imports"}`, `
import type Type from 'foo';

export { Type }; // is a type-only export
export default Type; // is a type-only export
export type { Type }; // is a type-only export
          `, `
import Type from 'foo';

export { Type }; // is a type-only export
export default Type; // is a type-only export
export type { Type }; // is a type-only export
          `},
	{37, `{"prefer": "no-type-imports"}`, `
import type { Type } from 'foo';

export { Type }; // is a type-only export
export default Type; // is a type-only export
export type { Type }; // is a type-only export
          `, `
import { Type } from 'foo';

export { Type }; // is a type-only export
export default Type; // is a type-only export
export type { Type }; // is a type-only export
          `},
	{38, `{"prefer": "no-type-imports"}`, `
import type * as Type from 'foo';

export { Type }; // is a type-only export
export default Type; // is a type-only export
export type { Type }; // is a type-only export
          `, `
import * as Type from 'foo';

export { Type }; // is a type-only export
export default Type; // is a type-only export
export type { Type }; // is a type-only export
          `},
	{39, `{"prefer": "no-type-imports"}`, `
import type /*comment*/ * as AllType from 'foo';
import type // comment
DefType from 'foo';
import type /*comment*/ { Type } from 'foo';

type T = { a: AllType; b: DefType; c: Type };
          `, `
import /*comment*/ * as AllType from 'foo';
import // comment
DefType from 'foo';
import /*comment*/ { Type } from 'foo';

type T = { a: AllType; b: DefType; c: Type };
          `},
	{40, `{"prefer": "type-imports"}`, `
import Default, * as Rest from 'module';
const a: Rest.A = '';
          `, `
import type * as Rest from 'module';
import Default from 'module';
const a: Rest.A = '';
          `},
	{41, `{"prefer": "type-imports"}`, `
import Default, * as Rest from 'module';
const a: Default = '';
          `, `
import type Default from 'module';
import * as Rest from 'module';
const a: Default = '';
          `},
	{42, `{"prefer": "type-imports"}`, `
import Default, * as Rest from 'module';
const a: Default = '';
const b: Rest.A = '';
          `, `
import type * as Rest from 'module';
import type Default from 'module';
const a: Default = '';
const b: Rest.A = '';
          `},
	{43, `{"prefer": "type-imports"}`, `
import Default, /*comment*/ * as Rest from 'module';
const a: Default = '';
          `, `
import type Default from 'module';
import /*comment*/ * as Rest from 'module';
const a: Default = '';
          `},
	{44, `{"prefer": "type-imports"}`, `
import Default /*comment1*/, /*comment2*/ { Data } from 'module';
const a: Default = '';
          `, `
import type Default /*comment1*/ from 'module';
import /*comment2*/ { Data } from 'module';
const a: Default = '';
          `},
	{45, `{"prefer": "no-type-imports"}`, `
import { type A, B } from 'foo';
type T = A;
const b = B;
          `, `
import { A, B } from 'foo';
type T = A;
const b = B;
          `},
	{46, `{"prefer": "type-imports"}`, `
import { A, B, type C } from 'foo';
type T = A | C;
const b = B;
          `, `
import type { A} from 'foo';
import { B, type C } from 'foo';
type T = A | C;
const b = B;
          `},
	{47, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import { A, B } from 'foo';
let foo: A;
let bar: B;
          `, `
import { type A, type B } from 'foo';
let foo: A;
let bar: B;
          `},
	{48, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import { A, B } from 'foo';

let foo: A;
B();
          `, `
import { type A, B } from 'foo';

let foo: A;
B();
          `},
	{49, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import { A, B } from 'foo';
type T = A;
B();
          `, `
import { type A, B } from 'foo';
type T = A;
B();
          `},
	{50, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import { A } from 'foo';
import { B } from 'foo';
type T = A;
type U = B;
          `, `
import { type A } from 'foo';
import { type B } from 'foo';
type T = A;
type U = B;
          `},
	{51, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import { A } from 'foo';
import B from 'foo';
type T = A;
type U = B;
          `, `
import { type A } from 'foo';
import type B from 'foo';
type T = A;
type U = B;
          `},
	{52, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import A, { B, C } from 'foo';
type T = B;
type U = C;
A();
          `, `
import A, { type B, type C } from 'foo';
type T = B;
type U = C;
A();
          `},
	{53, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import A, { B, C } from 'foo';
type T = B;
type U = C;
type V = A;
          `, `
import {type B, type C} from 'foo';
import type A from 'foo';
type T = B;
type U = C;
type V = A;
          `},
	{54, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import A, { B, C as D } from 'foo';
type T = B;
type U = D;
type V = A;
          `, `
import {type B, type C as D} from 'foo';
import type A from 'foo';
type T = B;
type U = D;
type V = A;
          `},
	{55, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import { /* comment */ A, B } from 'foo';
type T = A;
          `, `
import { /* comment */ type A, B } from 'foo';
type T = A;
          `},
	{56, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import { B, /* comment */ A } from 'foo';
type T = A;
          `, `
import { B, /* comment */ type A } from 'foo';
type T = A;
          `},
	{57, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import { A, B, C } from 'foo';
import type { D } from 'deez';

const foo: A = B();
let bar: C;
let baz: D;
          `, `
import { type A, B, type C } from 'foo';
import type { D } from 'deez';

const foo: A = B();
let bar: C;
let baz: D;
          `},
	{58, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import { A, B, type C } from 'foo';
import type { D } from 'deez';
const foo: A = B();
let bar: C;
let baz: D;
          `, `
import { type A, B, type C } from 'foo';
import type { D } from 'deez';
const foo: A = B();
let bar: C;
let baz: D;
          `},
	{59, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import A from 'foo';
export = {} as A;
          `, `
import type A from 'foo';
export = {} as A;
          `},
	{60, `{"fixStyle": "inline-type-imports", "prefer": "type-imports"}`, `
import { A } from 'foo';
export = {} as A;
          `, `
import { type A } from 'foo';
export = {} as A;
          `},
	{61, ``, `
import Foo from 'foo';
@deco
class A {
  constructor(foo: Foo) {}
}
          `, `
import type Foo from 'foo';
@deco
class A {
  constructor(foo: Foo) {}
}
          `},
	{62, ``, `
import Foo from 'foo';
class A {
  @deco
  foo: Foo;
}
          `, `
import type Foo from 'foo';
class A {
  @deco
  foo: Foo;
}
          `},
	{63, ``, `
import Foo from 'foo';
class A {
  @deco
  foo(foo: Foo) {}
}
          `, `
import type Foo from 'foo';
class A {
  @deco
  foo(foo: Foo) {}
}
          `},
	{64, ``, `
import Foo from 'foo';
class A {
  @deco
  foo(): Foo {}
}
          `, `
import type Foo from 'foo';
class A {
  @deco
  foo(): Foo {}
}
          `},
	{65, ``, `
import Foo from 'foo';
class A {
  foo(@deco foo: Foo) {}
}
          `, `
import type Foo from 'foo';
class A {
  foo(@deco foo: Foo) {}
}
          `},
	{66, ``, `
import Foo from 'foo';
class A {
  @deco
  set foo(value: Foo) {}
}
          `, `
import type Foo from 'foo';
class A {
  @deco
  set foo(value: Foo) {}
}
          `},
	{67, ``, `
import Foo from 'foo';
class A {
  @deco
  get foo() {}

  set foo(value: Foo) {}
}
          `, `
import type Foo from 'foo';
class A {
  @deco
  get foo() {}

  set foo(value: Foo) {}
}
          `},
	{68, ``, `
import Foo from 'foo';
class A {
  @deco
  get foo() {}

  set ['foo'](value: Foo) {}
}
          `, `
import type Foo from 'foo';
class A {
  @deco
  get foo() {}

  set ['foo'](value: Foo) {}
}
          `},
	{69, ``, `
import * as foo from 'foo';
@deco
class A {
  constructor(foo: foo.Foo) {}
}
          `, `
import type * as foo from 'foo';
@deco
class A {
  constructor(foo: foo.Foo) {}
}
          `},
	{70, ``, `
import 'foo';
import { Foo, Bar } from 'foo';
function test(foo: Foo) {}
          `, `
import 'foo';
import type { Foo} from 'foo';
import { Bar } from 'foo';
function test(foo: Foo) {}
          `},
	{71, ``, `
import {} from 'foo';
import { Foo, Bar } from 'foo';
function test(foo: Foo) {}
          `, `
import {} from 'foo';
import type { Foo} from 'foo';
import { Bar } from 'foo';
function test(foo: Foo) {}
          `},
}
