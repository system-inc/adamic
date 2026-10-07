Built: native Adamic ports of no-import-assign, prefer-exponentiation-operator and use-isnan.
Commits: claim 17d1c8d1 was pushed before code; implementation a18c8db9.
Commands and output: wave gate PASS 66.672s; bridge PASS 67.328s; filtered Node oracle PASS 19.288s; vet and gofmt logs empty.
Mutants: three rule mutants caught only by Go finding bytes; released-registry mutant caught by panic expectation; bridge and Node mutants detailed below.
Not covered: full repository gate, nondefault options, every upstream fixture, JSX/JavaScript projects, or the pinned CLI's unsupported .a lint selection.

## Selection and implementation

Branch: codex/typeaware-wave-16. Base:
0d540f413625f016f20fea39761c7b184f335de6 on origin/codex/tsgo-c-library.
The unit-specific base overrides the generic main-branch instruction.
The initial checkout tracked only main; all origin heads were explicitly fetched.
No matching existing port or claim was found on those heads before the claim push.
No rules were skipped or replaced.

The ranking is the combined compiler and repository counts in
validation-volume/compiler-all.counts and repository-all.counts, linked from
VOLUME_REPORT.md. Remove the sixteen oracle_volume subjects and the ten
validation-coverage/selection.json subjects before assigning positions; sort
counts descending, with lexical ties. These are the requested positions:

| Position | Rule | Compiler | Repository |
| --- | --- | ---: | ---: |
| 46 | no-import-assign | 1 | 0 |
| 47 | prefer-exponentiation-operator | 0 | 1 |
| 48 | use-isnan | 0 | 1 |

Each port has its own .a file. wave16_suite.a loads one checker program and uses
the existing native parser, numeric parent index and canonical diagnostics.
wave16_facts.a decodes existing resolved-name facts; Bindings.list is used with
alias following disabled to anchor imported bindings. Declaration identity,
namespace write targets, shadowing decisions, NaN references, report ranges,
exponentiation precedence, comment refusals and repair text are judged in Adamic.
No new checker question was necessary. No shared bridge, compiler, runner or
registration file was edited; submodule pins are unchanged.

The independent Go oracle is testdata/oracle_wave16.go, built through an overlay
inside pinned cohere. It imports the production registry, walks its own Go AST,
and invokes the three unmodified production Run methods with default options.
It imports no bridge code. Complete sorted diagnostics, fixes and suggestions
are compared, retaining duplicates and fix order. No edits are applied.

## Inputs and agreement

Compiler: the same 77 TypeScript src/compiler roots, v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8.
Repository: the existing frozen 287-root manifest, 212 .a and 75 .ts.
The manifests remain in validation-coverage; new wave sources are not added to
that selection population. Configured declaration roots are retained by both
loaders. Cohere remains 715ba94f3608a6500086b1076ce5cb7e51b836db;
typescript-go remains 8d550c837c90bd1805b047b7eeccc2baac2d5e7a.

| Population | Findings | Identical bytes, normal and ASan/UBSan/LSan |
| --- | ---: | ---: |
| Compiler | 1 | 6,379 |
| Repository | 2 | 19,351 |
| Controls, 16 sources plus helper | 65 | 29,690 |

The one compiler finding is no-import-assign. The repository contributes one
prefer-exponentiation-operator finding with its complete fix and one use-isnan
finding. No production-default suggestion is emitted by these three rules.
Positive controls are required for every rule; a zero without a positive control
cannot pass this test. Controls include imported-name shadowing, namespace versus
ordinary imported objects, distinct import anchors, shorthand and nested target
writes, parentheses, Object/Reflect mutation calls, shadowed globals, Number.NaN,
comma values, dead switch branches, default indexOf exemption, computed pow names,
wrong arity, spreads, comments versus string/regexp contents, operator precedence,
lexical spacing, alias exemptions, Unicode names and CRLF spans.

Compressed Go, native and sanitizer streams, hashes and input source hashes are
in validation-wave16. Stream headers include absolute paths, so relocated runs
must compare their own Go/native bytes rather than expecting identical hashes
with this machine. Native stderr was empty on all normal and sanitizer comparisons.
Go's checker heap is not instrumented by the C sanitizers.

## Mutants and ownership

All three rule mutants compile, exit 0 and produce no stderr. Only the independent
Go diagnostic byte comparison catches them:

| Rule | Mutation | First differing byte | Mutant findings |
| --- | --- | ---: | ---: |
| no-import-assign | Set the direct binding-write judgment false | 63 | 59 |
| prefer-exponentiation-operator | Omit base parentheses | 15,651 | 65 |
| use-isnan | Read a PlusToken rather than CommaToken sequence value | 12,061 | 63 |

The equal-count exponentiation mutant changes only repair bytes. The test never
applies that malformed repair, so neither a syntax check of the repaired input nor
clang can kill it. Mutant sources and generated controls use .a; existing .ts
imports point to the unchanged originals rather than new scratch .ts copies.

A released program queried through node-symbol-details panics with exit 70 and
exact stderr `adamic: panic: invalid or released checker handle`. The registry
mutant omits deletion of the released handle, exits 0 and fails that expectation.

The separate bridge regression gate passes its C ABI probe: 100 queries, output
strings surviving program release, zero/stale handles rejected, a distinct second
handle and Unicode output. Its independent sample checks 162 positions and 3,261
bytes under ASan, UBSan and LeakSanitizer. Its mutants are all caught:

| Bridge mutant | Check that catches it |
| --- | --- |
| Input length +1 | ASan heap-buffer-overflow |
| Output string length +1 | ASan heap-buffer-overflow |
| Released handle kept live | Stale-handle assertion |
| Type from source-file position | Independent Go byte oracle, byte 6 |
| Removed link opt-in guard | Refusal expectation |
| Omitted C output frees | LeakSanitizer |
| Region result allocated on heap | LeakSanitizer |

The filtered Node oracle passes functions, generic_functions, method_closures,
string_index, closures, sorting, lone_surrogates and maps_and_text, with native
sanitizers/leak checks and the JavaScript backend. TestTheOracleCatchesOneByte
also passes by detecting its injected output-byte mutation.

## Toolchain and recovery

bash cloud/setup.sh passed. Timing lines: Go ready 0s; clang ready 1s; Node ready
1s; submodules ready 1s; build cache warm 85s; done 85s. nproc: 5; cgroup cpu.max:
400000 100000. Environment: Go 1.27.1, clang 20.1.8, Node v24.19.0, Linux amd64.
Each Go command sourced /workspace/adamic-tools/env.sh, the path setup printed.

Initial native builds refused inferred never[], negation of boolean | undefined,
multiple-value push and string charAt. The new sources were adjusted to use a
typed array local, an explicit missing-map fallback, individual pushes and slices.
No compiler changes were made. The refusal logs are preserved; refused builds
are not reported as mutant kills or successful verification.

The pinned cohere CLI does not recognize .a paths for project lint selection.
The attempted dry lint exits 1 with `nothing to check`; its log is preserved and
no lint pass is claimed. Formatting used its stdin interface with virtual .ts
path metadata, without creating or renaming any source file. All five submitted
Adamic files remain .a and the formatted native source was revalidated in full.
Native compilation typechecks the new sources. The full go test ./... gate was
not run; the touched package's new gate, full bridge packages and filtered Node
oracle below are the exact checks performed.

## Commands

All test stdout/stderr went directly to log files. Tests retain child command
logs under ADAMIC_WAVE16_ARTIFACTS. The final run used:

```sh
source /workspace/adamic-tools/env.sh
TMPDIR=/workspace/wave16-artifacts \
ADAMIC_WAVE16_ARTIFACTS=/workspace/wave16-artifacts/formatted-final \
ADAMIC_WAVE16_REPOSITORY_MANIFEST=/workspace/wave16-artifacts/repository.manifest \
ADAMIC_WAVE16_COMPILER_MANIFEST=/workspace/wave16-artifacts/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave16-corpus/typescript \
go test ./stage1/cohere/typeaware -run '^TestWave16AgreementAndMutants$' \
    -v -count=1 -timeout=30m > /workspace/wave16-artifacts/formatted-full-test.log 2>&1

TMPDIR=/workspace/wave16-artifacts \
go test ./bridge/tsgo/... -v -count=1 -timeout=15m \
    > /workspace/wave16-artifacts/bridge-test.log 2>&1

TMPDIR=/workspace/wave16-artifacts go test ./internal/oracle \
    -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
    -v -count=1 -timeout=10m > /workspace/wave16-artifacts/node-oracle.log 2>&1

go vet ./stage1/cohere/typeaware > /workspace/wave16-artifacts/final-vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave16_test.go \
    stage1/cohere/typeaware/testdata/oracle_wave16.go > /workspace/wave16-artifacts/gofmt.log
```

Absolute manifests were materialized from the committed relative ones exactly as
validation-coverage/README.md describes, using the roots above. The corpus was
obtained with `git clone --depth 1 --branch v6.0.3
https://github.com/microsoft/TypeScript.git /workspace/wave16-corpus/typescript`;
its observed HEAD is the required compiler pin. Tests themselves never download.

## Native time against Go

Three interleaved rounds, alternating order, after builds and verification had
finished. Every run writes complete finding streams to files, including fixes;
the benchmark refuses different stream hashes. No competing build/test ran.
These are whole-process observations, including load, parse, lint, rendering,
sorting, output and teardown, rather than bare C crossing costs.

| Corpus / implementation | Median load | Median run | Median whole process |
| --- | ---: | ---: | ---: |
| Compiler native | 0.247s | 2.295s | 2.563s |
| Compiler Go | 0.234s | 0.082s | 0.336s |
| Repository native | 0.068s | 0.203s | 0.278s |
| Repository Go | 0.058s | 0.051s | 0.123s |

Native is 7.62x Go's whole-process time on the compiler and 2.26x on the
repository. Native makes 8,792 compiler queries and 337 repository queries.
Median query aggregate: 0.234s and 0.0164s respectively. Most compiler native
run time is outside the bridge query intervals. This unit ports coverage and
makes no performance-parity claim. Separate phase medians need not add to the
whole-process median.

Exact rounds, times and diagnostic hashes are in validation-wave16/measurements.json;
raw timing lines, benchmark.py and benchmark.log are adjacent. The timed command
was `python3 /workspace/wave16-artifacts/benchmark.py >
/workspace/wave16-artifacts/benchmark.log 2>&1`. Its inputs are the final native
runner, independent Go oracle, two absolute manifests and configs above.
