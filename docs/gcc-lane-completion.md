GCC comparison builds now keep warnings without making them fatal; production clang flags remain unchanged.
Continued `36f99309f0cd383c9684c94effffc6a0e85cae8b`; implementation and measurement source is `e627f42`, `a5c569e`, and measurement source `d0c559f64f794100e24c7fb85bc9c952a65909dc`.
Uncached read_order, 36 exception fixtures, and all 328 lowered fixtures passed in release and sanitized GCC builds, with zero byte or exit-code disagreements.
The real dedication-byte mutant was caught under GCC; independent warning-policy mutants failed their intended assertions; no cache was added.
Emitter/runtime code, other architectures, unlowered programs, and the complete repository gate are outside this report's coverage.

# GCC comparison lane completed

This supersedes the blocked-build conclusions in [the previous worker report](gcc-lane.md). The old report remains historical evidence, including its unsuccessful strict-GCC measurements. This continuation does not fix any emitter or runtime warning.

## Build and comparison behavior

Only `internal/oracle/gcc_lane_test.go` changes compilation. `native.Options{Compiler: "gcc"}` still supplies the lane's ordinary native flags, but the lane removes exactly `-Werror` before invoking GCC. Clang keeps `-Werror` and its existing suppressions. Production `native.Flags`, `native.Build`, emission, lowering, runtime sources, and `oracle_test.go` are unchanged.

The lane copies the checkout's runtime C and headers into a fresh test directory and compiles every translation unit, collecting diagnostics even when the compiler succeeds. It archives those objects once per invocation, then builds and executes each selected fixture. Clang uses the same fresh-runtime harness. This adds no cache and uses neither the existing runtime archive cache nor oracle observation caches. Setting `ADAMIC_GATE_UNCACHED=1` remains the normal command. `ADAMIC_LANE_REPORT` selects a directory for exact compiler commands, full diagnostics, emitted C, Node/native stdout, stderr, exit codes, and leak-check results.

The checked fixtures deliberately compare with emitted JavaScript carrying Adamic's inserted checks. Other fixtures compare with Node running source through `oracle/node.mjs`. stdout and stderr are compared byte for byte, with exact exit codes. The native sanitizer execution disables leaks for panic compatibility; every Node-finishing sanitized fixture runs again with LeakSanitizer enabled.

## Results and order

After `source /workspace/adamic-tools/env.sh`, every run used `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1`, `go test ./internal/oracle`, `-v -count=1 -timeout 30m`, and direct log redirection.

| Order | Selection | Optimization | Observed result | Log |
|---|---|---|---|---|
| 1 | `^TestGCCAgreesWithNode$/internal/oracle/testdata/read_order.a$` | O2 | pass, exact agreement | `/tmp/gcc-read-release.log` |
| 2 | Same | O1 ASan/UBSan | pass, exact agreement | `/tmp/gcc-read-sanitize.log` |
| 3 | 36 registered lowered fixtures selected by source `\b(throw\|catch)\b` | O2 | all pass | `/tmp/gcc-exceptions-release.log` |
| 4 | Same 36 | O1 ASan/UBSan | all pass | `/tmp/gcc-exceptions-sanitize.log` |
| 5 | `^TestGCCAgreesWithNode$`, all 328 lowered fixtures | O2 | all pass | `/tmp/gcc-full-release.log` |
| 6 | Same 328 | O1 ASan/UBSan | all pass; 278 finishing fixtures leak-clean | `/tmp/gcc-full-sanitize.log` |

The selected exception fixture names are preserved in the evidence archive. The selection was made by scanning `.a` sources for `throw` or `catch`, then intersecting with the registered lowered fixtures through the lane's subtest selector.

**Every fixture whose GCC binary differs from Node: none observed**, in either mode. There is therefore no first differing line or observed evaluation-order, aliasing, char-signedness, or shift cause to report. Agreement on this Linux amd64 machine does not establish portability to another implementation. In particular, the signed-char dependency below is present even though this machine produces the intended bytes.

