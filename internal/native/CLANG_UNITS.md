Built an opt-in deterministic function-group compiler and correctly keyed object cache; default C emission is unchanged.
Base e011f8f60899586d6373a5ccb07335ad82cfbf3c; measured implementation 61075f294ed0b9a1a8b9b282c9d6e626c99ad6b0; lint source f4d98cab50048692781da3599131317dc569d466.
Commands, raw timings, traces and output-parity evidence are saved beside this report; final gate results follow below.
Flags-key and duplicated-state mutants fail on executed behavior; literal-rewrite and unit-order mutants prove preservation checks can fail.
Not covered: general one-module rebuilds when emitter numbering/header declarations change, the complete pinned TypeScript corpus, or a default TSGo dispatcher.

## System-header portability correction

The original measurements below are historical, at commit 61075f2. They predate this correction and were not remeasured. The macOS Apple clang 21 failure was reported by integration; this worker verified the analogous failure on Linux clang 20.1.8, not on macOS.

The old `clang -E -P` snapshot discarded system-header line markers. Compiling that snapshot under `-pedantic -Werror` treats platform extensions, including stdio.h's `_Nullable` function pointers, as user-code extensions. The fix removes `-P`, retaining clang's `# ... "header" ... 3` system-header flag. Object compilation accepts only clang's generated line-marker syntax with `-Wno-gnu-line-marker`; nullability and other pedantic warnings remain errors in user code. The same preprocessed snapshot is hashed and compiled, preserving transitive header contents and provenance without reopening headers. Ordered build flags and the compiler path/full clang version remain in the key. The cache format advances to `adamic-units-v2` so previous entries cannot be reused.

`TestUnitSystemHeaderProvenance` plants `_Nullable` in a header marked with `#pragma clang system_header`, with `-Wno-system-headers`, native.Flags and `-pedantic -Werror`. It compiles successfully, observes an unchanged-header cache hit, changes a header macro from 1 to 2, and observes output `2` for cached and uncached builds. Removing the system pragma makes the same extension fail with `-Wnullability-extension`; no blanket diagnostic suppression is allowed.

Correction commands (test output always saved to logs):

```bash
source /workspace/adamic-tools/env.sh
go test ./internal/native -run 'TestUnitSystemHeaderProvenance|TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokens' -count=1 -v > /tmp/adamic-clang-system-header.log 2>&1
ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout=30m > /tmp/adamic-header-oracle.log 2>&1
go test ./internal/native -count=1 -timeout=30m > /tmp/adamic-header-native.log 2>&1
go vet ./... > /tmp/adamic-header-vet.log 2>&1
# Each saved correction mutant is mapped over internal/native/units.go with a Go overlay:
go test -overlay=/tmp/adamic-header-flatten-overlay.json ./internal/native -run '^TestUnitSystemHeaderProvenance$' -count=1 > /tmp/adamic-header-flatten-mutant.log 2>&1
go test -overlay=/tmp/adamic-header-header-key-overlay.json ./internal/native -run '^TestUnitSystemHeaderProvenance$' -count=1 > /tmp/adamic-header-header-key-mutant.log 2>&1
go test -overlay=/tmp/adamic-header-flags-overlay.json ./internal/native -run '^TestUnitCacheFlagsHoldSanitizer$' -count=1 > /tmp/adamic-header-flags-mutant.log 2>&1
go test -overlay=/tmp/adamic-header-blanket-warning-overlay.json ./internal/native -run '^TestUnitSystemHeaderProvenance$' -count=1 > /tmp/adamic-header-blanket-warning-mutant.log 2>&1
```

Run `bash internal/native/clang_units_evidence/run-system-header-mutants.sh` to recreate all four correction overlays from this checkout; the runner requires the expected diagnostic/output failure, not merely a nonzero exit.

Mutant observations:

- Reintroduce `-P`: the planted system-header extension fails with `-Werror,-Wnullability-extension`. This regression deliberately proves compiler acceptance of a platform header; this diagnostic is the requested observation, rather than a behavioral miscompile mutant.
- Hash only the original unit source instead of the preprocessed snapshot: compilation/linking succeed, but after changing the header the cached program prints `1` instead of `2`. The header-dependency check catches wrong objects on executed behavior.
- Drop ordered flags from the v2 key: release objects are reused after a sanitized build request; the signed-overflow probe no longer traps and the sanitizer check fails.
- Add blanket `-Wno-nullability-extension`: the ordinary-header negative control incorrectly compiles and fails with `user-header extension must be rejected: <nil>`.

Correction status: **PASS** focused provenance/shared-state/sanitizer checks; **PASS** full internal/native; **PASS** full uncached split internal/oracle; **PASS** go vet ./...; **PASS** reproduction runner (all four expected mutant failures). Correction evidence uses the `system-header-` filename prefix. Saved correction logs and mutant bodies accompany this report. The default single-file path and all emit*.go files remain unchanged. Only devtools/clang-units is pushed; integration owns main.

