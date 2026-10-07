# Where the six-rule time went

The measured bottleneck was eagerly rendering and decoding checker facts, with
allocation costs on both sides. The C crossing was small. Native whole-process
median fell from **18.708858 s to 7.247286 s** in three alternating release
rounds; production Go cohere took **2.069595 s**. All 1,763 findings and fixes
remain byte-identical. Native is still slower than Go.

## Workload and environment

Starting implementation: `c76e1457c868fa8213521adb011cea5514479f7d` on
`codex/tsgo-c-library`. The existing scanner/parser merge is preserved. No
protected emitter, lowering, native-driver or oracle-test file was edited.

Pins remain cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`, typescript-go `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`, and TypeScript
v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`.
All 77 recursive compiler roots, the original compiler tsconfig, the same six
rules and production defaults were used. Both processes load once and initialize
the checker pool during loading; type resolution stays lazy.

Linux amd64, Go 1.27.1, clang 20.1.8, Node 24.19.0. `nproc` = 5, CPU quota = 4,
17.6 GB memory. `bash cloud/setup.sh` succeeded: Go/clang/Node/submodules each
ready in 0s, build cache warm 52s, total 52s. Fetch succeeded. Environment sourced
from `/workspace/adamic-tools/env.sh`. No corpus files or npm dependencies were
changed; neither oracle requests a semantic-diagnostics pass.

## Attribution before changing the implementation

First, an unchanged-suite CPU profile showed 5.11 sampled CPU seconds in
`TypeToString`, 1.99 in fact text framing and 5.09 in background Go collection.
Those are inclusive CPU samples, not additive wall intervals: concurrent Go
collection makes total CPU exceed wall time. Native leaf samples also exposed
retain/release, frame decoding, numeric formatting and allocation.

Then scratch generated C was instrumented at six outermost functions using
[the phase tool](../../../bridge/tsgo/profile/native_phases.py), without editing
the compiler. Recursive walks count only their outermost invocation. Both runs
also enabled Go CPU/allocation profiling and C input/call/output clocks. The
following wall intervals do not overlap; the residual is subtraction, not a
separate timer:

| Work after loading | Before | After |
| --- | ---: | ---: |
| Native query adapters, including Go work | 11.139487 s | 3.560384 s |
| Native `types` fact decoding and validation | 5.269553 s | 1.826940 s |
| Adamic `Parser.file` | 0.953548 s | 0.973678 s |
| Parent indexing | 0.061076 s | 0.060936 s |
| UTF-16 to byte offset table | 0.190856 s | 0.187055 s |
| Residual rule decisions, sorting, I/O and setup/cleanup | 0.775098 s | 0.572818 s |
| Total after-load interval | 18.389618 s | 7.181811 s |

The shared rule walk's inclusive time was 16.993280 s before and 5.846097 s
after; those numbers include queries and fact decoding and must not be added to
the table. Parsing and parent indexing barely moved. The before/after sampled
`TypeToString` accounted for 5.04 CPU seconds before; the after profile recorded
no TypeToString samples at 100 Hz (not a claim of zero work). Go background
collection went from 5.14 to 0.70 CPU seconds. See the retained
[cumulative profiles](validation-profile/phases-before-cumulative.log) and
[after profile](validation-profile/phases-after-cumulative.log) for exact samples.

### Why queries were not batched

Before, C's public-call interval was 10.969247 s; timed Go inspect/parts bodies
accounted for 10.873025 s. The difference was **0.096223 s total**, approximately
**0.730 µs per call**. It includes C input copies, cgo transitions, registry/timer
overhead and scheduling, not a pure cgo microbenchmark. Native input conversion
was another 0.028856 s; output conversion/freeing was 0.124014 s.

Thus the C boundary and marshalling did not account for the sixteen-second gap.
Even recovering their entire measured cost would save only a small fraction of
a second. Batching was not justified by this profile. The optimization instead
eliminates work within each fact request and avoids native decoding allocations.
The optimized suite makes **more** calls, since names are now fetched lazily, and
still runs much faster. Post-change crossing estimate: 0.073516 s total,
0.550 µs/query; native input/output conversion: 0.021922/0.083869 s.

## What changed

- Added `raw-shape`, `type-shape` and `signature-shape` questions to the existing
  `tsgo_inspect` function. They return the same structural facts with empty names
  and skip TypeToString. Existing questions and ABI 1 retain their meanings.