**GCC fixture sanitizer output: no ASan, UBSan, or LeakSanitizer finding observed.** Every captured native stderr agrees with Node, all sanitized fixture runs exit as Node does, and all 278 finishing leak checks succeed. Compiler warnings are compile-time observations, not sanitizer findings.

## Emitter diagnostics and source provenance

There are two distinct emitter warning mechanisms requested here. GCC prints eleven different value pairs for the overflow mechanism; all exact message variants are also listed in the warning inventory artifact.

| Kind | Exact warning text, excluding path/line prefix | Emitted C construct | Adamic source that produces it |
|---|---|---|---|
| `-Wpedantic` | `warning: initializer element is not constant [-Wpedantic]` | `static adamic_maybe_number adamic_global_5_last = (adamic_maybe_number){false, 0.0};`, emitted `classes.a` main.c:26 | `internal/oracle/testdata/classes.a:48`: `const last = stack.pop();` creates the optional-number global; the C initializer sets its pre-execution undefined value. |
| `-Woverflow` | `warning: overflow in conversion from ‘int’ to ‘char’ changes value from ‘195’ to ‘-61’ [-Woverflow]` | `static const char adamic_bytes_1[4096] = { ... ,195,169};`, diagnosed at main.c:136 | `internal/oracle/testdata/long_literals.a:7`: the `past` string ends in `é`, UTF-8 bytes 195,169. |

The pedantic construct is static aggregate initialization by a compound-literal value instead of a constant initializer. GCC accepts it as an extension once the warning is nonfatal. Its presence is a C11 portability issue; the optional-number runtime value still agrees with Node in this lane.

The long strings are emitted as numerical byte arrays to avoid C's guaranteed string-literal length limit. Converting an integer beyond signed plain char's range is implementation-defined. On this machine GCC chooses the negative char value with the intended underlying byte. The other message variants have the same form with pairs `169 → -87`, `228 → -28`, `184 → -72`, `150 → -106`, `231 → -25`, `149 → -107`, `140 → -116`, `240 → -16`, `159 → -97`, and `141 → -115`. Each appears in the UTF-8 `static const char adamic_bytes_4[22800]` initializer beginning at main.c:140, from `long_literals.a:10`'s `const text = 'héllo 世界 🌍 ...';`. No emitter change or `-funsigned-char` workaround was applied.

## Every warning

[The compressed evidence](gcc-lane-completion-evidence.json.gz) preserves every diagnostic occurrence, including the source excerpt and notes, alongside complete emitted C and comparison observations for one final full run per mode. [The inventory](gcc-lane-completion-warnings.tsv) groups exact warning messages by fixture/runtime source and diagnosed location with occurrence counts. Repeated warnings are counted, not silently dropped.

| GCC warning category | O2 occurrences | O1 sanitized occurrences |
|---|---:|---:|
| `-Woverflow` | 14,402 | 14,402 |
| `-Wpedantic` | 32 | 32 |
| `-Wtype-limits` | 11 | 11 |
| `-Winfinite-recursion` | 1 | 1 |
| `-Wunknown-pragmas` | 3 | 3 |
| `-Wmaybe-uninitialized` | 3 | 1 |
| `-Warray-bounds` | 3 | 0 |
| **Total** | **14,455** | **14,450** |

The runtime warnings are the same mechanisms identified in the prior report: dtoa's inferred negative digit index/memset offset at O2; ignored FP_CONTRACT pragmas in hypot, ieee754 and radix; ieee754's `fq`; and input's `bytes` and `length` at O2. The redundant unsigned-byte lower bounds in regexp fixtures and intentional `stack_forever.a` recursion remain diagnosed. All now build and run. No runtime invariant violation was dynamically demonstrated by the sanitizers.

## Toolchain setup

