# GCC differential lane

The [completed comparison report](gcc-lane-completion.md) supersedes the blocked-build results below. GCC now runs with nonfatal warnings in this diagnostic lane; the text below preserves the previous worker's report.

Built an opt-in `native.Options{Compiler: "gcc"}` and an uncached differential test; clang stays the default.
Implementation commits and final validation are recorded below; base is `15ab80659cf70a3cd0ddd292b7c4bdb6084f2d33` from `origin/area/developer-tools`.
Release and sanitizer lane commands reached all 328 lowered fixtures, but GCC runtime warnings blocked every link; the clang baseline passed.
Compiler-path and compiler-selection mutants failed their intended tests; a real one-byte fixture mutant and a heap-overflow mutant were caught.
GCC binary disagreements and fixture sanitizer findings are unmeasured because the required builds fail; emitter and runtime bugs were left unchanged.

## Use

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -count=1 -timeout 30m > /tmp/gcc-lane-full.log 2>&1
ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_SANITIZE=1 go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -count=1 -timeout 30m > /tmp/gcc-lane-sanitize.log 2>&1
```

`ADAMIC_LANE_CC=clang` runs the same lane with clang for timing. The lane always executes Node directly and never stores observations. Checked fixtures use the emitted JavaScript with the same checks as the ordinary oracle. It compares stdout and stderr byte for byte and exit codes exactly, reporting the first differing line. Sanitized finishing runs additionally use LeakSanitizer. Unlowered fixtures remain covered by the ordinary oracle, rather than this compiler lane.

The runtime archive uses the existing key: compiler path, complete version output, ordered flags, platform, and all embedded runtime source/header bytes. No new cache was added. `ADAMIC_GATE_UNCACHED=1` bypasses oracle observations; the pre-existing runtime archive cache remains enabled, exactly as for the ordinary oracle. The lane evaluates a failed runtime build once, then independently compiles each emitted C translation unit with the same flags and the runtime headers, so a runtime failure cannot hide emitter diagnostics. These object-only checks cannot produce a runnable binary.

Compiler choices are empty/`clang` and `gcc`; any other value fails loudly. GCC with split compilation is refused before building, since the existing split prototype invokes clang. Neither an ambient compiler variable nor this opt-in test changes the default.

## Toolchain and flags

The installed compiler is `gcc (Debian 14.2.0-19) 14.2.0`. Installing/updating it was blocked by this box, not by a test: `apt-get update` exited 100 because uid 1000 cannot write `/var/lib/apt/lists/partial`; `sudo` is absent. A writable apt index directory got HTTP 403 for the image's `snapshot.debian.org` sources. The installed `gcc` meta-package is `4:14.2.0-1`, also the candidate in the available package metadata. A refreshed newest candidate could not be established, and no installation is claimed.

`bash cloud/setup.sh` ran before the explicit base fetch and printed Go ready 0s, clang ready 0s, Node ready 1s, submodules ready 1s, build cache warm 153s, done 153s on 5 processors. The environment file is `/workspace/adamic-tools/env.sh`. That preparation overlapped the first pre-change clang baseline, so that baseline is used as correctness evidence only. Its log is `/tmp/gcc-lane-before.log` and says `ok github.com/system-inc/adamic/internal/oracle`. Setup started from `39638d9e278d38bb5aeae887f46d55a70e47aaad`; the requested branch was explicitly fetched to base `15ab806` before measurement. The checkout's fetch refspec initially exposed only main, although the requested base existed remotely.

Both release builds use:

```text
-std=c11 -Wall -Wextra -Werror -pedantic
-Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter
-ffp-contract=off -fno-optimize-sibling-calls -O2
```

Clang additionally uses `-Wno-self-assign`. GCC has no corresponding warning, so that option is omitted, rather than relying on GCC silently accepting unknown negative warning options. All four other suppressions use the same GCC spelling. Sanitizers replace `-O2` with `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`.

No `-fno-strict-aliasing` was added: neither `native.Flags` nor the runtime build supplies it, and clang's `-O2 -###` driver invocation does not request relaxed aliasing. GCC therefore retains its usual strict-aliasing optimization assumptions.

## Observed GCC build failures

