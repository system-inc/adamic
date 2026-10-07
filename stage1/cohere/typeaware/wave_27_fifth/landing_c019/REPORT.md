Rebased wave 27 onto main c01907a7; twelve completed ports are green.
Tested code baa46eecb00aa6527a4d10a84358b7bd95fafffa; only codex/typeaware-wave-27 is published.
All four owned Go byte gates, sanitizer corpora, handles and focused regressions PASS.
Twelve rule mutants and five fact mutants exit 0; Go bytes catch every one.
Three React analysis claims remain parked; shared registry and emitted JS untested.

# Landing on c01907a7

Claim f2a14e47f was pushed before fifth-batch implementation. Rebase rewrites that
claim to df4f2d8a3 and the parking commit to 55c44ddba. The initial fifth-batch native
ports are bbbd34619581c9f673315e02a0aff8ef07acfed2; the optional generic-call
correction is baa46eecb00aa6527a4d10a84358b7bd95fafffa. All 16 original patches rebased without conflict.
The final fetch confirms c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 is still main.
Source changes received from main are in stage3 and its hook/documentation;
checker, parser, lowerer and native compiler inputs are unchanged by that rebase.
Main and area branches are never pushed by this unit.

## Commands and observed outputs

All shells source /workspace/adamic-tools/env.sh, with TMPDIR=/tmp/adamic-gate.
Earlier cloud/setup.sh passed: Go/clang/Node/submodules ready 0s, warm cache 116s,
total 116s. nproc=5. The current-main compiler and archives were rebuilt by the
first three runners; the fifth runner uses that just-rebuilt compiler.
All subprocess output is captured directly to log files, never piped.

```sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-27-scratch/typescript
export ADAMIC_WAVE27_REPOSITORY_MANIFEST=/workspace/wave-27-scratch/repository.manifest
export ADAMIC_WAVE27_COMPILER_MANIFEST=/workspace/wave-27-scratch/compiler.manifest
export ADAMIC_WAVE27_ARTIFACTS=/workspace/wave-27-scratch/f801-first
export ADAMIC_WAVE27_NEXT_ARTIFACTS=/workspace/wave-27-scratch/f801-next
export ADAMIC_WAVE27_THIRD_ARTIFACTS=/workspace/wave-27-scratch/f801-third
go test -v -count=1 -timeout 15m ./stage1/cohere/typeaware -run '^TestWave27AgreementAndMutants$' > /workspace/wave-27-c019-first.log 2>&1
python3 stage1/cohere/typeaware/wave_27_next/validate.py > /workspace/wave-27-c019-next.log 2>&1
python3 stage1/cohere/typeaware/wave_27_third/validate.py > /workspace/wave-27-c019-third.log 2>&1
export ADAMIC_WAVE27_FIFTH_ARTIFACTS=/workspace/wave-27-scratch/fifth-c019
export ADAMIC_WAVE27_STAGE0=/workspace/wave-27-scratch/f801-third/adamic
python3 stage1/cohere/typeaware/wave_27_fifth/validate.py > /workspace/wave-27-c019-fifth.log 2>&1
go test -v -count=1 -timeout 15m ./bridge/tsgo/checker ./stage1/cohere/typeaware -run '^(TestCoverageCheckerQuestions|TestFactEncoding|TestShapeAndNameFacts|TestExactIndexMatchesCompilerNodes|TestAdamicRootKeepsConfigDeclarations|TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$' > /workspace/wave-27-c019-regression.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /workspace/wave-27-c019-vet.log 2>&1
```

The first three runners ran independently in parallel; their benchmark samples
were contended and are not presented as quiet performance comparisons. The
fifth runner completed its timing after those runners and regression finished.
First gate PASS in 134.692s. Next, third and fifth validators each print PASS.
Regression packages PASS in 0.236s and 75.489s; vet exits zero with empty output.

| Owned batch | Parseable controls | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| ORM/Serializable/Verify parity | 105 | 39 | 17177 |
| Output/race/blocking stream correctness | 93 | 101 | 62201 |
| Regex/rest/Hook dependencies | 605 | 437 | 225707 |
| Await/symbol-description/typeof | 186 | 108 | 64739 |

