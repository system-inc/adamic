# Scanner performance follow-up

Branch `codex/typescript-scanner`, starting scanner commit `0e1a6b3`.
Only this directory changed. Compiler and runtime findings below are proposed
units, not changes to `internal/`.

Native throughput improved from 600,634 to 2,081,602 tokens/s, 3.47 times.
Callgrind instructions decreased from 9,322,686,277 to 2,429,614,991,
73.94%, or 21,441.81 to 5,588.02 instructions/token. Final native is
6.30 times slower than Go and 1.013 times slower than Node in the final
best-of-five round. Step 6 briefly beat Node; the final timing does not.
The single-piece change removes 3,739 allocations and 1,830,266 instructions
but its 0.66% timing regression is within run-to-run variation. Instruction
counts, not that small timing difference, support retaining the change.

## Method and measurements

TypeScript v6.0.3 commit `050880ce59e30b356b686bd3144efe24f875ebc8`,
77 compiler files, 434,790 tokens in all runs. The count driver still computes
token values, flags and errors. Timing includes process startup, file reads
and scanning, excludes token printing. This is the existing scanner driver's
protocol, including its explicitly requested rescans, not a parser benchmark.

AMD EPYC 9V74; `nproc` 5; cgroup quota 4 CPUs; 17.6 GB memory. Go 1.27.1,
Node 24.19.0, clang 20.1.8, release `-O2` with the repository's native flags.
Five fresh-process rounds interleave native, Go, Node. No other worker job
was run during timing. Load after each group (three averages and process counts):

- baseline: `baseline load recorded in original REPORT.md`
- 1-identifier: `0.52 0.54 2.00 1/138 44761`
- 2-advance: `0.39 0.47 1.65 1/139 45385`
- 3-count-driver: `0.28 0.38 1.31 1/142 46040`
- 4-bmp-read: `0.46 0.31 0.96 1/143 46712`
- 5-punctuation: `0.14 0.24 0.73 1/142 47307`
- 6-raw-units: `0.12 0.16 0.49 1/143 47976`
- 7-single-piece: `0.11 0.19 0.41 1/143 48603`

The original setup timing remains in [REPORT.md](REPORT.md): Go 0s, clang 0s,
Node 0s, submodules 1s, cache warm 14s, total 14s. Toolchain was reused.
Callgrind 3.24.0 was unavailable initially. Unprivileged apt install failed
with permission denied; sudo was absent; the configured snapshot mirror
returned HTTP 403. The Debian package `valgrind_3.24.0-3_amd64.deb` was
downloaded from deb.debian.org and extracted into scratch. `VALGRIND_LIB`
points at its `usr/libexec/valgrind`. Baseline through step 5 emitted a
nonfatal Valgrind brk-segment warning, completed with the exact token count;
steps 6 and 7 completed without that warning. Allocation counts are from a
separate `ADAMIC_COUNT` build, not the instrumented Callgrind build.

Callgrind records `Ir` on the release flags plus `-g`, without sanitizers or
RC instrumentation. Top functions collapse same-name DWARF inline records
and exclude anonymous addresses and `(below main)` from the named ranking.
All self costs, including excluded names, sum exactly to the Callgrind
summary; the script refuses a mismatch. Inclusive rows overlap and must not
be added. Full raw baseline/final profiles are in `performance/*.callgrind.gz`;
all intermediate top twenties and five timing samples are committed in
[performance/measurements.json](performance/measurements.json) and the adjacent
text logs. Comparison, mutant and tool diagnostic logs are saved there too.
Native source snapshot commits are listed below.

