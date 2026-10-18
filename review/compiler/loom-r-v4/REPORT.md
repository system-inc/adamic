Built: views V4 carrying V3 for #1p9ste9 #6wfvfhm on exactly dbdf977f5b, on compiler/loom-r-v4.
Commits: one commit above dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source 332f305e6489cd90ebcc2a073f24f8289cb881ea; pushed SHA is in the final response.
Commands and outputs: 299 focused tests/products completed with seven documented skips; 858 flow leaves passed; 53 changed/reader packages compiled and 48 focused reader checks completed; counts, guards and vet passed.
Mutants: all 18 mutants from the source report were caught, plus a main library-spread preservation mutant; all sources restored.
Not covered: whole packages/full gate, ahra loom check, pending TSan/region controls, and two documented reader skips. No PR or Loom submission.

The exact-base instruction takes precedence over the generic current-main instruction. Started with git fetch origin, git checkout -b compiler/loom-r-v4 dbdf977f5b, and git cherry-pick --no-commit 332f305e6489cd90ebcc2a073f24f8289cb881ea. This applies the source's own parent-to-commit change, never a main-to-source diff. No newer tip is substituted. No cohere implementation was copied; setup uses the pinned submodule and the isolated mutant worktree references it through a symlink.

There was one conflicted hunk, at the start of type lowering struct in internal/lower/lower.go. Main 7d75348e adds librarySpreads []librarySpread for closed-world refusal of library-object spreads whose internal storage is not enumerable data. V4 adds viewCallablePayloadTypes and viewCallableProducerPayloads for its callable certificates. Keep all three fields. Main's checkLibrarySpreads invocation, candidate collection and proven plain-source logic remain; V4's payload maps remain. conflicts.json holds the hunk detail. All seven TestLibrarySpread* controls pass. Omitting the main invocation makes TestLibrarySpreadRegExpUnion fail with “want RegExp refusal, got <nil>”, proving the newer meaning is still guarded.

The auto-merges retain main's async, parallelMap, weak-reference and newer archive infrastructure. Additional compatibility tests passed:

| Main compatibility test | Seconds | Observation |
|---|---:|---|
| TestParallelMapCompilerABI | 0.036 | Original-self callback and result reference flag ABI intact |
| TestAsyncCompilerSuspensionsMatchNode | 18.619 | Source Node, backend Node, ASan/UBSan native and leak check agree |
| TestParallelCompilerSharedReferencesMatchNode | 2.082 | Source Node, backend Node, ASan/UBSan native and leak check agree |

One test-only compatibility correction was needed beyond the conflict. The source's private cache fixtures link only closure.c, heap.c, weak.c and count.c, with aborting support functions for unrelated heap kinds. Main's heap.c now also references adamic_async_free_children. Ten fixtures initially failed to link because support.c lacked that symbol; none exceeded its budget. Add this matching aborting test stub in internal/native/testdata/view-adapters/support.c:

    void adamic_async_free_children(void *value, void (*release)(void *)) { (void)value; (void)release; abort(); }

These fixtures allocate closures, cells and weak handles, not async objects. Reaching the unrelated destructor aborts rather than silently replacing ownership behavior. All ten fixtures pass after this correction, including their built-in allocation, identity, ordering, canonical, resurrection, leak-root and sanitizer mutants. Initial failures remain in initial-cache-link-failures.json and their original logs; corrected results are in cache-fixture-fixed.jsonl and *-fixed.log. No production runtime C or header was edited for this correction.

Runtime clearance comparison: zero changed lines versus the source commit's own runtime +/- lines. All 452 lines match, per file and in order, using source~1 to source versus the fixed base to the recut tree. Main's newer runtime work remains outside that delta. runtime-lines.json names every file and records an empty differences mapping. The support.c addition above is test data outside internal/native/runtime.

Counts were never hand-merged. Restored the untouched table from the exact base, waited for compilation and the flow corpus to pass, then ran one complete regeneration on Linux with clang 20.1.8, following the command in main's counts.md header:

    source /workspace/adamic-tools/env.sh
    timeout 480 go test ./internal/oracle -run '^TestCountsAreRecorded' -count=1 -timeout 360s -parallel 4 -json -args -update-counts

