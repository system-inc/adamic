Built direct ASCII paths, counted substring views and allocation-free ASCII characters.
Commits: baseline 66534dc, ASCII paths 19be8f5, views d3ccac21a29973425cdf2224b8f97065358dd51e.
Commands and outputs: uncached native/oracle PASS; native 10,114.28 -> 10,083.68 -> 14,655.94 requests/s; parse 6.5036G -> 6.4800G -> 6.4937G Ir.
Mutants: 12 caught by Node answers, offset/lifetime/allocation checks, ASan/LSan, profile accounting and benchmark assertions; every case is tabulated below.
Not covered: Apple execution, deployed Cloudflare, full repository gate, hardware counters and simultaneous escaping-view high-water memory.

## Baseline

Read CLAUDE.md and both requested reports first. The request evidence at a4e0902 profiles f4ec96c, not this base. The 10.2% parser bucket was corrected later: input.c decode belongs to input, not indexing. Historical rates and instructions are not this machine or this base.

This container reports Intel Xeon Platinum 8370C at 2.80GHz, nproc 5, cpu.max 400000 100000. Setup: Go/clang/Node ready 0s, submodules 1s, cache warm/done 204s. Source /workspace/adamic-tools/env.sh in every shell. Go 1.27.1, clang 20.1.8, Node 24.19.0, Valgrind 3.24.0 extracted from official Debian package into scratch.

The exact six-route corpus is 100,000 requests, 417,755,377 bytes, SHA-256 cba64bd86fdd84d7086973a145a4e8419dd31937470728efe8bf8db156a75c39. Checksum 7,394,547 UTF-16 units. Reused a4e0902 command.a, generator and timing markers. The native-only adaptation removes the Wasm engine from measure.mjs, preserving warmup, checksum, serial Node/native ordering and five rounds. All rounds are in requests-before.json. Scratch native C is unchanged between runtime builds.

Native release flags: -O2, -ffp-contract=off, -fno-optimize-sibling-calls, no LTO, no sanitizers, no ADAMIC_COUNT. Parse and step 2/3 native snapshots also include -g for diagnostic symbols; the baseline native binary does not. Timings exclude startup, input splitting, warmup and teardown, using separate stderr serve markers as in the evidence harness. Baseline rounds 1-3 overlapped benchmark preparation builds; only rounds 4-5 were free of our other CPU work. Their native best is 8,652.71 requests/s. Overall best 9,257.69 and Node 36,049.47 are observations, not a speedup claim. Host load and rate variation are large.

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

## Step 2: direct ASCII paths

Flagged ASCII slice and at bypass locate, and indexOf compares bytes without walking either string as UTF-16 or asking for byte-to-unit translation. An ASCII needle skips the surrogate-edge walk in other affix/search operations too. charCodeAt already had the direct inline path and was not changed. Slices propagate last - first UTF-16 units even for copied surrogate boundaries, avoiding a later recount. No input.c change.

Five serial request rounds: native best **10,083.68 requests/s**, Node best **36,254.50**. All checksum checks passed. Whole-process parse: **6,479,991,960 Ir**, down **23,638,266 (0.3635%)**. Same generated C and corpus, same release optimization flags. The rate comparison is unpaired and noisy; before clean-round best is 8,652.71 and overall best is 9,257.69. Baseline native verify compared all 100,000 response lines byte for byte with Node, plus final checksum. Allocation sites are unchanged in this step; character indexing still mints short strings under the old sharing policy.

Validation commands, with stdout/stderr to the named logs:

```sh
ADAMIC_STRING_OPERATIONS=100000 ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/oracle -run 'TestStringsMatchJavaScript|TestStringIndexMatchesNode|TestStringBuildingMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringViewAfterAppendMatchesNode|TestNativeAgreesWithNode/.*/(10_unicode|shared_slices|regexp_split|regexp_unicode|lone_surrogates|library_string_indices)' -count=1 -v -timeout 30m > /tmp/string-views-step2-tests.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(shared_slices|regexp_split|regexp_unicode|lone_surrogates|library_string_indices)' -count=1 -v -timeout 30m > /tmp/string-views-step2-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/load/testdata/0.1/compile/10_unicode' -count=1 -v -timeout 30m > /tmp/string-views-step2-unicode.log 2>&1
```

