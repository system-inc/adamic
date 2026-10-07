# Flow lookup speed report

Build clarification: every uninstrumented native throughput number below uses
clang `-O2`, without `-g`, sanitizers or `-DADAMIC_COUNT`. Diagnostic phase
builds use `-O2 -g` and scratch timers; correctness builds use
`-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`. Go uses its
default optimized build. The historical literal clang commands were not
captured. [The explicit build rerun](BUILD_FLAGS_REPORT.md) records actual
clang argv and a release/sanitized comparison on the identical compiler input.

The hot compiler runner falls from **65.449756 s to 61.539752 s**, a **6.0%**
reduction in three isolated, alternating trials. Production Go takes 7.443676 s.
The earlier unit measured 68.427 s / 7.768 s process times; the fresh matched
baseline is the comparison here. Native remains 8.27 times Go for this runner.

The 68-second workload is `coverage_suite.ts`, containing the ten newest rules.
The other sixteen run in `volume_suite.ts`. Their process medians are
17.599842 s before, 17.404761 s after, and 4.264157 s Go. Summing the two existing
runners gives **83.049598 s before, 78.944513 s after, 11.707833 s Go** for all
26 rules. This sum contains two program loads. Each executable loads once and
retains its program across every rule and all 77 compiler files. No combined
26-rule executable was introduced in this speed unit.

All **30,821 compiler findings and 226 repository findings**, including fixes
and suggestions, agree byte for byte with production Go and the previous native
streams on unchanged inputs. The repository corpus is now an archive of this
unit's starting commit, `0d540f413625f016f20fea39761c7b184f335de6`, with the
same 287 roots. Raw snapshot output agrees directly between Go, native and
ASan. Comparing it to the old output replaces only the snapshot directory
prefix with `/workspace/adamic`; offsets, messages, fixes and suggestions are
untouched. [findings.json](validation-coverage-speed/findings.json) records both
raw and canonical hashes. The existing compressed golden streams remain in
`validation-coverage` and `validation-volume`.

This freeze matters: the live-source gate still has 46 old-rule findings, but
editing `Types.ts` moves its existing `consistent-return` report from bytes
668–672 to 843–847 in that gate's inputs. Its Go/native agreement alone cannot
prove agreement to an older input's offsets. The old-stream hash comparison
caught this. The frozen-source comparison passes. No report offsets are
normalized to hide a source edit.

Linux profiling uses Go's CPU sampler, which records native PCs, `nm` leaf
symbol resolution, and monotonic wall timers in scratch generated C. Linux
`perf` is not installed. No compiler or runtime source was instrumented.
Native CPU samples do not supply inclusive C call stacks; the phase timers
supply inclusive and nested wall intervals. Profiling data and release timing
are distinct measurements.

The top three after-load costs are **flow comparisons, adapters, and unused
variable analysis**. Query and graph-decoder time is subtracted from each rule
before its own time is reported:

| Diagnostic wall interval | Before | After |
| --- | ---: | ---: |
| Flow's own work | **16.095 s** | **11.758 s** |
| Query adapters, including checker and fact production | **14.224 s** | **14.650 s** |
| Unused-variable rule's own work | **14.142 s** | **14.590 s** |
| Before-use rule's own work | 12.841 s | 12.813 s |
| Type-graph decoding and validation | 6.583 s | 5.623 s |
| Adamic parsing | 0.938 s | 0.940 s |
| Five binding/span indexes | 1.545 s | 1.627 s |

Flow's recursive comparison accounts for 14.829 s of its 16.095 s own interval
before, and 10.726 s of 11.758 s afterward. The remaining AST dispatch is much
smaller. The flow cost falls about **27%**. These are diagnostic snapshots;
CPU sampling was off in the detailed before run and on in the after run.
They identify costs and support the mechanism; the isolated release medians
above measure the overall improvement. The initial coarse timer run independently
puts flow at 15.766 s, unused analysis at 14.274 s, and adapters at 14.005 s.
[phases.json](validation-coverage-speed/phases.json) preserves every interval.

The initial CPU sample attributes 7.17 s to release, 5.05 s to retain, 4.40 s
to map lookup, 3.95 s to string equality, 3.87 s to array index search, and
2.46 s directly to `Types.type`. Allocation itself has 1.09 sampled seconds;
libc has additional unattributed samples. These are leaf CPU samples, not
another set of additive wall costs. Repeated record scans retain and release
each visited fact, so cutting scans also cuts reference-count work.

The matched boundary diagnostic makes **2,222,043 queries** and returns
**245,444,486 UTF-8 fact bytes**. Its C public-call interval is 13.903838 s;
timed Go bodies take 13.163743 s. The difference is **0.740095 s**, about
**0.333 microseconds per query**, including cgo, registry, copy and timer work.
It is not a pure cgo latency measurement. Native input conversion takes
0.300454 s and output conversion/freeing 0.554989 s. These costs are much smaller
than the flow comparison cost. Afterward the question count and fact-byte count
are identical; Go bodies take 12.739878 s, C calls 13.496558 s, input conversion
0.293309 s and output conversion/freeing 0.582704 s. Questions were not batched.