`bash cloud/setup.sh > /tmp/gcc-continuation-setup.log 2>&1` exited 0. Its timing lines were: Node ready 0.071s, Go ready 0.091s, submodules ready 0.105s, clang ready 0.346s, markdown dependencies ready 0.990s, Go build ready 41.326s, test binaries deferred 41.448s, build cache warm 41.453s, done 41.526s on 5 processors. The environment file is `/workspace/adamic-tools/env.sh`; GCC is the installed Debian 14.2.0-19 compiler.

Setup build-flags line:

```text
commit=36f99309f0cd383c9684c94effffc6a0e85cae8b nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gate-inputs=false load-before=0.06 0.04 0.01 1/141 937 load-after=5.99 1.58 0.53 1/143 1926
```

The requested `git fetch origin devtools/gcc-lane` fetched FETCH_HEAD but did not create a tracking ref because this checkout's fetch refspec exposed only main. `git fetch origin devtools/gcc-lane:refs/remotes/origin/devtools/gcc-lane` then allowed the requested checkout. The branch was continued, without rebasing onto main or rewriting history.

## Mutation and validation

All source mutations were run independently and restored. No emitter or runtime source was mutated by this continuation.

| Mutant | Intended check and observed output | Log |
|---|---|---|
| Stop removing GCC's `-Werror` (change the GCC-only condition to an impossible compiler name) | `TestGCCLaneWarningPolicy` exits 1 and reports `gcc flags: [... -Werror ...]` | `/tmp/gcc-mutant-fatal.log` |
| Remove `-Werror` for both compilers (change the condition to `if true`) | The same test exits 1 and reports `clang flags: [...]` without `-Werror`; the GCC check passes first, so it cannot mask clang's failure | `/tmp/gcc-mutant-clang.log` |
| Append `!` to the dedication's lowered string | With `ADAMIC_GCC_LANE=1`, `TestGCCLaneComparisonCatchesMutants` builds with GCC, observes `stdout differs`, and requires first difference `line 1:`; the test passes only if both assertions catch the changed byte | `/tmp/gcc-final-mutants.log` |
| Read element 4 of a four-element initialized `calloc` allocation | GCC's O1 combined sanitizer build succeeds; execution exits 1 with UBSan `load of address ... with insufficient space for an object of type 'int'` | `/tmp/gcc-sanitizer-mutants/heap.log` |
| Lose a 17-byte `malloc` allocation after a volatile write | GCC's O1 combined sanitizer build succeeds; execution exits 1 with `ERROR: LeakSanitizer: detected memory leaks` and `17 byte(s) leaked in 1 allocation(s)` | `/tmp/gcc-sanitizer-mutants/leak.log` |

The sanitizer probe sources, exact flags, successful compiler output, exit codes, and complete diagnostics are embedded in the evidence archive. These are deliberate standalone mutants, not fixture findings. The first heap probe was also caught by UBSan; an initial expectation of the ASan wording was corrected to the observed UBSan report. Neither probe was stopped by a compiler warning. No cache was added, so there is no new cache-key component mutant to run. The previous worker's runtime-cache mutants remain historical, not claimed as rerun here.

