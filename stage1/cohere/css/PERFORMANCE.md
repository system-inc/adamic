Native CSS formatting rises from 763 to 1,917 stylesheets/s, a measured 2.51x.
The gap to Go falls from 7.72x to 3.08x on five interleaved rounds.
Callgrind instructions fall 66.1%; sample allocations fall 65.7%.
The complete byte oracle, sanitizer, leak and mutation results are recorded below.
Compiler/runtime proposals include a proved shared-slice append overflow.

# Scope and source

This is the speed follow-up to `NATIVE_REPORT.md`, on
`codex/stage1-css-printer`. Baseline is
`23974240d9043602de8f8679065f4b129bd0a7a5`, which already contains the requested
regex cycle proof `32f8106`. Only the CSS slice changes. Compiler/runtime
proposals below are proposals; no compiler or runtime source was modified.
The production changes are `convert.ts`, `print_doc.ts` and `print.ts`.
The source snapshots, generated C and corpus hashes are recorded in
`performance/{baseline,final}-sha256.json`; generated C and raw profiles are
archived with reproducible gzip headers. The final source is identified by those hashes and the commit containing
this report.

# Profile and port changes

Callgrind 3.24.0 records `Ir` for the complete native parse-and-print executable,
compiled with the ordinary release flags (`-O2`) plus `-g`. The profiler build
compiles the unchanged runtime C directly to preserve its symbols; timings use
`native.Build` and its ordinary cached runtime library, without sanitizers or
counting. Full outputs of both builds and the counting build are checked against
Go in both option sets, separately from their timing checksums.

The fixed sample is the first 32 shared successes, every sixteenth shared
success, and the largest shared input: 341 CSS/SCSS cases, 33,232 UTF-16 output
units. It exercises comments, rules, at-rules, selectors, values and Unicode
without spending Callgrind time on the 19,110 formatter refusals per option
set. The full byte-validation corpus includes those refusals and error
positions. `performance/sample.txt` and `manifest.tsv` retain the sample;
manifest columns are shared-corpus index, encoded input bytes and output units.
Both samples and their manifests are identical. Callgrind reports its usual
`brk segment overflow` warning for both binaries; both exit successfully and
produce the exact expected output checksum. Instrumented instruction counts
are not wall-time measurements.

The scanner's checked Callgrind parser is reused by `profile.py`. Same-name
DWARF inline records are collapsed, and every raw self instruction must add
up to `summary`. Inclusive rows overlap and must never be added. A summary
mutant adding one instruction is rejected with
`callgrind self costs do not sum to summary`.

The baseline shows two large avoidable port costs: `byteSlice` uses
1,171,974,355 inclusive instructions (26.71%), and regexp replacement uses
1,691,620,205 (38.56%). Their inclusive costs overlap with other rows.

* `byteSlice` now returns immediately for empty/reversed ranges, slices ASCII
  by its identical byte/UTF-16 coordinates, and stops Unicode iteration at the
  requested end. The original boundary-replacement behavior stays intact.
  Equality of UTF-8 byte length and UTF-16 length proves ASCII in this input
  model; every non-ASCII scalar or lone surrogate uses more bytes than units.
  Its final inclusive cost is 18,853,636 instructions (1.27%).
* The document printer trims only terminal ASCII space and tab with character
  reads, rather than searching the entire accumulated output with a regexp
  at every hard line. The already-unused trailing-whitespace bookkeeping in
  `fits` is removed; pending flat-line spaces are still measured.
* Document text is immutable, so display widths are cached lazily per document.
  The exact old ASCII range 0x20 through 0x7f is checked directly. Unicode,
  surrogate, emoji and wide-character width rules are unchanged.
* The immutable CSS unit list is built once rather than split for each number.

Total instructions: **4,387,115,814 -> 1,487,775,242**, a **66.09%** reduction.
`Documents.print` falls from 1,839,631,017 to 105,788,399 inclusive instructions.
The safe trim branch concatenates the sliced prefix and newline directly;
ordinary untrimmed lines retain amortized appending.

