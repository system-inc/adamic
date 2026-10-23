# Where two slow internal/native tests spend their time

Measured on Linux, cold, at `e50e502c21f12f9df64f7e716e980ed84695aa6f` (origin/main on 2026-10-23). Measurement only:
nothing here changes the tests. The instrumented copy of the tests lived in a throwaway worktree and was never
committed; its exact diff is `native-slow-tests/instrumentation.diff`, and every raw `-v` log is beside it in
`native-slow-tests/`.

## Answer

On this box neither test comes near 60 s. Each takes 12 to 14 s, alone or with the other at `-parallel 6`, and
nearly all of it is the test's own clang work. Neither waits on a build another test makes.

- **TestTheJobserverNeverReachesTheBytes** is three child processes, each building the release runtime archive
  from nothing (64 `.o`). The 1-token child is the long one (6.7 s): one token means one compile at a time. The
  other two take about 2.7 s each with 4 workers. Collecting and comparing take hundredths of a second.
- **TestTSGoCacheSanitized** is 92% one phase: the cold `tsgoRuntimeLibrary`, 12.1 s, which compiles the sanitized
  tsgo runtime archive (73 translation units) **one at a time**. `cachedRuntime` calls `compileRuntime` with
  `workers` 1 (`internal/native/library.go:445`). The build log of a separate cold run shows 73 `cCompile miss`
  lines whose clang seconds sum to 10.05 s, inside a 12.5 s test, so the compiles did not overlap. Everything else
  (load and lower, Node, stubs, both BuildTSGo links, both binary runs, the warm lookup) adds up to about 1 s.

What is observed and what is inferred: the timings, counts and build log are observed. That the tsgo compiles are
serial because of `workers` 1 is read from the code and agrees with the build log's sum; I did not change the
worker count to confirm it. I did not measure a box busier than this one, or either test inside a full
`internal/native` run, where another test with the same sanitized tsgo runtime key could build it first (or hold
the in-process lock while it does). `units_tsgo_test.go`'s `splitTSGoOptions` is also `Sanitize: true`; whether
it shares the key was not checked.

## The box

| | |
|---|---|
| CPU | Intel(R) Xeon(R) Processor @ 2.80GHz, `nproc` 4 |
| Memory | 15 GiB |
| Kernel | Linux 6.18.44-fc-v153 |
| `/proc/loadavg` at the start | `0.32 0.08 0.03 1/108 352` (each log records it before and after its run) |
| clang | 20.1.8 from `LLVM-20.1.8-Linux-X64.tar.xz`, extracted to /opt/llvm20 without `bin/flang*`, `bin/mlir-*`, `bin/lldb*`, `bin/llvm-bolt*`, `bin/clangd`, `lib/lib*.a`; its `bin` first on PATH |
| Node | v24.22.1 (linux-x64 tarball, sha256 checked against SHASUMS256.txt) |
| Go | go1.27.1 (go.mod: `go 1.27`) |
| User | root, `TMPDIR=/tmp/adamic-gate` (mode 1777) |

Setup: `git worktree add --detach <dir> e50e502c21f12f9df64f7e716e980ed84695aa6f`, then
`git submodule update --init --recursive --depth 1`. The package's test binary was built once beforehand
(`go test -count=1 -run '^$' ./internal/native`) so no run paid for Go compilation inside its wall clock; the
`--- PASS` seconds exclude it anyway.

## Exact commands

Every run went through `native-slow-tests/run.sh <worktree> <label> <go test arguments>`, which sets
`PATH=/opt/llvm20/bin:/opt/node24/bin:/usr/local/go/bin:/usr/bin:/bin`, `TMPDIR=/tmp/adamic-gate`, unsets
`ADAMIC_BUILD_OUTPUT_DIR`, points `ADAMIC_BUILD_CACHE_DIR` at a fresh empty directory from `mktemp -d`, runs
`go test ./internal/native <arguments>` from the worktree root into `<label>.log`, and removes the cache after. Runs
were one after another, never overlapping.

```
# alone, unmodified tests, three times each (plain-*.log)
go test ./internal/native -count=1 -v -run '^TestTheJobserverNeverReachesTheBytes$'
go test ./internal/native -count=1 -v -run '^TestTSGoCacheSanitized$'

# the same, in the instrumented copy, three times each (phased-*.log)

# both at once as Loom's unit runs them, three times unmodified (together-N.log), three instrumented (together-phased-N.log)
go test ./internal/native -count=1 -v -parallel 6 -run '^(TestTheJobserverNeverReachesTheBytes|TestTSGoCacheSanitized)$'

# one extra unmodified tsgo run with the build log on, not among the three (buildlog-tsgo.log, buildlog-tsgo.buildlog.txt)
ADAMIC_BUILD_LOG=/tmp/adamic-gate/tsgo-buildlog.txt go test ./internal/native -count=1 -v -run '^TestTSGoCacheSanitized$'
```