The final source passed:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 go test ./internal/oracle -run 'TestGCCLaneWarningPolicy|TestGCCLaneComparisonCatchesMutants' -v -count=1 -timeout 30m > /tmp/gcc-final-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/gcc-completion-oracle-gate.log 2>&1
go vet ./... > /tmp/gcc-completion-vet.log 2>&1
gofmt -l cmd internal > /tmp/gcc-completion-format.log
git diff --check > /tmp/gcc-completion-diff.log 2>&1
ADAMIC_GATE_UNCACHED=0 ADAMIC_GCC_LANE=1 ADAMIC_LANE_REPORT=/tmp/gcc-gate-switch-unset go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-gate-switch-unset.log 2>&1
```

All tests exited 0. The oracle package log reports `ok github.com/system-inc/adamic/internal/oracle`; vet, formatting, and diff logs are empty. The ordinary package gate includes the complete clang oracle, release and sanitized builds, and the stream/permission and other package tests. The GCC opt-in is off in that ordinary gate and covered by the separate twelve final full-lane timing invocations.

With `ADAMIC_GATE_UNCACHED=0`, the lane still runs fresh. Its 328 fixture observations were loaded independently and compared with the final uncached O2 GCC observations: all fixture identities, stdout bytes, stderr bytes, and exit codes are identical. This verifies the gate switch does not change the lane's answers.

Coverage limits: Linux amd64, GCC 14.2.0 and the registered 328 lowered fixtures. No new fixture, emitter fix, runtime fix, runtime cache, default-compiler change, split-GCC build, other architecture, or full-repository test gate is claimed. `go vet ./...` covered the repository statically.

## Whole-oracle cost

Same box and commit, C/G then G/C then C/G, best of three. Each run builds a fresh runtime archive and all 328 binaries, executes Node and native, writes full evidence, and (sanitized mode) leak-checks 278 finishing fixtures. Go action caching is enabled; compiler/runtime/fixture result caching is absent. No setup, tests, vet, or audits overlapped these final measurements. These are whole lowered-fixture oracle lane costs, including warning collection, not compiler-only benchmarks or the ordinary oracle's extra stream/permission probes.

Instrument command: `source /workspace/adamic-tools/env.sh && python3 docs/gcc-lane-measure.py > /tmp/gcc-lane-measure.log 2>&1`. Python records `time.monotonic()` immediately around each `subprocess.run`. [The timings JSON](gcc-lane-completion-timings.json) includes each exact command, exit, flags, and metadata. Every run exited 0.

| Loop | Before: clang seconds | After: GCC seconds | Instrument and exact commands |
|---|---:|---:|---|
| Whole release oracle lane, round 1 | 26.316294 | 33.221406 | Python monotonic; C: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=clang ADAMIC_LANE_SANITIZE=0 ADAMIC_LANE_REPORT=/tmp/gcc-final-release-1-clang go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-release-1-clang.log 2>&1`; G: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=gcc ADAMIC_LANE_SANITIZE=0 ADAMIC_LANE_REPORT=/tmp/gcc-final-release-1-gcc go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-release-1-gcc.log 2>&1` |
| Whole release oracle lane, round 2 | 25.480852 | 30.533290 | Python monotonic; C: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=clang ADAMIC_LANE_SANITIZE=0 ADAMIC_LANE_REPORT=/tmp/gcc-final-release-2-clang go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-release-2-clang.log 2>&1`; G: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=gcc ADAMIC_LANE_SANITIZE=0 ADAMIC_LANE_REPORT=/tmp/gcc-final-release-2-gcc go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-release-2-gcc.log 2>&1` |
| Whole release oracle lane, round 3 | 25.403104 | 32.503241 | Python monotonic; C: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=clang ADAMIC_LANE_SANITIZE=0 ADAMIC_LANE_REPORT=/tmp/gcc-final-release-3-clang go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-release-3-clang.log 2>&1`; G: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=gcc ADAMIC_LANE_SANITIZE=0 ADAMIC_LANE_REPORT=/tmp/gcc-final-release-3-gcc go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-release-3-gcc.log 2>&1` |
| Whole release oracle lane, best of 3 | 25.403104 | 30.533290 | Same three commands above |

Observed release lane cost: GCC is 20.20% above clang using the two minima. This measures the complete diagnostic harness, including GCC's much larger warning output.

| Whole sanitize oracle lane, round 1 | 42.134895 | 43.076062 | Python monotonic; C: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=clang ADAMIC_LANE_SANITIZE=1 ADAMIC_LANE_REPORT=/tmp/gcc-final-sanitize-1-clang go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-sanitize-1-clang.log 2>&1`; G: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=gcc ADAMIC_LANE_SANITIZE=1 ADAMIC_LANE_REPORT=/tmp/gcc-final-sanitize-1-gcc go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-sanitize-1-gcc.log 2>&1` |
| Whole sanitize oracle lane, round 2 | 41.831693 | 43.591496 | Python monotonic; C: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=clang ADAMIC_LANE_SANITIZE=1 ADAMIC_LANE_REPORT=/tmp/gcc-final-sanitize-2-clang go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-sanitize-2-clang.log 2>&1`; G: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=gcc ADAMIC_LANE_SANITIZE=1 ADAMIC_LANE_REPORT=/tmp/gcc-final-sanitize-2-gcc go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-sanitize-2-gcc.log 2>&1` |
| Whole sanitize oracle lane, round 3 | 42.751412 | 42.832297 | Python monotonic; C: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=clang ADAMIC_LANE_SANITIZE=1 ADAMIC_LANE_REPORT=/tmp/gcc-final-sanitize-3-clang go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-sanitize-3-clang.log 2>&1`; G: `ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=gcc ADAMIC_LANE_SANITIZE=1 ADAMIC_LANE_REPORT=/tmp/gcc-final-sanitize-3-gcc go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -v -count=1 -timeout 30m > /tmp/gcc-final-sanitize-3-gcc.log 2>&1` |
| Whole sanitize oracle lane, best of 3 | 41.831693 | 42.832297 | Same three commands above |

