# Readonly empty array slots, adaptation 95-readonly-empty-array-slots

Adamic, task #r11wppy item 1, the internal change, built as proposed. This is
type-only: the emitted JavaScript does not change. Every slot is local to
`createTypeChecker` or `@internal`, so the published declaration files do not
change, this directory carries no `api-sanctions.json`, and
`stage3/lane/sanctioned-api.json` is not regenerated.

## The edit

`adapt.cjs` inserts `readonly ` before the array type of 20 slots, 16 in
`src/compiler/checker.ts` and 4 in `src/compiler/types.ts`, for example:

```ts
function getUnionSignatures(signatureLists: readonly (readonly Signature[])[]): readonly Signature[] {
```

It parses both files with stock TypeScript 6.0.3 and finds each slot by name:
a checker function's return type or parameter (the function must be the only
one of its name and local to `createTypeChecker`), or a types.ts interface's
method return type, method parameter or property (the member or its interface
must carry `@internal`). It refuses before any write unless every slot's type
is upstream's, or already readonly, which it leaves, so a second run changes
nothing. No source coordinate is a rewrite address.

The directory sorts after 95-checker-axolotl-c and before 96-emit-axolotl-a.
95-checker-axolotl-c pins the whole text of `SymbolLinks` by its sha256 before
it edits `tupleLabelDeclaration`, so `SymbolLinks.variances` must still read
as upstream writes it when that pin runs. No other adaptation reads or pins
any of the 20 slots or a declaration that contains one.

## Why

Each slot holds a value that can be tsc's shared `emptyArray`, typed
`never[]`. Seen as a writable `T[]`, that's a view that could write a `T` where
never is read, and Adamic refuses it. A readonly slot is a read-only view of
it, which can't write, so the refusal clears at its site.

`emptyArray` and `resolvingEmptyArray` stay `never[]`, by Adamic's ruling. A
readonly sentinel would reach `CompilerOptions.lib`, `.types` and public
services types and never type-check, and nothing needs it, because the slot is
what clears a site.

## The slots

Sites are on apply's tree. A slot with no site is on the chain to one that has
them: its value flows into a slot that clears a site, or out of one.

| Slot | File | Sites it holds |
|---|---|---|
| getNamedOrIndexSignatureMembers, return | checker.ts | none |
| getImplementsTypes, return | checker.ts | none |
| createInstantiatedSymbolTable, `symbols` | checker.ts | none |
| addInheritedMembers, `baseSymbols` | checker.ts | none |
| getUnionSignatures, return | checker.ts | checker.ts:14320:57, checker.ts:14365:26 |
| getUnionIndexInfos, return | checker.ts | checker.ts:14499:16 |
| relateVariances, `variances` | checker.ts | none |
| excludeProperties, `properties` | checker.ts | none |
| getVariances, return | checker.ts | none |
| getVariancesWorker, return | checker.ts | none |
| hasCovariantVoidArgument, `variances` | checker.ts | none |
| findDiscriminantProperties, `sourceProperties` | checker.ts | none |
| getSpellingSuggestionForName, `symbols` | checker.ts | none |
| getNonInheritedProperties, `baseTypes` | checker.ts | none |
| getNonInheritedProperties, `properties` | checker.ts | none |
| getPropertiesOfContainerFunction, return | checker.ts | checker.ts:50927:20 |
| TypeChecker.createAnonymousType, `constructSignatures` | types.ts | none |
| TypeChecker.createAnonymousType, `indexInfos` | types.ts | none |
| EmitResolver.getPropertiesOfContainerFunction, return | types.ts | none |
| SymbolLinks.variances | types.ts | checker.ts:25035:31 |

## Left out

- The four written slots: `InterfaceType.resolvedBaseTypes`, the local
  `resolvedImplementsTypes` in getImplementsTypes, and services' two, the local
  `symbols` in completions.ts and the local `result` in findAllReferences.ts's
  mergeReferences. Each is written through, so a readonly view of it can't
  stand.
- `append`'s overload parameter `to` in core.ts: its implementation pushes into
  it.
- The slots whose readonly reaches `append` or services:
  getAlternativeContainingModules' return, `SymbolLinks.extendedContainers`,
  the local `indexInfos` in resolveAnonymousTypeMembers,
  getAllPossiblePropertiesOfTypes (its checker return and its `TypeChecker`
  method), and the internal `ObjectType.properties`.

## Proof

Sealed, by scan-all's instrumented probe over sys.ts's closure in apply's tree,
with stock TypeScript 6.0.3 (J-5j08c4v):

| Measurement | Before | After | Mutant |
|---|---:|---:|---:|
| Distinct stops, every rule | 3,481 | 3,475 | 3,477 |
| never[] stops, distinct | 98 | 92 | 94 |
| The five listed sites stopped | 5 | 0 | 2 |
| Stops new against before | | 0 | |
| Stock tsc errors over src/compiler | 1 | 1, the same | 1, the same |

"Before" is apply's tree with the 18 edited lines put back exactly as upstream
writes them. Six stops go with this adaptation: the five listed sites
(checker.ts:14320:57, 14365:26, 14499:16, 25035:31 and 50927:20) and
checker.ts:50930:74, a second never[] seen as Symbol[] in
getPropertiesOfContainerFunction. None is new. The mutant takes
getUnionSignatures' one `readonly ` back out and brings back exactly its two
sites, 14320:57 and 14365:26. The set type-checks clean under stock tsc with
nothing else changed (J-7tkx9gz), and the adapted tree's product test and
stage3/lane pass with it.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