## Each test alone, unmodified

`--- PASS` seconds, cold.

| Test | Run 1 | Run 2 | Run 3 | Median |
|---|---|---|---|---|
| TestTheJobserverNeverReachesTheBytes | 13.67 | 12.62 | 13.29 | **13.29** |
| TestTSGoCacheSanitized | 13.93 | 14.25 | 13.55 | **13.93** |

## TestTheJobserverNeverReachesTheBytes by phase

Instrumented, alone, cold, three runs. "Child" is the `inChild` call: a child test process doing
`request.Build(directory, 4, jobserver)` on the release runtime. "Collect" is reading the archive directory,
hashing every file and reading the recorded clang commands.

| Phase | Run 1 | Run 2 | Run 3 | Median (s) |
|---|---|---|---|---|
| child, no jobserver | 2.82 | 2.72 | 2.58 | **2.72** |
| collect, no jobserver | 0.02 | 0.02 | 0.02 | 0.02 |
| child, 1 token | 6.59 | 6.69 | 6.81 | **6.69** |
| collect, 1 token | 0.01 | 0.02 | 0.01 | 0.01 |
| child, 8 tokens | 2.59 | 2.79 | 2.57 | **2.59** |
| collect, 8 tokens | 0.02 | 0.02 | 0.02 | 0.02 |
| comparison | 0.00 | 0.00 | 0.00 | 0.00 |
| `--- PASS` | 12.06 | 12.26 | 12.02 | 12.06 |

Each run's archive directory held 162 files, **64 of them `.o`**, and the clang wrapper recorded 66 invocations,
identical in all three children of all three runs. (`internal/native/runtime/` has 73 `.c` files at this sha;
the release archive compiles 64 of them.)

## TestTSGoCacheSanitized by phase

Instrumented, alone, cold, three runs. Each phase is the time since the previous mark, so they add up to the test.

| Phase | Run 1 | Run 2 | Run 3 | Median (s) |
|---|---|---|---|---|
| load.Load plus lower.Lower | 0.06 | 0.07 | 0.06 | 0.06 |
| C(program) | 0.00 | 0.00 | 0.00 | 0.00 |
| node run | 0.16 | 0.14 | 0.10 | 0.14 |
| stubs compile and archive | 0.02 | 0.02 | 0.02 | 0.02 |
| tsgoRuntimeLibrary (cold) | 11.89 | 12.13 | 12.06 | **12.06** |
| stat and mkdir before first | 0.00 | 0.00 | 0.00 | 0.00 |
| BuildTSGo first | 0.38 | 0.42 | 0.38 | 0.38 |
| binary run first | 0.02 | 0.02 | 0.01 | 0.02 |
| first round's checks and mkdir before warm | 0.01 | 0.01 | 0.01 | 0.01 |
| BuildTSGo warm | 0.37 | 0.38 | 0.33 | 0.37 |
| binary run warm | 0.02 | 0.01 | 0.01 | 0.01 |
| warm round's binary checks | 0.01 | 0.01 | 0.01 | 0.01 |
| tsgoRuntimeLibrary (second) | 0.04 | 0.03 | 0.03 | 0.03 |
| `--- PASS` | 12.98 | 13.25 | 13.03 | 13.03 |

Both BuildTSGo calls find the runtime archive the cold `tsgoRuntimeLibrary` made (the build log shows 3
`cArchive hit` after the one `cArchive miss`), so each is only main.c's compile and the link.

## Both at once, `-parallel 6`

`--- PASS` seconds, cold, one fresh cache shared by the two tests in each run.

| Test | Run 1 | Run 2 | Run 3 | Median |
|---|---|---|---|---|
| TestTheJobserverNeverReachesTheBytes, unmodified | 12.37 | 12.27 | 12.76 | **12.37** |
| TestTSGoCacheSanitized, unmodified | 13.54 | 12.69 | 13.73 | **13.54** |
| TestTheJobserverNeverReachesTheBytes, instrumented | 12.85 | 12.81 | 13.17 | 12.85 |
| TestTSGoCacheSanitized, instrumented | 13.43 | 13.73 | 13.63 | 13.63 |

Phase medians of the instrumented runs together, against alone:

| Phase | Alone | Together |
|---|---|---|
| child, no jobserver | 2.72 | 3.05 |
| child, 1 token | 6.69 | 6.83 |
| child, 8 tokens | 2.59 | 3.00 |
| tsgoRuntimeLibrary (cold) | 12.06 | 12.55 |
| BuildTSGo first | 0.38 | 0.40 |
| BuildTSGo warm | 0.37 | 0.37 |

Run together on 4 cores, each test slows by under a second. The serial tsgo compile leaves cores free for the
jobserver test's children, and the two build different archives (release without tsgo, sanitized with
`-DADAMIC_TSGO`), so neither finds or waits on the other's.
