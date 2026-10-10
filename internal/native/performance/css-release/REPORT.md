CSS median throughput rises from 1,948 to 2,018 stylesheets/s (+3.56%).
Scanner median throughput rises from 1.49 to 1.53 million tokens/s (+2.43%).
Callgrind instructions fall 8.29% for CSS and 1.56% for the scanner.
Every CSS/scanner answer byte and all 261 counts-table rows are unchanged.
The unsafe shared-owner mutant compiles and is caught by ASan.

The gzip archives this report names (corpus, expected outputs, Callgrind profiles,
stdout, stderr and `logs/`) are kept off main, on branch `codex/perf-css-runtime` at
`ccfd64f`, under this same directory.

# Change and scope

Branch: `codex/perf-css-runtime`, initially cut from main `1a4e299`, then
rebased onto main `5d4c801` when the requested hot-file split landed.
Implementation after the requested rebase: `06040f80cf238bf677b614c435dab455026817e5`.
Parent: hot-file split `5d4c801`. The pre-rebase commit was `a60ec8a`.

The CSS report at `6476d7a:stage1/cohere/css/PERFORMANCE.md` identified ownership
traffic as the largest residual category: 23.60% of self instructions. This
unit implements a small runtime optimization against that cost. It does **not**
implement the proposed compiler inference for borrowed results or scalar
replacement; those proofs remain separate work. No owned return, retain,
release, allocation or counter is elided.

`heap.c` now returns after decrementing a reference that leaves the value alive,
and immediately for NULL and immortal values. Only a last reference enters
`release_last`, a non-inlined helper containing the existing destruction queue
and drain. An outermost drain always empties the queue before returning;
reentrant last releases still enqueue behind the active drain. Child releases,
free order, weak-reference clearing and nonrecursive destruction are unchanged.

Simply adding the guard removed 2.93% of CSS instructions, but clang still saved
seven registers before checking it. Moving destruction into a non-inlined
helper removes that prologue from the common path. `objdump` shows the final
normal release path has no stack frame. The ordinary flag disabling sibling
calls is retained; this optimization does not alter generated-program stack
checks or tail recursion.

Only `internal/native/runtime/heap.c` changes production behavior. Compiler
emission and lowering are unchanged. Generated CSS and scanner C are identical
between snapshots. The largest compiler proposal is still open; this is the
bounded RC runtime improvement, not a claimed implementation of borrowed-result
lifetimes.

# Fixed inputs and builds

Main lacks the CSS slice at this revision. Its immutable scratch source is
byte-for-byte `6476d7a`, including all six composed slices: CSS, selector, values,
mediaquery, cssstrings and cssnumbers. It was regenerated with the current main
compiler. `source-sha256.json` records every TS input, generated C, corpus and
runtime source hash. Both snapshots use the same generated C; baseline runtime
is main and final runtime is the committed change. The copied scanner artifact
sources also match their respective runtime snapshots.

Full CSS corpus: 24,076 cases, two option sets, 48,152 answer pairs. Go oracle
files from the preceding CSS unit contain 3,172,898 default-option bytes and
3,227,590 narrow-option bytes. Fresh Node execution of the identical TypeScript
source agrees with both. This unit does not rerun Prettier or rebuild the Go CSS
oracle; the original-library comparisons and known surrogate gaps remain in the
preceding CSS report. The stored Go outputs and corpus are archived here with
reproducible gzip headers.

CSS timing corpus: 4,952 shared successes, 510,298 UTF-16 output units/pass.
CSS Callgrind/counting sample: the unchanged 341 successes, 33,232 output units.
Scanner timing/profile corpus: 77 compiler files, 434,790 tokens. Scanner byte
gate: those files, 104 stage1 files and 18,236 generated inputs, producing
25,288,474 answer bytes. TypeScript v6.0.3 source is commit
`050880ce59e30b356b686bd3144efe24f875ebc8`.

Release/profiling/timing binaries use clang 20.1.8, `-std=c11 -O2 -g
-ffp-contract=off -fno-optimize-sibling-calls`, with runtime translation units
compiled directly. There is no sanitizer or counting in timed binaries. The
counting build separately adds `-DADAMIC_COUNT`; sanitizer builds use `-O1 -g
-fsanitize=address,undefined -fno-sanitize-recover=all`. The ordinary native
build and its stricter warning flags are also covered by the native/oracle gate.

Environment: Go 1.27.1, Node 24.19.0, Callgrind 3.24.0, Linux x86-64.
`nproc`: **5**; cgroup CPU quota: **4 cores**; memory: **17.6 GB**.
`bash cloud/setup.sh` timing lines:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (114s)
setup: done in 114s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

