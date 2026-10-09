Built: option (a) for #38vbx51, replacing eight divergent record runtime stops with narrow compile-time refusals.
Commits: 8f1e2f58 adds the refusal pass and fixtures; 93f4a6bf pins exact diagnostics, preserves uncalled nested bodies and refreshes counts; evidence follows on compiler/records-next.
Checks: admission-delta passes with 23 new admissions, all agreeing with Node, JavaScript and native, none omitted; lower shards, reader guard, focused lane checks and mutants pass.
Mutants: removing only the refusal pass admits all eight divergent stops in both backends; all existing lane source, runtime, ownership, observation, readiness and narrowing mutants remain caught.
Uncovered: census reports zero exact refusal-shape hits in stock tsc, classified Want; V5 dictionary views and runtime clearance remain pending.

This report supersedes the earlier red records-next report, preserved in Git at b740bd3b. The existing branch remains based on optional-presence-next c81ba261. No other worker's history was merged. The option (a) ruling treats a runtime stop on these valid Node programs as a divergence.

The new internal/lower/records_divergence.go pass runs before ordinary record lowering. It follows prototype names through const aliases and resolved direct call arguments. It refuses an observable prototype read, inherited membership or __proto__ setter when the receiver lacks an explicit own data entry. Existing own-key guards, discarded reads, scalar comparisons and supported own entries remain admitted. Literal prototype reads keep their existing refusal rather than introducing a duplicate guard.

For scalar invalidation, the pass requires a visible earlier call whose body writes or deletes the same literal slot of the same record symbol. It excludes runtime tag observations, reference reads, different-slot writes and uncalled nested function bodies. A value snapshot or a type recheck after the call remains supported. The supported-neighbor test covers ordinary reads, membership and sets; own guards; an explicit own __proto__ entry; unrelated writes; snapshots; rechecks; and an uncalled nested mutator. It passes in 0.26 seconds. The scope is these statically identified divergent shapes; this unit adds no general record prototype model or interprocedural effect solver.

Every refusal has an exact source location, proof gap and concrete fix. The original eight .a programs are retained as refusal fixtures. Tests also compile identical source through generated .ts files and assert the complete Where, What and Fix fields. The tests pin:

| Fixture under internal/oracle/testdata/ | Location | Gap and fix |
| --- | --- | --- |
| records_prototype_read.a | 2:54 | Prototype chain is not modeled; check Object.hasOwn immediately before the own read, or use a Map. |
| records_prototype_in.a | 2:53 | Prototype membership is not modeled; use Object.hasOwn or a Map. |
| records_prototype_set.a | 2:34 | The inherited __proto__ setter is not modeled; define an explicit computed own data entry first, or use a Map. |
| records_compare_properties_left.a | 9:9 | Observable toString prototype read; own-key guard or Map. |
| records_compare_properties_right.a | 9:20 | Observable constructor prototype read; own-key guard or Map. |
| records_environment_boundary.a | 3:12 | Observable toString prototype read; own-key guard or Map. |
| records_named_invalidated.a | 9:20 | A call changed the same narrowed scalar slot; snapshot before the call or read and recheck afterward. |
| records_narrowed_number.a | 3:58 | A call deleted the same narrowed scalar slot; snapshot before the call or read and recheck afterward. |

The oracle registration changes only these eight from checked runtime fixtures to refused fixtures. Other record operations retain their existing registration and pass the focused oracle shards.

Admission proof:

The tool is cmd/adamic-admission-delta from origin/compiler/admission-delta 841e335cf791ca7196f3719256e0187301b833f2, used in its isolated checkout. The base binary is built from c81ba261. The head binary is built from the committed compiler implementation with only the tool's admission-lower init adapter supplied through a Go overlay. The dependency branch was not merged.

Command, from the repository root:

```
timeout 600 /tmp/records-next-admission-tool --base c81ba261 --head HEAD \
  --base-binary /tmp/records-next-base-adamic --base-lower-binary /tmp/records-next-base-adamic \
  --head-binary /workspace/records-next-scratch/records-ruling-adamic \
  --head-lower-binary /workspace/records-next-scratch/records-ruling-adamic \
  --manifest review/compiler/records-next/ruling-a/admission-manifest.json \
  --manifest-generator cloud/admission-corpus/manifest.py \
  --manifest-generator-revision origin/compiler/admission-delta \
  --workers 4 --compile-timeout 60s --timeout 30s --json
```

