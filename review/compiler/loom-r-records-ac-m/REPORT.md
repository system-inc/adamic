Re-cut records onto current main, preserving main's newer JSON preparation and ownership infrastructure.
Commit: one commit on compiler/loom-r-records-ac-m; exact SHA is reported at handoff.
Commands/results: see final package-results and per-test logs below; Linux clang 20.1.8 counts regenerated once.
Mutants: real source-Node, allocation balance, sanitizers, refusal guards and unchanged a-check predicate provide the catchers below.
Not covered: Loom/ahra, full gate, PR creation, and unrequested performance or runtime feature expansion.

Base: 5a94465370ea0172e25aae8c13478e59e40c2bb4, newest origin/main at fetch. Source: e47a602d81b9b0533eb8d85df3df5f45e328cf25, whose parent is dbdf977f5b8b0979d414ebe6f20e7aadd312e309. Applied exactly that commit's own patch with cherry-pick --no-commit, corroborated with merge-tree --merge-base=<source>~1. No main-to-source diff was used. Branch naming preserves the complete old basename, including -ac. No cohere implementation was copied.

Conflicts and resolutions:

- internal/native/runtime/json_stringify.c: source's record container arm met main's rewritten typed-runtime/toJSON encoder. Retained main's preparation, transferred-result ownership, throws, ancestor cycle checks and borrowed object-slot view. Added the record arm to write_prepared_body; prepare each child once before omission/key output, release owned results and key snapshots on throws, and include records in reference ownership and cycle tracking. The nullable reference case from source applies unchanged. No main behavior was reverted. Record-in-union admission stays at the source compiler's conservative boundary; this integration does not expand union metadata support.
- internal/native/runtime/json_stringify.h: both edits extended the enum after object. Retained main's toJSON position and appended record, preserving main's enum values.
- internal/oracle/counts.md: discarded the conflicted table, restored the entire main table, regenerated once with house clang 20.1.8 on Linux. The audit below contains only source-owned rows.

RUNTIME-CLEARANCE.md lists every exact signed C/header line beyond the original commit's own lines. json_stringify.c-resolution.patch and json_stringify.h-resolution.patch give the complete ordered zero-context patches against main. Those are the only runtime production files changed. Main's newer changes remain; additional lines integrate records with those contracts.

Integration assertions discovered by the full oracle package run:

- method_values_mutant_test.go: record specialization removed the original first true HasOwn query, so replacing the first remaining query with false changed nothing. The same behavioral mutant now inverts the query, preserving its intended Node-only catcher regardless of the original answer. It passes with clean sanitizers and balanced ownership.
- entries_acceptance_test.go: scanner entries now uses homogeneous RecordCall rather than open-view ObjectCall. Accept RecordCall only for the proven-number contract; retain all ObjectCall.Checked assertions and all baseline/mutant runtime comparisons. The pinned acceptance corpus and expected Node observations were not edited.
- element_access_boundaries_test.go: pin the earlier exact unsupported representation diagnostics for array-derived and numeric-record shapes. String dictionaries now lower as intended; hold that witness to Node, sanitized/counted native, release C and JavaScript. No unsupported shape was admitted and no test was deleted.

The initial full oracle failure and the focused corrective run remain in their logs. Changes are confined to the affected oracle tests. The full oracle suite was repeated because those assertions changed; production lowering and the counted fixtures did not change, so counts were not regenerated a second time.

Counts changed rows (all under internal/oracle/testdata; columns allocations/frees/retains/releases/peak/in-regions; exact full lines in counts-audit.json):

