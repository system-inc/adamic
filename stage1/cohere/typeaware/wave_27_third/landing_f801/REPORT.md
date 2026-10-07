Rebased wave 27 onto current main f8013f0b; nine owned ports revalidated.
Tested code c7cd4c91366e60f107bf22df6920e821b8d00f85; prior pushed tip 3cdcfc89.
All nine-rule, inherited bridge, numeric listener, sanitizer and handle gates PASS.
All 19 rule and nine listener mutants caught by Go bytes; five fact mutants caught likewise.
Three React claims and handed-node migration remain blocked; no new claims.

# Landing on f8013f0b

Main advanced with lowering changes, override refusals, Map/Set iterator and
number hashing fixes, and pragma checking. Rebase completed without conflicts;
the range diff shows every existing patch unchanged. Revalidation rebuilt the
compiler and native rule binaries from the tested code above. Only the owned
wave branch is published, using an explicit lease on the previous remote tip.
Main and area branches are never pushed by this unit.

## Commands and observations

Every toolchain shell sourced /workspace/adamic-tools/env.sh. TMPDIR was
/tmp/adamic-gate. nproc is 5. Earlier setup for this workspace passed with
Go/clang/Node/submodules ready 0s, build-cache warm 116s and total 116s.
No toolchain reinstall was needed on this unchanged toolchain.

```sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-27-scratch/typescript
export ADAMIC_WAVE27_REPOSITORY_MANIFEST=/workspace/wave-27-scratch/repository.manifest
export ADAMIC_WAVE27_COMPILER_MANIFEST=/workspace/wave-27-scratch/compiler.manifest
export ADAMIC_WAVE27_ARTIFACTS=/workspace/wave-27-scratch/f801-first
export ADAMIC_WAVE27_NEXT_ARTIFACTS=/workspace/wave-27-scratch/f801-next
export ADAMIC_WAVE27_THIRD_ARTIFACTS=/workspace/wave-27-scratch/f801-third
go test -v -count=1 -timeout 15m ./stage1/cohere/typeaware -run '^TestWave27AgreementAndMutants$' > /workspace/wave-27-f801-first.log 2>&1
python3 stage1/cohere/typeaware/wave_27_next/validate.py > /workspace/wave-27-f801-next.log 2>&1
python3 stage1/cohere/typeaware/wave_27_third/validate.py > /workspace/wave-27-f801-third.log 2>&1
go test -v -count=1 -timeout 15m ./bridge/tsgo/checker ./stage1/cohere/typeaware -run '^(TestCoverageAgreementAndMutants|TestCoverageCheckerQuestions|TestFactEncoding|TestShapeAndNameFacts|TestExactIndexMatchesCompilerNodes|TestAdamicRootKeepsConfigDeclarations|TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$' > /workspace/wave-27-f801-regression.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /workspace/wave-27-f801-vet.log 2>&1
```

First gate PASS 131.584s. Both isolated validators printed PASS. Focused
regression packages PASS 0.272s and 258.951s. Vet exited zero with empty output.
All subprocess output was written to log files; test runs were not piped.
Normal and Go-ASan archives were rebuilt and native --sanitize binaries used.

| Batch | Parseable control files | Findings | Identical output bytes |
| --- | ---: | ---: | ---: |
| Decorator parity | 105 | 39 | 17,177 |
| Stream/race correctness | 93 | 101 | 62,201 |
| Regex/rest/Hook dependencies | 605 | 437 | 225,707 |

All three owned batches compare complete findings, fixes and suggestions,
normal and sanitized, against unchanged production Go rules. Each also matches
on the 287-file repository corpus (zero findings, 18,485 bytes) and 77-file
TypeScript compiler corpus (zero findings, 5,934 bytes). The third batch still
has 34 Go-parser exclusions from its 639 candidates; exclusions are not passes.
Complete candidates, sources and exclusions are retained in JSON/manifests.
The inherited ten bridge ports match 53 control findings in normal and
sanitized builds and pass all ten rule mutants. Their other corpora were not
supplied to that inherited gate in this run.

