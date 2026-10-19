# internal/native on 3d0a6486, Linux

The tests of `internal/native`, run whole on one sha, on one Linux box, on October 19, 2026. No code was changed.

- Sha: `3d0a648678c45ad247458e14c4028abc4d56b199` (branch `compiler/build-log-declared`, one commit on `ce508865b5f7`).
- Checkout: `git worktree add --detach`, then `git submodule update --init --recursive --depth 1` (cohere `0183ccf8`, cohere/TypeScript `d92d9bfe`).
- The logs beside this file are the raw output: `runtimeproducts.log`, `full.log`, `targeted.log`.

## The box

| | |
|---|---|
| `clang --version` | `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`, from the official `LLVM-20.1.8-Linux-X64.tar.xz`, its `bin` first on PATH |
| `node --version` | `v24.22.0` (newest v24 in nodejs.org's index), next on PATH |
| wasi-sdk | 27 (`wasi-sdk-27.0-x86_64-linux.tar.gz`), its wasm32 builtins linked into clang 20's resource directory as `cloud/setup.sh` does; `WASI_SYSROOT=/opt/adamic-tools/wasi-sdk/share/wasi-sysroot`. A malloc/free program linked for `wasm32-wasi` with the PATH clang. |
| Go | `go1.27.1 linux/amd64` (fetched by `GOTOOLCHAIN=auto` from go.mod) |
| `nproc` | 4 |
| `/proc/loadavg` | `1.51 0.84 0.35 1/110 522` at setup; `3.53 3.36 1.63 2/107 9184` before the full run; `4.43 4.09 2.73 1/119 31275` after it |
| user | root, so `TMPDIR=/tmp/adamic-gate`, mode 1777; Node lives at `/opt/adamic-tools/bin/node`, not under /root |

Before the runs, ASan and UBSan under this clang stopped a heap read past an array with `heap-buffer-overflow`.

## What ran and what it printed

| Command | Exit | Time |
|---|---|---|
| `gofmt -l internal/native` | 0, no files listed | |
| `go vet ./internal/native` | 0, nothing printed beyond Go's module downloads | |
| `go test ./internal/native/runtimeproducts -count=1 -v -timeout 20m` | 0 | `ok ... 137.055s` |
| `go test ./internal/native -count=1 -timeout 40m` | 0 | `ok  	github.com/system-inc/adamic/internal/native	533.576s` (540s wall) |
| `go test ./internal/native -count=1 -v -run '^(TestSplitBuildLogsEachUnitAndTheLink\|TestBuildLogKeepsANameOnOneLine\|TestRuntimeArchiveLogsItsCompilesAndItsHit\|TestOneUnitBuildLogsItsClang)$'` | 0 | `ok ... 1.919s` |

TestTheRuntimeBuildReachesOnlyItsDeclaredFiles, verbatim:

```
--- PASS: TestTheRuntimeBuildReachesOnlyItsDeclaredFiles (137.03s)
```

Its census line was `census_test.go:736: build go_list_-export__./internal/native_listing.json 9ae9912b1c60 miss 134.48`; the other two census tests hit that same listing.

The four targeted tests all passed:

```
--- PASS: TestRuntimeArchiveLogsItsCompilesAndItsHit (0.17s)
--- PASS: TestOneUnitBuildLogsItsClang (1.09s)
--- PASS: TestSplitBuildLogsEachUnitAndTheLink (0.65s)
--- PASS: TestBuildLogKeepsANameOnOneLine (0.00s)
```

## Failures and main

No test failed, so nothing was run on `ce508865b5f7`.

## What this does not show

The full run was not verbose, so `full.log` is its one `ok` line. It doesn't say which tests skipped, and in particular it doesn't show whether the WASI tests ran or skipped on this box. This was one run on one box, with no mutants. It shows that the tests pass on this sha, not that they can fail.
