# Stage 1 lint batch 4

Twenty TypeScript-specific syntax-only cohere rules, each in its own file with
one-line dispatch entries in `batch4_registry.ts`. None declares
`NeedsTypeChecker: true`; no binder, checker or inferred types are added.
The requested branch was `codex/stage1-lint-batch4`. It was already occupied
by another worker's a-to-m unit, so this implementation is published as
`codex/stage1-lint-batch4-typescript`, cut from scanner tip
`0090256e607c3f2de7d5b67cef680ec010f95c1d`, as requested.

Exclusion tips checked before selection and fetched again before finalization:

- `origin/codex/typescript-scanner`: `0090256e607c3f2de7d5b67cef680ec010f95c1d`.
- `origin/codex/stage1-lint-batch2`: `c4373c05259c7235cf1202d5cd4138c16d498caa`.
- Batch 3 was unpublished at selection; the user authorized proceeding. Final
  direct remote inspection found `fa9781c2c0911c52312002ba97ffd4b560d178ae`.
  Its `rules/registry.ts` has no overlapping rule. All twenty here come from
  cohere's TypeScript-specific directory and exclude scanner/batch2 implementations.

The unmodified Go rule source is the pinned cohere submodule
`715ba94f3608a6500086b1076ce5cb7e51b836db`. Its real `report.Write`
and `edit.FixText` are the outside oracles. An overlay captures every upstream
harness `Run`; the original upstream assertions still execute. The port
executes the same TS on Node and as native C under ASan/UBSan/LeakSanitizer.
All implementation, tests and evidence changes stay in `stage1/cohere/lint`.
No compiler, runtime, parser, scanner or submodule source is changed.

## What changed

`Batch4Context` exposes syntax indexes and source spans without owning parent
cycles. Parent indexes are filled before listeners run. Twenty independent
visitors retain Go messages, IDs, finding ranges, options and file gates.
Comment rules share the inherited reachable-comment collector and Go Unicode
whitespace predicate. U+0085 is whitespace in Go; U+FEFF is not.
`Finding` now carries additional edits; each prints its own canonical `edit`
record, preserving Go proposal order. The repair engine applies every edit,
reports Go's no-op rejection reason, and retains the original filename while
reparsing between passes. Suggestions remain unapplied.

## Own-rule corpus and mutants

There are 1,329 captured new-rule source/rule/options/file-extension combinations.
1,326 are compared; three are explicit port parser refusals below. Five of the
compared cases are findings-only recovery inputs: four unterminated block
comments and `/// <reference path=foo />`. Go's edit engine refuses malformed
input, so these do not claim converged fixed-source parity.
There are also twenty independently nonzero positive controls, a declaration-file
fix-convergence control, and three Unicode-whitespace probes. The directive
probe uses explicit findings-only recovery.

For every row below, a scratch mutant changes only that rule's enabled-name
check to an unregistered name. The control must first produce a nonzero Go
finding count. Each mutant must compile and complete on Node and sanitized
native, then differ in findings from the Go answer. A compiler error or crash
is never counted as a killed mutant.

| Rule (all `@typescript-eslint/`) | Captured cases | Mutant and check |
| --- | ---: | --- |
| `adjacent-overload-signatures` | 107 | listener omission, findings comparison |
| `ban-tslint-comment` | 42 | listener omission, findings comparison |
| `consistent-type-definitions` | 50 | listener omission, findings comparison |
| `default-param-last` | 117 | listener omission, findings comparison |
| `init-declarations` | 92 | listener omission, findings comparison |
| `no-dupe-class-members` | 30 | listener omission, findings comparison |
| `no-duplicate-enum-values` | 57 | listener omission, findings comparison |
| `no-dynamic-delete` | 42 | listener omission, findings comparison |
| `no-extra-non-null-assertion` | 19 | listener omission, findings comparison |
| `no-import-type-side-effects` | 26 | listener omission, findings comparison |
| `no-misused-new` | 45 | listener omission, findings comparison |
| `no-non-null-asserted-optional-chain` | 24 | listener omission, findings comparison |
| `no-this-alias` | 40 | listener omission, findings comparison |
| `no-unnecessary-parameter-property-assignment` | 92 | listener omission, findings comparison |
| `no-unsafe-function-type` | 25 | listener omission, findings comparison |
| `no-useless-empty-export` | 64 | listener omission, findings comparison |
| `prefer-as-const` | 69 | listener omission, findings comparison |
| `triple-slash-reference` | 70 | listener omission, findings comparison |
| `no-explicit-any` | 207 | listener omission, findings comparison |
| `no-inferrable-types` | 111 | listener omission, findings comparison |

