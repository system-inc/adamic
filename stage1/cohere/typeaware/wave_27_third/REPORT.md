# Wave 27, third batch

Claimed and pushed before implementation in 8e5789a9. The previous six ports,
tests and evidence were pushed at c0cd687567bcacbce9d6be2383ee3dd17ed3ca31.
Selection fetched all origin heads without submodule recursion and checked the
197 checker-dependent rules in descending combined compiler/repository volume,
with lexical ties. Ports on main and codex/tsgo-c-library and Markdown claims
on all 356 origin refs were excluded, including unqualified/underscore spellings.
The first three available rules were prefer-regex-literals, prefer-rest-params
and react-hooks/exhaustive-deps, each with combined count zero. The complete
selection audit is in evidence.

## Implementation and boundary

Each rule has its own .a file in this directory. suite.a loads one program and
runs the rules over each native Adamic parser tree, sorting complete canonical
lines and retaining duplicates. Complete findings include message IDs/text,
UTF-8 spans, fix lists, and all suggestion IDs/text/edits in original edit order.
The independent oracle has its own loader/walk, production program views and
file cache, and invokes the three unchanged Go cohere rules with default options.
It imports no bridge code and contains no copied rule predicate.

No new checker operation is needed. The existing wave-27-declaration-ancestry
facts supply raw symbol identity and source declaration origins. Existing
binding-declarations resolves local declarations, including shorthand values.
All rule predicates, dependency trees, stability/containment checks, regex
tracking, regex grammar/rewrite checks and repair construction execute natively.
No shared generator/harness, Go registration, or protected compiler file changed.

prefer-rest-params discriminates the implicit arguments symbol by its empty
declaration list, preserving declared shadows and the dotted-access exemption.
prefer-regex-literals follows aliases, global objects and destructuring, preserves
its production static-string whitelist, and builds literal suggestions only
when comments, preceding tokens, grammar and printable-character gates permit.
exhaustive-deps preserves component/callback containment, binding-based stability,
property-prefix satisfaction, entry findings, construction warnings, setter loops,
optional path spelling and conditional suggestion ordering. Default suggestions
for corrected Hook arrays contain descriptive text and zero edits, exactly as Go
cohere does; those zero-edit suggestions are serialized and compared too.

The fixed runner exposes default options only, as the base suites do. The regex
redundant-wrapping option and Hook additionalHooks, explicit-array, compiler-auto
array and dangerous-autofix option paths are outside this port's exposed interface.
The rule bodies must not be advertised as a full configurable cohere CLI.

## Reproduction

Source /workspace/adamic-tools/env.sh, then from the repository root:

```bash
ADAMIC_WAVE27_THIRD_ARTIFACTS=/workspace/wave-27-scratch/third-final \
ADAMIC_WAVE27_REPOSITORY_MANIFEST=/workspace/wave-27-scratch/repository.manifest \
ADAMIC_WAVE27_COMPILER_MANIFEST=/workspace/wave-27-scratch/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-27-scratch/typescript \
python3 stage1/cohere/typeaware/wave_27_third/validate.py \
  > /workspace/wave-27-third-final.log 2>&1
```

The validator builds native stage 0, the normal and Go -asan C archives, the
native normal and --sanitize programs, and the independent Go overlay oracle.
It records every subprocess's stdout/stderr directly in separate files. Timed
rounds alternate native/Go, run after builds/checks, and compare their complete
streams again. Source/diagnostic hashes and relative corpus manifests are retained.

Toolchain setup was completed in this same wave: Go 1.27.1 ready 0s, clang
20.1.8 ready 0s, Node v24.19.0 ready 0s, submodules 0s, build cache warm
116s, done 116s; nproc 5, CPU quota 4 cores. Setup log is retained. Cohere
715ba94f3608a6500086b1076ce5cb7e51b836db and typescript-go
8d550c837c90bd1805b047b7eeccc2baac2d5e7a pins are unchanged. Compiler
corpus is TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8; populations
remain all 77 compiler files and the base's frozen 287 repository files.

## Final observations

