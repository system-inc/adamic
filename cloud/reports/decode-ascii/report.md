Added the ASCII cache: units = length + 1 exactly when the scan consumes the complete input; otherwise units stays 0.
Branch codex/decode-ascii, extending 4eb5187; string-views ba9c9ef tested as a clean scratch merge, compiler fixed.
Both runtime variants passed 95,694,560 cases per target with native ASan/UBSan and WASI; all requested gates passed.
False-ASCII mutant failed UTF-16 indexing (195 instead of 233); omitted-cache mutant failed metadata; all six mutants caught.
Best standalone Wasm/native: 9,235/11,917 req/s; merged: 13,804/19,753; the cache flag’s isolated speedup was not measured.

**Cache follow-up**

The addition to input.c is three lines after allocation: only `ascii == length` sets `string->units = length + 1`. The ASCII prefix scan already establishes that every input byte is < 0x80. That includes empty input and NULs; decoded length equals input length on that path, so this is exactly the representation’s ASCII flag. Non-ASCII paths retain new_string’s units = 0. No other runtime source was modified on the pushed branch.

The decoder test now independently classifies every corpus input by its bytes, asserts that all-ASCII results have units == decoded length + 1, and asserts that every non-ASCII result is unflagged and still has units = 0. A separate fresh decode of `é😀` checks charCodeAt results E9/D83D/DE00 before the metadata assertions. This holds the cache to JavaScript UTF-16 indexing, not just to a stored number.

Standalone final test: native ASan/UBSan 175.23s, WASI 42.44s, total 227.387s. Cleanly merged runtime: native ASan/UBSan 174.69s, WASI 37.52s, total 229.543s. Every target/variant printed `prefixes=92014 cases=95694560 Node and baseline identical`. The bounded longer-prefix coverage remains as described in the historical section below.

The false-flag mutant changes only the flag condition to true, setting length + 1 on non-ASCII input. Its first failure was `decode indexing mismatch index=0 actual=195 expected=233`: byte C3 instead of code unit E9. It failed before any metadata assertion, with no sanitizer error. An omission mutant disables cache initialization and fails on empty ASCII input: `decode cache mismatch record=0 run=0 offset=0 ascii=1 units=0 bytes=0`. The original byte-80, one-byte-late and dropped-byte mutants now fail the stronger cache assertions; the word over-read still fails with ASan. All six compiled successfully and exited 1 at their intended checks. Failure excerpts are in cache-mutants.json.

The new object comparison again compiled all 48 native runtime objects at -O2. **47/47 non-input objects remain byte-identical** on this branch, with only input.o different; cache-objects.json preserves hashes. This comparison concerns the standalone code change, not the extra runtime changes in the scratch merged variant.

**Combined runtime and rates**

Fetched exact string-views commit `ba9c9ef87e38d5369e0eaa1b0b6d24f93bbd227a`. A conflict-free `git merge-tree` with the prior decode commit produced tree `77f739c8f0ea4061fcfc8aa5b6596f67c50bf968`; the pending cache version of input.c was overlaid because string-views has no input.c delta. The four additional runtime files are directory.c, string_search_impl.h, string_share.c and string_slice_impl.h. There are no compiler changes in this combination. This merged runtime was tested and measured in scratch; its extra files are not changes on this worker branch. cache-merge.json records the tree, input hash and scope.

One quiet interleaved best-of-five run, same accepted 100,000 requests and compiler as the original report. Both standalone and combined native/Wasm variants verified every response byte against Node outside timing; timed loops checked checksum 7,394,547. Wasm used one instance per variant. Native timing brackets only the warmed serve loop; whole-process times including read/split and warmup are retained in cache-results.json. No other build, test or analysis ran during the rounds. Machine: AMD EPYC 9V74, Debian 13, Node 24.19.0, clang 20.1.8, WASI SDK 27; nproc 5 and cgroup CPU quota four CPU equivalents.

| Round | Order (SW/CW/SN/CN = standalone/combined Wasm/native) | Load before / after | nproc before / after | SW req/s | CW req/s | SN req/s | CN req/s |
|---|---|---:|---:|---:|---:|---:|---:|
| 1 | SW, CW, SN, CN | 1.00 / 1.00 | 5 / 5 | 9,191 | 13,804 | 11,825 | 19,753 |
| 2 | CN, SN, CW, SW | 1.00 / 1.00 | 5 / 5 | 9,093 | 13,203 | 11,917 | 19,667 |
| 3 | CW, SN, SW, CN | 1.00 / 1.00 | 5 / 5 | 9,235 | 13,537 | 11,373 | 19,051 |
| 4 | SN, SW, CN, CW | 1.00 / 1.00 | 5 / 5 | 8,697 | 13,727 | 11,627 | 18,147 |
| 5 | SW, CN, CW, SN | 1.00 / 1.00 | 5 / 5 | 9,106 | 12,225 | 11,779 | 18,770 |

