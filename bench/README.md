# Benchmarks

Seven fair programs, the same source run three ways: Adamic's native binary (`adamic build`, clang `-O2`, no sanitizers), Node, and Bun when it's installed. Each program makes its own input from a fixed seed and prints a checksum, so there's no I/O in the timing and every runtime has to give the same answer.

```
go run ./bench                      # every program, 5 interleaved rounds, best of 5
go run ./bench -rounds 9 -only trees,nbody -timeout 300s
go run ./bench -only parallel_files -threads 1,2,4,8,16 -rounds 5
```

`-threads` interleaves native thread settings too; Node and Bun use the independent sequential `parallelMap` shim. Parallel counts use one thread so peak liveness is reproducible.

The runs are interleaved round by round, so a machine that slows down slows all three alike, and the best round is reported. Time is wall clock; memory is the process's peak resident set. The counted table comes from `adamic build --count`, run once, untimed: it's what explains a native time. The machine, the versions and the load before and after print with the numbers, and numbers without them aren't worth quoting.

| Program | What it exercises |
|---|---|
| `nbody.ts` | floating point on the fields of five objects, 1,000,000 steps |
| `spectral_norm.ts` | floating point over number arrays read and written by index, n = 1,500 |
| `tokenizer.ts` | a lexer walking a 3.5-million-code-unit text with `charCodeAt`, some of it non-ASCII |
| `word_count.ts` | `split`, then a `Map<string, number>` over a million words, then a sort |
| `sort.ts` | a million numbers sorted with a comparator, random and then nearly sorted |
| `parallel_files.a` | 4,096 deterministic in-memory files, pure tokenization and summaries in parallel, merged in order |
| `trees.ts` | binary trees: 68 million small objects allocated, walked and freed |

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
