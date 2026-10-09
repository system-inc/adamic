Built: records and MapLike rebuilt on c81ba261 for task #38vbx51 and the views V5 prerequisite; admission proof is red.
Commits: 542bddd72702a57f45568118535cc1707157a832 and a2a63f164ec9eafa11bee2c00303ea9da0fbc49a; 96a66d06 records counts, status pins and validation evidence on compiler/records-next.
Checks: focused fixtures, lower shards, entries, native checks, source/runtime mutants, TestCallTargetReaders, counts verification and named stage 1 gaps pass; admission-delta reports fail.
Mutants: all twelve lane source overlays plus the nullable-slot port guard overlay, three named source mutations, eleven operation mutations and the runtime/ownership/readiness/narrowing mutants were caught; details below.
Uncovered: a green Node admission proof cannot coexist with the lane's eight pinned checked stops; V5 dictionary views remain guarded and runtime clearance is pending.

The branch starts directly at origin/compiler/optional-presence-next c81ba26193ce58fe7eb8c15d3ea1e46a58052fef. Only b58629a2's parent diff was applied with cherry-pick --no-commit. No records branch merge history or unlanded side stack was merged. The complete commit message and stage3/records-maplike-rebuild/REPORT.md were read first. Historical reports and evidence are preserved under inherited/ and are not observations of this head.

The implementation supplies the record representation and MapLike lowering required by task #38vbx51. V5 prerequisites are present: recordElement in internal/lower/records.go, recordSlot in the same file, finitePartialRecordElement in internal/lower/records_partial.go and ir.RecordCall in internal/ir/records.go. This is a prerequisite port, not V5 admission. The chain's dictionary view contract boundaries remain in place. No partially implemented view is admitted: the 142 ruling remains binding.

Conflict decisions:

* Preserve the chain's typed arrays, Buffer, views, adapters, generators, exceptions, iteration, readiness and all refusal predicates except the blanket index refusal replaced by record admission.
* Combine constructionJSON enumeration with record enumeration and existing plain object enumeration in library_for_in.go.
* Keep checkedJSONIndexSignature, placeholderDeclaration, checkedAssertionSource and placeholderNullishTest in refusals.go alongside detachedOwnRefusal, recordLiteralRead, recordStorageView and explicit record storage boundaries.
* Preserve the chain nullable-union boxing. Refuse nullable dictionaries separately and retain the records lane's unsupported nullable scalar slot boundary in recordStorageType. The focused refusal tests caught the initially too-broad nullable scalar admission.
* Do not import the older blanket delete predicate over the chain's proven optional-field deletion. Retain the chain's optionalPresence checks and expression delete refusals; put dictionary deletion beside them. OptionalDeletion tests and all lower shards pass.
* Reuse the chain's reserved ir.Record enum value, preserving existing numeric tag order. Add its reference classification and update its comment.
* Keep chain test parallelism. Update two old optional-index diagnostic pins to the actual rejected storage/index boundaries; no rejection assertion was removed.
* The source Node oracle now exposes an uncaught ReferenceError as exit 1. The two readiness mutation tests require both that exit and ReferenceError in stderr, then require the mutant to finish cleanly. They no longer assume the old wrapper's exit 70.

Admission proof:

The command comes from origin/compiler/admission-delta 841e335cf791ca7196f3719256e0187301b833f2, used in an isolated checkout without merging that branch. Its original automatic base build exceeded its five-minute hard limit. Local base and head binaries were then built separately. The subsequent delivery commits change counts, stage0 status pins and review evidence only; no compiler/runtime Go or C source differs from the proved head a2a63f16. Both binaries have only the tool's admission-lower init adapter; base sources are c81ba261 and head compiler sources are a2a63f16. The head adapter is recorded as admission-head-main.go.txt with its Go overlay. Base module paths reference the existing cohere submodule rather than copying its files.

Command: /tmp/records-next-admission-tool --base c81ba261 --head a2a63f16 --base-binary /tmp/records-next-base-adamic --base-lower-binary /tmp/records-next-base-adamic --head-binary /tmp/records-next-head-adamic --head-lower-binary /tmp/records-next-head-adamic --manifest review/compiler/records-next/admission-manifest.json --manifest-generator cloud/admission-corpus/manifest.py --manifest-generator-revision origin/compiler/admission-delta --workers 4 --compile-timeout 60s --timeout 30s --json. External hard limit: 600 seconds, checked while running. The pinned manifest includes witnesses, fixtures, named gaps, review programs, fuzz and the changed program diff, with no filters, omitted inputs or runtime sampling budget.

The first complete run at a 15-second compile limit observed 1,145 unique programs: 975 accepted by both, 138 refused by both, 31 newly accepted and one base timeout in stage1/cohere/css/gaps/printer_boundaries.ts. All 31 new admissions were run against source Node, emitted JavaScript and native. Eight disagree. The longer-limit complete run in admission-final-60.json has no compiler errors or timeouts: 976 accepted by both, 138 refused by both, 31 newly accepted, all 31 sampled with no omissions, and the same eight disagreements. Verdict: fail. Total proof runtime: 70.150 seconds.