It passed in 128.363 s, maximum recorded count leaf 4.620 s. The test-only support correction does not alter emitted programs or runtime sources and needed no second regeneration. Main has 1,073 rows; this change has 1,096. Exactly 30 rows change: 26 additions, three compile-time refusal removals and namespace_method_receiver.a losing two retain/release pairs. counts-attribution.json verifies that every before/after row delta is exactly the source commit's own delta, with no differences and no unrelated main rows changed. These are observed measurements; attributing the namespace pair removal to declaration-only this filtering follows the lowering code. A/F/R/L/P/G below mean allocations, frees, retains, releases, peak live and in regions.

| Fixture | Base A/F/R/L/P/G | Recut A/F/R/L/P/G |
|---|---|---|
| internal/oracle/testdata/library_array_holes_length.a | absent | 17/17/3/19/6/0 |
| internal/oracle/testdata/library_array_holes_range.a | absent | 24/24/12/27/4/0 |
| internal/oracle/testdata/library_array_holes_scanner_probe.a | absent | 4/4/0/3/4/0 |
| internal/oracle/testdata/namespace_method_receiver.a | 7/7/9/16/5/0 | 7/7/7/14/5/0 |
| stage3/interface-downcasts/lane2/array-alias.a | absent | 3/1/9/8/3/0 |
| stage3/interface-downcasts/lane2/array-boolean.a | absent | 2/0/4/3/2/0 |
| stage3/interface-downcasts/lane2/array-bounds.a | absent | 5/5/5/10/5/0 |
| stage3/interface-downcasts/lane2/array-element.a | absent | 4/4/4/7/4/0 |
| stage3/interface-downcasts/lane2/array-field.a | absent | 4/4/3/6/4/0 |
| stage3/interface-downcasts/lane2/array-holes-view.a | absent | 7/7/9/15/5/0 |
| stage3/interface-downcasts/lane2/array-iteration-bad.a | absent | 2/0/4/2/2/0 |
| stage3/interface-downcasts/lane2/array-iteration.a | absent | 6/6/3/8/4/0 |
| stage3/interface-downcasts/lane2/array-lazy.a | absent | 4/4/5/7/4/0 |
| stage3/interface-downcasts/lane2/array-length-kind.a | absent | 2/0/2/1/2/0 |
| stage3/interface-downcasts/lane2/array-map-bad.a | absent | 4/0/4/1/4/0 |
| stage3/interface-downcasts/lane2/array-map.a | absent | 7/7/5/12/6/0 |
| stage3/interface-downcasts/lane2/array-missing.a | absent | 1/0/2/1/1/0 |
| stage3/interface-downcasts/lane2/array-pop.a | absent | 6/6/5/11/6/0 |
| stage3/interface-downcasts/lane2/array-second.a | absent | 2/0/6/3/2/0 |
| stage3/interface-downcasts/lane2/array-string-literal.a | absent | 2/0/4/2/2/0 |
| stage3/interface-downcasts/lane2/array-undefined.a | absent | 2/0/4/1/2/0 |
| stage3/interface-downcasts/lane2/generic-array-element.a | absent | 2/1/1/2/2/0 |
| stage3/interface-downcasts/lane2/non-array.a | absent | 1/0/2/1/1/0 |
| stage3/interface-downcasts/lane2/nullable-array-element.a | absent | 1/0/4/2/1/0 |
| stage3/interface-downcasts/lane2/nullable-array-missing.a | absent | 0/0/1/1/0/0 |
| stage3/interface-downcasts/lane2/nullable-array-string.a | absent | 2/0/4/2/2/0 |
| stage3/interface-downcasts/lane2/object-element.a | absent | 3/3/9/10/3/0 |
| stage3/interface-downcasts/untagged/fixtures/callable-union-nested.a | 3/0/6/4/3/0 | absent |
| stage3/interface-downcasts/untagged/fixtures/callable-union-wrong.a | 2/0/5/4/2/0 | absent |
| stage3/interface-downcasts/v2/callable-producer-wrong.a | 2/0/5/4/2/0 | absent |

Flow corpus command (not the whole package):

    timeout 720 go test ./internal/flow -run '^TestFlow' -count=1 -parallel 4 -timeout 660s -json

