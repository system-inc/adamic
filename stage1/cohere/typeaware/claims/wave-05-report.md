Built: three native rules and the raw usage-shape checker question for wave 05.
Commits: claim 739be38b; ports 52b0de63; preserved refusal registration 28eb1b25.
Commands and outputs: wave go test PASS 175.516s; bridge, checker and filtered Node oracle PASS; touched-package vet clean.
Mutants: all three rule mutants exit normally and fail independent byte comparison; released-handle and checker-kind mutants are also caught.
Not covered: nondefault rule options, the complete upstream fixture matrix, and the full repository gate.

# Wave 05

The branch starts from `origin/codex/tsgo-c-library`, pinned at `0d540f413625f016f20fea39761c7b184f335de6`, as the unit-specific base instruction requires. The claim-only commit was pushed before implementation. Existing port filenames and claims on all fetched origin branches were checked before claiming; none of the selected rules was already ported or claimed, and none was skipped.

The by-volume ranking was reconstructed from the report's compiler and repository counts, descending combined volume with lexical tie-breaking, excluding the 26 rules already ported on the base branch. Remaining positions 13, 14 and 15 are:

| Rule | Compiler | Repository | Total |
| --- | ---: | ---: | ---: |
| no-unnecessary-type-parameters | 30 | 1 | 31 |
| no-useless-assignment | 17 | 12 | 29 |
| restrict-template-expressions | 23 | 0 | 23 |

Each rule has its own `.a` file. `no_unnecessary_type_parameters.a` counts references and traverses raw checker topology in native Adamic, including constraint/default multiplicity and witness exemptions. `no_useless_assignment.a` constructs a native control-flow graph and solves backward liveness using checker declaration identities. `restrict_template_expressions.a` evaluates constrained union/intersection types and the production default primitive and library allowlists. The independent production Go oracle computes none of the native rule verdicts.

The new checker question is confined to `bridge/tsgo/checker/usage_shape.go`; its Adamic decoder is `usage_shape.a`. The wire format is documented in [usage_shape.md](../../../../bridge/tsgo/usage_shape.md). The only shared source edit is one added physical switch-registration line in `checker/facts.go`. It deliberately stays on one line to honor the wave's shared-file restriction; gofmt would expand this registration. The existing unknown-question refusal and its mutant remain intact. No protected compiler source was edited.

## Setup and pinned inputs

`bash cloud/setup.sh` passed. Its timing lines are preserved in [setup.log](wave-05-evidence/setup.log): Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 82s, total 82s. `source /workspace/adamic-tools/env.sh` selects Go 1.27.1, clang 20.1.8 and Node 24.19.0. `nproc` printed **5**; cgroup `cpu.max` is `400000 100000`.

| Input | Revision |
| --- | --- |
| Go cohere | `715ba94f3608a6500086b1076ce5cb7e51b836db` |
| typescript-go | `8d550c837c90bd1805b047b7eeccc2baac2d5e7a` |
| TypeScript 6.0.3 | `050880ce59e30b356b686bd3144efe24f875ebc8` |

The frozen [compiler manifest](../validation-coverage/compiler.manifest) contains 77 files under TypeScript's `src/compiler`; the frozen [repository manifest](../validation-coverage/repository.manifest) contains 287 files. Absolute manifests were generated outside the repository from these portable lists. Newly added wave files were not added to the frozen comparison population.

## Validation

The production Go oracle has its own compiler host and AST walk, invokes the three unmodified registered rules with default options, and imports no bridge code. Both implementations write canonical records containing byte ranges, rule/message IDs, escaped message text, every fix, and every suggestion and its fixes. Equality is of the complete byte stream. The 24 additional controls cover allowed and rejected template types, tagged templates, constraints, library inheritance, Unicode/CRLF, RegExp aliases, generic witnesses, arrays, branches, loops, closures, updates, destructuring, conditional expressions, and try/finally.

| Population | Type parameters | Dead assignments | Templates | Total findings | Suggestions | Fixes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Controls, 24 files | 6 | 6 | 8 | 20 | 6 | 0 |
| Repository, 287 files | 1 | 12 | 0 | 13 | 1 | 0 |
| Compiler, 77 files | 30 | 17 | 23 | 70 | 30 | 0 |

Each population is compared normally and under ASan, UBSan and LeakSanitizer with leak detection enabled. Production's type-parameter suggestions contain no repair edits; that absence is part of the comparison. Canonical oracle streams, the three mutant streams, byte hashes, and source hashes are retained under `wave-05-evidence`.

