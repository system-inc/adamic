Built: selected-rule split builds with GOMAXPROCS jobs, concurrent correct/mutant builds, independent whole-file witnesses.\
Commits: merge a79f985bb18650f4224f19c02e2bf71b4bfeaaa9; implementation 9c1c3349da798bc2befed012d98b91bdba1153b9.\
Commands and outputs: 72 interleaved author checks passed; parity, helper parity, cache guards and ordinary uncached all-rule checks passed.\
Mutants: seven lint cache omissions, eight native split/header mutants and four witness/concurrency guards caught; all owned mutants caught.\
Not covered: full repository gate or fifty-rule fleet; byte edit is one appended space; helper edits invalidate broadly.

The branch is `devtools/lint-rule-check`. It merged `origin/area/developer-tools` at `a50bb48389bb470c1fcb7cc4754b988cbe3e2a11`, without rebasing. Only this branch is pushed.

The fast path explicitly uses `native.Options{Sanitize: true, Split: true, Jobs: runtime.GOMAXPROCS(0)}` and starts the correct and mutant builds concurrently. Native port keys now include split mode and jobs. An inherited `ADAMIC_NATIVE_SPLIT` switch is cleared by the lint harness so existing full-registration builds and `ADAMIC_GATE_UNCACHED=1` retain whole-file compilation. The split compiler and its cache implementation were merged unchanged.

Best of three, wall-clock seconds on the same implementation commit, with whole-file and split commands interleaved. Each rule/round/compiler has a separate initially empty user cache; subsequent edits reuse that cache. The Go build cache is primed and fixed. The ordinary compiler baseline builds correct and mutant ports sequentially, matching the preceding fast path. The clang wrapper only records and forwards compiler invocations; both paths use it.

| loop | before whole-file | after split | instrument |
|---|---:|---:|---|
| no-var cold | 70.811 | 39.016 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-var`; before adds `-rule-whole` |
| no-var warm-byte-edit | 52.575 | 15.231 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-var -rule-byte-change`; before adds `-rule-whole` |
| no-var warm-unchanged | 6.978 | 7.254 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-var`; before adds `-rule-whole` |
| no-var helper-edit | 47.901 | 21.253 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-var -rule-helper-change`; before adds `-rule-whole` |
| no-empty cold | 62.162 | 39.267 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-empty`; before adds `-rule-whole` |
| no-empty warm-byte-edit | 47.790 | 17.570 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-empty -rule-byte-change`; before adds `-rule-whole` |
| no-empty warm-unchanged | 6.978 | 7.570 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-empty`; before adds `-rule-whole` |
| no-empty helper-edit | 43.776 | 21.256 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-empty -rule-helper-change`; before adds `-rule-whole` |
| eqeqeq cold | 69.799 | 38.605 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule eqeqeq`; before adds `-rule-whole` |
| eqeqeq warm-byte-edit | 47.680 | 15.631 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule eqeqeq -rule-byte-change`; before adds `-rule-whole` |
| eqeqeq warm-unchanged | 7.211 | 7.339 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule eqeqeq`; before adds `-rule-whole` |
| eqeqeq helper-edit | 47.587 | 24.868 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule eqeqeq -rule-helper-change`; before adds `-rule-whole` |

Build-flags line for every row: commit=9c1c3349da798bc2befed012d98b91bdba1153b9; nproc=5; cpu.max=400000 100000; go=go version go1.27.1 linux/amd64; clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOMAXPROCS=4; jobs=4 per split build; two builds concurrent; flags=-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all; GOCACHE=/home/agent/.cache/go-build. Cache mode and load before/after for each winning sample follow. [All 72 rows](measurements.json) also carry the full line, exact command, exit code, cache directory and trace.

