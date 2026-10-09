# Six type-aware rules, one program

This extends the first unary-minus pilot with five native Adamic ports of the
pinned production cohere rules, under their default options:

| Added rule | Required checker facts |
| --- | --- |
| related-getter-setter-pairs | Getter-to-setter assignability |
| no-unsafe-declaration-merging | Named and local symbol declarations |
| no-unsafe-argument | Resolved call/new/tagged signature, parameter/rest types, reference arguments |
| restrict-plus-operands | Constrained/widened types, unions and intersections |
| no-unnecessary-boolean-literal-compare | Constrained boolean/nullable types, strict-null option |

The additive `tsgo_inspect` API returns facts, never findings. Its exact selector,
framing, stable program-local identities and ownership are in
[the protocol](../../../bridge/tsgo/facts.md). Native Adamic decides every rule
predicate, message, span and fix. One program loads all roots. Each file is
parsed once by the merged Adamic TypeScript parser, indexed for parents, and
walked once for all six rule visitors. Go cohere independently loads its program
and runs the six unchanged production rules in one shared AST walk.

## Pins and environment

- Starting branch head: `73f3e125bb503c19b4d07d6ff61e69dcc1484247`.
- cohere: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
- typescript-go: `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
- TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`,
  all 77 recursive `src/compiler/*.ts` roots, including `_namespaces`.
- Linux amd64, Go 1.27.1, clang 20.1.8, Node 24.19.0; `nproc` = 5,
  cgroup CPU quota = 4 CPUs, 17.6 GB memory.
- `bash cloud/setup.sh`: Go ready 0s, clang ready 0s, Node ready 0s,
  submodules ready 0s, cache warm 53s, total 53s. Environment sourced from
  `/workspace/adamic-tools/env.sh`. Setup and fetch succeeded.

## Agreement and mutants

The independent oracle imports no bridge implementation. Its production rules
use their declared program views and shared file cache. Comparison includes
file headers, rule/message IDs, escaped messages, byte spans, repair counts and
boolean-fix replacement bytes. Lines are sorted lexically by their complete
canonical representation, preserving duplicates. This is a findings protocol,
not cohere's CLI display, suppression engine or fix application.

408 syntactically valid generated files produced 186 identical findings:
21 unary-minus, 33 getter/setter, 18 declaration merging, 31 unsafe argument,
42 plus operands and 41 boolean compare. The source extractor found 415
candidates; the independent Go parser excluded seven nonsource message strings
or intentionally malformed fixtures. Additional controls cover nonstrict
configuration and cross-file global declaration merging. Cross-file spans
preserve production cohere's current-source trivia trimming, including its
foreign-declaration positions. The corpus produced **1,763 byte-identical
findings**, including fixes, under ASan/UBSan/LSan.

| Real-input mutation | Catcher |
| --- | --- |
| Ask binary result instead of left operand | Production findings oracle |
| Select last declaration instead of earliest | Production findings oracle |
| Trim a foreign declaration against its own file | Production findings oracle |
| Reverse getter/setter assignability | Production findings oracle |
| Resolve first call instead of current call | Production findings oracle |
| Treat union as its root instead of constituents | Production findings oracle |
| Report nullable comparisons forbidden by default allowances | Production findings oracle |
| Query released program | Refused with panic exit 70; retaining live registry entry makes the real query succeed and fails the refusal expectation |
| Facts UTF-8 buffer length + 1 | ASan heap-buffer-overflow |
| Empty/noninteger/negative-length/schema-2 C facts | Native decoding checks, panic exit 70 |
| 19 malformed dynamic decoder inputs | Panic exit 70: framing, trailing data, schema/question, integer canonicality/precision, natural counts, booleans, IDs, tuple sentinel, duplicate records, missing graph links |
| Wrong exact kind / unknown question | Refused with panic 70; deleting each guard makes that query succeed and fails its refusal expectation |

Mutants compile and execute. Compile failures are not counted as catches.
See [suite.log](validation-six/suite.log), [guards.log](validation-six/guards.log)
and [requests.log](validation-six/requests.log) for each observed result.

## Measurements

Three alternating native/Go release runs use the same all-77 manifest, six rules
and count mode. No other builds or test suites ran during timed rounds. Both
loaders initialize the checker pool within load; type resolution remains lazy
and is charged to whichever query requests it. Each process loads once.

Observed medians from the final run:

| Metric | Native Adamic | Production Go cohere |
| --- | ---: | ---: |
| Program load | 0.292830 s | 0.286115 s |
| Run after load | 17.926922 s | 1.739172 s |
| Whole process | 18.270888 s | 2.045863 s |
| Findings/s after load | 98.344 | 1,013.701 |
| Findings/s whole process | 96.492 | 861.739 |
| Native adapter queries | 131,755 | Not instrumented as checker calls |
| Native total adapter time | 10.859240 s | Not equivalent to listener time |
| Native mean adapter query | 82.420 µs | See identical-facts baseline below |

