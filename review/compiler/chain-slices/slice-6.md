TypeScript binder.ts:1691:17 has a stale FlowLabel narrowing after bind(finallyBlock) can install FlowAssignment; .a needs adaptation (see narrowing-sites.md).
Built exact uncaught diagnostics, the .a writing-call refusal/control, TypeScript stop witness, memory correction and tab indentation fixes on main 98008bbb.
Member commits stay df93806d, bab6155e, d10e3e00, cd0b8e5f; this is one production fix on top of cleared 66ad4a67, with review-only delivery evidence afterward.
Lower shards: 281 pass, two existing skips; 14 admission agreements, zero .a witnesses; counts refresh and six follow-up mutants pass.
Not covered: a native tsc binary, WASI execution, unresolved indirect calls or a universal whole-program effect/admission proof.

# Chain slice 6

## Runtime condition and step 21 ruling follow-up

Native and JavaScript write exactly one diagnostic line. Error instances, including subclasses, use their actual shape fields: `Uncaught Name: message`. Other kinds print `undefined`, `null`, `number`, `boolean`, `string`, `function`, or `object` (arrays/maps included). CR and LF in fields are escaped; lone UTF-16 surrogates encode as U+FFFD in both backends. Native flushes stdout, writes/flushes stderr, releases the thrown payload and exits 1. The JavaScript handler synchronously writes the same bytes and exits 1. Source Node's renderer remains excluded by the step 21 ruling. TestStep21Uncaught pins exact bytes for sanitized native, release and JavaScript; TestStep21UncaughtKinds adds kind, subclass, CR/LF and Unicode coverage. Error creation indentation now uses tabs; docs/memory.md no longer describes uncaught exceptions as exit-70 panics.

The .a refusal is `:7:21: Adamic 0.1 refuses a narrowed read of value after change() can write it; narrow again after the call`. The original step21_builtin_narrow_terminal.a is unchanged. The control changes only the write target to another variable and prints `TypeError` then `finally` in source Node, JavaScript, sanitized native and release native, with no leaks. Captured/module binding assignments are propagated through resolved direct/transitive callees and recursive write summaries; writes to another activation's own locals are excluded. Fresh instanceof checks read the held union to establish the proof again. Calls that cannot write the binding remain admitted. Dynamic observations (including qualified reads) and slots accepting the entire declared type do not consume the stale proof.

The general rule also refuses existing deliberate stale-read .a fixtures. Their source files are intact; exact refusal tests replace runnable registrations. Existing narrowed-union IR stop mutation proofs use temporary TypeScript compatibility copies; no negative-witness-list support was added. The source analysis and precise coverage limits are recorded in ../chain-slice-6/narrowing-sites.md. It scans all 77 original pinned compiler files plus the pin's generated diagnostic map, with zero semantic diagnostics; ten pairs at seven distinct reads. Reads that save/restore correctly and unreachable switch paths are identified. The reported FlowLabel defect concerns the static narrowing; the .flags read remains valid for wider FlowNode values. No native tsc runtime result or JavaScript behavior defect is claimed.

The external admission-delta worker should add these facts under **step 21: unsound narrowing across a writing call**:

- Source: internal/oracle/testdata/step21_builtin_narrow_terminal.ts, byte-identical to the .a source.
- SHA-256: c6273c04f4e421e248443ab2c5b27196332d89409f2c54fb123b8a901054513b.
- Ruling: #bqj5drt's step 21, ruled Oct 9 18:0x by @system_adamic.
- Violated narrowing: internal/oracle/testdata/step21_builtin_narrow_terminal.ts:5 (`value instanceof TypeError`); binding declaration :2; consuming read :7:21.
- Node: exit 0, stdout `Error\nfinally\n`, empty stderr.
- Sanitized native, release native and JavaScript: exit 70, empty stdout, exactly `adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n`.

Machine-readable facts are in ../chain-slice-6/narrowing-witness.json. The witness list itself remains owned by compiler/admission-delta.

