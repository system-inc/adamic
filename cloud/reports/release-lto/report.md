Built: unchanged generated C and 48 separately archived runtime units, native -O2 versus ThinLTO; baseline production flags unchanged.  
Commits: runtime base 1740da37b800ad2793218bec38ebe723cb80ec33; evidence f362559fbec5ab56a5081195002d70ca69df1a22; harness pins below.  
Commands and outputs: pinned instrument check and five interleaved LTO rounds recorded with wall/user time; full Node/Go byte parity and native/oracle gate PASS.  
Mutants: compiled SourceFile-to-XourceFile and health-status-200-to-201 outputs caught by byte comparison; summary-plus-one and footer-event-plus-one accounting mutants rejected; pinned output-line mutant rejected.  
Not covered: full repository gate, other LLVM versions/platforms, WASI execution, a generated-C translation-unit split, a hardware-cycle ranking, host-wide isolation, or a production default change.

# Native release ThinLTO, October 7, 2026

## Inputs and tools

The branch starts at `origin/area/runtime`,
`1740da37b800ad2793218bec38ebe723cb80ec33`. No compiler, parser, scanner, runtime,
or sanitizer source was edited. Both variants use the same generated C bytes.
The parse-speed branch supplies the harness, not its subsequent compiler changes.

- Parse harness: `origin/codex/parse-speed`, `2cdf8dd3faff065390720cfc8bc9b09a1c9e62d2`,
  `internal/native/performance/parse-speed/prepare.py`. It reconstructs batch 8
  from `4189abd3490757e8abe13722ceb365c451293e92` and removes only the rule traversal
  in the scratch parse driver. Context construction, line table, parent map,
  parser/scanner, sorting and cleanup remain. Measured invocation:
  `parse --manifest compiler.txt --count`.
