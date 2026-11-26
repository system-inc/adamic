Adamic ruled this adaptation on Nov 22 2026 (#bbssdcy, family 36's `deduplicateSorted`, family 15 C of the Nov 24 scout pass) and accepts the head after its review and census.

The `-before` site is `src/compiler/core.ts:759:13`, the `case true:` in the private function `deduplicateSorted`. The scout identified it as the first lowering stop of the three `getPreEmitDiagnostics` units in `program.ts` and `performCompilation` in `executeCommandLine.ts`, with the reason "a case whose type differs from the switch's". Loom checks those units and the absence of new stops at main after review.

| Comparer result | Original switch | Adapted body |
| --- | --- | --- |
| `true` | Continue | Continue |
| `false` | Push | Push |
| `0` | Continue | Continue |
| `-0` | Continue | Continue |
| `-1` | Throw "Array is unsorted." | Throw "Array is unsorted." |
| `1` | Push | Push |
| `NaN` | Push | Push |
| `undefined` | Push | Push |
| `null` | Push | Push |
| String, for example `"0"` | Push | Push |

The original switch calls `comparer(next, last)` once per step. The replacement keeps that one result and uses the `typeof` narrowing in [rulings.md:145](../../../docs/rulings.md#L145). The numeric arm compares by strict equality, so `-0` matches `0`. Mixed-representation switches remain governed by [rulings.md:269](../../../docs/rulings.md#L269).

`test.cjs` reads this adapter's before and after refs, checks the output and the unchanged surrounding text, then runs both bodies with stock TypeScript's JavaScript emit and recording stubs in Node. Each drive compares both argument identities in order, call count, the returned elements by identity, and the thrown message. The counting witness observes four sorted calls in order, `1:1 2:1 2:2 4:2`, followed by one call on unsorted input, `2:4`, before the throw.

The after owner text pinned by `99-correctness-axolotl-r` is this adapter's before text. Every other statement, including `let last = array[0]!;` and the read into `next`, keeps its text. Family 1 E1's earlier stop in `sortAndDeduplicate` at that declaration is left to #3zvr9mq. The public declarations and every caller retain their source and types.
