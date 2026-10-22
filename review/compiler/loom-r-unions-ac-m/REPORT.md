Built: unions recut onto current main with the stale meter diagnostic witness fixed and main's narrow-array refusal preserved.
Commits: source 63604fed2dc1fd5f6b6f4c2198eb374ecec2d05a; base 8cfcf172d08e685850cfd38b1002cfd69e25d3e6; delivery SHA is the pushed branch tip.
Commands and outputs: requested package suites, touched stage1 packages, flow/fresh and stage3 consumers, guards, vet, formatting and counts passed; raw commands/results are below.
Mutants: four recorded union mutants caught; equality-observation stopped at an ENOSPC Go build, never reached a test and was not rerun; 21 further Go mutants were not run. Both header mutants caught; the existing IR check-removal oracle passed in the full oracle suite.
Not covered: the stopped equality-observation mutant and 21 unrun Go mutants (mutant-coverage.json), full repository gate, other compiler-reader packages, ahra/Loom submission, opt-in skips, and the source report's earlier typeaware timeout.

Observed red: only unions, not its base. Whole cmd/adamic-meter/... at dbdf977f5b8b0979d414ebe6f20e7aadd312e309 passed (15.503s command; 1.458s package). The corrected actual source tip 63604fed2dc1fd5f6b6f4c2198eb374ecec2d05a failed (16.902s command; full exact timing in comparison.json). Both raw JSON logs are retained. TestFixtureCorpusCountsEveryDiagnosticKind reported counts = map[string]int{"Refused":1}, want one Refused and one NotYet. TestJSONIsMachineReadable parsed valid JSON, then rejected a report with only one reason. No kind registration or JSON encoder defect was observed.

Fix: unions admits the old corpus's required boolean | undefined class field. Move that exact program body from testdata/corpus/not_yet.ts into testdata/compiled/main.a and add TestUnionClassFieldHasNoDiagnostic. Replace the NotYet corpus witness with a .a program exercising the still-unimplemented nullable object array lookup with distinct null/undefined tags. Retain every assertion in both originally failing tests, the meter implementation, and all other meter tests. No test is loosened or removed. The new files use .a. Independent actual-compiler/Node/native sanitizer observations and meter mutants are recorded separately.

Re-cut: fetched main 8cfcf172d08e685850cfd38b1002cfd69e25d3e6, then applied only the source's own parent-to-tip diff using cherry-pick --no-commit. Requested merge-tree --write-tree --merge-base=63604fed2dc1~1 origin/main 63604fed2dc1 was clean. There are no conflict hunks. Git automatically merged object.go and counts.md. All source production added/deleted lines are preserved; no main implementation line is reverted.

Main integration: its newer narrow-array-literal layout refusal (4941491215606601892b4f20190577a9a55d68e5) precedes the old nullable-storage spread NotYet on one source witness. The first full lower run failed only that stale expected stop; the log is retained. Split that one source into top-level TestNullableObjectSpreadKeepsNarrowerLiteralRefused, requiring the exact Refused type, reason and fix. Keep all three other storage view probes. The oracle spread path likewise keeps its Node check and now requires exactly that named refusal and fix, rather than accepting a broader class of diagnostics. Add only // a-check: refused union-typed array built from a narrower literal to array_spread.a; its complete body is unchanged. Path search found no pinned source position for it. Main's guard is not changed; its removal is a new mutant.

Counts: discarded the auto-merged table, took main's entire table, then regenerated the complete family exactly once on Linux/clang 20.1.8 after the rest compiled. Exactly sixteen added rows and the existing taste_stage3_representations.a row change; all resulting rows are byte-identical to source values. Existing taste retains/releases 11/34 become 13/36 because the change routes narrowed string fields through checked union helpers (the original source report records its counterfactual). No unrelated existing row changes. counts.json and counts.patch retain every old/new value.

Runtime clearance: no source-own or recut added/deleted runtime .c/.h line exists. The three native/union.go emission lines remain identical to source. No extra runtime resolution line needs clearance.

