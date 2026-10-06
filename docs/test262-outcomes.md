# test262 outcomes

`not-typescript` separates a checker rejection independently confirmed by stock
TypeScript from remaining Adamic refusals. Only an Adamic checker diagnostic is
eligible. The runner checks the exact adapted source, including the adapted
harness, with stock TypeScript and records the matching diagnostic code (for
example `TS2345`). If tsc accepts it, or only reports other codes, it remains
`refused`; its reason includes the Adamic code and tsc's codes or `accepted`.
Lowering refusals remain `refused`. Oracle failures are `crashed`, never a
confirmation of invalid TypeScript.

The compiler API uses the options in `internal/load/load.go`: `strict`,
`noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`, `noImplicitReturns`,
`noFallthroughCasesInSwitch`, `erasableSyntaxOnly`, `verbatimModuleSyntax`,
`allowImportingTsExtensions`, and `noEmit`, all true; module ESNext, forced module
detection, Bundler resolution, target ES2024, lib ES2024, and an empty `types`
list. It also loads `internal/load/prelude.d.ts` from the selected Adamic root.
Stock TypeScript's library declarations are unmodified. In particular, Adamic's
stricter RegExp library overlay is not used to decide what stock tsc accepts.

One Node process retains parsed declarations across requests. A per-run SHA-256
cache keys results by the complete adapted source bytes. Options, TypeScript
version, declarations and source location are fixed within that process; the
cache is not reused across runs. JSON and text reports include compiler version,
unique checks, cache hits and cumulative oracle request time in milliseconds.
The time includes communication and checking, excludes helper startup, and is
not a benchmark of a standalone tsc invocation. Both aggregate and per-directory
counts keep `not-typescript` separate from `refused`; TS codes have their own
reason counts.

## Controls and mutants

`TestTypescriptControls` exercises the real runner attempt path and real Adamic
and tsc compilers on the fixtures under `cmd/adamic-test262/testdata/typescript`:

| Fixture | Adamic | Stock tsc | Outcome |
|---|---|---|---|
| `argument.ts` | TS2345 | TS2345 | not-typescript |
| `missing.ts` | Cannot lower Object.getOwnPropertyNames yet | Accepts | refused |
| `disagree.ts` | TS2322 for the stricter RegExp split result, plus TS2304 | TS2304 only | refused, naming TS2322 and TS2304 |

The discrepancy fixture deliberately uses a real library declaration difference,
not a mocked compiler. `TestNotTypescriptTable` checks the directory counts,
aggregate counts, refusal reason membership, TS-code record and rendered table.

All mutants were actually run with `go test -count=1` and restored:

| Mutant | Check that failed | Observation |
|---|---|---|
| Replace the oracle call with Adamic's code alone | `TestTypescriptControls/disagree` | Incorrect not-typescript TS2322 |
| Accept any tsc code without comparing it | `TestTypescriptControls/disagree` | Incorrect not-typescript TS2322 |
| Also add not-typescript to refused totals | `TestNotTypescriptTable` | refused 2 instead of 1 |
| Use a constant content-cache key | `TestTypescriptControls/missing` and `/disagree` | Reused TS2345 for different source |

## Measurement

Linux is the gate of record. Both surveys use fetched main
`5d4c8012a0877094134e6c6bac367ff68f9313e8` as the compiler and runtime, test262 commit
`c8c798898646638cd0c24879f8e0374e847e7d74`, Node `v24.19.0`, and adaptation on.
The before runner was built from unchanged main; the after runner includes only
this outcome/reporting change. Stock TypeScript is npm `typescript@6.0.3`.
There are 64 top-level built-ins directories and no tests directly under
`built-ins/`. Every directory was run with no attempt limit, including skipped
programs in its total. `disagreement` below is the runner's `fail` field.
Tables are sorted by refused descending, then directory name.

Toolchain setup completed in 80 seconds: Go ready 0s, clang ready 1s, Node ready
1s, submodules ready 1s, build cache warm 80s. `nproc` reported 5; cgroup CPU
quota was 4 CPUs. Go was 1.27.1 and clang 20.1.8. The printed environment file
was `/workspace/adamic-tools/env.sh`.

Commands used (test output always redirected to logs):

