# Command usage and WASI sanitizer guards

This test-only unit guards audit survivors M02 and M07 for task #gqca183. It adds three top-level tests within step 79's test-unit budget and replaces the two status-only sanitizer rows in `TestBuildTargetParsing`. Base: `b3f83786a0ea00c23d47d774d3b6c98f9dd71336` from current origin/main.

## Behavior held by the tests

`TestUsageExitStatus` runs the existing isolated CLI test driver with no command arguments and pins process exit 2, empty stdout, and the full usage diagnostic including its newline.

`TestBuildWASISanitizeTargetBeforeSource` and `TestBuildWASISanitizeTargetAfterSource` cover both original argument orders. Each first runs `types testdata/build_target.a` and requires successful loading. Each then requires exit 1, empty stdout, and exactly:

```text
adamic: native: sanitizers are not supported for wasm32-wasi
```

The fixture is a valid two-line `.a` program with a string declaration and console output. The tests use child processes so they can run in parallel without replacing the parent's stdout or stderr. Every child command has a 20-second execution limit. No build or toolchain setup is added inside a test.

## Measured green runs

Every row ran alone in a fresh `go test` process with `-count=1`, after toolchain setup. The Go compilation cache was retained; test result caching was bypassed. Measurements include each test's subprocess setup. The process column includes the Go driver and test binary preparation as well.

| Test | Leaf seconds | Whole command seconds | Result |
| --- | ---: | ---: | --- |
| `TestUsageExitStatus` | 0.01 | 2.32 | pass |
| `TestBuildWASISanitizeTargetBeforeSource` | 0.05 | 2.12 | pass |
| `TestBuildWASISanitizeTargetAfterSource` | 0.05 | 2.17 | pass |
| `TestBuildTargetParsing` | 0.00 | 2.07 | pass |
| `TestWASIRejectsTSGoArchive` | 0.00 | 2.17 | pass |

The exact invocation for each row was:

```sh
source /workspace/adamic-tools/env.sh
GOMAXPROCS=4 go test ./cmd/adamic -run '^<test name>$' -count=1 -timeout 90s -json
```

The runner applied an additional 180-second outer timeout to each command. Full outputs are the named `.jsonl` files beside this report; commands and measurements are in `timings.json`. The initial combined baseline also passed (`baseline.jsonl`, package 0.094 seconds).

## Mutant evidence

The exact audit diffs were fetched from `test-audit/cmd-adamic` and replayed with `git apply`, one at a time. The original production file bytes were restored in a `finally` block after each run; no production edits are committed.

| Mutation | Tests catching it | Observed failure |
| --- | --- | --- |
| M02: final usage `return 2` becomes `return 1` | `TestUsageExitStatus` | exit 1 with unchanged usage text; want exit 2 |
| M07: `options.Sanitize = true` becomes `false`, the exact audit diff | both sanitizer argument-order tests | without WASI_SYSROOT: exit 1 with the wrong diagnostic; exact text assertion fails |
| M07 with the installed WASI SDK | both sanitizer argument-order tests | unsanitized build succeeds, exit 0 and empty diagnostic; want exit 1 and sanitizer refusal |
| Fixture initializer changed from a string to a number | both sanitizer argument-order tests | the load prerequisite fails with TS2322; a load failure cannot substitute for the sanitizer refusal |

M02 failed in 0.01 leaf seconds. M07 with the SDK failed in 6.84 and 6.73 leaf seconds, below the unit budget. The fixture load control failed in 0.05 and 0.04 seconds. All mutant commands exited 1 through the test assertions. The restored tests then passed individually as recorded above.

Mutant commands were:

```sh
go test ./cmd/adamic -run '^TestUsageExitStatus$' -count=1 -timeout 90s -json
go test ./cmd/adamic -run '^TestBuildWASISanitizeTarget(Before|After)Source$' -count=1 -timeout 90s -json
WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot \
  go test ./cmd/adamic -run '^TestBuildWASISanitizeTarget(Before|After)Source$' -count=1 -timeout 90s -json
```

The fixture load control used the second command. `M02.diff`, `M07.diff`, the corresponding `.jsonl` logs, and `mutant-results.json` preserve the exact mutations and results.

## Toolchain and scope

`GOPROXY='https://proxy.golang.org|direct' timeout 180 bash cloud/setup.sh` succeeded. Its timing lines were: Go ready 0.023 seconds; Node ready 0.025; submodules ready 0.069; markdown dependencies ready 0.083; clang ready 0.200; Go build ready 119.362; test binaries deferred 119.630; build cache warm 119.632; done 119.659. The printed environment `/workspace/adamic-tools/env.sh` was sourced for every Go command. `nproc` was 5 and cgroup `cpu.max` was `400000 100000`, a four-CPU quota. Go 1.27.1, clang 20.1.8, Node 24.19.0.

The full package, full gate and differential oracle were not run. The new fixture checks command loading and an option refusal; it adds no differential-oracle count row. Counts regeneration and integration lane checks are recorded separately. Only `_test.go`, testdata and review evidence are committed.

Counts command: `timeout 180 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 90s -args -update-counts`. It passed in 66.676 seconds; `internal/oracle/counts.md` had no diff. `git diff --check` passed.
