# JSON formatter performance follow-up

Baseline source: `adb4e0aa911b010c038c4925ec0a5c949f313acb`. Optimized source
is the commit containing this report. Only `stage1/cohere/json/` changed.
The model was `stage1/typescript/scanner/PERFORMANCE.md` from
`origin/codex/typescript-scanner`.

Native improved from **59.43 to 94.21 texts/s**, **1.59 times** on the same
full corpus. Node improved from 88.27 to 148.72; Go measured 109.37. Optimized
native remains 1.58 times slower than Node and 1.16 times slower than Go.
The original 62.28 / 91.17 / 109.28 baseline in GAPS.md came from the earlier
three rounds; the fresh baseline below is the controlled comparison here.

| Side | Seconds, five rounds | Median texts/s |
| --- | --- | ---: |
| baseline-native | 17.733 / 18.126 / 17.924 / 18.924 / 18.088 | 59.43 |
| native | 11.410 / 11.453 / 11.173 / 11.200 / 11.812 | 94.21 |
| baseline-Node | 12.178 / 12.017 / 12.331 / 11.970 / 12.957 | 88.27 |
| Node | 7.093 / 6.910 / 7.229 / 7.282 / 7.436 | 148.72 |
| Go | 9.665 / 9.785 / 9.982 / 9.829 / 9.849 | 109.37 |

## Method and scope

Go cohere remains the byte-for-byte contract, including all syntax refusals.
The corpus remains 1,075 texts: 1,041 checkout JSON files and 34 generated
inputs. All 1,040 original repository files, the numeric-separator proving
fixture, and all generated cases remain included. Prettier 3.9.6 is a separate
report with exactly the same nine known upstream differences.

The timing script runs five fresh-process, interleaved rounds of baseline
native, optimized native, baseline Node source, optimized Node source and Go.
Each reads the same 60,612,868-byte escaped input and writes the same
62,400,611-byte answer stream. Every timed output is checked byte for byte
against Go, with successful exit and empty stderr. Startup, file reading,
formatting/refusal and protocol output are included; compilation and comparison
are excluded. Refusals count as texts. Go buffers its output; Node/native use
console.log. These are driver measurements, not an isolated formatter loop.
No test or profiling job ran concurrently with these five rounds.

Machine: AMD EPYC 9V74, `nproc` 5, cgroup quota 4 CPUs, 17.6 GB RAM.
Go 1.27.1, clang 20.1.8, Node 24.19.0; the existing toolchain was reused.
Original setup timings remain in GAPS.md: Go 0s, clang 0s, Node 0s,
submodules 0s, build cache 13s, total 13s.

Callgrind 3.24.0 was absent. `apt-get update` failed with permission denied
opening `/etc/apt/apt.conf.d/80-applied-apt-retries` and creating
`/var/lib/apt/lists/partial`. Downloading Debian's
`valgrind_3.24.0-3_amd64.deb` and extracting it into scratch succeeded;
`VALGRIND_LIB` points at its `usr/libexec/valgrind`.
An initial whole-corpus instrumented run was stopped because it was too slow;
its incomplete profile is not used. Both completed profiles use an identical,
fixed, bounded sample: every generated case, every 16th corpus index with at
most 131,072 UTF-8 source bytes, and the largest other text within that limit.
There are **97 texts, 340,876 source bytes**. The committed manifest names each
one. Sample input SHA256:
`48cc2ac1454143609af5bf87a5f2b1d9a8b9c65957636dcc095d312fe0b11d18`.
The sample deliberately includes deep nesting and the long-array probes;
it excludes the largest multi-megabyte fixtures. Its instruction shares are
not claimed to be the full-corpus shares or extrapolated to full-corpus Ir.

