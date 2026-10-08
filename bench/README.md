# Benchmarks

Deterministic programs, the same source run three ways: Adamic's native binary (`adamic build`, clang `-O2`, no sanitizers), Node, and Bun when it's installed. Each program makes its own input from a fixed seed and prints a checksum, so there's no I/O in the timing and every runtime has to give the same answer.

```
go run ./bench                      # every program, 5 interleaved rounds, best of 5
go run ./bench -rounds 9 -only trees,nbody -timeout 300s
go run ./bench -only parallel_files -threads 1,2,4,8,16 -rounds 5
```

`-threads` interleaves native thread settings too; Node and Bun use the independent sequential `parallelMap` shim. Parallel counts use one thread so peak liveness is reproducible.

The runs are interleaved round by round, so a machine that slows down slows all three alike, and the best round is reported. New programs use `.a`; the runner discovers both `.a` and `.ts`. Node runs `.a` through oracle/node.mjs, which strips its types and supplies the sequential `parallelMap` shim, and Bun runs `.a` directly; neither runs Adamic's generated JavaScript in the timed comparison. Time is wall clock; memory is the process's peak resident set. The counted table comes from `adamic build --count`, run once, untimed: it's what explains a native time. The machine, the versions and the load before and after print with the numbers, and numbers without them aren't worth quoting.

| Program | What it exercises |
|---|---|
| `nbody.ts` | floating point on the fields of five objects, 1,000,000 steps |
| `spectral_norm.ts` | floating point over number arrays read and written by index, n = 1,500 |
| `tokenizer.ts` | a lexer walking a 3.5-million-code-unit text with `charCodeAt`, some of it non-ASCII |
| `word_count.ts` | `split`, then a `Map<string, number>` over a million words, then a sort |
| `sort.ts` | a million numbers sorted with a comparator, random and then nearly sorted |
| `parallel_files.a` | 4,096 deterministic in-memory files, pure tokenization and summaries in parallel, merged in order |
| `trees.ts` | binary trees: 68 million small objects allocated, walked and freed |
| `string_build.a` | 6,000 bounded formatter documents, short `+=` pieces, templates, array joins, slices, repeat and padding; every UTF-16 unit checksummed |
| `map_workload.a` | four cache/store batches, each with 2,048 initial entries and 16,384 mixed requests over runtime-built string keys, numeric record keys and an active-name Set |

## Reproducible wall and user time comparison

```sh
source /workspace/adamic-tools/env.sh  # use the path printed by cloud/setup.sh
bash bench/run.sh                     # validates outputs, times, writes bench/RESULTS.md
bash bench/run.sh --check-only        # builds and validates without timing
bash bench/run.sh --only primes,map_200k --output /tmp/bench-results.md
```

This script builds with `go run ./cmd/adamic build`, requires Node v24.19.0,
and locates Bun or installs Bun 1.3.14 from its GitHub release into a user cache.
Set `BUN` to an executable to select an installed version. All three runtimes are
required; a missing runtime, execution error, timeout or checksum disagreement
fails the run and leaves any prior report intact. Python 3 and curl are required.

The selected workloads are `primes.a` (ten Uint8Array sieves near two million),
`number_sum_sort.a` (full sums and sorts of two million-number arrays, adapted
from `sort.ts`), `string_build.a` and `word_count.ts` (existing building and
splitting programs), `map_200k.a` (200,000 distinct string keys), `trees.ts` and
`nbody.ts` (existing recursive allocation/walk and float workloads). These are
six workload categories, with two existing programs for the string category.
`json_encode.a` adds nested JSON encoding; decoding is not measured because
native refuses JSON.parse. An explicit JSON.stringify refusal is recorded as
not measured; any other build error stops the run.

One untimed output check precedes five fresh-process interleaved rounds. Every
measured run must also match stdout exactly. The starting runtime rotates each
round. The table reports minimum wall time and the user CPU time from that same
sample, both native/Node and native/Bun ratios, and explicit native losses. The
report includes machine, CPU/affinity/quota, versions, load before/after timing,
source checksums, exact outputs and every timing sample. Startup, type stripping,
workload checksum calculation and printing are included; builds are excluded.
[RESULTS.md](RESULTS.md) records the measured run and its limitations.

## Parallel files, October 6, 2026

