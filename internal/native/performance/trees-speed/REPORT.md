# Tree build/walk runtime profile

The historical 1.24× native/Bun loss at `fb714279a32b5888019bba8723a15470344a2096`
is not principally per-node ARC or the slab allocator. On the requested area/runtime
base `ae57b2d84bc8ee572830a49f4e9a7f78737f7205`, 97.70% of allocations in this
unchanged workload are regional. Construction and walking emit **zero retains**.
Temporary trees bypass `heap_free_children` completely. The ordinary heap paths
serve the stretch and long-lived trees and 38 output strings.

Two small runtime costs were removed:

- `drain_remote` now loads an empty queue before trying to exchange it. This workload
  previously performed **2,761,011 empty atomic exchanges**. The load does not consume
  queue nodes; a nonempty queue still uses the acquire exchange paired with remote
  producers' release CAS. A concurrent publication missed by the load waits for a
  later drain, as a publication just after the old exchange did.
- `region->count` is updated only in `ADAMIC_COUNT` builds. Uncounted release code
  previously performed **66,759,344 unnecessary memory increments**. The field is
  diagnostic; teardown uses block cursors and `holds_outside`. Counted reports are
  byte-identical before/after.

**No wall-time speedup was established.** Final best native wall time was 1.586 s,
versus baseline 1.567 s (+1.19%), Node 2.271 s and Bun 1.418 s. Final native/Bun
is 1.119×; baseline native/Bun is
1.105×. User CPU on those best-wall samples fell
from 1.065 s to 0.983 s, while system CPU rose from 0.502 s to 0.602 s. The
instruction reduction and eliminated atomic work are concrete; these wall samples
on a shared machine do not establish a repeatable win. Remaining cost is primarily
regional build/initialization, recursive walking and region block allocation/free,
not retains. Explaining Bun's internal allocation/GC behavior would need a separate
Bun profile; this report does not infer it from native instruction counts.

## Machine, setup and method

- Machine `cb1632b4fc8c`; `Linux-6.18.44-x86_64-with-glibc2.41`; **INTEL(R) XEON(R) PLATINUM 8573C**.
- Five logical/affinity CPUs (0–4); cgroup `cpu.max = 400000 100000`.
- Go 1.27.1, clang 20.1.8, Node **v24.19.0**, Bun **1.3.14**, Python 3.12.14.
- `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh --wasi-sdk`
  completed; sourced `/workspace/adamic-tools/env.sh`.
- No AGENTS.md was present at the workspace/repository roots or in the repository.
  Fetched both `area/runtime` and `runtime/honest-benchmarks` by branch name with
  explicit refspecs. Created `runtime/trees-speed` from fetched `origin/area/runtime`.
- The historical machine was an AMD EPYC 9V74; this machine is a Xeon. The ratios
  above are same-machine comparisons, not a comparison of seconds across machines.
- Same `bench/trees.ts` as `fb714279`, SHA256
  `743b83a8bce42948f155d02dd8e939c94a2d7f403274b2b1981da381aa6c7f50`. Exact stdout is below.
- CLI release builds: `go run ./cmd/adamic build bench/trees.ts -o BINARY`.
  Current release policy is clang `-O2 -flto=thin`, linked with lld; the historical
  report describes `-O2`. Both baseline/final here use the current shipped policy.
- One untimed validation per binary, then **five interleaved rounds** of baseline,
  final, Node, Bun, rotating the first runtime each round. All children run
  sequentially in fresh processes. Best is minimum wall time; user/system CPU
  belongs to that same sample. Python `perf_counter` measures launch through wait;
  `getrusage(RUSAGE_CHILDREN)` deltas measure child CPU. Startup and output included.
  No pinning or isolation; no builds, tests, setup or profiling overlapped timings.
  `NODE_OPTIONS`, `BUN_OPTIONS`, `ADAMIC_THREADS` cleared for children.
- Load (1/5/15 min), final comparison: before
  **0.43 / 3.14 / 4.17**;
  after **0.73 / 2.84 / 4.02**.
  Every raw timing, command and binary hash is in `final.json`.

## Final best-of-five seconds

| Runtime | Wall | User | System |
|---|---:|---:|---:|
| before | 1.566986 | 1.064902 | 0.501532 |
| final | 1.585689 | 0.982685 | 0.601942 |
| Node | 2.270595 | 2.927380 | 0.230703 |
| Bun | 1.417596 | 1.909226 | 0.400181 |

## Every final timing sample

Cells are wall / user / system seconds in round order.

