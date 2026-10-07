# Object static library slice

Branch: `codex/library-object`. Started from `origin/main` at `fe3b9f2`; the requested runner branch fast-forwarded it to `bb766d4`. No runner code was changed.

Test262 commit: `5992dc3b60faf62a48fd6be8a40ae9d9a8c84d81`. Node: `v24.19.0`. These are full `built-ins/Object` measurements with the same pinned checkout, not a limit or a classify-only run. The runner adapts the harness and runs the same adapted source on Node and natively.

| Measurement | Pass | Disagreements | Refused | Crashed | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| Before | 0 | 0 | 2335 | 0 | 1076 | 3411 |
| After | 11 | 0 | 2324 | 0 | 1076 | 3411 |

Observed: 11 newly passing, every one agreed with Node, no disagreements or crashes. The biggest remaining refusal is `var`: 394 tests both before and after. Full machine-readable reports: [before.json](before.json) and [after.json](after.json).

## What was built

- `Object.keys`: complete proven literal shapes, sorted with canonical array-index keys first, ascending; remaining strings preserve insertion order.
- `Object.values` and `Object.entries`: complete proven literal shapes whose tsc result has one homogeneous number, string or boolean representation. Entries are represented as tuples, with retained runtime strings and values.
- `Object.assign`: one to three typed sources into fields already in a proven target shape. Both directions of field assignability must agree. The returned object is the same target, with exactly the tsc intersection result; no type is widened. String ownership, source order and self-assignment are covered.
- `Object.freeze`: shallow freezing of proven plain objects, plus the identity operation for represented primitives. `Object.isFrozen` observes that state and distinguishes objects from primitives.
- `Object.is`: SameValue over represented values, including NaN equality, signed zero inequality, unions, undefined, strings and reference identity.
- `Object.hasOwn`: a literal key naming a declared public field or method, on a proven plain object or class instance. Class fields are own; prototype methods are not.

**Freeze enforcement is at runtime**, not a proof from readonly types. Objects carry a frozen flag. Every emitted field store and assign checks it, including updates and mutable aliases of a readonly freeze result. The write fixture creates such a mutable alias and agrees with Node on its TypeError. Nested objects remain writable: freezing is shallow. A spread must copy a frozen source instead of reusing it; a mutant proves that guard matters.

New implementation is in `internal/lower/library_object.go`, `internal/native/runtime/library_object.c`, and slice-specific IR, native emission and cycle-analysis files. Existing files have small dispatch, scalar-intersection representation, frozen-state initialization, write-check and reuse hooks. `internal/native/native.go` and `internal/lower/lower.go` were not edited. Six fixtures are registered in `internal/oracle/oracle_test.go` and recorded in `counts.md`.

## Explicit refusals and remaining coverage

| Operation or input | Result and reason in the diagnostic |
|---|---|
| defineProperty, defineProperties, getOwnPropertyDescriptor, getOwnPropertyDescriptors | Refused: property descriptors expose or change field presence, types and access behavior outside fixed plain loads and stores. |
| getPrototypeOf, setPrototypeOf, create | Refused: prototypes expose or replace fields outside the declared shape; use declared composition. |
| fromEntries | Refused: tsc returns an index-signature object with unproven keys; Adamic fixes shapes and refuses index signatures. Use Map. |
| groupBy | NotYet: tsc returns a partial record with dynamically present keys; this slice cannot represent it. Use Map and a typed grouping loop. |
| assign with no source or more than three sources | Refused when tsc chooses its any-returning overload; a proven result type is required. |
| assign with widening or conflicting field types | Refused: both directions must agree with the intersection result. |
| assign with mutable reference-valued fields | Conservatively refused; this slice only copies scalar values and strings and does not attempt a new cycle proof. |
| assign adding fields | NotYet: the target allocation has a fixed shape and aliases must keep the same object. |
| heterogeneous or unproven values/entries | Refused: no homogeneous, proven result element representation; any is not a type proof. |
| keys/values/entries/freeze on annotated, returned or otherwise unproven shapes | NotYet: this slice proves only literal shapes or their unannotated const bindings. Hidden fields and synthetic absent slots cannot be guessed. |
| hasOwn with arbitrary keys or unproven object views | Refused for undeclared keys; NotYet for unproven views, including optional object parameters. |
| catches around possibly frozen field writes or assign | NotYet: library TypeErrors currently panic natively rather than participate in catch cleanup. The diagnostic prevents a silent difference, conservatively across aliases and calls. |
| class/array/function freezing, null, symbols, heterogeneous union values/entries, computed names and special prototype/private-like literal names | Not covered; existing lowering or the shape proof reports a refusal or NotYet. |