Output went to flow-corpus.jsonl and flow-corpus.stderr. It passed in 116.900 s, with 858 leaves and a maximum leaf of 49.640 s. All discovered top-level programs lower; no new fixture is refused there. The three added top-level fixtures run through all four flow analyses in TestFlowCorpusRemainder: library_array_holes_length.a 0.250 s, library_array_holes_range.a 0.300 s, library_array_holes_scanner_probe.a 0.390 s. The six carried V4 compile-time refusal controls already live under views_v4_refused/. No new refusal was left at top level. flow-remainder-results.json records the discovery checks.

Focused tests ran through run-focused.py with the source report's inventory, current portability checks, seven main library-spread controls and the required guards. No added source test declaration is missing (source-test-inventory-audit.json). Each exact top-level leaf uses -test.timeout=90s, -test.parallel=1, an outer 95 s bound and its own log; two leaf processes run on this four-CPU-quota instance. Its watched 840-second driver ended before its last 52 tests; resume-focused.py ran only unfinished leaves and never reran a completed or over-budget leaf. The initial ten linker failures were corrected and rerun as described above. No leaf exceeded 60 seconds. This continuation is a driver bound, not an over-budget test result.

The final inventory is 299 tests/products, 7 documented skips, all other results passing; maximum command wall time 41.207 s. Source Node controls, backend Node, release native and ASan/UBSan/LSan run as defined by each positive/refusal fixture. test-seconds.csv lists every final focused leaf, flow leaf, counts leaf, guard and reader result, with command wall versus Go leaf timing distinguished. final-focused-results.json applies the corrected cache results while raw logs are retained.

Documented focused skips:

- internal/native TestViewAdapterPoolTSanPending
- internal/native TestViewAdapterProgramUnderlyingPending
- internal/native TestViewAdapterRegionAdoptionPending
- internal/native TestViewAdapterRuntimeStaticsPending
- internal/oracle TestV4EscapeAdapterBlameRuntimeCallback
- internal/oracle TestV4Lane5FactoryReadonlyTargetControl
- internal/oracle TestV4Lane5FactoryWiderWritePending

All changed Go packages and all direct production/test readers were identified with go list -json ./... and the import edges, not a hand-maintained old package list. All 53 compile with go test -c, each build under an outer 180 s bound. run-readers.py runs one bounded exact reader control in each of 48 test packages; these cover compiler profiles, explicit lowering gaps/refusals, ABI/input keys, bridge cache verification, adapters and fixture product tracking. These are focused samples, not exhaustive package tests. The other five packages have no tests. All completed reader checks passed, maximum command wall 38.941 s. Full package/selector inventories are consumers.json and reader-selection.json; results and each log are reader-results.jsonl and *-reader.log. The prior unit's over-budget cold TestProduct_ApiDriver was not rerun; this unit ran bridge's cache-integrity control and compiled its reader package, without claiming the cold API product was measured anew.

Reader skips, not passes claimed:

- internal/lower/cycleledger TestOriginalCycleLedger (see its reader log)
- stage1/cohere/estree TestCookedSurrogateLibraryGap (see its reader log)

Final guards were run on the restored final source using:

    timeout 180 go test ./internal/ir ./internal/lower ./internal/oracle ./cmd/adamic-gate -run '^(TestEveryMapAProgramHoldsIsSorted|TestCallTargetReaders|TestNativeAgreementRowsNameRegisteredPrograms|TestEveryTestIsParallelOrSaysWhy|TestRuntimeSetsAreEachUsed)$' -count=1 -parallel 4 -timeout 90s -json

| Guard | Seconds | How satisfied |
|---|---:|---|
| TestNativeAgreementRowsNameRegisteredPrograms | 0.050 | No unregistered lowering agreement literal is introduced. |
| TestEveryTestIsParallelOrSaysWhy | 0.760 | No new serial exception or baseline weakening. |
| TestEveryMapAProgramHoldsIsSorted | 0.000 | Source Program-held maps stay SortedMap, including OptionalViewFields. |
| TestCallTargetReaders | 5.030 | Shared call-target readers remain intact on the combined compiler. |
| TestRuntimeSetsAreEachUsed | 7.580 | Source-profile witness still enables exactly receiver plus regexp callback through main runtimeproducts. |

No new compiler-consuming package is introduced; runtimeproducts was already declared in cloud/fast-gate/compiler-dependencies.json. Main's placeless product recipes and keys are preserved; no new product/test identity names a build directory or machine path. The only newly authored fixture-support code is C, and no new .ts program was authored. Inherited TypeScript assertion probe inputs are carried unchanged as source-profile controls.