`parallel_files.a` generates 4,096 files, tokenizes and summarizes them through pure functions, then merges in input order. Linux amd64, AMD EPYC 9V74, four quota-limited CPUs (`nproc` 5), clang 20.1.8, Node 24.19.0, Bun 1.3.14. Best of five interleaved rounds: native 1/2/4 threads **0.744/0.454/0.312 s**, Node **0.637 s**, Bun **0.866 s**. All answers match. Load before **0.11/3.16/6.23**, after **0.69/3.10/6.14**; tests had finished, but the longer load averages still include them.

[Every round](parallel_files.measurements.log) and [the integration report](../docs/concurrency-integration.md) include memory, counts, proofs, limits and the 16-core Mac command. Node uses oracle/node.mjs for `.a` sources; Bun 1.3.14 executes `.a` directly. Both resolve the independent sequential shim; existing `.ts` programs remain unchanged.

## Noisy cloud numbers, October 5, 2026

Not the record: a shared cloud container, 4 vCPUs (Intel Xeon @ 2.80GHz), Linux 6.18, load 0.74 before and 1.04 after. Across the five rounds one runtime's times on one program spread 7% to 40%, so differences under about 1.4x here mean little. The quiet Mac reruns this for the record. Adamic `0955c44`, clang 18.1.3, Node 24.21.0, Bun 1.3.14.

| Benchmark | native | Node | Bun | native vs Node |
|---|---|---|---|---|
| nbody | 1.452 s, 6.7 MB | 1.256 s, 72.1 MB | 0.201 s, 41.0 MB | 1.16x, slower |
| sort | 0.575 s, 32.2 MB | 1.183 s, 207.2 MB | 0.876 s, 95.0 MB | 0.49x, faster |
| spectral_norm | 0.565 s, 5.0 MB | 0.376 s, 70.0 MB | 0.419 s, 40.4 MB | 1.50x, slower |
| tokenizer | did not finish in 120 s | 0.383 s, 146.1 MB | 0.425 s, 88.6 MB | over 300x, slower |
| trees | 5.385 s, 98.4 MB | 2.405 s, 308.4 MB | 2.004 s, 207.4 MB | 2.24x, slower |
| word_count | 0.324 s, 78.3 MB | 0.434 s, 157.8 MB | 0.525 s, 84.0 MB | 0.75x, faster |

| Benchmark | allocations | frees | retains | releases | peak live |
|---|---:|---:|---:|---:|---:|
| nbody | 8 | 8 | 52,000,119 | 52,000,123 | 7 |
| sort | 6 | 6 | 6,002,000 | 6,002,008 | 4 |
| spectral_norm | 4 | 4 | 3,063 | 3,070 | 4 |
| tokenizer | did not finish | | | | |
| trees | 68,332,244 | 68,332,244 | 375,302,854 | 306,970,688 | 2,097,149 |
| word_count | 1,020,129 | 1,020,129 | 10,160,699 | 9,160,756 | 1,020,006 |

Native uses the least memory in every row, 3x to 14x less than Node. On time it wins two and loses four. What explains each loss, from the generated C and the runtime:

- **tokenizer: the string walk.** `charCodeAt(i)` and `length` decode the UTF-8 from the start of the string on every call (runtime/string.c), so walking a string with them is quadratic: 5,000, 10,000 and 20,000 tokens took 2.0, 8.2 and 30.8 s, ASCII-only text the same. The ASCII-only flag docs/memory.md plans, and a cached UTF-16 index for other strings, are what fix it.
- **trees: allocation and retain and release traffic.** Every node is its own malloc and free (68 million of each), and there are 5.5 retains per node. Node and Bun bump-allocate in a young generation and free a whole dead tree at once. Reuse in place and arenas (docs/memory.md) are the answer this program is waiting for.
- **spectral_norm: array reads through a call.** Every `vector[j]` calls `adamic_array_at`, in the runtime's own translation unit where clang can't inline it, and each checks the index with `trunc` and three comparisons of doubles. The loop counters are doubles too. Inlining the read (or link-time optimization) and integer loop counters, where range analysis proves them, are the fix.
- **nbody: field access through a call, and retains on every element read.** Every field read and write is a call to `adamic_object_field` with a slot cache, even on a class instance whose layout is fixed, and each `bodies[i] ?? ...` costs four retains and three releases (52 million in all). Fixed offsets for class fields and borrowed element reads are the fix. Bun's 0.2 s here is JavaScriptCore's optimizing compiler keeping the five bodies' fields in registers.

The two wins, sort and word_count, are TimSort over unboxed doubles and the runtime's map, with no JIT warm-up to pay.


## Cloud worker 0aja3b1, October 7, 2026