| Change | Commit | Native tokens/s | Go tokens/s | Node tokens/s | Native Ir | Ir/token |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Baseline | `0e1a6b3` | 600,634 | 12,986,559 | 922,788 | 9,322,686,277 | 21,441.81 |
| Slice contiguous identifiers | `5b4a5a7` | 706,663 | 12,373,029 | 971,618 | 7,856,053,472 | 18,068.62 |
| Reuse decoded scalar on advance | `72fe01d` | 899,161 | 12,835,108 | 1,029,928 | 5,942,889,931 | 13,668.41 |
| Omit byte-offset table in count mode | `c98108c` | 1,250,492 | 12,165,127 | 1,744,522 | 3,806,256,243 | 8,754.24 |
| BMP charCodeAt fast path | `d5c60eb` | 1,521,489 | 12,059,307 | 1,722,360 | 3,084,480,085 | 7,094.18 |
| Dispatch punctuators without slices | `5cf0144` | 1,716,000 | 11,697,116 | 1,849,764 | 2,645,557,023 | 6,084.68 |
| Read raw bodies as code units | `9012623` | 2,095,365 | 11,877,364 | 1,930,873 | 2,431,445,257 | 5,592.23 |
| Return one cooked piece directly | `c230f01` | 2,081,602 | 13,120,602 | 2,107,845 | 2,429,614,991 | 5,588.02 |

Logical allocation counters, not allocated bytes. Frees equal allocations
at every step; peak live allocations 2,262; regions 0 throughout.

| Snapshot | Allocations | Retains | Releases |
| --- | ---: | ---: | ---: |
| baseline | 4168924 | 1831342 | 6209051 |
| 1-identifier | 1118972 | 2441827 | 3769584 |
| 2-advance | 1118972 | 2441827 | 3769584 |
| 3-count-driver | 1118972 | 2441750 | 3769507 |
| 4-bmp-read | 1118972 | 2441750 | 3769507 |
| 5-punctuation | 435938 | 2441910 | 2636846 |
| 6-raw-units | 435938 | 2441910 | 2636846 |
| 7-single-piece | 432199 | 2450881 | 2642078 |

## Top 20 inclusive instructions

Baseline:

| Rank | Function | Ir | Share |
| --- | --- | ---: | ---: |
| 1 | `main` | 9,322,507,326 | 100.00% |
| 2 | `run` | 9,321,783,208 | 99.99% |
| 3 | `Scanner_scan` | 6,752,900,489 | 72.44% |
| 4 | `adamic_string_code_point_at` | 4,608,743,384 | 49.44% |
| 5 | `Scanner_code` | 3,616,276,675 | 38.79% |
| 6 | `Scanner_template` | 2,696,020,840 | 28.92% |
| 7 | `adamic_string_locate` | 2,668,688,683 | 28.63% |
| 8 | `Scanner_identifier` | 2,541,482,449 | 27.26% |
| 9 | `Scanner_advance` | 1,992,294,730 | 21.37% |
| 10 | `unit_at` | 1,588,551,758 | 17.04% |
| 11 | `adamic_string_from_code_points` | 722,498,639 | 7.75% |
| 12 | `usable` | 569,065,410 | 6.10% |
| 13 | `adamic_string_append` | 551,490,381 | 5.92% |
| 14 | `adamic_release` | 433,506,484 | 4.65% |
| 15 | `adamic_string_concat` | 351,061,298 | 3.77% |
| 16 | `adamic_string_slice` | 324,488,745 | 3.48% |
| 17 | `adamic_string_units` | 251,888,985 | 2.70% |
| 18 | `adamic_string_join_halves` | 239,745,081 | 2.57% |
| 19 | `adamic_read_text_file` | 226,519,917 | 2.43% |
| 20 | `decode` | 225,777,641 | 2.42% |

Final:

| Rank | Function | Ir | Share |
| --- | --- | ---: | ---: |
| 1 | `main` | 2,429,435,818 | 99.99% |
| 2 | `run` | 2,428,711,381 | 99.96% |
| 3 | `Scanner_scan` | 2,144,325,404 | 88.26% |
| 4 | `Scanner_template` | 947,709,220 | 39.01% |
| 5 | `adamic_string_char_code` | 707,400,508 | 29.12% |
| 6 | `Scanner_code` | 509,850,527 | 20.98% |
| 7 | `Scanner_identifier` | 470,209,573 | 19.35% |
| 8 | `adamic_string_locate` | 441,849,530 | 18.19% |
| 9 | `unit_at` | 438,287,338 | 18.04% |
| 10 | `adamic_read_text_file` | 228,998,266 | 9.43% |
| 11 | `decode` | 225,777,620 | 9.29% |
| 12 | `usable` | 153,769,189 | 6.33% |
| 13 | `adamic_string_slice` | 122,762,521 | 5.05% |
| 14 | `adamic_string_units` | 104,515,436 | 4.30% |
| 15 | `adamic_release` | 103,717,182 | 4.27% |
| 16 | `adamic_array_join` | 100,712,463 | 4.15% |
| 17 | `isIdentifierPart` | 95,756,397 | 3.94% |
| 18 | `adamic_string_concat` | 89,162,854 | 3.67% |
| 19 | `adamic_string_join_halves` | 82,545,858 | 3.40% |
| 20 | `isIdentifierStart` | 73,690,719 | 3.03% |

