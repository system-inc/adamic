Merged current main for lowering landing candidate #wj4pmt1, preserving both sides’ dependency declarations and WASI guards; regenerated measured host status.
Merge commit: 5a8fd7142846b471613cc3a65f946f304152ad65; main parent: b8bcadb2c493173855f19d7e5c508b34f5eeb5b6; previous branch head: 0292780a0ab4549a0de2d24702167b9773558085.
Validation: lower 342 pass/3 existing skips; optional 21 groups pass; host 25 pass; WASI 39 and associated native 12 pass; reader guard and counts pass.
Mutants: optional runtime/JavaScript tests, seven source overlays, destination metadata, and actual guard revert caught; details below.
Uncovered: three existing lower census skips, opt-in split checker archive, shared Go-cache helper path and full repository gate; seven unsupported witnesses remain located refusals.

This branch carries the optional-presence fix, a6cb5660’s lowering corrections and the eep containment guards into the main-based candidate that integration can cut for #wj4pmt1. Ancestry checks confirm 897d0e79, a6cb5660, 21ceb23c and b8bcadb2 are present. No PR was opened.

The add/add compiler-dependencies conflict now includes all three consumer entries from both parents: cmd/adamic-gocacheprog, stage3/checked-any/probe and stage3/project-loader/probe. Every parent dependency is retained. The actual reader from origin/devtools/fast-gate accepted the candidate; all 19 ReverseDependencies reader tests passed. check-dependencies.py, dependency-census.log and dependency-reader-tests.log preserve the reader invocation and results. The harness initially unpacked the reader’s return value incorrectly; it was corrected to the actual API and rerun successfully.

The WASI merge retains main’s phase logging and this branch’s feature-aware runtime compilation and cache keys. Generated program feature flags reach every runtime translation unit and link, including counted request execution. Both parents’ test sets are present: main’s 36 leaves and this branch’s 39 leaves. The strict C11 runtime compilation, Node stdout/stderr/exit comparisons, request bounds and counted-memory guards remain intact. merge-invariants.json and wasm-parent-diff.patch record the comparison.

Host status was measured, not selected from a parent. The exact repository writer was:

`timeout 600 python3 stage3/fixtures/host/check.py --update-stage0 --compiler-repo /workspace/adamic --logs /workspace/scratch/host-main-observation`

All 25 sources were observed on Node, and successful native builds were compared with those observations. The writer regenerated stage0 outcomes and diagnostics; the source Node measurements and provenance are preserved. In particular 14_getCurrentDirectory now names line 13. host-regenerate.log and host-observations/ contain the measurements, including the frontend overlay saved as load.go.txt. All 25 fixtures also pass TestFixturesHost in five bounded selector groups; host-shard-results.json and host-union.json give complete coverage. Combined cold host runs hit their bounds and are superseded by this passing union.

Full internal/lower was enumerated with go test -list, then split by test-name prefix into 238 shards covering 345 top-level leaves. All 342 executed leaves pass; TestGenericBodyRelationsCensus, TestOptionalWideningCensus and TestOriginalCycleLedger keep their existing skips. Maximum shard wall time is 26.256 seconds; maximum leaf time is 24.240 seconds. The first 21 shards used go test. To avoid repeated linking, the remaining 217 used a test binary compiled from this exact merged tree, invoked through go tool test2json with -test.run, -test.count=1 and -test.timeout=80s, from internal/lower. Each shard has an 85-second external process-group limit. The initial runner was deliberately stopped; completed results were retained and unfinished selectors rerun. lower-union.json asserts the exact enumerated leaf union without omissions or duplicates. TestUndecidedCycleReadsUseReadyChecks, TestOptionalIndexingMapShapeRefused and TestOptionalIndexingKeepsUnsupportedStorageNotYet also pass isolated anchored go test selectors (16.927, 4.359 and 6.584 seconds wall time).

Optional-presence ran uncached in 21 bounded groups: all 20 optional-field/literal top-level tests and TestNativeAgreesWithNode’s 15 optional_field fixtures. All pass, maximum group wall time 28.495 seconds. The fixtures compare source Node with JavaScript and native release plus ASan/UBSan and leak checks; deliberately checked failures retain their pinned diagnostics. All seven eep witnesses were freshly observed on Node and compiled in release-native, sanitized-native and JavaScript modes. Every compiler response is a located NotYet with a workaround and empty stdout; no runtime internal stop occurs. The supported tuple/array/ordered-static neighbor prints 2, 2 and ok identically on source Node, JavaScript, native release and ASan/UBSan/LSan native. See optional-results.jsonl, results.json, refusals-verify.log and supported-results.json.

The complete counts writer ran once after the cache-enabled pre-test attempt was canceled:

`timeout 650 go test -buildvcs=false ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -parallel 4 -timeout 10m -v -args -update-counts`