## Scope and source inventory

The splitter runs after the existing emitter. It emits one shared declaration header, one state/data translation unit, main, and groups of sixteen consecutive functions. It does not recover source-module boundaries. All former top-level static identifiers get an `adamic_unit_` prefix and external linkage; runtime header inline functions retain their existing linkage. Initialized state is defined once, preserving class/shape identity, initialization order and main cleanup. Function groups and link order are deterministic. Unknown declaration syntax fails before clang. No emit*.go output was changed.

| Source | Emitted main.c lines | Bytes | Units | Header bytes |
|---|---:|---:|---:|---:|
| lint-harness | 43,197 | 3,203,114 | 34 | 309,458 |
| typescript-parser | 24,743 | 1,759,692 | 18 | 146,309 |
| cohere-typeaware | 26,403 | 1,933,616 | 19 | 195,293 |

Inventory scanned the sixteen static stage1 `main.ts` entrypoints. The largest other emitted program is `stage1/cohere/typeaware` (actual TSGoC output, checker linked), ahead of cohere/json (1,875,504 bytes). This is an entrypoint inventory, not every possible custom entry file. Lint is the requested lint-harness branch, with all five registered rules: no-debugger, no-empty, eqeqeq, no-var, no-duplicate-case. Parser includes its scanner dependency. Ordinary C inputs were emitted by the base stage0 compiler; emitter output was separately checked unchanged with the final compiler. Typeaware was emitted with EnableTSGo/TSGoC; the helper is saved as evidence.

Use `Options{Split:true, Jobs:5}` or `ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5`. Off is the default; jobs default to one. `BuildSplitTSGo` is an explicit prototype API. Existing BuildTSGo and tsgo.go remain unchanged, pending their owner’s integration decision.

## Build flags and method

Every timing below uses commit `61075f294ed0b9a1a8b9b282c9d6e626c99ad6b0`, nproc=5, cgroup cpu.max=`400000 100000`, `go version go1.27.1 linux/amd64`, `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`, node `v24.19.0`. The same box and commit were used throughout. Each raw row includes this complete build-flags line, exact ordered Flags, full before/after /proc/loadavg, cache state and instrument. Rows below select the best of three for each side; the three interleaved paired rounds remain in timings.jsonl. An individual build never exceeded five minutes. Incomplete samples from an interrupted environment connection were discarded, not combined with these samples.

Build-flags sanitized: `-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`.
Build-flags release: `-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2`.
Build-flags counted: `-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -DADAMIC_COUNT -O2`.

Runtime archives are warm for all measurements; the TSGo checker archive is prebuilt and warm. Full split runs remove only this benchmark’s object-cache directory before the timed build; incremental and no-edit runs use those objects. The whole-file path compiles main.c afresh. No program-result cache is used. Split cache is enabled (cold full / warm rebuild), not gate-uncached. C generation, Go test startup, runtime/checker preparation and process setup outside Build are excluded. Split includes splitting, all per-unit preprocessing/cache checks, object compilation and linking.

For typeaware, whole-file measurements deliberately link the **same** cached TSGo-enabled runtime and prebuilt checker as split; they measure clang on main.c rather than the current BuildTSGo API’s additional runtime recompilation. Flags for main.c are exactly native.Flags in both paths. Runtime preparation uses the existing correctly keyed cache plus ADAMIC_TSGO. These numbers do not claim a speedup of the full legacy BuildTSGo API.

Measurement instruments (no test output piped):
```bash
source /workspace/adamic-tools/env.sh
go build -buildmode=c-archive -o /tmp/adamic-clang-evidence/tsgo.a ./bridge/tsgo/archive
ADAMIC_CLANG_MEASURE=/tmp/adamic-clang-evidence ADAMIC_CLANG_PROGRAM=lint-harness go test ./internal/native -run '^TestMeasureClangUnits$' -count=1 -timeout=60m -v > /tmp/adamic-clang-measure-lint-harness.log 2>&1
ADAMIC_CLANG_MEASURE=/tmp/adamic-clang-evidence ADAMIC_CLANG_PROGRAM=typescript-parser go test ./internal/native -run '^TestMeasureClangUnits$' -count=1 -timeout=60m -v > /tmp/adamic-clang-measure-typescript-parser.log 2>&1
ADAMIC_CLANG_MEASURE=/tmp/adamic-clang-evidence ADAMIC_CLANG_PROGRAM=cohere-typeaware go test ./internal/native -run '^TestMeasureClangUnits$' -count=1 -timeout=60m -v > /tmp/adamic-clang-measure-cohere-typeaware.log 2>&1
```

