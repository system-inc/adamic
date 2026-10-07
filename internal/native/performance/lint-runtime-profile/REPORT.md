Built batch8 and the supplied batch4 TypeScript syntax driver on current main; scratch attribution includes integer fast paths and numeric Map hashing.
Runtime commits: 8b629af, b7059e0, 4ffed41, 0f58c62; fixture/count registration b02f762; clean integration tip 8c171be; report is the pushed profile branch tip.
Commands and outputs: best-of-five interleaved Go/native/Node, full byte parity, Linux perf call graphs, core.ts Callgrind, package gates; tables and logs below.
Mutants: release drain/decrement/children and wrong recorded counts caught by live/count checks; search offset/halves/miss and equality identity/bytes caught against Node.
Not covered: type-aware profiling belongs to the bridge worker by instruction; batch6 not rebuilt, full repository gate and other architectures not run; runtime branch stays from main, with no lint/emitter/map-hash merges.

# Native lint runtime profile

## Build flags and fresh release versus sanitizer comparison

These are new interleaved best-of-five, count-only measurements on the same pinned 77-file corpus. Before/after use identical generated C, integer fast paths a183e50 and numeric Map hash e7ea1a4; only the four runtime fixes differ. Each round rotates among Go, before release, before sanitized, after release and after sanitized. Both stages share the Go result from their group. Every timed execution returns exactly 161 findings (batch8) or 15,119 (batch4), with empty stderr. No compilation or other task tests were run concurrently with this timing sequence.

| Driver | Runtime fixes | Go gc/exe best s | Release -O2, no sanitizers s | Release / Go | Sanitized -O1, ASan + UBSan s | Sanitized / Go |
|---|---|---:|---:|---:|---:|---:|
| Batch8 syntax | before | 0.434827 | 2.569821 | 5.91x | 10.386376 | 23.89x |
| Batch8 syntax | after | 0.434827 | 2.537118 | 5.83x | 10.979861 | 25.25x |
| Batch4 syntax | before | 1.286065 | 7.714120 | 6.00x | 32.845238 | 25.54x |
| Batch4 syntax | after | 1.286065 | 6.701798 | 5.21x | 29.862276 | 23.22x |

Fresh best release times decrease 1.3% for batch8 and 13.1% for batch4; checked batch8 increases 5.7%, while checked batch4 decreases 9.1%. The earlier profiling campaign remains below with its own Go baseline and load.

The sanitized column is the complete checked build: `-fsanitize=address,undefined -fno-sanitize-recover=all`, with `ASAN_OPTIONS=detect_leaks=1` and `UBSAN_OPTIONS=print_stacktrace=1`. Sanitizers also select malloc/free rather than the release slab allocator. These measurements therefore compare build configurations, not sanitizer instrumentation alone. Neither release binary defines ADAMIC_COUNT, enables sanitizers, or uses LTO or architecture-specific flags. Debug symbols (`-g`) are present in both configurations and do not change the optimization level.

Exact expanded clang commands, source-unit order, working directories and binary Go metadata are in [flags-build-commands.txt](evidence/flags-build-commands.txt); [flags-builds.json](evidence/flags-builds.json) also records every generated/runtime C and header SHA-256. The compiler is `/workspace/adamic-tools/llvm/bin/clang`, clang 20.1.8, LLVM commit 87f0227cb60147a26a1eeb4fb06e3b505e9c7261, target x86_64-unknown-linux-gnu. Every command uses these exact common arguments, followed by the configuration arguments, `-o scanner` or `-o sanitized`, `main.c`, the recorded ordered runtime C files, and `-lm`:

```text
-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls
Release:   -O2 -g
Sanitized: -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all
```

Go is `go version go1.27.1 linux/amd64`, built with `go build -overlay=<overlay.json> -o <oracle> <virtual-main.go>` from the pinned cohere checkout. The overlay inserts the rule-selection driver, not a compiler/runtime change. This is default optimizing gc, `-buildmode=exe`, `-compiler=gc`, GOARCH=amd64, GOOS=linux, GOAMD64=v1, CGO_ENABLED=1; no race instrumentation, sanitizers, `-gcflags`, `-N`, or `-l`. Before/after use the byte-identical Go executable and manifest. Historical batch8 is regenerated from its own driver and uses the same Go version and build mode.

Machine and CPU quota are unchanged: Intel Xeon Platinum 8573C, Linux 6.18.44 x86_64, `nproc` 5, `cpu.max` 400000/100000 (four CPUs). No affinity or fixed frequency. Load strings below are `/proc/loadavg` (1/5/15-minute averages, runnable/total processes, last PID), not CPU utilization. Raw samples, rotation order, commands, output hashes and exit status are in `flags-group-*.json`. A measurement-guard mutant changes the expected batch8 count to 162; the first real Go scan returns 161 and the exact-output assertion rejects it (exit 1, `lint-flags-count-mutant.log.gz`). The production expectation remains 161.