Final vet passed with no output:

    timeout 180 go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/load ./internal/lower ./internal/native ./internal/native/runtimeproducts ./internal/oracle ./stage1/cohere/json

final-vet.log records the result; gofmt.log and source-whitespace.log are empty. No full package test run, full gate, ahra CLI, PR or Loom submission was run.

All 18 source-report mutants ran freshly on an isolated exact-base worktree. A nineteenth proves main's library-spread refusal survived the conflict. Each Go test uses -run '^<catcher>$' -count=1 -timeout 90s with an outer 120 s bound and ADAMIC_GATE_UNCACHED=1. Mutant sources are only .patch files in this evidence directory. Every mutant exits 1 from the expected assertion or sanitizer, not a build warning/link failure. Command wall includes Go rebuilding; individual Go leaf times are in the logs. mutants-restored.json verifies every mutated file matches the final recut byte for byte. Sanitizer byte dumps are decoded in sanitizer-mutants.json.

| Mutant | Catcher | Command wall seconds | Observed failure |
|---|---|---:|---|
| escape-Argument | TestV4EscapeArgumentRefusal | 39.899 | required refusal/path/fix absent |
| escape-Result | TestV4EscapeResultRefusal | 11.648 | required refusal/path/fix absent |
| escape-Passed | TestV4EscapePassedRefusal | 11.848 | required refusal/path/fix absent |
| escape-Returned | TestV4EscapeReturnedRefusal | 14.269 | required refusal/path/fix absent |
| escape-MixedProducer | TestV4EscapeMixedProducerRefusal | 19.193 | required refusal/path/fix absent |
| unsupported-domain | TestV4DirectRecursiveRefusal | 52.855 | required refusal/path/fix absent |
| unsupported-producer | TestV4DirectProducerRefusal | 41.850 | required refusal/path/fix absent |
| surface-borrow | TestV4EscapeAdapterName | 35.090 | ASan heap-use-after-free |
| bind-name-borrow | TestV4EscapeAdapterBindNameOwnership | 32.455 | ASan heap-use-after-free |
| identity-root | TestV4EscapeAdapterIdentity | 58.603 | Node stdout comparison |
| metadata-inline | TestPortElementMetadataEmission | 14.091 | 2,885 direct stores exceed zero bound |
| metadata-omitted | TestPortElementMetadataEmission | 12.986 | zero runtime certificates instead of 2,885 |
| native-order | TestV4DirectCalleeOrder | 16.320 | argument effects missing after premature callable check |
| javascript-order | TestV4DirectCalleeOrder | 16.539 | argument effects missing after premature callable check |
| construct-witness | TestV4EscapeAdapterWithoutConstructSignatureRefusal | 4.064 | required refusal/path/fix absent |
| boxed-witness | TestV4Lane5BoxedAdamicRefusal | 9.777 | unproven-read refusal replaced by another refusal |
| factory-witness | TestV4Lane5FactoryWiderWriteRefusal | 10.603 | required refusal/path/fix absent |
| runtime-witness-mismatch | TestRuntimeSetsAreEachUsed | 17.559 | dedication enables [] instead of receiver plus regexp callback |
| main-library-spread-omitted | TestLibrarySpreadRegExpUnion | 16.389 | RegExp library-spread refusal absent: got <nil> |

Setup: GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh; source the printed /workspace/adamic-tools/env.sh in build/test shells. setup.log contains all timing lines: Go ready 0.055 s; submodules ready 0.067 s; Node ready 0.086 s; markdown dependencies ready 0.190 s; clang ready 0.374 s; shared cache ready 1.156 s; go build ready 61.975 s; test binaries deferred 62.041 s; build cache warm 62.042 s; done 62.075 s. House versions: clang 20.1.8, Go 1.27.1, Node 24.19.0. nproc=5; cpu.max=400000 100000 means four CPU shares. Setup passed without a workaround. Every long command carried a hard bound, and watched batches were split before 15 minutes.

The commit keeps exactly the requested base and preserves the source runtime delta. Conservative choices are to retain both independent lowering-state meanings and to fail loudly on unrelated async destruction in the minimal private fixture. Main's newer async and parallelMap behavior was checked separately rather than removed or narrowed. Push only compiler/loom-r-v4, never main or another worker's branch, without force.