It passed in 368.632 seconds. The regenerated file is byte-identical to the merged table: 1,320 runtime rows and 15 predicate direction rows, with no moved, added or removed row. counts-attribution.md records this; the prior turn’s row attribution remains in ../counts-attribution.md. No second successful full-table regeneration was performed.

All 39 TestWASIUnit00 through 38 leaves pass without toolchain skips, each separately bounded below 90 seconds (maximum wall 15.064 seconds). This includes arguments_length_extended, closure_convention_regexp_count, closure_convention_nested and counted WASI requests. All 12 associated native leaves pass: both RuntimeFeatureMismatch tests, RuntimeFeatureSetsNative, RuntimeFeatureSetsWASI, TSGoBuildSeesTheProgramsFeatures, RecordsAgainstNode and RecordMutantsUnit00 through 05. Maximum wall across these checks is 74.250 seconds; each command has an 80-second test timeout and 85-second external bound. Feature matrices each check 32 matching masks, 160 mismatches and an unused mixed unit. Five completed Go test invocations were retained; the remaining 46 leaves ran individually from the exact merged native test binary through test2json, with package working directory and WASI SDK environment. native-union.json verifies all 51 terminal passes. No new Go test leaf was added; every merged WASI leaf is under the 60-second test grain. The opt-in split checker archive test was not run.

Mutants rerun on the merged sources:

| Mutation | Catcher |
| --- | --- |
| Omitted-slot loss; initially present slots; insertion rank replaced by layout rank; suppressed deletion | OptionalFieldWriteCatchesDroppedSlot / PresenceCatchesMutants: missing-field stop or clean wrong stdout held to Node |
| Lost spread reservation; static enumeration; copied presence/readiness/representation or overlapping metadata | ConstructionCatchesDroppedReservation / AliasCatchesStaticEnumeration / CopyState: Node differences or runtime state assertions, including sanitized native |
| Required lookup for optional literal; dropped checked store | LiteralOptionalOracleCatchesMutant / CheckedViewCatchesDroppedStore: Node difference |
| Missing publication, write guard or union boxing; absent read as zero; boolean decode | Corresponding CheckedViewCatches tests: native release/sanitized and generated JavaScript |
| Finite optional literal check removed; optional string undefined rejected; required string undefined accepted | CheckedViewLiteral / StringUndefinedMutant / RequiredStringMutant: both backends’ pinned contracts |
| Borrowed string self-store without retain | CheckedViewStringSelf: ASan heap-use-after-free |
| Boolean slot reservation removed | Source overlay: optional_field_checked_view fails with field write failed: flag |
| String retain removed from native emission | Source overlay: CheckedViewWrites catches Sanitizer failure |
| Cast-origin or plain-origin bypass | Two source overlays: OptionalDeletionRequiresPlainStorage catches wrongly admitted class storage |
| Names-only hasOwn proof | Source overlay: ObjectUnprovenShapesStayNotYet catches wrongly admitted getter storage |
| Stale binder refusal registration | Source overlay: 17_binder_flow rejects stale NotYet expectation |
| Literal undefined catalog patch | Source overlay: e4eec87_u01_undefined_field_widened catches stdout difference |
| Construction optionality metadata removed | Destination overlay: optional_field_checked_view_string_undefined catches expected string, found undefined in both backends |
| Real eep property guard removed | All seven TestEEPPresence refusal leaves fail; supported control passes; source restored in finally |

Mutant sources are .go.txt, never census-visible Go files. Source/destination runners reject build failures as kills. guard-revert-mutant.log, guard-mutant-summary.log and guard-restored.log preserve the actual revert and restored check. TestCallTargetReaders passes in 5.979 seconds.

Setup began with GOPROXY=https://proxy.golang.org|direct. The shared-cache-enabled setup/build and host commands reached their hard limits before producing test results. The documented ADAMIC_GOCACHE_OFF=1 override completed setup and all verification with local Go caching; this is an infrastructure limitation, not a claimed semantic pass for the shared helper. Successful timing lines: Go 0.045s, Node 0.052s, submodules 0.179s, Markdown 0.257s, clang 0.433s, WASI SDK 0.703s, shared cache disabled 0.705s, build ready 123.347s, test binaries deferred 128.758s, build cache warm 128.761s, done 128.975s. nproc=5; cgroup quota=4 CPUs. All output went to evidence logs. TMPDIR was set to /workspace/scratch/eep-temp for verification to avoid the constrained /tmp mount.

Final committed-state lane checks on evidence commit a1dae0d3 returned 0 after the required fetch: `lane checks 18.9 s: gofmt and tools on 293 Go files, t.Parallel on 23 test packages; a-check 102 .a files; vet 23 packages`. origin/main remained b8bcadb2 after that fetch. lane-checks.log and lane-checks.status.log retain the result. Changes after that check are review evidence only. Source formatting was verified with the configured toolchain; gofmt made no changes. Every command, selector, exit and elapsed time is retained in the plans, runners, JSON unions and logs. The full repository gate was not run.
