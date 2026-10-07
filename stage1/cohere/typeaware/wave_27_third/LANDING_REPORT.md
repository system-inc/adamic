# Wave 27 landing revalidation

Rebased codex/typeaware-wave-27 onto origin/main
`e8ba3d5d81de4d3773c723914fccd4c76248b965` and rebuilt all nine completed
ports. The tested code tip is `381c6172a7fb641c10a18254aff4c2b500fd99b2`.
The final evidence commit changes only this report and owned evidence.
The range diff preserves every original patch unchanged. The original remote
tip was `460188555e99bcd92ff1fcd3d33fff0c499f1b74`; the publish command uses an
explicit lease against that tip. This is the only branch pushed in this unit.
No new rules were claimed and no shared source conflict repairs were needed.

The first rebase onto e011f8f6 also passed. Before publishing, a remote check
found main had advanced with native devirtualization and call-target changes.
A second rebase and complete fresh validation were therefore performed.
Logs without v2 refer only to the earlier base. Logs with v2 are the final-base
results below. The final remote check confirmed main was still e8ba3d5d.

## Commands and results

Every toolchain shell sourced `/workspace/adamic-tools/env.sh` and set
`TMPDIR=/tmp/adamic-gate`. Toolchain setup earlier in this unit reported
Go/clang/Node/submodules ready 0s, cache warm 116s and total 116s; nproc is 5.
No toolchain reinstall was needed for this rebase.

The final gate used these environment variables:

```sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-27-scratch/typescript
export ADAMIC_WAVE27_REPOSITORY_MANIFEST=/workspace/wave-27-scratch/repository.manifest
export ADAMIC_WAVE27_COMPILER_MANIFEST=/workspace/wave-27-scratch/compiler.manifest
export ADAMIC_WAVE27_ARTIFACTS=/workspace/wave-27-scratch/landing-v2-first
export ADAMIC_WAVE27_NEXT_ARTIFACTS=/workspace/wave-27-scratch/landing-v2-next
export ADAMIC_WAVE27_THIRD_ARTIFACTS=/workspace/wave-27-scratch/landing-v2-third

go test -v -count=1 -timeout 15m ./stage1/cohere/typeaware -run '^TestWave27AgreementAndMutants$' > /workspace/wave-27-landing-v2-first.log 2>&1
python3 stage1/cohere/typeaware/wave_27_next/validate.py > /workspace/wave-27-landing-v2-next.log 2>&1
python3 stage1/cohere/typeaware/wave_27_third/validate.py > /workspace/wave-27-landing-v2-third.log 2>&1

go test -v -count=1 -timeout 15m ./bridge/tsgo/checker ./stage1/cohere/typeaware -run '^(TestCoverageAgreementAndMutants|TestCoverageCheckerQuestions|TestFactEncoding|TestShapeAndNameFacts|TestExactIndexMatchesCompilerNodes|TestAdamicRootKeepsConfigDeclarations|TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$' > /workspace/wave-27-landing-v2-regression.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /workspace/wave-27-landing-v2-vet.log 2>&1
```

All commands exited zero. The first gate passed in 126.995s; the independent
next and third validators each printed PASS. The regression packages passed
in 0.217s and 248.749s. Vet output is empty. Test output was written directly
to files, without test pipelines. Builds use the newly rebuilt native compiler,
normal and Go-ASan C archives, and native sanitizer compilation.

| Completed batch | Control findings | Identical output bytes |
| --- | ---: | ---: |
| Decorator parity | 39 | 17,807 |
| Stream and race correctness | 101 | 62,759 |
| Regex, rest parameters, Hook dependencies | 437 | 229,337 |

Each completed batch matches the independent unchanged production Go rules
on all findings, fixes and suggestions, in normal and sanitized native runs.
Each also matches over the frozen 287-file repository corpus (18,485 bytes,
zero findings) and 77-file compiler corpus (5,934 bytes, zero findings).
Source hashes and portable corpus manifests are retained. The third batch
uses 605 parseable controls from 639 candidates; 34 rejected by the Go parser
remain explicit exclusions and are not passes. Source strings and exclusions
are retained, including intentionally invalid candidates as JSON evidence.
The inherited ten bridge ports also passed their control byte comparison:
53 findings, normal and sanitized, with all ten semantic mutants.