# Measurement

Seven rounds interleave baseline, final and Node, reversing order on alternating
rounds. No compilation, gates or profiling run concurrently. Startup, file I/O,
case decoding and final checksum printing are included. Every subprocess writes
stdout/stderr to files; checksums and empty stderr are required. Full answer
bytes are checked separately. Use medians; these modest wall-time gains are not
claims of an 8.29% speedup or general formatter/scanner rankings.

| Workload | Baseline / s | Final / s | Node / s | Final/baseline |
|---|---:|---:|---:|---:|
| CSS stylesheets | 1,948 | 2,018 | 2,464 | 1.0356 |
| Scanner tokens | 1,493,287 | 1,529,541 | 1,375,302 | 1.0243 |

Raw seconds for every round, commands and output checksums are in
`measurements.json`. Neither Go nor Prettier throughput is remeasured in this
unit. The compiler and runtime baseline differ from the previous CSS unit;
comparisons with its 1,917/s number would mix changes from other workers.

Callgrind records `Ir` alone. The scanner's checked parser collapses same-name
inline DWARF records and requires the sum of all self costs to equal `summary`.
Inclusive costs overlap and are not added. Raw profiles and profiler stdout/
stderr are archived, including the guard-only intermediate candidate.

| Workload | Baseline Ir | Final Ir | Reduction |
|---|---:|---:|---:|
| CSS sample | 1,384,254,881 | 1,269,486,791 | 8.29% |
| Scanner compiler files | 2,516,568,688 | 2,477,408,930 | 1.56% |

CSS release self cost changes from 278,252,625 to 163,484,126 instructions,
summing `adamic_release` and the new `release_last` on the final side. Scanner
release self cost changes from 85,729,519 to 46,569,490. The instruction savings
are attributable to the runtime, while wall-time attribution remains subject
to timing variation and memory costs.

# Counts, correctness and mutation

`TestCountsAreRecorded` passes all **261** recorded fixtures without changing
`internal/oracle/counts.md`: every row and column is identical, so no row is
worse. Counts on the two profiled workloads are also exactly identical:

| Workload, both runtimes | Allocations | Frees | Retains | Releases | Peak live | Regions |
|---|---:|---:|---:|---:|---:|---:|
| CSS sample | 1,714,918 | 1,714,918 | 6,408,810 | 6,626,382 | 140,839 | 0 |
| Scanner compiler files | 432,199 | 432,199 | 2,053,192 | 2,244,312 | 2,262 | 0 |

The final CSS release and ASan/UBSan builds produce every expected Go byte in
both option sets. `ASAN_OPTIONS=detect_leaks=1` and
`UBSAN_OPTIONS=halt_on_error=1` require clean exits and empty stderr. Scanner's
full gate independently runs its sanitized native binary, Go scanner and Node,
including leak checks. Both actual measured scanner snapshots also pass the
complete answer-byte protocol against Go and Node.

New proving program: `internal/native/testdata/release_fastpath.ts`. It builds a
dynamic 200-character string, shares it between an array and a local, removes
the array's owner, then reads the surviving local. Node and native print
`200 xxxx\n`, with a clean leak check.

New unsafe mutant: change the release guard from `--references != 0` to
`--references > 1`. Dropping one of two references then incorrectly destroys
still-shared storage. The test builds the mutant runtime successfully and
requires **AddressSanitizer: heap-use-after-free** from this real program;
a compilation failure is not accepted as a catch. The final regression passes
and logs the ASan catch. Existing native/oracle mutants also run in the full
core gate. The scanner's three existing mutants (punctuator, decimal separator,
regex rescan) are caught again by its Go oracle.

# Commands and gate limits

All test output is saved to log files. `logs/*.gz` retain complete output.
After the final helper split:

```
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=30m ./internal/native ./internal/oracle > /tmp/css-runtime-final-core-gate.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=30m ./stage1/typescript/scanner > /tmp/css-runtime-final-scanner-gate.log 2>&1
python3 internal/native/performance/css-release/verify.py /workspace/scratch/css-runtime /workspace/scratch/css-perf/final --final-only > /tmp/css-runtime-final-verify.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_SCANNER_PROFILE_SNAPSHOTS=/workspace/scratch/css-runtime/baseline/scanner:/workspace/scratch/css-runtime/final/scanner go test -v -count=1 -timeout=30m ./stage1/typescript/scanner -run '^TestProfileSnapshotsAgree$' > /tmp/css-runtime-final-scanner-snapshots.log 2>&1
python3 internal/native/performance/css-release/measure.py /workspace/scratch/css-runtime /workspace/scratch/css-perf/final/source/css/print_main.ts /workspace/scratch/css-perf/final/shared.txt > /tmp/css-runtime-measure.log 2>&1
```