| Warm fact question | Native µs/query | Direct Go µs/query |
| --- | ---: | ---: |
| Assignability | 2.894 | 2.128 |
| Declarations | 5.407 | 5.485 |
| Resolved signature | 14.998 | 15.358 |
| Raw type | 6.405 | 6.322 |
| Nullable type | 17.556 | 17.828 |
| Widened union | 13.716 | 14.288 |
| Compiler options | 1.805 | 1.257 |

See [measurements.log](validation-six/measurements.log) for exact medians and
[measure.py](validation-six/measure.py) to reproduce them from raw logs.
`load_ns` excludes process startup; `run_ns` starts after successful create and
ends before release. Whole-process seconds include startup, loading and exit.
Native `query_ns` sums 131,755 adapter calls including input copies, exact-node
lookup, Go checker/serialization, owned UTF-8 decoding and freeing the C buffer.
Go's `rule_ns` is listener time, not an equivalent checker-query count: dividing
it by native call count would be misleading.

Seven tiny probes repeat an identical fact question 10,000 times, three rounds.
The table subtracts the first query and divides by 9,999, then takes the median.
Direct Go deliberately shares `Program.Inspect` for this cost baseline; it is
not the independent findings oracle. Native adds C copies/registry/owned
conversion; Go consumes UTF-16 unit counts for an identical checksum. Deltas
include this difference and timer noise; they are not pure cgo crossing costs.
Corpus query averages include large type graphs and lazy resolution, so the
small warm probes cannot predict whole-run cost.

Native remains slower than Go. Program loading is a small fraction of native
elapsed time. The residual outside query intervals includes Adamic parsing,
parent indexing, traversal, fact validation and diagnostic handling. No profile
was taken, so assigning that residual to any one operation would be inference.

## Commands and coverage limits

Every test run wrote a log directly, without a pipe. The final corpus command passed in 309.848 s (both tests PASS):

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript \
ADAMIC_TYPEAWARE_BENCH=1 ADAMIC_SIX_ARTIFACTS=/tmp/tsgo-six-final-full \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware \
  -run 'TestSixRuleAgreementAndMutants|TestSixPinnedFlags' > /tmp/tsgo-six-final-full.log 2>&1
```

Additional commands (all with `-v -count=1 -timeout 30m` and direct log output):

```sh
go test ./stage1/cohere/typeaware -run TestFactsDecoderGuards
go test ./stage1/cohere/typeaware -run TestInspectRequestRefusals
go test ./stage1/cohere/typeaware -run 'TestTypeAwareAgreementAndMutants|TestPinnedTypeFlags'
go test ./bridge/tsgo ./internal/load ./internal/lower ./internal/native
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)[.]a$'
go test ./internal/oracle -run TestTheOracleCatchesOneByte
```

Also ran `TestFactsDecoderGuards`, `TestInspectRequestRefusals`, and
`TestTypeAwareAgreementAndMutants|TestPinnedTypeFlags` separately in that package.
The original rule passed 53 files/36 findings and its existing mutants.
`go test -v -count=1 -timeout 30m ./bridge/tsgo ./internal/load ./internal/lower
./internal/native` passed; the foundation retained its 1,600 checker queries
across four compiler files and boundary mutants. The five filtered Node cases
were closures, method_closures, generic_functions, regions and regions_throw;
`TestTheOracleCatchesOneByte` also passed. The first Node filter selected no
Native children; it is retained transparently in `node-one-byte.log`, and the
corrected five-child filter is in `node-filtered.log`.

`gofmt -l cmd internal bridge stage1/cohere/typeaware`, `go vet ./...` and cohere
`--no-fix --no-cache` over all 15 pilot TypeScript files passed. Full `go test
./...` was not rerun: coverage is the touched packages, all pilot tests and the
stated filtered external oracle. Logs are retained in `validation-six/`.

The suite supports only these default rule options. It does not run all cohere
rules, apply fixes, or implement suppression/configuration processing. No npm
corpus dependencies were installed and no semantic-diagnostic pass was
requested; both sides use the same original compiler tsconfig and bundled
libraries. Agreement is established for these roots and valid fixtures, not
claimed for every legal TypeScript program. Go's checker runtime still owns its
managed heap; Adamic's side retains its existing explicit ownership model.
The old first-rule report is preserved and its 1.26 findings/s is a different
workload/measurement, not a directly comparable six-rule speedup.
