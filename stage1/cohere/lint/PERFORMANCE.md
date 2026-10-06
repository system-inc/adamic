# Lint runner performance

Twenty rules, starting at `0e97eeb`, on `codex/typescript-scanner`.
Changes stay in this directory. Compiler/runtime observations below are proposed
units; no `internal/` source is changed.

## Method

Pinned TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`,
77 compiler files, 8,215 findings. Count mode still parses whole files, runs every
rule, builds complete findings/messages/repair proposals, and sorts findings.
It skips reporting positions, printing, and applying fixes. Timings include
process startup and input reads, excluding builds. Five fresh-process rounds
interleave native, Go, and Node. No profiling/build/test jobs run during timing.

AMD EPYC 9V74, Linux 6.18.44, `nproc` 5, four-core quota, 17.6 GB memory.
Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup: Go ready 0s, clang/Node/submodules
ready 1s, build cache warm 14s, done 14s. Release uses repository `-O2` flags;
Callgrind uses the same flags plus `-g`, without sanitizers or allocation counters.
Valgrind 3.24.0 is the extracted scanner-unit package in scratch. Its nonfatal
brk-segment warning is preserved in tool logs; completed runs return the exact
8,215 count. Allocation counters come from a separate `ADAMIC_COUNT` binary.

Top twenties and raw profiles are saved under `performance/`. The shared scanner
accounting collapses same-name inline records; self instructions must sum exactly
to the Callgrind summary. Inclusive rows overlap, and recursive clang clones can
show multiple nested paths; do not sum inclusive costs. Anonymous functions remain
in total/self accounting, although excluded from the named top twenty.

A supplementary Go run uses `GOMAXPROCS=1 GODEBUG=asyncpreemptoff=1` for Callgrind,
not for timing: 5,301,243,964 instructions, 645,312.72 per finding, same 8,215
answer. Go scheduler inclusive attribution is not used. Baseline native executes
7.27 times as many instructions. Instruction ratios need not equal timing ratios.
Node is timed, not instruction-profiled.

## Baseline observations

38,559,511,949 instructions, 4,693,793.30 per finding. Best native 3.280604s,
2,504.11 findings/s; Go 0.484782s, 16,945.78/s; Node 1.516196s, 5,418.16/s.
Native is 6.77 times Go elapsed and 2.16 times Node. Load before
`0.47 0.66 0.39`, after `0.65 0.69 0.41`.

Warning-comment processing is 27.57 billion inclusive instructions (71.49%).
Map `find` self is 17.52 billion (45.42%). Caller edges show 6.57 billion in
folding's map reads, 5.58 billion in warning-position map reads, and 6.00 billion
in comment-anchor map writes. These costs overlap `find` and must not be added
as another category. No weak references are used; the `find` ranking here is the
map function, not the runtime's same-named weak-table helper.

The numeric hash in `runtime/map.c` mixes IEEE-754 bits with a right shift by 29
and multiplication, then masks low bits. A reduction of that exact expression
maps all integers 0..1023 to one initial bucket in a 2048-bucket table. For
0..4095 in 8192 buckets only two initial buckets are occupied. The saved
`hash-evidence.log` states initial bucket populations, not measured final probe
lengths. Inspection explains the observed excessive map lookup instructions.
This is a runtime hash-distribution defect, amplified by port representation.

Other disjoint self costs: releases 3.657 billion (9.48%), retains 919.57 million
(2.38%); allocate 425.73M, malloc 296.60M, free 488.81M, realloc 43.15M (named
allocator total 1.254B, 3.25%, excluding RC and anonymous libc). Logical counts:
11,417,906 allocations and frees, 118,809,787 retains, 90,317,497 releases,
peak 738,338 live allocations, zero regions. Literal membership arrays in
comment traversal allocate at every node and retain immortal strings per entry.

UTF-16 costs include locate self 1.132B (2.94%), usable 370.81M, units 184.47M,
char-code 534.58M. These are disjoint self costs, not inclusive string access
costs. Concat 72.22M, append 102.20M, slice 117.31M self are smaller than map/RC.
Finding construction is 1.15M self; reporting's method is 0.674M; sorting's runtime
body 0.201M. Formatting number-to-string is 2.489M self and from-number 0.784M,
under 0.009% combined. Count mode does not format finding positions.

Array access inline header lines 285..310 account for 286.76M self (0.74%);
that range includes access/conversion plumbing, so it is not an isolated count of
redundant bounds checks. Class shape checks, field access and stack checks also
live in generated method bodies. No claim that they explain the whole slowdown.

A separate full-output profile of `src/compiler/core.ts` is 758,875,429 Ir,
with exact Go/Node/native output including fixes. Number format/from-number self
are only 47,007/16,269; `written` is 95.63M inclusive (12.60%) for escaped output,
mostly the unchanged whole-source record. This single-file result is not an
all-corpus reporting benchmark. Position/message formatting is not the leading
problem in either measured workload.

## Port changes

1. Compare comment candidates as folded Unicode scalars, stopping at the first
   mismatch. ASCII folding uses arithmetic; non-ASCII keeps the generated Go
   simple-fold table, including Kelvin sign and long s. Removes folded term and
   candidate substring/string construction. Nothing is deferred or omitted in
   count mode. The Unicode generator emits the same helper.

2. Replace literal-end, anchor and reachable-comment numeric maps/sets with
   bounded position arrays. Literal ends initialize lazily, so a selected rule
   with no comment/sequence work does not allocate that array. Helper methods
   take readonly arrays as explicit visitor parameters, keeping borrowing without
   violating cohere's property-alias rule. Per-owner small child maps remain.

The first array trial was byte-identical but failed cohere on two property aliases;
its measurements are retained as `2-position-arrays`. The borrowed-parameter
adjustment is `2b-position-arrays`: 20,970,336,919 Ir, 2,552,688.61 per finding,
4,101.37 native findings/s, Go 17,098.74, Node 4,203.20. Compared with scalar
matching, it removes 9,967,576,018 Ir (32.22%). Load before `0.42 0.70 0.65`,
after `0.62 0.73 0.66`. The extra method boundaries add 29.39M Ir versus the
trial (0.14%); the small timing difference does not establish a speed gain.

Arrays require about 24 bytes per UTF-16 position plus headers/capacity. Separate
child peak-RSS measurements were 100,996 KiB baseline and 152,160 KiB array trial;
these concurrent runs are not timing samples. `/usr/bin/time` was unavailable,
so Python `resource.getrusage(RUSAGE_CHILDREN)` in separate worker processes was
used. Logical allocation peak alone hides these larger array buffers. Node
regresses from 5,702.12 to 4,203.20 findings/s after the dense initialization.
This is an explicit memory/Node tradeoff for native lookup performance.

Both array forms pass full sanitized Go/Node/native findings and fixed output.
The final array form has 578,745 fixture bytes plus 15,550,205 source-corpus bytes.
Its release/debug/Node snapshot check has 16,140,914 identical bytes. Differences
between comparison byte counts include changed port source and temporary fixture
path lengths; each run compares all bytes against Go over the identical inputs.
The new position mutant clears anchor zero after collecting anchors; both Node
and native lose the first TODO finding at case 100, line 2068. It compiled and
executed successfully. Cohere, vet and the filtered one-byte core oracle pass.

3. Initialize constant boolean masks with `new Array<boolean>(length).fill(false)`
   rather than `Array.from` callbacks. Saves 582,826,634 Ir (2.78%) and 154 closure
   allocations. Native best timing regresses 3.37%, from 2.002988s to 2.070502s;
   this timing does not support a native speedup. Node improves 1.501 times,
   from 4,203.20 to 6,309.38 findings/s, reversing its initialization regression.
   Retained because deterministic native work decreases and Node improves.
   Load before `0.42 0.63 0.66`, after `0.51 0.64 0.66`. Full sanitized parity:
   578,745 fixture bytes and 15,550,195 corpus bytes. Release/debug/Node snapshot
   parity: 16,140,904 bytes. Cohere passes. Memory-domain representation remains
   the same; fill does not remove the position arrays' larger memory cost.

## Proposed compiler and runtime units

- Numeric Map/Set hash distribution: use an avalanche with adequate low-bit
  distribution before a power-of-two mask. Evidence is the 45.42% `find` self,
  caller edges, and the exact hash-expression bucket reduction above. Preserve
  SameValueZero for NaN and signed zero, insertion order, tombstones and live
  iterators. The port workaround does not fix arbitrary numeric maps elsewhere.
- Compiler literal membership tests: generated C for `[...].includes(kind)`
  creates a heap array, pushes retained immortal strings, runs generic includes,
  then frees it. Comment anchors alone call array creation 1,763,938 times.
  Investigate scalarization or safe static data for immediate nonescaping
  membership. Mutable literal identity/JavaScript evaluation order must survive.
- Compiler ownership across node access: `Linter.node` delegates to `Parser.node`
  and receives an owned result. The baseline has 118.8M retains; node lookup,
  child-array iteration and nested field reads emit RC traffic even while Parser
  retains every node. Borrowed return/field lifetime proofs could reduce it.
  Counts measure all traffic, not a proof that every retain can be deleted.
- UTF-16/source views remain the scanner unit's proposed runtime task. This
  profile records 1.132B locate self and repeated public string operations.
  Existing ASCII fast paths/checkpoints already exist; this is not a claim that
  native rescans every source string from byte zero on every read.
- Finding formatting/dtoa: this profile does not justify a separate optimization
  unit. Dynamic finding strings, position mapping and number formatting do not
  dominate; larger printer work is the canonical escaped whole-source record.

## Reproduction

All test/tool outputs are redirected to files. Build one directory per source
snapshot; the helper saves generated C, runnable TS, runtime C, release, counted,
Go oracle, debug binary and the pinned manifest:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_LINT_PROFILE_DIR=/workspace/scratch/lint-perf/1-scalar-comments go test ./stage1/cohere/lint -run '^(TestProfileArtifacts|TestRulesAgree|TestCompilerAndStage1Agree)$' -count=1 -v -timeout 30m > /workspace/scratch/lint-perf/1-build-parity.log 2>&1
VALGRIND_LIB=/workspace/scratch/scanner-perf/valgrind/usr/libexec/valgrind python3 stage1/cohere/lint/profile.py /workspace/scratch/lint-perf/1-scalar-comments --valgrind /workspace/scratch/scanner-perf/valgrind/usr/bin/valgrind > /workspace/scratch/lint-perf/1-scalar-comments/profile.log 2>&1
```

