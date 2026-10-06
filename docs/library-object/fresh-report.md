Built Object sealing/extensibility, own-name reflection, stable literal bindings and exact-null SameValue; 41 newly passing tests agree with Node.<br>
Commits: claim `c48709e1a7c119f6f479ee396e04cfab1d6a66de`, implementation `d42b393f1e3dc4d7a34222f2758a45e3aa41c72c`, followed by this report commit.<br>
Commands: full adapted Object measurement, touched-package tests, uncached Object oracle/counts, formatting and vet all exit 0; final 58 pass, 0 disagreement, 2277 refused, 0 crashed, 1076 skipped.<br>
Mutants: seven executed; every one caught by Node stdout alone, listed with its receipt below.<br>
Not covered: descriptor/prototype mutation, unrepresented host values and constructors, arbitrary/absent shapes, array integrity mutation and nullable-union SameValue; explicit refusal reasons remain.

# Fresh Object library report

Branch: `codex/library-object`, from current main `5d4c801`. Linux is the gate of record. No PR was opened. Test262 is pinned to `5992dc3b60faf62a48fd6be8a40ae9d9a8c84d81`; Node is `v24.19.0`. Both measurements use the official `cmd/adamic-test262` with `-adapt`. No file under `cmd/adamic-test262/` was edited.

## Before and after

| Measurement | Pass | Disagreement | Refused | Crashed | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| Fresh before | 17 | 1 | 2317 | 0 | 1076 | 3411 |
| Fresh after | 58 | 0 | 2277 | 0 | 1076 | 3411 |

Observed: 41 newly passing, no previous pass lost, zero disagreements and zero crashes. Forty formerly refused tests now pass; the other new pass closes the baseline disagreement `built-ins/Object/is/not-same-value-x-y-undefined.js`. Full reports: [fresh-before.json](fresh-before.json), [fresh-after.json](fresh-after.json), with [before log](fresh-before.txt) and [after log](fresh-after.txt). The separate source-capture run matches the official baseline JSON exactly.

## Classify before building

The claim was committed and pushed before implementation as `c48709e`. Its original first-diagnostic classification was (a) 1679, (b) 388, (c) 250. After inspecting the generic constructor/value diagnostics, 37 were named as library dependencies instead of general representation gaps. Final buckets are (a) 1679, (b) 425, (c) 213. The final repeat uses real `typescript@6.0.3` with Adamic's exact pinned declaration text, its RegExp declaration corrections, the complete prelude and identical compiler options. The accepted/rejected partition remained identical; every rejection matches the original TS diagnostic code. The initial npm-declaration pass and the exact-declaration repeat are distinguished in [refusals.md](refusals.md).

Bucket (a)'s largest reasons are TS18048 (343), TS2339 (311), TS2345 (216), TS2322 (181) and TS7009 (166). They are real tsc rejections, not work owed to this library slice. Bucket (c)'s largest reason is unproven any (99). The language handoff in [refusals.md](refusals.md) names every observed language/representation reason with a representative adapted body on one line. Full exact sources, hashes and diagnostics are linked there.

The largest bucket (b) feature is descriptor mutation: defineProperty (185), then defineProperties (24). These remain refused because the bound `CLAUDE.md` and approved `docs/0.1.md` require fixed shapes and plain loads/stores, and specifically exclude descriptor/prototype mutation. Closing them would require a sound descriptor/shape representation design beyond this slice. I built the largest remaining family: seal (38), isExtensible (35), isSealed (28) and preventExtensions (9), then own-name reflection (16) and keys shape proofs (14). These counts are first-blocker candidates; closing a call can reveal another refusal. Unrepresented host constructors discovered behind seal stay refused and are named as library dependencies.

## What changed