An initial slice-then-append candidate was rejected: its release binary
segfaulted, the full comparison reported `missing document`, and ASan found a
heap-buffer-overflow. Its apparently better Callgrind number was invalid and
is excluded from every result here. An intermediate snapshot validation was
also interrupted by moving its scratch artifact directory; the final snapshots
were rebuilt and the complete validation rerun. The final evidence uses only
those immutable snapshots.

# Throughput

The unchanged shared-success corpus is 4,952 stylesheets producing 510,298
UTF-16 output units per pass. Each side runs one pass in each of five rounds;
round order reverses on alternate rounds. No compilation or other test runs
concurrently with these measurements. Every count/output-unit check must pass;
the complete output bytes are independently validated. Native and Node include
startup, source loading and case decoding. Go uses a compiled overlay test
binary and measures only in-process formatting after decoding; that timing
advantage is explicitly retained from the preceding report. No warmed-server
or print-only claim is made. Timing variation is visible below; use medians.

| Side | Round 1 / s | Round 2 / s | Round 3 / s | Round 4 / s | Round 5 / s | Median / s |
|---|---:|---:|---:|---:|---:|---:|
| native baseline | 763 | 774 | 799 | 583 | 742 | 763 |
| native final | 1435 | 1976 | 1936 | 1286 | 1917 | 1917 |
| Node baseline | 1145 | 1383 | 1389 | 1411 | 1327 | 1383 |
| Node final | 2521 | 2545 | 2668 | 1142 | 2571 | 2545 |
| Prettier fork | 1326 | 1365 | 1404 | 794 | 1282 | 1326 |
| Prettier npm 3.9.6 | 1350 | 1368 | 1263 | 1175 | 1144 | 1263 |
| Go | 6120 | 6009 | 5430 | 5896 | 5295 | 5896 |

Measured native gain: **2.51x**. Go/native median ratio: **3.08x**.
The previous 758/5,870 observation is reproduced closely by the fresh
763/5,896 baseline/reference. Node source also improves, from 1,383 to 2,545/s;
this supports attribution to the port rather than a changed compiler.
Final native is 1.45x the fork's median and 1.52x npm Prettier's here.
Those are corpus-specific observations, not general formatter rankings.
All raw rounds and stdout checksums are in `performance/measurements.json`.

# Allocation and residual costs

Counters are separate instrumented runs of the exact same 341 sample inputs.
They count heap values and RC calls, not every internal malloc buffer.
Allocations equal frees and regions are zero in both runs; LSan is an
independent check, not inferred from these counts.

| Counter | Baseline | Final |
|---|---:|---:|
| Allocations | 5,003,718 | 1,714,919 |
| Frees | 5,003,718 | 1,714,919 |
| Retains | 6,513,586 | 6,408,812 |
| Releases | 9,900,442 | 6,626,382 |
| Peak live values | 140,777 | 140,840 |
| Regions | 0 | 0 |

Allocations fall **65.73%**; peak live values slightly increase, consistent
with retaining the global unit list. A lower peak-memory claim is not made.
Retains hardly change, so remaining ownership traffic is substantial.

These categories sum **self** costs for explicitly selected, disjoint function
names. They are attribution evidence, not estimates of achievable speedup:

| Category | Baseline self Ir | Final self Ir | Final % |
|---|---:|---:|---:|
| RC (`retain`, `release`, `let_go`) | 605,263,559 | 351,181,789 | 23.60 |
| Allocation (`adamic_allocate`, malloc/free/calloc/realloc, slab helpers, grown) | 1,349,805,575 | 216,007,141 | 14.52 |
| Map (`find`, `adamic_map_*`) | 164,193,759 | 164,193,759 | 11.04 |
| Strings (`adamic_string_*`, `adamic_utf8_*`) | 791,705,935 | 156,990,213 | 10.55 |
| Regex VM (`regex_execute`, `regex_run`) | 559,559,594 | 57,225,046 | 3.85 |
| Virtual dispatch (`adamic_virtual`) | 29,005,656 | 29,221,896 | 1.96 |

