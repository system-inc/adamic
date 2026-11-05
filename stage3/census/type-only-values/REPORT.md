# Value-bearing imports used only as types

Adamic base: `6761c7374aa7169c73602b6a74bc281b879a640d` (fresh clone of origin/main on 2026-11-05).
Upstream: TypeScript `v6.0.3`, `050880ce59e30b356b686bd3144efe24f875ebc8`.
Stock checker: npm TypeScript **6.0.3**. Adapted product key:
`d3d7119e05963133d152a90e3d34fcea1975bc1bdcf4a28bc35eb68cb382fddf`.

This is a census, not an adaptation. No existing committed source was changed.
The complete README and adapter for `10-type-imports` were read. That adapter
explicitly retains value-bearing imports used only as types because TS1484 does
not require their conversion.

## Counts

**27 import specifiers** have a value-bearing target and one or more
references, all in type positions. **1 whole declaration**
(1 finding) could become declaration-level `import type`,
removing their evaluation edges. **9 mixed declarations**
(26 findings) retain other bindings and their evaluation edges.

| Symbol kind | Specifiers |
| --- | ---: |
| enum | 1 |
| const enum | 26 |
| class | 0 |
| function | 0 |
| namespace with values | 0 |
| variable | 0 |
| other value | 0 |

## Whole declarations and removed edges

These are proposed declaration-level removals, not edits made by this census.
Inline `type` specifiers alone retain the ordinary import declaration and its
module evaluation edge under verbatimModuleSyntax. Exact changes in Adamic's
evaluation order across the 79 roots are deferred to the next step. The table
supplies the removed edges rather than claiming any computed order changes.

| Importing file | Declaration line | Source module | Resolved target | Findings |
| --- | ---: | --- | --- | --- |
| src/compiler/types.ts | 1 | `./_namespaces/ts.js` | src/compiler/_namespaces/ts.ts | NodeFactoryFlags |

## Every finding

Paths and 1-based lines refer to the final adapted tree, not pristine upstream.
Whole means all remaining non-type bindings in that declaration are findings;
mixed means the declaration retains other bindings. Reference lines and context,
local aliases, symbol flags, remaining bindings, root paths, checker options and
SHA256 hashes of every root are in [census.json](census.json).

| File | Line | Imported name | Symbol kind | Source module | Declaration |
| --- | ---: | --- | --- | --- | --- |
| src/compiler/debug.ts | 6 | CheckMode | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 8 | EmitFlags | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 65 | ModifierFlags | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 68 | NodeCheckFlags | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 69 | NodeFlags | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 75 | RelationComparisonResult | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 76 | ScriptKind | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 78 | SignatureCheckMode | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 79 | SignatureFlags | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 80 | SnippetKind | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 85 | SyntaxKind | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 87 | TransformFlags | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/debug.ts | 89 | TypeFacts | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/factory/baseNodeFactory.ts | 4 | SyntaxKind | const enum | `../_namespaces/ts.js` | mixed |
| src/compiler/program.ts | 69 | EmitOnly | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/transformer.ts | 16 | EmitHint | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/transformer.ts | 18 | EmitOnly | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/transformers/module/esnextAnd2015.ts | 12 | EmitHint | const enum | `../../_namespaces/ts.js` | mixed |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | 4 | EmitHint | const enum | `../../_namespaces/ts.js` | mixed |
| src/compiler/types.ts | 11 | NodeFactoryFlags | const enum | `./_namespaces/ts.js` | whole |
| src/compiler/utilities.ts | 250 | InternalEmitFlags | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/utilities.ts | 461 | ObjectFlags | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/utilities.ts | 525 | SignatureFlags | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/watchPublic.ts | 56 | JSDocParsingMode | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/watchPublic.ts | 78 | ScriptTarget | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/watchPublic.ts | 92 | WatchDirectoryFlags | const enum | `./_namespaces/ts.js` | mixed |
| src/compiler/watchUtilities.ts | 46 | PollingInterval | enum | `./_namespaces/ts.js` | mixed |

## Method and validation

The input is the final output of `stage3/apply.sh`, with all adaptations on this
base, before any upstream oracle build. The project uses all **79**
`src/compiler/**/*.ts` roots, including generated files and namespace barrels,
and the same checker options as `internal/load/load.go` and adaptation 10.
No tsconfig omission or relaxed checker option is used.

[The read-only census driver](census.cjs) obtains each local import alias from
`checker.getSymbolAtLocation` and resolves it with `getAliasedSymbol`. A target
must have `SymbolFlags.Value`. Kind labels use ConstEnum, RegularEnum, Class,
Function, ValueModule, then Variable flags. All named, default and namespace
bindings are inventoried; already type-only clauses/specifiers are excluded.
References are matched by local alias symbol identity in the importing file,
with shorthand value symbols and local export targets handled explicitly.
Import declarations themselves are excluded from references. Qualified name
receivers are lifted to the complete name before asking `isPartOfTypeNode`.
Type queries (`typeof X`) and explicit type exports are type positions; class
extends expressions and ordinary value exports are value positions. Comments
and JSDoc prose are not executable/type references.

Unused bindings are listed separately in JSON and excluded from the headline:
**0** unused bindings; unresolved aliases: **4**.
The four unresolved aliases are the Node builtin namespaces `_fs`, `_path`,
`_os`, and `_crypto` in `sys.ts`. Their 19, 6, 2, and 1 references, respectively,
are all runtime uses, so they cannot qualify regardless of their target kind.
These inventories make zero-reference and resolution gaps visible.

The driver builds a second program with candidate inline `type` insertions in
memory only. Semantic diagnostics: **59 before,
59 after**; new diagnostic messages:
**0**; TS1361/TS1362 value-use errors:
**0**. This validates type-only eligibility without
modifying the adapted sources. No Adamic lowering or 79-root ordering comparison
is claimed. The source hash ledger identifies the exact measured tree. All 79 source hashes
were rechecked after the census and remained unchanged.

Reproduction from the Adamic checkout, with a new adapted output path and the
pinned upstream checkout available:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api --no-audit --no-fund > /workspace/scratch/census-api-install.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-pristine
stage3/apply.sh /workspace/scratch/typescript-adapted > /workspace/scratch/apply.log 2>&1
node stage3/census/type-only-values/census.cjs /workspace/scratch/typescript-adapted /workspace/scratch/census.json > /workspace/scratch/census.log 2>&1
```

The checked-in JSON adds base revision, upstream pin, product key and ordering
status to the driver's output. Setup, apply, and census logs remain external.
The upstream oracle suite is not needed for a report-only change and was not run.