## Measured steps

| Step | Native findings/s | Go findings/s | Node findings/s | Native Ir | Ir/finding |
| --- | ---: | ---: | ---: | ---: | ---: |
| baseline | 2,504.11 | 16,945.78 | 5,418.16 | 38,559,511,949 | 4,693,793.30 |
| 1-scalar-comments | 2,941.27 | 16,456.40 | 5,702.12 | 30,937,912,937 | 3,766,027.14 |
| 2-position-arrays | 4,018.26 | 16,606.44 | 4,083.68 | 20,940,946,893 | 2,549,111.00 |
| 2b-position-arrays | 4,101.37 | 17,098.74 | 4,203.20 | 20,970,336,919 | 2,552,688.61 |
| 3-filled-masks | 3,967.64 | 16,847.92 | 6,309.38 | 20,387,510,285 | 2,481,741.97 |

Scalar comparison removes 7,621,599,012 Ir (19.77%), 1,766,845 logical allocations, and improves native throughput 1.175 times. Load before `0.67 0.73 0.46`, after `0.79 0.75 0.48`. Full sanitized parity: 577,682 fixture bytes plus 15,548,752 source-corpus bytes, 16,126,434 total. Cohere passes the changed source.

## Baseline top twenties

### Inclusive

| Function | Instructions | Share |
| --- | ---: | ---: |
| `main` | 38,559,323,748 | 100.00% |
| `adamic_function_50_run` | 38,542,204,801 | 99.96% |
| `adamic_function_167_Linter_run` | 37,829,625,289 | 98.11% |
| `adamic_function_180_Linter_walk` | 31,755,401,905 | 82.35% |
| `adamic_function_187_Linter_additional` | 30,107,682,523 | 78.08% |
| `adamic_function_198_Linter_warnings` | 27,566,479,223 | 71.49% |
| `find` | 17,549,131,384 | 45.51% |
| `adamic_function_130_Statements_statement'2` | 14,325,736,465 | 37.15% |
| `adamic_map_get` | 12,688,359,186 | 32.91% |
| `adamic_function_197_Linter_commentAnchors` | 10,265,553,333 | 26.62% |
| `adamic_function_197_Linter_commentAnchors'2` | 10,265,217,929 | 26.62% |
| `adamic_function_112_Statements_block'2` | 8,545,395,290 | 22.16% |
| `adamic_function_38_fold` | 7,412,478,806 | 19.22% |
| `adamic_map_set` | 6,067,692,216 | 15.74% |
| `adamic_function_107_Parser_file` | 6,030,950,885 | 15.64% |
| `adamic_function_130_Statements_statement` | 6,029,824,794 | 15.64% |
| `adamic_function_125_Statements_functionDeclaration` | 5,251,487,167 | 13.62% |
| `adamic_function_112_Statements_block` | 5,053,503,671 | 13.11% |
| `adamic_function_125_Statements_functionDeclaration'2` | 4,648,204,059 | 12.05% |
| `adamic_release` | 4,207,107,650 | 10.91% |