These are implementation boundaries, not a claim that all ordinary JavaScript forms of Object are fundamentally unsound. Descriptor and prototype mutation conflict with Adamic fixed shapes; groupBy and general fromEntries would require a different representation. The lower tests contain adversarial widened-view, conflicting-type and incomplete-shape probes and check the diagnostic reasons.

## Node fixtures and mutants

Six source oracles cover ordering, SameValue, declared own fields, assign, freeze and a failing frozen write. Every successful fixture is compared with source on Node, the JavaScript backend, sanitized native and release native, and receives a separate leak check. The keys fixture exercises `0`, `2`, `10`, `01`, `-0`, `4294967294`, `4294967295` and `1e0`, alongside ordinary strings. Values and entries include strings constructed at runtime.

| Mutant | Fixture | What caught it |
|---|---|---|
| [keys-integer-last](mutants/keys-integer-last.log) | library_object_keys.a | Node stdout; test exit 1, no sanitizer or clang rejection |
| [keys-descending](mutants/keys-descending.log) | library_object_keys.a | Node stdout; test exit 1, no sanitizer or clang rejection |
| [values-first-slot](mutants/values-first-slot.log) | library_object_keys.a | Node stdout; test exit 1, no sanitizer or clang rejection |
| [entries-wrong-key](mutants/entries-wrong-key.log) | library_object_keys.a | Node stdout; test exit 1, no sanitizer or clang rejection |
| [assign-a-plus-one](mutants/assign-a-plus-one.log) | library_object_assign.a | Node stdout; test exit 1, no sanitizer or clang rejection |
| [same-value-nan-false](mutants/same-value-nan-false.log) | library_object_is.a | Node stdout; test exit 1, no sanitizer or clang rejection |
| [same-value-zero](mutants/same-value-zero.log) | library_object_is.a | Node stdout; test exit 1, no sanitizer or clang rejection |
| [has-own-always-false](mutants/has-own-always-false.log) | library_object_has_own.a | Node stdout; test exit 1, no sanitizer or clang rejection |
| [freeze-no-flag](mutants/freeze-no-flag.log) | library_object_freeze.a | Node stdout; test exit 1, no sanitizer or clang rejection |
| [freeze-write-ignored](mutants/freeze-write-ignored.log) | library_object_freeze_write.a | Node exit and stdout; test exit 1, no sanitizer or clang rejection |
| [spread-reuses-frozen](mutants/spread-reuses-frozen.log) | library_object_freeze.a | Node stdout; test exit 1, no sanitizer or clang rejection |

All eleven were run and restored. `run-mutants.py` rejects a mutant caught by compilation or a sanitizer. The first attempted assign mutant skipped the store while retaining/releasing strings and was caught by ASan; it does not qualify and was replaced by the numeric `assign-a-plus-one` mutation. A temporary fixture interpolated standalone undefined and was corrected to compare it with Object.is, since standalone undefined interpolation is an existing lowering gap.

## Commands and observed outputs

The toolchain environment is `/workspace/adamic-tools/env.sh`; each shell command sources it. `nproc` printed `5`; setup reported a cgroup quota of `400000 100000` and 17.6 GB.