## Top 20 self instructions

Baseline:

| Rank | Function | Ir | Share |
| --- | --- | ---: | ---: |
| 1 | `adamic_string_locate` | 2,099,623,273 | 22.52% |
| 2 | `adamic_string_code_point_at` | 1,463,437,128 | 15.70% |
| 3 | `Scanner_code` | 578,768,909 | 6.21% |
| 4 | `usable` | 569,062,855 | 6.10% |
| 5 | `run` | 486,970,422 | 5.22% |
| 6 | `adamic_release` | 378,108,866 | 4.06% |
| 7 | `unit_at` | 373,137,482 | 4.00% |
| 8 | `Scanner_advance` | 337,296,436 | 3.62% |
| 9 | `adamic_string_units` | 251,888,985 | 2.70% |
| 10 | `adamic_string_join_halves` | 239,745,081 | 2.57% |
| 11 | `decode` | 225,747,318 | 2.42% |
| 12 | `adamic_string_append` | 205,020,447 | 2.20% |
| 13 | `Scanner_scan` | 187,570,859 | 2.01% |
| 14 | `adamic_array_push` | 154,446,557 | 1.66% |
| 15 | `adamic_allocate` | 150,074,192 | 1.61% |
| 16 | `free` | 148,698,157 | 1.60% |
| 17 | `Scanner_template` | 129,927,701 | 1.39% |
| 18 | `adamic_string_slice` | 126,736,694 | 1.36% |
| 19 | `adamic_string_from_code_points` | 125,392,050 | 1.35% |
| 20 | `adamic_string_concat` | 119,583,485 | 1.28% |

Final:

| Rank | Function | Ir | Share |
| --- | --- | ---: | ---: |
| 1 | `Scanner_template` | 354,578,754 | 14.59% |
| 2 | `adamic_string_locate` | 288,080,341 | 11.86% |
| 3 | `Scanner_code` | 245,166,344 | 10.09% |
| 4 | `decode` | 225,747,678 | 9.29% |
| 5 | `Scanner_scan` | 189,069,591 | 7.78% |
| 6 | `adamic_string_char_code` | 157,435,393 | 6.48% |
| 7 | `usable` | 153,766,828 | 6.33% |
| 8 | `adamic_string_units` | 104,515,436 | 4.30% |
| 9 | `adamic_release` | 96,938,779 | 3.99% |
| 10 | `adamic_string_join_halves` | 82,545,858 | 3.40% |
| 11 | `isIdentifierStart` | 73,686,440 | 3.03% |
| 12 | `Scanner_identifier` | 69,674,447 | 2.87% |
| 13 | `unit_at` | 48,729,266 | 2.01% |
| 14 | `isIdentifierPart` | 39,254,925 | 1.62% |
| 15 | `adamic_string_slice` | 37,671,137 | 1.55% |
| 16 | `find` | 36,304,556 | 1.49% |
| 17 | `adamic_string_equal` | 33,785,653 | 1.39% |
| 18 | `adamic_retain` | 18,284,449 | 0.75% |
| 19 | `run` | 17,400,997 | 0.72% |
| 20 | `adamic_allocate` | 15,547,828 | 0.64% |

## Cost attribution and scanner changes

