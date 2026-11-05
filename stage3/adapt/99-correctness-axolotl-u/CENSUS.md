# 99u compiler census

The ordinary compiler and per-file driver are unmodified main b0efce6d, built
with `go build ./cmd/adamic ./stage3/census/tool`. Every one of the 79 TypeScript
roots is loaded and lowered independently before and after. Both ordinary
censuses completed. Non-TypeScript and whole-program rows remain in the raw
streams; they are excluded from the 79-root groups. The ordinary whole-program
row still reports the compiler's single-entry limitation.

The five complete owner bodies are additionally observed through the meter's
existing `LatentReplay` selector and full attempt path, using its unmodified
make_overlay.py output and `LATENT_FULL=1`. Both the ordinary loader and lowerer
no-output guards are checked. Each side loads all 79 roots, has zero checker
diagnostics, and attempts exactly the requested complete owner, including its
reachable dependencies. No roots or bodies are reduced, no diagnostic or
refusal is suppressed, and no usable IR is returned. owner-census.go.txt is the
measurement driver; the production compiler has no corresponding change.
It uses selection without requesting a finding-signature match, and retains
that API's signature-mismatch message as metadata, not as a compiler finding.

Two exploratory whole-selected-file scans were superseded before completion
by these supported complete-owner observations. full-file-pilots.json records
their incomplete status. Their partial headers are not treated as finished
censuses. The complete ordinary per-file census is independent of those pilots.
The historical meter label in records says "measured on a checker-rejected
program"; the actual zero-diagnostic headers govern interpretation.

## Ordinary per-file groups

| Side | Roots | Kind | First stop | Reason |
| --- | ---: | --- | --- | --- |
| before | 73 | Refused | src/compiler/core.ts:736:36 | a cast the runtime can't check |
| before | 1 | NotYet | src/compiler/debug.ts:243:59 | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| before | 2 | accepted | - | - |
| before | 1 | NotYet | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| before | 1 | NotYet | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration |
| before | 1 | NotYet | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration |
| after | 73 | Refused | src/compiler/core.ts:736:36 | a cast the runtime can't check |
| after | 1 | NotYet | src/compiler/debug.ts:243:59 | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| after | 2 | accepted | - | - |
| after | 1 | NotYet | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| after | 1 | NotYet | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration |
| after | 1 | NotYet | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration |

## The edited files' ordinary census lines

| Side | File | Kind | Where | Reason |
| --- | --- | --- | --- | --- |
| before | src/compiler/factory/utilities.ts | Refused | src/compiler/core.ts:736:36 | a cast the runtime can't check |
| before | src/compiler/transformers/declarations/diagnostics.ts | Refused | src/compiler/core.ts:736:36 | a cast the runtime can't check |
| before | src/compiler/transformers/utilities.ts | Refused | src/compiler/core.ts:736:36 | a cast the runtime can't check |
| before | src/compiler/utilities.ts | Refused | src/compiler/core.ts:736:36 | a cast the runtime can't check |
| after | src/compiler/factory/utilities.ts | Refused | src/compiler/core.ts:736:36 | a cast the runtime can't check |
| after | src/compiler/transformers/declarations/diagnostics.ts | Refused | src/compiler/core.ts:736:36 | a cast the runtime can't check |
| after | src/compiler/transformers/utilities.ts | Refused | src/compiler/core.ts:736:36 | a cast the runtime can't check |
| after | src/compiler/utilities.ts | Refused | src/compiler/core.ts:736:36 | a cast the runtime can't check |

## Every complete edited owner

A zero below means no finding was recorded in that full-body attempt. It is
not a claim that the whole compiler or all dependencies now lower. All complete
findings, dependency locations and failed-statement boundaries are retained in
census-results.json and the raw owners-before/after streams.

| Owner | Before census line | Before findings | After census line | After findings |
| --- | --- | ---: | --- | ---: |
| isDeclarationWithTypeParameters | src/compiler/utilities.ts:2232:1 | 4 | src/compiler/utilities.ts:2232:1 | 2 |
| isDeclarationWithTypeParameterChildren | src/compiler/utilities.ts:2246:1 | 4 | src/compiler/utilities.ts:2244:1 | 0 |
| isPrimitiveLiteralValue | src/compiler/utilities.ts:12033:1 | 8 | src/compiler/utilities.ts:12030:1 | 0 |
| hasInferredType | src/compiler/utilities.ts:12067:1 | 4 | src/compiler/utilities.ts:12064:1 | 0 |
| createGetIsolatedDeclarationErrors.getDiagnostic | src/compiler/transformers/declarations/diagnostics.ts:664:5 | 10 | src/compiler/transformers/declarations/diagnostics.ts:673:5 | 14 |

The first owner still encounters a dependency NotYet at binder.ts:1034:17 from
its child-predicate call. getDiagnostic still encounters findAncestor's overload
refusal and binder/destructuring dependency walls. Its additional reachable
findings are recorded, not hidden as an improvement. The other three owners
have no findings after this adaptation.

## Commands

```sh
source /workspace/adamic-tools/env.sh
go build -o /workspace/scratch/99u-bin/ ./cmd/adamic ./stage3/census/tool
/workspace/scratch/99u-bin/tool /workspace/scratch/99u-main/src/compiler /workspace/scratch/99u-ordinary-before.jsonl
/workspace/scratch/99u-bin/tool /workspace/scratch/99u-measured/src/compiler /workspace/scratch/99u-ordinary-after.jsonl
python3 stage3/census/latent/make_overlay.py /workspace/adamic /workspace/scratch/99u-meter-overlay
```

For the owner observations, extend that generated overlay only by mapping its
existing replay/worker/main.go entry to owner-census.go.txt, then build the
existing replay/worker package with `go build -overlay=...`. Pass `-project`
with the complete src/compiler directory and `-sites` with the corresponding
owner-sites-before.json or owner-sites-after.json. The same compiler sources,
registration, lexical captures, full attempt loop and no-output guards are used.

## Latest Go-only main advance

Landing base: a3f9467233562ef1d8834b064dddf1a149dd036c. The complete ordinary
79-root tables above remain bound to the b0efce6d compiler. The TypeScript
adaptation inputs and oracle harness did not change in the intervening commits,
and full apply confirms every source byte matches the oracle-tested tree.

A separate current compiler and the same generated meter overlay are built at
a3f94672. All 79 roots type-check with zero diagnostics before and after. Both
edited files are rerun through the unmodified ordinary per-file driver. Each
of the five complete owners is also attempted before and after using the same
full-body selector, all 79 roots and both no-output guards. No findings are
suppressed, and the owner findings are unchanged. The current mutant confirms
the restored Debug.type call still fails through a correlated NotYet cause.
Raw current owner streams and current-file-census.json are retained separately.

| Side | Current file | Kind | First stop |
| --- | --- | --- | --- |
| before | utilities.ts | Refused | compiler/core.ts:736:36 |
| before | transformers/declarations/diagnostics.ts | Refused | compiler/core.ts:736:36 |
| after | utilities.ts | Refused | compiler/core.ts:736:36 |
| after | transformers/declarations/diagnostics.ts | Refused | compiler/core.ts:736:36 |
