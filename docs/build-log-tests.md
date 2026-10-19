# #gydfg86 build log tests and their mutants on Linux

This run changed no code. It ran tests and records what they printed. The raw logs are beside this file, in
`docs/build-log-tests/node22/` (the first round) and `docs/build-log-tests/node24/` (the round that counts).

## The box

- Linux 6.18.44-fc-v139, x86_64, Intel(R) Xeon(R) Processor @ 2.10GHz, 15 GiB of memory.
- `nproc`: 4.
- `/proc/loadavg` before any test: `0.48 0.11 0.04 2/107 266`. Before the Node 24 full run: `1.74 4.52 3.86 2/107 803`;
  after it: `6.22 4.73 4.05 1/109 15834`. Each run's loadavg is in the logs of this session, not in the log files.
- clang: the official `LLVM-20.1.8-Linux-X64.tar.xz` from the llvm-project releases (sha256
  `1ead36b3dfcb774b57be530df42bec70ab2d239fbce9889447c7a29a4ddc1ae6`), extracted to `/opt/llvm-20.1.8`, with its
  `bin` first on PATH. `clang --version`:

  ```
  clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
  Target: x86_64-unknown-linux-gnu
  Thread model: posix
  InstalledDir: /opt/llvm-20.1.8/bin
  ```

  The box also has Ubuntu clang 18.1.3 at `/usr/bin/clang`; `which clang` printed `/opt/llvm-20.1.8/bin/clang` for
  every run.
- Go: `go version go1.27.1 linux/amd64`, as `go.mod` asks (`go 1.27`, `toolchain go1.27.1`).
- Node: the box came with v22.22.0 (`/opt/node22/bin/node`). The first round ran on it. Many internal/native tests
  refuse anything but Node 24 (`Node 24 required`), so for the second round I installed what `cloud/setup.sh` installs:
  the newest v24 from nodejs.org, which was v24.22.0, at `/opt/adamic-tools/node/bin`, second on PATH after clang.
- WASI: the first round had no WASI_SYSROOT. For the second round I installed wasi-sdk 27 as `cloud/setup.sh` and
  `docs/wasm.md` do (`wasi-sdk-27.0-x86_64-linux.tar.gz` from the WebAssembly/wasi-sdk releases) at
  `/opt/adamic-tools/wasi-sdk` and set `WASI_SYSROOT=/opt/adamic-tools/wasi-sdk/share/wasi-sysroot`.
- The shell runs as root, so the input tests drop to uid 65534: `TMPDIR=/tmp/adamic-gate`, mode 1777.
- I did not set ADAMIC_TOOLS or ADAMIC_SANDBOX. Outside a sandbox `internal/toolchain` falls back to PATH, which is how
  clang 20.1.8 and Node were found.
- Each sha ran from its own `git worktree add --detach`, with `git submodule update --init --recursive --depth 1`
  (cohere 0183ccf8, cohere/TypeScript d92d9bfe on every sha; `go.mod` replaces the checker with `./cohere`).

## The shas

| Sha | What | Parent |
| --- | --- | --- |
| `0becdc91ce30bf72065c5534dac36ec182b57b89` | the change (`compiler/build-log-runtime`) | |
| `8a8fbfcef0bcff8464a1cb3402e93b6e354a3705` | mutant 3 (`evidence/compiler/build-log-mutant-3`) | 0becdc91 |
| `2d28cb7329c0c9e678473c89e2317359786bd8e7` | mutant 4 (`evidence/compiler/build-log-mutant-4`) | 0becdc91 |

Mutant 3 changes one line in each of four files:

- `units.go`: a split unit's cache hit is logged as `miss`.
- `library.go`: a runtime compile is logged at `0` seconds.
- `native.go`: the one-unit clang call is logged as `cLink`, not `cBuild`.
- `buildlog.go`: only spaces in a name become `_`, not every whitespace character.

Mutant 4 changes one line in each of two files:

- `units.go`: a split compile is logged at `0` seconds.
- `library.go`: the cache's runtime archive is logged as `miss`, not `hit`.

## The targeted runs

On each sha: `gofmt -l internal/native` printed nothing, `go vet ./internal/native` exited 0, then

```
go test ./internal/native -count=1 -v -timeout 20m \
  -run '^(TestSplitBuildLogsEachUnitAndTheLink|TestBuildLogKeepsANameOnOneLine|TestRuntimeArchiveLogsItsCompilesAndItsHit|TestOneUnitBuildLogsItsClang|TestARuntimeIsTheSameOnAnyNumberOfWorkers|TestUnitsPreserveSharedState|TestUnitSystemHeaderProvenance)$'
```

I ran this twice per sha: on Node 22 without WASI_SYSROOT, then on Node 24 with it. Both rounds gave the same PASS and
FAIL for every test. The lines below are from the Node 24 round (`node24/<sha>.log`). Only the keys, unit names and
seconds in the messages changed between rounds.

### 0becdc91 (the change): exit 0

| Test | Result |
| --- | --- |
| TestSplitBuildLogsEachUnitAndTheLink | PASS |
| TestBuildLogKeepsANameOnOneLine | PASS |
| TestRuntimeArchiveLogsItsCompilesAndItsHit | PASS |
| TestOneUnitBuildLogsItsClang | PASS |
| TestARuntimeIsTheSameOnAnyNumberOfWorkers | PASS |
| TestUnitsPreserveSharedState | PASS |
| TestUnitSystemHeaderProvenance | PASS |

### 8a8fbfce (mutant 3): exit 1

| Test | Result |
| --- | --- |
| TestSplitBuildLogsEachUnitAndTheLink | FAIL |
| TestBuildLogKeepsANameOnOneLine | FAIL |
| TestRuntimeArchiveLogsItsCompilesAndItsHit | FAIL |
| TestOneUnitBuildLogsItsClang | FAIL |
| TestARuntimeIsTheSameOnAnyNumberOfWorkers | PASS |
| TestUnitsPreserveSharedState | PASS |
| TestUnitSystemHeaderProvenance | PASS |

- TestSplitBuildLogsEachUnitAndTheLink:
  `units_build_log_test.go:125: logged-hit: "cCompile functions_0000.c 9369fa21a67d miss 0.018", want a keyed hit`.
  This is the mutant's own: the `units.go` line logged the hit as `miss`.
- TestBuildLogKeepsANameOnOneLine:
  `units_build_log_test.go:148: wrote "cLink a_b\tc\ncCompile_forged_0123456789ab_hit_9.000 - off 1.000\n", want "cLink a_b_c_cCompile_forged_0123456789ab_hit_9.000 - off 1.000\n"`.
  This is the mutant's own: the `buildlog.go` line left the tab and newline in the name.
- TestRuntimeArchiveLogsItsCompilesAndItsHit:
  `runtime_build_log_test.go:92: "cCompile runtime/unit1.c 0123456789ab miss 0.000", want a miss under the archive's key with its seconds`.
  This is the mutant's own: the `library.go` line logged `0.000` seconds.
- TestOneUnitBuildLogsItsClang: `runtime_build_log_test.go:186: logged ["cLink program"], want cBuild program`.
  This is the mutant's own: the `native.go` line logged the kind as `cLink`.

### 2d28cb73 (mutant 4): exit 1

| Test | Result |
| --- | --- |
| TestSplitBuildLogsEachUnitAndTheLink | FAIL |
| TestBuildLogKeepsANameOnOneLine | PASS |
| TestRuntimeArchiveLogsItsCompilesAndItsHit | FAIL |
| TestOneUnitBuildLogsItsClang | PASS |
| TestARuntimeIsTheSameOnAnyNumberOfWorkers | PASS |
| TestUnitsPreserveSharedState | PASS |
| TestUnitSystemHeaderProvenance | PASS |

