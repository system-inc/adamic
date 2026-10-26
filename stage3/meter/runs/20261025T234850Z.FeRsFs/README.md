# Stage 3 meter 20261025T234850Z

The area label repeats the pinned main tree for reporter compatibility.

| Run | Main SHA | Whole program | Own file | Lowering | tsc entry diagnostics |
| --- | --- | ---: | ---: | ---: | ---: |
| Current | ff9ce59b73aa3aa1c15cbc3d3559e07e1cc8c4cb | 2/79 | 58/79 | 2/79 | 144 |
| 20261025T121703Z.hljfOJ | 8156b785d5ba5a286d5e3c97bcc969088530376b | 2/79 | 58/79 | 2/79 | 144 |

Lowering counts ordinary successful IR lowering. No native output is claimed.

Elapsed: 30m 42s.

## Compiler question #tzd3gjg

Pinned main and both per-ref compilers: `ff9ce59b73aa3aa1c15cbc3d3559e07e1cc8c4cb`. The required `git merge-base --is-ancestor 0ee00f60 ff9ce59b73aa3aa1c15cbc3d3559e07e1cc8c4cb` passed.

From `main/latent.jsonl.gz`: **214 distinct sites across 17 files** whose reason starts with "checked predicate membership by an enum tag another member can hold". **353 distinct sites** name a checked predicate of any kind in their reason, 52.8% of R-T80's 669 .ts sites for scale. This is a partial first-error blocker census, not an equivalent census of every predicate declaration.

From `main/tsc/latent.jsonl.gz`: **0 observed enum-tag sites and 0 observed checked-predicate sites**. The stream contains only a checker-rejected header; 144 diagnostics block entry lowering, so these zeros mean unmeasured, not absence of blockers.

| File | Main enum-tag sites | tsc entry enum-tag sites |
| --- | ---: | ---: |
| src/compiler/binder.ts | 2 | unmeasured |
| src/compiler/checker.ts | 2 | unmeasured |
| src/compiler/emitter.ts | 1 | unmeasured |
| src/compiler/factory/emitHelpers.ts | 1 | unmeasured |
| src/compiler/factory/utilities.ts | 17 | unmeasured |
| src/compiler/parser.ts | 2 | unmeasured |
| src/compiler/program.ts | 2 | unmeasured |
| src/compiler/transformers/classFields.ts | 4 | unmeasured |
| src/compiler/transformers/classThis.ts | 1 | unmeasured |
| src/compiler/transformers/declarations/diagnostics.ts | 1 | unmeasured |
| src/compiler/transformers/destructuring.ts | 2 | unmeasured |
| src/compiler/transformers/esnext.ts | 2 | unmeasured |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | 1 | unmeasured |
| src/compiler/transformers/namedEvaluation.ts | 1 | unmeasured |
| src/compiler/transformers/utilities.ts | 11 | unmeasured |
| src/compiler/utilities.ts | 123 | unmeasured |
| src/compiler/utilitiesPublic.ts | 41 | unmeasured |
| Total | 214 | unmeasured (0 records) |

Counts deduplicate source `file:line:column` sites, excluding repeated attempts. For all checked-predicate reasons, this also equals the meter's unique `(kind, where, reason, text)` finding count: 353. Reproduce with `python3 stage3/meter/runs/20261025T234850Z.FeRsFs/count-predicates.py`; exact results are in `checked-predicates.json`.

## Observations and execution

The daily command exited 0 without intervention in 1842 seconds (30m 42s), versus the previous run's 2086 seconds (34m 46s). Both independent copies agree: ordinary whole program 2/79, own file 58/79, lowering 2/79, and tsc entry 144 diagnostics, unchanged from `8156b785d5ba5a286d5e3c97bcc969088530376b`.

Each copy records one latent lowering panic: `src/compiler/moduleSpecifiers.ts:335:1`, `runtime error: invalid memory address or nil pointer dereference`. Previous main also recorded one panic. Errors remain 0. This is a recorded measurement finding, not a failing daily-command step.

Latent totals per copy: NotYet 1144 (previous 876), Refused 3083 (previous 3345), SkippedDependency 2 (previous 2). New top-ten reasons are enum-tag checked-predicate membership (214, previously 0) and checked-predicate membership over an open tag domain (45, previously 0). The unproven-return predicate reason fell from 388 to 255; a union/optional field read with record storage fell from 72 to 69; the TsConfigSourceFile record reason fell from 60 to 57. These observations change both compiler and adaptations and do not isolate a causal change.

The documented command was run in the background under its hard limit, with this supported explicit pin exported first:

```sh
export STAGE3_METER_MAIN_REF=ff9ce59b73aa3aa1c15cbc3d3559e07e1cc8c4cb
source /workspace/adamic-tools/env.sh && STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error timeout --kill-after=60s 7200 bash /workspace/adamic/stage3/meter/twice-daily.sh > /tmp/stage3-meter-daily.log 2>&1
```

Choice: use the supported explicit main-ref override because the command fetches main again; this preserves the SHA the branch was cut from. The branch date is October 26; the run directory records its actual October 25 UTC start.

Setup: `export GOPROXY='https://proxy.golang.org|direct'`, then `timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/stage3-meter-setup.log 2>&1`. It passed. Timing lines: Node ready 0.050s; Go ready 0.061s; clang ready 0.443s; tsc ready 0.503s; Markdown dependencies installed step-duration 0.750s and ready 1.353s; submodules ready 142.687s; shared cache off 142.689s; Go build ready 294.354s; test binaries deferred 294.447s; build cache warm 294.448s; done 294.477s. `nproc`: 5; CPU quota: `400000 100000`. Node: v24.19.0. Full logs are `setup.log.gz` and `daily.log.gz`.

No compiler, adaptation, meter, or check implementation changed, so no new mutant or repository test suite was run. The authorized daily command and its built-in report validation passed. All eight census/latent streams and every log are retained compressed.