| Fixture | Before | Regenerated |
|---|---|---|
| detached_own_objects.a | new | 6/6/4/8/5/0 |
| detached_own_parser.a | new | 4/4/0/4/2/0 |
| detached_own_records.a | new | 2/2/3/4/2/0 |
| detached_own_unready.a | new | 0/0/0/0/0/0 |
| json_replacer_duplicates_object.a | new | 12/12/12/18/7/0 |
| library_method_values.a | 137/137/154/295/42/0 | 136/136/152/291/41/0 |
| library_string_raw.a | 45/45/55/89/11/0 | 45/45/51/85/11/0 |
| method_coverage_object_descriptors.a | 47/47/88/127/7/0 | 46/46/61/98/6/0 |
| records_census.a | new | 18/18/38/47/9/0 |
| records_compare_missing_scalar.a | new | 63/63/75/87/9/0 |
| records_discarded.a | new | 4/4/4/10/4/0 |
| records_for_in.a | new | 6/6/28/24/3/0 |
| records_guarded_snapshot.a | new | 10/10/15/27/6/0 |
| records_json_replacer_duplicates.a | new | 12/12/22/24/8/0 |
| records_named_alias_unready.a | new | 0/0/0/0/0/0 |
| records_named_intrinsics.a | new | 14/14/13/28/7/0 |
| records_named_operations.a | new | 19/19/42/48/7/0 |
| records_named_options.a | new | 8/8/1/6/4/0 |
| records_named_union.a | new | 14/14/30/37/5/0 |
| records_narrowed_reference.a | new | 3/1/7/7/3/0 |
| records_operations.a | new | 64/64/136/98/21/0 |
| records_optional.a | new | 12/12/11/15/6/0 |
| records_ownership.a | new | 37/37/62/60/18/0 |
| records_partial_parser.a | new | 5/5/0/5/5/0 |
| records_partial_undefined.a | new | 15/15/13/23/8/0 |
| records_partial_views.a | new | 20/20/54/44/6/0 |
| records_scalars.a | new | 23/23/28/31/8/0 |
| records_type_only_composition.a | new | 0/0/0/0/0/0 |
| records_type_only_scanner.a | new | 0/0/0/0/0/0 |
| records_type_only_value.a | new | 5/5/7/12/4/0 |

Exactly 30 rows: 27 added and three existing changed. No row was removed and no unrelated row changed. Count inventory matches the original source patch exactly. Values match source for 29 rows. The sole value difference from the source table is records_census.a: source 16/16/38/45/9/0, regenerated 18/18/38/47/9/0. Observation: two extra allocations, frees and releases, unchanged retains/peak/live. Main’s newer array JSON loop allocates and releases numeric key strings for each prepared element; this fixture serializes a two-element array. That code explains the increase and was retained. No unrelated count row changed.

Toolchain and setup:

GOPROXY=https://proxy.golang.org|direct; ADAMIC_GOCACHE_OFF=1 timeout 600 bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. Setup succeeded in 41.740 seconds. nproc=5; cpu.max=400000 100000, so tests used GOMAXPROCS=4 and -parallel 4. Go 1.27.1, Node 24.19.0, Linux clang 20.1.8. Timing lines: submodules 0.061s; Node 0.079s; Go 0.086s; markdown dependencies 0.185s; clang 0.421s; shared cache off 0.424s; Go build 41.528s; deferred test binaries 41.705s; build cache warm 41.706s; done 41.740s. setup.log contains the exact lines. The unchanged checker archive was prepared as setup with a separately bounded c-archive build. No test exceeded its budget and was then rerun as a cold-build exception.

One runtime-edit auto-review request timed out without running. Retrying the same authorized edit succeeded; no permission remains blocked. The first-green push target of 30 minutes was missed: environment startup, required full-package validation, integration fixes and disk-space recovery took longer. No failing compiler state was published.

Validation commands use direct log files, never output pipelines. Package command arrays and process exits are in full-results.json; every completed test and elapsed time are in the matching *-tests.json. Go vet and gofmt were run again after the oracle assertion fixes. The full gate and ahra were not run.

Full package validation (GOMAXPROCS=4, -parallel 4, -p 1, -count=1, -json):

| Command | Result | Command seconds | Passed checks | Skipped |
|---|---|---:|---:|---:|
| go test ./internal/lower/... -count=1 -json -timeout 840s -parallel 4 -p 1 | PASS | 50.053 | 1577 | 2 |
| go test ./internal/native/... -count=1 -json -timeout 840s -parallel 4 -p 1 | PASS | 280.107 | 770 | 92 |
| go test ./internal/flow/... -count=1 -json -timeout 840s -parallel 4 -p 1 | PASS | 62.057 | 889 | 0 |
| go test ./internal/fresh/... -count=1 -json -timeout 840s -parallel 4 -p 1 | PASS | 14.022 | 928 | 0 |
| go test ./internal/ir/... -count=1 -json -timeout 240s -parallel 4 -p 1 | PASS | 4.012 | 16 | 0 |
| go test ./internal/javascript/... -count=1 -json -timeout 240s -parallel 4 -p 1 | PASS | 2.001 | 7 | 0 |
| go test ./stage3/fixtures/... -count=1 -json -timeout 840s -parallel 4 -p 1 | PASS | 86.047 | 722 | 1 |

