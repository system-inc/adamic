# Bridge archive build phases

The cold sanitized archive took 109.74 seconds. TypeScript Go compilation used 281.24 CPU seconds; sanitized clang compilation used 1.53 CPU seconds. Retaining the ordinary Go packages while invalidating every C-bearing package reduced elapsed time to 7.19 seconds. The expensive part is Go compilation, rather than sanitizer compilation or the archive link.

## Inputs and machine

- Repository: `20d538c0b68f4a4259e3f46dd9fdf7c2d340264b` (current main when the branch started).
- Cohere submodule: `7945d102a6c18dd36adf9114a758ce646e8b2359`.
- Linux amd64, Go 1.27.1, clang 20.1.8; tool environment `/workspace/adamic-tools/env.sh`.
- `nproc` printed 5; `cpu.max` was `400000 100000`, a four-CPU quota. All measured Go commands used `GOMAXPROCS=4`.
- The selected environment became ready, but its workspace and host were reused. These are empty compiler-cache measurements on the reference CPU quota, not a claim of a newly booted machine or an empty filesystem page cache. Module downloads were already present.

## Measured phases

Each cell is **wall / CPU seconds**. Wall is the union of that phase's tool execution intervals, excluding gaps and counting concurrent executions once. CPU is user plus system time summed across the tool processes. The total row is independently measured end-to-end elapsed and process-tree CPU.

cgo includes its clang children, and the final link includes its external archiver or compiler children. Phase rows overlap and must not be added. They identify where work happens, rather than forming an exclusive accounting partition. Tool wrappers and the Go driver are included in total CPU; compiler identity calls are excluded from their compile phase.

| Phase | Archive, empty Go cache | Archive, warm Go / cold C packages | Archive, repeated warm cache | Test binary, empty Go cache |
| --- | ---: | ---: | ---: | ---: |
| Go TypeScript packages | 92.57 / 281.24 | 0.00 / 0.00 | 0.00 / 0.00 | 101.00 / 267.50 |
| Go other packages | 14.76 / 36.41 | 3.68 / 8.93 | 0.00 / 0.00 | 65.21 / 188.20 |
| cgo generation (inclusive) | 0.90 / 1.41 | 0.59 / 1.21 | 0.32 / 0.88 | 0.70 / 0.85 |
| Sanitized clang compilation | 1.19 / 1.53 | 0.69 / 1.29 | 0.47 / 0.97 | 0.00 / 0.00 |
| Clang preprocessing / probes | 0.31 / 0.27 | 0.16 / 0.20 | 0.06 / 0.10 | 0.00 / 0.00 |
| Final Go link (inclusive) | 1.39 / 1.61 | 1.36 / 1.57 | 1.31 / 1.52 | 1.95 / 2.19 |
| Native ar (inside link) | 0.05 / 0.05 | 0.04 / 0.04 | 0.06 / 0.06 | 0.00 / 0.00 |
| Test binary GCC (inclusive) | 0.00 / 0.00 | 0.00 / 0.00 | 0.00 / 0.00 | 1.06 / 1.02 |
| go vet | 0.00 / 0.00 | 0.00 / 0.00 | 0.00 / 0.00 | 9.10 / 10.03 |
| Other tools / probes | 0.41 / 0.31 | 0.04 / 0.04 | 0.03 / 0.03 | 0.52 / 0.35 |
| **Total** | **109.74 / 335.35** | **7.19 / 14.43** | **2.56 / 4.39** | **146.97 / 493.48** |

Every command exited 0. `go test -c` built the binary without running a test; it also ran the default vet checks, included in its measurement. Its C compiler was the ordinary environment compiler, `/usr/bin/gcc`, without the archive sanitizer flags.

## What each cache removed

- Retaining non-C Go packages saved **102.56 elapsed seconds** and **320.92 CPU seconds** compared with the empty cache. TypeScript Go compile calls went from 60 to zero.
- Retaining the C-bearing package archives as well saved a further **4.62 elapsed seconds** and **10.04 CPU seconds**. This is a package-cache effect, not a separate clang-cache effect: Go stores generated Go and C objects together.
- There was no ccache or other persistent clang cache. The repeated warm archive still invoked cgo/clang for `runtime/cgo` and `net` and performed the final link. Warm Go cache does not mean zero C work.
- These sequential single runs establish the observed savings; differences also include changed parallel scheduling and process overhead. They are not isolated benchmark confidence intervals.

## Recipe and cold-cache procedure

The archive command and sanitizer flags are the recipe in `bridge/tsgo/products_test.go`. Instrumentation added only `-x`, timing wrappers and `-toolexec`; wrappers forwarded the real tools and their arguments unchanged.

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
export GOMAXPROCS=4
export GOCACHE=$(mktemp -d)
export CC=/workspace/bridge-archive-phases-scratch/cc
export CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all'
go build -x -toolexec=/workspace/bridge-archive-phases-scratch/tool.py \
  -buildmode=c-archive -o /workspace/bridge-archive-phases-scratch/cold.a \
  ./bridge/tsgo/archive
```

The runner timestamped every stderr line from `-x` with monotonic seconds since launch. Each tool wrapper measured its start/end with `time.monotonic()` and child user/system CPU with `resource.getrusage(resource.RUSAGE_CHILDREN)`, around the actual tool invocation. The parent runner measured the whole Go command by the same method. All build output went to files; no build was piped through a truncating reader.

The warm repeat reused the first cache and wrote `warm.a`. Before the warm-Go/cold-C run, `go list -deps -export -json -buildmode=c-archive` with the same environment and `-toolexec` located every dependency with `CgoFiles`. Only its exported cache data file was removed:

| C-bearing package | Removed cache archive bytes |
| --- | ---: |
| `runtime/cgo` | 687,068 |
| `net` | 4,112,730 |
| `github.com/system-inc/adamic/bridge/tsgo/archive` | 41,590,004 |

The measurement script initially asserted that there were two C-bearing packages; it found `net` as the third after removing all three files. It stopped before the timed cold-C command. The resumed runner retained the original Go cache and ran that command without rebuilding anything in between. The only Go compiler calls in this run were `runtime/cgo`, `net`, and the archive's `main`. All three C-bearing products were rebuilt; all TypeScript Go packages stayed cached. Because the cache combines Go and C objects, this run also recompiles their small Go portion.

The test binary used a second newly created empty Go cache, unset `CGO_CFLAGS`, and a wrapper forwarding `/usr/bin/gcc`:

```sh
GOCACHE=$(mktemp -d) go test -c -x \
  -toolexec=/workspace/bridge-archive-phases-scratch/tool.py \
  -o /workspace/bridge-archive-phases-scratch/bridge.test ./bridge/tsgo
```

## Slowest TypeScript Go compiler invocations in the cold archive

| Package under `github.com/microsoft/TypeScript/tsc/` | Wall seconds | CPU seconds |
| --- | ---: | ---: |
| `internal/checker` | 26.29 | 34.80 |
| `internal/ast` | 17.52 | 39.60 |
| `internal/parser` | 7.73 | 7.93 |
| `internal/module` | 7.22 | 7.58 |
| `internal/binder` | 6.88 | 7.36 |

No product code or tests changed. No mutants, oracle fixtures, product tests or full gate were run for this measurement-only unit. Raw timestamped logs and per-tool JSON remained in `/workspace/bridge-archive-phases-scratch/`; this branch commits only this report.