| Runtime | Best Wasm req/s (100k serve time) | Best native req/s (100k serve time) |
|---|---:|---:|
| ASCII cache alone | 9,234.60 (10.828841s) | 11,916.60 (8.391659s) |
| ASCII cache + string-views | 13,803.93 (7.244314s) | 19,753.13 (5.062488s) |

Observed: the combined runtime was faster in all five rounds for both targets; best rates increased 49.5% for Wasm and 65.8% for native relative to standalone. Both variants include the cache change. This is the combined result of the additional runtime changes, not a measurement of the flag’s isolated contribution. The older scan-only rates (8,610 Wasm, 11,732 native) were a separate run; no controlled cache-only speedup is inferred from them. No new CPU profile was captured for this addition.

**Follow-up validation and reproduction**

Logs: cache-only-oracle.log, cache-combined-oracle.log, cache-native-gate.log, cache-wasi-gate.log, cache-wasi-integration.log, cache-mutants.log, cache-vet.log and cache-gofmt.log, all under /tmp/decode-ascii. Uncached native/input oracle passed in 110.440s; WASI oracle passed in 167.694s; WASI integration passed in 28.896s. Vet, gofmt and diff checks passed. No full repository gate was run.

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_TEST_WASI=1 go test ./internal/native -run "^TestDecodeASCII$" -count=1 -v -timeout 30m > /tmp/decode-ascii/cache-only-oracle.log 2>&1
python3 internal/native/decode_ascii/mutants.py > /tmp/decode-ascii/cache-mutants.log 2>&1
git fetch origin codex/string-views codex/wasm-requests-profile
bash cloud/reports/decode-ascii/build.sh > /tmp/decode-ascii/cache-prepare.log 2>&1
python3 cloud/reports/decode-ascii/cache-build.py > /tmp/decode-ascii/cache-build.log 2>&1
ADAMIC_DECODE_RUNTIME=/tmp/decode-ascii/cache-combined-runtime ADAMIC_TEST_WASI=1 go test ./internal/native -run "^TestDecodeASCII$" -count=1 -v -timeout 30m > /tmp/decode-ascii/cache-combined-oracle.log 2>&1
# After all other work finishes:
ADAMIC_DECODE_RESULTS=cloud/reports/decode-ascii/cache-results.json node --disable-warning=ExperimentalWarning cloud/reports/decode-ascii/measure.mjs measure > /tmp/decode-ascii/cache-measure.log 2>&1
```

In cache-results.json, “before” means standalone cache and “after” means cache plus string-views; its variants field records this explicitly. Full measured artifact hashes are included. The new preparation script reconstructs the merged snapshot without editing the checkout.

**Historical scan-only unit at 4eb5187**

The following records the original scan-only implementation and experiment before cache initialization was added. Its commands and artifact hashes describe that commit; the cache follow-up above describes the current code.

Implemented a bounded leading-ASCII copy in input.c for every target, preserving the existing tail decoder.
Base 4d86c305 (area/runtime); claim 54ec0b1 pushed first on codex/decode-ascii; only input.c changes in runtime.
Native ASan/UBSan and WASI pass 95,694,560 cases each; requested gates pass; 47 other objects byte-identical.
Four mutants caught: byte 0x80, one-byte-late boundary, dropped last byte by output comparison; word over-read by ASan.
Best rates: Wasm 8,376 -> 8,610 req/s, native 11,808 -> 11,732; decode after 1.96% self time; bounded corpus limits below.

**Change and correctness argument**

Base: `4d86c305dda261768b35687d199c1b2188c7ab71`, fetched explicitly from `origin/area/runtime` because this clone normally fetches only main. Claim-only commit: `54ec0b1`. The implementation is local to `decode` and a new private `ascii_prefix` helper in `internal/native/runtime/input.c`. No header, compiler, emitter, ownership, string cache or other runtime source changed.

The helper loads eight-byte words with `memcpy` only when `length - offset >= sizeof(uint64_t)`. The repeated high-bit mask is endian independent; memcpy supports unaligned pointers without violating aliasing or alignment rules, including wasm32. The scalar tail stops at the first byte >= 0x80. Offsets remain <= length throughout, so neither loop reads past the buffer. Empty inputs cause no input load or copy.

Every scanned byte is ASCII, including NUL, and WHATWG decoding preserves those bytes one for one. Both original passes start at the exact first non-ASCII byte; sizing begins with the prefix length and that prefix is copied directly into the same newly allocated string representation. Allocation, ownership, capacity, UTF-16/cache metadata, BOM handling and the rest of decode_step/encoding are unchanged. Therefore the original replacement behavior on the remaining suffix is preserved; a high-bit byte is never copied by the fast path.

**Oracle and corpus coverage**

`internal/native/decode_ascii/baseline.c` snapshots the original decoder portion of input.c from the runtime base, with only the exported entry point renamed. It remains a separate translation unit with the old implementation. `corpus.mjs` uses Node `Buffer.toString("utf8")`, re-encodes the result as UTF-8 and emits binary oracle records. For every record it additionally compares the full ASCII-wrapped byte input to Node for each run length 0..64, with and without a trailing ASCII run. Pointer alignment does not change the input bytes or the Node expected output.

`probe.c` expands each of 92,014 records at allocation offsets 0..7, ASCII prefix runs 0..64, and both suffix variants. That is 95,694,560 runtime calls per target, plus an equal number of original-decoder calls. Every result is compared byte for byte to Node and the original decoder. Each input allocation ends exactly at input[length], so native ASan detects tail word over-reads. Outputs and inputs are released on both success and comparison failure.

Coverage: all 256 one-byte and all 65,536 two-byte strings; all 256 leads in three- and four-byte sequences with each continuation position independently swept over `00,7f,80,8f,90,9f,a0,bf,c0,ff`; full products of those continuation classes for representative valid lead classes E0/E1/ED/EF and F0/F1/F4; truncations of valid, overlong/out-of-range/surrogate forms; BOM; NUL runs; a 4,096-byte all-ASCII input; and 10,000 fixed-seed random sequences of lengths 0..64 (seed `0xdec0de`). Invalid leads are exhausted by the one-/two-byte sweep and repeated in the longer decision-boundary sweeps.

The phrase “all prefixes” is interpreted as this structured coverage, not every arbitrary three- and four-byte string crossed with every placement. Exhausting four-byte strings at all 520 run/alignment combinations alone exceeds two trillion cases. This limitation was raised during the unit and is explicit here; there is no claim of exhaustive arbitrary four-byte enumeration.

The final new test completed: native ASan+UBSan 139.97 seconds; wasm32 under node:wasi 35.92 seconds; entire test 186.32 seconds including Node corpus preparation and builds. Both printed `prefixes=92014 cases=95694560 Node and baseline identical`.

**Object identity**

Before editing and after editing, `objects.py` compiled every native runtime `.c` translation unit at strict C11 `-O2 -ffp-contract=off -fno-optimize-sibling-calls` with the same clang and source paths. It compared the entire ELF object bytes, not disassembly alone. Result: **47/47 non-input objects byte-identical**, out of 48 total objects; only `input.o` differs. `objects.json` preserves all before/after SHA-256 hashes and comparison results. The normal release builds also use -O2, without --sanitize or --count.

**Mutants**

All mutations are applied to scratch runtime copies by `mutants.py`, leaving the checkout intact. Each compiled successfully with strict warnings before execution. No compile-warning failure counts as a catch.

| Mutant | Catcher and observed failure |
|---|---|
| `byte-80-ascii` | Exit 1; decode mismatch; decode mismatch record=130 run=0 offset=0 tail=0 lead=00 |
| `boundary-one-late` | Exit 1; decode mismatch; decode mismatch record=130 run=0 offset=0 tail=0 lead=00 |
| `drop-run-last-byte` | Exit 1; decode mismatch; decode mismatch record=0 run=1 offset=0 tail=0 lead=00 |
| `word-past-end` | Exit 1; AddressSanitizer: heap-buffer-overflow; ASan heap-buffer-overflow in the eight-byte load, empty input allocation |

`boundary-one-late` advances the prefix stop exactly one byte beyond the ASCII/non-ASCII boundary. Its first failure is prefix `[00,80]`: the leading NUL is ASCII and the following 0x80 must become replacement bytes, but the mutant copies it unchanged. `drop-run-last-byte` shortens the output and copy together, so the output assertion catches an actual omission without a sanitizer error. `word-past-end` relaxes the complete-word bound and ASan catches its load at the end of a heap input. Complete failure excerpts and log paths are retained in `mutants.json`.

**Quiet paired measurement**

Machine: Debian 13, AMD EPYC 9V74 80-Core Processor; nproc 5, cgroup CPU quota `400000 100000` (four CPU equivalents), memory limit 16 GiB. Native clang 20.1.8, WASI SDK 27, Node 24.19.0. `bash cloud/setup.sh --wasi-sdk`: Go ready 0s; clang/Node/WASI/submodules ready 1s; cache warm 80s; done 80s; nproc 5. Setup output: `/tmp/decode-ascii-setup.log`.

No other container build, test, analysis or workload ran during the measurement rounds or CPU profile. Required load/nproc snapshots are the only monitoring commands inside the run. Host activity outside this container is not controlled.

The exact accepted service and seeded generator were extracted from f4ec96c; the command wrapper/harness comes from the a4e0902 profiling unit. No service code was edited. Corpus: 100,000 requests, 417,755,377 JSONL bytes, SHA-256 `cba64bd86fdd84d7086973a145a4e8419dd31937470728efe8bf8db156a75c39`, response checksum 7,394,547 UTF-16 units. Both release variants were built using `adamic build` on this runtime base, before and after the local decoder edit. The corpus is the same as the earlier profile; the runtime/compiler base is newer. Earlier rates are not reused as the baseline.

Each Wasm variant uses one persistent reactor instance with whole-corpus warmup and exact Node response comparison. Both native variants also compare all 100,000 response lines exactly to Node outside timing. Each timed loop checks the checksum. Wasm timing includes encoding, boundary input copy, input decode, handler, response decode/release and free. Native reads and splits the JSONL before markers, warms the handler, and the parent times receipt of separate stderr serve:start/serve:stop writes. Native whole-process times include startup, reading/splitting and warmup, and are retained separately. Native marker scheduling error is small but unquantified. All variants ran serially in each interleaved round; rates are best of five.

| Round | Order (WB/WA/NB/NA = before/after Wasm/native) | Load before / after | nproc before / after | WB req/s | WA req/s | NB req/s | NA req/s |
|---|---|---:|---:|---:|---:|---:|---:|
| 1 | WB, WA, NB, NA | 0.91 / 0.97 | 5 / 5 | 7,943 | 8,483 | 10,862 | 11,534 |
| 2 | NA, NB, WA, WB | 0.97 / 0.99 | 5 / 5 | 7,906 | 8,264 | 11,478 | 11,542 |
| 3 | WA, NB, WB, NA | 0.99 / 1.00 | 5 / 5 | 8,376 | 8,057 | 11,501 | 11,732 |
| 4 | NB, WB, NA, WA | 1.00 / 1.00 | 5 / 5 | 7,888 | 8,610 | 11,665 | 11,628 |
| 5 | WB, NA, WA, NB | 1.00 / 1.05 | 5 / 5 | 7,997 | 8,440 | 11,808 | 11,585 |

| Path | Best before serve time | Best after serve time | Best before -> after req/s | Rate change |
|---|---:|---:|---:|---:|
| Wasm reactor | 11.939482 s | 11.613907 s | 8,375.57 -> 8,610.37 | +2.80% |
| Native prepared-string loop | 8.468733 s | 8.523835 s | 11,808.14 -> 11,731.81 | -0.65% |

Native whole-process minima were 19.681565 seconds before and 19.701621 seconds after, including warmup. These do not establish a file-read startup improvement. They are not startup-only measurements.

**After profile**

A separate quiet `node --cpu-prof` run kept the ordinary named, unstripped after Wasm module. An inspector capture gated around one serve batch, with monotonic timestamps, excludes generation, warmup and profiler stop bookkeeping. Profile loop 11,595.30 ms; load 1.04 -> 1.11; nproc 5 before and after. Raw automatic full-process and gated profiles are `/tmp/decode-ascii/after-all.cpuprofile` and `after.cpuprofile`; `profile-after.json` preserves the frame summaries.

The runtime `decode` frame is **227.606 ms, 1.963% self time, 211 samples**. It is distinct from Node’s TextDecoder frame named decode (0.084%). Its Wasm profile byte position is 20844 (0x516c), matching function `decode` at disassembly offset 0x516a in the unstripped artifact, and input.c:119. The word scan is inlined; no separate ascii_prefix frame appears. Self time is sampled attribution, not a precise standalone decoder duration or inclusive call time.

| After Wasm function | Self ms | Self share | Source |
|---|---:|---:|---|
| `release_last` | 1631.24 | 14.07% | runtime/heap.c:322 |
| `adamic_string_slice` | 1479.77 | 12.76% | runtime/string_slice_impl.h |
| `adamic_string_share` | 785.29 | 6.77% | runtime/string_share.c |
| `adamic_string_locate` | 704.72 | 6.08% | runtime/string_index.c |
| `adamic_function_13_Parser_string` | 617.84 | 5.33% | emitted Parser.string, accepted service.a:19 |
| `adamic_string_append` | 585.53 | 5.05% | runtime/string_append.c |
| `adamic_function_11_Parser_peek` | 583.52 | 5.03% | emitted Parser.peek, accepted service.a:15 |
| `adamic_string_equal` | 548.41 | 4.73% | runtime/string_build_impl.h |
| `adamic_allocate` | 514.12 | 4.43% | runtime/heap.c |
| `find` | 414.25 | 3.57% | runtime/map.c:105 |

**Observations and inference**

Observed: Wasm’s best rate increased 2.80%, with the after variant faster in four of five rounds. Native’s best rate decreased 0.65%; there is no observed native serve-loop speedup. The decoder’s after Wasm self share is 1.96%. The earlier profiling unit found 6.43%, but its runtime base was f4ec96c; that earlier share is context, not a controlled before-profile for this unit. No before-profile was recaptured on this newer base.

Inference: avoiding per-byte decode/encode work on the leading ASCII run contributes to the Wasm improvement, where the reactor decodes a fresh input every request. The modest native serve-loop difference is consistent with noise: this loop operates on already decoded strings and does not call the boundary decoder per request. The optimization applies to native file/argument decoding too, but it only scans the leading ASCII run and this unit provides no separate native file-read microbenchmark establishing a gain. The dominant string slicing, sharing and release/destruction costs remain. No universal service-speedup claim follows from one machine and one corpus.

**Commands, checks and reproduction**

All test output went to log files; no test process was piped to head or tail. Final successful commands:

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_TEST_WASI=1 go test ./internal/native -run "^TestDecodeASCII$" -count=1 -v -timeout 30m > /tmp/decode-ascii/oracle-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run "^(TestNativeAgreesWithNode|TestInputAgreesWithNode)$" -count=1 -timeout 30m > /tmp/decode-ascii/native-gate.log 2>&1
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run "^TestWASI" -count=1 -timeout 30m > /tmp/decode-ascii/wasi-oracle.log 2>&1
ADAMIC_TEST_WASI=1 PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH go test ./internal/native -run "^TestWASI$" -count=1 -v -timeout 30m > /tmp/decode-ascii/wasi-integration.log 2>&1
python3 internal/native/decode_ascii/mutants.py > /tmp/decode-ascii/mutants.log 2>&1
go vet ./... > /tmp/decode-ascii/vet-final.log 2>&1
gofmt -l cmd internal > /tmp/decode-ascii/gofmt-final.log
git diff --check
```