Observations: baseline codePointAt has 4.609 billion inclusive instructions,
49.44% of the entire process. `adamic_string_locate` alone has 2.100 billion
self instructions (22.52%). The generated `Scanner.code` and `Scanner.advance`
have 578.8 and 337.3 million self instructions. The original scanner decodes
again when advancing and uses codePointAt for BMP and raw bodies. Go scans
bytes directly. A supplemental Go Callgrind run, with `GOMAXPROCS=1` and
`GODEBUG=asyncpreemptoff=1`, has 308,939,105 total instructions, 710.55/token:
baseline native executes about 30.2 times that work; final about 7.86 times.
The supplementary Go raw profile and tool output are also committed in
`performance/go.*`. Go's scheduler confuses Callgrind inclusive attribution, so its inclusive
rows are not used. Timed Go runs use the normal environment. Instruction
ratios are not expected to equal wall-time ratios. Node was timed but not
profiled; no Node allocation or instruction attribution is claimed.

1. Identifiers originally built strings one code point at a time. Contiguous
   identifiers now take one source slice; escapes append source spans and
   decoded escapes. Removes 3,049,952 logical allocations and 1.467 billion
   instructions. Runtime append is amortized linear when uniquely owned;
   the measured problem is repeated code-point string creation and generic
   appends, not proof of quadratic behavior.
2. Advance takes the already decoded scalar. Removes another 1.913 billion
   instructions without changing allocation counts. Identifier and escape
   positions still advance by scalar width.
3. Count mode previously built a UTF-16 to byte-offset array that no counted
   token used. Its removal saves 2.137 billion instructions. Printing mode
   retains the complete table and byte-range checks. This is a driver fix,
   not a reduction of work in the scanner's printable protocol.
4. Read BMP via charCodeAt, falling back to codePointAt only for a high
   surrogate. Removes 721.8 million instructions while retaining supplementary
   identifier semantics and EOF behavior.
5. Switch on punctuation code units, returning fixed spellings; attempt an
   identifier only when its first character qualifies. Removes speculative
   source slices and 683,034 allocations, saving 438.9 million instructions.
   The existing authoritative kind table remains in use, including `!=`.
6. Strings, templates, comments, shebang and conflict text walk raw UTF-16
   units; escapes and identifiers retain scalar reads. Saves 214.1 million
   instructions. Added 12 surrogate and Unicode-line-break inputs. JSX keeps
   its previous scalar behavior to preserve the Go scanner's whitespace flag
   behavior for supplementary characters.
7. Return the sole cooked piece directly rather than calling join. Go's
   strings.Join has that fast path. Multi-piece joining remains unchanged.
   Saves 3,739 allocations and 1.83 million instructions.

Baseline string-building self costs include append 205.0M, fromCodePoints
125.4M, concat 119.6M, slice 126.7M and join-halves 239.7M. These are disjoint
self costs; their inclusive costs also contain UTF-16 work and allocation.
Baseline release self is 378.1M (4.06%), retain 13.6M (0.15%). Logical RC
operations are much more numerous than heap allocations because temporary
values and immortal strings also pass through RC checks. Baseline named
allocate/malloc/free/realloc self sum to 431.9M (4.63%), excluding separately
counted release and anonymous libc costs. Small allocations use runtime
slabs, so one logical allocation does not mean one malloc call.

Bounds and field checks: baseline `adamic.h` object-field-cache lines 170-176
account for 94.0M self instructions (1.01%); generated class shape checks
also reside in main.c and are not separately isolated. Array bounds-access
lines 290-300 account for just 1,433 instructions in count mode. String
bounds/truncation checks are included in string-access costs; they are not
separately proven redundant. Stack checks inline at generated method entries.
These observations do not justify attributing the whole slowdown to checks.

Number to string is small here: baseline adamic_number_format self 934,289,
adamic_string_from_number self 221,009, the latter inclusive 1,546,168,
0.0166% of total. No dtoa optimization is warranted by this profile. Printing
positions and flags is excluded by count mode, so this is not a statement
about output-heavy formatting. Bigint decimal string arithmetic is likewise
not a leading function in this compiler corpus.

## Proposed runtime and compiler units

These are evidence-backed follow-ups, not measured gains from changes made here.