Five additional mutations isolate repair and Unicode-boundary checks:

| Mutant | Check that catches it |
| --- | --- |
| Drop every additional edit during application | fixed source differs for an inline-type import; initial findings and all proposal records remain identical |
| Replace original filename with `fixed source` when reparsing | fixed source drops an exempt empty export from a `.d.ts` after debugger removal; initial findings remain identical |
| Change no-op rejection reason | only rejection metadata differs; findings and fixed source remain identical |
| Use JavaScript trim instead of Go TrimSpace | directive descriptions differ at U+0085 and U+FEFF |
| Use JavaScript whitespace instead of Go Fields | a reference separated by U+0085 loses its finding |

The following 31 inherited mutations also compiled and executed on both Node
and sanitized native, then failed the named comparison. The position-anchor
mutation initially stopped because its anchor occurred twice; narrowing the
anchor to the warning collector made its intended executable check pass.

| Inherited mutant | Check that catches it |
| --- | --- |
| Suggestion applied as fix | repair category and fixed source |
| Empty function body reported | extra finding |
| Duplicate case suppressed | finding list |
| Continue listener changed to break | finding list |
| Destructuring hole exemption removed | extra finding |
| Nested generator yield crosses boundary | missing yield finding |
| Await crosses function boundary | extra await finding |
| int32Hint option inverted | operator findings |
| Regex fix eats extra byte | edit span and fixed source |
| Loose comparison inverse changed to strict | replacement and fixed source |
| Comment self-directive exemption removed | extra finding |
| BOM edit removes two marks | edit span |
| Label loop option inverted | label/jump findings |
| Directive prefix ignored | extra vars-on-top finding |
| Comma chain reports inner node | finding list |
| Empty template placeholder reported | extra finding |
| Overlap winner misreported | rejected-edit metadata |
| Type-predicate listener changed | finding list |
| Console receiver widened | finding list |
| For-afterthought option inverted | extra update finding |
| Method-signature listener lost | missing finding and fix |
| Wrapper Number ignored | missing finding |
| Enum bitwise option inverted | enum findings |
| Enum declaration listener lost | missing finding |
| Negated-condition polarity reversed | condition findings |
| Return parentheses option inverted | assignment findings |
| Interface type suffix accepted | missing suffix finding |
| Decoration range reduced to first character | missing option-dependent comment finding |
| Count mode incremented | count-only check; ordinary output remains identical |
| Unicode fold offset changed | comment finding list |
| Opening-comment anchor removed | missing opening comment finding |

The raw logs retain actual differing lines for all 56 lint mutations (25 new,
31 inherited), plus the filtered internal oracle's one-byte mutation.

## Explicit limits

These exact upstream cases remain in the inventory and are tested as refusals
on both Node and sanitized native, bounded at two seconds:

- `no-explicit-any`: `interface Greeter { constructor(param: Array<any>) {} }`.
- `no-explicit-any`: `type obj = { constructor(param: Array<any>) {} }`.
- `no-unnecessary-parameter-property-assignment`:
  `class Foo { constructor(public { a }: { a: string }) { this.a = a; } }`
  (the real case retains its original newlines).

The first two are malformed signatures that Go recovers. The third is a valid
parser dependency that stage1 refuses at the destructured parameter property.
The upstream Go answer is logged; ports exit 70 with `adamic: panic:` and no
successful parity is claimed for these three. Eighteen rules therefore have
complete own-fixture coverage, and two have the explicit limits above.
Inherited method-signature recovery limits remain unchanged.