Results: native **PASS, 456.511s**; oracle **PASS, 409.878s** (1,031 native
misses, 573 Node misses, 24 probe misses, no cache hits); scanner **PASS,
38.462s**; scanner snapshot bytes **PASS, 12.181s**; CSS byte, sanitizer, leak
and count checks **PASS**. The separate final unsafe-mutant regression also
passes (**70.056s**) while the heavily loaded preliminary gate runs.

An intermediate guard-only candidate was tested with
`ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout=30m ./...`.
Native, oracle, compiler packages and several slices passed, but
`internal/unicodeproperties/TestCanonicalizeUnicodeNode` failed with
`node: signal: killed`. Its batches have a four-minute context deadline;
the cgroup reported zero OOM kills. The remaining oversized preliminary run
was terminated after about twelve minutes to finish the final core and scanner
gates without competing builds. **A complete final all-package gate is not
claimed.** The isolated retry **passes in 486.086s**: 2,964 class scans and 2,110
singleton groups, **5,653,004,288 code points examined, zero disagreements**.
Its complete log is retained.

# Reproducing the snapshots

Recover the CSS source without adding it to this compiler branch:

```
git archive 6476d7a stage1/cohere/css stage1/cohere/selector stage1/cohere/values stage1/cohere/mediaquery stage1/cohere/cssstrings stage1/cohere/cssnumbers > /tmp/css-source.tar
```

Extract outside the repository and copy those six directories under a common
`source/`. The manifest verifies the source, and the archived cases, expected
outputs, shared successes and profile sample can be decompressed beside it.
Generate C once with `go run ./cmd/adamic c <source/css/print_main.ts>` and reuse
it for both runtimes. Obtain baseline runtime via `git archive 1a4e299
internal/native/runtime`, and final runtime from this branch. Compile the same
C against each runtime directory using the flags above.

Scanner artifacts are generated with:

```
ADAMIC_TYPESCRIPT_SOURCE=<TypeScript-v6.0.3> ADAMIC_SCANNER_PROFILE_DIR=<snapshot/scanner> go test -v -count=1 ./stage1/typescript/scanner -run '^TestProfileArtifacts$' > <artifact-log> 2>&1
```

Reuse its generated C, TS files, Go oracle and compiler manifest in both sides;
rebuild each native binary against its frozen runtime. The `measure.py` and
`verify.py` scripts describe the artifact names. Profile each binary with
`valgrind --tool=callgrind --callgrind-out-file=<file>` using `sample.txt count`
for CSS and `--manifest <compiler.txt> --count` for scanner, then summarize with
`stage1/typescript/scanner/profile.py`'s checked `summarize` function.

The hot-file split landed as `5d4c801` and the branch was rebased cleanly.
Both workloads regenerate byte-identical C. Both rebuilt baseline and final
binaries have identical `.text`, `.rodata` and `.data` sections to the measured
binaries, verified separately for each workload and side. Therefore the
measurements remain applicable after the split; no new timing is substituted.
`rebased-code-sha256.json` records the checked section hashes, and
`rebased-source-sha256.json` records the new runtime/header layout. The full CSS
byte and ASan/UBSan/LSan checks are repeated on a freshly rebuilt rebased runtime.
Post-rebase formatting, vet, filtered core and full scanner results are recorded
below. A full all-package post-rebase gate is not claimed.

Post-rebase commands and results:

```
gofmt -l cmd internal > /tmp/css-runtime-rebased-gofmt.log 2>&1
go vet ./... > /tmp/css-runtime-rebased-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=30m ./internal/native ./internal/oracle -run '^(TestReleaseSharedValueAndUnsafeMutant|TestCountsAreRecorded|TestNativeAgreesWithNode)$' > /tmp/css-runtime-rebased-core-gate.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=30m ./stage1/typescript/scanner > /tmp/css-runtime-rebased-scanner-gate.log 2>&1
```

Formatting and vet: **PASS, empty logs**. Native positive/unsafe-mutant test:
**PASS, 12.182s**. Oracle: **PASS, 91.268s**, 991 native and 514 Node uncached
checks, no cache hits; every recorded count remains unchanged. Scanner:
**PASS, 34.999s**, including its sanitizer, leak and three mutant checks.
Rebuilt rebased CSS: **PASS**, every Go byte in both option sets, empty stderr
under ASan/UBSan/LSan. The proving program subsequently receives repository
loop/spacing style only; its Node, sanitizer and unsafe-mutant regression is
rerun and **PASS, 0.620s**.
