# Async control flow validation

Main landed the identical async implementation as `926e19f4967c0f13a2b9a58560dc1eb61601d9c3` during validation. This branch commit carries the flow budget split, exact refusal assertions, and completed evidence. This worker pushed only its own branch and made no Loom submission.

Tests ran against the code saved as `935545154d9561235069be8fbf43371b10a2a6e6`. The final recut retains identical compiler, runtime, fixtures, records, and tested refinements.

| Check | Result | Package seconds |
| --- | --- | ---: |
| flow | pass | 79.417 |
| fresh | pass | 22.497 |
| lower | pass | 55.177 |
| native | pass; existing skip: TestNormalizeLongMeasurements, TestRecordBenchmark, TestMeasureClangUnits, TestWASIService, TestWASIServiceCounted, TestWASIUnit00, TestWASIUnit01, TestWASIUnit02, TestWASIUnit03, TestWASIUnit04, TestWASIUnit05, TestWASIUnit06, TestWASIUnit07, TestWASIUnit08, TestWASIUnit09, TestWASIUnit10, TestWASIUnit11, TestWASIUnit12, TestWASIUnit13, TestWASIUnit14, TestWASIUnit15, TestWASIUnit16, TestWASIUnit17, TestWASIUnit18, TestWASIUnit19, TestWASIUnit20, TestWASIUnit21, TestWASIUnit22, TestWASIUnit23, TestWASIUnit24, TestWASIUnit25, TestWASIUnit26, TestWASIUnit27, TestWASIUnit28, TestWASIUnit29, TestWASIUnit30, TestWASIUnit31, TestWASIUnit32, TestWASIUnit33, TestWASIUnit34, TestWASIUnit35, TestProduct_StrictWASIRuntime, TestProduct_WASIServiceNative, TestProduct_WASIServiceModules, TestWASIServiceMutantShrunkCount, TestWASIServiceMutantUnicodeEscape, TestWASIServiceMutantRetainedObject, TestWASIServiceMutantResponseByte, TestWASIServiceNativeAgrees_7, TestWASIServiceNativeAgrees_6, TestWASIServiceNativeAgrees_5, TestWASIServiceNativeAgrees_4, TestWASIServiceNativeAgrees_3, TestWASIServiceNativeAgrees_2, TestWASIServiceNativeAgrees_1, TestWASIServiceNativeAgrees_0, TestWasiAsyncHostPromises, TestDecodeASCIIWASIUnit44, TestDecodeASCIIWASIUnit42, TestDecodeASCIIWASIUnit43, TestDecodeASCIIWASIUnit41, TestDecodeASCIIWASIUnit40, TestDecodeASCIIWASIUnit39, TestDecodeASCIIWASIUnit38, TestDecodeASCIIWASIUnit37, TestDecodeASCIIWASIUnit36, TestDecodeASCIIWASIUnit35, TestDecodeASCIIWASIUnit34, TestDecodeASCIIWASIUnit33, TestDecodeASCIIWASIUnit32, TestDecodeASCIIWASIUnit31, TestDecodeASCIIWASIUnit30, TestDecodeASCIIWASIUnit29, TestDecodeASCIIWASIUnit28, TestDecodeASCIIWASIUnit27, TestDecodeASCIIWASIUnit26, TestDecodeASCIIWASIUnit25, TestDecodeASCIIWASIUnit24, TestDecodeASCIIWASIUnit23, TestDecodeASCIIWASIUnit22, TestDecodeASCIIWASIUnit21, TestDecodeASCIIWASIUnit20, TestDecodeASCIIWASIUnit19, TestDecodeASCIIWASIUnit18, TestDecodeASCIIWASIUnit17, TestDecodeASCIIWASIUnit16, TestDecodeASCIIWASIUnit15, TestDecodeASCIIWASIUnit14, TestDecodeASCIIWASIUnit13, TestDecodeASCIIWASIUnit12, TestDecodeASCIIWASIUnit11, TestDecodeASCIIWASIUnit10, TestDecodeASCIIWASIUnit09, TestDecodeASCIIWASIUnit08, TestDecodeASCIIWASIUnit07, TestDecodeASCIIWASIUnit06, TestDecodeASCIIWASIUnit05, TestDecodeASCIIWASIUnit04, TestDecodeASCIIWASIUnit03, TestDecodeASCIIWASIUnit02, TestDecodeASCIIWASIUnit01, TestDecodeASCIIWASIUnit00, TestProduct_DecodeASCIIWASI | 239.788 |
| oracle | pass; existing skip: TestStage3FixtureHook | 758.656 |
| oracle jobserver | pass | 11.855 |
| count regeneration | pass | 72.656 |
| acheck | pass | 6.384 |
| stage3 fixtures | pass; existing skip: TestTransformedNodeRunnerGuardHook | 74.542 |
| gate | pass | 0.749 |
| meter | pass | 1.499 |
| compiler census | pass | 0.814 |
| exact refusals | pass | 3.636 |
| read only | pass | 59.989 |
| async TSan | pass | 6.089 |
| host promise TSan | pass | 9.103 |

Whole oracle recorded 2073 top-level passes and one existing fixture-hook skip; the isolated jobserver proof adds one pass. All requested package units are covered. No final reading unit exceeded 30 seconds. The main-owned sanitized and WASI request product builders previously took 79.44 and 61.99 seconds; these run on Workshop outside the unit budget.

All 19 stage1 gap files passed. The census checked 786 promises with zero regressions: one async admission, six inherited admissions and three inherited refusal-reason changes. Full paths and diagnostics are in `recorded-outcome-summary.json`.

There are 73 batch count additions and 40 inherited additions since batch 13: 1474 records become 1587. Every existing record is unchanged. Every added record and value appears in `count-changes.json`.

All 41 traced owners reported zero refused operations. All 353 readers in `read-only-tests.json` passed with `ADAMIC_BUILD_CACHE=read` and `ADAMIC_GATE_UNCACHED=1`. The first attempt exposed JSONTypes missing from the proof preparation plan. Tracing its existing owner fixed that prerequisite.

| Production mutation | Named red test | Evidence |
| --- | --- | --- |
| early-live-collection-release | `TestAsyncLoopsForOfMutation` | heap-use-after-free |
| skip-iterator-cleanup-on-break | `TestAsyncLiveIterableMapBreakFunction` | LeakSanitizer |
| labeled-break-skips-finally | `TestAsyncLabelsAsyncLabelsFinally` | JavaScript: stdout differs |

The iterator mutant omits builtin ownership cleanup. User iterators and their `return()` dispatch remain exactly refused; builtin iterators do not implement that method.

ThreadSanitizer passed the accepted async worker-pool fixture, live Map clearing in function and module frames, labeled suspended finally, and native host promises. The temporary four-test harness is `async-tsan-proof.go.txt`; the fixture inventory is `async-worker-fixtures.json`.

The initial flow remainder took 40.34 seconds. Four deterministic top-level shards preserve its coverage under budget. Gate caught indirect Parallel calls; direct first calls fixed them. Exact-refusal tests compare full diagnostics and locations, and fail if lowering succeeds.

`validation-results.json` records commands, exits, timings, traced owners and corrected attempts. `validation-logs.tar.gz` preserves 1003 raw logs, including initial failures. Mutant patches and decoded diagnostics are under `mutants/`. Earlier bounded oracle inventory logs refer to the preceding JSON base; `recut-oracle.log` is the final whole run.