```text
go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
v24.19.0
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (16s)
setup: done in 16s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

```sh
git fetch origin
git checkout -b codex/library-object origin/main
git fetch origin cloud/grok-test262-runner:refs/remotes/origin/cloud/grok-test262-runner
git merge origin/cloud/grok-test262-runner
bash cloud/setup.sh > /tmp/library-object-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic-test262 -json -test262 /workspace/scratch/test262 -work /tmp/object-before built-ins/Object > /tmp/object-before.json 2> /tmp/object-before.log
go run ./cmd/adamic-test262 -json -test262 /workspace/scratch/test262 -work /tmp/object-final built-ins/Object > /tmp/object-final.json 2> /tmp/object-final.log
go test -count=1 ./internal/lower -run TestObject > /tmp/object-refusals.log 2>&1
go test -count=1 -timeout 10m ./internal/oracle -run TestNativeAgreesWithNode/internal/oracle/testdata/library_object > /tmp/object-oracle.log 2>&1
python3 docs/library-object/run-mutants.py > /tmp/object-mutants.log 2>&1
gofmt -l cmd internal > /tmp/object-format.log
go vet ./... > /tmp/object-vet.log 2>&1
go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/object-counts.log 2>&1
go test -count=1 -timeout 30m ./... > /tmp/object-gate.log 2>&1
```

Formatting and vet: exit 0, empty output. The final [refusal tests](refusals.log) passed. [Oracle log](oracle.log) records the six fixtures passing. [Mutant summary](mutants.log) records all eleven Node comparisons failing as intended. All test output was redirected to files, never piped from a running test.

Counts regeneration: exit 0, `ok github.com/system-inc/adamic/internal/oracle 103.086s`. Exactly six rows were added; no existing count changed. The failing frozen-write row ends at the panic, so it is not expected to free its live object. [Counts log](counts.log). The first counts attempt failed on the single-target assign fixture when the any-result guard was added; the fixture was corrected and the full counts command rerun.

The assign mutant was rechecked after the final fixture correction: [final assign mutant summary](assign-mutant-final.log), caught by Node comparison.

Full gate: exit 0. `go test -count=1 -timeout 30m ./...` passed every package, including native (267.994s), the complete oracle package (713.461s), and all stage1/cohere packages. [Full gate output](gate.log). This final run validates the corrected assign fixture and all current implementation changes; the earlier filtered oracle run is supplementary.

## Per-directory tables

### Before

| Directory | Pass | Disagreements | Refused | Crashed | Skipped |
|---|---:|---:|---:|---:|---:|
| built-ins/Object | 0 | 0 | 55 | 0 | 6 |
| built-ins/Object/assign | 0 | 0 | 17 | 0 | 21 |
| built-ins/Object/create | 0 | 0 | 280 | 0 | 40 |
| built-ins/Object/defineProperties | 0 | 0 | 361 | 0 | 271 |
| built-ins/Object/defineProperty | 0 | 0 | 722 | 0 | 409 |
| built-ins/Object/entries | 0 | 0 | 10 | 0 | 11 |
| built-ins/Object/freeze | 0 | 0 | 23 | 0 | 30 |
| built-ins/Object/fromEntries | 0 | 0 | 8 | 0 | 17 |
| built-ins/Object/getOwnPropertyDescriptor | 0 | 0 | 306 | 0 | 4 |
| built-ins/Object/getOwnPropertyDescriptors | 0 | 0 | 8 | 0 | 10 |
| built-ins/Object/getOwnPropertyNames | 0 | 0 | 37 | 0 | 8 |
| built-ins/Object/getOwnPropertySymbols | 0 | 0 | 0 | 0 | 12 |
| built-ins/Object/getPrototypeOf | 0 | 0 | 37 | 0 | 2 |
| built-ins/Object/groupBy | 0 | 0 | 10 | 0 | 4 |
| built-ins/Object/hasOwn | 0 | 0 | 47 | 0 | 15 |
| built-ins/Object/internals/DefineOwnProperty | 0 | 0 | 0 | 0 | 6 |
| built-ins/Object/is | 0 | 0 | 14 | 0 | 7 |
| built-ins/Object/isExtensible | 0 | 0 | 36 | 0 | 2 |
| built-ins/Object/isFrozen | 0 | 0 | 56 | 0 | 3 |
| built-ins/Object/isSealed | 0 | 0 | 30 | 0 | 3 |
| built-ins/Object/keys | 0 | 0 | 52 | 0 | 7 |
| built-ins/Object/preventExtensions | 0 | 0 | 14 | 0 | 26 |
| built-ins/Object/prototype | 0 | 0 | 9 | 0 | 6 |
| built-ins/Object/prototype/__defineGetter__ | 0 | 0 | 4 | 0 | 7 |
| built-ins/Object/prototype/__defineSetter__ | 0 | 0 | 4 | 0 | 7 |
| built-ins/Object/prototype/__lookupGetter__ | 0 | 0 | 9 | 0 | 7 |
| built-ins/Object/prototype/__lookupSetter__ | 0 | 0 | 9 | 0 | 7 |
| built-ins/Object/prototype/__proto__ | 0 | 0 | 7 | 0 | 8 |
| built-ins/Object/prototype/constructor | 0 | 0 | 2 | 0 | 0 |
| built-ins/Object/prototype/hasOwnProperty | 0 | 0 | 49 | 0 | 14 |
| built-ins/Object/prototype/isPrototypeOf | 0 | 0 | 3 | 0 | 7 |
| built-ins/Object/prototype/propertyIsEnumerable | 0 | 0 | 9 | 0 | 7 |
| built-ins/Object/prototype/toLocaleString | 0 | 0 | 9 | 0 | 3 |
| built-ins/Object/prototype/toString | 0 | 0 | 13 | 0 | 28 |
| built-ins/Object/prototype/valueOf | 0 | 0 | 17 | 0 | 3 |
| built-ins/Object/seal | 0 | 0 | 56 | 0 | 38 |
| built-ins/Object/setPrototypeOf | 0 | 0 | 4 | 0 | 8 |
| built-ins/Object/values | 0 | 0 | 8 | 0 | 12 |

### After

| Directory | Pass | Disagreements | Refused | Crashed | Skipped |
|---|---:|---:|---:|---:|---:|
| built-ins/Object | 0 | 0 | 55 | 0 | 6 |
| built-ins/Object/assign | 0 | 0 | 17 | 0 | 21 |
| built-ins/Object/create | 0 | 0 | 280 | 0 | 40 |
| built-ins/Object/defineProperties | 0 | 0 | 361 | 0 | 271 |
| built-ins/Object/defineProperty | 0 | 0 | 722 | 0 | 409 |
| built-ins/Object/entries | 0 | 0 | 10 | 0 | 11 |
| built-ins/Object/freeze | 4 | 0 | 19 | 0 | 30 |
| built-ins/Object/fromEntries | 0 | 0 | 8 | 0 | 17 |
| built-ins/Object/getOwnPropertyDescriptor | 0 | 0 | 306 | 0 | 4 |
| built-ins/Object/getOwnPropertyDescriptors | 0 | 0 | 8 | 0 | 10 |
| built-ins/Object/getOwnPropertyNames | 0 | 0 | 37 | 0 | 8 |
| built-ins/Object/getOwnPropertySymbols | 0 | 0 | 0 | 0 | 12 |
| built-ins/Object/getPrototypeOf | 0 | 0 | 37 | 0 | 2 |
| built-ins/Object/groupBy | 0 | 0 | 10 | 0 | 4 |
| built-ins/Object/hasOwn | 0 | 0 | 47 | 0 | 15 |
| built-ins/Object/internals/DefineOwnProperty | 0 | 0 | 0 | 0 | 6 |
| built-ins/Object/is | 2 | 0 | 12 | 0 | 7 |
| built-ins/Object/isExtensible | 0 | 0 | 36 | 0 | 2 |
| built-ins/Object/isFrozen | 5 | 0 | 51 | 0 | 3 |
| built-ins/Object/isSealed | 0 | 0 | 30 | 0 | 3 |
| built-ins/Object/keys | 0 | 0 | 52 | 0 | 7 |
| built-ins/Object/preventExtensions | 0 | 0 | 14 | 0 | 26 |
| built-ins/Object/prototype | 0 | 0 | 9 | 0 | 6 |
| built-ins/Object/prototype/__defineGetter__ | 0 | 0 | 4 | 0 | 7 |
| built-ins/Object/prototype/__defineSetter__ | 0 | 0 | 4 | 0 | 7 |
| built-ins/Object/prototype/__lookupGetter__ | 0 | 0 | 9 | 0 | 7 |
| built-ins/Object/prototype/__lookupSetter__ | 0 | 0 | 9 | 0 | 7 |
| built-ins/Object/prototype/__proto__ | 0 | 0 | 7 | 0 | 8 |
| built-ins/Object/prototype/constructor | 0 | 0 | 2 | 0 | 0 |
| built-ins/Object/prototype/hasOwnProperty | 0 | 0 | 49 | 0 | 14 |
| built-ins/Object/prototype/isPrototypeOf | 0 | 0 | 3 | 0 | 7 |
| built-ins/Object/prototype/propertyIsEnumerable | 0 | 0 | 9 | 0 | 7 |
| built-ins/Object/prototype/toLocaleString | 0 | 0 | 9 | 0 | 3 |
| built-ins/Object/prototype/toString | 0 | 0 | 13 | 0 | 28 |
| built-ins/Object/prototype/valueOf | 0 | 0 | 17 | 0 | 3 |
| built-ins/Object/seal | 0 | 0 | 56 | 0 | 38 |
| built-ins/Object/setPrototypeOf | 0 | 0 | 4 | 0 | 8 |
| built-ins/Object/values | 0 | 0 | 8 | 0 | 12 |

## Newly passing tests

- `built-ins/Object/freeze/15.2.3.9-1-1.js`
- `built-ins/Object/freeze/15.2.3.9-1-3.js`
- `built-ins/Object/freeze/15.2.3.9-1-4.js`
- `built-ins/Object/freeze/15.2.3.9-1.js`
- `built-ins/Object/is/same-value-x-y-boolean.js`
- `built-ins/Object/is/same-value-x-y-number.js`
- `built-ins/Object/isFrozen/15.2.3.12-1-1.js`
- `built-ins/Object/isFrozen/15.2.3.12-1-3.js`
- `built-ins/Object/isFrozen/15.2.3.12-1-4.js`
- `built-ins/Object/isFrozen/15.2.3.12-1.js`
- `built-ins/Object/isFrozen/15.2.3.12-4-1.js`

## Commits

Implementation: `a18b3035896488f330040a98ee682bdde8268857` (`Lower sound Object methods and refuse unproven shapes`). The following report commit contains this report, both test262 snapshots, setup and validation logs, and the executable mutant harness. No pull request was opened.
