# Accessors lowered as methods

Stage 0 opens this part of the 0.2 accessor decision. Named instance getters and
setters lower to ordinary typed methods, monomorphized with generic classes. Each descriptor
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
| `class C { static get x() { return 1; } }` and static setters | NotYet, `adamic/static-accessor`. A static getter's receiver is the class value: `Base.x` and `Derived.x` differ. The IR has no class-value representation yet. |
| `class C<T> { get x() { return 1; } }` and generic class setters | Admitted. Accessors are monomorphized with the class. Override proofs substitute the read and write types through each declaration's nominal class view; mutable contents and native representations are checked on those substituted types. |
| `interface View { get x(): number; }` or another descriptor declaration outside a class | NotYet, `adamic/accessor-declaration`. Structural descriptor syntax still does not prove a nominal virtual layout. |
| `function read<T extends A>(value: T) { return value.x; }` | NotYet, `adamic/generic-accessor-receiver`. The up-front pass cannot yet prove every instantiation has the constraint's nominal table rather than a structural property. |
| A derived getter returning `super.x` or a setter assigning `super.x` | Admitted. `super.x` directly calls the base getter and `super.x = value` directly calls the base setter with the current receiver. These are non-virtual calls; ordinary accessor reads inside the base body still dispatch virtually. |
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

## Static accessor stopping point

The ordered follow-on unit starts from `origin/codex/accessors` at `d785e87`:
that commit is not an ancestor of the fetched `origin/main` at `ef3d907`.
The first commit stopped before admitting static accessors. The follow-on work
continues with generic descriptors, super descriptor calls and update values.
Object-literal descriptors remain NotYet.

Observed on Node 24.19.0, `review/accessors-2/static_receiver.a` exits 0 and prints:

```
10 20
base 3
derived 4
```

The same getter and setter declarations execute with `Base` as `this` for
`Base.x`, and with `Derived` as `this` for `Derived.x`. Both `adamic c` and
`adamic js` reject this probe at the static getter with `adamic/static-accessor`,
before either backend emits code.

The implementation has only instance receivers: `instantiate` gives every
method an object-typed `thisLocal`, while `expression.go` rejects reading a
class name as a value. Static methods are excluded from instance tables, but
still receive that object parameter. Static fields separately return NotYet in
`inheritanceConstructor`. These are code observations, not a claim that static
accessors are impossible to implement.

The inference is that simply removing the refusal and emitting a function per
static descriptor is insufficient for the approved semantics. A receiver model
must preserve constructor identity for inherited reads and writes; static
member lookup must also preserve descriptor halves. Binding `this` to the class
that declares the function would print the wrong values in the probe above.
A restricted implementation could refuse every receiver-dependent body, but
that would leave the receiver part of this ordered item unfinished. This unit
keeps the named refusal instead of claiming the static row is admitted.

Unused getter and setter declarations now have dedicated refusal probes in
`TestAccessorRefusals`, alongside an inherited static receiver probe. Mutant:
remove only the static-accessor guard from `accessorRefusal`. Both unused
probes then lower successfully and fail the test with `got <nil>`; no backend
or clang diagnostic masks the missing guard. The mutant was restored. No new
feature runs natively, so this stopping-point probe has no sanitizer or leak
result. The existing six admitted accessor fixtures are the backend regression
coverage for this documentation and refusal-test change.

Validation for the stopping-point commit (logs in `/tmp/accessors-2-*.log`):

- `go test -count=1 -timeout 10m ./internal/lower`: passed, 7.826s.
- `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 10m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/accessors'`: passed, 10.724s.
  This covers all six existing accessor fixtures against source Node, both
  backends, the release build, ASan/UBSan and LeakSanitizer.
- `go vet ./internal/lower`: exit 0; `gofmt -l internal/lower/accessors_test.go`:
  empty; `git diff --check`: exit 0.
- Setup: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s,
  build cache warm 83s, done 83s; `nproc`: 5.

The full repository gate was not run. No fixture was added to the admitted
oracle list and no allocation-count row changed.

## Generic-class accessors

`accessors_generic.a` holds number, string and object instantiations, a derived
class whose base uses its second type argument, nominal base dispatch, owned
getter results kept across setter calls, and a throwing virtual getter that
prevents the operand and setter. The fixture passes source Node, the JavaScript
backend, native release, ASan/UBSan and LeakSanitizer. The counts update passed
and added only this fixture's row.

Two mutants were restored after failing:

- Leaving accessor signatures unsubstituted refuses the sound reordered-argument
  override in `TestGenericAccessorOverridesUseSubstitutedTypes`.
- Omitting the read/write type relation admits the mutable array covariance
  probe in `TestGenericAccessorOverrideRejectsMutableCovariance` (`got <nil>`).

Commands: `go test -count=1 -timeout 10m ./internal/lower`, the uncached
`TestNativeAgreesWithNode/internal/oracle/testdata/accessors` filter, and
`go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts`.
Logs are `/tmp/accessors-2-generic-*.log`. Static accessors and generic function
receivers remain NotYet. Getters remain virtual `ir.Call` nodes.

## Super descriptor calls

`accessors_super.a` passes source Node, both backends, ASan/UBSan and
LeakSanitizer. It covers three levels of descriptor overrides, direct base calls
with the leaf receiver, virtual calls made inside a base body, compound super
updates, fresh string results held across writes, generic base descriptors and
a throwing base getter that suppresses the operand and setter.

Four mutants were run and restored:

- Making super descriptor calls virtual fails
  `TestSuperAccessorsAreDirectCallsWithCurrentReceiver`.
- Skipping a super setter fails stdout parity against Node in both backends;
  native and JavaScript finish with exit 0, so clang and sanitizers do not mask it.
- Ignoring super in field-initializer validation admits the initializer probe.
- Skipping `useOfThis` for super admits the early constructor accessor probe.

Validation: full `internal/lower` passed (15.917s), the uncached accessor oracle
filter passed (1.445s), and the full counts update passed (18.729s), adding only
the super fixture's row. Logs are `/tmp/accessors-2-super-*.log`. These getter
reads remain direct `ir.Call` nodes, never loads.
