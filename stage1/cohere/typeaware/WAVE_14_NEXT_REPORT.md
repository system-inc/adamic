Built: two native continuation rules; leaked-number-render remains blocked on JSX parsing.
Commits: original implementation c043d14a; continuation claim 1c5cf1bd; implementation accompanies this report.
Commands and outputs: continuation PASS 87.690s; bridge PASS 81.420s; Node PASS 24.020s; vet and formatting clean.
Mutants: listener and namespace judgments caught only by complete bytes; released registry caught by required panic; bridge foundation mutants pass.
Not covered: leaked-number-render implementation or mutant, full repository gate, every upstream fixture, emitted-JavaScript comparison for these rules.

# Wave 14 continuation

The original three rules were already complete and pushed. The premature
continuation request arrived next; the claim update was pushed before any code.
Ahra's correction arrived after implementation of the two non-JSX rules and
while the final checks were running. No further claims or shared-file edits
were made after that correction. No pull request was opened.

All 320 origin refs were fetched without submodule recursion. The full ranking
has 197 checker-dependent entries. The original 26 ports exclude 25 ranked
entries; method-signature-style is outside this ranking. Ninety-three ranked
entries are named in claims across the fetched branches. The first three of
79 remaining entries were these zero-volume rules:

- nexus/correctness-no-global-listener-target-assertion: implemented.
- nexus/correctness-no-leaked-number-render: unimplemented, JSX parser blocker.
- nexus/correctness-no-mock-on-module-namespace: implemented.

The collection-misuse and discarded-result rules were already claimed by other
continuations. Selection and branch-tip evidence is in
[selection.json](validation-wave-14-next/selection.json). The compared remote
bridge tip was 5afbdb83; main was ef3d907e. Counts and configuration mentions
were distinguished from native ports; the original production Go oracle's 16
subjects and the 10 coverage selections identify the excluded native ports.

Each implemented rule lives in its own .a file. Three raw questions have their
own Go files and Adamic decoders: symbol-context, interface-bases and
signature-context. They return declaration origins, identities, nonnullable
identities, union members, interface bases, signature ancestors and the emit
module kind. They return no lint verdict, finding, fix or suggestion. Only
three switch entries, six gofmt lines, were added to the shared checker dispatch
before Ahra's correction. No shared registration generator, existing test
harness, protected compiler file or submodule pin changed.

## Complete diagnostic comparison

The independent Go oracle loads the pinned production rules and walks its own
compiler AST. Native judgment uses Adamic's parser and the raw checker facts.
Both serialize complete messages, byte spans, fixes and suggestions, retaining
all duplicates. These rules produce no fixes or suggestions; those fields are
still part of the comparison. All native sanitizer runs had empty stderr.

| Population | Roots | Findings | Identical bytes | Normal and ASan/UBSan/LSan |
| --- | ---: | ---: | ---: | --- |
| Generated controls, including ambient module declarations | 15 | 15 | 9142 | PASS |
| TypeScript src/compiler | 77 | 0 | 5318 | PASS |
| Frozen repository corpus | 287 | 0 | 18485 | PASS |

Controls exercise document/window library origins, inline and named handlers,
const versus mutable handlers, nested event captures and shadowing, narrow and
general element assertions, nullable unions, chained assertions, angle-bracket
assertions, already-proven instanceof guards, local receiver exemptions,
Unicode/CRLF spans, namespace/type-only/default/copy exemptions, all four mock
methods, wrappers and MockTracker declarations in nested namespaces. CommonJS
and NodeNext produce seven listener findings and no namespace findings;
Preserve and ES2020 produce all fifteen, matching Go completely.

| Mutant | Changed judgment | What caught it |
| --- | --- | --- |
| Listener assertion | Reverse the known-to-asserted assignability branch | Go byte comparison, byte 91 |
| Namespace mock | Accept another ambient module instead of node:test/test | Go byte comparison, byte 4619 |
| Released registry | Retain a released program handle | Required panic 70 disappears; mutant exits 0 |

Both rule mutants compile, exit 0 and have empty stderr. No compile failure or
sanitizer catches them. The normal released-handle probe exits 70 with
`invalid or released checker handle`. The bridge regression additionally kills
input/output length mutants with ASan, missing output-free and heap-region
mutants with LeakSanitizer, wrong-position with the independent byte oracle,
removed link opt-in with refusal, and retained handles with the stale-handle
check. Raw question tests check origins, type identities, signature ancestry
and malformed-request rejection.

