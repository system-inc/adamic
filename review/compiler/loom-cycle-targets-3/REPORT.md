Re-cut closure-target cycle analysis as one commit on current main.
Commit: delivery SHA is in the final response; base d79302e50c022c997c00f7c68ddb7475f9ee65fc; source de92dfd9b6a63b8062bc49e5bc6ddf5f80d53b98.
Commands: focused witnesses and four guards PASS (35 events, maximum 25.62s); counts PASS 58.340s; readers PASS 23.06s; five affected fixtures PASS uncached; consumer tests and changed-package vet PASS.
Mutants: all six recorded mutants caught by their intended assertions, with every production file restored.
Not covered: full gate, whole packages, unrelated consumer leaves, WASI, benchmarks, pristine-source cycle ledger or a renewed compiler admission census. No PR or Loom submission.

There were no textual squash conflicts. Main did not move between the two branch fetches. The branches are independent single commits on the same main SHA; this branch contains none of checked-writes. Main compiler infrastructure and test products are retained.

The only production adjustment beyond the source squash is two lines in internal/ir/call_targets.go: closureFacts.contents and its allocation use SortedMap[string, FunctionTargets]. Conservative assumption: the Program-held-map rule applies to this private cached map too, even though gob omits that cache. The source algorithm, unknown-input fallback, producer flow, capture traversal and admission behavior are retained. No other source correction was needed.

Counts: take main counts.md, then exactly one successful full -update-counts on this Linux box with house clang 20.1.8. Main 1f07ce2d has already moved digests to measurement-product keys and restored counts-only rows; preserve that current infrastructure rather than reintroducing the historical digest column. The regenerated table changes five existing rows, adds eleven, and removes none. Every numeric delta and added row equals the source branch. There are no unrelated row changes. The changed rows are below and in counts-row-changes.json; column order is allocations/frees/retains/releases/peak/regions.

| Fixture | Main | Re-cut |
|---|---|---|
| internal/oracle/testdata/route_targets_sort_values.a | 84/84/92/142/17/0 | 84/84/96/150/17/0 |
| internal/oracle/testdata/borrow_chain_unknown.a | 8/8/3/8/7/0 | 8/8/2/7/7/0 |
| internal/oracle/testdata/call_targets_element.a | 45/43/32/60/7/2 | 45/43/30/58/7/2 |
| internal/oracle/testdata/generic_functions.a | 35/35/32/55/8/0 | 35/35/29/52/8/0 |
| stage3/namespace-init-sys/cwd.a | new | 6/6/8/12/6/0 |
| internal/oracle/testdata/getters_census_binary_function.a | new | 39/39/37/62/25/0 |
| internal/oracle/testdata/getters_census_unary_function.a | new | 36/36/34/56/24/0 |
| internal/oracle/testdata/getters_census_primary_function.a | new | 33/33/31/50/23/0 |
| internal/oracle/testdata/getters_census_unary_node_function.a | new | 36/36/34/56/24/0 |
| internal/oracle/testdata/getters_census_update_node_function.a | new | 39/39/40/65/25/0 |
| internal/oracle/testdata/getters_census_optional_boolean_function.a | new | 36/36/34/56/24/0 |
| internal/oracle/testdata/getters_census_optional_type_function.a | new | 36/36/40/62/24/0 |
| internal/oracle/testdata/getters_census_update_type_function.a | new | 39/39/43/68/25/0 |
| internal/oracle/testdata/getters_census_optional_comment_function.a | new | 33/33/37/56/23/0 |
| internal/oracle/testdata/getters_census_update_comment_function.a | new | 36/36/40/62/24/0 |
| internal/oracle/testdata/user_iterators.a | 663/663/455/904/64/0 | 663/663/442/891/64/0 |

Runtime C delta: source and re-cut both have zero +/- lines under internal/native/runtime, so they are identical. Differences: none. runtime-delta-comparison.json records the source merge base and comparison.

Required guards:

