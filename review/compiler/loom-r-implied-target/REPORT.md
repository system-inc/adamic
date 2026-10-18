Built the node-keyed implied-target cache and exact plain-scalar cycle-shape sharing for large array literals (#d0bzeyh).
Delivery: one commit on pinned main dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source 3514d4ecfe25eea222744e25ca5ccce9fd9ee6bc; the final delivery message gives the pushed SHA.
Validation: all 52 reader packages pass focused tests or compile-only checks; eight new leaves pass; reader guard, vet, five oracle fixtures and the 2,048-element backend control pass.
Mutants: all nine source-report mutants fail their intended query-count or scalar-shape assertions; none is credited for a warning, timeout or build failure.
Not covered: full package suites, full gate, full flow corpus, mutable parallelMap admission or new performance profiling; the auxiliary cold counts-family test timed out and was replaced by direct production measurements, not rerun warm.

The branch was created from exactly dbdf977f5b8b0979d414ebe6f20e7aadd312e309, as requested, even if origin/main later moves. The source commit was fetched explicitly and applied with git cherry-pick --no-commit 3514d4ecfe25eea222744e25ca5ccce9fd9ee6bc. This applies its own parent-to-commit diff, never main..source. It applied cleanly: no conflicted hunk and no resolution was needed. scope-proof.json verifies that every added/deleted line in all five lowering files exactly matches the source's own change, preserving newer main content.

The three production files are internal/lower/invariance.go, internal/lower/cycles.go and internal/lower/lower.go. The two added test files are implied_target_test.go and cycle_scalar_shapes_test.go in the same package. No other package's test was edited. The parent-target cache includes nil answers and lives only within one lowering/checker. Scalar shape sharing is restricted to the source's plain, fresh, directly named scalar properties, preserving field order, exact type identities, field flags and readonly status. No map was added to an IR program. No compiler-consuming Go package or runtime feature set was added.

Runtime C lines changed: none. Runtime lines different from the source commit's own runtime delta: none. Source and candidate both have empty internal/native/runtime diffs. No runtime clearance is required for an added line because there is no added runtime line.

No registered oracle fixture was added or modified. The refused original mutable parallelMap example remains under review, outside the flow corpus; no refused top-level oracle fixture was introduced. The existing TestFlowCorpusUnitsCoverEveryProgram coverage audit passed. Full flow-corpus execution was not required by the new-fixture condition and was not run.

Conservative assumption: preserve main's admission. The exact original objects.a still refuses mutable parallelMap input. A production baseline CLI built from the pinned main's three original lowering files through .go.txt overlays and the candidate CLI produce byte-identical refusal stderr, exit 1 and empty stdout. original-admission.json preserves both observations. No cohere code was copied. The accepted sequential items.map control is the source report's unchanged 2,048-object literal and worker. Raw source Node, generated JavaScript, release native and ASan/UBSan with detect_leaks=1 all print exactly:

2048 moved objects, ordered digest 661084526

All four observations exit 0 with empty stderr. backend-results.json records commands and seconds. The release compilation takes 34.786 s and sanitized compilation 21.114 s; these are CLI operations, not test-leaf timings. No new wall-time speedup or universal linear-complexity claim is inferred from this re-cut. The deterministic regressions assert one parent query per array/conditional/logical parent, cached nil tuple results, and one shared scalar shape after every one of 3,072 registrations before the full cycle analysis.

Validation used /workspace/adamic-tools/env.sh, GOMAXPROCS=4, and the supplied -trimpath -buildvcs=false flags with -p=4 appended. Every potentially long command had a hard limit; every test's output went directly to a log file. Exact commands, selectors, wall seconds and test events are in focused-results.json, mutants.json and logs/. Every completed leaf's seconds are in test-seconds.json and TEST-SECONDS.md.

The reverse dependency closure was recomputed from current go list Imports/TestImports/XTestImports: 52 packages read internal/lower, directly or transitively. packages.jsonstream is the raw snapshot, reader-packages.json lists them, and focused-selection.json lists the exact related tests. run-focused.py runs up to four selected top-level tests per bounded invocation, or -run '^$' for reader packages with no relevant test. This resolves the instruction to run each reader package conservatively while honoring the explicit prohibition on whole-package suites. The pattern comes from the source's report and includes implied-target, scalar-shape, widening/invariance, cycle/fresh-write, enum, closure/refusal and checked-view controls. All 52 packages complete successfully. No complete package suite or full gate was run.

Additional exact commands:

- go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -json -timeout 90s: pass, package 25.750 s. No reader allowlist changes.
- go vet ./internal/lower, bounded at 120 s: exit 0 with empty output.
- go test ./internal/oracle -run '^TestNativeAgreesWithNode_[0-9]+$/^internal$/^oracle$/^testdata$/^(invariance_readonly|optional_widening_(class|declared|fresh)|node_options_widening)[.]a$' -count=1 -json -timeout 90s with ADAMIC_GATE_UNCACHED=1: five fixture leaves pass; package 26.265 s. The full fixture paths were selected and observed, not an empty match.
- go build ./cmd/adamic into a task-specific scratch binary, then compare-control.py: the large sequential source control agrees with Node across both backends and sanitizer/leak checks.
- gofmt -l on the five changed Go files and git diff --check: clean.

Counts rows changed: none. There is no new fixture and no observed count movement, so no -update-counts regeneration was needed or run. The auxiliary read-only TestCountsAreRecorded family selection hit its 90 s timeout before running a selected count-row subtest. Main's countShardsAhead prepares product measurements before t.Run applies the path filter; selecting every numbered wrapper therefore prepared all cold measurement products rather than only the five requested rows. The timeout and stack are retained in logs/counts-neighbors.jsonl. It was not rerun with a warmed cache, and this is not claimed as a passing counts-family check.

Instead, direct-counts.py builds the same five fixtures through the production CLI with --count, runs each with the documented 8 MiB stack limit on Linux and clang 20.1.8, and compares all six measured numbers to the pinned main table. Every row matches exactly, as recorded below and in direct-counts.json. This is direct evidence of unchanged rows; it does not claim full-table measurement. counts.md remains byte-identical to the pinned base, without hand-merging or rewriting unrelated rows.

| Fixture | Recorded and measured A/F/R/L/P/G | Seconds including build |
|---|---|---:|
| internal/oracle/testdata/invariance_readonly.a | 30/30/38/51/16/0 | 0.944 |
| internal/oracle/testdata/optional_widening_class.a | 3/3/2/7/3/0 | 0.925 |
| internal/oracle/testdata/optional_widening_declared.a | 6/6/2/12/4/0 | 1.094 |
| internal/oracle/testdata/optional_widening_fresh.a | 3/3/2/8/3/0 | 1.045 |
| internal/oracle/testdata/node_options_widening.a | 5/5/1/7/5/0 | 1.186 |

New top-level test leaves, setup included:

| Test | Seconds |
|---|---:|
| TestCycleScalarShapeKeysPreserveDistinctFields | 0.15 |
| TestCycleScalarLiteralShapesOnce | 0.96 |
| TestImpliedTargetLargeArrayQueriesOnce | 0.83 |
| TestImpliedTargetAndQueriesOnce | 2.97 |
| TestImpliedTargetConditionalQueriesOnce | 8.49 |
| TestImpliedTargetTupleCachesNil | 0.09 |
| TestImpliedTargetOrQueriesOnce | 0.57 |
| TestImpliedTargetNullishQueriesOnce | 1.47 |

All nine independent source-report mutations were rerun using Go overlays, so the working source was never mutated. Mutant sources end in .go.txt.

| Mutant | Expected catcher | Assertion text |
|---|---|---|
| array-per-element-query | ^TestImpliedTargetLargeArrayQueriesOnce$ | more than once |
| conditional-per-branch-query | ^TestImpliedTargetConditionalQueriesOnce$ | more than once |
| logical-per-operand-query | ^TestImpliedTarget(Or / And / Nullish)QueriesOnce$ | more than once |
| nil-cache-miss | ^TestImpliedTargetTupleCachesNil$ | more than once |
| shape-ignores-readonly | ^TestCycleScalarShapeKeysPreserveDistinctFields$ | shared a shape key |
| shape-per-literal | ^TestCycleScalarLiteralShapesOnce$ | scalar literal shapes = 2 |
| shape-ignores-type | ^TestCycleScalarShapeKeysPreserveDistinctFields$ | shared a shape key |
| shape-ignores-name | ^TestCycleScalarShapeKeysPreserveDistinctFields$ | shared a shape key |
| shape-includes-reference | ^TestCycleScalarShapeKeysPreserveDistinctFields$ | want 7/4 |

All nine overlay runs exit 1 through these assertions; the logical-operand mutant fails the Or, And and Nullish leaves separately. The per-query observer fails before a second expensive checker query, and the per-literal observer fails at the second unshared shape, so a restored quadratic path is caught without spending its unbounded work or relying on elapsed time.

Setup was run after exporting GOPROXY=https://proxy.golang.org|direct. ADAMIC_GOCACHE_OFF=1 selected the local cache. Timings: Go ready 0.073 s; Node ready 0.081 s; markdown ready 0.219 s; clang ready 0.451 s; submodules ready 1.426 s; shared cache off 1.428 s; Go build ready 50.643 s; test binaries deferred 50.837 s; build cache warm 50.839 s; done 50.867 s. nproc=5 and cgroup cpu.max=400000 100000 (four CPUs). Go 1.27.1, clang 20.1.8, Node 24.19.0. setup.log retains the complete output. Setup succeeded without a workaround.

No ahra command, PR, Loom submission or main push was made. The user submits this SHA for the lowering wave. Historical review/compiler/loom-implied-target evidence is preserved as part of the source's own diff; this report and directory contain the new pinned-base evidence.