The first combined filter selected native sweeps but no oracle fixtures. The two corrected oracle invocations above passed: six string/regex fixtures plus program 10, with source Node, backend Node, release native, ASan/UBSan and LeakSanitizer. Native passed **11,460 answers, zero mismatches** (the current sweep extends the historical 11,080), **100,000 stateful operations**, cache states, indexed BMP/supplementary/lone-surrogate sweeps and reads after append. Native 24.613s, focused oracle 10.846s, program 10 0.563s.

Mutant: only the new ASCII slice path returns one fewer byte. It compiles, then TestStringsMatchJavaScript fails with **890 of 11,460 Node answers differing**, exit 1. It was restored before further work; no warning/refusal kill is counted. The compressed mutant and validation logs are beside this report.

## Step 3: counted substring views

Slice and substring share whole code-point byte spans and retain the ultimate counted owner once. Nested views keep byte offsets into the original owner; their units and indexes describe only their own range. Whole slices retain their string. Empty slices reuse the immortal empty string. Surrogate-split boundaries must still build the required WTF-8 halves. Unmarked zero-count stack pieces copy even for whole slices: there is no parent count to keep. The input.c decoder is unchanged.

The new policy removes the 64-byte minimum. A counted parent's header plus max(capacity, length) may be at most eight times the view's header plus byte length; otherwise copy. This includes spare append capacity and uses overflow-safe quotient/remainder arithmetic. Immortal literal bytes pin no counted storage. Character indexing additionally reuses 128 immutable ASCII unit strings, including ASCII units within Unicode parents. That is the main allocation reduction; nonempty substring views still allocate a header. The pool occupies 128 string headers plus 128 bytes and follows the runtime's existing single-threaded cache model.

Final five serial native/Node request rounds: native best **14,655.94 requests/s**, Node **37,870.49**. All timed checksums passed. Native step 2 and step 3 separately compared **all 100,000 responses byte for byte** with Node, including the final checksum. CPU preparation and tests were excluded from these final rate rounds. Cross-stage rates remain unpaired observations on a noisy shared machine.

| Runtime | Native best requests/s | Whole parse Ir | Change from preceding step |
|---|---:|---:|---:|
| Before | 10,114.28 clean five-round repeat | 6,503,630,226 | baseline |
| Direct ASCII paths | 10,083.68 | 6,479,991,960 | -0.3635% Ir |
| Views and immutable ASCII characters | 14,655.94 | 6,493,671,247 | +0.2111% Ir |

Final parse is **0.1531% below baseline**, but the policy adds instructions relative to step 2. This parse driver does no character indexing, so the ASCII character pool gives it no benefit. The identical generated C, corpus and -O2 flags isolate runtime changes; this is not evidence of a parser speedup of the size seen in requests.

### Retention policy measurement

policy.c measures 41 non-whole slices across parent capacities 128, 256, 512, 1,024, 8,192 and 65,536 and sizes 1, 4, 8, 16, 32, 64 and 128. Each slice is released before the next: isolated_pinned_bytes sums the owner storage retained by those individually escaping views, rather than claiming a simultaneous high-water measurement.

| Bound | Views | Copies | Copied bytes | Sum of isolated pinned owner bytes |
|---|---:|---:|---:|---:|
| 4x | 11 | 30 | 897 | 3,008 |
| 8x | 19 | 22 | 636 | 7,360 |
| 16x | 26 | 15 | 507 | 13,952 |

Choose 8x: it permits short slices of small parents while limiting the retained storage; 16x nearly doubles pinned storage in this probe for seven fewer copies. This is a synthetic policy comparison, not a claim that 8x optimizes every workload. The final request and parse measurements use 8x. TestRuntimeStringViews checks the exact portable header-dependent boundary and a 64-byte logical string with 8,192-byte spare capacity.

### Final allocation attribution

Same disjoint scratch counters and serve-minus-control method. Total **51,586,613**, or **515.86613/request**, down **67.25%** from 1,575.00123. Slice allocations fall from 1,284.60190 to **225.47304/request**. Concatenation, grown append, number formatting and nonstring categories are unchanged; other strings fall by 624 across the corpus. The character pool and immortal empties remove allocations; changing copied substrings into views chiefly removes copied bytes. This is not a count of backing-buffer mallocs or a high-water memory comparison.

### Per-function follow-up

Separate nominal 1ms SIGPROF PC captures, not the rate runs. The step 2 capture has 2,429 samples over 9.746219 CPU seconds, step 3 has 1,582 over 6.317285 CPU seconds. Unresolved PCs stay in each denominator; inlining moves work between names, notably charCodeAt into at. Zero samples is not proof of zero work.