`read_order.a` was run first and its emitted C compiled, but `dtoa.c:153` blocked its runtime archive. The 36 registered exception fixtures were selected by `rg -l '\b(throw|catch)\b' internal/oracle/testdata -g '*.a'`, intersected with the main oracle's lowered fixtures; they ran next, followed by the full release lane and then the full sanitized lane. Full release and sanitizer runs each reported 328 failing subtests, every one blocked by the runtime. Independent emitted-C checks found diagnostics in 27 fixtures, with the same categories at both optimization levels. All 328 lowered fixtures were attempted, which includes every finishing fixture in `TestNativeAgreesWithNode`.

Every runtime `.c` was also compiled independently, so this inventory includes failures after the archive builder's first error:

| Runtime file and line | Warning made fatal by `-Werror` | Builds | Reading, not a dynamic finding |
|---|---|---|---|
| `dtoa.c:153` | `-Warray-bounds`: subscript -1; memset offset [-8589934592, -4] outside `delta_minus` | O2 | Likely missed range invariant. `bignum_init` zeroes every object and digit counts are meant to stay nonnegative. A reachable negative count would be a real out-of-bounds write, but this run did not demonstrate one. |
| `hypot.c:19` | `-Wunknown-pragmas`: ignores `#pragma STDC FP_CONTRACT OFF` | O2, O1 sanitized | Compiler compatibility failure. The required `-ffp-contract=off` flag already requests this behavior. |
| `ieee754.c:40` | Same `-Wunknown-pragmas` | O2, O1 sanitized | Same compatibility failure. |
| `ieee754.c:632` | `-Wmaybe-uninitialized`: `fq[0]` | O2, O1 sanitized | The preceding loop initializes `fq[0]` when `jz >= 0`; a negative `jz` would be a real problem. The numerical algorithm's range invariant is not dynamically established here. |
| `input.c:287` | `-Wmaybe-uninitialized`: `length` | O2 | `read_all` sets both outputs on its explicit success path. Its error path returns `errno`, which POSIX read errors make nonzero. Likely GCC cannot prove that external contract. No observed uninitialized read. |
| `input.c:292` | `-Wmaybe-uninitialized`: `bytes` | O2 | Same control-flow/errno contract as `length`. |
| `radix.c:27` | Same `-Wunknown-pragmas` | O2, O1 sanitized | Same compatibility failure. |

All other runtime translation units compiled under their respective audit flags. The full audit logs are `/tmp/gcc-lane-runtime-audit.log` and `/tmp/gcc-lane-runtime-sanitize-audit.log`.

| Fixture | Emitted `main.c` lines | GCC warning |
|---|---|---|
| `047cb0d_n_conditional.a` | 10 | `-Wpedantic` |
| `047cb0d_n_numparam.a` | 10 | `-Wpedantic` |
| `047cb0d_n_paren.a` | 10 | `-Wpedantic` |
| `047cb0d_n_typeof.a` | 13 | `-Wpedantic` |
| `classes.a` | 26,28,30 | `-Wpedantic` |
| `collections.a` | 120,124 | `-Wpedantic` |
| `find_shrinks.a` | 14 | `-Wpedantic` |
| `gaps.a` | 34 | `-Wpedantic` |
| `generic_values.a` | 19 | `-Wpedantic` |
| `indexing.a` | 19 | `-Wpedantic` |
| `json_stringify_scalars.a` | 18 | `-Wpedantic` |
| `literal_optional_shapes.a` | 17 | `-Wpedantic` |
| `long_literals.a` | 136,140-852 | `-Woverflow` |
| `maps_and_text.a` | 32 | `-Wpedantic` |
| `maybe_booleans.a` | 35,43,47,51 | `-Wpedantic` |
| `maybe_collections.a` | 33,35,39 | `-Wpedantic` |
| `maybe_numbers.a` | 13,17 | `-Wpedantic` |
| `narrowed_compared.a` | 14,16 | `-Wpedantic` |
| `narrowed_numbers.a` | 8 | `-Wpedantic` |
| `optional_numbers.a` | 14,16 | `-Wpedantic` |
| `regexp.a` | 3085,3097 | `-Wtype-limits` |
| `regexp_search.a` | 171 | `-Wtype-limits` |
| `regexp_unicode.a` | 51,63,128,140 | `-Wtype-limits` |
| `stack_forever.a` | 12 | `-Winfinite-recursion` |
| `sweeps/regexp_methods.a` | 178,190,3168,3180 | `-Wtype-limits` |
| `tuple_values.a` | 81 | `-Wpedantic` |
| `user_iterators.a` | 113 | `-Wpedantic` |

