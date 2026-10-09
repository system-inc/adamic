# Concurrency scaling

Stacked on 137afa5. Claims: runtime/string_index.c and share.c for lazy shared string indexes; runtime/parallel.c for adaptive range sizes; runtime/adamic.h and object.c plus runtime slot-cache declarations and the one declaration hook in emit_objects.go for packed atomic caches. Parallel C harnesses and Go tests, measurement scripts and this report will hold the proofs and measurements. No lower or JavaScript changes.

## Implementation

Sharing now traverses and marks the graph without counting string units or building indexes. The first worker that needs a string's UTF-16 view builds a complete private candidate and publishes its pointer with a release CAS. Readers acquire that pointer. A losing builder frees both its BMP view and its candidate. Length, checkpoints and the BMP view are published together; the original plain length field remains immutable after sharing. Shared cursors remain disabled. Empty, short and ASCII shared strings use a small length-only candidate. Unshared strings retain their existing caches and cursor.

Warm shared length and BMP reads are inline. The string index layout is internal to the C runtime; moving it into adamic.h lets a BMP read use one acquired pointer without another runtime call. string_decode_impl.h and string_slice_impl.h use the same published length helper. The three-argument parallelMap ABI and adamic_share ABI are unchanged.

Each scope chooses `max(1, min(256, items / (8 * threads)))`. Stealing and range claiming use that scope's grain, including nested calls. At 4,096 items and 16 threads this is 32 rather than 256; at a million items and four threads it remains 256. `adamic_parallel_grain` is a private runtime observation in parallel.h, not a compiler ABI.

A slot cache is now one naturally aligned, lock-free 64-bit word: the shape address in the low 48 bits and the slot index in the high 16. Reads and writes are relaxed atomics. Field and method lookup use one snapshot; method lookup does not reread the cache after another thread can replace it. A compile-time assertion requires lock-free 64-bit atomics, and a runtime check refuses a shape address outside 48 bits with `adamic: panic: shape address exceeds 48 bits`. Slots greater than 65,535 bypass the cache. The emitted declaration and the caches in map.c and exceptions.c are ordinary static caches; library_object.c's local initializer is adjusted to the new representation.

## Proofs

The existing parallel memory, map, nested-map, exception, lifecycle, panic and million-element harnesses run under ASan/UBSan, sanitized slabs, counted, malloc and TSan builds. The scaling boundary harness additionally checks that sharing leaves caches cold, that empty/short/ASCII/indexed strings read correctly, the grain choices, large-slot fallback, and the address guard. Alternating shapes stress both field caches and dispatch between an object's closure field and a class method.

A snapshot-only hook holds four builders immediately before CAS publication. This forces exactly three losing candidates without adding instrumentation to production. It passes each build variant; omitting the losing-copy frees fails LeakSanitizer. Darwin uses the existing unsanitized malloc build under `leaks --atExit -- <binary>`; this Linux worker cannot run macOS tests.

The first full gate exposed a scheduling-sensitive new mutant: a plain read of the cache pointer was not always observed racing with publication. That detector was replaced by non-atomic publication in the forced-four-builder harness. The final report records the repeated proof and reruns rather than treating the original mutant as sufficient.


## Measurements

Linux amd64, AMD EPYC 9V74. `nproc` is 5; affinity permits five CPUs, while `cpu.max` is `400000 100000`, so the pool defaults to four. Five is an explicit override measurement. Parent is exactly 137afa5. All release samples are interleaved, alternating version order, best of five. Every run's stdout is checked against the other version and thread counts. These are observations on homogeneous Linux cores, not a measurement of the Mac's efficiency-core tail.

Load (1/5/15 minute) for the release comparison: 0.11/3.31/6.10 at start, 1.01/3.14/5.93 at end. No tests or other builds ran during measurement. Raw samples, output hashes and spreads are in [concurrency-scaling-results.json](concurrency-scaling-results.json).