Machine `082e243695ca`: shared Linux/amd64 container, Intel Xeon Platinum 8573C,
5 logical CPUs (`nproc` = 5), cgroup quota 4 CPUs (`cpu.max` = `400000 100000`),
Linux 6.18.44. Compiler base `ef3d907` plus this unit's bench changes, clang 20.1.8,
Go 1.27.1, Node 24.19.0, Bun 1.3.14. Run began at 00:04:13 UTC. Load before:
`1.22 4.53 3.06`; after: `1.00 3.64 2.87` (1, 5, 15 minutes).
No setup, tests or profiles ran alongside the timing. These are noisy cloud measurements,
not a quiet-machine record. Memory belongs to the fastest run selected for each runtime.

Command: `source /workspace/adamic-tools/env.sh; go run ./bench -rounds 5`, exit 0,
best of five interleaved rounds. Every top-level program discovered by the runner was run.
There is no `parallel_files` in this base checkout; the six existing programs and the two
additions are the eight rows below. The subdirectory experiments in `regex/` and
`unwinding/` use their own harnesses and are outside this runner.

| benchmark | native time | native memory | node time | node memory | bun time | bun memory | native vs node | same answer |
|---|---|---|---|---|---|---|---|---|
| map_workload | 0.438 s | 3.3 MB | 0.150 s | 64.4 MB | 0.100 s | 89.4 MB | 2.93x | yes |
| nbody | 0.500 s | 3.8 MB | 0.941 s | 53.5 MB | 0.152 s | 56.8 MB | 0.53x | yes |
| sort | 0.300 s | 23.7 MB | 0.782 s | 188.0 MB | 0.758 s | 111.2 MB | 0.38x | yes |
| spectral_norm | 0.171 s | 4.9 MB | 0.307 s | 50.3 MB | 0.275 s | 58.1 MB | 0.56x | yes |
| string_build | 0.513 s | 5.9 MB | 0.471 s | 69.7 MB | 0.677 s | 96.1 MB | 1.09x | yes |
| tokenizer | 0.231 s | 66.2 MB | 0.320 s | 125.0 MB | 0.414 s | 108.7 MB | 0.72x | yes |
| trees | 2.416 s | 129.0 MB | 1.736 s | 263.8 MB | 1.333 s | 235.0 MB | 1.39x | yes |
| word_count | 0.217 s | 109.0 MB | 0.333 s | 139.8 MB | 0.502 s | 95.0 MB | 0.65x | yes |

| benchmark | allocations | frees | retains | releases | peak live | in regions |
|---|---:|---:|---:|---:|---:|---:|
| map_workload | 544,367 | 544,367 | 151,452 | 683,752 | 5,203 | 0 |
| nbody | 8 | 8 | 7,000,019 | 7,000,023 | 7 | 0 |
| sort | 6 | 6 | 2,000,002 | 2,000,010 | 4 | 0 |
| spectral_norm | 4 | 4 | 60 | 67 | 4 | 0 |
| string_build | 2,132,407 | 2,132,407 | 1,113,234 | 2,861,641 | 70 | 0 |
| tokenizer | 1,194,976 | 1,194,976 | 2,405,063 | 2,400,020 | 597,494 | 0 |
| trees | 68,332,244 | 1,572,900 | 67,982,686 | 67,982,728 | 2,097,149 | 66,759,344 |
| word_count | 1,020,129 | 1,020,129 | 6,100,372 | 5,100,429 | 1,020,006 | 0 |

The string program prints `13354836 units 1926777760`; the map program prints
`654186588`. Strings include BMP and supplementary characters. The map request mix is
20% insert/replace, 45% read (hits and misses), 20% update, 15% delete; final folds visit
both maps and the set in insertion order, so changing the order changes the answer.
All integer arithmetic in the rolling checksums stays below the exact-number limit.

### Native profiles

Both final `adamic build` binaries (clang `-O2`, no sanitizers or counters) were run under
Callgrind 3.24.0: `VALGRIND_LIB=/tmp/0aja3b1-valgrind/usr/libexec/valgrind /tmp/0aja3b1-valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=<profile> <binary>`.
The percentages below are self instruction shares, not wall-clock percentages; profiles
ran after timing and both printed their unchanged checksum.

**string_build:** 5,501,186,038 instructions; runtime UTF-16 lookup (`char_code` + `locate`) is 28.3%, index access/building (`usable`) 7.6%, surrogate joining 8.0%, append 3.2%.
The checksum algorithm's double remainder (`fmod`) is 15.3%; generated `main` is 10.7%, including inlined construction and checksum loops.
The main costs belong to runtime string handling and to the checksum algorithm as emitted.