The allocation helper selection is `adamic_allocate`, `malloc`, `free`,
`calloc`, `realloc`, `take`, `give`, `deallocate`, `new_chunk`, `grown`.
Other libc internals and child destruction remain in the unclassified balance.
Top rows and full call graphs can be regenerated from the archived raw profiles.

# Compiler and runtime proposals

These are inferences from the observations above. None is implemented here,
and no wall-time gain is asserted without a new benchmark and byte gate.

1. **Runtime correctness first: shared-slice append.** In
   `internal/native/runtime/string_append.c`, the sole-reference reuse test
   uses `capacity - length >= added`. For a shared slice `capacity == 0` and
   `length > 0`, unsigned subtraction wraps. Require `capacity >= length`
   before subtraction, and ensure the storage is owned by the string. The
   proving program `gaps/7_shared_slice_append.ts` prints 1152 on Node; native
   ASan reports heap-buffer-overflow in `adamic_string_put`, via append.
   `TestSharedSliceAppendGap` holds that observed gap. After a runtime fix,
   convert this to a positive native/Node/leak regression, also retaining an
   alias to the owner's bytes to catch silent mutation. The safe CSS workaround
   already passes full byte and memory checks. This is a correctness proposal,
   not permission to trade memory safety for the rejected profile gain.

2. **Compiler: borrowed result lifetimes and scalar replacement.** The sample
   still executes 6,408,812 retains and 6,626,382 releases. Their direct self
   work plus `let_go` is 23.60% of all instructions. The generated C retains
   results of `Tree.at`, immediately reads them through helper methods, then
   releases them. Existing borrowed parameters/lent reads do not remove this
   owned-return traffic. Prove a borrowed result tied to an arena owner, and
   allow a caller to keep it borrowed while no intervening operation can erase
   it or release the owner. Do not infer safety from a readonly view alone:
   test owner replacement, callbacks, exceptions and writable aliases with
   use-after-free mutants. Temporary printer `Node` wrappers are another
   scalar-replacement candidate: their constructor costs 22,620,227 inclusive
   instructions (1.52%), before downstream release traffic. Preserve identity
   whenever a wrapper escapes or identity is observed. The RC self share is
   an upper bound on directly targeted work, not a promised 23.60% gain.

3. **Compiler: exact-class direct calls.** There are 3,652,737 profiled calls
   to `adamic_virtual`; its direct cost is 29,221,896 instructions (1.96%).
   `Tree.at` alone has 39,521,532 self instructions (2.66%). Generated C calls
   it through a function-pointer lookup even when the receiver comes from a
   concrete `new Tree`. Prove the exact dynamic class and an unmodified method
   target before emitting a direct call. This could expose small lookup/read
   helpers to clang inlining and borrow-result analysis; those secondary gains
   are unmeasured. Inheritance, overrides and aliased method changes must stay
   dynamically dispatched; use dispatch-target mutants to prove that boundary.

4. **Runtime/compiler: small constant-key property maps.** Final map self work
   is 164,193,759 instructions (11.04%), unchanged in absolute terms;
   `find` alone is 108,246,554 (7.28%), with 1,905,678 recorded call edges.
   `ObjectNode` uses separate Map tables for strings, numbers, booleans,
   objects and lists, plus ordered keys and nulls. Tiny tables still hash the
   UTF-8 key at each lookup. Evaluate inline small-map storage or a compiler
   specialization for confined constant-key maps before a larger typed-tree
   port rewrite. Preserve insertion order, deletion/reinsertion, missing versus
   empty/null, iteration under mutation, and NaN/-0 equality. A source-level
   schema redesign belongs to a separate port optimization with canonical
   trees as its oracle; replacing heterogeneous maps with fixed slots without
   preserving those observable properties is not justified by this profile.

