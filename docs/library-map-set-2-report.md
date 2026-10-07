# Map and Set second library report

Built all seven Set methods with concrete Map operands and accepted never-returning Map/Set forEach callbacks.
Claim commit: `2d45095`; implementation is the commit containing this report.
Linux checks: lower 31.403s, flow 71.278s, library oracle 22.311s, counts 30.539s; vet and formatting clean.
Eight new mutants compiled and exited 0 without sanitizer output; only Node stdout comparison caught them.
Language gaps remain refused; the full repository gate was not run. See the complete [claim and compiler handoff](library-map-set-2-claim.md).

## Measurement

Runner source: `origin/codex/test262-ts-validity` at `af128990588aab7ee52ea3ab8527d638009c4979`.
test262: `c8c798898646638cd0c24879f8e0374e847e7d74`. TypeScript 6.0.3, Node 24.19.0.
The separately compiled runner uses this checkout for lowering and the runtime; no runner changes are committed.

| Stage | Slice | Pass | Disagreement | Refused | Not TypeScript | Crashed | Skipped |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Before | Set | 19 | 0 | 125 | 80 | 0 | 159 |
| Before | Map | 6 | 0 | 32 | 45 | 0 | 121 |
| After | Set | 23 | 0 | 121 | 80 | 0 | 159 |
| After | Map | 7 | 0 | 31 | 45 | 0 | 121 |

Every new pass agreed with source Node under the runner's sanitizer build:
- `built-ins/Set/prototype/forEach/throws-when-callback-throws.js`
- `built-ins/Set/prototype/isDisjointFrom/compares-Map.js`
- `built-ins/Set/prototype/isSubsetOf/compares-Map.js`
- `built-ins/Set/prototype/isSupersetOf/compares-Map.js`
- `built-ins/Map/prototype/forEach/callback-result-is-abrupt.js`

The four result-producing combines-Map tests still refuse at builtin instanceof. The new map-operands fixture proves the results against Node directly, including both size branches, insertion order, ties, zero, NaN, empty operands, object identity, strings, booleans and optional numeric keys.
The new forEach fixture checks abrupt propagation after entry deletion, dynamically allocated strings, thisArg evaluation, and empty iteration. It agrees in sanitized native, release native and the JavaScript backend, and its allocation count balances.

## Commands and artifacts

All test output was redirected to logs. Environment: `source /workspace/adamic-tools/env.sh`. Setup used `bash cloud/setup.sh`: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 61s, total 61s, nproc 5.

```sh
/tmp/map-set-2-runner -adapt -json -test262 /tmp/map-set-test262 built-ins/Set built-ins/Map > /tmp/map-set-2-before.json 2> /tmp/map-set-2-before.log
/tmp/map-set-2-runner -adapt -json -test262 /tmp/map-set-test262 built-ins/Set built-ins/Map > /tmp/map-set-2-after.json 2> /tmp/map-set-2-after.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/flow -count=1 -timeout 10m > /tmp/map-set-2-packages.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestLibraryMapSet|TestNativeAgreesWithNode/internal/oracle/testdata/library_map_set' -count=1 -v -timeout 10m > /tmp/map-set-2-library-oracle.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/map-set-2-counts.log 2>&1
go vet ./... > /tmp/map-set-2-vet.log 2>&1
gofmt -l cmd internal > /tmp/map-set-2-gofmt.log
git diff --check > /tmp/map-set-2-diff-check.log
```

All commands above exited 0. The exact bounded gate is reported here; packages outside lower/flow and the filtered library oracle were not tested.

## Mutants

| Method family | Mutation | Sole catching check |
| --- | --- | --- |
| union | Swap generated left/right parameters | Node stdout comparison |
| intersection | Swap generated left/right parameters | Node stdout comparison |
| difference | Swap generated left/right parameters | Node stdout comparison |
| symmetricDifference | Swap generated left/right parameters | Node stdout comparison |
| isSubsetOf | Swap generated left/right parameters | Node stdout comparison |
| isSupersetOf | Swap generated left/right parameters | Node stdout comparison |
| isDisjointFrom | Return false on the successful disjoint path | Node stdout comparison |
| forEach | Iterator reports exhaustion before visiting any callback | Node stdout comparison |

All eight mutations changed real generated code. Each passed compilation with Werror, ran under ASan/UBSan, exited 0 and produced no stderr. Existing Map/Set mutants and iterator-copy refusal probes also passed.

## Scope

No runtime change was needed: existing live Map iteration and exception cleanup implement these operations. No new V8 algorithm was ported, so THIRD_PARTY_NOTICES.md did not change. Compiler-owned object.go, cycles.go, lower.go, native.go, emit.go and oracle_test.go were untouched.
General set-like objects, array iterator protocols, dynamic receivers and builtin prototype identity remain refused rather than approximated. The claim records each language feature with a one-line reproducer, all 157 baseline paths, and the seven runner diagnostic-code mismatches.
