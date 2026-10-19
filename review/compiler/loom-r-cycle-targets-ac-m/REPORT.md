Built compiler/loom-r-cycle-targets-ac-m on current fetched main, preserving producer-based cycle membership and slice 2 adapters.
Source b1d42ea3917096b83841c570f05185cbe9a0b337; base 67733635b0252eb9d3bb3645c72262c7b18361e1; final delivery SHA is in the response.
Counts regenerated once on Linux clang 20.1.8; package, guard, vet, formatting and a-check results follow below.
All six recorded implementation mutants, one new Effects-resolution mutant and both header mutants were caught by their intended checks.
Not covered: full repository gate, full-tree a-check, ahra/Loom submission, the opt-in WASI integration matrix or a renewed whole admission census.

The newest origin/main at fetch was exactly 67733635b0252eb9d3bb3645c72262c7b18361e1. Apply only b1d42ea391's own parent-to-tip change with git merge-tree --write-tree --merge-base=b1d42ea3917096b83841c570f05185cbe9a0b337~1 origin/main b1d42ea3917096b83841c570f05185cbe9a0b337, then cherry-pick --no-commit. No main or existing source branch is pushed. The existing queue item chg_9zj2ggjbqsxqxz66h9xbvv1bnr is not submitted or altered.

Conflict resolution

Only internal/ir/call_targets.go conflicts, in ClosureTargets after selecting a callback expression. Main's 67733635 slice 2 unwraps Effects so an intrinsic adapter retains the identity of its returned closure. The source replaces literal/const-only queries with shared whole-program producer facts, which are also the cycle rule's evidence. Resolve by delegating ClosureTargets to ValueClosureTargets and moving main's result-identity behavior into closureFacts.value's Effects case. It recursively follows Effects.Result; the existing reflective solver still walks Effects.Body and gathers callback declarations and writes. This preserves the main behavior for direct calls and additionally keeps stored adapter values and returned callback bindings bounded. Main's unwrap loop is relocated into this shared rule, not dropped or replaced with Unknown. All other newer main code remains.

Conservative assumption: Effects evaluates its statements, then returns exactly Result, so its result identity belongs in the shared producer solver. Unsupported producers retain Unknown. No nominal type match supplies closure targets. TestClosureTargetsPreserveEffectsProducers covers nested wrappers, captured callback binding, stored values and nonthrowing target identity. Removing the Effects case fails that test with 'effects lost adapter identity', not a compile error.

Counts

Discard the source table in favor of main's entire counts.md, then regenerate once using timeout 600 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -json -timeout=570s -args -update-counts. PASS, 68.711s package. No row was merged by hand. Eight existing rows change and eleven rows are added; none removed. Six existing deltas and all eleven additions exactly match the original source's own change.

The two additional changed rows are newer main fixtures and are this change's effect, not table drift. Build a main-only compiler with an overlay restoring exactly the three production Go files this change edits (ir.go, call_targets.go and cycles.go) to origin/main, and compare its emitted C with this branch's compiler in the same tree. Both compile both fixtures. object-generic-emission.patch shows exactly one array iterator retain/release pair and three per-item pairs removed, accounting for -4 retains/-4 releases. The known callback cannot invalidate these borrows. controls-emission.patch shows exactly the iterator's next closure retain/release pair removed, accounting for -1/-1; the held iterator owns that stable field. Allocations, frees, peak and regions stay unchanged in both. baseline-overlay.json and the generated C files preserve this causal comparison. Their Node/sanitizer oracle checks are included in the package suite.

Counter order below is allocations / frees / retains / releases / peak live / in regions.

