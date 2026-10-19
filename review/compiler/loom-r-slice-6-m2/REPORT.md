Re-cut lowering slice 6 onto today's main, regenerate flow/fresh corpora, and repair uncaught Error text.
Branch compiler/loom-r-slice-6-m2; one commit on 5565a7cafb0ee8d538f8a5ece2304d5aa4254145; delivery SHA is in the final response.
Flow, fresh, JavaScript, focused native oracle fixtures, counts, guards, vet, formatting and a-check passed.
Both real always-separator mutants failed exact stderr assertions; restored renderer tests passed.
Full native and stage1 suites and the earlier separate 61-mutant campaign were not repeated for this re-cut.

## Base, method and conflicts

Base `5565a7cafb0ee8d538f8a5ece2304d5aa4254145`, fetched from origin/main. Source `4955f059e924f858e25aa425029ea846826934fa`. Used `git merge-tree --write-tree --merge-base=4955f059~1 origin/main 4955f059`, materialized tree `0f0182336a69a724d50bddb92cdfbaa186076754`, then regenerated conflicts. Exactly two conflicts: internal/flow/corpus_units_test.go and internal/fresh/corpus_units_test.go, both generated fixture lists. Ran their owner `python3 review/test-split-flow/generate.py`; no list line was hand-merged. Main's generator is unchanged. Committed its units.json manifest too. With the runtime review's five new fixtures, it generates 4,346 program units. A second generation reproduces both lists and manifest byte-for-byte (generated-digests.json). Main's regexp_native_input_cache.a remains in both corpora, plus the slice 6 fixtures and new Error cases. Slice 6's earlier intentional writing-call refusal moves still have dedicated flow/fresh refusal tests; no test was deleted to resolve these conflicts.

Assumptions: main's newer regexp runtime is authoritative; preserve it. The byte-identical runtime requirement applies to slice 6's runtime changes, without reverting main's already landed regexp.c, regexp.h and regexp_regular.c. Of the fourteen runtime files slice 6 changes, thirteen equal 4955f059 byte-for-byte; exceptions.c differs only by the runtime review's requested separator line. runtime-preservation.json records these checks. Keep the previously agreed `Uncaught ` prefix, escaped single diagnostic line, exit1 and payload release; change only Error.prototype.toString's name/message composition. No new runtime static or concurrency state.

## Error renderer repair and Node evidence

Native now writes `: ` only when name and message are both non-empty. JavaScript captures each field once and applies the same composition. Five .a fixtures and five top-level tests cover omitted message, empty message, empty name, both empty, and a named non-empty control. The tests ask source Node's actual Error.prototype.toString for each composition, then hold sanitized native, release native, the JavaScript backend and a leak-enabled run to the exact owned line. Expected stderr respectively: `Uncaught Error\n`, `Uncaught Error\n`, `Uncaught message\n`, `Uncaught \n`, `Uncaught Failure: message\n`. All exit1 with no stdout. Every fixture typechecks and lowers on the final branch compiler, with no a-check error header needed.

Separate top-level native and JavaScript mutant tests compile/run actual always-separator renderers and require a diagnostic mismatch with otherwise clean execution. Additionally, run-separator-mutants.py temporarily changed the real native runtime line and the real JavaScript formatter source, one at a time, and ran the normal fixture test. Native mutation: TestUncaughtErrorEmptyMessage failed on `Uncaught Error: \n` versus `Uncaught Error\n`. JavaScript mutation: TestUncaughtErrorEmptyName failed on `Uncaught : message\n` versus `Uncaught message\n`. Both test commands exit1 on the expected stderr assertion, not a build or sanitizer failure. Real mutation patches, raw red JSONL and separator-mutants.json are retained. Both sources were restored in finally blocks, and the final renderer run passed again in3.963s. The initial new helper run used a relative fixture path incompatible with the oracle cache; that harness mistake was fixed to filepath.Abs, and the raw failed log remains.

## Exact runtime line differing from 4955f059

No additional slice 6 C/header line differs. Main's three regexp runtime files are unchanged from current main, not new edits of this commit. The complete exceptions.c clearance delta is:

```diff
diff --git a/internal/native/runtime/exceptions.c b/internal/native/runtime/exceptions.c
index d8b7861e..2d80e7ab 100644
--- a/internal/native/runtime/exceptions.c
+++ b/internal/native/runtime/exceptions.c
@@ -124,7 +124,7 @@ _Noreturn void adamic_uncaught(void) {
 		adamic_slot_cache name_cache = {0};
 		adamic_slot_cache message_cache = {0};
 		uncaught_text(adamic_object_field(error, "name", &name_cache)->reference);
-		fputs(": ", stderr);
+		if (adamic_string_length(adamic_object_field(error, "name", &name_cache)->reference) != 0 && adamic_string_length(adamic_object_field(error, "message", &message_cache)->reference) != 0) { fputs(": ", stderr); }
 		uncaught_text(adamic_object_field(error, "message", &message_cache)->reference);
 	} else if (adamic_thrown == &adamic_null) {
 		fputs("null", stderr);
```

## Counts

