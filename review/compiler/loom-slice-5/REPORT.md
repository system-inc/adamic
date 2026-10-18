Re-cut slice 5 preserves placeholder lowering, Outcome 24 parser construction/speculation/stack machinery, per-backend stops and the eight-witness ledger.
Delivery is one squash commit on main; source 1e380ea12c93797e1552b57bc74b258e15d4d4b7; final base c0fd8945b872e57a341ccee5973e3f4d1b75c34b; the final delivery message gives its commit SHA.
Focused tests and 55 fixtures pass; one successful counts rewrite takes 59.061 s; TestCallTargetReaders passes in 6.04 s; changed-package vet passes.
All 27 independent overlays were rejected; embedded flow, payload, rewind, runtime and eight retained-stop repair mutants were also caught; catchers are below.
Not covered: the full gate, a fresh full admission census, the complete adapted parser, Program-region lifetime, optional guard-cost measurements, WASI or non-Linux systems.

This report supersedes the source branch's historical evidence for this re-cut. The old review tree is preserved because the requested squash imports the source's whole net change. No cohere code was copied, no other worker branch was merged, no PR was opened, and no Loom submission was made.

The first fetch and branch creation used main cd7880cf5145b33a5bf6bb2227e216d7fb9d63c0. Main then advanced to c0fd8945b872e57a341ccee5973e3f4d1b75c34b. Its intervening commits add only stage3/census/tsc-closure/rerun-2026-10-09 reports, evidence and Python/shell reproduction scripts. No compiler, runtime, Go tests, .a fixtures, module files or dependency manifest changed. The delivered branch was rebased once onto that newer tip, with no rebase conflicts. A diff against the pre-rebase commit confirms no change under cmd, internal, oracle or cloud/fast-gate. The rebase receipt is in rebase.log. The initial 30-minute push target was missed during cold setup and the provider-capacity interruption; no partial state was pushed.

The squash had exactly two conflict hunks, both in internal/ir/ir.go. For ViewContractTypes, retain main's SortedMap[int, ViewContractID] and add the source's PlaceholderSources, PlaceholderViews and PlaceholderChecks. For UninitializedFields, retain main's SortedMap[string, bool] and add FactoryFields. The three added maps also use SortedMap[string, string], so serialized IR remains deterministic. No allowlist entry was added to TestCallTargetReaders.

Main's additionalFixtureCounts is now []countCase rather than a list of functions returning many rows. Parser fixtures register individual countCase entries and have one top-level count test each. Main explicitly rejects the requested anchored ^TestCountsAreRecorded$ selector because that test is now a family. The bounded update command therefore uses unanchored TestCountsAreRecorded. There was exactly one successful rewrite. The first update attempt failed in 97.605 s because @types/node 25.3.3 was missing; TestMain refused to write. npm ci --prefix stage3/api installed the pinned declarations, then the successful attempt wrote the table in 59.061 s. No fixture was changed to hide that prerequisite failure.

The conservative reading of the 60 s law includes helper setup. Initial TestUnlistedBackendAgreement (60.13 s), TestRuledBackendOutcomes (60.40 s) and TestTerminalStackStop (60.51 s) passed their assertions but are recorded red and were never rerun. Their replacements split source/native/JavaScript/marked-stack work into independent top-level tests. Dedicated slice observations call Node and native directly rather than initializing the shared oracle result cache, whose setup builds three runtime variants. Readiness and view-field tables were also split into top-level functions. The largest final substantive slice leaf is TestParserConstructionRuntimeMutantDeclarationOrder, 12.87 s. A short displayed-name helper failure on bundled:/// library URIs was corrected before resuming the affected shard; its failure events remain recorded. Every observed leaf's seconds, including those failures and removed red leaves, are in [TEST-SECONDS.md](TEST-SECONDS.md) and [test-seconds.json](test-seconds.json).

Products keep main's displayed-source convention. Parser source tests use LoadDisplayed with repository-relative names, preserving bundled URI names. Relocation tests lower identical sources in two physical directories and compare emitted C/JavaScript plus placeholder provenance and check metadata. Independent mutants bypassing the displayed loader's names fail each of the three relocation checks. Runtime construction mutants use main's source feature flags, keyed runtime cache, link flags and file-prefix mappings. No new Go package was added, so cloud/fast-gate/compiler-dependencies.json needs no entry.

