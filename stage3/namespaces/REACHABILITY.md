Built: namespace initialization follows known calls and leaves unresolved accesses checked; the exact core Map probe prints 0:false, matching Node.
Commits: main merges ab6cead9 and 2751606d; implementation d7519a64b8bae99ddfa82f12ff36e6a4c479b479; delivery includes this report commit.
Validation: complete lower/load pass; final uncached namespace and signal oracle passes in 15.203s; counts pass in 51.274s; matrix 6 Compiles, 3 NotYet, 3 Refused.
Mutants: reaching-as-unreaching, missing runtime readiness, and inlining away the namespace const-enum read are each caught and restored.
Limits: full parser execution, namespace containers and merges, partial export observation, and general never-value representation are not established; main and area branches were untouched.

## Implementation

The former global pending count and blanket call/new refusal are removed. The static walk follows function declarations and immutable initializer chains by checker symbol, including imported aliases, callbacks, parameter defaults and recursive helpers. Active nodes break graph cycles without hiding later reads in a body. Functions and instance initializers remain deferred; static class initialization and extends expressions execute at declaration.

A statically reached read of a pending namespace remains located NotYet. Parameters, mutable function values, virtual methods and constructors leave unresolved edges. Their eventual qualified namespace accesses use a private generated boolean initialized false before module evaluation and set true after the namespace body. Ordinary IR checks that boolean and panics loudly; both backends and their flow/ownership passes see the calls and checks. This is not a namespace object and does not admit escape, reflection or merging. Private singleton reads retain the existing ready checks.

Readiness covers state, function-value reads, direct generic/non-generic calls, void calls, writes and const-enum members. Inlining a constant must not erase an earlier namespace-container read. The callee is checked before argument effects. A simple property write evaluates its RHS before the final PutValue check; nested containers are read before the RHS. Compound writes and increments check before reading the old value. Whole-body readiness is conservative: unresolved access to a partly assigned namespace stops even if an individual export may already exist. No partial-container semantics are claimed.

The Map probe then exposed a separate missing representation for never-key/value maps. An uninhabited key/value type uses an unused object storage slot, allowing an empty Map and readonly wider observations without inventing a never value. Writable widening from Map<never, never> to Map<string, number> remains Refused and is tested. Functions and values typed never still have their existing boundaries.

## Exact parser probe

Unchanged bytes from codex/stage3-parser-proof 77aaea49, stage3/drivers/parser/native-map-before-namespace.a, are registered as namespaces_map_before.a:

```typescript
// From TypeScript 6.0.3, src/compiler/core.ts:19
// From TypeScript 6.0.3, src/compiler/debug.ts:26
// Minimal namespace-initialization preflight stop: this Map call does not read Debug.
const emptyMap: ReadonlyMap<never, never> = new Map<never, never>();
namespace Debug {
    export let isDebugging = false;
}
console.log(`${emptyMap.size}:${Debug.isDebugging}`);
```

Node, sanitized native, release native and checked JavaScript all exit 0, print `0:false\n`, and have empty stderr. Successful executions pass leak checks. The former initialization stop at 4:45 is gone.

## Direct and helper reaching calls

The two exact programs are committed in internal/lower/testdata/namespaces_notyet/reaching_direct.a and reaching_helper.a:

```typescript
function read(): boolean { return Debug.isDebugging; }
console.log(`${read()}`);
namespace Debug { export let isDebugging = false; }
```

```typescript
function read(): boolean { return Debug.isDebugging; }
function helper(): boolean { return read(); }
console.log(`${helper()}`);
namespace Debug { export let isDebugging = false; }
```

Both builds exit 1, at the reached read at 1:35, with this pinned text:

```
stage 0 can't lower a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration yet
```

Independent Node exits 70 with empty stdout and `adamic: panic: TypeError: Cannot read properties of undefined (reading 'isDebugging')`. The oracle runtime normalizes Node's uncaught error to 70. The refusal is sound for the admitted qualified-name subset: it avoids manufacturing an initialized export or namespace object. It is intentionally conservative for partial namespace bodies. Exact CLI and Node observations are in /tmp/namespaces-reach-{direct,helper}-{node,build}.json.