5. **Runtime: regexp search/state buffers.** Native regex remains compiled to
   constant programs, but `regex_execute` tries starts one at a time and each
   `regex_run` allocates repeat/capture/state storage. Calls fall from 1,327,763
   to 148,800 after the port fixes; final execute/run self work is 3.85% and
   execute inclusive cost is 177,825,726 (11.95%). Consider one per-execution
   scratch buffer, test-only execution without result objects, and proved
   first-character/anchor filters. `regex_input` still allocates and fills a
   UTF-16 copy; its final inclusive share is only 0.74%, so that alone is a
   smaller target. Preserve first match, greedy order, captures, lookarounds,
   Unicode advancement, sticky/global lastIndex and instruction-limit behavior.
   Anchoring/filter/state-sharing mutants need independent V8 regex oracles,
   not this CSS corpus alone. Do not attribute all inclusive VM costs to malloc.

# Top instruction rows

The following are the top twenty named rows after same-name inline collapse.
Anonymous loader addresses and `(below main)` are omitted only from ranking;
the total includes every raw self cost. Quoted suffixed symbols are retained
as Callgrind reports them. Inclusive rows overlap.

## Baseline inclusive

| Function | Ir | % |
|---|---:|---:|
| `main` | 4,386,905,347 | 100.00 |
| `adamic_function_328_Printer_print'2` | 2,990,547,960 | 68.17 |
| `adamic_function_355_Documents_print` | 1,839,631,017 | 41.93 |
| `regex_execute` | 1,818,509,754 | 41.45 |
| `regex_run` | 1,737,639,520 | 39.61 |
| `adamic_regex_replace` | 1,691,620,205 | 38.56 |
| `adamic_function_328_Printer_print` | 1,512,904,539 | 34.49 |
| `adamic_function_326_Printer_sequence` | 1,509,111,265 | 34.40 |
| `adamic_function_326_Printer_sequence'2` | 1,234,441,420 | 28.14 |
| `adamic_function_133_byteSlice` | 1,171,974,355 | 26.71 |
| `adamic_function_292_nextEmpty` | 1,126,619,406 | 25.68 |
| `adamic_function_252_compose` | 923,495,103 | 21.05 |
| `free` | 765,957,102 | 17.46 |
| `adamic_release` | 699,695,436 | 15.95 |
| `adamic_function_291_hasNewline` | 584,979,191 | 13.33 |
| `malloc` | 486,319,474 | 11.09 |
| `adamic_function_251_nestedCSS` | 460,392,155 | 10.49 |
| `adamic_function_251_nestedCSS'2` | 460,351,864 | 10.49 |
| `adamic_function_359_closure'2` | 432,083,669 | 9.85 |
| `adamic_string_append` | 413,892,942 | 9.43 |

## Baseline self

| Function | Ir | % |
|---|---:|---:|
| `free` | 712,675,203 | 16.24 |
| `adamic_release` | 531,477,492 | 12.11 |
| `regex_run` | 499,680,946 | 11.39 |
| `malloc` | 440,650,307 | 10.04 |
| `adamic_allocate` | 182,227,715 | 4.15 |
| `adamic_string_append` | 175,591,012 | 4.00 |
| `adamic_string_join_halves` | 140,464,821 | 3.20 |
| `adamic_string_share` | 119,479,302 | 2.72 |
| `find` | 108,246,554 | 2.47 |
| `adamic_function_133_byteSlice` | 100,254,506 | 2.29 |
| `adamic_string_put` | 68,265,996 | 1.56 |
| `regex_execute` | 59,878,648 | 1.36 |
| `adamic_retain` | 55,880,366 | 1.27 |
| `adamic_string_allocate` | 49,097,048 | 1.12 |
| `adamic_string_free_index` | 48,539,106 | 1.11 |
| `regex_input` | 48,247,280 | 1.10 |
| `adamic_string_equal` | 41,170,995 | 0.94 |
| `adamic_function_7_Tree_at` | 39,521,532 | 0.90 |
| `adamic_object_free_children` | 34,243,742 | 0.78 |
| `adamic_string_units` | 31,560,484 | 0.72 |