| Runtime | 1 | 2 | 3 | 4 | 5 |
|---|---:|---:|---:|---:|---:|
| before | 1.617501 / 1.046032 / 0.570688 | 2.238304 / 1.390197 / 0.847435 | 1.566986 / 1.064902 / 0.501532 | 1.612200 / 1.118867 / 0.492673 | 1.624002 / 1.097639 / 0.525652 |
| final | 1.738039 / 1.179403 / 0.557235 | 1.585689 / 0.982685 / 0.601942 | 1.639062 / 1.040041 / 0.598326 | 1.875662 / 1.191532 / 0.681227 | 2.722529 / 1.742895 / 0.978322 |
| Node | 2.428589 / 3.006382 / 0.253923 | 2.670450 / 3.525113 / 0.254999 | 2.603005 / 3.200006 / 0.324087 | 2.270595 / 2.927380 / 0.230703 | 2.466340 / 3.185265 / 0.237604 |
| Bun | 1.417596 / 1.909226 / 0.400181 | 1.417994 / 1.983053 / 0.356547 | 1.507636 / 1.956088 / 0.405000 | 1.787482 / 2.244443 / 0.296277 | 1.451868 / 1.960592 / 0.305425 |

The earlier slab-only exploratory run is retained in `slab-only.json`: best wall
baseline 1.742438, slab fix 1.556504,
Node 2.238320, Bun 1.425431. Its
system-time difference was substantial, and the apparent wall gain did not survive
the final run. It is not substituted for the final result.

## Allocator and counting costs

A separate diagnostic copy of the baseline runtime added path counters, not timers
(`diagnose.py`, `diagnostic.txt`). It used the same generated `trees.c`, `-O2`,
ThinLTO, `-ffp-contract=off`, `-fno-optimize-sibling-calls`, pthread and lld. Its stdout
matched the release checksum. Its timing is not a release measurement.

| Path | Exact measurement |
|---|---:|
| Logical allocations | 68,332,244 |
| Regional objects | 66,759,344 |
| Ordinary heap objects freed | 1,572,900 |
| Tree nodes in ordinary heap | 1,572,862 |
| Retain calls | 0 |
| Release calls emitted by compiler (including NULL/output) | 42 |
| Peak live logical allocations | 2,097,149 |
| Region ends | 349,520 |
| Region strong-child walks | 0 |
| Region blocks allocated and freed | 495,744 |
| Cumulative region block capacity (not RSS) | 6,973,751,296 bytes |
| Object bytes / allocated stride | 56 / 64 bytes |
| Slab takes | 1,572,900 |
| Fresh / reused slab slots | 1,572,866 / 34 |
| New chunks / reused spare chunks | 1,542 / 0 |
| Empty remote drain exchanges, baseline | 2,761,011 |
| `heap_free_children` invocations | 1,572,900 |
| Child `let_go` callbacks / NULL callbacks | 1,572,898 / 38 |

Heap nodes use class index 3: sizes 49–64 occupy 64-byte slots. A 65,536-byte chunk
has a 96-byte header and 1,022 such slots. The two trees remain alive together until
main's final releases; each heap tree node starts at count one, gets no retain,
and has exactly one last-owner decrement (1,572,862 total). Internal child drops
are not added to the public release-call counter. Regional nodes have count zero;
they have neither per-node decrement nor slot destruction at region end. The
regional object footprint is 4,272,598,016 used bytes; growing blocks request 6.97 GB
cumulatively, reflecting unused capacity as well. This is allocator churn, not a
6.97 GB simultaneous live set. Both allocators' layouts/growth policy are unchanged.

Field initialization is already efficient in the shipped LTO build. Each node
initializes count, packed kind/slab, shape, class and frozen plus two child slots.
Assembly folds kind/slab into one word; leaf slots use a single 16-byte zero store.
The ordinary heap allocator's zeroing is eliminated when both child slots are
immediately overwritten. Regional `filled_in` already avoids memset. No new
uninitialized-object API or speculative initialization removal was added.

### Release instruction profile

Callgrind 3.24.0 profiled the **actual CLI release binaries**, without count hooks
or sanitizers. `before.cg.gz` / `final.cg.gz`, annotations and assembly are retained.
Instruction counts are not cycle counts or wall-time attribution. Callgrind replaces
malloc and reported a brk-growth limitation; the diagnostic allocator paths and
native timings are measured outside Callgrind. Recursive inclusive costs can exceed
100%, so build/walk below aggregate **self** counts across their recursion contexts.

| Work | Baseline instructions | Share of baseline |
|---|---:|---:|
| Regional build, including inlined allocation/initialization | 3,377,096,048 | 61.36% |
| Tree walk | 1,432,879,194 | 26.03% |
| Slab `allocate_storage`, inclusive | 98,027,628 | 1.78% |
| Complete ordinary heap release drain, inclusive | 229,682,234 | 4.17% |
| `heap_free_children`, self / inclusive | 84,935,752 / 128,976,174 | 1.54% / 2.34% |
| Child count/drop/queue (`let_go`), self | 44,040,422 | 0.80% |
| Slab deallocation, self | 53,517,180 | 0.97% |
| Region teardown, inclusive | 111,238,592 | 2.02% |
| libc malloc / free, inclusive, all paths | 203,041,368 / 96,384,779 | 3.69% / 1.75% |
| Whole program | 5,503,698,925 | 100% |