- Corpus: TypeScript 6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`, exactly 77
  `src/compiler/*.ts` files. All 77 SHA-256 entries match the upstream harness's
  corpus manifest. cohere is `715ba94f3608a6500086b1076ce5cb7e51b836db`.
- Service: the **native command arm** of `origin/codex/wasm-requests-profile`,
  `a4e0902afc35cdc79c09fa7e58fa8d56b2189f55`. Its command driver, service source,
  generator and Node expectation calculation are retained unchanged apart from
  scratch import/file paths. Six routes, 100,000 seeded requests, 417,755,377
  input bytes; input SHA-256
  `cba64bd86fdd84d7086973a145a4e8419dd31937470728efe8bf8db156a75c39`.
  Measured command: `service requests.jsonl run`. Warm timing uses the original
  `warm` argument, which runs a batch before its `serve:start` marker.

Linux 6.18.44, AMD EPYC 9V74, `nproc=5`, cgroup `cpu.max=400000 100000`, 17.6 GB
reported memory. Go 1.27.1, Node 24.19.0, Valgrind/Callgrind 3.24.0.
clang, lld and llvm-ar are **20.1.8**, LLVM revision
`87f0227cb60147a26a1eeb4fb06e3b505e9c7261`.
The baseline uses **GNU ld 2.44**; ThinLTO uses **lld 20.1.8**.
[versions.log](evidence/versions.log) preserves the complete version output.

`bash cloud/setup.sh` passed: Go ready 0s; clang ready 0s; Node ready 1s;
submodules ready 1s; build cache warm 111s; done 111s on 5 processors.
Every build/test shell sourced `/workspace/adamic-tools/env.sh`.
Valgrind was absent. `sudo` was unavailable and apt had no package candidate;
extracting Debian's `valgrind_3.24.0-3_amd64.deb` into scratch worked.

## Instrument check before the LTO comparison

Core 3 is permitted (affinity 0–4). Both programs used taskset -c 3 and
GOMAXPROCS=1 for Go. The same 77-file manifest and upstream Go cohere harness
were used: it forces the line table and keeps the parsed source file alive.
The native is today's unchanged runtime-base release executable.

**perf stat -r 10 -e cycles,instructions,cache-misses,branch-misses reports
“not supported” for all four events**, on a probe and the actual native/Go
commands. perf_event_paranoid is 2, but the result is unsupported events,
not permission denied. perf 6.12.107 and dependencies were extracted from
Debian trixie because perf was absent. The requested fallback is **hyperfine
user time**, ten interleaved pairs with alternating native/Go order, pinned
to core 3. No concurrent test, build, profiler or other benchmark ran.
One-minute load was 0.40–0.49. Background container services remained;
host-wide exclusive isolation cannot be certified from this container.

| Program | Whole-process Callgrind Ir | Best of 10 wall s | Best of 10 user s |
| --- | ---: | ---: | ---: |
| Native | 6,503,615,880 | 0.847628 | 0.797904 |
| Go | 1,598,015,620 | 0.170946 | 0.144318 |

Native/Go is **5.529x in user time**, versus **4.070x in this fresh instruction
profile**: 35.8% apart, outside 10%. The historical 6.50G/1.685G (3.9x) ratio
is also outside the band. The fresh Go count is 1.598G, not historical 1.685G;
no cause for that difference is established here. Go profiling uses
GODEBUG=asyncpreemptoff=1 to run under Valgrind; hyperfine uses normal preemption.
**The instruction ranking is not confirmed by this instrument check.**
User time measures scheduled execution, not cycles or IPC; no cycle ratio
can honestly be supplied on this box.

### Simulated cache/branch reranking

The requested simulation ran with --cache-sim=yes --branch-sim=yes on an
-O2 -g baseline build. All runtime and generated-program optimization flags
remain identical. Release and debug executable .text bytes are identical,
SHA-256 dcbd5108c285686ad30ffae8db0a27b9c55542cbbce54ae676d93525cc16121a.
DWARF separates inline field checks and inline slab code from generated callers.
The simulator model is I1/D1 32KiB, 64-byte lines, 8 ways; LL 256MiB,
direct mapped. This is a simulator model, not a measurement of the EPYC
cache hierarchy or allocator locality.

Callgrind has no hardware cycle event here. To give the permitted simulation
a concrete ranking, the central **hypothetical cycle-cost model** is
Ir + 4*(I1mr+D1mr+D1mw) + 50*(ILmr+DLmr+DLmw) + 15*(Bcm+Bim).
L1/LL penalties are additive. Overlap, memory-level parallelism, frequency,
out-of-order execution and real predictor behavior are not modeled.
Sensitivity uses (L1,LL,branch) = (2,40,8) and (10,200,25), versus central
(4,50,15). These assumptions must not be read as measured cycles.

| Disjoint bucket | Ir rank | Central model rank | Ir M | L1 misses M | LL misses M | Branch mispredicts M |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Remainder: parser, arrays, driver, libc and unattributed inline | 2 | 1 | 1514.968 | 55.806 | 1.128 | 7.606 |
| Scanner generated control | 1 | 2 | 1585.314 | 18.405 | 0.001 | 6.026 |
| Character reads and UTF-16 indexing | 3 | 3 | 960.752 | 1.740 | 0.105 | 0.818 |
| Releases and child destruction | 4 | 4 | 566.150 | 5.555 | 0.000 | 16.597 |
| Retains | 6 | 5 | 331.008 | 1.958 | 0.000 | 6.670 |
| String equality | 5 | 6 | 339.705 | 1.879 | 0.000 | 4.498 |
| File reading and UTF-8 input decode | 7 | 7 | 225.804 | 0.437 | 0.102 | 0.002 |
| Node construction generated control | 8 | 8 | 212.506 | 3.456 | 0.000 | 0.057 |
| Object field and call plumbing | 9 | 9 | 194.088 | 4.237 | 0.000 | 0.474 |
| Allocation entry and libc allocator | 10 | 10 | 146.184 | 2.402 | 0.162 | 1.101 |
| Slab allocator | 11 | 11 | 123.057 | 0.504 | 0.001 | 0.976 |
| Other string operations | 12 | 12 | 95.050 | 1.724 | 0.094 | 0.450 |
| Substrings | 14 | 13 | 81.235 | 1.900 | 0.000 | 0.028 |
| Line table generated control | 13 | 14 | 83.446 | 0.000 | 0.000 | 0.201 |
| Parent map generated control | 15 | 15 | 44.350 | 1.379 | 0.000 | 0.921 |


Moved in all three models: remainder 2→1, scanner 1→2, retains 6→5, string
equality 5→6, substrings 14→13, line table 13→14. In the high-penalty model,
releases additionally move 4→3 and character reads 3→4. **Slab allocator stays
11th**, separate from allocation entry/libc (10th), **retains** (5th modeled)
and **releases/child destruction** (4th central, 3rd high).
Slab includes inline take/give/chunk/list management and deallocation;
allocation entry bookkeeping and libc remain separate. This reranks the current
baseline using the upstream mechanism split extended for today's slab functions.
It does not assert that older release bucket definitions stayed identical.
The remainder includes generated parser/arrays and unattributed inline;
simulation cannot turn it into exact per-mechanism hardware cycle attribution.

All 13 simulated self-event sums equal the profile's totals footer.
Its summary exceeds that footer by exactly **2 Ir**, with zero difference
in every other event. That observed accounting discrepancy is retained,
not rounded away or assigned to a bucket. Ranks use self costs.
The footer's final event +1 mutant is rejected by the same reconciliation
function. Raw profiles, vectors, ranks, assumptions and logs are committed.
**An actual cycle ranking remains unmeasured** and needs a host with working
hardware PMU counters.

## Decision

**Propose ThinLTO as the native release default on supported clang/linker pairs.**
Both workloads exceed the requested 10% threshold in pinned, interleaved wall time.
Parse saves 13.19% wall / 16.81% user time; service saves 15.06% wall / 15.26% user time.
The parse instruction reduction is 9.14%; the service instruction reduction is 6.02%.
Every good output comparison passed. This is a proposal, with
measurements and reproduction helpers, not an implementation of the default.
Sanitized lanes stay as they are.

The material cost is the parser's link: a cached runtime archive still leaves
16.08 seconds of generated-C compilation and LTO linking, against 2.52 seconds
without LTO. Its binary also grows 37.8%. The service has a smaller binary and
cheaper cold build, but its cached build grows from 0.208 to 0.458 seconds.

## Measurements

These are the final runs after the instrument check: `taskset -c 3`, hyperfine
1.19.0, five pairs with alternating order and no concurrent build/test/profiler.
Each hyperfine invocation uses `--runs 1 --warmup 0 --shell none --show-output`,
so separate invocations alternate programs rather than timing one program five
times and then the other. The binaries/corpus had already been warmed.
Best wall and best user values are selected independently, as requested.
Percent saved is `100 * (1 - ThinLTO / baseline)`.

| Workload | Version | Callgrind Ir | Best wall s | Best user s |
| --- | --- | ---: | ---: | ---: |
| Parse | Baseline | 6,503,615,880 | 0.917109 | 0.898100 |
| Parse | ThinLTO | 5,908,993,714 | 0.796108 | 0.747111 |
| Service | Baseline | 135,285,492,290 | 11.096163 | 10.699734 |
| Service | ThinLTO | 127,147,493,270 | 9.425452 | 9.067095 |

| Round | Parse base wall/user s | Parse ThinLTO wall/user s | Service base wall/user s | Service ThinLTO wall/user s |
| --- | ---: | ---: | ---: | ---: |
| 1 | 0.934972 / 0.899309 | 0.815627 / 0.795878 | 11.567951 / 11.101677 | 9.769813 / 9.346118 |
| 2 | 0.956361 / 0.931587 | 0.839235 / 0.790218 | 11.461074 / 11.113848 | 9.727791 / 9.349506 |
| 3 | 0.940213 / 0.915040 | 0.823577 / 0.791032 | 11.146315 / 10.776314 | 9.571187 / 9.144419 |
| 4 | 0.917109 / 0.899976 | 0.799632 / 0.747111 | 11.096163 / 10.699734 | 9.425452 / 9.067095 |
| 5 | 0.919115 / 0.898100 | 0.796108 / 0.779740 | 11.096908 / 10.730139 | 10.026774 / 9.619812 |

One-minute load: parse 0.49–0.57; service 0.57–0.97. Raw JSON retains user,
system and wall time, command, output logs and all three load averages. No other
benchmark, build, test or profiler ran concurrently. Background container
services remained; host isolation and frequency are not independently verifiable.
Hardware cycles/cache/branch counts are unavailable, so user time is the requested
fallback, not a renamed cycle measurement. Earlier unpinned and warmed-marker
wall runs are preserved as supplemental evidence; the table and decision use
the final pinned series.

Ir is whole-process: startup, file reads/UTF-8 decoding, request splitting or
manifest handling, work, output and cleanup. The service is the harness's native
`run` command, including input preparation inside that command. Instructions were
collected separately, not under hyperfine. No sanitizers, `ADAMIC_COUNT`, PGO or
extra optimization was used. Callgrind's nonfatal `brk segment overflow` warning
is retained; each profile completed with exit 0, matching stdout and reconciled
self costs. Raw compressed profiles are committed. Only the separate simulation
build adds `-g`; its `.text` bytes equal the baseline's exactly.

## Build cost and size

These are native backend builds from already emitted C: runtime compilation,
archiving, generated-C compilation and linking. Go compiler construction,
Adamic checking/lowering/emission, downloads and setup are excluded. Cold means
a new empty artifact directory and no `runtime.a`, not flushed OS page caches.
Each workload was cold-built independently, then rebuilt with that same runtime
archive cached. No ThinLTO backend cache was enabled. All four independently
rebuilt binaries are byte-identical to the timed executables.

| Workload | Baseline cold wall/user s | ThinLTO cold wall/user s | Baseline cached wall/user s | ThinLTO cached wall/user s | Baseline bytes | ThinLTO bytes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Parse | 5.102981 / 4.481244 | 19.448743 / 18.735153 | 2.521755 / 2.415769 | 16.076042 / 16.489838 | 971,936 | 1,339,432 |
| Service | 2.705089 / 2.105590 | 2.270272 / 2.335112 | 0.207710 / 0.179252 | 0.458260 / 1.047572 | 406,816 | 163,056 |

The ordinary archive is 511,640 bytes; the ThinLTO archive is 667,544 bytes
(+30.5%). Parse executable size grows 37.8%; service size falls 59.9%.
Size is complete unstripped ELF file length, consistently without `-g`.
Runtime frontends defer machine-code generation with ThinLTO; that work moves
into the link and is paid again for a new generated program.

The final independent cold/cached builds ran with no concurrent benchmark,
test or profiler, at one-minute load 0.94–0.98. Builds use ordinary toolchain
parallelism, not core-3 execution pinning; ThinLTO backend workers can make
aggregate child user time exceed wall time. Every command and child user/system
time is recorded. Earlier builds with one Callgrind process in the background
are retained as supplemental evidence, not the primary table. All final fresh
binaries are byte-identical to the measured binaries.

## Byte parity and mutants

Every measured stdout and ordinary stderr matched its baseline. Parse count
mode prints only `0\n`, so it is insufficient as an AST correctness check.
Additional baseline and ThinLTO builds of the whole-tree parser driver matched both
independent typescript-go and source on Node for all 77 files:
**44,766,682 identical bytes**, SHA-256
`8ae015600498b915cc25abab82730299451ae990b50478980d5a3bc465801bfe`.
This uses the same parser/scanner but a separate AST-printing driver; it does
not claim that the count-only Context driver's internal parent array is fully
observed by its stdout.

Both service variants matched source on Node for **every full response**,
8,018,694 JSONL bytes (SHA-256
`7fa190641de38a460fcc666516e219969cef8f2adda92a0c6082a8f9cfdb384b`),
plus the final checksum line, 8,018,702 compared bytes. Checksum is 7,394,547
UTF-16 units. The original four independent service semantic pins passed too.
A checksum alone is not accepted as evidence of byte parity.

| Mutant actually compiled/run | Intended catcher | Observation |
| --- | --- | --- |
| AST generated C changes the SourceFile kind to XourceFile | Whole-AST byte comparison against Go/Node | Exit 0, equal 44,766,682-byte lengths; first differing byte 14 |
| Service generated C changes health status 200 to 201 | Full response comparison against Node | Exit 0, equal 8,018,702-byte lengths and unchanged checksum; first differing byte 12 |
| Callgrind summary only increased by 1 | Self-cost reconciliation | `callgrind self costs do not sum to summary` |

The simulation footer's final event increased by 1 in memory was rejected by
event-vector reconciliation. Changing the pinned captured output line from 0
to 1 was rejected by the exact stdout/stderr verifier used by the harness.
All 40 real pinned-run output captures passed that same verifier.

The two program mutants are isolated scratch ThinLTO executables, compiled with
all release warnings enabled. Neither was stopped by a warning, sanitizer or
refusal. They do not replace the good measured binaries or runtime sources.
These deliberate mutant differences are separate from the good-build comparisons;
no unexpected output difference was observed.

Complete ordinary native and oracle packages passed uncached:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/native ./internal/oracle > /workspace/scratch/release-lto/native-oracle.log 2>&1
```

Native: `ok` 196.198s; oracle: `ok` 187.858s. Formatting, `go vet ./...`, and
`git diff --check` passed with empty logs. These package results use the existing
release/sanitizer policies; they are not a claim that every oracle fixture was
rebuilt with ThinLTO. Workload-specific ThinLTO validation is the AST/service
comparison above. Test output was written to files, never piped.

## Release policy, runtime caching and developer tools

The existing `runtimeKey` already hashes source/header bytes, every flag,
compiler path/full version, OS and architecture. Adding `-flto=thin` to the
native release compilation policy naturally creates a distinct bitcode archive
cache entry. Use the compatible `llvm-ar` index, retain whole-archive linkage,
and apply ThinLTO to **every generated and runtime C unit**. Linking a new main
against an ordinary cached archive does not test this proposal.

Keep linker selection in the link policy: `-fuse-ld=lld` on this Linux toolchain,
or a verified platform ThinLTO linker. Do not put linker-only options into the
runtime's `clang -c` warning-strict command. Check toolchain support explicitly
and provide an explicit non-LTO release override. The sanitizer branch remains
its existing `-O1 -g -fsanitize=address,undefined` policy. WASI needs its own
measurement and support decision.

A cached `runtime.a` caches frontend bitcode, not all program-specific backend
optimization. Measure an lld `--thinlto-cache-dir` policy separately, with bounded
storage and version/flag isolation. No backend-cache speedup is claimed here.
Before landing a default, run the complete release oracle with the new policy
and cover supported macOS/Linux toolchains, option/cache separation, count
reporting constructors and unavailable-linker diagnostics.

From the developer-tools translation-unit split, I need the concrete emitted
unit boundaries, exported prototypes/data and compile/archive/link commands;
all units must consume the same release flag/link policy and compatible clang.
Their per-unit object cache must include the ThinLTO flag and toolchain identity.
Shared definitions need correct linkage instead of per-unit duplicate globals;
address identity and initialization order must retain their existing semantics.
ThinLTO can cross those boundaries without a unity build, but splitting the large
program module changes parallelism, import decisions, cache reuse and code size.
I need their actual split artifacts to repeat these two measurements and the
byte/mutant oracles. This report measures one generated program unit plus 48
runtime units, not that future split.

Disassembly confirms fewer named static call sites: parse calls to
`adamic_release` fall from 2,383 to 607, and `adamic_object_new` from 43 to 0;
service calls fall from 187 to 44 and 40 to 4 respectively. These are whole-binary
static sites, not dynamic call counts or an instruction attribution. Inlining
moves runtime self costs into callers. The baseline also already has header
fast paths for field writes, so the aggregate gain must not be presented as
an isolated gain from tonight's four hot calls. There was no lld-only ablation;
these observations concern the requested ThinLTO-plus-lld configuration.

## Exact commands and reproduction

The compiler executable was `/workspace/adamic-tools/llvm/bin/clang`.
The exact common argument list, in order, is:

```text
-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2
```

For every runtime `.c` in lexical order, baseline runs that compiler with those
arguments followed by `-c /workspace/adamic/internal/native/runtime/FILE.c -o
/workspace/scratch/release-lto/baseline/FILE.o`. ThinLTO appends `-flto=thin` before
`-c` and writes to `thin/FILE.o`. Both archive all 48 objects with
`/workspace/adamic-tools/llvm/bin/llvm-ar rcs MODE/runtime.a OBJECTS...`.

The baseline parse link is exactly:

```sh
/workspace/adamic-tools/llvm/bin/clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -I /workspace/adamic/internal/native/runtime -o /workspace/scratch/release-lto/baseline/parse /workspace/scratch/release-lto/parse.c -Xlinker --whole-archive /workspace/scratch/release-lto/baseline/runtime.a -Xlinker --no-whole-archive -lm
```

The ThinLTO parse link is exactly:

```sh
/workspace/adamic-tools/llvm/bin/clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -flto=thin -fuse-ld=lld -I /workspace/adamic/internal/native/runtime -o /workspace/scratch/release-lto/thin/parse /workspace/scratch/release-lto/parse.c -Xlinker --whole-archive /workspace/scratch/release-lto/thin/runtime.a -Xlinker --no-whole-archive -lm
```

Service links replace `parse`/`parse.c` with `service`/`service.c` only.
[commands.log](evidence/commands.log) and [cold-commands.log](evidence/cold-commands.log)
contain every expanded command, including all runtime units. Runtime snapshots
are hashed in [runtime.sha256](evidence/runtime.sha256); generated C, binaries,
archives and compared outputs in [artifact-hashes.json](evidence/artifact-hashes.json).

```sh
source /workspace/adamic-tools/env.sh
bash cloud/reports/release-lto/prepare.sh /workspace/scratch/release-lto > /tmp/release-lto-prepare.log 2>&1
python3 cloud/reports/release-lto/build.py /workspace/scratch/release-lto > /tmp/release-lto-build.log 2>&1
# Extract Valgrind 3.24.0 under scratch/valgrind, as above.
python3 cloud/reports/release-lto/measure.py /workspace/scratch/release-lto > /tmp/release-lto-measure.log 2>&1
# Extract hyperfine 1.19.0 and perf 6.12.107 under scratch/tools.
(cd cohere && go build -overlay=/workspace/scratch/release-lto/parse-overlay.json -o /workspace/scratch/release-lto/go-parse /workspace/adamic/cohere/adamic_parse.go) > /tmp/release-lto-go-build.log 2>&1
python3 cloud/reports/release-lto/instrument.py /workspace/scratch/release-lto > /tmp/release-lto-instrument.log 2>&1
python3 cloud/reports/release-lto/cache-profile.py /workspace/scratch/release-lto > /tmp/release-lto-cache.log 2>&1
python3 cloud/reports/release-lto/rerank.py /workspace/scratch/release-lto > /tmp/release-lto-rerank.log 2>&1
python3 cloud/reports/release-lto/validate.py /workspace/scratch/release-lto > /tmp/release-lto-validate.log 2>&1
python3 cloud/reports/release-lto/summarize.py /workspace/scratch/release-lto > /tmp/release-lto-summarize.log 2>&1
python3 cloud/reports/release-lto/cold-build.py /workspace/scratch/release-lto isolated > /tmp/release-lto-isolated-build.log 2>&1
```

Callgrind invocation, repeated for each measured executable with the workload's
arguments, is:

```sh
VALGRIND_LIB=/workspace/scratch/release-lto/valgrind/usr/libexec/valgrind /workspace/scratch/release-lto/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/release-lto/baseline-parse.callgrind /workspace/scratch/release-lto/baseline/parse --manifest /workspace/scratch/release-lto/compiler.txt --count > /workspace/scratch/release-lto/baseline-parse-profile.stdout 2> /workspace/scratch/release-lto/baseline-parse-profile.stderr
```


Tool acquisition URLs and package versions/checksums are in
[evidence/measurement-tool-urls.txt](evidence/measurement-tool-urls.txt) and
[evidence/measurement-tools.json](evidence/measurement-tools.json).
Extract hyperfine/perf/dependencies with dpkg-deb -x into scratch/tools and
Valgrind into scratch/valgrind. perf requires
LD_LIBRARY_PATH=/workspace/scratch/release-lto/tools/usr/lib/x86_64-linux-gnu.
Run each helper from the repository root in a fresh scratch artifact directory;
cache-debug and isolated-* must not already exist.
The exact actual counter attempts were:

```sh
GOMAXPROCS=1 /workspace/scratch/release-lto/tools/usr/bin/perf stat -r 10 -e cycles,instructions,cache-misses,branch-misses taskset -c 3 /workspace/scratch/release-lto/baseline/parse --manifest /workspace/scratch/release-lto/compiler.txt --count > /workspace/scratch/release-lto/perf-native.stdout 2> /workspace/scratch/release-lto/perf-native.stderr
GOMAXPROCS=1 /workspace/scratch/release-lto/tools/usr/bin/perf stat -r 10 -e cycles,instructions,cache-misses,branch-misses taskset -c 3 /workspace/scratch/release-lto/go-parse --manifest /workspace/scratch/release-lto/compiler.txt --count > /workspace/scratch/release-lto/perf-go.stdout 2> /workspace/scratch/release-lto/perf-go.stderr
```
The unsupported hardware events prevented an interleaved hardware-counter series;
the final hyperfine fallback is interleaved. The actual-command perf attempts
above are diagnostic only and are not the timing series used for decisions.
