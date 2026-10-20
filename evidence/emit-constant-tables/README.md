# Evidence: emit/constant-tables-main 655516d4 (#nda35wb)

These are the review records for the change. None of this file is part of the change.

## The change

An array literal of 64 or more constant numbers, or of array literals of them, is emitted as a
static C table plus a loop. It used to be one adamic_array_push per element.

## Emitted C (adamic c, before -> after)

| program | lines | main's lines | largest function |
|---|---|---|---|
| tsprinter | 59,096 -> 40,765 | 19,887 -> 541 | 1,693 |
| json | 25,248 -> 6,511 | 19,857 -> 240 | 848 |
| yaml | 47,565 -> 35,855 | 12,505 -> 224 | 843 |
| graphql/printer | 25,134 -> 13,424 | 12,534 -> 253 | 988 |

## clang -O2

Performance's clock is on evidence/compiler/emit-split-clock e07f4c9f. It measured 4 programs, on
one 4-core Linux VM with clang 20.1.8, taking the median of 3 interleaved runs. At -O2 they went
from 58.1 s to 12.1 s, and peak memory roughly halved.

## Checks

- `check 655516d4a376 green · gofmt vet gate plan phases census products tests(12) · 31m 31s on Golem`.
  tests.log: `ok internal/native 799.570s`, `ok internal/oracle 1253.467s` (ADAMIC_GATE_UNCACHED=1).
- An earlier all-packages check of the same sha, `check 655516d4a376 red · products, tests · 2h 33m 12s`,
  timed out: products ran past 60 minutes and tests past 90. Its stage1 reds read "settle ... owners
  first" or "shared setup did not complete", because their oracle-owner products never built.
  internal/native and internal/oracle passed in that run too.
- The counts row for typed_arrays_workers_stats.a was regenerated on Linux with clang 20.1.8 by
  cloud worker session_01Bp8Vufofy3T7v2sQheyBAJ. Retains went from 82 to 67 and releases from 584
  to 569. Allocations and frees stayed at 513.

## Mutants (TestConstantArraysAgreeWithNode)

Each must be seen, against Node, under the sanitizers:
- the flat loop one short
- every table read at index 0
- a row the outer array retains instead of taking (LeakSanitizer, Linux)
- each table's first entry written as 12

They passed as part of internal/native in both checks above.