- `Object.seal` and `Object.preventExtensions` preserve identity and primitive values. For complete plain data shapes, sealing prevents extensions and makes every public own field non-configurable while keeping its value writable. Existing frozen objects stay frozen.
- `Object.isExtensible` and `Object.isSealed` observe the runtime integrity state. `Object.isFrozen` now uses the same descriptor algorithm: an empty object is neither sealed nor frozen while extensible, but is both after preventing extensions. A nonempty sealed object remains writable and is not frozen.
- A spread of a sealed/nonextensible object must copy it, yielding a fresh extensible result. Runtime object and region initialization clear the new flags; fresh analysis preserves the identity of integrity-call results.
- `Object.getOwnPropertyNames` lists complete plain-shape fields using the existing canonical integer-key ordering. For primitive strings it returns UTF-16 indices followed by the non-enumerable `length`. Numeric and boolean primitives have no own keys. `Object.keys` gets the same primitive paths without `length`.
- Unannotated literal `let` bindings may provide an exact shape when no target assignment anywhere in their module can replace the binding. Direct replacement, destructuring and loop targets stay unproven. Field assignments are conservatively rejected by this proof too. A missing guard permits a hidden string slot to be read as a number; the behavioral mutant proves why the guard is needed.
- Exact null operands in `Object.is` have a statically proven SameValue result, with both operands still evaluated once in order. A union that may contain null is explicitly NotYet because current union tags cannot distinguish null from undefined. No language representation was expanded.

New lowering is in `internal/lower/library_object_integrity.go` and `library_object_names.go`; runtime in `internal/native/runtime/object_integrity.c`, `.h`, and `object_names.c`. The V8 ports are named in `THIRD_PARTY_NOTICES.md`: V8 13.6.233 `src/builtins/builtins-object.cc` and `src/objects/js-objects.cc`, restricted to proven plain data descriptors and primitive fast paths. Arbitrary descriptors, accessors and proxies do not reach the runtime algorithm.

Named shared hooks: `internal/native/library_object.go` dispatches the new methods; `internal/fresh/library_object.go` records returned identity and fresh name arrays; `internal/native/reuse.go` adds the nonextensible guard; `runtime/adamic.h` carries two object flags and includes the owned header; `runtime/region.c` initializes the flags. `runtime/object.c` initializes heap objects. Only the three expanded fixture rows moved in `internal/oracle/counts.md`. `internal/native/emit.go`, `internal/lower/lower.go`, `internal/native/native.go` and `internal/oracle/oracle_test.go` were untouched.

## Mutants

Every listed comparison mutant ran with uncached observations. Each compiled, ran and exited successfully, and differed from Node stdout; compilation failures and sanitizer reports were rejected as proof. Each edit was restored before final measurement and gates. The binding mutant compares a successful release native run to Node directly; its normal source is explicitly refused.

| Mutant | Probe | What caught it |
|---|---|---|
| [own-names-misses-last-unit](integrity-mutants/own-names-misses-last-unit.txt) | library_object_keys.a, `A🌍` UTF-16 boundary | Node stdout only |
| [seal-keeps-configurable](integrity-mutants/seal-keeps-configurable.txt) | library_object_freeze.a, sealed writable fields | Node stdout only |
| [prevent-keeps-extensible](integrity-mutants/prevent-keeps-extensible.txt) | library_object_freeze.a, empty/nonempty objects | Node stdout only |
| [integrity-skips-last-field](integrity-mutants/integrity-skips-last-field.txt) | library_object_freeze.a, one writable public field | Node stdout only |
| [spread-keeps-integrity](integrity-mutants/spread-keeps-integrity.txt) | library_object_freeze.a, fresh spread metadata | Node stdout only |
| [null-equals-undefined](integrity-mutants/null-equals-undefined.txt) | library_object_same.a, exact null and undefined | Node stdout only |
| [replaced-binding-trusted](binding-mutant.txt) | library_object_replaced.a, hidden string field | Node stdout only: Node `2\|built-wrong`, native `2\|4.64724167041843e-310`, both exit 0 |

[Mutant summary](fresh-mutants.txt). Reproducible harnesses: [run-integrity-mutants.py](run-integrity-mutants.py), [run-binding-mutant.py](run-binding-mutant.py). `library_object_replaced.a` is a refused manual probe, not registered as a passing oracle fixture. It does not need a counts row.