- Runtime codePointAt: `runtime/string.c:387-403` locates the UTF-16 position,
  then calls unit_at for BMP, which locates it again. Eliminate the second
  translation while preserving undefined/out-of-range and lone-surrogate
  behavior. The baseline 49.44% inclusive code-point path proves priority;
  this scanner now largely avoids it, so the baseline gain is not still
  available on the optimized scanner.
- Runtime UTF-16 views: final locate self 288.1M, usable 153.8M, unit_at 48.7M,
  units 104.5M and char_code 157.4M. `string_index.c` checkpoints every 32 units
  and keeps a forward cursor; backward lookahead or slicing can restart from
  a checkpoint. The header ASCII fast path already exists, but a file with
  even one non-ASCII character uses the indexed path throughout. Investigate
  sequential views and efficient nearby backward access, without weakening
  UTF-16 semantics. Do not mistake this for an absent ASCII fast path.
- Compiler scanner access: final generated Scanner.code self 245.2M (10.09%),
  3,827,809 calls. Generated C performs class shape checks, wraps optional
  offsets, executes entry stack checks and uses generic numeric/string APIs.
  Template and scan self are another 354.6M and 189.1M. Investigate proven
  monomorphic leaf access, integer indices and safe inlining. The profile
  measures whole method bodies, not an isolated cost for each check.
- Runtime input decoding: `input.c` decode performs sizing and writing passes
  for WHATWG UTF-8 decoding, 225.75M self instructions, 9.29% of final Ir.
  Investigate an ASCII/valid-UTF-8 fast path. Preserve Node replacement behavior
  for malformed bytes; the current scanner corpus is valid UTF-8, so that
  broader semantic validation belongs to the runtime unit.
- Runtime string slices and joins: final still allocates 432,199 objects,
  roughly one/token. `string_share.c` copies below 64 bytes or below one
  quarter of its owner. Most lexemes therefore copy even while Scanner.text
  keeps source alive; Go source slices are headers without an allocation.
  Investigate source-lifetime-anchored borrowed views rather than blindly
  changing the retention policy. A runtime single-piece join fast path and
  compiler elimination of the temporary pieces array are separate candidates.
- Compiler ownership/runtime RC: final retains 2,450,881 and releases 2,642,078;
  self costs 18.28M and 96.94M, together 4.74% of final instructions. Final
  named allocate/malloc/free/realloc self is 23.50M, 0.97%. Ownership-proven
  borrowed temporaries and fewer generic value transitions could remove part
  of the RC traffic. Do not add inclusive RC to allocator costs: it overlaps.

## Validation, mutants, commands and limits

Each step passed the complete scanner package: both gap refusal probes,
ASan/UBSan/LeakSanitizer native, Node, Go, and all three comparison mutants.
Final corpus: 77 compiler files, 61 stage1 files, 18,236 generated inputs;
23,816,619 identical answer bytes. This byte count differs from the original
report because the corpus includes the scanner's changing source itself.

Every measured snapshot, including baseline, additionally passed release,
-O2 -g profiled, and Node comparisons against the same final full Go corpus:
23,816,619 bytes each. `TestProfileSnapshotsAgree` passed in 41.005s.
The final step's complete scanner package passed in 48.363s.

Mutants, all observed successful executions with wrong answers:

| Mutant | What caught it |
| --- | --- |
| `!=` kind becomes `==` | Node and sanitized native byte comparison, line 348 |
| Repeated decimal separator accepted | Both byte comparisons, line 689: NumericLiteral instead of error 6189 |
| Regex rescan skipped | Both byte comparisons, line 126235: SlashToken instead of RegularExpressionLiteral |
| Release snapshot `!=` kind becomes `==` | TestProfileSnapshotsAgree fails, line 32502; demonstrates new release-build comparison |
| Callgrind summary increased by one | profile.py refuses with self costs do not sum to summary; demonstrates accounting check |

The filtered core oracle passed nine string/index/append/surrogate fixtures
in 16.555s (the substring filter also includes optional_strings and
undefined_strings). `go vet ./...`, gofmt and diff whitespace checks passed.
Cohere passed all four scanner TS files, 276 rules, 100% Adamic-ready.
No full repository gate was rerun for these scanner-only performance changes;
the original unit's full gate remains recorded in REPORT.md. No compiler or
runtime fix was made. Existing scanner API limits in GAPS.md remain.