General config decoding/validation, suppression, filesystem-writing integration,
JSX, JSDoc AST analysis, general binding, non-UTF-8 input and parser-diagnostic
parity are outside coverage. Options are cohere's already-decoded Go JSON.
The full repository gate was not run; the touched package's tests, its vet and
the named filtered internal oracle were run. Profile tests requiring saved
snapshot environment variables remain skipped.

## Toolchain and setup

`source /workspace/adamic-tools/env.sh`: Go 1.27.1, clang 20.1.8,
Node 24.19.0. `nproc` is 5; cgroup quota is four cores.

The first `bash cloud/setup.sh` failed in cache warmup because the branch was
switched while it was still enumerating/building packages. It reported missing
`internal/unicodeproperties`, `stage1/cohere/cssnumbers`, `cssstrings`, `json`
and `selector` files, and missing `volumeGenerated`/`checkRecoveryRefusal`.
Retrying on the stable selected branch succeeded:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (17s)
setup: done in 17s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

## Commands and evidence

All test stdout/stderr goes to a log file, never a pipeline. Commands run from
the repository after sourcing the toolchain environment. The compiler corpus
is scratch-only TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`; the test asserts the pin.

The final canonical comparisons are:

| Check | Observed output |
| --- | --- |
| Twenty nonzero controls | 13,277 identical bytes |
| 1,326 supported own-rule cases | 460,313 identical bytes; three separate refusals |
| 188 source files (77 compiler, 111 stage1) | 19,192,994 identical bytes |
| Declaration-file convergence | 509 identical bytes; filename mutant caught |
| Unicode boundary probes | 1,656 identical bytes; both whitespace mutants caught |
| Combined baseline/new upstream replay plus 145 generated runs | 1,210,274 identical bytes; eleven separately checked refusals |
| Twenty family mutants | PASS, 357.345s |
| No-op rejection mutant | PASS, 19.969s |
| Corrected inherited position mutant | PASS, 19.930s |
| Filtered internal one-byte oracle | PASS, 2.920s |
| Cohere source gate | exit 0, 276 rules, 27 checked, 100% Adamic-ready |
| Touched-package vet, gofmt, source diff check | exit 0, no output |

The regression command initially exited 1 solely because
`TestPositionIndexMutant` could not identify its duplicated anchor. Every other
executed check in that run passed. Narrowing the test's anchor and rerunning
that test passed on both ports. This is a harness fix, not a caught semantic
mutant; the actual successfully executed mutation is recorded above. The raw
failed regression log is retained. A corpus test in the earlier combined
upstream/filename run skipped because its environment variable was absent;
the explicit pinned-corpus rerun and final supported-parity run both passed.
No skipped run is counted as parity evidence.

```sh
source /workspace/adamic-tools/env.sh

go test ./stage1/cohere/lint -run '^TestBatch4Mutants$' \
  -count=1 -v -timeout 30m > /tmp/batch4-mutants.log 2>&1

go test ./stage1/cohere/lint \
  -run '^(TestRulesAgree|TestBatch4Controls|TestBatch4ExtraEditMutant|TestNestedConstructorGap|TestOptionAndComparatorGaps|TestMutants|TestDecorationOptionMutant|TestCountGuardMutant|TestCommentFoldMutant|TestPositionIndexMutant|TestVolumeMutants|TestProfileArtifacts|TestProfileSnapshotsAgree)$' \
  -count=1 -v -timeout 30m > /tmp/batch4-regression.log 2>&1

go test ./stage1/cohere/lint -run '^TestPositionIndexMutant$' \
  -count=1 -v -timeout 10m > /tmp/batch4-position-mutant.log 2>&1

go test ./stage1/cohere/lint \
  -run '^(TestBatch4FilenameMutant|TestBatch4UpstreamAgree|TestCompilerAndStage1Agree)$' \
  -count=1 -v -timeout 20m > /tmp/batch4-final-parity.log 2>&1