| Function | Step 2 sampled self ms | Step 3 sampled self ms | Step 3 parse self Ir |
| adamic_string_slice | 1163.61 | 263.55 | 57,077,781 |
| adamic_string_locate | 822.55 | 291.51 | 55,722,688 |
| adamic_string_at | 525.63 | 998.31 | 0 |
| adamic_string_char_code | 136.42 | 131.78 | 8,177 |
| adamic_string_index_of_at | 0.00 | 0.00 | 0 |
| adamic_string_share | 353.09 | 107.82 | 36,118,810 |
| adamic_allocate | 629.95 | 211.64 | 135,865,980 |
| adamic_string_units | 80.25 | 83.86 | 86,388,629 |

Flagged ASCII at/charCodeAt are O(1) byte reads; flagged slice computes byte bounds directly and returns a header view or an intentional pin-avoidance copy. Flagged indexOf no longer translates positions, but keeps the existing naive byte search, worst-case O(haystack times needle). Unicode indexing retains its existing unit cache/locate behavior. Views still allocate headers and retain owners; UTF-16 caches still take work. Those are the remaining costs.

## Mutant evidence

Every mutant was restored. All compiled; none was killed by a warning or static refusal. Logs and exact commands are preserved beside this report.

| Mutant | Catcher |
|---|---|
| ASCII slice one byte short, step 2 | Node sweep: 890 of 11,460 answers differ |
| View offset +1 | TestRuntimeStringViews: view offsets or holds differ |
| Omit owner retain | shared_slices oracle: AddressSanitizer heap-use-after-free |
| Retain owner twice | shared_slices oracle: LeakSanitizer, 5,468 bytes in 16 allocations |
| Ignore spare capacity | TestRuntimeStringViews: spare capacity was pinned |
| Use 16x instead of 8x | TestRuntimeStringViews: eightfold storage boundary differs |
| Treat unmarked borrowed stack storage as shareable | TestRuntimeStringViews: stack piece was retained |
| ASCII cached byte XOR 1 | TestRuntimeStringViews ordered hash differs from Node |
| Allocate and free a temporary before returning the correct cached character | TestRuntimeStringViews: ASCII indexing allocated |
| Callgrind summary +1 | Reconciliation rejects self costs not summing to summary |
| Add a response byte only after warmup | Timed batch checksum: 7,494,547 differs from 7,394,547 |
| Change a response byte without changing length | Exact Node response assertion fails at request 0 |

The new counted lifetime test releases the original parent and intermediate view while the last view remains, reads the surviving bytes against Node, and verifies that the last release leaves zero live heap values. It runs release and ASan/UBSan builds. Existing shared_slices supplies independent sanitizer checks on strings built at runtime; immortal literals cannot hide a missing hold.

## Reproduction and scope

service.a, host.mjs and command.a preserve a4e0902's accepted service, seeded generator and serve markers; only command.a's import location changed. measure.mjs is the native/Node adaptation described above. Fixed C snapshots were generated on the runtime base and reused for all three measurements. To prepare and measure the checked-out runtime:

```sh
bash cloud/setup.sh --wasi-sdk > /tmp/string-views-setup.log 2>&1
source /workspace/adamic-tools/env.sh
mkdir -p /tmp/wasm-requests-profile
go build -o /tmp/wasm-requests-profile/adamic ./cmd/adamic > /tmp/string-views-build.log 2>&1
/tmp/wasm-requests-profile/adamic c cloud/reports/string-views/command.a > /tmp/wasm-requests-profile/command.c 2> /tmp/string-views-emit.log
clang -std=c11 -O2 -g -ffp-contract=off -fno-optimize-sibling-calls -I internal/native/runtime /tmp/wasm-requests-profile/command.c internal/native/runtime/*.c -lm -o /tmp/wasm-requests-profile/service-native > /tmp/string-views-native-build.log 2>&1
node --disable-warning=ExperimentalWarning cloud/reports/string-views/measure.mjs prepare > /tmp/string-views-prepare.log 2>&1
node --disable-warning=ExperimentalWarning cloud/reports/string-views/measure.mjs measure /tmp/wasm-requests-profile/results.json > /tmp/string-views-measure.log 2>&1
python3 cloud/reports/string-views/instrument.py /tmp/wasm-requests-profile/counted-runtime > /tmp/string-views-instrument.log 2>&1
clang -std=c11 -O2 -ffp-contract=off -fno-optimize-sibling-calls -I /tmp/wasm-requests-profile/counted-runtime /tmp/wasm-requests-profile/command.c /tmp/wasm-requests-profile/counted-runtime/*.c -lm -o /tmp/wasm-requests-profile/counted > /tmp/string-views-count-build.log 2>&1
/tmp/wasm-requests-profile/counted /tmp/wasm-requests-profile/requests.jsonl run > /tmp/string-views-count-run.log 2>&1
/tmp/wasm-requests-profile/counted /tmp/wasm-requests-profile/requests.jsonl control > /tmp/string-views-count-control.log 2>&1
python3 cloud/reports/string-views/count-summary.py /tmp/string-views-count-run.log /tmp/string-views-count-control.log > /tmp/string-views-allocations.json
```

