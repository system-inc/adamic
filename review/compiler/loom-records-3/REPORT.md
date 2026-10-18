Re-cut records as one commit on current main for compiler/loom-records-3; push authorized after review of the cold consumer product build.
Commit: branch HEAD supplied at handoff; base 4011e74a8f59e40ea533f1e0613a2a0aa6b4997b; source 3bebe77258da2f80bc231a67c5e1e85d30869632.
Commands/results: focused lowering, Node/native oracle, runtime mutants, flow, stage3, guards, counts regeneration, readers and vet pass; exact commands and times are in validation.json and the runner scripts.
Mutants: all 16 lowering guard mutations, eight refusal-removal witnesses, runtime/oracle mutations, flow and stage3 classifications and the planted counts mismatch were caught; details below.
Not covered: remaining consumer selections after typeaware, optional estree library and cycle-ledger inputs, full packages/gate, a new admission census, benchmarks and non-Linux platforms.

The requested fetch, checkout from origin/main and squash merge succeeded without conflicts. No conflict markers or manual compiler resolutions were needed. The automatic merges in internal/lower/expression.go and internal/native/emit_values.go preserve main's newer Node buffer handling and typed-array representations alongside records lowering. Main's runtime header/object changes remain. There are no compiler or runtime edits beyond the source patch and main's automatic merge.

Counts assumption: current main supersedes the prompt's digest-column description. Commit 1f07ce2d moved measurement identity into products and removed that column. I kept current main's format and harness rather than restoring an obsolete schema. I took origin/main's counts.md and completed one Linux -update-counts regeneration, clang 20.1.8. It passed in 57.679s. The earlier interrupted attempt never reached a test and changed no table. TestCountsAreRecordedTable passes afterward.

Exactly 30 rows differ: 27 new rows and three existing rows (library_string_raw.a, library_method_values.a and method_coverage_object_descriptors.a). Every row belongs to the source change's recorded coverage inventory; zero unrelated changes and zero removed rows. counts-audit.json contains every path, complete before/after rows and the empty unrelated list. No Mac table was used.

internal/native/runtime added/removed lines are identical to the source change, in order: 23 lines, no differences. runtime-equivalence.json records the complete comparison against the source's merge base 199174800998904d8d646dfa5892f512cd5cb20f.

Guards and test grain:

- TestEveryMapAProgramHoldsIsSorted passes. Records introduces no plain map held by an IR Program; main's SortedMap fields remain.
- TestRuntimeSetsAreEachUsed passes. No new runtime feature set was needed.
- TestNativeAgreementRowsNameRegisteredPrograms passes. This source adds no native-agreement call needing registration.
- TestEveryTestIsParallelOrSaysWhy passes. All added tests already call t.Parallel first; no baseline relaxation was made.
- TestEveryDuplicatedProductDeclarationAgrees passes (1,041 declarations checked). No product declaration or recipe was added or changed; main's placeless infrastructure remains.
- No new Go package consumes the compiler, so cloud/fast-gate/compiler-dependencies.json needs no addition. The current-main registry, including cycleledger, is preserved. consumers.json records the 51 affected consumer packages.
- All 43 tests in changed test files were observed passing. test-leaf-seconds.json lists every test and child duration. The longest observed leaf in focused validation was TestCallTargetReaders, 18.96s; no added/touched test needed splitting. Tests ran on the four-CPU cgroup with GOMAXPROCS=4 and -parallel 4 for final focused runs.
- TestCallTargetReaders passed with -timeout 90s. Vet passed on flow, fresh, ir, javascript, lower, native and oracle. git diff --check passes.

The fixture selector ran the source change's 28 registered fixtures, including its deliberate refusal control. Runnable fixtures agree with source Node and backend Node, and native under ASan/UBSan and Linux leak checks. Focused lower tests also hold entries provenance and optional-indexing boundaries. Flow's remainder passes with the exact eight refusal controls classified; stage3's 19 records status pins pass unchanged.

Every recorded mutant and catcher:

- readonly-index, numeric-index, storage-views, mutable-invariance, record-cycles, spread-invariance, logical-record-conversion and nullable-record-container: TestRecordRefusals, intended missing refusal.
- prototype-literal: TestRecordPrototypeLiteralNames; opaque-own-argument: TestDetachedOwnRepresentation; shorthand-own-escape: TestDetachedOwnRefusals; type-only-index-admission: TestRecordForms.
- index-read-type: TestNamedRecordReadTypes; unrestricted-write, erase-named-contract and nullable-slot: TestNamedRecordRefusals.
- recordDivergenceRefusal removed: prototype_read, prototype_in, prototype_set, compare_properties_left, compare_properties_right, environment_boundary, named_invalidated and narrowed_number each admit a checked stop in both compiled backends while source Node finishes. The same overlay runs prototype read-through, scalar narrowing and named-member-kind guard mutations, caught by their exit pins.
- Runtime indices-in-insertion-order, uint32-max-as-index and deleted-key-iterated: Node behavior comparisons. overwrite-key-leaked: LeakSanitizer; stored-key-freed: ASan; own-slot-null-read: UBSan. prototype-membership-restored and missing-read-silent: diagnostic/exit pins; own-read-checked-as-missing: own-hit success pin.
- Duplicate replacer dedup skipped, separately for records and fixed objects: Node stdout, clean sanitizers and leaks.
- Oracle read, write, delete, in, has_own, keys, values, entries, spread, for_in and stringify mutations: Node stdout with clean native exit and sanitizers.
- Discarded-read, guarded-snapshot and scalar-comparison guards removed: native exit 70 versus Node finishing, with exact missing-member message.
- Record ownership releases removed: LeakSanitizer alone; eager/repeated coalescing fallback: Node stdout.
- Partial and named absent entries defined as own undefined: Node stdout in both backends. Named alias-copy readiness removed: normalized ReferenceError exit 70 in both backends.
- Detached own-property inherited-name mutations, on records and objects: Node stdout. Intrinsic readiness removed: Node's initialization stop.
- Flow refusal classification removed: TestFlowCorpusRemainder fails on the eight refusal controls.
- Stage3 07_optional_view.a falsely marked Compiles: TestFixturesRecords fails its stage0 pin.
- records_operations.a recorded allocation increased by one through main's family probe: the counts family fails the recorded/measured comparison.

Source/named overlays are freshly generated against this tree, under source-mutants/ and named-mutants/. Additional overlays and logs are in this directory. No production file was changed to run a mutant. Guard mutants exit 1 at the intended assertion; runtime mutation harnesses pass only after observing the intended bad behavior. No clang warning or build failure counts as a kill. mutants.json and extra-mutants.json record commands, times and results; native-final.log and oracle-checks-final.log contain the runtime/oracle witnesses.

Toolchain/setup observations:

nproc=5, cgroup cpu.max=400000 100000, Linux, Go 1.27.1, clang 20.1.8, Node 24.19.0. The printed /workspace/adamic-tools/env.sh was sourced for builds and tests. GOPROXY was https://proxy.golang.org|direct.

Initial setup and cold parallel build attempts were stopped with exit 143 before any test began. Initial setup printed Go 0.062s, Node 0.057s, clang 0.486s, markdown 0.884s, submodules 157.501s and shared cache ready 162.936s. An earlier IR attempt exited 1 because the TypeScript submodule's go.mod did not yet exist. These are not test passes. I switched to the local Go cache; the shared-cache delay is an inference, not a proven service diagnosis.

The successful ADAMIC_GOCACHE_OFF=1 setup printed Node 0.115s, Go 0.117s, submodules 0.160s, markdown 0.161s, clang 0.241s, shared cache explicitly off 0.243s, go build 145.132s, build cache warm 145.233s and done 145.257s. setup-local.log is authoritative. Test binaries were deferred. The initial logs are retained too.

Consumer validation uses focused gap/compiler tests in each package with tests, and compile-only checks for packages without tests. consumer-selection.json and consumer-results.json list exact coverage and timing; consumer-*.log retains full output. This does not claim every test in those packages ran. No full package or full gate was run. No PR or Loom submission was made, and no cohere code was copied.

Consumer setup correction: the first estree selection failed before comparison because the estree-go-oracle-v1 owner was missing. TestProduct_EstreeGoOracle in stage1/cohere/estree/oracles settled it in 11.81s, and the selection then passed (4.940s test time). The initial failure and resumed checks are retained separately in consumer-results.json and consumer-resume-results.json. TestRecoveryLibraryGaps, TestPinnedNumericGaps, TestTypeMemberLibraryGap and TestCookedSurrogateLibraryGap skipped because ADAMIC_ESTREE_LIBRARY is not configured; no claim is made for those optional library comparisons.

A formatting-helper attempt omitted the setup environment and could not find gofmt; its corrected check, with env.sh sourced, found no file needing formatting. The first consumer-resume helper had an incorrect results filename and exited before running tests; that filename was corrected. Neither helper failure is counted as validation.

Observed local limitation: TestInspectRequestRefusals in stage1/cohere/typeaware timed out at 90.067s (command 95.939s) while building the checker C archive cold. The last completed phase was stage0 at 1.961567s; no compiler disagreement was observed. Records changes no product recipe for this archive. The user clarified that Loom serves this product from cache, with the same pattern as bridge/tsgo products (about 26s p90 on Loom versus about 80s cold here), and explicitly directed the push. Those Loom timings are supplied context, not measurements from this worker. This timeout is recorded as an observed cold product build, not attributed to records. The test was not rerun and another package's tests were not changed. Remaining consumer selections were not run. TestOriginalCycleLedger skipped because ADAMIC_CYCLE_LEDGER_ROOT is absent; consumer-skips.json preserves skip events.

Final rebase: fetched main and rebased once, without conflicts, from d79302e50c022c997c00f7c68ddb7475f9ee65fc onto 4011e74a8f59e40ea533f1e0613a2a0aa6b4997b as explicitly requested. Main added command runtime product tests and census tooling/evidence; it changed no compiler production code or counts table. The validation results above were collected before this rebase. The records patch, its 30 counts rows and runtime equivalence are unchanged. No over-budget test was rerun. The unpushed commit was amended to keep the requested single-commit branch and this report accurate.