## The exact blocker

The .a witness is:

```tsx
declare const count:number;export const view=<p>{count&&'some'}</p>;
```

The external Go test presents those same .a file contents as virtual TSX to
Go's production typed rule harness. Go reports `leakedNumberRender` at bytes
49:54 with its complete message, zero fixes and zero suggestions. Native
refuses the .a witness with exit 70 and:

```
adamic: panic: parser slice expected CloseBraceToken, got AmpersandAmpersandToken at 54 in /workspace/wave-14-next-final/jsx-witness.a
```

The native parser has no JSX child representation for this rule to visit.
Zero findings on the two non-JSX corpora do not prove a port of this rule.
It has no native implementation or comparison-only mutant. Ahra instructed
workers to state blockers and stop rather than edit shared files, so no parser,
shared harness or registration-generator change was attempted. The original
three-rule wave remains complete; this continuation is partial.

## Native time against Go

Three alternating count-only rounds, run after the tests and builds finished.
These numbers measure the two implemented rules; the Go oracle also registers
the leaked-render rule, whose JSX listener has no matches in these populations.
Medians are seconds and include checker loading in process time.

| Population | Implementation | Process | Load | Run |
| --- | --- | ---: | ---: | ---: |
| Compiler | Native | 1.344788 | 0.247174 | 1.085612 |
| Compiler | Go | 0.294746 | 0.237537 | 0.040861 |
| Repository | Native | 0.193762 | 0.063544 | 0.124210 |
| Repository | Go | 0.120553 | 0.062967 | 0.044740 |

Native process time is 4.56 times Go on the compiler and 1.61 times Go on the
repository. No performance parity is claimed. Raw rounds are in timings.json
and bench.log. The earlier wave's timings remain in WAVE_14_REPORT.md.

## Commands and environment

Each Go command sources `/workspace/adamic-tools/env.sh`. Setup passed: Go,
clang, Node and submodules ready at 0s; cache warm 25s; total 25s. `nproc` is 5,
CPU quota is four cores, memory 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0.
The TypeScript source pin and both frozen populations match the earlier wave.

```sh
bash cloud/setup.sh > /tmp/wave-14-next-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
ADAMIC_WAVE14_NEXT_ARTIFACTS=/workspace/wave-14-next-final ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript ADAMIC_WAVE14_NEXT_COMPILER_MANIFEST=/workspace/wave-14-artifacts/compiler.manifest ADAMIC_WAVE14_NEXT_REPOSITORY_MANIFEST=/workspace/wave-14-artifacts/repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave14NextAgreementAndMutants$' -count=1 -v -timeout=30m > /tmp/wave-14-next-final.log 2>&1
go test ./bridge/tsgo/... -count=1 -v -timeout=30m > /tmp/wave-14-next-bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -v -timeout=30m > /tmp/wave-14-next-node.log 2>&1
go vet ./... > /tmp/wave-14-next-vet.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /tmp/wave-14-next-gofmt.log
python3 bridge/tsgo/profile/volume_bench.py /workspace/wave-14-next-final/wave-14-next /workspace/wave-14-next-final/wave-14-next-oracle /workspace/wave-14-next-bench --corpus compiler /workspace/wave-14-typescript/src/compiler/tsconfig.json /workspace/wave-14-artifacts/compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-14-artifacts/repository.manifest > /tmp/wave-14-next-bench.log 2>&1
```

All test output went directly to files. Logs, compressed complete outputs,
output hashes, source hashes and portable manifests are in
[validation-wave-14-next](validation-wave-14-next/). The six new .a source files
were formatted through the pinned cohere native printer with virtual TypeScript
filenames; no new .ts program was authored. That is source formatting, not a
claim that the pinned full CLI linted .a files. The correction assigns .a module
support and emitted-JavaScript comparison in the shared harness to another
worker, and this unit does not modify that work. The full repository gate,
the original 26 rules' corpus reruns, every upstream fixture and the leaked-rule
native mutant remain outside this continuation's measured coverage.
