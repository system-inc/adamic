# Upstream candidates

No reachable wrong-typed write was verified in this wave. No upstream issue was filed and the upstream ledger was not edited.

Two inspected writers were retained:

| Wider view | Declaration | Write | Reachability assessment |
|---|---|---|---|
| `SourceFileLike.lineMap` | `src/compiler/types.ts:4289:5` | `src/compiler/scanner.ts:504:35`, `sourceFile.lineMap = computeLineStarts(sourceFile.text)` | Latent wider-view capacity. The actual RHS is `number[]`; the stock checker confirms it fits the original `SourceFile.lineMap` slot, `readonly number[]`. This write does not insert `undefined`. The declaration is public and gets no edit. |
| `clear(array: unknown[])` | `src/compiler/core.ts:311:23` | `src/compiler/core.ts:312:5`, `array.length = 0` | Latent element-widening capacity. This actual write removes elements and inserts no wrong-typed value. The consumer writes, so its parameter stays mutable. |

The stock-checker observations are in `evidence/writers.json`. They cover these two inspected writer families, not every possible transitive writer for every retained refusal. Unreviewed sites are not asserted to be latent or reachable holes. `REMAINING.md` lists them with their exact census reasons.

Ledger format read from `origin/codex/stage3-regex-captures:stage3/upstream/LEDGER.md`. It requires an actual upstream counterexample, expected and observed behavior, and a draft issue. Neither inspected write supplies a wrong-typed counterexample, so neither receives a draft candidate entry.