```sh
git fetch origin && git checkout -b codex/test262-ts-validity origin/main
bash cloud/setup.sh > /tmp/test262-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm install --prefix /tmp/test262-typescript typescript@6.0.3 > /tmp/test262-npm.log 2>&1
export PATH=/tmp/test262-typescript/node_modules/.bin:$PATH
git clone --depth 1 https://github.com/tc39/test262.git /tmp/test262 > /tmp/test262-clone.log 2>&1
# Before editing the runner:
go build -o /tmp/test262-before ./cmd/adamic-test262 > /tmp/test262-build-before.log 2>&1
# After editing the runner:
go build -o /tmp/test262-final ./cmd/adamic-test262 > /tmp/test262-final-build.log 2>&1
```

Each survey passed every `built-ins/<directory>` as a separate filter, expanded
in sorted order from the checkout. The commands were:

```sh
/tmp/test262-before -adapt -json -test262 /tmp/test262 -work /tmp/test262-baseline \
  $(python3 -c 'from pathlib import Path; print(" ".join("built-ins/"+p.name for p in sorted(Path("/tmp/test262/test/built-ins").iterdir()) if p.is_dir()))') > /tmp/test262-before.json 2> /tmp/test262-before.log
/tmp/test262-final -adapt -json -test262 /tmp/test262 -work /tmp/test262-final-work \
  $(python3 -c 'from pathlib import Path; print(" ".join("built-ins/"+p.name for p in sorted(Path("/tmp/test262/test/built-ins").iterdir()) if p.is_dir()))') > /tmp/test262-after.json 2> /tmp/test262-after.log
go test -count=1 -timeout 15m ./cmd/adamic-test262 > /tmp/test262-tests-final.log 2>&1
go vet ./cmd/adamic-test262 > /tmp/test262-vet.log 2>&1
gofmt -l cmd internal > /tmp/test262-format.log
go vet ./... > /tmp/test262-vet-all.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 15m ./internal/oracle \
  -run 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp' > /tmp/test262-filtered-oracle.log 2>&1
```

The runner package passed in 26.551s. Both touched-package and repository-wide vet passed; repository formatting checks were clean.
The filtered uncached oracle passed all ten selected RegExp fixtures in 14.229s.
The full repository gate was not run. No lowering or native runtime was changed.
This survey retains the runner's existing skip and harness adaptation rules;
it measures agreement on adapted programs, not full JavaScript conformance.
In particular, the existing adapted assertThrows harness does not compare the
exact thrown constructor. Skipped programs have no execution observation.

## Complete before and after tables

### Before: unchanged main