- TestEveryMapAProgramHoldsIsSorted passes; manually converted the newly Program-held private contents map as well.
- TestRuntimeSetsAreEachUsed passes. Every required feature set already has an exact witness; no set or witness edit is needed.
- TestNativeAgreementRowsNameRegisteredPrograms passes. The lower cwd test uses the existing JavaScript agreement helper, while TestClosureCycleAmbientCwd supplies native/sanitizer agreement. No native helper call bypasses its literal registry.
- TestEveryTestIsParallelOrSaysWhy passes. All added/touched test functions call t.Parallel first; no serial exception was introduced.
- TestCallTargetReaders passes in 23.06s with -timeout 90s. The original call-target descendant and gob guards pass too.
- No new compiler-consuming package was introduced; existing declarations in cloud/fast-gate/compiler-dependencies.json are retained. No new product key names a build directory or machine path.
- Every completed own test is below 60s on GOMAXPROCS=4, -parallel=4 and cgroup cpu.max=400000 100000. No leaf needed splitting or deletion. The full per-event ledger is test-seconds.json.

The focused suite holds the ten newly admitted getter memoizers and ambient cwd to Node, JavaScript, release native, ASan/UBSan and leak checks. Direct, joined and map back-reference cycles still get their exact cycle refusal. Five affected existing fixtures were separately run with ADAMIC_GATE_UNCACHED=1: route_targets_sort_values.a, borrow_chain_unknown.a, call_targets_element.a, generic_functions.a and user_iterators.a. All five actual fixture bodies ran and passed (maximum event 7.71s); numbered wrappers and the user_iterators solo wrapper are named in affected-fixtures.log.

Counts update: 1069 passing events, maximum 5.23s. Counts table structure and generated-family ownership/currentness pass afterward. All three changed Go packages were vetted, with no output. Focused consumer checks exercise native borrowing, flow, fresh analysis and JavaScript. Setup builds every package; tests remain focused as requested.

Toolchain: export GOPROXY=https://proxy.golang.org|direct, source /workspace/adamic-tools/env.sh and set GOMAXPROCS=4. setup passed: Go .017s; Node .020s; submodules .058s; markdown .069s; clang .166s; shared cache 10.876s; Go build 48.401s; cache warm 48.495s; done 48.524s. nproc=5, quota 4 CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0. No setup workaround was needed on this branch, and no cohere code was copied.

Commands and outputs are retained in this directory. All test output was redirected directly to logs. Each command has an outer timeout; each go test below uses -count=1, -parallel=4 where relevant and -timeout 90s unless counts regeneration needs its family-wide 570s limit. Individual count events remain below 60s.

```text
timeout 240 bash cloud/setup.sh
timeout 180 go test ./internal/ir ./internal/lower ./internal/oracle ./cmd/adamic-gate -run "^(TestEveryMapAProgramHoldsIsSorted|TestEveryTestIsParallelOrSaysWhy|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed|TestClosureTargets.*|TestNamespaceAmbient.*|TestClosureCycle.*|TestGettersCensus.*)$" -count=1 -parallel=4 -json -timeout 90s
timeout 600 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -json -timeout 570s -args -update-counts
timeout 600 python3 review/compiler/loom-cycle-targets-3/run-mutants.py
timeout 180 python3 review/compiler/loom-cycle-targets-3/run-admission-mutant.py
timeout 120 go test ./internal/ir -run "^(TestCallTargetReaders|TestCallTargetsIncludeEveryDescendant|TestEveryMapAProgramHoldsIsSorted|TestAProgramIsTheSameBytesEveryTimeGobWritesIt|TestASortedMapIsNilWhereAPlainMapIs)$" -count=1 -parallel=4 -json -timeout 90s
timeout 150 go vet ./internal/ir ./internal/lower ./internal/oracle
ADAMIC_GATE_UNCACHED=1 timeout 150 go test ./internal/oracle -run '^(TestNativeAgreesWithNode_[0-9]+|TestNativeAgreesWithNode_user_iterators)$/^internal$/^oracle$/^testdata$/^(route_targets_sort_values|borrow_chain_unknown|call_targets_element|generic_functions|user_iterators)\.a$' -count=1 -parallel=4 -json -timeout 90s
timeout 150 go test ./internal/native ./internal/flow ./internal/fresh ./internal/javascript -run "^(TestCallTargetsElementBorrowPlan|TestBorrowChainTargets|TestLoopCallCoverage|TestDebuggerHasNoFlowInstruction|TestRegexOperandsStillJudgeCycleWrites|TestViewMixedUnionSelectsNumber)$" -count=1 -parallel=4 -json -timeout 90s
timeout 120 go test ./internal/oracle -run ^TestCountsAreRecordedTable$ -count=1 -timeout 90s
timeout 120 go test ./internal/oracle -run "^(TestFamiliesOwnEveryCaseOnce|TestFamilyWrappersAreCurrent)$" -count=1 -timeout 90s
```

