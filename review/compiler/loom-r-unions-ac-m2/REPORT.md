Built: Unions re-cut as one commit on fetched main.
Commits: source 225850deea88029ed269a8b0e14b6832ebb49134; base a22f4251067a5f53a00d07fc64b54a2ac72d340b; delivery SHA is the pushed branch tip.
Commands and outputs: complete touched-package suites, consumer checks, guards, vet, formatting and a-check results are recorded below.
Mutants: targeted integration and header mutant observations are recorded below; historical source evidence is retained.
Not covered: full repository gate, opt-in lanes listed in skips.json, other compiler-reader packages, ahra and Loom submission.

Conflict: internal/oracle/counts.md only. Retain main's entire table, then regenerate every counts row using Linux house clang 20.1.8. No table hunk is merged by hand. Main's newer fixture rows remain intact; the changed rows below are exclusively this change's. See merge-tree.log for every conflict hunk location and merge diagnostic.

Counts: took main's entire counts.md before the final complete -update-counts run. Exact changed rows follow (full before/after values are also in counts.json and counts.patch). No unrelated row changes.

Before: `| internal/oracle/testdata/taste_stage3_representations.a | 24 | 24 | 11 | 34 | 10 | 0 |`
After: `| internal/oracle/testdata/taste_stage3_representations.a | 24 | 24 | 13 | 36 | 10 | 0 |`
After: `| internal/oracle/testdata/scout_boxed_scalar_fields.a | 62 | 62 | 114 | 169 | 6 | 0 |`
After: `| internal/oracle/testdata/scout_boxed_scalar_stale_field.a | 2 | 1 | 3 | 3 | 2 | 0 |`
After: `| internal/oracle/testdata/scout_class_boolean_fields.a | 22 | 22 | 9 | 31 | 4 | 0 |`
After: `| internal/oracle/testdata/scout_selector_optional_boolean.a | 2 | 2 | 1 | 4 | 2 | 0 |`
After: `| internal/oracle/testdata/scout_values_optional_boolean.a | 2 | 2 | 2 | 5 | 2 | 0 |`
After: `| internal/oracle/testdata/scout_nullable_objects.a | 31 | 31 | 139 | 162 | 5 | 0 |`
After: `| internal/oracle/testdata/scout_nullable_object_stale_field.a | 2 | 1 | 6 | 5 | 2 | 0 |`
After: `| internal/oracle/testdata/scout_nullable_object_null_field.a | 2 | 2 | 11 | 12 | 2 | 0 |`
After: `| internal/oracle/testdata/scout_nullable_object_stale_local.a | 1 | 1 | 6 | 5 | 1 | 0 |`
After: `| internal/oracle/testdata/scout_string_number_fields_locals.a | 157 | 157 | 137 | 295 | 5 | 0 |`
After: `| internal/oracle/testdata/scout_string_number_stale_field.a | 2 | 0 | 3 | 3 | 2 | 0 |`
After: `| internal/oracle/testdata/scout_hir_optional_boolean.a | 1 | 1 | 0 | 1 | 1 | 0 |`
After: `| internal/oracle/testdata/scout_tsc_source_content.a | 0 | 0 | 2 | 2 | 0 | 0 |`
After: `| internal/oracle/testdata/scout_tsc_nullish_content.a | 0 | 0 | 6 | 6 | 0 | 0 |`
After: `| internal/oracle/testdata/scout_nullable_strings.a | 34 | 34 | 130 | 156 | 5 | 0 |`
After: `| internal/oracle/testdata/scout_nullable_string_stale_field.a | 1 | 0 | 4 | 3 | 1 | 0 |`

A-check status counts: {'checked': 17, 'NotYet': 7, 'refused': 1}. Every added/modified .a path including review evidence is compiled in place by this branch's own compiler. Carried program bytes remain identical to the source commit; any additional regression program is identified separately below. All existing headers remain correct; no header or positional expectation changed. The unchanged pinned Gate.aCheck accepts every file and rejects removed-header and wrong-reason mutants replaying the captured real compiler diagnostic, as the a-check README permits. statuses.json and predicate.json retain per-file results.