| Interleaved group | Load before | Load after |
|---|---|---|
| 8 | 3.82 5.08 2.51 1/172 32842 | 1.26 3.39 2.27 1/173 33063 |
| 4 | 1.26 3.39 2.27 1/173 33065 | 1.01 1.51 1.76 1/174 33265 |
| original8 | 1.01 1.51 1.76 1/174 33266 | 1.00 1.40 1.70 1/178 33341 |

### Audit of the original comparisons

Read the branch-specific BATCH reports and their build harnesses, rather than the older shared REPORT.md. All three historical syntax comparisons used unsanitized release builds, so their roughly 5 to 7x gap was not a sanitizer artifact.

| Original branch/report | Reported native / Go s | Ratio | Native build evidence |
|---|---:|---:|---|
| batch8 4189abd, BATCH8.md | 2.898728 / 0.409678 | 7.08x | Says "Release native"; batch8Throughput calls batch8Build with Sanitize=false, and native.go maps false to -O2 without sanitizers. |
| batch6, BATCH6.md | 1.632241 / 0.311383 | 5.24x | Explicitly says "Native used clang -O2 without sanitizers". |
| batch4-typescript 63782c5, BATCH4.md | 4.546550 / 0.945173 | 4.81x | Explicitly says "Native uses unsanitized clang -O2". |

Batch8 does not spell out `-O2` in its report, so its driver was also regenerated at exactly 4189abd, with its original compiler and runtime, without the later integer/hash integrations or this unit's runtime fixes. Fresh interleaved best-of-five results:

| Original batch8 build | Native best s | Go gc/exe best s | Native / Go |
|---|---:|---:|---:|
| Release -O2, no sanitizers | 2.413668 | 0.392837 | 6.14x |
| Sanitized -O1, ASan + UBSan | 11.048057 | 0.392837 | 28.12x |

Batch6 and batch4 reports already name the native flags, so their original revisions were not rebuilt under the conditional request. The fresh before/after batch4 rows above cover both build configurations with current-main runtime and the required scratch integrations. Archived original reports and harness flag evidence are in `evidence/flags-history/`.

All native-versus-Go timings elsewhere in this report are **release -O2 without sanitizers**, unless explicitly labeled ASan/UBSan or quoted from the separate bridge reports. Checked full-output parity uses **-O1 with ASan/UBSan** and is not used as the throughput numerator. The bridge's historical 68.118s/7.422s build flags were not audited here because that path belongs to its profiling worker; that quoted result is not a syntax-driver measurement.

## Scope and reproduction

Runtime base is origin/main ef3d907. The separate worktree `/workspace/lint-runtime-drivers`, branch `codex/lint-runtime-drivers`, merges batch8 4189abd and batch4-typescript 63782c5, then integer fast paths a183e50 and numeric Map hashing e7ea1a4. Its integration head is ab6de95. This branch is not merged into the runtime branch. Existing generated C is fixed throughout each runtime comparison. Runtime overlays are only heap.c, string_build_impl.h, string_search_impl.h and unchanged string_slice_impl.h. Integer/map runtime files remain those of the scratch integration. No changes to region.c, parallel runtime files or emitter.

Read the earlier batch2 through batch6 and batch8 REPORTs and batch4 profiling reports before profiling. Current main already contains slab allocation, UTF-16 BMP views, uniform field layouts and field caches. Earlier port work already changed comment scanning, Unicode folding, position masks and some literal-array membership tests. Those improvements were retained.

Pinned corpus: TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8, all 77 `src/compiler/*.ts` files. Go cohere is the scratch recursive submodule, 715ba94, and typescript-go 8d550c8. Batch8 runs its nine-rule driver, `all` manifest rows; batch4 runs its 50-rule syntax driver. Timing uses `--count`, excludes compilation, includes process startup, reading, parsing and rule execution. Five interleaved Go/native/Node executions per snapshot, best of each reported. Full output parity is checked separately, because position formatting/serialization materially changes timing. Node uses the same port snapshot and parser, via `oracle/node.mjs`.

Machine: Intel Xeon Platinum 8573C, Linux 6.18.44 x86_64, `nproc` 5; CPU cgroup quota 400000/100000, equivalent to four CPUs. Clang 20.1.8, Go 1.27.1, Node 24.19.0. No CPU affinity or fixed frequency; load averages are retained per timing run, so wall-time changes alone are not proof of causation. Counts and Callgrind use separate untimed executables. No other task compilation/tests were deliberately run during timings/profiles.