Reproduce each source commit with a unique scratch snapshot. Actual commands
(all test and tool output is redirected to files):

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_SCANNER_PROFILE_DIR=/workspace/scratch/scanner-perf/7-single-piece go test -count=1 -v -timeout 30m ./stage1/typescript/scanner > /tmp/scanner-perf-7-compare.log 2>&1
VALGRIND_LIB=/workspace/scratch/scanner-perf/valgrind/usr/libexec/valgrind python3 stage1/typescript/scanner/profile.py /workspace/scratch/scanner-perf/7-single-piece --valgrind /workspace/scratch/scanner-perf/valgrind/usr/bin/valgrind > /workspace/scratch/scanner-perf/7-single-piece/profile.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_SCANNER_PROFILE_SNAPSHOTS=/workspace/scratch/scanner-perf/baseline:/workspace/scratch/scanner-perf/1-identifier:/workspace/scratch/scanner-perf/2-advance:/workspace/scratch/scanner-perf/3-count-driver:/workspace/scratch/scanner-perf/4-bmp-read:/workspace/scratch/scanner-perf/5-punctuation:/workspace/scratch/scanner-perf/6-raw-units:/workspace/scratch/scanner-perf/7-single-piece go test -count=1 -v -run '^TestProfileSnapshotsAgree$' -timeout 30m ./stage1/typescript/scanner > /tmp/scanner-perf-snapshots.log 2>&1
go test -count=1 -v -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(strings|strings_more|string_positions|string_index|string_append|shared_slices|lone_surrogates)\.a$' -timeout 30m ./internal/oracle > /tmp/scanner-perf-filtered-oracle.log 2>&1
go vet ./... > /tmp/scanner-perf-vet.log 2>&1
/workspace/scratch/cohere --no-fix --no-cache stage1/typescript/scanner/main.ts stage1/typescript/scanner/scanner.ts stage1/typescript/scanner/characters.ts stage1/typescript/scanner/tokens.ts > /tmp/scanner-perf-cohere.log 2>&1
```

Supplemental Go profile command:

```sh
GOMAXPROCS=1 GODEBUG=asyncpreemptoff=1 VALGRIND_LIB=/workspace/scratch/scanner-perf/valgrind/usr/libexec/valgrind /workspace/scratch/scanner-perf/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/scanner-perf/go-profile/callgrind.out /workspace/scratch/scanner-perf/1-identifier/oracle --manifest /workspace/scratch/scanner-perf/compiler.txt --count > /workspace/scratch/scanner-perf/go-profile/stdout.log 2> /workspace/scratch/scanner-perf/go-profile/stderr.log
```

Additional mutant commands used scratch copies, leaving port source intact:
copy final four TS files into `mutant-release`, replace the single tokens.ts
pair `['!=', 'ExclamationEqualsToken']` with `['!=', 'EqualsEqualsToken']`,
then build and duplicate its release binary as profiled. For accounting,
copy final callgrind.out into `mutant-summary` and increase `summary:` by one.

```sh
go run ./cmd/adamic build /workspace/scratch/scanner-perf/mutant-release/main.ts -o /workspace/scratch/scanner-perf/mutant-release/scanner > /tmp/scanner-perf-mutant-build.log 2>&1
cp /workspace/scratch/scanner-perf/mutant-release/scanner /workspace/scratch/scanner-perf/mutant-release/profiled
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_SCANNER_PROFILE_SNAPSHOTS=/workspace/scratch/scanner-perf/mutant-release go test -count=1 -v -run '^TestProfileSnapshotsAgree$' -timeout 30m ./stage1/typescript/scanner > /tmp/scanner-perf-mutant-release.log 2>&1
python3 stage1/typescript/scanner/profile.py /workspace/scratch/scanner-perf/mutant-summary --summarize-only > /tmp/scanner-perf-mutant-summary.log 2>&1
```

Both mutant check commands exited nonzero for the intended reason; the
mutant executable built and ran normally.
