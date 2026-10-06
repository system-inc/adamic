# Concurrency runtime, part 1: C evidence

Runtime is implemented on `codex/concurrency`. The compiler and Node backend are on another branch; native source lowering and the parallel oracle variant remain the agreed third pass. No files in `internal/lower` or `internal/javascript` changed. The approved design commit `099f36d` remains.

## ABI and ownership

```c
adamic_array *adamic_parallel_map(adamic_array *items, adamic_closure *work, bool references);
void adamic_share(void *value);
```

The third argument is @system_adamic_runtime's final ABI decision. It supplies the result array's reference flag directly; there is no added closure metadata. Items and work are borrowed through the join. The callback receives the array's item union and the numeric index; its return is owned. Every reference result is marked shared before publication, including aliases and fresh object graphs. No unique handover is assumed. On an exception, the lowest throwing input index wins after all children finish, and owned results, including a returned reference accompanying an exception, are released.

The caller is one of N executors; N-1 fixed pthread workers are started lazily. `ADAMIC_THREADS=1` creates zero threads. Linux uses online CPUs capped by affinity and cgroup v2 quota, with common cgroup v1 paths as fallback; macOS uses online CPUs. The override wins. Deques hold ranges, stolen ranges split in half, and execution claims 256-element chunks. Nested joins help execute work. Exit joins workers before heap cleanup and the leak check. A direct C probe observed `unset: threads 4 workers 3`, `1: threads 1 workers 0`, `2: threads 2 workers 1`, `4: threads 4 workers 3`.

## Hazard audit

- Giving lists, spare lists, recursive freeing queues, exception state and stack limits are thread-local. Worker stacks have their own 8 MiB stack limit with a guard margin.
- Chunk numbers are atomic. A fixed two-level atomic pointer table never moves; pages and chunk pointers are published with release/acquire. Chunks record monotonically numbered owners. Foreign frees publish an atomic remote list, drained by the owner when allocating or exiting. Remote nodes are out of line so the whole freed slot remains ASan-poisoned, including its first word.
- The high reference-count bit permanently marks sharing; the header remains 16 bytes. Unshared counts retain plain increments/decrements. Shared retain is relaxed; release is release, with an acquire fence before destruction. The count query masks the tag and loads with acquire. Generated Perceus and string reuse use that query; a held reference with count one permits unique reuse even after sharing. Immortal zero counts stay immortal. Region zero counts have a separate marker so their children are visited.
- Shared containers are always revisited on publication: unique reuse can add a fresh child before the caller creates another alias. An iterative walk with a visited set covers deep and cyclic graphs. Shared scalar/string leaves can be skipped. A graph preparation mutex prevents publishing a partially marked graph.
- Heap string length/checkpoint/BMP caches are prepared before sharing, and shared cursors never mutate. Immortal literals and stack pieces skip mutable caches. A uniquely reused shared string whose cache was cleared uses uncached reads afterward.
- Emitted call-site caches and both runtime static cache sites (exceptions and Map reads) are thread-local. Automatic caches already belong to the invocation.
- Readonly Map iteration's hidden iterator tally is atomic. Weak's global side table has a mutex and an atomic empty-table fast path; shared target destruction uses that same protection. Weak handles still cannot cross tasks in part 1.
- Output writes and explicit flushes take a mutex. One atomic panic winner flushes buffered output, prints once and exits 70; losing panic callers wait for process termination. Workers block termination signals so the caller handles them. Fatal panics use the existing `_exit` behavior; normal exit performs pool join and leak checking.
- ADAMIC_COUNT uses atomic tallies and an atomic maximum for peak live values. Tallies are exact. A parallel peak, and work performed after an exception, can depend on scheduling; there is no promise that those counts equal the one-thread run. All six existing sequential benchmark counts remain byte-for-byte identical to main.
- Other audited state: sort state is per invocation; library identities and math tables are immutable; argument globals are initialized before source execution. RegExp's test-only step-limit setter must be configured before work starts; ordinary source cannot call it. Source-level I/O and mutation are the compiler half's responsibility.

## Controls and mutants

`internal/native/parallel_test.go` builds C harnesses with ASan+UBSan and leak detection, Linux TSan alone, and plain ADAMIC_COUNT. Controls cover numbers; long BMP and supplementary strings and UTF-16 views; readonly objects/nested arrays; a captured Map; fresh reference results; identical one/four-thread bytes; nested maps and nested exceptions; lowest-index exceptions; a million numbers; buffered worker panic; region roots; empty/singleton maps; deep graphs; shared unique string reuse; and fresh descendants after container reuse. Slab tests force cross-thread frees and growth beyond the first 1,024-chunk page.