| Fixture | Main A/F/R/L/P/G | Re-cut A/F/R/L/P/G | Delta |
|---|---|---|---|
| internal/oracle/testdata/borrow_chain_unknown.a | 8/8/3/8/7/0 | 8/8/2/7/7/0 | 0/0/-1/-1/0/0 |
| internal/oracle/testdata/call_targets_element.a | 45/43/32/60/7/2 | 45/43/30/58/7/2 | 0/0/-2/-2/0/0 |
| internal/oracle/testdata/callback_widening/object-generic.a | 13/13/20/27/10/0 | 13/13/16/23/10/0 | 0/0/-4/-4/0/0 |
| internal/oracle/testdata/generic_functions.a | 35/35/32/55/8/0 | 35/35/29/52/8/0 | 0/0/-3/-3/0/0 |
| internal/oracle/testdata/getters_census_binary_function.a | new | 39/39/37/62/25/0 | new |
| internal/oracle/testdata/getters_census_optional_boolean_function.a | new | 36/36/34/56/24/0 | new |
| internal/oracle/testdata/getters_census_optional_comment_function.a | new | 33/33/37/56/23/0 | new |
| internal/oracle/testdata/getters_census_optional_type_function.a | new | 36/36/40/62/24/0 | new |
| internal/oracle/testdata/getters_census_primary_function.a | new | 33/33/31/50/23/0 | new |
| internal/oracle/testdata/getters_census_unary_function.a | new | 36/36/34/56/24/0 | new |
| internal/oracle/testdata/getters_census_unary_node_function.a | new | 36/36/34/56/24/0 | new |
| internal/oracle/testdata/getters_census_update_comment_function.a | new | 36/36/40/62/24/0 | new |
| internal/oracle/testdata/getters_census_update_node_function.a | new | 39/39/40/65/25/0 | new |
| internal/oracle/testdata/getters_census_update_type_function.a | new | 39/39/43/68/25/0 | new |
| internal/oracle/testdata/instantiation_find.a | 13/13/28/31/7/0 | 13/13/22/25/7/0 | 0/0/-6/-6/0/0 |
| internal/oracle/testdata/miscompile_2a/controls.a | 33/32/34/56/10/1 | 33/32/33/55/10/1 | 0/0/-1/-1/0/0 |
| internal/oracle/testdata/route_targets_sort_values.a | 84/84/92/142/17/0 | 84/84/96/150/17/0 | 0/0/4/8/0/0 |
| internal/oracle/testdata/user_iterators.a | 663/663/455/904/64/0 | 663/663/442/891/64/0 | 0/0/-13/-13/0/0 |
| stage3/namespace-init-sys/cwd.a | new | 6/6/8/12/6/0 | new |

Runtime clearance

