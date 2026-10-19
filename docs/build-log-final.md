# internal/native on d55df6c3, Linux: the final run

Sha under test: `d55df6c354f2af9647a153cc45c14d094516bdb0` (branch `compiler/build-log`, one commit on origin/main
`23ff023b9ed986834d20404d43e2c363667edf9b`). No code was changed. The logs are beside this file, unedited:
`docs/build-log-final/full.log` and `docs/build-log-final/targeted.log`.

## Result

Every command passed. No test failed, so there was nothing to rerun on origin/main.

| Command | Exit | Wall time |
| --- | --- | --- |
| `gofmt -l internal/native` | 0, no files listed | |
| `go vet ./internal/native` | 0, no findings | |
| `go test ./internal/native -count=1 -timeout 40m` | 0, `ok ... 1127.754s` | 1142 s with the build |
| the four build-log tests, `-v` | 0, all four `--- PASS` | 6 s with the build |

## The box

| Fact | Value |
| --- | --- |
| Kernel | Linux 6.18.44-fc-v139 x86_64 |
| User | uid 0, so `TMPDIR=/tmp/adamic-gate`, mode 1777 |
| `clang --version` | `clang version 20.1.8`, llvm-project 87f0227cb601, Target x86_64-unknown-linux-gnu |
| clang source | official `LLVM-20.1.8-Linux-X64.tar.xz` from the llvm-project releases, its bin first on PATH |
| `node --version` | `v24.22.0` (the newest v24 in nodejs.org's index at install time), next on PATH |
| wasi-sdk | 27 (`wasi-sdk-27.0-x86_64-linux.tar.gz`), `WASI_SYSROOT` set to its `share/wasi-sysroot` |
| Go | go1.27.1 linux/amd64, as go.mod's toolchain line asks |
| `nproc` | 4 |
| Memory | about 15 GB |
| `/proc/loadavg` before setup checks | `1.15 1.11 0.54 1/106 600` |
| `/proc/loadavg` as the full run started (08:19:01Z) | `3.11 3.28 1.64 2/107 4498` |
| `/proc/loadavg` as the full run ended (08:38:03Z) | `4.04 5.51 4.08 1/106 29088` |
| `/proc/loadavg` before the targeted run | `3.15 5.24 4.01 1/107 29119` |

## How the box was set up

`cloud/setup.sh` was not run as a whole. Its clang step takes the first clang whose sanitizers work, and this image has
Ubuntu clang 18.1.3 at `/usr/bin/clang`, whose sanitizers work. Its Node step also checks the default `node` first,
which here is v22. So the exact tools were installed by hand into `/opt/adamic-tools`, following setup.sh's steps:

- LLVM 20.1.8, the same release tarball setup.sh downloads when it finds no working clang. Checked the way setup.sh
  checks a clang: a read past a malloc'd array, built with `-fsanitize=address,undefined`, is stopped with
  `heap-buffer-overflow`.
- Node v24.22.0 from nodejs.org, under `/opt`, not `/root`.
- wasi-sdk 27, the same tarball setup.sh's `--wasi-sdk` downloads. As setup.sh does, the SDK's
  `lib/clang/20/lib/wasm32-*` builtins were linked into the PATH clang's resource directory
  (`/opt/adamic-tools/llvm/lib/clang/20/lib`), and its probe program links with
  `clang --target=wasm32-wasi --sysroot=$WASI_SYSROOT -O2`.
- The environment was `PATH` (LLVM bin, then Node bin), `TMPDIR`, `WASI_SYSROOT`, `GOTOOLCHAIN=auto`, and
  setup.sh's `GOPROXY`. The `GOFLAGS=-buildvcs=false -trimpath` and `GOCACHEPROG` that setup.sh's `env.sh` writes were
  not set.
- The checkout was `git worktree add --detach` at the sha, then `git submodule update --init --recursive --depth 1`,
  which exited 0.

## The targeted tests

All four ran and passed. Here are their log lines, verbatim:

```
runtime_build_log_test.go:134: 4 units and ar logged as misses under the archive's key, the cache's archive as a hit, one archive's bytes
--- PASS: TestRuntimeArchiveLogsItsCompilesAndItsHit (0.26s)
runtime_build_log_test.go:188: logged cBuild program, and the binary is the same bytes without the log
--- PASS: TestOneUnitBuildLogsItsClang (1.75s)
units_build_log_test.go:132: 3 units: off, miss and hit each logged once a unit with one link, and five binaries the same bytes
--- PASS: TestSplitBuildLogsEachUnitAndTheLink (1.03s)
--- PASS: TestBuildLogKeepsANameOnOneLine (0.00s)
```

## What this does not show

The full run was not verbose, so `full.log` is its one `ok` line. It shows no per-test passes or skips. In particular,
it doesn't show whether the package's WASI tests ran or skipped on this box. The only WASI evidence here is that
setup.sh's own WASI probe links with this clang and sysroot.
