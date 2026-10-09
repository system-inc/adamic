Built compiler/optional-presence-next by merging the verified lowering chain and eep-presence onto optional-presence-chain.
Source commits: 5d6c68f3 (chain merge), 39c749d2 (eep-presence merge), 167aaf0c (reader reconciliation and regenerated counts).
Checks: lower 341 pass / 3 existing skips; three former reds pass alone; optional fixtures, both backends, sanitizers, seven refusals, supported neighbor and reader guard pass.
Mutants: optional runtime/IR/JavaScript checks, seven source/catalog overlays, construction metadata omission and guard revert all caught.
Not covered: whole repository gate, full native/oracle packages, WASI, Darwin, performance, or runtime implementation of the seven refused programs.

This is the explicitly requested integration branch for #r3chqza optional presence and #eep1m5z review-lane containment. It starts at 897d0e7989ecfc2175f25966b8def3efe51b1ec8, merges verified chain tip 50654a40 (including a6cb5660), then merges eep-presence through 21ceb23c. Neither parent branch is rewritten or pushed. No additional runtime semantics were introduced during reconciliation.

Two chain-merge conflicts were resolved. internal/lower/interface_cast_test.go retains the optional-presence admission of optional reads and its located refusal for optional nullable conversion, plus the chain's observable optional-target control. internal/ir/call_targets_guard_test.go retains current reader metadata and drops the obsolete throwsOutReadiness direct-sort reader; after the automatic merge retained the newer static test's CallTargets migration, its direct Call.Function allowlist entry was also removed. TestCallTargetReaders passes. The structural Map boundary in optional_chain.go, chain exception behavior, optional storage checks, and all three eep property guards remain. The eep merge had no conflicts.

Counts were regenerated once in full, not patched from a selected subset. `counts-attribution.md` names every changed/new numerical row relative to 897d0e79, and the order-only relocations. Ten newly measured values were stale on that parent: independent counted observations with freshly built exact-parent compilers confirm all ten match 897d0e79 behavior and that 50654a40 matches the former table. Sixteen added chain fixture rows and four callback-search corrections are attributed to their recorded chain commits. The complete writer passed in 283.723 seconds and retained all 1,320 runtime fixture rows and predicate-direction counts. Exact commands, values, source revisions and logs are retained here.

The lower suite was enumerated from the merged tree into 344 top-level tests, split into 237 disjoint prefix shards. Each command used `go test -buildvcs=false ./internal/lower -run '^PREFIX' -count=1 -timeout 80s -json`, with an external process-group limit of 85 seconds. All 237 commands pass; the union is 341 passed and 3 skipped, with no missing or duplicate top-level verdict. Maximum shard wall time is 24.305 seconds. Existing skips are TestGenericBodyRelationsCensus, TestOptionalWideningCensus and TestOriginalCycleLedger; inherited nested pending skips remain. lower-summary.json, shards-plan.json and shards-results.jsonl provide the full census and commands.

The three tests previously failing on 897d0e79 now pass both in the full suite and alone, with exact anchored selectors and -timeout 90s: TestUndecidedCycleReadsUseReadyChecks (7.204s command wall), TestOptionalIndexingMapShapeRefused (3.115s), and TestOptionalIndexingKeepsUnsupportedStorageNotYet (3.164s). three-tests.json and individual logs are authoritative. This confirms the a6cb5660 corrections coexist with optional-presence support.

Optional-presence validation ran uncached in 21 bounded groups: 20 top-level optional-field/literal tests, plus TestNativeAgreesWithNode's 15 optional_field fixtures. Every command passes. Commands use -count=1, -timeout 80s and an 85-second external process-group limit. Maximum wall time is 25.009 seconds; maximum top-level test duration is 8.13 seconds. The tests compare source Node with JavaScript and release native plus ASan/UBSan native; successful fixtures retain their leak checks. Existing intentionally checked failures are held to their explicit pinned diagnostic rather than claimed identical to unchecked TypeScript. optional-fixtures.json lists the 15 differential fixtures, and optional-results.jsonl records every command and verdict.

All seven eep witnesses were freshly observed on source Node and compiled in release-native, sanitized-native and JavaScript modes. Each mode gives a located NotYet with a workaround, empty compiler stdout and no internal runtime stop. results.json and refusals-verify.log record these checks. The existing supported tuple/array/ordered-static neighbor prints `2\n2\nok\n` on source Node and identically in release native, ASan/UBSan/LSan native and the JavaScript backend. supported-results.json records the four observations and builds. In the lower suite the eight presence leaves took 0.10 to 0.35 seconds; no new Go test leaf was added.