Each batch matches full finding ranges, ids and descriptions, automatic fixes,
ordered suggestions and their ordered edit ranges/text against independent
production Go cohere. Every batch matches the 287-file frozen repository corpus
(18485 bytes, zero findings) and TypeScript's 77 compiler files (5934 bytes, zero
findings), normally and with Go ASan archives plus native --sanitize binaries.
Sanitizer stderr is empty. Positive controls and normally exiting mutants keep
the clean corpus results from passing vacuously. The third batch retains 34
Go-parser exclusions from 639 candidates. Fifth has 24 exclusions from 210;
all candidates and exclusions are retained and exclusions are not passes.

The fifth batch's own kind-indexed driver caches parser nodes and passes each
numeric Node to its declared listener. Callbacks do not compare kind strings or
refetch the handed node. Its rule.json files declare numeric kinds, all matched
against Go. 351 numeric constants are independently checked. The string adapter
is confined to the existing parser/wire boundary. The previous nine ports have
unused numeric sidecars but have not completed that handed-node migration.

## Mutants and handle checks

Every rule mutant compiles and exits zero with empty stderr. Go byte comparison
alone catches ORM nullable inversion (57), Serializable optional-true inversion
(5466), Verify array suppression inversion (8264), process pending-output
inversion (80), race ownership inversion (52234), blocking early-return inversion
(581), rest implicit-declaration inversion (54), Hook missing-message selection
inversion (4962), regex arity inversion (112016), symbol first-declaration ambient
inversion (39039), typeof legal-string replacement (46485), and await contract
bypass (33546). These numbers are first differing bytes in their retained streams.

Five bridge fact mutants lose declaration-file status (56), mark all CFG blocks
unreachable (1318), force never return flags (12452), mark all imports type-only
(42069), and force declaration name kind Identifier (61584). They also exit zero
with empty stderr and are caught only by Go output. The next log contains each.

Raw-type, four existing wave questions, binding-declarations and all four new
checker-link operations reject released handles with exact panic 70. Each owned
batch's retained-registry mutant exits zero and fails the required refusal check.
The new bridge exposes raw symbol declarations without alias following, generic
signature links, type/property/index/call-signature links and class heritage type
identities. It supplies no lint verdict. All contract traversal and thenability
judgments are native.

The nineteen malformed-wire guards, pinned flags and checker question regressions
pass. Wrong-kind and unknown-question guard mutants exit zero and are caught by
required-panic assertions. Full results are in regression.log. The inherited ten
rule port mutants were checked on f8013f0b; they were not repeated in this landing.
Numeric sidecar mutants for the previous nine were likewise not rerun here.

The optional generic-call witness first produced Go=1 finding/native=0: the
QuestionDotToken was mistaken for an explicit type argument. The corrected native
walk ignores only that punctuation, and the permanent witness is part of the
186-file complete control comparison. Initial disagreement is retained separately.

## Native time against Go

Fifth batch, three alternating whole-process samples, includes program loading,
parsing, judgments, rendering and teardown. Repository median native 547.758ms
versus Go 154.696ms (3.541x); compiler native 3249.782ms versus Go 423.911ms
(7.666x). Streams remain identical in all six samples per corpus. These results
do not claim a speedup; numeric handed-node declarations address dispatch shape,
while native parsing and the compatibility adapter still contribute costs.

## Limits

React effect/render/static-component claims are parked on native high-level IR,
single-assignment/capture/post-dominance analysis and, for the JSX witness, parser
support. They count finished only under Ahra's landing-cap exception, not as
implemented or byte-green ports. See wave_27_fourth/PARKED.md and the claim file.

Shared registration discovery and emitted-JavaScript comparison remain untested.
The incoming shared registry needs numeric handed-node integration; this unit
uses its own typed context/factories/driver and does not edit that shared work.
ValidTypeof's nondefault requireStringLiterals constructor is not covered by the
published default-option oracle. Full repository tests and the other inherited
rule gates were not run. No protected compiler, shared parser, registration
generator or shared test harness file was edited. The shared Go dispatcher
changes one physical registration line; new facts live in their own Go/.a files.
