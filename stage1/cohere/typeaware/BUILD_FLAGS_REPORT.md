# Explicit release and sanitizer build comparison

The previous headline binaries were release builds: clang `-O2`, no sanitizer
flags, no `-g`, and no `-DADAMIC_COUNT`. This is established by the recorded
harness build paths and the native driver's `Flags`; the historical literal
clang argv was not captured. This rerun captures the actual commands rather
than presenting a reconstructed historical command as an observation.

Every native-versus-Go comparison from this unit forward names the build flags.
Native release means `-O2`, no sanitizers or allocation counting. Native
correctness means `-O1 -g -fsanitize=address,undefined
-fno-sanitize-recover=all`, no allocation counting. Go uses its default optimized
compiler, without race/ASan or disabling optimization; its embedded build
metadata is saved. The sanitizer elapsed time is a correctness-run cost, not a
release-performance baseline.

Clang is `/workspace/adamic-tools/llvm/bin/clang`, version 20.1.8. All native
commands also use these common flags, exactly as captured:

```text
-std=c11 -Wall -Wextra -Werror -pedantic
-Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
-Wno-unused-parameter -Wno-self-assign -ffp-contract=off
-fno-optimize-sibling-calls -DADAMIC_TSGO
```

The four complete literal clang commands, including every source path, archive,
output and link flag, are saved as shell commands and JSON argv:

- [Coverage release](validation-build-flags/coverage-release.clang)
- [Coverage sanitized](validation-build-flags/coverage-sanitized.clang)
- [Volume release](validation-build-flags/volume-release.clang)
- [Volume sanitized](validation-build-flags/volume-sanitized.clang)

A transparent wrapper records argv and executes the real clang without changing
any argument. Stage 0 removes its generated C scratch directory after building;
the literal commands retain those original paths. Reproduce by running the
saved stage-0 build commands through the saved wrapper, which generates fresh
scratch paths. The normal and sanitized checker archives and Go oracles are
reused from the previous passing gate; their SHA-256 hashes are recorded.
The archive build recipes are:

```sh
source /workspace/adamic-tools/env.sh
# cwd /workspace/adamic; normal archive
# CGO_CFLAGS defaults to -O2 -g for C glue; Go uses default optimization.
go build -buildmode=c-archive -o /workspace/tsgo-speed/validation-pass/checker.a ./bridge/tsgo/archive
# Sanitized archive C glue; Go compiler itself remains normally optimized.
CC=clang CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all' \
  go build -buildmode=c-archive -o /workspace/tsgo-speed/validation-pass/checker-asan.a ./bridge/tsgo/archive
```

The normal Go archive contains debug information, as ordinary Go builds do;
that does not add native sanitizer or allocation-count instrumentation. The
release frontend/runtime C command itself has no `-g`.

Fresh native builds use:

```sh
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/tsgo-build-flags
cd /workspace/adamic
go build -o /workspace/tsgo-build-flags/adamic ./cmd/adamic
# Repeat the following two commands for each suite.
for suite in coverage volume; do
CLANG_CAPTURE=/workspace/tsgo-build-flags/$suite-release.clang \
PATH=/workspace/tsgo-build-flags/wrapper:$PATH \
/workspace/tsgo-build-flags/adamic build stage1/cohere/typeaware/${suite}_suite.ts \
  -o /workspace/tsgo-build-flags/$suite-release \
  --tsgo /workspace/tsgo-speed/validation-pass/checker.a
CLANG_CAPTURE=/workspace/tsgo-build-flags/$suite-sanitized.clang \
PATH=/workspace/tsgo-build-flags/wrapper:$PATH \
/workspace/tsgo-build-flags/adamic build stage1/cohere/typeaware/${suite}_suite.ts \
  -o /workspace/tsgo-build-flags/$suite-sanitized \
  --tsgo /workspace/tsgo-speed/validation-pass/checker-asan.a --sanitize
done
```

No build uses `--count`. The runtime argument `--count` prints the total findings
instead of full diagnostics; it does not enable `ADAMIC_COUNT`. The original
headline used that runtime mode, so it is rerun separately from full-output
correctness checks. All runs use `ADAMIC_TSGO_TIMING=1`, exactly as the previous
headline; CPU sampling and scratch phase instrumentation are off. Build work
finishes before timing begins. Processes run sequentially.

Every run uses the identical 77-file manifest
`/tmp/tsgo-profile/final/compiler.manifest` and config
`/tmp/tsgo-typescript/src/compiler/tsconfig.json`. The config, manifest, each
root source, archives and Go oracle hashes are in
[provenance.json](validation-build-flags/provenance.json).
The production Go build recipes use the existing overlay JSON files:

