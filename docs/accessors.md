# Accessor views after integration

The current accessor implementation is main's descriptor dispatcher, described
in [class-features.md](class-features.md). There is one shared accessorSymbol.
The earlier implementation and its decisions below are historical, superseded
by that document except for the fresh data-literal refusal retained here.

`class A { get total(): number { return 1; } }` followed by
`const items = [new A(), { total: 2 }]` is refused with
`adamic/accessor-field-view`, in either element order. A fresh data field cannot
supply the accessor identity of the inferred nominal view. Literal freshness
still permits ordinary field covariance. A getter/setter property's write type
is its setter parameter, proved by the override check, rather than its getter
result type treated as mutable storage.

The 26 branch fixtures remain registered. Twenty-four compare against Node,
native sanitizers and the JavaScript backend. Two retain their original programs
as explicit NotYet probes: optional boolean accessor slots and different native
getter/setter representations. Unrelated descriptors use distinct names, and
optional references are read through un-narrowed receiver helpers before their
local results are narrowed. These keep main's documented dispatch limits true.

## Initial accessor implementation (historical)

Stage 0 opens this part of the 0.2 accessor decision. Named instance getters and
setters on non-generic classes lower to ordinary typed methods. Each descriptor
half has its own virtual slot. A getter can compute, mutate, allocate or throw;
it is never a field load in the IR. The existing flow, alias, freshness, borrow,
region and reuse analyses therefore see a call and its virtual targets.

`value.x` calls the getter. Statement assignments call the setter. Statement
compound assignments and increments hold the receiver, call and hold the getter,
evaluate the right operand, compute the result and call the setter, in that order.
Even a plain receiver variable is held because the getter or operand can reassign
it. Throwing before the setter prevents both the remaining operand and setter
from running. Strings and objects returned by getters use ordinary owned method
returns, including exception cleanup. Virtual returns keep the heap convention.

Nominal base views and base-typed arrays dispatch to the most-derived accessor.
Overrides preserve descriptor halves, accept the base setter's input type or a
wider type, and return the base getter's result type or a sound subtype. Mutable
contents remain invariant. Native parameter and result representations must match.
A getter-only descriptor is readonly in Adamic's own view and invariance checks,
including when TypeScript's structural relation would forget that fact.

## Refused programs

These are language decisions, including temporary `NotYet` limits. The examples
use `class A { get x(): number { return 1; } }` and `const a = new A()` unless a
row gives another declaration. Every admitted program must keep the distinction
between a call and a stored field true.