| Follow-up proof | Exact scope/result | Evidence under ../chain-slice-6/ |
| --- | --- | --- |
| Lower | run-followup-lower.py; 283 exact top-level tests in 15 shards, -count=1 -json -timeout 90s; 281 pass, two original skips; max shard 27.759 s. Qualified wider-slot reads were rechecked after the final precision adjustment. | followup-lower/lower-results.json, followup-lower/lower-union.json; followup-observations.jsonl |
| Exceptions and TS witness | go test ./internal/oracle -run '^TestStep21\|^TestNarrowedUnion' -count=1 -json -timeout 90s; pass 18.732 s. | followup-exceptions.jsonl |
| Positive control | TestStep21NonWritingControlAgreement; source/JS, ASan/UBSan, release, leaks; pass 5.872 s. Lower's independent control is the over-refusal mutant catcher. | followup-control.jsonl, narrowing-tests-final.jsonl |
| Admission agreements | TestNativeAgreesWithNode exact flat, nested review/refused and structural selectors; all 14 newly admitted ordinary .a fixtures pass, zero .a witnesses. | followup-admission-{flat,review,structural}.jsonl, admission-delta-fix.json |
| Reader guard | go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -json -timeout 90s; pass 0.660 s. First concurrent attempt did not complete; isolated rerun is credited. | followup-readers-final.jsonl |
| Counts | go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -json -timeout 180s -args -update-counts; pass 82.651 s. The preceding 180 s attempt timed out before writing. Table refreshed once. | followup-counts-final.jsonl, followup-counts-attribution.json |
| Feature/runtime link | Four exact feature selectors with ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/chain-slice-6-tsgo.a; 32 matching masks, mismatch probes and actual checker archive; pass 9.331 s. | followup-features.jsonl |
| Lint profile | TestProfileRuntimeFeaturesMatchNode; pass 0.318 s. | followup-lint.jsonl |
| WASI compile/opt-in | TestWASIRequestThrows compiles and skips under its existing ADAMIC_ORACLE_WASI opt-in; stderr expectation updated to `Uncaught Error: boundary\n`. | followup-wasi.jsonl |

Every follow-up mutant below failed at its intended behavioral/refusal assertion; build failures receive no credit.

| Mutant | Catcher |
| --- | --- |
| backend-renders-stack | TestStep21Uncaught exact diagnostic, multiline Node stack differs |
| backend-drops-line | TestStep21Uncaught exact diagnostic, empty line differs |
| native-renders-stack | TestStep21Uncaught exact diagnostic, extra frame line differs |
| native-drops-line | TestStep21Uncaught exact diagnostic, empty line differs |
| drop-write-set-test | TestStep21NonWritingCallControl refuses the non-writing control |
| admit-writing-call | TestStep21WritingCallRefusal receives nil instead of the exact refusal |

Results, commands, bounded logs and non-compilable sources are under followup-mutants/. followup-test-seconds.json records test seconds. All added test leaves are under 60 s. The uncaught renderer itself changes no runtime counts: its borrowed shape/text reads allocate and retain nothing. The combined narrowing fix moves three retained rows: library_failures loses two retain/release pairs, library_types and error_subclasses lose one pair each because fresh instanceof no longer calls a stale nominal-read helper. Eleven runtime-count rows disappear because their .a sources are now compile-time refusals. Every row is attributed, with generated-C diffs, in followup-counts-attribution.json and followup-counts-diffs/. Stdout/exit behavior of admitted programs does not change.

Runtime files changed by this follow-up: internal/native/runtime/exceptions.c and internal/native/runtime/adamic.h. The complete slice's ten runtime paths remain listed below for @system_adamic_runtime. The branch remains based on 98008bbb; fetching lane references does not merge later main or another worker's branch.

## Extraction and dependencies

Initial main: 5e33a17b186a8a2218d27b69b21e2de5acc5b750. Final main parent: 98008bbba2881939e07a9d74031994102a3bbe36. The chain head f5236b48 was inspected, never merged. No other slice was merged or copied. The source plan, refined ranges and complete source histories are preserved in ../chain-slice-6/.