| Directory | pass | disagreement | refused | not-typescript | crashed | skipped | total |
|---|---:|---:|---:|---:|---:|---:|---:|
| Array | 125 | 0 | 2339 | 0 | 8 | 610 | 3082 |
| Object | 17 | 1 | 2317 | 0 | 0 | 1076 | 3411 |
| String | 205 | 1 | 700 | 0 | 1 | 316 | 1223 |
| RegExp | 53 | 0 | 424 | 0 | 73 | 1329 | 1879 |
| Date | 0 | 0 | 324 | 0 | 0 | 270 | 594 |
| Iterator | 0 | 0 | 240 | 0 | 0 | 414 | 654 |
| Function | 0 | 0 | 228 | 0 | 0 | 281 | 509 |
| Set | 19 | 0 | 205 | 0 | 0 | 159 | 383 |
| Promise | 0 | 0 | 174 | 0 | 0 | 558 | 732 |
| DataView | 0 | 0 | 155 | 0 | 0 | 406 | 561 |
| Number | 152 | 0 | 114 | 0 | 0 | 74 | 340 |
| JSON | 7 | 0 | 95 | 0 | 0 | 63 | 165 |
| Map | 6 | 0 | 77 | 0 | 0 | 121 | 204 |
| WeakMap | 0 | 0 | 50 | 0 | 0 | 91 | 141 |
| ArrayBuffer | 0 | 0 | 48 | 0 | 0 | 173 | 221 |
| Math | 118 | 0 | 48 | 0 | 0 | 161 | 327 |
| WeakSet | 0 | 0 | 43 | 0 | 0 | 42 | 85 |
| Boolean | 0 | 0 | 37 | 0 | 0 | 14 | 51 |
| NativeErrors | 0 | 0 | 37 | 0 | 0 | 57 | 94 |
| ShadowRealm | 0 | 0 | 35 | 0 | 0 | 29 | 64 |
| Error | 0 | 0 | 31 | 0 | 0 | 62 | 93 |
| parseFloat | 20 | 0 | 31 | 0 | 0 | 3 | 54 |
| decodeURIComponent | 0 | 0 | 24 | 0 | 0 | 32 | 56 |
| decodeURI | 0 | 0 | 23 | 0 | 0 | 32 | 55 |
| global | 0 | 0 | 20 | 0 | 0 | 9 | 29 |
| parseInt | 33 | 0 | 19 | 0 | 0 | 3 | 55 |
| encodeURI | 0 | 0 | 16 | 0 | 0 | 15 | 31 |
| encodeURIComponent | 0 | 0 | 16 | 0 | 0 | 15 | 31 |
| ThrowTypeError | 0 | 0 | 11 | 0 | 0 | 3 | 14 |
| AggregateError | 0 | 0 | 6 | 0 | 0 | 19 | 25 |
| isFinite | 0 | 0 | 6 | 0 | 0 | 9 | 15 |
| isNaN | 0 | 0 | 5 | 0 | 0 | 10 | 15 |
| eval | 0 | 0 | 4 | 0 | 0 | 6 | 10 |
| Infinity | 0 | 0 | 3 | 0 | 0 | 3 | 6 |
| NaN | 0 | 0 | 3 | 0 | 0 | 3 | 6 |
| undefined | 0 | 0 | 3 | 0 | 0 | 5 | 8 |
| AsyncFunction | 0 | 0 | 1 | 0 | 0 | 17 | 18 |
| Atomics | 0 | 0 | 1 | 0 | 0 | 388 | 389 |
| AbstractModuleSource | 0 | 0 | 0 | 0 | 0 | 8 | 8 |
| ArrayIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 27 | 27 |
| AsyncDisposableStack | 0 | 0 | 0 | 0 | 0 | 104 | 104 |
| AsyncFromSyncIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 38 | 38 |
| AsyncGeneratorFunction | 0 | 0 | 0 | 0 | 0 | 23 | 23 |
| AsyncGeneratorPrototype | 0 | 0 | 0 | 0 | 0 | 48 | 48 |
| AsyncIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 13 | 13 |
| BigInt | 0 | 0 | 0 | 0 | 0 | 77 | 77 |
| DisposableStack | 0 | 0 | 0 | 0 | 0 | 93 | 93 |
| FinalizationRegistry | 0 | 0 | 0 | 0 | 0 | 47 | 47 |
| GeneratorFunction | 0 | 0 | 0 | 0 | 0 | 23 | 23 |
| GeneratorPrototype | 0 | 0 | 0 | 0 | 0 | 61 | 61 |
| MapIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 11 | 11 |
| Proxy | 0 | 0 | 0 | 0 | 0 | 311 | 311 |
| Reflect | 0 | 0 | 0 | 0 | 0 | 153 | 153 |
| RegExpStringIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 17 | 17 |
| SetIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 11 | 11 |
| SharedArrayBuffer | 0 | 0 | 0 | 0 | 0 | 104 | 104 |
| StringIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 7 | 7 |
| SuppressedError | 0 | 0 | 0 | 0 | 0 | 22 | 22 |
| Symbol | 0 | 0 | 0 | 0 | 0 | 98 | 98 |
| Temporal | 0 | 0 | 0 | 0 | 0 | 4605 | 4605 |
| TypedArray | 0 | 0 | 0 | 0 | 0 | 1453 | 1453 |
| TypedArrayConstructors | 0 | 0 | 0 | 0 | 0 | 738 | 738 |
| Uint8Array | 0 | 0 | 0 | 0 | 0 | 70 | 70 |
| WeakRef | 0 | 0 | 0 | 0 | 0 | 29 | 29 |
| **Total** | 755 | 2 | 7913 | 0 | 82 | 15069 | 23821 |

### After: independent TypeScript classification