`bash cloud/setup.sh` succeeded using `/workspace/adamic-tools/env.sh`. Main timing lines: go 0s, clang 1s, node 1s, submodules 1s, cache 247s, total 247s. Scratch: go 0s, clang/node/submodules 2s, cache 761s, total 761s. Logs are in evidence. Tool installation through apt lacked permission as uid1000; extracted perf 6.12.107 and Valgrind 3.24 locally. Linux perf works with the container's granted execution permissions, so full-input `perf record -F 999 --call-graph dwarf,16384` was used. Valgrind adds deterministic instruction counts on smaller input: only `src/compiler/core.ts`, 92,419 bytes, same rule selection as its driver.

Test, setup and parity logs are preserved as `evidence/*.log.gz` because the repository ignores raw `.log` files. Decompress with `gzip -cd FILE.log.gz`. Raw Callgrind and perf text reports are also compressed; JSON summaries and counters remain plain text.

Tools are preserved under `tools/`; they use the documented `/workspace` layout. Copy `tools/batch8_artifacts.go.fixture` into the scratch lint package as `runtime_profile_test.go`. For batch8 run `ADAMIC_TYPESCRIPT_SOURCE=/workspace/lint-typescript ADAMIC_LINT_PROFILE_DIR=/workspace/lint-baseline8 go test ./stage1/cohere/lint -run '^TestRuntimeBatch8Artifacts$' -count=1 -v`; batch4 uses the branch's existing `TestProfileArtifacts` with the analogous directory. Redirect test output directly into a log. Regenerate these artifacts after each scratch compiler integration. `lint-rebuild.py BASE TARGET REF` overlays a runtime revision; omit REF for working runtime, add `--sanitize` for ASan/UBSan. Then run `lint-measure.py TARGET bench`, `lint-callgrind.sh TARGET`, `lint-cg-summary.py TARGET`, `lint-instrument.py TARGET`, `lint-perf.sh TARGET`, and `lint-measure.py TARGET parity`, sequentially. Native flags are C11, warnings as errors, -O2, -ffp-contract=off and -fno-optimize-sibling-calls; profiling adds -g, sanitizers use -O1.

## Confirming the reported gap

Fresh main, before either extra scratch integration, gives batch8 161 findings in 2.434389s native versus 0.360729s Go: 66.1 versus 446.3 findings/s, a 6.75x gap. This confirms the reported scale (55 versus 393), not the exact throughput on a different machine/runtime. Native batch4 gives 15,119 findings in 7.196674s versus 1.160737s Go, a 6.20x gap. Node is 1.286111s and 3.234707s respectively. Initial load averages were high but falling: batch8 9.82/19.68/13.25 before and 7.14/18.26/12.95 after; batch4 7.14/18.26/12.95 before and 2.83/14.27/11.99 after. The one runnable-process field was low; load includes earlier work. Raw five samples are in JSON.

The supplied batch4-typescript driver is syntax-only, with no binder/checker. Its original REPORT's 68.674s is TestThroughput including builds; its timed syntax scan is 4.546550s native versus Go 0.945173s. That figure is distinct from the actual type-aware result. The user identified **68.118s native versus 7.422s Go** on `codex/tsgo-c-library` **0d540f4**. This unit leaves that path to the worker already profiling it. Batch6's earlier 295 versus 1,545 findings/s remains prior evidence, not a fresh run here; batch8 is the primary syntax walk attribution, with batch4 syntax as an additional runtime workload.

Read `stage1/cohere/typeaware/COVERAGE_REPORT.md`, `VOLUME_REPORT.md`, `VOLUME_PROFILE_REPORT.md` and `PROFILE_REPORT.md` at 0d540f4 to avoid duplicating bridge work. The coverage report records the new-ten-rule timing within the cumulative 26-rule bridge coverage: 2,222,043 adapter queries, mean 6.602 microseconds, summed query intervals 14.670s inside the 68.118s run. This is reported prior evidence, not remeasured here. Earlier six-rule profiling already removed eager TypeToString/fact names and reduced frame decoding allocations; the sixteen-rule profile already indexed immutable AST spans and shadow bindings. Its measured crossing/registry estimate was about 0.35s over 854,525 queries, not the whole native remainder. None of those bridge, fact-decoder or shadow-rule fixes were repeated. This unit's measurements and runtime changes concern the syntax drivers only.

Initial full output matches Go byte for byte: batch8 11,441,458 bytes on native, Node and ASan/UBSan; batch4 17,842,614 bytes on native and Node. Final native, Node and ASan/UBSan output matches Go at the same byte sizes, with empty stderr. Full output elapsed times are retained in parity logs, rather than conflated with count-only throughput. SHA-256 hashes are in `evidence/parity-sha256.json`.