- Added `name` LF type-ID, with exact selector, canonical known-identity and live
  program checks. Adamic requests argument names only while rendering findings.
  Type strings are not cached across checker operations. Plus retains named
  facts because its RegExp arm needs the rendering. The bridge contains no lint
  predicate, message or repair.
- Go counts UTF-16 units without rune/UTF-16 slices and writes numeric frames
  directly, using stack digit buffers rather than formatted printing.
- Adamic decodes numeric fields directly in the owned frame. Canonical syntax,
  safe-integer precision, lengths, schema, graph links and full consumption are
  still checked. String fields retain UTF-16 lengths inside UTF-8 C buffers.
- Metadata searches use direct loops. The shared walk avoids allocating an empty
  unary-finding slice at every node. Finding sort keys are rendered once after
  fixes are complete; `Diagnostic.written()` still renders current field values.
- Added opt-in CPU, Go allocation and C interval profiling. Registry locking
  protects Go profiling state; normal runs do not write profiles or metrics.

Profiling allocation counts, excluding program loading on Go's side:

| Count | Before | After |
| --- | ---: | ---: |
| Go allocated bytes | 3,188,202,792 | 671,209,536 |
| Go allocated objects | 50,572,140 | 4,395,247 |
| Go collections during profiled run | 10 | 5 |
| UTF-8 facts returned | 67,327,486 bytes | 44,681,859 bytes |
| Native allocations, all freed | 61,003,189 | 10,412,957 |
| Native retains | 333,200,401 | 213,028,710 |
| Native releases | 380,025,789 | 211,293,603 |
| Native peak live objects | 735,248 | 735,248 |

These allocation/phase builds are diagnostic experiments, separate from the
release throughput runs. Profiling counters and CPU sampling have overhead.
Go's collector belongs to the external checker; Adamic gains no collector.

## Release measurements

Three alternating before/after/Go processes used the same manifest and count
mode, with profiling disabled and no other builds/tests running. Each process
printed `findings 1763`. Native adapters include lookup, checker work, fact
encoding, C copies, owned UTF-8 decoding and output frees. Whole-process timing
includes startup, load and exit; after-load excludes program release.

| Median | Before native | After native | Production Go |
| --- | ---: | ---: | ---: |
| Program load | 0.298643 s | 0.303877 s | 0.285709 s |
| Run after load | 18.359974 s | 6.921510 s | 1.775663 s |
| Whole process | 18.708858 s | 7.247286 s | 2.069595 s |
| Whole-process findings/s | 94.233 | 243.263 | 851.857 |
| After-load findings/s | 96.024 | 254.713 | 992.868 |
| Native query count | 131,755 | 133,565 | Not counted as checker calls |
| Native query interval sum | 11.240360 s | 3.457084 s | Not equivalent to listener time |
| Mean native adapter query | 85.313 µs | 25.883 µs | See matched facts baseline below |

The final integration harness independently measured 7.439592 s native and
2.036689 s Go medians, with 27.356 µs/native query. The explicit paired run above
is the primary comparison with the previous implementation; the variation is
reported rather than choosing only the fastest run. Median values in a column
can come from different rounds.

Ten small identical-facts probes ran 10,000 queries, three rounds each. Subtract
the first query and divide by 9,999, then take the median:

| Warm query | Native µs | Direct Go µs |
| --- | ---: | ---: |
| Assignability | 2.163 | 1.562 |
| Declarations | 2.678 | 2.634 |
| Full signature | 10.783 | 10.682 |
| Full raw type | 4.043 | 3.647 |
| Full nullable type | 11.405 | 11.288 |
| Widened union | 8.818 | 8.967 |
| Signature shape | 3.674 | 3.412 |
| Raw shape | 2.711 | 2.248 |
| Constrained shape | 4.407 | 4.527 |
| Compiler options | 1.257 | 0.627 |

The direct Go fact baseline intentionally shares Inspect; it isolates adapter
cost and is not the independent findings oracle. It counts identical UTF-16 units
instead of producing an owned Adamic string. Small negative deltas are noise,
not evidence that crossing C is free. These warm, small graphs do not predict
corpus queries with large graphs and lazy resolution. Go cohere's listener time
cannot be divided by native call count to claim a production Go per-query cost.