- TestSplitBuildLogsEachUnitAndTheLink:
  `units_build_log_test.go:123: logged-miss: "cCompile state.c 54fb8631ec96 miss 0.000" says clang took no time`.
  This is the mutant's own: the `units.go` line logged `0.000` seconds.
- TestRuntimeArchiveLogsItsCompilesAndItsHit:
  `runtime_build_log_test.go:124: "cArchive runtime.a e6b0cb09aac8 miss 0.000", want the archive hit under e6b0cb09aac8`.
  This is the mutant's own: the `library.go` line logged the cache's archive as `miss`.

Every expected red was red, every other listed test passed, and every failure line names the outcome, seconds, kind or
name its mutant changed. No mutant died to `-Werror`, gofmt or vet: those were clean on both mutants.

## The full package on 0becdc91

`go test ./internal/native -count=1 -timeout 40m`

- **Node 24.22.0 with WASI_SYSROOT (`node24/full.log`): exit 0**, `ok github.com/system-inc/adamic/internal/native 294.753s`.
  No test failed.
- Node 22.22.0 without WASI_SYSROOT (`node22/full.log`): exit 1 after 721.278s, 59 tests failed. Every one comes from
  the box, not from the change:
  - 49 stop at the oracle's own guard, `Error: Node 24 required` (for example
    `regexp_test.go:860: Node oracle: exit status 1` with `Error: Node 24 required` in the output):
    TestRegExpBytecodeRandomNodeShard00 to TestRegExpBytecodeRandomNodeShard39, TestRegExpV8NativeShard0 to
    TestRegExpV8NativeShard3, TestRegExpSearchNode, TestRegExpLintPatternsNode, TestRegExpRegularEngineNode,
    TestRegExpRuntimeCompilerMatchesNode and TestRegExpRuntimeCompilerMatchesNodeWASM.
  - 2 meet Node 22's older V8, which refuses duplicate named groups: TestRegExpRuntimeParserWASI and
    TestRegExpRuntimeParserAcceptance (`reference versus Node "(?<x>a)|(?<x>b)" "": <nil> / Invalid regular expression: /(?<x>a)|(?<x>b)/: Duplicate capture group name`).
  - 3 need WASI_SYSROOT: TestRegExpRuntimeV8WASI and TestRegExpRuntimeBytecodeWASI (`WASI_SYSROOT required`), and
    TestSplitAttributedBindingsAgreeWithNode/wasm32-wasi (`native: wasm32-wasi requires WASI_SYSROOT`).
  - 4 meet Node 22's Unicode 16 against the runtime's Unicode 17 tables: TestCaseTablesMatchNodesUnicode
    (`case_tables.h is Unicode 17.0.0 and Node is Unicode 16.0: update case_generate.go and run go generate`),
    TestCaseMappingMatchesNodePoints0To2FFFF (102 of 196608 lines differ), TestNormalizeMatchesNodePoints00 (28 of
    65536) and TestNormalizeMatchesNodePoints01 (7 of 65536).
  - 1 meets Node 22's `Math.pow`: TestMathAndToFixedMatchJavaScript
    (`pow 40201604189374bc c03028f5c28f5c29: native "2.3358449403599265e-15", Node "2.335844940359927e-15"`).

  To check that, I ran exactly those 59 tests on `2b22b31ccc8b` (the origin/main the brief named) on the same Node 22 box
  without WASI_SYSROOT (`node22/main-node22.log`): exit 1, and the same 59 failed, no more and no fewer. Since nothing
  fails on the Node 24 box, no test is left to compare against main there.

## What this does not cover

- One machine, one run per configuration. I did not repeat runs to look for flakes. The timing assertions passed on a
  4-core box with loadavg up to about 6.
- origin/main had moved to `23ff023b9ed986834d20404d43e2c363667edf9b` (two commits past 2b22b31) by the time I wrote this.
  I ran the main comparison on 2b22b31 only. This branch is based on 23ff023b.
- I did not run the whole gate (`go test ./...`), only `internal/native`.