## Runtime changes, one measured commit each

1. **8b629af, Skip empty release drains.** A non-last release previously entered and left the freeing loop even with an empty queue. Return before setting `draining` in that case. Child destruction remains iterative. Whole-input empty drain entries fall from 94,301,585 to zero in batch8 and 217,308,954 to zero in batch4. This does not eliminate ownership calls.
2. **b7059e0, Search backward by bytes for whole-character needles.** lastIndexOf previously materialized UTF-16 arrays for both strings before searching. A needle containing no surrogate half can be searched backward by UTF-8 bytes, with the matching byte offset translated to UTF-16 once. Half-surrogate needles retain the old unit search. Batch4 eliminates 2,450 temporary buffers and 341,788,809 decoded UTF-16 units. Batch8's chosen input has no lastIndexOf calls and gets no instruction benefit. The core.ts batch4 subset invokes it only twice, so its instruction delta understates the conversion cost seen in full-input perf and counters. Unsuccessful reverse search stays linear; overlapping whole-character matches and UTF-16 result indices are covered.
3. **4ffed41, Skip comparing bytes for identical string headers.** Check length first, then identity, before memcmp. Undefined cases retain their old semantics. Batch4 avoids 23,617,385 byte comparisons; batch8 avoids 2,316,956. The first pointer-first prototype added unnecessary tests for unequal-length strings; it was discarded. The length-first version still adds 226,458 total core.ts instructions in batch8 (0.092%), while reducing batch4 by 6,257,518 (1.10%). Its whole-input benefit is concentrated in batch4. This tradeoff is stated, not attributed to an across-the-board instruction win.
4. **0f58c62, Keep destruction off the common release path.** Keep the last-reference queue/destruction path in a noinline helper. Before, clang saved seven registers even for null/immortal/shared releases. After, those cases return without stack work or queue access. Count combined `adamic_release` plus `release_last`, not only the symbol that got smaller. Child callbacks still use `let_go`, and the queue remains nonrecursive and guarded by `draining`.

| Runtime stage | Driver | Native release -O2 best s | Go gc/exe best s | Node best s | core.ts instructions | Load before (1/5/15 min) |
|---|---|---:|---:|---:|---:|---|
| both | 8 | 2.434763 | 0.387216 | 1.612352 | 253,140,036 | 0.99 0.98 2.40 |
| both | 4 | 7.750185 | 1.216820 | 3.186131 | 583,380,074 | 1.14 1.02 2.37 |
| both-release | 8 | 2.327105 | 0.375348 | 1.327604 | 247,300,964 | 1.04 1.01 2.26 |
| both-release | 4 | 6.951565 | 1.182298 | 3.034863 | 571,021,695 | 1.29 1.07 2.25 |
| both-search | 8 | 2.576781 | 0.362114 | 1.461939 | 247,300,933 | 1.18 1.07 2.16 |
| both-search | 4 | 7.135446 | 1.185969 | 3.684197 | 569,174,756 | 1.37 1.13 2.15 |
| identity | 8 | 2.388713 | 0.366808 | 1.485478 | 247,527,391 | 0.77 0.83 1.59 |
| identity | 4 | 6.216240 | 1.212898 | 3.195342 | 562,917,238 | 0.85 0.85 1.57 |
| split | 8 | 2.236907 | 0.372147 | 1.548837 | 223,796,780 | 1.10 0.97 1.35 |
| split | 4 | 6.194127 | 1.207319 | 3.163925 | 513,645,338 | 1.06 0.98 1.34 |

The stage names correspond to main plus both upstream integrations, then each cumulative runtime fix. Every stage has identical findings. Wall time is noisy: for example search lowers deterministic batch4 instructions while that run's wall time rises. Do not infer a regression or speedup from that one wall observation alone. Final comparison and instruction evidence below separate these effects.

## Counters and upstream integrations

| Counter (full compiler corpus) | Batch8 before | Batch8 after | Batch4 before | Batch4 after |
|---|---:|---:|---:|---:|
| Allocations / frees | 8,203,664 / 8,203,664 | 8,203,664 / 8,203,664 | 26,387,731 / 26,387,731 | 26,387,731 / 26,387,731 |
| Retains | 110,222,423 | 110,222,423 | 268,796,884 | 268,796,884 |
| Releases | 100,087,461 | 100,087,461 | 241,069,081 | 241,069,081 |
| Empty drains | 94,301,585 | 0 | 217,308,954 | 0 |
| Drain entries (including empty) | 100,087,461 | 5,785,876 | 241,069,081 | 23,760,127 |
| UTF-16 temporary buffers | 0 | 0 | 2,450 | 0 |
| UTF-16 units decoded into temporaries | 0 | 0 | 341,788,809 | 0 |
| String memcmp calls | 13,875,561 | 11,558,605 | 35,250,576 | 11,633,191 |
| Map probes (upstream hash already included) | 1,387,198 | 1,387,198 | 5,873,485 | 5,873,485 |
| Peak live values | 730,193 | 730,193 | 746,217 | 746,217 |