Final Linux clang20.1.8 regeneration: `timeout 180 go test -p 1 ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -timeout 150s -json -args -update-counts`, PASS65.320s. No hand merging. Earlier pre-repair regeneration passed83.566s and is retained separately.

153 rows affected relative to current main: 31 added, 9 removed, 113 changed. Every row's name, before, after and owner is listed exhaustively in [counts-rows.json](counts-rows.json). The carried147 changes match the previous cut. inherit-guards owns inherited_fields_guard.a; exceptions-21-main owns the other carried rows. Main's new regexp_native_input_cache.a measures148 allocations/148 frees/86 retains/160 releases/peak15/regions0, versus main's90 retains/164 releases; that row is attributed to the carried exceptions lowering on a newly present main fixture, without claiming a separately isolated optimization. The renderer repair itself adds only the following five rows and moves no existing row:

| internal/oracle/testdata/step21_error_empty_message.a | 1 | 1 | 4 | 4 | 1 | 0 |
| internal/oracle/testdata/step21_error_empty_name.a | 1 | 1 | 6 | 6 | 1 | 0 |
| internal/oracle/testdata/step21_error_empty_name_and_message.a | 1 | 1 | 6 | 6 | 1 | 0 |
| internal/oracle/testdata/step21_error_missing_message.a | 1 | 1 | 4 | 4 | 1 | 0 |
| internal/oracle/testdata/step21_error_named_message.a | 1 | 1 | 6 | 6 | 1 | 0 |

## Tests and grain

All tests run with GOMAXPROCS4 and parallel4 under a four-CPU cgroup (nproc5). Hard bounds and raw output retained. Final results:

- `timeout 180 go test -p 1 ./internal/flow/... ./internal/fresh/... ./internal/javascript/... -count=1 -parallel=4 -timeout 150s -json`: flow62.430s, fresh11.565s, JavaScript0.136s, all pass. No leaf over60s; slowest13.89s, flow timsort graph paths.
- `timeout 120 go test -p 1 ./internal/oracle -run '^(TestUncaughtError|TestStep21Uncaught)' -count=1 -parallel=4 -timeout 90s -json`: pass5.587s. Restored renderer-only controls pass3.963s. Sanitizers and leak checks included in each new fixture's test.
- Structural guards: TestEveryTestHasOneDescriptiveName, TestEveryTestIsParallelOrSaysWhy, TestEveryMapAProgramHoldsIsSorted, TestNativeAgreementRowsNameRegisteredPrograms, TestRuntimeSetsAreEachUsed and TestRuntimeStaticsAreListed, all pass in guards-final.jsonl. Four TestCallTargetReaders variants and their export setup pass in readers-*.jsonl, each under -timeout90s. Script uses independent disposable export product copies on the larger scratch volume, removes each only after its guard completes, and leaves source/active Go cache alone.
- go vet -p1 all16 changed normal Go packages passes; explicit build-tagged stage3/drivers/scanner/build-metrics.go vet passes. vet-packages.json and empty logs retained. gofmt -l all changed Go sources produces no output; git diff --check production/test sources clean.
- Final branch compiler built successfully, then unchanged pinned Gate.aCheck checked all46 changed .a files:29 checked,16 expected refusals,1 expected checker error; exit0. a-check.py, final.json.txt and a-check.log retain commands, exact diagnostics, header predicate and per-file timing.

New top-level tests, observed seconds:

- TestUncaughtErrorSeparatorMutantJavaScript: 3.3s
- TestUncaughtErrorEmptyMessage: 3.68s
- TestUncaughtErrorNamedMessage: 3.72s
- TestUncaughtErrorSeparatorMutantNative: 0.5s
- TestUncaughtErrorEmptyNameAndMessage: 0.5s
- TestUncaughtErrorEmptyName: 0.47s
- TestUncaughtErrorMissingMessage: 0.48s

Full per-package status, leaf limits and expected mutant failures are in final-validation-summary.json. The native root package, broad lower/oracle suites and stage1 suites were not repeated beyond requested guards and focused renderer/uncaught tests; previous source evidence remains available. No ahra, Loom submission or PR.

## Setup and execution observations

`export GOPROXY='https://proxy.golang.org|direct'`; bounded cloud/setup.sh reached Go, Node, markdown, submodules and clang, then exit1: `github.com/system-inc/adamic/cmd/adamic: /workspace/adamic-tools/go/pkg/tool/linux_amd64/compile: signal: killed`. Timing lines: go0.023s, node0.029s, markdown0.078s, submodules0.079s, clang0.174s, shared-cache-off0.175s; no final timing. Provisioned Go1.27.1, Node24.19.0 and house clang20.1.8 remained usable; sourced /workspace/adamic-tools/env.sh and retained task cache/WASI environment. nproc5, cpu.max400000/100000. Recovered memory by evicting a disposable previous-cut copied export product, then, only after all running Go commands finished, old unshared Go action bodies. No source or active Go cache file was removed. An attempted cross-volume hardlink dedup failed EXDEV; its raw log remains. Final compiler build, vet and reader commands used GOGC50/GOMEMLIMIT768MiB and completed. Setup/OOM is not hidden as a passing setup. No stage1 memory failure was investigated.