| Member | Source SHA | Fork point | Decision | Dependency evidence | Tasks served |
| --- | --- | --- | --- | --- | --- |
| exceptions-21-main | 205586a0899488595200f3fca470874ef9282b37 | 7a10c877 | kept | This slice supplies Read.Throws/Assign.Throws at internal/ir/readiness.go:4 and propagation at internal/lower/readiness.go:56. Main supplies local readiness. No required outside production symbol. | step 21; #dv99xzy, #agwccbw |
| inherit-guards | 16a0b626b14b8fd6b541edc6cbb92e63b8ac55f5 | 291ee604 | kept | Main already supplies the void/never guard at internal/lower/class_inheritance.go:188 and inherited layout at :237. Existing guards retained; regression evidence added. | #a898y40 regression coverage |
| feature-set-link-main | e1efb527a314d834fc6fb282108e992558048eca | 553ad06a | kept | Main supplies RuntimeLibraryForSource and sourceFlags at internal/native/library.go:32/:44. Feature relocation/definition and source-aware helper fixes are supplied here. | #hmab710; Outcome 32 prerequisite |
| lint-features | 3535387095d50f05d80e2717aed468843e59bc18 | 2a28375f | kept | Requires feature-set-link-main in this slice. SourceFlags export at internal/native/library.go:37; matching lint runtime/flags at stage1/cohere/lint/profile_compilation_main_test.go:357. | lint profile link repair |

No member dropped. The necessary lint/feature-link interaction is kept together. Tasks have local delivery evidence and await integration acceptance; this does not claim Outcome 32 itself complete.

Repair ownership: 71972e18 supplies lexical TDZ exceptions, ReferenceError and readiness propagation; a6cb5660 restores Node/JS/native cycle exit checks; c775dc61 supplies observation logging and its bypass mutant; aea87960 supplies the WASI request uncaught exit ruling. d8e6f59c exception payload/pending ownership is retained. Main's unrelated guards and behavior were kept when conflicts were resolved.

Excluded outside portions: Read.Unset, Defined.Throws and namespace readiness are absent from main and unnecessary for its lexical readiness model; absent-message-shape runtime changes belong to another member. The lint parser_construction_test.go hunk names an outside fixture absent on main. The 2b2b1095 named-outcome WASI hunk needs ruled-backend-outcomes (2391c655) for backendDisagreement at internal/oracle/wasi_test.go:76. That helper is absent on main; its named-stack/weak-stop fixtures were not copied. Main's disagreement helper already covers the exception-owned exit-1 contract. No slice 1-4 changes imported.

## Initial delivered commands and observed results (66ad4a67)

Setup exported GOPROXY=https://proxy.golang.org|direct. First timeout 240 bash cloud/setup.sh exited 124 without diagnostic; retry with ADAMIC_GOCACHE_OFF=1 completed. Printed elapsed seconds: Go 0.180, Node 0.234, submodules 0.484, markdown 0.522, clang 1.265, shared cache off 1.289, build 402.140, deferred 402.302, cache warm 402.304, done 402.429. nproc=5; Go capped at four CPUs. Go 1.27.1, Node 24.19, clang 20.1.8. Missing pinned Node types were installed with npm ci --prefix stage3/api. The checker archive build expired once and completed on retry. Setup delays and proof work exceeded the requested first-green window.

Build/test commands source /workspace/adamic-tools/env.sh, unset GOCACHEPROG, and use GOMAXPROCS=4 with GOFLAGS='-buildvcs=false -trimpath -p=4'. Every command has a hard bound; output goes to files. No whole-package tests or full gate ran.

