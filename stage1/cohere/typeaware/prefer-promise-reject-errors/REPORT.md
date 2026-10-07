Built: native core prefer-promise-reject-errors, prefer-regex-literals, and prefer-rest-params, including both options and regex suggestions.
Commits: prior six rules deb33e49; third claim 9026a958 pushed before code; implementation is the commit containing this report.
Commands and outputs: isolated gate PASS 682 projects/502 findings, both corpora normal and ASan/UBSan/LSan; checker, vet, source lint/formatting, filtered Node passed.
Mutants: Promise and rest decisions, regex suggestion end +1, binding-origin ambient answer, and released registry retention all caught.
Not covered: one invalid TSX fixture, full repository gate, shared-harness integration and emitted-JavaScript rule comparisons; automatic fixes are empty in production.

The previous six rules were ported, tested, and pushed before this claim. A fetch
of all origin heads found 356 refs, 33 distinct Markdown claim blobs, 132 claimed
checker-dependent inventory rules, and 25 existing checker-dependent ports on
main/the bridge branch. Forty candidates remained before this batch. The first
three by the recorded combined volume and lexical tie ranking were the three
core rules above, all volume zero. Claims were checked on every fetched origin
branch; compressed test evidence beneath claims/ was not treated as a claim.
The claim update was pushed before implementation. No additional rules were
claimed after this batch.

All new Adamic sources are `.a`, in the three owned rule directories. The only
shared code edit is the binding-origin switch registration in checker/facts.go,
one logical entry (two physical lines after gofmt). The new raw question is in
its own Go and Adamic files. No shared generator, test harness, parser, native
emitter/lowerer, protected oracle source, or submodule pin was edited.

The question supplies original and read symbol identities and declaration
file/kind/span/ambient metadata, without alias resolution or lint verdicts.
Native code decides implicit arguments, optimistic Error possibilities,
executor references, global-reference tracing, regexp scanning, reports and
suggestion edits. The independent oracle imports no bridge code and calls the
unchanged production Go cohere rules with identical options.

The environment reuses the successful unit setup: `bash cloud/setup.sh`, then
`source /workspace/adamic-tools/env.sh`. Timing lines: Go ready 0s, clang ready 0s,
Node ready 0s, submodules ready 0s, build cache warm 163s, done 163s. `nproc` is 5
with a four-CPU quota. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.

`README.md` gives the portable reproduction command. The final gate command was:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py \
  --repository /workspace/adamic \
  --artifacts /workspace/wave20-validation/third/passing \
  --compiler /workspace/wave20-typescript \
  --cases /workspace/wave20-validation/third/valid-cases \
  > /workspace/wave20-validation/third/passing-gate.log 2>&1
```

It passed 662 valid upstream projects and 20 added controls. The original Go
upstream assertions also passed, 0.485s. Their inputs and both options were
captured through scratch overlays, without changing the harness or tests on
disk. The running gate's final label says '682 upstream projects'; the true
breakdown is 662 upstream plus 20 controls, as recorded in coverage.json and
fixed in the committed script's label.

One upstream source writes `Promise.reject(<string>'x');` in input.tsx. The
independent Go loader rejects that invalid TypeScript/JSX combination with parse
diagnostics. It is excluded and preserved in excluded.json, not counted as a
successful agreement. The first full gate attempt also found that relocation
had omitted a nested fixture directory. The portable tape now retains relative
paths and every input recursively; the final run covers all 682 restored inputs.

The additional controls exposed a real missed case: `(resolve, ...reject)` as a
Promise executor parameter. Selecting the first child mistook the rest token
for the parameter name. Native now selects the binding name across modifiers;
the retained control agrees with Go and sanitizers. The failed extra-control
run is preserved. Stage 0 refused positional lastIndexOf; an owned source slice
before lastIndexOf uses supported operations with the same bounds. No compiler
workaround was made in shared files.

Every normal and ASan/UBSan/LSan project has identical full diagnostic bytes and
empty sanitizer stderr. Fields include spans, rule and message IDs, message text,
fix counts/edits, suggestion IDs/text and all suggestion edit spans/text. Option
flags are preserved on both sides, including false/default and true variants.

| Rule | Findings across project fixtures |
| --- | ---: |
| prefer-promise-reject-errors | 68 |
| prefer-regex-literals | 413 |
| prefer-rest-params | 21 |

The fixture runs total 502 findings, 0 automatic fixes, and 365
suggestions. All three production rules have no automatic fixes; no nonempty
automatic-fix coverage is claimed. The regex suggestions include actual edits
and the two alternative flag messages/edits. Comparison preserves duplicates
and repair order and sorts complete canonical finding lines.

The frozen repository manifest has 287 roots, 18,485 output bytes, zero findings;
the TypeScript v6.0.3 compiler manifest has 77 roots, 5,241 bytes, zero findings.
Both normal and sanitized streams agree with Go. Positive controls are required
for every rule, so the zero-finding corpora cannot make the gate vacuous.

Every rule and answer mutant compiles, exits 0, and produces empty stderr. Only
the independent Go/native byte comparison catches it:

| Mutant | Fixture | First differing byte | Detector |
| --- | --- | ---: | --- |
| promise | case-1001187216 | 77 | Full diagnostic bytes |
| regex-suggestion-span | case-102540695 | 615 | Full diagnostic bytes |
| rest | case-1124234137 | 106 | Full diagnostic bytes |
| binding-origin | case-1006578395 | 77 | Full diagnostic bytes |

The regex mutant increments a suggestion's end by one while retaining finding,
fix and suggestion counts and messages. Its kill proves the edit boundaries are
compared. The released binding-origin probe must panic 70 with
`invalid or released checker handle`. Retaining the released program in the
registry makes it exit 0, caught by that required-panic assertion. All bridge
mutations use scratch Go overlays, never repository edits.

Additional checks:

```sh
go test ./bridge/tsgo/checker -count=1
go vet ./...
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(strings|closures|generic_functions)[.]a$' -count=1 -v -timeout 5m
```

Checker PASS 0.449s, including raw symbol/declaration comparisons, shorthand read
identity, implicit arguments, ambient/missing symbols, and kind/suffix refusal.
Vet produced an empty log. Filtered Node oracle PASS 19.178s, with the one-byte
failure probe and the selected Node/backend/native/sanitizer fixtures. All 15
new Adamic files format idempotently and configured source lint has zero findings.
Source validation uses in-memory virtual TypeScript names against pinned Go
cohere, with no physical `.ts` ports. Gofmt and git diff checks are clean.
The full repository test gate was not run.

Three alternating native/Go benchmark rounds ran without concurrent builds or
tests. All timed streams are byte-identical; per-round hashes and phase timings
are in bench.json. Medians in seconds:

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| repository | 0.370926 | 0.221599 | 1.674 |
| compiler | 2.126633 | 0.516472 | 4.118 |

Observation: native is slower for these isolated entries; no speedup is claimed.
Shared profile compilation and emitted-JavaScript rule comparison remain outside
this unit. There is no remaining local rule blocker, and no fourth new rule was
claimed. Validation includes the complete per-command streams in one compressed
archive, final and failed-attempt logs, exclusions, fixture counts, mutant
evidence, and benchmark hashes. The fixture tape and gate reproduce the required
Go/native comparisons without editing shared files.