The source commit changes no internal/native/runtime/*.c or *.h line. This re-cut also changes none. There are no additional runtime lines to list or clear. No cohere source is copied; the existing submodule is used.

Toolchain and resource observations

GOPROXY=https://proxy.golang.org|direct, GOMAXPROCS=4 and -parallel=4. nproc reports 5; cpu.max=400000 100000 gives four CPUs. Go 1.27.1, Node 24.19.0 and house Linux clang 20.1.8. ADAMIC_GOCACHE_OFF=1 preserves the local-cache workaround already needed by this worker from the previous unit; the external Go cache is not deleted or modified in this unit.

I let the initial setup warm build overlap patch application; it saw conflict markers and failed with syntax errors at call_targets.go:72, :82 and :85 (setup.log). After resolution, bash cloud/setup.sh succeeds, and its printed /workspace/adamic-tools/env.sh is sourced in every Go shell. Successful timing lines: Go 0.022s; Node 0.022s; markdown 0.068s (step 0.008s); submodules 0.067s; clang 0.178s; shared cache off 0.180s; Go build 34.656s; test binaries deferred 34.787s; build cache warm 34.789s; done 34.815s. setup-resolved.log has the complete lines.

The first complete IR invocation failed on no-space errors while copying Go export products for its reader guards; completed leaves were under 20s. It had no semantic/SortedMap failure. To preserve caches and recover capacity, temporarily move the oldest completed export product to /tmp, compare only completed immutable export archives by full SHA256 plus owner/mode, hard-link identical copies, then restore the old product at exactly its original path. 3,173 archives verified; 2,762 duplicates linked; 14,552,389,066 duplicate bytes avoided. Every product's bytes and paths are preserved and the external Go cache is untouched. cache-deduplication.json lists each verified link. Free disk grew to about 14 GB. The complete IR package was rerun and passed after this environment fix; no test or assertion was weakened.

A-check

All three changed .a files are scanned in place using this branch's actual compiler and unchanged pinned Gate.aCheck. direct.a, joined.a and map.a under internal/oracle/testdata/cycle_closure_targets_refused/ are all refused with their existing expected reasons. Each complete file is byte-identical to b1d42ea391, including its header. Counts: zero checked, zero NotYet, three refused, zero checker errors. There are no header or position-expectation edits in this re-cut. a-check-status.json retains full diagnostics and body SHA256s. The predicate is pinned from fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py; only the harness redirects its executable output path into evidence. Both removed-header and wrong-reason mutants on direct.a fail the unchanged predicate with captured actual branch diagnostics, as the README specifies. The real file is restored and all three pass again. Header proof is predicate replay, not a claim of recompilation per comment mutation.

Mutants and catchers

- omit-boundary-inputs: TestClosureTargetsKnownCallerDoesNotHideUnknownCaller, 'known caller hid unbounded caller'.
- unknown-is-empty: TestClosureTargetsBoundOnlyProvenValues, 'bounded without proof'.
- omit-captures: TestClosureCycleDirect, 'want cycle refusal'.
- drop-return-target: TestClosureTargetsFollowParametersReturnsAndMapValues, 'want both returned targets'.
- omit-map-captures: TestClosureCycleMap, 'want cycle refusal'.
- type-matched-captures: TestNamespaceAmbientCwdInitialization and TestClosureCycleAmbientCwd both fail with the old adamic/cycle-capable refusal.
- drop-effects-result: TestClosureTargetsPreserveEffectsProducers, 'effects lost adapter identity'.
- removed-header and wrong-reason: unchanged Gate.aCheck rejects the missing expectation and mismatched reason, respectively.

All implementation mutants compile and fail their intended assertion, never a build error or timeout. Sources are restored in finally. Mutation sources are .patch or .go.txt, not compilable Go under review. The Effects test's fixture was subsequently made explicit about the captured callback Environment; effects-final.log rechecks that final fixture.

WASI prerequisite follow-up

The initial complete native suite finishes with only TestSplitAttributedBindingsAgreeWithNode/wasm32-wasi failing: native: wasm32-wasi requires WASI_SYSROOT. It is a mandatory main test even without optional ADAMIC_TEST_WASI. Install the repository-supported SDK with timeout 240 bash cloud/setup.sh --wasi-sdk, then source env.sh again. Setup passes in 20.410s: Go 0.019s; Node 0.026s; submodules 0.074s; markdown 0.075s (step 0.007s); clang 0.150s; WASI SDK 3.414s; shared cache off 3.415s; Go build 19.894s; deferred tests 20.235s; warm cache 20.241s; done 20.410s. Native remains house clang 20.1.8; WASI uses SDK clang 20.1.8-wasi-sdk. The full native suite is rerun with this prerequisite, without skipping or editing its test.

The final explicit captured-environment Effects fixture also catches the same removed-case mutant under a Go overlay (drop-effects-final-result.json), leaving production sources untouched while other suites run.

Verification commands and completed evidence

Every test command uses GOMAXPROCS=4 and -parallel=4, with output written directly to the named log. Complete direct packages are run, not a selected-test substitute and not the full repository gate: internal/ir, internal/lower, internal/native, internal/javascript, internal/flow, internal/fresh and internal/oracle. final-package-results.json, remaining-package-results.json and native-wasi-ready.log record the actual invocations and elapsed times. The initial native failure is retained separately from the final SDK-ready rerun. Optional ADAMIC_TEST_WASI remains unset; the mandatory native/WASI attributed-binding witness executes. The first IR run's ENOSPC is retained separately from its passing rerun.

Focused command: timeout 120 go test ./internal/ir ./internal/lower ./internal/oracle -run '^(TestClosureTargets.*|TestNamespaceAmbient.*|TestClosureCycle.*|TestGettersCensus.*)$' -count=1 -parallel=4 -json -timeout=90s. PASS. Slowest focused leaf 29.39s. Admitted witnesses are held to Node with native ASan/UBSan/leak checks by their oracle helpers; refused witnesses are held to Node and then to exact compiler refusals. Full oracle additionally checks the two newer main fixtures whose counts move.

Reader guard: timeout 120 go test ./internal/ir -run '^TestCallTargetReaders' -count=1 -parallel=4 -json -timeout=90s. PASS; prefix includes main's split Native/Lower/Fresh/Flow export and owner guards. Vet: timeout 180 go vet ./internal/ir ./internal/lower ./internal/oracle ./internal/native ./internal/javascript ./internal/flow ./internal/fresh. PASS.

Additional guard command: timeout 120 go test ./internal/ir ./internal/lower ./internal/oracle ./cmd/adamic-gate -run '^(TestEveryMapAProgramHoldsIsSorted|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed|TestCountsAreRecordedTable|TestFamiliesOwnEveryCaseOnce|TestFamilyWrappersAreCurrent|TestEveryTestIsParallelOrSaysWhy)$' -count=1 -parallel=4 -json -timeout=90s. PASS. The producer cache uses SortedMap for held maps; no new compiler-consuming package or runtime feature definition is introduced; the runtime-set witness guard passes. New tests call t.Parallel first. Existing native-agreement registry and family ownership rules pass without unrelated test edits. Products/tests gain no build directory or machine path.

Gofmt -l over every changed Go file prints nothing. git diff --check origin/main -- docs internal passes. Runtime source/re-cut diffs are both empty in runtime-lines.json. source-file-comparison.json verifies all source-commit files except the regenerated table are byte-identical to the source unless newer main/resolution affects ir.go, call_targets.go, call_targets_test.go or cycles.go; the complete original review evidence is carried through.

Timing limit: the 30-minute first-green-push target was missed while completing the mandatory full suites after repairing disk capacity and the missing WASI SDK. No unverified state was pushed to claim that target. All individual completed test leaves remain within their 60-second limit; full package aggregates are explicitly bounded and watched throughout.

Final complete package results

| Package | Package seconds | Slowest leaf seconds | Result |
|---|---:|---:|---|
| internal/ir | 22.656 | 15.82 | PASS |
| internal/lower | 50.264 | 8.62 | PASS |
| internal/native | 362.627 | 28.9 | PASS |
| internal/javascript | 0.236 | 0.09 | PASS |
| internal/flow | 103.425 | 22.73 | PASS |
| internal/fresh | 29.322 | 6.83 | PASS |
| internal/oracle | 605.855 | 49.34 | PASS |

Every completed final leaf is under 60 seconds. all-final-test-seconds.json records all terminal test events. New/changed declared test timings from complete final suites (the explicit capture-layout refinement is also checked separately in effects-final.log):

| Test | Seconds |
|---|---:|
| TestClosureTargetsFollowParametersReturnsAndMapValues | 0 |
| TestClosureTargetsKnownCallerDoesNotHideUnknownCaller | 0 |
| TestClosureTargetsPreserveEffectsProducers | 0 |
| TestNamespaceAmbientCwdInitialization | 1.12 |
| TestClosureCycleDirect | 0.19 |
| TestClosureCycleJoined | 0.19 |
| TestClosureCycleMap | 0.19 |
| TestClosureCycleAmbientCwd | 1.11 |

Delivery is one commit with parent 67733635b0252eb9d3bb3645c72262c7b18361e1, retaining the complete original message plus the requested re-cut line, conservative assumption and Ahra co-author trailer. Only compiler/loom-r-cycle-targets-ac-m is pushed, without force. The final response supplies the full delivery SHA and remote verification. No ahra CLI, PR or Loom submission is run.