go test ./stage1/cohere/lint -run '^TestBatch4NoopReasonMutant$' \
  -count=1 -v -timeout 10m > /tmp/batch4-noop-mutant.log 2>&1

go test ./stage1/cohere/lint -run '^TestBatch4UnicodeWhitespace$' \
  -count=1 -v -timeout 15m > /tmp/batch4-unicode.log 2>&1

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/batch4-typescript \
  go test ./stage1/cohere/lint \
  -run '^(TestBatch4UpstreamAgree|TestCompilerAndStage1Agree)$' \
  -count=1 -v -timeout 20m > /tmp/batch4-final-supported-parity.log 2>&1

go vet ./stage1/cohere/lint > /tmp/batch4-vet.log 2>&1
gofmt -l stage1/cohere/lint > /tmp/batch4-gofmt.log

go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' \
  -count=1 -v -timeout 10m > /tmp/batch4-filtered-oracle.log 2>&1
```

The source gate runs `/workspace/scratch/batch4-cohere --no-fix --no-cache`
with the twenty paths listed in `testdata/batch4_rules.json`, plus
`batch4_context.ts`, `batch4_registry.ts`, `extra_edit.ts`, `finding.ts`,
`main.ts` and `lint.ts`. Its actual output is retained in
`batch4_evidence/source-gate.log`.

## Measurement

```sh
ADAMIC_LINT_BENCH=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/batch4-typescript \
  go test ./stage1/cohere/lint -run '^TestThroughput$' \
  -count=1 -v -timeout 20m > /tmp/batch4-final-throughput.log 2>&1
```

Full release output is compared before timing. Five rounds interleave
Go/native/Node count mode on all fifty rules, including the scanner's thirty.
The sample includes process startup, reads, scanning, parsing and visitors;
it excludes output formatting, fixing and build time. Native uses unsanitized
clang -O2; correctness and mutant runs use sanitizers with leak checking.
This is a whole-file workload, not an isolated twenty-rule visitor benchmark.
Shared-machine timing is an observation, not a performance guarantee.

Every sample reported **15,119 findings** across 77 compiler files. Full
release output first matched Go: **17,993,804 bytes**.

| Implementation | Best of five seconds | Findings per second |
| --- | ---: | ---: |
| Go | 0.945173 | 15,996.01 |
| Native Adamic | 4.546550 | 3,325.38 |
| Node | 2.118871 | 7,135.40 |

Native takes 4.81 times Go's time and 2.15 times Node's on this workload.
No speedup claim is made. Machine: Linux x86_64, kernel 6.18.44,
AMD EPYC 9V74 80-Core Processor; `nproc` 5, four-core quota. Load before:
`0.75 1.04 1.21`; after: `0.98 1.06 1.21`.
`TestThroughput` passed in 68.674s. Every sample and the machine line are in
[the raw throughput log](batch4_evidence/throughput.log).

## Commits and raw artifacts

Implementation and tests:
`526954ab4cff2101fb7195433bdf60e9cc9a9469`,
`Port twenty TypeScript syntax lint rules and preserve Go edit output`.
A separate evidence commit adds this report and
[the raw logs](batch4_evidence/). The requested remote `codex/stage1-lint-batch4` already contains another
worker's a-to-m batch at `d486b03a15f3202acc317dc81c40cc21818e21f7`, based
on batch2. The first push was rejected without changing remote history.
Publication uses `codex/stage1-lint-batch4-typescript` to preserve that worker's
commits. No merge with the occupied branch is performed; its a-to-m unit also
contains three of these TypeScript rules, so a future integration must deduplicate
ban-tslint-comment, default-param-last and init-declarations. These three were
absent from the scanner and batch2 tips specified for this selection.
No pull request is opened.

Raw oracle whitespace is preserved through the local `.gitattributes` entry.
Source/document and staged whitespace checks pass. The retained failed setup
and regression logs are labeled as initial attempts and are not claimed as
successful checks.
