Built slice 5 on main: placeholder initializers, parser construction and speculation, named stack guards, and exact backend-stop rulings for Outcome 24.
Member commits: bf65c361, df996382, a47ac4cf; contract/counts correction ca547958; current-main merge 235bc94c.
Commands: bounded lowering and oracle shards pass; Node/native/JavaScript fixtures pass; counts regeneration and verification pass; reader guard, vet and lane checks pass.
Mutants: all 23 recorded compiler/backend mutations and both catalog entries caught; executed IR/runtime mutations are documented below.
Not covered: complete native TypeScript parser, Program-region lifetime, arbitrary-input admission equivalence, full gate, other operating systems, and runtime-owner clearance.

Base was origin/main 5e33a17b186a8a2218d27b69b21e2de5acc5b750. Current main 98008bbba2881939e07a9d74031994102a3bbe36 was merged only after it landed. No other slice, chain, area branch, or unlanded worker branch was merged. The chain used for source/repair audit was f5236b48ee11d680c5288e3074c2fec6ad26a09a. The delivery contains one feature commit per retained member, plus the main-contract/counts correction, current-main reconciliation and review evidence.

| Member | Source | Status | Own commit | Real prerequisite | Work closed on this branch |
| --- | --- | --- | --- | --- | --- |
| placeholder-nonnull-main | bafb9ef4 | kept | bf65c361 | none outside main | #9wc5q5j reproduction and proofs |
| step24-parser-main | e503868c | kept | df996382 | placeholders in this slice | Outcome 24 factory, NodeArray, speculation and stack slice; #ktz9fek and #cs3ehy8 contract work |
| per-backend-stops | 7a151f7f | kept | a47ac4cf | named stack guard in this slice | #z1vjxxd reproduction and proofs |

Task #wj4pmt1 receives this independently based slice; the complete chain task and integration acceptance are not claimed closed. Runtime-owner clearance remains outstanding.

Source extraction follows the refined ranges, excluding review evidence and stale counts, rather than branch-tip ancestry:

- placeholders: e8b02f8ca1ab02b5ecf1cd1b2b5a8f6d026d0703..bafb9ef4334a8fb36713c92b0bb372d43edd546a.
- parser: bafb9ef4334a8fb36713c92b0bb372d43edd546a..e503868c680ce67691e602446539f64f463d4e95.
- backend stops: b7c8d7a8e184fcbb7be6a4c80c1d91555634f188..7a151f7fdb2ec8431320b73d7aa515fa1318e243.

Dependency evidence: parser reads placeholderOrigin at internal/lower/locals.go:135; placeholders define it at internal/lower/placeholder.go:107. Backend-stop expectations require ADAMIC_CHECK_STACK_NAMED at internal/native/runtime/adamic.h:1074, supplied by the parser member. Main already supplies namespace/value flags, structural-view helpers, sameKeeping, libraryArrayBuilder and call-target APIs. Production CLI builds and all focused tests resolve their symbols on main plus these three own-net changes. No required outside-member symbol, IR field, helper or fixture was found. No member was dropped.

The parser's separate Program lifetime test still reads exactly t.Skip("awaits compiler/program-region-lowering: parser trees in the Program region"). This is an explicitly pending feature, not a missing prerequisite for the delivered heap-backed witnesses. Optional-only NodeArray presence and the generic factory scout retain their pinned refusals.

Repair/conflict decisions: retain current main's host-cycle refusal and source location; it was the only extraction conflict. Carry only the placeholder nil-property-symbol guard from repair 2b2b1095. The d8e6f59c exceptions merge preserved parser own-key semantics but contributes no additional required own-net parser hunk here. The chain repair's factory-use exit 1 and tiny-stack exception_repeat expectations belong to exceptions-21's outside uncaught/cleanup contract. This branch keeps main's uncaught exit 70 and parser's parseNested expectation. The initial incorrect exit-1 test adaptation was corrected in ca547958, without importing exceptions-21. When current main added per-emission metadata caching, resolve emit_objects.go once by retaining main's cache and including constructionNeeded in fieldTypes; do not replace the cache with the old repeated walk. Main's emit.go is inherited unchanged relative to final main.

Validation commands run from the repository root, with /workspace/adamic-tools/env.sh sourced, GOCACHEPROG unset, GOMAXPROCS=4 and GOFLAGS adding -p=4. Every command writes a complete log and has a hard timeout. No whole package test invocation or full gate was run.

