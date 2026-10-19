Built: re-cut Slice 3 as one commit on the newest main fetched on resumption, retaining both refusal checks.
Commit: delivery branch compiler/loom-r-slice-3-ac-m2; base 638f45766812bc57b7b8a74c3dcd2c806d7a1993; full delivery SHA is in the final response.
Commands and outputs: exact commands, attempts and test timings are listed below and in results.jsonl.
Mutants: 14/14 production mutants caught; removed-header and wrong-reason mutants rejected.
Not certified: the blanket 60-second grain requirement fails on the existing TestProduct_GrainMarkdownFormatter (97.60s) and TestProduct_MarkdownblocksTextFormatter (130.04s) cold builds. Neither was rerun. No ahra, Loom submission, full repository gate or macOS execution; skipped cases remain skips.

## Base, conflicts and assumptions

Source: a4b12a2c463e05ee612c064caa5d8db1c8486cf8, applying only its own parent-to-commit change. Initial main was 561aeb4b851ca0820f39440a09dc5fb3fd27ba40. The resume fetch returned 638f45766812bc57b7b8a74c3dcd2c806d7a1993; that newer main is the delivery base. Assumption: hold this fetched main snapshot throughout validation rather than continually chasing main. Its eep presence checks, tests and witnesses are retained unchanged. Every changed program is byte-identical to the source commit.

One refusal-pass hunk conflicted: main called knownLoweringGap(node), while Slice 3 called functionCallRefusal(node) at the same position. Retained both error-return blocks, with main's known-gap check first. Also retained main's literal-callable branch selection and the source's module-level merged-field and Function-annotation checks. refusal-resolution.patch records the final resolution. No newer main line was reverted by the resolution.

counts.md conflicted. Started from the final main table and regenerated it on Linux with house clang 20.1.8, never hand-merged. The new-main automatic merge retains its exception, expression, emission and runtime-header additions together with Slice 3. Private unreferenced validation snapshots make corpus inputs clean against HEAD; they are not delivery commits. The final branch contains exactly one commit above the final main base.

Completed IR, JavaScript, lower and flow runs on the earlier base are preserved separately under earlier-base-561aeb4b and are not credited as final-base coverage. Its partial native product run was interrupted to incorporate the newly fetched main. No completed over-budget leaf was rerun. The requested complete recursive validation takes priority over the older 30-minute first-green workflow.

## Runtime clearance

