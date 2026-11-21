# Benchmarks

Nine programs, the same source run three ways: Adamic's native binary (`adamic build`, clang `-O2`, no sanitizers), Node, and Bun when it's installed. Each program constructs deterministic input and prints a checksum, so there is no input-file I/O in the timing and every runtime has to give the same answer.

```
go run ./bench                      # every program, 5 interleaved rounds, best of 5
go run ./bench -rounds 9 -only trees,nbody -timeout 300s
go run ./bench -only parallel_files -threads 1,2,4,8,16 -rounds 5
```

`-threads` interleaves native thread settings too; Node and Bun use the independent sequential `parallelMap` shim. Parallel counts use one thread so peak liveness is reproducible.

The runs are interleaved round by round to spread changes in machine load across runtimes, and the best round is reported. Time is wall clock; memory is the process's peak resident set. The counted table comes from `adamic build --count`, run once, untimed: it reports allocations and reference-count calls alongside the timings. The machine, the versions and the load before and after print with the numbers, and numbers without them aren't worth quoting.

On Linux every program runs through [internal/benchmark/reaper/reaper.c](../internal/benchmark/reaper/reaper.c), a small C process bench builds with clang through internal/benchmark, which forks the program, waits for it, and hands bench its user time and peak resident set. Started directly, a program's peak was never its own: Go starts a child with `clone(CLONE_VFORK|CLONE_VM)`, and Linux's exec folds the starting memory's high-water resident set into the program's, so every program read at least bench's own (a bench holding 250 MB of inputs read 1321.4 MB for a C program that only prints, J-wme5jp2). Through the reaper, all that can carry over is the reaper's few hundred kilobytes. internal/benchmark's `TestPeakIsTheProgramsOwn` proves both halves. The reaper costs every timed run one more fork, a fraction of a millisecond: a paired comparison carries it on both sides alike, but an absolute wall time on a very short program carries it too. Every peak in the record below was read through it.

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
| `primes_parallel.a` | 1,000 prime-counting tasks, each over its own range of 10,000 numbers, through `parallelMap` across the pool's threads (#sp4m5ex) |

## Suites and comparing two commits

[suites.json](suites.json) names four sets, each with the rule that decides what belongs in it:

- `everyday`: the eight programs above but `primes_parallel.a`, whole-program workloads any mind can run in about a minute.
- `regression`: everyday plus every program a measured speed finding named, each with its finding's task id: `graph.ts` and `copy.ts` (#ahnc64w, #kha2ea0; their trailing timer is the point), `timer_burst.a` (#ezae5s0), `timer_dispatch.a` (#a6yb7d1), parallel_files at 1, 4 and 16 threads, and the record cost harness, which runs outside run.go. A program leaves only by Runtime's ruling on #re2xhwa, logged in `retired` with why.
- `holdout`: the Benchmarks Game programs everyday doesn't cover, ported in `holdout/`, run only at a release or before a claim, never tuned on.
- `claim`: the 1.0.0 claim's workload, native tsc's full check of src/compiler, once tsc builds natively.

```
go run ./bench -compare <baseline sha> <candidate sha> -suite everyday -samples 15
go run ./bench -compare <sha> <sha> -suite everyday -only trees -plant candidate -plant-ms 95
```

`-compare` builds `cmd/adamic` at each commit in a detached worktree, builds every program with both from this checkout's source, and makes eight runs of each program per sample: Node, baseline, Node, candidate, Node, candidate, Node, baseline, with the sides swapped on odd samples. Each runner's runs in a sample are combined as their mean. A program is judged on user time unless the suite says `wall` (parallel_files, whose user time rises with threads by design).

Each row gives the median of the per-sample ratios, candidate over baseline, and the sign test's 95% interval: the k-th smallest and k-th largest ratio (the 4th and 12th of 15), which assumes nothing about how times are distributed. `faster` means the whole interval is below 1.00, `slower` that it's all above, and anything else reads `no change at this confidence, none larger than` the interval's farther bound from 1.00: the change that run couldn't see, not proof of none. Fewer than six samples are refused. A program whose answer differs from Node's at either commit, or with a reading run.go would refuse, gets no verdict. Native against Node is a paired ratio with its own interval, and a program is listed as a loss or a win only when that interval excludes 1.00; the rest are listed as within the noise.