| Directory | pass | disagreement | refused | not-typescript | crashed | skipped | total |
|---|---:|---:|---:|---:|---:|---:|---:|
| Object | 17 | 1 | 638 | 1679 | 0 | 1076 | 3411 |
| Array | 125 | 0 | 495 | 1844 | 8 | 610 | 3082 |
| String | 205 | 1 | 267 | 433 | 1 | 316 | 1223 |
| Date | 0 | 0 | 201 | 123 | 0 | 270 | 594 |
| Set | 19 | 0 | 125 | 80 | 0 | 159 | 383 |
| Function | 0 | 0 | 76 | 152 | 0 | 281 | 509 |
| Number | 152 | 0 | 70 | 44 | 0 | 74 | 340 |
| JSON | 7 | 0 | 69 | 26 | 0 | 63 | 165 |
| RegExp | 53 | 0 | 59 | 365 | 73 | 1329 | 1879 |
| DataView | 0 | 0 | 57 | 98 | 0 | 406 | 561 |
| Math | 118 | 0 | 38 | 10 | 0 | 161 | 327 |
| ArrayBuffer | 0 | 0 | 37 | 11 | 0 | 173 | 221 |
| Promise | 0 | 0 | 36 | 138 | 0 | 558 | 732 |
| WeakSet | 0 | 0 | 34 | 9 | 0 | 42 | 85 |
| Map | 6 | 0 | 32 | 45 | 0 | 121 | 204 |
| WeakMap | 0 | 0 | 27 | 23 | 0 | 91 | 141 |
| parseFloat | 20 | 0 | 21 | 10 | 0 | 3 | 54 |
| Boolean | 0 | 0 | 19 | 18 | 0 | 14 | 51 |
| Error | 0 | 0 | 19 | 12 | 0 | 62 | 93 |
| NativeErrors | 0 | 0 | 18 | 19 | 0 | 57 | 94 |
| decodeURIComponent | 0 | 0 | 12 | 12 | 0 | 32 | 56 |
| decodeURI | 0 | 0 | 11 | 12 | 0 | 32 | 55 |
| encodeURI | 0 | 0 | 8 | 8 | 0 | 15 | 31 |
| encodeURIComponent | 0 | 0 | 8 | 8 | 0 | 15 | 31 |
| global | 0 | 0 | 5 | 15 | 0 | 9 | 29 |
| isFinite | 0 | 0 | 3 | 3 | 0 | 9 | 15 |
| AggregateError | 0 | 0 | 2 | 4 | 0 | 19 | 25 |
| eval | 0 | 0 | 2 | 2 | 0 | 6 | 10 |
| isNaN | 0 | 0 | 2 | 3 | 0 | 10 | 15 |
| Atomics | 0 | 0 | 1 | 0 | 0 | 388 | 389 |
| Infinity | 0 | 0 | 1 | 2 | 0 | 3 | 6 |
| parseInt | 33 | 0 | 1 | 18 | 0 | 3 | 55 |
| AbstractModuleSource | 0 | 0 | 0 | 0 | 0 | 8 | 8 |
| ArrayIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 27 | 27 |
| AsyncDisposableStack | 0 | 0 | 0 | 0 | 0 | 104 | 104 |
| AsyncFromSyncIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 38 | 38 |
| AsyncFunction | 0 | 0 | 0 | 1 | 0 | 17 | 18 |
| AsyncGeneratorFunction | 0 | 0 | 0 | 0 | 0 | 23 | 23 |
| AsyncGeneratorPrototype | 0 | 0 | 0 | 0 | 0 | 48 | 48 |
| AsyncIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 13 | 13 |
| BigInt | 0 | 0 | 0 | 0 | 0 | 77 | 77 |
| DisposableStack | 0 | 0 | 0 | 0 | 0 | 93 | 93 |
| FinalizationRegistry | 0 | 0 | 0 | 0 | 0 | 47 | 47 |
| GeneratorFunction | 0 | 0 | 0 | 0 | 0 | 23 | 23 |
| GeneratorPrototype | 0 | 0 | 0 | 0 | 0 | 61 | 61 |
| Iterator | 0 | 0 | 0 | 240 | 0 | 414 | 654 |
| MapIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 11 | 11 |
| NaN | 0 | 0 | 0 | 3 | 0 | 3 | 6 |
| Proxy | 0 | 0 | 0 | 0 | 0 | 311 | 311 |
| Reflect | 0 | 0 | 0 | 0 | 0 | 153 | 153 |
| RegExpStringIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 17 | 17 |
| SetIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 11 | 11 |
| ShadowRealm | 0 | 0 | 0 | 35 | 0 | 29 | 64 |
| SharedArrayBuffer | 0 | 0 | 0 | 0 | 0 | 104 | 104 |
| StringIteratorPrototype | 0 | 0 | 0 | 0 | 0 | 7 | 7 |
| SuppressedError | 0 | 0 | 0 | 0 | 0 | 22 | 22 |
| Symbol | 0 | 0 | 0 | 0 | 0 | 98 | 98 |
| Temporal | 0 | 0 | 0 | 0 | 0 | 4605 | 4605 |
| ThrowTypeError | 0 | 0 | 0 | 11 | 0 | 3 | 14 |
| TypedArray | 0 | 0 | 0 | 0 | 0 | 1453 | 1453 |
| TypedArrayConstructors | 0 | 0 | 0 | 0 | 0 | 738 | 738 |
| Uint8Array | 0 | 0 | 0 | 0 | 0 | 70 | 70 |
| WeakRef | 0 | 0 | 0 | 0 | 0 | 29 | 29 |
| undefined | 0 | 0 | 0 | 3 | 0 | 5 | 8 |
| **Total** | 755 | 2 | 2394 | 5519 | 82 | 15069 | 23821 |

