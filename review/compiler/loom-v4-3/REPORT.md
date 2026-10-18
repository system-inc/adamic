Built: re-cut checked callable views and carried array lowering on current main; authorized for push.
Commits: one compiler/loom-v4-3 commit above d79302e50c022c997c00f7c68ddb7475f9ee65fc; source 516829207651189faddb3cf97e8b9ce74d3d9feb; pushed commit SHA is in the final response.
Commands and outputs: 289 focused tests/products completed, 7 documented skips, all under 60 s; 52 reader/changed packages compiled; counts, five guards and vet passed.
Mutants: all 17 recorded source mutants and one exact-runtime-witness mismatch mutant were caught; every source was restored.
Not covered: remaining consumer tests after the local bridge/tsgo cold-build failure; full packages/full gate; no PR or Loom submission.

The requested branch has one commit on origin/main's fetched tip, d79302e50c022c997c00f7c68ddb7475f9ee65fc. The source is origin/compiler/loom-v4-2 at 516829207651189faddb3cf97e8b9ce74d3d9feb. Fetch initially tracked only main, so the source ref was fetched explicitly. No other branch or main was changed. The initial local stop was superseded by the user's explicit push instruction. A fresh fetch before pushing found origin/main unchanged at the reported base, so no rebase was needed. The first push was delayed beyond the original 30-minute target while validating the cold box.

The squash had two content conflicts, resolved once while retaining both meanings:

- internal/oracle/checked_views_v2_counts_test.go: main's two-phase countCase.prepare returns emitted C plus a measurement closure. Preserve that API and V4's three escape-refusal exclusions. Adapt the incoming array count registration and baseline count assertions to the same API.
- internal/oracle/runtimes_test.go: main moved runtime archive recipes and declarations into internal/native/runtimeproducts (a2261547, following its shared supplier extraction). Preserve Supply/Build, placeless Inputs and main's additional receiver-plus-node-host set. Declare V4's receiver-plus-regexp set and sanitized-counted variant there. Keep the source's Node-checked source-profile witness callback in the oracle's exact-feature guard; it lowers the existing v4ReceiverRegExpSource program and the guard compares native.Features exactly.

The initial admitted .a witness attempt refused the original escaping read. A positive admitted receiver/regexp control passed, but the final code retains the source-profile witness to avoid introducing another registered oracle fixture. The temporary .a control was removed. The final source-profile guard passed, and substituting dedication/dedication.a as the witness failed precisely because it enables no features. Logs retain these observations separately.

Main's current counts architecture supersedes the digest-column premise: 39b99b53 introduced that column, then 1f07ce2d moved digests into measurement product keys and restored counts-only rows. Preserve current main's counts-only table and two-phase measurement products. Starting from main's table, one completed Linux regeneration ran:

    source /workspace/adamic-tools/env.sh
    timeout 480 go test ./internal/oracle -run '^TestCountsAreRecorded' -count=1 -timeout 300s -parallel 4 -json -args -update-counts

It passed in 90.118 s; 1,048 count leaves passed, maximum 18.460 s. The initial cold-build attempt ended at its outer limit before any tests ran; it did not regenerate the table. counts-regenerated.jsonl and counts-regenerated.stderr hold the completed run. No Mac table was used.

Exactly 30 rows change, with the entire before/after values in counts-diff.json. counts-attribution.json verifies that the row delta is exactly identical to source versus its base 6935e5933fbb74b28825a2fe221af3e0a65c1066:

- 26 additions: three library_array_holes_* fixtures and 23 lane2 checked-array fixtures.
- Three removals: callable-union-nested.a, callable-union-wrong.a, and v2/callable-producer-wrong.a, now compile-time escape refusals with Node controls in their refusal tests.
- namespace_method_receiver.a: 7/7/9/16/5/0 becomes 7/7/7/14/5/0. Observed counts are measured; the attribution to excluding declaration-only this from argument slots follows internal/lower/functions.go's new parameter filtering.

All other current-main rows are retained. No counts row changed for a reader's work.

The internal/native/runtime added/deleted lines are exactly identical to the source delta: all 452 normalized +/- lines match, with no differences. runtime-delta.json records the refs and comparison. No runtime C or header was changed to satisfy the newer infrastructure.

