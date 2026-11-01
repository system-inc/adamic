# 96a: transformer.ts's lexical environment, honestly optional

Axolotl (#axolotl), task #vy7ncrg. This adaptation is byte-identical in JavaScript and changes no public API. It edits only `transformNodes` in src/compiler/transformer.ts.

## The cause

`transformNodes` keeps four declaration lists for the environment it's building: `lexicalEnvironmentVariableDeclarations`, `lexicalEnvironmentFunctionDeclarations`, `lexicalEnvironmentStatements` and `blockScopedVariableDeclarations`.

- **Declared:** each is `T[]`.
- **Held:** each starts an environment as `undefined!` (startLexicalEnvironment, startBlockScope, dispose).
- **Saved:** each is saved onto a stack when an environment starts and restored from it when one ends. A saved payload is undefined whenever the outer environment had accumulated nothing.

So restoring from a stack is `T[] | undefined` into `T[]`. Those are the four TS2322 diagnostics at 563, 564, 565 and 614. Adaptation 33 declined those reads: "its saved declaration-list payload intentionally can be undefined".

## The repair

- Each of the four locals becomes `T[] | undefined`.
- Each of their four stacks becomes `(T[] | undefined)[]`.
- Each local's placeholders, six in all, become plain `undefined`.

The two stacks' disposal placeholders in `dispose` stay `undefined!`, as #9wc5q5j reads them. Correctness, on #axochk (Nov 1): no word from Adamic is needed. A head that tells the type the truth may remove a placeholder, and the Oct 6 rule sanctions it for a non-public owner.

## Why it's sound

Every reader of the four locals already tests absence:

- `if (!x)` before `x = [..]` / `x.push(..)`: hoistVariableDeclaration, hoistFunctionDeclaration, addInitializationStatement (435, 453, 468).
- `x ||` and then `if (x)` before each spread or list: endLexicalEnvironment (527-556).
- `some(x) ? .. x.map(..) : undefined`: endBlockScope (603-608). `some` is core.ts's type guard.
- `(x || (x = [])).push(..)`: addBlockScopedVariable (623).

No reader dereferences a local without one of these, which is why the checker finds nothing new.

## Proof

- **The shape:** the adapter parses with stock TypeScript 6.0.3 and finds each local and stack by name inside `transformNodes`. It accepts only the inventoried annotations, before or after. It counts each local's placeholders and refuses any other count, and refuses a half-adapted file. It plans every edit before its single write.
- **Diagnostics:** `adamic types` on the adapted tree's transformer.ts, at main a23ba096:
  - whole program, 144 before and 140 after
  - transformer.ts own-file, 6 before and 2 after
  - The four removed are exactly 563, 564, 565 and 614, with no new diagnostic anywhere. The two left are 96b's.
- **JavaScript:** `ts.transpileModule` of transformer.ts before and after gives byte-identical output (26,861 bytes each).
- **Idempotence:** a second run edits nothing.

Not yet run: the sharded lane oracle (stage3/lane/shards.json's 398 units). It's owed before this goes to Loom.