## Final inclusive

| Function | Ir | % |
|---|---:|---:|
| `main` | 1,487,564,813 | 99.99 |
| `adamic_function_328_Printer_print'2` | 1,470,103,119 | 98.81 |
| `adamic_function_252_compose` | 923,171,385 | 62.05 |
| `adamic_function_251_nestedCSS` | 459,867,750 | 30.91 |
| `adamic_function_251_nestedCSS'2` | 459,838,509 | 30.91 |
| `adamic_release` | 399,214,697 | 26.83 |
| `adamic_function_328_Printer_print` | 347,597,928 | 23.36 |
| `adamic_function_326_Printer_sequence` | 343,793,200 | 23.11 |
| `adamic_function_229_parseValue` | 301,421,528 | 20.26 |
| `adamic_function_360_closure'2` | 270,119,864 | 18.16 |
| `adamic_function_245_calculateLoc` | 251,028,702 | 16.87 |
| `adamic_function_245_calculateLoc'2` | 243,123,414 | 16.34 |
| `adamic_function_360_closure` | 233,393,617 | 15.69 |
| `adamic_function_326_Printer_sequence'2` | 220,577,221 | 14.83 |
| `adamic_function_329_Printer_declaration` | 194,938,237 | 13.10 |
| `regex_execute` | 177,825,726 | 11.95 |
| `adamic_function_333_Printer_comma` | 171,427,225 | 11.52 |
| `regex_run` | 165,240,141 | 11.11 |
| `adamic_function_12_Tree_children` | 140,747,712 | 9.46 |
| `adamic_function_325_Printer_list` | 139,993,503 | 9.41 |

## Final self

| Function | Ir | % |
|---|---:|---:|
| `adamic_release` | 278,155,915 | 18.70 |
| `find` | 108,246,554 | 7.28 |
| `free` | 87,897,255 | 5.91 |
| `adamic_allocate` | 63,671,456 | 4.28 |
| `adamic_retain` | 55,120,830 | 3.70 |
| `malloc` | 50,530,924 | 3.40 |
| `regex_run` | 49,550,587 | 3.33 |
| `adamic_string_equal` | 41,170,995 | 2.77 |
| `adamic_function_7_Tree_at` | 39,521,532 | 2.66 |
| `adamic_string_units` | 31,443,885 | 2.11 |
| `adamic_object_free_children` | 30,296,277 | 2.04 |
| `adamic_virtual` | 29,221,896 | 1.96 |
| `adamic_map_set` | 17,963,856 | 1.21 |
| `let_go` | 17,905,044 | 1.20 |
| `adamic_function_12_Tree_children` | 14,449,691 | 0.97 |
| `adamic_string_slice` | 14,299,880 | 0.96 |
| `adamic_function_16_ObjectNode_clear` | 12,991,008 | 0.87 |
| `adamic_string_units_before` | 12,857,231 | 0.86 |
| `adamic_array_push` | 12,306,900 | 0.83 |
| `adamic_array_new` | 11,670,488 | 0.78 |

# Setup and reproduction

