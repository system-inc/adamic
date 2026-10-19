Re-cut slice 3 onto current main, preserving both sides of reuse.go.
Branch: compiler/loom-r-slice-3-ac-m3; base: fde41ee7d014e38b6d309f217630614507af2011; source: bb1cc531484bc1a10da05041b23bac330a69c6c9.
Native/lower recursive tests, counts regeneration, guards, vet, gofmt and a-check pass.
All 14 production mutants and both header mutants are caught by their intended checks.
Not covered: full repository gate, Loom submission and formatter reruns; two existing cold count-product leaves exceeded 60 seconds.

## Resolution and assumptions

The fetched origin/main tip is fixed for this verification and delivery. Applied only bb1cc531's own change using the requested merge-tree base and cherry-pick --no-commit. The single conflict is checkDefined in internal/native/reuse.go: keep slice 3's catchable TypeError branch and main's cArray(message) initializer in the invariant-panic fallback. Main's automatically merged closure constructors remain, with source_length initialization added. No other main line was reverted by resolution.

Counts were reset to main and regenerated once on Linux with house clang 20.1.8. Exactly 33 rows differ from main: 22 additions and 11 existing rows. All resulting moved-row values match bb1cc531. Every carried .a file remains byte-identical to bb1cc531. No header or pinned-position changes.

## Runtime lines

No additional runtime C/H lines beyond the source commit. The complete own delta remains exactly:

```
internal/native/runtime/adamic.h: + size_t source_length;
internal/native/runtime/closure.c: + closure->source_length = 0;
internal/native/runtime/union.c:
+ if (value != NULL && value->kind == adamic_kind_closure && strcmp(name, "length") == 0) {
+     return adamic_box_number((double)((adamic_closure *)value)->source_length);
+ }
```

source-runtime.patch and recut-runtime.patch have identical added/deleted lines; runtime-audit.txt records the comparison.

## Commands and results

Every Go test used -count=1 and a hard timeout, with direct log output. Four-CPU affinity and GOMAXPROCS=4 were used. Products preceded their ordinary consumers. Exact commands and timing events are in results.jsonl and the corresponding logs/seconds files.

| Check | Exit | Wall seconds | Notes |
|---|---:|---:|---|
| internal-native-products | 0 | 135.676 |  |
| internal-native | 0 | 486.953 |  |
| internal-lower-products | 0 | 18.487 |  |
| internal-lower | 0 | 78.126 |  |
| counts-products | 0 | 156.0 | TestProduct_CountMeasurements00 63.89s; TestProduct_CountMeasurements15 65.24s |
| counts-update | 0 | 95.759 |  |
| reader-guard | 1 | 48.66 |  Environment storage failure, repaired below. |
| reader-product-Native | 0 | 22.316 |  |
| reader-product-Lower | 0 | 20.557 |  |
| reader-product-Fresh | 0 | 14.125 |  |
| reader-product-Flow | 1 | 7.382 |  Environment storage failure, repaired below. |
| reader-product-repaired-Flow | 0 | 11.278 |  |
| reader-guard-repaired | 0 | 9.114 |  |
| oracle-guards | 0 | 26.97 |  |
| flow-guard | 0 | 3.906 |  |
| vet | 0 | 9.461 |  |
| a-check | 0 | 5.418 |  |
| header-mutants | 0 | 2.787 |  |
| production-mutants | 0 | 258.856 |  |

Formatting: gofmt.txt is empty; production git diff --check passes. Native/lower products and consumer leaves all stayed under 60 seconds. Two unmodified count-product leaves passed at 63.89 and 65.24 seconds (CountMeasurements00 and 15); neither was rerun or altered. No test was added, split or deleted in this re-cut. The user explicitly accepted existing formatter cold-product timings; those packages were not rerun.

The first reader run and the first serial flow product attempt failed only with no space left on device. Regenerable Go-export snapshots were evicted and byte-identical completed archives deduplicated. All four reader products and the reader/SortedMap guard then passed. The bounded second dedup attempt reached its 45-second limit after partial safe progress; a later inode-aware pass completed and reclaimed 3,034,699,730 duplicate bytes. No source workaround was made for storage.

Setup completed in 70.869s: node 0.036s; go 0.040s; submodules 0.110s; markdown 0.119s; clang 0.226s; WASI SDK 0.430s; shared cache 1.617s; go build 70.554s; test binaries deferred 70.753s; cache warm 70.755s. nproc=5; cgroup quota is four CPUs. Sourced /workspace/adamic-tools/env.sh. Go 1.27.1, Node 24.19.0, clang 20.1.8.