| core.ts release instruction accounting | Batch8 before split | Batch8 after split | Batch4 before split | Batch4 after split |
|---|---:|---:|---:|---:|
| adamic_release | 45,815,393 | 11,940,690 | 109,139,504 | 27,444,258 |
| release_last | 0 | 10,081,639 | 0 | 32,362,812 |
| total_instructions | 247,527,391 | 223,796,780 | 562,917,238 | 513,645,338 |
| release + destruction self instructions | 45,815,393 | 22,022,329 | 109,139,504 | 59,807,070 |

The allocation counters count Adamic heap values, not raw malloc/realloc or temporary UTF-16 buffers. Eliminating those buffers therefore does not change allocation totals. All allocated values are freed on these runs. No region allocation was introduced. The field cache is already effective: core.ts object_find calls are only 268/357, while shape-fast-path access stays frequent.

Credit e7ea1a4 to the map owner: on batch4 whole input, identical 7,151,132 lookups take 30,964,006 probes before its numeric hash and 5,873,485 after, an 81.0% drop. Batch8 probes drop only 1,390,679 to 1,387,198 on 1,199,004 lookups. Some lookups encounter an empty table and take zero probes. A local duplicate hash prototype was discarded when this upstream input arrived; no hash implementation or hash test is in the runtime branch. Integer a183e50 is also included before attribution. Remaining core.ts fmod calls are 80 and 972 instructions (about 0.0004%/0.0002%); number formatting is 441 calls and about 28,500 instructions (0.011%/0.005%). These are not the current bottleneck, and integer %/ToInt32 work was not redone.

## Attribution before and after

Self perf samples on the full 77-file corpus, after both upstream integrations, before and after this unit's runtime changes. Columns identify the principal location of each measured cost; they do not prove how much of the wall-time gap would disappear under a hypothetical rewrite. Runtime ownership machinery is called according to the generated ownership schedule, so those causes interact. Percentages are rounded, samples are finite, and functions absent from the report mean zero observed samples, not proof of zero possible execution. No exception raise/check routines were found in these valid-input driver paths or sampled.

| Driver | Function (self samples) | Runtime | Generated shape | Program |
|---|---|---:|---:|---:|
| 8 | release and destruction | 15.63 → 8.87% |  |  |
| 8 | retain | 8.02 → 7.42% |  |  |
| 8 | allocate | 2.24 → 2.40% |  |  |
| 8 | string equality | 4.12 → 5.56% |  |  |
| 8 | UTF-16 unit walker | 0.22 → 0.23% |  |  |
| 8 | string locate/index | 0.48 → 0.65% |  |  |
| 8 | string units before | 0.00 → 0.00% |  |  |
| 8 | lastIndexOf | 0.00 → 0.00% |  |  |
| 8 | string slice | 0.40 → 0.46% |  |  |
| 8 | Map find | 1.03 → 1.48% |  |  |
| 8 | field cache miss | 0.00 → 0.00% |  |  |
| 8 | field cache access | 0.00 → 0.00% |  |  |
| 8 | field write and frozen check | 5.04 → 7.38% |  |  |
| 8 | closure allocation | 0.04 → 0.00% |  |  |
| 8 | number formatting | 0.04 → 0.11% |  |  |
| 8 | fmod | 0.00 → 0.00% |  |  |
| 8 | exceptions | 0.00 → 0.00% |  |  |
| 8 | parent array callback |  | 0.04 → 0.00% |  |
| 8 | per-node suffix callbacks |  | 0.00 → 0.00% |  |
| 8 | virtual dispatch |  | 5.77 → 7.61% |  |
| 8 | Parser.node |  | 4.82 → 5.37% |  |
| 8 | Scanner.code |  | 5.00 → 5.14% |  |
| 8 | Scanner.scan |  |  | 3.24 → 3.46% |
| 8 | node visitor |  |  | 2.28 → 2.59% |
| 8 | private-name used scan |  |  | 0.00 → 0.00% |
| 4 | release and destruction | 13.53 → 9.40% |  |  |
| 4 | retain | 6.95 → 8.59% |  |  |
| 4 | allocate | 2.65 → 3.31% |  |  |
| 4 | string equality | 6.04 → 7.39% |  |  |
| 4 | UTF-16 unit walker | 9.40 → 0.67% |  |  |
| 4 | string locate/index | 0.30 → 0.27% |  |  |
| 4 | string units before | 0.01 → 2.68% |  |  |
| 4 | lastIndexOf | 1.76 → 0.01% |  |  |
| 4 | string slice | 0.42 → 0.53% |  |  |
| 4 | Map find | 1.22 → 1.35% |  |  |
| 4 | field cache miss | 0.00 → 0.00% |  |  |
| 4 | field cache access | 0.00 → 0.00% |  |  |
| 4 | field write and frozen check | 3.15 → 3.48% |  |  |
| 4 | closure allocation | 0.15 → 0.16% |  |  |
| 4 | number formatting | 0.01 → 0.00% |  |  |
| 4 | fmod | 0.00 → 0.00% |  |  |
| 4 | exceptions | 0.00 → 0.00% |  |  |
| 4 | parent array callback |  | 0.00 → 0.00% |  |
| 4 | per-node suffix callbacks |  | 0.14 → 0.25% |  |
| 4 | virtual dispatch |  | 3.97 → 5.63% |  |
| 4 | Parser.node |  | 4.26 → 5.24% |  |
| 4 | Scanner.code |  | 4.12 → 4.90% |  |
| 4 | Scanner.scan |  |  | 3.09 → 3.92% |
| 4 | node visitor |  |  | 1.95 → 2.75% |
| 4 | private-name used scan |  |  | 0.00 → 0.00% |

