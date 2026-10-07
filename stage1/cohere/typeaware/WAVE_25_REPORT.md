# Type-aware wave 25

Built native Adamic ports of GraphQL nullable parity and matching injection types; accessor-pairs was skipped.
Claim commit: 5864811a, pushed before implementation; base: 0d540f413625f016f20fea39761c7b184f335de6.
Final rule gate: PASS 48.264s; bridge: PASS 64.151s; filtered Node oracle: PASS 27.164s; Go vet and gofmt logs empty.
Rule mutants: remove undefined from GraphQL nullable mask and disable injection brand stripping; Go bytes alone catch both; released registry mutant caught by required panic 70.
Not covered: full repository test gate, other rule options or populations, CLI suppression/edit application, and cohere CLI lint for the new .a files.

## Selection and implementation

The VOLUME_REPORT.md full-table links hold 197 checker-dependent registry rows.
Combine compiler-all.counts and repository-all.counts, sort descending total with
lexical ties, and exclude the baseline ports. The baseline's method-signature-style
is not a checker-dependent row, leaving 172 rows after 25 exclusions. Remaining
positions 73, 74 and 75 are accessor-pairs,
base/correctness-require-graphql-nullable-parity and
base/correctness-require-matching-inject-type. Each has recorded combined volume zero.

All origin branch tips were fetched and distinct stage1/cohere source and claim
blobs inspected. accessor-pairs already exists as grouped_accessor_pairs.ts on
origin/codex/stage1-lint-batch4, tip d486b03a15f3202acc317dc81c40cc21818e21f7.
It is skipped, with no replacement. The other two were available and claimed in
claims/wave-25.md before any source implementation was written.

Each rule owns its .a file. decorator_shape.a reads syntax; wave_25_suite.a loads
one checker program, parses each source once, indexes parents and runs both rules.
All decisions, messages and spans live in Adamic. Existing raw-shape, call-returns,
type-symbol, reference-shape, type-properties, name and assignable-types questions
supply the facts, so no bridge or shared source edits were necessary. In particular,
type-properties uses GetTypeOfSymbol, matching the injection brand read exactly.
No protected compiler files, baseline source files or submodule pins changed.

GraphQL preserves first-signature return selection, one Promise unwrap, the wide
nullable mask, explicit null on inputs, ignored non-boolean option values,
relation exemptions, and Go's getter decline. Injection preserves optional-brand
undefined stripping, the original union fallback when several members remain,
any/unknown exemptions, assignability direction, parameter name-to-end spans,
call-member names and the bare-member <decorator> fallback.

## Independent agreement and mutants

The overlay oracle imports the production cohere registry and runs only the two
unchanged production implementations. It imports no bridge code. The canonical
comparison includes every file heading, UTF-8 span, rule and message ID, full
message text, fix and suggestion records and final count, preserving duplicates.
Neither rule offers fixes or suggestions, and both sides emit their zero counts.

64 distinct complete sources were extracted from the pinned Go production tests,
with their GraphQL preamble. Two more .a controls add Unicode/CRLF, a genuine union
brand, an undefined brand and a bare namespaced decorator. These 66 controls
produce 24 findings: 15 GraphQL and 9 injection. Normal and ASan/UBSan/LSan runs
agree on 7,334 bytes, with empty native stderr.

The two pinned populations are exactly the committed validation-coverage manifests:
287 repository files and TypeScript v6.0.3's 77 compiler files at
050880ce59e30b356b686bd3144efe24f875ebc8. Both have zero findings and zero checker
queries for these rules, agreeing on 18,485 and 5,010 bytes respectively, normal
and sanitized. These are measured negatives held by reporting controls, not a
claim that the corpus exercised the checker-dependent judgments.

| Mutant | Observed result | Catch |
| --- | --- | --- |
| GraphQL nullable mask omits Undefined | Exit 0, empty stderr | Production Go bytes differ at byte 784 |
| Injection no longer strips a single surviving brand member | Exit 0, empty stderr | Production Go bytes differ at byte 5090 |
| Released program kept in registry | Exit 0 instead of panic 70 | Released-handle expectation |

The released-handle probe itself panics 70 with invalid or released checker handle.
The full existing bridge gate additionally proves C input/output length checks
with ASan heap-buffer-overflow, output freeing and region ownership with LSan,
wrong-node queries with the byte oracle, and link opt-in and stale handles with
refusal expectations. It checks 100 C queries, output buffers surviving release,
zero/stale handles and distinct handles, plus 162 independently answered positions
and 3,261 identical bytes under ASan/UBSan/LSan. All these mutants were run and
caught; see bridge.log for their individual observed results.
The filtered Node oracle also kills its one-byte mutant and checks native/Node/JS
agreement, sanitizers and leaks on closures, method_closures and generic_functions.

## Native versus Go

Three alternating optimized whole-process rounds, after builds and tests finished,
run both rules and write complete diagnostic streams to files. Every round's
Go/native hashes match. Median seconds:

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.234705 | 0.131445 | 1.786 |
| Compiler | 1.403194 | 0.306441 | 4.579 |

Zero findings and queries mean these measurements primarily compare loading,
parsing, walking and output overhead, not warm checker cost. Raw nanoseconds,
phase timings, byte counts and hashes are in validation-wave-25/bench.json.
Compressed Go/native/sanitized streams, portable manifests, input hashes and
logs are adjacent. Finding stream hashes describe this run's absolute file headings.

## Commands and environment

Setup: bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh.
Go 1.27.1, clang 20.1.8 and Node 24.19.0 ready at 0s; submodules ready at 0s;
build cache warm 89s; done 89s. nproc: 5; cgroup cpu.max: 400000 100000.

```sh
ADAMIC_WAVE25_ARTIFACTS=/workspace/wave-25-validation \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-25-corpus \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware \
    -run '^TestWave25AgreementAndMutants$' > /workspace/wave-25-test.log 2>&1

go test -v -count=1 -timeout 15m ./bridge/tsgo/... \
    > /workspace/wave-25-bridge.log 2>&1

go test -v -count=1 -timeout 10m ./bridge/tsgo/checker ./internal/oracle \
    -run 'TestCoverageCheckerQuestions|TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions)\.a$' \
    > /workspace/wave-25-regression.log 2>&1

go vet ./... > /workspace/wave-25-vet-all.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware \
    > /workspace/wave-25-gofmt.log 2>&1
```

The first claim checkout failed because the initial fetch refspec fetched main
only. Fetching all origin heads made the specified base available. The specific
unit base takes precedence over the generic instruction to branch from main.
The first rule test failed before comparison with invalid tsconfig: ordinary
config discovery found no .ts inputs among the generated .a controls. Inheriting
the established test config and then using the explicit .a manifest fixed it.
A first passing run had 23 findings; review added the bare-member name control,
and the complete final rerun above has 24. Failed compilation or config checks
are not counted as comparison evidence or mutant kills.

The pinned cohere CLI command --no-cache --no-fix on the four new .a files exits
1 with nothing to check and says .a is not a TypeScript or JavaScript file.
Its format-only command exits 0 without useful coverage evidence. Neither is
claimed as a passing source lint/format check. Native stage0 builds type-check
and execute all four .a files. Updating the pinned cohere CLI's .a support is
outside this unit. The complete go test ./... gate and prior full rule suites
were not rerun; the precise targeted gates above are the validation claimed.
