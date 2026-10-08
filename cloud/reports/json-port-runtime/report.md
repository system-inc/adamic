# JSON port: serial allocation was scanning remote-free queues

Base: `origin/area/runtime`, fetched by name at `c1c6073021e8e5cd6005fcdee59eb3b377395fd8`. The native entry is `stage1/cohere/json/main.ts`. The only production change is in `internal/native/runtime/heap.c`; generated C is byte-for-byte unchanged (identity in [manifest.txt](manifest.txt)).

The allocator's `take` drained a giving chunk on every small allocation. When a size class had no giving chunk, it walked **every owned chunk**, exchanging each remote-free queue. This serial formatter had only one heap thread, so all those queues were empty. As large documents grew the slab pool, the scans became the dominant cost. The fix skips this work until the existing monotonic heap-thread counter exceeds one. After a second thread gets an identity, the original remote-draining behavior remains enabled. Cleanup still drains unconditionally. A concurrent first remote free can miss a particular allocation pass, just as it can arrive after an exchange; a later allocation or cleanup reclaims it.

## Self CPU split over the complete corpus

| Innermost source attribution | Before | After |
|---|---:|---:|
| Generated C | 3.52% | 14.81% |
| `internal/native/runtime` | 91.79% | 65.21% |
| External libraries | 4.58% | 19.49% |
| Unresolved native source | 0.11% | 0.49% |

These are percentages of sampled **user CPU**, not wall-time fractions or inclusive call costs. After shares have a much smaller denominator. Standard-library work stays separate, rather than being assigned speculatively to its caller. Most external samples lack private libc symbols; that is an attribution limitation, not missing samples. All self samples and periods are accounted for. Kernel/I/O wait time is outside this user-CPU split.

## Top runtime functions and requested cost classes

Self percentages below use the total sampled user CPU denominator, including generated and external code. LLVM clone suffixes are omitted for readability.

| Cost | Function | Before | After |
|---|---|---:|---:|
| Allocation / remote-free polling | `drain_remote` | 76.519% | 0.018% |
| Allocation | `take` | 0.337% | 1.962% |
| Allocation | `allocate_storage` | 0.337% | 1.273% |
| Array lookup | `adamic_array_at` | 2.686% | 11.028% |
| Retain | `adamic_retain` | 1.218% | 5.232% |
| Retain slow path | `adamic_retain_slow` | 0.043% | 0.247% |
| String unit read | `adamic_string_char_code_at` | 1.008% | 4.480% |
| String equality | `adamic_string_equal` | 0.899% | 3.729% |
| Release / field teardown | `release_field` | 0.824% | 3.579% |
| Ownership header resolution | `counted_heap` | 0.742% | 3.305% |
| String building | `adamic_array_join` | 0.408% | 1.635% |
| String slicing | `adamic_string_slice` | 0.386% | 1.697% |
| Release | `adamic_release` | 0.328% | 1.414% |
| Release last reference | `release_last` | 0.180% | 0.716% |
| String building | `adamic_string_concat` | 0.004% | 0.035% |
| Map lookup | `adamic_map_get` / lookup helpers | no samples | no samples |
| Number formatting | `adamic_number_format` / formatting helpers | no samples | no samples |

Map lookup and number formatting are not measured hotspots in this port. The printer preserves numeric lexemes; number parsing is distinct from formatting. Zero samples does not assert zero possible cost. The full function and source-location rankings are in [before-split.txt](before-split.txt) and [after-split.txt](after-split.txt).

## Profiling method

The shipped binaries were built with:

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic build stage1/cohere/json/main.ts -o /tmp/json-port-runtime/before
# Apply the runtime change, then the same command with -o .../after.
go run ./cmd/adamic c stage1/cohere/json/main.ts > /tmp/json-port-runtime/after.c
cmp /tmp/json-port-runtime/before.c /tmp/json-port-runtime/after.c
```

This is the shipped CLI release build: `-O2`, ThinLTO, `-ffp-contract=off`, `-fno-optimize-sibling-calls`, no sanitizers or allocation counters. A first `perf record -g` of the actual before release put 76.80% of self samples in `allocate_storage`. Its requested cycles event was recorded as `task-clock:uH` by this environment; no hardware-cycle conclusion is drawn from it.

For accurate source attribution through ThinLTO, profiling twins retained all those release flags and appended `-g` at compilation and link, using a scratch clang wrapper. They were not used for the final wall-time measurements. Each twin ran the complete corpus and matched Go's exact bytes. Canonical profiles used:

```sh
perf record -g --call-graph dwarf,16384 -e cpu-clock:u -F 499 \
  -o before.data -- /tmp/json-port-runtime/profile-before --cases /tmp/json-port-runtime/cases.txt
perf report -i before.data --stdio --no-children --sort overhead,symbol \
  --call-graph none --percent-limit 0
