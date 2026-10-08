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
The amended ruling keeps all eight debug-site refusals exactly; no prototype link
is admitted. Runtime commit 0e2af455 was explicitly requested and cherry-picked as
bd59c092c, following OrdinaryToPrimitive commit 12cc2e52 as a44503796. These are
the first commits of the runtime-dependency batch, after the already pushed lowering
slices. Neither dependency branch was merged. Step 21 `origin/codex/scout-exceptions` remains a
separate unlanded dependency for generalized error identity/throw support.

## Catchable error slice and dependency block

String.repeat now guards JavaScript count and length failures before calling the
native helper. It throws catchable Errors named RangeError, with Node's diagnostic
(including the original fractional count), through the existing exception cleanup.
Both receiver and count evaluate once; finally runs after a caught failure. Adamic
inserted checks and allocation failures remain terminal. Other conversion/property
error sites and generalized exception identity remain unfinished.

The new catchable_errors.a fixture covers negative/fractional counts, NaN, infinity,
the UTF-16 length limit, empty strings and evaluation order. Its mutant changes
RangeError names to Error; Node output detects the mutation in both backends.
Sanitized and release native and the JavaScript backend are compared with Node.
Focused lower checks pass (ruling6-lower.log). A sandboxed oracle retry failed
because LeakSanitizer cannot run under ptrace; its first unsandboxed retry reused
those cached failures. The final ADAMIC_GATE_UNCACHED=1 focused run outside ptrace
passes all four tests, with leak checks and both mutants caught
(runtime-merge-oracle-uncached.log). Counts refresh passes (ruling6-counts.log).
Logs are under /tmp/object-semantics.

The runtime-dependency test command selects TestOrdinaryToPrimitive, its mutants
and protocol guards, TestPrototypeNodeResearch and TestPrototypeDebugSiteRefusals.
It is blocked before compilation: internal/leakcheck is missing from this area base.
That shared package was introduced by f6eef5df9 and updated by a29a1b9be; those
commits also touch unrelated and protected files and have not been cherry-picked.
The prototype graph control also requires missing graph_regions.h. These are
dependencies, not passed checks. The lowering still uses its existing IR conversion
loop; wiring it to adamic_ordinary_to_primitive remains pending. No claim is made
that the C runtime contract is yet used by generated conversion code.

## Runtime base delivery

The branch is moved onto runtime/area-on-next b74158d2 as explicitly requested.
That base supplies shared leakcheck and graph regions. Neither step 22 runtime
commit is already present, so both remain: a44503796 becomes 0936496c6,
bd59c092c becomes 90ed82a59, and 6fb710351 becomes 37cc25cd6. Earlier cached
intrinsic and conversion lowering commits are retained on the new base.
Prototype operations stay refused. The earlier dependency block above describes
the old base only. The C conversion adapter remains unfinished.

On the runtime base, the focused native command selecting OrdinaryToPrimitive,
its five mutants and protocol guards, prototype Node research, all eight debug
refusals and the mixed counted-cycle probe passes. The mixed cycle leaks two
counted objects as expected; the graph-region control frees them. No prototype
operation is admitted. The uncached six object oracle tests pass in both backends
against Node, with sanitizer and leak checks. Their receiver-dispatch, missing
conversion TypeError and wrong RangeError-name mutants are caught. Focused lower
checks pass. Logs: runtime-base-native.log, runtime-base-oracle.log and
runtime-base-lower.log under /tmp/object-semantics.

The required TestCountsAreRecorded -count=1 -args -update-counts run passes and
refreshes counts on this base (runtime-base-counts.log). No full gate or whole
package was run.

## Step 22: native conversion adapter

Templates, String and string-producing addition now call
adamic_ordinary_to_primitive for admitted object and array operands. Lowering
records verified method signatures and emits typed callback helpers. Native
keeps the original receiver, transfers owned strings and objects, unboxes scalar
union results, distinguishes null from undefined, and checks the pending error
before using a result. Rejected objects are released by the runtime before the
next lookup. The original IR protocol remains executable for JavaScript and
visible to ownership/effect analysis. Both addition operands still evaluate
before conversion. The earlier statement that the C adapter is unfinished is
superseded by this delivery.

conversion_runtime.a returns owned objects from both conversion methods. Node
prints TypeError: Cannot convert object to primitive value for addition and
template conversion, with valueOf/toString order determined by the hint. A
valueOf object result followed by a primitive toString result is legal and stays
allowed. Controls cover returning this, method exceptions, optional numbers and
booleans, tagged boolean/object results and nullable object results. Sanitized
and release native and JavaScript agree with Node, with native leak checks.

The runtime-only mutant removes the exhausted-protocol TypeError. The generated
program then prints undefined and wrongly continued, exiting zero without leaks;
Node catches the wrong output. The mutant archive now uses RuntimeLibraryForSource
to match the generated program's runtime feature flags. A first attempt used an
incompatible archive and failed ASan; that failure is not counted as a killed
behavior mutant. The valid source-aware mutant is caught by output alone. The
existing IR exhaustion mutant also bypasses the native runtime adapter and is
caught in both backends.

The conversion fixture caught an existing nullable-object conversion error in
the JavaScript protocol: null fell through to toString. Lowering now tests null
explicitly instead of testing only undefined. A proposed wider primitive-union
return was rejected by the existing return-type boundary and is not admitted.
Accessors, optional callable members, specialized host objects, Symbol.toPrimitive
and directly interpolated object-containing unions retain their earlier NotYet
boundaries. Prototype operations stay refused.

Focused commands (all output under /tmp/object-semantics):
- internal/oracle: TestObjectOrdinaryPrimitive, its mutant and
  TestObjectConversionRuntime, uncached (conversion-adapter-fixtures.log).
- internal/native: TestOrdinaryConversionAdapterMutant
  (conversion-adapter-mutant.log).
- internal/lower: TestOrdinaryPrimitiveAdmission, TestOrdinaryPrimitiveBoundaries
  and TestLibraryStringRefusals (conversion-adapter-lower.log).
- internal/oracle: the terminal accessor-readiness check and existing
  library_string_conversion and library_method_values fixtures, uncached
  (conversion-adapter-regressions.log).
- internal/native: existing OrdinaryToPrimitive mutants and protocol guards
  (conversion-adapter-runtime-mutants.log).
- internal/oracle: TestCountsAreRecorded -count=1 -args -update-counts
  (conversion-adapter-counts.log).
No whole package or full gate was run.

All focused commands listed for this adapter delivery pass. The new fixture
allocates and frees 58 heap values; the existing conversion fixture allocates
and frees 91. The runtime TypeError mutant exits zero, is leak-clean, and is
rejected solely by disagreement with Node.