The change separates 5,519 tests from 7,913 refusals, leaving 2,394 refused. There are zero newly passing tests, zero new disagreements, and identical pass path sets in every directory. The two existing disagreements, one each in Object and String, are native exit 70 where Node exits 0; both are present on main. The 82 crashes are also unchanged. The full corpus therefore does not meet a zero-total-disagreement bar; this outcome change introduces no disagreement.

Oracle cost: TypeScript 6.0.3, 5,532 unique source checks, 10 content-cache hits, and 328,332ms (328.332s) cumulative request time. Cache-hit requests do not run tsc again.

Date: refused 324 before and 201 after, with 123 not-typescript. Its new TS-code counts are TS2345=100, TS2554=10, TS2683=7, TS2769=3, TS7005=3.

## Top ten directories by remaining refused

The following are the five most common normalized refusal reasons per directory, sorted by count descending, then exact reason. Only remaining refused tests enter these counts.

### Object: 638 refused

| Count | Reason |
|---:|---|
| 185 | refuses Object.defineProperty |
| 99 | not yet: a value of type any |
| 38 | not yet: Object.seal |
| 35 | not yet: Object.isExtensible |
| 28 | not yet: Object.isSealed |

### Array: 495 refused

| Count | Reason |
|---:|---|
| 51 | refuses a method read as a value (indexOf would lose its object, and this with it) |
| 50 | not yet: new an Identifier |
| 49 | refuses a method read as a value (lastIndexOf would lose its object, and this with it) |
| 26 | refuses delete |
| 23 | not yet: an array of never |

### String: 267 refused

| Count | Reason |
|---:|---|
| 49 | not yet: new an Identifier |
| 22 | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| 16 | not yet: a BinaryExpression with a string and a number |
| 14 | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| 13 | refuses != |

### Date: 201 refused

| Count | Reason |
|---:|---|
| 43 | not yet: a value of type any |
| 42 | refuses inherited library member prototype read as an own field |
| 14 | refuses a method read as a value (toString would lose its object, and this with it) |
| 12 | not yet: new an Identifier |
| 10 | refuses inherited library member UTC read as an own field |

### Set: 125 refused

| Count | Reason |
|---:|---|
| 19 | not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) |
| 13 | refuses a method read as a value (add would lose its object, and this with it) |
| 9 | not yet: instanceof against a value that isn't a declared class |
| 9 | refuses a method read as a value (clear would lose its object, and this with it) |
| 9 | refuses a method read as a value (delete would lose its object, and this with it) |

### Function: 76 refused

| Count | Reason |
|---:|---|
| 14 | refuses inherited library member caller read as an own field |
| 11 | refuses a method read as a value (toString would lose its object, and this with it) |
| 8 | not yet: reading Function |
| 8 | refuses inherited library member bind read as an own field |
| 5 | refuses var |

### Number: 70 refused

| Count | Reason |
|---:|---|
| 37 | not yet: toString through an object view (a value with a different native representation may be hidden by the view) |
| 6 | not yet: new an Identifier |
| 6 | refuses var |
| 3 | refuses isPrototypeOf |
| 2 | not yet: a try around toString, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) |

### JSON: 69 refused

| Count | Reason |
|---:|---|
| 33 | refuses JSON.parse: its result's type can't be proven from the text |
| 10 | not yet: new an Identifier |
| 6 | not yet: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) |
| 5 | refuses var |
| 4 | refuses a method read as a value (toString would lose its object, and this with it) |

### RegExp: 59 refused

| Count | Reason |
|---:|---|
| 9 | not yet: RegExp with a nonconstant pattern |
| 8 | refuses var |
| 6 | not yet: a BinaryExpression with a value and a string |
| 5 | refuses a method read as a value (exec would lose its object, and this with it) |
| 5 | refuses a method read as a value (test would lose its object, and this with it) |

### DataView: 57 refused

| Count | Reason |
|---:|---|
| 52 | not yet: new an Identifier |
| 3 | not yet: a PropertyAccessExpression as a statement |
| 1 | not yet: Object.isExtensible |
| 1 | not yet: reading DataView |

This directory has fewer than five distinct refusal reasons; all are shown.