### Self

| Function | Instructions | Share |
| --- | ---: | ---: |
| `find` | 17,515,650,254 | 45.42% |
| `adamic_release` | 3,657,053,349 | 9.48% |
| `adamic_string_equal` | 1,599,934,210 | 4.15% |
| `adamic_function_198_Linter_warnings` | 1,184,845,782 | 3.07% |
| `adamic_string_locate` | 1,132,261,690 | 2.94% |
| `adamic_map_set` | 1,039,519,344 | 2.70% |
| `adamic_retain` | 919,570,267 | 2.38% |
| `adamic_function_21_Scanner_code` | 795,854,329 | 2.06% |
| `adamic_array_index_of` | 753,685,754 | 1.95% |
| `adamic_array_push` | 612,540,077 | 1.59% |
| `adamic_function_32_Scanner_scan` | 601,635,118 | 1.56% |
| `adamic_string_char_code` | 534,582,963 | 1.39% |
| `free` | 488,805,618 | 1.27% |
| `adamic_function_54_Parser_node` | 470,693,280 | 1.22% |
| `adamic_allocate` | 425,731,134 | 1.10% |
| `adamic_function_197_Linter_commentAnchors'2` | 409,300,729 | 1.06% |
| `adamic_function_169_Linter_enabled` | 401,908,110 | 1.04% |
| `usable` | 370,808,392 | 0.96% |
| `adamic_function_187_Linter_additional` | 337,160,718 | 0.87% |
| `malloc` | 296,601,126 | 0.77% |


First-step release/debug/Node snapshot comparison passed for baseline and scalar
snapshots: 16,139,461 identical bytes each on the same final source corpus.
The scalar mutant changes ASCII folding by one: it compiles and executes, then
both Node and sanitized native lose the first TODO finding at generated case
100, line 2068. The accounting mutant increments only Callgrind `summary:` by
one; the summarizer refuses with `callgrind self costs do not sum to summary`.
Their complete logs are saved. These are wrong-answer/accounting failures,
not compilation failures.