## Mutants and released handles

Each of the nine owned rule mutants compiled, exited zero with empty stderr,
and was caught only by the independent Go byte comparison:

- ORM: reverse the nullable discriminator.
- Serializable: reverse explicit optional-true versus nullable matching.
- Verify array: invert the known suppressor decision.
- Process exit: report the empty output state instead of pending output.
- Race timeout: invert the lost timer ownership test.
- Blocking streams: return before checking stream setup.
- Rest parameters: reverse the implicit arguments declaration test.
- Hook dependencies: reverse selection of a missing-dependency warning.
- Regex literals: reject the accepted constructor argument counts.

Five normally exiting bridge fact mutants were also caught by Go bytes:
declaration ancestry loses declaration-file status; syntax-flow marks every
block unreachable; call-declaration forces never return flags; module-sources
makes every import type-only; ancestry-name-kind replaces destructuring kind
with Identifier. Exact first difference offsets are in the retained logs.

Inherited before/cast/methods/coercion/caller/parameter/invariant/optional/alias/
unused rule mutants all exited zero and failed their Go byte comparison.
These respectively change initialization order, cast flags, method traversal,
coercion truth, caller flags, parameter anchoring, assignability direction,
optional field naming, edit range and arrow-function usage tracking.

Released handles were refused with exact panic 70 for raw-type, the four new
stream questions, binding-declarations, node-symbol-details and type-name.
Each gate's registry-retention mutant exited zero and was caught by the
required-panic check. Wrong-kind and unknown-question guard mutants also exited
zero and failed the required refusal. All nineteen malformed wire cases were
refused with panic 70: empty, negative/short length, trailing data, wrong
version/question, invalid/rounded/noncanonical integer, negative natural,
bad boolean, zero ID, negative tuple count, missing root/constraint/element,
duplicate ID and missing links 17/18. Their exact messages are retained in the
regression log. Sanitizer controls and both owned corpora completed without
native sanitizer or leak diagnostics.

## Native versus Go timing

Three alternating process samples per engine retain and compare full output.
These timings include process startup, loading, native parsing, judgment,
rendering and teardown; they are not timings of just bridge calls.

| Batch and corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Stream/race repository | 443.971 ms | 155.124 ms | 2.862 |
| Stream/race compiler | 3,287.628 ms | 410.474 ms | 8.009 |
| Regex/rest/Hook repository | 414.270 ms | 148.642 ms | 2.787 |
| Regex/rest/Hook compiler | 3,241.534 ms | 424.010 ms | 7.645 |

Decorator parity's single process observations were repository native
273.236ms versus Go 213.634ms and compiler native 2.104s versus Go 458.858ms.
These early samples overlapped the regression gate and are not quiet medians.
The later stream/race and third-batch timing rounds ran after that gate ended.
Native is slower than Go on every observed corpus comparison.

## Pending claims and coverage limits

The effect-state, render-state and static-component claims remain pending.
The JSX parser probe was rebuilt on the final compiler and still exits 70
with `expected GreaterThanToken, got SlashToken at 84`; the independent Go
oracle exits zero with one staticComponents finding at bytes 74:83. Its exact
commands and streams are retained in wave-27-landing-v2-parser.json.
JSX support exists on origin/codex/stage1-jsx-lint at e715ef4a, with evidence
at a8a62d62, but is absent from main and this branch. Importing it modifies
shared parser and scanner files outside this unit's allowed scope. Ahra's
stop-on-other-blockers instruction therefore still applies. No always-empty
replacement was added, and no further claims were made.

No pending-rule byte gate, mutant, sanitizer or timing result is claimed.
Full repository tests, emitted-JavaScript comparison, all CLI options and the
other sixteen inherited rule gates were not rerun here. This report covers
the named nine owned ports, ten inherited bridge ports and focused bridge
regressions. Earlier reports remain historical evidence for their earlier
code tips. No protected compiler or shared harness source was edited by this
unit during landing; compiler changes in main were received by the rebase.