The ledger's eight top-level TestSlice5Witness* tests validate exact source hashes, declared witness paths, exact exit-70 diagnostics, repair hashes and the rule that repaired sources are not witness keys. Each repair runs against raw source Node, JavaScript and native ASan/UBSan, with LeakSanitizer on successful native runs. Every repair agrees with Node. Each test deliberately compares the repaired Node observation with retained stopping backend artifacts and rejects them. These are finite witness proofs; they do not reproduce the source branch's full external admission-delta census.

Commands run with the generated env.sh sourced, GOPROXY=https://proxy.golang.org|direct, GOMAXPROCS=4 and Go package concurrency 4. Initial focused commands set GOFLAGS=-p=4; final recorder, reader and verification commands preserve env.sh's -buildvcs=false -trimpath and append -p=4. Test runs wrote directly to log files. The exact selectors, commands, wall seconds and test events are preserved in the result JSON files.

- python3 review/compiler/loom-slice-5/run-focused.py: focused lower and native tests pass; then its grain monitor stops on the three red-duration leaves. Evidence: focused-results.json.
- python3 review/compiler/loom-slice-5/run-focused-after-grain.py: new backend-stop leaves and initial parser checks pass; stops at the short bundled URI helper failure. Evidence: focused-after-grain-results.json.
- python3 review/compiler/loom-slice-5/run-focused-resume.py: remaining focused tests and CLI leaves pass; expected Program-region and optional measurement skips remain. Evidence: focused-resume-results.json.
- python3 review/compiler/loom-slice-5/run-fixtures.py: all 55 selected fixtures run exactly once, held to source Node, JavaScript, sanitized native, release native and completed-program leaks as the oracle requires. This includes 23 member/control fixtures, 29 legacy non-null fixtures and three stack fixtures. Evidence: fixture-results.json; eight command wall times total 79.781 s.
- go test ./internal/oracle -run '^TestSlice5Witness' -count=1 -json -timeout 90s: all eight ledger tests pass; aggregate 7.129 s, largest leaf 5.09 s. Evidence: logs/ledger.jsonl.
- go test ./internal/oracle -run 'TestCountsAreRecorded' -count=1 -json -timeout 180s -args -update-counts: success, 59.061 s, largest recorded leaf 5.11 s. Evidence: logs/counts-complete.jsonl.
- go test ./internal/oracle -run '^(TestSmallStacksStillPanic|TestLongArgumentsLeaveTheStackItsLimit)$' -count=1 -json -timeout 90s: success, 7.089 s. Evidence: logs/stack-regression.jsonl.
- go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -json -timeout 90s: success, 6.04 s leaf and 6.081 s package. Initial cold attempts reached their shell limits; preparing main's TestProduct_CallTargetExports product succeeds in 26.67 s before the final reader run. Evidence: logs/readers-product.jsonl and logs/readers-green.jsonl.
- TestEveryMapAProgramHoldsIsSorted and TestAProgramIsTheSameBytesEveryTimeGobWritesIt pass (0.00 s and 0.01 s). The plain-map mutant fails the former. Evidence: logs/readers-final.jsonl and product-mutants-results.json.
- go vet ./cmd/adamic ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle: exit 0, no diagnostics. Oracle vet was confirmed after its final ledger edits: exit 0 in 0.169 s. Evidence: logs/vet.log and vet-oracle-confirmed.json.
- Focused test package compilation used -run '^$'; no whole-package test selection or full gate was run. Final gofmt and git diff --check are clean.

Toolchain setup initially reached hard limits at 480 s and 240 s while the checkout and caches were cold. Work continued using the installed tools and local cache. The successful setup, with ADAMIC_GOCACHE_OFF=1, reports Go ready 0.022 s, Node ready 0.022 s, submodules ready 0.064 s, markdown ready 0.070 s, clang ready 0.162 s, shared cache off 0.163 s, Go build ready 18.288 s, test binaries deferred 18.449 s, build cache warm 18.450 s and done 18.481 s. nproc=5, cgroup cpu.max=400000 100000 (4 CPUs). Go 1.27.1, Node 24.19.0, clang 20.1.8. The environment file is /workspace/adamic-tools/env.sh. All three setup logs are retained under this review directory.