| Proof | Command/scope and output | Evidence under ../chain-slice-6/ |
| --- | --- | --- |
| Lower | run-lower-shards.py: 279 top-level tests, exact selectors in lower-results.json, -count=1 -json -timeout 90s. 14 green shards; max 36.547 s. 277 pass, two original skips. | lower-union.json, lower-00..13.jsonl |
| Member fixtures | TestNativeAgreesWithNode: step21_*, inherited_fields_guard, lowering_chain_tdz_catch, arguments_length_value_count. Node, JS, release native, ASan/UBSan native and leaks. Host retry passed after Node types installation. Final main: 19 selected leaves pass. | member-oracle.jsonl, member-host-retry.jsonl, final-main-oracle.jsonl |
| Exception tests | go test ./internal/oracle -run '^TestStep21' -count=1 -json -timeout 90s: pass 5.65 s. Includes five executed IR mutants, unknown assertion barriers, throw liveness, uncaught rendering and oracle hashes. | step21-tests.jsonl |
| Composition gaps | go test ./stage1/cohere/css ./stage1/cohere/json ./stage1/cohere/graphql -run '^(TestClosedRepeatInTryGap|TestEachGapStandsWhereGapsMdSaysItDoes)$' -count=1 -json -timeout 90s: pass; Node, JS, sanitized native and leaks for closed exception gaps. | member-gap-tests.jsonl |
| Feature matrix | TestRuntimeFeatureMismatchProgramFeature, TestRuntimeFeatureMismatchRuntimeFeature, TestRuntimeFeatureSetsNative, TestSplitTSGoRuntimeFeatures: final pass 10.098 s. 32 matching masks, 160 one-bit mismatches, mixed unused unit and actual checker archive positive control. ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/chain-slice-6-tsgo.a. | features-final-main.jsonl, split-checker-positive.jsonl |
| Lint | go test ./stage1/cohere/lint -run '^TestProfileRuntimeFeaturesMatchNode$' -count=1 -json -timeout 90s: final pass 0.346 s. | final-main-lint.jsonl |
| Admission witnesses | Exact nested review/refused/fxspptb_(9984394_lib_dispatch,903f25b_try_assign,57f2d04_with_frozen) and statements_small_stopped/structural_error selectors: final main pass, three plus one leaves. | final-main-review-oracle.jsonl, final-main-structural-oracle.jsonl |
| Readiness | Recorded bypass mutant fails expected exit 1 and Node stdout; restored TestUndecidedCycleReadsUseReadyChecks passes 2.997 s. | mutants/readiness-bypass/test.log, readiness-restored.jsonl |
| Reader guard | go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -json -timeout 90s: final repair pass 0.715 s. Stale throwsOut allowlist renamed with compiler ownership. | readers-final-repairs.jsonl |
| Counts | go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -json -timeout 180s -args -update-counts: pass 99.910 s. Two earlier 90-second attempts expired without writing; table wrote once successfully. Maximum individual counts leaf 3.21 s. | counts-final.jsonl |
| Attribution | All 126 changed rows: 104 existing changes and 22 new registrations, zero removed rows. Existing changes have generated main/slice C diffs. New inherited-fields registration produces identical C on main. | counts-attribution.json, counts-diffs/ |
| Lane | Required fetch/show lane-checks pipeline passed formatting, tools, parallel analyzer and a-check. Its vet budget expired; separate go vet on all changed test packages passed. Final delivery lane: 69 Go files, 13 test packages, five a-check files and vet on all 13 packages; 4.5 s. | lane-checks.log, lane-checks-final.log, lane-delivery.log, vet.log, vet-final.log |
| WASI request | go test ./cmd/adamic -run '^TestWASIRequestThrows$' -count=1 -json -timeout 90s: compiles and skips; WASI execution opt-in/SDK not configured. | request-wasi-optin.jsonl |

Test seconds are in test-seconds.json; largest member leaf observed 38.58 s. No new test leaf exceeds 60 s. The counts aggregate writer is an existing test. Unsuccessful setup, stale allowlist and missing-type attempts are retained and not credited as passes.

## Admission delta on the final refusal policy

The original finite 1,293-file internal/oracle/testdata .a corpus was rerun against main and the final slice: 14 newly admitted agreements and zero .a negative witnesses, no timeouts. Every one of the 14 ran against source Node in JavaScript, sanitized native (ASan/UBSan), release native and leak checks when source Node exits successfully. Commands and leaves are in followup-admission-{flat,review,structural}.jsonl. The additional positive control is under internal/lower/testdata and has its own sanitized/release/JS/Node and leak proof; it is not included in this fixed corpus.

The general writing-call refusal intentionally removes 15 main admissions from that corpus. Every moved source has an exact refusal test, with the builtin and primitive fixtures tested separately. See admission-refusals-fix.json and TestStep21ExistingWritingCallRefusals. This includes a type-preserving `keep()` write: the ruling tests whether the write set reaches the narrowed binding, not whether one particular runtime write preserves its type. Dynamic observations and wider receiving slots remain admitted. No other admission regression was observed. The earlier 15th newly admitted checked-negative .a is now refused; its exact TypeScript witness facts follow below. This is a finite corpus proof, not a proof for all possible programs.

## Initial delivered mutants and catchers (66ad4a67)

Compilation failures were not credited. Go/C mutations use overlays. Node .mjs mutations write the file Node reads and restore in finally; the first attempted Go overlay did not affect direct Node reads and was not credited. M02 also failed native exit 70 against Node exit 0.