Go comparison profiles run ten full scans per driver for more samples, with zero lost samples. Batch8 self samples include map hashing 6.52%, mapaccess2 6.01%, scanner.Scan 5.09%, scanner.scanIdentifier 4.24%, ComputeECMALineStarts 3.82%, and AST ForEachChild 2.61%. Batch4 includes small-map lookup 10.92%, hashing 9.64%, mapaccess2 7.40%, AST ForEachChild 3.20%, comment collection 3.07%, and scanner.Scan 2.72%. These profiles show Go also pays parsing/map/GC costs, rather than establishing that every native runtime operation is pure overhead. Compressed self/call-graph reports are retained.

Instruction counts strengthen attribution where a small self-sample percentage is unreliable. On core.ts, retain is 1,657,721 calls / 13,254,063 instructions in batch8 and 3,737,163 / 30,311,427 in batch4. Virtual dispatch is 1,182,132 / 9,457,056 and 2,019,360 / 16,154,880. After these fixes these calls remain exactly the same. The UTF-16 locate/index work is 18,604 calls / 1,943,887 instructions and 29,813 / 3,505,858; units_next remains needed for slicing/walking even after removal of lastIndexOf conversion. String slicing and indexing are measured runtime costs, but the existing main BMP view/index changes are retained rather than duplicated.

### Evidence for @system_adamic_compiler

No compiler emitter change is made here. Generated Parser.node returns `adamic_retain` for every access; visitor emission retains node children and array temporaries and releases them afterward. Whole-input release classification shows only 5.8%/9.9% are last-reference events. The fast runtime reduces each redundant/no-op call's cost, but cannot remove semantically scheduled retains/releases. Borrowing/local ownership analysis is the remaining opportunity, with the deep shared-owner and sanitizer controls maintained.

`adamic_virtual` resolves a method on almost every node/character access. Generated Scanner.code invokes it through parser objects; known receiver/type devirtualization could remove these dispatches. Object write paths also repeatedly call freeze/write checks; dropping runtime checks without a compiler proof would change behavior.

The main recursive batch8 visitor calls itself directly; it is not a closure on every visit. The parent-array setup `parser.nodes.map(() => -1)` does call generated `adamic_function_225_closure` 13,224 times on core.ts, roughly the node count (13,214 visitor calls). That closure's body only returns -1, with stack checking and indirect array-map dispatch. This is specific evidence for pure callback inlining, not a claim about every visitor.

Batch4's no_explicit_any, no_this_alias and triple_slash_reference run `['.ts','.tsx','.mts','.cts'].some(suffix => ctx.parser.path.endsWith(suffix))` on every visited node before their kind gate. Generated closures 227/245/248 each execute 13,214 times on core.ts; each visit also constructs the suffix array and capture cell/closure. Whole-rule file-extension hoisting is the port owner's simplest fix; callback inlining and escape analysis are compiler opportunities. The core.ts closure_new count is 42,439 versus batch8's 983. The three suffix callback bodies alone take 1,228,902 self instructions, 0.24% of final core.ts instructions, excluding array/capture allocation and dispatch. Evidence locations are generated main.c lines 28662/28879, 30356/30733 and 30772/30881 in the saved scratch drivers.

### Evidence for the port owner

The port's Context.kind returns a string through parser.node, while Go listeners use ast.Kind enum values. The generated access retains the node and kind string, then releases temporaries. This representation difference is an observed source/code shape; an enum/tag port is a possible improvement, not a measured speedup here.