Profiles use stage 0's generated C and uninstrumented release code-generation
flags (`-O2`, `-ffp-contract=off`, `-fno-optimize-sibling-calls`) plus `-g`.
Both emitted a nonfatal Valgrind brk-segment warning and completed with exact Go
sample bytes. Both profiled binaries additionally passed the entire Go corpus,
62,400,611 answer bytes each. Separate `ADAMIC_COUNT` builds measured ownership
operations across the entire corpus and produced the exact same Go bytes.
The final lint-required template literal emits **identical C** to the profiled
indentation expression; `cmp final.c frozen.c` passed.

Raw profiles are [baseline.callgrind.gz](performance/baseline.callgrind.gz) and
[final.callgrind.gz](performance/final.callgrind.gz). Named top-twenty logs,
tool diagnostics, manifest, allocation counters, comparison/mutant logs and
all five timing samples are beside them. `profile.py` collapses same-name
DWARF inline records and requires all self costs, including anonymous names,
to sum exactly to the Callgrind summary. Anonymous addresses and `(below main)`
are excluded from the named ranking. Callgrind's recursive context suffix
`'2` is retained; inclusive rows overlap and must never be added. Generated
`adamic_function_N_` prefixes are omitted in the tables for readability.

## Top 20 inclusive instructions

Baseline:

| Rank | Function | Ir | Share |
| --- | --- | ---: | ---: |
| 1 | `main` | 1,554,972,940 | 99.99% |
| 2 | `format` | 1,225,573,365 | 78.81% |
| 3 | `Reader_value` | 376,279,126 | 24.20% |
| 4 | `Reader_value'2` | 314,575,936 | 20.23% |
| 5 | `adamic_release` | 293,673,055 | 18.88% |
| 6 | `escaped` | 262,744,042 | 16.90% |
| 7 | `Printer_print` | 254,781,616 | 16.38% |
| 8 | `adamic_array_join` | 233,254,763 | 15.00% |
| 9 | `Printer_print'2` | 206,512,618 | 13.28% |
| 10 | `adamic_string_split` | 199,299,515 | 12.82% |
| 11 | `Reader_peek` | 198,979,378 | 12.79% |
| 12 | `adamic_string_slice` | 177,888,672 | 11.44% |
| 13 | `Documents_fits` | 165,806,266 | 10.66% |
| 14 | `Reader_skip` | 165,239,367 | 10.63% |
| 15 | `adamic_string_concat` | 142,192,846 | 9.14% |
| 16 | `adamic_string_join_halves` | 125,243,675 | 8.05% |
| 17 | `adamic_string_equal` | 84,529,523 | 5.44% |
| 18 | `Documents_add` | 83,249,566 | 5.35% |
| 19 | `Reader_number` | 76,659,182 | 4.93% |
| 20 | `adamic_allocate` | 76,015,438 | 4.89% |

Optimized:

| Rank | Function | Ir | Share |
| --- | --- | ---: | ---: |
| 1 | `main` | 909,524,994 | 99.98% |
| 2 | `format` | 646,989,632 | 71.12% |
| 3 | `Printer_print` | 199,016,067 | 21.88% |
| 4 | `escaped` | 193,318,283 | 21.25% |
| 5 | `adamic_release` | 165,808,456 | 18.23% |
| 6 | `Reader_value` | 161,606,804 | 17.76% |
| 7 | `Reader_value'2` | 152,805,487 | 16.80% |
| 8 | `Printer_print'2` | 149,941,727 | 16.48% |
| 9 | `adamic_string_index_of` | 136,600,813 | 15.02% |
| 10 | `adamic_string_index_of_at` | 136,082,150 | 14.96% |
| 11 | `Documents_fits` | 101,853,393 | 11.20% |
| 12 | `Documents_add` | 88,843,272 | 9.77% |
| 13 | `adamic_string_split` | 70,001,725 | 7.70% |
| 14 | `adamic_array_join` | 64,130,201 | 7.05% |
| 15 | `adamic_string_concat` | 54,022,325 | 5.94% |
| 16 | `adamic_string_equal` | 52,822,790 | 5.81% |
| 17 | `adamic_string_slice` | 50,182,181 | 5.52% |
| 18 | `adamic_string_join_halves` | 48,445,952 | 5.33% |
| 19 | `adamic_allocate` | 39,890,094 | 4.38% |
| 20 | `Printer_blankAfter` | 39,689,261 | 4.36% |