# Repeat identically for profile-after / after.data.
```

Before: 46,616 samples; after: 11,316; zero lost samples in both. [before-self.txt](before-self.txt) and [after-self.txt](after-self.txt) are the requested self-time-sorted perf symbol reports. The before twin's symbol report places 77.21% in `allocate_storage`; inline source attribution identifies `drain_remote` as 76.52% of all samples, or 83.37% of runtime samples.

A plain symbol-prefix split would incorrectly count inlined runtime work as generated code. [split.py](split.py) reads only each sample's self IP/period (`perf script -G`), checks its executable relocation against ELF symbols, and uses `llvm-symbolizer`'s innermost DWARF inline frame to classify generated `main.c`, cached runtime sources/headers, external DSOs, or unresolved locations. It requires every sample line to parse and all weighted periods to sum to the recorded total. It never adds parent call-chain costs. Some source locations have DWARF line zero; their file/function still identifies the runtime. Tiny unresolved native regions are kept explicit. Sampling and optimized debug attribution give estimates, not an exact instruction-cost decomposition.

Reproduce source accounting with:

```sh
python3 cloud/reports/json-port-runtime/split.py before.data \
  /tmp/json-port-runtime/profile-before before-split.txt \
  --perf /tmp/nbody-speed/tools/usr/bin/perf --llvm /workspace/adamic-tools/llvm/bin
```

## Corpus and timings

The tests' `corpusCases`, `cohereAnswers`, `protocol`, and `buildGoDriver` helpers prepared the full corpus in scratch. A temporary preparation test was removed. No sampling environment was set. The corpus pin passed after provisioning its 18 locked npm files with `npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund`.

There are 2,790 texts: 1,696 tracked repository JSON inputs, 839 cohere inputs, 203 TypeScript inputs, 18 provisioned inputs, and 34 generated boundaries. The escaped input is 161,125,991 bytes; Go's answer stream is 163,928,745 bytes. This is larger than the corpus in the older stage1 performance report; that report's timings were not reused. Input/output digests and all raw timing samples are recorded in [measurements.txt](measurements.txt).

| Build | Best wall seconds | User seconds at that sample | All five wall samples (seconds) |
|---|---:|---:|---|
| Native before | 93.413982 | 92.698937 | 97.632 / 94.258 / 93.414 / 95.417 / 100.802 |
| Native after | 21.429778 | 20.665521 | 22.483 / 21.884 / 22.435 / 22.533 / 21.430 |

Best-of-five speedup: **4.36×**, **77.1% less wall time**. All five pairs improved. Load averages (1/5/15 minutes) before: 1.290 / 1.586 / 1.167; after: 1.030 / 1.082 / 1.077. Every timed output and the standalone Go-driver validation matched the full reference exactly.


[measure.py](measure.py) runs five fresh-process pairs, alternating before/after and after/before. Wall time uses Python `perf_counter` around launch and wait; user CPU uses `getrusage(RUSAGE_CHILDREN)`. Best means minimum wall time, with the user time from that same sample. Reading the corpus, formatting/refusal, protocol escaping and writing the complete answer stream to a regular file are included. Builds and subsequent byte comparison are excluded. Every timed run must exit successfully, write empty stderr, and match Go byte for byte. A standalone run of the Go driver also validates the saved reference. Both native releases already passed an untimed full-corpus parity check before timing.

No setup, builds, tests or profiles ran alongside the final timing pairs. No CPU pinning or isolation; `ADAMIC_THREADS`, `NODE_OPTIONS`, and `BUN_OPTIONS` are cleared. Machine: `3ee19de862f2`, AMD EPYC 9V74 80-Core Processor, Linux 6.18.44 x86-64/glibc 2.41, 17.6 GB RAM, affinity CPUs 0–4, cgroup quota `400000 100000` (four CPUs). Go 1.27.1, clang 20.1.8, Node 24.19.0, perf 6.12.107. Load averages before/after and after each pair are in the measurements file.

## Validation

No AGENTS.md was found in the repository or workspace ancestors. `area/runtime` was fetched by name before branching `runtime/json-port-runtime`. `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh --wasi-sdk` completed, and its printed `/workspace/adamic-tools/env.sh` was sourced. Perf and its shared libraries were reused from scratch, with `LD_LIBRARY_PATH=/tmp/nbody-speed/tools/usr/lib/x86_64-linux-gnu`; no system package changes were required.

Passed:

```sh
go test ./internal/native -run '^TestSerialSlabsSkipRemoteScans$' -count=1 -v
go test ./internal/native -run '^(TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestQuarantineCatchesAStaleReadOfAReusedSlot|TestParallelMemory)$' -count=1 -v
ADAMIC_JSON_PROFILE_BINARIES=/tmp/json-port-runtime/before:/tmp/json-port-runtime/after \
  go test ./stage1/cohere/json -run '^TestProfileSnapshotsAgree$' -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts
```

The new C allocator fixture fills several slabs in one thread, requires zero remote-drain calls, releases those cells in another thread, and requires their slabs to be reused with correct values. Tests inject a drain-call counter into a scratch runtime; production has no new counters. The fixture passes both unsanitized and ASan/UBSan slab lanes. Its always-scan mutant fails the zero-scan requirement; its never-scan mutant fails remote slot reuse, both with nonzero exits. See [mutants.txt](mutants.txt).

[Runtime checks](runtime-tests.txt) additionally cover poisoned freed values, slab sharing, quarantine, parallel ownership, malloc and counted modes, and three ThreadSanitizer runs. [Corpus parity](corpus-parity.txt) checks all output bytes against Go. Counts regeneration passed (75.787 s) and left `counts.md` unchanged, so no counts-only commit was needed. No oracle fixture registrations changed. No whole-package tests or full gate were run. Python AST checks and `git diff --check` passed.

Reproduce final measurements after preparing the same corpus and before/after/Go binaries:

```sh
python3 cloud/reports/json-port-runtime/measure.py \
  /tmp/json-port-runtime /tmp/json-port-runtime/measurements.txt
```
