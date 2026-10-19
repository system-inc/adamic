Fixed the records flow corpus refusal classification; current stage3 records pins pass unchanged.
Commit evidence: source c8fdf062f79b4a0e070b5b31e43886b6d242b5ac; base 199174800998904d8d646dfa5892f512cd5cb20f; measured squash 6e4ccee9ae4b246f517448651a6caac990354737; submitted SHA is the single branch HEAD reported at handoff.
Commands/results: both main baselines pass; fixed flow and uncached stage3 pass; 28 oracle fixtures, 30 counts rows, readers and vet pass; admission is 24/24 Node agreements with zero omissions.
Mutants: flow classification, stage3 classification, counts, 16 lowering guards, eight refusal witnesses, native runtime and oracle mutants all caught; names and catchers below.
Not covered: original stage3 Loom failure was not reproduced or diagnosed; its change endpoint returned HTTP 401. No full package gate, full repository gate, new benchmark or non-Linux run.

Observed cause and fix:

Main's flow programs() globs internal/oracle/testdata/*.a. The records branch adds eight deliberate compile-time refusal witnesses at that level. records_refusal_test.go already requires each witness, in .a and .ts form, to return lower.Refused with exact location, reason and fix. Flow attempted to lower them and failed before any graph existed. The unchanged squash reproduced all eight failures in TestFlowCorpusRemainder, including records_compare_properties_left.a:9:9. It ran 38 remainder programs in 5.22s.

The fix uses main's existing explicit refusal classification pattern: programs() now calls refusedRecordFixture(), which names exactly those eight witnesses. Main's non-null, constructor and getters-census classifications remain. The hidden-boundaries branch was read and its flow diff retained as hidden-boundaries-flow.patch.gz; no code or dependency from that branch was imported. The remaining 30 remainder programs receive the same single-assignment/read/reaching-definition, mutation-range, Node graph-path and liveness checks. No lowering error is generally ignored, and no records prefix is skipped.

On plain origin/main, flow passed in 1.795s and stage3 records in 24.163s. On the unmodified records squash, flow failed in 5.235s but stage3 records passed in 24.402s. After the flow classification fix, flow passed in 4.840s and stage3 records passed uncached in 9.910s. All 19 stage3 entries were checked, including six native fixtures. Stage3 already represents deliberate stops by exact Checker/Refused/NotYet status pins rather than treating them as runnable. Its current pins agree with the current compiler and source Node, so none were relaxed or changed. A controlled mutant falsely marking 07_optional_view.a Compiles fails exactly the stage0 status comparison while its Node comparison passes.

Inference/assumption: the stage3 pins are correct on this main tip. The historical stage3 red may depend on information unavailable here; no cause is claimed for it. Conservative choice: preserve the passing exact pins and report the reproduction limitation. This assumption is also in the commit message. No user question was needed to preserve the stricter checks.

Admission proof:

The existing admission tool was fetched from compiler/admission-delta at 4e0aab75a16382732b9f22fb53cbd1aa44e73d97 and built as a temporary local tool, then removed from the submission. No unlanded tool dependency was added. Complete generated manifest, no corpus filter, no shard, no sampling budget, four classification workers, compile limit 60s, runtime limit 30s. Result: 1,075 distinct programs, 883 accepted by both, 168 refused by both, 24 newly admitted, zero newly refused, zero omitted, all 24 agreeing with source Node, JavaScript and native. No negative-witness exemptions. Total measured tool time 119.124s. The proof measures the initial squash; admission-equivalence.json certifies the final change alters only flow's test classification and review evidence, with unchanged compiler/runtime sources and .a inputs.

Validation:

All exact commands, including failed setup/build attempts and discarded empty selectors, are in COMMANDS.md and commands.jsonl. Output always went to files. coverage.json lists all 28 actual registered oracle fixture paths (27 runnable and one pinned NotYet control) and all 30 measured count rows. These oracle fixtures ran uncached against source Node, backend Node and release/sanitized native, with leak checks on normal finishes. The 30 rows are the inherited records change's 27 added rows plus its three changed rows. TestCountsAreRecordedTable verifies the whole table's structure. No new fixture was introduced by this repair, and no counts changed, so the already refreshed inherited counts.md is retained.

Focused lower checks cover records, named/partial records, detached own-property calls and entries provenance. The two inherited optional-indexing diagnostic adaptations were separately rerun. TestCallTargetReaders passed in 27.21s, including its compiler-reader build. Vet passed on every changed Go package: flow, fresh, ir, javascript, lower, native and oracle. Formatting and git diff --check pass. This repair adds no top-level Test function; test-leaf-seconds.json records observed tests, with longest top-level execution 29.49s, below 60s on the four-CPU quota. Test command wall times additionally include compilation.

Every effective mutant and catcher:

