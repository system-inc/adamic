# tracing-fs-import resumed stop evidence

Task nqp4p7g. Base remains 9ad919acec5955055898f4978a100a72a3742f4e,
as explicitly requested. The resumed tree started byte-identical to checkpoint
4d8f53a26d4eb8ce670d0ef0f3d6760323d78f57 (a3c23a4b plus its checkpoint).
The previous stop report is retained as initial-stop.md and in Git history.

## Required closed-removal condition fails

helper-reference-census.json and helper-references/ record TypeScript checker
symbol identity for all ten emitted files, including shorthand value reads.
The definition is located in hostErrors' emitted section, also inside the
outer typescript.js IIFE. Only references resolving to that exact binding
count. The function's declaration itself and unrelated errorMessage locals
and property names are excluded.

| Emitted file | Before catch reads | Before export reads | After export reads |
| --- | ---: | ---: | ---: |
| _tsc.js | 1 | 0 | 0 |
| run.js | 1 | 3 | 3 |
| typescript.js | 1 | 3 | 3 |
| Other seven emitted files | 0 | 0 | 0 |
| Total | 3 | 6 | 6 |

Before: nine total references, only three in the removed catches.
After: six total references. The requested all-before-in-catch / after-zero
condition is false. In typescript.js the remaining reads are at 541:23,
184103:23, and 198938:3. In run.js they are at 197043:23, 200358:23,
and 209524:23. The last typescript.js read is a shorthand export read;
its value symbol is resolved with getShorthandAssignmentValueSymbol.
The helper remains exported through compiler/_namespaces/ts.ts.

final-public-contracts.json proves that both public declarations and all
2,256 runtime export keys match, including errorMessage. Removing those live
exports to force zero references would violate runtime export-key equality.
Only _tsc.js eliminates the helper; run.js and typescript.js retain it.
No implementation branch is published. The updated draft is checkpointed.

## Completed resumed observations

- 47 now drops its unused tracing errorMessage import, including the preceding
  lazy finished state. It still pins startTracing and checks qualified bindings.
- Fresh official final apply: exit 0. Fresh npm ci, build and harness: exit 0.
- Final adamic types: exit 0 on all 79 roots.
- Final full tracing census: exit 0, no checker diagnostics, complete trailer.
- Final core.ts per-file census: NotYet sys.ts:1654:13,
  node:inspector.Session.post.
- Restoring mutant: restores exact control tracing.ts bytes, including lazy
  require and ambient require declaration. Core's first stop returns to
  Refused tracing.ts:380:1, an ambient value declaration. Original unchanged.
- host-declarations.json lists and pins the complete hostErrors module body:
  errorCode, errorMessage, propertyOf. All are function declarations, with
  no other top-level statements, and control/final bytes match.
- Private catch instrumentation has an exact undo proof for all three bundles.
  Its forced-failure positive control counts one catch execution. That is not
  a whole-lane zero-execution observation.
- Both public declarations byte-identical; runtime exports identical.
- final-emitted-js.diff retains every final emitted JavaScript hunk. The hunk
  certificate is incomplete because its required closed-removal premise fails.

## Incomplete and failed observations, retained without passing claims

The global token certificate stopped on _tsc.js with a four-token count
mismatch; it did not certify the diff. Rename normalization may require
accounting for esbuild's shorthand destructuring rewrite. No passing claim.
The first all-file binding census missed nested IIFE declarations and is
retained under helper-references-initial-incomplete. The corrected census
walks all declarations and verifies the nested definition is found.
The first resumed repeated-trace driver stopped after the first control run
because the ordinary project CLI produces types.json and trace.json but no
legend.json. Four-run trace comparison is incomplete. No claim about that
comparison is made.
An inventory attempt incorrectly expected all final emitted files to equal
previous draft emits and failed before publishing a result. The subsequent
final inventory is generated directly against the unchanged control.
The 398-shard lane, control whole-lane catch-zero count, full hunk certificate,
and final zero-edit replay are not run to completion after the closure stop.
No incomplete result is treated as a pass.
