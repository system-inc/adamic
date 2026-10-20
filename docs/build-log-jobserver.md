# internal/native on cc5e21eb, Linux

Sha `cc5e21ebe102a28302017321fb24947c5b997852` (branch `compiler/build-log-jobserver`, one commit on origin/main
`1133862655ea`), checked out with `git worktree add --detach`, then `git submodule update --init --recursive --depth 1`
(cohere at `06af2e39`). No code was changed.

## The box

- Linux 6.18.44 x86_64, `nproc` 4, run as root, `TMPDIR=/tmp/adamic-gate` (mode 1777).
- clang: the official `LLVM-20.1.8-Linux-X64.tar.xz` from llvm-project releases, unpacked into `/opt/adamic-tools/llvm`
  before `cloud/setup.sh --wasi-sdk` ran, so setup took it as its first sane candidate and linked it into
  `/opt/adamic-tools/bin`, first on PATH. `clang --version`:
  `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`,
  `Target: x86_64-unknown-linux-gnu`, `InstalledDir: /opt/adamic-tools/llvm/bin`.
- Node: `v24.22.0`, the newest v24 in nodejs.org's index, installed by `cloud/setup.sh` at `/opt/adamic-tools/bin/node`.
- wasi-sdk 27 as `cloud/setup.sh --wasi-sdk` installs it (`clang version 20.1.8-wasi-sdk`), with
  `WASI_SYSROOT=/opt/adamic-tools/wasi-sdk/share/wasi-sysroot`.
- Go: `go1.27.1 linux/amd64`, fetched by GOTOOLCHAIN=auto as go.mod asks.
- `/proc/loadavg` when the box facts were recorded, just after setup's Go build: `5.64 2.70 1.08 1/108 3938`.
  Before the full package run: `3.40 2.60 1.14`; after it: `4.08 3.67 2.19`.

## What ran and what it printed

| Command | Result |
| --- | --- |
| `gofmt -l internal/native` | printed nothing |
| `go vet ./internal/native` | printed nothing, exit 0 |
| `go test ./internal/native/runtimeproducts -count=1 -v -timeout 20m` | exit=0, `ok ... 16.569s` |
| `go test ./internal/native -count=1 -timeout 40m` | exit=0, `ok ... 422.326s` (424 s wall) |
| `go test ./internal/native -count=1 -v -run '^(TestSplitBuildLogsEachUnitAndTheLink\|...)$'` | exit=0, four PASS |

runtimeproducts' outcome line for the test the commit message names:

```
--- PASS: TestTheRuntimeBuildReachesOnlyItsDeclaredFiles (16.55s)
```

The targeted run:

```
--- PASS: TestRuntimeArchiveLogsItsCompilesAndItsHit (0.16s)
--- PASS: TestOneUnitBuildLogsItsClang (0.76s)
--- PASS: TestSplitBuildLogsEachUnitAndTheLink (0.48s)
--- PASS: TestBuildLogKeepsANameOnOneLine (0.00s)
PASS
ok  	github.com/system-inc/adamic/internal/native	1.415s
```

No test failed, so nothing was run on origin/main `1133862655ea` for comparison.

## What this does not show

- `full.log` is the package run without `-v`, as asked, so it holds only the `ok` line: it does not say which tests
  skipped, including whether any WASI test skipped. Only the targeted and runtimeproducts runs list tests by name.
- A first start of the full package run was stopped after a few seconds because its background limit was shorter than
  the test timeout; its test binary was killed and its partial log discarded. The run reported here started fresh.
- One run each, on one box. No mutant was run here; the commit's own mutants are on `evidence/compiler/build-log-tests`.

Logs: [runtimeproducts.log](build-log-jobserver/runtimeproducts.log), [full.log](build-log-jobserver/full.log),
[targeted.log](build-log-jobserver/targeted.log).