The eight witnesses are records_compare_properties_left.a, records_compare_properties_right.a, records_environment_boundary.a, records_named_invalidated.a, records_narrowed_number.a, records_prototype_in.a, records_prototype_read.a and records_prototype_set.a, all under internal/oracle/testdata/. Node exits 0; both compiled backends exit 70. Examples: Node prints the inherited toString function in records_prototype_read.a; the lane deliberately stops with "records hold own keys only". Node prints NaN after deletion in records_narrowed_number.a; the lane stops because a call invalidated narrowing. These are precisely the checked stops required by the inherited oracle fixtures and their mutants, not newly invented port behavior. No fixture or check was deleted, no prototype guard was weakened and no proof input was filtered to hide them. Consequently this candidate is not merge ready under the requested every-new-admission-agrees-with-Node rule. Producing a green proof requires a ruling that changes these lane admissions or their pinned checked-stop requirements.

Validation:

All commands wrote directly to logs. Go commands use ADAMIC_GOCACHE_OFF=1, source /workspace/adamic-tools/env.sh, GOPROXY=https://proxy.golang.org|direct and TMPDIR=/workspace/records-next-scratch. The shared cache and /tmp capacity interfered with initial builds; local caching and workspace scratch resolved that. Every long command had an external hard limit. No whole compiler package suite or full fast gate was run.

* Focused lowering: go test ./internal/lower -run 'Record|DetachedOwn|Enumeration|Entries|LibraryMethod' -count=1 -timeout 85s; lower-focused.log passes.
* All internal/lower test names split into twelve groups, each go test -run anchored to its recorded names -count=1 -timeout 85s, external timeout 89. lower-shards.json records every name and all twelve exit 0. The two stale optional-index pins were corrected and shards 7/8 rerun. OptionalIndexing and OptionalDeletion were also run separately and pass.
* Focused oracle fixture set: oracle-selection.json records 137 fixture paths. Eight TestNativeAgreesWithNode shards in run-oracle-shards.py passed with ADAMIC_GATE_UNCACHED=1. Four record/detached-own mutant groups passed; groups 2/4 were rerun after the ReferenceError pin correction. oracle-shards.json records the final outcomes and exact patterns.
* Main entries acceptance: TestEntriesAcceptance, TestEntriesProvenance and TestEntriesRuntimeReadiness pass (entries.log). TestRecordCensusComparisonBuckets passes.
* Native: TestRecordsAgainstNode and TestRecordReadMutants pass (native.log). TestRecordMutantsUnit00 through Unit05 pass (native-mutants-final.log); all leaves are below the grain limit.
* Source mutants: run-source-mutants.py catches all twelve lane overlays; the additional nullable-slot guard overlay is caught by TestNamedRecordRefusals; run-named-mutants.py catches all three named-contract mutations. Source files were restored before final validation. Overlay evidence is .go.txt, never compilable .go.
* TestCallTargetReaders passes, latest run 19.297 seconds (call-targets-final.log).
* Counts regenerated exactly once successfully with TestCountsAreRecorded -update-counts, then independently verified without update (counts-final.log and counts-verify-final.log). An earlier cold command timed out without updating the file. counts-audit.json records all 33 added rows, all three moved rows and no removed rows. library_string_raw loses four retain/release pairs through record raw storage; method_coverage_object_descriptors and library_method_values replace the detached own helper closure with a readiness marker. The observed deltas match the lane's attribution.
* Stage 3: TestFixturesRecords, TestFixturesObjects and TestFixturesTaste pass, 11.868 seconds. The lane's old Compiles pin for records/01_has_property is corrected to the chain's explicit-any refusal; objects/06_watch_close retains the strong capture cycle refusal; taste/11_build_info_pending is refreshed using normal -update only after native and both recorded/current Node agree. Only stage0 JSON values changed.
* Stage 1: go test -p 4 ./stage1/... -run 'Gap|Gaps|Probes' -timeout 85s -v exercised named gap checks. Under initial concurrent build load estree and comments reached the package limit. The comments TestJsxParserGapIsExplicit rerun passes in 37.02 seconds. All estree gap names were split into four groups and pass, including TestParserRecoveryGap in 46.09 seconds. Existing library-oracle skips remain because optional pinned npm inputs were not provisioned; their exact names and reasons remain in the logs. stage1-estree-results.json and test-leaf-times.json record names, outcomes and seconds. Other packages' named gap/probe checks pass in stage1-final.log.
* Lane checks after committing: git fetch -q origin main devtools/fast-gate cloud/merge-tree; git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -. The first result passes in 58.5 seconds: gofmt/tools on 319 Go files, t.Parallel on 23 test packages and a-check on 103 .a files; vet exceeded that script's ten-second allowance and was skipped. Separate go vet ./internal/lower ./internal/oracle ./internal/ir ./internal/native ./internal/fresh ./internal/flow ./internal/javascript passes with empty output. Final lane output passes in 11.6 seconds: gofmt/tools on 319 Go files, t.Parallel on 23 test packages, a-check on 103 .a files and vet on 23 packages. It is retained as lane-checks-final.log.gz.