Assumptions: use the fetched actual branch SHA corrected by the user, rather than the mistyped full SHA. A diagnostic-coverage corpus must track real compiler boundaries; preserve its old newly admitted witness positively instead of changing the meter's truthful result or relaxing the corpus assertions. Apply current main's earlier precise refusal where it precedes the old stop. The latest whole-package instruction supersedes older focused-only instructions. Never use archived whole-file mutant source to overwrite current main; port only each mutation's diff. No new compiler-consuming Go package, IR Program map or runtime implementation is introduced. No cohere code was copied. No ahra command, PR or Loom submission is run.

Environment recovery: the first full native suite failed only TestSplitAttributedBindingsAgreeWithNode/wasm32-wasi, introduced on main by a01172d7, because WASI_SYSROOT was unset. Exact output: units_test.go:244: native: wasm32-wasi requires WASI_SYSROOT. The rest of that full run completed. Installed the repository-supported WASI SDK 27 with cloud/setup.sh --wasi-sdk, using a preserved symlink from the tools directory into the separate temporary volume to avoid filling the root disk. Main's test was not skipped, edited or weakened. Source the printed environment and rerun the full native package tree, retaining the original environmental red log. House native clang stays 20.1.8; the SDK also reports clang 20.1.8-wasi-sdk. Opt-in WASI oracle/integration lanes remain separately opt-in and are not claimed covered.

Consumer product ordering: the first full selector/... and values/... invocation ran root-package consumers before their oracle owner packages. Eight consumer leaves failed only with the exact missing-product message settle the Go oracle owner first. The later selector/oracles and values/oracles packages in that same invocation successfully built their Go oracle/answer products; those successful owner observations are in stage1-touched.log. Rerun the full two package trees after owners settle. No cohere implementation, test expectation, producer/consumer source or product key is changed to make them pass.

Resumption and explicit mutant limit: four mutants completed and were caught: null-box, runtime-typeof, lower-typeof, nullable-admission. The fifth, equality-observation, exited during Go compilation: compile: writing output: write $WORK/b132/_pkg_.a: no space left on device; a second line also reported inability to write vet.cfg. It did not reach any test and is not a semantic catch. The runner correctly stopped; the original log and false killed result are retained. The resumed user explicitly directed push rather than rerunning a stopping mutant. Therefore do not rerun this mutant or start the 21 remaining Go mutants; list all missing coverage in mutant-coverage.json. No live source was mutated by these overlay runs.

Resume assumptions: keep the already validated base 8cfcf172d08e685850cfd38b1002cfd69e25d3e6, the newest origin/main fetched when this unit began, instead of starting a fresh rebase and invalidating completed suites. Finish the separately requested CLI witness/header proofs before the single final commit so their evidence ships in that commit. The first standalone Node witness probe printed the expected true but also Node's experimental stripTypeScriptTypes warning because its command omitted the documented --disable-warning=ExperimentalWarning flag; preserve that failed harness attempt and use oracle/node.mjs's documented invocation for the final proof. This changes only the evidence harness, not a fixture, test assertion or compiler. No stopped Go mutant is rerun.


Exact original failure output:

```text
main_test.go:79: report = main.report{Root:"/workspace/adamic/cmd/adamic-meter/testdata/corpus", FilesExamined:2, JavaScriptSkipped:0, FilesReachingLowering:2, Reasons:[]*main.finding{(*main.finding)(0x2847c32f38c0)}, Adaptations:[]main.adaptation(nil)}
main_test.go:24: counts = map[string]int{"Refused":1}, want one Refused and one NotYet
```

Validation ran with GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1, -count=1, -parallel 4 and -p 1. Package suites have a 12-minute Go timeout and 840-second outer limit; focused guards/mutants have 90-second Go and bounded process limits. All test output goes directly to logs. Grouping parents are distinguished from leaves. Every measured leaf, including baseline failures and intentionally killed mutants, stayed below 60 seconds. Tests were split when main changed the classification; none was deleted.

The full lower suite includes native-agreement program registration; full oracle includes exact runtime-set witnesses and Node/native/JavaScript comparisons under ASan/UBSan/LeakSanitizer. Explicit IR TestCallTargetReaders (-timeout 90s), SortedMap and cmd/adamic-gate parallel-test guards passed. No new Go package, IR-held map, product recipe or runtime implementation is added. Full touched stage1 selector/values and full fixture-consuming flow/fresh suites passed. The full stage3 fixture package passed with no status refresh. Gofmt and source whitespace checks are recorded separately.