The named test alternates original whole / split full / edited whole / edited split / unchanged split, for j=1 and j=nproc=5, in each round. `units_measure_test.go` is the exact instrument; preprocessing and object clang arguments are in units.go. Whole and split APIs, and literal trace commands, are also recorded per raw row.

## Full build and one-line rebuild

Seconds. Each row is a before/after comparison. Instrument refers to the exact program-specific command above, with loop and jobs shown. Build-flags are the common metadata line plus the named mode line. Load cells give before→after **one-minute load** for the selected before and selected after sample, respectively; all five /proc/loadavg fields are preserved in raw rows. Full rows are whole→cold objects; one-line rows are edited whole→cached objects. The before and after minima may come from different paired rounds.

| Loop | Before s | After s | Instrument | Build-flags/cache | Load before sample; after sample |
|---|---:|---:|---|---|---|
| lint-harness sanitized full | 14.785 | 15.850 | TestMeasureClangUnits, j=1 | sanitized; runtime/checker warm; units cold | 1.43→1.33; 1.33→1.26 |
| lint-harness sanitized one-line | 15.222 | 1.376 | TestMeasureClangUnits, j=1 | sanitized; runtime/checker warm; units warm | 1.26→1.20; 1.09→1.09 |
| lint-harness sanitized full | 14.785 | 6.012 | TestMeasureClangUnits, j=5 | sanitized; runtime/checker warm; units cold | 1.43→1.33; 1.18→1.49 |
| lint-harness sanitized one-line | 14.412 | 0.899 | TestMeasureClangUnits, j=5 | sanitized; runtime/checker warm; units warm | 1.26→1.21; 1.38→1.38 |
| lint-harness release full | 5.044 | 6.857 | TestMeasureClangUnits, j=1 | release; runtime/checker warm; units cold | 1.39→1.36; 1.11→1.10 |
| lint-harness release one-line | 4.879 | 1.057 | TestMeasureClangUnits, j=1 | release; runtime/checker warm; units warm | 1.10→1.09; 1.09→1.09 |
| lint-harness release full | 5.044 | 2.809 | TestMeasureClangUnits, j=5 | release; runtime/checker warm; units cold | 1.39→1.36; 1.23→1.13 |
| lint-harness release one-line | 5.179 | 0.541 | TestMeasureClangUnits, j=5 | release; runtime/checker warm; units warm | 1.13→1.12; 1.12→1.12 |
| lint-harness counted full | 4.965 | 6.694 | TestMeasureClangUnits, j=1 | counted; runtime/checker warm; units cold | 1.22→1.20; 1.20→1.17 |
| lint-harness counted one-line | 4.676 | 1.129 | TestMeasureClangUnits, j=1 | counted; runtime/checker warm; units warm | 1.36→1.33; 1.33→1.33 |
| lint-harness counted full | 4.965 | 2.942 | TestMeasureClangUnits, j=5 | counted; runtime/checker warm; units cold | 1.22→1.20; 1.45→1.81 |
| lint-harness counted one-line | 4.853 | 0.532 | TestMeasureClangUnits, j=5 | counted; runtime/checker warm; units warm | 1.81→1.75; 1.43→1.43 |
| typescript-parser sanitized full | 8.927 | 9.299 | TestMeasureClangUnits, j=1 | sanitized; runtime/checker warm; units cold | 1.58→1.49; 1.49→1.42 |
| typescript-parser sanitized one-line | 8.756 | 1.110 | TestMeasureClangUnits, j=1 | sanitized; runtime/checker warm; units warm | 1.20→1.18; 1.35→1.35 |
| typescript-parser sanitized full | 8.927 | 4.579 | TestMeasureClangUnits, j=5 | sanitized; runtime/checker warm; units cold | 1.58→1.49; 1.16→1.15 |
| typescript-parser sanitized one-line | 8.405 | 0.671 | TestMeasureClangUnits, j=5 | sanitized; runtime/checker warm; units warm | 1.15→1.14; 1.14→1.53 |
| typescript-parser release full | 2.487 | 3.499 | TestMeasureClangUnits, j=1 | release; runtime/checker warm; units cold | 1.18→1.18; 1.18→1.24 |
| typescript-parser release one-line | 2.531 | 0.775 | TestMeasureClangUnits, j=1 | release; runtime/checker warm; units warm | 1.19→1.19; 1.24→1.22 |
| typescript-parser release full | 2.487 | 1.456 | TestMeasureClangUnits, j=5 | release; runtime/checker warm; units cold | 1.18→1.18; 1.19→1.19 |
| typescript-parser release one-line | 2.466 | 0.442 | TestMeasureClangUnits, j=5 | release; runtime/checker warm; units warm | 1.17→1.16; 1.19→1.18 |
| typescript-parser counted full | 2.663 | 3.481 | TestMeasureClangUnits, j=1 | counted; runtime/checker warm; units cold | 1.80→1.73; 1.73→1.68 |
| typescript-parser counted one-line | 2.537 | 0.807 | TestMeasureClangUnits, j=1 | counted; runtime/checker warm; units warm | 1.17→1.17; 1.13→1.13 |
| typescript-parser counted full | 2.663 | 1.433 | TestMeasureClangUnits, j=5 | counted; runtime/checker warm; units cold | 1.80→1.73; 1.17→1.80 |
| typescript-parser counted one-line | 2.455 | 0.458 | TestMeasureClangUnits, j=5 | counted; runtime/checker warm; units warm | 1.62→1.62; 1.62→1.62 |
| cohere-typeaware sanitized full | 11.004 | 10.853 | TestMeasureClangUnits, j=1 | sanitized; runtime/checker warm; units cold | 1.76→1.64; 1.46→1.39 |
| cohere-typeaware sanitized one-line | 10.420 | 1.153 | TestMeasureClangUnits, j=1 | sanitized; runtime/checker warm; units warm | 1.39→1.33; 1.27→1.27 |
| cohere-typeaware sanitized full | 11.004 | 4.483 | TestMeasureClangUnits, j=5 | sanitized; runtime/checker warm; units cold | 1.76→1.64; 1.42→1.71 |
| cohere-typeaware sanitized one-line | 10.310 | 0.841 | TestMeasureClangUnits, j=5 | sanitized; runtime/checker warm; units warm | 1.23→1.19; 1.60→1.60 |
| cohere-typeaware release full | 3.347 | 4.354 | TestMeasureClangUnits, j=1 | release; runtime/checker warm; units cold | 1.18→1.16; 1.16→1.15 |
| cohere-typeaware release one-line | 3.176 | 0.841 | TestMeasureClangUnits, j=1 | release; runtime/checker warm; units warm | 1.25→1.23; 1.23→1.23 |
| cohere-typeaware release full | 3.347 | 1.882 | TestMeasureClangUnits, j=5 | release; runtime/checker warm; units cold | 1.18→1.16; 1.14→1.14 |
| cohere-typeaware release one-line | 3.300 | 0.566 | TestMeasureClangUnits, j=5 | release; runtime/checker warm; units warm | 1.37→1.37; 1.37→1.37 |
| cohere-typeaware counted full | 3.266 | 4.255 | TestMeasureClangUnits, j=1 | counted; runtime/checker warm; units cold | 1.76→1.76; 1.76→1.70 |
| cohere-typeaware counted one-line | 3.257 | 0.803 | TestMeasureClangUnits, j=1 | counted; runtime/checker warm; units warm | 1.70→1.64; 1.64→1.64 |
| cohere-typeaware counted full | 3.266 | 1.783 | TestMeasureClangUnits, j=5 | counted; runtime/checker warm; units cold | 1.76→1.76; 1.64→1.64 |
| cohere-typeaware counted one-line | 3.304 | 0.529 | TestMeasureClangUnits, j=5 | counted; runtime/checker warm; units warm | 1.70→1.65; 1.65→1.65 |

