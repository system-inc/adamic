# Benchmarks

Eight programs, the same source run three ways: Adamic's native binary (`adamic build`, clang `-O2`, no sanitizers), Node, and Bun when it's installed. Each program constructs deterministic input and prints a checksum, so there is no input-file I/O in the timing and every runtime has to give the same answer.

```
go run ./bench                      # every program, 5 interleaved rounds, best of 5
go run ./bench -rounds 9 -only trees,nbody -timeout 300s
go run ./bench -only parallel_files -threads 1,2,4,8,16 -rounds 5
```

`-threads` interleaves native thread settings too; Node and Bun use the independent sequential `parallelMap` shim. Parallel counts use one thread so peak liveness is reproducible.

The runs are interleaved round by round to spread changes in machine load across runtimes, and the best round is reported. Time is wall clock; memory is the process's peak resident set. The counted table comes from `adamic build --count`, run once, untimed: it reports allocations and reference-count calls alongside the timings. The machine, the versions and the load before and after print with the numbers, and numbers without them aren't worth quoting.

| Program | What it exercises |
|---|---|
| `nbody.ts` | floating point on the fields of five objects, 1,000,000 steps |
| `spectral_norm.ts` | floating point over number arrays read and written by index, n = 1,500 |
| `tokenizer.ts` | a lexer walking 600,000 generated pieces with `charCodeAt`: identifiers, numbers, quoted strings, punctuation and whitespace, including non-ASCII text |
| `word_count.ts` | `split`, then a `Map<string, number>` over a million words, then a sort |
| `sort.ts` | a million numbers sorted with a comparator, random and then nearly sorted |
| `parallel_files.a` | 4,096 deterministic in-memory files, pure tokenization and summaries in parallel, merged in order |
| `trees.ts` | binary trees: 68 million small objects allocated, walked and freed |
| `devirtualize.a` | repeated rule visits over syntax-shaped records, with method calls in the hot loop |

## Parallel files, October 6, 2026

`parallel_files.a` generates 4,096 files, tokenizes and summarizes them through pure functions, then merges in input order. Linux amd64, AMD EPYC 9V74, four quota-limited CPUs (`nproc` 5), clang 20.1.8, Node 24.19.0, Bun 1.3.14. Best of five interleaved rounds: native 1/2/4 threads **0.744/0.454/0.312 s**, Node **0.637 s**, Bun **0.866 s**. All answers match. Load before **0.11/3.16/6.23**, after **0.69/3.10/6.14**; tests had finished, but the longer load averages still include them.

[Every round](parallel_files.measurements.log) includes memory and counts. Node uses oracle/node.mjs for `.a` sources; Bun 1.3.14 executes `.a` directly. Both resolve the independent sequential shim; existing `.ts` programs remain unchanged.

## TSan fixture units

The source batch's reported time was 82 seconds. Its full 4,096-file program now runs as six independent top-level oracle tests: three attempts with `ADAMIC_THREADS=1` and three with the override unset. Each attempt still compares the full ordered answer with Node; no input or thread setting was removed. Every subprocess has a one-minute deadline.

On the piece 6 Linux worker (four quota-limited CPUs, `nproc` 5), the one-thread units took 11.02, 19.02 and 15.18 seconds; the default-pool units took 10.39, 10.22 and 13.81 seconds. These are whole test-unit timings, including build and oracle work, from `go test -count=1 -v -timeout 90s ./internal/oracle -run 'Async|Parallel|Concurrency'`. The 82-second figure is supplied historical evidence, not a before measurement on this worker; the split establishes bounded units, not a speedup.

## Cloud, October 22, 2026

The record, until the next rerun. Cloud, a house box: AMD Ryzen Threadripper 3970X, 64 logical CPUs, Linux 6.18.40.1 under WSL2, its Loom serve stood down and inactive before and after the run. Adamic `f9d746475`, clang 20.1.8, Node 24.19.0, Bun 1.2.20. Load **0.36/0.23/0.13** before and **0.96/0.50/0.24** after (1, 5, 15 minutes). `go run ./bench -rounds 5 -timeout 120s`, best of five interleaved rounds, all answers matching. Time is wall clock and user is CPU time in user mode summed over every thread, both from the operating system for the waited process; memory is peak resident set. Instructions are not counted: WSL2 can't read the hardware counters, so that column waits for a run on Adam or Eve.