For before/after comparisons emit command.c and parse.c once on 8cb9d252, then rebuild those same files with each stage's runtime. Fetch the evidence refs explicitly if the clone tracks only main. Obtain parse-speed prepare.py with git show origin/codex/parse-speed:internal/native/performance/parse-speed/prepare.py and adjust only its repository root as stated above. That script retrieves batch 8 through git and uses the pinned TypeScript checkout. The upstream parser driver and corpus hashes are retained in scratch; raw Callgrind profiles and named results are committed here.

Allocation instrument.py modifies a scratch copy only. It supports both the old and new sharing implementations. The runtime allocator and input.c are unchanged. Policy comparison: compile policy.c and runtime/*.c with the named release flags and -DSHARE_FRACTION=4, 8 or 16. Mutants.py reproduces the reversible runtime mutants sequentially and restores each file in finally; run from /workspace/adamic with the toolchain environment sourced. Profile accounting uses stage1/typescript/scanner/profile.py's summarize function and rejects an altered summary.

WASI setup additionally reported Go 0s, clang/Node 1s, WASI SDK 5s, submodules 5s, cache warm/done 146s. The final counted reactor passed host.mjs over all 100,000 requests with exact Node answers, zero live values, zero regions, and linear memory fixed at **1,638,400 bytes**. Its timing ran alongside correctness tests and is not a release rate comparison.

Apple execution and a deployed Cloudflare Worker are unavailable in this Linux container. The shared C runtime and wasm32-wasi reactor were exercised; no platform-specific backend or decoder change was made. No hardware-counter measurement, exhaustive workload distribution, simultaneous escape high-water study or full repository test gate is claimed.

## Final gate

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/string-views-update-counts.log 2>&1
ADAMIC_STRING_OPERATIONS=100000 ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/native ./internal/oracle > /tmp/string-views-final-gate.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/string-views/typescript go test ./stage1/typescript/parser -run '^TestWholeCompilerAgrees$' -count=1 -v -timeout 30m > /tmp/string-views-parser-parity.log 2>&1
go vet ./... > /tmp/string-views-vet.log 2>&1
gofmt -l cmd internal > /tmp/string-views-gofmt.log
git diff --check > /tmp/string-views-diff-check.log 2>&1
```

Counts refresh PASS 73.080s, changing 51 fixture rows to the new ownership/allocation observations. The **uncached complete native and oracle packages PASS**, native 312.511s and oracle 330.538s. This includes native/JavaScript-backend agreement, program 10, surrogate-split regex fixtures and sanitizer/leak checks. Focused native validation additionally reports 11,460 UTF-16 answers with zero mismatches and 100,000 stateful operations. The current UTF-16 sweep extends the historical 11,080 answers; no answers were removed. Parser whole-tree comparison PASS 83.502s: **77 files and 44,766,682 identical bytes** against the independent Go parser on native and JavaScript. Vet, formatting, JavaScript syntax and diff checks passed. Protected compiler files and input.c are unchanged.

After all gates completed, the saved unmodified baseline binary was repeated for five serial rounds with no concurrent preparation/tests. Native best **10,114.28 requests/s**, Node **32,329.72**, all checksum assertions passed (requests-before-clean.json). Use this clean repeat in the stage table. The original before data remains preserved. Step 2's 10,083.68 is within this timing variation: no request-throughput improvement is established for that step, although its instruction reduction is deterministic. Final native best is 44.9% above the clean repeat, while Node also varies across captures; this remains an unpaired observation, not a causal effect estimate.

The exact generated command.c and parse.c used for runtime builds are compressed beside this report, with SHA-256 values in snapshots.json. compiler-corpus.sha256 pins all 77 source contents. Decompress the fixed C and compile against the runtime checkout for instruction comparisons without regenerating the parser. Scratch binaries and full raw profiles remain under /workspace/scratch/string-views.

Pushed each implementation stage only to codex/string-views. No main/area push or pull request.
