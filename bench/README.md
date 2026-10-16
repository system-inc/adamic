# Benchmarks

Six programs, the same source run three ways: Adamic's native binary (`adamic build`, clang `-O2`, no sanitizers), Node, and Bun when it's installed. Each program constructs deterministic input and prints a checksum, so there is no input-file I/O in the timing and every runtime has to give the same answer.

```
go run ./bench                      # every program, 5 interleaved rounds, best of 5
go run ./bench -rounds 9 -only trees,nbody -timeout 300s
```

The runs are interleaved round by round to spread changes in machine load across runtimes, and the best round is reported. Time is wall clock; memory is the process's peak resident set. The counted table comes from `adamic build --count`, run once, untimed: it reports allocations and reference-count calls alongside the timings. The machine, the versions and the load before and after print with the numbers, and numbers without them aren't worth quoting.

| Program | What it exercises |
|---|---|
| `nbody.ts` | floating point on the fields of five objects, 1,000,000 steps |
| `spectral_norm.ts` | floating point over number arrays read and written by index, n = 1,500 |
| `tokenizer.ts` | a lexer walking 600,000 generated pieces with `charCodeAt`: identifiers, numbers, quoted strings, punctuation and whitespace, including non-ASCII text |
| `word_count.ts` | `split`, then a `Map<string, number>` over a million words, then a sort |
| `sort.ts` | a million numbers sorted with a comparator, random and then nearly sorted |
| `trees.ts` | binary trees: 68 million small objects allocated, walked and freed |

## Noisy cloud numbers, October 5, 2026

Both tables below preserve the October 5, 2026 measurements. They are historical, pending a quiet-machine rerun on a house box, and do not describe the current compiler. The tokenizer row also predates the input-generator repair: its old text was almost entirely numbers, so that row does not measure the current mixed-token workload.

Not the record: a shared cloud container, 4 vCPUs (Intel Xeon @ 2.80GHz), Linux 6.18, load 0.74 before and 1.04 after. Across the five rounds one runtime's times on one program spread 7% to 40%, so differences under about 1.4x here mean little. Adamic `0955c44`, clang 18.1.3, Node 24.21.0, Bun 1.3.14.

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

Among the five programs native finished, it had the lowest peak resident memory; Node's peak was about 2.0 to 14 times native's (2.0 for word_count). Native's elapsed time was lower in two rows and higher in three; tokenizer did not finish within 120 seconds. The investigation at that revision found these costs in the generated C and runtime:

- **tokenizer: the string walk.** `charCodeAt(i)` and `length` decoded UTF-8 from the start on every call, making the walk quadratic. The degenerate input also left the identifier path almost unused. See docs/memory.md for the subsequent string-index work.
- **trees: allocation and retain and release traffic.** The counted row recorded about 68 million allocations and frees, and about 375 million retains. Regions and ownership optimizations were subsequent work; this table does not measure them.
- **spectral_norm: array reads through a call.** `vector[j]` called `adamic_array_at` in a separate runtime translation unit and checked a double-valued index. Inlining and proven integer counters were possible improvements, not results measured by this table.
- **nbody: field access through a call, and reference-count traffic.** Field access used `adamic_object_field` with a slot cache, and the counted row recorded about 52 million retains and releases. Bun's 0.201-second result alone does not establish which optimization produced it.

The two lower native elapsed times were sort and word_count. Those programs exercise numeric sorting and maps; the table does not isolate the contribution of either implementation or JIT warm-up.

The tokenizer generator uses a fixed Park-Miller seed with exact integer arithmetic below 2^53. Its intended mix is about 3/8 identifiers, 1/4 numbers, 1/4 punctuation and 1/8 quoted strings, separated by spaces, tabs and newlines. `go test ./bench -run '^TestTokenizer' -timeout 60s` checks the scanned identifier share stays at least 25% and proves the check catches the former overflowing generator. This is a workload check, not a new performance measurement.