| Benchmark | native | Node | Bun | native vs Node |
|---|---|---|---|---|
| devirtualize | 0.044 s, user 0.044 s, 4.1 MB | 0.128 s, user 0.117 s, 73.8 MB | 0.054 s, user 0.064 s, 48.7 MB | 0.34x |
| nbody | 0.200 s, user 0.200 s, 5.1 MB | 0.777 s, user 0.751 s, 71.7 MB | 0.142 s, user 0.144 s, 43.6 MB | 0.26x |
| parallel_files | 0.156 s, user 0.911 s, 57.6 MB | 0.698 s, user 0.634 s, 245.2 MB | 0.913 s, user 0.810 s, 220.3 MB | 0.22x |
| sort | 0.225 s, user 0.209 s, 24.4 MB | 0.815 s, user 0.743 s, 194.8 MB | 0.873 s, user 0.856 s, 89.8 MB | 0.28x |
| spectral_norm | 0.137 s, user 0.136 s, 5.1 MB | 0.232 s, user 0.224 s, 69.3 MB | 0.210 s, user 0.217 s, 45.0 MB | 0.59x |
| tokenizer | 0.150 s, user 0.130 s, 32.4 MB | 0.232 s, user 0.197 s, 137.4 MB | 0.434 s, user 0.454 s, 93.5 MB | 0.65x |
| trees | 2.459 s, user 1.515 s, 161.5 MB | 1.298 s, user 1.595 s, 326.7 MB | 0.895 s, user 0.934 s, 192.5 MB | 1.89x, slower |
| word_count | 0.175 s, user 0.122 s, 109.6 MB | 0.293 s, user 0.246 s, 159.1 MB | 0.604 s, user 0.588 s, 81.7 MB | 0.60x |

| Benchmark | allocations | frees | retains | releases | peak live | in regions |
|---|---:|---:|---:|---:|---:|---:|
| devirtualize | 772 | 771 | 12,800,513 | 12,800,517 | 772 | 1 |
| nbody | 8 | 8 | 10 | 14 | 7 | 0 |
| parallel_files | 3,952,651 | 3,952,651 | 8,151,045 | 8,654,930 | 8,839 | 0 |
| sort | 6 | 6 | 2,000,002 | 2,000,010 | 4 | 0 |
| spectral_norm | 4 | 4 | 0 | 7 | 4 | 0 |
| tokenizer | 526,266 | 526,266 | 3,298,050 | 2,624,293 | 151,005 | 0 |
| trees | 68,332,244 | 1,572,900 | 0 | 42 | 2,097,149 | 66,759,344 |
| word_count | 1,020,129 | 1,020,129 | 6,100,370 | 5,100,427 | 1,020,006 | 0 |

Native was faster than Node in seven of eight rows, from 0.22x (parallel_files) to 0.65x (tokenizer), and slower in one: trees, 1.89x. Against Bun it was faster in six and slower in two: nbody (0.200 s against 0.142 s, 1.41x) and trees (2.459 s against 0.895 s, 2.75x). Native's peak resident memory was the lowest in seven rows; in word_count Bun's was lower (81.7 MB against 109.6 MB). Node's peak was 1.5 times native's (word_count) to 18 times (devirtualize).

Across the five rounds, native's times on a program spread 1.5% to 6.8%, except trees at 38.8% (2.459 s to 3.413 s). Node's spread up to 24.4% (sort) and Bun's up to 74.1% (devirtualize, a 0.05-second program). parallel_files used the pool's default threads: 0.156 s of wall bought 0.911 s of user time across them. trees is the row where wall exceeds user by most, about 0.9 seconds of a single-threaded run spent outside user mode, and 66.8 million of its 68.3 million allocations end in regions; that gap is the next thing profiled, not a cause found by this table.

The tokenizer generator uses a fixed Park-Miller seed with exact integer arithmetic below 2^53. Its intended mix is about 3/8 identifiers, 1/4 numbers, 1/4 punctuation and 1/8 quoted strings, separated by spaces, tabs and newlines. `go test ./bench -run '^TestTokenizer' -timeout 60s` checks the scanned identifier share stays at least 25% and proves the check catches the former overflowing generator. This is a workload check, not a new performance measurement.