Observed: parallel full builds improve all three modes/programs. Sequential full builds are approximately unchanged for sanitized programs and slower for release/counted because each group repeats preprocessing and loses cross-group optimization. Incremental rebuilds still preprocess every group and always relink; they do not reach the small-fixture 0.01-second compile cost. Correct header dependencies are intentionally included in every object key.

## Clang trace

Single instrumented run per mode, separate from the best-of-three uninstrumented timings. Frontend, Optimizer and CodeGenPasses are clang total events in seconds. CodeGenPasses is backend code generation; Backend and OptModule contain nested work and must not be added to these columns. Trace wall time includes linking; ExecuteCompiler excludes the driver/linker. Optimization includes sanitizer and scalar/vector pipeline work. At 500µs granularity, tiny events are omitted from the detailed function list.

| Loop | Before wall s | After (instrumented wall) s | Frontend s | Optimization s | Code generation s | Instrument | Build-flags/cache; load |
|---|---:|---:|---:|---:|---:|---|---|
| lint-harness sanitized | 14.785 | 16.362 | 1.037 | 4.322 | 10.712 | T-lint-harness-sanitized | sanitized, warm runtime/checker; 0.82→0.86 |
| lint-harness release | 5.044 | 5.351 | 0.472 | 1.694 | 3.108 | T-lint-harness-release | release, warm runtime/checker; 1.35→1.32 |
| lint-harness counted | 4.965 | 5.014 | 0.443 | 1.583 | 2.919 | T-lint-harness-counted | counted, warm runtime/checker; 1.24→1.22 |
| typescript-parser sanitized | 8.927 | 9.221 | 0.600 | 2.219 | 6.147 | T-typescript-parser-sanitized | sanitized, warm runtime/checker; 1.69→1.58 |
| typescript-parser release | 2.487 | 2.661 | 0.276 | 0.789 | 1.522 | T-typescript-parser-release | release, warm runtime/checker; 1.25→1.23 |
| typescript-parser counted | 2.663 | 2.521 | 0.254 | 0.765 | 1.434 | T-typescript-parser-counted | counted, warm runtime/checker; 1.16→1.16 |
| cohere-typeaware sanitized | 11.004 | 10.611 | 0.600 | 2.514 | 7.005 | T-cohere-typeaware-sanitized | sanitized, warm runtime/checker; 1.90→1.76 |
| cohere-typeaware release | 3.347 | 4.125 | 0.303 | 1.332 | 2.270 | T-cohere-typeaware-release | release, warm runtime/checker; 1.18→1.18 |
| cohere-typeaware counted | 3.266 | 3.874 | 0.315 | 1.050 | 2.310 | T-cohere-typeaware-counted | counted, warm runtime/checker; 1.98→1.90 |