Native validation retained the prior task-local GCC wrapper: system GCC 14 lacks sanitizer/allocator_interface.h despite libasan exporting the required functions. The wrapper adds only the house LLVM allocator header and declares its digest in --version (94be5aa9568359b0a1e331836221d05713d6810c28c6ed688d2aea6f3861348d). WASI_SYSROOT points to the installed house SDK. Runtime sources and tests were not edited for either environment condition.

## Production mutants and catchers

| Mutant | Behavioral catcher |
|---|---|
| check-optional-receiver | TestOptionalAfterCall79 (5.00s) |
| javascript-required-panic | TestOptionalAfterCallRequired (4.07s) |
| native-required-panic | TestOptionalAfterCallRequired (4.39s) |
| drop-flow-throw-edge | TestOptionalAfterCallFlowEdges (0.04s) |
| drop-required-propagation | TestOptionalAfterCallPropagation (4.99s) |
| merged-computed-read | TestMergedFieldTypeScriptComputedChecked (4.79s) |
| merged-declaration | TestMergedFieldAdamicUnusedRefused (0.05s) |
| merged-read | TestMergedFieldTypeScriptMissingChecked (3.92s) |
| function-annotation | TestFunctionAdamicUnusedAnnotationRefused (0.04s) |
| function-call | TestFunctionTypeScriptCallRefused (0.05s) |
| function-length | TestFunctionTypeScriptLengthAgreesWithNode (5.94s) |
| erase-earlier-throw | TestOptionalAfterCallPreservesEarlierThrow (0.00s) |
| tsgo-absolute-main | TestTSGoCacheSanitized (1.81s) |
| tsgo-rewrite-runtime | TestTSGoCacheRelease (1.53s) |

All overlays affect actual production paths, are retained as .go.txt files, and failed by test assertions, without compile errors, timeout, OOM or storage errors. Header mutants removed the first-line refusal header or replaced its reason on function_call.a; both retained compiler status refused and failed the pinned a-check predicate with exit 1.

## A-check status

Counts: {"checked": 16, "refused": 6}. NotYet: 0. Type error: 0. Predicate exit: 0.

| File | Status |
|---|---|
| internal/oracle/testdata/call_spread_array.a | checked |
| internal/oracle/testdata/json_shorthand.a | checked |
| internal/oracle/testdata/optional_after_call_193.a | checked |
| internal/oracle/testdata/optional_after_call_79.a | checked |
| internal/oracle/testdata/optional_after_call_gaps/computed.a | checked |
| internal/oracle/testdata/optional_after_call_propagation.a | checked |
| internal/oracle/testdata/optional_after_call_required.a | checked |
| internal/oracle/testdata/optional_after_call_variants.a | checked |
| internal/oracle/testdata/refusal_rulings/function_call.a | refused |
| internal/oracle/testdata/refusal_rulings/function_length.a | refused |
| internal/oracle/testdata/refusal_rulings/function_signature.a | checked |
| internal/oracle/testdata/refusal_rulings/function_unused.a | refused |
| internal/oracle/testdata/refusal_rulings/merge_missing.a | refused |
| internal/oracle/testdata/refusal_rulings/merge_missing_computed.a | refused |
| internal/oracle/testdata/refusal_rulings/merge_present.a | checked |
| internal/oracle/testdata/refusal_rulings/merge_unused.a | refused |
| stage3/fixtures/assignment-proofs/01_scanner_keyword.a | checked |
| stage3/fixtures/assignment-proofs/02_if_narrowing.a | checked |
| stage3/fixtures/assignment-proofs/03_return_defined.a | checked |
| stage3/fixtures/assignment-proofs/04_chained.a | checked |
| stage3/fixtures/assignment-proofs/05_compound.a | checked |
| stage3/fixtures/assignment-proofs/06_wider_target.a | checked |

## Every moved counts row