Every mutation and its catcher:

| Mutation | Observed catcher |
| --- | --- |
| readonly-index, numeric-index, storage-views, mutable-invariance, record-cycles, spread-invariance, logical-record-conversion, nullable-record-container | TestRecordRefusals, intended admission assertion fails |
| prototype-literal | TestRecordPrototypeLiteralNames |
| opaque-own-argument | TestDetachedOwnRepresentation |
| shorthand-own-escape | TestDetachedOwnRefusals |
| type-only-index-admission, restoring blanket index refusal | TestRecordForms |
| nullable-slot guard removed | TestNamedRecordRefusals, union-valued dictionary becomes incorrectly admitted |
| named index-read-type | TestNamedRecordReadTypes |
| named unrestricted-write and erase-named-contract | TestNamedRecordRefusals |
| read, write, delete, in, hasOwn, keys, values, entries, spread, for-in, stringify | TestRecordOperationMutants, sanitizer-clean mutated output differs from Node |
| inherited record read | TestRecordPrototypeMutant, required missing-member checked stop |
| dropped record releases | TestRecordOwnershipMutant, LeakSanitizer |
| eager fallback | TestRecordCoalesceMutant, Node stdout |
| removed scalar narrowing guard | TestRecordNarrowingMutant, required checked exit 70 |
| removed own-read guards: discarded, guarded snapshot, scalar comparison | TestRecordObservationGuardMutants, required checked stop |
| detached own changed to inherited membership on record/fixed object | TestDetachedOwnInheritedMutant, Node stdout |
| removed detached alias readiness | TestDetachedOwnReadinessMutant, Node ReferenceError exit 1 versus clean mutant exit |
| partial/named absent entries inserted as present undefined | TestPartialRecordAbsentEntryMutant and TestNamedRecordAbsentEntryMutant, both backend stdout |
| removed named member-kind guard | TestNamedRecordTypeGuardMutant, required checked exit 70 |
| removed named alias-copy readiness | TestNamedRecordAliasReadinessMutant, source ReferenceError versus clean mutant in both backends |
| integer insertion order, uint32 maximum, deleted-key iteration | TestRecordMutantsUnit00/01/02, Node stdout |
| overwritten key leaked | TestRecordMutantsUnit03, LeakSanitizer |
| stored key freed | TestRecordMutantsUnit04, AddressSanitizer heap-use-after-free |
| own slot silently null | TestRecordMutantsUnit05, own-hit contract/UBSan |
| inherited membership, missing read silently null, own hit treated as missing | TestRecordReadMutants, exact own-read behavior and checked diagnostic |
| scanner end-position mutation | Independent scanner reference diff exits 1; unmutated control diff exits 0 |

Scanner reproduction uses the 3ea66219 driver by git show, without merging it. Both ADAMIC_NATIVE_SPLIT=0 and 1 stop at debug.ts:14:14, function viewed as unknown or object, after the MapLike/type-only dictionary barrier, exactly as the report requires. The independent Node reference covers 81 files, 1,369,432 tokens, 466 errors and 108,024,471 bytes, SHA256 a4a83298df7d9a6b1b779da1a02a40d26dc5c942dfa25870b7a378867f54e691. Control agrees; end-position mutant fails. The complete upstream scratch trees and huge streams stay in workspace scratch; scanner-evidence/ retains reports, hashes, inventories and stop diagnostics. No scanner-native end-to-end claim is made.

Setup observations: nproc=5; quota is four CPUs. First setup timing lines: Go .083, Node .168, clang .561, markdown duration 1.029 / ready 1.302, submodules 19.731 seconds. The initial setup overlapped my checkout, causing its Go dependency list to omit newly introduced project-loader files; it failed with undefined projectOptionsForRoots and related symbols. This overlap was my mistake. Stable retry: Go .076, Node .083, clang .523, markdown duration .013 / ready .219, submodules .227, shared cache ready 7.274; its dependency warming did not complete within the 240-second hard limit. No successful setup "done" line exists. Subsequent local builds and tests used Go 1.27.1, Node 24.19.0, clang 20.1.8. setup.log and setup-retry.log preserve the observations. No cohere source was copied.

Runtime clearance required from @system_adamic_runtime, exact changed runtime paths:

* internal/native/runtime/json_stringify.c
* internal/native/runtime/json_stringify.h

No clearance has been granted in this turn and no message was sent to the owner. This delivery exposes the preserved lane and the binding-rule conflict for review; it does not claim a green admission proof or completed V5 dictionary support.