no-unused-private-class-members loops over every private name, scans every sibling for accessor status, then recursively scans every sibling body via `used`. It does not even break the sibling loop when a read was found. Go collects declared names in a map and marks references during one body traversal. The port is O(private names × body size), versus Go's one pass, outside nested-shadowing details. This is algorithmic work that runtime speed cannot cure.

| Private names | Native control s | Native private rule s | Go private rule s | Node private rule s |
|---:|---:|---:|---:|---:|
| 100 | 0.003718 | 0.021870 | 0.005849 | 0.176717 |
| 200 | 0.004904 | 0.078179 | 0.005884 | 0.194553 |
| 400 | 0.008081 | 0.313543 | 0.007021 | 0.245622 |
| 800 | 0.013861 | 1.216100 | 0.007916 | 0.440260 |

This synthetic class has n private fields and n public methods, each returning its corresponding field; all three drivers report zero findings. The unrelated react/no-find-dom-node rule provides a parser/visitor control on the identical input. Best of five including startup makes the smallest points less reliable; the large-input divergence is the useful observation. All generated synthetic sources are `.a`. No `_used` calls were observed on core.ts or sampled on the compiler corpus; private-name examples there are comments/docs. Therefore this quadratic loop is not claimed to explain the measured whole-compiler gap. Replacing it belongs to the port owner, preserving nested-name shadowing, accessors and write-only use semantics pinned by Go.

## Validation and mutants

All commands sourced `/workspace/adamic-tools/env.sh` and wrote test output directly to the named log:

```sh
go test ./internal/native -count=1 -timeout 30m > /tmp/lint-native-package.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestRuntimeLastIndexOfMatchesNode|TestWeakReadsUndefinedOnceFreed|TestNativeAgreesWithNode/internal/oracle/testdata/(runtime_last_index_of|search_halves|search_from_sweep|shared_slices|shared_slice_append|strings|lone_surrogates|long_chain|weak_parent|doubly_linked|exceptions|map_iteration|sort_releases|class_oct6_release)\.a$' -count=1 -timeout 30m > /tmp/lint-filtered-oracle.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/lint-counts-update.log 2>&1
go vet ./internal/native ./internal/oracle > /tmp/lint-vet.log 2>&1
gofmt -l internal/native/runtime_profile_test.go internal/oracle/runtime_last_index_test.go internal/oracle/oracle_test.go > /tmp/lint-gofmt.log
git diff --check > /tmp/lint-diff-check.log
```

Full native package: PASS, 137.073s. Uncached relevant oracle: PASS, 5.281s. Counts update: PASS, 21.765s, only the new fixture row added (714 allocations/frees, 76 retains, 731 releases, peak 6, regions 0). Vet, formatting and diff checks: exit zero, empty logs. The complete repository gate, cohere lint/format gate, and other platforms were not run. Final controls: release paths PASS, 0.475s; full recorded counts PASS, 20.671s (`go test ./internal/native -run '^TestRuntimeReleasePaths$' -count=1` and `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1`). The wrong-count mutant ran all 305 countable fixture cases and failed on the one changed row. All mutant files were restored; git diff confirms only the new fixture count row remains.

| Mutant | Check that caught it | Observed failure |
|---|---|---|
| Skip empty-drain fix's last-reference drain | TestRuntimeReleasePaths, explicit live count | last release left 1 value |
| Decrement reference count twice | TestRuntimeReleasePaths, shared-value live count | shared release left 0 values |
| Return raw byte search offset | TestRuntimeLastIndexOfMatchesNode | stdout differs from Node |
| Treat surrogate half as whole-character bytes | TestRuntimeLastIndexOfMatchesNode | stdout differs from Node |
| Return zero for absent lastIndexOf match | TestRuntimeLastIndexOfMatchesNode | stdout differs from Node |
| Invert equality header identity | TestRuntimeStringEquality | distinct equal-length strings compare true, Node says false |
| Bypass equality memcmp | TestRuntimeStringEquality | same false equality |
| Skip final release_last helper | TestRuntimeReleasePaths | last release left 1 value |
| Double decrement in final common release path | TestRuntimeReleasePaths | shared release left 0 values |
| Skip object children in final destruction path | TestRuntimeReleasePaths, deep-chain live count | chain release left 199,999 values |
| Change new fixture's recorded allocation count to 715 | TestCountsAreRecorded | recorded 715 versus measured 714 allocations |