## Top 20 self instructions

Baseline:

| Rank | Function | Ir | Share |
| --- | --- | ---: | ---: |
| 1 | `adamic_release` | 255,418,348 | 16.42% |
| 2 | `adamic_string_split` | 127,402,159 | 8.19% |
| 3 | `adamic_string_join_halves` | 125,243,675 | 8.05% |
| 4 | `adamic_string_slice` | 75,906,100 | 4.88% |
| 5 | `adamic_allocate` | 72,918,049 | 4.69% |
| 6 | `stringWidth` | 58,748,278 | 3.78% |
| 7 | `adamic_array_join` | 57,871,666 | 3.72% |
| 8 | `adamic_string_equal` | 54,674,629 | 3.52% |
| 9 | `format` | 44,695,657 | 2.87% |
| 10 | `Reader_value'2` | 39,759,492 | 2.56% |
| 11 | `adamic_retain` | 38,527,550 | 2.48% |
| 12 | `adamic_array_push` | 38,043,720 | 2.45% |
| 13 | `Documents_fits` | 35,699,522 | 2.30% |
| 14 | `adamic_string_locate` | 31,527,704 | 2.03% |
| 15 | `Documents_get` | 31,485,604 | 2.02% |
| 16 | `Reader_skip` | 24,586,592 | 1.58% |
| 17 | `adamic_string_share` | 23,988,943 | 1.54% |
| 18 | `free` | 19,734,315 | 1.27% |
| 19 | `Documents_add` | 17,963,915 | 1.16% |
| 20 | `Printer_print'2` | 17,273,450 | 1.11% |

Optimized:

| Rank | Function | Ir | Share |
| --- | --- | ---: | ---: |
| 1 | `adamic_release` | 146,949,440 | 16.15% |
| 2 | `adamic_string_index_of_at` | 73,956,765 | 8.13% |
| 3 | `adamic_string_join_halves` | 48,445,952 | 5.33% |
| 4 | `adamic_string_split` | 43,937,909 | 4.83% |
| 5 | `Reader_value'2` | 43,844,749 | 4.82% |
| 6 | `adamic_allocate` | 37,905,044 | 4.17% |
| 7 | `adamic_string_equal` | 34,885,228 | 3.83% |
| 8 | `Documents_add` | 34,428,318 | 3.78% |
| 9 | `Documents_fits` | 26,761,265 | 2.94% |
| 10 | `Documents_get` | 25,097,408 | 2.76% |
| 11 | `format` | 24,356,221 | 2.68% |
| 12 | `adamic_retain` | 21,804,667 | 2.40% |
| 13 | `adamic_string_slice` | 20,952,890 | 2.30% |
| 14 | `adamic_string_units_before` | 20,308,907 | 2.23% |
| 15 | `Reader_skip` | 18,300,154 | 2.01% |
| 16 | `adamic_array_push` | 14,866,111 | 1.63% |
| 17 | `adamic_write_line` | 13,810,369 | 1.52% |
| 18 | `Printer_print'2` | 11,709,980 | 1.29% |
| 19 | `adamic_object_new` | 11,086,681 | 1.22% |
| 20 | `adamic_string_locate` | 9,500,949 | 1.04% |

## Attribution and changes in this port

The sample's total Ir fell from **1,555,146,045 to 909,698,061**, **41.50%**.
Its format entry inclusive cost fell from 1,225.6M to 647.0M; escaped-output
driver inclusive cost fell from 262.7M to 193.3M. This is a combined before/after
result; the changes were not individually profiled, so no individual gain is
inferred from these totals.

- Parser comment lookahead and punctuation use UTF-16 code units instead of
  allocating one- and two-character slices. Reader.skip inclusive cost fell
  from 165.2M to 21.9M and Reader.peek from 199.0M to 25.7M in this sample.
  The comment scanner, Unicode whitespace, EOF errors and unary spelling stay
  covered by Go's exact answers.