The complete corpus includes changed programs, witnesses, fixtures, named stage 1 gaps, review programs and the pinned empty fuzz corpus. There is no filter or runtime sampling budget. The result is pass: 976 accepted by both, 146 refused by both and 23 newly accepted. All 23 new admissions agree with source Node and both backends, with zero omitted, zero newly refused and no compiler failures or timeouts. The eight divergent programs are now refused by both the base and head. ruling-a/admission.json, admission-manifest.json and admission-summary.json hold the pinned result. The final evidence commit changes review files only; its compiler/runtime sources match the proved implementation commit.

Census:

The census walks untouched TypeScript v6.0.3 src/compiler at upstream commit 050880ce59e30b356b686bd3144efe24f875ebc8, extracted with git archive from the existing pinned upstream checkout. It scans all 77 .ts roots, including declaration files, and recognizes 727 record accesses. Each source SHA256 is recorded. It invokes the actual new refusal predicates at every node and records each hit's file, line, reason and fix.

The existing stage3/census/latent/make_overlay.py failed loudly because its refusal transformer does not support the chain's additional outer error returns. No production census or compiler source was changed to accommodate it. ruling-a/build-census.py uses that census's measurement-only loader approach and AST walker: checker diagnostics remain outside compilation, ordinary Load is disabled, Lower cannot produce IR, and the driver imports no backend. All overlay sources are .go.txt or .txt evidence. These observations are on a checker-rejected source tree and authorize no source admission.

The positive control runs the same walker over the eight generated .ts witnesses. It finds exactly one hit for every requested shape, with all eight expected source locations, while recognizing 13 record accesses. The untouched compiler result is:

| Refused shape | Count | File:line hits | Ruling classification |
| --- | ---:| --- | --- |
| prototype read through known-key provenance | 0 | none | Want |
| inherited record in | 0 | none | Want |
| inherited __proto__ setter | 0 | none | Want |
| comparison properties left observation | 0 | none | Want |
| comparison properties right observation | 0 | none | Want |
| environment record fallback observation | 0 | none | Want |
| named scalar invalidated by same-slot call | 0 | none | Want |
| number scalar invalidated by same-slot call | 0 | none | Want |

These counts measure the implemented known-key provenance and visible same-slot mutation shapes. The generic compareProperties helper and exotic process.env operations remain broader modeling work; their mere syntax is not an exact match to these narrow refusal predicates. There are no above-zero refusal shapes to file as modeling follow-ups in this census. ruling-a/census-summary.json records each count, every hit list, the source hashes and positive controls; census-stock.json and census-witnesses.json are the raw observations.

Checks, all with output written directly to logs:

* Setup: GOPROXY=https://proxy.golang.org|direct, ADAMIC_GOCACHE_OFF=1, timeout 120 bash cloud/setup.sh; then source /workspace/adamic-tools/env.sh. Setup succeeds. Go ready 0.017, Node ready 0.021, submodules 0.053, markdown duration 0.013 / ready 0.077, clang ready 0.152, shared cache off 0.154, build ready 38.915, deferred tests 39.069, cache warm 39.070, done 39.095 seconds. nproc=5; cpu.max=400000 100000. Go 1.27.1, Node 24.19.0, clang 20.1.8. Scratch uses the workspace to preserve /tmp capacity.
* go test ./internal/lower -run 'Record|DetachedOwn' -count=1 -timeout 85s passes. Final full lower runs in twelve shards, each with its recorded anchored test-name pattern, -count=1, -timeout 85s and external timeout 89. All twelve pass. ruling-a/lower-shards.json records every test name and exit.
* TestCallTargetReaders runs with -count=1 -timeout 85s and passes. Its final result is in readers.log.gz.
* Eight refusal test leaves pin .a and .ts; the three existing stop-pinning mutant tests also validate the refusal. All pass. Exact leaf durations are recorded in test-leaf-times.json; each new leaf is below one second in the normal run.
* The 137 focused oracle fixture paths run in eight TestNativeAgreesWithNode shards, with ADAMIC_GATE_UNCACHED=1. Four existing lane mutant groups, main TestEntriesAcceptance/Provenance/RuntimeReadiness and TestRecordCensusComparisonBuckets pass. ruling-a/oracle-shards.json records exact patterns and exits.
* go test ./internal/native -run '^TestRecordsAgainstNode$|^TestRecordReadMutants$|^TestRecordMutantsUnit' -count=1 -timeout 85s -v passes in 15.246 seconds.
* All twelve lane source overlays pass their intended failure assertions, all three named contract mutants are caught, and the nullable-slot port guard mutant is caught by TestNamedRecordRefusals. The named runner now uses Go overlays and preserves production sources during concurrent checks.
* Counts regenerate exactly once with go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 8m -args -update-counts, external timeout 600. It passes in 185.091 seconds. Independent verification without update passes, with 1,289 counted observations. Only the eight newly refused fixture rows disappear; no still-compiled row changes and no row is added. ruling-a/counts-audit.json attributes each removed row to the ruling. Refusal fixtures have no runtime allocation counts; their compile-time checks remain exercised.
* Lane checks run after committing: git fetch -q origin main devtools/fast-gate cloud/merge-tree; git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -. The committed-code run passes in 50.1 seconds: gofmt/tools on 322 Go files, t.Parallel on 23 test packages and a-check on 103 .a files. Its ten-second vet allowance was exceeded; separate go vet on lower, oracle, ir, native, fresh, flow and javascript passes with empty output. Final lane output is in ruling-a/lane-checks.log.gz. No full package suite or fast gate was run.