The first full oracle run (392.321 command seconds) found the three integration assertion issues described above. The focused corrective run passed all 27 checks, longest 4.49s. The repeated full oracle command, go test ./internal/oracle/... -count=1 -json -timeout 1200s -parallel 4 -p 1, ran 373.438 package seconds: 4,867 checks passed, 291 skipped, two leaves (and their parent reports) failed only because cache result writes returned "no space left on device". No integration failure recurred. Both leaves were retried after freeing the inactive gocacheprog cache and passed: TestCheckedViewInterfaces/interfaces-wrong-inherited and TestCheckedViewV2ReadsAfterWrites/fixed-tuple-good (3.539 package seconds, maximum 3.53s). A complete command did return exit 1 for that infrastructure error; this report does not relabel its exit as a pass. Combined full-run and focused-retry evidence covers all 4,871 oracle checks successfully. full-oracle-final.jsonl.gz and oracle-disk-retest.jsonl.gz retain the exact outputs. The root had zero free space; cache cleanup restored 7 GB without removing source files or active native products. cache-cleanup.json records the recovery.

Counts regeneration: timeout 660 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout 600s -parallel 4 -args -update-counts, PASS 73.827 package seconds; aggregate time exceeds a leaf budget, no leaf does. counts-update.log retains each leaf.

Explicit guards, all PASS:

- internal/ir TestCallTargetReaders and TestEveryMapAProgramHoldsIsSorted, -timeout 90s: maximum 2.30s. Record element metadata uses SortedMap; no exception added.
- internal/lower TestNativeAgreementRowsNameRegisteredPrograms, included in full lower run: 0.02s; literal registration requirements retained.
- internal/oracle TestRuntimeSetsAreEachUsed, included in full oracle run: 1.10s; existing runtime feature sets retain exact witnesses, no new runtime feature set was introduced.
- cmd/adamic-gate TestEveryTestIsParallelOrSaysWhy and TestEveryDuplicatedProductDeclarationAgrees, -timeout 90s: maximum 1.62s. All five new tests call t.Parallel first. No product recipe, product directory identity or new Go package was introduced.
- go vet ./internal/flow/... ./internal/fresh/... ./internal/ir/... ./internal/javascript/... ./internal/lower/... ./internal/native/... ./internal/oracle/... ./stage3/fixtures/..., timeout 180s, exit 0 after the integration fixes.

Five new native runtime test timings from the full native package:

- TestJSONRecordsRuntimeThrowCleanupMutant: 0.95s, pass.
- TestJSONRecordsRuntimeCycleGuard: 1.55s, pass.
- TestJSONRecordsRuntimeOwnershipMutant: 0.44s, pass.
- TestJSONRecordsRuntimePrepareMutant: 0.42s, pass.
- TestJSONRecordsRuntimePreparedValues: 0.4s, pass.

All completed positive-run leaves stayed below 60s. The longest reported positive test was 33.39s (full native run); oracle final maximum was 24.29s, full lower maximum 14.40s. No leaf was deleted or split because none exceeded its budget. The full package results retain optional skips; they are not reported as executions. No over-budget test was rerun.

Compiler guard mutant results (each process exits 1 and its intended assertion text must be present):

| Mutant | Catcher | Assertion witness | Command seconds |
|---|---|---|---:|
| readonly-index | TestRecordRefusals | want "signature" | 8.26 |
| numeric-index | TestRecordRefusals | want "signature" | 7.764 |
| prototype-literal | TestRecordPrototypeLiteralNames | want literal member refusal | 7.653 |
| storage-views | TestRecordRefusals | want "seen as" | 7.505 |
| mutable-invariance | TestRecordRefusals | want "invariant-mutable" | 7.607 |
| record-cycles | TestRecordRefusals | want "cycle" | 7.498 |
| spread-invariance | TestRecordRefusals | want "invariant-mutable" | 7.552 |
| logical-record-conversion | TestRecordRefusals | want "logical record operand" | 7.411 |
| nullable-record-container | TestRecordRefusals | want "value of type" | 7.211 |
| opaque-own-argument | TestDetachedOwnRepresentation | must not compile | 7.315 |
| shorthand-own-escape | TestDetachedOwnRefusals | want detached-method refusal | 7.56 |
| type-only-index-admission | TestRecordForms | refuses | 6.994 |
| index-read-type | TestNamedRecordReadTypes | named read used index type | 6.845 |
| unrestricted-write | TestNamedRecordRefusals | want NotYet named property named | 7.103 |
| erase-named-contract | TestNamedRecordRefusals | want NotYet seen as | 7.295 |
| nullable-slot | TestNamedRecordRefusals | slot representation | 7.261 |