Read `CLAUDE.md`, scanner `PERFORMANCE.md` and its profile tooling, and JSON's
complete `PERFORMANCE.md` at fetched `codex/stage1-json-format`
`67796a5a8419d5052e1cb7532d6b39bb5777da6c`. JSON guidance is read from that
commit because its report is not in this branch. It is not merged.
`bash cloud/setup.sh` succeeds, then source `/workspace/adamic-tools/env.sh`:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (72s)
setup: done in 72s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc`: 5; cgroup: four cores. Go 1.27.1, clang 20.1.8, Node 24.19.0.
Callgrind was missing. `sudo apt-get update` and `sudo apt-get install -y
valgrind` both fail exactly with `sudo: command not found`. Workaround:
download Debian `valgrind_3.24.0-3_amd64.deb`, extract with `dpkg-deb -x`,
set `VALGRIND_LIB` to the extraction's `usr/libexec/valgrind`, and use its
`usr/bin/valgrind`. No system package or compiler/runtime change was needed.

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_CSS_FIXTURES=/tmp/adamic-css-prettier
export ADAMIC_CSS_LIBRARY=/tmp/adamic-css-library
export ADAMIC_CSS_PRINTER_LIBRARY=/tmp/adamic-css-printer-library
# Run artifact generation at the baseline commit, then at the final port source:
ADAMIC_CSS_PROFILE_DIR=/workspace/scratch/css-perf/baseline go test -v -count=1 -timeout=30m ./stage1/cohere/css -run '^TestCSSProfileArtifacts$' > /tmp/css-speed-baseline-artifacts.log 2>&1
ADAMIC_CSS_PROFILE_DIR=/workspace/scratch/css-perf/final go test -v -count=1 -timeout=30m ./stage1/cohere/css -run '^(TestCSSProfileArtifacts|TestSharedSliceAppendGap)$' > /tmp/css-speed-final-artifacts.log 2>&1
# Repeat this profile command for each immutable artifact directory:
VALGRIND_LIB=/workspace/scratch/css-perf/valgrind/usr/libexec/valgrind /workspace/scratch/css-perf/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/css-perf/final/callgrind.out /workspace/scratch/css-perf/final/profiled /workspace/scratch/css-perf/final/sample.txt count once > /workspace/scratch/css-perf/final/profile.stdout 2> /workspace/scratch/css-perf/final/profile.stderr
cmp /workspace/scratch/css-perf/final/profile.stdout /workspace/scratch/css-perf/final/sample-expected.txt
python3 stage1/cohere/css/profile.py --summarize /workspace/scratch/css-perf/final/callgrind.out > /workspace/scratch/css-perf/final/checked-summary.json
python3 stage1/cohere/css/profile.py --benchmark /workspace/scratch/css-perf/baseline /workspace/scratch/css-perf/final --prettier /tmp/adamic-css-printer-library --output /workspace/scratch/css-perf/measurements.json > /tmp/css-speed-timings.log 2>&1
ADAMIC_CSS_PROFILE_SNAPSHOTS=/workspace/scratch/css-perf/baseline:/workspace/scratch/css-perf/final go test -v -count=1 -timeout=45m ./stage1/cohere/css > /tmp/css-speed-gate.log 2>&1
go vet ./... > /tmp/css-speed-vet.log 2>&1
go test -count=1 -timeout=30m ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/regexp' > /tmp/css-speed-regex-oracle.log 2>&1
gofmt -l stage1/cohere/css > /tmp/css-speed-gofmt.log
git diff --check > /tmp/css-speed-diff.log
```

Artifact generation on the old baseline requires the new opt-in harness files
from this unit while leaving the six port source directories at baseline.
It snapshots them before compilation. Existing corpus fixture and npm setup
commands are in `REPORT.md` and `PRINTER_REPORT.md`. Shared corpus generation
requires cohere's bundled Prettier; the external npm oracle is pinned 3.9.6.
The same original PostCSS surrogate and full-Prettier API boundary records
remain exact exceptions; no new output exception is introduced.

# Verification and limitations