Biggest backend function events (OptFunction; time includes that function’s machine-code pass pipeline):

| Program/mode | Biggest functions, seconds |
|---|---|
| lint-harness sanitized | adamic_function_136_Statements_statement 1.209; adamic_function_132_Statements_classDeclaration 0.892; main 0.758 |
| lint-harness release | main 0.587; adamic_function_132_Statements_classDeclaration 0.170; adamic_function_33_Scanner_scan 0.138 |
| lint-harness counted | main 0.560; adamic_function_132_Statements_classDeclaration 0.167; adamic_function_33_Scanner_scan 0.130 |
| typescript-parser sanitized | adamic_function_122_Statements_statement 1.107; adamic_function_118_Statements_classDeclaration 0.751; main 0.470 |
| typescript-parser release | main 0.344; adamic_function_33_Scanner_scan 0.133; adamic_function_122_Statements_statement 0.094 |
| typescript-parser counted | main 0.322; adamic_function_33_Scanner_scan 0.121; adamic_function_122_Statements_statement 0.082 |
| cohere-typeaware sanitized | adamic_function_124_Statements_statement 1.467; adamic_function_120_Statements_classDeclaration 0.978; main 0.572 |
| cohere-typeaware release | main 0.550; adamic_function_120_Statements_classDeclaration 0.179; adamic_function_33_Scanner_scan 0.137 |
| cohere-typeaware counted | adamic_function_33_Scanner_scan 0.421; main 0.372; adamic_function_120_Statements_classDeclaration 0.228 |

Observed: sanitized code generation dominates, then optimization, then the front end. Statements.statement and Statements.classDeclaration dominate sanitized function events; main and Scanner.scan become relatively prominent without sanitizers. Function durations are nested events, not additional wall time. Sanitized main.c is costly primarily because it produces large instrumented backend functions, not because parsing source text is the largest phase.

Exact trace instruments (the files and archive hashes identify this run):

T-lint-harness-sanitized:
```text
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all -ftime-trace=/tmp/adamic-clang-evidence/lint-harness-sanitized-trace.json -ftime-trace-granularity=500 -I /tmp/adamic-clang-evidence/cache/adamic/runtime/37daa12a0b7195d790a596409e073f72de4068614061c7869f71b4955a5cb49d -o /tmp/adamic-clang-evidence/program /tmp/adamic-clang-evidence/main.c -Xlinker --whole-archive /tmp/adamic-clang-evidence/cache/adamic/runtime/37daa12a0b7195d790a596409e073f72de4068614061c7869f71b4955a5cb49d/runtime.a -Xlinker --no-whole-archive -lm
```

T-lint-harness-release:
```text
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -ftime-trace=/tmp/adamic-clang-evidence/lint-harness-release-trace.json -ftime-trace-granularity=500 -I /tmp/adamic-clang-evidence/cache/adamic/runtime/cb98d35f1ea6b8ad99018110c4692ef6ab8ce936f8c1c72c626c683511ac2314 -o /tmp/adamic-clang-evidence/program /tmp/adamic-clang-evidence/main.c -Xlinker --whole-archive /tmp/adamic-clang-evidence/cache/adamic/runtime/cb98d35f1ea6b8ad99018110c4692ef6ab8ce936f8c1c72c626c683511ac2314/runtime.a -Xlinker --no-whole-archive -lm
```