Mutants rerun and their catchers:

| Mutation | Catcher |
| --- | --- |
| Drop omitted optional slot | TestOptionalFieldWriteCatchesDroppedSlot: restored exit-70 missing-field stop differs from Node |
| Initially present omitted slots, layout rank instead of insertion rank, suppressed deletion | TestOptionalFieldPresenceCatchesMutants: clean wrong stdout differs from Node |
| Drop spread reservation | TestOptionalFieldConstructionCatchesDroppedReservation: missing-field stop differs from Node |
| Restore static key enumeration | TestOptionalFieldAliasCatchesStaticEnumeration: stdout differs from Node, release and sanitized |
| Drop copied presence, readiness, representation; overlap rank/readiness | TestOptionalFieldCopyState: state assertions, release and sanitized |
| Required lookup replaces optional lookup | TestLiteralOptionalOracleCatchesMutant: missing-field stop after the green control |
| Drop checked store | TestOptionalFieldCheckedViewCatchesDroppedStore: clean wrong stdout differs from Node |
| Missing publication, missing write guard, absent read as zero, boolean decode, missing union boxing | Corresponding TestOptionalFieldCheckedViewCatches tests: native release/sanitized and generated JavaScript observations |
| Omit finite optional literal check | TestOptionalFieldCheckedViewLiteral: pinned diagnostic in both backends |
| Borrowed string self-store without retain | TestOptionalFieldCheckedViewStringSelf: ASan heap-use-after-free |
| Reject optional string undefined; allow required string undefined | TestOptionalFieldCheckedViewStringUndefinedMutant and RequiredStringMutant: both backends' pinned contracts |
| Drop boolean reservation | Source overlay: TestNativeAgreesWithNode optional_field_checked_view catches field write failed: flag |
| Drop string retain in emission | Source overlay: TestOptionalFieldCheckedViewWrites catches Sanitizer failure |
| Cast-origin bypass; plain-origin bypass | Source overlays: TestOptionalDeletionRequiresPlainStorage catches wrongly admitted class storage |
| Names-only hasOwn proof | Source overlay: TestObjectUnprovenShapesStayNotYet catches wrongly admitted getter shape |
| Stale binder refusal registration | Source overlay: TestNativeAgreesWithNode/17_binder_flow rejects stale NotYet expectation |
| Literal undefined catalog patch | Source overlay: TestNativeAgreesWithNode/e4eec87_u01_undefined_field_widened catches stdout difference |
| Drop construction optionality metadata | TestNativeAgreesWithNode/optional_field_checked_view_string_undefined catches expected string, found undefined in both backends |
| Remove real eep property guard | TestEEPPresence0 through 6 all fail on wrongly successful lowering; supported control still passes |

The source overlays, complete generated mutant sources (.go.txt), catalog patch and logs are in source-mutants/. run-source-mutants.py returns 0 after all seven semantic catches and rejects build failures as kills. run-destination-mutant.py returns 0. run-guard-mutant.py removes the real hook, requires all seven fixtures to fail, and restores object.go in finally; its summary and output are preserved. The restored TestEEPPresence selection passes in 0.250 seconds. No compiler warnings or build failures are counted as mutant kills.

TestCallTargetReaders passes in 7.119 seconds. Initial combined optional runs reached their outer limit (one included cold compilation); they are superseded by the complete bounded group union. Its runner initially failed to create a log filename containing a nested selector slash, after 20 groups passed; the filename was fixed and only the unexecuted 15-fixture group was run, passing. The first supported JavaScript invocation lacked oracle/node.mjs's module adapter; it was corrected and all four observations rerun. The initial reader export scan reached its outer bound; the warm complete reader run passes. These infrastructure/probe attempts are not counted as semantic passes.

Setup used GOPROXY=https://proxy.golang.org|direct and the printed /workspace/adamic-tools/env.sh. Timings: Go ready 0.042s, Node ready 0.062s, Markdown ready 0.136s, submodules ready 0.138s, clang ready 0.431s, build ready 137.132s, cache warm 138.297s, done 138.545s. nproc is 5; CPU quota is 4. Long commands had explicit limits and background work was inspected during execution. Every test output went to a log. The full repository gate was not run.

Final committed-state integration lane checks returned 0 after evidence commit 12a9c350: `lane checks 20.6 s: gofmt and tools on 293 Go files, t.Parallel on 23 test packages; a-check 101 .a files; vet 23 packages`. lane-checks.log and lane-checks.status.log retain the result. The required lane fetch preceded the check. Final changes after that check are review evidence only. Raw logs preserve tool-emitted trailing whitespace; source and report whitespace checks pass.