## Mutants

The owned rule mutants reverse ORM nullable discrimination, Serializable
optional-true matching, Verify array suppression, process-exit pending-output
state, race timer ownership, blocking-stream early return, rest-parameter
implicit declaration detection, Hook missing-dependency message selection,
and regex constructor arity admission. All compile and exit zero with empty
stderr; only the independent Go output comparison catches each. First differing
bytes for the first batch are 57/5466/8264, next batch 80/52234/581, and third
batch 54/4962/112016. Full observations are in the corresponding logs.

Inherited before/cast/methods/coercion/caller/parameter/invariant/optional/alias/
unused mutants change initialization order, cast flags, method traversal,
coercion truth, caller flags, parameter anchoring, assignability direction,
optional field naming, edit range and arrow usage tracking. All ten are caught
by the independent Go byte comparison after normal exit.

Five raw fact mutants lose declaration-file status, mark every flow block
unreachable, force never return flags, make every import type-only, and replace
binding-pattern name kind with Identifier. All exit zero with empty stderr and
are caught only by Go bytes. The next-batch log names every offset.

Nine numeric listener mutants increment the first kind of exactly one owned
declaration. Each compiles and exits zero with empty stderr, and the independent
Go enum bytes catch it. listeners/mutants.json records every file and offset.
Fresh native and --sanitize manifest builds match Go. The oracle is rebuilt
inside the pinned TypeScript/tsc module using the owned oracle.go.txt overlay.
These checks validate declarations, not handed-node dispatch or performance.

Released raw-type, four stream questions, binding-declarations,
node-symbol-details and type-name requests are refused with exact panic 70.
Each gate's retained-registry mutant exits zero and is caught by that required
refusal. Wrong-kind and unknown-question guard mutants exit zero and fail the
required refusal too. The nineteen malformed-wire inputs (empty, negative or
short length, trailing data, wrong version/question, invalid/rounded/noncanonical
integer, negative natural, bad boolean, zero ID, negative tuple count, missing
root/constraint/element, duplicate ID, missing links 17/18) all panic 70 with
their expected messages. Regression logs retain every case. Native sanitizer
runs produce no sanitizer or leak diagnostics.

## Timing and limits

Whole-process, three alternating samples per engine retain identical complete
output. Medians include loading, parsing, rule judgments, rendering and teardown.
They are not just bridge query timings. Stream/race repository native 536.436ms
versus Go 178.590ms (3.004x); compiler native 3,945.780ms versus Go 511.194ms
(7.719x). Third batch repository native 415.940ms versus Go 198.732ms (2.093x);
compiler native 3,385.326ms versus Go 519.214ms (6.520x). Earlier decorator single
samples overlapped the regression gate: repository 330.778ms versus 174.226ms;
compiler 2,229.710ms versus 410.073ms. They are not quiet medians.

The pending static-component positive control was rebuilt with the new native
compiler. Native still exits 70 with expected GreaterThanToken, got SlashToken
at 84; Go exits zero with one staticComponents finding at bytes 74:83.
Exact streams and commands are in wave-27-f801-parser.json. JSX support exists
on origin/codex/stage1-jsx-lint but has not landed on main. Shared ParseNode
still has only string kind; numeric handed-node migration remains incomplete.
The published rule.json registry requires actual factories and node callbacks,
which are not supplied by these numeric declaration sidecars. No new rule.json
registration or performance-compliance claim is made.

Ahra's instruction remains: "If anything else blocks you, say exactly what it
is and stop, rather than editing shared files." The parser/interface changes
are outside permitted rule directories. No further rules were claimed. Full
repository tests, emitted-JavaScript comparison, nondefault options and the
other sixteen inherited rule gates were not rerun here. No protected compiler,
shared generator or shared test harness source was edited by this unit; main's
compiler changes were received through the rebase. No shared finding-model sha
was named in this message, so no separate model branch was imported.