T-lint-harness-counted:
```text
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -DADAMIC_COUNT -O2 -ftime-trace=/tmp/adamic-clang-evidence/lint-harness-counted-trace.json -ftime-trace-granularity=500 -I /tmp/adamic-clang-evidence/cache/adamic/runtime/8248a0669571be5ef4a9ceb8951713b09814c4ff7b619dbdec8dca79877527e8 -o /tmp/adamic-clang-evidence/program /tmp/adamic-clang-evidence/main.c -Xlinker --whole-archive /tmp/adamic-clang-evidence/cache/adamic/runtime/8248a0669571be5ef4a9ceb8951713b09814c4ff7b619dbdec8dca79877527e8/runtime.a -Xlinker --no-whole-archive -lm
```

T-typescript-parser-sanitized:
```text
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all -ftime-trace=/tmp/adamic-clang-evidence/typescript-parser-sanitized-trace.json -ftime-trace-granularity=500 -I /tmp/adamic-clang-evidence/cache/adamic/runtime/37daa12a0b7195d790a596409e073f72de4068614061c7869f71b4955a5cb49d -o /tmp/adamic-clang-evidence/program /tmp/adamic-clang-evidence/main.c -Xlinker --whole-archive /tmp/adamic-clang-evidence/cache/adamic/runtime/37daa12a0b7195d790a596409e073f72de4068614061c7869f71b4955a5cb49d/runtime.a -Xlinker --no-whole-archive -lm
```

T-typescript-parser-release:
```text
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -ftime-trace=/tmp/adamic-clang-evidence/typescript-parser-release-trace.json -ftime-trace-granularity=500 -I /tmp/adamic-clang-evidence/cache/adamic/runtime/cb98d35f1ea6b8ad99018110c4692ef6ab8ce936f8c1c72c626c683511ac2314 -o /tmp/adamic-clang-evidence/program /tmp/adamic-clang-evidence/main.c -Xlinker --whole-archive /tmp/adamic-clang-evidence/cache/adamic/runtime/cb98d35f1ea6b8ad99018110c4692ef6ab8ce936f8c1c72c626c683511ac2314/runtime.a -Xlinker --no-whole-archive -lm
```

T-typescript-parser-counted:
```text
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -DADAMIC_COUNT -O2 -ftime-trace=/tmp/adamic-clang-evidence/typescript-parser-counted-trace.json -ftime-trace-granularity=500 -I /tmp/adamic-clang-evidence/cache/adamic/runtime/8248a0669571be5ef4a9ceb8951713b09814c4ff7b619dbdec8dca79877527e8 -o /tmp/adamic-clang-evidence/program /tmp/adamic-clang-evidence/main.c -Xlinker --whole-archive /tmp/adamic-clang-evidence/cache/adamic/runtime/8248a0669571be5ef4a9ceb8951713b09814c4ff7b619dbdec8dca79877527e8/runtime.a -Xlinker --no-whole-archive -lm
```

T-cohere-typeaware-sanitized:
```text
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all -ftime-trace=/tmp/adamic-clang-evidence/cohere-typeaware-sanitized-trace.json -ftime-trace-granularity=500 -I /tmp/adamic-clang-evidence/cache/adamic/runtime/c01403b2ff55b7b1f0708c112f7086811ddbba958a61025e2e60dd567342ff47 -o /tmp/adamic-clang-evidence/program /tmp/adamic-clang-evidence/main.c -Xlinker --whole-archive /tmp/adamic-clang-evidence/cache/adamic/runtime/c01403b2ff55b7b1f0708c112f7086811ddbba958a61025e2e60dd567342ff47/runtime.a -Xlinker --no-whole-archive /tmp/adamic-clang-evidence/tsgo.a -lpthread -ldl -lm
```

T-cohere-typeaware-release:
```text
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -ftime-trace=/tmp/adamic-clang-evidence/cohere-typeaware-release-trace.json -ftime-trace-granularity=500 -I /tmp/adamic-clang-evidence/cache/adamic/runtime/552c5e8f616e73ac8c5dadcfa52187e613e14d09e0138c55ffd04fa08665d1e1 -o /tmp/adamic-clang-evidence/program /tmp/adamic-clang-evidence/main.c -Xlinker --whole-archive /tmp/adamic-clang-evidence/cache/adamic/runtime/552c5e8f616e73ac8c5dadcfa52187e613e14d09e0138c55ffd04fa08665d1e1/runtime.a -Xlinker --no-whole-archive /tmp/adamic-clang-evidence/tsgo.a -lpthread -ldl -lm
```

T-cohere-typeaware-counted:
```text
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -DADAMIC_COUNT -O2 -ftime-trace=/tmp/adamic-clang-evidence/cohere-typeaware-counted-trace.json -ftime-trace-granularity=500 -I /tmp/adamic-clang-evidence/cache/adamic/runtime/451d5f835c4b32ff7550fbf52f1a38f4a11263f6e82d177ddd60826dc88abbd6 -o /tmp/adamic-clang-evidence/program /tmp/adamic-clang-evidence/main.c -Xlinker --whole-archive /tmp/adamic-clang-evidence/cache/adamic/runtime/451d5f835c4b32ff7550fbf52f1a38f4a11263f6e82d177ddd60826dc88abbd6/runtime.a -Xlinker --no-whole-archive /tmp/adamic-clang-evidence/tsgo.a -lpthread -ldl -lm
```