The final command above completed with PASS. Of 639 generated candidates, 605
parsed in the pinned Go loader and were compared. They produce 437 findings:
15 prefer-rest-params, 237 exhaustive-deps, and 185 prefer-regex-literals.
All 226,312 control bytes match, including zero automatic fixes and 306 complete
suggestions. The normal and sanitizer native stderr streams are empty.
Repository: 287 files, zero findings, 18,485 matching bytes normal/sanitized.
Compiler: 77 files, zero findings, 5,934 matching bytes normal/sanitized.
The sanitizer archive uses Go -asan and the native --sanitize build checks
ASan/UBSan/LeakSanitizer, including output-string ownership and program teardown.

Every rule mutant compiled, exited 0 with empty stderr, and was caught only by
the independent Go comparison:

| Rule | Mutation | First differing byte |
| --- | --- | ---: |
| prefer-rest-params | Reverse the empty declaration-list discriminator | 55 |
| exhaustive-deps | Reverse missing-list warning selection | 4986 |
| prefer-regex-literals | Reject the accepted constructor arities | 112387 |

Both reused questions, wave-27-declaration-ancestry and binding-declarations,
refuse a released program with exit 70 and exactly
`adamic: panic: invalid or released checker handle`. Keeping that handle in the
registry makes both exit 0, which the required-panic expectation catches.

Three alternating complete-process timing rounds run after compilation,
sanitation and mutants; every timed finding stream is re-compared:

| Corpus | Native median | Go median | Native / Go | Native queries/run |
| --- | ---: | ---: | ---: | ---: |
| repository | 471.009 ms | 214.973 ms | 2.191x | 25 |
| compiler | 3674.298 ms | 532.066 ms | 6.906x | 166 |

These are observed process times including loading, parsing, rules, rendering,
startup and teardown, not bare bridge-call latency. Native is slower. A zero
finding count does not imply a zero checker-question count. Raw phase timings,
three samples per implementation, stream bytes and hashes are preserved.

After timing, focused checker/framing/refusal regressions passed:

```bash
go test -v -count=1 -timeout 10m ./bridge/tsgo/checker ./stage1/cohere/typeaware \
  -run '^(TestCoverageCheckerQuestions|TestFactEncoding|TestShapeAndNameFacts|TestExactIndexMatchesCompilerNodes|TestAdamicRootKeepsConfigDeclarations|TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$' \
  > /workspace/wave-27-third-regression.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware \
  > /workspace/wave-27-third-vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave_27_third/oracle.go.txt \
  > /workspace/wave-27-third-gofmt.log 2>&1
```

PASS: checker 0.177s, type-aware package 49.060s; vet and gofmt logs empty.
The guard regression also catches 19 malformed wire mutants with panic 70:
empty, negative/short lengths, trailing data, wrong version/question, noninteger,
rounded/noncanonical integer, negative natural, bad boolean, zero identity,
negative tuple, missing root/constraint/element, duplicate identity and missing
links 17/18. Wrong-kind and unknown-question request guard mutants exit 0 and
fail the required-refusal check; normal requests refuse 70. The complete log
records every mutation and exact refusal. No compiler/runtime file changed,
and the previous wave's filtered Node and foundation sanitizer evidence remains
in its reports; those broader gates were not rerun in this batch.

## Recorded limitations and preliminary failures

The pinned Go loader reads .a as TypeScript, so JSX controls fail parsing.
The evidence records every candidate and the independent parser's exclusions;
excluded controls are not called passes. Two candidates are partial strings from
concatenated Go fixture expressions, an extraction limitation. The other 32 are
JSX cases. Dedicated complete numeric React.useState controls supplement those
partial extractions. Native parser JSX agreement, all upstream options/fixtures,
full go test ./..., emitted-JavaScript agreement and a clean stock .a source-lint
run are not claimed. The shared .a loader/harness work is owned by
codex/lint-harness-dot-a; it was not edited here.

The first control config used files:[] and the independent loader refused it.
A combined control run then refused the first JSX .a root. Both failures are
retained, rather than called comparisons. Native stage 0 rejected an untyped
empty-array fallback (never[]) and positioned lastIndexOf; owned source changes
use a typed fallback and substring scans. Two redundant recursive-object-tree
compilations were terminated after several CPU-bound minutes; the final tree
uses integer links and compiles successfully. One formatter attempt ran from the
wrong Go module and failed; the retry ran the same overlay in cohere and succeeded.
No shared/compiler workaround was used. Source formatting uses cohere's house
formatter on the .a bodies through the isolated overlay; stock source lint is
still limited by its .a import handling.