Outputs: native/input oracle `ok ... 86.603s`; WASI oracle `ok ... 165.001s`; WASI integration `ok ... 21.558s` with all 48 runtime translation units and 35/35 fixture outputs matching, plus 100,000 counted toy reactor calls, zero live values, flat memory and advancing regions. Final decoder oracle `ok ... 186.323s`. All four mutants caught, runner exit 0. Vet and gofmt exit 0 with no output; diff check clean. No full repository gate was run.

To reproduce the paired binaries after checkout, fetch `codex/wasm-requests-profile` if its a4e0902/f4ec96c source objects are absent, then run:

```bash
bash cloud/reports/decode-ascii/build.sh > /tmp/decode-ascii/reproduction-build.log 2>&1
node --disable-warning=ExperimentalWarning cloud/reports/decode-ascii/measure.mjs measure > /tmp/decode-ascii/measure.log 2>&1
node --cpu-prof --cpu-prof-dir=/tmp/decode-ascii --cpu-prof-name=after-all.cpuprofile --disable-warning=ExperimentalWarning cloud/reports/decode-ascii/measure.mjs profile > /tmp/decode-ascii/profile-after.log 2>&1
python3 cloud/reports/decode-ascii/profile.py > /tmp/decode-ascii/profile-analysis.log 2>&1
```

Run measurement/profile serially, after all builds/tests finish. The reproduction driver uses the unchanged compiler API and selects runtime snapshots without editing the checkout. The complete reproduction build regenerated all four binaries byte-identically to the original adamic-build artifacts, verified by full-file SHA-256 with the same output names. Original measured artifact hashes are retained in results.json. `objects.py before` must run before editing input.c; `objects.py after` writes the all-object comparison.

Durable evidence: results.json (every timing/load/order and hashes), objects.json (all object hashes), mutants.json (catchers/excerpts), profile-after.json (self-time frames), plus the oracle, mutation and reproduction scripts. Raw data, binaries, disassembly and logs are under `/tmp/decode-ascii`. Not covered: arbitrary exhaustive three-/four-byte products, hardware perf counters, other machines, a native file-read microbenchmark, or a complete repository gate.