Runtime clearance: neither source-own change nor this re-cut changes any internal/native/runtime/*.c or *.h line. Extra lines needing clearance: NONE.

Validation: GOPROXY=https://proxy.golang.org|direct; source /workspace/adamic-tools/env.sh; Linux clang 20.1.8; Go 1.27.1; Node 24.19.0; nproc=5, cpu.max=400000 100000 (four CPUs); GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1. Test output goes directly to log files. Complete package suites use -count=1 -timeout 12m -parallel 4 -p 1 -json and an outer hard bound; focused guards use -timeout 90s. Source Node, backend Node, native release, sanitizers and leaks are exercised by native/oracle suites. No test was deleted or weakened.

- counts-regenerate: exit 0, 55.682s; `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 12m -parallel 4 -p 1 -json -args -update-counts`.
- cmd-adamic-meter-all: exit 0, 9.674s; `go test ./cmd/adamic-meter/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-lower-all: exit 0, 55.108s; `go test ./internal/lower/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-native-all: exit 0, 231.125s; `go test ./internal/native/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-oracle-all: exit 0, 392.978s; `go test ./internal/oracle/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- stage1-cohere-selector-oracles-all: exit 0, 24.144s; `go test ./stage1/cohere/selector/oracles/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- stage1-cohere-values-oracles-all: exit 0, 8.04s; `go test ./stage1/cohere/values/oracles/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- stage1-cohere-selector-all: exit 0, 22.052s; `go test ./stage1/cohere/selector/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- stage1-cohere-values-all: exit 0, 29.316s; `go test ./stage1/cohere/values/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-flow-all: exit 0, 78.934s; `go test ./internal/flow/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-fresh-all: exit 0, 19.666s; `go test ./internal/fresh/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- stage3-fixtures-all: exit 0, 64.951s; `go test ./stage3/fixtures/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- reader-maps: exit 0, 32.612s; `go test ./internal/ir -run ^Test(CallTargetReaders.*|EveryMapAProgramHoldsIsSorted)$ -count=1 -timeout 90s -parallel 4 -json`.
- parallel-guard: exit 0, 1.27s; `go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -timeout 90s -json`.
- vet: exit 0, 8.363s; `go vet ./cmd/adamic-meter/... ./internal/lower/... ./internal/native/... ./internal/oracle/... ./stage1/cohere/selector/... ./stage1/cohere/values/... ./internal/flow/... ./internal/fresh/... ./internal/ir/... ./stage3/fixtures/...`.
- a-check: exit 0, 4.038s; `python3 review/compiler/loom-r-unions-ac-m2/header-check.py`.
- meter-node-witnesses: exit 0, 1.217s; `python3 review/compiler/loom-r-unions-ac-m2/fixture-proof.py`.
- integration-mutants: exit 0, 37.778s; `python3 review/compiler/loom-r-unions-ac-m2/integration-mutants.py`.
- count-mutants: exit 0, 23.657s; `python3 review/compiler/loom-r-unions-ac-m2/count-mutants.py`.

Targeted semantic mutant results:
- drop-map-adapter-exemption: exit 1, 18.575s; caught by TestCallbackWideningNullableObjectMapRelease.
- drop-nested-storage-match: exit 1, 19.14s; caught by TestNullableObjectMapKeepsArrayStorageViewNotYet.
- removed-count-row: exit 1, 3.839s; caught by TestCountsAreRecordedTable.
- wrong-count-row: exit 1, 19.779s; caught by TestCountsAreRecorded_010/internal/oracle/testdata/scout_class_boolean_fields.a, TestCountsAreRecorded_010.

Measured leaves: 11798; slowest 21.91s. Every leaf is below 60 seconds. leaves.json retains timings and skips.json retains opt-in omissions. Historical semantic mutants in the original report are retained as historical evidence and are not claimed rerun.

Setup timing lines:
```text
setup: go ready (0.018s)
setup: node ready (0.020s)
setup: submodules ready (0.054s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.006s
setup: markdown dependencies ready (0.057s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.165s)
setup: wasi sdk ready (/workspace/adamic-tools/wasi-sdk, linking with /workspace/adamic-tools/bin/clang) (0.243s)
setup: shared cache ready (0.684s)
setup: go build ready (43.378s)
setup: test binaries deferred (use --warm-tests) (43.466s)
setup: build cache warm (43.467s)
setup: build-flags commit=a22f4251067a5f53a00d07fc64b54a2ac72d340b nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=1.07 3.11 3.19 1/1581 918974 load-after=1.84 3.03 3.16 1/1570 919585
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (43.500s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.rpYnIK
```

Assumptions: use newest fetched origin/main when starting each branch, and preserve that base throughout its tests rather than chasing subsequent landings. The latest whole-package instruction supersedes earlier focused-only unit instructions. Preserve only source-own change plus measured integration fixes; never replace current files with historical whole-file mutants. No compiler-consuming Go package, IR-held map or runtime feature set is added. No cohere code was copied. Retire only prior task's reproducible scratch cache products when space is needed, recording exact paths in cache-space.json. One commit retains the original message plus the requested new recut line. No PR, ahra check, Loom submission, force push, main push or area-branch push is performed.

The additional semantic integration and its measured initial failure are detailed in [INTEGRATION.md](INTEGRATION.md). Main's map adapter stays unchanged. The carried shared-storage guard now recognizes a direct array-map callback only when its result type and nonprimitive parameter storage remain identical. Rest and other intrinsic callback views retain conservative stops. The new negative array_callback.a program prints null on Node and must remain NotYet; the existing nullable-object map passes all three backends with sanitizers.

The counts conflict has one hunk: main's stage3/fixtures/self-compare/{closure,array-fill,sweep}.a rows versus the source's sixteen scout insertions. Main's rows were preserved by taking its complete table before regeneration. The sixteen additions and one taste row change above are source-identical. No runtime C/H line is changed.

A-check freshly compiles all 24 carried programs, plus the new array_callback.a regression. All carried files are byte-identical to the source commit. The new program is NotYet and needs no header. Every original header is retained; no positional expectation changes. Fresh meter witness evidence is in fixture-proof.json. Targeted map/count/header mutants were all caught. The historical original Go mutant catalogue, including its earlier stopped and unrun entries, was not rerun or claimed covered by this re-cut.