The table gives exact inclusive line ranges in `native.C` output. [gcc-lane-diagnostics.tsv](gcc-lane-diagnostics.tsv) preserves all distinct diagnostic messages and every diagnosed line, grouped without duplicate columns or repetitions. The complete unabridged compiler output remains in `/tmp/gcc-lane-full.log` and `/tmp/gcc-lane-sanitize.log`.

Readings of the emitted diagnostics:

- `-Wpedantic`, "initializer element is not constant": static struct globals are initialized with compound-literal expressions, for example `static adamic_maybe_number ... = (adamic_maybe_number){false, 0.0};` in `classes.a:26`. This is a real C11 portability problem in static initialization, accepted by clang's extension. It is not evidence of an evaluation-order mismatch.
- `-Woverflow`: `long_literals.a` emits `static const char ... = {...,195,169,...}` for UTF-8 bytes. Converting these positive integers to signed plain `char` is implementation-defined when they do not fit. GCC reports, for example, 195 to -61 and 169 to -87. That is a real portability dependency even though the underlying bytes on this machine are the intended bytes. No output comparison was possible.
- `-Wtype-limits`: regex ASCII matchers test an `unsigned char` with `input[at] >= 0`. The test is always true; it is redundant, with no wrong matcher behavior established by this diagnostic.
- `-Winfinite-recursion`: `stack_forever.a` intentionally recurses until the runtime's stack guard panics. The recursion is the fixture's purpose, not an unexpected runtime bug. GCC still refuses it under the required warning policy.

## Binary and sanitizer results

There are no runnable GCC fixture binaries under the requested strict flags. Consequently the list of observed GCC-versus-Node byte mismatches is unavailable, not an empty list establishing agreement. No first differing output line or evaluation-order/aliasing/shift diagnosis can responsibly be assigned to these blocked fixtures.

GCC's sanitizers work on this box: a standalone initialized heap-overflow probe built at O1 with ASan and UBSan and exited 1 with UBSan's "load ... with insufficient space for an object of type 'int'". Clang built the same probe and exited 1 with ASan's heap-buffer-overflow. Both detect the deliberate fault. The initial uninitialized version of that probe was stopped by GCC's `-Wuninitialized`; it was replaced with `calloc`, so the reported sanitizer proof is not masked by `-Werror`.

The full GCC O1 fixture build is blocked first by `hypot.c:19`; the independent runtime audit also finds the `ieee754.c` and `radix.c` failures above. There are therefore no GCC fixture sanitizer findings to compare with clang. No warning was suppressed and no emitter/runtime source was changed to get past these failures.

## Mutation and validation

| Mutant run independently and restored | Check and observed result |
|---|---|
| Remove `part(compiler)` from `runtimeKey` | `TestRuntimeKeyIncludesEveryInput` exits 1: "compiler path missing from key". Log `/tmp/gcc-lane-mutant-key.log`. |
| Make the GCC compiler option select clang | `TestRuntimeCompilerChoice` exits 1: "gcc runtime: got exit status 1, want exit 2". Two wrappers expose identical version strings but produce different runtime code. Log `/tmp/gcc-lane-mutant-choice.log`. |
| Append `!` to the dedication fixture's lowered string | `TestGCCLaneComparisonCatchesMutants` catches "stdout differs" and line 1, using the real Node/C runners. Clang keeps this proof runnable despite GCC's runtime build failure. Log `/tmp/gcc-lane-mutant-byte.log`. |
| Read element 4 from a four-element initialized allocation | Both standalone sanitizer runtimes stop the program, as above. Log `/tmp/gcc-lane-sanitizer-probe.log`. |

No new cache was introduced. The compiler-identity mutation proves the existing runtime key component required by this unit can fail its check. The compiler-choice mutation exercises actual compiled output rather than only inspecting flags.