```sh
cd /workspace/adamic/cohere
go build -overlay /workspace/tsgo-speed/validation-pass/coverage-oracle-overlay.json \
  -o /workspace/tsgo-speed/validation-pass/coverage-oracle /workspace/adamic/cohere/adamic_coverage-oracle.go
go build -overlay /workspace/tsgo-speed/validation-pass/volume-oracle-overlay.json \
  -o /workspace/tsgo-speed/validation-pass/volume-oracle /workspace/adamic/cohere/adamic_volume-oracle.go
```

Results and exact runtime commands follow below. These are single isolated
reproduction runs, not new three-trial medians. No implementation or checker
question changed in this unit. Sanitized binaries serve correctness only.

Original headline mode, seconds:

| Native release (`-O2`, no sanitizers/counting) | Native correctness (`-O1 -g`, ASan/UBSan, no counting) | Go (default optimized) |
| ---: | ---: | ---: |
| 62.209701 | 234.834517 | 7.456767 |

Full-output correctness runs, seconds:

| Compiler runner | Native release (`-O2`, no sanitizers/counting) | Native correctness (`-O1 -g`, ASan/UBSan, no counting) | Go (default optimized) |
| --- | ---: | ---: | ---: |
| coverage | 60.764310 | 237.290099 | 7.578700 |
| volume | 17.803283 | 51.869769 | 4.449356 |

Load/run/adapter measurements for the original headline:

| Build | Load s | After-load run s | Adapter mean µs | Queries |
| --- | ---: | ---: | ---: | ---: |
| -O2; no sanitizers; no -DADAMIC_COUNT | 0.311909 | 61.832226 | 6.445 | 2222043 |
| -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all; no -DADAMIC_COUNT | 0.277418 | 233.721575 | 12.524 | 2222043 |
| Go default optimized; no race, ASan or gcflags | 0.281532 | 7.139782 | - | - |

Adapter time includes checker work and C/UTF conversion, excludes native frame
decoding and rule walks. Go production listeners do not make the same bridge
queries; a Go per-query value would not be comparable.

Exact timed invocations (stdout and stderr redirected to separate saved files):

```sh
ADAMIC_TSGO_TIMING=1 /workspace/tsgo-speed/validation-pass/coverage-oracle /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest
ADAMIC_TSGO_TIMING=1 /workspace/tsgo-build-flags/coverage-release /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest
ADAMIC_TSGO_TIMING=1 /workspace/tsgo-build-flags/coverage-sanitized /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest
ADAMIC_TSGO_TIMING=1 /workspace/tsgo-speed/validation-pass/volume-oracle /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest
ADAMIC_TSGO_TIMING=1 /workspace/tsgo-build-flags/volume-release /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest
ADAMIC_TSGO_TIMING=1 /workspace/tsgo-build-flags/volume-sanitized /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest
ADAMIC_TSGO_TIMING=1 /workspace/tsgo-speed/validation-pass/coverage-oracle /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest --count
ADAMIC_TSGO_TIMING=1 /workspace/tsgo-build-flags/coverage-release /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest --count
ADAMIC_TSGO_TIMING=1 /workspace/tsgo-build-flags/coverage-sanitized /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest --count
```

All processes exit 0. Full findings, fixes and suggestions match production Go
and the previous native streams byte for byte: coverage 16,589 findings,
7,120,228 bytes, SHA-256
`cb4d875ac853b1e27561cbcce3b68eccc4d0818e701c9ab2a4e3d8494c614ab3`;
volume 14,232 findings, 6,717,107 bytes, SHA-256
`6c1cafe2c044717bdb8fc427d4d5585b4d5cf357373fc759916638440ec60e59`.
That is 30,821 unchanged compiler findings and fixes across all 26 rules.
ASan/UBSan/LSan report no errors; stderr contains only requested timing fields.
All three count-only headline outputs are `findings 16589`.

Captured command checks pass for all four builds. Three altered-argv mutants
are rejected: adding `-O1` to release, adding sanitizer flags to release, and
adding `-DADAMIC_COUNT`. These are build-provenance checks, not new language or
checker mutants. No compiler, runtime, ABI or rule implementation changed.
The repository corpus, full Go gate and previous checker mutants were not rerun
in this flags-only unit. Prior correctness coverage remains in the speed report.

Setup succeeds: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s,
build cache warm 20s, total 20s; `nproc` is 5 (four-CPU quota), Linux x86_64,
Go 1.27.1, clang 20.1.8, Node 24.19.0. `git diff --check` has no output.
Raw timing records, exact argv, build metadata, provenance, reproduction script,
flag-check mutants and logs are in [validation-build-flags](validation-build-flags/).
The earlier release result was already an `-O2` result. This rerun confirms that
sanitizer overhead did not cause the reported release gap.