- `python3 review/compiler/chain-slice-5/run-lower-shards.py`: 16 exact-name shards, each `go test ./internal/lower -run <selection> -count=1 -json -timeout 90s`, subprocess limit 100s. All 310 enumerated top-level tests accounted for: 308 pass and two existing skips. Rerun after merging main; lower-results.json and lower-union.json record the final union.
- `python3 review/compiler/chain-slice-5/run-oracle-shards.py`: seven exact-name shards of at most eight tests, `-timeout 90s`, 51 selected top-level leaves pass, including checked reads, source parser fixtures, backend rulings, stack-removal, payload, rewind and readiness mutants. Final logs are oracle-shard-*.jsonl.
- `python3 review/compiler/chain-slice-5/run-member-fixtures.py`: 23 member/root-control fixture programs in three bounded selections of TestNativeAgreesWithNode, all pass. Dedicated parser-source tests cover non-root parser fixtures with source Node, both backends, ASan/UBSan and completed-program leaks.
- `python3 review/compiler/chain-slice-5/run-legacy-fixtures.py`: all 29 registered legacy non-null/readiness fixtures in four bounded selections pass. Checked failures use the existing explicit checked-fixture contract.
- `python3 review/compiler/chain-slice-5/run-remaining.py`: the five changed CLI leaves, seven native parser/metadata leaves plus current main's cache guard, and 12 additional oracle leaves pass; Program lifetime has its explicit pending skip.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -v -timeout 10m -args -update-counts`: successful full regeneration once, 126.438s aggregate. An earlier failed prerequisite sweep could not write the table because the test refuses to write after any failure. Installed recorded @types/node through `npm ci --prefix stage3/api`. `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -v -timeout 4m` verifies the merged state without rewriting, 219.522s aggregate. `TestParserAheadCounts` verifies parser rows, 9.273s.
- `go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s`: final pass, 13.87s; no allowlist additions.
- `go vet ./cmd/adamic ./internal/lower ./internal/native ./internal/oracle`: exit 0. Lane vet exceeded its short cutoff, so this separate bounded invocation completes it.
- `git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -`: passes. Output: lane checks 40.1 s: gofmt and tools on 68 Go files, t.Parallel on 4 test packages; a-check 22 .a files; vet skipped, over 10 s. The committed evidence-only follow-up is checked again before push.
- Changed Go files are gofmt-formatted and `git diff --check` is clean.

Admission evidence: compile the existing cmd/adamic twice, candidate normally and current main through a Go source overlay; no new executable Go probe is added. `run-admission-census.py` compares C-emission admission for every main recorded counts fixture and every added member .a: 1,071 programs, 30 newly admitted, zero regressions, zero per-program timeouts. Both production CLI builds pass. The census was repeated against final main after the cache merge. See main-baseline.json, admission-results.json, admission-delta.json and admission-proof.json.

This is a finite corpus proof, not a universal claim about all possible programs. Of the 30 new admissions, 21 agree with raw source Node on stdout/stderr/exit across both backends. Nine negative witnesses intentionally violate a declared-T use contract: seven placeholder checks, factory-use and speculative misfit. Both backends stop with the pinned exit-70 diagnostic; their independent raw Node observations and check-removal mutants remain recorded. Literal raw-Node equality for those nine is false and cannot be supplied without removing the requested checks. This report does not disguise checked stops as Node equality.

Counts attribution: 41 new rows, 21 changed existing rows, zero removed rows. All 62 rows have before/after values, a member and a specific reason in counts-attribution.json. The 40 placeholder rows match the author's independently recorded ledger exactly. The 19 parser rows measure factory publication, own keys, element-only JSON, metadata slots, speculation and normal recursion; the three backend-stop rows measure exact stopping points or catch controls. Stop counts describe resources live at exit, not successful-program leak freedom.

Recorded compiler/backend mutant outcomes:

| Mutant | Catcher | Result |
| --- | --- | --- |
| drop-flow-check | `^TestPlaceholderUseCheckSavedLeak$` | caught |
| retain-alias-proof | `^TestPlaceholderUseCheckAliasReset$` | caught |
| nullish-receiver-twice | `^TestNativeAgreesWithNode$/internal/oracle/testdata/placeholder_nonnull_null_loose[.]a$` | caught |
| json-ordinary-assertion | `^TestNativeAgreesWithNode$/internal/oracle/testdata/placeholder_nonnull_null_json[.]a$` | caught |
| admit-weak-slot | `^TestPlaceholderWeakSlotStaysNotYet$` | caught |
| omit-refusal-state | `^TestOptionalWideningSpreadOverwrite$` | caught |
| drop-native-nullish-tags | `^TestNativeAgreesWithNode$/internal/oracle/testdata/placeholder_nonnull_null_copy[.]a$` | caught |
| drop-javascript-union-tag | `^TestDefaultTaggedSourceViews$/default-boxed-write$` | caught |
| branch-union | `^TestParserFactory(Completion.*|ReadBeforeCompletion)$` | caught |
| omit-escape-missing | `^TestParserFactory(Completion.*|ReadBeforeCompletion)$` | caught |
| erase-earlier-read | `^TestParserFactory(Completion.*|ReadBeforeCompletion)$` | caught |
| trust-assertion | `^TestParserFactory(Completion.*|ReadBeforeCompletion)$` | caught |
| slot-pos | `^TestParserNodeArrayLayout$` | caught |
| slot-end | `^TestParserNodeArrayLayout$` | caught |
| slot-hasTrailingComma | `^TestParserNodeArrayLayout$` | caught |
| slot-transformFlags | `^TestParserNodeArrayLayout$` | caught |
| drop-unset-use | `^TestParserConstructionUnsetUse$` | caught |
| drop-speculation-check | `^TestParserSpeculationMisfit$` | caught |
| disable-tiny-stack | `^TestParserStackTinyLimit$` | caught |
| unlisted-output | `TestUnlistedBackendAgreement: stdout differs` | caught |
| drop-terminal-guard | `TestTerminalStackStop: exit codes differ` | caught |
| drop-range-error-kind | `TestNativeAgreesWithNode/backend_stop_catches: exit codes differ` | caught |
| drop-overflow-message | `TestNonOverflowRangeErrorCatch: non-overflow RangeError became terminal` | caught |