- Numeric values no longer convert to canonical number strings when they are
  not property keys. Canonical key comparison is retained. Separator removal
  avoids split/join when no underscore exists. Go's accepted malformed
  separator spellings and raw literal behavior remain unchanged.
- Comment-free inputs no longer allocate three comment arrays per AST node.
  The existing attachment and populated lists are retained for comment inputs.
  The printer returns a bare result when no comment wrapper is needed, instead
  of allocating a one-child concat document and parts array for every node.
- Immutable document text widths are computed once at construction and reused
  by fits and output. The ASCII width scan reads each code unit once. Width
  computation now inlines into Documents.add, so disappearance of its separate
  named row does not mean it costs zero. Indentation prefixes are cached once
  per print instead of allocating and joining a spaces array for each line.
- The escaped driver skips split/join for absent escape characters; unescaped
  strings with no backslash return directly. This improves driver work as well
  as formatter work; timings retain both. The remaining search scans are
  explicitly attributed below.

Disjoint named **self** instructions, not overlapping inclusive costs:

| Category | Baseline Ir | Baseline share | Optimized Ir | Optimized share |
| --- | ---: | ---: | ---: | ---: |
| Retains and releases | 293,945,898 | 18.90% | 168,754,107 | 18.55% |
| allocate / malloc / free / realloc | 110,043,067 | 7.08% | 52,104,955 | 5.73% |
| String building | 424,349,563 | 27.29% | 130,788,123 | 14.38% |
| UTF-16 views and access | 48,085,545 | 3.09% | 41,697,995 | 4.58% |
| Runtime number formatting | 1,430,787 | 0.092% | 1,276 | 0.00014% |

String building includes split, join-halves, array join, concat, slice, share,
append and the string-allocation wrapper. UTF-16 includes locate, units,
units-before, usable, unit_at, char-code and code-point wrappers. Runtime number
formatting includes number_format, from_number and shortest_digits. These
categories omit other functions and are not a partition of the entire process;
allocator implementation details and anonymous libc costs are not all isolated.
The port's own numberText baseline self/inclusive was 321,044 / 2,956,734;
ordinary number spelling is still held to Go. Sample number_format calls fell
from 12,532 to 11; this supports removing redundant port conversions, not a
proposal to rewrite dtoa.

Logical full-corpus counts, not bytes or malloc calls:

| Snapshot | Allocations | Frees | Retains | Releases | Peak live | Regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Baseline | 208,410,603 | 208,410,603 | 458,239,251 | 513,466,401 | 21,374,125 | 0 |
| Optimized | 103,275,657 | 103,275,657 | 302,444,925 | 301,716,973 | 13,759,774 | 0 |

Allocations fell **50.45%**. Retain/release hooks count even immortal constants
and undefined checks; small allocations use slabs. Equal frees and allocations
are observations from these counted runs, complemented by LeakSanitizer.

## Compiler and runtime proposals, without changes to internal/

These are candidates supported by measurements and source inspection. Their
possible gains have not been measured by implementing them.

1. Runtime ASCII search and includes: optimized index_of_at has **73.96M self
   and 136.08M inclusive Ir**, 8.13% and 14.96% of the sample. `string.c:520-553`
   scans each decoded byte sequence and calls memcmp; includes lowers through
   indexOf, which also computes a UTF-16 result that a boolean does not need.
   Investigate a one-byte ASCII needle fast path and a boolean contains path.
   Preserve pair/lone-surrogate searches, from-index semantics and UTF-16
   positions. The port's absent-escape guards reduce copying, but increase
   search work; index_of_at baseline self was 6.93M. This tradeoff is included
   in the measured total and wall time, not hidden as a runtime regression.
