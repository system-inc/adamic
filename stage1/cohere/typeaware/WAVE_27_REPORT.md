# Type-aware wave 27

Built three separate native Adamic decorator parity rules using existing raw-type bridge facts.
Claim commit: `d0f7d7bd4c4fdc1f32a16bd6876ba03c8b897142`; implementation is the commit containing this report.
Commands: setup 116s; final wave PASS 87.467s; checker PASS 0.438s; bridge PASS 100.455s; Node oracle PASS 87.821s; vet clean.
Mutants: all three rule decision inversions must exit normally and differ from Go; released-registry retention must violate panic 70.
Not covered: full repository gate, all upstream rule fixtures, and a clean stock-cohere source lint for `.a` imports.

## Selection and scope

Branch `codex/typeaware-wave-27` starts at requested
`origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.
The specific branch instruction takes precedence over the generic main instruction.
CLAUDE.md and its required language/memory documents, the type-aware README,
VOLUME_REPORT and COVERAGE_REPORT were read before edits.

The combined compiler/repository ranking is descending count, lexical name for
ties, excluding existing ports. The base has 26 ports, of which 25 occur in the
197-rule checker table (method-signature-style is outside it), leaving 172 entries.
Unported positions 79, 80 and 81 are:

| Position | Rule | Compiler / repository findings |
| --- | --- | --- |
| 79 | base/correctness-require-orm-column-nullable-parity | 0 / 0 |
| 80 | base/correctness-require-serializable-nullable-parity | 0 / 0 |
| 81 | base/correctness-require-verify-array-parity | 0 / 0 |

Before coding, every fetched origin branch's stage1 tree and claims were searched
for names and port filenames. Only inventory/evidence matches existed. None were
skipped. The claim was committed and pushed before implementation.

Each rule is in its own `.a` file. `parity_decorators.a` supplies shared local
syntax/option handling, raw-type lookup and diagnostic spans; `wave_27_suite.a`
is a standalone runner following the existing suite pattern. No new checker
question was necessary. No shared registration, compiler, pinned submodule or
existing rule file changed. Authored Adamic files and generated controls are `.a`.
The bridge's existing `.ts` implementation dependencies retain their names.

ORM relation exemptions, bare call names, literal nullable options and duplicate
reports follow Go. Serializable parity follows both directions, declines
nonliteral optional values and suppresses any explicit defaultValue. Verify
parity preserves the five array-level names, 25 known value names, unknown-Verify
suppression, sentinel decorators and parameter-property modifiers. Nullable
facts use pinned Any/Unknown/Undefined/Null/Void flags, including Void = 16.

## Independent oracle and controls

`testdata/oracle_wave_27.go` builds as an overlay in pinned cohere, loads programs
independently, and calls the unmodified production registry rules. It imports no
bridge code. It preserves config declaration roots and supports `.a` control
roots. Findings serialize byte positions, namespace, ID, exact message, fixes
and every suggestion/edit. Complete sorted streams, including file headers and
counts, are compared; no count-only shortcut or stderr normalization is used.
These three Go rules emit no fixes or suggestions, so their zero edit fields
are compared too.

There are 105 control files and 39 findings. They cover nullable unions, any,
unknown, void, never, explicit/nonliteral/absent options, default values,
relations, duplicate decorators/options, unsupported property keys, methods,
accessors, qualified/bare decorators, arrays/readonly arrays/tuples/aliases,
nonarray types, array suppressors/custom Verify decorators, constructor parameter
properties, Unicode identifiers, astral comments and CRLF byte offsets.

Frozen base manifests cover all 77 pinned TypeScript `src/compiler` files and
287 repository files. TypeScript commit:
`050880ce59e30b356b686bd3144efe24f875ebc8`.
Cohere: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
Typescript-go: `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
Both corpus comparisons have zero findings, matching the volume inventory;
positive controls ensure this cannot make the gate vacuous.

An initial run failed at byte 567 because I incorrectly used 16384 for Void.
Go's `void` control exposed the extra ORM report. The mask was corrected to 16
and subsequent complete gates passed. `controls-test.log` retains the failure;
it is not counted as a deliberate mutant or successful validation.

## Reproduction

Setup command: `bash cloud/setup.sh > /workspace/wave-27-setup.log 2>&1`, then
`source /workspace/adamic-tools/env.sh`. Go 1.27.1, clang 20.1.8 and Node
v24.19.0 each ready in 0s; submodules 0s; cache warm 116s; total 116s.
`nproc` = 5; cgroup `cpu.max` = `400000 100000`; memory 17.6 GB.

