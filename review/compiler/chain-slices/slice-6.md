# Chain slice 6

Built four own-net members on main, with catchable initialization throws and matching lint/runtime features.
Member commits: de2f1819, 1e9199d2, 090163eb, 250a56be; owning repairs and the main merge remain in history.
Lower: 277 pass, two existing skips; own fixtures, feature matrix, reader guard, counts and lane checks pass.
49 mutant cases were caught; each case and catcher is listed below with evidence.
Not covered: literal all-admission Node agreement, WASI execution, other platforms or the full gate.

## Extraction and dependencies

Initial main: 5e33a17b186a8a2218d27b69b21e2de5acc5b750. Current main merged: 98008bbba2881939e07a9d74031994102a3bbe36. The chain head f5236b48 was inspected, never merged. No other slice was merged or copied. The source plan, refined ranges and complete source histories are preserved in ../chain-slice-6/.

| Member | Source SHA | Fork point | Decision | Dependency evidence | Tasks served |
| --- | --- | --- | --- | --- | --- |
| exceptions-21-main | 205586a0899488595200f3fca470874ef9282b37 | 7a10c877 | kept | This slice supplies Read.Throws/Assign.Throws at internal/ir/readiness.go:4 and propagation at internal/lower/readiness.go:56. Main supplies local readiness. No required outside production symbol. | step 21; #dv99xzy, #agwccbw |
| inherit-guards | 16a0b626b14b8fd6b541edc6cbb92e63b8ac55f5 | 291ee604 | kept | Main already supplies the void/never guard at internal/lower/class_inheritance.go:188 and inherited layout at :237. Existing guards retained; regression evidence added. | #a898y40 regression coverage |
| feature-set-link-main | e1efb527a314d834fc6fb282108e992558048eca | 553ad06a | kept | Main supplies RuntimeLibraryForSource and sourceFlags at internal/native/library.go:32/:44. Feature relocation/definition and source-aware helper fixes are supplied here. | #hmab710; Outcome 32 prerequisite |
| lint-features | 3535387095d50f05d80e2717aed468843e59bc18 | 2a28375f | kept | Requires feature-set-link-main in this slice. SourceFlags export at internal/native/library.go:37; matching lint runtime/flags at stage1/cohere/lint/profile_compilation_main_test.go:357. | lint profile link repair |

No member dropped. The necessary lint/feature-link interaction is kept together. Tasks have local delivery evidence and await integration acceptance; this does not claim Outcome 32 itself complete.

Repair ownership: 71972e18 supplies lexical TDZ exceptions, ReferenceError and readiness propagation; a6cb5660 restores Node/JS/native cycle exit checks; c775dc61 supplies observation logging and its bypass mutant; aea87960 supplies the WASI request uncaught exit ruling. d8e6f59c exception payload/pending ownership is retained. Main's unrelated guards and behavior were kept when conflicts were resolved.

Excluded outside portions: Read.Unset, Defined.Throws and namespace readiness are absent from main and unnecessary for its lexical readiness model; absent-message-shape runtime changes belong to another member. The lint parser_construction_test.go hunk names an outside fixture absent on main. The 2b2b1095 named-outcome WASI hunk needs ruled-backend-outcomes (2391c655) for backendDisagreement at internal/oracle/wasi_test.go:76. That helper is absent on main; its named-stack/weak-stop fixtures were not copied. Main's disagreement helper already covers the exception-owned exit-1 contract. No slice 1-4 changes imported.

## Commands and observed results

Setup exported GOPROXY=https://proxy.golang.org|direct. First bounded cloud/setup.sh expired without diagnostic; retry with ADAMIC_GOCACHE_OFF=1 completed. Printed elapsed seconds: Go 0.180, Node 0.234, submodules 0.484, markdown 0.522, clang 1.265, shared cache off 1.289, build 402.140, deferred 402.302, cache warm 402.304, done 402.429. nproc=5; Go capped at four CPUs. Go 1.27.1, Node 24.19, clang 20.1.8. Missing pinned Node types were installed with npm ci --prefix stage3/api. The checker archive build expired once and completed on retry. Setup delays and proof work exceeded the requested first-green window.

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
| Lane | Required fetch/show lane-checks pipeline passed formatting, tools, parallel analyzer and a-check. Its vet budget expired; separate go vet on all changed test packages passed. Final lane also includes cmd/adamic request repair. | lane-checks.log, lane-checks-final.log, vet.log, vet-final.log |
| WASI request | go test ./cmd/adamic -run '^TestWASIRequestThrows$' -count=1 -json -timeout 90s: compiles and skips; WASI execution opt-in/SDK not configured. | request-wasi-optin.jsonl |

Test seconds are in test-seconds.json; largest member leaf observed 38.58 s. No new test leaf exceeds 60 s. The counts aggregate writer is an existing test. Unsuccessful setup, stale allowlist and missing-type attempts are retained and not credited as passes.

## Admission delta limitation

The main-overlay baseline and slice binaries were compared on all 1,293 .a files under internal/oracle/testdata. Fifteen newly admitted; zero admission regressions; zero census timeouts. The later main merge changes C metadata caching; frontend, IR, JS, Go dependencies and cohere revision are identical between the two main commits. Final-main backend witnesses were rerun. See admission-census.json and admission-delta.json.

Fourteen newly admitted ordinary witnesses agree with Node under the step-21 stdout/exit contract for uncaught exceptions. The fifteenth, step21_builtin_narrow_terminal.a, deliberately invalidates a nominal narrowing through a call. Both Adamic backends stop at the inserted proof check; raw Node continues. This is a checked-negative soundness witness, NOT raw Node agreement. Therefore the literal requested all-admission Node proof is not established. The finite corpus also cannot establish universal agreement for all possible programs. Observations and those limits are kept separate.

## Every mutant and catcher

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

Runtime clearance and integration acceptance remain pending. Delivery branch: compiler/chain-slice-6. No PR and no main/area push.