| Program | Threads | Before seconds | After seconds | Change |
|---|---:|---:|---:|---:|
| trees | 1 | 1.618343 | 1.585996 | -2.0% |
| parallel_files | 1 | 0.732378 | 0.685402 | -6.4% |
| parallel_files | 2 | 0.446724 | 0.394921 | -11.6% |
| parallel_files | 4 | 0.329601 | 0.256097 | -22.3% |
| parallel_files | 5 | 0.293238 | 0.253547 | -13.5% |

Trees' best time is 2.0% faster, inside the observed 10.7%/10.8% before/after spreads. parallel_files' spreads range from 7.7% to 30.8% before and 12.5% to 26.9% after; the table gives the requested best-of-five observations, not a confidence interval.

| Million-number map call | Before ms | After ms | Change |
|---:|---:|---:|---:|
| 1 threads | 7.486 | 7.129 | -4.8% |
| 2 threads | 5.073 | 5.206 | +2.6% |
| 4 threads | 4.483 | 4.521 | +0.8% |
| 5 threads | 4.690 | 4.364 | -6.9% |

The small slowdowns at two and four threads are inside the measured scatter (35.6% and 172.6% after). Grain remains 256 for the million-number map. Claiming consumes a batch from a remaining range; range allocation occurs on a steal, not per element. At 4,096 items on 16 threads the maximum batch becomes 32 rather than 256, providing at least 128 batches instead of a nominal 16. The C grain fixture refuses the old fixed-256 implementation.

To isolate cache preparation, a second C workload scans all UTF-16 units in 4,096 independently allocated 8 KiB non-ASCII strings (32 MiB UTF-8). Its share-plus-map timings are:

| Threads | Before ms | After ms |
|---:|---:|---:|
| 1 | 168.030 | 124.949 |
| 2 | 108.420 | 63.798 |
| 4 | 84.261 | 34.868 |
| 5 | 79.322 | 27.594 |

The real program was also built against isolated runtime snapshots with timing hooks immediately around parallelMap's items/work sharing. Instrumentation is absent from production. Every instrumented run produces the same program output. Load at measurement start/end: 0.07/0.16/2.16 and 0.40/0.23/2.14. Raw data is [concurrency-scaling-share.json](concurrency-scaling-share.json).

| Actual parallel_files serial sharing | Before ms | After ms |
|---:|---:|---:|
| 1 threads | 54.340 | 0.396 |
| 2 threads | 58.123 | 0.375 |
| 4 threads | 60.069 | 0.375 |
| 5 threads | 58.046 | 0.378 |

At four threads, serial sharing falls from 60.069 ms to 0.375 ms, 99.4% less. Cache construction now runs in the callbacks. The Mac's original 27 to 29 ms is a different machine's observation; it has not been remeasured here.

Counted totals are unchanged on both workloads at one thread. [concurrency-scaling-counts.json](concurrency-scaling-counts.json) preserves the exact lines: parallel_files has 3,952,651 allocations and frees, 8,151,046 retains, 8,654,931 releases and peak 8,839; trees has 68,332,244 allocations, 1,572,900 frees, 66,759,344 region allocations, 2 retains, 44 releases and peak 2,097,149. Index metadata uses ordinary malloc and is covered by the leak proofs rather than these value counters. Result references retain the existing policy: mark their whole graph shared before the parent sees them, because a result may alias items or captures.

## Commands and results