Every mutant and its intended catcher (assertion failures, no build-error or timeout kill):

| Mutant | Catcher | Assertion |
|---|---|---|
| omit-boundary-inputs | TestClosureTargetsKnownCallerDoesNotHideUnknownCaller | known caller hid unbounded caller |
| unknown-is-empty | TestClosureTargetsBoundOnlyProvenValues | bounded without proof |
| omit-captures | TestClosureCycleDirect | want cycle refusal |
| drop-return-target | TestClosureTargetsFollowParametersReturnsAndMapValues | want both returned targets |
| omit-map-captures | TestClosureCycleMap | want cycle refusal |
| type-matched-captures | TestNamespaceAmbientCwdInitialization and TestClosureCycleAmbientCwd | adamic/cycle-capable |

Every focused test event and seconds:

| Test | Seconds |
|---|---:|
| TestEveryMapAProgramHoldsIsSorted | 0 |
| TestClosureTargetsKnownCallerDoesNotHideUnknownCaller | 0 |
| TestClosureTargetsFollowParametersReturnsAndMapValues | 0 |
| TestClosureTargetsBoundOnlyProvenValues | 0 |
| TestNativeAgreementRowsNameRegisteredPrograms | 0.06 |
| TestEveryTestIsParallelOrSaysWhy | 0.52 |
| TestNamespaceAmbientContextsDoNotExecute | 0.12 |
| TestNamespaceAmbientHostInitialization | 0.76 |
| TestNamespaceAmbientCwdInitialization | 0.86 |
| TestGettersCensusSnapshotNext | 14.48 |
| TestClosureCycleDirect | 14.49 |
| TestGettersCensusClosureUnionNext | 0.18 |
| TestGettersCensusClosureNumberNext | 0.18 |
| TestGettersCensusClassMap | 14.86 |
| TestGettersCensusOverloadedFunction | 0.16 |
| TestGettersCensusThrowingObject | 10.94 |
| TestGettersCensusBinaryFunction | 25.62 |
| TestGettersCensusUnaryFunction | 10.95 |
| TestGettersCensusOptionalBooleanFunction | 10.61 |
| TestGettersCensusUpdateTypeFunction | 0.57 |
| TestGettersCensusOptionalCommentFunction | 0.57 |
| TestGettersCensusUpdateCommentFunction | 0.6 |
| TestGettersCensusOptionalTypeFunction | 0.6 |
| TestGettersCensusClosureNumberMethods | 0.19 |
| TestGettersCensusLazyObject | 0.16 |
| TestGettersCensusUnaryNodeFunction | 0.59 |
| TestGettersCensusUpdateNodeFunction | 0.59 |
| TestGettersCensusPrimaryFunction | 0.6 |
| TestClosureCycleMap | 0.2 |
| TestGettersCensusHigherCallbackPairMethods | 0.24 |
| TestGettersCensusCallbackPairMethods | 0.22 |
| TestClosureCycleJoined | 0.16 |
| TestRuntimeSetsAreEachUsed | 1.42 |
| TestGettersCensusBinaryInline | 11.46 |
| TestClosureCycleAmbientCwd | 12.1 |

Formatting: changed Go files are gofmt-clean and strict diff whitespace checks pass for docs/internal. Inherited .patch evidence retains unified-diff context whitespace. New logs are explicitly staged within this branch's review directory.