2. Compiler borrowed table reads and temporary commands: optimized retain and
   release self total **168.75M**, 18.55%. Documents.fits alone issues 530,553
   release calls in the sample, with 30.77M inclusive release-edge instructions.
   Generated C retains objects read from arrays and field references, allocates
   generic command objects, then releases them; fits self is 26.76M and get self
   25.10M. Investigate ownership-proven borrowed reads and specialized records
   or short-lived command storage. Port parallel primitive stacks are also a
   possible later experiment; this profile does not prove that compiler work
   is the only way to reduce that traffic. No checks or ownership guarantees
   should be removed without a proof and independent mutants.
3. Runtime joins and source slices: optimized join-halves self remains 48.45M,
   split 43.94M, slice 20.95M. `string_share.c` copies below 64 bytes or below a
   quarter of the owner, so many short lexemes allocate while Reader.text keeps
   the source alive. Investigate lifetime-anchored borrowed slices and join
   fast paths, preserving lone surrogates and source-retention bounds. Native
   still performs 103.3M logical allocations on the full corpus. The profile
   does not establish quadratic string append behavior.
4. Runtime UTF-16 views: units-before self rises from 3.84M to **20.31M** because
   the driver does more includes searches; locate falls from 31.53M to 9.50M.
   `string_index.c:143-175` already has an ASCII path and checkpoint/forward
   cursor. Investigate avoiding a position translation for boolean searches
   and cheaper nearby access on non-ASCII strings. Preserve JS UTF-16 semantics.
   The bounded sample's Unicode share is limited; no claim that this is the
   dominant full-corpus bottleneck is supported here.
5. Compiler proven field/index access: `adamic.h:170-176` inline field-cache
   lines account for 21.00M optimized self Ir (2.31%), and array access lines
   286-300 for 22.46M (2.47%). Generated class checks, entry stack checks and
   method bodies are not separately isolated. Investigate monomorphic field
   layouts and proven integer indexes, retaining required bounds and stack
   behavior. These measured shares do not support blaming the entire gap on
   checks. Number formatting is negligible after the port fix; no dtoa unit
   is warranted by this sample.

## Validation and mutants

The complete original JSON suite passed in **292.082s**, including all 1,075
Go/native-ASan/UBSan/LeakSanitizer/Node/emitted-JavaScript answers, 72 additional
boundaries, direct stdout drivers, three original port mutants, gap programs,
and the pinned external Prettier report. New snapshot and cache-mutant checks
ran separately with profiling snapshots enabled; both passed. The filtered
core oracle passed in **17.206s**, on the same eight string/collection/exception/
bitwise fixtures as GAPS.md. Vet and gofmt produced no output; cohere format
passed, and lint/type reported **276 rules, 7 checked, 100% Adamic-ready**.
`git diff --check` passed. All 25 controlled timing runs, both instrumented
sample streams and both full-corpus counted streams matched exact Go bytes.

The original three port mutants still compile and exit successfully with empty
stderr, and native/Node comparisons catch only their wrong bytes: removing the
final newline (four controls), removing colon space (three), and using ordinary
JSON for package.json (two). Additional performance checks:

- Cached text width becomes zero. The optimized code compiles and exits
  successfully on native release and Node; both byte comparisons catch a
  long string incorrectly kept on one line. `TestCachedWidthMutantIsCaught`
  passed; the mutant was caught.
- Callgrind summary increased by one. `profile.py` exits nonzero with
  `callgrind self costs do not sum to summary`. Actual instruction records
  are untouched; this proves the new accounting check rejects a bad total.

The profiled snapshot test passed on both entire-corpus binaries in 46.76s;
its additional cached-width mutant passed in 9.25s, 56.042s together. The
existing three stage 0 refusal proofs and separate Go-printer comparison
mutants also passed. No compiler or runtime files changed. No full repository
gate was rerun; this port-only follow-up used its complete package plus the
filtered oracle, as the unit allows. Existing API limits remain in GAPS.md.
The Callgrind sample is bounded; no Go or Node instruction/allocation profile,
custom formatter options, malformed-byte decoder or arbitrary input proof is
claimed.

## Reproduction