Independent overlay mutations and their catchers:

| Mutant | Catcher | Evidence |
|---|---|---|
| branch-union | ^TestParserFactory(Completion.* / ReadBeforeCompletion)$ | step24-parser-main-mutants-results.json |
| omit-escape-missing | ^TestParserFactory(Completion.* / ReadBeforeCompletion)$ | step24-parser-main-mutants-results.json |
| erase-earlier-read | ^TestParserFactory(Completion.* / ReadBeforeCompletion)$ | step24-parser-main-mutants-results.json |
| trust-assertion | ^TestParserFactory(Completion.* / ReadBeforeCompletion)$ | step24-parser-main-mutants-results.json |
| slot-pos | ^TestParserNodeArrayLayout$ | step24-parser-main-mutants-results.json |
| slot-end | ^TestParserNodeArrayLayout$ | step24-parser-main-mutants-results.json |
| slot-hasTrailingComma | ^TestParserNodeArrayLayout$ | step24-parser-main-mutants-results.json |
| slot-transformFlags | ^TestParserNodeArrayLayout$ | step24-parser-main-mutants-results.json |
| drop-unset-use | ^TestParserConstructionUnsetUse$ | step24-parser-main-mutants-results.json |
| drop-speculation-check | ^TestParserSpeculationMisfit$ | step24-parser-main-mutants-results.json |
| disable-tiny-stack | ^TestParserStackTinyLimit$ | step24-parser-main-mutants-results.json |
| drop-flow-check | ^TestPlaceholderUseCheckSavedLeak$ | placeholder-nonnull-main-mutants-results.json |
| retain-alias-proof | ^TestPlaceholderUseCheckAliasReset$ | placeholder-nonnull-main-mutants-results.json |
| nullish-receiver-twice | ^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/placeholder_nonnull_null_loose[.]a$ | placeholder-nonnull-main-mutants-results.json |
| json-ordinary-assertion | ^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/placeholder_nonnull_null_json[.]a$ | placeholder-nonnull-main-mutants-results.json |
| admit-weak-slot | ^TestPlaceholderWeakSlotStaysNotYet$ | placeholder-nonnull-main-mutants-results.json |
| omit-refusal-state | TestOptionalWideningSpreadOverwrite: nil-result Go panic; initialization guard only, not a backend behavioral kill | placeholder-nonnull-main-mutants-results.json |
| drop-native-nullish-tags | ^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/placeholder_nonnull_null_copy[.]a$ | placeholder-nonnull-main-mutants-results.json |
| drop-javascript-union-tag | ^TestDefaultTaggedSourceViewsDefaultBoxedWrite$ | placeholder-nonnull-main-mutants-results.json |
| plain-placeholder-map | is a plain | product-mutants-results.json |
| placeholder-physical-path | parser product changed | product-mutants-results.json |
| factory-physical-path | parser product changed | product-mutants-results.json |
| speculation-physical-path | parser product changed | product-mutants-results.json |
| unlisted-output | unlisted backend divergence: stdout differs | logs/backend-stop-mutants/unlisted-output.log |
| drop-terminal-guard | exit codes differ | logs/backend-stop-mutants/drop-terminal-guard.log |
| drop-range-error-kind | exit codes differ | logs/backend-stop-mutants/drop-range-error-kind.log |
| drop-overflow-message | non-overflow RangeError was terminal: exit codes differ | logs/backend-stop-mutants/drop-overflow-message.log |

All 27 independent overlays are rejected; no clang warning or compiler build failure is credited as a behavioral catch. The omit-refusal-state case is separately classified above. Embedded mutations include seven placeholder flow-check removals, null-tag collapse, spread readiness loss, seven readiness probes, field-check and initialization tracking removal, factory payload corruption, scanner rewind removal, stack-guard removal (ASan stack-overflow), three runtime construction/key/JSON mutations, non-null drop checks and the eight retained stopping artifacts checked against type-correct repairs. Complete test output and per-test seconds are retained.

