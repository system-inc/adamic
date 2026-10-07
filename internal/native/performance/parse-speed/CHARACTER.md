# File decoder UTF-16 cache, October 7, 2026

Built: propagated input-string UTF-16 length/ASCII flags during the decoder's existing sizing pass.
Commits: follows 3196034's release fast path; runtime/input.c only, no emission/parser/scanner edits.
Commands and outputs: Node input control PASS; parse 6,570,598,637 to 6,504,806,296 Ir; 77-file batch8 and whole-tree parity PASS.
Mutant: cached units missing the plus-one fails read_files against Node stdout, with both programs exiting 0.
Not covered: full repository gate, eliminated bounds/index conversions, a per-file region, or either dated speed target.

## Exact builds and measurement

Every Ir number below uses clang 20.1.8, `-O2 -g`, sanitizers off,
`-ffp-contract=off`, `-fno-optimize-sibling-calls`, no LTO or counted-runtime
hooks. The same runtime-parse.c from the area merge is used. The measured after
runtime was first a scratch prototype; byte comparison verifies every .c/.h
in it is identical to the promoted production runtime. Exact measured command:

```sh
source /workspace/adamic-tools/env.sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I /workspace/scratch/parse-speed/unit-cache-runtime /workspace/scratch/parse-speed/runtime-parse.c /workspace/scratch/parse-speed/unit-cache-runtime/*.c -lm -o /workspace/scratch/parse-speed/unit-cache-parse > /tmp/parse-speed-unit-cache-clang.log 2>&1
VALGRIND_LIB=/workspace/scratch/parse-speed/valgrind/usr/libexec/valgrind /workspace/scratch/parse-speed/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/parse-speed/unit-cache.callgrind /workspace/scratch/parse-speed/unit-cache-parse --manifest /workspace/scratch/parse-speed/compiler.txt --count > /workspace/scratch/parse-speed/unit-cache.stdout 2> /workspace/scratch/parse-speed/unit-cache.stderr
```

| Same release flags and 77 files | Before | After | Saved |
| --- | ---: | ---: | ---: |
| Whole parse process | 6,570,598,637 | 6,504,806,296 | 65,792,341 |
| adamic_string_units self | 86,952,614 | 2,351,072 | 84,601,542 |
| input.c decode self | 225,738,054 | 244,549,623 | -18,811,569 |
| Inline charCodeAt body self | 565,054,688 | 565,055,120 | -432 |

The gain is **1.0013%** of the release-fixed process. Maintaining the count
adds work to the decoder but avoids a separate source-length counting pass.
The inline read body itself is unchanged. Go parse baseline remains
1,684,637,174 Ir (Go 1.27.1 optimized default build, sanitizers/race off,
GOMAXPROCS=1 GODEBUG=asyncpreemptoff=1 under Callgrind); native is about **3.8613x**.

## How a read reaches bytes

The runtime area's adamic_string_char_code_at already checks the cached units
against byte length as an ASCII flag, checks the JavaScript position range,
truncates an in-range fractional position via the size_t conversion, and loads
one unsigned byte. Long non-ASCII heap strings lazily build their checkpoint
index and complete UTF-16 view once; subsequent reads directly load one unit,
including the two halves of supplementary characters. The unit view covers the
whole string even when nearly all of it is ASCII. String slices locate byte
boundaries separately through the checkpoint/cursor index. Short strings below
64 bytes and stack pieces have a bounded walk rather than allocating an index.

The missing cache was file/argument UTF-8 decoding: it validates input twice
(to size and then write), but previously discarded the UTF-16 count, so the
first length/read request walked the new string again. The existing first
pass now accumulates one unit for each decoded point, two for supplementary
points, and caches units+1. Invalid UTF-8's replacement points count as one;
empty input caches 1. Decoding, BOM handling and WTF-8 representation are unchanged.
The decoder's second writing pass and all read bounds checks remain.

An isolated counter records actual inline branches, not guessed ASCII ratios:
72 source files are byte-ASCII and five contain non-ASCII. Counts exclude
performance-probe overhead; performance numbers above use uninstrumented code.

| charCodeAt path | Source before | Source after | Other strings, both |
| --- | ---: | ---: | ---: |
| Direct ASCII byte load | 12,998,699 | 12,998,771 | 19,033 |
| Direct cached UTF-16 view | 8,016,017 | 8,016,017 | 40 |
| Slow fallback, including out-of-range | 154 | 82 | 547 |

Across **21,034,490 charCodeAt reads** (source and other), the before inline body
is **26.863 Ir/read** on average, under the exact release flags above. This
includes range/flag tests and double-to-size_t and unit-to-double conversions;
it excludes generated Scanner.code and the first-read cache build/length work.
The after inline body has essentially the same cost. There is no repeated
UTF-8 scan for long-source character reads. The sizeable remaining string/index
bucket includes one-time index construction, length access and token substring
boundary location, rather than all being paid per read. Input decode is in its
own source-aware bucket; its static decode function is not the character decoder.

## Correctness evidence

The off-by-one mutant changes only `string->units = units + 1` to `units`.
The existing read_files input oracle catches it by stdout against Node, not a
compiler diagnostic or memory fault: both exit 0 with empty stderr; first
length is 24 instead of Node's 25, with the final code unit missing. The source
is restored in finally before the good controls/builds. The Node input control
covers all six input fixtures, including malformed UTF-8 sweep, supplementary
pairs, BOM, NUL, empty input, truncated sequences, argv and file errors.

Full batch8 release output: **11,442,907 bytes identical** to typescript-go on
all 77 files. Full AST: **44,766,682 identical bytes**, all 77 files against Go
and Node, including sanitized native. Complete uncached native/oracle checks,
ASan/UBSan and LeakSanitizer included, are in the accompanying logs. Sanitized
test builds use `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`
with the same strict C/warning/contraction/sibling flags; they are not instruction
measurements. No emission hook or field-write edit is introduced.