The principal command is:

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE05_ARTIFACTS=/workspace/wave-05-release-validation \
ADAMIC_WAVE05_REPOSITORY_MANIFEST=/workspace/wave-05-repository.manifest \
ADAMIC_WAVE05_COMPILER_MANIFEST=/workspace/wave-05-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-05-typescript \
ADAMIC_WAVE05_BENCH=1 \
go test ./stage1/cohere/typeaware -run '^TestWave05AgreementAndMutants$' \
  -count=1 -timeout=30m -v > /workspace/wave-05-release-test.log 2>&1
```

Additional checks, with test output sent directly to logs:

```bash
go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-05-bridge-test.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave-05-checker-final.log 2>&1
go test ./stage1/cohere/typeaware \
  -run '^Test(PinnedTypeFlags|SixPinnedFlags|FactsDecoderGuards|InspectRequestRefusals)$' \
  -count=1 -timeout=10m -v > /workspace/wave-05-regression-final.log 2>&1
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout=10m -v > /workspace/wave-05-node-test.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /workspace/wave-05-vet-final.log 2>&1
```

The filtered Node oracle command also selects names containing the listed fragments, including generic functions and method closures. It passed in 13.136s, including its one-byte comparison mutant. The full bridge package suite passed in 100.550s. Its C ABI coverage includes 100 queries, strings surviving handle release, zero/stale handles, distinct handle identities, Unicode frames, and 162 oracle positions under all three sanitizers.

## Mutants

| Mutant | Observation | Check that catches it |
| --- | --- | --- |
| Templates accept `unknown` as string-like | Builds, exits 0, empty stderr, different findings | Independent Go byte comparison |
| Type-parameter threshold changes `count > 2` to `count > 1` | Builds, exits 0, empty stderr, loses single-use diagnostics/suggestions | Independent Go byte comparison |
| Dead-store judgment reverses `dead` | Builds, exits 0, empty stderr, different findings | Independent Go byte comparison |
| Registry retains a released program | Builds and exits 0 | Required released-handle exit 70 and exact panic text |
| usage-shape node-kind guard disabled | Go test fails with `usage-shape accepted a value node: <nil>` | Direct checker refusal assertion |

The complete bridge suite also runs its existing input/output-length ASan mutants, output-free and region-allocation LeakSanitizer mutants, released-registry mutant, wrong-position byte comparison, and unlinked-call guard mutant. All are caught; details are in [bridge.log](wave-05-evidence/bridge.log). Existing facts decoder malformed-wire checks and wrong-kind/unknown-question refusal mutants pass after the routing registration was corrected.

## Formatting and limits

Pinned cohere's CLI rejects `.a` input during lint selection and its formatter does not register `.a`. This is observed in [lint-refusal.log](wave-05-evidence/lint-refusal.log) and [format-skip.log](wave-05-evidence/format-skip.log). A scratch Go build overlay registered `.a` with cohere's TypeScript formatter; it changed all five new Adamic source files, and its final `--format-only --no-fix --no-cache` run reports zero files needing changes. The submodule was not edited. New Go files are gofmt-clean and vet passes. Native builds type-check the Adamic sources.

The full `go test ./...` gate and the old 26-rule complete corpus suite were not rerun. Validation covers the touched bridge packages, this wave's full specified corpora, selected shared facts/refusal regressions, and the filtered external Node oracle. No claim is made for nondefault option configurations or the entire upstream rule-test matrix. The native dead-assignment implementation is conservative for captured reads and writes in protected exception regions; optional-chain and generator behavior has not been exhaustively checked outside the frozen corpora.

## Final observations and timing

The final test passed in **175.516s**. Controls matched **6,597 bytes**, repository matched **24,608 bytes**, and compiler matched **23,501 bytes**, in both normal and sanitized runs. The comparison-only mutants diverged at bytes **235**, **1,379** and **3,255** respectively. The final shared facts/refusal regression run passed in **34.373s**; checker tests passed in **0.087s**. Touched-package vet emitted no findings.

Three alternating count-only Go/native rounds were run after the other test/build jobs finished. These are whole-process wall times, including loading, parsing, checking, rule execution and startup. They measure the complete three-rule driver, not individual rule CPU times. Raw rounds and load/query/rule timings are in [wave.log](wave-05-evidence/wave.log) and [timings.json](wave-05-evidence/timings.json).

| Population | Go median | Native median | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.183527s | 0.531729s | 2.90x |
| Compiler | 0.587785s | 14.029293s | 23.87x |

Native is slower on both populations. It issued 20,277 checker queries on the repository and 135,676 on the compiler. This wave establishes agreement on the specified populations; it does not establish a performance win or attribute the remaining cost to a particular native phase.

The earlier `nonunique mutant` regression failure is retained as [earlier-regression-failure.log](wave-05-evidence/earlier-regression-failure.log). It was resolved by restoring the old unsupported-question refusal and adding only the dedicated registration line; the successful final run is [decoder-regression.log](wave-05-evidence/decoder-regression.log).