## Commands and outputs

All test output was redirected to files. No test run was piped. Each toolchain command sources `/workspace/adamic-tools/env.sh`.

```text
go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
v24.19.0
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (81s)
setup: done in 81s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

[Setup receipt](fresh-setup.txt).

```sh
git fetch origin && git checkout -b codex/library-object origin/main
bash cloud/setup.sh > /tmp/library-object-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic-test262 -adapt -json -test262 /workspace/scratch/test262 -work /tmp/object-fresh-before built-ins/Object > /tmp/object-fresh-before.json 2> /tmp/object-fresh-before.log
go run ./cmd/adamic-test262 -adapt -json -test262 /workspace/scratch/test262 -work /tmp/object-fresh-after built-ins/Object > /tmp/object-fresh-after.json 2> /tmp/object-fresh-after.log
node docs/library-object/classify-tsc.cjs /workspace/scratch/object-tsc/node_modules/typescript /tmp/object-capture/captures.jsonl /tmp/object-tsc-exact.jsonl > /tmp/object-tsc-exact.log 2>&1
python3 docs/library-object/summarize-refusals.py /tmp/object-capture/captures.jsonl /tmp/object-tsc-exact.jsonl
python3 docs/library-object/run-integrity-mutants.py > /tmp/object-integrity-mutants.log 2>&1
python3 docs/library-object/run-binding-mutant.py > /tmp/object-binding-mutant.log 2>&1
go test -count=1 -timeout 30m ./internal/lower ./internal/fresh ./internal/native > /tmp/object-packages-gate.log 2>&1
go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/object-integrity-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 10m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_object|TestCountsAreRecorded' > /tmp/object-final-oracle.log 2>&1
gofmt -l cmd internal > /tmp/object-format-final.log
go vet ./... > /tmp/object-vet-final.log 2>&1
```

Observed exit 0:

- [Package gate](fresh-packages.txt): lower 29.762s, fresh 27.080s, native 89.928s.
- [Final uncached Object oracle plus recorded counts](fresh-oracle.txt): 18.710s. Object fixtures compare source Node, JavaScript backend, sanitized native and release native, with leak checks for successful programs.
- [Counts regeneration](fresh-counts.txt): 30.400s. Exactly keys, freeze and same fixtures changed; successful allocations equal frees in all three.
- [Real TypeScript pass](exact-tsc.txt): all 2317 refused adapted programs checked, TypeScript 6.0.3.
- [Formatting](fresh-format.txt) and [vet](fresh-vet.txt): empty output. `git diff --check` passed.

The full `go test ./...` gate was not run; the complete touched compiler packages and filtered uncached Object oracle/count checks above are the worker gate. Integration's full uncached gate remains with @system_adamic. Source capture can be reproduced with [capture-sources.py](capture-sources.py) without editing the official runner.

During fixture construction, `function value(): null { return null; }` and `function value(): undefined { return undefined; }` were observed to be NotYet. Those are additional language handoff reproducers, outside the baseline refusal counts; they were not implemented. The final evaluation-order probe uses a supported numeric-returning function.

## What remains refused

Descriptor/prototype mutation, dynamic fromEntries/groupBy records, arrays requiring integrity mutation or own-name reflection with holes, boxed/host constructors, builtin constructor objects, optional/widened/returned shape reflection, special prototype/private-like literal names and nullable-union SameValue remain explicit refusals or NotYet. Closing the initial Object call can expose one of these dependencies; candidate counts are not claims of implementability under the fixed-shape contract. No unsupported case was approximated.

## Newly passing tests

- `built-ins/Object/freeze/15.2.3.9-2-1.js`
- `built-ins/Object/freeze/15.2.3.9-3-1.js`
- `built-ins/Object/getOwnPropertyNames/15.2.3.4-1-4.js`
- `built-ins/Object/getOwnPropertyNames/15.2.3.4-1-5.js`
- `built-ins/Object/getOwnPropertyNames/15.2.3.4-1.js`
- `built-ins/Object/getOwnPropertyNames/15.2.3.4-2-3.js`
- `built-ins/Object/getOwnPropertyNames/15.2.3.4-4-38.js`
- `built-ins/Object/getOwnPropertyNames/15.2.3.4-4-b-4.js`
- `built-ins/Object/hasOwn/hasown_own_property_exists.js`
- `built-ins/Object/is/not-same-value-x-y-undefined.js`
- `built-ins/Object/isExtensible/15.2.3.13-1-1.js`
- `built-ins/Object/isExtensible/15.2.3.13-1-2.js`
- `built-ins/Object/isExtensible/15.2.3.13-1-3.js`
- `built-ins/Object/isExtensible/15.2.3.13-1-4.js`
- `built-ins/Object/isExtensible/15.2.3.13-1.js`
- `built-ins/Object/isExtensible/15.2.3.13-2-2.js`
- `built-ins/Object/isExtensible/15.2.3.13-2-22.js`
- `built-ins/Object/isExtensible/15.2.3.13-2-23.js`
- `built-ins/Object/isExtensible/15.2.3.13-2-3.js`
- `built-ins/Object/isSealed/15.2.3.11-1.js`
- `built-ins/Object/keys/15.2.3.14-1-1.js`
- `built-ins/Object/keys/15.2.3.14-1-2.js`
- `built-ins/Object/keys/15.2.3.14-1-3.js`
- `built-ins/Object/keys/15.2.3.14-2-4.js`
- `built-ins/Object/keys/15.2.3.14-2-5.js`
- `built-ins/Object/keys/15.2.3.14-2-6.js`
- `built-ins/Object/keys/15.2.3.14-3-1.js`
- `built-ins/Object/keys/15.2.3.14-3-5.js`
- `built-ins/Object/preventExtensions/15.2.3.10-1-1.js`
- `built-ins/Object/preventExtensions/15.2.3.10-1-2.js`
- `built-ins/Object/preventExtensions/15.2.3.10-1-3.js`
- `built-ins/Object/preventExtensions/15.2.3.10-1-4.js`
- `built-ins/Object/preventExtensions/15.2.3.10-1.js`
- `built-ins/Object/preventExtensions/15.2.3.10-2-1.js`
- `built-ins/Object/preventExtensions/15.2.3.10-3-1.js`
- `built-ins/Object/seal/object-seal-extensible-of-o-is-set-as-false-even-if-o-has-no-own-property.js`
- `built-ins/Object/seal/object-seal-returned-object-is-not-extensible.js`
- `built-ins/Object/seal/seal-boolean-literal.js`
- `built-ins/Object/seal/seal-infinity.js`
- `built-ins/Object/seal/seal-nan.js`
- `built-ins/Object/seal/seal-undefined.js`

## Before per-directory table

| Directory | Pass | Disagreement | Refused | Crashed | Skipped |
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
| built-ins/Object/is | 6 | 1 | 7 | 0 | 7 |
| built-ins/Object/isExtensible | 0 | 0 | 36 | 0 | 2 |
| built-ins/Object/isFrozen | 6 | 0 | 50 | 0 | 3 |
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
| built-ins/Object/prototype/hasOwnProperty | 1 | 0 | 48 | 0 | 14 |
| built-ins/Object/prototype/isPrototypeOf | 0 | 0 | 3 | 0 | 7 |
| built-ins/Object/prototype/propertyIsEnumerable | 0 | 0 | 9 | 0 | 7 |
| built-ins/Object/prototype/toLocaleString | 0 | 0 | 9 | 0 | 3 |
| built-ins/Object/prototype/toString | 0 | 0 | 13 | 0 | 28 |
| built-ins/Object/prototype/valueOf | 0 | 0 | 17 | 0 | 3 |
| built-ins/Object/seal | 0 | 0 | 56 | 0 | 38 |
| built-ins/Object/setPrototypeOf | 0 | 0 | 4 | 0 | 8 |
| built-ins/Object/values | 0 | 0 | 8 | 0 | 12 |

## After per-directory table

| Directory | Pass | Disagreement | Refused | Crashed | Skipped |
|---|---:|---:|---:|---:|---:|
| built-ins/Object | 0 | 0 | 55 | 0 | 6 |
| built-ins/Object/assign | 0 | 0 | 17 | 0 | 21 |
| built-ins/Object/create | 0 | 0 | 280 | 0 | 40 |
| built-ins/Object/defineProperties | 0 | 0 | 361 | 0 | 271 |
| built-ins/Object/defineProperty | 0 | 0 | 722 | 0 | 409 |
| built-ins/Object/entries | 0 | 0 | 10 | 0 | 11 |
| built-ins/Object/freeze | 6 | 0 | 17 | 0 | 30 |
| built-ins/Object/fromEntries | 0 | 0 | 8 | 0 | 17 |
| built-ins/Object/getOwnPropertyDescriptor | 0 | 0 | 306 | 0 | 4 |
| built-ins/Object/getOwnPropertyDescriptors | 0 | 0 | 8 | 0 | 10 |
| built-ins/Object/getOwnPropertyNames | 6 | 0 | 31 | 0 | 8 |
| built-ins/Object/getOwnPropertySymbols | 0 | 0 | 0 | 0 | 12 |
| built-ins/Object/getPrototypeOf | 0 | 0 | 37 | 0 | 2 |
| built-ins/Object/groupBy | 0 | 0 | 10 | 0 | 4 |
| built-ins/Object/hasOwn | 1 | 0 | 46 | 0 | 15 |
| built-ins/Object/internals/DefineOwnProperty | 0 | 0 | 0 | 0 | 6 |
| built-ins/Object/is | 7 | 0 | 7 | 0 | 7 |
| built-ins/Object/isExtensible | 9 | 0 | 27 | 0 | 2 |
| built-ins/Object/isFrozen | 6 | 0 | 50 | 0 | 3 |
| built-ins/Object/isSealed | 1 | 0 | 29 | 0 | 3 |
| built-ins/Object/keys | 8 | 0 | 44 | 0 | 7 |
| built-ins/Object/preventExtensions | 7 | 0 | 7 | 0 | 26 |
| built-ins/Object/prototype | 0 | 0 | 9 | 0 | 6 |
| built-ins/Object/prototype/__defineGetter__ | 0 | 0 | 4 | 0 | 7 |
| built-ins/Object/prototype/__defineSetter__ | 0 | 0 | 4 | 0 | 7 |
| built-ins/Object/prototype/__lookupGetter__ | 0 | 0 | 9 | 0 | 7 |
| built-ins/Object/prototype/__lookupSetter__ | 0 | 0 | 9 | 0 | 7 |
| built-ins/Object/prototype/__proto__ | 0 | 0 | 7 | 0 | 8 |
| built-ins/Object/prototype/constructor | 0 | 0 | 2 | 0 | 0 |
| built-ins/Object/prototype/hasOwnProperty | 1 | 0 | 48 | 0 | 14 |
| built-ins/Object/prototype/isPrototypeOf | 0 | 0 | 3 | 0 | 7 |
| built-ins/Object/prototype/propertyIsEnumerable | 0 | 0 | 9 | 0 | 7 |
| built-ins/Object/prototype/toLocaleString | 0 | 0 | 9 | 0 | 3 |
| built-ins/Object/prototype/toString | 0 | 0 | 13 | 0 | 28 |
| built-ins/Object/prototype/valueOf | 0 | 0 | 17 | 0 | 3 |
| built-ins/Object/seal | 6 | 0 | 50 | 0 | 38 |
| built-ins/Object/setPrototypeOf | 0 | 0 | 4 | 0 | 8 |
| built-ins/Object/values | 0 | 0 | 8 | 0 | 12 |