Overlapping inclusive rows must not be added. `let_go` averages about 28 instructions
per callback, including queueing; it is an upper bound on pure child-counting work,
not an isolated count decrement benchmark. Whole-tree release is iterative and
still costs 4.17% of instructions, but covers only 2.30% of logical allocations.
Those destruction and per-child counting costs were measured, not changed.

Final whole-program instructions: **5,436,559,409**, down **67,139,516 (1.22%)**.
The regional build loses exactly **66,759,344** instructions, one increment per
regional object. Final slab allocation inclusive cost is **97,645,918**; the
atomic exchange's latency/cross-core ownership cost is not represented by a count
of x86 instructions. The workload has no cross-thread frees.

### Count instrumentation overhead (diagnostic only)

The CLI `--count` build disables ThinLTO, so its timing is not used to isolate count
hook overhead. A separate `-DADAMIC_COUNT` build kept release ThinLTO and matched
both stdout and the CLI count report byte for byte. Five interleaved fresh-process
rounds (`count-cost.json`, same timer method), best wall:

- Uncounted final: **1.600642 s**, user **1.000377 s**,
  system **0.600710 s**.
- Counted with ThinLTO: **2.674053 s**, user
  **2.111699 s**, system **0.559957 s**.
- Instrumentation added **1.073410 s** best-wall
  (1.67×). These are diagnostic overhead numbers,
  never native-vs-Bun release scores. Count hooks use atomic allocation/live counters
  and peak tracking; the much larger diagnostic cost is absent from release.
- Load before 0.37/2.48/3.85,
  after 0.59/2.36/3.77.

## Validation and reproduction

Commands, all after sourcing the setup environment:

```sh
go test ./internal/native -run '^(TestRemoteSlabDrainAndMutant|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestRuntimeReleasePaths|TestRegionEndWeakTargets|TestRegionEndThrowInitialization|TestReleaseSharedValueAndUnsafeMutant)$|^TestParallelMemory/(asan_slabs|count|tsan)$' -count=1 -v
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/^(region_end|call_targets_region)[.]a$' -count=1 -v
ADAMIC_TEST_WASI=1 go test ./internal/native -run '^TestWASIHostPromises$' -count=1 -v
go run ./cmd/adamic build bench/trees.ts -o /tmp/trees-profile/sanitized --sanitize
ASAN_OPTIONS=detect_leaks=1:halt_on_error=1 UBSAN_OPTIONS=halt_on_error=1 /tmp/trees-profile/sanitized
go run ./cmd/adamic build bench/trees.ts -o /tmp/trees-profile/counted-final --count
```

All passed. Full tree ASan/UBSan/leak run had empty stderr and matching stdout.
New `TestRemoteSlabDrainAndMutant` probes remote slot reuse in both a giving chunk
and an exhausted chunk. Its mutant skips nonempty drains, compiles, and exits 4
with `remote slot was not recycled`, rather than failing to compile or timing out.
The existing last-owner mutant also compiled and was caught by ASan. Concurrent
memory harness passed under ASan slabs, counted slabs and three TSan executions.
The WASI host-promise ownership control and its existing missing-release mutant
also passed with `ADAMIC_TEST_WASI=1` (`wasi.log`). Counted baseline/final reports
matched exactly (`counted*.stderr`). No oracle
fixture source or registration changed, so counts.md regeneration was not needed.
No whole-package tests or full gate were run.

`measure.py OUTPUT before=BASELINE final=FINAL Node=NODE Bun=BUN` reproduces the
five-round timing protocol (its source path records this workspace). `diagnose.py`
expects the generated C and a baseline runtime copy in `/tmp/trees-profile`;
its added path counters and compilation flags are visible for review. Baseline
runtime can be extracted with `git archive ae57b2d84bc8ee572830a49f4e9a7f78737f7205
internal/native/runtime`, and C generated with `go run ./cmd/adamic c bench/trees.ts`.
For profiles: `VALGRIND_LIB=... valgrind --tool=callgrind
--callgrind-out-file=FILE BINARY`; annotate with `callgrind_annotate --threshold=100`.

## Identical stdout from every validation and timed execution

```text
stretch tree of depth 19	 check: 1048575
262144	 trees of depth 4	 check: 8126464
65536	 trees of depth 6	 check: 8323072
16384	 trees of depth 8	 check: 8372224
4096	 trees of depth 10	 check: 8384512
1024	 trees of depth 12	 check: 8387584
256	 trees of depth 14	 check: 8388352
64	 trees of depth 16	 check: 8388544
16	 trees of depth 18	 check: 8388592
long lived tree of depth 18	 check: 524287
```