## Runtime fallback fixtures

All fixtures below are source-Node comparisons, including stderr and exit code, rather than exceptions to the ordinary differential oracle.

| Fixture | Node and both backends |
| --- | --- |
| namespaces_safe_initialization.a | Exit 0; `0`, `0:true:false`, `false`, `0:true:false` on successive lines |
| namespaces_unknown_before.a | Exit 70; no stdout; TypeError reading isDebugging |
| namespaces_unknown_function_before.a | Exit 70; no stdout; TypeError reading answer, before the argument can print right |
| namespaces_unknown_void_before.a | Exit 70; no stdout; TypeError reading log, before the argument can print right |
| namespaces_unknown_write_before.a | Exit 70; stdout `right`; TypeError setting isDebugging, after the RHS effect |
| namespaces_unknown_enum_before.a | Exit 70; no stdout; TypeError reading Level, despite const-member inlining |

The safe fixture includes a pure helper constructing a Map before Debug, the same helper in a namespace initializer/body, a class whose instance field is deferred until after Debug, and a readonly wider view of an empty never map. Refusal tests additionally pin const aliases, callbacks, default parameters, recursion and static class helper calls.

## Mutants

Run with the setup environment sourced:

```
python3 stage3/namespaces/reachability-mutants.py
```

| Actual compiler mutation | Catcher |
| --- | --- |
| Resolve a reaching function declaration as no target | Direct/helper lower tests fail: the promised initialization NotYet becomes successful lowering |
| Drop runtime namespace-readiness operands | Node exits 70 while native and checked JavaScript exit 0 and print right then true |
| Remove only the const-enum namespace check hook | Node exits 70 while native and checked JavaScript exit 0 and print 1 |