Extra changed lines in internal/native/runtime/*.c or *.h beyond the source commit's own delta: NONE. Added and removed lines match by file in adamic.h, closure.c and union.c. runtime-clearance.json.txt and the two runtime patches retain the exact comparison. Main's newer header declarations and host-runtime include remain intact.

## Counts

Final-base command: timeout 840 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -json -parallel=4 -timeout=13m -args -update-counts. Complete output: counts-update.jsonl. Regenerated once on the final base, after a separate successful regeneration on the superseded base. All 33 changed rows belong to the source change and exactly equal its recorded values: 22 additions, 11 existing rows. No unrelated row changed.

| Fixture | Before (alloc/free/retain/release/peak/region) | After |
|---|---|---|
| internal/load/testdata/0.1/compile/09_tree.ts | 19/19/28/69/16/0 | 33/33/66/69/16/0 |
| internal/oracle/testdata/047cb0d_n_arrayindex.a | 1/1/1/2/1/0 | 2/1/4/4/1/0 |
| internal/oracle/testdata/047cb0d_n_element.a | 4/3/6/8/4/0 | 5/3/9/10/4/0 |
| internal/oracle/testdata/047cb0d_n_element_method.a | 4/3/6/9/4/0 | 5/3/9/11/4/0 |
| internal/oracle/testdata/047cb0d_n_element_plain.a | 4/3/5/8/4/0 | 5/3/8/10/4/0 |
| internal/oracle/testdata/call_spread_array.a | new | 56/56/112/115/14/0 |
| internal/oracle/testdata/element_access_presence.a | 4/3/5/7/4/0 | 5/3/8/10/4/0 |
| internal/oracle/testdata/json_shorthand.a | new | 20/20/8/29/4/0 |
| internal/oracle/testdata/narrowed_fields.a | 4/3/4/6/4/0 | 5/3/7/8/4/0 |
| internal/oracle/testdata/narrowed_methods.a | 3/3/3/6/3/0 | 4/3/6/8/3/0 |
| internal/oracle/testdata/narrowed_reads.a | 5/5/3/7/4/0 | 6/5/6/9/4/0 |
| internal/oracle/testdata/optional_after_call_193.a | new | 3/3/6/9/3/0 |
| internal/oracle/testdata/optional_after_call_79.a | new | 4/4/6/10/4/0 |
| internal/oracle/testdata/optional_after_call_gaps/computed.a | new | 3/3/1/6/2/0 |
| internal/oracle/testdata/optional_after_call_propagation.a | new | 18/18/13/20/7/0 |
| internal/oracle/testdata/optional_after_call_required.a | new | 3/3/5/7/2/0 |
| internal/oracle/testdata/optional_after_call_variants.a | new | 11/11/10/21/3/0 |
| internal/oracle/testdata/refusal_rulings/function_length.a (TypeScript mode) | new | 20/20/4/29/13/0 |
| internal/oracle/testdata/refusal_rulings/function_signature.a | new | 3/3/0/3/3/0 |
| internal/oracle/testdata/refusal_rulings/function_unused.a (TypeScript mode) | new | 0/0/0/0/0/0 |
| internal/oracle/testdata/refusal_rulings/merge_missing.a (TypeScript mode) | new | 1/0/1/1/1/0 |
| internal/oracle/testdata/refusal_rulings/merge_missing_computed.a (TypeScript mode) | new | 1/0/1/1/1/0 |
| internal/oracle/testdata/refusal_rulings/merge_present.a | new | 3/3/1/4/3/0 |
| internal/oracle/testdata/refusal_rulings/merge_present.a (TypeScript mode) | new | 3/3/1/4/3/0 |
| internal/oracle/testdata/refusal_rulings/merge_unused.a (TypeScript mode) | new | 0/0/0/0/0/0 |
| internal/oracle/testdata/regexp_null_narrowed.a | 4/4/8/5/3/0 | 5/4/11/7/3/0 |
| internal/oracle/testdata/reuse_narrowed.a | 7/5/6/8/5/0 | 8/5/9/10/5/0 |
| stage3/fixtures/assignment-proofs/01_scanner_keyword.a | new | 18/18/18/32/6/0 |
| stage3/fixtures/assignment-proofs/02_if_narrowing.a | new | 6/6/5/13/3/0 |
| stage3/fixtures/assignment-proofs/03_return_defined.a | new | 8/8/6/18/4/0 |
| stage3/fixtures/assignment-proofs/04_chained.a | new | 4/4/6/15/3/0 |
| stage3/fixtures/assignment-proofs/05_compound.a | new | 8/8/0/8/4/0 |
| stage3/fixtures/assignment-proofs/06_wider_target.a | new | 4/4/0/4/4/0 |

## Tests, reader guard and vet

Go runs use GOMAXPROCS=4 under the four-CPU cgroup quota; the recursive suite and mutant drivers also pin four-CPU affinity. Product owners run first with -parallel=1. Recursive consumer suites use -skip ^TestProduct_ -parallel=4 -p 1, so the product leaves are covered by their preceding owner runs and are not rerun. Child oracles packages have their own preceding product runs. Each command has an outer hard limit; tests also have a Go timeout. Output goes directly to logs. Results include setup and every failed attempt, if any.

| Command label | Exit | Wall seconds | Failed test events |
|---|---:|---:|---:|
| internal-ir-products | 0 | 105.265 | 0 |
| internal-ir | 0 | 12.622 | 0 |
| internal-javascript-products | 0 | 0.822 | 0 |
| internal-javascript | 0 | 0.828 | 0 |
| internal-lower-products | 0 | 6.176 | 0 |
| internal-lower | 0 | 73.214 | 0 |
| internal-flow-products | 0 | 4.89 | 0 |
| internal-flow | 0 | 99.102 | 0 |
| internal-native-products | 0 | 78.713 | 0 |
| internal-native | 1 | 492.588 | 3 |
| internal-oracle-products | 0 | 539.248 | 0 |
| internal-oracle | 0 | 495.856 | 0 |
| stage1-cohere-cssstrings-owners | 0 | 22.96 | 0 |
| stage1-cohere-cssstrings-products | 0 | 16.688 | 0 |
| stage1-cohere-cssstrings | 0 | 34.0 | 0 |
| stage1-cohere-json-owners | 0 | 66.619 | 0 |
| stage1-cohere-json-products | 0 | 42.292 | 0 |
| stage1-cohere-json | 0 | 503.064 | 0 |
| stage1-cohere-markdownblocks-owners | 0 | 314.275 | 0 |
| stage1-cohere-markdownblocks-products | 0 | 531.569 | 0 |
| stage1-cohere-markdownblocks | 0 | 498.332 | 0 |
| stage1-cohere-selector-owners | 0 | 9.418 | 0 |
| stage1-cohere-selector-products | 0 | 14.147 | 0 |
| stage1-cohere-selector | 0 | 19.829 | 0 |
| stage1-cohere-yaml-owners | 0 | 14.088 | 0 |
| stage1-cohere-yaml-products | 0 | 286.426 | 0 |
| stage1-cohere-yaml | 0 | 82.9 | 0 |
| stage1-typescript-parser-owners | 0 | 9.167 | 0 |
| stage1-typescript-parser-products | 0 | 319.312 | 0 |
| stage1-typescript-parser | 0 | 49.152 | 0 |
| stage1-typescript-scanner-owners | 0 | 4.948 | 0 |
| stage1-typescript-scanner-products | 0 | 38.062 | 0 |
| stage1-typescript-scanner | 0 | 29.971 | 0 |
| stage3-fixtures-products | 0 | 31.452 | 0 |
| stage3-fixtures | 0 | 73.68 | 0 |
| reader-guard | 0 | 3.035 | 0 |
| vet | 0 | 13.013 | 0 |
| internal-native-repaired | 0 | 301.597 | 0 |
| production-mutants | 0 | 237.107 | 0 |
| assignment-proof-driver | 0 | 36.135 | 0 |
| reader-guard-final | 0 | 3.226 | 0 |
| vet-final | 0 | 8.962 | 0 |

Final-base recorded test events: {'pass': 13549, 'skip': 511, 'fail': 3}. Maximum in source-touched test files: 26.770s. touched-test-seconds.json.txt lists every such test and its seconds. All timing events at or above 60s: [{"package": "github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles", "test": "TestProduct_GrainMarkdownFormatter", "action": "pass", "seconds": 97.6}, {"package": "github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles", "test": "TestProduct_MarkdownblocksTextFormatter", "action": "pass", "seconds": 130.04}].

Exact commands:

```text
timeout 840 go test ./internal/ir -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./internal/ir/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./internal/javascript -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./internal/javascript/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./internal/lower -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./internal/lower/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./internal/flow -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./internal/flow/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./internal/native -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./internal/native/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./internal/oracle -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./internal/oracle/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/cssstrings/oracles -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/cssstrings -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/cssstrings/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/json/oracles -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/json -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/json/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/markdownblocks/oracles -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/markdownblocks -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/markdownblocks/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/selector/oracles -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/selector -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/selector/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/yaml/oracles -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/yaml -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/cohere/yaml/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./stage1/typescript/parser/oracles -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/typescript/parser -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/typescript/parser/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./stage1/typescript/scanner/oracles -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/typescript/scanner -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage1/typescript/scanner/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 840 go test ./stage3/fixtures -run ^TestProduct_ -count=1 -json -parallel=1 -timeout=13m -p 1
timeout 840 go test ./stage3/fixtures/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 120 go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -timeout=90s -json
timeout 180 go vet -p 1 ./internal/ir/... ./internal/javascript/... ./internal/lower/... ./internal/flow/... ./internal/native/... ./internal/oracle/... ./stage1/cohere/cssstrings/... ./stage1/cohere/json/... ./stage1/cohere/markdownblocks/... ./stage1/cohere/selector/... ./stage1/cohere/yaml/... ./stage1/typescript/parser/... ./stage1/typescript/scanner/... ./stage3/fixtures/...
timeout 840 go test ./internal/native/... -skip ^TestProduct_ -count=1 -json -parallel=4 -timeout=13m -p 1
timeout 1440 python3 /workspace/adamic/review/compiler/loom-r-slice-3-ac-m2/run-mutants.py
timeout 240 python3 stage3/fixtures/assignment-proofs/check.py --require-adamic
timeout 120 go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -timeout=90s -json
timeout 180 go vet -p 1 ./internal/ir/... ./internal/javascript/... ./internal/lower/... ./internal/flow/... ./internal/native/... ./internal/oracle/... ./stage1/cohere/cssstrings/... ./stage1/cohere/json/... ./stage1/cohere/markdownblocks/... ./stage1/cohere/selector/... ./stage1/cohere/yaml/... ./stage1/typescript/parser/... ./stage1/typescript/scanner/... ./stage3/fixtures/...
```

## Production mutants

Every overlay source is .go.txt; each negative run must fail an actual test and excludes Go/C build errors and test timeouts as catchers. The source catalog supplies all 14 mutations. Embedded assignment, call-spread and JSON-binding source mutants are also covered by the oracle suite.

| Mutant | Catcher | Caught | Seconds |
|---|---|---|---:|
| check-optional-receiver | ^TestOptionalAfterCall79$ | True | 18.011 |
| javascript-required-panic | ^TestOptionalAfterCallRequired$ | True | 11.601 |
| native-required-panic | ^TestOptionalAfterCallRequired$ | True | 12.666 |
| drop-flow-throw-edge | ^TestOptionalAfterCallFlowEdges$ | True | 20.048 |
| drop-required-propagation | ^TestOptionalAfterCallPropagation$ | True | 19.138 |
| merged-computed-read | ^TestMergedFieldTypeScriptComputedChecked$ | True | 21.418 |
| merged-declaration | ^TestMergedFieldAdamicUnusedRefused$ | True | 15.312 |
| merged-read | ^TestMergedFieldTypeScriptMissingChecked$ | True | 22.687 |
| function-annotation | ^TestFunctionAdamicUnusedAnnotationRefused$ | True | 16.275 |
| function-call | ^TestFunctionTypeScriptCallRefused$ | True | 16.142 |
| function-length | ^TestFunctionTypeScriptLengthAgreesWithNode$ | True | 22.627 |
| erase-earlier-throw | ^TestOptionalAfterCallPreservesEarlierThrow$ | True | 10.881 |
| tsgo-absolute-main | ^TestTSGoCacheSanitized$ | True | 15.594 |
| tsgo-rewrite-runtime | ^TestTSGoCacheRelease$ | True | 14.477 |

## A-check

All 22 source-added or source-modified .a files were compiled in place using the final branch compiler. The unchanged pinned Gate.aCheck predicate passes all 22: 16 checked, six refused, zero NotYet, zero checker errors. No header or program body needed changing on this re-cut. program-byte-check.json.txt records the source-byte comparison. Both removed-header and wrong-reason mutants on refusal_rulings/function_call.a stay genuinely refused and make the predicate exit 1; mutants.json.txt records the catches.

| File | Status | Reason/code |
|---|---|---|
| internal/oracle/testdata/call_spread_array.a | checked |  |
| internal/oracle/testdata/json_shorthand.a | checked |  |
| internal/oracle/testdata/optional_after_call_193.a | checked |  |
| internal/oracle/testdata/optional_after_call_79.a | checked |  |
| internal/oracle/testdata/optional_after_call_gaps/computed.a | checked |  |
| internal/oracle/testdata/optional_after_call_propagation.a | checked |  |
| internal/oracle/testdata/optional_after_call_required.a | checked |  |
| internal/oracle/testdata/optional_after_call_variants.a | checked |  |
| internal/oracle/testdata/refusal_rulings/function_call.a | refused | the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number |
| internal/oracle/testdata/refusal_rulings/function_length.a | refused | the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number |
| internal/oracle/testdata/refusal_rulings/function_signature.a | checked |  |
| internal/oracle/testdata/refusal_rulings/function_unused.a | refused | the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number |
| internal/oracle/testdata/refusal_rulings/merge_missing.a | refused | class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view |
| internal/oracle/testdata/refusal_rulings/merge_missing_computed.a | refused | class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view |
| internal/oracle/testdata/refusal_rulings/merge_present.a | checked |  |
| internal/oracle/testdata/refusal_rulings/merge_unused.a | refused | class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view |
| stage3/fixtures/assignment-proofs/01_scanner_keyword.a | checked |  |
| stage3/fixtures/assignment-proofs/02_if_narrowing.a | checked |  |
| stage3/fixtures/assignment-proofs/03_return_defined.a | checked |  |
| stage3/fixtures/assignment-proofs/04_chained.a | checked |  |
| stage3/fixtures/assignment-proofs/05_compound.a | checked |  |
| stage3/fixtures/assignment-proofs/06_wider_target.a | checked |  |

## Toolchain and evidence

GOPROXY=https://proxy.golang.org|direct was exported before bash cloud/setup.sh, then /workspace/adamic-tools/env.sh was sourced. nproc=5; cgroup quota 400000/100000 is four CPUs. Go 1.27.1, Node 24.19.0, Linux amd64, clang 20.1.8. setup.log retains all output. Setup timing lines:

```text
setup: go ready (0.039s)
setup: node ready (0.046s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.011s
setup: markdown dependencies ready (0.120s)
setup: submodules ready (0.143s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.447s)
setup: shared cache ready (1.817s)
setup: go build ready (71.792s)
setup: test binaries deferred (use --warm-tests) (71.949s)
setup: build cache warm (71.951s)
setup: build-flags commit=561aeb4b851ca0820f39440a09dc5fb3fd27ba40 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.00 0.00 1/362 155013 load-after=2.71 0.76 0.26 1/369 156038
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (71.998s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.Z53NQB
```

Existing Go cache data was reused in the task-local cache. Only byte-verified identical export archives were hard-linked to bound disk use; product names and bytes were preserved. Cache records and deduplication logs are retained. Reclaimed 3,259,256,352 bytes of stale pre-unit Adamic compiler archives, identified by their Go compilation-unit package symbols; current-unit archives and Cohere/toolchain dependencies were retained. The complete eviction inventory is saved. Product output and result caches use task-local scratch; no machine path was added to products or tests. Historical source review artifacts remain unchanged, including their pre-existing patch/log whitespace. New evidence uses .json.txt so it does not alter the JSON corpus census. No code was copied from cohere.

The pinned JSON corpus was provisioned through its unchanged TestProduct_JSONPinnedCorpus in an isolated worktree at the same validation snapshot. This avoids matching historical /workspace/adamic strings against the current builder path; no guard or corpus byte was changed. The ordinary checkout then reads the identical product (0.27s). json-relocated-product.log and json-portable-product-read.log record both runs. The temporary worktree referenced the existing cohere submodule and was removed after the successful read. npm ci --prefix stage3/api --ignore-scripts provisioned the exact lockfile dependencies for the corpus and stock TypeScript 6.0.3 proof driver.

The original commit message is preserved, including its historical WIP note. This re-cut independently reproduced the existing markdown producer grain failure: TestProduct_GrainMarkdownFormatter took 97.60s and TestProduct_MarkdownblocksTextFormatter took 130.04s while cold-building their standalone Cohere commands. Neither was rerun and no unrelated producer was weakened or edited. Deliver the requested branch as reviewable code, but hold it from Loom until the owner splits or prepares that producer within the budget. Functional pass is not claimed to satisfy this timing requirement.

The first native recursive consumer run failed only on toolchain setup: GCC 14.2.0 lacked sanitizer/allocator_interface.h, and the unguarded attributed-binding WASI leaf had no WASI_SYSROOT. No leaf exceeded 60 seconds in that attempt. The GCC ASan library exports both required allocator functions; the task-local GCC wrapper supplies only the house LLVM header and adds its SHA256 to the compiler identity. gcc-header-workaround.json.txt and the successful ABI probe record it. No sudo is available. cloud/setup.sh --wasi-sdk provisioned SDK 27 with clang 20.1.8 and the supported sysroot. setup-wasi-sdk.log retains its full timings. The repaired-toolchain native rerun is separately recorded rather than hiding the first attempt.

All evidence belongs to review/compiler/loom-r-slice-3-ac-m2/. No PR, ahra CLI, Loom submission, main push or force push was performed.
