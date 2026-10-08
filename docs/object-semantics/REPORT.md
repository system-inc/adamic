# Step 22 object semantics

## Ruling 2: cached own-property intrinsic

The area base already traces const aliases to library declarations. The new dispatch uses
that provenance to call the borrowed `Object.prototype.hasOwnProperty` intrinsic without
looking up the receiver's overriding field. Its verified signature takes a present object
and a string key and returns boolean. No explicit `this` annotation or source adaptation
is needed. The JavaScript backend now emits the intrinsic call for `ir.HasOwn` too.

Conservative assumption: this first slice accepts complete plain-literal shapes and their
const bindings. Structural views, private fields, tuples and primitive boxing still require
their own presence proofs; their existing NotYet boundaries are retained.

Observation: the unmodified area base rejects the fixture at the cached call because it
spots the receiver's own `hasOwnProperty: false`. Node prints:

```
2: true
z: true
missing: true
hasOwnProperty: true
constructor: false
absent: false
true object key 
```

The fixture is reduced from TypeScript's own-key helpers. It covers inferred const aliases,
shadowing on the receiver, inherited and absent keys, a declared undefined field, and
receiver/key evaluation order. Sanitized native, release native and the independent IR
JavaScript backend are held to source Node; the sanitized program is checked for leaks.
The override-dispatch mutant replaces the queries with the receiver override's false
result. It exits zero without leaks; source Node catches the stdout mismatch in both
backends. This separates the observed output from the provenance inference above.

The setup installed the toolchain and submodules, then could not append to the read-only
`/home/agent/.bashrc`. Sourcing its generated environment works. Reported cumulative setup
lines: Go 0.087s; Node 0.105s; clang 0.625s; markdown ready 0.189s; submodules 219.698s.
`nproc` reports 5. GOPROXY was set to `https://proxy.golang.org|direct` before setup.

Focused commands and their logs are recorded with each delivery. No whole-package test or
full gate is run for this unit. Rulings 5, 3 and 6 follow this slice; ruling 4 belongs to the
TypeScript adaptation. Step 21's exception branch is not on the starting area tip and will
be named as a dependency rather than merged while unlanded.

Ruling 2 validation after merging current main: `go test ./internal/lower -run
'^TestLibraryMethodValue' -count=1` passes; `go test ./internal/oracle -run
'^TestObjectCachedIntrinsic' -count=1 -v` passes both the Node fixture and its mutant.
`go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(library_method_values|method_coverage_object_descriptors)'
-count=1 -v` passes the existing descriptor and alias fixtures. The initial count run
found missing pinned Node declarations and two intercepted descriptor paths; both were
fixed. `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts`
then passes and refreshes the table, including the new fixture.

## Ruling 5: ordinary conversion

Templates and String use the string hint (toString before valueOf); string-producing
addition uses the default hint (valueOf before toString). Addition holds both operands
before either conversion. A conversion method receives the original object as this.
Non-callable fields are skipped. Object results, including boxed empty-object results,
fall through to the next method. Primitive-or-object unions are tested at runtime;
only primitive results terminate the protocol. Exhaustion throws an Error named
TypeError, with Node's `Cannot convert object to primitive value`, through existing
exception cleanup. A caught failure runs finally and releases held conversion results.

The new fixture reduces semver's Version.toString and debug's declaration-kind array,
including its absent branch. Controls exercise receiver identity, argument order,
non-callable overrides, undefined/void conversion results, ordinary-object fallback,
class methods and varying primitive/object results. Sanitized native, release native,
JavaScript and source Node agree; the successful native runs are leak-clean.

Mutation: replace the conversion-exhaustion throw with an object-tag return. Native
exits zero without leaks and prints wrong output; source Node catches it in both
backends. The fixture also proves that an object-returning valueOf alone is not an
error: toString can still return a primitive, exactly as OrdinaryToPrimitive requires.

Conservative boundaries remain NotYet for conversion accessors, optional callable
members, host objects with specialized behavior, tuples held as native objects,
Symbol.toPrimitive, and object-containing unions interpolated directly. This is an
implemented admitted subset, not a claim that all dynamic/exotic conversion is finished.

Stage 3 `objects/15_accessor_absence.a` now lowers and reaches its existing readiness
check: its definite-assignment assertion lied about secondAccessor. Its record changes
from NotYet to CheckedStop, pinning exit 70 and the diagnostic; its Node record is
unchanged byte for byte. `taste/02_diagnostic_message.a` still needs dynamic object-union
conversion and its record is unchanged. No inserted soundness check becomes catchable.

Validation: focused OrdinaryPrimitive admission/boundary tests, the existing stage-zero
and library-string boundaries, library-method tests, both object-semantics Node/mutant
fixtures, and the existing library_string_conversion fixture pass. The filtered
stage3/fixtures run for objects/15 and taste/02 passes. Counts were refreshed with
`go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts` (PASS).
Output is in `/tmp/object-semantics/ruling5-*.log`; no full gate or whole package was run.

## Prototype ownership dependency

The runtime report on `origin/runtime/step22-prototype-links` at `0e2af4559` exposes no
implemented prototype API. Its mixed-cycle witness is accepted by Node even though
ordinary counted prototype ownership would leak: an object points to its prototype,
and a data field on that prototype points back to the object. A prototype-chain-only
cycle check cannot detect this ownership cycle and must not reject it as JavaScript's
cyclic-chain TypeError. Graph-region integration is absent on this compiler base.
A direction question is pending before changing this rule. That worker's unlanded
branch was inspected, not merged. Step 21 `origin/codex/scout-exceptions` remains a
separate unlanded dependency for generalized error identity/throw support.