- Flow refusal-list call removed: TestFlowCorpusRemainder exits 1 on all eight refusal witnesses. Its corrected run takes 14.24s. The first slash-filtered mutant selected no subtests and is explicitly discarded.
- Stage3 07_optional_view.a falsely marked Compiles: TestFixturesRecords exits 1, exact stage0 classification mismatch; source Node comparison passes.
- Recorded allocations for records_operations.a increased by one through main's family probe: TestCountsAreRecorded exits 1 with the recorded/measured count mismatch.
- readonly-index and numeric-index: TestRecordRefusals, missing signature refusal.
- prototype-literal: TestRecordPrototypeLiteralNames, missing literal-member refusal.
- storage-views: TestRecordRefusals, missing storage-view refusal.
- mutable-invariance and spread-invariance: TestRecordRefusals, missing invariant-mutable refusal.
- record-cycles: TestRecordRefusals, missing cycle refusal.
- logical-record-conversion: TestRecordRefusals, missing logical-record-operand refusal.
- nullable-record-container: TestRecordRefusals, missing value-type refusal.
- opaque-own-argument: TestDetachedOwnRepresentation, unsupported argument admitted.
- shorthand-own-escape: TestDetachedOwnRefusals, detached-method refusal missing.
- type-only-index-admission: TestRecordForms, valid type-only record rejected.
- index-read-type: TestNamedRecordReadTypes, named read uses index type.
- unrestricted-write and erase-named-contract: TestNamedRecordRefusals, named write/alias contracts erased.
- nullable-slot guard removed: TestNamedRecordRefusals admits a rejected nullable slot.
- recordDivergenceRefusal removed: prototype_read, prototype_in, prototype_set, compare_properties_left, compare_properties_right, environment_boundary, named_invalidated and narrowed_number all compile; source Node exits 0 while both compiled backends stop at 70. The refusal-removal harness requires the disagreement for every witness. The same overlay also exercises prototype read-through, stale scalar narrowing and declared-member-kind guard removal, caught by their checked-stop exit pins.
- Native runtime indices-in-insertion-order, uint32-max-as-index, deleted-key-iterated: Node behavior comparisons.
- Native overwrite-key-leaked: LeakSanitizer; stored-key-freed: ASan heap-use-after-free; own-slot-null-read: UBSan null access.
- Native prototype-membership-restored and missing-read-silent: exact diagnostic/exit checks; own-read-checked-as-missing: successful own-hit fixture.
- Duplicate replacer dedup skipped, independently for records and fixed objects: Node stdout disagreement, clean sanitizers/leaks.
- Oracle operation mutants read, write, delete, in, has_own, keys, values, entries, spread, for_in, stringify: Node stdout comparison with clean native exit and sanitizers.
- Discarded-read, guarded-snapshot and scalar-comparison own guards removed: native exit 70 versus source Node finishing, exact missing-member diagnostic.
- Record ownership releases removed: LeakSanitizer alone, preserving ordinary output.
- Coalescing fallback made eager/repeated: Node stdout.
- Partial absent entry and named absent entry defined as own undefined: both backends disagree with Node stdout.
- Named alias-copy readiness removed: both backends lose source Node's normalized ReferenceError exit 70.
- Detached own-property inherited-name mutants on records and objects: Node stdout; intrinsic readiness removed: source Node's initialization stop.

All 16 source/named/nullable guard test invocations exit 1 under their overlays. Runtime-mutant harnesses exit 0 only after observing the intended bad output, stop, sanitizer or leak report. No build-warning failure is counted as a caught mutant. .go.txt sources, overlays and individual logs are retained under this review directory. The first obsolete unsharded native-oracle selector ran no tests and is discarded; the corrected generated-family selector ran the 28 listed fixture leaves.

Toolchain and setup:

The managed environment started without a usable shell; runtime/network skill was read and the network-policy snapshot inspected. First setup was bounded at 240s and expired after submodules became ready at 217.470s and the shared cache helper at 223.538s. Second setup was bounded at 360s and expired during warming. Their logs are retained; no completed setup is claimed for those attempts. Local-cache setup, with ADAMIC_GOCACHE_OFF=1, succeeded: Go 0.051s, Node 0.034s, markdown dependencies 0.111s, submodules 0.130s, clang 0.417s, shared cache explicitly off 0.423s, build/cache warm 169.796s, done 169.877s. setup-local-cache.log is authoritative for all timing lines. nproc=5, cgroup cpu.max=400000 100000 (four CPUs), Go 1.27.1, clang 20.1.8, Node 24.19.0. Every later build/test shell sourced the printed /workspace/adamic-tools/env.sh. Cold attempts expired during compilation before test execution; none are claimed test passes.

Submission:

The squash was created exactly as requested from origin/main and origin/compiler/loom-records. The local squash commit is folded with this repair and its evidence into one submitted commit, whose sole parent is 199174800998904d8d646dfa5892f512cd5cb20f. A final fetch confirmed main did not move. Only compiler/loom-records-2 is pushed, without force. No PR or Loom submission is made. No cohere code is copied. The report's commit hash is supplied at handoff; embedding a commit's own SHA in its committed report would be circular.