All three are restored in finally blocks and caught without a compiler-warning or sanitizer build failure. Logs: /tmp/namespaces-reach-final-mutants.log and /tmp/namespaces-reach-mutants/*.log. Existing namespace semantic/state mutants also pass on the final IR. The exported-state mutant now changes the RHS snapshot feeding the write. The order fixture has a second post-initialization output, so swapping two outputs disagrees with Node cleanly without moving a qualified call before readiness.

## Integration and matrix

Started on f893faf2. Merged main 48c05d091f0a43c31cbe051b1d6578d99eeedf19, resolving enum and cycle-policy conflicts explicitly. Main's open numeric-enum and checked-cast policies are retained; namespace-scoped enum acceptance and parameter properties remain. The shared localRead helper retains main's union-tag and module-cycle checks and can now return its NotYet error; the scalar returned-assignment caller is updated accordingly. No protected compiler entry/emitter files were manually edited.

At completion main had advanced to ce0750f28ef3943057f1f852b3ae5d93e6c5d644. It was merged without conflicts in 2751606d131e0a350a153ced2e982ccf3c9c9c4d. Its runtime signal-reset change passes all six Node signal cases, including inherited ignored handlers. No force push, main/area merge, main/area push or PR was performed. No cohere code was copied; its pinned submodule supplies the rule runner and checker.

The original twelve source fixtures were not edited. The matrix at implementation d7519a64 is **6 Compiles, 3 NotYet, 3 Refused**, with every compiled row matching fresh Node, checked JavaScript and native ASan/UBSan/leak runs. Changes relative to f893faf2 are inherited main policies:

- 09_incremental_parser: Refused to Compiles because main opens whole numeric enums and no longer demands exhaustive numeric-enum switches. Its post-switch return remains unchanged. Output: `11 true -1`, `9 true -1`, `80 true -1`, `1 false -1`.
- 02_jsx_names and 03_react_names: NotYet to Refused because main rejects their primitive-brand casts as runtime-uncheckable casts.
- 06_binary_expression_state retains its exact generic-signature refusal. All other original rows retain their category.

This unit clears the separate Map discovery probe; no original twelve-slice row moved because of the initialization change. Five of ten normalized declaration shapes still lower. Debug's next normalized blocker is its namespace class at 93:1; Parser's is a bodyless function at 237:1. All ten shape results and the test expectations are refreshed. These are declaration-shape observations, not full implementations.

## Counts and validation

New count rows use allocations/frees/retains/releases/peak/regions:

- Map probe: 3/3/0/4/3/0.
- Safe initialization: 12/12/2/16/3/0.
- Each of the five stopping indirect-read/write fixtures: 2/1/2/3/2/0. Panic stops with an owned value, as the existing panic contract permits.

Existing namespace rows preserve allocations and frees. Retain/release pairs rise for observed narrowing 5/14 to 9/18, Debug modules 1/8 to 3/10, Debug state 8/23 to 13/28, namespaces 16/56 to 18/58 and namespace modules 3/15 to 5/17. The checked helper argument/return paths and RHS snapshots retain values; Debug module/state peak rises from 3 to 4. Readiness booleans themselves allocate no heap objects. No claim of a performance improvement is made.

Final commands, with /workspace/adamic-tools/env.sh sourced, write directly to logs:

```
go test ./internal/lower ./internal/load -count=1 -timeout 30m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNamespace|TestNativeAgreesWithNode/internal/oracle/testdata/namespaces|TestNativeAgreesWithNode/stage3/namespaces/shapes|TestASignalLeavesWhatWasPrinted' -count=1 -timeout 30m -v
python3 stage3/namespaces/progress-matrix.py --label 'Reachable namespace initialization and current main' --scratch /tmp/namespaces-reach-delivery-matrix
go vet ./...
gofmt -l cmd internal
git diff --check
```

Final lower 45.865s and load 2.240s pass; counts 51.274s pass; uncached oracle 15.203s passes, with native hits 0/misses 63, Node hits 0/misses 57 and probe hits 0/misses 6. The matrix passes. Vet, formatting and own-change whitespace checks are clean. Logs: /tmp/namespaces-reach-delivery-{lower-load,counts,oracle,matrix,vet,format}.log.

A broader affected-package command was also run uncached:

```
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/load ./internal/lower ./internal/native ./internal/oracle ./internal/ir ./internal/flow ./internal/fresh
```

Native 475.859s, IR 19.577s, flow 297.453s and fresh 128.671s pass. That initial command exits 1: the new direct static-field test was rejected by the checker before lowering (TS2729), and two old namespace mutants no longer matched the changed IR or the clean-exit precondition. The test now reaches the field through a checker-accepted helper, and the mutants are adapted as described above. Full lower/load and the affected namespace oracle/mutants were re-green on the final code. The whole Oracle package and whole repository gate were not repeated after those test repairs; the final filtered gate also covers the const-enum fix and refreshed-main runtime. Broader log: /tmp/namespaces-reach-packages.log.

Exploratory failures are retained: the missing never-map representation, an initial ReferenceError rather than Node's TypeError, and initially checking a simple write before its RHS were all corrected and re-oracled. A parenthesized const-enum receiver experiment reaches an upstream checker panic, `Unhandled case in GetFirstIdentifier`, during checkConstEnumAccess before lowering. Its exact scratch source is /tmp/namespaces-reach-parenthesized-enum.a and log is /tmp/namespaces-reach-enum-parentheses-oracle.log. This unit does not fix that checker issue or claim coverage of it.

## Toolchain

Initial setup failed with `no required module provides package github.com/system-inc/cohere/rule_runner` while the submodule checkout still preceded main's pinned runner. Updating the existing submodule to main's pin fixed it; no runner file was copied. Retry setup succeeds: Go 0.057s, Node 0.064s, submodules 0.156s, markdown 0.188s, clang 0.343s, Go build 52.757s, cache warm 52.993s, total 53.038s. nproc 5; CPU quota four; Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup logs: /tmp/namespaces-reach-setup.log and /tmp/namespaces-reach-setup-retry.log.