| Case | Catcher |
| --- | --- |
| missing-exception-edge | runtime error: |
| assignment-kills-handler-value | old text is dead |
| throw-temporaries-leak | leaks: |
| pending-finally-leak | leaks: |
| error-message-not-retained | AddressSanitizer: heap-use-after-free |
| range-error-is-error | stdout differs |
| wrong-range-message | stdout differs |
| type-error-is-error | stdout differs |
| hash-failure-is-panic | exit codes differ |
| host-type-error-is-error | stdout differs |
| soundness-check-is-catchable | want the inserted check to fire |
| rethrow-wrapped | stdout differs |
| builtin-narrow-forgets-subtype | want the inserted check to fire |
| saved-error-wrapped | stdout differs |
| backend-renders-stack | uncaught renderer wrote stderr |
| error-ancestry-missing | stdout differs |
| error-default-name-wrong | stdout differs |
| error-prefix-released-twice | AddressSanitizer: heap-use-after-free |
| error-fields-out-of-order | stdout differs |
| implicit-string-conversion | uncaught lifetime check: exit 70 |
| pending-undefined-lost | stdout differs |
| catch-assumes-error | exit codes differ |
| typeof-boolean-is-number | stdout differs |
| null-is-undefined | stdout differs |
| finally-releases-twice | AddressSanitizer: heap-use-after-free |
| region-payload-freed | AddressSanitizer: heap-use-after-free |
| uncaught-is-panic | uncaught lifetime check: exit 70 |
| changed-source-keeps-terminal-mode | changed source borrowed terminal convention |
| changed-dependency-keeps-terminal-mode | changed dependency borrowed terminal convention |
| backend-drops-entry-source | JavaScript backend: exit codes differ |
| readiness-old-panic | JavaScript language TDZ |
| readiness-old-effects | stdout differs |
| unknown-authorizes-assertion | want Refused, got <nil> |
| IR catch_callback | executed marker changes output against Node in all three backends |
| IR finally_callback | executed finalizer marker changes output |
| IR rethrow | executed handler marker changes output |
| IR finally_completion | executed finalizer marker changes output |
| IR liveness | executed handler marker changes output |
| M02 | B.Fields guard; native/Node exit disagreement |
| M07 | want NotYet for void result representation |
| P_LOWER | accepted inheritance must produce nonempty IR |
| fixed-symbol | mismatched runtime linked successfully |
| drop-reference | mismatched runtime linked successfully |
| drop-retain | mismatched runtime linked successfully |
| header-forces-runtime-features | mismatched runtime linked successfully |
| split-runtime-default-flags | feature link rejection and matching positive-control failure |
| split-original-checker-default-flags | original checker fixture fails at unsanitized feature link |
| lint-profile-default-runtime | undefined matching feature symbol |
| readiness-bypass c775dc61 | Node stdout and expected exit 1 differ from native |

Sources are non-compilable .go.txt/.c.txt/.mjs.txt, with overlays and logs under ../chain-slice-6/{exception-mutants,feature-mutants,mutants}/. exception-mutants/results.json records 33 exception source mutants; feature-mutants.log and inherit-mutants.log record other campaigns. The five IR mutations are observed in step21-tests.jsonl.

## Runtime clearance

For @system_adamic_runtime, own runtime files changed against current main:

- internal/native/runtime/adamic.h
- internal/native/runtime/class_inheritance.c
- internal/native/runtime/exceptions.c
- internal/native/runtime/features.c
- internal/native/runtime/node_crypto.c
- internal/native/runtime/node_fs_file.c
- internal/native/runtime/node_host.c
- internal/native/runtime/regexp_replace.c
- internal/native/runtime/sort.c
- internal/native/runtime/sort_undefined.c

Runtime cleared 66ad4a67 subject to the uncaught diagnostic condition. The follow-up below supplies that diagnostic; integration acceptance remains pending. Delivery branch: compiler/chain-slice-6. No PR and no main/area push.

Owning repairs were folded into one commit per member on current main. The tested tree is unchanged: 65191d9656fb6dbcfcefd550e94d692eb4aecd24 before and after history consolidation. See ../chain-slice-6/history-proof.json.