Observed sanitize lane cost: GCC is 2.39% above clang using the two minima. This measures the complete diagnostic harness, including GCC's much larger warning output.

Build-flags line applying to every timing row:

```text
commit=d0c559f64f794100e24c7fb85bc9c952a65909dc nproc=5 cgroup.cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 gcc=gcc (Debian 14.2.0-19) 14.2.0 cached=no (observations, runtime archives, fixture binaries); Go action cache=yes
```

Exact C flags for each compiler and mode:

- release clang: `-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2`
- release gcc: `-std=c11 -Wall -Wextra -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -ffp-contract=off -fno-optimize-sibling-calls -O2`
- sanitize clang: `-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`
- sanitize gcc: `-std=c11 -Wall -Wextra -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`

Load averages and runnable/process counts before and after every measurement:

| Mode | Round | Compiler | Load before | Load after |
|---|---:|---|---|---|
| release | 1 | clang | `1.38 2.54 1.50 1/152 40024` | `3.53 2.95 1.67 1/154 43816` |
| release | 1 | gcc | `3.53 2.95 1.67 1/154 43816` | `4.38 3.20 1.79 1/153 48362` |
| release | 2 | gcc | `4.38 3.20 1.79 1/153 48362` | `5.33 3.51 1.94 1/154 52910` |
| release | 2 | clang | `5.33 3.51 1.94 2/154 52910` | `5.52 3.72 2.06 1/155 56694` |
| release | 3 | clang | `5.52 3.72 2.06 1/155 56694` | `5.79 3.92 2.17 1/155 60478` |
| release | 3 | gcc | `5.79 3.92 2.17 1/155 60478` | `5.95 4.13 2.29 1/154 65024` |
| sanitize | 1 | clang | `5.95 4.13 2.29 2/154 65024` | `6.40 4.45 2.49 1/154 69424` |
| sanitize | 1 | gcc | `6.40 4.45 2.49 1/154 69424` | `4.91 4.31 2.53 1/154 74530` |
| sanitize | 2 | gcc | `4.91 4.31 2.53 2/154 74530` | `5.47 4.53 2.68 1/155 79638` |
| sanitize | 2 | clang | `5.47 4.53 2.68 1/155 79638` | `5.59 4.67 2.81 1/156 83992` |
| sanitize | 3 | clang | `5.59 4.67 2.81 1/156 83992` | `5.02 4.60 2.87 1/155 88344` |
| sanitize | 3 | gcc | `5.02 4.60 2.87 1/155 88344` | `5.10 4.67 2.97 1/155 93459` |