## Correctness and cache

The object key hashes the exact preprocessed snapshot and unit filename, all ordered native.Flags, compiler path and full --version, platform (GOOS/GOARCH via runtimeKey), and a cache-format domain. Source and shared/runtime headers are snapshotted; clang -E observes transitive system/ambient include dependencies on every lookup and preserves system-header line markers. The same preprocessed bytes are then compiled, so dependencies are not reopened between key creation and compilation. Macro flags remain in the key but are omitted from cpp-output compilation because they already took effect. A stable debug-prefix mapping removes random temporary directories from emitted debug paths. Runtime flags and checker ABI/header remain in the existing runtime cache key. Linking is never cached; checker/archive contents are read again for each link.

Cache publication uses a temporary directory on the cache filesystem and atomic rename. Per-key in-process locks avoid duplicate work; competing processes publish complete entries. ADAMIC_GATE_UNCACHED=1 bypasses all new object-cache hits and publication. It does not change the preexisting runtime archive cache contract. Cache-vs-uncached equality means executed output, not ELF bytes/debug paths. Cached objects are not authenticated against a hostile cache writer, matching the existing local-cache trust model.

Validation commands and observations:
```bash
ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout=30m > /tmp/adamic-clang-oracle-latest.log 2>&1
go test ./internal/native -count=1 -timeout=30m > /tmp/adamic-clang-native-latest.log 2>&1
go vet ./... > /tmp/adamic-clang-vet-latest.log 2>&1
ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 ADAMIC_GATE_UNCACHED=1 go test ./stage1/typescript/parser -run "TestExpressionsAgree|TestWholeGeneratedAgrees|TestEveryTypeNodeKindAgrees" -count=1 -timeout=30m -v > /tmp/adamic-clang-parser-final.log 2>&1
# Run in the detached lint worktree at f4d98ca with the saved native overlay:
ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 ADAMIC_GATE_UNCACHED=1 go test -overlay=/tmp/adamic-clang-lint-overlay.json ./stage1/cohere/lint -run "TestRulesAgree|TestCompleteSuggestionSerialization|TestSuggestionAlongsideAutomaticFix" -count=1 -timeout=30m -v > /tmp/adamic-clang-lint-final.log 2>&1
ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/adamic-clang-evidence/tsgo.a go test ./internal/native -run "TestSplitTSGoAgrees|TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokensDoNotRewriteLiterals" -count=1 -v > /tmp/adamic-clang-tsgo-parity-final.log 2>&1
```

Parser: 58 expression cases (17,463 identical bytes), 42 type-node kinds (250 bytes), 68 generated whole files (93,323 bytes), comparing the external Go/parser oracle and relevant Node behavior with native output. Lint: all five registered rules, 757,785 identical bytes across Go, Node, emitted JS and native; complete suggestion serialization 735 bytes and simultaneous automatic fix/suggestion 720 bytes. Tests explicitly exclude documented unsupported recovery cases. TSGo typeaware: single-file vs split cached and uncached output identical (86 bytes) on the supported .a-root bridge fixture. Shared-state test compares expected `1 2 2` output for cached and uncached split. Existing parser and lint planted mutants are caught in their logs.

Earlier full oracle found that the splitter unnecessarily rejected valid static helper names planted by array/library mutants. The naming restriction was removed; external names still receive the unique prefix. The entire uncached oracle then passed, including a final rerun on the measured implementation. An initial TestDotARename failed because dependency-download chatter violated its silent-build stderr assertion; the final lint command above uses the requested runtime parity tests after warmup. No claim is made that all lint package tests or the complete upstream pinned TypeScript corpus were run.

Mutants use Go overlays; production code is never edited. The relocatable `bash internal/native/clang_units_evidence/run-mutants.sh` recreates all four overlays and requires the observed failure message, rather than accepting unrelated compile failures:
```bash
go test -overlay=/tmp/adamic-clang-flags-latest-overlay.json ./internal/native -run "^TestUnitCacheFlagsHoldSanitizer$" -count=1 > /tmp/adamic-clang-flags-mutant-latest.log 2>&1
go test -overlay=/tmp/adamic-clang-state-latest-overlay.json ./internal/native -run "^TestUnitsPreserveSharedState$" -count=1 > /tmp/adamic-clang-state-mutant-latest.log 2>&1
```