| winning sample | round | cache | load before | load after | log |
|---|---:|---|---|---|---|
| no-var cold whole | 1 | empty lint/native caches | `3.20 1.96 1.20 1/162 119991` | `1.70 1.76 1.19 2/161 121752` | [no-var-1-cold-whole.log](no-var-1-cold-whole.log) |
| no-var cold split | 3 | empty lint/native caches | `1.34 1.54 1.54 2/159 190833` | `1.64 1.58 1.56 1/159 192690` | [no-var-3-cold-split.log](no-var-3-cold-split.log) |
| no-var warm-byte-edit whole | 1 | cached | `1.74 1.76 1.21 1/161 123638` | `1.29 1.63 1.19 1/161 125129` | [no-var-1-warm-byte-edit-whole.log](no-var-1-warm-byte-edit-whole.log) |
| no-var warm-byte-edit split | 2 | cached | `1.58 1.68 1.47 2/160 159623` | `2.26 1.83 1.52 1/164 161204` | [no-var-2-warm-byte-edit-split.log](no-var-2-warm-byte-edit-split.log) |
| no-var warm-unchanged whole | 3 | cached | `1.91 1.62 1.57 1/160 195687` | `1.84 1.61 1.56 1/159 196587` | [no-var-3-warm-unchanged-whole.log](no-var-3-warm-unchanged-whole.log) |
| no-var warm-unchanged split | 3 | cached | `1.84 1.61 1.56 1/159 196587` | `1.79 1.61 1.56 1/159 197489` | [no-var-3-warm-unchanged-split.log](no-var-3-warm-unchanged-split.log) |
| no-var helper-edit whole | 3 | cached | `1.79 1.61 1.56 1/159 197489` | `1.37 1.52 1.53 1/158 198964` | [no-var-3-helper-edit-whole.log](no-var-3-helper-edit-whole.log) |
| no-var helper-edit split | 3 | cached | `1.37 1.52 1.53 1/158 198964` | `2.02 1.67 1.58 1/162 200545` | [no-var-3-helper-edit-split.log](no-var-3-helper-edit-split.log) |
| no-empty cold whole | 1 | empty lint/native caches | `1.98 1.76 1.29 1/161 131527` | `1.33 1.61 1.27 1/162 133285` | [no-empty-1-cold-whole.log](no-empty-1-cold-whole.log) |
| no-empty cold split | 3 | empty lint/native caches | `1.29 1.52 1.53 1/158 202301` | `2.14 1.69 1.59 1/158 204158` | [no-empty-3-cold-split.log](no-empty-3-cold-split.log) |
| no-empty warm-byte-edit whole | 3 | cached | `2.14 1.69 1.59 1/158 204158` | `2.21 1.73 1.60 1/159 205618` | [no-empty-3-warm-byte-edit-whole.log](no-empty-3-warm-byte-edit-whole.log) |
| no-empty warm-byte-edit split | 2 | cached | `1.39 1.67 1.54 1/162 171151` | `1.70 1.72 1.56 1/163 172725` | [no-empty-2-warm-byte-edit-split.log](no-empty-2-warm-byte-edit-split.log) |
| no-empty warm-unchanged whole | 3 | cached | `3.59 2.05 1.71 1/159 207147` | `3.11 2.00 1.70 1/159 208044` | [no-empty-3-warm-unchanged-whole.log](no-empty-3-warm-unchanged-whole.log) |
| no-empty warm-unchanged split | 3 | cached | `3.11 2.00 1.70 1/159 208044` | `3.02 2.00 1.70 1/159 208946` | [no-empty-3-warm-unchanged-split.log](no-empty-3-warm-unchanged-split.log) |
| no-empty helper-edit whole | 3 | cached | `3.02 2.00 1.70 2/159 208946` | `1.95 1.86 1.67 1/159 210421` | [no-empty-3-helper-edit-whole.log](no-empty-3-helper-edit-whole.log) |
| no-empty helper-edit split | 3 | cached | `1.95 1.86 1.67 1/159 210421` | `2.24 1.93 1.69 1/159 211999` | [no-empty-3-helper-edit-split.log](no-empty-3-helper-edit-split.log) |
| eqeqeq cold whole | 2 | empty lint/native caches | `2.26 1.92 1.64 1/161 177581` | `1.39 1.73 1.60 1/161 179342` | [eqeqeq-2-cold-whole.log](eqeqeq-2-cold-whole.log) |
| eqeqeq cold split | 3 | empty lint/native caches | `1.43 1.74 1.64 1/160 213753` | `1.77 1.77 1.66 1/161 215625` | [eqeqeq-3-cold-split.log](eqeqeq-3-cold-split.log) |
| eqeqeq warm-byte-edit whole | 2 | cached | `1.38 1.66 1.58 1/160 181212` | `1.16 1.56 1.54 1/159 182674` | [eqeqeq-2-warm-byte-edit-whole.log](eqeqeq-2-warm-byte-edit-whole.log) |
| eqeqeq warm-byte-edit split | 3 | cached | `1.33 1.65 1.62 2/163 217079` | `1.65 1.70 1.64 1/163 218607` | [eqeqeq-3-warm-byte-edit-split.log](eqeqeq-3-warm-byte-edit-split.log) |
| eqeqeq warm-unchanged whole | 2 | cached | `1.20 1.55 1.54 1/159 184204` | `1.24 1.54 1.54 1/159 185104` | [eqeqeq-2-warm-unchanged-whole.log](eqeqeq-2-warm-unchanged-whole.log) |
| eqeqeq warm-unchanged split | 2 | cached | `1.24 1.54 1.54 1/159 185104` | `1.54 1.60 1.56 1/159 186003` | [eqeqeq-2-warm-unchanged-split.log](eqeqeq-2-warm-unchanged-split.log) |
| eqeqeq helper-edit whole | 1 | cached | `1.46 1.60 1.39 1/160 151438` | `1.46 1.56 1.39 1/160 152931` | [eqeqeq-1-helper-edit-whole.log](eqeqeq-1-helper-edit-whole.log) |
| eqeqeq helper-edit split | 2 | cached | `1.31 1.52 1.54 2/159 187490` | `2.04 1.68 1.59 1/159 189076` | [eqeqeq-2-helper-edit-split.log](eqeqeq-2-helper-edit-split.log) |

