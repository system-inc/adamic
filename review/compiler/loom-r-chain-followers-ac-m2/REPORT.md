Built: Chain followers re-cut as one commit on fetched main.
Commits: source 09aa4ee0aad721c1a027b15eba5b23344a588a83; base 561aeb4b851ca0820f39440a09dc5fb3fd27ba40; delivery SHA is the pushed branch tip.
Commands and outputs: complete touched-package suites, consumer checks, guards, vet, formatting and a-check results are recorded below.
Mutants: targeted integration and header mutant observations are recorded below; historical source evidence is retained.
Not covered: full repository gate, opt-in lanes listed in skips.json, other compiler-reader packages, ahra and Loom submission.

Conflict: one hunk in internal/lower/refusals.go. Main added knownLoweringGap immediately after literalCallableBranch; the source added phantomRefusal in the same position. Keep main's complete check first, then the complete phantom check. The syntax kinds are disjoint; main's earlier stop keeps its priority. No main preflight is removed.

Semantic integration: the first counts run stopped at generic_function_value.a:12:32 with "an intrinsic callback parameter x with an unknown representation". Slice 2's new adaptArrayCallback inspected a raw generic signature, while this change specializes identity for its concrete contextual callback slot. Have the adapter use contextualGenericType for both its function-type key and inspected signature. All of main's boxing, nullable, rest/default, and representation-refusal logic stays unchanged. Existing TestGenericFunctionValuesHaveSeparateInstances and the full Node/native oracle cover string and numeric map(identity). The only production delta beyond source-own lines is this callback-signature integration.

Setup: the initial setup overlapped my installation of the merge tree and saw temporary conflict markers; it exited 1. Its full error log is setup-initial-failed.log. After resolving the conflict I reran setup on the stable tree: exit 0. Setup timing lines are below. No test failure was hidden by this retry. Counts regeneration first failed on the measured callback integration, before writing a table. After the integration fix the complete regeneration passed once. counts-before-integration.log and callback-before-integration.log retain the red observations.

Counts: took main's entire counts.md before the final complete -update-counts run. Exact changed rows follow (full before/after values are also in counts.json and counts.patch). No unrelated row changes.

After: `| internal/oracle/testdata/arrow_this_super.a | 5 | 5 | 8 | 8 | 5 | 0 |`
After: `| internal/oracle/testdata/arrow_this_field.a | 7 | 7 | 9 | 11 | 6 | 0 |`
After: `| internal/oracle/testdata/arrow_this_stored.a | 7 | 7 | 9 | 11 | 7 | 0 |`
After: `| internal/oracle/testdata/native-generic-empty-array.a | 3 | 3 | 1 | 5 | 3 | 0 |`
After: `| internal/oracle/testdata/generic_empty_array.a | 32 | 32 | 24 | 60 | 17 | 0 |`
After: `| internal/oracle/testdata/generic_function_value_arms.a | 18 | 18 | 72 | 99 | 7 | 0 |`
After: `| internal/oracle/testdata/generic_function_value.a | 15 | 15 | 27 | 45 | 7 | 0 |`
After: `| internal/oracle/testdata/phantom_brands.a | 15 | 15 | 29 | 42 | 6 | 0 |`
After: `| internal/oracle/testdata/phantom_undefined.a | 1 | 0 | 5 | 4 | 1 | 0 |`
After: `| internal/oracle/testdata/phantom_catch.a | 8 | 8 | 10 | 18 | 2 | 0 |`
After: `| internal/oracle/testdata/host_phantom_overload_primitive.a | 25 | 25 | 13 | 43 | 7 | 0 |`
After: `| internal/oracle/testdata/host_phantom_optional_array.a | 5 | 5 | 2 | 7 | 2 | 0 |`
After: `| internal/oracle/testdata/scanner_expressions/string-number-concatenation.a | 2 | 2 | 3 | 6 | 2 | 0 |`
After: `| internal/oracle/testdata/scanner_expressions/number_spellings.a | 85 | 85 | 2 | 88 | 3 | 0 |`
After: `| internal/oracle/testdata/scanner_expressions/primitive_concatenation.a | 33 | 33 | 2 | 38 | 3 | 0 |`

A-check status counts: {'refused': 4, 'checked': 17}. Every added/modified .a path including review evidence is compiled in place by this branch's own compiler. All program bytes remain identical to the source commit. All existing headers remain correct; no header or positional expectation changed. The unchanged pinned Gate.aCheck accepts every file and rejects removed-header and wrong-reason mutants replaying the captured real compiler diagnostic, as the a-check README permits. statuses.json and predicate.json retain per-file results.