The implementation adds a stable sorted copy of each completed, read-only type
graph and uses lower-bound search followed by an exact ID check. Original wire
order stays intact, and a duplicate ID still selects the first record. Missing
IDs still panic. Property arrays gain a string-to-first-slot map, preserving
`indexOf` semantics, including duplicate and empty names. Flow's visited-pair
list becomes a map using the same source/target key. It is cleared for every
mutable/optional judgment and every site. Target property traversal order,
recursion depth, flags, relation questions, type rendering and diagnostic
construction are unchanged. No checker question or C ABI was added. These
indexes retain additional native storage until their owning graph or file dies;
ASan/UBSan/LSan checks cover their release. Native peak RSS was not measured.

Release measurements, seconds unless the column says otherwise:

| Runner / corpus / implementation | Load | After-load run | Process | Mean adapter query µs | Findings/process s |
| --- | ---: | ---: | ---: | ---: | ---: |
| coverage-compiler / before | 0.302 | 65.092 | 65.450 | 6.225 | 253.5 |
| coverage-compiler / native | 0.298 | 61.172 | 61.540 | 6.199 | 269.6 |
| coverage-compiler / go | 0.295 | 7.112 | 7.444 | - | 2228.6 |
| volume-compiler / before | 0.291 | 17.266 | 17.600 | 9.029 | 808.6 |
| volume-compiler / native | 0.305 | 17.067 | 17.405 | 9.246 | 817.7 |
| volume-compiler / go | 0.287 | 3.940 | 4.264 | - | 3337.6 |
| coverage-repository / before | 0.085 | 1.468 | 1.564 | 7.024 | 115.1 |
| coverage-repository / native | 0.084 | 1.496 | 1.586 | 7.084 | 113.5 |
| coverage-repository / go | 0.085 | 0.236 | 0.336 | - | 535.7 |
| volume-repository / before | 0.084 | 1.165 | 1.261 | 8.362 | 36.5 |
| volume-repository / native | 0.082 | 1.151 | 1.244 | 8.127 | 37.0 |
| volume-repository / go | 0.076 | 0.219 | 0.306 | - | 150.4 |

Query averages divide the median summed adapter interval by the unchanged
question count. They include checker work and C/UTF conversion, and exclude
native frame decoding and rule work. Compiler counts are 2,222,043 for coverage
and 854,525 for volume; frozen repository counts are 72,186 and 65,198. Go's
production listeners do not execute the bridge questions, so their run times
are not divided by native query counts. Medians of phases need not add to the
process median. The repository sum is 2.824701 s before and 2.830000 s after,
about **0.2% slower**; this unit shows a compiler-corpus improvement.

The initial coverage compiler rounds 2 and 3 overlapped the additional frozen
repository validation. They are saved but excluded. Accepted coverage compiler
measurements are initial round 1 and isolated repeats of rounds 2 and 3. The
volume compiler rounds and both frozen repository sets are three isolated
rounds. Ordering alternates Go/native/before and before/native/Go. No builds or
checks ran during accepted repeat rounds. All raw results and the accepted
selection are saved beside [medians.json](validation-coverage-speed/medians.json).

The fresh full gate passes normally and under ASan/UBSan/LSan: compiler coverage
7,120,228 bytes / 16,589 findings; compiler volume 6,717,107 / 14,232; frozen
repository coverage 85,151 canonical bytes / 180; frozen repository volume
31,862 / 46. Generated controls produce 59 findings and 27,302 identical bytes
normally and under sanitizers. Six compiled index mutants are checked:

| Mutant | Catch |
| --- | --- |
| Return record 0 after finding another ID | Go finding oracle, byte 1435; exits 0, 51 findings |
| Reverse the stored property slot | Go finding oracle, byte 13288; exits 0, 60 findings |
| Carry visited pairs into later judgments | Go finding oracle, byte 15396; exits 0, 54 findings |
| Keep the last duplicate property slot | Direct first-match expectation, slot 2 instead of 0; exits 0 |
| Remove the exact ID equality guard | Missing-ID expectation; exits 0 with ID 11 instead of panic 70 |
| Choose the midpoint past the array end | Panic 70, `missing checker type index` |

Sparse IDs, MAX_SAFE_INTEGER, duplicate IDs, duplicate/empty property names,
unchanged wire order and repeated widening sites are explicit controls.
The frozen-input helper also refuses a reused snapshot directory (exit 1),
and a one-byte wrong golden stream is caught by its exact-byte assertion
(exit 1). Their stderr is saved. Nineteen existing frame/schema/integer/link mutants still produce the required
panics. The first three direct-probe attempts were refused because `console.log`
requires a string and stage 0 cannot lower string-plus-number expressions.
The final probes use `toString`; those compile refusals were not counted as
mutant kills. The earlier 14.78 s adapter estimate in commentary was an
arithmetic error; the saved coarse measurement is 14.005 s.