Search mutations compile and execute: returning byte offsets, searching surrogate halves as ordinary bytes, and returning zero for absence each disagree with Node. Equality mutations invert the header identity condition or bypass memcmp and produce a false true for distinct equal-length heap strings, caught against Node. Release mutations skip last-reference draining or decrement twice; explicit live-count assertions catch leaked pending values or prematurely freed shared values. The initial uncounted skipped-drain control **survived**: the pending free queue kept values reachable to LeakSanitizer. The control was strengthened with runtime live counts, rerun, and killed the same mutant. This survivor and its correction are retained in logs. Compile-time rejection is not counted as a killed semantic mutant.

The lastIndexOf test exercises empty/absent/overlong needles, repeated ASCII, multibyte BMP characters, paired and lone surrogates, embedded NUL, every substring of its pieces, and sliced heap strings. Node, release, ASan/UBSan and the JavaScript backend agree on 758 output bytes. The release control covers null, immortal, shared and final references and a 100,000-object chain with heap strings, in normal and sanitizer modes.

## Clean runtime integration branch

`codex/lint-runtime-fixes` is based on current origin/main e011f8f60899586d6373a5ccb07335ad82cfbf3c and pushed at **8c171be2163f8bad0410e7061520812f766b863d**. It contains exactly the requested five cherry-picks, eight changed runtime/test/fixture files, and no profiling artifacts, lint batches, integer-fast-path or map-hash integrations.

| Profile source commit | Integration cherry-pick |
|---|---|
| 8b629af | a24561a72c5a8358c24d5f4953adb6f8259ed6f6 |
| b7059e0 | 77b48689c90bca7430757acd6a34e747af0923c2 |
| 4ffed41 | 59a085087eefb2c3bb3f0c2cb2874784449d0506 |
| 0f58c62 | cc0c246e4e0f590182966623922dbfef7455a0e6 |
| b02f762 | 8c171be2163f8bad0410e7061520812f766b863d |

Toolchain setup on this branch: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 139s, total 139s. `nproc` is 5; cgroup cpu.max is 400000/100000. Every command sourced `/workspace/adamic-tools/env.sh`; output went directly to a log:

```sh
ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh > /tmp/lint-fixes-setup.log 2>&1
go test ./internal/native -count=1 -timeout 30m > /tmp/lint-fixes-native.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m > /tmp/lint-fixes-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestRuntimeLastIndexOfMatchesNode$' -count=1 -timeout 30m > /tmp/lint-fixes-search.log 2>&1
```

Native package PASS 126.350s; all recorded counts PASS 21.400s; uncached search comparison against Node PASS 0.669s. No count-table or source adjustment was needed. `git diff --check` and working-tree status are clean. Existing semantic mutants and their failures are recorded above; these cherry-picks introduce no additional checks. The complete repository test gate was not rerun. Logs and exact source-to-integration commit mapping are preserved under `evidence/lint-fixes-*`; they stay on the profile branch.

## Remaining gap

The following times belong to the original profiling campaign. The fresh build-flags comparison near the top is a separate interleaved run of the same snapshots and reports its own Go baseline and load.

| Final driver | Findings | Native s / findings per s | Go s / findings per s | Node s | Gap to Go | Native improvement vs both-integrated baseline |
|---|---:|---:|---:|---:|---:|---:|
| 8 | 161 | 2.236907 / 72.0 | 0.372147 / 432.6 | 1.548837 | 6.01x | 8.1% |
| 4 | 15,119 | 6.194127 / 2440.9 | 1.207319 / 12522.8 | 3.163925 | 5.13x | 20.1% |

| Driver | Runtime remains | Generated shape remains | Program remains |
|---|---|---|---|
| Batch8, 6.01x Go | Release/destruction 8.87%, retain 7.42%, equality 5.56%, allocation 2.40%; core.ts total instructions down 11.6% from integrated baseline | Virtual dispatch 7.61%, Parser.node 5.37%, Scanner.code 5.14%; 110.2M retains/100.1M releases still scheduled; parent-array closure per node | Parser/node representation, per-rule node inspection and literal membership arrays; quadratic private-member loop absent from corpus samples |
| Batch4 syntax driver, 5.13x Go | Release/destruction 9.40%, retain 8.59%, equality 7.39%, allocation 3.31%, UTF-16 units_before 2.68%; core.ts total instructions down 12.0% | Virtual dispatch 5.63%, Parser.node 5.24%, Scanner.code 4.90%; 268.8M retains/241.1M releases; suffix closures/arrays constructed per node | Hoist three file-extension tests per file; repeated rule passes and parser representation; private-member quadratic cost is a separate synthetic finding |

These categories overlap causally: the port chooses traversals, emission chooses temporaries/calls, and the runtime implements them. The observations above explain where native work remains; assigning exact fractions of the native-minus-Go wall gap would require comparable rewrites, not just a profile. No algorithmic improvement is credited on a corpus where its loop does not run.
