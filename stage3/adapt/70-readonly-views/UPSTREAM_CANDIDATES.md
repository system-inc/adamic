# Upstream candidates

The initial waves verified no reachable wrong-typed write. Wave 5 adds the
pristine public-API counterexample below. No upstream issue was filed and the
global upstream ledger was not edited.

Two inspected writers were retained:

| Wider view | Declaration | Write | Reachability assessment |
|---|---|---|---|
| `SourceFileLike.lineMap` | `src/compiler/types.ts:4289:5` | `src/compiler/scanner.ts:504:35`, `sourceFile.lineMap = computeLineStarts(sourceFile.text)` | Latent wider-view capacity. The actual RHS is `number[]`; the stock checker confirms it fits the original `SourceFile.lineMap` slot, `readonly number[]`. This write does not insert `undefined`. Its public status is now sanctioned, but its actual writer prevents readonly. |
| `clear(array: unknown[])` | `src/compiler/core.ts:311:23` | `src/compiler/core.ts:312:5`, `array.length = 0` | Latent element-widening capacity. This actual write removes elements and inserts no wrong-typed value. The consumer writes, so its parameter stays mutable. |

The stock-checker observations are in `evidence/writers.json`. They cover these two inspected writer families, not every possible transitive writer for every retained refusal. Unreviewed sites are not asserted to be latent or reachable holes. `REMAINING.md` lists them with their exact census reasons.

Ledger format read from `origin/codex/stage3-regex-captures:stage3/upstream/LEDGER.md`. It requires an actual upstream counterexample, expected and observed behavior, and a draft issue. Neither inspected write supplies a wrong-typed counterexample, so neither receives a draft candidate entry.


## Candidate: setTextRange retains a narrower destination type after replacing its positions

Status: reachable on pristine upstream and the adapted compiler; draft issue,
not filed. The destination is deliberately unchanged by adaptation 70.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/factory/utilitiesPublic.ts:10-11 forwards range into
setTextRangePosEnd at utilities.ts:10664. Its checker-resolved callees write
(range as TextRange).pos at utilities.ts:10645:5 and
(range as TextRange).end at utilities.ts:10655:5, returning the same generic T. Exact resolved writer locations
are recorded with the candidate evidence.

Counterexample against the published API:

```typescript
import * as ts from "typescript";
const range = { pos: 0, end: 0 } as const;
const returned = ts.setTextRange(range, { pos: 1, end: 2 });
const promisedPos: 0 = returned.pos;
const promisedEnd: 0 = range.end;
console.log([promisedPos, promisedEnd]);
```

Expected from the literal-type promises: both readings are zero, or the writing
call is rejected or its resulting contract is widened. Observed: the stock
6.0.3 checker accepts this program with strict, exactOptionalPropertyTypes,
noUncheckedIndexedAccess, verbatimModuleSyntax and erasableSyntaxOnly. Node
prints [1, 2]. The write inserts values the original readonly literal type
forbids, through an actually executed exported function. The executable witness
uses an injected, version-checked compiled compiler binding with the exact
published declaration type; it emits the fixture with the stock checker before
executing it. Neither an unchecked cast on range nor a weakened option is used.

Pristine and adapted observations are saved in evidence/wave5. A disposable
compiled-function mutant returning range without writing produces [0, 0] and
fails the witness's Node assertion, proving the observed values are not a
hard-coded report. No narrow-position instance has been established among the
compiler's own callers. Public reachability is proven; internal reachability
is not claimed.

### Draft GitHub issue

Title: setTextRange writes new positions but preserves readonly literal destination types

Version: TypeScript 6.0.3, reproduced with stock checker and Node 24.19.0.

setTextRange<T extends TextRange> accepts a readonly literal destination and
returns T after mutating its pos and end. The reproduction above typechecks but
prints [1, 2] through variables whose types are both literal 0. A truthful
contract must account for the mutation and narrower incoming aliases; widening
only the return type would not repair the original range alias. No fix is
proposed in this unit. Its readonly location input is independent of this
writing destination. Kirk decides whether to file the drafted issue.