Every required guard was run on the final source, with JSON evidence in final-guards.jsonl:

| Guard | Seconds | How satisfied |
|---|---:|---|
| TestEveryMapAProgramHoldsIsSorted | 0.000 | The incoming OptionalViewFields and other Program-held maps remain SortedMap. |
| TestNativeAgreementRowsNameRegisteredPrograms | 0.050 | Incoming lowering tests introduce no unregistered agreement program. |
| TestEveryTestIsParallelOrSaysWhy | 0.980 | Incoming top-level tests start with Parallel or retain a documented reason; no baseline was relaxed. |
| TestCallTargetReaders | 4.430 | Incoming call-target readers retain the shared enumerator contract. |
| TestRuntimeSetsAreEachUsed | 6.760 | Source-profile witness enables exactly receiver plus regexp callback, through main shared runtime declarations. |

No new compiler-consuming package was introduced. internal/native/runtimeproducts is already declared in cloud/fast-gate/compiler-dependencies.json; changed-packages.json records the audit. Products preserve main's existing placeless Inputs and recipes; only the feature set/variant declarations were extended. No product or test source names a build directory or a machine path as an identity.

Focused validation used the source's complete new/touched-test declaration inventory, with current guards and bounded main compatibility controls. run-focused.py compiles test binaries and runs exact top-level tests with -test.timeout=90s, -test.parallel=1 and an outer 95-second leaf bound, two at once. Its watched 840-second driver ended before its last 31 tests. resume-focused.py ran only those unfinished tests; no completed or over-budget leaf was rerun. A declaration audit found no added/touched test declaration missing from selected-focused.json. Two obsolete portability test names were absent on main; current TestConstantPortabilityClang, TestLongUnicodePortabilityClang and TestLongUnicodeLiteralByteCoverage passed in reader validation instead.

All 289 focused tests/products completed with 7 documented skips. Maximum observed command wall time was 38.746 s. Each leaf's command wall seconds and result are in test-seconds.csv and focused-results.jsonl; full per-test logs are beside them. The oracle-focused helpers run source Node controls, native ASan/UBSan/LSan, release native and backend Node as prescribed by each positive/refusal fixture. No new test function was added by this recut beyond the source's carried functions.

Documented focused skips:

- internal/native TestViewAdapterPoolTSanPending
- internal/native TestViewAdapterProgramUnderlyingPending
- internal/native TestViewAdapterRegionAdoptionPending
- internal/native TestViewAdapterRuntimeStaticsPending
- internal/oracle TestV4EscapeAdapterBlameRuntimeCallback
- internal/oracle TestV4Lane5FactoryReadonlyTargetControl
- internal/oracle TestV4Lane5FactoryWiderWritePending

The local reader cold-build failure was observed:

    bridge/tsgo TestProduct_ApiDriver
    command wall 80.556 s; test reported 80.48 s; exit 1
    products_test.go:331: build bridge-test-tsgo-asan.a: tsgo-asan.a: signal: killed

Its own build deadline killed the build. cgroup memory.events reported oom=0 and oom_kill=0; this is not evidence of an OOM or silent miscompile. It exceeded the local 60-second leaf target and was run once, never rerun. This was a cold product build on this box, outside V4's territory. The user supplied Loom's record over the last day: bridge/tsgo p50 18.7 s, p90 25.9 s, 11 of 11 passed, because Loom serves the product from its cache. These Loom figures are user-provided evidence, not independently queried in this session. The user explicitly authorized pushing on that basis. bridge/tsgo/product_units_test.go and products_test.go were not edited or weakened; this local timing is not attributed to V4.

All 52 changed/direct-reader packages compiled successfully; consumers.json names every package and selected-consumer-tests.json every bounded sample. Remaining consumer testing stopped after this local cold-build failure; that coverage limit remains recorded despite the subsequent push authorization. consumer-results.jsonl records completed packages; consumer-driver.log also records completed leaves beyond the last buffered package result. consumer-stop.json records terminated child tests/compilers so they did not outlive the run. Incomplete consumer checks are not claimed as passes. No whole package test run or full gate was run.