The unit starts and ends on `codex/tsgo-c-library`. Go 1.27.1, clang 20.1.8,
Node 24.19.0, Linux x86_64. `bash cloud/setup.sh` succeeds: Go ready 0s, clang
ready 1s, Node ready 1s, submodules ready 1s, build cache warm 21s, total 21s.
`nproc` is **5**; cgroup quota is four CPUs and memory is 17.6 GB. Every shell
sources `/workspace/adamic-tools/env.sh`. Workspace scratch avoids the inherited
nearly-full `/tmp`. C-archive works on this toolchain.

Key commands (all test/process stdout and stderr go directly to saved files):

```sh
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/tsgo-coverage-scratch
# Phase build: put a copy of clang_phases.sh named clang first on PATH.
ADAMIC_PROFILE_CLANG=/workspace/adamic-tools/llvm/bin/clang \
ADAMIC_PHASE_INSTRUMENT=/workspace/adamic/bridge/tsgo/profile/coverage_phases.py \
PATH=/workspace/tsgo-speed/wrapper:$PATH \
/tmp/tsgo-coverage/adamic build stage1/cohere/typeaware/coverage_suite.ts \
  -o /workspace/tsgo-speed/after-phases \
  --tsgo /workspace/tsgo-coverage-scratch/final-controls/checker.a
ADAMIC_TSGO_TIMING=1 ADAMIC_TSGO_PROFILE=/workspace/tsgo-speed/after.pprof \
/workspace/tsgo-speed/after-phases /tmp/tsgo-typescript/src/compiler/tsconfig.json \
  /tmp/tsgo-profile/final/compiler.manifest --count
# findings 16589; phase/Go/C counters saved in after-phases.stderr.
ADAMIC_SPEED_ARTIFACTS=/workspace/tsgo-speed/validation-pass \
ADAMIC_SPEED_REPOSITORY_MANIFEST=/tmp/tsgo-volume/repository.manifest \
ADAMIC_SPEED_COMPILER_MANIFEST=/tmp/tsgo-profile/final/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript \
go test ./stage1/cohere/typeaware -run '^TestFlowIndexAgreementAndMutants$' \
  -count=1 -timeout=30m -v
# PASS 549.105s, both runners/corpora normal and ASan, five mutants.
ADAMIC_SPEED_BOUNDS_ARTIFACTS=/workspace/tsgo-speed/bounds \
go test ./stage1/cohere/typeaware -run '^TestFlowTypeIndexBoundsMutant$' \
  -count=1 -timeout=10m -v
# PASS 17.323s; past-end midpoint panics 70.
python3 bridge/tsgo/profile/frozen_repository.py /workspace/adamic \
  0d540f413625f016f20fea39761c7b184f335de6 \
  /tmp/tsgo-volume/repository.manifest /workspace/tsgo-speed/validation-pass \
  /workspace/tsgo-speed/frozen-final \
  --baseline coverage /tmp/tsgo-coverage/final-repository-native.stdout \
  --baseline volume /tmp/tsgo-coverage/old16/062-repository-native.stdout
# Fresh archive; all six outputs match Go and previous canonical bytes.
python3 bridge/tsgo/profile/volume_bench.py \
  /workspace/tsgo-speed/validation-pass/coverage \
  /workspace/tsgo-speed/validation-pass/coverage-oracle \
  /workspace/tsgo-speed/bench-coverage-repeat \
  --before /workspace/tsgo-coverage-scratch/final-controls/coverage \
  --corpus compiler /tmp/tsgo-typescript/src/compiler/tsconfig.json \
  /tmp/tsgo-profile/final/compiler.manifest --rounds 2 --start-round 2
# Counts identical in every round; medians and accepted-rounds.json saved.
go test ./stage1/cohere/typeaware \
  -run '^(TestFactsDecoderGuards|TestPinnedTypeFlags|TestSixPinnedFlags)$' \
  -count=1 -timeout=10m -v
# PASS 5.639s; 19 decoder mutants.
go vet ./stage1/cohere/typeaware
gofmt -l stage1/cohere/typeaware/speed_test.go
git diff --check
# Empty output.
```

The manual production-cohere lint attempt on the three changed TypeScript files
reports 92 findings across their 142-file dependency set, including three files
it would rewrite for formatting/import style. That is not a passing lint gate;
its output is saved. No protected emitter, lowerer, native driver or oracle-test
file was edited. Full `go test ./...`, native peak-memory profiling, concurrent
programs and additional rule/configuration coverage are outside this unit's gate.
The existing bridge ownership/question mutants were not all rerun: the ABI,
Go question implementations and runtime are unchanged.