New meter witnesses: fixture-proof.json records actual source Node and sanitized native output for the admitted boolean-field program, and actual Node plus the expected NotYet diagnostic for the replacement lookup. The original two failing assertions and the complete meter implementation remain byte-identical to main.

A-check: compiled every added/modified .a file, including the new meter programs, using the branch compiler and unchanged pinned Gate.aCheck. Totals: 17 checked, 6 NotYet, 1 refused, 0 type errors. Added one matching first-line header to scout_nullable_storage_refused/array_spread.a; preserved its body exactly and found no pinned positional reader. The existing headed non-null witness is the removed-header/wrong-reason control. Both failed the predicate, and restoration passed with unchanged bytes. Exact statuses and real diagnostics are in a-check-statuses.json and predicate-results.json.

Setup succeeded with GOPROXY=https://proxy.golang.org|direct and sourced /workspace/adamic-tools/env.sh. Linux house clang 20.1.8; Go 1.27.1; Node 24.19.0. nproc=5, cpu.max=400000 100000 (four CPUs). Used main's supported ADAMIC_BUILD_OUTPUT_DIR and ADAMIC_GOCACHE_DIR on the separate temporary volume for new products/cache writes, preserving existing cache contents. This affects the environment only, not product keys or source. Timing lines:

```text
setup: go ready (0.018s)
setup: node ready (0.021s)
setup: submodules ready (0.063s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.009s
setup: markdown dependencies ready (0.068s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.151s)
setup: shared cache ready (0.759s)
setup: go build ready (37.858s)
setup: test binaries deferred (use --warm-tests) (37.937s)
setup: build cache warm (37.938s)
setup: build-flags commit=8cfcf172d08e685850cfd38b1002cfd69e25d3e6 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.02 1.13 1.96 1/697 422108 load-after=1.18 1.29 1.98 1/703 423003
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (37.964s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.XVpmWE
```

WASI setup recovery timing lines (supported --wasi-sdk, success):

```text
setup: go ready (0.017s)
setup: node ready (0.021s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.007s
setup: submodules ready (0.060s)
setup: markdown dependencies ready (0.062s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.153s)
setup: wasi sdk ready (/workspace/adamic-tools/wasi-sdk, linking with /workspace/adamic-tools/bin/clang) (23.932s)
setup: shared cache ready (24.342s)
setup: go build ready (61.538s)
setup: test binaries deferred (use --warm-tests) (61.620s)
setup: build cache warm (61.622s)
setup: build-flags commit=8cfcf172d08e685850cfd38b1002cfd69e25d3e6 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=1.89 2.69 2.45 1/849 470809 load-after=1.62 2.42 2.37 1/846 471458
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (61.651s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.Wyi5Xn
```

Package commands and results:

- compiler-build: exit 0, 7.946s; `go build -o /tmp/unions-m-adamic ./cmd/adamic`.
- stage3-all: exit 0, 68.854s; `go test ./stage3/fixtures/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- counts-regenerate: exit 0, 52.357s; `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 12m -parallel 4 -p 1 -json -args -update-counts`.
- meter-all: exit 0, 5.151s; `go test ./cmd/adamic-meter/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- lower-all: exit 1, 54.483s; `go test ./internal/lower/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- lower-all-fixed: exit 0, 48.814s; `go test ./internal/lower/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- native-all-fixed: exit 1, 217.586s; `go test ./internal/native/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- native-all-wasi: exit 0, 181.37s; `go test ./internal/native/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- oracle-all-wasi: exit 0, 376.056s; `go test ./internal/oracle/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- stage1-touched: exit 1, 56.93s; `go test ./stage1/cohere/selector/... ./stage1/cohere/values/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- stage1-touched-ready: exit 0, 36.049s; `go test ./stage1/cohere/selector/... ./stage1/cohere/values/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- flow-fresh: exit 0, 92.986s; `go test ./internal/flow/... ./internal/fresh/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- readers-maps: exit 0, 19.425s; `go test ./internal/ir -run ^Test(CallTargetReaders|EveryMapAProgramHoldsIsSorted)$ -count=1 -timeout 90s -parallel 4 -json`.
- parallel-guard: exit 0, 2.58s; `go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -timeout 90s -parallel 4 -json`.
- vet: exit 0, 7.022s; `go vet ./cmd/adamic-meter/... ./internal/lower/... ./internal/native/... ./internal/oracle/... ./stage1/cohere/selector/... ./stage1/cohere/values/... ./internal/flow/... ./internal/fresh/... ./internal/ir/... ./stage3/fixtures/...`.