No compiler or clang failure is credited as a behavioral mutant catch. The omitted standalone-refusal-state mutant deliberately demonstrates the recorded nil-result Go panic. Exact commands, exits, seconds, overlays and complete logs are in compiler-mutants-results.json and the member-specific result files. Catalog 09's numeric-zero mutation produces `0 false` instead of Node's `undefined true`; its clean control passes. Catalog 12 removes suppression refusal and fails the corresponding expected-refusal cases; the clean control is in the lowering shards. Both are rerun on merged main.

Executed embedded mutations, with their catchers:

- Seven placeholder flow-check removals: TestPlaceholderFlowMutantBeforeUse, SavedLeak, NullBeforeUse, NullSavedLeak, AssignmentResult, ReturnAssignment and AliasReset. Each check removal reaches the raw Node outcome and fails the pinned checked-use contract in both backends.
- Null-to-undefined tag collapse: TestPlaceholderNullTagMutant catches Node equality output changing. Lost spread storage readiness: TestPlaceholderSpreadReadinessMutant catches an exit-70 stop where Node finishes.
- Seven readiness cases under TestReadinessMutants: replace-unset-check-with-zero, erase-without-proof, initialize-to-zero, miss-captured-read, miss-exception-path, replace-missing-assertion-with-empty and weak-generic-message. Pinned outcomes reject missing checks; the latter two execute their migrated drop-check controls.
- Nonliteral check removal and confusing unset with nullish: TestNonliteralInitializerCannotSkipCheck and TestUninitializedIsNotNullishMutant. Shared view-field check and initialization tracking removals: TestRequiredViewFieldPrimitive. Freed-Weak diagnostic mutation: TestNonNullWeakFreedNamesExpression.
- Factory payload changed from ready to wrong: TestParserAheadFactoryCompleteControl catches source Node stdout differing in both backends. Scanner rewind removal: TestParserSpeculationRewindMutant catches Node stdout differences. Runtime stack guard removal: TestParserStackGuardMutant executes and catches AddressSanitizer stack-overflow.
- Runtime synthesized-key publication, declaration-order key enumeration and array extras in JSON: TestParserConstructionRuntimeMutantSynthesizedKey, DeclarationOrder and ArrayExtrasInJSON. Each compiles and executes with sanitizer/leak checks and normal exit, then differs from actual source Node stdout.

Final per-leaf timings are in test-leaf-seconds.json; the largest oracle/runtime leaf is 37.20s. Lowering shard timings are in lower-results.json; the final largest shard is 15.556s. The optional guard-cost measurement is explicitly skipped by default and was not measured. Initial cold/concurrent runs crossed the requested leaf budget and one combined oracle selection exhausted its 90s aggregate limit; those attempts are retained but not used as green proofs. Final selections were split and run with warm setup/runtime caches. The requested first-green push ceiling was missed during setup/cache recovery; no partial state was pushed.

Toolchain: GOPROXY was set to https://proxy.golang.org|direct before setup. Setup first hit hard limits while shared-cache Go compilation was cold/stalled; disabling the shared cache and warming the local cache succeeded. Final timing lines: go 0.053s, Node 0.059s, submodules 0.118s, markdown 0.129s, clang 0.375s, shared cache off 0.377s, go build 101.499s, test binaries deferred 101.706s, build cache warm 101.709s, done 101.833s. nproc=5; CPU quota=4. Go 1.27.1, Node 24.19.0, clang 20.1.8. Complete setup output is logs/setup-final.log.

Automatic approval review rejected the proposed scratch Go admission probe outside review. That action was abandoned. The completed census instead uses the existing cmd/adamic and non-compilable overlay evidence; no requested proof remains blocked by that rejection.

Runtime files changed, listed for @system_adamic_runtime clearance:

- `internal/native/runtime/adamic.h`
- `internal/native/runtime/array.c`
- `internal/native/runtime/class_features.c`
- `internal/native/runtime/class_static.c`
- `internal/native/runtime/construction.c`
- `internal/native/runtime/heap.c`
- `internal/native/runtime/json_stringify.c`
- `internal/native/runtime/library_language.c`
- `internal/native/runtime/library_object.c`
- `internal/native/runtime/object.c`
- `internal/native/runtime/region.c`
- `internal/native/runtime/stack.c`
- `internal/native/runtime/union.c`
- `oracle/adamic.mjs`

Runtime-owner clearance has not been obtained by this worker. No cohere file was copied. The complete parser and Program lifetime, 301-project native corpus certification, fresh pinned scout remeasurement, optional-only array relation admission, whole gate, WASI and non-Linux platforms were not covered. Delivery serves Outcome 24's bounded native parser construction/speculation/stack machinery, pending integration acceptance.