| Program | Diagnostic and reason |
|---|---|
| `a.x = 2` on getter-only A | The checker refuses the readonly write; lowering also guards `adamic/getter-only-write`. There is no setter to execute. |
| `class S { set x(v: number) {} }` followed by `new S().x` | Refused, `adamic/setter-only-read`. Node reads undefined, while TypeScript types it as the setter's input. Setter-only writes are admitted. |
| `const v: { x: number } = a` or `const v: { readonly x: number } = a` | Refused, `adamic/accessor-field-view`. Structural property views do not prove getter or setter slots, even when readonly or declared with `interface View extends A {}`. A getter-only object can never become writable this way. |
| A getter class and a field class passed through the same `interface View { x: number }` | Refused at the accessor conversion with `adamic/accessor-field-view`. Choosing a load for one implementation and a call for another would lie about effects and storage. The rule also applies inside arrays, fields, signatures, casts and inferred views. |
| `const erased: {} = a`, or an upcast to a class omitting an accessor | Refused, `adamic/accessor-view-erasure`. Erasing the descriptor could later expose it as an optional structural field with unchecked writes. Keep a nominal view declaring that accessor. |
| `const o = { get x() { return 1; } }` or an object-literal setter | NotYet, `adamic/object-accessor`. Own enumerable descriptors need a separate object layout and spread implementation. |
| `class C { static get x() { return 1; } }` and static setters | NotYet, `adamic/static-accessor`. Static receivers and inherited static descriptors are not modeled by instance method tables. |
| `class C<T> { get x() { return 1; } }` and generic class setters | NotYet, `adamic/generic-accessor`. Declared generic accessors need substituted descriptor and override proofs. Inheriting an unchanged non-generic accessor does not add a new descriptor. |
| `interface View { get x(): number; }` or another descriptor declaration outside a class | NotYet, `adamic/accessor-declaration`. Structural descriptor syntax still does not prove a nominal virtual layout. |
| `function read<T extends A>(value: T) { return value.x; }` | NotYet, `adamic/generic-accessor-receiver`. The up-front pass cannot yet prove every instantiation has the constraint's nominal table rather than a structural property. |
| A derived getter returning `super.x` or a setter assigning `super.x` | NotYet, `adamic/super-accessor`. Super property access needs the base descriptor with the current receiver, rather than ordinary virtual dispatch. |
| `a?.x` | NotYet, `adamic/optional-accessor`. A short-circuiting call must preserve optional-chain order and result representation. |
| `a['x']`, `{ ...a }`, `const { x } = a`, or `function f({ x }: A) {}` | NotYet, `adamic/accessor-property-operation`. Existing indexed and destructured reads assume loads. Class accessors are inherited and absent from Node's own-field spread, so copying the type's accessor properties would be wrong. These operations conservatively refuse any accessor-bearing source. |
| `const old = a.x++`, `const sum = (a.x += 1)` or `const written = (a.x = 2)` | NotYet, `adamic/accessor-update-value`. Stage 0 has statement updates but no sequence expression carrying assignment or prefix/postfix results. Updates in expression statements and for-loop update positions are admitted. |
| `abstract get x(): number`, ambient accessors, `get ['x']()` or private accessor names | NotYet, `adamic/abstract-accessor` or `adamic/accessor-name`. These descriptors need additional declaration or name handling. |
| A field or method replacing an inherited accessor, or the converse | Refused, `adamic/accessor-member-kind` when the checker accepts the program. A virtual descriptor cannot silently turn into a storage slot or ordinary method. |
| A derived class overriding only the getter of a base getter/setter pair | NotYet, `adamic/accessor-descriptor-override`. JavaScript replaces the whole descriptor and loses the inherited setter. The same limit applies to removing a getter or adding a missing half. Getter-only overrides and complete pair overrides are admitted. |
| An override returning `Dog[]` where the base getter returns mutable `Animal[]`, or narrowing a setter's input | Refused, `adamic/accessor-override`. Base clients could mutate the narrower array or supply input the override cannot accept. |
| An otherwise sound override changing the native input or result representation | NotYet, `adamic/accessor-override-representation`. The existing virtual table requires one calling convention per slot. |
| Accessor dispatch through an un-narrowed union of class types | NotYet, `adamic/accessor-union`. Independent tables need a proven slot selection. Narrow with instanceof first. |
| A getter read narrowed to a different native representation | NotYet, `adamic/accessor-narrowing`. A second getter call can return a different value. Existing undefined checks still guard narrowed reference and optional scalar results. |
| Calling an accessor on this before required constructor fields are initialized, from a base constructor before derived initialization, or from a field initializer | Refused by the existing constructor and initializer rules. The accessor can call arbitrary code or read future fields, so it does not receive the initialized-field exception. |

## Evidence

Six `internal/oracle/testdata/accessors*.a` fixtures compare source Node with the
JavaScript backend, native release, ASan/UBSan and LeakSanitizer. They cover lazy
computation, validation, operand and receiver order, three-level dispatch,
base-typed arrays, throws and fresh strings and objects. Analysis probes include
a getter changing a global string, a getter retaining its result beyond the call,
and virtual getters and setters removing array elements while callers retain them.

The array-element borrow check now joins every virtual target. Previously it
consulted only the static implementation, which could miss an override's removal.
No other optimizer is disabled to admit accessors.

Fifteen mutants were run and caught, then restored:

| Mutant | Check that failed |
|---|---|
| Getter replaced with a field load | Node output and exit parity in accessors.a; native reported the missing field. |
| Setter skipped | Node output parity in accessors.a, hierarchy, order and ownership fixtures. |
| Compound receiver evaluated twice | Node output parity in accessors_order.a, including receiver redirection by the getter. |
| Static instead of virtual dispatch | Node output parity in hierarchy, ownership, throw and analysis fixtures. |
| Getter MayThrow dropped | accessors_throw.a ran the operand and setter after a throwing getter. Node output parity failed. |
| Getter-only writable interface admitted | Both direct and nested writable-interface refusal probes became admitted. |
| Borrow effects use only the static target | ASan heap-use-after-free in accessors_analyses.a's getterBorrow. |
| Accessor descriptor erasure permitted | The erased-accessor refusal probe became admitted. |
| Getter result covariance ignores mutable contents | The mutable getter-result override probe became admitted. |
| Structural accessor layout assumed nominal | Three interface inheritance and parameter probes became admitted. |
| Setter parameter contravariance omitted | The narrowed setter-input probe became admitted. |
| Override representation check omitted | The changed-representation probe became admitted. |
| Missing descriptor half inherited | The partial-descriptor override probe became admitted. |
| Setter-only read fabricated as zero | The setter-only read probe became admitted. |
| Spread and destructuring treated as loads | Three refusal probes became admitted. The indexed probe still reached an older NotYet guard, so that part is masked. |