[measure.py](validation-profile/measure.py) reproduces tables from the retained
[paired benchmark](validation-profile/paired-bench.log), [suite log](validation-profile/suite.log)
and phase logs. Raw CPU profiles are retained too. [bench.py](validation-profile/bench.py)
repeats the nine release processes.

## Agreement, mutants and checks

Production Go cohere independently loads its program and calls all six unchanged
rules. It imports no bridge code. The optimized ASan/UBSan/LSan run matches all
288,482 finding/fix bytes across 77 compiler files. Generated coverage is now
409 valid files, 186 findings with the same per-rule counts as the prior report;
seven nonsource/malformed table candidates are explicitly excluded. Nonstrict
and cross-file global declaration controls also match.

All prior findings mutants were rerun and caught by the byte oracle: wrong node,
last declaration, foreign-source name trimming, reversed assignability, wrong
resolved signature, missing union constituents and changed nullable defaults.
The new wrong-name mutant has the SAME 186 finding count, but different bytes,
and is caught. Constraining a raw argument shape wrongly reports the new
`T extends Array<any>` control (187 instead of 186 findings), also caught.

The first `T extends any` control did not expose that mutation; the test failed
and was strengthened with the reaching array case, whose production oracle
prints zero findings while the mutant prints one. This survivor is retained in
[first-attempt.log](validation-profile/first-attempt.log), not counted as a kill.

Released-program queries are refused with panic exit 70; retaining the released
registry entry makes the query succeed and fails the refusal expectation.
Released type-name queries are also refused. The output-byte-length + 1 mutant
still triggers ASan heap-buffer-overflow. Four malformed C fact payloads and
19 malformed dynamic decoder payloads still panic with their expected messages.
Exact-kind and unsupported-question guards are each proven by removing the guard
and observing that the previously refused query succeeds. New Go tests compare
Unicode/numeric framing and full/shape facts, lazy names, malformed/unknown IDs
and exact-selector refusals.

The six-rule test passed in 245.03 s and the decoder test in 5.74 s. The combined
run then ran out of scratch disk while building the unsupported-question mutant
archive. This is recorded in suite.log, which ends in FAIL. Obsolete archives
created by this thread were removed, preserving all source and log evidence.
The affected `TestInspectRequestRefusals` rerun passed in 16.814 s, including all
three refusal cases. Compilation failure is never counted as a mutant kill.

Commands, each with output written directly to a log, without piping a test:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript ADAMIC_TYPEAWARE_BENCH=1 \
ADAMIC_SIX_ARTIFACTS=/tmp/tsgo-profile/final \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware \
  -run 'TestSixRuleAgreementAndMutants|TestSixPinnedFlags|TestFactsDecoderGuards|TestInspectRequestRefusals'
ADAMIC_SIX_REQUEST_ARTIFACTS=/tmp/tsgo-profile/requests-final \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run TestInspectRequestRefusals
ADAMIC_TSGO_CORPUS=/tmp/tsgo-typescript \
go test -v -count=1 -timeout 30m ./bridge/tsgo ./bridge/tsgo/checker ./internal/native
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware \
  -run 'TestTypeAwareAgreementAndMutants|TestPinnedTypeFlags'
go test -v -count=1 -timeout 30m ./internal/oracle \
  -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)[.]a$|TestTheOracleCatchesOneByte'
go vet ./...
gofmt -l cmd internal bridge stage1/cohere/typeaware
/tmp/tsgo-lint-cohere --no-fix --no-cache stage1/cohere/typeaware/*.ts stage1/cohere/typeaware/testdata/*.ts
```

Bridge regressions passed in 95.403 s, checker tests in 0.022 s, native
regressions in 66.643 s, original unary-minus tests/mutants in 93.247 s, and
the filtered Node oracle in 11.430 s. Vet and gofmt logs are empty; cohere
reports zero findings, 15 of 15 files Adamic-ready.

Full `go test ./...` was not rerun. Coverage is the touched packages, the original
rule tests/mutants, the six-rule integration and guards, and five filtered Node
cases plus its one-byte oracle mutant. Source pins, commands, raw findings,
profiling files and logs are in `validation-profile/`. Default-rule/configuration,
semantic-diagnostic and arbitrary-TypeScript coverage limits from the previous
report still apply. Profiling covers this single-program workload, not concurrent
multi-program use. Native decoding and duplicated Adamic parsing remain visible
costs; removing them or further compressing fact graphs was not attempted here.