Every mutant compiles in an isolated runtime snapshot and must produce its particular failure. A timeout or build failure does not count as catching it.

| Mutant actually run | What caught it |
| --- | --- |
| Skip items sharing | TSan race in `adamic_retain`, exit 66 |
| Plain retain count on a shared value | TSan race in `adamic_retain`, exit 66 |
| Non-thread-local emitted field cache | TSan race in `adamic_object_find`, exit 66 |
| Foreign free directly into owner's lists | TSan race in `give_local`, exit 66 |
| Reverse result slots | Number comparison at index 0, exit 3 |
| Replace the no-worker path with two executors | Caller-thread/worker-count assertion, SIGABRT |
| Skip revisiting shared containers | Fresh descendant sharing assertion, SIGABRT |

Two earlier broad runs exposed weaknesses in the proof harnesses. The missing-mark mutant could be masked by early result sharing; a rendezvous now forces concurrent cold reads before any result publication. The torn-cache mutant could crash before TSan diagnosed it; both alternative shapes now have valid reference slots so the race is diagnosed. The repaired sharing checks passed three repeats, and the cache check ten repeats, before the final gate.

## Commands and observed outputs

Toolchain: `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`. Setup timing lines: Go 0s, clang 1s, Node 1s, submodules 1s, cache 80s, total 80s; `nproc` 5, `cpu.max` `400000 100000`, 17.6 GB. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.

```sh
go test -v -count=1 -timeout 15m ./internal/native -run TestParallel > /tmp/adamic-runtime-abi-proofs.log 2>&1
# PASS, 22.091s; all controls and the six required mutants
go test -v -count=1 -timeout 5m ./internal/native -run TestParallelChecksCatchMutants/reused_graph > /tmp/adamic-runtime-reuse-mutant.log 2>&1
# PASS, 2.901s; seventh mutant aborts
go test -count=1 -timeout 30m ./internal/native/... > /tmp/adamic-runtime-native-final.log 2>&1
# ok github.com/system-inc/adamic/internal/native 119.142s
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -parallel 4 ./internal/oracle > /tmp/adamic-runtime-oracle-final.log 2>&1
# ok github.com/system-inc/adamic/internal/oracle 122.481s
go vet ./... > /tmp/adamic-runtime-vet-final.log 2>&1
# exit 0, no diagnostics
git diff --check
# exit 0; gofmt also clean
```

A further forced-slab check compiled `memory.c` against `runtime/*.c` with `-std=c11 -Wall -Wextra -Werror -pedantic -O1 -g -pthread -fsanitize=address,undefined -fno-sanitize-recover=all -DADAMIC_SLABS -DADAMIC_COUNT`, then ran with `ASAN_OPTIONS=detect_leaks=1`, logging to `/tmp/adamic-remote-slabs.log`. It exited 0: `memory clean`; counts were allocations 720004, frees 720004, retains 80003, releases 760003, peak 300013, regions 0. This explicitly covers remote lists with poisoned slab slots.

The full native gate includes all seven mutants. Test output was written directly to logs. Baseline is a detached worktree at `5d4c801`. Its cohere directory points to the installed submodule. Building that checkout initially failed while stamping VCS status; `go build -buildvcs=false` worked around the worktree/submodule status issue.

## Performance observations

Best of five, same clang flags and machine, measured after gates finished. The portable runner's load averages before/after the maps and sequential programs were **0.66/2.03/1.86** and **0.81/1.91/1.83**. Quota is four CPUs despite `nproc` reporting five. All five samples, stdout and count lines are in `internal/native/testdata/parallel/measurements.json`, including the earlier run for comparison. These are observations on a shared cloud machine, not a statistical speed guarantee.

The string workload maps 32,768 entries alternating two deterministic long non-ASCII strings, reads their UTF-16 units and produces reference results. Number workload maps 1,000,000 elements. Timed maps include first pool startup and sharing. Process wall includes input preparation, verification and cleanup.

| Workload, milliseconds | 1 thread | 2 threads | 4 threads |
| --- | ---: | ---: | ---: |
| million, map only | 6.964 | 5.261 | 4.617 |
| million, process wall | 13.989 | 12.513 | 11.589 |
| strings, map only | 141.650 | 76.206 | 44.144 |
| strings, process wall | 143.594 | 78.780 | 46.251 |