Relocate the preserved relative manifests to absolute paths in your checkout;
compiler entries are relative to the pinned TypeScript checkout. Then run:

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE27_ARTIFACTS=/workspace/wave-27-scratch/final-source \
ADAMIC_WAVE27_REPOSITORY_MANIFEST=/workspace/wave-27-scratch/repository.manifest \
ADAMIC_WAVE27_COMPILER_MANIFEST=/workspace/wave-27-scratch/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-27-scratch/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware \
  -run '^TestWave27AgreementAndMutants$' > final-source-test.log 2>&1

go test -v -count=1 -timeout 10m ./bridge/tsgo/checker > checker-test.log 2>&1
go test -v -count=1 -timeout 15m ./bridge/tsgo > bridge-test.log 2>&1
go vet ./stage1/cohere/typeaware > vet.log 2>&1
go test -v -count=1 -timeout 10m ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  > node-oracle-test.log 2>&1
```

The Node regex also selects generic_functions and method_closures: eight fixtures
plus the one-byte oracle mutant. This filtered gate passed in 87.821s. Checker
tests passed in 0.438s; bridge gate passed in 100.455s; vet and gofmt checks
produced no findings. Test output was written directly to files, never piped.

## Mutants, memory and limits

The wave test compiles and runs each source mutant, requires exit 0 and empty
stderr, and only then requires diagnostic bytes to disagree with production Go:

| Mutant | Decision changed | Catcher |
| --- | --- | --- |
| ORM | nullable test inverted | Byte 59, exit 0, empty stderr |
| Serializable | optional-true nullable test inverted | Byte 5554, exit 0, empty stderr |
| Verify array | suppression test inverted | Byte 8412, exit 0, empty stderr |
| Released registry | deletion of released handle removed | Expected panic 70 becomes exit 0 |

The normal post-release query panics 70 with exactly
`adamic: panic: invalid or released checker handle`. The base bridge gate
additionally checks stale/zero handles, repeated handles, answers surviving
release, Unicode offsets, and normal/native sanitized byte equality. Its own
mutants are caught by ASan (input/output length +1), byte mismatch (wrong source
file fact), expected refusal (removed link opt-in), and LeakSanitizer (missing
C buffer free and region allocation changed to unowned heap). The filtered Node
oracle's one-byte mutant is caught by the three-way output comparison.

All three normal populations run again under native ASan, UBSan and
LeakSanitizer with the sanitizer-built archive and match production Go. Go's
managed heap is not instrumented by C sanitizers. This validates the exercised
bridge/native paths; it does not prove every possible checker fact graph.

Pinned cohere's stock `--no-cache --fix` CLI reported `no files to check` for
these `.a` files. This is not a lint pass. The pinned house formatter was
invoked directly with a TypeScript parser hint while writing the original `.a`
files. An independent source loader ran all 276 enabled house rules with their
real options. After fixing descriptive names, iteration style and unused-loop
naming, 138 findings remain, all unsafe-type diagnostics caused by unresolved
`.a` imports. The final native compiler resolves them and passes the differential
and sanitizer gates. The source-lint workaround and its output are preserved,
not represented as a clean source-lint gate. No existing loader was modified.

Evidence is in `validation-wave-27/`: final complete test logs, canonical Go,
native and sanitized output streams, mutant output streams, setup and supporting
checks, manifests, measurement helper and SHA256 hashes. Scratch paths in output
headers make byte lengths/hashes specific to this run; comparison within each
population uses the same paths. Binaries, archives and downloaded corpora are
not committed. The complete repository gate and all upstream rule fixtures were
not run; the exact focused checks above were run instead.

## Final results and timing

Final frozen-source gate: PASS, 87.467s. Controls: 17,387 identical bytes,
39 findings. Repository: 18,485 identical bytes, zero findings. Compiler:
5,934 identical bytes, zero findings. Sanitized outputs match each stream.

Three alternating complete-process runs per engine were measured after builds
and gates finished, with stdout redirected and equality asserted on every run.
The preserved `measure.py` and `timing.json` give every sample and output hash.

| Corpus | Native median | Go median | Native / Go |
| --- | --- | --- | --- |
| repository | 411.045 ms | 211.820 ms | 1.941x |
| compiler | 1803.816 ms | 394.694 ms | 4.570x |

Native is slower on these corpora. Measurements include program creation,
loading, Adamic parsing/traversal and output serialization; Go uses its production
compiler AST walk. All timed corpus runs made zero bridge queries because the
relevant decorators are absent. These figures do not measure per-question
latency or extrapolate performance on decorator-heavy inputs. Load/run timing
stderr is preserved separately.