Runtime clearance: neither source-own change nor this re-cut changes any internal/native/runtime/*.c or *.h line. Extra lines needing clearance: NONE.

Validation: GOPROXY=https://proxy.golang.org|direct; source /workspace/adamic-tools/env.sh; Linux clang 20.1.8; Go 1.27.1; Node 24.19.0; nproc=5, cpu.max=400000 100000 (four CPUs); GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1. Test output goes directly to log files. Complete package suites use -count=1 -timeout 12m -parallel 4 -p 1 -json and an outer hard bound; focused guards use -timeout 90s. Source Node, backend Node, native release, sanitizers and leaks are exercised by native/oracle suites. No test was deleted or weakened.

- counts-regenerate: exit 0, 66.712s; `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 12m -parallel 4 -args -update-counts`.
- internal-lower-all: exit 0, 57.922s; `go test ./internal/lower/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-native-all: exit 1, 320.484s; `go test ./internal/native/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-native-all-fixed: exit 0, 189.775s; `go test ./internal/native/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-oracle-all: exit 0, 448.14s; `go test ./internal/oracle/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-ir-all: exit 0, 62.46s; `go test ./internal/ir/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-javascript-all: exit 0, 2.673s; `go test ./internal/javascript/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-flow-all: exit 0, 66.788s; `go test ./internal/flow/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- internal-fresh-all: exit 0, 16.764s; `go test ./internal/fresh/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- stage3-fixtures-all: exit 0, 54.121s; `go test ./stage3/fixtures/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- reader-maps: exit 0, 6.063s; `go test ./internal/ir -run ^Test(CallTargetReaders.*|EveryMapAProgramHoldsIsSorted)$ -count=1 -timeout 90s -json`.
- parallel-guard: exit 0, 2.677s; `go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -timeout 90s -json`.
- vet: exit 0, 6.213s; `go vet ./internal/ir/... ./internal/lower/... ./internal/native/... ./internal/oracle/... ./internal/javascript/... ./internal/flow/... ./internal/fresh/... ./stage3/fixtures/...`.
- a-check: exit 0, 3.829s; `python3 review/compiler/loom-r-chain-followers-ac-m2/header-check.py`.
- integration-mutants: exit 0, 36.383s; `python3 /workspace/scratch/chain-m2-mutants.py`.

Targeted semantic mutant results:
- drop-main-preflight: exit 1, 16.459s; caught by TestMiscompile2AErrorSpread.
- drop-phantom-refusal: exit 1, 10.273s; caught by TestPhantomRefusals/prototype, TestPhantomRefusals/required_undefined, TestPhantomRefusals/nonvoid, TestPhantomRefusals/any_Path, TestPhantomRefusals/review_exact_message, TestPhantomRefusals/length, TestPhantomRefusals/string_index, TestPhantomRefusals.
- drop-callback-specialization: exit 1, 9.606s; caught by TestGenericFunctionValuesHaveSeparateInstances.

Measured leaves: 10656; slowest 21.38s. Every leaf is below 60 seconds. leaves.json retains timings and skips.json retains opt-in omissions. Historical semantic mutants in the original report are retained as historical evidence and are not claimed rerun.

Setup timing lines:
```text
setup: node ready (0.020s)
setup: go ready (0.021s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.009s
setup: markdown dependencies ready (0.061s)
setup: submodules ready (0.072s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.146s)
setup: wasi sdk ready (/workspace/adamic-tools/wasi-sdk, linking with /workspace/adamic-tools/bin/clang) (0.220s)
setup: shared cache ready (0.844s)
setup: go build ready (34.952s)
setup: test binaries deferred (use --warm-tests) (35.034s)
setup: build cache warm (35.035s)
setup: build-flags commit=561aeb4b851ca0820f39440a09dc5fb3fd27ba40 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=2.22 0.89 0.36 1/1229 731487 load-after=2.25 1.05 0.44 1/1235 732199
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (35.066s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.6ybb8r
```

Assumptions: use newest fetched origin/main when starting each branch, and preserve that base throughout its tests rather than chasing subsequent landings. The latest whole-package instruction supersedes earlier focused-only unit instructions. Preserve only source-own change plus measured integration fixes; never replace current files with historical whole-file mutants. No compiler-consuming Go package, IR-held map or runtime feature set is added. No cohere code was copied. Retire only prior task's reproducible scratch cache products when space is needed, recording exact paths in cache-space.json. One commit retains the original message plus the requested new recut line. No PR, ahra check, Loom submission, force push, main push or area-branch push is performed.

Environment recovery and approval review are documented in [ENVIRONMENT.md](ENVIRONMENT.md). The first full native suite had exactly one failing leaf, TestLongUnicodePortabilityGCC, due to the missing sanitizer header; the complete native suite passed with actual GCC and the command-local scratch wrapper. The shared-toolchain wrapper was rejected by automatic approval review and was never written. No runtime source was changed.

The second validation driver reached its 840-second whole-driver hard limit during stage3. It interrupted that suite after about 53 seconds, with no over-budget test leaf. The full stage3 suite then passed under its own bound in 54.121 seconds. The interrupted output is retained in stage3-fixtures-interrupted.log. This was a driver scheduling limit, not a test assertion failure.

Main's knownLoweringGap preflight prevents Error-storage spreads, unsupported numeric setter ABIs, overlong property names, unsupported computed derived-class members, and constructor-arrow captures before readiness. All remain before the new phantom-intersection refusal.
