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