Toolchain setup, logged to `/workspace/concurrency-scaling-setup.log`:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (20s)
setup: done in 20s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, clang 20.1.8, Node 24.19.0, Bun 1.3.14. Commands source `/workspace/adamic-tools/env.sh` first. Test output always went directly to log files.

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /workspace/concurrency-scaling-gate.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/native/... ./internal/oracle/... > /workspace/concurrency-scaling-final-gate.log 2>&1
go test -count=5 -v -run 'TestParallelChecksCatchMutants/lazy_cache$' ./internal/native > /workspace/scaling-cache-mutant-repeat.log 2>&1
go test -count=1 -v -run 'TestParallel(ChecksCatchMutants|ScalingGuardMutants)$' ./internal/native > /workspace/concurrency-scaling-final-mutants.log 2>&1
go vet ./internal/native/... > /workspace/concurrency-scaling-vet.log 2>&1
```

The first full gate finished with one failing scheduling-sensitive mutant; all other packages passed, including oracle (413.337 s). After correcting the detector and adding the inline BMP read, the full touched-package gate passed:

```text
ok github.com/system-inc/adamic/internal/native 368.162s
ok github.com/system-inc/adamic/internal/oracle 385.516s
```

The publication mutant was caught five consecutive times (33.732 s). The final combined thirteen-mutant run passed in 4.268 s. Vet and diff checks were clean. The complete final native suite includes ASan/UBSan, sanitized size classes, counted, malloc and Linux TSan harnesses; oracle includes the Node comparison, count table, parallel TSan variants, the ordering negative control and compiler refusals. The complete repository gate was not repeated after the correction; the full native and oracle packages were.

| Mutant run | What caught it |
|---|---|
| Skip items sharing | TSan data race in retain |
| Plain shared count increment | TSan data race in retain_slow |
| Plain slot-cache read | TSan data race against slot-cache store |
| Free directly into the foreign owner's lists | TSan data race in give_local |
| Reverse result slots | C ordered comparison, exit 3 |
| Skip revisiting shared containers | Lifecycle assertion, SIGABRT |
| Create a worker for ADAMIC_THREADS=1 | Caller/worker assertion, SIGABRT |
| Non-atomic index publication, four forced builders | TSan data race in shared_index; five repeated catches |
| Fixed grain 256 | Grain assertion, SIGABRT |
| Prepare strings at share time | Cold-cache assertion, SIGABRT |
| Truncate a large slot into the cache | Large-slot assertion, SIGABRT |
| Remove the 48-bit pointer guard | Expected panic status/message check; mutant instead aborts |
| Omit losing-copy frees | LeakSanitizer, exit 1, with three forced losing copies |

Build baseline binaries with the 137afa5 compiler and current binaries with this branch. For C harnesses, run `build_measure.go` in each checkout. The helper's optional `-runtime` flag builds an isolated supplied runtime snapshot for instrumentation. The exact measurements were:

```sh
python3 internal/native/testdata/parallel/measure_scaling.py --before /workspace/scaling-before --after /workspace/scaling-after --threads 1,2,4,5 --rounds 5 --output docs/concurrency-scaling-results.json > /workspace/concurrency-scaling-measurements.log 2>&1
python3 internal/native/testdata/parallel/profile_share.py --before-root /workspace/adamic-scaling-before --after-root /workspace/adamic --before-compiler /workspace/scaling-before-compiler --after-compiler /workspace/scaling-after-compiler --threads 1,2,4,5 --rounds 5 --output docs/concurrency-scaling-share.json > /workspace/concurrency-scaling-share-profile.log 2>&1
```

The release driver expects `PREFIX-files`, `PREFIX-trees`, `PREFIX-map` (built from map.c), and `PREFIX-share` (built from share_measure.c). It records every sample and verifies program bytes. Both Python drivers and their generated C builds were run successfully.

On the 16-core Mac, run on this branch:

```sh
go run ./bench -only parallel_files -threads 1,2,4,8,16 -rounds 5 > parallel-files-scaling-mac.log 2>&1
go run ./bench -only trees -rounds 5 > trees-scaling-mac.log 2>&1
go test -count=1 -timeout 30m ./internal/native/... ./internal/oracle/... > concurrency-scaling-mac-tests.log 2>&1
```

The existing benchmark runner includes Node and Bun and loads parallel_files.a correctly. For an interleaved Mac parent comparison, build the release binaries on both checkouts and run `measure_scaling.py` with `--threads 1,2,4,8,16`; its platform checks support Darwin. Mac TSan availability is probed by the existing harness helper. No Mac or efficiency cores were available in this worker, so those results remain for Kirk's machine.

Implementation commits: 9cad650 (claims), 1dc2a7f (lazy caches), 44ff0fe (grain and packed caches), fff6bfb (inline views and deterministic proofs). The final measurements/report commit follows them on codex/concurrency-scaling, stacked on 137afa5. No lower or JavaScript changes.