The nullable-container mutation uses main's explicit `if of == ir.Record` metadata stop instead of the old pre-main expression anchor. All mutations are freshly generated from current source; mutant Go sources are .go.txt files and never compilable package sources. The three named-member overlays replace the old script's worktree mutation with equivalent overlays.

Header and archived readers:

- The unchanged pinned Gate.aCheck accepts every one of the 110 changed .a files: 64 checked, four explicit NotYet, nine refused and 33 checker errors, all with the measured header. This includes the one added Node source witness for the direct runtime C API regression; its unsupported compiler lowering is explicitly recorded, not claimed as compiled Adamic. No file is excluded from this a-check inventory.
- Removed-header and wrong-refusal-reason mutants are rejected by that unchanged predicate while the actual compiler still reports Refused. The old-line-pin overlay fails the real oracle refusal diagnostic at 3:54 versus its planted 2:54 expectation (0.10s leaf).
- Archived stage3 position readers pass all 11 selected source-edited cases (2.713 package seconds); the archived intentionally incorrect 07_optional_view classification remains a mutant, not a positive case.
- The unchanged stock TypeScript 6.0.3 input reader reproduces all 30 exact observations and catches replacing lost own __proto__ with an ordinary constructor path through actual CLI output. No archived source body or golden observation was rewritten.

Additional recorded mutants:

| Mutation | Catcher | Result | Command seconds |
|---|---|---|---:|
| refusal-removal | eight refusal witnesses plus prototype/narrowing/named-guard mutants | Node finishes; removal admits native and JavaScript checked exit 70, then runtime mutations distinguish the guard | 14.554 |
| flow-refused-directory | TestFlowCorpusRemainder | corpus rejects deliberate records_compare_properties_left refusal | 16.795 |
| counts-mutant | TestCountsAreRecorded shards | planted records_operations counts mismatch fails | 6.394 |
| stage3-classification | TestFixturesRecords | 07_optional_view incorrectly Compiles expectation fails | 6.115 |

New runtime integration mutations, all executed in the full native run: skipped preparation changes only the observable source-Node output with clean sanitizers and balanced counts; omitted record reference ownership preserves stdout but fails balance/LeakSanitizer; omitted key-snapshot release on throw preserves stdout but leaks; removed cycle ancestry guard loses the exact cyclic-container refusal. The unmutated C control matches real Node and has balanced counts and clean ASan/UBSan/LeakSanitizer.

Source-owned embedded behavioral mutants were also exercised by the full native/oracle runs and the refusal-removal overlay. Their exact cases and output are preserved in the logs:

- JSON replacer dedup removal, separately on record and object receivers: Node stdout catches the change, with clean sanitizers/leaks.
- Detached own inherited-name and readiness mutations: Node stdout or initialization stop catches the change.
- Named-record type guard, alias readiness and absent-as-present mutations: checked exit 70, normalized ReferenceError, or Node stdout catches the change.
- Observation guard removal on discarded, guarded_snapshot and compare_missing_scalar: native missing-own-member exit 70 differs from successful Node.
- Partial-record absent-as-present mutation: Node stdout catches it in native and JavaScript.
- Record operation mutations: Node stdout with clean native exit and sanitizers; lost releases: LeakSanitizer; eager/repeated coalescing fallback: Node stdout; prototype bypass and stale narrowing: source-Node versus checked-stop comparison. The extra refusal-removal run ensures the refusal-covered mutants actually reach their intended runtime witnesses.
- The adapted pre-existing own-field mutant is also held to Node stdout, clean exit, sanitizers and leak checks; it can no longer silently change an already-false query to false.

Current-main reader coverage:

The fresh go-list import inventory finds 42 direct compiler reader packages; all were checked, along with 11 additional external readers declared in cloud/fast-gate/compiler-dependencies.json. All 53 jobs exited 0; no completed reader leaf reached 60 seconds (maximum 18.65s). consumer-selection.json gives each exact selector, consumer-results.json each command, exit and leaf time, and consumer-inventory.json the current imports and previous selections. Packages without selected tests were compile-only, not represented as runtime validation. Optional skips, including the external cycle ledger, remain explicit. Main replaced TestCompileProfiles in four lint-rule packages with three top-level leaves per package; all 12 replacements were run and passed. No consumer test was edited.