The final source passed:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/oracle -count=1 -timeout 30m > /tmp/gcc-lane-touched-gate.log 2>&1
gofmt -l cmd internal > /tmp/gcc-lane-format.log
go vet ./... > /tmp/gcc-lane-vet.log 2>&1
git diff --check > /tmp/gcc-lane-diff-check.log 2>&1
```

The package log reports `ok` for both packages; the other three logs are empty. This covers all `TestNativeAgreesWithNode` fixtures with clang, both release and sanitizers, plus all other tests in the touched packages. The opt-in GCC test skips in the default gate. The full repository test gate was not run; `go vet ./...` was run.

Implementation is in `ae22089e91ed7b0babcea5d2f2606ee108458760`; `3bcf9a8` makes the checked-fixture backend runner unconditionally uncached too. Measurements below use the final source at `3bcf9a8`, with no concurrent setup or audit jobs. The branch is `devtools/gcc-lane`, based on the requested developer-tools ref. No PR was opened.

## Cost


Same machine and final source, interleaved C/G, G/C, C/G, best of three. Instrument: Python `time.monotonic()` immediately before and after `subprocess.run`, including the Go test command's complete wall time. All test output went directly to individual log files.

Exact commands:

```sh
# C
ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=clang go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -count=1 -timeout 30m > /tmp/gcc-lane-final-paired-N-clang.log 2>&1
# G
ADAMIC_GATE_UNCACHED=1 ADAMIC_GCC_LANE=1 ADAMIC_LANE_CC=gcc go test ./internal/oracle -run '^TestGCCAgreesWithNode$' -count=1 -timeout 30m > /tmp/gcc-lane-final-paired-N-gcc.log 2>&1
```

Here `N` is the recorded round number (1, 2 or 3), not a shell variable. [gcc-lane-timings.json](gcc-lane-timings.json) records each actual command and log name, wall time, exit, complete flags and build-flags metadata.

| Loop | Before: clang wall seconds, exit 0 | After: GCC wall seconds, exit 1 | Instrument and exact command |
|---|---:|---:|---|
| Whole release fixture lane, round 1 | 19.845790 | 20.476957 | Python monotonic; C and G above, N=1 |
| Whole release fixture lane, round 2 | 19.938474 | 20.747523 | Python monotonic; C and G above, N=2 |
| Whole release fixture lane, round 3 | 19.749125 | 21.107913 | Python monotonic; C and G above, N=3 |
| Best of 3 | 19.749125 | 20.476957 | Same commands and instrument |

Build-flags line applying to all rows:

```text
commit=3bcf9a84b40392f673225dc640cde711fcb6d66e nproc=5 cgroup.cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 gcc=gcc (Debian 14.2.0-19) 14.2.0 uncached-observations=yes existing-runtime-archive-cache=yes
C flags and the clang-only suppression are recorded in the flags section and per run in the JSON artifact.
```

Each row's load samples (one-, five-, fifteen-minute averages, followed by runnable/process counts and last pid):

| Round | Compiler | Load before | Load after |
|---|---|---|---|
| 1 | clang | `1.71 4.01 3.00 1/158 58979` | `2.83 4.12 3.06 1/162 63042` |
| 1 | gcc | `2.83 4.12 3.06 2/162 63042` | `3.53 4.18 3.10 1/163 66471` |
| 2 | gcc | `3.53 4.18 3.10 1/163 66471` | `4.08 4.25 3.15 1/165 69942` |
| 2 | clang | `4.08 4.25 3.15 2/165 69942` | `4.17 4.25 3.17 1/167 74009` |
| 3 | clang | `4.17 4.25 3.17 1/167 74009` | `5.06 4.44 3.26 1/167 78076` |
| 3 | gcc | `5.06 4.44 3.26 2/167 78076` | `5.30 4.52 3.31 1/167 81495` |

These are complete fixture attempts, not equivalent successful compiler workloads: clang links and runs all binaries, while GCC checks emitted C objects and fails linking on runtime diagnostics. GCC also repeats failed runtime compilation because no failed archive is cached. The numbers cannot establish relative successful-lane cost or a compiler speedup. A GCC-versus-clang cost for the original whole `TestNativeAgreesWithNode` including runnable GCC sanitizers is unavailable under the required flags. The ordinary clang oracle passed both the pre-change baseline and the final touched-package gate.
