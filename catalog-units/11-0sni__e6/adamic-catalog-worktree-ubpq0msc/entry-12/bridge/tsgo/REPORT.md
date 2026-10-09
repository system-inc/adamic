# Milestone 5 bridge report

Task k4fm1vf, branch `codex/tsgo-c-library`, based on main
`fe3b9f236e0672e968bcb7ad53c1882cdde18f29`.

Built a typescript-go C archive, an explicit-length UTF-8 C API with owned
outputs and guarded program IDs, three Adamic prelude functions, opt-in stage 0
archive linking, an independent Go oracle, measurements, and executable mutants.
Both ordinary owned query results and results allocated in a statement's region
are supported. The four protected compiler files were not edited. Usage,
contracts and reproduction commands are in [README.md](README.md).

Pinned dependencies: cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`, its
TypeScript/typescript-go submodule `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`,
and the original TypeScript corpus at v6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`.

## Setup observed

`bash cloud/setup.sh > /tmp/tsgo-setup.log 2>&1`, exit 0, followed by
`source /workspace/adamic-tools/env.sh` in each build/test shell. Setup printed:

```text
go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
v24.19.0
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (12s)
setup: done in 12s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` printed `5`. Setup succeeded without a workaround.

## Archive and oracle observed

`go build -buildmode=c-archive -o /tmp/tsgo.a ./bridge/tsgo/archive` succeeded,
producing a 46 MB archive on Linux amd64. The C boundary was also built with
`CC=clang` and
`CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all'`.
Both archives linked into stage 0's native Adamic output.

The final command was:

```sh
ADAMIC_TSGO_CORPUS=/tmp/tsgo-typescript go test -v -count=1 -timeout 30m ./bridge/tsgo > /tmp/tsgo-validation-final.log 2>&1
```

It exited 0 in 45.091 s and printed:

```text
C ABI: 100 queries, outputs survive release, stale and zero handles rejected, second handle distinct; "世界🌍"
build, C and JavaScript commands refuse unlinked checker calls
oracle: 1600 positions across 4 files, 54982 bytes identical under ASan/UBSan/LSan
region probe: Go and native answer 6; adamic: counts: allocations 9 frees 8 retains 6 releases 14 peak 8 regions 1
PASS
ok github.com/system-inc/adamic/bridge/tsgo 45.091s
```

The positions are 400 uniformly spaced byte offsets per compiler file:
`checker.ts`, `parser.ts`, `types.ts`, and `utilities.ts`. All three returned
fields are compared, including their byte lengths and complete string bytes.
Semantic diagnostics are not requested; the bridge and direct Go oracle ask
the same checker questions. This does not establish that every compiler file
passes semantic checking.

## Measurements observed

Three interleaved rounds, optimized native Adamic with the ordinary archive
against the direct Go oracle, after the full gate finished:

| Round | Native load ms | Native 1,600 queries ms | Go load ms | Go 1,600 queries ms |
|---|---:|---:|---:|---:|
| 1 | 216.147 | 249.300 | 225.868 | 254.242 |
| 2 | 240.195 | 243.897 | 217.076 | 238.116 |
| 3 | 218.514 | 233.554 | 235.585 | 251.412 |
| Median | 218.514 | 243.897 | 225.868 | 251.412 |

These are API-call timings, excluding formatting/output and whole-process
startup. Native includes C input copies, boundary crossings and creation of
Adamic result objects and strings. Checker work is lazy; query time includes
that work when first requested. The observed median differences are about 3%,
inside the variation here. No speedup is established by these three rounds.

## Mutants observed

Every row ran successfully through compilation and failed at its intended check.
Production mutations use overlays, leaving the checkout unchanged.

| Mutant | What caught it |
|---|---|
| Caller passes config length plus one over an exact-size heap buffer | ASan `heap-buffer-overflow` in the C input copy |
| Returned type buffer advertises length plus one | ASan `heap-buffer-overflow` when its bytes are consumed |
| Release retains the program in the live registry | Assertion that querying a released handle returns `TSGO_HANDLE` |
| Type is queried from the source-file node instead of the selected node | Go/native byte oracle, first mismatch at byte 6 |
| Lowering's link opt-in guard removed | `TestTSGoRequiresLink`: lowering accepted an unlinked checker call |
| C output free omitted | LeakSanitizer `detected memory leaks` |
| Region entry allocates its result on the heap | LeakSanitizer `detected memory leaks` in the counted region probe |

An earlier region mutant was masked: the corpus print helper used `utf8Length`,
which the existing region analysis conservatively treats as escaping. That run
failed the test's requirement to catch the mutant. The final separate region
probe reads only string lengths; its counted run proves one region is used, and
the mutant then leaks. The corpus fixture uses ordinary owned results, while
the region fixture holds the second entry independently.

## Repository gate

`gofmt -l cmd internal bridge` produced no output. `go vet ./...` exited 0 with
no output. `git diff --check` produced no output.

`go test -count=1 -timeout 30m ./... > /tmp/tsgo-gate.log 2>&1` exited 0.
The main oracle package reported 702.864 s, native 274.545 s, flow 90.747 s,
load 1.383 s and lower 6.490 s. Every package passed, including the stage 1
cohere slices. This gate began before the final region adapter/probe refinements;
the final bridge corpus, all mutants, touched packages, and a filtered region
oracle were rerun afterwards. The final commands and logs are:

```sh
go test -count=1 -timeout 30m ./cmd/adamic ./internal/load ./internal/lower ./internal/native > /tmp/tsgo-touched-final.log 2>&1
go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/^regions' > /tmp/tsgo-oracle-filtered-final.log 2>&1
```

The final touched-package command exited 0: load 1.562 s, lower 5.914 s and
native 65.268 s. The filtered oracle exited 0 in 13.393 s; both `regions.a`
and `regions_throw.a` passed against Node.

## Not covered

Windows, macOS, c-shared, concurrent-call stress, exhaustive AST positions,
allocation-failure injection, semantic diagnostic retrieval, first-class bridge
function values, and the complete cohere lint engine are not covered. The
external checker still owns Go's runtime and collector; releasing its program
ID drops roots rather than unloading that runtime. ASan/LSan cover C and Adamic
memory, not Go heap reclamation. The tested C archive platform works, so there is
no build-mode blocker to report.