The compiler dependency census guards TestEveryCCompilerRunsThroughNative, TestDependenciesFollowTheMarkers and TestDependenciesCoverTheDeclaredMap passed in internal/compilercensus. No new compiler-consuming package was created, so no declaration change was needed. The reader coverage uses focused checks rather than running full packages outside the commit's touched packages. The broad stage3/lane Python wrappers were not repeated after the source report's deadline limitations; its Go manifest reader passed. This limitation does not apply to the full stage3/fixtures package, which was run and passed.

Formatting: gofmt -l on all 54 changed Go files, no output, exit 0 (gofmt-final.json). git diff --check is clean for production files. Runtime patch audit confirms the attached patches still match the final code exactly. No runtime source changed during overlay mutations. All refused top-level record witnesses remain under records_refused/, never in flow's lowering glob.

Exact additional runtime lines for clearance follow; these reproduce RUNTIME-CLEARANCE.md:

# Runtime clearance

These are the exact signed runtime lines in this re-cut that are beyond e47a602d81's own added/removed lines. Minus lines are current main lines extended by the integration; plus lines are final code. Common source-patch lines are excluded, with repeated lines counted by occurrence. The complete ordered patches are json_stringify.c-resolution.patch and json_stringify.h-resolution.patch.

Main's typed providers, toJSON, throw handling, cycle ancestors, nullable metadata and borrowed object slot views are retained. The record enum is appended after toJSON, preserving main's existing enum values. The record path prepares each field exactly once before testing omission or writing its key, releases transferred reference results, releases its key snapshot on throws, and participates in main's cycle guard. No main behavior is reverted. Node, balanced counts, ASan/UBSan/LeakSanitizer and four independent runtime mutants cover the new interactions.

### internal/native/runtime/json_stringify.c

```diff
-    return schema->kind == adamic_json_string || schema->kind == adamic_json_array || schema->kind == adamic_json_object || schema->kind == adamic_json_toJSON || schema->kind == adamic_json_union || schema->kind == adamic_json_maybe_boolean || schema->kind == adamic_json_map || schema->kind == adamic_json_function;
+    return schema->kind == adamic_json_string || schema->kind == adamic_json_array || schema->kind == adamic_json_object || schema->kind == adamic_json_toJSON || schema->kind == adamic_json_union || schema->kind == adamic_json_maybe_boolean || schema->kind == adamic_json_map || schema->kind == adamic_json_function || schema->kind == adamic_json_record;
-    bool container = s.kind == adamic_json_array || s.kind == adamic_json_tuple || s.kind == adamic_json_object || s.kind == adamic_json_toJSON;
+    bool container = s.kind == adamic_json_array || s.kind == adamic_json_tuple || s.kind == adamic_json_object || s.kind == adamic_json_toJSON || s.kind == adamic_json_record;
+		const adamic_record *record = s.value.reference;
+			bool owned;
+			adamic_json_result result = prepare(w, *slot, schema->element, key, &owned);
+			if (adamic_thrown != NULL) {
+				if (owned && reference_result(result.schema)) { adamic_release(result.value.reference); }
+				adamic_release(ordered); return false;
+			}
+			json_scalar item = scalar(result.value, result.schema, w->runtime);
+			if (item.kind != adamic_json_undefined && item.kind != adamic_json_function) {
+				if (written++ != 0) { ascii(w, ","); }
+				indent(w, depth + 1); quote(w, key); ascii(w, w->pretty ? ": " : ":");
+				(void)write_prepared(w, result.value, result.schema, depth + 1);
+			}
+			if (owned && reference_result(result.schema)) { adamic_release(result.value.reference); }
+			if (adamic_thrown != NULL) { adamic_release(ordered); return false; }
```

### internal/native/runtime/json_stringify.h

```diff
-	adamic_json_tuple, adamic_json_object, adamic_json_toJSON
+	adamic_json_tuple, adamic_json_object, adamic_json_toJSON, adamic_json_record
```

During development of the new runtime harness, two focused runs failed harness assertions: the leak classifier initially mistook LeakSanitizer's AddressSanitizer summary for a distinct memory error, and LeakSanitizer exited before buffered stdout was flushed. The harness now distinguishes actual "ERROR: AddressSanitizer:" errors and flushes the C witness's stdout before returning. The production encoder was unchanged by those test-only fixes. Final focused and full-native runs pass. runtime-merge-initial.log and runtime-merge-stdout-attempt.log preserve those observations.

Evidence archives: completed JSON-lines test logs are gzip-compressed without editing. evidence-archives.json records each original byte length and SHA-256 and the compressed archive hash. Scripts still write the original .jsonl names when rerun; the committed observations append .gz. Small summary JSON files and readable focused .log files remain uncompressed.
