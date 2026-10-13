Built step 18 / Outcome 24: function slots use the shared ClosureTargets producer facts, with transitive capture checks and conservative unknown origins.
Commits: 14d3f49b implementation, b2d6e207 landed-census consolidation, 39adfaa5 and 98eb464b current-main merges; delivery evidence follows.
Checks: ten R2 fixtures agree across Node, JavaScript, release native and sanitized native with leaks enabled; 1,346-program admission delta clean; counts, guard and focused tests pass.
Mutants: omitted boundary inputs, unknown-as-empty, omitted captures, dropped return target and omitted Map captures each fail the named assertion below.
Not covered: whole packages/full gate, the entire native parser, or a more precise per-instance collection analysis; collection and field target pools remain conservative.

The implementation keeps one cached closure-producer analysis in IR. Lowering readers and the cycle decision consult its facts. Parameter, return, collection and field flows are bounded only when every origin is named; unknown invocations and foreign inputs poison target sets. Captured objects retain type-graph traversal, and captured function slots recurse through the same producer facts. Pooling collections and fields by slot kind/name can refuse otherwise safe programs. This is the conservative interpretation used in the implementation commit.

The ten test-only R2 fixtures came from compiler/getters-census 9471e0a4. They were replayed against main production sources before the change and all ten refused with adamic/cycle-capable (red-replayed.log). Main subsequently landed the census tests, so their existing top-level TestGettersCensus functions now expect success. No production census changes were imported. Three local-memoizer witnesses pin the exact existing refusal: direct capture, joined targets with one cyclic closure, and indirect capture through a Map of functions. docs/memory.md includes the direct and indirect programs.

Admission comparison used standalone base/head compilers and every .a/.ts under internal/oracle/testdata. Base production sources are 88ec8b6e; the subsequent main changes through 79f2067b changed census/tests/evidence, not these compiler production sources. All 1,346 programs were classified. Exactly the ten R2 fixtures became admitted, none became refused, and every newly admitted program matched Node stdout and exit in JavaScript and release native. Sanitized native agreement and LeakSanitizer checks are covered separately by the ten census oracle leaves. Full observations are in admission-delta.json; runner in run-admission-delta.py.

Mutants were applied individually and restored in finally blocks. The runner demands assertion failure, not a build failure. Sources are .patch evidence:

- omit-boundary-inputs: TestClosureTargetsKnownCallerDoesNotHideUnknownCaller catches a known call concealing an unbounded caller.
- unknown-is-empty: TestClosureTargetsBoundOnlyProvenValues catches missing unknown-origin proof.
- omit-captures: TestClosureCycleDirect catches the direct memoizer cycle being admitted.
- drop-return-target: TestClosureTargetsFollowParametersReturnsAndMapValues catches omission of one returned target, including the throwing target.
- omit-map-captures: TestClosureCycleMap catches the indirect Map cycle being admitted.

Commands were run with ADAMIC_GOCACHE_OFF=1 and source /workspace/adamic-tools/env.sh. Every test output was redirected to a log, and commands had hard limits. Final successful commands:

```sh
# merged-oracle.log, PASS 20.091s; parallelism 4, uncached oracle runs
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestGettersCensus|^TestClosureCycle|^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(route_targets_sort_values|borrow_chain_unknown|call_targets_element|generic_functions|user_iterators)\.a$' -count=1 -timeout 90s -parallel 4 -v
# merged-counts.log, PASS 77.632s
go test ./internal/oracle -run TestCountsAreRecorded -timeout 90s -args -update-counts
# merged-targets.log, PASS 0.544s; reader leaf 0.53s
go test ./internal/ir -run '^TestClosureTargets|^TestCallTargetReaders$' -count=1 -timeout 90s -v
# element-plan.log, PASS 0.044s
go test ./internal/native -run '^TestCallTargetsElementBorrowPlan$' -count=1 -timeout 90s
# vet.log, exit 0
timeout 90 go vet ./internal/ir ./internal/lower ./internal/oracle
# merged-admission-delta.log: 1346 programs; 10 newly admitted, clean
timeout 480 python3 review/compiler/cycle-closure-targets/run-admission-delta.py
# mutants-final.log: all five caught
timeout 720 python3 review/compiler/cycle-closure-targets/run-mutants.py
# required lane command, lane-checks.log
 git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

The final lane script passed in 0.9s: gofmt/tools on six Go files, t.Parallel on two test packages, and vet on two packages. The earlier lane run skipped vet after its budget; explicit vet above also completed successfully. No test reads guarded call fields directly without the permitted IR interface. Latest main 79f2067b was merged on this branch; its stage3 census changes do not alter the oracle corpus or compiler under test.

Counts were refreshed for ten admitted fixtures. More precise shared target sets also changed five existing count rows: route_targets_sort_values retains 92→96/releases 142→150; borrow_chain_unknown 3→2/8→7; call_targets_element 32→30/60→58; generic_functions 32→29/55→52; user_iterators 455→442/904→891. Allocation, free, peak and region counts did not change for these rows. All five fixtures were explicitly checked against Node in both backends with native sanitizers in the final oracle command.

Setup: nproc reported 5 (the worker CPU quota is 4). GOPROXY was https://proxy.golang.org|direct. cloud/setup.sh reached Go 0.093s, Node 0.108s, clang 0.813s, markdown 1.374s, submodules 17.642s, shared cache 24.721s, then hit the 240s limit warming the cache. Initial cold builds also hit their limits. Disabling the optional remote Go cache allowed the build to finish. npm ci --prefix stage3/api installed the locked Node typings. setup.log and node-types.log preserve output. Cache startup and reconciling landed census tests delayed delivery beyond the requested first-push window.

New/touched oracle leaf durations (including setup and all backends), each below 60s:

```text
--- PASS: TestClosureCycleDirect (0.87s)
--- PASS: TestGettersCensusHigherCallbackPairMethods (0.55s)
--- PASS: TestGettersCensusCallbackPairMethods (0.48s)
--- PASS: TestGettersCensusSnapshotNext (0.41s)
--- PASS: TestGettersCensusBinaryFunction (2.04s)
--- PASS: TestGettersCensusClosureUnionNext (0.51s)
--- PASS: TestGettersCensusClosureNumberNext (0.35s)
--- PASS: TestGettersCensusClassMap (1.61s)
--- PASS: TestGettersCensusOverloadedFunction (0.59s)
--- PASS: TestGettersCensusThrowingObject (1.52s)
--- PASS: TestGettersCensusUnaryFunction (1.65s)
--- PASS: TestGettersCensusUpdateNodeFunction (1.80s)
--- PASS: TestGettersCensusUpdateCommentFunction (1.57s)
--- PASS: TestGettersCensusOptionalCommentFunction (1.41s)
--- PASS: TestGettersCensusUpdateTypeFunction (1.42s)
--- PASS: TestGettersCensusOptionalTypeFunction (1.31s)
--- PASS: TestGettersCensusUnaryNodeFunction (1.22s)
--- PASS: TestGettersCensusLazyObject (0.44s)
--- PASS: TestClosureCycleMap (0.39s)
--- PASS: TestGettersCensusClosureNumberMethods (0.36s)
--- PASS: TestGettersCensusPrimaryFunction (1.20s)
--- PASS: TestGettersCensusOptionalBooleanFunction (1.31s)
--- PASS: TestClosureCycleJoined (0.43s)
--- PASS: TestGettersCensusBinaryInline (20.03s)
```

The new IR leaves are effectively 0.00s in merged-targets.log. No PR was opened. This lands the cycle-membership refinement toward step 18 and Outcome 24; it does not claim completion of the full native parser roadmap.