The complete touched CSS package **PASS**, 856.325s. Raw parser: 61.91s;
composed parser: 99.92s; optimized printer: 20.24s; optimized parser: 15.18s;
full sanitized printer and original-library checks: 493.58s; immutable
snapshot byte checks: 92.89s. The final artifact generation and standalone
shared-slice proof **PASS**, 38.321s. Repository `go vet ./...`, gofmt and
`git diff --check` have no diagnostics; filtered native regex oracle **PASS**,
0.238s. Raw archived profiles/C, their checked totals, identical sample and
manifest bytes all pass the archive check.

Both default and narrow options have 24,076 exact Go printer answers, including
4,966 formats and 19,110 refusals. Native ASan/UBSan, source Node and the
JavaScript backend agree; separate LeakSanitizer runs are clean. The composed
parser has 24,076 exact Go trees/errors on those backends, also leak-clean.
Ordinary `-O2`, profiled `-O2 -g`, counting and source Node snapshots for
**both baseline and final** produce exactly 3,172,898 default and 3,227,590
narrow Go answer bytes. This validates every byte of each measured executable,
not only its timing checksum.

Each option set against each Prettier variant retains 4,952 exact formats,
19,080 shared refusals and 44 exact recorded API-boundary occurrences:
BOM 12, CR 2, NBSP 2, YAML 28. Raw PostCSS retains 24,074 exact answers and
the two proved surrogate-cut occurrences. No new discrepancy is admitted.

Three printer source mutants compile and run to exit 0 with clean stderr;
then both native ASan/UBSan and Node comparisons catch them in both modes:

| Mutant | First difference |
|---|---|
| Omit declaration semicolons | Line 2, byte 69 |
| Omit rule-body indentation | Line 6, byte 7 |
| Ignore remaining width in document groups | Line 26, byte 12 |

Existing raw-parser mutants are also rerun on native and Node, and on composed
Node: removing custom-property block values (raw line 162, byte 396; composed
line 162, byte 1,001), dropping value comments (raw line 3,126, byte 302;
composed line 3,126, byte 77), and excluding the closing brace from its offset
(raw line 6, byte 659; composed line 6, byte 2,264). Each is caught by the
external Go comparison, not by a compile failure. Public Range corruption
still prints `caught` on raw/composed backends, including native and JavaScript,
with clean leak checks.

Memory-check mutants **PASS**, 53.64s. All three preserve the ordinary
optimized composed printer's bytes on the test inputs; only the intended
instrument catches each:

| Temporary generated-C mutant | Observation |
|---|---|
| Read freed storage | ASan heap-use-after-free |
| Overflow a volatile signed integer with argc | UBSan signed integer overflow |
| Omit generated releases | ASan/UBSan with leaks off and output checks pass; LSan catches 607,327 bytes in 8,466 allocations |

The first two are instrumentation probes; the last leaks actual composed
printer allocations. They do not deploy faulty source. The independent
Callgrind summary +1 mutant is also caught, as recorded above. The shared-slice
append program proves a real runtime defect, rather than an injected defect.

Logs are retained in `verification/speed-*.log`, including the rejected trim
candidate's ASan diagnostic, setup/package-install failure, profile accounting
mutant and all successful final commands. Bench and artifact generation are
opt-in and skipped in the ordinary package gate because their successful
separate runs are already recorded. All test output goes to files, never piped.
The complete repository test gate is not run: this unit runs the whole touched
CSS package, repository vet and a filtered native regex oracle.

CSS/SCSS scope is unchanged. Less fixtures use those two supported grammars;
Less semantics, YAML delegation, HTML/CSS-in-JS orchestration, private files,
incremental formatting and arbitrary invalid UTF-8 output remain unclaimed.
The profile is a bounded success sample, not a profile of malformed refusals,
and wall timings are whole parse-and-print passes on the shared corpus.
No compiler/runtime proposal is claimed implemented or measured as a speedup.

Port, harness and profile artifacts commit: `55890551505d1217a4443282e4474f956f4bc21b`.
Verification logs are committed in the following report commit; both are pushed
to `codex/stage1-css-printer` without a pull request.