Recorded source mutants were run in /tmp/adamic-loom-v4-mutants, a detached current-main worktree with the exact recut patch, referring to cohere through a symlink to the pinned submodule. No cohere implementation was copied. Each mutant ran alone with go test -run '^<catcher>$' -count=1 -timeout 90s, outer 120 s, ADAMIC_GATE_UNCACHED=1. All exited 1 from their expected assertions or sanitizer observations, not compiler warning failures. Mutant wall times below include Go rebuilding; the individual Go test leaf times remain in the logs. All 17 mutated files match the final tree byte for byte (mutants-restored.json).

| Mutant | Catcher | Wall seconds | Failure |
|---|---|---:|---|
| escape-Argument | TestV4EscapeArgumentRefusal | 58.167 | required path/fix refusal missing |
| escape-Result | TestV4EscapeResultRefusal | 11.088 | required path/fix refusal missing |
| escape-Passed | TestV4EscapePassedRefusal | 12.778 | required path/fix refusal missing |
| escape-Returned | TestV4EscapeReturnedRefusal | 12.737 | required path/fix refusal missing |
| escape-MixedProducer | TestV4EscapeMixedProducerRefusal | 13.156 | required path/fix refusal missing |
| unsupported-domain | TestV4DirectRecursiveRefusal | 26.935 | required path/fix refusal missing |
| unsupported-producer | TestV4DirectProducerRefusal | 36.645 | required path/fix refusal missing |
| surface-borrow | TestV4EscapeAdapterName | 45.484 | ASan heap-use-after-free, decoded in sanitizer-mutants.json |
| bind-name-borrow | TestV4EscapeAdapterBindNameOwnership | 37.506 | ASan heap-use-after-free, decoded in sanitizer-mutants.json |
| identity-root | TestV4EscapeAdapterIdentity | 37.277 | Node stdout comparison |
| metadata-inline | TestPortElementMetadataEmission | 13.800 | 2,885 direct stores exceed bound zero |
| metadata-omitted | TestPortElementMetadataEmission | 9.968 | zero runtime certificates instead of 2,885 |
| native-order | TestV4DirectCalleeOrder | 13.199 | argument effects missing after premature callable check |
| javascript-order | TestV4DirectCalleeOrder | 12.440 | argument effects missing after premature callable check |
| construct-witness | TestV4EscapeAdapterWithoutConstructSignatureRefusal | 3.193 | required path/fix refusal missing |
| boxed-witness | TestV4Lane5BoxedAdamicRefusal | 6.235 | required unproven-read refusal replaced by another refusal |
| factory-witness | TestV4Lane5FactoryWiderWriteRefusal | 6.087 | required path/fix refusal missing |
| runtime-witness-mismatch | TestRuntimeSetsAreEachUsed | 9.569 | dedication enables [] instead of receiver plus regexp callback |

The new guard mutant was restored in a finally block; the unmutated guard had passed on those same source bytes. final-vet.log records exit 0 with no diagnostics for all ten changed Go packages. gofmt and diff whitespace checks cover implementation/test sources; source-review patches and raw logs retain their evidence content.

Setup was run with GOPROXY='https://proxy.golang.org|direct' and timeout 600 bash cloud/setup.sh, then its printed /workspace/adamic-tools/env.sh was sourced in build/test shells. setup.log records every line: Node ready 0.054 s; Go ready 0.067 s; clang ready 0.411 s; markdown ready 0.936 s; submodules ready 175.755 s; shared cache ready 180.984 s; go build ready 534.762 s; test binaries deferred 534.895 s; build cache warm 534.899 s; done 534.967 s. Versions: clang 20.1.8, Go 1.27.1, Node 24.19.0. nproc=5, cgroup cpu.max=400000 100000 (four CPUs). Initial guard/vet/count commands reached outer limits while the cold frontend was compiling and ran no tests; after warming, every required command above completed. No over-budget test leaf was rerun.

The user explicitly authorized pushing compiler/loom-v4-3 after identifying the local bridge product cold-build timing as outside V4 and supplying Loom's cached-product record. Preserve that observed local failure and the incomplete consumer coverage without attributing them to V4. There is only this one requested branch. No PR or Loom submission is made.