Mutation evidence:

| Mutation | Catcher |
| --- | --- |
| Only recordDivergenceRefusal removed, all eight witnesses | TestRecordRefusalPrototypeRead/In/Set, CompareLeft/Right, Environment, NamedInvalidated and NarrowedNumber: each compiles, source Node exits 0, both backends stop at 70 and disagree. No runtime guard is removed. |
| Existing prototype read-through mutant | TestRecordPrototypeMutant runs under refusal removal; its original JavaScript mutation finishes as Node while the admitted backend stops. |
| Existing scalar narrowing guard removal | TestRecordNarrowingMutant runs under refusal removal; unchecked output finishes while the admitted checked backend diverges. |
| Existing named member type guard removal | TestNamedRecordTypeGuardMutant runs under refusal removal; the old type-guard mutant runs as Node while the admitted backend diverges. |
| readonly-index, numeric-index, storage-views, mutable-invariance, record-cycles, spread-invariance, logical-record-conversion, nullable-record-container | TestRecordRefusals |
| prototype-literal | TestRecordPrototypeLiteralNames |
| opaque-own-argument, shorthand-own-escape, restored blanket index refusal | TestDetachedOwnRepresentation, TestDetachedOwnRefusals, TestRecordForms respectively |
| named index-read-type, unrestricted-write, erased named contract | TestNamedRecordReadTypes and TestNamedRecordRefusals |
| nullable-slot guard removed | TestNamedRecordRefusals, union slot incorrectly admitted |
| Eleven operation mutants: read/write/delete/in/hasOwn/keys/values/entries/spread/for-in/stringify | TestRecordOperationMutants, clean mutated output differs from Node |
| Lost releases | TestRecordOwnershipMutant, LeakSanitizer |
| Eager coalescing | TestRecordCoalesceMutant, Node stdout |
| Removed discarded/snapshot/scalar own-read guards | TestRecordObservationGuardMutants, mutant stop differs from Node |
| Inherited detached membership | TestDetachedOwnInheritedMutant, Node stdout |
| Detached alias and named alias readiness removed | TestDetachedOwnReadinessMutant and TestNamedRecordAliasReadinessMutant, Node ReferenceError versus clean mutant |
| Partial/named missing entries inserted as present undefined | TestPartialRecordAbsentEntryMutant and TestNamedRecordAbsentEntryMutant, both backend stdout |
| Integer order, uint32 maximum, deleted iteration | TestRecordMutantsUnit00/01/02, Node stdout |
| Overwritten key leak, stored key freed, own slot silently null | TestRecordMutantsUnit03/04/05, LeakSanitizer, AddressSanitizer and own-slot/UBSan check |
| Inherited membership, missing read silently null, own hit treated as missing | TestRecordReadMutants, required own-read behavior and diagnostic |

The refusal-removal runner sets ADAMIC_RECORD_REFUSAL_MUTANT=1 only for the overlay test command. This changes the test's expectation to require an admitted divergence; it is not a production bypass. Ordinary tests require the exact refusal. The overlay source and each original lane mutant source are stored as noncompilable evidence.

recordElement, finitePartialRecordElement, recordSlot and ir.RecordCall remain on this head. The view 142 ruling remains in force; this turn introduces no partial dictionary view support. Runtime files changed cumulatively by this lane, requiring @system_adamic_runtime clearance, are exactly internal/native/runtime/json_stringify.c and internal/native/runtime/json_stringify.h. Neither runtime file changed in this ruling turn. Clearance was not granted here and no message was sent to the owner.