`-plant candidate` appends a fixed busy loop, `-plant-ms` long (25 by default), to a copy of each candidate program, so an A/A comparison must read `slower` once the plant is larger than the program's sensitivity: the proof the tool can fail. An A/A run of eight programs without it should read `no change` throughout, though by the interval's own 3.5% miss rate about one run in four shows one false verdict somewhere. Both sides run from paths of the same length, and the header says whether each row's two binaries are byte-identical. On Golem, identical bytes read a change twice: an A/A read nbody a steady 3% slower (J-eewt3vf), and in M3 three of thirteen identical rows got a verdict, nbody slower at a median ratio of 1.032 (J-d1emr4h). Two causes stood, and both are cured. The order used to rotate, which kept who ran just before whom, so the candidate always ran right after the baseline; now every native run follows a Node run, so a cost carried over from the run before lands on both sides alike inside every sample, whether it's tied to the baseline as the run before or to the other side, and the baseline's positions (1 and 7) and the candidate's (3 and 5) mirror, so a steady drift through the sample lands alike too. That costs eight runs a program a sample, Node's four among them. And each side's binary kept its physical pages, and every run its environment, for the whole run, so cache aliasing and the stack's alignment were constants tied to a side; now every native binary gets fresh pages before each sample (copied and renamed over itself, at the same path), and every run gets `ADAMIC_BENCH_PAD`, a value of a random length from 0 to 4095, drawn from a seed the header prints and `-seed` repeats. A row whose two binaries are byte-identical is a control: it shows its ratio and interval with `control` in the verdict cell, never `faster` or `slower`, and a line above the table counts the controls and names any that would have claimed a change. A bias that remains shows there, and a verdict within the size a control reached is unproven on that run. Run an A/A beside any claim, on the box that makes it.

## Parallel files, October 6, 2026

`parallel_files.a` generates 4,096 files, tokenizes and summarizes them through pure functions, then merges in input order. Linux amd64, AMD EPYC 9V74, four quota-limited CPUs (`nproc` 5), clang 20.1.8, Node 24.19.0, Bun 1.3.14. Best of five interleaved rounds: native 1/2/4 threads **0.744/0.454/0.312 s**, Node **0.637 s**, Bun **0.866 s**. All answers match. Load before **0.11/3.16/6.23**, after **0.69/3.10/6.14**; tests had finished, but the longer load averages still include them.

[Every round](parallel_files.measurements.log) includes memory and counts. Node uses oracle/node.mjs for `.a` sources; Bun 1.3.14 executes `.a` directly. Both resolve the independent sequential shim; existing `.ts` programs remain unchanged.

## TSan fixture units

The source batch's reported time was 82 seconds. Its full 4,096-file program now runs as six independent top-level oracle tests: three attempts with `ADAMIC_THREADS=1` and three with the override unset. Each attempt still compares the full ordered answer with Node; no input or thread setting was removed. Every subprocess has a one-minute deadline.

On the piece 6 Linux worker (four quota-limited CPUs, `nproc` 5), the one-thread units took 11.02, 19.02 and 15.18 seconds; the default-pool units took 10.39, 10.22 and 13.81 seconds. These are whole test-unit timings, including build and oracle work, from `go test -count=1 -v -timeout 90s ./internal/oracle -run 'Async|Parallel|Concurrency'`. The 82-second figure is supplied historical evidence, not a before measurement on this worker; the split establishes bounded units, not a speedup.

## Golem, November 20, 2026

The record, until the next rerun. Golem, a house box: Intel Core i7-4770S, 8 logical CPUs, Linux 6.8.0-147, in a sealed Loom measure job alone on the box (J-vpa5442). Adamic `9233213075`, clang 20.1.8, Node 24.19.0; Bun isn't on Golem, so it has no column. The job waited until the one-minute load was under 0.5 and timed nothing otherwise: load **0.48/1.44/1.74** before and **1.59/2.62/2.29** after (1, 5, 15 minutes), with no process over 5% CPU at the start. `go run ./bench -rounds 5 -timeout 120s`, best of five interleaved rounds, all answers matching. Time is wall clock and user is CPU time in user mode summed over every thread; memory is each program's own peak resident set, read through the reaper. This table doesn't count instructions: run.go reads none yet, though Golem's counters can, which `-game` uses.

| Benchmark | native | Node | native vs Node |
|---|---|---|---|
| copy | 0.534 s, user 0.393 s, 48.0 MB | 0.406 s, user 0.498 s, 212.4 MB | 1.32x, slower |
| devirtualize | 0.484 s, user 0.482 s, 2.1 MB | 0.590 s, user 0.600 s, 76.3 MB | 0.82x |
| environment | 0.316 s, user 0.313 s, 2.1 MB | 0.134 s, user 0.143 s, 74.8 MB | 2.35x, slower |
| graph | 0.455 s, user 0.356 s, 116.0 MB | 0.408 s, user 0.493 s, 297.7 MB | 1.11x, slower |
| nbody | 0.198 s, user 0.196 s, 2.2 MB | 0.948 s, user 0.952 s, 73.3 MB | 0.21x |
| parallel_files | 0.311 s, user 1.658 s, 65.5 MB | 0.701 s, user 0.703 s, 246.6 MB | 0.44x |
| primes | 0.120 s, user 0.116 s, 4.1 MB | 0.177 s, user 0.189 s, 88.6 MB | 0.68x |
| primes_5m | 0.031 s, user 0.026 s, 6.6 MB | 0.111 s, user 0.111 s, 76.3 MB | 0.28x |
| sort | 0.226 s, user 0.215 s, 24.8 MB | 0.685 s, user 0.649 s, 207.7 MB | 0.33x |
| spectral_norm | 0.330 s, user 0.328 s, 2.2 MB | 0.407 s, user 0.415 s, 70.4 MB | 0.81x |
| timer_burst | 0.165 s, user 0.162 s, 5.8 MB | 0.341 s, user 0.353 s, 145.1 MB | 0.48x |
| timer_dispatch | 0.453 s, user 0.451 s, 2.4 MB | 0.540 s, user 0.545 s, 75.6 MB | 0.84x |
| tokenizer | 0.181 s, user 0.167 s, 33.1 MB | 0.227 s, user 0.231 s, 139.3 MB | 0.80x |
| trees | 1.608 s, user 1.301 s, 196.4 MB | 1.504 s, user 2.032 s, 342.1 MB | 1.07x, slower |
| word_count | 0.208 s, user 0.165 s, 110.4 MB | 0.288 s, user 0.270 s, 152.2 MB | 0.72x |

