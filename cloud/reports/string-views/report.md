Measured area/runtime request and 77-file parse baselines before runtime fixes.
Base: 8cb9d252a9dc79d563b2644e0368e8b2edb114c7, descendant of 4d86c305.
Commands and outputs: native best 9,257.69 requests/s; parse 6,503,630,226 Ir; allocations 1,575.00123/request.
Mutants: no baseline measurement mutant claimed yet; runtime validation is recorded in the follow-up.
Not covered yet: step 2/3 measurements, final ownership mutants and full gate.

## Baseline

Read CLAUDE.md and both requested reports first. The request evidence at a4e0902 profiles f4ec96c, not this base. The 10.2% parser bucket was corrected later: input.c decode belongs to input, not indexing. Historical rates and instructions are not this machine or this base.

This container reports Intel Xeon Platinum 8370C at 2.80GHz, nproc 5, cpu.max 400000 100000. Setup: Go/clang/Node ready 0s, submodules 1s, cache warm/done 204s. Source /workspace/adamic-tools/env.sh in every shell. Go 1.27.1, clang 20.1.8, Node 24.19.0, Valgrind 3.24.0 extracted from official Debian package into scratch.

The exact six-route corpus is 100,000 requests, 417,755,377 bytes, SHA-256 cba64bd86fdd84d7086973a145a4e8419dd31937470728efe8bf8db156a75c39. Checksum 7,394,547 UTF-16 units. Reused a4e0902 command.a, generator and timing markers. The native-only adaptation removes the Wasm engine from measure.mjs, preserving warmup, checksum, serial Node/native ordering and five rounds. All rounds are in requests-before.json. Scratch native C is unchanged between runtime builds.

Native release flags: -O2, -ffp-contract=off, -fno-optimize-sibling-calls, no LTO, no sanitizers, no ADAMIC_COUNT. Parse adds -g. Timings exclude startup, input splitting, warmup and teardown, using separate stderr serve markers as in the evidence harness. Baseline rounds 1-3 overlapped benchmark preparation builds; only rounds 4-5 were free of our other CPU work. Their native best is 8,652.71 requests/s. Overall best 9,257.69 and Node 36,049.47 are observations, not a speedup claim. Host load and rate variation are large.

## Per-function baseline

PC self time is a separate run: 23.045802 CPU seconds, 5,571 samples. Unresolved samples stay in the denominator. Parse self instructions collapse inline records and reconcile exactly to the whole-process summary. At and indexOf are not used by this parse driver.

| Function | Request self ms | Request self share | Parse self Ir |
|---|---:|---:|---:|
| adamic_string_slice | 2763.35 | 11.991% | 64,821,007 |
| adamic_string_locate | 2473.77 | 10.734% | 68,686,255 |
| adamic_string_char_code | 885.26 | 3.841% | 28,731 |
| adamic_string_at | 905.95 | 3.931% | 0 |
| adamic_string_index_of_at | 20.68 | 0.090% | 0 |
| adamic_string_share | 959.72 | 4.164% | 21,968,067 |
| adamic_allocate | 1385.81 | 6.013% | 135,951,493 |
| adamic_string_units | 368.17 | 1.598% | 86,952,614 |

Flagged ASCII charCodeAt already reads bytes inline. Locate returns a byte offset directly, but slice calls it twice and at routes through slice. indexOf walks the needle to test surrogate halves and checks UTF-8 sequence widths while advancing, even with a flagged ASCII haystack. No flag path is asymptotically quadratic from UTF-16 translation; the remaining issue is decoding/translation calls and allocation/copy overhead. Search retains its existing worst-case O(haystack times needle) byte comparisons; for a fixed needle it is linear.

## Allocations

A scratch runtime increments disjoint counters in share allocation paths, concat, from_number and grown append, and every allocation by heap kind. Run minus control removes identical read/split/setup. Slices includes character indexing and byte slices; whole-string retains are excluded. Empty slices and other string builders stay in other_strings. These are heap values, not backing-buffer mallocs. Sum is exactly 157,500,123.

| Category | Serve allocations | Per request |
|---|---:|---:|
| slices | 128,460,190 | 1284.60190 |
| concatenation | 7,847,806 | 78.47806 |
| number_formatting | 62,601 | 0.62601 |
| growing_append | 10,069,705 | 100.69705 |
| objects | 3,981,972 | 39.81972 |
| arrays | 3,788,122 | 37.88122 |
| maps | 2,974,371 | 29.74371 |
| cells | 20,000 | 0.20000 |
| closures | 20,000 | 0.20000 |
| map_iterators | 20,000 | 0.20000 |
| other_strings | 255,356 | 2.55356 |

## Parse reproduction

Fetched codex/stage1-lint-batch8 at 4189abd3490757e8abe13722ceb365c451293e92. Used parse-speed prepare.py with only its repository root adjusted for scratch. TypeScript v6.0.3 is pinned to 050880ce59e30b356b686bd3144efe24f875ebc8; exactly 77 compiler files. Driver removes only visit(context, root), retaining line table, parent map and cleanup. Current area/runtime parser/scanner and compiler generate parse.c. Whole-process output is 0.

```sh
python3 /workspace/scratch/string-views/internal/native/performance/parse-speed/prepare.py /workspace/scratch/string-views/parse /workspace/scratch/string-views/typescript > /tmp/string-views-prepare.log 2>&1
source /workspace/adamic-tools/env.sh
/workspace/scratch/string-views/adamic c /workspace/scratch/string-views/parse/batch8/parse.a > /workspace/scratch/string-views/parse/parse.c 2> /tmp/string-views-emit.log
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I internal/native/runtime /workspace/scratch/string-views/parse/parse.c internal/native/runtime/*.c -lm -o /workspace/scratch/string-views/parse/native-before > /tmp/string-views-parse-build.log 2>&1
VALGRIND_LIB=/workspace/scratch/string-views/valgrind/usr/libexec/valgrind /workspace/scratch/string-views/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/string-views/parse/before.callgrind /workspace/scratch/string-views/parse/native-before --manifest /workspace/scratch/string-views/parse/compiler.txt --count > /tmp/string-views-parse-before.stdout 2> /tmp/string-views-parse-before.stderr
```

The preserved brk segment overflow warning is nonfatal; Callgrind finishes and reports 6,503,630,226 Ir. Raw profile and reconciled named self costs are beside this report. Native sampling reused a4e0902 diagnostic.c through LD_PRELOAD, enabled only between serve markers. Profiles do not supply unprofiled timing components.