Flags-key mutant replaces `runtimeKey(files, append([]string{"adamic-units-v1"}, flags...), compiler, version)` with `runtimeKey(files, []string{"adamic-units-v1"}, compiler, version)`. Release then sanitized builds have the same preprocessed C; the wrong release object links successfully and executes without UBSan. TestUnitCacheFlagsHoldSanitizer fails because signed integer overflow no longer traps. The correct key makes UBSan report signed integer overflow and exit nonzero. This is a behavioral cache-soundness mutant; -Werror does not kill it.

Shared-state mutant replaces the extern state declaration with a per-unit static definition and removes its single state.c definition. Compilation and linking succeed, but main reads zero instead of two: output `1 2 0`, exit 1. TestUnitsPreserveSharedState fails. The mutant overlay bodies are saved as .go.txt files for inspection/reproduction.

Literal-rewrite mutant replaces identifiers inside string tokens; TestSplitTokensDoNotRewriteLiterals fails with `literal changed`. Unit-order mutant swaps the first two output units every second split; the repeated-split check fails with `split changed`. Both run through Go overlays against their named tests, with separate failing logs. This proves the lightweight lexer and deterministic-output assertions are able to fail.

## What the emitter owner would need to change

The measured edits preserve declarations and numbering: lint rules/no-var/rule.a changes `parent >= 0` to `parent >= 1`; parser grammar.ts changes `return 14` to `return 15`; typeaware unary_minus.ts changes `cursor = 0` to `cursor = 1`. The measurement harness proves each changes exactly one translation-unit body and leaves the shared header unchanged. These changes are timing inputs, not correctness fixtures.

Counterexample: change the parser PercentToken return from `14` to equivalent `[14].length`. The emitter adds temporaries: shared header changes, 12/18 unit sources change, and all 18 preprocessed object keys invalidate. This is observed in preflight-final.log. Therefore this branch is a **function-group prototype**, not a general source-module incremental compiler. It always rebuilds dependencies correctly rather than promising one object after an arbitrary line edit.

Precise follow-on changes, not made here:

1. Preserve source-module identity on IR functions/locals or provide a stable emitter unit map; assign deterministic module-qualified symbol IDs rather than whole-program ordinal names.
2. Make temporary and inline-cache numbering local to a function/module. Current emitter-wide temporary/cache counters cause edits to renumber subsequent definitions.
3. Give string constants, shape descriptors, class descriptors and other shared data stable content/owner identities and one definition. A module-local change must not renumber unrelated data. Preserve pointer identity where runtime uses it.
4. Emit per-unit types/declarations from actual dependency sets, plus stable shared ABI types. A single omnibus header deliberately invalidates every unit on any declaration change.
5. Keep main’s initialization and global-release order identical. Track interprocedural borrow, region and throw facts as dependencies: a changed proof or signature must rebuild affected callers, even when their source bytes do not change.
6. Route the existing TSGo build entrypoint to the prototype only after its owner approves the option and matching checker/runtime policy.

No emitter/lower/oracle code was changed to obtain these results. Test helpers must still be accepted if the oracle plants valid C functions.

## Reproduction and evidence

`clang_units_evidence/timings.jsonl` contains all 252 measured rows. The nine original traces, six timing C inputs and equivalent parser churn input are gzip-compressed with deterministic headers. trace-summary.json provides phase totals and top-five backend functions. The checker archive is intentionally omitted; rebuild it with the command above. Logs preserve setup, gates, external parity, preflight, mutants and single-file byte identity.

To reproduce the timing inputs from this checkout:
```bash
source /workspace/adamic-tools/env.sh
mkdir -p /tmp/adamic-clang-evidence
python3 - <<'PY'
import gzip
from pathlib import Path
for path in Path('internal/native/clang_units_evidence').glob('*.c.gz'):
    Path('/tmp/adamic-clang-evidence', path.name[:-3]).write_bytes(gzip.decompress(path.read_bytes()))
PY
go build -buildmode=c-archive -o /tmp/adamic-clang-evidence/tsgo.a ./bridge/tsgo/archive
# Then run each exact TestMeasureClangUnits command above.
```

Toolchain setup: `bash cloud/setup.sh > /tmp/adamic-clang-setup.log 2>&1`; setup log records Go ready 0s, clang ready 0s, Node ready 0s, submodules 1s, cache warm 177s, done 177s; nproc 5. The environment prints `/workspace/adamic-tools/env.sh`; `/opt/adamic-tools/env.sh` does not exist on this box. Setup timing is provisioning evidence, not a before/after compiler benchmark.

Final gate status: **PASS** full uncached split oracle; **PASS** full internal/native; **PASS** go vet ./...; **PASS** selected parser/lint stage1 byte-parity tests and cached/uncached checker parity; all four deliberately broken implementations fail their named checks. See corresponding saved logs. No full repository go test ./... gate was run.