**map_workload:** 3,457,359,803 instructions; runtime `map.c`'s `find` is 84.9% and `adamic_map_set` 4.8%, dominated by hash-table probing.
Source inspection and a standalone reproduction of its numeric hash formula put all initial IDs 0..2047 in one starting bucket at 4,096 or 8,192 buckets.
This supports runtime hash clustering as the cause; generated code and the checksum algorithm are minor here (`fmod` 1.7%).

Raw profiles: `/tmp/0aja3b1-string-final.callgrind` and
`/tmp/0aja3b1-map-final.callgrind`; annotated reports:
`/tmp/0aja3b1-string-final-annotate.log` and `/tmp/0aja3b1-map-final-annotate.log`.

### Validation and findings

- `bash cloud/setup.sh`: exit 0. Timing lines: Go ready 0s; clang ready 1s;
  Node ready 1s; submodules ready 1s; build cache warm 184s; done 184s on 5 processors.
  Sourced `/workspace/adamic-tools/env.sh` after setup; its configured tools directory
  is `/workspace/adamic-tools`, rather than `/opt/adamic-tools`.
- `go vet ./bench` and `go test -count=1 -timeout 30m ./bench`: exit 0; bench has no test files.
- `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -v ./internal/oracle -run '^(TestNativeAgreesWithNode|TestTheOracleCatchesOneByte)$/internal/oracle/testdata/(maps_and_text|sets|string_append|shared_slices|library_string_existing)\.a$'`:
  PASS, 36.602s, five differential fixtures plus the oracle's one-byte mutant;
  native cache hits 0, misses 16; Node hits 0, misses 11.
- Both new programs were built with `adamic build --sanitize` and run with
  `ASAN_OPTIONS=detect_leaks=1:halt_on_error=1` and `UBSAN_OPTIONS=halt_on_error=1`.
  They exit 0 with empty stderr and match source on Node and Bun and generated JavaScript
  through `node oracle/node.mjs <generated.mjs>`.
- No naturally written construct was refused or reported `NotYet` by Adamic. Node treats
  `.a` as untyped JavaScript, which required the runner's temporary source copy. The pinned
  cohere CLI also rejects `.a` paths as outside the program despite `sourceExtensions` in
  the root tsconfig. Identical scratch `.ts` copies were checked with the root compiler
  options, prelude and `CohereSettings.json`; both pass types, lint and formatting.
  Lint fixes were a suffixed interface name, local checksum accumulators instead of
  reassigning parameters, `cached ?? 10001`, and the formatter's output.
- Callgrind 3.24.0 was extracted from Debian's official `valgrind_3.24.0-3_amd64.deb` into
  `/tmp/0aja3b1-valgrind`. `sudo` is absent and `apt-get update` lacks permission to write
  `/var/lib/apt/lists`; the local extraction needs neither root nor a repository change.
- The full repository gate was not run. This bench-only unit used the touched package,
  the uncached oracle subset, new-program sanitizer comparisons and cohere checks above.
  No compiler or runtime files were changed; hash-map and string optimization fixes remain
  outside this unit's territory. Other machines and runtime versions were not measured.

### Mutants

Scratch copies of `bench/run.go` were tested with actual native and Node outputs.
The baseline passes; each changed output below makes its `sameAnswer` comparison report
`DIFFERS`. The generated-C mutants all compile with `-Werror`, exit 0 and emit no
ASan, UBSan or leak report, so only the output comparison kills them.

| Mutant | What caught it |
|---|---|
| Zero a runtime string's high-surrogate units during hashing | checksum changes to `2015033212`, with the same unit count |
| Join runtime-built pieces without separators | `12598836 units 934365399`, differing from Node |
| Reverse map/set iteration, keeping every live entry | checksum changes to `766933908` |
| Drop `.a` sources from discovery | scratch registration test expects two programs and finds one |
| Assign a runtime-built string to the numeric checksum local | cohere types-only check: TS2322, string not assignable to number |
| Existing oracle one-byte mutation | `TestTheOracleCatchesOneByte` passes by observing stdout disagreement |

Logs and generated mutant code are in `/tmp/0aja3b1-checks/`,
`/tmp/0aja3b1-bench-final.log`, `/tmp/0aja3b1-mutants-final.log`,
`/tmp/0aja3b1-oracle.log` and `/tmp/0aja3b1-cohere-final.log` on this machine.