Runtime paths changed against the final main base:

- internal/native/runtime/adamic.h
- internal/native/runtime/array.c
- internal/native/runtime/class_features.c
- internal/native/runtime/class_static.c
- internal/native/runtime/construction.c
- internal/native/runtime/heap.c
- internal/native/runtime/json_stringify.c
- internal/native/runtime/library_language.c
- internal/native/runtime/library_object.c
- internal/native/runtime/object.c
- internal/native/runtime/region.c
- internal/native/runtime/stack.c
- internal/native/runtime/union.c
- oracle/adamic.mjs

Counts: 41 added rows, 21 changed existing rows, no removed rows. All 62 before/after rows match the source slice's attributable changes; the successful regeneration introduced no unrelated row change. Exact values and member-specific reasons are also in counts-attribution.json. A/F/R/L/P/G mean allocations, frees, retains, releases, peak live and in-region values. Stop rows count resources at exit, not successful-program leak freedom.

| Fixture | Before A/F/R/L/P/G | After A/F/R/L/P/G | Member and reason |
|---|---|---|---|
| internal/oracle/testdata/4ddd17f_weak_single_narrowed.a | new | 4/2/4/6/4/0 | per-backend-stops: Counts stop at the exact ruled native Weak-after-free stop. |
| internal/oracle/testdata/backend_stop_catches.a | new | 2/2/6/4/1/0 | per-backend-stops: Non-overflow RangeError controls execute their ordinary catch blocks. |
| internal/oracle/testdata/d96d304_try_stack.a | new | 0/0/0/0/0/0 | per-backend-stops: Counts stop at the exact ruled native named recursion guard. |
| internal/oracle/testdata/non_null_initialized.ts | 0/0/0/0/0/0 | 55/55/50/109/18/0 | placeholder-nonnull-main: Completing writes now execute; scalar placeholder slots box their values and captured saves own their cells. |
| internal/oracle/testdata/non_null_literal_assignment.ts | 0/0/0/0/0/0 | 1/1/2/2/1/0 | placeholder-nonnull-main: Execute the reset, later completing write and observation; own and release the resulting value. |
| internal/oracle/testdata/non_null_static_initialized.ts | 1/0/0/1/1/0 | 9/9/12/22/6/0 | placeholder-nonnull-main: Complete the static writes; scalar boxes and strings are released on the successful path. |
| internal/oracle/testdata/non_null_static_uninitialized.ts | 1/0/0/1/1/0 | 1/0/3/2/1/0 | placeholder-nonnull-main: Allocate static storage and transfer its nullish value before the property-use check. |
| internal/oracle/testdata/non_null_uninitialized_capture.ts | 0/0/0/0/0/0 | 2/0/2/0/2/0 | placeholder-nonnull-main: Create the captured cell and closure, then stop when the callback uses the unset value as T. |
| internal/oracle/testdata/non_null_uninitialized_const.ts | 0/0/0/0/0/0 | 0/0/2/1/0/0 | placeholder-nonnull-main: Save the nullish reference then stop at the property receiver. |
| internal/oracle/testdata/non_null_uninitialized_exception.ts | 0/0/0/0/0/0 | 1/1/3/2/1/0 | placeholder-nonnull-main: Execute catch handling and its string output before the later typed-use stop. |
| internal/oracle/testdata/non_null_uninitialized_field.ts | 1/0/0/0/1/0 | 1/0/3/2/1/0 | placeholder-nonnull-main: Allocate the object, transfer its nullish field and stop at the typed property use. |
| internal/oracle/testdata/non_null_uninitialized_interface.ts | 1/0/1/0/1/0 | 1/0/2/1/1/0 | placeholder-nonnull-main: Read the nullish field through the checked structural view before stopping at use as T. |
| internal/oracle/testdata/non_null_uninitialized_iteration.ts | 1/0/0/1/1/0 | 6/1/10/8/6/0 | placeholder-nonnull-main: Allocate captured iteration cells and closures before the callback typed-use stop. |
| internal/oracle/testdata/non_null_uninitialized_loop.ts | 0/0/0/0/0/0 | 0/0/1/0/0/0 | placeholder-nonnull-main: Carry unset into the loop exit and stop at the typed use, after its reference transfer. |
| internal/oracle/testdata/non_null_uninitialized_map_entry.ts | 0/0/0/0/0/0 | 0/0/1/1/0/0 | placeholder-nonnull-main: Transfer the nullish reference before the map argument boundary stops. |
| internal/oracle/testdata/non_null_uninitialized_optional.ts | 1/0/0/0/1/0 | 3/3/2/6/3/0 | placeholder-nonnull-main: Observe unset through optional handling and complete rather than stopping at initialization. |
| internal/oracle/testdata/non_null_uninitialized_spread.ts | 1/0/0/0/1/0 | 2/0/5/5/2/0 | placeholder-nonnull-main: Copy the actual nullish field, allocate the copy and stop only at its later typed use. |
| internal/oracle/testdata/placeholder_nonnull_alias_reset.a | new | 2/1/6/7/2/0 | placeholder-nonnull-main: New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_assignment_result.a | new | 0/0/2/3/0/0 | placeholder-nonnull-main: New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_before_use.a | new | 0/0/1/1/0/0 | placeholder-nonnull-main: New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_field.a | new | 9/9/15/23/5/0 | placeholder-nonnull-main: New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_local.a | new | 12/12/11/24/4/0 | placeholder-nonnull-main: New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_before_use.a | new | 0/0/2/1/0/0 | placeholder-nonnull-main: New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_null_coalesce.a | new | 8/8/15/27/4/0 | placeholder-nonnull-main: New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_copy.a | new | 26/26/102/121/7/0 | placeholder-nonnull-main: New fixture; copy, save/restore and spread preserve nullish payloads and release their storage. |
| internal/oracle/testdata/placeholder_nonnull_null_equality.a | new | 4/4/10/16/2/0 | placeholder-nonnull-main: New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_json.a | new | 5/5/0/6/3/0 | placeholder-nonnull-main: New fixture; exact null/undefined JSON temporary fields use existing schema arms and release their objects and output strings. |
| internal/oracle/testdata/placeholder_nonnull_null_loose.a | new | 10/10/23/35/4/0 | placeholder-nonnull-main: New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_optional.a | new | 6/6/16/25/3/0 | placeholder-nonnull-main: New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a | new | 1/0/4/4/1/0 | placeholder-nonnull-main: New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_null_truthiness.a | new | 1/1/7/10/1/0 | placeholder-nonnull-main: New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_typeof.a | new | 3/3/9/11/3/0 | placeholder-nonnull-main: New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_observe.a | new | 27/27/48/79/9/0 | placeholder-nonnull-main: New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_return_assignment.a | new | 0/0/2/3/0/0 | placeholder-nonnull-main: New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_saved_leak.a | new | 1/0/3/4/1/0 | placeholder-nonnull-main: New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_scanner.a | new | 4/4/9/8/3/0 | placeholder-nonnull-main: New fixture; counted successful placeholder observations and ownership. |
| stage3/interface-downcasts/default-boxed-write.ts | 0/0/0/0/0/0 | 5/5/5/12/4/0 | placeholder-nonnull-main: Complete scalar field writes using union boxes and release their values. |
| stage3/interface-downcasts/default-read-before-set.ts | 0/0/0/0/0/0 | 1/0/4/2/1/0 | placeholder-nonnull-main: Allocate the placeholder object and check its typed structural field read. |
| stage3/interface-downcasts/default-staged.ts | 0/0/0/0/0/0 | 2/2/6/7/2/0 | placeholder-nonnull-main: Complete staged placeholder field writes and the structural-view reads. |
| stage3/interface-downcasts/readiness-identifier-uninitialized.ts | 0/0/0/0/0/0 | 1/0/4/3/1/0 | placeholder-nonnull-main: Transfer the unset reference through a structural view and stop at use as T. |
| stage3/interface-downcasts/readiness-identifier.ts | 0/0/0/0/0/0 | 2/2/6/8/2/0 | placeholder-nonnull-main: Complete reference field assignment and checked structural-view transfer. |
| stage3/interface-downcasts/readiness-number-uninitialized.ts | 0/0/0/0/0/0 | 1/0/5/3/1/0 | placeholder-nonnull-main: Preserve the numeric field nullish payload and stop at its typed structural read. |
| stage3/interface-downcasts/readiness-number.ts | 0/0/0/0/0/0 | 5/5/5/10/5/0 | placeholder-nonnull-main: Box completed numeric placeholder fields and release the boxes after structural reads. |
| stage3/parser-ahead/factories/already-complete.a | new | 1/1/4/4/1/0 | step24-parser-main: The one shared factory allocation is already complete. |
| stage3/parser-ahead/factories/complete.a | new | 1/1/5/4/1/0 | step24-parser-main: The shared factory allocation is completed by checked writes. |
| stage3/parser-ahead/factories/escaped.a | new | 1/1/3/4/1/0 | step24-parser-main: The shared factory allocation escapes before completion and retains checked writes. |
| stage3/parser-ahead/node-array/fields.a | new | 12/12/10/23/12/0 | step24-parser-main: Four NodeArray metadata fields retain their own slots. |
| stage3/parser-ahead/rulings/factory-unset.a | new | 3/3/13/20/3/0 | step24-parser-main: Presence and saved-local observations preserve unset factory fields. |
| stage3/parser-ahead/rulings/factory-use.a | new | 1/0/2/3/1/0 | step24-parser-main: Resources retained at the named declared-string use stop. |
| stage3/parser-ahead/rulings/node-array-json.a | new | 2/2/5/9/2/0 | step24-parser-main: JSON serializes elements while omitting metadata extras. |
| stage3/parser-ahead/rulings/node-array-keys.a | new | 15/15/12/16/9/0 | step24-parser-main: Indexed and extra metadata keys are enumerated. |
| stage3/parser-ahead/rulings/node-array-length.a | new | 4/4/4/9/2/0 | step24-parser-main: Length observes elements independently of metadata. |
| stage3/parser-ahead/rulings/node-array-unset.a | new | 4/4/18/29/4/0 | step24-parser-main: Absence observations precede metadata publication. |
| stage3/parser-ahead/rulings/null-placeholder-pending.a | new | 2/2/4/8/2/0 | step24-parser-main: Null payloads stay separate from undefined/unset factory state. |
| stage3/parser-ahead/rulings/own-key-order.a | new | 28/28/17/37/10/0 | step24-parser-main: Key arrays and JSON strings observe publication order. |
| stage3/parser-next/speculation/arrow-tryparse.a | new | 3/3/8/11/2/0 | step24-parser-main: Arrow callback result specialization preserves scanner state. |
| stage3/parser-next/speculation/misfit.a | new | 1/0/2/1/1/0 | step24-parser-main: Resources retained at the speculative result type check. |
| stage3/parser-next/speculation/named-lookahead.a | new | 5/5/1/7/4/0 | step24-parser-main: Named callback result specialization retains scanner snapshots. |
| stage3/parser-next/speculation/named-tryparse.a | new | 3/3/10/16/3/0 | step24-parser-main: Named callback result specialization commits or restores scanner state. |
| stage3/parser-next/speculation/truthiness-rewind.a | new | 34/34/46/85/6/0 | step24-parser-main: False results restore snapshots and later observations. |
| stage3/parser-next/stack/benchmark.a | new | 7/7/0/17200008/4/0 | step24-parser-main: Repeated legal recursive work releases its allocations. |
| stage3/parser-next/stack/legal.a | new | 7/7/0/1728/4/0 | step24-parser-main: Normal recursive parser work releases its 859 frames. |

Assumption retained in the commit message: Outcome 24 means the delivered, bounded heap-backed parser construction, NodeArray, speculation and stack witnesses. Program-region lifetime remains explicitly pending. The two ruled backend divergences and eight exact declared-type negative witnesses retain their separate contracts; no new exception or admission is inferred from their names.