All test/tool output goes to files. The original npm pin and setup instructions
are in GAPS.md. Scratch paths below are the actual paths used here.

```sh
source /workspace/adamic-tools/env.sh
mkdir -p /tmp/adamic-json-perf /tmp/adamic-json-artifacts
ADAMIC_JSON_ARTIFACTS=/tmp/adamic-json-artifacts ADAMIC_JSON_PRETTIER=/tmp/adamic-json-prettier go test -v -count=1 -timeout 30m ./stage1/cohere/json > /tmp/adamic-json-perf/suite.log 2>&1
python3 stage1/cohere/json/profile.py /tmp/adamic-json-artifacts --prepare-to /tmp/adamic-json-perf > /tmp/adamic-json-perf/prepare.log
# Repeat c/build for each source snapshot, with its entry path.
go run ./cmd/adamic c stage1/cohere/json/main.ts > /tmp/adamic-json-perf/final.c 2>/tmp/adamic-json-perf/final-build.log
clang -std=c11 -O2 -g -ffp-contract=off -fno-optimize-sibling-calls -Iinternal/native/runtime /tmp/adamic-json-perf/final.c internal/native/runtime/*.c -lm -o /tmp/adamic-json-perf/final >> /tmp/adamic-json-perf/final-build.log 2>&1
VALGRIND_LIB=/tmp/adamic-json-perf/valgrind/usr/libexec/valgrind /tmp/adamic-json-perf/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/tmp/adamic-json-perf/final.callgrind /tmp/adamic-json-perf/final --cases /tmp/adamic-json-perf/profile-cases.txt > /tmp/adamic-json-perf/final.stdout 2>/tmp/adamic-json-perf/final.stderr
python3 stage1/cohere/json/profile.py /tmp/adamic-json-perf/final.callgrind > /tmp/adamic-json-perf/final-top.txt
ADAMIC_JSON_PROFILE_BINARIES=/tmp/adamic-json-perf/baseline:/tmp/adamic-json-perf/final go test -v -count=1 -timeout 30m ./stage1/cohere/json -run 'TestProfileSnapshotsAgree|TestCachedWidthMutantIsCaught' > /tmp/adamic-json-perf/snapshots.log 2>&1
python3 stage1/cohere/json/benchmark.py /tmp/adamic-json-perf /tmp/adamic-json-artifacts/cases.txt > /tmp/adamic-json-perf/timing.log 2>&1
go test -v -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(strings|strings_more|collections|exceptions|bitwise)\.a$' > /tmp/adamic-json-perf/oracle.log 2>&1
go vet ./... > /tmp/adamic-json-perf/vet.log 2>&1
gofmt -l cmd internal stage1/cohere/json > /tmp/adamic-json-perf/gofmt.log 2>&1
/tmp/adamic-json-cohere --no-fix --format-only --format-all stage1/cohere/json/*.ts > /tmp/adamic-json-perf/format-check.log 2>&1
/tmp/adamic-json-cohere --no-fix --no-format stage1/cohere/json/*.ts > /tmp/adamic-json-perf/lint.log 2>&1
```

For the counted build, use the same C/clang command with `-DADAMIC_COUNT`
and no `-g`, output `final-counted`; run it on the complete cases file and
compare stdout with expected.txt. Baseline source was extracted into
`baseline-source/` with `git archive adb4e0aa stage1/cohere/json`, and compiled
from that entry into `baseline.c` and `baseline`. The same preparation script
reconstructs the manifest and both expected streams from independently saved
Go answers. Compare both instrumented sample outputs to profile-expected.txt.
For Go, build `testdata/cohere_driver.go` through an overlay of
`cohere/command/formatter_comparison/main.go`, as buildGoDriver does, into
`/tmp/adamic-json-perf/go-cohere`. The benchmark script uses these saved binaries
and the baseline/optimized source paths, and records every timing and load
sample in measurements.txt. Metadata uses .txt to keep it out of the JSON
corpus. The sample and full corpus must be unchanged between snapshots.
