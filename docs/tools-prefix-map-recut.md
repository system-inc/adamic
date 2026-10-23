# tools-prefix-map recut: internal/native, runtimeproducts and lower's agreement tests on Linux

Sha under test: `c8ec81f0670e4daf92d98b9be8cd9ae66ddf66e6` (branch `compiler/tools-prefix-map-2`, one commit on origin/main `e50e502c21`). No code was changed.

Verdict: every command exited 0. Nothing failed, so nothing was rerun on origin/main.

## The box

- Linux 6.18.44, x86_64, 4 cores, 15 GB memory, run as root.
- clang 20.1.8 from the official `LLVM-20.1.8-Linux-X64.tar.xz`, extracted with `--exclude` for `bin/flang*`, `bin/mlir-*`, `bin/lldb*`, `bin/llvm-bolt*`, `bin/clangd` and `lib/lib*.a`, its `bin` first on PATH. `clang --version`: `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`.
- Node v24.22.1 (the newest v24 in nodejs.org's index that day), next on PATH, under `/opt`, not `/root`.
- wasi-sdk 27 (`wasi-sdk-27.0-x86_64-linux.tar.gz`), its wasm32 builtins linked into clang 20.1.8's resource directory as `cloud/setup.sh` does, `WASI_SYSROOT` set to its `share/wasi-sysroot`. The setup's own probe (malloc and free, `--target=wasm32-wasi`) linked with the PATH clang. An ASan and UBSan probe also built and ran clean.
- go1.27.1, as go.mod's toolchain line asks. `GOFLAGS="-buildvcs=false -trimpath"` and `GOTOOLCHAIN=auto`, as `cloud/write-env.sh` sets them.
- `TMPDIR=/tmp/adamic-gate`, mode 1777.
- `git worktree add --detach` at the sha, then `git submodule update --init --recursive --depth 1` (cohere `06af2e3`, cohere/TypeScript `d92d9bfe`).
- `df` showed 17 GB free on `/` as the full run started (14 GB as it finished).
- The oracle cache was cold: this box had never run the suite. `ADAMIC_GATE_UNCACHED` was not set, as the commands asked.

## What ran and what it printed

Each log is beside this file in `docs/tools-prefix-map-recut/`.

| Command | exit | Log |
| --- | --- | --- |
| `gofmt -l internal/native` | 0, no files listed | `gofmt-vet.log` |
| `go vet ./internal/native` | 0, no findings | `gofmt-vet.log` |
| `go test ./internal/native/runtimeproducts -count=1 -v` | 0 | `runtimeproducts.log` |
| `go test ./internal/lower -count=1 -run 'Agreement\|Lean'` | 0 | `lower.log` |
| `go test ./internal/native -count=1 -timeout 40m` | 0 | `full.log` |

### runtimeproducts

The two outcome lines asked for, verbatim:

```
--- PASS: TestTheLeanBuildReachesOnlyItsDeclaredFiles (19.56s)
--- PASS: TestTheRuntimeBuildReachesOnlyItsDeclaredFiles (19.60s)
```

The other outcome lines: `TestAnArchivesBuildIsJournaledAsRuntimeproducts`, `TestTheAgreementCBuildReachesOnlyItsDeclaredFiles`, `TestProduct_LeanCensusExports`, `TestProduct_AgreementCensusExports`, `TestTheLeanCensusFollowsTheBuildAndFailsClosed` and `TestTheJobserverNeverReachesAnArchivesBytes` all PASS; package `ok ... 23.320s`.

No census finding text: the only census output is its build timing lines (`census_test.go:98`, `:111`, `:814`, `:835`, `:854`, `:902`), each a `go list -export` listing for `./internal/native` or `./internal/native/agreementproducts`, first a miss and then hits.

### lower

`ok  	github.com/system-inc/adamic/internal/lower	24.243s`. The run was not `-v`, so the log names no single test; it does not say `[no tests to run]`, so the pattern matched at least one test.

### internal/native

`ok  	github.com/system-inc/adamic/internal/native	760.230s`. The run was not `-v`, so the log does not show which tests skipped (for example any WASI test that found no toolchain); I did not check that separately.

## What was not done

- Nothing failed, so no test was rerun on origin/main `e50e502c21`.
- This is not the whole gate: no other package, no `TestProduct_*` waves beyond what these packages ran, and the oracle cache was not bypassed.
