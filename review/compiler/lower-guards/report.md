# Lowering mutant guards

Task #agwccbw: replayed all three exact audit diffs against targeted downstream tests before deciding which guard to add. M14 and M15 are already held by the differential oracle, so no duplicate lowering tests were added for them. M11 survived the targeted oracle, flow and freshness checks; one lowering guard and its reduced `.a` fixture were added. This contributes a bounded test leaf to step 79.

Base: `470cc0cde025aee921b7c7856bebe2465c1461b1`, current origin/main when this branch started. No production changes are committed. Each mutated production file was restored from its original bytes in a `finally` block.

## Existing downstream guards

| Mutant | Existing guard | What catches it | Mutated leaf seconds |
| --- | --- | --- | ---: |
| u038 M14, readiness declaration starts true | `TestNativeAgreesWithNode/internal/oracle/testdata/namespaces_unknown_before.a` | Node gives the namespace-property TypeError; both native and JavaScript give a ReferenceError instead. The exact stderr comparison fails. | 0.41 |
| u038 M15, completion assignment writes false | `TestNativeAgreesWithNode/internal/oracle/testdata/namespaces_safe_initialization.a` | Node exits 0 with the complete output; both native and JavaScript exit 70 after the first line. The exit/output comparisons fail. | 0.44 |

These are ordinary existing oracle fixture checks, not the oracle's own planted-mutant tests. Both exact production diffs exited 1 under those checks. The baseline and restored source passed. Oracle observations were uncached with `ADAMIC_GATE_UNCACHED=1`; the harness ran source Node, JavaScript, sanitized native and release native.

The restored namespace leaves took 0.44 and 0.47 seconds respectively. Existing `TestNewExpressionCacheCaptureRemainsNotYet` also passed, taking 0.24 seconds.

## M11 needs the new lowering guard

The exact M11 diff changes `len(value.Arguments) == 0` to `<= 1`. Existing constructor-cache oracle fixtures and `TestNewExpressionCacheCaptureRemainsNotYet` still passed under it. The two flow and three freshness corpus tests for existing constructor-cache fixtures also passed under M11, as they did at baseline. Their selectors and full outputs are recorded in `downstream-results.json`, `flow-fresh-results.json` and the matching `.jsonl` files.

`TestConstructorCacheOneArgumentRemainsNotYet` reads `internal/lower/testdata/constructor_cache_one_argument.a`, loads it through the checker, and pins the exact additional-capture NotYet reason. The fixture is reduced from the audit's `cache-one-argument.a`: a local constructor cache initialized through an ordinary function with one argument. Source Node prints `9`.

| New test run | Exit | Leaf seconds | Whole go-test command seconds |
| --- | ---: | ---: | ---: |
| baseline | 0 | 0.03 | 1.92 |
| mutant | 1 | 0.03 | 1.97 |
| restored | 0 | 0.03 | 1.93 |

The mutant is killed by the new assertion: `want additional-capture NotYet for a one-argument initializer, got <nil>`. It is not a load error, build warning or timeout. The initial reduction printed a number directly, which failed the prelude's console string parameter check. Those initial invalid-fixture runs are preserved with the `initial-invalid-` prefix and are excluded from mutant proof. Printing the number's string representation fixed the reduction before the valid baseline/mutant/restored proof.

Every new test runs in parallel, has no subtests and completes under 60 seconds including its load/lowering setup. The measured whole process includes Go driver/test-binary preparation with the Go cache prepared by setup; `-count=1` bypasses test result caching. Every Go run used `-timeout 90s` and an outer 180-second command limit. `GOMAXPROCS=4` held the final measurements to the CPU quota.

## Commands and evidence

The exact diffs are `M14.diff`, `M15.diff`, and `M11.diff` from the cited audit branches. They were applied one at a time with `git apply`, restored, and the unmutated checks rerun.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/namespaces_unknown_before[.]a$' \
  -count=1 -timeout 90s -json
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/namespaces_safe_initialization[.]a$' \
  -count=1 -timeout 90s -json
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^(TestNewExpressionCacheCaptureRemainsNotYet|TestNativeAgreesWithNode)$/internal/oracle/testdata/new_expression_class_cache_(local|capture_notyet)[.]a$' \
  -count=1 -timeout 90s -json
go test ./internal/flow ./internal/fresh \
  -run '^Test(FlowProgram|FreshWrites).*new_expression_class_cache' \
  -count=1 -timeout 90s -json
go test ./internal/lower -run '^TestConstructorCacheOneArgumentRemainsNotYet$' \
  -count=1 -timeout 90s -json
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestCountsAreRecorded$' -count=1 -timeout 90s -args -update-counts
```

Counts regeneration passed in 44.447 seconds; `internal/oracle/counts.md` did not change. The new lowering-only fixture creates no differential-oracle count row. `git diff --check` passed. Commands and return codes are recorded in the result JSON; full test output was written to files rather than piped through a truncating reader.

Source Node was checked with a bounded invocation of Node's `node:module.stripTypeScriptTypes` over the `.a` fixture, followed by evaluation. It exited 0 and printed `9` (`source-node.log`).

## Setup and limits

`export GOPROXY='https://proxy.golang.org|direct'` preceded `timeout 180 bash cloud/setup.sh`. Setup succeeded and printed `/workspace/adamic-tools/env.sh`, sourced for all Go runs. Timing lines: Go and Node ready 0.024 seconds; markdown dependencies ready 0.077; submodules ready 0.081; clang ready 0.179; Go build ready 39.461; test binaries deferred 39.758; build cache warm 39.760; done 39.786. `nproc` printed 5, with `cpu.max` `400000 100000` (four-CPU quota). Go 1.27.1, clang 20.1.8, Node 24.19.0.

This was targeted downstream replay, not a claim about all repo-wide coverage. Full packages, the full gate and stage1 port suites were not run. Integration lane results are recorded alongside this report. All branch changes are `_test.go`, testdata or review evidence.

Lane checks passed: `lane checks 1.4 s: gofmt and tools on 1 Go files, t.Parallel on 1 test packages; vet 1 packages`. Origin/main advanced to `6d890e0129beecb2df9f65c89b7733a6f02b5248` during checks and was merged without conflicts. Its changes were stage1 CSS and lint tests, with no production change to the guarded paths. The new lowering guard and both namespace oracle cases were rerun green on the merged tip (`merged-guard.jsonl`, `merged-downstream.jsonl`).

The first commit attempt failed because the disk was full. Only disposable Go caches from this worker's completed bridge archive measurement were removed, preserving its source and evidence; the commit then succeeded.