| Fixture | Before | After |
|---|---|
| internal/load/testdata/0.1/compile/09_tree.ts | 19, 19, 28, 69, 16, 0 | 33, 33, 66, 69, 16, 0 |
| internal/oracle/testdata/047cb0d_n_arrayindex.a | 1, 1, 1, 2, 1, 0 | 2, 1, 4, 4, 1, 0 |
| internal/oracle/testdata/047cb0d_n_element.a | 4, 3, 6, 8, 4, 0 | 5, 3, 9, 10, 4, 0 |
| internal/oracle/testdata/047cb0d_n_element_method.a | 4, 3, 6, 9, 4, 0 | 5, 3, 9, 11, 4, 0 |
| internal/oracle/testdata/047cb0d_n_element_plain.a | 4, 3, 5, 8, 4, 0 | 5, 3, 8, 10, 4, 0 |
| internal/oracle/testdata/call_spread_array.a | new row | 56, 56, 112, 115, 14, 0 |
| internal/oracle/testdata/element_access_presence.a | 4, 3, 5, 7, 4, 0 | 5, 3, 8, 10, 4, 0 |
| internal/oracle/testdata/json_shorthand.a | new row | 20, 20, 8, 29, 4, 0 |
| internal/oracle/testdata/narrowed_fields.a | 4, 3, 4, 6, 4, 0 | 5, 3, 7, 8, 4, 0 |
| internal/oracle/testdata/narrowed_methods.a | 3, 3, 3, 6, 3, 0 | 4, 3, 6, 8, 3, 0 |
| internal/oracle/testdata/narrowed_reads.a | 5, 5, 3, 7, 4, 0 | 6, 5, 6, 9, 4, 0 |
| internal/oracle/testdata/optional_after_call_193.a | new row | 3, 3, 6, 9, 3, 0 |
| internal/oracle/testdata/optional_after_call_79.a | new row | 4, 4, 6, 10, 4, 0 |
| internal/oracle/testdata/optional_after_call_gaps/computed.a | new row | 3, 3, 1, 6, 2, 0 |
| internal/oracle/testdata/optional_after_call_propagation.a | new row | 18, 18, 13, 20, 7, 0 |
| internal/oracle/testdata/optional_after_call_required.a | new row | 3, 3, 5, 7, 2, 0 |
| internal/oracle/testdata/optional_after_call_variants.a | new row | 11, 11, 10, 21, 3, 0 |
| internal/oracle/testdata/refusal_rulings/function_length.a (TypeScript mode) | new row | 20, 20, 4, 29, 13, 0 |
| internal/oracle/testdata/refusal_rulings/function_signature.a | new row | 3, 3, 0, 3, 3, 0 |
| internal/oracle/testdata/refusal_rulings/function_unused.a (TypeScript mode) | new row | 0, 0, 0, 0, 0, 0 |
| internal/oracle/testdata/refusal_rulings/merge_missing.a (TypeScript mode) | new row | 1, 0, 1, 1, 1, 0 |
| internal/oracle/testdata/refusal_rulings/merge_missing_computed.a (TypeScript mode) | new row | 1, 0, 1, 1, 1, 0 |
| internal/oracle/testdata/refusal_rulings/merge_present.a | new row | 3, 3, 1, 4, 3, 0 |
| internal/oracle/testdata/refusal_rulings/merge_present.a (TypeScript mode) | new row | 3, 3, 1, 4, 3, 0 |
| internal/oracle/testdata/refusal_rulings/merge_unused.a (TypeScript mode) | new row | 0, 0, 0, 0, 0, 0 |
| internal/oracle/testdata/regexp_null_narrowed.a | 4, 4, 8, 5, 3, 0 | 5, 4, 11, 7, 3, 0 |
| internal/oracle/testdata/reuse_narrowed.a | 7, 5, 6, 8, 5, 0 | 8, 5, 9, 10, 5, 0 |
| stage3/fixtures/assignment-proofs/01_scanner_keyword.a | new row | 18, 18, 18, 32, 6, 0 |
| stage3/fixtures/assignment-proofs/02_if_narrowing.a | new row | 6, 6, 5, 13, 3, 0 |
| stage3/fixtures/assignment-proofs/03_return_defined.a | new row | 8, 8, 6, 18, 4, 0 |
| stage3/fixtures/assignment-proofs/04_chained.a | new row | 4, 4, 6, 15, 3, 0 |
| stage3/fixtures/assignment-proofs/05_compound.a | new row | 8, 8, 0, 8, 4, 0 |
| stage3/fixtures/assignment-proofs/06_wider_target.a | new row | 4, 4, 0, 4, 4, 0 |

Column order: allocations, frees, retains, releases, peak live, in regions. No unrelated rows changed. No ahra command, full gate, PR or Loom submission was run. No formatter work was pursued. Earlier inherited source reports and commit-body notes describe earlier cuts; this report describes this re-cut.