| Benchmark | allocations | frees | retains | releases | peak live | in regions |
|---|---:|---:|---:|---:|---:|---:|
| copy | 10,200,070 | 200,070 | 400,027 | 400,035 | 400,002 | 10,000,000 |
| devirtualize | 772 | 771 | 102,400,513 | 102,400,517 | 772 | 1 |
| environment | 5,000,001 | 5,000,001 | 31,000,000 | 11,000,001 | 5 | 0 |
| graph | 10,485,818 | 68 | 27 | 31 | 1,048,575 | 10,485,750 |
| nbody | 8 | 8 | 10 | 14 | 7 | 0 |
| parallel_files | 3,952,651 | 3,952,651 | 8,151,045 | 8,654,930 | 8,839 | 0 |
| primes | 12 | 12 | 0 | 12 | 2 | 0 |
| primes_5m | 3 | 3 | 0 | 3 | 2 | 0 |
| sort | 6 | 6 | 2,000,002 | 2,000,010 | 4 | 0 |
| spectral_norm | 4 | 4 | 0 | 7 | 4 | 0 |
| timer_burst | 800,045 | 800,045 | 1,600,080 | 2,400,127 | 5 | 0 |
| timer_dispatch | 2,409 | 2,409 | 4,806 | 7,217 | 6 | 0 |
| tokenizer | 526,266 | 526,266 | 3,298,050 | 2,624,293 | 151,005 | 0 |
| trees | 68,332,244 | 1,572,900 | 0 | 42 | 2,097,149 | 66,759,344 |
| word_count | 1,020,129 | 1,020,129 | 6,100,370 | 5,100,427 | 1,020,006 | 0 |

Native was faster than Node in eleven of fifteen rows, from 0.21x (nbody) to 0.84x (timer_dispatch), and slower in four: environment 2.35x, copy 1.32x, graph 1.11x and trees 1.07x. Native's peak resident memory was the lower in every row, Node's 1.4 times native's (word_count) to 36 times (devirtualize).

The copy and graph rows are from a later run, J-68bfbtq on the same Golem at `eef9fd9a` (load 0.49/2.11/1.83 before, 0.80/2.01/1.80 after, nothing over 5% CPU, `go run ./bench -only copy,graph -rounds 5`, answers matching). Both programs used to end with a 400 ms timer before reading their idle resident set, and every timed run, native's and Node's alike, carried that sleep: J-vpa5442 read copy at 0.932 s against 0.809 s (1.15x) and graph at 0.850 s against 0.777 s (1.09x), each 0.4 s longer than its work (#dqfvqff, J-h66sxf2). The wait now runs only when `ADAMIC_BENCH_QUIET=1` asks for the idle read, which bench never sets. Their counted rows moved with it: the timer's closure is gone.

Across the five rounds, native's times on a program spread 0.0% to 5.7%, except tokenizer at 14.9% (0.181 s to 0.208 s) and graph at 17.1% (0.455 s to 0.533 s, J-68bfbtq); Node's spread up to 18.9% (graph, the same run), and 12.9% on the floor. parallel_files used the pool's default threads: 0.311 s of wall bought 1.658 s of user time across them. trees spends about 0.3 s of its 1.608 s outside user mode, where the October run on Cloud spent about 0.9 s: 66.8 million of its 68.3 million allocations end in regions, and since region blocks became huge pages (#cgn2cwt, `10f8f6b4`) they fault far less: on Golem that took trees from 2.01 s to 1.64 s and its faults from 626,554 to 124,024 (J-6sv2t7d).

The tokenizer generator uses a fixed Park-Miller seed with exact integer arithmetic below 2^53. Its intended mix is about 3/8 identifiers, 1/4 numbers, 1/4 punctuation and 1/8 quoted strings, separated by spaces, tabs and newlines. `go test ./bench -run '^TestTokenizer' -timeout 60s` checks the scanned identifier share stays at least 25% and proves the check catches the former overflowing generator. This is a workload check, not a new performance measurement.