Split object commands additionally use `-Wno-gnu-line-marker`, `-fdebug-prefix-map=<snapshot>=/adamic-units`, `-x cpp-output -c`; every exact compiler invocation is retained in its trace. Whole-file commands retain the ordinary build flags.

The byte edit appends exactly one space to the private rule module. Its native artifact key misses, but generated C remains identical: all three rounds compile zero objects for both correct and mutant ports. This measures source-byte invalidation and object reuse, not a behavior-changing edit. The unchanged check reuses completed native artifacts and still executes both sanitized binaries.

The helper edit adds a used numeric identity helper and routes `visit` through it. It changes declarations and program numbering while preserving findings. Counts below are actual clang `-c -x cpp-output` calls across the concurrent correct/mutant pair, not cache file counts. All 34 unit names (32 function groups, state and main) compile again. Shared unchanged objects need only one physical compilation across the pair; additional calls represent differing correct/mutant objects.

| rule | cold object builds, rounds 1/2/3 | byte edit | unchanged | helper object builds, rounds 1/2/3 |
|---|---|---|---|---|
| no-var | 35/35/35 | 0/0/0 | 0/0/0 | 35/35/35 |
| no-empty | 35/35/35 | 0/0/0 | 0/0/0 | 35/35/35 |
| eqeqeq | 44/44/44 | 0/0/0 | 0/0/0 | 44/44/44 |

Every `.clang.json` retains compiler arguments, working directory and start time. The two preprocessing windows overlap, confirming concurrent builds. Each split build uses four jobs; together they can request eight compiler workers against the four-core quota. No compiler scheduling or emitter changes were made.

Validation commands (all test output is saved, not piped):

```bash
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/split-evidence/measure.py
python3 stage1/cohere/lint/split-evidence/checks.py
```

[checks.py](checks.py) records the exact filtered Go commands. [parity.log](parity.log) covers selected-versus-full registration, split-versus-fresh-whole native, cached-versus-uncached answers, and compiler selection even with `ADAMIC_NATIVE_SPLIT=1` inherited. [helper-parity.log](helper-parity.log) repeats the complete four-runtime and fresh whole-file comparison after the helper edit. [ordinary-uncached-all-rule.log](ordinary-uncached-all-rule.log) runs `ADAMIC_GATE_UNCACHED=1 ADAMIC_NATIVE_SPLIT=1 go test ./stage1/cohere/lint -count=1 -timeout 30m -v -run '^TestRulesAgree$'`, including every upstream package and inherited corpus row; the build log must say `split=false jobs=0`.

Focused ordinary-build oracle: `GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 ADAMIC_NATIVE_SPLIT=0 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(functions|closures|generic_functions)\.a$' -count=1 -timeout 30m -v` passed all three intended fixture subtests: [filtered-oracle.log](filtered-oracle.log). `go vet ./stage1/cohere/lint ./cmd/adamic-lint-check` passed: [vet.log](vet.log).

Observed byte-identical corpus sizes:

| rule | unedited | helper edit |
|---|---:|---:|
| no-var | 427484 | 427484 |
| no-empty | 427760 | 427760 |
| eqeqeq | 437914 | 437914 |

Mutants run and caught:

