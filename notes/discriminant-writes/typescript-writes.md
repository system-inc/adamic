# TypeScript v6.0.3 discriminant writes

Checkout: `050880ce59e30b356b686bd3144efe24f875ebc8` (tag `v6.0.3`). Stock npm checker: `typescript@6.0.3`.

Observed count: **30 writes outside constructor bodies**. The checker recognizes `kind`, `operator`, `isTypeOnly`, `version`, and tuple `length` as discriminants in unions this program uses. This is a count of write sites, not distinct source lines.

The census exposes the stock checker's internal `isDiscriminantProperty` and property-name resolver in memory. It uses checker types and assignability to match union members to writable views; it does not search source text for assignments. Only members whose declared field type is one literal supply a protected field. A wider view of such a member is still protected. A wider member which cannot contain the literal member is not protected by that member.

Construction in this count means a constructor body. Factory functions which initialize already created nodes are outside that exemption. A checker by itself does not prove IR holder confinement. This boundary matches the up-front refusal pass in this branch; constructor writes additionally go through the existing IR fresh-holder proof.

Run from the repository root:

```sh
node notes/discriminant-writes/count-typescript.mjs /path/to/TypeScript-v6.0.3 /path/to/typescript-npm-package > census.json
```

Prepare the checkout by running its own `scripts/processDiagnosticMessages.mjs src/compiler/diagnosticMessages.json`, and install `@types/node`, `source-map-support`, and `@types/source-map-support` into its `node_modules`. The npm checker package directory contains `lib/typescript.js`.

| File:line | Column | Field | View |
|---|---:|---|---|
| src/compiler/builder.ts:1424 | 16 | length | `IncrementalBuildInfoRoot[]` |
| src/compiler/checker.ts:1974 | 9 | length | `Signature[]` |
| src/compiler/checker.ts:8129 | 25 | length | `TrackedSymbol[]` |
| src/compiler/checker.ts:8132 | 25 | length | `(() => void)[]` |
| src/compiler/checker.ts:16171 | 13 | length | `Type[]` |
| src/compiler/core.ts:307 | 5 | length | `T[]` |
| src/compiler/core.ts:312 | 5 | length | `unknown[]` |
| src/compiler/core.ts:1596 | 13 | length | `(T &#124; undefined)[]` |
| src/compiler/factory/emitNode.ts:308 | 9 | length | `EmitHelper[]` |
| src/compiler/factory/nodeFactory.ts:3370 | 9 | operator | `Mutable<PrefixUnaryExpression>` |
| src/compiler/factory/nodeFactory.ts:4658 | 9 | isTypeOnly | `Mutable<ImportEqualsDeclaration>` |
| src/compiler/factory/nodeFactory.ts:4732 | 9 | isTypeOnly | `Mutable<ImportClause>` |
| src/compiler/factory/nodeFactory.ts:4897 | 9 | isTypeOnly | `Mutable<ImportSpecifier>` |
| src/compiler/factory/nodeFactory.ts:4956 | 9 | isTypeOnly | `Mutable<ExportDeclaration>` |
| src/compiler/factory/nodeFactory.ts:5016 | 9 | isTypeOnly | `Mutable<ExportSpecifier>` |
| src/compiler/parser.ts:2283 | 17 | length | `DiagnosticWithDetachedLocation[]` |
| src/compiler/parser.ts:8867 | 13 | length | `DiagnosticWithDetachedLocation[]` |
| src/compiler/program.ts:364 | 17 | length | `string[]` |
| src/compiler/program.ts:371 | 13 | length | `string[]` |
| src/compiler/sourcemap.ts:316 | 13 | length | `number[]` |
| src/compiler/sys.ts:233 | 17 | length | `(T &#124; undefined)[]` |
| src/compiler/tracing.ts:71 | 9 | length | `Type[]` |
| src/compiler/tracing.ts:163 | 9 | length | `{ phase: Phase; name: string; args?: Args &#124; undefined; time: number; separateBeginAndEnd: boolean; }[]` |
| src/compiler/tracing.ts:170 | 9 | length | `{ phase: Phase; name: string; args?: Args &#124; undefined; time: number; separateBeginAndEnd: boolean; }[]` |
| src/compiler/transformers/esnext.ts:351 | 25 | length | `VariableDeclaration[]` |
| src/compiler/utilities.ts:8509 | 5 | kind | `Mutable<Node>` |
| src/compiler/utilities.ts:8523 | 5 | kind | `Mutable<Node>` |
| src/compiler/utilities.ts:8535 | 5 | kind | `Mutable<Node>` |
| src/compiler/watchPublic.ts:773 | 21 | version | `FilePresentOnHost &#124; FilePresenceUnknownOnHost` |
| src/compiler/watchPublic.ts:808 | 17 | version | `FilePresenceUnknownOnHost` |

Independent agreement with @system_adamic_typescript has not been observed in this workspace. The user was asked for that result so the lists can be reconciled.

The final checker program has 78 compiler source files, including TypeScript's generated diagnostic map. One diagnostic remains in the unmodified pinned source: TS2488 at `src/compiler/transformers/classFields.ts:1503`, where a local set to undefined is filled by a visitor and then spread. That site is not among the counted field writes. The raw result records this diagnostic rather than hiding it.