Counts additions (A/F/retain/release/peak/region):

```text
| internal/oracle/testdata/scout_boxed_scalar_fields.a | 62 | 62 | 114 | 169 | 6 | 0 |
| internal/oracle/testdata/scout_boxed_scalar_stale_field.a | 2 | 1 | 3 | 3 | 2 | 0 |
| internal/oracle/testdata/scout_class_boolean_fields.a | 22 | 22 | 9 | 31 | 4 | 0 |
| internal/oracle/testdata/scout_selector_optional_boolean.a | 2 | 2 | 1 | 4 | 2 | 0 |
| internal/oracle/testdata/scout_values_optional_boolean.a | 2 | 2 | 2 | 5 | 2 | 0 |
| internal/oracle/testdata/scout_nullable_objects.a | 31 | 31 | 139 | 162 | 5 | 0 |
| internal/oracle/testdata/scout_nullable_object_stale_field.a | 2 | 1 | 6 | 5 | 2 | 0 |
| internal/oracle/testdata/scout_nullable_object_null_field.a | 2 | 2 | 11 | 12 | 2 | 0 |
| internal/oracle/testdata/scout_nullable_object_stale_local.a | 1 | 1 | 6 | 5 | 1 | 0 |
| internal/oracle/testdata/scout_string_number_fields_locals.a | 157 | 157 | 137 | 295 | 5 | 0 |
| internal/oracle/testdata/scout_string_number_stale_field.a | 2 | 0 | 3 | 3 | 2 | 0 |
| internal/oracle/testdata/scout_hir_optional_boolean.a | 1 | 1 | 0 | 1 | 1 | 0 |
| internal/oracle/testdata/scout_tsc_source_content.a | 0 | 0 | 2 | 2 | 0 | 0 |
| internal/oracle/testdata/scout_tsc_nullish_content.a | 0 | 0 | 6 | 6 | 0 | 0 |
| internal/oracle/testdata/scout_nullable_strings.a | 34 | 34 | 130 | 156 | 5 | 0 |
| internal/oracle/testdata/scout_nullable_string_stale_field.a | 1 | 0 | 4 | 3 | 1 | 0 |
```

Counts existing-row changes:

```text
| internal/oracle/testdata/taste_stage3_representations.a | 24 | 24 | 11 | 34 | 10 | 0 |
| internal/oracle/testdata/taste_stage3_representations.a | 24 | 24 | 13 | 36 | 10 | 0 |
```

Completed mutant catchers (test assertions, not build failures or timeouts):

- null-box: TestNativeAgreesWithNode_035/internal/oracle/testdata/scout_nullable_strings.a, TestNativeAgreesWithNode_035; 16.252s command.
- runtime-typeof: TestNativeAgreesWithNode_091/internal/oracle/testdata/scout_tsc_nullish_content.a, TestNativeAgreesWithNode_091; 15.852s command.
- lower-typeof: TestNativeAgreesWithNode_091/internal/oracle/testdata/scout_tsc_nullish_content.a, TestNativeAgreesWithNode_091; 15.69s command.
- nullable-admission: TestNullableStringAdmission; 14.427s command.

New test timings (successful observations):

- TestNullableObjectSpreadKeepsNarrowerLiteralRefused: 0.02s (lower-all-fixed.log).
- TestUnionClassFieldHasNoDiagnostic: 0.12s (meter-all.log).
- TestUnionClassFieldHasNoDiagnostic: 0.11s (meter-fixed.log).

Measured leaves: 12332; maximum 19.53 seconds. Exact per-leaf records, expected-red observations, skips and complete logs are retained.