Existing sequential programs, seconds; both backends' native binaries produce identical stdout:

| bench/ program | main | runtime branch | Change |
| --- | ---: | ---: | ---: |
| nbody | 0.331453 | 0.341920 | +3.16% |
| trees | 1.601757 | 1.636039 | +2.14% |
| spectral_norm | 0.170110 | 0.163550 | -3.86% |
| sort | 0.224637 | 0.224359 | -0.12% |
| word_count | 0.153240 | 0.168652 | +10.06% |
| tokenizer | 0.159594 | 0.166566 | +4.37% |

**There is a single-thread regression.** Word count is 10.06% slower in this run, tokenizer 4.37%, nbody 3.16% and trees 2.14%. The earlier run instead showed spectral norm +9.29% and sort +5.97%, which did not reproduce here; their variability is preserved in the raw data. The retain/release fast path also costs more despite keeping plain counts: a 50-million-pair loop is 26.75% slower (about 1.017 ns extra per pair). This unit does not claim single-thread performance neutrality.

| Microbenchmark | main seconds | branch seconds | Extra ns/operation | Change |
| --- | ---: | ---: | ---: | ---: |
| unshared retain/release pairs (50,000,000) | 0.190027604 | 0.240867690 | 1.017 | +26.75% |
| Weak lookup, empty table (20,000,000) | 0.028448678 | 0.030953918 | 0.125 | +8.81% |
| Weak lookup, populated table (20,000,000) | 0.040589105 | 0.172442499 | 6.593 | +324.85% |

The populated Weak-table mutex adds about 6.593 ns per lookup in this uncontended measurement. Multiworker memory harnesses exercise its correctness under contention; a contended throughput benchmark was not run.

Counts match exactly for both versions:

| Program | Allocations | Frees | Retains | Releases | Peak | Regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| nbody | 8 | 8 | 7000019 | 7000023 | 7 | 0 |
| trees | 68332244 | 1572900 | 2 | 44 | 2097149 | 66759344 |
| spectral_norm | 4 | 4 | 60 | 67 | 4 | 0 |
| sort | 6 | 6 | 2000002 | 2000010 | 4 | 0 |
| word_count | 1020129 | 1020129 | 6100372 | 5100429 | 1020006 | 0 |
| tokenizer | 1194976 | 1194976 | 2405063 | 2400020 | 597494 | 0 |

Reproduce, including builds and five interleaved runs per sequential version:

```sh
source /workspace/adamic-tools/env.sh
python3 internal/native/testdata/parallel/measure.py /tmp/adamic-concurrency-main /tmp/adamic-runtime-measurement --threads 1,2,4 > /tmp/adamic-runtime-measurement.log 2>&1
# exit 0; outputs/counts match; measurements.json contains every sample
```

On the 16-core Mac, pass `--threads 1,2,16` and use its Go/clang toolchain for both checkouts. The C harnesses and runner are committed; no source-level parallel lowering is needed to measure this runtime.

## Commits and limits

- `099f36d`: approved design.
- `acefbbd`, `21a3922`, `b26c2c7`: runtime claims, pushed before implementation.
- `aef9598`: heap/thread-local state, counts, caches, Weak and strings.
- `811edc4`: structured range-stealing pool and C controls.
- `b2221de`: six executable mutants.
- `da8d3bc`: force cold sharing overlap.
- `d0851d9`: fast count tag and lifecycle edges.
- `4629844`, `bfcecad`: final ABI and complete graph re-publication (the latter completes the ABI conversion of the lifecycle harness too).
- The final evidence commit adds this report, raw samples, the Weak cost harness and portable runner.

Everything is pushed on `codex/concurrency`; no PR opened. This runtime unit does not implement Shareable/purity checks, Node/front-end handling, native source lowering, the parallel oracle variant, or the cohere-shaped source benchmark. Those remain with the other worker/third pass by agreement. Existing oracle fixtures pass, but the new parallel behaviors are proven here from C, not from Node comparisons. macOS, cgroup v1 fallback and nested cgroup mount layouts were not validated on this machine; TSan evidence is Linux only. A whole-repository `go test ./...` and a contended Weak performance measurement were not run.

## Inline count follow-up

The original timing regressions above are historical. The requested branch merges, inline count path, new slab controls and repeat measurements are recorded in [concurrency-inline.md](concurrency-inline.md).