| mutant | change | check and observation |
|---|---|---|
| lint oracle | drop `oracle/adapter/no-var`; no-var adapter becomes silent | `TestLintCacheInvalidation/oracle` fails with `stale oracle answer` and `cached_matches_before=true`; [cache-guards.log](cache-guards.log) |
| lint capture | drop `capture/filter`; fires filter becomes stays-silent | `TestLintCacheInvalidation/capture` fails with `stale capture answer` and `cached_matches_before=true`; [cache-guards.log](cache-guards.log) |
| lint go-output | drop `go-output/manifest`; input var becomes let | `TestLintCacheInvalidation/go-output` fails with `stale go-output answer` and `cached_matches_before=true`; [cache-guards.log](cache-guards.log) |
| lint port | drop `port/modules`; main prints an extra line | `TestLintCacheInvalidation/port` fails with `stale port answer` and `cached_matches_before=true`; [cache-guards.log](cache-guards.log) |
| lint node | drop `node/modules`; main prints an extra line | `TestLintCacheInvalidation/node` fails with `stale node answer` and `cached_matches_before=true`; [cache-guards.log](cache-guards.log) |
| lint javascript | drop `javascript/modules`; emitted module prints an extra line | `TestLintCacheInvalidation/javascript` fails with `stale javascript answer` and `cached_matches_before=true`; [cache-guards.log](cache-guards.log) |
| lint javascript-build | drop `javascript-build/modules`; main prints an extra line | `TestLintCacheInvalidation/javascript-build` fails with `stale javascript-build answer` and `cached_matches_before=true`; [cache-guards.log](cache-guards.log) |
| native-flags | drop ordered flags from object key | `TestUnitCacheFlagsHoldSanitizer`: `sanitized rebuild reused uninstrumented object`; [native-flags-mutant.log](native-flags-mutant.log) |
| native-state | duplicate state across units | `TestUnitsPreserveSharedState`: `shared state (uncached=0)`; [native-state-mutant.log](native-state-mutant.log) |
| native-literal | rewrite identifiers inside string tokens | `TestSplitTokensDoNotRewriteLiterals`: `literal changed`; [native-literal-mutant.log](native-literal-mutant.log) |
| native-determinism | swap unit order on repeated split | `TestUnitsPreserveSharedState`: `split changed`; [native-determinism-mutant.log](native-determinism-mutant.log) |
| native-header-flatten | remove system-header provenance markers | `TestUnitSystemHeaderProvenance`: `system-header provenance lost`; [native-header-flatten-mutant.log](native-header-flatten-mutant.log) |
| native-header-header-key | drop header bytes from object key; change header macro 1 to 2 | `TestUnitSystemHeaderProvenance`: `output="1\n" want="2\n"`; [native-header-header-key-mutant.log](native-header-header-key-mutant.log) |
| native-header-flags | drop flags and reuse release object | `TestUnitCacheFlagsHoldSanitizer`: `sanitized rebuild reused uninstrumented object`; [native-header-flags-mutant.log](native-header-flags-mutant.log) |
| native-header-blanket-warning | disable user-code diagnostics globally | `TestUnitSystemHeaderProvenance`: `user-header extension must be rejected`; [native-header-blanket-warning-mutant.log](native-header-blanket-warning-mutant.log) |
| inherited-split-switch | remove clearing of inherited opt-in switch | `TestWholeBuildSelection`: `inherited split switch can replace the whole-file witness`; [inherited-split-switch-mutant.log](inherited-split-switch-mutant.log) |
| uncached-selection | remove uncached guard from selected build options | `TestWholeBuildSelection`: `uncached build options:`; [uncached-selection-mutant.log](uncached-selection-mutant.log) |
| serial-build | remove goroutine launches; findings still agree and owned mutants are caught | `compilation_overlap`: `correct and mutant compilation windows do not overlap`; [serial-build-mutant.log](serial-build-mutant.log) |
| parity-output | add puts to only the actual whole-file C witness, compiled and run cleanly | `TestSplitWholeParity/no-var`: `split and whole-file outputs differ`; [parity-output-mutant.log](parity-output-mutant.log) |

Every measured author check also runs its owned rule mutant on Node, emitted JavaScript and sanitized native. The mutants are `var declaration suppressed`, `empty function body reported`, and `suggestion applied as fix`; none survived. Native split guards pass without mutations in [native-split-guards.log](native-split-guards.log).

Setup completed successfully: [setup.log](setup.log), including tool timings and its complete build-flags/load line. It reports done at 44.164 seconds, nproc 5, cpu.max `400000 100000`, Go 1.27.1, clang 20.1.8, Node 24.19.0; test binaries were deferred by the merged setup script.

An initial pre-commit check was refused because harness bytes changed during its oracle build: [input-change-refusal.log](input-change-refusal.log). After edits stopped, the checks passed. This refusal was not a benchmark sample.

Limits: this checkout has five registered baseline rules, not a completed fifty-rule fleet. No full repository `go test ./...`, compiler corpus, throughput or profile gate was rerun. The focused lint and native checks above were run. The broad helper invalidation is observed; stable module-qualified numbering and smaller declaration dependency sets remain compiler work, as described in CLANG_UNITS.md. No rule directory, submodule, registry validation or native implementation was edited by this fast-path change.
